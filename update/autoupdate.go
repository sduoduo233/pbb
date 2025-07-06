package update

import (
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"syscall"
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

func AutoUpdate(name string) error {

	// check update
	slog.Info("check update", "current_version", strings.Trim(CURRENT_VERSION, "\n"))
	latestVersion, err := getLatestVersion()
	if err != nil {
		return fmt.Errorf("get latest version: %w", err)
	}
	slog.Info("latest version", "version", latestVersion)

	if latestVersion == strings.Trim(CURRENT_VERSION, "\n") || strings.Trim(CURRENT_VERSION, "\n") == "DEV" {
		return nil
	}

	// check architecture
	var uname syscall.Utsname
	err = syscall.Uname(&uname)
	if err != nil {
		return fmt.Errorf("uname: %w", err)
	}

	downloadUrl := "https://dl.exec.li/" + name + "-"
	unameArch := ""
	for _, c := range uname.Machine {
		if c == 0 {
			break
		}
		unameArch += string(c)
	}

	if unameArch == "x86_64" {
		downloadUrl += "amd64"
	} else if strings.Contains(unameArch, "aarch64") || strings.Contains(unameArch, "arm64") {
		downloadUrl += "arm64"
	} else if strings.Contains(unameArch, "arm") {
		downloadUrl += "arm32-v7a"
	} else if strings.Contains(unameArch, "i386") || strings.Contains(unameArch, "i686") {
		downloadUrl += "386"
	} else {
		return fmt.Errorf("unsupported architecture: %s", unameArch)
	}

	// download binary

	slog.Info("download binary", "url", downloadUrl)

	req, _ := http.NewRequest("GET", downloadUrl, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http: bad status code %d", resp.StatusCode)
	}

	// replace current binary

	path, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	slog.Info("replace current binary", "path", path)

	err = syscall.Unlink(path)
	if err != nil {
		return fmt.Errorf("unlink current binary: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create new binary: %w", err)
	}

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		f.Close()
		return fmt.Errorf("write binary: %w", err)
	}
	f.Close()

	err = syscall.Chmod(path, 0755)
	if err != nil {
		return fmt.Errorf("chmod new binary: %w", err)
	}

	err = syscall.Exec(path, []string{path}, os.Environ())
	if err != nil {
		return fmt.Errorf("exec new binary: %w", err)
	}

	return nil
}
