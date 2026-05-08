<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { NInput, NButton, NEmpty, NSpin } from 'naive-ui'
import { SearchNotes, GetTagsWithCount, GetNotesByTag } from '../../wailsjs/go/main/App'

const props = defineProps(['user'])
const emit = defineEmits(['close', 'open-note'])

const searchQuery = ref('')
const isSearching = ref(false)
const searchResults = ref([])
const recentSearches = ref([])
const hotTags = ref([])
const showContent = ref(false)

async function loadHotTags() {
  try {
    const tags = await GetTagsWithCount()
    hotTags.value = tags || []
  } catch (e) {
    hotTags.value = []
  }
}

onMounted(() => {
  loadHotTags()
  setTimeout(() => { showContent.value = true }, 50)
})

function getDelay(index) {
  return (index * 50) + 'ms'
}

async function searchByTag(tag) {
  isSearching.value = true
  searchResults.value = []
  try {
    const results = await GetNotesByTag(tag.id)
    searchResults.value = results || []
  } catch (e) {
    searchResults.value = []
  }
  isSearching.value = false
}

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
    searchResults.value = []
  }
  isSearching.value = false
}

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

function escapeHtml(str) {
  return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}

function escapeRegex(str) {
  return str.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function highlightMatch(text, query) {
  if (!query.trim()) return escapeHtml(text)
  const escaped = escapeHtml(text)
  const regex = new RegExp(`(${escapeRegex(query)})`, 'gi')
  return escaped.replace(regex, '<mark>$1</mark>')
}

function handleOverlayClick(e) {
  if (e.target === e.currentTarget) {
    emit('close')
  }
}
</script>

<template>
  <Transition name="modal-fade">
    <div v-if="showContent" class="search-overlay" @click="handleOverlayClick">
      <div class="search-container glass-panel">
        <!-- Header -->
        <div class="search-header">
          <div class="header-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="11" cy="11" r="8"/>
              <path d="M21 21l-4.35-4.35"/>
            </svg>
            <span>搜索</span>
          </div>
          <button class="close-btn" @click="$emit('close')">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
          </button>
        </div>

        <!-- Search input -->
        <div class="search-input-wrapper">
          <div class="input-container">
            <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
              <circle cx="11" cy="11" r="8"/>
              <path d="M21 21l-4.35-4.35"/>
            </svg>
            <input
              v-model="searchQuery"
              class="search-input"
              placeholder="输入关键词搜索笔记..."
              autofocus
              @keyup.enter="addToRecent(searchQuery)"
            />
            <button v-if="searchQuery" class="clear-btn" @click="searchQuery = ''">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"/>
                <line x1="6" y1="6" x2="18" y2="18"/>
              </svg>
            </button>
          </div>
        </div>

        <!-- Results area -->
        <div class="search-results">
          <!-- Loading -->
          <div v-if="isSearching" class="search-loading">
            <div class="loading-spinner"></div>
            <span>正在搜索...</span>
          </div>

          <!-- Results list -->
          <div v-else-if="searchResults.length > 0" class="results-list">
            <div class="results-meta">找到 {{ searchResults.length }} 条结果</div>
            <div
              v-for="(note, index) in searchResults"
              :key="note.id"
              class="result-item"
              :style="{ animationDelay: getDelay(index) }"
              @click="openNote(note)"
            >
              <div class="result-icon">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                  <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                </svg>
              </div>
              <div class="result-content">
                <div class="result-title" v-html="highlightMatch(note.title, searchQuery)"></div>
                <div class="result-excerpt" v-if="note.content" v-html="highlightMatch(note.content.substring(0, 120), searchQuery)"></div>
              </div>
              <svg class="result-arrow" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M9 18l6-6-6-6"/>
              </svg>
            </div>
          </div>

          <!-- No results -->
          <div v-else-if="searchResults.length === 0 && searchQuery.trim()" class="no-results">
            <div class="no-results-icon">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
                <circle cx="11" cy="11" r="8"/>
                <path d="M21 21l-4.35-4.35"/>
                <path d="M8 8l6 6M14 8l-6 6" opacity="0.5"/>
              </svg>
            </div>
            <p class="no-results-text">未找到相关笔记</p>
            <span class="no-results-hint">尝试使用不同的关键词</span>
          </div>

          <!-- Default state - recent searches & hot tags -->
          <div v-else-if="searchResults.length === 0 && !searchQuery.trim()" class="default-state">
            <!-- Recent searches -->
            <div v-if="recentSearches.length > 0" class="recent-section">
              <div class="section-label">最近搜索</div>
              <div class="recent-list">
                <button
                  v-for="(query, idx) in recentSearches"
                  :key="idx"
                  class="recent-item"
                  @click="useRecentSearch(query)"
                >
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <circle cx="12" cy="12" r="10"/>
                    <polyline points="12 6 12 12 16 14"/>
                  </svg>
                  <span>{{ query }}</span>
                </button>
              </div>
            </div>

            <!-- Hot tags -->
            <div v-if="hotTags.length > 0" class="tags-section">
              <div class="section-label">热门标签</div>
              <div class="tags-cloud">
                <button
                  v-for="tag in hotTags"
                  :key="tag.id"
                  class="tag-item"
                  @click="searchByTag(tag)"
                >
                  <span class="tag-hash">#</span>
                  <span class="tag-name">{{ tag.name }}</span>
                  <span class="tag-count">{{ tag.noteCount }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="search-footer">
          <div class="footer-hint">
            <kbd>Enter</kbd>
            <span>保存搜索历史</span>
          </div>
          <div class="footer-shortcut">
            <kbd>ESC</kbd>
            <span>关闭</span>
          </div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
.search-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(16px);
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding-top: 10vh;
  z-index: 200;
}

.search-container {
  width: min(580px, 92vw);
  border-radius: 20px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 40px 100px rgba(0, 0, 0, 0.5), 0 0 0 1px var(--glass-border);
}

/* Header */
.search-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--glass-border);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.header-title svg {
  color: var(--accent);
}

.close-btn {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border: none;
  border-radius: 8px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

/* Input */
.search-input-wrapper {
  padding: 16px 24px;
}

.input-container {
  display: flex;
  align-items: center;
  gap: 12px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  padding: 12px 16px;
  transition: all 0.2s;
}

.input-container:focus-within {
  border-color: var(--accent);
  background: rgba(var(--accent-rgb), 0.05);
}

.input-icon {
  color: var(--text-secondary);
  flex-shrink: 0;
}

.search-input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  font-size: 15px;
  color: var(--text-primary);
}

.search-input::placeholder {
  color: var(--text-secondary);
  opacity: 0.5;
}

.clear-btn {
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.06);
  border: none;
  border-radius: 6px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.clear-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}

/* Results */
.search-results {
  flex: 1;
  overflow-y: auto;
  padding: 0 24px 20px;
  min-height: 280px;
  max-height: 50vh;
}

.search-results::-webkit-scrollbar {
  width: 4px;
}

.search-results::-webkit-scrollbar-track {
  background: transparent;
}

.search-results::-webkit-scrollbar-thumb {
  background: var(--glass-border);
  border-radius: 2px;
}

/* Loading */
.search-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 50px;
  color: var(--text-secondary);
}

.loading-spinner {
  width: 20px;
  height: 20px;
  border: 2px solid var(--glass-border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Results list */
.results-meta {
  font-size: 11px;
  color: var(--text-secondary);
  letter-spacing: 0.5px;
  margin-bottom: 12px;
  opacity: 0.7;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 14px 16px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--glass-border);
  border-radius: 12px;
  margin-bottom: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  opacity: 0;
  animation: itemFadeIn 0.3s ease-out forwards;
}

@keyframes itemFadeIn {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: translateY(0); }
}

.result-item:hover {
  background: rgba(var(--accent-rgb), 0.08);
  border-color: rgba(var(--accent-rgb), 0.3);
  transform: translateX(4px);
}

.result-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(var(--accent-rgb), 0.1);
  border-radius: 10px;
  color: var(--accent);
  flex-shrink: 0;
}

.result-content {
  flex: 1;
  min-width: 0;
}

.result-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.result-excerpt {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-arrow {
  color: var(--text-secondary);
  opacity: 0;
  transition: all 0.2s;
  flex-shrink: 0;
}

.result-item:hover .result-arrow {
  opacity: 0.6;
  transform: translateX(2px);
}

:deep(mark) {
  background: rgba(var(--accent-rgb), 0.25);
  color: var(--accent);
  border-radius: 3px;
  padding: 0 2px;
}

/* No results */
.no-results {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 50px 20px;
  text-align: center;
}

.no-results-icon {
  color: var(--text-secondary);
  opacity: 0.3;
  margin-bottom: 16px;
}

.no-results-text {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0 0 6px 0;
}

.no-results-hint {
  font-size: 12px;
  color: var(--text-secondary);
  opacity: 0.6;
}

/* Default state */
.default-state {
  padding: 8px 0;
}

.section-label {
  font-size: 11px;
  color: var(--text-secondary);
  letter-spacing: 1px;
  text-transform: uppercase;
  margin-bottom: 12px;
  opacity: 0.6;
}

.recent-section {
  margin-bottom: 24px;
}

.recent-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.recent-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  background: transparent;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
  color: inherit;
  text-align: left;
  width: 100%;
}

.recent-item:hover {
  background: rgba(255, 255, 255, 0.05);
}

.recent-item svg {
  color: var(--text-secondary);
  opacity: 0.5;
}

.recent-item span {
  font-size: 13px;
  color: var(--text-primary);
}

/* Tags */
.tags-section {
  padding-top: 8px;
}

.tags-cloud {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--glass-border);
  border-radius: 20px;
  font-size: 12px;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
  color: inherit;
}

.tag-item:hover {
  background: rgba(var(--accent-rgb), 0.1);
  border-color: var(--accent);
  transform: translateY(-2px);
}

.tag-hash {
  color: var(--accent);
  font-weight: 600;
}

.tag-name {
  color: var(--text-primary);
}

.tag-count {
  color: var(--text-secondary);
  font-size: 11px;
  opacity: 0.6;
  background: rgba(255, 255, 255, 0.06);
  padding: 1px 6px;
  border-radius: 8px;
}

/* Footer */
.search-footer {
  display: flex;
  justify-content: space-between;
  padding: 14px 24px;
  border-top: 1px solid var(--glass-border);
}

.footer-hint, .footer-shortcut {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.5;
}

.footer-hint kbd, .footer-shortcut kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 18px;
  padding: 0 5px;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid var(--glass-border);
  border-radius: 4px;
  font-size: 10px;
  font-family: inherit;
}

/* Modal transition */
.modal-fade-enter-active {
  transition: all 0.3s ease-out;
}

.modal-fade-leave-active {
  transition: all 0.2s ease-in;
}

.modal-fade-enter-from {
  opacity: 0;
}

.modal-fade-enter-from .search-container {
  transform: scale(0.95) translateY(-20px);
  opacity: 0;
}

.modal-fade-leave-to {
  opacity: 0;
}

.modal-fade-leave-to .search-container {
  transform: scale(0.95);
  opacity: 0;
}
</style>