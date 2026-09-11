import { defineConfig } from 'astro/config'

export default defineConfig({
  // Tutorial pages import ../examples/*/main.go?raw so the code shown is the real code.
  vite: { server: { fs: { allow: ['..'] } } },
})
