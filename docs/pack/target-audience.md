# MiniGraph target audience

Who MiniGraph is for, what they are trying to do, where they are, and what
they will say when they first see the README. Six personas, then an explicit
"not for" list. Community sizes, venues and survey figures were verified on
2026-09-11; sources are linked inline.

## Market context in three numbers

- The 2025 Go Developer Survey (5,379 respondents) found 78% are not building
  AI-powered features into their Go software, 11% build "ML models, tools, or
  agents", and 55% build both CLIs and API services
  ([go.dev/blog/survey2025](https://go.dev/blog/survey2025)). The pool of Go
  developers shipping agents is small and the pool that *could* is large.
- The most-starred Go AI frameworks are ports or vendor kits: langchaingo
  9.7k stars, Eino 13.0k, Google ADK Go 8.8k; the six community LangGraph
  ports on GitHub top out at 304 stars (`smallnest/langgraphgo`) and 290
  (`tmc/langgraphgo`, unmaintained since April 2024) (GitHub API, 2026-09-11).
  Nobody owns "LangGraph's ideas, in idiomatic Go, small".
- Go engineers "tend to avoid frameworks and prefer the standard library plus
  small dependencies"; a bare agent loop is "about forty lines of Go", but the
  durable boundary between steps, state persistence and human-in-the-loop
  points are what teams end up bolting on with Temporal or Hatchet
  ([Zep, Building Agents in Go Without a Framework](https://blog.getzep.com/agentic-development-in-go/)).
  Those three things are exactly what the remaining 380 lines of MiniGraph are.

## Persona 1: Go backend engineer adding an LLM feature to an existing service

**Job to be done.** Add one agentic feature (triage, enrichment, a support
bot, a report generator) to a Go service that already exists, ship it behind
the same deploy pipeline, and be able to explain it in code review.

**Pains with current options.**
- A Python sidecar means a second runtime, a second CI matrix, a second
  on-call surface.
- langchaingo "trails the original in coverage and pace"
  ([Fastio, 2026](https://fast.io/resources/best-ai-agent-frameworks-for-golang-2026/));
  Eino and ADK pull in large dependency trees and vendor concepts.
- Hand-rolled loops work until someone asks for "pause and ask a human" or
  "resume after the pod restarts".

**What they search for.** "golang llm agent library", "go langgraph
equivalent", "go agent loop tool calling", "human in the loop go", "durable
workflow go without temporal", "go state machine library generic".

**Where they hang out.** [r/golang](https://www.reddit.com/r/golang/);
[Gophers Slack](https://invite.slack.golangbridge.org/) (`#general`,
`#golang-newbies`, plus channels like `#showandtell` to confirm on join);
[Gophers Discord](https://discord.com/invite/golang) (44k members);
[Go Forum](https://forum.golangbridge.org/);
[Golang Weekly](https://golangweekly.com/) (616+ issues since 2015);
[Cup o' Go](https://cupogo.dev/episodes) podcast; company Slack `#ai` channels.

**Objections and answers.**
- "I can write this loop myself." Yes, in forty lines; the checkpoint, resume
  and interrupt semantics are the other 380, tested, and you can still read
  all of it.
- "Where are the model bindings?" A node is `func(ctx, S) (S, error)`; call
  your existing client inside it. No adapter to learn or to be abandoned.
- "Is it maintained?" Point at CI, coverage (97.8%), the pinned invariants in
  `AGENTS.md`, and the release cadence in `strategy.md`.

**Success looks like.** The feature ships in the existing binary; the graph
is drawn in the design doc and the code matches it; a teammate reviews the
whole runtime before approving the dependency.

## Persona 2: Platform or infra team refusing a Python sidecar

**Job to be done.** Provide an "agents" capability to product teams without
introducing a Python runtime, a new deploy target, or a framework whose
upgrade cadence the platform team does not control.

**Pains with current options.**
- LangGraph's runtime and platform tier are Python and JavaScript; the
  self-hosted LangGraph stack has had a remote code execution flaw chain
  disclosed in June 2026
  ([The Hacker News](https://thehackernews.com/2026/06/langgraph-flaw-chain-exposes-self.html)),
  which is the kind of surface a platform team is asked to own.
- Large frameworks are hard to vendor, audit and pin; "we spent as much time
  understanding internals as building features" is the recurring complaint
  ([dev.to survey of LangChain complaints](https://dev.to/mujib77/i-spent-a-week-researching-why-devs-hate-supabase-langchain-posthog-neon-and-what-i-found-6f7)).

**What they search for.** "agent runtime go single binary", "langgraph self
hosted alternative", "checkpointer interface go", "durable execution go
library", "workflow engine go minimal dependencies", "govulncheck zero
dependencies".

**Where they hang out.** Gophers Slack `#performance`, `#kubernetes`-style
channels; [r/devops](https://www.reddit.com/r/devops/) and
[r/sre](https://www.reddit.com/r/sre/); CNCF and platform-engineering Slacks;
[GopherCon](https://www.gophercon.com/) and
[GopherCon EU](https://www.gophercon.eu/) hallway tracks;
[console.dev](https://console.dev/) newsletter for tool discovery.

**Objections and answers.**
- "Zero dependencies means you reimplemented things badly." It means the
  runtime uses `context`, `errors`, `iter`, `maps`, `slices`, `sync` and
  nothing else; there is nothing to reimplement. Persistence is an interface
  they implement over the store they already run.
- "What about audit and supply chain?" `govulncheck` on a zero-dependency
  module checks only the standard library; the whole runtime is a one-hour
  code read; a `SECURITY.md` and an OpenSSF Scorecard are on the roadmap.
- "Will it be abandoned?" Stability is a feature here: the API is small
  enough to be finished. See the v1.0 criteria in `strategy.md`.

**Success looks like.** MiniGraph is on the internal allow-list; a template
service ("agent-service-starter") with a Postgres `Checkpointer` exists;
product teams add nodes, not infrastructure.

## Persona 3: Solo developer shipping a CLI agent as a single binary

**Job to be done.** Build a coding assistant, a research tool, a
personal automation, or a paid CLI, and distribute it as one file via
`go install` or GitHub Releases.

**Pains with current options.**
- Python tools need users to manage environments; Node tools need `npm`. Go's
  static binary is the selling point, and a Python-shaped framework throws it
  away.
- They want a loop that can pause for `y/n` on the terminal and resume, and
  a way to survive Ctrl-C mid-run.

**What they search for.** "build a coding agent in go", "go react agent
example", "ollama go agent", "cli agent go tool calling loop", "go
iter.Seq2 example", "resume interrupted agent run".

**Where they hang out.** [r/golang](https://www.reddit.com/r/golang/),
[r/LocalLLaMA](https://www.reddit.com/r/LocalLLaMA/),
[r/LLMDevs](https://www.reddit.com/r/LLMDevs/) (161k members,
[GummySearch](https://gummysearch.com/r/LLMDevs/)),
[r/AI_Agents](https://www.reddit.com/r/AI_Agents/) (425k members,
[Prowlo](https://prowlo.com/tools/subreddit-stats/ai_agents));
[Hacker News](https://news.ycombinator.com/); YouTube channels such as
[Anthony GG](https://www.youtube.com/@anthonygg_), Melkey and Dreams of Code
([2026 channel list](https://learnwithpath.com/blog/best-youtube-channels-for-go-programming-2026));
Charm's Bubble Tea community (they already ship TUIs in Go).

**Objections and answers.**
- "PocketFlow does this in 100 lines of Python." Same spirit, different
  runtime; MiniGraph is typed, compiled, and adds checkpoints and interrupts
  that PocketFlow leaves out.
- "I need streaming tokens in the terminal." Stream inside the node with your
  client's API and print; the graph streams steps, the node streams tokens.

**Success looks like.** `go run ./examples/react` works in under a minute;
their own tool ships as a release binary; the approval flow is
`examples/approval` with a real prompt.

## Persona 4: LangGraph user with Go in the stack, tired of size and upgrades

**Job to be done.** Move (or duplicate) an agent from a Python LangGraph
service into the Go service that already owns the data, without relearning
the mental model.

**Pains with current options.**
- LangGraph core plus checkpoints is roughly 33,700 lines (README claim,
  worth re-measuring each release); upgrades touch reducers, `Send`,
  subgraph semantics and platform APIs they do not use.
- The Go ports either chase feature parity (`smallnest/langgraphgo`,
  "aims for feature parity", published January 2026) or stalled
  (`tmc/langgraphgo`, last push April 2024).

**What they search for.** "langgraph go", "langgraph golang port",
"langgraph alternative typed state", "StateGraph in go", "langgraph
interrupt equivalent go", "langgraph checkpointer postgres go".

**Where they hang out.** LangChain's Discord and GitHub Discussions;
[r/LangChain](https://www.reddit.com/r/LangChain/) (verify current rules on
the sidebar); [Latent Space](https://www.latent.space/) newsletter and
podcast; [AI Engineer World's Fair](https://www.ai.engineer/worldsfair)
(June 29 - July 2, 2026, San Francisco; next edition 2027); the
[langchaingo Discord](https://github.com/tmc/langchaingo).

**Objections and answers.**
- "No reducers, no `Send`, no subgraph API: it is not LangGraph." Correct,
  and the README's mapping table says which idea replaces each: whole-state
  return replaces reducers, `Parallel` replaces `Send`, `App.Invoke` as a
  node replaces the subgraph API.
- "No time travel or checkpoint history." The `Checkpointer` interface is two
  methods; store every Step under an index and you have history.
- "No LangSmith." Wrap the `Stream` loop with OpenTelemetry; every Step
  carries the node name.

**Success looks like.** A one-page migration note maps their graph node by
node; the Go version has fewer lines than the Python one; their
`Checkpointer` writes to the same Postgres.

## Persona 5: Learner or evaluator who wants to understand agent runtimes

**Job to be done.** Understand what an agent runtime actually does
(state, cycles, checkpoints, interrupts) by reading a complete, tested one;
or evaluate whether their team needs a framework at all.

**Pains with current options.**
- Reading LangGraph to learn the ideas means reading tens of thousands of
  lines; the tutorials teach the API, not the mechanism.
- Blog posts explain "agents" without a runnable reference.

**What they search for.** "how does langgraph work internally", "agent
runtime explained", "minimal agent framework source", "state machine agent
loop", "nanoGPT for agents", "read the whole framework".

**Where they hang out.** Hacker News (the audience that starred minGPT
24.9k, nanoGPT 63.0k, llm.c 31.0k, tinygrad 33.6k); university courses and
reading groups; [Latent Space](https://www.latent.space/); Bluesky and X
"AI engineering" circles; dev.to and Medium.

**Objections and answers.**
- "Toy." The whole runtime has more test code than library code and pinned
  invariants; `examples/approval` is a durable HITL flow, not a demo loop.
- "Go, not Python." That is the point: the compiler checks the state, and
  the code is short enough that language is not a barrier.

**Success looks like.** They read all four files in one sitting, can draw
the three yield conventions from memory, and cite the annotated-source post.

## Persona 6: DevOps or SRE building operations automations

**Job to be done.** Automate runbooks and incident response with an LLM in
the loop: gather signals in parallel, propose an action, require an approval,
execute, and survive the fact that the automation itself runs on the
infrastructure being fixed.

**Pains with current options.**
- Approval gates and durable resume are the hard parts, and most agent
  libraries treat them as platform features.
- Ops tooling is already Go (Kubernetes, Terraform providers, Prometheus
  exporters); a Python agent is an outlier on the on-call laptop.

**What they search for.** "llm runbook automation go", "approval gate
workflow go", "human approval before kubectl agent", "incident agent
checkpoint resume", "parallel fan-out go context cancel".

**Where they hang out.** [r/devops](https://www.reddit.com/r/devops/),
[r/sre](https://www.reddit.com/r/sre/), [r/kubernetes](https://www.reddit.com/r/kubernetes/);
Kubernetes and CNCF Slacks; SREcon and KubeCon; Gophers Slack.

**Objections and answers.**
- "I would never let an agent run `kubectl` unattended." Agreed; the
  `approve` node in `examples/approval` is the unattended boundary, and the
  thread survives a restart while waiting.
- "What if the process dies mid-remediation?" Failed steps are not saved, so
  a rerun retries the failed step; successful steps are not repeated.

**Success looks like.** A runbook agent runs as a Kubernetes Job, pauses in
Slack for approval, and resumes from a checkpoint stored in the cluster's
existing database.

## Not for

- Teams standardized on Python or TypeScript with no Go in the stack; use
  LangGraph, Pydantic AI or smolagents.
- Anyone who needs per-key reducers, retry policies, token streaming at the
  graph level, provider bindings, or a hosted platform; these are out of scope
  by design (`AGENTS.md`).
- Projects that need checkpoint history, time travel or branching threads
  out of the box; the interface allows it, the library does not ship it.
- Dynamic graphs whose topology the model decides at run time; MiniGraph
  graphs are fixed at `Compile` (build a new one per request if needed).
- Distributed execution across machines; `Parallel` is goroutines in one
  process. Use Temporal or Hatchet for cross-process durability.
- Users who want a visual builder, a studio, or a UI.
- Anyone whose organization measures frameworks by feature checklists; the
  checklist is short on purpose.
