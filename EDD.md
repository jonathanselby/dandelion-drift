# Dandelion Drift — Engineering Design Document

**Version:** 0.1
**Status:** Draft for V0.1.0
**Implementation:** Go 1.25+, Ebitengine 2.10.5

## 1. Purpose

Dandelion Drift is one small endless flight game. A bunny floats under a dandelion puff that is her lift, her health and, through the seeds she spends, the source of future landing spots.

> **Keep the game as small as the game actually is.**

Richness comes from four interacting rules: flaps spend seeds, seeds set lift and glide, flowers refill seeds, and spent seeds plant flowers. Hazards only add pressure to that loop.

## 2. Design constraints

1. One authoritative game state.
2. One explicit, single-threaded update loop at a fixed 60 ticks per second.
3. Deterministic gameplay from run seed and input.
4. No abstraction before a demonstrated need.
5. Bounded gameplay and cosmetic resources.
6. Rendering and sound react to state and events; they own no game truth.
7. Every gameplay rule testable without opening a window.
8. Visual polish comes from art and animation, not architecture.

## 3. Runtime dependency

Ebitengine for the window, fixed update loop, keyboard, 2D drawing, text and audio. Nothing else.

```text
Ebitengine
    |
    v
thin adapter (cmd/dandeliondrift)
    |
    v
Game (internal/game): plain Go rules
```

## 4. Units and coordinates

| Quantity | Unit |
| --- | --- |
| time | ticks (60 per second) |
| horizontal position | world pixels, increasing to the right |
| vertical position | screen pixels, 0 at top (no vertical scrolling) |
| velocity | pixels per tick |
| distance shown | metres; 1 m = 40 px |

The logical screen is 960×720.

```text
ceiling        y = 60
ground surface y = 640   (grass or water)
camera         world x of the screen's left edge, advanced by Scroll each tick
bunny          screen x between 120 and 600; world x = camera + screen x
```

The bunny's position is the centre of her body. The puff is drawn above her, centred 56 px higher; the puff's collision circle (radius 30) and the body's collision box (36×44) are both tested against hazards.

## 5. Authoritative state

```go
type Game struct {
    Screen Screen // Attract, Play, Initials

    Camera float64 // world x of screen left
    Scroll float64 // px per tick

    Bunny Bunny
    Seeds int // 0..24

    Flowers   []Flower
    Planters  []Planter  // released seeds that will plant
    Bees      []Bee
    Thistles  []Thistle
    Ponds     []Pond     // each with its frog
    Gusts     []Gust
    Fireflies []Firefly

    genX float64 // world x up to which the world has been generated

    Released  int // seeds released this flight, for the 1-in-4 planting rule
    Chain     int // firefly multiplier 1..5
    ChainTicks int

    Score, Metres int
    Ticks         int

    EndTicks int    // end-of-flight overlay countdown
    Scores   HighScores
    Initials string

    Events []Event

    RNG         *rand.Rand // gameplay: world generation only
    CosmeticRNG *rand.Rand
    Particles   []Particle
}
```

The gameplay RNG is used only by world generation. Bee bob, frog timing and planting are derived from ticks and counters, so a run replays exactly from seed and input. Cosmetic seeds, petals and splashes use the cosmetic RNG.

## 6. The bunny

```go
type Bunny struct {
    SX, Y  float64 // screen x, screen y
    VX, VY float64
    State  BunnyState // Flying, Perched, Falling (puff gone), Landed
    Perch  int        // index of perched flower, -1 if none
    Guard  int        // ticks of knock protection remaining
    FlapCooldown int
    RegrowTimer  int
}
```

## 7. Concepts deliberately kept as values

```text
seeds        = int (the puff IS this number)
lift, glide  = functions of seeds
scroll       = float64, function of distance
chain        = int
score        = int
```

No `Puff`, `Physics`, `Camera`, `Spawner` or `Level` types.

## 8. Tuning constants (starting values)

```go
const (
    TPS        = 60
    PxPerMetre = 40.0
    Ceiling    = 60.0
    GroundY    = 640.0
    MinSX, MaxSX = 120.0, 600.0

    MaxSeeds = 24

    ScrollStart = 1.5
    ScrollMax   = 3.0
    ScrollRamp  = 30000.0 // px of distance to reach ScrollMax (750 m)

    Gravity      = 0.10
    FlapImpulse  = 3.0   // scaled by lift
    FlapCooldown = 8
    MaxRise      = -4.0
    DriftSpeed   = 1.6

    MaxFallFull  = 1.0 // terminal fall with a full puff
    MaxFallEmpty = 2.4 // terminal fall with one seed
    MaxFallStem  = 2.0 // no puff, floating down on the bare stem

    PerchMaxVY   = 2.0
    RegrowTicks  = 15  // per seed while perched
    NectarWild   = 10
    NectarPlanted = 6

    PlantEvery   = 4   // one released seed in four plants
    PlanterVX    = 0.4 // world px/tick ahead of scroll
    PlanterVY    = 0.6
    SproutTicks  = 240
    MaxPlanted   = 6

    BeeSeeds, ThistleSeeds, GroundSeeds, WaterSeeds, FrogSeeds = 3, 3, 1, 4, 6
    GuardTicks   = 60
    FrogGuard    = 90
    BounceVY     = -2.5
    SplashVY     = -3.5
    FlingVY      = -4.0

    ChainWindow  = 120
    ChainMax     = 5
    FireflyPoints = 10

    EndTicks     = 150
    MaxParticles = 256
)
```

Tests assert relationships (sections 15 and 22), not these numbers.

## 9. Update loop

```text
clear Events
    |
if Landed: count down the end overlay; route to Initials or Attract; return
    |
Scroll = ScrollStart + (ScrollMax - ScrollStart) * min(1, Camera / ScrollRamp)
Camera += Scroll
    |
generate world ahead while genX < Camera + 2,000
    |
input: flap, drift
    |
bunny motion (or perch carry and regrow)
    |
planters fall; sprouts bloom
    |
bees, frogs move
    |
collisions: flowers (perch), hazards (knocks), fireflies
    |
seeds == 0 -> Falling; Falling touches ground -> Landed
    |
metres, score, chain timer
    |
prune entities behind Camera - 200
    |
cosmetic particles
```

No goroutines, channels, event bus or actors.

## 10. Lift and glide

```text
f      = Seeds / MaxSeeds
lift   = 0.45 + 0.55 * f
maxFall = MaxFallFull + (MaxFallEmpty - MaxFallFull) * (1 - f)
```

### Flap

Space just pressed, `State == Flying` or `Perched`, `Seeds > 0`, `FlapCooldown == 0`:

```text
VY = max(MaxRise, VY - FlapImpulse * lift)
Seeds--; Released++
if Released % PlantEvery == 0: add Planter at bunny world position
if Perched: State = Flying, Perch = -1, emit TakeOff
FlapCooldown = 8; emit Flap
```

Lift is computed **before** the seed is spent, so the flap that empties the puff still lifts.

### Fall

```text
VY = min(maxFall, VY + Gravity + gust)
Y += VY; clamp at Ceiling (VY = 0 there)
SX += VX (Left/Right sets VX to ±DriftSpeed, else 0); clamp to MinSX..MaxSX
```

## 11. Perching

### Landing on a flower

A flower head is a circle of radius 28 at `(flower.X, GroundY - flower.Height)`. The bunny perches when:

```text
State == Flying
VY >= 0 and VY <= PerchMaxVY
bunny's feet (Y + 22) are within the head circle
flower is bloomed and has Nectar > 0
```

Then `State = Perched`, `VY = 0`, the bunny's feet snap to the top of the head, emit `Perch`.

### While perched

```text
SX = flower.X - Camera            // carried left with the world
every RegrowTicks: if Seeds < MaxSeeds and Nectar > 0: Seeds++, Nectar--, emit Regrow
Nectar == 0: flower wilts (drawn drooping); stays as a perch until she leaves
SX < MinSX: auto hop: State = Flying, VY = -2.0, emit TakeOff
```

A spent flower cannot be perched again.

## 12. Planting

```go
type Planter struct{ X, Y, VX, VY float64 } // world x, screen y
```

Each tick: `X += Scroll + PlanterVX`, `Y += PlanterVY`. At `Y >= GroundY`:

- over grass, and fewer than `MaxPlanted` planted flowers alive → Sprout at `X`, height 90–140 px from a counter-derived pattern, blooms after `SproutTicks`, `Nectar = NectarPlanted`, emit `Sprouted` on bloom;
- over water, inside a thistle, or within 80 px of another flower → lost.

Every released seed also spawns a cosmetic seed particle, so the child sees all seeds blow away; only one in four becomes gameplay.

## 13. Hazards

All knocks are ignored while `Guard > 0`. A knock: `Seeds = max(0, Seeds - n)`, `Guard = GuardTicks`, emit the hazard's event and `SeedsLost(n)`.

| Hazard | Geometry | Behaviour |
| --- | --- | --- |
| Bee | circle r = 18; world `VX = -1.0`; `Y = baseY + 30·sin(ticks/30 + phase)` | knock `BeeSeeds`; bunny pushed `VX = -2` for 10 ticks |
| Thistle | rectangle 30 wide, height 80–180 from the ground | knock `ThistleSeeds`; `VY = BounceVY` |
| Grass | ground outside ponds, flowers and thistles | knock `GroundSeeds`; `VY = BounceVY` |
| Water | ground inside a pond span | knock `WaterSeeds`; `VY = SplashVY`; emit `Splash` |
| Frog | sits at pond centre; leap arc `VY0 = -6`, gravity 0.25 | leaps when bunny's world x is within ±80 px and `Y > GroundY - 180`, at most once per 150 ticks; touching the leaping frog knocks `FrogSeeds`, `VY = FlingVY`, `Guard = FrogGuard` |
| Gust | world x span, all heights | adds −0.06 (updraft) or +0.05 (downdraft) to gravity while inside |

Grass and water knocks ignore `Guard` for the bounce (she must never sink into the ground) but not for seed loss.

## 14. Fireflies and scoring

Firefly: circle r = 20. Collecting:

```text
if ChainTicks > 0: Chain = min(ChainMax, Chain + 1) else Chain = 1
Score += FireflyPoints * Chain
ChainTicks = ChainWindow
emit Firefly(Chain)
```

`ChainTicks` counts down each tick. Metres = floor(Camera / PxPerMetre); total score = Metres + firefly points.

## 15. World generation

The world is generated in chunks ahead of the camera from the gameplay RNG.

| Chunk | Width (px) | Content |
| --- | --- | --- |
| Meadow | 600 | 1–2 flowers, optional fireflies |
| Gap | 400 | grass only |
| ThistlePatch | 500 | 2–3 thistles |
| BeeLine | 600 | 2–4 bees at staggered heights |
| Pond | 400–600 | water span and a frog |
| Updraft | 400 | updraft gust, firefly column |
| Downdraft | 300 | downdraft gust |
| FireflyArc | 500 | 6 fireflies in an arc |

Weights shift with distance: early worlds are mostly Meadow, Gap and FireflyArc; hazards thicken toward 750 m and then hold steady.

### Invariants

1. **Gentle start.** The first 1,200 px contain one flower and no hazards.
2. **Reachable flowers (FR10).** The gap between consecutive generated flower heads never exceeds

   ```text
   MaxGap = 0.8 * (GroundY - Ceiling - 90) / MaxFallFull * Scroll
          = 392 * Scroll  -> 588 px at the start, 1,176 px at full speed
   ```

   which is 80% of a no-flap full-puff glide from the ceiling to the lowest flower head. Planted flowers only add to this.
3. **Clear approach.** No bee path, thistle or gust within 120 px horizontally of a flower head.
4. **Passable bees.** A BeeLine leaves a vertical gap of at least 160 px between bee paths.
5. **Bounded ponds.** Ponds are at most 600 px wide and never directly follow a Gap.

The generator rejects and redraws a chunk that would break an invariant; after 64 failed redraws it places a Meadow, which is always valid.

## 16. End of flight

```text
Seeds == 0 (after a flap or a knock): State = Falling, emit PuffGone
Falling: no flap; maxFall = MaxFallStem; hazards no longer knock
Falling touches ground or water: State = Landed, EndTicks = 150, emit FlightOver
EndTicks reaches 0: Initials if score qualifies (emit HighScore), else Attract
```

Landing in water is drawn on a lily pad. The end is always soft.

## 17. Screens

```text
ATTRACT   title, top five, demo bunny; Space starts a flight
PLAY      flight, with the end-of-flight overlay
INITIALS  1-3 letters, Enter confirms (Enter with no letters saves "?")
```

Esc quits from any screen. The demo bunny is the real `Game` running a fixed seed with a scripted flap pattern, so the attract screen also exercises the rules.

The attract screen and end-of-flight presentation are first-class graphical scenes. They must be deliberately composed and polished to the same standard as gameplay. A finished flight must visibly conclude, then route back to the attract screen so a new flight starts from a clean state. This is a retained Astro Answers UAT lesson, not optional polish.

## 18. Input

```go
type Input struct {
    Flap        bool // Space just pressed
    Left, Right bool // held
    Enter       bool
    Backspace   bool
    Chars       []rune
}
```

## 19. Events and sound

```go
const (
    EvFlap Event = iota
    EvTakeOff
    EvPerch
    EvRegrow
    EvSprouted
    EvBee
    EvThistle
    EvBounce
    EvSplash
    EvFrogLeap
    EvFrogCatch
    EvFirefly    // with chain level, pitch rises with chain
    EvPuffGone
    EvFlightOver
    EvHighScore
)
```

About 13 WAV cues, 44.1 kHz mono, embedded. Firefly chain plays the same chime at five pre-rendered pitches rather than adding a synthesiser. Missing sounds degrade gracefully; missing sprites fail startup visibly.

## 20. Rendering

The renderer reads state and draws; it decides nothing.

### Code-drawn

- sky gradient, drifting clouds, two parallax hill layers;
- grass strip, pond water with ripples;
- **the puff**: a central seed head with exactly `Seeds` tufts radiating evenly, re-spaced as seeds go; this is the health display and must never need digits;
- flower stems, sprouts growing, wilted flowers drooping;
- wind streaks in gust zones;
- fireflies as soft glows; chain multiplier pops as "×3";
- particles: blown seeds on every flap, petals on perch, splashes, regrow sparkles;
- distance in metres and score;
- a deliberately composed attract/start scene using game art, not a developer-menu presentation;
- a visually finished end-of-flight scene that keeps the completed run visible, presents the result clearly, and returns to attract.

### Raster assets (five PNGs)

| Asset | Content |
| --- | --- |
| `bunny.png` | sheet: hang, flap, perched, landed |
| `bee.png` | two wing frames |
| `thistle.png` | thistle head and leaves (stem stretched in code) |
| `flower.png` | flower head |
| `frog.png` | sitting, leaping |

## 21. Persistence

One local `scores.json`, single top-five table, temp-file then rename. Absent starts empty; malformed reports and starts empty. Carried over from Astro Answers' `scores.go`.

## 22. Testing

All gameplay tests run headless in `internal/game`.

### Flight
- `TestFlapRisesAndSpendsOneSeed`.
- `TestLiftAndGlideMonotonicInSeeds`: more seeds → stronger lift, slower max fall.
- `TestLastFlapStillLifts`.
- `TestCeilingClamp`.
- `TestDriftBounds`.

### Perching
- `TestPerchOnlyOnGentleDescent`: rising or fast-falling bunny does not perch.
- `TestRegrowUntilFullOrNectarSpent`.
- `TestPerchCarriesLeftAndAutoHops`.
- `TestSpentFlowerCannotBePerched`.

### Planting
- `TestOneInFourSeedsPlants`.
- `TestPlantingOnGrassOnly`.
- `TestSproutBloomsAfterDelay`.
- `TestPlantedFlowerCap`.

### Hazards
- `TestEachHazardRemovesItsSeeds`.
- `TestGuardPreventsDoubleKnock`.
- `TestGroundAlwaysBouncesEvenWhenGuarded`.
- `TestFrogLeapsOnlyWhenBunnyLowOverhead`.
- `TestGustsChangeFall`.

### Fireflies and score
- `TestChainRisesWithinWindowAndResets`.
- `TestChainCapsAtFive`.
- `TestScoreIsMetresPlusFireflies`.

### End of flight
- `TestEmptyPuffFallsSoftlyAndEnds`.
- `TestEndRoutesToInitialsOrAttract`.
- `TestCompletedFlightReturnsAndStartsFresh`.

### Generation
- `TestGenerationDeterministicFromSeed`.
- `TestInvariantsHoldOverLongFlights`: invariants 1–5 over 3,000 m for seeds 1–200.
- `TestNoFlapGlideReachesNextFlower`: from the ceiling with a full puff and no input, every next flower head is reachable at the scroll speed in force.

### Determinism
- `TestReplayDeterministic`: same seed and inputs → identical state after 5,000 ticks.
- `TestCosmeticRNGDoesNotAffectGameplay`.
- `TestEventsMatchStateChanges`.

Score persistence tests carried over from Astro Answers.

## 23. Resource budgets

```text
live gameplay entities  <= 120 (world window ~3,200 px)
planted flowers         <= 6
planters in flight      <= 6
particles               <= 256, overflow dropped
events per tick         <= 8
```

Entities behind `Camera − 200` are pruned every tick, so memory is flat however long the flight.

## 24. Repository shape

```text
cmd/
    dandeliondrift/
        main.go      Ebitengine adapter, rendering, audio, input
        assets/      5 PNG, ~13 WAV
internal/
    game/
        game.go      state, update loop, flight, perch, planting, hazards
        world.go     chunk catalogue, generation, invariants
        scores.go
        game_test.go
        world_test.go
        scores_test.go
```

## 25. Failure behaviour

| Failure | Behaviour |
| --- | --- |
| score file absent | start empty |
| score file malformed | report, start empty |
| sound missing | continue silently |
| sprite missing | fail startup visibly |
| generator redraw limit | place Meadow (always valid) |
| entity cap reached | generation pauses until pruning frees space |
| programming error | fail visibly |

## 26. CI and release

Copy Astro Answers' workflow: module tidy contract, `gofmt`, `go vet`, `go test -race -count=1`, Linux/amd64 and Linux/ARM64 builds, named core acceptance tests, AppImage. Then the generated EDD and comparison pinned to the release commit.

## 27. Deliberately absent

```text
physics engine       particle engine        camera type
entity component system   scene graph       spawner framework
lives                power-ups              music
gamepad / touch      network / telemetry / AI
```

## 28. Design conclusion

The puff is the whole game:

- flapping spends it;
- its size decides how well she flies;
- flowers give it back;
- what she spends grows into the flowers she will need.

Lift, health and future landing spots share one resource. The red-team test is unchanged:

> **Does every piece of machinery remove more complexity than it introduces?**
