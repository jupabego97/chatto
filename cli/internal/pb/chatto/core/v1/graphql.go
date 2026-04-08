package corev1

// Implement GraphQL SpaceEventType interface for space-scoped events.
func (*RoomCreatedEvent) IsSpaceEventType()        {}
func (*RoomUpdatedEvent) IsSpaceEventType()        {}
func (*RoomDeletedEvent) IsSpaceEventType()        {}
func (*RoomArchivedEvent) IsSpaceEventType()       {}
func (*RoomUnarchivedEvent) IsSpaceEventType()     {}
func (*UserJoinedRoomEvent) IsSpaceEventType()     {}
func (*UserLeftRoomEvent) IsSpaceEventType()       {}
func (*SpaceMemberDeletedEvent) IsSpaceEventType() {}
func (*MessagePostedEvent) IsSpaceEventType()              {}
func (*MessageUpdatedEvent) IsSpaceEventType()             {}
func (*MessageDeletedEvent) IsSpaceEventType()             {}
func (*ReactionAddedEvent) IsSpaceEventType()              {}
func (*ReactionRemovedEvent) IsSpaceEventType()            {}
func (*UserTypingEvent) IsSpaceEventType()                 {}
func (*PresenceChangedEvent) IsSpaceEventType()            {}
func (*VideoProcessingCompletedEvent) IsSpaceEventType()   {}
func (*CallParticipantJoinedEvent) IsSpaceEventType()      {}
func (*CallParticipantLeftEvent) IsSpaceEventType()        {}

// Implement GraphQL InstanceEventType interface for instance-scoped events.
func (*InstanceConfigUpdatedEvent) IsInstanceEventType()          {}
func (*UserCreatedEvent) IsInstanceEventType()                    {}
func (*UserDeletedEvent) IsInstanceEventType()                    {}
func (*UserProfileUpdatedEvent) IsInstanceEventType()             {}
func (*InstanceUserPreferencesUpdatedEvent) IsInstanceEventType() {}
func (*NotificationLevelChangedEvent) IsInstanceEventType()       {}
func (*ThreadFollowChangedEvent) IsInstanceEventType()            {}
func (*UserJoinedSpaceEvent) IsInstanceEventType()                {}
func (*UserLeftSpaceEvent) IsInstanceEventType()                  {}
func (*SpaceCreatedEvent) IsInstanceEventType()                   {}
func (*SpaceUpdatedEvent) IsInstanceEventType()                   {}
func (*SpaceDeletedEvent) IsInstanceEventType()                   {}
func (*MentionNotificationEvent) IsInstanceEventType()            {}
func (*NewDirectMessageNotificationEvent) IsInstanceEventType()   {}
func (*NotificationCreatedEvent) IsInstanceEventType()            {}
func (*NotificationDismissedEvent) IsInstanceEventType()          {}
func (*NewMessageInSpaceEvent) IsInstanceEventType()              {}
func (*RoomMarkedAsReadEvent) IsInstanceEventType()               {}
func (*RoomLayoutUpdatedEvent) IsInstanceEventType()              {}
func (*SessionTerminatedEvent) IsInstanceEventType()              {}
