/**
 * Centralized time formatting utilities that respect user settings.
 *
 * All functions accept either a Date object or an ISO string.
 * When timezone/timeFormat are unset, browser defaults are used.
 *
 * Intl.DateTimeFormat instances are cached because construction is expensive
 * (parses locale + timezone data). The cache is keyed by serialized options,
 * so formatters are reused across calls with the same settings.
 */

import type { UserSettingsState } from '$lib/state/userSettings.svelte';

function toDate(date: Date | string): Date {
  return typeof date === 'string' ? new Date(date) : date;
}

/** Cache of Intl.DateTimeFormat instances keyed by locale + options. */
const formatterCache = new Map<string, Intl.DateTimeFormat>();

function getFormatter(
  locale: string | undefined,
  options: Intl.DateTimeFormatOptions
): Intl.DateTimeFormat {
  const key = `${locale ?? ''}:${JSON.stringify(options)}`;
  let fmt = formatterCache.get(key);
  if (!fmt) {
    fmt = new Intl.DateTimeFormat(locale, options);
    formatterCache.set(key, fmt);
  }
  return fmt;
}

/**
 * Format a message timestamp (e.g., "2:30 PM" or "14:30").
 */
export function formatMessageTime(date: Date | string, settings: UserSettingsState): string {
  const fmt = getFormatter('en-US', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: settings.effectiveHour12,
    timeZone: settings.effectiveTimezone
  });
  return fmt.format(toDate(date));
}

/**
 * Format a date for display (e.g., "Jan 15, 2025").
 */
export function formatDate(date: Date | string, settings: UserSettingsState): string {
  const fmt = getFormatter(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    timeZone: settings.effectiveTimezone
  });
  return fmt.format(toDate(date));
}

/**
 * Format a date with time for display (e.g., "November 15, 2025, 02:30 PM").
 */
export function formatDateTime(date: Date | string, settings: UserSettingsState): string {
  const fmt = getFormatter(undefined, {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    hour12: settings.effectiveHour12,
    timeZone: settings.effectiveTimezone
  });
  return fmt.format(toDate(date));
}

/**
 * Check if two dates fall on the same calendar day in the user's timezone.
 */
export function isSameDay(date1: Date, date2: Date, settings: UserSettingsState): boolean {
  const fmt = getFormatter('en-US', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    timeZone: settings.effectiveTimezone
  });
  return fmt.format(date1) === fmt.format(date2);
}

/**
 * Format a day separator label ("Today", "Yesterday", or a full date).
 */
export function formatDayLabel(date: Date | string, settings: UserSettingsState): string {
  const d = toDate(date);
  const now = new Date();

  if (isSameDay(d, now, settings)) {
    return 'Today';
  }

  const yesterday = new Date(now);
  yesterday.setDate(yesterday.getDate() - 1);
  if (isSameDay(d, yesterday, settings)) {
    return 'Yesterday';
  }

  const tz = settings.effectiveTimezone;
  const yearFmt = getFormatter('en-US', { year: 'numeric', timeZone: tz });
  const sameYear = yearFmt.format(d) === yearFmt.format(now);

  const labelFmt = getFormatter('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    year: sameYear ? undefined : 'numeric',
    timeZone: tz
  });
  return labelFmt.format(d);
}
