package logger

import (
	"bytes"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestNew_DevWritesText(t *testing.T) {
	var out bytes.Buffer
	lg, _, err := New("development", "", &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	lg.Info("hello", slog.String("key", "val"))
	if !strings.Contains(out.String(), "key=val") {
		t.Errorf("expected text format, got: %s", out.String())
	}
}

func TestNew_ProdWritesJSON(t *testing.T) {
	var out bytes.Buffer
	lg, _, err := New("production", "", &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	lg.Info("hello", slog.String("key", "val"))
	if !strings.Contains(out.String(), `"key":"val"`) {
		t.Errorf("expected JSON format, got: %s", out.String())
	}
}

func TestNew_ErrorsGoToStderr(t *testing.T) {
	var out, errOut bytes.Buffer
	lg, _, err := New("development", "", &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	lg.Error("boom", slog.String("key", "val"))
	if !strings.Contains(errOut.String(), "boom") {
		t.Errorf("expected error in stderr, got: %s", errOut.String())
	}
}

func TestNew_InfoDoesNotGoToStderr(t *testing.T) {
	var out, errOut bytes.Buffer
	lg, _, err := New("development", "", &out, &errOut)
	if err != nil {
		t.Fatal(err)
	}
	lg.Info("routine message")
	if errOut.Len() != 0 {
		t.Errorf("expected stderr to be empty for info log, got: %s", errOut.String())
	}
}

func TestNew_DevDebugVisible(t *testing.T) {
	var out bytes.Buffer
	lg, _, err := New("development", "", &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	lg.Debug("verbose")
	if !strings.Contains(out.String(), "verbose") {
		t.Errorf("expected debug log in dev output, got: %s", out.String())
	}
}

func TestNew_ProdDebugHidden(t *testing.T) {
	var out bytes.Buffer
	lg, _, err := New("production", "", &out, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	lg.Debug("verbose")
	if out.Len() != 0 {
		t.Errorf("expected debug log hidden in prod, got: %s", out.String())
	}
}
