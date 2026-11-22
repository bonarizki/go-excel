package importer

import (
	"context"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/bonarizki/go-excel/excel"
)

// newSingleCellFile builds a minimal one-cell workbook for timeout tests,
// which only care about context propagation and not about row content.
func newSingleCellFile(t *testing.T) *excelize.File {
	t.Helper()
	f := excelize.NewFile()
	if err := f.SetCellValue("Sheet1", "A1", "x"); err != nil {
		t.Fatalf("SetCellValue: %v", err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatalf("WriteToBuffer: %v", err)
	}
	opened, err := excelize.OpenReader(buf)
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	return opened
}

// TestImporter_Timeout_SetsContextDeadline verifies that WithTimeout wraps
// the context passed through Import with a deadline, since it was
// previously accepted into Config but never actually applied.
func TestImporter_Timeout_SetsContextDeadline(t *testing.T) {
	imp := NewImporterWithFile(newSingleCellFile(t), excel.WithTimeout(50*time.Millisecond), excel.WithHeadings(false))

	var sawDeadline bool
	imp.On(excel.BeforeImport, func(ctx context.Context, _ *excel.Event) error {
		_, sawDeadline = ctx.Deadline()
		return nil
	})

	if err := imp.Import(context.Background()); err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if !sawDeadline {
		t.Error("expected context to have a deadline when Timeout is set")
	}
}

// TestImporter_NoTimeout_NoContextDeadline verifies that Import does not
// impose a deadline when Timeout is left at its zero value.
func TestImporter_NoTimeout_NoContextDeadline(t *testing.T) {
	imp := NewImporterWithFile(newSingleCellFile(t), excel.WithHeadings(false))

	var hasDeadline bool
	imp.On(excel.BeforeImport, func(ctx context.Context, _ *excel.Event) error {
		_, hasDeadline = ctx.Deadline()
		return nil
	})

	if err := imp.Import(context.Background()); err != nil {
		t.Fatalf("Import failed: %v", err)
	}
	if hasDeadline {
		t.Error("expected no deadline on context when Timeout is not set")
	}
}

// TestImporter_Timeout_Expires verifies that an elapsed WithTimeout
// causes Import to fail with a context error rather than succeeding.
//
// The timeout is applied when Import starts (not at construction), so a
// pre-Import sleep cannot expire it. Waiting on ctx.Done() inside
// BeforeImport is deterministic on Windows, where a 1ns timer often does
// not fire before a tiny in-memory import finishes.
func TestImporter_Timeout_Expires(t *testing.T) {
	imp := NewImporterWithFile(newSingleCellFile(t), excel.WithTimeout(10*time.Millisecond), excel.WithHeadings(false))

	imp.On(excel.BeforeImport, func(ctx context.Context, _ *excel.Event) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	})

	err := imp.Import(context.Background())
	if err == nil {
		t.Fatal("expected Import to fail once the configured timeout has elapsed")
	}
}
