package main

import (
	"path/filepath"
	"testing"
)

func TestScorePathUsesXDGDataHome(t *testing.T) {
	got, err := scorePathFrom("/tmp/xdg-data", "/ignored")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/xdg-data", "dandelion-drift", "scores.json")
	if got != want {
		t.Fatalf("score path=%q want %q", got, want)
	}
}

func TestScorePathFallsBackToUserDataDirectory(t *testing.T) {
	got, err := scorePathFrom("", "/home/player")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/home/player", ".local", "share", "dandelion-drift", "scores.json")
	if got != want {
		t.Fatalf("score path=%q want %q", got, want)
	}
}

func TestScorePathRequiresHomeWithoutXDG(t *testing.T) {
	if _, err := scorePathFrom("", ""); err == nil {
		t.Fatal("expected error for empty home directory")
	}
}
