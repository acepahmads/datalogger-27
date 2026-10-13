<template>
  <div class="operator-command-center font-sans text-slate-800 dark:text-slate-100 select-none pb-14 transition-colors duration-150">
    <!-- Sub-Header / SCADA Operational Status Bar -->
    <div class="flex flex-wrap items-center justify-between gap-2 px-3 py-2 mb-2 rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] shadow-xs dark:shadow-lg transition-colors">
      <!-- Left: Station & Console Title -->
      <div class="flex items-center space-x-3">
        <div class="flex items-center space-x-2">
          <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center text-white shadow-glow-blue">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5m.75-9l3-3 2.25 2.25L15 6" />
            </svg>
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h1 class="text-xs font-black tracking-wider uppercase text-slate-900 dark:text-white font-mono flex items-center gap-1.5">
                <span>{{ $t('operator.title') || 'OPERATOR COMMAND CENTER' }}</span>
                <span class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-cyan-500/10 dark:bg-cyan-500/20 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30 dark:border-cyan-500/40">
                  SCADA
                </span>
              </h1>
            </div>
            <div class="text-3xs text-slate-500 dark:text-slate-400 font-mono flex items-center space-x-2">
              <span>{{ $t('operator.station') || 'Station' }}: <strong class="text-cyan-600 dark:text-cyan-400">{{ systemStatus.hostname || 'Edge-Node-01' }}</strong></span>
              <span>·</span>
              <span>{{ $t('operator.site') || 'Site' }}: <span class="text-slate-700 dark:text-slate-300">Local Autonomous Edge</span></span>
            </div>
          </div>
        </div>
      </div>

      <!-- Center: Acquisition & Health Status Badges -->
      <div class="hidden md:flex items-center space-x-2.5 text-2xs font-mono">
        <!-- App Status -->
        <div class="flex items-center space-x-1.5 px-2.5 py-1 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B]">
          <span class="text-slate-500 dark:text-slate-400 text-3xs">{{ $t('operator.systemHealth') || 'App' }}:</span>
          <span class="font-bold flex items-center space-x-1" :class="appHealthColorClass">
            <span class="w-2 h-2 rounded-full" :class="appHealthDotClass"></span>
            <span>{{ overallAppHealth }}</span>
          </span>
        </div>

        <!-- Acquisition Engine Status -->
        <div class="flex items-center space-x-1.5 px-2.5 py-1 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B]">
          <span class="text-slate-500 dark:text-slate-400 text-3xs">{{ $t('operator.acquisitionStatus') || 'Acquisition' }}:</span>
          <span class="text-emerald-600 dark:text-emerald-400 font-bold flex items-center space-x-1">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-ping"></span>
            <span>{{ $t('operator.pollingActive') || 'POLLING ACTIVE' }}</span>
          </span>
        </div>

        <!-- Telemetry Rate -->
        <div class="flex items-center space-x-1.5 px-2.5 py-1 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B]">
          <span class="text-slate-500 dark:text-slate-400 text-3xs">{{ $t('operator.throughput') || 'Throughput' }}:</span>
          <span class="text-cyan-600 dark:text-cyan-400 font-bold tabular-nums">{{ (systemStatus.data_per_sec || telemetryRate).toFixed(1) }} pkt/s</span>
        </div>

        <!-- Live Clock in SCADA bar (visible especially when global header is hidden) -->
        <div class="flex items-center space-x-1 px-2.5 py-1 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-200 font-bold text-3xs tabular-nums">
          <span class="text-slate-400 dark:text-slate-500">🕒</span>
          <span>{{ currentTime }}</span>
        </div>
      </div>

      <!-- Right: Audio Controls, Theme, Language, Refresh, & Fullscreen Toggles -->
      <div class="flex items-center space-x-1.5">
        <!-- Quick Theme Toggle (☀ / 🌙) -->
        <button
          @click="toggleTheme"
          class="px-2 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-300 text-3xs font-mono transition"
          :title="isDark ? ($t('operator.lightMode') || 'Switch to Light Mode') : ($t('operator.darkMode') || 'Switch to Dark Mode')"
        >
          <span>{{ isDark ? '🌙' : '☀' }}</span>
        </button>

        <!-- Quick Language Toggle (EN / ID) -->
        <button
          @click="toggleLanguage"
          class="px-2 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-300 text-3xs font-mono font-bold transition"
          :title="$t('header.language') || 'Language'"
        >
          {{ currentLocale.toUpperCase() }}
        </button>

        <!-- Sound Mute/Unmute Toggle -->
        <button
          @click="toggleAudioMute"
          class="px-2.5 py-1 rounded-lg border text-3xs font-mono font-semibold transition-colors flex items-center space-x-1"
          :class="isAudioMuted ? 'bg-amber-500/10 border-amber-500/30 text-amber-600 dark:text-amber-400' : 'bg-cyan-500/10 border-cyan-500/30 text-cyan-600 dark:text-cyan-400 hover:bg-cyan-500/20'"
          :title="isAudioMuted ? ($t('operator.audioMuted') || 'Sound Muted') : ($t('operator.audioActive') || 'Sound Active')"
        >
          <span>{{ isAudioMuted ? '🔇' : '🔊' }}</span>
          <span class="hidden sm:inline">{{ isAudioMuted ? ($t('operator.audioMuted') || 'MUTED') : ($t('operator.audioActive') || 'AUDIO') }}</span>
        </button>

        <!-- Test Chime -->
        <button
          @click="testAlarmChime('CRITICAL')"
          class="px-2 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-300 text-3xs font-mono transition"
          :title="$t('operator.testAudioCritical') || 'Test Emergency Chime'"
        >
          🚨
        </button>

        <!-- Manual Refresh -->
        <button
          @click="refreshDashboard"
          :disabled="isRefreshing"
          class="p-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white transition disabled:opacity-50"
          :title="$t('common.refresh') || 'Refresh All'"
        >
          <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': isRefreshing }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>

        <!-- Fullscreen / Sidebar & Top Panel Layout Toggle -->
        <button
          @click="$emit('toggle-sidebar')"
          class="px-2.5 py-1 rounded-lg border text-3xs font-mono font-semibold transition flex items-center space-x-1"
          :class="isSidebarCollapsed
            ? 'bg-cyan-600/10 hover:bg-cyan-600/20 text-cyan-600 dark:text-cyan-400 border-cyan-500/40 ring-1 ring-cyan-500/30'
            : 'bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] border-slate-200 dark:border-[#20354B] text-cyan-600 dark:text-cyan-400'"
          :title="isSidebarCollapsed ? ($t('operator.exitFullWidth') || 'Exit Full Width (Restore Panels)') : ($t('operator.fullWidth') || 'Full Width SCADA (Hide Panels)')"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path v-if="!isSidebarCollapsed" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
            <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 4H5a1 1 0 00-1 1v4m0 0l5-5m6 0h4a1 1 0 011 1v4m0 0l-5-5M9 20H5a1 1 0 01-1-1v-4m0 0l5 5m6 0h4a1 1 0 001-1v-4m0 0l-5 5" />
          </svg>
          <span class="hidden sm:inline">{{ isSidebarCollapsed ? ($t('operator.exitFullWidth') || 'EXIT FULL') : ($t('operator.fullWidth') || 'FULL WIDTH') }}</span>
        </button>
      </div>
    </div>

    <!-- C. FOUR OPERATIONAL SUMMARY CARDS -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-2 mb-2.5">
      <!-- 1. Device / System Status -->
      <div
        class="scada-panel p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border relative overflow-hidden transition"
        :class="summaryStats.offlineCount > 0
          ? 'border-rose-500/70 ring-1 ring-rose-500/30'
          : 'border-slate-200 dark:border-[#20354B] hover:border-cyan-500/50'"
      >
        <div class="flex items-center justify-between text-3xs font-mono text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-1">
          <div class="flex items-center space-x-1.5">
            <span
              class="w-2 h-2 rounded-full"
              :class="summaryStats.offlineCount > 0
                ? (summaryStats.onlineCount === 0 ? 'bg-rose-500 animate-ping' : 'bg-rose-500 animate-pulse')
                : 'bg-emerald-500'"
            ></span>
            <span class="font-bold text-slate-800 dark:text-slate-200">{{ $t('operator.summaryDevices') || 'Device Fleet & Comm' }}</span>
          </div>
          <span
            class="font-bold"
            :class="summaryStats.offlineCount > 0 ? 'text-rose-600 dark:text-rose-400' : 'text-cyan-600 dark:text-cyan-400'"
          >
            {{ summaryStats.onlineCount }}/{{ summaryStats.totalDevices }} ONLINE
          </span>
        </div>
        <div class="flex items-baseline justify-between mt-1">
          <div class="flex items-baseline space-x-1.5">
            <span class="text-xl font-bold font-mono text-slate-900 dark:text-white">{{ summaryStats.totalDevices }}</span>
            <span class="text-3xs text-slate-500 dark:text-slate-400 font-mono">devices</span>
          </div>
          <div class="flex items-center space-x-1 font-mono text-3xs">
            <span class="px-1.5 py-0.2 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-bold" :title="'Online: ' + summaryStats.onlineCount">
              {{ summaryStats.onlineCount }} ON
            </span>
            <span
              class="px-1.5 py-0.2 rounded font-bold"
              :class="summaryStats.offlineCount > 0 ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 animate-pulse' : 'bg-slate-100 dark:bg-slate-800 text-slate-500'"
              :title="'Offline: ' + summaryStats.offlineCount"
            >
              {{ summaryStats.offlineCount }} OFF
            </span>
            <span v-if="summaryStats.disabledCount > 0" class="px-1.5 py-0.2 rounded bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400" :title="'Disabled: ' + summaryStats.disabledCount">
              {{ summaryStats.disabledCount }} DIS
            </span>
          </div>
        </div>
        <div class="flex items-center justify-between text-4xs text-slate-500 dark:text-slate-400 font-mono mt-1 pt-1 border-t border-slate-100 dark:border-[#1a2d42]">
          <span>Latency: <strong class="text-slate-800 dark:text-slate-200">{{ systemStatus.avg_latency_ms || 1.2 }}ms</strong></span>
          <span>Success: <strong class="text-emerald-600 dark:text-emerald-400">{{ (systemStatus.success_rate || 99.8).toFixed(1) }}%</strong></span>
        </div>
      </div>

      <!-- 2. Active Alarms -->
      <div
        class="scada-panel p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border relative overflow-hidden transition cursor-pointer"
        :class="activeCriticalCount > 0 ? 'border-rose-500/60 ring-1 ring-rose-500/30' : activeAlarmsList.length > 0 ? 'border-amber-500/60' : 'border-slate-200 dark:border-[#20354B]'"
        @click="activeRecentTab = 'alarms'"
      >
        <div class="flex items-center justify-between text-3xs font-mono uppercase tracking-wider mb-1">
          <div class="flex items-center space-x-1.5">
            <span class="w-2 h-2 rounded-full" :class="activeCriticalCount > 0 ? 'bg-rose-500 animate-ping' : activeAlarmsList.length > 0 ? 'bg-amber-500' : 'bg-emerald-500'"></span>
            <span class="font-bold text-slate-800 dark:text-slate-200">{{ $t('operator.summaryAlarms') || 'Active Incidents' }}</span>
          </div>
          <router-link to="/alarm/active" class="text-3xs text-cyan-600 dark:text-cyan-400 hover:underline font-mono">
            {{ $t('operator.viewConsole') || 'Console →' }}
          </router-link>
        </div>
        <div class="flex items-baseline justify-between mt-1">
          <div class="flex items-baseline space-x-1.5">
            <span class="text-xl font-bold font-mono" :class="activeCriticalCount > 0 ? 'text-rose-600 dark:text-rose-400' : activeAlarmsList.length > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'">
              {{ activeAlarmsList.length }}
            </span>
            <span class="text-3xs text-slate-500 dark:text-slate-400 font-mono">active</span>
          </div>
          <div class="flex items-center space-x-1 font-mono text-3xs">
            <span class="px-1.5 py-0.2 rounded font-bold" :class="activeCriticalCount > 0 ? 'bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/30 animate-pulse' : 'bg-slate-100 dark:bg-slate-800 text-slate-500'">
              {{ activeCriticalCount }} CRIT
            </span>
            <span class="px-1.5 py-0.2 rounded font-bold" :class="activeWarningCount > 0 ? 'bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/30' : 'bg-slate-100 dark:bg-slate-800 text-slate-500'">
              {{ activeWarningCount }} WARN
            </span>
            <span class="px-1.5 py-0.2 rounded font-bold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
              {{ activeAckedCount }} ACK
            </span>
          </div>
        </div>
        <div class="flex items-center justify-between text-4xs text-slate-500 dark:text-slate-400 font-mono mt-1 pt-1 border-t border-slate-100 dark:border-[#1a2d42]">
          <span v-if="highestPriorityAlarm" class="truncate max-w-[170px] text-rose-600 dark:text-rose-400 font-semibold" :title="highestPriorityAlarm.message">
            Top: {{ highestPriorityAlarm.alarm_code }}
          </span>
          <span v-else class="text-emerald-600 dark:text-emerald-400">
            ✓ Normal Operating Limits
          </span>
          <span class="text-slate-500 font-mono">{{ activeAlarmsList.length }} incident(s)</span>
        </div>
      </div>

      <!-- 3. Latest Telemetry & Freshness -->
      <div class="scada-panel p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] relative overflow-hidden group hover:border-cyan-500/50 transition">
        <div class="flex items-center justify-between text-3xs font-mono text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-1">
          <div class="flex items-center space-x-1.5">
            <span class="w-2 h-2 rounded-full" :class="isTelemetryStale ? 'bg-rose-500 animate-pulse' : 'bg-emerald-500 animate-ping'"></span>
            <span class="font-bold text-slate-800 dark:text-slate-200">{{ $t('operator.summaryTelemetry') || 'Data Pipeline' }}</span>
          </div>
          <span
            class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold"
            :class="isTelemetryStale ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/30' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30'"
          >
            {{ isTelemetryStale ? ($t('operator.stale') || 'STALE') : ($t('operator.live') || 'LIVE < 1s') }}
          </span>
        </div>
        <div class="flex items-baseline justify-between mt-1">
          <div class="flex items-baseline space-x-1.5">
            <span class="text-xl font-bold font-mono text-slate-900 dark:text-white">{{ formatRelativeSeconds(lastTelemetrySecondsAgo) }}</span>
            <span class="text-3xs text-slate-500 dark:text-slate-400 font-mono">since last packet</span>
          </div>
          <div class="text-right font-mono text-3xs">
            <span class="text-emerald-600 dark:text-emerald-400 font-bold">{{ summaryStats.goodSensors }}</span>
            <span class="text-slate-500 dark:text-slate-400">/{{ allParameters.length }} Good</span>
          </div>
        </div>
        <div class="flex items-center justify-between text-4xs text-slate-500 dark:text-slate-400 font-mono mt-1 pt-1 border-t border-slate-100 dark:border-[#1a2d42]">
          <span>Last: <strong class="text-slate-800 dark:text-slate-200 font-mono">{{ lastTelemetryTimeStr || 'Just now' }}</strong></span>
          <span :class="summaryStats.staleSensors > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-slate-500 dark:text-slate-400'">
            {{ summaryStats.staleSensors }} timed out
          </span>
        </div>
      </div>

      <!-- 4. System Monitor -->
      <div class="scada-panel p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] relative overflow-hidden group hover:border-cyan-500/50 transition">
        <div class="flex items-center justify-between text-3xs font-mono text-slate-500 dark:text-slate-400 uppercase tracking-wider mb-1">
          <div class="flex items-center space-x-1.5">
            <span class="w-2 h-2 rounded-full bg-blue-500"></span>
            <span class="font-bold text-slate-800 dark:text-slate-200">{{ $t('operator.summarySystem') || 'System Health' }}</span>
          </div>
          <span class="text-emerald-600 dark:text-emerald-400 font-bold font-mono">{{ systemStatus.service_status || 'RUNNING' }}</span>
        </div>
        <div class="grid grid-cols-3 gap-1 mt-1 text-center font-mono">
          <!-- CPU -->
          <div class="bg-slate-50 dark:bg-[#101B2B] p-1 rounded border border-slate-200 dark:border-[#20354B]">
            <div class="text-4xs text-slate-500 dark:text-slate-400 uppercase">CPU</div>
            <div class="text-xs font-bold text-slate-900 dark:text-white mt-0.5">{{ Math.round(systemStatus.cpu_percent || 0) }}%</div>
            <div class="w-full bg-slate-200 dark:bg-slate-800 h-1 rounded-full mt-1 overflow-hidden">
              <div class="bg-cyan-500 h-full" :style="{ width: Math.min(systemStatus.cpu_percent || 0, 100) + '%' }"></div>
            </div>
          </div>
          <!-- RAM -->
          <div class="bg-slate-50 dark:bg-[#101B2B] p-1 rounded border border-slate-200 dark:border-[#20354B]">
            <div class="text-4xs text-slate-500 dark:text-slate-400 uppercase">RAM</div>
            <div class="text-xs font-bold text-slate-900 dark:text-white mt-0.5">{{ Math.round(systemStatus.ram_percent || 0) }}%</div>
            <div class="w-full bg-slate-200 dark:bg-slate-800 h-1 rounded-full mt-1 overflow-hidden">
              <div class="bg-indigo-500 h-full" :style="{ width: Math.min(systemStatus.ram_percent || 0, 100) + '%' }"></div>
            </div>
          </div>
          <!-- DISK -->
          <div class="bg-slate-50 dark:bg-[#101B2B] p-1 rounded border border-slate-200 dark:border-[#20354B]">
            <div class="text-4xs text-slate-500 dark:text-slate-400 uppercase">DISK</div>
            <div class="text-xs font-bold text-slate-900 dark:text-white mt-0.5">{{ Math.round(systemStatus.disk_percent || 0) }}%</div>
            <div class="w-full bg-slate-200 dark:bg-slate-800 h-1 rounded-full mt-1 overflow-hidden">
              <div class="bg-emerald-500 h-full" :style="{ width: Math.min(systemStatus.disk_percent || 0, 100) + '%' }"></div>
            </div>
          </div>
        </div>
        <div class="flex items-center justify-between text-4xs text-slate-500 dark:text-slate-400 font-mono mt-1 pt-1 border-t border-slate-100 dark:border-[#1a2d42]">
          <span>DB: <strong class="text-emerald-600 dark:text-emerald-400">{{ systemStatus.db_latency_ms != null ? systemStatus.db_latency_ms + 'ms' : '0.8ms' }}</strong></span>
          <span>WAL: <strong class="text-cyan-600 dark:text-cyan-400">{{ systemStatus.queue_status || 'HEALTHY' }}</strong></span>
        </div>
      </div>
    </div>

    <!-- MAIN TWO-COLUMN OPERATIONAL MONITORING AREA (12 Columns) -->
    <div class="grid grid-cols-1 xl:grid-cols-12 gap-2.5">
      <!-- ======================================================== -->
      <!-- LEFT COLUMN (7 COLS): DEVICE & PARAMETER MONITORING       -->
      <!-- ======================================================== -->
      <div class="xl:col-span-7 flex flex-col space-y-2">
        <!-- Panel Header & Filter Controls Bar -->
        <div class="p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] shadow-xs dark:shadow-md flex flex-wrap items-center justify-between gap-2">
          <!-- Left: Title & Mode Toggle -->
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-500"></span>
            <h2 class="text-xs font-bold tracking-wider uppercase text-slate-900 dark:text-white font-mono">
              {{ $t('operator.parameterMonitoring') || 'DEVICE & PARAMETER MONITORING' }}
            </h2>
            <span class="text-4xs font-mono px-1.5 py-0.5 rounded bg-slate-100 dark:bg-[#101B2B] text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-[#20354B]">
              {{ filteredParameters.length }} / {{ allParameters.length }}
            </span>
          </div>

          <!-- View Mode & Quick Filter Toggles -->
          <div class="flex items-center space-x-1.5">
            <!-- Pinned Only Toggle -->
            <button
              @click="toggleFilterPinned"
              class="px-2 py-1 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1"
              :class="filterPinnedOnly
                ? 'bg-amber-500/20 text-amber-600 dark:text-amber-300 border-amber-500/40'
                : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B] hover:text-slate-900 dark:hover:text-slate-200'"
              :title="$t('operator.pinnedOnly') || 'Show Pinned Favorites Only'"
            >
              <span>⭐</span>
              <span class="hidden sm:inline">{{ $t('operator.pinnedOnly') || 'Pinned' }}</span>
            </button>

            <!-- Cards vs Table Switcher -->
            <div class="flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">
              <button
                @click="viewMode = 'cards'"
                class="px-2 py-1 rounded text-3xs font-mono font-semibold transition flex items-center space-x-1"
                :class="viewMode === 'cards'
                  ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border border-cyan-500/40 shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
                :title="$t('operator.viewCards') || 'SCADA Card Tiles'"
              >
                <span>▦</span>
                <span class="hidden sm:inline">{{ $t('operator.viewCards') || 'Cards' }}</span>
              </button>
              <button
                @click="viewMode = 'table'"
                class="px-2 py-1 rounded text-3xs font-mono font-semibold transition flex items-center space-x-1"
                :class="viewMode === 'table'
                  ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border border-cyan-500/40 shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
                :title="$t('operator.viewTable') || 'Live Telemetry Table'"
              >
                <span>☰</span>
                <span class="hidden sm:inline">{{ $t('operator.viewTable') || 'Table' }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Filter Dropdowns Row -->
        <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 p-2 rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] text-2xs font-mono">
          <!-- Search Input -->
          <div class="col-span-2 sm:col-span-1">
            <input
              type="text"
              v-model="searchQuery"
              @input="onSearchDebounced"
              :placeholder="$t('operator.searchPlaceholder') || 'Search params, dev...'"
              class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-white text-3xs placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-cyan-500"
            />
          </div>

          <!-- Device Filter -->
          <div>
            <select
              v-model="selectedDeviceFilter"
              class="w-full px-2 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-800 dark:text-slate-300 text-3xs focus:outline-none focus:border-cyan-500"
            >
              <option value="">{{ $t('operator.filterDevice') || 'All Devices' }}</option>
              <option v-for="dev in devicesList" :key="dev.id" :value="dev.id">
                {{ dev.device_code || dev.code }} ({{ dev.device_name || dev.name }})
              </option>
            </select>
          </div>

          <!-- Quality Filter -->
          <div>
            <select
              v-model="selectedQualityFilter"
              class="w-full px-2 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-800 dark:text-slate-300 text-3xs focus:outline-none focus:border-cyan-500"
            >
              <option value="">{{ $t('operator.filterQuality') || 'All Qualities' }}</option>
              <option value="GOOD">GOOD (Healthy)</option>
              <option value="UNCERTAIN">UNCERTAIN</option>
              <option value="BAD">BAD (Anomaly)</option>
              <option value="STALE">STALE (Timeout)</option>
            </select>
          </div>

          <!-- Alarm Status Filter -->
          <div>
            <select
              v-model="selectedAlarmFilter"
              class="w-full px-2 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-800 dark:text-slate-300 text-3xs focus:outline-none focus:border-cyan-500"
            >
              <option value="">{{ $t('operator.filterAlarm') || 'All Alarms' }}</option>
              <option value="CRITICAL">Critical Alarms</option>
              <option value="WARNING">Warning Alarms</option>
              <option value="NORMAL">Normal Operating</option>
            </select>
          </div>
        </div>

        <!-- VIEW MODE 1: SCADA PARAMETER CARD TILES (High Density) -->
        <div
          v-if="viewMode === 'cards'"
          class="grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-2 overflow-y-auto max-h-[500px] pr-1 custom-scroller"
        >
          <div
            v-for="param in filteredParameters"
            :key="param.param_id"
            @click="selectParamForTrend(param)"
            class="scada-param-tile p-2.5 rounded-xl border transition relative cursor-pointer group shadow-xs dark:shadow-none"
            :class="[
              isParamAlarmOrStaleBad(param)
                ? 'bg-slate-200/90 dark:bg-slate-800/90 border-2 border-rose-500 animate-alarm-pulse'
                : isParamSelectedForTrend(param)
                  ? 'border-cyan-500 dark:border-cyan-400 ring-1 ring-cyan-500/40 bg-cyan-50/40 dark:bg-[#0e1b2e]'
                  : 'bg-white dark:bg-[#0B1422] border-slate-200 dark:border-[#20354B] hover:border-cyan-400 dark:hover:border-slate-600',
              liveBlinks[param.param_id] ? 'bg-cyan-50 dark:bg-cyan-950/40' : ''
            ]"
          >
            <!-- Tile Header -->
            <div class="flex items-center justify-between text-4xs font-mono text-slate-500 dark:text-slate-400 mb-1">
              <div class="flex items-center space-x-1.5 truncate">
                <span class="w-1.5 h-1.5 rounded-full" :class="param.device_status === 'CONNECTED' ? 'bg-emerald-500' : 'bg-rose-500'"></span>
                <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ param.device_code }}</span>
                <span>:</span>
                <span class="text-slate-700 dark:text-slate-300 truncate font-semibold" :title="param.param_name">{{ param.param_code }}</span>
              </div>
              <!-- Pin Favorite Button -->
              <button
                @click.stop="togglePin(param.param_id)"
                class="text-xs transition hover:scale-110 p-0.5"
                :class="isPinned(param.param_id) ? 'text-amber-500 dark:text-amber-400' : 'text-slate-300 dark:text-slate-600 hover:text-amber-400'"
                :title="isPinned(param.param_id) ? 'Unpin' : 'Pin to top'"
              >
                ★
              </button>
            </div>

            <!-- Value Display & Unit -->
            <div class="flex items-baseline justify-between mt-1">
              <div class="flex items-baseline space-x-1">
                <span
                  class="text-2xl font-bold font-mono tracking-tight"
                  :class="paramValueColorClass(param)"
                >
                  {{ formatParamValue(param) }}
                </span>
                <span class="text-3xs font-mono text-slate-500 dark:text-slate-400 font-semibold">{{ param.unit || '' }}</span>
              </div>

              <!-- Quality Badge -->
              <span
                class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold uppercase"
                :class="qualityBadgeClass(param.quality)"
              >
                {{ param.quality || 'GOOD' }}
              </span>
            </div>

            <!-- Sub indicators: Thresholds & Held/Formula -->
            <div class="flex items-center justify-between text-4xs font-mono text-slate-500 dark:text-slate-400 mt-1">
              <span class="truncate">
                <span v-if="param.low_limit != null || param.high_limit != null" class="text-slate-600 dark:text-slate-400">
                  Lim: [{{ param.low_limit != null ? param.low_limit : '--' }} .. {{ param.high_limit != null ? param.high_limit : '--' }}]
                </span>
                <span v-else class="text-slate-400 dark:text-slate-600">Standard range</span>
              </span>

              <div class="flex items-center space-x-1">
                <span v-if="param.is_held_value" class="px-1 rounded bg-amber-500/10 dark:bg-amber-500/20 text-amber-600 dark:text-amber-300 font-bold" title="Hold last value active">
                  HELD
                </span>
                <span v-if="param.formula" class="px-1 rounded bg-blue-500/10 dark:bg-blue-500/20 text-blue-600 dark:text-blue-300 font-bold" :title="'Formula: ' + param.formula">
                  FX
                </span>
              </div>
            </div>

            <!-- Sparkline Trend Chart (SVG) -->
            <div class="mt-1.5 h-7 w-full bg-slate-50 dark:bg-[#080E18] border border-slate-100 dark:border-transparent rounded-md overflow-hidden relative">
              <svg class="w-full h-full" viewBox="0 0 100 24" preserveAspectRatio="none">
                <defs>
                  <linearGradient :id="'sparkGrad-' + param.param_id" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" :stop-color="getSparklineColor(param)" stop-opacity="0.3" />
                    <stop offset="100%" :stop-color="getSparklineColor(param)" stop-opacity="0.0" />
                  </linearGradient>
                </defs>
                <polygon
                  :points="getSparklinePolygon(param.param_id)"
                  :fill="'url(#sparkGrad-' + param.param_id + ')'"
                />
                <polyline
                  :points="getSparklinePoints(param.param_id)"
                  fill="none"
                  :stroke="getSparklineColor(param)"
                  stroke-width="1.8"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
            </div>

            <!-- Tile Footer: Timestamp & Quick Action Link -->
            <div class="flex items-center justify-between text-4xs font-mono text-slate-400 dark:text-slate-500 mt-1">
              <span>{{ param.received_at ? formatTimeShort(param.received_at) : 'Waiting data' }}</span>
              <span class="text-cyan-600 dark:text-cyan-400 group-hover:underline">
                {{ isParamSelectedForTrend(param) ? '● Selected in Chart' : '+ Plot Trend' }}
              </span>
            </div>
          </div>

          <div v-if="filteredParameters.length === 0" class="col-span-full p-8 text-center text-slate-500 font-mono text-xs">
            No parameters matching current filter criteria.
          </div>
        </div>

        <!-- VIEW MODE 2: COMPACT LIVE TELEMETRY TABLE -->
        <div v-else class="scada-panel rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] overflow-hidden shadow-xs dark:shadow-lg">
          <div class="overflow-x-auto max-h-[500px] custom-scroller">
            <table class="w-full text-left text-xs font-mono border-collapse">
              <thead class="sticky top-0 bg-slate-50 dark:bg-[#080E18] text-3xs font-bold text-slate-600 dark:text-slate-400 uppercase tracking-wider border-b border-slate-200 dark:border-[#20354B] z-10">
                <tr>
                  <th class="py-2 px-2.5 text-center w-8">⭐</th>
                  <th class="py-2 px-2.5">{{ $t('telemetry.device') || 'Device' }}</th>
                  <th class="py-2 px-2.5">{{ $t('telemetry.parameterName') || 'Parameter' }}</th>
                  <th class="py-2 px-2.5 text-right">{{ $t('common.value') || 'Value' }}</th>
                  <th class="py-2 px-2">{{ $t('common.unit') || 'Unit' }}</th>
                  <th class="py-2 px-2 text-center">{{ $t('common.quality') || 'Quality' }}</th>
                  <th class="py-2 px-2.5 text-center">Status</th>
                  <th class="py-2 px-2.5 text-right">Last Received</th>
                  <th class="py-2 px-2.5 text-right">Actions</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-[#18283E] text-2xs">
                <tr
                  v-for="p in filteredParameters"
                  :key="p.param_id"
                  class="transition-colors"
                  :class="[
                    isParamAlarmOrStaleBad(p)
                      ? 'bg-slate-200/70 dark:bg-slate-800/80 border-l-4 border-l-rose-500'
                      : 'hover:bg-slate-50 dark:hover:bg-[#101B2B]',
                    liveBlinks[p.param_id] ? 'bg-cyan-50 dark:bg-cyan-950/40' : '',
                    isPinned(p.param_id) ? 'bg-amber-50/40 dark:bg-[#0f1d30]/40' : ''
                  ]"
                >
                  <!-- Pin -->
                  <td class="py-2 px-2.5 text-center">
                    <button @click="togglePin(p.param_id)" class="text-xs" :class="isPinned(p.param_id) ? 'text-amber-500 dark:text-amber-400' : 'text-slate-300 dark:text-slate-600 hover:text-amber-400'">
                      ★
                    </button>
                  </td>

                  <!-- Device -->
                  <td class="py-2 px-2.5 text-cyan-600 dark:text-cyan-400 font-bold whitespace-nowrap">
                    <router-link :to="'/monitoring/devices/' + p.device_id" class="hover:underline">
                      {{ p.device_code }}
                    </router-link>
                  </td>

                  <!-- Parameter Name & Code -->
                  <td class="py-2 px-2.5 text-slate-700 dark:text-slate-200">
                    <div class="font-bold text-slate-900 dark:text-white leading-tight">{{ p.param_code }}</div>
                    <div class="text-4xs text-slate-500 dark:text-slate-400 font-sans truncate max-w-[130px]">{{ p.param_name }}</div>
                  </td>

                  <!-- Value -->
                  <td class="py-2 px-2.5 text-right font-bold text-sm whitespace-nowrap" :class="paramValueColorClass(p)">
                    <span>{{ formatParamValue(p) }}</span>
                    <span v-if="p.is_held_value" class="text-4xs text-amber-500 dark:text-amber-400 ml-1">H</span>
                  </td>

                  <!-- Unit -->
                  <td class="py-2 px-2 text-slate-500 dark:text-slate-400 font-semibold text-3xs">{{ p.unit || '--' }}</td>

                  <!-- Quality -->
                  <td class="py-2 px-2 text-center">
                    <span class="px-1.5 py-0.5 rounded text-4xs font-bold uppercase inline-block" :class="qualityBadgeClass(p.quality)">
                      {{ p.quality || 'GOOD' }}
                    </span>
                  </td>

                  <!-- Threshold / Limit Status -->
                  <td class="py-2 px-2.5 text-center">
                    <span
                      class="px-1.5 py-0.5 rounded text-4xs font-bold"
                      :class="p.alarm_status === 'CRITICAL' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400' : p.alarm_status === 'WARNING' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'"
                    >
                      {{ p.alarm_status || 'NORMAL' }}
                    </span>
                  </td>

                  <!-- Last Received -->
                  <td class="py-2 px-2.5 text-right text-slate-500 dark:text-slate-400 text-3xs whitespace-nowrap">
                    <div>{{ p.received_at ? formatTimeShort(p.received_at) : '--' }}</div>
                    <div class="text-4xs text-slate-400 dark:text-slate-500">{{ p.received_at ? formatRelative(p.received_at) : '' }}</div>
                  </td>

                  <!-- Actions -->
                  <td class="py-2 px-2.5 text-right whitespace-nowrap space-x-1">
                    <button
                      @click="selectParamForTrend(p)"
                      class="px-1.5 py-0.5 rounded text-4xs font-bold transition"
                      :class="isParamSelectedForTrend(p) ? 'bg-cyan-500 text-slate-950 font-black' : 'bg-slate-100 hover:bg-slate-200 dark:bg-[#18283E] dark:hover:bg-[#20354B] text-cyan-700 dark:text-cyan-400'"
                    >
                      {{ isParamSelectedForTrend(p) ? 'PLOTTED' : 'PLOT' }}
                    </button>
                    <router-link
                      :to="'/monitoring/analysis?device_id=' + p.device_id + '&parameter_id=' + p.param_id"
                      class="px-1.5 py-0.5 rounded text-4xs bg-slate-100 hover:bg-slate-200 dark:bg-[#18283E] dark:hover:bg-[#20354B] text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white"
                      title="Inspect in Analysis"
                    >
                      FX
                    </router-link>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- ======================================================== -->
      <!-- RIGHT COLUMN (5 COLS): TREND ANALYSIS & RECENT EVENTS     -->
      <!-- ======================================================== -->
      <div class="xl:col-span-5 flex flex-col space-y-2">
        <!-- 1. TREND ANALYSIS CHART PANEL -->
        <div class="scada-panel p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] shadow-xs dark:shadow-md flex flex-col space-y-2">
          <!-- Chart Header & Timeframe Buttons -->
          <div class="flex flex-wrap items-center justify-between gap-1.5 pb-1.5 border-b border-slate-200 dark:border-[#20354B]">
            <div class="flex items-center space-x-2">
              <span class="w-2 h-2 rounded-full bg-cyan-500"></span>
              <h3 class="text-xs font-bold tracking-wider uppercase text-slate-900 dark:text-white font-mono">
                {{ $t('operator.trendAnalysis') || 'TREND ANALYSIS' }}
              </h3>
            </div>

            <!-- Timeframe Preset Buttons -->
            <div class="flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-3xs font-mono">
              <button
                v-for="tf in ['15m', '1h', '6h', '24h', '7d']"
                :key="tf"
                @click="setTimeframe(tf)"
                class="px-2 py-0.5 rounded transition font-semibold"
                :class="currentTimeframe === tf
                  ? 'bg-cyan-500 text-slate-950 font-black shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
              >
                {{ tf }}
              </button>
            </div>
          </div>

          <!-- Series Selector Chips Bar -->
          <div class="flex flex-wrap items-center gap-1.5 text-3xs font-mono">
            <span class="text-slate-400 dark:text-slate-500 text-4xs uppercase">Series:</span>
            <div
              v-for="s in activeTrendSeries"
              :key="s.id"
              class="flex items-center space-x-1 px-2 py-0.5 rounded-md border text-3xs"
              :style="{ borderColor: s.color, backgroundColor: s.color + '15' }"
            >
              <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: s.color }"></span>
              <span class="font-bold text-slate-900 dark:text-white">{{ s.deviceCode }}:{{ s.paramCode }}</span>
              <span v-if="s.latestVal != null" class="font-bold" :style="{ color: s.color }">
                {{ s.latestVal }} {{ s.unit }}
              </span>
              <button @click="removeTrendSeries(s.id)" class="text-slate-400 hover:text-slate-700 dark:hover:text-white ml-0.5">✕</button>
            </div>

            <button
              v-if="activeTrendSeries.length === 0"
              @click="addDefaultTrendSeries"
              class="text-cyan-600 dark:text-cyan-400 text-3xs hover:underline"
            >
              + Click any parameter card to plot
            </button>
          </div>

          <!-- SVG Industrial Chart Canvas -->
          <div
            ref="scadaChartContainer"
            class="w-full h-44 bg-slate-50 dark:bg-[#070D18] rounded-lg border border-slate-200 dark:border-[#20354B] relative overflow-hidden select-none"
            @mousemove="onChartScrub"
            @mouseleave="scrubPoint = null"
          >
            <!-- Loading Indicator -->
            <div v-if="isChartLoading" class="absolute inset-0 bg-white/80 dark:bg-[#070D18]/80 flex items-center justify-center text-cyan-600 dark:text-cyan-400 text-xs font-mono z-20">
              <span class="animate-spin w-4 h-4 border-2 border-cyan-500 border-t-transparent rounded-full mr-2"></span>
              Loading Telemetry Trend...
            </div>

            <!-- Empty State -->
            <div v-else-if="activeTrendSeries.length === 0" class="w-full h-full flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 text-3xs font-mono space-y-1">
              <span>Select parameters from left monitoring tiles to display trend graph</span>
            </div>

            <!-- SVG Multi-Series Trend Line -->
            <svg v-else class="w-full h-full" viewBox="0 0 460 170" preserveAspectRatio="none">
              <!-- Grid Lines (dynamic stroke for light / dark contrast) -->
              <g class="grid-lines" :stroke="gridStroke" stroke-dasharray="3 3" stroke-width="1">
                <line x1="35" y1="20" x2="450" y2="20" />
                <line x1="35" y1="55" x2="450" y2="55" />
                <line x1="35" y1="90" x2="450" y2="90" />
                <line x1="35" y1="125" x2="450" y2="125" />
                <line x1="35" y1="150" x2="450" y2="150" stroke-dasharray="none" :stroke="axisStroke" />
              </g>

              <!-- Threshold Reference Lines (e.g. Critical High / Warning High) -->
              <line x1="35" y1="35" x2="450" y2="35" stroke="#FF4559" stroke-width="1" stroke-dasharray="4 4" opacity="0.7" />
              <text x="445" y="32" fill="#FF4559" font-size="8" text-anchor="end" font-family="monospace">CRIT LIMIT</text>

              <line x1="35" y1="65" x2="450" y2="65" stroke="#FFB020" stroke-width="1" stroke-dasharray="4 4" opacity="0.7" />
              <text x="445" y="62" fill="#FFB020" font-size="8" text-anchor="end" font-family="monospace">WARN LIMIT</text>

              <!-- Series Paths -->
              <g v-for="s in activeTrendSeries" :key="'plot-' + s.id">
                <!-- Area gradient -->
                <polygon
                  :points="generateChartPolygon(s)"
                  :fill="s.color"
                  fill-opacity="0.14"
                />
                <!-- Path line -->
                <polyline
                  :points="generateChartPolyline(s)"
                  fill="none"
                  :stroke="s.color"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </g>

              <!-- Scrubbing Hover Crosshair -->
              <g v-if="scrubPoint">
                <line :x1="scrubPoint.x" y1="15" :x2="scrubPoint.x" y2="150" stroke="#00CFFF" stroke-width="1.2" stroke-dasharray="2 2" />
                <circle :cx="scrubPoint.x" :cy="scrubPoint.y" r="3.5" fill="#00CFFF" stroke="#ffffff" stroke-width="1.5" />
              </g>
            </svg>

            <!-- Floating Hover Tooltip -->
            <div
              v-if="scrubPoint"
              class="absolute z-30 pointer-events-none bg-white dark:bg-[#0B1422] border border-cyan-500 rounded px-2 py-1 shadow-xl text-4xs font-mono text-slate-900 dark:text-white"
              :style="{ left: Math.min(Math.max(scrubPoint.x - 40, 10), 340) + 'px', top: '10px' }"
            >
              <div class="text-slate-500 dark:text-slate-400">{{ scrubPoint.time }}</div>
              <div class="font-bold text-cyan-600 dark:text-cyan-400">{{ scrubPoint.label }}: {{ scrubPoint.val }} {{ scrubPoint.unit }}</div>
            </div>
          </div>

          <!-- Chart Footer Controls -->
          <div class="flex items-center justify-between text-4xs font-mono text-slate-500 dark:text-slate-400">
            <div class="flex items-center space-x-2">
              <span>Mode:</span>
              <button
                @click="chartMode = 'processed'"
                :class="chartMode === 'processed' ? 'text-cyan-600 dark:text-cyan-400 font-bold' : 'text-slate-400 dark:text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'"
              >
                PROCESSED
              </button>
              <span>·</span>
              <button
                @click="chartMode = 'raw'"
                :class="chartMode === 'raw' ? 'text-cyan-600 dark:text-cyan-400 font-bold' : 'text-slate-400 dark:text-slate-500 hover:text-slate-700 dark:hover:text-slate-300'"
              >
                RAW
              </button>
            </div>
            <router-link to="/monitoring/analysis" class="text-cyan-600 dark:text-cyan-400 hover:underline">
              Deep Telemetry Analysis →
            </router-link>
          </div>
        </div>

        <!-- 2. RECENT ALARMS, NOTIFICATIONS & EVENTS PANEL -->
        <div class="scada-panel p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] shadow-xs dark:shadow-md flex-1 flex flex-col min-h-[280px]">
          <!-- Panel Segmented Tabs -->
          <div class="flex items-center justify-between pb-2 border-b border-slate-200 dark:border-[#20354B]">
            <div class="flex items-center space-x-1 bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-3xs font-mono">
              <!-- Tab 1: Active Alarms -->
              <button
                @click="activeRecentTab = 'alarms'"
                class="px-2.5 py-1 rounded transition font-semibold flex items-center space-x-1.5"
                :class="activeRecentTab === 'alarms'
                  ? 'bg-cyan-500 text-slate-950 font-black shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
              >
                <span>{{ $t('operator.tabActiveAlarms') || 'Alarms' }}</span>
                <span v-if="activeAlarmsList.length > 0" class="px-1 py-0.2 rounded-full text-4xs font-bold" :class="activeRecentTab === 'alarms' ? 'bg-slate-900 text-cyan-300' : 'bg-rose-500 text-white'">
                  {{ activeAlarmsList.length }}
                </span>
              </button>

              <!-- Tab 2: Recent Events -->
              <button
                @click="activeRecentTab = 'events'"
                class="px-2.5 py-1 rounded transition font-semibold flex items-center space-x-1"
                :class="activeRecentTab === 'events'
                  ? 'bg-cyan-500 text-slate-950 font-black shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
              >
                <span>{{ $t('operator.tabRecentEvents') || 'Events' }}</span>
              </button>

              <!-- Tab 3: Notification Delivery Queue -->
              <button
                @click="activeRecentTab = 'notifications'"
                class="px-2.5 py-1 rounded transition font-semibold flex items-center space-x-1"
                :class="activeRecentTab === 'notifications'
                  ? 'bg-cyan-500 text-slate-950 font-black shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
              >
                <span>{{ $t('operator.tabNotifications') || 'Delivery Queue' }}</span>
                <span v-if="notificationQueueStats.pending > 0" class="px-1 py-0.2 rounded-full bg-amber-500 text-slate-950 text-4xs font-bold">
                  {{ notificationQueueStats.pending }}
                </span>
              </button>
            </div>

            <!-- Tab Quick Actions -->
            <div class="text-3xs font-mono text-slate-500 dark:text-slate-400">
              <span v-if="activeRecentTab === 'alarms'">
                <button @click="fetchActiveAlarms" class="hover:text-cyan-600 dark:hover:text-cyan-400">↻ Sync</button>
              </span>
              <span v-else-if="activeRecentTab === 'events'">
                <router-link to="/system/audit" class="hover:text-cyan-600 dark:hover:text-cyan-400">All Audits →</router-link>
              </span>
              <span v-else>
                <router-link to="/alarm/queue" class="hover:text-cyan-600 dark:hover:text-cyan-400">Full Queue →</router-link>
              </span>
            </div>
          </div>

          <!-- TAB CONTENT 1: ACTIVE ALARMS LIST -->
          <div v-if="activeRecentTab === 'alarms'" class="flex-1 overflow-y-auto max-h-56 divide-y divide-slate-100 dark:divide-[#18283E] text-2xs font-mono custom-scroller mt-1">
            <div
              v-for="a in activeAlarmsList"
              :key="a.id"
              class="py-2 px-1.5 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-[#101B2B] transition rounded-lg"
            >
              <div class="flex items-center space-x-2 truncate flex-1 min-w-0 pr-2">
                <span class="text-xs">{{ a.severity === 'CRITICAL' ? '🚨' : '⚠️' }}</span>
                <span
                  class="px-1.5 py-0.2 rounded text-4xs font-bold uppercase flex-shrink-0"
                  :class="a.severity === 'CRITICAL' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/30' : 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/30'"
                >
                  {{ a.severity }}
                </span>
                <span class="font-bold text-slate-900 dark:text-white flex-shrink-0">[{{ a.device ? a.device.device_code : 'SYS' }}]</span>
                <span class="truncate text-slate-700 dark:text-slate-300 font-sans text-xs" :title="a.message">{{ a.message || a.alarm_code }}</span>
                <span v-if="a.trigger_value != null" class="text-rose-600 dark:text-rose-400 font-bold flex-shrink-0">
                  Val: {{ Number(a.trigger_value).toFixed(1) }}
                </span>
              </div>
              <div class="flex items-center space-x-1.5 flex-shrink-0">
                <span class="text-4xs text-slate-400 dark:text-slate-500">{{ formatRelative(a.triggered_at) }}</span>
                <button
                  v-if="a.status === 'ACTIVE'"
                  @click="openAckModal(a)"
                  class="px-2 py-0.5 rounded bg-blue-600 hover:bg-blue-500 text-white text-3xs font-bold font-mono transition"
                >
                  {{ $t('operator.acknowledge') || 'ACK' }}
                </button>
                <span v-else class="text-4xs text-emerald-600 dark:text-emerald-400 font-bold">ACKED</span>
              </div>
            </div>

            <div v-if="activeAlarmsList.length === 0" class="py-8 text-center text-slate-400 dark:text-slate-500 text-3xs font-mono">
              ✓ {{ $t('operator.noActiveAlarms') || 'No active alarms. Fleet functioning within safe limits.' }}
            </div>
          </div>

          <!-- TAB CONTENT 2: RECENT AUDIT & SYSTEM EVENTS -->
          <div v-else-if="activeRecentTab === 'events'" class="flex-1 overflow-y-auto max-h-56 divide-y divide-slate-100 dark:divide-[#18283E] text-2xs font-mono custom-scroller mt-1">
            <div
              v-for="e in recentEvents"
              :key="e.id"
              class="py-1.5 px-1 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-[#101B2B] transition"
            >
              <div class="flex items-center space-x-2 truncate flex-1 min-w-0 pr-2">
                <span class="w-1.5 h-1.5 rounded-full bg-cyan-500"></span>
                <span class="text-cyan-600 dark:text-cyan-400 font-bold flex-shrink-0">{{ e.action || 'EVENT' }}</span>
                <span class="truncate text-slate-700 dark:text-slate-300 font-sans text-xs" :title="e.details">{{ e.details || e.description }}</span>
              </div>
              <div class="text-4xs text-slate-400 dark:text-slate-500 flex-shrink-0 font-mono">
                {{ formatRelative(e.created_at || e.timestamp) }}
              </div>
            </div>

            <div v-if="recentEvents.length === 0" class="py-8 text-center text-slate-400 dark:text-slate-500 text-3xs font-mono">
              {{ $t('operator.noEvents') || 'No recent events recorded.' }}
            </div>
          </div>

          <!-- TAB CONTENT 3: NOTIFICATION DELIVERY QUEUE -->
          <div v-else class="flex-1 overflow-y-auto max-h-56 divide-y divide-slate-100 dark:divide-[#18283E] text-2xs font-mono custom-scroller mt-1">
            <div
              v-for="n in notificationQueue"
              :key="n.id"
              class="py-1.5 px-1 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-[#101B2B] transition"
            >
              <div class="flex items-center space-x-2 truncate flex-1 min-w-0 pr-2">
                <span class="text-xs">{{ n.channel === 'EMAIL' ? '✉️' : n.channel === 'TELEGRAM' ? '✈️' : '🔗' }}</span>
                <span
                  class="px-1.5 py-0.2 rounded text-4xs font-bold uppercase flex-shrink-0"
                  :class="getNotifStatusClass(n.status)"
                >
                  {{ n.status }}
                </span>
                <span class="font-mono text-slate-700 dark:text-slate-300">{{ n.recipient }}</span>
                <span class="truncate text-slate-500 dark:text-slate-400 text-4xs font-sans">{{ n.subject || n.message }}</span>
              </div>
              <div class="flex items-center space-x-1.5 flex-shrink-0">
                <span class="text-4xs text-slate-400 dark:text-slate-500">Tries: {{ n.attempts }}/{{ n.max_retries || 3 }}</span>
                <button
                  v-if="n.status === 'FAILED'"
                  @click="retryNotification(n.id)"
                  class="px-1.5 py-0.5 rounded bg-amber-600 hover:bg-amber-500 text-white text-4xs font-bold"
                >
                  {{ $t('operator.retryDelivery') || 'Retry' }}
                </button>
              </div>
            </div>

            <div v-if="notificationQueue.length === 0" class="py-8 text-center text-slate-400 dark:text-slate-500 text-3xs font-mono">
              ✓ {{ $t('operator.noNotifications') || 'Notification queue is clear.' }}
            </div>
          </div>

          <!-- Panel Bottom Status: Channels Health -->
          <div class="pt-2 mt-auto border-t border-slate-200 dark:border-[#20354B] flex items-center justify-between text-4xs font-mono text-slate-500 dark:text-slate-400">
            <span class="text-slate-400 dark:text-slate-500">{{ $t('operator.channelsHealth') || 'Channels' }}:</span>
            <div class="flex items-center space-x-2">
              <span class="flex items-center space-x-1">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                <span>Email</span>
              </span>
              <span class="flex items-center space-x-1">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                <span>Telegram</span>
              </span>
              <span class="flex items-center space-x-1 text-slate-400 dark:text-slate-500">
                <span class="w-1.5 h-1.5 rounded-full bg-slate-300 dark:bg-slate-600"></span>
                <span>Webhook ({{ $t('operator.notConfigured') || 'NOT CONFIGURED' }})</span>
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- I. QUICK ACTIONS & EXPORT BAR (Docked SCADA Footer) -->
    <div class="fixed bottom-0 left-0 right-0 z-30 bg-white/95 dark:bg-[#0B1422]/95 backdrop-blur-md border-t border-slate-200 dark:border-[#20354B] px-4 py-2 flex flex-wrap items-center justify-between gap-2 shadow-2xl transition-colors">
      <!-- Left: Quick Operational Commands -->
      <div class="flex items-center space-x-2">
        <span class="text-3xs uppercase font-mono font-bold text-cyan-600 dark:text-cyan-400 hidden sm:inline">
          {{ $t('operator.quickActions') || 'QUICK ACTIONS' }}:
        </span>

        <!-- Export Data -->
        <button
          @click="openExportModal"
          class="px-3 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-3xs font-mono transition-colors shadow-glow-blue flex items-center space-x-1.5"
        >
          <svg class="w-3.5 h-3.5 text-slate-950" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
          </svg>
          <span>{{ $t('operator.exportData') || 'Export Data' }}</span>
        </button>

        <!-- Generate Report -->
        <button
          @click="openReportModal"
          class="px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-bold text-3xs font-mono transition-colors flex items-center space-x-1.5 shadow-sm"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
          </svg>
          <span>{{ $t('operator.generateReport') || 'Operational Report' }}</span>
        </button>

        <!-- Alarm History -->
        <router-link
          to="/alarm/history"
          class="px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-[#20354B] text-3xs font-mono font-semibold transition hidden md:flex items-center space-x-1"
        >
          <span>📋</span>
          <span>{{ $t('operator.alarmHistory') || 'Alarm History' }}</span>
        </router-link>
      </div>

      <!-- Right: Navigation Shortcuts & Audio State -->
      <div class="flex items-center space-x-2 text-3xs font-mono">
        <router-link
          to="/monitoring/devices"
          class="px-2 py-1 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-[#20354B] transition hidden lg:inline"
        >
          {{ $t('operator.deviceMonitoring') || 'Devices' }}
        </router-link>
        <router-link
          to="/monitoring/analysis"
          class="px-2 py-1 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-[#20354B] transition hidden lg:inline"
        >
          {{ $t('operator.telemetryAnalysis') || 'Analysis' }}
        </router-link>
        <router-link
          to="/monitoring/system"
          class="px-2 py-1 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-[#20354B] transition hidden lg:inline"
        >
          {{ $t('operator.systemHealthAction') || 'System Health' }}
        </router-link>

        <!-- Audio Indicator Badge -->
        <span
          class="px-2 py-1 rounded border text-4xs font-bold"
          :class="isAudioMuted ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30' : 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border-cyan-500/30'"
        >
          {{ isAudioMuted ? '🔇 MUTED' : '🔊 SOUND ON' }}
        </span>
      </div>
    </div>

    <!-- MODAL 1: OPERATOR ACKNOWLEDGE MODAL -->
    <div
      v-if="showAckModal && targetAckAlarm"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-slate-950/80 backdrop-blur-sm"
    >
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] p-5 shadow-2xl text-slate-900 dark:text-slate-100 space-y-4 font-mono text-xs">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-[#20354B] pb-3">
          <div class="flex items-center space-x-2">
            <span class="text-base">🛡️</span>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              {{ $t('operator.ackModalTitle') || 'Acknowledge Alarm Incident' }}
            </h3>
          </div>
          <button @click="showAckModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">✕</button>
        </div>

        <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] space-y-1 text-2xs">
          <div class="flex items-center justify-between">
            <span class="font-bold text-slate-900 dark:text-white">[{{ targetAckAlarm.device ? targetAckAlarm.device.device_code : 'SYS' }}]</span>
            <span class="font-bold" :class="targetAckAlarm.severity === 'CRITICAL' ? 'text-rose-600 dark:text-rose-400' : 'text-amber-600 dark:text-amber-400'">
              {{ targetAckAlarm.severity }}
            </span>
          </div>
          <div class="text-slate-700 dark:text-slate-300 font-sans text-xs">
            {{ targetAckAlarm.message || targetAckAlarm.alarm_code }}
          </div>
          <div v-if="targetAckAlarm.trigger_value != null" class="text-4xs text-slate-500 dark:text-slate-400">
            Trigger value: <strong class="text-rose-600 dark:text-rose-400">{{ targetAckAlarm.trigger_value }}</strong>
          </div>
        </div>

        <div class="space-y-1.5 font-sans">
          <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
            {{ $t('operator.operatorNotes') || 'Operator Notes (Optional)' }}
          </label>
          <textarea
            v-model="ackNotes"
            rows="3"
            class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-[#20354B] bg-slate-50 dark:bg-[#101B2B] text-slate-900 dark:text-slate-200 text-xs focus:ring-1 focus:ring-cyan-500 focus:outline-none"
            :placeholder="$t('operator.notesPlaceholder') || 'Action taken, field observation...'"
          ></textarea>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
          <button
            @click="showAckModal = false"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="confirmAckAlarm"
            :disabled="isAckSubmitting"
            class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50"
          >
            {{ isAckSubmitting ? 'Acknowledging...' : ($t('operator.confirmAck') || 'Confirm Acknowledgement') }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 2: EXPORT TELEMETRY MODAL -->
    <div
      v-if="showExportModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-slate-950/80 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] p-5 shadow-2xl text-slate-900 dark:text-slate-100 space-y-4 font-mono text-xs">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-[#20354B] pb-3">
          <div class="flex items-center space-x-2">
            <span class="text-base text-cyan-500">📦</span>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              {{ $t('operator.exportModalTitle') || 'Export Operational Telemetry' }}
            </h3>
          </div>
          <button @click="showExportModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">✕</button>
        </div>

        <!-- Export Controls Form -->
        <div class="space-y-3 font-sans text-xs">
          <!-- Time Range -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
              {{ $t('operator.selectDateRange') || 'Time Range' }}
            </label>
            <div class="grid grid-cols-2 gap-2">
              <input
                type="datetime-local"
                v-model="exportForm.startTime"
                class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 font-mono text-2xs focus:outline-none focus:border-cyan-500"
              />
              <input
                type="datetime-local"
                v-model="exportForm.endTime"
                class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 font-mono text-2xs focus:outline-none focus:border-cyan-500"
              />
            </div>
          </div>

          <!-- Data Source Mode -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
              {{ $t('operator.dataSource') || 'Data Source Mode' }}
            </label>
            <div class="grid grid-cols-3 gap-2 font-mono text-2xs">
              <button
                @click="exportForm.mode = 'PROCESSED'"
                class="px-2 py-1.5 rounded-lg border font-semibold transition"
                :class="exportForm.mode === 'PROCESSED'
                  ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border-cyan-500/50'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B]'"
              >
                PROCESSED
              </button>
              <button
                @click="exportForm.mode = 'RAW'"
                class="px-2 py-1.5 rounded-lg border font-semibold transition"
                :class="exportForm.mode === 'RAW'
                  ? 'bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 border-emerald-500/50'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B]'"
              >
                RAW SENSOR
              </button>
              <button
                @click="exportForm.mode = 'AGGREGATED'"
                class="px-2 py-1.5 rounded-lg border font-semibold transition"
                :class="exportForm.mode === 'AGGREGATED'
                  ? 'bg-indigo-500/20 text-indigo-700 dark:text-indigo-300 border-indigo-500/50'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B]'"
              >
                AGGREGATED
              </button>
            </div>
          </div>

          <!-- Target Device -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
              {{ $t('operator.selectDevices') || 'Device Selection' }}
            </label>
            <select
              v-model="exportForm.deviceId"
              class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 text-xs focus:outline-none focus:border-cyan-500 font-mono"
            >
              <option value="">{{ $t('operator.allDevicesSelected') || 'All Fleet Devices (Unified)' }}</option>
              <option v-for="dev in devicesList" :key="'exp-dev-' + dev.id" :value="dev.id">
                {{ dev.device_code || dev.code }} — {{ dev.device_name || dev.name }}
              </option>
            </select>
          </div>

          <!-- Format Selection -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
              {{ $t('operator.selectFormat') || 'Export Format' }}
            </label>
            <div class="flex items-center space-x-3 font-mono text-2xs">
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input type="radio" value="csv" v-model="exportForm.format" class="text-cyan-500" />
                <span class="text-slate-700 dark:text-slate-300">CSV Spreadsheet</span>
              </label>
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input type="radio" value="json" v-model="exportForm.format" class="text-cyan-500" />
                <span class="text-slate-700 dark:text-slate-300">JSON Records</span>
              </label>
            </div>
          </div>
        </div>

        <!-- Export Progress / Feedback -->
        <div v-if="exportStatusMessage" class="p-2.5 rounded-lg bg-cyan-50 dark:bg-cyan-950/40 border border-cyan-200 dark:border-cyan-800/40 text-cyan-800 dark:text-cyan-300 text-2xs font-mono">
          {{ exportStatusMessage }}
        </div>

        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
          <button
            @click="showExportModal = false"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="executeTelemetryExport"
            :disabled="isExporting"
            class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50 flex items-center space-x-1.5"
          >
            <span v-if="isExporting" class="animate-spin w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full"></span>
            <span>{{ isExporting ? 'Generating...' : ($t('operator.downloadExport') || 'Download Export') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 3: OPERATIONAL REPORT MODAL -->
    <div
      v-if="showReportModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-slate-950/80 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] p-5 shadow-2xl text-slate-900 dark:text-slate-100 space-y-4 font-mono text-xs">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-[#20354B] pb-3">
          <div class="flex items-center space-x-2">
            <span class="text-base text-blue-500">📊</span>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              {{ $t('operator.reportModalTitle') || 'Generate Operational Report' }}
            </h3>
          </div>
          <button @click="showReportModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">✕</button>
        </div>

        <div class="space-y-3 font-sans text-xs">
          <!-- Report Type -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">Report Focus</label>
            <select
              v-model="reportForm.type"
              class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 text-xs focus:outline-none focus:border-cyan-500 font-mono"
            >
              <option value="COMPREHENSIVE">Comprehensive Operational (Fleet, Quality, Alarms)</option>
              <option value="TELEMETRY">Telemetry Quality & Data Gaps Focus</option>
              <option value="ALARMS">Alarm Incidents & Audit History</option>
              <option value="COMMUNICATION">Hardware Comm Availability</option>
            </select>
          </div>

          <!-- Preset Interval -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">Time Interval</label>
            <div class="grid grid-cols-4 gap-2 font-mono text-2xs">
              <button
                v-for="p in ['1h', '24h', '7d', '30d']"
                :key="'rep-' + p"
                @click="reportForm.preset = p"
                class="px-2 py-1.5 rounded-lg border font-semibold transition"
                :class="reportForm.preset === p ? 'bg-blue-600 text-white border-blue-500' : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B]'"
              >
                {{ p }}
              </button>
            </div>
          </div>

          <!-- Format Choice -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">Output Format</label>
            <div class="flex items-center space-x-3 font-mono text-2xs">
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input type="radio" value="html" v-model="reportForm.format" class="text-blue-500" />
                <span class="text-slate-700 dark:text-slate-300">Printable HTML / PDF</span>
              </label>
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input type="radio" value="csv" v-model="reportForm.format" class="text-blue-500" />
                <span class="text-slate-700 dark:text-slate-300">CSV Data Table</span>
              </label>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
          <button
            @click="showReportModal = false"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="executeReportGenerate"
            class="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-bold text-xs shadow-md transition"
          >
            Generate & Download
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import audioAlert from '../utils/audioAlert';

const TREND_PALETTE = ['#00CFFF', '#168BFF', '#00D6A0', '#FFB020', '#A855F7', '#EC4899'];

export default {
  name: 'OperatorCommandCenterView',
  props: {
    isSidebarCollapsed: {
      type: Boolean,
      default: false,
    },
    isTopPanelHidden: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      devicesList: [],
      liveMap: {},
      liveBlinks: {},
      sparklinesMap: {},
      pinnedIds: JSON.parse(localStorage.getItem('dl_pinned_params') || '[]'),
      searchQuery: '',
      searchDebounceTimer: null,
      selectedDeviceFilter: '',
      selectedQualityFilter: '',
      selectedAlarmFilter: '',
      filterPinnedOnly: false,
      viewMode: 'cards', // 'cards' | 'table'

      // Trend Chart State
      currentTimeframe: '24h',
      chartMode: 'processed', // 'processed' | 'raw'
      activeTrendSeries: [],
      isChartLoading: false,
      scrubPoint: null,

      // Recent Activity & Queue
      activeRecentTab: 'alarms', // 'alarms' | 'events' | 'notifications'
      recentEvents: [],
      notificationQueue: [],
      notificationQueueStats: { total: 0, pending: 0, sent: 0, failed: 0 },

      // Telemetry Freshness Ticker & Live Clock
      lastTelemetryTime: new Date(),
      lastTelemetrySecondsAgo: 0,
      currentTimeMs: Date.now(),
      tickerInterval: null,
      pollingInterval: null,
      clockInterval: null,
      currentTime: new Date().toLocaleTimeString('en-GB'),
      isRefreshing: false,
      telemetryRate: 24.5,

      // Modals
      showAckModal: false,
      targetAckAlarm: null,
      ackNotes: '',
      isAckSubmitting: false,

      showExportModal: false,
      exportForm: {
        startTime: '',
        endTime: '',
        mode: 'PROCESSED',
        deviceId: '',
        format: 'csv',
      },
      isExporting: false,
      exportStatusMessage: '',

      showReportModal: false,
      reportForm: {
        type: 'COMPREHENSIVE',
        preset: '24h',
        format: 'html',
      },
    };
  },
  computed: {
    isDark() {
      const pref = this.$store.state.uiPreferences;
      return pref ? pref.effectiveTheme === 'dark' : true;
    },
    currentLocale() {
      return this.$store.state.language || 'en';
    },
    gridStroke() {
      return this.isDark ? '#132337' : '#e2e8f0';
    },
    axisStroke() {
      return this.isDark ? '#20354B' : '#cbd5e1';
    },
    systemStatus() {
      return this.$store.state.systemStatus || {};
    },
    activeAlarmsList() {
      return (this.$store.state.activeAlarms || []).filter(a => a && a.status !== 'CLEARED');
    },
    activeCriticalCount() {
      return this.activeAlarmsList.filter(a => a.severity === 'CRITICAL' && a.status === 'ACTIVE').length;
    },
    activeWarningCount() {
      return this.activeAlarmsList.filter(a => a.severity === 'WARNING' && a.status === 'ACTIVE').length;
    },
    activeAckedCount() {
      return this.activeAlarmsList.filter(a => a.status === 'ACKNOWLEDGED').length;
    },
    highestPriorityAlarm() {
      if (this.activeAlarmsList.length === 0) return null;
      const sorted = [...this.activeAlarmsList].sort((a, b) => {
        if (a.severity === 'CRITICAL' && b.severity !== 'CRITICAL') return -1;
        if (a.severity !== 'CRITICAL' && b.severity === 'CRITICAL') return 1;
        return 0;
      });
      return sorted[0];
    },
    isAudioMuted() {
      const prefs = this.$store.state.alarmPreferences;
      return prefs ? prefs.isMuted : audioAlert.isMuted;
    },
    overallAppHealth() {
      if (this.systemStatus.database_status && String(this.systemStatus.database_status).toLowerCase().includes('fail')) {
        return 'FAILURE';
      }
      if (this.summaryStats.offlineCount > 0 || this.activeCriticalCount > 0) {
        return 'DEGRADED';
      }
      return 'HEALTHY';
    },
    appHealthColorClass() {
      if (this.overallAppHealth === 'FAILURE') return 'text-rose-600 dark:text-rose-400';
      if (this.overallAppHealth === 'DEGRADED') return 'text-amber-600 dark:text-amber-400';
      return 'text-emerald-600 dark:text-emerald-400';
    },
    appHealthDotClass() {
      if (this.overallAppHealth === 'FAILURE') return 'bg-rose-500 animate-ping';
      if (this.overallAppHealth === 'DEGRADED') return 'bg-amber-500 animate-pulse';
      return 'bg-emerald-500';
    },
    allParameters() {
      const list = [];
      const nowMs = this.currentTimeMs || Date.now();
      for (const dev of this.devicesList) {
        if (!dev.parameters) continue;
        const isDevOffline = dev.connection_status && (
          dev.connection_status === 'OFFLINE' ||
          dev.connection_status === 'DISCONNECTED' ||
          dev.connection_status === 'ERROR' ||
          dev.connection_status === 'FAILED'
        );
        for (const p of dev.parameters) {
          const live = this.liveMap[p.id] || {};
          const isHeld = live.is_held_value !== undefined ? live.is_held_value : !!p.is_current_held;
          const val = live.value !== undefined ? live.value : p.current_value;
          const alarmState = this.determineAlarmState(p, val);

          const receivedAt = live.received_at || p.current_received_at || p.last_updated;
          const staleSec = p.stale_timeout_seconds || 120;
          let isTimeStale = false;
          if (receivedAt) {
            const ageSec = (nowMs - new Date(receivedAt).getTime()) / 1000;
            if (ageSec > staleSec) {
              isTimeStale = true;
            }
          }

          let rawQuality = live.quality || p.current_quality || 'UNKNOWN';
          let finalQuality = rawQuality;
          // If the device is offline/disconnected: all parameters have communication failure -> BAD!
          // If the device is online, but data has not arrived past stale timeout -> STALE!
          if (isDevOffline) {
            finalQuality = 'BAD';
          } else if (isTimeStale) {
            if (finalQuality !== 'BAD') {
              finalQuality = 'STALE';
            }
          }

          const finalReason = isDevOffline
            ? (live.quality_reason || 'COMMUNICATION_ERROR')
            : (isTimeStale && (!live.quality_reason && !p.current_quality_reason))
              ? 'STALE_TIMEOUT'
              : (live.quality_reason || p.current_quality_reason || '');

          const finalTime = isDevOffline
            ? (live.received_at || this.lastTelemetryTime || receivedAt)
            : receivedAt;

          list.push({
            device_id: dev.id,
            device_code: dev.device_code || dev.code,
            device_name: dev.device_name || dev.name,
            device_status: dev.connection_status,
            param_id: p.id,
            param_code: p.parameter_code || p.code,
            param_name: p.parameter_name || p.name,
            unit: p.unit,
            data_type: p.data_type,
            precision: p.precision !== undefined ? p.precision : 2,
            value: val,
            is_held_value: isHeld,
            formula: p.formula || '',
            quality: finalQuality,
            quality_reason: finalReason,
            low_limit: p.low_critical != null ? p.low_critical : (p.low_warning != null ? p.low_warning : null),
            high_limit: p.high_critical != null ? p.high_critical : (p.high_warning != null ? p.high_warning : null),
            alarm_status: alarmState,
            received_at: finalTime,
          });
        }
      }
      return list;
    },
    filteredParameters() {
      let list = this.allParameters.filter((p) => {
        if (this.selectedDeviceFilter && p.device_id !== Number(this.selectedDeviceFilter)) {
          return false;
        }
        if (this.selectedQualityFilter && p.quality !== this.selectedQualityFilter) {
          return false;
        }
        if (this.selectedAlarmFilter && p.alarm_status !== this.selectedAlarmFilter) {
          return false;
        }
        if (this.filterPinnedOnly && !this.isPinned(p.param_id)) {
          return false;
        }
        if (this.searchQuery) {
          const q = this.searchQuery.toLowerCase();
          const matchCode = (p.param_code || '').toLowerCase().includes(q);
          const matchName = (p.param_name || '').toLowerCase().includes(q);
          const matchDev = (p.device_code || '').toLowerCase().includes(q);
          if (!matchCode && !matchName && !matchDev) return false;
        }
        return true;
      });

      // Pinned parameters always bubble to the top
      list.sort((a, b) => {
        const aPinned = this.isPinned(a.param_id) ? 1 : 0;
        const bPinned = this.isPinned(b.param_id) ? 1 : 0;
        return bPinned - aPinned;
      });

      return list;
    },
    summaryStats() {
      const devs = this.devicesList || [];
      const totalDevs = devs.length;
      const onlineDevs = devs.filter(d => d.connection_status === 'CONNECTED' || d.is_active).length;
      const offlineDevs = totalDevs - onlineDevs;
      const disabledDevs = devs.filter(d => d.admin_status === 'DISABLED' || !d.is_active).length;

      const params = this.allParameters;
      const goodParams = params.filter(p => p.quality === 'GOOD').length;
      const staleParams = params.filter(p => p.quality === 'STALE').length;

      return {
        totalDevices: totalDevs,
        onlineCount: onlineDevs,
        offlineCount: offlineDevs,
        disabledCount: disabledDevs,
        goodSensors: goodParams,
        staleSensors: staleParams,
      };
    },
    isTelemetryStale() {
      return this.lastTelemetrySecondsAgo > 30 || this.summaryStats.staleSensors > 0;
    },
    lastTelemetryTimeStr() {
      if (!this.lastTelemetryTime) return '';
      return new Date(this.lastTelemetryTime).toLocaleTimeString('en-GB');
    },
  },
  mounted() {
    this.initDashboard();

    // Telemetry Freshness Ticker
    this.tickerInterval = setInterval(() => {
      this.currentTimeMs = Date.now();
      if (this.lastTelemetryTime) {
        this.lastTelemetrySecondsAgo = Math.floor((Date.now() - new Date(this.lastTelemetryTime).getTime()) / 1000);
      }
    }, 1000);

    // Live Digital Clock
    this.clockInterval = setInterval(() => {
      this.currentTime = new Date().toLocaleTimeString('en-GB');
    }, 1000);

    // Background Poll Interval for Telemetry & Alarms
    this.pollingInterval = setInterval(() => {
      this.refreshLatestOnly();
      this.fetchActiveAlarms();
    }, 10000);

    // Listen to real WebSocket Telemetry events
    window.addEventListener('device-telemetry-event', this.onTelemetryEvent);
  },
  beforeDestroy() {
    if (this.tickerInterval) clearInterval(this.tickerInterval);
    if (this.clockInterval) clearInterval(this.clockInterval);
    if (this.pollingInterval) clearInterval(this.pollingInterval);
    window.removeEventListener('device-telemetry-event', this.onTelemetryEvent);
  },
  methods: {
    toggleTheme() {
      const nextTheme = this.isDark ? 'light' : 'dark';
      this.$store.dispatch('setTheme', nextTheme);
    },
    toggleLanguage() {
      const nextLang = this.currentLocale === 'en' ? 'id' : 'en';
      this.$store.dispatch('setLanguage', nextLang);
    },
    determineAlarmState(p, val) {
      if (val == null || isNaN(val)) return 'NORMAL';
      const num = Number(val);
      if (p.high_critical != null && num >= p.high_critical) return 'CRITICAL';
      if (p.low_critical != null && num <= p.low_critical) return 'CRITICAL';
      if (p.high_warning != null && num >= p.high_warning) return 'WARNING';
      if (p.low_warning != null && num <= p.low_warning) return 'WARNING';
      return 'NORMAL';
    },
    isParamAlarmOrStaleBad(param) {
      if (!param) return false;
      // 1. If parameter is GOOD and has no alarm, it MUST NOT have gray card or red border
      if (param.quality === 'GOOD' && (!param.alarm_status || param.alarm_status === 'NORMAL')) {
        return false;
      }
      // 2. If parameter itself has an active threshold alarm (CRITICAL / WARNING)
      if (param.alarm_status === 'CRITICAL' || param.alarm_status === 'WARNING') {
        return true;
      }
      // 3. If parameter data quality is STALE, BAD, or UNCERTAIN
      if (param.quality === 'STALE' || param.quality === 'BAD' || param.quality === 'UNCERTAIN') {
        return true;
      }
      // 4. If an active alarm specifically targets this parameter
      if (this.activeAlarmsList && this.activeAlarmsList.length > 0) {
        const hasSpecificParamAlarm = this.activeAlarmsList.some(a => {
          if (a.status !== 'ACTIVE') return false;
          if (a.parameter_id && a.parameter_id === param.param_id) return true;
          if (a.parameter && (a.parameter.parameter_code === param.param_code || a.parameter.code === param.param_code) && (a.device_id === param.device_id)) return true;
          return false;
        });
        if (hasSpecificParamAlarm) return true;
      }
      return false;
    },
    formatParamValue(param) {
      if (param.value == null || param.value === '') return '--';
      const num = Number(param.value);
      if (isNaN(num)) return String(param.value);
      return num.toFixed(param.precision != null ? param.precision : 2);
    },
    paramValueColorClass(param) {
      if (param.alarm_status === 'CRITICAL') return 'text-rose-600 dark:text-rose-400 font-black';
      if (param.alarm_status === 'WARNING') return 'text-amber-600 dark:text-amber-400 font-bold';
      return 'text-slate-900 dark:text-white font-bold';
    },
    qualityBadgeClass(q) {
      switch (q) {
        case 'GOOD':
          return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/30';
        case 'UNCERTAIN':
          return 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/30';
        case 'BAD':
          return 'bg-rose-500/10 text-rose-700 dark:text-rose-400 border border-rose-500/30';
        case 'STALE':
          return 'bg-purple-500/10 text-purple-700 dark:text-purple-400 border border-purple-500/30';
        default:
          return 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700';
      }
    },
    getNotifStatusClass(status) {
      switch (status) {
        case 'SENT':
        case 'DELIVERED':
          return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/30';
        case 'PENDING':
        case 'SENDING':
          return 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-400 border border-cyan-500/30';
        case 'FAILED':
          return 'bg-rose-500/10 text-rose-700 dark:text-rose-400 border border-rose-500/30';
        case 'RETRYING':
          return 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/30';
        default:
          return 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400';
      }
    },
    getSparklineColor(param) {
      if (param.alarm_status === 'CRITICAL' || param.quality === 'BAD') return '#FF4559';
      if (param.alarm_status === 'WARNING') return '#FFB020';
      if (param.quality === 'STALE') return '#A855F7';
      return this.isDark ? '#00CFFF' : '#0284c7';
    },
    recordSparkline(paramId, val) {
      if (val == null || isNaN(val)) return;
      if (!this.sparklinesMap[paramId]) {
        this.$set(this.sparklinesMap, paramId, []);
      }
      const arr = this.sparklinesMap[paramId];
      arr.push(Number(val));
      if (arr.length > 20) arr.shift();
    },
    getSparklinePoints(paramId) {
      const arr = this.sparklinesMap[paramId] || [];
      if (arr.length < 2) return '0,12 100,12';
      const min = Math.min(...arr);
      const max = Math.max(...arr);
      const span = max - min || 1;
      const step = 100 / (arr.length - 1);

      return arr.map((val, idx) => {
        const x = idx * step;
        const norm = (val - min) / span;
        const y = 22 - norm * 20;
        return `${x.toFixed(1)},${y.toFixed(1)}`;
      }).join(' ');
    },
    getSparklinePolygon(paramId) {
      const pts = this.getSparklinePoints(paramId);
      return `${pts} 100,24 0,24`;
    },
    onTelemetryEvent(e) {
      const data = e.detail;
      if (!data) return;
      this.lastTelemetryTime = new Date();
      this.lastTelemetrySecondsAgo = 0;

      if (data.parameter_id && data.value !== undefined) {
        this.$set(this.liveMap, data.parameter_id, data);
        this.recordSparkline(data.parameter_id, data.value);

        // Flash row/tile briefly
        this.$set(this.liveBlinks, data.parameter_id, true);
        setTimeout(() => {
          this.$set(this.liveBlinks, data.parameter_id, false);
        }, 600);
      }
    },
    async initDashboard() {
      this.isRefreshing = true;
      try {
        await Promise.all([
          this.fetchDevices(),
          this.$store.dispatch('fetchSystemStatus'),
          this.fetchActiveAlarms(),
          this.fetchEvents(),
          this.fetchNotifications(),
        ]);
        this.addDefaultTrendSeries();
      } catch (err) {
        console.error('Failed to initialize operator dashboard:', err);
      } finally {
        this.isRefreshing = false;
      }
    },
    async refreshDashboard() {
      this.isRefreshing = true;
      try {
        await Promise.all([
          this.fetchDevices(),
          this.$store.dispatch('fetchSystemStatus'),
          this.fetchActiveAlarms(),
          this.fetchEvents(),
          this.fetchNotifications(),
          this.fetchTrendData(),
        ]);
      } finally {
        this.isRefreshing = false;
      }
    },
    async refreshLatestOnly() {
      for (const dev of this.devicesList) {
        try {
          const res = await axios.get(`/api/devices/${dev.id}/telemetry/latest`);
          if (res.data && res.data.data) {
            for (const item of res.data.data) {
              this.$set(this.liveMap, item.parameter_id, item);
              this.recordSparkline(item.parameter_id, item.value);
            }
          }
        } catch (e) {
          // ignore error
        }
      }
    },
    async fetchDevices() {
      try {
        const res = await axios.get('/api/devices?page_size=100');
        if (res.data && res.data.data) {
          this.devicesList = res.data.data.items || res.data.data || [];
          await this.refreshLatestOnly();
        }
      } catch (err) {
        console.error('Failed to load device catalog:', err);
      }
    },
    async fetchActiveAlarms() {
      await this.$store.dispatch('fetchActiveAlarms');
    },
    async fetchEvents() {
      try {
        const res = await axios.get('/api/audit-trails?page=1&page_size=30');
        if (res.data && res.data.data) {
          this.recentEvents = res.data.data.items || res.data.data || [];
        }
      } catch (err) {
        console.error('Failed to load audit events:', err);
      }
    },
    async fetchNotifications() {
      try {
        const res = await axios.get('/api/notifications?page=1&page_size=30');
        if (res.data && res.data.data) {
          this.notificationQueue = res.data.data.items || res.data.data || [];
          this.notificationQueueStats.total = this.notificationQueue.length;
          this.notificationQueueStats.pending = this.notificationQueue.filter(n => n.status === 'PENDING' || n.status === 'RETRYING').length;
        }
      } catch (err) {
        console.error('Failed to load notification queue:', err);
      }
    },
    togglePin(paramId) {
      const idx = this.pinnedIds.indexOf(paramId);
      if (idx !== -1) {
        this.pinnedIds.splice(idx, 1);
      } else {
        this.pinnedIds.push(paramId);
      }
      localStorage.setItem('dl_pinned_params', JSON.stringify(this.pinnedIds));
    },
    isPinned(paramId) {
      return this.pinnedIds.includes(paramId);
    },
    toggleFilterPinned() {
      this.filterPinnedOnly = !this.filterPinnedOnly;
    },
    onSearchDebounced() {
      clearTimeout(this.searchDebounceTimer);
      this.searchDebounceTimer = setTimeout(() => {
        // search query updated reactively in computed
      }, 250);
    },
    toggleAudioMute() {
      const nextMuted = !this.isAudioMuted;
      audioAlert.setMuted(nextMuted);
      this.$store.commit('SET_ALARM_PREFERENCES', { isMuted: audioAlert.isMuted });
    },
    testAlarmChime(severity) {
      audioAlert.playChime(severity || 'CRITICAL');
    },

    // ----------------------------------------------------
    // TREND ANALYSIS TIME-SERIES MAPPING
    // ----------------------------------------------------
    selectParamForTrend(param) {
      const existingIdx = this.activeTrendSeries.findIndex(s => s.parameterId === param.param_id);
      if (existingIdx !== -1) {
        this.activeTrendSeries.splice(existingIdx, 1);
      } else {
        if (this.activeTrendSeries.length >= 4) {
          this.activeTrendSeries.shift();
        }
        const color = TREND_PALETTE[this.activeTrendSeries.length % TREND_PALETTE.length];
        this.activeTrendSeries.push({
          id: `${param.device_id}_${param.param_id}`,
          deviceId: param.device_id,
          parameterId: param.param_id,
          deviceCode: param.device_code,
          paramCode: param.param_code,
          unit: param.unit,
          latestVal: this.formatParamValue(param),
          color,
          dataPoints: [],
        });
        this.fetchTrendData();
      }
    },
    isParamSelectedForTrend(param) {
      return this.activeTrendSeries.some(s => s.parameterId === param.param_id);
    },
    removeTrendSeries(id) {
      this.activeTrendSeries = this.activeTrendSeries.filter(s => s.id !== id);
    },
    addDefaultTrendSeries() {
      if (this.activeTrendSeries.length > 0 || this.allParameters.length === 0) return;
      const firstTwo = this.allParameters.slice(0, 2);
      for (const p of firstTwo) {
        this.selectParamForTrend(p);
      }
    },
    setTimeframe(tf) {
      this.currentTimeframe = tf;
      this.fetchTrendData();
    },
    async fetchTrendData() {
      if (this.activeTrendSeries.length === 0) return;
      this.isChartLoading = true;
      try {
        const now = new Date();
        let startTime = new Date(now.getTime() - 24 * 3600 * 1000);
        if (this.currentTimeframe === '15m') startTime = new Date(now.getTime() - 15 * 60 * 1000);
        else if (this.currentTimeframe === '1h') startTime = new Date(now.getTime() - 3600 * 1000);
        else if (this.currentTimeframe === '6h') startTime = new Date(now.getTime() - 6 * 3600 * 1000);
        else if (this.currentTimeframe === '7d') startTime = new Date(now.getTime() - 7 * 86400 * 1000);

        for (const s of this.activeTrendSeries) {
          const endpoint = this.chartMode === 'raw'
            ? `/api/devices/${s.deviceId}/telemetry/raw`
            : `/api/devices/${s.deviceId}/telemetry/history`;

          const res = await axios.get(endpoint, {
            params: {
              parameter_id: s.parameterId,
              start_time: startTime.toISOString(),
              end_time: now.toISOString(),
              page_size: 60,
            },
          }).catch(() => null);

          if (res && res.data && res.data.data) {
            const items = res.data.data.items || [];
            s.dataPoints = items.map(item => ({
              x: new Date(item.received_at || item.timestamp).getTime(),
              y: Number(item.value) || 0,
            })).sort((a, b) => a.x - b.x);
          }
        }
      } finally {
        this.isChartLoading = false;
      }
    },
    generateChartPolyline(series) {
      const pts = series.dataPoints;
      if (!pts || pts.length === 0) return '';
      const minX = pts[0].x;
      const maxX = pts[pts.length - 1].x;
      const spanX = maxX - minX || 1;

      const yVals = pts.map(p => p.y);
      const minY = Math.min(...yVals);
      const maxY = Math.max(...yVals);
      const spanY = maxY - minY || 1;

      return pts.map(p => {
        const px = 35 + ((p.x - minX) / spanX) * (450 - 35);
        const py = 150 - ((p.y - minY) / spanY) * (150 - 20);
        return `${px.toFixed(1)},${py.toFixed(1)}`;
      }).join(' ');
    },
    generateChartPolygon(series) {
      const polyline = this.generateChartPolyline(series);
      if (!polyline) return '';
      return `${polyline} 450,150 35,150`;
    },
    onChartScrub(e) {
      if (!this.$refs.scadaChartContainer || this.activeTrendSeries.length === 0) return;
      const rect = this.$refs.scadaChartContainer.getBoundingClientRect();
      const mouseX = e.clientX - rect.left;
      const firstSeries = this.activeTrendSeries[0];
      if (!firstSeries || !firstSeries.dataPoints.length) return;

      const scaleX = 460 / rect.width;
      const chartX = mouseX * scaleX;
      this.scrubPoint = {
        x: Math.min(Math.max(chartX, 35), 450),
        y: 85,
        time: new Date().toLocaleTimeString('en-GB'),
        label: `${firstSeries.deviceCode}:${firstSeries.paramCode}`,
        val: firstSeries.latestVal || '--',
        unit: firstSeries.unit || '',
      };
    },

    // ----------------------------------------------------
    // ALARM ACKNOWLEDGEMENT
    // ----------------------------------------------------
    openAckModal(alarm) {
      this.targetAckAlarm = alarm;
      this.ackNotes = '';
      this.showAckModal = true;
    },
    async confirmAckAlarm() {
      if (!this.targetAckAlarm) return;
      this.isAckSubmitting = true;
      try {
        await this.$store.dispatch('acknowledgeAlarm', {
          id: this.targetAckAlarm.id,
          notes: this.ackNotes,
        });
        this.showAckModal = false;
        await this.fetchActiveAlarms();
      } catch (err) {
        alert('Failed to acknowledge alarm: ' + (err.message || 'Server error'));
      } finally {
        this.isAckSubmitting = false;
      }
    },
    async retryNotification(id) {
      try {
        await axios.post(`/api/notifications/${id}/retry`);
        await this.fetchNotifications();
      } catch (err) {
        console.error('Failed to retry notification:', err);
      }
    },

    // ----------------------------------------------------
    // EXPORT & REPORT WORKFLOWS
    // ----------------------------------------------------
    openExportModal() {
      const now = new Date();
      const past24h = new Date(now.getTime() - 24 * 3600 * 1000);
      this.exportForm.endTime = this.toLocalISO(now);
      this.exportForm.startTime = this.toLocalISO(past24h);
      this.exportStatusMessage = '';
      this.showExportModal = true;
    },
    async executeTelemetryExport() {
      this.isExporting = true;
      this.exportStatusMessage = 'Querying dataset from edge database...';
      try {
        const startISO = new Date(this.exportForm.startTime).toISOString();
        const endISO = new Date(this.exportForm.endTime).toISOString();

        let endpoint = '/api/telemetry/history';
        if (this.exportForm.mode === 'RAW') endpoint = '/api/telemetry/raw';
        else if (this.exportForm.mode === 'AGGREGATED') endpoint = '/api/aggregations/results';

        const params = {
          start_time: startISO,
          end_time: endISO,
          page_size: 5000,
        };
        if (this.exportForm.deviceId) params.device_id = this.exportForm.deviceId;

        const res = await axios.get(endpoint, { params });
        const items = res.data?.data?.items || res.data?.data || [];

        if (items.length === 0) {
          this.exportStatusMessage = 'No records found for specified interval.';
          return;
        }

        this.exportStatusMessage = `Preparing download of ${items.length} records...`;

        if (this.exportForm.format === 'json') {
          const blob = new Blob([JSON.stringify(items, null, 2)], { type: 'application/json' });
          this.triggerDownloadBlob(blob, `telemetry_export_${Date.now()}.json`);
        } else {
          // Generate CSV
          const headers = ['Timestamp', 'Device ID', 'Parameter ID', 'Value', 'Quality'];
          const rows = items.map(i => [
            i.received_at || i.timestamp || '',
            i.device_id || '',
            i.parameter_id || '',
            i.value != null ? i.value : '',
            i.quality || 'GOOD',
          ]);
          const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
          const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
          this.triggerDownloadBlob(blob, `telemetry_export_${Date.now()}.csv`);
        }

        setTimeout(() => {
          this.showExportModal = false;
        }, 1200);
      } catch (err) {
        this.exportStatusMessage = 'Export failed: ' + (err.message || 'API error');
      } finally {
        this.isExporting = false;
      }
    },
    openReportModal() {
      this.showReportModal = true;
    },
    executeReportGenerate() {
      const now = new Date();
      let startTime = new Date(now.getTime() - 24 * 3600 * 1000);
      if (this.reportForm.preset === '1h') startTime = new Date(now.getTime() - 3600 * 1000);
      else if (this.reportForm.preset === '7d') startTime = new Date(now.getTime() - 7 * 86400 * 1000);
      else if (this.reportForm.preset === '30d') startTime = new Date(now.getTime() - 30 * 86400 * 1000);

      const params = new URLSearchParams({
        start_time: startTime.toISOString(),
        end_time: now.toISOString(),
        report_type: this.reportForm.type,
        format: this.reportForm.format,
      });

      if (this.reportForm.format === 'html') {
        params.append('download', '1');
      }
      window.open(`/api/reports/generate?${params.toString()}`, '_blank');
      this.showReportModal = false;
    },
    triggerDownloadBlob(blob, filename) {
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    },
    toLocalISO(d) {
      const pad = n => (n < 10 ? '0' + n : n);
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    },
    formatTimeShort(ts) {
      if (!ts) return '';
      const d = new Date(ts);
      return d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    formatRelative(ts) {
      if (!ts) return '';
      const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
      if (diff < 5) return 'Just now';
      if (diff < 60) return `${diff}s ago`;
      if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
      return `${Math.floor(diff / 3600)}h ago`;
    },
    formatRelativeSeconds(sec) {
      if (sec < 5) return '< 1s';
      if (sec < 60) return `${sec}s`;
      return `${Math.floor(sec / 60)}m`;
    },
  },
};
</script>

<style scoped>
/* Industrial SCADA Styling Tokens */
.custom-scroller::-webkit-scrollbar {
  width: 5px;
  height: 5px;
}
.custom-scroller::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroller::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.4);
  border-radius: 4px;
}
.custom-scroller::-webkit-scrollbar-thumb:hover {
  background: #168BFF;
}

.scada-panel {
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}
:global(.dark) .scada-panel {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
}

/* Bright-dim breathing animation for alarmed and stale/bad parameter cards */
@keyframes alarm-pulse-glow {
  0%, 100% {
    border-color: rgba(239, 68, 68, 0.45);
    box-shadow: 0 0 4px rgba(239, 68, 68, 0.25);
    opacity: 0.88;
  }
  50% {
    border-color: rgba(239, 68, 68, 1);
    box-shadow: 0 0 16px rgba(239, 68, 68, 0.7), inset 0 0 8px rgba(239, 68, 68, 0.2);
    opacity: 1;
  }
}

.animate-alarm-pulse {
  animation: alarm-pulse-glow 1.8s ease-in-out infinite;
}

@media (prefers-reduced-motion: reduce) {
  .animate-alarm-pulse {
    animation: none;
    box-shadow: 0 0 8px rgba(239, 68, 68, 0.5);
    opacity: 1;
  }
}
</style>
