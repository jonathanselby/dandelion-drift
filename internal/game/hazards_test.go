package game

import (
	"reflect"
	"testing"
)

func isolatedGame() *Game {
	g := New(1)
	g.StartFlight()
	g.Flowers = nil
	g.Planters = nil
	g.Bees = nil
	g.Thistles = nil
	g.Ponds = nil
	g.Gusts = nil
	g.Fireflies = nil
	g.genX = 1e9
	return g
}

func TestEachHazardRemovesItsSeeds(t *testing.T) {
	t.Run("bee", func(t *testing.T) {
		g := isolatedGame()
		worldX := g.Camera + g.Bunny.SX
		g.Bees = []Bee{{X: worldX + 1, BaseY: g.Bunny.Y, Phase: -1.0 / 30}}
		g.Update(Input{})
		if g.Seeds != MaxSeeds-BeeSeeds {
			t.Fatalf("bee seeds=%d", g.Seeds)
		}
		if g.Bunny.PushTicks != BeePushTicks {
			t.Fatalf("bee push=%d", g.Bunny.PushTicks)
		}
	})
	t.Run("thistle", func(t *testing.T) {
		g := isolatedGame()
		worldX := g.Camera + g.Bunny.SX
		g.Bunny.Y = GroundY - 100
		g.Thistles = []Thistle{{X: worldX + g.Scroll, Height: 180}}
		g.Update(Input{})
		if g.Seeds != MaxSeeds-ThistleSeeds {
			t.Fatalf("thistle seeds=%d", g.Seeds)
		}
		if g.Bunny.VY != BounceVY {
			t.Fatalf("thistle vy=%v", g.Bunny.VY)
		}
	})
	t.Run("water", func(t *testing.T) {
		g := isolatedGame()
		worldX := g.Camera + g.Bunny.SX + g.Scroll
		g.Ponds = []Pond{{X: worldX - 50, Width: 100, Frog: Frog{Y: GroundY - 18, LastLeap: 10000}}}
		g.Bunny.Y = GroundY - BunnyFeet
		g.Bunny.VY = 1
		g.Update(Input{})
		if g.Seeds != MaxSeeds-WaterSeeds {
			t.Fatalf("water seeds=%d", g.Seeds)
		}
		if g.Bunny.VY != SplashVY {
			t.Fatalf("water vy=%v", g.Bunny.VY)
		}
	})
	t.Run("frog", func(t *testing.T) {
		g := isolatedGame()
		worldX := g.Camera + g.Bunny.SX + g.Scroll
		g.Ponds = []Pond{{X: worldX - 100, Width: 200, Frog: Frog{X: worldX, Y: g.Bunny.Y, Leaping: true, LastLeap: g.Ticks}}}
		g.Update(Input{})
		if g.Seeds != MaxSeeds-FrogSeeds {
			t.Fatalf("frog seeds=%d", g.Seeds)
		}
		if g.Bunny.VY != FlingVY {
			t.Fatalf("frog vy=%v", g.Bunny.VY)
		}
		if g.Bunny.Guard != FrogGuard {
			t.Fatalf("frog guard=%d", g.Bunny.Guard)
		}
	})
}

func TestGuardPreventsDoubleKnock(t *testing.T) {
	g := isolatedGame()
	worldX := g.Camera + g.Bunny.SX
	g.Bees = []Bee{{X: worldX + 1, BaseY: g.Bunny.Y, Phase: -1.0 / 30}}
	g.Update(Input{})
	first := g.Seeds
	g.Update(Input{})
	if g.Seeds != first {
		t.Fatalf("guard failed: %d -> %d", first, g.Seeds)
	}
}

func TestGroundAlwaysBouncesEvenWhenGuarded(t *testing.T) {
	g := isolatedGame()
	g.Bunny.Guard = 20
	g.Bunny.Y = GroundY - BunnyFeet
	g.Bunny.VY = 1
	before := g.Seeds
	g.Update(Input{})
	if g.Bunny.VY != BounceVY {
		t.Fatalf("vy=%v", g.Bunny.VY)
	}
	if g.Seeds != before {
		t.Fatalf("guarded ground lost seeds")
	}
}

func TestFrogLeapsOnlyWhenBunnyLowOverhead(t *testing.T) {
	g := isolatedGame()
	worldX := g.Camera + g.Bunny.SX + g.Scroll
	g.Ponds = []Pond{{X: worldX - 100, Width: 200, Frog: Frog{Y: GroundY - 18, LastLeap: -FrogLeapWait}}}
	g.Bunny.Y = GroundY - 240
	g.Update(Input{})
	if g.Ponds[0].Frog.Leaping {
		t.Fatal("frog leapt while bunny high")
	}
	g.Bunny.Y = GroundY - 100
	g.Ponds[0].Frog.LastLeap = g.Ticks - FrogLeapWait
	g.Update(Input{})
	if !g.Ponds[0].Frog.Leaping {
		t.Fatal("frog did not leap")
	}
}

func TestGustsChangeFall(t *testing.T) {
	plain := isolatedGame()
	plain.Bunny.VY = 0
	plain.Update(Input{})
	up := isolatedGame()
	up.Gusts = []Gust{{X: up.Camera, Width: 1000, Force: UpdraftForce}}
	up.Bunny.VY = 0
	up.Update(Input{})
	down := isolatedGame()
	down.Gusts = []Gust{{X: down.Camera, Width: 1000, Force: DowndraftForce}}
	down.Bunny.VY = 0
	down.Update(Input{})
	if !(up.Bunny.VY < plain.Bunny.VY && down.Bunny.VY > plain.Bunny.VY) {
		t.Fatalf("up=%v plain=%v down=%v", up.Bunny.VY, plain.Bunny.VY, down.Bunny.VY)
	}
}

func placeFireflyForNextTick(g *Game) {
	g.Fireflies = append(g.Fireflies, Firefly{X: g.Camera + g.Scroll + g.Bunny.SX, Y: g.Bunny.Y})
}

func TestChainRisesWithinWindowAndResets(t *testing.T) {
	g := isolatedGame()
	placeFireflyForNextTick(g)
	g.Update(Input{})
	placeFireflyForNextTick(g)
	g.Update(Input{})
	if g.Chain != 2 {
		t.Fatalf("chain=%d", g.Chain)
	}
	for i := 0; i < ChainWindow; i++ {
		g.Update(Input{})
	}
	if g.Chain != 0 || g.ChainTicks != 0 {
		t.Fatalf("expired chain=%d ticks=%d", g.Chain, g.ChainTicks)
	}
}

func TestChainCapsAtFive(t *testing.T) {
	g := isolatedGame()
	g.Chain = ChainMax
	g.ChainTicks = 10
	placeFireflyForNextTick(g)
	g.Update(Input{})
	if g.Chain != ChainMax {
		t.Fatalf("chain=%d", g.Chain)
	}
}

func TestScoreIsMetresPlusFireflies(t *testing.T) {
	g := isolatedGame()
	placeFireflyForNextTick(g)
	g.Update(Input{})
	if g.Score != g.Metres+g.FireflyScore {
		t.Fatalf("score=%d metres=%d fly=%d", g.Score, g.Metres, g.FireflyScore)
	}
	if g.FirefliesCollected != 1 {
		t.Fatalf("collected=%d", g.FirefliesCollected)
	}
}

func TestDeterministicGameplayWithHazards(t *testing.T) {
	g1 := New(55)
	g1.StartFlight()
	g2 := New(55)
	g2.StartFlight()
	for i := 0; i < 5000; i++ {
		in := Input{Flap: i%37 == 0, Left: i%211 < 15, Right: i%307 < 15}
		g1.Update(in)
		g2.Update(in)
	}
	if g1.Screen != g2.Screen || g1.Bunny != g2.Bunny || g1.Seeds != g2.Seeds || g1.Score != g2.Score || g1.Chain != g2.Chain {
		t.Fatal("core state diverged")
	}
	if !reflect.DeepEqual(g1.Flowers, g2.Flowers) || !reflect.DeepEqual(g1.Bees, g2.Bees) || !reflect.DeepEqual(g1.Thistles, g2.Thistles) || !reflect.DeepEqual(g1.Ponds, g2.Ponds) || !reflect.DeepEqual(g1.Gusts, g2.Gusts) || !reflect.DeepEqual(g1.Fireflies, g2.Fireflies) {
		t.Fatal("world state diverged")
	}
}
