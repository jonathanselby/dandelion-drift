package game

import "testing"

func TestStartScreenSpaceStartsFlightWithoutSpendingSeed(t *testing.T) {
	g := New()
	g.Update(Input{Flap: true})
	if g.Screen != Play {
		t.Fatalf("screen = %v, want Play", g.Screen)
	}
	if g.Seeds != MaxSeeds {
		t.Fatalf("seeds = %d, want %d", g.Seeds, MaxSeeds)
	}
}

func TestFlapRisesAndSpendsOneSeed(t *testing.T) {
	g := New()
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
	g := New()
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
	g := New()
	g.StartFlight()
	g.Bunny.Y = Ceiling + 0.1
	g.Bunny.VY = -4
	g.Update(Input{})
	if g.Bunny.Y != Ceiling || g.Bunny.VY != 0 {
		t.Fatalf("ceiling clamp failed: y=%v vy=%v", g.Bunny.Y, g.Bunny.VY)
	}
}

func TestDriftBounds(t *testing.T) {
	g := New()
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
	g := New()
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
	g := New()
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
	g := New()
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
	g := New()
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
	g := New()
	g.StartFlight()
	for i := 0; i < 100; i++ {
		g.Update(Input{Flap: i%10 == 0})
	}
	if g.Score != g.Metres {
		t.Fatalf("score=%d metres=%d", g.Score, g.Metres)
	}
}

func TestIdleFlightEventuallyEnds(t *testing.T) {
	g := New()
	g.StartFlight()
	for i := 0; i < 20000 && g.Bunny.State != Landed; i++ {
		g.Update(Input{})
	}
	if g.Bunny.State != Landed {
		t.Fatal("idle flight did not eventually reach a gentle landing")
	}
}

func TestCompletedFlightReturnsAndStartsFresh(t *testing.T) {
	g := New()
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
