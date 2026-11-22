package excel

import "github.com/bonarizki/go-excel/internal/cellvalue"

// NormalizeCell converts v into the value a spreadsheet cell should hold:
// nil pointers and interfaces become nil, non-nil pointers are
// dereferenced, and driver.Valuer types such as sql.NullString are
// unwrapped. Exporters apply it to every cell, so writing *int or
// sql.NullInt64 produces the number rather than a memory address.
func NormalizeCell(v any) any {
	return cellvalue.Normalize(v)
}
