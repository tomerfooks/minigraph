<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import Code from '../../../components/Code.vue'
import source from '../../../../examples/supervisor/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('supervisor')
const output = `supervisor → researcher
researcher: gathered notes
supervisor → coder
coder: wrote Map
supervisor → writer
writer: assembled report
supervisor → FINISH

Goal: write a generic Map helper
Notes: Go generics (1.18+) allow func Map[T, U any](xs []T, f func(T) U) []U
Code: func Map[T, U any](xs []T, f func(T) U) []U { out := make([]U, len(xs)); for i, x := range xs { out[i] = f(x) }; return out }`
</script>

<template>
  <Tutorial title="Supervisor and workers" pattern="multi-agent" :primitives="['Router', 'cycle', 'Invoke']"
    slug="supervisor" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">LangGraph's most-searched multi-agent tutorial is one decision node, three worker nodes, and a router that reads the decision.</p>
    </template>
    <template #story>
      <p>
        A supervisor looks at the state and names who should work next. Workers do their job and hand control back. In
        LangGraph this is the <em>supervisor</em> graph with <code>Command(goto=…)</code>; here the supervisor writes its
        decision into <code>s.Next</code> and the router returns it. The name of the next node comes from the state, which
        is the entire trick.
      </p>
      <p>
        Every worker has a static edge back to <code>supervisor</code>, so the graph is a star with a loop through the
        centre. When the supervisor says <code>FINISH</code> the router returns <code>End</code>. The scripted policy
        here fills gaps in order; a model would do the same from a prompt that lists the worker names.
      </p>
    </template>
    <template #graph>
      <Graph :w="520" :h="260"
        :nodes="[{id:'supervisor',x:110,y:130},{id:'researcher',x:340,y:40},{id:'coder',x:360,y:130},{id:'writer',x:340,y:220},{id:'end',x:110,y:240,kind:'end'}]"
        :edges="[{from:'supervisor',to:'researcher',bend:-18},{from:'researcher',to:'supervisor',bend:-18},{from:'supervisor',to:'coder',bend:-14},{from:'coder',to:'supervisor',bend:-14},{from:'supervisor',to:'writer',bend:-18},{from:'writer',to:'supervisor',bend:-18},{from:'supervisor',to:'end',label:'FINISH'}]" />
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> A router can only send the run to a node that exists, and it is checked at run time against the compiled node set: a supervisor that hallucinates <code>"designer"</code> produces <code>routed to unknown node</code>, not a silent no-op. Validate the model's choice in the supervisor node if you want a friendlier error.</p>
      </div>
      <div class="callout go">
        <p><strong>Workers as subgraphs.</strong> A worker can be a whole compiled graph: <code>AddNode("coder", coderApp.Invoke)</code>. The supervisor does not know or care. That is hierarchical multi-agent with no extra API.</p>
      </div>
      <h2>Make it real</h2>
      <Code lang="go" code="func supervise(ctx context.Context, s State) (State, error) {
    choice, err := llm(ctx, &quot;Workers: researcher, coder, writer. State: &quot;+describe(s)+&quot;\nWho next, or FINISH?&quot;)
    if err != nil {
        return s, err
    }
    s.Next = strings.TrimSpace(choice)
    return s, nil
}" />
    </template>
  </Tutorial>
</template>
