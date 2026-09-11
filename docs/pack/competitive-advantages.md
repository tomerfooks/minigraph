# MiniGraph: competitive advantages

Companion to [alternatives.md](alternatives.md). Every claim below is tied to
a measurement, a source, or a concrete argument; where the evidence is thin
it says so. Research date: 2026-09-11.

---

## 1. Readable in one sitting

**Claim.** The whole engine fits in one reading session, so an engineer can
own it the way they own their own code.

**Evidence.**
- Library size is 420 lines across four files (`graph.go` 239,
  `checkpoint.go` 93, `interrupt.go` 31, `parallel.go` 57), with 723 lines of
  tests. CI fails if the README's line-count claim drifts.
- Measured on shallow clones the same day: LangGraph core + checkpoint base is
  33,849 Python LOC (77,057 across all `libs/`); LangGraph.js core + checkpoint
  is ~34k TS LOC; Eino 59k; ADK Go 67k; Genkit Go 76k; Temporal Go SDK 82k;
  smallnest/langgraphgo (the largest Go clone) 34k. See
  [alternatives.md](alternatives.md) for method.
- Ratio: 420 / 33,849 = 1.24%, the README's "1.2%".

**Why it matters.** Frameworks are debugged by reading their source when the
abstraction leaks. LangGraph's own docs devote pages to reducer semantics,
superstep ordering, and durability modes; community write-ups describe reducers
as "a concurrency policy ... that reviewers, diagrams, and type checkers all
miss" ([Ranjan Kumar](https://ranjankumar.in/langgraph-reducers-concurrent-state-writes)).
MiniGraph has one loop (`StreamFrom`, `graph.go:161`) and one routing function
(`route`, `graph.go:226`); every behavior, including resume and interrupt
semantics, is visible in those 80 lines.

## 2. Zero dependencies and a small supply-chain surface

**Claim.** `go.mod` is `module` + `go 1.24`. Nothing to audit, pin, or patch.

**Evidence.**
- Measured `go.mod` requirements: MiniGraph 0/0; Eino 11 direct / 38 total;
  ADK Go 35/97; Genkit Go 40/130; langchaingo 64/286; Temporal SDK 15/25 plus
  a server. A clean `uv pip install langgraph` pulled 38 packages (29 MB).
- LangGraph's checkpoint layer produced three CVEs in a year: SQL injection in
  `langgraph-checkpoint-sqlite` (CVE-2025-67644), unsafe msgpack
  deserialization of checkpoints in `langgraph <1.0.10` (CVE-2026-28277), and
  RediSearch query injection in the Redis checkpointer (CVE-2026-27022); the
  first two chain to RCE on self-hosted deployments
  ([The Hacker News, Jun 2026](https://thehackernews.com/2026/06/langgraph-flaw-chain-exposes-self.html)).
  MiniGraph ships no serializer and no storage driver; `Checkpointer[S]` is an
  interface the user implements with whatever they already trust.
- Go's module system verifies every download against `go.sum` and the public
  checksum database, a stronger default than PyPI/npm
  ([Go module proxy FAQ](https://www.gofaq.org/en/how-the-go-module-proxy-works-goproxy/)),
  though a 2026 survey found one in six Go pipelines disables that check
  ([Safeguard](https://safeguard.sh/resources/blog/go-module-checksum-database-bypass-risks)).
  With zero dependencies the question is moot: `go.sum` is empty.

**Honest limit.** The user's node code still imports an LLM client. The
advantage is that the orchestration layer adds nothing on top of it.

## 3. Compile-time typed state instead of runtime schemas and reducers

**Claim.** `S` is any Go type; nodes are `func(ctx, S) (S, error)`. The
compiler checks every read and write; there is no schema object, no reducer
registry, and no key-level merge policy to get wrong.

**Argument.**
- In LangGraph the state is a `TypedDict`/Pydantic schema validated at
  runtime; each key optionally has a reducer, and parallel branches writing the
  same key with the default reducer silently overwrite each other
  ([itsourcecode](https://itsourcecode.com/runtimeerror/langgraph-state-reducer-conflicts-fix/)).
  LangGraph.js uses zod, which also validates at runtime.
- Eino is typed, but per node: `NewGraph[I, O]` plus a "type alignment" rule
  where each node's output type must equal the next node's input type; shared
  state is a separate opt-in mechanism with pre/post handlers
  ([Eino docs](https://www.cloudwego.io/docs/eino/core_modules/chain_and_graph_orchestration/chain_graph_introduction/)).
- In MiniGraph a node returns the whole next state. Merging exists in exactly
  one place, the `merge` function the user passes to `Parallel`, so the
  concurrency policy is explicit code with a signature, reviewable in a diff.
- Renaming a field is a compile error at every node, not a `KeyError` at
  runtime; `go vet` and IDE refactors work on the state as ordinary code.

## 4. One static binary: memory and cold start

**Claim.** A MiniGraph agent deploys as a single static executable with a
small footprint; a Python LangGraph service deploys as an interpreter plus a
package tree.

**Evidence.**
- Measured locally: `examples/react` builds to a **1.86 MB** static binary
  (`CGO_ENABLED=0 -ldflags="-s -w"`, Go 1.26) and runs with a **2.5 MB peak
  RSS**. The `langgraph` install alone is 29 MB of site-packages before the
  Python runtime or any provider SDK.
- Cold starts on AWS Lambda (independent of this library): Go ~35–50 ms versus
  Python ~100–300 ms for simple functions
  ([HunCoding, updated Sep 2026](https://huncoding.com/lambda-golang-coldstart-en/));
  a June 2026 workload benchmark reports Go clustered at 62–69 ms regardless
  of memory size, Python 88–91 ms bare and 472–495 ms with NumPy loaded
  ([Harding, Medium](https://jack-harding.medium.com/go-vs-python-on-aws-lambda-a-geospatial-benchmark-b14a02502884)).
- Practitioners report LangGraph services accumulating RAM when
  `MemorySaver` history grows per thread
  ([Aadhil Imam](https://aadhil96.github.io/posts/langgraph-memory-leakage/));
  MiniGraph's `MemorySaver` stores one `Step` per thread by design, so memory
  is bounded by thread count times state size.

**Honest limit.** No end-to-end benchmark of "LangGraph service vs MiniGraph
service" exists yet; the numbers above are Go-vs-Python platform figures plus
local measurements of the example binary. Producing a like-for-like benchmark
would strengthen this section considerably.

## 5. Real parallelism with goroutines

**Claim.** `Parallel` runs branches on goroutines, cancels siblings on first
error via `context.WithCancel`, and returns results in branch order. That is
OS-level parallelism for I/O-bound LLM and tool calls, in 57 lines.

**Argument.** LangGraph's supersteps run concurrent nodes under Python
threading/asyncio; the merge is governed by reducers. MiniGraph's fan-out is a
combinator, so the graph stays sequential and deterministic (one step, one
yield), while the merge is a plain function the user tests in isolation.
Because `App.Invoke` has a `Node`'s signature, an entire compiled subgraph is
a valid branch, giving nested parallelism with no engine support
(`README.md`, "Parallel fan-out"; `TestParallel*` in `parallel_test.go`).

## 6. Step-as-checkpoint

**Claim.** `Step[S]{Node, State}` is simultaneously the streamed event, the
checkpoint, and the resume argument. Durability, interrupts, and retries are
consequences of that one type rather than three subsystems.

**Argument.**
- `Stream` yields a `Step` after every node; `InvokeFrom(step)` resumes from
  it; `StreamThread` saves each one. There is no separate checkpoint schema,
  channel versioning, or history format, which is where LangGraph's
  serializer CVEs lived.
- The three yield conventions (success, plain error, interrupt) are documented
  in `AGENTS.md` and pinned by `TestErrorStepRetriesFailedNode`,
  `TestInterruptAndResume`, and `TestResumeFromMidRun`.
- Eino's equivalent requires `CheckPointStore`, `WithCheckPointID`,
  hierarchical `AddressSegment`s, and `CompositeInterrupt` wrapping for nested
  components ([Eino HITL guide](https://www.cloudwego.io/docs/eino/core_modules/eino_adk/agent_hitl/)).
  ADK Go 2.0 routes HITL through session persistence and `RequestInputEvent`.
  Both are more capable (see trade-offs) and much larger.

## 7. No platform lock-in

**Claim.** There is no MiniGraph platform to buy, and no feature is gated on
one.

**Evidence.** LangGraph's durable deployment story is LangGraph Platform /
LangSmith Deployment: $39/seat/month on Plus with one free small serverless
deployment, enterprise self-hosting by contract
([langchain.com/pricing](https://www.langchain.com/pricing)). Temporal is
$100/month minimum on Cloud or a self-hosted cluster with Cassandra/MySQL/
Postgres ([Temporal pricing](https://docs.temporal.io/cloud/pricing)).
Genkit and ADK are strongest on Google Cloud. MiniGraph's persistence
interface is two methods (`Save`, `Load`); a Postgres implementation is a
few dozen lines against the driver the service already uses.

## 8. Testability: nodes are plain functions

**Claim.** Nodes, routers, and merges are ordinary Go functions over a plain
struct; graphs are values. Unit tests need no harness, mocks of the engine, or
fixtures.

**Evidence.** The repo has more test code than library code (723 vs 420
lines), and the tests are table-free, behavior-named functions
(`TestConcurrentInvokes`, `TestAppIsolatedFromBuilder`, and so on). A test
for a router is a call with a state literal. A test for a full run is
`app.Invoke(ctx, State{...})` with a mock model closure, as `examples/react`
shows with a swappable offline LLM. Compare LangGraph, where testing a node
in isolation still means constructing state dicts that satisfy the schema and
reasoning about which reducer will apply on merge.

## 9. Context cancellation is the cancellation model

**Claim.** `context.Context` flows into every node, router, and branch, and
the engine checks `ctx.Err()` before each routing decision
(`graph.go:169`, added in commit `51ec765`). Cancel the context and the run
stops with the last good `Step`, resumable later.

**Argument.** This is the same cancellation idiom as `net/http`, database
drivers, and every Go LLM client, so deadlines and request-scoped
cancellation propagate without adapters. `Parallel` derives a child context
and cancels siblings on the first error. Python frameworks must bridge
asyncio cancellation, thread pools, and their own stop signals; LangGraph
exposes this through its own config and streaming modes rather than the
language's primitive.

---

## Advantages that are really trade-offs

Be candid about these in any pitch; competitors will raise them.

| Framed as an advantage | The cost |
|---|---|
| 420 lines, nothing you didn't ask for | No token streaming, no retry/backoff, no timeouts, no tracing hooks, no visualizer, no prebuilt ReAct agent. Every team re-implements a few of these outside the engine. |
| Whole-state return, no reducers | The user writes the merge for every `Parallel`; there is no `add_messages`-style helper, and forgetting to copy a slice inside a branch is a data race the engine cannot catch. |
| Step-as-checkpoint, latest step only | No history, no time-travel, no forking a thread from step N. LangGraph and Temporal keep full histories. |
| Checkpointer is a two-method interface | No shipped SQLite/Postgres/Redis savers; serialization of `S` is the user's job. |
| Zero dependencies | Also zero integrations: no provider clients, no MCP, no A2A, no vector stores. Composition from outside is the design, but the catalog is empty. |
| Interrupt is just an error type | Cannot pause inside a `Parallel` branch (downgraded to a plain error); no addressing of nested interrupts as in Eino. |
| Durable threads | Not durable execution. No crash detection, no work queue, no exactly-once; a node that partly ran before a crash re-runs on resume (the same caveat LangGraph documents and Diagrid criticizes). Temporal solves this class; MiniGraph does not. |
| One author, seven commits, readable in an hour | Also: no track record at scale, no company behind it, no security disclosure history. Eino runs inside ByteDance; ADK and Genkit are Google-maintained. |
| `MaxSteps` default 25 | Safer default than LangGraph's 1000, but long tool loops fail until raised. |

---

## Positioning statements

### Candidate one-liners

1. LangGraph for Go, 1.2% of the code. *(current; measured 420 / 33,849)*
2. The agent graph engine you can read before lunch and own by dinner.
3. Typed state, cyclic graphs, checkpoints, interrupts, fan-out. 420 lines.
   Zero dependencies.
4. Every step is a checkpoint. That one idea replaces three subsystems.
5. Nodes are plain Go functions. Everything else is a dependency you didn't
   need.

### 30-second pitch

MiniGraph is a LangGraph-shaped graph engine for Go: nodes over a typed state,
static or conditional edges, cycles, invoke and stream, checkpointers,
human-in-the-loop interrupts, and parallel fan-out. It is 420 lines with the
comments in, has zero dependencies, and ships as part of your static binary.
You can read all of it in one sitting, which means when an agent misbehaves
at 3 a.m. you are debugging your code, not a framework.

### 2-minute pitch

Teams building agents in Go have three choices today. Hand-roll a `for` loop
around the model and reinvent resume, interrupts, and step streaming per
service. Adopt a large framework, ByteDance's Eino at 59k lines and 38
modules, Google's ADK at 67k lines, or Genkit at 76k, and take on their
runtime, their concepts, and their release cadence. Or port LangGraph, which
several projects have tried; the credible ones land at 17k to 34k lines and
none combines typed generics, checkpoints, interrupts, and fan-out under a
thousand.

MiniGraph is the fourth option. It keeps the ideas that carry their weight:
typed state checked by the compiler rather than a runtime schema, one
outgoing edge per node so control flow has one place to look, and a `Step`
that is both the streamed event and the checkpoint, so durability, interrupts,
and retries fall out of one type instead of three subsystems. It drops
per-key reducers, token streaming, retry policies, LLM bindings, and the
platform layer, because nodes are plain functions and all of that composes
from outside. Zero dependencies means an empty `go.sum` and none of the
serializer CVEs that hit LangGraph's checkpoint layer this year. A compiled
example is under 2 MB and runs in under 3 MB of RSS.

It is not durable execution; if you need multi-day workflows that survive
infrastructure loss, run MiniGraph nodes inside Temporal. It is the smallest
engine that makes a Go agent resumable, interruptible, and streamable.

### 5-minute pitch

**The problem.** Agent frameworks have converged on the same core: typed
state, a graph or loop, a checkpoint per step, a way to pause for a human,
and a parallel step. Every vendor then wraps that core in a provider catalog,
tracing, a dev UI, and a hosted platform. In Python that is LangGraph at
34k lines of core plus a $39/seat platform; in Go it is Eino, ADK, and Genkit,
each 60k to 80k lines with 38 to 130 modules. Teams that only want the core
pay for the wrapper in learning curve, dependency surface, and debugging
time.

**What MiniGraph is.** The core, in Go, and nothing else. `New[S]()`,
`AddNode`, `AddEdge`, `AddRouter`, `Compile`. `Invoke`, `Stream` as an
`iter.Seq2`, `InvokeFrom` and `StreamFrom` to resume any yielded `Step`,
`InvokeThread` and `StreamThread` for durable runs against a two-method
`Checkpointer`. `Parallel(merge, branches...)` for goroutine fan-out,
`&Interrupt{}` to pause for a human. A compiled `App.Invoke` has a `Node`'s
signature, so subgraphs are free. The README's "entire API" block is the
actual API.

**Why the shape matters.** Three design decisions do the work. First, a node
returns the whole next state, which deletes reducers, LangGraph's largest
subsystem and the source of its best-known foot-gun (lost updates on
fan-out). Second, every node has exactly one outgoing edge, static or routed,
so `Compile` can validate the whole graph and report every mistake at once
with `errors.Join`. Third, a run is a `(node, state)` pair advancing, so the
same `Step` value is the stream event, the checkpoint, and the resume point.
Interrupts are an error type that keeps its state; plain errors yield the last
good step so a rerun retries the failure; `StreamThread` saves successful and
interrupted steps and skips failed ones. Three conventions, pinned by tests.

**Evidence, not adjectives.** 420 library lines, 723 test lines, CI checks
the number. Zero dependencies, verified by `go.mod`. A 1.86 MB static binary
with 2.5 MB peak RSS for the ReAct example. Go cold starts on Lambda are
roughly 3x to 8x faster than Python in public benchmarks. LangGraph's
checkpoint serializers had three CVEs in twelve months; MiniGraph ships no
serializer. The credible Go alternatives are 8k to 82k lines each.

**What it is not.** It is not a durable-execution engine: no crash detection,
no work queue, no exactly-once, no history beyond the latest step. It has no
token streaming, no retry policy, no provider clients, no tracing UI. These
are deliberate; each composes from outside or belongs in Temporal. Say so
up front and the comparison with Eino, ADK, and LangGraph becomes a choice
about how much runtime a team wants to own, which is the conversation
MiniGraph wins.

**Ask.** If you are in Go and want the whole agent runtime in your head, read
the four files. It takes one coffee.
