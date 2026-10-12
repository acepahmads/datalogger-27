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
              :class="activeCount > 0 ? 'bg-rose-500/10 text-rose-500 dark:text-rose-400 border border-rose-500/20' : 'bg-emerald-500/10 text-emerald-500 dark:text-emerald-400 border border-emerald-500/20'">
          {{ $t('alarms.activeCount', { count: activeCount }) }}
        </span>

        <!-- View Switcher Tabs -->
        <div class="flex items-center space-x-1 bg-white dark:bg-[#111827] p-1 rounded-lg border border-slate-200 dark:border-slate-800 text-xs font-sans shadow-sm">
          <button @click="activeTab = 'alarms'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'alarms' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('alarms.tabAlarms') }}
          </button>
          <button @click="activeTab = 'queue'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'queue' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('alarms.tabQueue') }}
          </button>
          <button @click="activeTab = 'channels'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'channels' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('alarms.tabChannels') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Alert / Toast Banner -->
    <div v-if="toastMessage"
         class="p-3.5 rounded-xl border flex items-center justify-between text-xs transition-all"
         :class="toastType === 'error' ? 'bg-rose-500/10 border-rose-500/30 text-rose-500 dark:text-rose-400' : 'bg-emerald-500/10 border-emerald-500/30 text-emerald-600 dark:text-emerald-400'">
      <div class="flex items-center space-x-2">
        <span>{{ toastMessage }}</span>
      </div>
      <button @click="toastMessage = null" class="text-xs opacity-75 hover:opacity-100 font-bold">&times;</button>
    </div>

    <!-- TAB 1: ALARMS CONSOLE -->
    <div v-if="activeTab === 'alarms'" class="space-y-4">
      <div class="saas-card overflow-hidden bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-sm">
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
          <h2 class="text-sm font-bold text-slate-900 dark:text-white font-sans tracking-tight">
            {{ $t('alarms.activeAndHistorical') }}
          </h2>
          <button @click="refreshAlarms" class="text-2xs text-indigo-600 dark:text-indigo-400 hover:underline font-sans font-medium">
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
                <th class="py-3 px-4 w-40">{{ $t('alarms.triggeredAt') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.status') }}</th>
                <th class="py-3 px-4 w-28 text-center">{{ $t('alarms.action') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 text-slate-700 dark:text-slate-300">
              <tr v-if="alarms.length === 0">
                <td colspan="7" class="py-8 text-center text-slate-500 font-sans">
                  {{ $t('alarms.noActiveAlarms') }}
                </td>
              </tr>
              <tr v-for="alarm in alarms" :key="alarm.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/25 transition-colors">
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-bold font-mono tracking-wider"
                        :class="alarm.severity === 'CRITICAL' ? 'bg-rose-500/10 text-rose-500 dark:text-rose-400 border border-rose-500/20' : 'bg-amber-500/10 text-amber-500 dark:text-amber-400 border border-amber-500/20'">
                    {{ $t('alarms.' + alarm.severity.toLowerCase()) }}
                  </span>
                </td>
                <td class="py-3 px-4 text-slate-900 dark:text-slate-200 font-semibold font-mono text-2xs">{{ alarm.alarm_code }}</td>
                <td class="py-3 px-4 text-xs text-slate-700 dark:text-slate-100 font-sans">{{ alarm.message }}</td>
                <td class="py-3 px-4 text-indigo-600 dark:text-blue-400 font-mono text-2xs">
                  {{ alarm.device ? (alarm.device.device_code || alarm.device.code) : 'System' }}
                </td>
                <td class="py-3 px-4 text-slate-500 dark:text-slate-400 font-mono text-3xs">{{ alarm.triggered_at | formatDate }}</td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-semibold font-mono"
                        :class="alarm.status === 'ACTIVE' ? 'bg-rose-500/10 text-rose-500 dark:text-rose-400' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400'">
                    {{ alarm.status }}
                  </span>
                </td>
                <td class="py-3 px-4 text-center">
                  <button v-if="alarm.status === 'ACTIVE'"
                          @click="ackAlarm(alarm.id)"
                          class="px-2.5 py-1 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 border border-slate-300 dark:border-slate-700 text-slate-700 dark:text-slate-200 rounded-lg text-2xs font-sans transition-colors">
                    {{ $t('alarms.acknowledge') }}
                  </button>
                  <span v-else class="text-slate-400 text-3xs font-mono">{{ $t('alarms.acked') }}</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: NOTIFICATION DELIVERY QUEUE & HISTORY -->
    <div v-if="activeTab === 'queue'" class="space-y-4">
      <!-- Metrics Overview KPI Cards -->
      <div class="grid grid-cols-2 sm:grid-cols-5 gap-3">
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-semibold block">{{ $t('alarms.dispatched') }}</span>
          <span class="text-lg font-bold font-mono text-slate-900 dark:text-white mt-1 block">{{ metrics.dispatched_count || 0 }}</span>
        </div>
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-emerald-500 font-semibold block">{{ $t('alarms.delivered') }}</span>
          <span class="text-lg font-bold font-mono text-emerald-500 mt-1 block">{{ metrics.delivered_count || 0 }}</span>
        </div>
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-amber-500 font-semibold block">{{ $t('alarms.retrying') }}</span>
          <span class="text-lg font-bold font-mono text-amber-500 mt-1 block">{{ metrics.retried_count || 0 }}</span>
        </div>
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-rose-500 font-semibold block">{{ $t('alarms.failureReason') }}</span>
          <span class="text-lg font-bold font-mono text-rose-500 mt-1 block">{{ metrics.failed_count || 0 }}</span>
        </div>
        <div class="p-3.5 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider text-slate-400 font-semibold block">{{ $t('alarms.cancelled') }}</span>
          <span class="text-lg font-bold font-mono text-slate-400 mt-1 block">{{ metrics.cancelled_count || 0 }}</span>
        </div>
      </div>

      <!-- Filters & Action Bar -->
      <div class="p-4 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-2">
          <!-- Status Filter -->
          <select v-model="filterStatus" @change="fetchNotifications"
                  class="px-2.5 py-1.5 rounded-lg text-xs bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300">
            <option value="">{{ $t('alarms.filterStatus') }}: ALL</option>
            <option value="PENDING">PENDING</option>
            <option value="SENDING">SENDING</option>
            <option value="SENT">SENT</option>
            <option value="RETRYING">RETRYING</option>
            <option value="FAILED">FAILED</option>
            <option value="CANCELLED">CANCELLED</option>
          </select>

          <!-- Channel Filter -->
          <select v-model="filterChannel" @change="fetchNotifications"
                  class="px-2.5 py-1.5 rounded-lg text-xs bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300">
            <option value="">{{ $t('alarms.filterChannel') }}: ALL</option>
            <option value="WEB">WEB</option>
            <option value="EMAIL">EMAIL</option>
            <option value="TELEGRAM">TELEGRAM</option>
            <option value="WEBHOOK">WEBHOOK</option>
          </select>

          <!-- Search Input -->
          <input v-model="searchQuery" @input="debounceSearch"
                 :placeholder="$t('alarms.searchPlaceholder')"
                 class="px-3 py-1.5 rounded-lg text-xs bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 w-64" />
        </div>

        <button @click="fetchNotifications"
                class="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
          <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          {{ $t('common.refresh') }}
        </button>
      </div>

      <!-- Queue & History Table -->
      <div class="saas-card overflow-hidden bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs border-collapse font-sans">
            <thead>
              <tr class="border-b border-slate-200 dark:border-slate-800/80 bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                <th class="py-3 px-4 w-20">ID</th>
                <th class="py-3 px-4 w-24">{{ $t('alarms.channel') }}</th>
                <th class="py-3 px-4 w-32">{{ $t('alarms.recipient') }}</th>
                <th class="py-3 px-4 w-28">{{ $t('alarms.status') }}</th>
                <th class="py-3 px-4 w-20">{{ $t('alarms.attempts') }}</th>
                <th class="py-3 px-4">{{ $t('alarms.message') }}</th>
                <th class="py-3 px-4 w-36">{{ $t('alarms.nextRetry') }} / {{ $t('alarms.sentAt') }}</th>
                <th class="py-3 px-4 w-28 text-center">{{ $t('alarms.action') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/50 text-slate-700 dark:text-slate-300">
              <tr v-if="notifications.length === 0">
                <td colspan="8" class="py-8 text-center text-slate-500 font-sans">
                  {{ $t('alarms.noNotifications') }}
                </td>
              </tr>
              <tr v-for="notif in notifications" :key="notif.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/25 transition-colors">
                <td class="py-3 px-4 font-mono text-2xs text-slate-500">#{{ notif.id }}</td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded text-3xs font-bold font-mono tracking-wider"
                        :class="getChannelBadgeClass(notif.channel)">
                    {{ notif.channel }}
                  </span>
                </td>
                <td class="py-3 px-4 text-xs font-mono text-slate-600 dark:text-slate-300 truncate max-w-xs">{{ notif.recipient }}</td>
                <td class="py-3 px-4">
                  <span class="px-2 py-0.5 rounded text-3xs font-bold font-mono"
                        :class="getStatusBadgeClass(notif.status)">
                    {{ notif.status }}
                  </span>
                </td>
                <td class="py-3 px-4 font-mono text-2xs text-slate-600 dark:text-slate-300">
                  {{ notif.attempts }} / {{ notif.max_attempts }}
                </td>
                <td class="py-3 px-4">
                  <div class="font-semibold text-slate-900 dark:text-slate-100 text-xs">{{ notif.subject }}</div>
                  <div class="text-3xs text-slate-500 dark:text-slate-400 truncate max-w-sm">{{ notif.message }}</div>
                  <div v-if="notif.error_msg" class="text-3xs text-rose-500 dark:text-rose-400 truncate max-w-sm mt-0.5 font-mono">
                    {{ notif.error_msg }}
                  </div>
                </td>
                <td class="py-3 px-4 font-mono text-3xs text-slate-500 dark:text-slate-400">
                  <div v-if="notif.status === 'SENT'">{{ notif.sent_at | formatDate }}</div>
                  <div v-else-if="notif.status === 'RETRYING'">{{ notif.next_retry_at | formatDate }}</div>
                  <div v-else>{{ notif.created_at | formatDate }}</div>
                </td>
                <td class="py-3 px-4 text-center">
                  <div class="flex items-center justify-center space-x-1">
                    <!-- Inspect Details -->
                    <button @click="openDetailsModal(notif)"
                            class="p-1 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:text-white hover:bg-slate-700"
                            :title="$t('alarms.details')">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                    </button>
                    <!-- Retry Button -->
                    <button v-if="notif.status === 'FAILED' || notif.status === 'CANCELLED'"
                            @click="retryNotification(notif.id)"
                            class="p-1 rounded bg-indigo-50 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-600 hover:text-white transition-colors"
                            :title="$t('alarms.retry')">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                      </svg>
                    </button>
                    <!-- Cancel Button -->
                    <button v-if="notif.status === 'PENDING' || notif.status === 'RETRYING'"
                            @click="cancelNotification(notif.id)"
                            class="p-1 rounded bg-rose-50 dark:bg-rose-900/30 text-rose-600 dark:text-rose-400 hover:bg-rose-600 hover:text-white transition-colors"
                            :title="$t('alarms.cancelDelivery')">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800/80 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
          <button @click="changeQueuePage(queuePage - 1)"
                  :disabled="queuePage <= 1"
                  class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 disabled:opacity-40 text-2xs font-medium">
            {{ $t('telemetry.prev') || 'Previous' }}
          </button>
          <span class="font-mono text-2xs">{{ $t('telemetry.pageOf', { page: queuePage, totalPages: Math.ceil(queueTotal / queuePageSize) || 1 }) }}</span>
          <button @click="changeQueuePage(queuePage + 1)"
                  :disabled="queuePage >= Math.ceil(queueTotal / queuePageSize)"
                  class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 disabled:opacity-40 text-2xs font-medium">
            {{ $t('telemetry.next') || 'Next' }}
          </button>
        </div>
      </div>
    </div>

    <!-- TAB 3: NOTIFICATION CHANNELS & ROUTING -->
    <div v-if="activeTab === 'channels'" class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div v-for="ch in channels" :key="ch.id"
             class="p-5 rounded-2xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-sm flex flex-col justify-between space-y-4">
          <div class="space-y-3">
            <div class="flex items-start justify-between">
              <div class="flex items-center space-x-2.5">
                <span class="p-2 rounded-xl" :class="getChannelIconBg(ch.type)">
                  <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path v-if="ch.type === 'EMAIL'" stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                    <path v-else-if="ch.type === 'TELEGRAM'" stroke-linecap="round" stroke-linejoin="round" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
                    <path v-else-if="ch.type === 'WEBHOOK'" stroke-linecap="round" stroke-linejoin="round" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                    <path v-else stroke-linecap="round" stroke-linejoin="round" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                </span>
                <div>
                  <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">{{ ch.name }}</h3>
                  <span class="text-3xs font-mono text-slate-500">{{ ch.type }}</span>
                </div>
              </div>

              <!-- Channel Status Pill -->
              <span class="px-2 py-0.5 rounded-full text-3xs font-bold font-mono uppercase tracking-wider"
                    :class="ch.enabled ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20' : 'bg-slate-200 dark:bg-slate-800 text-slate-500'">
                {{ ch.enabled ? $t('alarms.enabled') : $t('alarms.disabled') }}
              </span>
            </div>

            <!-- Configuration Summary Badges -->
            <div class="grid grid-cols-2 gap-2 text-2xs font-mono">
              <div class="p-2 rounded-lg bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800">
                <span class="text-3xs text-slate-500 uppercase block font-sans">{{ $t('alarms.minSeverity') }}</span>
                <span class="font-bold text-slate-800 dark:text-slate-200">{{ ch.min_severity }}</span>
              </div>
              <div class="p-2 rounded-lg bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800">
                <span class="text-3xs text-slate-500 uppercase block font-sans">{{ $t('alarms.rateLimit') }}</span>
                <span class="font-bold text-slate-800 dark:text-slate-200">{{ ch.rate_limit_per_min }} / min</span>
              </div>
              <div class="p-2 rounded-lg bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800">
                <span class="text-3xs text-slate-500 uppercase block font-sans">{{ $t('alarms.quietHours') }}</span>
                <span class="font-bold text-slate-800 dark:text-slate-200">
                  {{ ch.quiet_hours_start ? `${ch.quiet_hours_start} - ${ch.quiet_hours_end}` : 'None' }}
                </span>
              </div>
              <div class="p-2 rounded-lg bg-slate-50 dark:bg-slate-900/50 border border-slate-100 dark:border-slate-800">
                <span class="text-3xs text-slate-500 uppercase block font-sans">Provider Status</span>
                <span class="font-bold" :class="ch.is_configured ? 'text-emerald-500' : 'text-amber-500'">
                  {{ ch.is_configured ? $t('alarms.configured') : $t('alarms.notConfigured') }}
                </span>
              </div>
            </div>

            <!-- Config Details (Sanitized Secrets) -->
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0B0F19]/50 border border-slate-100 dark:border-slate-800 text-3xs font-mono text-slate-500 dark:text-slate-400 space-y-1">
              <div v-for="(v, k) in ch.config" :key="k" class="flex justify-between">
                <span class="text-slate-400">{{ k }}:</span>
                <span class="text-slate-700 dark:text-slate-200 truncate max-w-[200px]">{{ v }}</span>
              </div>
            </div>
          </div>

          <!-- Channel Action Buttons -->
          <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-100 dark:border-slate-800/60">
            <button @click="openTestSendModal(ch)"
                    class="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs font-semibold transition-colors flex items-center">
              <svg class="w-3.5 h-3.5 mr-1 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              {{ $t('alarms.testSend') }}
            </button>
            <button @click="openEditChannelModal(ch)"
                    class="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
              <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              {{ $t('alarms.editChannel') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- MODAL 1: NOTIFICATION DETAILS MODAL -->
    <div v-if="showDetailsModal && selectedNotif"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-xl overflow-hidden shadow-2xl space-y-4 p-6">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
            {{ $t('alarms.notificationDetails') }} (#{{ selectedNotif.id }})
          </h3>
          <button @click="showDetailsModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">&times;</button>
        </div>

        <div class="space-y-3 text-xs font-sans">
          <div class="grid grid-cols-2 gap-3 font-mono">
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-800">
              <span class="text-3xs text-slate-500 uppercase block font-sans">Status</span>
              <span class="font-bold" :class="getStatusBadgeClass(selectedNotif.status)">{{ selectedNotif.status }}</span>
            </div>
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-800">
              <span class="text-3xs text-slate-500 uppercase block font-sans">Channel & Recipient</span>
              <span class="font-bold text-slate-900 dark:text-slate-200">{{ selectedNotif.channel }} : {{ selectedNotif.recipient }}</span>
            </div>
          </div>

          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-800 space-y-2">
            <div>
              <span class="text-3xs text-slate-500 uppercase block font-sans font-semibold">Subject</span>
              <span class="font-bold text-slate-900 dark:text-slate-100">{{ selectedNotif.subject }}</span>
            </div>
            <div>
              <span class="text-3xs text-slate-500 uppercase block font-sans font-semibold">Message Payload</span>
              <p class="text-slate-700 dark:text-slate-300 font-sans mt-0.5">{{ selectedNotif.message }}</p>
            </div>
          </div>

          <div v-if="selectedNotif.error_msg" class="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-500 dark:text-rose-400 space-y-1">
            <span class="text-3xs uppercase font-bold block">{{ $t('alarms.failureReason') }}</span>
            <p class="font-mono text-2xs">{{ selectedNotif.error_msg }}</p>
          </div>

          <div v-if="selectedNotif.provider_response" class="p-3 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-100 dark:border-slate-800 space-y-1 font-mono">
            <span class="text-3xs text-slate-500 uppercase block font-sans font-semibold">{{ $t('alarms.providerResponse') }}</span>
            <p class="text-2xs text-slate-700 dark:text-slate-300 break-all">{{ selectedNotif.provider_response }}</p>
          </div>
        </div>

        <div class="flex justify-end pt-3 border-t border-slate-200 dark:border-slate-800">
          <button @click="showDetailsModal = false" class="px-4 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
            {{ $t('alarms.close') }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 2: EDIT CHANNEL MODAL -->
    <div v-if="showEditModal && editingChannel"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl space-y-4 p-6">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
            {{ $t('alarms.editChannel') }}: {{ editingChannel.name }}
          </h3>
          <button @click="showEditModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">&times;</button>
        </div>

        <form @submit.prevent="saveChannelConfig" class="space-y-3 text-xs font-sans">
          <div class="flex items-center justify-between">
            <label class="text-slate-700 dark:text-slate-300 font-semibold">{{ $t('alarms.enabled') }}</label>
            <input type="checkbox" v-model="editForm.enabled" class="w-4 h-4 rounded text-indigo-600" />
          </div>

          <div>
            <label class="block text-slate-500 dark:text-slate-400 mb-1">Channel Name</label>
            <input v-model="editForm.name" required
                   class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white" />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.minSeverity') }}</label>
              <select v-model="editForm.min_severity"
                      class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white">
                <option value="INFO">INFO</option>
                <option value="WARNING">WARNING</option>
                <option value="CRITICAL">CRITICAL</option>
              </select>
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.rateLimit') }} (msg/min)</label>
              <input type="number" min="1" max="600" v-model.number="editForm.rate_limit_per_min" required
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.maxRetries') }}</label>
              <input type="number" min="0" max="10" v-model.number="editForm.max_retries" required
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.retryInterval') }}</label>
              <input type="number" min="5" max="3600" v-model.number="editForm.retry_interval_sec" required
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.quietHours') }} Start</label>
              <input type="text" placeholder="22:00" v-model="editForm.quiet_hours_start"
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('alarms.quietHours') }} End</label>
              <input type="text" placeholder="06:00" v-model="editForm.quiet_hours_end"
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono" />
            </div>
          </div>

          <!-- Dynamic Configuration Fields based on Type -->
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-900/60 border border-slate-200 dark:border-slate-800 space-y-2">
            <span class="text-3xs uppercase font-bold text-slate-500 block">Channel Provider Settings</span>

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

    <!-- MODAL 3: TEST SEND MODAL -->
    <div v-if="showTestModal && testingChannel"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-md overflow-hidden shadow-2xl space-y-4 p-6">
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
               :class="testResult.success ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-500' : 'bg-rose-500/10 border-rose-500/30 text-rose-500'">
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
      activeTab: 'alarms', // 'alarms', 'queue', 'channels'
      toastMessage: null,
      toastType: 'success',

      // Queue state
      notifications: [],
      queueTotal: 0,
      queuePage: 1,
      queuePageSize: 20,
      filterStatus: '',
      filterChannel: '',
      searchQuery: '',
      searchDebounceTimer: null,
      metrics: {},

      // Channels state
      channels: [],

      // Modals
      showDetailsModal: false,
      selectedNotif: null,

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
    activeCount() {
      return this.alarms.filter((a) => a.status === 'ACTIVE').length;
    },
  },
  mounted() {
    this.refreshAlarms();
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
    async ackAlarm(id) {
      try {
        await this.$store.dispatch('acknowledgeAlarm', id);
        this.showToast('Alarm acknowledged successfully');
      } catch (err) {
        this.showToast('Failed acknowledging alarm: ' + err.message, 'error');
      }
    },
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
    getStatusBadgeClass(status) {
      switch (status) {
        case 'SENT':
          return 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20';
        case 'SENDING':
          return 'bg-blue-500/10 text-blue-500 border border-blue-500/20';
        case 'RETRYING':
          return 'bg-amber-500/10 text-amber-500 border border-amber-500/20';
        case 'FAILED':
          return 'bg-rose-500/10 text-rose-500 border border-rose-500/20';
        case 'CANCELLED':
          return 'bg-slate-200 dark:bg-slate-800 text-slate-500 border border-slate-300 dark:border-slate-700';
        default:
          return 'bg-indigo-500/10 text-indigo-500 border border-indigo-500/20';
      }
    },
    getChannelBadgeClass(chType) {
      switch (chType) {
        case 'EMAIL':
          return 'bg-blue-500/10 text-blue-500 border border-blue-500/20';
        case 'TELEGRAM':
          return 'bg-cyan-500/10 text-cyan-500 border border-cyan-500/20';
        case 'WEBHOOK':
          return 'bg-purple-500/10 text-purple-500 border border-purple-500/20';
        default:
          return 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20';
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
