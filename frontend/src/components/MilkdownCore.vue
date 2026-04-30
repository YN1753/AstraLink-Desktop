<script setup>
import { Milkdown, useEditor } from '@milkdown/vue'
import { Editor, rootCtx, defaultValueCtx } from '@milkdown/core'
import { commonmark } from '@milkdown/preset-commonmark'
import { gfm } from '@milkdown/preset-gfm'
import { history } from '@milkdown/plugin-history'
import { prism } from '@milkdown/plugin-prism'
import { indent } from '@milkdown/plugin-indent'
import { listener, listenerCtx } from '@milkdown/plugin-listener'
import { nord } from '@milkdown/theme-nord'

// 全局 Prism
import Prism from 'prismjs'
window.Prism = Prism
import 'prismjs/themes/prism-tomorrow.css'
import 'prismjs/components/prism-go'
import 'prismjs/components/prism-javascript'

const props = defineProps(['modelValue'])
const emit = defineEmits(['update:modelValue'])

useEditor((root) => {
  return Editor.make()
      .config((ctx) => {
        ctx.set(rootCtx, root)
        ctx.set(defaultValueCtx, props.modelValue || "# 未命名星谱\n\n在此输入正文...")
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
}, [])

// 监听滚动到行事件
if (typeof window !== 'undefined') {
  window.addEventListener('scroll-to-line', (e) => {
    const { line } = e.detail
    // 通过 DOM 操作滚动到对应行
    setTimeout(() => {
      const editor = document.querySelector('.ProseMirror')
      if (editor) {
        const lines = editor.querySelectorAll('p, h1, h2, h3, h4, h5, h6, li')
        if (lines[line]) {
          lines[line].scrollIntoView({ behavior: 'smooth', block: 'start' })
        }
      }
    }, 100)
  })
}
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
}

/* 占位符提示文字 */
:deep(.ProseMirror p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  color: var(--text-secondary);
  opacity: 0.4;
  pointer-events: none;
  float: left;
  height: 0;
}

/* 标题样式 */
:deep(.ProseMirror h1) {
  font-weight: 200; color: var(--text-primary); border-bottom: 1px solid var(--accent);
  padding-bottom: 10px; margin: 30px 0;
}
:deep(.ProseMirror h2),
:deep(.ProseMirror h3),
:deep(.ProseMirror h4) {
  color: var(--text-primary);
}

/* 链接样式 */
:deep(.ProseMirror a) {
  color: var(--accent);
  text-decoration: none;
}
:deep(.ProseMirror a:hover) {
  text-decoration: underline;
}

/* 双链 [[笔记名]] 样式 */
:deep(.ProseMirror .wiki-link) {
  color: var(--accent);
  background: rgba(59, 130, 246, 0.1);
  padding: 2px 6px;
  border-radius: 4px;
  cursor: pointer;
  transition: 0.2s;
}
:deep(.ProseMirror .wiki-link:hover) {
  background: rgba(59, 130, 246, 0.2);
}
:deep(.ProseMirror .wiki-link.missing) {
  color: #ef4444;
  background: rgba(239, 68, 68, 0.1);
}

/* 代码块 */
:deep(.milkdown .code-fence) {
  background: var(--bg-app) !important;
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  padding: 1.5rem;
  margin: 20px 0;
}
</style>