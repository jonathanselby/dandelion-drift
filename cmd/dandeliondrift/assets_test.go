package main

import (
	"bytes"
	"encoding/base64"
	"image"
	_ "image/png"
	"strings"
	"testing"
)

func TestEmbeddedAssetsDecode(t *testing.T) {
	rawBunny, err := base64.StdEncoding.DecodeString(bunnyPNGBase64)
	if err != nil {
		t.Fatalf("embedded bunny base64: %v", err)
	}
	if _, _, err := image.Decode(bytes.NewReader(rawBunny)); err != nil {
		t.Fatalf("embedded bunny image: %v", err)
	}

	for _, path := range []string{
		"assets/bee.b64",
		"assets/flower.b64",
		"assets/frog.b64",
		"assets/thistle.b64",
	} {
		raw, err := assetFS.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if _, _, err := image.Decode(bytes.NewReader(raw)); err == nil {
			continue
		}
		pngData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
		if err != nil {
			t.Fatalf("base64 %s: %v", path, err)
		}
		if _, _, err := image.Decode(bytes.NewReader(pngData)); err != nil {
			t.Fatalf("image %s: %v", path, err)
		}
	}
}

func TestProceduralSoundCues(t *testing.T) {
	for name := range cueSpecs {
		pcm, err := cuePCM(name)
		if err != nil {
			t.Fatalf("cue %s: %v", name, err)
		}
		if len(pcm) == 0 || len(pcm)%4 != 0 {
			t.Fatalf("cue %s produced invalid stereo PCM length %d", name, len(pcm))
		}
	}
}
