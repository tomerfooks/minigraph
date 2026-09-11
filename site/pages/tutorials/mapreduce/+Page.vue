<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import Code from '../../../components/Code.vue'
import source from '../../../../examples/mapreduce/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('mapreduce')
const output = `5 chunks summarized in 200ms (each takes 200ms)
§1: Go was designed… (15 words)
§2: The language is… (12 words)
§3: Goroutines and channels… (11 words)
§4: The standard library… (13 words)
§5: Generics arrived in… (11 words)`
</script>

<template>
  <Tutorial title="Dynamic fan-out with Parallel" pattern="map-reduce" :primitives="['Parallel', 'merge', 'Node']"
    slug="mapreduce" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">The number of branches is decided at run time from the state. Five chunks, five goroutines, one step, results in order.</p>
    </template>
    <template #story>
      <p>
        LangGraph does this with <code>Send</code> and a reducer on the results key. MiniGraph does it with the fact
        that <code>Parallel</code> returns a <code>Node</code>: build it inside a node, after you know how many chunks
        there are, and call it. Each branch receives a shallow copy of the state, summarizes its own chunk, and returns
        a fresh one-element <code>Summaries</code>. The <code>merge</code> you write folds the results, which arrive
        in branch order regardless of which goroutine finished first.
      </p>
      <p>
        Every branch sleeps 200ms to stand in for a model call. Five of them finish in 200ms, not a second: that is the
        goroutine story in one printed line.
      </p>
    </template>
    <template #graph>
      <Graph :w="600" :h="220"
        :nodes="[{id:'split',x:60,y:110},{id:'§1',x:260,y:30},{id:'§2',x:260,y:70},{id:'§3',x:260,y:110},{id:'§4',x:260,y:150},{id:'§5',x:260,y:190},{id:'reduce',x:460,y:110},{id:'end',x:560,y:110,kind:'end'}]"
        :edges="[{from:'split',to:'§1'},{from:'split',to:'§2'},{from:'split',to:'§3'},{from:'split',to:'§4'},{from:'split',to:'§5'},{from:'§1',to:'reduce'},{from:'§2',to:'reduce'},{from:'§3',to:'reduce'},{from:'§4',to:'reduce'},{from:'§5',to:'reduce'},{from:'reduce',to:'end'}]" />
      <p>The five branches are one node called <code>map</code>; the graph itself stays <code>split → map → reduce</code>.</p>
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> Branches share reference fields (the <code>Chunks</code> slice) and must treat them as read-only. Each branch writes a <em>new</em> slice into its own copy. The merge is the only place where results meet, so there is no concurrent write and nothing for the race detector to find.</p>
      </div>
      <div class="callout go">
        <p><strong>First error wins.</strong> If one branch fails, the sibling context is cancelled and the node returns that error with the branch index. Check <code>ctx.Done()</code> inside long branches, as <code>summarize</code> does, so cancellation is prompt.</p>
      </div>
      <h2>Make it real</h2>
      <p>Bound the concurrency when the fan-out is large: wrap the branch in a semaphore.</p>
      <Code lang="go" code="sem := make(chan struct{}, 8) // at most 8 model calls in flight
branch := func(ctx context.Context, s State) (State, error) {
    select {
    case sem <- struct{}{}:
        defer func() { <-sem }()
    case <-ctx.Done():
        return s, ctx.Err()
    }
    return summarize(i)(ctx, s)
}" />
    </template>
  </Tutorial>
</template>
