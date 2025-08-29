package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "Zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "Small bytes",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "Kilobytes",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "Megabytes",
			bytes:    1048576,
			expected: "1.0 MB",
		},
		{
			name:     "Gigabytes",
			bytes:    1073741824,
			expected: "1.0 GB",
		},
		{
			name:     "Large value",
			bytes:    52428800, // 50MB
			expected: "50.0 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatPercentage(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{
			name:     "Zero percentage",
			value:    0.0,
			expected: "0.0%",
		},
		{
			name:     "Half percentage",
			value:    50.0,
			expected: "50.0%",
		},
		{
			name:     "Full percentage",
			value:    100.0,
			expected: "100.0%",
		},
		{
			name:     "Decimal percentage",
			value:    75.5,
			expected: "75.5%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPercentage(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "Seconds",
			duration: 30 * time.Second,
			expected: "30s",
		},
		{
			name:     "Minutes",
			duration: 5 * time.Minute,
			expected: "5m",
		},
		{
			name:     "Hours",
			duration: 2 * time.Hour,
			expected: "2h",
		},
		{
			name:     "Days",
			duration: 3 * 24 * time.Hour,
			expected: "3d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDateTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "Today",
			time:     now,
			expected: now.Format("15:04"),
		},
		{
			name:     "This year",
			time:     now.AddDate(0, -1, 0), // 1 month ago
			expected: now.AddDate(0, -1, 0).Format("Jan 2, 15:04"),
		},
		{
			name:     "Different year",
			time:     time.Date(2020, 1, 15, 10, 30, 0, 0, time.UTC),
			expected: "Jan 15, 2020",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDateTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "Just now",
			time:     now.Add(-30 * time.Second),
			expected: "just now",
		},
		{
			name:     "Minutes ago",
			time:     now.Add(-5 * time.Minute),
			expected: "5 minutes ago",
		},
		{
			name:     "Hours ago",
			time:     now.Add(-2 * time.Hour),
			expected: "2 hours ago",
		},
		{
			name:     "Days ago",
			time:     now.AddDate(0, 0, -3),
			expected: "3 days ago",
		},
		{
			name:     "Future time",
			time:     now.Add(5 * time.Minute),
			expected: "in 5 minutes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRelativeTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		expected  string
	}{
		{
			name:      "Short string",
			input:     "Hello",
			maxLength: 10,
			expected:  "Hello",
		},
		{
			name:      "Exact length",
			input:     "Hello World",
			maxLength: 11,
			expected:  "Hello World",
		},
		{
			name:      "Long string",
			input:     "Hello World This Is A Long String",
			maxLength: 15,
			expected:  "Hello World ...",
		},
		{
			name:      "Very short max length",
			input:     "Hello",
			maxLength: 3,
			expected:  "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLength)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "String with control characters",
			input:    "Hello\x00World\x07",
			expected: "HelloWorld",
		},
		{
			name:     "String with newlines",
			input:    "Hello\nWorld\r\n",
			expected: "HelloWorld",
		},
		{
			name:     "String with tabs",
			input:    "Hello\tWorld",
			expected: "HelloWorld",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
