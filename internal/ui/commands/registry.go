package commands

import (
	"context"
	"fmt"
)

type Args map[string]any

type Handler func(context.Context, Args) error

type Command struct {
	Name        string
	Description string
	Handler     Handler
}

type Registry struct {
	m map[string]Command
}

func NewRegistry() Registry {
	return Registry{m: make(map[string]Command)}
}

func (r *Registry) Register(cmd Command) {
	r.m[cmd.Name] = cmd
}

func (r *Registry) Run(ctx context.Context, name string, args Args) error {
	if c, ok := r.m[name]; ok && c.Handler != nil {
		return c.Handler(ctx, args)
	}
	return fmt.Errorf("command not found: %s", name)
}
