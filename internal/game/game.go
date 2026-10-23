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

	MaxSeeds = 240

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

	BeeSeeds       = 3
	ThistleSeeds   = 3
	GroundSeeds    = 1
	WaterSeeds     = 4
	FrogSeeds      = 6
	GuardTicks     = 60
	FrogGuard      = 90
	BounceVY       = -2.5
	SplashVY       = -3.5
	FlingVY        = -4.0
	BeePushTicks   = 10
	FrogLeapWait   = 150
	UpdraftForce   = -0.06
	DowndraftForce = 0.05

	ChainWindow   = 120
	ChainMax      = 5
	FireflyPoints = 10
	FireflySeeds  = 24

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
	EvBee
	EvThistle
	EvSplash
	EvFrogLeap
	EvFrogCatch
	EvFirefly
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
	SX, Y        float64
	VX, VY       float64
	State        BunnyState
	Guard        int
	PushTicks    int
	Perch        int
	RegrowTimer  int
	FlapCooldown int
}

type Flower struct {
	X, Height        float64
	Nectar           int
	Planted, Bloomed bool
	GrowTicks        int
}
type Planter struct{ X, Y, VX, VY float64 }
type Bee struct{ X, BaseY, Phase float64 }
type Thistle struct{ X, Height float64 }
type Frog struct {
	X, Y, VY float64
	Leaping  bool
	LastLeap int
}
type Pond struct {
	X, Width float64
	Frog     Frog
}
type Gust struct{ X, Width, Force float64 }
type Firefly struct {
	X, Y      float64
	Collected bool
}

type Game struct {
	Screen               Screen
	Camera, Scroll       float64
	Bunny                Bunny
	Seeds                int
	Flowers              []Flower
	Planters             []Planter
	Bees                 []Bee
	Thistles             []Thistle
	Ponds                []Pond
	Gusts                []Gust
	Fireflies            []Firefly
	Released             int
	Chain, ChainTicks    int
	FireflyScore         int
	FirefliesCollected   int
	Score, Metres, Ticks int
	EndTicks             int
	Events               []Event
	baseSeed, flight     int64
	rng                  *rand.Rand
	genX                 float64
	lastFlowerX          float64
	lastChunk            chunkKind
}

func New(seed int64) *Game {
	g := &Game{Screen: Attract, baseSeed: seed}
	g.resetFlight(seed)
	g.Screen = Attract
	return g
}

func (g *Game) resetFlight(seed int64) {
	g.Camera, g.Scroll = 0, ScrollStart
	g.Bunny = Bunny{SX: 260, Y: 340, State: Flying, Perch: -1}
	g.Seeds = MaxSeeds
	g.Flowers = g.Flowers[:0]
	g.Planters = g.Planters[:0]
	g.Bees = g.Bees[:0]
	g.Thistles = g.Thistles[:0]
	g.Ponds = g.Ponds[:0]
	g.Gusts = g.Gusts[:0]
	g.Fireflies = g.Fireflies[:0]
	g.Released, g.Chain, g.ChainTicks, g.FireflyScore = 0, 0, 0, 0
	g.FirefliesCollected = 0
	g.Score, g.Metres, g.Ticks, g.EndTicks = 0, 0, 0, 0
	g.Events = g.Events[:0]
	g.rng = rand.New(rand.NewSource(seed))
	g.genX = 0
	g.lastFlowerX = 0
	g.lastChunk = chunkMeadow
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
	if g.ChainTicks > 0 {
		g.ChainTicks--
		if g.ChainTicks == 0 {
			g.Chain = 0
		}
	}
	g.Scroll = ScrollStart + (ScrollMax-ScrollStart)*math.Min(1, g.Camera/ScrollRamp)
	g.Camera += g.Scroll
	g.generateWorldAhead()
	if g.Bunny.Guard > 0 {
		g.Bunny.Guard--
	}
	if g.Bunny.FlapCooldown > 0 {
		g.Bunny.FlapCooldown--
	}
	if g.Bunny.PushTicks > 0 {
		g.Bunny.PushTicks--
	}

	g.applyInput(in)
	if g.Bunny.State == Perched {
		g.updatePerched()
	} else {
		g.moveBunny()
	}
	g.updatePlanters()
	g.updateSprouts()
	g.updateBees()
	g.updateFrogs()

	if g.Bunny.State == Flying {
		g.resolvePerch()
	}
	hitThistle := false
	if g.Bunny.State == Flying {
		hitThistle = g.resolveHazards()
		g.resolveFireflies()
	}
	if g.Bunny.State != Perched && !hitThistle {
		g.resolveGround()
	}

	if g.Seeds == 0 && g.Bunny.State == Flying {
		g.Bunny.State = Falling
		g.Bunny.Perch = -1
		g.Events = append(g.Events, Event{Type: EvPuffGone})
	}
	g.Metres = int(math.Floor(g.Camera / PxPerMetre))
	g.Score = g.Metres + g.FireflyScore
	g.pruneWorld()
}

func (g *Game) applyInput(in Input) {
	if g.Bunny.PushTicks > 0 {
		g.Bunny.VX = -2
	} else if in.Left == in.Right || g.Bunny.State == Perched {
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
		g.Planters = append(g.Planters, Planter{X: g.Camera + g.Bunny.SX, Y: g.Bunny.Y - 65, VX: PlanterVX, VY: PlanterVY})
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
	gravity := Gravity
	if g.Bunny.State == Flying {
		maxFall = MaxFallForSeeds(g.Seeds)
		gravity += g.gustAtBunny()
	}
	g.Bunny.VY = math.Min(maxFall, g.Bunny.VY+gravity)
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
		g.Bunny.VY = -2
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
		h := 90 + float64((g.Released+i*17)%51)
		g.Flowers = append(g.Flowers, Flower{X: p.X, Height: h, Nectar: NectarPlanted, Planted: true, GrowTicks: SproutTicks})
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
	for _, t := range g.Thistles {
		if math.Abs(t.X-x) < 35 {
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

func (g *Game) updateBees() {
	for i := range g.Bees {
		g.Bees[i].X -= 1.0
	}
}
func (g *Game) BeeY(b Bee) float64 { return b.BaseY + 30*math.Sin(float64(g.Ticks)/30+b.Phase) }
func (g *Game) updateFrogs() {
	for i := range g.Ponds {
		p := &g.Ponds[i]
		f := &p.Frog
		f.X = p.X + p.Width/2
		if !f.Leaping {
			if g.Bunny.State == Flying && math.Abs((g.Camera+g.Bunny.SX)-f.X) <= 80 && g.Bunny.Y > GroundY-180 && g.Ticks-f.LastLeap >= FrogLeapWait {
				f.Leaping = true
				f.Y = GroundY - 18
				f.VY = -6
				f.LastLeap = g.Ticks
				g.Events = append(g.Events, Event{Type: EvFrogLeap})
			}
			continue
		}
		f.VY += 0.25
		f.Y += f.VY
		if f.Y >= GroundY-18 {
			f.Y = GroundY - 18
			f.VY = 0
			f.Leaping = false
		}
	}
}

func (g *Game) resolveHazards() bool {
	wx := g.Camera + g.Bunny.SX
	for _, b := range g.Bees {
		if circleHit(g.Bunny.SX, g.Bunny.Y, 25, b.X-g.Camera, g.BeeY(b), 18) || (g.Seeds > 0 && circleHit(g.Bunny.SX, g.Bunny.Y-60, 30, b.X-g.Camera, g.BeeY(b), 18)) {
			if g.knock(BeeSeeds, GuardTicks, EvBee) {
				g.Bunny.PushTicks = BeePushTicks
			}
			break
		}
	}
	for _, t := range g.Thistles {
		if math.Abs(wx-t.X) <= 30 && g.Bunny.Y+BunnyFeet >= GroundY-t.Height {
			if g.knock(ThistleSeeds, GuardTicks, EvThistle) {
				g.Bunny.VY = BounceVY
				g.Bunny.Y = math.Min(g.Bunny.Y, GroundY-t.Height-BunnyFeet)
				return true
			}
		}
	}
	for i := range g.Ponds {
		f := &g.Ponds[i].Frog
		if f.Leaping && circleHit(g.Bunny.SX, g.Bunny.Y, 25, f.X-g.Camera, f.Y, 24) {
			if g.knock(FrogSeeds, FrogGuard, EvFrogCatch) {
				g.Bunny.VY = FlingVY
			}
			break
		}
	}
	return false
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
	wx := g.Camera + g.Bunny.SX
	for _, p := range g.Ponds {
		if wx >= p.X && wx <= p.X+p.Width {
			g.Bunny.VY = SplashVY
			if g.Bunny.Guard > 0 {
				g.Events = append(g.Events, Event{Type: EvSplash})
			} else {
				g.knock(WaterSeeds, GuardTicks, EvSplash)
			}
			return
		}
	}
	g.Bunny.VY = BounceVY
	if g.Bunny.Guard > 0 {
		g.Events = append(g.Events, Event{Type: EvBounce})
	} else {
		g.knock(GroundSeeds, GuardTicks, EvBounce)
	}
}

func (g *Game) knock(n, guard int, ev EventType) bool {
	if g.Bunny.Guard > 0 || g.Seeds <= 0 {
		return false
	}
	lost := min(n, g.Seeds)
	g.Seeds -= lost
	g.Bunny.Guard = guard
	g.Events = append(g.Events, Event{Type: ev}, Event{Type: EvSeedsLost, Value: lost})
	return true
}
func circleHit(x1, y1, r1, x2, y2, r2 float64) bool {
	dx := x1 - x2
	dy := y1 - y2
	rr := r1 + r2
	return dx*dx+dy*dy <= rr*rr
}
func (g *Game) gustAtBunny() float64 {
	wx := g.Camera + g.Bunny.SX
	for _, u := range g.Gusts {
		if wx >= u.X && wx <= u.X+u.Width {
			return u.Force
		}
	}
	return 0
}

func (g *Game) resolveFireflies() {
	for i := range g.Fireflies {
		f := &g.Fireflies[i]
		if f.Collected {
			continue
		}
		if !circleHit(g.Bunny.SX, g.Bunny.Y, 25, f.X-g.Camera, f.Y, 20) {
			continue
		}
		f.Collected = true
		if g.ChainTicks > 0 {
			g.Chain = min(ChainMax, g.Chain+1)
		} else {
			g.Chain = 1
		}
		g.ChainTicks = ChainWindow
		pts := FireflyPoints * g.Chain
		g.FireflyScore += pts
		g.FirefliesCollected++
		g.Seeds = min(MaxSeeds, g.Seeds+FireflySeeds)
		g.Events = append(g.Events, Event{Type: EvFirefly, Value: g.Chain})
	}
}

func LiftForSeeds(seeds int) float64 { f := clampSeeds(seeds) / float64(MaxSeeds); return .45 + .55*f }
func MaxFallForSeeds(seeds int) float64 {
	f := clampSeeds(seeds) / float64(MaxSeeds)
	return MaxFallFull + (MaxFallEmpty-MaxFallFull)*(1-f)
}
func ScrollAtWorldX(x float64) float64 {
	return ScrollStart + (ScrollMax-ScrollStart)*math.Min(1, math.Max(0, x)/ScrollRamp)
}
func MaxFlowerGap(scroll float64) float64 {
	return .8 * (GroundY - Ceiling - 90) / MaxFallFull * scroll
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
