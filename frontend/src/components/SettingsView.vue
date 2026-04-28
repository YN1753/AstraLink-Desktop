<script setup>
import { ref } from 'vue'

const props = defineProps(['user', 'config'])
const emit = defineEmits(['close', 'update-user', 'update-config', 'update-avatar'])

const editedUser = ref({ ...props.user })
const editedConfig = ref({ ...props.config })

const themes = [
  { id: 'light', name: '现代极简', color: '#ffffff' },
  { id: 'dark', name: '深邃星空', color: '#111827' },
  { id: 'sepia', name: '纸质书写', color: '#f4ecd8' }
]

const fonts = ['Inter', 'Sarasa Fixed SC', 'JetBrains Mono', 'EB Garamond']

const save = () => {
  emit('update-user', editedUser.value)
  emit('update-config', editedConfig.value)
  emit('close')
}

function triggerAvatarUpload() {
  document.getElementById('avatar-input')?.click()
}

function onAvatarSelected(event) {
  const file = event.target.files?.[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (e) => {
    const base64 = e.target?.result
    if (typeof base64 === 'string') {
      emit('update-avatar', base64)
    }
  }
  reader.readAsDataURL(file)
}
</script>

<template>
  <div class="settings-overlay" @click.self="$emit('close')">
    <div class="settings-card glass-panel">
      <aside class="settings-sidebar">
        <h2>CONFIG / 设置</h2>
        <nav>
          <a href="#" class="active">个人资料</a>
          <a href="#">外观定制</a>
          <a href="#">关于项目</a>
        </nav>
        <div class="version-info">
          AstraLink v1.2.4<br>Build: Stardust
        </div>
      </aside>

      <main class="settings-content">
        <section id="profile">
          <h3>个人资料 / PROFILE</h3>
          <div class="avatar-section">
            <div class="avatar-preview">
              <img v-if="editedUser.avatar" :src="editedUser.avatar" class="avatar-img" />
              <div v-else class="avatar-placeholder">{{ editedUser.username?.[0] || 'U' }}</div>
            </div>
            <div class="avatar-actions">
              <button class="btn-avatar" @click="triggerAvatarUpload">更换头像</button>
              <span class="avatar-hint">支持 JPG、PNG、SVG 格式</span>
            </div>
            <input
              id="avatar-input"
              type="file"
              accept="image/jpeg,image/png,image/svg+xml"
              style="display: none"
              @change="onAvatarSelected"
            />
          </div>
          <div class="form-group">
            <label>用户名称 / USERNAME</label>
            <input v-model="editedUser.username" />
          </div>
          <div class="form-group">
            <label>个性签名 / MOTTO</label>
            <textarea v-model="editedUser.motto" rows="2"></textarea>
          </div>
        </section>

        <section id="appearance">
          <h3>外观定制 / APPEARANCE</h3>
          <div class="form-group">
            <label>视觉主题 / THEME</label>
            <div class="theme-grid">
              <div v-for="t in themes" :key="t.id"
                   class="theme-item"
                   :class="{ active: editedConfig.theme === t.id }"
                   @click="editedConfig.theme = t.id"
              >
                <div class="color-preview" :style="{ background: t.color }"></div>
                <span>{{ t.name }}</span>
              </div>
            </div>
          </div>
          <div class="form-group">
            <label>阅读字体 / FONT</label>
            <select v-model="editedConfig.font">
              <option v-for="f in fonts" :key="f" :value="f">{{ f }}</option>
            </select>
          </div>
        </section>

        <footer class="settings-footer">
          <button class="btn-cancel" @click="$emit('close')">取消 / CANCEL</button>
          <button class="btn-save" @click="save">应用保存 / SAVE</button>
        </footer>
      </main>
    </div>
  </div>
</template>

<style scoped>
.settings-overlay {
  position: fixed; inset: 0; background: rgba(0,0,0,0.5);
  backdrop-filter: blur(10px); display: flex; justify-content: center; align-items: center; z-index: 200;
}

.settings-card {
  width: 850px; height: 600px;
  display: flex; overflow: hidden;
  box-shadow: 0 50px 100px rgba(0,0,0,0.5);
  border-radius: 20px;
}

.settings-sidebar {
  width: 220px;
  background: rgba(0, 0, 0, 0.2);
  padding: 40px 20px;
  border-right: 1px solid var(--glass-border);
  display: flex; flex-direction: column;
}

.settings-sidebar h2 { font-size: 12px; color: var(--text-secondary); letter-spacing: 2px; margin-bottom: 30px; opacity: 0.6; }
.settings-sidebar nav a {
  display: block; padding: 12px 15px; color: var(--text-secondary); text-decoration: none;
  font-size: 14px; border-radius: 8px; margin-bottom: 5px; transition: 0.2s;
}
.settings-sidebar nav a.active { background: rgba(255,255,255,0.05); color: var(--accent); font-weight: 600; }

.settings-content { flex: 1; padding: 40px 50px; overflow-y: auto; color: var(--text-primary); }
.settings-content h3 { font-size: 18px; font-weight: 200; margin-bottom: 30px; border-bottom: 1px solid var(--glass-border); padding-bottom: 15px; }

.form-group { margin-bottom: 25px; }
.form-group label { display: block; font-size: 10px; color: var(--text-secondary); margin-bottom: 10px; letter-spacing: 1px; }

/* 输入框深色化 */
input, select, textarea {
  width: 100%; background: rgba(255,255,255,0.05); border: 1px solid var(--glass-border);
  border-radius: 10px; padding: 12px; color: #fff; outline: none; transition: 0.3s;
}
input:focus { border-color: var(--accent); background: rgba(255,255,255,0.08); }

.avatar-section {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-bottom: 30px;
  padding: 20px;
  background: rgba(255,255,255,0.02);
  border-radius: 12px;
  border: 1px solid var(--glass-border);
}

.avatar-preview {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  background: var(--text-primary);
  color: var(--bg-app);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 800;
}

.avatar-actions {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.btn-avatar {
  background: rgba(255,255,255,0.08);
  border: 1px solid var(--glass-border);
  color: var(--text-primary);
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  transition: 0.2s;
  width: fit-content;
}

.btn-avatar:hover {
  background: rgba(255,255,255,0.12);
  border-color: var(--accent);
}

.avatar-hint {
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.6;
}

.theme-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 15px; }
.theme-item {
  border: 1px solid var(--glass-border); padding: 12px; border-radius: 12px;
  cursor: pointer; text-align: center; transition: 0.3s; background: rgba(255,255,255,0.02);
}
.theme-item.active { border-color: var(--accent); background: rgba(59, 130, 246, 0.1); }
.color-preview { height: 40px; border-radius: 8px; margin-bottom: 10px; }

.settings-footer {
  margin-top: 40px; display: flex; justify-content: flex-end; gap: 15px;
}
.btn-save { background: #fff; color: #000; border: none; padding: 12px 30px; border-radius: 10px; font-weight: 700; cursor: pointer; }
.btn-save:hover { background: var(--accent); color: #fff; }
.btn-cancel { background: transparent; border: 1px solid var(--glass-border); color: var(--text-secondary); padding: 12px 25px; border-radius: 10px; cursor: pointer; }
</style>