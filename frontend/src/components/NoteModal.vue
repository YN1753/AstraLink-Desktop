<script setup>
import { ref } from 'vue'

const emit = defineEmits(['close', 'save'])
const content = ref('')
const isSaving = ref(false)

const handleSave = () => {
  if (!content.value) return
  isSaving.value = true
  // 模拟保存逻辑
  setTimeout(() => {
    emit('save', content.value)
    isSaving.value = false
    content.value = ''
  }, 800)
}
</script>

<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="modal-box">
      <textarea
          v-model="content"
          placeholder="在此记录灵感..."
          autofocus
      ></textarea>

      <div class="modal-btns">
        <span class="hint">ESC 或点击空白处取消</span>
        <button class="confirm-btn" @click="handleSave">
          {{ isSaving ? '同步中...' : '确认存入 / SAVE' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-mask {
  position: fixed; inset: 0; background: rgba(0,0,0,0.8);
  backdrop-filter: blur(20px); z-index: 100;
  display: flex; align-items: center; justify-content: center;
}

.modal-box {
  width: 100%; max-width: 600px; padding: 40px;
}

textarea {
  width: 100%; height: 150px; background: transparent;
  border: none; outline: none; color: #fff;
  font-size: 24px; font-weight: 200; resize: none;
  line-height: 1.6;
}

.modal-btns {
  margin-top: 40px; padding-top: 20px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  display: flex; justify-content: space-between; align-items: center;
}

.hint { font-size: 10px; color: #444; letter-spacing: 1px; }

.confirm-btn {
  background: none; border: none; color: #3b82f6;
  font-weight: bold; cursor: pointer; font-size: 14px;
}
</style>