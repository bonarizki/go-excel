package importer

import (
	"context"
	"errors"
	"testing"

	"github.com/bonarizki/go-excel/excel"
)

type ageRow struct {
	Name string `excel:"name"`
	Age  int    `excel:"age"`
}

// Sheet rows: 1 heading, 2 Ann, 3 empty (skipped), 4 Bob, 5 bad age, 6 Cy.
const ageCSV = "name,age\nAnn,30\n,\nBob,40\nEve,old\nCy,20\n"

func TestImportToStructs_ErrorReportsSheetRow(t *testing.T) {
	path := writeCSVFile(t, ageCSV)

	_, err := ImportToStructs[ageRow](context.Background(), path)
	var ve *excel.ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want *excel.ValidationError", err)
	}
	if ve.Row != 5 || ve.Field != "age" {
		t.Errorf("Row, Field = %d, %q, want 5, \"age\"", ve.Row, ve.Field)
	}
}

func TestImportToStructs_StopOnErrorFalseKeepsGoodRows(t *testing.T) {
	path := writeCSVFile(t, ageCSV)

	var callbackRows []int
	got, err := ImportToStructs[ageRow](context.Background(), path,
		excel.WithStopOnError(false),
		excel.WithErrorHandler(func(row int, _ error) bool {
			callbackRows = append(callbackRows, row)
			return true
		}),
	)
	if err == nil {
		t.Fatal("err = nil, want the joined row error")
	}
	if len(got) != 3 || got[2].Name != "Cy" {
		t.Errorf("got = %+v, want Ann, Bob, Cy", got)
	}
	if len(callbackRows) != 1 || callbackRows[0] != 5 {
		t.Errorf("ErrorCallback rows = %v, want [5]", callbackRows)
	}
}
