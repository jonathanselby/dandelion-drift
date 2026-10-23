package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jonathanselby/dandelion-drift/internal/game"
	"golang.org/x/image/font"
)

func (a *app) drawAttract(screen *ebiten.Image) {
	green := color.RGBA{43, 104, 71, 255}
	dark := color.RGBA{34, 55, 43, 235}
	cream := color.RGBA{255, 251, 230, 255}
	pink := color.RGBA{239, 158, 177, 255}

	// Title: restrained sophistication from typography, shadow and the game art itself.
	drawStyledCentered(screen, a.titleFace, "DANDELION", game.LogicalWidth/2, 105, cream, color.RGBA{97, 128, 92, 220}, dark)
	drawStyledCentered(screen, a.titleFace, "DRIFT", game.LogicalWidth/2, 171, pink, color.RGBA{97, 128, 92, 210}, dark)

	bob := math.Sin(float64(a.frame)/42) * 8
	a.drawBunnySprite(screen, 0, 480, 360+bob, 1.35)
	a.drawPuff(screen, 480, 260+bob, game.MaxSeeds, 1.25)
	ebitenutil.DrawLine(screen, 481, 284+bob, 468, 337+bob, color.RGBA{89, 118, 58, 240})

	// Clear child-facing start affordance; no menu system.
	drawRoundedPanel(screen, 300, 472, 360, 92, color.RGBA{255, 252, 232, 232}, color.RGBA{74, 132, 83, 255})
	drawStyledCentered(screen, a.headingFace, "PRESS SPACE", game.LogicalWidth/2, 515, green, color.RGBA{255, 255, 255, 160}, color.RGBA{255, 255, 255, 0})
	drawStyledCentered(screen, a.labelFace, "TO FLOAT", game.LogicalWidth/2, 548, color.RGBA{75, 94, 76, 255}, color.RGBA{}, color.RGBA{})

	drawStyledCentered(screen, a.labelFace, "SPACE  FLAP     <-  ->  DRIFT", game.LogicalWidth/2, 608, color.RGBA{52, 84, 62, 255}, color.RGBA{}, color.RGBA{})
	drawStyledCentered(screen, a.labelFace, "Every flap spends your wings", game.LogicalWidth/2, 678, color.RGBA{57, 91, 67, 220}, color.RGBA{}, color.RGBA{})
}

func (a *app) drawPlay(screen *ebiten.Image) {
	b := a.game.Bunny
	a.drawFlowers(screen)
	a.drawPlanters(screen)

	frame := 0
	if b.State == game.Perched {
		frame = 2
	} else if b.State == game.Falling {
		frame = 0
	} else if b.State == game.Landed {
		frame = 3
	} else if b.VY < -0.5 {
		frame = 1
	}

	a.drawBunnySprite(screen, frame, b.SX, b.Y, 0.92)
	if a.game.Seeds > 0 {
		a.drawPuff(screen, b.SX, b.Y-76, a.game.Seeds, 0.78)
		ebitenutil.DrawLine(screen, b.SX, b.Y-56, b.SX-5, b.Y-18, color.RGBA{84, 112, 56, 235})
	} else {
		ebitenutil.DrawLine(screen, b.SX, b.Y-58, b.SX-4, b.Y-18, color.RGBA{84, 112, 56, 235})
		vector.DrawFilledCircle(screen, float32(b.SX), float32(b.Y-59), 5, color.RGBA{151, 129, 73, 255}, true)
	}

	for _, p := range a.particles {
		alpha := uint8(220 * p.life / max(1, p.maxLife))
		clr := color.RGBA{249, 248, 231, alpha}
		ebitenutil.DrawLine(screen, p.x, p.y, p.x+math.Cos(p.rotation)*8, p.y+math.Sin(p.rotation)*8, clr)
		vector.DrawFilledCircle(screen, float32(p.x), float32(p.y), 2.2, clr, true)
	}

	a.drawHUD(screen)
	if b.State == game.Landed {
		a.drawEndScreen(screen)
	}
}

func (a *app) drawHUD(screen *ebiten.Image) {
	dark := color.RGBA{46, 73, 54, 220}
	cream := color.RGBA{255, 252, 233, 235}
	drawRoundedPanel(screen, 22, 20, 196, 52, cream, color.RGBA{94, 143, 93, 210})
	drawStyledCentered(screen, a.labelFace, fmt.Sprintf("%04d m", a.game.Metres), 120, 53, dark, color.RGBA{}, color.RGBA{})

	// The puff is the health display, but this tiny seed row makes development/UAT state visible
	// without turning into a second health system.
	for i := 0; i < game.MaxSeeds; i++ {
		c := color.RGBA{246, 245, 223, 80}
		if i < a.game.Seeds {
			c = color.RGBA{255, 255, 244, 225}
		}
		vector.DrawFilledCircle(screen, float32(728+i*8), 45, 2.5, c, true)
	}
}

func (a *app) drawEndScreen(screen *ebiten.Image) {
	cream := color.RGBA{255, 251, 231, 248}
	green := color.RGBA{55, 109, 71, 255}
	muted := color.RGBA{82, 105, 87, 255}
	pink := color.RGBA{232, 150, 172, 255}

	// Preserve the finished meadow behind the result; the run should feel concluded, not erased.
	drawRoundedPanel(screen, 185, 148, 590, 385, cream, color.RGBA{75, 129, 83, 255})

	// Decorative drifting seeds make this a first-class scene rather than a text modal.
	for i := 0; i < 9; i++ {
		a := float64(i)*0.72 + float64(a.frame)/90
		x := 242 + math.Mod(float64(i*71)+math.Cos(a)*18, 470)
		y := 190 + float64((i*43)%270) + math.Sin(a)*10
		ebitenutil.DrawLine(screen, x, y, x+8, y-5, color.RGBA{173, 164, 128, 145})
		vector.DrawFilledCircle(screen, float32(x+9), float32(y-6), 2, color.RGBA{245, 243, 220, 190}, true)
	}

	drawStyledCentered(screen, a.headingFace, "A GENTLE LANDING", game.LogicalWidth/2, 220, green, color.RGBA{255, 255, 255, 210}, color.RGBA{})
	drawStyledCentered(screen, a.titleFace, fmt.Sprintf("%d m", a.game.Metres), game.LogicalWidth/2, 310, pink, color.RGBA{119, 139, 100, 150}, color.RGBA{})
	drawStyledCentered(screen, a.labelFace, fmt.Sprintf("%d seeds carried on the wind", a.game.Released), game.LogicalWidth/2, 360, muted, color.RGBA{}, color.RGBA{})

	remaining := (a.game.EndTicks + game.TPS - 1) / game.TPS
	message := "Returning to the meadow..."
	if remaining <= 1 {
		message = "Ready for another drift"
	}
	drawStyledCentered(screen, a.labelFace, message, game.LogicalWidth/2, 446, muted, color.RGBA{}, color.RGBA{})

	for i := 0; i < 3; i++ {
		c := color.RGBA{195, 214, 174, 170}
		if i < (game.EndTicks/(game.TPS) + 1) {
			c = color.RGBA{84, 142, 91, 220}
		}
		vector.DrawFilledCircle(screen, float32(456+i*24), 486, 4, c, true)
	}
}

func (a *app) drawBunnySprite(screen *ebiten.Image, frame int, x, y, scale float64) {
	img := a.bunny[frame]
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-50, -50)
	op.GeoM.Scale(scale, scale)
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func (a *app) drawPuff(screen *ebiten.Image, x, y float64, seeds int, scale float64) {
	if seeds <= 0 {
		return
	}
	core := float32(7 * scale)
	outer := 31.0 * scale
	vector.DrawFilledCircle(screen, float32(x), float32(y), core, color.RGBA{198, 174, 118, 255}, true)
	for i := 0; i < seeds; i++ {
		angle := -math.Pi/2 + 2*math.Pi*float64(i)/float64(seeds)
		x1 := x + math.Cos(angle)*7*scale
		y1 := y + math.Sin(angle)*7*scale
		x2 := x + math.Cos(angle)*outer
		y2 := y + math.Sin(angle)*outer
		ebitenutil.DrawLine(screen, x1, y1, x2, y2, color.RGBA{244, 240, 216, 220})
		vector.DrawFilledCircle(screen, float32(x2), float32(y2), float32(3.0*scale), color.RGBA{255, 254, 239, 235}, true)
		// Two tiny hairs make each seed read as a dandelion tuft at gameplay scale.
		for _, d := range []float64{-0.38, 0.38} {
			a := angle + d
			ebitenutil.DrawLine(screen, x2, y2, x2+math.Cos(a)*7*scale, y2+math.Sin(a)*7*scale, color.RGBA{255, 254, 241, 180})
		}
	}
}

func (a *app) spawnReleasedSeed() {
	b := a.game.Bunny
	a.particles = append(a.particles, seedParticle{
		x: b.SX + 5, y: b.Y - 72,
		vx: 1.8, vy: -0.35,
		life: 180, maxLife: 180,
		rotation: -0.6,
	})
}

func (a *app) spawnPuffRelease() {
	b := a.game.Bunny
	for i := 0; i < 7; i++ {
		a.particles = append(a.particles, seedParticle{
			x: b.SX + float64(i-3)*3, y: b.Y - 75 + float64((i%3)-1)*3,
			vx: 1.2 + float64(i)*0.12, vy: -0.7 + float64(i%3)*0.2,
			life: 150, maxLife: 150,
			rotation: float64(i) * 0.5,
		})
	}
}

func (a *app) updateParticles() {
	out := a.particles[:0]
	for _, p := range a.particles {
		p.x += p.vx
		p.y += p.vy
		p.vy += 0.002
		p.rotation += 0.025
		p.life--
		if p.life > 0 && p.x < game.LogicalWidth+30 {
			out = append(out, p)
		}
	}
	a.particles = out
}

func drawRoundedPanel(screen *ebiten.Image, x, y, w, h float64, fill, border color.RGBA) {
	// Rounded enough for this game's soft visual language without introducing a UI toolkit.
	ebitenutil.DrawRect(screen, x+16, y, w-32, h, fill)
	ebitenutil.DrawRect(screen, x, y+16, w, h-32, fill)
	vector.DrawFilledCircle(screen, float32(x+16), float32(y+16), 16, fill, true)
	vector.DrawFilledCircle(screen, float32(x+w-16), float32(y+16), 16, fill, true)
	vector.DrawFilledCircle(screen, float32(x+16), float32(y+h-16), 16, fill, true)
	vector.DrawFilledCircle(screen, float32(x+w-16), float32(y+h-16), 16, fill, true)
	// Deliberately thin frame.
	ebitenutil.DrawLine(screen, x+16, y, x+w-16, y, border)
	ebitenutil.DrawLine(screen, x+16, y+h, x+w-16, y+h, border)
	ebitenutil.DrawLine(screen, x, y+16, x, y+h-16, border)
	ebitenutil.DrawLine(screen, x+w, y+16, x+w, y+h-16, border)
}

func drawStyledCentered(screen *ebiten.Image, face font.Face, value string, centerX, baselineY int, fill, glow, outline color.RGBA) {
	width := font.MeasureString(face, value).Round()
	x := centerX - width/2
	if glow.A != 0 {
		for _, off := range [][2]int{{-2, 0}, {2, 0}, {0, -2}, {0, 2}} {
			text.Draw(screen, value, face, x+off[0], baselineY+off[1], glow)
		}
	}
	if outline.A != 0 {
		for _, off := range [][2]int{{-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
			text.Draw(screen, value, face, x+off[0], baselineY+off[1], outline)
		}
	}
	text.Draw(screen, value, face, x, baselineY, fill)
}

func mix(a, b color.RGBA, t float64) color.RGBA {
	lerp := func(x, y uint8) uint8 { return uint8(float64(x) + (float64(y)-float64(x))*t) }
	return color.RGBA{lerp(a.R, b.R), lerp(a.G, b.G), lerp(a.B, b.B), 255}
}
