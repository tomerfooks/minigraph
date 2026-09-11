<script setup lang="ts">
import { usePageContext } from 'vike-vue/usePageContext'
import '../styles/global.css'

const pc = usePageContext()
const nav = [
  { href: '/', label: 'Home' },
  { href: '/why', label: 'Why' },
  { href: '/vs-langgraph', label: 'vs LangGraph' },
  { href: '/tutorials', label: 'Tutorials' },
  { href: '/api', label: 'API' },
]
const active = (href: string) =>
  href === '/' ? pc.urlPathname === '/' : pc.urlPathname.startsWith(href)
</script>

<template>
  <div class="shell">
    <header class="top">
      <a href="/" class="brand">
        <img src="/logo.svg" alt="" width="28" height="28" />
        <span>MiniGraph</span>
      </a>
      <nav>
        <a v-for="n in nav" :key="n.href" :href="n.href" :class="{ active: active(n.href) }">{{ n.label }}</a>
        <a href="https://github.com/tomerfooks/minigraph" class="gh" rel="noopener">GitHub ↗</a>
      </nav>
    </header>
    <main>
      <slot />
    </main>
    <footer class="foot">
      <div>
        <strong>MiniGraph</strong> · LangGraph for Go, 1.2% of the code · MIT ©
        <a href="https://github.com/tomerfooks">Tomer Fooks</a>
      </div>
      <div class="foot-links">
        <a href="https://pkg.go.dev/github.com/tomerfooks/minigraph">pkg.go.dev</a>
        <a href="https://github.com/tomerfooks/minigraph">Source</a>
        <a href="https://github.com/tomerfooks/minigraph/issues">Issues</a>
      </div>
    </footer>
  </div>
</template>
