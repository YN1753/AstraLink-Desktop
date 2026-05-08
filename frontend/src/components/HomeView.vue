<script setup>
import { ref, onMounted } from 'vue'

defineProps(['user', 'status', 'stats', 'recentNotes'])
const emit = defineEmits(['start-note', 'enter-galaxy', 'open-search', 'open-note'])

const mounted = ref(false)

onMounted(() => {
  setTimeout(() => { mounted.value = true }, 50)
})

function getDelay(index) {
  return (index * 60) + 'ms'
}
</script>

<template>
  <div class="home-wrapper">
    <div class="home-container" :class="{ mounted }">
      <!-- Ambient glow effects -->
      <div class="ambient-glow glow-1"></div>
      <div class="ambient-glow glow-2"></div>

      <!-- Hero Section -->
      <div class="hero-section">
        <!-- Status indicator -->
        <div class="status-indicator">
          <div class="status-dot" :class="status === 'READY' ? 'ready' : 'saving'"></div>
          <span class="status-text">STELLAR LINK ESTABLISHED</span>
        </div>

        <!-- Main title -->
        <h1 class="main-title">
          <span class="title-word">Astra</span><span class="title-word accent">Link</span>
        </h1>

        <!-- Subtitle / Motto -->
        <p class="motto-text">{{ user.motto }}</p>

        <!-- Search trigger -->
        <button class="search-trigger" @click="$emit('open-search')">
          <div class="search-icon-wrap">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="11" cy="11" r="8"/>
              <path d="M21 21l-4.35-4.35"/>
            </svg>
          </div>
          <span class="search-placeholder-text">探索你的知识星系...</span>
          <div class="search-shortcut">
            <kbd>⌘</kbd><kbd>K</kbd>
          </div>
        </button>

        <!-- Action buttons -->
        <div class="action-buttons">
          <button class="action-btn primary" @click="$emit('start-note')">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 5v14M5 12h14"/>
            </svg>
            <span>开始书写</span>
          </button>
          <button class="action-btn secondary" @click="$emit('enter-galaxy')">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="12" cy="12" r="10"/>
              <circle cx="12" cy="12" r="3"/>
              <path d="M12 2v4M12 18v4M2 12h4M18 12h4"/>
            </svg>
            <span>进入星系</span>
          </button>
        </div>
      </div>

      <!-- Stats section -->
      <div class="stats-section">
        <div class="stat-card" v-if="stats?.noteCount !== undefined">
          <div class="stat-value">{{ stats.noteCount }}</div>
          <div class="stat-label">笔记</div>
          <div class="stat-decoration"></div>
        </div>
        <div class="stat-card" v-if="stats?.galaxyCount !== undefined">
          <div class="stat-value">{{ stats.galaxyCount }}</div>
          <div class="stat-label">星系</div>
          <div class="stat-decoration"></div>
        </div>
        <div class="stat-card" v-if="stats?.tagCount !== undefined">
          <div class="stat-value">{{ stats.tagCount }}</div>
          <div class="stat-label">标签</div>
          <div class="stat-decoration"></div>
        </div>
        <div class="stat-card" v-if="stats?.totalSize !== undefined">
          <div class="stat-value">{{ (stats.totalSize / 1024 / 1024).toFixed(1) }}</div>
          <div class="stat-label">MB</div>
          <div class="stat-decoration"></div>
        </div>
      </div>

      <!-- Recent notes -->
      <div class="recent-section" v-if="recentNotes && recentNotes.length > 0">
        <div class="section-header">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
          <span>最近编辑</span>
        </div>
        <div class="recent-list">
          <button
            v-for="(note, index) in recentNotes.slice(0, 5)"
            :key="note.id"
            class="recent-item"
            :style="{ animationDelay: getDelay(index) }"
            @click="$emit('open-note', note.id, note.name)"
          >
            <div class="note-icon-wrap">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="8" y1="13" x2="16" y2="13"/>
                <line x1="8" y1="17" x2="12" y2="17"/>
              </svg>
            </div>
            <span class="note-name">{{ note.name }}</span>
            <svg class="arrow-icon" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M9 18l6-6-6-6"/>
            </svg>
          </button>
        </div>
      </div>

      <!-- Empty state -->
      <div class="empty-hint" v-else>
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
          <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
        </svg>
        <p>点击「开始书写」创建你的第一颗星辰</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-wrapper {
  width: 100%;
  height: 100%;
  overflow-y: auto;
  overflow-x: hidden;
  scroll-behavior: smooth;
}

.home-container {
  width: 100%;
  min-height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: clamp(40px, 8vw, 80px) clamp(16px, 4vw, 40px);
  position: relative;
  box-sizing: border-box;
}

.home-container.mounted {
  animation: fadeInUp 0.8s ease-out forwards;
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(30px); }
  to { opacity: 1; transform: translateY(0); }
}

/* Ambient glows */
.ambient-glow {
  position: fixed;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.35;
  pointer-events: none;
  z-index: 0;
}

.glow-1 {
  width: min(400px, 60vw);
  height: min(400px, 60vw);
  background: radial-gradient(circle, var(--accent) 0%, transparent 70%);
  top: -10%;
  right: -10%;
  animation: pulse 8s ease-in-out infinite;
}

.glow-2 {
  width: min(300px, 50vw);
  height: min(300px, 50vw);
  background: radial-gradient(circle, rgba(99, 102, 241, 0.4) 0%, transparent 70%);
  bottom: 5%;
  left: -5%;
  animation: pulse 10s ease-in-out infinite 2s;
}

@keyframes pulse {
  0%, 100% { opacity: 0.25; transform: scale(1); }
  50% { opacity: 0.45; transform: scale(1.1); }
}

/* Hero Section */
.hero-section {
  width: 100%;
  max-width: 600px;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  z-index: 1;
  margin-bottom: clamp(24px, 5vw, 48px);
}

/* Status indicator */
.status-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: clamp(16px, 3vw, 24px);
  opacity: 0;
  animation: fadeSlideIn 0.6s ease-out 0.2s forwards;
}

@keyframes fadeSlideIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  animation: statusPulse 2s ease-in-out infinite;
}

.status-dot.ready {
  background: #22c55e;
  box-shadow: 0 0 8px #22c55e80;
}

.status-dot.saving {
  background: #eab308;
  box-shadow: 0 0 8px #eab30880;
}

@keyframes statusPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.status-text {
  font-size: clamp(9px, 1.2vw, 11px);
  letter-spacing: 2px;
  color: var(--text-secondary);
  font-weight: 500;
}

/* Main title */
.main-title {
  font-size: clamp(36px, 10vw, 96px);
  font-weight: 800;
  margin: 0 0 clamp(8px, 2vw, 14px) 0;
  letter-spacing: clamp(-2px, -0.5vw, -4px);
  line-height: 1.1;
  opacity: 0;
  animation: fadeSlideIn 0.6s ease-out 0.35s forwards;
}

.title-word {
  display: inline-block;
  background: linear-gradient(135deg, var(--text-primary) 0%, var(--text-secondary) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.title-word.accent {
  background: linear-gradient(135deg, var(--accent) 0%, rgba(99, 102, 241, 0.6) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Motto */
.motto-text {
  font-size: clamp(12px, 2vw, 15px);
  color: var(--text-secondary);
  margin: 0 0 clamp(20px, 4vw, 40px) 0;
  letter-spacing: clamp(1px, 0.3vw, 3px);
  font-weight: 300;
  opacity: 0;
  animation: fadeSlideIn 0.6s ease-out 0.5s forwards;
}

/* Search trigger */
.search-trigger {
  width: 100%;
  max-width: 480px;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: clamp(12px, 2vw, 16px) clamp(16px, 3vw, 20px);
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: clamp(12px, 2vw, 16px);
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  opacity: 0;
  animation: fadeSlideIn 0.6s ease-out 0.65s forwards;
  box-sizing: border-box;
}

.search-trigger::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent 0%, rgba(var(--accent-rgb), 0.05) 50%, transparent 100%);
  opacity: 0;
  transition: opacity 0.3s;
}

.search-trigger:hover {
  transform: translateY(-2px);
  border-color: var(--accent);
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.35), 0 0 30px rgba(var(--accent-rgb), 0.08);
}

.search-trigger:hover::before {
  opacity: 1;
}

.search-icon-wrap {
  width: clamp(32px, 6vw, 40px);
  height: clamp(32px, 6vw, 40px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--accent-rgb), 0.1);
  border-radius: 10px;
  color: var(--accent);
  flex-shrink: 0;
}

.search-placeholder-text {
  flex: 1;
  text-align: left;
  font-size: clamp(12px, 1.8vw, 14px);
  color: var(--text-secondary);
  transition: color 0.2s;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.search-trigger:hover .search-placeholder-text {
  color: var(--text-primary);
}

.search-shortcut {
  display: flex;
  gap: 4px;
  flex-shrink: 0;
}

.search-shortcut kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 22px;
  height: 20px;
  padding: 0 5px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--glass-border);
  border-radius: 4px;
  font-size: 10px;
  color: var(--text-secondary);
  font-family: inherit;
}

/* Action buttons */
.action-buttons {
  display: flex;
  gap: clamp(8px, 2vw, 14px);
  margin-top: clamp(16px, 4vw, 32px);
  opacity: 0;
  animation: fadeSlideIn 0.6s ease-out 0.8s forwards;
  flex-wrap: wrap;
  justify-content: center;
}

.action-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: clamp(10px, 2vw, 14px) clamp(20px, 4vw, 28px);
  border-radius: 10px;
  font-size: clamp(11px, 1.5vw, 13px);
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  border: none;
  font-family: inherit;
  white-space: nowrap;
}

.action-btn.primary {
  background: linear-gradient(135deg, var(--accent) 0%, rgba(var(--accent-rgb), 0.85) 100%);
  color: #fff;
  box-shadow: 0 6px 24px rgba(var(--accent-rgb), 0.3);
}

.action-btn.primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 32px rgba(var(--accent-rgb), 0.4);
}

.action-btn.primary:active {
  transform: translateY(0);
}

.action-btn.secondary {
  background: transparent;
  border: 1px solid var(--glass-border);
  color: var(--text-primary);
  backdrop-filter: blur(10px);
}

.action-btn.secondary:hover {
  background: rgba(255, 255, 255, 0.06);
  border-color: var(--accent);
  transform: translateY(-2px);
}

.action-btn svg {
  flex-shrink: 0;
}

/* Stats section */
.stats-section {
  display: flex;
  gap: clamp(10px, 2.5vw, 20px);
  margin-bottom: clamp(24px, 5vw, 44px);
  opacity: 0;
  animation: fadeSlideIn 0.6s ease-out 1s forwards;
  z-index: 1;
  flex-wrap: wrap;
  justify-content: center;
  width: 100%;
  max-width: 600px;
}

.stat-card {
  position: relative;
  padding: clamp(14px, 3vw, 22px) clamp(16px, 4vw, 28px);
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: clamp(12px, 2vw, 16px);
  text-align: center;
  overflow: hidden;
  transition: all 0.3s ease;
  flex: 1;
  min-width: clamp(100px, 20vw, 130px);
}

.stat-card:hover {
  transform: translateY(-3px);
  border-color: rgba(var(--accent-rgb), 0.3);
  box-shadow: 0 12px 36px rgba(0, 0, 0, 0.25);
}

.stat-value {
  font-size: clamp(22px, 4vw, 30px);
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: clamp(9px, 1.2vw, 11px);
  color: var(--text-secondary);
  letter-spacing: 1px;
  text-transform: uppercase;
}

.stat-decoration {
  position: absolute;
  top: 0;
  right: 0;
  width: 50px;
  height: 50px;
  background: radial-gradient(circle at top right, rgba(var(--accent-rgb), 0.12) 0%, transparent 70%);
  pointer-events: none;
}

/* Recent section */
.recent-section {
  width: 100%;
  max-width: 480px;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: clamp(12px, 2vw, 16px);
  padding: clamp(14px, 3vw, 22px);
  box-sizing: border-box;
  opacity: 0;
  animation: fadeSlideIn 0.6s ease-out 1.15s forwards;
  z-index: 1;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: clamp(10px, 1.3vw, 12px);
  color: var(--text-secondary);
  letter-spacing: 1px;
  text-transform: uppercase;
  margin-bottom: clamp(10px, 2vw, 14px);
  opacity: 0.7;
}

.section-header svg {
  opacity: 0.7;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.recent-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: clamp(8px, 1.5vw, 12px) clamp(10px, 2vw, 14px);
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
  font-family: inherit;
  color: inherit;
  width: 100%;
  box-sizing: border-box;
  opacity: 0;
  animation: itemFadeIn 0.4s ease-out forwards;
}

@keyframes itemFadeIn {
  from { opacity: 0; transform: translateX(-10px); }
  to { opacity: 1; transform: translateX(0); }
}

.recent-item:hover {
  background: rgba(255, 255, 255, 0.05);
  border-color: var(--glass-border);
}

.recent-item:hover .note-icon-wrap {
  background: rgba(var(--accent-rgb), 0.18);
  color: var(--accent);
}

.recent-item:hover .arrow-icon {
  opacity: 0.6;
  transform: translateX(2px);
}

.note-icon-wrap {
  width: clamp(28px, 5vw, 32px);
  height: clamp(28px, 5vw, 32px);
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.04);
  border-radius: 7px;
  color: var(--text-secondary);
  flex-shrink: 0;
  transition: all 0.2s ease;
}

.note-name {
  flex: 1;
  font-size: clamp(11px, 1.5vw, 13px);
  color: var(--text-primary);
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.arrow-icon {
  color: var(--text-secondary);
  opacity: 0;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

/* Empty hint */
.empty-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  margin-top: clamp(20px, 4vw, 40px);
  color: var(--text-secondary);
  opacity: 0;
  animation: fadeSlideIn 0.6s ease-out 1.15s forwards;
  z-index: 1;
}

.empty-hint svg {
  opacity: 0.3;
}

.empty-hint p {
  font-size: clamp(11px, 1.5vw, 13px);
  margin: 0;
  opacity: 0.6;
}

/* Scrollbar */
.home-wrapper::-webkit-scrollbar {
  width: 6px;
}

.home-wrapper::-webkit-scrollbar-track {
  background: transparent;
}

.home-wrapper::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 3px;
}

.home-wrapper::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.12);
}
</style>