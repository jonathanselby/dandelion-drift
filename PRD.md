# Dandelion Drift — Product Requirements Document

**Version:** 0.1  
**Status:** Final — V0.1.0  
**Product name:** Dandelion Drift

## Problem

Many children's games create engagement by adding progression systems, rewards, inventories or layered controls. Dandelion Drift takes the opposite approach: one visible resource drives movement, risk and recovery.

A bunny floats through a meadow beneath a dandelion puff. Flapping spends the puff. Flowers restore it. Hazards and fireflies drain it. Some released seeds become future flowers.

> **Every flap spends your wings. Every flap plants tomorrow.**

The game should feel calm enough for a young child to enjoy immediately, but constrained enough that a run can end and another run is tempting.

## Product objective

Build a polished, lightweight endless side-scrolling flight game in Go in which one shrinking dandelion resource connects lift, health, recovery and the future world.

The V0.1 product question is:

> **Can one button, one visible resource and a small set of interacting rules make a child want another flight immediately?**

## Users

Primary users are children from roughly five years old playing on a keyboard.

Secondary users are siblings or adults chasing distance and score.

The repository also serves as a Jonno's Garage engineering example: a visually polished application produced from a deliberately small runtime design.

## Product principles

- complete play is possible with Space alone;
- Left and Right are optional, not mandatory;
- the game can be understood without reading instructions;
- failure is soft rather than frightening;
- the same puff is movement reserve and health;
- the remaining puff count is visible numerically and graphically;
- challenge comes from interacting rules, not feature accumulation;
- beginner runs should be forgiving; skilled runs can continue for several minutes;
- the player must be able to lose if they spend poorly or collide repeatedly.

## Core loop

```text
flap -> spend one puff -> glide -> avoid / accept risk -> land on flower -> recharge
  |                                                                ^
  +--> every fourth released seed may grow into a future flower ---+
  |
  +--> hazards and fireflies drain the puff
  |
  +--> puff reaches zero -> soft fall -> result -> another flight
```

## In scope

- endless right-to-left scrolling meadow and ponds;
- flap on Space;
- optional horizontal drift on Left/Right;
- 50-puff starting reserve;
- each flap spends exactly one puff;
- lift and glide depend on puffs remaining;
- exact remaining puff count shown in the HUD;
- puff graphic shrinking proportionally with remaining reserve;
- perching on flowers to recharge from limited nectar;
- every fourth released seed becoming a potential future flower;
- bees, thistles, grass, ponds, frogs and wind gusts;
- fireflies as a score/chain opportunity that drains 15 puffs on contact;
- soft end of flight when the puff reaches zero;
- score from distance plus firefly points;
- persistent local top-five table with one-to-three-character initials;
- attract screen with top five and a deterministic drifting demo bunny;
- immediate restart;
- polished 2D artwork, animation and restrained effects;
- concise procedural sound cues;
- deterministic core gameplay from a seed;
- Linux desktop V0.1;
- full-screen fixed 960×720 logical playfield with aspect preservation;
- Linux/ARM64 build proof in CI.

## Out of scope

- lives;
- multiplayer;
- levels, stages or a level editor;
- upgrades, currencies, unlocks or inventories;
- multiple playable characters;
- music;
- learning content;
- player accounts;
- online leaderboards;
- telemetry;
- network services;
- runtime AI;
- gamepad or touch input in V0.1;
- Windows, macOS or mobile packaging in V0.1.

## Functional requirements

### FR1 — Scrolling

The world scrolls from right to left continuously. Scroll speed begins slowly and rises gradually with distance to a fixed cap.

### FR2 — Flap and glide

Pressing Space during flight spends exactly one puff and applies upward lift.

A fuller puff produces a stronger flap and slower terminal descent. A depleted puff produces weaker lift and a faster glide.

The flap that spends the final puff still produces its lift before the bunny transitions to the empty-puff fall.

### FR3 — Horizontal drift

Left and Right move the bunny horizontally within bounded screen limits. The game remains fully playable without horizontal input.

### FR4 — Puff reserve

Each flight starts with **50 puffs**.

The reserve cannot exceed 50 or fall below zero.

The HUD shows the exact remaining number. The dandelion drawing also becomes visually smaller/thinner as the reserve falls.

### FR5 — Flowers and perching

Flowers have limited nectar.

A bunny descending gently onto a bloomed flower head perches. While perched, the flower restores one puff at a time until either:

- the puff reaches 50;
- the flower's nectar is exhausted; or
- the bunny takes off or is carried off the left edge.

Wild and planted flowers may have different nectar quantities.

### FR6 — Planting

Every fourth puff spent through flapping releases a gameplay seed that can plant ahead.

A released gameplay seed that reaches valid grass may sprout and later bloom into a flower. Seeds landing in water, too close to an existing flower or in an invalid obstacle area are lost.

Planting is bounded so the live world cannot grow without limit.

### FR7 — Hazards

| Hazard | V0.1 behaviour |
| --- | --- |
| Bee | removes 3 puffs and gives a short horizontal push |
| Thistle | removes 3 puffs and bounces the bunny upward |
| Grass | removes 1 puff and bounces the bunny upward |
| Pond | removes 4 puffs and splashes the bunny upward |
| Frog | removes 6 puffs and flings the bunny upward |
| Wind | changes vertical acceleration without directly changing puffs |

After a normal collision the bunny receives a short protection period to prevent repeated immediate damage. Ground/water bounce still occurs while protected so the bunny cannot sink into terrain.

### FR8 — Fireflies

Fireflies are deliberate risk/reward objects.

Touching one:

- collects it;
- removes **15 puffs** immediately, capped at the current reserve;
- scores `10 × current chain multiplier`;
- continues or starts a two-second chain;
- raises the chain multiplier up to ×5.

A firefly can therefore end the flight if fewer than 15 puffs remain.

### FR9 — End of flight

When the puff reaches zero, flapping stops and the bunny falls softly on the bare stem.

On ground or water contact the run ends. A result overlay remains briefly, showing the completed run before routing either to high-score initials entry or back to the attract screen.

No additional "gentle landing" message is required.

### FR10 — Fairness

Generated worlds must remain structurally passable.

The world generator keeps flowers sufficiently frequent, provides a gentle initial section and prevents impossible hazard placement around flower approaches. Ponds and bee patterns remain bounded.

### FR11 — Scoring

Score is distance travelled plus firefly chain points.

Distance is measured from world-camera progress. Fireflies contribute 10 points multiplied by the current chain multiplier.

### FR12 — High scores

One persistent top-five score table is stored locally.

A qualifying run allows one to three letters. Enter confirms. Empty initials are stored as `?`.

Only the five highest valid scores are retained.

### FR13 — Attract screen

The idle/start screen presents:

- Dandelion Drift title treatment;
- top-five scores;
- a deterministic demo bunny using the real game rules;
- a clear Space-to-start prompt.

It must look like part of the finished game, not a developer menu.

### FR14 — Visual quality

The visual system is soft, bright and high-contrast.

Quality is concentrated on:

- bunny motion that reads as fluid rather than stepped;
- the shrinking puff;
- flower landing and recharge;
- readable hazards and fireflies;
- the attract screen;
- the result screen;
- small responsive effects that do not obscure play.

### FR15 — Display

The game uses a fixed 960×720 logical playfield and launches full-screen.

The logical 4:3 image is uniformly scaled and centred. Wider or taller physical displays use black pillarboxing or letterboxing rather than stretching.

### FR16 — Privacy and runtime boundary

V0.1 contains no network access, accounts, telemetry, advertising or runtime model/AI dependency.

Only the local top-five score file persists.

## Acceptance criteria

V0.1 is accepted when automated and playable evidence shows that:

1. a flight starts with exactly 50 puffs;
2. each valid flap spends exactly one puff and raises the bunny;
3. more remaining puffs produce stronger lift and slower descent;
4. the final flap still produces lift before the empty-puff fall;
5. the exact remaining puff count is visible during play;
6. perching occurs only on a valid gentle descent onto a flower;
7. flowers recharge one puff at a time and stop at full reserve or exhausted nectar;
8. one in four released flap seeds becomes eligible to plant;
9. planting occurs only on valid grass and remains bounded;
10. each normal hazard produces the specified damage and physical reaction;
11. fireflies remove exactly 15 puffs, score correctly and can end a run;
12. firefly chains rise only within the chain window and cap at ×5;
13. the world generator remains deterministic for a fixed seed;
14. long generated worlds preserve the required safety/resource invariants;
15. gameplay remains playable using Space alone;
16. empty reserve transitions to a soft fall and then a completed run;
17. completed runs route cleanly to initials or attract and a new flight starts from clean state;
18. the top-five table persists correctly across application runs;
19. the attract screen and result screen are visually finished;
20. gameplay logic is testable without the renderer;
21. full-screen 4:3 geometry is preserved without stretching;
22. `gofmt`, `go vet`, race tests, Linux/amd64 build and Linux/ARM64 build pass in CI;
23. the production Linux x86-64 release executable survives the graphical launch smoke test.

## Definition of done

V0.1 is done when the acceptance criteria have executable or observable evidence, human UAT has approved the gameplay and presentation, the final repository documents describe the shipped system truthfully, and the release build can be published without carrying UAT scaffolding or obsolete assets.

The hand-written EDD is then compared independently with a code-derived implementation view. Material divergence is corrected or recorded rather than hidden.

The final product test remains:

> **She lands, and immediately flaps to go again.**
