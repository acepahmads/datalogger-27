<template>
  <div class="space-y-2.5 font-sans compact-telemetry-workbench">
    <!-- BAR 1: UNIFIED HEADER, MODE SWITCHER, QUICK TIME PILLS, LAYOUT & ACTIONS -->
    <div class="saas-card ta-bar1-card px-3 py-2 bg-white dark:bg-[#0F172A]/95 border border-slate-200 dark:border-slate-800 rounded-xl flex flex-wrap items-center justify-between gap-2 shadow-xs">
      <!-- Left: Compact Breadcrumb + Title + Segmented Mode + Time Presets -->
      <div class="ta-bar1-left flex flex-wrap items-center gap-2">
        <div class="ta-title-group flex items-center space-x-2 pr-1 min-w-0">
          <router-link
            :to="isOperator ? '/monitoring/dashboard-operator' : '/monitoring/devices'"
            class="p-1 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-600 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 transition flex-shrink-0"
            :title="$t('navigation.monitoring')"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </router-link>
          <h1 class="text-sm font-bold tracking-tight text-slate-900 dark:text-white whitespace-nowrap truncate">
            {{ $t('analysis.breadcrumb') }}
          </h1>
        </div>

        <span class="hidden md:inline-block h-4 w-px bg-slate-200 dark:bg-slate-800"></span>

        <!-- Segmented Data Source Mode Tabs -->
        <div class="ta-mode-tabs inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 text-2xs">
          <button
            @click="setMode('HISTORICAL')"
            :title="$t('analysis.historicalSemantic')"
            class="px-2.5 py-1 rounded-md font-semibold transition flex items-center justify-center space-x-1.5 min-w-0"
            :class="activeMode === 'HISTORICAL'
              ? 'bg-blue-600 text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
          >
            <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="activeMode === 'HISTORICAL' ? 'bg-white' : 'bg-blue-500'"></span>
            <span class="truncate">{{ $t('analysis.modeHistorical') }}</span>
          </button>

          <button
            v-if="!isOperator"
            @click="setMode('RAW')"
            :title="$t('analysis.rawSemantic')"
            class="px-2.5 py-1 rounded-md font-semibold transition flex items-center justify-center space-x-1.5 min-w-0"
            :class="activeMode === 'RAW'
              ? 'bg-emerald-600 text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
          >
            <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="activeMode === 'RAW' ? 'bg-white' : 'bg-emerald-500'"></span>
            <span class="truncate">{{ $t('analysis.modeRaw') }}</span>
          </button>

          <button
            @click="setMode('AGGREGATED')"
            :title="$t('analysis.aggregatedSemantic')"
            class="px-2.5 py-1 rounded-md font-semibold transition flex items-center justify-center space-x-1.5 min-w-0"
            :class="activeMode === 'AGGREGATED'
              ? 'bg-indigo-600 text-white shadow-xs'
              : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
          >
            <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="activeMode === 'AGGREGATED' ? 'bg-white' : 'bg-indigo-500'"></span>
            <span class="truncate">{{ $t('analysis.modeAggregated') }}</span>
          </button>
        </div>

        <!-- 1-Click Quick Time Range Pills -->
        <div class="ta-time-presets inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 text-3xs font-mono">
          <button
            v-for="tp in availableTimePresets"
            :key="tp.value"
            @click="selectTimePreset(tp.value)"
            class="px-2 py-1 rounded-md font-semibold transition text-center"
            :class="filters.timePreset === tp.value
              ? 'is-active-preset bg-white text-slate-900 shadow-xs border border-slate-300/80 dark:bg-slate-800 dark:text-white dark:border-slate-700'
              : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
          >
            {{ tp.label }}
          </button>
        </div>
        <span
          v-if="activeMode === 'RAW'"
          class="ta-raw-badge px-1.5 py-0.5 rounded bg-cyan-50 text-cyan-700 border border-cyan-200 dark:bg-cyan-950/50 dark:text-cyan-300 dark:border-cyan-800/60 text-4xs font-mono font-semibold"
        >
          Max {{ rawMaxRangeShort }}
        </span>
      </div>

      <!-- Right: Workspace Layout Switcher + Compact Actions -->
      <div class="ta-bar1-right flex flex-wrap items-center gap-1.5">
        <!-- Layout Density Mode (Split / Chart / Table — Desktop Bar 1; Mobile rendered on Trend/Table header) -->
        <div class="ta-desktop-layout-switcher inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 text-3xs font-semibold">
          <button
            @click="setWorkspaceLayout('SPLIT')"
            class="px-2 py-1 rounded-md transition"
            :class="workspaceLayout === 'SPLIT'
              ? 'bg-white text-slate-900 shadow-xs dark:bg-slate-800 dark:text-white'
              : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
            title="Split View: Trend Chart + Per-Series Stats + Data Table"
          >
            ⬒ {{ $t('analysis.layoutSplit') }}
          </button>
          <button
            @click="setWorkspaceLayout('CHART')"
            class="px-2 py-1 rounded-md transition"
            :class="workspaceLayout === 'CHART'
              ? 'bg-white text-slate-900 shadow-xs dark:bg-slate-800 dark:text-white'
              : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
            title="Chart Focus View"
          >
            📈 {{ $t('analysis.layoutChart') }}
          </button>
          <button
            @click="setWorkspaceLayout('TABLE')"
            class="px-2 py-1 rounded-md transition"
            :class="workspaceLayout === 'TABLE'
              ? 'bg-white text-slate-900 shadow-xs dark:bg-slate-800 dark:text-white'
              : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
            title="Table Focus View"
          >
            📋 {{ $t('analysis.layoutTable') }}
          </button>
        </div>

        <div class="ta-actions-cluster flex items-center gap-1.5">
          <!-- Refresh -->
          <button
            @click="fetchData"
            :disabled="loading"
            :title="$t('common.refresh')"
            class="px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 border border-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 dark:border-slate-700 text-2xs font-semibold transition flex items-center space-x-1 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span class="hidden xl:inline">{{ $t('common.refresh') }}</span>
          </button>

          <!-- Unified Export Dropdown (XLS, PDF, CSV, JSON + Chart PNG, JPG) -->
          <div ref="exportMenuWrap" class="relative">
            <button
              @click.stop="exportDropdownOpen = !exportDropdownOpen"
              :disabled="isExporting"
              class="px-2.5 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white border border-blue-500 text-2xs font-semibold transition flex items-center space-x-1.5 shadow-xs disabled:opacity-50"
            >
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
              </svg>
              <span>{{ $t('common.export') }}</span>
              <svg class="w-3 h-3 opacity-80" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>

            <div
              v-if="exportDropdownOpen"
              class="absolute right-0 mt-1.5 w-52 max-w-[calc(100vw-24px)] bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-700 rounded-xl shadow-2xl z-40 py-1.5 divide-y divide-slate-100 dark:divide-slate-800"
            >
              <!-- Section 1: Data Export -->
              <div class="pb-1">
                <div class="px-3 py-1 text-[10px] font-mono font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500">
                  {{ $t('analysis.exportDataGroup') }}
                </div>
                <button
                  @click="exportXLS"
                  :disabled="records.length === 0"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800 flex items-center justify-between disabled:opacity-40"
                >
                  <span>📊 {{ $t('analysis.exportXls') }}</span>
                  <span class="text-4xs font-mono text-emerald-600 dark:text-emerald-400 font-bold">XLS</span>
                </button>
                <button
                  @click="exportPDF"
                  :disabled="records.length === 0"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800 flex items-center justify-between disabled:opacity-40"
                >
                  <span>📑 {{ $t('analysis.exportPdf') }}</span>
                  <span class="text-4xs font-mono text-rose-600 dark:text-rose-400 font-bold">PDF</span>
                </button>
                <button
                  @click="exportCSV"
                  :disabled="records.length === 0"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800 flex items-center justify-between disabled:opacity-40"
                >
                  <span>📄 {{ $t('analysis.exportCsv') }}</span>
                  <span class="text-4xs font-mono text-slate-400 font-bold">CSV</span>
                </button>
                <button
                  @click="exportJSON"
                  :disabled="records.length === 0"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800 flex items-center justify-between disabled:opacity-40"
                >
                  <span>📦 {{ $t('analysis.exportJson') }}</span>
                  <span class="text-4xs font-mono text-slate-400 font-bold">JSON</span>
                </button>
              </div>

              <!-- Section 2: Chart Image Export -->
              <div class="pt-1">
                <div class="px-3 py-1 text-[10px] font-mono font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500">
                  {{ $t('analysis.exportChartGroup') }}
                </div>
                <button
                  @click="exportChart('png')"
                  :disabled="visibleSeries.length === 0"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800 flex items-center justify-between disabled:opacity-40"
                >
                  <span>🖼️ {{ $t('analysis.exportChartPng') }}</span>
                  <span class="text-4xs font-mono text-blue-600 dark:text-blue-400 font-bold">PNG</span>
                </button>
                <button
                  @click="exportChart('jpg')"
                  :disabled="visibleSeries.length === 0"
                  class="w-full text-left px-3 py-1.5 text-xs text-slate-700 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-800 flex items-center justify-between disabled:opacity-40"
                >
                  <span>🖼️ {{ $t('analysis.exportChartJpg') }}</span>
                  <span class="text-4xs font-mono text-amber-600 dark:text-amber-400 font-bold">JPG</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Superadmin Purge All Raw Telemetry -->
          <button
            v-if="isSuperAdmin && activeMode === 'RAW'"
            id="btn-analysis-purge-raw"
            @click="openPurgeRawModal"
            :disabled="isPurgingRaw || loading"
            class="px-2 sm:px-2.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 text-white border border-rose-500 text-2xs font-semibold transition flex items-center space-x-1 shadow-xs disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            <span class="hidden sm:inline">{{ $t('retention.purgeAllRawBtn') }}</span>
          </button>

          <!-- Superadmin Telemetry Injector Shortcut -->
          <router-link
            v-if="isSuperAdmin"
            :to="filters.deviceId ? ('/monitoring/telemetry-injection?device_id=' + filters.deviceId) : '/monitoring/telemetry-injection'"
            class="hidden sm:flex px-2.5 py-1.5 rounded-lg bg-violet-600/15 hover:bg-violet-600/25 text-violet-700 dark:text-violet-300 border border-violet-500/30 text-2xs font-semibold transition items-center space-x-1"
          >
            <span>⚡</span>
            <span class="hidden sm:inline">Suntik</span>
          </router-link>

          <!-- Shortcut to Live Telemetry (Hidden for Operator role; desktop/tablet bar button, mobile uses Drawer) -->
          <router-link
            v-if="!isOperator"
            to="/monitoring/telemetry"
            class="hidden sm:flex px-2.5 py-1.5 rounded-lg bg-emerald-50 hover:bg-emerald-100 text-emerald-700 border border-emerald-200 dark:bg-emerald-600/20 dark:hover:bg-emerald-600/30 dark:text-emerald-400 dark:border-emerald-500/30 text-2xs font-semibold transition items-center space-x-1.5"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 dark:bg-emerald-400 animate-pulse"></span>
            <span class="hidden xl:inline">{{ $t('navigation.liveTelemetry') }}</span>
            <span class="xl:hidden">Live</span>
          </router-link>
        </div>
      </div>
    </div>

    <!-- BAR 2: SLIM INLINE FILTERS + LIVE TELEMETRY KPI STRIP -->
    <div class="saas-card ta-bar2-card px-3 py-2 bg-white dark:bg-[#0F172A]/90 border border-slate-200 dark:border-slate-800 rounded-xl space-y-2 shadow-xs">
      <div class="ta-bar2-row flex flex-wrap items-center justify-between gap-2">
        <!-- Compact Inline Filter Controls -->
        <div class="ta-filter-grid flex flex-wrap items-center gap-1.5">
          <!-- Device Filter (Hidden for Operator) -->
          <select
            v-if="!isOperator"
            v-model="filters.deviceId"
            @change="onDeviceChange"
            :aria-label="$t('telemetry.device')"
            class="px-2.5 py-1 bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-800 dark:text-slate-200 focus:outline-none focus:border-blue-500 max-w-[190px] truncate"
          >
            <option value="">{{ $t('analysis.allDevices') }}</option>
            <option v-for="d in devicesList" :key="d.id" :value="d.id">
              {{ d.device_code }} — {{ d.device_name }}
            </option>
          </select>

          <!-- Parameter Filter -->
          <select
            v-model="filters.parameterId"
            @change="fetchData"
            :aria-label="$t('telemetry.parameterName')"
            class="px-2.5 py-1 bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-800 dark:text-slate-200 focus:outline-none focus:border-blue-500 max-w-[210px] truncate"
          >
            <option value="">{{ $t('analysis.allParameters') }}</option>
            <option v-for="p in availableParameters" :key="p.id" :value="p.id">
              {{ p.parameter_code }} ({{ p.parameter_name }})
            </option>
          </select>

          <!-- Quality Filter -->
          <select
            v-model="filters.quality"
            @change="fetchData"
            :aria-label="$t('analysis.qualityFilter')"
            class="px-2.5 py-1 bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-800 dark:text-slate-200 focus:outline-none focus:border-blue-500 truncate"
          >
            <option value="">{{ $t('analysis.allQualities') }}</option>
            <option value="GOOD">{{ $t('telemetry.qualityGood') }}</option>
            <option value="UNCERTAIN">{{ $t('telemetry.qualityUncertain') }}</option>
            <option value="BAD">{{ $t('telemetry.qualityBad') }}</option>
            <option value="STALE">{{ $t('telemetry.qualityStale') }}</option>
          </select>

          <!-- Contextual Aggregation Function OR Raw Plot Mode -->
          <select
            v-if="activeMode === 'AGGREGATED'"
            v-model="filters.aggFunction"
            @change="fetchData"
            class="px-2.5 py-1 bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs font-mono text-slate-800 dark:text-slate-200 focus:outline-none focus:border-blue-500 truncate"
          >
            <option value="">{{ $t('analysis.allFunctions') }}</option>
            <option value="AVG">Fn: AVG</option>
            <option value="MIN">Fn: MIN</option>
            <option value="MAX">Fn: MAX</option>
            <option value="SUM">Fn: SUM</option>
            <option value="COUNT">Fn: COUNT</option>
            <option value="FIRST">Fn: FIRST</option>
            <option value="LAST">Fn: LAST</option>
          </select>

          <select
            v-else-if="activeMode === 'RAW' && canAccessRawMode"
            v-model="rawChartMode"
            class="px-2.5 py-1 bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-800 dark:text-slate-200 focus:outline-none focus:border-blue-500 truncate"
          >
            <option value="scaled">{{ $t('analysis.viewScaledValue') }}</option>
            <option value="raw">{{ $t('analysis.viewRawValue') }}</option>
          </select>

          <!-- Page Size (Desktop in Bar 2; Mobile in Data Table footer/header) -->
          <select
            v-model="filters.pageSize"
            @change="fetchData"
            class="ta-pagesize-select px-2 py-1 bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs font-mono text-slate-700 dark:text-slate-300 focus:outline-none focus:border-blue-500"
          >
            <option :value="50">50 / pg</option>
            <option :value="100">100 / pg</option>
            <option :value="250">250 / pg</option>
            <option :value="500">500 / pg</option>
          </select>
        </div>

        <!-- Inline Monospace KPI Ribbon (Replaces 6 Tall KPI Cards) -->
        <div
          v-if="activeSingleSeriesStats"
          class="ta-kpi-strip flex flex-wrap items-center justify-between sm:justify-start gap-x-2.5 gap-y-1 px-2.5 py-1.5 sm:py-1 rounded-lg bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 text-2xs font-mono tabular-nums"
        >
          <span class="font-bold text-slate-900 dark:text-white">
            {{ activeSingleSeriesStats.label }}:
            <span class="text-blue-600 dark:text-blue-400">{{ activeSingleSeriesStats.latest !== null ? activeSingleSeriesStats.latest : '--' }}</span>
            <span class="text-3xs text-slate-500">{{ activeSingleSeriesStats.unit }}</span>
          </span>
          <span class="hidden sm:inline text-slate-300 dark:text-slate-700">|</span>
          <div class="flex items-center gap-2">
            <span class="text-emerald-600 dark:text-emerald-400">Min <b>{{ activeSingleSeriesStats.min !== null ? activeSingleSeriesStats.min : '--' }}</b></span>
            <span class="text-rose-600 dark:text-rose-400">Max <b>{{ activeSingleSeriesStats.max !== null ? activeSingleSeriesStats.max : '--' }}</b></span>
            <span class="text-blue-600 dark:text-blue-400">Avg <b>{{ activeSingleSeriesStats.avg !== null ? activeSingleSeriesStats.avg : '--' }}</b></span>
          </div>
          <span class="hidden sm:inline text-slate-300 dark:text-slate-700">|</span>
          <div class="flex items-center gap-1.5 ml-auto sm:ml-0">
            <span class="text-slate-600 dark:text-slate-300">n=<b>{{ activeSingleSeriesStats.sampleCount }}</b></span>
            <span class="px-1.5 py-0.5 sm:py-0.2 rounded-md bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 font-bold">{{ activeSingleSeriesStats.goodPercent }}% GOOD</span>
          </div>
        </div>

        <div
          v-else
          class="ta-kpi-strip flex flex-wrap items-center justify-between sm:justify-start gap-x-2.5 gap-y-1 px-2.5 py-1.5 sm:py-1 rounded-lg bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 text-2xs font-mono tabular-nums"
        >
          <div class="flex items-center gap-2 min-w-0">
            <span class="text-slate-700 dark:text-slate-300">
              {{ $t('analysis.activeSeries') }}: <b class="text-slate-900 dark:text-white">{{ visibleSeries.length }}/{{ seriesList.length }}</b>
            </span>
            <span class="hidden sm:inline text-slate-300 dark:text-slate-700">|</span>
            <span class="sm:hidden text-slate-400 dark:text-slate-600">&middot;</span>
            <span class="text-slate-700 dark:text-slate-300">
              {{ $t('analysis.samples') }}: <b class="text-slate-900 dark:text-white">{{ totalRecords }}</b>
            </span>
          </div>
          <span class="hidden sm:inline text-slate-300 dark:text-slate-700">|</span>
          <span class="px-2 sm:px-1.5 py-0.5 sm:py-0.2 rounded-full sm:rounded bg-emerald-500/15 sm:bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 font-bold ml-auto sm:ml-0">
            {{ allSeriesOverview.goodPercent }}% GOOD
          </span>
          <template v-if="allSeriesOverview.latestMeasurement">
            <span class="text-slate-300 dark:text-slate-700 hidden sm:inline">|</span>
            <span class="text-slate-500 dark:text-slate-400 hidden sm:inline truncate max-w-[240px]">
              {{ (!isOperator && allSeriesOverview.latestMeasurement.deviceCode ? allSeriesOverview.latestMeasurement.deviceCode + ':' : '') + allSeriesOverview.latestMeasurement.paramCode }}
              <b class="text-slate-900 dark:text-white">{{ formatVal(allSeriesOverview.latestMeasurement.value) }}</b>
              {{ allSeriesOverview.latestMeasurement.unit || '' }}
              <span class="text-4xs">({{ formatRelative(allSeriesOverview.latestMeasurement.timestamp) }})</span>
            </span>
          </template>
        </div>
      </div>

      <!-- Compact Custom Time Range Row -->
      <div v-if="filters.timePreset === 'custom'" class="pt-1.5 border-t border-slate-200 dark:border-slate-800 grid grid-cols-1 sm:flex sm:flex-wrap items-center gap-2 text-xs">
        <div class="flex items-center justify-between sm:justify-start space-x-1.5">
          <span class="text-slate-500 dark:text-slate-400 text-3xs uppercase font-bold">{{ $t('analysis.startDate') }}:</span>
          <input
            type="datetime-local"
            step="1"
            v-model="filters.startTime"
            class="flex-1 sm:flex-initial min-w-0 px-2 py-1 sm:py-0.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-800 dark:text-slate-200 font-mono focus:outline-none focus:border-blue-500"
          />
        </div>
        <div class="flex items-center justify-between sm:justify-start space-x-1.5">
          <span class="text-slate-500 dark:text-slate-400 text-3xs uppercase font-bold">{{ $t('analysis.endDate') }}:</span>
          <input
            type="datetime-local"
            step="1"
            v-model="filters.endTime"
            class="flex-1 sm:flex-initial min-w-0 px-2 py-1 sm:py-0.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-800 dark:text-slate-200 font-mono focus:outline-none focus:border-blue-500"
          />
        </div>
        <button
          @click="fetchData"
          class="w-full sm:w-auto px-2.5 py-1.5 sm:py-0.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold transition text-center"
        >
          {{ $t('common.apply') || 'Apply' }}
        </button>
      </div>
    </div>

    <!-- Purge Success Banner -->
    <div
      v-if="purgeSuccessMsg"
      class="px-3 py-2 rounded-xl bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-between text-xs text-emerald-700 dark:text-emerald-300"
    >
      <div class="flex items-center space-x-2">
        <svg class="w-4 h-4 text-emerald-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
        </svg>
        <span class="font-semibold">{{ purgeSuccessMsg }}</span>
      </div>
      <button @click="purgeSuccessMsg = ''" class="text-emerald-600 dark:text-emerald-400 hover:opacity-75 text-2xs font-bold px-2 py-0.5">✕</button>
    </div>

    <!-- ZONE 3A: ADAPTIVE CHART + DOCKED PER-SERIES MATRIX -->
    <div
      v-show="workspaceLayout !== 'TABLE'"
      class="grid grid-cols-1 lg:grid-cols-12 gap-2.5 items-stretch"
    >
      <!-- Multi-Series SVG Trend Chart Panel -->
      <div
        :class="workspaceLayout === 'CHART' ? 'lg:col-span-12' : 'lg:col-span-8'"
        class="saas-card bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden flex flex-col shadow-xs"
      >
        <div class="px-3 py-2 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-1.5 sm:gap-2">
          <div class="flex items-center space-x-1.5 sm:space-x-2 min-w-0">
            <span class="w-2 h-2 rounded-full shrink-0" :class="activeMode === 'HISTORICAL' ? 'bg-blue-500' : activeMode === 'RAW' ? 'bg-emerald-500' : 'bg-indigo-500'"></span>
            <h3 class="text-[11px] sm:text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider truncate">{{ $t('analysis.trendChart') }}</h3>
            <span class="text-3xs text-slate-500 dark:text-slate-400 font-mono truncate hidden sm:inline">
              <span v-if="filters.parameterId">• {{ activeParameterLabel }}</span>
              <span v-else>• {{ visibleSeries.length }}/{{ seriesList.length }} series</span>
              <span v-if="activeMode === 'RAW'" class="ml-1">({{ rawChartMode === 'raw' ? 'Raw Reg' : 'Scaled' }})</span>
            </span>
          </div>

          <!-- Quick Chart Image Export (PNG/JPG) + Mobile Layout Switcher + Scale Switcher + Reset -->
          <div class="flex flex-wrap items-center gap-1 sm:gap-1.5 text-3xs font-mono">
            <!-- Mobile Compact Layout Switcher (< 768px) -->
            <div class="md:hidden inline-flex items-center bg-slate-100 dark:bg-[#0B0F19] rounded-lg p-0.5 border border-slate-200 dark:border-slate-800 font-sans font-semibold">
              <button
                @click="setWorkspaceLayout('SPLIT')"
                class="px-1.5 py-0.5 rounded transition"
                :class="workspaceLayout === 'SPLIT' ? 'bg-white text-slate-900 shadow-xs dark:bg-slate-800 dark:text-white' : 'text-slate-500 dark:text-slate-400'"
              >
                {{ $t('analysis.layoutSplit') }}
              </button>
              <button
                @click="setWorkspaceLayout('CHART')"
                class="px-1.5 py-0.5 rounded transition"
                :class="workspaceLayout === 'CHART' ? 'bg-white text-slate-900 shadow-xs dark:bg-slate-800 dark:text-white' : 'text-slate-500 dark:text-slate-400'"
              >
                {{ $t('analysis.layoutChart') }}
              </button>
              <button
                @click="setWorkspaceLayout('TABLE')"
                class="px-1.5 py-0.5 rounded transition"
                :class="workspaceLayout === 'TABLE' ? 'bg-white text-slate-900 shadow-xs dark:bg-slate-800 dark:text-white' : 'text-slate-500 dark:text-slate-400'"
              >
                {{ $t('analysis.layoutTable') }}
              </button>
            </div>

            <!-- Direct 1-Click PNG / JPG Chart Export (>= 640px; on mobile available inside unified Export menu) -->
            <div class="hidden sm:inline-flex items-center bg-slate-100 dark:bg-[#0B0F19] rounded-lg p-0.5 border border-slate-200 dark:border-slate-800">
              <button
                @click="exportChart('png')"
                :disabled="visibleSeries.length === 0 || isExporting"
                :title="$t('analysis.exportChartPng')"
                class="px-1.5 py-0.5 rounded font-bold text-blue-600 hover:bg-white dark:text-blue-400 dark:hover:bg-slate-800 transition disabled:opacity-40"
              >
                ⬇ PNG
              </button>
              <button
                @click="exportChart('jpg')"
                :disabled="visibleSeries.length === 0 || isExporting"
                :title="$t('analysis.exportChartJpg')"
                class="px-1.5 py-0.5 rounded font-bold text-amber-600 hover:bg-white dark:text-amber-400 dark:hover:bg-slate-800 transition disabled:opacity-40"
              >
                ⬇ JPG
              </button>
            </div>

            <!-- Scale switcher buttons -->
            <div class="inline-flex items-center bg-slate-100 dark:bg-[#0B0F19] rounded-lg p-0.5 border border-slate-200 dark:border-slate-800">
              <button
                @click="setChartScaleMode('normalized')"
                class="px-1.5 sm:px-2 py-0.5 rounded transition font-semibold"
                :class="chartScaleMode === 'normalized' ? 'bg-white text-slate-900 border border-slate-300 shadow-xs dark:bg-amber-500/20 dark:text-amber-300 dark:border-amber-500/40' : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
              >
                Norm %
              </button>
              <button
                @click="setChartScaleMode('native')"
                class="px-1.5 sm:px-2 py-0.5 rounded transition font-semibold"
                :class="chartScaleMode === 'native' ? 'bg-white text-slate-900 border border-slate-300 shadow-xs dark:bg-blue-500/20 dark:text-blue-300 dark:border-blue-500/40' : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
              >
                Native
              </button>
            </div>

            <button
              v-if="focusedSeriesId || hiddenSeriesIds.length > 0"
              @click="resetSeriesFilters"
              class="px-2 py-0.5 rounded bg-blue-50 hover:bg-blue-100 text-blue-600 border border-blue-200 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-blue-400 dark:border-slate-700 font-semibold transition"
            >
              ↺ {{ $t('analysis.showAllSeries') }}
            </button>
          </div>
        </div>

        <!-- Mobile Interactive Series Legend Strip (< 640px) -->
        <div
          v-if="seriesList.length > 0"
          class="sm:hidden px-3 py-1.5 bg-slate-50/70 dark:bg-[#0B0F19]/80 border-b border-slate-100 dark:border-slate-800/70 flex items-center gap-2 overflow-x-auto no-scrollbar"
        >
          <button
            v-for="s in seriesList"
            :key="'mob-chip-' + s.id"
            type="button"
            @click="toggleFocusSeries(s.id)"
            class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-3xs font-mono font-semibold whitespace-nowrap border transition shrink-0"
            :class="[
              hiddenSeriesIds.includes(s.id) ? 'opacity-40 line-through border-slate-200 dark:border-slate-800' : '',
              focusedSeriesId === s.id
                ? 'bg-blue-600/15 text-blue-700 dark:text-blue-300 border-blue-500/40 ring-1 ring-blue-500/20'
                : 'bg-white dark:bg-slate-900/80 text-slate-700 dark:text-slate-300 border-slate-200/80 dark:border-slate-800'
            ]"
          >
            <span class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: s.color }"></span>
            <span>{{ (!isOperator && s.deviceCode ? s.deviceCode + ':' : '') + s.paramCode }}</span>
          </button>
        </div>

        <!-- SVG Chart Canvas Area -->
        <div
          ref="chartContainer"
          :class="workspaceLayout === 'CHART' ? 'h-64 sm:h-80' : 'h-[215px] sm:h-60'"
          class="w-full bg-white dark:bg-[#0B0F19] relative select-none overflow-hidden flex-1"
        >
          <!-- Empty State -->
          <div v-if="seriesList.length === 0 || visibleSeries.length === 0" class="w-full h-full flex flex-col items-center justify-center space-y-1.5 text-center p-4">
            <svg class="w-8 h-8 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
            </svg>
            <p class="text-xs text-slate-600 dark:text-slate-400 font-medium">{{ $t('analysis.noRecords') }}</p>
            <p class="text-3xs text-slate-500">{{ isOperator ? ($t('analysis.noRecordsSubtitleOperator') || $t('analysis.noRecordsSubtitle')) : $t('analysis.noRecordsSubtitle') }}</p>
          </div>

          <!-- SVG Multi-Series Chart -->
          <svg
            v-else
            ref="chartSvg"
            class="w-full h-full overflow-visible chart-svg"
            :viewBox="`0 0 ${containerWidth} ${containerHeight}`"
            @mousemove="onChartMouseMove"
            @mouseleave="onChartMouseLeave"
            @touchstart.passive="onChartTouchMove"
            @touchmove.passive="onChartTouchMove"
          >
            <defs>
              <linearGradient
                v-for="s in chartPlotData.seriesPlots"
                :key="'grad-' + s.id"
                :id="'grad-' + s.id"
                x1="0"
                y1="0"
                x2="0"
                y2="1"
              >
                <stop offset="0%" :stop-color="s.color" stop-opacity="0.08" />
                <stop offset="60%" :stop-color="s.color" stop-opacity="0.02" />
                <stop offset="100%" :stop-color="s.color" stop-opacity="0.00" />
              </linearGradient>
            </defs>

            <!-- Horizontal Grid Lines & Y-Axis Labels -->
            <g class="grid-lines">
              <g v-for="(tick, idx) in chartPlotData.gridTicks" :key="'grid-' + idx">
                <line
                  :x1="chartPlotData.padLeft"
                  :y1="tick.y"
                  :x2="containerWidth - chartPlotData.padRight"
                  :y2="tick.y"
                  :stroke="isDark ? '#1E293B' : '#E2E8F0'"
                  stroke-dasharray="4 4"
                  stroke-width="0.8"
                  stroke-opacity="0.8"
                />
                <text
                  :x="chartPlotData.padLeft - 8"
                  :y="tick.y + 4"
                  text-anchor="end"
                  :fill="isDark ? '#94A3B8' : '#64748B'"
                  font-size="10"
                  font-family="monospace"
                >
                  {{ tick.label }}
                </text>
              </g>
            </g>

            <!-- Vertical Ticks & Timestamps on X-Axis -->
            <g class="time-ticks">
              <g v-for="(tick, idx) in chartPlotData.timeTicks" :key="'ttick-' + idx">
                <line
                  :x1="tick.x"
                  :y1="containerHeight - chartPlotData.padBottom"
                  :x2="tick.x"
                  :y2="containerHeight - chartPlotData.padBottom + 5"
                  :stroke="isDark ? '#334155' : '#CBD5E1'"
                  stroke-width="1"
                />
                <text
                  :x="tick.x"
                  :y="containerHeight - 12"
                  text-anchor="middle"
                  :fill="isDark ? '#94A3B8' : '#64748B'"
                  font-size="10"
                  font-family="monospace"
                >
                  {{ tick.label }}
                </text>
              </g>
            </g>

            <!-- Area Fills (Rendered only when 1 series is visible) -->
            <g v-if="chartPlotData.seriesPlots.length === 1">
              <path
                v-for="s in chartPlotData.seriesPlots"
                :key="'area-' + s.id"
                :d="s.areaPath"
                :fill="`url(#grad-${s.id})`"
              />
            </g>

            <!-- Distinct Trend Lines for Each Series -->
            <g class="series-lines">
              <path
                v-for="s in chartPlotData.seriesPlots"
                :key="'line-' + s.id"
                :d="s.linePath"
                fill="none"
                :stroke="s.color"
                :stroke-width="hoveredSeriesId === s.id || focusedSeriesId === s.id ? 2.2 : 1.8"
                :opacity="focusedSeriesId && focusedSeriesId !== s.id ? 0.2 : hoveredSeriesId && hoveredSeriesId !== s.id ? 0.35 : 1"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="transition-all duration-150"
              />
            </g>

            <!-- Scatter Dots -->
            <g v-if="totalRecords <= 160" class="series-dots">
              <template v-for="s in chartPlotData.seriesPlots">
                <circle
                  v-for="(pt, idx) in s.points"
                  :key="'pt-' + s.id + '-' + idx"
                  :cx="pt.x"
                  :cy="pt.y"
                  :r="activePoint && activePoint.seriesId === s.id && activePoint.record.id === pt.record.id ? 5 : 2.5"
                  :fill="activePoint && activePoint.seriesId === s.id && activePoint.record.id === pt.record.id ? '#FFFFFF' : s.color"
                  :stroke="s.color"
                  stroke-width="1.5"
                  :opacity="focusedSeriesId && focusedSeriesId !== s.id ? 0.2 : 1"
                  class="transition-all duration-75"
                />
              </template>
            </g>

            <!-- Hover Crosshair & Tooltip Indicator -->
            <g v-if="activePoint" class="chart-hover-crosshair">
              <line
                :x1="activePoint.x"
                :y1="chartPlotData.padTop"
                :x2="activePoint.x"
                :y2="containerHeight - chartPlotData.padBottom"
                :stroke="isDark ? '#94A3B8' : '#64748B'"
                stroke-width="1"
                stroke-dasharray="3 3"
              />
              <circle
                :cx="activePoint.x"
                :cy="activePoint.y"
                r="5.5"
                fill="#FFFFFF"
                :stroke="activePoint.color"
                stroke-width="2.5"
              />
            </g>
          </svg>

          <!-- Interactive Floating Tooltip -->
          <div
            v-if="activePoint"
            class="absolute z-20 pointer-events-none bg-white/95 dark:bg-[#0F172A]/95 border border-slate-200 dark:border-slate-700/80 rounded-xl px-3 py-2 text-xs shadow-2xl text-slate-800 dark:text-slate-200 font-sans backdrop-blur-md max-w-[205px]"
            :style="{
              left: `${Math.min(Math.max(activePoint.x - 80, 8), Math.max(8, containerWidth - 205))}px`,
              top: `${Math.max(activePoint.y - 80, 8)}px`
            }"
          >
            <div class="flex items-center space-x-2 text-3xs text-slate-500 dark:text-slate-400 font-mono">
              <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: activePoint.color }"></span>
              <span>{{ formatTimestamp(activePoint.timestamp) }}</span>
            </div>
            <div class="flex items-baseline space-x-1.5 mt-0.5">
              <span class="text-sm font-bold text-slate-900 dark:text-white font-mono">{{ formatVal(activePoint.value) }}</span>
              <span class="text-2xs font-mono font-bold text-slate-600 dark:text-slate-300">{{ activePoint.unit || '' }}</span>
              <span class="ml-1 px-1.5 py-0.2 rounded text-4xs font-mono font-bold" :class="qualityBadgeClass(activePoint.quality)">
                {{ activePoint.quality }}
              </span>
            </div>
            <div class="text-3xs text-slate-700 dark:text-slate-300 font-medium truncate max-w-[190px] mt-0.5">
              <span v-if="!isOperator">{{ activePoint.deviceCode }}: </span>{{ activePoint.paramCode }}
            </div>
            <div class="text-4xs text-slate-500 dark:text-slate-400 truncate max-w-[190px]">
              {{ activePoint.paramName }}<span v-if="!isOperator && activePoint.deviceName"> ({{ activePoint.deviceName }})</span>
            </div>
            <div v-if="activePoint.qualityReason && activePoint.qualityReason !== 'NONE'" class="text-4xs text-rose-600 dark:text-rose-400 font-mono mt-0.5">
              Reason: {{ activePoint.qualityReason }}
            </div>
            <div v-if="canAccessRawMode && activeMode === 'RAW' && activePoint.rawHex" class="text-4xs text-emerald-600 dark:text-emerald-400 font-mono mt-0.5">
              Hex: {{ activePoint.rawHex }} (Reg: {{ activePoint.rawValue }})
            </div>
          </div>
        </div>
      </div>

      <!-- Docked Per-Parameter Stats Matrix (Replaces Separate Legend + Bottom Breakdown Cards) -->
      <div
        :class="workspaceLayout === 'CHART' ? 'lg:col-span-12' : 'lg:col-span-4'"
        class="saas-card bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden flex flex-col shadow-xs"
      >
        <div class="px-3 py-2 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between gap-2">
          <div class="flex items-center space-x-1.5">
            <h4 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">{{ $t('analysis.perSeriesStats') }}</h4>
            <span class="px-1.5 py-0.2 rounded bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300 text-4xs font-mono font-bold">
              {{ seriesList.length }}
            </span>
          </div>
          <span class="text-4xs font-mono text-slate-400 dark:text-slate-500 truncate">{{ $t('analysis.clickMuteSoloHint') }}</span>
        </div>

        <div
          v-if="seriesList.length > 0"
          :class="workspaceLayout === 'CHART' ? 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 divide-y sm:divide-y-0 sm:divide-x divide-slate-100 dark:divide-slate-800/80' : 'divide-y divide-slate-100 dark:divide-slate-800/80 max-h-60 overflow-y-auto flex-1'"
        >
          <div
            v-for="s in seriesList"
            :key="'matrix-' + s.id"
            @mouseenter="hoveredSeriesId = s.id"
            @mouseleave="hoveredSeriesId = null"
            class="px-2.5 py-1.5 flex items-center justify-between gap-2 text-2xs font-mono tabular-nums transition hover:bg-slate-50 dark:hover:bg-slate-800/40 select-none"
            :class="[
              hiddenSeriesIds.includes(s.id) ? 'opacity-40 line-through' : '',
              focusedSeriesId === s.id ? 'bg-blue-50/70 dark:bg-blue-950/30' : ''
            ]"
          >
            <!-- Toggle Visibility on Row Click -->
            <div @click="toggleSeriesVisibility(s.id)" class="flex items-center space-x-2 min-w-0 flex-1 cursor-pointer">
              <span
                class="w-2.5 h-2.5 rounded-full shrink-0 transition-transform"
                :style="{ backgroundColor: s.color }"
                :class="{ 'scale-125 ring-2 ring-slate-400/50': hoveredSeriesId === s.id }"
              ></span>
              <div class="min-w-0">
                <div class="font-bold text-slate-800 dark:text-slate-200 truncate text-2xs">
                  <span v-if="!isOperator">{{ s.deviceCode }}: </span>{{ s.paramCode }}
                  <span v-if="s.unit" class="text-4xs font-normal text-slate-400">({{ s.unit }})</span>
                </div>
                <div class="text-4xs text-slate-500 dark:text-slate-400 truncate font-sans">
                  <span class="text-emerald-600 dark:text-emerald-400">↓{{ s.stats.min !== null ? s.stats.min : '--' }}</span>
                  <span class="mx-1 text-rose-600 dark:text-rose-400">↑{{ s.stats.max !== null ? s.stats.max : '--' }}</span>
                  <span class="text-blue-600 dark:text-blue-400">μ{{ s.stats.avg !== null ? s.stats.avg : '--' }}</span>
                </div>
              </div>
            </div>

            <!-- Right: Latest + %Good + Solo Button -->
            <div class="flex items-center space-x-1.5 shrink-0">
              <div class="text-right">
                <div class="font-bold text-xs text-slate-900 dark:text-white leading-tight">
                  {{ s.stats.latest !== null ? s.stats.latest : '--' }}
                </div>
                <div class="text-[9px] text-emerald-600 dark:text-emerald-400 leading-tight">
                  {{ s.stats.goodPercent }}%G
                </div>
              </div>
              <button
                @click.stop="toggleFocusSeries(s.id)"
                class="px-1.5 py-0.5 rounded border text-3xs font-bold transition"
                :class="focusedSeriesId === s.id
                  ? 'bg-blue-600 text-white border-blue-500'
                  : 'bg-slate-100 hover:bg-slate-200 text-slate-500 border-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-700'"
                :title="focusedSeriesId === s.id ? 'Exit focus' : 'Solo focus this series'"
              >
                {{ focusedSeriesId === s.id ? '✕' : '⊙' }}
              </button>
            </div>
          </div>
        </div>

        <div v-else class="p-4 text-center text-slate-500 text-3xs font-mono flex-1 flex items-center justify-center">
          {{ loading ? $t('analysis.loadingData') : $t('analysis.noRecords') }}
        </div>
      </div>
    </div>

    <!-- ZONE 3B: DENSE STICKY-HEADER TELEMETRY DATA TABLE -->
    <div
      v-show="workspaceLayout !== 'CHART'"
      class="saas-card overflow-hidden border border-slate-200 dark:border-slate-800 bg-white dark:bg-[#0F172A] rounded-xl space-y-0 shadow-xs"
    >
      <!-- Compact Table Header, Quick Data Export, Search & Inline Pagination -->
      <div class="px-3 py-2 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-2">
        <div class="flex items-center justify-between sm:justify-start w-full sm:w-auto gap-2">
          <div class="flex items-center space-x-2">
            <h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">{{ $t('analysis.dataTable') }}</h3>
            <span class="px-1.5 py-0.5 rounded bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300 border border-slate-200 dark:border-transparent text-3xs font-mono font-bold">
              {{ totalRecords }} items
            </span>
          </div>

          <!-- Mobile Layout Switcher when in Table-Only View (< 768px) -->
          <div
            v-if="workspaceLayout === 'TABLE'"
            class="md:hidden inline-flex items-center bg-slate-100 dark:bg-[#0B0F19] rounded-lg p-0.5 border border-slate-200 dark:border-slate-800 text-3xs font-semibold"
          >
            <button
              @click="setWorkspaceLayout('SPLIT')"
              class="px-1.5 py-0.5 rounded text-slate-600 dark:text-slate-300"
            >
              {{ $t('analysis.layoutSplit') }}
            </button>
            <button
              @click="setWorkspaceLayout('CHART')"
              class="px-1.5 py-0.5 rounded text-slate-600 dark:text-slate-300"
            >
              {{ $t('analysis.layoutChart') }}
            </button>
            <button
              @click="setWorkspaceLayout('TABLE')"
              class="px-1.5 py-0.5 rounded bg-white text-slate-900 shadow-xs dark:bg-slate-800 dark:text-white"
            >
              {{ $t('analysis.layoutTable') }}
            </button>
          </div>

          <!-- Quick 1-Click XLS / PDF Data Buttons -->
          <div class="hidden sm:inline-flex items-center bg-slate-100 dark:bg-[#0B0F19] rounded-lg p-0.5 border border-slate-200 dark:border-slate-800 text-3xs font-mono">
            <button
              @click="exportXLS"
              :disabled="records.length === 0 || isExporting"
              :title="$t('analysis.exportXls')"
              class="px-1.5 py-0.5 rounded font-bold text-emerald-600 hover:bg-white dark:text-emerald-400 dark:hover:bg-slate-800 transition disabled:opacity-40"
            >
              ⬇ XLS
            </button>
            <button
              @click="exportPDF"
              :disabled="records.length === 0 || isExporting"
              :title="$t('analysis.exportPdf')"
              class="px-1.5 py-0.5 rounded font-bold text-rose-600 hover:bg-white dark:text-rose-400 dark:hover:bg-slate-800 transition disabled:opacity-40"
            >
              ⬇ PDF
            </button>
          </div>
        </div>

        <div class="flex items-center justify-between sm:justify-end w-full sm:w-auto gap-1.5 sm:gap-2 text-xs">
          <input
            type="text"
            v-model="tableSearch"
            :placeholder="$t('analysis.searchTable')"
            class="px-2.5 py-1 bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 flex-1 min-w-0 sm:flex-initial sm:w-44"
          />

          <!-- Mobile Page Size Selector (< 768px) -->
          <select
            v-model="filters.pageSize"
            @change="fetchData"
            aria-label="Page Size"
            class="md:hidden px-1.5 py-1 bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 rounded-lg text-2xs font-mono text-slate-700 dark:text-slate-300 focus:outline-none focus:border-blue-500 shrink-0"
          >
            <option :value="50">50/pg</option>
            <option :value="100">100/pg</option>
            <option :value="250">250/pg</option>
            <option :value="500">500/pg</option>
          </select>

          <!-- Inline Compact Pagination -->
          <div class="inline-flex items-center space-x-1 font-mono text-2xs shrink-0">
            <button
              @click="prevPage"
              :disabled="currentPage <= 1 || loading"
              class="px-2 py-1 rounded-md bg-slate-100 hover:bg-slate-200 text-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 disabled:opacity-40 transition font-bold"
              :title="$t('analysis.prev')"
            >
              ‹
            </button>
            <span class="px-1 sm:px-1.5 text-slate-600 dark:text-slate-400 tabular-nums">
              {{ currentPage }} / {{ Math.ceil(totalRecords / filters.pageSize) || 1 }}
            </span>
            <button
              @click="nextPage"
              :disabled="currentPage * filters.pageSize >= totalRecords || loading"
              class="px-2 py-1 rounded-md bg-slate-100 hover:bg-slate-200 text-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 disabled:opacity-40 transition font-bold"
              :title="$t('analysis.next')"
            >
              ›
            </button>
          </div>
        </div>
      </div>

      <!-- Dense Scrollable Table Body with Sticky Header -->
      <div
        :class="workspaceLayout === 'TABLE' ? 'max-h-[540px]' : 'max-h-[290px]'"
        class="overflow-auto compact-analysis-table-wrap"
      >
        <!-- MODE 1: HISTORICAL / PROCESSED TABLE -->
        <table v-if="activeMode === 'HISTORICAL'" class="w-full text-left text-xs font-sans compact-analysis-table">
          <thead class="sticky top-0 z-10 bg-slate-50 dark:bg-[#0B0F19] text-3xs font-bold text-slate-600 dark:text-slate-400 uppercase tracking-wider border-b border-slate-200 dark:border-slate-800">
            <tr>
              <th class="py-2.5 px-4">{{ $t('telemetry.receivedAt') }}</th>
              <th v-if="!isOperator" class="py-2.5 px-4">{{ $t('telemetry.device') }}</th>
              <th class="py-2.5 px-4">{{ $t('telemetry.parameterName') }}</th>
              <th class="py-2.5 px-4 text-right">{{ $t('analysis.processedVal') }}</th>
              <th v-if="canAccessRawMode" class="py-2.5 px-4 text-right">{{ $t('analysis.rawVal') }}</th>
              <th class="py-2.5 px-4">{{ $t('common.unit') }}</th>
              <th class="py-2.5 px-4 text-center">{{ $t('common.quality') }}</th>
              <th class="py-2.5 px-4">{{ $t('deviceDetail.details') || 'Reason / Flags' }}</th>
              <th v-if="!isOperator" class="py-2.5 px-4 text-right">{{ $t('deviceDetail.actions') || 'Actions' }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-800/60 text-2xs">
            <tr v-for="r in paginatedRecords" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition font-mono tabular-nums">
              <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">{{ formatTimestamp(r.received_at || r.timestamp) }}</td>
              <td v-if="!isOperator" class="py-2.5 px-4">
                <router-link :to="'/monitoring/devices/' + r.device_id" class="text-blue-600 dark:text-blue-400 hover:underline font-bold">
                  {{ getDeviceCode(r.device_id) }}
                </router-link>
              </td>
              <td class="py-2.5 px-4 text-slate-800 dark:text-slate-200">
                <span>{{ getParamCode(r.parameter_id) }}</span>
                <span
                  v-if="isSuperAdmin && r.is_injected"
                  :title="'Disuntik Superadmin · Batch ' + (r.injection_batch_id || '')"
                  class="ml-1.5 px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-violet-500/15 text-violet-600 dark:text-violet-300 border border-violet-500/30"
                >INJ</span>
              </td>
              <td class="py-2.5 px-4 text-right font-bold text-slate-900 dark:text-white text-xs">
                {{ formatVal(r.processed_value !== undefined ? r.processed_value : r.value) }}
              </td>
              <td v-if="canAccessRawMode" class="py-2.5 px-4 text-right text-slate-500 dark:text-slate-400">{{ formatVal(r.raw_value) }}</td>
              <td class="py-2.5 px-4 text-slate-500 dark:text-slate-400 font-sans">{{ getParamUnit(r.parameter_id) }}</td>
              <td class="py-2.5 px-4 text-center">
                <span class="px-1.5 py-0.2 rounded text-3xs font-mono font-bold" :class="qualityBadgeClass(r.quality)">
                  {{ r.quality || 'GOOD' }}
                </span>
              </td>
              <td class="py-2.5 px-4 text-3xs text-slate-500 dark:text-slate-400 font-mono">
                <span v-if="r.is_held_value" class="text-amber-600 dark:text-amber-400 mr-1">[HELD]</span>
                <span>{{ r.quality_reason && r.quality_reason !== 'NONE' ? r.quality_reason : '--' }}</span>
              </td>
              <td v-if="!isOperator" class="py-2.5 px-4 text-right font-sans">
                <router-link
                  :to="'/monitoring/devices/' + r.device_id + '?tab=history'"
                  class="text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 hover:underline font-semibold"
                >
                  Detail →
                </router-link>
              </td>
            </tr>
            <tr v-if="paginatedRecords.length === 0">
              <td :colspan="canAccessRawMode ? (isOperator ? 7 : 9) : (isOperator ? 6 : 8)" class="py-6 text-center text-xs text-slate-500 dark:text-slate-400 font-sans">
                {{ loading ? $t('analysis.loadingData') : $t('analysis.noRecords') }}
              </td>
            </tr>
          </tbody>
        </table>

        <!-- MODE 2: RAW TELEMETRY TABLE -->
        <table v-else-if="activeMode === 'RAW'" class="w-full text-left text-xs font-sans compact-analysis-table">
          <thead class="sticky top-0 z-10 bg-slate-50 dark:bg-[#0B0F19] text-3xs font-bold text-slate-600 dark:text-slate-400 uppercase tracking-wider border-b border-slate-200 dark:border-slate-800">
            <tr>
              <th class="py-2.5 px-4">{{ $t('deviceDetail.receivedTime') }}</th>
              <th v-if="!isOperator" class="py-2.5 px-4">{{ $t('telemetry.device') }}</th>
              <th class="py-2.5 px-4">{{ $t('deviceDetail.paramLabel') }}</th>
              <th v-if="canAccessRawMode" class="py-2.5 px-4">{{ $t('analysis.rawHex') }}</th>
              <th v-if="canAccessRawMode" class="py-2.5 px-4 text-right">{{ $t('analysis.rawVal') }}</th>
              <th v-if="!isOperator" class="py-2.5 px-4 text-right">Sebelum Kalibrasi (x)</th>
              <th class="py-2.5 px-4 text-right">{{ canAccessRawMode ? $t('analysis.scaledVal') : $t('deviceDetail.rawOperatorValue') }}</th>
              <th v-if="!isOperator" class="py-2.5 px-4 text-right">Selisih Kalibrasi (Δ)</th>
              <th class="py-2.5 px-4 text-center">{{ $t('common.quality') }}</th>
              <th class="py-2.5 px-4">{{ $t('deviceDetail.protocol') }}</th>
              <th class="py-2.5 px-4 text-right">UUID</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-800/60 text-2xs font-mono tabular-nums">
            <tr v-for="r in paginatedRecords" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition">
              <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">{{ formatTimestamp(r.received_at || r.timestamp) }}</td>
              <td v-if="!isOperator" class="py-2.5 px-4 text-blue-600 dark:text-blue-400 font-bold">
                <router-link :to="'/monitoring/devices/' + r.device_id" class="hover:underline">
                  {{ getDeviceCode(r.device_id) }}
                </router-link>
              </td>
              <td class="py-2.5 px-4 text-slate-800 dark:text-slate-200">
                <span>{{ getParamCode(r.parameter_id) }}</span>
                <span
                  v-if="!isOperator && isRecordCalibrated(r)"
                  :title="'Kurva Kalibrasi: ' + getParamCalEquation(r.parameter_id)"
                  class="ml-1.5 px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-sky-50 text-sky-700 border border-sky-200 dark:bg-sky-500/15 dark:text-sky-300 dark:border-sky-500/30"
                >CAL</span>
                <span
                  v-if="isSuperAdmin && r.is_injected"
                  :title="'Disuntik Superadmin · Batch ' + (r.injection_batch_id || '')"
                  class="ml-1.5 px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-violet-500/15 text-violet-600 dark:text-violet-300 border border-violet-500/30"
                >INJ</span>
              </td>
              <td v-if="canAccessRawMode" class="py-2.5 px-4 text-emerald-600 dark:text-emerald-400 font-bold tracking-widest">{{ r.raw_hex || '--' }}</td>
              <td v-if="canAccessRawMode" class="py-2.5 px-4 text-right text-slate-600 dark:text-slate-300">{{ formatVal(r.raw_value) }}</td>
              <td v-if="!isOperator" class="py-2.5 px-4 text-right text-slate-600 dark:text-slate-300">{{ formatUncalibratedVal(r) }}</td>
              <td class="py-2.5 px-4 text-right text-slate-900 dark:text-white font-bold">{{ formatVal(r.value) }}</td>
              <td v-if="!isOperator" class="py-2.5 px-4 text-right">
                <span
                  v-if="isRecordCalibrated(r)"
                  class="px-1.5 py-0.5 rounded text-3xs font-mono font-bold bg-sky-50 text-sky-700 border border-sky-200 dark:bg-sky-500/15 dark:text-sky-300 dark:border-sky-500/30"
                  :title="'Rumus: ' + getParamCalEquation(r.parameter_id)"
                >{{ formatCalibrationDelta(r) }}</span>
                <span v-else class="text-slate-400 dark:text-slate-500 text-3xs">{{ formatCalibrationDelta(r) }}</span>
              </td>
              <td class="py-2.5 px-4 text-center">
                <span class="px-1.5 py-0.2 rounded text-3xs font-mono font-bold" :class="qualityBadgeClass(r.quality)">
                  {{ r.quality || 'GOOD' }}
                </span>
              </td>
              <td class="py-2.5 px-4 text-slate-500 dark:text-slate-400">{{ r.source || 'MODBUS_TCP' }}</td>
              <td class="py-2.5 px-4 text-right text-3xs text-slate-400 dark:text-slate-500 truncate max-w-[100px]">{{ r.record_uuid || r.id }}</td>
            </tr>
            <tr v-if="paginatedRecords.length === 0">
              <td :colspan="(isOperator ? 8 : 11) - (canAccessRawMode ? 0 : 2)" class="py-6 text-center text-xs text-slate-500 dark:text-slate-400 font-sans">
                {{ loading ? $t('analysis.loadingData') : $t('analysis.noRecords') }}
              </td>
            </tr>
          </tbody>
        </table>

        <!-- MODE 3: AGGREGATION & ROLLUP TABLE -->
        <table v-else class="w-full text-left text-xs font-sans compact-analysis-table">
          <thead class="sticky top-0 z-10 bg-slate-50 dark:bg-[#0B0F19] text-3xs font-bold text-slate-600 dark:text-slate-400 uppercase tracking-wider border-b border-slate-200 dark:border-slate-800">
            <tr>
              <th class="py-2.5 px-4">{{ $t('analysis.period') }}</th>
              <th class="py-2.5 px-4">{{ $t('analysis.identifier') }}</th>
              <th v-if="!isOperator" class="py-2.5 px-4">{{ $t('telemetry.device') }}</th>
              <th class="py-2.5 px-4">{{ $t('telemetry.parameterName') }}</th>
              <th class="py-2.5 px-4 text-right">{{ $t('common.value') }}</th>
              <th class="py-2.5 px-4">{{ $t('analysis.function') }}</th>
              <th class="py-2.5 px-4">{{ $t('analysis.interval') }}</th>
              <th class="py-2.5 px-4 text-right">Min / Max</th>
              <th class="py-2.5 px-4 text-center">{{ $t('analysis.samples') }}</th>
              <th class="py-2.5 px-4 text-center">{{ $t('common.quality') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-800/60 text-2xs font-mono tabular-nums">
            <tr v-for="r in paginatedRecords" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition">
              <td class="py-2.5 px-4 text-slate-600 dark:text-slate-300">
                {{ formatTimestamp(r.period_start) }} → {{ formatTimeOnly(r.period_end) }}
              </td>
              <td class="py-2.5 px-4 text-indigo-600 dark:text-indigo-400 font-bold">{{ r.identifier }}</td>
              <td v-if="!isOperator" class="py-2.5 px-4 text-blue-600 dark:text-blue-400 font-semibold">
                <router-link :to="'/monitoring/devices/' + r.device_id" class="hover:underline">
                  {{ getDeviceCode(r.device_id) }}
                </router-link>
              </td>
              <td class="py-2.5 px-4 text-slate-800 dark:text-slate-200">
                <span>{{ getParamCode(r.parameter_id) }}</span>
                <span
                  v-if="isSuperAdmin && r.has_injected_samples"
                  title="Mengandung sampel suntikan Superadmin (Hanya terlihat oleh Superadmin)"
                  class="ml-1.5 px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-violet-500/15 text-violet-600 dark:text-violet-300 border border-violet-500/30"
                >INJ</span>
              </td>
              <td class="py-2.5 px-4 text-right font-bold text-slate-900 dark:text-white text-xs">
                {{ formatVal(r.value) }}
              </td>
              <td class="py-2.5 px-4 text-indigo-600 dark:text-indigo-300 font-bold">{{ r.function }}</td>
              <td class="py-2.5 px-4 text-slate-500 dark:text-slate-400">{{ formatInterval(r.interval_seconds) }}</td>
              <td class="py-2.5 px-4 text-right text-slate-500 dark:text-slate-400 text-3xs">
                {{ formatVal(r.min_value, 1) }} / {{ formatVal(r.max_value, 1) }}
              </td>
              <td class="py-2.5 px-4 text-center text-3xs text-slate-500 dark:text-slate-400">
                <span class="text-slate-900 dark:text-white font-bold">{{ r.sample_count }}</span>
                <span class="text-emerald-600 dark:text-emerald-400 ml-1">({{ r.good_count }}G)</span>
              </td>
              <td class="py-2.5 px-4 text-center">
                <span class="px-1.5 py-0.2 rounded text-3xs font-mono font-bold" :class="qualityBadgeClass(r.quality)">
                  {{ r.quality || 'GOOD' }}
                </span>
              </td>
            </tr>
            <tr v-if="paginatedRecords.length === 0">
              <td :colspan="isOperator ? 9 : 10" class="py-6 text-center text-xs text-slate-500 dark:text-slate-400 font-sans">
                {{ loading ? $t('analysis.loadingData') : $t('analysis.noRecords') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- SUPERADMIN PURGE ALL RAW TELEMETRY MODAL -->
    <div
      v-if="showPurgeRawModal && isSuperAdmin"
      class="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4"
    >
      <div class="bg-white dark:bg-[#111827] border border-rose-300 dark:border-rose-900/80 rounded-2xl w-full max-w-md overflow-hidden shadow-2xl space-y-4 p-6">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <div class="flex items-center space-x-3">
            <div class="w-9 h-9 rounded-full bg-rose-100 dark:bg-rose-950/90 border border-rose-300 dark:border-rose-800 flex items-center justify-center flex-shrink-0">
              <svg class="w-5 h-5 text-rose-600 dark:text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
            </div>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
                {{ $t('retention.purgeRawModalTitle') }}
              </h3>
              <span class="text-3xs text-rose-600 dark:text-rose-400 font-semibold font-sans block">
                {{ $t('retention.purgeRawModalSubtitle') }}
              </span>
            </div>
          </div>
          <button @click="showPurgeRawModal = false" class="text-slate-400 hover:text-slate-700 dark:hover:text-white">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <p class="text-xs text-slate-600 dark:text-slate-300 font-sans leading-relaxed">
          {{ $t('retention.purgeRawModalDesc') }}
        </p>

        <div class="p-2.5 rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-800/60 flex items-start space-x-2 text-2xs text-emerald-800 dark:text-emerald-300 font-sans">
          <svg class="w-4 h-4 text-emerald-600 dark:text-emerald-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
          </svg>
          <span>{{ $t('retention.purgeRawSafeNote') }}</span>
        </div>

        <div class="space-y-1">
          <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 font-sans">
            {{ $t('retention.purgeRawKeywordLabel') }}
          </label>
          <input
            v-model="purgeConfirmText"
            type="text"
            autocomplete="off"
            placeholder="PURGE / HAPUS"
            class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 text-slate-900 dark:text-white font-mono text-xs uppercase tracking-wider focus:outline-none focus:border-rose-500"
          >
        </div>

        <div class="flex items-start space-x-2 pt-0.5">
          <input
            type="checkbox"
            id="confirmAnalysisPurgeRaw"
            v-model="purgeConfirmCheck"
            class="mt-0.5 rounded border-slate-300 dark:border-slate-700 text-rose-600 focus:ring-rose-500"
          >
          <label for="confirmAnalysisPurgeRaw" class="text-2xs text-slate-600 dark:text-slate-300 cursor-pointer select-none font-sans">
            {{ $t('retention.purgeRawConfirmCheck') }}
          </label>
        </div>

        <div class="flex items-center justify-end space-x-2 border-t border-slate-200 dark:border-slate-800 pt-3">
          <button
            @click="showPurgeRawModal = false"
            :disabled="isPurgingRaw"
            class="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold"
          >
            {{ $t('common.cancel') }}
          </button>
          <button
            @click="executePurgeAllRaw"
            :disabled="!isPurgeRawValid || isPurgingRaw"
            class="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 disabled:opacity-40 text-white text-xs font-semibold shadow-sm transition-colors flex items-center"
          >
            <span v-if="isPurgingRaw" class="animate-spin w-3 h-3 mr-1.5 border-2 border-white border-t-transparent rounded-full"></span>
            {{ isPurgingRaw ? $t('retention.purgingRaw') : $t('retention.purgeRawExecuteBtn') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import { getParamSeriesColor } from '../utils/trendPalette';
import { formatCalibrationEquation } from '../utils/calibrationMath';
import {
  exportAnalysisXLS,
  exportAnalysisPDF,
  exportAnalysisChartImage,
} from '../utils/telemetryAnalysisExport';

const SERIES_PALETTE = [
  '#00E5FF', // Cyan
  '#3B82F6', // Blue
  '#10B981', // Emerald
  '#8B5CF6', // Purple
  '#06B6D4', // Deep Cyan
  '#14B8A6', // Teal
  '#6366F1', // Indigo
  '#2DD4BF', // Mint
  '#A855F7', // Vivid Purple
  '#38BDF8', // Light Sky Blue
];

export default {
  name: 'TelemetryAnalysisView',
  data() {
    return {
      activeMode: 'HISTORICAL', // 'HISTORICAL' | 'RAW' | 'AGGREGATED'
      workspaceLayout: 'SPLIT', // 'SPLIT' | 'CHART' | 'TABLE'
      rawChartMode: 'scaled', // 'scaled' | 'raw'
      loading: false,
      isExporting: false,
      exportDropdownOpen: false,
      devicesList: [],
      records: [],
      totalRecords: 0,
      currentPage: 1,
      anchoredTimeBounds: null,
      tableSearch: '',
      filters: {
        deviceId: '',
        parameterId: '',
        timePreset: '24h',
        startTime: '',
        endTime: '',
        quality: '',
        aggFunction: '',
        pageSize: 50,
      },
      // Chart canvas state
      containerWidth: 800,
      containerHeight: 240,
      activePoint: null,
      hoveredSeriesId: null,
      focusedSeriesId: null,
      hiddenSeriesIds: [],
      chartScaleMode: 'normalized', // 'normalized' | 'native'
      resizeObserver: null,
      // Superadmin Purge All Raw Telemetry state
      showPurgeRawModal: false,
      purgeConfirmText: '',
      purgeConfirmCheck: false,
      isPurgingRaw: false,
      purgeSuccessMsg: '',
    };
  },
  computed: {
    isDark() {
      return this.$store.state.theme === 'dark' || (typeof document !== 'undefined' && document.documentElement.classList.contains('dark'));
    },
    userRole() {
      const u = this.$store.state.user || {};
      return (u.role || '').toUpperCase();
    },
    isOperator() {
      return (this.userRole || '').toUpperCase() === 'OPERATOR';
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    isPurgeRawValid() {
      const kw = (this.purgeConfirmText || '').trim().toUpperCase();
      return this.purgeConfirmCheck && (kw === 'PURGE' || kw === 'HAPUS');
    },
    canAccessRawMode() {
      if (this.$store && this.$store.getters && this.$store.getters.canAccessRawTrendMode !== undefined) {
        return !!this.$store.getters.canAccessRawTrendMode;
      }
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR' || this.userRole === 'ADMIN';
    },
    rawMaxRangeHours() {
      if (this.$store && this.$store.getters && this.$store.getters.rawTelemetryMaxRangeHours) {
        return Number(this.$store.getters.rawTelemetryMaxRangeHours) || 168;
      }
      const st = (this.$store && this.$store.state && this.$store.state.systemStatus) || {};
      return Number(st.raw_telemetry_max_range_hours) > 0 ? Number(st.raw_telemetry_max_range_hours) : 168;
    },
    rawMaxRangeShort() {
      const h = this.rawMaxRangeHours;
      if (h >= 24 && Number.isInteger(h / 24)) return `${h / 24}d`;
      return `${h}h`;
    },
    availableParameters() {
      if (!this.filters.deviceId) {
        // Collect all parameters across all devices
        const all = [];
        this.devicesList.forEach((d) => {
          if (d.parameters) {
            d.parameters.forEach((p) => {
              all.push({
                ...p,
                parameter_name: this.isOperator ? p.parameter_name : `${d.device_code}: ${p.parameter_name}`,
              });
            });
          }
        });
        return all;
      }
      const dev = this.devicesList.find((d) => d.id === Number(this.filters.deviceId));
      return dev && dev.parameters ? dev.parameters : [];
    },
    activeParameter() {
      if (!this.filters.parameterId) return null;
      for (const d of this.devicesList) {
        if (d.parameters) {
          const match = d.parameters.find((p) => p.id === Number(this.filters.parameterId));
          if (match) return match;
        }
      }
      return null;
    },
    activeParameterLabel() {
      if (this.activeParameter) {
        return `${this.activeParameter.parameter_code} (${this.activeParameter.parameter_name})`;
      }
      return this.$t('analysis.allParameters');
    },
    activeUnit() {
      if (this.activeParameter && this.activeParameter.unit) {
        return this.activeParameter.unit;
      }
      return '';
    },
    seriesList() {
      if (!this.records || this.records.length === 0) return [];

      const groups = new Map();

      this.records.forEach((r) => {
        const devId = r.device_id;
        const paramId = r.parameter_id;
        const fn = this.activeMode === 'AGGREGATED' ? (r.function || 'VAL') : '';
        const key = fn ? `${devId}_${paramId}_${fn}` : `${devId}_${paramId}`;

        if (!groups.has(key)) {
          const deviceCode = r.device?.device_code || this.getDeviceCode(devId);
          const deviceName = r.device?.device_name || this.getDeviceName(devId);
          const paramCode = r.parameter?.parameter_code || this.getParamCode(paramId);
          const paramName = r.parameter?.parameter_name || this.getParamName(paramId);
          const unit = r.parameter?.unit !== undefined ? r.parameter.unit : this.getParamUnit(paramId);

          groups.set(key, {
            id: key,
            deviceId: devId,
            parameterId: paramId,
            aggFunction: fn,
            deviceCode,
            deviceName,
            paramCode,
            paramName,
            unit: unit || '',
            records: [],
          });
        }

        groups.get(key).records.push(r);
      });

      const seriesArr = Array.from(groups.values());

      seriesArr.forEach((series, idx) => {
        series.color = getParamSeriesColor(
          {
            param_code: series.paramCode,
            param_name: series.paramName,
            unit: series.unit,
            param_id: series.parameterId,
          },
          this.isDark,
          idx
        );

        // Sort records strictly by timestamp ascending, tie-breaking by id
        series.records.sort((a, b) => {
          const tA = new Date(a.received_at || a.timestamp || a.period_start || 0).getTime();
          const tB = new Date(b.received_at || b.timestamp || b.period_start || 0).getTime();
          if (tA !== tB) return tA - tB;
          return (a.id || 0) - (b.id || 0);
        });

        // Compute per-series statistics
        let sum = 0;
        let min = Infinity;
        let max = -Infinity;
        let validCount = 0;
        let goodCount = 0;

        series.records.forEach((r) => {
          const v = this.extractValue(r);
          if (v !== null && v !== undefined && !isNaN(v) && isFinite(v)) {
            sum += v;
            validCount++;
            if (v < min) min = v;
            if (v > max) max = v;
          }
          if ((r.quality || 'GOOD').toUpperCase() === 'GOOD') {
            goodCount++;
          }
        });

        const lastRec = series.records[series.records.length - 1];
        let latestVal = null;
        let latestTime = null;
        if (lastRec) {
          latestVal = this.extractValue(lastRec);
          latestTime = lastRec.received_at || lastRec.timestamp || lastRec.period_end || lastRec.period_start;
        }

        series.stats = {
          latest: latestVal !== null && latestVal !== undefined && !isNaN(latestVal) ? Number(Number(latestVal).toFixed(2)) : null,
          latestTime,
          min: min !== Infinity && min !== -Infinity && !isNaN(min) ? Number(min.toFixed(2)) : null,
          max: max !== -Infinity && max !== Infinity && !isNaN(max) ? Number(max.toFixed(2)) : null,
          avg: validCount > 0 && !isNaN(sum / validCount) ? Number((sum / validCount).toFixed(2)) : null,
          sampleCount: series.records.length,
          validCount,
          goodCount,
          goodPercent: series.records.length > 0 ? Math.round((goodCount / series.records.length) * 100) : 100,
        };
      });

      return seriesArr;
    },
    visibleSeries() {
      let list = this.seriesList.filter((s) => !this.hiddenSeriesIds.includes(s.id));
      if (this.focusedSeriesId) {
        const focused = list.find((s) => s.id === this.focusedSeriesId);
        if (focused) return [focused];
      }
      return list;
    },
    singleParamStats() {
      if (this.seriesList.length === 0) {
        return {
          latest: null,
          latestTime: null,
          min: null,
          max: null,
          avg: null,
          sampleCount: 0,
          goodCount: 0,
          goodPercent: 100,
        };
      }
      if (this.filters.parameterId) {
        const found = this.seriesList.find((s) => s.parameterId === Number(this.filters.parameterId));
        if (found) return found.stats;
      }
      return this.seriesList[0].stats;
    },
    allSeriesOverview() {
      const totalCount = this.records.length;
      let goodCount = 0;
      this.records.forEach((r) => {
        if ((r.quality || 'GOOD').toUpperCase() === 'GOOD') {
          goodCount++;
        }
      });

      const distinctUnits = Array.from(new Set(this.seriesList.map((s) => s.unit).filter(Boolean)));

      let latestRec = null;
      let latestTime = -Infinity;
      this.records.forEach((r) => {
        const t = new Date(r.received_at || r.timestamp || r.period_end || r.period_start || 0).getTime();
        if (t > latestTime) {
          latestTime = t;
          latestRec = r;
        }
      });

      let latestMeasurement = null;
      if (latestRec) {
        const deviceCode = latestRec.device?.device_code || this.getDeviceCode(latestRec.device_id);
        const deviceName = latestRec.device?.device_name || this.getDeviceName(latestRec.device_id);
        const paramCode = latestRec.parameter?.parameter_code || this.getParamCode(latestRec.parameter_id);
        const paramName = latestRec.parameter?.parameter_name || this.getParamName(latestRec.parameter_id);
        const unit = latestRec.parameter?.unit !== undefined ? latestRec.parameter.unit : this.getParamUnit(latestRec.parameter_id);
        const val = this.extractValue(latestRec);

        latestMeasurement = {
          value: val,
          unit,
          deviceCode,
          deviceName,
          paramCode,
          paramName,
          timestamp: latestRec.received_at || latestRec.timestamp || latestRec.period_end || latestRec.period_start,
          quality: latestRec.quality || 'GOOD',
        };
      }

      return {
        totalCount,
        goodCount,
        goodPercent: totalCount > 0 ? Math.round((goodCount / totalCount) * 100) : 100,
        distinctUnitsCount: distinctUnits.length,
        distinctUnits,
        latestMeasurement,
      };
    },
    availableTimePresets() {
      const maxH = this.rawMaxRangeHours;
      const isRaw = this.activeMode === 'RAW';
      const list = [
        { value: '15m', label: '15m', hours: 0.25 },
        { value: '1h', label: '1h', hours: 1 },
        { value: '6h', label: '6h', hours: 6 },
        { value: '24h', label: '24h', hours: 24 },
        { value: '7d', label: '7d', hours: 168 },
        { value: '30d', label: '30d', hours: 720 },
      ].filter((tp) => !isRaw || tp.hours <= maxH);
      list.push({ value: 'custom', label: 'Custom' });
      return list;
    },
    activeSingleSeriesStats() {
      if (this.focusedSeriesId) {
        const focused = this.seriesList.find((s) => s.id === this.focusedSeriesId);
        if (focused) {
          return {
            label: (!this.isOperator && focused.deviceCode ? `${focused.deviceCode}:` : '') + focused.paramCode,
            unit: focused.unit || '',
            ...focused.stats,
          };
        }
      }
      if (this.filters.parameterId) {
        const found = this.seriesList.find((s) => s.parameterId === Number(this.filters.parameterId));
        if (found) {
          return {
            label: (!this.isOperator && found.deviceCode ? `${found.deviceCode}:` : '') + found.paramCode,
            unit: found.unit || '',
            ...found.stats,
          };
        }
        return {
          label: this.activeParameter ? this.activeParameter.parameter_code : 'Param',
          unit: this.activeUnit,
          ...this.singleParamStats,
        };
      }
      return null;
    },
    exportTimeWindowLabel() {
      if (this.filters.timePreset === 'custom') {
        return `${this.filters.startTime || '--'} s/d ${this.filters.endTime || '--'}`;
      }
      return `Last ${this.filters.timePreset.toUpperCase()}`;
    },
    chartPlotData() {
      const visible = this.visibleSeries;
      const isNarrowPlot = this.containerWidth < 460;
      const padLeft = isNarrowPlot ? 40 : 52;
      const padRight = isNarrowPlot ? 12 : 20;
      const padTop = 18;
      const padBottom = 30;

      if (visible.length === 0) {
        return {
          seriesPlots: [],
          timeTicks: [],
          gridTicks: [],
          isNormalized: false,
          padLeft,
          padRight,
          padTop,
          padBottom,
        };
      }

      const plotW = Math.max(this.containerWidth - padLeft - padRight, 100);
      const plotH = Math.max(this.containerHeight - padTop - padBottom, 80);

      // Determine Time Bounds across all visible series
      let minTime = Infinity;
      let maxTime = -Infinity;

      visible.forEach((s) => {
        s.records.forEach((r) => {
          const t = new Date(r.received_at || r.timestamp || r.period_start || 0).getTime();
          if (!isNaN(t)) {
            if (t < minTime) minTime = t;
            if (t > maxTime) maxTime = t;
          }
        });
      });

      if (minTime === Infinity || maxTime === -Infinity) {
        minTime = Date.now() - 3600000;
        maxTime = Date.now();
      }
      if (minTime === maxTime) {
        minTime -= 60000;
        maxTime += 60000;
      }
      const timeSpan = maxTime - minTime || 1;

      // Determine Scale Mode (Normalized % vs Native absolute values)
      const isNormalized = this.chartScaleMode === 'normalized' && (visible.length > 1 || !this.filters.parameterId);

      let globalMinVal = Infinity;
      let globalMaxVal = -Infinity;
      if (!isNormalized) {
        visible.forEach((s) => {
          s.records.forEach((r) => {
            const v = this.extractValue(r);
            if (v !== null && v !== undefined && !isNaN(v) && isFinite(v)) {
              if (v < globalMinVal) globalMinVal = v;
              if (v > globalMaxVal) globalMaxVal = v;
            }
          });
        });
        if (globalMinVal === Infinity || globalMaxVal === -Infinity) {
          globalMinVal = 0;
          globalMaxVal = 100;
        }
        if (globalMinVal === globalMaxVal) {
          globalMinVal -= 1;
          globalMaxVal += 1;
        }
      }
      const globalValSpan = globalMaxVal - globalMinVal || 1;

      // Map Points for Each Series Independently
      const seriesPlots = visible.map((series) => {
        let sMin = series.stats.min !== null ? series.stats.min : 0;
        let sMax = series.stats.max !== null ? series.stats.max : 100;
        if (sMin === sMax) {
          sMin -= 1;
          sMax += 1;
        }
        const sSpan = sMax - sMin || 1;

        const points = [];
        series.records.forEach((r) => {
          const v = this.extractValue(r);
          if (v === null || v === undefined || isNaN(v) || !isFinite(v)) return;
          const t = new Date(r.received_at || r.timestamp || r.period_start || 0).getTime();
          if (isNaN(t)) return;

          const x = padLeft + ((t - minTime) / timeSpan) * plotW;
          let y;
          if (isNormalized) {
            const normFrac = Math.max(0, Math.min(1, (v - sMin) / sSpan));
            y = padTop + plotH - normFrac * plotH;
          } else {
            const frac = Math.max(0, Math.min(1, (v - globalMinVal) / globalValSpan));
            y = padTop + plotH - frac * plotH;
          }

          points.push({
            x,
            y,
            value: v,
            timestamp: r.received_at || r.timestamp || r.period_start,
            quality: r.quality || 'GOOD',
            qualityReason: r.quality_reason,
            seriesId: series.id,
            deviceCode: series.deviceCode,
            deviceName: series.deviceName,
            paramCode: series.paramCode,
            paramName: series.paramName,
            unit: series.unit,
            rawHex: r.raw_hex,
            rawValue: r.raw_value,
            record: r,
          });
        });

        // Generate line path: never connect adjacent records from different series!
        const linePath = points.map((pt, i) => (i === 0 ? `M ${pt.x.toFixed(1)},${pt.y.toFixed(1)}` : `L ${pt.x.toFixed(1)},${pt.y.toFixed(1)}`)).join(' ');

        let areaPath = '';
        if (points.length > 0 && visible.length === 1) {
          const baseLineY = padTop + plotH;
          const first = points[0];
          const last = points[points.length - 1];
          areaPath = `${linePath} L ${last.x.toFixed(1)},${baseLineY} L ${first.x.toFixed(1)},${baseLineY} Z`;
        }

        return {
          ...series,
          points,
          linePath,
          areaPath,
        };
      });

      // Generate Y-axis grid ticks
      const gridTicks = [];
      const steps = 4;
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const y = padTop + plotH - frac * plotH;
        let label = '';
        if (isNormalized) {
          label = `${Math.round(frac * 100)}%`;
        } else {
          const val = globalMinVal + frac * (globalMaxVal - globalMinVal);
          label = isNaN(val) ? '--' : val.toFixed(1);
        }
        gridTicks.push({ y, label });
      }

      // Generate X-axis time ticks
      const timeTicks = [];
      const timeSteps = Math.min(Math.max(Math.floor(this.containerWidth / 130), 2), 6);
      for (let i = 0; i <= timeSteps; i++) {
        const frac = i / timeSteps;
        const t = new Date(minTime + frac * (maxTime - minTime));
        const x = padLeft + frac * plotW;
        const hours = String(t.getHours()).padStart(2, '0');
        const mins = String(t.getMinutes()).padStart(2, '0');
        const label = `${hours}:${mins}`;
        timeTicks.push({ x, label });
      }

      return {
        seriesPlots,
        timeTicks,
        gridTicks,
        isNormalized,
        padLeft,
        padRight,
        padTop,
        padBottom,
      };
    },
    filteredRecords() {
      if (!this.tableSearch.trim()) return this.records;
      const q = this.tableSearch.toLowerCase().trim();
      return this.records.filter((r) => {
        const devMatch = !this.isOperator && (r.device?.device_code || this.getDeviceCode(r.device_id)).toLowerCase().includes(q);
        const param = (r.parameter?.parameter_code || this.getParamCode(r.parameter_id)).toLowerCase();
        const paramName = (r.parameter?.parameter_name || this.getParamName(r.parameter_id)).toLowerCase();
        const hexMatch = !this.isOperator && (r.raw_hex || '').toLowerCase().includes(q);
        const ident = (r.identifier || '').toLowerCase();
        return devMatch || param.includes(q) || paramName.includes(q) || hexMatch || ident.includes(q);
      });
    },
    paginatedRecords() {
      return this.filteredRecords;
    },
  },
  watch: {
    '$route.query': {
      handler(newQuery) {
        this.applyUrlParams(newQuery);
      },
    },
    isOperator(val) {
      if (val) {
        this.filters.deviceId = '';
        if (this.activeMode === 'RAW') {
          this.setMode('HISTORICAL');
        }
      }
    },
    rawChartMode() {
      // Re-trigger layout/statistics updates
    },
  },
  mounted() {
    this.initResizeObserver();
    if (typeof document !== 'undefined') {
      document.addEventListener('click', this.handleOutsideClick);
    }
    this.fetchDevices().then(() => {
      this.applyUrlParams(this.$route.query);
      this.fetchData();
    });
  },
  beforeDestroy() {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
    }
    if (typeof document !== 'undefined') {
      document.removeEventListener('click', this.handleOutsideClick);
    }
  },
  methods: {
    handleOutsideClick(e) {
      if (!this.exportDropdownOpen) return;
      const wrap = this.$refs.exportMenuWrap;
      if (wrap && !wrap.contains(e.target)) {
        this.exportDropdownOpen = false;
      }
    },
    setWorkspaceLayout(mode) {
      this.workspaceLayout = mode;
      this.$nextTick(() => {
        if (this.$refs.chartContainer) {
          const rect = this.$refs.chartContainer.getBoundingClientRect();
          if (rect.width > 0) this.containerWidth = rect.width;
          if (rect.height > 0) this.containerHeight = rect.height;
        }
      });
    },
    selectTimePreset(val) {
      if (this.filters.timePreset === val && val !== 'custom') return;
      this.filters.timePreset = val;
      this.currentPage = 1;
      this.onTimePresetChange();
    },
    extractValue(r) {
      if (this.activeMode === 'HISTORICAL') {
        return r.processed_value !== undefined && r.processed_value !== null
          ? r.processed_value
          : (r.value !== undefined && r.value !== null ? r.value : null);
      }
      if (this.activeMode === 'RAW') {
        return (this.canAccessRawMode && this.rawChartMode === 'raw')
          ? (r.raw_value !== undefined && r.raw_value !== null ? r.raw_value : null)
          : (r.value !== undefined && r.value !== null ? r.value : null);
      }
      return r.value !== undefined && r.value !== null ? r.value : null;
    },
    toggleSeriesVisibility(id) {
      const idx = this.hiddenSeriesIds.indexOf(id);
      if (idx >= 0) {
        this.hiddenSeriesIds.splice(idx, 1);
      } else {
        // Prevent hiding every series
        if (this.hiddenSeriesIds.length < this.seriesList.length - 1) {
          this.hiddenSeriesIds.push(id);
        }
      }
    },
    toggleFocusSeries(id) {
      if (this.focusedSeriesId === id) {
        this.focusedSeriesId = null;
      } else {
        this.focusedSeriesId = id;
        this.chartScaleMode = 'native'; // Automatically use native scale for single focused series
      }
    },
    resetSeriesFilters() {
      this.focusedSeriesId = null;
      this.hiddenSeriesIds = [];
    },
    setChartScaleMode(mode) {
      this.chartScaleMode = mode;
    },
    initResizeObserver() {
      this.$nextTick(() => {
        if (this.$refs.chartContainer && window.ResizeObserver) {
          this.resizeObserver = new ResizeObserver((entries) => {
            for (const entry of entries) {
              if (entry.contentRect.width > 0) {
                this.containerWidth = entry.contentRect.width;
              }
              if (entry.contentRect.height > 0) {
                this.containerHeight = entry.contentRect.height;
              }
            }
          });
          this.resizeObserver.observe(this.$refs.chartContainer);
        }
      });
    },
    applyUrlParams(query) {
      if (query.mode && ['HISTORICAL', 'RAW', 'AGGREGATED'].includes(query.mode.toUpperCase())) {
        const targetMode = query.mode.toUpperCase();
        this.activeMode = (this.isOperator && targetMode === 'RAW') ? 'HISTORICAL' : targetMode;
        if (!this.canAccessRawMode) {
          this.rawChartMode = 'scaled';
        }
      } else if (this.isOperator && this.activeMode === 'RAW') {
        this.activeMode = 'HISTORICAL';
      }
      if (this.isOperator) {
        this.filters.deviceId = '';
      } else if (query.device_id) {
        this.filters.deviceId = Number(query.device_id);
      }
      if (query.parameter_id) {
        this.filters.parameterId = Number(query.parameter_id);
      }
      if (query.preset) {
        this.filters.timePreset = query.preset;
      }
      if (query.start_time) {
        this.filters.startTime = query.start_time;
        this.filters.timePreset = 'custom';
      }
      if (query.end_time) {
        this.filters.endTime = query.end_time;
        this.filters.timePreset = 'custom';
      }
      if (query.quality) {
        this.filters.quality = query.quality;
      }
    },
    async fetchDevices() {
      try {
        const res = await axios.get('/api/devices?page_size=100');
        if (res.data && res.data.data) {
          this.devicesList = res.data.data.items || res.data.data;
        }
      } catch (err) {
        console.error('Failed to fetch devices list for telemetry analysis:', err);
      }
    },
    setMode(mode) {
      if (this.isOperator && mode === 'RAW') {
        mode = 'HISTORICAL';
      }
      if (!this.canAccessRawMode) {
        this.rawChartMode = 'scaled';
      }
      if (this.activeMode === mode) return;
      this.activeMode = mode;
      if (mode === 'RAW') {
        const maxH = this.rawMaxRangeHours;
        const presetHours = { '15m': 0.25, '1h': 1, '6h': 6, '24h': 24, '7d': 168, '30d': 720 };
        if (presetHours[this.filters.timePreset] && presetHours[this.filters.timePreset] > maxH) {
          this.filters.timePreset = maxH >= 24 ? '24h' : '1h';
        }
      }
      this.currentPage = 1;
      this.focusedSeriesId = null;
      this.hiddenSeriesIds = [];
      this.fetchData();
    },
    onDeviceChange() {
      // If selected parameter doesn't belong to newly selected device, reset parameter filter
      if (this.filters.deviceId && this.filters.parameterId) {
        const dev = this.devicesList.find((d) => d.id === Number(this.filters.deviceId));
        if (dev && dev.parameters) {
          const exists = dev.parameters.some((p) => p.id === Number(this.filters.parameterId));
          if (!exists) {
            this.filters.parameterId = '';
          }
        }
      }
      this.focusedSeriesId = null;
      this.hiddenSeriesIds = [];
      this.fetchData();
    },
    onTimePresetChange() {
      if (this.filters.timePreset !== 'custom') {
        this.fetchData();
      }
    },
    calculateTimeBounds() {
      const now = new Date();
      let start = new Date();

      switch (this.filters.timePreset) {
        case '15m':
          start = new Date(now.getTime() - 15 * 60 * 1000);
          break;
        case '1h':
          start = new Date(now.getTime() - 60 * 60 * 1000);
          break;
        case '6h':
          start = new Date(now.getTime() - 6 * 3600 * 1000);
          break;
        case '24h':
          start = new Date(now.getTime() - 24 * 3600 * 1000);
          break;
        case '7d':
          start = new Date(now.getTime() - 7 * 86400 * 1000);
          break;
        case '30d':
          start = new Date(now.getTime() - 30 * 86400 * 1000);
          break;
        case 'custom':
          if (this.filters.startTime) start = new Date(this.filters.startTime);
          if (this.filters.endTime) now.setTime(new Date(this.filters.endTime).getTime());
          break;
      }

      if (this.activeMode === 'RAW') {
        const maxMs = this.rawMaxRangeHours * 3600 * 1000;
        if (now.getTime() - start.getTime() > maxMs) {
          start = new Date(now.getTime() - maxMs);
        }
      }

      return {
        startTime: start.toISOString(),
        endTime: now.toISOString(),
      };
    },
    async fetchData(opts) {
      this.loading = true;
      try {
        const keepWindow = Boolean(opts && opts.keepWindow === true);
        let bounds;
        if (keepWindow && this.filters.timePreset !== 'custom' && this.anchoredTimeBounds) {
          bounds = this.anchoredTimeBounds;
        } else {
          bounds = this.calculateTimeBounds();
          this.anchoredTimeBounds = bounds;
        }
        const { startTime, endTime } = bounds;
        const baseParams = {
          page: this.currentPage,
          page_size: this.filters.pageSize,
          start_time: startTime,
          end_time: endTime,
        };

        if (this.filters.deviceId) {
          baseParams.device_id = this.filters.deviceId;
        }
        if (this.filters.parameterId) {
          baseParams.parameter_id = this.filters.parameterId;
        }
        if (this.filters.quality) {
          baseParams.quality = this.filters.quality;
        }

        let endpoint = '';
        if (this.activeMode === 'HISTORICAL') {
          endpoint = this.filters.deviceId
            ? `/api/devices/${this.filters.deviceId}/telemetry/history`
            : '/api/telemetry/history';
        } else if (this.activeMode === 'RAW') {
          baseParams.raw_window = 1;
          if (this.canAccessRawMode) {
            endpoint = this.filters.deviceId
              ? `/api/devices/${this.filters.deviceId}/telemetry/raw`
              : '/api/telemetry/raw';
          } else {
            endpoint = this.filters.deviceId
              ? `/api/devices/${this.filters.deviceId}/telemetry/history`
              : '/api/telemetry/history';
          }
        } else if (this.activeMode === 'AGGREGATED') {
          endpoint = '/api/aggregations/results';
          if (this.filters.aggFunction) {
            baseParams.function = this.filters.aggFunction;
          }
        }

        const res = await axios.get(endpoint, { params: baseParams });
        if (res.data && res.data.data) {
          this.records = res.data.data.items || [];
          const apiTotal = Number(res.data.data.total || 0);
          if (apiTotal > 0 || this.currentPage <= 1) {
            this.totalRecords = apiTotal || this.records.length;
          } else if (this.records.length > 0) {
            this.totalRecords = Math.max(
              this.totalRecords,
              (this.currentPage - 1) * this.filters.pageSize + this.records.length,
            );
          }
        } else {
          this.records = [];
          this.totalRecords = 0;
        }
      } catch (err) {
        console.error('Error fetching telemetry analysis dataset:', err);
        this.records = [];
        this.totalRecords = 0;
      } finally {
        this.loading = false;
      }
    },
    prevPage() {
      if (this.currentPage > 1) {
        this.currentPage--;
        this.fetchData({ keepWindow: true });
      }
    },
    nextPage() {
      if (this.currentPage * this.filters.pageSize < this.totalRecords) {
        this.currentPage++;
        this.fetchData({ keepWindow: true });
      }
    },
    onChartTouchMove(e) {
      if (!e.touches || e.touches.length === 0) return;
      const touch = e.touches[0];
      this.onChartMouseMove({
        clientX: touch.clientX,
        clientY: touch.clientY,
        isTouch: true,
      });
    },
    onChartMouseMove(e) {
      if (!this.$refs.chartContainer) return;
      const rect = this.$refs.chartContainer.getBoundingClientRect();
      const mouseX = e.clientX - rect.left;
      const mouseY = e.clientY - rect.top;

      const allPoints = [];
      this.chartPlotData.seriesPlots.forEach((sp) => {
        sp.points.forEach((pt) => {
          allPoints.push({
            ...pt,
            color: sp.color,
          });
        });
      });

      if (allPoints.length === 0) {
        this.activePoint = null;
        return;
      }

      let closest = null;
      let minDistance = Infinity;

      for (const pt of allPoints) {
        const dx = pt.x - mouseX;
        const dy = pt.y - mouseY;
        // Weight X distance heavier for natural time scrubbing while distinguishing vertical series
        const yWeight = e.isTouch ? 0.25 : 0.5;
        const dist = Math.sqrt(dx * dx + (dy * yWeight) * (dy * yWeight));
        if (dist < minDistance) {
          minDistance = dist;
          closest = pt;
        }
      }

      const maxThreshold = e.isTouch ? 180 : 90;
      if (minDistance < maxThreshold) {
        this.activePoint = closest;
      } else {
        this.activePoint = null;
      }
    },
    onChartMouseLeave() {
      this.activePoint = null;
      this.hoveredSeriesId = null;
    },
    getDeviceCode(devId) {
      const dev = this.devicesList.find((d) => d.id === devId);
      return dev ? dev.device_code : `DEV-${devId}`;
    },
    getDeviceName(devId) {
      const dev = this.devicesList.find((d) => d.id === devId);
      return dev ? dev.device_name : '';
    },
    getParamCode(paramId) {
      for (const d of this.devicesList) {
        if (d.parameters) {
          const match = d.parameters.find((p) => p.id === paramId);
          if (match) return match.parameter_code;
        }
      }
      return `PARAM-${paramId}`;
    },
    getParamName(paramId) {
      for (const d of this.devicesList) {
        if (d.parameters) {
          const match = d.parameters.find((p) => p.id === paramId);
          if (match) return match.parameter_name;
        }
      }
      return '';
    },
    getParamUnit(paramId) {
      for (const d of this.devicesList) {
        if (d.parameters) {
          const match = d.parameters.find((p) => p.id === paramId);
          if (match) return match.unit || '';
        }
      }
      return '';
    },
    getParamObj(paramId) {
      for (const d of this.devicesList) {
        if (d.parameters) {
          const match = d.parameters.find((p) => p.id === Number(paramId));
          if (match) return match;
        }
      }
      return null;
    },
    getParamCalEquation(paramId) {
      const p = this.getParamObj(paramId);
      if (!p) return 'y = x (Default)';
      return formatCalibrationEquation(
        !!p.calibration_enabled,
        p.calibration_model,
        p.cal_coeff_c2,
        p.cal_coeff_c1,
        p.cal_coeff_c0,
      );
    },
    isRecordCalibrated(r) {
      if (!r) return false;
      if (r.uncalibrated_value !== undefined && r.uncalibrated_value !== null && r.value !== undefined && r.value !== null) {
        if (Math.abs(Number(r.value) - Number(r.uncalibrated_value)) > 1e-7) return true;
      }
      const p = this.getParamObj(r.parameter_id);
      return !!(p && p.calibration_enabled);
    },
    formatUncalibratedVal(r) {
      if (this.isOperator || !r) return '--';
      const p = this.getParamObj(r.parameter_id);
      const prec = p && p.precision !== undefined ? p.precision : 2;
      if (r.uncalibrated_value !== undefined && r.uncalibrated_value !== null && !isNaN(r.uncalibrated_value)) {
        return Number(r.uncalibrated_value).toFixed(prec);
      }
      if ((!p || !p.calibration_enabled) && r.value !== undefined && r.value !== null && !isNaN(r.value)) {
        return Number(r.value).toFixed(prec);
      }
      return '--';
    },
    formatCalibrationDelta(r) {
      if (this.isOperator || !r) return '--';
      const p = this.getParamObj(r.parameter_id);
      const prec = p && p.precision !== undefined ? p.precision : 2;
      if (r.uncalibrated_value !== undefined && r.uncalibrated_value !== null && r.value !== undefined && r.value !== null) {
        const delta = Number(r.value) - Number(r.uncalibrated_value);
        if (isNaN(delta)) return '--';
        if (!this.isRecordCalibrated(r) && Math.abs(delta) < 1e-7) {
          return '0.00 (y = x)';
        }
        return `${delta >= 0 ? '+' : ''}${delta.toFixed(prec)}`;
      }
      if (!p || !p.calibration_enabled) {
        return '0.00 (y = x)';
      }
      return '--';
    },
    formatVal(v, decimals = 2) {
      if (v === null || v === undefined || isNaN(v)) return '--';
      return Number(v).toFixed(decimals);
    },
    formatInterval(seconds) {
      if (!seconds) return '--';
      if (seconds < 60) return `${seconds}s`;
      if (seconds < 3600) return `${Math.round(seconds / 60)}m`;
      return `${Math.round(seconds / 3600)}h`;
    },
    formatTimestamp(iso) {
      if (!iso) return '--';
      const d = new Date(iso);
      return d.toLocaleString([], {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    },
    formatTimeOnly(iso) {
      if (!iso) return '--';
      const d = new Date(iso);
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    formatRelative(iso) {
      if (!iso) return '';
      const diffMs = Date.now() - new Date(iso).getTime();
      const diffSec = Math.floor(diffMs / 1000);
      if (diffSec < 60) return `${diffSec}s ago`;
      if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
      return `${Math.floor(diffSec / 3600)}h ago`;
    },
    qualityBadgeClass(q) {
      switch ((q || '').toUpperCase()) {
        case 'GOOD':
          return 'bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20';
        case 'UNCERTAIN':
          return 'bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-500/10 dark:text-amber-400 dark:border-amber-500/20';
        case 'BAD':
          return 'bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20';
        case 'STALE':
          return 'bg-purple-50 text-purple-700 border border-purple-200 dark:bg-purple-500/10 dark:text-purple-400 dark:border-purple-500/20';
        default:
          return 'bg-slate-100 text-slate-600 border border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700';
      }
    },
    exportCSV() {
      this.exportDropdownOpen = false;
      if (this.records.length === 0) return;
      if (this.activeMode === 'RAW' && !this.canAccessRawMode) return;

      let csv = '';
      if (this.activeMode === 'HISTORICAL') {
        if (this.canAccessRawMode) {
          csv = this.isOperator
            ? 'Timestamp,Parameter,ProcessedValue,RawValue,Quality,QualityReason\n'
            : 'Timestamp,Device,Parameter,ProcessedValue,RawValue,Quality,QualityReason\n';
          this.records.forEach((r) => {
            const devPart = this.isOperator ? '' : `"${this.getDeviceCode(r.device_id)}",`;
            csv += `"${r.received_at || r.timestamp}",${devPart}"${this.getParamCode(r.parameter_id)}",${r.processed_value !== undefined ? r.processed_value : r.value},${r.raw_value},"${r.quality}","${r.quality_reason || ''}"\n`;
          });
        } else {
          csv = this.isOperator
            ? 'Timestamp,Parameter,ProcessedValue,Quality,QualityReason\n'
            : 'Timestamp,Device,Parameter,ProcessedValue,Quality,QualityReason\n';
          this.records.forEach((r) => {
            const devPart = this.isOperator ? '' : `"${this.getDeviceCode(r.device_id)}",`;
            csv += `"${r.received_at || r.timestamp}",${devPart}"${this.getParamCode(r.parameter_id)}",${r.processed_value !== undefined ? r.processed_value : r.value},"${r.quality}","${r.quality_reason || ''}"\n`;
          });
        }
      } else if (this.activeMode === 'RAW' && this.canAccessRawMode) {
        csv = this.isOperator
          ? 'ReceivedAt,Parameter,RawHex,RawValue,ScaledValue,Quality,Source,RecordUUID\n'
          : 'ReceivedAt,Device,Parameter,RawHex,RawValue,UncalibratedValue,ScaledValue,Quality,Source,RecordUUID\n';
        this.records.forEach((r) => {
          const devPart = this.isOperator ? '' : `"${this.getDeviceCode(r.device_id)}",`;
          const uncalPart = this.isOperator ? '' : `${r.uncalibrated_value !== undefined && r.uncalibrated_value !== null ? r.uncalibrated_value : r.value},`;
          csv += `"${r.received_at || r.timestamp}",${devPart}"${this.getParamCode(r.parameter_id)}","${r.raw_hex || ''}",${r.raw_value},${uncalPart}${r.value},"${r.quality}","${r.source || ''}","${r.record_uuid || ''}"\n`;
        });
      } else {
        csv = this.isOperator
          ? 'PeriodStart,PeriodEnd,Identifier,Parameter,Value,Function,IntervalSeconds,SampleCount,Quality\n'
          : 'PeriodStart,PeriodEnd,Identifier,Device,Parameter,Value,Function,IntervalSeconds,SampleCount,Quality\n';
        this.records.forEach((r) => {
          const devPart = this.isOperator ? '' : `"${this.getDeviceCode(r.device_id)}",`;
          csv += `"${r.period_start}","${r.period_end}","${r.identifier}",${devPart}"${this.getParamCode(r.parameter_id)}",${r.value},"${r.function}",${r.interval_seconds},${r.sample_count},"${r.quality}"\n`;
        });
      }

      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `telemetry_${this.activeMode.toLowerCase()}_${Date.now()}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    },
    exportJSON() {
      this.exportDropdownOpen = false;
      if (this.records.length === 0) return;
      if (this.activeMode === 'RAW' && !this.canAccessRawMode) return;

      const exportedData = this.isOperator
        ? this.records.map(({ device_id, device, ...rest }) => rest)
        : this.records;

      const payload = {
        mode: this.activeMode,
        timestamp: new Date().toISOString(),
        total_records: exportedData.length,
        data: exportedData,
      };

      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `telemetry_${this.activeMode.toLowerCase()}_${Date.now()}.json`;
      a.click();
      URL.revokeObjectURL(url);
    },
    buildExportContext() {
      return {
        activeMode: this.activeMode,
        records: this.filteredRecords,
        seriesList: this.seriesList,
        visibleSeries: this.visibleSeries,
        svgEl: this.$refs.chartSvg || null,
        isDark: this.isDark,
        isOperator: this.isOperator,
        canAccessRawMode: this.canAccessRawMode,
        isSuperAdmin: this.isSuperAdmin,
        timeWindowLabel: this.exportTimeWindowLabel,
        getDeviceCode: this.getDeviceCode,
        getParamCode: this.getParamCode,
        getParamUnit: this.getParamUnit,
        formatTimestamp: this.formatTimestamp,
        formatTimeOnly: this.formatTimeOnly,
        formatVal: this.formatVal,
        formatUncalibratedVal: this.formatUncalibratedVal,
        formatCalibrationDelta: this.formatCalibrationDelta,
        formatInterval: this.formatInterval,
      };
    },
    exportXLS() {
      this.exportDropdownOpen = false;
      if (this.filteredRecords.length === 0) return;
      if (this.activeMode === 'RAW' && !this.canAccessRawMode) return;
      this.isExporting = true;
      try {
        exportAnalysisXLS(this.buildExportContext());
      } catch (err) {
        console.error('[TelemetryAnalysis XLS Export Error]:', err);
        alert('Gagal mengekspor Excel: ' + (err.message || 'Unknown error'));
      } finally {
        this.isExporting = false;
      }
    },
    async exportPDF() {
      this.exportDropdownOpen = false;
      if (this.filteredRecords.length === 0) return;
      if (this.activeMode === 'RAW' && !this.canAccessRawMode) return;
      this.isExporting = true;
      try {
        await exportAnalysisPDF(this.buildExportContext());
      } catch (err) {
        console.error('[TelemetryAnalysis PDF Export Error]:', err);
        alert('Gagal mengekspor PDF: ' + (err.message || 'Unknown error'));
      } finally {
        this.isExporting = false;
      }
    },
    async exportChart(format = 'png') {
      this.exportDropdownOpen = false;
      if (this.visibleSeries.length === 0 || !this.$refs.chartSvg) return;
      if (this.activeMode === 'RAW' && !this.canAccessRawMode) return;
      this.isExporting = true;
      try {
        await exportAnalysisChartImage({
          svgEl: this.$refs.chartSvg,
          activeMode: this.activeMode,
          isDark: this.isDark,
          format,
          titleText: `${this.$t('analysis.breadcrumb')} — ${this.activeMode}`,
          subtitleText: `${this.exportTimeWindowLabel} • ${this.visibleSeries.length}/${this.seriesList.length} series`,
          visibleSeries: this.visibleSeries,
          isOperator: this.isOperator,
        });
      } catch (err) {
        console.error('[TelemetryAnalysis Chart Export Error]:', err);
        alert('Gagal mengekspor gambar grafik: ' + (err.message || 'Unknown error'));
      } finally {
        this.isExporting = false;
      }
    },
    getAuthHeaders() {
      const token = (this.$store && this.$store.state && this.$store.state.token) ||
                    localStorage.getItem('datalogger_token') ||
                    localStorage.getItem('token');
      return token ? { Authorization: `Bearer ${token}` } : {};
    },
    openPurgeRawModal() {
      if (!this.isSuperAdmin) return;
      this.purgeConfirmText = '';
      this.purgeConfirmCheck = false;
      this.showPurgeRawModal = true;
    },
    async executePurgeAllRaw() {
      if (!this.isSuperAdmin || !this.isPurgeRawValid) return;
      this.isPurgingRaw = true;
      try {
        const res = await axios.post(
          '/api/retention/raw-telemetry/purge',
          {
            confirm: true,
            confirm_text: this.purgeConfirmText.trim().toUpperCase(),
          },
          {
            headers: this.getAuthHeaders(),
          }
        );
        if (res.data && res.data.success) {
          const log = res.data.data || {};
          const deletedCount = Number(log.deleted_count || 0).toLocaleString();
          const durMs = Number(log.duration_ms || 0);
          this.showPurgeRawModal = false;
          this.purgeSuccessMsg = `${this.$t('retention.purgeRawSuccessBanner')}: ${deletedCount} rows (${durMs}ms)`;
          this.currentPage = 1;
          await this.fetchData();
        } else {
          alert('Purge failed: ' + ((res.data && res.data.error) || 'Unknown error'));
        }
      } catch (err) {
        const errMsg = err.response && err.response.data && err.response.data.error
          ? err.response.data.error
          : (err.message || 'Unknown network error');
        alert(`Purge error: ${errMsg}`);
      } finally {
        this.isPurgingRaw = false;
      }
    },
  },
};
</script>

<style scoped>
.compact-analysis-table th {
  padding-top: 0.42rem !important;
  padding-bottom: 0.42rem !important;
  padding-left: 0.7rem !important;
  padding-right: 0.7rem !important;
}
.compact-analysis-table td {
  padding-top: 0.34rem !important;
  padding-bottom: 0.34rem !important;
  padding-left: 0.7rem !important;
  padding-right: 0.7rem !important;
}
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

@media (max-width: 767px) {
  .ta-bar1-card {
    display: grid !important;
    grid-template-columns: minmax(0, 1fr) auto;
    row-gap: 8px;
    column-gap: 8px;
    align-items: center;
    padding: 10px 11px 8px !important;
    border-bottom-left-radius: 0 !important;
    border-bottom-right-radius: 0 !important;
    border-bottom-width: 0 !important;
    margin-bottom: 0 !important;
  }
  .ta-bar1-left,
  .ta-bar1-right {
    display: contents !important;
  }
  .ta-desktop-layout-switcher {
    display: none !important;
  }
  .ta-title-group {
    grid-column: 1;
    grid-row: 1;
    min-width: 0;
  }
  .ta-actions-cluster {
    grid-column: 2;
    grid-row: 1;
    justify-self: end;
  }
  .ta-mode-tabs {
    grid-column: 1 / -1;
    grid-row: 2;
    width: 100%;
    display: grid !important;
    grid-template-columns: repeat(auto-fit, minmax(0, 1fr));
    gap: 3px;
    padding: 3px !important;
  }
  .ta-mode-tabs > button {
    width: 100%;
    padding: 6px 6px !important;
  }
  .ta-time-presets {
    grid-column: 1 / -1;
    grid-row: 3;
    width: 100%;
    display: grid !important;
    grid-template-columns: repeat(auto-fit, minmax(0, 1fr));
    gap: 2px;
    padding: 3px !important;
  }
  .ta-time-presets > button {
    width: 100%;
    padding: 5px 2px !important;
  }
  .ta-time-presets > button.is-active-preset {
    background-color: #2563EB !important;
    color: #FFFFFF !important;
    border-color: #2563EB !important;
  }
  .ta-raw-badge {
    grid-column: 1 / -1;
    grid-row: 4;
    justify-self: start;
  }
  .ta-bar1-card + .ta-bar2-card {
    margin-top: 0 !important;
    border-top-left-radius: 0 !important;
    border-top-right-radius: 0 !important;
    padding: 8px 11px 10px !important;
  }
  .ta-bar2-row {
    flex-direction: column;
    align-items: stretch !important;
    gap: 8px !important;
  }
  .ta-filter-grid {
    display: grid !important;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 6px !important;
    width: 100%;
  }
  .ta-pagesize-select {
    display: none !important;
  }
  .ta-filter-grid > select {
    width: 100% !important;
    max-width: none !important;
    min-width: 0 !important;
    height: 33px;
  }
  .ta-filter-grid > select:last-child:nth-child(odd) {
    grid-column: 1 / -1;
  }
  .ta-kpi-strip {
    width: 100%;
  }
  .compact-analysis-table th,
  .compact-analysis-table td {
    padding-left: 0.5rem !important;
    padding-right: 0.5rem !important;
  }
}
</style>
