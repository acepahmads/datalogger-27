<template>
  <div id="app" class="flex flex-col h-screen overflow-hidden bg-transparent dark:bg-slate-950 text-slate-800 dark:text-slate-100 transition-colors duration-150 relative">
    <!-- Top System Header (Hidden in Full-Width SCADA Mode or on Login) -->
    <Header
      v-if="!isLoginRoute && isShellReady"
      v-show="!isTopPanelHidden"
      @toggle-sidebar="toggleSidebar"
      @toggle-autohide="onSidebarAutoHideToggle"
    />

    <!-- Global Alarm Banner (Directly beneath header) -->
    <GlobalAlarmBanner v-if="!isLoginRoute && isShellReady" v-show="!isTopPanelHidden" />

    <!-- Main Content Area with Sidebar -->
    <div class="flex flex-1 overflow-hidden relative" v-if="!isLoginRoute && isShellReady">
      <!-- 1. Normal Pinned Sidebar (rendered when auto-hide is FALSE — Desktop >= 768px only) -->
      <div v-if="!sidebarAutoHide" class="hidden md:flex h-full flex-shrink-0">
        <Sidebar
          :is-collapsed="isSidebarCollapsed"
          :is-auto-hide="false"
          @toggle-autohide="onSidebarAutoHideToggle"
          @request-logout="openLogoutConfirmModal"
        />
      </div>

      <!-- 2. Auto-Hide Edge Hover Trigger Zone (rendered when auto-hide is TRUE — Desktop >= 768px only) -->
      <div
        v-if="sidebarAutoHide"
        @mouseenter="openSidebarHover"
        @click="openSidebarHover"
        class="hidden md:flex fixed left-0 top-0 bottom-0 w-4 z-40 cursor-pointer group items-center justify-start pointer-events-auto"
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
          class="hidden md:flex fixed left-0 top-0 bottom-0 z-50 w-60 shadow-2xl flex-col bg-white dark:bg-[#0B0F19] border-r border-slate-200 dark:border-slate-800"
        >
          <Sidebar
            :is-collapsed="false"
            :is-auto-hide="true"
            @toggle-autohide="onSidebarAutoHideToggle"
            @navigate="closeSidebarHoverImmediate"
            @request-logout="openLogoutConfirmModal"
          />
        </div>
      </transition>

      <!-- 4. Subtle backdrop when auto-hide sidebar is popped open -->
      <transition name="fade">
        <div
          v-if="sidebarAutoHide && isSidebarHovered"
          @click="closeSidebarHoverImmediate"
          class="hidden md:block fixed inset-0 z-40 bg-black/20 dark:bg-black/40 backdrop-blur-xs transition-opacity"
        ></div>
      </transition>

      <!-- 5. Mobile Full-Height Slide-Over Navigation Drawer & Backdrop (< 768px) -->
      <transition name="fade">
        <div
          v-if="isMobileViewport && isMobileDrawerOpen"
          @click="closeMobileDrawer"
          class="fixed inset-0 z-[65] bg-slate-950/55 backdrop-blur-xs transition-opacity md:hidden"
        ></div>
      </transition>

      <transition name="sidebar-slide">
        <div
          v-if="isMobileViewport && isMobileDrawerOpen"
          id="mobile-sidebar-drawer"
          class="fixed inset-y-0 left-0 z-[70] w-[78vw] max-w-[280px] shadow-2xl flex flex-col bg-white dark:bg-[#0B0F19] border-r border-slate-200 dark:border-slate-800 md:hidden"
        >
          <Sidebar
            :is-collapsed="false"
            :is-auto-hide="false"
            :is-mobile-drawer="true"
            @navigate="closeMobileDrawer"
            @close-mobile="closeMobileDrawer"
            @request-logout="openLogoutConfirmModal"
          />
        </div>
      </transition>

      <!-- Viewport / Route Views -->
      <main
        class="flex-1 min-w-0 transition-colors duration-150"
        :class="(isOperatorRoute || isDashboardOperatorRoute) ? 'px-1 sm:px-1.5 pt-0.5 sm:pt-1 pb-0 overflow-y-auto bg-[#EDF4F8] dark:bg-[#081421]' : (isTelemetryRoute ? 'p-2 sm:p-2.5 sm:pb-2 overflow-y-auto lg:overflow-hidden flex flex-col min-h-0 bg-transparent dark:bg-slate-950' : 'p-2.5 sm:p-4 overflow-y-auto bg-transparent dark:bg-slate-950')"
      >
        <router-view
          :key="'ctx-' + ($store.state.contextRevision || 0)"
          :is-sidebar-collapsed="isMobileViewport ? !isMobileDrawerOpen : (isSidebarCollapsed || sidebarAutoHide)"
          :is-top-panel-hidden="isTopPanelHidden"
          :is-sidebar-auto-hide="sidebarAutoHide"
          @toggle-sidebar="toggleSidebar"
          @toggle-fullwidth="toggleSidebar"
          @toggle-autohide="onSidebarAutoHideToggle"
        />
      </main>
    </div>

    <!-- Login Route Viewport (when isLoginRoute is true) -->
    <main v-else-if="isLoginRoute" class="flex-1 overflow-y-auto p-0 bg-transparent">
      <router-view />
    </main>

    <!-- Initial Auth Check Placeholder (prevents dashboard flash before Cloud /login redirect) -->
    <div v-else class="flex-1 bg-slate-50 dark:bg-slate-950"></div>

    <!-- Global Logout Confirmation Modal (Accessible from Sidebar even when Top Header is hidden) -->
    <div
      v-if="showLogoutConfirmModal"
      @click.self="showLogoutConfirmModal = false"
      class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs font-sans"
    >
      <div class="bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 rounded-3xl w-full max-w-sm p-6 shadow-2xl space-y-4 text-center m-auto max-h-[90vh] overflow-y-auto">
        <div class="w-12 h-12 rounded-2xl bg-rose-500/10 text-rose-500 flex items-center justify-center mx-auto">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </div>

        <div>
          <h2 class="text-base font-bold text-slate-900 dark:text-white">{{ $t('auth.logoutConfirmTitle') || 'Confirm Session Logout' }}</h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
            {{ $t('auth.logoutConfirmMsg') || 'Are you sure you want to end your active session? You will be redirected to the sign-in screen.' }}
          </p>
        </div>

        <div class="pt-3 flex items-center justify-center gap-2 border-t border-slate-100 dark:border-slate-800">
          <button
            type="button"
            @click="showLogoutConfirmModal = false"
            class="px-4 py-2 rounded-xl border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-semibold cursor-pointer"
          >
            {{ $t('auth.cancel') || 'Cancel' }}
          </button>
          <button
            type="button"
            @click="confirmAppLogout"
            class="px-4 py-2 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold shadow-xs cursor-pointer"
          >
            {{ $t('auth.confirmLogout') || 'Yes, Logout' }}
          </button>
        </div>
      </div>
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
      isSidebarCollapsed: true, // Default to fullscreen SCADA mode for Operator Dashboard
      isTopPanelHidden: true,
      isSidebarHovered: false,
      isMobileDrawerOpen: false,
      viewportWidth: typeof window !== 'undefined' ? window.innerWidth : 1280,
      hoverCloseTimer: null,
      showLogoutConfirmModal: false,
    };
  },
  computed: {
    isMobileViewport() {
      return this.viewportWidth < 768;
    },
    isLoginRoute() {
      return this.$route && (this.$route.path === '/login' || this.$route.name === 'Login');
    },
    isShellReady() {
      return Boolean(this.$store && this.$store.getters && this.$store.getters.isAuthenticated);
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
        this.isMobileDrawerOpen = false;
        this.closeSidebarHoverImmediate();
      },
    },
  },
  mounted() {
    if (typeof window !== 'undefined') {
      window.addEventListener('resize', this.onWindowResize, { passive: true });
      window.addEventListener('keydown', this.onGlobalKeydown);
    }
  },
  beforeDestroy() {
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', this.onWindowResize);
      window.removeEventListener('keydown', this.onGlobalKeydown);
    }
  },
  methods: {
    onWindowResize() {
      if (typeof window === 'undefined') return;
      this.viewportWidth = window.innerWidth;
      if (this.viewportWidth >= 768 && this.isMobileDrawerOpen) {
        this.isMobileDrawerOpen = false;
      }
    },
    onGlobalKeydown(e) {
      if (e && e.key === 'Escape' && this.isMobileDrawerOpen) {
        this.isMobileDrawerOpen = false;
      }
    },
    closeMobileDrawer() {
      this.isMobileDrawerOpen = false;
    },
    toggleSidebar() {
      if (this.isMobileViewport) {
        this.isMobileDrawerOpen = !this.isMobileDrawerOpen;
        return;
      }
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
    openLogoutConfirmModal() {
      this.closeMobileDrawer();
      this.closeSidebarHoverImmediate();
      this.showLogoutConfirmModal = true;
    },
    async confirmAppLogout() {
      this.showLogoutConfirmModal = false;
      this.closeMobileDrawer();
      this.closeSidebarHoverImmediate();
      await this.$store.dispatch('logout');
      if (this.$route.path !== '/login') {
        this.$router.push('/login');
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
