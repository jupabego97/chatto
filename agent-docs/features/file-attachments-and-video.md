# File Attachments & Video Processing

## Attachments

- Users can upload files as part of a message (images, videos, documents).
- The composer supports drag-and-drop, paste, and file picker for adding attachments.
- Draft attachments survive component remounts during room switches (module-level state).
- Default size limit is 25 MB for general files, 100 MB for videos when video processing is enabled.

## Image Handling

- Image dimensions are extracted at upload time.
- Images can be dynamically resized via signed URL parameters (width, height, fit mode).
- An optional image cache stores resized images as WebP with auto-expiry.
- Animated GIFs are routed to the video pipeline instead of being treated as images.

## Video Processing

- Videos (and animated GIFs) are processed asynchronously into H.264 MP4 quality variants.
- Processing status is tracked as: PENDING → PROCESSING → COMPLETED (or FAILED).
- Quality variants are selected based on source resolution (e.g., a 1080p source gets 720p and 480p variants).
- After transcoding, the original file is replaced by the variants.
- A thumbnail is generated from an early frame.
- A completion event is published so the frontend can refresh the message and show the video player.
- Processing failures are non-fatal — the message is posted regardless, and the UI shows an error state.

## Storage

- Two storage backends are supported: NATS ObjectStore (default, good for development) and S3-compatible storage (production).
- The system tries both backends when retrieving files for backward compatibility.

## Permissions

- No separate upload permission exists. Uploading requires room membership and the relevant message posting permission (`message.post` or `message.post-in-thread`).
