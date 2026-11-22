// Package cellvalue turns arbitrary Go values into what a spreadsheet cell
// should hold. It lives in internal so excel, excel/csvio and stream can
// share one implementation without import cycles.
package cellvalue

import (
	"database/sql/driver"
	"reflect"
	"time"
)

var valuerType = reflect.TypeOf((*driver.Valuer)(nil)).Elem()

// maxDerefDepth bounds pointer chasing so a self-referential pointer
// cannot loop forever.
const maxDerefDepth = 16

// Normalize dereferences pointers and interfaces (nil becomes nil) and
// unwraps driver.Valuer values such as sql.NullString, so a cell gets the
// underlying value instead of a memory address or struct dump.
func Normalize(v any) any {
	if v == nil {
		return nil
	}
	switch v.(type) {
	case string, bool, time.Time, []byte,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return v
	}

	rv := reflect.ValueOf(v)
	for depth := 0; depth < maxDerefDepth; depth++ {
		if rv.Type().Implements(valuerType) {
			if rv.Kind() == reflect.Pointer && rv.IsNil() {
				return nil
			}
			val, err := rv.Interface().(driver.Valuer).Value()
			if err != nil {
				return rv.Interface()
			}
			return val
		}
		if rv.Kind() != reflect.Pointer && rv.Kind() != reflect.Interface {
			return rv.Interface()
		}
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}
	return rv.Interface()
}
