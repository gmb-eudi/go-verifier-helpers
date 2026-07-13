package trustcache_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-quicktest/qt"

	"github.com/gmb-eudi/go-verifier-helpers/trustcache"
)

type fakeGetter map[string][]byte

func (f fakeGetter) Get(_ context.Context, key string) ([]byte, error) {
	v, ok := f[key]
	if !ok {
		return nil, nil // absent — the Getter contract
	}
	return v, nil
}

func entryJSON[T any](t *testing.T, e T) []byte {
	t.Helper()
	raw, err := json.Marshal(e)
	qt.Assert(t, qt.IsNil(err))
	return raw
}

func TestReaderAnchorSet(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	ctx := context.Background()

	valid := trustcache.AnchorSetEntry{
		SnapshotID: "snap-1",
		Type:       trustcache.TypePIDProvider,
		Territory:  "LV",
		FetchedAt:  now.Add(-5 * time.Minute),
		ValidUntil: now.Add(40 * time.Minute),
		Anchors:    []trustcache.Anchor{{CertDER: []byte{0x30, 0x82}, Territory: "LV", Status: "granted", ValidUntil: now.Add(24 * time.Hour), TLSequence: 51}},
	}

	t.Run("hit", func(t *testing.T) {
		g := fakeGetter{trustcache.AnchorSetKey(trustcache.TypePIDProvider, "LV"): entryJSON(t, valid)}
		got, err := trustcache.NewReader(g, clock).AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
		qt.Assert(t, qt.IsNil(err))
		qt.Check(t, qt.Equals(got.SnapshotID, "snap-1"))
		qt.Check(t, qt.Equals(len(got.Anchors), 1))
	})

	// Negative first: fail closed on every degraded shape.
	t.Run("missing_key_fails_closed", func(t *testing.T) {
		_, err := trustcache.NewReader(fakeGetter{}, clock).AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
		qt.Check(t, qt.IsTrue(errors.Is(err, trustcache.ErrCacheExpired)))
	})

	t.Run("corrupt_value_is_error_not_data", func(t *testing.T) {
		g := fakeGetter{trustcache.AnchorSetKey(trustcache.TypePIDProvider, "LV"): []byte("{not json")}
		_, err := trustcache.NewReader(g, clock).AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
		qt.Check(t, qt.IsNotNil(err))
		qt.Check(t, qt.IsFalse(errors.Is(err, trustcache.ErrCacheExpired))) // distinct: corruption is not expiry
	})

	t.Run("territory_lookup_normalized", func(t *testing.T) {
		g := fakeGetter{trustcache.AnchorSetKey(trustcache.TypePIDProvider, "LV"): entryJSON(t, valid)}
		got, err := trustcache.NewReader(g, clock).AnchorSet(ctx, trustcache.TypePIDProvider, "lv")
		qt.Assert(t, qt.IsNil(err))
		qt.Check(t, qt.Equals(got.Territory, "LV"))
	})

	// missing_key_but_type_fresh: the per-territory key was never written,
	// but the type's OWN freshness record proves the most recent sync
	// completed and is not stale/expired. That combination means "checked,
	// confirmed zero anchors for this territory" — a writer only
	// materializes per-territory keys for territories that actually have
	// data, while it refreshes the type's freshness record on every cycle
	// regardless. Report this as a confirmed-empty set, not an error, so a
	// caller resolving a chain across multiple territories can keep trying
	// the next one instead of aborting on a false "cache expired".
	t.Run("missing_key_but_type_fresh_is_confirmed_empty_not_error", func(t *testing.T) {
		fr := trustcache.TypeFreshness{SnapshotID: "snap-9", FetchedAt: now.Add(-5 * time.Minute), ValidUntil: now.Add(40 * time.Minute)}
		g := fakeGetter{trustcache.FreshnessKey(trustcache.TypePIDProvider): entryJSON(t, fr)}
		got, err := trustcache.NewReader(g, clock).AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
		qt.Assert(t, qt.IsNil(err))
		qt.Check(t, qt.Equals(len(got.Anchors), 0))
	})

	// missing_key_and_no_freshness_record: fail-closed negative — no
	// disambiguating signal exists at all, so behave exactly as before.
	t.Run("missing_key_and_no_freshness_record_fails_closed", func(t *testing.T) {
		_, err := trustcache.NewReader(fakeGetter{}, clock).AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
		qt.Check(t, qt.IsTrue(errors.Is(err, trustcache.ErrCacheExpired)))
	})

	// missing_key_and_stale_freshness: fail-closed negative — a freshness
	// record exists but reports the upstream source as stale, so the
	// missing territory key must still be treated as genuine degradation.
	t.Run("missing_key_and_stale_freshness_fails_closed", func(t *testing.T) {
		fr := trustcache.TypeFreshness{SnapshotID: "snap-9", FetchedAt: now.Add(-5 * time.Minute), ValidUntil: now.Add(40 * time.Minute), UpstreamStale: true}
		g := fakeGetter{trustcache.FreshnessKey(trustcache.TypePIDProvider): entryJSON(t, fr)}
		_, err := trustcache.NewReader(g, clock).AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
		qt.Check(t, qt.IsTrue(errors.Is(err, trustcache.ErrCacheExpired)))
	})

	// missing_key_and_expired_freshness: fail-closed negative — the
	// freshness record's own ValidUntil has already passed, so the missing
	// territory key must still fail closed, not be read as confirmed-empty.
	t.Run("missing_key_and_expired_freshness_fails_closed", func(t *testing.T) {
		fr := trustcache.TypeFreshness{SnapshotID: "snap-9", FetchedAt: now.Add(-time.Hour), ValidUntil: now.Add(-time.Minute)}
		g := fakeGetter{trustcache.FreshnessKey(trustcache.TypePIDProvider): entryJSON(t, fr)}
		_, err := trustcache.NewReader(g, clock).AnchorSet(ctx, trustcache.TypePIDProvider, "LV")
		qt.Check(t, qt.IsTrue(errors.Is(err, trustcache.ErrCacheExpired)))
	})
}

func TestReaderFreshnessAndSnapshotHealth(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	ctx := context.Background()

	fr := trustcache.TypeFreshness{SnapshotID: "snap-1", FetchedAt: now, ValidUntil: now.Add(45 * time.Minute)}
	frRaw, _ := json.Marshal(fr)
	sh := trustcache.SnapshotHealth{SnapshotID: "snap-1", LOTLSequence: 388, CheckedAt: now, PendingBootstrap: true,
		Territories: map[string]trustcache.TerritoryHealth{"LV": {TLSequence: 51, Stale: false}}}
	shRaw, _ := json.Marshal(sh)

	g := fakeGetter{
		trustcache.FreshnessKey(trustcache.TypePIDProvider): frRaw,
		trustcache.SnapshotHealthKey:                        shRaw,
	}
	r := trustcache.NewReader(g, clock)

	gotFr, err := r.Freshness(ctx, trustcache.TypePIDProvider)
	qt.Assert(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(gotFr.SnapshotID, "snap-1"))

	gotSh, err := r.SnapshotHealth(ctx)
	qt.Assert(t, qt.IsNil(err))
	qt.Check(t, qt.Equals(gotSh.LOTLSequence, uint64(388)))
	qt.Check(t, qt.IsTrue(gotSh.PendingBootstrap))

	_, err = r.Freshness(ctx, trustcache.TypeAccessCA)
	qt.Check(t, qt.IsTrue(errors.Is(err, trustcache.ErrCacheExpired)))
}

// TestAnchorUseCasesRoundTrip guards that Anchor.UseCases survives a JSON
// round-trip through the Reader, rather than silently zeroing out on every
// cache read.
func TestAnchorUseCasesRoundTrip(t *testing.T) {
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	clock := func() time.Time { return now }
	ctx := context.Background()

	entry := trustcache.AnchorSetEntry{
		SnapshotID: "snap-1",
		Type:       trustcache.TypeEAAProvider,
		Territory:  "LV",
		FetchedAt:  now,
		ValidUntil: now.Add(time.Hour),
		Anchors: []trustcache.Anchor{{
			CertDER:    []byte{0x30, 0x82},
			Territory:  "LV",
			Status:     "granted",
			ValidUntil: now.Add(24 * time.Hour),
			TLSequence: 51,
			UseCases:   []string{"mDL", "EHIC"},
		}},
	}
	g := fakeGetter{trustcache.AnchorSetKey(trustcache.TypeEAAProvider, "LV"): entryJSON(t, entry)}

	got, err := trustcache.NewReader(g, clock).AnchorSet(ctx, trustcache.TypeEAAProvider, "LV")
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.Equals(len(got.Anchors), 1))
	qt.Check(t, qt.DeepEquals(got.Anchors[0].UseCases, []string{"mDL", "EHIC"}))
}
