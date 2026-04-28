<script setup>
defineProps(['user', 'status'])
defineEmits(['start-note', 'enter-galaxy'])

</script>

<template>
  <div class="home-container">
    <div class="hero-box">
      <div class="badge">
        <span class="dot" :class="status"></span>
        {{ status }} // 系统连结正常
      </div>
      <h1 class="title">AstraLink</h1>
      <p class="motto">{{ user.motto }}</p>

      <div class="actions">
        <button class="btn-main" @click="$emit('start-note')">
          快速记录 / WRITE
        </button>
        <button class="btn-sub" @click="$emit('enter-galaxy')">
          进入星系 / GALAXY
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.home-container { width: 100%; height: 100%; display: flex; justify-content: center; align-items: center; }
.hero-box { text-align: center; }
.title { font-size: 120px; font-weight: 100; margin: 0; letter-spacing: -6px; background: linear-gradient(to bottom, #fff, #444); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
.motto { font-size: 18px; color: #666; margin-top: 15px; letter-spacing: 5px; font-weight: 300; }
.badge { font-size: 10px; color: #444; letter-spacing: 3px; margin-bottom: 25px; display: flex; align-items: center; justify-content: center; gap: 10px; }
.dot { width: 4px; height: 4px; border-radius: 50%; background: #22c55e; }
.actions { margin-top: 60px; display: flex; gap: 20px; justify-content: center; }

/* 🌟 修正：白底黑字按钮 */
.btn-main {
  padding: 18px 45px; background: #ffffff !important; color: #000000 !important;
  border-radius: 12px; font-weight: 900; border: none; cursor: pointer; transition: 0.2s; font-size: 14px;
}
.btn-main:hover { transform: translateY(-3px); box-shadow: 0 10px 40px rgba(255,255,255,0.15); }

.btn-sub {
  padding: 18px 45px; background: rgba(255,255,255,0.02); border: 1px solid rgba(255,255,255,0.1);
  color: white; border-radius: 12px; cursor: pointer; font-size: 14px;
}
.btn-sub:hover { background: rgba(255,255,255,0.08); }

</style>