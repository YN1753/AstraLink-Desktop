<script setup>
import { ref } from 'vue'
import { MilkdownProvider } from '@milkdown/vue'
import MilkdownCore from './MilkdownCore.vue'

defineProps(['status'])
const emit = defineEmits(['save', 'back'])

const content = ref('')

// 智能提取正文第一个 H1 作为保存标题
const extractTitle = (md) => {
  const match = md.match(/^#\s+(.+)$/m)
  return match ? match[1].trim() : '未命名星谱'
}

const onSave = () => {
  const title = extractTitle(content.value)
  emit('save', { title: title, content: content.value })
}
</script>

<template>
  <div class="editor-view-full">
    <MilkdownProvider>
      <div class="editor-shell">
        <header class="editor-header">
          <div class="header-left">
            <button class="back-link" @click="$emit('back')">← 退出 / EXIT</button>
          </div>
          <div class="header-right">
            <div class="status-box">
              <span class="dot"></span>
              {{ status }} / ASTRALINK_READY
            </div>
            <button class="sync-btn" @click="onSave">同步星链 / SYNC</button>
          </div>
        </header>

        <div class="editor-scroll-area">
          <MilkdownCore v-model="content" />
        </div>
      </div>
    </MilkdownProvider>
  </div>
</template>

<style scoped>
.editor-view-full {
  width: 100%; height: 100%; display: flex; justify-content: center; align-items: center;
  padding: 20px; box-sizing: border-box;
}
.editor-shell {
  width: 100%; max-width: 1200px; height: 90vh;
  background: rgba(10, 10, 10, 0.8); backdrop-filter: blur(40px);
  border: 1px solid rgba(255, 255, 255, 0.08); border-radius: 16px;
  display: flex; flex-direction: column; overflow: hidden;
  box-shadow: 0 50px 100px rgba(0,0,0,0.8);
}
.editor-header {
  padding: 15px 40px; display: flex; justify-content: space-between; align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}
.sync-btn {
  background: #fff; color: #000; padding: 10px 28px; border-radius: 6px;
  font-weight: 800; border: none; cursor: pointer; font-size: 12px; transition: 0.2s;
}
.sync-btn:hover { background: #3b82f6; color: #fff; }
.back-link { background: transparent; border: none; color: #555; cursor: pointer; font-size: 13px; }
.back-link:hover { color: #fff; }
.status-box { font-size: 10px; color: #444; display: flex; align-items: center; gap: 8px; margin-right: 25px; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: #22c55e; }
.editor-scroll-area { flex: 1; overflow-y: auto; padding: 40px 12%; cursor: text; }
.editor-scroll-area::-webkit-scrollbar { width: 4px; }
.editor-scroll-area::-webkit-scrollbar-thumb { background: #222; }
</style>