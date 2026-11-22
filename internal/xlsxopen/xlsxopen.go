// Package xlsxopen opens XLSX workbooks with the size limits configured by
// callers, so every importer entry point applies the same zip-bomb
// protection.
package xlsxopen

import (
	"io"

	"github.com/xuri/excelize/v2"
)

// Options builds excelize open options from the configured limits. Zero
// values defer to excelize's defaults. An XML limit above the total is
// clamped to the total, since excelize rejects that pair outright.
func Options(maxUnzip, maxXML int64) excelize.Options {
	if maxUnzip > 0 && maxXML > maxUnzip {
		maxXML = maxUnzip
	}
	return excelize.Options{UnzipSizeLimit: maxUnzip, UnzipXMLSizeLimit: maxXML}
}

// File opens the workbook at path with the given limits.
func File(path string, maxUnzip, maxXML int64) (*excelize.File, error) {
	return excelize.OpenFile(path, Options(maxUnzip, maxXML))
}

// Reader opens the workbook read from r with the given limits.
func Reader(r io.Reader, maxUnzip, maxXML int64) (*excelize.File, error) {
	return excelize.OpenReader(r, Options(maxUnzip, maxXML))
}
