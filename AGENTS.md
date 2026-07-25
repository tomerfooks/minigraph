# AGENTS.md

Guidance for AI coding agents working in this repository.

## What this is

MiniGraph: a single Go package (`minigraph`, at the repo root) implementing a
LangGraph-shaped graph engine in ~420 lines with **zero dependencies**. Small
size is the product, not an accident — the README advertises the line count and
"you can read all of it". Weigh every addition against that.

## Commands

```sh
go test -race ./...                      # full suite
go test -race -run TestInterruptAndResume ./...   # single test
gofmt -l .                               # must print nothing; CI fails otherwise
go vet ./...
go build ./...                           # also compiles the examples
go run ./examples/react                  # agent | react | approval | fanout (no API keys)
```

CI (`.github/workflows/ci.yml`) runs exactly: gofmt check, vet, `go test -race`,
`go build ./...`, and a README line-count honesty check, on Go 1.24.

## Architecture

Four library files, one package:

- `graph.go` — `Graph[S]` builder, `Compile`, immutable `App[S]`, and the
  execution engine.
- `checkpoint.go` — `Checkpointer[S]`, `MemorySaver[S]`, durable-thread wrappers.
- `interrupt.go` — `Interrupt`, the pause-for-a-human error type.
- `parallel.go` — `Parallel`, fan-out/join packaged as a single `Node`.

**`StreamFrom` is the only engine.** `Stream`, `Invoke`, `InvokeFrom`,
`StreamThread`, and `InvokeThread` all funnel into `graph.go:161`. Behavior
changes belong there, not duplicated in the wrappers.

**`Step[S]{Node, State}` is both an emitted event and a resume point.** Durability,
interrupts, and retries all fall out of that one fact. Resume semantics: routing
restarts *from* `from.Node`, so the checkpointed node is **not** re-executed.

**The three yield conventions are load-bearing** — change one and resume breaks:

| Situation | Step yielded | Effect on resume |
|---|---|---|
| Node succeeded | that node + its new state | continues onward |
| Node returned a plain error | **last completed** node + last good state | retries the failed node |
| Node returned `*Interrupt` | the **interrupted** node + the state it returned | routes onward; node does not re-run |

Pinned by `TestErrorStepRetriesFailedNode`, `TestInterruptAndResume`,
`TestResumeFromMidRun`.

**Other invariants:**

- Every node has exactly one outgoing edge; a static edge is just an edge with a
  nil `router`. All routing goes through `App.route` (`graph.go:226`).
- Builder mistakes accumulate silently in `Graph.errs` and surface only at
  `Compile`, joined via `errors.Join` — one report, all problems.
- `Compile` clones the maps, so `App` is immutable and safe for concurrent use
  (`TestAppIsolatedFromBuilder`, `TestConcurrentInvokes`).
- `MaxSteps` (default 25) counts from zero on each `StreamFrom` call.
- `StreamThread` saves successful steps and interrupt steps; plain errors are
  **not** saved, so rerunning a thread retries from the last good step.
- Interrupts are detected with `errors.As(err, &intr)` on `*Interrupt`. Inside a
  `Parallel` branch an interrupt is treated as a plain error — cannot pause.
- `Parallel` gives branches shallow state copies (reference fields are shared;
  treat as read-only), keeps results in branch order, and cancels siblings on the
  first error.
- Subgraphs need no support code: `App.Invoke` already has `Node`'s signature.

## Conventions

- No dependencies. `go.mod` stays at module + `go 1.24` (needs `iter`, `maps`,
  `slices`). Never add a dep for something a few lines can do.
- Doc comments are user-facing documentation — the package is meant to be read
  end to end. Match the existing explanatory style, including the runnable
  snippets in `interrupt.go` and `parallel.go`.
- Tests are table-free, behavior-named, and there is more test code than library
  code. New behavior needs a test in the matching `*_test.go`.
- If library line count changes, the README's line-count claims ("420 lines",
  "420 loc", "1.2%") must move with it — CI enforces the exact number.

## Deliberately out of scope

Per-key reducers, retry policies, token streaming, LLM/provider bindings, a
platform layer. A node is `func(ctx, S) (S, error)`; anything else composes from
outside. Proposals to add these should be pushed back on, not implemented
silently.
