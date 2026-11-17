<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import {
  Chart,
  LineController,
  LineElement,
  PointElement,
  LinearScale,
  CategoryScale,
  Legend,
  Tooltip,
} from 'chart.js'

Chart.register(LineController, LineElement, PointElement, LinearScale, CategoryScale, Legend, Tooltip)

const props = defineProps({
  readings: { type: Array, default: () => [] },
  maxPoints: { type: Number, default: 50 },
})

const canvas = ref(null)
let chart = null

function updateChart() {
  if (!chart) return
  const recent = props.readings.slice(-props.maxPoints)
  chart.data.labels = recent.map((r) => new Date(r.timestamp).toLocaleTimeString())
  chart.data.datasets[0].data = recent.map((r) => r.temperature)
  chart.data.datasets[1].data = recent.map((r) => r.humidity)
  chart.update()
}

onMounted(() => {
  chart = new Chart(canvas.value, {
    type: 'line',
    data: {
      labels: [],
      datasets: [
        { label: 'อุณหภูมิ (°C)', data: [], borderColor: '#ef4444', yAxisID: 'yTemp', tension: 0.3 },
        { label: 'ความชื้น (%)', data: [], borderColor: '#3b82f6', yAxisID: 'yHum', tension: 0.3 },
      ],
    },
    options: {
      responsive: true,
      animation: false,
      scales: {
        yTemp: { type: 'linear', position: 'left', title: { display: true, text: '°C' } },
        yHum: {
          type: 'linear',
          position: 'right',
          title: { display: true, text: '%' },
          grid: { drawOnChartArea: false },
        },
      },
    },
  })
  updateChart()
})

// เมื่อข้อมูลเปลี่ยน ให้อัปเดตกราฟ
watch(() => props.readings, updateChart)

onUnmounted(() => chart?.destroy())
</script>

<template>
  <canvas ref="canvas"></canvas>
</template>
