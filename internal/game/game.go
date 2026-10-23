package game

import "math"

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

	FlapCooldown int
}

type Game struct {
	Screen Screen

	Camera float64
	Scroll float64

	Bunny Bunny
	Seeds int

	Released int
	Score    int
	Metres   int
	Ticks    int

	EndTicks int
	Events   []Event
}

func New() *Game {
	g := &Game{Screen: Attract}
	g.resetFlight()
	g.Screen = Attract
	return g
}

func (g *Game) resetFlight() {
	g.Camera = 0
	g.Scroll = ScrollStart
	g.Bunny = Bunny{SX: 260, Y: 340, State: Flying}
	g.Seeds = MaxSeeds
	g.Released = 0
	g.Score = 0
	g.Metres = 0
	g.Ticks = 0
	g.EndTicks = 0
	g.Events = g.Events[:0]
}

func (g *Game) StartFlight() {
	g.resetFlight()
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

	if g.Bunny.Guard > 0 {
		g.Bunny.Guard--
	}
	if g.Bunny.FlapCooldown > 0 {
		g.Bunny.FlapCooldown--
	}

	g.applyInput(in)
	g.moveBunny()
	g.resolveGround()

	if g.Seeds == 0 && g.Bunny.State == Flying {
		g.Bunny.State = Falling
		g.Events = append(g.Events, Event{Type: EvPuffGone})
	}

	g.Metres = int(math.Floor(g.Camera / PxPerMetre))
	g.Score = g.Metres
}

func (g *Game) applyInput(in Input) {
	if in.Left == in.Right {
		g.Bunny.VX = 0
	} else if in.Left {
		g.Bunny.VX = -DriftSpeed
	} else {
		g.Bunny.VX = DriftSpeed
	}

	if !in.Flap || g.Bunny.State != Flying || g.Seeds <= 0 || g.Bunny.FlapCooldown != 0 {
		return
	}

	lift := LiftForSeeds(g.Seeds)
	g.Bunny.VY = math.Max(MaxRise, g.Bunny.VY-FlapImpulse*lift)
	g.Seeds--
	g.Released++
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

func (g *Game) resolveGround() {
	const feet = 22.0
	if g.Bunny.Y+feet < GroundY {
		return
	}

	g.Bunny.Y = GroundY - feet
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
		g.Seeds = max(0, g.Seeds-GroundSeeds)
		g.Bunny.Guard = GuardTicks
		g.Events = append(g.Events, Event{Type: EvSeedsLost, Value: GroundSeeds})
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

func clampSeeds(seeds int) float64 {
	if seeds < 0 {
		return 0
	}
	if seeds > MaxSeeds {
		return MaxSeeds
	}
	return float64(seeds)
}
