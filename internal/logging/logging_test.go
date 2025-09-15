package logging

import (
    "bytes"
    "strings"
    "testing"
)

func TestLoggerLevels(t *testing.T) {
    var b bytes.Buffer

    // Debug level: everything prints
    lg := NewWithWriter("debug", &b)
    lg.Debugf("d %d", 1)
    lg.Infof("i")
    lg.Warnf("w")
    lg.Errorf("e")
    out := b.String()
    if !strings.Contains(out, "[DEBUG] d 1") || !strings.Contains(out, "[INFO] i") || !strings.Contains(out, "[WARN] w") || !strings.Contains(out, "[ERROR] e") {
        t.Fatalf("unexpected output for debug level: %q", out)
    }

    // Info level: no debug
    b.Reset()
    lg = NewWithWriter("info", &b)
    lg.Debugf("should not appear")
    lg.Infof("hello")
    out = b.String()
    if strings.Contains(out, "should not appear") {
        t.Fatalf("debug line printed at info level: %q", out)
    }
    if !strings.Contains(out, "[INFO] hello") {
        t.Fatalf("info line missing: %q", out)
    }

    // Warn level: only warn+error
    b.Reset()
    lg = NewWithWriter("warn", &b)
    lg.Infof("nope")
    lg.Warnf("warned")
    lg.Errorf("boom")
    out = b.String()
    if strings.Contains(out, "nope") {
        t.Fatalf("info printed at warn level")
    }
    if !strings.Contains(out, "[WARN] warned") || !strings.Contains(out, "[ERROR] boom") {
        t.Fatalf("warn/error missing: %q", out)
    }

    // Error level: only error
    b.Reset()
    lg = NewWithWriter("error", &b)
    lg.Warnf("nope")
    lg.Errorf("only")
    out = b.String()
    if strings.Contains(out, "nope") || !strings.Contains(out, "[ERROR] only") {
        t.Fatalf("error-only expected: %q", out)
    }
}

