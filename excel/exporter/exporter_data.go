package exporter

import (
	"context"
	"fmt"
	"reflect"

	"github.com/bonarizki/go-excel/excel"
)

// Input-to-row conversion for BasicExporter: type switches, Exportable
// unwrapping, and the reflection fallback used for structs.

// maxExportableDepth bounds how many times prepareData will unwrap a
// nested excel.Exportable before giving up. Without a limit, a Data()
// implementation that returns itself (directly or through a cycle of
// Exportable values) would recurse until the goroutine stack overflows.
const maxExportableDepth = 10

// isEmptyExportData reports whether data should be treated as empty
// input to Export: a nil interface, or a non-nil slice/array/map with
// zero elements. Anything else (structs, scalars, pointers to either)
// is left for prepareData to interpret.
func isEmptyExportData(data any) bool {
	if data == nil {
		return true
	}

	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return val.Len() == 0
	default:
		return false
	}
}

// prepareData converts the input data to a uniform row format, where
// each row is []interface{}.
func (e *BasicExporter) prepareData(ctx context.Context) ([][]any, error) {
	return e.prepareDataAtDepth(ctx, 0)
}

// prepareDataAtDepth is the recursive implementation behind prepareData;
// depth counts how many Exportable.Data() calls have been unwrapped so
// far in this chain.
func (e *BasicExporter) prepareDataAtDepth(ctx context.Context, depth int) ([][]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if depth > maxExportableDepth {
		return nil, fmt.Errorf("excel.Exportable nesting exceeds max depth of %d; check for a Data() implementation that returns itself or forms a cycle", maxExportableDepth)
	}

	switch v := e.data.(type) {
	case [][]any:
		return v, nil

	case []any:
		return e.convertSliceToRows(v)

	case excel.Exportable:
		e.data = v.Data()
		return e.prepareDataAtDepth(ctx, depth+1)

	default:
		return e.convertUsingReflection(v)
	}
}

// convertSliceToRows converts []interface{} to [][]interface{}, where
// each element becomes a single-cell row.
func (e *BasicExporter) convertSliceToRows(slice []any) ([][]any, error) {
	var rows [][]any
	if e.config.PreallocateRows && len(slice) > 0 {
		rows = make([][]any, 0, len(slice))
	}

	for _, item := range slice {
		rows = append(rows, []any{item})
	}

	return rows, nil
}

// reflectSliceRows turns each element of a slice or array into a row. For
// struct elements the heading row comes from the excel tags, not from the
// first record, which would otherwise be consumed as the header by
// exportRows.
func (e *BasicExporter) reflectSliceRows(val reflect.Value) ([][]any, error) {
	var rows [][]any
	if e.config.PreallocateRows && val.Len() > 0 {
		rows = make([][]any, 0, val.Len()+1)
	}

	if e.config.WithHeadings {
		header, ok, err := structHeaderRow(val.Type().Elem())
		if err != nil {
			return nil, err
		}
		if ok {
			rows = append(rows, header)
		}
	}

	for i := 0; i < val.Len(); i++ {
		row, err := e.reflectStructToRow(val.Index(i))
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// convertUsingReflection is the fallback for data types not explicitly
// handled by prepareDataAtDepth: pointers are dereferenced, slices and
// arrays become one row per element, structs become a single row, and
// anything else becomes a single-cell row.
func (e *BasicExporter) convertUsingReflection(data any) ([][]any, error) {
	val := reflect.ValueOf(data)

	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	if val.Kind() == reflect.Slice || val.Kind() == reflect.Array {
		return e.reflectSliceRows(val)
	}

	if val.Kind() == reflect.Struct {
		row, err := e.reflectStructToRow(val)
		if err != nil {
			return nil, err
		}
		if e.config.WithHeadings {
			header, _, err := structHeaderRow(val.Type())
			if err != nil {
				return nil, err
			}
			return [][]any{header, row}, nil
		}
		return [][]any{row}, nil
	}

	return [][]any{{data}}, nil
}

// structHeaderRow returns one heading cell per exported column of t (or
// *t), named by its excel tag, falling back to the Go field name. ok is
// false when t is not a struct type.
func structHeaderRow(t reflect.Type) (header []any, ok bool, err error) {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil, false, nil
	}
	tags, err := excel.GetFieldTags(t)
	if err != nil {
		return nil, false, err
	}
	header = make([]any, len(tags))
	for i := range tags {
		header[i] = tags[i].Name
	}
	return header, true, nil
}

// reflectStructToRow converts a struct (or pointer to struct) to a row,
// one cell per column reported by excel.GetFieldTags, the same walker
// the importer uses, so exported files round-trip through
// ImportToStructs. Fields under a nil embedded pointer become blank
// cells, keeping every later column in place.
func (e *BasicExporter) reflectStructToRow(item reflect.Value) ([]any, error) {
	if item.Kind() == reflect.Pointer {
		if item.IsNil() {
			return nil, fmt.Errorf("expected struct or pointer to struct, got nil pointer")
		}
		item = item.Elem()
	}

	if item.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct or pointer to struct, got %s", item.Kind())
	}

	tags, err := excel.GetFieldTags(item.Type())
	if err != nil {
		return nil, err
	}

	row := make([]any, len(tags))
	for i := range tags {
		fieldValue, err := item.FieldByIndexErr(tags[i].Index)
		if err != nil {
			continue
		}
		if tags[i].OmitEmpty && fieldValue.IsZero() {
			continue
		}
		row[i] = excel.NormalizeCell(fieldValue.Interface())
	}
	return row, nil
}
