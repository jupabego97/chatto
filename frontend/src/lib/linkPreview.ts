/**
 * URL detection utilities for link previews.
 *
 * Uses linkify-it (same library as markdown-it's auto-linker) with the full
 * IANA TLD list so that bare-domain URLs like www.hmans.dev are detected.
 */

import LinkifyIt from 'linkify-it';
import tlds from 'tlds';

/** Shared linkify-it instance configured with the full IANA TLD list. */
export const linkify = new LinkifyIt();
linkify.tlds(tlds);

/**
 * Extracts unique URLs from text, including bare-domain URLs (e.g. www.hmans.dev).
 * Returns at most maxURLs URLs, in the order they appear.
 * Bare-domain URLs are normalized to https://.
 */
export function extractURLs(text: string, maxURLs = 1): string[] {
  const matches = linkify.match(text);
  if (!matches) return [];

  const seen = new Set<string>();
  const result: string[] = [];

  for (const match of matches) {
    // linkify-it adds http:// to bare domains (schema === '');
    // upgrade those to https:// since it's the safer default.
    // Explicit http:// URLs (schema === 'http://') are kept as-is.
    const url = match.schema === '' ? match.url.replace(/^http:\/\//, 'https://') : match.url;
    const normalized = normalizeURL(url);

    if (!seen.has(normalized)) {
      seen.add(normalized);
      result.push(url);
      if (result.length >= maxURLs) {
        break;
      }
    }
  }

  return result;
}

/**
 * Normalizes a URL for deduplication.
 */
function normalizeURL(url: string): string {
  try {
    const parsed = new URL(url);
    // Lowercase scheme and host, remove fragment
    return `${parsed.protocol.toLowerCase()}//${parsed.host.toLowerCase()}${parsed.pathname}${parsed.search}`;
  } catch {
    return url.toLowerCase();
  }
}

// Valid YouTube hostnames
const YOUTUBE_HOSTS = new Set(['youtube.com', 'www.youtube.com', 'm.youtube.com', 'youtu.be']);

// YouTube path/query patterns (applied after hostname validation)
const YOUTUBE_PATH_REGEX = /^\/(?:watch\?(?:.*&)?v=|embed\/|v\/|shorts\/)([a-zA-Z0-9_-]{11})/;

/**
 * Checks if a URL is a YouTube video URL.
 */
export function isYouTubeURL(url: string): boolean {
  return parseYouTubeVideoID(url) !== null;
}

/**
 * Extracts the video ID from a YouTube URL.
 * Returns null if the URL is not a valid YouTube video URL.
 */
export function parseYouTubeVideoID(rawUrl: string): string | null {
  let parsed: URL;
  try {
    parsed = new URL(rawUrl);
  } catch {
    return null;
  }

  const host = parsed.hostname.toLowerCase();
  if (!YOUTUBE_HOSTS.has(host)) {
    return null;
  }

  // For youtu.be short URLs, the video ID is the path
  if (host === 'youtu.be') {
    const id = parsed.pathname.slice(1); // Remove leading /
    return id.length === 11 ? id : null;
  }

  // For youtube.com, match path/query patterns
  const pathAndQuery = parsed.pathname + parsed.search;
  const match = pathAndQuery.match(YOUTUBE_PATH_REGEX);
  return match ? match[1] : null;
}
