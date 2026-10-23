package game

const (
	worldAhead      = 2000.0
	worldBehind     = 200.0
	initialWorldEnd = 1200.0
	meadowWidth     = 600.0
	gapWidth        = 400.0
)

func (g *Game) seedInitialWorld() {
	// The gentle opening contains exactly one wild flower and no hazards.
	g.Flowers = append(g.Flowers, Flower{
		X:       760,
		Height:  125,
		Nectar:  NectarWild,
		Bloomed: true,
	})
	g.genX = initialWorldEnd
	// Force the first generated chunk to be a meadow, keeping the opening
	// flower comfortably inside the conservative full-puff reach limit.
	g.lastChunkWasGap = true
}

func (g *Game) generateWorldAhead() {
	for g.genX < g.Camera+worldAhead {
		g.generateChunk()
	}
}

func (g *Game) generateChunk() {
	makeGap := !g.lastChunkWasGap && g.rng.Intn(4) == 0
	if makeGap {
		g.genX += gapWidth
		g.lastChunkWasGap = true
		return
	}

	start := g.genX
	g.addWildFlower(start+80, 90+float64(g.rng.Intn(51)))
	g.addWildFlower(start+meadowWidth-80, 90+float64(g.rng.Intn(51)))
	g.genX += meadowWidth
	g.lastChunkWasGap = false
}

func (g *Game) addWildFlower(x, height float64) {
	g.Flowers = append(g.Flowers, Flower{
		X:       x,
		Height:  height,
		Nectar:  NectarWild,
		Bloomed: true,
	})
}

func (g *Game) pruneWorld() {
	cutoff := g.Camera - worldBehind
	if cutoff <= 0 || len(g.Flowers) == 0 {
		return
	}

	perched := g.Bunny.Perch
	newPerch := -1
	out := g.Flowers[:0]
	for i := range g.Flowers {
		keep := g.Flowers[i].X >= cutoff || i == perched
		if !keep {
			continue
		}
		if i == perched {
			newPerch = len(out)
		}
		out = append(out, g.Flowers[i])
	}
	g.Flowers = out
	if perched >= 0 {
		g.Bunny.Perch = newPerch
	}

	ponds := g.Ponds[:0]
	for _, p := range g.Ponds {
		if p.X+p.Width >= cutoff {
			ponds = append(ponds, p)
		}
	}
	g.Ponds = ponds
}
