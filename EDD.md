# Dandelion Drift — Engineering Design Document

**Version:** 0.2  
**Status:** Final — V0.1.0, reconciled against the approved implementation  
**Implementation:** Go 1.25+, Ebitengine 2.10.5

## 1. Purpose

Dandelion Drift is one small endless flight game. A bunny floats beneath a dandelion puff that is both movement reserve and health. Flowers restore it. Hazards and fireflies drain it. Released flap seeds can grow into future flowers.

> **Keep the game as small as the game actually is.**

Richness comes from a few interacting rules: flapping spends the reserve, reserve level changes flight, flowers restore it, released seeds alter the future world, and hazards apply pressure.

The implementation is not a reusable game engine.

## 2. Design constraints

1. one authoritative gameplay state;
2. one explicit single-threaded update loop at 60 ticks per second;
3. deterministic gameplay from run seed and input;
4. no abstraction before demonstrated need;
5. bounded gameplay and cosmetic resources;
6. renderer and sound own no gameplay decisions;
7. every gameplay rule is testable without opening a window;
8. high-score entry and cosmetic presentation stay outside gameplay truth;
9. visual quality comes from art, animation and composition rather than frameworks;
10. release packaging must remain simpler than the game.

If the implementation becomes materially more complicated than this document, first ask what can be deleted.

## 3. Runtime dependency

Ebitengine provides:

- window and fixed game loop;
- keyboard input;
- 2D drawing and transforms;
- text;
- audio output.

Go's standard library provides persistence, deterministic RNG and the remaining runtime services.

No second game, UI, physics or audio framework is required.

```text
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
| vertical position | screen pixels, 0 at top |
| velocity | pixels per tick |
| displayed distance | metres, 40 px = 1 m |

Logical playfield: **960×720**.

The world camera advances rightward while scenery is rendered moving left. Bunny horizontal position remains bounded within the playfield.

## 5. Authoritative gameplay state

Gameplay truth lives in one `Game` in `internal/game`.

Conceptually the implementation owns:

```go
type Game struct {
    Screen Screen

    Camera float64
    Scroll float64
    Bunny  Bunny
    Seeds  int

    Flowers   []Flower
    Planters  []Planter
    Bees      []Bee
    Thistles  []Thistle
    Ponds     []Pond
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
    // generation counters / boundaries
}
```

The exact field layout is implementation detail. The design boundary is not: gameplay truth lives here.

The core has only `Attract` and `Play` screens. High-score initials entry is presentation/application state in the Ebitengine adapter because it does not alter game rules.

## 6. Adapter-owned state

`cmd/dandeliondrift` owns things that affect presentation but not gameplay truth:

- Ebitengine images and fonts;
- particle/effect lists;
- procedural audio players;
- persisted high-score table;
- initials-entry text and mode;
- deterministic attract-demo game;
- animation frame counters.

This separation is deliberate. Moving initials, renderer particles or sound machinery into `Game` would make the deterministic core less clear without improving behaviour.

## 7. Bunny

```go
type Bunny struct {
    SX, Y        float64
    VX, VY       float64
    State        BunnyState // Flying, Perched, Falling, Landed
    Guard        int
    PushTicks    int
    Perch        int
    RegrowTimer  int
    FlapCooldown int
}
```

The bunny's position is the centre of the body. Collision geometry is simple circles/rectangles around the visible body and puff.

## 8. Core tuning constants

The final V0.1 release uses:

```text
TPS               60
logical size      960 × 720
MaxSeeds          50
ScrollStart       1.5 px/tick
ScrollMax         3.0 px/tick
FlapCooldown      8 ticks
RegrowTicks       15 ticks per puff
NectarWild        10
NectarPlanted     6
PlantEvery        4 released flap seeds
Bee damage        3
Thistle damage    3
Ground damage     1
Water damage      4
Frog damage       6
Firefly damage    15
ChainWindow       120 ticks
ChainMax          5
FireflyPoints     10 × chain
EndTicks          180
```

Tests should assert important relationships and release values where they form product contract.

## 9. Update model

The core update is explicit and single-threaded.

Conceptually:

```text
clear Events
    |
Attract? Space may start fresh flight; otherwise return
    |
Landed? count down result interval -> Attract; return
    |
advance camera / scroll
    |
generate world ahead
    |
read flap / optional drift
    |
update bunny flight or perch recharge
    |
update planters / sprouts
    |
update bees / frogs
    |
resolve flower perch
    |
resolve hazards
    |
resolve fireflies
    |
resolve ground / water
    |
update score / chain / distance
    |
prune world behind camera
    |
empty reserve -> Falling
```

No goroutines, channels, ECS, actors or event bus.

## 10. Puff, lift and flap

The puff reserve is an integer from 0 to 50.

Lift and terminal fall are monotonic functions of remaining reserve:

```text
f       = Seeds / MaxSeeds
lift    = 0.45 + 0.55 * f
maxFall = MaxFallFull + (MaxFallEmpty - MaxFallFull) * (1 - f)
```

A valid Space press while flying or perched:

```text
calculate lift from current reserve
apply upward impulse
Seeds--
Released++
if Released % 4 == 0: create Planter
if perched: leave flower
start flap cooldown
emit Flap / TakeOff events as appropriate
```

Lift is calculated before the puff is spent, so the final flap still lifts.

When reserve reaches zero, state becomes `Falling` and further flaps are ignored.

## 11. Perching and recharge

A bloomed flower can be perched only during a controlled downward approach to its head.

While perched:

```text
carry bunny with flower as camera advances
hold vertical position
if Seeds < 50 and Nectar > 0:
    every 15 ticks: Seeds++, Nectar--, emit Regrow
if carried beyond left bound:
    auto-hop into Flying
```

A spent flower can no longer provide recharge.

Recharge is intentionally slow enough that resting is a decision rather than an instant refill.

## 12. Planting

Every fourth puff spent through flapping creates a gameplay `Planter`.

A planter moves ahead/downward until it reaches the ground. It becomes a planted flower only when the location is valid:

- grass, not water;
- not inside/too near a thistle;
- not too close to another flower;
- planted-flower cap not exceeded.

Planted flowers sprout after a fixed delay and carry less nectar than wild flowers.

The renderer may show cosmetic blown-seed particles for every flap; cosmetic particles are not gameplay planters.

## 13. Hazards

Normal hazard damage is applied by one small `knock` rule:

```text
if Guard > 0 or Seeds <= 0: ignore damage
lost = min(damage, Seeds)
Seeds -= lost
grant guard interval
emit hazard event + SeedsLost(lost)
```

Hazard-specific physical responses remain small:

| Hazard | Damage | Physical response |
| --- | ---: | --- |
| Bee | 3 | short leftward push |
| Thistle | 3 | upward bounce |
| Grass | 1 | upward bounce |
| Pond | 4 | stronger splash upward |
| Frog | 6 | stronger fling upward |
| Gust | 0 | modifies vertical acceleration |

Ground and water bounce still occur during protection; only repeat puff loss is suppressed.

## 14. Fireflies

Fireflies deliberately do not use the normal guarded-hazard rule because collecting one is an intentional score tradeoff.

On contact:

```text
mark collected
advance/reset chain
ChainTicks = 120
FireflyScore += 10 * Chain
FirefliesCollected++
lost = min(15, Seeds)
Seeds -= lost
emit Firefly(chain) + SeedsLost(lost)
```

This means a firefly can empty the puff and end the flight.

## 15. World generation

The world is generated in bounded chunks ahead of the camera using the gameplay RNG.

Chunk families are:

- Meadow;
- Gap;
- ThistlePatch;
- BeeLine;
- Pond;
- Updraft;
- Downdraft;
- FireflyArc.

Early play strongly favours open chunks. Hazard pressure increases with distance and then caps.

Structural rules prevent impossible worlds:

1. the first 1,200 px are gentle and contain an early flower;
2. flower-free chunks are followed by a flower-bearing meadow;
3. ponds do not directly follow gaps;
4. hazard patterns leave usable paths;
5. ponds and live-world extent are bounded.

Generated entities behind the camera are pruned, so long flights do not grow memory without limit.

## 16. Scoring and end of flight

```text
Metres = floor(Camera / 40)
Score  = Metres + FireflyScore
```

When reserve reaches zero, the bunny enters `Falling`. Hazards no longer remove reserve. On ground/water contact the bunny becomes `Landed` and a 180-tick result interval starts.

When that interval expires, the core returns to `Attract`.

The adapter then decides whether the completed score qualifies for initials entry. This keeps score-table UI out of the gameplay state machine.

## 17. Persistence

The only persisted state is one top-five score table in local JSON.

`internal/game/scores.go` provides normalisation, qualification, load and atomic temp-file/rename save behaviour.

Missing score file starts empty. Malformed data is reported and the application continues with an empty table.

The adapter owns the path and initials interaction.

## 18. Attract and initials

The attract screen is a finished game scene containing:

- title treatment;
- top five;
- deterministic demo bunny;
- start prompt.

The demo uses a real `Game` with a fixed seed and scripted input, so it exercises real rules rather than a separate animation implementation.

Initials entry is one-to-three letters; empty entry saves as `?`.

## 19. Rendering

Rendering reads state and decides no gameplay outcomes.

The presentation layer is split by responsibility rather than framework:

```text
main.go          app loop, input, assets, fonts
scenes.go        attract/play/initials/result composition
world_render.go  meadow, flowers, hazards, world pressure
bunny_asset.go   embedded bunny sheet
 effects.go      cosmetic effects
audio.go         short procedural sound cues
```

The split is intentionally larger than Astro Answers' adapter because Dandelion Drift has a richer scrolling world, but it remains one thin application layer.

### Puff display

The HUD deliberately uses two views of one value:

- a proportional graphical puff;
- an exact numeric remaining count.

The graphical puff uses a bounded number of tufts rather than drawing one tuft per puff. This keeps rendering readable after UAT increased the reserve from the original prototype value.

### Bunny motion

The renderer adds small continuous bob and rotation to the sprite while flying. This is presentation-only and does not alter collision or gameplay state.

## 20. Assets

Final V0.1 visual assets are embedded with the executable:

- bunny sprite sheet embedded as a Go base64 constant;
- bee sheet;
- frog sheet;
- thistle image;
- flower image.

The non-bunny image files are embedded from `assets/*.b64` and decoded at startup.

Obsolete prototype bunny/audio bundle files are not part of the release source.

## 21. Sound

Short cues are generated procedurally in `audio.go` from small cue specifications.

There is no WAV asset pack, synthesiser subsystem or music engine.

If a Linux audio device is unavailable, the game continues silently rather than failing startup.

## 22. Testing

Gameplay tests run headless under `internal/game`.

The release suite covers:

- flap/lift/glide relationships;
- perch rules and recharge;
- planting and sprout behaviour;
- each hazard and protection behaviour;
- firefly chain/scoring/damage;
- end-of-flight routing;
- deterministic generation;
- world invariants and live-entity bounds;
- repeated-run reset hygiene;
- high-score persistence and malformed-file handling;
- final release balance: 50 starting puffs, flower recharge, 15-puff firefly cost.

`cmd/dandeliondrift/assets_test.go` also proves required embedded images decode and all procedural sound cues produce valid PCM.

## 23. CI

Normal CI enforces:

```text
go mod tidy contract
gofmt cleanliness
go vet ./...
go test -race -count=1 ./...
Linux amd64 build
Linux arm64 build
named release hardening acceptance tests
```

This is the primary engineering gate.

## 24. Release packaging

The public V0.1 game release follows the Astro Answers Garage pattern: one useful end-user Linux x86-64 asset attached to the GitHub Release.

The release workflow:

1. validates `VERSION == 0.1.0`;
2. reruns source/test gates;
3. builds the native Linux x86-64 executable;
4. compresses it with UPX;
5. graphical-smoke-tests it under Xvfb;
6. packages `Dandelion-Drift-v0.1.0-linux-x86_64.tar.gz` preserving executable permissions;
7. uploads that archive as the workflow artefact for manual attachment to the GitHub Release.

Linux ARM64 remains a CI build proof, not a V0.1 release asset.

AppImage is not the V0.1 delivery format because UAT exposed environment/FUSE friction on the target Ubuntu desktop. The proven native archive is simpler and more reliable.

## 25. Repository shape

```text
README.md
PRD.md
EDD.md
DESIGN-NOTES.md
TESTING.md
RELEASE.md
CHANGELOG.md
LICENSE
VERSION
architecture.txt
generated/
    EDD.md
    COMPARISON.md
cmd/
    dandeliondrift/
        main.go
        scenes.go
        world_render.go
        effects.go
        audio.go
        bunny_asset.go
        assets/
internal/
    game/
        game.go
        world.go
        scores.go
        *_test.go
.github/workflows/
    test.yml
    release-build.yml
```

## 26. Failure behaviour

| Failure | Behaviour |
| --- | --- |
| score file absent | start empty |
| score file malformed | report and continue empty |
| Linux audio device unavailable | continue silently |
| required sprite cannot decode | fail startup visibly |
| live cosmetic capacity exhausted | drop cosmetic work, preserve gameplay |
| programming error | fail visibly |

## 27. Deliberately absent

```text
physics engine          reusable game engine
ECS                     event bus
scene graph             service layer
repository abstraction  database
music system            online leaderboard
accounts                telemetry
network service         runtime AI
upgrade system          currency / inventory
```

## 28. Design conclusion

Dandelion Drift remains a small game despite a finished presentation.

The puff is the central rule:

- flapping spends it;
- its level changes flight;
- flowers restore it;
- hazards drain it;
- fireflies trade 15 of it for scoring opportunity;
- what is spent can grow future recovery points.

The final red-team question is unchanged:

> **Does every piece of machinery remove more complexity than it introduces?**
