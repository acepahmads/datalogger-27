<template>
  <div id="app" class="flex flex-col h-screen overflow-hidden bg-slate-100 dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150">
    <!-- Top System Header -->
    <Header />

    <!-- Global Alarm Banner (Directly beneath header) -->
    <GlobalAlarmBanner />

    <!-- Main Content Area with Sidebar -->
    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar Navigation -->
      <Sidebar :is-collapsed="isSidebarCollapsed" />

      <!-- Viewport / Route Views -->
      <main
        class="flex-1 overflow-y-auto transition-colors duration-150"
        :class="isOperatorRoute ? 'p-1.5 sm:p-2.5 bg-[#050B16]' : 'p-4 bg-slate-50 dark:bg-slate-950'"
      >
        <router-view :is-sidebar-collapsed="isSidebarCollapsed" @toggle-sidebar="toggleSidebar" />
      </main>
    </div>
  </div>
</template>

<script>
import Header from './components/Header.vue';
import Sidebar from './components/Sidebar.vue';
import GlobalAlarmBanner from './components/GlobalAlarmBanner.vue';

export default {
  name: 'App',
  components: {
    Header,
    Sidebar,
    GlobalAlarmBanner,
  },
  data() {
    return {
      isSidebarCollapsed: false,
    };
  },
  computed: {
    isOperatorRoute() {
      return this.$route && (this.$route.path === '/monitoring/operator' || this.$route.path === '/operator');
    },
  },
  methods: {
    toggleSidebar() {
      this.isSidebarCollapsed = !this.isSidebarCollapsed;
    },
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
