package stream

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/bonarizki/go-excel/excel/config"
)

// TestImportStream_RowNumIsSheetRow verifies rowNum is the row's position
// in the file (heading counted) even when rows before it were skipped by
// the start row or as empty rows.
func TestImportStream_RowNumIsSheetRow(t *testing.T) {
	// row 1 heading, row 2 skipped by start row, row 3 empty, rows 4-5 data
	// (encoding/csv drops blank lines, so the empty row is written as ",").
	content := "id,v\nskip,x\n,\nA,1\nB,2\n"
	si, err := NewStreamImporter(strings.NewReader(content),
		WithImportFormat(config.FileFormatCSV),
		WithImportHeadings(),
		WithImportRows(3, 0),
		WithImportSkipEmptyRows(true),
	)
	if err != nil {
		t.Fatalf("NewStreamImporter: %v", err)
	}
	defer si.Close()

	var mu sync.Mutex
	got := map[int]string{}
	err = si.ImportStream(context.Background(), func(_ context.Context, rowNum int, row []any) ([]any, error) {
		mu.Lock()
		defer mu.Unlock()
		got[rowNum] = row[0].(string)
		return row, nil
	})
	if err != nil {
		t.Fatalf("ImportStream: %v", err)
	}
	if len(got) != 2 || got[4] != "A" || got[5] != "B" {
		t.Errorf("rows by number = %v, want map[4:A 5:B]", got)
	}
}
