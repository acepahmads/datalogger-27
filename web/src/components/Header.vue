<template>
  <header class="bg-white/95 dark:bg-[#0B0F19]/95 backdrop-blur-md border-b border-slate-200 dark:border-slate-800/60 text-slate-800 dark:text-slate-200 px-5 py-2.5 flex items-center justify-between sticky top-0 z-40 select-none transition-colors duration-150">
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

    <!-- Center: Clean Inline Host/Station, App Health vs Comm Health, & Live Time -->
    <div class="hidden lg:flex items-center space-x-4 text-2xs">
      <!-- Station / Host & OS -->
      <div class="flex items-center space-x-1.5 text-slate-500 dark:text-slate-400">
        <span class="text-slate-400 dark:text-slate-500 font-semibold">{{ $t('operator.station') || 'Station' }}:</span>
        <span class="font-mono text-cyan-600 dark:text-cyan-400 font-bold">{{ systemStatus.hostname || 'Edge-01' }}</span>
        <span class="text-slate-300 dark:text-slate-600">·</span>
        <span class="text-slate-600 dark:text-slate-400 font-mono text-3xs">{{ systemStatus.os || 'Linux · arm64' }}</span>
      </div>

      <!-- Application Health (HEALTHY / DEGRADED / FAILURE) -->
      <div class="flex items-center space-x-1.5">
        <span class="text-slate-400 dark:text-slate-500 font-semibold">{{ $t('operator.appStatus') || 'App' }}:</span>
        <span class="px-2 py-0.5 rounded-full font-mono font-bold text-3xs border flex items-center space-x-1" :class="appStatusBadgeClass">
          <span class="w-1.5 h-1.5 rounded-full" :class="overallAppStatus === 'HEALTHY' ? 'bg-emerald-400' : overallAppStatus === 'DEGRADED' ? 'bg-amber-400 animate-pulse' : 'bg-rose-400 animate-ping'"></span>
          <span>{{ overallAppStatus }}</span>
        </span>
      </div>

      <!-- Device Communication Health -->
      <div class="flex items-center space-x-1.5 text-slate-500 dark:text-slate-400">
        <span class="text-slate-400 dark:text-slate-500 font-semibold">{{ $t('operator.commHealth') || 'Comm' }}:</span>
        <span class="font-mono font-semibold text-3xs" :class="systemStatus.offline_devices > 0 ? 'text-amber-500' : 'text-emerald-500'">
          {{ systemStatus.online_devices != null ? systemStatus.online_devices : 3 }}/{{ systemStatus.total_devices != null ? systemStatus.total_devices : 3 }} {{ $t('header.online') || 'ONLINE' }}
        </span>
      </div>

      <!-- Live Clock -->
      <div class="flex items-center space-x-1.5 text-slate-500 dark:text-slate-400 pl-1 border-l border-slate-200 dark:border-slate-800">
        <span class="text-slate-400 dark:text-slate-500">{{ $t('header.time') }}:</span>
        <span class="font-mono text-slate-700 dark:text-slate-200 font-medium tabular-nums">{{ currentTime }}</span>
      </div>
    </div>

    <!-- Right: Controls & User -->
    <div class="flex items-center space-x-2.5">
      <!-- Operator Console Quick Button -->
      <router-link
        to="/monitoring/operator"
        class="px-2.5 py-1.5 rounded-lg border text-2xs font-bold font-mono transition-colors flex items-center space-x-1.5 shadow-xs"
        :class="$route.path.includes('/operator') ? 'bg-cyan-500/20 text-cyan-400 border-cyan-500/50 ring-1 ring-cyan-500/30' : 'bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/80 dark:hover:bg-slate-700/80 text-cyan-600 dark:text-cyan-400 border-cyan-500/30'"
        :title="$t('operator.title') || 'Operator Command Center'"
      >
        <span class="relative flex h-2 w-2">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2 w-2 bg-cyan-500"></span>
        </span>
        <span class="hidden sm:inline">{{ $t('navigation.operatorConsole') || 'Operator Console' }}</span>
      </router-link>

      <!-- Notification Bell with Unread / Active Alarms Badge -->
      <router-link
        to="/alarm/active"
        class="relative p-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 transition-colors"
        :title="$t('navigation.alarmsEvents') || 'Active Alarms'"
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
      <!-- Language Selector Dropdown -->
      <div class="relative" ref="langDropdown">
        <button
          @click="toggleLangMenu"
          class="flex items-center space-x-1.5 px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 text-2xs font-medium transition-colors"
          :title="$t('header.language')"
          aria-haspopup="true"
          :aria-expanded="isLangMenuOpen"
        >
          <span class="font-bold text-3xs tracking-wide uppercase font-mono">{{ currentLanguageLabel }}</span>
          <svg class="w-3 h-3 text-slate-400 transition-transform duration-200" :class="{ 'rotate-180': isLangMenuOpen }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
          </svg>
        </button>

        <!-- Dropdown Menu -->
        <div
          v-if="isLangMenuOpen"
          class="absolute right-0 mt-1.5 w-44 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-700/80 shadow-xl py-1.5 z-50 text-xs font-sans"
        >
          <div class="px-3 py-1 text-3xs font-semibold text-slate-400 uppercase tracking-wider border-b border-slate-100 dark:border-slate-800/80 mb-1">
            {{ $t('header.language') }}
          </div>
          <button
            @click="selectLanguage('en')"
            class="w-full text-left px-3 py-1.5 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/70 transition-colors"
            :class="language === 'en' ? 'text-blue-600 dark:text-blue-400 font-semibold' : 'text-slate-700 dark:text-slate-300'"
          >
            <span>English</span>
            <span v-if="language === 'en'" class="text-xs text-blue-500">✓</span>
          </button>
          <button
            @click="selectLanguage('id')"
            class="w-full text-left px-3 py-1.5 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/70 transition-colors"
            :class="language === 'id' ? 'text-blue-600 dark:text-blue-400 font-semibold' : 'text-slate-700 dark:text-slate-300'"
          >
            <span>Bahasa Indonesia</span>
            <span v-if="language === 'id'" class="text-xs text-blue-500">✓</span>
          </button>
        </div>
      </div>

      <!-- Theme Selector Dropdown (Light / Dark / System) -->
      <div class="relative" ref="themeDropdown">
        <button
          @click="toggleThemeMenu"
          class="flex items-center space-x-1.5 px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 text-2xs font-medium transition-colors"
          :title="$t('header.appearance')"
          aria-haspopup="true"
          :aria-expanded="isThemeMenuOpen"
        >
          <!-- Active Icon -->
          <span v-if="themeSetting === 'light'" class="text-amber-500 text-xs">☀</span>
          <span v-else-if="themeSetting === 'dark'" class="text-blue-400 text-xs">🌙</span>
          <span v-else class="text-indigo-500 dark:text-indigo-400 text-xs">◐</span>
          <span class="capitalize text-3xs font-mono font-semibold">{{ currentThemeLabel }}</span>
          <svg class="w-3 h-3 text-slate-400 transition-transform duration-200" :class="{ 'rotate-180': isThemeMenuOpen }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path>
          </svg>
        </button>

        <!-- Dropdown Menu -->
        <div
          v-if="isThemeMenuOpen"
          class="absolute right-0 mt-1.5 w-44 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-700/80 shadow-xl py-1.5 z-50 text-xs font-sans"
        >
          <div class="px-3 py-1 text-3xs font-semibold text-slate-400 uppercase tracking-wider border-b border-slate-100 dark:border-slate-800/80 mb-1">
            {{ $t('header.appearance') }}
          </div>
          <button
            @click="selectTheme('light')"
            class="w-full text-left px-3 py-1.5 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/70 transition-colors"
            :class="themeSetting === 'light' ? 'text-amber-600 dark:text-amber-400 font-semibold' : 'text-slate-700 dark:text-slate-300'"
          >
            <div class="flex items-center space-x-2">
              <span class="text-amber-500">☀</span>
              <span>{{ $t('header.light') }}</span>
            </div>
            <span v-if="themeSetting === 'light'" class="text-xs text-amber-500">✓</span>
          </button>
          <button
            @click="selectTheme('dark')"
            class="w-full text-left px-3 py-1.5 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/70 transition-colors"
            :class="themeSetting === 'dark' ? 'text-blue-600 dark:text-blue-400 font-semibold' : 'text-slate-700 dark:text-slate-300'"
          >
            <div class="flex items-center space-x-2">
              <span class="text-blue-400">🌙</span>
              <span>{{ $t('header.dark') }}</span>
            </div>
            <span v-if="themeSetting === 'dark'" class="text-xs text-blue-500">✓</span>
          </button>
          <button
            @click="selectTheme('system')"
            class="w-full text-left px-3 py-1.5 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/70 transition-colors"
            :class="themeSetting === 'system' ? 'text-indigo-600 dark:text-indigo-400 font-semibold' : 'text-slate-700 dark:text-slate-300'"
          >
            <div class="flex items-center space-x-2">
              <span class="text-indigo-500 dark:text-indigo-400">◐</span>
              <span>{{ $t('header.system') }}</span>
            </div>
            <span v-if="themeSetting === 'system'" class="text-xs text-indigo-500">✓</span>
          </button>
        </div>
      </div>

      <!-- User Profile Badge -->
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
    themeSetting() {
      return (this.$store.state.uiPreferences && this.$store.state.uiPreferences.theme) || 'system';
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
      if (s.offline_devices > 0 || this.activeAlarmsCount > 0 || (s.critical_alarms && s.critical_alarms > 0)) {
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
    document.addEventListener('click', this.handleOutsideClick);
  },
  beforeDestroy() {
    if (this.clockInterval) {
      clearInterval(this.clockInterval);
    }
    document.removeEventListener('click', this.handleOutsideClick);
  },
  methods: {
    toggleLangMenu() {
      this.isLangMenuOpen = !this.isLangMenuOpen;
      if (this.isLangMenuOpen) this.isThemeMenuOpen = false;
    },
    toggleThemeMenu() {
      this.isThemeMenuOpen = !this.isThemeMenuOpen;
      if (this.isThemeMenuOpen) this.isLangMenuOpen = false;
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
