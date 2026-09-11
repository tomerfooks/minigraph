// Command reflection shows the generate → critique → revise loop: a writer
// drafts, a critic returns feedback (or nothing when satisfied), and the
// router loops back until the critic is happy or a round budget runs out.
// This is LangGraph's "reflection" tutorial as two nodes and one router. The
// writer and critic are scripted so the example runs offline.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Prompt   string
	Draft    string
	Critique string // empty means "approved"
	Round    int
}

const maxRounds = 3

func generate(_ context.Context, s State) (State, error) {
	s.Round++
	switch {
	case s.Critique == "":
		s.Draft = "Go is a programming language."
	case strings.Contains(s.Critique, "specific"):
		s.Draft = "Go is a compiled, statically typed language with goroutines for concurrency."
	default:
		s.Draft = "Go is a compiled, statically typed language with goroutines for concurrency and a single-binary deploy story."
	}
	return s, nil
}

func critique(_ context.Context, s State) (State, error) {
	switch {
	case len(s.Draft) < 40:
		s.Critique = "too vague — be specific about what makes Go different"
	case !strings.Contains(s.Draft, "binary"):
		s.Critique = "mention deployment: single static binary"
	default:
		s.Critique = "" // approved
	}
	return s, nil
}

func build() (*minigraph.App[State], error) {
	return minigraph.New[State]().
		AddNode("generate", generate).
		AddNode("critique", critique).
		AddEdge(minigraph.Start, "generate").
		AddEdge("generate", "critique").
		AddRouter("critique", func(_ context.Context, s State) (string, error) {
			if s.Critique == "" || s.Round >= maxRounds {
				return minigraph.End, nil
			}
			return "generate", nil
		}).
		Compile()
}

func main() {
	app, err := build()
	if err != nil {
		panic(err)
	}
	var final State
	for step, err := range app.Stream(context.Background(), State{Prompt: "describe Go in one sentence"}) {
		if err != nil {
			panic(err)
		}
		final = step.State
		if step.Node == "generate" {
			fmt.Printf("round %d draft:    %s\n", final.Round, final.Draft)
		} else if final.Critique != "" {
			fmt.Printf("round %d critique: %s\n", final.Round, final.Critique)
		}
	}
	fmt.Printf("approved after %d rounds: %s\n", final.Round, final.Draft)
}
