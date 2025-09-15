package plugins

import (
	"context"
	"time"
)

// Request is a message sent to a plugin process.
type Request struct {
	Op     string         `json:"op"`
	Path   string         `json:"path,omitempty"`
	Action string         `json:"action,omitempty"`
	Args   map[string]any `json:"args,omitempty"`
}

// Response is a message received from a plugin process.
type Response struct {
	OK      bool     `json:"ok"`
	Message string   `json:"message,omitempty"`
	Kind    string   `json:"kind,omitempty"`
	Content string   `json:"content,omitempty"`
	Mime    string   `json:"mime,omitempty"`
	Actions []string `json:"actions,omitempty"`
}

// Plugin defines the minimal interface for external plugins.
type Plugin interface {
	Name() string
	Run(ctx context.Context, req Request) (Response, error)
	Timeout() time.Duration
}
