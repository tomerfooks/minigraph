<script setup lang="ts">
import Tutorial from '../../../components/Tutorial.vue'
import Graph from '../../../components/Graph.vue'
import source from '../../../../examples/planexec/main.go?raw'
import { nav } from '../list'
const { prev, next } = nav('planexec')
const output = `Objective: How many seats do we need for the meetup?
Plan: 3 steps
  done: count the RSVPs for the meetup => 48
Replan: 2 left
  done: look up the usual no-show rate => 25%
Replan: 1 left
  done: multiply to get expected attendance => 36
Replan: 0 left
Answer: Expect about 36 people; book the 40-seat room.`
</script>

<template>
  <Tutorial title="Plan, execute, replan" pattern="plan-and-execute" :primitives="['Router', 'cycle', 'Stream']"
    slug="planexec" :source="source" :output="output" :prev="prev" :next="next">
    <template #lede>
      <p class="lede">Separate thinking about the whole task from doing one step of it. Three nodes and a loop that closes when the plan is empty.</p>
    </template>
    <template #story>
      <p>
        Plan-and-execute keeps a list of remaining steps in the state. <code>plan</code> fills it once. <code>execute</code>
        pops the head, runs it, and appends the result to <code>Done</code>. <code>replan</code> looks at the whole picture:
        here it declares the answer when nothing is left; a model could also append new steps when a result was
        surprising, and the loop keeps going.
      </p>
      <p>
        Because the plan lives in the state, every <code>Step</code> yielded by <code>Stream</code> carries the exact
        to-do list at that moment. Save one with a <code>Checkpointer</code> and a crash mid-plan resumes at the right step.
      </p>
    </template>
    <template #graph>
      <Graph :w="560" :h="180"
        :nodes="[{id:'plan',x:60,y:60},{id:'execute',x:230,y:60},{id:'replan',x:400,y:60},{id:'end',x:520,y:60,kind:'end'}]"
        :edges="[{from:'plan',to:'execute'},{from:'execute',to:'replan'},{from:'replan',to:'execute',label:'steps left',bend:60},{from:'replan',to:'end',label:'answer'}]" />
    </template>
    <template #notes>
      <div class="callout">
        <p><strong>What to notice.</strong> The main loop switches on <code>step.Node</code> to print different things per node. That is the whole "streaming" story: you get a typed Step per node and decide what to show. No event schema to learn.</p>
      </div>
      <div class="callout go">
        <p><strong>Budget the loop.</strong> A replanner that keeps adding steps is the classic runaway. <code>MaxSteps</code> caps node executions per run; set it to your worst-case plan length times three (plan, execute, replan).</p>
      </div>
      <h2>Make it real</h2>
      <p>Have <code>plan</code> and <code>replan</code> ask a model for a JSON list of steps and unmarshal it into <code>s.Plan</code>. Give <code>execute</code> a map of tools keyed by verb. Nothing about the wiring changes.</p>
    </template>
  </Tutorial>
</template>
