// Package minigraph is a minimal LangGraph for Go: a state machine of nodes
// that transform a shared, typed state, wired with static or conditional
// edges. Build a Graph, Compile it, then Invoke or Stream.
package minigraph

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"maps"
	"slices"
)

// Start and End are the reserved endpoints of every run: wire Start to your
// entry node, and route to End to finish.
const (
	Start = "__start__"
	End   = "__end__"
)

// ErrMaxSteps is returned when a run exceeds App.MaxSteps node
// executions — usually a cycle that never routes to End.
var ErrMaxSteps = errors.New("minigraph: max steps exceeded")

// Node transforms the state: it receives the current state and returns the
// next one.
type Node[S any] func(ctx context.Context, state S) (S, error)

// Router inspects the state after a node ran and names the next node, or End.
type Router[S any] func(ctx context.Context, state S) (string, error)

type edge[S any] struct {
	to     string    // static target; empty when conditional
	router Router[S] // conditional target; nil when static
}

// Graph is a mutable builder. Add nodes and edges, then Compile. Building
// mistakes accumulate silently and are all reported by Compile.
type Graph[S any] struct {
	nodes map[string]Node[S]
	edges map[string]edge[S]
	errs  []error
}

// New returns an empty graph over state type S.
func New[S any]() *Graph[S] {
	return &Graph[S]{nodes: map[string]Node[S]{}, edges: map[string]edge[S]{}}
}

func (g *Graph[S]) fail(format string, args ...any) *Graph[S] {
	g.errs = append(g.errs, fmt.Errorf(format, args...))
	return g
}

// AddNode registers fn under name. Names must be unique and not Start or End.
func (g *Graph[S]) AddNode(name string, fn Node[S]) *Graph[S] {
	if name == "" || name == Start || name == End {
		return g.fail("AddNode: reserved or empty name %q", name)
	}
	if fn == nil {
		return g.fail("AddNode %q: nil func", name)
	}
	if _, dup := g.nodes[name]; dup {
		return g.fail("AddNode: duplicate node %q", name)
	}
	g.nodes[name] = fn
	return g
}

// AddEdge wires from → to unconditionally. Every node has exactly one
// outgoing edge; from may be Start, to may be End.
func (g *Graph[S]) AddEdge(from, to string) *Graph[S] {
	return g.addEdge(from, edge[S]{to: to})
}

// AddRouter wires from → r(state), letting the state pick the next
// node at run time. r must return a node name or End.
func (g *Graph[S]) AddRouter(from string, r Router[S]) *Graph[S] {
	if r == nil {
		return g.fail("AddRouter %q: nil router", from)
	}
	return g.addEdge(from, edge[S]{router: r})
}

func (g *Graph[S]) addEdge(from string, e edge[S]) *Graph[S] {
	if _, dup := g.edges[from]; dup {
		return g.fail("edge from %q already set", from)
	}
	g.edges[from] = e
	return g
}

// Compile validates the graph and freezes it into a App. The builder can
// keep changing afterwards without affecting compiled runnables.
func (g *Graph[S]) Compile() (*App[S], error) {
	errs := slices.Clone(g.errs)
	check := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf(format, args...))
	}

	if _, ok := g.edges[Start]; !ok {
		check("no entry edge: AddEdge(Start, ...) missing")
	}
	for _, name := range slices.Sorted(maps.Keys(g.nodes)) {
		if _, ok := g.edges[name]; !ok {
			check("node %q has no outgoing edge (route it to End to finish)", name)
		}
	}
	for _, from := range slices.Sorted(maps.Keys(g.edges)) {
		e := g.edges[from]
		if _, ok := g.nodes[from]; !ok && from != Start {
			check("edge from unknown node %q", from)
		}
		if e.router == nil && e.to != End {
			if _, ok := g.nodes[e.to]; !ok {
				check("edge %q -> %q: unknown target", from, e.to)
			}
		}
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	return &App[S]{
		nodes:    maps.Clone(g.nodes),
		edges:    maps.Clone(g.edges),
		MaxSteps: 25,
	}, nil
}

// App is a compiled, immutable graph. Safe for concurrent use.
type App[S any] struct {
	nodes map[string]Node[S]
	edges map[string]edge[S]

	// MaxSteps caps node executions per run, catching endless cycles.
	// Compile sets it to 25; override freely.
	MaxSteps int
}

// Step is one node execution: the node's name and the state it produced.
// Every Step is also a checkpoint — feed one to InvokeFrom or StreamFrom to
// continue a run from exactly that point.
type Step[S any] struct {
	Node  string
	State S
}

// Stream runs the graph from Start, yielding after every node. An error
// arrives as the final pair, with the Step beside it carrying the last
// completed node and last good state — so resuming an error Step retries the
// step that failed. Breaking out of the loop stops the run.
func (r *App[S]) Stream(ctx context.Context, state S) iter.Seq2[Step[S], error] {
	return r.StreamFrom(ctx, Step[S]{Node: Start, State: state})
}

// StreamFrom continues a run from a previously yielded Step: routing
// restarts from from.Node with from.State, so the checkpointed node is not
// re-executed. MaxSteps counts from zero again.
func (r *App[S]) StreamFrom(ctx context.Context, from Step[S]) iter.Seq2[Step[S], error] {
	return func(yield func(Step[S], error) bool) {
		at, state := from.Node, from.State
		if _, ok := r.nodes[at]; !ok && at != Start {
			yield(from, fmt.Errorf("resume from unknown node %q", at))
			return
		}
		for steps := 0; ; steps++ {
			next, err := r.route(ctx, at, state)
			if err != nil {
				yield(Step[S]{Node: at, State: state}, err)
				return
			}
			if next == End {
				return
			}
			if err := ctx.Err(); err != nil {
				yield(Step[S]{Node: at, State: state}, err)
				return
			}
			if steps >= r.MaxSteps {
				yield(Step[S]{Node: at, State: state}, fmt.Errorf("%w (%d)", ErrMaxSteps, r.MaxSteps))
				return
			}
			out, err := r.nodes[next](ctx, state)
			if intr := (*Interrupt)(nil); errors.As(err, &intr) {
				// A pause, not a failure: the returned state is kept, and the
				// yielded Step is the checkpoint to InvokeFrom from.
				yield(Step[S]{Node: next, State: out}, &Interrupt{Payload: intr.Payload, Node: next})
				return
			}
			if err != nil {
				// The failed node did not complete: the Step points at the
				// last completed node, so resuming it retries the failure.
				yield(Step[S]{Node: at, State: state}, fmt.Errorf("node %q: %w", next, err))
				return
			}
			state = out
			if !yield(Step[S]{Node: next, State: state}, nil) {
				return
			}
			at = next
		}
	}
}

// Invoke runs the graph to completion and returns the final state. On error
// it returns the last state a node produced successfully.
func (r *App[S]) Invoke(ctx context.Context, state S) (S, error) {
	return r.InvokeFrom(ctx, Step[S]{Node: Start, State: state})
}

// InvokeFrom runs the graph to completion from a previously yielded Step.
func (r *App[S]) InvokeFrom(ctx context.Context, from Step[S]) (S, error) {
	final := from.State
	for step, err := range r.StreamFrom(ctx, from) {
		final = step.State // on error, the Step carries the last good state
		if err != nil {
			return final, err
		}
	}
	return final, nil
}

// route resolves the outgoing edge of at against the current state.
func (r *App[S]) route(ctx context.Context, at string, state S) (string, error) {
	e := r.edges[at]
	if e.router == nil {
		return e.to, nil
	}
	to, err := e.router(ctx, state)
	if err != nil {
		return "", fmt.Errorf("edge from %q: %w", at, err)
	}
	if _, ok := r.nodes[to]; !ok && to != End {
		return "", fmt.Errorf("edge from %q: routed to unknown node %q", at, to)
	}
	return to, nil
}
