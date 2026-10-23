package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jonathanselby/dandelion-drift/internal/game"
)

type effectParticle struct {
	x, y, vx, vy  float64
	life, maxLife int
	size          float64
	colour        color.RGBA
}

func newDemoGame() *game.Game {
	g := game.New(4242)
	g.StartFlight()
	return g
}

func (a *app) updateDemo() {
	if a.demo == nil || a.demo.Screen != game.Play || a.demo.Bunny.State == game.Landed {
		a.demo = newDemoGame()
		a.demoFrame = 0
	}
	a.demoFrame++
	in := game.Input{
		Flap:  a.demoFrame%34 == 0,
		Right: a.demoFrame%240 < 65,
		Left:  a.demoFrame%240 >= 145 && a.demoFrame%240 < 190,
	}
	a.demo.Update(in)
}

func (a *app) reactToEvent(ev game.Event) {
	switch ev.Type {
	case game.EvFlap:
		a.spawnReleasedSeed()
	case game.EvPuffGone:
		a.spawnPuffRelease()
	case game.EvPerch:
		a.spawnBurst(a.game.Bunny.SX, a.game.Bunny.Y+22, color.RGBA{246, 180, 192, 220}, 7, 1.8)
	case game.EvRegrow:
		a.spawnBurst(a.game.Bunny.SX, a.game.Bunny.Y-72, color.RGBA{255, 249, 184, 235}, 6, 1.5)
	case game.EvSplash:
		a.spawnBurst(a.game.Bunny.SX, game.GroundY-4, color.RGBA{196, 239, 250, 235}, 11, 2.4)
	case game.EvFirefly:
		a.chainPop = ev.Value
		a.chainPopTicks = 42
		a.spawnBurst(a.game.Bunny.SX, a.game.Bunny.Y, color.RGBA{255, 237, 115, 235}, 8, 1.8)
	}
}

func (a *app) spawnBurst(x, y float64, c color.RGBA, count int, speed float64) {
	for i := 0; i < count && len(a.effects) < 256; i++ {
		angle := 2 * math.Pi * float64(i) / float64(max(1, count))
		v := speed * (0.65 + 0.35*float64((i%3)+1)/3)
		a.effects = append(a.effects, effectParticle{
			x: x, y: y,
			vx:   math.Cos(angle) * v,
			vy:   math.Sin(angle)*v - 0.4,
			life: 42 + i%12, maxLife: 54,
			size:   2.2 + float64(i%3)*0.6,
			colour: c,
		})
	}
}

func (a *app) updateEffects() {
	out := a.effects[:0]
	for _, p := range a.effects {
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.035
		p.vx *= 0.985
		p.life--
		if p.life > 0 {
			out = append(out, p)
		}
	}
	a.effects = out
	if a.chainPopTicks > 0 {
		a.chainPopTicks--
	}
}

func (a *app) drawEffects(screen *ebiten.Image) {
	for _, p := range a.effects {
		c := p.colour
		c.A = uint8(int(c.A) * p.life / max(1, p.maxLife))
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), float32(p.size), c, true)
	}
	if a.chainPopTicks > 0 && a.chainPop > 1 {
		y := a.game.Bunny.Y - 105 - float64(42-a.chainPopTicks)*0.5
		drawStyledCentered(screen, a.headingFace, fmtChain(a.chainPop), int(a.game.Bunny.SX), int(y), color.RGBA{142, 103, 34, 245}, color.RGBA{255, 250, 196, 170}, color.RGBA{})
	}
}

func fmtChain(n int) string {
	return "×" + string(rune('0'+n))
}

func (a *app) drawDemoBunny(screen *ebiten.Image) {
	if a.demo == nil {
		return
	}
	b := a.demo.Bunny
	x := 650 + (b.SX-260)*0.55
	y := 350 + (b.Y-340)*0.32
	frame := 0
	if b.State == game.Perched {
		frame = 2
	} else if b.VY < -0.5 {
		frame = 1
	}
	a.drawBunnySprite(screen, frame, x, y, 0.85)
	if a.demo.Seeds > 0 {
		a.drawPuff(screen, x, y-70, a.demo.Seeds, 0.70)
		ebitenutil.DrawLine(screen, x, y-51, x-5, y-17, color.RGBA{84, 112, 56, 220})
	}
}
