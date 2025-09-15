package preview

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// SupportsIterm2 reports whether the environment likely supports iTerm2/WezTerm inline images.
func SupportsIterm2() bool {
	if os.Getenv("TFM_NO_INLINE_IMAGES") != "" {
		return false
	}
	v := strings.ToLower(os.Getenv("TFM_INLINE"))
	switch v {
	case "off", "none", "0", "false":
		return false
	case "iterm2", "wezterm":
		return true
	}
	if os.Getenv("ITERM_SESSION_ID") != "" {
		return true
	}
	if os.Getenv("WEZTERM_PANE") != "" || strings.Contains(strings.ToLower(os.Getenv("TERM_PROGRAM")), "wezterm") {
		return true
	}
	return false
}

// BuildIterm2Inline builds OSC 1337 inline-image sequences for the given file,
// sized in terminal cells (cols x rows). Returns lines to render and true on success.
func BuildIterm2Inline(path string, cols, rows int) ([]string, bool) {
	if cols <= 0 || rows <= 0 {
		return nil, false
	}
	fi, err := os.Stat(path)
	if err != nil || fi.IsDir() {
		return nil, false
	}
	maxBytes := int64(1_572_864)
	if s := os.Getenv("TFM_INLINE_MAX_BYTES"); s != "" {
		if n, err := strconv.ParseInt(s, 10, 64); err == nil && n > 0 {
			maxBytes = n
		}
	}
	if fi.Size() > maxBytes {
		return nil, false
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer f.Close()
	var data []byte
	if fi.Size() > 0 && fi.Size() <= maxBytes {
		data = make([]byte, fi.Size())
		n, _ := io.ReadFull(f, data)
		data = data[:n]
	} else {
		b, _ := io.ReadAll(io.LimitReader(f, maxBytes))
		data = b
	}
	if len(data) == 0 {
		return nil, false
	}
	b64 := base64.StdEncoding.EncodeToString(data)
	name := base64.StdEncoding.EncodeToString([]byte(filepath.Base(path)))
	seq := fmt.Sprintf("\x1b]1337;File=name=%s;size=%d;inline=1;width=%d;height=%d;preserveAspectRatio=1:%s\x07", name, len(data), cols, rows, b64)
	pad := strings.Repeat(" ", cols)
	lines := make([]string, 0, rows)
	lines = append(lines, seq+pad)
	for i := 1; i < rows; i++ {
		lines = append(lines, pad)
	}
	return lines, true
}
