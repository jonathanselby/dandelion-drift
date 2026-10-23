package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"image"
	_ "image/png"
	"strings"
	"testing"
)

func TestEmbeddedAssetsDecode(t *testing.T) {
	for _, path := range []string{
		"assets/bunny.b64",
		"assets/bee.b64",
		"assets/flower.b64",
		"assets/frog.b64",
		"assets/thistle.b64",
	} {
		raw, err := assetFS.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		pngData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
		if err != nil {
			t.Fatalf("base64 %s: %v", path, err)
		}
		if _, _, err := image.Decode(bytes.NewReader(pngData)); err != nil {
			t.Fatalf("image %s: %v", path, err)
		}
	}

	raw, err := assetFS.ReadFile("assets/sounds.zip.b64")
	if err != nil {
		t.Fatal(err)
	}
	zipData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("sound bundle base64: %v", err)
	}
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		t.Fatalf("sound bundle zip: %v", err)
	}
	if len(zr.File) == 0 {
		t.Fatal("sound bundle is empty")
	}
}
