import vue from '@vitejs/plugin-vue'
import vike from 'vike/plugin'
import type { UserConfig } from 'vite'

export default {
  plugins: [vike(), vue()],
  // Tutorial pages import ../examples/*/main.go?raw so the code shown is the real code.
  server: { fs: { allow: ['..'] } },
} satisfies UserConfig
