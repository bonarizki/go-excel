package importer

import "github.com/xuri/excelize/v2"

// setFile replaces the importer's excelize file, closing the previous one
// if this importer opened it. owned reports whether f must be closed by
// this importer later.
func (i *BasicImporter) setFile(f *excelize.File, owned bool) {
	if i.shouldCloseFile && i.file != nil && i.file != f {
		_ = i.file.Close()
	}
	i.file = f
	i.shouldCloseFile = owned
}

// resetRunState clears everything a previous import left behind so a
// reused importer reports only the current run's rows and errors.
func (i *BasicImporter) resetRunState() {
	i.rows = make([][]any, 0)
	i.headings = make([]string, 0)
	i.errors = make([]error, 0)
	i.errorCount = 0
	i.rowNums = i.rowNums[:0]
	i.firstRowNum = 1
	i.currentRow = 0
	i.totalRows = 0
}
