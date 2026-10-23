# Dandelion Drift

A gentle side-scrolling flight game for children, built in Go with Ebitengine.

A bunny floats beneath a dandelion puff. The puff is the game's single visible resource: it provides lift, takes damage and can be restored by landing on flowers.

**Current release: v0.1.0**

## How it plays

- Start each flight with **50 puffs**.
- `Space` flaps upward and spends one puff.
- Lift and glide weaken as the puff runs down.
- Flowers can be landed on during a gentle descent and **restore puffs**, limited by each flower's nectar.
- Fireflies add score and chain points but **drain 15 puffs** when touched.
- Bees, thistles, water, frogs and ground contact can also drain puffs or disrupt flight.
- Every fourth released seed can become a planted flower.
- When the puff reaches zero, the bunny falls softly and the flight ends on landing.

The game includes deterministic world generation, persistent local top-five scores, initials entry, finished sprites, sound cues, visual effects and a full-screen 960×720 logical playfield.

## Controls

- `Space` — start / flap
- `Left`, `Right` — optional drift
- `Esc` — quit

## Download

Linux release packages are published under **Releases**:

- `Dandelion-Drift-v0.1.0-linux-x86_64.tar.gz`
- `Dandelion-Drift-v0.1.0-linux-arm64.tar.gz`
- `SHA256SUMS.txt`

Extract the archive for your architecture and run `Dandelion-Drift`. No installer is required.

## Engineering shape

Gameplay truth lives in `internal/game`. Ebitengine in `cmd/dandeliondrift` is a thin input/rendering adapter; rendering owns no gameplay decisions.

The build is deliberately small: one Go application, deterministic 60 TPS gameplay, bounded live entities and no ECS, event bus or background runtime architecture.

The design record is retained in [`PRD.md`](PRD.md) and [`EDD.md`](EDD.md). Those documents record the design process; this README describes the shipped v0.1.0 behaviour.
