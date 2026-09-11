# MiniGraph site

Landing and tutorial site for [MiniGraph](https://github.com/tomerfooks/minigraph). Vike + Vue, prerendered to static HTML.

```sh
npm install
npm run dev      # http://localhost:5173
npm run build    # static output in dist/client — deploy that folder anywhere
```

Tutorial pages import `../examples/*/main.go?raw`, so the code shown is the code CI compiles. Console output on each page is pasted from a real run; re-run the example when you change it.

Design tokens: `styles/global.css`. Intent and rules: `../docs/pack/design-language.md`, `../docs/pack/voice-and-terms.md`.
