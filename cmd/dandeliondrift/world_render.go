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

		petal := color.RGBA{244, 174, 190, 255}
		if f.Planted {
			petal = color.RGBA{248, 194, 151, 255}
		}
		if f.Nectar <= 0 {
			petal = color.RGBA{188, 158, 154, 230}
			headY += 7
		}
		for i := 0; i < 8; i++ {
			ang := 2 * math.Pi * float64(i) / 8
			px := sx + math.Cos(ang)*18
			py := headY + math.Sin(ang)*12
			vector.DrawFilledCircle(screen, float32(px), float32(py), 10, petal, true)
		}
		vector.DrawFilledCircle(screen, float32(sx), float32(headY), 11, color.RGBA{232, 184, 76, 255}, true)
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
