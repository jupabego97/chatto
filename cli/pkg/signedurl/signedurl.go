// Package signedurl provides HMAC-signed URL path generation and verification.
// It creates tamper-proof URL path components by signing parameters with
// HMAC-SHA256, and verifies signatures on the way back in.
package signedurl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// TransformParams holds the parameters for an image transformation.
type TransformParams struct {
	Width  int    `json:"w"`
	Height int    `json:"h"`
	Fit    string `json:"f"`
}

// SignedTransformPath generates a signed path component for an image transformation URL.
// Returns a string in the format: {base64params}.{signature}
// where base64params is base64url-encoded JSON: {"w":width,"h":height,"f":"fit"}
// and signature is a truncated HMAC-SHA256 of {resourceID1}/{resourceID2}/{base64params}
//
// The resourceID1 and resourceID2 parameters are opaque strings that identify the resource.
// This function has no knowledge of what they represent.
func SignedTransformPath(secret, resourceID1, resourceID2 string, width, height int, fit string) string {
	// Encode params as JSON then base64url
	params := TransformParams{Width: width, Height: height, Fit: fit}
	paramsJSON, _ := json.Marshal(params)
	paramsB64 := base64.RawURLEncoding.EncodeToString(paramsJSON)

	// Sign: {resourceID1}/{resourceID2}/{paramsB64}
	message := fmt.Sprintf("%s/%s/%s", resourceID1, resourceID2, paramsB64)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	// Use first 16 bytes (32 hex chars) for shorter URLs while still secure
	signature := hex.EncodeToString(h.Sum(nil)[:16])

	return paramsB64 + "." + signature
}

// ParseSignedTransformPath parses and verifies a signed transform path.
// Input format: {base64params}.{signature}
// Returns the transform params if valid, or an error if invalid.
//
// The resourceID1 and resourceID2 parameters are opaque strings that identify the resource.
// This function has no knowledge of what they represent.
func ParseSignedTransformPath(secret, resourceID1, resourceID2, signedPath string) (*TransformParams, error) {
	// Split into params and signature
	parts := strings.SplitN(signedPath, ".", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid signed path format")
	}
	paramsB64, signature := parts[0], parts[1]

	// Verify signature first (constant-time comparison)
	message := fmt.Sprintf("%s/%s/%s", resourceID1, resourceID2, paramsB64)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	expectedSig := hex.EncodeToString(h.Sum(nil)[:16])
	if !hmac.Equal([]byte(expectedSig), []byte(signature)) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode base64 params
	paramsJSON, err := base64.RawURLEncoding.DecodeString(paramsB64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 params: %w", err)
	}

	// Parse JSON
	var params TransformParams
	if err := json.Unmarshal(paramsJSON, &params); err != nil {
		return nil, fmt.Errorf("invalid params JSON: %w", err)
	}

	// Validate params
	if params.Width < 1 || params.Width > 2048 {
		return nil, fmt.Errorf("width out of range [1, 2048]: %d", params.Width)
	}
	if params.Height < 1 || params.Height > 2048 {
		return nil, fmt.Errorf("height out of range [1, 2048]: %d", params.Height)
	}
	if params.Fit != "contain" && params.Fit != "cover" && params.Fit != "exact" {
		return nil, fmt.Errorf("invalid fit mode: %s", params.Fit)
	}

	return &params, nil
}
