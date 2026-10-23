package game

import "testing"

func TestUATV03StartsAtOneHundredPuffs(t *testing.T) {
	if MaxSeeds != 100 {
		t.Fatalf("MaxSeeds=%d want 100", MaxSeeds)
	}
	g := New(303)
	g.StartFlight()
	if g.Seeds != 100 {
		t.Fatalf("starting seeds=%d want 100", g.Seeds)
	}
}

func TestFireflyIsRechargeSource(t *testing.T) {
	g := isolatedGame()
	g.Seeds = 40
	beforeLift := LiftForSeeds(g.Seeds)
	placeFireflyForNextTick(g)
	g.Update(Input{})
	if g.Seeds != 40+FireflySeeds {
		t.Fatalf("seeds=%d want %d", g.Seeds, 40+FireflySeeds)
	}
	if LiftForSeeds(g.Seeds) <= beforeLift {
		t.Fatalf("firefly did not restore lift: before=%v after=%v", beforeLift, LiftForSeeds(g.Seeds))
	}
}

func TestFireflyRechargeCapsAtOneHundred(t *testing.T) {
	g := isolatedGame()
	g.Seeds = MaxSeeds - 5
	placeFireflyForNextTick(g)
	g.Update(Input{})
	if g.Seeds != MaxSeeds {
		t.Fatalf("seeds=%d want cap %d", g.Seeds, MaxSeeds)
	}
}

func TestFlowerDrainsPuffInsteadOfRecharging(t *testing.T) {
	g := New(303)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + 500, Height: 120, Nectar: 2, Bloomed: true}}
	g.Bunny = Bunny{SX: 500, Y: GroundY - 120 - FlowerRadius - BunnyFeet, State: Perched, Perch: 0}
	g.Seeds = 20
	for i := 0; i < RegrowTicks*3; i++ {
		g.Update(Input{})
	}
	if g.Seeds != 18 || g.Flowers[0].Nectar != 0 {
		t.Fatalf("flower drain seeds=%d capacity=%d", g.Seeds, g.Flowers[0].Nectar)
	}
}

func TestFlowerCanDrainLastPuffAndEndFlight(t *testing.T) {
	g := New(303)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + 500, Height: 120, Nectar: 1, Bloomed: true}}
	g.Bunny = Bunny{SX: 500, Y: GroundY - 120 - FlowerRadius - BunnyFeet, State: Perched, Perch: 0}
	g.Seeds = 1
	for i := 0; i < RegrowTicks; i++ {
		g.Update(Input{})
	}
	if g.Seeds != 0 || g.Bunny.State != Falling {
		t.Fatalf("last-puff flower drain seeds=%d state=%v", g.Seeds, g.Bunny.State)
	}
}

func TestFlightIsLoseableWithoutFireflies(t *testing.T) {
	g := isolatedGame()
	for i := 0; i < MaxSeeds && g.Bunny.State == Flying; i++ {
		g.Bunny.FlapCooldown = 0
		g.Bunny.Y = 300
		g.Bunny.VY = 0
		g.Update(Input{Flap: true})
	}
	if g.Seeds != 0 || g.Bunny.State != Falling {
		t.Fatalf("flight did not become loseable: seeds=%d state=%v", g.Seeds, g.Bunny.State)
	}
}
