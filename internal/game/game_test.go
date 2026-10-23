package game

import "testing"

func TestStartScreenSpaceStartsFlightWithoutSpendingSeed(t *testing.T) {
	g := New(1)
	g.Update(Input{Flap: true})
	if g.Screen != Play {
		t.Fatalf("screen = %v, want Play", g.Screen)
	}
	if g.Seeds != MaxSeeds {
		t.Fatalf("seeds = %d, want %d", g.Seeds, MaxSeeds)
	}
}

func TestFlapRisesAndSpendsOneSeed(t *testing.T) {
	g := New(1)
	g.StartFlight()
	beforeY := g.Bunny.Y
	g.Update(Input{Flap: true})
	if g.Seeds != MaxSeeds-1 {
		t.Fatalf("seeds = %d, want %d", g.Seeds, MaxSeeds-1)
	}
	if g.Bunny.VY >= 0 || g.Bunny.Y >= beforeY {
		t.Fatalf("flap did not raise bunny: vy=%v y=%v before=%v", g.Bunny.VY, g.Bunny.Y, beforeY)
	}
}

func TestLiftAndGlideMonotonicInSeeds(t *testing.T) {
	if LiftForSeeds(MaxSeeds) <= LiftForSeeds(1) {
		t.Fatal("full puff should provide stronger lift")
	}
	if MaxFallForSeeds(MaxSeeds) >= MaxFallForSeeds(1) {
		t.Fatal("full puff should fall more slowly")
	}
}

func TestLastFlapStillLifts(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Seeds = 1
	g.Bunny.VY = 0
	g.Update(Input{Flap: true})
	if g.Seeds != 0 {
		t.Fatalf("seeds = %d, want 0", g.Seeds)
	}
	if g.Bunny.VY >= 0 {
		t.Fatalf("last flap did not lift: vy=%v", g.Bunny.VY)
	}
	if g.Bunny.State != Falling {
		t.Fatalf("state = %v, want Falling", g.Bunny.State)
	}
}

func TestCeilingClamp(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Bunny.Y = Ceiling + 0.1
	g.Bunny.VY = -4
	g.Update(Input{})
	if g.Bunny.Y != Ceiling || g.Bunny.VY != 0 {
		t.Fatalf("ceiling clamp failed: y=%v vy=%v", g.Bunny.Y, g.Bunny.VY)
	}
}

func TestDriftBounds(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Bunny.SX = MinSX
	g.Update(Input{Left: true})
	if g.Bunny.SX != MinSX {
		t.Fatalf("left bound = %v", g.Bunny.SX)
	}
	g.Bunny.SX = MaxSX
	g.Update(Input{Right: true})
	if g.Bunny.SX != MaxSX {
		t.Fatalf("right bound = %v", g.Bunny.SX)
	}
}

func TestGroundKnockUsesGuard(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Bunny.Y = GroundY - 21
	g.Bunny.VY = 1
	g.Update(Input{})
	if g.Seeds != MaxSeeds-1 {
		t.Fatalf("first ground touch seeds=%d", g.Seeds)
	}
	g.Bunny.Y = GroundY - 21
	g.Bunny.VY = 1
	g.Update(Input{})
	if g.Seeds != MaxSeeds-1 {
		t.Fatalf("guard did not prevent second loss: seeds=%d", g.Seeds)
	}
}

func TestEmptyPuffFallsSoftlyAndEnds(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Seeds = 0
	g.Bunny.State = Falling
	g.Bunny.Y = GroundY - 23
	g.Bunny.VY = 1
	g.Update(Input{})
	if g.Bunny.State != Landed {
		t.Fatalf("state=%v want Landed", g.Bunny.State)
	}
	if g.EndTicks != EndTicks {
		t.Fatalf("end ticks=%d want %d", g.EndTicks, EndTicks)
	}
}

func TestEndScreenReturnsToAttract(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Bunny.State = Landed
	g.EndTicks = 2
	g.Update(Input{})
	if g.Screen != Play {
		t.Fatal("returned too early")
	}
	g.Update(Input{})
	if g.Screen != Attract {
		t.Fatalf("screen=%v want Attract", g.Screen)
	}
}

func TestAttractCanStartAgainAfterCompletedFlight(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Bunny.State = Landed
	g.EndTicks = 1
	g.Update(Input{})
	g.Update(Input{Flap: true})
	if g.Screen != Play || g.Bunny.State != Flying || g.Seeds != MaxSeeds {
		t.Fatalf("restart failed: screen=%v state=%v seeds=%d", g.Screen, g.Bunny.State, g.Seeds)
	}
}

func TestScoreIsMetres(t *testing.T) {
	g := New(1)
	g.StartFlight()
	for i := 0; i < 100; i++ {
		g.Update(Input{Flap: i%10 == 0})
	}
	if g.Score != g.Metres {
		t.Fatalf("score=%d metres=%d", g.Score, g.Metres)
	}
}

func TestIdleFlightEventuallyEnds(t *testing.T) {
	g := New(1)
	g.StartFlight()
	for i := 0; i < 20000 && g.Bunny.State != Landed; i++ {
		g.Update(Input{})
	}
	if g.Bunny.State != Landed {
		t.Fatal("idle flight did not eventually reach a gentle landing")
	}
}

func TestCompletedFlightReturnsAndStartsFresh(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Seeds = 0
	g.Bunny.State = Falling
	g.Bunny.Y = GroundY - 23
	g.Bunny.VY = 1
	g.Update(Input{})
	for g.Screen != Attract {
		g.Update(Input{})
	}
	g.Update(Input{Flap: true})
	if g.Screen != Play || g.Seeds != MaxSeeds || g.Metres != 0 || g.Released != 0 {
		t.Fatalf("fresh start failed: screen=%v seeds=%d metres=%d released=%d", g.Screen, g.Seeds, g.Metres, g.Released)
	}
}

func TestPerchOnlyOnGentleDescent(t *testing.T) {
	newGame := func() *Game {
		g := New(1)
		g.StartFlight()
		g.Flowers = []Flower{{X: g.Camera + g.Bunny.SX, Height: 120, Nectar: NectarWild, Bloomed: true}}
		g.Bunny.Y = GroundY - 120 - BunnyFeet
		return g
	}

	g := newGame()
	g.Bunny.VY = -1
	g.Update(Input{})
	if g.Bunny.State == Perched {
		t.Fatal("rising bunny perched")
	}

	g = newGame()
	g.Seeds = 1
	g.Bunny.VY = 2.4
	g.Update(Input{})
	if g.Bunny.State == Perched {
		t.Fatal("fast-falling bunny perched")
	}

	g = newGame()
	g.Bunny.VY = 0.5
	g.Update(Input{})
	if g.Bunny.State != Perched {
		t.Fatalf("gentle descent state=%v want Perched", g.Bunny.State)
	}
}

func TestRegrowUntilFullOrNectarSpent(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + 500, Height: 120, Nectar: 2, Bloomed: true}}
	g.Bunny = Bunny{SX: 500, Y: GroundY - 120 - FlowerRadius - BunnyFeet, State: Perched, Perch: 0}
	g.Seeds = 20
	for i := 0; i < RegrowTicks*3; i++ {
		g.Update(Input{})
	}
	if g.Seeds != 22 || g.Flowers[0].Nectar != 0 {
		t.Fatalf("nectar-limited regrow seeds=%d nectar=%d", g.Seeds, g.Flowers[0].Nectar)
	}

	g = New(1)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + 500, Height: 120, Nectar: 10, Bloomed: true}}
	g.Bunny = Bunny{SX: 500, Y: GroundY - 120 - FlowerRadius - BunnyFeet, State: Perched, Perch: 0}
	g.Seeds = MaxSeeds - 1
	for i := 0; i < RegrowTicks*3; i++ {
		g.Update(Input{})
	}
	if g.Seeds != MaxSeeds || g.Flowers[0].Nectar != 9 {
		t.Fatalf("full-puff regrow seeds=%d nectar=%d", g.Seeds, g.Flowers[0].Nectar)
	}
}

func TestPerchCarriesLeftAndAutoHops(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + MinSX + 1, Height: 120, Nectar: NectarWild, Bloomed: true}}
	g.Bunny = Bunny{SX: MinSX + 1, Y: GroundY - 120 - FlowerRadius - BunnyFeet, State: Perched, Perch: 0}
	g.Update(Input{})
	if g.Bunny.State != Flying || g.Bunny.VY >= 0 || g.Bunny.SX != MinSX {
		t.Fatalf("auto hop failed: state=%v vy=%v sx=%v", g.Bunny.State, g.Bunny.VY, g.Bunny.SX)
	}
}

func TestFlapTakesOffFromPerch(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + 400, Height: 120, Nectar: NectarWild, Bloomed: true}}
	g.Bunny = Bunny{SX: 400, Y: GroundY - 120 - FlowerRadius - BunnyFeet, State: Perched, Perch: 0}
	before := g.Seeds
	g.Update(Input{Flap: true})
	if g.Bunny.State != Flying || g.Bunny.Perch != -1 {
		t.Fatalf("state=%v perch=%d", g.Bunny.State, g.Bunny.Perch)
	}
	if g.Seeds != before-1 {
		t.Fatalf("seeds=%d want %d", g.Seeds, before-1)
	}
}

func TestSpentFlowerCannotBePerched(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + g.Bunny.SX, Height: 120, Nectar: 0, Bloomed: true}}
	g.Bunny.Y = GroundY - 120 - BunnyFeet
	g.Bunny.VY = 0.5
	g.Update(Input{})
	if g.Bunny.State == Perched {
		t.Fatal("spent flower accepted a perch")
	}
}

func TestOneInFourSeedsPlants(t *testing.T) {
	g := New(1)
	g.StartFlight()
	for i := 0; i < PlantEvery; i++ {
		g.Bunny.FlapCooldown = 0
		g.Bunny.Y = 250
		g.Bunny.VY = 0
		g.Update(Input{Flap: true})
	}
	if g.Released != PlantEvery {
		t.Fatalf("released=%d want %d", g.Released, PlantEvery)
	}
	if len(g.Planters) != 1 {
		t.Fatalf("planters=%d want 1", len(g.Planters))
	}
}

func TestPlantingOnGrassOnly(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Ponds = []Pond{{X: 890, Width: 80}}
	g.Planters = []Planter{{X: 900, Y: GroundY, VX: PlanterVX, VY: PlanterVY}}
	before := g.plantedFlowerCount()
	g.Update(Input{})
	if g.plantedFlowerCount() != before {
		t.Fatal("planter created a flower in water")
	}

	g.Planters = []Planter{{X: 1000, Y: GroundY, VX: PlanterVX, VY: PlanterVY}}
	g.Update(Input{})
	if g.plantedFlowerCount() != before+1 {
		t.Fatalf("grass planter did not create sprout: before=%d after=%d", before, g.plantedFlowerCount())
	}
}

func TestSproutBloomsAfterDelay(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Flowers = []Flower{{X: 900, Height: 100, Nectar: NectarPlanted, Planted: true, Bloomed: false, GrowTicks: 2}}
	g.Update(Input{})
	if g.Flowers[0].Bloomed {
		t.Fatal("sprout bloomed early")
	}
	g.Update(Input{})
	if !g.Flowers[0].Bloomed || g.Flowers[0].GrowTicks != 0 {
		t.Fatalf("sprout did not bloom: %+v", g.Flowers[0])
	}
}

func TestPlantedFlowerCap(t *testing.T) {
	g := New(1)
	g.StartFlight()
	g.Flowers = g.Flowers[:0]
	for i := 0; i < MaxPlanted; i++ {
		g.Flowers = append(g.Flowers, Flower{X: 2000 + float64(i)*100, Height: 100, Nectar: NectarPlanted, Planted: true, Bloomed: true})
	}
	g.Planters = []Planter{{X: 900, Y: GroundY, VX: PlanterVX, VY: PlanterVY}}
	g.Update(Input{})
	if got := g.plantedFlowerCount(); got != MaxPlanted {
		t.Fatalf("planted flowers=%d want cap %d", got, MaxPlanted)
	}
}

func TestGenerationDeterministicFromSeed(t *testing.T) {
	g1 := New(77)
	g1.StartFlight()
	g2 := New(77)
	g2.StartFlight()

	if len(g1.Flowers) != len(g2.Flowers) {
		t.Fatalf("flower counts differ: %d vs %d", len(g1.Flowers), len(g2.Flowers))
	}
	for i := range g1.Flowers {
		if g1.Flowers[i] != g2.Flowers[i] {
			t.Fatalf("flower %d differs: %+v vs %+v", i, g1.Flowers[i], g2.Flowers[i])
		}
	}
}

func TestGentleStartHasOneFlower(t *testing.T) {
	g := New(1)
	g.StartFlight()
	count := 0
	for _, f := range g.Flowers {
		if f.X < initialWorldEnd {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("opening flowers=%d want 1", count)
	}
}

func TestGeneratedFlowersStayWithinReach(t *testing.T) {
	g := New(99)
	g.StartFlight()
	g.Camera = 120000
	g.generateWorldAhead()
	maxGap := MaxFlowerGap(ScrollStart)
	var last float64
	for i, f := range g.Flowers {
		if !f.Bloomed || f.Planted {
			continue
		}
		if i > 0 && last != 0 && f.X-last > maxGap+0.001 {
			t.Fatalf("flower gap %.1f exceeds conservative max %.1f at x=%.1f", f.X-last, maxGap, f.X)
		}
		last = f.X
	}
}

func TestPlantersAreBounded(t *testing.T) {
	g := New(1)
	g.StartFlight()
	for i := 0; i < 40; i++ {
		g.Bunny.FlapCooldown = 0
		g.Bunny.Y = 250
		g.Bunny.VY = 0
		g.Update(Input{Flap: true})
		if g.Seeds == 0 {
			g.Seeds = MaxSeeds
			g.Bunny.State = Flying
		}
	}
	if len(g.Planters) > MaxPlanted {
		t.Fatalf("planters=%d cap=%d", len(g.Planters), MaxPlanted)
	}
}
