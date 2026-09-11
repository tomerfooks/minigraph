// Command server embeds an agent in a plain net/http service and streams
// every step to the client as server-sent events. Nothing here is specific to
// MiniGraph beyond ranging over Stream: the request context cancels the run
// when the client disconnects, and each Step is one event.
//
//	go run ./examples/server &
//	curl -N 'localhost:8080/run?q=twice+the+population+of+France'
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Question string
	Steps    []string
	Answer   string
}

func build() (*minigraph.App[State], error) {
	return minigraph.New[State]().
		AddNode("think", func(ctx context.Context, s State) (State, error) {
			time.Sleep(300 * time.Millisecond) // stand-in for a model call
			s.Steps = append(s.Steps, "thinking about: "+s.Question)
			return s, nil
		}).
		AddNode("lookup", func(ctx context.Context, s State) (State, error) {
			time.Sleep(300 * time.Millisecond)
			s.Steps = append(s.Steps, "lookup: population of France = 68,000,000")
			return s, nil
		}).
		AddNode("answer", func(ctx context.Context, s State) (State, error) {
			s.Answer = "136,000,000"
			return s, nil
		}).
		AddEdge(minigraph.Start, "think").
		AddRouter("think", func(_ context.Context, s State) (string, error) {
			if strings.Contains(s.Question, "France") {
				return "lookup", nil
			}
			return "answer", nil
		}).
		AddEdge("lookup", "answer").
		AddEdge("answer", minigraph.End).
		Compile()
}

func main() {
	app, err := build()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		flush, _ := w.(http.Flusher)
		// r.Context() is cancelled when the client goes away; the run stops
		// at the next node boundary and the handler returns.
		for step, err := range app.Stream(r.Context(), State{Question: r.URL.Query().Get("q")}) {
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: %q\n\n", err.Error())
				return
			}
			b, _ := json.Marshal(step)
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", step.Node, b)
			if flush != nil {
				flush.Flush()
			}
		}
		fmt.Fprint(w, "event: done\ndata: {}\n\n")
	})
	log.Println("listening on :8080 — try: curl -N 'localhost:8080/run?q=twice+the+population+of+France'")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
