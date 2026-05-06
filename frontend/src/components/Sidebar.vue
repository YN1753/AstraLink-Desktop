<script setup>
import { ref, watch } from 'vue'
import TagCloud from './TagCloud.vue'

const props = defineProps({
  user: Object,
  currentView: String,
  tags: Array,
  recentNotes: Array,
  selectedTag: String,
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'open-note', 'open-settings', 'go-home', 'tag-click'])

const collapsed = ref(props.modelValue)

watch(() => props.modelValue, (val) => {
  collapsed.value = val
})

watch(collapsed, (val) => {
  localStorage.setItem('sidebarCollapsed', JSON.stringify(val))
  emit('update:modelValue', val)
})

function toggleCollapse() {
  collapsed.value = !collapsed.value
}

function relativeTime(timestamp) {
  const diff = Date.now() - timestamp
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}小时前`
  const days = Math.floor(hours / 24)
  return `${days}天前`
}
</script>

<template>
  <div class="sidebar glass-panel" :class="{ collapsed }">
    <!-- Header -->
    <div class="sidebar-header">
      <div class="header-left" @click="$emit('go-home')">
        <div class="logo-icon"></div>
        <span v-if="!collapsed" class="brand-name">ASTRALINK</span>
      </div>
      <button class="collapse-toggle" @click="toggleCollapse" :title="collapsed ? '展开' : '折叠'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path v-if="collapsed" d="M9 18l6-6-6-6" />
          <path v-else d="M15 18l-6-6 6-6" />
        </svg>
      </button>
    </div>

    <!-- Pin section -->
    <div class="sidebar-section">
      <div class="section-header" :title="collapsed ? '置顶星系' : ''">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
        </svg>
        <span v-if="!collapsed" class="section-title">置顶星系</span>
      </div>
      <div v-if="!collapsed" class="section-empty">
        <span>在星系中右键置顶</span>
      </div>
    </div>

    <!-- Tags section -->
    <div class="sidebar-section">
      <div class="section-header" :title="collapsed ? '标签' : ''">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z" />
          <line x1="7" y1="7" x2="7.01" y2="7" />
        </svg>
        <span v-if="!collapsed" class="section-title">标签</span>
        <span v-if="!collapsed && tags && tags.length" class="section-count">{{ tags.length }}</span>
      </div>
      <TagCloud
        v-if="!collapsed"
        :tags="tags"
        :selectedTag="selectedTag"
        :collapsed="false"
        @tag-click="(tag) => emit('tag-click', tag)"
      />
      <TagCloud
        v-if="collapsed && tags && tags.length"
        :tags="tags"
        :selectedTag="selectedTag"
        :collapsed="true"
        @tag-click="(tag) => emit('tag-click', tag)"
      />
    </div>

    <!-- Recent section -->
    <div class="sidebar-section">
      <div class="section-header" :title="collapsed ? '最近浏览' : ''">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="10" />
          <polyline points="12 6 12 12 16 14" />
        </svg>
        <span v-if="!collapsed" class="section-title">最近浏览</span>
        <span v-if="!collapsed && recentNotes && recentNotes.length" class="section-count">{{ recentNotes.length }}</span>
      </div>
      <div v-if="!collapsed" class="section-list recent-list">
        <div v-if="!recentNotes || recentNotes.length === 0" class="section-empty">
          <span>暂无浏览记录</span>
        </div>
        <div
          v-for="note in (recentNotes || [])"
          :key="note.id"
          class="list-item recent-item"
          @click="$emit('open-note', { id: note.id, name: note.name })"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
            <polyline points="14 2 14 8 20 8" />
          </svg>
          <span class="item-name">{{ note.name }}</span>
          <span class="item-time">{{ relativeTime(note.timestamp) }}</span>
        </div>
      </div>
      <div v-if="collapsed && recentNotes && recentNotes.length" class="badge-dot">{{ recentNotes.length > 9 ? '9+' : recentNotes.length }}</div>
    </div>

    <!-- Footer -->
    <div class="sidebar-footer" @click="$emit('open-settings')" :title="collapsed ? user?.username || '设置' : ''">
      <img v-if="user?.avatar" :src="user.avatar" class="user-avatar-img" />
      <div v-else class="user-avatar">{{ user?.username?.[0] || 'U' }}</div>
      <div v-if="!collapsed" class="user-info">
        <div class="user-name">{{ user?.username || '用户' }}</div>
        <div class="user-hint">设置</div>
      </div>
      <svg v-if="!collapsed" class="settings-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z" />
      </svg>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  width: 260px;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 24px 12px;
  box-sizing: border-box;
  border-right: 1px solid var(--glass-border);
  flex-shrink: 0;
  transition: width 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
  position: relative;
}

.sidebar.collapsed {
  width: 60px;
  padding: 24px 8px;
}

/* Header */
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px 28px;
  white-space: nowrap;
  overflow: hidden;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  flex: 1;
  min-width: 0;
}

.sidebar.collapsed .sidebar-header {
  justify-content: center;
  padding: 0 0 28px;
}

.sidebar.collapsed .header-left {
  justify-content: center;
}

.logo-icon {
  width: 4px;
  height: 22px;
  background: var(--accent);
  border-radius: 2px;
  box-shadow: 0 0 12px var(--accent);
  flex-shrink: 0;
}

.brand-name {
  font-weight: 900;
  font-size: 16px;
  letter-spacing: 2px;
  color: var(--text-primary);
}

/* Collapse toggle */
.collapse-toggle {
  width: 28px;
  height: 28px;
  background: transparent;
  border: 1px solid var(--glass-border);
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  flex-shrink: 0;
  transition: color 0.2s, background 0.2s, opacity 0.2s;
  opacity: 0.6;
}

.sidebar.collapsed .collapse-toggle {
  opacity: 0.4;
}

.collapse-toggle:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.06);
  opacity: 1;
}

/* Sections */
.sidebar-section {
  margin-bottom: 4px;
  position: relative;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  border-radius: 6px;
  transition: background 0.15s;
}

.sidebar.collapsed .section-header {
  justify-content: center;
  padding: 10px;
}

.section-header svg {
  flex-shrink: 0;
  opacity: 0.7;
}

.section-title {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.5px;
  opacity: 0.8;
}

.section-count {
  margin-left: auto;
  font-size: 10px;
  background: rgba(255, 255, 255, 0.06);
  padding: 1px 6px;
  border-radius: 10px;
  color: var(--text-secondary);
  opacity: 0.7;
}

.badge-dot {
  position: absolute;
  top: 4px;
  right: 4px;
  min-width: 16px;
  height: 16px;
  background: var(--accent);
  color: var(--bg-app);
  font-size: 9px;
  font-weight: 700;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 3px;
}

.section-empty {
  padding: 8px 10px 8px 34px;
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.4;
}

/* Lists */
.section-list {
  max-height: 200px;
  overflow-y: auto;
}

.section-list::-webkit-scrollbar {
  width: 3px;
}
.section-list::-webkit-scrollbar-track {
  background: transparent;
}
.section-list::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 3px;
}

.list-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px 6px 34px;
  border-radius: 5px;
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  transition: background 0.15s;
}

.tag-item {
  cursor: default;
}

.item-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  background: var(--accent);
  opacity: 0.5;
  flex-shrink: 0;
}

.item-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}

.item-time {
  font-size: 10px;
  opacity: 0.4;
  flex-shrink: 0;
}

.recent-item {
  cursor: pointer;
}

.recent-item svg {
  flex-shrink: 0;
  opacity: 0.4;
}

.recent-item:hover {
  background: rgba(255, 255, 255, 0.04);
  color: var(--text-primary);
}

/* Footer */
.sidebar-footer {
  margin-top: auto;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 10px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
  overflow: hidden;
  border: 1px solid var(--glass-border);
  background: rgba(255, 255, 255, 0.02);
}

.sidebar.collapsed .sidebar-footer {
  justify-content: center;
  padding: 10px;
}

.sidebar-footer:hover {
  background: rgba(255, 255, 255, 0.06);
}

.user-avatar {
  width: 32px;
  height: 32px;
  background: var(--text-primary);
  color: var(--bg-app);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 800;
  font-size: 13px;
  flex-shrink: 0;
}

.user-avatar-img {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--glass-border);
  flex-shrink: 0;
}

.user-info {
  flex: 1;
  overflow: hidden;
}

.user-name {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
}

.user-hint {
  font-size: 10px;
  color: var(--text-secondary);
  opacity: 0.6;
}

.settings-icon {
  color: var(--text-secondary);
  opacity: 0.4;
  flex-shrink: 0;
}
</style>
