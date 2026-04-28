<script setup>
import { ref } from 'vue'
const props = defineProps(['user'])
const emit = defineEmits(['close', 'update', 'upload'])

const form = ref({ username: props.user.username, motto: props.user.motto })
const fileInput = ref(null)

const onFileChange = (e) => {
  const file = e.target.files[0]
  if (!file || file.size > 2 * 1024 * 1024) return // 限制 2MB
  const reader = new FileReader()
  reader.onload = (ev) => emit('upload', ev.target.result)
  reader.readAsDataURL(file)
}
</script>

<template>
  <div class="modal-mask" @click.self="$emit('close')">
    <div class="settings-card">
      <div class="side-bar">
        <div class="avatar-uploader" @click="fileInput.click()">
          <img v-if="user.avatar" :src="user.avatar" class="avatar-img" />
          <div v-else class="avatar-placeholder">{{ form.username[0] }}</div>
          <div class="hover-tip">UPLOAD</div>
        </div>
        <input type="file" ref="fileInput" hidden @change="onFileChange" accept="image/*">
        <div class="side-info">
          <span class="label">IDENTITY</span>
          <span class="val">NODE-01</span>
        </div>
      </div>

      <div class="main-body">
        <div class="form-header">
          <h2>Profile Config</h2>
          <p>修改你的星链身份凭证</p>
        </div>

        <div class="form-group">
          <label>USERNAME / 昵称</label>
          <input v-model="form.username" type="text" placeholder="输入代号..." />
        </div>

        <div class="form-group">
          <label>MOTTO / 座右铭</label>
          <textarea v-model="form.motto" rows="3" placeholder="写下你的航行哲学..."></textarea>
        </div>

        <div class="form-actions">
          <button class="btn-cancel" @click="$emit('close')">CANCEL</button>
          <button class="btn-confirm" @click="$emit('update', { ...form })">CONFIRM / 确认保存</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-mask {
  position: fixed; inset: 0; background: rgba(0,0,0,0.85);
  backdrop-filter: blur(20px); z-index: 1000;
  display: flex; align-items: center; justify-content: center;
}

.settings-card {
  width: 660px; background: #0c0c0c; border: 1px solid rgba(255,255,255,0.05);
  border-radius: 24px; display: flex; overflow: hidden;
  box-shadow: 0 50px 100px rgba(0,0,0,0.5);
}

.side-bar {
  width: 200px; background: rgba(255,255,255,0.02);
  border-right: 1px solid rgba(255,255,255,0.05);
  display: flex; flex-direction: column; align-items: center; padding: 40px 0;
}

.avatar-uploader {
  width: 100px; height: 100px; border-radius: 35px; background: #151515;
  border: 1px solid rgba(255,255,255,0.1); overflow: hidden;
  cursor: pointer; position: relative; transition: 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}
.avatar-uploader:hover { border-color: #3b82f6; transform: scale(1.05); }

.avatar-img { width: 100%; height: 100%; object-fit: cover; }
.avatar-placeholder { height: 100%; display: flex; align-items: center; justify-content: center; font-size: 32px; font-weight: 200; color: #333; }

.hover-tip {
  position: absolute; inset: 0; background: rgba(59, 130, 246, 0.7);
  display: flex; align-items: center; justify-content: center;
  font-size: 10px; font-weight: 900; letter-spacing: 1px; opacity: 0; transition: 0.2s;
}
.avatar-uploader:hover .hover-tip { opacity: 1; }

.side-info { margin-top: auto; text-align: center; }
.side-info .label { display: block; font-size: 9px; color: #444; letter-spacing: 2px; }
.side-info .val { font-size: 12px; color: #222; font-family: monospace; }

.main-body { flex: 1; padding: 40px 50px; display: flex; flex-direction: column; }
.form-header h2 { font-size: 24px; font-weight: 200; margin: 0; color: #fff; }
.form-header p { font-size: 13px; color: #555; margin: 5px 0 35px 0; }

.form-group { margin-bottom: 25px; }
.form-group label { display: block; font-size: 10px; color: #444; margin-bottom: 8px; letter-spacing: 1px; }

input, textarea {
  width: 100%; background: #111; border: 1px solid #1a1a1a;
  border-radius: 12px; padding: 14px; color: #fff; font-size: 15px; outline: none; transition: 0.3s;
}
input:focus, textarea:focus { border-color: #333; background: #141414; }

.form-actions { margin-top: auto; display: flex; gap: 12px; }
.btn-confirm { flex: 1; background: #fff; color: #000; border: none; padding: 14px; border-radius: 10px; font-weight: 600; cursor: pointer; transition: 0.3s; }
.btn-confirm:hover { background: #3b82f6; color: #fff; }
.btn-cancel { padding: 14px 20px; border: 1px solid #222; background: transparent; color: #555; border-radius: 10px; cursor: pointer; }
</style>