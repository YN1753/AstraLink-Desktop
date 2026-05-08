<script setup>
import { onMounted, onUnmounted } from 'vue'
import { Milkdown, useEditor } from '@milkdown/vue'
import { Editor, rootCtx, defaultValueCtx, editorViewCtx } from '@milkdown/core'
import { commonmark } from '@milkdown/preset-commonmark'
import { gfm } from '@milkdown/preset-gfm'
import { history } from '@milkdown/plugin-history'
import { prism } from '@milkdown/plugin-prism'
import { indent } from '@milkdown/plugin-indent'
import { listener, listenerCtx } from '@milkdown/plugin-listener'
import { clipboard } from '@milkdown/plugin-clipboard'
import { TooltipProvider } from '@milkdown/plugin-tooltip'
import { SlashProvider } from '@milkdown/plugin-slash'
import { nord } from '@milkdown/theme-nord'

import Prism from 'prismjs'
window.Prism = Prism
import 'prismjs/themes/prism-tomorrow.css'
import 'prismjs/components/prism-go'
import 'prismjs/components/prism-javascript'
import 'prismjs/components/prism-typescript'
import 'prismjs/components/prism-python'
import 'prismjs/components/prism-java'
import 'prismjs/components/prism-c'
import 'prismjs/components/prism-cpp'
import 'prismjs/components/prism-rust'
import 'prismjs/components/prism-sql'
import 'prismjs/components/prism-bash'
import 'prismjs/components/prism-json'
import 'prismjs/components/prism-yaml'
import 'prismjs/components/prism-markdown'

const props = defineProps(['modelValue'])
const emit = defineEmits(['update:modelValue'])

let tooltipProvider = null
let slashProvider = null
let tooltipEl = null
let slashEl = null
let pollTimer = null
let lastSelection = null

function createTooltipElement() {
  const el = document.createElement('div')
  el.className = 'md-tooltip'
  el.innerHTML = `
    <button data-action="bold" title="加粗 (Ctrl+B)"><b>B</b></button>
    <button data-action="italic" title="斜体 (Ctrl+I)"><i>I</i></button>
    <button data-action="strike" title="删除线"><s>S</s></button>
    <button data-action="code" title="行内代码"><code>&lt;/&gt;</code></button>
    <button data-action="link" title="链接">🔗</button>
  `
  el.style.cssText = 'position:absolute;z-index:9999;display:none;gap:2px;padding:4px 6px;background:rgba(20,20,20,0.95);backdrop-filter:blur(20px);border:1px solid var(--glass-border);border-radius:8px;box-shadow:0 8px 24px rgba(0,0,0,0.6);'
  el.addEventListener('mousedown', (e) => {
    const btn = e.target.closest('button[data-action]')
    if (btn) { e.preventDefault(); executeCommand(btn.dataset.action) }
  })
  return el
}

function createSlashElement() {
  const el = document.createElement('div')
  el.className = 'md-slash-dropdown'
  el.style.cssText = 'position:absolute;z-index:9999;display:none;min-width:200px;max-height:280px;overflow-y:auto;padding:6px;background:rgba(20,20,20,0.95);backdrop-filter:blur(20px);border:1px solid var(--glass-border);border-radius:10px;box-shadow:0 12px 32px rgba(0,0,0,0.6);'

  const items = [
    { icon: 'H1', label: '一级标题', action: 'h1' },
    { icon: 'H2', label: '二级标题', action: 'h2' },
    { icon: 'H3', label: '三级标题', action: 'h3' },
    { icon: '—', label: '分割线', action: 'hr' },
    { icon: '•', label: '无序列表', action: 'bulletList' },
    { icon: '1.', label: '有序列表', action: 'orderedList' },
    { icon: '☐', label: '任务列表', action: 'taskList' },
    { icon: '>', label: '引用', action: 'blockquote' },
    { icon: '{}', label: '代码块', action: 'codeBlock' },
    { icon: '▦', label: '表格', action: 'table' },
  ]

  el.innerHTML = items.map(item => `
    <div class="slash-item" data-action="${item.action}" style="display:flex;align-items:center;gap:10px;padding:8px 12px;border-radius:6px;cursor:pointer;color:var(--text-secondary);transition:all 0.12s;">
      <span class="slash-icon" style="width:28px;height:28px;display:flex;align-items:center;justify-content:center;border-radius:6px;background:rgba(var(--accent-rgb),0.15);color:var(--accent);font-size:11px;font-weight:700;">${item.icon}</span>
      <span style="font-size:13px;font-weight:500;">${item.label}</span>
    </div>
  `).join('')

  el.addEventListener('mousedown', (e) => {
    const item = e.target.closest('.slash-item')
    if (item) { e.preventDefault(); executeSlashCommand(item.dataset.action) }
  })
  return el
}

function executeCommand(action) {
  if (!lastSelection) return
  const editor = get()
  if (!editor) return

  editor.action((ctx) => {
    const view = ctx.get(editorViewCtx)
    if (!view) return
    const { state, dispatch } = view
    const { from, to } = lastSelection
    const schema = state.schema
    let tr = state.tr

    function toggleMark(markType, attrs) {
      if (!markType) return
      const hasMark = state.doc.rangeHasMark(from, to, markType)
      if (hasMark) {
        tr = tr.removeMark(from, to, markType)
      } else {
        tr = tr.addMark(from, to, markType.create(attrs))
      }
    }

    switch (action) {
      case 'bold': toggleMark(schema.marks.strong); break
      case 'italic': toggleMark(schema.marks.emphasis || schema.marks.em); break
      case 'strike': toggleMark(schema.marks.strike_through || schema.marks.strikeThrough); break
      case 'code': toggleMark(schema.marks.code_inline || schema.marks.inlineCode || schema.marks.code); break
      case 'link': toggleMark(schema.marks.link, { href: 'https://example.com' }); break
    }

    if (tr.docChanged) {
      dispatch(tr)
      view.focus()
      lastSelection = null
    }
  })
}

function executeSlashCommand(action) {
  const editor = get()
  if (!editor) return

  editor.action((ctx) => {
    const view = ctx.get(editorViewCtx)
    if (!view) return
    const { state, dispatch } = view
    const { $from } = state.selection
    const schema = state.schema

    const slashPos = $from.pos - 1
    let tr = state.tr
    if (state.doc.textBetween(slashPos, $from.pos) === '/') {
      tr = tr.delete(slashPos, $from.pos)
    }

    const $newFrom = tr.doc.resolve(slashPos)
    const paraPos = $newFrom.before()
    const paraNode = tr.doc.nodeAt(paraPos)

    if (!paraNode || paraNode.type.name !== 'paragraph') {
      dispatch(tr)
      view.focus()
      if (slashProvider) slashProvider.hide()
      return
    }

    const textContent = paraNode.content

    switch (action) {
      case 'h1':
      case 'h2':
      case 'h3': {
        const level = action === 'h1' ? 1 : action === 'h2' ? 2 : 3
        const headingType = schema.nodes.heading
        if (headingType) {
          const heading = headingType.create({ level }, textContent)
          tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, heading)
        }
        break
      }
      case 'hr': {
        const hrType = schema.nodes.horizontalRule || schema.nodes.horizontal_rule
        if (hrType) {
          const hr = hrType.create()
          tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, hr)
        }
        break
      }
      case 'bulletList': {
        const listType = schema.nodes.bullet_list
        const itemType = schema.nodes.list_item
        if (listType && itemType) {
          const para = schema.nodes.paragraph.create(null, textContent)
          const item = itemType.create(null, para)
          const list = listType.create(null, [item])
          tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, list)
        }
        break
      }
      case 'orderedList': {
        const listType = schema.nodes.ordered_list
        const itemType = schema.nodes.list_item
        if (listType && itemType) {
          const para = schema.nodes.paragraph.create(null, textContent)
          const item = itemType.create(null, para)
          const list = listType.create(null, [item])
          tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, list)
        }
        break
      }
      case 'taskList': {
        const listType = schema.nodes.bullet_list
        const itemType = schema.nodes.task_list_item || schema.nodes.taskListItem
        if (listType && itemType) {
          const para = schema.nodes.paragraph.create(null, textContent)
          const item = itemType.create({ checked: false }, para)
          const list = listType.create(null, [item])
          tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, list)
        }
        break
      }
      case 'blockquote': {
        const bqType = schema.nodes.blockquote
        if (bqType) {
          const para = schema.nodes.paragraph.create(null, textContent)
          const bq = bqType.create(null, para)
          tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, bq)
        }
        break
      }
      case 'codeBlock': {
        const codeType = schema.nodes.code_block
        if (codeType) {
          const code = codeType.create(null, schema.text(''))
          tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, code)
        }
        break
      }
      case 'table': {
        const tableType = schema.nodes.table
        const rowType = schema.nodes.table_row
        const cellType = schema.nodes.table_cell || schema.nodes.tableCell
        if (tableType && rowType && cellType) {
          const cell1 = cellType.create(null, schema.nodes.paragraph.create(null, schema.text('列1')))
          const cell2 = cellType.create(null, schema.nodes.paragraph.create(null, schema.text('列2')))
          const cell3 = cellType.create(null, schema.nodes.paragraph.create(null, schema.text('列3')))
          const row = rowType.create(null, [cell1, cell2, cell3])
          const table = tableType.create(null, [row])
          tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, table)
        }
        break
      }
    }

    dispatch(tr)
    view.focus()
    if (slashProvider) slashProvider.hide()
  })
}

const { get } = useEditor((root) => {
  return Editor.make()
      .config((ctx) => {
        ctx.set(rootCtx, root)
        ctx.set(defaultValueCtx, props.modelValue || "")
        ctx.get(listenerCtx).markdownUpdated((_, markdown) => {
          emit('update:modelValue', markdown)
        })
      })
      .config(nord)
      .use(commonmark)
      .use(gfm)
      .use(history)
      .use(prism)
      .use(indent)
      .use(listener)
      .use(clipboard)
}, [])

onMounted(() => {
  tooltipEl = createTooltipElement()
  slashEl = createSlashElement()
  document.body.appendChild(tooltipEl)
  document.body.appendChild(slashEl)

  tooltipProvider = new TooltipProvider({ content: tooltipEl, debounce: 50 })
  slashProvider = new SlashProvider({ content: slashEl, debounce: 0, offset: 8 })

  pollTimer = setInterval(() => {
    const editor = get()
    if (!editor) return
    editor.action((ctx) => {
      try {
        const view = ctx.get(editorViewCtx)
        if (view && tooltipProvider) {
          tooltipProvider.update(view)
          const { from, to } = view.state.selection
          if (from !== to) {
            lastSelection = { from, to }
          } else {
            lastSelection = null
          }
        }
        if (view && slashProvider) {
          const { $from } = view.state.selection
          const lineTextBefore = $from.parent.textBetween(0, $from.parentOffset)
          const isSlash = lineTextBefore === '/' || lineTextBefore.endsWith(' /')
          if (isSlash) {
            try {
              const coords = view.coordsAtPos($from.pos)
              slashProvider.show({
                getBoundingClientRect() {
                  return {
                    width: 0, height: 0,
                    x: coords.left, y: coords.top,
                    top: coords.top, left: coords.left,
                    right: coords.left, bottom: coords.top,
                  }
                }
              })
            } catch (e) {
              slashProvider.hide()
            }
          } else {
            slashProvider.hide()
          }
        }
      } catch (e) {}
    })
  }, 100)
})

onUnmounted(() => {
  if (pollTimer) clearInterval(pollTimer)
  if (tooltipProvider) tooltipProvider.destroy()
  if (slashProvider) slashProvider.destroy()
  if (tooltipEl?.parentNode) tooltipEl.parentNode.removeChild(tooltipEl)
  if (slashEl?.parentNode) slashEl.parentNode.removeChild(slashEl)
})
</script>

<template>
  <div class="milkdown-wrapper">
    <Milkdown />
  </div>
</template>

<style scoped>
.milkdown-wrapper {
  width: 100%;
  height: 100%;
  text-align: left;
}

/* Base editor styles */
:deep(.milkdown) {
  background: transparent !important;
  color: var(--text-primary);
}

:deep(.ProseMirror) {
  outline: none !important;
  color: var(--text-primary);
  line-height: 1.85;
  font-size: clamp(15px, 2vw, 17px);
  min-height: 100%;
  padding: clamp(16px, 3vw, 28px) clamp(16px, 4%, 48px);
  white-space: pre-wrap;
  word-wrap: break-word;
}

/* Placeholder */
:deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  color: var(--text-secondary);
  opacity: 0.45;
  pointer-events: none;
  float: left;
  height: 0;
}

/* Headings */
:deep(.ProseMirror h1) {
  font-size: clamp(24px, 4vw, 32px);
  font-weight: 700;
  color: var(--text-primary);
  border-bottom: 2px solid var(--accent);
  padding-bottom: 12px;
  margin: clamp(24px, 4vw, 36px) 0 clamp(16px, 3vw, 24px);
  line-height: 1.3;
}

:deep(.ProseMirror h2) {
  font-size: clamp(20px, 3vw, 26px);
  font-weight: 600;
  color: var(--text-primary);
  margin: clamp(20px, 3vw, 30px) 0 clamp(12px, 2vw, 18px);
  line-height: 1.4;
}

:deep(.ProseMirror h3) {
  font-size: clamp(17px, 2.5vw, 22px);
  font-weight: 600;
  color: var(--text-primary);
  margin: clamp(16px, 3vw, 24px) 0 clamp(10px, 2vw, 14px);
  line-height: 1.5;
}

:deep(.ProseMirror h4),
:deep(.ProseMirror h5),
:deep(.ProseMirror h6) {
  font-size: clamp(15px, 2vw, 18px);
  font-weight: 600;
  color: var(--text-secondary);
  margin: clamp(14px, 2vw, 20px) 0 clamp(8px, 1.5vw, 12px);
}

/* Paragraphs */
:deep(.ProseMirror p) {
  margin: clamp(8px, 1.5vw, 12px) 0;
}

/* Links */
:deep(.ProseMirror a) {
  color: var(--accent);
  text-decoration: none;
  border-bottom: 1px solid transparent;
  transition: border-color 0.2s;
}

:deep(.ProseMirror a:hover) {
  border-bottom-color: var(--accent);
}

/* Strong and emphasis */
:deep(.ProseMirror strong) {
  font-weight: 700;
  color: var(--text-primary);
}

:deep(.ProseMirror em) {
  font-style: italic;
  color: var(--text-secondary);
}

/* Strikethrough */
:deep(.ProseMirror s) {
  text-decoration: line-through;
  opacity: 0.7;
}

/* Inline code */
:deep(.ProseMirror code) {
  background: rgba(var(--accent-rgb), 0.12);
  color: var(--accent);
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.9em;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

/* Blockquotes */
:deep(.ProseMirror blockquote) {
  border-left: 3px solid var(--accent);
  margin: clamp(12px, 2vw, 18px) 0;
  padding: clamp(10px, 2vw, 16px) clamp(16px, 3vw, 24px);
  background: rgba(var(--accent-rgb), 0.06);
  border-radius: 0 8px 8px 0;
  color: var(--text-secondary);
  font-style: italic;
}

:deep(.ProseMirror blockquote p) {
  margin: 0;
}

/* Horizontal rule */
:deep(.ProseMirror hr) {
  border: none;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--glass-border), transparent);
  margin: clamp(20px, 4vw, 32px) 0;
}

/* Lists */
:deep(.ProseMirror ul),
:deep(.ProseMirror ol) {
  padding-left: clamp(18px, 3vw, 28px);
  margin: clamp(10px, 2vw, 16px) 0;
}

:deep(.ProseMirror li) {
  margin: clamp(4px, 0.8vw, 6px) 0;
}

:deep(.ProseMirror li p) {
  margin: 0;
}

:deep(.ProseMirror ul) {
  list-style-type: disc;
}

:deep(.ProseMirror ol) {
  list-style-type: decimal;
}

/* Task lists */
:deep(.ProseMirror ul[data-type="taskList"]) {
  list-style: none;
  padding-left: 0;
}

:deep(.ProseMirror ul[data-type="taskList"] li) {
  display: flex;
  align-items: flex-start;
  gap: 10px;
}

:deep(.ProseMirror ul[data-type="taskList"] li > label) {
  flex-shrink: 0;
  margin-top: 3px;
}

:deep(.ProseMirror ul[data-type="taskList"] li > div) {
  flex: 1;
}

:deep(.ProseMirror ul[data-type="taskList"] input[type="checkbox"]) {
  width: 16px;
  height: 16px;
  accent-color: var(--accent);
  cursor: pointer;
}

/* Code blocks */
:deep(.ProseMirror pre) {
  background: var(--glass-bg) !important;
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  padding: clamp(14px, 2.5vw, 20px);
  margin: clamp(14px, 2.5vw, 20px) 0;
  overflow-x: auto;
  position: relative;
}

:deep(.ProseMirror pre)::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, var(--accent), rgba(var(--accent-rgb), 0.3), transparent);
  border-radius: 12px 12px 0 0;
}

:deep(.ProseMirror pre code) {
  background: transparent;
  color: var(--text-primary);
  padding: 0;
  font-size: clamp(12px, 1.5vw, 14px);
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
  line-height: 1.6;
}

/* Tables */
:deep(.ProseMirror table) {
  border-collapse: collapse;
  width: 100%;
  margin: clamp(14px, 2.5vw, 20px) 0;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--glass-border);
}

:deep(.ProseMirror table th) {
  background: rgba(var(--accent-rgb), 0.1);
  color: var(--text-primary);
  font-weight: 600;
  padding: clamp(8px, 1.5vw, 12px) clamp(12px, 2vw, 16px);
  text-align: left;
  border-bottom: 1px solid var(--glass-border);
}

:deep(.ProseMirror table td) {
  padding: clamp(8px, 1.5vw, 12px) clamp(12px, 2vw, 16px);
  border-bottom: 1px solid var(--glass-border);
  color: var(--text-secondary);
}

:deep(.ProseMirror table tr:last-child td) {
  border-bottom: none;
}

:deep(.ProseMirror table tr:hover td) {
  background: rgba(255, 255, 255, 0.02);
}

/* Images */
:deep(.ProseMirror img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
  margin: clamp(10px, 2vw, 16px) 0;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
}

/* Selection */
:deep(.ProseMirror ::selection) {
  background: rgba(var(--accent-rgb), 0.3);
}

/* Tooltip styles */
:deep(.md-tooltip button) {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 30px;
  height: 28px;
  padding: 0 7px;
  border: none;
  border-radius: 5px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
  font-size: 13px;
  font-family: inherit;
}

:deep(.md-tooltip button:hover) {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

/* Slash dropdown hover */
:deep(.slash-item:hover) {
  background: rgba(var(--accent-rgb), 0.1) !important;
  color: var(--text-primary) !important;
}

/* Scrollbar for code blocks */
:deep(.ProseMirror pre)::-webkit-scrollbar {
  height: 6px;
}

:deep(.ProseMirror pre)::-webkit-scrollbar-track {
  background: transparent;
}

:deep(.ProseMirror pre)::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 3px;
}
</style>

<style>
/* Global tooltip and slash dropdown styles */
.md-tooltip[data-show="true"] {
  display: flex !important;
}

.md-slash-dropdown[data-show="true"] {
  display: block !important;
}
</style>