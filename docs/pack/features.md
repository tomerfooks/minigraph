# MiniGraph feature inventory

Exact as of commit `51ec765` (2026-09-11). Line counts and behaviors below were
taken from the source, not the README. Library: `graph.go` (239 lines),
`checkpoint.go` (93), `interrupt.go` (31), `parallel.go` (57) = 420 lines, zero
dependencies beyond the standard library (`context`, `errors`, `fmt`, `iter`,
`maps`, `slices`, `sync`). Test coverage of the root package: 97.8% of
statements (`go test -cover .`). Module: `github.com/tomerfooks/minigraph`,
`go 1.24`.

## Public symbols

Every exported identifier, what it does, a one-line example, and the guarantee
behind it. "Pinned by" names the test that fails if the guarantee changes.

### Constants and errors

| Symbol | What it does | Example | Guarantee |
|---|---|---|---|
| `Start` (`"__start__"`) | Reserved source of the entry edge. | `AddEdge(minigraph.Start, "agent")` | `AddNode` rejects the name; `Compile` fails without an edge from it. |
| `End` (`"__end__"`) | Reserved terminal target. | `return minigraph.End, nil` from a router | Routing to `End` ends the run without yielding a Step. Valid as a static target or router result. |
| `ErrMaxSteps` | Sentinel for runaway cycles. | `errors.Is(err, minigraph.ErrMaxSteps)` | Wrapped with the limit: `"minigraph: max steps exceeded (25)"`. Pinned by `TestMaxSteps`. |

### Function types

| Symbol | What it does | Example | Guarantee |
|---|---|---|---|
| `Node[S] = func(ctx, S) (S, error)` | Transforms the whole state and returns the next one. | `func(ctx context.Context, s State) (State, error) { s.N++; return s, nil }` | Nodes are plain functions; no interface, no registration beyond `AddNode`. Anything with this signature is a node, including `(*App[S]).Invoke`. |
| `Router[S] = func(ctx, S) (string, error)` | Names the next node (or `End`) after a node ran. | `func(_ context.Context, s State) (string, error) { if s.Done { return minigraph.End, nil }; return "tools", nil }` | Router errors are wrapped as `edge from "<node>": <err>`; an unknown target is an error at run time, not compile time (`TestRouterError`, `TestRouterUnknownTarget`). |

### Builder: `Graph[S]`

| Symbol | What it does | Example | Guarantee |
|---|---|---|---|
| `New[S]() *Graph[S]` | Empty builder over state type `S`. | `g := minigraph.New[State]()` | `S` is any Go type; the compiler checks every node against it. |
| `(*Graph[S]).AddNode(name, fn)` | Registers a node. | `g.AddNode("agent", callModel)` | Rejects empty, `Start`, `End`, nil `fn`, and duplicates. Mistakes are recorded, not returned; chaining continues. |
| `(*Graph[S]).AddEdge(from, to)` | Unconditional edge. | `g.AddEdge("tools", "agent")` | Exactly one outgoing edge per node; a second `AddEdge`/`AddRouter` from the same node is an error. `to` may be `End`. |
| `(*Graph[S]).AddRouter(from, r)` | Conditional edge decided by the state at run time. | `g.AddRouter("agent", decide)` | Nil router is an error. Internally a static edge is just an edge with a nil router; all routing goes through one function (`App.route`). |
| `(*Graph[S]).Compile() (*App[S], error)` | Validates and freezes. | `app, err := g.Compile()` | Reports **all** problems at once via `errors.Join`, in deterministic (sorted) order: accumulated builder errors, missing `Start` edge, nodes with no outgoing edge, edges from unknown nodes, static edges to unknown targets. Maps are cloned, so the builder can keep mutating without affecting the app (`TestCompileErrors`, `TestAppIsolatedFromBuilder`). |

### Compiled app: `App[S]`

| Symbol | What it does | Example | Guarantee |
|---|---|---|---|
| `App[S].MaxSteps` (field) | Cap on node executions per run. | `app.MaxSteps = 100` | Set to 25 by `Compile`. Checked after routing and before executing a node: with the default, 25 executions succeed and the 26th attempt yields `ErrMaxSteps`. Counts from zero on every `StreamFrom`/`InvokeFrom` call, including resumes. |
| `(*App[S]).Invoke(ctx, state) (S, error)` | Runs to completion. | `final, err := app.Invoke(ctx, State{Q: "..."})` | Returns the last successfully produced state on error (`TestNodeError`). `App` is immutable and safe for concurrent `Invoke`s (`TestConcurrentInvokes`). |
| `(*App[S]).Stream(ctx, state) iter.Seq2[Step[S], error]` | Yields after every node. | `for step, err := range app.Stream(ctx, s) { ... }` | One `Step` per executed node, in order (`TestStreamYieldsEveryStep`). An error arrives as the final pair. Breaking out of the loop stops the run (`TestStreamBreak`). No Step is yielded for `End`. |
| `(*App[S]).InvokeFrom(ctx, step) (S, error)` | Resumes from any yielded Step. | `app.InvokeFrom(ctx, minigraph.Step[State]{Node: intr.Node, State: edited})` | Routing restarts *from* `step.Node`; that node is **not** re-executed (`TestResumeFromMidRun`). Unknown node name is an error (`TestResumeUnknownNode`). |
| `(*App[S]).StreamFrom(ctx, step) iter.Seq2[Step[S], error]` | The single engine; everything else wraps it. | `for step, err := range app.StreamFrom(ctx, cp) { ... }` | Checks `ctx.Err()` before every routing decision and yields the current position with the context error (`TestContextCancel`). |
| `(*App[S]).StreamThread(ctx, saver, thread, initial)` | `Stream` with durability. | `for step, err := range app.StreamThread(ctx, saver, "t-1", s)` | If the thread has a saved Step, the run resumes from it and `initial` is ignored. Every successful Step and every Interrupt Step is saved **before** it is yielded; plain node errors are not saved (`TestStreamThreadResumesAfterBreak`, `TestInterruptWithThread`). A `Save` failure ends the run with `save thread "<id>": <err>` (`TestStreamThreadSaveError`). |
| `(*App[S]).InvokeThread(ctx, saver, thread, initial) (S, error)` | Runs a thread to completion or its next Interrupt. | `final, err := app.InvokeThread(ctx, saver, "t-1", s)` | Calling it on a finished thread is a no-op that returns the saved final state (`TestInvokeThread`). To answer an Interrupt: edit the state, `saver.Save(ctx, thread, Step{Node: intr.Node, State: edited})`, call again. |

### Data types

| Symbol | What it does | Example | Guarantee |
|---|---|---|---|
| `Step[S]{Node string; State S}` | One executed step, and a resume point. | `Step[State]{Node: "agent", State: s}` | Both the event type of `Stream` and the argument of `InvokeFrom`, so every event is a checkpoint. The three yield conventions are load-bearing (see below). |
| `Checkpointer[S]` interface | `Save(ctx, thread, Step[S]) error`; `Load(ctx, thread) (Step[S], bool, error)`. | `type PgSaver struct{...}` implementing both | Only the **latest** Step per thread is required. Implementations must serialize `S` themselves (JSON, gob, ...). |
| `MemorySaver[S]` | In-process `Checkpointer`. | `saver := &minigraph.MemorySaver[State]{}` | Zero value is ready; mutex-guarded; stores Steps by value, so reference fields inside `S` still alias shared data (`TestMemorySaver`). |
| `Interrupt{Payload any; Node string}` | Error type that pauses a run instead of failing it. | `return s, &minigraph.Interrupt{Payload: "approve?"}` | Detected with `errors.As`, so wrapping with `%w` still works. `Node` is set by the engine. The state returned alongside it is **kept** and yielded as the final Step. `Error()` renders `minigraph: interrupted at "<node>": <payload>`. Pinned by `TestInterruptAndResume`. |

### Combinator

| Symbol | What it does | Example | Guarantee |
|---|---|---|---|
| `Parallel[S](merge, branches...) Node[S]` | Runs branches concurrently on the same starting state, then folds results with `merge`. | `minigraph.Parallel(mergeFindings, searchWeb, searchDocs, sub.Invoke)` | Results are delivered to `merge` in branch order regardless of completion order (`TestParallelMergesInBranchOrder`). First branch error cancels siblings via a derived context and fails the node with `parallel branch <i>: <err>` (`TestParallelBranchErrorCancelsSiblings`). Branches get shallow copies of `S`. Nil merge is a run-time error (`TestParallelNilMerge`). Whole compiled subgraphs are valid branches (`TestParallelSubgraphBranches`). Counts as one step toward `MaxSteps`. |

## The three yield conventions

These are the contract that makes `Step` a checkpoint. Changing any one of
them breaks resume.

| Situation | Step yielded | Error yielded | Effect of resuming that Step |
|---|---|---|---|
| Node succeeded | `{Node: that node, State: its output}` | `nil` | Continues onward. |
| Node returned a plain error | `{Node: last completed node, State: last good state}` | `node "<name>": <err>` | Re-runs the failed node (`TestErrorStepRetriesFailedNode`). |
| Node returned `*Interrupt` | `{Node: the interrupted node, State: the state it returned}` | fresh `*Interrupt{Payload, Node}` | Routes onward; the node does not re-run (`TestInterruptAndResume`). |
| Router or context error | `{Node: current node, State: current state}` | wrapped error | Re-runs the router (and so re-attempts the transition). |
| `MaxSteps` exceeded | `{Node: current node, State: current state}` | `ErrMaxSteps (N)` | Resuming gets a fresh budget of `MaxSteps`. |

Two details worth knowing when building on this:

- The engine rebuilds the Interrupt it yields (`&Interrupt{Payload, Node}`),
  so any wrapping context added around the node's Interrupt is dropped; the
  payload survives, the wrapper does not.
- `Invoke` on an Interrupt returns the interrupted node's returned state, not
  the pre-node state. This is what lets a human edit "the draft" rather than
  the input.

## Observed discrepancy: Interrupt inside Parallel

`parallel.go`, `interrupt.go` and `AGENTS.md` all say an Interrupt returned
inside a `Parallel` branch "is treated as a plain error". A probe run against
the current code shows otherwise: `Parallel` wraps the branch error with `%w`,
the engine detects Interrupts with `errors.As`, so the run **pauses** at the
`Parallel` node with `Interrupt.Node` set to the Parallel node's name and the
Step's state equal to the pre-branch base state (branch output is discarded).
Resuming routes past the Parallel node, so the branches never re-run. Siblings
are still cancelled. Either the docs or the behavior should be aligned; this
inventory records the behavior as it is. No existing test covers the case.

## Patterns you get for free

Nothing below needs support code; each falls out of the types above.

- **Subgraphs.** `(*App[S]).Invoke` has `Node[S]`'s signature, so
  `g.AddNode("plan", planner.Invoke)` nests a compiled graph as a node.
  Interrupts inside the subgraph surface through the outer node as Interrupts
  (same `errors.As` path); the outer Step's state is whatever the subgraph
  returned.
- **Map-reduce.** `Parallel(reduce, mapA, mapB, mapC)` is fan-out over a
  fixed set of branches with a user-written reduce; the graph stays
  sequential and deterministic. Dynamic fan-out (N branches decided at run
  time) is a loop that builds the `branches...` slice before calling
  `Parallel`.
- **Human-in-the-loop.** A node returns `&Interrupt{Payload: question}`;
  the caller inspects `errors.As(err, &intr)`, edits the state, and calls
  `InvokeFrom` (or `saver.Save` + `InvokeThread`). The router after the
  interrupted node reads the human's answer out of the state. See
  `examples/approval`.
- **Durable threads.** `InvokeThread` with any `Checkpointer` gives
  crash-resume: only the latest Step per thread is stored, and it is stored
  before it is yielded. A process can die at any point and the next call
  continues from the last saved node.
- **Retries by rerun.** Failed nodes are not checkpointed, so rerunning a
  thread (or `InvokeFrom` on the error Step) retries exactly the node that
  failed with the last good state. Backoff is a `for` loop around the call.
- **Streaming progress.** `for step, err := range app.Stream(ctx, s)` is
  the progress feed; the loop body can print, publish to a channel, or write
  to a websocket. Breaking out cancels the run, and the last Step you saw is
  a valid resume point.
- **Cancellation and deadlines.** `context.Context` flows into every node,
  router and branch; the engine checks it before each transition, and
  `Parallel` cancels siblings on the first failure.
- **Fail-fast wiring.** All builder mistakes surface at `Compile` as one
  joined error, so a test that calls `Compile()` is a complete wiring check.
- **Concurrency.** A compiled `App` is immutable; one instance can serve
  many goroutines and many threads at once.

## Things you build outside

Each is one sentence on how, consistent with `AGENTS.md`'s out-of-scope list.

- **LLM calls.** Call any client (OpenAI, Anthropic, Ollama, Bedrock) inside a
  node; keep the model's output in the state so routers can branch on it.
- **Tool calling.** Represent pending calls as a field in `S`; a `tools` node
  executes them and appends results; the router loops back to the model node
  (`examples/agent`, `examples/react`).
- **Per-key reducers.** Write the merge you want in `Parallel`'s `merge`
  argument; for sequential nodes, a node returns the whole next state, so
  "reduce" is just field assignment.
- **Retry policies and backoff.** Wrap a `Node` in a function that loops
  with `time.Sleep` or a library backoff, or rerun the thread from the
  outside.
- **Token streaming.** Stream inside the node (the client's stream API) and
  publish partial tokens to a channel the caller owns; the node still returns
  one final state.
- **Persistent checkpointers.** Implement `Checkpointer[S]` over Postgres,
  Redis, SQLite or S3 by JSON-encoding `Step[S]`; two methods, latest-only.
- **Checkpoint history and time travel.** Save every Step under
  `thread + step index` in your own `Checkpointer`, then `InvokeFrom` any of
  them.
- **Observability.** Log or trace around the `Stream` loop (every Step has
  the node name) and inside nodes with OpenTelemetry or `log/slog`.
- **Timeouts per node.** `context.WithTimeout` inside the node, or a wrapper
  node that applies one.
- **Dynamic graphs.** Build a new `Graph` per request and `Compile` it;
  compile is cheap (two map clones plus validation).
- **Multi-agent.** Each agent is an `App`; compose them as nodes of a parent
  graph, or as `Parallel` branches.
- **Server or platform layer.** MiniGraph is a library; expose threads over
  HTTP/gRPC in your own service, mapping `thread` to your request or user id.

## Compile-time versus run-time checks

Knowing which mistakes surface where matters for testing strategy: a unit test
that calls `Compile()` covers the left column; the right column needs a run.

| Checked at `Compile` (joined, all at once) | Checked at run time (first occurrence) |
|---|---|
| Empty, reserved, or duplicate node names | Router returns an unknown node name |
| Nil node or nil router | Router returns an error |
| Second outgoing edge from the same node | Node returns an error or Interrupt |
| No edge from `Start` | `MaxSteps` exceeded |
| Node with no outgoing edge | Context cancelled or deadline passed |
| Edge from an unregistered node | Resume from an unknown node name |
| Static edge to an unregistered target | `Checkpointer.Load`/`Save` failures |
| | `Parallel` with nil merge; branch failure |

## Minimal end-to-end

The smallest program that exercises build, run, stream, interrupt and resume.
It compiles against the current module and needs no API keys.

```go
type S struct{ N int; OK bool }

app, err := minigraph.New[S]().
    AddNode("inc", func(_ context.Context, s S) (S, error) { s.N++; return s, nil }).
    AddNode("gate", func(_ context.Context, s S) (S, error) {
        if !s.OK { return s, &minigraph.Interrupt{Payload: "continue?"} }
        return s, nil
    }).
    AddEdge(minigraph.Start, "inc").
    AddEdge("inc", "gate").
    AddRouter("gate", func(_ context.Context, s S) (string, error) {
        if s.N >= 3 { return minigraph.End, nil }
        return "inc", nil
    }).
    Compile()

final, err := app.Invoke(ctx, S{})          // pauses at "gate" with N == 1
var intr *minigraph.Interrupt
for errors.As(err, &intr) {                 // answer, resume, repeat
    final.OK = true
    final, err = app.InvokeFrom(ctx, minigraph.Step[S]{Node: intr.Node, State: final})
}
// final.N == 3, err == nil
```

Note the loop: `OK` is set once, and because the gate node is not re-executed
on resume, the only remaining pauses come from the next `inc` -> `gate` pass,
where `OK` is already true and the node returns normally. Human-in-the-loop
protocols are therefore expressed entirely in the state type.

## Non-goals encoded in the types

The type signatures rule these out rather than a policy document:

- A node cannot return a partial update; `Node[S]` returns `S`. There is no
  reducer to configure because there is nothing to reduce.
- A node cannot pick its successor; only `Router[S]` does, and there is one
  per node. Control flow has exactly one place to read.
- The engine cannot store more than one Step per thread; `Checkpointer[S]`
  has no history method. History is the implementer's choice.
- The engine has no model, tool, or prompt types. Provider bindings cannot
  creep in without adding a new file, which the README line count would
  expose.
