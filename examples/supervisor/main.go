// Command supervisor shows the multi-agent "supervisor" pattern: one node
// looks at the state and names the specialist that should work next; the
// specialists are ordinary nodes that hand control back. In LangGraph this is
// the supervisor + workers graph; here it is one router and a cycle. The
// supervisor's "LLM" is a scripted policy so the example runs offline.
package main

import (
	"context"
	"fmt"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Goal     string
	Research string
	Code     string
	Report   string
	Next     string   // the supervisor's decision
	Log      []string // who did what, in order
}

// supervise is the decision node. Swap the policy for a model call that
// returns one of the worker names or "FINISH".
func supervise(_ context.Context, s State) (State, error) {
	switch {
	case s.Research == "":
		s.Next = "researcher"
	case s.Code == "":
		s.Next = "coder"
	case s.Report == "":
		s.Next = "writer"
	default:
		s.Next = "FINISH"
	}
	s.Log = append(s.Log, "supervisor → "+s.Next)
	return s, nil
}

func researcher(_ context.Context, s State) (State, error) {
	s.Research = "Go generics (1.18+) allow func Map[T, U any](xs []T, f func(T) U) []U"
	s.Log = append(s.Log, "researcher: gathered notes")
	return s, nil
}

func coder(_ context.Context, s State) (State, error) {
	s.Code = "func Map[T, U any](xs []T, f func(T) U) []U { out := make([]U, len(xs)); for i, x := range xs { out[i] = f(x) }; return out }"
	s.Log = append(s.Log, "coder: wrote Map")
	return s, nil
}

func writer(_ context.Context, s State) (State, error) {
	s.Report = fmt.Sprintf("Goal: %s\nNotes: %s\nCode: %s", s.Goal, s.Research, s.Code)
	s.Log = append(s.Log, "writer: assembled report")
	return s, nil
}

func build() (*minigraph.App[State], error) {
	g := minigraph.New[State]().
		AddNode("supervisor", supervise).
		AddNode("researcher", researcher).
		AddNode("coder", coder).
		AddNode("writer", writer).
		AddEdge(minigraph.Start, "supervisor").
		AddRouter("supervisor", func(_ context.Context, s State) (string, error) {
			if s.Next == "FINISH" {
				return minigraph.End, nil
			}
			return s.Next, nil // the state names the next node
		})
	for _, w := range []string{"researcher", "coder", "writer"} {
		g.AddEdge(w, "supervisor") // every worker reports back
	}
	return g.Compile()
}

func main() {
	app, err := build()
	if err != nil {
		panic(err)
	}
	final, err := app.Invoke(context.Background(), State{Goal: "write a generic Map helper"})
	if err != nil {
		panic(err)
	}
	for _, l := range final.Log {
		fmt.Println(l)
	}
	fmt.Println("\n" + final.Report)
}
