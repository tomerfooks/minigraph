# MiniGraph: alternatives analysis

Research date: 2026-09-11. Star counts and release dates were read from GitHub on
that day; line counts were measured on shallow clones with `find … | wc -l`,
excluding tests and examples, and are approximate. "LOC" means non-test source
lines of the library itself, not docs or examples.

MiniGraph for reference: 420 library lines (`graph.go` 239, `checkpoint.go` 93,
`interrupt.go` 31, `parallel.go` 57), 723 test lines, zero dependencies,
`go 1.24`, MIT. A compiled `examples/react` is a 1.86 MB static binary
(`-ldflags="-s -w"`, `CGO_ENABLED=0`) with a 2.5 MB peak RSS on macOS.

---

## 1. LangGraph (Python)

- **What it is.** "Low-level orchestration framework for building stateful
  agents", the reference design MiniGraph mirrors.
  [github.com/langchain-ai/langgraph](https://github.com/langchain-ai/langgraph)
- **Size.** 41.5k stars, 7.0k forks, MIT, created Aug 2023. Current core
  release `langgraph==1.2.11` (Aug 2026); 1.0 GA was
  [Oct 22, 2025](https://changelog.langchain.com/announcements/langgraph-1-0-is-now-generally-available).
  Measured: `libs/langgraph` 27,951 LOC + `libs/checkpoint` 5,898 LOC =
  **33,849 LOC** for the engine plus checkpoint base (the README's "~33.7k").
  The whole `libs/` tree (prebuilt, sdk, cli, postgres/sqlite checkpointers) is
  77,057 LOC.
- **Dependencies.** Six direct (`langchain-core`, `langgraph-checkpoint`,
  `langgraph-sdk`, `langgraph-prebuilt`, `xxhash`, `pydantic`). A clean
  `uv pip install langgraph` on Python 3.14 pulled **38 packages, 29 MB** of
  site-packages before any model provider is installed.
- **State model.** `TypedDict`/Pydantic/dataclass schema; each key may carry a
  reducer via `Annotated` (default last-write-wins; `add_messages` for chat).
  Execution is a Pregel-style loop of supersteps; parallel branches merge via
  reducers. Dynamic fan-out via `Send`, combined routing+update via `Command`.
  Default recursion limit is 1000 since 1.0.6.
  [Graph API docs](https://docs.langchain.com/oss/python/langgraph/graph-api)
- **Durability.** Checkpointer saves a snapshot per superstep, keyed by
  `thread_id`; in-memory, SQLite, Postgres, Redis, Mongo, DynamoDB backends.
  Resume re-executes the interrupted node from its start, so side effects must
  be wrapped in `@task` and made idempotent.
  [Durable execution docs](https://docs.langchain.com/oss/python/langgraph/durable-execution)
- **Human-in-the-loop.** `interrupt(value)` inside a node; resume with
  `Command(resume=...)`. [Interrupts docs](https://docs.langchain.com/oss/python/langgraph/interrupts)
- **Parallelism.** Supersteps run ready nodes concurrently; `Send` for
  map-reduce. Python-level concurrency (threads/async), not OS parallelism.
- **Platform.** LangGraph Platform / LangSmith Deployment: Developer $0/seat,
  Plus $39/seat/month with one free small serverless deployment, Enterprise
  custom with self-hosted options; compute metered in LCUs at $1.50.
  [langchain.com/pricing](https://www.langchain.com/pricing). Third-party
  breakdowns cite 100k node executions included then $0.001 each
  ([TrueFoundry](https://www.truefoundry.com/blog/langgraph-pricing),
  [ZenML](https://www.zenml.io/blog/langgraph-pricing)).
- **Criticism.** (a) Reducers are an implicit concurrency policy that "reviewers,
  diagrams, and type checkers all miss"
  ([Ranjan Kumar](https://ranjankumar.in/langgraph-reducers-concurrent-state-writes));
  lost updates on fan-out are a recurring support topic
  ([itsourcecode](https://itsourcecode.com/runtimeerror/langgraph-state-reducer-conflicts-fix/)).
  (b) "Programming languages already are graphs with compile-time validation"
  ([Honchar](https://www.vitaliihonchar.com/insights/go-ai-agent-library)).
  (c) A June 2026 chain of CVE-2025-67644 (SQLi in SQLite checkpointer) and
  CVE-2026-28277 (unsafe msgpack deserialization of checkpoints, `<1.0.10`)
  gave RCE on self-hosted deployments
  ([The Hacker News](https://thehackernews.com/2026/06/langgraph-flaw-chain-exposes-self.html)).
  (d) Checkpointing alone is not durable execution: no failure detection,
  no duplicate-run prevention, no task queue
  ([Diagrid](https://www.diagrid.io/blog/checkpoints-are-not-durable-execution-why-langgraph-crewai-google-adk-and-others-fall-short-for-production-agent-workflows)).
- **Choose LangGraph when** you are in Python, want the LangChain integration
  catalog, LangSmith tracing, prebuilt agents, time-travel, or a managed
  deployment. **Choose MiniGraph when** you are in Go and want the same
  shape with compile-time state and no runtime schema.

## 2. LangGraph.js

- 3.3k stars, 576 forks, MIT, `@langchain/langgraph` 1.0.47.
  [github.com/langchain-ai/langgraphjs](https://github.com/langchain-ai/langgraphjs)
- Measured `libs/langgraph-core/src` 30,633 TS LOC + `libs/checkpoint/src`
  3,326 = ~34k LOC. Peer deps `@langchain/core` and `zod`.
- Same concepts as Python (Annotation/zod state, reducers, `Send`, `Command`,
  `interrupt`), same Platform. Zod schemas give runtime, not compile-time,
  state validation; TypeScript types are erased at runtime.
- Choose it for Node/edge runtimes and React UI SDKs. Not a Go option.

## 3. CloudWeGo Eino (ByteDance) — Go

- **What it is.** "The ultimate LLM/AI application development framework in
  Go": components (ChatModel, Tool, Retriever), a `compose` orchestration layer
  (Chain/Graph/Workflow), and an ADK for agents.
  [github.com/cloudwego/eino](https://github.com/cloudwego/eino)
- **Size.** 13.0k stars, 1.1k forks, Apache-2.0, created Dec 2024. Latest
  stable `v0.9.14` (Aug 13, 2026); `v0.10.0-alpha.32` (Sep 11, 2026); no 1.0.
  Measured **59,203 LOC** (`compose` 11,863; `adk` 28,610; `schema` 7,147).
  11 direct / 38 total modules in `go.mod`; `go 1.18`.
- **State model.** `NewGraph[I, O]()` is typed on graph input/output; each
  node's output type must match the next node's input ("type alignment").
  Shared graph state is opt-in via `WithGenLocalState` plus
  `StatePreHandler`/`StatePostHandler`/`ProcessState`.
  [Chain/Graph intro](https://www.cloudwego.io/docs/eino/core_modules/chain_and_graph_orchestration/chain_graph_introduction/)
- **Durability.** `CheckPointStore` (`Get`/`Set` with `[]byte`), keyed by
  `WithCheckPointID`; JSON or gob; custom types must be registered with
  `schema.RegisterName`; unexported fields are not persisted; the graph
  structure must be identical on resume.
  [Interrupt & CheckPoint manual](https://www.cloudwego.io/docs/eino/core_modules/chain_and_graph_orchestration/checkpoint_interrupt/)
- **Human-in-the-loop.** Static (`WithInterruptBeforeNodes/AfterNodes`) and
  dynamic (`Interrupt`, `StatefulInterrupt`, `CompositeInterrupt`) interrupts,
  addressed by hierarchical `AddressSegment`s; resume via `Resume`,
  `ResumeWithData`, `BatchResumeWithData`, or `Runner.ResumeWithParams` in the
  ADK. [ADK HITL guide](https://www.cloudwego.io/docs/eino/core_modules/eino_adk/agent_hitl/)
- **Parallelism.** Graph-level parallel branches and a Pregel-style runner;
  four streaming paradigms (Invoke/Stream/Collect/Transform) handled by the
  framework.
- **Strengths.** Production use at ByteDance ("battle-tested internally for
  over six months" before release,
  [CloudWeGo](https://www.cloudwego.io/docs/eino/overview/eino_open_source/));
  first-class streaming; rich provider ecosystem (eino-ext); visual devtools.
- **Weaknesses.** Large API surface; still 0.x with alpha releases; English
  documentation is thin ("Almost nothing has been written about it in
  English", [dev.to, Feb 2026](https://dev.to/itodona/4-eino-bytedances-langgraph-for-go-building-a-6-node-agentic-workflow-4nj2));
  the addressing/composite-interrupt model puts real burden on the developer.
- **Choose Eino when** you need token streaming through the graph, a
  component ecosystem, or multi-agent ADK patterns in Go. **Choose MiniGraph
  when** you want one state type, one edge per node, and an engine you can
  read in full.

## 4. langchaingo — Go

- 9.7k stars, 1.1k forks, MIT, `v0.1.14` (Oct 20, 2025); last push Jan 11,
  2026. Measured **50,462 LOC**; **64 direct / 286 total** modules; `go 1.24`.
  [github.com/tmc/langchaingo](https://github.com/tmc/langchaingo)
- A port of classic LangChain: LLM clients, chains, prompts, memory,
  vector stores, MRKL/conversational/OpenAI-functions agent executors. No
  graph engine, no checkpointer, no interrupt primitive.
- Choose it for provider adapters and RAG building blocks; it is
  complementary to MiniGraph (call it inside a node), not a substitute. Watch
  its release cadence and dependency weight.

## 5. Firebase Genkit for Go

- Repo 6.4k stars, Apache-2.0. Genkit Go 1.0 shipped
  [Sep 10, 2025](https://developers.googleblog.com/en/announcing-genkit-go-10-and-enhanced-ai-assisted-development/);
  latest Go tag `go/v1.13.1` (Sep 3, 2026). Measured Go module **75,844 LOC**
  (`core` 6,328; `ai` 18,885; `genkit` 5,912); **40 direct / 130 total**
  modules; `go 1.25`. [github.com/firebase/genkit](https://github.com/firebase/genkit)
- **Model.** Flows are typed functions with tracing, a dev UI, and one-line
  HTTP deployment; `generate()` runs the tool-calling loop. Not a graph.
- **Durability / HITL.** Interrupts are tool-level: a tool interrupts
  `generate()`, the developer stores the messages and calls `generate()` again
  with `resume` data. Persistence of that history is the developer's job.
  [Interrupts docs](https://genkit.dev/docs/interrupts/). v1.13.0 added
  "agent progress preservation and background sub-agents with resumable
  snapshots" ([releases](https://github.com/firebase/genkit/releases)).
- Choose Genkit for Gemini/Firebase/Cloud Run integration, tracing UI, and
  prompt tooling. Choose MiniGraph when control flow is the product and you
  do not want a framework runtime.

## 6. Google ADK Go

- 8.8k stars, 1.0k forks, Apache-2.0, created May 2025. 1.0 on
  [Mar 31, 2026](https://developers.googleblog.com/adk-go-10-arrives/);
  2.0 on [Jun 30, 2026](https://developers.googleblog.com/announcing-adk-go-20/)
  introduced "a new graph-based orchestration engine"; latest `v2.3.0`
  (Aug 31, 2026). Measured **66,627 LOC**; **35 direct / 97 total** modules;
  `go 1.26`. [github.com/google/adk-go](https://github.com/google/adk-go)
- **Model.** LLMAgent, Sequential/Parallel/Loop workflow agents, plus a
  `workflow` package with function/agent/tool/join/dynamic nodes and typed
  routes. State lives in sessions (`session.Service`: in-memory, database,
  Vertex AI). HITL via `RequestInputEvent`, resumed by handoff or re-entry,
  durable across restarts through session persistence. Built-in retry with
  backoff, per-node timeouts, concurrency limits, A2A protocol.
- Choose ADK Go for Gemini-centric multi-agent systems with Google Cloud
  deployment. It is the most capable Go framework on this list and also the
  largest; MiniGraph is the opposite trade.

## 7. Temporal Go SDK (durable-workflow alternative)

- Server 23k stars; `sdk-go` 962 stars, `v1.48.0` (Aug 18, 2026), MIT,
  `go 1.26`. Measured SDK **81,746 LOC**; 15 direct / 25 total modules.
  [github.com/temporalio/sdk-go](https://github.com/temporalio/sdk-go)
- **Model.** Workflows are deterministic Go functions replayed from event
  history; every LLM/tool call goes in an Activity with retry policies,
  timeouts, heartbeats. HITL via Signals/Updates. Requires a Temporal
  Service (Cassandra/MySQL/PostgreSQL, optional Elasticsearch) or Temporal
  Cloud (Essentials from $100/month or 5% of consumption, Business $500/month
  with 2.5M actions, $50 per million actions beyond)
  ([pricing](https://docs.temporal.io/cloud/pricing),
  [self-hosted](https://docs.temporal.io/self-hosted-guide/deployment)).
  Official Go integration exists for Google ADK
  ([Durable AI](https://docs.temporal.io/ai)).
- **Strengths.** True durable execution: crash recovery, exactly-once-ish
  activities, timers, queues, multi-day waits, horizontal workers.
- **Weaknesses.** Operational footprint; determinism rules (no direct I/O,
  time, or goroutines in workflow code) and replay bugs
  ([TMPRL1100](https://github.com/temporalio/rules/blob/main/rules/TMPRL1100.md));
  per-action billing.
- Choose Temporal when you need workflows that survive infrastructure loss
  and run for days at scale. MiniGraph nodes can run inside a Temporal
  Activity; MiniGraph itself is not a durable-execution engine (see gaps).

## 8. Raw SDK loops in Go

- [`sashabaranov/go-openai`](https://github.com/sashabaranov/go-openai) 10.8k
  stars, Apache-2.0, pure client, no agent loop.
- [`openai/openai-go`](https://github.com/openai/openai-go) 3.5k stars,
  `v3.61.0`, official client.
- [`anthropics/anthropic-sdk-go`](https://github.com/anthropics/anthropic-sdk-go)
  1.2k stars, `v1.72.0`, includes a `toolrunner` helper for a tool-use loop.
- A hand-written `for { call model; run tools }` is ~50 lines and needs no
  library. It has no resume point, no interrupt convention, no step
  streaming, and no cycle guard; each of those gets reinvented per service.
  MiniGraph is the smallest step up from this that adds those four things.

## 9. Go LangGraph ports and clones (GitHub, Sep 2026)

| Project | Stars | LOC | Deps (direct/total) | Last push | Notes |
|---|---|---|---|---|---|
| [smallnest/langgraphgo](https://github.com/smallnest/langgraphgo) | 304 | 34,257 | 15/61 | 2026-07-16 | Aims at LangGraph parity: Redis/Postgres/SQLite checkpointers, interrupts, streaming modes, prebuilt ReAct/supervisor, `v0.8.5` (Jan 2026). State is untyped map-style. The most credible Go clone. |
| [tmc/langgraphgo](https://github.com/tmc/langgraphgo) | 290 | 141 | 0/3 | 2024-04-23 | `MessageGraph` over langchaingo; dormant since 2024. |
| [paulnegz/langgraphgo](https://github.com/paulnegz/langgraphgo) | 36 | 3,311 | 2/4 | 2025-09-05 | Fork of tmc with checkpointable graphs, streaming, Langfuse; depends on langchaingo. |
| [piotrlaczkowski/GoLangGraph](https://github.com/piotrlaczkowski/GoLangGraph) | 12 | 20,247 | 13/35 | 2026-08-07 | Batteries-included (REST server, Prometheus, tools); `v0.2.2`. |
| [dshills/langgraph-go](https://github.com/dshills/langgraph-go) | 8 | 17,146 | 16/78 | 2025-11-18 | Generics-typed state, deterministic replay, SQLite/MySQL persistence; alpha, stalled. |
| [futurxlab/golanggraph](https://github.com/futurxlab/golanggraph) | 8 | 3,540 | 10/29 | 2026-03-04 | Redis/in-memory checkpoints, MCP tools, langchaingo. |
| [denizumutdereli/langgraphgo](https://github.com/denizumutdereli/langgraphgo) | 1 | 7,935 | 6/63 | 2026-03-22 | Pregel-style, "parity with LangGraphJS is not complete". |
| [vitalii-honchar/go-agent](https://github.com/vitalii-honchar/go-agent) | 82 | — | — | 2026-05-09 | Not a graph: typed ReAct loop, explicitly anti-LangGraph. |
| [microsoft/agent-framework-go](https://github.com/microsoft/agent-framework-go) | 606 | — | — | active | Public preview of the .NET/Python Agent Framework for Go; workflows with checkpointing and HITL; handoff and Foundry hosting not yet implemented. |

Observation: every port with typed state and persistence is 8k–35k LOC; the
only sub-1k-LOC option (tmc) is dormant and untyped. None combines typed
generics, a checkpointer, interrupts, and parallel fan-out under 1k lines.

## 10. Non-Go frameworks (for calibration)

- **OpenAI Agents SDK** — Python 29.4k stars (`v0.22.2`, Sep 2026), JS 3.8k.
  Agents, handoffs, guardrails, sessions, HITL approvals; durability via the
  [Temporal integration](https://temporal.io/blog/announcing-openai-agents-sdk-integration).
  No graph; control flow is the LLM's handoff choice.
- **Pydantic AI** — 19.9k stars. Typed agents plus `pydantic-graph`; durable
  execution is delegated to Temporal, DBOS, or Prefect as attachable
  capabilities ([docs](https://pydantic.dev/docs/ai/capabilities/durable_execution/overview/)).
- **Mastra** — 27.9k stars, Apache-2.0 core with an enterprise-licensed `ee/`
  directory. Workflow engine with `.then/.branch/.parallel`, suspend/resume
  backed by storage. Reviewers note it "requires separate platform
  durability" ([Developers Digest](https://www.developersdigest.tech/blog/mastra-durable-typescript-agents)).
- **Vercel AI SDK** — 26.7k stars. `ToolLoopAgent` and tool approval in v6;
  durability via the separate Workflow DevKit `DurableAgent`
  ([AI SDK 6](https://vercel.com/blog/ai-sdk-6)).
- **Microsoft Agent Framework** — 13.5k stars, MIT; merged AutoGen and
  Semantic Kernel; 1.0 for Python/.NET on
  [Apr 3, 2026](https://visualstudiomagazine.com/articles/2026/04/06/microsoft-ships-production-ready-agent-framework-1-0-for-net-and-python.aspx);
  graph workflows, checkpointing, HITL request/response, OpenTelemetry.

Pattern across all of them: the industry has converged on (typed state,
graph or loop, checkpoint, interrupt, parallel step) as the core, and pushes
real durability to Temporal/DBOS-class runtimes. MiniGraph implements exactly
that core and nothing else.

---

## Feature matrix

| Capability | MiniGraph | LangGraph (Py) | Eino | langchaingo | Genkit Go | Temporal Go |
|---|---|---|---|---|---|---|
| Library LOC (measured) | 420 | ~33.8k core+ckpt (77k all libs) | ~59k | ~50k | ~76k (Go module) | ~82k (SDK) |
| Direct / total deps | 0 / 0 | 6 / 38 pkgs | 11 / 38 | 64 / 286 | 40 / 130 | 15 / 25 + server |
| Stars | new | 41.5k | 13.0k | 9.7k | 6.4k (repo) | 962 (SDK), 23k (server) |
| Stable release | 0.x | 1.2.11 | 0.9.14 (0.10 alpha) | 0.1.14 | 1.13.1 | 1.48.0 |
| Typed state | compile-time generic `S` | runtime schema (TypedDict/Pydantic) | per-node I/O generics + opt-in state | n/a | typed flow I/O | typed workflow args |
| Reducers / merge | whole-state return; merge in `Parallel` | per-key reducers | state handlers | n/a | n/a | n/a |
| Cycles | yes, `MaxSteps` (25) | yes, recursion limit (1000) | yes | agent executor loop | `generate()` loop | yes (deterministic code) |
| Conditional routing | `Router` | conditional edges, `Command`, `Send` | Branch | n/a | code | code |
| Streaming | per-step `iter.Seq2` | per-step + token modes | per-step + token streams | tokens via callbacks | token streaming | n/a |
| Checkpoint granularity | every `Step` | every superstep | on interrupt (`CheckPointStore`) | none | none (developer) | every event |
| Durable across crash | last step, via `Checkpointer` | yes, via checkpointer | on interrupt only | no | no | yes, fully |
| Interrupt / HITL | `*Interrupt` + `InvokeFrom` | `interrupt()` + `Command(resume)` | static + dynamic interrupts, addresses | no | tool interrupts | Signals / Updates |
| Parallel fan-out | `Parallel` node, goroutines | supersteps, `Send` | parallel branches | no | no | child workflows / activities |
| Subgraphs | free (`App.Invoke` is a `Node`) | supported | supported | n/a | flows call flows | child workflows |
| Retry policy | none (rerun thread) | Python `RetryPolicy` | none built-in | none | none | built-in |
| Time-travel / history | no (latest step only) | yes | limited | no | no | full history |
| LLM bindings | none | LangChain | eino-ext | many | many | none |
| Managed platform | none | LangGraph Platform | none | none | Firebase / Cloud Run | Temporal Cloud |
| Runs in one static binary | yes | no (Python) | yes | yes | yes | yes + server |

## Honest gaps

What MiniGraph lacks that matters, depending on the workload:

1. **No token streaming.** `Stream` yields whole steps. A chat UI that wants
   tokens must stream from inside a node (channel or callback); the engine
   cannot help. Eino, Genkit, LangGraph, and the Vercel SDK all do this.
2. **No per-key reducers.** Parallel branches return whole states and the user
   writes the merge. That is simpler but puts the concurrency policy on the
   caller; LangGraph's `add_messages`-style helpers do not exist.
3. **Latest-step-only persistence.** `Checkpointer` keeps one `Step` per
   thread; there is no history, no time travel, no forking a thread from an
   earlier point. LangGraph and Temporal keep full histories.
4. **No retry policy, timeout, or backoff.** A failed step is retried only by
   rerunning the thread. ADK Go 2.0 and Temporal ship these.
5. **Not durable execution.** Like every checkpoint framework (Diagrid's
   critique applies), MiniGraph does not detect crashed processes, prevent two
   workers resuming the same thread, or queue work. A node that partially
   ran before a crash re-runs on resume; side effects must be idempotent.
6. **Interrupts do not work inside `Parallel` branches**; they are downgraded
   to errors. LangGraph and Eino can pause inside fan-out.
7. **Serialization is the user's problem.** `Checkpointer[S]` stores `Step[S]`;
   persisting outside memory means writing a JSON/DB implementation. Every
   competitor ships at least SQLite/Postgres/Redis savers.
8. **No provider ecosystem, tracing, or dev UI.** No OpenTelemetry hooks, no
   LangSmith equivalent, no visualizer. Observability is `for step := range`
   plus your logger.
9. **Untested at scale and unknown maintainer base.** Seven commits, one
   author. Eino runs inside ByteDance; ADK and Genkit are Google-maintained.
10. **`MaxSteps` defaults to 25** versus LangGraph's 1000; long tool loops
    must raise it explicitly.
