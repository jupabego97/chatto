package linkpreview

import (
	"net/url"
	"regexp"
	"strings"
)

// urlRegex matches HTTP/HTTPS URLs in text.
// This is a simplified regex that captures most common URL patterns.
var urlRegex = regexp.MustCompile(`https?://[^\s<>"'` + "`" + `\[\]{}|\\^]+`)

// ExtractURLs extracts unique HTTP/HTTPS URLs from text.
// Returns at most maxURLs URLs, in the order they appear.
func ExtractURLs(text string, maxURLs int) []string {
	if maxURLs <= 0 {
		return nil
	}

	matches := urlRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var result []string

	for _, match := range matches {
		// Clean up trailing punctuation that might have been captured
		match = strings.TrimRight(match, ".,;:!?)")

		// Validate the URL
		u, err := url.Parse(match)
		if err != nil {
			continue
		}

		// Only allow http/https schemes
		if u.Scheme != "http" && u.Scheme != "https" {
			continue
		}

		// Skip if we've already seen this URL
		normalized := normalizeURL(u)
		if seen[normalized] {
			continue
		}
		seen[normalized] = true

		result = append(result, match)
		if len(result) >= maxURLs {
			break
		}
	}

	return result
}

// normalizeURL normalizes a URL for deduplication and caching.
func normalizeURL(u *url.URL) string {
	// Lowercase scheme and host
	normalized := &url.URL{
		Scheme:   strings.ToLower(u.Scheme),
		Host:     strings.ToLower(u.Host),
		Path:     u.Path,
		RawQuery: u.RawQuery,
		Fragment: "", // Ignore fragments
	}
	return normalized.String()
}

// NormalizeURLString normalizes a URL string for caching.
func NormalizeURLString(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return normalizeURL(u)
}

// youtubeHosts is the set of valid YouTube hostnames.
var youtubeHosts = map[string]bool{
	"youtube.com":     true,
	"www.youtube.com": true,
	"m.youtube.com":   true,
	"youtu.be":        true,
}

// youtubePathRegex matches YouTube video path/query patterns.
// Only applied after hostname validation to prevent matching non-YouTube domains.
var youtubePathRegex = regexp.MustCompile(
	`^/(?:watch\?(?:.*&)?v=|embed/|v/|shorts/)([a-zA-Z0-9_-]{11})`,
)

// ParseYouTubeVideoID extracts the video ID from a YouTube URL.
// Returns empty string if the URL is not a valid YouTube video URL.
func ParseYouTubeVideoID(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	host := strings.ToLower(u.Hostname())
	if !youtubeHosts[host] {
		return ""
	}

	// For youtu.be short URLs, the video ID is the path
	if host == "youtu.be" {
		id := strings.TrimPrefix(u.Path, "/")
		if len(id) == 11 {
			return id
		}
		return ""
	}

	// For youtube.com, match path/query patterns
	pathAndQuery := u.Path
	if u.RawQuery != "" {
		pathAndQuery += "?" + u.RawQuery
	}
	matches := youtubePathRegex.FindStringSubmatch(pathAndQuery)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// IsYouTubeURL checks if a URL is a YouTube video URL.
func IsYouTubeURL(rawURL string) bool {
	return ParseYouTubeVideoID(rawURL) != ""
}
