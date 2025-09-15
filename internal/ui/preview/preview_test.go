package preview

import "testing"

func TestIsText(t *testing.T) {
	if !isText([]byte("hello, world\n")) {
		t.Fatalf("ASCII should be text")
	}
	if isText([]byte{'a', 0x00, 'b'}) {
		t.Fatalf("NUL-containing should be non-text")
	}
	// 90% printable should pass
	buf := make([]byte, 100)
	for i := 0; i < 90; i++ {
		buf[i] = 'x'
	}
	for i := 90; i < 100; i++ {
		buf[i] = 1 // control, non-printable
	}
	if !isText(buf) {
		t.Fatalf("85%%+ printable should be text")
	}
}
