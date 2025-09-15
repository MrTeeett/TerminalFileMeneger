package plugins

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

type dummy struct{}

func (dummy) Name() string           { return "dummy" }
func (dummy) Timeout() time.Duration { return 123 * time.Millisecond }
func (dummy) Run(ctx context.Context, req Request) (Response, error) {
	_ = ctx
	return Response{OK: true, Message: req.Op, Kind: "text", Content: "ok", Mime: "text/plain", Actions: []string{"a"}}, nil
}

func TestRequestResponseJSON(t *testing.T) {
	r := Request{Op: "do", Path: "/tmp/x", Action: "open", Args: map[string]any{"x": 1}}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal Request: %v", err)
	}
	var r2 Request
	if err := json.Unmarshal(b, &r2); err != nil {
		t.Fatalf("unmarshal Request: %v", err)
	}
	if r2.Op != r.Op || r2.Path != r.Path || r2.Action != r.Action || r2.Args["x"].(float64) != 1 {
		t.Fatalf("roundtrip mismatch: %#v", r2)
	}

	resp := Response{OK: true, Message: "m", Kind: "k", Content: "c", Mime: "text/plain", Actions: []string{"x"}}
	b, err = json.Marshal(resp)
	if err != nil {
		t.Fatalf("marshal Response: %v", err)
	}
	var resp2 Response
	if err := json.Unmarshal(b, &resp2); err != nil {
		t.Fatalf("unmarshal Response: %v", err)
	}
	if !resp2.OK || resp2.Message != "m" || resp2.Kind != "k" || resp2.Content != "c" || resp2.Mime != "text/plain" || len(resp2.Actions) != 1 || resp2.Actions[0] != "x" {
		t.Fatalf("roundtrip resp mismatch: %#v", resp2)
	}
}

func TestPluginInterface(t *testing.T) {
	var p Plugin = dummy{}
	if p.Name() != "dummy" {
		t.Fatalf("name")
	}
	if p.Timeout() != 123*time.Millisecond {
		t.Fatalf("timeout")
	}
	out, err := p.Run(context.Background(), Request{Op: "hello"})
	if err != nil || !out.OK || out.Message != "hello" {
		t.Fatalf("run: out=%#v err=%v", out, err)
	}
}
