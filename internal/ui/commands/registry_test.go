package commands

import (
    "context"
    "errors"
    "testing"
)

func TestRegistry(t *testing.T) {
    r := NewRegistry()
    ran := false
    r.Register(Command{
        Name:        "ping",
        Description: "test",
        Handler: func(ctx context.Context, args Args) error {
            ran = true
            if args["x"] != 42 {
                return errors.New("bad arg")
            }
            return nil
        },
    })
    if err := r.Run(context.Background(), "ping", Args{"x": 42}); err != nil {
        t.Fatalf("run ping: %v", err)
    }
    if !ran {
        t.Fatalf("handler did not run")
    }
    if err := r.Run(context.Background(), "nope", nil); err == nil {
        t.Fatalf("expected error for unknown command")
    }
}

