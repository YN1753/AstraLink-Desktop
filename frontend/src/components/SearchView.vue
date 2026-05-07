<script setup>
import { ref, computed, watch } from 'vue'
import { NInput, NButton, NEmpty, NSpin } from 'naive-ui'
import { SearchNotes } from '../../wailsjs/go/main/App'

const props = defineProps(['user'])
const emit = defineEmits(['close', 'open-note'])

// 搜索相关
const searchQuery = ref('')
const isSearching = ref(false)
const searchResults = ref([])
const recentSearches = ref([])

// 真实搜索方法 - 使用 Bleve 全文搜索
async function performSearch(query) {
  if (!query.trim()) {
    searchResults.value = []
    return
  }

  isSearching.value = true

  try {
    const results = await SearchNotes(query)
    searchResults.value = results || []
  } catch (e) {
    console.error('搜索失败:', e)
    searchResults.value = []
  }

  isSearching.value = false
}

// 防抖搜索
let searchTimeout = null
watch(searchQuery, (newQuery) => {
  clearTimeout(searchTimeout)
  searchTimeout = setTimeout(() => {
    performSearch(newQuery)
  }, 300)
})

function openNote(note) {
  emit('open-note', { id: note.id, name: note.title })
}

function addToRecent(query) {
  if (!query.trim()) return
  const existing = recentSearches.value.indexOf(query)
  if (existing !== -1) {
    recentSearches.value.splice(existing, 1)
  }
  recentSearches.value.unshift(query)
  if (recentSearches.value.length > 5) {
    recentSearches.value.pop()
  }
}

function useRecentSearch(query) {
  searchQuery.value = query
}

// 转义 HTML 特殊字符
function escapeHtml(str) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

// 转义正则特殊字符
function escapeRegex(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

// 高亮匹配文本
function highlightMatch(text, query) {
  if (!query.trim()) return escapeHtml(text)
  const escaped = escapeHtml(text)
  const regex = new RegExp(`(${escapeRegex(query)})`, 'gi')
  return escaped.replace(regex, '<mark>$1</mark>')
}
</script>

<template>
  <div class="search-overlay" @click.self="$emit('close')">
    <div class="search-container glass-panel">
      <!-- 搜索头部 -->
      <div class="search-header">
        <div class="search-title">🔍 搜索 / SEARCH</div>
        <button class="close-btn" @click="$emit('close')">×</button>
      </div>

      <!-- 搜索框 -->
      <div class="search-input-wrapper">
        <NInput
          v-model:value="searchQuery"
          placeholder="输入关键词搜索笔记..."
          size="large"
          autofocus
          clearable
          @keyup.enter="addToRecent(searchQuery)"
        />
      </div>

      <!-- 搜索结果 -->
      <div class="search-results">
        <!-- 加载状态 -->
        <div v-if="isSearching" class="search-loading">
          <NSpin size="medium" />
          <span>搜索中...</span>
        </div>

        <!-- 搜索结果列表 -->
        <div v-else-if="searchResults.length > 0" class="results-list">
          <div class="results-count">找到 {{ searchResults.length }} 条结果</div>
          <div
            v-for="note in searchResults"
            :key="note.id"
            class="result-item"
            @click="openNote(note)"
          >
            <div class="result-title" v-html="highlightMatch(note.title, searchQuery)"></div>
            <div class="result-content" v-if="note.content" v-html="highlightMatch(note.content.substring(0, 100), searchQuery)"></div>
          </div>
        </div>

        <!-- 无结果 -->
        <div v-else-if="searchQuery.trim()" class="no-results">
          <NEmpty description="未找到相关笔记">
            <template #extra>
              <span class="hint-text">尝试使用不同的关键词</span>
            </template>
          </NEmpty>
        </div>

        <!-- 最近搜索 -->
        <div v-else class="recent-searches">
          <div class="section-title">最近搜索</div>
          <div class="recent-list">
            <div
              v-for="(query, idx) in recentSearches"
              :key="idx"
              class="recent-item"
              @click="useRecentSearch(query)"
            >
              <span class="recent-icon">🕐</span>
              <span class="recent-text">{{ query }}</span>
            </div>
          </div>

          <div class="section-title" style="margin-top: 30px;">热门标签</div>
          <div class="tags-cloud">
            <span class="tag-item"># 项目</span>
            <span class="tag-item"># 学习</span>
            <span class="tag-item"># 工作</span>
            <span class="tag-item"># 总结</span>
            <span class="tag-item"># 读书</span>
          </div>
        </div>
      </div>

      <!-- 搜索提示 -->
      <div class="search-footer">
        <span class="hint">按 Enter 键保存搜索历史</span>
        <span class="shortcut">ESC 关闭</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(10px);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 10vh;
  z-index: 200;
}

.search-container {
  width: 650px;
  max-height: 80vh;
  border-radius: 16px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.search-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 25px;
  border-bottom: 1px solid var(--glass-border);
}

.search-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.close-btn {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  font-size: 24px;
  cursor: pointer;
  border-radius: 6px;
  transition: 0.2s;
}
.close-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary);
}

.search-input-wrapper {
  padding: 20px 25px;
}

.search-results {
  flex: 1;
  overflow-y: auto;
  padding: 0 25px 20px;
  min-height: 200px;
  max-height: 50vh;
}

.search-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 40px;
  color: var(--text-secondary);
}

.results-count {
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 15px;
  opacity: 0.7;
}

.result-item {
  padding: 15px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--glass-border);
  margin-bottom: 10px;
  cursor: pointer;
  transition: 0.2s;
}
.result-item:hover {
  background: rgba(59, 130, 246, 0.1);
  border-color: var(--accent);
}

.result-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.result-content {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

:deep(mark) {
  background: rgba(59, 130, 246, 0.3);
  color: var(--accent);
  border-radius: 2px;
  padding: 0 2px;
}

.no-results {
  padding: 40px;
  text-align: center;
}

.hint-text {
  font-size: 12px;
  color: var(--text-secondary);
  opacity: 0.6;
}

.recent-searches {
  padding: 10px 0;
}

.section-title {
  font-size: 11px;
  color: var(--text-secondary);
  letter-spacing: 1px;
  margin-bottom: 12px;
  opacity: 0.6;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.recent-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: 0.2s;
}
.recent-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.recent-icon { font-size: 12px; }
.recent-text { font-size: 13px; color: var(--text-primary); }

.tags-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-item {
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 20px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: 0.2s;
}
.tag-item:hover {
  background: rgba(59, 130, 246, 0.1);
  border-color: var(--accent);
  color: var(--accent);
}

.search-footer {
  padding: 12px 25px;
  border-top: 1px solid var(--glass-border);
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.5;
}

.shortcut {
  background: rgba(255, 255, 255, 0.05);
  padding: 2px 8px;
  border-radius: 4px;
}
</style>