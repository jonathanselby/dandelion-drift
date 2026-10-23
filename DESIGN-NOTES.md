# Dandelion Drift — Design Notes

Dandelion Drift was designed by reducing the game until movement, health and recovery all depended on the same visible thing.

> **Every flap spends your wings. Every flap plants tomorrow.**

The useful design history is therefore mostly a record of what was simplified, what UAT changed, and what was deliberately not added.

## One resource, not several

The core product idea is the dandelion puff.

A separate stamina bar, lives counter, boost meter and health bar were all unnecessary. One reserve is enough:

- flapping spends it;
- its level changes flight behaviour;
- hazards remove it;
- flowers restore it;
- spent flap seeds can create future recovery points.

The original prototype expressed the reserve as 24 individual seeds. Human UAT showed that the run was too short and that the exact remaining amount was useful. The final release therefore starts at 50, displays the number explicitly, and keeps the puff as a proportional visual rather than pretending every unit must be rendered as a separate hair.

## UAT changed balance, not architecture

The UAT loop deliberately changed the resource economy several times while preserving the same architecture.

The final accepted rule is:

- start at 50;
- each flap spends one;
- flowers recharge;
- fireflies drain 15 while still scoring;
- other animals/hazards drain smaller fixed amounts.

The important engineering result is that these changes required tuning constants and small rule changes rather than new systems.

## Fireflies became a real decision

Early versions treated fireflies as pure score/recharge objects. That made survival too easy.

The final version makes a firefly a deliberate trade:

- score and chain opportunity;
- immediate cost of 15 puffs;
- possible run-ending contact if reserve is low.

No separate power-up type, inventory or boost system was introduced.

## Flowers remain recovery and future consequence

Flowers provide the recovery loop.

Perching restores puffs slowly from finite nectar. Every fourth puff spent through flapping can create a future planter. That preserves the original idea that movement changes the world ahead without adding a crafting or resource-management layer.

## Presentation was allowed to become richer than the core

The final game has a more elaborate presentation layer than Astro Answers because the meadow world needs more drawing responsibilities. That does not justify more gameplay architecture.

The Ebitengine adapter therefore owns:

- sprites;
- procedural sounds;
- attract/demo presentation;
- initials-entry UI;
- particles/effects;
- continuous bunny bob/rotation.

The deterministic `Game` remains unaware of those things.

## Fluidity without physics machinery

UAT identified that the bunny looked too stepped even though the deterministic movement was correct.

The fix was presentation-only interpolation-like motion: a small continuous bob and rotation derived from render frame and vertical velocity. Collision state and gameplay coordinates were not changed.

A physics engine, animation state machine or tweening framework would have solved a much larger problem than existed.

## Sound became code, not an asset subsystem

The design draft expected embedded WAV cues. During runtime hardening, generated short cues proved simpler and more robust.

The final implementation generates PCM from a tiny table of frequency/duration/volume specifications and plays it through Ebitengine audio. If Linux audio is unavailable, the game continues silently.

The obsolete encoded sound bundle was removed before release.

## Packaging lesson

The first AppImage path launched on some runs but exposed FUSE/environment friction on the target Ubuntu desktop. A later native executable proved more reliable, while browser download handling exposed the separate issue of executable-bit preservation.

The release format was therefore simplified to a tar.gz containing the native executable with mode 755 preserved.

This is not a workaround layered on top of the product. It is a smaller delivery mechanism.

One consequence was easy to miss. Astro Answers' AppImage launcher chose the score directory; a bare executable has no launcher, so scores followed whatever directory it was started from. The executable now resolves the standard per-user data directory itself. That is one small function, not a configuration system.

## Assets are just files

During UAT, sprites were carried as base64 text. Nothing in the final runtime needed that encoding. The final release embeds five ordinary PNGs, as Astro Answers does, so the art is visible in the repository and decoded by one code path.

## What was deliberately rejected

V0.1 does not contain:

- levels or campaigns;
- upgrades or power-up inventories;
- currencies;
- multiple characters;
- player accounts;
- online leaderboards;
- telemetry;
- network services;
- runtime AI;
- a reusable game engine;
- ECS;
- event bus;
- physics engine;
- UI framework;
- animation framework;
- database.

The governing test remained:

> **Does every piece of machinery remove more complexity than it introduces?**
