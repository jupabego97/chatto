package core

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strings"
	"testing"
)

// containsString is a helper to check if a string contains a substring.
func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ============================================================================
// Instance OG Image Tests
// ============================================================================

// createOGTestImage creates a simple PNG image for testing OG image uploads.
func createOGTestImage(width, height int) *bytes.Reader {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a solid color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return bytes.NewReader(buf.Bytes())
}

func TestChattoCore_UploadInstanceOGImage(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("upload valid image", func(t *testing.T) {
		imageReader := createOGTestImage(1200, 630)

		asset, err := core.UploadInstanceOGImage(ctx, imageReader)
		if err != nil {
			t.Fatalf("Failed to upload OG image: %v", err)
		}

		if asset == nil {
			t.Fatal("Expected asset, got nil")
		}

		// Verify asset has a valid key
		if asset.GetNats() != nil {
			if asset.GetNats().Key == "" {
				t.Error("Expected non-empty asset key")
			}
		} else {
			t.Error("Expected NATS asset (S3 not configured in tests)")
		}
	})

	t.Run("upload small image", func(t *testing.T) {
		// Small images should be processed and resized
		imageReader := createOGTestImage(100, 100)

		asset, err := core.UploadInstanceOGImage(ctx, imageReader)
		if err != nil {
			t.Fatalf("Failed to upload small OG image: %v", err)
		}

		if asset == nil {
			t.Fatal("Expected asset, got nil")
		}
	})

	t.Run("upload invalid data fails", func(t *testing.T) {
		invalidData := bytes.NewReader([]byte("not an image"))

		_, err := core.UploadInstanceOGImage(ctx, invalidData)
		if err == nil {
			t.Error("Expected error for invalid image data")
		}
	})
}

func TestChattoCore_SetAndGetInstanceOGImage(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("set and get OG image", func(t *testing.T) {
		// First upload an image to get an asset
		imageReader := createOGTestImage(1200, 630)
		asset, err := core.UploadInstanceOGImage(ctx, imageReader)
		if err != nil {
			t.Fatalf("Failed to upload OG image: %v", err)
		}

		// Set the OG image
		err = core.SetInstanceOGImage(ctx, "test-admin", asset)
		if err != nil {
			t.Fatalf("Failed to set OG image: %v", err)
		}

		// Get the OG image back
		retrievedAsset, err := core.GetInstanceOGImage(ctx)
		if err != nil {
			t.Fatalf("Failed to get OG image: %v", err)
		}

		if retrievedAsset == nil {
			t.Fatal("Expected asset, got nil")
		}

		// Verify the asset key matches
		if retrievedAsset.GetNats().Key != asset.GetNats().Key {
			t.Errorf("Asset key mismatch: expected %s, got %s",
				asset.GetNats().Key, retrievedAsset.GetNats().Key)
		}
	})

	t.Run("get OG image when none set returns nil", func(t *testing.T) {
		// Use a fresh core instance
		freshCore, _ := setupTestCore(t)
		freshCtx := testContext(t)

		asset, err := freshCore.GetInstanceOGImage(freshCtx)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if asset != nil {
			t.Error("Expected nil asset when none set")
		}
	})
}

func TestChattoCore_GetInstanceOGImageURL(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("returns empty string when no image set", func(t *testing.T) {
		url, err := core.GetInstanceOGImageURL(ctx, nil, nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if url != "" {
			t.Errorf("Expected empty URL, got %s", url)
		}
	})

	t.Run("returns direct URL when no dimensions specified", func(t *testing.T) {
		// Upload and set an image
		imageReader := createOGTestImage(1200, 630)
		asset, err := core.UploadInstanceOGImage(ctx, imageReader)
		if err != nil {
			t.Fatalf("Failed to upload OG image: %v", err)
		}

		err = core.SetInstanceOGImage(ctx, "test-admin", asset)
		if err != nil {
			t.Fatalf("Failed to set OG image: %v", err)
		}

		// Get URL without dimensions
		url, err := core.GetInstanceOGImageURL(ctx, nil, nil)
		if err != nil {
			t.Fatalf("Failed to get OG image URL: %v", err)
		}

		if url == "" {
			t.Error("Expected non-empty URL")
		}

		// Should be a direct asset URL
		expectedPrefix := "/assets/instance/"
		if len(url) < len(expectedPrefix) || url[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("Expected URL to start with %s, got %s", expectedPrefix, url)
		}
	})

	t.Run("returns transform URL when dimensions specified", func(t *testing.T) {
		// Upload and set an image
		imageReader := createOGTestImage(1200, 630)
		asset, err := core.UploadInstanceOGImage(ctx, imageReader)
		if err != nil {
			t.Fatalf("Failed to upload OG image: %v", err)
		}

		err = core.SetInstanceOGImage(ctx, "test-admin", asset)
		if err != nil {
			t.Fatalf("Failed to set OG image: %v", err)
		}

		// Get URL with dimensions (OG standard size)
		width, height := 1200, 630
		url, err := core.GetInstanceOGImageURL(ctx, &width, &height)
		if err != nil {
			t.Fatalf("Failed to get OG image URL: %v", err)
		}

		if url == "" {
			t.Error("Expected non-empty URL")
		}

		// Should be a transform URL (contains /t/ with transform parameters)
		expectedContains := "/t/"
		if !containsString(url, expectedContains) {
			t.Errorf("Expected URL to contain %s, got %s", expectedContains, url)
		}
	})
}

func TestChattoCore_DeleteInstanceOGImage(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("delete existing OG image", func(t *testing.T) {
		// Upload and set an image
		imageReader := createOGTestImage(1200, 630)
		asset, err := core.UploadInstanceOGImage(ctx, imageReader)
		if err != nil {
			t.Fatalf("Failed to upload OG image: %v", err)
		}

		err = core.SetInstanceOGImage(ctx, "test-admin", asset)
		if err != nil {
			t.Fatalf("Failed to set OG image: %v", err)
		}

		// Verify it exists
		existing, err := core.GetInstanceOGImage(ctx)
		if err != nil {
			t.Fatalf("Failed to get OG image: %v", err)
		}
		if existing == nil {
			t.Fatal("Expected OG image to be set")
		}

		// Delete it
		err = core.DeleteInstanceOGImage(ctx, "test-admin")
		if err != nil {
			t.Fatalf("Failed to delete OG image: %v", err)
		}

		// Verify it's gone
		deleted, err := core.GetInstanceOGImage(ctx)
		if err != nil {
			t.Fatalf("Failed to get OG image after delete: %v", err)
		}
		if deleted != nil {
			t.Error("Expected OG image to be nil after delete")
		}
	})

	t.Run("delete non-existent OG image is no-op", func(t *testing.T) {
		// Use a fresh core instance
		freshCore, _ := setupTestCore(t)
		freshCtx := testContext(t)

		// Should not error when nothing to delete
		err := freshCore.DeleteInstanceOGImage(freshCtx, "test-admin")
		if err != nil {
			t.Errorf("Unexpected error deleting non-existent image: %v", err)
		}
	})
}

func TestChattoCore_ReplaceInstanceOGImage(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("replace existing OG image", func(t *testing.T) {
		// Upload and set first image
		image1 := createOGTestImage(1200, 630)
		asset1, err := core.UploadInstanceOGImage(ctx, image1)
		if err != nil {
			t.Fatalf("Failed to upload first OG image: %v", err)
		}

		err = core.SetInstanceOGImage(ctx, "test-admin", asset1)
		if err != nil {
			t.Fatalf("Failed to set first OG image: %v", err)
		}

		// Get the first asset key
		firstKey := asset1.GetNats().Key

		// Upload and set second image (replaces first)
		image2 := createOGTestImage(800, 600)
		asset2, err := core.UploadInstanceOGImage(ctx, image2)
		if err != nil {
			t.Fatalf("Failed to upload second OG image: %v", err)
		}

		err = core.SetInstanceOGImage(ctx, "test-admin", asset2)
		if err != nil {
			t.Fatalf("Failed to set second OG image: %v", err)
		}

		// Verify the new image is set
		current, err := core.GetInstanceOGImage(ctx)
		if err != nil {
			t.Fatalf("Failed to get OG image: %v", err)
		}

		if current.GetNats().Key == firstKey {
			t.Error("Expected new asset key, got the old one")
		}

		if current.GetNats().Key != asset2.GetNats().Key {
			t.Errorf("Asset key mismatch: expected %s, got %s",
				asset2.GetNats().Key, current.GetNats().Key)
		}
	})
}
