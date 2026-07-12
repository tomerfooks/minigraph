// Command react runs a ReAct (Reason + Act) agent on tomergraph.
//
// The loop is the textbook one: the model emits a Thought and an Action, a
// tool runs and appends an Observation, and the model reasons again over the
// growing trace — until it emits a Final Answer. The LLM here is a scripted
// mock so the example runs offline; swap `llm` for a real model call and the
// rest of the code is unchanged, including the output parsing.
package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/tomerfooks/tomergraph"
)

// LLM turns a prompt into a completion. Replace the mock with a real client.
type LLM func(ctx context.Context, prompt string) (string, error)

// Turn is one Thought → Action → Observation cycle in the trace.
type Turn struct {
	Thought     string
	Action      string // tool name; empty once the agent answers
	Input       string
	Observation string
}

// State is the agent's full memory, flowing through every node.
type State struct {
	Question string
	Trace    []Turn
	Answer   string
}

// tools maps action names to implementations.
var tools = map[string]func(input string) (string, error){
	"lookup": func(input string) (string, error) {
		facts := map[string]string{
			"population of France":             "68000000",
			"height of Eiffel Tower in meters": "330",
		}
		if fact, ok := facts[input]; ok {
			return fact, nil
		}
		return "", fmt.Errorf("no fact for %q", input)
	},
	"calculate": func(input string) (string, error) {
		f := strings.Fields(input) // "a op b"
		if len(f) != 3 {
			return "", fmt.Errorf("want %q, got %q", "a op b", input)
		}
		a, errA := strconv.ParseFloat(f[0], 64)
		b, errB := strconv.ParseFloat(f[2], 64)
		if errA != nil || errB != nil {
			return "", fmt.Errorf("bad operands in %q", input)
		}
		switch f[1] {
		case "+":
			return strconv.FormatFloat(a+b, 'f', -1, 64), nil
		case "-":
			return strconv.FormatFloat(a-b, 'f', -1, 64), nil
		case "*":
			return strconv.FormatFloat(a*b, 'f', -1, 64), nil
		case "/":
			return strconv.FormatFloat(a/b, 'f', -1, 64), nil
		}
		return "", fmt.Errorf("unknown operator %q", f[1])
	},
}

// prompt renders the question and trace the way a real ReAct prompt would.
func prompt(s State) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Question: %s\n", s.Question)
	for _, t := range s.Trace {
		fmt.Fprintf(&b, "Thought: %s\nAction: %s[%s]\nObservation: %s\n",
			t.Thought, t.Action, t.Input, t.Observation)
	}
	return b.String()
}

// parse extracts either an action turn or a final answer from a completion:
//
//	Thought: <text>
//	Action: <tool>[<input>]        -- or --        Final Answer: <text>
func parse(completion string) (Turn, string, error) {
	var turn Turn
	for _, line := range strings.Split(completion, "\n") {
		switch {
		case strings.HasPrefix(line, "Thought: "):
			turn.Thought = strings.TrimPrefix(line, "Thought: ")
		case strings.HasPrefix(line, "Final Answer: "):
			return turn, strings.TrimPrefix(line, "Final Answer: "), nil
		case strings.HasPrefix(line, "Action: "):
			call := strings.TrimPrefix(line, "Action: ")
			open := strings.Index(call, "[")
			if open < 0 || !strings.HasSuffix(call, "]") {
				return turn, "", fmt.Errorf("malformed action %q", call)
			}
			turn.Action = call[:open]
			turn.Input = call[open+1 : len(call)-1]
		}
	}
	if turn.Action == "" {
		return turn, "", fmt.Errorf("completion had no Action or Final Answer:\n%s", completion)
	}
	return turn, "", nil
}

// mockLLM plays a fixed two-tool episode, keyed off how many observations the
// prompt already contains — the shape a real model produces, without the API.
func mockLLM(_ context.Context, p string) (string, error) {
	switch strings.Count(p, "Observation: ") {
	case 0:
		return "Thought: I need France's population before I can double it.\n" +
			"Action: lookup[population of France]", nil
	case 1:
		return "Thought: Now I double the population I found.\n" +
			"Action: calculate[68000000 * 2]", nil
	default:
		return "Thought: I have the doubled figure.\n" +
			"Final Answer: Twice the population of France is 136000000.", nil
	}
}

func buildAgent(llm LLM) (*tomergraph.App[State], error) {
	g := tomergraph.New[State]()

	// reason: one LLM step. Either queues an action or produces the answer.
	g.AddNode("reason", func(ctx context.Context, s State) (State, error) {
		completion, err := llm(ctx, prompt(s))
		if err != nil {
			return s, err
		}
		turn, answer, err := parse(completion)
		if err != nil {
			return s, err
		}
		if answer != "" {
			s.Answer = answer
			return s, nil
		}
		s.Trace = append(s.Trace, turn)
		return s, nil
	})

	// act: run the queued tool, write its observation into the trace. Tool
	// failures become observations too, so the model can react to them.
	g.AddNode("act", func(ctx context.Context, s State) (State, error) {
		turn := &s.Trace[len(s.Trace)-1]
		tool, ok := tools[turn.Action]
		if !ok {
			turn.Observation = fmt.Sprintf("error: unknown tool %q", turn.Action)
			return s, nil
		}
		out, err := tool(turn.Input)
		if err != nil {
			turn.Observation = "error: " + err.Error()
			return s, nil
		}
		turn.Observation = out
		return s, nil
	})

	g.AddEdge(tomergraph.Start, "reason")
	g.AddRouter("reason", func(_ context.Context, s State) (string, error) {
		if s.Answer != "" {
			return tomergraph.End, nil
		}
		return "act", nil
	})
	g.AddEdge("act", "reason")

	return g.Compile()
}

func main() {
	app, err := buildAgent(mockLLM)
	if err != nil {
		panic(err)
	}

	initial := State{Question: "What is twice the population of France?"}
	fmt.Println("Question:", initial.Question)

	var final State
	for step, err := range app.Stream(context.Background(), initial) {
		if err != nil {
			panic(err)
		}
		final = step.State
		switch step.Node {
		case "reason":
			if final.Answer != "" {
				continue // printed below
			}
			t := final.Trace[len(final.Trace)-1]
			fmt.Printf("Thought: %s\nAction: %s[%s]\n", t.Thought, t.Action, t.Input)
		case "act":
			fmt.Printf("Observation: %s\n", final.Trace[len(final.Trace)-1].Observation)
		}
	}
	fmt.Println("Final Answer:", final.Answer)
}
