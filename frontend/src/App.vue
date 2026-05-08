<script setup>
import { ref, onMounted, watch, onUnmounted } from 'vue'
import Sidebar from './components/Sidebar.vue'
import HomeView from './components/HomeView.vue'
import EditorView from './components/EditorView.vue'
import GalaxyView from './components/GalaxyView.vue'
import SettingsView from './components/SettingsView.vue'
import SearchView from './components/SearchView.vue'
import { CheckUserNode, MergeUserInfo, GetAvatar, UploadAvatar, GetDataSpace, GetTagCount, GetNoteCount, GetGalaxyCount, GetTagsWithCount, GetRecentNotes } from '../wailsjs/go/main/App'

const currentView = ref('home')
const previousView = ref('home')
const showSettings = ref(false)
const showSearch = ref(false)
const systemStatus = ref('READY')
const editingNoteId = ref(null)
const newNoteRequest = ref(0)

const themes = [
  { id: 'nebula', name: '深邃星云', bg: '#050505', text: '#ffffff', accent: '#3b82f6' },
  { id: 'void', name: '星际虚空', bg: '#0a0010', text: '#e8e0f0', accent: '#9d4edd' },
  { id: 'dusk', name: '薄暮黄昏', bg: '#2d1f1f', text: '#f5e6d3', accent: '#e07a5f' },
  { id: 'cyberpunk', name: '赛博朋克', bg: '#0d0221', text: '#f0f0f0', accent: '#ff00ff' },
  { id: 'cloud', name: '星云微光', bg: '#f8fafc', text: '#1e293b', accent: '#6366f1' },
  { id: 'ancient', name: '古老星图', bg: '#f4ecd8', text: '#5c4b37', accent: '#8b5e3c' },
]

const config = ref({
  theme: localStorage.getItem('theme') || 'nebula',
  font: localStorage.getItem('font') || 'Inter'
})

const user = ref({
  id: localStorage.getItem('userId') || '',
  username: 'VOYAGER',
  motto: '书写有序的人生',
  avatar: ''
})

const stats = ref({
  noteCount: 0,
  galaxyCount: 0,
  tagCount: 0,
  totalSize: 0
})

const sidebarCollapsed = ref(localStorage.getItem('sidebarCollapsed') === 'true')
const tags = ref([])
const recentNotes = ref([])
const selectedTag = ref(null)

async function loadRecentNotes() {
  try {
    const notes = await GetRecentNotes(10)
    recentNotes.value = (notes || []).map(n => ({
      id: n.id,
      name: n.name,
      timestamp: n.update_time ? n.update_time * 1000 : Date.now()
    }))
  } catch (e) {}
}

function trackRecentNote(noteId, noteName) {
  loadRecentNotes()
}

function applyTheme(themeId) {
  document.documentElement.setAttribute('data-theme', themeId)
  localStorage.setItem('theme', themeId)
}

function applyFont(fontId) {
  document.body.style.fontFamily = `'${fontId}', sans-serif`
  localStorage.setItem('font', fontId)
}

watch(() => config.value.theme, (newTheme) => {
  applyTheme(newTheme)
})

watch(() => config.value.font, (newFont) => {
  applyFont(newFont)
})

async function init() {
  try {
    const userNode = await CheckUserNode()
    if (userNode) {
      user.value = {
        id: userNode.id,
        username: userNode.name,
        motto: userNode.others?.motto || '记录，是为了更好的思考',
        avatar: userNode.address || '',
        others: userNode.others || {}
      }

      localStorage.setItem('userId', userNode.id)

      if (userNode.others?.theme) {
        config.value.theme = userNode.others.theme
        applyTheme(userNode.others.theme)
      }
      if (userNode.others?.font) {
        config.value.font = userNode.others.font
      }

      if (user.value.id) {
        try {
          const avatar = await GetAvatar(user.value.id)
          if (avatar) {
            user.value.avatar = avatar
          }
        } catch (e) {}

        try {
          const [noteCount, galaxyCount, tagCount, totalSize] = await Promise.all([
            GetNoteCount(),
            GetGalaxyCount(),
            GetTagCount(),
            GetDataSpace()
          ])
          stats.value = { noteCount, galaxyCount, tagCount, totalSize }
        } catch (e) {}

        try {
          const tagList = await GetTagsWithCount()
          tags.value = tagList || []
        } catch (e) {}

        await loadRecentNotes()
      }
    }
  } catch (e) {}
}

function openNoteById(noteId, noteName) {
  trackRecentNote(noteId, noteName)
  editingNoteId.value = noteId
  if (currentView.value !== 'editor') {
    previousView.value = currentView.value
    currentView.value = 'editor'
  }
}

function openNoteFromGalaxy(note) {
  if (note && note.id) {
    trackRecentNote(note.id, note.name)
    editingNoteId.value = note.id
    previousView.value = 'galaxy'
    currentView.value = 'editor'
  }
}

function handleSidebarNewNote() {
  if (currentView.value === 'editor') {
    newNoteRequest.value++
  } else {
    editingNoteId.value = null
    previousView.value = 'home'
    currentView.value = 'editor'
  }
}

async function saveUserConfig() {
  if (!user.value.id) return
  try {
    await MergeUserInfo({
      id: user.value.id,
      name: user.value.username,
      type: 'user',
      file: '',
      parentPath: '',
      others: {
        ...user.value.others,
        theme: config.value.theme,
        font: config.value.font,
        motto: user.value.motto
      }
    })
  } catch (e) {}
}

async function handleUpdateUser(updatedUser) {
  user.value = updatedUser
  try {
    await MergeUserInfo({
      id: updatedUser.id,
      name: updatedUser.username,
      type: 'user',
      file: '',
      parentPath: '',
      others: {
        ...updatedUser.others,
        motto: updatedUser.motto
      }
    })
  } catch (e) {}
}

function handleEditorSaved(savedNote) {
  loadRecentNotes()
  refreshTags()
  Promise.all([
    GetNoteCount(),
    GetGalaxyCount(),
    GetTagCount(),
    GetDataSpace()
  ]).then(([noteCount, galaxyCount, tagCount, totalSize]) => {
    stats.value = { noteCount, galaxyCount, tagCount, totalSize }
  }).catch(() => {})
}

async function handleUpdateAvatar(base64Data) {
  if (!user.value.id) return
  try {
    const path = await UploadAvatar(user.value.id, base64Data)
    if (path) {
      const avatar = await GetAvatar(user.value.id)
      if (avatar) {
        user.value.avatar = avatar
      }
    }
  } catch (e) {}
}

function handleKeydown(e) {
  if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
    e.preventDefault()
    showSearch.value = true
  }
  if (e.key === 'Escape') {
    showSearch.value = false
    showSettings.value = false
  }
}

function handleTagClick(tag) {
  if (selectedTag.value === tag.id) {
    selectedTag.value = null
  } else {
    selectedTag.value = tag.id
  }
}

async function refreshTags() {
  try {
    const tagList = await GetTagsWithCount()
    tags.value = tagList || []
  } catch (e) {}
}

onMounted(() => {
  applyTheme(config.value.theme)
  applyFont(config.value.font)
  init()
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div class="app-shell" :class="`theme-${config.theme}`">
    <div class="global-bg"></div>
    <div class="bg-grain"></div>

    <Sidebar
        v-model="sidebarCollapsed"
        :user="user"
        :currentView="currentView"
        :tags="tags"
        :recentNotes="recentNotes"
        :selectedTag="selectedTag"
        @open-note="(note) => openNoteById(note.id, note.name)"
        @open-settings="showSettings = true"
        @go-home="currentView = 'home'"
        @tag-click="handleTagClick"
    />

    <main class="app-viewport">
      <HomeView
          v-if="currentView === 'home'"
          :user="user"
          :status="systemStatus"
          :stats="stats"
          :recentNotes="recentNotes"
          @start-note="editingNoteId = null; previousView = 'home'; currentView = 'editor'"
          @enter-galaxy="currentView = 'galaxy'"
          @open-search="showSearch = true"
          @open-note="openNoteById"
      />

      <GalaxyView
          v-else-if="currentView === 'galaxy'"
          :user="user"
          @back="currentView = 'home'"
          @open-note="openNoteFromGalaxy"
      />

      <EditorView
          v-else-if="currentView === 'editor'"
          :status="systemStatus"
          :noteId="editingNoteId"
          :user="user"
          :newNoteRequest="newNoteRequest"
          @back="editingNoteId = null; currentView = previousView"
          @saved="handleEditorSaved"
          @refresh-tags="refreshTags"
      />
    </main>

    <SettingsView
        v-if="showSettings"
        :user="user"
        :config="config"
        :themes="themes"
        @close="showSettings = false"
        @update-user="handleUpdateUser"
        @update-config="c => { config = c; saveUserConfig() }"
        @update-avatar="handleUpdateAvatar"
    />

    <SearchView
        v-if="showSearch"
        :user="user"
        @close="showSearch = false"
        @open-note="(note) => { showSearch = false; openNoteById(note.id, note.name) }"
    />
  </div>
</template>

<style>
:root {
  --transition-speed: 0.35s;
  --bg-app: #050505;
  --glass-bg: rgba(15, 15, 15, 0.75);
  --glass-border: rgba(255, 255, 255, 0.08);
  --text-primary: #ffffff;
  --text-secondary: #94a3b8;
  --accent: #3b82f6;
  --accent-rgb: 59, 130, 246;
  --bg-gradient: #0d1425;
}

[data-theme="nebula"] {
  --bg-app: #050505;
  --glass-bg: rgba(15, 15, 15, 0.75);
  --glass-border: rgba(255, 255, 255, 0.08);
  --text-primary: #ffffff;
  --text-secondary: #94a3b8;
  --accent: #3b82f6;
  --accent-rgb: 59, 130, 246;
  --bg-gradient: #0d1425;
}

[data-theme="void"] {
  --bg-app: #0a0010;
  --glass-bg: rgba(20, 5, 30, 0.85);
  --glass-border: rgba(157, 78, 221, 0.2);
  --text-primary: #e8e0f0;
  --text-secondary: #a89cb5;
  --accent: #9d4edd;
  --accent-rgb: 157, 78, 221;
  --bg-gradient: #050008;
}

[data-theme="cloud"] {
  --bg-app: #f8fafc;
  --glass-bg: rgba(255, 255, 255, 0.92);
  --glass-border: rgba(99, 102, 241, 0.15);
  --text-primary: #1e293b;
  --text-secondary: #64748b;
  --accent: #6366f1;
  --accent-rgb: 99, 102, 241;
  --bg-gradient: #e2e8f0;
}

[data-theme="ancient"] {
  --bg-app: #f4ecd8;
  --glass-bg: rgba(244, 236, 216, 0.92);
  --glass-border: rgba(92, 75, 55, 0.15);
  --text-primary: #5c4b37;
  --text-secondary: #8b7355;
  --accent: #8b5e3c;
  --accent-rgb: 139, 94, 60;
  --bg-gradient: #d4c4a8;
}

[data-theme="dusk"] {
  --bg-app: #2d1f1f;
  --glass-bg: rgba(45, 31, 31, 0.82);
  --glass-border: rgba(224, 122, 95, 0.15);
  --text-primary: #f5e6d3;
  --text-secondary: #b8a89a;
  --accent: #e07a5f;
  --accent-rgb: 224, 122, 95;
  --bg-gradient: #1a1212;
}

[data-theme="cyberpunk"] {
  --bg-app: #0d0221;
  --glass-bg: rgba(13, 2, 33, 0.88);
  --glass-border: rgba(255, 0, 255, 0.2);
  --text-primary: #f0f0f0;
  --text-secondary: #a0a0a0;
  --accent: #ff00ff;
  --accent-rgb: 255, 0, 255;
  --bg-gradient: #050010;
}

* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  overflow: hidden;
  background: var(--bg-app);
  color: var(--text-primary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.app-shell {
  display: flex;
  height: 100vh;
  width: 100vw;
  position: relative;
  transition: background var(--transition-speed);
}

.global-bg {
  position: absolute;
  inset: 0;
  z-index: -2;
  background: radial-gradient(ellipse at 50% 30%, var(--bg-gradient) 0%, var(--bg-app) 70%);
  transition: background var(--transition-speed);
}

.bg-grain {
  position: absolute;
  inset: 0;
  z-index: -1;
  opacity: 0.03;
  pointer-events: none;
  background-image: url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='noise'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.65' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23noise)'/%3E%3C/svg%3E");
}

.app-viewport {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.glass-panel {
  background: var(--glass-bg);
  backdrop-filter: blur(24px) saturate(180%);
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  border: 1px solid var(--glass-border);
}

/* Scrollbar */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.15);
}
</style>