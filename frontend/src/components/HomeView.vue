<script setup>
defineProps(['user', 'status', 'stats', 'recentNotes'])
const emit = defineEmits(['start-note', 'enter-galaxy', 'open-search', 'open-note'])
</script>

<template>
  <div class="home-container">
    <div class="hero-box">
      <!-- 状态徽章 -->
      <div class="badge">
        <span class="dot" :class="status === 'READY' ? 'ready' : 'saving'"></span>
        {{ status }} // 系统连结正常
      </div>

      <!-- 主标题 -->
      <h1 class="title">AstraLink</h1>
      <p class="motto">{{ user.motto }}</p>

      <!-- 搜索框 -->
      <div class="search-bar" @click="$emit('open-search')">
        <span class="search-icon">🔍</span>
        <span class="search-placeholder">搜索笔记、标签、内容...</span>
        <span class="search-shortcut">Ctrl+K</span>
      </div>

      <!-- 操作按钮 -->
      <div class="actions">
        <button class="btn-main" @click="$emit('start-note')">
          快速记录 / WRITE
        </button>
        <button class="btn-sub" @click="$emit('enter-galaxy')">
          进入星系 / GALAXY
        </button>
      </div>

      <!-- 统计数据 -->
      <div class="stats-row">
        <div class="stat-item">
          <span class="stat-num">{{ stats?.noteCount || 0 }}</span>
          <span class="stat-label">笔记</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-num">{{ stats?.galaxyCount || 0 }}</span>
          <span class="stat-label">星系</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-num">{{ stats?.tagCount || 0 }}</span>
          <span class="stat-label">标签</span>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <span class="stat-num">{{ ((stats?.totalSize || 0) / 1024 / 1024).toFixed(1) }}</span>
          <span class="stat-label">MB</span>
        </div>
      </div>

      <!-- 最近笔记 -->
      <div class="recent-section">
        <div class="section-title">最近编辑</div>
        <div class="recent-list">
          <div
            v-for="note in (recentNotes || [])"
            :key="note.id"
            class="recent-item"
            @click="$emit('open-note', note.id, note.name)"
          >
            <span class="note-icon">📄</span>
            <span class="note-name">{{ note.name }}</span>
          </div>
        </div>
        <div v-if="!recentNotes || recentNotes.length === 0" class="recent-empty">
          暂无笔记，开始创作吧
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-container {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.hero-box {
  text-align: center;
  max-width: min(600px, 90vw);
  padding: 0 clamp(10px, 3vw, 20px);
  box-sizing: border-box;
}

.badge {
  font-size: clamp(9px, 1.2vw, 10px);
  color: var(--text-secondary);
  letter-spacing: clamp(2px, 0.3vw, 3px);
  margin-bottom: clamp(15px, 3vw, 25px);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: clamp(6px, 1vw, 10px);
}

.dot {
  width: clamp(3px, 0.5vw, 4px);
  height: clamp(3px, 0.5vw, 4px);
  border-radius: 50%;
}

.dot.ready { background: #22c55e; }
.dot.saving { background: #eab308; }

.title {
  font-size: clamp(32px, 10vw, 100px);
  font-weight: 100;
  margin: 0;
  letter-spacing: clamp(-2px, -0.3vw, -4px);
  background: linear-gradient(to bottom, var(--text-primary), var(--text-secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  word-break: break-word;
}

.motto {
  font-size: clamp(12px, 2vw, 16px);
  color: var(--text-secondary);
  margin-top: clamp(10px, 2vw, 15px);
  letter-spacing: clamp(2px, 0.5vw, 4px);
  font-weight: 300;
}

/* 搜索框 */
.search-bar {
  margin-top: clamp(16px, 4vw, 40px);
  padding: clamp(8px, 1.5vw, 15px) clamp(12px, 2.5vw, 25px);
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 10px;
  cursor: pointer;
  transition: 0.3s;
  max-width: 450px;
  margin-left: auto;
  margin-right: auto;
}

.search-bar:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(var(--accent-rgb), 0.15);
}

.search-icon { font-size: 14px; opacity: 0.5; }
.search-placeholder { flex: 1; text-align: left; font-size: clamp(11px, 1.5vw, 14px); color: var(--text-secondary); }
.search-shortcut {
  font-size: 11px;
  color: var(--text-secondary);
  background: rgba(255, 255, 255, 0.05);
  padding: 3px 8px;
  border-radius: 4px;
  display: none;
}

.actions {
  margin-top: clamp(16px, 4vw, 40px);
  display: flex;
  gap: clamp(8px, 1.5vw, 15px);
  justify-content: center;
  flex-wrap: wrap;
}

.btn-main {
  padding: clamp(10px, 2vw, 15px) clamp(25px, 5vw, 40px);
  background: var(--text-primary);
  color: var(--bg-app);
  border-radius: clamp(8px, 1.5vw, 12px);
  font-weight: 800;
  border: none;
  cursor: pointer;
  transition: 0.2s;
  font-size: clamp(11px, 1.5vw, 13px);
}
.btn-main:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(255, 255, 255, 0.1);
}

.btn-sub {
  padding: clamp(10px, 2vw, 15px) clamp(25px, 5vw, 40px);
  background: transparent;
  border: 1px solid var(--glass-border);
  color: var(--text-primary);
  border-radius: clamp(8px, 1.5vw, 12px);
  cursor: pointer;
  font-size: clamp(11px, 1.5vw, 13px);
  transition: 0.2s;
}
.btn-sub:hover {
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(var(--accent-rgb), 0.15);
}

/* 统计 */
.stats-row {
  margin-top: clamp(20px, 4vw, 50px);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: clamp(12px, 3vw, 30px);
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: clamp(3px, 0.5vw, 5px);
}

.stat-num {
  font-size: clamp(20px, 4vw, 28px);
  font-weight: 600;
  color: var(--text-primary);
}

.stat-label {
  font-size: clamp(9px, 1.3vw, 11px);
  color: var(--text-secondary);
  letter-spacing: clamp(0.5px, 0.15vw, 1px);
}

.stat-divider {
  width: 1px;
  height: clamp(25px, 5vw, 40px);
  background: var(--glass-border);
}

/* 最近笔记 */
.recent-section {
  margin-top: clamp(20px, 4vw, 50px);
  text-align: left;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: clamp(8px, 1.5vw, 12px);
  padding: clamp(12px, 2vw, 20px) clamp(12px, 2.5vw, 25px);
  max-height: 200px;
  overflow-y: auto;
}

.section-title {
  font-size: clamp(9px, 1.2vw, 10px);
  color: var(--text-secondary);
  letter-spacing: clamp(1px, 0.2vw, 2px);
  margin-bottom: clamp(10px, 1.5vw, 15px);
  opacity: 0.6;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: clamp(5px, 1vw, 8px);
}

.recent-item {
  display: flex;
  align-items: center;
  gap: clamp(8px, 1.5vw, 12px);
  padding: clamp(8px, 1.3vw, 10px) clamp(8px, 1.5vw, 12px);
  border-radius: clamp(5px, 1vw, 8px);
  cursor: pointer;
  transition: 0.2s;
}

.recent-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.note-icon { font-size: clamp(12px, 1.8vw, 14px); }
.note-name { flex: 1; font-size: clamp(11px, 1.5vw, 13px); color: var(--text-primary); }
.note-time { font-size: clamp(9px, 1.3vw, 11px); color: var(--text-secondary); opacity: 0.6; }

.recent-empty {
  text-align: center;
  color: var(--text-secondary);
  font-size: clamp(10px, 1.4vw, 12px);
  opacity: 0.5;
  padding: 20px;
}
</style>