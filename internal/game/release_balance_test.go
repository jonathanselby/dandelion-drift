package game

import "testing"

func TestReleaseStartsAtFiftyPuffs(t *testing.T) {
	if MaxSeeds != 50 {
		t.Fatalf("MaxSeeds=%d want 50", MaxSeeds)
	}
	g := New(404)
	g.StartFlight()
	if g.Seeds != 50 {
		t.Fatalf("starting seeds=%d want 50", g.Seeds)
	}
}

func TestFlowerRechargesPuffs(t *testing.T) {
	g := New(404)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + 500, Height: 120, Nectar: 2, Bloomed: true}}
	g.Bunny = Bunny{SX: 500, Y: GroundY - 120 - FlowerRadius - BunnyFeet, State: Perched, Perch: 0}
	g.Seeds = 20
	for i := 0; i < RegrowTicks*3; i++ {
		g.Update(Input{})
	}
	if g.Seeds != 22 || g.Flowers[0].Nectar != 0 {
		t.Fatalf("flower recharge seeds=%d nectar=%d", g.Seeds, g.Flowers[0].Nectar)
	}
}

func TestFlowerRechargeCapsAtFull(t *testing.T) {
	g := New(404)
	g.StartFlight()
	g.Flowers = []Flower{{X: g.Camera + 500, Height: 120, Nectar: 10, Bloomed: true}}
	g.Bunny = Bunny{SX: 500, Y: GroundY - 120 - FlowerRadius - BunnyFeet, State: Perched, Perch: 0}
	g.Seeds = MaxSeeds - 1
	for i := 0; i < RegrowTicks*3; i++ {
		g.Update(Input{})
	}
	if g.Seeds != MaxSeeds || g.Flowers[0].Nectar != 9 {
		t.Fatalf("full recharge seeds=%d nectar=%d", g.Seeds, g.Flowers[0].Nectar)
	}
}

func TestFireflyDrainsPuffs(t *testing.T) {
	g := isolatedGame()
	g.Seeds = 40
	placeFireflyForNextTick(g)
	g.Update(Input{})
	want := 40 - FireflySeeds
	if want < 0 {
		want = 0
	}
	if g.Seeds != want {
		t.Fatalf("seeds=%d want %d", g.Seeds, want)
	}
	if g.FirefliesCollected != 1 {
		t.Fatalf("collected=%d want 1", g.FirefliesCollected)
	}
}

func TestFireflyDrainCannotGoBelowZero(t *testing.T) {
	g := isolatedGame()
	g.Seeds = 5
	placeFireflyForNextTick(g)
	g.Update(Input{})
	if g.Seeds != 0 || g.Bunny.State != Falling {
		t.Fatalf("firefly drain seeds=%d state=%v", g.Seeds, g.Bunny.State)
	}
}
