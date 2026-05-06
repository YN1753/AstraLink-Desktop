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

// Calculate font size based on note count
function getFontSize(noteCount) {
  const minSize = 11
  const maxSize = 18
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
    >
      {{ tag.name }}
    </div>
  </div>

  <!-- Expanded mode -->
  <div v-else class="tag-cloud">
    <div v-if="!tags || tags.length === 0" class="section-empty">
      <span>暂无标签</span>
    </div>
    <div v-else class="tag-wrapper">
      <span
        v-for="tag in tags"
        :key="tag.id"
        class="tag-chip"
        :class="{ selected: selectedTag === tag.id }"
        :style="{ fontSize: getFontSize(tag.noteCount) + 'px' }"
        @click="handleTagClick(tag)"
      >
        {{ tag.name }}
      </span>
    </div>
  </div>
</template>

<style scoped>
.tag-cloud-collapsed {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 0;
}

.collapsed-tag {
  color: var(--text-secondary);
  opacity: 0.6;
  cursor: pointer;
  transition: opacity 0.2s;
}

.collapsed-tag:hover {
  opacity: 1;
  color: var(--accent);
}

.tag-cloud {
  max-height: 180px;
  overflow-y: auto;
}

.tag-cloud::-webkit-scrollbar {
  width: 3px;
}
.tag-cloud::-webkit-scrollbar-track {
  background: transparent;
}
.tag-cloud::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 3px;
}

.tag-wrapper {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 8px 10px;
}

.tag-chip {
  padding: 4px 10px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--glass-border);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  white-space: nowrap;
}

.tag-chip:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: var(--accent);
  color: var(--accent);
}

.tag-chip.selected {
  background: var(--accent);
  border-color: var(--accent);
  color: var(--bg-app);
}

.section-empty {
  padding: 8px 10px;
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.4;
}
</style>
