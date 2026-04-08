package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"

	"hmans.de/chatto/internal/assets"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// instanceOGImageKey is the KV key for the instance OpenGraph image asset reference.
const instanceOGImageKey = "config.og_image"

// UploadInstanceOGImage processes and uploads an OpenGraph image for the instance.
// The image is resized and converted to WebP format (same as space logos).
// Returns an Asset reference that can be passed to SetInstanceOGImage.
func (c *ChattoCore) UploadInstanceOGImage(ctx context.Context, reader io.Reader) (*corev1.Asset, error) {
	// Process image: resize and convert to WebP (same as space logos)
	webpReader, err := assets.ProcessLogoImageWithConfig(reader, c.AssetsConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to process OG image: %w", err)
	}

	// Read the processed image into bytes (needed for both NATS and S3)
	webpData, err := io.ReadAll(webpReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read processed OG image: %w", err)
	}

	// Upload to storage with unique asset ID
	assetID := NewAssetID()
	var asset *corev1.Asset

	if c.ShouldUseS3() {
		// Upload to S3
		s3Key := S3KeyInstanceAsset(assetID)
		_, err := c.s3Client.PutObjectFromBytes(ctx, s3Key, webpData, "image/webp")
		if err != nil {
			return nil, fmt.Errorf("failed to upload OG image to S3: %w", err)
		}
		asset = &corev1.Asset{
			Asset: &corev1.Asset_S3{
				S3: &corev1.S3Asset{
					Key:    assetID,
					Bucket: proto.String(c.s3Client.Bucket()),
				},
			},
		}
		c.logger.Info("Uploaded instance OG image to S3", "asset_id", assetID, "size", len(webpData))
	} else {
		// Upload to NATS ObjectStore
		headers := nats.Header{}
		headers.Set("Content-Type", "image/webp")
		meta := jetstream.ObjectMeta{
			Name:    assetID,
			Headers: headers,
		}
		info, err := c.storage.instanceStore.Put(ctx, meta, bytes.NewReader(webpData))
		if err != nil {
			return nil, fmt.Errorf("failed to upload OG image: %w", err)
		}
		asset = &corev1.Asset{
			Asset: &corev1.Asset_Nats{
				Nats: &corev1.NATSAsset{
					Key: assetID,
				},
			},
		}
		c.logger.Info("Uploaded instance OG image", "size", info.Size)
	}

	return asset, nil
}

// SetInstanceOGImage stores the instance OpenGraph image asset reference.
// Uses optimistic locking to prevent race conditions.
// If an old image exists, it will be cleaned up after the new one is saved.
func (c *ChattoCore) SetInstanceOGImage(ctx context.Context, actorID string, asset *corev1.Asset) error {
	const maxRetries = 5

	// Marshal the new asset
	assetData, err := proto.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal OG image asset: %w", err)
	}

	// Optimistic locking loop
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get current entry (if any) with its revision
		var revision uint64
		var oldAsset *corev1.Asset

		entry, err := c.storage.instanceKV.Get(ctx, instanceOGImageKey)
		if err == nil {
			// Key exists - get revision and unmarshal old asset for cleanup
			revision = entry.Revision()
			oldAsset = &corev1.Asset{}
			if unmarshalErr := proto.Unmarshal(entry.Value(), oldAsset); unmarshalErr != nil {
				c.logger.Warn("Failed to unmarshal old OG image asset", "error", unmarshalErr)
				oldAsset = nil
			}
		} else if !errors.Is(err, jetstream.ErrKeyNotFound) {
			return fmt.Errorf("failed to get current OG image: %w", err)
		}
		// If ErrKeyNotFound, revision stays 0 and oldAsset stays nil

		// Try atomic update
		var updateErr error
		if revision == 0 {
			// No existing key - use Create for atomic insert
			_, updateErr = c.storage.instanceKV.Create(ctx, instanceOGImageKey, assetData)
		} else {
			// Existing key - use Update with revision check
			_, updateErr = c.storage.instanceKV.Update(ctx, instanceOGImageKey, assetData, revision)
		}

		if updateErr == nil {
			// Success! Now clean up the old image
			if oldAsset != nil {
				c.deleteAsset(ctx, oldAsset, "og_image", "instance")
			}
			c.logger.Info("Set instance OG image", "actor_id", actorID)
			return nil
		}

		// Check if it's a conflict error (someone else updated)
		if errors.Is(updateErr, jetstream.ErrKeyExists) {
			c.logger.Debug("OG image update conflict, retrying", "attempt", attempt+1)
			continue
		}

		return fmt.Errorf("failed to set OG image: %w", updateErr)
	}

	return fmt.Errorf("failed to set OG image after %d retries due to concurrent updates", maxRetries)
}

// GetInstanceOGImage retrieves the instance OpenGraph image asset reference.
// Returns (nil, nil) if no OG image is set.
func (c *ChattoCore) GetInstanceOGImage(ctx context.Context) (*corev1.Asset, error) {
	entry, err := c.storage.instanceKV.Get(ctx, instanceOGImageKey)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return nil, nil // No OG image set is not an error
		}
		return nil, fmt.Errorf("failed to get OG image: %w", err)
	}

	asset := &corev1.Asset{}
	if err := proto.Unmarshal(entry.Value(), asset); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OG image asset: %w", err)
	}

	return asset, nil
}

// GetInstanceOGImageURL returns the URL for the instance OpenGraph image.
// If width and height are provided (non-nil), returns a URL to a resized version.
// Returns empty string if no OG image is set.
func (c *ChattoCore) GetInstanceOGImageURL(ctx context.Context, width, height *int) (string, error) {
	image, err := c.GetInstanceOGImage(ctx)
	if err != nil {
		return "", err
	}

	// No OG image set
	if image == nil {
		return "", nil
	}

	// Get the asset ID (same format for both NATS and S3)
	var assetID string
	switch asset := image.Asset.(type) {
	case *corev1.Asset_Nats:
		assetID = asset.Nats.Key
	case *corev1.Asset_S3:
		assetID = asset.S3.Key
	default:
		return "", fmt.Errorf("unknown asset type")
	}

	// Use transform URL for resizing, or direct URL for full size
	if width != nil && height != nil {
		return c.GetTransformedInstanceAssetURL(assetID, *width, *height, "cover"), nil
	}
	return c.assetURL(fmt.Sprintf("/assets/instance/%s", assetID)), nil
}

// DeleteInstanceOGImage removes the instance OpenGraph image.
// Uses optimistic locking to prevent race conditions.
func (c *ChattoCore) DeleteInstanceOGImage(ctx context.Context, actorID string) error {
	const maxRetries = 5

	// Optimistic locking loop
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get current entry with revision
		entry, err := c.storage.instanceKV.Get(ctx, instanceOGImageKey)
		if err != nil {
			if errors.Is(err, jetstream.ErrKeyNotFound) {
				// No image to delete
				return nil
			}
			return fmt.Errorf("failed to get current OG image: %w", err)
		}

		revision := entry.Revision()

		// Unmarshal the asset for cleanup
		image := &corev1.Asset{}
		if unmarshalErr := proto.Unmarshal(entry.Value(), image); unmarshalErr != nil {
			c.logger.Warn("Failed to unmarshal OG image asset for deletion", "error", unmarshalErr)
			// Continue with deletion anyway - the KV entry is corrupted
			image = nil
		}

		// Try to delete with revision check
		if err := c.storage.instanceKV.Delete(ctx, instanceOGImageKey, jetstream.LastRevision(revision)); err != nil {
			if errors.Is(err, jetstream.ErrKeyExists) {
				c.logger.Debug("OG image delete conflict, retrying", "attempt", attempt+1)
				continue
			}
			return fmt.Errorf("failed to delete OG image: %w", err)
		}

		// Success! Now clean up the asset from storage
		if image != nil {
			c.deleteAsset(ctx, image, "og_image", "instance")
		}

		c.logger.Info("Deleted instance OG image", "actor_id", actorID)
		return nil
	}

	return fmt.Errorf("failed to delete OG image after %d retries due to concurrent updates", maxRetries)
}
