package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
)

type ExportFormat string

const (
	CSV           ExportFormat = "csv"
	JSON          ExportFormat = "json"
	JSONEncrypted ExportFormat = "encrypted_json"
)

type Bitwarden struct {
	key []byte
}

func NewClient(cfg *Config) (Bitwarden, error) {
	if cfg.BitwardenServer != "" {
		if err := exec.Command("bw", "config", "server", cfg.BitwardenServer).Run(); err != nil {
			return Bitwarden{}, err
		}
	}

	cmd := exec.Command("bw", "login", "--apikey")
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("BW_CLIENTID=%s", cfg.BitwardenID),
		fmt.Sprintf("BW_CLIENTSECRET=%s", cfg.BitwardenSecret),
	}...)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return Bitwarden{}, err
	}

	cmd = exec.Command("bw", "unlock", "--passwordenv=BW_PASSWORD")
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("BW_PASSWORD=%s", cfg.BitwardenMasterPassword),
	}...)

	out, err = cmd.CombinedOutput()
	if err != nil {
		return Bitwarden{}, err
	}

	key, err := extractSessionKey(out)
	if err != nil {
		return Bitwarden{}, err
	}

	return Bitwarden{key}, nil
}

func extractSessionKey(buf []byte) ([]byte, error) {
	re := regexp.MustCompile(`\$ export BW_SESSION="([^"]+)"`)
	matches := re.FindSubmatch(buf)
	return matches[1], nil
}

func (b *Bitwarden) Close() error {
	return exec.Command("bw", "logout").Run()
}

func (b *Bitwarden) Export(format ExportFormat) (io.Reader, error) {
	cmd := exec.Command("bw", "export", "--format", string(format), "--raw")
	cmd.Env = append(os.Environ(), fmt.Sprintf("BW_SESSION=%s", string(b.key)))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(out), nil
}
