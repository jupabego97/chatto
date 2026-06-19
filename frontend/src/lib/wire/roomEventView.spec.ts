import { Timestamp } from '@bufbuild/protobuf';
import { describe, expect, it } from 'vitest';
import { PresenceStatus } from '$lib/chatTypes';
import { User } from '$lib/pb/chatto/core/v1/models_pb';
import {
  CurrentUserPresenceStatus,
  MessagePostedView,
  ReactionSummaryView,
  RoomEventPayload,
  RoomEventView,
  UserAvatarView
} from '$lib/pb/chatto/api/v1/chat_pb';
import { wireRoomEventViewToFragment } from './roomEventView';

describe('wireRoomEventViewToFragment', () => {
  it('preserves hydrated avatar and presence data from wire user views', () => {
    const fragment = wireRoomEventViewToFragment(
      new RoomEventView({
        id: 'evt-1',
        createdAt: Timestamp.fromDate(new Date('2026-01-02T03:04:05Z')),
        actorId: 'u_actor',
        actor: userAvatar(
          'u_actor',
          'actor',
          'Actor',
          '/avatars/actor.webp',
          CurrentUserPresenceStatus.ONLINE
        ),
        event: new RoomEventPayload({
          payload: {
            case: 'messagePosted',
            value: new MessagePostedView({
              roomId: 'room-1',
              body: 'hello',
              threadParticipants: [
                userAvatar(
                  'u_thread',
                  'thread',
                  'Thread',
                  '/avatars/thread.webp',
                  CurrentUserPresenceStatus.AWAY
                )
              ],
              reactions: [
                new ReactionSummaryView({
                  emoji: ':wave:',
                  count: 1,
                  users: [
                    userAvatar(
                      'u_reaction',
                      'reaction',
                      'Reaction',
                      '/avatars/reaction.webp',
                      CurrentUserPresenceStatus.DO_NOT_DISTURB
                    )
                  ]
                })
              ]
            })
          }
        })
      })
    );

    expect(fragment?.actor).toMatchObject({
      id: 'u_actor',
      avatarUrl: '/avatars/actor.webp',
      presenceStatus: PresenceStatus.Online
    });
    expect(fragment?.event.__typename).toBe('MessagePostedEvent');
    if (fragment?.event.__typename !== 'MessagePostedEvent') return;
    expect(fragment.event.threadParticipants[0]).toMatchObject({
      id: 'u_thread',
      avatarUrl: '/avatars/thread.webp',
      presenceStatus: PresenceStatus.Away
    });
    expect(fragment.event.reactions[0]?.users[0]).toMatchObject({
      id: 'u_reaction',
      displayName: 'Reaction'
    });
  });
});

function userAvatar(
  id: string,
  login: string,
  displayName: string,
  avatarUrl: string,
  presenceStatus: CurrentUserPresenceStatus
) {
  return new UserAvatarView({
    user: new User({ id, login, displayName }),
    avatarUrl,
    presenceStatus
  });
}
