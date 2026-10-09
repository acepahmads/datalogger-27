<template>
  <aside class="w-60 bg-white dark:bg-[#0B0F19] border-r border-slate-200 dark:border-slate-800/60 flex flex-col flex-shrink-0 select-none text-slate-700 dark:text-slate-300 font-sans transition-colors duration-150">
    <!-- Brand Header -->
    <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800/40">
      <div class="text-sm font-black text-slate-900 dark:text-white tracking-wider">DATALOGGER</div>
      <div class="text-2xs text-slate-500 dark:text-slate-400 font-medium">Analysis Application</div>
      <div class="flex items-center space-x-1.5 mt-2">
        <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
        <span class="text-3xs text-emerald-600 dark:text-emerald-400 font-medium tracking-wide">{{ $t('navigation.edgeSystem') }}</span>
      </div>
    </div>

    <!-- Navigation List -->
    <nav class="flex-1 overflow-y-auto px-3 py-4 space-y-6">
      <!-- DEVELOPMENT -->
      <div>
        <div class="px-3 text-3xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2 font-sans">
          {{ $t('navigation.development') }}
        </div>
        <div class="space-y-1">
          <router-link to="/development" class="nav-item" :class="{ 'nav-active': $route.path === '/development' || $route.path === '/development/phases' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <rect x="3" y="3" width="7" height="7" rx="1"></rect>
              <rect x="14" y="3" width="7" height="7" rx="1"></rect>
              <rect x="14" y="14" width="7" height="7" rx="1"></rect>
              <rect x="3" y="14" width="7" height="7" rx="1"></rect>
            </svg>
            <span>{{ $t('navigation.overviewPhases') }}</span>
          </router-link>

          <router-link to="/development/tasks" class="nav-item" :class="{ 'nav-active': $route.path === '/development/tasks' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 11l3 3L22 4"></path>
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"></path>
            </svg>
            <span>{{ $t('navigation.tasksManager') }}</span>
            <span class="ml-auto text-3xs font-mono px-1.5 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
              {{ progress.total_tasks_count || 117 }}
            </span>
          </router-link>

          <router-link to="/development/activity" class="nav-item" :class="{ 'nav-active': $route.path === '/development/activity' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
            </svg>
            <span>{{ $t('navigation.activityLog') }}</span>
          </router-link>

          <router-link to="/development/release" class="nav-item" :class="{ 'nav-active': $route.path === '/development/release' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"></path>
              <polyline points="14 2 14 8 20 8"></polyline>
              <line x1="16" y1="13" x2="8" y2="13"></line>
              <line x1="16" y1="17" x2="8" y2="17"></line>
              <polyline points="10 9 9 9 8 9"></polyline>
            </svg>
            <span>{{ $t('navigation.releaseDocs') }}</span>
          </router-link>
        </div>
      </div>

      <!-- MONITORING -->
      <div>
        <div class="px-3 text-3xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2 font-sans">
          {{ $t('navigation.monitoring') }}
        </div>
        <div class="space-y-1">
          <router-link to="/monitoring/system" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/system' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <rect x="4" y="4" width="16" height="16" rx="2"></rect>
              <rect x="9" y="9" width="6" height="6"></rect>
              <line x1="9" y1="1" x2="9" y2="4"></line>
              <line x1="15" y1="1" x2="15" y2="4"></line>
              <line x1="9" y1="20" x2="9" y2="23"></line>
              <line x1="15" y1="20" x2="15" y2="23"></line>
              <line x1="20" y1="9" x2="23" y2="9"></line>
              <line x1="20" y1="14" x2="23" y2="14"></line>
              <line x1="1" y1="9" x2="4" y2="9"></line>
              <line x1="1" y1="14" x2="4" y2="14"></line>
            </svg>
            <span>{{ $t('navigation.systemMonitoring') }}</span>
          </router-link>

          <router-link to="/monitoring/devices" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/devices' || $route.path === '/monitoring/parameters' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <circle cx="12" cy="12" r="2"></circle>
              <path d="M16.24 7.76a6 6 0 0 1 0 8.49m-8.48-.01a6 6 0 0 1 0-8.49m11.31-2.82a10 10 0 0 1 0 14.14m-14.14 0a10 10 0 0 1 0-14.14"></path>
            </svg>
            <span>{{ $t('navigation.devicesSensors') }}</span>
          </router-link>

          <router-link to="/monitoring/telemetry" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/telemetry' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-emerald-500 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
            </svg>
            <span class="font-medium text-emerald-600 dark:text-emerald-300">{{ $t('navigation.liveTelemetry') }}</span>
            <span class="ml-auto flex h-2 w-2 relative">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
          </router-link>

          <router-link to="/monitoring/analysis" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/analysis' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-blue-500 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
            </svg>
            <span class="font-medium text-slate-200">{{ $t('navigation.telemetryAnalysis') }}</span>
          </router-link>

          <router-link to="/dashboard" class="nav-item" :class="{ 'nav-active': $route.path === '/dashboard' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
            </svg>
            <span>{{ $t('common.quality') }}</span>
          </router-link>
        </div>
      </div>

      <!-- ALARMS -->
      <div>
        <div class="px-3 text-3xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2 font-sans flex items-center justify-between">
          <span>{{ $t('navigation.alarmsEvents') }}</span>
          <span v-if="systemStatus.active_alarms > 0" class="w-1.5 h-1.5 rounded-full bg-rose-500 animate-ping"></span>
        </div>
        <div class="space-y-1">
          <router-link to="/alarm/active" class="nav-item" :class="{ 'nav-active': $route.path === '/alarm/active' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
            </svg>
            <span>{{ $t('navigation.activeAlarms') }}</span>
            <span v-if="systemStatus.active_alarms > 0" class="ml-auto text-3xs font-mono px-1.5 py-0.5 rounded-full bg-rose-100 dark:bg-rose-950 text-rose-600 dark:text-rose-400 border border-rose-300 dark:border-rose-800">
              {{ systemStatus.active_alarms }}
            </span>
          </router-link>

          <router-link to="/alarm/history" class="nav-item" :class="{ 'nav-active': $route.path === '/alarm/history' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
            <span>{{ $t('navigation.alarmHistory') }}</span>
          </router-link>
        </div>
      </div>

      <!-- SYSTEM -->
      <div>
        <div class="px-3 text-3xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2 font-sans">
          {{ $t('navigation.systemAdmin') }}
        </div>
        <div class="space-y-1">
          <router-link to="/system/logs" class="nav-item" :class="{ 'nav-active': $route.path === '/system/logs' || $route.path === '/system/audit' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <polyline points="4 17 10 11 4 5"></polyline>
              <line x1="12" y1="19" x2="20" y2="19"></line>
            </svg>
            <span>{{ $t('navigation.systemLogs') }}</span>
          </router-link>

          <router-link to="/administration/config" class="nav-item" :class="{ 'nav-active': $route.path === '/administration/config' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <circle cx="12" cy="12" r="3"></circle>
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
            </svg>
            <span>{{ $t('navigation.configuration') }}</span>
          </router-link>

          <router-link to="/administration/backup" class="nav-item" :class="{ 'nav-active': $route.path === '/administration/backup' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 7v10c0 2 1.5 3 3.5 3h9c2 0 3.5-1 3.5-3V7M4 7c0-2 1.5-3 3.5-3h9c2 0 3.5 1 3.5 3M4 7h16m-8 4v6m0 0l-3-3m3 3l3-3"></path>
            </svg>
            <span>{{ $t('navigation.backupRestore') }}</span>
          </router-link>

          <router-link to="/administration/retention" class="nav-item" :class="{ 'nav-active': $route.path === '/administration/retention' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
            </svg>
            <span>{{ $t('navigation.retentionStorage') }}</span>
          </router-link>

          <router-link to="/administration/config" class="nav-item">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <circle cx="18" cy="5" r="3"></circle>
              <circle cx="6" cy="12" r="3"></circle>
              <circle cx="18" cy="19" r="3"></circle>
              <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line>
              <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>
            </svg>
            <span>{{ $t('navigation.integrations') }}</span>
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Bottom Database Indicator (MariaDB Edge ● LOCAL) -->
    <div class="p-4 border-t border-slate-200 dark:border-slate-800/50 bg-slate-50 dark:bg-[#0A0E18]/60">
      <div class="flex items-center justify-between text-2xs mb-1.5">
        <span class="font-medium text-slate-700 dark:text-slate-200">{{ $t('navigation.edgeDatabase') }}</span>
        <span class="flex items-center text-emerald-600 dark:text-emerald-400 text-3xs font-semibold tracking-wide">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1"></span>
          LOCAL
        </span>
      </div>
      <div class="text-3xs text-slate-500 font-mono flex items-center justify-between">
        <span>127.0.0.1:3306</span>
        <span class="text-slate-600 dark:text-slate-400">10.4.x</span>
      </div>
    </div>
  </aside>
</template>

<script>
export default {
  name: 'Sidebar',
  computed: {
    progress() {
      return this.$store.state.progress;
    },
    systemStatus() {
      return this.$store.state.systemStatus;
    },
  },
};
</script>

<style scoped>
.nav-item {
  @apply relative flex items-center px-3 py-2 rounded-lg text-xs font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100 hover:bg-slate-100 dark:hover:bg-slate-800/40 transition-colors;
}
.nav-active {
  @apply bg-blue-50 dark:bg-blue-600/10 text-blue-600 dark:text-blue-400 font-semibold;
}
.nav-active .nav-icon {
  @apply text-blue-600 dark:text-blue-400;
}
.active-indicator {
  @apply hidden;
}
.nav-active .active-indicator {
  @apply block absolute left-0 top-2 bottom-2 w-0.5 bg-blue-600 dark:bg-blue-500 rounded-r;
}
</style>
