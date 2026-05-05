<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { MilkdownProvider } from '@milkdown/vue'
import MilkdownCore from './MilkdownCore.vue'
import { GetNoteContent, CreateNote, UpdateNoteContent } from '../../wailsjs/go/main/App'

const props = defineProps(['status', 'noteId', 'user', 'newNoteRequest'])
const emit = defineEmits(['saved', 'back'])

// ==================== 多标签页 ====================
const openTabs = ref([])
const currentTabId = ref(null)
const tabCounter = ref(0)

const currentTab = computed(() => openTabs.value.find(t => t.id === currentTabId.value))

// 打开笔记（新建标签或切换）
async function openNote(note = null) {
  if (!note) {
    // 新建空白笔记
    const id = `new-${++tabCounter.value}`
    openTabs.value.push({ id, name: '未命名', content: '', isNew: true })
    currentTabId.value = id
    return
  }

  // 已打开则激活
  const existing = openTabs.value.find(t => t.id === note.id)
  if (existing) {
    currentTabId.value = note.id
    return
  }

  // 从后端加载笔记内容
  let content = note.content || ''
  if (!content && note.id && !note.id.startsWith('new-')) {
    try {
      content = await GetNoteContent(note.id)
    } catch (e) {
      console.error('Failed to load note content:', e)
    }
  }

  openTabs.value.push({ id: note.id, name: note.name, content, isNew: false })
  currentTabId.value = note.id
}

// 关闭标签
function closeTab(id) {
  const idx = openTabs.value.findIndex(t => t.id === id)
  if (idx === -1) return

  openTabs.value.splice(idx, 1)

  if (currentTabId.value === id) {
    currentTabId.value = openTabs.value[idx - 1]?.id || openTabs.value[idx]?.id || null
  }

  if (openTabs.value.length === 0) {
    goHome()
  }
}

// 初始打开一个空白标签或加载指定笔记
onMounted(async () => {
  if (props.noteId) {
    await openNote({ id: props.noteId, name: '' })
  } else {
    openNote()
  }
})

// 监听 noteId 变化
watch(() => props.noteId, async (newId) => {
  if (newId) {
    await openNote({ id: newId, name: '' })
  }
})

// 监听侧边栏"快速记录"请求（编辑器内新建空白标签）
watch(() => props.newNoteRequest, () => {
  openNote()
})

// ==================== 大纲视图 ====================
const outline = computed(() => {
  const content = currentTab.value?.content || ''
  return parseOutline(content)
})

function parseOutline(markdown) {
  if (!markdown) return []
  const lines = markdown.split('\n')
  const result = []
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i]
    const match = line.match(/^(#{1,6})\s+(.+)/)
    if (match) {
      result.push({
        level: match[1].length,
        text: match[2].trim(),
        line: i
      })
    }
  }
  return result
}

function scrollToHeading(line) {
  // 通知 MilkdownCore 滚动到指定行
  const event = new CustomEvent('scroll-to-line', { detail: { line } })
  window.dispatchEvent(event)
}
// 内容双向绑定
const currentContent = computed({
  get: () => currentTab.value?.content || '',
  set: (val) => {
    if (currentTab.value) {
      currentTab.value.content = val
    }
  }
})

function onContentChange(content) {
  if (!currentTab.value) return
  currentTab.value.content = content
}

// ==================== 保存 ====================
const extractTitle = (md) => {
  if (!md) return '未命名星谱'
  const match = md.match(/^#\s+(.+)$/m)
  return match ? match[1].trim() : '未命名星谱'
}

async function onSave() {
  if (!currentTab.value) return
  const title = extractTitle(currentTab.value.content)
  const content = currentTab.value.content

  try {
    if (currentTab.value.id && !currentTab.value.id.startsWith('new-')) {
      // 更新已有笔记
      await UpdateNoteContent(currentTab.value.id, content)
    } else {
      // 创建新笔记并回写真实 ID
      const newId = await CreateNote({
        name: title, file: content,
        parentPath: "root", parentId: props.user?.id || "", others: {}
      })
      if (newId) {
        currentTab.value.id = newId
      }
    }
    // 更新标签名
    currentTab.value.name = title
    currentTab.value.isNew = false
    emit('saved', { id: currentTab.value.id, name: currentTab.value.name })
  } catch (e) {
    console.error('保存失败:', e)
  }
}

function goHome() {
  openTabs.value = []
  currentTabId.value = null
  emit('back')
}
</script>

<template>
  <div class="editor-view-full">
    <!-- 标签栏 -->
    <div class="tab-bar">
      <div class="tabs-container">
        <div
          v-for="tab in openTabs"
          :key="tab.id"
          :class="['tab', { active: tab.id === currentTabId }]"
          @click="currentTabId = tab.id"
        >
          <span class="tab-name">{{ tab.name || '未命名' }}</span>
          <span class="tab-close" @click.stop="closeTab(tab.id)">×</span>
        </div>
      </div>
      <button class="tab-add" @click="openNote()">+</button>
    </div>

    <div class="editor-shell">
      <header class="editor-header">
        <div class="header-left">
          <button class="back-link" @click="goHome">← 退出 / EXIT</button>
        </div>
        <div class="header-right">
          <div class="status-box">
            <span class="dot"></span>
            {{ status }} / ASTRALINK_READY
          </div>
          <button class="sync-btn" @click="onSave">同步星链 / SYNC</button>
        </div>
      </header>

      <div class="editor-main">
        <!-- 侧边栏：大纲 -->
        <aside class="editor-sidebar">
          <div class="sidebar-section">
            <div class="outline-list">
              <div
                v-for="(item, idx) in outline"
                :key="idx"
                :class="['outline-item', `level-${item.level}`]"
                :style="{ paddingLeft: `${(item.level - 1) * 12 + 10}px` }"
                @click="scrollToHeading(item.line)"
              >
                {{ item.text }}
              </div>
              <div v-if="outline.length === 0" class="outline-empty">暂无标题</div>
            </div>
          </div>
        </aside>

        <!-- 编辑器内容 -->
        <div class="editor-scroll-area">
          <MilkdownProvider>
            <MilkdownCore v-model="currentContent" @update:modelValue="onContentChange" />
          </MilkdownProvider>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.editor-view-full {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: clamp(10px, 2vw, 20px);
  box-sizing: border-box;
}

/* 标签栏 */
.tab-bar {
  display: flex;
  align-items: center;
  gap: clamp(6px, 1.5vw, 10px);
  padding: 0 clamp(10px, 2vw, 20px);
  margin-bottom: clamp(8px, 1.5vw, 15px);
}

.tabs-container {
  display: flex;
  gap: clamp(3px, 0.8vw, 5px);
  flex: 1;
  overflow-x: auto;
}

.tab {
  display: flex;
  align-items: center;
  gap: clamp(5px, 1vw, 8px);
  padding: clamp(6px, 1.2vw, 8px) clamp(10px, 2vw, 15px);
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: clamp(5px, 1vw, 8px) clamp(5px, 1vw, 8px) 0 0;
  cursor: pointer;
  font-size: clamp(10px, 1.4vw, 12px);
  color: var(--text-secondary);
  transition: 0.2s;
  white-space: nowrap;
}

.tab.active {
  background: rgba(var(--accent-rgb), 0.15);
  border-color: var(--accent);
  color: var(--text-primary);
}

.tab:hover:not(.active) {
  background: rgba(255, 255, 255, 0.05);
}

.tab-name {
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  line-height: 1;
  opacity: 0.5;
  transition: 0.2s;
}

.tab-close:hover {
  opacity: 1;
  background: rgba(255, 255, 255, 0.1);
}

.tab-add {
  width: clamp(28px, 4vw, 32px);
  height: clamp(28px, 4vw, 32px);
  border: 1px dashed var(--glass-border);
  background: transparent;
  border-radius: clamp(4px, 0.8vw, 6px);
  color: var(--text-secondary);
  cursor: pointer;
  font-size: clamp(14px, 2vw, 18px);
  transition: 0.2s;
}

.tab-add:hover {
  border-color: var(--accent);
  color: var(--accent);
}

/* 编辑器主体 */
.editor-shell {
  flex: 1;
  background: var(--glass-bg);
  backdrop-filter: blur(40px);
  border: 1px solid var(--glass-border);
  border-radius: clamp(10px, 2vw, 16px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 50px 100px rgba(0,0,0,0.8);
}

.editor-header {
  padding: clamp(10px, 2vw, 15px) clamp(20px, 5vw, 40px);
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--glass-border);
}

.sync-btn {
  background: var(--accent);
  color: #fff;
  padding: clamp(8px, 1.5vw, 10px) clamp(18px, 4vw, 28px);
  border-radius: clamp(4px, 0.8vw, 6px);
  font-weight: 700;
  border: none;
  cursor: pointer;
  font-size: clamp(10px, 1.4vw, 12px);
  transition: 0.2s;
}
.sync-btn:hover { opacity: 0.9; }

.back-link {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: clamp(11px, 1.5vw, 13px);
}
.back-link:hover { color: var(--text-primary); }

.status-box {
  font-size: clamp(9px, 1.2vw, 10px);
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: clamp(5px, 1vw, 8px);
  margin-right: clamp(15px, 3vw, 25px);
}

.dot { width: clamp(5px, 0.8vw, 6px); height: clamp(5px, 0.8vw, 6px); border-radius: 50%; background: #22c55e; }

/* 编辑器主体区域 */
.editor-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* 侧边栏 */
.editor-sidebar {
  width: min(250px, 30%);
  border-right: 1px solid var(--glass-border);
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.sidebar-tabs {
  display: flex;
  border-bottom: 1px solid var(--glass-border);
  padding: clamp(8px, 1.5vw, 12px) clamp(8px, 1.5vw, 12px) 0;
  gap: clamp(4px, 0.8vw, 6px);
}

.sidebar-tab {
  flex: 1;
  padding: clamp(6px, 1vw, 10px) clamp(8px, 1.5vw, 12px);
  text-align: center;
  font-size: clamp(10px, 1.4vw, 12px);
  color: var(--text-secondary);
  border-radius: clamp(4px, 0.8vw, 6px) clamp(4px, 0.8vw, 6px) 0 0;
  cursor: pointer;
  transition: 0.2s;
}

.sidebar-tab:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.05);
}

.sidebar-tab.active {
  color: var(--accent);
  font-weight: 700;
  background: rgba(var(--accent-rgb), 0.1);
}

.sidebar-section {
  flex: 1;
  padding: clamp(12px, 2vw, 20px) clamp(10px, 1.8vw, 15px);
  overflow-y: auto;
}

/* NTree 主题适配 */
:deep(.n-tree) {
  --n-node-text-color: var(--text-secondary);
  --n-node-text-color-hover: var(--text-primary);
  --n-node-text-color-active: var(--accent);
  --n-arrow-color: var(--text-secondary);
  background: transparent;
}

:deep(.n-tree-node) {
  padding: clamp(2px, 0.5vw, 4px) 0;
}

:deep(.n-tree-node-content) {
  padding: clamp(4px, 0.8vw, 6px) clamp(6px, 1vw, 8px);
  border-radius: clamp(3px, 0.6vw, 5px);
  font-size: clamp(10px, 1.4vw, 12px);
  transition: 0.15s;
}

:deep(.n-tree-node-content:hover) {
  background: rgba(255, 255, 255, 0.05);
}

:deep(.n-tree-node--selected .n-tree-node-content) {
  background: rgba(var(--accent-rgb), 0.15);
  color: var(--accent);
}

.section-title {
  font-size: clamp(9px, 1.2vw, 10px);
  color: var(--text-secondary);
  letter-spacing: clamp(0.5px, 0.15vw, 1px);
  margin-bottom: clamp(6px, 1vw, 10px);
  opacity: 0.6;
}

/* 大纲列表 */
.outline-list {
  display: flex;
  flex-direction: column;
  gap: clamp(1px, 0.3vw, 2px);
}

.outline-item {
  padding: clamp(4px, 0.8vw, 6px) clamp(6px, 1.2vw, 10px);
  border-radius: clamp(2px, 0.5vw, 4px);
  font-size: clamp(10px, 1.4vw, 12px);
  color: var(--text-secondary);
  cursor: pointer;
  transition: 0.2s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.outline-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.outline-item.level-1 { font-weight: 600; color: var(--text-primary); }
.outline-item.level-2 { font-size: clamp(9px, 1.3vw, 11px); }
.outline-item.level-3 { font-size: clamp(9px, 1.3vw, 11px); opacity: 0.8; }

.outline-empty {
  font-size: clamp(9px, 1.3vw, 11px);
  color: var(--text-secondary);
  opacity: 0.5;
  padding: clamp(6px, 1.2vw, 10px);
  text-align: center;
}

/* 编辑区域 */
.editor-scroll-area {
  flex: 1;
  overflow-y: auto;
  padding: clamp(25px, 5vw, 40px) clamp(8%, 12%, 15%);
  cursor: text;
  position: relative;
}

.editor-scroll-area::-webkit-scrollbar { width: 4px; }
.editor-scroll-area::-webkit-scrollbar-thumb { background: var(--glass-border); }

/* 动画 */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>