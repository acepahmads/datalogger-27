<template>
  <div id="app" class="flex flex-col h-screen overflow-hidden bg-transparent dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150 relative">
    <!-- Top System Header (Hidden in Full-Width SCADA Mode or on Login) -->
    <Header
      v-if="!isLoginRoute"
      v-show="!isTopPanelHidden"
      @toggle-sidebar="toggleSidebar"
      @toggle-autohide="onSidebarAutoHideToggle"
    />

    <!-- Global Alarm Banner (Directly beneath header) -->
    <GlobalAlarmBanner v-if="!isLoginRoute" v-show="!isTopPanelHidden" />

    <!-- Main Content Area with Sidebar -->
    <div class="flex flex-1 overflow-hidden relative" v-if="!isLoginRoute">
      <!-- 1. Normal Pinned Sidebar (rendered when auto-hide is FALSE) -->
      <Sidebar
        v-if="!sidebarAutoHide"
        :is-collapsed="isSidebarCollapsed"
        :is-auto-hide="false"
        @toggle-autohide="onSidebarAutoHideToggle"
      />

      <!-- 2. Auto-Hide Edge Hover Trigger Zone (rendered when auto-hide is TRUE) -->
      <div
        v-if="sidebarAutoHide"
        @mouseenter="openSidebarHover"
        @click="openSidebarHover"
        class="fixed left-0 top-0 bottom-0 w-4 z-40 cursor-pointer group flex items-center justify-start pointer-events-auto"
        :title="$t('navigation.sidebarHoverTitle') || 'Hover to view navigation sidebar (Auto-Hide active)'"
      >
        <!-- Subtle cyan indicator beacon on edge -->
        <div class="h-20 w-1 rounded-r bg-slate-400/60 dark:bg-slate-700/80 group-hover:bg-cyan-500 group-hover:w-1.5 group-hover:shadow-glow-cyan transition-all duration-200"></div>
      </div>

      <!-- 3. Auto-Hide Sliding Drawer (rendered over content with smooth transition) -->
      <transition name="sidebar-slide">
        <div
          v-if="sidebarAutoHide && isSidebarHovered"
          @mouseleave="scheduleCloseSidebarHover"
          @mouseenter="cancelCloseSidebarHover"
          class="fixed left-0 top-0 bottom-0 z-50 w-60 shadow-2xl flex flex-col bg-white dark:bg-[#0B0F19] border-r border-slate-200 dark:border-slate-800"
        >
          <Sidebar
            :is-collapsed="false"
            :is-auto-hide="true"
            @toggle-autohide="onSidebarAutoHideToggle"
            @navigate="closeSidebarHoverImmediate"
          />
        </div>
      </transition>

      <!-- 4. Subtle backdrop when auto-hide sidebar is popped open -->
      <transition name="fade">
        <div
          v-if="sidebarAutoHide && isSidebarHovered"
          @click="closeSidebarHoverImmediate"
          class="fixed inset-0 z-40 bg-black/20 dark:bg-black/40 backdrop-blur-xs transition-opacity"
        ></div>
      </transition>

      <!-- Viewport / Route Views -->
      <main
        class="flex-1 transition-colors duration-150"
        :class="(isOperatorRoute || isDashboardOperatorRoute) ? 'px-1 sm:px-1.5 pt-0.5 sm:pt-1 pb-0 overflow-y-auto bg-[#EDF4F8] dark:bg-[#081421]' : (isTelemetryRoute ? 'p-2 sm:p-2.5 sm:pb-2 overflow-y-auto lg:overflow-hidden flex flex-col min-h-0 bg-transparent dark:bg-slate-950' : 'p-4 overflow-y-auto bg-transparent dark:bg-slate-950')"
      >
        <router-view
          :is-sidebar-collapsed="isSidebarCollapsed || sidebarAutoHide"
          :is-top-panel-hidden="isTopPanelHidden"
          :is-sidebar-auto-hide="sidebarAutoHide"
          @toggle-sidebar="toggleSidebar"
          @toggle-fullwidth="toggleSidebar"
          @toggle-autohide="onSidebarAutoHideToggle"
        />
      </main>
    </div>

    <!-- Login Route Viewport (when isLoginRoute is true) -->
    <main v-else class="flex-1 overflow-y-auto p-0 bg-transparent">
      <router-view />
    </main>
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
      isSidebarCollapsed: true, // Default to fullscreen SCADA mode for Operator Dashboard
      isTopPanelHidden: true,
      isSidebarHovered: false,
      hoverCloseTimer: null,
    };
  },
  computed: {
    isLoginRoute() {
      return this.$route && (this.$route.path === '/login' || this.$route.name === 'Login');
    },
    isOperatorRoute() {
      return this.$route && (this.$route.path === '/monitoring/operator' || this.$route.path === '/operator');
    },
    isDashboardOperatorRoute() {
      return this.$route && this.$route.path === '/monitoring/dashboard-operator';
    },
    isTelemetryRoute() {
      return this.$route && (this.$route.path === '/monitoring/telemetry' || this.$route.name === 'LiveTelemetry');
    },
    sidebarAutoHide() {
      return !!((this.$store.state.uiPreferences && this.$store.state.uiPreferences.sidebarAutoHide) || this.$store.state.sidebarAutoHide);
    },
  },
  watch: {
    $route: {
      immediate: true,
      handler(to) {
        if (to && (to.path === '/monitoring/operator' || to.path === '/operator' || to.path === '/monitoring/dashboard-operator')) {
          this.isSidebarCollapsed = true;
          this.isTopPanelHidden = true;
        } else {
          this.isTopPanelHidden = false;
          this.isSidebarCollapsed = false;
        }
        this.closeSidebarHoverImmediate();
      },
    },
  },
  methods: {
    toggleSidebar() {
      this.isSidebarCollapsed = !this.isSidebarCollapsed;
      if (this.isOperatorRoute || this.isDashboardOperatorRoute) {
        this.isTopPanelHidden = this.isSidebarCollapsed;
      } else {
        this.isTopPanelHidden = false;
      }
    },
    onSidebarAutoHideToggle() {
      if (!this.sidebarAutoHide) {
        this.closeSidebarHoverImmediate();
      }
    },
    toggleSidebarAutoHide() {
      this.$store.dispatch('toggleSidebarAutoHide');
      if (!this.sidebarAutoHide) {
        this.closeSidebarHoverImmediate();
      }
    },
    openSidebarHover() {
      this.cancelCloseSidebarHover();
      this.isSidebarHovered = true;
    },
    scheduleCloseSidebarHover() {
      this.cancelCloseSidebarHover();
      this.hoverCloseTimer = setTimeout(() => {
        this.isSidebarHovered = false;
      }, 250);
    },
    cancelCloseSidebarHover() {
      if (this.hoverCloseTimer) {
        clearTimeout(this.hoverCloseTimer);
        this.hoverCloseTimer = null;
      }
    },
    closeSidebarHoverImmediate() {
      this.cancelCloseSidebarHover();
      this.isSidebarHovered = false;
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

.sidebar-slide-enter-active,
.sidebar-slide-leave-active {
  transition: transform 0.22s cubic-bezier(0.16, 1, 0.3, 1);
}
.sidebar-slide-enter,
.sidebar-slide-leave-to {
  transform: translateX(-100%);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}
.fade-enter,
.fade-leave-to {
  opacity: 0;
}
</style>
