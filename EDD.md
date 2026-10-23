# Dandelion Drift — Engineering Design Document

**Version:** 0.3  
**Status:** Final — V0.1.0, red-teamed against Garage engineering rules and reconciled against the approved implementation  
**Implementation:** Go 1.25+, Ebitengine 2.10.5

## 1. Purpose

Dandelion Drift is one small endless flight game. A bunny floats beneath a
dandelion puff that is both movement reserve and health. Flowers restore it.
Hazards and fireflies drain it. Released flap seeds can grow into future
flowers.

The governing engineering rule is:

> **Keep the game as small as the game actually is.**

The implementation is not a reusable game engine. It is one game whose
richness comes from a few interacting rules: flapping spends the reserve,
reserve level changes flight, flowers restore it, released seeds alter the
future world, and hazards apply pressure.

The core knows nothing about sprites, sound or initials. A bunny has a
position, a velocity and a state; a flower has a position, a height and some
nectar.

## 2. Design constraints

The design follows these constraints:

1.  one authoritative gameplay state;
2.  one explicit single-threaded update loop at 60 ticks per second;
3.  deterministic gameplay from run seed and input;
4.  no abstraction before a demonstrated need;
5.  bounded gameplay and cosmetic resources;
6.  renderer and sound own no gameplay decisions;
7.  every gameplay rule is testable without opening a window;
8.  high-score entry and cosmetic presentation stay outside gameplay truth;
9.  visual quality comes from art, animation and composition rather than
    frameworks;
10. release packaging must remain simpler than the game.

If the implementation becomes materially more complicated than this
document, first ask what can be deleted.

## 3. Runtime dependency

Use Ebitengine for:

-   window and fixed game loop;
-   keyboard input;
-   2D drawing and transforms;
-   text;
-   audio output.

Go's standard library provides persistence, deterministic RNG and the
remaining runtime services. The `gofont` bold face from `golang.org/x/image`
supplies text.

Do not add another game, UI, physics, animation or audio framework.

``` text
Ebitengine
    |
    v
thin adapter (cmd/dandeliondrift)
    |
    v
Game (internal/game)
    |
    v
plain deterministic Go rules
```

## 4. Units and coordinates

| Quantity | Unit |
| --- | --- |
| time | ticks, 60 per second |
| horizontal world position | pixels increasing to the right |
| horizontal bunny position | screen pixels (`SX`) |
| vertical position | screen pixels, 0 at top |
| velocity | pixels per tick |
| displayed distance | metres, 40 px = 1 m |

Logical playfield: **960×720**.

The world camera advances rightward while scenery is rendered moving left.
World X minus `Camera` gives screen X. The bunny's horizontal position is held
in screen space and bounded within the playfield, so it never needs to be
re-based as the camera moves.

## 5. Authoritative gameplay state

Gameplay truth lives in one `Game` in `internal/game`.

Conceptually:

``` go
type Game struct {
    Screen Screen

    Camera, Scroll float64
    Bunny          Bunny
    Seeds          int

    Flowers   []Flower
    Planters  []Planter
    Bees      []Bee
    Thistles  []Thistle
    Ponds     []Pond   // each pond owns its frog
    Gusts     []Gust
    Fireflies []Firefly

    Released           int
    Chain, ChainTicks  int
    FireflyScore       int
    FirefliesCollected int
    Score, Metres      int
    Ticks, EndTicks    int

    Events []Event

    rng *rand.Rand
    // generation cursors: genX, lastFlowerX, lastChunk
}
```

This is not a required final field layout. It states the design boundary:
gameplay truth lives in one place.

The core has only two screens, `Attract` and `Play`. High-score initials
entry is presentation/application state in the Ebitengine adapter because it
does not alter game rules.

Rendering and sound read or react to this state. They do not own gameplay
truth.

## 6. Bunny

The bunny is plain data.

``` go
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
```

There are four states:

``` go
const (
    Flying BunnyState = iota
    Perched
    Falling
    Landed
)
```

State determines which movement rule runs, whether flaps are accepted,
whether hazards apply and whether the flight has ended.

The bunny's position is the centre of the body. Collision geometry is simple
circles and horizontal/vertical extents around the visible body and puff.
There is no general collision engine.

`Perch` is the index of the flower being perched on, or `-1`. Pruning the
world remaps that index when flowers behind the camera are removed. This is
the only place an index into a world slice survives across a tick, and it is
kept because a perched bunny always refers to exactly one flower.

## 7. Concepts deliberately kept as values

These do not need objects or subsystems:

``` text
puff reserve    = integer 0..50
guard           = tick counter
bee push        = tick counter
flap cooldown   = tick counter
regrow timer    = tick counter
chain           = integer 1..5 + tick counter
nectar          = integer per flower
sprout          = tick counter per flower
frog            = value inside its pond
distance        = derived from Camera
score           = derived each tick
```

The puff does not need a `Puff` type.

Health does not need a separate value: the puff is health.

The frog does not need its own world list: it exists only inside a pond.

Score does not need a scoring service.

### Events

`Events` is a per-tick slice cleared at the start of every update. Rules
append small `{Type, Value}` records — flap, perch, regrow, hazard type,
puffs lost, firefly chain, flight over — and the adapter reads them once to
play sounds and spawn cosmetic effects.

This is not an event bus. There are no subscribers, no queues, no
cross-tick delivery and no gameplay rule reacts to an event. It exists so the
adapter does not have to diff consecutive `Game` snapshots to discover what
just happened.

## 8. Adapter-owned state

`cmd/dandeliondrift` owns things that affect presentation but not gameplay
truth:

-   Ebitengine images and font faces;
-   cosmetic seed particles and effect bursts;
-   procedural audio players;
-   persisted high-score table and its path;
-   initials-entry text and mode;
-   the deterministic attract-demo game;
-   animation frame counters.

This separation is deliberate. Moving initials, particles or sound machinery
into `Game` would make the deterministic core less clear without improving
behaviour.

## 9. Core tuning constants

The final V0.1 release uses:

``` text
Timing and space
  TPS               60
  logical size      960 × 720
  PxPerMetre        40
  Ceiling           60 px
  GroundY           640 px
  MinSX / MaxSX     120 / 600 px

Reserve
  MaxSeeds          50

Scroll
  ScrollStart       1.5 px/tick
  ScrollMax         3.0 px/tick
  ScrollRamp        30,000 px of world (750 m)

Flight
  Gravity           0.10 px/tick²
  FlapImpulse       3.0 px/tick
  MaxRise           -4.0 px/tick
  FlapCooldown      8 ticks
  DriftSpeed        1.6 px/tick
  MaxFallFull       1.0 px/tick
  MaxFallEmpty      2.4 px/tick
  MaxFallStem       2.0 px/tick

Perching
  BunnyFeet         22 px
  FlowerRadius      28 px
  PerchMaxVY        2.0 px/tick
  RegrowTicks       15 ticks per puff
  NectarWild        10
  NectarPlanted     6

Planting
  PlantEvery        4 released flap seeds
  PlanterVX / VY    0.4 / 0.6 px/tick
  SproutTicks       240
  MaxPlanted        6

Hazards
  Bee damage        3
  Thistle damage    3
  Ground damage     1
  Water damage      4
  Frog damage       6
  GuardTicks        60
  FrogGuard         90
  BounceVY          -2.5
  SplashVY          -3.5
  FlingVY           -4.0
  BeePushTicks      10
  FrogLeapWait      150
  UpdraftForce      -0.06
  DowndraftForce    0.05

Fireflies
  ChainWindow       120 ticks
  ChainMax          5
  FireflyPoints     10 × chain
  FireflySeeds      15

End of flight
  EndTicks          180
```

Exact values are tuning constants. Tests assert important relationships and
the release values that form product contract: 50 starting puffs, one puff
per flap, one-in-four planting, the hazard damage table and the 15-puff
firefly cost.

## 10. Update loop

The update sequence remains explicit and single-threaded.

``` text
clear Events
    |
Attract? Space starts a fresh flight; return
    |
Landed? count down result interval -> Attract; return
    |
tick counters: chain, guard, cooldown, push
    |
advance scroll and camera
    |
generate world ahead
    |
apply drift; apply flap (spend, plant, take off)
    |
perched? recharge : move bunny
    |
update planters / sprouts
    |
update bees / frogs
    |
resolve flower perch
    |
resolve hazards and fireflies
    |
resolve ground / water
    |
empty reserve -> Falling
    |
derive metres and score
    |
prune world behind camera
```

No gameplay goroutines.

No channels.

No event bus.

No asynchronous actors.

The sequence is intentionally visible because order matters. For example,
perch is resolved before hazards so a clean landing on a flower beside a
thistle is a landing, and the empty-reserve transition happens after hazards
so a firefly or bee that empties the puff ends the flight in the same tick.

## 11. Time and movement

Use Ebitengine's fixed update cycle.

Each flying tick:

``` go
maxFall := MaxFallForSeeds(Seeds)
VY = min(maxFall, VY + Gravity + gustAtBunny())
Y += VY
SX += VX
clamp Y to Ceiling; clamp SX to MinSX..MaxSX
```

A falling bunny uses `MaxFallStem` and ignores gusts.

Scroll speed is a pure function of camera distance:

``` go
Scroll = ScrollStart + (ScrollMax - ScrollStart) * min(1, Camera / ScrollRamp)
```

so the world reaches full speed at 750 m and then holds.

Bees move left at a constant 1 px/tick in world space and bob on a sine of
`Ticks`. Frogs leap on a fixed parabola when the bunny is low and near.

There is no:

-   physics engine;
-   rigid-body simulation;
-   steering or pathfinding;
-   drag model;
-   entity-to-entity collision between world objects;
-   tweening or animation framework.

The combination of reserve-dependent lift, scrolling, gusts and hazard
placement supplies the movement complexity.

## 12. Puff, lift and flap

The puff reserve is an integer from 0 to 50.

Lift and terminal fall are monotonic functions of remaining reserve:

``` text
f       = Seeds / MaxSeeds
lift    = 0.45 + 0.55 * f
maxFall = MaxFallFull + (MaxFallEmpty - MaxFallFull) * (1 - f)
```

A full puff flaps at full strength and falls at 1.0 px/tick. An almost empty
puff flaps at less than half strength and falls at up to 2.4 px/tick.

A valid Space press while flying or perched:

``` text
lift from current reserve
VY = max(MaxRise, VY - FlapImpulse * lift)
Seeds--
Released++
if Released % 4 == 0 and planters < MaxPlanted: create Planter
if perched: leave flower, emit TakeOff
start flap cooldown
emit Flap
```

Lift is calculated before the puff is spent, so the final flap still lifts.

When reserve reaches zero, state becomes `Falling` and further flaps are
ignored.

## 13. Perching and recharge

A bloomed flower with nectar can be perched only during a controlled
downward approach to its head:

``` text
0 <= VY <= PerchMaxVY
distance(bunny feet, flower head) <= FlowerRadius
```

Rising through a flower or dropping onto it too fast does not perch.

While perched:

``` text
carry bunny with flower as camera advances
hold vertical position on the flower head
if Seeds < 50 and Nectar > 0:
    every 15 ticks: Seeds++, Nectar--, emit Regrow
if carried beyond MinSX:
    auto-hop into Flying with a small upward velocity
```

A spent flower can no longer be perched on.

Recharge is intentionally slow — a full wild flower restores 10 puffs over
2.5 seconds — so resting is a decision rather than an instant refill.

## 14. Planting

Every fourth puff spent through flapping creates a gameplay `Planter`, up to
`MaxPlanted` in flight.

A planter drifts ahead and down until it reaches the ground. It becomes a
planted flower only when the location is valid:

-   grass, not water;
-   more than 35 px from a thistle;
-   more than 80 px from another flower;
-   planted-flower cap of 6 not exceeded.

Otherwise the planter is lost.

A planted flower sprouts after `SproutTicks` (four seconds) and carries 6
nectar rather than 10. Its height is derived arithmetically from the release
count, not from the gameplay RNG, so planting cannot perturb world
generation.

The renderer shows cosmetic blown-seed particles for every flap. Cosmetic
particles are not gameplay planters.

## 15. Hazards

Normal hazard damage is applied by one small `knock` rule:

``` text
if Guard > 0 or Seeds <= 0: ignore damage
lost = min(damage, Seeds)
Seeds -= lost
grant guard interval
emit hazard event + SeedsLost(lost)
```

Hazard-specific physical responses remain small:

| Hazard | Damage | Guard | Physical response |
| --- | ---: | ---: | --- |
| Bee | 3 | 60 | 10-tick leftward push |
| Thistle | 3 | 60 | upward bounce |
| Grass | 1 | 60 | upward bounce |
| Pond | 4 | 60 | stronger splash upward |
| Frog | 6 | 90 | stronger fling upward |
| Gust | 0 | — | modifies vertical acceleration |

Bees hit both the body and, while the puff exists, the puff above the head.

Ground and water bounce still occur during protection so the bunny cannot
sink into terrain; only repeat puff loss is suppressed. A guarded bunny
passing over a thistle is not bounced, so protection does not trap it on the
spikes.

A falling bunny is not damaged by hazards. Its flight is already over.

## 16. Fireflies

Fireflies deliberately do not use the normal guarded-hazard rule because
collecting one is an intentional score trade-off.

On contact:

``` text
mark collected
if ChainTicks > 0: Chain = min(5, Chain + 1) else Chain = 1
ChainTicks = 120
FireflyScore += 10 * Chain
FirefliesCollected++
lost = min(15, Seeds)
Seeds -= lost
emit Firefly(chain) + SeedsLost(lost)
```

Guard does not protect against a firefly. A firefly can empty the puff and
end the flight.

## 17. World generation and difficulty

The world is generated in bounded chunks ahead of the camera using the
gameplay RNG. Generation keeps 2,000 px of world ahead of the camera.

Chunk families are:

| Chunk | Width | Contents |
| --- | ---: | --- |
| Meadow | 600 | two wild flowers; one-in-three chance of a firefly |
| Gap | 400 | nothing |
| ThistlePatch | 500 | flower each side; 2–3 thistles in the middle |
| BeeLine | 600 | flower each side; 2–4 bees in alternating lanes |
| Pond | 400–600 | water and one frog |
| Updraft | 400 | rising gust with a five-firefly column |
| Downdraft | 300 | sinking gust |
| FireflyArc | 500 | flower each side; six fireflies on an arc |

Difficulty has one deliberate input: distance. Chunk weights are chosen from
three bands:

``` text
0–250 m     open: mostly meadow, firefly arcs and gaps; no ponds or downdrafts
250–750 m   mixed: all families
750 m+      capped: hazard-heavy weights, then held
```

Scroll speed rises over the same 750 m and then holds. There is no dynamic
difficulty, skill estimation or difficulty profile.

Structural rules prevent impossible worlds:

1.  the first 1,200 px are gentle and contain one early flower;
2.  if more than 100 px of world has no flower, the next chunk is a
    flower-bearing meadow;
3.  ponds never directly follow gaps;
4.  bee lanes leave a vertical gap;
5.  ponds and live-world extent are bounded.

Rule 2 is structural rather than probabilistic, so flower reachability does
not depend on a lucky roll.

Entities more than 200 px behind the camera are pruned every tick, so long
flights do not grow memory without limit.

## 18. Scoring and end of flight

``` text
Metres = floor(Camera / 40)
Score  = Metres + FireflyScore
```

Both are derived every tick rather than accumulated, so they cannot drift.

When reserve reaches zero the bunny enters `Falling`. Hazards no longer
remove reserve. On ground or water contact the bunny becomes `Landed` and a
180-tick result interval starts.

When that interval expires the core returns to `Attract` and emits
`ReturnToAttract`.

The adapter then decides whether the completed score qualifies for initials
entry. This keeps score-table UI out of the gameplay state machine.

No combos beyond the firefly chain, multipliers, currencies or experience.

## 19. Randomness and determinism

Use one seedable gameplay RNG per flight for:

-   chunk family selection;
-   flower heights;
-   thistle count and height;
-   bee count;
-   pond width;
-   meadow firefly placement.

Every other varying value is derived arithmetically from existing state:
bee bob from `Ticks`, planted-flower height from `Released`, firefly arcs
from their index.

Each flight is seeded from `baseSeed + flightNumber`, so successive flights
differ but a run can be reproduced from base seed, flight number, input
sequence and tick count.

The application's base seed comes from the clock at startup. The attract
demo uses the fixed seed 4242 and scripted input.

Cosmetic presentation uses no random numbers at all. Particle spread, bunny
bob and effect bursts derive from frame counters and particle index, so
cosmetic work can never consume or perturb the gameplay stream.

## 20. Rendering boundary

The renderer reads game state and draws it.

It does not decide:

-   whether the bunny perches;
-   whether a hazard or firefly removes puffs;
-   whether a seed plants;
-   whether the flight has ended;
-   what the score is.

The adapter decides only presentation-level questions: whether a completed
score qualifies for the table, and how initials are typed.

Rendering is an adapter over the authoritative game state.

The presentation layer is split by responsibility rather than framework:

``` text
main.go          app loop, input, assets, fonts, initials input
storage.go       high-score file location
scenes.go        attract / play / initials / result composition, HUD, puff
world_render.go  meadow, flowers, planters, hazards, fireflies
effects.go       event reactions, effect bursts, attract demo
audio.go         procedural sound cues
```

The split is larger than Astro Answers' single adapter file because the
scrolling meadow has more drawing responsibilities. It remains one thin
application layer over one gameplay core.

## 21. Visual presentation

The logical playfield is fixed at 960×720.

The Linux V0.1 application launches full-screen by default and preserves the
4:3 logical aspect ratio on wider or taller physical displays. The final
image is scaled uniformly with linear filtering and centred; unused display
area is filled black.

Do not create a bespoke viewport system. Use Ebitengine's normal logical
layout and final-screen hook.

The game uses five embedded production sprites:

-   bunny sheet (four 100×100 frames);
-   bee sheet (two 64×64 frames);
-   frog sheet (two 64×64 frames);
-   thistle;
-   flower.

Sky, hills, clouds, ground, water, stems, planters, fireflies, gusts, the
puff, HUD panels and effects are drawn in code.

Hero sprites remain raster assets. The puff is drawn in code because it must
shrink continuously with the reserve.

### Bunny motion

Bunny frame is chosen from state and vertical velocity. While flying the
renderer adds a small continuous bob and a velocity-linked tilt of at most
±0.14 rad. During guard the sprite blinks.

This is presentation-only and does not alter collision or gameplay
coordinates.

### Readability

Hazards are always drawn above scenery and below the HUD. Effects must not
obscure the bunny, the puff or hazards.

Readability wins over decoration.

## 22. HUD

``` text
[ 0000 m   00000 ]        [ x3 ]        [ (puff)  × 050 ]
```

-   left: metres and score;
-   centre: firefly chain multiplier, shown only while a chain above ×1 is
    live;
-   right: proportional puff graphic plus exact remaining count.

The HUD deliberately uses two views of one value. The graphical puff uses
between 5 and 36 tufts in proportion to the reserve rather than one tuft per
puff. This keeps rendering readable after UAT increased the reserve from the
original prototype value, while the number gives the exact count.

Do not display redundant information.

## 23. Cosmetic particles and effects

Two small cosmetic slices are justified by the visual requirement:

``` text
seed particles   blown seeds from flaps and the final puff release
effect bursts    radial bursts for perch, regrow, hazard, firefly, high score
```

Each tick:

``` text
move
decrement life
remove when expired or off screen
```

Effect-burst particles are capped at 256 in total; additional particles are
dropped.

Seed particles have no separate cap because they are bounded by rule: one
per accepted flap, at most one flap every 8 ticks, each living at most 180
ticks, plus one seven-seed release when the puff empties. That is at most
about 30 live seed particles.

Both slices are cleared when a flight returns to attract and after initials
entry.

Never drop gameplay state to preserve an effect.

There is no general effects engine.

## 24. Sound

Required cues:

-   flap;
-   perch;
-   regrow;
-   bee;
-   thistle;
-   bounce;
-   splash;
-   frog leap;
-   frog catch;
-   puff gone;
-   landing;
-   high score;
-   firefly chain ×1 through ×5.

Cues are generated at startup from a 17-entry table of frequency, duration
and volume. Each becomes a short PCM buffer played through Ebitengine audio
directly.

There is no WAV asset pack, synthesiser subsystem, mixer or music engine.

If a Linux audio device is unavailable, the game continues silently rather
than failing startup.

Music is out of V0.1.

## 25. Attract and initials

The attract screen is a finished game scene containing:

-   title treatment;
-   top five;
-   deterministic demo bunny;
-   start prompt.

The demo uses a real `Game` with a fixed seed and scripted input, so it
exercises real rules rather than a separate animation implementation. When
the demo bunny lands, a fresh demo starts.

Space starts a flight from attract without spending a puff.

Initials entry accepts one to three letters, upper-cased; Backspace deletes;
Enter confirms. An empty entry saves as `?`. Escape quits from any screen.

## 26. High scores

High scores are the only persistent state.

Maintain:

``` text
one top-five table
```

Each entry contains:

``` text
one-to-three-character initials
score
```

Use one small local JSON file at:

``` text
$XDG_DATA_HOME/dandelion-drift/scores.json
```

falling back to `~/.local/share/dandelion-drift/scores.json`. The location
does not depend on the working directory, so scores persist however the
executable is launched.

Persistence remains ordinary functions:

``` go
LoadScores()
SaveScores()
(HighScores).Add()
(HighScores).Qualifies()
```

No repository pattern, adapter or database.

Create the directory if needed and write safely using a temporary file
followed by rename.

If the file is absent, start empty.

If it is malformed, report the problem and start empty.

High-score corruption must not prevent the game from running.

## 27. Assets

Embed required assets in the executable.

Expected asset categories are small:

``` text
bunny
bees
frogs
thistle
flower
font (Go module data)
```

The five PNGs live in `cmd/dandeliondrift/assets/` and are embedded with
`//go:embed assets/*.png`. Sound has no asset files.

Do not create asset quantity as a proxy for quality.

Missing or undecodable required visual assets fail startup visibly, and a
test proves every embedded image decodes.

## 28. Resource budgets

V0.1 has explicit runtime ceilings:

``` text
live world entities     120   (tested over 100 seeds at 3,000 m)
events per tick         8     (tested over long representative runs)
planters in flight      6     (MaxPlanted)
planted flowers         6     (MaxPlanted)
effect-burst particles  256
seed particles          ~30   (bounded by flap cooldown and lifetime)
world ahead of camera   2,000 px
world behind camera     200 px
```

World entities are bounded by generation distance and pruning rather than by
an allocation cap. Additional effect particles are dropped when the cosmetic
ceiling is full; gameplay state is never dropped for an effect.

Sound uses 17 short generated cues and Ebitengine audio directly. V0.1 does
not add a sound pool or player manager. Asset memory is dominated by five
small PNGs and is recorded as build evidence rather than guarded by an
arbitrary runtime limit.

Do not introduce object pooling unless measurement shows garbage collection
materially affects play.

> **Bound gameplay state. Measure the rest. Optimise only when evidence
> says to.**

## 29. Repository shape

The V0.1 implementation remains deliberately small:

``` text
cmd/
    dandeliondrift/
        main.go
        storage.go
        scenes.go
        world_render.go
        effects.go
        audio.go
        assets_test.go
        storage_test.go
        assets/
            5 PNG sprites

internal/
    game/
        game.go
        world.go
        scores.go
        game_test.go
        hazards_test.go
        world_invariants_test.go
        hardening_test.go
        release_balance_test.go
        scores_test.go
```

Assets live beside the only adapter that embeds and renders them. World
generation earned its own file because chunk rules are a separate concern
from per-tick flight rules. Scores earned their own file because persistence
is separate. The score location earned a small adapter file because it is
the one piece of platform policy in the application.

Split files only when the code becomes materially clearer.

Do not create package hierarchies from architectural vocabulary.

No default:

``` text
engine/
domain/
services/
providers/
repositories/
controllers/
managers/
systems/
```

## 30. Failure behaviour

| Failure | Behaviour |
| --- | --- |
| high-score file absent | start empty |
| high-score file malformed | report; start empty |
| high-score directory unwritable | report on save; continue |
| home directory unresolvable | fail startup visibly |
| Linux audio device unavailable | continue silently |
| individual sound cue fails to generate | continue without it |
| required sprite cannot decode | fail startup visibly |
| cosmetic capacity exhausted | drop cosmetic work, preserve gameplay |
| programming/runtime failure | fail visibly |

Failure behaviour remains explicit but proportionate to a local arcade
game.

## 31. Testing

The game rules must run without opening a graphical window.

### Flight

Test:

-   Space on attract starts a flight without spending a puff;
-   a flap rises and spends exactly one puff;
-   lift and glide are monotonic in remaining reserve;
-   the last flap still lifts;
-   ceiling clamp and horizontal drift bounds;
-   gusts change fall.

### Perching

Test:

-   perch occurs only on a gentle descent;
-   recharge continues until full or until nectar is spent;
-   recharge caps at 50;
-   a perched bunny is carried left and auto-hops;
-   a flap takes off from a perch;
-   a spent flower cannot be perched on.

### Planting

Test:

-   one in four released flap seeds plants;
-   planting succeeds on grass only;
-   sprouts bloom after the delay;
-   planted flowers and planters are capped.

### Hazards

Test:

-   each hazard removes its specified puffs;
-   guard prevents a double knock;
-   ground always bounces even when guarded;
-   guard suppresses the thistle bounce;
-   frogs leap only when the bunny is low and overhead;
-   a falling bunny ignores hazard damage.

### Fireflies

Test:

-   the chain rises within the window and resets after it;
-   the chain caps at ×5;
-   a firefly drains 15 puffs;
-   firefly drain cannot take the reserve below zero.

### End of flight

Test:

-   an empty puff falls softly and ends the flight;
-   an idle flight eventually ends;
-   the result interval returns to attract;
-   a new flight can start after a completed one and starts clean;
-   repeated completed flights leave no stale state.

### Scoring

Test:

-   score equals metres with no fireflies;
-   score equals metres plus firefly points.

### World generation

Test:

-   generation is deterministic from seed;
-   the gentle start has one flower;
-   generated flowers stay within reach;
-   ponds never follow gaps;
-   bee lanes leave a vertical gap;
-   invariants hold over long generation;
-   live entities stay within budget over long worlds.

### Scores

Test:

-   descending order, only five retained;
-   qualification;
-   save/load round trip;
-   missing file starts empty;
-   malformed file is reported and starts empty;
-   initials accept one to three letters and store empty entry as `?`;
-   the score path uses `XDG_DATA_HOME`, falls back to the user data
    directory and refuses an empty home.

### Determinism

Given the same:

``` text
seed
input sequence
tick count
```

the gameplay state, including hazards, should reproduce.

### Hardening

Representative long runs keep state sane: no NaN or infinite values, reserve
within bounds, planters and events within budget. A Space-only run advances
without horizontal input.

### Release balance

The release values are pinned by name: 50 starting puffs, flower recharge,
recharge cap, firefly drain and zero flooring.

### Adapter

Every embedded image decodes and every procedural cue produces valid stereo
PCM.

## 32. Build evidence

CI runs:

``` text
go mod tidy contract
gofmt cleanliness
go vet ./...
go test -race -count=1 ./...
linux/amd64 build
linux/arm64 build
named release hardening acceptance tests
```

This is the primary engineering gate.

## 33. Release packaging

The public V0.1 game release follows the Garage pattern: one useful end-user
Linux x86-64 asset attached to the GitHub Release.

The release workflow:

1.  validates `VERSION == 0.1.0`;
2.  reruns source and test gates;
3.  builds the native Linux x86-64 executable;
4.  compresses it with UPX;
5.  graphical-smoke-tests it under Xvfb;
6.  packages `Dandelion-Drift-v0.1.0-linux-x86_64.tar.gz`, preserving
    executable permissions;
7.  uploads that archive as the workflow artefact for attachment to the
    GitHub Release.

Linux ARM64 remains a CI build proof, not a V0.1 release asset.

AppImage is not the V0.1 delivery format because UAT exposed
environment/FUSE friction on the target Ubuntu desktop. The proven native
archive is simpler and more reliable. Because a native executable has no
AppImage launcher to choose a data directory, the executable resolves its own
score location (§26).

## 34. Constrained-hardware benchmark

After the game is correct and polished, run the standard Garage
constrained-hardware benchmark.

Record:

-   hardware;
-   OS;
-   logical and displayed resolution;
-   frame rate;
-   CPU;
-   memory;
-   executable size;
-   maximum observed live world entities;
-   maximum observed effect and seed particles;
-   distance reached during the measured run.

The benchmark proves the design; it does not drive premature optimisation.

## 35. Privacy and network boundary

V0.1:

-   requires no network;
-   sends no telemetry;
-   has no account;
-   has no advertising;
-   has no online leaderboard;
-   has no runtime model or AI dependency;
-   requires no secrets.

Only local initials and scores persist.

## 36. Deliberately absent architecture

EDD v0.3 deliberately does not contain:

``` text
reusable game engine
ECS
event bus or subscribers
scene graph
physics engine
general collision engine
animation or tweening framework
UI framework
Puff, Health or Stamina type
lives counter
separate frog world list
power-up, upgrade or inventory system
currency
difficulty subsystem
dynamic difficulty / skill estimation
score service
score-store abstraction
database
service layer
music system
sound pool or mixer
online leaderboard
accounts
telemetry
network service
runtime AI
gameplay goroutines
channels
object pooling without evidence
large package hierarchy
```

Two earlier design choices were reversed:

``` text
Initials screen in Game  ->  initials in the adapter
24 seeds drawn one by one ->  50-puff reserve with tufts + exact count
```

The first removes a screen from the deterministic core because entering
initials never changes a gameplay rule. The second came from UAT: the
original reserve made flights too short, and drawing each seed separately
stopped being readable once the reserve grew.

One addition has been kept:

``` text
Events []Event
```

It earns its keep because it removes snapshot-diffing from the adapter while
adding no subscribers or delivery machinery (§7).

## 37. Architecture summary

``` text
                  KEYBOARD
                     |
                     v
               +-----------+
               |   Game    |
               +-----------+
                |    |    |
                |    |    +---- camera / scroll / score / chain
                |    |
                |    +--------- Bunny + Seeds (puff)
                |
                +-------------- world slices
                                    |
                     flowers / planters / bees / thistles
                     ponds (+frog) / gusts / fireflies

                  Game.Update(Input)
                         |
      +-------------+----+-------+------------+
      |             |            |            |
    flap          scroll       hazards      fireflies
      |             |            |            |
   spend          generate      knock        chain
      |             |            |            |
   plant          prune        guard        drain
      |                          |
   perch / regrow           ground / water
      |
   empty -> fall -> land -> attract


                  Game.Draw()
                       |
                  Ebitengine
```

Persistent state:

``` text
one top-five score table in the user data directory
```

Content:

``` text
eight chunk families
+
seeded placement and distance-banded weights
```

Cosmetics:

``` text
five embedded sprites + code-drawn meadow and puff
+
generated sound cues
+
two bounded cosmetic slices
```

That is the application.

## 38. Design conclusion

The game should look substantially more sophisticated than its
implementation.

The puff is the central rule:

-   flapping spends it;
-   its level changes flight;
-   flowers restore it;
-   hazards drain it;
-   fireflies trade 15 of it for scoring opportunity;
-   what is spent can grow future recovery points.

Scrolling creates movement pressure.

Chunk families create variety.

Distance bands create rising difficulty.

Planting turns spending into future recovery.

Good artwork, bunny motion, the shrinking puff and soft effects create
visual richness.

A top-five score table creates replay motivation.

The red-team test is therefore not whether the design has the fewest
possible structs or fields.

It is:

> **Does every piece of machinery remove more complexity than it
> introduces?**

For EDD v0.3, the answer leaves us with one authoritative gameplay state,
plain bunny and world data, ordinary functions, one update loop, one local
score file and a thin 2D runtime dependency.

Anything substantially beyond that must earn its place.
