# go-verifier-helpers

Small, framework-free Go packages defining the Valkey wire/cache contracts
shared between the components of an EUDI wallet verifier. Importing these gives
producer and consumer a single source of truth for key names and value schemas,
so no key string is ever restated in two places.

```
go-verifier-helpers/
  handoffwire/   result-handoff Valkey key & queue wire contract:
                 one producer enqueues an Envelope, one consumer reads and
                 delivers it. The verification report is value-free (claim
                 NAMES only); disclosed claim VALUES travel only inside the
                 encrypted ResultJWE.
  trustcache/    trust-anchor cache Valkey key contract + a consumer-side
                 Reader: a single writer materializes trusted-list anchors
                 (valid_until honored as the key TTL, so expiry fails closed);
                 consumers read them warm behind a minimal Getter seam.
                 Reader.AnchorSet distinguishes a per-territory entry that is
                 missing because the type's cache is degraded (fails closed,
                 ErrCacheExpired) from one missing because the type's own
                 freshness record is fresh and the territory legitimately has
                 no anchors (returns a confirmed-empty set, nil error) — a
                 writer only materializes per-territory keys for territories
                 that actually have data.
```

```go
import "github.com/gmb-eudi/go-verifier-helpers/handoffwire"
import "github.com/gmb-eudi/go-verifier-helpers/trustcache"
```

Requires Go 1.26. Production code in both packages is **stdlib-only** — no web
framework, no logging, bring your own Valkey client. `trustcache`'s test suite
additionally uses `github.com/go-quicktest/qt`,
`github.com/alicebob/miniredis/v2`, and `github.com/valkey-io/valkey-go` as
test-only dependencies; these never enter a consumer's production build.

## License

MIT.
