package game

import "testing"

func TestUATV02StartsWithTenTimesOriginalPuffReserve(t *testing.T) {
	if MaxSeeds != 240 {
		t.Fatalf("MaxSeeds=%d want 240", MaxSeeds)
	}
	g := New(202)
	g.StartFlight()
	if g.Seeds != 240 {
		t.Fatalf("starting seeds=%d want 240", g.Seeds)
	}
}

func TestFireflyRestoresPuffsAndLift(t *testing.T) {
	g := isolatedGame()
	g.Seeds = 100
	beforeLift := LiftForSeeds(g.Seeds)
	placeFireflyForNextTick(g)
	g.Update(Input{})
	if g.Seeds != 100+FireflySeeds {
		t.Fatalf("seeds=%d want %d", g.Seeds, 100+FireflySeeds)
	}
	if LiftForSeeds(g.Seeds) <= beforeLift {
		t.Fatalf("firefly did not restore lift: before=%v after=%v", beforeLift, LiftForSeeds(g.Seeds))
	}
	if g.FirefliesCollected != 1 {
		t.Fatalf("collected=%d want 1", g.FirefliesCollected)
	}
}

func TestFireflyRechargeCapsAtFullPuff(t *testing.T) {
	g := isolatedGame()
	g.Seeds = MaxSeeds - 5
	placeFireflyForNextTick(g)
	g.Update(Input{})
	if g.Seeds != MaxSeeds {
		t.Fatalf("seeds=%d want cap %d", g.Seeds, MaxSeeds)
	}
}
