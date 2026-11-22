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
// can run min/max validation against it.
func (m *RowMapper) convertInto(field reflect.Value, raw string, custom Converter) (any, error) {
	if custom != nil {
		return assignCustom(field, raw, custom)
	}

	s := strings.TrimSpace(raw)
	ft := field.Type()

	if ft == timeType {
		t, err := parseTime(s, m.dateLayouts, m.location)
		if err != nil {
			return nil, err
		}
		field.Set(reflect.ValueOf(t))
		return t, nil
	}

	if ft.Kind() == reflect.Pointer {
		if s == "" {
			field.Set(reflect.Zero(ft))
			return nil, nil
		}
		ptr := reflect.New(ft.Elem())
		converted, err := m.convertInto(ptr.Elem(), s, nil)
		if err != nil {
			return nil, err
		}
		field.Set(ptr)
		return converted, nil
	}

	if field.CanAddr() {
		addr := field.Addr()
		if addr.Type().Implements(sqlScannerType) {
			var src any
			if s != "" {
				src = s
			}
			if err := addr.Interface().(sql.Scanner).Scan(src); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrTypeMismatch, err)
			}
			return field.Interface(), nil
		}
		if addr.Type().Implements(textUnmarshalerType) {
			if err := addr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(s)); err != nil {
				return nil, fmt.Errorf("%w: %v", ErrTypeMismatch, err)
			}
			return field.Interface(), nil
		}
	}

	switch ft.Kind() {
	case reflect.String:
		field.SetString(s)
		return s, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := convertToInt(s)
		if err != nil {
			return nil, err
		}
		if field.OverflowInt(n) {
			return nil, fmt.Errorf("%w: %d overflows %s", ErrTypeMismatch, n, ft)
		}
		field.SetInt(n)
		return n, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := convertToUint(s)
		if err != nil {
			return nil, err
		}
		if field.OverflowUint(n) {
			return nil, fmt.Errorf("%w: %d overflows %s", ErrTypeMismatch, n, ft)
		}
		field.SetUint(n)
		return n, nil
	case reflect.Float32, reflect.Float64:
		f, err := convertToFloat(s)
		if err != nil {
			return nil, err
		}
		if field.OverflowFloat(f) {
			return nil, fmt.Errorf("%w: %v overflows %s", ErrTypeMismatch, f, ft)
		}
		field.SetFloat(f)
		return f, nil
	case reflect.Bool:
		b, err := convertToBool(s)
		if err != nil {
			return nil, err
		}
		field.SetBool(b)
		return b, nil
	}

	return nil, fmt.Errorf("%w: unsupported type: %s", ErrTypeMismatch, ft)
}

// assignCustom runs a per-column Converter and stores its result, failing
// instead of silently skipping when the result does not fit the field.
func assignCustom(field reflect.Value, raw string, custom Converter) (any, error) {
	v, err := custom(raw, field.Type())
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}

	rv := reflect.ValueOf(v)
	ft := field.Type()
	switch {
	case rv.Type().AssignableTo(ft):
		field.Set(rv)
	case sameKindFamily(rv.Kind(), ft.Kind()) && rv.Type().ConvertibleTo(ft):
		field.Set(rv.Convert(ft))
	default:
		return nil, fmt.Errorf("%w: converter returned %s, field is %s", ErrTypeMismatch, rv.Type(), ft)
	}
	return v, nil
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
