<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import Code from '../../../components/Code.vue'
import source from '../../../../examples/agent/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('agent')
const output = `── agent
   {Question:what do add(2,3) and mul(4,5) make together? Pending:[add 2 3 mul 4 5] Results:[] Answer:}
── tools
   {Question:what do add(2,3) and mul(4,5) make together? Pending:[] Results:[5 20] Answer:}
── agent
   {Question:what do add(2,3) and mul(4,5) make together? Pending:[] Results:[5 20] Answer:tool results [5 20] combine to 25}`
</script>

<template>
  <Tutorial title="Your first loop" pattern="agent ⇄ tools" :primitives="['AddNode', 'AddRouter', 'Stream']"
    slug="agent" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">The shape every other tutorial builds on: a model node, a tools node, and a router that decides whether to loop or stop.</p>
    </template>
    <template #story>
      <p>
        Start with the state. It is a plain struct: the question, the tool calls the agent still wants, the results gathered so
        far, and the answer. No <code>TypedDict</code>, no <code>Annotated</code>, no reducer. A node receives the whole state
        and returns the whole next state, so the compiler checks every field you touch.
      </p>
      <p>
        The <code>agent</code> node stands in for a model. With no results yet it queues two tool calls; with results it
        writes the answer. The <code>tools</code> node drains the queue. The router after <code>agent</code> is the only
        decision in the graph: answer present, go to <code>End</code>; otherwise go to <code>tools</code>, which always
        returns to <code>agent</code>. That return edge is the cycle, and the cycle is what makes this an agent instead of a
        pipeline.
      </p>
    </template>
    <template #graph>
      <Graph :w="460" :h="190"
        :nodes="[{id:'agent',x:80,y:70},{id:'tools',x:250,y:70},{id:'end',x:410,y:70,kind:'end'}]"
        :edges="[{from:'agent',to:'tools',label:'pending calls'},{from:'tools',to:'agent',label:'results',bend:55},{from:'agent',to:'end',label:'answer',bend:-70}]" />
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> <code>Stream</code> yields after every node, so the loop in <code>main</code> prints the state as it changes. Each yielded <code>Step</code> is also a checkpoint you could resume from. You get observability and resumability from the same range loop.</p>
      </div>
      <div class="callout go">
        <p><strong>Runaway protection.</strong> A router that never returns <code>End</code> stops after <code>app.MaxSteps</code> (default 25) with <code>ErrMaxSteps</code>. Raise it per app; it counts from zero on every run and every resume.</p>
      </div>
      <h2>Make it real</h2>
      <p>Replace the body of <code>agent</code> with a model call that returns either tool calls or a final answer. The router does not change:</p>
      <Code lang="go" code="g.AddNode(&quot;agent&quot;, func(ctx context.Context, s State) (State, error) {
    reply, err := client.Complete(ctx, prompt(s)) // any client, any provider
    if err != nil {
        return s, err // the run stops; resume retries this node
    }
    s.Pending, s.Answer = parse(reply)
    return s, nil
})" />
    </template>
  </Tutorial>
</template>
