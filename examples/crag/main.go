// Command crag shows corrective RAG: retrieve documents, grade their
// relevance, and if they miss, rewrite the query and retrieve again before
// generating. The retriever is a keyword index over a tiny corpus and the
// grader/rewriter are scripted, so it runs offline; each node is the place
// to drop in a vector store or a model call.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Question string
	Query    string
	Docs     []string
	Relevant bool
	Rewrites int
	Answer   string
}

var corpus = []string{
	"Type parameters are declared in square brackets after the function name.",
	"A type constraint is an interface that restricts which types may be used as a type argument.",
	"Goroutines are lightweight threads managed by the Go runtime.",
}

func retrieve(_ context.Context, s State) (State, error) {
	s.Docs = nil
	for _, d := range corpus {
		if strings.Contains(strings.ToLower(d), s.Query) {
			s.Docs = append(s.Docs, d)
		}
	}
	return s, nil
}

func grade(_ context.Context, s State) (State, error) {
	s.Relevant = len(s.Docs) > 0
	return s, nil
}

func rewrite(_ context.Context, s State) (State, error) {
	s.Rewrites++
	s.Query = "type constraint" // a model would rephrase; we know the answer
	return s, nil
}

func generate(_ context.Context, s State) (State, error) {
	if !s.Relevant {
		s.Answer = "I couldn't find anything relevant."
		return s, nil
	}
	s.Answer = "Based on the docs: " + s.Docs[0]
	return s, nil
}

func build() (*minigraph.App[State], error) {
	return minigraph.New[State]().
		AddNode("retrieve", retrieve).
		AddNode("grade", grade).
		AddNode("rewrite", rewrite).
		AddNode("generate", generate).
		AddEdge(minigraph.Start, "retrieve").
		AddEdge("retrieve", "grade").
		AddRouter("grade", func(_ context.Context, s State) (string, error) {
			if s.Relevant || s.Rewrites >= 2 {
				return "generate", nil
			}
			return "rewrite", nil
		}).
		AddEdge("rewrite", "retrieve").
		AddEdge("generate", minigraph.End).
		Compile()
}

func main() {
	app, err := build()
	if err != nil {
		panic(err)
	}
	q := "How do I restrict a type parameter?"
	var final State
	for step, err := range app.Stream(context.Background(), State{Question: q, Query: "restrict a type parameter"}) {
		if err != nil {
			panic(err)
		}
		final = step.State
		switch step.Node {
		case "retrieve":
			fmt.Printf("retrieve %q → %d docs\n", final.Query, len(final.Docs))
		case "grade":
			fmt.Println("grade    → relevant:", final.Relevant)
		case "rewrite":
			fmt.Printf("rewrite  → %q\n", final.Query)
		}
	}
	fmt.Println("answer   →", final.Answer)
}
