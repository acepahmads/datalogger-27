<template>
  <div class="space-y-6 font-sans text-slate-800 dark:text-slate-200 pb-12">
    <!-- Non-Superadmin Warning Fallback (if accessed directly) -->
    <div
      v-if="!isSuperAdmin"
      class="p-8 rounded-3xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 text-center max-w-xl mx-auto my-12"
    >
      <div class="w-16 h-16 rounded-2xl bg-rose-100 dark:bg-rose-900/40 text-rose-600 dark:text-rose-400 flex items-center justify-center mx-auto mb-4">
        <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
        </svg>
      </div>
      <h2 class="text-xl font-bold text-rose-900 dark:text-rose-200 mb-2">Akses Terbatas: Superadmin Saja</h2>
      <p class="text-xs text-rose-700 dark:text-rose-400 mb-6">
        Halaman System Activity Monitor hanya dapat diakses oleh akun dengan hak akses Superadmin.
      </p>
      <router-link
        to="/monitoring/operator"
        class="inline-flex items-center px-4 py-2 text-xs font-semibold rounded-xl bg-slate-900 text-white dark:bg-white dark:text-slate-900 hover:opacity-90 transition-opacity"
      >
        Kembali ke Dashboard
      </router-link>
    </div>

    <!-- MAIN DASHBOARD CONTENT (Superadmin Only) -->
    <template v-else>
      <!-- SECTION A: TOP BAR & SYSTEM OVERVIEW -->
      <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <div>
          <div class="flex items-center space-x-2 text-xs text-slate-400 dark:text-slate-500 mb-1">
            <router-link to="/alarm/active" class="hover:text-blue-500 transition-colors">
              {{ $t('navigation.alarmsEvents') || 'Alarms & Events' }}
            </router-link>
            <span>/</span>
            <span class="text-slate-700 dark:text-slate-300 font-semibold">System Activity Monitor</span>
          </div>
          <div class="flex items-center space-x-3">
            <h1 class="text-2xl font-bold tracking-tight text-slate-900 dark:text-white flex items-center gap-2.5">
              <span class="p-2.5 rounded-2xl bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20 shadow-2xs">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                </svg>
              </span>
              System Activity Monitor
            </h1>

            <span class="px-2.5 py-1 rounded-full text-2xs font-mono font-bold uppercase tracking-wider bg-indigo-100 text-indigo-700 border border-indigo-200 dark:bg-indigo-950/60 dark:text-indigo-300 dark:border-indigo-800">
              SUPERADMIN CONSOLE
            </span>

            <!-- WebSocket Live Stream Indicator -->
            <span
              class="px-2.5 py-1 rounded-full text-2xs font-mono font-semibold flex items-center space-x-1.5 border"
              :class="wsConnected
                ? 'bg-emerald-50 text-emerald-700 border-emerald-300 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800'
                : 'bg-amber-50 text-amber-700 border-amber-300 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800'"
            >
              <span class="w-2 h-2 rounded-full" :class="wsConnected ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500'"></span>
              <span>{{ wsConnected ? 'FEED STREAM: LIVE' : 'FEED: RECONNECTING' }}</span>
            </span>
          </div>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-1 max-w-3xl">
            Pusat observabilitas sistem: amati komunikasi sensor, polling engine, retry backoff, dan auto-recovery secara real-time tanpa terminal.
          </p>
        </div>

        <!-- Global Header Actions -->
        <div class="flex items-center space-x-2.5 flex-wrap">
          <!-- Live Feed Pause/Resume Button -->
          <button
            @click="toggleFeedPause"
            class="px-3.5 py-2 rounded-xl text-xs font-semibold transition-all border flex items-center space-x-2 shadow-2xs"
            :class="isFeedPaused
              ? 'bg-amber-500 hover:bg-amber-600 text-white border-amber-600'
              : 'bg-white hover:bg-slate-50 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 dark:border-slate-700'"
            :title="isFeedPaused ? 'Klik untuk melanjutkan streaming aktivitas' : 'Klik untuk jeda pembaruan tampilan aktivitas'"
          >
            <svg v-if="isFeedPaused" class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM9.555 7.168A1 1 0 008 8v4a1 1 0 001.555.832l3-2a1 1 0 000-1.664l-3-2z" clip-rule="evenodd" />
            </svg>
            <svg v-else class="w-3.5 h-3.5 text-amber-500" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zM7 8a1 1 0 012 0v4a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v4a1 1 0 102 0V8a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
            <span>{{ isFeedPaused ? 'Lanjutkan Feed' : 'Jeda Feed' }}</span>
          </button>

          <!-- Refresh Data -->
          <button
            @click="refreshAll(true)"
            :disabled="loading"
            class="px-3.5 py-2 rounded-xl text-xs font-semibold bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-slate-700 transition-all flex items-center space-x-1.5 shadow-2xs disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
            <span>Refresh</span>
          </button>

          <!-- Export Dropdown -->
          <div class="relative" ref="exportDropdownRef">
            <button
              @click="showExportMenu = !showExportMenu"
              class="px-3.5 py-2 rounded-xl text-xs font-semibold bg-indigo-600 hover:bg-indigo-700 text-white transition-all flex items-center space-x-1.5 shadow-xs"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
              </svg>
              <span>Export Log</span>
              <svg class="w-3 h-3 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
              </svg>
            </button>
            <div
              v-if="showExportMenu"
              class="absolute right-0 mt-2 w-44 rounded-xl bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 shadow-lg py-1 z-30 text-xs"
            >
              <button
                @click="exportLogs('csv')"
                class="w-full text-left px-4 py-2 hover:bg-slate-100 dark:hover:bg-slate-700/60 flex items-center space-x-2 text-slate-700 dark:text-slate-200"
              >
                <span class="font-mono text-indigo-500 font-bold">CSV</span>
                <span>Unduh CSV Spreadsheet</span>
              </button>
              <button
                @click="exportLogs('json')"
                class="w-full text-left px-4 py-2 hover:bg-slate-100 dark:hover:bg-slate-700/60 flex items-center space-x-2 text-slate-700 dark:text-slate-200"
              >
                <span class="font-mono text-emerald-500 font-bold">JSON</span>
                <span>Unduh JSON Data</span>
              </button>
            </div>
          </div>

          <!-- Clear In-Memory Ring Buffer -->
          <button
            @click="confirmClearLogs"
            class="px-3.5 py-2 rounded-xl text-xs font-semibold bg-rose-50 hover:bg-rose-100 text-rose-700 dark:bg-rose-950/40 dark:hover:bg-rose-900/50 dark:text-rose-300 border border-rose-200 dark:border-rose-900/60 transition-all flex items-center space-x-1.5 shadow-2xs"
            title="Bersihkan ring buffer memori log aktivitas"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            <span>Clear Buffer</span>
          </button>
        </div>
      </div>

      <!-- SECTION A.1: EXECUTIVE STATUS METRICS CARDS -->
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3.5">
        <!-- Overall System Status -->
        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-2xs relative overflow-hidden">
          <div class="flex items-center justify-between mb-2">
            <span class="text-3xs font-bold text-slate-400 uppercase tracking-wider">Overall Status</span>
            <span class="w-2.5 h-2.5 rounded-full" :class="statusBgColor(overview.overall_status)"></span>
          </div>
          <div class="text-base font-bold font-mono tracking-tight" :class="statusTextColor(overview.overall_status)">
            {{ overview.overall_status || 'UNKNOWN' }}
          </div>
          <div class="text-3xs text-slate-500 mt-1 flex items-center space-x-1">
            <span>Uptime:</span>
            <span class="font-mono font-medium">{{ formatUptime(overview.uptime_seconds) }}</span>
          </div>
        </div>

        <!-- Polling Engine Status -->
        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-2xs">
          <div class="flex items-center justify-between mb-2">
            <span class="text-3xs font-bold text-slate-400 uppercase tracking-wider">Polling Engine</span>
            <span class="w-2.5 h-2.5 rounded-full" :class="overview.polling_engine_status === 'RUNNING' ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500'"></span>
          </div>
          <div class="text-base font-bold font-mono tracking-tight text-slate-900 dark:text-white">
            {{ overview.polling_engine_status || 'RUNNING' }}
          </div>
          <div class="text-3xs text-slate-500 mt-1">
            {{ overview.active_polls_count || 0 }} worker aktif berjalan
          </div>
        </div>

        <!-- Devices Online -->
        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-2xs">
          <div class="flex items-center justify-between mb-2">
            <span class="text-3xs font-bold text-slate-400 uppercase tracking-wider">Devices Online</span>
            <span class="p-1 rounded bg-emerald-100 text-emerald-600 dark:bg-emerald-950/60 dark:text-emerald-400">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>
            </span>
          </div>
          <div class="text-2xl font-bold font-mono text-emerald-600 dark:text-emerald-400">
            {{ overview.devices_online || 0 }}
            <span class="text-xs text-slate-400 font-normal">/ {{ overview.total_devices || 0 }}</span>
          </div>
          <div class="text-3xs text-slate-500 mt-1">Komunikasi aktif terverifikasi</div>
        </div>

        <!-- Devices Reconnecting / Degraded -->
        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-2xs">
          <div class="flex items-center justify-between mb-2">
            <span class="text-3xs font-bold text-slate-400 uppercase tracking-wider">Reconnecting</span>
            <span class="p-1 rounded bg-amber-100 text-amber-600 dark:bg-amber-950/60 dark:text-amber-400">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
            </span>
          </div>
          <div class="text-2xl font-bold font-mono text-amber-600 dark:text-amber-400">
            {{ overview.devices_reconnecting || 0 }}
          </div>
          <div class="text-3xs text-slate-500 mt-1">Auto-recovery loop aktif</div>
        </div>

        <!-- Devices Offline -->
        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-2xs">
          <div class="flex items-center justify-between mb-2">
            <span class="text-3xs font-bold text-slate-400 uppercase tracking-wider">Devices Offline</span>
            <span class="p-1 rounded bg-rose-100 text-rose-600 dark:bg-rose-950/60 dark:text-rose-400">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
            </span>
          </div>
          <div class="text-2xl font-bold font-mono text-rose-600 dark:text-rose-400">
            {{ overview.devices_offline || 0 }}
          </div>
          <div class="text-3xs text-slate-500 mt-1">Terputus / port hilang</div>
        </div>

        <!-- Waiting / Timeout / Active Errors -->
        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-2xs">
          <div class="flex items-center justify-between mb-2">
            <span class="text-3xs font-bold text-slate-400 uppercase tracking-wider">Error & Timeout</span>
            <span class="p-1 rounded bg-indigo-100 text-indigo-600 dark:bg-indigo-950/60 dark:text-indigo-400">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
            </span>
          </div>
          <div class="text-2xl font-bold font-mono text-slate-800 dark:text-slate-100">
            {{ overview.active_errors_count || 0 }}
          </div>
          <div class="text-3xs text-slate-500 mt-1">
            Terakhir: {{ formatTimeOnly(overview.last_activity_time) }}
          </div>
        </div>
      </div>

      <!-- SECTION C: POLLING ACTIVITY MONITOR TABLE -->
      <div class="rounded-3xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-xs overflow-hidden">
        <div class="p-5 border-b border-slate-200 dark:border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <div class="flex items-center space-x-3">
            <span class="p-2 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
              </svg>
            </span>
            <div>
              <h2 class="text-base font-bold text-slate-900 dark:text-white flex items-center space-x-2">
                <span>Polling Activity Monitor</span>
                <span class="text-2xs font-mono font-medium px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                  {{ pollingSnapshots.length }} Perangkat
                </span>
              </h2>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                Status polling per perangkat secara real-time. Klik baris perangkat untuk melihat detail diagnostik komunikasi.
              </p>
            </div>
          </div>
        </div>

        <!-- Table View -->
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs whitespace-nowrap">
            <thead class="bg-slate-50 dark:bg-[#0A0E18] text-slate-500 dark:text-slate-400 font-semibold border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="py-3 px-4">Perangkat</th>
                <th class="py-3 px-3">Protokol / Port</th>
                <th class="py-3 px-3 text-center">Status Transport</th>
                <th class="py-3 px-3 text-center">Status Komunikasi</th>
                <th class="py-3 px-3 text-center">Polling State</th>
                <th class="py-3 px-3 text-right">Latency</th>
                <th class="py-3 px-3 text-center">Failures / Retry</th>
                <th class="py-3 px-3 text-center">Parameter (OK/Fail)</th>
                <th class="py-3 px-3">Respons Terakhir</th>
                <th class="py-3 px-3">Status Pemulihan</th>
                <th class="py-3 px-4 text-center">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
              <template v-for="dev in pollingSnapshots">
                <tr
                  :key="'dev-' + dev.device_id"
                  @click="toggleDeviceExpand(dev.device_id)"
                  class="cursor-pointer transition-colors hover:bg-blue-50/40 dark:hover:bg-blue-950/20"
                  :class="{ 'bg-blue-50/60 dark:bg-blue-950/30 font-medium': expandedDeviceId === dev.device_id }"
                >
                  <!-- Device Code, Name & Administrative Status -->
                  <td class="py-3.5 px-4">
                    <div class="flex items-center space-x-2.5">
                      <button class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">
                        <svg
                          class="w-4 h-4 transform transition-transform"
                          :class="{ 'rotate-90': expandedDeviceId === dev.device_id }"
                          fill="none" stroke="currentColor" viewBox="0 0 24 24"
                        >
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
                        </svg>
                      </button>
                      <div>
                        <div class="font-bold text-slate-900 dark:text-white flex items-center space-x-1.5">
                          <span>{{ dev.device_code }}</span>
                          <span class="text-3xs font-mono text-slate-400">#{{ dev.device_id }}</span>
                          <span
                            class="inline-flex items-center px-1.5 py-0.2 rounded text-4xs font-mono font-medium border"
                            :class="adminStatusBadge(dev.administrative_status)"
                            :title="'Administrative: ' + (dev.administrative_status || 'ACTIVE')"
                          >
                            {{ dev.administrative_status || 'ACTIVE' }}
                          </span>
                        </div>
                        <div class="text-3xs text-slate-500">{{ dev.device_name }}</div>
                      </div>
                    </div>
                  </td>

                  <!-- Protocol & Transport Path -->
                  <td class="py-3.5 px-3">
                    <div class="font-mono text-2xs font-semibold text-slate-700 dark:text-slate-300">
                      {{ dev.protocol || 'MODBUS' }}
                    </div>
                    <div class="text-3xs text-slate-400 font-mono">
                      {{ dev.transport || 'N/A' }}
                    </div>
                  </td>

                  <!-- Status Transport (AVAILABLE / PORT_MISSING / OPEN_FAILED / DISCONNECTED) -->
                  <td class="py-3.5 px-3 text-center">
                    <span
                      class="inline-flex items-center px-2 py-0.5 rounded-full text-3xs font-mono font-bold uppercase tracking-wider border shadow-2xs"
                      :class="transportStatusBadge(dev.transport_status)"
                      :title="'Transport: ' + (dev.transport_status || 'UNKNOWN')"
                    >
                      <span class="w-1.5 h-1.5 rounded-full mr-1.5" :class="transportStatusDot(dev.transport_status)"></span>
                      {{ dev.transport_status || 'UNKNOWN' }}
                    </span>
                  </td>

                  <!-- Status Komunikasi (ONLINE / OFFLINE / DEGRADED) + Freshness -->
                  <td class="py-3.5 px-3 text-center">
                    <div class="flex flex-col items-center space-y-0.5">
                      <span
                        class="inline-flex items-center px-2.5 py-0.5 rounded-full text-3xs font-mono font-bold uppercase tracking-wider border shadow-2xs"
                        :class="commStatusBadge(dev.communication_status)"
                        :title="'Communication: ' + (dev.communication_status || 'UNKNOWN')"
                      >
                        <span class="w-1.5 h-1.5 rounded-full mr-1.5" :class="commStatusDot(dev.communication_status)"></span>
                        {{ dev.communication_status || 'UNKNOWN' }}
                      </span>
                      <span
                        v-if="dev.communication_status === 'ONLINE' && dev.is_fresh === false"
                        class="inline-flex items-center px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300 border border-amber-300"
                        title="Data respons kedaluwarsa melampaui batas toleransi freshness"
                      >
                        STALE
                      </span>
                      <span
                        v-else-if="dev.communication_status === 'ONLINE' && dev.is_fresh"
                        class="text-4xs text-emerald-600 dark:text-emerald-400 font-mono font-medium"
                      >
                        ● Fresh
                      </span>
                    </div>
                  </td>

                  <!-- Polling State -->
                  <td class="py-3.5 px-3 text-center">
                    <span
                      class="inline-flex items-center px-2 py-0.5 rounded-md text-3xs font-mono font-semibold"
                      :class="pollingStateBadge(dev.polling_state)"
                    >
                      <span v-if="dev.polling_state === 'POLLING'" class="w-1.5 h-1.5 rounded-full bg-blue-500 animate-ping mr-1.5"></span>
                      {{ dev.polling_state || 'WAITING' }}
                    </span>
                  </td>

                  <!-- Latency -->
                  <td class="py-3.5 px-3 text-right font-mono font-semibold">
                    <span :class="latencyColor(dev.last_latency_ms)">
                      {{ dev.last_latency_ms ? dev.last_latency_ms + ' ms' : '-' }}
                    </span>
                  </td>

                  <!-- Failures & Retry -->
                  <td class="py-3.5 px-3 text-center font-mono">
                    <span v-if="dev.consecutive_failures > 0" class="text-rose-600 dark:text-rose-400 font-bold">
                      {{ dev.consecutive_failures }} fail
                    </span>
                    <span v-else class="text-slate-400">0</span>
                    <span class="text-slate-400 text-3xs ml-1">
                      ({{ dev.retry_count || 0 }} retries)
                    </span>
                  </td>

                  <!-- Parameters Summary -->
                  <td class="py-3.5 px-3 text-center">
                    <div v-if="dev.total_parameters > 0" class="flex items-center justify-center space-x-1 font-mono text-2xs">
                      <span class="text-emerald-600 dark:text-emerald-400 font-semibold">{{ dev.success_parameters }}</span>
                      <span class="text-slate-400">/</span>
                      <span v-if="dev.failed_parameters > 0" class="text-rose-600 font-bold">{{ dev.failed_parameters }} err</span>
                      <span v-else class="text-slate-400">0 err</span>
                      <span class="text-slate-400 text-3xs">({{ dev.total_parameters }} total)</span>
                    </div>
                    <span v-else class="text-3xs text-slate-400">-</span>
                  </td>

                  <!-- Last Successful Response -->
                  <td class="py-3.5 px-3 font-mono text-3xs text-slate-500">
                    <div>{{ formatTimestamp(dev.last_successful_response) }}</div>
                    <div v-if="dev.last_successful_response" class="text-slate-400">
                      {{ formatTimeAgo(dev.last_successful_response) }}
                    </div>
                  </td>

                  <!-- Recovery State & Error Detail -->
                  <td class="py-3.5 px-3">
                    <div class="flex items-center space-x-1.5">
                      <span
                        class="text-3xs font-mono font-semibold px-1.5 py-0.2 rounded"
                        :class="recoveryStateBadge(dev.recovery_state)"
                      >
                        {{ dev.recovery_state || 'NORMAL' }}
                      </span>
                    </div>
                    <div v-if="dev.last_error" class="text-3xs text-rose-500 truncate max-w-xs" :title="dev.last_error">
                      {{ dev.last_error }}
                    </div>
                  </td>

                  <!-- Detail Action -->
                  <td class="py-3.5 px-4 text-center">
                    <button
                      @click.stop="toggleDeviceExpand(dev.device_id)"
                      class="px-2.5 py-1 rounded-lg text-3xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition-colors"
                    >
                      {{ expandedDeviceId === dev.device_id ? 'Tutup' : 'Detail' }}
                    </button>
                  </td>
                </tr>

                <!-- SECTION D: INLINE EXPANDED DEVICE ACTIVITY DETAIL -->
                <tr v-if="expandedDeviceId === dev.device_id" :key="'detail-' + dev.device_id">
                  <td colspan="11" class="p-0 bg-slate-50/70 dark:bg-[#0A0E18]/80 border-y border-slate-200 dark:border-slate-800">
                    <div class="p-6 space-y-4">
                      <div class="flex flex-col md:flex-row md:items-center justify-between gap-3 border-b border-slate-200 dark:border-slate-800 pb-3">
                        <div class="flex items-center space-x-3">
                          <span class="p-2 rounded-xl bg-indigo-500/10 text-indigo-600 dark:text-indigo-400">
                            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>
                          </span>
                          <div>
                            <h3 class="text-sm font-bold text-slate-900 dark:text-white">
                              Diagnostik & Aktivitas Transaksi: {{ dev.device_code }} ({{ dev.device_name }})
                            </h3>
                            <p class="text-3xs text-slate-500">
                              Data register dan jejak transaksi komunikasi serial/TCP terkini tanpa menampilkan secret.
                            </p>
                          </div>
                        </div>

                        <div class="flex items-center space-x-2">
                          <button
                            @click="fetchDeviceDetail(dev.device_id)"
                            :disabled="detailLoading"
                            class="px-3 py-1.5 rounded-lg text-3xs font-semibold bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 flex items-center space-x-1"
                          >
                            <svg class="w-3 h-3" :class="{ 'animate-spin': detailLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/></svg>
                            <span>Refresh Detail</span>
                          </button>
                        </div>
                      </div>

                      <!-- Sub-grid: Connection specs + Parameters + Recent Frames -->
                      <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
                        <!-- Safe Connection Parameters -->
                        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-2xs space-y-2.5">
                          <h4 class="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center space-x-1.5">
                            <svg class="w-3.5 h-3.5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/></svg>
                            <span>Konfigurasi Link</span>
                          </h4>

                          <div v-if="deviceDetailData && deviceDetailData.connection" class="space-y-1.5 text-3xs font-mono">
                            <div class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Protokol:</span>
                              <span class="font-bold text-slate-800 dark:text-slate-200">{{ deviceDetailData.connection.protocol }}</span>
                            </div>
                            <div v-if="deviceDetailData.connection.serial_port" class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Serial Port:</span>
                              <span class="text-indigo-600 dark:text-indigo-400 font-bold">{{ deviceDetailData.connection.serial_port }}</span>
                            </div>
                            <div v-if="deviceDetailData.connection.baud_rate" class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Baud / Frame:</span>
                              <span>{{ deviceDetailData.connection.baud_rate }} bps ({{ deviceDetailData.connection.data_bits }}{{ deviceDetailData.connection.parity }}{{ deviceDetailData.connection.stop_bits }})</span>
                            </div>
                            <div v-if="deviceDetailData.connection.host" class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">TCP Endpoint:</span>
                              <span class="text-indigo-600 dark:text-indigo-400 font-bold">{{ deviceDetailData.connection.host }}:{{ deviceDetailData.connection.port }}</span>
                            </div>
                            <div class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Slave ID:</span>
                              <span>{{ deviceDetailData.connection.slave_id }}</span>
                            </div>
                            <div class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Timeout / Retries:</span>
                              <span>{{ deviceDetailData.connection.timeout_ms }}ms (max {{ deviceDetailData.connection.retry_count }}x)</span>
                            </div>
                            <div class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Polling Cadence:</span>
                              <span>{{ deviceDetailData.connection.polling_interval }}ms</span>
                            </div>
                            <div class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Batch Read:</span>
                              <span :class="deviceDetailData.connection.batch_read_enabled ? 'text-emerald-500 font-bold' : 'text-slate-400'">
                                {{ deviceDetailData.connection.batch_read_enabled ? 'ENABLED' : 'DISABLED' }}
                              </span>
                            </div>
                            <div class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Status Transport:</span>
                              <span class="font-bold font-mono text-3xs px-1.5 py-0.5 rounded" :class="transportStatusBadge((deviceDetailData.snapshot && deviceDetailData.snapshot.transport_status) || dev.transport_status)">
                                {{ (deviceDetailData.snapshot && deviceDetailData.snapshot.transport_status) || dev.transport_status || 'UNKNOWN' }}
                              </span>
                            </div>
                            <div class="flex justify-between py-1 border-b border-slate-100 dark:border-slate-800">
                              <span class="text-slate-400">Status Komunikasi:</span>
                              <span class="font-bold font-mono text-3xs px-1.5 py-0.5 rounded" :class="commStatusBadge((deviceDetailData.snapshot && deviceDetailData.snapshot.communication_status) || dev.communication_status)">
                                {{ (deviceDetailData.snapshot && deviceDetailData.snapshot.communication_status) || dev.communication_status || 'UNKNOWN' }}
                              </span>
                            </div>
                            <div class="flex justify-between py-1">
                              <span class="text-slate-400">Data Freshness:</span>
                              <span class="font-bold font-mono text-3xs" :class="((deviceDetailData.snapshot && deviceDetailData.snapshot.is_fresh) || dev.is_fresh) ? 'text-emerald-500' : 'text-amber-500'">
                                {{ ((deviceDetailData.snapshot && deviceDetailData.snapshot.is_fresh) || dev.is_fresh) ? 'VALID & FRESH' : 'STALE / EXPIRED' }}
                              </span>
                            </div>
                          </div>
                          <div v-else class="text-3xs text-slate-400 py-3 text-center">Memuat konfigurasi...</div>
                        </div>

                        <!-- Parameter List -->
                        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-2xs space-y-2.5">
                          <h4 class="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center space-x-1.5">
                            <svg class="w-3.5 h-3.5 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/></svg>
                            <span>Parameter Sensor</span>
                          </h4>

                          <div v-if="deviceDetailData && deviceDetailData.parameters && deviceDetailData.parameters.length" class="max-h-48 overflow-y-auto space-y-1 pr-1">
                            <div
                              v-for="param in deviceDetailData.parameters"
                              :key="param.id"
                              class="p-2 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800 flex items-center justify-between text-3xs"
                            >
                              <div>
                                <span class="font-bold font-mono text-slate-800 dark:text-slate-200">{{ param.parameter_code }}</span>
                                <span class="text-slate-400 ml-1">{{ param.parameter_name }}</span>
                              </div>
                              <div class="font-mono text-slate-500 flex items-center space-x-1">
                                <span>Reg: {{ param.register_address || param.address || '-' }}</span>
                                <span class="text-slate-400">({{ param.data_type }})</span>
                              </div>
                            </div>
                          </div>
                          <div v-else class="text-3xs text-slate-400 py-4 text-center">Tidak ada parameter terdaftar.</div>
                        </div>

                        <!-- Recent Frame Transactions -->
                        <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-2xs space-y-2.5">
                          <h4 class="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider flex items-center space-x-1.5">
                            <svg class="w-3.5 h-3.5 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
                            <span>Transaksi Terkini (TX/RX)</span>
                          </h4>

                          <div v-if="deviceDetailData && deviceDetailData.recent_frames && deviceDetailData.recent_frames.length" class="max-h-48 overflow-y-auto space-y-1.5 pr-1 font-mono text-3xs">
                            <div
                              v-for="(frame, fIdx) in deviceDetailData.recent_frames.slice(0, 5)"
                              :key="fIdx"
                              class="p-2 rounded-xl bg-slate-50 dark:bg-slate-800/50 border border-slate-100 dark:border-slate-800"
                            >
                              <div class="flex justify-between text-slate-400 mb-0.5">
                                <span>{{ formatTimestamp(frame.timestamp) }}</span>
                                <span :class="frame.success ? 'text-emerald-500 font-bold' : 'text-rose-500 font-bold'">
                                  {{ frame.success ? 'OK (' + frame.duration_ms + 'ms)' : 'FAILED' }}
                                </span>
                              </div>
                              <div v-if="frame.request_hex" class="text-slate-600 dark:text-slate-300 truncate">
                                <span class="text-blue-500 font-bold">TX:</span> {{ frame.request_hex }}
                              </div>
                              <div v-if="frame.response_hex" class="text-slate-600 dark:text-slate-300 truncate">
                                <span class="text-emerald-500 font-bold">RX:</span> {{ frame.response_hex }}
                              </div>
                              <div v-if="frame.error_message" class="text-rose-500 truncate">
                                {{ frame.error_message }}
                              </div>
                            </div>
                          </div>
                          <div v-else class="text-3xs text-slate-400 py-4 text-center">
                            Belum ada frame sniffer tersimpan. Aktifkan Live Sensor Comm untuk inspeksi paket raw.
                          </div>
                        </div>
                      </div>
                    </div>
                  </td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>

      <!-- SECTION B & F: LIVE ACTIVITY FEED WITH INTEGRATED CONTROLS -->
      <div class="rounded-3xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-xs overflow-hidden">
        <!-- Feed Header & Controls -->
        <div class="p-5 border-b border-slate-200 dark:border-slate-800 space-y-4">
          <div class="flex flex-col md:flex-row md:items-center justify-between gap-3">
            <div class="flex items-center space-x-3">
              <span class="p-2 rounded-xl bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </span>
              <div>
                <h2 class="text-base font-bold text-slate-900 dark:text-white flex items-center space-x-2">
                  <span>Live Activity Feed</span>
                  <span
                    class="text-2xs font-mono font-semibold px-2 py-0.5 rounded-full"
                    :class="isFeedPaused
                      ? 'bg-amber-100 text-amber-700 dark:bg-amber-950/60 dark:text-amber-400'
                      : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-950/60 dark:text-emerald-400'"
                  >
                    {{ isFeedPaused ? 'STREAM DIJEDA' : 'STREAM AKTIF' }}
                  </span>
                </h2>
                <p class="text-xs text-slate-500 dark:text-slate-400">
                  Event log operasional sistem diperbarui secara otomatis melalui WebSocket.
                </p>
              </div>
            </div>

            <!-- Auto-scroll & Filter toggles -->
            <div class="flex items-center space-x-3 text-xs">
              <label class="flex items-center space-x-2 cursor-pointer select-none">
                <input
                  type="checkbox"
                  v-model="autoScroll"
                  class="rounded text-indigo-600 focus:ring-indigo-500 h-4 w-4 bg-white dark:bg-slate-800 border-slate-300 dark:border-slate-700"
                />
                <span class="text-slate-600 dark:text-slate-300 font-medium text-3xs uppercase tracking-wider">Auto-Scroll</span>
              </label>

              <button
                @click="clearFilters"
                v-if="hasActiveFilters"
                class="px-2.5 py-1 text-3xs font-semibold rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 transition-colors"
              >
                Reset Filter
              </button>
            </div>
          </div>

          <!-- Filter Toolbar (Search, Severity, Category, Device) -->
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5 pt-2">
            <!-- Search Input -->
            <div class="relative">
              <input
                v-model="filters.search"
                @input="applyFiltersDebounced"
                type="text"
                placeholder="Cari event, pesan, error..."
                class="w-full pl-9 pr-3 py-2 text-xs rounded-xl bg-slate-50 dark:bg-[#0A0E18] border border-slate-200 dark:border-slate-800 text-slate-800 dark:text-slate-200 placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
              />
              <svg class="w-4 h-4 text-slate-400 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
              </svg>
            </div>

            <!-- Severity Dropdown -->
            <select
              v-model="filters.severity"
              @change="applyFilters"
              class="px-3 py-2 text-xs rounded-xl bg-slate-50 dark:bg-[#0A0E18] border border-slate-200 dark:border-slate-800 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
            >
              <option value="ALL">Semua Severity</option>
              <option value="DEBUG">DEBUG</option>
              <option value="INFO">INFO</option>
              <option value="WARNING">WARNING</option>
              <option value="ERROR">ERROR</option>
              <option value="CRITICAL">CRITICAL</option>
            </select>

            <!-- Category Dropdown -->
            <select
              v-model="filters.category"
              @change="applyFilters"
              class="px-3 py-2 text-xs rounded-xl bg-slate-50 dark:bg-[#0A0E18] border border-slate-200 dark:border-slate-800 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
            >
              <option value="ALL">Semua Kategori</option>
              <option value="POLLING">POLLING</option>
              <option value="COMMUNICATION">COMMUNICATION</option>
              <option value="SERIAL">SERIAL</option>
              <option value="MODBUS">MODBUS</option>
              <option value="TELEMETRY">TELEMETRY</option>
              <option value="ALARM">ALARM</option>
              <option value="SCHEDULER">SCHEDULER</option>
              <option value="SYSTEM">SYSTEM</option>
              <option value="DATABASE">DATABASE</option>
            </select>

            <!-- Device Dropdown -->
            <select
              v-model="filters.deviceId"
              @change="applyFilters"
              class="px-3 py-2 text-xs rounded-xl bg-slate-50 dark:bg-[#0A0E18] border border-slate-200 dark:border-slate-800 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500/20"
            >
              <option :value="0">Semua Perangkat</option>
              <option v-for="dev in pollingSnapshots" :key="dev.device_id" :value="dev.device_id">
                {{ dev.device_code }} ({{ dev.device_name }})
              </option>
            </select>
          </div>
        </div>

        <!-- Activity Feed List / Timeline -->
        <div ref="feedScrollRef" class="max-h-[520px] overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/60 font-sans">
          <div
            v-if="filteredEvents.length === 0"
            class="py-16 text-center text-slate-400 dark:text-slate-500 text-xs"
          >
            <svg class="w-8 h-8 mx-auto mb-2 text-slate-300 dark:text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
            </svg>
            Belum ada aktivitas yang sesuai dengan filter.
          </div>

          <div
            v-for="event in filteredEvents"
            :key="event.id"
            class="p-4 hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors flex flex-col sm:flex-row sm:items-start justify-between gap-3 text-xs"
          >
            <!-- Left: Severity Icon, Badges, Name, Description -->
            <div class="flex items-start space-x-3 min-w-0">
              <!-- Severity Dot Indicator -->
              <div class="mt-0.5 flex-shrink-0">
                <span
                  class="w-3 h-3 rounded-full flex items-center justify-center text-3xs font-bold"
                  :class="severityBadgeClass(event.severity)"
                ></span>
              </div>

              <div class="space-y-1 min-w-0">
                <div class="flex items-center space-x-2 flex-wrap gap-y-1">
                  <!-- Severity Badge -->
                  <span
                    class="px-2 py-0.5 rounded text-3xs font-mono font-bold tracking-wider uppercase border"
                    :class="severityBadgeStyle(event.severity)"
                  >
                    {{ event.severity }}
                  </span>

                  <!-- Category Badge -->
                  <span class="px-2 py-0.5 rounded text-3xs font-mono font-semibold bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
                    {{ event.category }}
                  </span>

                  <!-- Device Badge if present -->
                  <span
                    v-if="event.device_code"
                    class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-indigo-50 text-indigo-700 dark:bg-indigo-950/60 dark:text-indigo-300 border border-indigo-200 dark:border-indigo-800"
                  >
                    {{ event.device_code }}
                  </span>

                  <!-- Event Name -->
                  <span class="font-bold text-slate-900 dark:text-white font-mono text-2xs">
                    {{ event.event_name }}
                  </span>
                </div>

                <!-- Description -->
                <div class="text-slate-700 dark:text-slate-300 text-xs">
                  {{ event.description }}
                </div>

                <!-- Error Detail or Duration -->
                <div v-if="event.error_detail" class="text-3xs font-mono text-rose-600 dark:text-rose-400 bg-rose-50 dark:bg-rose-950/40 p-1.5 rounded-lg border border-rose-100 dark:border-rose-900/40 mt-1">
                  Detail: {{ event.error_detail }}
                </div>
              </div>
            </div>

            <!-- Right: Timestamp, Component, Duration -->
            <div class="flex-shrink-0 text-right sm:self-center font-mono text-3xs text-slate-400 space-y-0.5">
              <div class="font-semibold text-slate-600 dark:text-slate-300">
                {{ formatTimestamp(event.timestamp) }}
              </div>
              <div class="text-slate-400">
                {{ event.component || 'System' }}
                <span v-if="event.duration_ms" class="text-indigo-500 font-bold ml-1">({{ event.duration_ms }}ms)</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- SECTION E: BACKGROUND JOBS & SERVICE ACTIVITY -->
      <div class="rounded-3xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 shadow-xs overflow-hidden">
        <div class="p-5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <div class="flex items-center space-x-3">
            <span class="p-2 rounded-xl bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
              </svg>
            </span>
            <div>
              <h2 class="text-base font-bold text-slate-900 dark:text-white">
                Background Jobs & Service Activity
              </h2>
              <p class="text-xs text-slate-500 dark:text-slate-400">
                Status runtime proses internal backend (scheduler, pipeline rollup, data retention, backup, broadcaster).
              </p>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-5">
          <div
            v-for="job in backgroundJobs"
            :key="job.job_name"
            class="p-4 rounded-2xl bg-slate-50/60 dark:bg-[#0A0E18]/60 border border-slate-200 dark:border-slate-800/80 shadow-2xs space-y-2.5"
          >
            <div class="flex items-center justify-between">
              <span class="font-bold text-xs text-slate-900 dark:text-white font-mono">
                {{ formatJobName(job.job_name) }}
              </span>
              <span
                class="px-2 py-0.5 rounded text-3xs font-mono font-bold tracking-wider border uppercase"
                :class="jobStatusBadge(job.status)"
              >
                {{ job.status || 'RUNNING' }}
              </span>
            </div>

            <p class="text-3xs text-slate-500 dark:text-slate-400">
              {{ job.details || job.last_result || 'Proses berjalan di background' }}
            </p>

            <div class="text-3xs font-mono text-slate-400 space-y-1 pt-1 border-t border-slate-200/60 dark:border-slate-800">
              <div class="flex justify-between">
                <span>Terakhir Mulai:</span>
                <span class="text-slate-600 dark:text-slate-300 font-semibold">{{ formatTimestamp(job.last_started) }}</span>
              </div>
              <div v-if="job.last_duration_ms" class="flex justify-between">
                <span>Durasi:</span>
                <span class="text-indigo-500 font-bold">{{ job.last_duration_ms }} ms</span>
              </div>
              <div v-if="job.last_error" class="text-rose-500 truncate" :title="job.last_error">
                Err: {{ job.last_error }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- CONFIRMATION MODAL FOR CLEARING LOG BUFFER -->
    <div
      v-if="showClearModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
    >
      <div class="bg-white dark:bg-[#111827] rounded-3xl max-w-md w-full p-6 border border-slate-200 dark:border-slate-800 shadow-2xl space-y-4">
        <div class="w-12 h-12 rounded-2xl bg-rose-100 dark:bg-rose-950/60 text-rose-600 dark:text-rose-400 flex items-center justify-center mx-auto">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>
        </div>
        <div class="text-center">
          <h3 class="text-base font-bold text-slate-900 dark:text-white mb-1">
            Bersihkan Ring Buffer Log Aktivitas?
          </h3>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            Aksi ini akan mengosongkan riwayat aktivitas di memori server untuk monitor ini. Log kritis di database tetap tersimpan.
          </p>
        </div>
        <div class="flex space-x-3 pt-2">
          <button
            @click="showClearModal = false"
            class="flex-1 py-2.5 rounded-xl text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300"
          >
            Batal
          </button>
          <button
            @click="executeClearLogs"
            :disabled="clearLoading"
            class="flex-1 py-2.5 rounded-xl text-xs font-semibold bg-rose-600 hover:bg-rose-700 text-white shadow-xs"
          >
            <span v-if="clearLoading">Menghapus...</span>
            <span v-else>Ya, Bersihkan</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'SystemActivityMonitorView',
  data() {
    return {
      loading: false,
      detailLoading: false,
      clearLoading: false,
      wsConnected: false,
      isFeedPaused: false,
      autoScroll: true,
      showExportMenu: false,
      showClearModal: false,

      overview: {
        overall_status: 'RUNNING',
        backend_service_status: 'RUNNING',
        polling_engine_status: 'RUNNING',
        devices_online: 0,
        devices_offline: 0,
        devices_reconnecting: 0,
        total_devices: 0,
        active_polls_count: 0,
        waiting_or_timeout_count: 0,
        active_errors_count: 0,
        last_activity_time: null,
        uptime_seconds: 0,
      },

      pollingSnapshots: [],
      backgroundJobs: [],
      rawEvents: [],
      expandedDeviceId: null,
      deviceDetailData: null,

      filters: {
        search: '',
        severity: 'ALL',
        category: 'ALL',
        deviceId: 0,
      },

      debounceTimer: null,
      ws: null,
      pollFallbackTimer: null,
    };
  },

  computed: {
    userRole() {
      const u = this.$store.state.user || {};
      return (u.role || '').toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    hasActiveFilters() {
      return (
        this.filters.search !== '' ||
        this.filters.severity !== 'ALL' ||
        this.filters.category !== 'ALL' ||
        this.filters.deviceId !== 0
      );
    },
    filteredEvents() {
      let list = this.rawEvents;

      if (this.filters.severity !== 'ALL') {
        list = list.filter(e => (e.severity || '').toUpperCase() === this.filters.severity);
      }
      if (this.filters.category !== 'ALL') {
        list = list.filter(e => (e.category || '').toUpperCase() === this.filters.category);
      }
      if (this.filters.deviceId > 0) {
        list = list.filter(e => e.device_id === this.filters.deviceId);
      }
      if (this.filters.search.trim()) {
        const q = this.filters.search.toLowerCase();
        list = list.filter(e => {
          const combined = `${e.event_name || ''} ${e.description || ''} ${e.device_code || ''} ${e.component || ''} ${e.error_detail || ''}`.toLowerCase();
          return combined.includes(q);
        });
      }

      return list;
    },
  },

  mounted() {
    if (!this.isSuperAdmin) {
      return;
    }
    this.refreshAll();
    this.initWebSocket();

    // Fallback sync every 5 seconds to ensure snapshot counters remain up-to-date
    this.pollFallbackTimer = setInterval(() => {
      this.fetchOverview();
      this.fetchPollingSnapshots();
    }, 5000);

    // Click outside to close export menu
    document.addEventListener('click', this.handleClickOutside);
  },

  beforeDestroy() {
    if (this.ws) {
      try {
        this.ws.close();
      } catch (e) {
        // ignore
      }
    }
    if (this.pollFallbackTimer) {
      clearInterval(this.pollFallbackTimer);
    }
    document.removeEventListener('click', this.handleClickOutside);
  },

  methods: {
    getAuthHeaders() {
      const token = this.$store.state.token;
      return token ? { Authorization: `Bearer ${token}` } : {};
    },

    async refreshAll(manual = false) {
      this.loading = true;
      try {
        await Promise.all([
          this.fetchOverview(),
          this.fetchEvents(),
          this.fetchPollingSnapshots(),
          this.fetchBackgroundJobs(),
        ]);
        if (manual && this.$toast) {
          this.$toast.success('System Activity Monitor diperbarui');
        }
      } catch (err) {
        console.error('Failed refreshing activity monitor:', err);
      } finally {
        this.loading = false;
      }
    },

    async fetchOverview() {
      try {
        const res = await axios.get('/api/system/activity-monitor/overview', {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.data) {
          this.overview = res.data.data;
        }
      } catch (e) {
        console.warn('Overview fetch error:', e);
      }
    },

    async fetchEvents() {
      try {
        const res = await axios.get('/api/system/activity-monitor/events?limit=100', {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.data && res.data.data.events) {
          this.rawEvents = res.data.data.events;
        }
      } catch (e) {
        console.warn('Events fetch error:', e);
      }
    },

    async fetchPollingSnapshots() {
      try {
        const res = await axios.get('/api/system/activity-monitor/polling', {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.data) {
          this.pollingSnapshots = res.data.data;
        }
      } catch (e) {
        console.warn('Polling snapshots fetch error:', e);
      }
    },

    async fetchBackgroundJobs() {
      try {
        const res = await axios.get('/api/system/activity-monitor/background-jobs', {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.data) {
          this.backgroundJobs = res.data.data;
        }
      } catch (e) {
        console.warn('Background jobs fetch error:', e);
      }
    },

    async toggleDeviceExpand(deviceId) {
      if (this.expandedDeviceId === deviceId) {
        this.expandedDeviceId = null;
        this.deviceDetailData = null;
        return;
      }
      this.expandedDeviceId = deviceId;
      await this.fetchDeviceDetail(deviceId);
    },

    async fetchDeviceDetail(deviceId) {
      this.detailLoading = true;
      try {
        const res = await axios.get(`/api/system/activity-monitor/device/${deviceId}`, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.data) {
          this.deviceDetailData = res.data.data;
        }
      } catch (e) {
        console.warn('Device detail fetch error:', e);
      } finally {
        this.detailLoading = false;
      }
    },

    initWebSocket() {
      try {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          this.wsConnected = true;
        };

        this.ws.onclose = () => {
          this.wsConnected = false;
          // Reconnect with exponential delay
          setTimeout(() => {
            if (!this._isDestroyed) {
              this.initWebSocket();
            }
          }, 3000);
        };

        this.ws.onerror = () => {
          this.wsConnected = false;
        };

        this.ws.onmessage = (event) => {
          try {
            const msg = JSON.parse(event.data);
            this.handleWebSocketMessage(msg);
          } catch (err) {
            // non-json frame
          }
        };
      } catch (err) {
        console.warn('WebSocket connection error:', err);
      }
    },

    handleWebSocketMessage(msg) {
      if (!msg) return;

      // 1. New Activity Event Broadcast
      if (msg.type === 'SYSTEM_ACTIVITY_EVENT' && msg.data) {
        if (!this.isFeedPaused) {
          this.rawEvents.unshift(msg.data);
          // Ring buffer limit in browser
          if (this.rawEvents.length > 250) {
            this.rawEvents.pop();
          }
          if (this.autoScroll) {
            this.$nextTick(() => {
              if (this.$refs.feedScrollRef) {
                this.$refs.feedScrollRef.scrollTop = 0;
              }
            });
          }
        }

        // Live update polling snapshots based on activity events
        const ev = msg.data;
        if (ev.device_id && this.pollingSnapshots) {
          const target = this.pollingSnapshots.find(s => s.device_id === ev.device_id);
          if (target) {
            if (ev.event_name === 'USB_PORT_MISSING') {
              target.transport_status = 'PORT_MISSING';
              target.communication_status = 'OFFLINE';
              target.polling_state = 'STOPPED';
              target.is_fresh = false;
              target.last_error = ev.description;
            } else if (ev.event_name === 'SERIAL_OPEN_FAILED') {
              target.transport_status = 'OPEN_FAILED';
              target.communication_status = 'OFFLINE';
              target.is_fresh = false;
              target.last_error = ev.description;
            } else if (ev.event_name === 'COMMUNICATION_TIMEOUT') {
              target.polling_state = 'TIMEOUT';
              target.last_error = ev.description;
            } else if (ev.event_name === 'DEVICE_OFFLINE') {
              target.communication_status = 'OFFLINE';
              target.is_fresh = false;
              target.last_error = ev.description;
            } else if (ev.event_name === 'RECONNECT_STARTED') {
              target.recovery_state = 'RECONNECTING';
              target.polling_state = 'RECONNECTING';
            } else if (ev.event_name === 'SERIAL_PORT_REOPENED') {
              target.transport_status = 'AVAILABLE';
              target.recovery_state = 'RECONNECTING';
            } else if (ev.event_name === 'COMMUNICATION_RECOVERED') {
              target.transport_status = 'AVAILABLE';
              target.communication_status = 'ONLINE';
              target.polling_state = 'SUCCESS';
              target.recovery_state = 'NORMAL';
              target.is_fresh = true;
              target.consecutive_failures = 0;
            }
          }
        }
      }

      // 2. Real-time Device Connection Status change
      if (msg.type === 'device.connection.changed' && msg.data) {
        const dId = msg.data.device_id || msg.data.id;
        const newStatus = msg.data.connection_status || msg.data.status;
        const target = this.pollingSnapshots.find(s => s.device_id === dId);
        if (target) {
          target.communication_status = newStatus;
          if (msg.data.transport_status) {
            target.transport_status = msg.data.transport_status;
          }
          if (msg.data.administrative_status) {
            target.administrative_status = msg.data.administrative_status;
          }
          if (newStatus === 'ONLINE') {
            target.polling_state = 'SUCCESS';
            target.consecutive_failures = 0;
            target.recovery_state = 'NORMAL';
            target.is_fresh = true;
          } else if (newStatus === 'OFFLINE') {
            target.is_fresh = false;
            if (target.transport_status === 'PORT_MISSING') {
              target.polling_state = 'STOPPED';
            } else {
              target.polling_state = 'TIMEOUT';
            }
          } else if (newStatus === 'ERROR') {
            target.polling_state = 'TIMEOUT';
            target.is_fresh = false;
          }
        }
        // Update overview counters
        this.fetchOverview();
      }

      // 3. Telemetry Ingestion / Communication Success
      if (msg.type === 'device.communication.success' && msg.data) {
        const dId = msg.data.device_id;
        const target = this.pollingSnapshots.find(s => s.device_id === dId);
        if (target) {
          target.transport_status = 'AVAILABLE';
          target.communication_status = 'ONLINE';
          target.polling_state = 'SUCCESS';
          target.consecutive_failures = 0;
          target.is_fresh = true;
          target.last_latency_ms = msg.data.latency_ms || target.last_latency_ms;
          const nowStr = new Date().toISOString();
          target.last_successful_response = nowStr;
          target.last_successful_telemetry = nowStr;
        }
      }
    },

    toggleFeedPause() {
      this.isFeedPaused = !this.isFeedPaused;
    },

    clearFilters() {
      this.filters = {
        search: '',
        severity: 'ALL',
        category: 'ALL',
        deviceId: 0,
      };
    },

    applyFiltersDebounced() {
      clearTimeout(this.debounceTimer);
      this.debounceTimer = setTimeout(() => {
        // dynamic filter in computed
      }, 200);
    },

    applyFilters() {
      // computed handles filtering
    },

    confirmClearLogs() {
      this.showClearModal = true;
    },

    async executeClearLogs() {
      this.clearLoading = true;
      try {
        await axios.post('/api/system/activity-monitor/events/clear', {}, {
          headers: this.getAuthHeaders(),
        });
        this.rawEvents = [];
        this.showClearModal = false;
        if (this.$toast) {
          this.$toast.success('Ring buffer log aktivitas berhasil dibersihkan');
        }
      } catch (err) {
        console.error('Failed clearing log buffer:', err);
      } finally {
        this.clearLoading = false;
      }
    },

    exportLogs(format = 'csv') {
      this.showExportMenu = false;
      const url = `/api/system/activity-monitor/export?format=${format}`;
      const token = this.$store.state.token;

      // Use fetch to pass auth bearer header and download blob
      fetch(url, {
        headers: { Authorization: `Bearer ${token}` },
      })
        .then(res => res.blob())
        .then(blob => {
          const downloadUrl = window.URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.style.display = 'none';
          a.href = downloadUrl;
          a.download = `system_activity_${new Date().toISOString().replace(/[:.]/g, '-')}.${format}`;
          document.body.appendChild(a);
          a.click();
          window.URL.revokeObjectURL(downloadUrl);
          document.body.removeChild(a);
        })
        .catch(err => {
          console.error('Failed exporting logs:', err);
        });
    },

    handleClickOutside(e) {
      if (this.$refs.exportDropdownRef && !this.$refs.exportDropdownRef.contains(e.target)) {
        this.showExportMenu = false;
      }
    },

    formatJobName(raw) {
      if (!raw) return '';
      return raw.replace(/_/g, ' ').toUpperCase();
    },

    formatTimestamp(ts) {
      if (!ts) return '-';
      const d = new Date(ts);
      if (isNaN(d.getTime())) return '-';
      return d.toLocaleString('id-ID', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    },

    formatTimeOnly(ts) {
      if (!ts) return '-';
      const d = new Date(ts);
      if (isNaN(d.getTime())) return '-';
      return d.toLocaleTimeString('id-ID');
    },

    formatTimeAgo(ts) {
      if (!ts) return '';
      const seconds = Math.floor((new Date() - new Date(ts)) / 1000);
      if (seconds < 5) return 'baru saja';
      if (seconds < 60) return `${seconds} dtk lalu`;
      const minutes = Math.floor(seconds / 60);
      if (minutes < 60) return `${minutes} mnt lalu`;
      const hours = Math.floor(minutes / 60);
      return `${hours} jam lalu`;
    },

    formatUptime(sec) {
      if (!sec) return '0s';
      const d = Math.floor(sec / 86400);
      const h = Math.floor((sec % 86400) / 3600);
      const m = Math.floor((sec % 3600) / 60);
      const s = sec % 60;
      if (d > 0) return `${d}d ${h}h ${m}m`;
      if (h > 0) return `${h}h ${m}m ${s}s`;
      return `${m}m ${s}s`;
    },

    statusBgColor(status) {
      switch ((status || '').toUpperCase()) {
        case 'RUNNING':
          return 'bg-emerald-500 animate-pulse';
        case 'DEGRADED':
          return 'bg-amber-500';
        case 'STOPPED':
          return 'bg-rose-500';
        default:
          return 'bg-slate-400';
      }
    },

    statusTextColor(status) {
      switch ((status || '').toUpperCase()) {
        case 'RUNNING':
          return 'text-emerald-600 dark:text-emerald-400';
        case 'DEGRADED':
          return 'text-amber-600 dark:text-amber-400';
        case 'STOPPED':
          return 'text-rose-600 dark:text-rose-400';
        default:
          return 'text-slate-600 dark:text-slate-400';
      }
    },

    adminStatusBadge(status) {
      switch ((status || '').toUpperCase()) {
        case 'ACTIVE':
          return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900';
        case 'INACTIVE':
          return 'bg-slate-100 text-slate-500 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700';
        default:
          return 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700';
      }
    },

    transportStatusBadge(status) {
      switch ((status || '').toUpperCase()) {
        case 'AVAILABLE':
          return 'bg-emerald-50 text-emerald-700 border-emerald-300 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800';
        case 'PORT_MISSING':
          return 'bg-rose-50 text-rose-700 border-rose-300 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800 font-bold';
        case 'OPEN_FAILED':
          return 'bg-rose-50 text-rose-700 border-rose-300 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800';
        case 'DISCONNECTED':
          return 'bg-amber-50 text-amber-700 border-amber-300 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800';
        default:
          return 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700';
      }
    },

    transportStatusDot(status) {
      switch ((status || '').toUpperCase()) {
        case 'AVAILABLE':
          return 'bg-emerald-500';
        case 'PORT_MISSING':
          return 'bg-rose-500 animate-pulse';
        case 'OPEN_FAILED':
          return 'bg-rose-500';
        case 'DISCONNECTED':
          return 'bg-amber-500';
        default:
          return 'bg-slate-400';
      }
    },

    commStatusBadge(status) {
      switch ((status || '').toUpperCase()) {
        case 'ONLINE':
          return 'bg-emerald-50 text-emerald-700 border-emerald-300 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800';
        case 'RECONNECTING':
        case 'CONNECTING':
          return 'bg-amber-50 text-amber-700 border-amber-300 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800';
        case 'OFFLINE':
          return 'bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700';
        case 'ERROR':
          return 'bg-rose-50 text-rose-700 border-rose-300 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800';
        default:
          return 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400';
      }
    },

    commStatusDot(status) {
      switch ((status || '').toUpperCase()) {
        case 'ONLINE':
          return 'bg-emerald-500';
        case 'RECONNECTING':
        case 'CONNECTING':
          return 'bg-amber-500 animate-pulse';
        case 'OFFLINE':
          return 'bg-slate-400';
        case 'ERROR':
          return 'bg-rose-500';
        default:
          return 'bg-slate-400';
      }
    },

    pollingStateBadge(state) {
      switch ((state || '').toUpperCase()) {
        case 'POLLING':
          return 'bg-blue-100 text-blue-800 dark:bg-blue-950/60 dark:text-blue-300 border border-blue-200 dark:border-blue-800';
        case 'SUCCESS':
          return 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/60 dark:text-emerald-300';
        case 'TIMEOUT':
          return 'bg-amber-100 text-amber-800 dark:bg-amber-950/60 dark:text-amber-300';
        case 'RETRYING':
        case 'RECONNECTING':
          return 'bg-purple-100 text-purple-800 dark:bg-purple-950/60 dark:text-purple-300';
        case 'STOPPED':
          return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400';
        default:
          return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400';
      }
    },

    latencyColor(ms) {
      if (!ms) return 'text-slate-400';
      if (ms < 150) return 'text-emerald-600 dark:text-emerald-400';
      if (ms < 500) return 'text-amber-600 dark:text-amber-400';
      return 'text-rose-600 dark:text-rose-400';
    },

    recoveryStateBadge(state) {
      switch ((state || '').toUpperCase()) {
        case 'NORMAL':
          return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400';
        case 'BACKOFF':
          return 'bg-amber-100 text-amber-700 dark:bg-amber-950/60 dark:text-amber-400 font-bold';
        case 'RECONNECTING':
          return 'bg-purple-100 text-purple-700 dark:bg-purple-950/60 dark:text-purple-400 font-bold';
        default:
          return 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400';
      }
    },

    severityBadgeClass(sev) {
      switch ((sev || '').toUpperCase()) {
        case 'CRITICAL':
          return 'bg-rose-600 animate-ping';
        case 'ERROR':
          return 'bg-rose-500';
        case 'WARNING':
          return 'bg-amber-500';
        case 'INFO':
          return 'bg-blue-500';
        default:
          return 'bg-slate-400';
      }
    },

    severityBadgeStyle(sev) {
      switch ((sev || '').toUpperCase()) {
        case 'CRITICAL':
          return 'bg-rose-100 text-rose-800 border-rose-300 dark:bg-rose-950 dark:text-rose-200 dark:border-rose-800 font-extrabold';
        case 'ERROR':
          return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-900';
        case 'WARNING':
          return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-900';
        case 'INFO':
          return 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-950/40 dark:text-blue-300 dark:border-blue-900';
        default:
          return 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700';
      }
    },

    jobStatusBadge(status) {
      switch ((status || '').toUpperCase()) {
        case 'RUNNING':
          return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-950/40 dark:text-emerald-300 dark:border-emerald-800';
        case 'IDLE':
          return 'bg-slate-100 text-slate-600 border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700';
        case 'DEGRADED':
          return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-950/40 dark:text-amber-300 dark:border-amber-800';
        case 'STOPPED':
          return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-950/40 dark:text-rose-300 dark:border-rose-800';
        default:
          return 'bg-slate-100 text-slate-500 border-slate-200 dark:bg-slate-800 dark:text-slate-400';
      }
    },
  },
};
</script>

<style scoped>
/* High readability and custom scroll styling */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}
::-webkit-scrollbar-track {
  background: transparent;
}
::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.4);
  border-radius: 9999px;
}
::-webkit-scrollbar-thumb:hover {
  background: rgba(148, 163, 184, 0.6);
}
</style>
