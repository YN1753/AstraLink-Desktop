<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { MilkdownProvider } from '@milkdown/vue'
import MilkdownCore from './MilkdownCore.vue'
import TagInput from './TagInput.vue'
import LinkPickerView from './LinkPickerView.vue'
import { GetNoteContent, CreateNote, UpdateNoteContent, UpdateNodeInfo, LinkTagToNode, GetTagsWithCount, GetTagsByNodeID, GetNodeByPath } from '../../wailsjs/go/main/App'

const props = defineProps(['status', 'noteId', 'user', 'newNoteRequest'])
const emit = defineEmits(['saved', 'back', 'refresh-tags', 'open-note'])

const openTabs = ref([])
const currentTabId = ref(null)
const tabCounter = ref(0)
const mounted = ref(false)
const saveToast = ref('')
let toastTimer = null

function showToast(msg, duration = 2000) {
  saveToast.value = msg
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { saveToast.value = '' }, duration)
}

function onKeyDown(e) {
  if ((e.ctrlKey || e.metaKey) && e.key === 's') {
    e.preventDefault()
    onSave()
  }
}

const currentTab = computed(() => openTabs.value.find(t => t.id === currentTabId.value))

async function openNote(note = null, forceReload = false) {
  if (!note) {
    const id = `new-${++tabCounter.value}`
    openTabs.value.push({ id, name: '未命名', content: '', isNew: true, dirty: false })
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
    existing.dirty = false
    currentTabId.value = note.id
  } else {
    openTabs.value.push({ id: note.id, name: title, content: content || note.content || '', isNew: false, dirty: false })
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
  window.addEventListener('keydown', onKeyDown)
  if (props.noteId) {
    await openNote({ id: props.noteId, name: '' })
  } else {
    openNote()
  }
  loadFileTree()
})

onUnmounted(() => {
  window.removeEventListener('keydown', onKeyDown)
  if (toastTimer) clearTimeout(toastTimer)
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
  currentTab.value.dirty = true
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
    currentTab.value.dirty = false

    const tags = noteTags.value[currentTab.value.id] || []
    if (tags.length > 0) {
      await linkTagsToNote(currentTab.value.id, tags)
    }

    emit('saved', { id: currentTab.value.id, name: currentTab.value.name })
    emit('refresh-tags')
    showToast('保存成功')
  } catch (e) {
    console.error('保存失败:', e)
    showToast('保存失败')
  }
}

function goHome() {
  openTabs.value = []
  currentTabId.value = null
  emit('back')
}

const noteTags = ref({})
const linkPickerPos = ref(null)
const editorCoreRef = ref(null)

function showLinkPicker(insertPos) {
  linkPickerPos.value = insertPos
}

function hideLinkPicker() {
  linkPickerPos.value = null
}

function handleLinkSelect(note) {
  if (linkPickerPos.value === null) return
  const editorRef = editorCoreRef.value
  if (!editorRef) return
  editorRef.insertLink(note.id, note.title)
  hideLinkPicker()
}

function handleOpenNoteFromLink(noteId) {
  emit('open-note', noteId)
}

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

// Sidebar mode: 'outline' or 'tree'
const sidebarMode = ref('outline')
const fileTree = ref([])
const expandedPaths = ref(new Set())

async function loadFileTree() {
  try {
    // GetNodeByPath('root') returns ALL nodes (everything starts with 'root')
    const allNodes = await GetNodeByPath('root')
    const nodes = (allNodes || []).filter(n => n.type === 'galaxy' || n.type === 'note')
    fileTree.value = buildFileTree(nodes)
  } catch (e) {
    console.error('Failed to load file tree:', e)
    fileTree.value = []
  }
}

function buildFileTree(nodes) {
  const root = { id: '__root__', name: '根目录', type: 'galaxy', path: 'root', children: [], depth: 0 }

  for (const node of nodes) {
    const pathParts = (node.path || 'root').split('/')
    // path is like "root/userID" or "root/userID/galaxyID"
    // Skip "root" and userID, the rest are galaxy IDs forming the hierarchy
    const hierarchyIds = pathParts.slice(2)
    insertIntoTree(root, hierarchyIds, node)
  }

  sortTree(root)
  return root.children
}

function insertIntoTree(parent, hierarchyIds, node) {
  if (hierarchyIds.length === 0) {
    // This node is a direct child of parent
    parent.children.push({
      id: node.id,
      name: node.name,
      type: node.type,
      path: node.path,
      children: node.type === 'galaxy' ? [] : undefined,
      depth: parent.depth + 1
    })
    return
  }

  // Find or create the intermediate galaxy
  const nextId = hierarchyIds[0]
  let intermediate = parent.children.find(c => c.id === nextId)
  if (!intermediate) {
    // The intermediate galaxy might not be in the nodes list (e.g., it's a parent galaxy)
    // We need to create a placeholder
    intermediate = {
      id: nextId,
      name: nextId, // Will be updated if found in nodes
      type: 'galaxy',
      path: parent.path + '/' + nextId,
      children: [],
      depth: parent.depth + 1
    }
    parent.children.push(intermediate)
  }
  insertIntoTree(intermediate, hierarchyIds.slice(1), node)
}

function sortTree(node) {
  if (!node.children) return
  node.children.sort((a, b) => {
    // Galaxies first, then notes
    if (a.type !== b.type) return a.type === 'galaxy' ? -1 : 1
    return a.name.localeCompare(b.name)
  })
  for (const child of node.children) {
    sortTree(child)
  }
}

function toggleExpand(nodeId) {
  if (expandedPaths.value.has(nodeId)) {
    expandedPaths.value.delete(nodeId)
  } else {
    expandedPaths.value.add(nodeId)
  }
}

function isExpanded(nodeId) {
  return expandedPaths.value.has(nodeId)
}

async function switchSidebarMode(mode) {
  sidebarMode.value = mode
  if (mode === 'tree' && fileTree.value.length === 0) {
    await loadFileTree()
  }
}

function openNoteFromTree(node) {
  if (node.type === 'note') {
    openNote({ id: node.id, name: node.name })
  }
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
          <span class="tab-name">{{ tab.name || '未命名' }}<span v-if="tab.dirty" class="tab-dirty"> *</span></span>
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
          <button class="save-btn" @click="onSave" title="Ctrl+S">
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
        <!-- Sidebar -->
        <aside class="editor-sidebar">
          <div class="sidebar-header">
            <div class="sidebar-tabs">
              <button
                :class="['sidebar-tab', { active: sidebarMode === 'outline' }]"
                @click="switchSidebarMode('outline')"
              >
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <line x1="8" y1="6" x2="21" y2="6"/>
                  <line x1="8" y1="12" x2="21" y2="12"/>
                  <line x1="8" y1="18" x2="21" y2="18"/>
                  <line x1="3" y1="6" x2="3.01" y2="6"/>
                  <line x1="3" y1="12" x2="3.01" y2="12"/>
                  <line x1="3" y1="18" x2="3.01" y2="18"/>
                </svg>
                大纲
              </button>
              <button
                :class="['sidebar-tab', { active: sidebarMode === 'tree' }]"
                @click="switchSidebarMode('tree')"
              >
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/>
                </svg>
                文件
              </button>
            </div>
          </div>

          <!-- Outline view -->
          <div v-if="sidebarMode === 'outline'" class="outline-list">
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

          <!-- File tree view -->
          <div v-else class="file-tree-list">
            <template v-for="node in fileTree" :key="node.id">
              <div
                :class="['tree-item', { active: currentTabId === node.id }]"
                :style="{ paddingLeft: `${(node.depth - 1) * 14 + 8}px` }"
                @click="node.type === 'galaxy' ? toggleExpand(node.id) : openNoteFromTree(node)"
              >
                <svg v-if="node.type === 'galaxy'" :class="['tree-arrow', { expanded: isExpanded(node.id) }]" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M9 18l6-6-6-6"/>
                </svg>
                <svg v-if="node.type === 'galaxy'" class="tree-icon galaxy-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/>
                </svg>
                <svg v-else class="tree-icon note-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                </svg>
                <span class="tree-name">{{ node.name }}</span>
              </div>
              <template v-if="node.type === 'galaxy' && isExpanded(node.id) && node.children">
                <div
                  v-for="child in node.children"
                  :key="child.id"
                  :class="['tree-item', 'tree-child', { active: currentTabId === child.id }]"
                  :style="{ paddingLeft: `${(child.depth - 1) * 14 + 8}px` }"
                  @click="child.type === 'galaxy' ? toggleExpand(child.id) : openNoteFromTree(child)"
                >
                  <svg v-if="child.type === 'galaxy'" :class="['tree-arrow', { expanded: isExpanded(child.id) }]" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M9 18l6-6-6-6"/>
                  </svg>
                  <svg v-if="child.type === 'galaxy'" class="tree-icon galaxy-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/>
                  </svg>
                  <svg v-else class="tree-icon note-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                    <polyline points="14 2 14 8 20 8"/>
                  </svg>
                  <span class="tree-name">{{ child.name }}</span>
                </div>
              </template>
            </template>
            <div v-if="fileTree.length === 0" class="outline-empty">暂无文件</div>
          </div>
        </aside>

        <!-- Editor content -->
        <div class="editor-scroll-area">
          <MilkdownProvider>
            <MilkdownCore
              ref="editorCoreRef"
              :key="currentTabId"
              v-model="currentContent"
              :noteId="currentTabId"
              @update:modelValue="onContentChange"
              @show-link-picker="showLinkPicker"
              @open-note="handleOpenNoteFromLink"
            />
          </MilkdownProvider>
        </div>
      </div>
    </div>

    <!-- Link Picker Modal -->
    <LinkPickerView
      v-if="linkPickerPos !== null"
      :insertPos="linkPickerPos"
      @close="hideLinkPicker"
      @select-note="handleLinkSelect"
    />

    <!-- Save Toast -->
    <Transition name="toast-fade">
      <div v-if="saveToast" class="save-toast" :class="{ error: saveToast === '保存失败' }">
        {{ saveToast }}
      </div>
    </Transition>
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

.tab-dirty {
  color: var(--accent);
  font-weight: 600;
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
  padding: 10px 10px 0;
  border-bottom: 1px solid var(--glass-border);
}

.sidebar-tabs {
  display: flex;
  gap: 2px;
  flex: 1;
}

.sidebar-tab {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 8px 10px;
  border-radius: 6px 6px 0 0;
  font-size: 11px;
  color: var(--text-secondary);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.15s;
  font-family: inherit;
  letter-spacing: 0.3px;
}

.sidebar-tab:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.04);
}

.sidebar-tab.active {
  color: var(--accent);
  background: rgba(var(--accent-rgb), 0.08);
  border-bottom: 1.5px solid var(--accent);
  margin-bottom: -1px;
}

.sidebar-tab svg {
  opacity: 0.7;
}

.sidebar-tab.active svg {
  opacity: 1;
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

/* File tree */
.file-tree-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px 4px;
}

.file-tree-list::-webkit-scrollbar {
  width: 3px;
}

.file-tree-list::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 2px;
}

.tree-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border-radius: 5px;
  cursor: pointer;
  transition: all 0.12s;
  font-size: 12px;
  color: var(--text-secondary);
  user-select: none;
}

.tree-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.tree-item.active {
  background: rgba(var(--accent-rgb), 0.1);
  color: var(--accent);
}

.tree-arrow {
  flex-shrink: 0;
  transition: transform 0.15s;
  opacity: 0.5;
}

.tree-arrow.expanded {
  transform: rotate(90deg);
}

.tree-icon {
  flex-shrink: 0;
  opacity: 0.6;
}

.galaxy-icon {
  color: var(--accent);
}

.note-icon {
  color: var(--text-secondary);
}

.tree-item.active .tree-icon {
  opacity: 1;
}

.tree-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
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

/* Save toast */
.save-toast {
  position: fixed;
  bottom: 40px;
  left: 50%;
  transform: translateX(-50%);
  padding: 10px 28px;
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.25);
  border-radius: 10px;
  font-size: 13px;
  font-weight: 500;
  backdrop-filter: blur(20px);
  z-index: 9999;
  pointer-events: none;
}

.save-toast.error {
  background: rgba(248, 113, 113, 0.15);
  color: #f87171;
  border-color: rgba(248, 113, 113, 0.25);
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: opacity 0.3s ease, transform 0.3s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(10px);
}
</style>