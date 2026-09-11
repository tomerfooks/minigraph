<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import Code from '../../../components/Code.vue'
import source from '../../../../examples/fanout/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('fanout')
const output = `── research
── write
3 sources agree: web says "solar sails" is promising; docs says "solar sails" is promising; db says "solar sails" is promising`
</script>

<template>
  <Tutorial title="Parallel researchers, merged" pattern="fan-out / join" :primitives="['Parallel', 'merge', 'subgraph']"
    slug="fanout" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">Three researchers run concurrently as one graph step. A merge you write folds their findings. A compiled graph is a node, so any researcher could be a whole subgraph.</p>
    </template>
    <template #story>
      <p>
        <code>Parallel(merge, branches...)</code> packages fan-out and join as a single <code>Node</code>. Each branch
        gets a shallow copy of the state and writes its finding into its own scalar field. When all branches are done,
        <code>merge</code> receives the base state and the results in branch order and produces the next state. The
        graph sees one step called <code>research</code>; <code>Stream</code> yields once for it.
      </p>
      <p>
        Because the merge is a function you write, it is the reducer, defined once, next to the fan-out, with the
        types the compiler already knows. There is no per-key annotation and no surprise when two branches touch the same
        field: you decide what happens in the merge.
      </p>
    </template>
    <template #graph>
      <Graph :w="560" :h="200"
        :nodes="[{id:'web',x:180,y:40},{id:'docs',x:180,y:100},{id:'db',x:180,y:160},{id:'merge',x:340,y:100},{id:'write',x:460,y:100},{id:'end',x:530,y:100,kind:'end'}]"
        :edges="[{from:'web',to:'merge'},{from:'docs',to:'merge'},{from:'db',to:'merge'},{from:'merge',to:'write'},{from:'write',to:'end'}]" />
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> The branches write <code>Finding</code>, a scalar, on their own copy. Writing to a shared slice from a branch would be a data race; the merge appends to <code>Findings</code> once the goroutines are done. Run it with <code>-race</code> and it stays quiet.</p>
      </div>
      <div class="callout go">
        <p><strong>Interrupts in branches.</strong> A branch cannot pause the run. Design approval steps as nodes in the outer graph, after the parallel step.</p>
      </div>
      <h2>Subgraphs, for free</h2>
      <p>An <code>App.Invoke</code> has a <code>Node</code>'s signature. So a whole compiled graph can be a branch, or a plain node:</p>
      <Code lang="go" code="deep, _ := minigraph.New[State]().
    AddNode(&quot;search&quot;, search).AddNode(&quot;rank&quot;, rank).
    AddEdge(minigraph.Start, &quot;search&quot;).AddEdge(&quot;search&quot;, &quot;rank&quot;).AddEdge(&quot;rank&quot;, minigraph.End).
    Compile()

research := minigraph.Parallel(merge, researcher(&quot;web&quot;), deep.Invoke) // a subgraph as a branch" />
      <p>Nested parallelism, hierarchical agents, reusable sub-flows: all the same trick.</p>
    </template>
  </Tutorial>
</template>
