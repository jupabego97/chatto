package graph

import (
	"hmans.de/chatto/internal/graph/model"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
)

// protoLayoutToModel converts a proto RoomLayout to the GraphQL model,
// attaching the pre-resolved room map (all rooms in the space) so
// sub-resolvers can efficiently resolve room IDs to Room objects.
func protoLayoutToModel(layout *corev1.RoomLayout, viewerRooms map[string]*corev1.Room) *model.RoomLayoutModel {
	sections := make([]*model.RoomLayoutSectionModel, len(layout.Sections))
	for i, s := range layout.Sections {
		sections[i] = &model.RoomLayoutSectionModel{
			ID:          s.Id,
			Name:        s.Name,
			RoomIds:     s.RoomIds,
			ViewerRooms: viewerRooms,
		}
	}

	return &model.RoomLayoutModel{
		Sections:           sections,
		UnsectionedRoomIds: layout.UnsortedRoomIds,
		ViewerRooms:        viewerRooms,
	}
}
