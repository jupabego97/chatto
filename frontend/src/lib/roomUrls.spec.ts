import { describe, expect, it } from 'vitest';
import { RoomType } from '$lib/gql/graphql';
import {
  looksLikeRoomIDSegment,
  roomMessagePathForSegment,
  roomPath,
  roomPathForSegment,
  roomThreadPathForSegment,
  roomURLSegment
} from './roomUrls';

describe('room URL helpers', () => {
  it('uses channel names as canonical room URL segments', () => {
    expect(roomURLSegment({ id: 'R1', name: 'General', type: RoomType.Channel })).toBe('General');
  });

  it('recognizes legacy room ID segments for fallback lookup', () => {
    expect(looksLikeRoomIDSegment('R7gUDvZNyvHkk4K')).toBe(true);
    expect(looksLikeRoomIDSegment('deadbeef123456')).toBe(true);
    expect(looksLikeRoomIDSegment('general')).toBe(false);
  });

  it('keeps ordinary capital-R channel names as name URLs', () => {
    expect(roomURLSegment({ id: 'R1', name: 'Random', type: RoomType.Channel })).toBe('Random');
  });

  it('uses DM IDs as canonical room URL segments', () => {
    expect(roomURLSegment({ id: 'deadbeef123456', name: '', type: RoomType.Dm })).toBe(
      'deadbeef123456'
    );
  });

  it('builds room paths from canonical targets and raw segments', () => {
    expect(roomPath('origin', { id: 'R1', name: 'General', type: RoomType.Channel })).toBe(
      '/chat/-/r/General'
    );
    expect(roomPathForSegment('-', 'R1')).toBe('/chat/-/R1');
    expect(roomPathForSegment('-', 'General', 'name')).toBe('/chat/-/r/General');
    expect(roomThreadPathForSegment('-', 'General', 'Eroot', 'name')).toBe(
      '/chat/-/r/General/Eroot'
    );
    expect(roomMessagePathForSegment('-', 'General', 'Emsg', 'name')).toBe(
      '/chat/-/r/General/m/Emsg'
    );
  });
});
