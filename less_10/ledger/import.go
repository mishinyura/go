type ImportResult struct {
	Accepted int64         `json:"accepted"`
	Rejected int64         `json:"rejected"`
	Errors   []ImportError `json:"errors"`
}

type ImportError struct {
	Index int    `json:"index"`
	Error string `json:"error"`
}

func (l *Ledger) BulkImport(ctx context.Context, txs []Transaction, workers int) (*ImportResult, error) {
	jobs := make(chan job)
	results := make(chan result)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
					if err := j.tx.Validate(); err != nil {
						results <- result{index: j.index, err: err}
						continue
					}
					if err := l.AddTransaction(ctx, j.tx); err != nil {
						results <- result{index: j.index, err: err}
					} else {
						results <- result{index: j.index, err: nil}
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		for i, tx := range txs {
			select {
			case <-ctx.Done():
				return
			default:
				jobs <- job{index: i, tx: tx}
			}
		}
		close(jobs)
	}()

	summary := &ImportResult{}
	for r := range results {
		if r.err != nil {
			atomic.AddInt64(&summary.Rejected, 1)
			summary.Errors = append(summary.Errors, ImportError{
				Index: r.index,
				Error: r.err.Error(),
			})
		} else {
			atomic.AddInt64(&summary.Accepted, 1)
		}
	}

	if ctx.Err() != nil {
		return summary, ctx.Err()
	}
	return summary, nil
}

type job struct {
	index int
	tx    Transaction
}

type result struct {
	index int
	err   error
}
