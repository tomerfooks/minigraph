package tomergraph

import (
	"context"
	"fmt"
	"sync"
)

// Parallel composes branches into a single Node: every branch runs
// concurrently on the same starting state, then merge folds the results —
// ordered like the branches — into the next state. It is fan-out/join as a
// combinator: the graph stays sequential and this counts as one step.
//
// Because a App's Invoke has a Node's signature, branches can be whole
// compiled subgraphs: Parallel(merge, subA.Invoke, subB.Invoke).
//
// Branches receive shallow copies of the state, so reference fields (slices,
// maps, pointers) are shared: treat them as read-only inside a branch and
// put new data in the branch's own copy for merge to reconcile. The first
// branch error cancels the siblings and fails the node; an Interrupt inside
// a branch is treated as a plain error.
func Parallel[S any](merge func(ctx context.Context, base S, results []S) (S, error), branches ...Node[S]) Node[S] {
	return func(ctx context.Context, state S) (S, error) {
		if merge == nil {
			return state, fmt.Errorf("tomergraph: Parallel with nil merge")
		}
		bctx, cancel := context.WithCancel(ctx)
		defer cancel()

		results := make([]S, len(branches))
		var (
			wg    sync.WaitGroup
			once  sync.Once
			first error
		)
		for i, branch := range branches {
			wg.Add(1)
			go func() {
				defer wg.Done()
				out, err := branch(bctx, state)
				if err != nil {
					once.Do(func() {
						first = fmt.Errorf("parallel branch %d: %w", i, err)
						cancel()
					})
					return
				}
				results[i] = out
			}()
		}
		wg.Wait()
		if first != nil {
			return state, first
		}
		return merge(ctx, state, results)
	}
}
