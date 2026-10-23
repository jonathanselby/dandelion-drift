# Release checklist

## Contract

- [x] `VERSION` is `0.1.0`.
- [x] PRD is final for V0.1.0 and matches shipped behaviour.
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
- [x] embedded image decode and procedural sound generation are tested.
- [x] production Linux x86-64 executable passes graphical launch smoke test.

## Evidence

- [x] PRD acceptance criteria map to tests or observable play evidence.
- [x] code-derived EDD is pinned to approved gameplay snapshot `74ec654e7ac7c92ff88befe3f2fd800431648173`.
- [x] generated comparison records design/implementation matches and material corrections.
- [x] design process is recorded in `DESIGN-NOTES.md`.
- [x] automated, hardening, UAT and packaging/runtime failures are recorded in `TESTING.md`.
- [ ] constrained-hardware benchmark recorded on the standard Garage target.

## Publication

- [x] README is a finished product/repository front page rather than build-session notes.
- [x] obsolete UAT packaging workflows are absent.
- [x] obsolete embedded prototype assets are absent.
- [x] release workflow produces one Linux x86-64 end-user archive.
- [x] Linux ARM64 remains CI proof rather than unnecessary release clutter.
- [x] MIT licence included.

## Release rule

The source and V0.1 release may be published before the physical constrained-hardware benchmark. The benchmark remains an explicit open Garage evidence item rather than a fabricated claim.

Manual publication steps remain outside automation: create tag `v0.1.0`, publish GitHub Release, attach the x86-64 archive, and complete the About description.
