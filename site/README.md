# MiniGraph site

Landing and tutorial site for [MiniGraph](https://github.com/tomerfooks/minigraph). Astro, fully static, no client-side JavaScript.

```sh
npm install
npm run dev      # http://localhost:4321
npm run build    # static output in dist/ — deploy that folder anywhere
```

Tutorial pages import `../examples/*/main.go?raw`, so the code shown is the code CI compiles. Console output on each page is pasted from a real run; re-run the example when you change it.

Layout: `src/layouts/Base.astro`. Tutorial shell: `src/components/Tutorial.astro`. Tutorial order and blurbs: `src/data/tutorials.ts`. Design tokens: `src/styles/global.css`. Intent and rules: `../docs/pack/design-language.md`, `../docs/pack/voice-and-terms.md`.
