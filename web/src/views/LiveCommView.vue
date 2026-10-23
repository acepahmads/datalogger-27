<template>
  <div class="space-y-5 font-sans text-slate-800 dark:text-slate-200">
    <!-- Top Bar Navigation & Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center space-x-2 text-xs text-slate-400 dark:text-slate-500 mb-1">
          <router-link to="/alarm/active" class="hover:text-blue-500 transition-colors">
            {{ $t('navigation.alarmsEvents') }}
          </router-link>
          <span>/</span>
          <span class="text-slate-700 dark:text-slate-300 font-semibold">{{ $t('liveComm.title') }}</span>
        </div>
        <div class="flex items-center space-x-3">
          <h1 class="text-xl font-bold tracking-tight text-slate-900 dark:text-white flex items-center gap-2.5">
            <span class="p-2 rounded-xl bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
              </svg>
            </span>
            {{ $t('liveComm.title') }}
          </h1>

          <!-- Active Pulsing Badge -->
          <span
            v-if="status && status.active"
            class="px-2.5 py-1 rounded-full text-2xs font-mono font-bold bg-emerald-50 text-emerald-700 border border-emerald-300 dark:bg-emerald-500/20 dark:text-emerald-300 dark:border-emerald-500/30 flex items-center space-x-1.5 shadow-2xs"
          >
            <span class="w-2 h-2 rounded-full bg-emerald-500 animate-ping"></span>
            <span>LIVE SNIFFER ACTIVE</span>
          </span>
          <span
            v-else
            class="px-2.5 py-1 rounded-full text-2xs font-mono font-bold bg-slate-100 text-slate-600 border border-slate-300 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700"
          >
            STANDBY (INACTIVE)
          </span>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1 max-w-2xl">
          {{ $t('liveComm.subtitle') }}
        </p>
      </div>

      <!-- Quick Superadmin Activation / Deactivation Toolbar -->
      <div class="flex items-center space-x-2.5 flex-shrink-0">
        <!-- Target Device Filter Selection when activating -->
        <select
          v-if="canToggle && (!status || !status.active)"
          v-model="targetDeviceId"
          class="px-3 py-2 text-xs rounded-xl bg-white dark:bg-[#111827] border border-[#C7D5E8] dark:border-slate-800 text-slate-800 dark:text-slate-200 shadow-2xs focus:ring-2 focus:ring-blue-500/20 focus:outline-none"
        >
          <option :value="0">{{ $t('liveComm.allDevices') }}</option>
          <option v-for="dev in devices" :key="dev.id" :value="dev.id">
            {{ dev.device_code || dev.name }} (ID: {{ dev.id }})
          </option>
        </select>

        <!-- Toggle Button (Superadmin Only) -->
        <button
          v-if="canToggle"
          @click="toggleTrace"
          :disabled="toggleLoading"
          class="px-4 py-2 rounded-xl text-xs font-semibold transition-all shadow-xs flex items-center space-x-2 disabled:opacity-50"
          :class="(status && status.active)
            ? 'bg-rose-600 hover:bg-rose-700 text-white shadow-rose-600/20'
            : 'bg-emerald-600 hover:bg-emerald-700 text-white shadow-emerald-600/20'"
        >
          <svg v-if="toggleLoading" class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <svg v-else-if="status && status.active" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z" />
          </svg>
          <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{{ (status && status.active) ? $t('liveComm.stopSniffer') : $t('liveComm.startSniffer') }}</span>
        </button>

        <!-- Viewer / Non-Superadmin Badge -->
        <span
          v-else
          class="px-3 py-2 rounded-xl text-3xs font-semibold bg-slate-100 text-slate-500 border border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700 flex items-center space-x-1.5"
          :title="$t('liveComm.superadminOnlyBadge')"
        >
          <svg class="w-3.5 h-3.5 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"/>
          </svg>
          <span>{{ $t('liveComm.superadminOnlyBadge') }}</span>
        </span>
      </div>
    </div>

    <!-- Active State Information Banner with Countdown Timer -->
    <div
      v-if="status && status.active"
      class="p-4 rounded-2xl border border-emerald-200 bg-emerald-50/80 dark:bg-emerald-950/20 dark:border-emerald-500/30 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs shadow-2xs"
    >
      <div class="flex items-center space-x-3">
        <div class="w-8 h-8 rounded-xl bg-emerald-100 dark:bg-emerald-500/20 text-emerald-700 dark:text-emerald-400 flex items-center justify-center flex-shrink-0">
          <svg class="w-4 h-4 animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
          </svg>
        </div>
        <div>
          <div class="font-bold text-emerald-900 dark:text-emerald-300 flex items-center space-x-2">
            <span>{{ $t('liveComm.autoTimeoutNotice') }}</span>
            <span v-if="status.activated_by" class="font-mono text-3xs font-semibold px-2 py-0.5 rounded bg-emerald-100 dark:bg-emerald-900/50 text-emerald-800 dark:text-emerald-300">
              @{{ status.activated_by }}
            </span>
          </div>
          <p class="text-3xs text-emerald-700 dark:text-emerald-400/90 mt-0.5">
            Streaming raw Modbus serial and socket packets in real-time. Frames are buffered in memory and broadcast over WebSocket.
          </p>
        </div>
      </div>

      <div class="flex items-center space-x-3 flex-shrink-0">
        <div class="px-3 py-1.5 rounded-xl bg-white/90 dark:bg-slate-900 border border-emerald-300 dark:border-emerald-600/50 text-emerald-800 dark:text-emerald-300 font-mono font-bold text-xs flex items-center space-x-1.5 shadow-2xs">
          <svg class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>{{ countdownFormatted }}</span>
        </div>
      </div>
    </div>

    <!-- Inactive Notice Banner -->
    <div
      v-else
      class="p-4 rounded-2xl border border-[#C7D5E8] bg-slate-50 dark:bg-[#0E2034] dark:border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs shadow-2xs"
    >
      <div class="flex items-center space-x-3">
        <div class="w-8 h-8 rounded-xl bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400 flex items-center justify-center flex-shrink-0">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636" />
          </svg>
        </div>
        <div>
          <div class="font-bold text-slate-800 dark:text-slate-200">
            {{ $t('liveComm.inactiveNotice') }}
          </div>
          <p class="text-3xs text-slate-500 dark:text-slate-400 mt-0.5">
            Zero CPU and serial bus overhead while inactive. Click "Start Live Stream" above to begin monitoring live sensor communication.
          </p>
        </div>
      </div>
    </div>

    <!-- Real-Time Metrics Row -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3.5">
      <div class="p-3.5 bg-white dark:bg-[#0E2034] rounded-xl border border-[#C7D5E8] dark:border-slate-800 min-h-[72px] flex flex-col justify-between shadow-2xs">
        <div class="text-3xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">{{ $t('liveComm.packetsCaptured') }}</div>
        <div class="text-lg font-mono font-bold text-slate-900 dark:text-white mt-1">{{ filteredFrames.length }}</div>
      </div>

      <div class="p-3.5 bg-white dark:bg-[#0E2034] rounded-xl border border-[#C7D5E8] dark:border-slate-800 min-h-[72px] flex flex-col justify-between shadow-2xs">
        <div class="text-3xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">{{ $t('liveComm.txFrames') }}</div>
        <div class="text-lg font-mono font-bold text-blue-600 dark:text-blue-400 mt-1">{{ txCount }}</div>
      </div>

      <div class="p-3.5 bg-white dark:bg-[#0E2034] rounded-xl border border-[#C7D5E8] dark:border-slate-800 min-h-[72px] flex flex-col justify-between shadow-2xs">
        <div class="text-3xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">{{ $t('liveComm.rxFrames') }}</div>
        <div class="text-lg font-mono font-bold text-emerald-600 dark:text-emerald-400 mt-1">{{ rxCount }}</div>
      </div>

      <div class="p-3.5 bg-white dark:bg-[#0E2034] rounded-xl border border-[#C7D5E8] dark:border-slate-800 min-h-[72px] flex flex-col justify-between shadow-2xs">
        <div class="text-3xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">{{ $t('liveComm.avgLatency') }}</div>
        <div class="text-lg font-mono font-bold text-indigo-600 dark:text-indigo-400 mt-1">
          {{ avgLatency }} <span class="text-xs font-sans text-slate-400 font-normal">ms</span>
        </div>
      </div>

      <div class="p-3.5 bg-white dark:bg-[#0E2034] rounded-xl border border-[#C7D5E8] dark:border-slate-800 min-h-[72px] flex flex-col justify-between shadow-2xs">
        <div class="text-3xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wider">{{ $t('liveComm.errors') }}</div>
        <div class="text-lg font-mono font-bold" :class="errorCount > 0 ? 'text-rose-600 dark:text-rose-400' : 'text-slate-700 dark:text-slate-300'">
          {{ errorCount }}
        </div>
      </div>
    </div>

    <!-- Packet Stream Console Card -->
    <div class="bg-white dark:bg-[#0E2034] rounded-2xl border border-[#C7D5E8] dark:border-slate-800 shadow-sm overflow-hidden flex flex-col">
      <!-- Toolbar & Filter Controls -->
      <div class="p-4 border-b border-[#E2EAF4] dark:border-slate-800/80 bg-slate-50/70 dark:bg-[#0B1524] flex flex-col lg:flex-row lg:items-center justify-between gap-3 text-xs">
        <div class="flex flex-wrap items-center gap-2.5">
          <!-- Device Filter -->
          <div class="flex items-center space-x-1.5">
            <span class="text-3xs font-semibold uppercase text-slate-500">{{ $t('liveComm.device') }}:</span>
            <select
              v-model="filterDevice"
              class="px-2.5 py-1 text-xs rounded-lg bg-white dark:bg-slate-900 border border-[#C7D5E8] dark:border-slate-800 text-slate-800 dark:text-slate-200 focus:outline-none"
            >
              <option value="">{{ $t('liveComm.allDevices') }}</option>
              <option v-for="dev in devices" :key="dev.id" :value="dev.device_code">
                {{ dev.device_code || dev.name }}
              </option>
            </select>
          </div>

          <!-- Status Filter -->
          <div class="flex items-center space-x-1.5">
            <span class="text-3xs font-semibold uppercase text-slate-500">{{ $t('liveComm.status') }}:</span>
            <select
              v-model="filterStatus"
              class="px-2.5 py-1 text-xs rounded-lg bg-white dark:bg-slate-900 border border-[#C7D5E8] dark:border-slate-800 text-slate-800 dark:text-slate-200 focus:outline-none"
            >
              <option value="">{{ $t('liveComm.allStatuses') }}</option>
              <option value="SUCCESS">SUCCESS</option>
              <option value="TIMEOUT">TIMEOUT</option>
              <option value="CRC_ERROR">CRC ERROR</option>
              <option value="EXCEPTION">EXCEPTION</option>
              <option value="ERROR">ERROR</option>
            </select>
          </div>

          <!-- Text Search -->
          <div class="relative min-w-[200px]">
            <input
              type="text"
              v-model="searchQuery"
              :placeholder="$t('liveComm.searchPlaceholder')"
              class="w-full pl-7 pr-3 py-1 text-xs rounded-lg bg-white dark:bg-slate-900 border border-[#C7D5E8] dark:border-slate-800 text-slate-800 dark:text-slate-200 focus:outline-none"
            />
            <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"/>
            </svg>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex flex-wrap items-center gap-2">
          <!-- Auto-Scroll Toggle -->
          <button
            @click="autoScroll = !autoScroll"
            class="px-2.5 py-1 rounded-lg border text-2xs font-semibold transition flex items-center space-x-1.5"
            :class="autoScroll ? 'bg-blue-50 text-blue-700 border-blue-200 dark:bg-blue-600/20 dark:text-blue-300 dark:border-blue-500/30' : 'bg-white text-slate-600 border-slate-300 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700'"
            :title="$t('liveComm.autoScroll')"
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="autoScroll ? 'bg-blue-500' : 'bg-slate-400'"></span>
            <span>{{ $t('liveComm.autoScroll') }}</span>
          </button>

          <!-- Pause / Resume -->
          <button
            @click="isPaused = !isPaused"
            class="px-2.5 py-1 rounded-lg border text-2xs font-semibold transition flex items-center space-x-1.5"
            :class="isPaused ? 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-500/20 dark:text-amber-300 dark:border-amber-500/30' : 'bg-white text-slate-600 border-slate-300 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700'"
          >
            <span>{{ isPaused ? $t('liveComm.resumeStream') : $t('liveComm.pauseStream') }}</span>
          </button>

          <!-- Clear Buffer -->
          <button
            @click="clearConsole"
            class="px-2.5 py-1 rounded-lg bg-white hover:bg-slate-100 text-slate-700 border border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-700 text-2xs font-semibold transition"
          >
            {{ $t('liveComm.clearBuffer') }}
          </button>

          <!-- Export Actions -->
          <button
            @click="exportLogs('json')"
            class="px-2 py-1 rounded-lg bg-white hover:bg-slate-100 text-slate-700 border border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-700 text-2xs font-semibold transition"
            :title="$t('liveComm.exportJson')"
          >
            JSON
          </button>
          <button
            @click="exportLogs('txt')"
            class="px-2 py-1 rounded-lg bg-white hover:bg-slate-100 text-slate-700 border border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-700 text-2xs font-semibold transition"
            :title="$t('liveComm.exportTxt')"
          >
            TXT
          </button>
        </div>
      </div>

      <!-- Live Packets Table / Terminal -->
      <div ref="packetContainer" class="overflow-x-auto max-h-[560px] overflow-y-auto">
        <table class="w-full text-left text-xs font-sans border-collapse">
          <thead class="bg-slate-100/90 dark:bg-slate-900/90 text-slate-600 dark:text-slate-400 text-3xs font-semibold uppercase tracking-wider sticky top-0 z-10 border-b border-[#C7D5E8] dark:border-slate-800 backdrop-blur-xs">
            <tr>
              <th class="py-2.5 px-3 w-28">{{ $t('liveComm.timestamp') }}</th>
              <th class="py-2.5 px-3 w-28">{{ $t('liveComm.device') }}</th>
              <th class="py-2.5 px-3 w-40">{{ $t('liveComm.function') }}</th>
              <th class="py-2.5 px-3 w-36">{{ $t('liveComm.address') }}</th>
              <th class="py-2.5 px-3">{{ $t('liveComm.hexPayload') }}</th>
              <th class="py-2.5 px-3 w-20 text-right">{{ $t('liveComm.latency') }}</th>
              <th class="py-2.5 px-3 w-24 text-center">{{ $t('liveComm.status') }}</th>
              <th class="py-2.5 px-3 w-20 text-center">Inspect</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#E2ECF7] dark:divide-slate-800/60 font-mono text-3xs bg-white dark:bg-[#0B1524]">
            <tr v-if="filteredFrames.length === 0">
              <td colspan="8" class="py-16 text-center text-slate-500 font-sans">
                <div class="flex flex-col items-center justify-center space-y-2">
                  <div class="w-10 h-10 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center text-slate-400">
                    <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                    </svg>
                  </div>
                  <h4 class="text-xs font-bold text-slate-700 dark:text-slate-300">{{ $t('liveComm.noPackets') }}</h4>
                  <p class="text-3xs text-slate-400 max-w-sm">{{ $t('liveComm.noPacketsDesc') }}</p>
                </div>
              </td>
            </tr>

            <tr
              v-for="frame in filteredFrames"
              :key="frame.id"
              @click="openInspectModal(frame)"
              class="hover:bg-blue-50/60 dark:hover:bg-slate-800/40 transition cursor-pointer"
            >
              <!-- Timestamp -->
              <td class="py-2 px-3 text-slate-600 dark:text-slate-400 whitespace-nowrap">
                {{ formatTime(frame.timestamp) }}
              </td>

              <!-- Device Code & Transport -->
              <td class="py-2 px-3">
                <div class="flex flex-col">
                  <span class="font-bold text-slate-900 dark:text-slate-100">{{ frame.device_code }}</span>
                  <span class="text-4xs text-slate-400 truncate max-w-[110px]" :title="frame.transport">{{ frame.transport }}</span>
                </div>
              </td>

              <!-- Function Code -->
              <td class="py-2 px-3">
                <span class="px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-500/10 dark:text-blue-300 dark:border-blue-500/20 font-semibold">
                  {{ frame.fc_name }}
                </span>
              </td>

              <!-- Address & Qty -->
              <td class="py-2 px-3 whitespace-nowrap">
                <span class="font-bold text-slate-900 dark:text-slate-200">#{{ frame.display_address }}</span>
                <span class="text-slate-400"> (PDU 0x{{ frame.starting_address.toString(16).toUpperCase().padStart(4, '0') }})</span>
                <span class="text-slate-500 block text-4xs">Qty: {{ frame.quantity }}</span>
              </td>

              <!-- Raw HEX Stream (TX & RX) -->
              <td class="py-2 px-3">
                <div class="space-y-1">
                  <!-- TX Request Frame -->
                  <div class="flex items-center space-x-1.5">
                    <span class="px-1 py-0.2 rounded text-4xs font-bold bg-blue-100 text-blue-800 border border-blue-300 dark:bg-blue-500/20 dark:text-blue-300 dark:border-blue-500/40">TX</span>
                    <span class="text-3xs font-semibold text-blue-700 dark:text-blue-300 tracking-wider select-all" :title="'Request: ' + frame.tx_hex">
                      {{ frame.tx_hex || '--' }}
                    </span>
                    <span class="text-4xs text-slate-400 font-sans">({{ frame.tx_bytes }}B)</span>
                  </div>

                  <!-- RX Response Frame -->
                  <div class="flex items-center space-x-1.5">
                    <span class="px-1 py-0.2 rounded text-4xs font-bold" :class="frame.status === 'SUCCESS' ? 'bg-emerald-100 text-emerald-800 border border-emerald-300 dark:bg-emerald-500/20 dark:text-emerald-300 dark:border-emerald-500/40' : 'bg-rose-100 text-rose-800 border border-rose-300 dark:bg-rose-500/20 dark:text-rose-300 dark:border-rose-500/40'">RX</span>
                    <span v-if="frame.rx_hex" class="text-3xs font-semibold tracking-wider select-all" :class="frame.status === 'SUCCESS' ? 'text-emerald-700 dark:text-emerald-300' : 'text-rose-600 dark:text-rose-400'" :title="'Response: ' + frame.rx_hex">
                      {{ frame.rx_hex }}
                    </span>
                    <span v-else class="text-rose-600 dark:text-rose-400 text-4xs italic">
                      {{ frame.error_message || 'No response received (Timeout / I/O error)' }}
                    </span>
                    <span v-if="frame.rx_bytes > 0" class="text-4xs text-slate-400 font-sans">({{ frame.rx_bytes }}B)</span>
                  </div>
                </div>
              </td>

              <!-- Latency -->
              <td class="py-2 px-3 text-right font-bold text-slate-900 dark:text-slate-100 tabular-nums">
                {{ frame.latency_ms }} ms
              </td>

              <!-- Status Badge -->
              <td class="py-2 px-3 text-center">
                <span
                  class="px-1.5 py-0.5 rounded text-4xs font-bold uppercase tracking-wider"
                  :class="statusBadgeClass(frame.status)"
                >
                  {{ frame.status }}
                </span>
              </td>

              <!-- Inspect Action -->
              <td class="py-2 px-3 text-center">
                <button
                  type="button"
                  @click.stop="openInspectModal(frame)"
                  class="p-1 rounded text-slate-400 hover:text-blue-600 dark:hover:text-blue-400 transition"
                  title="Inspect packet details"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Packet Inspector Modal / Detail Drawer -->
    <div
      v-if="inspectedFrame"
      class="fixed inset-0 z-50 bg-black/60 backdrop-blur-xs flex items-center justify-center p-4 animate-fade-in"
      @click.self="inspectedFrame = null"
    >
      <div class="bg-white dark:bg-[#0E2034] border border-[#C7D5E8] dark:border-slate-800 rounded-2xl w-full max-w-2xl shadow-2xl overflow-hidden font-sans space-y-0">
        <!-- Modal Header -->
        <div class="px-5 py-4 border-b border-[#E2EAF4] dark:border-slate-800 flex items-center justify-between bg-slate-50/80 dark:bg-[#0B1524]">
          <div class="flex items-center space-x-2.5">
            <span class="p-1.5 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
              </svg>
            </span>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white">
                {{ $t('liveComm.packetDetails') }} #{{ inspectedFrame.id }}
              </h3>
              <p class="text-3xs text-slate-500 dark:text-slate-400">
                {{ inspectedFrame.device_code }} • {{ inspectedFrame.summary }}
              </p>
            </div>
          </div>

          <button
            @click="inspectedFrame = null"
            class="text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 text-sm font-bold p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          >
            ✕
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-5 space-y-4 max-h-[70vh] overflow-y-auto text-xs">
          <!-- Metadata Grid -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5 font-sans">
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800">
              <div class="text-4xs text-slate-500 uppercase font-semibold">Timestamp</div>
              <div class="font-mono text-3xs font-bold text-slate-900 dark:text-slate-100 mt-0.5">{{ formatTime(inspectedFrame.timestamp) }}</div>
            </div>

            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800">
              <div class="text-4xs text-slate-500 uppercase font-semibold">Slave / Unit ID</div>
              <div class="font-mono text-3xs font-bold text-amber-600 dark:text-amber-400 mt-0.5">#{{ inspectedFrame.slave_id }}</div>
            </div>

            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800">
              <div class="text-4xs text-slate-500 uppercase font-semibold">Roundtrip Latency</div>
              <div class="font-mono text-3xs font-bold text-indigo-600 dark:text-indigo-400 mt-0.5">{{ inspectedFrame.latency_ms }} ms</div>
            </div>

            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800">
              <div class="text-4xs text-slate-500 uppercase font-semibold">Execution Status</div>
              <div class="font-mono text-3xs font-bold mt-0.5" :class="inspectedFrame.status === 'SUCCESS' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">
                {{ inspectedFrame.status }}
              </div>
            </div>
          </div>

          <!-- TX Request Inspector Panel -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
              <span>TX Request Frame ({{ inspectedFrame.tx_bytes }} Bytes)</span>
              <button
                @click="copyText(inspectedFrame.tx_hex, 'TX Hex copied')"
                class="text-blue-600 dark:text-blue-400 hover:underline flex items-center space-x-1"
              >
                <span>{{ $t('liveComm.copyHex') }}</span>
              </button>
            </div>
            <div class="p-3 bg-slate-100 dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 font-mono text-xs text-blue-700 dark:text-blue-300 break-all select-all">
              {{ inspectedFrame.tx_hex || '--' }}
            </div>
          </div>

          <!-- RX Response Inspector Panel -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
              <span>RX Response Frame ({{ inspectedFrame.rx_bytes }} Bytes)</span>
              <button
                v-if="inspectedFrame.rx_hex"
                @click="copyText(inspectedFrame.rx_hex, 'RX Hex copied')"
                class="text-emerald-600 dark:text-emerald-400 hover:underline flex items-center space-x-1"
              >
                <span>{{ $t('liveComm.copyHex') }}</span>
              </button>
            </div>
            <div class="p-3 bg-slate-100 dark:bg-slate-950 rounded-xl border border-slate-200 dark:border-slate-800 font-mono text-xs break-all select-all" :class="inspectedFrame.status === 'SUCCESS' ? 'text-emerald-700 dark:text-emerald-300' : 'text-rose-600 dark:text-rose-400'">
              {{ inspectedFrame.rx_hex || inspectedFrame.error_message || 'No response bytes' }}
            </div>
          </div>

          <!-- Error Details if any -->
          <div v-if="inspectedFrame.error_message" class="p-3 rounded-xl bg-rose-50 dark:bg-rose-950/20 border border-rose-200 dark:border-rose-500/30 text-rose-800 dark:text-rose-300 text-xs">
            <div class="font-bold flex items-center space-x-1.5 mb-1">
              <svg class="w-3.5 h-3.5 text-rose-600 dark:text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
              </svg>
              <span>Communication Failure Diagnostic</span>
            </div>
            <p class="font-mono text-3xs">{{ inspectedFrame.error_message }}</p>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="px-5 py-3 border-t border-[#E2EAF4] dark:border-slate-800 flex justify-end bg-slate-50/80 dark:bg-[#0B1524]">
          <button
            @click="inspectedFrame = null"
            class="px-4 py-1.5 rounded-xl bg-slate-200 hover:bg-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 text-xs font-semibold transition"
          >
            {{ $t('liveComm.close') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Floating Toast Notification -->
    <div
      v-if="toastMessage"
      class="fixed bottom-6 right-6 z-50 flex items-center space-x-2.5 px-4 py-2.5 rounded-xl shadow-xl border text-xs font-semibold transition-all animate-fade-in"
      :class="toastType === 'error' ? 'bg-rose-950 text-rose-200 border-rose-500/40' : 'bg-emerald-950 text-emerald-200 border-emerald-500/40'"
    >
      <span>{{ toastMessage }}</span>
      <button @click="toastMessage = null" class="opacity-75 hover:opacity-100 text-3xs font-mono ml-1">✕</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'LiveCommView',
  data() {
    return {
      status: null,
      statusTimer: null,
      countdownTimer: null,
      remainingSeconds: 0,
      toggleLoading: false,
      targetDeviceId: 0,
      devices: [],
      frames: [],
      maxInMemoryFrames: 200,
      autoScroll: true,
      isPaused: false,
      filterDevice: '',
      filterStatus: '',
      searchQuery: '',
      inspectedFrame: null,
      toastMessage: null,
      toastType: 'success',
      toastTimeout: null,
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
    canToggle() {
      return this.isSuperAdmin || (this.status && this.status.can_toggle);
    },
    countdownFormatted() {
      const sec = this.remainingSeconds > 0 ? this.remainingSeconds : (this.status ? this.status.remaining_seconds : 0);
      if (sec <= 0) return '00:00';
      const m = Math.floor(sec / 60);
      const s = sec % 60;
      return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')} remaining`;
    },
    filteredFrames() {
      return this.frames.filter(f => {
        if (this.filterDevice && f.device_code !== this.filterDevice) {
          return false;
        }
        if (this.filterStatus && f.status !== this.filterStatus) {
          return false;
        }
        if (this.searchQuery) {
          const q = this.searchQuery.toLowerCase();
          const matchHex = (f.tx_hex && f.tx_hex.toLowerCase().includes(q)) || (f.rx_hex && f.rx_hex.toLowerCase().includes(q));
          const matchErr = f.error_message && f.error_message.toLowerCase().includes(q);
          const matchSummary = f.summary && f.summary.toLowerCase().includes(q);
          if (!matchHex && !matchErr && !matchSummary) {
            return false;
          }
        }
        return true;
      });
    },
    txCount() {
      return this.filteredFrames.filter(f => f.tx_hex).length;
    },
    rxCount() {
      return this.filteredFrames.filter(f => f.rx_hex).length;
    },
    avgLatency() {
      const valid = this.filteredFrames.filter(f => f.latency_ms > 0);
      if (valid.length === 0) return 0;
      const sum = valid.reduce((acc, f) => acc + f.latency_ms, 0);
      return Math.round(sum / valid.length);
    },
    errorCount() {
      return this.filteredFrames.filter(f => f.status !== 'SUCCESS').length;
    },
  },
  methods: {
    statusBadgeClass(status) {
      switch (status) {
        case 'SUCCESS':
          return 'bg-emerald-100 text-emerald-800 border border-emerald-300 dark:bg-emerald-500/20 dark:text-emerald-300 dark:border-emerald-500/30';
        case 'TIMEOUT':
          return 'bg-amber-100 text-amber-800 border border-amber-300 dark:bg-amber-500/20 dark:text-amber-300 dark:border-amber-500/30';
        case 'CRC_ERROR':
        case 'EXCEPTION':
        case 'ERROR':
        default:
          return 'bg-rose-100 text-rose-800 border border-rose-300 dark:bg-rose-500/20 dark:text-rose-300 dark:border-rose-500/30';
      }
    },
    formatTime(ts) {
      if (!ts) return '--:--:--';
      const d = new Date(ts);
      const h = String(d.getHours()).padStart(2, '0');
      const m = String(d.getMinutes()).padStart(2, '0');
      const s = String(d.getSeconds()).padStart(2, '0');
      const ms = String(d.getMilliseconds()).padStart(3, '0');
      return `${h}:${m}:${s}.${ms}`;
    },
    async loadStatus() {
      const res = await this.$store.dispatch('fetchLiveCommStatus');
      if (res) {
        this.status = res;
        this.remainingSeconds = res.remaining_seconds || 0;
      }
    },
    async loadRecentFrames() {
      const data = await this.$store.dispatch('fetchLiveCommFrames', { limit: 100 });
      if (data && Array.isArray(data)) {
        this.frames = data;
        this.scrollToBottom();
      }
    },
    async loadDevices() {
      try {
        const res = await this.$store.dispatch('fetchDevices');
        this.devices = (res && res.data) ? res.data : (res || []);
      } catch (e) {
        // ignore
      }
    },
    async toggleTrace() {
      if (!this.canToggle) return;
      this.toggleLoading = true;
      const targetActive = !(this.status && this.status.active);
      try {
        const payload = {
          active: targetActive,
          target_device_id: this.targetDeviceId,
          duration_minutes: 10,
        };
        const res = await this.$store.dispatch('toggleLiveCommTrace', payload);
        this.showToast(res.message || (targetActive ? 'Live sniffer started' : 'Live sniffer stopped'), 'success');
        await this.loadStatus();
        if (targetActive) {
          await this.loadRecentFrames();
        }
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.error) || err.message;
        this.showToast(msg, 'error');
      } finally {
        this.toggleLoading = false;
      }
    },
    async clearConsole() {
      try {
        if (this.canToggle) {
          await this.$store.dispatch('clearLiveCommFrames');
        }
        this.frames = [];
        this.showToast('Console cleared', 'success');
      } catch (e) {
        this.frames = [];
      }
    },
    onLiveFrame(event) {
      if (this.isPaused) return;
      if (event && event.detail) {
        this.frames.push(event.detail);
        if (this.frames.length > this.maxInMemoryFrames) {
          this.frames.shift();
        }
        this.scrollToBottom();
      }
    },
    onTraceStatus(event) {
      if (event && event.detail) {
        this.status = event.detail;
        this.remainingSeconds = event.detail.remaining_seconds || 0;
      }
    },
    scrollToBottom() {
      if (!this.autoScroll) return;
      this.$nextTick(() => {
        const c = this.$refs.packetContainer;
        if (c) {
          c.scrollTop = c.scrollHeight;
        }
      });
    },
    openInspectModal(frame) {
      this.inspectedFrame = frame;
    },
    copyText(text, msg) {
      if (!text) return;
      if (navigator && navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
          this.showToast(msg || 'Copied to clipboard', 'success');
        });
      } else {
        const el = document.createElement('textarea');
        el.value = text;
        document.body.appendChild(el);
        el.select();
        document.execCommand('copy');
        document.body.removeChild(el);
        this.showToast(msg || 'Copied to clipboard', 'success');
      }
    },
    exportLogs(format) {
      const data = this.filteredFrames;
      if (data.length === 0) {
        this.showToast('No frames to export', 'error');
        return;
      }

      let content = '';
      let mime = 'text/plain';
      let filename = `comm_trace_${new Date().toISOString().replace(/[:.]/g, '-')}`;

      if (format === 'json') {
        content = JSON.stringify(data, null, 2);
        mime = 'application/json';
        filename += '.json';
      } else {
        filename += '.txt';
        content = data.map(f => {
          return `[${f.timestamp}] [${f.device_code}] ${f.fc_name} Addr:#${f.display_address} Status:${f.status} Latency:${f.latency_ms}ms\nTX: ${f.tx_hex}\nRX: ${f.rx_hex || f.error_message}\n`;
        }).join('\n----------------------------------------\n');
      }

      const blob = new Blob([content], { type: mime });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
      this.showToast(`Exported ${data.length} frames (${format.toUpperCase()})`, 'success');
    },
    showToast(msg, type = 'success') {
      this.toastMessage = msg;
      this.toastType = type;
      if (this.toastTimeout) clearTimeout(this.toastTimeout);
      this.toastTimeout = setTimeout(() => {
        this.toastMessage = null;
      }, 3000);
    },
  },
  mounted() {
    this.loadStatus();
    this.loadDevices();
    this.loadRecentFrames();

    // Event listeners
    window.addEventListener('comm-live-frame', this.onLiveFrame);
    window.addEventListener('comm-trace-status', this.onTraceStatus);

    // 1-second countdown clock ticker
    this.countdownTimer = setInterval(() => {
      if (this.status && this.status.active && this.remainingSeconds > 0) {
        this.remainingSeconds--;
        if (this.remainingSeconds <= 0) {
          this.loadStatus();
        }
      }
    }, 1000);

    // Background status check every 15 seconds
    this.statusTimer = setInterval(() => {
      this.loadStatus();
    }, 15000);
  },
  beforeDestroy() {
    window.removeEventListener('comm-live-frame', this.onLiveFrame);
    window.removeEventListener('comm-trace-status', this.onTraceStatus);
    if (this.countdownTimer) clearInterval(this.countdownTimer);
    if (this.statusTimer) clearInterval(this.statusTimer);
    if (this.toastTimeout) clearTimeout(this.toastTimeout);
  },
};
</script>

<style scoped>
.shadow-2xs {
  box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.03);
}
</style>
