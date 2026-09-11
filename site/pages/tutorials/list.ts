// Ordered tutorial index. Also drives prev/next links.
export const tutorials = [
  { slug: 'agent', pattern: 'agent ⇄ tools', title: 'Your first loop', promise: 'The minimal agent: one model node, one tools node, a router that decides when to stop.' },
  { slug: 'react', pattern: 'ReAct', title: 'ReAct: Thought, Action, Observation', promise: 'The canonical tool-using agent, with a swappable mock model and a trace you can print.' },
  { slug: 'supervisor', pattern: 'multi-agent', title: 'Supervisor and workers', promise: 'One node decides who works next; specialists hand control back. A router and a cycle.' },
  { slug: 'planexec', pattern: 'plan-and-execute', title: 'Plan, execute, replan', promise: 'A planner writes steps, an executor runs one per turn, a replanner decides when the objective is met.' },
  { slug: 'reflection', pattern: 'reflection', title: 'Generate, critique, revise', promise: 'A writer and a critic in a loop with a round budget. Self-correction as two nodes and a router.' },
  { slug: 'crag', pattern: 'corrective RAG', title: 'Retrieve, grade, rewrite', promise: 'Grade retrieved documents; if they miss, rewrite the query and try again before answering.' },
  { slug: 'mapreduce', pattern: 'map-reduce', title: 'Dynamic fan-out with Parallel', promise: 'Decide the number of branches at run time, run them concurrently, merge in order. LangGraph\'s Send, no reducer.' },
  { slug: 'approval', pattern: 'human-in-the-loop', title: 'Pause for approval', promise: 'A node returns an Interrupt; a human edits the state; the run continues where it stopped.' },
  { slug: 'durable', pattern: 'durable execution', title: 'Crash, run again, resume', promise: 'A Checkpointer that writes JSON files. Kill the process mid-run; the next run picks up at the last saved step.' },
  { slug: 'server', pattern: 'net/http + SSE', title: 'An agent behind an HTTP endpoint', promise: 'Stream every step to the browser as server-sent events; client disconnect cancels the run.' },
  { slug: 'llm', pattern: 'real model', title: 'Swap in a real model', promise: 'Mock by default, live with one env var: a model call is an HTTP request inside a node. No SDK required.' },
  { slug: 'fanout', pattern: 'fan-out / join', title: 'Parallel researchers, merged', promise: 'Three researchers run at once as one step; a merge you write folds their findings. Subgraphs come free.' },
]
export const nav = (slug: string) => {
  const i = tutorials.findIndex(t => t.slug === slug)
  return { prev: tutorials[i - 1], next: tutorials[i + 1] }
}
