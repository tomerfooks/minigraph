<script setup lang="ts">
import { computed } from 'vue'
// Tiny Go/shell highlighter: enough for docs, zero dependencies.
const props = defineProps<{ code: string; lang?: string; label?: string }>()
const esc = (s: string) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
const KW = /\b(package|import|func|return|type|struct|interface|var|const|if|else|for|range|switch|case|default|go|defer|chan|select|map|nil|true|false|break|continue|any|error|string|int|int64|float64|bool|byte|context)\b/g
const html = computed(() => {
  const src = esc(props.code.trim())
  if (props.lang === 'sh' || props.lang === 'console') {
    return src.replace(/^(\$ )(.*)$/gm, '<span class="prompt">$1</span>$2').replace(/^(#.*)$/gm, '<span class="cm">$1</span>')
  }
  const out: string[] = []
  const re = /(\/\/.*$)|(`[^`]*`|"(?:[^"\\]|\\.)*")|([A-Za-z_][A-Za-z0-9_]*)(?=\()/gm
  let last = 0
  let m: RegExpExecArray | null
  const kw = (s: string) => s.replace(KW, '<span class="kw">$1</span>')
  while ((m = re.exec(src))) {
    out.push(kw(src.slice(last, m.index)))
    if (m[1]) out.push(`<span class="cm">${m[1]}</span>`)
    else if (m[2]) out.push(`<span class="st">${m[2]}</span>`)
    else if (m[3]) out.push(KW.test(m[3] + ' ') ? kw(m[3]) : `<span class="fn">${m[3]}</span>`)
    last = re.lastIndex
  }
  out.push(kw(src.slice(last)))
  return out.join('')
})
</script>

<template>
  <div class="code-wrap">
    <span v-if="label" class="label">{{ label }}</span>
    <pre :class="lang === 'console' ? 'term' : 'code'"><code v-html="html"></code></pre>
  </div>
</template>
