package game

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestScoresSortedAndFiveOnly(t *testing.T) {
	h := HighScores{}
	for _, s := range []int{20, 100, 40, 80, 60, 120} {
		h.Add(ScoreEntry{Initials: "js", Score: s})
	}
	want := []int{120, 100, 80, 60, 40}
	if len(h.Scores) != len(want) {
		t.Fatalf("entries=%d", len(h.Scores))
	}
	for i, score := range want {
		if h.Scores[i].Score != score {
			t.Fatalf("scores[%d]=%d", i, h.Scores[i].Score)
		}
		if h.Scores[i].Initials != "JS" {
			t.Fatalf("initials=%q", h.Scores[i].Initials)
		}
	}
}

func TestScoreQualification(t *testing.T) {
	h := HighScores{Scores: []ScoreEntry{
		{"AAA", 100}, {"BBB", 90}, {"CCC", 80}, {"DDD", 70}, {"EEE", 60},
	}}
	if !h.Qualifies(61) {
		t.Fatal("61 should qualify")
	}
	if h.Qualifies(60) {
		t.Fatal("tie with fifth should not qualify")
	}
	if !(HighScores{}).Qualifies(0) {
		t.Fatal("empty table should accept a score")
	}
}

func TestScoreSaveLoadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	want := HighScores{Scores: []ScoreEntry{{"AAA", 100}, {"BB", 50}, {"?", 25}}}
	if err := SaveScores(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := LoadScores(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
}

func TestMissingScoresStartEmpty(t *testing.T) {
	got, err := LoadScores(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Scores) != 0 {
		t.Fatal("missing score file did not start empty")
	}
}

func TestMalformedScoresReportAndStartEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	if err := os.WriteFile(path, []byte("not-json"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadScores(path)
	if err == nil {
		t.Fatal("malformed score file should report an error")
	}
	if len(got.Scores) != 0 {
		t.Fatal("malformed score file should return empty scores")
	}
}

func TestInitialsAcceptOneToThreeLettersAndQuestionMark(t *testing.T) {
	h := HighScores{}
	h.Add(ScoreEntry{Initials: "j", Score: 10})
	h.Add(ScoreEntry{Initials: "on", Score: 20})
	h.Add(ScoreEntry{Initials: "selby", Score: 30})
	h.Add(ScoreEntry{Initials: "?", Score: 40})
	got := []string{h.Scores[0].Initials, h.Scores[1].Initials, h.Scores[2].Initials, h.Scores[3].Initials}
	want := []string{"?", "SEL", "ON", "J"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
