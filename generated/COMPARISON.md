# Dandelion Drift — Design / Implementation Comparison

**Implementation evidence:** `6c70eea985b58d6fe3511b26ba5b7c7d2d902dfa`  
**Generated:** 2026-10-23

| Design area | Final contract | Recovered implementation | Assessment |
|---|---|---|---|
| authoritative state | one gameplay `Game` | one `Game` owns gameplay truth | MATCH |
| presentation state | adapter owns initials/effects/audio | adapter owns these concerns | MATCH |
| starting reserve | 50 | `MaxSeeds = 50` | MATCH |
| flap cost | 1 | one reserve unit per valid flap | MATCH |
| flowers | recharge | timed recharge from finite nectar | MATCH |
| fireflies | score + drain 15 | score/chain plus `FireflySeeds = 15` loss | MATCH |
| other hazards | fixed small losses | central guarded `knock` rule | MATCH |
| planting | every fourth flap seed | `PlantEvery = 4` | MATCH |
| puff display | graphic + exact count | both rendered in HUD | MATCH |
| end state | soft fall, result, restart | Falling → Landed → Attract | MATCH |
| scores | local top five | JSON top-five persistence | MATCH |
| deterministic core | fixed-seed gameplay | seeded RNG, headless tests | MATCH |
| renderer | thin Ebitengine layer | renderer owns no gameplay decisions | MATCH |
| sound | small procedural cues | generated PCM cue table | MATCH |
| full-screen display | fixed 960×720, aspect preserved | logical layout + black final-screen fill | MATCH |
| update order | explicit single-threaded loop; perch before hazards; empty check after hazards | `Update` follows that order | MATCH |
| perching | gentle descent only; recharge every 15 ticks from finite nectar | `resolvePerch` velocity window + radius; `updatePerched` | MATCH |
| hazards | bee 3, thistle 3, grass 1, pond 4, frog 6; guard; ground still bounces | constants + `knock`; guarded ground bounce | MATCH |
| world generation | eight chunk families; three distance bands; structural flower rule | `chooseChunkKind` bands at 250 m and 750 m; forced meadow after flower-free gap | MATCH |
| RNG | gameplay RNG for generation only; cosmetics derived | `rng` used only in `world.go` and seeding; adapter has no RNG | MATCH |
| resource bounds | 120 world entities, 8 events, 6 planters, 256 effect particles | asserted in hardening/invariant tests and `effects.go` | MATCH |
| score location | user data directory, independent of launch path | `storage.go` resolves `XDG_DATA_HOME` or `~/.local/share` | MATCH after correction |
| hero assets | raster sprites as files | five embedded PNGs | MATCH after correction |
| CI | format/vet/race/amd64/arm64/hardening | encoded in `test.yml` | MATCH |
| release | simple native x86-64 archive | native executable smoke-tested and tar packaged | MATCH after publication cleanup |
| network/privacy | none | no network/account/telemetry/runtime AI path | MATCH |

## Material corrections made during development

The prototype PRD/EDD described a 24-unit puff rendered seed-for-seed. UAT showed that the run duration and visibility were wrong. The final product starts at 50 and combines a proportional puff graphic with an exact numeric counter.

The resource loop was deliberately tested through several variants. Final UAT locked flowers as recharge and fireflies as a 15-puff score risk. The architecture did not expand to support these changes.

The first bunny presentation looked mechanically stepped. The final renderer adds small continuous bob/rotation without changing gameplay coordinates or collision state.

The design draft expected embedded WAV assets. Runtime hardening produced a smaller final design: procedural cues generated directly in Go with silent fallback if Linux audio is unavailable.

Early packaging work used AppImage and raw native binaries. Target-Ubuntu testing exposed FUSE/environment friction and executable-bit delivery issues. The final release path is a native x86-64 executable inside tar.gz, preserving mode 755.

Two obsolete encoded UAT artefacts remained after the approved gameplay snapshot (`assets/bunny.b64` and `assets/sounds.zip.b64`). They were not used by the final runtime and were removed during publication cleanup. No gameplay behaviour changed.

A final benchmark against Astro Answers found that the native executable wrote `scores.json` to whatever directory it was launched from; Astro avoids this through its AppImage launcher. The executable now resolves the per-user data directory. The same pass converted the remaining base64 sprite data to plain PNG files. Neither change affects gameplay.

## Deliberate differences from Astro Answers

Dandelion Drift has more adapter source files because the scrolling meadow requires separate scene/world/effect rendering responsibilities. This is not additional application architecture: all files remain inside one Ebitengine adapter and share one gameplay core.

Dandelion Drift uses a native tar.gz rather than AppImage because the latter was less reliable on the tested Ubuntu target. The Garage rule is reliable simplicity, not identical packaging for its own sake. The cost is one small score-location function that AppImage's launcher would otherwise have provided.

## Open evidence

The standard Garage constrained-hardware benchmark remains a separate physical evidence item and is not claimed here.

## Result

**MATCH after correction.** The final PRD and EDD now describe the approved V0.1 implementation. Material UAT divergences are incorporated explicitly, obsolete build artefacts are removed, and no unexplained architecture mismatch remains.
