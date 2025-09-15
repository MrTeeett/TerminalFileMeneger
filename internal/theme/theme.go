package theme

// Theme holds style names; concrete styling will be assigned later (e.g., Lip Gloss).
type Theme struct {
	Name       string
	Primary    string
	Secondary  string
	Accent     string
	Error      string
	Warning    string
	Background string
	Foreground string
	Selection  string
}

func Default() Theme {
	return Theme{
		Name:       "default",
		Primary:    "blue",
		Secondary:  "magenta",
		Accent:     "cyan",
		Error:      "red",
		Warning:    "yellow",
		Background: "default",
		Foreground: "default",
		Selection:  "reverse",
	}
}
