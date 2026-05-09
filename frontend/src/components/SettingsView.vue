<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import QRCode from 'qrcode'
import { CheckUpdate, GetVersion } from '../../wailsjs/go/main/App'

const props = defineProps(['user', 'config', 'themes'])
const emit = defineEmits(['close', 'update-user', 'update-config', 'update-avatar'])

const editedUser = ref({ ...props.user })
const activeTab = ref('profile')
const contentRef = ref(null)

const sectionRefs = {}
let observer = null

const fonts = [
  { id: 'DM Sans', name: 'DM Sans', desc: '现代无衬线' },
  { id: 'LXGW WenKai', name: 'LXGW WenKai', desc: '霞鹜文楷' },
  { id: 'Noto Serif SC', name: 'Noto Serif SC', desc: '思源宋体' },
  { id: 'JetBrains Mono', name: 'JetBrains Mono', desc: '等宽代码' }
]

const tabs = [
  { id: 'profile', name: '个人资料', icon: 'user' },
  { id: 'appearance', name: '外观定制', icon: 'palette' },
  { id: 'galaxy', name: '星系设置', icon: 'orbit' },
  { id: 'about', name: '关于', icon: 'info' },
  { id: 'donate', name: '赞赏作者', icon: 'heart' },
]

const showDonateModal = ref(false)
const donateQrWx = ref('')
const donateQrAlipay = ref('')

const currentVersion = ref('')
const checkingUpdate = ref(false)
const showUpdateModal = ref(false)
const updateInfo = ref(null)
const updateToast = ref('')
const updateError = ref('')

async function loadVersion() {
  try {
    currentVersion.value = await GetVersion()
  } catch (e) {
    currentVersion.value = '1.0.0'
  }
}

async function handleCheckUpdate() {
  checkingUpdate.value = true
  updateError.value = ''
  updateToast.value = ''
  try {
    const info = await CheckUpdate()
    if (info.hasUpdate) {
      updateInfo.value = info
      showUpdateModal.value = true
    } else {
      updateToast.value = '已是最新版本'
      setTimeout(() => { updateToast.value = '' }, 2500)
    }
  } catch (e) {
    updateError.value = '检查更新失败，请检查网络'
    setTimeout(() => { updateError.value = '' }, 3000)
  } finally {
    checkingUpdate.value = false
  }
}

function formatSize(bytes) {
  if (!bytes) return ''
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(0) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

async function generateDonateQr() {
  const style = getComputedStyle(document.documentElement)
  const accent = style.getPropertyValue('--accent').trim() || '#3b82f6'
  const bg = style.getPropertyValue('--bg-app').trim() || '#050505'
  try {
    donateQrWx.value = await QRCode.toDataURL('wxp://f2f07HN77AF5DnnfXWofPd-3BAcybco-S-yaqxmPi6udHQs', {
      width: 180, margin: 2, color: { dark: accent, light: bg }
    })
    donateQrAlipay.value = await QRCode.toDataURL('https://qr.alipay.com/2m615984gq3t4kgsifs70fc', {
      width: 180, margin: 2, color: { dark: accent, light: bg }
    })
  } catch (e) {
    console.error('QR generation failed:', e)
  }
}

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

function setSectionRef(tabId, el) {
  if (el) sectionRefs[tabId] = el
}

function selectTab(tabId) {
  activeTab.value = tabId
  const el = sectionRefs[tabId]
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

onMounted(() => {
  loadVersion()
  const container = contentRef.value
  if (!container) return
  observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (entry.isIntersecting && entry.intersectionRatio >= 0.3) {
        const id = entry.target.dataset.section
        if (id) activeTab.value = id
      }
    }
  }, { root: container, threshold: 0.3 })
  for (const el of Object.values(sectionRefs)) {
    observer.observe(el)
  }
})

onUnmounted(() => {
  if (observer) observer.disconnect()
})

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
      <!-- Left nav -->
      <aside class="settings-sidebar">
        <div class="sidebar-header">
          <div class="logo-mark">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
              <path d="M12 2L2 7l10 5 10-5-10-5z" fill="var(--accent)" opacity="0.8"/>
              <path d="M2 17l10 5 10-5" stroke="var(--accent)" stroke-width="2" stroke-linecap="round"/>
              <path d="M2 12l10 5 10-5" stroke="var(--accent)" stroke-width="2" stroke-linecap="round" opacity="0.6"/>
            </svg>
          </div>
          <span class="logo-text">设置</span>
        </div>

        <nav class="tab-nav">
          <button
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
            <svg v-else-if="tab.icon === 'orbit'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="3"/>
              <ellipse cx="12" cy="12" rx="10" ry="4"/>
              <ellipse cx="12" cy="12" rx="10" ry="4" transform="rotate(60 12 12)"/>
              <ellipse cx="12" cy="12" rx="10" ry="4" transform="rotate(120 12 12)"/>
            </svg>
            <svg v-else-if="tab.icon === 'info'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="9"/>
              <path d="M12 16v-4M12 8h.01"/>
            </svg>
            <svg v-else-if="tab.icon === 'heart'" class="tab-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <path d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z"/>
            </svg>
            <span class="tab-name">{{ tab.name }}</span>
            <svg v-if="activeTab === tab.id" class="tab-indicator" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 18l6-6-6-6"/>
            </svg>
          </button>
        </nav>

        <div class="sidebar-footer">
          <div class="version-tag">v{{ currentVersion || '1.0.0' }}</div>
        </div>
      </aside>

      <!-- Content -->
      <main class="settings-content" ref="contentRef">
        <!-- Profile -->
        <div :ref="el => setSectionRef('profile', el)" data-section="profile" class="page">
          <div class="page-header">
            <h2 class="page-title">个人资料</h2>
            <p class="page-desc">管理你的个人信息和头像</p>
          </div>

          <div class="profile-card">
            <button class="avatar-section" @click="showAvatarPicker = true">
              <div class="avatar-wrapper">
                <img v-if="editedUser.avatar" :src="editedUser.avatar" class="avatar-img" />
                <div v-else class="avatar-placeholder">
                  <span>{{ editedUser.username?.[0] || 'U' }}</span>
                </div>
                <div class="avatar-overlay">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
                  </svg>
                </div>
              </div>
            </button>
            <div class="profile-details">
              <div class="profile-name-row">
                <h3 class="profile-name">{{ editedUser.username || '未设置用户名' }}</h3>
                <button class="edit-btn" @click="showProfileModal = true">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                    <path d="M18.5 2.5a2.12 2.12 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                  </svg>
                  编辑资料
                </button>
              </div>
              <p class="profile-motto">{{ editedUser.motto || '暂无个性签名' }}</p>
            </div>
          </div>
        </div>

        <!-- Appearance -->
        <div :ref="el => setSectionRef('appearance', el)" data-section="appearance" class="page">
          <div class="page-header">
            <h2 class="page-title">外观定制</h2>
            <p class="page-desc">自定义应用的外观和字体</p>
          </div>

          <div class="settings-section">
            <label class="section-label">视觉主题</label>
            <div class="theme-grid">
              <button
                v-for="t in themes"
                :key="t.id"
                :class="['theme-card', { active: config.theme === t.id }]"
                @click="selectTheme(t.id)"
              >
                <div class="theme-preview" :style="{ background: t.bg }">
                  <div class="theme-inner" :style="{ borderColor: t.accent }">
                    <div class="theme-accent-dot" :style="{ background: t.accent }"></div>
                    <div class="theme-text-bar" :style="{ background: t.text }"></div>
                  </div>
                </div>
                <span class="theme-name">{{ t.name }}</span>
                <div v-if="config.theme === t.id" class="theme-check">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                    <path d="M20 6L9 17l-5-5"/>
                  </svg>
                </div>
              </button>
            </div>
          </div>

          <div class="settings-section">
            <label class="section-label">系统字体</label>
            <div class="font-grid">
              <button
                v-for="f in fonts"
                :key="f.id"
                :class="['font-card', { active: config.font === f.id }]"
                @click="selectFont(f.id)"
              >
                <span class="font-preview" :style="{ fontFamily: f.id }">Aa</span>
                <div class="font-info">
                  <span class="font-name">{{ f.name }}</span>
                  <span class="font-desc">{{ f.desc }}</span>
                </div>
              </button>
            </div>
          </div>
        </div>

        <!-- Galaxy -->
        <div :ref="el => setSectionRef('galaxy', el)" data-section="galaxy" class="page">
          <div class="page-header">
            <h2 class="page-title">星系设置</h2>
            <p class="page-desc">自定义星系图的节点和连线外观</p>
          </div>

          <div class="settings-section">
            <label class="section-label">节点大小</label>
            <div class="slider-row">
              <input
                type="range"
                class="galaxy-slider"
                :value="config.nodeScale ?? 1"
                min="0.6"
                max="1.6"
                step="0.1"
                @input="emit('update-config', { ...config, nodeScale: parseFloat($event.target.value) })"
              />
              <span class="slider-value">{{ ((config.nodeScale ?? 1) * 100).toFixed(0) }}%</span>
            </div>
          </div>

          <div class="settings-section">
            <label class="section-label">连线样式</label>
            <div class="link-style-grid">
              <button
                :class="['style-card', { active: (config.linkStyle ?? 'dashed') === 'solid' }]"
                @click="emit('update-config', { ...config, linkStyle: 'solid' })"
              >
                <svg width="48" height="12" viewBox="0 0 48 12">
                  <line x1="0" y1="6" x2="48" y2="6" stroke="currentColor" stroke-width="2"/>
                </svg>
                <span class="style-name">实线</span>
              </button>
              <button
                :class="['style-card', { active: (config.linkStyle ?? 'dashed') === 'dashed' }]"
                @click="emit('update-config', { ...config, linkStyle: 'dashed' })"
              >
                <svg width="48" height="12" viewBox="0 0 48 12">
                  <line x1="0" y1="6" x2="48" y2="6" stroke="currentColor" stroke-width="2" stroke-dasharray="6 4"/>
                </svg>
                <span class="style-name">虚线</span>
              </button>
            </div>
          </div>

          <div class="settings-section">
            <label class="section-label">连线粗细</label>
            <div class="slider-row">
              <input
                type="range"
                class="galaxy-slider"
                :value="config.linkWidth ?? 2"
                min="1"
                max="4"
                step="0.5"
                @input="emit('update-config', { ...config, linkWidth: parseFloat($event.target.value) })"
              />
              <span class="slider-value">{{ (config.linkWidth ?? 2).toFixed(1) }}px</span>
            </div>
          </div>
        </div>

        <!-- About -->
        <div :ref="el => setSectionRef('about', el)" data-section="about" class="page">
          <div class="page-header">
            <h2 class="page-title">关于项目</h2>
            <p class="page-desc">了解 AstraLink 的更多信息</p>
          </div>

          <div class="about-hero">
            <h3 class="about-title">AstraLink</h3>
            <p class="about-subtitle">双向链接笔记应用</p>
          </div>

          <div class="about-desc">
            <p>整合 Galaxy 星系视图与树形目录，让你的笔记如星辰般互联。</p>
            <p>记录是为了更好的思考，每一个想法都值得被连接。</p>
          </div>

          <div class="feature-list">
            <div class="feature-card">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="12" cy="12" r="9"/>
                <circle cx="12" cy="12" r="3"/>
                <path d="M12 3v3M12 18v3M3 12h3M18 12h3"/>
              </svg>
              <span>Galaxy 星系视图</span>
            </div>
            <div class="feature-card">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/>
                <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/>
              </svg>
              <span>双向链接</span>
            </div>
            <div class="feature-card">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="11" cy="11" r="7"/>
                <path d="M21 21l-4.35-4.35"/>
              </svg>
              <span>全文搜索</span>
            </div>
            <div class="feature-card">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <circle cx="12" cy="12" r="9"/>
                <circle cx="8" cy="10" r="1.5" fill="currentColor"/>
                <circle cx="15" cy="8" r="1.5" fill="currentColor"/>
                <circle cx="16" cy="14" r="1.5" fill="currentColor"/>
              </svg>
              <span>多主题切换</span>
            </div>
          </div>

          <div class="version-info-card">
            <div class="version-info-row">
              <span class="version-label">当前版本</span>
              <span class="version-value">v{{ currentVersion || '1.0.0' }}</span>
            </div>
            <button
              class="check-update-btn"
              :class="{ checking: checkingUpdate }"
              :disabled="checkingUpdate"
              @click="handleCheckUpdate"
            >
              <svg v-if="!checkingUpdate" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M21 12a9 9 0 11-6.22-8.56"/>
                <path d="M21 3v6h-6"/>
              </svg>
              <div v-else class="btn-spinner"></div>
              <span>{{ checkingUpdate ? '检查中...' : '检查更新' }}</span>
            </button>
            <Transition name="toast-fade">
              <div v-if="updateToast" class="update-toast success">{{ updateToast }}</div>
            </Transition>
            <Transition name="toast-fade">
              <div v-if="updateError" class="update-toast error">{{ updateError }}</div>
            </Transition>
          </div>
        </div>

        <!-- Donate -->
        <div :ref="el => setSectionRef('donate', el)" data-section="donate" class="page">
          <div class="page-header">
            <h2 class="page-title">赞赏作者</h2>
            <p class="page-desc">如果 AstraLink 对你有帮助，欢迎请作者喝杯咖啡</p>
          </div>

          <button class="donate-card" @click="showDonateModal = true; generateDonateQr()">
            <div class="donate-icon">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 00-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 000-7.78z"/>
              </svg>
            </div>
            <div class="donate-text">
              <span class="donate-title">支持开发</span>
              <span class="donate-desc">微信支付 / 支付宝</span>
            </div>
            <svg class="donate-arrow" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 18l6-6-6-6"/>
            </svg>
          </button>
        </div>
      </main>

      <button class="close-btn" @click="$emit('close')">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M18 6L6 18M6 6l12 12"/>
        </svg>
      </button>

      <!-- Avatar picker -->
      <Transition name="modal-fade">
        <div v-if="showAvatarPicker" class="modal-overlay" @click.self="showAvatarPicker = false">
          <div class="avatar-picker-modal">
            <div class="picker-preview">
              <img v-if="editedUser.avatar" :src="editedUser.avatar" class="preview-img" />
              <div v-else class="preview-placeholder">{{ editedUser.username?.[0] || 'U' }}</div>
            </div>
            <div class="picker-actions">
              <button class="picker-btn primary" @click="triggerAvatarUpload">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M17 8l-5-5-5 5M12 3v12"/>
                </svg>
                更换头像
              </button>
              <button class="picker-btn" @click="showAvatarPicker = false">取消</button>
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

      <!-- Profile edit modal -->
      <Transition name="modal-fade">
        <div v-if="showProfileModal" class="modal-overlay" @click.self="showProfileModal = false">
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
              <button class="modal-btn cancel" @click="showProfileModal = false">取消</button>
              <button class="modal-btn save" @click="saveUser(); showProfileModal = false">保存</button>
            </div>
          </div>
        </div>
      </Transition>

      <!-- Donate modal -->
      <Transition name="modal-fade">
        <div v-if="showDonateModal" class="modal-overlay" @click.self="showDonateModal = false">
          <div class="donate-modal">
            <div class="modal-header">
              <h3>赞赏作者</h3>
              <button class="modal-close" @click="showDonateModal = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M18 6L6 18M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="donate-body">
              <p class="donate-thanks">感谢你的支持，这将鼓励我持续改进 AstraLink</p>
              <div class="donate-qr-grid">
                <div class="donate-qr-card">
                  <div class="qr-img-wrapper">
                    <img v-if="donateQrWx" :src="donateQrWx" class="qr-img" alt="微信支付" />
                    <div v-else class="qr-placeholder">
                      <div class="qr-loading-spinner"></div>
                    </div>
                  </div>
                  <span class="qr-label">微信支付</span>
                </div>
                <div class="donate-qr-card">
                  <div class="qr-img-wrapper">
                    <img v-if="donateQrAlipay" :src="donateQrAlipay" class="qr-img" alt="支付宝" />
                    <div v-else class="qr-placeholder">
                      <div class="qr-loading-spinner"></div>
                    </div>
                  </div>
                  <span class="qr-label">支付宝</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Transition>

      <!-- Update modal -->
      <Transition name="modal-fade">
        <div v-if="showUpdateModal && updateInfo" class="modal-overlay" @click.self="showUpdateModal = false">
          <div class="update-modal">
            <div class="modal-header">
              <h3>发现新版本</h3>
              <button class="modal-close" @click="showUpdateModal = false">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M18 6L6 18M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="update-body">
              <div class="update-version-badge">
                <span class="update-from">v{{ updateInfo.currentVer }}</span>
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="20" height="20">
                  <path d="M5 12h14M12 5l7 7-7 7"/>
                </svg>
                <span class="update-to">v{{ updateInfo.latestVer }}</span>
              </div>
              <div v-if="updateInfo.releaseNote" class="update-notes">
                <h4>更新日志</h4>
                <pre>{{ updateInfo.releaseNote }}</pre>
              </div>
              <div class="update-downloads">
                <button
                  v-if="!updateInfo.isPortable && updateInfo.exeUrl"
                  class="download-btn"
                  @click="window.open(updateInfo.exeUrl)"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/>
                  </svg>
                  <div class="download-text">
                    <span class="download-name">下载安装版 (EXE)</span>
                    <span class="download-size" v-if="updateInfo.exeSize">{{ formatSize(updateInfo.exeSize) }}</span>
                  </div>
                </button>
                <button
                  v-if="updateInfo.isPortable && updateInfo.zipUrl"
                  class="download-btn"
                  @click="window.open(updateInfo.zipUrl)"
                >
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/>
                  </svg>
                  <div class="download-text">
                    <span class="download-name">下载便携版 (ZIP)</span>
                    <span class="download-size" v-if="updateInfo.zipSize">{{ formatSize(updateInfo.zipSize) }}</span>
                  </div>
                </button>
              </div>
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
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(16px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 200;
}

.settings-card {
  width: min(760px, 94vw);
  height: min(560px, 88vh);
  display: flex;
  position: relative;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 50px 120px rgba(0, 0, 0, 0.5), 0 0 0 1px var(--glass-border);
  animation: slideUp 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

@keyframes slideUp {
  from { opacity: 0; transform: translateY(24px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Sidebar */
.settings-sidebar {
  width: min(200px, 24%);
  background: linear-gradient(180deg, rgba(0,0,0,0.3) 0%, rgba(0,0,0,0.15) 100%);
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
  margin-bottom: 32px;
}

.logo-mark {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--accent-rgb), 0.1);
  border-radius: 10px;
}

.logo-text {
  font-size: 15px;
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
  gap: 12px;
  padding: 12px 14px;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  color: var(--text-secondary);
  font-size: 13px;
  border: none;
  background: transparent;
  font-family: inherit;
  text-align: left;
  width: 100%;
}

.tab-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.tab-btn.active {
  background: linear-gradient(90deg, rgba(var(--accent-rgb), 0.12) 0%, transparent 70%);
  color: var(--accent);
}

.tab-icon {
  width: 18px;
  height: 18px;
  flex-shrink: 0;
}

.tab-name {
  flex: 1;
}

.tab-indicator {
  width: 14px;
  height: 14px;
  opacity: 0.5;
}

.sidebar-footer {
  padding: 0 20px;
}

.version-tag {
  font-size: 10px;
  color: var(--text-secondary);
  opacity: 0.4;
  letter-spacing: 0.5px;
}

/* Content */
.settings-content {
  flex: 1;
  overflow-y: auto;
  background: var(--glass-bg);
  scroll-behavior: smooth;
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
  padding: clamp(20px, 4vw, 36px) clamp(24px, 5vw, 40px);
  scroll-margin-top: 0;
}

.page + .page {
  border-top: 1px solid var(--glass-border);
}

.page:last-child {
  padding-bottom: clamp(32px, 6vw, 56px);
}

.page-header {
  margin-bottom: 28px;
}

.page-title {
  font-size: clamp(20px, 3vw, 26px);
  font-weight: 700;
  margin: 0 0 8px 0;
  color: var(--text-primary);
}

.page-desc {
  font-size: clamp(12px, 1.5vw, 14px);
  color: var(--text-secondary);
  margin: 0;
}

/* Profile card */
.profile-card {
  display: flex;
  align-items: center;
  gap: 24px;
  padding: 28px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--glass-border);
  border-radius: 20px;
}

.avatar-section {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.avatar-wrapper {
  position: relative;
  width: clamp(72px, 12vw, 96px);
  height: clamp(72px, 12vw, 96px);
}

.avatar-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 50%;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.3);
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  background: linear-gradient(135deg, var(--accent) 0%, rgba(var(--accent-rgb), 0.7) 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 24px rgba(var(--accent-rgb), 0.3);
}

.avatar-placeholder span {
  font-size: clamp(28px, 5vw, 36px);
  font-weight: 700;
  color: #fff;
}

.avatar-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.avatar-section:hover .avatar-overlay {
  opacity: 1;
}

.avatar-overlay svg {
  width: 24px;
  height: 24px;
  color: #fff;
}

.profile-details {
  flex: 1;
  min-width: 0;
}

.profile-name-row {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 8px;
}

.profile-name {
  font-size: clamp(18px, 3vw, 22px);
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.edit-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: rgba(var(--accent-rgb), 0.1);
  border: 1px solid var(--accent);
  color: var(--accent);
  border-radius: 10px;
  cursor: pointer;
  font-size: 12px;
  font-weight: 500;
  transition: all 0.2s ease;
  font-family: inherit;
}

.edit-btn svg {
  width: 14px;
  height: 14px;
}

.edit-btn:hover {
  background: rgba(var(--accent-rgb), 0.2);
}

.profile-motto {
  font-size: clamp(13px, 1.8vw, 15px);
  color: var(--text-secondary);
  margin: 0;
  line-height: 1.5;
}

/* Settings section */
.settings-section {
  margin-bottom: 32px;
}

.section-label {
  display: block;
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 16px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.theme-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}

.theme-card {
  position: relative;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  padding: 14px;
  border-radius: 14px;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s ease;
  font-family: inherit;
  color: inherit;
}

.theme-card:hover {
  border-color: rgba(var(--accent-rgb), 0.4);
  transform: translateY(-2px);
}

.theme-card.active {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.08);
}

.theme-preview {
  height: 48px;
  border-radius: 8px;
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 6px;
}

.theme-inner {
  width: 100%;
  height: 100%;
  border-radius: 5px;
  border: 1.5px solid;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
}

.theme-accent-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.theme-text-bar {
  width: 20px;
  height: 5px;
  border-radius: 3px;
  opacity: 0.8;
}

.theme-name {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.theme-check {
  position: absolute;
  top: 10px;
  right: 10px;
  width: 20px;
  height: 20px;
  background: var(--accent);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.theme-check svg {
  width: 11px;
  height: 11px;
  color: #fff;
}

.font-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.font-card {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  color: inherit;
}

.font-card:hover {
  border-color: rgba(var(--accent-rgb), 0.4);
  background: rgba(var(--accent-rgb), 0.05);
}

.font-card.active {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.1);
}

.font-preview {
  font-size: 24px;
  color: var(--text-primary);
  width: 40px;
  text-align: center;
  flex-shrink: 0;
}

.font-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.font-name {
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
}

.font-desc {
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.7;
}

/* About */
.about-hero {
  text-align: center;
  padding: 32px 0;
  margin-bottom: 24px;
}

.about-title {
  font-size: 28px;
  font-weight: 800;
  margin: 0 0 8px 0;
  background: linear-gradient(135deg, var(--text-primary) 0%, var(--accent) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.about-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

.about-desc {
  font-size: 14px;
  line-height: 1.8;
  color: var(--text-secondary);
  text-align: center;
  margin-bottom: 28px;
}

.about-desc p {
  margin: 0 0 10px 0;
}

.feature-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.feature-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  font-size: 13px;
  color: var(--text-primary);
  transition: all 0.2s ease;
}

.feature-card:hover {
  border-color: rgba(var(--accent-rgb), 0.3);
  background: rgba(var(--accent-rgb), 0.03);
}

.feature-card svg {
  width: 20px;
  height: 20px;
  color: var(--accent);
  flex-shrink: 0;
}

/* Close button */
.close-btn {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 36px;
  height: 36px;
  border: none;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 10px;
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

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 300;
}

.avatar-picker-modal {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 24px;
  padding: 36px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 24px;
  animation: scaleIn 0.25s ease;
}

@keyframes scaleIn {
  from { opacity: 0; transform: scale(0.9); }
  to { opacity: 1; transform: scale(1); }
}

.picker-preview {
  width: 140px;
  height: 140px;
  border-radius: 50%;
  overflow: hidden;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
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
  font-size: 52px;
  font-weight: 700;
}

.picker-actions {
  display: flex;
  gap: 12px;
}

.picker-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 14px 24px;
  border-radius: 12px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  font-family: inherit;
}

.picker-btn.primary {
  background: var(--accent);
  border: none;
  color: #fff;
}

.picker-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(var(--accent-rgb), 0.35);
}

.picker-btn:not(.primary) {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--glass-border);
  color: var(--text-secondary);
}

.picker-btn:not(.primary):hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.picker-btn svg {
  width: 16px;
  height: 16px;
}

/* Profile modal */
.profile-modal {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 20px;
  width: min(420px, 90vw);
  overflow: hidden;
  animation: scaleIn 0.25s ease;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 24px;
  border-bottom: 1px solid var(--glass-border);
}

.modal-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.modal-close {
  width: 30px;
  height: 30px;
  border: none;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: 8px;
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
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.form-label {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 500;
}

.form-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--glass-border);
  border-radius: 10px;
  padding: 14px 16px;
  color: var(--text-primary);
  font-size: 14px;
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
  min-height: 90px;
}

.modal-footer {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  padding: 18px 24px;
  border-top: 1px solid var(--glass-border);
}

.modal-btn {
  padding: 12px 24px;
  border-radius: 10px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
  font-family: inherit;
}

.modal-btn.cancel {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--glass-border);
  color: var(--text-secondary);
}

.modal-btn.cancel:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

.modal-btn.save {
  background: var(--accent);
  border: none;
  color: #fff;
}

.modal-btn.save:hover {
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(var(--accent-rgb), 0.3);
}

/* Modal transition */
.modal-fade-enter-active {
  transition: all 0.25s ease;
}

.modal-fade-leave-active {
  transition: all 0.2s ease;
}

.modal-fade-enter-from, .modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-enter-from .avatar-picker-modal,
.modal-fade-enter-from .profile-modal {
  transform: scale(0.92);
}

/* Galaxy settings */
.slider-row {
  display: flex;
  align-items: center;
  gap: 16px;
}

.galaxy-slider {
  flex: 1;
  -webkit-appearance: none;
  appearance: none;
  height: 4px;
  background: var(--glass-border);
  border-radius: 2px;
  outline: none;
}

.galaxy-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--accent);
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(var(--accent-rgb), 0.4);
  transition: transform 0.15s ease;
}

.galaxy-slider::-webkit-slider-thumb:hover {
  transform: scale(1.15);
}

.slider-value {
  font-size: 13px;
  color: var(--accent);
  font-weight: 600;
  min-width: 44px;
  text-align: right;
  font-variant-numeric: tabular-nums;
}

.link-style-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.style-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  padding: 20px 16px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
  color: var(--text-secondary);
}

.style-card:hover {
  border-color: rgba(var(--accent-rgb), 0.4);
  background: rgba(var(--accent-rgb), 0.05);
}

.style-card.active {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.1);
  color: var(--accent);
}

.style-name {
  font-size: 13px;
  font-weight: 500;
}

/* Donate section */
.donate-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px 24px;
  background: rgba(var(--accent-rgb), 0.06);
  border: 1px solid rgba(var(--accent-rgb), 0.2);
  border-radius: 16px;
  cursor: pointer;
  width: 100%;
  transition: all 0.2s ease;
  font-family: inherit;
  color: var(--text-primary);
}

.donate-card:hover {
  background: rgba(var(--accent-rgb), 0.12);
  border-color: var(--accent);
  transform: translateY(-1px);
}

.donate-icon {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--accent-rgb), 0.15);
  border-radius: 12px;
  flex-shrink: 0;
}

.donate-icon svg {
  width: 22px;
  height: 22px;
  color: var(--accent);
}

.donate-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 4px;
  text-align: left;
}

.donate-title {
  font-size: 15px;
  font-weight: 600;
}

.donate-desc {
  font-size: 12px;
  color: var(--text-secondary);
}

.donate-arrow {
  width: 18px;
  height: 18px;
  color: var(--text-secondary);
  flex-shrink: 0;
}

/* Donate modal */
.donate-modal {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 20px;
  width: min(420px, 90vw);
  overflow: hidden;
  animation: scaleIn 0.25s ease;
}

.donate-body {
  padding: 28px 24px 32px;
  text-align: center;
}

.donate-thanks {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0 0 24px 0;
  line-height: 1.6;
}

.donate-qr-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.donate-qr-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.qr-img-wrapper {
  width: 160px;
  height: 160px;
  background: var(--bg-app);
  border: 1px solid var(--glass-border);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 10px;
}

.qr-img {
  width: 100%;
  height: 100%;
  border-radius: 8px;
  image-rendering: pixelated;
}

.qr-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-loading-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--glass-border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.qr-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

/* Version info card */
.version-info-card {
  margin-top: 24px;
  padding: 20px 24px;
  background: rgba(var(--accent-rgb), 0.04);
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  display: flex;
  align-items: center;
  gap: 16px;
  flex-wrap: wrap;
}

.version-info-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.version-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.version-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  font-family: 'JetBrains Mono', monospace;
}

.check-update-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 18px;
  background: rgba(var(--accent-rgb), 0.12);
  border: 1px solid rgba(var(--accent-rgb), 0.3);
  border-radius: 10px;
  color: var(--accent);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.check-update-btn:hover:not(:disabled) {
  background: rgba(var(--accent-rgb), 0.2);
  border-color: var(--accent);
}

.check-update-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.check-update-btn svg {
  width: 16px;
  height: 16px;
}

.check-update-btn.checking svg {
  animation: spin 1s linear infinite;
}

.btn-spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(var(--accent-rgb), 0.3);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.update-toast {
  width: 100%;
  text-align: center;
  font-size: 12px;
  padding: 6px 0;
  border-radius: 8px;
}

.update-toast.success {
  color: #4ade80;
  background: rgba(74, 222, 128, 0.08);
}

.update-toast.error {
  color: #f87171;
  background: rgba(248, 113, 113, 0.08);
}

.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: opacity 0.3s ease;
}

.toast-fade-enter-from,
.toast-fade-leave-to {
  opacity: 0;
}

/* Update modal */
.update-modal {
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 20px;
  width: min(480px, 90vw);
  overflow: hidden;
  animation: scaleIn 0.25s ease;
}

.update-body {
  padding: 24px 28px 28px;
}

.update-version-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  margin-bottom: 20px;
}

.update-from {
  font-size: 16px;
  color: var(--text-secondary);
  font-family: 'JetBrains Mono', monospace;
}

.update-to {
  font-size: 18px;
  font-weight: 700;
  color: var(--accent);
  font-family: 'JetBrains Mono', monospace;
}

.update-notes {
  margin-bottom: 24px;
}

.update-notes h4 {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
  margin: 0 0 10px 0;
}

.update-notes pre {
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-primary);
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 200px;
  overflow-y: auto;
  padding: 14px 16px;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  margin: 0;
  font-family: inherit;
}

.update-downloads {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.download-btn {
  display: flex;
  align-items: center;
  gap: 14px;
  width: 100%;
  padding: 14px 18px;
  background: rgba(var(--accent-rgb), 0.08);
  border: 1px solid rgba(var(--accent-rgb), 0.2);
  border-radius: 14px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s ease;
  font-family: inherit;
}

.download-btn:hover {
  background: rgba(var(--accent-rgb), 0.15);
  border-color: var(--accent);
  transform: translateY(-1px);
}

.download-btn svg {
  width: 22px;
  height: 22px;
  color: var(--accent);
  flex-shrink: 0;
}

.download-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  text-align: left;
}

.download-name {
  font-size: 14px;
  font-weight: 600;
}

.download-size {
  font-size: 12px;
  color: var(--text-secondary);
}
</style>