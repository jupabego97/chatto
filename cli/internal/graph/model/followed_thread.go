package model

import "google.golang.org/protobuf/types/known/timestamppb"

// FollowedThread is the GraphQL model for a thread the user is following.
type FollowedThread struct {
	SpaceID           string                 `json:"spaceId"`
	RoomID            string                 `json:"roomId"`
	ThreadRootEventID string                 `json:"threadRootEventId"`
	ReplyCount        int32                  `json:"replyCount"`
	LastReplyAt       *timestamppb.Timestamp `json:"lastReplyAt"`
	HasUnread         bool                   `json:"hasUnread"`

	// Internal fields for resolvers (not exposed in GraphQL)
	ParticipantIDs []string `json:"-"`
}
