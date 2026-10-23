package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func scorePath() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return scorePathFrom(xdg, "")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return scorePathFrom("", home)
}

func scorePathFrom(xdgDataHome, home string) (string, error) {
	if xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "dandelion-drift", "scores.json"), nil
	}
	if home == "" {
		return "", fmt.Errorf("home directory is empty")
	}
	return filepath.Join(home, ".local", "share", "dandelion-drift", "scores.json"), nil
}
