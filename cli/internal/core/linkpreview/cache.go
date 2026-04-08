package linkpreview

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// ErrCachedFailure is returned by Cache.Get when the URL was previously fetched
// and failed. This distinguishes a negative cache hit from a cache miss (which returns nil, nil).
var ErrCachedFailure = fmt.Errorf("cached failure")

const (
	// CacheBucketName is the name of the KV bucket for link preview cache.
	CacheBucketName = "LINK_PREVIEW_CACHE"

	// SuccessTTL is how long successful previews are cached.
	SuccessTTL = 24 * time.Hour

	// FailureTTL is how long failed previews are cached.
	FailureTTL = 1 * time.Hour

	// BucketTTL is the bucket-level TTL (entries auto-expire).
	BucketTTL = 48 * time.Hour
)

// Cache provides caching for link preview results.
type Cache struct {
	kv jetstream.KeyValue
}

// NewCache creates or opens the link preview cache KV bucket.
func NewCache(ctx context.Context, js jetstream.JetStream, replicas int) (*Cache, error) {
	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      CacheBucketName,
		Description: "Cached link preview metadata",
		Storage:     jetstream.FileStorage,
		TTL:         BucketTTL,
		Replicas:    replicas,
	})
	if err != nil {
		return nil, err
	}
	return &Cache{kv: kv}, nil
}

// cacheKey generates a cache key from a URL.
func cacheKey(rawURL string) string {
	normalized := NormalizeURLString(rawURL)
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// Get retrieves a cached link preview.
// Returns nil, nil if not found or stale.
func (c *Cache) Get(ctx context.Context, url string) (*corev1.LinkPreview, error) {
	key := cacheKey(url)

	entry, err := c.kv.Get(ctx, key)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var cached corev1.CachedLinkPreview
	if err := proto.Unmarshal(entry.Value(), &cached); err != nil {
		return nil, err
	}

	// Check staleness
	fetchedAt := time.Unix(cached.FetchedAtUnix, 0)
	maxAge := SuccessTTL
	if cached.FetchFailed {
		maxAge = FailureTTL
	}

	if time.Since(fetchedAt) > maxAge {
		return nil, nil // Stale entry
	}

	// Signal negative cache hit so callers can distinguish from a cache miss
	if cached.FetchFailed {
		return nil, ErrCachedFailure
	}

	return cached.Preview, nil
}

// Set stores a link preview in the cache.
func (c *Cache) Set(ctx context.Context, url string, preview *corev1.LinkPreview) error {
	cached := &corev1.CachedLinkPreview{
		Url:           url,
		Preview:       preview,
		FetchFailed:   false,
		FetchedAtUnix: time.Now().Unix(),
	}

	data, err := proto.Marshal(cached)
	if err != nil {
		return err
	}

	_, err = c.kv.Put(ctx, cacheKey(url), data)
	return err
}

// SetFailure stores a failed fetch in the cache (negative caching).
func (c *Cache) SetFailure(ctx context.Context, url string, reason string) error {
	cached := &corev1.CachedLinkPreview{
		Url:           url,
		Preview:       nil,
		FetchFailed:   true,
		ErrorReason:   reason,
		FetchedAtUnix: time.Now().Unix(),
	}

	data, err := proto.Marshal(cached)
	if err != nil {
		return err
	}

	_, err = c.kv.Put(ctx, cacheKey(url), data)
	return err
}
