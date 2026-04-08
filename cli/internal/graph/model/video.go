package model

// VideoProcessing represents the processing state of a video attachment.
// SpaceID and ThumbnailAttachmentID are internal fields used by field resolvers.
type VideoProcessing struct {
	Status                VideoProcessingStatus
	DurationMs            *int64
	Width                 *int32
	Height                *int32
	ErrorMessage          *string
	Variants              []*VideoVariant
	SpaceID               string // internal: for building URLs
	ThumbnailAttachmentID string // internal: for thumbnailUrl resolver
}

// VideoVariant represents a transcoded quality variant of a video.
// SpaceID and AttachmentID are internal fields used by field resolvers.
type VideoVariant struct {
	Quality      string
	Width        int32
	Height       int32
	Size         int64
	SpaceID      string // internal: for building URL
	AttachmentID string // internal: for building URL
}
