<script setup lang="ts">
import Code from './Code.vue'
// One tutorial page: header, story, graph, real source, real output, notes.
defineProps<{
  title: string
  pattern: string
  primitives: string[]
  slug: string
  source: string
  output: string
  prev?: { slug: string; title: string }
  next?: { slug: string; title: string }
}>()
</script>

<template>
  <article class="prose">
    <header class="tut-head">
      <div class="tut-meta">
        <span class="pill go">{{ pattern }}</span>
        <span v-for="p in primitives" :key="p" class="pill">{{ p }}</span>
        <span class="pill amber">offline · no API key</span>
      </div>
      <h1>{{ title }}</h1>
      <slot name="lede" />
    </header>

    <slot name="story" />

    <h2>The graph</h2>
    <slot name="graph" />

    <h2>Run it</h2>
    <Code lang="console" :code="`$ go run ./examples/${slug}\n${output}`" />

    <h2>The code</h2>
    <p>This is <code>examples/{{ slug }}/main.go</code> in the repository, imported at build time. CI compiles it.</p>
    <Code lang="go" :label="`examples/${slug}/main.go`" :code="source" />

    <slot name="notes" />

    <nav class="next">
      <a v-if="prev" :href="`/tutorials/${prev.slug}`">← {{ prev.title }}</a><span v-else></span>
      <a v-if="next" :href="`/tutorials/${next.slug}`">{{ next.title }} →</a>
      <a v-else href="/tutorials">All tutorials →</a>
    </nav>
  </article>
</template>
