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

// 🌟 核心：必须在所有逻辑之前初始化全局 Prism
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
        // 这里的空格非常关键，防止空文档导致的渲染节点丢失
        ctx.set(defaultValueCtx, props.modelValue || "# 未命名星谱\n\n在此输入正文...")
        ctx.get(listenerCtx).markdownUpdated((_, markdown) => {
          emit('update:modelValue', markdown)
        })
      })
      .config(nord)
      // 🌟 必须严格遵守此顺序：commonmark 必须第一个加载
      .use(commonmark)
      .use(gfm)
      .use(history)
      .use(prism)
      .use(indent)
      .use(listener)
}, [])
</script>

<template>
  <div class="milkdown-pro-wrapper">
    <Milkdown />
  </div>
</template>

<style scoped>
.milkdown-pro-wrapper { width: 100%; min-height: 500px; text-align: left; }

/* 🌟 强制显示光标和正文区域 */
:deep(.milkdown) { background: transparent !important; }
:deep(.ProseMirror) {
  outline: none !important;
  color: #d1d5db !important;
  line-height: 1.8;
  font-size: 17px;
  min-height: 600px; /* 确保大面积可点 */
  padding-bottom: 200px;
}

/* 标题样式：发光特效 */
:deep(.ProseMirror h1) {
  font-weight: 200; color: #fff; border-bottom: 1px solid #3b82f6;
  padding-bottom: 10px; margin: 30px 0;
  text-shadow: 0 0 10px rgba(59, 130, 246, 0.3);
}

/* 代码块增强 */
:deep(.milkdown .code-fence) {
  background: #1e1e1e !important; border: 1px solid #333;
  border-radius: 8px; padding: 1.5rem; margin: 20px 0;
}
</style>