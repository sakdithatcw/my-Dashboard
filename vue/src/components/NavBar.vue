<template>
  <header class="navbar">
    <span class="brand">🌡️ Sensor Dashboard</span>
    <nav>
      <RouterLink to="/">Dashboard</RouterLink>
      <RouterLink to="/devices">Devices</RouterLink>
    </nav>
  </header>
</template>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 24px;
  background: #111827;
  color: white;
}
.brand {
  font-weight: bold;
}
nav {
  display: flex;
  gap: 16px;
}
nav a {
  color: #9ca3af;
  text-decoration: none;
}
nav a.router-link-exact-active {
  color: white;
  font-weight: bold;
}
</style>
