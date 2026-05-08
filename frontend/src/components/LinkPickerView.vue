<script setup>
import { ref, onMounted } from 'vue'
import { GetRecentNotes } from '../../wailsjs/go/main/App'

const props = defineProps(['insertPos'])
const emit = defineEmits(['close', 'select-note'])

const notes = ref([])
const loading = ref(true)
const showContent = ref(false)

async function loadNotes() {
  loading.value = true
  try {
    const result = await GetRecentNotes(20)
    notes.value = result || []
  } catch (e) {
    notes.value = []
  }
  loading.value = false
}

function selectNote(note) {
  emit('select-note', { id: note.id, title: note.name || note.title || '未命名' })
}

function handleOverlayClick(e) {
  if (e.target === e.currentTarget) {
    emit('close')
  }
}

function getDelay(index) {
  return (index * 40) + 'ms'
}

onMounted(() => {
  loadNotes()
  setTimeout(() => { showContent.value = true }, 50)
})
</script>

<template>
  <Transition name="modal-fade">
    <div v-if="showContent" class="link-picker-overlay" @click="handleOverlayClick">
      <div class="link-picker-container glass-panel">
        <!-- Header -->
        <div class="link-picker-header">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/>
            <path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/>
          </svg>
          <span>链接笔记</span>
        </div>

        <!-- Notes list -->
        <div class="link-picker-list">
          <div v-if="loading" class="link-picker-loading">
            <div class="loading-spinner"></div>
            <span>加载中...</span>
          </div>

          <div v-else-if="notes.length === 0" class="link-picker-empty">
            <span>暂无笔记</span>
          </div>

          <div
            v-else
            v-for="(note, index) in notes"
            :key="note.id"
            class="link-picker-item"
            :style="{ animationDelay: getDelay(index) }"
            @click="selectNote(note)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="flex-shrink:0;color:var(--accent);">
              <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
            </svg>
            <span class="note-title">{{ note.name || '未命名' }}</span>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.link-picker-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 300;
}

.link-picker-container {
  width: min(380px, 90vw);
  max-height: 60vh;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5), 0 0 0 1px var(--glass-border);
}

.link-picker-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px 20px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  border-bottom: 1px solid var(--glass-border);
}

.link-picker-header svg {
  color: var(--accent);
}

.link-picker-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.link-picker-list::-webkit-scrollbar {
  width: 4px;
}

.link-picker-list::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 2px;
}

.link-picker-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s;
  opacity: 0;
  animation: itemFadeIn 0.25s ease-out forwards;
}

@keyframes itemFadeIn {
  from { opacity: 0; transform: translateY(6px); }
  to { opacity: 1; transform: translateY(0); }
}

.link-picker-item:hover {
  background: rgba(var(--accent-rgb), 0.1);
}

.note-title {
  font-size: 13px;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-picker-loading,
.link-picker-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 40px;
  color: var(--text-secondary);
  font-size: 13px;
}

.loading-spinner {
  width: 18px;
  height: 18px;
  border: 2px solid var(--glass-border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Modal transition */
.modal-fade-enter-active {
  transition: all 0.2s ease-out;
}

.modal-fade-leave-active {
  transition: all 0.15s ease-in;
}

.modal-fade-enter-from {
  opacity: 0;
}

.modal-fade-enter-from .link-picker-container {
  transform: scale(0.95) translateY(-10px);
}

.modal-fade-leave-to {
  opacity: 0;
}
</style>