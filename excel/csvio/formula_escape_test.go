package csvio

import "testing"

func TestFormatValue_FormulaEscape(t *testing.T) {
	tests := []struct {
		name   string
		value  any
		escape bool
		want   string
	}{
		{name: "formula string escaped", value: "=1+1", escape: true, want: "'=1+1"},
		{name: "at sign escaped", value: "@sum", escape: true, want: "'@sum"},
		{name: "string negative escaped", value: "-5", escape: true, want: "'-5"},
		{name: "int negative untouched", value: -5, escape: true, want: "-5"},
		{name: "escape disabled", value: "=1+1", escape: false, want: "=1+1"},
		{name: "tab prefix escaped", value: "\t=cmd", escape: true, want: "'\t=cmd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatValue(tt.value, "", tt.escape)
			if got != tt.want {
				t.Fatalf("FormatValue(%#v, \"\", %v) = %q, want %q", tt.value, tt.escape, got, tt.want)
			}
		})
	}
}
