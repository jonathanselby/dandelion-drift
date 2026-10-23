package game

import "math"

const (
	worldAhead      = 2000.0
	worldBehind     = 200.0
	initialWorldEnd = 1200.0
)

type chunkKind int

const (
	chunkMeadow chunkKind = iota
	chunkGap
	chunkThistle
	chunkBeeLine
	chunkPond
	chunkUpdraft
	chunkDowndraft
	chunkFireflyArc
)

func (g *Game) seedInitialWorld() {
	g.addWildFlower(760, 125)
	g.genX = initialWorldEnd
	g.lastChunk = chunkMeadow
}

func (g *Game) generateWorldAhead() {
	for g.genX < g.Camera+worldAhead {
		g.generateChunk()
	}
}

func (g *Game) generateChunk() {
	kind := g.chooseChunkKind(g.genX / PxPerMetre)
	start := g.genX
	switch kind {
	case chunkGap:
		g.genX += 400
	case chunkThistle:
		width := 500.0
		g.addWildFlower(start+70, 100+float64(g.rng.Intn(31)))
		g.addWildFlower(start+width-70, 100+float64(g.rng.Intn(31)))
		n := 2 + g.rng.Intn(2)
		for i := 0; i < n; i++ {
			x := start + 235 + float64(i-(n-1)/2)*28
			g.Thistles = append(g.Thistles, Thistle{X: x, Height: 90 + float64(g.rng.Intn(91))})
		}
		g.genX += width
	case chunkBeeLine:
		width := 600.0
		g.addWildFlower(start+70, 100+float64(g.rng.Intn(31)))
		g.addWildFlower(start+width-70, 100+float64(g.rng.Intn(31)))
		n := 2 + g.rng.Intn(3)
		base := []float64{180, 420, 180, 420}
		for i := 0; i < n; i++ {
			g.Bees = append(g.Bees, Bee{X: start + 230 + float64(i)*55, BaseY: base[i], Phase: float64(i) * 1.2})
		}
		g.genX += width
	case chunkPond:
		width := 400.0 + float64(g.rng.Intn(201))
		frogX := start + width/2
		g.Ponds = append(g.Ponds, Pond{X: start, Width: width, Frog: Frog{X: frogX, Y: GroundY - 8, LastLeap: -150}})
		g.genX += width
	case chunkUpdraft:
		width := 400.0
		g.Gusts = append(g.Gusts, Gust{X: start + 145, Width: 110, Force: UpdraftForce})
		for i := 0; i < 5; i++ {
			g.Fireflies = append(g.Fireflies, Firefly{X: start + 200, Y: 500 - float64(i)*70})
		}
		g.genX += width
	case chunkDowndraft:
		width := 300.0
		g.Gusts = append(g.Gusts, Gust{X: start + 145, Width: 60, Force: DowndraftForce})
		g.genX += width
	case chunkFireflyArc:
		width := 500.0
		g.addWildFlower(start+60, 105+float64(g.rng.Intn(25)))
		g.addWildFlower(start+width-60, 105+float64(g.rng.Intn(25)))
		for i := 0; i < 6; i++ {
			t := float64(i) / 5
			y := 390 - 115*math.Sin(math.Pi*t)
			g.Fireflies = append(g.Fireflies, Firefly{X: start + 110 + t*280, Y: y})
		}
		g.genX += width
	default:
		width := 600.0
		g.addWildFlower(start+80, 90+float64(g.rng.Intn(51)))
		g.addWildFlower(start+width-80, 90+float64(g.rng.Intn(51)))
		if g.rng.Intn(3) == 0 {
			g.Fireflies = append(g.Fireflies, Firefly{X: start + 300, Y: 360})
		}
		g.genX += width
		kind = chunkMeadow
	}
	g.lastChunk = kind
}

func (g *Game) chooseChunkKind(metres float64) chunkKind {
	// A flower-free chunk is always followed by a flower-bearing meadow.
	// This keeps the reachability rule structural rather than probabilistic.
	if g.genX-g.lastFlowerX > 100 {
		return chunkMeadow
	}
	// Ponds never follow a gap. Early play strongly favours open chunks;
	// pressure increases toward 750 m and then holds.
	r := g.rng.Intn(100)
	if metres < 250 {
		switch {
		case r < 46:
			return chunkMeadow
		case r < 64:
			return chunkFireflyArc
		case r < 78:
			return chunkGap
		case r < 86:
			return chunkUpdraft
		case r < 92:
			return chunkBeeLine
		default:
			return chunkThistle
		}
	}
	if metres < 750 {
		switch {
		case r < 26:
			return chunkMeadow
		case r < 39:
			return chunkFireflyArc
		case r < 49:
			return chunkGap
		case r < 61:
			return chunkBeeLine
		case r < 73:
			return chunkThistle
		case r < 83:
			return chunkUpdraft
		case r < 91:
			return chunkDowndraft
		default:
			if g.lastChunk == chunkGap {
				return chunkMeadow
			}
			return chunkPond
		}
	}
	switch {
	case r < 18:
		return chunkMeadow
	case r < 28:
		return chunkFireflyArc
	case r < 36:
		return chunkGap
	case r < 51:
		return chunkBeeLine
	case r < 66:
		return chunkThistle
	case r < 76:
		return chunkUpdraft
	case r < 86:
		return chunkDowndraft
	default:
		if g.lastChunk == chunkGap {
			return chunkMeadow
		}
		return chunkPond
	}
}

func (g *Game) addWildFlower(x, height float64) {
	g.Flowers = append(g.Flowers, Flower{X: x, Height: height, Nectar: NectarWild, Bloomed: true})
	if x > g.lastFlowerX {
		g.lastFlowerX = x
	}
}

func (g *Game) pruneWorld() {
	cutoff := g.Camera - worldBehind
	if cutoff <= 0 {
		return
	}
	perched := g.Bunny.Perch
	newPerch := -1
	flowers := g.Flowers[:0]
	for i := range g.Flowers {
		if g.Flowers[i].X < cutoff && i != perched {
			continue
		}
		if i == perched {
			newPerch = len(flowers)
		}
		flowers = append(flowers, g.Flowers[i])
	}
	g.Flowers = flowers
	if perched >= 0 {
		g.Bunny.Perch = newPerch
	}
	bees := g.Bees[:0]
	for _, b := range g.Bees {
		if b.X >= cutoff {
			bees = append(bees, b)
		}
	}
	g.Bees = bees
	thistles := g.Thistles[:0]
	for _, t := range g.Thistles {
		if t.X >= cutoff {
			thistles = append(thistles, t)
		}
	}
	g.Thistles = thistles
	ponds := g.Ponds[:0]
	for _, p := range g.Ponds {
		if p.X+p.Width >= cutoff {
			ponds = append(ponds, p)
		}
	}
	g.Ponds = ponds
	gusts := g.Gusts[:0]
	for _, u := range g.Gusts {
		if u.X+u.Width >= cutoff {
			gusts = append(gusts, u)
		}
	}
	g.Gusts = gusts
	flies := g.Fireflies[:0]
	for _, f := range g.Fireflies {
		if !f.Collected && f.X >= cutoff {
			flies = append(flies, f)
		}
	}
	g.Fireflies = flies
}
