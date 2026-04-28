<script setup>
import { onMounted, ref } from 'vue'
import * as d3 from 'd3'

const props = defineProps(['user'])
const emit = defineEmits(['back'])

const svgRef = ref(null)

onMounted(() => {
  initGalaxy()
})

function initGalaxy() {
  const width = svgRef.value.clientWidth
  const height = svgRef.value.clientHeight

  d3.select(svgRef.value).selectAll("*").remove()

  const svg = d3.select(svgRef.value)
      .append("g")

  const nodes = [
    {
      id: "user-center",
      name: props.user.username,
      isRoot: true,
      x: width / 2, // 初始位置设为中心
      y: height / 2
    }
  ]

  const simulation = d3.forceSimulation(nodes)
      .force("charge", d3.forceManyBody().strength(-50))
      .force("center", d3.forceCenter(width / 2, height / 2))
      .on("tick", ticked)

  const node = svg.append("g")
      .selectAll("g")
      .data(nodes)
      .join("g")
      .call(d3.drag()
          .on("start", (event) => {
            if (!event.active) simulation.alphaTarget(0.3).restart()
            event.subject.fx = event.subject.x
            event.subject.fy = event.subject.y
          })
          .on("drag", (event) => {
            event.subject.fx = event.x
            event.subject.fy = event.y
          })
          .on("end", (event) => {
            if (!event.active) simulation.alphaTarget(0)
            // 拖拽结束后固定位置，不自动回弹
            event.subject.fx = event.x
            event.subject.fy = event.y
          }))
      .attr("class", "galaxy-node")

  node.append("circle")
      .attr("r", 45)
      .attr("class", "node-glow")

  node.append("circle")
      .attr("r", 40)
      .attr("class", "node-main")

  node.append("text")
      .text(d => d.name ? d.name[0].toUpperCase() : "V")
      .attr("text-anchor", "middle")
      .attr("dy", ".35em")
      .attr("class", "node-text")

  node.append("text")
      .attr("dy", 70)
      .attr("text-anchor", "middle")
      .attr("class", "node-label")
      .text(d => d.name)

  function ticked() {
    node.attr("transform", d => `translate(${d.x},${d.y})`)
  }
}
</script>

<template>
  <div class="galaxy-container">
    <button class="back-btn" @click="$emit('back')">← 返回主页 / EXIT GALAXY</button>
    <svg ref="svgRef" class="galaxy-svg"></svg>
  </div>
</template>

<style scoped>
.galaxy-container { width: 100%; height: 100%; background: #050505; position: relative; }
.galaxy-svg { width: 100%; height: 100%; }
.back-btn {
  position: absolute; top: 30px; left: 30px; z-index: 10;
  background: rgba(255,255,255,0.05); border: 1px solid rgba(255,255,255,0.1);
  color: #fff; padding: 10px 20px; border-radius: 8px; cursor: pointer;
  backdrop-filter: blur(10px);
}
:deep(.node-main) { fill: #0f172a; stroke: #3b82f6; stroke-width: 2px; }
:deep(.node-glow) { fill: #3b82f6; opacity: 0.15; filter: blur(8px); }
:deep(.node-text) { fill: #fff; font-size: 24px; font-weight: 800; pointer-events: none; }
:deep(.node-label) { fill: #94a3b8; font-size: 12px; pointer-events: none; }
:deep(.galaxy-node) { cursor: grab; }
:deep(.galaxy-node:active) { cursor: grabbing; }
</style>