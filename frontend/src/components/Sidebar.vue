<script setup>
defineProps(['user'])
defineEmits(['open-note', 'open-settings'])
</script>

<template>
  <div class="sidebar-container glass-panel">
    <div class="brand">
      <div class="logo-icon"></div>
      <span class="brand-name">ASTRALINK</span>
    </div>

    <nav class="nav-list">
      <div class="nav-group">NAVIGATE</div>
      <div class="nav-item" @click="$emit('open-note')">
        <i class="fa-solid fa-plus"></i> 快速记录 / QUICK WRITE
      </div>
      <div class="nav-item">
        <i class="fa-solid fa-box-archive"></i> 所有记录 / ARCHIVE
      </div>
    </nav>

    <div class="sidebar-footer" @click="$emit('open-settings')">
      <img v-if="user.avatar" :src="user.avatar" class="user-avatar-img" />
      <div v-else class="user-avatar">{{ user.username?.[0] || 'U' }}</div>
      <div class="user-info">
        <div class="uname">{{ user.username }}</div>
        <div class="ustatus">设置 / SETTINGS</div>
      </div>
      <div class="cog-icon">⚙️</div>
    </div>
  </div>
</template>

<style scoped>
.sidebar-container {
  width: 260px;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 30px 15px;
  box-sizing: border-box;
  /* 移除实色边框，改用变量 */
  border-right: 1px solid var(--glass-border);
}

.brand { display: flex; align-items: center; gap: 12px; padding: 0 15px 40px; }
.logo-icon { width: 4px; height: 24px; background: var(--accent); border-radius: 2px; box-shadow: 0 0 15px var(--accent); }
.brand-name { font-weight: 900; font-size: 18px; letter-spacing: 2px; color: var(--text-primary); }

.nav-group { font-size: 11px; color: var(--text-secondary); letter-spacing: 1px; padding: 20px 15px 10px; opacity: 0.6; }
.nav-item {
  padding: 12px 15px; border-radius: 8px; color: var(--text-secondary); font-size: 13px;
  cursor: pointer; transition: 0.2s; font-weight: 500;
}
.nav-item:hover { background: rgba(255, 255, 255, 0.05); color: var(--text-primary); }

.sidebar-footer {
  margin-top: auto;
  padding: 15px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  cursor: pointer;
  transition: 0.2s;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
}
.sidebar-footer:hover { background: rgba(255, 255, 255, 0.08); }

.user-avatar {
  width: 36px; height: 36px; background: var(--text-primary); color: var(--bg-app);
  border-radius: 50%; display: flex; align-items: center; justify-content: center;
  font-weight: 800; font-size: 14px;
}

.user-avatar-img {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  object-fit: cover;
  border: 2px solid var(--glass-border);
}

.uname { font-size: 13px; font-weight: 700; color: var(--text-primary); }
.ustatus { font-size: 10px; color: var(--text-secondary); opacity: 0.7; }
.cog-icon { font-size: 14px; color: var(--text-secondary); opacity: 0.5; }
</style>