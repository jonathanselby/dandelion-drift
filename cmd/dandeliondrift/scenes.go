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

	drawStyledCentered(screen, a.titleFace, "DANDELION", game.LogicalWidth/2, 96, cream, color.RGBA{97, 128, 92, 220}, dark)
	drawStyledCentered(screen, a.titleFace, "DRIFT", game.LogicalWidth/2, 160, pink, color.RGBA{97, 128, 92, 210}, dark)

	// High scores and the live demo share the meadow instead of creating a menu layer.
	drawRoundedPanel(screen, 72, 215, 340, 315, color.RGBA{255, 252, 232, 235}, color.RGBA{74, 132, 83, 235})
	drawStyledCentered(screen, a.headingFace, "TOP FIVE", 242, 260, green, color.RGBA{255, 255, 255, 150}, color.RGBA{})
	a.drawHighScores(screen, 112, 300)

	drawRoundedPanel(screen, 500, 215, 388, 315, color.RGBA{244, 250, 229, 150}, color.RGBA{104, 156, 104, 160})
	drawStyledCentered(screen, a.labelFace, "A LITTLE DRIFT", 694, 255, color.RGBA{69, 101, 72, 230}, color.RGBA{}, color.RGBA{})
	a.drawDemoBunny(screen)

	drawRoundedPanel(screen, 300, 555, 360, 82, color.RGBA{255, 252, 232, 238}, color.RGBA{74, 132, 83, 255})
	drawStyledCentered(screen, a.headingFace, "PRESS SPACE", game.LogicalWidth/2, 598, green, color.RGBA{255, 255, 255, 160}, color.RGBA{})
	drawStyledCentered(screen, a.labelFace, "TO FLOAT", game.LogicalWidth/2, 627, color.RGBA{75, 94, 76, 255}, color.RGBA{}, color.RGBA{})
	drawStyledCentered(screen, a.labelFace, "Every flap spends your wings", game.LogicalWidth/2, 682, color.RGBA{57, 91, 67, 220}, color.RGBA{}, color.RGBA{})
}

func (a *app) drawHighScores(screen *ebiten.Image, x, y int) {
	entries := a.scores.Normalized().Scores
	for i := 0; i < 5; i++ {
		initials := "---"
		score := "-----"
		if i < len(entries) {
			initials = entries[i].Initials
			score = fmt.Sprintf("%05d", entries[i].Score)
		}
		lineY := y + i*43
		text.Draw(screen, fmt.Sprintf("%d", i+1), a.labelFace, x, lineY, color.RGBA{92, 120, 88, 220})
		text.Draw(screen, initials, a.headingFace, x+42, lineY+2, color.RGBA{55, 91, 66, 245})
		w := font.MeasureString(a.headingFace, score).Round()
		text.Draw(screen, score, a.headingFace, x+260-w, lineY+2, color.RGBA{118, 93, 49, 240})
	}
}

func (a *app) drawInitials(screen *ebiten.Image) {
	cream := color.RGBA{255, 251, 231, 248}
	green := color.RGBA{55, 109, 71, 255}
	muted := color.RGBA{82, 105, 87, 255}
	pink := color.RGBA{232, 150, 172, 255}

	drawRoundedPanel(screen, 190, 125, 580, 460, cream, color.RGBA{75, 129, 83, 255})
	drawStyledCentered(screen, a.headingFace, "NEW TOP FIVE", game.LogicalWidth/2, 200, green, color.RGBA{255, 255, 255, 190}, color.RGBA{})
	drawStyledCentered(screen, a.titleFace, fmt.Sprintf("%d", a.game.Score), game.LogicalWidth/2, 285, pink, color.RGBA{119, 139, 100, 140}, color.RGBA{})
	drawStyledCentered(screen, a.labelFace, "TYPE 1-3 LETTERS", game.LogicalWidth/2, 345, muted, color.RGBA{}, color.RGBA{})

	runes := []rune(a.initials)
	for i := 0; i < 3; i++ {
		x := 385 + i*95
		display := "_"
		if i < len(runes) {
			display = string(runes[i])
		}
		drawStyledCentered(screen, a.titleFace, display, x, 430, green, color.RGBA{255, 255, 255, 150}, color.RGBA{})
	}
	drawStyledCentered(screen, a.labelFace, "ENTER TO SAVE   •   BACKSPACE TO CHANGE", game.LogicalWidth/2, 500, muted, color.RGBA{}, color.RGBA{})
	drawStyledCentered(screen, a.labelFace, "Empty initials save as ?", game.LogicalWidth/2, 540, color.RGBA{112, 122, 104, 220}, color.RGBA{}, color.RGBA{})
	a.drawEffects(screen)
}

func (a *app) drawPlay(screen *ebiten.Image) {
	b := a.game.Bunny
	a.drawWorldPressure(screen)
	a.drawFlowers(screen)
	a.drawPlanters(screen)
	a.drawFireflies(screen)
	a.drawHazardSprites(screen)

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

	if b.State == game.Landed && a.landedOverWater() {
		vector.DrawFilledCircle(screen, float32(b.SX), float32(b.Y+27), 34, color.RGBA{87, 154, 82, 245}, true)
		vector.DrawFilledCircle(screen, float32(b.SX+18), float32(b.Y+20), 9, color.RGBA{103, 181, 96, 245}, true)
	}
	blinkVisible := b.Guard == 0 || (a.frame/4)%2 == 0
	if blinkVisible {
		a.drawBunnySprite(screen, frame, b.SX, b.Y, 0.92)
	}
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
	a.drawEffects(screen)

	a.drawHUD(screen)
	if b.State == game.Landed {
		a.drawEndScreen(screen)
	}
}

func (a *app) drawHUD(screen *ebiten.Image) {
	dark := color.RGBA{46, 73, 54, 220}
	cream := color.RGBA{255, 252, 233, 235}
	drawRoundedPanel(screen, 22, 20, 210, 58, cream, color.RGBA{94, 143, 93, 210})
	drawStyledCentered(screen, a.labelFace, fmt.Sprintf("%04d m   %05d", a.game.Metres, a.game.Score), 127, 55, dark, color.RGBA{}, color.RGBA{})

	if a.game.Chain > 1 && a.game.ChainTicks > 0 {
		drawRoundedPanel(screen, 420, 20, 120, 58, color.RGBA{255, 249, 210, 225}, color.RGBA{198, 164, 75, 220})
		drawStyledCentered(screen, a.headingFace, fmt.Sprintf("x%d", a.game.Chain), 480, 58, color.RGBA{126, 94, 36, 255}, color.RGBA{}, color.RGBA{})
	}

	for i := 0; i < game.MaxSeeds; i++ {
		c := color.RGBA{246, 245, 223, 80}
		if i < a.game.Seeds {
			c = color.RGBA{255, 255, 244, 225}
		}
		vector.DrawFilledCircle(screen, float32(728+i*8), 45, 2.5, c, true)
	}
}

func (a *app) landedOverWater() bool {
	wx := a.game.Camera + a.game.Bunny.SX
	for _, p := range a.game.Ponds {
		if wx >= p.X && wx <= p.X+p.Width {
			return true
		}
	}
	return false
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
	drawStyledCentered(screen, a.titleFace, fmt.Sprintf("%d", a.game.Score), game.LogicalWidth/2, 302, pink, color.RGBA{119, 139, 100, 150}, color.RGBA{})
	drawStyledCentered(screen, a.labelFace, fmt.Sprintf("%d m   •   %d fireflies", a.game.Metres, a.game.FirefliesCollected), game.LogicalWidth/2, 348, muted, color.RGBA{}, color.RGBA{})
	drawStyledCentered(screen, a.labelFace, fmt.Sprintf("%d seeds carried on the wind", a.game.Released), game.LogicalWidth/2, 380, muted, color.RGBA{}, color.RGBA{})

	remaining := (a.game.EndTicks + game.TPS - 1) / game.TPS
	message := "Returning to the meadow..."
	if a.scores.Qualifies(a.game.Score) {
		message = "A top-five flight — initials next"
	} else if remaining <= 1 {
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
