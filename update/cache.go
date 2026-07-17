package update

import (
	"context"
	"encoding/json"
	"time"
)

// Cache is the storage backend a [CacheProvider] memoizes lookups through. It is
// satisfied by the application's cache root, which already scopes writes beneath
// an application-specific directory, so the provider addresses entries by a bare
// file name.
type Cache interface {
	// ReadFile reads the entire contents of the named cache entry.
	ReadFile(name string) ([]byte, error)

	// WriteFile writes data to the named cache entry, creating it if absent.
	WriteFile(name string, data []byte) error
}

// CacheProvider memoizes a [Provider] through a [Cache], so repeated lookups
// within TTL do not re-query the underlying channel.
//
// The cache and clock are collaborators supplied by the constructor. Only
// successful lookups are cached: an error from the wrapped provider is returned
// as-is and leaves any existing cache entry intact.
type CacheProvider struct {
	// Provider is the wrapped provider consulted on a cache miss.
	Provider Provider

	// Source is the stem of the cache file; it should match the source name the
	// provider is registered under.
	Source string

	// TTL is how long a cached entry is considered fresh.
	TTL time.Duration

	// Cache is the backend cached lookups are read from and written to.
	Cache Cache

	// Now reports the current time.
	Now func() time.Time
}

// cacheRecord is the on-disk representation of a cached lookup.
type cacheRecord struct {
	Version   string    `json:"version"`
	CheckedAt time.Time `json:"checked_at"`
}

// LatestVersion returns the cached version when a fresh entry exists, otherwise
// it consults the wrapped provider and caches a successful result. It returns
// the wrapped provider's error unchanged, without caching it.
func (c *CacheProvider) LatestVersion(ctx context.Context) (string, error) {
	if version, ok := c.cached(); ok {
		return version, nil
	}
	version, err := c.Provider.LatestVersion(ctx)
	if err != nil {
		return "", err
	}
	c.store(version)
	return version, nil
}

// cached returns the cached version and whether a fresh entry was found. A
// missing, unreadable, or expired entry reports false.
func (c *CacheProvider) cached() (string, bool) {
	data, err := c.Cache.ReadFile(c.path())
	if err != nil {
		return "", false
	}
	var record cacheRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return "", false
	}
	if c.Now().Sub(record.CheckedAt) >= c.TTL {
		return "", false
	}
	return record.Version, true
}

// store best-effort writes version to the cache. Write failures are ignored: a
// failed cache write must not fail a lookup that otherwise succeeded.
func (c *CacheProvider) store(version string) {
	data, _ := json.Marshal(cacheRecord{Version: version, CheckedAt: c.Now()})
	_ = c.Cache.WriteFile(c.path(), data)
}

// path returns the cache entry name for this provider's source.
func (c *CacheProvider) path() string {
	return c.Source + ".json"
}

var _ Provider = (*CacheProvider)(nil)
