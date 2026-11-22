package excel

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type status int

type upperText string

func (u *upperText) UnmarshalText(b []byte) error {
	*u = upperText(strings.ToUpper(string(b)))
	return nil
}

type richRow struct {
	Status status        `excel:"status"`
	Count  *int          `excel:"count"`
	Name   *string       `excel:"name"`
	When   *time.Time    `excel:"when"`
	Code   upperText     `excel:"code"`
	Score  sql.NullInt64 `excel:"score"`
	Tiny   int8          `excel:"tiny"`
}

func TestRowMapper_ConvertsByKindNotTypeName(t *testing.T) {
	headers := []string{"status", "count", "name", "when", "code", "score", "tiny"}
	m := NewRowMapper(headers)

	var got richRow
	row := []any{"2", "7", "bob", "2026-11-23", "ab", "42", "5"}
	if err := m.MapRow(row, &got); err != nil {
		t.Fatalf("MapRow: %v", err)
	}
	if got.Status != 2 {
		t.Errorf("Status = %d, want 2", got.Status)
	}
	if got.Count == nil || *got.Count != 7 {
		t.Errorf("Count = %v, want 7", got.Count)
	}
	if got.Name == nil || *got.Name != "bob" {
		t.Errorf("Name = %v, want bob", got.Name)
	}
	if got.When == nil || got.When.Format("2006-01-02") != "2026-11-23" {
		t.Errorf("When = %v, want 2026-11-23", got.When)
	}
	if got.Code != "AB" {
		t.Errorf("Code = %q, want AB", got.Code)
	}
	if !got.Score.Valid || got.Score.Int64 != 42 {
		t.Errorf("Score = %+v, want valid 42", got.Score)
	}
	if got.Tiny != 5 {
		t.Errorf("Tiny = %d, want 5", got.Tiny)
	}
}

func TestRowMapper_EmptyCellsLeavePointersNil(t *testing.T) {
	m := NewRowMapper([]string{"count", "name", "when", "score"})

	var got richRow
	if err := m.MapRow([]any{"", "", "", ""}, &got); err != nil {
		t.Fatalf("MapRow: %v", err)
	}
	if got.Count != nil || got.Name != nil || got.When != nil {
		t.Errorf("pointers = %v %v %v, want all nil", got.Count, got.Name, got.When)
	}
	if got.Score.Valid {
		t.Errorf("Score.Valid = true, want false for empty cell")
	}
}

func TestRowMapper_IntOverflowIsAnError(t *testing.T) {
	m := NewRowMapper([]string{"tiny"})

	var got richRow
	err := m.MapRow([]any{"300"}, &got)
	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("MapRow(300 into int8) = %v, want ErrTypeMismatch", err)
	}
}

func TestRowMapper_ConverterWrongTypeIsAnError(t *testing.T) {
	m := NewRowMapper([]string{"tiny"})
	m.SetConverters(map[string]Converter{
		"tiny": func(string, reflect.Type) (any, error) { return "nope", nil },
	})

	var got richRow
	err := m.MapRow([]any{"1"}, &got)
	if !errors.Is(err, ErrTypeMismatch) {
		t.Fatalf("MapRow with mismatched converter = %v, want ErrTypeMismatch", err)
	}
}
