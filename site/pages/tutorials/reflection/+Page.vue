<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import source from '../../../../examples/reflection/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('reflection')
const output = `round 1 draft:    Go is a programming language.
round 1 critique: too vague — be specific about what makes Go different
round 2 draft:    Go is a compiled, statically typed language with goroutines for concurrency.
round 2 critique: mention deployment: single static binary
round 3 draft:    Go is a compiled, statically typed language with goroutines for concurrency and a single-binary deploy story.
approved after 3 rounds: Go is a compiled, statically typed language with goroutines for concurrency and a single-binary deploy story.`
</script>

<template>
  <Tutorial title="Generate, critique, revise" pattern="reflection" :primitives="['Router', 'cycle']"
    slug="reflection" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">Self-correction is a writer and a critic taking turns until the critic has nothing to say or the round budget runs out.</p>
    </template>
    <template #story>
      <p>
        <code>generate</code> drafts, using the previous critique if there is one. <code>critique</code> writes feedback
        into the state, or the empty string for "approved". The router after <code>critique</code> ends the run when the
        critique is empty or <code>Round</code> hits the budget; otherwise it loops to <code>generate</code>. Two nodes, one
        router, one field that doubles as the stop signal.
      </p>
      <p>
        LangGraph's reflection tutorial does the same with a message list and a length check. Here the budget is an
        integer in the state and the approval is a typed field, so the stop condition is one line you can read.
      </p>
    </template>
    <template #graph>
      <Graph :w="460" :h="190"
        :nodes="[{id:'generate',x:80,y:70},{id:'critique',x:250,y:70},{id:'end',x:410,y:70,kind:'end'}]"
        :edges="[{from:'generate',to:'critique'},{from:'critique',to:'generate',label:'feedback',bend:55},{from:'critique',to:'end',label:'approved or budget'}]" />
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> The round counter is incremented by <code>generate</code>, not by the engine. MiniGraph does not count for you beyond <code>MaxSteps</code>; anything the graph needs to remember goes in the state, where it is visible and checkpointed.</p>
      </div>
      <h2>Make it real</h2>
      <p>Use two prompts, or two models: a cheap generator and a stricter critic. Ask the critic for a structured verdict (<code>{"approved": true}</code> or a list of issues) and unmarshal it, so "approved" is a boolean rather than the absence of text.</p>
    </template>
  </Tutorial>
</template>
