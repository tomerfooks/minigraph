<script setup lang="ts">
import Code from '../../components/Code.vue'
const outside = `// Retry: wrap the node.
func retry[S any](n int, node minigraph.Node[S]) minigraph.Node[S] {
    return func(ctx context.Context, s S) (S, error) {
        var err error
        for i := 0; i < n; i++ {
            var out S
            if out, err = node(ctx, s); err == nil {
                return out, nil
            }
        }
        return s, err
    }
}

// Timeout: wrap the context.
func within[S any](d time.Duration, node minigraph.Node[S]) minigraph.Node[S] {
    return func(ctx context.Context, s S) (S, error) {
        ctx, cancel := context.WithTimeout(ctx, d)
        defer cancel()
        return node(ctx, s)
    }
}

// Logging, metrics, tracing: same shape.`
</script>

<template>
  <article class="prose">
    <div class="eyebrow">Why</div>
    <h1>The size is the product</h1>
    <p class="lede">A runtime you can hold in your head is a different kind of dependency from one you trust because it is popular.</p>

    <h2>Four files, one engine</h2>
    <div class="table-wrap"><table>
      <thead><tr><th>File</th><th>Lines</th><th>What it buys</th></tr></thead>
      <tbody>
        <tr><td><code>graph.go</code></td><td>239</td><td>Builder, <code>Compile</code> validation, immutable <code>App</code>, and <code>StreamFrom</code>: the only execution loop. Every public method funnels into it.</td></tr>
        <tr><td><code>checkpoint.go</code></td><td>93</td><td><code>Checkpointer</code> (two methods), <code>MemorySaver</code>, and the thread wrappers that save every good step.</td></tr>
        <tr><td><code>parallel.go</code></td><td>57</td><td>Fan-out/join as a single node: goroutines, first-error cancel, ordered results, your merge.</td></tr>
        <tr><td><code>interrupt.go</code></td><td>31</td><td>The pause-for-a-human error type and its doc comment.</td></tr>
      </tbody>
    </table></div>
    <p>420 lines with the comments in. CI fails if the README's number drifts from <code>wc -l</code>. There is more test code than library code.</p>

    <h2>What the smallness gets you</h2>
    <h3>You can audit it before lunch</h3>
    <p>The whole runtime, including the resume semantics, is a coffee's worth of reading. When something surprises you in production, you open <code>graph.go:161</code> and read the loop instead of a discussion thread. Your dependency review is the file, not the changelog.</p>
    <h3>Zero dependencies, zero supply chain</h3>
    <p><code>go.mod</code> is the module line and <code>go 1.24</code>. Nothing to pin, nothing to audit, nothing to upgrade on a CVE. The compiled ReAct example is a static binary under 2 MB that starts in milliseconds; the same shape in Python pulls dozens of packages before your code runs.</p>
    <h3>State the compiler checks</h3>
    <p>State is any Go type. A node returns the whole next state, so a typo in a field name is a compile error, not a runtime <code>KeyError</code>, and there are no reducers: the two paths in LangGraph that need them, parallel writes and appends, are a <code>merge</code> you write and ordinary slice code.</p>
    <h3>One place to look for control flow</h3>
    <p>Every node has exactly one outgoing edge; a static edge is a router that ignores the state. All routing goes through <code>App.route</code>. Wiring mistakes are reported at <code>Compile</code>, joined, all at once.</p>
    <h3>Every step is a checkpoint</h3>
    <p>A run is a <code>(node, state)</code> pair advancing. That pair is what <code>Stream</code> yields, what a <code>Checkpointer</code> saves, and what <code>InvokeFrom</code> accepts. Interrupts, durability, and retries are not features layered on the engine; they are consequences of the engine having one loop.</p>
    <h3>It fits inside the program you already have</h3>
    <p>No agent server, no sidecar, no platform. A handler in your service ranges over <code>Stream</code> and writes events. A CLI ships as one binary. A cron job runs a thread and exits.</p>

    <h2>What the smallness costs you</h2>
    <p>Honesty is part of the pitch. These are real gaps, each with the compose-from-outside answer.</p>
    <div class="table-wrap"><table>
      <thead><tr><th>Not in the library</th><th>How you get it</th></tr></thead>
      <tbody>
        <tr><td>Per-key reducers</td><td>A node returns the whole state; <code>Parallel</code> gives you one <code>merge</code>.</td></tr>
        <tr><td>Retry and timeout policies</td><td>Wrap the node (below). Durable threads already retry failed steps on rerun.</td></tr>
        <tr><td>Token streaming</td><td>Inside the node, with your client. The engine streams steps.</td></tr>
        <tr><td>Provider bindings</td><td>A model call is a function call inside a node. <a href="/tutorials/llm">Tutorial.</a></td></tr>
        <tr><td>Checkpoint history and time travel</td><td>Keep the Steps you care about; resume any of them with <code>InvokeFrom</code>. Store history by appending in your <code>Checkpointer</code>.</td></tr>
        <tr><td>SQL/Redis savers</td><td>Two methods. <a href="/tutorials/durable">The file-backed one is 20 lines.</a></td></tr>
        <tr><td>Interrupts inside parallel branches</td><td>Put approval nodes in the outer graph after the parallel step.</td></tr>
        <tr><td>A platform, a studio, a tracing UI</td><td>Your existing logging and tracing, in the node or in the range loop.</td></tr>
      </tbody>
    </table></div>
    <Code lang="go" label="compose from outside" :code="outside" />

    <h2>The test for a feature request</h2>
    <p>Can a user do it with a wrapper around a <code>Node</code>, a <code>Router</code>, or a <code>Checkpointer</code>? Then it is not going in. Does it change what a yielded <code>Step</code> means? Then it is a resume-semantics change and needs a test that pins the three yield conventions. Does it add a dependency? No.</p>
    <p>That rule is written down in the repository's <code>AGENTS.md</code> so that contributors, human or otherwise, hold the line the same way.</p>
  </article>
</template>
