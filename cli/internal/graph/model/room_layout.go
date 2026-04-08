package model

import corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"

// RoomLayoutModel is the GraphQL model for RoomLayout.
// It wraps the proto RoomLayout with pre-resolved viewer room data
// so sub-resolvers can efficiently resolve room IDs to Room objects.
type RoomLayoutModel struct {
	// Sections from the proto layout
	Sections []*RoomLayoutSectionModel

	// UnsectionedRoomIds is the ordered list of unsectioned room IDs.
	// When non-empty, the Unsectioned resolver uses this ordering.
	UnsectionedRoomIds []string

	// ViewerRooms maps room ID → Room for all rooms in the space.
	// Used by sub-resolvers (Unsectioned, Rooms) to resolve room IDs.
	ViewerRooms map[string]*corev1.Room
}

// RoomLayoutSectionModel is the GraphQL model for RoomLayoutSection.
type RoomLayoutSectionModel struct {
	ID      string
	Name    string
	RoomIds []string

	// ViewerRooms is a reference to the parent layout's ViewerRooms map.
	ViewerRooms map[string]*corev1.Room
}
