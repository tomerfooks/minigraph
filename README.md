# tomergraph

A minimal LangGraph for Go. Zero dependencies, a few hundred lines, the whole
runtime: typed state, conditional edges, cycles, streaming, checkpointing,
human-in-the-loop interrupts, and parallel fan-out.

Build a graph of **nodes** that transform a shared, typed **state**, wire them
with **edges** (static or conditional), **compile**, then **invoke** or
**stream**. Cycles are first-class — that's what makes it an agent runtime
rather than a pipeline.

```go
type State struct {
    Question string
    Answer   string
}

g := tomergraph.New[State]()
g.AddNode("agent", agentFn)
g.AddNode("tools", toolsFn)

g.AddEdge(tomergraph.Start, "agent")
g.AddRouter("agent", func(ctx context.Context, s State) (string, error) {
    if s.Answer == "" {
        return "tools", nil // loop until the agent has an answer
    }
    return tomergraph.End, nil
})
g.AddEdge("tools", "agent")

app, err := g.Compile()
final, err := app.Invoke(ctx, State{Question: "..."})
```

Or watch every step as it happens — `Stream` returns a standard Go iterator:

```go
for step, err := range app.Stream(ctx, initial) {
    if err != nil { ... }
    fmt.Println(step.Node, step.State)
}
```

Run the examples:

- `go run ./examples/agent` — minimal agent-and-tools loop
- `go run ./examples/react` — full ReAct loop: Thought → Action → Observation
  with a trace, tool registry, and completion parsing; swap the mock LLM for a
  real one and the rest stands
- `go run ./examples/approval` — human-in-the-loop: interrupt, reject with
  feedback, redraft, approve, all on a durable thread
- `go run ./examples/fanout` — three researchers in parallel as one graph step

## Checkpoints and durable threads

Every `Step` a run yields is a checkpoint: `InvokeFrom(ctx, step)` continues the
run from exactly that point. The `Checkpointer` interface (with an in-memory
`MemorySaver` included) persists the latest step per thread, and the
`*Thread` methods make runs durable — die at any point, run again, continue
where it stopped:

```go
saver := &tomergraph.MemorySaver[State]{}
final, err := app.InvokeThread(ctx, saver, "thread-42", initial)
```

## Human-in-the-loop

A node pauses a run by returning an `Interrupt`; unlike an error, the state
it returns is kept. The outside world answers by editing the state and
resuming from the interrupted node:

```go
final, err := app.Invoke(ctx, initial)
var intr *tomergraph.Interrupt
if errors.As(err, &intr) {
    fmt.Println("agent asks:", intr.Payload)
    final.Approved = true // the human's answer, written into the state
    final, err = app.InvokeFrom(ctx, tomergraph.Step[State]{Node: intr.Node, State: final})
}
```

## Parallel fan-out

`Parallel` composes nodes into one node: branches run concurrently on copies
of the state, a merge you write folds the results, and the graph stays
sequential. Since `App.Invoke` has a node's signature, whole compiled
subgraphs can be branches:

```go
research := tomergraph.Parallel(
    func(ctx context.Context, base State, results []State) (State, error) {
        for _, r := range results {
            base.Findings = append(base.Findings, r.Finding)
        }
        return base, nil
    },
    searchWeb, searchDocs, subgraph.Invoke,
)
g.AddNode("research", research)
```

## Concepts

| LangGraph | tomergraph |
|---|---|
| `StateGraph(State)` | `New[State]()` — state is any Go type, checked at compile time |
| node | `func(ctx, S) (S, error)` — takes state, returns next state |
| `add_edge` | `AddEdge(from, to)` |
| `add_conditional_edges` | `AddRouter(from, router)` — router returns the next node's name |
| `Start` / `End` | `tomergraph.Start` / `tomergraph.End` |
| `compile()` | `Compile()` — validates the whole graph, returns an immutable `App` |
| `invoke` | `Invoke(ctx, state)` |
| `stream` | `Stream(ctx, state)` — an `iter.Seq2[Step[S], error]`, so it's just `for ... range` |
| recursion limit | `App.MaxSteps` (default 25) |
| checkpointer + `thread_id` | `Checkpointer` / `MemorySaver`, `InvokeThread` / `StreamThread` |
| `interrupt()` | return `&Interrupt{Payload: ...}` from a node, `InvokeFrom` to continue |
| `Send` / parallel supersteps | `Parallel(merge, branches...)` — fan-out/join as a node combinator |
| subgraphs | free: `App.Invoke` is a valid `Node`, so `g.AddNode("sub", sub.Invoke)` |

## Design choices

- **Typed state via generics.** No `map[string]any`, no reducers, no channel
  merging. A node returns the whole next state; branching logic reads it
  directly. The Go compiler is the schema validator.
- **Exactly one outgoing edge per node.** Static edges and routers are the
  same thing internally — a static edge is just a router that ignores the
  state. Branching is always an explicit `Router` decision, so control flow
  has one place to look.
- **Fail at `Compile`, not mid-run.** Builder mistakes (duplicate nodes,
  unknown edge targets, missing entry edge, dead-end nodes) accumulate and
  come back joined from `Compile()`, all at once. The only unavoidable
  runtime check is a router returning an unknown name.
- **`App` is immutable and concurrency-safe.** Compile snapshots the
  graph; keep mutating the builder, run the compiled app from many
  goroutines.
- **Errors keep the last good state.** `Invoke` returns the state produced by
  the last successful node alongside the error, node errors are wrapped with
  the node's name, and an error Step points at the last *completed* node — so
  resuming one retries the step that failed.
- **Every Step is a checkpoint.** There is no separate checkpoint type or
  replay machinery: a run is just a `(node, state)` pair advancing, so any
  yielded pair is a resume point. Interrupts, durability, and retries all
  fall out of that one fact.
- **Fork/join instead of reducers.** Parallelism lives inside `Parallel`,
  where you write the merge, rather than in per-key channel reducers. The
  graph stays deterministic and the state stays a plain typed value. The
  cost: inside a branch, treat shared reference fields (slices, maps) as
  read-only and let the merge reconcile.

Deliberately not included: per-key state reducers, multiple streaming modes,
token streaming, and the platform layer (Studio, tracing, deployment). Nodes
are plain functions — LLM calls, retries, and telemetry compose from the
outside.

## Test

```sh
go test -race ./...
```
