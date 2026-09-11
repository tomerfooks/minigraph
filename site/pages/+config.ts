import type { Config } from 'vike/types'
import vikeVue from 'vike-vue/config'

export default {
  extends: [vikeVue],
  prerender: true,
  title: 'MiniGraph — LangGraph for Go, 1.2% of the code',
  description:
    'MiniGraph is a 420-line, zero-dependency graph engine for Go agents: typed state, cycles, checkpoints, interrupts, parallel fan-out. You can read all of it.',
  lang: 'en',
} satisfies Config
