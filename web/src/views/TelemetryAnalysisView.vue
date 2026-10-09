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
    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
      <!-- Latest Reading -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-400">{{ $t('analysis.latestValue') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-white tracking-tight">{{ stats.latest !== null ? stats.latest : '--' }}</span>
          <span class="text-3xs font-mono text-slate-400">{{ activeUnit }}</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono truncate">{{ stats.latestTime ? formatRelative(stats.latestTime) : '--' }}</div>
      </div>

      <!-- Minimum -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-emerald-400">{{ $t('analysis.min') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-emerald-400 tracking-tight">{{ stats.min !== null ? stats.min : '--' }}</span>
          <span class="text-3xs font-mono text-slate-400">{{ activeUnit }}</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono">{{ $t('analysis.min') }} in window</div>
      </div>

      <!-- Maximum -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-rose-400">{{ $t('analysis.max') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-rose-400 tracking-tight">{{ stats.max !== null ? stats.max : '--' }}</span>
          <span class="text-3xs font-mono text-slate-400">{{ activeUnit }}</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono">{{ $t('analysis.max') }} in window</div>
      </div>

      <!-- Average -->
      <div class="saas-card p-3.5 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-blue-400">{{ $t('analysis.avg') }}</div>
        <div class="flex items-baseline space-x-1.5">
          <span class="text-2xl font-mono font-bold text-blue-400 tracking-tight">{{ stats.avg !== null ? stats.avg : '--' }}</span>
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
          <span class="text-2xl font-mono font-bold text-emerald-400 tracking-tight">{{ stats.goodPercent }}%</span>
          <span class="text-3xs text-slate-400">Good</span>
        </div>
        <div class="text-4xs text-slate-500 font-mono">{{ stats.goodCount }} / {{ stats.sampleCount }} valid</div>
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
              {{ activeParameterLabel }} • {{ activeUnit }}
              <span v-if="activeMode === 'RAW'" class="ml-1 text-slate-500">({{ rawChartMode === 'raw' ? 'Unscaled Register' : 'Scaled Engineering' }})</span>
            </p>
          </div>
        </div>

        <div class="flex items-center space-x-2 text-2xs font-mono text-slate-400">
          <span class="flex items-center space-x-1">
            <span class="w-3 h-0.5 bg-blue-500 inline-block"></span>
            <span>Trend Curve</span>
          </span>
          <span class="flex items-center space-x-1 ml-2">
            <span class="w-3 h-0.5 bg-amber-500 border-dashed inline-block"></span>
            <span>Avg Guide</span>
          </span>
        </div>
      </div>

      <!-- SVG Chart Canvas Area -->
      <div
        ref="chartContainer"
        class="w-full h-80 bg-[#0B0F19] rounded-xl border border-slate-800 relative select-none overflow-hidden"
      >
        <!-- Empty State -->
        <div v-if="chartPoints.length === 0" class="w-full h-full flex flex-col items-center justify-center space-y-2 text-center p-6">
          <svg class="w-10 h-10 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
          </svg>
          <p class="text-xs text-slate-400 font-medium">{{ $t('analysis.noRecords') }}</p>
          <p class="text-3xs text-slate-500">{{ $t('analysis.noRecordsSubtitle') }}</p>
        </div>

        <!-- SVG Chart -->
        <svg
          v-else
          class="w-full h-full overflow-visible chart-svg"
          :viewBox="`0 0 ${containerWidth} ${containerHeight}`"
          @mousemove="onChartMouseMove"
          @mouseleave="onChartMouseLeave"
        >
          <defs>
            <linearGradient id="analysisAreaGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" :stop-color="chartColor" stop-opacity="0.25" />
              <stop offset="60%" :stop-color="chartColor" stop-opacity="0.05" />
              <stop offset="100%" :stop-color="chartColor" stop-opacity="0.00" />
            </linearGradient>
            <linearGradient id="analysisStrokeGradient" x1="0" y1="0" x2="1" y2="0">
              <stop offset="0%" :stop-color="chartLightColor" />
              <stop offset="100%" :stop-color="chartColor" />
            </linearGradient>
          </defs>

          <!-- Horizontal Grid Lines & Y-Axis Labels -->
          <g class="grid-lines">
            <g v-for="(tick, idx) in chartGridTicks" :key="'grid-' + idx">
              <line
                :x1="60"
                :y1="tick.y"
                :x2="containerWidth - 30"
                :y2="tick.y"
                stroke="#1E293B"
                stroke-dasharray="4 4"
                stroke-width="1"
              />
              <text
                :x="52"
                :y="tick.y + 4"
                text-anchor="end"
                fill="#94A3B8"
                font-size="11"
              >
                {{ tick.label }}
              </text>
            </g>
          </g>

          <!-- Vertical Ticks & Timestamps on X-Axis -->
          <g class="time-ticks">
            <g v-for="(tick, idx) in chartTimeTicks" :key="'ttick-' + idx">
              <line
                :x1="tick.x"
                :y1="containerHeight - 40"
                :x2="tick.x"
                :y2="containerHeight - 34"
                stroke="#334155"
                stroke-width="1.2"
              />
              <text
                :x="tick.x"
                :y="containerHeight - 20"
                text-anchor="middle"
                fill="#94A3B8"
                font-size="10"
                font-family="monospace"
              >
                {{ tick.label }}
              </text>
            </g>
          </g>

          <!-- Average Reference Line -->
          <g v-if="chartAvgY !== null">
            <line
              :x1="60"
              :y1="chartAvgY"
              :x2="containerWidth - 30"
              :y2="chartAvgY"
              stroke="#F59E0B"
              stroke-dasharray="5 3"
              stroke-width="1.2"
              opacity="0.75"
            />
            <text
              :x="containerWidth - 35"
              :y="chartAvgY - 4"
              text-anchor="end"
              fill="#F59E0B"
              font-size="10"
              font-family="monospace"
            >
              avg: {{ stats.avg }}
            </text>
          </g>

          <!-- Area Fill Under Curve -->
          <path :d="chartAreaPath" fill="url(#analysisAreaGradient)" />

          <!-- Main Trend Line -->
          <path
            :d="chartLinePath"
            fill="none"
            stroke="url(#analysisStrokeGradient)"
            stroke-width="2.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />

          <!-- Scatter Dots for points if reasonably sized -->
          <g v-if="chartPoints.length <= 150">
            <circle
              v-for="(pt, idx) in chartPoints"
              :key="'pt-' + idx"
              :cx="pt.x"
              :cy="pt.y"
              :r="activePointIndex === idx ? 5 : 2.5"
              :fill="activePointIndex === idx ? '#FFFFFF' : chartColor"
              :stroke="chartLightColor"
              stroke-width="1.5"
              class="transition-all duration-75"
            />
          </g>

          <!-- Hover Crosshair & Tooltip Indicator -->
          <g v-if="activePoint">
            <line
              :x1="activePoint.x"
              :y1="25"
              :x2="activePoint.x"
              :y2="containerHeight - 40"
              stroke="#94A3B8"
              stroke-width="1"
              stroke-dasharray="3 3"
            />
            <circle
              :cx="activePoint.x"
              :cy="activePoint.y"
              r="6"
              fill="#FFFFFF"
              :stroke="chartColor"
              stroke-width="2.5"
            />
          </g>
        </svg>

        <!-- Interactive Floating Tooltip -->
        <div
          v-if="activePoint"
          class="absolute z-20 pointer-events-none bg-[#0F172A]/95 border border-slate-700/80 rounded-xl px-3 py-2 text-xs shadow-xl text-slate-200 font-sans backdrop-blur-sm"
          :style="{
            left: `${Math.min(Math.max(activePoint.x - 70, 10), containerWidth - 160)}px`,
            top: `${Math.max(activePoint.y - 75, 10)}px`
          }"
        >
          <div class="text-3xs text-slate-400 font-mono">{{ formatTimestamp(activePoint.timestamp) }}</div>
          <div class="flex items-baseline space-x-1.5 mt-0.5">
            <span class="text-sm font-bold text-white font-mono">{{ activePoint.value.toFixed(2) }}</span>
            <span class="text-3xs text-slate-400 font-mono">{{ activeUnit }}</span>
            <span class="ml-1 px-1 py-0.2 rounded text-4xs font-mono font-bold" :class="qualityBadgeClass(activePoint.quality)">
              {{ activePoint.quality }}
            </span>
          </div>
          <div class="text-4xs text-slate-400 truncate max-w-[140px] mt-0.5">
            {{ activePoint.deviceName }} • {{ activePoint.paramCode }}
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
                {{ r.value !== null ? r.value.toFixed(2) : '--' }}
              </td>
              <td class="py-2.5 px-4 text-indigo-300 font-bold">{{ r.function }}</td>
              <td class="py-2.5 px-4 text-slate-400">{{ formatInterval(r.interval_seconds) }}</td>
              <td class="py-2.5 px-4 text-right text-slate-400 text-3xs">
                {{ r.min_value !== null ? r.min_value.toFixed(1) : '--' }} / {{ r.max_value !== null ? r.max_value.toFixed(1) : '--' }}
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
      activePointIndex: null,
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
    chartColor() {
      if (this.activeMode === 'RAW') return '#10B981'; // emerald
      if (this.activeMode === 'AGGREGATED') return '#6366F1'; // indigo
      return '#3B82F6'; // blue
    },
    chartLightColor() {
      if (this.activeMode === 'RAW') return '#34D399';
      if (this.activeMode === 'AGGREGATED') return '#818CF8';
      return '#60A5FA';
    },
    filteredRecords() {
      if (!this.tableSearch.trim()) return this.records;
      const q = this.tableSearch.toLowerCase().trim();
      return this.records.filter((r) => {
        const dev = this.getDeviceCode(r.device_id).toLowerCase();
        const param = this.getParamCode(r.parameter_id).toLowerCase();
        const hex = (r.raw_hex || '').toLowerCase();
        const ident = (r.identifier || '').toLowerCase();
        return dev.includes(q) || param.includes(q) || hex.includes(q) || ident.includes(q);
      });
    },
    paginatedRecords() {
      return this.filteredRecords;
    },
    stats() {
      if (this.records.length === 0) {
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

      let sum = 0;
      let min = Infinity;
      let max = -Infinity;
      let validCount = 0;
      let goodCount = 0;

      this.records.forEach((r) => {
        let val = null;
        if (this.activeMode === 'HISTORICAL') {
          val = r.processed_value !== undefined ? r.processed_value : r.value;
        } else if (this.activeMode === 'RAW') {
          val = this.rawChartMode === 'raw' ? r.raw_value : r.value;
        } else {
          val = r.value;
        }

        if (val !== null && val !== undefined && !isNaN(val)) {
          sum += val;
          validCount++;
          if (val < min) min = val;
          if (val > max) max = val;
        }

        if (r.quality === 'GOOD') {
          goodCount++;
        }
      });

      const first = this.records[0];
      let latestVal = null;
      let latestTime = null;
      if (first) {
        latestTime = first.received_at || first.timestamp || first.period_end;
        if (this.activeMode === 'HISTORICAL') {
          latestVal = first.processed_value !== undefined ? first.processed_value : first.value;
        } else if (this.activeMode === 'RAW') {
          latestVal = this.rawChartMode === 'raw' ? first.raw_value : first.value;
        } else {
          latestVal = first.value;
        }
      }

      return {
        latest: latestVal !== null && latestVal !== undefined ? Number(latestVal.toFixed(2)) : null,
        latestTime,
        min: min !== Infinity ? Number(min.toFixed(2)) : null,
        max: max !== -Infinity ? Number(max.toFixed(2)) : null,
        avg: validCount > 0 ? Number((sum / validCount).toFixed(2)) : null,
        sampleCount: this.records.length,
        goodCount,
        goodPercent: this.records.length > 0 ? Math.round((goodCount / this.records.length) * 100) : 100,
      };
    },
    // Chart Calculations
    chartPoints() {
      if (this.records.length === 0) return [];

      // Sort chronological ascending for line plotting
      const sorted = [...this.records].sort((a, b) => {
        const tA = new Date(a.received_at || a.timestamp || a.period_start || 0).getTime();
        const tB = new Date(b.received_at || b.timestamp || b.period_start || 0).getTime();
        return tA - tB;
      });

      const padLeft = 60;
      const padRight = 30;
      const padTop = 25;
      const padBottom = 40;

      const plotW = Math.max(this.containerWidth - padLeft - padRight, 100);
      const plotH = Math.max(this.containerHeight - padTop - padBottom, 80);

      const times = sorted.map((r) => new Date(r.received_at || r.timestamp || r.period_start || 0).getTime());
      const minTime = Math.min(...times);
      const maxTime = Math.max(...times);
      const timeSpan = maxTime - minTime || 1;

      const vals = sorted.map((r) => {
        if (this.activeMode === 'HISTORICAL') {
          return r.processed_value !== undefined ? r.processed_value : r.value;
        }
        if (this.activeMode === 'RAW') {
          return this.rawChartMode === 'raw' ? r.raw_value : r.value;
        }
        return r.value !== null ? r.value : 0;
      });

      let minVal = Math.min(...vals);
      let maxVal = Math.max(...vals);
      if (minVal === maxVal) {
        minVal -= 1;
        maxVal += 1;
      }
      const valSpan = maxVal - minVal;

      return sorted.map((r, idx) => {
        const t = times[idx];
        const v = vals[idx];
        const x = padLeft + ((t - minTime) / timeSpan) * plotW;
        const y = padTop + plotH - ((v - minVal) / valSpan) * plotH;
        return {
          x,
          y,
          value: v,
          timestamp: r.received_at || r.timestamp || r.period_start,
          quality: r.quality || 'GOOD',
          deviceName: this.getDeviceCode(r.device_id),
          paramCode: this.getParamCode(r.parameter_id),
          record: r,
        };
      });
    },
    chartLinePath() {
      if (this.chartPoints.length === 0) return '';
      return this.chartPoints.reduce((acc, pt, idx) => {
        return idx === 0 ? `M ${pt.x},${pt.y}` : `${acc} L ${pt.x},${pt.y}`;
      }, '');
    },
    chartAreaPath() {
      if (this.chartPoints.length === 0) return '';
      const padBottom = 40;
      const baseLineY = this.containerHeight - padBottom;
      const first = this.chartPoints[0];
      const last = this.chartPoints[this.chartPoints.length - 1];
      const line = this.chartLinePath;
      return `${line} L ${last.x},${baseLineY} L ${first.x},${baseLineY} Z`;
    },
    chartAvgY() {
      if (this.stats.avg === null || this.chartPoints.length === 0) return null;
      const padTop = 25;
      const padBottom = 40;
      const plotH = Math.max(this.containerHeight - padTop - padBottom, 80);
      const vals = this.chartPoints.map((p) => p.value);
      let minVal = Math.min(...vals);
      let maxVal = Math.max(...vals);
      if (minVal === maxVal) {
        minVal -= 1;
        maxVal += 1;
      }
      const span = maxVal - minVal;
      return padTop + plotH - ((this.stats.avg - minVal) / span) * plotH;
    },
    chartGridTicks() {
      if (this.chartPoints.length === 0) return [];
      const padTop = 25;
      const padBottom = 40;
      const plotH = Math.max(this.containerHeight - padTop - padBottom, 80);
      const vals = this.chartPoints.map((p) => p.value);
      let minVal = Math.min(...vals);
      let maxVal = Math.max(...vals);
      if (minVal === maxVal) {
        minVal -= 1;
        maxVal += 1;
      }

      const ticks = [];
      const steps = 4;
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const val = minVal + frac * (maxVal - minVal);
        const y = padTop + plotH - frac * plotH;
        ticks.push({
          y,
          label: val.toFixed(1),
        });
      }
      return ticks;
    },
    chartTimeTicks() {
      if (this.chartPoints.length === 0) return [];
      const padLeft = 60;
      const padRight = 30;
      const plotW = Math.max(this.containerWidth - padLeft - padRight, 100);
      const times = this.chartPoints.map((p) => new Date(p.timestamp).getTime());
      const minTime = Math.min(...times);
      const maxTime = Math.max(...times);

      const ticks = [];
      const steps = Math.min(Math.max(Math.floor(this.containerWidth / 130), 2), 6);
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const t = new Date(minTime + frac * (maxTime - minTime));
        const x = padLeft + frac * plotW;
        const hours = String(t.getHours()).padStart(2, '0');
        const mins = String(t.getMinutes()).padStart(2, '0');
        const label = `${hours}:${mins}`;
        ticks.push({ x, label });
      }
      return ticks;
    },
    activePoint() {
      if (this.activePointIndex === null || !this.chartPoints[this.activePointIndex]) {
        return null;
      }
      return this.chartPoints[this.activePointIndex];
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
      if (!this.$refs.chartContainer || this.chartPoints.length === 0) return;
      const rect = this.$refs.chartContainer.getBoundingClientRect();
      const mouseX = e.clientX - rect.left;

      let nearestIdx = 0;
      let minDistance = Infinity;

      for (let i = 0; i < this.chartPoints.length; i++) {
        const d = Math.abs(this.chartPoints[i].x - mouseX);
        if (d < minDistance) {
          minDistance = d;
          nearestIdx = i;
        }
      }

      this.activePointIndex = nearestIdx;
    },
    onChartMouseLeave() {
      this.activePointIndex = null;
    },
    getDeviceCode(devId) {
      const dev = this.devicesList.find((d) => d.id === devId);
      return dev ? dev.device_code : `DEV-${devId}`;
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
    getParamUnit(paramId) {
      for (const d of this.devicesList) {
        if (d.parameters) {
          const match = d.parameters.find((p) => p.id === paramId);
          if (match) return match.unit || '';
        }
      }
      return '';
    },
    formatVal(v) {
      if (v === null || v === undefined || isNaN(v)) return '--';
      return Number(v).toFixed(2);
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
