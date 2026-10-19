<template>
  <header class="bg-white/95 dark:bg-[#0B0F19]/95 backdrop-blur-md border-b border-slate-200 dark:border-slate-800/60 text-slate-800 dark:text-slate-200 px-4 sm:px-5 py-2.5 flex items-center justify-between sticky top-0 z-40 select-none transition-colors duration-150">
    <!-- Left: Logo & App Title -->
    <div class="flex items-center space-x-3">
      <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-600 to-indigo-700 flex items-center justify-center shadow-soft text-white font-bold text-xs tracking-wider">
        DL
      </div>
      <div class="flex items-center space-x-2">
        <h1 class="text-xs font-bold tracking-tight text-slate-900 dark:text-white uppercase font-sans">
          {{ $t('header.appTitle') }}
        </h1>
        <span class="bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20 text-3xs px-2 py-0.5 rounded-full font-mono font-medium">
          {{ $t('header.version') }}
        </span>
      </div>
    </div>

    <!-- Right: Controls & User -->
    <div class="flex items-center space-x-2 sm:space-x-2.5">
      <!-- 0. Sidebar Auto-Hide Setup Icon Button -->
      <button
        type="button"
        id="btn-sidebar-autohide-toggle"
        @click="toggleSidebarAutoHide"
        class="relative w-8 h-8 rounded-lg border transition-colors flex items-center justify-center shadow-xs"
        :class="sidebarAutoHide
          ? 'bg-cyan-500/20 text-cyan-400 border-cyan-500/50 ring-1 ring-cyan-500/30'
          : 'bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border-slate-200 dark:border-slate-700/60'"
        :title="sidebarAutoHide ? ($t('header.sidebarAutoHideOn') || 'Sidebar: Auto-Hide Enabled (Click to Pin)') : ($t('header.sidebarAutoHideOff') || 'Sidebar: Pinned (Click for Auto-Hide)')"
        aria-label="Toggle Sidebar Auto-Hide"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <path d="M9 3v18" />
          <path v-if="!sidebarAutoHide" d="m14 9-3 3 3 3" stroke-linecap="round" stroke-linejoin="round" />
          <path v-else d="m12 9 3 3-3 3" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>

      <!-- 1. Operator Console Quick Icon Button -->
      <router-link
        to="/monitoring/operator"
        class="relative w-8 h-8 rounded-lg border transition-colors flex items-center justify-center shadow-xs"
        :class="$route.path.includes('/operator')
          ? 'bg-cyan-500/20 text-cyan-400 border-cyan-500/50 ring-1 ring-cyan-500/30'
          : 'bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-cyan-600 dark:text-cyan-400 hover:text-cyan-500 dark:hover:text-cyan-300 border-slate-200 dark:border-slate-700/60'"
        :title="$t('navigation.operatorConsole') || 'Operator Console'"
        aria-label="Operator Console"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5m.75-9l3-3 2.25 2.25L15 6" />
        </svg>
        <!-- Pulsing live indicator dot -->
        <span class="absolute top-1 right-1 flex h-1.5 w-1.5">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-cyan-500"></span>
        </span>
      </router-link>

      <!-- 2. Notification Bell with Unread / Active Alarms Badge -->
      <router-link
        to="/alarm/active"
        class="relative w-8 h-8 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 transition-colors flex items-center justify-center shadow-xs"
        :title="$t('navigation.alarmsEvents') || 'Active Alarms'"
        aria-label="Active Alarms"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        <span
          v-if="activeAlarmsCount > 0"
          class="absolute -top-1 -right-1 px-1.5 py-0.2 rounded-full text-4xs font-mono font-bold bg-rose-500 text-white shadow-xs animate-pulse"
        >
          {{ activeAlarmsCount }}
        </span>
      </router-link>

      <!-- 3. Language Switch Icon Button (EN / ID Slide Switch) -->
      <button
        type="button"
        @click="toggleLanguage"
        class="h-8 px-2 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 flex items-center space-x-1.5 transition-all shadow-xs"
        :title="$t('header.language') + ': ' + (language === 'id' ? 'Bahasa Indonesia (Click to switch to English)' : 'English (Click to switch to Bahasa Indonesia)')"
        aria-label="Toggle Language"
      >
        <svg class="w-3.5 h-3.5 text-cyan-600 dark:text-cyan-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <circle cx="12" cy="12" r="10" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" />
        </svg>
        <div class="relative w-8 h-4 bg-slate-200 dark:bg-slate-700 rounded-full flex items-center p-0.5 transition-colors">
          <span
            class="text-[7.5px] font-bold font-mono text-center flex-1 z-10 transition-colors"
            :class="language === 'en' ? 'text-cyan-700 dark:text-cyan-300 font-extrabold' : 'text-slate-400 dark:text-slate-500'"
          >EN</span>
          <span
            class="text-[7.5px] font-bold font-mono text-center flex-1 z-10 transition-colors"
            :class="language === 'id' ? 'text-cyan-700 dark:text-cyan-300 font-extrabold' : 'text-slate-400 dark:text-slate-500'"
          >ID</span>
          <div
            class="absolute top-0.5 bottom-0.5 w-[15px] bg-white dark:bg-cyan-500 rounded-full shadow-xs transition-transform duration-200"
            :class="language === 'id' ? 'translate-x-[15px]' : 'translate-x-0'"
          ></div>
        </div>
      </button>

      <!-- 4. Theme Switch Icon Button (Light / Dark Slide Switch) -->
      <button
        type="button"
        @click="toggleTheme"
        class="h-8 px-2 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 flex items-center space-x-1.5 transition-all shadow-xs"
        :title="$t('header.appearance') + ': ' + (isDarkMode ? $t('header.dark') : $t('header.light')) + ' (Click to toggle)'"
        aria-label="Toggle Dark / Light Theme"
      >
        <span v-if="!isDarkMode" class="text-amber-500 text-xs leading-none flex-shrink-0">☀</span>
        <span v-else class="text-blue-400 text-xs leading-none flex-shrink-0">🌙</span>
        <div class="relative w-8 h-4 bg-slate-200 dark:bg-slate-700 rounded-full flex items-center p-0.5 transition-colors">
          <span class="text-[8px] text-center flex-1 z-10 leading-none">☀</span>
          <span class="text-[8px] text-center flex-1 z-10 leading-none">🌙</span>
          <div
            class="absolute top-0.5 bottom-0.5 w-[15px] bg-white dark:bg-blue-500 rounded-full shadow-xs transition-transform duration-200"
            :class="isDarkMode ? 'translate-x-[15px]' : 'translate-x-0'"
          ></div>
        </div>
      </button>

      <!-- 5. User Profile Badge -->
      <div class="flex items-center space-x-2 pl-3 border-l border-slate-200 dark:border-slate-800">
        <div class="w-6 h-6 rounded-full bg-blue-600/20 border border-blue-500/40 flex items-center justify-center text-3xs font-bold text-blue-500 dark:text-blue-400 font-sans">
          {{ user.username ? user.username.charAt(0).toUpperCase() : 'A' }}
        </div>
        <div class="hidden sm:block text-left font-sans">
          <div class="text-xs font-medium text-slate-800 dark:text-slate-200 leading-tight">{{ user.username }}</div>
          <div class="text-3xs text-slate-500 dark:text-slate-400 leading-tight">{{ user.role || $t('header.userRole') }}</div>
        </div>
      </div>
    </div>
  </header>
</template>

<script>
export default {
  name: 'Header',
  data() {
    return {
      currentTime: new Date().toLocaleTimeString('en-GB'),
      clockInterval: null,
      isLangMenuOpen: false,
      isThemeMenuOpen: false,
    };
  },
  computed: {
    systemStatus() {
      return this.$store.state.systemStatus;
    },
    user() {
      return this.$store.state.user;
    },
    sidebarAutoHide() {
      return (this.$store.state.uiPreferences && this.$store.state.uiPreferences.sidebarAutoHide) || false;
    },
    themeSetting() {
      return (this.$store.state.uiPreferences && this.$store.state.uiPreferences.theme) || 'system';
    },
    isDarkMode() {
      const eff = this.$store.state.uiPreferences && this.$store.state.uiPreferences.effectiveTheme;
      if (eff) return eff === 'dark';
      return this.themeSetting === 'dark' || (this.themeSetting === 'system' && typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches);
    },
    language() {
      return (this.$store.state.uiPreferences && this.$store.state.uiPreferences.language) || 'en';
    },
    currentLanguageLabel() {
      return this.language === 'id' ? 'ID' : 'EN';
    },
    currentThemeLabel() {
      if (this.themeSetting === 'light') return this.$t('header.light');
      if (this.themeSetting === 'dark') return this.$t('header.dark');
      return this.$t('header.system');
    },
    activeAlarmsCount() {
      const active = this.$store.state.activeAlarms || [];
      return active.filter(a => a && a.status === 'ACTIVE').length;
    },
    overallAppStatus() {
      const s = this.systemStatus;
      if (!s) return 'HEALTHY';
      if (s.database_status && typeof s.database_status === 'string' && s.database_status.toLowerCase().includes('fail')) {
        return 'FAILURE';
      }
      // Distinguishes application health from systemStatus.online_devices communication health
      if (s.offline_devices > 0 || this.activeAlarmsCount > 0 || (s.critical_alarms && s.critical_alarms > 0) || (this.systemStatus && this.systemStatus.online_devices === 0)) {
        return 'DEGRADED';
      }
      return 'HEALTHY';
    },
    appStatusBadgeClass() {
      if (this.overallAppStatus === 'FAILURE') {
        return 'bg-rose-500/20 text-rose-500 dark:text-rose-400 border-rose-500/40';
      }
      if (this.overallAppStatus === 'DEGRADED') {
        return 'bg-amber-500/20 text-amber-600 dark:text-amber-400 border-amber-500/40';
      }
      return 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border-emerald-500/40';
    },
  },
  mounted() {
    this.clockInterval = setInterval(() => {
      this.currentTime = new Date().toLocaleTimeString('en-GB');
    }, 1000);
  },
  beforeDestroy() {
    if (this.clockInterval) {
      clearInterval(this.clockInterval);
    }
  },
  methods: {
    toggleSidebarAutoHide() {
      this.$store.dispatch('toggleSidebarAutoHide');
      this.$emit('toggle-autohide');
    },
    toggleLanguage() {
      const next = this.language === 'id' ? 'en' : 'id';
      this.$store.dispatch('setLanguage', next);
    },
    toggleTheme() {
      const next = this.isDarkMode ? 'light' : 'dark';
      this.$store.dispatch('setTheme', next);
    },
    toggleLangMenu() {
      this.toggleLanguage();
    },
    toggleThemeMenu() {
      this.toggleTheme();
    },
    selectLanguage(lang) {
      this.$store.dispatch('setLanguage', lang);
      this.isLangMenuOpen = false;
    },
    selectTheme(theme) {
      this.$store.dispatch('setTheme', theme);
      this.isThemeMenuOpen = false;
    },
    handleOutsideClick(event) {
      if (this.$refs.langDropdown && !this.$refs.langDropdown.contains(event.target)) {
        this.isLangMenuOpen = false;
      }
      if (this.$refs.themeDropdown && !this.$refs.themeDropdown.contains(event.target)) {
        this.isThemeMenuOpen = false;
      }
    },
  },
};
</script>
