package update

import (
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

//go:embed version.txt
var CURRENT_VERSION string

func getLatestVersion() (string, error) {
	resp, err := http.Get("https://dl.exec.li/version.txt")
	if err != nil {
		return "", fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http: bad status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("http: %w", err)
	}

	version := strings.Trim(string(data), "\n")

	return version, nil
}

func AutoUpdate(installScript string) error {

	// check update
	slog.Info("check update", "current_version", CURRENT_VERSION)
	latestVersion, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("get latest version: %w", err)
	}
	slog.Info("latest version", "version", latestVersion)

	if latestVersion == CURRENT_VERSION || CURRENT_VERSION == "DEV" {
		return nil
	}

	slog.Info("auto update begin", "script", installScript)

	// download install script

	req, _ := http.NewRequest("GET", installScript, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http: bad status code %d", resp.StatusCode)
	}

	path := "/opt/hub/" + path.Base(req.URL.Path)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("save install script: %w", err)
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		return fmt.Errorf("save install script: %w", err)
	}
	f.Close()

	slog.Info("download install script", "path", path)

	// run install script
	slog.Info("run install script", "path", path)

	cmd := exec.Command("/bin/bash", path)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Start()
	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("run install script: %w", err)
	}
	slog.Info("auto update end", "script", installScript)
	return nil
}
