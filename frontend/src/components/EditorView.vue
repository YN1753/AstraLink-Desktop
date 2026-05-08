<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { MilkdownProvider } from '@milkdown/vue'
import MilkdownCore from './MilkdownCore.vue'
import TagInput from './TagInput.vue'
import { GetNoteContent, CreateNote, UpdateNoteContent, UpdateNodeInfo, LinkTagToNode, GetTagsWithCount, GetTagsByNodeID } from '../../wailsjs/go/main/App'

const props = defineProps(['status', 'noteId', 'user', 'newNoteRequest'])
const emit = defineEmits(['saved', 'back', 'refresh-tags'])

const openTabs = ref([])
const currentTabId = ref(null)
const tabCounter = ref(0)
const mounted = ref(false)

const currentTab = computed(() => openTabs.value.find(t => t.id === currentTabId.value))

async function openNote(note = null, forceReload = false) {
  if (!note) {
    const id = `new-${++tabCounter.value}`
    openTabs.value.push({ id, name: '未命名', content: '', isNew: true })
    currentTabId.value = id
    noteTags.value[id] = []
    return
  }

  const existing = openTabs.value.find(t => t.id === note.id)
  if (existing && !forceReload) {
    currentTabId.value = note.id
    return
  }

  let content = ''
  let title = note.name || ''
  if (note.id && !note.id.startsWith('new-')) {
    try {
      content = await GetNoteContent(note.id)
      if (!title && content) {
        title = extractTitle(content)
      }
    } catch (e) {}
  }

  if (existing && forceReload) {
    existing.content = content || note.content || ''
    currentTabId.value = note.id
  } else {
    openTabs.value.push({ id: note.id, name: title, content: content || note.content || '', isNew: false })
    currentTabId.value = note.id
  }

  await loadNoteTags(note.id)
}

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

onMounted(async () => {
  setTimeout(() => { mounted.value = true }, 50)
  if (props.noteId) {
    await openNote({ id: props.noteId, name: '' })
  } else {
    openNote()
  }
})

watch(() => props.noteId, async (newId) => {
  if (newId) {
    await openNote({ id: newId, name: '' }, true)
  }
})

watch(() => props.newNoteRequest, () => {
  openNote()
})

const outline = computed(() => {
  const content = currentTab.value?.content || ''
  return parseOutline(content)
})

function stripInlineMarkdown(text) {
  return text
    .replace(/\*\*(.+?)\*\*/g, '$1')
    .replace(/__(.+?)__/g, '$1')
    .replace(/\*(.+?)\*/g, '$1')
    .replace(/_(.+?)_/g, '$1')
    .replace(/`(.+?)`/g, '$1')
    .replace(/~~(.+?)~~/g, '$1')
    .replace(/\[(.+?)\]\(.+?\)/g, '$1')
    .replace(/!\[.*?\]\(.+?\)/g, '')
}

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
        text: stripInlineMarkdown(match[2].trim()),
        line: i
      })
    }
  }
  return result
}

function scrollToHeading(line) {
  const event = new CustomEvent('scroll-to-line', { detail: { line } })
  window.dispatchEvent(event)
}

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
      await UpdateNoteContent(currentTab.value.id, content)
      await UpdateNodeInfo({
        id: currentTab.value.id,
        title: title,
        path: '',
        others: {}
      })
    } else {
      const newId = await CreateNote({
        name: title, file: content,
        parentPath: "root", parentId: props.user?.id || "", others: {}
      })
      if (newId) {
        currentTab.value.id = newId
      }
    }
    currentTab.value.name = title
    currentTab.value.isNew = false

    const tags = noteTags.value[currentTab.value.id] || []
    if (tags.length > 0) {
      await linkTagsToNote(currentTab.value.id, tags)
    }

    emit('saved', { id: currentTab.value.id, name: currentTab.value.name })
    emit('refresh-tags')
  } catch (e) {
    console.error('保存失败:', e)
  }
}

function goHome() {
  openTabs.value = []
  currentTabId.value = null
  emit('back')
}

const noteTags = ref({})

async function loadNoteTags(noteId) {
  if (!noteId || noteId.startsWith('new-')) {
    noteTags.value[noteId] = []
    return
  }
  try {
    const tagMessages = await GetTagsByNodeID(noteId)
    if (tagMessages && tagMessages.length > 0) {
      noteTags.value[noteId] = tagMessages.map(t => ({
        id: t.id,
        name: t.name,
        noteCount: 0
      }))
    } else {
      noteTags.value[noteId] = []
    }
  } catch (e) {
    noteTags.value[noteId] = []
  }
}

function handleTagsChanged(noteId, tags) {
  noteTags.value[noteId] = tags
}

async function linkTagsToNote(noteId, tags) {
  if (!noteId || noteId.startsWith('new-') || !tags || tags.length === 0) return
  try {
    for (const tag of tags) {
      await LinkTagToNode(tag.id, noteId)
    }
  } catch (e) {}
}
</script>

<template>
  <div class="editor-view-full" :class="{ mounted }">
    <!-- Tab bar -->
    <div class="tab-bar">
      <div class="tabs-container">
        <button
          v-for="tab in openTabs"
          :key="tab.id"
          :class="['tab', { active: tab.id === currentTabId }]"
          @click="currentTabId = tab.id"
        >
          <span class="tab-name">{{ tab.name || '未命名' }}</span>
          <span class="tab-close" @click.stop="closeTab(tab.id)">
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </span>
        </button>
      </div>
      <button class="tab-add" @click="openNote()">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19"/>
          <line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
      </button>
    </div>

    <div class="editor-shell">
      <!-- Header -->
      <header class="editor-header">
        <div class="header-left">
          <button class="back-btn" @click="goHome">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M19 12H5M12 19l-7-7 7-7"/>
            </svg>
            <span>退出</span>
          </button>
        </div>
        <div class="header-right">
          <div class="status-indicator">
            <span class="status-dot"></span>
            <span>{{ status }}</span>
          </div>
          <button class="save-btn" @click="onSave">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M19 21H5a2 2 0 01-2-2V5a2 2 0 012-2h11l5 5v11a2 2 0 01-2 2z"/>
              <polyline points="17 21 17 13 7 13 7 21"/>
              <polyline points="7 3 7 8 15 8"/>
            </svg>
            <span>同步</span>
          </button>
        </div>
      </header>

      <!-- Tags -->
      <TagInput
        v-if="currentTab && !currentTab.isNew"
        :noteId="currentTab.id"
        :noteTags="noteTags[currentTab.id] || []"
        @tags-changed="(tags) => handleTagsChanged(currentTab.id, tags)"
      />

      <!-- Main area -->
      <div class="editor-main">
        <!-- Sidebar - Outline -->
        <aside class="editor-sidebar">
          <div class="sidebar-header">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
            <span>大纲</span>
          </div>
          <div class="outline-list">
            <button
              v-for="(item, idx) in outline"
              :key="idx"
              :class="['outline-item', `level-${item.level}`]"
              :style="{ paddingLeft: `${(item.level - 1) * 12 + 12}px` }"
              @click="scrollToHeading(item.line)"
            >
              {{ item.text }}
            </button>
            <div v-if="outline.length === 0" class="outline-empty">暂无标题</div>
          </div>
        </aside>

        <!-- Editor content -->
        <div class="editor-scroll-area">
          <MilkdownProvider>
            <MilkdownCore :key="currentTabId" v-model="currentContent" @update:modelValue="onContentChange" />
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
  padding: clamp(8px, 2vw, 16px);
  box-sizing: border-box;
  opacity: 0;
  transition: opacity 0.4s ease;
}

.editor-view-full.mounted {
  opacity: 1;
}

/* Tab bar */
.tab-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 clamp(8px, 2vw, 16px);
  margin-bottom: 12px;
}

.tabs-container {
  display: flex;
  gap: 4px;
  flex: 1;
  overflow-x: auto;
  padding: 4px 0;
}

.tabs-container::-webkit-scrollbar {
  height: 3px;
}

.tabs-container::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 2px;
}

.tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid transparent;
  border-radius: 8px 8px 0 0;
  cursor: pointer;
  font-size: 12px;
  color: var(--text-secondary);
  transition: all 0.2s;
  font-family: inherit;
  white-space: nowrap;
}

.tab:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-primary);
}

.tab.active {
  background: var(--glass-bg);
  border-color: var(--glass-border);
  border-bottom-color: transparent;
  color: var(--text-primary);
}

.tab-name {
  max-width: 100px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tab-close {
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  opacity: 0.5;
  transition: all 0.15s;
}

.tab-close:hover {
  opacity: 1;
  background: rgba(255, 255, 255, 0.1);
}

.tab-add {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px dashed var(--glass-border);
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  flex-shrink: 0;
}

.tab-add:hover {
  border-color: var(--accent);
  color: var(--accent);
  border-style: solid;
}

/* Editor shell */
.editor-shell {
  flex: 1;
  background: var(--glass-bg);
  backdrop-filter: blur(40px);
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  box-shadow: 0 40px 100px rgba(0, 0, 0, 0.6);
}

/* Header */
.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14px 24px;
  border-bottom: 1px solid var(--glass-border);
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  font-family: inherit;
  transition: all 0.2s;
}

.back-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 16px;
}

.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: var(--text-secondary);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #22c55e;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.save-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: var(--accent);
  border: none;
  border-radius: 10px;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.save-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(var(--accent-rgb), 0.35);
}

.save-btn:active {
  transform: translateY(0);
}

/* Editor main */
.editor-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* Sidebar */
.editor-sidebar {
  width: min(220px, 25%);
  border-right: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
  background: rgba(0, 0, 0, 0.2);
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 16px 16px 12px;
  font-size: 11px;
  color: var(--text-secondary);
  letter-spacing: 0.5px;
  text-transform: uppercase;
  opacity: 0.6;
  border-bottom: 1px solid var(--glass-border);
}

.sidebar-header svg {
  opacity: 0.7;
}

.outline-list {
  flex: 1;
  overflow-y: auto;
  padding: 12px 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.outline-list::-webkit-scrollbar {
  width: 3px;
}

.outline-list::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 2px;
}

.outline-item {
  padding: 8px 12px;
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.15s;
  text-align: left;
  font-family: inherit;
  border: none;
  background: transparent;
  width: 100%;
}

.outline-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.outline-item.level-1 {
  font-weight: 600;
  color: var(--text-primary);
}

.outline-item.level-2 {
  font-size: 11px;
}

.outline-item.level-3 {
  font-size: 11px;
  opacity: 0.8;
}

.outline-empty {
  padding: 20px;
  text-align: center;
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.5;
}

/* Editor area */
.editor-scroll-area {
  flex: 1;
  overflow-y: auto;
  padding: clamp(16px, 3vw, 28px) clamp(20px, 5%, 60px);
  cursor: text;
  position: relative;
}

.editor-scroll-area::-webkit-scrollbar {
  width: 6px;
}

.editor-scroll-area::-webkit-scrollbar-track {
  background: transparent;
}

.editor-scroll-area::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 3px;
}
</style>