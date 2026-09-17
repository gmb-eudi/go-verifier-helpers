# Changelog

Notable changes to this library, newest first. Versions are git tags; this file is written
for whoever bumps the dependency.

## v0.0.4

**Requires Go 1.27.0.** The `go` directive moves up from 1.26.6, so a consumer on an older
toolchain will not build this version. Nothing else changed here — no source, no signature, no
message text, and the dependency graph is untouched.

### Added

- **`LICENSE`** — the MIT text this library has always been offered under, and which its README has
  always stated, is now in the repository. Until now the published source carried the statement
  without the grant, which left a consumer unable to rely on it. The text is byte-identical to the
  other libraries in this organisation: MIT, `Copyright (c) 2026 go-make-bytes contributors`.
  **v0.0.3 and earlier remain as published** — a module version on the Go proxy is immutable, so the
  licence reaches consumers from this version on.
- **`SECURITY.md`** — private vulnerability reporting through GitHub Security Advisories, the same
  wording as the sibling libraries; and **`CODEOWNERS`**, so dependency and workflow changes need a
  maintainer review.

### Changed

- **`go` directive 1.26.6 → 1.27.0** — the minimum Go version a consumer needs. The services
  in this project already required 1.27.0 while the libraries were the half still behind, so
  they are brought up together and the whole codebase now asks for one toolchain.

### Notes

- The gate is green on the new directive: `go mod verify`, `go mod tidy -diff`, build, vet,
  `gofmt`, and `go test -race` with **0 races** under a Go 1.27.0 toolchain. `govulncheck`
  reports **no vulnerabilities found** — this library requires no `golang.org/x/crypto`, so
  it does not carry the module-level advisory its siblings do.

- The README's "Requires Go 1.26" line moved with the directive.

- Repository hygiene, with no effect on code that uses the library: CI now also runs on pushes
  to `develop`, the pinned GitHub Actions moved to their current commits, the `setup-go` pin
  rolled forward to v7.0.0, and `.gitattributes` now pins its own line endings.

## v0.0.3

No source change: every exported symbol, signature and behaviour is exactly as in v0.0.2.
This release exists to move the declared Go version.

### Notes

- The `go` directive is now `1.26.6`, up from `1.26.4`, and that is the one thing a consumer
  feels: it is the minimum Go version required to build this module. The bump is deliberate
  rather than incidental — the 1.26.4 and 1.26.5 standard libraries carry security fixes that
  arrived in 1.26.6, and a library that declares an older patch lets a consumer build against
  the unfixed one without noticing.
- The repository now runs the same build, formatting, lint and vulnerability gate as the other
  libraries here, on every push and pull request. No shipped code changed as a result — the
  gate found nothing in this module.
