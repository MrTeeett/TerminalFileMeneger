package preview

import (
	"bufio"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"
	"time"
)

// Result is a generic preview result.
type Result struct {
	Kind    string // e.g., "text", "info"
	Content string
	Mime    string
}

// Provider renders preview for a given path.
type Provider interface {
	Preview(path string, maxBytes int) (Result, error)
}

// BasicProvider shows text snippets for text-like files and metadata otherwise.
type BasicProvider struct{}

func (BasicProvider) Preview(path string, maxBytes int) (Result, error) {
	if maxBytes <= 0 {
		maxBytes = 8192
	}
	fi, err := os.Stat(path)
	if err != nil {
		return Result{}, err
	}
	if fi.IsDir() {
		return Result{Kind: "info", Content: "directory", Mime: "inode/directory"}, nil
	}
	// Try to detect image quickly using DecodeConfig
	f, err := os.Open(path)
	if err != nil {
		return Result{}, err
	}
	defer f.Close()
	cfg, format, cfgErr := image.DecodeConfig(f)
	if cfgErr == nil && (format == "png" || format == "jpeg" || format == "gif") {
		// Rewind and decode full image (best-effort, may be heavy for huge files)
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			// Seek failed (unlikely) — reopen
			f.Close()
			f, err = os.Open(path)
			if err != nil {
				return Result{}, err
			}
		}
		img, _, err := image.Decode(f)
		if err == nil {
			// Produce ASCII thumbnail preview
			targetW := 64
			if cfg.Width > 0 && cfg.Width < targetW {
				targetW = cfg.Width
			}
			ascii := asciiPreview(img, targetW)
			mime := "image/" + format
			return Result{Kind: "image", Content: ascii, Mime: mime}, nil
		}
		// If decode failed, fall through to text path: rewind for buffered read
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			// Best effort: reopen
			f.Close()
			if rf, er2 := os.Open(path); er2 == nil {
				f = rf
				defer f.Close()
			}
		}
	} else {
		// Not an image (or unknown): rewind before reading bytes
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			f.Close()
			f, _ = os.Open(path)
		}
	}
	// Try read a small chunk and detect as text
	r := bufio.NewReader(f)
	buf := make([]byte, 0, maxBytes)
	for len(buf) < maxBytes {
		b := make([]byte, 1024)
		n, er := r.Read(b)
		if n > 0 {
			if len(buf)+n > maxBytes {
				n = maxBytes - len(buf)
			}
			buf = append(buf, b[:n]...)
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			break
		}
	}
	if isText(buf) {
		s := string(buf)
		s = strings.ReplaceAll(s, "\r\n", "\n")
		s = strings.ReplaceAll(s, "\r", "\n")
		lines := strings.Split(s, "\n")
		if len(lines) > 256 {
			lines = lines[:256]
			s = strings.Join(lines, "\n") + "\n…"
		} else {
			s = strings.Join(lines, "\n")
		}
		return Result{Kind: "text", Content: s, Mime: "text/plain"}, nil
	}
	// Fallback to metadata
	info := fmt.Sprintf("%s (%d bytes)\nmode: %s\nmodified: %s",
		fi.Name(), fi.Size(), fi.Mode(), fi.ModTime().Format(time.RFC3339))
	return Result{Kind: "info", Content: info, Mime: "application/octet-stream"}, nil
}

func isText(b []byte) bool {
	if len(b) == 0 {
		return true
	}
	// Reject if contains NUL
	for _, c := range b {
		if c == 0x00 {
			return false
		}
	}
	// Count printable or allowed whitespace/control
	printable := 0
	for _, c := range b {
		if c == '\n' || c == '\r' || c == '\t' || c == '\f' || c == '\v' {
			printable++
			continue
		}
		if c >= 0x20 && c <= 0x7E { // ASCII printable
			printable++
		}
	}
	// If at least 85% printable, call it text
	return float64(printable) >= 0.85*float64(len(b))
}

// asciiPreview builds a simple grayscale ASCII preview of the image limited
// to targetW characters wide. Height is adjusted to approximate terminal
// aspect ratio.
func asciiPreview(img image.Image, targetW int) string {
	if targetW < 8 {
		targetW = 8
	}
	b := img.Bounds()
	iw, ih := b.Dx(), b.Dy()
	if iw <= 0 || ih <= 0 {
		return ""
	}
	// Keep aspect ratio. Terminal chars are ~2x taller than wide; reduce rows.
	targetH := ih * targetW / iw
	if targetH < 1 {
		targetH = 1
	}
	targetH = targetH / 2
	if targetH < 1 {
		targetH = 1
	}
	if targetH > 80 {
		targetH = 80
	}

	ramp := []rune(" .:-=+*#%@")
	nr := len(ramp)

	var out strings.Builder
	for y := 0; y < targetH; y++ {
		sy := b.Min.Y + y*ih/targetH
		for x := 0; x < targetW; x++ {
			sx := b.Min.X + x*iw/targetW
			r, g, bl, _ := img.At(sx, sy).RGBA()
			// Convert to 0..255 luma
			// Using Rec. 601 luma coefficients
			y8 := (299*r + 587*g + 114*bl) / 1000
			y8 = y8 >> 8
			idx := int((y8 * uint32(nr-1)) / 255)
			out.WriteRune(ramp[idx])
		}
		if y < targetH-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}
