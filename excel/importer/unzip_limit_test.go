package importer

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bonarizki/go-excel/excel"
	"github.com/xuri/excelize/v2"
)

func xlsxWithBigCell(t *testing.T) []byte {
	t.Helper()
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	if err := f.SetCellValue("Sheet1", "A1", strings.Repeat("x", 64<<10)); err != nil {
		t.Fatal(err)
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestImportFromReader_UnzipLimit verifies a workbook whose uncompressed
// size exceeds MaxUnzipSize is rejected, while the default limit accepts it.
func TestImportFromReader_UnzipLimit(t *testing.T) {
	data := xlsxWithBigCell(t)

	limited := NewImporter(excel.WithHeadings(false), excel.WithUnzipLimits(16<<10, 0))
	defer func() { _ = limited.Close() }()
	if err := limited.ImportFromReader(context.Background(), bytes.NewReader(data)); err == nil {
		t.Error("ImportFromReader over the unzip limit = nil error, want error")
	}

	def := NewImporter(excel.WithHeadings(false))
	defer func() { _ = def.Close() }()
	if err := def.ImportFromReader(context.Background(), bytes.NewReader(data)); err != nil {
		t.Errorf("ImportFromReader with default limits: %v", err)
	}
}
