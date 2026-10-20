<template>
  <div class="space-y-4 p-3 sm:p-5 max-w-[1920px] mx-auto min-h-screen text-slate-800 dark:text-slate-100">
    <!-- Header & Breadcrumb (Sticky Top Navigation) -->
    <div class="sticky top-0 z-30 bg-slate-50/95 dark:bg-[#0B0F19]/95 backdrop-blur-md pt-2 pb-3.5 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 shadow-xs">
      <div class="flex items-center space-x-3">
        <router-link
          to="/dashboard"
          class="p-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700 transition"
          title="Back to Dashboard"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </router-link>
        <div>
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-400 animate-pulse"></span>
            <h1 class="text-lg sm:text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white uppercase">
              OUTBOUND INTEGRATION &amp; DELIVERY
            </h1>
            <span class="px-2 py-0.5 rounded text-4xs font-mono font-bold bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30">
              PHASE 8
            </span>
          </div>
          <p class="text-3xs text-slate-500 dark:text-slate-400 font-mono mt-0.5">
            Aggregation-based outbound routing, configurable JSON payloads, offline buffering &amp; queue inspection.
          </p>
        </div>
      </div>

      <!-- Navigation Tabs (Setup & Queue) -->
      <div class="flex items-center bg-slate-200/80 dark:bg-slate-800/90 p-1 rounded-xl border border-slate-300 dark:border-slate-700/80 font-mono text-xs shadow-inner">
        <!-- TAB 1: DESTINATIONS SETUP -->
        <button
          @click="selectTab('destinations')"
          class="px-3 py-1.5 rounded-lg transition font-semibold flex items-center space-x-2"
          :class="activeTab === 'destinations'
            ? 'bg-cyan-500 text-white shadow-xs'
            : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10" stroke-width="2"/>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
          </svg>
          <span>Destinations Setup ({{ destinations.length }})</span>
        </button>

        <!-- TAB 2: CONFIGURABLE JSON -->
        <button
          @click="selectTab('config')"
          class="px-3 py-1.5 rounded-lg transition font-semibold flex items-center space-x-2"
          :class="activeTab === 'config'
            ? 'bg-cyan-500 text-white shadow-xs'
            : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
          </svg>
          <span>Configurable JSON</span>
        </button>

        <!-- TAB 3: DELIVERY QUEUE -->
        <button
          @click="selectTab('queue')"
          class="px-3 py-1.5 rounded-lg transition font-semibold flex items-center space-x-2"
          :class="activeTab === 'queue'
            ? 'bg-cyan-500 text-white shadow-xs'
            : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
          </svg>
          <span>Delivery Queue</span>
          <span v-if="summary.pending > 0" class="ml-1 px-1.5 py-0.2 rounded-full text-4xs bg-amber-400 text-slate-900 font-bold">
            {{ summary.pending }}
          </span>
        </button>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- TAB 1: DELIVERY QUEUE & INSPECTION                             -->
    <!-- ============================================================== -->
    <div v-if="activeTab === 'queue'" class="space-y-4">
      <!-- Summary Counts KPI Cards -->
      <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-7 gap-2.5">
        <!-- PENDING -->
        <div class="p-3 rounded-xl border bg-amber-500/10 border-amber-500/30 text-amber-500">
          <div class="text-3xs font-mono font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Pending</span>
            <span class="w-2 h-2 rounded-full bg-amber-400 animate-pulse"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono mt-1 text-amber-600 dark:text-amber-400">
            {{ summary.pending || 0 }}
          </div>
          <div class="text-4xs text-slate-500 font-mono mt-0.5">Awaiting dispatch</div>
        </div>

        <!-- SENDING -->
        <div class="p-3 rounded-xl border bg-blue-500/10 border-blue-500/30 text-blue-500">
          <div class="text-3xs font-mono font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Sending</span>
            <span class="w-2 h-2 rounded-full bg-blue-400 animate-ping"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono mt-1 text-blue-600 dark:text-blue-400">
            {{ summary.sending || 0 }}
          </div>
          <div class="text-4xs text-slate-500 font-mono mt-0.5">In-flight transmission</div>
        </div>

        <!-- DELIVERED -->
        <div class="p-3 rounded-xl border bg-emerald-500/10 border-emerald-500/30 text-emerald-500">
          <div class="text-3xs font-mono font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Delivered</span>
            <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono mt-1 text-emerald-600 dark:text-emerald-400">
            {{ summary.delivered || 0 }}
          </div>
          <div class="text-4xs text-slate-500 font-mono mt-0.5">HTTP 2xx / Socket OK</div>
        </div>

        <!-- ACKNOWLEDGED -->
        <div class="p-3 rounded-xl border bg-purple-500/10 border-purple-500/30 text-purple-500">
          <div class="text-3xs font-mono font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Acknowledged</span>
            <span class="w-2 h-2 rounded-full bg-purple-400"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono mt-1 text-purple-600 dark:text-purple-400">
            {{ summary.acknowledged || 0 }}
          </div>
          <div class="text-4xs text-slate-500 font-mono mt-0.5">Upstream receipt confirmed</div>
        </div>

        <!-- RETRYING -->
        <div class="p-3 rounded-xl border bg-orange-500/10 border-orange-500/30 text-orange-500">
          <div class="text-3xs font-mono font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Retrying</span>
            <span class="w-2 h-2 rounded-full bg-orange-400"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono mt-1 text-orange-600 dark:text-orange-400">
            {{ summary.retrying || 0 }}
          </div>
          <div class="text-4xs text-slate-500 font-mono mt-0.5">Backoff & jitter active</div>
        </div>

        <!-- FAILED -->
        <div class="p-3 rounded-xl border bg-rose-500/10 border-rose-500/30 text-rose-500">
          <div class="text-3xs font-mono font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Failed</span>
            <span class="w-2 h-2 rounded-full bg-rose-400"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono mt-1 text-rose-600 dark:text-rose-400">
            {{ summary.failed || 0 }}
          </div>
          <div class="text-4xs text-slate-500 font-mono mt-0.5">Permanent / exhausted</div>
        </div>

        <!-- TOTAL -->
        <div class="p-3 rounded-xl border bg-slate-500/10 border-slate-500/30 text-slate-500 dark:text-slate-400">
          <div class="text-3xs font-mono font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Total Batches</span>
            <span>Σ</span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono mt-1 text-slate-900 dark:text-white">
            {{ summary.total || 0 }}
          </div>
          <div class="text-4xs text-slate-500 font-mono mt-0.5">Lifetime buffered</div>
        </div>
      </div>

      <!-- Search, Filter & Refresh Bar -->
      <div class="flex flex-wrap items-center justify-between gap-2.5 p-3 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 font-mono text-xs">
        <div class="flex flex-wrap items-center gap-2">
          <!-- Status Filter -->
          <div class="flex items-center space-x-1.5 bg-slate-100 dark:bg-slate-800 px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700">
            <span class="text-3xs text-slate-500">Status:</span>
            <select v-model="filterStatus" @change="fetchQueue" class="bg-transparent font-bold focus:outline-hidden cursor-pointer">
              <option value="ALL">ALL STATUSES</option>
              <option value="PENDING">PENDING</option>
              <option value="SENDING">SENDING</option>
              <option value="DELIVERED">DELIVERED</option>
              <option value="ACKNOWLEDGED">ACKNOWLEDGED</option>
              <option value="RETRYING">RETRYING</option>
              <option value="FAILED">FAILED</option>
              <option value="CANCELLED">CANCELLED</option>
            </select>
          </div>

          <!-- Destination Filter -->
          <div class="flex items-center space-x-1.5 bg-slate-100 dark:bg-slate-800 px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700">
            <span class="text-3xs text-slate-500">Destination:</span>
            <select v-model="filterDestinationId" @change="fetchQueue" class="bg-transparent font-bold focus:outline-hidden cursor-pointer">
              <option :value="null">ALL DESTINATIONS</option>
              <option v-for="d in destinations" :key="d.id" :value="d.id">
                {{ d.name }} ({{ d.type }})
              </option>
            </select>
          </div>

          <!-- Search Input -->
          <div class="relative">
            <input
              v-model="searchQuery"
              @keyup.enter="fetchQueue"
              type="text"
              placeholder="Search Batch ID / Key..."
              class="w-48 sm:w-64 bg-slate-100 dark:bg-slate-800 text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 focus:outline-hidden focus:border-cyan-500"
            />
          </div>
        </div>

        <div class="flex items-center space-x-2">
          <!-- Auto Refresh Toggle -->
          <button
            @click="autoRefresh = !autoRefresh"
            class="px-2.5 py-1.5 rounded-lg border text-3xs font-semibold flex items-center space-x-1.5 transition"
            :class="autoRefresh
              ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-400 border-cyan-500/50'
              : 'bg-slate-100 dark:bg-slate-800 text-slate-500 border-slate-200 dark:border-slate-700'"
          >
            <span class="w-1.5 h-1.5 rounded-full" :class="autoRefresh ? 'bg-cyan-400 animate-pulse' : 'bg-slate-400'"></span>
            <span>Auto Refresh (5s)</span>
          </button>

          <!-- Refresh Button -->
          <button
            @click="fetchQueue"
            :disabled="loading"
            class="px-3 py-1.5 rounded-lg bg-cyan-500 text-white hover:bg-cyan-600 transition flex items-center space-x-1.5 shadow-xs"
          >
            <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
            <span>Refresh</span>
          </button>
        </div>
      </div>

      <!-- Queue Items Table -->
      <div class="bg-white dark:bg-slate-900 rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden shadow-xs">
        <div class="overflow-x-auto">
          <table class="w-full text-left font-mono text-2xs">
            <thead class="bg-slate-50 dark:bg-slate-800/60 text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800 uppercase tracking-wider">
              <tr>
                <th class="py-2.5 px-3">Batch ID / Key</th>
                <th class="py-2.5 px-3">Devices</th>
                <th class="py-2.5 px-3">Destination</th>
                <th class="py-2.5 px-3">Window (UTC)</th>
                <th class="py-2.5 px-3 text-center">Interval</th>
                <th class="py-2.5 px-3 text-center">Params</th>
                <th class="py-2.5 px-3">Status</th>
                <th class="py-2.5 px-3 text-center">Attempts</th>
                <th class="py-2.5 px-3">Last Code / Error</th>
                <th class="py-2.5 px-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
              <tr v-if="queueItems.length === 0">
                <td colspan="10" class="py-12 text-center text-slate-400 font-mono">
                  <div class="text-sm">No outbound delivery queue items found</div>
                  <div class="text-3xs text-slate-500 mt-1">
                    Aggregation batches are automatically enqueued when an aggregation window completes.
                  </div>
                </td>
              </tr>
              <tr
                v-for="item in queueItems"
                :key="item.id"
                class="hover:bg-slate-50 dark:hover:bg-slate-800/40 transition"
              >
                <!-- Batch ID -->
                <td class="py-2.5 px-3">
                  <div class="font-bold text-slate-900 dark:text-white flex items-center space-x-1.5">
                    <span>{{ item.batch_id }}</span>
                    <span v-if="item.is_replay" class="px-1.5 py-0.2 rounded text-4xs bg-purple-500/20 text-purple-400 border border-purple-500/30">
                      REPLAY
                    </span>
                  </div>
                  <div class="text-4xs text-slate-400 truncate max-w-[200px]" :title="item.idempotency_key">
                    {{ item.idempotency_key }}
                  </div>
                </td>

                <!-- Devices (Unified Multi-Device Support) -->
                <td class="py-2.5 px-3">
                  <div v-if="item.device_count > 1 || (item.devices_summary && item.devices_summary.includes(','))">
                    <span class="px-2 py-0.5 rounded-md font-bold bg-cyan-500/15 text-cyan-500 dark:text-cyan-400 border border-cyan-500/30 text-3xs" :title="item.devices_summary">
                      {{ item.device_count || 2 }} Devices
                    </span>
                    <div class="text-4xs text-slate-400 truncate max-w-[150px] mt-0.5" :title="item.devices_summary">
                      {{ item.devices_summary }}
                    </div>
                  </div>
                  <span v-else class="px-2 py-0.5 rounded-md font-semibold bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 text-3xs">
                    {{ item.devices_summary || (item.device ? item.device.device_code : (item.device_id ? 'Station #' + item.device_id : 'All Devices')) }}
                  </span>
                </td>

                <!-- Destination -->
                <td class="py-2.5 px-3">
                  <div class="font-semibold text-slate-800 dark:text-slate-200">
                    {{ item.destination ? item.destination.name : ('Dest #' + item.destination_id) }}
                  </div>
                  <span class="px-1.5 py-0.2 rounded text-4xs font-bold" :class="protocolBadgeClass(item.destination ? item.destination.type : 'REST')">
                    {{ item.destination ? item.destination.type : 'REST_API' }}
                  </span>
                </td>

                <!-- Window -->
                <td class="py-2.5 px-3 text-3xs text-slate-500">
                  <div>{{ formatTime(item.period_start) }}</div>
                  <div>→ {{ formatTime(item.period_end) }}</div>
                </td>

                <!-- Interval -->
                <td class="py-2.5 px-3 text-center">
                  <span class="px-2 py-0.5 rounded-full text-3xs font-bold bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/20">
                    {{ Math.round(item.interval_seconds / 60) }}m
                  </span>
                </td>

                <!-- Params Count -->
                <td class="py-2.5 px-3 text-center font-bold">
                  {{ item.parameter_count }}
                </td>

                <!-- Status -->
                <td class="py-2.5 px-3">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-bold" :class="statusBadgeClass(item.status)">
                    {{ item.status }}
                  </span>
                  <div v-if="item.next_retry_at && item.status === 'RETRYING'" class="text-4xs text-orange-400 mt-0.5">
                    Retry at {{ formatTime(item.next_retry_at) }}
                  </div>
                </td>

                <!-- Attempts -->
                <td class="py-2.5 px-3 text-center">
                  <span :class="item.attempts >= item.max_retries ? 'text-rose-500 font-bold' : 'text-slate-600 dark:text-slate-300'">
                    {{ item.attempts }} / {{ item.max_retries }}
                  </span>
                </td>

                <!-- Last HTTP Code / Error -->
                <td class="py-2.5 px-3 max-w-[200px]">
                  <div v-if="item.last_http_code > 0" class="font-bold" :class="httpCodeClass(item.last_http_code)">
                    HTTP {{ item.last_http_code }}
                  </div>
                  <div v-if="item.last_error" class="text-4xs text-rose-400 truncate" :title="item.last_error">
                    {{ item.last_error }}
                  </div>
                  <div v-else-if="item.last_response" class="text-4xs text-emerald-500 truncate" :title="item.last_response">
                    {{ item.last_response }}
                  </div>
                </td>

                <!-- Actions -->
                <td class="py-2.5 px-3 text-right space-x-1 whitespace-nowrap">
                  <!-- Inspect Detail Modal -->
                  <button
                    @click="openInspectModal(item.id)"
                    class="px-2 py-1 rounded bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition text-3xs"
                    title="Deep Inspect: Source Aggregations vs Serialized JSON vs Response"
                  >
                    Inspect
                  </button>

                  <!-- Retry (if Retrying or Failed) -->
                  <button
                    v-if="canManageQueue && (item.status === 'FAILED' || item.status === 'RETRYING')"
                    @click="retryQueueItem(item.id)"
                    class="px-2 py-1 rounded bg-amber-500/20 hover:bg-amber-500/30 text-amber-500 border border-amber-500/30 transition text-3xs font-semibold"
                    title="Reset to PENDING for immediate retry"
                  >
                    Retry
                  </button>

                  <!-- Replay (Creates new delivery item) -->
                  <button
                    v-if="canManageQueue && item.status !== 'SENDING'"
                    @click="replayQueueItem(item.id)"
                    class="px-2 py-1 rounded bg-purple-500/20 hover:bg-purple-500/30 text-purple-400 border border-purple-500/30 transition text-3xs font-semibold"
                    title="Replay / Re-enqueue as a brand new delivery batch"
                  >
                    Replay
                  </button>

                  <!-- Cancel (if Pending or Retrying) -->
                  <button
                    v-if="canManageQueue && (item.status === 'PENDING' || item.status === 'RETRYING')"
                    @click="cancelQueueItem(item.id)"
                    class="px-2 py-1 rounded bg-rose-500/20 hover:bg-rose-500/30 text-rose-400 border border-rose-500/30 transition text-3xs"
                    title="Cancel pending delivery"
                  >
                    Cancel
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div class="px-4 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-2xs font-mono text-slate-500">
          <div>
            Showing {{ queueItems.length }} of {{ queueTotal }} items (Page {{ page }} of {{ totalPages }})
          </div>
          <div class="flex items-center space-x-1">
            <button
              @click="changePage(page - 1)"
              :disabled="page <= 1"
              class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 disabled:opacity-30 cursor-pointer"
            >
              Prev
            </button>
            <button
              @click="changePage(page + 1)"
              :disabled="page >= totalPages"
              class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 disabled:opacity-30 cursor-pointer"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- TAB 2: DESTINATIONS & PROTOCOLS                                -->
    <!-- ============================================================== -->
    <div v-if="activeTab === 'destinations'" class="space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h2 class="text-sm font-bold font-mono text-slate-900 dark:text-white uppercase">
            Upstream Outbound Destinations
          </h2>
          <p class="text-3xs text-slate-500 font-mono">
            Configure target servers, webhooks, brokers, and telemetry interval filters.
          </p>
        </div>
        <button
          v-if="canManageDestinations"
          @click="openAddDestinationModal"
          class="px-3 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-600 text-white font-mono text-xs font-bold transition flex items-center space-x-1.5 shadow-xs"
        >
          <span>+ Add Destination</span>
        </button>
      </div>

      <!-- Empty State when 0 Destinations -->
      <div
        v-if="destinations.length === 0"
        class="p-10 rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 text-center bg-white/40 dark:bg-slate-900/40 font-mono space-y-3"
      >
        <div class="w-12 h-12 rounded-full bg-cyan-500/10 text-cyan-500 flex items-center justify-center mx-auto">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10" stroke-width="2"/>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2 12h20M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
          </svg>
        </div>
        <h3 class="text-sm font-bold text-slate-800 dark:text-slate-200">
          No Upstream Destinations Configured
        </h3>
        <p class="text-2xs text-slate-500 max-w-md mx-auto">
          Configure downstream servers (REST Webhooks, MQTT Brokers, TCP/UDP Sockets) to automatically dispatch aggregated telemetry batches.
        </p>
        <div v-if="canManageDestinations" class="pt-2">
          <button
            @click="openAddDestinationModal"
            class="px-4 py-2 rounded-xl bg-cyan-500 hover:bg-cyan-600 text-white font-bold text-xs transition inline-flex items-center space-x-2 shadow-md cursor-pointer"
          >
            <span>+ Add Your First Destination</span>
          </button>
        </div>
        <div v-else class="text-3xs text-slate-400 italic">
          (Only Superadmin and Admin can add or configure destinations)
        </div>
      </div>

      <!-- Destination Cards Grid -->
      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
        <div
          v-for="dest in destinations"
          :key="dest.id"
          class="p-4 rounded-xl border bg-white dark:bg-slate-900 border-slate-200 dark:border-slate-800 flex flex-col justify-between shadow-xs transition hover:border-cyan-500/50"
        >
          <div>
            <div class="flex items-start justify-between">
              <div>
                <span class="px-2 py-0.5 rounded text-4xs font-mono font-bold" :class="protocolBadgeClass(dest.type)">
                  {{ dest.type }}
                </span>
                <h3 class="text-sm font-bold font-mono text-slate-900 dark:text-white mt-1.5">
                  {{ dest.name }}
                </h3>
              </div>
              <!-- Enable / Disable Switch -->
              <button
                v-if="canManageDestinations"
                @click="toggleDestinationEnabled(dest)"
                class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold border transition cursor-pointer"
                :class="dest.enabled
                  ? 'bg-emerald-500/20 text-emerald-400 border-emerald-500/30'
                  : 'bg-slate-200 dark:bg-slate-800 text-slate-500 border-slate-300 dark:border-slate-700'"
              >
                {{ dest.enabled ? '● ACTIVE' : '○ DISABLED' }}
              </button>
            </div>

            <!-- Details List -->
            <div class="mt-3 space-y-1.5 font-mono text-2xs text-slate-600 dark:text-slate-300">
              <div class="flex items-center justify-between">
                <span class="text-slate-400">Endpoint:</span>
                <span class="font-mono text-3xs truncate max-w-[200px]" :title="dest.endpoint">{{ dest.endpoint }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-400">Interval:</span>
                <span class="font-bold text-cyan-500">
                  {{ dest.interval_seconds > 0 ? (dest.interval_seconds / 60 + ' min') : 'All Intervals' }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-400">Auth Mode:</span>
                <span>{{ dest.auth_type || 'NONE' }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-400">Deliveries:</span>
                <span class="text-emerald-500 font-bold">{{ dest.success_deliveries }} OK</span>
                <span v-if="dest.failed_deliveries > 0" class="text-rose-500 font-bold ml-1">{{ dest.failed_deliveries }} ERR</span>
              </div>
            </div>
          </div>

          <!-- Card Actions -->
          <div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800/80 flex items-center justify-between font-mono text-3xs">
            <button
              @click="selectDestinationForConfig(dest.id)"
              class="text-cyan-500 hover:text-cyan-400 font-bold transition"
            >
              Configure JSON Schema →
            </button>
            <div class="space-x-1">
              <button
                @click="openEditDestinationModal(dest)"
                class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300"
              >
                Edit
              </button>
              <button
                v-if="canManageDestinations"
                @click="deleteDestination(dest.id)"
                class="px-2 py-1 rounded bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/30"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- TAB 3: CONFIGURABLE JSON & FIELD MAPPINGS                      -->
    <!-- ============================================================== -->
    <div v-if="activeTab === 'config'" class="space-y-4 font-mono">
      <!-- Empty State when 0 Destinations -->
      <div
        v-if="destinations.length === 0"
        class="p-10 rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 text-center bg-white/40 dark:bg-slate-900/40 space-y-3"
      >
        <h3 class="text-sm font-bold text-slate-800 dark:text-slate-200">
          No Destination Configured
        </h3>
        <p class="text-2xs text-slate-500 max-w-md mx-auto">
          Please add a destination in the "Destinations Setup" tab first before configuring payload formatting.
        </p>
        <button
          @click="selectTab('destinations')"
          class="mt-2 px-4 py-2 rounded-xl bg-cyan-500 hover:bg-cyan-600 text-white font-bold text-xs transition inline-flex items-center space-x-1.5 shadow-md cursor-pointer"
        >
          <span>→ Go to Destinations Setup</span>
        </button>
      </div>

      <div v-else class="space-y-4">
        <!-- Destination Selector Header -->
        <div class="flex flex-wrap items-center justify-between gap-3 p-3.5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
        <div class="flex items-center space-x-2">
          <span class="text-xs text-slate-500 font-bold">Active Destination:</span>
          <select
            v-model="selectedDestinationId"
            @change="loadDestinationPayloadConfig"
            class="bg-slate-100 dark:bg-slate-800 text-xs px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 font-bold text-slate-900 dark:text-white cursor-pointer focus:outline-hidden"
          >
            <option v-for="d in destinations" :key="d.id" :value="d.id">
              {{ d.name }} ({{ d.type }})
            </option>
          </select>
          <span class="text-3xs text-slate-400">Version #{{ payloadConfig.version || 1 }}</span>
        </div>

        <div class="flex items-center space-x-2">
          <!-- Test Payload Action -->
          <button
            @click="testPayloadTransmission"
            :disabled="testingPayload"
            class="px-3 py-1.5 rounded-lg bg-purple-500 hover:bg-purple-600 text-white text-xs font-bold transition flex items-center space-x-1.5 shadow-xs"
            title="Dispatch a test serialized payload directly to this destination"
          >
            <span v-if="testingPayload" class="animate-spin">⏳</span>
            <span>⚡ Test Payload Delivery</span>
          </button>

          <!-- Save Config Action -->
          <button
            v-if="canManageDestinations"
            @click="savePayloadConfig"
            :disabled="savingConfig"
            class="px-4 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-600 text-white text-xs font-bold transition flex items-center space-x-1.5 shadow-xs"
          >
            <span>💾 Save Configuration</span>
          </button>
        </div>
      </div>

      <!-- Test Result Alert Banner -->
      <div
        v-if="testResult"
        class="p-3.5 rounded-xl border text-2xs flex items-start justify-between"
        :class="testResult.success
          ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400'
          : 'bg-rose-500/10 border-rose-500/30 text-rose-400'"
      >
        <div>
          <div class="font-bold flex items-center space-x-2">
            <span>{{ testResult.success ? '✔ TEST DELIVERY SUCCEEDED' : '✖ TEST DELIVERY FAILED' }}</span>
            <span v-if="testResult.http_code">HTTP {{ testResult.http_code }}</span>
            <span>({{ testResult.latency_ms }}ms)</span>
          </div>
          <div class="mt-1 text-3xs opacity-90">{{ testResult.response_text || testResult.error_message }}</div>
        </div>
        <button @click="testResult = null" class="text-slate-400 hover:text-white">✕</button>
      </div>

      <!-- Config & Live Preview Columns -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <!-- Column 1: Schema Editor Form -->
        <div class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-4">
          <h3 class="text-xs font-bold uppercase tracking-wider text-cyan-500 border-b border-slate-200 dark:border-slate-800 pb-2">
            Payload Architecture & Mappings
          </h3>

          <!-- Template Preset -->
          <div class="grid grid-cols-2 gap-3 text-2xs">
            <div>
              <label class="block text-slate-500 mb-1">Template Preset</label>
              <select
                v-model="payloadConfig.template_type"
                @change="applyTemplatePreset"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              >
                <option value="STANDARD">Standard Datalogger</option>
                <option value="MULTI_DEVICE_FLAT">Multi-Device Unified Flat</option>
                <option value="MULTI_DEVICE_NESTED">Multi-Device Nested by Device</option>
                <option value="MULTI_DEVICE_ARRAY">Multi-Device Array of Devices</option>
                <option value="KLHK_SPARING">KLHK SPARING (Flat Standard)</option>
                <option value="NESTED_SCADA">Nested SCADA Telemetry</option>
                <option value="CUSTOM">Custom Schema</option>
              </select>
            </div>

            <!-- Structure Toggle -->
            <div>
              <label class="block text-slate-500 mb-1">JSON Structure</label>
              <select
                v-model="payloadConfig.structure"
                @change="triggerLivePreview"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-bold text-cyan-600 dark:text-cyan-400"
              >
                <option value="FLAT">FLAT (Root Key-Value)</option>
                <option value="NESTED">NESTED (Sub-Object under key)</option>
                <option value="NESTED_BY_DEVICE">NESTED_BY_DEVICE (Grouped by Device)</option>
                <option value="ARRAY_BY_DEVICE">ARRAY_BY_DEVICE (Array of Devices)</option>
              </select>
            </div>
          </div>

          <!-- Multi-Device Identity & Collision Handling -->
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800 space-y-3">
            <div class="flex items-center justify-between">
              <span class="font-bold text-slate-700 dark:text-slate-300 text-2xs uppercase tracking-wider">Multi-Device Identity & Conflicts</span>
              <label class="flex items-center space-x-1.5 cursor-pointer text-3xs font-semibold text-cyan-600 dark:text-cyan-400">
                <input
                  v-model="payloadConfig.include_device_identity"
                  @change="triggerLivePreview"
                  type="checkbox"
                  class="rounded border-slate-700 text-cyan-500 focus:ring-0"
                />
                <span>Include Device Identity in Payload</span>
              </label>
            </div>

            <div class="grid grid-cols-2 gap-3 text-2xs">
              <div>
                <label class="block text-slate-500 mb-1">Device Identity Mode</label>
                <select
                  v-model="payloadConfig.device_identity_mode"
                  @change="triggerLivePreview"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 text-3xs font-mono"
                >
                  <option value="NONE">NONE (Omit from external JSON)</option>
                  <option value="FIELD">FIELD (Top-level Station ID field)</option>
                  <option value="NESTED_BY_DEVICE">NESTED_BY_DEVICE (Keyed by Device Code)</option>
                  <option value="ARRAY_BY_DEVICE">ARRAY_BY_DEVICE (List of Device Records)</option>
                  <option value="PREFIX">PREFIX (Prefix keys: aqms01_pm25)</option>
                </select>
              </div>

              <div>
                <label class="block text-slate-500 mb-1">Parameter Collision Handling</label>
                <select
                  v-model="payloadConfig.conflict_resolution"
                  @change="triggerLivePreview"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 text-3xs font-mono font-bold text-amber-500"
                >
                  <option value="FAIL_ON_CONFLICT">FAIL_ON_CONFLICT (Validate & Error)</option>
                  <option value="PREFIX">PREFIX (Auto-prefix: aqms01_pm25)</option>
                  <option value="CUSTOM_MAPPING">CUSTOM_MAPPING (Explicit Aliases)</option>
                  <option value="ARRAY">ARRAY (Keep both in array [v1, v2])</option>
                  <option value="PRIORITY_PRIMARY">PRIORITY_PRIMARY (Primary device only)</option>
                  <option value="AVERAGE">AVERAGE (Cross-device mean average)</option>
                </select>
              </div>
            </div>
            <p class="text-4xs text-slate-400">
              * Internal aggregation ID, device ID, and parameter ID are always stored in DB audit records regardless of external JSON format.
            </p>
          </div>

          <!-- Nested Key Name (if NESTED) -->
          <div v-if="payloadConfig.structure === 'NESTED'" class="text-2xs">
            <label class="block text-slate-500 mb-1">Nested Data Key Name</label>
            <input
              v-model="payloadConfig.nested_data_key"
              @input="triggerLivePreview"
              type="text"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              placeholder="e.g. data, measurements, telemetry"
            />
          </div>

          <!-- Station ID & Override -->
          <div class="grid grid-cols-2 gap-3 text-2xs">
            <div>
              <label class="block text-slate-500 mb-1">Station ID Field Name</label>
              <input
                v-model="payloadConfig.station_id_field"
                @input="triggerLivePreview"
                type="text"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
                placeholder="station_id"
              />
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Station ID Value Override</label>
              <input
                v-model="payloadConfig.station_id_value"
                @input="triggerLivePreview"
                type="text"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
                placeholder="Auto (from Station Code)"
              />
            </div>
          </div>

          <!-- Timestamp Field & Format -->
          <div class="grid grid-cols-2 gap-3 text-2xs">
            <div>
              <label class="block text-slate-500 mb-1">Timestamp Field Name</label>
              <input
                v-model="payloadConfig.timestamp_field"
                @input="triggerLivePreview"
                type="text"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
                placeholder="timestamp"
              />
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Timestamp Format</label>
              <select
                v-model="payloadConfig.timestamp_format"
                @change="triggerLivePreview"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              >
                <option value="ISO8601">ISO 8601 / RFC 3339</option>
                <option value="UNIX_MS">UNIX Epoch Milliseconds</option>
                <option value="UNIX_SEC">UNIX Epoch Seconds</option>
                <option value="CUSTOM">Custom (YYYY-MM-DD HH:mm:ss)</option>
              </select>
            </div>
          </div>

          <!-- Interval Field Name -->
          <div class="text-2xs">
            <label class="block text-slate-500 mb-1">Aggregation Interval Field Name</label>
            <input
              v-model="payloadConfig.interval_field"
              @input="triggerLivePreview"
              type="text"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              placeholder="interval_minutes"
            />
          </div>

          <!-- Metadata & Quality Toggles -->
          <div class="flex flex-wrap items-center gap-4 text-2xs pt-2">
            <label class="flex items-center space-x-2 cursor-pointer">
              <input
                v-model="payloadConfig.include_units"
                @change="triggerLivePreview"
                type="checkbox"
                class="rounded border-slate-700 text-cyan-500 focus:ring-0"
              />
              <span>Include Parameter Units</span>
            </label>

            <label class="flex items-center space-x-2 cursor-pointer">
              <input
                v-model="payloadConfig.include_quality"
                @change="triggerLivePreview"
                type="checkbox"
                class="rounded border-slate-700 text-cyan-500 focus:ring-0"
              />
              <span>Include Quality Flags (GOOD/BAD)</span>
            </label>

            <label class="flex items-center space-x-2 cursor-pointer">
              <input
                v-model="payloadConfig.include_metadata"
                @change="triggerLivePreview"
                type="checkbox"
                class="rounded border-slate-700 text-cyan-500 focus:ring-0"
              />
              <span>Include Metadata Headers</span>
            </label>
          </div>

          <!-- Parameter Aliases Table -->
          <div class="space-y-2 pt-2 border-t border-slate-200 dark:border-slate-800">
            <div class="flex items-center justify-between text-2xs">
              <span class="font-bold text-slate-400">Parameter External Aliases & Filters</span>
              <span class="text-3xs text-slate-500">Maps internal code to downstream field name</span>
            </div>
            <div class="max-h-48 overflow-y-auto rounded-lg border border-slate-200 dark:border-slate-800">
              <table class="w-full text-left text-3xs">
                <thead class="bg-slate-50 dark:bg-slate-800/80 text-slate-500">
                  <tr>
                    <th class="p-2">Enabled</th>
                    <th class="p-2">Internal Code</th>
                    <th class="p-2">External Field Name</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr v-for="param in parameterMappings" :key="param.parameter_code">
                    <td class="p-2">
                      <input
                        v-model="param.enabled"
                        @change="triggerLivePreview"
                        type="checkbox"
                        class="rounded border-slate-700 text-cyan-500 focus:ring-0"
                      />
                    </td>
                    <td class="p-2 font-bold text-slate-700 dark:text-slate-300">
                      {{ param.parameter_code }} ({{ param.parameter_name }})
                    </td>
                    <td class="p-2">
                      <input
                        v-model="param.external_name"
                        @input="triggerLivePreview"
                        type="text"
                        class="w-full bg-slate-100 dark:bg-slate-800 px-2 py-1 rounded border border-slate-300 dark:border-slate-700 text-3xs"
                        :placeholder="param.parameter_code"
                      />
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- Column 2: Live Preview Card -->
        <div class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 flex flex-col justify-between">
          <div>
            <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-2">
              <div class="flex items-center space-x-2">
                <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
                <span class="text-xs font-bold uppercase tracking-wider text-emerald-500">
                  Live Serialized JSON Preview
                </span>
              </div>
              <button
                @click="copyToClipboard(previewJSON)"
                class="px-2.5 py-1 rounded-md text-3xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition"
              >
                📋 Copy JSON
              </button>
            </div>

            <!-- Code block -->
            <pre class="mt-3 p-3.5 rounded-lg bg-slate-950 text-emerald-400 font-mono text-2xs overflow-x-auto max-h-[500px] border border-slate-800 shadow-inner leading-relaxed select-all">{{ previewJSON || '// Generating live preview from real aggregation data...' }}</pre>
          </div>

          <div class="mt-4 pt-3 border-t border-slate-200 dark:border-slate-800 text-3xs text-slate-400 flex items-center justify-between">
            <span>● Matches actual serialized bytes dispatched by OutboundService</span>
            <span>UTF-8 JSON</span>
          </div>
        </div>
      </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- MODAL: DEEP QUEUE INSPECTION                                   -->
    <!-- ============================================================== -->
    <div
      v-if="showInspectModal && inspectDetail"
      class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-slate-950/80 backdrop-blur-xs font-mono"
    >
      <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-4xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
        <!-- Modal Header -->
        <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50 dark:bg-slate-800/40">
          <div>
            <div class="flex items-center space-x-2">
              <span class="px-2 py-0.5 rounded text-3xs font-bold" :class="statusBadgeClass(inspectDetail.status)">
                {{ inspectDetail.status }}
              </span>
              <h3 class="text-sm sm:text-base font-bold text-slate-900 dark:text-white">
                Queue Inspection: {{ inspectDetail.batch_id }}
              </h3>
            </div>
            <div class="text-3xs text-slate-400 mt-0.5">
              Idempotency Key: {{ inspectDetail.idempotency_key }}
            </div>
          </div>
          <button @click="showInspectModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
        </div>

        <!-- Modal Content Tabs -->
        <div class="p-5 overflow-y-auto space-y-5 text-2xs">
          <!-- SECTION 1: Source Aggregation Values -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <h4 class="font-bold text-cyan-500 uppercase tracking-wider flex items-center space-x-1.5">
                <span>1. Persisted Source Aggregation Values</span>
                <span class="text-3xs text-slate-400 font-normal">({{ inspectDetail.source_values ? inspectDetail.source_values.length : 0 }} records)</span>
              </h4>
              <span class="text-4xs text-slate-500">Source Type: CUSTOMER_PROCESSED</span>
            </div>
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
              <table class="w-full text-left text-3xs">
                <thead class="bg-slate-100 dark:bg-slate-800 text-slate-500 uppercase">
                  <tr>
                    <th class="p-2">Device</th>
                    <th class="p-2">Param Code</th>
                    <th class="p-2">Name</th>
                    <th class="p-2 text-right">Aggregated Value</th>
                    <th class="p-2">Unit</th>
                    <th class="p-2">Quality</th>
                    <th class="p-2">Source Agg ID</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr v-for="val in inspectDetail.source_values" :key="val.aggregation_result_id + '-' + val.parameter_id">
                    <td class="p-2 font-bold text-cyan-600 dark:text-cyan-400">
                      {{ val.device_code || ('Device #' + val.device_id) }}
                    </td>
                    <td class="p-2 font-bold text-slate-900 dark:text-white">{{ val.parameter_code }}</td>
                    <td class="p-2 text-slate-600 dark:text-slate-300">{{ val.parameter_name }}</td>
                    <td class="p-2 text-right font-bold" :class="val.value !== null ? 'text-cyan-500' : 'text-slate-400'">
                      {{ val.value !== null ? val.value : 'null' }}
                    </td>
                    <td class="p-2 text-slate-400">{{ val.unit || '-' }}</td>
                    <td class="p-2">
                      <span class="px-1.5 py-0.2 rounded font-bold" :class="val.quality === 'GOOD' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-rose-500/20 text-rose-400'">
                        {{ val.quality }}
                      </span>
                    </td>
                    <td class="p-2 text-slate-500">#{{ val.aggregation_result_id }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- SECTION 2: Exact Serialized Outbound Payload -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <h4 class="font-bold text-emerald-500 uppercase tracking-wider">
                2. Serialized Outbound Payload Snapshot (v{{ inspectDetail.payload_version }})
              </h4>
              <button
                @click="copyToClipboard(inspectDetail.payload_snapshot)"
                class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-300 hover:text-white text-3xs font-semibold"
              >
                📋 Copy Outbound JSON
              </button>
            </div>
            <pre class="p-3.5 rounded-xl bg-slate-950 text-emerald-400 text-3xs font-mono overflow-x-auto max-h-56 border border-slate-800 select-all leading-relaxed">{{ formatPrettyJSON(inspectDetail.payload_snapshot) }}</pre>
          </div>

          <!-- SECTION 3: Downstream Response / ACK -->
          <div>
            <h4 class="font-bold text-purple-400 uppercase tracking-wider mb-2">
              3. Destination Response / Acknowledgement
            </h4>
            <div class="p-3.5 rounded-xl bg-slate-100 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 space-y-2 text-3xs">
              <div class="flex items-center space-x-3">
                <span class="font-bold">HTTP Code:</span>
                <span class="font-bold" :class="httpCodeClass(inspectDetail.last_http_code)">
                  {{ inspectDetail.last_http_code || 'N/A' }}
                </span>
                <span class="text-slate-500">Last Attempt: {{ formatTime(inspectDetail.last_attempt_at) }}</span>
              </div>
              <div v-if="inspectDetail.last_response" class="text-emerald-400 font-mono">
                <div class="text-slate-400 mb-0.5">Response Body:</div>
                <div class="p-2 rounded bg-slate-900 border border-slate-800 break-all">{{ inspectDetail.last_response }}</div>
              </div>
              <div v-if="inspectDetail.last_error" class="text-rose-400 font-mono">
                <div class="text-slate-400 mb-0.5">Last Error:</div>
                <div class="p-2 rounded bg-rose-950/30 border border-rose-900/50 break-all">{{ inspectDetail.last_error }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50 dark:bg-slate-800/40 text-2xs">
          <div class="space-x-2">
            <button
              v-if="canManageQueue && (inspectDetail.status === 'FAILED' || inspectDetail.status === 'RETRYING')"
              @click="retryQueueItem(inspectDetail.id); showInspectModal = false;"
              class="px-3 py-1.5 rounded-lg bg-amber-500/20 text-amber-500 font-bold border border-amber-500/30"
            >
              Retry Item
            </button>
            <button
              v-if="canManageQueue"
              @click="replayQueueItem(inspectDetail.id); showInspectModal = false;"
              class="px-3 py-1.5 rounded-lg bg-purple-500/20 text-purple-400 font-bold border border-purple-500/30"
            >
              Replay as New Delivery
            </button>
          </div>
          <button
            @click="showInspectModal = false"
            class="px-4 py-1.5 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-white font-bold"
          >
            Close
          </button>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- MODAL: ADD / EDIT DESTINATION                                  -->
    <!-- ============================================================== -->
    <div
      v-if="showDestModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-slate-950/80 backdrop-blur-xs font-mono"
    >
      <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg p-5 space-y-4 shadow-2xl">
        <h3 class="text-sm font-bold text-slate-900 dark:text-white uppercase">
          {{ editingDest.id ? 'Edit Destination' : 'Add New Outbound Destination' }}
        </h3>

        <div class="space-y-3 text-2xs">
          <div>
            <label class="block text-slate-500 mb-1">Destination Name</label>
            <input
              v-model="editingDest.name"
              type="text"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              placeholder="e.g. KLHK SPARING Webhook"
            />
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 mb-1">Protocol Type</label>
              <select
                v-model="editingDest.type"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-bold"
              >
                <option value="REST_API">REST API (HTTP/S)</option>
                <option value="MQTT">MQTT Broker</option>
                <option value="TCP">Raw TCP Socket</option>
                <option value="UDP">UDP Datagram</option>
                <option value="WEBSOCKET">WebSocket Stream</option>
                <option value="CLOUD">Cloud Forwarder</option>
              </select>
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Interval Filter</label>
              <select
                v-model="editingDest.interval_seconds"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-bold"
              >
                <option :value="0">All Configured Intervals</option>
                <option :value="120">2 Minutes</option>
                <option :value="300">5 Minutes</option>
                <option :value="1800">30 Minutes</option>
                <option :value="3600">60 Minutes</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-slate-500 mb-1">Endpoint (URL or host:port)</label>
            <input
              v-model="editingDest.endpoint"
              type="text"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
              placeholder="https://sparing.menlhk.go.id/api/telemetry or 192.168.1.50:5000"
            />
          </div>

          <!-- Auth Settings -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 mb-1">Authentication Type</label>
              <select
                v-model="editingDest.auth_type"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              >
                <option value="NONE">None / Open</option>
                <option value="BEARER_TOKEN">Bearer Token</option>
                <option value="API_KEY">API Key Header</option>
                <option value="BASIC_AUTH">HTTP Basic Auth</option>
              </select>
            </div>
            <div v-if="editingDest.auth_type === 'BEARER_TOKEN'">
              <label class="block text-slate-500 mb-1">Bearer Token</label>
              <input
                v-model="editingDest.auth_token"
                type="password"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
                placeholder="***REDACTED***"
              />
            </div>
            <div v-if="editingDest.auth_type === 'API_KEY'">
              <label class="block text-slate-500 mb-1">API Key Value</label>
              <input
                v-model="editingDest.api_key_value"
                type="password"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
                placeholder="***REDACTED***"
              />
            </div>
          </div>

          <!-- Basic Auth Username & Password -->
          <div v-if="editingDest.auth_type === 'BASIC_AUTH'" class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 mb-1">Username</label>
              <input
                v-model="editingDest.username"
                type="text"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              />
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Password</label>
              <input
                v-model="editingDest.password"
                type="password"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
                placeholder="***REDACTED***"
              />
            </div>
          </div>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-3 border-t border-slate-200 dark:border-slate-800 text-2xs">
          <button
            @click="showDestModal = false"
            class="px-3 py-1.5 rounded-lg bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-300 font-bold"
          >
            Cancel
          </button>
          <button
            @click="saveDestination"
            class="px-4 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-600 text-white font-bold"
          >
            Save Destination
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'IntegrationView',
  data() {
    let initialTab = 'destinations';
    if (this.$route.query && this.$route.query.tab) {
      initialTab = this.$route.query.tab;
    } else if (this.$route.path && this.$route.path.endsWith('/queue')) {
      initialTab = 'queue';
    } else if (this.$route.path && this.$route.path.endsWith('/config')) {
      initialTab = 'config';
    }

    return {
      activeTab: initialTab,
      loading: false,
      autoRefresh: true,
      timer: null,

      // Queue state
      queueItems: [],
      queueTotal: 0,
      page: 1,
      pageSize: 20,
      filterStatus: 'ALL',
      filterDestinationId: null,
      searchQuery: '',
      summary: {
        pending: 0,
        sending: 0,
        delivered: 0,
        retrying: 0,
        failed: 0,
        acknowledged: 0,
        total: 0,
      },

      // Inspection Modal
      showInspectModal: false,
      inspectDetail: null,

      // Destinations state
      destinations: [],
      showDestModal: false,
      editingDest: {
        id: 0,
        name: '',
        type: 'REST_API',
        endpoint: '',
        interval_seconds: 0,
        auth_type: 'NONE',
        auth_token: '',
        api_key_header: 'X-API-Key',
        api_key_value: '',
        username: '',
        password: '',
        retry_limit: 5,
        enabled: true,
      },

      // Configurable Payload state
      selectedDestinationId: 1,
      payloadConfig: {
        destination_id: 1,
        template_type: 'STANDARD',
        station_id_field: 'station_id',
        station_id_value: '',
        timestamp_field: 'timestamp',
        timestamp_format: 'ISO8601',
        interval_field: 'interval_minutes',
        structure: 'FLAT',
        nested_data_key: 'data',
        include_device_identity: false,
        device_identity_mode: 'NONE',
        conflict_resolution: 'FAIL_ON_CONFLICT',
        include_units: false,
        include_quality: false,
        include_metadata: true,
        static_metadata_json: '',
        parameter_mappings_json: '[]',
        required_fields_json: '[]',
        version: 1,
      },
      parameterMappings: [
        { parameter_id: 1, parameter_code: 'pm25', parameter_name: 'PM2.5', external_name: 'PM25', enabled: true },
        { parameter_id: 2, parameter_code: 'pm10', parameter_name: 'PM10', external_name: 'PM10', enabled: true },
        { parameter_id: 3, parameter_code: 'so2', parameter_name: 'SO2', external_name: 'SO2', enabled: true },
        { parameter_id: 4, parameter_code: 'co', parameter_name: 'CO', external_name: 'CO', enabled: true },
        { parameter_id: 5, parameter_code: 'o3', parameter_name: 'O3', external_name: 'O3', enabled: true },
        { parameter_id: 6, parameter_code: 'no2', parameter_name: 'NO2', external_name: 'NO2', enabled: true },
      ],
      previewJSON: '',
      savingConfig: false,
      testingPayload: false,
      testResult: null,
    };
  },
  computed: {
    totalPages() {
      return Math.ceil(this.queueTotal / this.pageSize) || 1;
    },
    currentUserRole() {
      const u = this.$store.state.user || (this.$store.state.auth && this.$store.state.auth.user);
      const role = (u && u.role) || (this.$store.getters && this.$store.getters.userRole) || 'OPERATOR';
      return String(role).toUpperCase();
    },
    isSuperAdmin() {
      if (this.$store.getters && this.$store.getters.isSuperAdmin) return true;
      return this.currentUserRole === 'SUPERADMIN' || this.currentUserRole === 'ADMINISTRATOR';
    },
    isAdmin() {
      if (this.$store.getters && this.$store.getters.isAdmin) return true;
      return this.currentUserRole === 'ADMIN' || this.isSuperAdmin;
    },
    isEngineer() {
      if (this.$store.getters && this.$store.getters.isEngineer) return true;
      return this.currentUserRole === 'ENGINEER' || this.isAdmin;
    },
    isOperator() {
      return !this.isAdmin && !this.isEngineer;
    },
    canManageDestinations() {
      // Khusus untuk SUPERADMIN dan ADMIN
      return this.isSuperAdmin || this.isAdmin;
    },
    canManageQueue() {
      // Khusus untuk SUPERADMIN dan ADMIN
      return this.isSuperAdmin || this.isAdmin;
    },
  },
  watch: {
    '$route.path'(newPath) {
      if (newPath.endsWith('/queue')) {
        this.activeTab = 'queue';
      } else if (newPath === '/integration' && this.activeTab === 'queue') {
        this.activeTab = 'destinations';
      }
    },
    '$route.query.tab'(newTab) {
      if (newTab && ['destinations', 'config', 'queue'].includes(newTab)) {
        this.activeTab = newTab;
      }
    },
  },
  mounted() {
    this.$nextTick(() => {
      const main = document.querySelector('main');
      if (main) main.scrollTop = 0;
      window.scrollTo(0, 0);
    });

    this.fetchSummary();
    this.fetchDestinations();
    this.fetchQueue();

    this.timer = setInterval(() => {
      if (this.autoRefresh && this.activeTab === 'queue') {
        this.fetchSummary();
        this.fetchQueue();
      }
    }, 5000);
  },
  beforeDestroy() {
    if (this.timer) {
      clearInterval(this.timer);
    }
  },
  methods: {
    selectTab(tab) {
      this.activeTab = tab;
      if (this.$route.query.tab !== tab) {
        this.$router.replace({ query: { ...this.$route.query, tab } }).catch(() => {});
      }
    },
    async fetchSummary() {
      try {
        const res = await axios.get('/api/integration/queue/summary');
        if (res.data && res.data.success) {
          this.summary = res.data.data;
        }
      } catch (err) {
        // silent fail
      }
    },
    async fetchDestinations() {
      try {
        const res = await axios.get('/api/destinations');
        if (res.data && res.data.success) {
          this.destinations = res.data.data;
          if (this.destinations.length > 0 && !this.selectedDestinationId) {
            this.selectedDestinationId = this.destinations[0].id;
          }
        }
      } catch (err) {
        console.error('Failed to load destinations:', err);
      }
    },
    async fetchQueue() {
      this.loading = true;
      try {
        const params = {
          page: this.page,
          page_size: this.pageSize,
        };
        if (this.filterStatus && this.filterStatus !== 'ALL') {
          params.status = this.filterStatus;
        }
        if (this.filterDestinationId) {
          params.destination_id = this.filterDestinationId;
        }
        if (this.searchQuery) {
          params.search = this.searchQuery;
        }

        const res = await axios.get('/api/integration/queue', { params });
        if (res.data && res.data.success) {
          this.queueItems = res.data.data.items || [];
          this.queueTotal = res.data.data.total || 0;
        }
      } catch (err) {
        console.error('Failed to load queue:', err);
      } finally {
        this.loading = false;
      }
    },
    changePage(p) {
      if (p < 1 || p > this.totalPages) return;
      this.page = p;
      this.fetchQueue();
    },
    async openInspectModal(id) {
      try {
        const res = await axios.get(`/api/integration/queue/${id}`);
        if (res.data && res.data.success) {
          this.inspectDetail = res.data.data;
          this.showInspectModal = true;
        }
      } catch (err) {
        alert('Failed to load queue detail: ' + err.message);
      }
    },
    async retryQueueItem(id) {
      try {
        await axios.post(`/api/integration/queue/${id}/retry`);
        this.fetchSummary();
        this.fetchQueue();
      } catch (err) {
        alert('Retry failed: ' + err.message);
      }
    },
    async replayQueueItem(id) {
      if (!confirm('Replay this delivery? A new queue item will be created.')) return;
      try {
        await axios.post(`/api/integration/queue/${id}/replay`);
        this.fetchSummary();
        this.fetchQueue();
      } catch (err) {
        alert('Replay failed: ' + err.message);
      }
    },
    async cancelQueueItem(id) {
      if (!confirm('Cancel this pending delivery item?')) return;
      try {
        await axios.post(`/api/integration/queue/${id}/cancel`);
        this.fetchSummary();
        this.fetchQueue();
      } catch (err) {
        alert('Cancel failed: ' + err.message);
      }
    },
    selectDestinationForConfig(destId) {
      this.selectedDestinationId = destId;
      this.activeTab = 'config';
      this.loadDestinationPayloadConfig();
    },
    async loadDestinationPayloadConfig() {
      if (!this.selectedDestinationId) return;
      try {
        const res = await axios.get(`/api/destinations/${this.selectedDestinationId}/payload-config`);
        if (res.data && res.data.success) {
          this.payloadConfig = res.data.data;
          if (this.payloadConfig.parameter_mappings_json) {
            try {
              const parsed = JSON.parse(this.payloadConfig.parameter_mappings_json);
              if (Array.isArray(parsed) && parsed.length > 0) {
                this.parameterMappings = parsed;
              }
            } catch (e) {
              // fallback
            }
          }
        }
        this.triggerLivePreview();
      } catch (err) {
        console.error('Failed to load payload config:', err);
      }
    },
    async triggerLivePreview() {
      if (!this.selectedDestinationId) return;
      try {
        this.payloadConfig.parameter_mappings_json = JSON.stringify(this.parameterMappings);
        const res = await axios.post(`/api/destinations/${this.selectedDestinationId}/preview`, this.payloadConfig);
        if (res.data && res.data.success) {
          this.previewJSON = res.data.data.preview;
        }
      } catch (err) {
        this.previewJSON = '// Preview validation error:\n' + (err.response?.data?.error || err.message);
      }
    },
    async savePayloadConfig() {
      this.savingConfig = true;
      try {
        this.payloadConfig.parameter_mappings_json = JSON.stringify(this.parameterMappings);
        const res = await axios.put(`/api/destinations/${this.selectedDestinationId}/payload-config`, this.payloadConfig);
        if (res.data && res.data.success) {
          this.payloadConfig = res.data.data;
          alert('Configuration version #' + this.payloadConfig.version + ' saved successfully!');
          this.triggerLivePreview();
        }
      } catch (err) {
        alert('Validation error: ' + (err.response?.data?.error || err.message));
      } finally {
        this.savingConfig = false;
      }
    },
    async testPayloadTransmission() {
      this.testingPayload = true;
      this.testResult = null;
      try {
        const res = await axios.post(`/api/destinations/${this.selectedDestinationId}/test`);
        if (res.data && res.data.success) {
          this.testResult = res.data.data;
        }
      } catch (err) {
        this.testResult = {
          success: false,
          error_message: err.response?.data?.error || err.message,
          latency_ms: 0,
        };
      } finally {
        this.testingPayload = false;
      }
    },
    openAddDestinationModal() {
      this.editingDest = {
        id: 0,
        name: '',
        type: 'REST_API',
        endpoint: '',
        interval_seconds: 0,
        auth_type: 'NONE',
        auth_token: '',
        api_key_header: 'X-API-Key',
        api_key_value: '',
        username: '',
        password: '',
        retry_limit: 5,
        enabled: true,
      };
      this.showDestModal = true;
    },
    openEditDestinationModal(dest) {
      this.editingDest = JSON.parse(JSON.stringify(dest));
      this.showDestModal = true;
    },
    async saveDestination() {
      try {
        if (this.editingDest.id) {
          await axios.put(`/api/destinations/${this.editingDest.id}`, this.editingDest);
        } else {
          await axios.post('/api/destinations', this.editingDest);
        }
        this.showDestModal = false;
        this.fetchDestinations();
      } catch (err) {
        alert('Failed to save destination: ' + (err.response?.data?.error || err.message));
      }
    },
    async toggleDestinationEnabled(dest) {
      try {
        const updated = { ...dest, enabled: !dest.enabled };
        await axios.put(`/api/destinations/${dest.id}`, updated);
        dest.enabled = updated.enabled;
      } catch (err) {
        alert('Toggle failed: ' + err.message);
      }
    },
    async deleteDestination(id) {
      if (!confirm('Delete this destination? Queued deliveries will be detached.')) return;
      try {
        await axios.delete(`/api/destinations/${id}`);
        this.fetchDestinations();
      } catch (err) {
        alert('Delete failed: ' + err.message);
      }
    },
    applyTemplatePreset() {
      const p = this.payloadConfig.template_type;
      if (p === 'MULTI_DEVICE_FLAT') {
        this.payloadConfig.station_id_field = 'station_id';
        this.payloadConfig.timestamp_field = 'timestamp';
        this.payloadConfig.timestamp_format = 'ISO8601';
        this.payloadConfig.interval_field = 'interval_minutes';
        this.payloadConfig.structure = 'FLAT';
        this.payloadConfig.nested_data_key = 'data';
        this.payloadConfig.include_device_identity = false;
        this.payloadConfig.device_identity_mode = 'NONE';
        this.payloadConfig.conflict_resolution = 'PREFIX';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
      } else if (p === 'MULTI_DEVICE_NESTED') {
        this.payloadConfig.station_id_field = 'station_id';
        this.payloadConfig.timestamp_field = 'timestamp';
        this.payloadConfig.timestamp_format = 'ISO8601';
        this.payloadConfig.interval_field = 'interval_minutes';
        this.payloadConfig.structure = 'NESTED_BY_DEVICE';
        this.payloadConfig.nested_data_key = 'data';
        this.payloadConfig.include_device_identity = true;
        this.payloadConfig.device_identity_mode = 'NESTED_BY_DEVICE';
        this.payloadConfig.conflict_resolution = 'FAIL_ON_CONFLICT';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
      } else if (p === 'MULTI_DEVICE_ARRAY') {
        this.payloadConfig.station_id_field = 'station_id';
        this.payloadConfig.timestamp_field = 'timestamp';
        this.payloadConfig.timestamp_format = 'ISO8601';
        this.payloadConfig.interval_field = 'interval_minutes';
        this.payloadConfig.structure = 'ARRAY_BY_DEVICE';
        this.payloadConfig.nested_data_key = 'devices';
        this.payloadConfig.include_device_identity = true;
        this.payloadConfig.device_identity_mode = 'ARRAY_BY_DEVICE';
        this.payloadConfig.conflict_resolution = 'FAIL_ON_CONFLICT';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
      } else if (p === 'KLHK_SPARING') {
        this.payloadConfig.station_id_field = 'station_id';
        this.payloadConfig.timestamp_field = 'timestamp';
        this.payloadConfig.timestamp_format = 'ISO8601';
        this.payloadConfig.interval_field = 'interval_minutes';
        this.payloadConfig.structure = 'FLAT';
        this.payloadConfig.include_device_identity = false;
        this.payloadConfig.device_identity_mode = 'NONE';
        this.payloadConfig.conflict_resolution = 'PREFIX';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
      } else if (p === 'NESTED_SCADA') {
        this.payloadConfig.station_id_field = 'stationCode';
        this.payloadConfig.timestamp_field = 'datetime';
        this.payloadConfig.timestamp_format = 'UNIX_MS';
        this.payloadConfig.interval_field = 'periodMin';
        this.payloadConfig.structure = 'NESTED';
        this.payloadConfig.nested_data_key = 'measurements';
        this.payloadConfig.include_device_identity = true;
        this.payloadConfig.device_identity_mode = 'FIELD';
        this.payloadConfig.conflict_resolution = 'PREFIX';
        this.payloadConfig.include_units = true;
        this.payloadConfig.include_quality = true;
      } else {
        this.payloadConfig.structure = 'FLAT';
      }
      this.triggerLivePreview();
    },
    formatTime(t) {
      if (!t) return '-';
      try {
        const d = new Date(t);
        return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
      } catch (e) {
        return t;
      }
    },
    formatPrettyJSON(str) {
      if (!str) return '';
      try {
        const parsed = JSON.parse(str);
        return JSON.stringify(parsed, null, 2);
      } catch (e) {
        return str;
      }
    },
    copyToClipboard(text) {
      if (!text) return;
      navigator.clipboard.writeText(text);
      alert('Copied to clipboard!');
    },
    statusBadgeClass(st) {
      switch (st) {
        case 'DELIVERED':
          return 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30';
        case 'ACKNOWLEDGED':
          return 'bg-purple-500/20 text-purple-400 border border-purple-500/30';
        case 'SENDING':
          return 'bg-blue-500/20 text-blue-400 border border-blue-500/30 animate-pulse';
        case 'RETRYING':
          return 'bg-orange-500/20 text-orange-400 border border-orange-500/30';
        case 'FAILED':
          return 'bg-rose-500/20 text-rose-400 border border-rose-500/30';
        case 'CANCELLED':
          return 'bg-slate-500/20 text-slate-400 border border-slate-500/30';
        case 'PENDING':
        default:
          return 'bg-amber-500/20 text-amber-400 border border-amber-500/30';
      }
    },
    protocolBadgeClass(type) {
      switch (type) {
        case 'REST_API':
          return 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/30';
        case 'MQTT':
          return 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30';
        case 'TCP':
        case 'UDP':
          return 'bg-amber-500/20 text-amber-400 border border-amber-500/30';
        case 'WEBSOCKET':
          return 'bg-purple-500/20 text-purple-400 border border-purple-500/30';
        default:
          return 'bg-slate-500/20 text-slate-400 border border-slate-500/30';
      }
    },
    httpCodeClass(code) {
      if (code >= 200 && code < 300) return 'text-emerald-500';
      if (code >= 400 && code < 500) return 'text-amber-500';
      if (code >= 500) return 'text-rose-500';
      return 'text-slate-400';
    },
  },
};
</script>
