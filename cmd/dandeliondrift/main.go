package main

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"log"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/jonathanselby/dandelion-drift/internal/game"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
)

//go:embed assets/*.b64
var assetFS embed.FS

type seedParticle struct {
	x, y     float64
	vx, vy   float64
	life     int
	maxLife  int
	rotation float64
}

type app struct {
	game   *game.Game
	demo   *game.Game
	scores game.HighScores

	bunny   [4]*ebiten.Image
	bee     [2]*ebiten.Image
	frog    [2]*ebiten.Image
	thistle *ebiten.Image
	flower  *ebiten.Image

	titleFace   font.Face
	headingFace font.Face
	labelFace   font.Face

	particles               []seedParticle
	effects                 []effectParticle
	sounds                  *soundBank
	scoresPath              string
	initialsMode            bool
	initials                string
	chainPop, chainPopTicks int
	demoFrame               int
	frame                   int
}

func main() {
	const scoresPath = "scores.json"
	scores, err := game.LoadScores(scoresPath)
	if err != nil {
		log.Printf("high scores unavailable, starting empty: %v", err)
	}
	g := game.New(time.Now().UnixNano())
	a := &app{game: g, demo: newDemoGame(), scores: scores, scoresPath: scoresPath}
	if err := a.loadAssets(); err != nil {
		log.Fatal(err)
	}
	if err := a.loadFonts(); err != nil {
		log.Fatal(err)
	}
	a.sounds = loadSoundBank()

	ebiten.SetWindowSize(game.LogicalWidth, game.LogicalHeight)
	ebiten.SetWindowResizable(false)
	ebiten.SetWindowTitle("Dandelion Drift")
	ebiten.SetFullscreen(true)
	if err := ebiten.RunGame(a); err != nil {
		log.Fatal(err)
	}
}

func (a *app) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	a.frame++

	if a.initialsMode {
		a.updateInitialsInput()
		a.updateEffects()
		return nil
	}

	in := game.Input{
		Flap:  inpututil.IsKeyJustPressed(ebiten.KeySpace),
		Left:  ebiten.IsKeyPressed(ebiten.KeyArrowLeft),
		Right: ebiten.IsKeyPressed(ebiten.KeyArrowRight),
	}

	beforeScreen := a.game.Screen
	a.game.Update(in)
	for _, ev := range a.game.Events {
		a.sounds.playEvent(ev)
		a.reactToEvent(ev)
	}

	if beforeScreen == game.Play && a.game.Screen == game.Attract && a.scores.Qualifies(a.game.Score) {
		a.initialsMode = true
		a.initials = ""
		a.sounds.playName("highscore")
		a.spawnBurst(game.LogicalWidth/2, 300, color.RGBA{255, 224, 118, 235}, 18, 2.6)
	}
	if a.game.Screen == game.Attract && !a.initialsMode {
		a.updateDemo()
	}
	if beforeScreen != game.Attract && a.game.Screen == game.Attract && !a.initialsMode {
		a.particles = a.particles[:0]
		a.effects = a.effects[:0]
	}
	a.updateParticles()
	a.updateEffects()
	return nil
}

func (a *app) updateInitialsInput() {
	runes := []rune(a.initials)
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(runes) > 0 {
		runes = runes[:len(runes)-1]
	}
	for _, r := range ebiten.AppendInputChars(nil) {
		if len(runes) >= 3 {
			break
		}
		if r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		if r < 'A' || r > 'Z' {
			continue
		}
		runes = append(runes, r)
	}
	a.initials = string(runes)
	if !inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		return
	}
	initials := a.initials
	if initials == "" {
		initials = "?"
	}
	a.scores.Add(game.ScoreEntry{Initials: initials, Score: a.game.Score})
	if err := game.SaveScores(a.scoresPath, a.scores); err != nil {
		log.Printf("save high scores: %v", err)
	}
	a.initialsMode = false
	a.initials = ""
	a.particles = a.particles[:0]
	a.effects = a.effects[:0]
	a.demo = newDemoGame()
	a.demoFrame = 0
}

func (a *app) Draw(screen *ebiten.Image) {
	a.drawMeadow(screen)
	if a.initialsMode {
		a.drawInitials(screen)
		return
	}
	switch a.game.Screen {
	case game.Attract:
		a.drawAttract(screen)
	case game.Play:
		a.drawPlay(screen)
	}
}

func (a *app) Layout(_, _ int) (int, int) {
	return game.LogicalWidth, game.LogicalHeight
}

func (a *app) DrawFinalScreen(screen ebiten.FinalScreen, offscreen *ebiten.Image, geoM ebiten.GeoM) {
	screen.Fill(color.Black)
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM = geoM
	screen.DrawImage(offscreen, op)
}

func (a *app) loadAssets() error {
	sheet, err := loadImage("assets/bunny.b64")
	if err != nil {
		return err
	}
	for i := range a.bunny {
		r := image.Rect(i*100, 0, (i+1)*100, 100)
		a.bunny[i] = sheet.SubImage(r).(*ebiten.Image)
	}
	beeSheet, err := loadImage("assets/bee.b64")
	if err != nil {
		return err
	}
	for i := range a.bee {
		a.bee[i] = beeSheet.SubImage(image.Rect(i*64, 0, (i+1)*64, 64)).(*ebiten.Image)
	}
	frogSheet, err := loadImage("assets/frog.b64")
	if err != nil {
		return err
	}
	for i := range a.frog {
		a.frog[i] = frogSheet.SubImage(image.Rect(i*64, 0, (i+1)*64, 64)).(*ebiten.Image)
	}
	if a.thistle, err = loadImage("assets/thistle.b64"); err != nil {
		return err
	}
	if a.flower, err = loadImage("assets/flower.b64"); err != nil {
		return err
	}
	return nil
}

func loadImage(path string) (*ebiten.Image, error) {
	raw, err := assetFS.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if img, _, directErr := image.Decode(bytes.NewReader(raw)); directErr == nil {
		return ebiten.NewImageFromImage(img), nil
	}
	pngData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("decode base64 %s: %w", path, err)
	}
	img, _, err := image.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", path, err)
	}
	return ebiten.NewImageFromImage(img), nil
}

func (a *app) loadFonts() error {
	parsed, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return fmt.Errorf("parse built-in font: %w", err)
	}
	newFace := func(size float64) (font.Face, error) {
		return opentype.NewFace(parsed, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	}
	if a.titleFace, err = newFace(64); err != nil {
		return err
	}
	if a.headingFace, err = newFace(28); err != nil {
		return err
	}
	if a.labelFace, err = newFace(18); err != nil {
		return err
	}
	return nil
}
