<script setup>
// การ์ดแสดงค่าหนึ่งค่า ใช้ซ้ำได้กี่ใบก็ได้ โดยส่งค่าผ่าน props
defineProps({
  title: String,
  icon: String,
  unit: String,
  value: { type: Number, default: null },
})
</script>

<template>
  <div class="device-card">
    <div class="icon">{{ icon }}</div>
    <div class="title">{{ title }}</div>
    <div class="value">
      {{ value !== null ? value.toFixed(1) : '--' }}
      <span class="unit">{{ unit }}</span>
    </div>
  </div>
</template>

<style scoped>
.device-card {
  flex: 1;
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  text-align: center;
  background: white;
}
.icon {
  font-size: 1.8rem;
}
.title {
  color: #6b7280;
}
.value {
  font-size: 2rem;
  font-weight: bold;
}
.unit {
  font-size: 1rem;
  font-weight: normal;
  color: #6b7280;
}
</style>
