package game

import (
	"math"
	"math/rand"
)

const (
	LogicalWidth  = 960
	LogicalHeight = 720
	TPS           = 60

	PxPerMetre = 40.0
	Ceiling    = 60.0
	GroundY    = 640.0
	MinSX      = 120.0
	MaxSX      = 600.0

	MaxSeeds = 24

	ScrollStart = 1.5
	ScrollMax   = 3.0
	ScrollRamp  = 30000.0

	Gravity      = 0.10
	FlapImpulse  = 3.0
	FlapCooldown = 8
	MaxRise      = -4.0
	DriftSpeed   = 1.6

	MaxFallFull  = 1.0
	MaxFallEmpty = 2.4
	MaxFallStem  = 2.0

	BunnyFeet     = 22.0
	FlowerRadius  = 28.0
	PerchMaxVY    = 2.0
	RegrowTicks   = 15
	NectarWild    = 10
	NectarPlanted = 6

	PlantEvery  = 4
	PlanterVX   = 0.4
	PlanterVY   = 0.6
	SproutTicks = 240
	MaxPlanted  = 6

	GroundSeeds = 1
	GuardTicks  = 60
	BounceVY    = -2.5

	EndTicks = 180
)

type Screen int

const (
	Attract Screen = iota
	Play
)

type BunnyState int

const (
	Flying BunnyState = iota
	Perched
	Falling
	Landed
)

type EventType int

const (
	EvFlap EventType = iota
	EvBounce
	EvSeedsLost
	EvPuffGone
	EvFlightOver
	EvReturnToAttract
	EvPerch
	EvRegrow
	EvTakeOff
	EvSprouted
)

type Event struct {
	Type  EventType
	Value int
}

type Input struct {
	Flap        bool
	Left, Right bool
}

type Bunny struct {
	SX, Y  float64
	VX, VY float64
	State  BunnyState
	Guard  int

	Perch        int
	RegrowTimer  int
	FlapCooldown int
}

type Flower struct {
	X, Height float64
	Nectar    int
	Planted   bool
	Bloomed   bool
	GrowTicks int
}

type Planter struct {
	X, Y   float64
	VX, VY float64
}

type Pond struct {
	X, Width float64
}

type Game struct {
	Screen Screen

	Camera float64
	Scroll float64

	Bunny Bunny
	Seeds int

	Flowers  []Flower
	Planters []Planter
	Ponds    []Pond

	Released int
	Score    int
	Metres   int
	Ticks    int

	EndTicks int
	Events   []Event

	baseSeed        int64
	flight          int64
	rng             *rand.Rand
	genX            float64
	lastChunkWasGap bool
}

func New(seed int64) *Game {
	g := &Game{Screen: Attract, baseSeed: seed}
	g.resetFlight(seed)
	g.Screen = Attract
	return g
}

func (g *Game) resetFlight(seed int64) {
	g.Camera = 0
	g.Scroll = ScrollStart
	g.Bunny = Bunny{SX: 260, Y: 340, State: Flying, Perch: -1}
	g.Seeds = MaxSeeds
	g.Flowers = g.Flowers[:0]
	g.Planters = g.Planters[:0]
	g.Ponds = g.Ponds[:0]
	g.Released = 0
	g.Score = 0
	g.Metres = 0
	g.Ticks = 0
	g.EndTicks = 0
	g.Events = g.Events[:0]
	g.rng = rand.New(rand.NewSource(seed))
	g.genX = 0
	g.lastChunkWasGap = false
	g.seedInitialWorld()
	g.generateWorldAhead()
}

func (g *Game) StartFlight() {
	seed := g.baseSeed + g.flight
	g.flight++
	g.resetFlight(seed)
	g.Screen = Play
}

func (g *Game) Update(in Input) {
	g.Events = g.Events[:0]

	if g.Screen == Attract {
		if in.Flap {
			g.StartFlight()
		}
		return
	}

	if g.Bunny.State == Landed {
		if g.EndTicks > 0 {
			g.EndTicks--
		}
		if g.EndTicks == 0 {
			g.Screen = Attract
			g.Events = append(g.Events, Event{Type: EvReturnToAttract})
		}
		return
	}

	g.Ticks++
	g.Scroll = ScrollStart + (ScrollMax-ScrollStart)*math.Min(1, g.Camera/ScrollRamp)
	g.Camera += g.Scroll
	g.generateWorldAhead()

	if g.Bunny.Guard > 0 {
		g.Bunny.Guard--
	}
	if g.Bunny.FlapCooldown > 0 {
		g.Bunny.FlapCooldown--
	}

	g.applyInput(in)
	if g.Bunny.State == Perched {
		g.updatePerched()
	} else {
		g.moveBunny()
	}

	g.updatePlanters()
	g.updateSprouts()

	if g.Bunny.State == Flying {
		g.resolvePerch()
	}
	if g.Bunny.State != Perched {
		g.resolveGround()
	}

	if g.Seeds == 0 && g.Bunny.State == Flying {
		g.Bunny.State = Falling
		g.Bunny.Perch = -1
		g.Events = append(g.Events, Event{Type: EvPuffGone})
	}

	g.Metres = int(math.Floor(g.Camera / PxPerMetre))
	g.Score = g.Metres
	g.pruneWorld()
}

func (g *Game) applyInput(in Input) {
	if in.Left == in.Right || g.Bunny.State == Perched {
		g.Bunny.VX = 0
	} else if in.Left {
		g.Bunny.VX = -DriftSpeed
	} else {
		g.Bunny.VX = DriftSpeed
	}

	if !in.Flap || (g.Bunny.State != Flying && g.Bunny.State != Perched) || g.Seeds <= 0 || g.Bunny.FlapCooldown != 0 {
		return
	}

	lift := LiftForSeeds(g.Seeds)
	g.Bunny.VY = math.Max(MaxRise, g.Bunny.VY-FlapImpulse*lift)
	g.Seeds--
	g.Released++
	if g.Released%PlantEvery == 0 && len(g.Planters) < MaxPlanted {
		g.Planters = append(g.Planters, Planter{
			X:  g.Camera + g.Bunny.SX,
			Y:  g.Bunny.Y - 65,
			VX: PlanterVX,
			VY: PlanterVY,
		})
	}
	if g.Bunny.State == Perched {
		g.Bunny.State = Flying
		g.Bunny.Perch = -1
		g.Bunny.RegrowTimer = 0
		g.Events = append(g.Events, Event{Type: EvTakeOff})
	}
	g.Bunny.FlapCooldown = FlapCooldown
	g.Events = append(g.Events, Event{Type: EvFlap})
}

func (g *Game) moveBunny() {
	maxFall := MaxFallStem
	if g.Bunny.State == Flying {
		maxFall = MaxFallForSeeds(g.Seeds)
	}

	g.Bunny.VY = math.Min(maxFall, g.Bunny.VY+Gravity)
	g.Bunny.Y += g.Bunny.VY
	if g.Bunny.Y < Ceiling {
		g.Bunny.Y = Ceiling
		g.Bunny.VY = 0
	}

	g.Bunny.SX += g.Bunny.VX
	if g.Bunny.SX < MinSX {
		g.Bunny.SX = MinSX
	}
	if g.Bunny.SX > MaxSX {
		g.Bunny.SX = MaxSX
	}
}

func (g *Game) resolvePerch() {
	if g.Bunny.VY < 0 || g.Bunny.VY > PerchMaxVY {
		return
	}
	feetY := g.Bunny.Y + BunnyFeet
	for i := range g.Flowers {
		f := &g.Flowers[i]
		if !f.Bloomed || f.Nectar <= 0 {
			continue
		}
		sx := f.X - g.Camera
		headY := GroundY - f.Height
		dx := g.Bunny.SX - sx
		dy := feetY - headY
		if dx*dx+dy*dy > FlowerRadius*FlowerRadius {
			continue
		}
		g.Bunny.State = Perched
		g.Bunny.Perch = i
		g.Bunny.VX = 0
		g.Bunny.VY = 0
		g.Bunny.Y = headY - FlowerRadius - BunnyFeet
		g.Bunny.RegrowTimer = 0
		g.Events = append(g.Events, Event{Type: EvPerch})
		return
	}
}

func (g *Game) updatePerched() {
	if g.Bunny.Perch < 0 || g.Bunny.Perch >= len(g.Flowers) {
		g.Bunny.State = Flying
		g.Bunny.Perch = -1
		return
	}
	f := &g.Flowers[g.Bunny.Perch]
	g.Bunny.SX = f.X - g.Camera
	g.Bunny.Y = GroundY - f.Height - FlowerRadius - BunnyFeet
	g.Bunny.VY = 0

	if g.Bunny.SX < MinSX {
		g.Bunny.SX = MinSX
		g.Bunny.State = Flying
		g.Bunny.Perch = -1
		g.Bunny.RegrowTimer = 0
		g.Bunny.VY = -2.0
		g.Events = append(g.Events, Event{Type: EvTakeOff})
		return
	}

	if g.Seeds >= MaxSeeds || f.Nectar <= 0 {
		g.Bunny.RegrowTimer = 0
		return
	}
	g.Bunny.RegrowTimer++
	if g.Bunny.RegrowTimer >= RegrowTicks {
		g.Bunny.RegrowTimer = 0
		g.Seeds++
		f.Nectar--
		g.Events = append(g.Events, Event{Type: EvRegrow})
	}
}

func (g *Game) updatePlanters() {
	if len(g.Planters) == 0 {
		return
	}
	out := g.Planters[:0]
	for i := range g.Planters {
		p := g.Planters[i]
		p.X += g.Scroll + p.VX
		p.Y += p.VY
		if p.Y < GroundY {
			out = append(out, p)
			continue
		}
		if !g.canPlantAt(p.X) || g.plantedFlowerCount() >= MaxPlanted {
			continue
		}
		height := 90.0 + float64((g.Released+i*17)%51)
		g.Flowers = append(g.Flowers, Flower{
			X:         p.X,
			Height:    height,
			Nectar:    NectarPlanted,
			Planted:   true,
			Bloomed:   false,
			GrowTicks: SproutTicks,
		})
	}
	g.Planters = out
}

func (g *Game) updateSprouts() {
	for i := range g.Flowers {
		f := &g.Flowers[i]
		if f.Bloomed || f.GrowTicks <= 0 {
			continue
		}
		f.GrowTicks--
		if f.GrowTicks == 0 {
			f.Bloomed = true
			g.Events = append(g.Events, Event{Type: EvSprouted})
		}
	}
}

func (g *Game) canPlantAt(x float64) bool {
	for _, p := range g.Ponds {
		if x >= p.X && x <= p.X+p.Width {
			return false
		}
	}
	for _, f := range g.Flowers {
		if math.Abs(f.X-x) < 80 {
			return false
		}
	}
	return true
}

func (g *Game) plantedFlowerCount() int {
	n := 0
	for _, f := range g.Flowers {
		if f.Planted {
			n++
		}
	}
	return n
}

func (g *Game) resolveGround() {
	if g.Bunny.Y+BunnyFeet < GroundY {
		return
	}

	g.Bunny.Y = GroundY - BunnyFeet
	if g.Bunny.State == Falling {
		g.Bunny.VY = 0
		g.Bunny.State = Landed
		g.EndTicks = EndTicks
		g.Events = append(g.Events, Event{Type: EvFlightOver})
		return
	}

	g.Bunny.VY = BounceVY
	g.Events = append(g.Events, Event{Type: EvBounce})
	if g.Bunny.Guard == 0 && g.Seeds > 0 {
		lost := min(GroundSeeds, g.Seeds)
		g.Seeds -= lost
		g.Bunny.Guard = GuardTicks
		g.Events = append(g.Events, Event{Type: EvSeedsLost, Value: lost})
	}
}

func LiftForSeeds(seeds int) float64 {
	f := clampSeeds(seeds) / float64(MaxSeeds)
	return 0.45 + 0.55*f
}

func MaxFallForSeeds(seeds int) float64 {
	f := clampSeeds(seeds) / float64(MaxSeeds)
	return MaxFallFull + (MaxFallEmpty-MaxFallFull)*(1-f)
}

func MaxFlowerGap(scroll float64) float64 {
	return 0.8 * (GroundY - Ceiling - 90) / MaxFallFull * scroll
}

func clampSeeds(seeds int) float64 {
	if seeds < 0 {
		return 0
	}
	if seeds > MaxSeeds {
		return MaxSeeds
	}
	return float64(seeds)
}
