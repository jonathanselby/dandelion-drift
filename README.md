# Dandelion Drift

A gentle side-scrolling flight game for children, built in Go with Ebitengine.

A bunny floats beneath a dandelion puff. Every flap releases a seed, so the same visible resource is lift, health and — as the game grows — the source of future flowers.

> Every flap spends your wings. Every flap plants tomorrow.

## Build status

**V0.1 — Chunk 1 complete.**

The first playable vertical slice establishes the complete game lifecycle:

- polished attract/start scene;
- full-screen 960×720 logical playfield with aspect preservation;
- Space to start and flap, Left/Right optional drift;
- 24-seed puff drawn directly from authoritative seed state;
- lift and glide weaken as seeds are lost;
- grass contact bounces and removes a seed with knock protection;
- empty puff transitions to a soft fall and landing;
- visually finished end-of-flight scene;
- automatic return to attract and clean fresh restart;
- headless tests for the lifecycle and flight rules.

Flowers, planting, hazards, scoring, persistent high scores, sound and final effects are deliberately later chunks. They are not represented by temporary systems here.

## Controls

- `Space` — start / flap
- `Left`, `Right` — optional drift
- `Esc` — quit

## Engineering shape

Gameplay truth lives in `internal/game`. Ebitengine in `cmd/dandeliondrift` is a thin input/rendering adapter. Rendering owns no gameplay decisions.
