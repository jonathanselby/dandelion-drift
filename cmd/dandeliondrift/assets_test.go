package main

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"
)

func TestEmbeddedAssetsDecode(t *testing.T) {
	for _, path := range []string{
		"assets/bunny.png",
		"assets/bee.png",
		"assets/flower.png",
		"assets/frog.png",
		"assets/thistle.png",
	} {
		raw, err := assetFS.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if _, _, err := image.Decode(bytes.NewReader(raw)); err != nil {
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
