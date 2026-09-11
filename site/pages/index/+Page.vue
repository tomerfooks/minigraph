<script setup lang="ts">
import Code from '../../components/Code.vue'
import Graph from '../../components/Graph.vue'

const sixty = `app, _ := minigraph.New[State]().
    AddNode("agent", callModel).
    AddNode("tools", runTools).
    AddEdge(minigraph.Start, "agent").
    AddRouter("agent", func(ctx context.Context, s State) (string, error) {
        if s.Done {
            return minigraph.End, nil
        }
        return "tools", nil // else: run tools, then loop back
    }).
    AddEdge("tools", "agent").
    Compile()

final, err := app.Invoke(ctx, State{Question: "..."})`

const react = `$ go run ./examples/react
Question: What is twice the population of France?
Thought: I need France's population before I can double it.
Action: lookup[population of France]
Observation: 68000000
Thought: Now I double the population I found.
Action: calculate[68000000 * 2]
Observation: 136000000
Final Answer: Twice the population of France is 136000000.`

const api = `minigraph.Start, minigraph.End          // reserved endpoints of every run
type Node[S any]   = func(ctx, S) (S, error)      // transforms the state
type Router[S any] = func(ctx, S) (string, error) // picks the next node

New[S]().AddNode(…).AddEdge(…).AddRouter(…).Compile() → (*App[S], error)

app.Invoke(ctx, state)                  // run to completion
app.Stream(ctx, state)                  // iter.Seq2[Step[S], error] — range over it
app.InvokeFrom(ctx, step)               // resume any yielded Step
app.StreamFrom(ctx, step)
app.InvokeThread(ctx, saver, id, state) // durable runs: load, run, save every step
app.StreamThread(ctx, saver, id, state)
app.MaxSteps                            // runaway-cycle guard, default 25

Parallel(merge, branches...)            // concurrent fan-out/join, as one Node
&Interrupt{Payload: …}                  // return from a node to pause for a human
Checkpointer[S] · MemorySaver[S]        // persistence interface + in-memory impl
Step[S]{Node, State}                    // one executed step — and a resume point`

const hitl = `final, err := app.Invoke(ctx, initial)
var intr *minigraph.Interrupt
if errors.As(err, &intr) {
    final.Approved = true // the human's answer, written into the state
    final, err = app.InvokeFrom(ctx, minigraph.Step[State]{Node: intr.Node, State: final})
}`

const durable = `saver := &minigraph.MemorySaver[State]{}
final, err := app.InvokeThread(ctx, saver, "thread-42", initial)
// crash anywhere, run it again: it continues from the last saved Step`

const fan = `research := minigraph.Parallel(mergeFindings, searchWeb, searchDocs, subgraph.Invoke)`
</script>

<template>
  <section class="hero">
    <div>
      <div class="eyebrow">Go · zero dependencies · MIT</div>
      <h1>LangGraph for Go,<br /><em>1.2%</em> of the code.</h1>
      <p class="lede">
        A graph engine for agents: nodes over a typed state, edges the state decides at run time, cycles as a first-class thing.
        420 lines with the comments in. You can read all of it.
      </p>
      <div class="btn-row">
        <a class="btn primary mono" href="https://github.com/tomerfooks/minigraph">go get github.com/tomerfooks/minigraph</a>
        <a class="btn" href="/tutorials">Read the tutorials</a>
      </div>
      <div class="stats">
        <div class="stat">420<small>lines, comments included</small></div>
        <div class="stat">0<small>dependencies</small></div>
        <div class="stat">1.2%<small>of LangGraph core + checkpoints</small></div>
      </div>
    </div>
    <Code lang="go" label="sixty seconds" :code="sixty" />
  </section>

  <section class="section narrow">
    <blockquote>
      Go is the best language for building web servers and agents — compiled, lean, simple. One static binary, no runtime zoo,
      code that still reads clearly at 3 a.m. Agents don't need a fancy framework. One small graph engine covers lean agents
      <em>and</em> dynamic workflows. MiniGraph is that engine — and nothing you didn't ask for.
    </blockquote>
  </section>

  <section class="section">
    <div class="two">
      <div>
        <div class="eyebrow">The shape</div>
        <h2>An agent is a graph that loops</h2>
        <p>
          Nodes transform a shared, typed state. Edges are static or decided by the state at run time. Cycles are the whole
          difference between a flowchart and an agent that keeps going until it is done.
        </p>
        <Graph :w="460" :h="190"
          :nodes="[{id:'agent',x:80,y:70},{id:'tools',x:250,y:70},{id:'end',x:410,y:70,kind:'end'}]"
          :edges="[{from:'agent',to:'tools',label:'needs tools'},{from:'tools',to:'agent',label:'observation',bend:55},{from:'agent',to:'end',label:'done',bend:-70}]" />
        <p>Here is a full ReAct loop running on it, offline, right now:</p>
      </div>
      <Code lang="console" :code="react" />
    </div>
  </section>

  <section class="section">
    <div class="eyebrow">Why so small</div>
    <h2>Same shape. Compiler-checked state. No machinery you rarely touch.</h2>
    <div class="two">
      <div>
        <p>
          LangGraph is excellent — and enormous. MiniGraph keeps the ideas that carry their weight: typed state, conditional
          edges, cycles, invoke and stream, checkpointers, interrupts, parallel fan-out. It drops reducers, retry policies,
          token streaming, provider bindings, and the platform layer. Those compose from outside, in your code, where you can
          see them.
        </p>
        <div class="bar"><span>LangGraph core + checkpoints</span><div class="track"><div class="fill" style="width:100%"></div></div><span>~33,700 loc</span></div>
        <div class="bar"><span>MiniGraph</span><div class="track"><div class="fill go" style="width:1.2%"></div></div><span>420 loc</span></div>
        <p style="margin-top:1rem"><a href="/why">Why the size is the product →</a></p>
      </div>
      <Code lang="go" label="the entire API" :code="api" />
    </div>
  </section>

  <section class="section">
    <div class="eyebrow">The patterns</div>
    <h2>Three things fall out of one fact: every step is a checkpoint</h2>
    <div class="grid">
      <div class="card">
        <h3>Human in the loop</h3>
        <p>A node pauses by returning an <code>Interrupt</code>. The state it returns is kept. Answer by editing that state and resuming.</p>
        <Code lang="go" :code="hitl" />
        <a href="/tutorials/approval">Tutorial →</a>
      </div>
      <div class="card">
        <h3>Durable threads</h3>
        <p>A <code>Checkpointer</code> persists the latest Step per thread. Crash anywhere, run it again, it continues. Failed steps are not saved, so a rerun retries them.</p>
        <Code lang="go" :code="durable" />
        <a href="/tutorials/durable">Tutorial →</a>
      </div>
      <div class="card">
        <h3>Parallel fan-out</h3>
        <p><code>Parallel</code> folds N concurrent branches into one node with a merge you write. A compiled graph's <code>Invoke</code> is a node, so subgraphs are branches too.</p>
        <Code lang="go" :code="fan" />
        <a href="/tutorials/mapreduce">Tutorial →</a>
      </div>
    </div>
  </section>

  <section class="section narrow">
    <div class="eyebrow">Coming from LangGraph</div>
    <h2>The mapping fits on a napkin</h2>
    <div class="table-wrap">
      <table>
        <thead><tr><th>LangGraph</th><th>MiniGraph</th></tr></thead>
        <tbody>
          <tr><td><code>StateGraph(State)</code></td><td><code>New[State]()</code> — any Go type, checked at compile time</td></tr>
          <tr><td>node function</td><td><code>func(ctx, S) (S, error)</code></td></tr>
          <tr><td><code>add_edge</code> / <code>add_conditional_edges</code></td><td><code>AddEdge(from, to)</code> / <code>AddRouter(from, router)</code></td></tr>
          <tr><td><code>compile()</code> → <code>invoke</code> / <code>stream</code></td><td><code>Compile()</code> → <code>Invoke</code> / <code>Stream</code></td></tr>
          <tr><td>recursion limit</td><td><code>App.MaxSteps</code></td></tr>
          <tr><td>checkpointer + <code>thread_id</code></td><td><code>Checkpointer</code>, <code>InvokeThread</code></td></tr>
          <tr><td><code>interrupt()</code></td><td>return <code>&amp;Interrupt{…}</code>; <code>InvokeFrom</code> to continue</td></tr>
          <tr><td><code>Send</code> / parallel supersteps</td><td><code>Parallel(merge, branches...)</code></td></tr>
          <tr><td>subgraphs</td><td>free — <code>App.Invoke</code> is a valid <code>Node</code></td></tr>
        </tbody>
      </table>
    </div>
    <p><a href="/vs-langgraph">The full comparison, including what MiniGraph does not do →</a></p>
  </section>

  <section class="section narrow">
    <h2>Twelve runnable examples, no API keys</h2>
    <p>Every tutorial on this site is a <code>main.go</code> in the repository that CI compiles. The code on the page is the code in the file.</p>
    <Code lang="sh" code="go run ./examples/react       # ReAct: Thought → Action → Observation
go run ./examples/approval    # human-in-the-loop on a durable thread
go run ./examples/durable     # crash, run again, resume from a JSON file
go run ./examples/mapreduce   # dynamic fan-out, one Parallel step
go run ./examples/server      # agent behind net/http, steps as SSE" />
    <div class="btn-row">
      <a class="btn primary" href="/tutorials">All tutorials</a>
      <a class="btn" href="https://github.com/tomerfooks/minigraph/tree/master/examples">Browse examples on GitHub</a>
    </div>
  </section>

  <section class="section narrow">
    <h2>FAQ</h2>
    <p><strong>Is it production-ready?</strong> It is 420 lines with more test code than library code. Read it over one coffee and you will know it better than most of your dependencies.</p>
    <p><strong>Where are the LLM bindings?</strong> There are none. A node is <code>func(ctx, S) (S, error)</code>. Call your model inside one: any client, any provider, no adapter layer. <a href="/tutorials/llm">See the real-model tutorial.</a></p>
    <p><strong>Why not just use LangGraph?</strong> If you are in Python, do. If you are in Go and want the whole runtime in your head, welcome home.</p>
  </section>
</template>
