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
)

//go:embed version.txt
var CURRENT_VERSION string

func AutoUpdate(installScript string) error {
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
