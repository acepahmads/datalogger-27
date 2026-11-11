<template>
  <div class="sysmon-root space-y-3 font-sans select-none" :class="{ 'sysmon-dark': isDarkMode }">
    <!-- TIER 1: Compact Command Header Strip -->
    <div class="flex flex-col lg:flex-row lg:items-center lg:justify-between gap-2.5">
      <div class="min-w-0">
        <div class="flex flex-wrap items-center gap-2">
          <h1 class="text-lg sm:text-xl font-bold tracking-tight text-slate-900 dark:text-white leading-tight">
            {{ $t('systemMonitoring.hostServiceTitle') }}
          </h1>
          <span
            class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-mono font-bold tracking-wider uppercase border"
            :class="isCloudMode
              ? 'bg-indigo-50 text-indigo-700 border-indigo-200 dark:bg-indigo-500/15 dark:text-indigo-300 dark:border-indigo-500/30'
              : 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-500/15 dark:text-emerald-300 dark:border-emerald-500/30'"
          >
            {{ isCloudMode ? $t('systemMonitoring.cloudReceiverBadge') : $t('systemMonitoring.edgeAutonomousBadge') }}
          </span>
          <span class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-mono font-medium bg-slate-100 text-slate-600 border border-slate-200/80 dark:bg-slate-800/80 dark:text-slate-300 dark:border-slate-700/70">
            {{ osArchLabel }} &bull; {{ status.num_cpu || 4 }} vCPU
          </span>
          <span
            v-if="displayHostname"
            class="hidden sm:inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-mono font-medium bg-slate-100 text-slate-600 border border-slate-200/80 dark:bg-slate-800/80 dark:text-slate-400 dark:border-slate-700/70"
          >
            {{ displayHostname }}
          </span>
        </div>
        <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5 truncate">
          {{ $t('systemMonitoring.hostServiceSubtitle') }}
        </p>
      </div>

      <!-- Right Vital Pills & Controls -->
      <div class="flex flex-wrap items-center gap-1.5 flex-shrink-0">
        <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-[11px] font-mono font-medium bg-white dark:bg-slate-900/90 text-slate-700 dark:text-slate-200 border border-slate-200/90 dark:border-slate-800 shadow-2xs">
          <svg class="w-3.5 h-3.5 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <circle cx="12" cy="12" r="9" />
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 7v5l3 2" />
          </svg>
          <span>{{ $t('systemMonitoring.uptime') }} <strong>{{ status.uptime_human || '0m' }}</strong></span>
        </span>

        <span
          v-if="!isOperator"
          class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-[11px] font-mono font-medium bg-white dark:bg-slate-900/90 text-slate-700 dark:text-slate-200 border border-slate-200/90 dark:border-slate-800 shadow-2xs"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
          <span>MariaDB &bull; <strong class="text-blue-600 dark:text-sky-400">{{ formattedDbLatency }}</strong></span>
        </span>

        <button
          type="button"
          @click="cyclePollInterval"
          class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-[11px] font-mono font-bold border transition-colors cursor-pointer"
          :class="pollIntervalSec > 0
            ? 'bg-emerald-50 text-emerald-700 border-emerald-200 hover:bg-emerald-100 dark:bg-emerald-500/15 dark:text-emerald-300 dark:border-emerald-500/30 dark:hover:bg-emerald-500/25'
            : 'bg-amber-50 text-amber-700 border-amber-200 hover:bg-amber-100 dark:bg-amber-500/15 dark:text-amber-300 dark:border-amber-500/30'"
          title="Click to cycle auto-refresh interval (5s / 10s / 30s / Paused)"
        >
          <span
            class="w-2 h-2 rounded-full"
            :class="pollIntervalSec > 0 ? 'bg-emerald-500 animate-pulse' : 'bg-amber-500'"
          ></span>
          <span>{{ pollIntervalSec > 0 ? $t('systemMonitoring.liveInterval', { sec: pollIntervalSec }) : $t('systemMonitoring.paused') }}</span>
        </button>

        <button
          type="button"
          @click="manualRefresh"
          :disabled="isRefreshing"
          class="inline-flex items-center gap-1.5 px-3 py-1 rounded-lg text-[11px] font-semibold bg-blue-600 hover:bg-blue-700 active:bg-blue-800 text-white shadow-xs transition-colors cursor-pointer disabled:opacity-60"
        >
          <svg
            class="w-3.5 h-3.5"
            :class="{ 'animate-spin': isRefreshing }"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            stroke-width="2.2"
          >
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>{{ $t('systemMonitoring.refresh') }}</span>
        </button>
      </div>
    </div>

    <!-- TIER 2: 4 Hardware & Network Gauge Bento Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3">
      <!-- 1. CPU Compute & Process Load -->
      <div class="saas-card p-3.5 flex flex-col justify-between space-y-2.5">
        <div class="flex items-center justify-between gap-2">
          <span class="text-xs font-bold text-slate-800 dark:text-slate-100 truncate">
            {{ $t('systemMonitoring.cpuCardTitle') }}
          </span>
          <span
            class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[10px] font-mono font-semibold bg-blue-50 text-blue-700 border border-blue-200/80 dark:bg-blue-500/15 dark:text-sky-300 dark:border-blue-500/30 flex-shrink-0"
            :title="$t('systemMonitoring.appCpuTooltip')"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-blue-500 dark:bg-sky-400"></span>
            <span>{{ $t('systemMonitoring.app') }} {{ processCpuFixed }}%</span>
          </span>
        </div>

        <div class="flex items-center gap-3">
          <!-- Radial Ring Gauge -->
          <div class="relative w-[72px] h-[72px] flex-shrink-0">
            <svg viewBox="0 0 80 80" class="w-full h-full -rotate-90">
              <circle
                cx="40"
                cy="40"
                r="32"
                fill="none"
                stroke-width="7"
                class="stroke-slate-200/90 dark:stroke-slate-800"
              />
              <circle
                cx="40"
                cy="40"
                r="32"
                fill="none"
                stroke-width="7"
                stroke-linecap="round"
                :stroke="cpuStrokeColor"
                :stroke-dasharray="ringCircumference"
                :stroke-dashoffset="ringDashOffset(status.cpu_percent)"
                class="transition-all duration-500 sysmon-ring-glow"
              />
            </svg>
            <div class="absolute inset-0 flex flex-col items-center justify-center leading-none">
              <span class="text-[13px] font-mono font-extrabold text-slate-900 dark:text-white tabular-nums">
                {{ cpuPercentFixed }}%
              </span>
              <span class="text-[9px] font-mono text-slate-400 mt-0.5">HOST</span>
            </div>
          </div>

          <!-- Mini 60s Sparkline + Per-Core Bars -->
          <div class="flex-1 min-w-0 space-y-1.5">
            <div class="relative h-9 w-full">
              <svg viewBox="0 0 180 38" preserveAspectRatio="none" class="w-full h-full overflow-visible">
                <defs>
                  <linearGradient id="sysmonCpuGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="#3B82F6" stop-opacity="0.34" />
                    <stop offset="100%" stop-color="#3B82F6" stop-opacity="0.02" />
                  </linearGradient>
                </defs>
                <path :d="buildSparklineArea(cpuHistory, 180, 38, 100)" fill="url(#sysmonCpuGrad)" />
                <path :d="buildSparklinePath(cpuHistory, 180, 38, 100)" fill="none" stroke="#3B82F6" stroke-width="1.75" stroke-linecap="round" />
              </svg>
              <span class="absolute right-0 bottom-0 text-[9px] font-mono text-slate-400 bg-white/80 dark:bg-slate-900/80 px-1 rounded">60s</span>
            </div>

            <div class="grid grid-cols-2 gap-x-2 gap-y-1">
              <div v-for="(c, idx) in visibleCores" :key="'core-' + idx" class="text-[10px] font-mono">
                <div class="flex items-center justify-between leading-none mb-0.5">
                  <span class="text-slate-500 dark:text-slate-400">C{{ idx }}</span>
                  <span :class="c > 80 ? 'text-rose-600 dark:text-rose-400 font-bold' : 'text-slate-700 dark:text-slate-300 font-semibold'" class="tabular-nums">{{ c.toFixed(0) }}%</span>
                </div>
                <div class="w-full h-1 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all duration-300"
                    :class="c > 80 ? 'bg-rose-500' : c > 65 ? 'bg-amber-500' : 'bg-sky-500 dark:bg-sky-400'"
                    :style="{ width: clampPercent(c) + '%' }"
                  ></div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800/80 text-[10.5px] text-slate-500 dark:text-slate-400">
          <span>{{ $t('systemMonitoring.logicalCores', { count: status.num_cpu || 4 }) }} &bull; <strong class="font-mono text-slate-700 dark:text-slate-300">{{ status.arch || 'amd64' }}</strong></span>
          <span class="font-mono text-[10px] text-slate-500 dark:text-slate-400 tabular-nums">
            {{ status.goroutines || status.go_num_goroutine || 0 }} {{ $t('systemMonitoring.goroutines') }}
          </span>
        </div>
      </div>

      <!-- 2. Memory Allocation (htop) -->
      <div class="saas-card p-3.5 flex flex-col justify-between space-y-2.5">
        <div class="flex items-center justify-between gap-2">
          <span class="text-xs font-bold text-slate-800 dark:text-slate-100 truncate">
            {{ $t('systemMonitoring.memoryCardTitle') }}
          </span>
          <span
            v-if="status.process_rss_mb"
            class="inline-flex items-center px-1.5 py-0.5 rounded-md text-[10px] font-mono font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200/80 dark:bg-emerald-500/15 dark:text-emerald-300 dark:border-emerald-500/30 flex-shrink-0 tabular-nums"
          >
            {{ $t('systemMonitoring.processRss') }}: {{ Number(status.process_rss_mb).toFixed(0) }} MB
          </span>
        </div>

        <div class="flex items-center gap-3">
          <div class="relative w-[72px] h-[72px] flex-shrink-0">
            <svg viewBox="0 0 80 80" class="w-full h-full -rotate-90">
              <circle cx="40" cy="40" r="32" fill="none" stroke-width="7" class="stroke-slate-200/90 dark:stroke-slate-800" />
              <circle
                cx="40"
                cy="40"
                r="32"
                fill="none"
                stroke-width="7"
                stroke-linecap="round"
                :stroke="ramStrokeColor"
                :stroke-dasharray="ringCircumference"
                :stroke-dashoffset="ringDashOffset(status.ram_percent)"
                class="transition-all duration-500 sysmon-ring-glow"
              />
            </svg>
            <div class="absolute inset-0 flex flex-col items-center justify-center leading-none">
              <span class="text-[13px] font-mono font-extrabold text-slate-900 dark:text-white tabular-nums">
                {{ ramPercentFixed }}%
              </span>
              <span class="text-[9px] font-mono text-slate-400 mt-0.5">RAM</span>
            </div>
          </div>

          <div class="flex-1 min-w-0 space-y-2">
            <div class="relative h-9 w-full">
              <svg viewBox="0 0 180 38" preserveAspectRatio="none" class="w-full h-full overflow-visible">
                <defs>
                  <linearGradient id="sysmonRamGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="#10B981" stop-opacity="0.34" />
                    <stop offset="100%" stop-color="#10B981" stop-opacity="0.02" />
                  </linearGradient>
                </defs>
                <path :d="buildSparklineArea(ramHistory, 180, 38, 100)" fill="url(#sysmonRamGrad)" />
                <path :d="buildSparklinePath(ramHistory, 180, 38, 100)" fill="none" stroke="#10B981" stroke-width="1.75" stroke-linecap="round" />
              </svg>
            </div>

            <div>
              <div class="w-full h-2 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-300"
                  :class="status.ram_percent > 90 ? 'bg-rose-500' : status.ram_percent > 75 ? 'bg-amber-500' : 'bg-emerald-500'"
                  :style="{ width: clampPercent(status.ram_percent) + '%' }"
                ></div>
              </div>
              <div class="flex items-center justify-between text-[10px] font-mono mt-1 tabular-nums">
                <span class="font-semibold text-slate-800 dark:text-slate-200">{{ ramUsedLabel }} {{ $t('systemMonitoring.used') }}</span>
                <span class="text-slate-500 dark:text-slate-400">{{ ramTotalLabel }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800/80 text-[10.5px]">
          <span
            class="inline-flex items-center px-2 py-0.5 rounded-md text-[9.5px] font-mono font-bold tracking-wide border"
            :class="memoryPressureBadgeClass"
          >
            {{ memoryPressureLabel }}
          </span>
          <span class="font-mono text-[10.5px] text-slate-500 dark:text-slate-400 tabular-nums">
            {{ ramFreeLabel }} {{ $t('systemMonitoring.free') }}
          </span>
        </div>
      </div>

      <!-- 3. Storage & Retention Volume -->
      <div class="saas-card p-3.5 flex flex-col justify-between space-y-2.5">
        <div class="flex items-center justify-between gap-2">
          <span class="text-xs font-bold text-slate-800 dark:text-slate-100 truncate">
            {{ $t('systemMonitoring.storageCardTitle') }}
          </span>
          <span class="text-[10px] font-mono font-semibold text-indigo-600 dark:text-indigo-400 tabular-nums flex-shrink-0">
            {{ diskFreeLabel }} {{ $t('systemMonitoring.free') }}
          </span>
        </div>

        <div class="flex items-center gap-3">
          <div class="relative w-[72px] h-[72px] flex-shrink-0">
            <svg viewBox="0 0 80 80" class="w-full h-full -rotate-90">
              <circle cx="40" cy="40" r="32" fill="none" stroke-width="7" class="stroke-slate-200/90 dark:stroke-slate-800" />
              <circle
                cx="40"
                cy="40"
                r="32"
                fill="none"
                stroke-width="7"
                stroke-linecap="round"
                :stroke="diskStrokeColor"
                :stroke-dasharray="ringCircumference"
                :stroke-dashoffset="ringDashOffset(status.disk_percent)"
                class="transition-all duration-500 sysmon-ring-glow"
              />
            </svg>
            <div class="absolute inset-0 flex flex-col items-center justify-center leading-none">
              <span class="text-[13px] font-mono font-extrabold text-slate-900 dark:text-white tabular-nums">
                {{ diskPercentFixed }}%
              </span>
              <span class="text-[9px] font-mono text-slate-400 mt-0.5">DISK</span>
            </div>
          </div>

          <div class="flex-1 min-w-0 space-y-2">
            <div>
              <div class="flex items-center justify-between text-[10px] text-slate-500 dark:text-slate-400 mb-1">
                <span>{{ $t('systemMonitoring.capacity') }}</span>
                <span class="font-mono font-semibold text-slate-700 dark:text-slate-300 tabular-nums">{{ diskUsedLabel }} / {{ diskTotalLabel }}</span>
              </div>
              <div class="w-full h-2.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-300"
                  :class="status.disk_percent > 90 ? 'bg-rose-500' : status.disk_percent > 80 ? 'bg-amber-500' : 'bg-indigo-500'"
                  :style="{ width: clampPercent(status.disk_percent) + '%' }"
                ></div>
              </div>
            </div>

            <div class="flex flex-wrap gap-1">
              <span class="inline-flex items-center px-1.5 py-0.5 rounded text-[9.5px] font-mono font-semibold bg-emerald-50 text-emerald-700 border border-emerald-200/80 dark:bg-emerald-500/15 dark:text-emerald-300 dark:border-emerald-500/30">
                {{ $t('systemMonitoring.walQueue') }}: {{ shortQueueStatus }}
              </span>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800/80 text-[10.5px] text-slate-500 dark:text-slate-400">
          <span class="font-mono">
            {{ $t('systemMonitoring.rawRetention') }}: <strong class="text-slate-700 dark:text-slate-200">{{ retentionDaysLabel }}</strong>
          </span>
          <span class="font-mono text-[10px] tabular-nums">
            Spool: {{ formatTotalBytes(status.queue_spool_size_bytes || 0) }}
          </span>
        </div>
      </div>

      <!-- 4. Network & Connectivity (Jaringan: LAN, WAN & Live Traffic) -->
      <div class="saas-card p-3.5 flex flex-col justify-between space-y-2">
        <div class="flex items-center justify-between gap-1.5">
          <span class="text-xs font-bold text-slate-800 dark:text-slate-100 truncate">
            {{ $t('systemMonitoring.networkCardTitle') }}
          </span>
          <div class="flex items-center gap-1 flex-shrink-0">
            <span
              class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-[9.5px] font-mono font-bold border"
              :class="networkBadgeClass"
            >
              <span class="w-1.5 h-1.5 rounded-full" :class="networkDotClass"></span>
              <span>{{ networkOverallLabel }}</span>
            </span>
            <button
              type="button"
              @click="showNetworkModal = true"
              class="px-1.5 py-0.5 rounded-md text-[10px] font-semibold bg-slate-100 hover:bg-sky-50 text-slate-600 hover:text-sky-700 border border-slate-200/80 dark:bg-slate-800 dark:hover:bg-sky-500/15 dark:text-slate-300 dark:hover:text-sky-300 dark:border-slate-700 transition-colors cursor-pointer"
            >
              {{ $t('systemMonitoring.networkDetailBtn') }}
            </button>
          </div>
        </div>

        <!-- Live RX/TX Sparkline + Counters -->
        <div class="flex items-center gap-2.5">
          <div class="relative w-[64px] h-[64px] flex-shrink-0">
            <svg viewBox="0 0 80 80" class="w-full h-full -rotate-90">
              <circle cx="40" cy="40" r="32" fill="none" stroke-width="7" class="stroke-slate-200/90 dark:stroke-slate-800" />
              <circle
                cx="40"
                cy="40"
                r="32"
                fill="none"
                stroke-width="7"
                stroke-linecap="round"
                :stroke="networkStrokeColor"
                :stroke-dasharray="ringCircumference"
                :stroke-dashoffset="ringDashOffset(networkHealthScore)"
                class="transition-all duration-500 sysmon-ring-glow"
              />
            </svg>
            <div class="absolute inset-0 flex flex-col items-center justify-center leading-none text-center px-1">
              <span class="text-[10.5px] font-mono font-extrabold text-slate-900 dark:text-white tabular-nums truncate max-w-full">
                {{ netCombinedRateCompact }}
              </span>
              <span class="text-[8.5px] font-mono text-cyan-600 dark:text-cyan-400 mt-0.5">RX+TX</span>
            </div>
          </div>

          <div class="flex-1 min-w-0 space-y-1">
            <div class="h-7 w-full">
              <svg viewBox="0 0 180 28" preserveAspectRatio="none" class="w-full h-full overflow-visible">
                <defs>
                  <linearGradient id="sysmonNetGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="#06B6D4" stop-opacity="0.34" />
                    <stop offset="100%" stop-color="#06B6D4" stop-opacity="0.02" />
                  </linearGradient>
                </defs>
                <path :d="buildSparklineArea(netHistory, 180, 28, netHistoryPeak)" fill="url(#sysmonNetGrad)" />
                <path :d="buildSparklinePath(netHistory, 180, 28, netHistoryPeak)" fill="none" stroke="#06B6D4" stroke-width="1.75" stroke-linecap="round" />
              </svg>
            </div>
            <div class="grid grid-cols-2 gap-1 text-[10px] font-mono tabular-nums">
              <div class="px-1.5 py-0.5 rounded bg-sky-50/80 dark:bg-sky-500/10 text-sky-700 dark:text-sky-300 truncate" :title="'Total RX: ' + formatTotalBytes(status.net_rx_total_bytes)">
                &darr; {{ formatByteRate(status.net_rx_bytes_per_sec) }}
              </div>
              <div class="px-1.5 py-0.5 rounded bg-indigo-50/80 dark:bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 truncate" :title="'Total TX: ' + formatTotalBytes(status.net_tx_total_bytes)">
                &uarr; {{ formatByteRate(status.net_tx_bytes_per_sec) }}
              </div>
            </div>
          </div>
        </div>

        <!-- Compact LAN / Gateway / WAN Summary -->
        <div class="pt-1.5 border-t border-slate-100 dark:border-slate-800/80 space-y-1 text-[10.5px]">
          <div class="flex items-center justify-between gap-2">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.lanStatus') }}</span>
            <span class="font-mono font-semibold text-slate-800 dark:text-slate-200 truncate">
              {{ displayPrimaryIp }}
              <span class="text-[9.5px] text-emerald-600 dark:text-emerald-400">({{ $t('systemMonitoring.interfacesUp', { count: activeNicCount }) }})</span>
            </span>
          </div>
          <div class="flex items-center justify-between gap-2">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.gateway') }}</span>
            <span class="font-mono truncate" :class="networkData.local_network.gateway_reachable ? 'text-emerald-600 dark:text-emerald-400 font-semibold' : 'text-slate-500 dark:text-slate-400'">
              {{ displayGateway }}
            </span>
          </div>
          <div class="flex items-center justify-between gap-2">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.internetWan') }}</span>
            <span class="font-mono font-semibold truncate" :class="wanTextClass">
              {{ wanSummaryText }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- TIER 3: 4 Subsystem Diagnostic Matrix Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3">
      <!-- A. Engine & Services -->
      <div class="saas-card p-3.5 space-y-2 text-xs">
        <div class="flex items-center justify-between pb-1.5 border-b border-slate-100 dark:border-slate-800/80">
          <span class="text-[11px] font-bold uppercase tracking-wider text-blue-600 dark:text-sky-400">
            {{ $t('systemMonitoring.engineServicesTitle') }}
          </span>
          <span class="text-[10px] font-mono text-slate-400">{{ status.app_version || 'v1.0.0' }}</span>
        </div>

        <div class="space-y-1.5 text-[11.5px]">
          <div class="flex items-center justify-between">
            <span class="text-slate-600 dark:text-slate-400">{{ $t('systemMonitoring.coreEngine') }}</span>
            <span class="sysmon-pill sysmon-pill-emerald">{{ status.service_status || 'RUNNING' }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-600 dark:text-slate-400">{{ $t('systemMonitoring.ingestionWorkers') }}</span>
            <span class="sysmon-pill sysmon-pill-emerald">{{ isCloudMode ? 'PASSIVE PUSH' : (status.worker_status || 'READY') }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-600 dark:text-slate-400">{{ $t('systemMonitoring.rollupScheduler') }}</span>
            <span class="sysmon-pill sysmon-pill-emerald">{{ status.scheduler_status || 'ACTIVE' }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-600 dark:text-slate-400">{{ $t('systemMonitoring.outboundWalQueue') }}</span>
            <span class="sysmon-pill" :class="shortQueueStatus === 'HEALTHY' || shortQueueStatus === 'OPTIMAL' ? 'sysmon-pill-emerald' : 'sysmon-pill-amber'">
              {{ shortQueueStatus }}
            </span>
          </div>
          <div v-if="!isOperator" class="flex items-center justify-between">
            <span class="text-slate-600 dark:text-slate-400">{{ $t('systemMonitoring.databasePool') }}</span>
            <span class="sysmon-pill sysmon-pill-emerald">ONLINE</span>
          </div>
        </div>
      </div>

      <!-- B. Station & Device Fleet -->
      <div class="saas-card p-3.5 flex flex-col justify-between space-y-2 text-xs">
        <div class="flex items-center justify-between pb-1.5 border-b border-slate-100 dark:border-slate-800/80">
          <span class="text-[11px] font-bold uppercase tracking-wider text-emerald-600 dark:text-emerald-400">
            {{ $t('systemMonitoring.stationFleetTitle') }}
          </span>
          <span class="text-[10px] font-mono font-semibold text-emerald-600 dark:text-emerald-400 tabular-nums">
            {{ fleetOnlinePercent }}% Online
          </span>
        </div>

        <!-- Stacked Fleet Bar -->
        <div class="w-full h-2 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden flex">
          <div class="h-full bg-emerald-500 transition-all duration-300" :style="{ width: fleetBarSegments.online + '%' }"></div>
          <div class="h-full bg-amber-500 transition-all duration-300" :style="{ width: fleetBarSegments.offline + '%' }"></div>
          <div class="h-full bg-rose-500 transition-all duration-300" :style="{ width: fleetBarSegments.error + '%' }"></div>
        </div>

        <!-- 2x2 Mini KPI Counter Grid -->
        <div class="grid grid-cols-2 gap-1.5">
          <div class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200/70 dark:border-slate-800 text-center">
            <div class="text-sm font-mono font-extrabold text-slate-900 dark:text-white leading-tight tabular-nums">{{ status.total_devices || 0 }}</div>
            <div class="text-[9.5px] text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.configuredShort') }}</div>
          </div>
          <div class="px-2.5 py-1.5 rounded-lg bg-emerald-50/70 dark:bg-emerald-500/10 border border-emerald-200/70 dark:border-emerald-500/20 text-center">
            <div class="text-sm font-mono font-extrabold text-emerald-700 dark:text-emerald-400 leading-tight tabular-nums">{{ status.online_devices || 0 }}</div>
            <div class="text-[9.5px] text-emerald-700/80 dark:text-emerald-400/80">{{ $t('systemMonitoring.onlineShort') }}</div>
          </div>
          <div class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200/70 dark:border-slate-800 text-center">
            <div class="text-sm font-mono font-bold text-slate-700 dark:text-slate-300 leading-tight tabular-nums">{{ status.offline_devices || 0 }}</div>
            <div class="text-[9.5px] text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.offlineShort') }}</div>
          </div>
          <div class="px-2.5 py-1.5 rounded-lg bg-rose-50/60 dark:bg-rose-500/10 border border-rose-200/60 dark:border-rose-500/20 text-center">
            <div class="text-sm font-mono font-bold text-rose-600 dark:text-rose-400 leading-tight tabular-nums">{{ status.error_devices || 0 }}</div>
            <div class="text-[9.5px] text-rose-600/80 dark:text-rose-400/80">{{ $t('systemMonitoring.errorShort') }}</div>
          </div>
        </div>

        <div class="flex items-center justify-between text-[10.5px] pt-1 border-t border-slate-100 dark:border-slate-800/80">
          <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.lastPacket') }}:</span>
          <span class="font-mono font-semibold text-blue-600 dark:text-sky-400 tabular-nums">{{ status.last_device_comm || '--' }}</span>
        </div>
      </div>

      <!-- C. Telemetry Ingestion -->
      <div class="saas-card p-3.5 flex flex-col justify-between space-y-2 text-xs">
        <div class="flex items-center justify-between pb-1.5 border-b border-slate-100 dark:border-slate-800/80">
          <span class="text-[11px] font-bold uppercase tracking-wider text-amber-600 dark:text-amber-400">
            {{ $t('systemMonitoring.telemetryIngestionTitle') }}
          </span>
          <span class="px-1.5 py-0.5 rounded-md text-[10px] font-mono font-bold bg-blue-50 text-blue-700 border border-blue-200/80 dark:bg-blue-500/15 dark:text-sky-300 dark:border-blue-500/30 tabular-nums">
            {{ Number(status.data_per_sec || 0).toFixed(1) }} /sec
          </span>
        </div>

        <div>
          <div class="flex items-baseline gap-1.5">
            <span class="text-xl font-mono font-extrabold text-slate-900 dark:text-white tracking-tight tabular-nums">
              {{ Number(status.data_received_count || 0).toLocaleString() }}
            </span>
            <span class="text-[11px] font-medium text-slate-500 dark:text-slate-400">
              {{ $t('systemMonitoring.samplesUnit') }}
            </span>
          </div>

          <div class="w-full h-2 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden mt-2">
            <div class="h-full rounded-full bg-emerald-500 transition-all duration-300" :style="{ width: qualityPercent + '%' }"></div>
          </div>
        </div>

        <div class="space-y-1 text-[11px] pt-1 border-t border-slate-100 dark:border-slate-800/80">
          <div class="flex items-center justify-between">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.quality') }}</span>
            <span class="font-mono font-bold text-emerald-600 dark:text-emerald-400">{{ status.data_quality || '100.0% GOOD' }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.invalid') }}</span>
            <span class="font-mono font-semibold text-slate-700 dark:text-slate-300 tabular-nums">{{ status.invalid_data_count || 0 }}</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.lastSample') }}</span>
            <span class="font-mono font-semibold text-slate-800 dark:text-slate-200 tabular-nums">{{ status.last_data_received || '--' }}</span>
          </div>
        </div>
      </div>

      <!-- D. Communication, Outbound & Alarms -->
      <div class="saas-card p-3.5 flex flex-col justify-between space-y-2 text-xs">
        <div class="flex items-center justify-between pb-1.5 border-b border-slate-100 dark:border-slate-800/80">
          <span class="text-[11px] font-bold uppercase tracking-wider text-purple-600 dark:text-purple-400">
            {{ $t('systemMonitoring.communicationApiTitle') }}
          </span>
          <span class="text-[10px] font-mono font-bold text-emerald-600 dark:text-emerald-400 tabular-nums">
            {{ Number(status.success_rate || 0).toFixed(1) }}% OK
          </span>
        </div>

        <div>
          <div class="flex items-center justify-between text-[11px] mb-1">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.successRate') }}</span>
            <span class="font-mono font-bold text-slate-900 dark:text-white tabular-nums">
              {{ Number(status.success_rate || 0).toFixed(1) }}%
              <span class="text-[10px] font-normal text-slate-400">({{ Number(status.failed_rate || 0).toFixed(1) }}% err)</span>
            </span>
          </div>
          <div class="w-full h-2 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-300"
              :class="(status.success_rate || 0) >= 95 ? 'bg-sky-500 dark:bg-sky-400' : 'bg-amber-500'"
              :style="{ width: clampPercent(status.success_rate || 0) + '%' }"
            ></div>
          </div>
        </div>

        <div class="space-y-1 text-[11px] pt-1 border-t border-slate-100 dark:border-slate-800/80">
          <div class="flex items-center justify-between">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.avgLatency') }}:</span>
            <span class="font-mono font-bold text-blue-600 dark:text-sky-400 tabular-nums">{{ status.avg_latency_ms || 0 }} ms</span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.activeChannels') }}:</span>
            <span class="font-mono font-semibold text-slate-800 dark:text-slate-200 tabular-nums">
              {{ status.total_connections || 0 }} {{ $t('systemMonitoring.active') }}
              <span v-if="outboundTotalLinks > 0" class="text-[10px] text-emerald-600 dark:text-emerald-400">
                &bull; Out {{ outboundHealthyLinks }}/{{ outboundTotalLinks }}
              </span>
            </span>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.activeAlarms') }}:</span>
            <span
              class="font-mono font-bold tabular-nums"
              :class="(status.critical_alarms || 0) > 0 ? 'text-rose-600 dark:text-rose-400' : (status.warning_alarms || 0) > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-slate-700 dark:text-slate-300'"
            >
              {{ status.critical_alarms || 0 }} {{ $t('systemMonitoring.criticalShort') }}
              <span v-if="status.warning_alarms" class="text-amber-600 dark:text-amber-400"> &bull; {{ status.warning_alarms }} {{ $t('systemMonitoring.warningShort') }}</span>
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- TIER 4: Live 5-Minute System Resource, Network & Ingestion Timeline -->
    <div class="saas-card p-3.5">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 mb-2">
        <div class="flex items-center gap-2">
          <h2 class="text-xs font-bold text-slate-900 dark:text-white">
            {{ $t('systemMonitoring.timelineTitle') }}
          </h2>
        </div>

        <!-- Interactive Series Toggle Pills -->
        <div class="flex flex-wrap items-center gap-1.5">
          <button
            v-for="s in timelineSeriesDefs"
            :key="s.key"
            type="button"
            @click="toggleTimelineSeries(s.key)"
            class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10.5px] font-medium border transition-all cursor-pointer"
            :class="visibleSeries[s.key]
              ? 'bg-slate-100/90 dark:bg-slate-800 text-slate-800 dark:text-slate-100 border-slate-300 dark:border-slate-700 shadow-2xs'
              : 'bg-transparent text-slate-400 dark:text-slate-500 border-slate-200/60 dark:border-slate-800 opacity-55'"
          >
            <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: s.color }"></span>
            <span>{{ s.label }}</span>
          </button>
        </div>
      </div>

      <!-- Interactive Multi-Series SVG Timeline -->
      <div
        class="relative w-full h-[126px]"
        @mousemove="onTimelineMouseMove"
        @mouseleave="hoverIndex = -1"
      >
        <svg viewBox="0 0 960 112" preserveAspectRatio="none" class="w-full h-[104px] overflow-visible">
          <defs>
            <linearGradient id="tlGradCpu" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#3B82F6" stop-opacity="0.24" />
              <stop offset="100%" stop-color="#3B82F6" stop-opacity="0.0" />
            </linearGradient>
            <linearGradient id="tlGradRam" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#10B981" stop-opacity="0.20" />
              <stop offset="100%" stop-color="#10B981" stop-opacity="0.0" />
            </linearGradient>
            <linearGradient id="tlGradNet" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#06B6D4" stop-opacity="0.22" />
              <stop offset="100%" stop-color="#06B6D4" stop-opacity="0.0" />
            </linearGradient>
          </defs>

          <!-- Subtle Horizontal Gridlines -->
          <line x1="0" y1="12" x2="960" y2="12" stroke="currentColor" stroke-dasharray="3 3" class="text-slate-200/80 dark:text-slate-800/90" stroke-width="1" />
          <line x1="0" y1="56" x2="960" y2="56" stroke="currentColor" stroke-dasharray="3 3" class="text-slate-200/80 dark:text-slate-800/90" stroke-width="1" />
          <line x1="0" y1="100" x2="960" y2="100" stroke="currentColor" class="text-slate-200 dark:text-slate-800" stroke-width="1" />

          <!-- Series Area & Smooth Lines -->
          <g v-if="visibleSeries.ram">
            <path :d="buildTimelineArea('ram', 100)" fill="url(#tlGradRam)" />
            <path :d="buildTimelinePath('ram', 100)" fill="none" stroke="#10B981" stroke-width="1.8" stroke-linecap="round" />
          </g>
          <g v-if="visibleSeries.network">
            <path :d="buildTimelineArea('netKbps', timelineNetPeak)" fill="url(#tlGradNet)" />
            <path :d="buildTimelinePath('netKbps', timelineNetPeak)" fill="none" stroke="#06B6D4" stroke-width="1.8" stroke-linecap="round" />
          </g>
          <g v-if="visibleSeries.cpu">
            <path :d="buildTimelineArea('cpu', 100)" fill="url(#tlGradCpu)" />
            <path :d="buildTimelinePath('cpu', 100)" fill="none" stroke="#3B82F6" stroke-width="2" stroke-linecap="round" />
          </g>
          <g v-if="visibleSeries.throughput">
            <path :d="buildTimelinePath('throughput', timelineThroughputPeak)" fill="none" stroke="#8B5CF6" stroke-width="1.6" stroke-dasharray="4 2" stroke-linecap="round" />
          </g>

          <!-- Hover Crosshair -->
          <line
            v-if="hoverIndex >= 0 && timelineSamples[hoverIndex]"
            :x1="hoverXCoord"
            y1="6"
            :x2="hoverXCoord"
            y2="102"
            stroke="#64748B"
            stroke-width="1"
            stroke-dasharray="2 2"
          />
        </svg>

        <!-- Hover Tooltip -->
        <div
          v-if="hoverSample"
          class="pointer-events-none absolute top-1 z-20 px-2.5 py-1.5 rounded-lg text-[10.5px] font-mono bg-white/95 dark:bg-slate-900/95 text-slate-800 dark:text-slate-100 border border-slate-200 dark:border-slate-700 shadow-lg flex items-center gap-3"
          :style="{ left: hoverTooltipLeft }"
        >
          <span class="font-bold text-slate-600 dark:text-slate-300">{{ hoverSample.time }}</span>
          <span v-if="visibleSeries.cpu" class="text-blue-600 dark:text-sky-400">CPU {{ hoverSample.cpu.toFixed(1) }}%</span>
          <span v-if="visibleSeries.ram" class="text-emerald-600 dark:text-emerald-400">RAM {{ hoverSample.ram.toFixed(1) }}%</span>
          <span v-if="visibleSeries.network" class="text-cyan-600 dark:text-cyan-400">Net {{ hoverSample.netKbps.toFixed(1) }} KB/s</span>
          <span v-if="visibleSeries.throughput" class="text-purple-600 dark:text-purple-400">{{ hoverSample.throughput.toFixed(1) }} /s</span>
        </div>

        <!-- Time Axis Labels -->
        <div class="flex items-center justify-between text-[10px] font-mono text-slate-400 dark:text-slate-500 pt-1">
          <span v-for="(lbl, idx) in timelineAxisLabels" :key="'tl-lbl-' + idx">{{ lbl }}</span>
        </div>
      </div>
    </div>

    <!-- MODAL: Detailed Network Interfaces (NIC), WAN & Outbound Links -->
    <div
      v-if="showNetworkModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs"
      @click.self="showNetworkModal = false"
    >
      <div class="w-full max-w-3xl rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 shadow-2xl overflow-hidden">
        <!-- Modal Header -->
        <div class="flex items-center justify-between px-5 py-3.5 border-b border-slate-100 dark:border-slate-800">
          <div>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white">
              {{ $t('systemMonitoring.networkModalTitle') }}
            </h3>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 mt-0.5">
              {{ $t('systemMonitoring.networkModalSubtitle') }}
            </p>
          </div>
          <button
            type="button"
            @click="showNetworkModal = false"
            class="w-7 h-7 rounded-lg flex items-center justify-center text-slate-400 hover:text-slate-700 hover:bg-slate-100 dark:hover:text-white dark:hover:bg-slate-800 cursor-pointer"
          >
            &times;
          </button>
        </div>

        <!-- Modal Body -->
        <div class="p-5 space-y-4 max-h-[75vh] overflow-y-auto text-xs">
          <!-- Summary Cards -->
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5">
            <div class="p-2.5 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-200/80 dark:border-slate-800">
              <div class="text-[10px] text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.lanStatus') }}</div>
              <div class="font-mono font-bold text-slate-900 dark:text-white mt-0.5 truncate">{{ displayPrimaryIp }}</div>
            </div>
            <div class="p-2.5 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-200/80 dark:border-slate-800">
              <div class="text-[10px] text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.gateway') }}</div>
              <div class="font-mono font-bold mt-0.5 truncate" :class="networkData.local_network.gateway_reachable ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-500'">
                {{ displayGateway }}
              </div>
            </div>
            <div class="p-2.5 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-200/80 dark:border-slate-800">
              <div class="text-[10px] text-slate-500 dark:text-slate-400">{{ $t('systemMonitoring.internetWan') }}</div>
              <div class="font-mono font-bold mt-0.5 truncate" :class="wanTextClass">{{ wanSummaryText }}</div>
            </div>
            <div class="p-2.5 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-200/80 dark:border-slate-800">
              <div class="text-[10px] text-slate-500 dark:text-slate-400">Total RX / TX</div>
              <div class="font-mono font-bold text-cyan-600 dark:text-cyan-400 mt-0.5 truncate">
                {{ formatTotalBytes(status.net_rx_total_bytes) }} / {{ formatTotalBytes(status.net_tx_total_bytes) }}
              </div>
            </div>
          </div>

          <!-- Active Network Interfaces Table -->
          <div>
            <div class="text-[11px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">
              {{ $t('systemMonitoring.nicInterfaces') }}
            </div>
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
              <table class="w-full text-left text-[11px]">
                <thead class="bg-slate-50 dark:bg-slate-900/80 text-slate-500 dark:text-slate-400 font-semibold">
                  <tr>
                    <th class="px-3 py-2">Interface</th>
                    <th class="px-3 py-2">Type</th>
                    <th class="px-3 py-2">IP Address</th>
                    <th class="px-3 py-2">MAC</th>
                    <th class="px-3 py-2 text-right">Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800/80 font-mono">
                  <tr v-for="(nic, idx) in sanitizedInterfaces" :key="'nic-' + idx">
                    <td class="px-3 py-2 font-semibold text-slate-800 dark:text-slate-200">{{ nic.name }}</td>
                    <td class="px-3 py-2 text-slate-500 dark:text-slate-400">{{ nic.type }}</td>
                    <td class="px-3 py-2 text-slate-700 dark:text-slate-300">{{ nic.ip || '-' }}</td>
                    <td class="px-3 py-2 text-slate-500 dark:text-slate-400">{{ nic.mac || '-' }}</td>
                    <td class="px-3 py-2 text-right">
                      <span :class="nic.is_up ? 'sysmon-pill sysmon-pill-emerald' : 'sysmon-pill sysmon-pill-slate'">
                        {{ nic.is_up ? 'UP' : 'DOWN' }}
                      </span>
                    </td>
                  </tr>
                  <tr v-if="!sanitizedInterfaces.length">
                    <td colspan="5" class="px-3 py-4 text-center text-slate-400 font-sans">No active network interfaces reported</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Outbound Delivery Destinations -->
          <div v-if="!isCloudMode || isSuperAdmin">
            <div class="text-[11px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-2">
              {{ $t('systemMonitoring.outboundLinks') }}
            </div>
            <div v-if="outboundDestinations.length" class="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
              <table class="w-full text-left text-[11px]">
                <thead class="bg-slate-50 dark:bg-slate-900/80 text-slate-500 dark:text-slate-400 font-semibold">
                  <tr>
                    <th class="px-3 py-2">Destination</th>
                    <th class="px-3 py-2">Protocol</th>
                    <th class="px-3 py-2">Success Rate</th>
                    <th class="px-3 py-2">Queue</th>
                    <th class="px-3 py-2 text-right">Health</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800/80 font-mono">
                  <tr v-for="dest in outboundDestinations" :key="'dest-' + dest.id">
                    <td class="px-3 py-2 font-sans font-semibold text-slate-800 dark:text-slate-200">{{ dest.name }}</td>
                    <td class="px-3 py-2 text-slate-500 dark:text-slate-400">{{ dest.type }}</td>
                    <td class="px-3 py-2 text-emerald-600 dark:text-emerald-400">{{ Number(dest.success_rate || 0).toFixed(1) }}%</td>
                    <td class="px-3 py-2 text-slate-700 dark:text-slate-300">{{ dest.queue_count || 0 }}</td>
                    <td class="px-3 py-2 text-right">
                      <span
                        class="sysmon-pill"
                        :class="dest.status === 'HEALTHY' ? 'sysmon-pill-emerald' : dest.status === 'FAILED' ? 'sysmon-pill-rose' : 'sysmon-pill-amber'"
                      >
                        {{ dest.status }}
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="px-3 py-3 rounded-xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 text-slate-500 dark:text-slate-400 text-center">
              {{ $t('systemMonitoring.noOutboundConfigured') }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

const RING_RADIUS = 32;
const RING_CIRCUMFERENCE = 2 * Math.PI * RING_RADIUS;
const MAX_SPARK_POINTS = 24;
const MAX_TIMELINE_POINTS = 60;

export default {
  name: 'SystemMonitoringView',
  data() {
    return {
      timer: null,
      pollIntervalSec: 5,
      pollTickCount: 0,
      isRefreshing: false,
      showNetworkModal: false,
      hoverIndex: -1,
      ringCircumference: RING_CIRCUMFERENCE,
      cpuHistory: [],
      ramHistory: [],
      netHistory: [],
      timelineSamples: [],
      visibleSeries: {
        cpu: true,
        ram: true,
        network: true,
        throughput: true,
      },
      networkData: {
        local_network: {
          status: 'UP',
          primary_ip: '',
          primary_interface: '',
          gateway: '',
          gateway_reachable: true,
          gateway_latency_ms: 0,
          interfaces: [],
        },
        internet: {
          status: 'CONNECTED',
          connected: true,
          probe_target: '1.1.1.1:53',
          latency_ms: 0,
        },
        outbound: {
          total_links: 0,
          healthy_count: 0,
          overall_status: 'UNKNOWN',
          destinations: [],
        },
      },
    };
  },
  computed: {
    status() {
      return this.$store.state.systemStatus || {};
    },
    isDarkMode() {
      if (this.$store && this.$store.state) {
        if (this.$store.state.uiPreferences && this.$store.state.uiPreferences.effectiveTheme) {
          return this.$store.state.uiPreferences.effectiveTheme === 'dark';
        }
        if (this.$store.state.theme) {
          return this.$store.state.theme === 'dark';
        }
      }
      return typeof document !== 'undefined' && document.documentElement.classList.contains('dark');
    },
    userRole() {
      const u = (this.$store && this.$store.state && this.$store.state.user) || {};
      return String(u.role || '').toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    isOperator() {
      return !!this.$store.getters.isOperator || this.userRole === 'OPERATOR';
    },
    isCloudMode() {
      const sys = this.status || {};
      return Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
    },
    canViewHostPrivateNetwork() {
      return !this.isCloudMode || this.isSuperAdmin;
    },
    osArchLabel() {
      const rawOs = String(this.status.os || 'linux').split('/')[0].trim().toLowerCase();
      const arch = String(this.status.arch || 'amd64').trim();
      return `${rawOs}/${arch}`;
    },
    displayHostname() {
      if (!this.canViewHostPrivateNetwork) {
        return '';
      }
      return this.status.hostname || '';
    },
    formattedDbLatency() {
      const ms = this.status.db_latency_ms;
      if (ms == null || Number.isNaN(Number(ms))) {
        return '0.8 ms';
      }
      return `${Number(ms).toFixed(1)} ms`;
    },
    cpuPercentFixed() {
      return Number(this.status.cpu_percent || 0).toFixed(1);
    },
    processCpuFixed() {
      return Number(this.status.process_cpu_percent || 0).toFixed(1);
    },
    ramPercentFixed() {
      return Number(this.status.ram_percent || 0).toFixed(1);
    },
    diskPercentFixed() {
      return Number(this.status.disk_percent || 0).toFixed(1);
    },
    cpuStrokeColor() {
      const v = Number(this.status.cpu_percent || 0);
      if (v > 85) return '#EF4444';
      if (v > 70) return '#F59E0B';
      return '#3B82F6';
    },
    ramStrokeColor() {
      const v = Number(this.status.ram_percent || 0);
      if (v > 90) return '#EF4444';
      if (v > 75) return '#F59E0B';
      return '#10B981';
    },
    diskStrokeColor() {
      const v = Number(this.status.disk_percent || 0);
      if (v > 90) return '#EF4444';
      if (v > 80) return '#F59E0B';
      return '#6366F1';
    },
    visibleCores() {
      const cores = Array.isArray(this.status.cpu_per_core) ? this.status.cpu_per_core : [];
      if (cores.length > 0) {
        return cores.slice(0, 4);
      }
      const base = Number(this.status.cpu_percent || 0);
      return [base, Math.max(0, base * 0.8), Math.max(0, base * 0.9), Math.max(0, base * 0.6)];
    },
    ramUsedLabel() {
      const usedMb = Number(this.status.ram_used_mb || 0);
      const totalMb = Number(this.status.ram_total_mb || 0);
      return totalMb >= 1024 ? `${(usedMb / 1024).toFixed(2)} GB` : `${usedMb.toFixed(0)} MB`;
    },
    ramTotalLabel() {
      const totalMb = Number(this.status.ram_total_mb || 0);
      return totalMb >= 1024 ? `${(totalMb / 1024).toFixed(2)} GB` : `${totalMb.toFixed(0)} MB`;
    },
    ramFreeLabel() {
      const usedMb = Number(this.status.ram_used_mb || 0);
      const totalMb = Number(this.status.ram_total_mb || 0);
      const freeMb = Math.max(0, totalMb - usedMb);
      return totalMb >= 1024 ? `${(freeMb / 1024).toFixed(2)} GB` : `${freeMb.toFixed(0)} MB`;
    },
    memoryPressureLabel() {
      const pct = Number(this.status.ram_percent || 0);
      if (pct > 85) return this.$t('systemMonitoring.highPressure');
      if (pct > 68) return this.$t('systemMonitoring.moderateLoad');
      return this.$t('systemMonitoring.optimalHeadroom');
    },
    memoryPressureBadgeClass() {
      const pct = Number(this.status.ram_percent || 0);
      if (pct > 85) {
        return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-500/15 dark:text-rose-300 dark:border-rose-500/30';
      }
      if (pct > 68) {
        return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-500/15 dark:text-amber-300 dark:border-amber-500/30';
      }
      return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-500/15 dark:text-emerald-300 dark:border-emerald-500/30';
    },
    diskUsedLabel() {
      return `${Number(this.status.disk_used_gb || 0).toFixed(1)} GB`;
    },
    diskTotalLabel() {
      return `${Number(this.status.disk_total_gb || 0).toFixed(1)} GB`;
    },
    diskFreeLabel() {
      const free = Math.max(0, Number(this.status.disk_total_gb || 0) - Number(this.status.disk_used_gb || 0));
      return `${free.toFixed(1)} GB`;
    },
    shortQueueStatus() {
      const raw = String(this.status.queue_status || 'HEALTHY');
      return raw.length > 16 ? raw.slice(0, 16) : raw;
    },
    retentionDaysLabel() {
      const hrs = Number(this.status.raw_telemetry_max_range_hours || 168);
      if (hrs >= 24 && hrs % 24 === 0) {
        return `${hrs / 24}d`;
      }
      return `${hrs}h`;
    },
    // Network (Jaringan) Computed Properties
    lanUp() {
      return String((this.networkData.local_network && this.networkData.local_network.status) || 'UP').toUpperCase() === 'UP';
    },
    wanStatus() {
      return String((this.networkData.internet && this.networkData.internet.status) || 'CONNECTED').toUpperCase();
    },
    networkHealthScore() {
      if (!this.lanUp) return 8;
      if (this.wanStatus === 'CONNECTED') return 100;
      if (this.wanStatus === 'DEGRADED') return 72;
      return 55;
    },
    networkStrokeColor() {
      if (!this.lanUp) return '#EF4444';
      if (this.wanStatus === 'CONNECTED') return '#06B6D4';
      return '#F59E0B';
    },
    networkOverallLabel() {
      if (!this.lanUp) return this.$t('systemMonitoring.networkOffline');
      if (this.wanStatus === 'CONNECTED') return this.$t('systemMonitoring.networkOnline');
      if (this.wanStatus === 'DEGRADED') return this.$t('systemMonitoring.networkDegraded');
      return this.$t('systemMonitoring.networkLanOnly');
    },
    networkBadgeClass() {
      if (!this.lanUp) {
        return 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-500/15 dark:text-rose-300 dark:border-rose-500/30';
      }
      if (this.wanStatus === 'CONNECTED') {
        return 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-500/15 dark:text-emerald-300 dark:border-emerald-500/30';
      }
      return 'bg-amber-50 text-amber-700 border-amber-200 dark:bg-amber-500/15 dark:text-amber-300 dark:border-amber-500/30';
    },
    networkDotClass() {
      if (!this.lanUp) return 'bg-rose-500';
      if (this.wanStatus === 'CONNECTED') return 'bg-emerald-500';
      return 'bg-amber-500';
    },
    netCombinedRateCompact() {
      const total = Number(this.status.net_rx_bytes_per_sec || 0) + Number(this.status.net_tx_bytes_per_sec || 0);
      return this.formatByteRate(total);
    },
    netHistoryPeak() {
      const maxVal = Math.max(10, ...this.netHistory);
      return maxVal * 1.15;
    },
    activeNicCount() {
      const list = (this.networkData.local_network && this.networkData.local_network.interfaces) || [];
      const count = list.filter(i => i && i.is_up && i.type !== 'LOOPBACK').length;
      return count || (this.lanUp ? 1 : 0);
    },
    displayPrimaryIp() {
      if (!this.canViewHostPrivateNetwork) {
        return this.$t('systemMonitoring.cloudVpcProtected');
      }
      const ln = this.networkData.local_network || {};
      if (!ln.primary_ip) return this.lanUp ? 'Active' : '-';
      return ln.primary_interface ? `${ln.primary_ip} (${ln.primary_interface})` : ln.primary_ip;
    },
    displayGateway() {
      const ln = this.networkData.local_network || {};
      const lat = ln.gateway_latency_ms ? `${Number(ln.gateway_latency_ms).toFixed(1)} ms` : '';
      if (!this.canViewHostPrivateNetwork) {
        return ln.gateway_reachable
          ? `${this.$t('systemMonitoring.reachable')}${lat ? ' · ' + lat : ''}`
          : this.$t('systemMonitoring.unreachable');
      }
      if (!ln.gateway) {
        return ln.gateway_reachable ? this.$t('systemMonitoring.reachable') : '-';
      }
      return lat ? `${ln.gateway} · ${lat}` : ln.gateway;
    },
    wanSummaryText() {
      const inet = this.networkData.internet || {};
      const st = this.wanStatus;
      const lat = inet.latency_ms ? `${Number(inet.latency_ms).toFixed(0)} ms` : '';
      if (st === 'CONNECTED') {
        return lat ? `CONNECTED · ${lat}` : 'CONNECTED';
      }
      return st;
    },
    wanTextClass() {
      if (this.wanStatus === 'CONNECTED') return 'text-emerald-600 dark:text-emerald-400';
      if (this.wanStatus === 'DEGRADED') return 'text-amber-600 dark:text-amber-400';
      return 'text-rose-600 dark:text-rose-400';
    },
    sanitizedInterfaces() {
      const list = (this.networkData.local_network && this.networkData.local_network.interfaces) || [];
      return list
        .filter(i => i && i.type !== 'LOOPBACK')
        .map((nic, idx) => ({
          name: this.canViewHostPrivateNetwork ? nic.name : `eth${idx}`,
          type: nic.type || 'ETHERNET',
          ip: this.canViewHostPrivateNetwork ? nic.ip : this.$t('systemMonitoring.cloudVpcProtected'),
          mac: this.canViewHostPrivateNetwork ? nic.mac : '••:••:••:••:••:••',
          is_up: !!nic.is_up,
        }));
    },
    outboundTotalLinks() {
      return Number((this.networkData.outbound && this.networkData.outbound.total_links) || 0);
    },
    outboundHealthyLinks() {
      return Number((this.networkData.outbound && this.networkData.outbound.healthy_count) || 0);
    },
    outboundDestinations() {
      return (this.networkData.outbound && this.networkData.outbound.destinations) || [];
    },
    // Subsystem Fleet & Quality
    fleetOnlinePercent() {
      const total = Number(this.status.total_devices || 0);
      if (total <= 0) return 100;
      return Math.round((Number(this.status.online_devices || 0) / total) * 100);
    },
    fleetBarSegments() {
      const total = Math.max(1, Number(this.status.total_devices || 0));
      if (!this.status.total_devices) {
        return { online: 100, offline: 0, error: 0 };
      }
      const online = (Number(this.status.online_devices || 0) / total) * 100;
      const error = (Number(this.status.error_devices || 0) / total) * 100;
      const offline = Math.max(0, 100 - online - error);
      return { online, offline, error };
    },
    qualityPercent() {
      const raw = String(this.status.data_quality || '100');
      const match = raw.match(/([\d.]+)/);
      return match ? this.clampPercent(parseFloat(match[1])) : 100;
    },
    // Bottom 5-Minute Live Timeline
    timelineSeriesDefs() {
      return [
        { key: 'cpu', label: this.$t('systemMonitoring.seriesHostCpu'), color: '#3B82F6' },
        { key: 'ram', label: this.$t('systemMonitoring.seriesMemory'), color: '#10B981' },
        { key: 'network', label: this.$t('systemMonitoring.seriesNetwork'), color: '#06B6D4' },
        { key: 'throughput', label: this.$t('systemMonitoring.seriesThroughput'), color: '#8B5CF6' },
      ];
    },
    timelineNetPeak() {
      const vals = this.timelineSamples.map(s => s.netKbps || 0);
      return Math.max(20, ...vals) * 1.2;
    },
    timelineThroughputPeak() {
      const vals = this.timelineSamples.map(s => s.throughput || 0);
      return Math.max(10, ...vals) * 1.25;
    },
    timelineAxisLabels() {
      const pts = this.timelineSamples;
      if (!pts.length) return ['--:--', '--:--', '--:--', '--:--', '--:--'];
      const indices = [0, Math.floor(pts.length * 0.25), Math.floor(pts.length * 0.5), Math.floor(pts.length * 0.75), pts.length - 1];
      return indices.map(i => (pts[i] && pts[i].timeShort) || '--:--');
    },
    hoverSample() {
      if (this.hoverIndex < 0 || this.hoverIndex >= this.timelineSamples.length) return null;
      return this.timelineSamples[this.hoverIndex];
    },
    hoverXCoord() {
      const count = Math.max(2, this.timelineSamples.length);
      return (this.hoverIndex / (count - 1)) * 960;
    },
    hoverTooltipLeft() {
      const ratio = this.timelineSamples.length > 1 ? this.hoverIndex / (this.timelineSamples.length - 1) : 0.5;
      if (ratio > 0.68) return 'auto';
      return `${Math.max(8, Math.round(ratio * 72))}%`;
    },
  },
  watch: {
    isCloudMode() {
      this.enforceCloudOperatorGuard();
    },
    isOperator() {
      this.enforceCloudOperatorGuard();
    },
  },
  methods: {
    enforceCloudOperatorGuard() {
      if (this.isOperator && this.isCloudMode) {
        this.clearPollTimer();
        this.$router.replace({
          path: '/access-denied',
          query: { from: this.$route.fullPath },
        });
        return true;
      }
      return false;
    },
    clampPercent(val) {
      const n = Number(val || 0);
      if (Number.isNaN(n)) return 0;
      return Math.max(0, Math.min(100, n));
    },
    ringDashOffset(pct) {
      const clamped = this.clampPercent(pct);
      return RING_CIRCUMFERENCE * (1 - clamped / 100);
    },
    formatByteRate(bytesPerSec) {
      const b = Math.max(0, Number(bytesPerSec || 0));
      if (b >= 1024 * 1024) {
        return `${(b / (1024 * 1024)).toFixed(1)} MB/s`;
      }
      if (b >= 1024) {
        return `${(b / 1024).toFixed(1)} KB/s`;
      }
      return `${b.toFixed(0)} B/s`;
    },
    formatTotalBytes(bytes) {
      const b = Math.max(0, Number(bytes || 0));
      if (b >= 1024 * 1024 * 1024) {
        return `${(b / (1024 * 1024 * 1024)).toFixed(2)} GB`;
      }
      if (b >= 1024 * 1024) {
        return `${(b / (1024 * 1024)).toFixed(1)} MB`;
      }
      if (b >= 1024) {
        return `${(b / 1024).toFixed(0)} KB`;
      }
      return `${b} B`;
    },
    async fetchNetworkDiagnostics() {
      try {
        const res = await axios.get('/api/system/network');
        const payload = (res && res.data && (res.data.data || res.data)) || {};
        if (payload.local_network) {
          this.networkData = {
            local_network: payload.local_network || this.networkData.local_network,
            internet: payload.internet || this.networkData.internet,
            outbound: payload.outbound || this.networkData.outbound,
          };
        }
      } catch (_) {
        // Keep last known network diagnostics gracefully
      }
    },
    seedInitialHistories() {
      const cpu = Number(this.status.cpu_percent || 4);
      const ram = Number(this.status.ram_percent || 12);
      const netKbps = (Number(this.status.net_rx_bytes_per_sec || 0) + Number(this.status.net_tx_bytes_per_sec || 0)) / 1024;
      const throughput = Number(this.status.data_per_sec || 0);
      const now = Date.now();

      this.cpuHistory = Array.from({ length: 16 }, () => cpu);
      this.ramHistory = Array.from({ length: 16 }, () => ram);
      this.netHistory = Array.from({ length: 16 }, () => netKbps);

      this.timelineSamples = Array.from({ length: 24 }, (_, idx) => {
        const ts = new Date(now - (23 - idx) * 5000);
        return {
          time: ts.toTimeString().slice(0, 8),
          timeShort: ts.toTimeString().slice(0, 5),
          cpu,
          ram,
          netKbps,
          throughput,
        };
      });
    },
    recordSample() {
      const cpu = this.clampPercent(this.status.cpu_percent);
      const ram = this.clampPercent(this.status.ram_percent);
      const netKbps = Math.max(
        0,
        (Number(this.status.net_rx_bytes_per_sec || 0) + Number(this.status.net_tx_bytes_per_sec || 0)) / 1024
      );
      const throughput = Math.max(0, Number(this.status.data_per_sec || 0));

      if (!this.timelineSamples.length) {
        this.seedInitialHistories();
      }

      this.cpuHistory = [...this.cpuHistory.slice(-(MAX_SPARK_POINTS - 1)), cpu];
      this.ramHistory = [...this.ramHistory.slice(-(MAX_SPARK_POINTS - 1)), ram];
      this.netHistory = [...this.netHistory.slice(-(MAX_SPARK_POINTS - 1)), netKbps];

      const now = new Date();
      this.timelineSamples = [
        ...this.timelineSamples.slice(-(MAX_TIMELINE_POINTS - 1)),
        {
          time: now.toTimeString().slice(0, 8),
          timeShort: now.toTimeString().slice(0, 5),
          cpu,
          ram,
          netKbps,
          throughput,
        },
      ];
    },
    buildSmoothSvgCoords(values, width, height, maxVal, padTop = 4, padBottom = 4) {
      const list = Array.isArray(values) && values.length ? values : [0, 0];
      const safeMax = Math.max(1, Number(maxVal || 100));
      const usableH = Math.max(8, height - padTop - padBottom);
      const count = Math.max(2, list.length);

      return list.map((v, i) => {
        const x = (i / (count - 1)) * width;
        const ratio = Math.max(0, Math.min(1, Number(v || 0) / safeMax));
        const y = padTop + (1 - ratio) * usableH;
        return { x, y };
      });
    },
    coordsToSmoothPath(pts) {
      if (!pts.length) return '';
      if (pts.length === 1) return `M ${pts[0].x.toFixed(1)} ${pts[0].y.toFixed(1)}`;
      let d = `M ${pts[0].x.toFixed(1)} ${pts[0].y.toFixed(1)}`;
      for (let i = 0; i < pts.length - 1; i++) {
        const p0 = pts[i];
        const p1 = pts[i + 1];
        const midX = ((p0.x + p1.x) / 2).toFixed(1);
        d += ` C ${midX} ${p0.y.toFixed(1)}, ${midX} ${p1.y.toFixed(1)}, ${p1.x.toFixed(1)} ${p1.y.toFixed(1)}`;
      }
      return d;
    },
    buildSparklinePath(values, width, height, maxVal) {
      const pts = this.buildSmoothSvgCoords(values, width, height, maxVal, 3, 3);
      return this.coordsToSmoothPath(pts);
    },
    buildSparklineArea(values, width, height, maxVal) {
      const pts = this.buildSmoothSvgCoords(values, width, height, maxVal, 3, 3);
      if (!pts.length) return '';
      const line = this.coordsToSmoothPath(pts);
      const last = pts[pts.length - 1];
      const first = pts[0];
      return `${line} L ${last.x.toFixed(1)} ${height} L ${first.x.toFixed(1)} ${height} Z`;
    },
    buildTimelinePath(field, maxVal) {
      const vals = this.timelineSamples.map(s => s[field] || 0);
      const pts = this.buildSmoothSvgCoords(vals, 960, 102, maxVal, 10, 4);
      return this.coordsToSmoothPath(pts);
    },
    buildTimelineArea(field, maxVal) {
      const vals = this.timelineSamples.map(s => s[field] || 0);
      const pts = this.buildSmoothSvgCoords(vals, 960, 102, maxVal, 10, 4);
      if (!pts.length) return '';
      const line = this.coordsToSmoothPath(pts);
      return `${line} L 960 100 L 0 100 Z`;
    },
    toggleTimelineSeries(key) {
      const next = !this.visibleSeries[key];
      const activeCount = Object.values(this.visibleSeries).filter(Boolean).length;
      if (!next && activeCount <= 1) return;
      this.$set(this.visibleSeries, key, next);
    },
    onTimelineMouseMove(evt) {
      if (!this.timelineSamples.length) return;
      const rect = evt.currentTarget.getBoundingClientRect();
      const relX = Math.max(0, Math.min(1, (evt.clientX - rect.left) / Math.max(1, rect.width)));
      this.hoverIndex = Math.round(relX * (this.timelineSamples.length - 1));
    },
    async manualRefresh() {
      if (this.enforceCloudOperatorGuard()) return;
      this.isRefreshing = true;
      try {
        await Promise.all([
          this.$store.dispatch('fetchSystemStatus'),
          this.fetchNetworkDiagnostics(),
        ]);
        if (!this.enforceCloudOperatorGuard()) {
          this.recordSample();
        }
      } finally {
        this.isRefreshing = false;
      }
    },
    cyclePollInterval() {
      const cycle = [5, 10, 30, 0];
      const idx = cycle.indexOf(this.pollIntervalSec);
      this.pollIntervalSec = cycle[(idx + 1) % cycle.length];
      this.restartPollTimer();
    },
    clearPollTimer() {
      if (this.timer) {
        clearInterval(this.timer);
        this.timer = null;
      }
    },
    restartPollTimer() {
      this.clearPollTimer();
      if (this.pollIntervalSec <= 0) return;
      this.timer = setInterval(async () => {
        if (this.enforceCloudOperatorGuard()) return;
        this.pollTickCount += 1;
        await this.$store.dispatch('fetchSystemStatus');
        if (this.pollTickCount % 3 === 0) {
          await this.fetchNetworkDiagnostics();
        }
        if (!this.enforceCloudOperatorGuard()) {
          this.recordSample();
        }
      }, this.pollIntervalSec * 1000);
    },
  },
  async mounted() {
    if (this.enforceCloudOperatorGuard()) {
      return;
    }
    await Promise.all([
      this.$store.dispatch('fetchSystemStatus'),
      this.fetchNetworkDiagnostics(),
    ]);
    if (this.enforceCloudOperatorGuard()) {
      return;
    }
    this.seedInitialHistories();
    this.recordSample();
    this.restartPollTimer();
  },
  beforeDestroy() {
    this.clearPollTimer();
  },
};
</script>

<style scoped>
.sysmon-pill {
  display: inline-flex;
  align-items: center;
  padding: 1px 7px;
  border-radius: 9999px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 10px;
  font-weight: 700;
  line-height: 1.35;
  border: 1px solid transparent;
}

.sysmon-pill-emerald {
  background-color: #ecfdf5;
  color: #047857;
  border-color: #a7f3d0;
}

.sysmon-pill-amber {
  background-color: #fffbeb;
  color: #b45309;
  border-color: #fde68a;
}

.sysmon-pill-rose {
  background-color: #fff1f2;
  color: #be123c;
  border-color: #fecdd3;
}

.sysmon-pill-slate {
  background-color: #f1f5f9;
  color: #475569;
  border-color: #cbd5e1;
}

.sysmon-dark .sysmon-pill-emerald {
  background-color: rgba(16, 185, 129, 0.14);
  color: #6ee7b7;
  border-color: rgba(16, 185, 129, 0.32);
}

.sysmon-dark .sysmon-pill-amber {
  background-color: rgba(245, 158, 11, 0.15);
  color: #fcd34d;
  border-color: rgba(245, 158, 11, 0.32);
}

.sysmon-dark .sysmon-pill-rose {
  background-color: rgba(244, 63, 94, 0.15);
  color: #fda4af;
  border-color: rgba(244, 63, 94, 0.32);
}

.sysmon-dark .sysmon-pill-slate {
  background-color: rgba(51, 65, 85, 0.45);
  color: #cbd5e1;
  border-color: rgba(100, 116, 139, 0.35);
}

.sysmon-dark .sysmon-ring-glow {
  filter: drop-shadow(0 0 5px rgba(56, 189, 248, 0.22));
}
</style>
