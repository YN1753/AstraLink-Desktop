<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { GetTagsWithCount, CreateTag, DeleteTagRelation } from '../../wailsjs/go/main/App'

const props = defineProps({
  noteId: String,
  noteTags: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['tags-changed', 'refresh-tags', 'tag-removed'])

const allTags = ref([])
const selectedTags = ref([])
const inputValue = ref('')
const showDropdown = ref(false)
const dropdownPosition = ref({ top: 0, left: 0 })
const inputRef = ref(null)

async function loadTags() {
  try {
    const tags = await GetTagsWithCount()
    allTags.value = tags || []
  } catch (e) {}
}

const filteredTags = computed(() => {
  if (!inputValue.value) return allTags.value
  const search = inputValue.value.toLowerCase()
  return allTags.value.filter(t => t.name.toLowerCase().includes(search))
})

function isSelected(tag) {
  return selectedTags.value.some(t => t.id === tag.id)
}

function addTag(tag) {
  if (!isSelected(tag)) {
    selectedTags.value.push(tag)
    emit('tags-changed', selectedTags.value)
  }
  inputValue.value = ''
  showDropdown.value = false
}

async function removeTag(tag) {
  if (props.noteId && !props.noteId.startsWith('new-')) {
    try {
      await DeleteTagRelation(tag.id, props.noteId)
      await loadTags()
    } catch (e) {}
  }
  selectedTags.value = selectedTags.value.filter(t => t.id !== tag.id)
  emit('tags-changed', selectedTags.value)
  emit('tag-removed', tag)
}

async function createNewTag() {
  if (!inputValue.value.trim()) return
  try {
    const newTag = await CreateTag({
      name: inputValue.value.trim(),
      parentId: '',
      others: {}
    })
    await loadTags()
    const created = allTags.value.find(t => t.id === newTag || t.name === inputValue.value.trim())
    if (created) {
      addTag(created)
    } else {
      addTag({ id: newTag, name: inputValue.value.trim(), noteCount: 0 })
    }
    emit('refresh-tags')
  } catch (e) {}
  inputValue.value = ''
  showDropdown.value = false
}

function handleInput(e) {
  inputValue.value = e.target.value
  showDropdown.value = true
  updateDropdownPosition()
}

function handleKeydown(e) {
  if (e.key === 'Enter' && inputValue.value.trim()) {
    e.preventDefault()
    const exact = filteredTags.value.find(t => t.name.toLowerCase() === inputValue.value.toLowerCase())
    if (exact) {
      addTag(exact)
    } else {
      createNewTag()
    }
  }
  if (e.key === 'Escape') {
    showDropdown.value = false
  }
}

function updateDropdownPosition() {
  if (!inputRef.value) return
  const rect = inputRef.value.getBoundingClientRect()
  dropdownPosition.value = {
    top: rect.bottom + 6,
    left: rect.left
  }
}

function handleClickOutside(e) {
  if (!e.target.closest('.tag-input-container')) {
    showDropdown.value = false
  }
}

watch(() => props.noteTags, (newTags) => {
  selectedTags.value = newTags ? [...newTags] : []
}, { immediate: true, deep: true })

onMounted(() => {
  loadTags()
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div class="tag-input-container">
    <div class="selected-tags">
      <button
        v-for="tag in selectedTags"
        :key="tag.id"
        class="tag-chip selected"
        @click="removeTag(tag)"
      >
        <span class="chip-hash">#</span>
        <span class="chip-name">{{ tag.name }}</span>
        <svg class="chip-remove" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"/>
          <line x1="6" y1="6" x2="18" y2="18"/>
        </svg>
      </button>
    </div>

    <div class="input-wrapper" ref="inputRef">
      <span class="input-prefix">@</span>
      <input
        type="text"
        class="tag-input-field"
        placeholder="添加标签..."
        :value="inputValue"
        @input="handleInput"
        @keydown="handleKeydown"
        @focus="showDropdown = true"
      />
    </div>

    <Transition name="dropdown-fade">
      <div
        v-if="showDropdown && (filteredTags.length > 0 || inputValue.trim())"
        class="tag-dropdown"
        :style="{ top: dropdownPosition.top + 'px', left: dropdownPosition.left + 'px' }"
      >
        <button
          v-for="tag in filteredTags.filter(t => !isSelected(t)).slice(0, 8)"
          :key="tag.id"
          class="dropdown-item"
          @click="addTag(tag)"
        >
          <span class="dropdown-hash">#</span>
          <span class="dropdown-name">{{ tag.name }}</span>
          <span class="dropdown-count">{{ tag.noteCount }}</span>
        </button>

        <button
          v-if="inputValue.trim() && !filteredTags.some(t => t.name.toLowerCase() === inputValue.toLowerCase())"
          class="dropdown-item create-new"
          @click="createNewTag"
        >
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          <span>创建 "{{ inputValue.trim() }}"</span>
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.tag-input-container {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: rgba(0, 0, 0, 0.2);
  border-bottom: 1px solid var(--glass-border);
}

.selected-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 8px 5px 10px;
  background: rgba(var(--accent-rgb), 0.15);
  border: 1px solid rgba(var(--accent-rgb), 0.3);
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
  color: inherit;
}

.tag-chip:hover {
  background: rgba(var(--accent-rgb), 0.25);
}

.tag-chip:hover .chip-remove {
  opacity: 1;
}

.chip-hash {
  color: var(--accent);
  font-weight: 600;
  font-size: 11px;
}

.chip-name {
  color: var(--text-primary);
  font-size: 12px;
  font-weight: 500;
}

.chip-remove {
  color: var(--text-secondary);
  opacity: 0.5;
  transition: opacity 0.15s;
}

.input-wrapper {
  display: flex;
  align-items: center;
  gap: 6px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 16px;
  padding: 6px 12px;
  transition: all 0.2s;
  min-width: 100px;
}

.input-wrapper:focus-within {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

.input-prefix {
  color: var(--accent);
  font-weight: 600;
  font-size: 13px;
}

.tag-input-field {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: 12px;
  width: 80px;
}

.tag-input-field::placeholder {
  color: var(--text-secondary);
  opacity: 0.4;
}

.tag-dropdown {
  position: fixed;
  min-width: 180px;
  max-height: 220px;
  overflow-y: auto;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
  z-index: 1000;
  padding: 6px;
}

.dropdown-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 12px;
  color: var(--text-secondary);
  transition: all 0.15s;
  border: none;
  background: transparent;
  font-family: inherit;
  text-align: left;
  width: 100%;
}

.dropdown-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
}

.dropdown-hash {
  color: var(--accent);
  font-weight: 600;
}

.dropdown-name {
  flex: 1;
  color: var(--text-primary);
}

.dropdown-count {
  font-size: 10px;
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-secondary);
  padding: 2px 6px;
  border-radius: 6px;
}

.dropdown-item.create-new {
  border-top: 1px solid var(--glass-border);
  margin-top: 4px;
  padding-top: 10px;
  color: var(--accent);
}

.dropdown-item.create-new:hover {
  background: rgba(var(--accent-rgb), 0.1);
}

.dropdown-item.create-new svg {
  flex-shrink: 0;
}

.dropdown-fade-enter-active,
.dropdown-fade-leave-active {
  transition: all 0.2s ease;
}

.dropdown-fade-enter-from,
.dropdown-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>