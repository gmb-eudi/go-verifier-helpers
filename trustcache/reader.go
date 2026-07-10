package trustcache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrCacheExpired reports a missing (TTL-evicted or never-written) cache
// entry. Consumers MUST fail closed on it — never serve stale anchors past
// their validity.
var ErrCacheExpired = errors.New("trustcache: anchor cache entry missing or expired")

// Getter is the minimal read seam over a Valkey client. Get returns
// (nil, nil) when the key does not exist — absence is data here, not an
// error, so the Reader can distinguish expiry (fail closed) from transport
// failure (also fail closed, but reported distinctly).
type Getter interface {
	Get(ctx context.Context, key string) ([]byte, error)
}

// Reader is the consumer-side view of the trust cache, over the Getter seam.
type Reader struct {
	g   Getter
	now func() time.Time
}

// NewReader builds a Reader; nil clock = time.Now (injectable for tests).
func NewReader(g Getter, now func() time.Time) *Reader {
	if now == nil {
		now = time.Now
	}
	return &Reader{g: g, now: now}
}

// AnchorSet returns the materialized anchor set for (type, territory), or
// ErrCacheExpired when the key is absent. Key TTL (set by the writer to the
// entry's ValidUntil) is the expiry authority; the embedded timestamps are
// provenance for reports and for a consumer's in-memory grace policy.
func (r *Reader) AnchorSet(ctx context.Context, keyType, territory string) (*AnchorSetEntry, error) {
	raw, err := r.g.Get(ctx, AnchorSetKey(keyType, territory))
	if err != nil {
		return nil, fmt.Errorf("trustcache: get anchor set: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%w: %s/%s", ErrCacheExpired, keyType, NormalizeTerritory(territory))
	}
	var e AnchorSetEntry
	if err := json.Unmarshal(raw, &e); err != nil {
		return nil, fmt.Errorf("trustcache: decode anchor set %s/%s: %w", keyType, territory, err)
	}
	return &e, nil
}

// Freshness returns the per-type freshness record, ErrCacheExpired when absent.
func (r *Reader) Freshness(ctx context.Context, keyType string) (*TypeFreshness, error) {
	raw, err := r.g.Get(ctx, FreshnessKey(keyType))
	if err != nil {
		return nil, fmt.Errorf("trustcache: get freshness: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%w: freshness %s", ErrCacheExpired, keyType)
	}
	var f TypeFreshness
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("trustcache: decode freshness %s: %w", keyType, err)
	}
	return &f, nil
}

// SnapshotHealth returns the writer's snapshot telemetry, ErrCacheExpired
// when absent (writer never ran or its telemetry TTL lapsed — consumers
// treat that as degraded).
func (r *Reader) SnapshotHealth(ctx context.Context) (*SnapshotHealth, error) {
	raw, err := r.g.Get(ctx, SnapshotHealthKey)
	if err != nil {
		return nil, fmt.Errorf("trustcache: get snapshot health: %w", err)
	}
	if raw == nil {
		return nil, fmt.Errorf("%w: snapshot health", ErrCacheExpired)
	}
	var h SnapshotHealth
	if err := json.Unmarshal(raw, &h); err != nil {
		return nil, fmt.Errorf("trustcache: decode snapshot health: %w", err)
	}
	return &h, nil
}
