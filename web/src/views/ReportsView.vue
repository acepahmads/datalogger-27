<template>
  <div class="space-y-4 font-sans text-slate-800 dark:text-slate-200">
    <!-- Header with Title & Action Buttons -->
    <div class="no-print bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center space-x-3">
        <div class="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-600 dark:text-blue-400 flex items-center justify-center flex-shrink-0">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
          </svg>
        </div>
        <div>
          <h1 class="text-base font-bold text-slate-900 dark:text-white tracking-tight flex items-center gap-2">
            {{ $t('reports.title') || 'Operational Reporting' }}
            <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20">
              #79
            </span>
          </h1>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            {{ $t('reports.subtitle') || 'High-integrity operational telemetry, fleet availability, alarm audits, and compliance export' }}
          </p>
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex flex-wrap items-center gap-2">
        <button id="btn-print-report"
                @click="printReport"
                :disabled="loading || !reportData"
                class="px-3 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs font-semibold transition-colors flex items-center gap-1.5 disabled:opacity-50">
          <svg class="w-4 h-4 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <polyline points="6 9 6 2 18 2 18 9"></polyline>
            <path d="M6 18H4a2 2 0 01-2-2v-5a2 2 0 012-2h16a2 2 0 012 2v5a2 2 0 01-2 2h-2"></path>
            <rect x="6" y="14" width="12" height="8"></rect>
          </svg>
          {{ $t('reports.printPdf') || 'Print / Save PDF' }}
        </button>

        <button id="btn-export-csv"
                @click="exportReport('csv')"
                :disabled="loading || !reportData"
                class="px-3 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs font-semibold transition-colors flex items-center gap-1.5 disabled:opacity-50">
          <svg class="w-4 h-4 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
          </svg>
          {{ $t('reports.exportCsv') || 'CSV Export' }}
        </button>

        <button id="btn-download-html"
                @click="exportReport('html')"
                :disabled="loading || !reportData"
                class="px-3 py-1.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold shadow-xs transition-colors flex items-center gap-1.5 disabled:opacity-50">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path>
          </svg>
          {{ $t('reports.downloadHtml') || 'Download HTML' }}
        </button>
      </div>
    </div>

    <!-- Filter & Configuration Toolbar -->
    <div class="no-print bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm space-y-4">
      <!-- Quick Range Presets -->
      <div class="flex flex-wrap items-center justify-between gap-3 text-xs">
        <div class="flex items-center gap-1 p-1 bg-slate-100 dark:bg-[#0B0F19] rounded-xl border border-slate-200 dark:border-slate-800">
          <button v-for="preset in presets"
                  :key="preset.id"
                  @click="applyPreset(preset.id)"
                  class="px-3 py-1.5 rounded-lg transition-all"
                  :class="selectedPreset === preset.id ? 'bg-white dark:bg-slate-800 text-blue-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ preset.label }}
          </button>
        </div>

        <!-- Report Type Selection -->
        <div class="flex items-center gap-2">
          <label class="text-3xs uppercase tracking-wider text-slate-500 font-semibold">{{ $t('reports.type') || 'Report Type' }}:</label>
          <select v-model="filters.report_type"
                  class="bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 text-slate-800 dark:text-slate-200 rounded-xl px-3 py-1.5 text-xs focus:ring-1 focus:ring-blue-500">
            <option value="COMPREHENSIVE">Comprehensive Operational</option>
            <option value="TELEMETRY">Telemetry & Quality Focus</option>
            <option value="ALARMS">Alarms & Events Audit</option>
            <option value="COMMUNICATION">Hardware Comm Availability</option>
          </select>
        </div>
      </div>

      <!-- Specific Filters Row -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 text-xs">
        <!-- Start Time -->
        <div>
          <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('reports.startTime') || 'Start Time' }}</label>
          <input type="datetime-local"
                 v-model="filters.startTime"
                 class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200 font-mono text-2xs focus:ring-1 focus:ring-blue-500" />
        </div>

        <!-- End Time -->
        <div>
          <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('reports.endTime') || 'End Time' }}</label>
          <input type="datetime-local"
                 v-model="filters.endTime"
                 class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200 font-mono text-2xs focus:ring-1 focus:ring-blue-500" />
        </div>

        <!-- Device Filter -->
        <div>
          <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('reports.device') || 'Device Filter' }}</label>
          <select v-model="filters.deviceId"
                  class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200 focus:ring-1 focus:ring-blue-500">
            <option value="">{{ $t('reports.allDevices') || 'All Fleet Devices' }}</option>
            <option v-for="d in devices" :key="d.id" :value="d.id">
              {{ d.name }} ({{ d.code }})
            </option>
          </select>
        </div>

        <!-- Timezone & Refresh Action -->
        <div class="flex items-end gap-2">
          <div class="flex-1">
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('reports.timezone') || 'Timezone' }}</label>
            <select v-model="filters.timezone"
                    class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200 focus:ring-1 focus:ring-blue-500">
              <option value="Asia/Jakarta">Asia/Jakarta (WIB, UTC+7)</option>
              <option value="UTC">UTC (Universal Coordinated)</option>
            </select>
          </div>

          <button id="btn-fetch-report"
                  @click="fetchReportPreview"
                  :disabled="loading"
                  class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-semibold transition-colors flex items-center gap-1.5 shadow-xs disabled:opacity-50">
            <svg class="w-4 h-4" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            {{ loading ? ($t('common.loading') || 'Compiling...') : ($t('reports.preview') || 'Preview') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-if="error" class="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-600 dark:text-rose-400 text-xs flex items-center justify-between">
      <div class="flex items-center gap-2">
        <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
        <span>{{ error }}</span>
      </div>
      <button @click="fetchReportPreview" class="px-3 py-1 bg-rose-500/20 hover:bg-rose-500/30 rounded-lg font-semibold">
        {{ $t('common.retry') || 'Retry' }}
      </button>
    </div>

    <!-- Loading Skeleton -->
    <div v-if="loading && !reportData" class="space-y-4 animate-pulse">
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div v-for="i in 4" :key="i" class="h-28 bg-slate-200 dark:bg-slate-800 rounded-2xl"></div>
      </div>
      <div class="h-64 bg-slate-200 dark:bg-slate-800 rounded-2xl"></div>
    </div>

    <!-- Empty State -->
    <div v-else-if="!loading && (!reportData || (reportData.total_samples === 0 && reportData.alarm_summary.total_alarms === 0))"
         class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-12 text-center space-y-3">
      <div class="w-12 h-12 rounded-full bg-amber-500/10 border border-amber-500/20 text-amber-500 flex items-center justify-center mx-auto">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <circle cx="12" cy="12" r="10"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
      </div>
      <h3 class="text-sm font-bold text-slate-800 dark:text-slate-200">{{ $t('reports.noDataTitle') || 'No Telemetry Records in Period' }}</h3>
      <p class="text-xs text-slate-500 dark:text-slate-400 max-w-md mx-auto">
        {{ $t('reports.noDataDesc') || 'No raw sensor data or alarm records were captured during the selected interval. Try adjusting the date range or select all devices.' }}
      </p>
      <button @click="applyPreset('24h')" class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-semibold transition-colors">
        {{ $t('reports.resetLast24h') || 'View Last 24 Hours' }}
      </button>
    </div>

    <!-- Active Report Preview Body -->
    <div v-else-if="reportData" class="space-y-4">
      <!-- Report Metadata Header (Visible in print too) -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 shadow-sm border-l-4 border-l-blue-500">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-100 dark:border-slate-800/80 pb-3">
          <div>
            <div class="text-3xs uppercase tracking-widest text-blue-600 dark:text-blue-400 font-black">DATALOGGER 27 &bull; OPERATIONAL REPORT</div>
            <h2 class="text-lg font-bold text-slate-900 dark:text-white">{{ reportData.title }}</h2>
          </div>
          <div class="text-right">
            <span class="inline-block px-2.5 py-1 rounded-lg text-2xs font-mono font-bold bg-slate-100 dark:bg-[#0B0F19] text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-800">
              {{ reportData.report_type }}
            </span>
          </div>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-3 text-xs text-slate-600 dark:text-slate-400">
          <div>
            <span class="block text-3xs uppercase text-slate-400 font-semibold">{{ $t('reports.period') || 'Reporting Period' }}</span>
            <span class="font-medium text-slate-800 dark:text-slate-200">{{ reportData.period_formatted }}</span>
          </div>
          <div>
            <span class="block text-3xs uppercase text-slate-400 font-semibold">{{ $t('reports.generatedAt') || 'Generated At' }}</span>
            <span class="font-mono text-slate-800 dark:text-slate-200">{{ reportData.generated_at_formatted }}</span>
          </div>
          <div>
            <span class="block text-3xs uppercase text-slate-400 font-semibold">{{ $t('reports.serverAggregation') || 'Engine Mode' }}</span>
            <span class="text-emerald-600 dark:text-emerald-400 font-medium">Server-Side Aggregation &bull; High Integrity</span>
          </div>
        </div>
      </div>

      <!-- KPI Executive Cards Grid -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <!-- System Availability -->
        <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm">
          <div class="text-3xs font-semibold uppercase tracking-wider text-slate-500 mb-1">
            {{ $t('reports.kpiAvailability') || 'Fleet Availability' }}
          </div>
          <div class="text-2xl font-bold font-mono text-emerald-600 dark:text-emerald-400">
            {{ reportData.overall_availability.toFixed(1) }}%
          </div>
          <div class="text-3xs text-slate-500 mt-1 flex items-center justify-between">
            <span>{{ reportData.communication_summary.success_frames }} OK</span>
            <span class="text-rose-500">{{ reportData.communication_summary.error_frames }} ERR</span>
          </div>
        </div>

        <!-- Good Data Quality -->
        <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm">
          <div class="text-3xs font-semibold uppercase tracking-wider text-slate-500 mb-1">
            {{ $t('reports.kpiDataQuality') || 'Good Data Quality' }}
          </div>
          <div class="text-2xl font-bold font-mono text-blue-600 dark:text-blue-400">
            {{ reportData.overall_quality_good.toFixed(1) }}%
          </div>
          <div class="text-3xs text-slate-500 mt-1">
            Total {{ reportData.total_samples.toLocaleString() }} samples
          </div>
        </div>

        <!-- Alarms Count -->
        <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm">
          <div class="text-3xs font-semibold uppercase tracking-wider text-slate-500 mb-1">
            {{ $t('reports.kpiAlarms') || 'Active / Total Alarms' }}
          </div>
          <div class="text-2xl font-bold font-mono" :class="reportData.alarm_summary.active_count > 0 ? 'text-rose-500' : 'text-emerald-500'">
            {{ reportData.alarm_summary.active_count }} / {{ reportData.alarm_summary.total_alarms }}
          </div>
          <div class="text-3xs text-slate-500 mt-1">
            {{ reportData.alarm_summary.critical_count }} Crit &bull; {{ reportData.alarm_summary.warning_count }} Warn
          </div>
        </div>

        <!-- Fleet Scale -->
        <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm">
          <div class="text-3xs font-semibold uppercase tracking-wider text-slate-500 mb-1">
            {{ $t('reports.kpiFleet') || 'Active Equipment' }}
          </div>
          <div class="text-2xl font-bold font-mono text-slate-800 dark:text-slate-100">
            {{ reportData.total_devices }} Devices
          </div>
          <div class="text-3xs text-slate-500 mt-1 font-mono">
            {{ reportData.total_parameters }} Parameters &bull; {{ reportData.communication_summary.avg_latency_ms.toFixed(1) }}ms
          </div>
        </div>
      </div>

      <!-- SECTION 1: Device Fleet Availability -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-blue-500"></span>
            <h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
              {{ $t('reports.deviceAvailabilityTitle') || 'Device Fleet Availability & Communication Audit' }}
            </h3>
          </div>
          <span class="text-3xs font-mono text-slate-500">{{ reportData.device_summaries.length }} nodes</span>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-4 py-2.5">Device Code</th>
                <th class="px-4 py-2.5">Name</th>
                <th class="px-4 py-2.5">Status</th>
                <th class="px-4 py-2.5">Latency</th>
                <th class="px-4 py-2.5">Packets (OK / ERR)</th>
                <th class="px-4 py-2.5">Parameters</th>
                <th class="px-4 py-2.5 text-right">Availability</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-2xs">
              <tr v-for="d in reportData.device_summaries" :key="d.device_id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30">
                <td class="px-4 py-3 font-bold text-slate-900 dark:text-white">{{ d.device_code }}</td>
                <td class="px-4 py-3 font-sans text-xs text-slate-700 dark:text-slate-300">{{ d.device_name }}</td>
                <td class="px-4 py-3">
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-3xs font-bold uppercase"
                        :class="d.status === 'ONLINE' ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-500 border border-rose-500/20'">
                    <span class="w-1.5 h-1.5 rounded-full" :class="d.status === 'ONLINE' ? 'bg-emerald-500' : 'bg-rose-500'"></span>
                    {{ d.status }}
                  </span>
                </td>
                <td class="px-4 py-3 text-slate-500">{{ d.latency_ms }} ms</td>
                <td class="px-4 py-3">{{ d.success_count }} / <span class="text-rose-500">{{ d.failed_count }}</span></td>
                <td class="px-4 py-3 text-slate-500">{{ d.parameters_count }} active</td>
                <td class="px-4 py-3 text-right font-bold text-emerald-500">{{ d.availability_percent.toFixed(1) }}%</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- SECTION 2: Parameter Telemetry & Statistical Breakdown -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm page-break">
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
            <h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
              {{ $t('reports.telemetryStatisticsTitle') || 'Parameter Telemetry & Statistical Quality Metrics' }}
            </h3>
          </div>
          <span class="text-3xs font-mono text-slate-500">{{ reportData.parameter_summaries.length }} parameters</span>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-4 py-2.5">Device</th>
                <th class="px-4 py-2.5">Parameter</th>
                <th class="px-4 py-2.5">Unit</th>
                <th class="px-4 py-2.5">Samples</th>
                <th class="px-4 py-2.5">Min</th>
                <th class="px-4 py-2.5">Max</th>
                <th class="px-4 py-2.5">Average</th>
                <th class="px-4 py-2.5">Quality (Good / Bad / Unc)</th>
                <th class="px-4 py-2.5 text-right">Data Source</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-2xs">
              <tr v-for="p in reportData.parameter_summaries" :key="p.parameter_id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30">
                <td class="px-4 py-3 text-slate-500">{{ p.device_code }}</td>
                <td class="px-4 py-3 font-sans">
                  <span class="font-bold text-slate-900 dark:text-white">{{ p.parameter_name }}</span>
                  <span class="text-3xs font-mono text-slate-400 block">{{ p.parameter_code }}</span>
                </td>
                <td class="px-4 py-3 text-slate-500">{{ p.unit || '-' }}</td>
                <td class="px-4 py-3">{{ p.sample_count.toLocaleString() }}</td>
                <td class="px-4 py-3">{{ p.min_value !== null ? p.min_value.toFixed(2) : '-' }}</td>
                <td class="px-4 py-3">{{ p.max_value !== null ? p.max_value.toFixed(2) : '-' }}</td>
                <td class="px-4 py-3 font-bold text-blue-600 dark:text-blue-400">{{ p.avg_value !== null ? p.avg_value.toFixed(2) : '-' }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-1.5">
                    <span class="text-emerald-500 font-bold">{{ p.good_percent.toFixed(0) }}%</span>
                    <span class="text-slate-400">/</span>
                    <span class="text-rose-500">{{ p.bad_percent.toFixed(0) }}%</span>
                    <span class="text-slate-400">/</span>
                    <span class="text-amber-500">{{ p.uncertain_percent.toFixed(0) }}%</span>
                  </div>
                </td>
                <td class="px-4 py-3 text-right">
                  <span class="inline-block px-2 py-0.5 rounded text-3xs font-bold uppercase tracking-wider"
                        :class="p.data_source === 'RAW_TELEMETRY' ? 'bg-blue-500/10 text-blue-500' : p.data_source === 'PROCESSED' ? 'bg-emerald-500/10 text-emerald-500' : 'bg-slate-500/10 text-slate-400'">
                    {{ p.data_source === 'RAW_TELEMETRY' ? 'RAW' : p.data_source }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- SECTION 3: Alarms & Events Audit Log -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full bg-rose-500"></span>
            <h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
              {{ $t('reports.alarmAuditTitle') || 'Operational Alarms & Exception Log' }}
            </h3>
          </div>
          <span class="text-3xs font-mono text-slate-500">{{ reportData.alarm_summary.total_alarms }} recorded</span>
        </div>

        <div v-if="reportData.alarm_summary.recent_alarms.length === 0" class="p-6 text-center text-xs text-slate-500 dark:text-slate-400 font-sans">
          {{ $t('reports.noAlarmsRecorded') || 'No critical or warning alarms occurred during this reporting interval.' }}
        </div>

        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-4 py-2.5">Code</th>
                <th class="px-4 py-2.5">Device</th>
                <th class="px-4 py-2.5">Severity</th>
                <th class="px-4 py-2.5">Status</th>
                <th class="px-4 py-2.5">Message</th>
                <th class="px-4 py-2.5 text-right">Triggered Time</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-2xs">
              <tr v-for="a in reportData.alarm_summary.recent_alarms" :key="a.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30">
                <td class="px-4 py-3 font-bold text-slate-900 dark:text-white">{{ a.alarm_code }}</td>
                <td class="px-4 py-3 font-sans text-slate-700 dark:text-slate-300">{{ a.device_name }}</td>
                <td class="px-4 py-3">
                  <span class="inline-block px-2 py-0.5 rounded text-3xs font-bold uppercase tracking-wider"
                        :class="a.severity === 'CRITICAL' ? 'bg-rose-500/10 text-rose-500 border border-rose-500/20' : 'bg-amber-500/10 text-amber-500 border border-amber-500/20'">
                    {{ a.severity }}
                  </span>
                </td>
                <td class="px-4 py-3 text-slate-500">{{ a.status }}</td>
                <td class="px-4 py-3 font-sans text-xs text-slate-800 dark:text-slate-200 max-w-xs truncate">{{ a.message }}</td>
                <td class="px-4 py-3 text-right text-slate-500">{{ formatTimestamp(a.triggered_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'ReportsView',
  data() {
    return {
      loading: false,
      error: null,
      selectedPreset: '24h',
      presets: [
        { id: '1h', label: 'Last 1 Hour' },
        { id: '24h', label: 'Last 24 Hours' },
        { id: '7d', label: 'Last 7 Days' },
        { id: '30d', label: 'Last 30 Days' },
        { id: 'custom', label: 'Custom' },
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
  mounted() {
    this.initDefaultDates();
    this.fetchDevices();
    this.fetchReportPreview();
  },
  methods: {
    initDefaultDates() {
      const now = new Date();
      const past24h = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      this.filters.endTime = this.toLocalISOString(now);
      this.filters.startTime = this.toLocalISOString(past24h);
    },
    toLocalISOString(date) {
      const pad = n => (n < 10 ? '0' + n : n);
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
    exportReport(format) {
      const params = this.buildQueryParams();
      params.append('format', format);
      if (format === 'html') {
        params.append('download', '1');
      }
      window.open(`/api/reports/generate?${params.toString()}`, '_blank');
    },
    printReport() {
      window.print();
    },
    formatTimestamp(ts) {
      if (!ts) return '-';
      const d = new Date(ts);
      return d.toLocaleString('en-GB', { hour12: false });
    },
  },
};
</script>

<style scoped>
@media print {
  .no-print {
    display: none !important;
  }
  .page-break {
    page-break-before: always;
  }
  body, :host {
    background: #ffffff !important;
    color: #000000 !important;
  }
}
</style>
