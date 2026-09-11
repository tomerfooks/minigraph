<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import Code from '../../../components/Code.vue'
import source from '../../../../examples/durable/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('durable')
const output = `fetch:   loading items for #7
charge:  card charged
crash:   node "deliver": carrier API: connection reset
         checkpoint kept; run again to resume
exit status 1

$ go run ./examples/durable
checkpoint found at /tmp/minigraph-order-7.json — resuming
deliver: shipped [keyboard mouse]
done:    charged=true shipped=true`
</script>

<template>
  <Tutorial title="Crash, run again, resume" pattern="durable execution" :primitives="['Checkpointer', 'InvokeThread', 'Step']"
    slug="durable" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">The whole durability story in one file: a Checkpointer that writes JSON, a pipeline that dies mid-run, and a second process that finishes it.</p>
    </template>
    <template #story>
      <p>
        LangGraph's durable execution needs a checkpointer backend, usually Postgres or SQLite, plus a thread ID, plus a
        serializer for your state. MiniGraph's <code>Checkpointer</code> is two methods, <code>Save</code> and
        <code>Load</code>, over a <code>Step</code>. A <code>Step</code> is the node that just ran and the state it produced,
        and it is also the point to resume from. Here the backend is a file per thread.
      </p>
      <p>
        The pipeline fetches an order, charges the card, and delivers. The first run of a thread has no checkpoint, so the
        program wires <code>deliver</code> to fail with a network error. <code>InvokeThread</code> has already saved
        <code>fetch</code> and <code>charge</code>. Failed steps are not saved, so when you run the program again it loads the
        <code>charge</code> checkpoint and routes onward: straight into <code>deliver</code>. The card is not charged twice.
      </p>
    </template>
    <template #graph>
      <Graph :w="560" :h="120"
        :nodes="[{id:'fetch',x:60,y:50},{id:'charge',x:220,y:50},{id:'deliver',x:380,y:50},{id:'end',x:510,y:50,kind:'end'}]"
        :edges="[{from:'fetch',to:'charge'},{from:'charge',to:'deliver'},{from:'deliver',to:'end'}]" />
      <p>A straight line. The interesting part is what <code>StreamThread</code> saves and does not save:</p>
      <div class="table-wrap"><table>
        <thead><tr><th>Node result</th><th>Step saved</th><th>Rerun does</th></tr></thead>
        <tbody>
          <tr><td>success</td><td>that node + its new state</td><td>continues onward</td></tr>
          <tr><td>plain error</td><td>nothing new; last good step stays</td><td>retries the failed node</td></tr>
          <tr><td><code>Interrupt</code></td><td>the interrupted node + its state</td><td>routes onward; node does not re-run</td></tr>
        </tbody>
      </table></div>
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> The checkpointer is 20 lines and there is no schema: <code>encoding/json</code> serializes your state struct. That is the compile-time-typed state paying off a second time. Swap the file for a database row and nothing else changes.</p>
      </div>
      <div class="callout go">
        <p><strong>At-least-once, by design.</strong> If the process dies between a node returning and <code>Save</code> completing, that node runs again on the next run. Make side-effecting nodes idempotent (an order ID, an idempotency key) exactly as you would in any queue consumer.</p>
      </div>
      <h2>Make it real</h2>
      <p>Replace <code>FileSaver</code> with anything that can store a blob per key. The interface is the whole contract:</p>
      <Code lang="go" code="type Checkpointer[S any] interface {
    Save(ctx context.Context, thread string, step Step[S]) error
    Load(ctx context.Context, thread string) (Step[S], bool, error)
}" />
      <p>For a table, <code>INSERT … ON CONFLICT (thread) DO UPDATE</code> on a <code>jsonb</code> column is enough. Keep a history by appending instead of overwriting; resume from any row with <code>InvokeFrom</code>.</p>
    </template>
  </Tutorial>
</template>
