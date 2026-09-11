# MiniGraph positioning pack

Written September 2026. Star counts, LOC figures, and release dates are as measured then; re-check before quoting publicly.

| File | What it is |
|---|---|
| [architecture.md](architecture.md) | How the engine works: the single loop, Step as checkpoint, the three yield conventions, traces, what is composed from outside. |
| [code-review.md](code-review.md) | Honest senior-engineer review: strengths, findings, edge cases, coverage, ranked suggestions with line cost. |
| [features.md](features.md) | Every public symbol with its guarantee and pinning test; patterns you get for free; things you build outside. |
| [alternatives.md](alternatives.md) | LangGraph (Py/JS), Eino, langchaingo, Genkit Go, ADK Go, Temporal, Go ports; feature matrix; honest gaps. |
| [competitive-advantages.md](competitive-advantages.md) | Differentiators with evidence, trade-offs, positioning one-liners and pitches. |
| [target-audience.md](target-audience.md) | Six personas, where they are, what they search, objections and answers, "not for". |
| [strategy.md](strategy.md) | Growth plan grounded in comparable projects: launch, content, community, trust signals, goals, risks. |
| [use-cases.md](use-cases.md) | What people build with LangGraph, where it hurts, and the demos/tutorials that answer each. |
| [design-language.md](design-language.md) | Colour, type, layout, components, diagram rules for the site. |
| [voice-and-terms.md](voice-and-terms.md) | How we talk, banned words, the glossary, pushback answers. |

Open item surfaced by two independent reviews: an `Interrupt` returned inside a `Parallel` branch is documented as "treated as a plain error" but actually pauses the run at the parallel node (see code-review.md, finding 2). Fix is about four lines in `parallel.go`; not applied here because it changes the library line count.
