package subjects

import "testing"

func TestSpaceRoomMessage(t *testing.T) {
	tests := []struct {
		name     string
		spaceID  string
		roomID   string
		eventID  string
		expected string
	}{
		{
			name:     "basic message subject",
			spaceID:  "space1",
			roomID:   "room1",
			eventID:  "evt123",
			expected: "space.space1.room.room1.msg.evt123",
		},
		{
			name:     "with nanoid-style IDs",
			spaceID:  "Sp6IQDs4Hm6gLIb",
			roomID:   "R7IFBV0AV1UBYTK",
			eventID:  "E8ShdnxI4BouAIl",
			expected: "space.Sp6IQDs4Hm6gLIb.room.R7IFBV0AV1UBYTK.msg.E8ShdnxI4BouAIl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SpaceRoomMessage(tt.spaceID, tt.roomID, tt.eventID)
			if got != tt.expected {
				t.Errorf("SpaceRoomMessage() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSpaceRoomThread(t *testing.T) {
	tests := []struct {
		name        string
		spaceID     string
		roomID      string
		rootEventID string
		eventID     string
		expected    string
	}{
		{
			name:        "basic thread subject",
			spaceID:     "space1",
			roomID:      "room1",
			rootEventID: "evt123",
			eventID:     "evt456",
			expected:    "space.space1.room.room1.msg.evt123.replies.evt456",
		},
		{
			name:        "with nanoid-style IDs",
			spaceID:     "Sp6IQDs4Hm6gLIb",
			roomID:      "R7IFBV0AV1UBYTK",
			rootEventID: "E7RootEventId",
			eventID:     "E8ShdnxI4BouAIl",
			expected:    "space.Sp6IQDs4Hm6gLIb.room.R7IFBV0AV1UBYTK.msg.E7RootEventId.replies.E8ShdnxI4BouAIl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SpaceRoomThread(tt.spaceID, tt.roomID, tt.rootEventID, tt.eventID)
			if got != tt.expected {
				t.Errorf("SpaceRoomThread() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSpaceRoomThreadFilter(t *testing.T) {
	got := SpaceRoomThreadFilter("space1", "room1", "evt123")
	expected := "space.space1.room.room1.msg.evt123.replies.>"
	if got != expected {
		t.Errorf("SpaceRoomThreadFilter() = %q, want %q", got, expected)
	}
}

func TestSpaceRoomThreadLookup(t *testing.T) {
	got := SpaceRoomThreadLookup("space1", "room1", "evt456")
	expected := "space.space1.room.room1.msg.*.replies.evt456"
	if got != expected {
		t.Errorf("SpaceRoomThreadLookup() = %q, want %q", got, expected)
	}
}

func TestSpaceRoomAllThreads(t *testing.T) {
	got := SpaceRoomAllThreads("space1", "room1")
	expected := "space.space1.room.room1.msg.*.replies.>"
	if got != expected {
		t.Errorf("SpaceRoomAllThreads() = %q, want %q", got, expected)
	}
}

func TestSpaceRoomRootMessages(t *testing.T) {
	got := SpaceRoomRootMessages("space1", "room1")
	expected := "space.space1.room.room1.msg.*"
	if got != expected {
		t.Errorf("SpaceRoomRootMessages() = %q, want %q", got, expected)
	}
}

func TestSpaceRoomAllEvents(t *testing.T) {
	got := SpaceRoomAllEvents("space1", "room1")
	expected := "space.space1.room.room1.>"
	if got != expected {
		t.Errorf("SpaceRoomAllEvents() = %q, want %q", got, expected)
	}
}

func TestParseRoomIDFromSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected string
	}{
		{
			name:     "root message",
			subject:  "space.space1.room.room1.msg.evt123",
			expected: "room1",
		},
		{
			name:     "thread reply",
			subject:  "space.space1.room.room1.msg.evt123.replies.evt456",
			expected: "room1",
		},
		{
			name:     "meta event",
			subject:  "space.space1.room.room1.meta",
			expected: "room1",
		},
		{
			name:     "space-level event (not a room)",
			subject:  "space.space1.joined",
			expected: "",
		},
		{
			name:     "invalid subject",
			subject:  "invalid.subject",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseRoomIDFromSubject(tt.subject)
			if got != tt.expected {
				t.Errorf("ParseRoomIDFromSubject(%q) = %q, want %q", tt.subject, got, tt.expected)
			}
		})
	}
}

func TestParseThreadRootEventIDFromSubject(t *testing.T) {
	tests := []struct {
		name              string
		subject           string
		expectedEventID   string
		expectedOK        bool
	}{
		{
			name:            "thread reply",
			subject:         "space.space1.room.room1.msg.evt123.replies.evt456",
			expectedEventID: "evt123",
			expectedOK:      true,
		},
		{
			name:            "root message",
			subject:         "space.space1.room.room1.msg.evt123",
			expectedEventID: "",
			expectedOK:      false,
		},
		{
			name:            "meta event",
			subject:         "space.space1.room.room1.meta",
			expectedEventID: "",
			expectedOK:      false,
		},
		{
			name:            "nanoid-style IDs",
			subject:         "space.Sp6IQDs4Hm6gLIb.room.R7IFBV0AV1UBYTK.msg.E7RootEventId.replies.E8ShdnxI4BouAIl",
			expectedEventID: "E7RootEventId",
			expectedOK:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventID, ok := ParseThreadRootEventIDFromSubject(tt.subject)
			if eventID != tt.expectedEventID || ok != tt.expectedOK {
				t.Errorf("ParseThreadRootEventIDFromSubject(%q) = (%q, %v), want (%q, %v)", tt.subject, eventID, ok, tt.expectedEventID, tt.expectedOK)
			}
		})
	}
}

func TestIsRootMessageSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected bool
	}{
		{
			name:     "root message",
			subject:  "space.space1.room.room1.msg.evt123",
			expected: true,
		},
		{
			name:     "thread reply",
			subject:  "space.space1.room.room1.msg.evt123.replies.evt456",
			expected: false,
		},
		{
			name:     "meta event",
			subject:  "space.space1.room.room1.meta",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRootMessageSubject(tt.subject)
			if got != tt.expected {
				t.Errorf("IsRootMessageSubject(%q) = %v, want %v", tt.subject, got, tt.expected)
			}
		})
	}
}

func TestIsMetaSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected bool
	}{
		{
			name:     "meta event",
			subject:  "space.space1.room.room1.meta",
			expected: true,
		},
		{
			name:     "root message",
			subject:  "space.space1.room.room1.msg.evt123",
			expected: false,
		},
		{
			name:     "thread reply",
			subject:  "space.space1.room.room1.msg.evt123.replies.evt456",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMetaSubject(tt.subject)
			if got != tt.expected {
				t.Errorf("IsMetaSubject(%q) = %v, want %v", tt.subject, got, tt.expected)
			}
		})
	}
}

func TestIsThreadSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected bool
	}{
		{
			name:     "thread reply",
			subject:  "space.space1.room.room1.msg.evt123.replies.evt456",
			expected: true,
		},
		{
			name:     "root message",
			subject:  "space.space1.room.room1.msg.evt123",
			expected: false,
		},
		{
			name:     "meta event",
			subject:  "space.space1.room.room1.meta",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsThreadSubject(tt.subject)
			if got != tt.expected {
				t.Errorf("IsThreadSubject(%q) = %v, want %v", tt.subject, got, tt.expected)
			}
		})
	}
}

func TestParseEventIDFromSubject(t *testing.T) {
	tests := []struct {
		name     string
		subject  string
		expected string
	}{
		{
			name:     "root message",
			subject:  "space.space1.room.room1.msg.evt123",
			expected: "evt123",
		},
		{
			name:     "thread reply",
			subject:  "space.space1.room.room1.msg.evt123.replies.evt456",
			expected: "evt456",
		},
		{
			name:     "meta event (no event ID)",
			subject:  "space.space1.room.room1.meta",
			expected: "",
		},
		{
			name:     "space-level event",
			subject:  "space.space1.joined",
			expected: "",
		},
		{
			name:     "nanoid-style event ID",
			subject:  "space.Sp6IQDs4Hm6gLIb.room.R7IFBV0AV1UBYTK.msg.E8ShdnxI4BouAIl",
			expected: "E8ShdnxI4BouAIl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseEventIDFromSubject(tt.subject)
			if got != tt.expected {
				t.Errorf("ParseEventIDFromSubject(%q) = %q, want %q", tt.subject, got, tt.expected)
			}
		})
	}
}

func TestSpaceRoomRootEventsFilters(t *testing.T) {
	got := SpaceRoomRootEventsFilters("space1", "room1")
	expected := []string{
		"space.space1.room.room1.msg.*",
		"space.space1.room.room1.meta",
	}
	if len(got) != len(expected) {
		t.Fatalf("SpaceRoomRootEventsFilters() returned %d elements, want %d", len(got), len(expected))
	}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("SpaceRoomRootEventsFilters()[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestSpaceAllRoomEventsFilters(t *testing.T) {
	got := SpaceAllRoomEventsFilters("space1")
	expected := []string{
		"space.space1.room.*.msg.>",
		"space.space1.room.*.meta",
	}
	if len(got) != len(expected) {
		t.Fatalf("SpaceAllRoomEventsFilters() returned %d elements, want %d", len(got), len(expected))
	}
	for i, v := range got {
		if v != expected[i] {
			t.Errorf("SpaceAllRoomEventsFilters()[%d] = %q, want %q", i, v, expected[i])
		}
	}
}
