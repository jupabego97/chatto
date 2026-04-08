const STORAGE_KEY = 'chatto:threadPaneWidth';
const DEFAULT_WIDTH = 50; // percentage

export const THREAD_PANE_MIN_WIDTH = 25;
export const THREAD_PANE_MAX_WIDTH = 75;

export function getThreadPaneWidth(): number {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      const value = parseFloat(stored);
      if (!isNaN(value) && value >= THREAD_PANE_MIN_WIDTH && value <= THREAD_PANE_MAX_WIDTH) {
        return value;
      }
    }
  } catch {
    // Ignore storage errors
  }
  return DEFAULT_WIDTH;
}

export function setThreadPaneWidth(width: number): void {
  const clamped = Math.min(THREAD_PANE_MAX_WIDTH, Math.max(THREAD_PANE_MIN_WIDTH, width));
  try {
    localStorage.setItem(STORAGE_KEY, String(clamped));
  } catch {
    // Ignore storage errors
  }
}
