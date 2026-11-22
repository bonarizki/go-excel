package cellvalue

import (
	"database/sql"
	"testing"
	"time"
)

func TestNormalize(t *testing.T) {
	n := 7
	pn := &n
	var nilInt *int
	var nilNull *sql.NullInt64
	now := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		in   any
		want any
	}{
		{"nil", nil, nil},
		{"plain", 3, 3},
		{"pointer", &n, 7},
		{"pointer to pointer", &pn, 7},
		{"nil pointer", nilInt, nil},
		{"time pointer", &now, now},
		{"valid null string", sql.NullString{String: "a", Valid: true}, "a"},
		{"invalid null int", sql.NullInt64{}, nil},
		{"nil valuer pointer", nilNull, nil},
		{"any holding pointer", any(&n), 7},
	}
	for _, tt := range tests {
		if got := Normalize(tt.in); got != tt.want {
			t.Errorf("%s: Normalize(%v) = %#v, want %#v", tt.name, tt.in, got, tt.want)
		}
	}
}
