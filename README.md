# Dandelion Drift

**Every flap spends your wings. Every flap plants tomorrow.**

Dandelion Drift is a gentle full-screen side-scrolling flight game for children, written in Go. A bunny floats beneath a dandelion puff. Each flap spends one puff; flowers restore it; hazards and fireflies drain it; released seeds can grow into future flowers.

The repository is also a Jonno's Garage engineering example: polished output from a deliberately small runtime design.

## Play

```text
Space        start / flap
Left/Right   optional drift
Esc          quit
```

A flight begins with **50 puffs**. Each flap spends one. A fuller puff gives stronger lift and a slower glide; a depleted puff weakens both.

Land gently on a flower to restore puffs from its limited nectar. Every fourth released seed can grow into a future flower. Bees, thistles, grass, ponds and frogs remove puffs. Fireflies are a deliberate risk/reward object: they add score and chain points but drain **15 puffs** when touched.

The HUD shows the remaining puff count explicitly while the drawn puff also shrinks proportionally.

## V0.1

V0.1 targets Linux desktop. It launches full-screen using a fixed 960×720 logical playfield and preserves 4:3 geometry on wider or taller displays with centred black pillarboxing or letterboxing.

The game includes:

- endless deterministic meadow generation;
- 50-puff starting health/lift reserve;
- Space-only complete play with optional horizontal drift;
- flower perching and limited recharge;
- one-in-four seed planting;
- bees, thistles, grass, ponds, frogs and wind gusts;
- firefly scoring chains with a 15-puff collection cost;
- persistent local top-five scores with one-to-three-character initials;
- an attract screen with a live deterministic demo;
- finished sprites, procedural sound cues and restrained visual effects;
- no network, telemetry, accounts, advertising or runtime AI.

## Build

Dandelion Drift uses Go 1.25+ and Ebitengine 2.10.5.

```bash
go build -trimpath -ldflags="-s -w" -o ./bin/dandelion-drift ./cmd/dandeliondrift
./bin/dandelion-drift
```

The executable embeds five PNG sprites and generates its short sound cues in code. High scores are stored locally in `$XDG_DATA_HOME/dandelion-drift/scores.json`, falling back to `~/.local/share/dandelion-drift/scores.json`, so they persist however the game is launched.

## Engineering gate

```bash
gofmt -w ./cmd ./internal
go vet ./...
go test -race -count=1 ./...
go build ./cmd/dandeliondrift
GOOS=linux GOARCH=arm64 go build ./cmd/dandeliondrift
```

The gameplay package is deterministic for a fixed seed and can be tested without opening a graphical window. CI enforces formatting, vet, race tests, Linux builds, long-run hardening and release-balance acceptance tests.

## Architecture

```text
keyboard
   |
   v
Game.Update()  <---- one authoritative gameplay state
   |
   +-- flap / drift
   +-- scroll / generate / prune
   +-- perch / regrow / plant
   +-- hazards / fireflies
   +-- score / end-of-flight
   |
   v
Game state ----> thin Ebitengine adapter ----> screen + procedural sound

user data dir JSON <---- top-five scores only
```

There is no reusable game engine, event bus, service layer, repository abstraction, physics framework or UI framework. Presentation state such as initials entry and particles stays in the Ebitengine adapter rather than contaminating gameplay truth. See [`EDD.md`](EDD.md) and [`architecture.txt`](architecture.txt).

## Design and evidence

- [`PRD.md`](PRD.md) — final product contract and acceptance criteria.
- [`EDD.md`](EDD.md) — final hand-written engineering design.
- [`DESIGN-NOTES.md`](DESIGN-NOTES.md) — material design decisions and rejected complexity.
- [`TESTING.md`](TESTING.md) — automated testing, human UAT and failures found during packaging/runtime testing.
- [`RELEASE.md`](RELEASE.md) — release gate and evidence status.
- [`generated/EDD.md`](generated/EDD.md) — code-derived engineering view pinned to the final source commit.
- [`generated/COMPARISON.md`](generated/COMPARISON.md) — design versus implementation comparison.

## Licence

MIT. See [`LICENSE`](LICENSE).
