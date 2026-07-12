package tomergraph

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func linearGraph(t *testing.T) *App[state] {
	t.Helper()
	app, err := New[state]().
		AddNode("a", visit("a", func(s state) state { s.N++; return s })).
		AddNode("b", visit("b", func(s state) state { s.N *= 10; return s })).
		AddEdge(Start, "a").
		AddEdge("a", "b").
		AddEdge("b", End).
		Compile()
	if err != nil {
		t.Fatal(err)
	}
	return app
}

func TestResumeFromMidRun(t *testing.T) {
	app := linearGraph(t)

	var cp Step[state]
	for step, err := range app.Stream(t.Context(), state{}) {
		if err != nil {
			t.Fatal(err)
		}
		cp = step
		break // stop after "a", keep its Step as a checkpoint
	}
	if cp.Node != "a" {
		t.Fatalf("checkpoint node = %q, want a", cp.Node)
	}

	final, err := app.InvokeFrom(t.Context(), cp)
	if err != nil {
		t.Fatal(err)
	}
	if final.N != 10 || strings.Join(final.Path, ",") != "a,b" {
		t.Errorf("resumed final = %+v, want N=10 Path=a,b", final)
	}
}

func TestResumeUnknownNode(t *testing.T) {
	app := linearGraph(t)
	_, err := app.InvokeFrom(t.Context(), Step[state]{Node: "ghost"})
	if err == nil || !strings.Contains(err.Error(), `unknown node "ghost"`) {
		t.Fatalf("err = %v, want resume error", err)
	}
}

func TestMemorySaver(t *testing.T) {
	saver := &MemorySaver[state]{}
	ctx := t.Context()

	if _, ok, err := saver.Load(ctx, "t1"); err != nil || ok {
		t.Fatalf("empty load: ok=%v err=%v, want miss", ok, err)
	}
	want := Step[state]{Node: "a", State: state{N: 3}}
	if err := saver.Save(ctx, "t1", want); err != nil {
		t.Fatal(err)
	}
	got, ok, err := saver.Load(ctx, "t1")
	if err != nil || !ok || got.Node != "a" || got.State.N != 3 {
		t.Fatalf("load = %+v ok=%v err=%v, want %+v", got, ok, err, want)
	}
}

func TestInvokeThread(t *testing.T) {
	app := linearGraph(t)
	saver := &MemorySaver[state]{}
	ctx := t.Context()

	final, err := app.InvokeThread(ctx, saver, "t1", state{})
	if err != nil || final.N != 10 {
		t.Fatalf("final = %+v, err = %v", final, err)
	}

	// Finished thread: rerunning is a no-op, initial is ignored.
	again, err := app.InvokeThread(ctx, saver, "t1", state{N: 999})
	if err != nil {
		t.Fatal(err)
	}
	if again.N != 10 || strings.Join(again.Path, ",") != "a,b" {
		t.Errorf("rerun = %+v, want unchanged final", again)
	}
}

func TestStreamThreadResumesAfterBreak(t *testing.T) {
	app := linearGraph(t)
	saver := &MemorySaver[state]{}
	ctx := t.Context()

	for _, err := range app.StreamThread(ctx, saver, "t1", state{}) {
		if err != nil {
			t.Fatal(err)
		}
		break // simulate dying after the first node; its Step is saved
	}

	final, err := app.InvokeThread(ctx, saver, "t1", state{})
	if err != nil {
		t.Fatal(err)
	}
	// "a" ran once before the crash and was not re-run after it.
	if final.N != 10 || strings.Join(final.Path, ",") != "a,b" {
		t.Errorf("final = %+v, want N=10 Path=a,b", final)
	}
}

type failingSaver struct{ *MemorySaver[state] }

func (f failingSaver) Save(context.Context, string, Step[state]) error {
	return errors.New("disk full")
}

func TestStreamThreadSaveError(t *testing.T) {
	app := linearGraph(t)
	saver := failingSaver{&MemorySaver[state]{}}

	_, err := app.InvokeThread(t.Context(), saver, "t1", state{})
	if err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("err = %v, want save error", err)
	}
}
