<template>
  <div class="space-y-6 font-sans">
    <!-- Header & Breadcrumb -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center space-x-2 text-xs text-slate-400 mb-1">
          <router-link to="/monitoring/devices" class="hover:text-blue-400 flex items-center space-x-1">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            <span>{{ $t('navigation.monitoring') }}</span>
          </router-link>
          <span>/</span>
          <span class="text-blue-400 font-semibold">{{ $t('analysis.breadcrumb') }}</span>
        </div>
        <div class="flex items-center space-x-3">
          <h1 class="text-xl font-bold tracking-tight text-white flex items-center space-x-2">
            <span>{{ $t('analysis.title') }}</span>
          </h1>
          <span class="px-2 py-0.5 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/20 text-3xs font-mono font-bold uppercase tracking-wider">
            {{ $t('analysis.semanticNotice') }}
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          {{ $t('analysis.subtitle') }}
        </p>
      </div>

      <!-- Quick Action Buttons -->
      <div class="flex flex-wrap items-center gap-2">
        <button
          @click="fetchData"
          :disabled="loading"
          class="px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition flex items-center space-x-2 disabled:opacity-50"
        >
          <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>{{ $t('common.refresh') }}</span>
        </button>

        <!-- Export Dropdown -->
        <div class="relative">
          <button
            @click="exportDropdownOpen = !exportDropdownOpen"
            class="px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition flex items-center space-x-2"
          >
            <svg class="w-3.5 h-3.5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            <span>{{ $t('common.export') }}</span>
            <svg class="w-3 h-3 text-slate-400 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
            </svg>
          </button>
          <div
            v-if="exportDropdownOpen"
            class="absolute right-0 mt-1 w-36 bg-[#0F172A] border border-slate-700 rounded-xl shadow-xl z-30 py-1"
          >
            <button
              @click="exportCSV"
              class="w-full text-left px-3 py-1.5 text-xs text-slate-200 hover:bg-slate-800 flex items-center space-x-2"
            >
              <span>📄</span>
              <span>{{ $t('analysis.exportCsv') }}</span>
            </button>
            <button
              @click="exportJSON"
              class="w-full text-left px-3 py-1.5 text-xs text-slate-200 hover:bg-slate-800 flex items-center space-x-2"
            >
              <span>📦</span>
              <span>{{ $t('analysis.exportJson') }}</span>
            </button>
          </div>
        </div>

        <!-- Shortcut to Live Telemetry -->
        <router-link
          to="/monitoring/telemetry"
          class="px-3.5 py-2 rounded-xl bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-400 border border-emerald-500/30 text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
          <span>{{ $t('navigation.liveTelemetry') }}</span>
        </router-link>
      </div>
    </div>

    <!-- Data Source / Semantic Mode Selector Bar -->
    <div class="saas-card p-2 bg-[#0F172A]/90 border border-slate-800 rounded-2xl flex flex-col md:flex-row items-stretch md:items-center justify-between gap-2">
      <div class="grid grid-cols-3 gap-1.5 w-full md:w-auto">
        <!-- MODE 1: HISTORICAL / PROCESSED -->
        <button
          @click="setMode('HISTORICAL')"
          class="px-4 py-2.5 rounded-xl text-xs font-semibold transition flex items-center justify-center space-x-2 border"
          :class="activeMode === 'HISTORICAL'
            ? 'bg-blue-600/20 text-blue-400 border-blue-500/40 shadow-sm'
            : 'bg-transparent text-slate-400 border-transparent hover:text-slate-200 hover:bg-slate-800/40'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <div class="text-left">
            <div class="font-bold leading-tight">{{ $t('analysis.modeHistorical') }}</div>
            <div class="text-4xs text-slate-400 hidden sm:block">{{ $t('analysis.modeHistoricalDesc') }}</div>
          </div>
        </button>

        <!-- MODE 2: RAW TELEMETRY -->
        <button
          @click="setMode('RAW')"
          class="px-4 py-2.5 rounded-xl text-xs font-semibold transition flex items-center justify-center space-x-2 border"
          :class="activeMode === 'RAW'
            ? 'bg-emerald-600/20 text-emerald-400 border-emerald-500/40 shadow-sm'
            : 'bg-transparent text-slate-400 border-transparent hover:text-slate-200 hover:bg-slate-800/40'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
          </svg>
          <div class="text-left">
            <div class="font-bold leading-tight">{{ $t('analysis.modeRaw') }}</div>
            <div class="text-4xs text-slate-400 hidden sm:block">{{ $t('analysis.modeRawDesc') }}</div>
          </div>
        </button>

        <!-- MODE 3: AGGREGATION & ROLLUP -->
        <button
          @click="setMode('AGGREGATED')"
          class="px-4 py-2.5 rounded-xl text-xs font-semibold transition flex items-center justify-center space-x-2 border"
          :class="activeMode === 'AGGREGATED'
            ? 'bg-indigo-600/20 text-indigo-400 border-indigo-500/40 shadow-sm'
            : 'bg-transparent text-slate-400 border-transparent hover:text-slate-200 hover:bg-slate-800/40'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
          </svg>
          <div class="text-left">
            <div class="font-bold leading-tight">{{ $t('analysis.modeAggregated') }}</div>
            <div class="text-4xs text-slate-400 hidden sm:block">{{ $t('analysis.modeAggregatedDesc') }}</div>
          </div>
        </button>
      </div>

      <!-- Mode Semantic Badge Description -->
      <div class="px-3 py-1.5 rounded-xl bg-[#0B0F19] border border-slate-800/80 text-3xs font-mono text-slate-300 flex items-center space-x-2">
        <span class="w-2 h-2 rounded-full" :class="activeMode === 'HISTORICAL' ? 'bg-blue-400' : activeMode === 'RAW' ? 'bg-emerald-400' : 'bg-indigo-400'"></span>
        <span>
          {{ activeMode === 'HISTORICAL' ? $t('analysis.historicalSemantic') : activeMode === 'RAW' ? $t('analysis.rawSemantic') : $t('analysis.aggregatedSemantic') }}
        </span>
      </div>
    </div>

    <!-- Unified Filters Bar -->
    <div class="saas-card p-4 bg-[#0F172A]/90 border border-slate-800 rounded-2xl space-y-3">
      <div class="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-3">
        <!-- 1. Device Filter -->
        <div class="space-y-1">
          <label class="text-3xs font-bold text-slate-400 uppercase tracking-wider">{{ $t('telemetry.device') }}</label>
          <select
            v-model="filters.deviceId"
            @change="onDeviceChange"
            class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option value="">{{ $t('analysis.allDevices') }}</option>
            <option v-for="d in devicesList" :key="d.id" :value="d.id">
              {{ d.device_code }} — {{ d.device_name }}
            </option>
          </select>
        </div>

        <!-- 2. Parameter Filter -->
        <div class="space-y-1">
          <label class="text-3xs font-bold text-slate-400 uppercase tracking-wider">{{ $t('telemetry.parameterName') }}</label>
          <select
            v-model="filters.parameterId"
            @change="fetchData"
            class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option value="">{{ $t('analysis.allParameters') }}</option>
            <option v-for="p in availableParameters" :key="p.id" :value="p.id">
              {{ p.parameter_code }} ({{ p.parameter_name }})
            </option>
          </select>
        </div>

        <!-- 3. Time Preset Filter -->
        <div class="space-y-1">
          <label class="text-3xs font-bold text-slate-400 uppercase tracking-wider">{{ $t('analysis.timeRange') }}</label>
          <select
            v-model="filters.timePreset"
            @change="onTimePresetChange"
            class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option value="15m">{{ $t('analysis.last15m') }}</option>
            <option value="1h">{{ $t('analysis.last1h') }}</option>
            <option value="6h">{{ $t('analysis.last6h') }}</option>
            <option value="24h">{{ $t('analysis.last24h') }}</option>
            <option value="7d">{{ $t('analysis.last7d') }}</option>
            <option value="30d">{{ $t('analysis.last30d') }}</option>
            <option value="custom">{{ $t('analysis.custom') }}</option>
          </select>
        </div>

        <!-- 4. Quality Filter -->
        <div class="space-y-1">
          <label class="text-3xs font-bold text-slate-400 uppercase tracking-wider">{{ $t('analysis.qualityFilter') }}</label>
          <select
            v-model="filters.quality"
            @change="fetchData"
            class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option value="">{{ $t('analysis.allQualities') }}</option>
            <option value="GOOD">{{ $t('telemetry.qualityGood') }}</option>
            <option value="UNCERTAIN">{{ $t('telemetry.qualityUncertain') }}</option>
            <option value="BAD">{{ $t('telemetry.qualityBad') }}</option>
            <option value="STALE">{{ $t('telemetry.qualityStale') }}</option>
          </select>
        </div>

        <!-- 5. Mode Specific Filter (Aggregation function or Raw display toggle) -->
        <div v-if="activeMode === 'AGGREGATED'" class="space-y-1">
          <label class="text-3xs font-bold text-slate-400 uppercase tracking-wider">{{ $t('analysis.function') }}</label>
          <select
            v-model="filters.aggFunction"
            @change="fetchData"
            class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option value="">{{ $t('analysis.allFunctions') }}</option>
            <option value="AVG">AVG</option>
            <option value="MIN">MIN</option>
            <option value="MAX">MAX</option>
            <option value="SUM">SUM</option>
            <option value="COUNT">COUNT</option>
            <option value="FIRST">FIRST</option>
            <option value="LAST">LAST</option>
          </select>
        </div>

        <div v-else-if="activeMode === 'RAW'" class="space-y-1">
          <label class="text-3xs font-bold text-slate-400 uppercase tracking-wider">{{ $t('analysis.trendChart') }}</label>
          <select
            v-model="rawChartMode"
            class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option value="scaled">{{ $t('analysis.viewScaledValue') }}</option>
            <option value="raw">{{ $t('analysis.viewRawValue') }}</option>
          </select>
        </div>

        <!-- Page Size -->
        <div class="space-y-1">
          <label class="text-3xs font-bold text-slate-400 uppercase tracking-wider">{{ $t('common.limit') || 'Page Size' }}</label>
          <select
            v-model="filters.pageSize"
            @change="fetchData"
            class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option :value="50">50 records</option>
            <option :value="100">100 records</option>
            <option :value="250">250 records</option>
            <option :value="500">500 records</option>
          </select>
        </div>
      </div>

      <!-- Custom Date Pickers (When custom preset selected) -->
      <div v-if="filters.timePreset === 'custom'" class="pt-2 border-t border-slate-800 flex flex-wrap items-center gap-3 text-xs">
        <div class="flex items-center space-x-2">
          <span class="text-slate-400 text-3xs uppercase font-bold">{{ $t('analysis.startDate') }}:</span>
          <input
            type="datetime-local"
            v-model="filters.startTime"
            class="px-2.5 py-1 bg-[#0B0F19] border border-slate-700 rounded-lg text-xs text-slate-200 font-mono focus:outline-none focus:border-blue-500"
          />
        </div>
        <div class="flex items-center space-x-2">
          <span class="text-slate-400 text-3xs uppercase font-bold">{{ $t('analysis.endDate') }}:</span>
          <input
            type="datetime-local"
            v-model="filters.endTime"
            class="px-2.5 py-1 bg-[#0B0F19] border border-slate-700 rounded-lg text-xs text-slate-200 font-mono focus:outline-none focus:border-blue-500"
          />
        </div>
        <button
          @click="fetchData"
          class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold transition"
        >
          {{ $t('common.apply') || 'Apply' }}
        </button>
      </div>
    </div>

    <!-- Summary KPI Cards -->
    <!-- Case 1: Single Parameter Selected -->
    <div v-if="filters.parameterId" class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
      <!-- Latest Reading -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-400">{{ $t('analysis.latestValue') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-white tracking-tight">{{ singleParamStats.latest !== null ? singleParamStats.latest : '--' }}</span>
          <span class="text-3xs font-mono text-slate-400">{{ activeUnit }}</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono truncate">{{ singleParamStats.latestTime ? formatRelative(singleParamStats.latestTime) : '--' }}</div>
      </div>

      <!-- Minimum -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-emerald-400">{{ $t('analysis.min') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-emerald-400 tracking-tight">{{ singleParamStats.min !== null ? singleParamStats.min : '--' }}</span>
          <span class="text-3xs font-mono text-slate-400">{{ activeUnit }}</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono">{{ $t('analysis.min') }} in window</div>
      </div>

      <!-- Maximum -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-rose-400">{{ $t('analysis.max') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-rose-400 tracking-tight">{{ singleParamStats.max !== null ? singleParamStats.max : '--' }}</span>
          <span class="text-3xs font-mono text-slate-400">{{ activeUnit }}</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono">{{ $t('analysis.max') }} in window</div>
      </div>

      <!-- Average -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-blue-400">{{ $t('analysis.avg') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-blue-400 tracking-tight">{{ singleParamStats.avg !== null ? singleParamStats.avg : '--' }}</span>
          <span class="text-3xs font-mono text-slate-400">{{ activeUnit }}</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono">Mean across samples</div>
      </div>

      <!-- Sample Count -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-400">{{ $t('analysis.count') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-white tracking-tight">{{ totalRecords }}</span>
          <span class="text-3xs text-slate-400">items</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono">{{ totalRecords }} points in set</div>
      </div>

      <!-- Quality Ratio -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-emerald-500/20">
        <div class="text-3xs uppercase tracking-wider font-semibold text-emerald-400">{{ $t('analysis.qualityRatio') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-emerald-400 tracking-tight">{{ singleParamStats.goodPercent }}%</span>
          <span class="text-3xs text-slate-400">Good</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono">{{ singleParamStats.goodCount }} / {{ singleParamStats.sampleCount }} valid</div>
      </div>
    </div>

    <!-- Case 2: All Parameters Selected (Overview Cards) -->
    <div v-else class="space-y-4">
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
        <!-- Active Series Count -->
        <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
          <div class="text-3xs uppercase tracking-wider font-semibold text-blue-400">{{ $t('analysis.activeSeries') }}</div>
          <div class="flex items-baseline space-x-1.5">
            <span class="text-2xl font-mono font-bold text-white tracking-tight">{{ seriesList.length }}</span>
            <span class="text-3xs font-mono text-slate-400">series</span>
          </div>
          <div class="text-4xs text-slate-400 font-mono">{{ allSeriesOverview.distinctUnitsCount }} distinct unit(s)</div>
        </div>

        <!-- Total Dataset Samples -->
        <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
          <div class="text-3xs uppercase tracking-wider font-semibold text-slate-400">{{ $t('analysis.count') }}</div>
          <div class="flex items-baseline space-x-1.5">
            <span class="text-2xl font-mono font-bold text-white tracking-tight">{{ totalRecords }}</span>
            <span class="text-3xs text-slate-400">records</span>
          </div>
          <div class="text-4xs text-slate-500 font-mono">Total points across set</div>
        </div>

        <!-- Overall Quality Health -->
        <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-emerald-500/20">
          <div class="text-3xs uppercase tracking-wider font-semibold text-emerald-400">{{ $t('analysis.qualityRatio') }}</div>
          <div class="flex items-baseline space-x-1.5">
            <span class="text-2xl font-mono font-bold text-emerald-400 tracking-tight">{{ allSeriesOverview.goodPercent }}%</span>
            <span class="text-3xs text-slate-400">Good</span>
          </div>
          <div class="text-4xs text-slate-500 font-mono">{{ allSeriesOverview.goodCount }} / {{ allSeriesOverview.totalCount }} valid</div>
        </div>

        <!-- Most Recent Measurement (With explicit device, param, unit) -->
        <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800 col-span-2">
          <div class="flex items-center justify-between">
            <span class="text-3xs uppercase tracking-wider font-semibold text-purple-400">{{ $t('analysis.mostRecentPacket') }}</span>
            <span v-if="allSeriesOverview.latestMeasurement" class="text-4xs font-mono text-slate-400">
              {{ formatRelative(allSeriesOverview.latestMeasurement.timestamp) }}
            </span>
          </div>
          <div v-if="allSeriesOverview.latestMeasurement" class="flex items-baseline space-x-2">
            <span class="text-xl font-mono font-bold text-white tracking-tight">
              {{ formatVal(allSeriesOverview.latestMeasurement.value) }}
            </span>
            <span class="text-xs font-mono text-purple-300 font-bold">
              {{ allSeriesOverview.latestMeasurement.unit || '--' }}
            </span>
            <span class="text-3xs text-slate-400 truncate font-mono ml-1">
              [{{ allSeriesOverview.latestMeasurement.deviceCode }} : {{ allSeriesOverview.latestMeasurement.paramCode }}]
            </span>
          </div>
          <div v-else class="text-xs text-slate-500 font-mono">--</div>
          <div class="text-4xs text-slate-500 font-mono truncate">
            {{ allSeriesOverview.latestMeasurement ? allSeriesOverview.latestMeasurement.paramName : 'No recent packet' }}
          </div>
        </div>

        <!-- Scale Mode Control Pill -->
        <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
          <div class="text-3xs uppercase tracking-wider font-semibold text-amber-400">{{ $t('analysis.scaleMode') }}</div>
          <div class="flex items-center space-x-1 pt-0.5">
            <button
              @click="setChartScaleMode('normalized')"
              class="px-2 py-1 rounded text-3xs font-semibold font-mono transition"
              :class="chartScaleMode === 'normalized' ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40' : 'bg-slate-800 text-slate-400 hover:text-slate-200'"
              title="Scale 0-100% per series to visually compare parameters of different units without distortion"
            >
              % Norm
            </button>
            <button
              @click="setChartScaleMode('native')"
              class="px-2 py-1 rounded text-3xs font-semibold font-mono transition"
              :class="chartScaleMode === 'native' ? 'bg-blue-500/20 text-blue-300 border border-blue-500/40' : 'bg-slate-800 text-slate-400 hover:text-slate-200'"
              title="Show actual physical engineering values"
            >
              Native
            </button>
          </div>
          <div class="text-4xs text-slate-500 font-mono truncate">
            {{ chartScaleMode === 'normalized' ? '0–100% relative span' : 'Engineering units' }}
          </div>
        </div>
      </div>

      <!-- Per-Series Metric Breakdown Grid -->
      <div v-if="seriesList.length > 0" class="space-y-2">
        <div class="flex items-center justify-between px-1">
          <div>
            <h4 class="text-xs font-bold text-white uppercase tracking-wider flex items-center space-x-2">
              <span>{{ $t('analysis.perSeriesBreakdown') }}</span>
              <span class="px-1.5 py-0.2 rounded bg-slate-800 text-slate-300 text-4xs font-mono">
                {{ seriesList.length }} {{ $t('analysis.activeSeries').toLowerCase() }}
              </span>
            </h4>
            <p class="text-4xs text-slate-400">
              {{ $t('analysis.perSeriesBreakdownSubtitle') }}
            </p>
          </div>
          <button
            v-if="focusedSeriesId || hiddenSeriesIds.length > 0"
            @click="resetSeriesFilters"
            class="text-3xs text-blue-400 hover:text-blue-300 hover:underline font-mono flex items-center space-x-1"
          >
            <span>↺</span>
            <span>{{ $t('analysis.showAllSeries') }}</span>
          </button>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-2.5">
          <div
            v-for="s in seriesList"
            :key="'card-' + s.id"
            class="saas-card p-3 rounded-xl bg-[#0F172A]/90 border transition hover:border-slate-700"
            :class="focusedSeriesId === s.id ? 'border-blue-500 ring-1 ring-blue-500/30' : 'border-slate-800'"
          >
            <!-- Series Header -->
            <div class="flex items-center justify-between pb-2 border-b border-slate-800/80">
              <div class="flex items-center space-x-2 min-w-0">
                <span class="w-2.5 h-2.5 rounded-full shrink-0" :style="{ backgroundColor: s.color }"></span>
                <div class="min-w-0">
                  <div class="text-xs font-bold text-white truncate font-mono">
                    {{ s.deviceCode }}: {{ s.paramCode }}
                  </div>
                  <div class="text-4xs text-slate-400 truncate">
                    {{ s.paramName }}
                  </div>
                </div>
              </div>
              <div class="flex items-center space-x-1.5 shrink-0">
                <span v-if="s.unit" class="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 text-4xs font-mono font-bold">
                  {{ s.unit }}
                </span>
                <button
                  @click="toggleFocusSeries(s.id)"
                  class="px-1.5 py-0.5 rounded text-4xs font-mono transition"
                  :class="focusedSeriesId === s.id ? 'bg-blue-600 text-white' : 'bg-slate-800 hover:bg-slate-700 text-slate-300'"
                  :title="focusedSeriesId === s.id ? 'Click to show all series' : 'Focus this series alone on the chart'"
                >
                  {{ focusedSeriesId === s.id ? $t('analysis.isolatedSeries') : $t('analysis.focusSeries') }}
                </button>
              </div>
            </div>

            <!-- Series Metrics Grid -->
            <div class="grid grid-cols-4 gap-1.5 pt-2 text-center">
              <div class="bg-[#0B0F19] p-1.5 rounded-lg">
                <div class="text-4xs text-slate-500 uppercase font-mono">{{ $t('analysis.latestValue') }}</div>
                <div class="text-xs font-bold font-mono text-white truncate mt-0.5">
                  {{ s.stats.latest !== null ? s.stats.latest : '--' }}
                </div>
              </div>
              <div class="bg-[#0B0F19] p-1.5 rounded-lg">
                <div class="text-4xs text-emerald-400 uppercase font-mono">{{ $t('analysis.min') }}</div>
                <div class="text-xs font-bold font-mono text-emerald-400 truncate mt-0.5">
                  {{ s.stats.min !== null ? s.stats.min : '--' }}
                </div>
              </div>
              <div class="bg-[#0B0F19] p-1.5 rounded-lg">
                <div class="text-4xs text-rose-400 uppercase font-mono">{{ $t('analysis.max') }}</div>
                <div class="text-xs font-bold font-mono text-rose-400 truncate mt-0.5">
                  {{ s.stats.max !== null ? s.stats.max : '--' }}
                </div>
              </div>
              <div class="bg-[#0B0F19] p-1.5 rounded-lg">
                <div class="text-4xs text-blue-400 uppercase font-mono">{{ $t('analysis.avg') }}</div>
                <div class="text-xs font-bold font-mono text-blue-400 truncate mt-0.5">
                  {{ s.stats.avg !== null ? s.stats.avg : '--' }}
                </div>
              </div>
            </div>

            <!-- Card Sub-Footer -->
            <div class="flex items-center justify-between pt-2 mt-1 border-t border-slate-800/60 text-4xs font-mono text-slate-400">
              <span>{{ s.stats.sampleCount }} samples</span>
              <span class="text-emerald-400 font-bold">{{ s.stats.goodPercent }}% Good</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Trend Chart Card -->
    <div class="saas-card p-5 bg-[#0F172A] border border-slate-800 rounded-2xl space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div class="flex items-center space-x-2.5">
          <span class="w-2.5 h-2.5 rounded-full" :class="activeMode === 'HISTORICAL' ? 'bg-blue-500' : activeMode === 'RAW' ? 'bg-emerald-500' : 'bg-indigo-500'"></span>
          <div>
            <h3 class="text-xs font-bold text-white uppercase tracking-wider">{{ $t('analysis.trendChart') }}</h3>
            <p class="text-3xs text-slate-400 mt-0.5">
              <span v-if="filters.parameterId">{{ activeParameterLabel }} • {{ activeUnit }}</span>
              <span v-else>{{ $t('analysis.allParameters') }} ({{ visibleSeries.length }} / {{ seriesList.length }} series active)</span>
              <span v-if="activeMode === 'RAW'" class="ml-1 text-slate-500">({{ rawChartMode === 'raw' ? 'Unscaled Register' : 'Scaled Engineering' }})</span>
            </p>
          </div>
        </div>

        <!-- Scale Mode Switcher & Reset in Chart Header -->
        <div class="flex items-center space-x-2 text-2xs font-mono">
          <!-- Scale switcher buttons -->
          <div class="flex items-center bg-[#0B0F19] rounded-lg p-0.5 border border-slate-800 text-3xs">
            <button
              @click="setChartScaleMode('normalized')"
              class="px-2 py-0.5 rounded transition font-semibold"
              :class="chartScaleMode === 'normalized' ? 'bg-amber-500/20 text-amber-300 border border-amber-500/40' : 'text-slate-400 hover:text-slate-200'"
            >
              {{ $t('analysis.scaleNormalized') }}
            </button>
            <button
              @click="setChartScaleMode('native')"
              class="px-2 py-0.5 rounded transition font-semibold"
              :class="chartScaleMode === 'native' ? 'bg-blue-500/20 text-blue-300 border border-blue-500/40' : 'text-slate-400 hover:text-slate-200'"
            >
              {{ $t('analysis.scaleNative') }}
            </button>
          </div>

          <!-- Reset Filter if series hidden or focused -->
          <button
            v-if="focusedSeriesId || hiddenSeriesIds.length > 0"
            @click="resetSeriesFilters"
            class="px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 text-blue-400 text-3xs font-semibold transition"
          >
            {{ $t('analysis.showAllSeries') }}
          </button>
        </div>
      </div>

      <!-- Interactive Multi-Series Legend Bar -->
      <div v-if="seriesList.length > 0" class="flex flex-wrap items-center gap-2 pt-1 border-t border-slate-800/80">
        <div
          v-for="s in seriesList"
          :key="'legend-' + s.id"
          @mouseenter="hoveredSeriesId = s.id"
          @mouseleave="hoveredSeriesId = null"
          class="flex items-center space-x-1.5 px-2.5 py-1 rounded-lg text-3xs font-mono transition border cursor-pointer select-none"
          :class="[
            hiddenSeriesIds.includes(s.id)
              ? 'opacity-40 bg-slate-900 border-slate-800 line-through text-slate-500'
              : focusedSeriesId === s.id
                ? 'bg-blue-950/40 border-blue-500 text-white shadow-sm'
                : 'bg-[#0B0F19] border-slate-800 text-slate-300 hover:border-slate-700'
          ]"
        >
          <!-- Visibility Toggle Click -->
          <div @click="toggleSeriesVisibility(s.id)" class="flex items-center space-x-1.5">
            <span
              class="w-2.5 h-2.5 rounded-full inline-block transition-transform"
              :style="{ backgroundColor: s.color }"
              :class="{ 'ring-2 ring-white/50 scale-110': hoveredSeriesId === s.id }"
            ></span>
            <span class="font-bold">{{ s.deviceCode }}: {{ s.paramCode }}</span>
            <span v-if="s.unit" class="text-slate-400">({{ s.unit }})</span>
            <span v-if="s.stats.latest !== null" class="text-white font-bold ml-1">{{ s.stats.latest }}</span>
          </div>

          <!-- Quick Solo Focus Button -->
          <button
            @click.stop="toggleFocusSeries(s.id)"
            class="ml-1 text-slate-400 hover:text-blue-400 text-4xs uppercase px-1 py-0.2 rounded hover:bg-slate-800"
            :title="focusedSeriesId === s.id ? 'Exit focus' : 'Focus only this series'"
          >
            {{ focusedSeriesId === s.id ? '✕' : '⊙' }}
          </button>
        </div>
      </div>

      <!-- SVG Chart Canvas Area -->
      <div
        ref="chartContainer"
        class="w-full h-80 bg-[#0B0F19] rounded-xl border border-slate-800 relative select-none overflow-hidden"
      >
        <!-- Empty State -->
        <div v-if="seriesList.length === 0 || visibleSeries.length === 0" class="w-full h-full flex flex-col items-center justify-center space-y-2 text-center p-6">
          <svg class="w-10 h-10 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
          </svg>
          <p class="text-xs text-slate-400 font-medium">{{ $t('analysis.noRecords') }}</p>
          <p class="text-3xs text-slate-500">{{ $t('analysis.noRecordsSubtitle') }}</p>
        </div>

        <!-- SVG Multi-Series Chart -->
        <svg
          v-else
          class="w-full h-full overflow-visible chart-svg"
          :viewBox="`0 0 ${containerWidth} ${containerHeight}`"
          @mousemove="onChartMouseMove"
          @mouseleave="onChartMouseLeave"
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
              <stop offset="0%" :stop-color="s.color" stop-opacity="0.25" />
              <stop offset="60%" :stop-color="s.color" stop-opacity="0.05" />
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
                stroke="#1E293B"
                stroke-dasharray="4 4"
                stroke-width="1"
              />
              <text
                :x="chartPlotData.padLeft - 8"
                :y="tick.y + 4"
                text-anchor="end"
                fill="#94A3B8"
                font-size="11"
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
                :y2="containerHeight - chartPlotData.padBottom + 6"
                stroke="#334155"
                stroke-width="1.2"
              />
              <text
                :x="tick.x"
                :y="containerHeight - 18"
                text-anchor="middle"
                fill="#94A3B8"
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

          <!-- Distinct Trend Lines for Each Series (Never interleaved!) -->
          <g class="series-lines">
            <path
              v-for="s in chartPlotData.seriesPlots"
              :key="'line-' + s.id"
              :d="s.linePath"
              fill="none"
              :stroke="s.color"
              :stroke-width="hoveredSeriesId === s.id || focusedSeriesId === s.id ? 3.2 : 2.2"
              :opacity="focusedSeriesId && focusedSeriesId !== s.id ? 0.2 : hoveredSeriesId && hoveredSeriesId !== s.id ? 0.35 : 1"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="transition-all duration-150"
            />
          </g>

          <!-- Scatter Dots for points if dataset size is reasonable -->
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
          <g v-if="activePoint">
            <line
              :x1="activePoint.x"
              :y1="chartPlotData.padTop"
              :x2="activePoint.x"
              :y2="containerHeight - chartPlotData.padBottom"
              stroke="#94A3B8"
              stroke-width="1"
              stroke-dasharray="3 3"
            />
            <circle
              :cx="activePoint.x"
              :cy="activePoint.y"
              r="6"
              fill="#FFFFFF"
              :stroke="activePoint.color"
              stroke-width="2.5"
            />
          </g>
        </svg>

        <!-- Interactive Floating Tooltip -->
        <div
          v-if="activePoint"
          class="absolute z-20 pointer-events-none bg-[#0F172A]/95 border border-slate-700/80 rounded-xl px-3.5 py-2.5 text-xs shadow-2xl text-slate-200 font-sans backdrop-blur-md"
          :style="{
            left: `${Math.min(Math.max(activePoint.x - 80, 10), containerWidth - 210)}px`,
            top: `${Math.max(activePoint.y - 85, 10)}px`
          }"
        >
          <div class="flex items-center space-x-2 text-3xs text-slate-400 font-mono">
            <span class="w-2 h-2 rounded-full" :style="{ backgroundColor: activePoint.color }"></span>
            <span>{{ formatTimestamp(activePoint.timestamp) }}</span>
          </div>
          <div class="flex items-baseline space-x-1.5 mt-1">
            <span class="text-base font-bold text-white font-mono">{{ formatVal(activePoint.value) }}</span>
            <span class="text-xs font-mono font-bold text-slate-300">{{ activePoint.unit || '' }}</span>
            <span class="ml-1 px-1.5 py-0.2 rounded text-4xs font-mono font-bold" :class="qualityBadgeClass(activePoint.quality)">
              {{ activePoint.quality }}
            </span>
          </div>
          <div class="text-3xs text-slate-300 font-medium truncate max-w-[190px] mt-0.5">
            {{ activePoint.deviceCode }}: {{ activePoint.paramCode }}
          </div>
          <div class="text-4xs text-slate-500 truncate max-w-[190px]">
            {{ activePoint.paramName }} ({{ activePoint.deviceName }})
          </div>
          <div v-if="activePoint.qualityReason && activePoint.qualityReason !== 'NONE'" class="text-4xs text-rose-400 font-mono mt-0.5">
            Reason: {{ activePoint.qualityReason }}
          </div>
          <div v-if="activeMode === 'RAW' && activePoint.rawHex" class="text-4xs text-emerald-400 font-mono mt-0.5">
            Hex: {{ activePoint.rawHex }} (Reg: {{ activePoint.rawValue }})
          </div>
        </div>
      </div>
    </div>

    <!-- Telemetry Data Records Table -->
    <div class="saas-card overflow-hidden border border-slate-800 bg-[#0F172A] rounded-2xl space-y-0">
      <!-- Table Header & Controls -->
      <div class="p-4 border-b border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div class="flex items-center space-x-2">
          <h3 class="text-xs font-bold text-white uppercase tracking-wider">{{ $t('analysis.dataTable') }}</h3>
          <span class="px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 text-3xs font-mono font-bold">
            {{ totalRecords }} items
          </span>
        </div>

        <div class="flex items-center space-x-3 text-xs">
          <!-- Filter Search -->
          <input
            type="text"
            v-model="tableSearch"
            :placeholder="$t('analysis.searchTable')"
            class="px-3 py-1 bg-[#0B0F19] border border-slate-700 rounded-lg text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 w-48"
          />
        </div>
      </div>

      <!-- Table Body based on Active Mode -->
      <div class="overflow-x-auto">
        <!-- MODE 1: HISTORICAL / PROCESSED TABLE -->
        <table v-if="activeMode === 'HISTORICAL'" class="w-full text-left text-xs font-sans">
          <thead class="bg-[#0B0F19] text-3xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-800">
            <tr>
              <th class="py-2.5 px-4">{{ $t('telemetry.receivedAt') }}</th>
              <th class="py-2.5 px-4">{{ $t('telemetry.device') }}</th>
              <th class="py-2.5 px-4">{{ $t('telemetry.parameterName') }}</th>
              <th class="py-2.5 px-4 text-right">{{ $t('analysis.processedVal') }}</th>
              <th class="py-2.5 px-4 text-right">{{ $t('analysis.rawVal') }}</th>
              <th class="py-2.5 px-4">{{ $t('common.unit') }}</th>
              <th class="py-2.5 px-4 text-center">{{ $t('common.quality') }}</th>
              <th class="py-2.5 px-4">{{ $t('deviceDetail.details') || 'Reason / Flags' }}</th>
              <th class="py-2.5 px-4 text-right">{{ $t('deviceDetail.actions') || 'Actions' }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-2xs">
            <tr v-for="r in paginatedRecords" :key="r.id" class="hover:bg-slate-800/30 transition font-mono">
              <td class="py-2.5 px-4 text-slate-300">{{ formatTimestamp(r.received_at || r.timestamp) }}</td>
              <td class="py-2.5 px-4">
                <router-link :to="'/monitoring/devices/' + r.device_id" class="text-blue-400 hover:underline font-bold">
                  {{ getDeviceCode(r.device_id) }}
                </router-link>
              </td>
              <td class="py-2.5 px-4 text-slate-200">{{ getParamCode(r.parameter_id) }}</td>
              <td class="py-2.5 px-4 text-right font-bold text-white text-xs">
                {{ formatVal(r.processed_value !== undefined ? r.processed_value : r.value) }}
              </td>
              <td class="py-2.5 px-4 text-right text-slate-400">{{ formatVal(r.raw_value) }}</td>
              <td class="py-2.5 px-4 text-slate-400 font-sans">{{ getParamUnit(r.parameter_id) }}</td>
              <td class="py-2.5 px-4 text-center">
                <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold" :class="qualityBadgeClass(r.quality)">
                  {{ r.quality || 'GOOD' }}
                </span>
              </td>
              <td class="py-2.5 px-4 text-3xs text-slate-400 font-mono">
                <span v-if="r.is_held_value" class="text-amber-400 mr-1">[HELD]</span>
                <span>{{ r.quality_reason && r.quality_reason !== 'NONE' ? r.quality_reason : '--' }}</span>
              </td>
              <td class="py-2.5 px-4 text-right font-sans">
                <router-link
                  :to="'/monitoring/devices/' + r.device_id + '?tab=history'"
                  class="text-blue-400 hover:text-blue-300 hover:underline"
                >
                  Detail →
                </router-link>
              </td>
            </tr>
            <tr v-if="paginatedRecords.length === 0">
              <td colspan="9" class="py-8 text-center text-xs text-slate-400 font-sans">
                {{ loading ? $t('analysis.loadingData') : $t('analysis.noRecords') }}
              </td>
            </tr>
          </tbody>
        </table>

        <!-- MODE 2: RAW TELEMETRY TABLE -->
        <table v-else-if="activeMode === 'RAW'" class="w-full text-left text-xs font-mono">
          <thead class="bg-[#0B0F19] text-3xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-800">
            <tr>
              <th class="py-2.5 px-4">{{ $t('deviceDetail.receivedTime') }}</th>
              <th class="py-2.5 px-4">{{ $t('telemetry.device') }}</th>
              <th class="py-2.5 px-4">{{ $t('deviceDetail.paramLabel') }}</th>
              <th class="py-2.5 px-4">{{ $t('analysis.rawHex') }}</th>
              <th class="py-2.5 px-4 text-right">{{ $t('analysis.rawVal') }}</th>
              <th class="py-2.5 px-4 text-right">{{ $t('analysis.scaledVal') }}</th>
              <th class="py-2.5 px-4 text-center">{{ $t('common.quality') }}</th>
              <th class="py-2.5 px-4">{{ $t('deviceDetail.protocol') }}</th>
              <th class="py-2.5 px-4 text-right">UUID</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-2xs">
            <tr v-for="r in paginatedRecords" :key="r.id" class="hover:bg-slate-800/30 transition">
              <td class="py-2.5 px-4 text-slate-300">{{ formatTimestamp(r.received_at || r.timestamp) }}</td>
              <td class="py-2.5 px-4 text-blue-400 font-bold">
                <router-link :to="'/monitoring/devices/' + r.device_id" class="hover:underline">
                  {{ getDeviceCode(r.device_id) }}
                </router-link>
              </td>
              <td class="py-2.5 px-4 text-slate-200">{{ getParamCode(r.parameter_id) }}</td>
              <td class="py-2.5 px-4 text-emerald-400 font-bold tracking-widest">{{ r.raw_hex || '--' }}</td>
              <td class="py-2.5 px-4 text-right text-slate-300">{{ formatVal(r.raw_value) }}</td>
              <td class="py-2.5 px-4 text-right text-white font-bold">{{ formatVal(r.value) }}</td>
              <td class="py-2.5 px-4 text-center">
                <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold" :class="qualityBadgeClass(r.quality)">
                  {{ r.quality || 'GOOD' }}
                </span>
              </td>
              <td class="py-2.5 px-4 text-slate-400">{{ r.source || 'MODBUS_TCP' }}</td>
              <td class="py-2.5 px-4 text-right text-3xs text-slate-500 truncate max-w-[100px]">{{ r.record_uuid || r.id }}</td>
            </tr>
            <tr v-if="paginatedRecords.length === 0">
              <td colspan="9" class="py-8 text-center text-xs text-slate-400 font-sans">
                {{ loading ? $t('analysis.loadingData') : $t('analysis.noRecords') }}
              </td>
            </tr>
          </tbody>
        </table>

        <!-- MODE 3: AGGREGATION & ROLLUP TABLE -->
        <table v-else class="w-full text-left text-xs font-sans">
          <thead class="bg-[#0B0F19] text-3xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-800">
            <tr>
              <th class="py-2.5 px-4">{{ $t('analysis.period') }}</th>
              <th class="py-2.5 px-4">{{ $t('analysis.identifier') }}</th>
              <th class="py-2.5 px-4">{{ $t('telemetry.device') }}</th>
              <th class="py-2.5 px-4">{{ $t('telemetry.parameterName') }}</th>
              <th class="py-2.5 px-4 text-right">{{ $t('common.value') }}</th>
              <th class="py-2.5 px-4">{{ $t('analysis.function') }}</th>
              <th class="py-2.5 px-4">{{ $t('analysis.interval') }}</th>
              <th class="py-2.5 px-4 text-right">Min / Max</th>
              <th class="py-2.5 px-4 text-center">{{ $t('analysis.samples') }}</th>
              <th class="py-2.5 px-4 text-center">{{ $t('common.quality') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-2xs font-mono">
            <tr v-for="r in paginatedRecords" :key="r.id" class="hover:bg-slate-800/30 transition">
              <td class="py-2.5 px-4 text-slate-300">
                {{ formatTimestamp(r.period_start) }} → {{ formatTimeOnly(r.period_end) }}
              </td>
              <td class="py-2.5 px-4 text-indigo-400 font-bold">{{ r.identifier }}</td>
              <td class="py-2.5 px-4 text-blue-400">
                <router-link :to="'/monitoring/devices/' + r.device_id" class="hover:underline">
                  {{ getDeviceCode(r.device_id) }}
                </router-link>
              </td>
              <td class="py-2.5 px-4 text-slate-200">{{ getParamCode(r.parameter_id) }}</td>
              <td class="py-2.5 px-4 text-right font-bold text-white text-xs">
                {{ formatVal(r.value) }}
              </td>
              <td class="py-2.5 px-4 text-indigo-300 font-bold">{{ r.function }}</td>
              <td class="py-2.5 px-4 text-slate-400">{{ formatInterval(r.interval_seconds) }}</td>
              <td class="py-2.5 px-4 text-right text-slate-400 text-3xs">
                {{ formatVal(r.min_value, 1) }} / {{ formatVal(r.max_value, 1) }}
              </td>
              <td class="py-2.5 px-4 text-center text-3xs text-slate-400">
                <span class="text-white font-bold">{{ r.sample_count }}</span>
                <span class="text-emerald-400 ml-1">({{ r.good_count }}G)</span>
              </td>
              <td class="py-2.5 px-4 text-center">
                <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold" :class="qualityBadgeClass(r.quality)">
                  {{ r.quality || 'GOOD' }}
                </span>
              </td>
            </tr>
            <tr v-if="paginatedRecords.length === 0">
              <td colspan="10" class="py-8 text-center text-xs text-slate-400 font-sans">
                {{ loading ? $t('analysis.loadingData') : $t('analysis.noRecords') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div class="p-3 border-t border-slate-800 flex items-center justify-between text-xs">
        <button
          @click="prevPage"
          :disabled="currentPage <= 1 || loading"
          class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold"
        >
          {{ $t('analysis.prev') }}
        </button>
        <span class="text-slate-400 text-2xs font-mono">
          {{ $t('analysis.pageOf', { page: currentPage, totalPages: Math.ceil(totalRecords / filters.pageSize) || 1 }) }}
        </span>
        <button
          @click="nextPage"
          :disabled="currentPage * filters.pageSize >= totalRecords || loading"
          class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold"
        >
          {{ $t('analysis.next') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

const SERIES_PALETTE = [
  '#3B82F6', // Blue
  '#10B981', // Emerald
  '#F59E0B', // Amber
  '#EC4899', // Pink
  '#8B5CF6', // Purple
  '#06B6D4', // Cyan
  '#F97316', // Orange
  '#14B8A6', // Teal
  '#6366F1', // Indigo
  '#84CC16', // Lime
];

export default {
  name: 'TelemetryAnalysisView',
  data() {
    return {
      activeMode: 'HISTORICAL', // 'HISTORICAL' | 'RAW' | 'AGGREGATED'
      rawChartMode: 'scaled', // 'scaled' | 'raw'
      loading: false,
      exportDropdownOpen: false,
      devicesList: [],
      records: [],
      totalRecords: 0,
      currentPage: 1,
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
      containerHeight: 320,
      activePoint: null,
      hoveredSeriesId: null,
      focusedSeriesId: null,
      hiddenSeriesIds: [],
      chartScaleMode: 'normalized', // 'normalized' | 'native'
      resizeObserver: null,
    };
  },
  computed: {
    availableParameters() {
      if (!this.filters.deviceId) {
        // Collect all parameters across all devices
        const all = [];
        this.devicesList.forEach((d) => {
          if (d.parameters) {
            d.parameters.forEach((p) => {
              all.push({
                ...p,
                parameter_name: `${d.device_code}: ${p.parameter_name}`,
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
        series.color = SERIES_PALETTE[idx % SERIES_PALETTE.length];

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
    chartPlotData() {
      const visible = this.visibleSeries;
      const padLeft = 65;
      const padRight = 35;
      const padTop = 30;
      const padBottom = 45;

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
        const dev = (r.device?.device_code || this.getDeviceCode(r.device_id)).toLowerCase();
        const param = (r.parameter?.parameter_code || this.getParamCode(r.parameter_id)).toLowerCase();
        const hex = (r.raw_hex || '').toLowerCase();
        const ident = (r.identifier || '').toLowerCase();
        return dev.includes(q) || param.includes(q) || hex.includes(q) || ident.includes(q);
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
    rawChartMode() {
      // Re-trigger layout/statistics updates
    },
  },
  mounted() {
    this.initResizeObserver();
    this.fetchDevices().then(() => {
      this.applyUrlParams(this.$route.query);
      this.fetchData();
    });
  },
  beforeDestroy() {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
    }
  },
  methods: {
    extractValue(r) {
      if (this.activeMode === 'HISTORICAL') {
        return r.processed_value !== undefined && r.processed_value !== null
          ? r.processed_value
          : (r.value !== undefined && r.value !== null ? r.value : null);
      }
      if (this.activeMode === 'RAW') {
        return this.rawChartMode === 'raw'
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
              this.containerWidth = entry.contentRect.width || 800;
              this.containerHeight = entry.contentRect.height || 320;
            }
          });
          this.resizeObserver.observe(this.$refs.chartContainer);
        }
      });
    },
    applyUrlParams(query) {
      if (query.mode && ['HISTORICAL', 'RAW', 'AGGREGATED'].includes(query.mode.toUpperCase())) {
        this.activeMode = query.mode.toUpperCase();
      }
      if (query.device_id) {
        this.filters.deviceId = Number(query.device_id);
      }
      if (query.parameter_id) {
        this.filters.parameterId = Number(query.parameter_id);
      }
      if (query.preset) {
        this.filters.timePreset = query.preset;
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
      if (this.activeMode === mode) return;
      this.activeMode = mode;
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

      return {
        startTime: start.toISOString(),
        endTime: now.toISOString(),
      };
    },
    async fetchData() {
      this.loading = true;
      try {
        const { startTime, endTime } = this.calculateTimeBounds();
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
          endpoint = this.filters.deviceId
            ? `/api/devices/${this.filters.deviceId}/telemetry/raw`
            : '/api/telemetry/raw';
        } else if (this.activeMode === 'AGGREGATED') {
          endpoint = '/api/aggregations/results';
          if (this.filters.aggFunction) {
            baseParams.function = this.filters.aggFunction;
          }
        }

        const res = await axios.get(endpoint, { params: baseParams });
        if (res.data && res.data.data) {
          this.records = res.data.data.items || [];
          this.totalRecords = res.data.data.total || this.records.length;
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
        this.fetchData();
      }
    },
    nextPage() {
      if (this.currentPage * this.filters.pageSize < this.totalRecords) {
        this.currentPage++;
        this.fetchData();
      }
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
        const dist = Math.sqrt(dx * dx + (dy * 0.5) * (dy * 0.5));
        if (dist < minDistance) {
          minDistance = dist;
          closest = pt;
        }
      }

      if (minDistance < 90) {
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
          return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
        case 'UNCERTAIN':
          return 'bg-amber-500/10 text-amber-400 border border-amber-500/20';
        case 'BAD':
          return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
        case 'STALE':
          return 'bg-purple-500/10 text-purple-400 border border-purple-500/20';
        default:
          return 'bg-slate-800 text-slate-400 border border-slate-700';
      }
    },
    exportCSV() {
      this.exportDropdownOpen = false;
      if (this.records.length === 0) return;

      let csv = '';
      if (this.activeMode === 'HISTORICAL') {
        csv = 'Timestamp,Device,Parameter,ProcessedValue,RawValue,Quality,QualityReason\n';
        this.records.forEach((r) => {
          csv += `"${r.received_at || r.timestamp}","${this.getDeviceCode(r.device_id)}","${this.getParamCode(r.parameter_id)}",${r.processed_value !== undefined ? r.processed_value : r.value},${r.raw_value},"${r.quality}","${r.quality_reason || ''}"\n`;
        });
      } else if (this.activeMode === 'RAW') {
        csv = 'ReceivedAt,Device,Parameter,RawHex,RawValue,ScaledValue,Quality,Source,RecordUUID\n';
        this.records.forEach((r) => {
          csv += `"${r.received_at || r.timestamp}","${this.getDeviceCode(r.device_id)}","${this.getParamCode(r.parameter_id)}","${r.raw_hex || ''}",${r.raw_value},${r.value},"${r.quality}","${r.source || ''}","${r.record_uuid || ''}"\n`;
        });
      } else {
        csv = 'PeriodStart,PeriodEnd,Identifier,Device,Parameter,Value,Function,IntervalSeconds,SampleCount,Quality\n';
        this.records.forEach((r) => {
          csv += `"${r.period_start}","${r.period_end}","${r.identifier}","${this.getDeviceCode(r.device_id)}","${this.getParamCode(r.parameter_id)}",${r.value},"${r.function}",${r.interval_seconds},${r.sample_count},"${r.quality}"\n`;
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

      const payload = {
        mode: this.activeMode,
        timestamp: new Date().toISOString(),
        total_records: this.records.length,
        data: this.records,
      };

      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `telemetry_${this.activeMode.toLowerCase()}_${Date.now()}.json`;
      a.click();
      URL.revokeObjectURL(url);
    },
  },
};
</script>
