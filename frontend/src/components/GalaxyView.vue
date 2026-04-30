<script setup>
import { onMounted, ref, nextTick } from 'vue'
import * as d3 from 'd3'
import { GetDirectoryTree } from '../../wailsjs/go/main/App'

const props = defineProps(['user'])
const emit = defineEmits(['back', 'open-note'])

const svgRef = ref(null)
const showUserCard = ref(false)
const selectedUser = ref(null)
const stats = ref({ noteCount: 0, galaxyCount: 0, tagCount: 0, totalSize: 0 })

let simulation = null

onMounted(async () => {
  await loadData()
  await nextTick()
  initGraph()
})

async function loadData() {
  console.log('=== loadData called ===')
  try {
    // 获取目录树
    const treeData = await GetDirectoryTree()
    console.log('GetDirectoryTree result:', treeData)

    // 统计信息 - 计算总节点数
    let nodeCount = 0
    function countNodes(node) {
      if (!node) return
      nodeCount++
      if (node.children) {
        node.children.forEach(countNodes)
      }
    }
    countNodes(treeData)

    stats.value = {
      noteCount: nodeCount - 1, // 减去 root 节点
      galaxyCount: 1,
      tagCount: 0,
      totalSize: 0
    }

    // 保存树数据到全局
    window.galaxyTreeData = treeData
  } catch (e) {
    console.error('Failed to load galaxy data:', e)
    window.galaxyTreeData = null
  }
}

// 获取节点半径
function getNodeRadius(d) {
  if (d.data.type === 'user') return 35
  if (d.data.type === 'folder') return 22
  return 18
}

// 计算连线与圆边界的交点
function getCircleEdgePoint(x1, y1, x2, y2, cx, cy, r) {
  const dx = x2 - x1
  const dy = y2 - y1
  const fx = x1 - cx
  const fy = y1 - cy

  const a = dx * dx + dy * dy
  const b = 2 * (fx * dx + fy * dy)
  const c = fx * fx + fy * fy - r * r

  const discriminant = b * b - 4 * a * c
  if (discriminant < 0) return { x: x2, y: y2 }

  const t = (-b - Math.sqrt(discriminant)) / (2 * a)
  if (t < 0 || t > 1) return { x: x2, y: y2 }

  return {
    x: x1 + t * dx,
    y: y1 + t * dy
  }
}

function initGraph() {
  if (!svgRef.value) return

  const treeData = window.galaxyTreeData
  if (!treeData) return

  const svg = d3.select(svgRef.value)
  const width = svgRef.value.clientWidth
  const height = svgRef.value.clientHeight

  svg.attr('width', width).attr('height', height)

  // 清空之前的元素
  svg.selectAll('*').remove()

  // 添加缩放行为
  const zoom = d3.zoom()
    .scaleExtent([0.3, 3])
    .on('zoom', (event) => {
      container.attr('transform', event.transform)
    })

  svg.call(zoom)

  // 创建容器组
  const container = svg.append('g')

  // 找到用户节点作为树的根
  let userNode = null
  if (treeData.type === 'user') {
    userNode = treeData
  } else if (treeData.children) {
    userNode = treeData.children.find(c => c.type === 'user')
  }

  // 如果找到用户节点，用它作为根构建树
  const root = userNode
    ? d3.hierarchy(userNode)
    : d3.hierarchy(treeData)

  const links = root.links()
  const nodes = root.descendants()

  console.log('Hierarchy nodes:', nodes)
  console.log('Hierarchy links:', links)

  // 创建力模拟 - 使用树的父子关系
  simulation = d3.forceSimulation(nodes)
    .force('link', d3.forceLink(links).id(d => d.id).distance(100).strength(1))
    .force('charge', d3.forceManyBody().strength(-250))
    .force('center', d3.forceCenter(width / 2, height / 2))
    .force('collision', d3.forceCollide().radius(45))
    .force('y', d3.forceY(height / 2).strength(0.1))

  // 绘制链接
  const link = container.append('g')
    .selectAll('line')
    .data(links)
    .join('line')
    .attr('stroke', 'rgba(59, 130, 246, 0.4)')
    .attr('stroke-width', 1.5)

  // 绘制节点组
  const node = container.append('g')
    .selectAll('g')
    .data(nodes)
    .join('g')
    .attr('cursor', 'pointer')
    .call(d3.drag()
      .on('start', dragstarted)
      .on('drag', dragged)
      .on('end', dragended))

  // 用户节点 - 特殊样式（更大，带头像支持）
  node.filter(d => d.data.type === 'user')
    .append('circle')
    .attr('r', 35)
    .attr('fill', 'rgba(59, 130, 246, 0.2)')
    .attr('stroke', '#3b82f6')
    .attr('stroke-width', 2)

  // 用户节点头像（优先用 d.data.avatar，其次用 props.user.avatar）
  node.filter(d => d.data.type === 'user' && (d.data.avatar || props.user.avatar))
    .append('clipPath')
    .attr('id', d => `clip-${d.id}`)
    .append('circle')
    .attr('r', 33)

  node.filter(d => d.data.type === 'user' && (d.data.avatar || props.user.avatar))
    .append('image')
    .attr('href', d => d.data.avatar || props.user.avatar)
    .attr('x', -33)
    .attr('y', -33)
    .attr('width', 66)
    .attr('height', 66)
    .attr('clip-path', d => `url(#clip-${d.id})`)
    .attr('preserveAspectRatio', 'xMidYMid slice')

  // 用户节点首字母（当没有头像时显示）
  node.filter(d => d.data.type === 'user' && !d.data.avatar && !props.user.avatar)
    .append('text')
    .attr('text-anchor', 'middle')
    .attr('dominant-baseline', 'central')
    .attr('fill', '#3b82f6')
    .attr('font-size', '24px')
    .attr('font-weight', 'bold')
    .text(d => d.data.name ? d.data.name[0].toUpperCase() : 'V')

  // 文件夹节点 - 绿色边框
  node.filter(d => d.data.type === 'folder')
    .append('circle')
    .attr('r', 22)
    .attr('fill', 'rgba(16, 185, 129, 0.15)')
    .attr('stroke', '#10b981')
    .attr('stroke-width', 1.5)

  // 笔记节点 - 蓝色边框
  node.filter(d => d.data.type === 'note')
    .append('circle')
    .attr('r', 18)
    .attr('fill', 'rgba(59, 130, 246, 0.15)')
    .attr('stroke', '#3b82f6')
    .attr('stroke-width', 1.5)

  // 节点首字母（用户节点除外）
  node.filter(d => d.data.type !== 'user')
    .append('text')
    .attr('text-anchor', 'middle')
    .attr('dominant-baseline', 'central')
    .attr('fill', d => d.data.type === 'note' ? '#3b82f6' : '#10b981')
    .attr('font-size', '12px')
    .attr('font-weight', 'bold')
    .text(d => d.data.name ? d.data.name[0].toUpperCase() : '?')

  // 节点名称标签
  node.filter(d => d.data.type === 'user')
    .append('text')
    .attr('text-anchor', 'middle')
    .attr('y', 50)
    .attr('fill', '#cbd5e1')
    .attr('font-size', '12px')
    .attr('font-weight', '500')
    .text(d => {
      const name = d.data.name || 'VOYAGER'
      return name.length > 8 ? name.substring(0, 8) + '...' : name
    })

  node.filter(d => d.data.type !== 'user')
    .append('text')
    .attr('text-anchor', 'middle')
    .attr('y', 35)
    .attr('fill', '#94a3b8')
    .attr('font-size', '10px')
    .text(d => d.data.name ? (d.data.name.length > 6 ? d.data.name.substring(0, 6) + '...' : d.data.name) : '')

  // 点击事件
  node.on('click', (event, d) => {
    if (d.data.type === 'user') {
      selectedUser.value = {
        username: d.data.name,
        avatar: d.data.avatar,
        title: '星际漫游者',
        noteCount: stats.value.noteCount,
        galaxyCount: stats.value.galaxyCount,
        tagCount: stats.value.tagCount,
        totalSize: stats.value.totalSize
      }
      showUserCard.value = true
    } else if (d.data.type === 'note') {
      emit('open-note', { id: d.data.id, name: d.data.name })
    }
  })

  // 更新位置 - 连线在节点边界停止
  simulation.on('tick', () => {
    link.each(function(d) {
      const sourceRadius = getNodeRadius(d.source)
      const targetRadius = getNodeRadius(d.target)

      const start = getCircleEdgePoint(
        d.target.x, d.target.y, d.source.x, d.source.y,
        d.source.x, d.source.y, sourceRadius
      )
      const end = getCircleEdgePoint(
        d.source.x, d.source.y, d.target.x, d.target.y,
        d.target.x, d.target.y, targetRadius
      )

      d3.select(this)
        .attr('x1', start.x)
        .attr('y1', start.y)
        .attr('x2', end.x)
        .attr('y2', end.y)
    })

    node.attr('transform', d => `translate(${d.x},${d.y})`)
  })

  function dragstarted(event, d) {
    if (!event.active) simulation.alphaTarget(0.3).restart()
    d.fx = d.x
    d.fy = d.y
  }

  function dragged(event, d) {
    d.fx = event.x
    d.fy = event.y
  }

  function dragended(event, d) {
    if (!event.active) simulation.alphaTarget(0)
    d.fx = null
    d.fy = null
  }
}

function closeUserCard() {
  showUserCard.value = false
}
</script>

<template>
  <div class="galaxy-container">
    <button class="back-btn" @click="$emit('back')">← 返回主页 / EXIT GALAXY</button>

    <svg ref="svgRef" class="galaxy-graph"></svg>

    <!-- 用户信息卡片 -->
    <Transition name="fade">
      <div v-if="showUserCard" class="user-card-overlay" @click.self="closeUserCard">
        <div class="user-card glass-panel">
          <button class="card-close" @click="closeUserCard">×</button>

          <div class="card-header">
            <img v-if="selectedUser.avatar" :src="selectedUser.avatar" class="card-avatar" />
            <div v-else class="card-avatar-placeholder">{{ selectedUser.username?.[0] || 'V' }}</div>
            <div class="card-title">{{ selectedUser.title }}</div>
          </div>

          <div class="card-body">
            <div class="card-name">{{ selectedUser.username }}</div>
            <div class="card-motto">{{ user.motto || '记录有序的人生' }}</div>
          </div>

          <div class="card-divider"></div>

          <div class="card-stats">
            <div class="stat-row">
              <span class="stat-icon">📄</span>
              <span class="stat-label">笔记</span>
              <span class="stat-value">{{ selectedUser.noteCount }} 篇</span>
            </div>
            <div class="stat-row">
              <span class="stat-icon">🌌</span>
              <span class="stat-label">星系</span>
              <span class="stat-value">{{ selectedUser.galaxyCount }} 个</span>
            </div>
            <div class="stat-row">
              <span class="stat-icon">🏷️</span>
              <span class="stat-label">标签</span>
              <span class="stat-value">{{ selectedUser.tagCount }} 个</span>
            </div>
            <div class="stat-row">
              <span class="stat-icon">💾</span>
              <span class="stat-label">占用</span>
              <span class="stat-value">{{ selectedUser.totalSize }} MB</span>
            </div>
          </div>

          <div class="card-footer">
            <div class="title-hint">称号：{{ selectedUser.title }}</div>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style>
.galaxy-graph canvas {
  display: block;
  user-select: none;
  outline: none;
  -webkit-tap-highlight-color: transparent;
}
.galaxy-container {
  width: 100%;
  height: 100%;
  background: var(--bg-app);
  position: relative;
}
.galaxy-graph {
  width: 100%;
  height: 100%;
}
.back-btn {
  position: absolute;
  top: 30px;
  left: 30px;
  z-index: 10;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  color: var(--text-primary);
  padding: 10px 20px;
  border-radius: 8px;
  cursor: pointer;
  backdrop-filter: blur(10px);
  transition: 0.2s;
}
.back-btn:hover {
  background: var(--accent);
  color: #fff;
}

.user-card-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 100;
}

.user-card {
  width: 320px;
  padding: 30px;
  border-radius: 20px;
  position: relative;
}

.card-close {
  position: absolute;
  top: 15px;
  right: 15px;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  font-size: 20px;
  cursor: pointer;
}
.card-close:hover { color: var(--text-primary); }

.card-header {
  text-align: center;
  margin-bottom: 20px;
}

.card-avatar {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  object-fit: cover;
  border: 3px solid var(--accent);
  margin-bottom: 15px;
}

.card-avatar-placeholder {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--accent);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: 800;
  margin: 0 auto 15px;
}

.card-title {
  font-size: 12px;
  color: var(--accent);
  letter-spacing: 2px;
  text-transform: uppercase;
}

.card-body {
  text-align: center;
  margin-bottom: 20px;
}

.card-name {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.card-motto {
  font-size: 13px;
  color: var(--text-secondary);
  font-style: italic;
}

.card-divider {
  height: 1px;
  background: var(--glass-border);
  margin: 20px 0;
}

.card-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 15px;
}

.stat-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.stat-icon { font-size: 14px; }
.stat-label { color: var(--text-secondary); flex: 1; }
.stat-value { color: var(--text-primary); font-weight: 600; }

.card-footer {
  margin-top: 20px;
  text-align: center;
}

.title-hint {
  font-size: 11px;
  color: var(--text-secondary);
  opacity: 0.7;
}

/* 动画 */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>