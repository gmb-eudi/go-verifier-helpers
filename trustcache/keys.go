// Package trustcache is the Valkey key contract of a trust-anchor cache.
//
// A single writer materializes trusted-list anchors into namespaced Valkey
// keys (valid_until honored as the key TTL); consumers read the same keys to
// serve anchors warm without a thundering herd on the upstream trust service.
// Writer and consumers both import THIS package for key names and value
// schemas — never restate a key string elsewhere. Production code here is
// stdlib-only; bring your own Valkey client behind the Getter seam.
package trustcache

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Wire anchor-type taxonomy — the trust-anchor `type=` filter values. These
// strings are load-bearing: they appear in Valkey keys and must match the
// upstream trust-service API.
const (
	TypePIDProvider    = "pid_provider"
	TypeQEAAProvider   = "qeaa_provider"
	TypePubEAAProvider = "pub_eaa_provider"
	TypeEAAProvider    = "eaa_provider"
	TypeWalletProvider = "wallet_provider"
	TypeAccessCA       = "access_ca"
	TypeWRPRCIssuer    = "wrprc_issuer"
	// Status-signer anchor types: the Token Status List / Identifiers List
	// signer may be a distinct service from the issuer.
	TypePIDProviderStatus    = "pid_provider_status"
	TypeQEAAProviderStatus   = "qeaa_provider_status"
	TypePubEAAProviderStatus = "pub_eaa_provider_status"
	TypeEAAProviderStatus    = "eaa_provider_status"
)

// AllTypes returns the full taxonomy in stable order.
func AllTypes() []string {
	return []string{
		TypePIDProvider, TypeQEAAProvider, TypePubEAAProvider, TypeEAAProvider,
		TypeWalletProvider, TypeAccessCA, TypeWRPRCIssuer,
		TypePIDProviderStatus, TypeQEAAProviderStatus, TypePubEAAProviderStatus, TypeEAAProviderStatus,
	}
}

// Fixed keys.
const (
	// SnapshotHealthKey holds the JSON SnapshotHealth written by the writer's
	// snapshot telemetry poll; readiness probes consume it.
	SnapshotHealthKey = "trust:freshness:snapshot"

	// StatusRefsKey is a ZSET of recently referenced status-list URIs:
	// member = URI, score = unix seconds of last reference. Consumers ZADD on
	// every status resolution; the writer prefetches the top-N and trims the
	// set by rank.
	StatusRefsKey = "trust:statuslist:refs"
)

// AnchorSetKey is the materialized anchor set for one (type, territory):
// value = JSON AnchorSetEntry, TTL = the entry's ValidUntil horizon.
func AnchorSetKey(keyType, territory string) string {
	return "trust:anchors:" + keyType + ":" + NormalizeTerritory(territory)
}

// TerritoryIndexKey lists (JSON []string) the territories currently
// materialized for a type — the swap uses it to DEL vanished territories.
func TerritoryIndexKey(keyType string) string {
	return "trust:anchors:territories:" + keyType
}

// ETagKey holds the snapshot id (strong ETag) last materialized for a type.
func ETagKey(keyType string) string {
	return "trust:anchors:etag:" + keyType
}

// FreshnessKey holds the JSON TypeFreshness for one type, refreshed on every
// successful poll (200 and 304) — the readiness signal for consumers.
func FreshnessKey(keyType string) string {
	return "trust:freshness:" + keyType
}

// StatusListKey addresses one cached raw status-list token by its URI hash.
// Hashed so keys stay uniform and raw issuer URIs never appear in keyspace
// listings or logs.
func StatusListKey(uri string) string {
	sum := sha256.Sum256([]byte(uri))
	return "trust:statuslist:" + hex.EncodeToString(sum[:])
}

// NormalizeTerritory upper-cases ISO 3166-1 alpha-2 codes; the empty string
// (EU-level lists — cross-country anchors honored when country="") maps to "EU".
func NormalizeTerritory(code string) string {
	if code == "" {
		return "EU"
	}
	return strings.ToUpper(code)
}
