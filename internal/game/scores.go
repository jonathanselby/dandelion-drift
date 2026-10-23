package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type ScoreEntry struct {
	Initials string `json:"initials"`
	Score    int    `json:"score"`
}

type HighScores struct {
	Scores []ScoreEntry `json:"scores"`
}

func LoadScores(path string) (HighScores, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return HighScores{}, nil
	}
	if err != nil {
		return HighScores{}, err
	}
	var scores HighScores
	if err := json.Unmarshal(data, &scores); err != nil {
		return HighScores{}, fmt.Errorf("decode high scores: %w", err)
	}
	return scores.Normalized(), nil
}

func SaveScores(path string, scores HighScores) error {
	scores = scores.Normalized()
	data, err := json.MarshalIndent(scores, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func (h HighScores) Normalized() HighScores {
	h.Scores = normalizeScores(h.Scores)
	return h
}

func (h *HighScores) Add(entry ScoreEntry) {
	entry.Initials = normalizeInitials(entry.Initials)
	if entry.Initials == "" {
		entry.Initials = "?"
	}
	h.Scores = normalizeScores(append(h.Scores, entry))
}

func (h HighScores) Qualifies(score int) bool {
	table := h.Normalized().Scores
	if len(table) < 5 {
		return true
	}
	return score > table[len(table)-1].Score
}

func normalizeScores(in []ScoreEntry) []ScoreEntry {
	out := make([]ScoreEntry, 0, len(in))
	for _, e := range in {
		e.Initials = normalizeInitials(e.Initials)
		if e.Initials == "" || e.Score < 0 {
			continue
		}
		out = append(out, e)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	if len(out) > 5 {
		out = out[:5]
	}
	return out
}

func normalizeInitials(s string) string {
	s = strings.TrimSpace(s)
	if s == "?" {
		return s
	}
	runes := make([]rune, 0, 3)
	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}
		runes = append(runes, unicode.ToUpper(r))
		if len(runes) == 3 {
			break
		}
	}
	if len(runes) == 0 {
		return ""
	}
	return string(runes)
}
