package importer

import "github.com/bonarizki/go-excel/excel/config"

// beginImport clears per-import state and closes any file opened by a
// previous ImportFromFile call on the same importer instance.
func (i *BasicImporter) beginImport() {
	if i.shouldCloseFile && i.file != nil {
		_ = i.file.Close()
	}
	i.file = nil
	i.shouldCloseFile = false
	i.rows = nil
	i.headings = nil
	i.errors = nil
	i.currentRow = 0
	i.totalRows = 0
	i.format = config.FileFormatXLSX
}