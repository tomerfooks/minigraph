package tomergraph

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// gateGraph pauses at "gate" until the state carries N >= 10.
func gateGraph(t *testing.T) *App[state] {
	t.Helper()
	app, err := New[state]().
		AddNode("prep", visit("prep", func(s state) state { s.N = 1; return s })).
		AddNode("gate", func(_ context.Context, s state) (state, error) {
			s.Path = append(s.Path, "gate")
			if s.N < 10 {
				return s, &Interrupt{Payload: "need N >= 10"}
			}
			return s, nil
		}).
		AddNode("done", visit("done", func(s state) state { s.N *= 2; return s })).
		AddEdge(Start, "prep").
		AddEdge("prep", "gate").
		AddRouter("gate", func(_ context.Context, s state) (string, error) {
			if s.N < 10 {
				return "gate", nil // re-ask after a rejected resume
			}
			return "done", nil
		}).
		AddEdge("done", End).
		Compile()
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestInterruptAndResume(t *testing.T) {
	app := gateGraph(t)
	ctx := t.Context()

	paused, err := app.Invoke(ctx, state{})
	var intr *Interrupt
	if !errors.As(err, &intr) {
		t.Fatalf("err = %v, want *Interrupt", err)
	}
	if intr.Node != "gate" || intr.Payload != "need N >= 10" {
		t.Errorf("interrupt = %+v, want gate / payload", intr)
	}
	// State returned alongside the Interrupt is kept, including the node's
	// own writes before pausing.
	if strings.Join(paused.Path, ",") != "prep,gate" {
		t.Errorf("paused path = %v, want prep,gate", paused.Path)
	}

	paused.N = 10 // the human answers
	final, err := app.InvokeFrom(ctx, Step[state]{Node: intr.Node, State: paused})
	if err != nil {
		t.Fatal(err)
	}
	if final.N != 20 {
		t.Errorf("final N = %d, want 20", final.N)
	}
	// InvokeFrom routed onward from gate without re-running it.
	if strings.Join(final.Path, ",") != "prep,gate,done" {
		t.Errorf("final path = %v, want prep,gate,done", final.Path)
	}
}

func TestInterruptWithThread(t *testing.T) {
	app := gateGraph(t)
	saver := &MemorySaver[state]{}
	ctx := t.Context()

	paused, err := app.InvokeThread(ctx, saver, "t1", state{})
	var intr *Interrupt
	if !errors.As(err, &intr) {
		t.Fatalf("err = %v, want *Interrupt", err)
	}

	// The interrupt Step was saved: a fresh process could pick it up.
	cp, ok, err := saver.Load(ctx, "t1")
	if err != nil || !ok || cp.Node != "gate" {
		t.Fatalf("saved checkpoint = %+v ok=%v err=%v, want gate", cp, ok, err)
	}

	// Answer: edit the state, save it under the interrupted node, rerun.
	paused.N = 10
	if err := saver.Save(ctx, "t1", Step[state]{Node: intr.Node, State: paused}); err != nil {
		t.Fatal(err)
	}
	final, err := app.InvokeThread(ctx, saver, "t1", state{})
	if err != nil {
		t.Fatal(err)
	}
	if final.N != 20 || strings.Join(final.Path, ",") != "prep,gate,done" {
		t.Errorf("final = %+v, want N=20 Path=prep,gate,done", final)
	}
}
