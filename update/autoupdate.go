package update

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"syscall"
)

const RSA_PUBLIC_KEY = `
-----BEGIN PUBLIC KEY-----
MIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEAomjdVuUgZK7MfSmho6Rq
Ql5SvCDf1n9YCg/1Ofi2Q5chnGkIF1mENzJif50OstbElBEzrpojc8qXv8NCIY4G
QK18BmpLFVLeeX4jGeu6A2CLxPsvldqGcf2+sOMS7OxlrGwsOt0Zur7F4eM2ViHg
Jb7aVenmGFaFqt37XR0N5h/esf1afGhD95K2tO56f1epM7hFbJEn8pgQff/AzD7w
Z3kt/iivH0OMvGcpgnJAGeBdmndHqp1Aq8+J3uPKsDxNELlb9b52fSTmI0qUd064
Qbt4aI896885AAEopp+pfDC7BV2mXLwTwGPkSycLvRc4ChfRFgl1EvG5GnaN6CBe
FSBsuVeauB6kgJsoDYtYgtiKU7DtbNfS99eTj/04nEv4tXvJct73SA5mvBpbSikk
fCAeewOeRKH1pIAn7Lov3PL3aHN9lXzb+0jSHjEdqEp1DVcouPcJo+0/+t6P8PRq
GnaWNZfx/Tw5rIzaRA1XmMybvffvxwy+KrgA9RzAGnrbQpSLfYw1V5NFmP0SDMzc
mcrzn1z6NEbNY7R6m4qNaCndKwtPUk31+o63I1YgRrRoEnD3obWsFjQ07UtsSl07
OOh+lxE+OSIIrZVMde5SoibTBaiB0wYMLE8/7qZIr4TfCKJk8MVcHC1sdcM2O08s
b8HaxLp1R7RphvYjI2fRPOECAwEAAQ==
-----END PUBLIC KEY-----`

//go:embed version.txt
var CURRENT_VERSION string

func downloadFile(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http: bad status code %d", resp.StatusCode)
	}

	binaryBuffer := bytes.Buffer{}

	_, err = io.Copy(&binaryBuffer, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("write binary: %w", err)
	}

	return binaryBuffer.Bytes(), nil
}

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
		unameArch += string(rune(c))
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
	binaryBytes, err := downloadFile(downloadUrl)
	if err != nil {
		return fmt.Errorf("download binary: %w", err)
	}
	binarySig, err := downloadFile(downloadUrl + ".sig")
	if err != nil {
		return fmt.Errorf("download sig: %w", err)
	}

	// check signature

	err = Verify(binaryBytes, RSA_PUBLIC_KEY, binarySig)
	if err != nil {
		return fmt.Errorf("bad signature: %w", err)
	}

	slog.Info("signature verification passed")

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
	_, err = f.Write(binaryBytes)
	if err != nil {
		return fmt.Errorf("write new binary: %w", err)
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
