package stream

import (
	"context"
	"strings"
	"testing"

	"github.com/bonarizki/go-excel/internal/concurrency"
)

// TestFinishDispatch_ReportsRowsWithoutResults verifies finishDispatch
// fails when fewer results were drained than tasks were submitted, which
// is what happens when the pool abandons buffered tasks during Stop.
// Pool stats alone cannot detect this: abandoned tasks are counted as
// TasksFailed, so Submitted == Completed + Failed still holds.
func TestFinishDispatch_ReportsRowsWithoutResults(t *testing.T) {
	pool := concurrency.NewWorkerPool(1, func(context.Context, concurrency.Task) (any, error) {
		return nil, nil
	})
	si := &StreamImporter{workerPool: pool}

	for i := 1; i <= 2; i++ {
		if err := pool.Submit(concurrency.Task{ID: uint64(i)}); err != nil {
			t.Fatalf("Submit: %v", err)
		}
	}

	workerErrChan := make(chan error)
	close(workerErrChan)
	drainerDone := make(chan struct{})
	close(drainerDone)

	err := si.finishDispatch(workerErrChan, drainerDone, nil)
	if err == nil || !strings.Contains(err.Error(), "import incomplete: 2 row(s)") {
		t.Fatalf("finishDispatch() = %v, want 'import incomplete: 2 row(s)'", err)
	}
}
