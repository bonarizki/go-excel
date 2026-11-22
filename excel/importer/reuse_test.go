package importer

import (
	"context"
	"testing"

	"github.com/bonarizki/go-excel/excel"
)

// TestImporter_Reuse_DoesNotCarryOverRows verifies a second import on the
// same importer reports only the second file's rows and headings, with
// and without preallocation.
func TestImporter_Reuse_DoesNotCarryOverRows(t *testing.T) {
	for _, prealloc := range []bool{true, false} {
		first := writeCSVFile(t, "id,name\n1,Alice\n2,Bob\n3,Carol\n")
		second := writeCSVFile(t, "code\nX\n")

		imp := NewImporter(excel.WithHeadings(true), excel.WithPreallocation(prealloc))
		if err := imp.ImportFromFile(context.Background(), first); err != nil {
			t.Fatalf("first import: %v", err)
		}
		if err := imp.ImportFromFile(context.Background(), second); err != nil {
			t.Fatalf("second import: %v", err)
		}

		if got := len(imp.GetRows()); got != 1 {
			t.Errorf("prealloc=%v: len(GetRows()) = %d, want 1", prealloc, got)
		}
		if got := imp.GetHeadings(); len(got) != 1 || got[0] != "code" {
			t.Errorf("prealloc=%v: headings = %v, want [code]", prealloc, got)
		}
		_ = imp.Close()
	}
}

// TestImportFromFile_CSVLazyQuotes verifies a bare quote is rejected by
// default and accepted with WithCSVLazyQuotes(true).
func TestImportFromFile_CSVLazyQuotes(t *testing.T) {
	path := writeCSVFile(t, "name\nsay \"hi\" now\n")

	strict := NewImporter(excel.WithHeadings(true))
	if err := strict.ImportFromFile(context.Background(), path); err == nil {
		t.Error("default importer accepted a bare quote, want error")
	}
	_ = strict.Close()

	lazy := NewImporter(excel.WithHeadings(true), excel.WithCSVLazyQuotes(true))
	defer func() { _ = lazy.Close() }()
	if err := lazy.ImportFromFile(context.Background(), path); err != nil {
		t.Fatalf("lazy importer: %v", err)
	}
	if got := len(lazy.GetRows()); got != 1 {
		t.Errorf("len(GetRows()) = %d, want 1", got)
	}
}
