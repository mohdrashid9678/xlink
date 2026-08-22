package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestStructuredJSONLoggerOutputsRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{
		Level:  "info",
		Format: "json",
		Output: buf,
	})

	ctx := WithRequestID(context.Background(), "req-12345")
	log.InfoContext(ctx, "test event", slog.String("key", "value"))

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse JSON log output: %v", err)
	}

	if entry["msg"] != "test event" {
		t.Errorf("expected msg 'test event', got %v", entry["msg"])
	}
	if entry["request_id"] != "req-12345" {
		t.Errorf("expected request_id 'req-12345', got %v", entry["request_id"])
	}
	if entry["key"] != "value" {
		t.Errorf("expected key 'value', got %v", entry["key"])
	}
	if entry["level"] != "INFO" {
		t.Errorf("expected level 'INFO', got %v", entry["level"])
	}
}

func TestTextLoggerFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	log := New(Config{
		Level:  "debug",
		Format: "text",
		Output: buf,
	})

	ctx := WithRequestID(context.Background(), "req-debug")
	log.DebugContext(ctx, "debugging message")

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("request_id=req-debug")) {
		t.Errorf("expected text log to contain request_id=req-debug, got %s", output)
	}
}
