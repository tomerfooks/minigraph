// Command approval shows human-in-the-loop with interrupts and a durable
// thread: an agent drafts an email, pauses for approval, incorporates
// rejection feedback, and only sends once a human says yes. The "human" is
// scripted so the example runs offline; each InvokeThread call could just as
// well happen in a different process — the thread checkpoint carries the run.
package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/tomerfooks/tomergraph"
)

type State struct {
	Recipient string
	Feedback  string // rejection feedback for the next draft
	Draft     string
	Approved  bool
	Sent      bool
}

func build() (*tomergraph.App[State], error) {
	g := tomergraph.New[State]()

	g.AddNode("draft", func(ctx context.Context, s State) (State, error) {
		s.Draft = fmt.Sprintf("Dear %s, our Q3 numbers are strong.", s.Recipient)
		if s.Feedback != "" {
			s.Draft = fmt.Sprintf("Hi %s — Q3 was strong. (revised per: %s)", s.Recipient, s.Feedback)
		}
		return s, nil
	})

	// approve: always pause and ask. InvokeFrom routes onward from here, so the
	// router below decides what the human's edit means.
	g.AddNode("approve", func(ctx context.Context, s State) (State, error) {
		return s, &tomergraph.Interrupt{Payload: "send this? " + s.Draft}
	})

	g.AddNode("send", func(ctx context.Context, s State) (State, error) {
		s.Sent = true
		return s, nil
	})

	g.AddEdge(tomergraph.Start, "draft")
	g.AddEdge("draft", "approve")
	g.AddRouter("approve", func(_ context.Context, s State) (string, error) {
		if s.Approved {
			return "send", nil
		}
		return "draft", nil // rejected: redraft with the feedback
	})
	g.AddEdge("send", tomergraph.End)

	return g.Compile()
}

func main() {
	app, err := build()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	saver := &tomergraph.MemorySaver[State]{}
	const thread = "email-42"

	// The scripted human: rejects the first draft, approves the second.
	answers := []func(s State) State{
		func(s State) State { s.Feedback = "too formal, make it shorter"; return s },
		func(s State) State { s.Approved = true; return s },
	}

	state := State{Recipient: "Ada"}
	for {
		state, err = app.InvokeThread(ctx, saver, thread, state)
		var intr *tomergraph.Interrupt
		if !errors.As(err, &intr) {
			break // finished (or a real error)
		}
		fmt.Println("agent asks:", intr.Payload)

		state = answers[0](state) // the human answers by editing the state
		answers = answers[1:]
		fmt.Printf("human answers: approved=%v feedback=%q\n\n", state.Approved, state.Feedback)

		// Store the answer under the interrupted node; the next InvokeThread
		// resumes there.
		if err := saver.Save(ctx, thread, tomergraph.Step[State]{Node: intr.Node, State: state}); err != nil {
			panic(err)
		}
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("sent=%v final draft: %s\n", state.Sent, state.Draft)
}
