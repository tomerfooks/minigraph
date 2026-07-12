// Command agent shows the classic LangGraph shape — an agent that keeps
// calling tools until it can answer — as a tomergraph over a typed state.
// The "LLM" is mocked so the example runs offline.
package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tomerfooks/tomergraph"
)

// State flows through every node. Typed — no map[string]any.
type State struct {
	Question string
	Pending  []string // tool calls the agent still wants
	Results  []int    // tool outputs gathered so far
	Answer   string
}

func main() {
	g := tomergraph.New[State]()

	// agent: decide to call tools or answer. A real one would call an LLM.
	g.AddNode("agent", func(ctx context.Context, s State) (State, error) {
		if len(s.Results) == 0 {
			s.Pending = []string{"add 2 3", "mul 4 5"}
			return s, nil
		}
		sum := 0
		for _, r := range s.Results {
			sum += r
		}
		s.Answer = fmt.Sprintf("tool results %v combine to %d", s.Results, sum)
		return s, nil
	})

	// tools: execute every pending call, feed results back to the agent.
	g.AddNode("tools", func(ctx context.Context, s State) (State, error) {
		for _, call := range s.Pending {
			f := strings.Fields(call)
			a, _ := strconv.Atoi(f[1])
			b, _ := strconv.Atoi(f[2])
			switch f[0] {
			case "add":
				s.Results = append(s.Results, a+b)
			case "mul":
				s.Results = append(s.Results, a*b)
			default:
				return s, fmt.Errorf("unknown tool %q", f[0])
			}
		}
		s.Pending = nil
		return s, nil
	})

	g.AddEdge(tomergraph.Start, "agent")
	g.AddRouter("agent", func(ctx context.Context, s State) (string, error) {
		if len(s.Pending) > 0 {
			return "tools", nil
		}
		return tomergraph.End, nil
	})
	g.AddEdge("tools", "agent")

	app, err := g.Compile()
	if err != nil {
		panic(err)
	}

	initial := State{Question: "what do add(2,3) and mul(4,5) make together?"}
	for step, err := range app.Stream(context.Background(), initial) {
		if err != nil {
			panic(err)
		}
		fmt.Printf("── %s\n   %+v\n", step.Node, step.State)
	}
}
