package trustcache

import "time"

// Anchor is the cached projection of one trust anchor. CertDER is the raw
// certificate (public material — loggable identifiers only, never attribute
// values ride through this cache). ValidUntil mirrors the anchor's own
// validity (notAfter) as served by the trust service; TLSequence is the
// issuing trusted-list sequence for provenance; UseCases lists the accredited
// use cases this anchor is scoped to (empty = not use-case-scoped).
type Anchor struct {
	CertDER    []byte    `json:"certDer"`
	Territory  string    `json:"territory"`
	Status     string    `json:"status"`
	ValidUntil time.Time `json:"validUntil"`
	TLSequence int64     `json:"tlSequence"`
	UseCases   []string  `json:"useCases,omitempty"`
}

// AnchorSetEntry is the value of one trust:anchors:<type>:<territory> key.
// ValidUntil is the cache-validity horizon (fetchedAt + cache TTL); the key's
// Valkey TTL is set to the same instant — TTL expiry IS the fail-closed
// mechanism. The snapshot is authoritative: an anchor absent from the entry
// does not exist.
type AnchorSetEntry struct {
	SnapshotID    string    `json:"snapshotId"`
	Type          string    `json:"type"`
	Territory     string    `json:"territory"`
	FetchedAt     time.Time `json:"fetchedAt"`
	ValidUntil    time.Time `json:"validUntil"`
	UpstreamStale bool      `json:"upstreamStale"` // upstream marked the source stale at fetch time
	Anchors       []Anchor  `json:"anchors"`
}

// TypeFreshness is the value of trust:freshness:<type> — updated on every
// successful poll including 304s, so consumers can tell "worker alive, data
// unchanged" from "worker dead, data aging out".
type TypeFreshness struct {
	SnapshotID    string    `json:"snapshotId"`
	FetchedAt     time.Time `json:"fetchedAt"`
	ValidUntil    time.Time `json:"validUntil"`
	UpstreamStale bool      `json:"upstreamStale"`
}

// TerritoryHealth mirrors one territory summary from the trust service's
// snapshot endpoint.
type TerritoryHealth struct {
	TLSequence uint64     `json:"tlSequence"`
	Stale      bool       `json:"stale"`
	NextUpdate *time.Time `json:"nextUpdate,omitempty"`
}

// SnapshotHealth is the value of SnapshotHealthKey — the writer's freshness
// telemetry: trusted-list sequence, per-territory staleness, upstream
// staleness, and the pending-bootstrap alert flag.
type SnapshotHealth struct {
	SnapshotID       string                     `json:"snapshotId"`
	LOTLSequence     uint64                     `json:"lotlSequence"`
	CheckedAt        time.Time                  `json:"checkedAt"`
	UpstreamStale    bool                       `json:"upstreamStale"`
	PendingBootstrap bool                       `json:"pendingBootstrap"`
	Territories      map[string]TerritoryHealth `json:"territories"`
}
