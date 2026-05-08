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
      <button class="logo-btn" @click="$emit('go-home')" :title="collapsed ? 'AstraLink' : ''">
        <div class="logo-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
            <path d="M12 2L2 7l10 5 10-5-10-5z" fill="var(--accent)" opacity="0.8"/>
            <path d="M2 17l10 5 10-5" stroke="var(--accent)" stroke-width="2" stroke-linecap="round"/>
            <path d="M2 12l10 5 10-5" stroke="var(--accent)" stroke-width="2" stroke-linecap="round" opacity="0.6"/>
          </svg>
        </div>
        <span v-if="!collapsed" class="brand-name">AstraLink</span>
      </button>
      <button class="collapse-btn" @click="toggleCollapse" :title="collapsed ? '展开' : '折叠'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round">
          <path v-if="collapsed" d="M9 18l6-6-6-6" />
          <path v-else d="M15 18l-6-6 6-6" />
        </svg>
      </button>
    </div>

    <!-- Tags section -->
    <div class="sidebar-section">
      <div class="section-header" :title="collapsed ? '标签' : ''">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
          <path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z"/>
          <line x1="7" y1="7" x2="7.01" y2="7"/>
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
      <div v-if="collapsed && tags && tags.length" class="collapsed-indicator">
        <div v-for="tag in tags.slice(0, 4)" :key="tag.id" class="mini-tag" @click="emit('tag-click', tag)">
          {{ tag.name }}
        </div>
      </div>
    </div>

    <!-- Recent section -->
    <div class="sidebar-section">
      <div class="section-header" :title="collapsed ? '最近' : ''">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
          <circle cx="12" cy="12" r="10"/>
          <polyline points="12 6 12 12 16 14"/>
        </svg>
        <span v-if="!collapsed" class="section-title">最近</span>
      </div>
      <div v-if="!collapsed" class="section-list recent-list">
        <div v-if="!recentNotes || recentNotes.length === 0" class="section-empty">
          <span>暂无记录</span>
        </div>
        <button
          v-for="note in (recentNotes || [])"
          :key="note.id"
          class="list-item recent-item"
          @click="$emit('open-note', { id: note.id, name: note.name })"
        >
          <svg class="item-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
          </svg>
          <span class="item-name">{{ note.name }}</span>
          <span class="item-time">{{ relativeTime(note.timestamp) }}</span>
        </button>
      </div>
      <div v-if="collapsed && recentNotes && recentNotes.length" class="collapsed-badge">
        {{ recentNotes.length > 9 ? '9+' : recentNotes.length }}
      </div>
    </div>

    <!-- Footer -->
    <div class="sidebar-footer" @click="$emit('open-settings')" :title="collapsed ? '设置' : ''">
      <div class="user-avatar-wrap">
        <img v-if="user?.avatar" :src="user.avatar" class="user-avatar-img" />
        <div v-else class="user-avatar">{{ user?.username?.[0] || 'U' }}</div>
      </div>
      <div v-if="!collapsed" class="user-info">
        <div class="user-name">{{ user?.username || '用户' }}</div>
        <div class="user-hint">点击设置</div>
      </div>
      <svg v-if="!collapsed" class="settings-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
        <circle cx="12" cy="12" r="3"/>
        <path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z"/>
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
  padding: 20px 10px;
  box-sizing: border-box;
  border-right: 1px solid var(--glass-border);
  flex-shrink: 0;
  transition: width 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  overflow: hidden;
}

.sidebar.collapsed {
  width: 64px;
  padding: 20px 8px;
}

/* Header */
.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 6px 24px;
  gap: 8px;
}

.logo-btn {
  display: flex;
  align-items: center;
  gap: 10px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 4px;
  border-radius: 8px;
  transition: background 0.2s;
  flex: 1;
  min-width: 0;
}

.logo-btn:hover {
  background: rgba(255, 255, 255, 0.05);
}

.sidebar.collapsed .logo-btn {
  justify-content: center;
}

.logo-icon {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--accent-rgb), 0.1);
  border-radius: 8px;
  flex-shrink: 0;
}

.brand-name {
  font-weight: 700;
  font-size: 14px;
  letter-spacing: 1px;
  color: var(--text-primary);
  white-space: nowrap;
}

.collapse-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 6px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: all 0.2s;
  flex-shrink: 0;
}

.collapse-btn:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: var(--glass-border);
  color: var(--text-primary);
}

.sidebar.collapsed .collapse-btn {
  opacity: 0.5;
}

.sidebar.collapsed .collapse-btn:hover {
  opacity: 1;
}

/* Sections */
.sidebar-section {
  margin-bottom: 6px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  border-radius: 8px;
  transition: background 0.15s;
}

.sidebar.collapsed .section-header {
  justify-content: center;
  padding: 10px;
}

.section-header svg {
  flex-shrink: 0;
  opacity: 0.6;
}

.section-title {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.5px;
  opacity: 0.7;
}

.section-count {
  margin-left: auto;
  font-size: 10px;
  background: rgba(var(--accent-rgb), 0.15);
  padding: 2px 8px;
  border-radius: 10px;
  color: var(--accent);
  font-weight: 600;
}

.collapsed-indicator {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 0;
}

.mini-tag {
  font-size: 10px;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.15s;
}

.mini-tag:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
}

/* Lists */
.section-list {
  max-height: 180px;
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
  border-radius: 2px;
}

.list-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 7px 10px 7px 34px;
  border-radius: 8px;
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: all 0.15s;
  border: none;
  background: transparent;
  cursor: pointer;
  font-family: inherit;
  text-align: left;
  width: 100%;
}

.list-item:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-primary);
}

.item-icon {
  flex-shrink: 0;
  opacity: 0.4;
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

.collapsed-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  background: var(--accent);
  color: var(--bg-app);
  font-size: 10px;
  font-weight: 700;
  border-radius: 10px;
  margin: 0 auto;
}

.section-empty {
  padding: 12px 10px 12px 34px;
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.4;
}

/* Footer */
.sidebar-footer {
  margin-top: auto;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 10px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
  overflow: hidden;
  border: 1px solid transparent;
}

.sidebar-footer:hover {
  background: rgba(255, 255, 255, 0.04);
  border-color: var(--glass-border);
}

.sidebar.collapsed .sidebar-footer {
  justify-content: center;
  padding: 12px 8px;
}

.user-avatar-wrap {
  flex-shrink: 0;
}

.user-avatar {
  width: 34px;
  height: 34px;
  background: linear-gradient(135deg, var(--accent) 0%, rgba(var(--accent-rgb), 0.7) 100%);
  color: #fff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 14px;
}

.user-avatar-img {
  width: 34px;
  height: 34px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--glass-border);
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-hint {
  font-size: 10px;
  color: var(--text-secondary);
  opacity: 0.5;
}

.settings-icon {
  color: var(--text-secondary);
  opacity: 0.4;
  flex-shrink: 0;
}

.sidebar-footer:hover .settings-icon {
  opacity: 0.7;
}
</style>