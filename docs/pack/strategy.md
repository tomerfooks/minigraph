# MiniGraph open-source growth strategy

A plan for taking MiniGraph from 0 stars (repo created 2026-07-12) to a
project the Go and agents communities reach for, grounded in how comparable
small-and-readable projects grew. Figures were pulled from the GitHub API and
the linked sources on 2026-09-11.

## 1. What the comparables teach

| Project | Stars (2026-09) | Created | What drove adoption | Lesson for MiniGraph |
|---|---|---|---|---|
| [htmx](https://github.com/bigskysoftware/htmx) | 49.4k | 2020-04 | Essays plus memes on X, grown from 2k to 30k followers; a self-deprecating "htmx sucks" essay; a book ([htmx lore](https://htmx.org/essays/lore/), [SE Radio 671](https://se-radio.net/2025/06/se-radio-671-carson-gross-on-htmx/)) | A written thesis ("Go is the right runtime for agents") beats feature marketing; the README already has one. |
| [sqlc](https://github.com/sqlc-dev/sqlc) | 18.3k | 2019-06 | One clear job ("compile SQL to type-safe Go"), an [introduction post](https://conroy.org/introducing-sqlc), sustained HN threads from users ([example](https://news.ycombinator.com/item?id=28462162)) | Users' own write-ups drive later waves; make them easy to write. |
| [chi](https://github.com/go-chi/chi) | 22.8k | 2015-10 | "no external dependencies", 100% `net/http` compatible, core under 1000 LOC, production logos (Cloudflare, Heroku) ([README](https://github.com/go-chi/chi)) | Stdlib-shaped and small is a durable Go positioning; collect production users early. |
| [Bubble Tea](https://github.com/charmbracelet/bubbletea) | 44.9k | 2020-01 | A borrowed, well-known architecture (Elm), a large examples gallery, a strong brand | Borrow LangGraph's vocabulary deliberately; ship many small examples. |
| [instructor](https://github.com/567-labs/instructor) | 13.9k | 2023-06 | One primitive done well, docs-first, podcast appearances ([Latent Space](https://www.latent.space/p/instructor)) | Stay a single primitive; talk about the idea, not the library. |
| [Pydantic AI](https://github.com/pydantic/pydantic-ai) | 19.9k | launched 2024-12-02 | Brand leverage plus a feeling ("the FastAPI feeling") ([Real Python](https://realpython.com/pydantic-ai/)) | MiniGraph's feeling is "you can read all of it"; repeat it everywhere. |
| [smolagents](https://github.com/huggingface/smolagents) | 29.3k | 2024-12 | A falsifiable number in the pitch ("~1,000 lines") ([HF blog](https://huggingface.co/blog/smolagents)); the number later drifted to 1,813 lines while the README kept the claim ([analysis](https://www.aledlie.com/1813-lines-of-simplicity/)) | The number is load-bearing; MiniGraph's CI already enforces it. Keep that and say so. |
| [PocketFlow](https://github.com/The-Pocket/PocketFlow) | 11.2k | 2024-12 | "100-line LLM framework", a dev.to essay comparing LangChain's 405K lines ([essay](https://dev.to/zachary62/i-built-an-llm-framework-in-just-100-lines-here-is-why-35b0)) | A size comparison chart is a proven hook; MiniGraph's README has one. |
| [tinygrad](https://github.com/tinygrad/tinygrad) | 33.6k | 2020-10 | A public rule: "will always be below 1000 lines; if it isn't, we revert commits" ([HN](https://news.ycombinator.com/item?id=30910706), [issue #405](https://github.com/geohot/tinygrad/issues/405)) | Publish the rule, not just the number: "the runtime stays under 500 lines". |
| [minGPT](https://github.com/karpathy/minGPT) / [nanoGPT](https://github.com/karpathy/nanoGPT) / [llm.c](https://github.com/karpathy/llm.c) | 24.9k / 63.0k / 31.0k | 2020 / 2022 / 2024 | Readable reference implementations; "build from scratch to understand" ([Simon Willison on llm.c](https://simonwillison.net/2024/Apr/9/llmc/)) | The educational audience is large and star-generous; write the "read the whole runtime" post. |
| [langchaingo](https://github.com/tmc/langchaingo) / [Eino](https://github.com/cloudwego/eino) / [ADK Go](https://github.com/google/adk-go) | 9.7k / 13.0k / 8.8k | 2023 / 2024 / 2025 | Port of a brand; vendor backing | The Go AI ceiling is ~10-13k for well-backed projects; a solo library should target the hundreds-to-low-thousands. |
| LangGraph-in-Go ports ([smallnest](https://github.com/smallnest/langgraphgo), [tmc](https://github.com/tmc/langgraphgo), others) | 304 / 290 / 36 / 8 / 3 / 1 | 2024-2025 | Feature-parity ports with little promotion | Parity ports stall; a differentiated position ("1.2% of the code") plus promotion is the gap. |

Common thread: every project above sells one idea in one sentence, backs it
with something verifiable (a number, a benchmark, a compatibility claim), and
publishes writing that stands on its own without the library.

## 2. Positioning and messaging pillars

1. **Read all of it.** 420 lines, CI-enforced, more tests than code. The
   README line-count check is the differentiator smolagents lacked; mention
   the check itself in talks and posts.
2. **Go-native, not ported.** Typed state checked by the compiler, `iter.Seq2`
   streaming, `context` everywhere, zero dependencies, one binary. Mirror
   chi's language: "stdlib-shaped".
3. **LangGraph's ideas, LangGraph's words.** Keep the mapping table
   (`StateGraph` -> `New[S]`, `interrupt()` -> `&Interrupt{}`) so switching
   cost is one page.
4. **Everything else composes from outside.** Reducers, retries, token
   streaming, bindings are out of scope on purpose; state it as a promise,
   not an apology.

One-line pitch for every venue: *LangGraph for Go, 1.2% of the code: typed
state, cycles, checkpoints, interrupts, parallel, 420 lines you can read over
one coffee.*

## 3. Launch plan

### Pre-launch checklist (do before any public post)

Stars concentrate in the first 48 hours after a Hacker News post (92% of
star-getting is over by then), so documentation must be finished before
posting ([Show HN by the numbers, 188k posts](https://danfking.github.io/blog/2026/04/23/show-hn-by-the-numbers/)).

- Tag `v0.1.0` (SemVer; required later by awesome-go).
- Confirm [pkg.go.dev](https://pkg.go.dev/github.com/tomerfooks/minigraph)
  renders the doc comments and the runnable snippets.
- [Go Report Card](https://goreportcard.com/report/github.com/tomerfooks/minigraph)
  at A+; add a coverage badge (Codecov or Coveralls; current coverage 97.8%).
- Add `CONTRIBUTING.md`, `SECURITY.md`, issue and PR templates, enable
  Discussions, label 5-8 `good first issue`s (section 5).
- Fix the Interrupt-inside-Parallel doc/behavior mismatch recorded in
  `features.md` (either way, add the test).
- A 90-second terminal recording of `go run ./examples/react` and
  `examples/approval` (asciinema or a GIF) for social posts.
- Draft the annotated-source post (section 4) so it can be linked from the
  HN thread within the hour.

### Hacker News (Show HN)

- Rules: it must be something people can try, the author must be around to
  answer, no asking friends to upvote
  ([Show HN guidelines](https://news.ycombinator.com/showhn.html),
  [HN FAQ](https://news.ycombinator.com/newsfaq.html)).
- Title: `Show HN: MiniGraph - LangGraph for Go in 420 lines, zero deps`.
  Numbers and "zero deps" are concrete; avoid "blazing" and "simple".
- Timing: Monday 00:00 UTC (Sunday 7pm US Eastern) has the best odds of 50+
  points (10.8%); Thursday 06:00 UTC the worst (2.6%). Median Show HN scores
  2 points; 50 points is the top 6%
  ([Show HN by the numbers](https://danfking.github.io/blog/2026/04/23/show-hn-by-the-numbers/)).
- First comment: post a maintainer comment immediately with why it exists,
  what is out of scope, and the three yield conventions; HN rewards candor
  about limits.
- Conversion: about 1.4 GitHub stars per HN point within 48 hours; comments
  do not convert (r = 0.10). Plan for 70-200 stars from a 50-150 point post.
- If it sinks, email hn@ycombinator.com for the
  [second-chance pool](https://news.ycombinator.com/item?id=26998308); do
  not repost the same URL within days
  ([syften guide](https://syften.com/blog/hacker-news-marketing/)).

### Reddit

- **r/golang**: read the sidebar rules before posting (reddit.com could not
  be fetched for this document; confirm current flair and self-promotion
  rules there). Post as a technical write-up with the design decisions, not
  a link drop; disclose authorship; answer every comment for 24 hours.
  Site-wide, keep self-promotion under roughly 10% of activity
  ([2026 Reddit self-promotion guide](https://redship.io/blog/reddit-self-promotion-rules)).
- **r/LocalLLaMA**: self-promotion tolerated but policed; "frame the post as
  a lesson or a genuine contribution and the tool as context"; lead with a
  local-model demo (Ollama inside a node), not the library
  ([Intoru](https://intoru.ai/subreddits/localllama),
  [LaunchWake](https://www.launchwake.com/channels/r-localllama)).
- **r/LLMDevs** (161k) and **r/AI_Agents** (425k): post the "LangGraph vs
  Go" comparison there, a week after HN, with a different angle each time.
- Never cross-post identical text; stagger by days.

### Newsletters and lists

- **Golang Weekly**: no public submission form was found on the site or in a
  recent issue; the editors curate from HN, r/golang and X. Tag
  [@golangweekly](https://x.com/golangweekly) when posting, and reply to the
  newsletter email with a two-line pitch ([golangweekly.com](https://golangweekly.com/)).
- **awesome-go**: requires 5 months of history (eligible from 2026-12-12),
  an open-source license, a `go.mod`, at least one SemVer release, coverage
  >= 80%, links to pkg.go.dev, Go Report Card and a coverage report in the PR
  body, and responses to issues within ~2 weeks
  ([CONTRIBUTING.md](https://github.com/avelino/awesome-go/blob/main/CONTRIBUTING.md)).
  Submit in December under "Artificial Intelligence".
- **console.dev**: email hello@console.dev; they review 2-3 tools weekly and
  publish [selection criteria](https://console.dev/selection-criteria).
- **Hacker Newsletter** ([hackernewsletter.com](https://hackernewsletter.com/),
  60k+ readers) and [TLDR AI](https://tldr.tech/ai) pick from HN; a good HN
  thread is the submission.

### Chat communities

- [Gophers Slack](https://invite.slack.golangbridge.org/): share in the
  show-and-tell style channel once, then answer questions in `#general` and
  `#golang-newbies` when agents come up.
- [Gophers Discord](https://discord.com/invite/golang) (44k members): same
  rule, one post, then be useful.
- langchaingo Discord and LangChain Discord: only in threads where someone
  asks for Go.

### Podcasts

- Go Time ended on 2024-12-18 after 340 episodes
  ([changelog.com/gotime](https://changelog.com/gotime)); do not pitch it.
- Pitch [Cup o' Go](https://cupogo.dev/episodes) (weekly news; they cover
  new libraries), go podcast() and the Ardan Labs Podcast
  ([2026 list](https://podcast.feedspot.com/golang_podcasts/)).
- [Latent Space](https://www.latent.space/) featured instructor and Pydantic
  AI; realistic only after a visible HN result and one production user.

### Conferences

- GopherCon 2026 (Seattle, Aug 3-6) and GopherCon EU 2026 (Berlin, June
  15-18) have passed; the GopherCon CFP window was Jan 19 - Mar 4, so plan a
  2027 talk proposal in January ([gophercon.com](https://www.gophercon.com/),
  [gophercon.eu](https://www.gophercon.eu/)).
- [GoLab 2026](https://golab.io/), Bologna, Nov 1-3 2026: attend or propose
  a lightning talk if the CFP is still open.
- [AI Engineer World's Fair](https://www.ai.engineer/worldsfair) (next
  edition 2027): the agents audience; a talk titled "An agent runtime you can
  read in an hour".
- Full list: [Go wiki Conferences](https://github.com/golang/wiki/blob/master/Conferences.md).

### Written channels and cadence

- **dev.to and Medium**: one long post per month, cross-posted with a
  canonical URL (PocketFlow's dev.to essay was its main hook).
- **X and Bluesky**: two posts per week, alternating a code screenshot with a
  design note; tag @golangweekly, @golang. Follow htmx's pattern of essays
  over announcements.
- **LinkedIn**: one post per launch milestone, written for Persona 2
  (platform teams): "no Python sidecar".
- **YouTube**: a 15-minute walkthrough reading the four files top to bottom;
  pitch Anthony GG, Melkey or Dreams of Code for a guest walkthrough after
  the HN launch ([channel list](https://learnwithpath.com/blog/best-youtube-channels-for-go-programming-2026)).

## 4. Content plan (first six months)

| Month | Piece | Target persona | Venue |
|---|---|---|---|
| 1 | "Read the whole runtime": annotated walk through `graph.go` line by line, the three yield conventions, why `Step` is a checkpoint | 5, 4 | Blog, HN, dev.to |
| 1 | "LangGraph vs Go: the same agent, side by side" with line counts and a migration table | 4, 1 | r/LLMDevs, LinkedIn |
| 2 | "Human-in-the-loop without a platform": `examples/approval` on a Postgres `Checkpointer` (publish the 40-line saver as a gist, not a package) | 2, 6 | r/golang, Gophers Slack |
| 2 | "A ReAct agent on Ollama in Go", local models only | 3 | r/LocalLLaMA |
| 3 | "Why no reducers": the design essay | 5, 4 | Blog, X, Bluesky |
| 3 | Runbook agent with approval in Slack, running as a Kubernetes Job | 6 | r/devops, r/sre |
| 4 | Benchmarks: steps/second, allocations per step, `Parallel` scaling; compare with a hand-written loop to show overhead is near zero | 1, 2 | Blog |
| 4 | "What we said no to": feature requests declined and why | all | GitHub Discussions, X |
| 5 | Case study with the first production user | 1, 2 | Blog, Golang Weekly |
| 6 | v1.0 announcement and the compatibility promise | all | HN (not Show HN), r/golang |

## 5. Community mechanics

- **Good first issues** (small and real): a Postgres `Checkpointer` example
  under `examples/`, a fuzz test for `Compile`, a `Benchmark` for
  `StreamFrom`, a test for Interrupt inside `Parallel`, an `iter.Pull`
  usage example, a Bluesky-sized logo variant.
- **CONTRIBUTING.md**: state the line budget ("library files stay under 450
  lines; README counts must move with them; CI enforces"), the
  no-dependencies rule, the "test in the matching `*_test.go`" rule, and the
  out-of-scope list, so reviews cite policy rather than taste.
- **Issue templates**: bug (with a minimal graph), feature request (must
  answer "can this be built outside?"), question -> redirect to Discussions.
- **Discussions**: enable, with pinned "Show and tell" and "Design" categories.
- **Release cadence**: patch releases as needed, a minor release at most
  monthly, each with a `CHANGELOG.md` entry; use Go module versioning rules
  ([go.dev version numbers](https://go.dev/doc/modules/version-numbers)).
- **v1.0 criteria**: three months without an exported API change; two
  independent production users; fuzz targets for `Compile` and `StreamFrom`
  running in CI; the Interrupt-in-Parallel semantics documented and tested;
  awesome-go listing merged. After v1.0, breaking changes require a `/v2`
  path, which is a strong incentive to say no.

## 6. Trust signals

- Already present: pkg.go.dev badge, CI badge, Go Report Card badge, MIT
  license, `-race` tests, README line-count check.
- Add: coverage badge (97.8% now); `go test -fuzz` targets
  ([Go fuzzing](https://go.dev/doc/security/fuzz/)); `govulncheck` in CI
  ([tutorial](https://go.dev/doc/tutorial/govulncheck)); `SECURITY.md` with a
  disclosure address ([GitHub docs](https://docs.github.com/en/code-security/getting-started/adding-a-security-policy-to-your-repository));
  [OpenSSF Scorecard](https://securityscorecards.dev/) badge; published
  benchmarks in the README; a "used by" section once two users agree to be
  named.

## 7. Metrics and goals

Baselines: the best-promoted LangGraph-in-Go port has 304 stars; PocketFlow
reached ~8.4k in its first year with a full-time author and Python's
audience; a top-6% Show HN yields 70-200 stars in 48 hours.

| Horizon | Base | Target | Stretch | Leading indicators |
|---|---|---|---|---|
| 30 days | 150 stars | 300 | 600 | Show HN >= 50 points; 5 pkg.go.dev importers; 3 external issues |
| 90 days | 400 | 800 | 1,500 | Golang Weekly mention; 15 importers; 2 outside contributors; 1 podcast |
| 180 days | 800 | 1,500 | 3,000 | awesome-go merged; 40 importers; 1 named production user; v1.0 shipped |

Track: stars (GitHub), "Imported by" on pkg.go.dev (Go has no download
counts), Discussions per month, median time-to-first-response on issues (goal
under 48 hours), and referrals in GitHub traffic. Review monthly; if 90-day
base is missed, the fix is more writing, not more features.

## 8. Risks and mitigations

| Risk | Likelihood | Mitigation |
|---|---|---|
| LangChain ships an official Go LangGraph (today only `langsmith-go` and a Go reference page exist; no runtime) | Medium | Position on size and readability, not parity; an official port will be large by construction. Keep the mapping table current so migration in either direction is easy. |
| Scope-creep pressure (reducers, retries, bindings) | High | The decision framework below, published in CONTRIBUTING; a "declined with reasons" Discussion thread; the CI line-count check as a hard stop. |
| "Toy" perception | High early | Production user case study by month 5; benchmarks; fuzzing; the approval example on a real store; more test code than library code, stated with numbers. |
| Maintainer bandwidth (solo) | High | Small API is finished-able; monthly cadence, not weekly; auto-reply issue templates; recruit one co-maintainer from the first good contributors by month 4. |
| Google ADK Go and Eino absorb attention | Medium | They are frameworks with model bindings; MiniGraph is the runtime under a 40-line loop. Cite the Zep "no framework" essay and say MiniGraph is what you write after the 40 lines. |
| Line-count claim drifts (smolagents' failure mode) | Low | Already CI-enforced; publish the rule as tinygrad does. |
| Go generics ergonomics deter newcomers | Low | Examples never show a type parameter beyond `New[State]()`. |

## 9. Decision framework for feature requests

Apply in order; the first "yes" decides.

1. **Can it be built outside with the current API?** If yes, decline, and
   add an example or a doc snippet showing how (this is the answer to
   retries, backoff, token streaming, observability, history).
2. **Is it on the out-of-scope list in `AGENTS.md`** (per-key reducers,
   retry policies, token streaming, LLM bindings, platform layer)? Decline
   with a link to the design section of the README.
3. **Does it change a yield convention or resume semantics?** Decline unless
   it fixes a documented discrepancy; those three rows are the product.
4. **Does it add a dependency?** Decline.
5. **Does it push the library over the line budget** (450 lines with
   comments) or reduce coverage below 95%? Decline or require an equivalent
   removal.
6. **Otherwise**: accept only with a test in the matching `*_test.go`, a doc
   comment in the existing explanatory style, and README counts updated in
   the same PR.

Bug reports about documented behavior that is wrong (as with Interrupt inside
`Parallel`) skip the framework: fix or re-document, add the test, release a
patch.
