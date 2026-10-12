<template>
  <div class="space-y-4 font-sans text-slate-800 dark:text-slate-200">
    <!-- Header with Title, Tab Switcher & Live Refresh Status -->
    <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center space-x-3">
        <div class="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-600 dark:text-blue-400 flex items-center justify-center flex-shrink-0">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <polyline points="4 17 10 11 4 5"></polyline>
            <line x1="12" y1="19" x2="20" y2="19"></line>
          </svg>
        </div>
        <div>
          <h1 class="text-base font-bold text-slate-900 dark:text-white tracking-tight flex items-center gap-2">
            {{ $t('logs.title') || 'Logs & Observability' }}
            <span v-if="refreshInterval > 0" class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-ping"></span>
              {{ $t('logs.liveStreaming') || 'LIVE' }} ({{ refreshInterval }}s)
            </span>
          </h1>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            {{ $t('logs.subtitle') || 'System events, operator audit trails, and hardware communication frames' }}
          </p>
        </div>
      </div>

      <!-- Tab Switcher -->
      <div class="flex flex-wrap items-center gap-2">
        <div class="flex items-center p-1 bg-slate-100 dark:bg-[#0B0F19] rounded-xl border border-slate-200 dark:border-slate-800 text-xs font-medium">
          <button id="tab-system-logs"
                  @click="switchTab('logs')"
                  class="px-3 py-1.5 rounded-lg transition-all"
                  :class="activeTab === 'logs' ? 'bg-white dark:bg-slate-800 text-blue-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('logs.systemLogs') || 'System Logs' }}
          </button>
          <button id="tab-audit-trails"
                  @click="switchTab('audit')"
                  class="px-3 py-1.5 rounded-lg transition-all"
                  :class="activeTab === 'audit' ? 'bg-white dark:bg-slate-800 text-blue-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('logs.auditTrails') || 'Audit Trails' }}
          </button>
          <button id="tab-communication-logs"
                  @click="switchTab('comm')"
                  class="px-3 py-1.5 rounded-lg transition-all"
                  :class="activeTab === 'comm' ? 'bg-white dark:bg-slate-800 text-blue-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('logs.communicationLogs') || 'Communication Logs' }}
          </button>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- TAB 1: SYSTEM LOGS                                             -->
    <!-- ============================================================== -->
    <div v-if="activeTab === 'logs'" class="space-y-4">
      <!-- Filter Bar -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm space-y-3">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-3 text-xs">
          <!-- Log Level Filter -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.level') || 'Level' }}</label>
            <select v-model="systemFilters.level" @change="fetchLogs(1)"
                    class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white">
              <option value="">{{ $t('logs.allLevels') || 'All Levels' }}</option>
              <option value="ERROR">ERROR</option>
              <option value="WARN">WARN</option>
              <option value="INFO">INFO</option>
              <option value="DEBUG">DEBUG</option>
            </select>
          </div>

          <!-- Component Filter -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.component') || 'Component' }}</label>
            <input type="text" v-model="systemFilters.component" @input="debounceFetchLogs"
                   :placeholder="$t('logs.allComponents') || 'All components'"
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
          </div>

          <!-- Start Date -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.startDate') || 'Start Date' }}</label>
            <input type="date" v-model="systemFilters.startDate" @change="fetchLogs(1)"
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono text-2xs" />
          </div>

          <!-- End Date -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.endDate') || 'End Date' }}</label>
            <input type="date" v-model="systemFilters.endDate" @change="fetchLogs(1)"
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono text-2xs" />
          </div>

          <!-- Text Search (Spans 2 columns on lg) -->
          <div class="sm:col-span-2">
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('common.search') || 'Search' }}</label>
            <div class="relative">
              <input type="text" v-model="systemFilters.search" @input="debounceFetchLogs"
                     :placeholder="$t('logs.searchLogs') || 'Search logs by message or component...'"
                     class="w-full px-3 py-1.5 pl-8 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
              <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </div>
        </div>

        <!-- Controls: Refresh Interval, Export & Refresh -->
        <div class="flex flex-wrap items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800/80 text-xs">
          <div class="flex items-center space-x-2">
            <span class="text-3xs text-slate-500 font-semibold uppercase">{{ $t('logs.liveStreaming') || 'Live Refresh' }}:</span>
            <div class="inline-flex rounded-lg border border-slate-200 dark:border-slate-800 p-0.5 bg-slate-50 dark:bg-slate-900 text-3xs font-mono">
              <button @click="setRefreshInterval(0)" class="px-2 py-0.5 rounded" :class="refreshInterval === 0 ? 'bg-indigo-600 text-white font-bold' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'">
                {{ $t('logs.paused') || 'Paused' }}
              </button>
              <button @click="setRefreshInterval(3)" class="px-2 py-0.5 rounded" :class="refreshInterval === 3 ? 'bg-indigo-600 text-white font-bold' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'">
                3s
              </button>
              <button @click="setRefreshInterval(5)" class="px-2 py-0.5 rounded" :class="refreshInterval === 5 ? 'bg-indigo-600 text-white font-bold' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'">
                5s
              </button>
              <button @click="setRefreshInterval(10)" class="px-2 py-0.5 rounded" :class="refreshInterval === 10 ? 'bg-indigo-600 text-white font-bold' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'">
                10s
              </button>
            </div>
          </div>

          <div class="flex items-center space-x-2 mt-2 sm:mt-0">
            <button @click="exportCSV('logs')" class="px-3 py-1 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 rounded-lg text-2xs font-semibold flex items-center gap-1 transition-colors">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
              {{ $t('logs.exportCSV') || 'Export CSV' }}
            </button>
            <button @click="fetchLogs(systemPage)" class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-2xs font-semibold flex items-center gap-1 transition-colors shadow-xs">
              <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loadingLogs }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
              {{ $t('common.refresh') || 'Refresh' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Logs Table -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse font-sans">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider font-sans">
                <th class="py-3 px-4 w-40">{{ $t('common.timestamp') || 'Timestamp' }}</th>
                <th class="py-3 px-4 w-24">{{ $t('logs.level') || 'Level' }}</th>
                <th class="py-3 px-4 w-32">{{ $t('logs.component') || 'Component' }}</th>
                <th class="py-3 px-4">{{ $t('logs.message') || 'Message' }}</th>
                <th class="py-3 px-4 w-1/3">{{ $t('logs.details') || 'Details' }}</th>
                <th class="py-3 px-4 w-16 text-center">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 text-slate-700 dark:text-slate-300 font-mono text-2xs">
              <!-- Loading State -->
              <tr v-if="loadingLogs">
                <td colspan="6" class="py-12 text-center text-slate-500 font-sans">
                  <div class="flex items-center justify-center space-x-2">
                    <span class="animate-spin w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full"></span>
                    <span>Loading logs...</span>
                  </div>
                </td>
              </tr>
              <!-- Error State -->
              <tr v-else-if="logsError">
                <td colspan="6" class="py-12 text-center text-rose-500 font-sans">
                  <div class="flex flex-col items-center justify-center space-y-2">
                    <svg class="w-8 h-8 text-rose-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                    <span>{{ logsError }}</span>
                    <button @click="fetchLogs(1)" class="px-3 py-1 bg-rose-500/10 text-rose-600 dark:text-rose-400 rounded-lg text-xs font-semibold hover:bg-rose-500/20">Try again</button>
                  </div>
                </td>
              </tr>
              <!-- Empty State -->
              <tr v-else-if="systemLogs.length === 0">
                <td colspan="6" class="py-12 text-center text-slate-500 font-sans">
                  <div class="flex flex-col items-center justify-center space-y-2">
                    <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                    <span>{{ $t('logs.noLogsFound') || 'No system logs match the current criteria.' }}</span>
                  </div>
                </td>
              </tr>
              <!-- Data Rows -->
              <tr v-for="log in systemLogs" :key="log.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/25 transition-colors">
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 text-3xs">{{ log.created_at | formatDate }}</td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded text-3xs font-bold tracking-wider font-mono" :class="getLevelBadgeClass(log.level)">
                    {{ log.level }}
                  </span>
                </td>
                <td class="py-3 px-4 text-slate-700 dark:text-slate-200 font-bold font-mono">{{ log.component }}</td>
                <td class="py-3 px-4 font-sans text-xs text-slate-900 dark:text-slate-100">
                  <div class="line-clamp-2">{{ log.message }}</div>
                </td>
                <td class="py-3 px-4 font-mono text-3xs text-slate-500 dark:text-slate-400">
                  <div class="line-clamp-2 truncate max-w-md">{{ log.details || '-' }}</div>
                </td>
                <td class="py-3 px-4 text-center">
                  <button @click="openInspectorModal('log', log)" class="p-1 text-slate-400 hover:text-blue-500 transition-colors" title="Inspect log">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-sans">
          <div>
            {{ $t('logs.page') || 'Page' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ systemPage }}</span> {{ $t('logs.of') || 'of' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ systemTotalPages }}</span>
            <span class="ml-2 font-mono text-3xs">({{ systemTotal }} {{ $t('logs.totalRecords') || 'total records' }})</span>
          </div>
          <div class="flex items-center space-x-1">
            <button @click="fetchLogs(systemPage - 1)" :disabled="systemPage <= 1"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 disabled:opacity-40 transition-colors text-2xs">
              {{ $t('common.prev') || 'Prev' }}
            </button>
            <button @click="fetchLogs(systemPage + 1)" :disabled="systemPage >= systemTotalPages"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 disabled:opacity-40 transition-colors text-2xs">
              {{ $t('common.next') || 'Next' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- TAB 2: AUDIT TRAILS                                            -->
    <!-- ============================================================== -->
    <div v-else-if="activeTab === 'audit'" class="space-y-4">
      <!-- Filter Bar -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm space-y-3">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-3 text-xs">
          <!-- Actor Filter -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.actor') || 'Actor' }}</label>
            <input type="text" v-model="auditFilters.username" @input="debounceFetchAudit"
                   :placeholder="$t('logs.allActors') || 'All actors'"
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
          </div>

          <!-- Action Filter -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.action') || 'Action' }}</label>
            <input type="text" v-model="auditFilters.action" @input="debounceFetchAudit"
                   :placeholder="$t('logs.allActions') || 'All actions'"
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
          </div>

          <!-- Resource Filter -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.resource') || 'Resource' }}</label>
            <input type="text" v-model="auditFilters.resource" @input="debounceFetchAudit"
                   :placeholder="$t('logs.allResources') || 'All resources'"
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
          </div>

          <!-- Start Date -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.startDate') || 'Start Date' }}</label>
            <input type="date" v-model="auditFilters.startDate" @change="fetchAudit(1)"
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono text-2xs" />
          </div>

          <!-- End Date -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.endDate') || 'End Date' }}</label>
            <input type="date" v-model="auditFilters.endDate" @change="fetchAudit(1)"
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono text-2xs" />
          </div>

          <!-- Text Search -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('common.search') || 'Search' }}</label>
            <div class="relative">
              <input type="text" v-model="auditFilters.search" @input="debounceFetchAudit"
                     :placeholder="$t('logs.searchAudit') || 'Search details...'"
                     class="w-full px-3 py-1.5 pl-8 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
              <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </div>
        </div>

        <!-- Controls: Export & Refresh -->
        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-100 dark:border-slate-800/80 text-xs">
          <button @click="exportCSV('audit')" class="px-3 py-1 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 rounded-lg text-2xs font-semibold flex items-center gap-1 transition-colors">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
            {{ $t('logs.exportCSV') || 'Export CSV' }}
          </button>
          <button @click="fetchAudit(auditPage)" class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-2xs font-semibold flex items-center gap-1 transition-colors shadow-xs">
            <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loadingAudit }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
            {{ $t('common.refresh') || 'Refresh' }}
          </button>
        </div>
      </div>

      <!-- Audit Table -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse font-sans">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider font-sans">
                <th class="py-3 px-4 w-40">{{ $t('common.timestamp') || 'Timestamp' }}</th>
                <th class="py-3 px-4 w-32">{{ $t('common.user') || 'Actor' }}</th>
                <th class="py-3 px-4 w-36">{{ $t('common.action') || 'Action' }}</th>
                <th class="py-3 px-4 w-28">{{ $t('logs.resource') || 'Resource' }}</th>
                <th class="py-3 px-4">{{ $t('logs.details') || 'Details' }}</th>
                <th class="py-3 px-4 w-28">{{ $t('logs.ip') || 'IP' }}</th>
                <th class="py-3 px-4 w-16 text-center">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 text-slate-700 dark:text-slate-300 font-mono text-2xs">
              <!-- Loading -->
              <tr v-if="loadingAudit">
                <td colspan="7" class="py-12 text-center text-slate-500 font-sans">
                  <div class="flex items-center justify-center space-x-2">
                    <span class="animate-spin w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full"></span>
                    <span>Loading audit records...</span>
                  </div>
                </td>
              </tr>
              <!-- Error -->
              <tr v-else-if="auditError">
                <td colspan="7" class="py-12 text-center text-rose-500 font-sans">
                  <div class="flex flex-col items-center justify-center space-y-2">
                    <svg class="w-8 h-8 text-rose-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                    <span>{{ auditError }}</span>
                    <button @click="fetchAudit(1)" class="px-3 py-1 bg-rose-500/10 text-rose-600 dark:text-rose-400 rounded-lg text-xs font-semibold hover:bg-rose-500/20">Try again</button>
                  </div>
                </td>
              </tr>
              <!-- Empty -->
              <tr v-else-if="auditTrails.length === 0">
                <td colspan="7" class="py-12 text-center text-slate-500 font-sans">
                  <div class="flex flex-col items-center justify-center space-y-2">
                    <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                    <span>{{ $t('logs.noAuditFound') || 'No audit trails match the current criteria.' }}</span>
                  </div>
                </td>
              </tr>
              <!-- Rows -->
              <tr v-for="trail in auditTrails" :key="trail.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/25 transition-colors">
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 text-3xs">{{ trail.created_at | formatDate }}</td>
                <td class="py-3 px-4">
                  <div class="flex items-center space-x-1.5">
                    <div class="w-5 h-5 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-600 dark:text-blue-400 font-bold flex items-center justify-center text-4xs">
                      {{ (trail.username || 'U').charAt(0).toUpperCase() }}
                    </div>
                    <span class="font-bold text-slate-900 dark:text-slate-100 font-mono">{{ trail.username }}</span>
                  </div>
                </td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded text-3xs font-semibold tracking-wider font-mono" :class="getActionBadgeClass(trail.action)">
                    {{ trail.action }}
                  </span>
                </td>
                <td class="py-3 px-4 text-indigo-600 dark:text-indigo-400 font-bold font-mono">{{ trail.resource }}</td>
                <td class="py-3 px-4 font-sans text-xs text-slate-800 dark:text-slate-200">
                  <div class="line-clamp-2 truncate max-w-md">{{ trail.details }}</div>
                </td>
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 text-3xs">{{ trail.ip_address || '-' }}</td>
                <td class="py-3 px-4 text-center">
                  <button @click="openInspectorModal('audit', trail)" class="p-1 text-slate-400 hover:text-blue-500 transition-colors" title="Inspect audit entry">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-sans">
          <div>
            {{ $t('logs.page') || 'Page' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ auditPage }}</span> {{ $t('logs.of') || 'of' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ auditTotalPages }}</span>
            <span class="ml-2 font-mono text-3xs">({{ auditTotal }} {{ $t('logs.totalRecords') || 'total records' }})</span>
          </div>
          <div class="flex items-center space-x-1">
            <button @click="fetchAudit(auditPage - 1)" :disabled="auditPage <= 1"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 disabled:opacity-40 transition-colors text-2xs">
              {{ $t('common.prev') || 'Prev' }}
            </button>
            <button @click="fetchAudit(auditPage + 1)" :disabled="auditPage >= auditTotalPages"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 disabled:opacity-40 transition-colors text-2xs">
              {{ $t('common.next') || 'Next' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- TAB 3: COMMUNICATION LOGS                                      -->
    <!-- ============================================================== -->
    <div v-else-if="activeTab === 'comm'" class="space-y-4">
      <!-- Filter Bar -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm space-y-3">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-3 text-xs">
          <!-- Device Selector -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.device') || 'Device' }}</label>
            <select v-model="commFilters.deviceId" @change="fetchComm(1)"
                    class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white">
              <option value="">{{ $t('logs.allDevices') || 'All Devices' }}</option>
              <option v-for="dev in devices" :key="dev.id" :value="dev.id">
                {{ dev.device_code || dev.name }} (ID: {{ dev.id }})
              </option>
            </select>
          </div>

          <!-- Direction Filter -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.direction') || 'Direction' }}</label>
            <select v-model="commFilters.direction" @change="fetchComm(1)"
                    class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white">
              <option value="">{{ $t('logs.allDirections') || 'All Directions' }}</option>
              <option value="TX">TX (Outgoing)</option>
              <option value="RX">RX (Incoming)</option>
            </select>
          </div>

          <!-- Status Filter -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.status') || 'Status' }}</label>
            <select v-model="commFilters.status" @change="fetchComm(1)"
                    class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white">
              <option value="">{{ $t('logs.allStatuses') || 'All Statuses' }}</option>
              <option value="SUCCESS">SUCCESS</option>
              <option value="TIMEOUT">TIMEOUT</option>
              <option value="ERROR">ERROR</option>
            </select>
          </div>

          <!-- Protocol Filter -->
          <div>
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('logs.protocol') || 'Protocol' }}</label>
            <input type="text" v-model="commFilters.protocol" @input="debounceFetchComm"
                   placeholder="Modbus-RTU, TCP..."
                   class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
          </div>

          <!-- Search (Hex or Error) -->
          <div class="sm:col-span-2">
            <label class="block text-3xs font-semibold text-slate-500 uppercase tracking-wider mb-1">{{ $t('common.search') || 'Search' }}</label>
            <div class="relative">
              <input type="text" v-model="commFilters.search" @input="debounceFetchComm"
                     :placeholder="$t('logs.searchComm') || 'Search hex frame or error message...'"
                     class="w-full px-3 py-1.5 pl-8 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
              <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
          </div>
        </div>

        <!-- Controls: Export & Refresh -->
        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-100 dark:border-slate-800/80 text-xs">
          <button @click="exportCSV('comm')" class="px-3 py-1 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 rounded-lg text-2xs font-semibold flex items-center gap-1 transition-colors">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" /></svg>
            {{ $t('logs.exportCSV') || 'Export CSV' }}
          </button>
          <button @click="fetchComm(commPage)" class="px-3 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-2xs font-semibold flex items-center gap-1 transition-colors shadow-xs">
            <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loadingComm }" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
            {{ $t('common.refresh') || 'Refresh' }}
          </button>
        </div>
      </div>

      <!-- Comm Logs Table -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse font-sans">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider font-sans">
                <th class="py-3 px-4 w-40">{{ $t('common.timestamp') || 'Timestamp' }}</th>
                <th class="py-3 px-4 w-28">{{ $t('logs.device') || 'Device' }}</th>
                <th class="py-3 px-4 w-28">{{ $t('logs.protocol') || 'Protocol' }}</th>
                <th class="py-3 px-4 w-24">{{ $t('logs.direction') || 'Direction' }}</th>
                <th class="py-3 px-4 w-28">{{ $t('logs.status') || 'Status' }}</th>
                <th class="py-3 px-4 w-20 text-center">{{ $t('logs.latency') || 'Latency' }}</th>
                <th class="py-3 px-4">{{ $t('logs.details') || 'Hex Payload / Error' }}</th>
                <th class="py-3 px-4 w-20 text-center">Inspect</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 text-slate-700 dark:text-slate-300 font-mono text-2xs">
              <!-- Loading -->
              <tr v-if="loadingComm">
                <td colspan="8" class="py-12 text-center text-slate-500 font-sans">
                  <div class="flex items-center justify-center space-x-2">
                    <span class="animate-spin w-4 h-4 border-2 border-blue-600 border-t-transparent rounded-full"></span>
                    <span>Loading communication frames...</span>
                  </div>
                </td>
              </tr>
              <!-- Error -->
              <tr v-else-if="commError">
                <td colspan="8" class="py-12 text-center text-rose-500 font-sans">
                  <div class="flex flex-col items-center justify-center space-y-2">
                    <svg class="w-8 h-8 text-rose-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
                    <span>{{ commError }}</span>
                    <button @click="fetchComm(1)" class="px-3 py-1 bg-rose-500/10 text-rose-600 dark:text-rose-400 rounded-lg text-xs font-semibold hover:bg-rose-500/20">Try again</button>
                  </div>
                </td>
              </tr>
              <!-- Empty -->
              <tr v-else-if="communicationLogs.length === 0">
                <td colspan="8" class="py-12 text-center text-slate-500 font-sans">
                  <div class="flex flex-col items-center justify-center space-y-2">
                    <svg class="w-8 h-8 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" /></svg>
                    <span>{{ $t('logs.noCommFound') || 'No communication logs match the current criteria.' }}</span>
                  </div>
                </td>
              </tr>
              <!-- Rows -->
              <tr v-for="cLog in communicationLogs" :key="cLog.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/25 transition-colors">
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 text-3xs">{{ cLog.created_at | formatDate }}</td>
                <td class="py-3 px-4">
                  <span class="text-blue-600 dark:text-blue-400 font-bold font-mono">Dev #{{ cLog.device_id }}</span>
                </td>
                <td class="py-3 px-4 text-slate-700 dark:text-slate-300 font-mono">{{ cLog.protocol }}</td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded text-3xs font-bold tracking-wider font-mono"
                        :class="cLog.direction === 'TX' ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20'">
                    {{ cLog.direction === 'TX' ? '▲ TX' : '▼ RX' }}
                  </span>
                </td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded text-3xs font-semibold font-mono" :class="getCommStatusBadgeClass(cLog.status)">
                    {{ cLog.status }}
                  </span>
                </td>
                <td class="py-3 px-4 text-center font-mono text-3xs" :class="getLatencyClass(cLog.latency_ms)">
                  {{ cLog.latency_ms ? cLog.latency_ms + ' ms' : '-' }}
                </td>
                <td class="py-3 px-4 font-mono text-3xs">
                  <div v-if="cLog.payload_hex" class="text-slate-800 dark:text-slate-200 truncate max-w-md">
                    {{ cLog.payload_hex }}
                  </div>
                  <div v-if="cLog.error_message" class="text-rose-500 truncate max-w-md font-sans">
                    Error: {{ cLog.error_message }}
                  </div>
                </td>
                <td class="py-3 px-4 text-center">
                  <button @click="openInspectorModal('comm', cLog)" class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 text-3xs font-semibold transition-colors">
                    Inspect
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-sans">
          <div>
            {{ $t('logs.page') || 'Page' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ commPage }}</span> {{ $t('logs.of') || 'of' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ commTotalPages }}</span>
            <span class="ml-2 font-mono text-3xs">({{ commTotal }} {{ $t('logs.totalRecords') || 'total records' }})</span>
          </div>
          <div class="flex items-center space-x-1">
            <button @click="fetchComm(commPage - 1)" :disabled="commPage <= 1"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 disabled:opacity-40 transition-colors text-2xs">
              {{ $t('common.prev') || 'Prev' }}
            </button>
            <button @click="fetchComm(commPage + 1)" :disabled="commPage >= commTotalPages"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 disabled:opacity-40 transition-colors text-2xs">
              {{ $t('common.next') || 'Next' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- MODAL: INSPECT LOG / AUDIT / COMM RECORD                       -->
    <!-- ============================================================== -->
    <div v-if="showModal && selectedItem" class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-xl overflow-hidden shadow-2xl p-6 space-y-4 font-sans text-xs">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <div class="flex items-center space-x-2">
            <div class="w-7 h-7 rounded-lg bg-blue-500/10 text-blue-600 dark:text-blue-400 flex items-center justify-center font-bold">
              🔍
            </div>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white">
              {{ modalTitle }}
            </h3>
          </div>
          <button @click="showModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white text-base font-bold">&times;</button>
        </div>

        <!-- Content Area -->
        <div class="space-y-3 font-sans">
          <!-- Metadata Table -->
          <div class="grid grid-cols-2 gap-2 p-3 bg-slate-50 dark:bg-slate-900/60 rounded-xl border border-slate-200/80 dark:border-slate-800 text-2xs">
            <div>
              <span class="text-slate-400 block text-3xs">ID:</span>
              <strong class="font-mono text-slate-800 dark:text-slate-200">#{{ selectedItem.id }}</strong>
            </div>
            <div>
              <span class="text-slate-400 block text-3xs">Timestamp:</span>
              <span class="font-mono text-slate-800 dark:text-slate-200">{{ selectedItem.created_at | formatDate }}</span>
            </div>

            <!-- Specific metadata for comm -->
            <template v-if="modalType === 'comm'">
              <div>
                <span class="text-slate-400 block text-3xs">Device:</span>
                <span class="font-mono font-bold text-blue-500">Device #{{ selectedItem.device_id }}</span>
              </div>
              <div>
                <span class="text-slate-400 block text-3xs">Latency:</span>
                <span class="font-mono font-bold">{{ selectedItem.latency_ms }} ms</span>
              </div>
              <div>
                <span class="text-slate-400 block text-3xs">Direction & Protocol:</span>
                <span class="font-mono">{{ selectedItem.direction }} / {{ selectedItem.protocol }}</span>
              </div>
              <div>
                <span class="text-slate-400 block text-3xs">Status:</span>
                <span class="font-mono font-bold">{{ selectedItem.status }}</span>
              </div>
            </template>

            <!-- Specific metadata for audit -->
            <template v-if="modalType === 'audit'">
              <div>
                <span class="text-slate-400 block text-3xs">Actor:</span>
                <span class="font-mono font-bold text-blue-500">{{ selectedItem.username }}</span>
              </div>
              <div>
                <span class="text-slate-400 block text-3xs">Action:</span>
                <span class="font-mono font-bold">{{ selectedItem.action }}</span>
              </div>
              <div>
                <span class="text-slate-400 block text-3xs">Resource:</span>
                <span class="font-mono">{{ selectedItem.resource }}</span>
              </div>
              <div>
                <span class="text-slate-400 block text-3xs">Client IP:</span>
                <span class="font-mono">{{ selectedItem.ip_address || '-' }}</span>
              </div>
            </template>

            <!-- Specific metadata for log -->
            <template v-if="modalType === 'log'">
              <div>
                <span class="text-slate-400 block text-3xs">Level:</span>
                <span class="font-mono font-bold">{{ selectedItem.level }}</span>
              </div>
              <div>
                <span class="text-slate-400 block text-3xs">Component:</span>
                <span class="font-mono font-bold">{{ selectedItem.component }}</span>
              </div>
            </template>
          </div>

          <!-- Hex payload / Raw details section -->
          <div v-if="modalType === 'comm' && selectedItem.payload_hex" class="space-y-1">
            <div class="flex items-center justify-between">
              <span class="text-3xs uppercase font-bold text-slate-500 tracking-wider">Hex Payload ({{ getHexByteCount(selectedItem.payload_hex) }} Bytes)</span>
              <button @click="copyToClipboard(selectedItem.payload_hex)" class="text-3xs text-blue-500 hover:underline flex items-center gap-1 font-mono">
                {{ isCopied ? '✓ Copied' : 'Copy Hex' }}
              </button>
            </div>
            <pre class="p-3 rounded-xl bg-slate-900 text-emerald-400 font-mono text-2xs overflow-x-auto whitespace-pre-wrap select-all border border-slate-800">{{ formatHexDump(selectedItem.payload_hex) }}</pre>
          </div>

          <!-- Message or Details -->
          <div v-if="selectedItem.message" class="space-y-1">
            <span class="text-3xs uppercase font-bold text-slate-500 tracking-wider">Message</span>
            <div class="p-3 rounded-xl bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-xs">
              {{ selectedItem.message }}
            </div>
          </div>

          <div v-if="selectedItem.details" class="space-y-1">
            <span class="text-3xs uppercase font-bold text-slate-500 tracking-wider">Full Details</span>
            <pre class="p-3 rounded-xl bg-slate-50 dark:bg-slate-900 text-slate-800 dark:text-slate-200 font-mono text-2xs overflow-x-auto whitespace-pre-wrap border border-slate-200 dark:border-slate-800">{{ selectedItem.details }}</pre>
          </div>

          <div v-if="selectedItem.error_message" class="space-y-1">
            <span class="text-3xs uppercase font-bold text-rose-500 tracking-wider">Error Diagnostics</span>
            <pre class="p-3 rounded-xl bg-rose-500/10 text-rose-500 dark:text-rose-400 font-mono text-2xs overflow-x-auto whitespace-pre-wrap border border-rose-500/20">{{ selectedItem.error_message }}</pre>
          </div>
        </div>

        <div class="flex items-center justify-end pt-3 border-t border-slate-200 dark:border-slate-800">
          <button @click="showModal = false" class="px-4 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 font-semibold text-xs transition-colors">
            {{ $t('logs.close') || 'Close' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'LogsView',
  data() {
    return {
      activeTab: 'logs',
      refreshInterval: 0,
      refreshTimer: null,

      // System Logs state
      systemLogs: [],
      systemTotal: 0,
      systemPage: 1,
      systemPageSize: 50,
      systemTotalPages: 1,
      loadingLogs: false,
      logsError: null,
      systemFilters: {
        level: '',
        component: '',
        search: '',
        startDate: '',
        endDate: '',
      },
      systemDebounceTimer: null,

      // Audit Trails state
      auditTrails: [],
      auditTotal: 0,
      auditPage: 1,
      auditPageSize: 50,
      auditTotalPages: 1,
      loadingAudit: false,
      auditError: null,
      auditFilters: {
        username: '',
        action: '',
        resource: '',
        search: '',
        startDate: '',
        endDate: '',
      },
      auditDebounceTimer: null,

      // Communication Logs state
      communicationLogs: [],
      commTotal: 0,
      commPage: 1,
      commPageSize: 50,
      commTotalPages: 1,
      loadingComm: false,
      commError: null,
      commFilters: {
        deviceId: '',
        protocol: '',
        direction: '',
        status: '',
        search: '',
      },
      commDebounceTimer: null,

      // Inspector Modal state
      showModal: false,
      modalType: 'log',
      selectedItem: null,
      isCopied: false,
    };
  },
  computed: {
    devices() {
      return (this.$store && this.$store.state.devices) || [];
    },
    modalTitle() {
      if (this.modalType === 'comm') return 'Communication Packet Diagnostics';
      if (this.modalType === 'audit') return 'Security Audit Trail Details';
      return 'System Log Details';
    },
  },
  watch: {
    '$route.path'(path) {
      this.syncTabFromRoute(path);
    },
    '$route.query.tab'(tab) {
      if (tab) {
        this.activeTab = tab;
      }
    },
  },
  created() {
    this.syncTabFromRoute(this.$route.path);
  },
  mounted() {
    this.fetchLogs(1);
    this.fetchAudit(1);
    this.fetchComm(1);
    if (this.$store && this.$store.dispatch) {
      this.$store.dispatch('fetchDevices');
    }
  },
  beforeDestroy() {
    this.clearRefreshTimer();
    if (this.systemDebounceTimer) clearTimeout(this.systemDebounceTimer);
    if (this.auditDebounceTimer) clearTimeout(this.auditDebounceTimer);
    if (this.commDebounceTimer) clearTimeout(this.commDebounceTimer);
  },
  methods: {
    syncTabFromRoute(path) {
      if (path.includes('/audit')) {
        this.activeTab = 'audit';
      } else if (path.includes('/communication')) {
        this.activeTab = 'comm';
      } else {
        this.activeTab = 'logs';
      }
    },
    switchTab(tab) {
      this.activeTab = tab;
      if (tab === 'logs') {
        this.fetchLogs(this.systemPage);
      } else if (tab === 'audit') {
        this.fetchAudit(this.auditPage);
      } else if (tab === 'comm') {
        this.fetchComm(this.commPage);
      }
    },

    // Timer management
    setRefreshInterval(seconds) {
      this.refreshInterval = seconds;
      this.clearRefreshTimer();
      if (seconds > 0) {
        this.refreshTimer = setInterval(() => {
          if (this.activeTab === 'logs') {
            this.fetchLogs(this.systemPage, true);
          } else if (this.activeTab === 'comm') {
            this.fetchComm(this.commPage, true);
          }
        }, seconds * 1000);
      }
    },
    clearRefreshTimer() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer);
        this.refreshTimer = null;
      }
    },

    // Fetch System Logs
    async fetchLogs(page = 1, isBackground = false) {
      this.systemPage = page;
      if (!isBackground) this.loadingLogs = true;
      this.logsError = null;
      try {
        const params = {
          page: this.systemPage,
          page_size: this.systemPageSize,
          level: this.systemFilters.level || undefined,
          component: this.systemFilters.component || undefined,
          search: this.systemFilters.search || undefined,
          start_date: this.systemFilters.startDate || undefined,
          end_date: this.systemFilters.endDate || undefined,
        };
        const res = await axios.get('/api/logs', { params });
        if (res.data && res.data.data) {
          const d = res.data.data;
          if (Array.isArray(d)) {
            this.systemLogs = d;
            this.systemTotal = d.length;
            this.systemTotalPages = 1;
          } else if (d.items) {
            this.systemLogs = d.items;
            this.systemTotal = d.total || 0;
            this.systemTotalPages = d.total_pages || 1;
          }
        }
      } catch (err) {
        this.logsError = (err.response && err.response.data && err.response.data.message) || err.message || 'Failed to load system logs';
      } finally {
        this.loadingLogs = false;
      }
    },
    debounceFetchLogs() {
      if (this.systemDebounceTimer) clearTimeout(this.systemDebounceTimer);
      this.systemDebounceTimer = setTimeout(() => {
        this.fetchLogs(1);
      }, 400);
    },

    // Fetch Audit Trails
    async fetchAudit(page = 1) {
      this.auditPage = page;
      this.loadingAudit = true;
      this.auditError = null;
      try {
        const params = {
          page: this.auditPage,
          page_size: this.auditPageSize,
          username: this.auditFilters.username || undefined,
          action: this.auditFilters.action || undefined,
          resource: this.auditFilters.resource || undefined,
          search: this.auditFilters.search || undefined,
          start_date: this.auditFilters.startDate || undefined,
          end_date: this.auditFilters.endDate || undefined,
        };
        const res = await axios.get('/api/audit-trails', { params });
        if (res.data && res.data.data) {
          const d = res.data.data;
          if (Array.isArray(d)) {
            this.auditTrails = d;
            this.auditTotal = d.length;
            this.auditTotalPages = 1;
          } else if (d.items) {
            this.auditTrails = d.items;
            this.auditTotal = d.total || 0;
            this.auditTotalPages = d.total_pages || 1;
          }
        }
      } catch (err) {
        this.auditError = (err.response && err.response.data && err.response.data.message) || err.message || 'Failed to load audit trails';
      } finally {
        this.loadingAudit = false;
      }
    },
    debounceFetchAudit() {
      if (this.auditDebounceTimer) clearTimeout(this.auditDebounceTimer);
      this.auditDebounceTimer = setTimeout(() => {
        this.fetchAudit(1);
      }, 400);
    },

    // Fetch Communication Logs
    async fetchComm(page = 1, isBackground = false) {
      this.commPage = page;
      if (!isBackground) this.loadingComm = true;
      this.commError = null;
      try {
        const params = {
          page: this.commPage,
          page_size: this.commPageSize,
          device_id: this.commFilters.deviceId || undefined,
          protocol: this.commFilters.protocol || undefined,
          direction: this.commFilters.direction || undefined,
          status: this.commFilters.status || undefined,
          search: this.commFilters.search || undefined,
        };
        const res = await axios.get('/api/communication-logs', { params });
        if (res.data && res.data.data) {
          const d = res.data.data;
          if (Array.isArray(d)) {
            this.communicationLogs = d;
            this.commTotal = d.length;
            this.commTotalPages = 1;
          } else if (d.items) {
            this.communicationLogs = d.items;
            this.commTotal = d.total || 0;
            this.commTotalPages = d.total_pages || 1;
          }
        }
      } catch (err) {
        this.commError = (err.response && err.response.data && err.response.data.message) || err.message || 'Failed to load communication logs';
      } finally {
        this.loadingComm = false;
      }
    },
    debounceFetchComm() {
      if (this.commDebounceTimer) clearTimeout(this.commDebounceTimer);
      this.commDebounceTimer = setTimeout(() => {
        this.fetchComm(1);
      }, 400);
    },

    // Inspect Modal
    openInspectorModal(type, item) {
      this.modalType = type;
      this.selectedItem = item;
      this.isCopied = false;
      this.showModal = true;
    },
    copyToClipboard(text) {
      if (!text) return;
      if (navigator && navigator.clipboard) {
        navigator.clipboard.writeText(text);
        this.isCopied = true;
        setTimeout(() => {
          this.isCopied = false;
        }, 3000);
      }
    },
    getHexByteCount(hex) {
      if (!hex) return 0;
      const clean = hex.replace(/[^0-9A-Fa-f]/g, '');
      return Math.floor(clean.length / 2);
    },
    formatHexDump(hex) {
      if (!hex) return '';
      const clean = hex.replace(/[^0-9A-Fa-f]/g, '');
      const chunks = clean.match(/.{1,2}/g) || [];
      return chunks.join(' ').toUpperCase();
    },

    // Styling Helpers
    getLevelBadgeClass(level) {
      switch (level) {
        case 'ERROR':
          return 'bg-rose-500/10 text-rose-500 dark:text-rose-400 border border-rose-500/20';
        case 'WARN':
          return 'bg-amber-500/10 text-amber-500 border border-amber-500/20';
        case 'INFO':
          return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20';
        case 'DEBUG':
          return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20';
        default:
          return 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400';
      }
    },
    getActionBadgeClass(action) {
      if (action.includes('ACKNOWLEDGE')) {
        return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20';
      }
      if (action.includes('CLEAR')) {
        return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20';
      }
      if (action.includes('DELETE')) {
        return 'bg-rose-500/10 text-rose-500 border border-rose-500/20';
      }
      if (action.includes('CREATE') || action.includes('RESTORE')) {
        return 'bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20';
      }
      return 'bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300';
    },
    getCommStatusBadgeClass(status) {
      switch (status) {
        case 'SUCCESS':
          return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20';
        case 'TIMEOUT':
          return 'bg-amber-500/10 text-amber-500 border border-amber-500/20';
        case 'ERROR':
          return 'bg-rose-500/10 text-rose-500 border border-rose-500/20';
        default:
          return 'bg-slate-100 dark:bg-slate-800 text-slate-600';
      }
    },
    getLatencyClass(ms) {
      if (!ms) return 'text-slate-400';
      if (ms < 50) return 'text-emerald-500 font-bold';
      if (ms < 200) return 'text-amber-500 font-semibold';
      return 'text-rose-500 font-bold';
    },

    // CSV Exporter
    exportCSV(type) {
      let csvContent = '';
      let filename = '';
      const now = new Date().toISOString().replace(/[:.]/g, '-');

      if (type === 'logs') {
        filename = `system_logs_${now}.csv`;
        csvContent = 'ID,Timestamp,Level,Component,Message,Details\n';
        this.systemLogs.forEach((l) => {
          csvContent += `"${l.id}","${l.created_at}","${l.level}","${l.component}","${(l.message || '').replace(/"/g, '""')}","${(l.details || '').replace(/"/g, '""')}"\n`;
        });
      } else if (type === 'audit') {
        filename = `audit_trails_${now}.csv`;
        csvContent = 'ID,Timestamp,Username,Action,Resource,Details,IPAddress\n';
        this.auditTrails.forEach((a) => {
          csvContent += `"${a.id}","${a.created_at}","${a.username}","${a.action}","${a.resource}","${(a.details || '').replace(/"/g, '""')}","${a.ip_address || ''}"\n`;
        });
      } else if (type === 'comm') {
        filename = `communication_logs_${now}.csv`;
        csvContent = 'ID,Timestamp,DeviceID,Protocol,Direction,Status,LatencyMs,PayloadHex,ErrorMessage\n';
        this.communicationLogs.forEach((c) => {
          csvContent += `"${c.id}","${c.created_at}","${c.device_id}","${c.protocol}","${c.direction}","${c.status}","${c.latency_ms}","${c.payload_hex || ''}","${(c.error_message || '').replace(/"/g, '""')}"\n`;
        });
      }

      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const link = document.createElement('a');
      link.href = URL.createObjectURL(blob);
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    },
  },
};
</script>
