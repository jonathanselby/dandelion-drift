# Dandelion Drift

A gentle side-scrolling flight game for children, built in Go with Ebitengine.

A bunny floats beneath a dandelion puff. Every flap releases a seed, so the same visible resource is lift, health and — as the game grows — the source of future flowers.

> Every flap spends your wings. Every flap plants tomorrow.

## Build status

**V0.1 — Chunk 4 complete.**

The fourth coherent build retains the complete game and adds the production presentation layer:

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
- conservative generated flower spacing proved against the full-puff glide limit;
- bees with deterministic bobbing, three-seed knocks and brief pushback;
- thistles, ponds/water knocks and low-overhead frog leaps;
- updraft and downdraft gust zones;
- one-second knock protection, with grass/water bounce preserved while guarded;
- collectible fireflies with a two-second chain window and multiplier capped at ×5;
- total score = metres + firefly points;
- richer deterministic chunk catalogue whose pressure increases with distance;
- long-run generation tests for gentle opening, reachable flowers, clear approaches, passable bee lanes, bounded ponds and live-entity budget;
- the final bee, thistle, flower and frog art integrated at gameplay resolution;
- one persistent local top-five table with atomic save/replace;
- one-to-three-letter initials entry, with `?` for an empty entry;
- a live fixed-seed demo bunny on the attract screen using the real game rules;
- top five displayed as part of the finished attract scene;
- event-driven sound cues, including five rising firefly-chain notes;
- perch petals, regrow sparkles, splash/firefly bursts and chain pop feedback;
- soft lily-pad landing when the final fall ends over water;
- polished high-score initials scene and clean return to attract.

The game feature set is now present. The remaining chunks are hardening/UAT and release/publication work; they do not add new gameplay systems.

## Controls

- `Space` — start / flap
- `Left`, `Right` — optional drift
- `Esc` — quit

## Engineering shape

Gameplay truth lives in `internal/game`. Ebitengine in `cmd/dandeliondrift` is a thin input/rendering adapter. Rendering owns no gameplay decisions.
