// Command mapreduce shows dynamic fan-out: split a document into chunks,
// summarize every chunk concurrently, then reduce the summaries. The number
// of branches is decided at run time from the state — LangGraph's Send —
// by building the Parallel node inside a node. Each "summarize" sleeps to
// stand in for a model call, so the wall-clock time shows the concurrency.
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Doc       string
	Chunks    []string
	Summaries []string
	Summary   string
}

func split(_ context.Context, s State) (State, error) {
	s.Chunks = strings.Split(strings.TrimSpace(s.Doc), "\n\n")
	return s, nil
}

// summarize returns a branch that handles chunk i. Branches get a shallow
// copy of the state, so each writes only its own slot in a fresh slice.
func summarize(i int) minigraph.Node[State] {
	return func(ctx context.Context, s State) (State, error) {
		select {
		case <-time.After(200 * time.Millisecond): // "the model call"
		case <-ctx.Done():
			return s, ctx.Err()
		}
		words := strings.Fields(s.Chunks[i])
		s.Summaries = []string{fmt.Sprintf("§%d: %s… (%d words)", i+1, strings.Join(words[:3], " "), len(words))}
		return s, nil
	}
}

// fanout builds one branch per chunk and runs them as a single Parallel step.
func fanout(ctx context.Context, s State) (State, error) {
	branches := make([]minigraph.Node[State], len(s.Chunks))
	for i := range s.Chunks {
		branches[i] = summarize(i)
	}
	merge := func(_ context.Context, base State, results []State) (State, error) {
		base.Summaries = nil
		for _, r := range results { // results keep branch order
			base.Summaries = append(base.Summaries, r.Summaries...)
		}
		return base, nil
	}
	return minigraph.Parallel(merge, branches...)(ctx, s)
}

func reduce(_ context.Context, s State) (State, error) {
	s.Summary = strings.Join(s.Summaries, "\n")
	return s, nil
}

func build() (*minigraph.App[State], error) {
	return minigraph.New[State]().
		AddNode("split", split).
		AddNode("map", fanout).
		AddNode("reduce", reduce).
		AddEdge(minigraph.Start, "split").
		AddEdge("split", "map").
		AddEdge("map", "reduce").
		AddEdge("reduce", minigraph.End).
		Compile()
}

const doc = `
Go was designed at Google in 2007 by Robert Griesemer, Rob Pike, and Ken Thompson.

The language is statically typed and compiles to a single native binary.

Goroutines and channels make concurrent programs easy to write and read.

The standard library covers HTTP, JSON, crypto, and testing out of the box.

Generics arrived in Go 1.18 and range-over-func iterators in Go 1.23.
`

func main() {
	app, err := build()
	if err != nil {
		panic(err)
	}
	start := time.Now()
	final, err := app.Invoke(context.Background(), State{Doc: doc})
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d chunks summarized in %s (each takes 200ms)\n", len(final.Chunks), time.Since(start).Round(10*time.Millisecond))
	fmt.Println(final.Summary)
}
