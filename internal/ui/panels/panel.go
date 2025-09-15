package panels

import (
	"os"
	"path/filepath"
	"sort"
	"unicode/utf8"
)

// Entry represents a file or directory in the panel.
type Entry struct {
	Name  string
	IsDir bool
}

// Panel holds current directory state and entries.
type Panel struct {
	Cwd        string
	Entries    []Entry
	ShowHidden bool
	// MaxDirName stores the maximum rune-length of directory names in Entries (plus slash), for layout hints.
	MaxDirName int
}

func NewPanel(cwd string, showHidden bool) *Panel {
	return &Panel{Cwd: cwd, ShowHidden: showHidden}
}

// Refresh reads the current directory and updates entries.
func (p *Panel) Refresh() error {
	if p.Cwd == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		p.Cwd = cwd
	}
	ents, err := os.ReadDir(p.Cwd)
	if err != nil {
		return err
	}
	p.Entries = p.Entries[:0]
	p.MaxDirName = 0
	for _, e := range ents {
		name := e.Name()
		if !p.ShowHidden && len(name) > 0 && name[0] == '.' {
			continue
		}
		isDir := e.IsDir()
		p.Entries = append(p.Entries, Entry{Name: name, IsDir: isDir})
		if isDir {
			l := utf8.RuneCountInString(name) + 1 // account for '/'
			if l > p.MaxDirName {
				p.MaxDirName = l
			}
		}
	}
	sort.Slice(p.Entries, func(i, j int) bool {
		if p.Entries[i].IsDir != p.Entries[j].IsDir {
			return p.Entries[i].IsDir
		}
		return p.Entries[i].Name < p.Entries[j].Name
	})
	return nil
}

// Join returns a path within the panel's CWD.
func (p *Panel) Join(name string) string {
	return filepath.Join(p.Cwd, name)
}

// Chdir attempts to change current directory to dir and refresh entries.
// It only updates p.Cwd on success.
func (p *Panel) Chdir(dir string) error {
	if dir == "" {
		return os.ErrInvalid
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	// Build a new list based on ShowHidden
	next := make([]Entry, 0, len(ents))
	maxDir := 0
	for _, e := range ents {
		name := e.Name()
		if !p.ShowHidden && len(name) > 0 && name[0] == '.' {
			continue
		}
		isDir := e.IsDir()
		next = append(next, Entry{Name: name, IsDir: isDir})
		if isDir {
			l := utf8.RuneCountInString(name) + 1
			if l > maxDir {
				maxDir = l
			}
		}
	}
	sort.Slice(next, func(i, j int) bool {
		if next[i].IsDir != next[j].IsDir {
			return next[i].IsDir
		}
		return next[i].Name < next[j].Name
	})
	p.Cwd = dir
	p.Entries = next
	p.MaxDirName = maxDir
	return nil
}
