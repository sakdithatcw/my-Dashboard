import { ref, onMounted, onUnmounted } from 'vue'

// ดึงรายการอุปกรณ์ทั้งหมด + ค่าล่าสุดของแต่ละตัว
export function useDevices(intervalMs = 5000) {
  const devices = ref([])
  const error = ref('')

  let timer = null

  async function load() {
    try {
      const res = await fetch('/api/devices')
      if (!res.ok) throw new Error('HTTP ' + res.status)
      devices.value = await res.json()
      error.value = ''
    } catch (e) {
      error.value = 'เชื่อมต่อ backend ไม่ได้: ' + e.message
    }
  }

  onMounted(() => {
    load()
    timer = setInterval(load, intervalMs)
  })

  onUnmounted(() => clearInterval(timer))

  return { devices, error, reload: load }
}
