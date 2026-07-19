package updatecheck_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"testing"
	"time"

	"github.com/bitwizeshift/go-cli/internal/updatecheck"
	"github.com/bitwizeshift/go-cli/update/updatetest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// testNow is the fixed clock used across the cache tests; the seeded cache
// timestamps below are expressed relative to it.
var testNow = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// cacheEntry is the name the source "github" memoizes its lookup under.
const cacheEntry = "github.json"

func TestCacheProvider_LatestVersion(t *testing.T) {
	t.Parallel()

	errLookup := errors.New("release lookup")
	errUnavailable := errors.New("cache unavailable")

	testCases := []struct {
		name     string
		cache    updatecheck.Cache
		provider updatecheck.Provider
		want     string
		wantErr  error
	}{
		{
			name:     "CacheMiss",
			cache:    fakeCache{},
			provider: updatetest.Provider("v1.0.0"),
			want:     "v1.0.0",
		}, {
			name:     "FreshHit",
			cache:    fakeCache{cacheEntry: []byte(`{"version":"v9.9.9","checked_at":"2025-12-31T23:30:00Z"}`)},
			provider: updatetest.Provider("v1.0.0"),
			want:     "v9.9.9",
		}, {
			name:     "StaleRefetch",
			cache:    fakeCache{cacheEntry: []byte(`{"version":"v9.9.9","checked_at":"2025-12-31T22:00:00Z"}`)},
			provider: updatetest.Provider("v1.0.0"),
			want:     "v1.0.0",
		}, {
			name:     "CorruptCache",
			cache:    fakeCache{cacheEntry: []byte("{")},
			provider: updatetest.Provider("v1.0.0"),
			want:     "v1.0.0",
		}, {
			name:     "CacheUnavailable",
			cache:    errCache{err: errUnavailable},
			provider: updatetest.Provider("v1.0.0"),
			want:     "v1.0.0",
		}, {
			name:     "ProviderError",
			cache:    fakeCache{},
			provider: updatetest.ErrProvider(errLookup),
			want:     "",
			wantErr:  errLookup,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			sut := &updatecheck.CacheProvider{
				Provider: tc.provider,
				Source:   "github",
				TTL:      time.Hour,
				Cache:    tc.cache,
				Now:      func() time.Time { return testNow },
			}
			ctx := context.Background()

			// Act
			version, err := sut.LatestVersion(ctx)

			// Assert
			if got, want := err, tc.wantErr; !cmp.Equal(got, want, cmpopts.EquateErrors()) {
				t.Fatalf("CacheProvider.LatestVersion() error = %v, want %v", got, want)
			}
			if got, want := version, tc.want; !cmp.Equal(got, want) {
				t.Errorf("CacheProvider.LatestVersion() = %q, want %q", got, want)
			}
		})
	}
}

func TestCacheProvider_LatestVersion_OnMiss_PersistsResult(t *testing.T) {
	t.Parallel()

	// Arrange
	cache := fakeCache{}
	sut := &updatecheck.CacheProvider{
		Provider: updatetest.Provider("v4.5.6"),
		Source:   "github",
		TTL:      time.Hour,
		Cache:    cache,
		Now:      func() time.Time { return testNow },
	}
	ctx := context.Background()

	// Act
	version, err := sut.LatestVersion(ctx)

	// Assert
	if err != nil {
		t.Fatalf("CacheProvider.LatestVersion() error = %v, want %v", err, error(nil))
	}
	if got, want := version, "v4.5.6"; !cmp.Equal(got, want) {
		t.Errorf("CacheProvider.LatestVersion() = %q, want %q", got, want)
	}
	record := decodeRecord(t, cache[cacheEntry])
	want := cacheJSON{Version: "v4.5.6", CheckedAt: testNow}
	if got, want := record, want; !cmp.Equal(got, want) {
		t.Errorf("cached record = %+v, want %+v", got, want)
	}
}

// fakeCache is an in-memory [updatecheck.Cache] keyed by entry name, standing in for
// the application's cache root so tests can seed entries and read them back.
type fakeCache map[string][]byte

func (c fakeCache) ReadFile(name string) ([]byte, error) {
	data, ok := c[name]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return data, nil
}

func (c fakeCache) WriteFile(name string, data []byte) error {
	c[name] = data
	return nil
}

var _ updatecheck.Cache = (fakeCache)(nil)

// errCache is an [updatecheck.Cache] whose reads and writes always fail, modelling a
// cache root that cannot be reached.
type errCache struct {
	err error
}

func (c errCache) ReadFile(string) ([]byte, error) { return nil, c.err }
func (c errCache) WriteFile(string, []byte) error  { return c.err }

var _ updatecheck.Cache = errCache{}

// cacheJSON mirrors the on-disk representation of a memoized lookup, so tests can
// assert the persisted record without reaching into unexported types.
type cacheJSON struct {
	Version   string    `json:"version"`
	CheckedAt time.Time `json:"checked_at"`
}

// decodeRecord decodes a persisted cache entry into its [cacheJSON] form.
func decodeRecord(t *testing.T, data []byte) cacheJSON {
	t.Helper()

	var record cacheJSON
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatalf("json.Unmarshal(record) error = %v, want %v", err, error(nil))
	}
	return record
}
