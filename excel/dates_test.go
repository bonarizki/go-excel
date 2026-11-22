package excel

import (
	"bytes"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

type dateRow struct {
	When time.Time `excel:"when"`
}

func TestRowMapper_DateParsing(t *testing.T) {
	day := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		in      string
		layouts []string
		want    time.Time
		wantErr bool
	}{
		{name: "iso", in: "2026-11-23", want: day},
		{name: "excelize numFmt 14", in: "11-23-26", want: day},
		{name: "excel serial", in: "46349", want: day},
		{name: "day-first rejected by default", in: "23-11-2026", wantErr: true},
		{name: "day-first with layout", in: "23-11-2026", layouts: []string{"02-01-2006"}, want: day},
	}
	for _, tt := range tests {
		m := NewRowMapper([]string{"when"})
		m.SetDateLayouts(tt.layouts)
		var got dateRow
		err := m.MapRow([]any{tt.in}, &got)
		if tt.wantErr {
			if err == nil {
				t.Errorf("%s: MapRow(%q) = %v, want error", tt.name, tt.in, got.When)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: MapRow(%q): %v", tt.name, tt.in, err)
			continue
		}
		if !got.When.Equal(tt.want) {
			t.Errorf("%s: MapRow(%q) = %v, want %v", tt.name, tt.in, got.When, tt.want)
		}
	}
}

func TestRowMapper_TimeLocation(t *testing.T) {
	jakarta := time.FixedZone("WIB", 7*3600)
	m := NewRowMapper([]string{"when"})
	m.SetTimeLocation(jakarta)

	var got dateRow
	if err := m.MapRow([]any{"2026-11-23 08:00"}, &got); err != nil {
		t.Fatalf("MapRow: %v", err)
	}
	if want := time.Date(2026, 11, 23, 8, 0, 0, 0, jakarta); !got.When.Equal(want) {
		t.Errorf("When = %v, want %v", got.When, want)
	}
}

// TestRowMapper_ExcelizeDateFormatsRoundTrip confirms the strings excelize
// renders for built-in date formats 14 and 22 parse with the defaults.
func TestRowMapper_ExcelizeDateFormatsRoundTrip(t *testing.T) {
	when := time.Date(2026, 11, 23, 8, 30, 0, 0, time.UTC)

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()
	for i, numFmt := range []int{14, 22} {
		style, err := f.NewStyle(&excelize.Style{NumFmt: numFmt})
		if err != nil {
			t.Fatal(err)
		}
		cell, _ := excelize.CoordinatesToCellName(1, i+1)
		if err := f.SetCellValue("Sheet1", cell, when); err != nil {
			t.Fatal(err)
		}
		if err := f.SetCellStyle("Sheet1", cell, cell, style); err != nil {
			t.Fatal(err)
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatal(err)
	}
	g, err := excelize.OpenReader(&buf)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = g.Close() }()
	rows, err := g.GetRows("Sheet1")
	if err != nil {
		t.Fatal(err)
	}

	wants := []time.Time{time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC), when}
	for i, row := range rows {
		m := NewRowMapper([]string{"when"})
		var got dateRow
		if err := m.MapRow([]any{row[0]}, &got); err != nil {
			t.Errorf("numFmt row %d (%q): %v", i+1, row[0], err)
			continue
		}
		if !got.When.Equal(wants[i]) {
			t.Errorf("numFmt row %d (%q) = %v, want %v", i+1, row[0], got.When, wants[i])
		}
	}
}
