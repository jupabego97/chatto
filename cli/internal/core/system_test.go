package core

import (
	"testing"
)

// ============================================================================
// System Info Tests
// ============================================================================

func TestChattoCore_GetSystemInfo(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("returns system info for fresh instance", func(t *testing.T) {
		info, err := core.GetSystemInfo(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if info == nil {
			t.Fatal("expected non-nil system info")
		}

		// Should have initialized slices (not nil)
		if info.Streams == nil {
			t.Error("Streams should be initialized slice, not nil")
		}
		if info.KVBuckets == nil {
			t.Error("KVBuckets should be initialized slice, not nil")
		}
		if info.ObjectStores == nil {
			t.Error("ObjectStores should be initialized slice, not nil")
		}
	})

	t.Run("excludes KV_ and OBJ_ prefixed streams", func(t *testing.T) {
		info, err := core.GetSystemInfo(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check that no stream names start with KV_ or OBJ_
		for _, s := range info.Streams {
			if len(s.Name) >= 3 && s.Name[:3] == "KV_" {
				t.Errorf("stream list should exclude KV_ streams, found: %s", s.Name)
			}
			if len(s.Name) >= 4 && s.Name[:4] == "OBJ_" {
				t.Errorf("stream list should exclude OBJ_ streams, found: %s", s.Name)
			}
		}
	})

	t.Run("includes KV buckets", func(t *testing.T) {
		info, err := core.GetSystemInfo(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Fresh instance should have at least the INSTANCE KV bucket
		found := false
		for _, kv := range info.KVBuckets {
			if kv.Name == "INSTANCE" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected INSTANCE KV bucket to be present")
		}
	})

	t.Run("stream info has valid fields after creating space", func(t *testing.T) {
		// Create a space to generate a stream
		user, err := core.CreateUser(ctx, SystemActorID, "sysinfo-user", "System Info User", "password123")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		space, err := core.CreateSpace(ctx, user.Id, "sysinfo-space", "System Info Space")
		if err != nil {
			t.Fatalf("failed to create space: %v", err)
		}

		info, err := core.GetSystemInfo(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Find the space stream
		var spaceStream *StreamInfo
		expectedStreamName := "SPACE_" + space.Id + "_EVENTS"
		for i := range info.Streams {
			if info.Streams[i].Name == expectedStreamName {
				spaceStream = &info.Streams[i]
				break
			}
		}

		if spaceStream == nil {
			t.Errorf("expected to find stream %s", expectedStreamName)
			return
		}

		// Verify stream info fields
		if spaceStream.Name != expectedStreamName {
			t.Errorf("expected stream name %s, got %s", expectedStreamName, spaceStream.Name)
		}
		if spaceStream.Created == "" {
			t.Error("stream Created should not be empty")
		}
	})
}

func TestChattoCore_GetConnectionInfo(t *testing.T) {
	core, _ := setupTestCore(t)

	t.Run("returns connection info", func(t *testing.T) {
		info := core.GetConnectionInfo()
		if info == nil {
			t.Fatal("expected non-nil connection info")
		}

		// Should be connected in tests
		if !info.Connected {
			t.Error("expected to be connected")
		}
	})

	t.Run("has valid server info when connected", func(t *testing.T) {
		info := core.GetConnectionInfo()
		if !info.Connected {
			t.Skip("not connected, skipping server info tests")
		}

		// Server ID should be non-empty
		if info.ServerID == "" {
			t.Error("expected non-empty ServerID when connected")
		}

		// Version should be non-empty
		if info.Version == "" {
			t.Error("expected non-empty Version when connected")
		}

		// MaxPayload should be > 0
		if info.MaxPayload <= 0 {
			t.Errorf("expected positive MaxPayload, got %d", info.MaxPayload)
		}
	})
}

func TestChattoCore_GetAccountInfo(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("returns account info", func(t *testing.T) {
		info, err := core.GetAccountInfo(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if info == nil {
			t.Fatal("expected non-nil account info")
		}

		// StreamsUsed should be >= 0
		if info.StreamsUsed < 0 {
			t.Errorf("expected non-negative StreamsUsed, got %d", info.StreamsUsed)
		}
	})

	t.Run("reflects usage after creating resources", func(t *testing.T) {
		// Get initial count
		initialInfo, err := core.GetAccountInfo(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		initialStreams := initialInfo.StreamsUsed

		// Create a space (which creates a stream)
		user, err := core.CreateUser(ctx, SystemActorID, "acctinfo-user", "Account Info User", "password123")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		_, err = core.CreateSpace(ctx, user.Id, "acctinfo-space", "Account Info Space")
		if err != nil {
			t.Fatalf("failed to create space: %v", err)
		}

		// Get updated count
		updatedInfo, err := core.GetAccountInfo(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have more streams now (at least the space event stream)
		if updatedInfo.StreamsUsed <= initialStreams {
			t.Errorf("expected StreamsUsed to increase after creating space, was %d now %d",
				initialStreams, updatedInfo.StreamsUsed)
		}
	})
}

func TestChattoCore_GetStreamSubjects(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	// Create a space with some activity
	user, err := core.CreateUser(ctx, SystemActorID, "streamsubj-user", "Stream Subject User", "password123")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	space, err := core.CreateSpace(ctx, user.Id, "streamsubj-space", "Stream Subject Space")
	if err != nil {
		t.Fatalf("failed to create space: %v", err)
	}

	// Create a room to add some subjects
	room, err := core.CreateRoom(ctx, user.Id, space.Id, "test-room", "Test Room")
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	// Join the room (actorID, space_id, user_id, room_id)
	_, err = core.JoinRoom(ctx, user.Id, space.Id, user.Id, room.Id)
	if err != nil {
		t.Fatalf("failed to join room: %v", err)
	}

	// Post a message (space_id, room_id, user_id, body, attachments, inThread, inReplyTo)
	_, err = core.PostMessage(ctx, space.Id, room.Id, user.Id, "Hello world", nil, "", "", nil, false)
	if err != nil {
		t.Fatalf("failed to post message: %v", err)
	}

	streamName := "SPACE_" + space.Id + "_EVENTS"

	t.Run("returns subjects for existing stream", func(t *testing.T) {
		subjects, err := core.GetStreamSubjects(ctx, streamName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(subjects) == 0 {
			t.Error("expected at least one subject in stream")
		}
	})

	t.Run("subjects are sorted by name", func(t *testing.T) {
		subjects, err := core.GetStreamSubjects(ctx, streamName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := 1; i < len(subjects); i++ {
			if subjects[i].Subject < subjects[i-1].Subject {
				t.Errorf("subjects not sorted: %s should come before %s",
					subjects[i].Subject, subjects[i-1].Subject)
			}
		}
	})

	t.Run("subject info has message counts", func(t *testing.T) {
		subjects, err := core.GetStreamSubjects(ctx, streamName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// At least one subject should have messages > 0
		hasMessages := false
		for _, s := range subjects {
			if s.Messages > 0 {
				hasMessages = true
				break
			}
		}
		if !hasMessages {
			t.Error("expected at least one subject with messages > 0")
		}
	})

	t.Run("returns error for non-existent stream", func(t *testing.T) {
		_, err := core.GetStreamSubjects(ctx, "NONEXISTENT_STREAM")
		if err == nil {
			t.Error("expected error for non-existent stream")
		}
	})
}

func TestChattoCore_GetKVKeys(t *testing.T) {
	core, _ := setupTestCore(t)
	ctx := testContext(t)

	t.Run("returns keys for INSTANCE bucket", func(t *testing.T) {
		keys, err := core.GetKVKeys(ctx, "INSTANCE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have at least some keys after initialization
		if keys == nil {
			t.Error("keys should not be nil")
		}
	})

	t.Run("returns keys after adding data", func(t *testing.T) {
		// Create a user (which adds keys to INSTANCE)
		_, err := core.CreateUser(ctx, SystemActorID, "kvkeys-user", "KV Keys User", "password123")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		keys, err := core.GetKVKeys(ctx, "INSTANCE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should have user-related keys
		if len(keys) == 0 {
			t.Error("expected at least one key after creating user")
		}
	})

	t.Run("keys are sorted", func(t *testing.T) {
		keys, err := core.GetKVKeys(ctx, "INSTANCE")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for i := 1; i < len(keys); i++ {
			if keys[i] < keys[i-1] {
				t.Errorf("keys not sorted: %s should come before %s", keys[i], keys[i-1])
			}
		}
	})

	t.Run("returns error for non-existent bucket", func(t *testing.T) {
		_, err := core.GetKVKeys(ctx, "NONEXISTENT_BUCKET")
		if err == nil {
			t.Error("expected error for non-existent bucket")
		}
	})

	t.Run("returns empty slice for empty bucket", func(t *testing.T) {
		// Create a new space to get an empty KV bucket
		user, err := core.CreateUser(ctx, SystemActorID, "emptykvuser", "Empty KV User", "password123")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}

		space, err := core.CreateSpace(ctx, user.Id, "emptykv-space", "Empty KV Space")
		if err != nil {
			t.Fatalf("failed to create space: %v", err)
		}

		// RUNTIME bucket should have minimal keys initially
		runtimeBucket := "SPACE_" + space.Id + "_RUNTIME"
		keys, err := core.GetKVKeys(ctx, runtimeBucket)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should return a slice (possibly empty, but not nil)
		if keys == nil {
			t.Error("keys should be an empty slice, not nil")
		}
	})
}

func TestStreamInfo_Fields(t *testing.T) {
	t.Run("StreamInfo has all expected fields", func(t *testing.T) {
		info := StreamInfo{
			Name:        "TEST_STREAM",
			Messages:    100,
			Bytes:       1024,
			Consumers:   2,
			Created:     "2024-01-01 00:00:00",
			FirstSeq:    1,
			LastSeq:     100,
			NumSubjects: 5,
		}

		if info.Name != "TEST_STREAM" {
			t.Errorf("expected Name 'TEST_STREAM', got '%s'", info.Name)
		}
		if info.Messages != 100 {
			t.Errorf("expected Messages 100, got %d", info.Messages)
		}
		if info.Bytes != 1024 {
			t.Errorf("expected Bytes 1024, got %d", info.Bytes)
		}
		if info.Consumers != 2 {
			t.Errorf("expected Consumers 2, got %d", info.Consumers)
		}
		if info.Created != "2024-01-01 00:00:00" {
			t.Errorf("expected Created '2024-01-01 00:00:00', got '%s'", info.Created)
		}
		if info.FirstSeq != 1 {
			t.Errorf("expected FirstSeq 1, got %d", info.FirstSeq)
		}
		if info.LastSeq != 100 {
			t.Errorf("expected LastSeq 100, got %d", info.LastSeq)
		}
		if info.NumSubjects != 5 {
			t.Errorf("expected NumSubjects 5, got %d", info.NumSubjects)
		}
	})
}

func TestKVBucketInfo_Fields(t *testing.T) {
	t.Run("KVBucketInfo has all expected fields", func(t *testing.T) {
		info := KVBucketInfo{
			Name:    "TEST_BUCKET",
			Keys:    50,
			Bytes:   2048,
			History: 5,
			TTL:     "1h0m0s",
		}

		if info.Name != "TEST_BUCKET" {
			t.Errorf("expected Name 'TEST_BUCKET', got '%s'", info.Name)
		}
		if info.Keys != 50 {
			t.Errorf("expected Keys 50, got %d", info.Keys)
		}
		if info.Bytes != 2048 {
			t.Errorf("expected Bytes 2048, got %d", info.Bytes)
		}
		if info.History != 5 {
			t.Errorf("expected History 5, got %d", info.History)
		}
		if info.TTL != "1h0m0s" {
			t.Errorf("expected TTL '1h0m0s', got '%s'", info.TTL)
		}
	})
}

func TestObjectStoreInfo_Fields(t *testing.T) {
	t.Run("ObjectStoreInfo has all expected fields", func(t *testing.T) {
		info := ObjectStoreInfo{
			Name:   "TEST_STORE",
			Size:   4096,
			Sealed: false,
		}

		if info.Name != "TEST_STORE" {
			t.Errorf("expected Name 'TEST_STORE', got '%s'", info.Name)
		}
		if info.Size != 4096 {
			t.Errorf("expected Size 4096, got %d", info.Size)
		}
		if info.Sealed {
			t.Error("expected Sealed false")
		}
	})
}

func TestConnectionInfo_Fields(t *testing.T) {
	t.Run("ConnectionInfo has all expected fields", func(t *testing.T) {
		info := ConnectionInfo{
			Connected:  true,
			ServerID:   "server-123",
			ServerName: "test-server",
			Version:    "2.10.0",
			MaxPayload: 1048576,
			RTT:        "1ms",
		}

		if !info.Connected {
			t.Error("expected Connected true")
		}
		if info.ServerID != "server-123" {
			t.Errorf("expected ServerID 'server-123', got '%s'", info.ServerID)
		}
		if info.ServerName != "test-server" {
			t.Errorf("expected ServerName 'test-server', got '%s'", info.ServerName)
		}
		if info.Version != "2.10.0" {
			t.Errorf("expected Version '2.10.0', got '%s'", info.Version)
		}
		if info.MaxPayload != 1048576 {
			t.Errorf("expected MaxPayload 1048576, got %d", info.MaxPayload)
		}
		if info.RTT != "1ms" {
			t.Errorf("expected RTT '1ms', got '%s'", info.RTT)
		}
	})
}

func TestAccountInfo_Fields(t *testing.T) {
	t.Run("AccountInfo has all expected fields", func(t *testing.T) {
		info := AccountInfo{
			Memory:        1073741824,
			MemoryUsed:    536870912,
			Storage:       10737418240,
			StorageUsed:   5368709120,
			Streams:       100,
			StreamsUsed:   50,
			Consumers:     1000,
			ConsumersUsed: 250,
		}

		if info.Memory != 1073741824 {
			t.Errorf("expected Memory 1073741824, got %d", info.Memory)
		}
		if info.MemoryUsed != 536870912 {
			t.Errorf("expected MemoryUsed 536870912, got %d", info.MemoryUsed)
		}
		if info.Storage != 10737418240 {
			t.Errorf("expected Storage 10737418240, got %d", info.Storage)
		}
		if info.StorageUsed != 5368709120 {
			t.Errorf("expected StorageUsed 5368709120, got %d", info.StorageUsed)
		}
		if info.Streams != 100 {
			t.Errorf("expected Streams 100, got %d", info.Streams)
		}
		if info.StreamsUsed != 50 {
			t.Errorf("expected StreamsUsed 50, got %d", info.StreamsUsed)
		}
		if info.Consumers != 1000 {
			t.Errorf("expected Consumers 1000, got %d", info.Consumers)
		}
		if info.ConsumersUsed != 250 {
			t.Errorf("expected ConsumersUsed 250, got %d", info.ConsumersUsed)
		}
	})
}
