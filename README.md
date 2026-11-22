# feralos-dbus-audit

ADR-010 supervision audit for FeralOS (standalone Go binary): discovers
every D-Bus activation target on BOTH buses (session + system), installs
the aports supervision package when a target has no unit, and reports the
full coverage state. Runs at install (twice: INSTALL + REPORT passes), at
boot (oneshot, writes `/var/log/feralos-dbus-audit.log`), and manually (CLI).

This is the standalone home of the tool (moved out of the feralos installer
repo — the aports package `feralos-dbus-audit` builds from this repo's
release tag).

## Usage

```
feralos-dbus-audit [--log-file <path>] [--user <name>] [--no-fail]
```

- `--log-file` — also write the report to this file (boot mode)
- `--no-fail` — report only: always exit 0 (installer mode)
- without `--no-fail` a missing supervision package exits 1 (boot + CLI are strict)

Exit semantics on OpenRC: a missing DINIT unit is expected (the vendor
dbus-daemon-launch-helper activates natively) — not a gap; the strict gate
applies to dinit boots only.

## Build

```
go build -o feralos-dbus-audit .
```

Static binary (CGO disabled on target builds). The `shell` package is a
minimal silent-exec core extracted from the FeralOS utils/shell package.

## CI

The full FeralOS gate, ALL BLOCKING: gofmt, vet, golangci-lint (gosec
included), tests, build, govulncheck, gitleaks, yamllint,
editorconfig-checker, typos, markdownlint-cli2. Same posture as the
feralos installer repo.

## License

ISC — see `LICENSE`. Contribution terms in `CONTRIBUTING.md`.
