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

// Load all tags
async function loadTags() {
  try {
    const tags = await GetTagsWithCount()
    allTags.value = tags || []
  } catch (e) {
    console.error('加载标签失败:', e)
  }
}

// Filter tags based on input
const filteredTags = computed(() => {
  if (!inputValue.value) return allTags.value
  const search = inputValue.value.toLowerCase()
  return allTags.value.filter(t => t.name.toLowerCase().includes(search))
})

// Check if a tag is already selected
function isSelected(tag) {
  return selectedTags.value.some(t => t.id === tag.id)
}

// Add tag to selected
function addTag(tag) {
  if (!isSelected(tag)) {
    selectedTags.value.push(tag)
    emit('tags-changed', selectedTags.value)
  }
  inputValue.value = ''
  showDropdown.value = false
}

// Remove tag from selected
async function removeTag(tag) {
  // If noteId exists and is not a new note, delete the relation
  if (props.noteId && !props.noteId.startsWith('new-')) {
    try {
      await DeleteTagRelation(tag.id, props.noteId)
      // Refresh the tag cache
      await loadTags()
    } catch (e) {
      console.error('删除标签关联失败:', e)
    }
  }
  selectedTags.value = selectedTags.value.filter(t => t.id !== tag.id)
  emit('tags-changed', selectedTags.value)
  emit('tag-removed', tag)
}

// Create new tag
async function createNewTag() {
  if (!inputValue.value.trim()) return
  try {
    const newTag = await CreateTag({
      name: inputValue.value.trim(),
      parentId: '',
      others: {}
    })
    // Reload tags to get the new one with correct ID
    await loadTags()
    // Find the newly created tag
    const created = allTags.value.find(t => t.id === newTag || t.name === inputValue.value.trim())
    if (created) {
      addTag(created)
    } else {
      // Fallback: create tag object manually
      addTag({ id: newTag, name: inputValue.value.trim(), noteCount: 0 })
    }
    emit('refresh-tags')
  } catch (e) {
    console.error('创建标签失败:', e)
  }
  inputValue.value = ''
  showDropdown.value = false
}

// Handle input
function handleInput(e) {
  inputValue.value = e.target.value
  showDropdown.value = true
  updateDropdownPosition()
}

// Handle keydown
function handleKeydown(e) {
  if (e.key === 'Enter' && inputValue.value.trim()) {
    e.preventDefault()
    // Check if exact match exists
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

// Update dropdown position
function updateDropdownPosition() {
  if (!inputRef.value) return
  const rect = inputRef.value.getBoundingClientRect()
  dropdownPosition.value = {
    top: rect.bottom + 4,
    left: rect.left
  }
}

// Handle click outside
function handleClickOutside(e) {
  if (!e.target.closest('.tag-input-container')) {
    showDropdown.value = false
  }
}

// Initialize selected tags from props
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
    <!-- Selected tags -->
    <div class="selected-tags">
      <span
        v-for="tag in selectedTags"
        :key="tag.id"
        class="tag-chip selected"
        @click="removeTag(tag)"
      >
        {{ tag.name }}
        <svg class="remove-icon" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18" />
          <line x1="6" y1="6" x2="18" y2="18" />
        </svg>
      </span>
    </div>

    <!-- Input -->
    <div class="input-wrapper">
      <span class="input-prefix">@</span>
      <input
        ref="inputRef"
        type="text"
        class="tag-input"
        placeholder="添加标签..."
        :value="inputValue"
        @input="handleInput"
        @keydown="handleKeydown"
        @focus="showDropdown = true"
      />
    </div>

    <!-- Dropdown -->
    <div
      v-if="showDropdown && (filteredTags.length > 0 || inputValue.trim())"
      class="tag-dropdown"
      :style="{ top: dropdownPosition.top + 'px', left: dropdownPosition.left + 'px' }"
    >
      <!-- Existing tags -->
      <div
        v-for="tag in filteredTags.filter(t => !isSelected(t)).slice(0, 8)"
        :key="tag.id"
        class="dropdown-item"
        @click="addTag(tag)"
      >
        <span class="tag-name">{{ tag.name }}</span>
        <span class="tag-count">{{ tag.noteCount }}</span>
      </div>

      <!-- Create new -->
      <div
        v-if="inputValue.trim() && !filteredTags.some(t => t.name.toLowerCase() === inputValue.toLowerCase())"
        class="dropdown-item create-new"
        @click="createNewTag"
      >
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="12" y1="5" x2="12" y2="19" />
          <line x1="5" y1="12" x2="19" y2="12" />
        </svg>
        <span>创建 "{{ inputValue.trim() }}"</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tag-input-container {
  position: relative;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.02);
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
  padding: 3px 8px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--glass-border);
  color: var(--text-secondary);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.2s;
}

.tag-chip:hover {
  background: rgba(255, 255, 255, 0.1);
  border-color: var(--accent);
}

.tag-chip.selected {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--bg-app);
}

.remove-icon {
  opacity: 0.6;
}

.tag-chip:hover .remove-icon {
  opacity: 1;
}

.input-wrapper {
  display: flex;
  align-items: center;
  gap: 4px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  padding: 2px 8px;
  transition: border-color 0.2s;
}

.input-wrapper:focus-within {
  border-color: var(--accent);
}

.input-prefix {
  color: var(--accent);
  font-weight: 600;
  font-size: 12px;
}

.tag-input {
  background: transparent;
  border: none;
  outline: none;
  color: var(--text-primary);
  font-size: 12px;
  width: 100px;
}

.tag-input::placeholder {
  color: var(--text-secondary);
  opacity: 0.5;
}

.tag-dropdown {
  position: fixed;
  min-width: 180px;
  max-height: 240px;
  overflow-y: auto;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  z-index: 1000;
  padding: 4px;
}

.dropdown-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 12px;
  color: var(--text-secondary);
  transition: background 0.15s;
}

.dropdown-item:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
}

.dropdown-item.create-new {
  border-top: 1px solid var(--glass-border);
  margin-top: 4px;
  padding-top: 8px;
  color: var(--accent);
}

.dropdown-item.create-new:hover {
  background: rgba(59, 130, 246, 0.1);
}

.tag-name {
  flex: 1;
}

.tag-count {
  font-size: 10px;
  opacity: 0.5;
  background: rgba(255, 255, 255, 0.06);
  padding: 1px 6px;
  border-radius: 8px;
}
</style>
