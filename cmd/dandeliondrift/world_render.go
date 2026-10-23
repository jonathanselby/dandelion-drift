package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/jonathanselby/dandelion-drift/internal/game"
)

func (a *app) drawMeadow(screen *ebiten.Image) {
	for y := 0; y < game.LogicalHeight; y += 4 {
		t := float64(y) / game.LogicalHeight
		c := mix(color.RGBA{164, 222, 246, 255}, color.RGBA{241, 245, 214, 255}, t)
		ebitenutil.DrawRect(screen, 0, float64(y), game.LogicalWidth, 4, c)
	}

	cloudX := func(base float64, rate float64) float64 {
		v := math.Mod(base+float64(a.frame)*rate, game.LogicalWidth+240)
		return v - 120
	}
	drawCloud(screen, cloudX(80, 0.08), 110, 1.0)
	drawCloud(screen, cloudX(540, 0.05), 185, 0.75)
	drawCloud(screen, cloudX(820, 0.065), 82, 0.55)

	// Two calm parallax hill bands made from deliberately simple overlapping circles.
	hillFar := color.RGBA{151, 196, 139, 255}
	hillNear := color.RGBA{102, 166, 104, 255}
	for i := -1; i < 7; i++ {
		x := float32(i*190) - float32(math.Mod(a.game.Camera*0.08, 190))
		vector.DrawFilledCircle(screen, x, 615, 165, hillFar, true)
	}
	for i := -1; i < 8; i++ {
		x := float32(i*155) - float32(math.Mod(a.game.Camera*0.16, 155))
		vector.DrawFilledCircle(screen, x, 655, 135, hillNear, true)
	}

	ebitenutil.DrawRect(screen, 0, game.GroundY, game.LogicalWidth, game.LogicalHeight-game.GroundY, color.RGBA{75, 143, 76, 255})
	for i := 0; i < 42; i++ {
		x := float64((i*47)%game.LogicalWidth) - math.Mod(a.game.Camera*0.55, 47)
		h := 7.0 + float64((i*13)%9)
		ebitenutil.DrawLine(screen, x, game.GroundY+2, x+2, game.GroundY-h, color.RGBA{47, 112, 58, 210})
	}

}

func drawCloud(screen *ebiten.Image, x, y, scale float64) {
	c := color.RGBA{255, 255, 255, 176}
	vector.DrawFilledCircle(screen, float32(x), float32(y), float32(30*scale), c, true)
	vector.DrawFilledCircle(screen, float32(x+34*scale), float32(y-8*scale), float32(38*scale), c, true)
	vector.DrawFilledCircle(screen, float32(x+76*scale), float32(y), float32(29*scale), c, true)
	ebitenutil.DrawRect(screen, x-2*scale, y, 82*scale, 25*scale, c)
}

func (a *app) drawFlowers(screen *ebiten.Image) {
	for _, f := range a.game.Flowers {
		sx := f.X - a.game.Camera
		if sx < -70 || sx > game.LogicalWidth+70 {
			continue
		}

		if !f.Bloomed {
			progress := 1 - float64(f.GrowTicks)/float64(game.SproutTicks)
			if progress < 0 {
				progress = 0
			}
			if progress > 1 {
				progress = 1
			}
			height := f.Height * (0.12 + 0.88*progress)
			top := game.GroundY - height
			ebitenutil.DrawLine(screen, sx, game.GroundY, sx, top, color.RGBA{55, 126, 70, 235})
			vector.DrawFilledCircle(screen, float32(sx-4), float32(top+8), 4, color.RGBA{93, 161, 86, 235}, true)
			vector.DrawFilledCircle(screen, float32(sx+4), float32(top+8), 4, color.RGBA{93, 161, 86, 235}, true)
			continue
		}

		headY := game.GroundY - f.Height
		stem := color.RGBA{55, 126, 70, 245}
		ebitenutil.DrawLine(screen, sx, game.GroundY, sx, headY+8, stem)
		ebitenutil.DrawLine(screen, sx, game.GroundY, sx+1, headY+8, stem)

		if f.Nectar <= 0 {
			headY += 7
		}
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Translate(-32, -32)
		op.GeoM.Scale(0.84, 0.84)
		if f.Nectar <= 0 {
			op.ColorScale.Scale(0.72, 0.72, 0.72, 0.86)
		}
		op.GeoM.Translate(sx, headY)
		screen.DrawImage(a.flower, op)
	}
}

func (a *app) drawPlanters(screen *ebiten.Image) {
	for _, p := range a.game.Planters {
		sx := p.X - a.game.Camera
		if sx < -20 || sx > game.LogicalWidth+20 {
			continue
		}
		clr := color.RGBA{250, 248, 228, 225}
		ebitenutil.DrawLine(screen, sx-7, p.Y+4, sx+2, p.Y-3, clr)
		vector.DrawFilledCircle(screen, float32(sx+3), float32(p.Y-4), 2.5, clr, true)
		ebitenutil.DrawLine(screen, sx+3, p.Y-4, sx+8, p.Y-11, color.RGBA{255, 254, 238, 175})
		ebitenutil.DrawLine(screen, sx+3, p.Y-4, sx+11, p.Y-5, color.RGBA{255, 254, 238, 175})
	}
}

func (a *app) drawWorldPressure(screen *ebiten.Image) {
	for _, p := range a.game.Ponds {
		sx := p.X - a.game.Camera
		if sx+p.Width < 0 || sx > game.LogicalWidth {
			continue
		}
		x := math.Max(0, sx)
		w := math.Min(game.LogicalWidth, sx+p.Width) - x
		ebitenutil.DrawRect(screen, x, game.GroundY-4, w, game.LogicalHeight-game.GroundY+4, color.RGBA{91, 180, 211, 235})
		for i := 0; i < int(w)/44+2; i++ {
			wx := x + float64(i*44)
			ebitenutil.DrawLine(screen, wx, game.GroundY+8, wx+22, game.GroundY+5, color.RGBA{217, 241, 245, 145})
		}
	}
	for _, u := range a.game.Gusts {
		sx := u.X - a.game.Camera
		if sx+u.Width < 0 || sx > game.LogicalWidth {
			continue
		}
		for i := 0; i < 7; i++ {
			y := 120.0 + float64(i)*68 + math.Mod(float64(a.frame+i*23)*0.9, 36)
			dx := 28.0
			dy := 7.0
			if u.Force < 0 {
				dy = -10
			} else {
				dy = 10
			}
			x := sx + 12 + math.Mod(float64(i*53+a.frame), math.Max(28, u.Width-24))
			ebitenutil.DrawLine(screen, x, y, x+dx, y+dy, color.RGBA{240, 249, 247, 118})
		}
	}
}

func (a *app) drawHazardSprites(screen *ebiten.Image) {
	for _, t := range a.game.Thistles {
		sx := t.X - a.game.Camera
		if sx < -60 || sx > game.LogicalWidth+60 {
			continue
		}
		top := game.GroundY - t.Height
		ebitenutil.DrawLine(screen, sx, game.GroundY, sx, top+22, color.RGBA{47, 111, 56, 245})
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Translate(-32, -32)
		op.GeoM.Scale(0.96, 0.96)
		op.GeoM.Translate(sx, top+16)
		screen.DrawImage(a.thistle, op)
	}
	for _, b := range a.game.Bees {
		sx := b.X - a.game.Camera
		if sx < -50 || sx > game.LogicalWidth+50 {
			continue
		}
		img := a.bee[(a.frame/8)%2]
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Translate(-32, -32)
		op.GeoM.Scale(0.84, 0.84)
		op.GeoM.Translate(sx, a.game.BeeY(b))
		screen.DrawImage(img, op)
	}
	for _, p := range a.game.Ponds {
		f := p.Frog
		sx := f.X - a.game.Camera
		if sx < -60 || sx > game.LogicalWidth+60 {
			continue
		}
		frame := 0
		if f.Leaping {
			frame = 1
		}
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Translate(-32, -32)
		op.GeoM.Scale(1.0, 1.0)
		op.GeoM.Translate(sx, f.Y-15)
		screen.DrawImage(a.frog[frame], op)
	}
}

func (a *app) drawFireflies(screen *ebiten.Image) {
	for _, f := range a.game.Fireflies {
		if f.Collected {
			continue
		}
		sx := f.X - a.game.Camera
		if sx < -30 || sx > game.LogicalWidth+30 {
			continue
		}
		pulse := 1.0 + 0.15*math.Sin(float64(a.frame)/8+f.X*0.01)
		vector.DrawFilledCircle(screen, float32(sx), float32(f.Y), float32(13*pulse), color.RGBA{255, 232, 111, 42}, true)
		vector.DrawFilledCircle(screen, float32(sx), float32(f.Y), float32(5*pulse), color.RGBA{255, 240, 124, 230}, true)
	}
}
