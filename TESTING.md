# Dandelion Drift — Testing and Debugging Record

Testing was split between deterministic rule tests, long-run hardening and real human UAT. All three were required.

## Automated game-rule tests

`internal/game` runs without Ebitengine or a graphical window.

The suite covers:

- flap spending and lift/glide relationships;
- final-flap behaviour;
- ceiling and horizontal bounds;
- perching only on a valid gentle descent;
- recharge until full or nectar exhaustion;
- automatic hop-off when a flower is carried left;
- planting every fourth released flap seed;
- planting restrictions and bloom delay;
- planted-flower caps;
- bee, thistle, grass, pond and frog damage;
- guard/protection behaviour;
- gust influence;
- firefly chain scoring;
- 15-puff firefly damage and zero-flooring;
- soft empty-puff fall;
- completed-flight routing and clean restart;
- deterministic generation;
- long-world invariants and live-entity bounds;
- top-five score normalisation, qualification and persistence;
- score-file location resolved from `XDG_DATA_HOME` or the user data directory;
- missing and malformed score-file behaviour;
- final release balance: 50 starting puffs and flower recharge.

`cmd/dandeliondrift/assets_test.go` additionally verifies that all five embedded PNG sprites decode and that every procedural cue specification generates valid stereo PCM. `cmd/dandeliondrift/storage_test.go` verifies the high-score location.

Release commands are:

```bash
gofmt -w ./cmd ./internal
go vet ./...
go test -race -count=1 ./...
go build -trimpath -ldflags="-s -w" -o ./bin/dandelion-drift ./cmd/dandeliondrift
GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w" -o ./bin/dandelion-drift-linux-arm64 ./cmd/dandeliondrift
```

## Hardening tests

Chunk 5 added explicit long-run checks rather than relying only on feature tests.

These exercise:

- representative long deterministic runs;
- repeated completed flights to detect stale state;
- Space-only progression without horizontal control;
- state sanity over extended generation;
- live-entity budgets;
- falling-state immunity from further hazard damage.

The purpose is to prove that the small architecture remains stable under duration, not merely that individual mechanics pass isolated tests.

## Human UAT

Human UAT materially improved the game while keeping the architecture small.

### Resource duration

The first playable build exhausted the dandelion reserve too quickly. The reserve was increased substantially for testing, then deliberately brought back down after survival became too easy.

Final accepted starting reserve: **50**.

### Recovery loop

UAT tested alternative recharge/drain arrangements for flowers and fireflies. The final product rule is:

- flowers recharge;
- fireflies score but drain 15;
- other animals/hazards drain their normal amounts.

This restored a genuine lose condition without adding new systems.

### Health visibility

The original design intended the puff alone to communicate remaining reserve. With the larger reserve, that became too imprecise.

The HUD now shows the exact remaining count while the puff remains a proportional visual cue.

### Bunny fluidity

Gameplay motion was correct but the sprite presentation looked too rigid. A small continuous render-only bob and velocity-linked rotation were added. Gameplay coordinates and collision rules were unchanged.

### End presentation

The explicit "gentle landing" wording was removed. The visual behaviour remains soft; the UI no longer narrates something the player can already see.

## Runtime failures found before release

Several important failures were outside the deterministic game rules.

### Malformed embedded bunny asset

An early Linux UAT executable compiled successfully but failed at graphical startup because an embedded bunny base64 asset was malformed.

The correction was twofold:

1. move the live bunny sheet to a known-good embedded source;
2. add a decode gate so CI proves required embedded images are actually decodable rather than merely compilable.

This is why the final repository includes asset-level tests.

### Linux audio availability

A later launch path exposed Linux audio-device sensitivity. Startup was changed so unavailable audio causes a logged silent fallback rather than application failure.

No alternate audio subsystem was introduced.

### AppImage / desktop launch path

AppImage packaging was tested because Astro Answers uses it successfully. On the Dandelion target Ubuntu environment, AppImage/FUSE behaviour was not reliable enough to remain the sole delivery path.

A native executable then exposed a different desktop issue: browsers/archive handling may lose or ignore the executable bit when delivering a raw ELF.

The final native tar.gz preserves mode 755 and is the simpler V0.1 release format.

### Launch gate duration

The release build performs a graphical X11/Xvfb smoke launch against the actual production executable, rather than treating compilation as proof of startup.

## Publication cleanup

The approved gameplay snapshot at commit `74ec654e7ac7c92ff88befe3f2fd800431648173` still carried two obsolete encoded prototype artefacts:

- `assets/bunny.b64`;
- `assets/sounds.zip.b64`.

Neither was used by the final runtime path. They were removed during release-publication cleanup so dead UAT bytes are not embedded into the shipped binary.

No gameplay behaviour changed during that cleanup.

## Release benchmark corrections

A final benchmark against the Astro Answers repository found two release-level problems that automated tests had not caught.

### Score location depended on the launch directory

The game wrote `scores.json` to the current working directory. Astro Answers avoids this because its AppImage launcher changes into a per-user data directory first. The native tar.gz has no launcher, so launching from a file manager could put scores somewhere unpredictable, or fail to save in a read-only directory, and PRD acceptance criterion 18 was at risk.

The executable now resolves `$XDG_DATA_HOME/dandelion-drift/scores.json`, falling back to `~/.local/share/dandelion-drift/scores.json`. `SaveScores` already created missing directories. Three tests pin the path rules.

### Sprites were stored as base64 text

Four sprites were embedded as `.b64` text and the bunny as a Go base64 string constant. That inflated the source by about a third, hid the art from repository viewers and kept a second decode path alive. All five are now plain PNGs embedded with `//go:embed assets/*.png`, and the asset test decodes each one.

No gameplay behaviour changed.

## Play-test evidence for V0.1

The final game was human-tested on Linux with:

- full-screen launch;
- Space-only flight;
- optional horizontal drift;
- 50-puff starting reserve;
- flower recharge;
- firefly 15-puff cost;
- all major hazards;
- puff counter and proportional graphic;
- score/high-score flow;
- attract/demo scene;
- result scene;
- smoother bunny presentation;
- repeated fresh runs.

The final UAT judgement was that the game looked good and required no further gameplay or presentation changes beyond the final firefly cost of 15.

## Constrained-hardware benchmark

As with Astro Answers, the Garage constrained-hardware benchmark is a separate physical evidence step. It should record hardware, OS, displayed resolution, frame rate, CPU, memory, executable size and maximum live-entity counts.

Do not fabricate this evidence from CI or a desktop synthetic run. It is not a blocker to publishing V0.1, but remains an explicit physical benchmark item until recorded.
