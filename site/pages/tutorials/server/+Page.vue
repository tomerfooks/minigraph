<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import Code from '../../../components/Code.vue'
import source from '../../../../examples/server/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('server')
const output = `2026/09/11 15:40:02 listening on :8080 — try: curl -N 'localhost:8080/run?q=twice+the+population+of+France'

$ curl -N 'localhost:8080/run?q=twice+the+population+of+France'
event: think
data: {"Node":"think","State":{"Question":"twice the population of France","Steps":["thinking about: twice the population of France"],"Answer":""}}

event: lookup
data: {"Node":"lookup","State":{"Question":"twice the population of France","Steps":["thinking about: twice the population of France","lookup: population of France = 68,000,000"],"Answer":""}}

event: answer
data: {"Node":"answer","State":{"Question":"twice the population of France","Steps":["thinking about: twice the population of France","lookup: population of France = 68,000,000"],"Answer":"136,000,000"}}

event: done
data: {}`
</script>

<template>
  <Tutorial title="An agent behind an HTTP endpoint" pattern="net/http + SSE" :primitives="['Stream', 'context', 'Step']"
    slug="server" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">Embed the graph in the service you already have. Each Step becomes one server-sent event; a client that disconnects cancels the run.</p>
    </template>
    <template #story>
      <p>
        This is the deployment story most teams actually want: not a separate agent server, but a handler in the Go
        service that owns the data. The handler ranges over <code>app.Stream(r.Context(), …)</code>, JSON-encodes each
        <code>Step</code>, and flushes it as an SSE event named after the node. The whole thing is one static binary
        with no runtime to install next to it.
      </p>
      <p>
        The request context does the cancellation. When the client goes away, <code>r.Context()</code> is cancelled,
        the engine checks it before routing to the next node, and the handler returns. Run the second curl with a short
        timeout and watch the server stop after the first event.
      </p>
    </template>
    <template #graph>
      <Graph :w="560" :h="150"
        :nodes="[{id:'think',x:70,y:60},{id:'lookup',x:250,y:60},{id:'answer',x:420,y:60},{id:'end',x:520,y:60,kind:'end'}]"
        :edges="[{from:'think',to:'lookup',label:'mentions France'},{from:'think',to:'answer',bend:-50},{from:'lookup',to:'answer'},{from:'answer',to:'end'}]" />
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> <code>Step[S]</code> marshals with <code>encoding/json</code> because <code>S</code> is your struct. The SSE payload, the checkpoint format, and the type your handler code sees are the same thing.</p>
      </div>
      <div class="callout go">
        <p><strong>Cancellation is at node boundaries.</strong> The engine checks <code>ctx.Err()</code> before each route. A long model call inside a node should pass <code>ctx</code> to its HTTP client so it aborts promptly too.</p>
      </div>
      <h2>Make it real</h2>
      <p>Add a thread ID from the request and switch to <code>StreamThread</code> with a database-backed <code>Checkpointer</code>: a client that reconnects gets the run continued, not restarted.</p>
      <Code lang="go" code="thread := r.URL.Query().Get(&quot;thread&quot;)
for step, err := range app.StreamThread(r.Context(), saver, thread, initial) { … }" />
    </template>
  </Tutorial>
</template>
