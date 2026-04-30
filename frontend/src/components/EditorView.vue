<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { MilkdownProvider } from '@milkdown/vue'
import MilkdownCore from './MilkdownCore.vue'
import { NTree, NInput, NButton, NPopover } from 'naive-ui'

const props = defineProps(['status'])
const emit = defineEmits(['save', 'back'])

// ==================== 多标签页 ====================
const openTabs = ref([])
const currentTabId = ref(null)
const tabCounter = ref(0)

const currentTab = computed(() => openTabs.value.find(t => t.id === currentTabId.value))

// 打开笔记（新建标签或切换）
function openNote(note = null) {
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

  openTabs.value.push({ id: note.id, name: note.name, content: note.content || '' })
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
    emit('back')
  }
}

// 初始打开一个空白标签
onMounted(() => {
  openNote()
})

// ==================== 侧边栏标签切换 ====================
const sidebarTab = ref('outline')

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

// ==================== 树形导航 ====================
// 模拟笔记列表（用于树形导航）
const mockNotes = [
  { id: 'note-1', name: '项目笔记', content: '# 项目笔记\n\n这是项目相关的内容。' },
  { id: 'note-2', name: 'Go学习笔记', content: '# Go学习笔记\n\n## 变量\n\n学习Go语言的基础。\n\n### 数据类型\n\n- string\n- int\n- bool' },
  { id: 'note-3', name: '周报总结', content: '# 周报总结\n\n本周工作内容。' },
  { id: 'note-4', name: '读书笔记', content: '# 读书笔记\n\n读《深入理解计算机系统》。' },
]

// 模拟树数据
const treeData = ref([
  {
    key: 'root',
    label: '📁 root',
    children: [
      {
        key: 'root/work',
        label: '📁 工作',
        children: [
          { key: 'note-1', label: '📄 项目笔记', isLeaf: true },
          { key: 'note-3', label: '📄 周报总结', isLeaf: true },
        ]
      },
      {
        key: 'root/study',
        label: '📁 学习',
        children: [
          { key: 'note-2', label: '📄 Go学习笔记', isLeaf: true },
          { key: 'note-4', label: '📄 读书笔记', isLeaf: true },
        ]
      }
    ]
  }
])

const expandedKeys = ref(['root', 'root/work', 'root/study'])
const selectedKey = ref(null)

function onTreeSelect(keys) {
  if (keys.length > 0) {
    selectedKey.value = keys[0]
    // 找到对应的笔记打开
    const note = mockNotes.find(n => `note-${n.id}` === keys[0] || n.id === keys[0])
    if (note) {
      openNote({ id: note.id, name: note.name })
    }
  }
}

// ==================== 双链笔记 ====================
// 内容双向绑定
const currentContent = computed({
  get: () => currentTab.value?.content || '',
  set: (val) => {
    if (currentTab.value) {
      currentTab.value.content = val
    }
  }
})

const showLinkPicker = ref(false)
const linkSearchQuery = ref('')
const linkPickerPosition = ref({ top: 0, left: 0 })

const filteredNotes = computed(() => {
  const query = linkSearchQuery.value.toLowerCase()
  if (!query) return mockNotes.slice(0, 5)
  return mockNotes.filter(n => n.name.toLowerCase().includes(query)).slice(0, 5)
})

// 监听内容中的 [[ 触发链接选择器
function onContentChange(content) {
  if (!currentTab.value) return
  currentTab.value.content = content

  // 检测 [[ 触发链接选择器（简化实现）
  const lastBracket = content.lastIndexOf('[[')
  if (lastBracket !== -1) {
    const afterBracket = content.slice(lastBracket + 2)
    if (!afterBracket.includes(']]')) {
      linkSearchQuery.value = afterBracket
      showLinkPicker.value = true
    }
  } else {
    showLinkPicker.value = false
  }
}

function insertLink(noteName) {
  if (!currentTab.value) return

  const content = currentTab.value.content
  const lastBracket = content.lastIndexOf('[[')

  if (lastBracket !== -1) {
    const before = content.slice(0, lastBracket)
    const after = content.slice(lastBracket + 2)
    const bracketIdx = after.indexOf(']]')
    const finalAfter = bracketIdx !== -1 ? after.slice(bracketIdx + 2) : after

    currentTab.value.content = `${before}[[${noteName}]]${finalAfter}`
  }

  showLinkPicker.value = false
  linkSearchQuery.value = ''
}

// ==================== 保存 ====================
const extractTitle = (md) => {
  if (!md) return '未命名星谱'
  const match = md.match(/^#\s+(.+)$/m)
  return match ? match[1].trim() : '未命名星谱'
}

const onSave = () => {
  if (!currentTab.value) return
  const title = extractTitle(currentTab.value.content)
  emit('save', { title, content: currentTab.value.content })

  // 更新标签名
  if (currentTab.value.isNew) {
    currentTab.value.name = title
    currentTab.value.isNew = false
  }
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

        <div class="editor-main">
          <!-- 侧边栏：标签切换（大纲/文件树） -->
          <aside class="editor-sidebar">
            <div class="sidebar-tabs">
              <div
                class="sidebar-tab"
                :class="{ active: sidebarTab === 'outline' }"
                @click="sidebarTab = 'outline'"
              >
                📑 大纲
              </div>
              <div
                class="sidebar-tab"
                :class="{ active: sidebarTab === 'tree' }"
                @click="sidebarTab = 'tree'"
              >
                📂 文件
              </div>
            </div>

            <!-- 大纲视图 -->
            <div v-show="sidebarTab === 'outline'" class="sidebar-section">
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

            <!-- 树形导航 -->
            <div v-show="sidebarTab === 'tree'" class="sidebar-section">
              <NTree
                :data="treeData"
                :default-expanded-keys="expandedKeys"
                :selected-keys="selectedKey ? [selectedKey] : []"
                block-line
                selectable
                @update:selected-keys="onTreeSelect"
              />
            </div>
          </aside>

          <!-- 编辑器内容 -->
          <div class="editor-scroll-area">
            <MilkdownCore v-model="currentContent" @update:modelValue="onContentChange" />

            <!-- 双链选择器 -->
            <Transition name="fade">
              <div v-if="showLinkPicker" class="link-picker glass-panel">
                <div class="link-picker-header">插入链接</div>
                <NInput
                  v-model:value="linkSearchQuery"
                  placeholder="搜索笔记..."
                  size="small"
                  autofocus
                />
                <div class="link-picker-list">
                  <div
                    v-for="note in filteredNotes"
                    :key="note.id"
                    class="link-picker-item"
                    @click="insertLink(note.name)"
                  >
                    📄 {{ note.name }}
                  </div>
                  <div v-if="filteredNotes.length === 0" class="link-picker-empty">
                    未找到笔记
                  </div>
                </div>
              </div>
            </Transition>
          </div>
        </div>
      </div>
    </MilkdownProvider>
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

/* 双链选择器 */
.link-picker {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: min(300px, 80vw);
  max-height: min(350px, 60vh);
  border-radius: clamp(8px, 1.5vw, 12px);
  padding: clamp(10px, 2vw, 15px);
  z-index: 50;
  overflow: hidden;
}

.link-picker-header {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 10px;
  letter-spacing: 1px;
}

.link-picker-list {
  margin-top: 10px;
  max-height: 250px;
  overflow-y: auto;
}

.link-picker-item {
  padding: 10px 12px;
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-primary);
  cursor: pointer;
  transition: 0.2s;
}

.link-picker-item:hover {
  background: rgba(59, 130, 246, 0.15);
}

.link-picker-empty {
  text-align: center;
  color: var(--text-secondary);
  font-size: 12px;
  padding: 20px;
  opacity: 0.6;
}

/* 动画 */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>