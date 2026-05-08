<script setup>
import { computed } from 'vue'

const props = defineProps({
  tags: {
    type: Array,
    default: () => []
  },
  selectedTag: {
    type: String,
    default: null
  },
  collapsed: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['tag-click'])

function getFontSize(noteCount) {
  const minSize = 11
  const maxSize = 17
  if (props.tags.length === 0) return minSize

  const counts = props.tags.map(t => t.noteCount || 0)
  const minCount = Math.min(...counts)
  const maxCount = Math.max(...counts)

  if (maxCount === minCount) return (minSize + maxSize) / 2
  return minSize + ((noteCount - minCount) / (maxCount - minCount)) * (maxSize - minSize)
}

function handleTagClick(tag) {
  emit('tag-click', tag)
}
</script>

<template>
  <!-- Collapsed mode -->
  <div v-if="collapsed" class="tag-cloud-collapsed">
    <div
      v-for="tag in (tags || []).slice(0, 5)"
      :key="tag.id"
      class="collapsed-tag"
      :style="{ fontSize: getFontSize(tag.noteCount) + 'px' }"
      @click="handleTagClick(tag)"
    >
      {{ tag.name }}
    </div>
  </div>

  <!-- Expanded mode -->
  <div v-else class="tag-cloud">
    <div v-if="!tags || tags.length === 0" class="cloud-empty">
      <span>暂无标签</span>
    </div>
    <div v-else class="tag-wrapper">
      <button
        v-for="tag in tags"
        :key="tag.id"
        class="tag-btn"
        :class="{ selected: selectedTag === tag.id }"
        :style="{ fontSize: getFontSize(tag.noteCount) + 'px' }"
        @click="handleTagClick(tag)"
      >
        <span class="tag-hash">#</span>
        <span class="tag-text">{{ tag.name }}</span>
        <span v-if="tag.noteCount > 0" class="tag-badge">{{ tag.noteCount }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.tag-cloud-collapsed {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
}

.collapsed-tag {
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.2s;
}

.collapsed-tag:hover {
  color: var(--accent);
  background: rgba(var(--accent-rgb), 0.1);
}

.tag-cloud {
  max-height: 160px;
  overflow-y: auto;
  padding: 4px 2px;
}

.tag-cloud::-webkit-scrollbar {
  width: 3px;
}

.tag-cloud::-webkit-scrollbar-track {
  background: transparent;
}

.tag-cloud::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 2px;
}

.cloud-empty {
  padding: 12px 10px;
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.4;
  text-align: center;
}

.tag-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 4px 6px;
}

.tag-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 5px 10px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid transparent;
  border-radius: 14px;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
  color: inherit;
}

.tag-btn:hover {
  background: rgba(var(--accent-rgb), 0.08);
  border-color: rgba(var(--accent-rgb), 0.3);
}

.tag-btn.selected {
  background: var(--accent);
  border-color: var(--accent);
}

.tag-btn.selected .tag-hash,
.tag-btn.selected .tag-text {
  color: var(--bg-app);
}

.tag-btn.selected .tag-badge {
  background: rgba(0, 0, 0, 0.2);
  color: var(--bg-app);
}

.tag-hash {
  color: var(--accent);
  font-weight: 600;
  font-size: 0.85em;
}

.tag-text {
  color: var(--text-primary);
  font-weight: 500;
}

.tag-badge {
  font-size: 10px;
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-secondary);
  padding: 1px 6px;
  border-radius: 8px;
  margin-left: 2px;
}
</style>