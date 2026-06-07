package renamer

import (
	"testing"
	"time"
)

func TestIsLegacyPattern(t *testing.T) {
	tests := []struct {
		pattern  string
		isLegacy bool
	}{
		{"%Y/%m/%d", false},
		{"%Y/%Y-%M/%Y-%M-%D-%S", true}, // Contains %D, %M, %S -> Legacy (contains %D)
		{"%P%Y%M%D-%h%m%s%S%C%F", true}, // Contains %P, %D, %h, %C, %F
		{"%p%Y%m%d-%H%M%S%s%c%e", false}, // Standard
		{"literal-string", false},
	}

	for _, tt := range tests {
		got := IsLegacyPattern(tt.pattern)
		if got != tt.isLegacy {
			t.Errorf("IsLegacyPattern(%q) = %t; want %t", tt.pattern, got, tt.isLegacy)
		}
	}
}

func TestTranslateLegacyToStandard(t *testing.T) {
	tests := []struct {
		legacy   string
		standard string
	}{
		{"%Y-%M-%D", "%Y-%m-%d"},
		{"%P%Y%M%D-%h%m%s%S%C%F", "%p%Y%m%d-%H%M%S%s%c%e"},
		{"%Y/%Y-%M/%Y-%M-%D%S", "%Y/%Y-%m/%Y-%m-%d%s"},
		{"literal", "literal"},
	}

	for _, tt := range tests {
		got := TranslateLegacyToStandard(tt.legacy)
		if got != tt.standard {
			t.Errorf("TranslateLegacyToStandard(%q) = %q; want %q", tt.legacy, got, tt.standard)
		}
	}
}

func TestTranslatePattern(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"%Y/%Y-%M/%Y-%M-%D%S", "%Y/%Y-%m/%Y-%m-%d%s"},
		{"%p%Y%m%d-%H%M%S%s%c%e", "%p%Y%m%d-%H%M%S%s%c%e"},
	}

	for _, tt := range tests {
		got := TranslatePattern(tt.input)
		if got != tt.expected {
			t.Errorf("TranslatePattern(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFormatPattern(t *testing.T) {
	testTime := time.Date(2023, 8, 3, 19, 40, 5, 0, time.UTC)

	tests := []struct {
		pattern   string
		prefix    string
		suffix    string
		collision string
		ext       string
		expected  string
	}{
		// Legacy Pattern
		{
			pattern:   "%P%Y%M%D-%h%m%s%S%C%F",
			prefix:    "Vacation-",
			suffix:    "-Playa",
			collision: "_1",
			ext:       ".jpg",
			expected:  "Vacation-20230803-194005-Playa_1.jpg",
		},
		// Standard Pattern
		{
			pattern:   "%p%Y%m%d-%H%M%S%s%c%e",
			prefix:    "Vacation-",
			suffix:    "-Playa",
			collision: "_1",
			ext:       ".jpg",
			expected:  "Vacation-20230803-194005-Playa_1.jpg",
		},
		// Subset of components
		{
			pattern:   "%Y/%Y-%m/%Y-%m-%d%s",
			prefix:    "",
			suffix:    "-Playa",
			collision: "",
			ext:       "",
			expected:  "2023/2023-08/2023-08-03-Playa",
		},
		// Literal output
		{
			pattern:   "FixedName",
			prefix:    "Ignore",
			suffix:    "Ignore",
			collision: "",
			ext:       "",
			expected:  "FixedName",
		},
		// 2-digit Year
		{
			pattern:   "%y",
			prefix:    "",
			suffix:    "",
			collision: "",
			ext:       "",
			expected:  "23",
		},
	}

	for _, tt := range tests {
		got := FormatPattern(tt.pattern, testTime, tt.prefix, tt.suffix, tt.collision, tt.ext)
		if got != tt.expected {
			t.Errorf("FormatPattern(%q) = %q; want %q", tt.pattern, got, tt.expected)
		}
	}
}
