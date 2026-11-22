package importer

import (
	"context"
	"errors"
	"reflect"

	"github.com/bonarizki/go-excel/excel"
)

// ImportToStructs imports data from an Excel file into a slice of structs.
// T must be a struct type.
// The file is read row by row, and each row is mapped to a new instance of T.
// Accepts the same options as NewImporter; in particular, WithValidation(true)
// enables enforcement of min:/max:/pattern: struct tag constraints.
//
// Mapping errors are *excel.ValidationError values whose Row is the 1-based
// sheet row. With StopOnError(true) (the default) the first one is
// returned. With StopOnError(false) failing rows are skipped, ErrorCallback
// is consulted for each, and the rows that did map are returned together
// with an errors.Join of every failure.
func ImportToStructs[T any](ctx context.Context, filename string, opts ...Option) ([]T, error) {
	var zero T
	if reflect.TypeOf(zero).Kind() != reflect.Struct {
		return nil, excel.NewConfigError("type", reflect.TypeOf(zero).String(), "ImportToStructs requires a struct type")
	}

	imp := NewImporter(opts...)
	defer func() {
		_ = imp.Close()
	}()

	// ImportFromFile loads the whole file into memory; see the stream
	// package for a row-at-a-time alternative on very large files.
	err := imp.ImportFromFile(ctx, filename)
	if err != nil {
		return nil, err
	}
	rows := imp.GetRows()

	if len(rows) == 0 {
		return []T{}, nil
	}

	headers := imp.GetHeadings()
	if len(headers) == 0 {
		return nil, excel.NewConfigError("headers", nil, "no headers found in file")
	}

	cfg := imp.GetConfig()
	mapper := excel.NewRowMapper(headers)
	mapper.EnableValidation(cfg.ValidateData)
	mapper.SetConverters(cfg.Converters)
	mapper.SetDateLayouts(cfg.DateLayouts)
	mapper.SetTimeLocation(cfg.TimeLocation)

	results := make([]T, 0, len(rows))
	var rowErrs []error

	for i, row := range rows {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var item T
		if err := mapper.MapRow(row, &item); err != nil {
			rowNum := imp.sheetRowNum(i)
			err = withRow(err, rowNum)
			if cfg.StopOnError {
				return nil, err
			}
			rowErrs = append(rowErrs, err)
			if cfg.ErrorCallback != nil && !cfg.ErrorCallback(rowNum, err) {
				return nil, errors.Join(rowErrs...)
			}
			continue
		}

		results = append(results, item)
	}

	return results, errors.Join(rowErrs...)
}

// withRow stamps rowNum onto the ValidationError inside err, wrapping
// other errors in a ValidationError that carries the row.
func withRow(err error, rowNum int) error {
	var ve *excel.ValidationError
	if errors.As(err, &ve) {
		ve.Row = rowNum
		return err
	}
	return excel.NewValidationError(rowNum, "", nil, err.Error(), err)
}

// sheetRowNum returns the 1-based sheet row of GetRows()[idx].
func (i *BasicImporter) sheetRowNum(idx int) int {
	if idx < len(i.rowNums) {
		return i.rowNums[idx]
	}
	return 0
}
