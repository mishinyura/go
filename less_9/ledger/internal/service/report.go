func (s *Service) CalculateSummary(ctx context.Context, from, to time.Time) (map[string]float64, error) {
	categories, err := s.expenseRepo.GetCategories(ctx)
	if err != nil {
		return nil, err
	}

	results := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errCh := make(chan error, len(categories))

	// heartbeat
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("report: still calculating...")
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	for _, cat := range categories {
		wg.Add(1)
		go func(cat string) {
			defer wg.Done()
			sum, err := s.expenseRepo.SumByCategory(ctx, cat, from, to)
			if err != nil {
				errCh <- err
				return
			}

			mu.Lock()
			results[cat] = sum
			mu.Unlock()
		}(cat)
	}

	wg.Wait()
	ticker.Stop()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return results, nil
	}
}
