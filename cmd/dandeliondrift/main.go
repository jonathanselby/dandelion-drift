package main

import (
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
	game *game.Game

	bunny   [4]*ebiten.Image
	bee     [2]*ebiten.Image
	frog    [2]*ebiten.Image
	thistle *ebiten.Image
	flower  *ebiten.Image

	titleFace   font.Face
	headingFace font.Face
	labelFace   font.Face

	particles []seedParticle
	frame     int
}

func main() {
	a := &app{game: game.New(time.Now().UnixNano())}
	if err := a.loadAssets(); err != nil {
		log.Fatal(err)
	}
	if err := a.loadFonts(); err != nil {
		log.Fatal(err)
	}

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

	in := game.Input{
		Flap:  inpututil.IsKeyJustPressed(ebiten.KeySpace),
		Left:  ebiten.IsKeyPressed(ebiten.KeyArrowLeft),
		Right: ebiten.IsKeyPressed(ebiten.KeyArrowRight),
	}

	beforeScreen := a.game.Screen
	a.game.Update(in)
	for _, ev := range a.game.Events {
		switch ev.Type {
		case game.EvFlap:
			a.spawnReleasedSeed()
		case game.EvPuffGone:
			a.spawnPuffRelease()
		}
	}
	if beforeScreen != game.Attract && a.game.Screen == game.Attract {
		a.particles = a.particles[:0]
	}
	a.updateParticles()
	return nil
}

func (a *app) Draw(screen *ebiten.Image) {
	a.drawMeadow(screen)

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
	pngData, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(raw)))
	if err != nil {
		return nil, fmt.Errorf("decode base64 %s: %w", path, err)
	}
	img, _, err := image.Decode(strings.NewReader(string(pngData)))
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
