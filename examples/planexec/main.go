// Command planexec shows plan-and-execute: a planner writes a list of steps,
// an executor runs one step per turn, and a replanner decides whether the
// objective is met or more steps are needed. The "LLM" calls are scripted so
// the example runs offline; every node is a plain function you can swap.
package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Objective string
	Plan      []string // steps still to do
	Done      []string // "step => result"
	Answer    string
}

func plan(_ context.Context, s State) (State, error) {
	s.Plan = []string{
		"count the RSVPs for the meetup",
		"look up the usual no-show rate",
		"multiply to get expected attendance",
	}
	return s, nil
}

// execute pops the first step and runs it. Real tools go here.
func execute(_ context.Context, s State) (State, error) {
	step := s.Plan[0]
	s.Plan = s.Plan[1:]
	var result string
	switch {
	case strings.Contains(step, "RSVPs"):
		result = "48"
	case strings.Contains(step, "no-show"):
		result = "25%"
	case strings.Contains(step, "multiply"):
		result = "36"
	}
	s.Done = append(s.Done, step+" => "+result)
	return s, nil
}

// replan looks at progress. Here it finishes when the plan is empty; a model
// could also append new steps when a result was surprising.
func replan(_ context.Context, s State) (State, error) {
	if len(s.Plan) == 0 {
		s.Answer = "Expect about 36 people; book the 40-seat room."
	}
	return s, nil
}

func build() (*minigraph.App[State], error) {
	return minigraph.New[State]().
		AddNode("plan", plan).
		AddNode("execute", execute).
		AddNode("replan", replan).
		AddEdge(minigraph.Start, "plan").
		AddEdge("plan", "execute").
		AddEdge("execute", "replan").
		AddRouter("replan", func(_ context.Context, s State) (string, error) {
			if s.Answer != "" {
				return minigraph.End, nil
			}
			return "execute", nil
		}).
		Compile()
}

func main() {
	app, err := build()
	if err != nil {
		panic(err)
	}
	initial := State{Objective: "How many seats do we need for the meetup?"}
	fmt.Println("Objective:", initial.Objective)
	var final State
	for step, err := range app.Stream(context.Background(), initial) {
		if err != nil {
			panic(err)
		}
		final = step.State
		switch step.Node {
		case "plan":
			fmt.Printf("Plan: %d steps\n", len(final.Plan))
		case "execute":
			fmt.Println("  done:", final.Done[len(final.Done)-1])
		case "replan":
			fmt.Printf("Replan: %d left\n", len(final.Plan))
		}
	}
	fmt.Println("Answer:", final.Answer)
}
