<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { Milkdown, useEditor } from '@milkdown/vue'
import { Editor, rootCtx, defaultValueCtx, editorViewCtx } from '@milkdown/core'
import { commonmark } from '@milkdown/preset-commonmark'
import { gfm } from '@milkdown/preset-gfm'
import { history } from '@milkdown/plugin-history'
import { prism } from '@milkdown/plugin-prism'
import { indent } from '@milkdown/plugin-indent'
import { listener, listenerCtx } from '@milkdown/plugin-listener'
import { clipboard } from '@milkdown/plugin-clipboard'
import { $prose } from '@milkdown/utils'
import { Plugin, PluginKey } from '@milkdown/prose/state'
import { Slice, Fragment } from '@milkdown/prose/model'
import { OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime'
import { TooltipProvider } from '@milkdown/plugin-tooltip'
import { SlashProvider } from '@milkdown/plugin-slash'
import { nord } from '@milkdown/theme-nord'
import { parserCtx, serializerCtx } from '@milkdown/core'
import { SearchNotes, GetNoteContent, SaveAttachment, GetAttachmentPath, OpenURL, ReadFileAsDataUrl } from '../../wailsjs/go/main/App'

import Prism from 'prismjs'
window.Prism = Prism
// PrismJS theme handled by custom CSS below (theme-adaptive via CSS variables)
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

const props = defineProps(['modelValue', 'noteId'])
const emit = defineEmits(['update:modelValue', 'open-note', 'show-link-picker'])

let tooltipProvider = null
let linkProvider = null
let tooltipEl = null
let slashEl = null
let linkEl = null
let pollTimer = null
let lastSelection = null
let linkInsertPos = null
let hoverTimeout = null

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
    { icon: '📎', label: '附件', action: 'attachment' },
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

function createBidirectionalLinkElement() {
  const el = document.createElement('div')
  el.className = 'md-link-dropdown'
  el.style.cssText = 'position:absolute;z-index:9999;display:none;min-width:240px;max-height:320px;overflow-y:auto;padding:6px;background:rgba(20,20,20,0.98);backdrop-filter:blur(20px);border:1px solid var(--glass-border);border-radius:12px;box-shadow:0 12px 40px rgba(0,0,0,0.7);'
  el.innerHTML = '<div class="link-search-input" style="padding:8px 12px;border-bottom:1px solid var(--glass-border);"><input type="text" placeholder="搜索笔记..." style="width:100%;background:transparent;border:none;outline:none;color:var(--text-primary);font-size:13px;" /></div><div class="link-results" style="max-height:260px;overflow-y:auto;"></div>'
  return el
}

function showLinkDropdown(coords, searchQuery = '') {
  if (!linkEl || !linkProvider) return
  linkProvider.show({
    getBoundingClientRect() {
      return {
        width: 0, height: 0,
        x: coords.left, y: coords.bottom + 8,
        top: coords.top, left: coords.left,
        right: coords.left, bottom: coords.bottom + 8,
      }
    }
  })
  const input = linkEl.querySelector('input')
  if (input) {
    input.value = searchQuery
    input.focus()
  }
  if (searchQuery) {
    searchLinkNotes(searchQuery)
  } else {
    searchLinkNotes('')
  }
}

function hideLinkDropdown() {
  if (linkProvider) linkProvider.hide()
  linkInsertPos = null
}

let linkSearchTimeout = null
async function searchLinkNotes(query) {
  if (!linkEl) return
  const resultsEl = linkEl.querySelector('.link-results')
  if (!resultsEl) return
  try {
    const results = await SearchNotes(query || '')
    if (results.length === 0) {
      resultsEl.innerHTML = '<div style="padding:16px;text-align:center;color:var(--text-secondary);font-size:12px;">未找到笔记</div>'
      return
    }
    resultsEl.innerHTML = results.slice(0, 10).map(note => `
      <div class="link-result-item" data-id="${note.id}" data-title="${note.title || note.name || '未命名'}" style="display:flex;align-items:center;gap:10px;padding:10px 12px;border-radius:8px;cursor:pointer;color:var(--text-secondary);transition:all 0.12s;">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="flex-shrink:0;color:var(--accent);">
          <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
        </svg>
        <span style="font-size:13px;color:var(--text-primary);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${note.title || note.name || '未命名'}</span>
      </div>
    `).join('')
  } catch (e) {
    resultsEl.innerHTML = '<div style="padding:16px;text-align:center;color:var(--text-secondary);font-size:12px;">搜索失败</div>'
  }
}

function insertBidirectionalLink(id, title) {
  const editor = get()
  if (!editor || linkInsertPos === null) return
  editor.action((ctx) => {
    const view = ctx.get(editorViewCtx)
    if (!view) return
    const { state, dispatch } = view
    const { tr } = state
    const linkText = `[[${id}|${title}]]`
    tr.insertText(linkText, linkInsertPos)
    dispatch(tr)
    view.focus()
  })
  hideLinkDropdown()
}

function handleLinkInput(e) {
  if (e.target.classList.contains('link-search-input') || e.target.closest('.link-search-input')) {
    const input = e.target.tagName === 'INPUT' ? e.target : e.target.querySelector('input')
    if (input) {
      clearTimeout(linkSearchTimeout)
      linkSearchTimeout = setTimeout(() => {
        searchLinkNotes(input.value)
      }, 200)
    }
  }
}

function handleLinkClick(e) {
  const item = e.target.closest('.link-result-item')
  if (item) {
    e.preventDefault()
    e.stopPropagation()
    insertBidirectionalLink(item.dataset.id, item.dataset.title)
  }
}

function insertLink(id, title) {
  const editor = get()
  if (!editor || linkInsertPos === null) return

  const insertPos = linkInsertPos
  linkInsertPos = null

  editor.action((ctx) => {
    const view = ctx.get(editorViewCtx)
    if (!view) return
    const { state } = view
    const { schema, tr } = state

    const linkTextLength = 2
    const pos = insertPos - linkTextLength

    // 方式1: 尝试作为 link 节点插入 (Milkdown 7.x)
    const linkNodeType = schema.nodes.link
    if (linkNodeType) {
      try {
        const linkNode = linkNodeType.create(
          { href: `note:${id}`, title: '' },
          schema.text(title)
        )
        const newTr = tr.delete(insertPos - linkTextLength, insertPos)
          .insert(pos, linkNode)
        view.dispatch(newTr)
        view.focus()
        return
      } catch (e) {
        console.error('[insertLink] Node insert failed:', e)
      }
    }

    // 方式2: 尝试作为 link mark 插入
    const linkMarkType = schema.marks.link
    if (linkMarkType) {
      try {
        const mark = linkMarkType.create({ href: `note:${id}`, title: '' })
        const textNode = schema.text(title, [mark])
        const newTr = tr.delete(insertPos - linkTextLength, insertPos)
          .insert(pos, textNode)
        view.dispatch(newTr)
        view.focus()
        return
      } catch (e) {
        console.error('[insertLink] Mark insert failed:', e)
      }
    }

    // 方式3: 回退到文本插入，然后重新解析文档
    const newTr = tr.delete(insertPos - linkTextLength, insertPos)
      .insertText(`[${title}](note:${id})`, pos)
    view.dispatch(newTr)
    view.focus()
  })

  // 方式3 后续: 延迟重新解析整个文档，让 Milkdown 将文本解析为 link 节点
  setTimeout(() => {
    if (!editor) return
    editor.action((ctx) => {
      try {
        const view = ctx.get(editorViewCtx)
        const parser = ctx.get(parserCtx)
        const serializer = ctx.get(serializerCtx)
        if (!view || !parser || !serializer) return

        const markdown = serializer(view.state.doc)
        const newDoc = parser(markdown)
        if (!newDoc) return

        const tr = view.state.tr
        const docSize = view.state.doc.content.size
        tr.replaceWith(0, docSize, newDoc)
        view.dispatch(tr)
      } catch (e) {
        console.error('[insertLink] Re-parse failed:', e)
      }
    })
  }, 50)
}

defineExpose({ insertLink })

const languages = [
  'javascript', 'typescript', 'python', 'go', 'rust', 'java', 'c', 'cpp',
  'csharp', 'ruby', 'php', 'swift', 'kotlin', 'scala', 'sql', 'bash',
  'shell', 'powershell', 'json', 'yaml', 'xml', 'html', 'css', 'scss',
  'markdown', 'dockerfile', 'toml', 'ini', 'lua', 'r', 'matlab', 'perl',
  'haskell', 'elixir', 'clojure', 'latex', 'graphql', 'nginx', 'apache',
]

let langPickerEl = null
function createLangPicker() {
  const el = document.createElement('div')
  el.className = 'lang-picker'
  el.style.cssText = 'position:fixed;z-index:10000;display:none;width:200px;background:var(--glass-bg);backdrop-filter:blur(20px);border:1px solid var(--glass-border);border-radius:8px;box-shadow:0 12px 32px rgba(0,0,0,0.6);overflow:hidden;'
  el.innerHTML = '<input type="text" placeholder="搜索语言..." style="width:100%;padding:8px 12px;background:transparent;border:none;border-bottom:1px solid var(--glass-border);outline:none;color:var(--text-primary);font-size:13px;font-family:inherit;box-sizing:border-box;" /><div class="lang-list" style="max-height:200px;overflow-y:auto;padding:4px;"></div>'
  return el
}

function showLangPicker(preEl, codeBlockPos) {
  if (!langPickerEl) {
    langPickerEl = createLangPicker()
    document.body.appendChild(langPickerEl)
  }
  const rect = preEl.getBoundingClientRect()
  langPickerEl.style.left = (rect.right - 200) + 'px'
  langPickerEl.style.top = (rect.bottom + 8) + 'px'
  langPickerEl.style.display = 'block'
  langPickerEl.dataset.pos = codeBlockPos

  const input = langPickerEl.querySelector('input')
  const listEl = langPickerEl.querySelector('.lang-list')
  input.value = ''
  input.focus()
  renderLangList(listEl, '')

  input.oninput = () => renderLangList(listEl, input.value)
  input.onkeydown = (e) => {
    if (e.key === 'Escape') { hideLangPicker(); return }
    if (e.key === 'Enter') {
      const first = listEl.querySelector('.lang-item')
      if (first) selectLanguage(first.dataset.lang)
    }
  }

  // Close on click outside
  setTimeout(() => {
    const closeHandler = (ev) => {
      if (!langPickerEl?.contains(ev.target)) {
        hideLangPicker()
        document.removeEventListener('mousedown', closeHandler)
      }
    }
    document.addEventListener('mousedown', closeHandler)
  }, 50)
}

function renderLangList(listEl, filter) {
  const filtered = filter
    ? languages.filter(l => l.includes(filter.toLowerCase()))
    : languages
  listEl.innerHTML = filtered.map(lang =>
    `<div class="lang-item" data-lang="${lang}" style="padding:6px 12px;border-radius:6px;cursor:pointer;font-size:12px;color:var(--text-secondary);font-family:'JetBrains Mono',monospace;transition:all 0.12s;">${lang}</div>`
  ).join('')
  listEl.querySelectorAll('.lang-item').forEach(el => {
    el.onclick = () => selectLanguage(el.dataset.lang)
  })
}

function selectLanguage(lang) {
  const pos = parseInt(langPickerEl?.dataset.pos)
  if (isNaN(pos)) return
  const editor = get()
  if (!editor) return
  editor.action((ctx) => {
    const view = ctx.get(editorViewCtx)
    if (!view) return
    const node = view.state.doc.nodeAt(pos)
    if (!node || node.type.name !== 'code_block') return
    const tr = view.state.tr.setNodeMarkup(pos, null, { ...node.attrs, language: lang })
    view.dispatch(tr)
    view.focus()
  })
  hideLangPicker()
}

function hideLangPicker() {
  if (langPickerEl) langPickerEl.style.display = 'none'
}

function handleEditorClick(e) {
  // Code block language picker
  const preEl = e.target.closest('.ProseMirror pre')
  if (preEl) {
    const rect = preEl.getBoundingClientRect()
    const clickX = e.clientX - rect.left
    const clickY = e.clientY - rect.top
    // Click in bottom-right area (language label)
    if (clickY > rect.height - 34 && clickX > rect.width * 0.5) {
      const editor = get()
      if (editor) {
        editor.action((ctx) => {
          const view = ctx.get(editorViewCtx)
          if (!view) return
          view.state.doc.descendants((node, pos) => {
            if (node.type.name === 'code_block') {
              const dom = view.nodeDOM(pos)
              if (dom === preEl || dom?.contains(preEl)) {
                showLangPicker(preEl, pos)
                return false
              }
            }
          })
        })
      }
      return
    }
  }

  const link = e.target.closest('.note-link') || e.target.closest('a[href^="note:"]')
  if (link) {
    e.preventDefault()
    const noteId = link.getAttribute('data-note-id') || link.getAttribute('href')?.replace(/^note:/, '')
    if (noteId) emit('open-note', noteId)
    return
  }

  const attachLink = e.target.closest('a[href^="assets/"]')
  if (attachLink) {
    const href = attachLink.getAttribute('href')
    if (href) {
      const ext = href.split('.').pop()?.toLowerCase() || ''
      if (isDocumentFile(ext)) {
        e.preventDefault()
        const parts = href.split('/')
        if (parts.length >= 3) {
          const noteId = parts[1]
          const filename = parts.slice(2).join('/')
          GetAttachmentPath(noteId, filename).then((fullPath) => {
            if (fullPath) OpenURL(fullPath)
          })
        }
      }
    }
  }
}

let currentHoverLink = null
let linkPreviewEl = null
let linkPreviewContainer = null
let mouseIsOverTooltip = false

function handleEditorHover(e) {
  const link = e.target.closest('.note-link') || e.target.closest('a[href^="note:"]')
  if (link && link !== currentHoverLink) {
    currentHoverLink = link
    clearTimeout(hoverTimeout)
    hoverTimeout = setTimeout(() => {
      showLinkPreview(link)
    }, 400)
  } else if (!link && currentHoverLink) {
    hideLinkPreview()
  }
}

function hideLinkPreview() {
  clearTimeout(hoverTimeout)
  if (mouseIsOverTooltip) return
  if (linkPreviewContainer) {
    linkPreviewContainer.style.display = 'none'
  }
  currentHoverLink = null
}

function escapeHtml(text) {
  const el = document.createElement('div')
  el.textContent = text
  return el.innerHTML
}

async function showLinkPreview(link) {
  if (!linkPreviewContainer) {
    linkPreviewContainer = document.createElement('div')
    linkPreviewContainer.className = 'link-preview-tooltip'
    linkPreviewContainer.addEventListener('click', () => {
      const id = linkPreviewContainer.dataset.noteId
      if (id) emit('open-note', id)
    })
    linkPreviewContainer.addEventListener('mouseenter', () => {
      mouseIsOverTooltip = true
    })
    linkPreviewContainer.addEventListener('mouseleave', () => {
      mouseIsOverTooltip = false
      hideLinkPreview()
    })
    document.body.appendChild(linkPreviewContainer)
  }
  const noteId = link.getAttribute('data-note-id') || link.getAttribute('href')?.replace(/^note:/, '')
  if (!noteId) return

  try {
    const content = await GetNoteContent(noteId)

    let title = '未命名'
    const titleMatch = content?.match(/^#\s+(.+)$/m)
    if (titleMatch) title = titleMatch[1].trim()

    let previewText = content || ''
    previewText = previewText.replace(/^#\s+.+$/m, '').trim()
    previewText = previewText
      .replace(/\*\*(.+?)\*\*/g, '$1')
      .replace(/\*(.+?)\*/g, '$1')
      .replace(/`(.+?)`/g, '$1')
      .replace(/\[(.+?)\]\(.+?\)/g, '$1')
      .replace(/!\[.*?\]\(.+?\)/g, '[图片]')

    const maxLen = 280
    const preview = previewText.substring(0, maxLen) + (previewText.length > maxLen ? '...' : '')
    const wordCount = content?.length || 0

    const rect = link.getBoundingClientRect()
    const left = Math.min(rect.left, window.innerWidth - 360)
    const top = rect.bottom + 12

    linkPreviewContainer.dataset.noteId = noteId
    linkPreviewContainer.style.cssText = `position:fixed;left:${left}px;top:${top}px;z-index:9999;width:320px;background:var(--glass-bg);backdrop-filter:blur(20px);border:1px solid var(--glass-border);border-radius:12px;box-shadow:0 12px 40px rgba(0,0,0,0.4);overflow:hidden;display:block;cursor:pointer;`

    linkPreviewContainer.innerHTML = `
      <div style="padding:14px 16px;border-bottom:1px solid var(--glass-border);display:flex;align-items:center;gap:10px;">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="color:var(--accent);flex-shrink:0;">
          <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
        </svg>
        <span style="color:var(--text-primary);font-size:14px;font-weight:600;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;">${escapeHtml(title)}</span>
      </div>
      <div style="padding:14px 16px;color:var(--text-secondary);font-size:12px;line-height:1.7;white-space:pre-wrap;word-break:break-word;max-height:200px;overflow-y:auto;">${escapeHtml(preview || '空白笔记')}</div>
      <div style="padding:10px 16px;border-top:1px solid var(--glass-border);color:var(--text-secondary);font-size:11px;opacity:0.6;display:flex;align-items:center;justify-content:space-between;">
        <span>点击打开笔记</span>
        <span style="font-size:10px;">${wordCount} 字</span>
      </div>
    `
  } catch (e) {
    if (linkPreviewContainer) linkPreviewContainer.style.display = 'none'
  }
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
      if (slashEl) slashEl.dataset.show = 'false'
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
        const hrType = schema.nodes.hr
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
        const itemType = schema.nodes.list_item
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
          const code = codeType.createAndFill()
          if (code) {
            tr = tr.replaceWith(paraPos, paraPos + paraNode.nodeSize, code)
          }
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
      case 'attachment': {
        const input = document.createElement('input')
        input.type = 'file'
        input.accept = 'image/png,image/jpeg,image/gif,image/webp,image/svg+xml,application/pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.csv'
        input.multiple = true
        input.onchange = async () => {
          const files = input.files
          if (!files || files.length === 0) return

          const editor = get()
          if (!editor) return

          for (const file of files) {
            try {
              editor.action((ctx) => {
                const view = ctx.get(editorViewCtx)
                if (!view) return
                const { schema: sch } = view.state
                uploadSingleFile(file, sch).then(node => {
                  if (!node) return
                  const { state, dispatch: disp } = view
                  const insertPos = state.selection.from
                  disp(state.tr.insert(insertPos, node))
                  view.focus()
                })
              })
            } catch (e) {
              console.error('Attachment upload failed:', e)
            }
          }
        }
        input.click()
        break
      }
    }

    dispatch(tr)
    view.focus()
    if (slashEl) slashEl.dataset.show = 'false'
  })
}

function readFileAsDataUrl(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result)
    reader.onerror = reject
    reader.readAsDataURL(file)
  })
}

const documentExts = ['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'csv']

function isImageFile(fileOrExt) {
  if (typeof fileOrExt === 'string') return ['png','jpg','jpeg','gif','webp','svg'].includes(fileOrExt)
  return fileOrExt.type?.startsWith('image/')
}

function isDocumentFile(fileOrExt) {
  if (typeof fileOrExt === 'string') return documentExts.includes(fileOrExt)
  const ext = fileOrExt.name?.split('.').pop()?.toLowerCase() || ''
  return documentExts.includes(ext)
}

// Upload a single file and return a ProseMirror node, or null
async function uploadSingleFile(file, schema) {
  const dataUrl = await readFileAsDataUrl(file)

  if (isImageFile(file)) {
    return schema.nodes.image.createAndFill({ src: dataUrl, alt: file.name })
  }

  if (isDocumentFile(file)) {
    const noteId = props.noteId
    if (!noteId || noteId.startsWith('new-')) return null
    const relativePath = await SaveAttachment(noteId, dataUrl, file.name)
    const linkMark = schema.marks.link.create({ href: relativePath })
    return schema.text(file.name, [linkMark])
  }

  return null
}

// Plugin: ensure there's always an empty paragraph at the end if the last node is a code block
const trailingParagraphKey = new PluginKey('trailing-paragraph')
const trailingParagraphPlugin = $prose((ctx) => {
  return new Plugin({
    key: trailingParagraphKey,
    appendTransaction: (_transactions, oldState, newState) => {
      const lastNode = newState.doc.lastChild
      if (!lastNode || lastNode.type.name === 'paragraph') return null
      const paragraph = newState.schema.nodes.paragraph.create()
      const tr = newState.tr.insert(newState.doc.content.size, paragraph)
      return tr
    },
  })
})

// Auto-indent in code blocks: preserve indentation on Enter
const autoIndentKey = new PluginKey('auto-indent')
const autoIndentPlugin = $prose(() => {
  return new Plugin({
    key: autoIndentKey,
    props: {
      handleKeyDown(view, event) {
        if (event.key !== 'Enter') return false
        const { $from } = view.state.selection
        if ($from.parent.type.name !== 'code_block') return false

        const textBefore = $from.parent.textContent.slice(0, $from.parentOffset)
        const lastLine = textBefore.split('\n').pop() || ''
        const indent = lastLine.match(/^(\s*)/)[1]

        const tr = view.state.tr.insertText('\n' + indent)
        view.dispatch(tr)
        return true
      },
    },
  })
})

// Auto-close brackets and quotes in code blocks
const bracketMatchKey = new PluginKey('bracket-match')
const bracketMatchPlugin = $prose(() => {
  const pairs = { '(': ')', '{': '}', '[': ']', '"': '"', "'": "'", '`': '`' }
  return new Plugin({
    key: bracketMatchKey,
    props: {
      handleTextInput(view, from, to, text) {
        const { $from } = view.state.selection
        if ($from.parent.type.name !== 'code_block') return false

        const closing = pairs[text]
        if (!closing) return false

        // Don't auto-close quotes if previous char is alphanumeric
        if (text === '"' || text === "'" || text === '`') {
          const prevChar = $from.parent.textContent[$from.parentOffset - 1]
          if (prevChar && /\w/.test(prevChar)) return false
        }

        const tr = view.state.tr.insertText(text + closing, from, to)
        // Place cursor between the pair
        const newPos = from + 1
        view.dispatch(tr.setSelection(view.state.selection.constructor.near(tr.doc.resolve(newPos))))
        return true
      },
    },
  })
})

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
      .use(trailingParagraphPlugin)
      .use(autoIndentPlugin)
      .use(bracketMatchPlugin)
}, [])

onMounted(() => {
  tooltipEl = createTooltipElement()
  slashEl = createSlashElement()
  linkEl = createBidirectionalLinkElement()
  document.body.appendChild(tooltipEl)
  document.body.appendChild(slashEl)
  document.body.appendChild(linkEl)

  tooltipProvider = new TooltipProvider({ content: tooltipEl, debounce: 50 })
  linkProvider = new SlashProvider({ content: linkEl, debounce: 0, offset: 8 })

  linkEl.addEventListener('input', handleLinkInput)
  linkEl.addEventListener('click', handleLinkClick)

  // Handle paste with images/documents directly on the editor DOM
  const editorEl = document.querySelector('.milkdown-wrapper')
  if (editorEl) {
    editorEl.addEventListener('paste', (e) => {
      const files = e.clipboardData?.files
      if (!files || files.length === 0) return
      const hasAttachment = Array.from(files).some(f =>
        isImageFile(f) || isDocumentFile(f)
      )
      if (!hasAttachment) return

      e.preventDefault()
      e.stopPropagation()

      const editor = get()
      if (!editor) return

      for (const file of files) {
        editor.action((ctx) => {
          const view = ctx.get(editorViewCtx)
          if (!view) return
          const { schema } = view.state
          const insertPos = view.state.selection.from

          uploadSingleFile(file, schema).then(node => {
            if (!node) return
            view.dispatch(view.state.tr.insert(insertPos, node))
            view.focus()
          }).catch(err => console.error('Paste upload failed:', err))
        })
      }
    }, true)
  }

  // Handle native file drops via Wails runtime
  // (Wails intercepts file drops at native level, webview DOM events don't fire)
  OnFileDrop(async (_x, _y, paths) => {
    if (!paths || paths.length === 0) return

    const editor = get()
    if (!editor) return

    for (const filePath of paths) {
      try {
        const fileName = filePath.split(/[/\\]/).pop() || 'file'
        const ext = fileName.split('.').pop()?.toLowerCase() || ''
        if (!isImageFile(ext) && !isDocumentFile(ext)) continue

        // Read file from disk via Go backend
        const dataUrl = await ReadFileAsDataUrl(filePath)

        editor.action((ctx) => {
          const view = ctx.get(editorViewCtx)
          if (!view) return
          const { schema } = view.state
          const insertPos = view.state.selection.from

          if (isImageFile(ext)) {
            const node = schema.nodes.image.createAndFill({ src: dataUrl, alt: fileName })
            if (node) {
              view.dispatch(view.state.tr.insert(insertPos, node))
              view.focus()
            }
          } else {
            // Document: save to assets and insert as link
            const noteId = props.noteId
            if (!noteId || noteId.startsWith('new-')) return
            SaveAttachment(noteId, dataUrl, fileName).then(relativePath => {
              const linkMark = schema.marks.link.create({ href: relativePath })
              const textNode = schema.text(fileName, [linkMark])
              view.dispatch(view.state.tr.insert(insertPos, textNode))
              view.focus()
            })
          }
        })
      } catch (e) {
        console.error('File drop upload failed:', e)
      }
    }
  }, false)

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
        if (view && slashEl) {
          const { $from } = view.state.selection
          const lineTextBefore = $from.parent.textBetween(0, $from.parentOffset)
          const isSlash = lineTextBefore === '/' || lineTextBefore.endsWith(' /')
          if (isSlash) {
            try {
              const coords = view.coordsAtPos($from.pos)
              const dropdownH = 320
              const spaceBelow = window.innerHeight - coords.bottom
              slashEl.style.left = coords.left + 'px'
              if (spaceBelow < dropdownH) {
                slashEl.style.top = (coords.top - dropdownH - 4) + 'px'
              } else {
                slashEl.style.top = (coords.bottom + 4) + 'px'
              }
              slashEl.dataset.show = 'true'
            } catch (e) {
              slashEl.dataset.show = 'false'
            }
          } else {
            slashEl.dataset.show = 'false'
          }
        }
        if (view) {
          const { $from } = view.state.selection
          const lineTextBefore = $from.parent.textBetween(0, $from.parentOffset)
          const endsWithLink = lineTextBefore.endsWith('[[') || lineTextBefore.endsWith(' [[')
          if (endsWithLink) {
            const pos = $from.pos
            if (linkInsertPos !== pos) {
              linkInsertPos = pos
              emit('show-link-picker', pos)
            }
          }
        }
      } catch (e) {}
    })
  }, 100)
})

onUnmounted(() => {
  try { OnFileDropOff() } catch (e) {}
  if (pollTimer) clearInterval(pollTimer)
  if (hoverTimeout) clearTimeout(hoverTimeout)
  if (tooltipProvider) tooltipProvider.destroy()
  if (linkProvider) linkProvider.destroy()
  if (tooltipEl?.parentNode) tooltipEl.parentNode.removeChild(tooltipEl)
  if (slashEl?.parentNode) slashEl.parentNode.removeChild(slashEl)
  if (linkEl?.parentNode) linkEl.parentNode.removeChild(linkEl)
  if (linkPreviewContainer?.parentNode) linkPreviewContainer.parentNode.removeChild(linkPreviewContainer)
  if (langPickerEl?.parentNode) langPickerEl.parentNode.removeChild(langPickerEl)
})
</script>

<template>
  <div class="milkdown-wrapper" @click="handleEditorClick" @mouseover="handleEditorHover" @mouseleave="hideLinkPreview">
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

/* Note links rendered by Milkdown default <a> tag - card style */
:deep(.ProseMirror a[href^="note:"]) {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: rgba(var(--accent-rgb), 0.08);
  border: 1px solid rgba(var(--accent-rgb), 0.2);
  border-radius: 8px;
  text-decoration: none;
  color: var(--accent);
  font-weight: 600;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
  max-width: 280px;
  vertical-align: middle;
}

:deep(.ProseMirror a[href^="note:"]:hover) {
  background: rgba(var(--accent-rgb), 0.15);
  border-color: rgba(var(--accent-rgb), 0.4);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

:deep(.ProseMirror a[href^="note:"]::before) {
  content: '';
  display: inline-block;
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  background-color: var(--accent);
  opacity: 0.8;
  -webkit-mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='2'%3E%3Cpath d='M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z'/%3E%3Cpolyline points='14 2 14 8 20 8'/%3E%3C/svg%3E");
  -webkit-mask-size: contain;
  -webkit-mask-repeat: no-repeat;
  mask-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='14' height='14' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='2'%3E%3Cpath d='M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z'/%3E%3Cpolyline points='14 2 14 8 20 8'/%3E%3C/svg%3E");
  mask-size: contain;
  mask-repeat: no-repeat;
}

/* Note links (custom node view fallback) */
:deep(.ProseMirror .note-link) {
  display: inline-block;
  vertical-align: middle;
  cursor: pointer;
  margin: 2px 0;
}

:deep(.ProseMirror .note-link-inner) {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: rgba(var(--accent-rgb), 0.08);
  border: 1px solid rgba(var(--accent-rgb), 0.2);
  border-radius: 8px;
  transition: all 0.2s;
  font-size: 13px;
  max-width: 280px;
}

:deep(.ProseMirror .note-link:hover .note-link-inner) {
  background: rgba(var(--accent-rgb), 0.15);
  border-color: rgba(var(--accent-rgb), 0.4);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

:deep(.ProseMirror .note-link-title) {
  color: var(--accent);
  font-weight: 600;
  font-size: 13px;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(.ProseMirror a[href^="note:"]) {
  color: var(--accent);
  font-weight: 600;
  font-size: 13px;
  line-height: 1.4;
}

:deep(.ProseMirror .note-link-icon) {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  color: var(--accent);
  opacity: 0.8;
  display: flex;
  align-items: center;
}

/* Document attachment links (PDF, Word, Excel, etc.) */
:deep(.ProseMirror a[href^="assets/"]) {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.25);
  border-radius: 6px;
  color: #f87171;
  font-size: 13px;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}

:deep(.ProseMirror a[href^="assets/"][href$=".pdf"]::before) {
  content: 'PDF';
  font-size: 10px;
  font-weight: 700;
  background: #ef4444;
  color: white;
  padding: 1px 4px;
  border-radius: 3px;
}

:deep(.ProseMirror a[href^="assets/"][href$=".doc"]::before),
:deep(.ProseMirror a[href^="assets/"][href$=".docx"]::before) {
  content: 'Word';
  font-size: 10px;
  font-weight: 700;
  background: #2b579a;
  color: white;
  padding: 1px 4px;
  border-radius: 3px;
}

:deep(.ProseMirror a[href^="assets/"][href$=".xls"]::before),
:deep(.ProseMirror a[href^="assets/"][href$=".xlsx"]::before) {
  content: 'Excel';
  font-size: 10px;
  font-weight: 700;
  background: #217346;
  color: white;
  padding: 1px 4px;
  border-radius: 3px;
}

:deep(.ProseMirror a[href^="assets/"][href$=".ppt"]::before),
:deep(.ProseMirror a[href^="assets/"][href$=".pptx"]::before) {
  content: 'PPT';
  font-size: 10px;
  font-weight: 700;
  background: #d24726;
  color: white;
  padding: 1px 4px;
  border-radius: 3px;
}

:deep(.ProseMirror a[href^="assets/"][href$=".csv"]::before) {
  content: 'CSV';
  font-size: 10px;
  font-weight: 700;
  background: #16a34a;
  color: white;
  padding: 1px 4px;
  border-radius: 3px;
}

:deep(.ProseMirror a[href^="assets/"]:hover) {
  background: rgba(239, 68, 68, 0.18);
  transform: translateY(-1px);
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
  height: 2px;
  background: linear-gradient(90deg, transparent 0%, var(--text-secondary) 15%, var(--text-secondary) 85%, transparent 100%);
  opacity: 0.4;
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

/* Code blocks — Typora style */
:deep(.ProseMirror pre) {
  background: rgba(128, 128, 128, 0.12);
  border: 1px solid rgba(128, 128, 128, 0.2);
  border-radius: 8px;
  padding: 0;
  margin: clamp(14px, 2.5vw, 20px) 0;
  overflow: hidden;
  position: relative;
}

/* Language label — bottom-right inside the code block */
:deep(.ProseMirror pre::after) {
  content: '选择语言';
  position: absolute;
  bottom: 6px;
  right: 6px;
  padding: 3px 10px;
  font-size: 11px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  color: var(--text-secondary);
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  cursor: pointer;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  transition: all 0.15s;
  z-index: 1;
}

:deep(.ProseMirror pre::after:hover) {
  color: var(--text-primary);
  border-color: rgba(255, 255, 255, 0.25);
  background: rgba(0, 0, 0, 0.45);
}

:deep(.ProseMirror pre[data-language]::after) {
  content: attr(data-language);
  color: var(--accent);
  background: rgba(0, 0, 0, 0.35);
  border-color: rgba(var(--accent-rgb), 0.4);
}

:deep(.ProseMirror pre[data-language]::after:hover) {
  border-color: var(--accent);
  background: rgba(0, 0, 0, 0.5);
}

:deep(.ProseMirror pre code) {
  display: block;
  background: transparent;
  color: var(--text-primary);
  padding: clamp(14px, 2.5vw, 20px);
  padding-top: clamp(18px, 2.5vw, 24px);
  font-size: clamp(12px, 1.5vw, 14px);
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;
  line-height: 1.7;
}

/* Theme-adaptive syntax highlighting via PrismJS */
:deep(.ProseMirror pre .token.comment),
:deep(.ProseMirror pre .token.prolog),
:deep(.ProseMirror pre .token.doctype),
:deep(.ProseMirror pre .token.cdata) {
  color: var(--text-secondary);
  opacity: 0.6;
  font-style: italic;
}

:deep(.ProseMirror pre .token.keyword),
:deep(.ProseMirror pre .token.tag),
:deep(.ProseMirror pre .token.boolean),
:deep(.ProseMirror pre .token.important) {
  color: var(--accent);
}

:deep(.ProseMirror pre .token.string),
:deep(.ProseMirror pre .token.char),
:deep(.ProseMirror pre .token.template-string),
:deep(.ProseMirror pre .token.attr-value) {
  color: #a3e635;
}

:deep(.ProseMirror pre .token.number),
:deep(.ProseMirror pre .token.variable),
:deep(.ProseMirror pre .token.constant) {
  color: #fbbf24;
}

:deep(.ProseMirror pre .token.function),
:deep(.ProseMirror pre .token.class-name),
:deep(.ProseMirror pre .token.method) {
  color: #60a5fa;
}

:deep(.ProseMirror pre .token.operator),
:deep(.ProseMirror pre .token.punctuation),
:deep(.ProseMirror pre .token.bracket) {
  color: var(--text-secondary);
}

:deep(.ProseMirror pre .token.property),
:deep(.ProseMirror pre .token.attr-name) {
  color: #c084fc;
}

:deep(.ProseMirror pre .token.regex),
:deep(.ProseMirror pre .token.important) {
  color: #f472b6;
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

.md-link-dropdown[data-show="true"] {
  display: block !important;
}

.link-result-item:hover {
  background: rgba(var(--accent-rgb), 0.1) !important;
  color: var(--text-primary) !important;
}

/* Language picker */
.lang-item:hover {
  background: rgba(var(--accent-rgb), 0.1) !important;
  color: var(--text-primary) !important;
}

.lang-picker::-webkit-scrollbar {
  width: 3px;
}
.lang-picker::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 2px;
}
</style>