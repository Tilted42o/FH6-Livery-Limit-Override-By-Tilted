package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type compatibilityReport struct {
	ToolVersion    string      `json:"tool_version"`
	CreatedUTC     string      `json:"created_utc"`
	ExecutableName string      `json:"executable_name"`
	StorefrontHint string      `json:"storefront_hint"`
	Diagnostics    diagnostics `json:"diagnostics"`
	Note           string      `json:"note"`
}

func inferStorefront(path string) string {
	s := strings.ToLower(filepath.Clean(path))
	switch {
	case strings.Contains(s, "steamapps"):
		return "Steam path detected"
	case strings.Contains(s, "xboxgames") || strings.Contains(s, "windowsapps"):
		return "Microsoft Store / Xbox app path detected"
	default:
		return "unknown; include your storefront when sending this report"
	}
}

func reportDirectory() string {
	exe, err := os.Executable()
	if err == nil {
		return filepath.Dir(exe)
	}
	return os.TempDir()
}

func createCompatibilityReport(path string, d diagnostics) (string, error) {
	r := compatibilityReport{
		ToolVersion:    detectorVersion,
		CreatedUTC:     time.Now().UTC().Format(time.RFC3339),
		ExecutableName: filepath.Base(path),
		StorefrontHint: inferStorefront(path),
		Diagnostics:    d,
		Note:           "This report contains hashes, PE metadata, detector counts and at most a few tiny machine-code windows near possible layer-table anchors. It does not contain the game executable or a memory dump.",
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	name := fmt.Sprintf("Layer-Limit-Compatibility-%s.json", shortHash(d.SHA256))
	dirs := []string{reportDirectory(), os.TempDir()}
	var last error
	for _, dir := range dirs {
		if err = os.MkdirAll(dir, 0755); err != nil {
			last = err
			continue
		}
		dest := filepath.Join(dir, name)
		if err = os.WriteFile(dest, data, 0600); err == nil {
			return dest, nil
		}
		last = err
	}
	return "", last
}
