# Dandelion Drift

A gentle side-scrolling flight game for children, built in Go with Ebitengine.

A bunny floats beneath a dandelion puff. Every flap releases a seed, so the same visible resource is lift, health and — as the game grows — the source of future flowers.

> Every flap spends your wings. Every flap plants tomorrow.

## Build status

**V0.1 — Chunk 2 complete.**

The second coherent build retains the complete lifecycle and adds the core Dandelion Drift resource loop:

- polished attract/start scene;
- full-screen 960×720 logical playfield with aspect preservation;
- Space to start and flap, Left/Right optional drift;
- 24-seed puff drawn directly from authoritative seed state;
- lift and glide weaken as seeds are lost;
- grass contact bounces and removes a seed with knock protection;
- empty puff transitions to a soft fall and landing;
- visually finished end-of-flight scene;
- automatic return to attract and clean fresh restart;
- headless tests for the lifecycle and flight rules;
- deterministic meadow/gap world generation with a gentle opening flower;
- wild flowers with gentle-descent perching;
- seed-by-seed regrowth limited by flower nectar;
- flap take-off and automatic hop when a perch reaches the left edge;
- one gameplay planter for every four released seeds;
- grass-only planting, sprouts, delayed bloom and a six-flower planted cap;
- conservative generated flower spacing proved against the full-puff glide limit.

Hazards, fireflies/scoring, persistent high scores, sound and final presentation effects remain later chunks. They are not represented by temporary systems here.

## Controls

- `Space` — start / flap
- `Left`, `Right` — optional drift
- `Esc` — quit

## Engineering shape

Gameplay truth lives in `internal/game`. Ebitengine in `cmd/dandeliondrift` is a thin input/rendering adapter. Rendering owns no gameplay decisions.
