// Command llm is the "swap in a real model" demo. Nodes call llm(), which is
// a scripted mock unless ANTHROPIC_API_KEY is set — then it POSTs to the
// Messages API with net/http. No SDK, no adapter layer: a node is a function,
// and a model call is an HTTP request inside it. The graph is a two-round
// draft → tighten chain; the shape does not change when the model does.
//
//	go run ./examples/llm                              # offline mock
//	ANTHROPIC_API_KEY=sk-ant-... go run ./examples/llm # real model
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/tomerfooks/minigraph"
)

type State struct {
	Topic string
	Draft string
	Final string
}

// llm is the only place that knows about a provider.
var llm = mock

func mock(_ context.Context, prompt string) (string, error) {
	if strings.HasPrefix(prompt, "Tighten") {
		return "MiniGraph: LangGraph's ideas, 420 lines of Go, zero deps.", nil
	}
	return "MiniGraph brings LangGraph's graph-of-nodes model to Go in about 420 lines with no dependencies at all.", nil
}

func anthropic(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{
		"model":      "claude-opus-5",
		"max_tokens": 1024,
		"fallbacks":  "default", // re-run on another model if this one declines
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	})
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("content-type", "application/json")
	req.Header.Set("x-api-key", os.Getenv("ANTHROPIC_API_KEY"))
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-beta", "server-side-fallback-2026-07-01")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var out struct {
		StopReason string `json:"stop_reason"`
		Content    []struct{ Type, Text string }
		Error      *struct{ Message string }
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.Error != nil {
		return "", fmt.Errorf("anthropic: %s", out.Error.Message)
	}
	if out.StopReason == "refusal" {
		return "", fmt.Errorf("anthropic: request declined")
	}
	var text strings.Builder
	for _, c := range out.Content {
		if c.Type == "text" {
			text.WriteString(c.Text)
		}
	}
	return strings.TrimSpace(text.String()), nil
}

func build() (*minigraph.App[State], error) {
	return minigraph.New[State]().
		AddNode("draft", func(ctx context.Context, s State) (State, error) {
			var err error
			s.Draft, err = llm(ctx, "Write one sentence introducing "+s.Topic+".")
			return s, err
		}).
		AddNode("tighten", func(ctx context.Context, s State) (State, error) {
			var err error
			s.Final, err = llm(ctx, "Tighten to under 12 words, keep the facts: "+s.Draft)
			return s, err
		}).
		AddEdge(minigraph.Start, "draft").
		AddEdge("draft", "tighten").
		AddEdge("tighten", minigraph.End).
		Compile()
}

func main() {
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		llm = anthropic
		fmt.Println("model: claude-opus-5 (live)")
	} else {
		fmt.Println("model: mock (set ANTHROPIC_API_KEY to go live)")
	}
	app, err := build()
	if err != nil {
		panic(err)
	}
	final, err := app.Invoke(context.Background(), State{Topic: "MiniGraph, a 420-line Go graph engine for agents"})
	if err != nil {
		panic(err)
	}
	fmt.Println("draft:", final.Draft)
	fmt.Println("final:", final.Final)
}
