# MiniGraph code review

Scope: `graph.go`, `checkpoint.go`, `interrupt.go`, `parallel.go`, all tests,
the four examples, README, AGENTS.md, CI. Commit `51ec765`. Everything below
was checked by reading the code and, where behaviour was in doubt, by running
probe programs against the library from a separate module (no repository
files were modified).

Baseline: `go test -race ./...` passes; `go test -cover` reports 97.8%;
`go vet` and `gofmt -l` are clean. Uncovered: `Interrupt.Error`
(`interrupt.go:29-31`) and the `Load`-error branch of `StreamThread`
(`checkpoint.go:54-56`).

The constraint that frames every suggestion: `AGENTS.md` says small size is
the product and CI fails if the README's "420 lines" claim drifts. Every
suggestion states its line cost against that budget.

---

## Strengths

- **One engine.** `StreamFrom` (`graph.go:161-205`) is 45 lines and every
  public entry point funnels through it. There is nothing to keep in sync.
- **`Step` as event and checkpoint** is the right primitive. Resume is
  "route from `from.Node`", full stop. Interrupts and retries are not
  features; they are consequences of which Step gets yielded
  (`graph.go:189` vs `graph.go:195`).
- **Push iterator, no goroutines.** `iter.Seq2` means `break` is a return, not
  a leaked goroutine or a drained channel (`graph_test.go:260`).
- **Compile reports everything at once**, in sorted order, via `errors.Join`
  (`graph.go:98-122`). Builder methods never panic and never return errors, so
  chains stay readable.
- **App is genuinely immutable** after `maps.Clone` (`graph.go:126-127`) and
  the test suite proves both isolation from the builder and concurrent use.
- **Error Steps carry the last good state.** `Invoke` returning usable state
  on failure (`graph.go:214-223`) is unusual and practical.
- **Documentation is the code.** The doc comments in `interrupt.go` and
  `checkpoint.go:74-80` contain the exact snippet a user needs.
- **Tests are behaviour-named and pin the load-bearing conventions.**
  `TestErrorStepRetriesFailedNode`, `TestInterruptAndResume`,
  `TestResumeFromMidRun`, `TestStreamThreadResumesAfterBreak` each guard one
  sentence of the design.

---

## Findings, ranked by severity

### 1. `Interrupt` inside a `Parallel` branch is not "a plain error" — it corrupts the checkpoint

`parallel.go:20-21`, `interrupt.go:22-23`, and `AGENTS.md` all state that an
`Interrupt` returned by a branch is treated as a plain error. The code does
not do that. `parallel.go:43` wraps with `%w`:

```go
first = fmt.Errorf("parallel branch %d: %w", i, err)
```

and the engine at `graph.go:186` uses `errors.As`, which unwraps. Verified
behaviour with a two-branch `Parallel` node named `fan` whose first branch
interrupts:

- `Invoke` returns an `*Interrupt{Node: "fan"}` with the **pre-fan-out**
  state (Parallel returns its input on error, `parallel.go:53`). The
  interrupting branch's state and the sibling's result are both lost.
- `StreamThread` saves `Step{fan, <input state>}` as a checkpoint
  (`checkpoint.go:61`), because it is an interrupt.
- `InvokeFrom` / a thread rerun routes **onward from `fan`**, so the entire
  fan-out is skipped and the next node runs on the pre-fan-out state.

This is worse than either documented alternative: a plain error would retry
the fan-out; a real interrupt would keep state. A user who puts an approval
gate inside a parallel research branch gets silent skip-on-resume.

**Fix** (`parallel.go`, +4 lines including the `errors` import):

```go
if err != nil {
    if errors.As(err, new(*Interrupt)) {
        err = errors.New(err.Error()) // a branch cannot pause the run
    }
    once.Do(...)
```

Do not "fix" this by changing `%w` to `%v`; that would also hide ordinary
branch errors from `errors.Is`. Add a test (`parallel_test.go`, ~20 lines)
asserting `errors.As` fails on the node's error and that resume re-runs the
node. README would move to 424. **Worth it: this is a correctness bug that
persists a wrong checkpoint.**

### 2. A panic in a `Parallel` branch kills the process — after the merge may have run

A panic in a sequential node unwinds through the iterator into the caller's
`range` loop and is recoverable there; nothing is yielded or saved. A panic
in a `Parallel` goroutine (`parallel.go:38-49`) is on another goroutine and
cannot be recovered by the caller. Two effects, both verified:

- The process terminates. `README.md:163` ("Crash anywhere … run the thread
  again") still holds because the last checkpoint is intact, but a library
  that wraps user callbacks in goroutines usually converts panics to errors.
- Ordering: `defer wg.Done()` runs *during* unwinding, so `wg.Wait()`
  returns, `merge` runs with a **zero-value** `results[i]` for the panicked
  branch, the engine yields (and `StreamThread` may `Save`) that state, and
  only then does the runtime abort. A durable checkpointer can persist a
  Step built from a zero-valued branch result.

**Fix** (`parallel.go`, +6 lines): hoist the `once.Do` body into a `fail`
closure and add `defer func() { if p := recover(); p != nil { fail(fmt.Errorf("parallel branch %d: panic: %v", i, p)) } }()`
before the branch call. **Worth it**, but lower than finding 1 because a
panic is already a programming error and the claim in the README survives.
Alternative at 0 lines: document it in the `Parallel` comment.

### 3. Shallow-copy aliasing reaches yielded Steps and checkpoints, and the project's own example does it

`S` is passed by value, so slices, maps and pointers inside it are shared
between the input state, the returned state, every yielded Step, and every
`MemorySaver` entry (`checkpoint.go:22` says so for `MemorySaver`; nothing
says it for `Step`). `examples/react/main.go:152` does:

```go
turn := &s.Trace[len(s.Trace)-1]
...
turn.Observation = out
```

which mutates the backing array shared with the `reason` Step yielded one
iteration earlier. Verified: after `act` runs, the Step previously yielded
for `reason` reports `Observation` set. In the example it is harmless because
the consumer already printed, but the same pattern on a durable thread
rewrites history in a saved checkpoint, and inside `Parallel` two branches
doing `append` on a slice with spare capacity race (the race detector would
flag it; no test exercises it). The test helper `visit` at
`graph_test.go:16-22` clones defensively, so the author knows; `gateGraph`
at `interrupt_test.go:16` does not.

**Fix**: change the example to copy the last element, mutate, and write it
back (+2 example lines, 0 library lines); add one sentence to the `Step` doc
comment (`graph.go:142-144`) — "reference fields are shared with earlier
Steps; clone before mutating in place" (+1 line). **Worth it.**

### 4. After an interrupt, the gate must live in the router, and a crash-rerun proceeds as if answered

Because resume routes *onward* from the interrupted node (`graph.go:158-160`),
a node that pauses cannot re-check anything on resume. `examples/approval`
gets this right (`main.go:35-53`): `approve` only pauses; the router reads
`s.Approved`. But the natural first attempt — `approve` node checks
`s.Approved` and interrupts if false, followed by `AddEdge("approve","send")`
— compiles, passes the first run, and on resume **sends without approval**.

Related: `StreamThread` saves the interrupt Step (`checkpoint.go:61`), and
the `Interrupt` (including `Payload`) is not part of the Step. After a process
restart, rerunning the thread *without* editing the state routes onward
exactly as if the human had answered, and the question the node asked is
gone unless the node also wrote it into `S`. Both `gateGraph` and the
approval example are safe only because their routers re-ask when the answer
is missing.

**Fix**: two doc lines in `interrupt.go` ("put the decision in the router
after the node; write the payload into the state if it must survive a
restart") (+2 lines). No code change is appropriate — the alternative
(re-running the interrupted node on resume) would break the retry
convention. **Worth it.**

### 5. Durable threads are at-least-once, and `MaxSteps` resets on every rerun

- A crash between a node returning and `Save` completing, or any `Save`
  error (`checkpoint.go:62-65`), re-runs that node on the next
  `InvokeThread`. Side-effecting nodes need idempotency keys in `S`. Not
  stated anywhere.
- `MaxSteps` counts from zero per `StreamFrom` call (`graph.go:160`,
  `graph.go:168`). Verified: a `loop → loop` graph with `MaxSteps = 3` on a
  thread reaches `N = 3, 6, 9` across three `InvokeThread` calls, each
  returning `ErrMaxSteps`. A retry loop around `InvokeThread` therefore
  never terminates on a runaway graph. Documented in `graph.go:160` but not
  in `checkpoint.go`, where the consequence is.

**Fix**: one sentence each in the `StreamThread` comment (+2 lines). **Worth it.**

### 6. Smaller edge cases (verified)

| Case | Behaviour | Assessment |
|---|---|---|
| Router returns `Start` or `""` | `edge from "a": routed to unknown node "__start__"` (`graph.go:235-236`) | Correct; message could say "reserved" but not worth a line |
| `AddEdge("a", Start)` | Compile error `unknown target` (`graph.go:116-119`) | Correct |
| `AddRouter(Start, r)` | Compiles and works: conditional entry point | Undocumented feature; one README bullet would be free |
| `AddEdge(Start, End)` | Compiles; `Invoke` returns input with no steps | Fine |
| Context pre-cancelled | Yields `Step{Start, initial}` with `context.Canceled` | Fine; note `Step.Node == Start` is possible in error Steps |
| Context cancelled after last node, before routing to `End` | Reports `Canceled` although the run is complete (`graph.go:169` precedes `graph.go:178`) | Cosmetic; `InvokeFrom` still returns the final state |
| Wrapped interrupt `fmt.Errorf("x: %w", &Interrupt{Node:"user"})` | Detected; engine yields a **fresh** `*Interrupt`, dropping wrapper text and user-set `Node` (`graph.go:189`) | Reasonable; document "Node is overwritten" is already there |
| `nil` ctx | Nil-pointer panic at `graph.go:169` | Standard Go; do not guard |
| Resume from a node removed after a graph change | `resume from unknown node` yielded with the original Step (`graph.go:164-167`); not saved; thread is stuck until someone `Save`s under a valid node | Acceptable; note `Checkpointer` has no `Delete`, so "reset thread" means `Save(Step{Start, s})` |
| Resume from a node whose edge changed | Routes with the new edge | By design ("every Step is a checkpoint") |
| `Parallel` with a branch that ignores `ctx` | First error is recorded but the node blocks in `wg.Wait()` until the slow branch returns (`parallel.go:51`) | Doc says "cancels the siblings"; "signals" is accurate. 0-line wording fix |
| `Parallel` with zero branches | `merge(ctx, state, [])` | Fine |
| `Parallel` second and later errors | Dropped; only the first is reported | Fine; `errors.Join` would cost lines for little |
| `App.MaxSteps` mutated while runs are in flight | Data race (read at `graph.go:181`) | "Set once after Compile" is implied; not worth a setter |
| `MemorySaver` on a long-running server | Threads are never deleted; grows without bound | Interface has no `Delete`; see suggestions |
| JSON of `Step[S]` | Exported fields, no tags: `{"Node":…, "State":…}`; works iff `S` does | Fine; `Interrupt.Payload any` is not in the Step so it never hits the serializer |
| `InvokeThread` | Calls `Load` twice (`checkpoint.go:83`, then `checkpoint.go:54`) | One extra read per call; not worth restructuring |

---

## Test coverage assessment

Pinned well: linear and cyclic runs, `MaxSteps` boundary (exactly N
executions), node and router errors with last-good state, all nine compile
error classes, `break`, cancellation mid-run, retry convention, interrupt
convention, resume-from-Step convention, builder isolation, concurrent
`Invoke`, finished-thread no-op, resume after break, `Save` error, branch
order, sibling cancel, subgraph branches.

Not pinned (each verified by hand for this review; none has a test):

1. Interrupt inside `Parallel` (finding 1). Highest priority; the current
   doc claim is false and nothing would catch a fix regressing.
2. Panic in a `Parallel` branch (finding 2).
3. Two `Parallel` branches appending to a shared slice under `-race`.
4. `StreamThread` when `Load` fails (`checkpoint.go:54-56`) — the only
   uncovered library branch besides `Interrupt.Error`.
5. `MaxSteps` resets across `InvokeThread` reruns.
6. Rerunning an interrupted thread *without* editing state routes onward.
7. Router returning `Start` / `""`.
8. `AddRouter(Start, …)` conditional entry.
9. Pre-cancelled context yields `Step{Start, …}`.
10. Wrapped `Interrupt` is detected and `Node` is overwritten.
11. Examples are compiled by CI but never executed; the README's ReAct
    transcript is unverified. A `go run ./examples/react | diff` step in CI
    would be ~5 YAML lines.

Also: `AGENTS.md` says "Tests are table-free", but `TestCompileErrors`
(`graph_test.go:175-258`) is a table test. Either is fine; the doc is wrong.

---

## API ergonomics

- `New[State]()` needs an explicit type argument (no value to infer from);
  `Parallel(merge, …)` infers from `merge`; `Parallel[S](nil)` only in the
  degenerate case. Normal for Go generics, no complaint.
- `App.Invoke` being a `Node[S]` by method-value is the best trick in the
  library; it should be the first thing the README says about subgraphs (it
  is nearly that already).
- `Step[S]{Node: intr.Node, State: s}` is the resume incantation and appears
  in three docs and two examples. A `func Resume[S any](intr *Interrupt, s S) Step[S]`
  would be +3 lines and save nothing important. Not recommended.
- `StreamThread`'s `initial` being **ignored** when a checkpoint exists
  (`checkpoint.go:45-46`) is documented but surprising for chat-style
  multi-turn use, where each call brings new input. The supported path is
  the interrupt path: `Load`, edit, `Save`, `InvokeThread`. Worth one README
  sentence; a helper would cost ~10 lines and is not recommended.
- Router targets are strings validated at run time (`graph.go:235`). Typos
  surface as `routed to unknown node "tols"` on the first run that takes that
  branch, not at `Compile`. An optional `targets ...string` on `AddRouter`
  validated in `Compile` would be ~8 lines. Marginal; the run-time message is
  clear. Not recommended unless users ask.
- A forgotten `Compile` error check yields a nil `*App` and a nil-pointer
  panic on first use. Standard Go; no change.

---

## Suggestions, ranked by value against line cost

| # | Change | Library lines | Verdict |
|---|---|---|---|
| 1 | Strip `*Interrupt` from branch errors in `Parallel` (finding 1) + test | +4 (README → 424) | **Do it.** Correctness bug, wrong checkpoint persisted. |
| 2 | Fix `examples/react` aliasing; one sentence on `Step` about shared reference fields (finding 3) | +1 | **Do it.** |
| 3 | Two sentences in `interrupt.go`: gate goes in the router; payload is not persisted (finding 4) | +2 | **Do it.** Cheapest way to prevent the "sends without approval" mistake. |
| 4 | Two sentences in `checkpoint.go`: at-least-once; `MaxSteps` resets per rerun (finding 5) | +2 | **Do it.** |
| 5 | Add the ten missing tests listed above | 0 | **Do it.** Test code is explicitly unbounded. |
| 6 | Recover panics in `Parallel` goroutines into a branch error (finding 2) | +6 | Probably worth it; converts process death into the existing error path and closes the zero-value merge window. If not, document (+1). |
| 7 | Reword "cancels the siblings" to "signals the siblings and waits for them" (`parallel.go:19-20`) | 0 | Do it. |
| 8 | Fix `AGENTS.md` "table-free" claim; note `AddRouter(Start, …)` works | 0 | Do it. |
| 9 | CI: run `./examples/react` and diff against the README transcript | 0 (CI only) | Cheap honesty check, same spirit as the line-count check. |
| 10 | Add `Delete` to `Checkpointer` + `MemorySaver` | +8 and a breaking interface change | **Not worth it.** Users own their backend; `MemorySaver` is for tests. Document that finished threads are never removed. |
| 11 | `MaxSteps` as a `Compile` option instead of an exported field | +6, API change | Not worth it. |
| 12 | Declared router targets validated at `Compile` | +8 | Not now. |
| 13 | `Resume(intr, s)` helper | +3 | No. |
| 14 | Recover panics in sequential nodes | +4 | No; leave to the caller, as Go does. |
| 15 | Remove the double `Load` in `InvokeThread` | ~0 net, harder to read | No. |

Net if 1–4 and 6 are all taken: 420 → 435 lines, README claims move once.
If only 1–4: 420 → 429.

---

## Summary

The engine is sound and the three yield conventions do exactly what the
documentation says they do. The one real bug is at the seam between
`Parallel` and the interrupt detector: the `%w` at `parallel.go:43` lets an
interrupt escape a branch, and the resulting checkpoint skips the fan-out on
resume. It costs four lines to fix. Everything else of substance is
documentation: aliasing through shared reference fields, the router-as-gate
rule after an interrupt, at-least-once execution on threads, and `MaxSteps`
being per-call. The test suite is strong on the sequential engine and thin
on `Parallel`'s failure modes.
