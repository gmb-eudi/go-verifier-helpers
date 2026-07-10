package trustcache_test

import (
	"strings"
	"testing"

	"github.com/go-quicktest/qt"

	"github.com/gmb-eudi/go-verifier-helpers/trustcache"
)

func TestKeyLayout(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"anchor_set", trustcache.AnchorSetKey(trustcache.TypePIDProvider, "LV"), "trust:anchors:pid_provider:LV"},
		{"anchor_set_normalizes_territory", trustcache.AnchorSetKey(trustcache.TypeAccessCA, "lv"), "trust:anchors:access_ca:LV"},
		{"anchor_set_eu_level_for_empty", trustcache.AnchorSetKey(trustcache.TypeWalletProvider, ""), "trust:anchors:wallet_provider:EU"},
		{"territory_index", trustcache.TerritoryIndexKey(trustcache.TypePIDProvider), "trust:anchors:territories:pid_provider"},
		{"etag", trustcache.ETagKey(trustcache.TypePIDProvider), "trust:anchors:etag:pid_provider"},
		{"freshness", trustcache.FreshnessKey(trustcache.TypePIDProvider), "trust:freshness:pid_provider"},
		{"snapshot_health", trustcache.SnapshotHealthKey, "trust:freshness:snapshot"},
		{"status_refs", trustcache.StatusRefsKey, "trust:statuslist:refs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qt.Check(t, qt.Equals(tt.got, tt.want))
		})
	}
}

// Status-list keys hash the URI: URIs contain ':' '/' and unbounded length —
// hashing keeps the keyspace uniform and avoids logging raw issuer URIs.
func TestStatusListKeyIsHashed(t *testing.T) {
	k := trustcache.StatusListKey("https://issuer.example.eu/statuslists/1")
	qt.Check(t, qt.IsTrue(strings.HasPrefix(k, "trust:statuslist:")))
	qt.Check(t, qt.Equals(len(k), len("trust:statuslist:")+64)) // sha256 hex
	// Deterministic + distinct per URI.
	qt.Check(t, qt.Equals(k, trustcache.StatusListKey("https://issuer.example.eu/statuslists/1")))
	qt.Check(t, qt.Not(qt.Equals(k, trustcache.StatusListKey("https://issuer.example.eu/statuslists/2"))))
}

// AllTypes covers the full 11-type taxonomy (7 issuer/CA types + 4
// status-signer types).
func TestAllTypesCoversTaxonomy(t *testing.T) {
	qt.Assert(t, qt.DeepEquals(trustcache.AllTypes(), []string{
		"pid_provider", "qeaa_provider", "pub_eaa_provider", "eaa_provider",
		"wallet_provider", "access_ca", "wrprc_issuer",
		"pid_provider_status", "qeaa_provider_status", "pub_eaa_provider_status", "eaa_provider_status",
	}))
}
