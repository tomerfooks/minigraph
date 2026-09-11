<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import source from '../../../../examples/crag/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('crag')
const output = `retrieve "restrict a type parameter" → 0 docs
grade    → relevant: false
rewrite  → "type constraint"
retrieve "type constraint" → 1 docs
grade    → relevant: true
answer   → Based on the docs: A type constraint is an interface that restricts which types may be used as a type argument.`
</script>

<template>
  <Tutorial title="Retrieve, grade, rewrite" pattern="corrective RAG" :primitives="['Router', 'cycle', 'Stream']"
    slug="crag" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">Plain RAG trusts the first retrieval. Corrective RAG grades it, and rewrites the query when the documents miss.</p>
    </template>
    <template #story>
      <p>
        Four nodes. <code>retrieve</code> fills <code>Docs</code> from an index (here a keyword match over three
        sentences). <code>grade</code> sets <code>Relevant</code>. The router after <code>grade</code> goes to
        <code>generate</code> when the docs are relevant or the rewrite budget is spent, and to <code>rewrite</code>
        otherwise. <code>rewrite</code> changes <code>Query</code> and loops back to <code>retrieve</code>.
      </p>
      <p>
        The first query, "restrict a type parameter", misses. The rewritten query, "type constraint", hits, and the
        answer is grounded in the document that came back. The fallback path is visible in the router: after two rewrites
        it generates anyway, and <code>generate</code> says so.
      </p>
    </template>
    <template #graph>
      <Graph :w="600" :h="200"
        :nodes="[{id:'retrieve',x:70,y:70},{id:'grade',x:230,y:70},{id:'generate',x:400,y:70},{id:'end',x:550,y:70,kind:'end'},{id:'rewrite',x:150,y:160}]"
        :edges="[{from:'retrieve',to:'grade'},{from:'grade',to:'generate',label:'relevant'},{from:'generate',to:'end'},{from:'grade',to:'rewrite',label:'miss',bend:-20},{from:'rewrite',to:'retrieve',bend:-20}]" />
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> <code>Docs</code> is reset at the top of <code>retrieve</code>. A node returns the whole next state, so "replace" and "append" are both ordinary Go on a slice: no reducer to declare, no <code>operator.add</code> surprise when a second retrieval doubles the list.</p>
      </div>
      <h2>Make it real</h2>
      <p>Point <code>retrieve</code> at a vector store or a full-text index, have <code>grade</code> ask a small model "is this relevant, yes or no" per document, and let <code>rewrite</code> ask for a better search query. The router and the budget stay.</p>
    </template>
  </Tutorial>
</template>
