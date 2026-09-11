<script setup lang="ts">
// Inline SVG graph diagram. nodes: {id,x,y,kind?}, edges: {from,to,label?,curve?}
defineProps<{
  w: number; h: number
  nodes: { id: string; x: number; y: number; kind?: 'end' | 'human' | 'start' }[]
  edges: { from: string; to: string; label?: string; bend?: number }[]
}>()
</script>

<template>
  <svg class="diagram" :viewBox="`0 0 ${w} ${h}`" :width="w" :height="h" role="img">
    <defs>
      <marker id="arr" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">
        <path d="M0,0 L10,5 L0,10 z" fill="var(--muted)" />
      </marker>
    </defs>
    <template v-for="e in edges" :key="e.from + e.to + e.label">
      <path
        v-if="true"
        class="edge"
        marker-end="url(#arr)"
        :d="(() => {
          const a = nodes.find(n => n.id === e.from)!, b = nodes.find(n => n.id === e.to)!
          const dx = b.x - a.x, dy = b.y - a.y, len = Math.hypot(dx, dy) || 1
          const ux = dx / len, uy = dy / len, r = 22
          const sx = a.x + ux * r, sy = a.y + uy * r, ex = b.x - ux * r, ey = b.y - uy * r
          const bend = e.bend ?? 0
          const mx = (sx + ex) / 2 - uy * bend, my = (sy + ey) / 2 + ux * bend
          return `M${sx},${sy} Q${mx},${my} ${ex},${ey}`
        })()"
      />
      <text
        v-if="e.label"
        class="lab"
        text-anchor="middle"
        :x="(() => { const a = nodes.find(n => n.id === e.from)!, b = nodes.find(n => n.id === e.to)!; const dx=b.x-a.x, dy=b.y-a.y, l=Math.hypot(dx,dy)||1; return (a.x+b.x)/2 - (dy/l)*((e.bend??0)*0.6+14) })()"
        :y="(() => { const a = nodes.find(n => n.id === e.from)!, b = nodes.find(n => n.id === e.to)!; const dx=b.x-a.x, dy=b.y-a.y, l=Math.hypot(dx,dy)||1; return (a.y+b.y)/2 + (dx/l)*((e.bend??0)*0.6+14) + 4 })()"
      >{{ e.label }}</text>
    </template>
    <template v-for="n in nodes" :key="n.id">
      <rect v-if="n.kind === 'human'" class="node human" :x="n.x - 40" :y="n.y - 16" width="80" height="32" rx="6" />
      <circle v-else :class="['node', n.kind]" :cx="n.x" :cy="n.y" :r="n.kind === 'end' ? 12 : 16" />
      <text :x="n.x" :y="n.y + 34" text-anchor="middle">{{ n.id }}</text>
    </template>
  </svg>
</template>
