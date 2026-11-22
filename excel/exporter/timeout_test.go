package exporter

import (
	"context"
	"testing"
	"time"

	"github.com/bonarizki/go-excel/excel"
)

// TestExporter_Timeout_SetsContextDeadline verifies that WithTimeout wraps
// the context passed through Export with a deadline, since it was
// previously accepted into Config but never actually applied.
func TestExporter_Timeout_SetsContextDeadline(t *testing.T) {
	exp := NewExporter([][]any{{"a"}}, excel.WithTimeout(50*time.Millisecond))

	var sawDeadline bool
	exp.On(excel.BeforeExport, func(ctx context.Context, _ *excel.Event) error {
		_, sawDeadline = ctx.Deadline()
		return nil
	})

	if err := exp.Export(context.Background()); err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if !sawDeadline {
		t.Error("expected context to have a deadline when Timeout is set")
	}
}

// TestExporter_NoTimeout_NoContextDeadline verifies that Export does not
// impose a deadline when Timeout is left at its zero value.
func TestExporter_NoTimeout_NoContextDeadline(t *testing.T) {
	exp := NewExporter([][]any{{"a"}})

	var hasDeadline bool
	exp.On(excel.BeforeExport, func(ctx context.Context, _ *excel.Event) error {
		_, hasDeadline = ctx.Deadline()
		return nil
	})

	if err := exp.Export(context.Background()); err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if hasDeadline {
		t.Error("expected no deadline on context when Timeout is not set")
	}
}

// TestExporter_Timeout_Expires verifies that an elapsed WithTimeout
// causes Export to fail with a context error rather than succeeding.
//
// The timeout is applied when Export starts (not at construction), so a
// pre-Export sleep cannot expire it. Waiting on ctx.Done() inside
// BeforeExport is deterministic on Windows, where a 1ns timer often does
// not fire before a tiny export finishes.
func TestExporter_Timeout_Expires(t *testing.T) {
	exp := NewExporter([][]any{{"a"}, {"b"}}, excel.WithTimeout(10*time.Millisecond))

	exp.On(excel.BeforeExport, func(ctx context.Context, _ *excel.Event) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	})

	if err := exp.Export(context.Background()); err == nil {
		t.Fatal("expected Export to fail once the configured timeout has elapsed")
	}
}
