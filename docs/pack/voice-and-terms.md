# How MiniGraph talks

## The voice in one line

**A senior engineer who read the whole thing and is telling you what it does.** Calm, exact, slightly dry. Confident because the claims are checkable, not because the adjectives are big.

## Rules

1. **Say the number, then what it is measured against.** "420 lines, comments included, against ~33,700 for LangGraph core + checkpoints." Never "tiny", "blazing", "lightweight" on their own.
2. **Respect LangGraph.** It is the reference design and it is excellent. We are the same ideas, in Go, without the machinery. Never mock it; compare it.
3. **Prefer the verb to the adjective.** "A node returns the whole next state" beats "simple, elegant state handling."
4. **Write for someone who will read the source.** Every claim on the site should be verifiable by opening one file. Link to the file when you can.
5. **Admit scope.** "Out of scope, on purpose" is a feature of the voice. List what we do not do and how to compose it from outside. Never hide a gap.
6. **Short sentences, real paragraphs.** No bullet walls in prose pages. Bullets are for lists of parallel items.
7. **No hype vocabulary.** Banned: revolutionary, game-changing, blazing, next-gen, supercharge, unleash, effortless, magical, seamless, robust (unless about a specific failure mode), powerful, cutting-edge, AI-powered.
8. **Second person, present tense.** "You return an Interrupt; the run pauses." Not "users can leverage interrupts."
9. **Code over prose when code is shorter.** If a sentence would take longer than the four lines of Go it describes, show the Go.
10. **One idea per headline.** Headlines are statements, not teasers: "Every step is a checkpoint", not "Durability, reimagined."

## Key terms (use these exactly)

| Term | Meaning | Not |
|---|---|---|
| **MiniGraph** | the library. One word, capital M and G. | Minigraph, mini-graph, minigraph (except in code/import paths) |
| **node** | `func(ctx, S) (S, error)`; transforms the whole state | step, task, agent, tool |
| **router** | `func(ctx, S) (string, error)`; picks the next node from the state | conditional edge (use only when mapping from LangGraph) |
| **edge** | a static link `from → to`; a router is an edge that reads the state | transition |
| **state** | the single typed value flowing through the graph; a Go struct | context (that word is for `context.Context`), memory |
| **graph** | the builder (`New[S]()`) before `Compile` | workflow, pipeline (use those for what the user is building, not for the object) |
| **app** | the compiled, immutable graph (`*App[S]`) | runnable, executor |
| **run** | one execution from a Step to End or an error | invocation, session |
| **step** | one executed node and the state it produced; `Step[S]{Node, State}`; also a checkpoint | event, frame, tick |
| **checkpoint** | any Step you keep to resume from | snapshot, save point |
| **thread** | a named durable run persisted by a Checkpointer | conversation, session |
| **Checkpointer** | the two-method interface `Save`/`Load` | saver (except `MemorySaver`, the type), store, backend |
| **interrupt** | a node returning `*Interrupt`; the run pauses and keeps the returned state | breakpoint, pause (fine in prose, but the noun is interrupt) |
| **resume** | `InvokeFrom`/`StreamFrom` with a Step; routing restarts *from* that node | replay, restart |
| **parallel / fan-out** | `Parallel(merge, branches...)`: branches run concurrently, `merge` folds results, counts as one step | superstep, Send, map-reduce (map-reduce is a pattern built on it) |
| **subgraph** | a compiled `App` used as a node via `app.Invoke` | nested graph |
| **MaxSteps** | per-run cap on node executions, default 25 | recursion limit (LangGraph's term; use in the mapping table only) |
| **zero dependencies** | `go.mod` has only the module line and `go 1.24` | "no deps", "dependency-free" |
| **420 lines** | `graph.go + checkpoint.go + interrupt.go + parallel.go`, comments included, enforced by CI | "~400", "under 500" |
| **1.2%** | 420 / ~33,700 (LangGraph core + checkpoint packages, Python) | "100× smaller" |

## Sentence patterns that sound like us

- "A node returns the whole next state, so there are no reducers."
- "Every yielded Step is a checkpoint. Interrupts, durability, and retries fall out of that one fact."
- "If you are in Python, use LangGraph. If you are in Go and want the runtime in your head, this is it."
- "Out of scope, on purpose: …"
- "Read it over one coffee."

## Sentence patterns that do not

- "MiniGraph makes building AI agents effortless!"
- "Unlock the power of graph-based orchestration."
- "The ultimate LangGraph alternative for Go developers."
- "Battle-tested, production-grade, enterprise-ready."

## Where the voice shows up

| Surface | Register |
|---|---|
| README / site home | Thesis first, then the sixty-second example, then the whole API. Confident, brief. |
| Tutorials | Second person, numbered walkthrough, real code, real output, one "what to notice" callout per pattern, one "make it real" section pointing at where a model/store plugs in. |
| Doc comments | User-facing documentation; explanatory; include runnable snippets (see `interrupt.go`). |
| Issues / PR replies | Friendly, direct, quote the invariant that applies. Decline scope creep by naming the out-of-scope list and offering the compose-from-outside path. |
| Show HN / Reddit | Lead with the honest number and the "read all of it" promise, then the LangGraph mapping table. Invite criticism of the design decisions explicitly. |
| Release notes | What changed, the new line count, what it means for resume semantics if anything. |

## Handling common pushback

| They say | We say |
|---|---|
| "It's a toy." | It is 420 lines with more test code than library code, and the three resume conventions are pinned by tests. Here is the file. What is missing for your case? |
| "No reducers means no multi-writer state." | Correct. Parallel gives you `merge`, which is the reducer, written by you, once, where the fan-out happens. |
| "No token streaming." | Right — a node is a function; stream tokens inside it with your client. Step streaming is what the engine gives you. |
| "Why not LangGraph.js / Python?" | If you are there, stay there. This is for teams that are already in Go. |
| "Will it grow?" | The out-of-scope list is in AGENTS.md. Proposals to add reducers, retry policies, or bindings get a compose-from-outside answer. |
