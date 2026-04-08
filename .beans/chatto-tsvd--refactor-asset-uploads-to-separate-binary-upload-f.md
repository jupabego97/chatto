---
# chatto-tsvd
title: Refactor asset uploads to separate binary upload from metadata
status: draft
type: task
created_at: 2026-01-17T17:19:14Z
updated_at: 2026-01-17T17:19:14Z
---

## Problem

Currently, asset upload mutations (avatars, space logos, banners) handle both binary file upload and metadata operations in a single GraphQL mutation that bypasses the Core API layer:

```go
// Current flow in mutation.resolvers.go
func (r *mutationResolver) UploadMyAvatar(..., file graphql.Upload) {
    asset, _ := r.core.UploadUserAvatar(ctx, userID, file.File)  // Binary + processing
    r.core.SetUserAvatar(ctx, userID, asset)                     // Metadata
}
```

This creates inconsistency - most operations go through Core API (`r.svc.*`), but file uploads call `r.core.*` directly.

## Proposed Solution

Separate binary upload from metadata operations:

1. **Binary upload endpoint** - Client uploads file first, gets back an asset ID
   - Could be a dedicated REST endpoint or a generic `uploadAsset` GraphQL mutation
   - Returns asset ID (and maybe a presigned URL for the uploaded file)

2. **Metadata operations via Core API** - Set avatar/logo/banner using just the asset ID
   ```go
   // New Core API operations
   svc.Users().SetAvatar(ctx, actorID, assetID)
   svc.Users().DeleteAvatar(ctx, actorID)
   svc.Spaces().SetLogo(ctx, actorID, spaceID, assetID)
   // etc.
   ```

## Benefits

- Core API layer is used consistently for all business logic
- Binary upload is decoupled from business operations
- Easier to add upload progress, resumable uploads, etc. later
- GraphQL mutations become simpler (just pass asset ID, no multipart handling)

## Affected Operations

- `uploadMyAvatar` / `deleteMyAvatar`
- `uploadSpaceLogo` / `deleteSpaceLogo`  
- `uploadSpaceBanner` / `deleteSpaceBanner`

## Notes

- This is a refactoring task, not a bug fix
- Current implementation works correctly, just bypasses the architectural pattern
- Consider whether asset upload should validate ownership/permissions before accepting