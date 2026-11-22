package importer

import (
	"context"
	"testing"

	"github.com/bonarizki/go-excel/excel"
)

type nameEmail struct {
	Name  string `excel:"name"`
	Email string `excel:"email"`
}

// TestWithColumns_HeadingsFollowColumnWindow verifies the heading row is
// cut to the same column window as data rows, so struct fields line up
// with the right cells.
func TestWithColumns_HeadingsFollowColumnWindow(t *testing.T) {
	path := writeCSVFile(t, "id,name,email\n1,Alice,a@x.io\n2,Bob,b@x.io\n")

	imp := NewImporter(excel.WithHeadings(true), excel.WithColumns(2, 3))
	defer func() { _ = imp.Close() }()
	if err := imp.ImportFromFile(context.Background(), path); err != nil {
		t.Fatalf("ImportFromFile: %v", err)
	}
	if got := imp.GetHeadings(); len(got) != 2 || got[0] != "name" || got[1] != "email" {
		t.Fatalf("GetHeadings() = %v, want [name email]", got)
	}

	people, err := ImportToStructs[nameEmail](context.Background(), path,
		excel.WithHeadings(true), excel.WithColumns(2, 3))
	if err != nil {
		t.Fatalf("ImportToStructs: %v", err)
	}
	if len(people) != 2 || people[0].Name != "Alice" || people[0].Email != "a@x.io" {
		t.Fatalf("people = %+v, want Alice/a@x.io first", people)
	}
}
