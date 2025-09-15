package layout

import "testing"

func TestNormalizeTokenRUUA(t *testing.T) {
    // RU: 'р' -> 'h', 'О' -> 'J'
    if got := NormalizeToken("р"); got != "h" {
        t.Fatalf("NormalizeToken(р) = %q; want h", got)
    }
    if got := NormalizeToken("О"); got != "J" {
        t.Fatalf("NormalizeToken(О) = %q; want J", got)
    }
    // UA: 'є' -> "'"
    if got := NormalizeToken("є"); got != "'" {
        t.Fatalf("NormalizeToken(є) = %q; want ' ", got)
    }
    // Non-single or chord remains unchanged
    if got := NormalizeToken("ctrl+x"); got != "ctrl+x" {
        t.Fatalf("chord changed: %q", got)
    }
    // Unknown stays the same
    if got := NormalizeToken("?"); got != "?" {
        t.Fatalf("unknown changed: %q", got)
    }
}

