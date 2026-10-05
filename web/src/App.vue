<template>
  <div id="app" class="flex flex-col h-screen overflow-hidden bg-slate-950 text-slate-100">
    <!-- Top System Header -->
    <Header />

    <!-- Main Content Area with Sidebar -->
    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar Navigation -->
      <Sidebar />

      <!-- Viewport / Route Views -->
      <main class="flex-1 overflow-y-auto p-4 bg-slate-950">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script>
import Header from './components/Header.vue';
import Sidebar from './components/Sidebar.vue';

export default {
  name: 'App',
  components: {
    Header,
    Sidebar,
  },
};
</script>

<style>
/* Global styling adjustments */
html, body {
  height: 100%;
  margin: 0;
  padding: 0;
}
</style>
