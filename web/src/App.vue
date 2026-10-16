<template>
  <div id="app" class="flex flex-col h-screen overflow-hidden bg-transparent dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150">
    <!-- Top System Header (Hidden in Full-Width SCADA Mode) -->
    <Header v-show="!isTopPanelHidden" />

    <!-- Global Alarm Banner (Directly beneath header) -->
    <GlobalAlarmBanner v-show="!isTopPanelHidden" />

    <!-- Main Content Area with Sidebar -->
    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar Navigation -->
      <Sidebar :is-collapsed="isSidebarCollapsed" />

      <!-- Viewport / Route Views -->
      <main
        class="flex-1 overflow-y-auto transition-colors duration-150"
        :class="isOperatorRoute ? 'px-1 sm:px-1.5 pt-0.5 sm:pt-1 pb-0 bg-transparent dark:bg-[#050B16]' : 'p-4 bg-transparent dark:bg-slate-950'"
      >
        <router-view
          :is-sidebar-collapsed="isSidebarCollapsed"
          :is-top-panel-hidden="isTopPanelHidden"
          @toggle-sidebar="toggleSidebar"
          @toggle-fullwidth="toggleSidebar"
        />
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
      isTopPanelHidden: false,
    };
  },
  computed: {
    isOperatorRoute() {
      return this.$route && (this.$route.path === '/monitoring/operator' || this.$route.path === '/operator');
    },
  },
  watch: {
    $route(to) {
      if (to.path !== '/monitoring/operator' && to.path !== '/operator') {
        this.isTopPanelHidden = false;
      }
    },
  },
  methods: {
    toggleSidebar() {
      this.isSidebarCollapsed = !this.isSidebarCollapsed;
      if (this.isOperatorRoute) {
        this.isTopPanelHidden = this.isSidebarCollapsed;
      } else {
        this.isTopPanelHidden = false;
      }
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
