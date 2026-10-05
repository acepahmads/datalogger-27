<template>
  <header class="bg-[#0B0F19]/95 backdrop-blur-md border-b border-slate-800/60 text-slate-200 px-5 py-2.5 flex items-center justify-between sticky top-0 z-40 select-none">
    <!-- Left: Logo & App Title -->
    <div class="flex items-center space-x-3">
      <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-600 to-indigo-700 flex items-center justify-center shadow-soft text-white font-bold text-xs tracking-wider">
        DL
      </div>
      <div class="flex items-center space-x-2">
        <h1 class="text-xs font-bold tracking-tight text-white uppercase font-sans">
          Datalogger Analysis Application
        </h1>
        <span class="bg-blue-500/10 text-blue-400 border border-blue-500/20 text-3xs px-2 py-0.5 rounded-full font-mono font-medium">
          v1.0.0
        </span>
      </div>
    </div>

    <!-- Center: Clean Inline Host, Status & Live Time (No heavy boxes) -->
    <div class="hidden lg:flex items-center space-x-5 text-2xs">
      <!-- Host & OS -->
      <div class="flex items-center space-x-1.5 text-slate-400">
        <span class="text-slate-500">Host:</span>
        <span class="font-mono text-slate-200 font-medium">{{ systemStatus.hostname || 'LAPTOP-5UE284DP' }}</span>
        <span class="text-slate-600">·</span>
        <span class="text-slate-300 font-mono">{{ systemStatus.os || 'Windows · amd64' }}</span>
      </div>

      <!-- Status Indicator -->
      <div class="flex items-center space-x-1.5 text-slate-400">
        <span class="text-slate-500">Status:</span>
        <span class="flex items-center text-emerald-400 font-semibold font-mono text-3xs tracking-wider">
          <span class="relative flex h-2 w-2 mr-1.5">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
          </span>
          ONLINE
        </span>
      </div>

      <!-- Live Clock -->
      <div class="flex items-center space-x-1.5 text-slate-400">
        <span class="text-slate-500">Time:</span>
        <span class="font-mono text-slate-200 font-medium tabular-nums">{{ currentTime }}</span>
      </div>
    </div>

    <!-- Right: Controls & User -->
    <div class="flex items-center space-x-3">
      <!-- Language Toggle -->
      <button @click="toggleLanguage"
              class="px-2 py-1 rounded-md bg-slate-800/40 hover:bg-slate-800 text-slate-300 hover:text-white text-3xs font-mono font-semibold transition-colors"
              title="Change Language">
        {{ language.toUpperCase() }}
      </button>

      <!-- Theme Toggle -->
      <button @click="toggleTheme"
              class="p-1.5 rounded-md bg-slate-800/40 hover:bg-slate-800 text-slate-400 hover:text-white transition-colors"
              title="Toggle Theme">
        <svg v-if="theme === 'dark'" class="w-3.5 h-3.5 text-amber-400" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 2a1 1 0 011 1v1a1 1 0 11-2 0V3a1 1 0 011-1zm4 8a4 4 0 11-8 0 4 4 0 018 0zm-.464 4.95l.707.707a1 1 0 001.414-1.414l-.707-.707a1 1 0 00-1.414 1.414zm2.12-10.607a1 1 0 010 1.414l-.706.707a1 1 0 11-1.414-1.414l.707-.707a1 1 0 011.414 0zM17 11a1 1 0 100-2h-1a1 1 0 100 2h1zm-7 4a1 1 0 011 1v1a1 1 0 11-2 0v-1a1 1 0 011-1zM5.05 6.464A1 1 0 106.465 5.05l-.708-.707a1 1 0 00-1.414 1.414l.707.707zm1.414 8.486l-.707.707a1 1 0 01-1.414-1.414l.707-.707a1 1 0 011.414 1.414zM4 11a1 1 0 100-2H3a1 1 0 000 2h1z" clip-rule="evenodd"></path>
        </svg>
        <svg v-else class="w-3.5 h-3.5 text-slate-400" fill="currentColor" viewBox="0 0 20 20">
          <path d="M17.293 13.293A8 8 0 016.707 2.707a8.001 8.001 0 1010.586 10.586z"></path>
        </svg>
      </button>

      <!-- User Profile Badge -->
      <div class="flex items-center space-x-2 pl-3 border-l border-slate-800">
        <div class="w-6 h-6 rounded-full bg-blue-600/20 border border-blue-500/40 flex items-center justify-center text-3xs font-bold text-blue-400 font-sans">
          {{ user.username ? user.username.charAt(0).toUpperCase() : 'A' }}
        </div>
        <div class="hidden sm:block text-left font-sans">
          <div class="text-xs font-medium text-slate-200 leading-tight">{{ user.username }}</div>
          <div class="text-3xs text-slate-400 leading-tight">{{ user.role || 'Lead Engineer' }}</div>
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
    };
  },
  computed: {
    systemStatus() {
      return this.$store.state.systemStatus;
    },
    user() {
      return this.$store.state.user;
    },
    theme() {
      return this.$store.state.theme;
    },
    language() {
      return this.$store.state.language;
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
    toggleTheme() {
      const nextTheme = this.theme === 'dark' ? 'light' : 'dark';
      this.$store.commit('SET_THEME', nextTheme);
    },
    toggleLanguage() {
      const nextLang = this.language === 'en' ? 'id' : 'en';
      this.$store.commit('SET_LANGUAGE', nextLang);
    },
  },
};
</script>
