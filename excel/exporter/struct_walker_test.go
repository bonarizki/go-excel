package exporter_test

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/bonarizki/go-excel/excel"
	"github.com/bonarizki/go-excel/excel/exporter"
	"github.com/bonarizki/go-excel/excel/importer"
)

type roundTripUser struct {
	ID    int    `excel:"id"`
	Name  string `excel:"full_name"`
	Email string `excel:"email"`
}

// TestExport_StructRoundTrip verifies exporting []T with headings writes
// the tag names as the header and keeps every record, so the file imports
// back to the same slice.
func TestExport_StructRoundTrip(t *testing.T) {
	users := []roundTripUser{
		{ID: 1, Name: "Ada", Email: "ada@x.io"},
		{ID: 2, Name: "Linus", Email: "linus@x.io"},
	}

	for ext, format := range map[string]excel.FileFormat{"xlsx": excel.FileFormatXLSX, "csv": excel.FileFormatCSV} {
		exp := exporter.NewExporter(users, excel.WithHeadings(true), excel.WithFormat(format))
		var buf bytes.Buffer
		if err := exp.ExportToWriter(context.Background(), &buf); err != nil {
			t.Fatalf("%s: ExportToWriter: %v", ext, err)
		}
		_ = exp.Close()

		path := filepath.Join(t.TempDir(), "users."+ext)
		if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}

		got, err := importer.ImportToStructs[roundTripUser](context.Background(), path, excel.WithHeadings(true))
		if err != nil {
			t.Fatalf("%s: ImportToStructs: %v", ext, err)
		}
		if !reflect.DeepEqual(got, users) {
			t.Errorf("%s: round trip = %+v, want %+v", ext, got, users)
		}
	}
}

type ptrRow struct {
	Count *int           `excel:"count"`
	Note  *string        `excel:"note"`
	Score sql.NullString `excel:"score"`
}

// TestExport_PointerValuesAreDereferenced verifies pointer and
// driver.Valuer fields export their value, not a memory address.
func TestExport_PointerValuesAreDereferenced(t *testing.T) {
	n, s := 7, "hi"
	rows := []ptrRow{
		{Count: &n, Note: &s, Score: sql.NullString{String: "A", Valid: true}},
		{},
	}

	exp := exporter.NewExporter(rows, excel.WithHeadings(true), excel.WithFormat(excel.FileFormatCSV))
	defer func() { _ = exp.Close() }()
	var buf bytes.Buffer
	if err := exp.ExportToWriter(context.Background(), &buf); err != nil {
		t.Fatalf("ExportToWriter: %v", err)
	}

	want := "\ufeffcount,note,score\r\n7,hi,A\r\n,,\r\n"
	if got := buf.String(); got != want {
		t.Errorf("CSV = %q, want %q", got, want)
	}
}

type EmbeddedBase struct {
	A int `excel:"a"`
	B int `excel:"b"`
}

type embeddedRow struct {
	*EmbeddedBase
	C    string `excel:"c"`
	Zero int    `excel:"omitempty"`
}

// TestExport_NilEmbeddedPointerKeepsColumns verifies a nil embedded
// pointer yields one blank cell per embedded column, and that a column
// literally named "omitempty" is not treated as the option.
func TestExport_NilEmbeddedPointerKeepsColumns(t *testing.T) {
	rows := []embeddedRow{
		{EmbeddedBase: &EmbeddedBase{A: 1, B: 2}, C: "x"},
		{C: "y"},
	}

	exp := exporter.NewExporter(rows, excel.WithHeadings(true), excel.WithFormat(excel.FileFormatCSV))
	defer func() { _ = exp.Close() }()
	var buf bytes.Buffer
	if err := exp.ExportToWriter(context.Background(), &buf); err != nil {
		t.Fatalf("ExportToWriter: %v", err)
	}

	want := "\ufeffa,b,c,omitempty\r\n1,2,x,0\r\n,,y,0\r\n"
	if got := buf.String(); got != want {
		t.Errorf("CSV = %q, want %q", got, want)
	}
}
