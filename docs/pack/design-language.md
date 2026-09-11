# MiniGraph design language

The site sells one idea: **you can read all of it**. Everything visual should feel like a well-set book about code — calm paper, ink, one accent, and monospace where the product itself appears. Not a SaaS dashboard, not a neon "AI" gradient.

Source of truth: `site/styles/global.css` (tokens) and this file (intent).

## Principles

1. **Reading room, not launch pad.** Generous line length (max 760px for prose), serif headlines, sans body, real paragraphs. No hero video, no floating cards, no glassmorphism.
2. **The code is the product.** Code blocks are the most "designed" element on every page: dark, high-contrast, always the real source (tutorial pages import `examples/*/main.go` at build time). Never show pseudo-code where real code exists.
3. **One accent, used for meaning.** Go cyan marks nodes, links, and the active state. Amber marks the human (interrupts, approvals) and the number that matters ("1.2%"). Nothing else gets colour.
4. **Numbers are claims.** "420 lines", "1.2%", "0 deps" appear as set stats, always with the thing they are measured against. If a number changes in the README it changes here (CI enforces the README; keep the site in step).
5. **Dark and light are equals.** Both themes come from the same tokens; the code block is dark in both so screenshots look identical.

## Colour

| Token | Light | Dark | Use |
|---|---|---|---|
| `--paper` | `#FAF8F3` | `#0F1115` | page background |
| `--paper-2` | `#F1EEE6` | `#161920` | inline code, pills, callouts |
| `--ink` | `#16171B` | `#E8E6E1` | body text, headings |
| `--ink-2` | `#4A4C55` | `#B9B6AE` | secondary text, nav |
| `--muted` | `#6E6A63` | `#9AA0A8` | captions, edges in diagrams |
| `--line` | `#E4E0D8` | `#262A33` | borders, rules |
| `--go` | `#00ADD8` | same | nodes, active nav, primary highlights (the Go brand cyan, also in the logo) |
| `--go-deep` | `#007D9C` | same | links, italic emphasis in headlines |
| `--amber` | `#E9A23B` | same | the human in a diagram, the key number, callout rule |
| `--code-bg` | `#14161B` | `#1A1D24` | code and terminal blocks |

Code syntax colours (`--code-kw` cyan-ish, `--code-str` amber-ish, `--code-fn` green-ish, `--code-cmt` grey italic) are deliberately low-saturation so keywords do not shout.

Contrast: body text on paper is ≥ 12:1 in both themes; links (`--go-deep` on `--paper`) ≈ 4.6:1; never put `--go` text on white for body copy.

## Type

- **Headlines:** IBM Plex Serif 600. `h1` clamps 2.2–3.6rem. Italic + `--go-deep` is allowed for exactly one phrase per headline (e.g. *1.2%*).
- **Body:** IBM Plex Sans 400/500, 17px, line-height 1.6.
- **Code, labels, eyebrows, stats:** IBM Plex Mono. Eyebrows are uppercase, 0.8rem, letter-spaced 0.08em, in `--go-deep`.
- One family (Plex) so serif/sans/mono share x-height and feel like one voice.
- Fallbacks are system fonts; the page must read fine if Google Fonts is blocked.

## Layout

- Max content width 1080px; prose column 760px; 20px side gutters at every width.
- Hero is a two-column grid (copy | code) that collapses under 860px.
- Sections are separated by whitespace, not background bands. At most one `hr` per page.
- Grid of cards: `auto-fit, minmax(240px, 1fr)`. Cards have 1px `--line` borders and no shadows.
- Radius is 10px everywhere (`--radius`); pills are fully round.

## Components

| Component | Where | Notes |
|---|---|---|
| `Code.astro` | every page | Astro's built-in Shiki `<Code>` (theme `vitesse-dark`) with the background overridden to `--code-bg`. `label` prop shows the file name top-right. `lang="console"` renders a terminal. |
| `Graph.astro` | tutorials, home | Inline SVG: cyan filled circles are nodes, hollow cyan circle is End, amber rounded rectangle is the human. Edges are grey with arrowheads; `bend` curves a return edge. Labels in mono. |
| `.stat` | home, why | Big serif number + small caption underneath. |
| `.pill` | tutorial headers | Mono tags: pattern name, primitives used, "offline". `go` and `amber` variants. |
| `.callout` | tutorials | Amber left rule = "what to notice"; `go` variant = a tip. |
| `.steps` | tutorials | Numbered walkthrough; numbers are cyan discs. |
| `.bar` | why, vs page | Horizontal LOC comparison bar; MiniGraph's bar is cyan, everything else grey. |

## Diagrams

- Nodes are named exactly as in the code (`retrieve`, `grade`, not "Retriever"). Lower-case, mono.
- Left-to-right flow. Cycles curve underneath (`bend` > 0). Start is implied; End is the hollow circle.
- A human is always the amber rectangle labelled `human`.
- No icons, no clip art, no gradients.

## Imagery

- Only the logo (`docs/logo.svg`: three cyan nodes fanning out and joining, one hollow end node, a faint return loop) and screenshots of terminals.
- The social preview (`site/public/og.png`) is the headline on paper with the logo, nothing more.

## Motion

None beyond hover underline on links and border colour on cards. The site should feel like a printed page that happens to be live.

## Don'ts

- No emoji in UI copy.
- No gradients, glows, or blur.
- No "AI" iconography (sparkles, brains, robots).
- No stock illustrations.
- No marketing superlatives in headings (see `voice-and-terms.md`).
