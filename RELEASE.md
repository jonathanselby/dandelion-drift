# Release checklist

## Contract

- [x] `VERSION` is `0.1.0`.
- [x] PRD is final for V0.1.0, matches shipped behaviour and records that no V0.2 direction is committed.
- [x] EDD is final and aligned to the approved implementation.
- [x] V0.1 is Linux-only, full-screen and keyboard-first.
- [x] No network, telemetry, accounts, advertising or runtime AI.
- [x] Final UAT balance is 50 starting puffs, flower recharge, firefly drain 15.

## Engineering gate

- [x] `gofmt` clean.
- [x] `go vet ./...` passes in CI.
- [x] `go test -race -count=1 ./...` passes in CI.
- [x] Linux/amd64 build passes in CI.
- [x] Linux/arm64 build passes in CI.
- [x] hardening and final-balance acceptance tests pass.
- [x] embedded PNG decode, procedural sound generation and high-score location are tested.
- [x] production Linux x86-64 executable passes graphical launch smoke test.

## Evidence

- [x] PRD acceptance criteria map to tests or observable play evidence.
- [x] code-derived EDD is pinned to the final source commit `d93668ef91d7fc743e1ec4f56995240c97a23adb`.
- [x] generated comparison records design/implementation matches and material corrections.
- [x] design process is recorded in `DESIGN-NOTES.md`.
- [x] automated, hardening, UAT and packaging/runtime failures are recorded in `TESTING.md`.
- [ ] constrained-hardware benchmark recorded on the standard Garage target.

## Publication

- [x] README is a finished product/repository front page rather than build-session notes.
- [x] obsolete UAT packaging workflows are absent.
- [x] obsolete embedded prototype assets are absent; sprites are plain PNG files.
- [x] high scores persist in the user data directory regardless of launch location.
- [x] release workflow produces one Linux x86-64 end-user archive.
- [x] Linux ARM64 remains CI proof rather than unnecessary release clutter.
- [x] MIT licence included.

## Release rule

The source and V0.1 release may be published before the physical constrained-hardware benchmark. The benchmark remains an explicit open Garage evidence item rather than a fabricated claim.

Manual publication steps remain outside automation: create tag `v0.1.0`, publish GitHub Release, attach the x86-64 archive, and complete the About description.
