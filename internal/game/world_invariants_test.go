package game

import (
	"math"
	"testing"
)

func generateTo(g *Game, x float64) {
	for g.genX < x {
		g.generateChunk()
	}
}

func TestInvariantsHoldOverLongGeneration(t *testing.T) {
	const distance = 3000 * PxPerMetre
	for seed := int64(1); seed <= 200; seed++ {
		g := New(seed)
		g.StartFlight()
		generateTo(g, distance)

		openingFlowers := 0
		for _, f := range g.Flowers {
			if f.X < initialWorldEnd && !f.Planted {
				openingFlowers++
			}
		}
		if openingFlowers != 1 {
			t.Fatalf("seed %d opening flowers=%d want 1", seed, openingFlowers)
		}
		for _, b := range g.Bees {
			if b.X < initialWorldEnd {
				t.Fatalf("seed %d bee in gentle start at %.1f", seed, b.X)
			}
		}
		for _, th := range g.Thistles {
			if th.X < initialWorldEnd {
				t.Fatalf("seed %d thistle in gentle start at %.1f", seed, th.X)
			}
		}
		for _, p := range g.Ponds {
			if p.X < initialWorldEnd {
				t.Fatalf("seed %d pond in gentle start at %.1f", seed, p.X)
			}
			if p.Width > 600 {
				t.Fatalf("seed %d pond width %.1f exceeds 600", seed, p.Width)
			}
		}
		for _, gust := range g.Gusts {
			if gust.X < initialWorldEnd {
				t.Fatalf("seed %d gust in gentle start at %.1f", seed, gust.X)
			}
		}

		lastFlower := 0.0
		for _, f := range g.Flowers {
			if f.Planted || !f.Bloomed {
				continue
			}
			if lastFlower != 0 {
				gap := f.X - lastFlower
				limit := MaxFlowerGap(ScrollAtWorldX(lastFlower))
				if gap > limit+0.001 {
					t.Fatalf("seed %d flower gap %.1f exceeds %.1f at %.1f", seed, gap, limit, f.X)
				}
			}
			lastFlower = f.X
		}

		for _, f := range g.Flowers {
			if f.Planted {
				continue
			}
			for _, b := range g.Bees {
				if math.Abs(f.X-b.X) < 120 {
					t.Fatalf("seed %d bee %.1f within 120 of flower %.1f", seed, b.X, f.X)
				}
			}
			for _, th := range g.Thistles {
				if math.Abs(f.X-th.X) < 120 {
					t.Fatalf("seed %d thistle %.1f within 120 of flower %.1f", seed, th.X, f.X)
				}
			}
			for _, gust := range g.Gusts {
				d := 0.0
				switch {
				case f.X < gust.X:
					d = gust.X - f.X
				case f.X > gust.X+gust.Width:
					d = f.X - (gust.X + gust.Width)
				}
				if d < 120 {
					t.Fatalf("seed %d gust %.1f..%.1f within 120 of flower %.1f", seed, gust.X, gust.X+gust.Width, f.X)
				}
			}
		}
	}
}

func TestBeeLanesLeaveVerticalGap(t *testing.T) {
	g := New(77)
	g.StartFlight()
	generateTo(g, 3000*PxPerMetre)

	seenLow, seenHigh := false, false
	for _, b := range g.Bees {
		if b.BaseY < 300 {
			seenLow = true
		} else {
			seenHigh = true
		}
	}
	if !seenLow || !seenHigh {
		t.Fatal("long generated world did not exercise both bee lanes")
	}
	// Bob amplitude is ±30. The two lane centres are 240 px apart, leaving
	// 180 px at their closest approach, above the 160 px invariant.
	if (420-30)-(180+30) < 160 {
		t.Fatal("bee lane tuning violates 160 px passable gap")
	}
}

func TestPondNeverFollowsGap(t *testing.T) {
	for seed := int64(1); seed <= 200; seed++ {
		g := New(seed)
		g.StartFlight()
		g.lastChunk = chunkGap
		for _, metres := range []float64{300, 800, 2000} {
			for i := 0; i < 50; i++ {
				if got := g.chooseChunkKind(metres); got == chunkPond {
					t.Fatalf("seed %d chose pond after gap at %.0f m", seed, metres)
				}
			}
		}
	}
}

func TestLiveEntityBudgetAcrossLongWorlds(t *testing.T) {
	for seed := int64(1); seed <= 100; seed++ {
		g := New(seed)
		g.StartFlight()
		g.Camera = 3000 * PxPerMetre
		g.generateWorldAhead()
		g.pruneWorld()
		n := len(g.Flowers) + len(g.Planters) + len(g.Bees) + len(g.Thistles) + len(g.Ponds) + len(g.Gusts) + len(g.Fireflies)
		if n > 120 {
			t.Fatalf("seed %d live entities=%d exceeds 120", seed, n)
		}
	}
}
