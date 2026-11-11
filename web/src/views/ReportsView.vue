<template>
  <div class="reports-workbench space-y-2.5 font-sans" :class="{ 'reports-dark': isDarkMode }">
    <!-- Print-Only Official Report Banner -->
    <div v-if="reportData" class="print-only-header">
      <div class="flex items-center justify-between border-b-2 border-slate-800 pb-2 mb-3">
        <div>
          <div class="text-[10px] font-bold uppercase tracking-widest text-blue-700">
            DATALOGGER &bull; OPERATIONAL TELEMETRY REPORT
          </div>
          <h1 class="text-base font-bold text-slate-900">{{ reportData.title || $t('reports.title') }}</h1>
        </div>
        <div class="text-right text-[10px] font-mono text-slate-600">
          <div>{{ reportData.period_formatted }}</div>
          <div>{{ $t('reports.generatedShort') }}: {{ reportData.generated_at_formatted }}</div>
        </div>
      </div>
    </div>

    <!-- TIER 1: UNIFIED COMMAND, SCOPE TABS & INLINE FILTER RIBBON -->
    <div class="no-print reports-card rounded-xl px-3.5 py-2.5 space-y-2">
      <!-- Top Row: Title + Metadata Pills + Complete Export Suite -->
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div class="flex flex-wrap items-center gap-2 min-w-0">
          <div class="w-7 h-7 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-600 flex items-center justify-center flex-shrink-0 reports-title-icon">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
          <h1 class="text-sm font-bold tracking-tight reports-heading whitespace-nowrap">
            {{ $t('reports.title') || 'Operational Reporting' }}
          </h1>

          <template v-if="reportData">
            <span class="reports-meta-pill font-mono" :title="$t('reports.period')">
              {{ reportData.period_formatted }}
            </span>
            <span class="reports-meta-pill font-mono hidden xl:inline-flex" :title="$t('reports.generatedAt')">
              {{ $t('reports.generatedShort') || 'Generated' }} {{ shortGeneratedTime }}
            </span>
            <span class="reports-engine-badge font-mono">
              {{ isCloudMode ? ($t('reports.cloudBadge') || 'CLOUD AGGREGATION') : ($t('reports.serverBadge') || 'SERVER AGGREGATION') }}
            </span>
          </template>
        </div>

        <!-- Compact Export Button Group -->
        <div class="flex flex-wrap items-center gap-1.5">
          <button
            id="btn-print-report"
            type="button"
            @click="printReport"
            :disabled="loading || !reportData"
            :title="$t('reports.printPdf') || 'Print / Save PDF'"
            class="reports-btn-secondary px-2.5 py-1 rounded-lg text-[11px] font-semibold transition flex items-center gap-1 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5 opacity-80" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <polyline points="6 9 6 2 18 2 18 9"></polyline>
              <path d="M6 18H4a2 2 0 01-2-2v-5a2 2 0 012-2h16a2 2 0 012 2v5a2 2 0 01-2 2h-2"></path>
              <rect x="6" y="14" width="12" height="8"></rect>
            </svg>
            <span>{{ $t('reports.printShort') || 'Print' }}</span>
          </button>

          <button
            id="btn-export-pdf"
            type="button"
            @click="exportPdfReport"
            :disabled="loading || !reportData"
            class="reports-btn-secondary px-2.5 py-1 rounded-lg text-[11px] font-semibold transition flex items-center gap-1 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5 text-rose-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
            </svg>
            <span>{{ $t('reports.exportPdf') || 'PDF' }}</span>
          </button>

          <button
            id="btn-export-xls"
            type="button"
            @click="exportXlsReport"
            :disabled="loading || !reportData"
            class="reports-btn-emerald px-2.5 py-1 rounded-lg text-[11px] font-semibold transition flex items-center gap-1 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 10h18M3 14h18m-9-4v8m-7 0h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
            </svg>
            <span>{{ $t('reports.exportXls') || 'Excel .XLS' }}</span>
          </button>

          <button
            id="btn-export-csv"
            type="button"
            @click="exportReport('csv')"
            :disabled="loading || !reportData"
            class="reports-btn-secondary px-2.5 py-1 rounded-lg text-[11px] font-semibold transition flex items-center gap-1 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <span>{{ $t('reports.exportCsv') || 'CSV' }}</span>
          </button>

          <button
            id="btn-download-html"
            type="button"
            @click="exportReport('html')"
            :disabled="loading || !reportData"
            class="px-2.5 py-1 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-[11px] font-semibold shadow-2xs transition flex items-center gap-1 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            <span>{{ $t('reports.downloadHtml') || 'Download HTML' }}</span>
          </button>
        </div>
      </div>

      <!-- Middle Row: Segmented Report Focus Tabs -->
      <div class="reports-tab-bar flex flex-wrap items-center justify-between gap-2 pt-0.5 border-b">
        <div class="flex flex-wrap items-center gap-1">
          <button
            v-for="tab in reportTabs"
            :key="tab.value"
            type="button"
            @click="selectReportTab(tab.value)"
            class="reports-tab-btn px-2.5 py-1.5 text-[11px] font-semibold transition flex items-center gap-1.5 border-b-2 -mb-px"
            :class="filters.report_type === tab.value ? 'reports-tab-active' : 'reports-tab-idle'"
          >
            <span>{{ tab.label }}</span>
            <span v-if="tab.count !== null" class="reports-tab-count font-mono">{{ tab.count }}</span>
          </button>
        </div>

        <span class="text-[10px] reports-muted hidden md:inline">
          {{ $t('reports.subtitle') }}
        </span>
      </div>

      <!-- Bottom Row: Single-Line Inline Filter Strip -->
      <div class="flex flex-wrap items-center gap-1.5 pt-0.5 text-[11px]">
        <!-- Quick Time Presets -->
        <div class="reports-segmented inline-flex items-center p-0.5 rounded-lg font-mono">
          <button
            v-for="preset in presets"
            :key="preset.id"
            type="button"
            @click="applyPreset(preset.id)"
            class="px-2 py-1 rounded-md text-[10.5px] font-semibold transition"
            :class="selectedPreset === preset.id ? 'reports-preset-active' : 'reports-preset-idle'"
          >
            {{ preset.shortLabel }}
          </button>
        </div>

        <!-- Inline Date Range -->
        <div class="flex items-center gap-1">
          <input
            type="datetime-local"
            v-model="filters.startTime"
            @change="onManualDateChange"
            :aria-label="$t('reports.startTime')"
            class="reports-input rounded-lg px-2 py-1 font-mono text-[11px]"
          />
          <span class="reports-muted font-mono">&rarr;</span>
          <input
            type="datetime-local"
            v-model="filters.endTime"
            @change="onManualDateChange"
            :aria-label="$t('reports.endTime')"
            class="reports-input rounded-lg px-2 py-1 font-mono text-[11px]"
          />
        </div>

        <!-- Device Selector -->
        <select
          v-model="filters.deviceId"
          @change="fetchReportPreview"
          :aria-label="$t('reports.device')"
          class="reports-input rounded-lg px-2.5 py-1 text-[11px] min-w-[155px] max-w-[220px]"
        >
          <option value="">{{ $t('reports.allDevices') || 'All Fleet Devices' }} ({{ devices.length }})</option>
          <option v-for="d in devices" :key="d.id" :value="d.id">
            {{ d.name }} ({{ d.code }})
          </option>
        </select>

        <!-- Timezone Selector -->
        <select
          v-model="filters.timezone"
          @change="fetchReportPreview"
          :aria-label="$t('reports.timezone')"
          class="reports-input rounded-lg px-2 py-1 text-[11px]"
        >
          <option value="Asia/Jakarta">WIB (UTC+7)</option>
          <option value="UTC">UTC</option>
        </select>

        <!-- Quick Search Box -->
        <div class="relative flex-1 min-w-[150px]">
          <svg class="w-3.5 h-3.5 reports-muted absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input
            type="text"
            v-model.trim="searchQuery"
            :placeholder="$t('reports.searchPlaceholder') || 'Filter parameter, device, alarm...'"
            class="reports-input w-full rounded-lg pl-7 pr-6 py-1 text-[11px]"
          />
          <button
            v-if="searchQuery"
            type="button"
            @click="searchQuery = ''"
            class="absolute right-2 top-1/2 -translate-y-1/2 reports-muted hover:opacity-80 text-xs font-bold"
          >
            &times;
          </button>
        </div>

        <!-- Compile / Refresh Button -->
        <button
          id="btn-fetch-report"
          type="button"
          @click="fetchReportPreview"
          :disabled="loading"
          class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-[11px] font-semibold transition flex items-center gap-1.5 shadow-2xs disabled:opacity-50 ml-auto"
        >
          <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>{{ loading ? ($t('common.loading') || 'Compiling...') : ($t('reports.compile') || 'Compile') }}</span>
        </button>
      </div>
    </div>

    <!-- Error Banner -->
    <div v-if="error" class="p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-500 text-xs flex items-center justify-between">
      <div class="flex items-center gap-2">
        <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <span>{{ error }}</span>
      </div>
      <button type="button" @click="fetchReportPreview" class="px-2.5 py-1 bg-rose-500/20 hover:bg-rose-500/30 rounded-lg font-semibold text-[11px]">
        {{ $t('common.retry') || 'Retry' }}
      </button>
    </div>

    <!-- Compact Loading Skeleton -->
    <div v-if="loading && !reportData" class="space-y-2.5 animate-pulse">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-2.5">
        <div v-for="i in 4" :key="i" class="h-20 reports-skeleton rounded-xl"></div>
      </div>
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-2.5">
        <div class="lg:col-span-8 h-80 reports-skeleton rounded-xl"></div>
        <div class="lg:col-span-4 space-y-2.5">
          <div class="h-32 reports-skeleton rounded-xl"></div>
          <div class="h-44 reports-skeleton rounded-xl"></div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div
      v-else-if="!loading && (!reportData || (reportData.total_samples === 0 && alarmSummary.total_alarms === 0 && deviceSummaries.length === 0))"
      class="reports-card rounded-xl p-8 text-center space-y-2.5"
    >
      <div class="w-10 h-10 rounded-full bg-amber-500/10 border border-amber-500/20 text-amber-500 flex items-center justify-center mx-auto">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
      </div>
      <h3 class="text-xs font-bold reports-heading">{{ $t('reports.noDataTitle') || 'No Telemetry Records in Period' }}</h3>
      <p class="text-[11px] reports-muted max-w-md mx-auto">
        {{ $t('reports.noDataDesc') }}
      </p>
      <button type="button" @click="applyPreset('24h')" class="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold transition">
        {{ $t('reports.resetLast24h') || 'View Last 24 Hours' }}
      </button>
    </div>

    <!-- ACTIVE COMPACT BENTO REPORT BODY -->
    <div v-else-if="reportData" class="space-y-2.5">
      <!-- TIER 2: COMPACT 4-COLUMN EXECUTIVE KPI STRIP -->
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-2.5">
        <!-- 1. Fleet Availability -->
        <div class="reports-card rounded-xl px-3.5 py-2.5 flex flex-col justify-between">
          <div class="text-[10px] font-bold uppercase tracking-wider reports-muted">
            {{ $t('reports.kpiAvailability') || 'Fleet Availability' }}
          </div>
          <div class="flex items-center gap-2.5 my-1">
            <span class="text-xl font-bold font-mono tabular-nums text-emerald-500 leading-none">
              {{ (reportData.overall_availability || 0).toFixed(1) }}%
            </span>
            <div class="flex-1 h-2 rounded-full reports-track overflow-hidden">
              <div
                class="h-full rounded-full bg-emerald-500 transition-all duration-300"
                :style="{ width: clampPct(reportData.overall_availability) + '%' }"
              ></div>
            </div>
          </div>
          <div class="text-[10.5px] reports-muted font-mono tabular-nums flex items-center justify-between">
            <span>
              <strong class="reports-strong">{{ commSummary.success_frames || 0 }}</strong> OK &bull;
              <span :class="(commSummary.error_frames || 0) > 0 ? 'text-rose-500 font-bold' : ''">
                {{ commSummary.error_frames || 0 }} ERR
              </span>
            </span>
            <span>Avg {{ (commSummary.avg_latency_ms || 0).toFixed(1) }} ms</span>
          </div>
        </div>

        <!-- 2. Good Data Quality -->
        <div class="reports-card rounded-xl px-3.5 py-2.5 flex flex-col justify-between">
          <div class="text-[10px] font-bold uppercase tracking-wider reports-muted">
            {{ $t('reports.kpiDataQuality') || 'Good Data Quality' }}
          </div>
          <div class="flex items-center gap-2.5 my-1">
            <span class="text-xl font-bold font-mono tabular-nums text-blue-500 leading-none">
              {{ (reportData.overall_quality_good || 0).toFixed(1) }}%
            </span>
            <div class="flex-1 h-2 rounded-full reports-track overflow-hidden flex">
              <div class="h-full bg-blue-500" :style="{ width: overallQualityBreakdown.good + '%' }"></div>
              <div v-if="overallQualityBreakdown.uncertain > 0" class="h-full bg-amber-500" :style="{ width: overallQualityBreakdown.uncertain + '%' }"></div>
              <div v-if="overallQualityBreakdown.bad > 0" class="h-full bg-rose-500" :style="{ width: overallQualityBreakdown.bad + '%' }"></div>
            </div>
          </div>
          <div class="text-[10.5px] reports-muted font-mono tabular-nums truncate">
            <strong class="reports-strong">{{ Number(reportData.total_samples || 0).toLocaleString() }}</strong>
            {{ $t('reports.validSamples') || 'Valid Samples' }} &bull; {{ (reportData.overall_quality_good || 0).toFixed(0) }}% Good
          </div>
        </div>

        <!-- 3. Active / Total Alarms -->
        <div class="reports-card rounded-xl px-3.5 py-2.5 flex flex-col justify-between">
          <div class="text-[10px] font-bold uppercase tracking-wider reports-muted">
            {{ $t('reports.kpiAlarms') || 'Active / Total Alarms' }}
          </div>
          <div class="my-0.5">
            <span
              class="text-xl font-bold font-mono tabular-nums leading-none"
              :class="(alarmSummary.active_count || 0) > 0 ? 'text-rose-500' : 'reports-heading'"
            >
              {{ alarmSummary.active_count || 0 }} / {{ alarmSummary.total_alarms || 0 }}
            </span>
          </div>
          <div class="flex flex-wrap items-center gap-1 pt-0.5 font-mono text-[10px]">
            <span class="px-1.5 py-0.5 rounded font-bold bg-rose-500/15 text-rose-500">
              {{ alarmSummary.critical_count || 0 }} CRIT
            </span>
            <span class="px-1.5 py-0.5 rounded font-bold bg-amber-500/15 text-amber-500">
              {{ alarmSummary.warning_count || 0 }} WARN
            </span>
            <span class="px-1.5 py-0.5 rounded font-bold bg-emerald-500/15 text-emerald-500">
              {{ alarmSummary.cleared_count || 0 }} CLEARED
            </span>
          </div>
        </div>

        <!-- 4. Monitored Fleet Scope -->
        <div class="reports-card rounded-xl px-3.5 py-2.5 flex flex-col justify-between">
          <div class="text-[10px] font-bold uppercase tracking-wider reports-muted">
            {{ $t('reports.kpiFleet') || 'Monitored Fleet Scope' }}
          </div>
          <div class="my-0.5">
            <span class="text-xl font-bold font-mono tabular-nums reports-heading leading-none">
              {{ onlineDevicesCount }} / {{ reportData.total_devices || 0 }} {{ $t('reports.onlineNodes') || 'Online' }}
            </span>
          </div>
          <div class="text-[10.5px] reports-muted font-mono tabular-nums truncate">
            <strong class="reports-strong">{{ reportData.total_parameters || 0 }}</strong>
            {{ $t('reports.activeParameters') || 'Active Parameters' }} &bull; {{ dominantSourceLabel }}
          </div>
        </div>
      </div>

      <!-- TIER 3: ASYMMETRIC SPLIT BENTO WORKBENCH (65% LEFT / 35% RIGHT IN COMPREHENSIVE) -->
      <div class="reports-split-grid grid grid-cols-1 lg:grid-cols-12 gap-2.5 items-start">
        <!-- LEFT COLUMN: PARAMETER TELEMETRY & STATISTICAL QUALITY METRICS -->
        <div
          v-if="showTelemetrySection"
          :class="filters.report_type === 'COMPREHENSIVE' ? 'lg:col-span-8' : 'lg:col-span-12'"
          class="reports-card rounded-xl overflow-hidden flex flex-col"
        >
          <div class="px-3.5 py-2.5 border-b reports-divider flex items-center justify-between gap-2">
            <div class="flex items-center gap-2 min-w-0">
              <span class="w-2 h-2 rounded-full bg-emerald-500 flex-shrink-0"></span>
              <h2 class="text-[11px] font-bold uppercase tracking-wider reports-heading truncate">
                {{ $t('reports.telemetryStatisticsTitle') || 'Parameter Telemetry & Statistical Quality Metrics' }}
              </h2>
            </div>
            <span class="text-[10.5px] font-mono reports-muted flex-shrink-0">
              ({{ filteredParameters.length }} parameters)
            </span>
          </div>

          <div
            class="overflow-x-auto"
            :class="filters.report_type === 'COMPREHENSIVE' ? 'reports-scroll-pane max-h-[430px] overflow-y-auto' : ''"
          >
            <table class="w-full text-left border-collapse">
              <thead class="reports-thead sticky top-0 z-10 text-[10px] font-bold uppercase tracking-wider border-b reports-divider">
                <tr>
                  <th class="px-3 py-2">Parameter / Device</th>
                  <th class="px-2.5 py-2">Unit</th>
                  <th class="px-2.5 py-2 text-right">Samples</th>
                  <th class="px-3 py-2 text-center">{{ $t('reports.visualRange') || 'Min — Avg — Max (Visual Range)' }}</th>
                  <th class="px-3 py-2 text-center">{{ $t('reports.firstLastDelta') || 'First → Last (Δ)' }}</th>
                  <th class="px-2.5 py-2 text-center">Quality</th>
                  <th class="px-3 py-2 text-right">Source</th>
                </tr>
              </thead>
              <tbody class="divide-y reports-row-divide text-[11.5px]">
                <tr v-if="filteredParameters.length === 0">
                  <td colspan="7" class="px-4 py-6 text-center text-xs reports-muted">
                    No matching telemetry parameters found.
                  </td>
                </tr>
                <tr
                  v-for="p in filteredParameters"
                  :key="p.parameter_id + '-' + p.device_id"
                  class="reports-row transition-colors"
                >
                  <!-- Parameter & Device -->
                  <td class="px-3 py-2">
                    <div class="font-bold reports-heading leading-tight">{{ p.parameter_name }}</div>
                    <div class="text-[10px] font-mono reports-muted leading-tight">({{ p.device_code }})</div>
                  </td>

                  <!-- Unit -->
                  <td class="px-2.5 py-2 font-mono text-[11px] reports-muted whitespace-nowrap">
                    {{ p.unit || '-' }}
                  </td>

                  <!-- Samples -->
                  <td class="px-2.5 py-2 font-mono tabular-nums text-[11px] text-right reports-strong">
                    {{ Number(p.sample_count || 0).toLocaleString() }}
                  </td>

                  <!-- Min — Avg — Max + Visual Range Bar -->
                  <td class="px-3 py-2 min-w-[175px]">
                    <div class="font-mono tabular-nums text-[11px] text-center whitespace-nowrap">
                      <span class="reports-muted">{{ formatNum(p.min_value) }}</span>
                      <span class="reports-muted mx-1">&mdash;</span>
                      <strong class="text-blue-500 font-bold">{{ formatNum(p.avg_value) }}</strong>
                      <span class="reports-muted mx-1">&mdash;</span>
                      <span class="reports-muted">{{ formatNum(p.max_value) }}</span>
                    </div>
                    <div class="mt-1 h-1.5 rounded-full reports-track relative overflow-hidden">
                      <div
                        class="absolute top-0 bottom-0 rounded-full bg-blue-500/75"
                        :style="computeVisualRangeStyle(p)"
                      ></div>
                    </div>
                  </td>

                  <!-- First -> Last (Delta) -->
                  <td class="px-3 py-2 font-mono tabular-nums text-[11px] text-center whitespace-nowrap">
                    <template v-if="hasFirstLast(p)">
                      <span class="reports-strong">{{ formatNum(p.first_value) }}</span>
                      <span class="reports-muted mx-0.5">&rarr;</span>
                      <span class="reports-strong">{{ formatNum(p.last_value) }}</span>
                      <span class="ml-1 text-[10px] font-semibold" :class="deltaClass(p)">
                        ({{ formatDelta(p) }})
                      </span>
                    </template>
                    <span v-else class="reports-muted">-</span>
                  </td>

                  <!-- Quality -->
                  <td class="px-2.5 py-2 font-mono tabular-nums text-[11px] text-center whitespace-nowrap">
                    <div class="inline-flex items-center gap-1">
                      <span class="text-emerald-500 font-bold">{{ (p.good_percent || 0).toFixed(0) }}% Good</span>
                      <span v-if="(p.bad_percent || 0) > 0" class="text-rose-500 text-[10px]">
                        &bull; {{ p.bad_percent.toFixed(0) }}%B
                      </span>
                      <span v-if="(p.uncertain_percent || 0) > 0" class="text-amber-500 text-[10px]">
                        &bull; {{ p.uncertain_percent.toFixed(0) }}%U
                      </span>
                    </div>
                  </td>

                  <!-- Data Source -->
                  <td class="px-3 py-2 text-right whitespace-nowrap">
                    <span
                      class="inline-block px-1.5 py-0.5 rounded text-[10px] font-mono font-bold uppercase"
                      :class="sourceBadgeClass(p.data_source)"
                    >
                      {{ shortSourceLabel(p.data_source) }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Compact Footer Strip -->
          <div class="px-3.5 py-2 border-t reports-divider reports-footer-strip text-[10.5px] font-mono flex flex-wrap items-center justify-between gap-2">
            <span>
              Total <strong class="reports-strong">{{ Number(reportData.total_samples || 0).toLocaleString() }}</strong> samples across
              <strong class="reports-strong">{{ filteredParameters.length }}</strong> parameters
              &bull; Interval: {{ activePresetShortLabel }}
            </span>
            <span>{{ badUncertainFooterText }}</span>
          </div>
        </div>

        <!-- RIGHT COLUMN: STACKED FLEET COMM AUDIT + ALARMS LOG (OR FULL-WIDTH WHEN TAB SELECTED) -->
        <div
          v-if="showCommSection || showAlarmsSection"
          :class="filters.report_type === 'COMPREHENSIVE' ? 'lg:col-span-4 space-y-2.5' : 'lg:col-span-12 space-y-2.5'"
        >
          <!-- CARD A: DEVICE FLEET AVAILABILITY & COMM AUDIT -->
          <div v-if="showCommSection" class="reports-card rounded-xl overflow-hidden">
            <div class="px-3.5 py-2.5 border-b reports-divider flex items-center justify-between gap-2">
              <div class="flex items-center gap-2 min-w-0">
                <span class="w-2 h-2 rounded-full bg-blue-500 flex-shrink-0"></span>
                <h2 class="text-[11px] font-bold uppercase tracking-wider reports-heading truncate">
                  {{ $t('reports.deviceAvailabilityTitle') || 'Device Fleet Availability & Comm Audit' }}
                </h2>
              </div>
              <span class="text-[10.5px] font-mono reports-muted flex-shrink-0">
                ({{ filteredDevices.length }} {{ filteredDevices.length === 1 ? 'node' : 'nodes' }})
              </span>
            </div>

            <!-- Compact Node Rows (Comprehensive Split Mode) -->
            <div v-if="filters.report_type === 'COMPREHENSIVE'" class="divide-y reports-row-divide max-h-[165px] overflow-y-auto reports-scroll-pane">
              <div v-if="filteredDevices.length === 0" class="px-3.5 py-4 text-center text-xs reports-muted">
                No fleet nodes matched filter.
              </div>
              <div v-for="d in filteredDevices" :key="d.device_id" class="px-3.5 py-2.5 space-y-1.5 reports-row">
                <div class="flex flex-wrap items-center justify-between gap-1.5 text-[11px] font-mono">
                  <div class="flex items-center gap-1.5 min-w-0">
                    <span class="font-bold reports-heading truncate" :title="d.device_name">{{ d.device_code }}</span>
                    <span
                      class="px-1.5 py-0.5 rounded text-[9.5px] font-bold uppercase"
                      :class="isDeviceOnline(d.status) ? 'bg-emerald-500/15 text-emerald-500' : 'bg-rose-500/15 text-rose-500'"
                    >
                      {{ d.status }}
                    </span>
                  </div>
                  <div class="flex items-center gap-2 tabular-nums text-[10.5px]">
                    <span class="reports-muted">{{ d.latency_ms }} ms</span>
                    <span>
                      {{ d.success_count }} OK /
                      <span :class="d.failed_count > 0 ? 'text-rose-500 font-bold' : 'text-rose-400'">{{ d.failed_count }} ERR</span>
                    </span>
                    <strong class="text-emerald-500">{{ (d.availability_percent || 0).toFixed(1) }}%</strong>
                  </div>
                </div>
                <div class="h-1.5 rounded-full reports-track overflow-hidden">
                  <div
                    class="h-full rounded-full transition-all"
                    :class="(d.availability_percent || 0) >= 95 ? 'bg-emerald-500' : 'bg-amber-500'"
                    :style="{ width: clampPct(d.availability_percent) + '%' }"
                  ></div>
                </div>
              </div>
            </div>

            <!-- Expanded Table (Fleet Comm Tab Mode) -->
            <div v-else class="overflow-x-auto">
              <table class="w-full text-left border-collapse">
                <thead class="reports-thead text-[10px] font-bold uppercase tracking-wider border-b reports-divider">
                  <tr>
                    <th class="px-3.5 py-2">Device Code</th>
                    <th class="px-3 py-2">Device Name</th>
                    <th class="px-3 py-2">Status</th>
                    <th class="px-3 py-2 text-right">Latency</th>
                    <th class="px-3 py-2 text-right">Packets (OK / ERR)</th>
                    <th class="px-3 py-2 text-right">Parameters</th>
                    <th class="px-3.5 py-2 text-right">Availability</th>
                  </tr>
                </thead>
                <tbody class="divide-y reports-row-divide text-xs">
                  <tr v-for="d in filteredDevices" :key="d.device_id" class="reports-row">
                    <td class="px-3.5 py-2.5 font-mono font-bold reports-heading">{{ d.device_code }}</td>
                    <td class="px-3 py-2.5 reports-strong">{{ d.device_name }}</td>
                    <td class="px-3 py-2.5">
                      <span
                        class="px-2 py-0.5 rounded-full text-[10px] font-bold uppercase"
                        :class="isDeviceOnline(d.status) ? 'bg-emerald-500/15 text-emerald-500' : 'bg-rose-500/15 text-rose-500'"
                      >
                        {{ d.status }}
                      </span>
                    </td>
                    <td class="px-3 py-2.5 text-right font-mono tabular-nums reports-muted">{{ d.latency_ms }} ms</td>
                    <td class="px-3 py-2.5 text-right font-mono tabular-nums">
                      {{ d.success_count }} / <span class="text-rose-500">{{ d.failed_count }}</span>
                    </td>
                    <td class="px-3 py-2.5 text-right font-mono tabular-nums reports-strong">{{ d.parameters_count }} active</td>
                    <td class="px-3.5 py-2.5 text-right font-mono font-bold tabular-nums text-emerald-500">
                      {{ (d.availability_percent || 0).toFixed(1) }}%
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div class="px-3.5 py-1.5 border-t reports-divider reports-footer-strip text-[10.5px] font-mono text-center">
              {{ $t('reports.totalFrames') || 'Total Frames' }}:
              <strong class="reports-strong">{{ commSummary.total_frames || commSummary.success_frames || 0 }}</strong>
              &bull; {{ $t('reports.successRate') || 'Success Rate' }}:
              <strong class="text-emerald-500">{{ (commSummary.success_rate_percent || reportData.overall_availability || 0).toFixed(1) }}%</strong>
            </div>
          </div>

          <!-- CARD B: OPERATIONAL ALARMS & EXCEPTION LOG -->
          <div v-if="showAlarmsSection" class="reports-card rounded-xl overflow-hidden page-break">
            <div class="px-3.5 py-2.5 border-b reports-divider space-y-1.5">
              <div class="flex items-center justify-between gap-2">
                <div class="flex items-center gap-2 min-w-0">
                  <span class="w-2 h-2 rounded-full bg-rose-500 flex-shrink-0"></span>
                  <h2 class="text-[11px] font-bold uppercase tracking-wider reports-heading truncate">
                    {{ $t('reports.alarmAuditTitle') || 'Operational Alarms & Exception Log' }}
                  </h2>
                </div>
                <span class="text-[10.5px] font-mono reports-muted flex-shrink-0">
                  ({{ alarmSummary.total_alarms || 0 }} recorded)
                </span>
              </div>

              <div class="flex flex-wrap items-center justify-between gap-1.5 text-[10px] font-mono">
                <span class="uppercase font-bold tracking-wider reports-muted">Exception Log</span>
                <div class="flex items-center gap-1">
                  <span class="px-1.5 py-0.5 rounded font-semibold reports-neutral-pill">
                    {{ alarmSummary.active_count || 0 }} Active
                  </span>
                  <span class="px-1.5 py-0.5 rounded font-bold bg-rose-600 text-white">
                    {{ alarmSummary.critical_count || 0 }} Crit
                  </span>
                  <span class="px-1.5 py-0.5 rounded font-bold bg-amber-500/20 text-amber-500">
                    {{ alarmSummary.warning_count || 0 }} Warn
                  </span>
                </div>
              </div>
            </div>

            <div v-if="filteredAlarms.length === 0" class="px-4 py-6 text-center text-[11.5px] reports-muted">
              {{ $t('reports.noAlarmsRecorded') || 'No critical or warning alarms occurred during this reporting interval.' }}
            </div>

            <!-- Compact Exception Feed (Comprehensive Split Mode) -->
            <div
              v-else-if="filters.report_type === 'COMPREHENSIVE'"
              class="divide-y reports-row-divide max-h-[255px] overflow-y-auto reports-scroll-pane"
            >
              <div
                v-for="a in filteredAlarms"
                :key="a.id"
                class="px-3.5 py-2 flex items-center justify-between gap-2 text-[11px] reports-row"
              >
                <div class="flex items-center gap-2 min-w-0">
                  <span class="font-mono font-bold text-[10.5px] reports-heading whitespace-nowrap">
                    {{ a.alarm_code }}
                  </span>
                  <span
                    class="px-1.5 py-0.5 rounded text-[9.5px] font-bold uppercase flex-shrink-0"
                    :class="a.severity === 'CRITICAL' ? 'bg-rose-600 text-white' : 'bg-amber-500/20 text-amber-500'"
                  >
                    {{ a.severity }}
                  </span>
                  <div class="min-w-0 leading-tight">
                    <div
                      class="text-[9.5px] font-mono font-bold uppercase"
                      :class="a.status === 'ACTIVE' ? 'text-rose-500' : 'text-emerald-500'"
                    >
                      {{ a.status }}
                    </div>
                    <div class="truncate reports-strong text-[11px]" :title="a.message">
                      {{ a.message }}
                    </div>
                  </div>
                </div>
                <span class="font-mono tabular-nums text-[10.5px] reports-muted whitespace-nowrap flex-shrink-0">
                  {{ formatShortTime(a.triggered_at) }}
                </span>
              </div>
            </div>

            <!-- Full-Width Alarms Table (Alarms Audit Tab Mode) -->
            <div v-else class="overflow-x-auto">
              <table class="w-full text-left border-collapse">
                <thead class="reports-thead text-[10px] font-bold uppercase tracking-wider border-b reports-divider">
                  <tr>
                    <th class="px-3.5 py-2">Code</th>
                    <th class="px-3 py-2">Device</th>
                    <th class="px-3 py-2">Severity</th>
                    <th class="px-3 py-2">Status</th>
                    <th class="px-3 py-2">Message</th>
                    <th class="px-3.5 py-2 text-right">Triggered Time</th>
                  </tr>
                </thead>
                <tbody class="divide-y reports-row-divide text-xs">
                  <tr v-for="a in filteredAlarms" :key="a.id" class="reports-row">
                    <td class="px-3.5 py-2.5 font-mono font-bold reports-heading">{{ a.alarm_code }}</td>
                    <td class="px-3 py-2.5 reports-strong">{{ a.device_name || '-' }}</td>
                    <td class="px-3 py-2.5">
                      <span
                        class="px-1.5 py-0.5 rounded text-[10px] font-bold uppercase"
                        :class="a.severity === 'CRITICAL' ? 'bg-rose-600 text-white' : 'bg-amber-500/20 text-amber-500'"
                      >
                        {{ a.severity }}
                      </span>
                    </td>
                    <td class="px-3 py-2.5 font-mono text-[11px]" :class="a.status === 'ACTIVE' ? 'text-rose-500 font-bold' : 'text-emerald-500 font-semibold'">
                      {{ a.status }}
                    </td>
                    <td class="px-3 py-2.5 reports-strong max-w-md truncate" :title="a.message">{{ a.message }}</td>
                    <td class="px-3.5 py-2.5 text-right font-mono tabular-nums text-[11px] reports-muted">
                      {{ formatTimestamp(a.triggered_at) }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import {
  exportOperationalReportXLS,
  exportOperationalReportPDF,
} from '../utils/reportsExport';

export default {
  name: 'ReportsView',
  data() {
    return {
      loading: false,
      error: null,
      isDarkMode: false,
      themeObserver: null,
      searchQuery: '',
      selectedPreset: '24h',
      presets: [
        { id: '1h', label: 'Last 1 Hour', shortLabel: '1H' },
        { id: '24h', label: 'Last 24 Hours', shortLabel: '24H' },
        { id: '7d', label: 'Last 7 Days', shortLabel: '7D' },
        { id: '30d', label: 'Last 30 Days', shortLabel: '30D' },
        { id: 'custom', label: 'Custom', shortLabel: 'Custom' },
      ],
      filters: {
        startTime: '',
        endTime: '',
        deviceId: '',
        report_type: 'COMPREHENSIVE',
        timezone: 'Asia/Jakarta',
      },
      devices: [],
      reportData: null,
    };
  },
  computed: {
    isCloudMode() {
      const sys = (this.$store && this.$store.state && this.$store.state.systemStatus) || {};
      return Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
    },
    activeTenantSiteKey() {
      const tId = (this.$store && this.$store.state && this.$store.state.activeTenantId) || 0;
      const sId = (this.$store && this.$store.state && this.$store.state.activeSiteId) || 0;
      return `${tId}:${sId}`;
    },
    commSummary() {
      return (this.reportData && this.reportData.communication_summary) || {};
    },
    alarmSummary() {
      return (this.reportData && this.reportData.alarm_summary) || {
        total_alarms: 0,
        critical_count: 0,
        warning_count: 0,
        active_count: 0,
        cleared_count: 0,
        recent_alarms: [],
      };
    },
    deviceSummaries() {
      return (this.reportData && Array.isArray(this.reportData.device_summaries))
        ? this.reportData.device_summaries
        : [];
    },
    parameterSummaries() {
      return (this.reportData && Array.isArray(this.reportData.parameter_summaries))
        ? this.reportData.parameter_summaries
        : [];
    },
    recentAlarms() {
      return Array.isArray(this.alarmSummary.recent_alarms) ? this.alarmSummary.recent_alarms : [];
    },
    filteredParameters() {
      if (!this.searchQuery) return this.parameterSummaries;
      const q = this.searchQuery.toLowerCase();
      return this.parameterSummaries.filter((p) => (
        String(p.parameter_name || '').toLowerCase().includes(q) ||
        String(p.parameter_code || '').toLowerCase().includes(q) ||
        String(p.device_code || '').toLowerCase().includes(q) ||
        String(p.device_name || '').toLowerCase().includes(q) ||
        String(p.unit || '').toLowerCase().includes(q)
      ));
    },
    filteredDevices() {
      if (!this.searchQuery) return this.deviceSummaries;
      const q = this.searchQuery.toLowerCase();
      return this.deviceSummaries.filter((d) => (
        String(d.device_code || '').toLowerCase().includes(q) ||
        String(d.device_name || '').toLowerCase().includes(q) ||
        String(d.status || '').toLowerCase().includes(q)
      ));
    },
    filteredAlarms() {
      if (!this.searchQuery) return this.recentAlarms;
      const q = this.searchQuery.toLowerCase();
      return this.recentAlarms.filter((a) => (
        String(a.alarm_code || '').toLowerCase().includes(q) ||
        String(a.device_name || '').toLowerCase().includes(q) ||
        String(a.message || '').toLowerCase().includes(q) ||
        String(a.severity || '').toLowerCase().includes(q) ||
        String(a.status || '').toLowerCase().includes(q)
      ));
    },
    reportTabs() {
      return [
        {
          value: 'COMPREHENSIVE',
          label: this.$t('reports.tabComprehensive') || 'Comprehensive',
          count: null,
        },
        {
          value: 'TELEMETRY',
          label: this.$t('reports.tabTelemetry') || 'Telemetry & Quality',
          count: this.parameterSummaries.length,
        },
        {
          value: 'COMMUNICATION',
          label: this.$t('reports.tabComm') || 'Fleet Comm',
          count: this.deviceSummaries.length,
        },
        {
          value: 'ALARMS',
          label: this.$t('reports.tabAlarms') || 'Alarms Audit',
          count: Number(this.alarmSummary.total_alarms || 0),
        },
      ];
    },
    showTelemetrySection() {
      return this.filters.report_type === 'COMPREHENSIVE' || this.filters.report_type === 'TELEMETRY';
    },
    showCommSection() {
      return this.filters.report_type === 'COMPREHENSIVE' || this.filters.report_type === 'COMMUNICATION';
    },
    showAlarmsSection() {
      return this.filters.report_type === 'COMPREHENSIVE' || this.filters.report_type === 'ALARMS';
    },
    onlineDevicesCount() {
      return this.deviceSummaries.filter((d) => this.isDeviceOnline(d.status)).length;
    },
    dominantSourceLabel() {
      const hasRaw = this.parameterSummaries.some((p) => p.data_source === 'RAW_TELEMETRY');
      const hasProc = this.parameterSummaries.some((p) => p.data_source === 'PROCESSED');
      if (hasRaw && hasProc) return 'Multi-Source RAW + PROC';
      if (hasRaw) return 'Multi-Source RAW';
      if (hasProc) return 'Processed Stream';
      return 'Live Snapshot';
    },
    overallQualityBreakdown() {
      let good = 0;
      let bad = 0;
      let unc = 0;
      for (const p of this.parameterSummaries) {
        good += Number(p.good_count || 0);
        bad += Number(p.bad_count || 0);
        unc += Number(p.uncertain_count || 0);
      }
      const total = good + bad + unc;
      if (total <= 0) {
        const g = this.clampPct(this.reportData ? this.reportData.overall_quality_good : 100);
        return { good: g, uncertain: 0, bad: Math.max(0, 100 - g), badCount: 0, uncCount: 0 };
      }
      return {
        good: Math.max(0, Math.min(100, (good / total) * 100)),
        uncertain: Math.max(0, Math.min(100, (unc / total) * 100)),
        bad: Math.max(0, Math.min(100, (bad / total) * 100)),
        badCount: bad,
        uncCount: unc,
      };
    },
    badUncertainFooterText() {
      const { badCount, uncCount } = this.overallQualityBreakdown;
      if (badCount === 0 && uncCount === 0) {
        return this.$t('reports.zeroBadFrames') || 'Zero bad/uncertain frames';
      }
      return `${badCount} bad • ${uncCount} uncertain frames`;
    },
    shortGeneratedTime() {
      if (!this.reportData || !this.reportData.generated_at) {
        return (this.reportData && this.reportData.generated_at_formatted) || '';
      }
      const d = new Date(this.reportData.generated_at);
      if (Number.isNaN(d.getTime())) return this.reportData.generated_at_formatted || '';
      return d.toLocaleTimeString('en-GB', { hour12: false });
    },
    activePresetShortLabel() {
      const found = this.presets.find((p) => p.id === this.selectedPreset);
      return found ? found.shortLabel : 'Custom';
    },
  },
  watch: {
    '$store.state.theme'() {
      this.syncDarkMode();
    },
    activeTenantSiteKey() {
      this.filters.deviceId = '';
      this.fetchDevices();
      this.fetchReportPreview();
    },
  },
  mounted() {
    this.syncDarkMode();
    if (typeof MutationObserver !== 'undefined' && typeof document !== 'undefined') {
      this.themeObserver = new MutationObserver(() => this.syncDarkMode());
      this.themeObserver.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['class'],
      });
    }
    this.initDefaultDates();
    this.fetchDevices();
    this.fetchReportPreview();
  },
  beforeDestroy() {
    if (this.themeObserver) {
      this.themeObserver.disconnect();
      this.themeObserver = null;
    }
  },
  methods: {
    syncDarkMode() {
      if (typeof document !== 'undefined' && document.documentElement) {
        this.isDarkMode = document.documentElement.classList.contains('dark');
      }
    },
    clampPct(val) {
      const n = Number(val);
      if (!Number.isFinite(n)) return 0;
      return Math.max(0, Math.min(100, n));
    },
    isDeviceOnline(status) {
      const s = String(status || '').toUpperCase();
      return s === 'ONLINE' || s === 'ACTIVE' || s === 'CONNECTED' || s === 'READY';
    },
    selectReportTab(type) {
      this.filters.report_type = type;
    },
    onManualDateChange() {
      this.selectedPreset = 'custom';
    },
    initDefaultDates() {
      const now = new Date();
      const past24h = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      this.filters.endTime = this.toLocalISOString(now);
      this.filters.startTime = this.toLocalISOString(past24h);
    },
    toLocalISOString(date) {
      const pad = (n) => (n < 10 ? '0' + n : n);
      return (
        date.getFullYear() +
        '-' +
        pad(date.getMonth() + 1) +
        '-' +
        pad(date.getDate()) +
        'T' +
        pad(date.getHours()) +
        ':' +
        pad(date.getMinutes())
      );
    },
    applyPreset(presetId) {
      this.selectedPreset = presetId;
      if (presetId === 'custom') return;

      const now = new Date();
      this.filters.endTime = this.toLocalISOString(now);

      if (presetId === '1h') {
        this.filters.startTime = this.toLocalISOString(new Date(now.getTime() - 60 * 60 * 1000));
      } else if (presetId === '24h') {
        this.filters.startTime = this.toLocalISOString(new Date(now.getTime() - 24 * 60 * 60 * 1000));
      } else if (presetId === '7d') {
        this.filters.startTime = this.toLocalISOString(new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000));
      } else if (presetId === '30d') {
        this.filters.startTime = this.toLocalISOString(new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000));
      }
      this.fetchReportPreview();
    },
    async fetchDevices() {
      try {
        const res = await axios.get('/api/devices');
        if (res.data && res.data.data) {
          this.devices = res.data.data;
        }
      } catch (err) {
        console.warn('Failed to load device catalog for reporting:', err);
      }
    },
    buildQueryParams() {
      const params = new URLSearchParams();
      if (this.filters.startTime) {
        params.append('start_time', new Date(this.filters.startTime).toISOString());
      }
      if (this.filters.endTime) {
        params.append('end_time', new Date(this.filters.endTime).toISOString());
      }
      if (this.filters.deviceId) {
        params.append('device_id', this.filters.deviceId);
      }
      if (this.filters.report_type) {
        params.append('report_type', this.filters.report_type);
      }
      if (this.filters.timezone) {
        params.append('timezone', this.filters.timezone);
      }
      return params;
    },
    async fetchReportPreview() {
      this.loading = true;
      this.error = null;
      try {
        const params = this.buildQueryParams();
        const res = await axios.get(`/api/reports/preview?${params.toString()}`);
        if (res.data && res.data.data) {
          this.reportData = res.data.data;
        }
      } catch (err) {
        this.error = err.response?.data?.error || err.message || 'Failed to generate report preview';
      } finally {
        this.loading = false;
      }
    },
    exportXlsReport() {
      if (!this.reportData) return;
      try {
        exportOperationalReportXLS(this.reportData, this.filters.report_type || 'COMPREHENSIVE');
      } catch (err) {
        console.error('[ReportsView XLS Export Error]:', err);
        alert('Gagal mengekspor Excel: ' + (err.message || err));
      }
    },
    exportPdfReport() {
      if (!this.reportData) return;
      try {
        exportOperationalReportPDF(this.reportData, this.filters.report_type || 'COMPREHENSIVE');
      } catch (err) {
        console.error('[ReportsView PDF Export Error]:', err);
        alert('Gagal mengekspor PDF: ' + (err.message || err));
      }
    },
    async exportReport(format) {
      try {
        const params = this.buildQueryParams();
        params.append('format', format);
        if (format === 'html') {
          params.append('download', '1');
        }
        const token = this.$store?.state?.token || localStorage.getItem('token') || localStorage.getItem('datalogger_token') || '';
        if (token) {
          params.append('token', token);
        }

        const res = await axios.get('/api/reports/generate', {
          params: Object.fromEntries(params.entries()),
          responseType: 'blob',
          timeout: 45000,
        });

        const isCsv = format === 'csv';
        const mimeType = isCsv ? 'text/csv;charset=utf-8' : 'text/html;charset=utf-8';
        const ext = isCsv ? 'csv' : 'html';
        const slug = String(this.filters.report_type || 'comprehensive').toLowerCase();
        const filename = `report_${slug}_${Date.now()}.${ext}`;
        const blob = new Blob([res.data], { type: mimeType });

        if (isCsv) {
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = filename;
          document.body.appendChild(a);
          a.click();
          document.body.removeChild(a);
          URL.revokeObjectURL(url);
        } else {
          const blobUrl = URL.createObjectURL(blob);
          const reportWin = window.open(blobUrl, '_blank');
          if (!reportWin) {
            const a = document.createElement('a');
            a.href = blobUrl;
            a.download = filename;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            URL.revokeObjectURL(blobUrl);
          }
        }
      } catch (err) {
        console.error('[ReportsView Export Error]:', err);
        alert('Gagal mengekspor laporan: ' + (err?.response?.data?.error || err.message));
      }
    },
    printReport() {
      window.print();
    },
    formatNum(val) {
      if (val === null || val === undefined || Number.isNaN(Number(val))) return '-';
      return Number(val).toFixed(2);
    },
    hasFirstLast(p) {
      return (
        p &&
        p.first_value !== null &&
        p.first_value !== undefined &&
        p.last_value !== null &&
        p.last_value !== undefined &&
        !Number.isNaN(Number(p.first_value)) &&
        !Number.isNaN(Number(p.last_value))
      );
    },
    formatDelta(p) {
      if (!this.hasFirstLast(p)) return '0.00';
      const diff = Number(p.last_value) - Number(p.first_value);
      const sign = diff > 0 ? '+' : '';
      return `${sign}${diff.toFixed(2)}`;
    },
    deltaClass(p) {
      if (!this.hasFirstLast(p)) return 'reports-muted';
      const diff = Number(p.last_value) - Number(p.first_value);
      if (Math.abs(diff) < 0.005) return 'reports-muted';
      return diff > 0 ? 'text-emerald-500' : 'text-amber-500';
    },
    computeVisualRangeStyle(p) {
      if (!p || p.min_value === null || p.max_value === null || p.avg_value === null) {
        return { left: '0%', width: '0%' };
      }
      const min = Number(p.min_value);
      const max = Number(p.max_value);
      const avg = Number(p.avg_value);
      if (!Number.isFinite(min) || !Number.isFinite(max) || !Number.isFinite(avg)) {
        return { left: '0%', width: '0%' };
      }
      if (max <= min) {
        return { left: '30%', width: '40%' };
      }
      const ratio = Math.max(0, Math.min(1, (avg - min) / (max - min)));
      const widthPct = 42;
      const leftPct = Math.max(4, Math.min(96 - widthPct, ratio * (100 - widthPct)));
      return {
        left: `${leftPct.toFixed(1)}%`,
        width: `${widthPct}%`,
      };
    },
    shortSourceLabel(src) {
      if (src === 'RAW_TELEMETRY') return 'RAW';
      if (src === 'PROCESSED') return 'PROC';
      if (src === 'CURRENT_SNAPSHOT') return 'SNAP';
      return src || '-';
    },
    sourceBadgeClass(src) {
      if (src === 'RAW_TELEMETRY') return 'bg-blue-500/15 text-blue-500';
      if (src === 'PROCESSED') return 'bg-emerald-500/15 text-emerald-500';
      return 'bg-slate-500/15 reports-muted';
    },
    formatTimestamp(ts) {
      if (!ts) return '-';
      const d = new Date(ts);
      if (Number.isNaN(d.getTime())) return '-';
      return d.toLocaleString('en-GB', { hour12: false });
    },
    formatShortTime(ts) {
      if (!ts) return '-';
      const d = new Date(ts);
      if (Number.isNaN(d.getTime())) return '-';
      return d.toLocaleTimeString('en-GB', { hour12: false });
    },
  },
};
</script>

<style scoped>
.reports-workbench {
  color: #1e293b;
}

.reports-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}

.reports-heading {
  color: #0f172a;
}

.reports-strong {
  color: #1e293b;
}

.reports-muted {
  color: #64748b;
}

.reports-divider {
  border-color: #e2e8f0;
}

.reports-row-divide > :not([hidden]) ~ :not([hidden]) {
  border-color: #f1f5f9;
}

.reports-thead {
  background: #f8fafc;
  color: #64748b;
}

.reports-row:hover {
  background: rgba(241, 245, 249, 0.65);
}

.reports-footer-strip {
  background: #f8fafc;
  color: #64748b;
}

.reports-track {
  background: #e2e8f0;
}

.reports-meta-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 10.5px;
  background: #f1f5f9;
  color: #334155;
  border: 1px solid #e2e8f0;
}

.reports-engine-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 10px;
  font-weight: 700;
  background: rgba(16, 185, 129, 0.14);
  color: #047857;
  border: 1px solid rgba(16, 185, 129, 0.28);
}

.reports-btn-secondary {
  background: #f8fafc;
  color: #1e293b;
  border: 1px solid #cbd5e1;
}

.reports-btn-secondary:hover:not(:disabled) {
  background: #f1f5f9;
}

.reports-btn-emerald {
  background: #059669;
  color: #ffffff;
  border: 1px solid #047857;
}

.reports-btn-emerald:hover:not(:disabled) {
  background: #10b981;
}

.reports-tab-bar {
  border-color: #e2e8f0;
}

.reports-tab-active {
  color: #2563eb;
  border-color: #2563eb;
}

.reports-tab-idle {
  color: #64748b;
  border-color: transparent;
}

.reports-tab-idle:hover {
  color: #0f172a;
}

.reports-tab-count {
  padding: 1px 5px;
  border-radius: 9999px;
  font-size: 9.5px;
  background: #f1f5f9;
  color: #475569;
}

.reports-segmented {
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
}

.reports-preset-active {
  background: #ffffff;
  color: #2563eb;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.06);
}

.reports-preset-idle {
  color: #64748b;
}

.reports-preset-idle:hover {
  color: #0f172a;
}

.reports-input {
  background: #f8fafc;
  border: 1px solid #cbd5e1;
  color: #0f172a;
  outline: none;
}

.reports-input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.25);
}

.reports-neutral-pill {
  background: #f1f5f9;
  color: #334155;
  border: 1px solid #cbd5e1;
}

.reports-skeleton {
  background: #e2e8f0;
}

/* =========================================================================
   HIGH-CONTRAST INDUSTRIAL DARK MODE (Vue 2 Scoped Class — No :global)
   ========================================================================= */
.reports-dark {
  color: #e2e8f0;
}

.reports-dark .reports-card {
  background: #111827;
  border-color: #1e293b;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
}

.reports-dark .reports-title-icon {
  color: #60a5fa;
}

.reports-dark .reports-heading {
  color: #f8fafc;
}

.reports-dark .reports-strong {
  color: #e2e8f0;
}

.reports-dark .reports-muted {
  color: #94a3b8;
}

.reports-dark .reports-divider {
  border-color: #1e293b;
}

.reports-dark .reports-row-divide > :not([hidden]) ~ :not([hidden]) {
  border-color: rgba(30, 41, 59, 0.75);
}

.reports-dark .reports-thead {
  background: #0b0f19;
  color: #94a3b8;
}

.reports-dark .reports-row:hover {
  background: rgba(30, 41, 59, 0.45);
}

.reports-dark .reports-footer-strip {
  background: rgba(11, 15, 25, 0.75);
  color: #94a3b8;
}

.reports-dark .reports-track {
  background: #1e293b;
}

.reports-dark .reports-meta-pill {
  background: #0b0f19;
  color: #cbd5e1;
  border-color: #1e293b;
}

.reports-dark .reports-engine-badge {
  background: rgba(16, 185, 129, 0.16);
  color: #34d399;
  border-color: rgba(16, 185, 129, 0.32);
}

.reports-dark .reports-btn-secondary {
  background: #0b0f19;
  color: #e2e8f0;
  border-color: #334155;
}

.reports-dark .reports-btn-secondary:hover:not(:disabled) {
  background: #1e293b;
}

.reports-dark .reports-tab-bar {
  border-color: #1e293b;
}

.reports-dark .reports-tab-active {
  color: #60a5fa;
  border-color: #3b82f6;
}

.reports-dark .reports-tab-idle {
  color: #94a3b8;
}

.reports-dark .reports-tab-idle:hover {
  color: #f8fafc;
}

.reports-dark .reports-tab-count {
  background: #1e293b;
  color: #cbd5e1;
}

.reports-dark .reports-segmented {
  background: #0b0f19;
  border-color: #1e293b;
}

.reports-dark .reports-preset-active {
  background: #1e293b;
  color: #60a5fa;
}

.reports-dark .reports-preset-idle {
  color: #94a3b8;
}

.reports-dark .reports-preset-idle:hover {
  color: #f8fafc;
}

.reports-dark .reports-input {
  background: #0b0f19;
  border-color: #1e293b;
  color: #e2e8f0;
  color-scheme: dark;
}

.reports-dark .reports-neutral-pill {
  background: #1e293b;
  color: #cbd5e1;
  border-color: #334155;
}

.reports-dark .reports-skeleton {
  background: #1e293b;
}

.print-only-header {
  display: none;
}

@media print {
  .no-print {
    display: none !important;
  }
  .print-only-header {
    display: block !important;
  }
  .reports-split-grid {
    display: block !important;
  }
  .reports-scroll-pane {
    max-height: none !important;
    overflow: visible !important;
  }
  .reports-card {
    background: #ffffff !important;
    border-color: #cbd5e1 !important;
    box-shadow: none !important;
    margin-bottom: 12px !important;
  }
  .page-break {
    page-break-before: auto;
  }
  body, :host {
    background: #ffffff !important;
    color: #000000 !important;
  }
}
</style>
