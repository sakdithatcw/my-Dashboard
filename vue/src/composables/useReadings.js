import { ref, unref, watch, onMounted, onUnmounted } from 'vue'

// deviceId: ref (หรือค่าธรรมดา) ของอุปกรณ์ที่ต้องการดู
export function useReadings(deviceId, intervalMs = 5000) {
  const readings = ref([])
  const latest = ref(null)
  const error = ref('')

  let timer = null

  async function load() {
    const id = unref(deviceId)
    if (!id) {
      readings.value = []
      latest.value = null
      return
    }
    try {
      const res = await fetch(`/api/readings?device_id=${encodeURIComponent(id)}&limit=200`)
      if (!res.ok) throw new Error('HTTP ' + res.status)
      const data = await res.json()
      readings.value = data
      latest.value = data.length ? data[data.length - 1] : null
      error.value = ''
    } catch (e) {
      error.value = 'เชื่อมต่อ backend ไม่ได้: ' + e.message
    }
  }

  // เปลี่ยนอุปกรณ์แล้วโหลดใหม่ทันที
  watch(() => unref(deviceId), load)

  onMounted(() => {
    load()
    timer = setInterval(load, intervalMs)
  })

  onUnmounted(() => clearInterval(timer))

  return { readings, latest, error, reload: load }
}
