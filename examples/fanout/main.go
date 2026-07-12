// Command fanout shows parallel fan-out/join: three researchers run
// concurrently as one graph step via Parallel, a merge folds their findings,
// and a writer summarizes. Branches write scalar fields on their own state
// copy; the merge reconciles — no reducers, no shared writes.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Topic    string
	Finding  string   // set by each branch on its own copy
	Findings []string // folded together by the merge
	Report   string
}

func researcher(source string) minigraph.Node[State] {
	return func(ctx context.Context, s State) (State, error) {
		s.Finding = fmt.Sprintf("%s says %q is promising", source, s.Topic)
		return s, nil
	}
}

func main() {
	research := minigraph.Parallel(
		func(_ context.Context, base State, results []State) (State, error) {
			for _, r := range results {
				base.Findings = append(base.Findings, r.Finding)
			}
			return base, nil
		},
		researcher("web"), researcher("docs"), researcher("db"),
	)

	g := minigraph.New[State]()
	g.AddNode("research", research)
	g.AddNode("write", func(ctx context.Context, s State) (State, error) {
		s.Report = fmt.Sprintf("%d sources agree: %s", len(s.Findings), strings.Join(s.Findings, "; "))
		return s, nil
	})
	g.AddEdge(minigraph.Start, "research")
	g.AddEdge("research", "write")
	g.AddEdge("write", minigraph.End)

	app, err := g.Compile()
	if err != nil {
		panic(err)
	}

	for step, err := range app.Stream(context.Background(), State{Topic: "solar sails"}) {
		if err != nil {
			panic(err)
		}
		fmt.Printf("── %s\n", step.Node)
		if step.Node == "write" {
			fmt.Println(step.State.Report)
		}
	}
}
