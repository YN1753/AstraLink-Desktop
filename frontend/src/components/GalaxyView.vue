<script setup>
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import * as d3 from 'd3'
import { GetNodeByPath, CreateGalaxy, CreateNote } from '../../wailsjs/go/main/App'

const props = defineProps(['user'])
const emit = defineEmits(['back', 'open-note'])

const canvasRef = ref(null)
const nodes = ref([])
const loading = ref(true)
const showCreateMenu = ref(false)
const createType = ref('')

let simulation = null
let canvas = null
let ctx = null
let transform = d3.zoomIdentity
let graphNodes = []
let graphLinks = []
let width = 0
let height = 0
let hoveredNode = null
let dragMoved = false
let avatarImg = null

const NODE_RADIUS = { user: 36, galaxy: 28, note: 18 }

// Read theme colors from CSS variables
function getThemeColors() {
  const style = getComputedStyle(document.documentElement)
  const accent = style.getPropertyValue('--accent').trim() || '#6366f1'
  const textPrimary = style.getPropertyValue('--text-primary').trim() || '#ffffff'
  const textSecondary = style.getPropertyValue('--text-secondary').trim() || '#94a3b8'
  const bgApp = style.getPropertyValue('--bg-app').trim() || '#050505'

  // Derive node colors from accent
  return {
    user: {
      fill: bgApp,
      stroke: accent,
      glow: accent + '40', // 25% opacity
      label: textPrimary
    },
    galaxy: {
      fill: bgApp,
      stroke: accent,
      glow: accent + '30',
      label: textSecondary
    },
    note: {
      fill: bgApp,
      stroke: accent,
      glow: accent + '25',
      label: textSecondary
    },
    link: textPrimary + '14', // ~8% opacity
    linkHover: textPrimary + '30',
    textPrimary,
    textSecondary,
    accent
  }
}

function loadAvatar() {
  avatarImg = null
  if (!props.user.avatar) return
  const img = new Image()
  img.onload = () => {
    avatarImg = img
    render()
  }
  img.src = props.user.avatar
}

function buildTree(allNodes) {
  // Find the user node (path === 'root')
  const userNode = allNodes.find(n => n.path === 'root')
  // Children: everything except root
  const children = allNodes.filter(n => n.path !== 'root')

  const graphNodes = []
  const graphLinks = []
  const added = new Set()

  // Add user node at center
  if (userNode) {
    graphNodes.push({
      id: userNode.id,
      name: userNode.name || props.user.username || '用户',
      type: 'user',
      path: userNode.path,
      others: userNode.others,
      radius: NODE_RADIUS.user,
      x: width / 2,
      y: height / 2
    })
    added.add(userNode.id)
  }

  // Build path map for children
  const pathMap = new Map()
  children.forEach(n => pathMap.set(n.path, n))

  children.forEach(n => {
    if (!added.has(n.id)) {
      graphNodes.push({
        id: n.id,
        name: n.name || '未命名',
        type: n.type,
        path: n.path,
        others: n.others,
        radius: NODE_RADIUS[n.type] || 18
      })
      added.add(n.id)
    }

    // Find parent by path hierarchy
    const parts = n.path.split('/')
    if (parts.length > 2) {
      const parentPath = parts.slice(0, -1).join('/')
      const parent = pathMap.get(parentPath)
      if (parent) {
        graphLinks.push({ source: parent.id, target: n.id })
      } else if (parentPath === 'root' && userNode) {
        graphLinks.push({ source: userNode.id, target: n.id })
      }
    } else if (parts.length === 2 && parts[0] === 'root' && userNode) {
      graphLinks.push({ source: userNode.id, target: n.id })
    }
  })

  return { nodes: graphNodes, links: graphLinks }
}

async function loadNodes() {
  loading.value = true
  try {
    const allNodes = await GetNodeByPath('root')
    nodes.value = allNodes || []
    const tree = buildTree(nodes.value)
    graphNodes = tree.nodes
    graphLinks = tree.links
    initGraph()
  } catch (e) {
    console.error('Failed to load nodes:', e)
  } finally {
    loading.value = false
  }
}

function initGraph() {
  const container = canvasRef.value
  if (!container) return

  width = container.clientWidth
  height = container.clientHeight

  d3.select(container).select('canvas').remove()

  canvas = d3.select(container)
    .append('canvas')
    .attr('width', width * devicePixelRatio)
    .attr('height', height * devicePixelRatio)
    .style('width', width + 'px')
    .style('height', height + 'px')

  ctx = canvas.node().getContext('2d')
  ctx.scale(devicePixelRatio, devicePixelRatio)

  // Drag — register BEFORE zoom so it fires first
  let dragTarget = null
  let dragStartX = 0
  let dragStartY = 0

  canvas.on('mousedown.drag', (event) => {
    const [mx, my] = d3.pointer(event, canvas.node())
    const [sx, sy] = transform.invert([mx, my])
    const node = findNode(sx, sy)
    if (!node) return

    // Block zoom from seeing this event
    event.stopImmediatePropagation()
    event.preventDefault()

    dragTarget = node
    dragMoved = false
    dragStartX = mx
    dragStartY = my
    simulation.alphaTarget(0.3).restart()
    node.fx = node.x
    node.fy = node.y

    const onMove = (event2) => {
      const [mmx, mmy] = d3.pointer(event2, canvas.node())
      if (!dragMoved && (Math.abs(mmx - dragStartX) > 3 || Math.abs(mmy - dragStartY) > 3)) {
        dragMoved = true
      }
      const [ix, iy] = transform.invert([mmx, mmy])
      dragTarget.fx = ix
      dragTarget.fy = iy
    }

    const onUp = () => {
      simulation.alphaTarget(0)
      if (dragTarget) {
        dragTarget.fx = null
        dragTarget.fy = null
      }
      dragTarget = null
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }

    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
  })

  // Zoom — registered after drag
  const zoom = d3.zoom()
    .scaleExtent([0.3, 4])
    .on('zoom', (event) => {
      transform = event.transform
      render()
    })

  canvas.call(zoom)
    .on('click', handleClick)
    .on('mousemove', handleMouseMove)
    .on('mouseleave', () => { hoveredNode = null; render() })

  // Forces — pull toward center
  simulation = d3.forceSimulation(graphNodes)
    .force('link', d3.forceLink(graphLinks).id(d => d.id).distance(100).strength(0.8))
    .force('charge', d3.forceManyBody().strength(-200))
    .force('x', d3.forceX(width / 2).strength(0.05))
    .force('y', d3.forceY(height / 2).strength(0.05))
    .on('tick', render)
}

function findNode(x, y) {
  for (let i = graphNodes.length - 1; i >= 0; i--) {
    const n = graphNodes[i]
    const dx = x - n.x
    const dy = y - n.y
    if (dx * dx + dy * dy < n.radius * n.radius) return n
  }
  return null
}

function handleClick(event) {
  if (dragMoved) return
  const [x, y] = transform.invert([event.offsetX, event.offsetY])
  const node = findNode(x, y)
  if (node && node.type === 'note') {
    emit('open-note', { id: node.id, name: node.name })
  }
}

function handleMouseMove(event) {
  const [x, y] = transform.invert([event.offsetX, event.offsetY])
  const node = findNode(x, y)
  if (node !== hoveredNode) {
    hoveredNode = node
    canvas.style('cursor', node && node.type === 'note' ? 'pointer' : 'default')
    render()
  }
}

function render() {
  if (!ctx) return
  const colors = getThemeColors()

  ctx.save()
  ctx.clearRect(0, 0, width, height)
  ctx.translate(transform.x, transform.y)
  ctx.scale(transform.k, transform.k)

  // Draw links
  graphLinks.forEach(link => {
    const source = typeof link.source === 'object' ? link.source : graphNodes.find(n => n.id === link.source)
    const target = typeof link.target === 'object' ? link.target : graphNodes.find(n => n.id === link.target)
    if (!source || !target) return

    ctx.beginPath()
    ctx.moveTo(source.x, source.y)
    ctx.lineTo(target.x, target.y)
    ctx.strokeStyle = colors.link
    ctx.lineWidth = 1.5
    ctx.setLineDash([4, 4])
    ctx.stroke()
    ctx.setLineDash([])
  })

  // Draw nodes
  graphNodes.forEach(n => {
    const isHovered = hoveredNode === n
    const typeColors = colors[n.type] || colors.note
    const r = n.radius

    // Glow on hover
    if (isHovered) {
      ctx.beginPath()
      ctx.arc(n.x, n.y, r + 14, 0, Math.PI * 2)
      const gradient = ctx.createRadialGradient(n.x, n.y, r, n.x, n.y, r + 14)
      gradient.addColorStop(0, typeColors.glow)
      gradient.addColorStop(1, 'transparent')
      ctx.fillStyle = gradient
      ctx.fill()
    }

    // Node circle
    ctx.beginPath()
    ctx.arc(n.x, n.y, r, 0, Math.PI * 2)
    ctx.fillStyle = typeColors.fill
    ctx.fill()
    ctx.strokeStyle = isHovered ? colors.textPrimary : typeColors.stroke
    ctx.lineWidth = isHovered ? 2.5 : 1.5
    ctx.stroke()

    // User node: avatar or fallback icon
    if (n.type === 'user') {
      if (avatarImg) {
        ctx.save()
        ctx.beginPath()
        ctx.arc(n.x, n.y, r - 2, 0, Math.PI * 2)
        ctx.clip()
        ctx.drawImage(avatarImg, n.x - r + 2, n.y - r + 2, (r - 2) * 2, (r - 2) * 2)
        ctx.restore()
        // Redraw border on top
        ctx.beginPath()
        ctx.arc(n.x, n.y, r - 2, 0, Math.PI * 2)
        ctx.strokeStyle = isHovered ? colors.textPrimary : typeColors.stroke
        ctx.lineWidth = isHovered ? 2.5 : 1.5
        ctx.stroke()
      } else {
        // Fallback: head + shoulders icon
        ctx.strokeStyle = typeColors.stroke
        ctx.lineWidth = 1.5
        ctx.beginPath()
        ctx.arc(n.x, n.y - 6, 7, 0, Math.PI * 2)
        ctx.stroke()
        ctx.beginPath()
        ctx.arc(n.x, n.y + 14, 13, Math.PI * 1.15, Math.PI * 1.85)
        ctx.stroke()
      }
    }

    // Galaxy: inner ring
    if (n.type === 'galaxy') {
      ctx.beginPath()
      ctx.arc(n.x, n.y, r * 0.55, 0, Math.PI * 2)
      ctx.strokeStyle = typeColors.stroke
      ctx.lineWidth = 0.8
      ctx.globalAlpha = 0.5
      ctx.stroke()
      ctx.globalAlpha = 1
    }

    // Note: doc icon
    if (n.type === 'note') {
      const ix = n.x - 5
      const iy = n.y - 6
      ctx.strokeStyle = typeColors.stroke
      ctx.lineWidth = 1
      ctx.beginPath()
      ctx.rect(ix, iy, 10, 12)
      ctx.stroke()
      ctx.beginPath()
      ctx.moveTo(ix + 2, iy + 4)
      ctx.lineTo(ix + 8, iy + 4)
      ctx.moveTo(ix + 2, iy + 7)
      ctx.lineTo(ix + 8, iy + 7)
      ctx.stroke()
    }

    // Label
    const maxLen = n.type === 'user' ? 10 : 8
    const label = n.name.length > maxLen ? n.name.slice(0, maxLen) + '...' : n.name
    ctx.fillStyle = isHovered ? colors.textPrimary : typeColors.label
    ctx.font = `${n.type === 'user' ? '600' : isHovered ? '600' : '500'} ${n.type === 'user' ? 12 : 11}px "SF Pro Display", "Helvetica Neue", sans-serif`
    ctx.textAlign = 'center'
    ctx.textBaseline = 'top'
    ctx.fillText(label, n.x, n.y + r + 6)
  })

  ctx.restore()
}

function openCreateMenu(type) {
  createType.value = type
  showCreateMenu.value = true
}

async function createNode() {
  if (!createType.value) return
  const name = createType.value === 'galaxy' ? '新星系' : '新笔记'
  const parentId = props.user.id
  try {
    if (createType.value === 'galaxy') {
      await CreateGalaxy({
        id: '', name,
        parentId,
        parentPath: 'root',
        others: {}
      })
    } else {
      await CreateNote({
        name,
        file: '',
        parentId,
        parentPath: 'root',
        others: {}
      })
    }
    showCreateMenu.value = false
    createType.value = ''
    await loadNodes()
  } catch (e) {
    console.error('Create node failed:', e)
  }
}

function handleResize() {
  if (!canvasRef.value || !canvas) return
  width = canvasRef.value.clientWidth
  height = canvasRef.value.clientHeight
  canvas
    .attr('width', width * devicePixelRatio)
    .attr('height', height * devicePixelRatio)
    .style('width', width + 'px')
    .style('height', height + 'px')
  ctx = canvas.node().getContext('2d')
  ctx.scale(devicePixelRatio, devicePixelRatio)
  simulation.force('x', d3.forceX(width / 2).strength(0.05))
  simulation.force('y', d3.forceY(height / 2).strength(0.05))
  simulation.alpha(0.1).restart()
}

// Re-render when theme changes
const themeObserver = new MutationObserver(() => {
  render()
})

onMounted(() => {
  nextTick(() => {
    loadAvatar()
    loadNodes()
    window.addEventListener('resize', handleResize)
    themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
  })
})

onUnmounted(() => {
  if (simulation) simulation.stop()
  window.removeEventListener('resize', handleResize)
  themeObserver.disconnect()
})
</script>

<template>
  <div class="galaxy-container">
    <header class="galaxy-header">
      <button class="back-btn" @click="$emit('back')">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M9 2L4 7L9 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <span>返回主页</span>
      </button>
      <div class="header-info">
        <span class="node-count" v-if="!loading">{{ nodes.length }} 个节点</span>
      </div>
      <div class="header-spacer"></div>
    </header>

    <div class="galaxy-canvas" ref="canvasRef">
      <div v-if="loading" class="loading-overlay">
        <div class="loading-spinner"></div>
        <span>正在加载星系...</span>
      </div>
      <div v-else-if="nodes.length === 0" class="empty-state">
        <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
          <circle cx="24" cy="24" r="20" stroke="rgba(255,255,255,0.15)" stroke-width="1.5"/>
          <path d="M24 12l6 3.5v7L24 26l-6-3.5v-7L24 12z" fill="rgba(99,102,241,0.1)" stroke="rgba(99,102,241,0.4)" stroke-width="1"/>
        </svg>
        <p>星系还是空的</p>
        <span>点击右下角 + 创建第一个节点</span>
      </div>
    </div>

    <footer class="create-area">
      <Transition name="fade">
        <div v-if="showCreateMenu" class="create-menu">
          <div class="create-option" @click="openCreateMenu('galaxy')">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.2"/>
              <circle cx="8" cy="8" r="2" fill="currentColor"/>
            </svg>
            <span>新建星系</span>
          </div>
          <div class="create-option" @click="openCreateMenu('note')">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <rect x="3" y="2" width="10" height="12" rx="1" stroke="currentColor" stroke-width="1.2"/>
              <line x1="5" y1="6" x2="11" y2="6" stroke="currentColor" stroke-width="1"/>
              <line x1="5" y1="9" x2="11" y2="9" stroke="currentColor" stroke-width="1"/>
            </svg>
            <span>新建笔记</span>
          </div>
          <div v-if="createType" class="create-confirm">
            <span class="confirm-hint">创建{{ createType === 'galaxy' ? '星系' : '笔记' }}？</span>
            <button class="confirm-btn" @click="createNode">确认</button>
          </div>
        </div>
      </Transition>
      <button v-if="!showCreateMenu" class="add-btn" @click="showCreateMenu = true">
        <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
          <line x1="10" y1="4" x2="10" y2="16" stroke="white" stroke-width="1.5" stroke-linecap="round"/>
          <line x1="4" y1="10" x2="16" y2="10" stroke="white" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
      </button>
    </footer>
  </div>
</template>

<style scoped>
.galaxy-container {
  width: 100%;
  height: 100%;
  background: var(--bg-app);
  position: relative;
  font-family: 'SF Pro Display', 'Helvetica Neue', 'Switzer', sans-serif;
}

.galaxy-canvas {
  width: 100%;
  height: 100%;
  position: relative;
}

.galaxy-canvas :deep(canvas) {
  display: block;
}

/* Header */
.galaxy-header {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  padding: 20px 32px;
  gap: 16px;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 14px;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  backdrop-filter: blur(12px);
  transition: background 0.2s, border-color 0.2s;
}
.back-btn:hover {
  border-color: var(--text-secondary);
}

.header-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-count {
  font-size: 11px;
  color: var(--text-secondary);
  padding: 4px 10px;
  background: var(--glass-bg);
  border-radius: 4px;
  border: 1px solid var(--glass-border);
}

.header-spacer { flex: 1; }

/* Loading & Empty states */
.loading-overlay, .empty-state {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--text-secondary);
  font-size: 13px;
}

.loading-spinner {
  width: 24px;
  height: 24px;
  border: 2px solid var(--glass-border);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty-state p {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin: 0;
}

.empty-state span {
  font-size: 12px;
}

/* Create Area */
.create-area {
  position: absolute;
  bottom: 32px;
  right: 32px;
  z-index: 10;
}

.add-btn {
  width: 48px;
  height: 48px;
  border-radius: 6px;
  background: var(--accent);
  border: none;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s, transform 0.15s, box-shadow 0.2s;
  box-shadow: 0 4px 16px rgba(0,0,0,0.3);
}
.add-btn:hover { opacity: 0.85; }
.add-btn:active { transform: scale(0.97); }

.create-menu {
  position: absolute;
  bottom: 64px;
  right: 0;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 8px;
  padding: 8px;
  min-width: 150px;
  backdrop-filter: blur(20px);
}

.create-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
  font-size: 13px;
  color: var(--text-primary);
}
.create-option:hover { background: var(--glass-border); }
.create-option svg { color: var(--accent); }

.create-confirm {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  margin-top: 4px;
  border-top: 1px solid var(--glass-border);
}
.confirm-hint { font-size: 12px; color: var(--text-secondary); flex: 1; }
.confirm-btn {
  padding: 5px 12px;
  border-radius: 4px;
  background: var(--accent);
  border: none;
  color: var(--bg-app);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}
.confirm-btn:hover { opacity: 0.85; }

/* Transitions */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
