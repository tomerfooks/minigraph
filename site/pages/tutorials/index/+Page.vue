<script setup lang="ts">
import { tutorials } from '../list'
</script>

<template>
  <div class="prose">
    <div class="eyebrow">Tutorials</div>
    <h1>The patterns people build with LangGraph, in Go</h1>
    <p class="lede">
      Each tutorial is a runnable <code>main.go</code> in the repository, shown on the page exactly as CI compiles it.
      No API keys: models are scripted so you can read the control flow. Every one ends with where the real model or
      store plugs in.
    </p>
    <div class="grid" style="margin-top:2rem">
      <a v-for="t in tutorials" :key="t.slug" class="card link" :href="`/tutorials/${t.slug}`">
        <div class="tut-meta"><span class="pill go">{{ t.pattern }}</span></div>
        <h3>{{ t.title }}</h3>
        <p>{{ t.promise }}</p>
      </a>
    </div>
  </div>
</template>
