package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/jonathanselby/dandelion-drift/internal/game"
)

const soundSampleRate = 44100

type soundBank struct {
	context *audio.Context
	players map[string]*audio.Player
}

func loadSoundBank() *soundBank {
	bank := &soundBank{context: audio.NewContext(soundSampleRate), players: map[string]*audio.Player{}}
	archive, err := loadSoundArchive("assets/sounds.zip.b64")
	if err != nil {
		log.Printf("sound archive unavailable: %v", err)
		return bank
	}
	for _, name := range []string{"flap", "perch", "regrow", "bee", "thistle", "bounce", "splash", "frog_leap", "frog_catch", "puff", "landing", "highscore", "firefly1", "firefly2", "firefly3", "firefly4", "firefly5"} {
		wavData, ok := archive["sound_"+name+".wav"]
		if !ok {
			continue
		}
		pcm, err := monoWAVAsStereo(wavData)
		if err != nil {
			log.Printf("sound %s unavailable: %v", name, err)
			continue
		}
		bank.players[name] = bank.context.NewPlayerFromBytes(pcm)
	}
	return bank
}

func loadSoundArchive(path string) (map[string][]byte, error) {
	raw, err := assetFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	zipData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, f := range zr.File {
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		buf := new(bytes.Buffer)
		_, copyErr := buf.ReadFrom(r)
		closeErr := r.Close()
		if copyErr != nil {
			return nil, copyErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		out[f.Name] = buf.Bytes()
	}
	return out, nil
}

func monoWAVAsStereo(wavData []byte) ([]byte, error) {
	if len(wavData) < 12 || string(wavData[:4]) != "RIFF" || string(wavData[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not RIFF/WAVE")
	}
	var formatOK bool
	var mono []byte
	for off := 12; off+8 <= len(wavData); {
		id := string(wavData[off : off+4])
		n := int(binary.LittleEndian.Uint32(wavData[off+4 : off+8]))
		start := off + 8
		end := start + n
		if end > len(wavData) {
			return nil, fmt.Errorf("truncated %s chunk", id)
		}
		switch id {
		case "fmt ":
			if n < 16 {
				return nil, fmt.Errorf("short fmt chunk")
			}
			formatOK = binary.LittleEndian.Uint16(wavData[start:start+2]) == 1 && binary.LittleEndian.Uint16(wavData[start+2:start+4]) == 1 && binary.LittleEndian.Uint32(wavData[start+4:start+8]) == soundSampleRate && binary.LittleEndian.Uint16(wavData[start+14:start+16]) == 16
		case "data":
			mono = wavData[start:end]
		}
		off = end
		if off%2 != 0 {
			off++
		}
	}
	if !formatOK || len(mono) == 0 || len(mono)%2 != 0 {
		return nil, fmt.Errorf("must be 44.1 kHz mono PCM16")
	}
	stereo := make([]byte, 0, len(mono)*2)
	for i := 0; i < len(mono); i += 2 {
		stereo = append(stereo, mono[i], mono[i+1], mono[i], mono[i+1])
	}
	return stereo, nil
}

func (b *soundBank) playName(name string) {
	if b == nil {
		return
	}
	p := b.players[name]
	if p == nil {
		return
	}
	if err := p.Rewind(); err != nil {
		log.Printf("rewind sound %s: %v", name, err)
		return
	}
	p.Play()
}

func (b *soundBank) playEvent(ev game.Event) {
	if b == nil {
		return
	}
	name := ""
	switch ev.Type {
	case game.EvFlap, game.EvTakeOff:
		name = "flap"
	case game.EvPerch:
		name = "perch"
	case game.EvRegrow, game.EvSprouted:
		name = "regrow"
	case game.EvBee:
		name = "bee"
	case game.EvThistle:
		name = "thistle"
	case game.EvBounce:
		name = "bounce"
	case game.EvSplash:
		name = "splash"
	case game.EvFrogLeap:
		name = "frog_leap"
	case game.EvFrogCatch:
		name = "frog_catch"
	case game.EvPuffGone:
		name = "puff"
	case game.EvFlightOver:
		name = "landing"
	case game.EvFirefly:
		level := ev.Value
		if level < 1 {
			level = 1
		}
		if level > game.ChainMax {
			level = game.ChainMax
		}
		name = fmt.Sprintf("firefly%d", level)
	}
	p := b.players[name]
	if p == nil {
		return
	}
	if err := p.Rewind(); err != nil {
		log.Printf("rewind sound %s: %v", name, err)
		return
	}
	p.Play()
}
