package clipboard

// Clipboard is a simple in-memory clipboard for the UI.
// Kind: "file" (list of paths) or "path" (single path).
type Clipboard struct {
	kind  string
	items []string
}

func New() Clipboard { return Clipboard{} }

func (c *Clipboard) Clear() { c.kind, c.items = "", nil }

func (c *Clipboard) SetFiles(paths []string) {
	c.kind = "file"
	c.items = append([]string(nil), paths...)
}

func (c *Clipboard) SetPath(p string) {
	c.kind = "path"
	if p == "" {
		c.items = nil
	} else {
		c.items = []string{p}
	}
}

func (c *Clipboard) Kind() string    { return c.kind }
func (c *Clipboard) Items() []string { return append([]string(nil), c.items...) }
