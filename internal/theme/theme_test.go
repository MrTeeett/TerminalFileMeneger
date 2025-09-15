package theme

import "testing"

func TestDefaultTheme(t *testing.T) {
	th := Default()
	if th.Name != "default" {
		t.Fatalf("Name = %q; want default", th.Name)
	}
	if th.Primary == "" || th.Foreground == "" || th.Selection == "" {
		t.Fatalf("default theme fields should be non-empty: %#v", th)
	}
}
