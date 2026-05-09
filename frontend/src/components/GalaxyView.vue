<script setup>
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import * as d3 from 'd3'
import { GetRelationById, CreateGalaxy, CreateNote, DeleteNode, UpdateNodeInfo, GetUserInfo, GetNoteCount, GetGalaxyCount, GetTagCount, GetAvatar } from '../../wailsjs/go/main/App'

const props = defineProps(['user', 'config'])
const emit = defineEmits(['back', 'open-note'])

const canvasRef = ref(null)
const nodes = ref([])
const loading = ref(true)
const showCreateMenu = ref(false)
const contextMenu = ref({ show: false, x: 0, y: 0, node: null })

// Sub-galaxy navigation
let currentCenterId = null
let currentCenterPath = null
let currentCenterName = ''
let parentGalaxyId = null

// Profile popup
const showProfile = ref(false)
const profileData = ref({ user: null, avatar: '', noteCount: 0, galaxyCount: 0, tagCount: 0 })

// Breadcrumb
const breadcrumbs = ref([])

// Rename modal
const showRename = ref(false)
const renameTarget = ref({ id: '', name: '' })
const renameInput = ref('')

// Click timer for single/double click
let clickTimer = null

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

function getNodeRadius(type) {
  const scale = props.config?.nodeScale ?? 1
  return (NODE_RADIUS[type] || 18) * scale
}

// Read theme colors from CSS variables
function getThemeColors() {
  const style = getComputedStyle(document.documentElement)
  const accent = style.getPropertyValue('--accent').trim() || '#6366f1'
  const textPrimary = style.getPropertyValue('--text-primary').trim() || '#ffffff'
  const textSecondary = style.getPropertyValue('--text-secondary').trim() || '#94a3b8'
  const bgApp = style.getPropertyValue('--bg-app').trim() || '#050505'

  return {
    user: { fill: bgApp, stroke: accent, label: textPrimary },
    galaxy: { fill: accent + '18', stroke: accent, label: textPrimary },
    note: { fill: accent + '10', stroke: accent + 'aa', label: textSecondary },
    link: textPrimary + '14',
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

function buildGraph(data) {
  const isRoot = currentCenterId === props.user.id

  // GetRelationById returns descendants but NOT the center node itself.
  // The links also won't include the center node since its ID wasn't in the query set.
  // We need to: (1) add the center node manually, (2) fetch links involving the center.

  // Find center in the returned data (only happens for root user)
  const centerInData = data.nodes.find(n => n.id === currentCenterId)
  const centerGraphNode = centerInData
    ? {
        id: centerInData.id,
        name: centerInData.title || '未命名',
        type: centerInData.type,
        radius: getNodeRadius(centerInData.type),
        x: width / 2,
        y: height / 2,
        fx: width / 2,
        fy: height / 2
      }
    : {
        id: currentCenterId,
        name: currentCenterName || '未命名',
        type: isRoot ? 'user' : 'galaxy',
        radius: getNodeRadius(isRoot ? 'user' : 'galaxy'),
        x: width / 2,
        y: height / 2,
        fx: width / 2,
        fy: height / 2
      }

  // Build child nodes (exclude the center node to avoid duplicates)
  const childNodes = (data.nodes || [])
    .filter(n => n.id !== currentCenterId)
    .map((n, i, arr) => {
      const angle = (2 * Math.PI * i) / arr.length
      const dist = 180 + Math.random() * 60
      return {
        id: n.id,
        name: n.title || '未命名',
        type: n.type,
        radius: getNodeRadius(n.type),
        x: width / 2 + Math.cos(angle) * dist,
        y: height / 2 + Math.sin(angle) * dist
      }
    })

  const gNodes = [centerGraphNode, ...childNodes]
  const nodeIdSet = new Set(gNodes.map(n => n.id))

  // For non-root views, GetRelationById won't have links to the center node
  // because the center wasn't in its node set. We need to fetch those links
  // from the raw relations. But since we only have the D3Graph output,
  // we create center links based on the child node paths.
  // All returned child nodes are direct children of the center, so link them.
  const gLinks = []

  // Include links from the API where both endpoints exist in our graph
  for (const l of (data.links || [])) {
    if (nodeIdSet.has(l.source) && nodeIdSet.has(l.target)) {
      gLinks.push({ source: l.source, target: l.target, type: l.type })
    }
  }

  // For non-root: the API won't have links from center to children,
  // so we add them manually (all returned nodes are children of the center)
  if (!isRoot) {
    const existingLinkTargets = new Set(gLinks.map(l => l.target))
    for (const child of childNodes) {
      // Only add if no link already exists to this child
      if (!existingLinkTargets.has(child.id)) {
        gLinks.push({ source: currentCenterId, target: child.id, type: child.type })
      }
    }
  }

  return { nodes: gNodes, links: gLinks }
}

async function loadNodes(fullRebuild = false) {
  loading.value = true
  try {
    const container = canvasRef.value
    if (container) {
      width = container.clientWidth
      height = container.clientHeight
    }

    if (!currentCenterId) {
      currentCenterId = props.user.id
      currentCenterPath = 'root'
      parentGalaxyId = null
    }

    const isRoot = currentCenterId === props.user.id

    let data
    if (isRoot) {
      currentCenterName = props.user.username
      data = await GetRelationById(props.user.id)
    } else {
      try {
        const nodeInfo = await GetUserInfo(currentCenterId)
        currentCenterName = nodeInfo.name
      } catch (e) {
        console.error('Failed to get node info:', e)
      }
      data = await GetRelationById(currentCenterId)
    }

    nodes.value = data.nodes || []

    // If simulation already exists and not a full rebuild (entering sub-galaxy),
    // update graph incrementally to preserve node positions
    if (simulation && !fullRebuild) {
      updateGraph(data)
      simulation.alpha(0.3).restart()
    } else {
      const graph = buildGraph(data)
      graphNodes = graph.nodes
      graphLinks = graph.links
    }

    await buildBreadcrumbs()

    if (!simulation || fullRebuild) {
      initGraph()
    }
  } catch (e) {
    console.error('Failed to load graph:', e)
  } finally {
    loading.value = false
  }
}

// Incremental graph update: preserve existing node positions
function updateGraph(data) {
  const isRoot = currentCenterId === props.user.id
  const existingMap = new Map(graphNodes.map(n => [n.id, n]))

  // Build new node list
  const newNodes = []

  // Center node
  const existingCenter = existingMap.get(currentCenterId)
  if (existingCenter) {
    existingCenter.name = currentCenterName || existingCenter.name
    newNodes.push(existingCenter)
  } else {
    newNodes.push({
      id: currentCenterId,
      name: currentCenterName || '未命名',
      type: isRoot ? 'user' : 'galaxy',
      radius: getNodeRadius(isRoot ? 'user' : 'galaxy'),
      x: width / 2,
      y: height / 2,
      fx: width / 2,
      fy: height / 2
    })
  }

  // Child nodes: keep existing positions, add new ones in a circle
  const childData = (data.nodes || []).filter(n => n.id !== currentCenterId)
  const newChildIds = new Set()

  for (const n of childData) {
    const existing = existingMap.get(n.id)
    if (existing) {
      existing.name = n.title || existing.name
      newNodes.push(existing)
    } else {
      newChildIds.add(n.id)
      newNodes.push({
        id: n.id,
        name: n.title || '未命名',
        type: n.type,
        radius: getNodeRadius(n.type),
        x: 0,
        y: 0
      })
    }
  }

  // Position new children in a circle around center
  const centerNode = newNodes[0]
  const newChildren = newNodes.filter(n => n.id !== currentCenterId && newChildIds.has(n.id))
  newChildren.forEach((n, i) => {
    const angle = (2 * Math.PI * i) / Math.max(newChildren.length, 1)
    const dist = 180 + Math.random() * 60
    n.x = centerNode.x + Math.cos(angle) * dist
    n.y = centerNode.y + Math.sin(angle) * dist
  })

  // Update links
  const nodeIdSet = new Set(newNodes.map(n => n.id))
  const newLinks = []

  for (const l of (data.links || [])) {
    if (nodeIdSet.has(l.source) && nodeIdSet.has(l.target)) {
      newLinks.push({ source: l.source, target: l.target, type: l.type })
    }
  }

  if (!isRoot) {
    const existingLinkTargets = new Set(newLinks.map(l => l.target))
    for (const n of newNodes) {
      if (n.id !== currentCenterId && !existingLinkTargets.has(n.id)) {
        newLinks.push({ source: currentCenterId, target: n.id, type: n.type })
      }
    }
  }

  // Replace arrays in-place so D3 simulation references stay valid
  graphNodes.length = 0
  graphNodes.push(...newNodes)
  graphLinks.length = 0
  graphLinks.push(...newLinks)
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

  const canvasEl = canvas.node()

  // IMPORTANT: Register dblclick BEFORE D3 zoom to ensure our capture-phase handler
  // fires before D3's internal bubble-phase handler (which calls stopImmediatePropagation)
  canvasEl.addEventListener('dblclick', handleDblClick, { capture: true })

  // Drag
  let dragTarget = null
  let dragStartX = 0
  let dragStartY = 0

  canvas.on('mousedown.drag', (event) => {
    const [mx, my] = d3.pointer(event, canvas.node())
    const [sx, sy] = transform.invert([mx, my])
    const node = findNode(sx, sy)
    if (!node) return

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

  // Zoom
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

  // Context menu on canvas
  canvasEl.addEventListener('contextmenu', (e) => {
    e.preventDefault()
    e.stopPropagation()
    const [x, y] = transform.invert([e.offsetX, e.offsetY])
    const node = findNode(x, y)
    if (!node || node.type === 'user') return
    contextMenu.value = {
      show: true,
      x: e.clientX,
      y: e.clientY,
      node
    }
  })

  simulation = d3.forceSimulation(graphNodes)
    .force('link', d3.forceLink(graphLinks).id(d => d.id).distance(180).strength(0.5))
    .force('charge', d3.forceManyBody().strength(-350))
    .force('x', d3.forceX(width / 2).strength(0.05))
    .force('y', d3.forceY(height / 2).strength(0.05))
    .force('collision', d3.forceCollide().radius(d => d.radius + 8).strength(0.7))
    .on('tick', render)

  simulation.alpha(0.6).restart()

  // Release center fixed position after layout stabilizes
  setTimeout(() => {
    const center = graphNodes.find(n => n.id === currentCenterId)
    if (center) {
      center.fx = null
      center.fy = null
    }
  }, 2500)
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
  if (!node) return

  if (node.type === 'user') {
    showUserProfile()
    return
  }

  if (node.type === 'note') {
    emit('open-note', { id: node.id, name: node.name })
    return
  }

  if (node.type === 'galaxy') {
    // Single click on galaxy: no action (use double-click to enter)
    return
  }
}

function handleDblClick(event) {
  event.preventDefault()
  event.stopPropagation()
  if (clickTimer) { clearTimeout(clickTimer); clickTimer = null }
  const [x, y] = transform.invert([event.offsetX, event.offsetY])
  const node = findNode(x, y)
  if (node && node.type === 'galaxy') {
    enterGalaxy(node.id, node.name)
  }
}

async function showUserProfile() {
  showProfile.value = true
  try {
    const [userNode, noteCount, galaxyCount, tagCount] = await Promise.all([
      GetUserInfo(props.user.id),
      GetNoteCount(),
      GetGalaxyCount(),
      GetTagCount()
    ])
    let avatar = ''
    try { avatar = await GetAvatar(props.user.id) } catch (e) {}
    profileData.value = { user: userNode, avatar, noteCount, galaxyCount, tagCount }
  } catch (e) {
    console.error('Failed to load profile:', e)
  }
}

function closeProfile() {
  showProfile.value = false
}

function handleMouseMove(event) {
  const [x, y] = transform.invert([event.offsetX, event.offsetY])
  const node = findNode(x, y)
  if (node !== hoveredNode) {
    hoveredNode = node
    canvas.style('cursor', node ? 'pointer' : 'default')
    render()
  }
}

function render() {
  if (!ctx) return
  const colors = getThemeColors()
  const bodyFont = getComputedStyle(document.documentElement).getPropertyValue('--font-body').trim() || "'DM Sans', sans-serif"

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
    ctx.lineWidth = props.config?.linkWidth ?? 2
    if ((props.config?.linkStyle ?? 'dashed') === 'dashed') {
      ctx.setLineDash([6, 4])
    } else {
      ctx.setLineDash([])
    }
    ctx.stroke()
    ctx.setLineDash([])
  })

  // Draw nodes
  graphNodes.forEach(n => {
    const isHovered = hoveredNode === n
    const typeColors = colors[n.type] || colors.note
    const r = n.radius

    // User node
    if (n.type === 'user') {
      ctx.beginPath()
      ctx.arc(n.x, n.y, r, 0, Math.PI * 2)
      ctx.fillStyle = typeColors.fill
      ctx.fill()
      ctx.strokeStyle = isHovered ? colors.textPrimary : typeColors.stroke
      ctx.lineWidth = isHovered ? 2.5 : 1.5
      ctx.stroke()

      if (avatarImg) {
        ctx.save()
        ctx.beginPath()
        ctx.arc(n.x, n.y, r - 2, 0, Math.PI * 2)
        ctx.clip()
        ctx.drawImage(avatarImg, n.x - r + 2, n.y - r + 2, (r - 2) * 2, (r - 2) * 2)
        ctx.restore()
        ctx.beginPath()
        ctx.arc(n.x, n.y, r - 2, 0, Math.PI * 2)
        ctx.strokeStyle = isHovered ? colors.textPrimary : typeColors.stroke
        ctx.lineWidth = isHovered ? 2.5 : 1.5
        ctx.stroke()
      } else {
        // No avatar: show first character of username
        const initial = (props.user.username || '?').charAt(0).toUpperCase()
        ctx.fillStyle = typeColors.label
        ctx.font = `600 ${r * 0.7}px ${bodyFont}`
        ctx.textAlign = 'center'
        ctx.textBaseline = 'middle'
        ctx.fillText(initial, n.x, n.y)
      }
    }

    // Galaxy
    if (n.type === 'galaxy') {
      ctx.beginPath()
      ctx.arc(n.x, n.y, r, 0, Math.PI * 2)
      ctx.fillStyle = typeColors.fill
      ctx.fill()
      ctx.strokeStyle = isHovered ? colors.accent : typeColors.stroke
      ctx.lineWidth = 1.5
      ctx.stroke()

      // Inner ring
      ctx.beginPath()
      ctx.arc(n.x, n.y, r * 0.5, 0, Math.PI * 2)
      ctx.strokeStyle = typeColors.stroke + '55'
      ctx.lineWidth = 1
      ctx.stroke()

      // Center dot
      ctx.beginPath()
      ctx.arc(n.x, n.y, 2, 0, Math.PI * 2)
      ctx.fillStyle = colors.accent + '88'
      ctx.fill()
    }

    // Note
    if (n.type === 'note') {
      ctx.beginPath()
      ctx.arc(n.x, n.y, r, 0, Math.PI * 2)
      ctx.fillStyle = typeColors.fill
      ctx.fill()
      ctx.strokeStyle = isHovered ? colors.accent : typeColors.stroke
      ctx.lineWidth = 1.5
      ctx.stroke()

      // Simple document icon
      const s = r * 0.38
      ctx.strokeStyle = isHovered ? colors.accent : colors.accent + '77'
      ctx.lineWidth = 1.2
      ctx.beginPath()
      ctx.rect(n.x - s, n.y - s * 1.2, s * 2, s * 2.4)
      ctx.stroke()
      ctx.beginPath()
      ctx.moveTo(n.x - s * 0.55, n.y - s * 0.3)
      ctx.lineTo(n.x + s * 0.55, n.y - s * 0.3)
      ctx.moveTo(n.x - s * 0.55, n.y + s * 0.3)
      ctx.lineTo(n.x + s * 0.3, n.y + s * 0.3)
      ctx.stroke()
    }

    // Label
    const maxLen = n.type === 'user' ? 10 : 8
    const label = n.name.length > maxLen ? n.name.slice(0, maxLen) + '...' : n.name
    ctx.fillStyle = isHovered ? colors.textPrimary : typeColors.label
    ctx.font = `${n.type === 'user' ? '600' : isHovered ? '600' : '500'} ${n.type === 'user' ? 12 : 11}px ${bodyFont}`
    ctx.textAlign = 'center'
    ctx.textBaseline = 'top'
    ctx.fillText(label, n.x, n.y + r + 6)
  })

  ctx.restore()
}

async function createNode(type) {
  const name = type === 'galaxy' ? '新星系' : '新笔记'
  const parentId = currentCenterId
  const isRoot = currentCenterId === props.user.id
  const parentPath = isRoot ? 'root' : currentCenterPath
  try {
    if (type === 'galaxy') {
      await CreateGalaxy({
        id: '', name,
        parentId,
        parentPath,
        others: {}
      })
    } else {
      await CreateNote({
        name,
        file: '# ' + name + '\n\n在此输入正文...',
        parentId,
        parentPath,
        others: {}
      })
    }
    showCreateMenu.value = false
    await loadNodes(false) // incremental update, preserve positions
  } catch (e) {
    console.error('Create node failed:', e)
  }
}

function handleContextMenu(event) {
  event.preventDefault()
  const [x, y] = transform.invert([event.offsetX, event.offsetY])
  const node = findNode(x, y)
  if (!node || node.type === 'user') return
  contextMenu.value = {
    show: true,
    x: event.offsetX,
    y: event.offsetY,
    node
  }
}

function closeContextMenu() {
  contextMenu.value.show = false
}

function handleKeydown(e) {
  if (e.key === 'Escape') {
    closeContextMenu()
    showProfile.value = false
    showRename.value = false
  }
}

async function deleteContextNode() {
  if (!contextMenu.value.node) return
  try {
    await DeleteNode(contextMenu.value.node.id)
    closeContextMenu()
    await loadNodes(false)
  } catch (e) {
    console.error('Delete node failed:', e)
  }
}

function openRenameModal() {
  if (!contextMenu.value.node) return
  renameTarget.value = { id: contextMenu.value.node.id, name: contextMenu.value.node.name }
  renameInput.value = contextMenu.value.node.name
  closeContextMenu()
  showRename.value = true
}

async function confirmRename() {
  if (!renameInput.value.trim()) return
  try {
    await UpdateNodeInfo({ id: renameTarget.value.id, title: renameInput.value.trim() })
    showRename.value = false
    // Update local state
    const centerNode = graphNodes.find(n => n.id === renameTarget.value.id)
    if (centerNode) {
      centerNode.name = renameInput.value.trim()
      if (centerNode.id === currentCenterId) currentCenterName = renameInput.value.trim()
    }
    const labelNode = graphNodes.find(n => n.id === renameTarget.value.id)
    if (labelNode) labelNode.name = renameInput.value.trim()
    await buildBreadcrumbs()
    render()
    await loadNodes(false)
  } catch (e) {
    console.error('Rename failed:', e)
  }
}

// Navigation: enter a galaxy as center
async function enterGalaxy(galaxyId, galaxyName) {
  parentGalaxyId = currentCenterId
  currentCenterId = galaxyId
  currentCenterName = galaxyName

  // Get the node's actual path from backend for accurate breadcrumbs
  try {
    const nodeInfo = await GetUserInfo(galaxyId)
    currentCenterPath = nodeInfo.path || 'root'
  } catch (e) {
    // Fallback: compute from breadcrumbs
    const isRootParent = parentGalaxyId === props.user.id
    if (isRootParent) {
      currentCenterPath = `root/${props.user.id}`
    } else {
      currentCenterPath = `${currentCenterPath}/${parentGalaxyId}`
    }
  }

  // Determine parent from path
  const pathParts = currentCenterPath.split('/')
  if (pathParts.length >= 3) {
    parentGalaxyId = pathParts[pathParts.length - 1]
  } else {
    parentGalaxyId = props.user.id
  }

  await loadNodes(true)
}

// Navigation: go back to parent
async function goBack() {
  if (currentCenterId === props.user.id) {
    emit('back')
    return
  }
  showProfile.value = false

  // Use the current path to determine the parent
  const pathParts = currentCenterPath.split('/')
  if (pathParts.length <= 2) {
    // Current path is like "root/userID", parent is root
    currentCenterId = props.user.id
    currentCenterPath = 'root'
    parentGalaxyId = null
  } else {
    // Current path is like "root/userID/g1/g2", parent is g1
    // Remove last segment to get parent's path
    const parentPath = pathParts.slice(0, -1).join('/')
    const parentId = pathParts[pathParts.length - 1]

    // Set parent galaxy
    if (pathParts.length === 3) {
      // Parent is user level: root/userID/g1 → go to user
      currentCenterId = props.user.id
      currentCenterPath = 'root'
      parentGalaxyId = null
    } else {
      // Parent is another galaxy: root/userID/g1/g2 → go to g1
      currentCenterId = parentId
      currentCenterPath = parentPath
      // Set grandparent
      const grandparentParts = parentPath.split('/')
      parentGalaxyId = grandparentParts.length >= 3 ? grandparentParts[grandparentParts.length - 1] : props.user.id
    }
  }

  await loadNodes(true)
}

// Build breadcrumb trail from the actual path
async function buildBreadcrumbs() {
  const crumbs = [{ id: props.user.id, name: props.user.username, parentId: null }]

  if (currentCenterId !== props.user.id) {
    // path is like "root/userID/g1/g2/g3"
    const pathParts = currentCenterPath.split('/')
    // Segments after "root" and userID are galaxy IDs
    const galaxyIds = pathParts.slice(2)

    let prevId = props.user.id
    for (const gid of galaxyIds) {
      // Try to find the name from graph nodes, otherwise fetch from backend
      const existingNode = graphNodes.find(n => n.id === gid)
      let name = existingNode ? existingNode.name : null
      if (!name) {
        try {
          const nodeInfo = await GetUserInfo(gid)
          name = nodeInfo.name || gid
        } catch (e) {
          name = gid
        }
      }
      crumbs.push({ id: gid, name, parentId: prevId })
      prevId = gid
    }

    // Add current center if not already in crumbs
    if (!crumbs.find(c => c.id === currentCenterId)) {
      crumbs.push({ id: currentCenterId, name: currentCenterName, parentId: prevId })
    }
  }

  breadcrumbs.value = crumbs
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

  // Recenter the center node
  const center = graphNodes.find(n => n.id === currentCenterId)
  if (center) {
    center.x = width / 2
    center.y = height / 2
    center.fx = width / 2
    center.fy = height / 2
  }

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
    window.addEventListener('click', closeContextMenu)
    window.addEventListener('keydown', handleKeydown)
    themeObserver.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] })
  })
})

onUnmounted(() => {
  if (simulation) simulation.stop()
  if (clickTimer) clearTimeout(clickTimer)
  window.removeEventListener('resize', handleResize)
  window.removeEventListener('click', closeContextMenu)
  window.removeEventListener('keydown', handleKeydown)
  themeObserver.disconnect()
})

watch(() => props.config, () => {
  graphNodes.forEach(n => {
    n.radius = getNodeRadius(n.type)
  })
  render()
}, { deep: true })
</script>

<template>
  <div class="galaxy-container">
    <header class="galaxy-header">
      <button class="back-btn" @click="goBack">
        <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
          <path d="M9 2L4 7L9 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <span v-if="currentCenterId === props.user.id">返回主页</span>
      </button>
      <div class="header-info">
        <span class="node-count" v-if="!loading">{{ nodes.length + 1 }} 个节点</span>
      </div>
      <div class="header-spacer"></div>
    </header>

    <!-- Breadcrumb -->
    <nav class="breadcrumb-nav" v-if="breadcrumbs.length > 1">
      <template v-for="(crumb, idx) in breadcrumbs" :key="crumb.id">
        <span
          v-if="idx > 0"
          class="breadcrumb-sep"
        >/</span>
        <span
          class="breadcrumb-item"
          :class="{ active: crumb.id === currentCenterId }"
          @click="crumb.id !== currentCenterId && enterGalaxy(crumb.id, crumb.name)"
        >{{ crumb.name }}</span>
      </template>
    </nav>

    <div class="galaxy-canvas" ref="canvasRef"></div>

    <div v-if="loading" class="loading-overlay">
      <div class="loading-spinner"></div>
      <span>正在加载星系...</span>
    </div>

    <div v-else-if="nodes.length === 0 && currentCenterId === props.user.id" class="empty-state">
      <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
        <circle cx="24" cy="24" r="20" stroke="rgba(255,255,255,0.15)" stroke-width="1.5"/>
        <path d="M24 12l6 3.5v7L24 26l-6-3.5v-7L24 12z" fill="rgba(99,102,241,0.1)" stroke="rgba(99,102,241,0.4)" stroke-width="1"/>
      </svg>
      <p>{{ currentCenterId === props.user.id ? '星系还是空的' : '这个星系还是空的' }}</p>
      <span>点击右下角 + 创建第一个节点</span>
    </div>

    <!-- Profile Card -->
    <Transition name="fade">
      <div v-if="showProfile" class="profile-overlay" @click.self="closeProfile">
        <div class="profile-card glass-panel">
          <button class="profile-close" @click="closeProfile">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
              <path d="M10 4L4 10M4 4l6 6" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
            </svg>
          </button>
          <div class="profile-header">
            <div class="profile-avatar" :class="{ 'has-avatar': profileData.avatar }">
              <img v-if="profileData.avatar" :src="profileData.avatar" alt="avatar" />
              <span v-else>{{ (props.user.username || '?').charAt(0).toUpperCase() }}</span>
            </div>
            <div class="profile-name">{{ profileData.user?.name || props.user.username }}</div>
            <div class="profile-motto">{{ profileData.user?.others?.motto || '' }}</div>
          </div>
          <div class="profile-stats">
            <div class="stat-item">
              <div class="stat-value">{{ profileData.noteCount }}</div>
              <div class="stat-label">笔记</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ profileData.galaxyCount }}</div>
              <div class="stat-label">星系</div>
            </div>
            <div class="stat-item">
              <div class="stat-value">{{ profileData.tagCount }}</div>
              <div class="stat-label">标签</div>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Rename Modal -->
    <Transition name="fade">
      <div v-if="showRename" class="profile-overlay" @click.self="showRename = false">
        <div class="rename-modal glass-panel">
          <h3>重命名星系</h3>
          <input
            v-model="renameInput"
            class="rename-input"
            placeholder="输入新名称"
            @keydown.enter="confirmRename"
            autofocus
          />
          <div class="rename-actions">
            <button class="rename-cancel" @click="showRename = false">取消</button>
            <button class="rename-confirm" @click="confirmRename">确认</button>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Context Menu -->
    <Transition name="fade">
      <div
        v-if="contextMenu.show"
        class="context-menu"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
      >
        <div v-if="contextMenu.node?.type === 'galaxy'" class="context-item" @click.stop="openRenameModal">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M8.5 2.5l3 3M2 9l5.5-5.5 3 3L5 12H2V9z" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <span>重命名</span>
        </div>
        <div class="context-item danger" @click.stop="deleteContextNode">
          <svg width="14" height="14" viewBox="0 0 14 14" fill="none">
            <path d="M2 3.5h10M5 3.5V2.5a.5.5 0 01.5-.5h3a.5.5 0 01.5.5v1M11 3.5v8a.5.5 0 01-.5.5H3.5a.5.5 0 01-.5-.5v-8" stroke="currentColor" stroke-width="1.2" stroke-linecap="round"/>
          </svg>
          <span>删除节点</span>
        </div>
      </div>
    </Transition>

    <footer class="create-area">
      <Transition name="fade">
        <div v-if="showCreateMenu" class="create-menu" @click.stop>
          <div class="create-option cursor-pointer" @click="createNode('galaxy')">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <circle cx="8" cy="8" r="6" stroke="currentColor" stroke-width="1.2"/>
              <circle cx="8" cy="8" r="2" fill="currentColor"/>
            </svg>
            <span>创建星系</span>
          </div>
          <div class="create-option cursor-pointer" @click="createNode('note')">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <rect x="3" y="2" width="10" height="12" rx="1" stroke="currentColor" stroke-width="1.2"/>
              <line x1="5" y1="6" x2="11" y2="6" stroke="currentColor" stroke-width="1"/>
              <line x1="5" y1="9" x2="11" y2="9" stroke="currentColor" stroke-width="1"/>
            </svg>
            <span>创建笔记</span>
          </div>
        </div>
      </Transition>

      <button class="add-btn" @click="showCreateMenu = !showCreateMenu; closeContextMenu()">
        <svg v-if="!showCreateMenu" width="20" height="20" viewBox="0 0 20 20" fill="none">
          <line x1="10" y1="4" x2="10" y2="16" stroke="white" stroke-width="1.5" stroke-linecap="round"/>
          <line x1="4" y1="10" x2="16" y2="10" stroke="white" stroke-width="1.5" stroke-linecap="round"/>
        </svg>
        <svg v-else width="18" height="18" viewBox="0 0 18 18" fill="none">
          <line x1="14" y1="4" x2="4" y2="14" stroke="white" stroke-width="1.5" stroke-linecap="round"/>
          <line x1="4" y1="4" x2="14" y2="14" stroke="white" stroke-width="1.5" stroke-linecap="round"/>
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
  font-family: var(--font-body, 'DM Sans', sans-serif);
}

.galaxy-canvas {
  width: 100%;
  height: 100%;
  position: relative;
}

.galaxy-canvas :deep(canvas) {
  display: block;
}

.cursor-pointer { cursor: pointer; }

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

/* Breadcrumb */
.breadcrumb-nav {
  position: absolute;
  top: 64px;
  left: 32px;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 6px;
  padding: 6px 12px;
  backdrop-filter: blur(12px);
}

.breadcrumb-item {
  cursor: pointer;
  transition: color 0.2s;
  padding: 2px 4px;
  border-radius: 3px;
}

.breadcrumb-item:hover {
  color: var(--text-primary);
}

.breadcrumb-item.active {
  color: var(--accent);
  font-weight: 500;
  cursor: default;
}

.breadcrumb-sep {
  color: var(--glass-border);
  user-select: none;
}

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

/* Profile Card */
.profile-overlay {
  position: absolute;
  inset: 0;
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
}

.profile-card {
  width: 280px;
  padding: 28px 24px;
  border-radius: 12px;
  position: relative;
  text-align: center;
}

.profile-close {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background 0.15s, color 0.15s;
}

.profile-close:hover {
  background: rgba(255,255,255,0.08);
  color: var(--text-primary);
}

.profile-header {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
}

.profile-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  border: 2px solid var(--accent);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  background: var(--bg-app);
}

.profile-avatar.has-avatar {
  border-color: var(--glass-border);
}

.profile-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.profile-avatar span {
  font-size: 22px;
  font-weight: 600;
  color: var(--accent);
}

.profile-name {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.profile-motto {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.profile-stats {
  display: flex;
  justify-content: center;
  gap: 24px;
  padding-top: 16px;
  border-top: 1px solid var(--glass-border);
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: var(--text-primary);
}

.stat-label {
  font-size: 11px;
  color: var(--text-secondary);
  margin-top: 2px;
}

/* Rename Modal */
.rename-modal {
  width: 300px;
  padding: 24px;
  border-radius: 12px;
}

.rename-modal h3 {
  margin: 0 0 16px 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.rename-input {
  width: 100%;
  padding: 10px 12px;
  background: var(--bg-app);
  border: 1px solid var(--glass-border);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 13px;
  font-family: var(--font-body, 'DM Sans', sans-serif);
  outline: none;
  transition: border-color 0.2s;
}

.rename-input:focus {
  border-color: var(--accent);
}

.rename-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.rename-cancel, .rename-confirm {
  padding: 7px 16px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid var(--glass-border);
  transition: background 0.15s, border-color 0.15s;
}

.rename-cancel {
  background: transparent;
  color: var(--text-secondary);
}

.rename-cancel:hover {
  background: rgba(255,255,255,0.05);
  color: var(--text-primary);
}

.rename-confirm {
  background: var(--accent);
  border-color: var(--accent);
  color: white;
}

.rename-confirm:hover {
  opacity: 0.9;
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

/* Context Menu */
.context-menu {
  position: fixed;
  background: var(--glass-bg);
  border: 1px solid var(--glass-border);
  border-radius: 6px;
  padding: 4px;
  backdrop-filter: blur(20px);
  transform: translate(-50%, -100%);
  margin-top: -8px;
}

.context-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 4px;
  cursor: pointer;
  transition: background 0.15s;
  font-size: 12px;
  color: var(--text-primary);
  white-space: nowrap;
}
.context-item:hover { background: rgba(255, 255, 255, 0.08); }
.context-item.danger:hover {
  background: rgba(239, 68, 68, 0.15);
  color: #ef4444;
}
.context-item.danger svg { color: #ef4444; }

/* Transitions */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.25s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
