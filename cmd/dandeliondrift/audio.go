package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/jonathanselby/dandelion-drift/internal/game"
)

const soundSampleRate = 44100

type soundBank struct {
	context *audio.Context
	players map[string]*audio.Player
}

type cueSpec struct {
	frequency float64
	duration  float64
	volume    float64
}

var cueSpecs = map[string]cueSpec{
	"flap":       {frequency: 520, duration: 0.07, volume: 0.20},
	"perch":      {frequency: 720, duration: 0.10, volume: 0.18},
	"regrow":     {frequency: 880, duration: 0.09, volume: 0.16},
	"bee":        {frequency: 180, duration: 0.12, volume: 0.20},
	"thistle":    {frequency: 240, duration: 0.10, volume: 0.20},
	"bounce":     {frequency: 330, duration: 0.08, volume: 0.16},
	"splash":     {frequency: 210, duration: 0.16, volume: 0.18},
	"frog_leap":  {frequency: 300, duration: 0.11, volume: 0.18},
	"frog_catch": {frequency: 150, duration: 0.16, volume: 0.22},
	"puff":       {frequency: 460, duration: 0.18, volume: 0.14},
	"landing":    {frequency: 390, duration: 0.16, volume: 0.16},
	"highscore":  {frequency: 980, duration: 0.22, volume: 0.18},
	"firefly1":   {frequency: 660, duration: 0.08, volume: 0.16},
	"firefly2":   {frequency: 740, duration: 0.08, volume: 0.16},
	"firefly3":   {frequency: 830, duration: 0.08, volume: 0.16},
	"firefly4":   {frequency: 930, duration: 0.08, volume: 0.16},
	"firefly5":   {frequency: 1040, duration: 0.10, volume: 0.18},
}

func loadSoundBank() *soundBank {
	bank := &soundBank{players: map[string]*audio.Player{}}
	if !audioDeviceLikelyAvailable() {
		log.Printf("audio device unavailable; continuing silently")
		return bank
	}
	bank.context = audio.NewContext(soundSampleRate)
	for name := range cueSpecs {
		pcm, err := cuePCM(name)
		if err != nil {
			log.Printf("sound %s unavailable: %v", name, err)
			continue
		}
		bank.players[name] = bank.context.NewPlayerFromBytes(pcm)
	}
	return bank
}

func audioDeviceLikelyAvailable() bool {
	if runtime.GOOS != "linux" {
		return true
	}
	if os.Getenv("PULSE_SERVER") != "" {
		return true
	}
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		if _, err := os.Stat(filepath.Join(runtimeDir, "pulse", "native")); err == nil {
			return true
		}
	}
	if _, err := os.Stat("/dev/snd"); err == nil {
		return true
	}
	return false
}

func cuePCM(name string) ([]byte, error) {
	spec, ok := cueSpecs[name]
	if !ok {
		return nil, fmt.Errorf("unknown cue %q", name)
	}
	frames := int(float64(soundSampleRate) * spec.duration)
	if frames < 1 {
		return nil, fmt.Errorf("invalid cue duration")
	}
	pcm := make([]byte, frames*4)
	for i := 0; i < frames; i++ {
		t := float64(i) / float64(soundSampleRate)
		progress := float64(i) / float64(frames)
		envelope := math.Sin(math.Pi * progress)
		wave := math.Sin(2 * math.Pi * spec.frequency * t)
		wave += 0.22 * math.Sin(4*math.Pi*spec.frequency*t)
		sample := int16(math.Max(-1, math.Min(1, wave*envelope*spec.volume)) * 32767)
		binary.LittleEndian.PutUint16(pcm[i*4:i*4+2], uint16(sample))
		binary.LittleEndian.PutUint16(pcm[i*4+2:i*4+4], uint16(sample))
	}
	return pcm, nil
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
	b.playName(name)
}
