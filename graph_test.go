package minigraph

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

type state struct {
	N    int
	Path []string
}

func visit(name string, f func(s state) state) Node[state] {
	return func(_ context.Context, s state) (state, error) {
		s = f(s)
		s.Path = append(append([]string(nil), s.Path...), name)
		return s, nil
	}
}

func TestLinear(t *testing.T) {
	g := New[state]().
		AddNode("a", visit("a", func(s state) state { s.N++; return s })).
		AddNode("b", visit("b", func(s state) state { s.N *= 10; return s })).
		AddEdge(Start, "a").
		AddEdge("a", "b").
		AddEdge("b", End)
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	final, err := app.Invoke(t.Context(), state{})
	if err != nil {
		t.Fatal(err)
	}
	if final.N != 10 {
		t.Errorf("N = %d, want 10", final.N)
	}
	if got := strings.Join(final.Path, ","); got != "a,b" {
		t.Errorf("Path = %s, want a,b", got)
	}
}

func TestStreamYieldsEveryStep(t *testing.T) {
	g := New[state]().
		AddNode("a", visit("a", func(s state) state { return s })).
		AddNode("b", visit("b", func(s state) state { return s })).
		AddEdge(Start, "a").
		AddEdge("a", "b").
		AddEdge("b", End)
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	var order []string
	for step, err := range app.Stream(t.Context(), state{}) {
		if err != nil {
			t.Fatal(err)
		}
		order = append(order, step.Node)
	}
	if got := strings.Join(order, ","); got != "a,b" {
		t.Errorf("stream order = %s, want a,b", got)
	}
}

func TestConditionalLoop(t *testing.T) {
	g := New[state]().
		AddNode("inc", visit("inc", func(s state) state { s.N++; return s })).
		AddEdge(Start, "inc").
		AddRouter("inc", func(_ context.Context, s state) (string, error) {
			if s.N < 3 {
				return "inc", nil
			}
			return End, nil
		})
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	final, err := app.Invoke(t.Context(), state{})
	if err != nil {
		t.Fatal(err)
	}
	if final.N != 3 {
		t.Errorf("N = %d, want 3", final.N)
	}
}

func TestMaxSteps(t *testing.T) {
	g := New[state]().
		AddNode("loop", visit("loop", func(s state) state { s.N++; return s })).
		AddEdge(Start, "loop").
		AddEdge("loop", "loop")
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}
	app.MaxSteps = 5

	final, err := app.Invoke(t.Context(), state{})
	if !errors.Is(err, ErrMaxSteps) {
		t.Fatalf("err = %v, want ErrMaxSteps", err)
	}
	if final.N != 5 {
		t.Errorf("last good state N = %d, want 5", final.N)
	}
}

func TestNodeError(t *testing.T) {
	boom := errors.New("boom")
	g := New[state]().
		AddNode("a", visit("a", func(s state) state { s.N = 7; return s })).
		AddNode("b", func(_ context.Context, s state) (state, error) { return state{}, boom }).
		AddEdge(Start, "a").
		AddEdge("a", "b").
		AddEdge("b", End)
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	final, err := app.Invoke(t.Context(), state{})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrapped boom", err)
	}
	if !strings.Contains(err.Error(), `node "b"`) {
		t.Errorf("err = %v, want node name in message", err)
	}
	if final.N != 7 {
		t.Errorf("last good state N = %d, want 7", final.N)
	}
}

func TestRouterError(t *testing.T) {
	boom := errors.New("router boom")
	g := New[state]().
		AddNode("a", visit("a", func(s state) state { return s })).
		AddEdge(Start, "a").
		AddRouter("a", func(_ context.Context, _ state) (string, error) {
			return "", boom
		})
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := app.Invoke(t.Context(), state{}); !errors.Is(err, boom) {
		t.Fatalf("err = %v, want wrapped router boom", err)
	}
}

func TestRouterUnknownTarget(t *testing.T) {
	g := New[state]().
		AddNode("a", visit("a", func(s state) state { return s })).
		AddEdge(Start, "a").
		AddRouter("a", func(_ context.Context, _ state) (string, error) {
			return "ghost", nil
		})
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}
	_, err = app.Invoke(t.Context(), state{})
	if err == nil || !strings.Contains(err.Error(), `unknown node "ghost"`) {
		t.Fatalf("err = %v, want unknown node error", err)
	}
}

func TestCompileErrors(t *testing.T) {
	noop := func(_ context.Context, s state) (state, error) { return s, nil }
	for _, tc := range []struct {
		name  string
		build func() *Graph[state]
		want  string
	}{
		{
			"no entry edge",
			func() *Graph[state] {
				return New[state]().AddNode("a", noop).AddEdge("a", End)
			},
			"no entry edge",
		},
		{
			"unknown static target",
			func() *Graph[state] {
				return New[state]().AddNode("a", noop).AddEdge(Start, "a").AddEdge("a", "ghost")
			},
			"unknown target",
		},
		{
			"node without outgoing edge",
			func() *Graph[state] {
				return New[state]().AddNode("a", noop).AddEdge(Start, "a")
			},
			"no outgoing edge",
		},
		{
			"edge from unknown node",
			func() *Graph[state] {
				return New[state]().AddNode("a", noop).AddEdge(Start, "a").
					AddEdge("a", End).AddEdge("ghost", End)
			},
			"edge from unknown node",
		},
		{
			"duplicate node",
			func() *Graph[state] {
				return New[state]().AddNode("a", noop).AddNode("a", noop).
					AddEdge(Start, "a").AddEdge("a", End)
			},
			"duplicate node",
		},
		{
			"duplicate edge",
			func() *Graph[state] {
				return New[state]().AddNode("a", noop).AddEdge(Start, "a").
					AddEdge("a", End).AddEdge("a", End)
			},
			"already set",
		},
		{
			"reserved name",
			func() *Graph[state] {
				return New[state]().AddNode(End, noop).AddNode("a", noop).
					AddEdge(Start, "a").AddEdge("a", End)
			},
			"reserved",
		},
		{
			"nil node func",
			func() *Graph[state] {
				return New[state]().AddNode("a", nil).AddEdge(Start, "a").AddEdge("a", End)
			},
			"nil func",
		},
		{
			"nil router",
			func() *Graph[state] {
				return New[state]().AddNode("a", noop).AddEdge(Start, "a").
					AddRouter("a", nil).AddEdge("a", End)
			},
			"nil router",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.build().Compile()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestStreamBreak(t *testing.T) {
	g := New[state]().
		AddNode("loop", visit("loop", func(s state) state { s.N++; return s })).
		AddEdge(Start, "loop").
		AddEdge("loop", "loop")
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	ran := 0
	for _, err := range app.Stream(t.Context(), state{}) {
		if err != nil {
			t.Fatal(err)
		}
		if ran++; ran == 2 {
			break
		}
	}
	if ran != 2 {
		t.Errorf("ran = %d, want 2", ran)
	}
}

func TestContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	g := New[state]().
		AddNode("loop", func(_ context.Context, s state) (state, error) {
			s.N++
			if s.N == 2 {
				cancel()
			}
			return s, nil
		}).
		AddEdge(Start, "loop").
		AddEdge("loop", "loop")
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	final, err := app.Invoke(ctx, state{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if final.N != 2 {
		t.Errorf("last good state N = %d, want 2", final.N)
	}
}

func TestErrorStepRetriesFailedNode(t *testing.T) {
	attempts := 0
	g := New[state]().
		AddNode("a", visit("a", func(s state) state { s.N = 1; return s })).
		AddNode("flaky", func(_ context.Context, s state) (state, error) {
			if attempts++; attempts == 1 {
				return state{}, errors.New("transient")
			}
			return visit("flaky", func(s state) state { s.N++; return s })(t.Context(), s)
		}).
		AddEdge(Start, "a").
		AddEdge("a", "flaky").
		AddEdge("flaky", End)
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	var cp Step[state]
	for step, err := range app.Stream(t.Context(), state{}) {
		cp = step
		if err != nil {
			break
		}
	}
	// The error Step points at the last completed node, not the failed one.
	if cp.Node != "a" {
		t.Fatalf("error step node = %q, want a", cp.Node)
	}

	final, err := app.InvokeFrom(t.Context(), cp)
	if err != nil {
		t.Fatal(err)
	}
	// Retrying re-ran "flaky" instead of skipping it.
	if final.N != 2 || strings.Join(final.Path, ",") != "a,flaky" {
		t.Errorf("final = %+v, want N=2 Path=a,flaky", final)
	}
}

func TestAppIsolatedFromBuilder(t *testing.T) {
	g := New[state]().
		AddNode("a", visit("a", func(s state) state { return s })).
		AddEdge(Start, "a").
		AddEdge("a", End)
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	g.AddNode("bad", nil) // poison the builder after compiling
	if _, err := g.Compile(); err == nil {
		t.Fatal("second Compile should fail")
	}
	if _, err := app.Invoke(t.Context(), state{}); err != nil {
		t.Errorf("compiled runnable broke after builder mutation: %v", err)
	}
}

func TestConcurrentInvokes(t *testing.T) {
	g := New[state]().
		AddNode("inc", visit("inc", func(s state) state { s.N++; return s })).
		AddEdge(Start, "inc").
		AddRouter("inc", func(_ context.Context, s state) (string, error) {
			if s.N < 10 {
				return "inc", nil
			}
			return End, nil
		})
	app, err := g.Compile()
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			final, err := app.Invoke(t.Context(), state{})
			if err != nil || final.N != 10 {
				t.Errorf("final = %+v, err = %v", final, err)
			}
		}()
	}
	wg.Wait()
}
