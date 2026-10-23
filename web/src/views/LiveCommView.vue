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
          <option v-for="dev in deviceList" :key="dev.id" :value="dev.id">
            {{ dev.device_code || dev.device_name || dev.name || ('Device #' + dev.id) }} (ID: {{ dev.id }})
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
              <option v-for="dev in deviceList" :key="dev.id" :value="dev.device_code">
                {{ dev.device_code || dev.device_name || dev.name || ('Device #' + dev.id) }}
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
            @click="toggleAutoScroll"
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
              <th class="py-2.5 px-3 w-28">{{ $t('liveComm.timestamp') }} <span class="text-3xs text-blue-500 font-bold ml-0.5">↓</span></th>
              <th class="py-2.5 px-3 w-28">{{ $t('liveComm.device') }}</th>
              <th class="py-2.5 px-3 w-36">{{ $t('liveComm.parameter') }}</th>
              <th class="py-2.5 px-3 w-36">{{ $t('liveComm.function') }}</th>
              <th class="py-2.5 px-3 w-36">{{ $t('liveComm.address') }}</th>
              <th class="py-2.5 px-3">{{ $t('liveComm.hexPayload') }}</th>
              <th class="py-2.5 px-3 min-w-[150px]">{{ $t('liveComm.convertedValue') }}</th>
              <th class="py-2.5 px-3 w-20 text-right">{{ $t('liveComm.latency') }}</th>
              <th class="py-2.5 px-3 w-24 text-center">{{ $t('liveComm.status') }}</th>
              <th class="py-2.5 px-3 w-20 text-center">Inspect</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-[#E2ECF7] dark:divide-slate-800/60 font-mono text-3xs bg-white dark:bg-[#0B1524]">
            <tr v-if="filteredFrames.length === 0">
              <td colspan="10" class="py-16 text-center text-slate-500 font-sans">
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
              v-for="(frame, idx) in filteredFrames"
              :key="frame.id"
              @click="openInspectModal(frame)"
              class="hover:bg-blue-50/60 dark:hover:bg-slate-800/40 transition cursor-pointer"
              :class="{ 'bg-blue-50/20 dark:bg-blue-950/10': idx === 0 }"
            >
              <!-- Timestamp -->
              <td class="py-2 px-3 text-slate-600 dark:text-slate-400 whitespace-nowrap">
                <div class="flex items-center space-x-1.5">
                  <span>{{ formatTime(frame.timestamp) }}</span>
                  <span v-if="idx === 0" class="px-1 py-0.2 rounded text-4xs font-bold bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300">
                    NEW
                  </span>
                </div>
              </td>

              <!-- Device Code & Transport -->
              <td class="py-2 px-3">
                <div class="flex flex-col">
                  <span class="font-bold text-slate-900 dark:text-slate-100">{{ frame.device_code }}</span>
                  <span class="text-4xs text-slate-400 truncate max-w-[110px]" :title="frame.transport">{{ frame.transport }}</span>
                </div>
              </td>

              <!-- Target Parameter -->
              <td class="py-2 px-3">
                <div v-if="getFrameParams(frame)" class="flex flex-col">
                  <span class="font-bold text-slate-900 dark:text-slate-100 flex items-center gap-1">
                    <span class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold bg-indigo-50 text-indigo-700 border border-indigo-200 dark:bg-indigo-500/20 dark:text-indigo-300 dark:border-indigo-500/40">
                      {{ getFrameParams(frame).codes }}
                    </span>
                  </span>
                  <span v-if="getFrameParams(frame).name" class="text-4xs text-slate-500 dark:text-slate-400 truncate max-w-[130px] mt-0.5" :title="getFrameParams(frame).name">
                    {{ getFrameParams(frame).name }}
                  </span>
                </div>
                <div v-else class="text-slate-400 dark:text-slate-500 text-4xs italic">
                  --
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

              <!-- Converted Value (Sesuai Setup Parameter) -->
              <td class="py-2 px-3 whitespace-nowrap">
                <div v-if="getDecodedValues(frame).length > 0" class="space-y-1">
                  <div
                    v-for="(valItem, vIdx) in getDecodedValues(frame)"
                    :key="vIdx"
                    class="flex items-center space-x-1.5"
                  >
                    <span
                      v-if="getDecodedValues(frame).length > 1"
                      class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-slate-100 text-slate-700 border border-slate-300 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700"
                    >
                      {{ valItem.parameter_code }}
                    </span>
                    <span class="font-mono font-bold text-xs text-emerald-600 dark:text-emerald-400">
                      {{ valItem.formatted }}
                    </span>
                    <span v-if="valItem.unit" class="text-4xs font-sans text-slate-500 dark:text-slate-400 font-medium">
                      {{ valItem.unit }}
                    </span>
                  </div>
                </div>
                <div v-else-if="frame.converted_values" class="font-mono font-bold text-xs text-emerald-600 dark:text-emerald-400">
                  {{ frame.converted_values }}
                </div>
                <div v-else-if="frame.status !== 'SUCCESS'" class="text-rose-500 dark:text-rose-400 text-4xs italic">
                  --
                </div>
                <div v-else class="text-slate-400 dark:text-slate-500 text-4xs italic">
                  --
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

          <!-- Target Parameter Card -->
          <div v-if="getFrameParams(inspectedFrame)" class="p-3 rounded-xl bg-indigo-50/70 dark:bg-indigo-950/20 border border-indigo-200 dark:border-indigo-500/30 flex items-center justify-between">
            <div class="flex items-center space-x-2.5">
              <span class="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                </svg>
              </span>
              <div>
                <div class="text-4xs text-indigo-600 dark:text-indigo-400 uppercase font-semibold">{{ $t('liveComm.targetParameter') }}</div>
                <div class="font-bold text-xs text-indigo-950 dark:text-indigo-200">
                  {{ getFrameParams(inspectedFrame).codes }}
                  <span v-if="getFrameParams(inspectedFrame).name" class="font-normal text-slate-500 dark:text-slate-400 text-3xs ml-1">
                    ({{ getFrameParams(inspectedFrame).name }})
                  </span>
                </div>
              </div>
            </div>
            <span class="font-mono text-3xs font-semibold px-2 py-0.5 rounded bg-white dark:bg-slate-900 border border-indigo-200 dark:border-indigo-500/30 text-indigo-800 dark:text-indigo-300">
              #{{ inspectedFrame.display_address }} (Qty: {{ inspectedFrame.quantity }})
            </span>
          </div>

          <!-- Converted Value Inspection Panel in Modal -->
          <div v-if="getDecodedValues(inspectedFrame).length > 0" class="p-3.5 rounded-xl bg-emerald-50/70 dark:bg-emerald-950/20 border border-emerald-200 dark:border-emerald-500/30 space-y-2.5">
            <div class="flex items-center justify-between">
              <div class="text-4xs text-emerald-800 dark:text-emerald-300 uppercase font-bold flex items-center gap-1.5">
                <svg class="w-3.5 h-3.5 text-emerald-600 dark:text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
                </svg>
                <span>{{ $t('liveComm.convertedValue') }} (Sesuai Setup Parameter)</span>
              </div>
              <span class="text-4xs font-mono font-semibold px-2 py-0.5 rounded bg-emerald-100 dark:bg-emerald-900/40 text-emerald-800 dark:text-emerald-300">
                {{ getDecodedValues(inspectedFrame).length }} Parameter
              </span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <div
                v-for="(valItem, vIdx) in getDecodedValues(inspectedFrame)"
                :key="vIdx"
                class="p-2.5 rounded-lg bg-white dark:bg-slate-900 border border-emerald-200/80 dark:border-emerald-500/30 flex items-center justify-between"
              >
                <div>
                  <div class="font-bold text-xs text-slate-800 dark:text-slate-100 flex items-center gap-1.5">
                    <span>{{ valItem.parameter_code }}</span>
                    <span v-if="valItem.parameter_name && valItem.parameter_name !== valItem.parameter_code" class="text-4xs font-normal text-slate-500 truncate max-w-[120px]">
                      ({{ valItem.parameter_name }})
                    </span>
                  </div>
                  <div v-if="valItem.raw_value !== undefined" class="text-4xs text-slate-400 font-mono mt-0.5">
                    Raw: {{ valItem.raw_value }}
                  </div>
                </div>
                <div class="text-right pl-2">
                  <div class="font-mono font-bold text-sm text-emerald-600 dark:text-emerald-400">
                    {{ valItem.formatted }} <span v-if="valItem.unit" class="text-xs font-normal text-slate-500">{{ valItem.unit }}</span>
                  </div>
                </div>
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
    deviceList() {
      if (Array.isArray(this.devices) && this.devices.length > 0) {
        return this.devices;
      }
      if (Array.isArray(this.$store.state.devices) && this.$store.state.devices.length > 0) {
        return this.$store.state.devices;
      }
      return [];
    },
    countdownFormatted() {
      const sec = this.remainingSeconds > 0 ? this.remainingSeconds : (this.status ? this.status.remaining_seconds : 0);
      if (sec <= 0) return '00:00';
      const m = Math.floor(sec / 60);
      const s = sec % 60;
      return `${String(m).padStart(2, '0')}:${String(s).padStart(2, '0')} remaining`;
    },
    filteredFrames() {
      const list = this.frames.filter(f => {
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
          const matchVal = f.converted_values && f.converted_values.toLowerCase().includes(q);
          const pInfo = this.getFrameParams(f);
          const matchParam = pInfo && ((pInfo.codes && pInfo.codes.toLowerCase().includes(q)) || (pInfo.name && pInfo.name.toLowerCase().includes(q)) || (pInfo.primary && pInfo.primary.toLowerCase().includes(q)));
          if (!matchHex && !matchErr && !matchSummary && !matchParam && !matchVal) {
            return false;
          }
        }
        return true;
      });
      // Always keep newest frame at index 0 (top of table)
      return list.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
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
    getFrameParams(frame) {
      if (!frame) return null;

      // 1. If backend frame already has ParameterCodes / ParameterNames populated:
      if (frame.parameter_codes || frame.parameter_names) {
        const codes = frame.parameter_codes || frame.parameter_names;
        const name = frame.parameter_names && frame.parameter_names !== codes ? frame.parameter_names : '';
        return {
          codes,
          name,
          primary: name ? `${codes} (${name})` : codes,
        };
      }

      // 2. Client-side resolution fallback using loaded devices with parameters
      const dev = (this.deviceList || []).find(d => (frame.device_id && d.id === frame.device_id) || (frame.device_code && d.device_code === frame.device_code));
      if (!dev || !dev.parameters || dev.parameters.length === 0) {
        return null;
      }

      const matches = [];
      const reqStart = Number(frame.starting_address || 0);
      const reqQty = Number(frame.quantity || 1);
      const reqEnd = reqStart + reqQty - 1;

      for (const p of dev.parameters) {
        if (p.enabled === false) continue;
        const rawAddr = Number(p.register_address);
        let pduAddr = rawAddr;
        let fc = 3;
        const regType = String(p.register_type || 'HOLDING_REGISTER').toUpperCase();

        if (rawAddr >= 400001 && rawAddr <= 465536) {
          pduAddr = rawAddr - 400001;
          fc = 3;
        } else if (rawAddr >= 300001 && rawAddr <= 365536) {
          pduAddr = rawAddr - 300001;
          fc = 4;
        } else if (rawAddr >= 40001 && rawAddr <= 49999) {
          pduAddr = rawAddr - 40001;
          fc = 3;
        } else if (rawAddr >= 30001 && rawAddr <= 39999) {
          pduAddr = rawAddr - 30001;
          fc = 4;
        } else if (rawAddr >= 10001 && rawAddr <= 19999) {
          pduAddr = rawAddr - 10001;
          fc = 2;
        } else if (rawAddr >= 1 && rawAddr <= 9999 && regType.includes('COIL')) {
          pduAddr = rawAddr - 1;
          fc = 1;
        } else {
          if (regType.includes('INPUT')) fc = 4;
          else if (regType.includes('DISCRETE')) fc = 2;
          else if (regType.includes('COIL')) fc = 1;
          else fc = 3;
        }

        if (frame.function_code && Number(frame.function_code) !== fc) {
          continue;
        }

        let qty = 2;
        const dt = String(p.data_type || 'FLOAT32').toUpperCase();
        if (dt.includes('INT16') || dt.includes('UINT16')) qty = 1;
        else if (dt.includes('INT64') || dt.includes('FLOAT64') || dt.includes('DOUBLE')) qty = 4;

        const pEnd = pduAddr + qty - 1;
        if (reqStart <= pEnd && reqEnd >= pduAddr) {
          const code = p.parameter_code || p.code || p.parameter_name || p.name;
          const name = p.parameter_name || p.name;
          matches.push({ code, name });
        }
      }

      if (matches.length === 0) return null;
      const codes = matches.map(m => m.code).join(', ');
      const names = matches.map(m => m.name).filter(Boolean).join(', ');
      return {
        codes,
        name: names !== codes ? names : '',
        primary: names && names !== codes ? `${codes} (${names})` : codes,
      };
    },
    getDecodedValues(frame) {
      if (!frame) return [];
      // 1. Backend-provided structured decoded values
      if (Array.isArray(frame.decoded_values) && frame.decoded_values.length > 0) {
        return frame.decoded_values;
      }
      // 2. Converted values formatted string directly on frame
      if (frame.converted_values) {
        return [{
          parameter_code: frame.parameter_codes || '',
          parameter_name: frame.parameter_names || '',
          formatted: frame.converted_values,
          unit: '',
        }];
      }
      // 3. Fallback resolution via loaded devices list
      if (frame.status !== 'SUCCESS') return [];
      const dev = (this.deviceList || []).find(d => (frame.device_id && d.id === frame.device_id) || (frame.device_code && d.device_code === frame.device_code));
      if (!dev || !dev.parameters || dev.parameters.length === 0) {
        return [];
      }

      const matches = [];
      const reqStart = Number(frame.starting_address || 0);
      const reqQty = Number(frame.quantity || 1);
      const reqEnd = reqStart + reqQty - 1;

      for (const p of dev.parameters) {
        if (p.enabled === false) continue;
        const rawAddr = Number(p.register_address);
        let pduAddr = rawAddr;
        let fc = 3;
        const regType = String(p.register_type || 'HOLDING_REGISTER').toUpperCase();

        if (rawAddr >= 400001 && rawAddr <= 465536) {
          pduAddr = rawAddr - 400001;
          fc = 3;
        } else if (rawAddr >= 300001 && rawAddr <= 365536) {
          pduAddr = rawAddr - 300001;
          fc = 4;
        } else if (rawAddr >= 40001 && rawAddr <= 49999) {
          pduAddr = rawAddr - 40001;
          fc = 3;
        } else if (rawAddr >= 30001 && rawAddr <= 39999) {
          pduAddr = rawAddr - 30001;
          fc = 4;
        } else if (rawAddr >= 10001 && rawAddr <= 19999) {
          pduAddr = rawAddr - 10001;
          fc = 2;
        } else if (rawAddr >= 1 && rawAddr <= 9999 && regType.includes('COIL')) {
          pduAddr = rawAddr - 1;
          fc = 1;
        } else {
          if (regType.includes('INPUT')) fc = 4;
          else if (regType.includes('DISCRETE')) fc = 2;
          else if (regType.includes('COIL')) fc = 1;
          else fc = 3;
        }

        if (frame.function_code && Number(frame.function_code) !== fc) {
          continue;
        }

        let qty = 2;
        const dt = String(p.data_type || 'FLOAT32').toUpperCase();
        if (dt.includes('INT16') || dt.includes('UINT16')) qty = 1;
        else if (dt.includes('INT64') || dt.includes('FLOAT64') || dt.includes('DOUBLE')) qty = 4;

        const pEnd = pduAddr + qty - 1;
        if (reqStart <= pEnd && reqEnd >= pduAddr) {
          const code = p.parameter_code || p.code || p.parameter_name || p.name;
          const name = p.parameter_name || p.name;
          const val = (p.current_processed_value !== undefined && p.current_processed_value !== null)
            ? p.current_processed_value
            : ((p.current_value !== undefined && p.current_value !== null) ? p.current_value : null);
          if (val !== null) {
            const prec = (p.precision !== undefined && p.precision !== null) ? p.precision : 2;
            matches.push({
              parameter_code: code,
              parameter_name: name,
              formatted: Number(val).toFixed(prec),
              unit: p.unit || '',
              value: val,
            });
          }
        }
      }
      return matches;
    },
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
        this.frames = data.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
        this.scrollToTop();
      }
    },
    async loadDevices() {
      try {
        const res = await this.$store.dispatch('fetchDevices', { all: true, pageSize: 100 });
        let list = [];
        if (Array.isArray(res)) {
          list = res;
        } else if (res && Array.isArray(res.items)) {
          list = res.items;
        } else if (res && res.data && Array.isArray(res.data.items)) {
          list = res.data.items;
        } else if (res && Array.isArray(res.data)) {
          list = res.data;
        } else if (Array.isArray(this.$store.state.devices)) {
          list = this.$store.state.devices;
        }
        this.devices = list;
      } catch (e) {
        if (Array.isArray(this.$store.state.devices)) {
          this.devices = this.$store.state.devices;
        }
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
        // Prepend newest frame to top of list
        this.frames.unshift(event.detail);
        if (this.frames.length > this.maxInMemoryFrames) {
          this.frames.pop();
        }
        this.scrollToTop();
      }
    },
    onTraceStatus(event) {
      if (event && event.detail) {
        this.status = event.detail;
        this.remainingSeconds = event.detail.remaining_seconds || 0;
      }
    },
    toggleAutoScroll() {
      this.autoScroll = !this.autoScroll;
      if (this.autoScroll) {
        this.scrollToTop();
      }
    },
    scrollToTop() {
      if (!this.autoScroll) return;
      this.$nextTick(() => {
        const c = this.$refs.packetContainer;
        if (c) {
          c.scrollTop = 0;
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
        const enriched = data.map(f => {
          const pInfo = this.getFrameParams(f);
          return {
            ...f,
            target_parameter: pInfo ? pInfo.primary : (f.parameter_names || null),
            target_parameter_code: pInfo ? pInfo.codes : (f.parameter_codes || null),
          };
        });
        content = JSON.stringify(enriched, null, 2);
        mime = 'application/json';
        filename += '.json';
      } else {
        filename += '.txt';
        content = data.map(f => {
          const pInfo = this.getFrameParams(f);
          const pStr = pInfo ? ` [Param: ${pInfo.primary}]` : '';
          const valStr = f.converted_values ? ` [Val: ${f.converted_values}]` : '';
          return `[${f.timestamp}] [${f.device_code}]${pStr}${valStr} ${f.fc_name} Addr:#${f.display_address} Status:${f.status} Latency:${f.latency_ms}ms\nTX: ${f.tx_hex}\nRX: ${f.rx_hex || f.error_message}\n`;
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
