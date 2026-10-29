<template>
  <div class="flex flex-col h-full min-h-0 space-y-2 font-sans">
    <!-- Header Section: Title & Actions in Single Line on Desktop -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 flex-shrink-0">
      <div class="flex items-center flex-wrap gap-x-2.5 gap-y-1 min-w-0">
        <!-- Breadcrumb & Back -->
        <router-link
          to="/monitoring/devices"
          class="hover:text-blue-500 text-slate-400 dark:text-slate-500 flex items-center space-x-1 text-2xs transition-colors flex-shrink-0"
          :title="$t('navigation.monitoring')"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
          <span class="font-medium">{{ $t('navigation.monitoring') }}</span>
          <span class="text-slate-300 dark:text-slate-700">/</span>
        </router-link>

        <!-- Main Title with live pulse -->
        <h1 class="text-sm sm:text-base font-bold tracking-tight text-slate-900 dark:text-white flex items-center space-x-1.5 flex-shrink-0">
          <span>{{ $t('telemetry.incomingTitle') }}</span>
          <span class="relative flex h-2.5 w-2.5">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
          </span>
        </h1>

        <!-- Status Badge -->
        <span class="px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20 text-3xs font-mono font-bold whitespace-nowrap flex-shrink-0">
          {{ $t('telemetry.realtimeIngestionActive') }}
        </span>

        <!-- Compact Subtitle (Inline on desktop) -->
        <span class="hidden xl:inline text-slate-300 dark:text-slate-700 text-xs flex-shrink-0">•</span>
        <p class="text-2xs text-slate-500 dark:text-slate-400 truncate max-w-xs md:max-w-sm xl:max-w-md hidden sm:block" :title="$t('telemetry.incomingSubtitle')">
          {{ $t('telemetry.incomingSubtitle') }}
        </p>
      </div>

      <!-- Quick Actions (Single line on desktop) -->
      <div class="flex items-center space-x-1.5 flex-shrink-0">
        <button
          @click="fetchTelemetry"
          :disabled="loading"
          class="h-7 px-2.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 border border-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 dark:border-slate-700 text-2xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50 shadow-2xs"
        >
          <svg class="w-3 h-3" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>{{ $t('common.refresh') }}</span>
        </button>

        <router-link
          to="/monitoring/analysis"
          class="h-7 px-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-2xs font-semibold transition flex items-center space-x-1.5 shadow-2xs shadow-indigo-600/20"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <span class="whitespace-nowrap">{{ $t('analysis.title') }}</span>
        </router-link>

        <router-link
          to="/monitoring/devices"
          class="h-7 px-2.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-2xs font-semibold transition flex items-center space-x-1.5 shadow-2xs shadow-blue-600/20"
        >
          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="2"></circle>
            <path d="M16.24 7.76a6 6 0 0 1 0 8.49m-8.48-.01a6 6 0 0 1 0-8.49m11.31-2.82a10 10 0 0 1 0 14.14m-14.14 0a10 10 0 0 1 0-14.14"></path>
          </svg>
          <span class="whitespace-nowrap">{{ $t('telemetry.manageDevices') }}</span>
        </router-link>
      </div>
    </div>

    <!-- Summary KPI Cards in a Single Row -->
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-2 flex-shrink-0">
      <!-- 1. Connected Devices -->
      <div class="saas-card px-3 py-1.5 bg-white dark:bg-[#0F172A]/70 border border-slate-200 dark:border-slate-800 rounded-xl flex flex-col justify-between">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-500 dark:text-slate-400 truncate">{{ $t('telemetry.connectedDevices') }}</div>
        <div class="flex items-baseline justify-between gap-1 mt-0.5">
          <div class="flex items-baseline space-x-1 font-mono">
            <span class="text-base sm:text-lg font-bold text-slate-900 dark:text-white leading-none">{{ stats.activeDevices }}</span>
            <span class="text-2xs text-slate-500 dark:text-slate-400">/ {{ stats.totalDevices }}</span>
          </div>
          <span class="text-3xs text-emerald-600 dark:text-emerald-400 font-mono truncate">{{ $t('telemetry.enginesRunning') }}</span>
        </div>
      </div>

      <!-- 2. Health Index -->
      <div class="saas-card px-3 py-1.5 bg-white dark:bg-[#0F172A]/70 border border-slate-200 dark:border-slate-800 rounded-xl flex flex-col justify-between">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-500 dark:text-slate-400 truncate">{{ $t('telemetry.healthIndex') }}</div>
        <div class="flex items-baseline justify-between gap-1 mt-0.5">
          <div class="flex items-baseline space-x-1 font-mono">
            <span class="text-base sm:text-lg font-bold leading-none" :class="stats.healthPercent >= 90 ? 'text-emerald-600 dark:text-emerald-400' : (stats.healthPercent >= 70 ? 'text-amber-600 dark:text-amber-400' : 'text-rose-600 dark:text-rose-400')">{{ stats.healthPercent }}%</span>
            <span class="text-3xs text-slate-500 dark:text-slate-400 truncate">{{ $t('telemetry.overallHealth') }}</span>
          </div>
          <span class="text-3xs text-slate-600 dark:text-slate-400 font-mono truncate">{{ stats.goodSensors }}/{{ parametersList.length }} {{ $t('telemetry.qualityGood') }}</span>
        </div>
      </div>

      <!-- 3. Good Quality -->
      <div class="saas-card px-3 py-1.5 bg-white dark:bg-[#0F172A]/70 border border-emerald-200 dark:border-emerald-500/20 rounded-xl flex flex-col justify-between">
        <div class="text-3xs uppercase tracking-wider font-semibold text-emerald-600 dark:text-emerald-400 truncate">{{ $t('telemetry.qualityGood') }}</div>
        <div class="flex items-baseline justify-between gap-1 mt-0.5">
          <div class="flex items-baseline space-x-1 font-mono">
            <span class="text-base sm:text-lg font-bold text-emerald-600 dark:text-emerald-400 leading-none">{{ stats.goodSensors }}</span>
            <span class="text-2xs text-slate-500 dark:text-slate-400">{{ $t('telemetry.healthySensors') }}</span>
          </div>
          <span class="text-3xs text-emerald-700 dark:text-emerald-400/80 font-medium truncate">{{ $t('telemetry.goodQualityDesc') }}</span>
        </div>
      </div>

      <!-- 4. Uncertain / Bad -->
      <div class="saas-card px-3 py-1.5 bg-white dark:bg-[#0F172A]/70 border border-amber-200 dark:border-amber-500/20 rounded-xl flex flex-col justify-between">
        <div class="text-3xs uppercase tracking-wider font-semibold text-amber-600 dark:text-amber-400 truncate">{{ $t('telemetry.uncertainBad') }}</div>
        <div class="flex items-baseline justify-between gap-1 mt-0.5">
          <div class="flex items-baseline space-x-1 font-mono">
            <span class="text-base sm:text-lg font-bold text-amber-600 dark:text-amber-400 leading-none">{{ stats.uncertainSensors }}</span>
            <span class="text-2xs text-rose-600 dark:text-rose-400 font-mono">/ {{ stats.badSensors }} {{ $t('telemetry.qualityBad') }}</span>
          </div>
          <span class="text-3xs text-amber-700 dark:text-amber-400/80 font-medium truncate">{{ $t('telemetry.attentionRequired') }}</span>
        </div>
      </div>

      <!-- 5. Stale Data -->
      <div class="saas-card px-3 py-1.5 bg-white dark:bg-[#0F172A]/70 border border-purple-200 dark:border-purple-500/20 rounded-xl flex flex-col justify-between">
        <div class="text-3xs uppercase tracking-wider font-semibold text-purple-600 dark:text-purple-400 truncate">{{ $t('telemetry.qualityStale') }}</div>
        <div class="flex items-baseline justify-between gap-1 mt-0.5">
          <div class="flex items-baseline space-x-1 font-mono">
            <span class="text-base sm:text-lg font-bold text-purple-600 dark:text-purple-400 leading-none">{{ stats.staleSensors }}</span>
            <span class="text-2xs text-slate-500 dark:text-slate-400">{{ $t('telemetry.timedOutSensors') }}</span>
          </div>
          <span class="text-3xs text-purple-700 dark:text-purple-400/80 font-medium truncate">{{ $t('telemetry.noDataReceived') }}</span>
        </div>
      </div>
    </div>

    <!-- Filters & Controls Bar in Single Line -->
    <div class="px-3 py-1.5 saas-card bg-white dark:bg-[#0F172A]/90 border border-slate-200 dark:border-slate-800 rounded-xl flex flex-wrap md:flex-nowrap gap-2 items-center justify-between flex-shrink-0 shadow-2xs">
      <div class="flex items-center gap-2 flex-1 min-w-0">
        <!-- Search Input -->
        <div class="relative flex-1 min-w-[140px] max-w-xs">
          <input
            v-model="searchQuery"
            type="text"
            :placeholder="$t('telemetry.searchPlaceholder')"
            class="w-full h-7 pl-8 pr-3 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-2xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 transition-colors"
          />
          <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>

        <!-- Filter Device -->
        <select
          v-model="selectedDeviceId"
          class="h-7 px-2.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-2xs text-slate-800 dark:text-white focus:outline-none focus:border-blue-500 max-w-[180px] truncate"
        >
          <option value="">{{ $t('telemetry.allDevices') }} ({{ devices.length }})</option>
          <option v-for="d in devices" :key="d.id" :value="d.id">
            {{ d.device_code || d.code }} - {{ d.device_name || d.name }}
          </option>
        </select>

        <!-- Filter Quality -->
        <select
          v-model="selectedQuality"
          class="h-7 px-2.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-2xs text-slate-800 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
        >
          <option value="">{{ $t('telemetry.allQualities') }}</option>
          <option value="GOOD">{{ $t('telemetry.goodQualityOnly') }}</option>
          <option value="BAD">{{ $t('telemetry.badQualityOnly') }}</option>
          <option value="UNCERTAIN">{{ $t('telemetry.uncertainOnly') }}</option>
          <option value="STALE">{{ $t('telemetry.staleOnly') }}</option>
        </select>

        <!-- Record Count Summary -->
        <div class="hidden md:flex items-center space-x-1 text-3xs font-mono text-slate-500 dark:text-slate-400 pl-2 border-l border-slate-200 dark:border-slate-800 whitespace-nowrap">
          <span>{{ filteredParameters.length }} / {{ parametersList.length }}</span>
          <span class="text-4xs text-slate-400 dark:text-slate-500 uppercase">{{ $t('parameters.parameters') || 'Sensors' }}</span>
        </div>
      </div>

      <!-- View Switcher -->
      <div class="flex items-center space-x-1.5 flex-shrink-0">
        <span class="text-3xs text-slate-400 dark:text-slate-500 uppercase font-semibold mr-0.5">{{ $t('telemetry.display') }}</span>
        <button
          @click="setViewMode('grid')"
          :class="viewMode === 'grid' ? 'bg-blue-600 text-white shadow-xs' : 'bg-slate-100 hover:bg-slate-200 text-slate-600 border border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:hover:text-white dark:border-transparent'"
          class="w-7 h-7 rounded-lg transition flex items-center justify-center cursor-pointer"
          :title="$t('telemetry.cardsGrid')"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <rect x="3" y="3" width="7" height="7" rx="1"></rect>
            <rect x="14" y="3" width="7" height="7" rx="1"></rect>
            <rect x="14" y="14" width="7" height="7" rx="1"></rect>
            <rect x="3" y="14" width="7" height="7" rx="1"></rect>
          </svg>
        </button>
        <button
          @click="setViewMode('table')"
          :class="viewMode === 'table' ? 'bg-blue-600 text-white shadow-xs' : 'bg-slate-100 hover:bg-slate-200 text-slate-600 border border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:hover:text-white dark:border-transparent'"
          class="w-7 h-7 rounded-lg transition flex items-center justify-center cursor-pointer"
          :title="$t('telemetry.dataTable')"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="filteredParameters.length === 0" class="saas-card p-8 text-center space-y-2 bg-white dark:bg-[#0F172A]/50 border border-slate-200 dark:border-slate-800 rounded-xl flex-1 flex flex-col items-center justify-center">
      <div class="w-10 h-10 rounded-full bg-slate-100 dark:bg-slate-800 flex items-center justify-center mx-auto text-slate-400">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="9"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
      </div>
      <h3 class="text-xs font-bold text-slate-900 dark:text-white">{{ $t('telemetry.noSensorsMatch') }}</h3>
      <p class="text-3xs text-slate-500 dark:text-slate-400 max-w-sm mx-auto">
        {{ $t('telemetry.verifyNotice') }}
      </p>
      <div class="pt-1">
        <router-link to="/monitoring/devices" class="px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-3xs font-semibold inline-block">
          {{ $t('telemetry.openDevices') }}
        </router-link>
      </div>
    </div>

    <!-- VIEW 1: SENSOR CARDS GRID -->
    <div v-else-if="viewMode === 'grid'" class="flex-1 min-h-0 overflow-y-auto pr-0.5">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2.5">
        <div
          v-for="item in filteredParameters"
          :key="item.param_id"
          :class="[
            liveBlinks[item.param_id] ? 'border-emerald-500 shadow-md shadow-emerald-500/10' : 'border-slate-200 dark:border-slate-800/80',
            item.quality === 'BAD' ? 'bg-rose-50/50 dark:bg-rose-950/10' : 'bg-white dark:bg-[#0F172A]'
          ]"
          class="saas-card p-3 space-y-2 rounded-xl border transition-all duration-200 hover:border-slate-300 dark:hover:border-slate-700 flex flex-col justify-between"
        >
          <div>
            <!-- Card Header: Device + Parameter Code -->
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex items-center space-x-1.5">
                  <span class="w-1.5 h-1.5 rounded-full" :class="item.device_status === 'CONNECTED' ? 'bg-emerald-500' : 'bg-slate-400'"></span>
                  <router-link :to="'/monitoring/devices/' + item.device_id" class="text-3xs font-mono font-bold text-slate-500 hover:text-blue-600 dark:text-slate-400 dark:hover:text-blue-400 uppercase tracking-wide truncate">
                    {{ item.device_code }}
                  </router-link>
                </div>
                <h3 class="text-xs font-bold text-slate-900 dark:text-white mt-0.5 truncate" :title="item.param_name">
                  {{ item.param_name }}
                </h3>
              </div>

              <!-- Status & Quality Badges -->
              <div class="flex items-center space-x-1.5 flex-shrink-0">
                <span
                  v-if="item.is_held_value"
                  class="px-1.5 py-0.5 rounded text-3xs font-mono font-bold uppercase bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-500/20 dark:text-amber-300 dark:border-amber-500/30 animate-pulse flex items-center space-x-1"
                  :title="$t('parameters.heldTooltip')"
                >
                  <span>🛡️</span>
                  <span>{{ $t('parameters.heldBadge') }}</span>
                </span>
                <div class="flex flex-col items-end space-y-0.5">
                  <span
                    :class="qualityBadgeClass(item.quality)"
                    class="px-1.5 py-0.5 rounded text-3xs font-mono font-bold uppercase flex items-center space-x-1"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="qualityDotClass(item.quality)"></span>
                    <span>{{ item.quality || 'WAITING' }}</span>
                  </span>
                  <span v-if="item.quality_reason" class="text-4xs font-mono text-amber-600 dark:text-amber-400/90 truncate max-w-[100px]" :title="item.quality_reason">
                    {{ item.quality_reason }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Parameter Code Tag -->
            <div class="mt-1 flex items-center space-x-2">
              <span class="px-1.5 py-0.2 rounded bg-slate-100 text-slate-700 border border-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-transparent font-mono text-3xs">
                {{ item.param_code }}
              </span>
              <span v-if="item.data_type" class="text-3xs font-mono text-slate-500 font-medium">
                {{ item.data_type }}
              </span>
            </div>

            <!-- Primary Metric Value -->
            <div class="mt-2.5 p-2 bg-slate-50/70 dark:bg-[#0B0F19] rounded-lg border border-slate-200 dark:border-slate-800/80 space-y-1.5">
              <div class="flex items-baseline justify-between">
                <div class="text-xl font-mono font-extrabold tracking-tight" :class="valueColorClass(item)">
                  {{ formatValue(item) }}
                </div>
                <div class="text-xs font-bold text-slate-600 dark:text-slate-400 font-mono">
                  {{ item.unit || '--' }}
                </div>
              </div>
              <!-- Formula Display if configured (Admin and Superadmin only) -->
              <div v-if="canAccessRawMode && item.formula_value !== undefined && item.formula_value !== null" class="pt-1.5 border-t border-slate-200 dark:border-slate-800/70 flex items-center justify-between text-3xs font-mono">
                <div class="flex items-center space-x-1 text-blue-600 dark:text-blue-400 min-w-0">
                  <span class="font-bold">ƒ(x):</span>
                  <span class="text-slate-600 dark:text-slate-400 truncate max-w-[100px]" :title="'Formula: ' + item.formula">{{ item.formula }}</span>
                </div>
                <div class="font-bold text-blue-700 dark:text-blue-300 bg-blue-50 dark:bg-blue-500/10 px-1.5 py-0.2 rounded border border-blue-200 dark:border-blue-500/20 whitespace-nowrap">
                  {{ formatFormula(item) }} {{ item.unit || '' }}
                </div>
              </div>
            </div>
          </div>

          <!-- Card Footer -->
          <div class="pt-1.5 border-t border-slate-200 dark:border-slate-800/80 space-y-1.5">
            <div class="flex items-center justify-between text-3xs font-mono text-slate-500 dark:text-slate-400">
              <span>{{ $t('telemetry.lastPacketLabel') }}</span>
              <span class="text-slate-700 dark:text-slate-300">{{ item.received_at ? formatRelative(item.received_at) : $t('common.noData') }}</span>
            </div>

            <!-- Quick Navigation Actions -->
            <div class="flex items-center justify-between gap-1 pt-0.5">
              <router-link
                :to="'/monitoring/devices/' + item.device_id + '?tab=telemetry'"
                class="text-3xs text-emerald-700 dark:text-emerald-400 hover:text-emerald-800 dark:hover:text-emerald-300 font-semibold px-1.5 py-0.5 rounded bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/40 hover:bg-emerald-100 dark:hover:bg-emerald-900/40 transition"
              >
                {{ $t('navigation.liveTelemetry') }}
              </router-link>
              <div class="flex items-center space-x-1">
                <router-link
                  :to="'/monitoring/devices/' + item.device_id + '?tab=history'"
                  class="text-3xs text-blue-700 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 font-medium px-1.5 py-0.5 rounded bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800/40 hover:bg-blue-100 dark:hover:bg-blue-900/40 transition"
                >
                  {{ $t('telemetry.chart') }}
                </router-link>
              <router-link
                v-if="canAccessRawMode"
                :to="'/monitoring/devices/' + item.device_id + '?tab=raw'"
                class="text-3xs text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200 font-medium px-1.5 py-0.5 rounded bg-slate-100 hover:bg-slate-200 border border-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-700 dark:border-transparent transition"
              >
                  {{ $t('telemetry.raw') }}
                </router-link>
                <router-link
                  :to="'/monitoring/analysis?device_id=' + item.device_id + '&parameter_id=' + item.param_id"
                  class="text-3xs text-indigo-700 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 font-medium px-1.5 py-0.5 rounded bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800/40 hover:bg-indigo-100 dark:hover:bg-indigo-900/40 transition"
                  :title="$t('analysis.openInAnalysis')"
                >
                  📊
                </router-link>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- VIEW 2: TELEMETRY DATA TABLE (Default View) -->
    <div v-else class="saas-card overflow-hidden border border-slate-200 dark:border-slate-800 bg-white dark:bg-[#0F172A] rounded-xl flex-1 min-h-0 flex flex-col shadow-xs">
      <div class="overflow-auto flex-1 min-h-0 telemetry-table-scroll">
        <table class="w-full text-left border-collapse text-xs">
          <thead class="sticky top-0 z-10 bg-slate-50 dark:bg-[#0B0F19] border-b border-slate-200 dark:border-slate-800 text-slate-500 dark:text-slate-400 text-3xs font-semibold uppercase tracking-wider">
            <tr>
              <th class="py-2 px-3">{{ $t('telemetry.device') }}</th>
              <th class="py-2 px-3">{{ $t('telemetry.parameterName') }}</th>
              <th class="py-2 px-2.5">{{ $t('common.code') }}</th>
              <th class="py-2 px-3 text-right">{{ $t('telemetry.currentValue') }}</th>
              <th v-if="canAccessRawMode" class="py-3 px-4 text-right">{{ $t('parameters.formula') }}</th>
              <th class="py-2 px-2.5">{{ $t('common.unit') }}</th>
              <th class="py-2 px-2.5 text-center">{{ $t('common.quality') }}</th>
              <th class="py-2 px-3">{{ $t('telemetry.lastTimestamp') }}</th>
              <th class="py-2 px-3 text-right">{{ $t('telemetry.inspect') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 font-sans">
            <tr
              v-for="item in filteredParameters"
              :key="item.param_id"
              :class="liveBlinks[item.param_id] ? 'bg-emerald-500/10' : 'hover:bg-slate-50 dark:hover:bg-slate-800/30'"
              class="transition-colors leading-tight"
            >
              <!-- Device -->
              <td class="py-1.5 px-3 font-mono text-xs whitespace-nowrap">
                <div class="flex items-center space-x-1.5">
                  <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="item.device_status === 'CONNECTED' ? 'bg-emerald-500' : 'bg-slate-400'"></span>
                  <router-link :to="'/monitoring/devices/' + item.device_id" class="text-blue-600 dark:text-blue-400 hover:underline font-semibold">
                    {{ item.device_code }}
                  </router-link>
                  <span v-if="item.device_name && item.device_name !== item.device_code" class="text-3xs text-slate-400 dark:text-slate-500 font-sans truncate max-w-[120px]" :title="item.device_name">
                    ({{ item.device_name }})
                  </span>
                </div>
              </td>

              <!-- Parameter Name -->
              <td class="py-1.5 px-3 font-medium text-slate-900 dark:text-white truncate max-w-[180px]" :title="item.param_name">
                {{ item.param_name }}
              </td>

              <!-- Code -->
              <td class="py-1.5 px-2.5 font-mono text-3xs text-slate-500 dark:text-slate-400 whitespace-nowrap">
                <span class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-transparent font-medium">
                  {{ item.param_code }}
                </span>
              </td>

              <!-- Value -->
              <td class="py-1.5 px-3 text-right font-mono font-bold text-xs sm:text-sm whitespace-nowrap" :class="valueColorClass(item)">
                <div class="flex items-center justify-end space-x-1.5">
                  <span
                    v-if="item.is_held_value"
                    class="px-1.5 py-0.2 rounded text-3xs font-mono font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30 animate-pulse"
                    :title="$t('parameters.heldTooltip')"
                  >
                    {{ $t('parameters.heldBadge') }}
                  </span>
                  <span>{{ formatValue(item) }}</span>
                </div>
              </td>

              <!-- Formula Result (Admin and Superadmin only) -->
              <td v-if="canAccessRawMode" class="py-1.5 px-3 text-right font-mono text-3xs whitespace-nowrap">
                <div v-if="item.formula_value !== undefined && item.formula_value !== null" class="text-blue-600 dark:text-blue-400 font-bold flex items-center justify-end space-x-1">
                  <span>{{ formatFormula(item) }}</span>
                  <span v-if="item.formula" class="text-slate-400 text-4xs truncate max-w-[90px] font-normal" :title="'Formula: ' + item.formula">
                    ({{ item.formula }})
                  </span>
                </div>
                <div v-else class="text-slate-400 dark:text-slate-600">--</div>
              </td>

              <!-- Unit -->
              <td class="py-1.5 px-2.5 font-mono text-2xs text-slate-500 dark:text-slate-400 whitespace-nowrap">
                {{ item.unit || '--' }}
              </td>

              <!-- Quality -->
              <td class="py-1.5 px-2.5 text-center whitespace-nowrap">
                <span
                  :class="qualityBadgeClass(item.quality)"
                  class="px-1.5 py-0.5 rounded text-3xs font-mono font-bold uppercase inline-flex items-center space-x-1"
                >
                  <span class="w-1.5 h-1.5 rounded-full" :class="qualityDotClass(item.quality)"></span>
                  <span>{{ item.quality || 'WAITING' }}</span>
                </span>
                <div v-if="item.quality_reason" class="text-4xs font-mono text-amber-600 dark:text-amber-400/90 truncate max-w-[90px] mx-auto mt-0.5" :title="item.quality_reason">
                  {{ item.quality_reason }}
                </div>
              </td>

              <!-- Timestamp -->
              <td class="py-1.5 px-3 font-mono text-3xs text-slate-500 dark:text-slate-400 whitespace-nowrap">
                <div class="flex items-center space-x-1">
                  <span class="text-slate-800 dark:text-slate-200 font-medium">{{ item.received_at ? formatRelative(item.received_at) : '--' }}</span>
                  <span v-if="item.received_at" class="text-slate-400 dark:text-slate-500 text-3xs">({{ formatTimestamp(item.received_at) }})</span>
                </div>
              </td>

              <!-- Actions -->
              <td class="py-1.5 px-3 text-right whitespace-nowrap space-x-1">
                <router-link
                  :to="'/monitoring/devices/' + item.device_id + '?tab=telemetry'"
                  class="inline-block text-emerald-700 dark:text-emerald-400 hover:text-emerald-800 dark:hover:text-emerald-300 text-3xs font-semibold px-1.5 py-0.5 rounded bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800/40 transition"
                >
                  {{ $t('devices.live') }}
                </router-link>
                <router-link
                  :to="'/monitoring/devices/' + item.device_id + '?tab=history'"
                  class="inline-block text-blue-700 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 text-3xs font-semibold px-1.5 py-0.5 rounded bg-blue-50 dark:bg-blue-950/40 border border-blue-200 dark:border-blue-800/40 transition"
                >
                  {{ $t('deviceDetail.history') }}
                </router-link>
                <router-link
                  :to="'/monitoring/analysis?device_id=' + item.device_id + '&parameter_id=' + item.param_id"
                  class="inline-block text-indigo-700 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 text-3xs font-semibold px-1.5 py-0.5 rounded bg-indigo-50 dark:bg-indigo-950/40 border border-indigo-200 dark:border-indigo-800/40 transition"
                  :title="$t('analysis.openInAnalysis')"
                >
                  {{ $t('analysis.breadcrumb') }}
                </router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'TelemetryMonitorView',
  data() {
    let initialMode = 'table';
    try {
      const stored = localStorage.getItem('telemetry_view_mode');
      if (stored === 'grid' || stored === 'table') {
        initialMode = stored;
      }
    } catch (e) {
      // fallback
    }
    return {
      devices: [],
      liveMap: {},
      liveBlinks: {},
      loading: false,
      searchQuery: '',
      selectedDeviceId: '',
      selectedQuality: '',
      viewMode: initialMode,
      lastPacketTime: null,
      pollTimer: null,
    };
  },
  computed: {
    parametersList() {
      const list = [];
      for (const dev of this.devices) {
        if (!dev.parameters) continue;
        for (const p of dev.parameters) {
          const live = this.liveMap[p.id] || {};
          const isHeld = live.is_held_value !== undefined ? live.is_held_value : !!p.is_current_held;
          const formulaVal = live.formula_value !== undefined && live.formula_value !== null ? live.formula_value : p.current_formula_value;
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
            value: live.value !== undefined ? live.value : p.current_value,
            value_text: live.value_text,
            formula: p.formula || live.formula || '',
            formula_value: formulaVal,
            is_held_value: isHeld,
            hold_enabled: p.hold_last_value_enabled,
            quality: live.quality || p.current_quality || 'UNKNOWN',
            quality_reason: (live.quality_reason && live.quality_reason !== 'NONE') ? live.quality_reason : (p.current_quality_reason && p.current_quality_reason !== 'NONE' ? p.current_quality_reason : ''),
            quality_flags: live.quality_flags || p.current_quality_flags || '',
            processed_value: live.processed_value !== undefined ? live.processed_value : p.current_processed_value,
            received_at: live.received_at || p.current_received_at || p.last_updated,
          });
        }
      }
      return list;
    },
    filteredParameters() {
      return this.parametersList.filter(item => {
        if (this.selectedDeviceId && item.device_id !== parseInt(this.selectedDeviceId, 10)) {
          return false;
        }
        if (this.selectedQuality && item.quality !== this.selectedQuality) {
          return false;
        }
        if (this.searchQuery) {
          const q = this.searchQuery.toLowerCase();
          const matchName = item.param_name && item.param_name.toLowerCase().includes(q);
          const matchCode = item.param_code && item.param_code.toLowerCase().includes(q);
          const matchDevice = item.device_code && item.device_code.toLowerCase().includes(q);
          const matchUnit = item.unit && item.unit.toLowerCase().includes(q);
          const matchFormula = item.formula && item.formula.toLowerCase().includes(q);
          const matchReason = item.quality_reason && item.quality_reason.toLowerCase().includes(q);
          if (!matchName && !matchCode && !matchDevice && !matchUnit && !matchFormula && !matchReason) return false;
        }
        return true;
      });
    },
    stats() {
      const activeDevices = this.devices.filter(d => d.connection_status === 'CONNECTED').length;
      const goodSensors = this.parametersList.filter(p => p.quality === 'GOOD').length;
      const uncertainSensors = this.parametersList.filter(p => p.quality === 'UNCERTAIN').length;
      const badSensors = this.parametersList.filter(p => p.quality === 'BAD').length;
      const staleSensors = this.parametersList.filter(p => p.quality === 'STALE').length;
      const total = this.parametersList.length;
      const healthPercent = total > 0 ? Math.round((goodSensors / total) * 1000) / 10 : 0;
      return {
        totalDevices: this.devices.length,
        activeDevices,
        goodSensors,
        uncertainSensors,
        badSensors,
        staleSensors,
        healthPercent,
      };
    },
    userRole() {
      const u = this.$store.state.user || {};
      return (u.role || '').toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    canAccessRawMode() {
      if (this.$store && this.$store.getters && this.$store.getters.canAccessRawTrendMode !== undefined) {
        return !!this.$store.getters.canAccessRawTrendMode;
      }
      return ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN'].includes(this.userRole);
    },
  },
  mounted() {
    this.fetchTelemetry();
    window.addEventListener('device-telemetry-event', this.onTelemetryEvent);
    this.pollTimer = setInterval(this.fetchLatestTelemetryAll, 10000);
  },
  beforeDestroy() {
    window.removeEventListener('device-telemetry-event', this.onTelemetryEvent);
    if (this.pollTimer) clearInterval(this.pollTimer);
  },
  methods: {
    setViewMode(mode) {
      this.viewMode = mode;
      try {
        localStorage.setItem('telemetry_view_mode', mode);
      } catch (e) {
        // ignore
      }
    },
    async fetchTelemetry() {
      this.loading = true;
      try {
        const res = await axios.get('/api/devices?limit=100');
        this.devices = res.data.data.items || [];
        await this.fetchLatestTelemetryAll();
      } catch (err) {
        console.error('Failed to load devices/telemetry:', err);
      } finally {
        this.loading = false;
      }
    },
    async fetchLatestTelemetryAll() {
      for (const dev of this.devices) {
        try {
          const res = await axios.get(`/api/devices/${dev.id}/telemetry/latest`);
          if (res.data && res.data.data) {
            for (const item of res.data.data) {
              this.$set(this.liveMap, item.parameter_id, item);
              if (item.received_at && (!this.lastPacketTime || item.received_at > this.lastPacketTime)) {
                this.lastPacketTime = item.received_at;
              }
            }
          }
        } catch (e) {
          // ignore error per device
        }
      }
    },
    onTelemetryEvent(e) {
      if (!e.detail) return;
      const data = e.detail;
      this.$set(this.liveMap, data.parameter_id, data);
      this.$set(this.liveBlinks, data.parameter_id, true);
      this.lastPacketTime = data.received_at;
      setTimeout(() => {
        this.$set(this.liveBlinks, data.parameter_id, false);
      }, 1200);
    },
    formatValue(item) {
      if (!item) return '--';
      const isOffline = item.device_status === 'OFFLINE' ||
        item.device_status === 'ERROR' ||
        item.device_status === 'DISCONNECTED' ||
        item.device_status === 'FAILED';
      const isBad = item.quality === 'BAD' && !item.is_held_value;
      const isCommErr = item.quality_reason === 'COMMUNICATION_ERROR' ||
        item.quality_reason === 'DEVICE_OFFLINE' ||
        item.quality_reason === 'TIMEOUT' ||
        item.quality_reason === 'SENSOR_FAULT';
      if (isOffline || isBad || isCommErr) {
        return '--';
      }
      if (item.value_text) return item.value_text;
      if (item.value !== null && item.value !== undefined) {
        return Number(item.value).toFixed(item.precision || 2);
      }
      return '--';
    },
    formatFormula(item) {
      if (!this.canAccessRawMode) return '--';
      if (item.formula_value !== null && item.formula_value !== undefined) {
        return Number(item.formula_value).toFixed(item.precision || 2);
      }
      return '--';
    },
    valueColorClass(item) {
      if (item.quality === 'BAD') return 'text-rose-600 dark:text-rose-400';
      if (item.quality === 'UNCERTAIN') return 'text-amber-600 dark:text-amber-400';
      if (item.value !== null && item.value !== undefined) return 'text-emerald-600 dark:text-emerald-400';
      return 'text-slate-400 dark:text-slate-500';
    },
    qualityBadgeClass(quality) {
      switch (quality) {
        case 'GOOD': return 'bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-500/15 dark:text-emerald-400 dark:border-emerald-500/30';
        case 'BAD': return 'bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-500/15 dark:text-rose-400 dark:border-rose-500/30';
        case 'UNCERTAIN': return 'bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-500/15 dark:text-amber-400 dark:border-amber-500/30';
        case 'STALE': return 'bg-purple-50 text-purple-700 border border-purple-200 dark:bg-purple-500/15 dark:text-purple-300 dark:border-purple-500/30';
        default: return 'bg-slate-100 text-slate-700 border border-slate-200 dark:bg-slate-500/15 dark:text-slate-400 dark:border-slate-500/30';
      }
    },
    qualityDotClass(quality) {
      switch (quality) {
        case 'GOOD': return 'bg-emerald-500 dark:bg-emerald-400';
        case 'BAD': return 'bg-rose-500 dark:bg-rose-400';
        case 'UNCERTAIN': return 'bg-amber-500 dark:bg-amber-400';
        case 'STALE': return 'bg-purple-500 dark:bg-purple-400';
        default: return 'bg-slate-400';
      }
    },
    formatTimestamp(ts) {
      if (!ts) return '--';
      try {
        const d = new Date(ts);
        return d.toLocaleTimeString('id-ID', { hour12: false });
      } catch (e) {
        return ts;
      }
    },
    formatRelative(ts) {
      if (!ts) return '--';
      try {
        const d = new Date(ts);
        const diff = Math.floor((new Date() - d) / 1000);
        if (diff < 2) return 'just now';
        if (diff < 60) return `${diff}s ago`;
        if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
        return `${Math.floor(diff / 3600)}h ago`;
      } catch (e) {
        return ts;
      }
    },
  },
};
</script>

<style scoped>
.telemetry-table-scroll th {
  padding-top: 0.45rem !important;
  padding-bottom: 0.45rem !important;
  padding-left: 0.65rem !important;
  padding-right: 0.65rem !important;
}
</style>
