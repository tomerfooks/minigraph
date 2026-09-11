// Command durable shows crash-and-resume across two processes with a
// Checkpointer that writes JSON files — no database, no server. The first
// run of a thread "crashes" while delivering an order; run the program again
// and it continues from the last saved step without redoing the earlier work.
//
//	go run ./examples/durable   # fetch, charge, then crash at deliver
//	go run ./examples/durable   # resumes at deliver, finishes, cleans up
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Order   string
	Items   []string
	Charged bool
	Shipped bool
}

// FileSaver is a Checkpointer: one JSON file per thread.
type FileSaver[S any] struct{ Dir string }

func (f FileSaver[S]) path(thread string) string { return filepath.Join(f.Dir, thread+".json") }

func (f FileSaver[S]) Save(_ context.Context, thread string, step minigraph.Step[S]) error {
	b, err := json.Marshal(step)
	if err != nil {
		return err
	}
	return os.WriteFile(f.path(thread), b, 0o644)
}

func (f FileSaver[S]) Load(_ context.Context, thread string) (minigraph.Step[S], bool, error) {
	var step minigraph.Step[S]
	b, err := os.ReadFile(f.path(thread))
	if errors.Is(err, os.ErrNotExist) {
		return step, false, nil
	}
	if err != nil {
		return step, false, err
	}
	return step, true, json.Unmarshal(b, &step)
}

func build(crash bool) (*minigraph.App[State], error) {
	return minigraph.New[State]().
		AddNode("fetch", func(_ context.Context, s State) (State, error) {
			fmt.Println("fetch:   loading items for", s.Order)
			s.Items = []string{"keyboard", "mouse"}
			return s, nil
		}).
		AddNode("charge", func(_ context.Context, s State) (State, error) {
			fmt.Println("charge:  card charged")
			s.Charged = true
			return s, nil
		}).
		AddNode("deliver", func(_ context.Context, s State) (State, error) {
			if crash {
				return s, errors.New("carrier API: connection reset")
			}
			fmt.Println("deliver: shipped", s.Items)
			s.Shipped = true
			return s, nil
		}).
		AddEdge(minigraph.Start, "fetch").
		AddEdge("fetch", "charge").
		AddEdge("charge", "deliver").
		AddEdge("deliver", minigraph.End).
		Compile()
}

func main() {
	ctx := context.Background()
	saver := FileSaver[State]{Dir: os.TempDir()}
	const thread = "minigraph-order-7"

	// First process of a thread: no checkpoint yet, so deliver will fail.
	_, resuming, _ := saver.Load(ctx, thread)
	app, err := build(!resuming)
	if err != nil {
		panic(err)
	}
	if resuming {
		fmt.Println("checkpoint found at", saver.path(thread), "— resuming")
	}

	final, err := app.InvokeThread(ctx, saver, thread, State{Order: "#7"})
	if err != nil {
		fmt.Println("crash:  ", err)
		fmt.Println("         checkpoint kept; run again to resume")
		os.Exit(1)
	}
	fmt.Printf("done:    charged=%v shipped=%v\n", final.Charged, final.Shipped)
	os.Remove(saver.path(thread)) // thread finished; forget it
}
