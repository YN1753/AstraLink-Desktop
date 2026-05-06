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

// 全局 Prism
import Prism from 'prismjs'
window.Prism = Prism
import 'prismjs/themes/prism-tomorrow.css'
import 'prismjs/components/prism-go'
import 'prismjs/components/prism-javascript'

const props = defineProps(['modelValue'])
const emit = defineEmits(['update:modelValue'])

let tooltipProvider = null
let slashProvider = null
let tooltipEl = null
let slashEl = null
let pollTimer = null
let lastSelection = null
let lastSelectedText = ''

function createTooltipElement() {
  const el = document.createElement('div')
  el.className = 'tooltip'
  el.innerHTML = `
    <button data-action="bold" title="加粗 (Ctrl+B)"><b>B</b></button>
    <button data-action="italic" title="斜体 (Ctrl+I)"><i>I</i></button>
    <button data-action="strike" title="删除线"><s>S</s></button>
    <button data-action="code" title="行内代码"><code>&lt;/&gt;</code></button>
    <button data-action="link" title="链接">🔗</button>
  `
  el.style.position = 'absolute'
  el.style.zIndex = '1000'
  el.dataset.show = 'false'
  el.addEventListener('mousedown', (e) => {
    const btn = e.target.closest('button[data-action]')
    if (btn) { e.preventDefault(); executeCommand(btn.dataset.action) }
  })
  return el
}

function createSlashElement() {
  const el = document.createElement('div')
  el.className = 'slash-dropdown'
  el.style.position = 'absolute'
  el.style.zIndex = '1000'
  el.dataset.show = 'false'

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
    { icon: '📊', label: '表格', action: 'table' },
  ]

  el.innerHTML = items.map(item => `
    <div class="slash-dropdown-item" data-action="${item.action}">
      <span class="icon">${item.icon}</span>
      <span class="label">${item.label}</span>
    </div>
  `).join('')

  el.addEventListener('mousedown', (e) => {
    const item = e.target.closest('.slash-dropdown-item')
    if (item) { e.preventDefault(); executeSlashCommand(item.dataset.action) }
  })
  return el
}

function getEditorView() {
  const el = document.querySelector('.ProseMirror')
  if (!el) return null
  // Milkdown 暴露 view 的方式
  return el.pmViewDesc?.view || null
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
      case 'bold':
        toggleMark(schema.marks.strong)
        break
      case 'italic':
        toggleMark(schema.marks.emphasis || schema.marks.em)
        break
      case 'strike':
        toggleMark(schema.marks.strike_through || schema.marks.strikeThrough)
        break
      case 'code':
        toggleMark(schema.marks.code_inline || schema.marks.inlineCode || schema.marks.code)
        break
      case 'link':
        toggleMark(schema.marks.link, { href: 'https://example.com' })
        break
    }

    if (tr.docChanged) {
      dispatch(tr)
      view.focus()
      lastSelection = null
      lastSelectedText = ''
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

    // 删除 '/' 字符
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
            lastSelectedText = view.state.doc.textBetween(from, to)
          } else {
            lastSelection = null
            lastSelectedText = ''
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
  <div class="milkdown-pro-wrapper">
    <Milkdown />
  </div>
</template>

<style scoped>
.milkdown-pro-wrapper { width: 100%; min-height: 500px; text-align: left; }

:deep(.milkdown) { background: transparent !important; }
:deep(.ProseMirror) {
  outline: none !important;
  color: var(--text-primary);
  line-height: 1.8;
  font-size: 17px;
  min-height: 600px;
  padding-bottom: 200px;
  white-space: pre-wrap;
}

:deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  color: var(--text-secondary);
  opacity: 0.4;
  pointer-events: none;
  float: left;
  height: 0;
}

:deep(.ProseMirror h1) {
  font-weight: 200; color: var(--text-primary); border-bottom: 1px solid var(--accent);
  padding-bottom: 10px; margin: 30px 0;
}
:deep(.ProseMirror h2), :deep(.ProseMirror h3), :deep(.ProseMirror h4) {
  color: var(--text-primary);
}

:deep(.ProseMirror a) { color: var(--accent); text-decoration: none; }
:deep(.ProseMirror a:hover) { text-decoration: underline; }

:deep(.milkdown .code-fence) {
  background: var(--bg-app) !important;
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  padding: 1.5rem;
  margin: 20px 0;
}
</style>

<style>
.tooltip {
  display: none;
  gap: 2px;
  padding: 4px 6px;
  background: rgba(20, 20, 20, 0.95);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.6);
}
.tooltip[data-show="true"] { display: flex; }

.tooltip button {
  display: flex; align-items: center; justify-content: center;
  min-width: 28px; height: 28px; padding: 0 6px;
  border: none; border-radius: 4px; background: transparent;
  color: #94a3b8; cursor: pointer; transition: all 0.15s;
  font-size: 13px; font-family: inherit;
}
.tooltip button:hover { background: rgba(255,255,255,0.12); color: #fff; }

.slash-dropdown {
  display: none; min-width: 200px; max-height: 260px;
  overflow-y: auto; padding: 4px;
  background: rgba(20, 20, 20, 0.95);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.6);
}
.slash-dropdown[data-show="true"] { display: block; }

.slash-dropdown-item {
  display: flex; align-items: center; gap: 10px;
  padding: 7px 10px; border-radius: 6px;
  cursor: pointer; transition: background 0.12s; color: #94a3b8;
}
.slash-dropdown-item:hover { background: rgba(255,255,255,0.08); color: #fff; }

.slash-dropdown-item .icon {
  width: 26px; height: 26px; display: flex; align-items: center; justify-content: center;
  border-radius: 5px; background: rgba(59, 130, 246, 0.15);
  color: #3b82f6; font-size: 11px; font-weight: 700; flex-shrink: 0;
}
.slash-dropdown-item .label { font-size: 13px; font-weight: 500; }
</style>
