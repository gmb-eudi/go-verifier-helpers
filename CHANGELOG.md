# Changelog

Notable changes to this library, newest first. Versions are git tags; this file is written
for whoever bumps the dependency.

## v0.0.5

**Requires Go 1.27.2.** The `go` directive moves up from 1.27.0, so a consumer on an older
toolchain will not build this version. **No source change**: every exported symbol, signature and
behaviour is exactly as in v0.0.4.

### Changed

- **`go` directive 1.27.0 → 1.27.2**, the minimum Go version a consumer needs. No vulnerability in
  Go 1.27.0 reached this library's code (`govulncheck` found none before the move either). The
  minimum moves with the sibling libraries, which do need Go 1.27.2's security fixes, so a consumer
  has one Go version to meet. Raise your own module's `go` directive to `1.27.2`; from there the go
  command downloads and uses that toolchain by itself.
- **Test-only dependencies moved, and none of them reaches your build**:
  `alicebob/miniredis/v2` 2.39.0 → 2.40.0, `valkey-io/valkey-go` 1.0.77 → 1.0.78, and the indirect
  `yuin/gopher-lua`, `rogpeppe/go-internal` and `golang.org/x/sys` with them.

### Notes

- The gate is green on Go 1.27.2: `go mod verify`, `go mod tidy -diff`, build, vet, `gofmt`,
  golangci-lint v2.14.0, and `go test -race` with **0 races**. `govulncheck` reports **no
  vulnerabilities found**.

- Repository hygiene, with no effect on code that uses the library: CI's linter moved to
  golangci-lint v2.14.0 (the earlier release cannot read Go 1.27.2's compiled standard library).

## v0.0.4

**Requires Go 1.27.0.** The `go` directive moves up from 1.26.6, so a consumer on an older
toolchain will not build this version. **No source change**: every exported symbol, signature and
behaviour is exactly as in v0.0.3, and nothing a consumer links moved.

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
- **Test-only dependencies moved, and none of them reaches your build**:
  `alicebob/miniredis/v2` 2.38.0 → 2.39.0, `valkey-io/valkey-go` 1.0.76 → 1.0.77, and the indirect
  `golang.org/x/sys` 0.43.0 → 0.47.0. **Measured, not assumed**: no non-test file in this library
  imports any of them, and `go list -deps ./...` — the packages a consumer actually links — contains
  none of the three. They exist for this library's own tests (an in-memory Valkey and a real client
  against it), so bumping to v0.0.4 changes nothing in a consumer's dependency graph beyond the
  toolchain line.

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
