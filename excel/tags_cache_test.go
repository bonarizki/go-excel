package excel

import (
	"reflect"
	"testing"
)

type cachedTagsRow struct {
	A string `excel:"a"`
}

// TestGetFieldTags_ReturnsCopy verifies mutating a returned slice does not
// corrupt the cache seen by later callers such as RowMapper.
func TestGetFieldTags_ReturnsCopy(t *testing.T) {
	typ := reflect.TypeOf(cachedTagsRow{})
	first, err := GetFieldTags(typ)
	if err != nil {
		t.Fatal(err)
	}
	first[0].Name = "mutated"

	second, err := GetFieldTags(typ)
	if err != nil {
		t.Fatal(err)
	}
	if second[0].Name != "a" {
		t.Errorf("cached Name = %q after caller mutation, want %q", second[0].Name, "a")
	}
}
