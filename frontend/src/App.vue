<script setup>
import { ref, onMounted, watch, onUnmounted } from 'vue'
import Sidebar from './components/Sidebar.vue'
import HomeView from './components/HomeView.vue'
import EditorView from './components/EditorView.vue'
import GalaxyView from './components/GalaxyView.vue'
import SettingsView from './components/SettingsView.vue'
import SearchView from './components/SearchView.vue'
import { GetNodeByType, CreateNewNode, UpdateAvatar, GetAvatar } from '../wailsjs/go/main/App'

const currentView = ref('home')
const showSettings = ref(false)
const showSearch = ref(false)
const systemStatus = ref('READY')

// 6种主题配置
const themes = [
  { id: 'nebula', name: '深邃星云', bg: '#050505', text: '#ffffff', accent: '#3b82f6' },
  { id: 'void', name: '星际虚空', bg: '#0a0010', text: '#e8e0f0', accent: '#9d4edd' },
  { id: 'dusk', name: '薄暮黄昏', bg: '#2d1f1f', text: '#f5e6d3', accent: '#e07a5f' },
  { id: 'cyberpunk', name: '赛博朋克', bg: '#0d0221', text: '#f0f0f0', accent: '#ff00ff' },
  { id: 'cloud', name: '星云微光', bg: '#f8fafc', text: '#1e293b', accent: '#6366f1' },
  { id: 'ancient', name: '古老星图', bg: '#f4ecd8', text: '#5c4b37', accent: '#8b5e3c' },
]

// 全局配置
const config = ref({
  theme: localStorage.getItem('theme') || 'nebula',
  font: localStorage.getItem('font') || 'Inter'
})

// 用户信息
const user = ref({ id: '', username: 'VOYAGER', motto: '书写有序的人生', avatar: '' })

// 存储统计
const stats = ref({
  noteCount: 0,
  galaxyCount: 0,
  tagCount: 0,
  totalSize: 0
})

// 应用主题到CSS变量
function applyTheme(themeId) {
  const theme = themes.find(t => t.id === themeId) || themes[0]
  const root = document.documentElement
  root.setAttribute('data-theme', themeId)
  root.style.setProperty('--bg-app', theme.bg)
  root.style.setProperty('--text-primary', theme.text)
  root.style.setProperty('--accent', theme.accent)
  localStorage.setItem('theme', themeId)
}

// 应用字体
function applyFont(fontId) {
  document.body.style.fontFamily = `'${fontId}', sans-serif`
  localStorage.setItem('font', fontId)
}

// 初始化主题和字体
onMounted(() => {
  applyTheme(config.value.theme)
  applyFont(config.value.font)
})

// 监听主题变化
watch(() => config.value.theme, (newTheme) => {
  applyTheme(newTheme)
})

// 监听字体变化
watch(() => config.value.font, (newFont) => {
  applyFont(newFont)
})

async function handleUpdateAvatar(base64Data) {
  if (!user.value.id) return
  try {
    await UpdateAvatar(user.value.id, base64Data)
    // 重新获取头像的 base64 数据
    const avatarData = await GetAvatar(user.value.id)
    if (avatarData) {
      user.value = { ...user.value, avatar: avatarData }
    }
  } catch (e) {
    console.error('更新头像失败:', e)
  }
}

async function init() {
  try {
    const nodes = await GetNodeByType('user')
    if (nodes && nodes.length > 0) {
      const userNode = nodes[0]
      user.value = {
        id: userNode.id,
        username: userNode.name,
        motto: userNode.others?.motto || '记录，是为了更好的思考',
        avatar: userNode.address || '',
        others: userNode.others || {}
      }

      // 从数据库读取主题和字体配置
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
      }
    }
    // 加载统计数据
    const allNodes = await GetNodeByType('note')
    if (allNodes) {
      stats.value.noteCount = allNodes.length
      stats.value.galaxyCount = new Set(allNodes.filter(n => n.others?.galaxy).size)
      stats.value.tagCount = new Set(allNodes.flatMap(n => n.others?.tags || [])).size
    }
  } catch (e) {}
}

async function handleSaveNote({ title, content }) {
  systemStatus.value = 'SAVING'
  try {
    await CreateNewNode({ id: "", name: title, type: "note", file: content, parentPath: "root", parentId: user.value.id, others: {} })
    systemStatus.value = 'READY'
    currentView.value = 'home'
    // 刷新统计
    init()
  } catch (err) {}
}

// 保存用户配置到数据库
async function saveUserConfig() {
  if (!user.value.id) return
  try {
    await CreateNewNode({
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
  } catch (e) {
    console.error('保存配置失败:', e)
  }
}

// 键盘快捷键
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

onMounted(() => {
  init()
  window.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
  <div class="app-shell" :class="`theme-${config.theme}`">
    <div class="global-background"></div>

    <Sidebar
        :user="user"
        :stats="stats"
        :currentView="currentView"
        :showSearch="showSearch"
        @open-note="currentView = 'editor'"
        @open-settings="showSettings = true"
        @open-search="showSearch = true"
        @go-home="currentView = 'home'"
    />

    <main class="app-viewport">
      <HomeView
          v-if="currentView === 'home'"
          :user="user"
          :status="systemStatus"
          @start-note="currentView = 'editor'"
          @enter-galaxy="currentView = 'galaxy'"
          @open-search="showSearch = true"
      />

      <GalaxyView
          v-else-if="currentView === 'galaxy'"
          :user="user"
          @back="currentView = 'home'"
      />

      <EditorView
          v-else-if="currentView === 'editor'"
          :status="systemStatus"
          @back="currentView = 'home'"
          @save="handleSaveNote"
      />
    </main>

    <SettingsView
        v-if="showSettings"
        :user="user"
        :config="config"
        :themes="themes"
        @close="showSettings = false"
        @update-user="u => user = u"
        @update-config="c => { config = c; saveUserConfig() }"
        @update-avatar="handleUpdateAvatar"
    />

    <SearchView
        v-if="showSearch"
        :user="user"
        @close="showSearch = false"
        @open-note="currentView = 'editor'"
    />
  </div>
</template>

<style>
:root {
  --transition-speed: 0.3s;
  /* 默认深色主题 */
  --bg-app: #050505;
  --glass-bg: rgba(15, 15, 15, 0.7);
  --glass-border: rgba(255, 255, 255, 0.08);
  --text-primary: #ffffff;
  --text-secondary: #94a3b8;
  --accent: #3b82f6;
  --accent-rgb: 59, 130, 246;
  --bg-gradient: #0d1425;
}

/* 各主题CSS变量 */
[data-theme="nebula"] {
  --bg-app: #050505;
  --glass-bg: rgba(15, 15, 15, 0.7);
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
  --glass-bg: rgba(255, 255, 255, 0.9);
  --glass-border: rgba(99, 102, 241, 0.15);
  --text-primary: #1e293b;
  --text-secondary: #64748b;
  --accent: #6366f1;
  --accent-rgb: 99, 102, 241;
  --bg-gradient: #e2e8f0;
}

[data-theme="ancient"] {
  --bg-app: #f4ecd8;
  --glass-bg: rgba(244, 236, 216, 0.9);
  --glass-border: rgba(92, 75, 55, 0.15);
  --text-primary: #5c4b37;
  --text-secondary: #8b7355;
  --accent: #8b5e3c;
  --accent-rgb: 139, 94, 60;
  --bg-gradient: #d4c4a8;
}

[data-theme="dusk"] {
  --bg-app: #2d1f1f;
  --glass-bg: rgba(45, 31, 31, 0.8);
  --glass-border: rgba(224, 122, 95, 0.15);
  --text-primary: #f5e6d3;
  --text-secondary: #b8a89a;
  --accent: #e07a5f;
  --accent-rgb: 224, 122, 95;
  --bg-gradient: #1a1212;
}

[data-theme="cyberpunk"] {
  --bg-app: #0d0221;
  --glass-bg: rgba(13, 2, 33, 0.85);
  --glass-border: rgba(255, 0, 255, 0.2);
  --text-primary: #f0f0f0;
  --text-secondary: #a0a0a0;
  --accent: #ff00ff;
  --accent-rgb: 255, 0, 255;
  --bg-gradient: #050010;
}

body {
  margin: 0;
  font-family: 'Inter', sans-serif;
  overflow: hidden;
  background: var(--bg-app);
  color: var(--text-primary);
  transition: background var(--transition-speed), color var(--transition-speed);
}

.app-shell {
  display: flex;
  height: 100vh;
  width: 100vw;
  position: relative;
  transition: background var(--transition-speed);
}

.global-background {
  position: absolute;
  inset: 0;
  z-index: -1;
  background: radial-gradient(circle at 50% 50%, var(--bg-gradient) 0%, var(--bg-app) 100%);
  transition: background var(--transition-speed);
}

.app-viewport { flex: 1; position: relative; overflow: hidden; display: flex; align-items: center; justify-content: center; }

.glass-panel {
  background: var(--glass-bg);
  backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid var(--glass-border);
}
</style>