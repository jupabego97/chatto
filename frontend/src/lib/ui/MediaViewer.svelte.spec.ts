import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { render } from 'vitest-browser-svelte';
import MediaViewer from './MediaViewer.svelte';

const LOCKED_VIEWPORT =
  'width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no, viewport-fit=cover, interactive-widget=resizes-content';

function ensureViewportMeta(): HTMLMetaElement {
  let viewport = document.querySelector<HTMLMetaElement>('meta[name="viewport"]');
  if (!viewport) {
    viewport = document.createElement('meta');
    viewport.name = 'viewport';
    document.head.append(viewport);
  }
  viewport.setAttribute('content', LOCKED_VIEWPORT);
  return viewport;
}

beforeEach(() => {
  ensureViewportMeta();
  document.documentElement.style.touchAction = '';
  document.body.style.touchAction = '';
});

afterEach(() => {
  ensureViewportMeta();
  document.documentElement.style.touchAction = '';
  document.body.style.touchAction = '';
});

describe('MediaViewer', () => {
  it('opens the original image action in a new window', async () => {
    const open = vi.spyOn(window, 'open').mockImplementation(() => null);
    const { container } = render(MediaViewer, {
      props: {
        items: [
          {
            kind: 'image',
            src: 'https://cdn.example.com/current.jpg',
            filename: 'image.jpg'
          }
        ],
        onclose: () => {}
      }
    });

    const button = [...container.querySelectorAll<HTMLButtonElement>('button')].find((candidate) =>
      candidate.textContent?.includes('Open original')
    )!;
    button.click();

    expect(open).toHaveBeenCalledWith(
      'https://cdn.example.com/current.jpg',
      '_blank',
      'noopener,noreferrer'
    );
  });

  it('hides navigation and counter for a single item', async () => {
    const { container } = render(MediaViewer, {
      props: {
        items: [
          {
            kind: 'image',
            src: 'https://cdn.example.com/current.jpg',
            filename: 'image.jpg'
          }
        ],
        onclose: () => {}
      }
    });

    expect(container.querySelector('[aria-label="Previous media"]')).toBeNull();
    expect(container.querySelector('[aria-label="Next media"]')).toBeNull();
    await expect.element(container).not.toHaveTextContent(/\d+ \/ \d+/);
  });

  it('supports keyboard and button navigation for multiple items', async () => {
    const { container } = render(MediaViewer, {
      props: {
        items: [
          {
            kind: 'image',
            src: 'https://cdn.example.com/first.jpg',
            filename: 'first.jpg'
          },
          {
            kind: 'image',
            src: 'https://cdn.example.com/second.jpg',
            filename: 'second.jpg'
          }
        ],
        onclose: () => {}
      }
    });

    const dialog = container.querySelector('dialog')!;
    await expect.element(container).toHaveTextContent('1 / 2');

    dialog.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight', bubbles: true }));
    await expect.element(container).toHaveTextContent('2 / 2');
    await expect.element(container).toHaveTextContent('second.jpg');

    container
      .querySelector<HTMLButtonElement>('[aria-label="Next media"]')!
      .click();
    await expect.element(container).toHaveTextContent('1 / 2');
  });

  it('allows native zoom while an image is active and restores it on close', async () => {
    const viewport = ensureViewportMeta();
    const rendered = render(MediaViewer, {
      props: {
        items: [
          {
            kind: 'image',
            src: 'https://cdn.example.com/current.jpg',
            filename: 'image.jpg'
          }
        ],
        onclose: () => {}
      }
    });

    await vi.waitFor(() => {
      expect(viewport.getAttribute('content')).toContain('user-scalable=yes');
      expect(document.documentElement.style.touchAction).toBe('auto');
      expect(document.body.style.touchAction).toBe('auto');
    });

    rendered.unmount();

    await vi.waitFor(() => {
      expect(viewport.getAttribute('content')).toBe(LOCKED_VIEWPORT);
      expect(document.documentElement.style.touchAction).toBe('');
      expect(document.body.style.touchAction).toBe('');
    });
  });

  it('does not allow native zoom for a video item', async () => {
    const viewport = ensureViewportMeta();
    render(MediaViewer, {
      props: {
        items: [
          {
            kind: 'video',
            src: 'https://cdn.example.com/video.mp4',
            filename: 'clip.mp4',
            source: { kind: 'asset' }
          }
        ],
        onclose: () => {}
      }
    });

    await vi.waitFor(() => {
      expect(viewport.getAttribute('content')).toBe(LOCKED_VIEWPORT);
      expect(document.documentElement.style.touchAction).toBe('');
      expect(document.body.style.touchAction).toBe('');
    });
  });
});
