package excel

import (
	"database/sql"
	"encoding"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	timeType            = reflect.TypeOf(time.Time{})
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	sqlScannerType      = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

// convertInto parses raw into field according to field's type rather than
// its type name, so named types (type Status int), pointers,
// encoding.TextUnmarshaler and sql.Scanner implementations all work. It
// returns the value that was stored (dereferenced for pointers) so callers
// can run min/max validation against it; the result is the zero
// reflect.Value when nothing was stored (an empty cell for a pointer field,
// or a Converter returning nil).
func (m *RowMapper) convertInto(field reflect.Value, raw string, custom Converter) (reflect.Value, error) {
	if custom != nil {
		return assignCustom(field, raw, custom)
	}

	s := strings.TrimSpace(raw)
	ft := field.Type()

	switch {
	case ft == timeType:
		t, err := parseTime(s, m.dateLayouts, m.location)
		if err != nil {
			return reflect.Value{}, err
		}
		field.Set(reflect.ValueOf(t))
		return field, nil
	case ft.Kind() == reflect.Pointer:
		return m.convertPointer(field, s)
	}

	if stored, handled, err := convertViaInterface(field, s); handled {
		return stored, err
	}
	return convertKind(field, s)
}

// convertPointer leaves field nil for an empty cell and otherwise converts
// into a newly allocated element.
func (m *RowMapper) convertPointer(field reflect.Value, s string) (reflect.Value, error) {
	if s == "" {
		field.Set(reflect.Zero(field.Type()))
		return reflect.Value{}, nil
	}
	ptr := reflect.New(field.Type().Elem())
	stored, err := m.convertInto(ptr.Elem(), s, nil)
	if err != nil {
		return reflect.Value{}, err
	}
	field.Set(ptr)
	return stored, nil
}

// convertViaInterface uses sql.Scanner or encoding.TextUnmarshaler when
// *field implements one. handled is false when it implements neither.
func convertViaInterface(field reflect.Value, s string) (stored reflect.Value, handled bool, err error) {
	if !field.CanAddr() {
		return reflect.Value{}, false, nil
	}
	addr := field.Addr()
	switch {
	case addr.Type().Implements(sqlScannerType):
		var src any
		if s != "" {
			src = s
		}
		err = addr.Interface().(sql.Scanner).Scan(src)
	case addr.Type().Implements(textUnmarshalerType):
		err = addr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s))
	default:
		return reflect.Value{}, false, nil
	}
	if err != nil {
		return reflect.Value{}, true, fmt.Errorf("%w: %w", ErrTypeMismatch, err)
	}
	return field, true, nil
}

// convertKind handles the primitive kinds, rejecting values that overflow
// the field's size.
func convertKind(field reflect.Value, s string) (reflect.Value, error) {
	ft := field.Type()
	switch ft.Kind() {
	case reflect.String:
		field.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := convertToInt(s)
		if err != nil {
			return reflect.Value{}, err
		}
		if field.OverflowInt(n) {
			return reflect.Value{}, fmt.Errorf("%w: %d overflows %s", ErrTypeMismatch, n, ft)
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := convertToUint(s)
		if err != nil {
			return reflect.Value{}, err
		}
		if field.OverflowUint(n) {
			return reflect.Value{}, fmt.Errorf("%w: %d overflows %s", ErrTypeMismatch, n, ft)
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := convertToFloat(s)
		if err != nil {
			return reflect.Value{}, err
		}
		if field.OverflowFloat(f) {
			return reflect.Value{}, fmt.Errorf("%w: %v overflows %s", ErrTypeMismatch, f, ft)
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := convertToBool(s)
		if err != nil {
			return reflect.Value{}, err
		}
		field.SetBool(b)
	default:
		return reflect.Value{}, fmt.Errorf("%w: unsupported type: %s", ErrTypeMismatch, ft)
	}
	return field, nil
}

// assignCustom runs a per-column Converter and stores its result, failing
// instead of silently skipping when the result does not fit the field.
func assignCustom(field reflect.Value, raw string, custom Converter) (reflect.Value, error) {
	v, err := custom(raw, field.Type())
	if err != nil {
		return reflect.Value{}, err
	}
	if v == nil {
		return reflect.Value{}, nil
	}

	rv := reflect.ValueOf(v)
	ft := field.Type()
	switch {
	case rv.Type().AssignableTo(ft):
		field.Set(rv)
	case sameKindFamily(rv.Kind(), ft.Kind()) && rv.Type().ConvertibleTo(ft):
		field.Set(rv.Convert(ft))
	default:
		return reflect.Value{}, fmt.Errorf("%w: converter returned %s, field is %s", ErrTypeMismatch, rv.Type(), ft)
	}
	return rv, nil
}

// sameKindFamily guards reflect.Convert against conversions that are legal
// in Go but never what a converter meant, such as int to string (which
// yields a rune).
func sameKindFamily(a, b reflect.Kind) bool {
	family := func(k reflect.Kind) int {
		switch {
		case isSignedInt(k), isUnsignedInt(k), isFloat(k):
			return 1
		case k == reflect.String:
			return 2
		case k == reflect.Bool:
			return 3
		default:
			return 0
		}
	}
	fa := family(a)
	return fa != 0 && fa == family(b)
}
