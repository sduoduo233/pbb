package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type ServerMetric struct {
	Cpu            float32 `json:"cpu"`
	MemoryPercent  float32 `json:"memory_percent"`
	MemoryUsed     uint64  `json:"memory_used"`
	MemoryTotal    uint64  `json:"memory_total"`
	DiskPercent    float32 `json:"disk_percent"`
	DiskUsed       uint64  `json:"disk_used"`
	DiskTotal      uint64  `json:"disk_total"`
	NetworkOutRate uint64  `json:"network_out_rate"`
	NetworkInRate  uint64  `json:"network_in_rate"`
	SwapUsed       uint64  `json:"swap_used"`
	SwapTotal      uint64  `json:"swap_total"`
	SwapPercent    float32 `json:"swap_percent"`
}

var URL = os.Getenv("AGENT_URL")
var SECRET = os.Getenv("AGENT_SECRET")

func main() {
	slog.Warn("agent")

	slog.Warn("reporting to", "url", URL)

	for {

		m, err := getMetirc()
		if err != nil {
			slog.Error("could not get metrics", "err", err)
			time.Sleep(time.Second * 5)
			continue
		}

		err = sendReport(m)
		if err != nil {
			slog.Error("could not send request", "err", err)
			time.Sleep(time.Second * 5)
			continue
		}

		time.Sleep(time.Second * 5)

	}
}

func sendReport(m *ServerMetric) error {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, URL, bytes.NewReader(jsonBytes))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("x-secret", SECRET)
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("read body: %w", err)
		}
		return fmt.Errorf("bad response: status=%s, body=%s", resp.Status, string(body))
	}

	return nil
}

var (
	lastTime      time.Time = time.Now()
	lastIoCounter *net.IOCountersStat
)

func getMetirc() (*ServerMetric, error) {
	m := ServerMetric{}

	// cpu

	cpuPercent, err := cpu.Percent(time.Millisecond*100, false)
	if err != nil {
		return nil, fmt.Errorf("cpu: %w", err)
	}
	m.Cpu = float32(cpuPercent[0])

	// disk

	diskUsage, err := disk.Usage("/")
	if err != nil {
		return nil, fmt.Errorf("disk: %w", err)
	}
	m.DiskPercent = float32(diskUsage.UsedPercent)
	m.DiskTotal = uint64(diskUsage.Total)
	m.DiskUsed = uint64(diskUsage.Used)

	// network

	ioCounters, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("net: %w", err)
	}
	ioCounter := ioCounters[0]

	if lastIoCounter != nil {
		m.NetworkInRate = (uint64(ioCounter.BytesRecv) - uint64(lastIoCounter.BytesRecv)) / uint64(time.Since(lastTime).Seconds())
		m.NetworkOutRate = (uint64(ioCounter.BytesSent) - uint64(lastIoCounter.BytesSent)) / uint64(time.Since(lastTime).Seconds())
	}
	lastIoCounter = &ioCounter
	lastTime = time.Now()

	// memory

	memoryStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("mem: %w", err)
	}

	m.MemoryPercent = float32(memoryStat.UsedPercent)
	m.MemoryTotal = uint64(memoryStat.Total)
	m.MemoryUsed = uint64(memoryStat.Used)

	return &m, nil
}
