package tomergraph

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestParallelMergesInBranchOrder(t *testing.T) {
	addN := func(n int) Node[state] {
		return func(_ context.Context, s state) (state, error) {
			s.N += n // own copy: scalar writes are branch-local
			return s, nil
		}
	}
	node := Parallel(func(_ context.Context, base state, results []state) (state, error) {
		for _, r := range results {
			base.N += r.N
			base.Path = append(base.Path, "merged")
		}
		return base, nil
	}, addN(1), addN(2), addN(3))

	app, err := New[state]().
		AddNode("fan", node).
		AddEdge(Start, "fan").
		AddEdge("fan", End).
		Compile()
	if err != nil {
		t.Fatal(err)
	}

	final, err := app.Invoke(t.Context(), state{N: 10})
	if err != nil {
		t.Fatal(err)
	}
	// base 10 + (10+1) + (10+2) + (10+3) = 46, and one whole graph step.
	if final.N != 46 || len(final.Path) != 3 {
		t.Errorf("final = %+v, want N=46 with 3 merges", final)
	}
}

func TestParallelBranchErrorCancelsSiblings(t *testing.T) {
	block := func(ctx context.Context, s state) (state, error) {
		<-ctx.Done() // released only by sibling cancellation
		return s, ctx.Err()
	}
	boom := func(_ context.Context, s state) (state, error) {
		return s, errors.New("boom")
	}
	node := Parallel(func(_ context.Context, base state, _ []state) (state, error) {
		return base, nil
	}, block, boom)

	_, err := node(t.Context(), state{})
	if err == nil || !strings.Contains(err.Error(), "branch 1: boom") {
		t.Fatalf("err = %v, want the causing branch's error", err)
	}
}

func TestParallelSubgraphBranches(t *testing.T) {
	sub := func(name string, n int) *App[state] {
		app, err := New[state]().
			AddNode(name, visit(name, func(s state) state { s.N += n; return s })).
			AddEdge(Start, name).
			AddEdge(name, End).
			Compile()
		if err != nil {
			t.Fatal(err)
		}
		return app
	}

	// Whole compiled graphs as parallel branches: Invoke is a Node.
	node := Parallel(func(_ context.Context, base state, results []state) (state, error) {
		for _, r := range results {
			base.N += r.N
		}
		return base, nil
	}, sub("x", 1).Invoke, sub("y", 2).Invoke)

	out, err := node(t.Context(), state{})
	if err != nil {
		t.Fatal(err)
	}
	if out.N != 3 {
		t.Errorf("N = %d, want 3", out.N)
	}
}

func TestParallelNilMerge(t *testing.T) {
	node := Parallel[state](nil)
	if _, err := node(t.Context(), state{}); err == nil {
		t.Fatal("want error for nil merge")
	}
}
