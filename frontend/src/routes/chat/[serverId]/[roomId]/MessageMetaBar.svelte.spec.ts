import { describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { q } from '$lib/test-utils';
import MessageMetaBar from './MessageMetaBar.svelte';

const mocks = vi.hoisted(() => ({
  tryWireAddReaction: vi.fn(),
  tryWireRemoveReaction: vi.fn(),
  toastError: vi.fn()
}));

vi.mock('$lib/wire', () => ({
  tryWireAddReaction: mocks.tryWireAddReaction,
  tryWireRemoveReaction: mocks.tryWireRemoveReaction
}));

vi.mock('$lib/ui/toast', () => ({
  toast: {
    error: mocks.toastError
  }
}));

const baseProps = {
  roomId: 'room-1',
  messageEventId: 'message-1',
  reactions: [],
  onOpenThread: vi.fn()
};

function buttonWithText(container: HTMLElement, text: string): HTMLButtonElement {
  const button = Array.from(container.querySelectorAll('button')).find((candidate) =>
    candidate.textContent?.replace(/\s+/g, ' ').trim().includes(text)
  );
  if (!(button instanceof HTMLButtonElement)) {
    throw new Error(`Button not found: ${text}`);
  }
  return button;
}

describe('MessageMetaBar', () => {
  it('renders the reply count button and opens the thread', () => {
    const onOpenThread = vi.fn();
    const { container } = render(MessageMetaBar, {
      props: {
        ...baseProps,
        onOpenThread,
        replyCount: 2
      }
    });

    buttonWithText(container, '2 replies').click();

    expect(onOpenThread).toHaveBeenCalledOnce();
  });

  it('renders the echo thread button and opens the thread', () => {
    const onOpenThread = vi.fn();
    const { container } = render(MessageMetaBar, {
      props: {
        ...baseProps,
        onOpenThread,
        isEchoEvent: true
      }
    });

    buttonWithText(container, 'Thread').click();

    expect(onOpenThread).toHaveBeenCalledOnce();
  });

  it('keeps follow toggles as buttons', () => {
    const { container } = render(MessageMetaBar, {
      props: {
        ...baseProps,
        replyCount: 1,
        isFollowingThread: true,
        onToggleThreadFollow: vi.fn()
      }
    });

    const followButton = q(container, 'button[title="Unfollow thread"]');

    expect(followButton).not.toBeNull();
  });

  it('adds reactions through the wire helper', async () => {
    mocks.tryWireAddReaction.mockResolvedValue(true);
    const { container } = render(MessageMetaBar, {
      props: {
        ...baseProps,
        canReact: true,
        reactions: [
          {
            emoji: 'thumbs_up',
            count: 1,
            hasReacted: false,
            users: [{ id: 'user-1', displayName: 'Ada' }]
          }
        ]
      }
    });

    (q(container, 'button[aria-pressed="false"]') as HTMLButtonElement).click();

    await vi.waitFor(() => {
      expect(mocks.tryWireAddReaction).toHaveBeenCalledWith({
        roomId: 'room-1',
        messageEventId: 'message-1',
        emoji: 'thumbs_up'
      });
    });
  });

  it('removes reactions through the wire helper', async () => {
    mocks.tryWireRemoveReaction.mockResolvedValue(true);
    const { container } = render(MessageMetaBar, {
      props: {
        ...baseProps,
        canReact: true,
        reactions: [
          {
            emoji: 'thumbs_up',
            count: 1,
            hasReacted: true,
            users: [{ id: 'user-1', displayName: 'Ada' }]
          }
        ]
      }
    });

    (q(container, 'button[aria-pressed="true"]') as HTMLButtonElement).click();

    await vi.waitFor(() => {
      expect(mocks.tryWireRemoveReaction).toHaveBeenCalledWith({
        roomId: 'room-1',
        messageEventId: 'message-1',
        emoji: 'thumbs_up'
      });
    });
  });
});
