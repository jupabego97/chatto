/**
 * Prevents pinch-to-zoom on trackpads via wheel and gesture events.
 * Call during component initialization — sets up listeners in an $effect.
 */

const ZOOMABLE_VIEWPORT =
  'width=device-width, initial-scale=1, maximum-scale=10, user-scalable=yes, viewport-fit=cover, interactive-widget=resizes-content';

let nativeZoomAllowanceCount = 0;
let originalViewportContent: string | null = null;
let originalHtmlTouchAction: string | null = null;
let originalBodyTouchAction: string | null = null;

function nativeZoomAllowed() {
  return nativeZoomAllowanceCount > 0;
}

function setNativeZoomAllowed(allowed: boolean) {
  if (!allowed || typeof document === 'undefined') return;

  const viewport = document.querySelector<HTMLMetaElement>('meta[name="viewport"]');
  if (nativeZoomAllowanceCount === 0) {
    originalViewportContent = viewport?.getAttribute('content') ?? null;
    originalHtmlTouchAction = document.documentElement.style.touchAction;
    originalBodyTouchAction = document.body.style.touchAction;
    viewport?.setAttribute('content', ZOOMABLE_VIEWPORT);
    document.documentElement.style.touchAction = 'auto';
    document.body.style.touchAction = 'auto';
  }

  nativeZoomAllowanceCount += 1;
}

function clearNativeZoomAllowed() {
  if (nativeZoomAllowanceCount === 0 || typeof document === 'undefined') return;

  nativeZoomAllowanceCount -= 1;
  if (nativeZoomAllowanceCount > 0) return;

  const viewport = document.querySelector<HTMLMetaElement>('meta[name="viewport"]');
  if (viewport && originalViewportContent !== null) {
    viewport.setAttribute('content', originalViewportContent);
  }
  document.documentElement.style.touchAction = originalHtmlTouchAction ?? '';
  document.body.style.touchAction = originalBodyTouchAction ?? '';
  originalViewportContent = null;
  originalHtmlTouchAction = null;
  originalBodyTouchAction = null;
}

export function usePinchZoomPrevention() {
  $effect(() => {
    function onWheel(e: WheelEvent) {
      if (e.ctrlKey && !nativeZoomAllowed()) e.preventDefault();
    }
    function onGesture(e: Event) {
      if (!nativeZoomAllowed()) e.preventDefault();
    }

    document.addEventListener('wheel', onWheel, { passive: false });
    document.addEventListener('gesturestart', onGesture);
    document.addEventListener('gesturechange', onGesture);

    return () => {
      document.removeEventListener('wheel', onWheel);
      document.removeEventListener('gesturestart', onGesture);
      document.removeEventListener('gesturechange', onGesture);
    };
  });
}

export function useNativeZoomAllowance(getAllowed: () => boolean) {
  $effect(() => {
    if (!getAllowed()) return;

    setNativeZoomAllowed(true);
    return () => clearNativeZoomAllowed();
  });
}
