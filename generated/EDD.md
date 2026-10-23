# Dandelion Drift — Generated Engineering Design Document

**Evidence commit:** `74ec654e7ac7c92ff88befe3f2fd800431648173`  
**Method:** code-derived review using the Jonno's Garage `codebase-to-EDD` evidence discipline  
**Generated:** 2026-10-23

This document describes the approved gameplay implementation at the evidence commit. The hand-written `PRD.md` and `EDD.md` are comparison targets, not runtime evidence.

## 1. Executive description

**KNOWN** — Dandelion Drift is a small Linux desktop flight game written in Go. Deterministic gameplay rules live in `internal/game`; `cmd/dandeliondrift` is a thin Ebitengine adapter for input, rendering, embedded artwork, procedural audio, initials entry and local score persistence.

**KNOWN** — one `Game` owns gameplay truth. The adapter owns presentation-only state.

## 2. Evidence inspected

- `internal/game/game.go` — authoritative gameplay state, update loop, puff economy, perching, planting, hazards, fireflies, scoring and run lifecycle;
- `internal/game/world.go` — deterministic chunk generation and pruning;
- `internal/game/scores.go` — local top-five score persistence;
- `internal/game/*_test.go` — rules, hardening, determinism, world invariants and final balance evidence;
- `cmd/dandeliondrift/main.go` — Ebitengine adapter, input, assets, fonts, score/initials integration;
- `cmd/dandeliondrift/scenes.go` — attract/play/result/initials presentation;
- `cmd/dandeliondrift/world_render.go` and `effects.go` — world and cosmetic rendering;
- `cmd/dandeliondrift/audio.go` — procedural sound cues and silent fallback;
- `cmd/dandeliondrift/assets_test.go` — asset decode and procedural audio validation;
- `.github/workflows/test.yml` — engineering gate;
- `VERSION` — release boundary.

## 3. Runtime structure

```text
keyboard
   |
   v
cmd/dandeliondrift
   |
   v
Game.Update(Input)
   |
   +-- flap / drift
   +-- camera / generation / pruning
   +-- perch / recharge
   +-- planter / sprout
   +-- hazards / fireflies
   +-- score / chain / end-of-flight
   |
   v
Game state
   |
   +--> renderer + embedded sprites
   +--> procedural sound
   +--> adapter-owned high scores / initials
```

There is no game engine abstraction, ECS, event bus, service layer, repository abstraction, physics framework, network path or runtime model/AI dependency.

## 4. Authoritative state

**KNOWN** — `Game` owns camera/scroll, bunny state, puff reserve, generated world entities, released-seed count, firefly chain/score, run score/distance/timing, events, deterministic RNG and generation cursors.

**KNOWN** — core screens are `Attract` and `Play`. `Flying`, `Perched`, `Falling` and `Landed` are bunny states.

**KNOWN** — initials entry and persistent score-table state are intentionally outside the core `Game` in the adapter.

## 5. Puff economy

**KNOWN** — `MaxSeeds = 50`.

**KNOWN** — a fresh flight starts full. Each valid flap spends one reserve unit. Lift is calculated before the spend, so the final flap still applies lift.

**KNOWN** — `LiftForSeeds` and `MaxFallForSeeds` vary monotonically with remaining reserve.

**KNOWN** — reserve at zero transitions the bunny to `Falling`.

## 6. Flowers

**KNOWN** — a valid gentle descent onto a bloomed flower with nectar enters `Perched` state.

**KNOWN** — while perched, one puff is restored every `RegrowTicks = 15` until either the reserve reaches 50 or flower nectar reaches zero.

**KNOWN** — wild flowers start with 10 nectar; planted flowers with 6.

**KNOWN** — the bunny is carried left with the world while perched and auto-hops when the left bound is crossed.

## 7. Planting

**KNOWN** — `PlantEvery = 4`. Every fourth puff spent by flapping creates a gameplay planter.

**KNOWN** — planters that reach valid grass may become planted flowers after `SproutTicks = 240`.

**KNOWN** — planting is blocked by ponds, nearby thistles, nearby flowers and the planted-flower cap.

## 8. Hazards

**KNOWN** — normal hazard damage is centralised through `knock`, which clamps loss to current reserve and applies a guard period.

Final fixed damage values are:

```text
bee       3
thistle   3
grass      1
water      4
frog       6
```

**KNOWN** — gusts alter vertical acceleration but do not directly remove reserve.

## 9. Fireflies

**KNOWN** — fireflies are not a recharge path.

**KNOWN** — contact marks the firefly collected, advances the chain, scores `10 × chain`, and removes `min(15, Seeds)` puffs.

**KNOWN** — a firefly can therefore reduce reserve to zero and end a flight.

## 10. World generation

**KNOWN** — generated chunk families are meadow, gap, thistle, bee line, pond, updraft, downdraft and firefly arc.

**KNOWN** — world generation uses the gameplay RNG, advances ahead of the camera and prunes entities behind it.

**KNOWN** — structural generation rules force an early flower, follow flower-free space with flower-bearing meadow, prevent pond-after-gap combinations and bound generated content.

## 11. Scoring and run lifecycle

**KNOWN** — displayed metres are derived from camera distance at 40 px per metre.

**KNOWN** — total score is metres plus accumulated firefly score.

**KNOWN** — at zero reserve the bunny falls. Once the falling bunny reaches ground/water it becomes landed and a finite end interval starts.

**KNOWN** — when that interval expires the core returns to attract. The adapter then routes qualifying scores to initials entry.

## 12. Persistence

**KNOWN** — only one local top-five table persists.

**KNOWN** — score writes use temp-file then rename. Missing file starts empty. Malformed data is reported and normalised/empty behaviour is used rather than crashing the application.

## 13. Adapter and rendering

**KNOWN** — the presentation layer is split into small responsibility files rather than a framework.

**KNOWN** — bunny sprite motion adds render-only bob and rotation derived from frame/velocity. It does not alter gameplay position.

**KNOWN** — the HUD includes the exact remaining puff count and a proportional graphical puff.

**KNOWN** — attract mode uses a second real deterministic `Game` with scripted input for the demo.

## 14. Assets and sound

**KNOWN** — final visual identity uses one embedded bunny sheet plus embedded bee, frog, thistle and flower image data.

**KNOWN** — sound cues are generated in code from small frequency/duration/volume specifications.

**KNOWN** — Linux audio unavailability produces silent fallback rather than startup failure.

## 15. Tests and resource bounds

**KNOWN** — gameplay tests are headless.

**KNOWN** — hardening tests include representative long runs, repeated fresh flights and bounded live-world checks.

**KNOWN** — final balance tests explicitly prove 50 starting puffs, flower recharge, firefly drain and no negative reserve.

**KNOWN** — CI enforces module tidiness, formatting, vet, race tests, amd64/arm64 builds and named hardening acceptance tests.

## 16. Material development corrections

Human UAT changed the resource balance from the prototype 24-unit puff to the final 50-unit reserve and added an explicit HUD count.

Fireflies moved through several trial roles before the final rule was locked: they remain score/chain objects but cost 15 puffs on contact.

Sprite fluidity was improved in the renderer without changing deterministic gameplay motion.

Audio moved from embedded bundle expectations to generated cues with silent fallback.

Packaging moved away from AppImage after target-Ubuntu/FUSE friction. The final delivery path is a native Linux x86-64 executable preserved inside a tar.gz.

## 17. Open evidence

The standard Garage constrained-hardware benchmark is not encoded in this source snapshot and remains a separate physical evidence item.

## Result

The approved V0.1 gameplay source is a lean deterministic mini-monolith: one core game state, one thin Ebitengine application layer and no unexplained runtime architecture.
