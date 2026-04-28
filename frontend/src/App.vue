<script setup>
import { ref, onMounted } from 'vue'
import Sidebar from './components/Sidebar.vue'
import HomeView from './components/HomeView.vue'
import EditorView from './components/EditorView.vue'
import GalaxyView from './components/GalaxyView.vue'
import SettingsView from './components/SettingsView.vue'
import { GetNodeByType, CreateNewNode, UpdateAvatar, GetAvatar } from '../wailsjs/go/main/App'

const currentView = ref('home')
const showSettings = ref(false)
const systemStatus = ref('READY')

// 全局配置
const config = ref({ theme: 'dark', font: 'Inter' })
const user = ref({ id: '', username: 'VOYAGER', motto: '书写有序的人生', avatar: '' })

async function handleUpdateAvatar(base64Data) {
  if (!user.value.id) return
  try {
    const newAvatar = await UpdateAvatar(user.value.id, base64Data)
    user.value.avatar = newAvatar // 后端直接返回 base64 data URL
  } catch (e) {
    console.error('更新头像失败:', e)
  }
}

async function init() {
  try {
    const nodes = await GetNodeByType('user')
    if (nodes && nodes.length > 0) {
      user.value = {
        id: nodes[0].id,
        username: nodes[0].name,
        motto: nodes[0].others?.motto || '记录，是为了更好的思考',
        avatar: nodes[0].address || ''
      }
      // 获取头像的 base64 数据
      if (user.value.id) {
        try {
          const avatar = await GetAvatar(user.value.id)
          if (avatar) {
            user.value.avatar = avatar
          }
        } catch (e) {
          // 头像获取失败不影响其他功能
        }
      }
    }
  } catch (e) {}
}

async function handleSaveNote({ title, content }) {
  systemStatus.value = 'SAVING'
  try {
    await CreateNewNode({ id: "", name: title, type: "note", file: content, parentPath: "root", others: {} })
    systemStatus.value = 'READY'
    currentView.value = 'home'
  } catch (err) {}
}

onMounted(init)
</script>

<template>
  <div class="app-shell" :class="`theme-${config.theme}`">
    <div class="global-background"></div>

    <Sidebar
        :user="user"
        @open-note="currentView = 'editor'"
        @open-settings="showSettings = true"
    />

    <main class="app-viewport">
      <HomeView
          v-if="currentView === 'home'"
          :user="user"
          :status="systemStatus"
          @start-note="currentView = 'editor'"
          @enter-galaxy="currentView = 'galaxy'"
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
        @close="showSettings = false"
        @update-user="u => user = u"
        @update-config="c => config = c"
        @update-avatar="handleUpdateAvatar"
    />
  </div>
</template>

<style>
:root {
  --transition-speed: 0.3s;
}

.theme-dark {
  --bg-app: #050505;
  --glass-bg: rgba(15, 15, 15, 0.7);
  --glass-border: rgba(255, 255, 255, 0.08);
  --text-primary: #ffffff;
  --text-secondary: #94a3b8;
  --accent: #3b82f6;
}

.theme-light {
  --bg-app: #f3f4f6;
  --glass-bg: rgba(255, 255, 255, 0.7);
  --glass-border: rgba(0, 0, 0, 0.08);
  --text-primary: #111827;
  --text-secondary: #4b5563;
  --accent: #2563eb;
}

body {
  margin: 0;
  font-family: 'Inter', sans-serif;
  overflow: hidden;
  background: var(--bg-app);
  color: var(--text-primary);
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
  background: radial-gradient(circle at 50% 50%, #0d1425 0%, #050505 100%);
}

.app-viewport { flex: 1; position: relative; overflow: hidden; display: flex; align-items: center; justify-content: center; }

.glass-panel {
  background: var(--glass-bg);
  backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid var(--glass-border);
}
</style>