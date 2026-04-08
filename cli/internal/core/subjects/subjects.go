package subjects

import "fmt"

// This file provides a single source of truth for all NATS subject patterns in the system.
// All functions are pure - they construct subjects from entity IDs.
//
// Simplified Subject Patterns (optimized for low cardinality):
//   - Instance: instance.user.{userId}.{eventType}
//   - Space:    space.{spaceId}.{eventType}
//   - Room:     space.{spaceId}.room.{roomId}.{eventType}
//   - User:     user.{userId}.event (transient, NATS Core only)
//
// Note: Actor/user context is stored in event payloads, not subjects.
// This minimizes subject cardinality for optimal NATS memory usage.

// ===== SPACE STREAM SUBJECTS (space.{spaceId}.>) =====

// SpaceEvent returns the subject for space-level events in a SPACE stream.
// Pattern: space.{spaceId}.{eventType}
// Examples: joined, left
// Note: Actor/user information is in the event payload, not the subject
func SpaceEvent(spaceID, eventType string) string {
	return fmt.Sprintf("space.%s.%s", spaceID, eventType)
}

// SpaceAllEvents returns the wildcard subject for all space-level events.
// Pattern: space.{spaceId}.>
func SpaceAllEvents(spaceID string) string {
	return fmt.Sprintf("space.%s.>", spaceID)
}

// ===== ROOM EVENT SUBJECTS =====
// Room events are stored in the SPACE stream, not separate room streams.
// This simplifies consumption - one subscription gets all space events.
//
// Subject structure (event IDs enable O(1) lookup via GetLastMsgForSubject):
//   - Messages: space.{spaceId}.room.{roomId}.msg.{eventId}                             (root messages)
//   - Threads:  space.{spaceId}.room.{roomId}.msg.{rootEventId}.replies.{eventId}       (thread replies)
//   - Meta:     space.{spaceId}.room.{roomId}.meta                                      (lifecycle + membership)
//
// Filtering patterns:
//   - All messages:    msg.>
//   - Root only:       msg.*
//   - All threads:     msg.*.replies.>
//   - Specific thread: msg.{rootEventId}.replies.>

// SpaceRoomMessage returns the subject for root message events in a room.
// Pattern: space.{spaceId}.room.{roomId}.msg.{eventId}
// Used for: Posting new top-level messages (not thread replies)
// The eventId enables O(1) lookup via GetLastMsgForSubject.
func SpaceRoomMessage(spaceID, roomID, eventID string) string {
	return fmt.Sprintf("space.%s.room.%s.msg.%s", spaceID, roomID, eventID)
}

// SpaceRoomThread returns the subject for a thread reply message.
// Pattern: space.{spaceId}.room.{roomId}.msg.{rootEventId}.replies.{eventId}
// Used for: Posting replies to a specific thread
// The eventId enables O(1) lookup via GetLastMsgForSubject.
func SpaceRoomThread(spaceID, roomID, rootEventID, eventID string) string {
	return fmt.Sprintf("space.%s.room.%s.msg.%s.replies.%s", spaceID, roomID, rootEventID, eventID)
}

// SpaceRoomThreadFilter returns the wildcard subject for all replies in a specific thread.
// Pattern: space.{spaceId}.room.{roomId}.msg.{rootEventId}.replies.>
// Used for: Consuming all thread replies regardless of their event IDs.
func SpaceRoomThreadFilter(spaceID, roomID, rootEventID string) string {
	return fmt.Sprintf("space.%s.room.%s.msg.%s.replies.>", spaceID, roomID, rootEventID)
}

// SpaceRoomThreadLookup returns the wildcard subject for looking up a thread reply by event ID.
// Pattern: space.{spaceId}.room.{roomId}.msg.*.replies.{eventId}
// Used for: O(1) lookup of any thread reply by event ID via GetLastMsgForSubject.
func SpaceRoomThreadLookup(spaceID, roomID, eventID string) string {
	return fmt.Sprintf("space.%s.room.%s.msg.*.replies.%s", spaceID, roomID, eventID)
}

// SpaceRoomAllThreads returns the wildcard subject for all thread events in a room.
// Pattern: space.{spaceId}.room.{roomId}.msg.*.replies.>
// Used for: Consuming all thread activity in a room
func SpaceRoomAllThreads(spaceID, roomID string) string {
	return fmt.Sprintf("space.%s.room.%s.msg.*.replies.>", spaceID, roomID)
}

// SpaceRoomMeta returns the subject for non-message room events.
// Pattern: space.{spaceId}.room.{roomId}.meta
// Used for: Room lifecycle (created, updated, deleted) and membership (joined, left)
// Event type is determined by the event payload, not the subject.
func SpaceRoomMeta(spaceID, roomID string) string {
	return fmt.Sprintf("space.%s.room.%s.meta", spaceID, roomID)
}

// SpaceRoomAllMessages returns the wildcard subject for all messages (root + thread) in a room.
// Pattern: space.{spaceId}.room.{roomId}.msg.>
// Used for: Deriving room's last message timestamp from JetStream (includes thread activity).
func SpaceRoomAllMessages(spaceID, roomID string) string {
	return fmt.Sprintf("space.%s.room.%s.msg.>", spaceID, roomID)
}

// SpaceRoomRootMessages returns the wildcard subject for root messages only in a room.
// Pattern: space.{spaceId}.room.{roomId}.msg.*
// The single wildcard (*) matches one token (eventId for root messages) but excludes
// thread replies which have subjects like msg.{rootId}.replies.{eventId} (3 tokens).
// Used for: Deriving room's last root message sequence for unread tracking.
func SpaceRoomRootMessages(spaceID, roomID string) string {
	return fmt.Sprintf("space.%s.room.%s.msg.*", spaceID, roomID)
}

// SpaceRoomAllEvents returns the filter subject for all events in a specific room.
// Pattern: space.{spaceId}.room.{roomId}.>
// Matches all room events (messages, threads, meta) since they all have suffixes.
func SpaceRoomAllEvents(spaceID, roomID string) string {
	return fmt.Sprintf("space.%s.room.%s.>", spaceID, roomID)
}

// SpaceAllRoomEvents returns the wildcard subject for all room events across a space.
// Pattern: space.{spaceId}.room.>
// Useful for indexers, notification services, etc. that need all room activity.
func SpaceAllRoomEvents(spaceID string) string {
	return fmt.Sprintf("space.%s.room.>", spaceID)
}

// SpaceRoomRootEventsFilters returns filter subjects for root messages and meta events in a room.
// Excludes thread replies (which have subjects like msg.{rootId}.replies.{eventId}).
// Returns: ["space.{s}.room.{r}.msg.*", "space.{s}.room.{r}.meta"]
// Use with JetStream consumer FilterSubjects for efficient server-side filtering.
func SpaceRoomRootEventsFilters(spaceID, roomID string) []string {
	return []string{
		fmt.Sprintf("space.%s.room.%s.msg.*", spaceID, roomID),
		fmt.Sprintf("space.%s.room.%s.meta", spaceID, roomID),
	}
}

// SpaceAllRoomEventsFilters returns filter subjects for ALL messages (root + thread replies)
// and meta events across all rooms in a space.
// Returns: ["space.{s}.room.*.msg.>", "space.{s}.room.*.meta"]
// Use with JetStream consumer FilterSubjects for live subscriptions that need all messages.
func SpaceAllRoomEventsFilters(spaceID string) []string {
	return []string{
		fmt.Sprintf("space.%s.room.*.msg.>", spaceID),
		fmt.Sprintf("space.%s.room.*.meta", spaceID),
	}
}

// ParseRoomIDFromSubject extracts the room ID from a space event subject.
// Returns the room ID if the subject is a room event, or empty string if it's a space-level event.
// Handles all room subject patterns:
//   - space.{spaceId}.room.{roomId}.msg.{eventId}                        (root messages)
//   - space.{spaceId}.room.{roomId}.meta                                 (lifecycle/membership)
//   - space.{spaceId}.room.{roomId}.msg.{rootEventId}.replies.{eventId}  (thread replies)
func ParseRoomIDFromSubject(subject string) string {
	parts := splitSubject(subject)
	// Minimum 5 parts: space.{s}.room.{r}.meta (or .msg.{id} for messages)
	if len(parts) >= 5 && parts[0] == "space" && parts[2] == "room" {
		return parts[3] // roomId is always at index 3
	}
	return ""
}

// ParseThreadRootEventIDFromSubject extracts the root event ID from a thread message subject.
// Returns (rootEventId, true) if the subject is a thread reply, or ("", false) if it's a root message or non-message event.
// Pattern: space.{spaceId}.room.{roomId}.msg.{rootEventId}.replies.{eventId}
func ParseThreadRootEventIDFromSubject(subject string) (string, bool) {
	parts := splitSubject(subject)
	// Thread subjects have 8 parts: space.{s}.room.{r}.msg.{rootEventId}.replies.{eventId}
	if len(parts) == 8 && parts[0] == "space" && parts[2] == "room" && parts[4] == "msg" && parts[6] == "replies" {
		return parts[5], true
	}
	return "", false
}

// IsRootMessageSubject checks if a subject is for a top-level (root) message.
// Pattern: space.{spaceId}.room.{roomId}.msg.{eventId} (6 parts with .msg. segment)
func IsRootMessageSubject(subject string) bool {
	parts := splitSubject(subject)
	return len(parts) == 6 && parts[0] == "space" && parts[2] == "room" && parts[4] == "msg"
}

// IsMetaSubject checks if a subject is for a meta event (lifecycle/membership).
// Pattern: space.{spaceId}.room.{roomId}.meta
func IsMetaSubject(subject string) bool {
	parts := splitSubject(subject)
	return len(parts) == 5 && parts[0] == "space" && parts[2] == "room" && parts[4] == "meta"
}

// IsThreadSubject checks if a subject is for a thread reply message.
// Pattern: space.{spaceId}.room.{roomId}.msg.{rootEventId}.replies.{eventId}
func IsThreadSubject(subject string) bool {
	parts := splitSubject(subject)
	return len(parts) == 8 && parts[0] == "space" && parts[2] == "room" && parts[4] == "msg" && parts[6] == "replies"
}

// ParseEventIDFromSubject extracts the event ID from a message subject.
// Returns the event ID if the subject is a message event (root or thread), or empty string otherwise.
// Patterns:
//   - space.{spaceId}.room.{roomId}.msg.{eventId}                        -> eventId at index 5
//   - space.{spaceId}.room.{roomId}.msg.{rootEventId}.replies.{eventId}  -> eventId at index 7
func ParseEventIDFromSubject(subject string) string {
	parts := splitSubject(subject)
	if len(parts) < 5 || parts[0] != "space" || parts[2] != "room" {
		return ""
	}
	// Root message: space.{s}.room.{r}.msg.{eventId} (6 parts)
	if len(parts) == 6 && parts[4] == "msg" {
		return parts[5]
	}
	// Thread reply: space.{s}.room.{r}.msg.{rootId}.replies.{eventId} (8 parts)
	if len(parts) == 8 && parts[4] == "msg" && parts[6] == "replies" {
		return parts[7]
	}
	return ""
}

// splitSubject splits a NATS subject by dots.
func splitSubject(subject string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(subject); i++ {
		if subject[i] == '.' {
			parts = append(parts, subject[start:i])
			start = i + 1
		}
	}
	if start < len(subject) {
		parts = append(parts, subject[start:])
	}
	return parts
}

// ===== LIVE SUBJECT PATTERNS (live.>) =====
// Live subjects are used for transient events that bypass JetStream storage.
// Events are published directly via publishLiveSpaceEvent()/publishInstanceEvent() for real-time
// notifications (reactions, message updates/deletes, space member removal, etc.).
//
// Space live events use publishLiveSpaceEvent(), instance live events use publishInstanceEvent().

// LiveInstanceAllEvents returns the live subject for all instance-level events.
// Pattern: live.instance.>
// Used for: Server-side subscription to all instance events with authorization filtering.
func LiveInstanceAllEvents() string {
	return "live.instance.>"
}

// LiveInstanceUserAllEvents returns the live subject for all events for a specific user.
// Pattern: live.instance.user.{userId}.>
// Used for: Real-time instance notifications (space created/updated/deleted, profile changes)
func LiveInstanceUserAllEvents(userID string) string {
	return fmt.Sprintf("live.instance.user.%s.>", userID)
}

// LiveInstanceUserEvent returns the live subject for a specific user's instance event.
// Pattern: live.instance.user.{userId}.{eventType}
// Used for: Transient events that bypass JetStream storage entirely (e.g., joined_space, left_space)
func LiveInstanceUserEvent(userID, eventType string) string {
	return fmt.Sprintf("live.instance.user.%s.%s", userID, eventType)
}

// LiveSpaceAllEvents returns the live subject for all space-level events.
// Pattern: live.space.{spaceId}.>
// Used for: Real-time space notifications (includes room events)
func LiveSpaceAllEvents(spaceID string) string {
	return fmt.Sprintf("live.space.%s.>", spaceID)
}

// LiveSpaceLevelEvents returns the live subject for space-level events only (not room events).
// Pattern: live.space.{spaceId}.*
// Uses single wildcard (*) to match only direct children like joined/left, excluding room.> subjects.
// Used for: Real-time space membership notifications without room events
func LiveSpaceLevelEvents(spaceID string) string {
	return fmt.Sprintf("live.space.%s.*", spaceID)
}

// LiveSpaceEvent returns the live subject for a space-level event (direct publish, bypasses JetStream).
// Pattern: live.space.{spaceId}.{eventType}
// Used for: Transient space-level events that don't need storage (e.g., member_deleted)
func LiveSpaceEvent(spaceID, eventType string) string {
	return fmt.Sprintf("live.space.%s.%s", spaceID, eventType)
}

// LiveSpaceRoomEvent returns the live subject for a room event (direct publish, bypasses JetStream).
// Pattern: live.space.{spaceId}.room.{roomId}.{eventType}
// Used for: Transient room events that don't need storage (e.g., reactions)
func LiveSpaceRoomEvent(spaceID, roomID, eventType string) string {
	return fmt.Sprintf("live.space.%s.room.%s.%s", spaceID, roomID, eventType)
}

// LiveSpaceRoomAllEvents returns the live subject for all transient room events in a space.
// Pattern: live.space.{spaceId}.room.>
// Used for: Subscribing to all live-only room events (reactions, typing indicators, etc.)
func LiveSpaceRoomAllEvents(spaceID string) string {
	return fmt.Sprintf("live.space.%s.room.>", spaceID)
}

// LiveSpaceRoomReactionEvents returns the subscription subject for all live-only reaction events.
// Pattern: live.space.{spaceId}.room.*.reaction_*
// Used for: Subscribing to transient reaction events that bypass JetStream storage
func LiveSpaceRoomReactionEvents(spaceID string) string {
	return fmt.Sprintf("live.space.%s.room.*.reaction_*", spaceID)
}

// ===== INSTANCE LIVE SUBJECT PATTERNS =====
// For transient instance-level events that bypass JetStream (config changes, etc.)

// LiveInstanceConfigUpdated returns the subject for instance config update events.
// Pattern: live.instance.config.updated
// Used for: Broadcasting config changes to all connected clients
func LiveInstanceConfigUpdated() string {
	return "live.instance.config.updated"
}

// LiveInstanceConfigAllEvents returns the wildcard subject for all instance config events.
// Pattern: live.instance.config.>
// Used for: Subscribing to all instance config-related live events
func LiveInstanceConfigAllEvents() string {
	return "live.instance.config.>"
}

// LiveInstanceSpaceEvent returns the live subject for a space-wide instance event.
// Pattern: live.instance.space.{spaceId}.{eventType}
// Used for: Transient events broadcast to all space members (e.g., new_message_in_space)
// These events bypass JetStream and are delivered via server-side filtering.
func LiveInstanceSpaceEvent(spaceID, eventType string) string {
	return fmt.Sprintf("live.instance.space.%s.%s", spaceID, eventType)
}

