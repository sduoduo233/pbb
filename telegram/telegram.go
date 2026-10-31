package telegram

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sduoduo233/pbb/db"
)

type user struct {
	Id int64 `json:"id"`
}

type chat struct {
	Id int64 `json:"id"`
}

type message struct {
	From *user  `json:"from"`
	Chat chat   `json:"chat"`
	Text string `json:"text"`
}

type update struct {
	UpdateId int64    `json:"update_id"`
	Message  *message `json:"message"`
}

type response[T any] struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Result      T      `json:"result"`
}

var client = &http.Client{Timeout: 60 * time.Second}

var (
	mu     sync.Mutex
	cancel context.CancelFunc
	reload = make(chan struct{}, 1)
)

func Init() {
	go run()
}

// Reload interrupts the current poll so that a new bot token takes effect immediately
func Reload() {
	mu.Lock()
	if cancel != nil {
		cancel()
	}
	mu.Unlock()

	select {
	case reload <- struct{}{}:
	default:
	}
}

func wait(d time.Duration) {
	select {
	case <-reload:
	case <-time.After(d):
	}
}

func run() {
	var offset int64
	var lastToken string

	for {
		token, err := getSetting("telegram_bot_token")
		if err != nil {
			slog.Error("telegram", "err", err)
			wait(5 * time.Second)
			continue
		}

		if token != lastToken {
			offset = 0
			lastToken = token
		}

		if token == "" {
			wait(5 * time.Second)
			continue
		}

		ctx, c := context.WithCancel(context.Background())
		mu.Lock()
		cancel = c
		mu.Unlock()

		updates, err := call[[]update](ctx, token, "getUpdates", map[string]any{
			"offset":          offset,
			"timeout":         30,
			"allowed_updates": []string{"message"},
		})
		c()

		if err != nil {
			if errors.Is(err, context.Canceled) {
				continue
			}
			slog.Error("telegram get updates", "err", err)
			wait(5 * time.Second)
			continue
		}

		for _, u := range updates {
			offset = u.UpdateId + 1
			handle(token, u.Message)
		}
	}
}

func handle(token string, m *message) {
	if m == nil || m.From == nil || !strings.HasPrefix(m.Text, "/") {
		return
	}

	cmd, _, _ := strings.Cut(strings.Fields(m.Text)[0], "@")
	if cmd != "/start" && cmd != "/servers" {
		return
	}

	authorized, err := isAuthorized(m.From.Id)
	if err != nil {
		slog.Error("telegram", "err", err)
		return
	}

	var reply string
	if !authorized {
		reply = fmt.Sprintf("You are not authorized.\nYour user ID is %d.", m.From.Id)
	} else {
		switch cmd {
		case "/start":
			reply = "Available commands:\n/start - show this message\n/servers - list servers"
		case "/servers":
			reply, err = listServers()
			if err != nil {
				slog.Error("telegram", "err", err)
				reply = "Something went wrong"
			}
		}
	}

	err = sendMessage(context.Background(), token, m.Chat.Id, reply)
	if err != nil {
		slog.Error("telegram send message", "err", err)
	}
}

func listServers() (string, error) {
	type server struct {
		Label      string         `db:"label"`
		LastReport sql.NullInt64  `db:"last_report"`
		GroupLabel sql.NullString `db:"group_label"`
	}
	var servers []server
	err := db.DB.Select(&servers, "SELECT s.label, s.last_report, g.label AS group_label FROM servers AS s LEFT JOIN groups AS g ON s.group_id = g.id ORDER BY s.id ASC")
	if err != nil {
		return "", fmt.Errorf("db: %w", err)
	}

	if len(servers) == 0 {
		return "No servers.", nil
	}

	threshold := time.Now().Add(-time.Second * 30).Unix()

	var sb strings.Builder
	for _, s := range servers {
		switch {
		case !s.LastReport.Valid:
			sb.WriteString("⚪ ")
		case s.LastReport.Int64 > threshold:
			sb.WriteString("🟢 ")
		default:
			sb.WriteString("🔴 ")
		}
		sb.WriteString(s.Label)
		if s.GroupLabel.Valid {
			sb.WriteString(" (")
			sb.WriteString(s.GroupLabel.String)
			sb.WriteString(")")
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// Notify sends text to all authorized users if notifications are enabled
func Notify(text string) {
	enabled, err := getSetting("telegram_notify")
	if err != nil {
		slog.Error("telegram notify", "err", err)
		return
	}
	if enabled != "1" {
		return
	}

	token, err := getSetting("telegram_bot_token")
	if err != nil {
		slog.Error("telegram notify", "err", err)
		return
	}
	if token == "" {
		return
	}

	var ids []int64
	err = db.DB.Select(&ids, "SELECT user_id FROM telegram_users")
	if err != nil {
		slog.Error("telegram notify", "err", fmt.Errorf("db: %w", err))
		return
	}

	go func() {
		for _, id := range ids {
			err := sendMessage(context.Background(), token, id, text)
			if err != nil {
				slog.Error("telegram notify", "err", err, "user_id", id)
			}
		}
	}()
}

func isAuthorized(id int64) (bool, error) {
	var cnt int
	err := db.DB.Get(&cnt, "SELECT COUNT(*) FROM telegram_users WHERE user_id = ?", id)
	if err != nil {
		return false, fmt.Errorf("db: %w", err)
	}
	return cnt > 0, nil
}

func getSetting(key string) (string, error) {
	var value string
	err := db.DB.Get(&value, "SELECT value FROM settings WHERE key = ?", key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("db: %w", err)
	}
	return value, nil
}

func sendMessage(ctx context.Context, token string, chatId int64, text string) error {
	_, err := call[json.RawMessage](ctx, token, "sendMessage", map[string]any{
		"chat_id": chatId,
		"text":    text,
	})
	return err
}

func call[T any](ctx context.Context, token string, method string, params any) (T, error) {
	var zero T

	body, err := json.Marshal(params)
	if err != nil {
		return zero, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.telegram.org/bot"+token+"/"+method, bytes.NewReader(body))
	if err != nil {
		return zero, fmt.Errorf("%s: new request failed", method)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		// strip the url, it contains the bot token
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return zero, fmt.Errorf("%s: %w", method, err)
	}
	defer resp.Body.Close()

	var r response[T]
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return zero, fmt.Errorf("%s: decode: %w", method, err)
	}
	if !r.Ok {
		return zero, fmt.Errorf("%s: %s", method, r.Description)
	}

	return r.Result, nil
}
