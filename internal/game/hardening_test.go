package game

import (
	"math"
	"testing"
)

func TestHardeningRepresentativeRunsKeepStateSane(t *testing.T) {
	for seed := int64(1); seed <= 32; seed++ {
		g := New(seed)
		g.StartFlight()

		for tick := 0; tick < 12000; tick++ {
			in := Input{}
			if g.Screen == Play {
				switch g.Bunny.State {
				case Flying:
					if g.Seeds > 0 && g.Bunny.FlapCooldown == 0 && (g.Bunny.Y > 360 || tick%53 == 0) {
						in.Flap = true
					}
					phase := tick % 900
					if phase < 120 {
						in.Left = true
					} else if phase >= 450 && phase < 570 {
						in.Right = true
					}
				case Perched:
					if g.Seeds >= MaxSeeds-3 || tick%90 == 0 {
						in.Flap = true
					}
				}
			}

			g.Update(in)
			assertHardenedState(t, seed, tick, g)

			if g.Screen == Attract {
				g.Update(Input{Flap: true})
				if g.Screen != Play || g.Seeds != MaxSeeds || g.Metres != 0 || g.Released != 0 {
					t.Fatalf("seed %d tick %d dirty restart: screen=%v seeds=%d metres=%d released=%d", seed, tick, g.Screen, g.Seeds, g.Metres, g.Released)
				}
			}
		}
	}
}

func assertHardenedState(t *testing.T, seed int64, tick int, g *Game) {
	t.Helper()
	if g.Seeds < 0 || g.Seeds > MaxSeeds {
		t.Fatalf("seed %d tick %d seeds=%d", seed, tick, g.Seeds)
	}
	if g.Chain < 0 || g.Chain > ChainMax {
		t.Fatalf("seed %d tick %d chain=%d", seed, tick, g.Chain)
	}
	if g.Scroll < ScrollStart-0.0001 || g.Scroll > ScrollMax+0.0001 {
		t.Fatalf("seed %d tick %d scroll=%v", seed, tick, g.Scroll)
	}
	if g.Screen == Play && (g.Bunny.SX < MinSX-0.0001 || g.Bunny.SX > MaxSX+0.0001) {
		t.Fatalf("seed %d tick %d sx=%v", seed, tick, g.Bunny.SX)
	}
	for name, v := range map[string]float64{
		"camera": g.Camera,
		"scroll": g.Scroll,
		"sx":     g.Bunny.SX,
		"y":      g.Bunny.Y,
		"vx":     g.Bunny.VX,
		"vy":     g.Bunny.VY,
		"genX":   g.genX,
	} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			t.Fatalf("seed %d tick %d %s=%v", seed, tick, name, v)
		}
	}
	if len(g.Planters) > MaxPlanted {
		t.Fatalf("seed %d tick %d planters=%d", seed, tick, len(g.Planters))
	}
	if g.plantedFlowerCount() > MaxPlanted {
		t.Fatalf("seed %d tick %d planted flowers=%d", seed, tick, g.plantedFlowerCount())
	}
	if len(g.Events) > 8 {
		t.Fatalf("seed %d tick %d events=%d exceeds budget", seed, tick, len(g.Events))
	}
	live := len(g.Flowers) + len(g.Planters) + len(g.Bees) + len(g.Thistles) + len(g.Ponds) + len(g.Gusts) + len(g.Fireflies)
	if live > 120 {
		t.Fatalf("seed %d tick %d live entities=%d", seed, tick, live)
	}
	if g.Score != g.Metres+g.FireflyScore {
		t.Fatalf("seed %d tick %d score=%d metres=%d fireflies=%d", seed, tick, g.Score, g.Metres, g.FireflyScore)
	}
}

func TestSpaceOnlyRunAdvancesWithoutHorizontalInput(t *testing.T) {
	g := New(991)
	g.StartFlight()
	g.Flowers = nil
	g.Planters = nil
	g.Bees = nil
	g.Thistles = nil
	g.Ponds = nil
	g.Gusts = nil
	g.Fireflies = nil
	g.genX = 1e12
	startX := g.Bunny.SX

	for tick := 0; tick < 600 && g.Screen == Play; tick++ {
		flap := g.Bunny.State == Flying && g.Seeds > 0 && g.Bunny.FlapCooldown == 0 && g.Bunny.Y > 350
		g.Update(Input{Flap: flap})
	}
	if g.Camera <= 0 || g.Metres <= 0 {
		t.Fatalf("space-only run did not advance: camera=%v metres=%d", g.Camera, g.Metres)
	}
	if g.Bunny.SX != startX {
		t.Fatalf("space-only run drifted horizontally: got %v want %v", g.Bunny.SX, startX)
	}
}

func TestFallingStateIgnoresHazardDamage(t *testing.T) {
	g := New(7)
	g.StartFlight()
	g.Seeds = 0
	g.Bunny.State = Falling
	g.Bunny.Y = 300
	g.Bunny.VY = 0
	wx := g.Camera + g.Bunny.SX
	g.Bees = []Bee{{X: wx + 1, BaseY: g.Bunny.Y}}
	g.Thistles = []Thistle{{X: wx, Height: 400}}
	g.Ponds = []Pond{{X: wx - 100, Width: 200, Frog: Frog{X: wx, Y: g.Bunny.Y, Leaping: true}}}

	g.Update(Input{})
	if g.Seeds != 0 {
		t.Fatalf("falling state regained/lost seeds: %d", g.Seeds)
	}
	for _, ev := range g.Events {
		switch ev.Type {
		case EvBee, EvThistle, EvFrogCatch, EvSeedsLost:
			t.Fatalf("falling state emitted hazard event %v", ev.Type)
		}
	}
}

func TestRepeatedCompletedFlightsStartClean(t *testing.T) {
	g := New(123)
	for flight := 0; flight < 40; flight++ {
		g.StartFlight()
		for i := 0; i < 40; i++ {
			g.Update(Input{Flap: i%12 == 0})
		}
		g.Seeds = 0
		g.Bunny.State = Falling
		g.Bunny.Y = GroundY - BunnyFeet - 1
		g.Bunny.VY = 1
		g.Update(Input{})
		for g.Screen != Attract {
			g.Update(Input{})
		}
		g.Update(Input{Flap: true})
		if g.Screen != Play || g.Seeds != MaxSeeds || g.Score != 0 || g.Metres != 0 || g.Released != 0 || g.Chain != 0 || g.FireflyScore != 0 || g.FirefliesCollected != 0 {
			t.Fatalf("flight %d dirty restart: screen=%v seeds=%d score=%d metres=%d released=%d chain=%d flyScore=%d collected=%d", flight, g.Screen, g.Seeds, g.Score, g.Metres, g.Released, g.Chain, g.FireflyScore, g.FirefliesCollected)
		}
	}
}
