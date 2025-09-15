package keymap

import "testing"

func TestTokensOf(t *testing.T) {
    cases := []struct{
        in   string
        want []string
    }{
        {"g g", []string{"g", "g"}},
        {"ctrl+x ctrl+e", []string{"ctrl+x", "ctrl+e"}},
        {"gg", []string{"g", "g"}},
        {"enter", []string{"enter"}},
        {"G", []string{"G"}},
        {"ctrl+c", []string{"ctrl+c"}},
        {"f12", []string{"f12"}},
        {"up", []string{"up"}},
    }
    for _, tc := range cases {
        got := TokensOf(tc.in)
        if len(got) != len(tc.want) {
            t.Fatalf("TokensOf(%q) len=%d; want %d (%v)", tc.in, len(got), len(tc.want), got)
        }
        for i := range got {
            if got[i] != tc.want[i] {
                t.Fatalf("TokensOf(%q)[%d] = %q; want %q", tc.in, i, got[i], tc.want[i])
            }
        }
    }
}

