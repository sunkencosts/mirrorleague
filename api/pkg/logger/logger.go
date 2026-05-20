package logger

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type CloseFunc func() error

// multiHandler fans out log records to multiple slog.Handler implementations.
type multiHandler struct {
	handlers []slog.Handler
}

func (h *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, hh := range h.handlers {
		if hh.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, hh := range h.handlers {
		if hh.Enabled(ctx, r.Level) {
			if err := hh.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, len(h.handlers))
	for i, hh := range h.handlers {
		next[i] = hh.WithAttrs(attrs)
	}
	return &multiHandler{next}
}

func (h *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, len(h.handlers))
	for i, hh := range h.handlers {
		next[i] = hh.WithGroup(name)
	}
	return &multiHandler{next}
}

// New returns a logger configured for the given environment.
//
// development: text format to stdout + logFile (if set); errors also to stderr.
// other envs:  JSON format to stdout (info+); errors also to stderr.
//
// stdout/stderr are passed in so callers (e.g. tests) can pass io.Discard to suppress output.
func New(env, logFile string, stdout, stderr io.Writer) (*slog.Logger, CloseFunc, error) {
	noop := func() error { return nil }

	if env == "development" {
		dest := stdout
		cleanup := noop

		if logFile != "" {
			f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
			if err != nil {
				return nil, nil, fmt.Errorf("opening log file: %w", err)
			}
			buf := bufio.NewWriterSize(f, 8192)
			dest = io.MultiWriter(stdout, buf)
			cleanup = func() error {
				if err := buf.Flush(); err != nil {
					return err
				}
				if err := f.Sync(); err != nil {
					return err
				}
				return f.Close()
			}
		}

		main := slog.NewTextHandler(dest, &slog.HandlerOptions{Level: slog.LevelDebug})
		errs := slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: slog.LevelError})
		return slog.New(&multiHandler{[]slog.Handler{main, errs}}), cleanup, nil
	}

	main := slog.NewJSONHandler(stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	errs := slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: slog.LevelError})
	return slog.New(&multiHandler{[]slog.Handler{main, errs}}), noop, nil
}
