# Changelog

Notable changes to this library, newest first. Versions are git tags; this file is written
for whoever bumps the dependency.

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
