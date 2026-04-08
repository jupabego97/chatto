import { describe, it, expect } from 'vitest';
import { extractURLs } from './linkPreview';

describe('extractURLs', () => {
  describe('protocol URLs', () => {
    it('extracts https URLs', () => {
      expect(extractURLs('Check https://example.com')).toEqual(['https://example.com']);
    });

    it('extracts http URLs and keeps them as-is', () => {
      expect(extractURLs('Check http://example.com')).toEqual(['http://example.com']);
    });

    it('extracts URLs with paths and query strings', () => {
      expect(extractURLs('See https://example.com/path?q=1')).toEqual([
        'https://example.com/path?q=1'
      ]);
    });
  });

  describe('bare-domain URLs', () => {
    it('detects www-prefixed URLs', () => {
      expect(extractURLs('Visit www.example.com')).toEqual(['https://www.example.com']);
    });

    it('detects bare domains with common TLDs', () => {
      expect(extractURLs('Visit example.com')).toEqual(['https://example.com']);
    });

    it('detects bare domains with newer TLDs like .dev', () => {
      expect(extractURLs('check www.hmans.dev')).toEqual(['https://www.hmans.dev']);
    });

    it('detects bare domains with .io TLD', () => {
      expect(extractURLs('try app.example.io')).toEqual(['https://app.example.io']);
    });

    it('detects bare domains with .app TLD', () => {
      expect(extractURLs('see myapp.app')).toEqual(['https://myapp.app']);
    });

    it('detects bare domains with paths', () => {
      expect(extractURLs('read www.hmans.dev/blog/chatto')).toEqual([
        'https://www.hmans.dev/blog/chatto'
      ]);
    });

    it('normalizes bare domains to https://', () => {
      const urls = extractURLs('www.example.com');
      expect(urls[0]).toMatch(/^https:\/\//);
    });
  });

  describe('deduplication and limits', () => {
    it('returns at most maxURLs results', () => {
      expect(extractURLs('https://a.com https://b.com https://c.com', 2)).toHaveLength(2);
    });

    it('defaults to 1 URL', () => {
      expect(extractURLs('https://a.com https://b.com')).toHaveLength(1);
    });

    it('deduplicates identical URLs', () => {
      expect(extractURLs('https://example.com https://example.com', 5)).toEqual([
        'https://example.com'
      ]);
    });

    it('deduplicates bare and protocol URLs pointing to the same host', () => {
      const urls = extractURLs('www.example.com and https://www.example.com', 5);
      expect(urls).toHaveLength(1);
    });
  });

  describe('edge cases', () => {
    it('returns empty array for text with no URLs', () => {
      expect(extractURLs('no urls here')).toEqual([]);
    });

    it('returns empty array for empty string', () => {
      expect(extractURLs('')).toEqual([]);
    });

    it('handles URL at start of text', () => {
      expect(extractURLs('https://example.com is cool')).toEqual(['https://example.com']);
    });

    it('handles URL at end of text', () => {
      expect(extractURLs('check out https://example.com')).toEqual(['https://example.com']);
    });
  });
});
