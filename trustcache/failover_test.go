package trustcache_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-quicktest/qt"
	valkey "github.com/valkey-io/valkey-go"

	"github.com/gmb-eudi/go-verifier-helpers/trustcache"
)

// valkeyGetter adapts a valkey-go client to the trustcache.Getter seam.
type valkeyGetter struct{ c valkey.Client }

func (g valkeyGetter) Get(ctx context.Context, key string) ([]byte, error) {
	resp := g.c.Do(ctx, g.c.B().Get().Key(key).Build())
	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}
		return nil, err
	}
	return resp.AsBytes()
}

// TestFailoverWorkerDownServesUntilExpiryThenFailsClosed proves anchors
// materialized with TTL=valid_until keep serving after the writer stops
// writing, and vanish (fail closed) at expiry.
func TestFailoverWorkerDownServesUntilExpiryThenFailsClosed(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := valkey.NewClient(valkey.ClientOption{
		InitAddress:  []string{mr.Addr()},
		DisableCache: true, // miniredis has no CLIENT TRACKING
	})
	qt.Assert(t, qt.IsNil(err))
	defer client.Close()

	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	cacheTTL := 45 * time.Minute
	key := trustcache.AnchorSetKey(trustcache.TypePIDProvider, "LV")

	// The writer materializes an entry and is then "down" — no further
	// writes, no TTL extensions.
	entry := trustcache.AnchorSetEntry{
		SnapshotID: "snap-1",
		Type:       trustcache.TypePIDProvider,
		Territory:  "LV",
		FetchedAt:  now,
		ValidUntil: now.Add(cacheTTL), // valid_until honored: it IS the key TTL
		Anchors:    []trustcache.Anchor{{CertDER: []byte{0x30, 0x82}, Territory: "LV", Status: "granted", ValidUntil: now.Add(24 * time.Hour), TLSequence: 51}},
	}
	raw, err := json.Marshal(entry)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.IsNil(mr.Set(key, string(raw))))
	mr.SetTTL(key, cacheTTL)

	ctx := context.Background()
	rd := trustcache.NewReader(valkeyGetter{client}, func() time.Time { return now })

	// t+0: serves warm.
	got, err := rd.AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
	qt.Assert(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(len(got.Anchors), 1))

	// t+44m — worker still down, entry not yet expired: still serves.
	mr.FastForward(44 * time.Minute)
	_, err = rd.AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
	qt.Assert(t, qt.IsNil(err))

	// t+46m — past valid_until: Valkey evicted the key; the reader MUST fail
	// closed.
	mr.FastForward(2 * time.Minute)
	_, err = rd.AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
	qt.Check(t, qt.IsTrue(errors.Is(err, trustcache.ErrCacheExpired)))
}
