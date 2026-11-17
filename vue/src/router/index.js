import { createRouter, createWebHistory } from 'vue-router'
import DashboardView from '../views/DashboardView.vue'

export default createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'dashboard', component: DashboardView },
    {
      path: '/devices',
      name: 'devices',
      component: () => import('../views/DevicesView.vue'), // โหลดเมื่อเข้าหน้านี้
    },
  ],
})
