<script setup>
import { ref, watch } from 'vue'

const props = defineProps(['user', 'config', 'themes'])
const emit = defineEmits(['close', 'update-user', 'update-config', 'update-avatar'])

const editedUser = ref({ ...props.user })
const activeTab = ref('profile')
const contentRef = ref(null)

const fonts = ['Inter', 'Sarasa Fixed SC', 'JetBrains Mono', 'EB Garamond']

const tabs = [
  { id: 'profile', name: '个人资料', icon: 'user' },
  { id: 'appearance', name: '外观定制', icon: 'palette' },
  { id: 'about', name: '关于', icon: 'info' },
]

watch(() => props.user, (newUser) => {
  editedUser.value = { ...newUser }
}, { deep: true })

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

function selectTab(tabId) {
  activeTab.value = tabId
  if (contentRef.value) {
    contentRef.value.scrollTop = 0
  }
}

function selectTheme(themeId) {
  emit('update-config', { ...props.config, theme: themeId })
}

function selectFont(fontId) {
  emit('update-config', { ...props.config, font: fontId })
}

function saveUser() {
  emit('update-user', editedUser.value)
}

const showAvatarPicker = ref(false)
const showProfileModal = ref(false)
</script>

<template>
  <div class="settings-overlay" @click.self="$emit('close')">
    <div class="settings-card">
      <!-- 左侧导航 -->
      <aside class="settings-sidebar">
        <div class="sidebar-header">
          <svg class="logo-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 3l8 4.5v9L12 21l-8-4.5v-9L12 3z"/>
            <path d="M12 12l8-4.5M12 12v9M12 12L4 7.5"/>
          </svg>
          <span class="logo-text">设置</span>
        </div>

        <nav class="tab-nav">
          <div
            v-for="tab in tabs"
            :key="tab.id"
            :class="['tab-btn', { active: activeTab === tab.id }]"
            @click="selectTab(tab.id)"
          >
            <svg v-if="tab.icon === 'user'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="8" r="4"/>
              <path d="M4 20c0-4 4-6 8-6s8 2 8 6"/>
            </svg>
            <svg v-else-if="tab.icon === 'palette'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="9"/>
              <circle cx="8" cy="10" r="1.5" fill="currentColor"/>
              <circle cx="15" cy="8" r="1.5" fill="currentColor"/>
              <circle cx="16" cy="14" r="1.5" fill="currentColor"/>
            </svg>
            <svg v-else-if="tab.icon === 'info'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="9"/>
              <path d="M12 16v-4M12 8h.01"/>
            </svg>
            <svg v-else-if="tab.icon === 'data'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/>
            </svg>
            <span class="tab-name">{{ tab.name }}</span>
            <svg v-if="activeTab === tab.id" class="tab-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 18l6-6-6-6"/>
            </svg>
          </div>
        </nav>

        <div class="sidebar-footer">
          <div class="version-badge">v1.0 Stardust</div>
        </div>
      </aside>

      <!-- 内容区 -->
      <main class="settings-content" ref="contentRef">
        <!-- 个人资料 -->
        <div v-show="activeTab === 'profile'" class="page">
          <h2 class="page-title">个人资料</h2>
          <p class="page-desc">管理你的个人信息和头像</p>

          <div class="profile-card">
            <div class="avatar-preview" @click="showAvatarPicker = true">
              <img v-if="editedUser.avatar" :src="editedUser.avatar" class="avatar-img" />
              <div v-else class="avatar-placeholder">{{ editedUser.username?.[0] || 'U' }}</div>
              <div class="avatar-overlay">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
                </svg>
              </div>
            </div>
            <div class="profile-info">
              <div class="profile-name-row">
                <span class="profile-name">{{ editedUser.username || '未设置用户名' }}</span>
                <button class="btn-edit-profile" @click="showProfileModal = true">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                    <path d="M18.5 2.5a2.12 2.12 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                  </svg>
                  修改资料
                </button>
              </div>
              <p class="profile-motto">{{ editedUser.motto || '暂无个性签名' }}</p>
            </div>
          </div>
        </div>

        <!-- 外观定制 -->
        <div v-show="activeTab === 'appearance'" class="page">
          <h2 class="page-title">外观定制</h2>
          <p class="page-desc">自定义应用的外观和系统字体</p>

          <div class="section-block">
            <label class="section-label">视觉主题</label>
            <div class="theme-grid">
              <div
                v-for="t in themes"
                :key="t.id"
                :class="['theme-item', { active: config.theme === t.id }]"
                @click="selectTheme(t.id)"
              >
                <div class="theme-preview" :style="{ background: t.bg }">
                  <div class="theme-preview-inner" :style="{ borderColor: t.accent }">
                    <div class="theme-accent" :style="{ background: t.accent }"></div>
                    <div class="theme-text" :style="{ background: t.text }"></div>
                  </div>
                </div>
                <span class="theme-name">{{ t.name }}</span>
                <div v-if="config.theme === t.id" class="theme-check">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                    <path d="M20 6L9 17l-5-5"/>
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <div class="section-block">
            <label class="section-label">系统字体</label>
            <div class="font-grid">
              <div
                v-for="f in fonts"
                :key="f"
                :class="['font-item', { active: config.font === f }]"
                @click="selectFont(f)"
              >
                <span class="font-preview" :style="{ fontFamily: f }">Aa</span>
                <span class="font-name">{{ f }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 关于 -->
        <div v-show="activeTab === 'about'" class="page">
          <h2 class="page-title">关于项目</h2>
          <p class="page-desc">了解 AstraLink 的更多信息</p>

          <div class="about-hero">
            <div class="about-logo">
              <svg viewBox="0 0 48 48" fill="none">
                <circle cx="24" cy="24" r="20" stroke="var(--accent)" stroke-width="1.5"/>
                <path d="M24 8l6 3.5v7L24 22l-6-3.5v-7L24 8z" fill="var(--accent)" opacity="0.6"/>
                <path d="M24 22l6 3.5v7L24 35.5l-6-3.5v-7L24 22z" fill="var(--accent)"/>
                <circle cx="24" cy="24" r="3" fill="var(--accent)"/>
              </svg>
            </div>
            <h3 class="about-title">AstraLink</h3>
            <p class="about-subtitle">双向链接笔记应用</p>
          </div>

          <div class="about-desc">
            <p>整合 Galaxy 星系视图与树形目录，让你的笔记如星辰般互联。</p>
            <p>记录是为了更好的思考，每一个想法都值得被连接。</p>
          </div>

          <div class="feature-list">
            <div class="feature-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="12" cy="12" r="9"/>
                <circle cx="12" cy="12" r="3"/>
                <path d="M12 3v3M12 18v3M3 12h3M18 12h3"/>
              </svg>
              <span>Galaxy 星系视图</span>
            </div>
            <div class="feature-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/>
                <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/>
              </svg>
              <span>双向链接</span>
            </div>
            <div class="feature-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="11" cy="11" r="7"/>
                <path d="M21 21l-4.35-4.35"/>
              </svg>
              <span>全文搜索</span>
            </div>
            <div class="feature-item">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="12" cy="12" r="9"/>
                <circle cx="8" cy="10" r="1.5" fill="currentColor"/>
                <circle cx="15" cy="8" r="1.5" fill="currentColor"/>
                <circle cx="16" cy="14" r="1.5" fill="currentColor"/>
              </svg>
              <span>多主题切换</span>
            </div>
          </div>
        </div>
      </main>

      <button class="close-btn" @click="$emit('close')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 6L6 18M6 6l12 12"/>
        </svg>
      </button>

      <!-- 头像预览弹窗 -->
      <Transition name="fade">
        <div v-if="showAvatarPicker" class="avatar-picker-overlay" @click.self="showAvatarPicker = false">
          <div class="avatar-picker-modal">
            <div class="avatar-picker-preview">
              <img v-if="editedUser.avatar" :src="editedUser.avatar" class="preview-img" />
              <div v-else class="preview-placeholder">{{ editedUser.username?.[0] || 'U' }}</div>
            </div>
            <div class="avatar-picker-actions">
              <button class="btn-change-avatar" @click="triggerAvatarUpload">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
                </svg>
                更换头像
              </button>
              <button class="btn-cancel-picker" @click="showAvatarPicker = false">取消</button>
            </div>
            <input
              id="avatar-input"
              type="file"
              accept="image/jpeg,image/png,image/svg+xml"
              style="display: none"
              @change="onAvatarSelected"
            />
          </div>
        </div>
      </Transition>

      <!-- 编辑个人资料弹窗 -->
      <Transition name="fade">
        <div v-if="showProfileModal" class="profile-modal-overlay" @click.self="showProfileModal = false">
          <div class="profile-modal">
            <div class="modal-header">
              <h3>编辑个人资料</h3>
              <button class="modal-close" @click="showProfileModal = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M18 6L6 18M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="modal-body">
              <div class="form-group">
                <label class="form-label">用户名称</label>
                <input v-model="editedUser.username" class="form-input" placeholder="输入用户名" />
              </div>
              <div class="form-group">
                <label class="form-label">个性签名</label>
                <textarea v-model="editedUser.motto" class="form-input" rows="3" placeholder="描述你自己..."></textarea>
              </div>
            </div>
            <div class="modal-footer">
              <button class="btn-cancel-modal" @click="showProfileModal = false">取消</button>
              <button class="btn-save-modal" @click="saveUser(); showProfileModal = false">保存</button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.settings-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(12px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 200;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.settings-card {
  width: min(720px, 92vw);
  height: min(540px, 85vh);
  display: flex;
  position: relative;
  border-radius: clamp(12px, 2vw, 20px);
  overflow: hidden;
  box-shadow: 0 25px 80px rgba(0, 0, 0, 0.5), 0 0 0 1px var(--glass-border);
  animation: slideUp 0.3s ease;
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

/* 左侧导航 */
.settings-sidebar {
  width: min(200px, 25%);
  background: linear-gradient(180deg, var(--glass-bg) 0%, rgba(0, 0, 0, 0.15) 100%);
  border-right: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
  padding: 24px 0;
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
  margin-bottom: 30px;
}

.logo-icon {
  width: 24px;
  height: 24px;
  color: var(--accent);
}

.logo-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 1px;
}

.tab-nav {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 0 12px;
  flex: 1;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 14px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-secondary);
  font-size: 13px;
  position: relative;
}

.tab-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.tab-btn.active {
  background: linear-gradient(90deg, rgba(var(--accent-rgb), 0.12) 0%, transparent 70%);
  color: var(--accent);
  font-weight: 500;
}

.tab-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.tab-name {
  flex: 1;
}

.tab-arrow {
  width: 14px;
  height: 14px;
  opacity: 0.5;
}

.sidebar-footer {
  padding: 0 20px;
}

.version-badge {
  font-size: 10px;
  color: var(--text-secondary);
  opacity: 0.5;
  letter-spacing: 0.5px;
}

/* 内容区 */
.settings-content {
  flex: 1;
  overflow-y: auto;
  background: var(--glass-bg);
  color: var(--text-primary);
}

.settings-content::-webkit-scrollbar {
  width: 4px;
}

.settings-content::-webkit-scrollbar-track {
  background: transparent;
}

.settings-content::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 2px;
}

.page {
  padding: clamp(16px, 3vw, 32px) clamp(20px, 4vw, 36px);
}

.page-title {
  font-size: clamp(16px, 2.5vw, 22px);
  font-weight: 600;
  margin: 0 0 6px 0;
  color: var(--text-primary);
}

.page-desc {
  font-size: clamp(11px, 1.5vw, 13px);
  color: var(--text-secondary);
  margin: 0 0 clamp(16px, 3vw, 28px) 0;
}

/* 个人资料卡片 */
.profile-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 16px;
  border: 1px solid var(--glass-border);
}

.avatar-preview {
  width: clamp(64px, 10vw, 80px);
  height: clamp(64px, 10vw, 80px);
  border-radius: 50%;
  overflow: hidden;
  flex-shrink: 0;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  cursor: pointer;
  position: relative;
  transition: transform 0.2s ease, box-shadow 0.2s ease;
}

.avatar-preview:hover {
  transform: scale(1.05);
  box-shadow: 0 8px 25px rgba(var(--accent-rgb), 0.3);
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s ease;
  border-radius: 50%;
}

.avatar-preview:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay svg {
  width: 24px;
  height: 24px;
  color: #fff;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, var(--accent) 0%, rgba(var(--accent-rgb), 0.7) 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: clamp(24px, 4vw, 32px);
  font-weight: 700;
}

.profile-info {
  flex: 1;
  min-width: 0;
}

.profile-name-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.profile-name {
  font-size: clamp(16px, 2.5vw, 20px);
  font-weight: 600;
  color: var(--text-primary);
}

.btn-edit-profile {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(var(--accent-rgb), 0.1);
  border: 1px solid var(--accent);
  color: var(--accent);
  padding: 8px 14px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.btn-edit-profile svg {
  width: 14px;
  height: 14px;
}

.btn-edit-profile:hover {
  background: rgba(var(--accent-rgb), 0.2);
}

.profile-motto {
  font-size: clamp(12px, 1.5vw, 14px);
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.5;
}

.hint {
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.6;
}

/* 表单卡片 */
.form-card {
  background: rgba(255, 255, 255, 0.02);
  border-radius: 14px;
  border: 1px solid var(--glass-border);
  padding: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.form-label svg {
  width: 14px;
  height: 14px;
  opacity: 0.7;
}

.form-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  padding: 12px 14px;
  color: var(--text-primary);
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: all 0.2s ease;
  box-sizing: border-box;
}

.form-input:focus {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

.form-input::placeholder {
  color: var(--text-secondary);
  opacity: 0.4;
}

textarea.form-input {
  resize: vertical;
  min-height: 80px;
}

.form-divider {
  height: 1px;
  background: var(--glass-border);
  margin: 16px 0;
}

/* 主题网格 */
.section-block {
  margin-bottom: 24px;
}

.section-label {
  display: block;
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 12px;
  font-weight: 500;
  letter-spacing: 0.5px;
}

.theme-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.theme-item {
  position: relative;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  padding: 12px;
  border-radius: 12px;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s ease;
}

.theme-item:hover {
  border-color: var(--accent);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}

.theme-item.active {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.08);
}

.theme-preview {
  height: 44px;
  border-radius: 6px;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px;
}

.theme-preview-inner {
  width: 100%;
  height: 100%;
  border-radius: 4px;
  border: 1.5px solid;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
}

.theme-accent {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.theme-text {
  width: 16px;
  height: 4px;
  border-radius: 2px;
  opacity: 0.8;
}

.theme-name {
  font-size: 11px;
  color: var(--text-secondary);
  font-weight: 500;
}

.theme-check {
  position: absolute;
  top: 8px;
  right: 8px;
  width: 18px;
  height: 18px;
  background: var(--accent);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.theme-check svg {
  width: 10px;
  height: 10px;
  color: #fff;
}

/* 字体网格 */
.font-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.font-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.font-item:hover {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

.font-item.active {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.1);
}

.font-preview {
  font-size: 22px;
  color: var(--text-primary);
  width: 36px;
  text-align: center;
}

.font-name {
  font-size: 12px;
  color: var(--text-secondary);
}

/* 关于页 */
.about-hero {
  text-align: center;
  padding: 30px 0;
  margin-bottom: 20px;
}

.about-logo {
  width: 72px;
  height: 72px;
  margin: 0 auto 16px;
}

.about-logo svg {
  width: 100%;
  height: 100%;
}

.about-title {
  font-size: 24px;
  font-weight: 700;
  margin: 0 0 6px 0;
  background: linear-gradient(135deg, var(--text-primary) 0%, var(--accent) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.about-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
}

.about-desc {
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-secondary);
  text-align: center;
  margin-bottom: 24px;
}

.about-desc p {
  margin: 0 0 8px 0;
}

.feature-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 10px;
  font-size: 13px;
  color: var(--text-primary);
  transition: all 0.2s ease;
}

.feature-item:hover {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

.feature-item svg {
  width: 18px;
  height: 18px;
  color: var(--accent);
  flex-shrink: 0;
}

/* 关闭按钮 */
.close-btn {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 8px;
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn svg {
  width: 16px;
  height: 16px;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

/* 编辑个人资料弹窗 */
.profile-modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 300;
}

.profile-modal {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  width: min(400px, 90vw);
  overflow: hidden;
  animation: scaleIn 0.2s ease;
}

@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--glass-border);
}

.modal-header h3 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.modal-close {
  width: 28px;
  height: 28px;
  border: none;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.modal-close svg {
  width: 14px;
  height: 14px;
}

.modal-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.modal-body {
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.modal-footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  padding: 16px 20px;
  border-top: 1px solid var(--glass-border);
}

.btn-cancel-modal {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--glass-border);
  color: var(--text-secondary);
  padding: 10px 20px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s ease;
}

.btn-cancel-modal:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.btn-save-modal {
  background: var(--accent);
  border: none;
  color: #fff;
  padding: 10px 24px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn-save-modal:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

/* 头像预览弹窗 */
.avatar-picker-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 300;
}

.avatar-picker-modal {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 20px;
  padding: 30px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  animation: scaleIn 0.2s ease;
}

@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

.avatar-picker-preview {
  width: 140px;
  height: 140px;
  border-radius: 50%;
  overflow: hidden;
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.4);
}

.preview-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.preview-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, var(--accent) 0%, rgba(var(--accent-rgb), 0.7) 100%);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  font-weight: 700;
}

.avatar-picker-actions {
  display: flex;
  gap: 12px;
}

.btn-change-avatar {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--accent);
  border: none;
  color: #fff;
  padding: 12px 24px;
  border-radius: 10px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn-change-avatar svg {
  width: 16px;
  height: 16px;
}

.btn-change-avatar:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-cancel-picker {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--glass-border);
  color: var(--text-secondary);
  padding: 12px 20px;
  border-radius: 10px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s ease;
}

.btn-cancel-picker:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

</style>