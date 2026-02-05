package worker

import (
	"context"
	"sync"
)

type WorkFunc[I any, O any] func(context.Context, I) (O, error)

// Pool runs a fixed number of workers and preserves input order in results.
type Pool[I any, O any] struct {
	Workers int
	Work    WorkFunc[I, O]
}

func (p Pool[I, O]) Run(ctx context.Context, inputs []I) ([]O, error) {
	if p.Workers < 1 {
		p.Workers = 1
	}

	type job struct {
		idx int
		in  I
	}
	type result struct {
		idx int
		out O
		err error
	}

	jobs := make(chan job)
	results := make(chan result)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for i := 0; i < p.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				out, err := p.Work(ctx, j.in)
				select {
				case results <- result{idx: j.idx, out: out, err: err}:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		for i, in := range inputs {
			select {
			case jobs <- job{idx: i, in: in}:
			case <-ctx.Done():
				close(jobs)
				wg.Wait()
				close(results)
				return
			}
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	outs := make([]O, len(inputs))
	var firstErr error
	for res := range results {
		if res.err != nil && firstErr == nil {
			firstErr = res.err
			cancel()
		}
		outs[res.idx] = res.out
	}
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return outs, nil
}