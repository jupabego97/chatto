package graph

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"hmans.de/chatto/internal/core"
	"hmans.de/chatto/internal/graph/model"
)

func userSuspensionToModel(suspension core.UserSuspension) *model.UserSuspension {
	var expiresAt *timestamppb.Timestamp
	if suspension.ExpiresAt != nil {
		expiresAt = timestamppb.New(*suspension.ExpiresAt)
	}
	return &model.UserSuspension{
		ID:          suspension.EventID,
		UserID:      suspension.UserID,
		ModeratorID: suspension.ModeratorID,
		Reason:      suspension.Reason,
		CreatedAt:   timestamppb.New(suspension.CreatedAt),
		ExpiresAt:   expiresAt,
	}
}
