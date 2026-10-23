# Dandelion Drift — Product Requirements Document

**Version:** 0.1
**Status:** Draft for V0.1.0
**Product name:** Dandelion Drift

## The idea

- Dandelion Drift is a gentle side-scrolling flying game for children from 5 years and up.
- It takes the calm of cosy games and adds the pull of a game you want to play again.
- A bunny floats across meadows and ponds holding a dandelion puff.
- Each flap blows a seed away, so the puff shrinks and she flies a little less well.
- Landing softly on a flower regrows her puff.
- The seeds she blows away drift ahead, land and grow into new flowers she can reach later.

> **Every flap spends your wings. Every flap plants tomorrow.**

## Why it works

- **Floating is calm; something running out is exciting.** The puff makes a peaceful game tense without making it scary.
- **One thing to understand.** The puff is lift, health and fuel at once, and it is drawn with exactly the seeds remaining, so its state needs no explanation.
- **Spending has a future.** Flapping hard now plants landing spots later, so even a mistake leaves something good behind.
- **Rest is a decision.** Perching regrows seeds, but the flower drifts away with the world, so staying longer is a trade.
- **Nothing is frightening.** Bees bump, the frog is comic, and the end of a flight is a soft landing.
- **One button.** Space flaps. Left and Right are optional.

## Product objective

Build a polished, lightweight endless side-scrolling flight game in Go in which one shrinking puff connects lift, health and the world ahead.

The V0.1 product question is:

> **Can one button and one puff make a child flap to go again the moment she lands?**

## Players

- Primary: children from 5 years, playing on a keyboard, who may not read yet.
- Secondary: siblings and adults chasing distance and firefly chains.

The repository is also a Jonno's Garage engineering example, built with the same method and budget as Astro Answers.

## Design principles

- Playable with Space alone.
- Playable without reading.
- Nothing scary; the end of a flight is soft.
- Beginner flights of about 30 seconds, skilled flights of several minutes.
- The puff is the only health display; no numbers needed.

## Core loop

```text
flap (rise, lose a seed) -> glide -> perch on a flower (regrow) -> take off
   |                                                                  ^
   +--> spent seeds drift ahead and grow into flowers ----------------+
   |
   +--> hazards knock seeds off ... last seed gone -> soft landing -> again
```

## In scope

- endless right-to-left scrolling meadow and ponds;
- flap on Space; gentle horizontal drift on Left/Right;
- dandelion puff of 24 seeds that is lift and health;
- each flap releases one seed;
- lift and glide depend on seeds remaining;
- perching on flowers to regrow seeds;
- released seeds planting new flowers further along;
- bees, thistles, ponds with a frog, wind gusts;
- fireflies with a chain bonus;
- gentle end of flight when the puff is empty;
- score from distance and fireflies;
- persistent top-five table with one-to-three letter initials;
- attract screen with the top five and a drifting demo bunny;
- immediate restart;
- high-quality 2D art and a code-drawn puff that visibly loses seeds;
- concise sound effects;
- deterministic core gameplay from a seed;
- Linux desktop V0.1, full-screen 4:3, Linux/ARM64 build proof.

## Out of scope

- lives;
- versus or multiplayer modes;
- levels, stages or a level editor;
- power-ups, upgrades, currencies, unlocks;
- multiple characters;
- music;
- learning content;
- accounts, online leaderboards, telemetry, network;
- runtime AI;
- gamepad and touch input in V0.1;
- Windows, macOS and mobile packaging in V0.1.

## Functional requirements

### FR1 — Scrolling

The world scrolls right to left continuously. Scroll speed starts slow and rises gradually with distance to a cap.

### FR2 — Flap and glide

Space flaps: the bunny rises. Gravity pulls her down, but the puff slows her fall to a glide. A fuller puff gives a stronger flap and a slower glide; a thin puff gives a weak flap and a faster fall.

### FR3 — Horizontal drift

Left and Right drift the bunny within the left two-thirds of the screen. Play is complete without them.

### FR4 — Seeds

The puff holds up to 24 seeds and starts full. Each flap releases one seed; hazards knock seeds off. The puff is drawn with exactly as many seeds as remain.

### FR5 — Flowers and perching

Flowers stand on the ground at varying heights. Floating gently down onto a flower head perches the bunny.

While perched she regrows seeds, one at a time, until the puff is full or the flower is spent. The flower travels with the world, so staying perched carries her toward the left edge. Flapping takes off; at the left edge she hops off automatically.

### FR6 — Planting

One released seed in four drifts ahead on the wind, lands on grass, sprouts and becomes a flower a few seconds later. Seeds that land in water are lost.

### FR7 — Hazards

| Hazard | Behaviour |
| --- | --- |
| Bee | drifts toward the bunny with a gentle bob; a bump knocks off 3 seeds |
| Thistle | tall spiky plant; touching knocks off 3 seeds and bounces her up |
| Grass | touching bare ground bounces her up and knocks off 1 seed |
| Pond | touching water splashes her up and knocks off 4 seeds |
| Frog | sits in a pond and leaps when she floats low overhead; a catch knocks off 6 seeds and flings her up |
| Wind | updrafts lift, downdrafts push down; drawn as streaks |

After any knock she is briefly protected and blinks.

### FR8 — Fireflies

Fireflies float in loose arcs. Collecting one scores points; collecting the next within two seconds raises a chain multiplier, up to ×5.

### FR9 — End of flight

When the last seed is gone the puff pops and the bunny floats down on the bare stem and lands softly. A short overlay shows distance and fireflies, then initials or the attract screen.

### FR10 — Fairness

The next flower is never further away than a full puff can glide at the current scroll speed, and no hazard sits directly over a flower head.

### FR11 — Scoring

Score = metres travelled + firefly points (10 × chain multiplier each).

### FR12 — High scores

One persistent top-five table; one to three letters, Enter to confirm.

### FR13 — Attract screen

Title, top five, a demo bunny drifting across, "Press Space".

The attract/start screen is a first-class game scene, not application plumbing. It must use the same deliberate graphical finish as gameplay, be immediately recognisable as the same world, and never look like a developer menu.

### FR14 — Visual quality

Soft, bright, high-contrast art executed to a high standard, concentrated on:

- the puff, readable at a glance as full, half or nearly gone;
- seeds blowing away on every flap;
- the perch and seed-by-seed regrow;
- the frog leap, which must be funny;
- the attract/start scene, which must be graphically deliberate and finished;
- the end-of-flight scene, which must give the run a visually satisfying conclusion before returning to the attract screen.

### FR15 — Display

Fixed 960×720 logical playfield, full-screen, aspect preserved with black bars.

## Acceptance criteria

V0.1 is accepted when automated and playable evidence shows that:

1. the world scrolls and speed rises to a cap;
2. a flap raises the bunny and releases exactly one seed;
3. lift and glide vary with seeds as specified;
4. perching happens only on a gentle descent onto a flower head;
5. perching regrows seeds until full or the flower is spent;
6. perching carries the bunny left and auto-hops at the edge;
7. one in four released seeds plants a flower on grass, none in water;
8. each hazard removes the specified seeds, with protection afterwards;
9. the frog leaps only when the bunny is low overhead;
10. firefly chains multiply and reset after two seconds;
11. the flight ends gently when seeds reach zero;
12. FR10 holds across generated worlds, proved by test;
13. the top-five table persists between runs;
14. the game is playable with Space alone;
15. gameplay logic is testable without the renderer and replays deterministically from a seed;
16. full-screen 4:3 presentation with letterboxing;
17. `gofmt`, `go vet`, race tests, Linux/amd64 and Linux/ARM64 builds pass in CI;
18. the attract/start and end-of-flight scenes are visually finished to the same standard as gameplay, and a completed flight returns cleanly to the attract screen ready for a fresh game.

## Definition of done

V0.1 is done when all acceptance criteria have executable or observable evidence, a child has flown and landed without help, the score table persists, and the visual execution makes the small implementation invisible.

The hand-written EDD is then recovered independently from the implementation and compared with the intended design.

The product test is:

> **She lands, and immediately flaps to go again.**
