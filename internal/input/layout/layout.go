package layout

// Package layout provides keyboard layout normalization for single-key tokens.
// It allows registering multiple language maps that translate a rune produced
// by a non-US layout to the equivalent US-QWERTY key token. This is used so
// bindings like "h/j/k/l" work regardless of active layout.

// Mapper maps a single input rune to a US-QWERTY token (as string).
// The value is string (not rune) to accommodate symbols like "[", "]", ";", "'", etc.
type Mapper map[rune]string

var registry []Mapper

// Register adds a mapper to the global registry. Order matters: earlier maps
// take precedence on conflicts.
func Register(m Mapper) { registry = append(registry, m) }

// NormalizeToken converts a single-key token into its US-QWERTY equivalent if
// found in any registered mapper. Non-letter keys and chord tokens are passed
// through unchanged.
//
// Examples:
//
//	"р" (RU) -> "h"
//	"О" (RU) -> "J"
//	"["       -> "[" (unchanged)
//	"ctrl+x"  -> "ctrl+x" (unchanged)
func NormalizeToken(s string) string {
	// Skip chords or non-single-rune tokens
	if s == "" || len([]rune(s)) != 1 || containsPlus(s) {
		return s
	}
	r := []rune(s)[0]
	for _, m := range registry {
		if out, ok := m[r]; ok {
			return out
		}
	}
	return s
}

func containsPlus(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '+' {
			return true
		}
	}
	return false
}
