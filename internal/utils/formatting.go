// Package utils provides utility functions for the Arcanus VPN bot.
// It includes formatting helpers and common utilities used across the application.
package utils

import (
	"fmt"
	"time"
)

// FormatBytes formats bytes into human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatPercentage formats a float64 as a percentage
func FormatPercentage(value float64) string {
	return fmt.Sprintf("%.1f%%", value)
}

// FormatDuration formats a duration in a human-readable format
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0fm", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	return fmt.Sprintf("%.0fd", d.Hours()/24)
}

// FormatDateTime formats a time.Time in a user-friendly format
func FormatDateTime(t time.Time) string {
	now := time.Now()

	// If it's today, show time only
	if t.Year() == now.Year() && t.YearDay() == now.YearDay() {
		return t.Format("15:04")
	}

	// If it's this year, show date and time
	if t.Year() == now.Year() {
		return t.Format("Jan 2, 15:04")
	}

	// Otherwise show full date
	return t.Format("Jan 2, 2006")
}

// FormatRelativeTime formats a time relative to now
func FormatRelativeTime(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < 0 {
		duration = -duration
		if duration < time.Minute {
			return "in a few seconds"
		}
		if duration < time.Hour {
			return fmt.Sprintf("in %.0f minutes", duration.Minutes())
		}
		if duration < 24*time.Hour {
			return fmt.Sprintf("in %.0f hours", duration.Hours())
		}
		return fmt.Sprintf("in %.0f days", duration.Hours()/24)
	}

	if duration < time.Minute {
		return "just now"
	}
	if duration < time.Hour {
		return fmt.Sprintf("%.0f minutes ago", duration.Minutes())
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%.0f hours ago", duration.Hours())
	}
	if duration < 7*24*time.Hour {
		return fmt.Sprintf("%.0f days ago", duration.Hours()/24)
	}
	return t.Format("Jan 2, 2006")
}

// TruncateString truncates a string to the specified length and adds ellipsis
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	if maxLength <= 3 {
		return "..."
	}
	return s[:maxLength-3] + "..."
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(s string) string {
	// Remove control characters and other potentially dangerous characters
	var result []rune
	for _, r := range s {
		if r >= 32 && r != 127 {
			result = append(result, r)
		}
	}
	return string(result)
}
