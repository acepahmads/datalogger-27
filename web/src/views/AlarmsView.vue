<template>
  <div class="space-y-6 font-sans">
    <!-- Header with Title, Status & Tabs -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-slate-900 dark:text-white font-sans flex items-center gap-2">
          <svg class="w-5 h-5 text-rose-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
          </svg>
          {{ $t('alarms.title') }}
        </h1>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5 font-sans">
          {{ $t('alarms.subtitle') }}
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-2">
        <!-- Active Alarms Counter Badge -->
        <span class="px-3 py-1.5 rounded-lg text-xs font-semibold font-mono"
              :class="activeCount > 0 ? 'bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20 font-bold' : 'bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/20 font-bold'">
          {{ $t('alarms.activeCount', { count: activeCount }) }}
        </span>

        <!-- View Switcher Tabs -->
        <div class="flex items-center space-x-1 bg-white dark:bg-[#111827] p-1 rounded-lg border border-slate-200 dark:border-slate-800 text-xs font-sans shadow-sm">
          <button id="tab-active-incidents"
                  @click="activeTab = 'alarms'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'alarms' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('alarms.tabActive') || 'Active Incidents' }}
          </button>
          <button id="tab-alarm-history"
                  @click="activeTab = 'history'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'history' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('alarms.tabHistory') || 'Alarm History' }}
          </button>
          <button id="tab-delivery-queue"
                  @click="activeTab = 'queue'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'queue' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('alarms.tabQueue') || 'Delivery Queue' }}
          </button>
          <button id="tab-notification-channels"
                  @click="activeTab = 'channels'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'channels' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('alarms.tabChannels') || 'Channels' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Alert / Toast Banner -->
    <div v-if="toastMessage"
         class="p-3.5 rounded-xl border flex items-center justify-between text-xs transition-all"
         :class="toastType === 'error' ? 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-500/10 dark:border-rose-500/30 dark:text-rose-400' : 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-500/10 dark:border-emerald-500/30 dark:text-emerald-400'">
      <div class="flex items-center space-x-2">
        <span>{{ toastMessage }}</span>
      </div>
      <button @click="toastMessage = null" class="text-xs opacity-75 hover:opacity-100 font-bold">&times;</button>
    </div>

    <!-- ========================================== -->
    <!-- TAB 1: ACTIVE INCIDENTS                    -->
    <!-- ========================================== -->
    <div v-if="activeTab === 'alarms'" class="space-y-4">
      <div class="saas-card overflow-hidden bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-sm">
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-rose-500 animate-pulse"></span>
            <h2 class="text-sm font-bold text-slate-900 dark:text-white font-sans tracking-tight">
              {{ $t('alarms.tabActive') || 'Active Incidents' }}
            </h2>
          </div>
          <button @click="refreshAlarms" class="text-2xs text-indigo-600 dark:text-indigo-400 hover:underline font-sans font-medium flex items-center gap-1">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
            {{ $t('common.refresh') }}
          </button>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse font-sans">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider font-sans">
                <th class="py-3 px-4 w-28">{{ $t('alarms.severity') }}</th>
                <th class="py-3 px-4 w-32">{{ $t('alarms.alarmCode') }}</th>
                <th class="py-3 px-4">{{ $t('alarms.message') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.device') }}</th>
                <th class="py-3 px-4 w-32">{{ $t('alarms.triggerValue') || 'Trigger / Limit' }}</th>
                <th class="py-3 px-4 w-40">{{ $t('alarms.triggeredAt') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.status') }}</th>
                <th class="py-3 px-4 w-36 text-center">{{ $t('alarms.action') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 text-slate-700 dark:text-slate-300">
              <tr v-if="activeAlarmsList.length === 0">
                <td colspan="8" class="py-12 text-center text-slate-500 font-sans">
                  <div class="flex flex-col items-center justify-center space-y-2">
                    <svg class="w-8 h-8 text-emerald-500/80" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    <span class="text-xs font-medium">{{ $t('alarms.noActiveAlarms') }}</span>
                  </div>
                </td>
              </tr>
              <tr v-for="alarm in activeAlarmsList" :key="alarm.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/25 transition-colors">
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-bold font-mono tracking-wider"
                        :class="getSeverityBadgeClass(alarm.severity)">
                    {{ $t('alarms.' + (alarm.severity || '').toLowerCase()) || alarm.severity }}
                  </span>
                </td>
                <td class="py-3 px-4 text-slate-900 dark:text-slate-200 font-semibold font-mono text-2xs">{{ alarm.alarm_code }}</td>
                <td class="py-3 px-4 text-xs text-slate-700 dark:text-slate-100 font-sans">
                  <div class="line-clamp-2">{{ alarm.message }}</div>
                  <div v-if="alarm.notes" class="text-3xs text-indigo-500 dark:text-indigo-400 mt-0.5 italic">
                    Note: {{ alarm.notes }}
                  </div>
                </td>
                <td class="py-3 px-4 text-indigo-600 dark:text-blue-400 font-mono text-2xs">
                  {{ alarm.device ? (alarm.device.device_code || alarm.device.code || alarm.device.name) : 'System' }}
                </td>
                <td class="py-3 px-4 font-mono text-2xs text-slate-600 dark:text-slate-400">
                  <span v-if="alarm.trigger_value != null" class="font-bold text-rose-500">{{ Number(alarm.trigger_value).toFixed(2) }}</span>
                  <span v-else class="text-slate-400">-</span>
                  <span v-if="alarm.threshold_value != null" class="text-3xs text-slate-400 block">Lim: {{ Number(alarm.threshold_value).toFixed(2) }}</span>
                </td>
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 font-mono text-3xs">{{ alarm.triggered_at | formatDate }}</td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-semibold font-mono"
                        :class="alarm.status === 'ACTIVE' ? 'bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/20 font-bold' : 'bg-amber-50 text-amber-700 border border-amber-200 dark:bg-amber-500/10 dark:text-amber-400 dark:border-amber-500/20'">
                    {{ alarm.status }}
                  </span>
                </td>
                <td class="py-3 px-4 text-center">
                  <div class="flex items-center justify-center space-x-1.5">
                    <button v-if="alarm.status === 'ACTIVE'"
                            id="btn-ack-alarm"
                            @click="openAckModal(alarm)"
                            class="px-2.5 py-1 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-2xs font-sans font-medium transition-colors shadow-xs">
                      {{ $t('alarms.acknowledge') }}
                    </button>
                    <span v-else class="text-slate-400 text-3xs font-mono">{{ $t('alarms.acked') }}</span>

                    <button @click="openIncidentModal(alarm)"
                            class="p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
                            :title="$t('alarms.details') || 'Details'">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- TAB 2: ALARM HISTORY (Task #66)            -->
    <!-- ========================================== -->
    <div v-if="activeTab === 'history'" class="space-y-4">
      <!-- Search & Multi-criteria Filters -->
      <div class="p-4 rounded-2xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-sm space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2 text-xs font-sans">
            <!-- Status Filter -->
            <select v-model="historyStatus" @change="fetchAlarmHistory"
                    class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 text-xs">
              <option value="">{{ $t('alarms.filterStatus') || 'All Statuses' }}</option>
              <option value="ACTIVE">ACTIVE</option>
              <option value="ACKNOWLEDGED">ACKNOWLEDGED</option>
              <option value="CLEARED">CLEARED</option>
            </select>

            <!-- Severity Filter -->
            <select v-model="historySeverity" @change="fetchAlarmHistory"
                    class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 text-xs">
              <option value="">{{ $t('alarms.filterSeverity') || 'All Severities' }}</option>
              <option value="CRITICAL">CRITICAL</option>
              <option value="WARNING">WARNING</option>
              <option value="INFO">INFO</option>
            </select>

            <!-- Device Filter -->
            <select v-model="historyDeviceID" @change="fetchAlarmHistory"
                    class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 text-xs">
              <option value="">{{ $t('alarms.filterDevice') || 'All Devices' }}</option>
              <option v-for="d in devices" :key="d.id" :value="d.id">
                {{ d.code || d.name }}
              </option>
            </select>

            <!-- Start Date -->
            <div class="flex items-center space-x-1">
              <span class="text-3xs text-slate-400">{{ $t('alarms.startDate') || 'From' }}:</span>
              <input type="date" v-model="historyStartDate" @change="fetchAlarmHistory"
                     class="px-2 py-1 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 text-xs font-mono" />
            </div>

            <!-- End Date -->
            <div class="flex items-center space-x-1">
              <span class="text-3xs text-slate-400">{{ $t('alarms.endDate') || 'To' }}:</span>
              <input type="date" v-model="historyEndDate" @change="fetchAlarmHistory"
                     class="px-2 py-1 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 text-xs font-mono" />
            </div>
          </div>

          <!-- Reset Filter & Refresh -->
          <div class="flex items-center space-x-2">
            <button @click="resetHistoryFilters"
                    class="px-2.5 py-1.5 text-2xs text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-800 rounded-lg">
              {{ $t('common.reset') || 'Reset' }}
            </button>
            <button @click="fetchAlarmHistory"
                    class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium transition-colors shadow-xs flex items-center gap-1">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" /></svg>
              {{ $t('common.refresh') || 'Refresh' }}
            </button>
          </div>
        </div>

        <!-- Search Input -->
        <div class="relative">
          <input type="text"
                 v-model="historySearch"
                 @input="debounceHistorySearch"
                 :placeholder="$t('alarms.searchHistoryPlaceholder') || 'Search code, message, notes, operator...'"
                 class="w-full px-3 py-2 pl-9 rounded-xl bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white text-xs placeholder:text-slate-400 focus:outline-none focus:border-indigo-500 font-sans" />
          <svg class="w-4 h-4 text-slate-400 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </div>

      <!-- History Table -->
      <div class="saas-card overflow-hidden bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse font-sans">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider font-sans">
                <th class="py-3 px-4 w-24">{{ $t('alarms.severity') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.status') }}</th>
                <th class="py-3 px-4 w-32">{{ $t('alarms.alarmCode') }}</th>
                <th class="py-3 px-4">{{ $t('alarms.message') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.device') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.duration') || 'Duration' }}</th>
                <th class="py-3 px-4 w-36">{{ $t('alarms.triggeredAt') }}</th>
                <th class="py-3 px-4 w-36">{{ $t('alarms.ackedBy') || 'Acknowledged' }}</th>
                <th class="py-3 px-4 w-36">{{ $t('alarms.clearedAt') || 'Cleared At' }}</th>
                <th class="py-3 px-4 w-20 text-center">{{ $t('alarms.details') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 text-slate-700 dark:text-slate-300">
              <tr v-if="historyLoading">
                <td colspan="10" class="py-12 text-center text-slate-500 font-sans">
                  <div class="flex items-center justify-center space-x-2">
                    <span class="animate-spin w-4 h-4 border-2 border-indigo-600 border-t-transparent rounded-full"></span>
                    <span>Loading history records...</span>
                  </div>
                </td>
              </tr>
              <tr v-else-if="historyItems.length === 0">
                <td colspan="10" class="py-12 text-center text-slate-500 font-sans">
                  {{ $t('alarms.noNotifications') || 'No alarm history records matching current criteria.' }}
                </td>
              </tr>
              <tr v-for="item in historyItems" :key="item.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/25 transition-colors">
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-bold font-mono tracking-wider"
                        :class="getSeverityBadgeClass(item.severity)">
                    {{ item.severity }}
                  </span>
                </td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-semibold font-mono"
                        :class="getStatusBadgeClass(item.status)">
                    {{ item.status }}
                  </span>
                </td>
                <td class="py-3 px-4 text-slate-900 dark:text-slate-200 font-semibold font-mono text-2xs">{{ item.alarm_code }}</td>
                <td class="py-3 px-4 text-xs text-slate-700 dark:text-slate-100 font-sans">
                  <div class="line-clamp-1">{{ item.message }}</div>
                  <div v-if="item.notes" class="text-3xs text-indigo-500 dark:text-indigo-400 italic">
                    Note: {{ item.notes }}
                  </div>
                </td>
                <td class="py-3 px-4 text-indigo-600 dark:text-blue-400 font-mono text-2xs">
                  {{ item.device_code || item.device_name || (item.device ? (item.device.code || item.device.name) : 'System') }}
                </td>
                <td class="py-3 px-4 font-mono text-2xs font-semibold"
                    :class="item.is_ongoing ? 'text-amber-700 dark:text-amber-400' : 'text-slate-600 dark:text-slate-400'">
                  {{ item.duration_formatted }}
                  <span v-if="item.is_ongoing" class="text-3xs block text-amber-700 dark:text-amber-400 font-bold uppercase">{{ $t('alarms.ongoing') || 'Ongoing' }}</span>
                </td>
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 font-mono text-3xs">{{ item.triggered_at | formatDate }}</td>
                <td class="py-3 px-4 text-3xs font-sans">
                  <div v-if="item.acknowledged_at">
                    <span class="font-semibold text-slate-800 dark:text-slate-200 block">{{ item.acknowledged_by_name || item.acknowledged_by }}</span>
                    <span class="text-slate-400 font-mono text-4xs">{{ item.acknowledged_at | formatDate }}</span>
                  </div>
                  <span v-else class="text-slate-400">-</span>
                </td>
                <td class="py-3 px-4 text-3xs font-mono">
                  <span v-if="item.cleared_at" class="text-emerald-700 dark:text-emerald-400 font-semibold">{{ item.cleared_at | formatDate }}</span>
                  <span v-else class="text-amber-700 dark:text-amber-400 font-sans font-bold">{{ $t('alarms.ongoing') || 'Active' }}</span>
                </td>
                <td class="py-3 px-4 text-center">
                  <button @click="openIncidentModal(item)"
                          class="p-1 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 transition-colors"
                          :title="$t('alarms.details') || 'Inspect Details'">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" /></svg>
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- History Pagination Footer -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400 font-sans">
          <div>
            {{ $t('alarms.page') || 'Page' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ historyPage }}</span> {{ $t('alarms.of') || 'of' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ historyTotalPages }}</span>
            <span class="ml-2 font-mono text-3xs">({{ historyTotal }} {{ $t('alarms.totalRecords') || 'total records' }})</span>
          </div>

          <div class="flex items-center space-x-1">
            <button @click="changeHistoryPage(historyPage - 1)"
                    :disabled="historyPage <= 1"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 disabled:opacity-40 transition-colors text-2xs font-medium">
              {{ $t('common.prev') || 'Prev' }}
            </button>
            <button @click="changeHistoryPage(historyPage + 1)"
                    :disabled="historyPage >= historyTotalPages"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 disabled:opacity-40 transition-colors text-2xs font-medium">
              {{ $t('common.next') || 'Next' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- TAB 3: NOTIFICATION DELIVERY QUEUE         -->
    <!-- ========================================== -->
    <div v-if="activeTab === 'queue'" class="space-y-4">
      <!-- Metrics Overview KPI Cards -->
      <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-semibold block">{{ $t('alarms.dispatched') }}</span>
          <span class="text-lg font-bold font-mono text-slate-900 dark:text-white mt-1 block">{{ metrics.dispatched_count || 0 }}</span>
        </div>
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-emerald-700 dark:text-emerald-400 font-semibold block">{{ $t('alarms.delivered') }}</span>
          <span class="text-lg font-bold font-mono text-emerald-700 dark:text-emerald-400 mt-1 block">{{ metrics.delivered_count || 0 }}</span>
        </div>
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-amber-700 dark:text-amber-400 font-semibold block">{{ $t('alarms.retrying') }}</span>
          <span class="text-lg font-bold font-mono text-amber-700 dark:text-amber-400 mt-1 block">{{ metrics.retried_count || 0 }}</span>
        </div>
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-rose-700 dark:text-rose-400 font-semibold block">{{ $t('alarms.failureReason') }}</span>
          <span class="text-lg font-bold font-mono text-rose-700 dark:text-rose-400 mt-1 block">{{ metrics.failed_count || 0 }}</span>
        </div>
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-semibold block">{{ $t('alarms.cancelled') }}</span>
          <span class="text-lg font-bold font-mono text-slate-600 dark:text-slate-400 mt-1 block">{{ metrics.cancelled_count || 0 }}</span>
        </div>
      </div>

      <!-- Filters & Action Bar -->
      <div class="p-4 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <!-- Status Filter -->
          <select v-model="filterStatus" @change="fetchNotifications"
                  class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 text-xs">
            <option value="">{{ $t('alarms.filterStatus') }}</option>
            <option value="PENDING">{{ $t('alarms.pending') }}</option>
            <option value="SENDING">{{ $t('alarms.sending') }}</option>
            <option value="RETRYING">{{ $t('alarms.retrying') }}</option>
            <option value="SENT">{{ $t('alarms.delivered') }}</option>
            <option value="FAILED">{{ $t('alarms.failureReason') }}</option>
            <option value="CANCELLED">{{ $t('alarms.cancelled') }}</option>
          </select>

          <!-- Channel Filter -->
          <select v-model="filterChannel" @change="fetchNotifications"
                  class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 text-xs">
            <option value="">{{ $t('alarms.filterChannel') }}</option>
            <option value="EMAIL">Email</option>
            <option value="TELEGRAM">Telegram</option>
            <option value="WEBHOOK">Webhook</option>
            <option value="WEB">In-App (WebSocket)</option>
          </select>

          <!-- Search Query -->
          <div class="relative">
            <input type="text"
                   v-model="searchQuery"
                   @input="debounceSearch"
                   :placeholder="$t('alarms.searchPlaceholder')"
                   class="w-64 px-3 py-1.5 pl-8 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white text-xs placeholder:text-slate-400 focus:outline-none focus:border-indigo-500 font-sans" />
            <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </div>
        </div>

        <button @click="fetchNotifications" class="text-2xs text-indigo-600 dark:text-indigo-400 hover:underline font-medium">
          {{ $t('common.refresh') }}
        </button>
      </div>

      <!-- Queue Table -->
      <div class="saas-card overflow-hidden bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse font-sans">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                <th class="py-3 px-4 w-28">{{ $t('alarms.channel') }}</th>
                <th class="py-3 px-4 w-48">{{ $t('alarms.recipient') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.severity') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.status') }}</th>
                <th class="py-3 px-4 w-36">{{ $t('alarms.sentAt') }}</th>
                <th class="py-3 px-4 w-36">{{ $t('alarms.nextRetry') }}</th>
                <th class="py-3 px-4 w-20 text-center">{{ $t('alarms.attempts') }}</th>
                <th class="py-3 px-4 w-32 text-center">{{ $t('alarms.action') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 text-slate-700 dark:text-slate-300">
              <tr v-if="notifications.length === 0">
                <td colspan="8" class="py-8 text-center text-slate-500">
                  {{ $t('alarms.noNotifications') }}
                </td>
              </tr>
              <tr v-for="notif in notifications" :key="notif.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/25 transition-colors">
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-bold font-mono tracking-wider"
                        :class="getChannelBadgeClass(notif.channel)">
                    {{ notif.channel }}
                  </span>
                </td>
                <td class="py-3 px-4 font-mono text-2xs truncate max-w-xs text-slate-900 dark:text-slate-200">
                  {{ notif.recipient }}
                </td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-bold font-mono tracking-wider"
                        :class="getSeverityBadgeClass(notif.severity)">
                    {{ notif.severity }}
                  </span>
                </td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-semibold font-mono"
                        :class="getStatusBadgeClass(notif.status)">
                    {{ notif.status }}
                  </span>
                </td>
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 font-mono text-3xs">
                  {{ notif.sent_at ? (notif.sent_at | formatDate) : '-' }}
                </td>
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 font-mono text-3xs">
                  {{ notif.next_retry_at ? (notif.next_retry_at | formatDate) : '-' }}
                </td>
                <td class="py-3 px-4 text-center font-mono text-2xs">
                  <span class="font-bold" :class="notif.attempts >= notif.max_retries ? 'text-rose-500' : 'text-slate-700 dark:text-slate-300'">
                    {{ notif.attempts }}
                  </span> / {{ notif.max_retries }}
                </td>
                <td class="py-3 px-4 text-center">
                  <div class="flex items-center justify-center space-x-1.5">
                    <button v-if="notif.status === 'FAILED' || notif.status === 'RETRYING'"
                            @click="retryNotification(notif.id)"
                            class="px-2 py-0.5 bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 rounded text-3xs font-semibold hover:bg-indigo-100 transition-colors">
                      {{ $t('alarms.retry') }}
                    </button>
                    <button v-if="notif.status === 'PENDING' || notif.status === 'RETRYING'"
                            @click="cancelNotification(notif.id)"
                            class="px-2 py-0.5 bg-rose-50 dark:bg-rose-900/30 text-rose-600 dark:text-rose-400 rounded text-3xs font-semibold hover:bg-rose-100 transition-colors">
                      {{ $t('alarms.cancelDelivery') }}
                    </button>
                    <button @click="openDetailsModal(notif)"
                            class="p-1 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors"
                            :title="$t('alarms.details')">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
          <div>
            {{ $t('alarms.page') || 'Page' }} <span class="font-bold text-slate-800 dark:text-slate-200">{{ queuePage }}</span>
            ({{ queueTotal }} {{ $t('alarms.totalRecords') || 'total' }})
          </div>
          <div class="flex items-center space-x-1">
            <button @click="changeQueuePage(queuePage - 1)"
                    :disabled="queuePage <= 1"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 disabled:opacity-40 transition-colors text-2xs font-medium">
              {{ $t('common.prev') }}
            </button>
            <button @click="changeQueuePage(queuePage + 1)"
                    :disabled="queuePage * queuePageSize >= queueTotal"
                    class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 disabled:opacity-40 transition-colors text-2xs font-medium">
              {{ $t('common.next') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- TAB 4: NOTIFICATION CHANNELS & ROUTING     -->
    <!-- ========================================== -->
    <div v-if="activeTab === 'channels'" class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div v-for="ch in channels" :key="ch.id"
             class="p-5 rounded-2xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-sm space-y-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <div class="w-10 h-10 rounded-xl flex items-center justify-center text-white"
                   :class="getChannelIconBg(ch.type)">
                <span class="font-bold font-mono text-xs">{{ ch.type.substring(0, 2) }}</span>
              </div>
              <div>
                <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">{{ ch.name }}</h3>
                <span class="text-3xs font-mono text-slate-500 uppercase">{{ ch.type }}</span>
              </div>
            </div>

            <!-- Enabled Switch / Status Pill -->
            <div class="flex items-center space-x-2">
              <span class="px-2 py-0.5 rounded-full text-3xs font-semibold font-mono"
                    :class="ch.enabled ? 'bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-500 dark:border-emerald-500/20' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-500 border border-slate-200 dark:border-slate-700'">
                {{ ch.enabled ? $t('alarms.enabled') : $t('alarms.disabled') }}
              </span>
            </div>
          </div>

          <!-- Parameters Summary -->
          <div class="grid grid-cols-2 gap-2 text-2xs font-sans bg-slate-50 dark:bg-slate-900/50 p-3 rounded-xl border border-slate-200 dark:border-slate-800/60">
            <div>
              <span class="text-slate-500 dark:text-slate-400 block text-3xs">{{ $t('alarms.minSeverity') }}</span>
              <span class="font-bold font-mono" :class="getSeverityTextColor(ch.min_severity)">{{ ch.min_severity }}</span>
            </div>
            <div>
              <span class="text-slate-500 dark:text-slate-400 block text-3xs">{{ $t('alarms.rateLimit') }}</span>
              <span class="font-mono text-slate-800 dark:text-slate-300 font-semibold">{{ ch.rate_limit_per_min }}/min</span>
            </div>
            <div>
              <span class="text-slate-500 dark:text-slate-400 block text-3xs">{{ $t('alarms.maxRetries') }} / {{ $t('alarms.retryInterval') }}</span>
              <span class="font-mono text-slate-800 dark:text-slate-300 font-medium">{{ ch.max_retries }} retries / {{ ch.retry_interval_sec }}s</span>
            </div>
            <div>
              <span class="text-slate-500 dark:text-slate-400 block text-3xs">{{ $t('alarms.quietHours') }}</span>
              <span class="font-mono text-slate-800 dark:text-slate-300 font-medium">
                {{ ch.quiet_hours_start ? `${ch.quiet_hours_start} - ${ch.quiet_hours_end}` : 'None' }}
              </span>
            </div>
          </div>

          <!-- Channel Actions -->
          <div class="flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800/80">
            <span class="text-3xs font-mono font-bold" :class="ch.is_configured ? 'text-emerald-700 dark:text-emerald-500' : 'text-amber-700 dark:text-amber-400'">
              {{ ch.is_configured ? `✓ ${$t('alarms.configured')}` : `⚠ ${$t('alarms.notConfigured')}` }}
            </span>
            <div class="flex items-center space-x-2">
              <button @click="openTestSendModal(ch)"
                      class="px-2.5 py-1 text-2xs rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 font-semibold transition-colors">
                {{ $t('alarms.testSend') }}
              </button>
              <button @click="openEditChannelModal(ch)"
                      class="px-2.5 py-1 text-2xs rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white font-semibold transition-colors shadow-xs">
                {{ $t('alarms.editChannel') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- MODAL 1: OPERATOR ACKNOWLEDGEMENT (Task 65)-->
    <!-- ========================================== -->
    <div v-if="showAckModal && selectedAckAlarm"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl space-y-4 p-6 font-sans">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <div class="flex items-center space-x-2">
            <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white">
              {{ $t('alarms.ackModalTitle') || 'Acknowledge Alarm Incident' }}
            </h3>
          </div>
          <button @click="showAckModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">&times;</button>
        </div>

        <p class="text-xs text-slate-600 dark:text-slate-400">
          {{ $t('alarms.ackModalDesc') || 'Operator acknowledgement records your identity, timestamp, and optional field notes in the system audit trail.' }}
        </p>

        <!-- Incident Metadata Summary -->
        <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 space-y-2 text-xs">
          <div class="flex items-center justify-between">
            <span class="font-bold font-mono text-slate-900 dark:text-slate-100 text-sm">{{ selectedAckAlarm.alarm_code }}</span>
            <span class="px-2 py-0.5 rounded text-3xs font-bold font-mono" :class="getSeverityBadgeClass(selectedAckAlarm.severity)">
              {{ selectedAckAlarm.severity }}
            </span>
          </div>
          <p class="text-slate-700 dark:text-slate-300 text-2xs">{{ selectedAckAlarm.message }}</p>
          <div class="grid grid-cols-2 gap-2 pt-2 border-t border-slate-200/50 dark:border-slate-800 text-3xs text-slate-500 dark:text-slate-400">
            <div>
              <span>Device: </span>
              <strong class="font-mono text-slate-700 dark:text-slate-300">
                {{ selectedAckAlarm.device ? (selectedAckAlarm.device.code || selectedAckAlarm.device.name) : 'System' }}
              </strong>
            </div>
            <div>
              <span>Triggered: </span>
              <span class="font-mono text-slate-700 dark:text-slate-300">{{ selectedAckAlarm.triggered_at | formatDate }}</span>
            </div>
            <div v-if="selectedAckAlarm.trigger_value != null">
              <span>Trigger Value: </span>
              <strong class="font-mono text-rose-500">{{ Number(selectedAckAlarm.trigger_value).toFixed(2) }}</strong>
            </div>
            <div v-if="selectedAckAlarm.threshold_value != null">
              <span>Threshold: </span>
              <strong class="font-mono text-slate-600 dark:text-slate-400">{{ Number(selectedAckAlarm.threshold_value).toFixed(2) }}</strong>
            </div>
          </div>
        </div>

        <!-- Operator Identity (Server-Derived) -->
        <div class="flex items-center justify-between p-2.5 rounded-lg bg-indigo-50/60 dark:bg-indigo-900/20 border border-indigo-100 dark:border-indigo-900/40 text-2xs">
          <div class="flex items-center space-x-2">
            <div class="w-6 h-6 rounded-full bg-indigo-600 text-white flex items-center justify-center font-bold text-3xs">
              {{ operatorDisplayName.charAt(0).toUpperCase() }}
            </div>
            <div>
              <span class="text-slate-500 dark:text-slate-400 block text-3xs">Operator Identity</span>
              <strong class="text-slate-900 dark:text-white">{{ operatorDisplayName }}</strong>
            </div>
          </div>
          <span class="text-3xs font-mono text-indigo-600 dark:text-indigo-400">Authenticated (JWT)</span>
        </div>

        <!-- Operator Notes -->
        <div>
          <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 mb-1">
            {{ $t('alarms.operatorNotes') || 'Operator Notes (Optional)' }}
          </label>
          <textarea id="input-ack-notes"
                    v-model="ackNotes"
                    rows="3"
                    :placeholder="$t('alarms.notesPlaceholder') || 'Describe action taken, root cause, or field observations...'"
                    class="w-full px-3 py-2 rounded-xl bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white text-xs placeholder:text-slate-400 focus:outline-none focus:border-indigo-500 font-sans"></textarea>
        </div>

        <div class="flex justify-end space-x-2 pt-3 border-t border-slate-200 dark:border-slate-800">
          <button type="button" @click="showAckModal = false"
                  class="px-3.5 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
            {{ $t('common.cancel') }}
          </button>
          <button id="btn-confirm-ack"
                  type="button"
                  @click="submitAckModal"
                  :disabled="isAckSubmitting"
                  class="px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
            <span v-if="isAckSubmitting" class="animate-spin w-3 h-3 mr-1.5 border-2 border-white border-t-transparent rounded-full"></span>
            {{ isAckSubmitting ? 'Acknowledging...' : ($t('alarms.confirmAck') || 'Confirm Acknowledgement') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- MODAL 2: INCIDENT DETAILS & TIMELINE       -->
    <!-- ========================================== -->
    <div v-if="showIncidentModal && selectedIncident"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl space-y-4 p-6 font-sans">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <div class="flex items-center space-x-2">
            <h3 class="text-sm font-bold text-slate-900 dark:text-white">
              {{ $t('alarms.incidentDetails') || 'Incident Lifecycle Details' }}
            </h3>
            <span class="px-2 py-0.5 rounded text-3xs font-bold font-mono" :class="getSeverityBadgeClass(selectedIncident.severity)">
              {{ selectedIncident.severity }}
            </span>
            <span class="px-2 py-0.5 rounded text-3xs font-semibold font-mono" :class="getStatusBadgeClass(selectedIncident.status)">
              {{ selectedIncident.status }}
            </span>
          </div>
          <button @click="showIncidentModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">&times;</button>
        </div>

        <!-- Alarm Summary Header -->
        <div class="space-y-1">
          <div class="text-base font-bold font-mono text-slate-900 dark:text-white">{{ selectedIncident.alarm_code }}</div>
          <div class="text-xs text-slate-700 dark:text-slate-300">{{ selectedIncident.message }}</div>
        </div>

        <!-- Device & Metric Context -->
        <div class="grid grid-cols-2 gap-3 p-3 rounded-xl bg-slate-50 dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 text-xs">
          <div>
            <span class="text-3xs text-slate-400 block">{{ $t('alarms.device') }}</span>
            <strong class="font-mono text-slate-800 dark:text-slate-200">
              {{ selectedIncident.device_code || selectedIncident.device_name || (selectedIncident.device ? (selectedIncident.device.code || selectedIncident.device.name) : 'System') }}
            </strong>
          </div>
          <div>
            <span class="text-3xs text-slate-400 block">{{ $t('alarms.duration') || 'Duration' }}</span>
            <strong class="font-mono" :class="selectedIncident.is_ongoing ? 'text-amber-700 dark:text-amber-400 font-bold' : 'text-slate-700 dark:text-slate-300'">
              {{ selectedIncident.duration_formatted || '-' }}
            </strong>
          </div>
          <div>
            <span class="text-3xs text-slate-400 block">{{ $t('alarms.triggerValue') || 'Trigger Value' }}</span>
            <span v-if="selectedIncident.trigger_value != null" class="font-mono font-bold text-rose-500">
              {{ Number(selectedIncident.trigger_value).toFixed(2) }}
            </span>
            <span v-else class="text-slate-400 font-mono text-3xs">{{ $t('alarms.legacyUnavailable') || 'N/A' }}</span>
          </div>
          <div>
            <span class="text-3xs text-slate-400 block">{{ $t('alarms.threshold') || 'Threshold Limit' }}</span>
            <span v-if="selectedIncident.threshold_value != null" class="font-mono text-slate-600 dark:text-slate-400">
              {{ Number(selectedIncident.threshold_value).toFixed(2) }}
            </span>
            <span v-else class="text-slate-400 font-mono text-3xs">{{ $t('alarms.legacyUnavailable') || 'N/A' }}</span>
          </div>
        </div>

        <!-- Lifecycle Timeline -->
        <div class="space-y-3">
          <h4 class="text-2xs font-bold uppercase tracking-wider text-slate-400 font-sans">
            {{ $t('alarms.timeline') || 'Lifecycle Timeline' }}
          </h4>

          <div class="relative pl-6 space-y-4 before:absolute before:left-2 before:top-2 before:bottom-2 before:w-0.5 before:bg-slate-200 dark:before:bg-slate-800">
            <!-- Step 1: Triggered -->
            <div class="relative">
              <span class="absolute -left-6 top-1 w-2.5 h-2.5 rounded-full bg-rose-500 ring-4 ring-white dark:ring-[#111827]"></span>
              <div class="text-xs font-semibold text-slate-900 dark:text-white">{{ $t('alarms.triggered') || 'Alarm Triggered' }}</div>
              <div class="text-3xs font-mono text-slate-400">{{ selectedIncident.triggered_at | formatDate }}</div>
            </div>

            <!-- Step 2: Acknowledged -->
            <div class="relative">
              <span class="absolute -left-6 top-1 w-2.5 h-2.5 rounded-full ring-4 ring-white dark:ring-[#111827]"
                    :class="selectedIncident.acknowledged_at ? 'bg-amber-500' : 'bg-slate-300 dark:bg-slate-700'"></span>
              <div class="text-xs font-semibold"
                   :class="selectedIncident.acknowledged_at ? 'text-slate-900 dark:text-white' : 'text-slate-400'">
                {{ $t('alarms.acknowledgedState') || 'Operator Acknowledgement' }}
              </div>
              <div v-if="selectedIncident.acknowledged_at" class="space-y-1">
                <div class="text-3xs font-mono text-slate-400">{{ selectedIncident.acknowledged_at | formatDate }}</div>
                <div class="text-2xs text-slate-700 dark:text-slate-300">
                  By: <strong>{{ selectedIncident.acknowledged_by_name || selectedIncident.acknowledged_by }}</strong>
                </div>
                <div v-if="selectedIncident.notes" class="text-2xs italic text-indigo-700 dark:text-indigo-400 bg-indigo-50/70 dark:bg-indigo-900/20 p-2 rounded-lg border border-indigo-100 dark:border-indigo-900/40">
                  "{{ selectedIncident.notes }}"
                </div>
              </div>
              <div v-else class="text-3xs text-slate-400 italic">Pending operator acknowledgement</div>
            </div>

            <!-- Step 3: Cleared -->
            <div class="relative">
              <span class="absolute -left-6 top-1 w-2.5 h-2.5 rounded-full ring-4 ring-white dark:ring-[#111827]"
                    :class="selectedIncident.cleared_at ? 'bg-emerald-500' : 'bg-slate-300 dark:bg-slate-700'"></span>
              <div class="text-xs font-semibold"
                   :class="selectedIncident.cleared_at ? 'text-slate-900 dark:text-white' : 'text-slate-400'">
                {{ $t('alarms.clearedState') || 'Cleared / Recovered' }}
              </div>
              <div v-if="selectedIncident.cleared_at" class="space-y-1">
                <div class="text-3xs font-mono text-slate-400">{{ selectedIncident.cleared_at | formatDate }}</div>
                <div v-if="selectedIncident.recovery_reason" class="text-2xs text-emerald-600 dark:text-emerald-400">
                  Reason: {{ selectedIncident.recovery_reason }}
                </div>
              </div>
              <div v-else class="text-3xs text-amber-700 dark:text-amber-400 font-semibold flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-ping"></span>
                Incident currently active and awaiting resolution
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end space-x-2 pt-3 border-t border-slate-200 dark:border-slate-800">
          <button v-if="selectedIncident.status === 'ACTIVE'"
                  @click="showIncidentModal = false; openAckModal(selectedIncident)"
                  class="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold transition-colors">
            {{ $t('alarms.acknowledge') }}
          </button>
          <button @click="showIncidentModal = false" class="px-4 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
            {{ $t('alarms.close') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- MODAL 3: NOTIFICATION ATTEMPT DETAILS      -->
    <!-- ========================================== -->
    <div v-if="showDetailsModal && selectedNotif"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl space-y-4 p-6 font-sans">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
            {{ $t('alarms.notificationDetails') }}
          </h3>
          <button @click="showDetailsModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">&times;</button>
        </div>

        <div class="space-y-3 text-xs font-sans">
          <div class="grid grid-cols-2 gap-2">
            <div>
              <span class="text-slate-400 block text-3xs">{{ $t('alarms.channel') }}</span>
              <span class="font-bold font-mono">{{ selectedNotif.channel }}</span>
            </div>
            <div>
              <span class="text-slate-400 block text-3xs">{{ $t('alarms.recipient') }}</span>
              <span class="font-mono text-slate-900 dark:text-white">{{ selectedNotif.recipient }}</span>
            </div>
            <div>
              <span class="text-slate-400 block text-3xs">{{ $t('alarms.status') }}</span>
              <span class="px-2 py-0.5 rounded text-3xs font-semibold font-mono" :class="getStatusBadgeClass(selectedNotif.status)">
                {{ selectedNotif.status }}
              </span>
            </div>
            <div>
              <span class="text-slate-400 block text-3xs">{{ $t('alarms.attempts') }}</span>
              <span class="font-mono">{{ selectedNotif.attempts }} / {{ selectedNotif.max_retries }}</span>
            </div>
          </div>

          <div v-if="selectedNotif.provider_response">
            <span class="text-slate-400 block text-3xs mb-1">{{ $t('alarms.providerResponse') }}</span>
            <pre class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 font-mono text-3xs text-slate-700 dark:text-slate-300 overflow-x-auto whitespace-pre-wrap">{{ selectedNotif.provider_response }}</pre>
          </div>

          <div v-if="selectedNotif.failure_reason">
            <span class="text-rose-600 dark:text-rose-400 block text-3xs mb-1 font-semibold">{{ $t('alarms.failureReason') }}</span>
            <pre class="p-2.5 rounded-lg bg-rose-50 dark:bg-rose-500/10 border border-rose-200 dark:border-rose-500/20 font-mono text-3xs text-rose-700 dark:text-rose-400 overflow-x-auto whitespace-pre-wrap">{{ selectedNotif.failure_reason }}</pre>
          </div>
        </div>

        <div class="flex justify-end pt-3 border-t border-slate-200 dark:border-slate-800">
          <button @click="showDetailsModal = false" class="px-4 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
            {{ $t('alarms.close') }}
          </button>
        </div>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- MODAL 4: CHANNEL CONFIGURATION             -->
    <!-- ========================================== -->
    <div v-if="showEditModal && editingChannel"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl space-y-4 p-6 font-sans">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
            {{ $t('alarms.editChannel') }}: {{ editingChannel.name }}
          </h3>
          <button @click="showEditModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">&times;</button>
        </div>

        <form @submit.prevent="saveChannelConfig" class="space-y-3 text-xs font-sans">
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">Name</label>
              <input v-model="editForm.name" required
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.minSeverity') }}</label>
              <select v-model="editForm.min_severity"
                      class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white">
                <option value="INFO">INFO</option>
                <option value="WARNING">WARNING</option>
                <option value="CRITICAL">CRITICAL</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-3 gap-3">
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.rateLimit') }}</label>
              <input type="number" v-model.number="editForm.rate_limit_per_min" min="1" max="100"
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.maxRetries') }}</label>
              <input type="number" v-model.number="editForm.max_retries" min="0" max="10"
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.retryInterval') }}</label>
              <input type="number" v-model.number="editForm.retry_interval_sec" min="5" max="3600"
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
            </div>
          </div>

          <!-- Channel-Specific Credentials -->
          <div class="p-3 rounded-xl bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-2">
            <h4 class="text-3xs uppercase font-bold text-slate-400 tracking-wider">Channel Specific Provider Configuration</h4>

            <!-- EMAIL CONFIG -->
            <div v-if="editingChannel.type === 'EMAIL'" class="space-y-2">
              <div class="grid grid-cols-2 gap-2">
                <div>
                  <label class="block text-3xs text-slate-500 mb-0.5">SMTP Host</label>
                  <input v-model="editForm.config.smtp_host" placeholder="smtp.gmail.com"
                         class="w-full px-2.5 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-3xs text-slate-500 mb-0.5">SMTP Port</label>
                  <input type="number" v-model.number="editForm.config.smtp_port" placeholder="587"
                         class="w-full px-2.5 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white font-mono" />
                </div>
              </div>
              <div class="grid grid-cols-2 gap-2">
                <div>
                  <label class="block text-3xs text-slate-500 mb-0.5">Username</label>
                  <input v-model="editForm.config.username" placeholder="user@gmail.com"
                         class="w-full px-2.5 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white" />
                </div>
                <div>
                  <label class="block text-3xs text-slate-500 mb-0.5">Password / Secret</label>
                  <input type="password" v-model="editForm.config.password" placeholder="********"
                         class="w-full px-2.5 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white font-mono" />
                </div>
              </div>
            </div>

            <!-- TELEGRAM CONFIG -->
            <div v-else-if="editingChannel.type === 'TELEGRAM'" class="space-y-2">
              <div>
                <label class="block text-3xs text-slate-500 mb-0.5">Bot Token</label>
                <input type="password" v-model="editForm.config.bot_token" placeholder="********"
                       class="w-full px-2.5 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white font-mono" />
              </div>
              <div>
                <label class="block text-3xs text-slate-500 mb-0.5">Chat ID / Channel ID</label>
                <input v-model="editForm.config.chat_id" placeholder="-10012345678"
                       class="w-full px-2.5 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white font-mono" />
              </div>
            </div>

            <!-- WEBHOOK CONFIG -->
            <div v-else-if="editingChannel.type === 'WEBHOOK'" class="space-y-2">
              <div>
                <label class="block text-3xs text-slate-500 mb-0.5">Webhook URL</label>
                <input v-model="editForm.config.webhook_url" placeholder="https://api.ops.corp/alerts"
                       class="w-full px-2.5 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white" />
              </div>
              <div>
                <label class="block text-3xs text-slate-500 mb-0.5">Auth Token (Optional)</label>
                <input type="password" v-model="editForm.config.auth_token" placeholder="********"
                       class="w-full px-2.5 py-1 rounded bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white font-mono" />
              </div>
            </div>

            <!-- WEB / IN-APP -->
            <div v-else class="text-3xs text-slate-400">
              In-app alert channel broadcasts live events directly to active browser WebSocket clients.
            </div>
          </div>

          <div class="flex justify-end space-x-2 pt-3 border-t border-slate-200 dark:border-slate-800">
            <button type="button" @click="showEditModal = false" class="px-3.5 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" class="px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm transition-colors">
              {{ $t('alarms.saveConfig') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- MODAL 5: TEST SEND MODAL                   -->
    <!-- ========================================== -->
    <div v-if="showTestModal && testingChannel"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-md overflow-hidden shadow-2xl space-y-4 p-6 font-sans">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
            {{ $t('alarms.testSend') }}: {{ testingChannel.name }}
          </h3>
          <button @click="showTestModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">&times;</button>
        </div>

        <div class="space-y-3 text-xs font-sans">
          <p class="text-slate-600 dark:text-slate-400">
            Send an authorized diagnostic notification to test credential validity and delivery connectivity.
          </p>

          <div>
            <label class="block text-slate-500 dark:text-slate-400 mb-1">Recipient Override (Optional)</label>
            <input v-model="testRecipient" placeholder="e.g. alerts@domain.com or chat ID"
                   class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
          </div>

          <!-- Diagnostic Test Result -->
          <div v-if="testResult"
               class="p-3 rounded-lg border text-2xs font-mono space-y-1"
               :class="testResult.success ? 'bg-emerald-50 dark:bg-emerald-500/10 border-emerald-200 dark:border-emerald-500/30 text-emerald-700 dark:text-emerald-400' : 'bg-rose-50 dark:bg-rose-500/10 border-rose-200 dark:border-rose-500/30 text-rose-700 dark:text-rose-400'">
            <div class="font-bold uppercase">{{ testResult.success ? $t('alarms.testSuccess') : $t('alarms.testFailed') }}</div>
            <div class="break-all">{{ testResult.response || testResult.error }}</div>
          </div>
        </div>

        <div class="flex justify-end space-x-2 pt-3 border-t border-slate-200 dark:border-slate-800">
          <button @click="showTestModal = false" class="px-3.5 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
            {{ $t('alarms.close') }}
          </button>
          <button @click="executeTestSend"
                  :disabled="isTesting"
                  class="px-4 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
            <span v-if="isTesting" class="animate-spin w-3 h-3 mr-1.5 border-2 border-white border-t-transparent rounded-full"></span>
            {{ isTesting ? 'Testing...' : 'Dispatch Test' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'AlarmsView',
  data() {
    return {
      activeTab: 'alarms', // 'alarms', 'history', 'queue', 'channels'
      toastMessage: null,
      toastType: 'success',

      // History state (Task #66)
      historyItems: [],
      historyTotal: 0,
      historyPage: 1,
      historyPageSize: 20,
      historyTotalPages: 1,
      historyStatus: '',
      historySeverity: '',
      historyDeviceID: '',
      historyStartDate: '',
      historyEndDate: '',
      historySearch: '',
      historyLoading: false,
      historyDebounceTimer: null,

      // Acknowledgement Modal (Task #65)
      showAckModal: false,
      selectedAckAlarm: null,
      ackNotes: '',
      isAckSubmitting: false,

      // Incident Details Modal
      showIncidentModal: false,
      selectedIncident: null,

      // Queue state (Wave 3)
      notifications: [],
      queueTotal: 0,
      queuePage: 1,
      queuePageSize: 20,
      filterStatus: '',
      filterChannel: '',
      searchQuery: '',
      searchDebounceTimer: null,
      metrics: {},

      // Channels state (Wave 3)
      channels: [],

      // Notification Details Modal
      showDetailsModal: false,
      selectedNotif: null,

      // Channel Config Modal
      showEditModal: false,
      editingChannel: null,
      editForm: {
        name: '',
        enabled: true,
        min_severity: 'WARNING',
        rate_limit_per_min: 10,
        max_retries: 3,
        retry_interval_sec: 30,
        quiet_hours_start: '',
        quiet_hours_end: '',
        timezone: 'UTC',
        config: {},
      },

      // Test Send Modal
      showTestModal: false,
      testingChannel: null,
      testRecipient: '',
      isTesting: false,
      testResult: null,
    };
  },
  computed: {
    alarms() {
      return this.$store.state.alarms || [];
    },
    activeAlarmsList() {
      return this.alarms.filter((a) => {
        if (a.status !== 'ACTIVE') return false;
        const devId = a.device_id || (a.device && a.device.id);
        if (devId && this.devices && this.devices.length) {
          const d = this.devices.find(x => Number(x.id) === Number(devId));
          if (!d) return false;
          if (d.enabled === false || (d.status || d.admin_status || d.administrative_status || '').toUpperCase() === 'DISABLED') {
            return false;
          }
        }
        return true;
      });
    },
    activeCount() {
      return this.activeAlarmsList.length;
    },
    devices() {
      return this.$store.state.devices || [];
    },
    operatorDisplayName() {
      const user = this.$store.state.user;
      if (user && user.full_name) return user.full_name;
      if (user && user.username) return user.username;
      return 'Lead Operator';
    },
  },
  watch: {
    activeTab(newTab) {
      if (newTab === 'history') {
        this.fetchAlarmHistory();
      } else if (newTab === 'alarms') {
        this.refreshAlarms();
      } else if (newTab === 'queue') {
        this.fetchNotifications();
        this.fetchMetrics();
      } else if (newTab === 'channels') {
        this.fetchChannels();
      }
    },
    '$route.query.tab'(tab) {
      if (tab && ['alarms', 'history', 'queue', 'channels'].includes(tab)) {
        this.activeTab = tab;
      }
    },
    '$route.path'(path) {
      if (path.includes('/queue') || path.includes('/delivery')) {
        this.activeTab = 'queue';
      } else if (path.includes('/history')) {
        this.activeTab = 'history';
      } else if (path.includes('/channels')) {
        this.activeTab = 'channels';
      } else if (path.includes('/active')) {
        this.activeTab = 'alarms';
      }
    },
  },
  mounted() {
    this.$store.dispatch('fetchDevices');
    if (this.$route.query.tab && ['alarms', 'history', 'queue', 'channels'].includes(this.$route.query.tab)) {
      this.activeTab = this.$route.query.tab;
    } else if (this.$route.path.includes('/queue') || this.$route.path.includes('/delivery')) {
      this.activeTab = 'queue';
    } else if (this.$route.path.includes('/history')) {
      this.activeTab = 'history';
    } else if (this.$route.path.includes('/channels')) {
      this.activeTab = 'channels';
    }
    this.refreshAlarms();
    this.fetchAlarmHistory();
    this.fetchNotifications();
    this.fetchChannels();
    this.fetchMetrics();
  },
  methods: {
    getAuthHeaders() {
      const token = (this.$store && this.$store.state.token) ||
                    localStorage.getItem('datalogger_token') ||
                    localStorage.getItem('token');
      return token ? { Authorization: `Bearer ${token}` } : {};
    },
    showToast(msg, type = 'success') {
      this.toastMessage = msg;
      this.toastType = type;
      setTimeout(() => {
        if (this.toastMessage === msg) {
          this.toastMessage = null;
        }
      }, 5000);
    },
    refreshAlarms() {
      this.$store.dispatch('fetchAlarms');
    },

    // ------------------------------------------
    // Task #65: Alarm Acknowledgement
    // ------------------------------------------
    openAckModal(alarm) {
      this.selectedAckAlarm = alarm;
      this.ackNotes = '';
      this.showAckModal = true;
    },
    async submitAckModal() {
      if (!this.selectedAckAlarm) return;
      this.isAckSubmitting = true;
      try {
        await this.$store.dispatch('acknowledgeAlarm', {
          id: this.selectedAckAlarm.id,
          notes: this.ackNotes,
        });
        this.showToast(this.$t('alarms.ackSuccess') || 'Alarm acknowledged successfully');
        this.showAckModal = false;
        this.refreshAlarms();
        if (this.activeTab === 'history') {
          this.fetchAlarmHistory();
        }
      } catch (err) {
        const errorMsg = (err.response && err.response.data && (err.response.data.error || err.response.data.message)) || err.message;
        this.showToast('Failed to acknowledge alarm: ' + errorMsg, 'error');
      } finally {
        this.isAckSubmitting = false;
      }
    },

    // ------------------------------------------
    // Task #66: Alarm History
    // ------------------------------------------
    async fetchAlarmHistory() {
      this.historyLoading = true;
      try {
        let url = `/api/alarms/history?page=${this.historyPage}&page_size=${this.historyPageSize}`;
        if (this.historyStatus) url += `&status=${encodeURIComponent(this.historyStatus)}`;
        if (this.historySeverity) url += `&severity=${encodeURIComponent(this.historySeverity)}`;
        if (this.historyDeviceID) url += `&device_id=${encodeURIComponent(this.historyDeviceID)}`;
        if (this.historyStartDate) url += `&start_time=${encodeURIComponent(this.historyStartDate + 'T00:00:00Z')}`;
        if (this.historyEndDate) url += `&end_time=${encodeURIComponent(this.historyEndDate + 'T23:59:59Z')}`;
        if (this.historySearch) url += `&search=${encodeURIComponent(this.historySearch)}`;

        const res = await axios.get(url, { headers: this.getAuthHeaders() });
        if (res.data && res.data.success) {
          this.historyItems = res.data.data.items || [];
          this.historyTotal = res.data.data.total || 0;
          this.historyTotalPages = res.data.data.total_pages || 1;
        }
      } catch (err) {
        console.error('Failed fetching alarm history:', err);
      } finally {
        this.historyLoading = false;
      }
    },
    changeHistoryPage(page) {
      if (page >= 1 && page <= this.historyTotalPages) {
        this.historyPage = page;
        this.fetchAlarmHistory();
      }
    },
    debounceHistorySearch() {
      clearTimeout(this.historyDebounceTimer);
      this.historyDebounceTimer = setTimeout(() => {
        this.historyPage = 1;
        this.fetchAlarmHistory();
      }, 300);
    },
    resetHistoryFilters() {
      this.historyStatus = '';
      this.historySeverity = '';
      this.historyDeviceID = '';
      this.historyStartDate = '';
      this.historyEndDate = '';
      this.historySearch = '';
      this.historyPage = 1;
      this.fetchAlarmHistory();
    },

    // ------------------------------------------
    // Incident Details Modal
    // ------------------------------------------
    openIncidentModal(alarm) {
      this.selectedIncident = alarm;
      this.showIncidentModal = true;
    },

    // ------------------------------------------
    // Wave 3: Notifications & Channels
    // ------------------------------------------
    debounceSearch() {
      clearTimeout(this.searchDebounceTimer);
      this.searchDebounceTimer = setTimeout(() => {
        this.queuePage = 1;
        this.fetchNotifications();
      }, 300);
    },
    changeQueuePage(page) {
      if (page >= 1) {
        this.queuePage = page;
        this.fetchNotifications();
      }
    },
    async fetchNotifications() {
      try {
        let url = `/api/notifications?page=${this.queuePage}&page_size=${this.queuePageSize}`;
        if (this.filterStatus) url += `&status=${this.filterStatus}`;
        if (this.filterChannel) url += `&channel=${this.filterChannel}`;
        if (this.searchQuery) url += `&search=${encodeURIComponent(this.searchQuery)}`;

        const res = await axios.get(url, { headers: this.getAuthHeaders() });
        if (res.data && res.data.success) {
          this.notifications = res.data.data.items || [];
          this.queueTotal = res.data.data.total || 0;
        }
      } catch (err) {
        console.error('Failed fetching notifications:', err);
      }
    },
    async fetchChannels() {
      try {
        const res = await axios.get('/api/notifications/channels', { headers: this.getAuthHeaders() });
        if (res.data && res.data.success) {
          this.channels = res.data.data || [];
        }
      } catch (err) {
        console.error('Failed fetching channels:', err);
      }
    },
    async fetchMetrics() {
      try {
        const res = await axios.get('/api/notifications/metrics', { headers: this.getAuthHeaders() });
        if (res.data && res.data.success) {
          this.metrics = res.data.data || {};
        }
      } catch (err) {
        console.error('Failed fetching metrics:', err);
      }
    },
    async retryNotification(id) {
      try {
        const res = await axios.post(`/api/notifications/${id}/retry`, {}, { headers: this.getAuthHeaders() });
        if (res.data && res.data.success) {
          this.showToast('Notification re-queued for delivery');
          this.fetchNotifications();
          this.fetchMetrics();
        }
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.error) || err.message;
        this.showToast('Failed to retry notification: ' + msg, 'error');
      }
    },
    async cancelNotification(id) {
      try {
        const res = await axios.post(`/api/notifications/${id}/cancel`, {}, { headers: this.getAuthHeaders() });
        if (res.data && res.data.success) {
          this.showToast('Notification delivery cancelled');
          this.fetchNotifications();
          this.fetchMetrics();
        }
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.error) || err.message;
        this.showToast('Failed to cancel notification: ' + msg, 'error');
      }
    },
    openDetailsModal(notif) {
      this.selectedNotif = notif;
      this.showDetailsModal = true;
    },
    openEditChannelModal(ch) {
      this.editingChannel = ch;
      this.editForm = {
        name: ch.name,
        enabled: ch.enabled,
        min_severity: ch.min_severity,
        rate_limit_per_min: ch.rate_limit_per_min,
        max_retries: ch.max_retries,
        retry_interval_sec: ch.retry_interval_sec,
        quiet_hours_start: ch.quiet_hours_start || '',
        quiet_hours_end: ch.quiet_hours_end || '',
        timezone: ch.timezone || 'UTC',
        config: ch.config ? JSON.parse(JSON.stringify(ch.config)) : {},
      };
      this.showEditModal = true;
    },
    async saveChannelConfig() {
      try {
        const payload = {
          name: this.editForm.name,
          enabled: this.editForm.enabled,
          min_severity: this.editForm.min_severity,
          rate_limit_per_min: this.editForm.rate_limit_per_min,
          max_retries: this.editForm.max_retries,
          retry_interval_sec: this.editForm.retry_interval_sec,
          quiet_hours_start: this.editForm.quiet_hours_start,
          quiet_hours_end: this.editForm.quiet_hours_end,
          timezone: this.editForm.timezone,
          config: this.editForm.config,
        };

        const res = await axios.put(`/api/notifications/channels/${this.editingChannel.id}`, payload, { headers: this.getAuthHeaders() });
        if (res.data && res.data.success) {
          this.showToast('Channel updated successfully');
          this.showEditModal = false;
          this.fetchChannels();
        }
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.error) || err.message;
        this.showToast('Failed saving channel: ' + msg, 'error');
      }
    },
    openTestSendModal(ch) {
      this.testingChannel = ch;
      this.testRecipient = '';
      this.testResult = null;
      this.showTestModal = true;
    },
    async executeTestSend() {
      this.isTesting = true;
      this.testResult = null;
      try {
        const res = await axios.post(`/api/notifications/channels/${this.testingChannel.id}/test`, {
          recipient: this.testRecipient,
        }, { headers: this.getAuthHeaders() });

        if (res.data && res.data.success) {
          this.testResult = {
            success: true,
            response: res.data.data.response || res.data.data.message || 'Delivered successfully',
          };
        }
      } catch (err) {
        const errorMsg = (err.response && err.response.data && (err.response.data.error || err.response.data.details)) || err.message;
        this.testResult = {
          success: false,
          error: errorMsg,
        };
      } finally {
        this.isTesting = false;
      }
    },

    // ------------------------------------------
    // Style & Badge Helpers
    // ------------------------------------------
    getSeverityBadgeClass(sev) {
      switch (sev) {
        case 'CRITICAL':
          return 'bg-rose-50 dark:bg-rose-500/10 text-rose-700 dark:text-rose-400 border border-rose-200 dark:border-rose-500/20 font-bold';
        case 'WARNING':
          return 'bg-amber-50 dark:bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-500/20 font-bold';
        default:
          return 'bg-blue-50 dark:bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-200 dark:border-blue-500/20 font-bold';
      }
    },
    getSeverityTextColor(sev) {
      switch (sev) {
        case 'CRITICAL':
          return 'text-rose-600 dark:text-rose-400 font-bold';
        case 'WARNING':
          return 'text-amber-700 dark:text-amber-400 font-bold';
        default:
          return 'text-blue-600 dark:text-blue-400 font-bold';
      }
    },
    getStatusBadgeClass(status) {
      switch (status) {
        case 'ACTIVE':
          return 'bg-rose-50 dark:bg-rose-500/10 text-rose-700 dark:text-rose-400 border border-rose-200 dark:border-rose-500/20 font-bold';
        case 'ACKNOWLEDGED':
          return 'bg-amber-50 dark:bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-500/20 font-bold';
        case 'CLEARED':
          return 'bg-emerald-50 dark:bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/20 font-bold';
        case 'SENT':
          return 'bg-emerald-50 dark:bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/20 font-bold';
        case 'SENDING':
          return 'bg-blue-50 dark:bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-200 dark:border-blue-500/20 font-bold';
        case 'RETRYING':
          return 'bg-amber-50 dark:bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-500/20 font-bold';
        case 'FAILED':
          return 'bg-rose-50 dark:bg-rose-500/10 text-rose-700 dark:text-rose-400 border border-rose-200 dark:border-rose-500/20 font-bold';
        case 'CANCELLED':
          return 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700';
        default:
          return 'bg-indigo-50 dark:bg-indigo-500/10 text-indigo-700 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-500/20 font-bold';
      }
    },
    getChannelBadgeClass(chType) {
      switch (chType) {
        case 'EMAIL':
          return 'bg-blue-50 dark:bg-blue-500/10 text-blue-700 dark:text-blue-400 border border-blue-200 dark:border-blue-500/20 font-bold';
        case 'TELEGRAM':
          return 'bg-cyan-50 dark:bg-cyan-500/10 text-cyan-700 dark:text-cyan-400 border border-cyan-200 dark:border-cyan-500/20 font-bold';
        case 'WEBHOOK':
          return 'bg-purple-50 dark:bg-purple-500/10 text-purple-700 dark:text-purple-400 border border-purple-200 dark:border-purple-500/20 font-bold';
        default:
          return 'bg-emerald-50 dark:bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/20 font-bold';
      }
    },
    getChannelIconBg(chType) {
      switch (chType) {
        case 'EMAIL':
          return 'bg-blue-600';
        case 'TELEGRAM':
          return 'bg-cyan-600';
        case 'WEBHOOK':
          return 'bg-purple-600';
        default:
          return 'bg-emerald-600';
      }
    },
  },
};
</script>
