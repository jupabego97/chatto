package linkpreview

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractURLs(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxURLs  int
		expected []string
	}{
		{"no URLs", "hello world", 5, nil},
		{"single URL", "check out https://example.com please", 5, []string{"https://example.com"}},
		{"multiple URLs", "see https://a.com and https://b.com", 5, []string{"https://a.com", "https://b.com"}},
		{"respects maxURLs", "see https://a.com and https://b.com", 1, []string{"https://a.com"}},
		{"deduplicates", "see https://example.com and https://example.com again", 5, []string{"https://example.com"}},
		{"strips trailing punctuation", "visit https://example.com.", 5, []string{"https://example.com"}},
		{"http scheme", "visit http://example.com", 5, []string{"http://example.com"}},
		{"zero maxURLs", "https://example.com", 0, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractURLs(tt.text, tt.maxURLs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseYouTubeVideoID(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		// Valid YouTube URLs
		{"watch URL", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"watch URL with params", "https://www.youtube.com/watch?feature=share&v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"embed URL", "https://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"short URL", "https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"shorts URL", "https://www.youtube.com/shorts/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"v/ URL", "https://www.youtube.com/v/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"mobile URL", "https://m.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"no www", "https://youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},

		// Non-YouTube URLs that should NOT match (hostname-anchored)
		{"not youtube domain", "https://notyoutube.com/watch?v=dQw4w9WgXcQ", ""},
		{"youtube in path", "https://evil.com/redirect?to=youtube.com/watch?v=dQw4w9WgXcQ", ""},
		{"youtube in subdomain", "https://fakeyoutube.com/watch?v=dQw4w9WgXcQ", ""},
		{"evil redirect", "https://evil.com/youtube.com/watch?v=dQw4w9WgXcQ", ""},

		// Invalid URLs
		{"empty string", "", ""},
		{"not a URL", "not-a-url", ""},
		{"wrong ID length", "https://youtu.be/short", ""},
		{"no video ID", "https://www.youtube.com/watch", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseYouTubeVideoID(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsYouTubeURL(t *testing.T) {
	assert.True(t, IsYouTubeURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ"))
	assert.True(t, IsYouTubeURL("https://youtu.be/dQw4w9WgXcQ"))
	assert.False(t, IsYouTubeURL("https://example.com"))
	assert.False(t, IsYouTubeURL("https://notyoutube.com/watch?v=dQw4w9WgXcQ"))
}
