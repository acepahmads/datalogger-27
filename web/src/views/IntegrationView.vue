<template>
  <div class="space-y-4 p-3 sm:p-5 max-w-[1920px] mx-auto min-h-screen text-slate-800 dark:text-slate-100 font-sans">
    <!-- Header & Breadcrumb (Sticky Top Navigation) -->
    <div class="sticky top-0 z-30 bg-slate-50/95 dark:bg-[#0B0F19]/95 backdrop-blur-md pt-2 pb-3.5 border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 shadow-xs">
      <div class="flex items-center space-x-3">
        <router-link
          :to="isSuperAdmin ? '/dashboard' : '/monitoring/operator'"
          class="p-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700 transition"
          :title="isSuperAdmin ? 'Back to Dashboard' : 'Back to Operator Console'"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </router-link>
        <div>
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-400 animate-pulse"></span>
            <h1 class="text-lg sm:text-xl font-bold tracking-tight text-slate-900 dark:text-white uppercase">
              OUTBOUND INTEGRATION &amp; DELIVERY
            </h1>
            <span class="px-2 py-0.5 rounded text-4xs font-mono font-bold bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30">
              PHASE 8
            </span>
          </div>
          <p class="text-2xs text-slate-500 dark:text-slate-400 mt-0.5">
            Aggregation-based outbound routing, configurable JSON payloads, offline buffering &amp; queue inspection.
          </p>
        </div>
      </div>

      <!-- Navigation Tabs (Setup & Queue) -->
      <div class="flex items-center bg-slate-200/80 dark:bg-slate-800/90 p-1 rounded-xl border border-slate-300 dark:border-slate-700/80 text-xs shadow-inner">
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
          <span v-if="summary.pending > 0" class="ml-1 px-1.5 py-0.2 rounded-full text-4xs font-mono tabular-nums bg-amber-400 text-slate-900 font-bold">
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
          <div class="text-3xs font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Pending</span>
            <span class="w-2 h-2 rounded-full bg-amber-400 animate-pulse"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono tabular-nums mt-1 text-amber-600 dark:text-amber-400">
            {{ summary.pending || 0 }}
          </div>
          <div class="text-4xs text-slate-500 mt-0.5">Awaiting dispatch</div>
        </div>

        <!-- SENDING -->
        <div class="p-3 rounded-xl border bg-blue-500/10 border-blue-500/30 text-blue-500">
          <div class="text-3xs font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Sending</span>
            <span class="w-2 h-2 rounded-full bg-blue-400 animate-ping"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono tabular-nums mt-1 text-blue-600 dark:text-blue-400">
            {{ summary.sending || 0 }}
          </div>
          <div class="text-4xs text-slate-500 mt-0.5">In-flight transmission</div>
        </div>

        <!-- DELIVERED -->
        <div class="p-3 rounded-xl border bg-emerald-500/10 border-emerald-500/30 text-emerald-500">
          <div class="text-3xs font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Delivered</span>
            <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono tabular-nums mt-1 text-emerald-600 dark:text-emerald-400">
            {{ summary.delivered || 0 }}
          </div>
          <div class="text-4xs text-slate-500 mt-0.5">HTTP 2xx / Socket OK</div>
        </div>

        <!-- ACKNOWLEDGED -->
        <div class="p-3 rounded-xl border bg-purple-500/10 border-purple-500/30 text-purple-500">
          <div class="text-3xs font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Acknowledged</span>
            <span class="w-2 h-2 rounded-full bg-purple-400"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono tabular-nums mt-1 text-purple-600 dark:text-purple-400">
            {{ summary.acknowledged || 0 }}
          </div>
          <div class="text-4xs text-slate-500 mt-0.5">Upstream receipt confirmed</div>
        </div>

        <!-- RETRYING -->
        <div class="p-3 rounded-xl border bg-orange-500/10 border-orange-500/30 text-orange-500">
          <div class="text-3xs font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Retrying</span>
            <span class="w-2 h-2 rounded-full bg-orange-400"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono tabular-nums mt-1 text-orange-600 dark:text-orange-400">
            {{ summary.retrying || 0 }}
          </div>
          <div class="text-4xs text-slate-500 mt-0.5">Backoff & jitter active</div>
        </div>

        <!-- FAILED -->
        <div class="p-3 rounded-xl border bg-rose-500/10 border-rose-500/30 text-rose-500">
          <div class="text-3xs font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Failed</span>
            <span class="w-2 h-2 rounded-full bg-rose-400"></span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono tabular-nums mt-1 text-rose-600 dark:text-rose-400">
            {{ summary.failed || 0 }}
          </div>
          <div class="text-4xs text-slate-500 mt-0.5">Permanent / exhausted</div>
        </div>

        <!-- TOTAL -->
        <div class="p-3 rounded-xl border bg-slate-500/10 border-slate-500/30 text-slate-500 dark:text-slate-400">
          <div class="text-3xs font-semibold uppercase tracking-wider flex items-center justify-between">
            <span>Total Batches</span>
            <span>Σ</span>
          </div>
          <div class="text-xl sm:text-2xl font-bold font-mono tabular-nums mt-1 text-slate-900 dark:text-white">
            {{ summary.total || 0 }}
          </div>
          <div class="text-4xs text-slate-500 mt-0.5">Lifetime buffered</div>
        </div>
      </div>

      <!-- Search, Filter & Refresh Bar -->
      <div class="flex flex-wrap items-center justify-between gap-2.5 p-3 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-xs">
        <div class="flex flex-wrap items-center gap-2">
          <!-- Status Filter -->
          <div class="flex items-center space-x-1.5 bg-slate-100 dark:bg-slate-800 px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700">
            <span class="text-3xs text-slate-500">Status:</span>
            <select v-model="filterStatus" @change="onQueueFilterChange" class="bg-transparent font-bold focus:outline-hidden cursor-pointer">
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
            <select v-model="filterDestinationId" @change="onQueueFilterChange" class="bg-transparent font-bold focus:outline-hidden cursor-pointer">
              <option :value="null">ALL DESTINATIONS</option>
              <option v-for="d in destinations" :key="d.id" :value="d.id">
                {{ d.name }} ({{ d.type }})
              </option>
            </select>
          </div>

          <!-- Window Date Filter -->
          <div class="flex items-center space-x-1.5 bg-slate-100 dark:bg-slate-800 px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700">
            <span class="text-3xs text-slate-500">Tanggal Data:</span>
            <input
              type="date"
              v-model="filterWindowDate"
              @change="onQueueFilterChange"
              class="bg-transparent font-mono font-bold text-xs focus:outline-hidden cursor-pointer"
            />
            <button
              v-if="filterWindowDate"
              type="button"
              @click="filterWindowDate = ''; onQueueFilterChange()"
              class="text-slate-400 hover:text-rose-400 font-mono text-3xs px-1"
              title="Reset filter tanggal data"
            >
              ✕
            </button>
          </div>

          <!-- Search Input -->
          <div class="relative">
            <input
              v-model="searchQuery"
              @keyup.enter="onQueueFilterChange"
              type="text"
              placeholder="Search Batch ID / Device / Key..."
              class="w-48 sm:w-64 bg-slate-100 dark:bg-slate-800 text-xs px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 focus:outline-hidden focus:border-cyan-500"
            />
          </div>
        </div>

        <div class="flex items-center space-x-2">
          <!-- Purge Queue Button (Admin / Superadmin only) -->
          <button
            v-if="canManageQueue"
            @click="openPurgeModal"
            class="px-2.5 py-1.5 rounded-lg border text-3xs font-semibold flex items-center space-x-1.5 transition bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 border-rose-500/30"
            title="Bersihkan / Purge antrean pengiriman secara manual"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            <span>Purge Antrean</span>
          </button>

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
          <table class="w-full text-left text-2xs">
            <thead class="bg-slate-50 dark:bg-slate-800/60 text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800 uppercase tracking-wider font-semibold">
              <tr>
                <th class="py-2.5 px-3">Batch ID / Key</th>
                <th v-if="!isOperator" class="py-2.5 px-3">Devices</th>
                <th class="py-2.5 px-3">Destination</th>
                <th class="py-2.5 px-3">Waktu Pertama Kirim</th>
                <th class="py-2.5 px-3">Window Data</th>
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
                <td :colspan="isOperator ? 10 : 11" class="py-12 text-center text-slate-400">
                  <div class="text-sm font-medium">No outbound delivery queue items found</div>
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
                <td class="py-2.5 px-3 font-mono">
                  <div class="font-bold text-slate-900 dark:text-white flex items-center space-x-1.5">
                    <span>{{ item.batch_id }}</span>
                    <span v-if="item.is_replay" class="px-1.5 py-0.2 rounded text-4xs font-sans bg-purple-500/20 text-purple-400 border border-purple-500/30">
                      REPLAY
                    </span>
                  </div>
                  <div class="text-4xs text-slate-400 truncate max-w-[200px]" :title="item.idempotency_key">
                    {{ item.idempotency_key }}
                  </div>
                </td>

                <!-- Devices (Unified Multi-Device Support - Hidden for Operator) -->
                <td v-if="!isOperator" class="py-2.5 px-3">
                  <div v-if="item.device_count > 1 || (item.devices_summary && item.devices_summary.includes(','))">
                    <span class="px-2 py-0.5 rounded-md font-bold bg-cyan-500/15 text-cyan-500 dark:text-cyan-400 border border-cyan-500/30 text-3xs" :title="item.devices_summary">
                      {{ item.device_count || 2 }} Devices
                    </span>
                    <div class="text-4xs text-slate-400 font-mono truncate max-w-[150px] mt-0.5" :title="item.devices_summary">
                      {{ item.devices_summary }}
                    </div>
                  </div>
                  <span v-else class="px-2 py-0.5 rounded-md font-semibold font-mono bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 text-3xs">
                    {{ item.devices_summary || (item.device ? item.device.device_code : (item.device_id ? 'Station #' + item.device_id : 'All Devices')) }}
                  </span>
                </td>

                <!-- Destination -->
                <td class="py-2.5 px-3">
                  <div class="font-semibold text-slate-800 dark:text-slate-200">
                    {{ item.destination ? item.destination.name : ('Dest #' + item.destination_id) }}
                  </div>
                  <span class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold" :class="protocolBadgeClass(item.destination ? item.destination.type : 'REST')">
                    {{ item.destination ? item.destination.type : 'REST_API' }}
                  </span>
                </td>

                <!-- Waktu Pertama Kirim (Enqueued / Dispatched) -->
                <td class="py-2.5 px-3 text-3xs font-mono tabular-nums whitespace-nowrap">
                  <div class="font-semibold text-slate-800 dark:text-slate-200">{{ formatTime(item.created_at) }}</div>
                  <div class="text-4xs text-slate-400">{{ formatDate(item.created_at) }}</div>
                  <div v-if="item.attempts > 1 && item.last_attempt_at" class="text-4xs text-amber-500 font-medium mt-0.5" :title="'Terakhir dicoba: ' + formatTime(item.last_attempt_at)">
                    Last: {{ formatTime(item.last_attempt_at) }}
                  </div>
                </td>

                <!-- Window -->
                <td class="py-2.5 px-3 text-3xs font-mono tabular-nums text-slate-500 whitespace-nowrap">
                  <div class="font-semibold text-cyan-600 dark:text-cyan-400">{{ formatDate(item.period_start) }}</div>
                  <div>{{ formatTime(item.period_start) }}</div>
                  <div>→ {{ formatTime(item.period_end) }}</div>
                </td>

                <!-- Interval -->
                <td class="py-2.5 px-3 text-center">
                  <span class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/20">
                    {{ formatQueueIntervalBadge(item) }}
                  </span>
                </td>

                <!-- Params Count -->
                <td class="py-2.5 px-3 text-center font-mono tabular-nums font-bold">
                  {{ item.parameter_count }}
                </td>

                <!-- Status -->
                <td class="py-2.5 px-3">
                  <span class="px-2 py-0.5 rounded-md text-3xs font-bold" :class="statusBadgeClass(item.status)">
                    {{ item.status }}
                  </span>
                  <div v-if="item.next_retry_at && item.status === 'RETRYING'" class="text-4xs font-mono tabular-nums text-orange-400 mt-0.5">
                    Retry at {{ formatTime(item.next_retry_at) }}
                  </div>
                </td>

                <!-- Attempts -->
                <td class="py-2.5 px-3 text-center font-mono tabular-nums">
                  <span :class="item.attempts >= item.max_retries ? 'text-rose-500 font-bold' : 'text-slate-600 dark:text-slate-300'">
                    {{ item.attempts }} / {{ item.max_retries }}
                  </span>
                </td>

                <!-- Last HTTP Code / Error -->
                <td class="py-2.5 px-3 max-w-[200px]">
                  <div v-if="item.last_http_code > 0" class="font-mono font-bold" :class="httpCodeClass(item.last_http_code)">
                    HTTP {{ item.last_http_code }}
                  </div>
                  <div v-if="item.last_error" class="text-4xs font-mono text-rose-400 truncate" :title="item.last_error">
                    {{ item.last_error }}
                  </div>
                  <div v-else-if="item.last_response" class="text-4xs font-mono text-emerald-500 truncate" :title="item.last_response">
                    {{ item.last_response }}
                  </div>
                </td>

                <!-- Actions -->
                <td class="py-2.5 px-3 text-right space-x-1 whitespace-nowrap">
                  <!-- Inspect Detail Modal -->
                  <button
                    @click="openInspectModal(item.id)"
                    class="px-2 py-1 rounded bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 transition text-3xs font-medium"
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
                    class="px-2 py-1 rounded bg-rose-500/20 hover:bg-rose-500/30 text-rose-400 border border-rose-500/30 transition text-3xs font-medium"
                    title="Cancel pending delivery"
                  >
                    Cancel
                  </button>

                  <!-- Delete (Manual item cleanup by Admin/Superadmin) -->
                  <button
                    v-if="canManageQueue"
                    @click="deleteQueueItem(item.id)"
                    class="px-2 py-1 rounded bg-rose-500/10 hover:bg-rose-500/20 text-rose-500 border border-rose-500/30 transition text-3xs font-semibold"
                    title="Hapus permanen antrean ini dari database"
                  >
                    Delete
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div class="px-4 py-3 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between text-2xs tabular-nums text-slate-500">
          <div>
            Showing {{ queueItems.length }} of {{ queueTotal }} items (Page {{ page }} of {{ totalPages }})
          </div>
          <div class="flex items-center space-x-1">
            <button
              @click="changePage(page - 1)"
              :disabled="page <= 1"
              class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 disabled:opacity-30 cursor-pointer font-medium"
            >
              Prev
            </button>
            <button
              @click="changePage(page + 1)"
              :disabled="page >= totalPages"
              class="px-2 py-1 rounded border border-slate-300 dark:border-slate-700 disabled:opacity-30 cursor-pointer font-medium"
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
          <h2 class="text-sm font-bold tracking-tight text-slate-900 dark:text-white uppercase">
            Upstream Outbound Destinations
          </h2>
          <p class="text-2xs text-slate-500 dark:text-slate-400 mt-0.5">
            Configure target servers, webhooks, brokers, and telemetry interval filters.
          </p>
        </div>
        <button
          v-if="canManageDestinations"
          @click="openAddDestinationModal"
          class="px-3 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-600 text-white text-xs font-semibold transition flex items-center space-x-1.5 shadow-xs"
        >
          <span>+ Add Destination</span>
        </button>
      </div>

      <!-- Empty State when 0 Destinations -->
      <div
        v-if="destinations.length === 0"
        class="p-10 rounded-2xl border border-dashed border-slate-300 dark:border-slate-800 text-center bg-white/40 dark:bg-slate-900/40 space-y-3"
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
            class="px-4 py-2 rounded-xl bg-cyan-500 hover:bg-cyan-600 text-white font-semibold text-xs transition inline-flex items-center space-x-2 shadow-md cursor-pointer"
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
                <h3 class="text-base font-bold text-slate-900 dark:text-white mt-1.5">
                  {{ dest.name }}
                </h3>
              </div>
              <!-- Enable / Disable Switch (Admin/Superadmin clickable, Operator read-only badge) -->
              <button
                v-if="canManageDestinations"
                @click="toggleDestinationEnabled(dest)"
                class="px-2 py-0.5 rounded-full text-3xs font-semibold border transition cursor-pointer shadow-xs hover:scale-105"
                :class="dest.enabled
                  ? 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border-emerald-500/30 hover:bg-emerald-500/30'
                  : 'bg-slate-200 dark:bg-slate-800 text-slate-500 border-slate-300 dark:border-slate-700 hover:bg-slate-300 dark:hover:bg-slate-700'"
                :title="dest.enabled ? 'Klik untuk nonaktifkan (disable) pengiriman ke destinasi ini' : 'Klik untuk aktifkan (enable) pengiriman ke destinasi ini'"
              >
                {{ dest.enabled ? '● ACTIVE' : '○ DISABLED' }}
              </button>
              <span
                v-else
                class="px-2 py-0.5 rounded-full text-3xs font-semibold border cursor-default"
                :class="dest.enabled
                  ? 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border-emerald-500/30'
                  : 'bg-slate-200 dark:bg-slate-800 text-slate-500 border-slate-300 dark:border-slate-700'"
                :title="dest.enabled ? 'Status: Pengiriman Aktif' : 'Status: Pengiriman Dinonaktifkan'"
              >
                {{ dest.enabled ? '● ACTIVE' : '○ DISABLED' }}
              </span>
            </div>

            <!-- Details List -->
            <div class="mt-3 space-y-1.5 text-2xs text-slate-600 dark:text-slate-300">
              <div v-if="(dest.auth_type === 'LOGIN_BEARER' || dest.auth_type === 'SPARING_JWT_2STEP') && dest.auth_url" class="flex items-center justify-between">
                <span class="text-slate-400">{{ dest.auth_type === 'SPARING_JWT_2STEP' ? 'Secret URL:' : 'Auth URL:' }}</span>
                <span class="font-mono text-3xs truncate max-w-[200px] text-amber-500 dark:text-amber-400" :title="dest.auth_url">{{ dest.auth_url }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-400">{{ dest.auth_type === 'LOGIN_BEARER' ? 'Data URL:' : (dest.auth_type === 'SPARING_JWT_2STEP' || dest.auth_type === 'SPARING_JWT_STATIC' ? 'Send URL:' : 'Endpoint:') }}</span>
                <span class="font-mono text-3xs truncate max-w-[200px]" :title="dest.endpoint">{{ dest.endpoint }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-400">Schedule:</span>
                <span class="font-semibold text-cyan-600 dark:text-cyan-400">
                  {{ formatDestinationSchedule(dest) }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-400">Auth Mode:</span>
                <span :class="(dest.auth_type === 'LOGIN_BEARER' || dest.auth_type === 'SPARING_JWT_2STEP' || dest.auth_type === 'SPARING_JWT_STATIC' || dest.auth_type === 'KLHK_ONLIMO_BODY') ? 'text-cyan-600 dark:text-cyan-400 font-semibold' : 'font-medium'">
                  {{ formatAuthModeLabel(dest.auth_type) }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-400">Deliveries:</span>
                <div class="font-mono tabular-nums text-3xs">
                  <span class="text-emerald-600 dark:text-emerald-400 font-bold">{{ dest.success_deliveries }} OK</span>
                  <span v-if="dest.failed_deliveries > 0" class="text-rose-500 font-bold ml-1.5">{{ dest.failed_deliveries }} ERR</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Card Actions -->
          <div class="mt-4 pt-3 border-t border-slate-100 dark:border-slate-800/80 flex items-center justify-between text-xs">
            <button
              @click="selectDestinationForConfig(dest.id)"
              class="text-cyan-600 dark:text-cyan-400 hover:text-cyan-500 font-semibold transition"
            >
              Configure JSON Schema →
            </button>
            <div class="space-x-1">
              <button
                @click="openEditDestinationModal(dest)"
                class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 text-2xs font-medium"
              >
                Edit
              </button>
              <button
                v-if="canManageDestinations"
                @click="deleteDestination(dest.id)"
                class="px-2.5 py-1 rounded bg-rose-500/10 hover:bg-rose-500/20 text-rose-500 dark:text-rose-400 border border-rose-500/30 text-2xs font-medium"
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
    <div v-if="activeTab === 'config'" class="space-y-4">
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
          class="mt-2 px-4 py-2 rounded-xl bg-cyan-500 hover:bg-cyan-600 text-white font-semibold text-xs transition inline-flex items-center space-x-1.5 shadow-md cursor-pointer"
        >
          <span>→ Go to Destinations Setup</span>
        </button>
      </div>

      <div v-else class="space-y-4">
        <!-- Destination Selector Header -->
        <div class="flex flex-wrap items-center justify-between gap-3 p-3.5 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
        <div class="flex items-center space-x-2">
          <span class="text-xs text-slate-500 font-semibold">Active Destination:</span>
          <select
            v-model="selectedDestinationId"
            @change="loadDestinationPayloadConfig"
            class="bg-slate-100 dark:bg-slate-800 text-xs px-3 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 font-semibold text-slate-900 dark:text-white cursor-pointer focus:outline-hidden"
          >
            <option v-for="d in destinations" :key="d.id" :value="d.id">
              {{ d.name }} ({{ d.type }})
            </option>
          </select>
          <span class="text-3xs font-mono text-slate-400">Version #{{ payloadConfig.version || 1 }}</span>
        </div>

        <div class="flex items-center space-x-2">
          <!-- Test Payload Action -->
          <button
            @click="testPayloadTransmission"
            :disabled="testingPayload"
            class="px-3 py-1.5 rounded-lg bg-purple-500 hover:bg-purple-600 text-white text-xs font-semibold transition flex items-center space-x-1.5 shadow-xs"
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
            class="px-4 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-600 text-white text-xs font-semibold transition flex items-center space-x-1.5 shadow-xs"
          >
            <span>💾 Save Configuration</span>
          </button>
        </div>
      </div>

      <!-- Test Result Alert Banner -->
      <div
        v-if="testResult"
        class="p-3.5 rounded-xl border text-2xs space-y-2.5"
        :class="testResult.success
          ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-600 dark:text-emerald-400'
          : 'bg-rose-500/10 border-rose-500/30 text-rose-600 dark:text-rose-400'"
      >
        <div class="flex items-start justify-between">
          <div>
            <div class="font-bold flex flex-wrap items-center gap-2">
              <span>
                {{ testResult.is_two_step
                  ? (testResult.success ? '✔ 2-STEP TEST DELIVERY SUCCEEDED' : '✖ 2-STEP TEST DELIVERY FAILED')
                  : (testResult.success ? '✔ TEST DELIVERY SUCCEEDED' : '✖ TEST DELIVERY FAILED') }}
              </span>
              <span v-if="testResult.http_code" class="font-mono">HTTP {{ testResult.http_code }}</span>
              <span class="font-mono">({{ testResult.latency_ms }}ms)</span>
            </div>
            <div v-if="!testResult.is_two_step" class="mt-1 text-3xs opacity-90 font-mono break-all">
              {{ testResult.error_message || testResult.response_text }}
            </div>
          </div>
          <button @click="testResult = null" class="text-slate-400 hover:text-white ml-2">✕</button>
        </div>

        <!-- 2-Step Auth + Data Breakdown -->
        <div v-if="testResult.is_two_step" class="grid grid-cols-1 md:grid-cols-2 gap-2.5 pt-1 text-3xs">
          <!-- Step 1 Card -->
          <div
            class="p-2.5 rounded-lg border space-y-1"
            :class="testResult.auth_success
              ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-700 dark:text-emerald-300'
              : 'bg-rose-500/10 border-rose-500/30 text-rose-700 dark:text-rose-300'"
          >
            <div class="flex items-center justify-between font-bold">
              <span>
                {{ isActiveDestSparing2Step
                  ? (testResult.auth_success ? '✔ STEP 1 — SECRET SENSOR & JWT OK' : '✖ STEP 1 — SECRET SENSOR FAILED')
                  : (testResult.auth_success ? '✔ STEP 1 — AUTH LOGIN OK' : '✖ STEP 1 — AUTH LOGIN FAILED') }}
              </span>
              <span class="font-mono">
                <template v-if="testResult.auth_used_cache">CACHE</template>
                <template v-else>
                  <span v-if="testResult.auth_http_code">HTTP {{ testResult.auth_http_code }} · </span>{{ testResult.auth_latency_ms || 0 }}ms
                </template>
              </span>
            </div>
            <div v-if="testResult.auth_url" class="text-slate-400 font-mono truncate" :title="testResult.auth_url">
              {{ isActiveDestSparing2Step ? 'GET' : 'POST' }} {{ testResult.auth_url }}
            </div>
            <div class="opacity-95 font-mono break-all">
              {{ testResult.auth_message }}
              <span v-if="testResult.auth_token_preview" class="text-cyan-500 dark:text-cyan-400 ml-1">
                [{{ isActiveDestSparing2Step ? 'JWT' : 'Bearer' }} {{ testResult.auth_token_preview }}]
              </span>
            </div>
          </div>

          <!-- Step 2 Card -->
          <div
            class="p-2.5 rounded-lg border space-y-1"
            :class="!testResult.auth_success
              ? 'bg-slate-800/40 border-slate-700/60 text-slate-400'
              : testResult.success
                ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-700 dark:text-emerald-300'
                : 'bg-rose-500/10 border-rose-500/30 text-rose-700 dark:text-rose-300'"
          >
            <div class="flex items-center justify-between font-bold">
              <span>
                {{ !testResult.auth_success
                  ? '○ STEP 2 — DATA SKIPPED'
                  : (testResult.success ? '✔ STEP 2 — DATA ACCEPTED' : '✖ STEP 2 — DATA REJECTED / ERROR') }}
              </span>
              <span v-if="testResult.auth_success" class="font-mono">
                <span v-if="testResult.http_code">HTTP {{ testResult.http_code }} · </span>{{ testResult.data_latency_ms || 0 }}ms
              </span>
            </div>
            <div v-if="testResult.data_url" class="text-slate-400 font-mono truncate" :title="testResult.data_url">
              POST {{ testResult.data_url }}{{ isActiveDestSparing2Step ? '  {"token": "<jwt_hs256>"}' : '' }}
            </div>
            <div v-if="testResult.auth_success" class="opacity-95 font-mono break-all">
              {{ testResult.response_text || testResult.error_message }}
            </div>
            <div v-else class="opacity-75">
              {{ isActiveDestSparing2Step
                ? 'Menunggu Step 1 Secret Sensor berhasil diambil sebelum mengirim token JWT.'
                : 'Menunggu Step 1 Login berhasil sebelum mengirim data sensor.' }}
            </div>
          </div>
        </div>
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
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-medium"
              >
                <option value="STANDARD">Standard Datalogger</option>
                <option value="KLHK_WQMS_ONLIMO">KLHK WQMS / ONLIMO (data{IDStasiun, Tanggal, Jam, 13 Sensor} + apikey)</option>
                <option value="KLHK_SPARING_HOURLY">KLHK SPARING — Per 1 Jam / 30 Data (uid + data[30])</option>
                <option value="KLHK_SPARING">KLHK SPARING — Per 2 Menit (Flat: uid + datetime + sensor)</option>
                <option value="KLHK_ISPU_AQMS">KLHK ISPU AQMS (id_stasiun + waktu + stat_*)</option>
                <option value="MULTI_TICK_SENSOR">Multi-Sensor Tick Array (multi[tick])</option>
                <option value="MULTI_DEVICE_FLAT">Multi-Device Unified Flat</option>
                <option value="MULTI_DEVICE_NESTED">Multi-Device Nested by Device</option>
                <option value="MULTI_DEVICE_ARRAY">Multi-Device Array of Devices</option>
                <option value="NESTED_SCADA">Nested CBI Telemetry</option>
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
                <option value="ONLIMO_DATA_OBJECT">ONLIMO_DATA_OBJECT (Root data{IDStasiun, Tanggal, Jam, ...} + apikey)</option>
                <option value="SPARING_DATA_ARRAY">SPARING_DATA_ARRAY (Root uid + data[datetime, ...])</option>
                <option value="MULTI_TICK_ARRAY">MULTI_TICK_ARRAY (Array multi[tick])</option>
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
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 text-3xs font-medium"
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
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 text-3xs font-bold text-amber-500"
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

          <!-- Nested Key Name (if NESTED, SPARING_DATA_ARRAY, or ONLIMO_DATA_OBJECT) -->
          <div v-if="payloadConfig.structure === 'NESTED' || payloadConfig.structure === 'SPARING_DATA_ARRAY' || payloadConfig.structure === 'ONLIMO_DATA_OBJECT'" class="text-2xs">
            <label class="block text-slate-500 mb-1">Nested Array / Object Key Name</label>
            <input
              v-model="payloadConfig.nested_data_key"
              @input="triggerLivePreview"
              type="text"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono"
              placeholder="e.g. data, measurements, telemetry"
            />
          </div>

          <!-- Station ID & Override -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-2xs" :class="{'sm:grid-cols-3': payloadConfig.structure === 'MULTI_TICK_ARRAY'}">
            <div>
              <label class="block text-slate-500 mb-1">Station ID / UID Field Name</label>
              <input
                v-model="payloadConfig.station_id_field"
                @input="triggerLivePreview"
                type="text"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono"
                placeholder="IDStasiun / uid / station_id"
              />
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Station ID / UID Value Override</label>
              <input
                v-model="payloadConfig.station_id_value"
                @input="triggerLivePreview"
                type="text"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono"
                placeholder="e.g. STASIUN_WQMS_01 or 79456d9c51128700b78b6fd0"
              />
            </div>
            <div v-if="payloadConfig.structure === 'MULTI_TICK_ARRAY'">
              <label class="block text-slate-500 mb-1">Module ID Value (tick.module_id)</label>
              <input
                v-model="payloadConfig.module_id_value"
                @input="triggerLivePreview"
                type="text"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono"
                placeholder="e.g. 3453 or 1"
              />
            </div>
          </div>

          <!-- Timestamp Field, Format & Timezone -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3 text-2xs">
            <div>
              <label class="block text-slate-500 mb-1">
                {{ payloadConfig.structure === 'ONLIMO_DATA_OBJECT' ? 'Date & Time Fields (Split)' : 'Timestamp Field Name' }}
              </label>
              <input
                v-model="payloadConfig.timestamp_field"
                @input="triggerLivePreview"
                type="text"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono"
                :placeholder="payloadConfig.structure === 'ONLIMO_DATA_OBJECT' ? 'Tanggal,Jam' : 'datetime / timestamp'"
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
                <option value="UNIX_SEC">UNIX Epoch Seconds (int_datetime)</option>
                <option value="CUSTOM">Custom (YYYY-MM-DD HH:mm:ss)</option>
              </select>
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Timestamp Timezone</label>
              <select
                v-model="payloadConfig.timestamp_timezone"
                @change="triggerLivePreview"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              >
                <option value="">Default (UTC / Server)</option>
                <option value="Asia/Jakarta">WIB (Asia/Jakarta UTC+7)</option>
                <option value="Asia/Makassar">WITA (Asia/Makassar UTC+8)</option>
                <option value="Asia/Jayapura">WIT (Asia/Jayapura UTC+9)</option>
                <option value="UTC">UTC</option>
              </select>
            </div>
          </div>

          <!-- Timestamp Bucket Source & Numeric Formatting -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-2xs">
            <div>
              <label class="block text-slate-500 mb-1">Sumber Waktu Bucket Agregasi</label>
              <select
                v-model="payloadConfig.timestamp_source"
                @change="triggerLivePreview"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              >
                <option value="PERIOD_END">Akhir Periode (Period End — Default)</option>
                <option value="PERIOD_START">Awal Periode (Period Start :00..:58 — SPARING)</option>
              </select>
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Format Angka Sensor JSON</label>
              <select
                v-model="payloadConfig.numeric_precision_mode"
                @change="triggerLivePreview"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
              >
                <option value="FIXED_2">2 Desimal Tetap (41.00, 0.00)</option>
                <option value="NATURAL_2">Natural Maks 2 Desimal (41, 41.1, 40.99, 0 — SPARING)</option>
              </select>
            </div>
          </div>

          <!-- Interval Field Name & Omit Toggle -->
          <div class="text-2xs">
            <div class="flex items-center justify-between mb-1">
              <label class="block text-slate-500">Aggregation Interval Field Name</label>
              <label class="flex items-center space-x-1.5 cursor-pointer text-3xs font-semibold text-amber-600 dark:text-amber-400">
                <input
                  v-model="payloadConfig.omit_interval_field"
                  @change="triggerLivePreview"
                  type="checkbox"
                  class="rounded border-slate-700 text-amber-500 focus:ring-0"
                />
                <span>Sembunyikan Field Interval (Omit)</span>
              </label>
            </div>
            <input
              v-model="payloadConfig.interval_field"
              @input="triggerLivePreview"
              :disabled="payloadConfig.omit_interval_field"
              type="text"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono disabled:opacity-40"
              placeholder="interval_minutes"
            />
          </div>

          <!-- Metadata, Quality & KLHK ISPU Sensor Status Flags -->
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

          <!-- KLHK WQMS / ONLIMO Canonical 13-Parameter Options -->
          <div
            v-if="payloadConfig.template_type === 'KLHK_WQMS_ONLIMO' || payloadConfig.structure === 'ONLIMO_DATA_OBJECT'"
            class="p-3 rounded-lg bg-emerald-500/5 border border-emerald-500/30 space-y-2 text-2xs"
          >
            <div class="flex flex-wrap items-center justify-between gap-2">
              <label class="flex items-center space-x-2 cursor-pointer font-semibold text-emerald-600 dark:text-emerald-400">
                <input
                  v-model="payloadConfig.fill_missing_preset_fields"
                  @change="triggerLivePreview"
                  type="checkbox"
                  class="rounded border-slate-700 text-emerald-500 focus:ring-0"
                />
                <span>Lengkapi Otomatis 13 Parameter Baku KLHK WQMS (Isi 0 jika sensor tidak terpasang / null)</span>
              </label>
            </div>
            <p class="text-4xs text-slate-400 leading-relaxed">
              Urutan key baku di dalam <span class="font-mono text-emerald-400">data</span>:
              <span class="font-mono">IDStasiun, Tanggal (%Y-%m-%d), Jam (%H:%M:%S), Suhu, DHL, TDS, Salinitas, DO, PH, Turbidity, Kedalaman, Nitrat, Amonia, COD, BOD, TSS</span>
            </p>
            <p v-if="!isActiveDestONLIMOBody" class="text-4xs text-amber-500 dark:text-amber-400">
              💡 Tip: Pilih Authentication Type <span class="font-bold">"API Key + Secret in JSON Body (KLHK WQMS / ONLIMO)"</span> pada menu Edit Destination untuk menyuntikkan <span class="font-mono">apikey</span> &amp; <span class="font-mono">apisecret</span> secara aman.
            </p>
          </div>

          <!-- KLHK ISPU Sensor Status Flags (stat_pm10 .. stat_hc) -->
          <div class="p-3 rounded-lg bg-cyan-500/5 border border-cyan-500/20 space-y-2 text-2xs">
            <div class="flex flex-wrap items-center justify-between gap-2">
              <label class="flex items-center space-x-2 cursor-pointer font-semibold text-cyan-600 dark:text-cyan-400">
                <input
                  v-model="payloadConfig.include_sensor_stat_flags"
                  @change="triggerLivePreview"
                  type="checkbox"
                  class="rounded border-slate-700 text-cyan-500 focus:ring-0"
                />
                <span>Sertakan Status Sensor ISPU (stat_pm10 .. stat_hc)</span>
              </label>
              <select
                v-if="payloadConfig.include_sensor_stat_flags"
                v-model="payloadConfig.sensor_stat_mode"
                @change="triggerLivePreview"
                class="bg-white dark:bg-slate-800 px-2.5 py-1 rounded-lg border border-slate-300 dark:border-slate-700 text-3xs font-medium"
              >
                <option value="STATIC_1">Static 1 (Selalu bernilai 1)</option>
                <option value="AUTO_QUALITY">Auto Quality (1 = GOOD, 0 = BAD/Null)</option>
              </select>
            </div>
            <p v-if="payloadConfig.include_sensor_stat_flags" class="text-4xs text-slate-400">
              Menambahkan key otomatis: <span class="font-mono">stat_pm10, stat_pm25, stat_so2, stat_co, stat_o3, stat_no2, stat_hc</span>
            </p>
          </div>

          <!-- Parameter Aliases Table -->
          <div class="space-y-2 pt-2 border-t border-slate-200 dark:border-slate-800">
            <div class="flex flex-wrap items-center justify-between gap-2 text-2xs">
              <div class="flex items-center space-x-2">
                <span class="font-bold text-slate-700 dark:text-slate-300">Parameter External Aliases & Filters</span>
                <span class="px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 font-mono tabular-nums text-3xs font-semibold">
                  {{ activeParameterCount }} / {{ parameterMappings.length }} Aktif
                </span>
              </div>
              <div class="flex items-center space-x-1.5">
                <button
                  type="button"
                  @click="selectAllParameters(true)"
                  class="px-2 py-0.5 rounded text-3xs bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-300 dark:border-slate-700 font-medium transition cursor-pointer"
                  title="Pilih / Ceklis Semua Parameter"
                >
                  ✓ Pilih Semua
                </button>
                <button
                  type="button"
                  @click="selectAllParameters(false)"
                  class="px-2 py-0.5 rounded text-3xs bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-300 dark:border-slate-700 font-medium transition cursor-pointer"
                  title="Hapus Semua Ceklis Parameter"
                >
                  ✗ Hapus Semua
                </button>
                <button
                  type="button"
                  @click="syncDeviceParameters(true)"
                  :disabled="loadingParameters"
                  class="px-2 py-0.5 rounded text-3xs bg-cyan-50 dark:bg-cyan-950/40 hover:bg-cyan-100 dark:hover:bg-cyan-900/60 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30 font-medium transition flex items-center space-x-1 cursor-pointer"
                  title="Ambil dan sinkronkan semua parameter dari seluruh perangkat fisik terpasang"
                >
                  <span :class="{'animate-spin': loadingParameters}">🔄</span>
                  <span>Sinkronkan Alat</span>
                </button>
                <button
                  type="button"
                  @click="addCustomParamRow"
                  class="px-2 py-0.5 rounded text-3xs bg-emerald-50 dark:bg-emerald-950/40 hover:bg-emerald-100 dark:hover:bg-emerald-900/60 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 font-medium transition cursor-pointer"
                  title="Tambah baris mapping parameter baru"
                >
                  + Tambah
                </button>
              </div>
            </div>

            <!-- Search box if many parameters -->
            <div v-if="parameterMappings.length > 4" class="relative">
              <input
                v-model="paramSearch"
                type="text"
                placeholder="🔍 Cari kode atau nama parameter..."
                class="w-full bg-slate-50 dark:bg-slate-800/80 px-2.5 py-1 rounded-lg border border-slate-200 dark:border-slate-700 text-3xs"
              />
            </div>

            <div class="max-h-72 overflow-y-auto rounded-lg border border-slate-200 dark:border-slate-800">
              <table class="w-full text-left text-3xs">
                <thead class="bg-slate-50 dark:bg-slate-800/80 text-slate-500 sticky top-0 z-10 shadow-xs font-semibold">
                  <tr>
                    <th class="p-2 w-12 text-center">Aktif</th>
                    <th class="p-2">Internal Code & Asal Alat</th>
                    <th class="p-2">External Field Name (Sensor Name)</th>
                    <th class="p-2 w-8 text-center">Aksi</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr
                    v-for="(param, idx) in filteredParameterMappings"
                    :key="(param.device_id || 'g') + '_' + param.parameter_code + '_' + idx"
                    class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition"
                  >
                    <td class="p-2 text-center">
                      <input
                        v-model="param.enabled"
                        @change="triggerLivePreview"
                        type="checkbox"
                        class="rounded border-slate-700 text-cyan-500 focus:ring-0 cursor-pointer"
                      />
                    </td>
                    <td class="p-2">
                      <div class="font-bold text-slate-700 dark:text-slate-300 font-mono">
                        {{ param.parameter_code }}
                        <span v-if="param.parameter_name && param.parameter_name !== param.parameter_code" class="text-slate-400 font-sans font-normal">
                          ({{ param.parameter_name }})
                        </span>
                      </div>
                      <div v-if="param.device_code" class="text-4xs text-slate-400 font-mono">
                        DEV: {{ param.device_code }}
                      </div>
                    </td>
                    <td class="p-2">
                      <input
                        v-model="param.external_name"
                        @input="triggerLivePreview"
                        type="text"
                        class="w-full bg-slate-100 dark:bg-slate-800 px-2 py-1 rounded border border-slate-300 dark:border-slate-700 text-3xs font-mono font-semibold text-cyan-600 dark:text-cyan-400 focus:border-cyan-500"
                        :placeholder="param.parameter_code"
                      />
                    </td>
                    <td class="p-2 text-center">
                      <button
                        type="button"
                        @click="removeParamRow(param)"
                        class="text-slate-400 hover:text-rose-500 transition text-2xs cursor-pointer"
                        title="Hapus parameter dari mapping"
                      >
                        ✕
                      </button>
                    </td>
                  </tr>
                  <tr v-if="filteredParameterMappings.length === 0">
                    <td colspan="4" class="p-4 text-center text-slate-400 italic">
                      Tidak ada parameter yang cocok dengan pencarian.
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
            <span class="font-mono">UTF-8 JSON</span>
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
      class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-slate-950/80 backdrop-blur-xs"
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
                Queue Inspection: <span class="font-mono">{{ inspectDetail.batch_id }}</span>
              </h3>
            </div>
            <div class="text-3xs text-slate-400 mt-0.5">
              Idempotency Key: <span class="font-mono">{{ inspectDetail.idempotency_key }}</span>
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
              <span class="text-4xs text-slate-500">Source Type: <span class="font-mono">CUSTOMER_PROCESSED</span></span>
            </div>
            <div class="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
              <table class="w-full text-left text-3xs">
                <thead class="bg-slate-100 dark:bg-slate-800 text-slate-500 uppercase font-semibold">
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
                    <td class="p-2 font-mono font-bold text-cyan-600 dark:text-cyan-400">
                      {{ val.device_code || ('Device #' + val.device_id) }}
                    </td>
                    <td class="p-2 font-mono font-bold text-slate-900 dark:text-white">{{ val.parameter_code }}</td>
                    <td class="p-2 text-slate-600 dark:text-slate-300">{{ val.parameter_name }}</td>
                    <td class="p-2 text-right font-mono tabular-nums font-bold" :class="val.value !== null && val.value !== undefined ? 'text-cyan-500' : 'text-slate-400'">
                      {{ formatAggValue(val.value) }}
                    </td>
                    <td class="p-2 text-slate-400">{{ val.unit || '-' }}</td>
                    <td class="p-2">
                      <span class="px-1.5 py-0.2 rounded font-bold" :class="val.quality === 'GOOD' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-rose-500/20 text-rose-400'">
                        {{ val.quality }}
                      </span>
                    </td>
                    <td class="p-2 font-mono text-slate-500">#{{ val.aggregation_result_id }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- SECTION 2: Exact Serialized Outbound Payload -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <h4 class="font-bold text-emerald-500 uppercase tracking-wider">
                2. Serialized Outbound Payload Snapshot (<span class="font-mono">v{{ inspectDetail.payload_version }}</span>)
              </h4>
              <button
                @click="copyToClipboard(inspectDetail.payload_snapshot)"
                class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white text-3xs font-semibold"
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
                <span class="font-mono font-bold" :class="httpCodeClass(inspectDetail.last_http_code)">
                  {{ inspectDetail.last_http_code || 'N/A' }}
                </span>
                <span class="text-slate-500">Last Attempt: <span class="font-mono">{{ formatTime(inspectDetail.last_attempt_at) }}</span></span>
              </div>
              <div v-if="inspectDetail.last_response" class="text-emerald-400">
                <div class="text-slate-400 mb-0.5 font-medium">Response Body:</div>
                <div class="p-2 rounded bg-slate-900 border border-slate-800 font-mono break-all">{{ inspectDetail.last_response }}</div>
              </div>
              <div v-if="inspectDetail.last_error" class="text-rose-400">
                <div class="text-slate-400 mb-0.5 font-medium">Last Error:</div>
                <div class="p-2 rounded bg-rose-950/30 border border-rose-900/50 font-mono break-all">{{ inspectDetail.last_error }}</div>
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
              class="px-3 py-1.5 rounded-lg bg-amber-500/20 text-amber-500 font-semibold border border-amber-500/30"
            >
              Retry Item
            </button>
            <button
              v-if="canManageQueue"
              @click="replayQueueItem(inspectDetail.id); showInspectModal = false;"
              class="px-3 py-1.5 rounded-lg bg-purple-500/20 text-purple-400 font-semibold border border-purple-500/30"
            >
              Replay as New Delivery
            </button>
          </div>
          <button
            @click="showInspectModal = false"
            class="px-4 py-1.5 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-white font-semibold"
          >
            Close
          </button>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- MODAL: PURGE / BERSIHKAN ANTREAN & LOG (ADMIN/SUPERADMIN)       -->
    <!-- ============================================================== -->
    <div
      v-if="showPurgeModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4"
    >
      <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg p-6 shadow-2xl space-y-4">
        <div class="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-slate-800">
          <div class="flex items-center space-x-2.5">
            <span class="w-8 h-8 rounded-lg bg-rose-500/10 text-rose-500 flex items-center justify-center font-bold text-base">🗑️</span>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white uppercase">Pembersihan Antrean & Log</h3>
              <p class="text-3xs text-slate-500">Hapus manual antrean dan riwayat pengiriman (Admin / Superadmin)</p>
            </div>
          </div>
          <button @click="closePurgeModal" class="text-slate-400 hover:text-slate-200 text-lg">✕</button>
        </div>

        <!-- FORM OPTIONS (When idle / ready to configure a new purge) -->
        <div v-if="!purgingQueue && purgeProgress.status === 'IDLE'" class="space-y-3.5 text-xs">
          <!-- Filter Status -->
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Kriteria Status Antrean</label>
            <select
              v-model="purgeConfig.status"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2.5 rounded-lg border border-slate-300 dark:border-slate-700 text-xs font-semibold"
            >
              <option value="ALL">SEMUA STATUS (Total Purge Seluruh Antrean)</option>
              <option value="FAILED">Hanya Antrean GAGAL (FAILED)</option>
              <option value="DELIVERED">Hanya Antrean TERKIRIM (DELIVERED)</option>
              <option value="ACKNOWLEDGED">Hanya Antrean TERKONFIRMASI (ACKNOWLEDGED)</option>
              <option value="CANCELLED">Hanya Antrean DIBATALKAN (CANCELLED)</option>
              <option value="RETRYING">Hanya Antrean RETRYING</option>
              <option value="PENDING">Hanya Antrean PENDING</option>
            </select>
          </div>

          <!-- Filter Destinasi -->
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Target Destinasi</label>
            <select
              v-model="purgeConfig.destination_id"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2.5 rounded-lg border border-slate-300 dark:border-slate-700 text-xs"
            >
              <option :value="null">Semua Destinasi (Termasuk yang setupnya sudah dihapus)</option>
              <option v-for="d in destinations" :key="d.id" :value="d.id">
                #{{ d.id }} - {{ d.name }} ({{ d.type }})
              </option>
            </select>
          </div>

          <!-- Opsi Purge Logs -->
          <div class="p-3 rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700 flex items-start space-x-3">
            <input
              id="purge-logs-cb"
              v-model="purgeConfig.purge_logs"
              type="checkbox"
              class="mt-1 w-4 h-4 rounded text-rose-500 border-slate-400 focus:ring-rose-500 cursor-pointer"
            />
            <label for="purge-logs-cb" class="cursor-pointer text-2xs text-slate-700 dark:text-slate-300 leading-snug">
              <span class="font-bold block text-slate-900 dark:text-white">Ikut bersihkan Delivery Logs</span>
              Hapus riwayat log transaksi pengiriman upstream yang sesuai kriteria di atas.
            </label>
          </div>

          <div class="text-3xs text-rose-500 bg-rose-500/10 p-2.5 rounded-lg border border-rose-500/20">
            ⚠️ Perhatian: Tindakan ini permanen dan data antrean/log yang terhapus tidak dapat dikembalikan.
          </div>
        </div>

        <!-- PROGRESS BAR DISPLAY (When running or completed) -->
        <div v-else class="space-y-4 py-2">
          <!-- Header Status -->
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <span
                v-if="purgingQueue"
                class="w-3 h-3 rounded-full bg-cyan-400 animate-ping"
              ></span>
              <span
                v-else-if="purgeProgress.status === 'COMPLETED'"
                class="w-5 h-5 rounded-full bg-emerald-500/20 text-emerald-500 flex items-center justify-center font-bold text-xs"
              >✓</span>
              <span
                v-else
                class="w-5 h-5 rounded-full bg-rose-500/20 text-rose-500 flex items-center justify-center font-bold text-xs"
              >✕</span>
              <span
                class="text-xs font-bold"
                :class="{
                  'text-cyan-500': purgingQueue,
                  'text-emerald-500': purgeProgress.status === 'COMPLETED',
                  'text-rose-500': purgeProgress.status === 'FAILED' || purgeProgress.status === 'CANCELLED'
                }"
              >
                {{
                  purgingQueue
                    ? 'Sedang membersihkan antrean di database...'
                    : (purgeProgress.status === 'COMPLETED' ? 'Pembersihan Selesai!' : 'Proses Terhenti: ' + purgeProgress.status)
                }}
              </span>
            </div>
            <div class="text-sm font-bold font-mono tabular-nums text-slate-900 dark:text-white">
              {{ (purgeProgress.percent || 0).toFixed(1) }}%
            </div>
          </div>

          <!-- Progress Bar Track -->
          <div class="w-full bg-slate-200 dark:bg-slate-800 rounded-full h-3.5 overflow-hidden p-0.5 border border-slate-300 dark:border-slate-700 shadow-inner">
            <div
              class="h-full rounded-full transition-all duration-300 relative overflow-hidden"
              :class="purgeProgress.status === 'COMPLETED' ? 'bg-gradient-to-r from-emerald-500 to-teal-400' : 'bg-gradient-to-r from-cyan-500 via-blue-500 to-rose-500'"
              :style="{ width: Math.max(purgeProgress.percent || 0, 3) + '%' }"
            >
              <div v-if="purgingQueue" class="absolute inset-0 bg-white/25 animate-pulse"></div>
            </div>
          </div>

          <!-- Progress Counts & Summary Grid -->
          <div class="grid grid-cols-2 gap-2 text-2xs">
            <div class="p-2.5 rounded-xl bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700">
              <span class="text-4xs text-slate-400 block uppercase font-semibold">Antrean Terhapus</span>
              <div class="font-bold font-mono tabular-nums text-xs text-rose-500 mt-0.5">
                {{ purgeProgress.purged_queue.toLocaleString() }}
                <span class="text-4xs text-slate-400 font-normal">/ {{ purgeProgress.total_queue.toLocaleString() }}</span>
              </div>
            </div>
            <div class="p-2.5 rounded-xl bg-slate-100 dark:bg-slate-800/80 border border-slate-200 dark:border-slate-700">
              <span class="text-4xs text-slate-400 block uppercase font-semibold">Log Terhapus</span>
              <div class="font-bold font-mono tabular-nums text-xs text-amber-500 mt-0.5">
                {{ purgeProgress.purged_logs.toLocaleString() }}
                <span class="text-4xs text-slate-400 font-normal">/ {{ purgeProgress.total_logs.toLocaleString() }}</span>
              </div>
            </div>
          </div>

          <div v-if="purgeProgress.status === 'COMPLETED'" class="text-3xs text-emerald-600 dark:text-emerald-400 bg-emerald-500/10 p-2.5 rounded-lg border border-emerald-500/20">
            ✓ Database telah dibersihkan secara aman tanpa mengganggu sistem monitoring.
          </div>
          <div v-else-if="purgeProgress.error_message" class="text-3xs text-rose-500 bg-rose-500/10 p-2.5 rounded-lg border border-rose-500/20 font-mono">
            Error: {{ purgeProgress.error_message }}
          </div>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-3 border-t border-slate-200 dark:border-slate-800">
          <template v-if="!purgingQueue && purgeProgress.status === 'IDLE'">
            <button
              @click="closePurgeModal"
              class="px-4 py-2 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-200 text-xs font-semibold hover:bg-slate-300 dark:hover:bg-slate-600 transition cursor-pointer"
            >
              Batal
            </button>
            <button
              @click="executePurgeQueue"
              class="px-4 py-2 rounded-lg bg-rose-600 text-white text-xs font-semibold hover:bg-rose-700 transition flex items-center space-x-1.5 shadow-md cursor-pointer"
            >
              <span>Eksekusi Purge</span>
            </button>
          </template>
          <template v-else-if="purgingQueue">
            <button
              @click="cancelPurgeExecution"
              class="px-4 py-2 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-200 text-xs font-semibold hover:bg-slate-300 dark:hover:bg-slate-600 transition cursor-pointer"
            >
              Batalkan Pembersihan
            </button>
          </template>
          <template v-else>
            <button
              @click="resetPurgeProgressState"
              class="px-4 py-2 rounded-lg bg-slate-200 dark:bg-slate-700 text-slate-800 dark:text-slate-200 text-xs font-semibold hover:bg-slate-300 dark:hover:bg-slate-600 transition cursor-pointer"
            >
              Bersihkan Lagi
            </button>
            <button
              @click="closePurgeModal"
              class="px-5 py-2 rounded-lg bg-cyan-600 hover:bg-cyan-700 text-white text-xs font-semibold transition shadow-md cursor-pointer"
            >
              Selesai & Tutup
            </button>
          </template>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- MODAL: ADD / EDIT DESTINATION                                  -->
    <!-- ============================================================== -->
    <div
      v-if="showDestModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-slate-950/80 backdrop-blur-xs"
    >
      <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-xl max-h-[90vh] overflow-y-auto p-5 space-y-4 shadow-2xl">
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
              placeholder="e.g. KLHK SPARING Outbound"
            />
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
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
              <label class="block text-slate-500 mb-1">Interval Agregasi Sumber</label>
              <select
                v-model.number="editingDest.interval_seconds"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-bold"
              >
                <option :value="0">All Configured Intervals</option>
                <option :value="120">2 Menit (Standar SPARING)</option>
                <option :value="300">5 Menit</option>
                <option :value="1800">30 Menit (Standar ISPU AQMS)</option>
                <option :value="3600">60 Menit (Rata-rata 1 Jam)</option>
              </select>
            </div>
          </div>

          <!-- Jadwal Pengiriman (Immediate Per-Bucket vs Bundled Per-Jam / 30 Data) -->
          <div>
            <label class="block text-slate-500 mb-1">Jadwal Pengiriman Data (Dispatch Schedule)</label>
            <select
              v-model.number="editingDest.dispatch_interval_seconds"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-bold text-cyan-600 dark:text-cyan-400"
            >
              <option :value="0">Kirim Langsung Tiap Selesai Agregasi (mis. Per 2 Menit / Realtime)</option>
              <option :value="1800">Kumpulkan & Kirim Tiap 30 Menit (Bundle 15 Data × 2 Menit)</option>
              <option :value="3600">Kumpulkan & Kirim Tiap 1 Jam (Bundle 30 Data × 2 Menit — SPARING)</option>
            </select>
            <p v-if="editingDest.dispatch_interval_seconds > 0" class="mt-1 text-3xs text-cyan-600 dark:text-cyan-400">
              📦 Mode Bundling: Mengumpulkan {{ Math.max(1, Math.round(editingDest.dispatch_interval_seconds / (editingDest.interval_seconds || 120))) }} data agregasi {{ (editingDest.interval_seconds || 120) / 60 }} menit ({{ ':00..' + String(Math.max(0, Math.round(editingDest.dispatch_interval_seconds / 60) - Math.round((editingDest.interval_seconds || 120) / 60))).padStart(2, '0') }}) lalu dikirim setiap {{ editingDest.dispatch_interval_seconds / 60 }} menit dalam format array "data": [...].
            </p>
          </div>

          <!-- Auth Settings -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div :class="(editingDest.auth_type === 'LOGIN_BEARER' || editingDest.auth_type === 'SPARING_JWT_2STEP' || editingDest.auth_type === 'KLHK_ONLIMO_BODY') ? 'sm:col-span-2' : ''">
              <label class="block text-slate-500 mb-1">Authentication Type</label>
              <select
                v-model="editingDest.auth_type"
                @change="onEditingAuthTypeChange"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-medium"
              >
                <option value="NONE">None / Open</option>
                <option value="KLHK_ONLIMO_BODY">API Key + Secret in JSON Body (KLHK WQMS / ONLIMO)</option>
                <option value="SPARING_JWT_2STEP">2-Step Secret Sensor → JWT HS256 (KLHK SPARING)</option>
                <option value="SPARING_JWT_STATIC">Static Secret Key → JWT HS256 (KLHK SPARING)</option>
                <option value="LOGIN_BEARER">2-Step Auth Login → Bearer Token (KLHK / ISPU AQMS)</option>
                <option value="BEARER_TOKEN">Static Bearer Token</option>
                <option value="API_KEY">API Key Header</option>
                <option value="BASIC_AUTH">HTTP Basic Auth</option>
              </select>
            </div>
            <div v-if="editingDest.auth_type === 'BEARER_TOKEN' || editingDest.auth_type === 'SPARING_JWT_STATIC'">
              <label class="block text-slate-500 mb-1">
                {{ editingDest.auth_type === 'SPARING_JWT_STATIC' ? 'SPARING JWT Secret Key (HS256)' : 'Bearer Token' }}
              </label>
              <input
                v-model="editingDest.auth_token"
                type="password"
                class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
                placeholder="***REDACTED***"
              />
            </div>
            <div v-if="editingDest.auth_type === 'API_KEY'" class="sm:col-span-2 grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-500 mb-1">API Key Header Name</label>
                <input
                  v-model="editingDest.api_key_header"
                  type="text"
                  class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                  placeholder="X-API-Key"
                />
              </div>
              <div>
                <label class="block text-slate-500 mb-1">API Key Value</label>
                <input
                  v-model="editingDest.api_key_value"
                  type="password"
                  class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700"
                  placeholder="***REDACTED***"
                />
              </div>
            </div>
          </div>

          <!-- KLHK WQMS / ONLIMO Root JSON Body Credentials (apikey + apisecret) -->
          <div
            v-if="editingDest.auth_type === 'KLHK_ONLIMO_BODY'"
            class="p-3 rounded-xl bg-emerald-500/5 border border-emerald-500/30 space-y-3"
          >
            <div class="flex items-center justify-between">
              <span class="font-bold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider text-3xs">
                KLHK WQMS / ONLIMO — Root JSON Credentials (apikey + apisecret)
              </span>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-500 mb-1">API Key (apikey)</label>
                <input
                  v-model="editingDest.api_key_value"
                  type="password"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                  placeholder="***REDACTED***"
                />
              </div>
              <div>
                <label class="block text-slate-500 mb-1">API Secret (apisecret)</label>
                <input
                  v-model="editingDest.auth_token"
                  type="password"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                  placeholder="***REDACTED***"
                />
              </div>
            </div>
            <p class="text-4xs text-slate-400 leading-relaxed">
              Menyuntikkan <code class="text-emerald-400">{"data": {...}, "apikey": "...", "apisecret": "..."}</code> secara otomatis saat pengiriman HTTP POST. Kredensial tetap di-redact di antrean database &amp; UI.
            </p>
          </div>

          <!-- 2-Step KLHK SPARING Secret Sensor -> JWT HS256 -->
          <div
            v-if="editingDest.auth_type === 'SPARING_JWT_2STEP'"
            class="p-3 rounded-xl bg-emerald-500/5 border border-emerald-500/30 space-y-3"
          >
            <div class="flex items-center justify-between">
              <span class="font-bold text-emerald-600 dark:text-emerald-400 uppercase tracking-wider text-3xs">
                Step 1 — KLHK SPARING Secret Sensor (GET secret → Sign JWT HS256)
              </span>
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Secret Sensor URL (KLHK_SECRET_URL_2STEP — HTTP GET)</label>
              <input
                v-model="editingDest.auth_url"
                type="text"
                class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                placeholder="https://sparing.kemenlh.go.id/api/secret-sensor"
              />
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-500 mb-1">Siklus Pengambilan Secret</label>
                <select
                  v-model.number="editingDest.auth_cache_seconds"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 text-3xs"
                >
                  <option :value="0">Ambil Secret Baru Setiap Kirim (Strict 2-Step)</option>
                  <option :value="900">Cache Secret 15 Menit (Auto Refresh jika ditolak)</option>
                  <option :value="1800">Cache Secret 30 Menit (Auto Refresh jika ditolak)</option>
                  <option :value="3300">Cache Secret 55 Menit (Auto Refresh jika ditolak)</option>
                </select>
              </div>
              <div>
                <label class="block text-slate-500 mb-1">JSON Field Secret (Opsional)</label>
                <input
                  v-model="editingDest.auth_token_path"
                  type="text"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                  placeholder="Auto (Plain text / secret / token)"
                />
              </div>
            </div>
            <p class="text-4xs text-slate-400 leading-relaxed">
              Otomatis mengambil dynamic secret via GET, mengenkode payload JSON ke JWT HS256, dan mengirim <code class="text-emerald-400">{"token": "&lt;jwt&gt;"}</code> ke URL Step 2.
            </p>
          </div>

          <!-- 2-Step Auth Login (Step 1: POST auth_url -> Token) -->
          <div
            v-if="editingDest.auth_type === 'LOGIN_BEARER'"
            class="p-3 rounded-xl bg-cyan-500/5 border border-cyan-500/30 space-y-3"
          >
            <div class="flex items-center justify-between">
              <span class="font-bold text-cyan-600 dark:text-cyan-400 uppercase tracking-wider text-3xs">
                Step 1 — Auth Login Endpoint (POST JSON { username, password })
              </span>
            </div>
            <div>
              <label class="block text-slate-500 mb-1">Auth Login URL (auth_url)</label>
              <input
                v-model="editingDest.auth_url"
                type="text"
                class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                placeholder="https://ispu.kemenlh.go.id/api/v1/auth"
              />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-500 mb-1">Username</label>
                <input
                  v-model="editingDest.username"
                  type="text"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                  placeholder="pt_cakrawala_bima_instrument"
                />
              </div>
              <div>
                <label class="block text-slate-500 mb-1">Password</label>
                <input
                  v-model="editingDest.password"
                  type="password"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                  placeholder="***REDACTED***"
                />
              </div>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <div>
                <label class="block text-slate-500 mb-1">Mode Siklus Pengiriman Auth</label>
                <select
                  v-model.number="editingDest.auth_cache_seconds"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 text-3xs"
                >
                  <option :value="0">2x Kirim Setiap Batch (Auth dulu → Data)</option>
                  <option :value="900">Cache Token 15 Menit (Auto Re-Auth jika 401)</option>
                  <option :value="1800">Cache Token 30 Menit (Auto Re-Auth jika 401)</option>
                  <option :value="3300">Cache Token 55 Menit (Auto Re-Auth jika 401)</option>
                </select>
              </div>
              <div>
                <label class="block text-slate-500 mb-1">JSON Field Token (Opsional)</label>
                <input
                  v-model="editingDest.auth_token_path"
                  type="text"
                  class="w-full bg-white dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
                  placeholder="Auto (token / access_token / data.token)"
                />
              </div>
            </div>
          </div>

          <div>
            <label class="block text-slate-500 mb-1">
              {{ editingDest.auth_type === 'LOGIN_BEARER'
                ? 'Step 2 — Data Endpoint URL (data_url, Bearer Token otomatis)'
                : (editingDest.auth_type === 'SPARING_JWT_2STEP' || editingDest.auth_type === 'SPARING_JWT_STATIC'
                  ? 'Step 2 — KLHK Send URL (KLHK_URL_2STEP — POST {"token": "<jwt_hs256>"})'
                  : (editingDest.auth_type === 'KLHK_ONLIMO_BODY'
                    ? 'KLHK WQMS / ONLIMO Endpoint URL (POST JSON { data, apikey, apisecret })'
                    : 'Endpoint (URL or host:port)')) }}
            </label>
            <input
              v-model="editingDest.endpoint"
              type="text"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
              :placeholder="editingDest.auth_type === 'LOGIN_BEARER'
                ? 'https://ispu.kemenlh.go.id/api/v1/aqmdata'
                : (editingDest.auth_type === 'SPARING_JWT_2STEP' || editingDest.auth_type === 'SPARING_JWT_STATIC'
                  ? 'https://sparing.kemenlh.go.id/api/send'
                  : (editingDest.auth_type === 'KLHK_ONLIMO_BODY'
                    ? 'https://ppkl.menlhk.go.id/onlimo/api/...'
                    : 'https://sparing.kemenlh.go.id/api/send or 192.168.1.50:5000'))"
            />
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

          <!-- Optional Custom HTTP Headers -->
          <div v-if="editingDest.type === 'REST_API' || editingDest.type === 'CLOUD'">
            <label class="block text-slate-500 mb-1">Custom HTTP Headers (Opsional JSON)</label>
            <input
              v-model="editingDest.custom_headers"
              type="text"
              class="w-full bg-slate-100 dark:bg-slate-800 p-2 rounded-lg border border-slate-300 dark:border-slate-700 font-mono text-3xs"
              placeholder='{"X-Client-ID": "cbi-aqms"}'
            />
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
      filterWindowDate: (this.$route.query && this.$route.query.window_date) || '',
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

      // Purge Modal & Progress
      showPurgeModal: false,
      purgingQueue: false,
      purgePollingTimer: null,
      purgeConfig: {
        status: 'ALL',
        destination_id: null,
        purge_logs: true,
      },
      purgeProgress: {
        is_running: false,
        status: 'IDLE',
        total_queue: 0,
        purged_queue: 0,
        total_logs: 0,
        purged_logs: 0,
        percent: 0,
        error_message: '',
      },

      // Destinations state
      destinations: [],
      showDestModal: false,
      editingDest: {
        id: 0,
        name: '',
        type: 'REST_API',
        endpoint: '',
        interval_seconds: 0,
        dispatch_interval_seconds: 0,
        auth_type: 'NONE',
        auth_url: '',
        auth_token_path: '',
        auth_cache_seconds: 0,
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
        module_id_value: '',
        timestamp_field: 'timestamp',
        timestamp_format: 'ISO8601',
        timestamp_source: 'PERIOD_END',
        timestamp_timezone: '',
        numeric_precision_mode: 'FIXED_2',
        interval_field: 'interval_minutes',
        omit_interval_field: false,
        structure: 'FLAT',
        nested_data_key: 'data',
        include_device_identity: false,
        device_identity_mode: 'NONE',
        conflict_resolution: 'FAIL_ON_CONFLICT',
        include_units: false,
        include_quality: false,
        include_metadata: true,
        include_sensor_stat_flags: false,
        sensor_stat_mode: 'STATIC_1',
        fill_missing_preset_fields: false,
        static_metadata_json: '',
        parameter_mappings_json: '[]',
        required_fields_json: '[]',
        version: 1,
      },
      paramSearch: '',
      loadingParameters: false,
      parameterMappings: [],
      previewJSON: '',
      savingConfig: false,
      testingPayload: false,
      testResult: null,
    };
  },
  computed: {
    activeParameterCount() {
      return (this.parameterMappings || []).filter(p => p.enabled).length;
    },
    isActiveDestSparing2Step() {
      const dest = (this.destinations || []).find(d => d.id === this.selectedDestinationId);
      return Boolean(dest && dest.auth_type === 'SPARING_JWT_2STEP');
    },
    isActiveDestONLIMOBody() {
      const dest = (this.destinations || []).find(d => d.id === this.selectedDestinationId);
      return Boolean(dest && dest.auth_type === 'KLHK_ONLIMO_BODY');
    },
    filteredParameterMappings() {
      if (!this.paramSearch || !this.paramSearch.trim()) {
        return this.parameterMappings || [];
      }
      const q = this.paramSearch.trim().toLowerCase();
      return (this.parameterMappings || []).filter(p => {
        const code = (p.parameter_code || '').toLowerCase();
        const name = (p.parameter_name || '').toLowerCase();
        const ext = (p.external_name || '').toLowerCase();
        const dev = (p.device_code || '').toLowerCase();
        return code.includes(q) || name.includes(q) || ext.includes(q) || dev.includes(q);
      });
    },
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
        if (newTab === 'config') {
          this.loadDestinationPayloadConfig();
        }
      }
    },
    '$route.query.window_date'(newDate) {
      const normalized = newDate || '';
      if (this.filterWindowDate !== normalized) {
        this.filterWindowDate = normalized;
        this.page = 1;
        this.fetchQueue();
      }
    },
    selectedDestinationId(newId) {
      if (newId && this.activeTab === 'config') {
        this.loadDestinationPayloadConfig();
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

    if (this.activeTab === 'config') {
      this.loadDestinationPayloadConfig();
    }

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
    if (this.purgePollingTimer) {
      clearInterval(this.purgePollingTimer);
    }
  },
  methods: {
    selectTab(tab) {
      this.activeTab = tab;
      if (this.$route.query.tab !== tab) {
        this.$router.replace({ query: { ...this.$route.query, tab } }).catch(() => {});
      }
      if (tab === 'config') {
        this.loadDestinationPayloadConfig();
      }
    },
    async fetchSummary() {
      try {
        const res = await axios.get('/api/integration/queue/summary', {
          params: { _t: Date.now() },
        });
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
          if (this.activeTab === 'config' && this.selectedDestinationId) {
            this.loadDestinationPayloadConfig();
          }
        }
      } catch (err) {
        console.error('Failed to load destinations:', err);
      }
    },
    onQueueFilterChange() {
      this.page = 1;
      this.fetchQueue();
    },
    async fetchQueue() {
      this.loading = true;
      try {
        const params = {
          page: this.page,
          page_size: this.pageSize,
          _t: Date.now(),
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
        if (this.filterWindowDate && /^\d{4}-\d{2}-\d{2}$/.test(this.filterWindowDate)) {
          const [yy, mm, dd] = this.filterWindowDate.split('-').map(Number);
          const dayStart = new Date(yy, mm - 1, dd, 0, 0, 0, 0);
          const dayEnd = new Date(yy, mm - 1, dd + 1, 0, 0, 0, 0);
          if (!isNaN(dayStart.getTime()) && !isNaN(dayEnd.getTime())) {
            params.window_start = dayStart.toISOString();
            params.window_end = dayEnd.toISOString();
          }
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
    async deleteQueueItem(id) {
      if (!confirm('Hapus item antrean ini secara permanen dari database?')) return;
      try {
        await axios.delete(`/api/integration/queue/${id}`);
        this.fetchSummary();
        this.fetchQueue();
      } catch (err) {
        alert('Hapus item antrean gagal: ' + (err.response?.data?.error || err.message));
      }
    },
    resetPurgeProgressState() {
      this.purgingQueue = false;
      this.purgeProgress = {
        is_running: false,
        status: 'IDLE',
        total_queue: 0,
        purged_queue: 0,
        total_logs: 0,
        purged_logs: 0,
        percent: 0,
        error_message: '',
      };
    },
    openPurgeModal() {
      this.purgeConfig = {
        status: 'ALL',
        destination_id: null,
        purge_logs: true,
      };
      this.resetPurgeProgressState();
      this.checkPurgeStatus();
      this.showPurgeModal = true;
    },
    async checkPurgeStatus() {
      try {
        const res = await axios.get('/api/integration/queue/purge/status', {
          params: { _t: Date.now() },
        });
        // Only resume progress display if a purge job is actively running right now;
        // ignore stale COMPLETED/FAILED/CANCELLED states from previous runs so user can trigger a new purge.
        if (res.data && res.data.success && res.data.data && res.data.data.is_running) {
          this.purgeProgress = res.data.data;
          this.purgingQueue = true;
          this.startPurgePolling();
        }
      } catch (e) {}
    },
    async executePurgeQueue() {
      const statusLabel = this.purgeConfig.status === 'ALL' ? 'SEMUA status' : `status ${this.purgeConfig.status}`;
      if (!confirm(`Konfirmasi: Bersihkan antrean pengiriman dengan ${statusLabel}? Tindakan ini permanen.`)) return;

      this.purgingQueue = true;
      this.purgeProgress = {
        is_running: true,
        status: 'RUNNING',
        total_queue: 0,
        purged_queue: 0,
        total_logs: 0,
        purged_logs: 0,
        percent: 0,
        error_message: '',
      };

      try {
        const res = await axios.post('/api/integration/queue/purge', this.purgeConfig);
        if (res.data && res.data.success) {
          this.purgeProgress = res.data.data;
          if (this.purgeProgress.status === 'COMPLETED') {
            this.onPurgeFinished();
            return;
          }
          this.startPurgePolling();
        }
      } catch (err) {
        this.purgingQueue = false;
        this.purgeProgress.status = 'IDLE';
        alert('Purge antrean gagal: ' + (err.response?.data?.error || err.message));
      }
    },
    startPurgePolling() {
      if (this.purgePollingTimer) clearInterval(this.purgePollingTimer);
      this.purgePollingTimer = setInterval(async () => {
        try {
          const res = await axios.get('/api/integration/queue/purge/status', {
            params: { _t: Date.now() },
          });
          if (res.data && res.data.success) {
            this.purgeProgress = res.data.data;
            if (!this.purgeProgress.is_running) {
              clearInterval(this.purgePollingTimer);
              this.purgePollingTimer = null;
              this.onPurgeFinished();
            }
          }
        } catch (err) {
          // ignore transient poll error
        }
      }, 600);
    },
    onPurgeFinished() {
      this.purgingQueue = false;
      this.page = 1;
      this.fetchSummary();
      this.fetchQueue();
      if (this.purgeProgress.status === 'FAILED') {
        alert('Proses purge gagal: ' + (this.purgeProgress.error_message || 'Unknown error'));
      }
    },
    async cancelPurgeExecution() {
      try {
        await axios.post('/api/integration/queue/purge/cancel');
      } catch (e) {}
    },
    closePurgeModal() {
      if (this.purgePollingTimer) {
        clearInterval(this.purgePollingTimer);
        this.purgePollingTimer = null;
      }
      this.showPurgeModal = false;
      this.resetPurgeProgressState();
      this.fetchSummary();
      this.fetchQueue();
    },
    selectDestinationForConfig(destId) {
      this.selectedDestinationId = destId;
      this.activeTab = 'config';
      this.loadDestinationPayloadConfig();
    },
    async loadDestinationPayloadConfig() {
      if (!this.selectedDestinationId) return;
      try {
        this.parameterMappings = [];
        const res = await axios.get(`/api/destinations/${this.selectedDestinationId}/payload-config`);
        if (res.data && res.data.success) {
          this.payloadConfig = {
            timestamp_source: 'PERIOD_END',
            timestamp_timezone: '',
            numeric_precision_mode: 'FIXED_2',
            omit_interval_field: false,
            include_sensor_stat_flags: false,
            sensor_stat_mode: 'STATIC_1',
            fill_missing_preset_fields: false,
            ...res.data.data,
          };
          if (!this.payloadConfig.timestamp_source) {
            this.payloadConfig.timestamp_source = 'PERIOD_END';
          }
          if (!this.payloadConfig.numeric_precision_mode) {
            this.payloadConfig.numeric_precision_mode = 'FIXED_2';
          }
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
        await this.syncDeviceParameters(false);
        this.triggerLivePreview();
      } catch (err) {
        console.error('Failed to load payload config:', err);
      }
    },
    getPresetAliasMap(templateType) {
      if (templateType === 'KLHK_SPARING' || templateType === 'KLHK_SPARING_HOURLY') {
        return {
          'ph': 'pH',
          'p_h': 'pH',
          'cod': 'cod',
          'tss': 'tss',
          'nh3n': 'nh3n',
          'nh3-n': 'nh3n',
          'nh3_n': 'nh3n',
          'ammonia': 'nh3n',
          'amonia': 'nh3n',
          'debit': 'debit',
          'flow': 'debit',
          'flowrate': 'debit',
          'flow_rate': 'debit',
          'flow-rate': 'debit',
        };
      }
      if (templateType === 'KLHK_ISPU_AQMS') {
        return {
          'pm10': 'pm10',
          'pm2.5': 'pm25',
          'pm25': 'pm25',
          'pm2_5': 'pm25',
          'so2': 'so2',
          'co': 'co',
          'o3': 'o3',
          'no2': 'no2',
          'hc': 'hc',
          'nmhc': 'hc',
          'w-speed': 'ws',
          'wspeed': 'ws',
          'ws': 'ws',
          'wind_speed': 'ws',
          'wind-speed': 'ws',
          'wind-angle': 'wd',
          'wind_angle': 'wd',
          'wd': 'wd',
          'wind_dir': 'wd',
          'wind-dir': 'wd',
          'humidity': 'humidity',
          'hum': 'humidity',
          'rh': 'humidity',
          'temperature': 'temperature',
          'temp': 'temperature',
          'air-pressure': 'pressure',
          'air_pressure': 'pressure',
          'pressure': 'pressure',
          'press': 'pressure',
          'solar-radiation': 'sr',
          'solar_radiation': 'sr',
          'intensity': 'sr',
          'sr': 'sr',
          'rain-fall': 'rain_intensity',
          'rain_fall': 'rain_intensity',
          'rainfall': 'rain_intensity',
          'tprecipitation': 'rain_intensity',
          'rain_intensity': 'rain_intensity',
          'rain-intensity': 'rain_intensity',
        };
      }
      return {
        'no2': 'no2',
        'so2': 'so2',
        'o3': 'o3',
        'co': 'co',
        'pm25': 'pm25',
        'pm2.5': 'pm25',
        'pm10': 'pm10',
        'temp': 'temp',
        'temperature': 'temp',
        'humidity': 'humidity',
        'hum': 'humidity',
        'pressure': 'pressure',
        'air-pressure': 'pressure',
        'wspeed': 'wspeed',
        'w-speed': 'wspeed',
        'intensity': 'intensity',
        'solar-radiation': 'intensity',
        'tprecipitation': 'tprecipitation',
        'rain-fall': 'tprecipitation',
        'hc': 'hc',
        'wind_angle': 'wind_angle',
        'wind-angle': 'wind_angle',
        'rain_intensity': 'rain_intensity',
      };
    },
    resolveKLHKWQMSAlias(code, name) {
      const normalize = (raw) =>
        String(raw || '')
          .trim()
          .toLowerCase()
          .replace(/[-/.()[\],\s]+/g, '_')
          .replace(/_+/g, '_')
          .replace(/^_+|_+$/g, '');

      const matchToken = (raw) => {
        const tok = normalize(raw);
        if (!tok) return '';
        const exactMap = {
          suhu: 'Suhu',
          temperature: 'Suhu',
          temp: 'Suhu',
          suhu_air: 'Suhu',
          water_temp: 'Suhu',
          water_temperature: 'Suhu',
          wtemp: 'Suhu',
          dhl: 'DHL',
          cl: 'DHL',
          conductivity: 'DHL',
          cond: 'DHL',
          ec: 'DHL',
          electrical_conductivity: 'DHL',
          conductivity_dhl: 'DHL',
          daya_hantar_listrik: 'DHL',
          konduktivitas: 'DHL',
          tds: 'TDS',
          fl: 'TDS',
          total_dissolved_solids: 'TDS',
          padatan_terlarut: 'TDS',
          salinitas: 'Salinitas',
          salinity: 'Salinitas',
          sal: 'Salinitas',
          do: 'DO',
          dissolved_oxygen: 'DO',
          oksigen_terlarut: 'DO',
          ph: 'PH',
          p_h: 'PH',
          derajat_keasaman: 'PH',
          turbidity: 'Turbidity',
          turbudity: 'Turbidity',
          turb: 'Turbidity',
          ntu: 'Turbidity',
          kekeruhan: 'Turbidity',
          turbidity_ntu: 'Turbidity',
          turbudity_ntu: 'Turbidity',
          kedalaman: 'Kedalaman',
          depth: 'Kedalaman',
          water_depth: 'Kedalaman',
          water_level: 'Kedalaman',
          level: 'Kedalaman',
          tma: 'Kedalaman',
          tinggi_muka_air: 'Kedalaman',
          nitrat: 'Nitrat',
          nitrate: 'Nitrat',
          no3: 'Nitrat',
          no3_n: 'Nitrat',
          no3n: 'Nitrat',
          amonia: 'Amonia',
          ammonia: 'Amonia',
          nh4: 'Amonia',
          nh4_n: 'Amonia',
          nh4n: 'Amonia',
          nh3: 'Amonia',
          nh3_n: 'Amonia',
          nh3n: 'Amonia',
          ammonia_nh3_n: 'Amonia',
          amonia_nh3_n: 'Amonia',
          ammonia_nh4: 'Amonia',
          amonia_nh4: 'Amonia',
          cod: 'COD',
          chemical_oxygen_demand: 'COD',
          bod: 'BOD',
          bod5: 'BOD',
          biological_oxygen_demand: 'BOD',
          biochemical_oxygen_demand: 'BOD',
          tss: 'TSS',
          total_suspended_solids: 'TSS',
          padatan_tersuspensi: 'TSS',
        };
        if (exactMap[tok]) return exactMap[tok];

        const parts = tok.split('_').filter(Boolean);
        const set = new Set(parts);
        if (set.has('toc') || set.has('no2') || set.has('so2') || set.has('pm10') || set.has('pm25')) {
          return '';
        }
        if (set.has('dhl') || set.has('conductivity') || set.has('konduktivitas')) return 'DHL';
        if (set.has('tds')) return 'TDS';
        if (set.has('salinitas') || set.has('salinity')) return 'Salinitas';
        if (set.has('turbidity') || set.has('turbudity') || set.has('kekeruhan')) return 'Turbidity';
        if (set.has('kedalaman') || set.has('depth')) return 'Kedalaman';
        if (set.has('nitrat') || set.has('nitrate') || set.has('no3')) return 'Nitrat';
        if (set.has('amonia') || set.has('ammonia') || set.has('nh4') || set.has('nh3') || set.has('nh3n')) return 'Amonia';
        if (set.has('suhu') || set.has('temperature')) return 'Suhu';
        if (set.has('cod')) return 'COD';
        if (set.has('bod') || set.has('bod5')) return 'BOD';
        if (set.has('tss')) return 'TSS';
        if (set.has('do') && parts.length <= 2) return 'DO';
        if (set.has('ph') && parts.length <= 2) return 'PH';
        return '';
      };

      return matchToken(code) || matchToken(name) || '';
    },
    async syncDeviceParameters(forceOverwrite = false) {
      try {
        this.loadingParameters = true;
        const res = await axios.get('/api/devices?all=true&page_size=200');
        const raw = res.data?.data;
        let devices = [];
        if (Array.isArray(raw)) {
          devices = raw;
        } else if (raw && Array.isArray(raw.items)) {
          devices = raw.items;
        } else if (Array.isArray(res.data)) {
          devices = res.data;
        }

        const aliasMap = this.getPresetAliasMap(this.payloadConfig?.template_type);

        const existingByKey = {};
        const existingByCode = {};
        for (const m of this.parameterMappings || []) {
          const lowerCode = (m.parameter_code || '').trim().toLowerCase();
          if (lowerCode) {
            existingByKey[`${m.device_id || 0}_${lowerCode}`] = m;
            if (!existingByCode[lowerCode]) {
              existingByCode[lowerCode] = m;
            }
          }
        }

        const newMappings = [];
        const seen = new Set();

        for (const dev of devices) {
          // Hanya proses device yang aktif (kecuali yang inactive)
          if (dev.enabled === false) continue;
          if (dev.status && String(dev.status).toUpperCase() === 'INACTIVE') continue;
          if (dev.deleted_at) continue;

          let params = dev.parameters || [];
          if (!params || params.length === 0) {
            try {
              const pRes = await axios.get(`/api/devices/${dev.id}/parameters`);
              const pData = pRes.data?.data;
              params = Array.isArray(pData) ? pData : (pData?.items || []);
            } catch (e) {
              params = [];
            }
          }

          for (const param of params) {
            // Hanya tampilkan parameter yang aktif (kecuali yang inactive)
            if (param.enabled === false) continue;
            if (param.status && String(param.status).toUpperCase() === 'INACTIVE') continue;
            if (param.deleted_at) continue;

            const code = (param.parameter_code || param.code || '').trim();
            if (!code) continue;
            const lowerCode = code.toLowerCase();
            const compositeKey = `${dev.id}_${lowerCode}`;
            if (seen.has(compositeKey)) continue;
            seen.add(compositeKey);

            const existing = existingByKey[compositeKey] || existingByCode[lowerCode];
            const pName = param.parameter_name || param.name || existing?.parameter_name || code;
            const isONLIMO = this.payloadConfig?.template_type === 'KLHK_WQMS_ONLIMO';
            const wqmsAlias = isONLIMO ? this.resolveKLHKWQMSAlias(code, pName) : '';

            if (existing && !forceOverwrite) {
              newMappings.push({
                device_id: dev.id,
                device_code: dev.device_code || dev.code || `DEV-${dev.id}`,
                parameter_id: param.id,
                parameter_code: code,
                parameter_name: pName,
                external_name: existing.external_name || (isONLIMO ? (wqmsAlias || code) : (aliasMap[lowerCode] || code)),
                enabled: existing.enabled !== undefined ? Boolean(existing.enabled) : (isONLIMO ? Boolean(wqmsAlias) : true),
              });
            } else {
              newMappings.push({
                device_id: dev.id,
                device_code: dev.device_code || dev.code || `DEV-${dev.id}`,
                parameter_id: param.id,
                parameter_code: code,
                parameter_name: pName,
                external_name: isONLIMO ? (wqmsAlias || code) : (aliasMap[lowerCode] || code),
                enabled: isONLIMO ? Boolean(wqmsAlias) : true,
              });
            }
          }
        }

        if (newMappings.length > 0) {
          this.parameterMappings = newMappings;
        }
        this.triggerLivePreview();
      } catch (err) {
        console.warn('Failed to sync device parameters for outbound config:', err);
      } finally {
        this.loadingParameters = false;
      }
    },
    selectAllParameters(state) {
      for (const p of this.parameterMappings || []) {
        p.enabled = state;
      }
      this.triggerLivePreview();
    },
    addCustomParamRow() {
      const defaultCode = `sensor_${(this.parameterMappings || []).length + 1}`;
      this.parameterMappings.push({
        parameter_code: defaultCode,
        parameter_name: defaultCode,
        external_name: defaultCode,
        enabled: true,
      });
      this.triggerLivePreview();
    },
    removeParamRow(target) {
      const idx = this.parameterMappings.indexOf(target);
      if (idx !== -1) {
        this.parameterMappings.splice(idx, 1);
        this.triggerLivePreview();
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
          const requeued = Number(this.payloadConfig.requeued_batches || 0);
          if (requeued > 0) {
            alert(
              `Configuration version #${this.payloadConfig.version} saved successfully!\n` +
              `${requeued} antrean batch yang tertunda/gagal otomatis diperbarui payload-nya & dijadwalkan kirim ulang.`
            );
          } else {
            alert('Configuration version #' + this.payloadConfig.version + ' saved successfully!');
          }
          this.triggerLivePreview();
          this.fetchDestinations();
          this.fetchSummary();
          this.fetchQueue();
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
        this.payloadConfig.parameter_mappings_json = JSON.stringify(this.parameterMappings);
        const res = await axios.post(`/api/destinations/${this.selectedDestinationId}/test`, this.payloadConfig);
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
        dispatch_interval_seconds: 0,
        auth_type: 'NONE',
        auth_url: '',
        auth_token_path: '',
        auth_cache_seconds: 0,
        auth_token: '',
        api_key_header: 'X-API-Key',
        api_key_value: '',
        username: '',
        password: '',
        custom_headers: '',
        retry_limit: 5,
        enabled: true,
      };
      this.showDestModal = true;
    },
    openEditDestinationModal(dest) {
      this.editingDest = {
        dispatch_interval_seconds: 0,
        auth_url: '',
        auth_token_path: '',
        auth_cache_seconds: 0,
        api_key_header: 'X-API-Key',
        custom_headers: '',
        ...JSON.parse(JSON.stringify(dest)),
      };
      this.showDestModal = true;
    },
    onEditingAuthTypeChange() {
      if (this.editingDest.auth_type === 'SPARING_JWT_2STEP') {
        if (!this.editingDest.interval_seconds) {
          this.editingDest.interval_seconds = 120;
        }
        if (!this.editingDest.id && !this.editingDest.dispatch_interval_seconds) {
          this.editingDest.dispatch_interval_seconds = 3600;
        }
        if (!this.editingDest.auth_url) {
          this.editingDest.auth_url = 'https://sparing.kemenlh.go.id/api/secret-sensor';
        }
        if (!this.editingDest.endpoint) {
          this.editingDest.endpoint = 'https://sparing.kemenlh.go.id/api/send';
        }
      } else if (this.editingDest.auth_type === 'SPARING_JWT_STATIC') {
        if (!this.editingDest.interval_seconds) {
          this.editingDest.interval_seconds = 120;
        }
        if (!this.editingDest.endpoint) {
          this.editingDest.endpoint = 'https://sparing.kemenlh.go.id/api/send';
        }
      } else if (this.editingDest.auth_type === 'KLHK_ONLIMO_BODY') {
        if (!this.editingDest.interval_seconds) {
          this.editingDest.interval_seconds = 3600;
        }
        this.editingDest.dispatch_interval_seconds = 0;
      }
    },
    async saveDestination() {
      try {
        let requeued = 0;
        if (this.editingDest.id) {
          const res = await axios.put(`/api/destinations/${this.editingDest.id}`, this.editingDest);
          requeued = Number(res.data?.data?.requeued_batches || 0);
        } else {
          await axios.post('/api/destinations', this.editingDest);
        }
        this.showDestModal = false;
        this.fetchDestinations();
        this.fetchSummary();
        this.fetchQueue();
        if (requeued > 0) {
          alert(`Destinasi berhasil disimpan! ${requeued} antrean batch otomatis dijadwalkan kirim ulang dengan kredensial/header terbaru.`);
        }
      } catch (err) {
        alert('Failed to save destination: ' + (err.response?.data?.error || err.message));
      }
    },
    async toggleDestinationEnabled(dest) {
      try {
        const updated = { ...dest, enabled: !dest.enabled };
        await axios.put(`/api/destinations/${dest.id}`, updated);
        dest.enabled = updated.enabled;
        this.fetchSummary();
        this.fetchQueue();
      } catch (err) {
        alert('Toggle failed: ' + err.message);
      }
    },
    async deleteDestination(id) {
      if (!confirm('Hapus konfigurasi destinasi integrasi ini? Antrean pengiriman dan riwayat log tetap tersimpan di database untuk dikelola atau dihapus manual oleh Admin.')) return;
      try {
        await axios.delete(`/api/destinations/${id}`);
        await this.fetchDestinations();
        this.fetchSummary();
        this.fetchQueue();
        if (this.selectedDestinationId === id) {
          this.selectedDestinationId = this.destinations.length > 0 ? this.destinations[0].id : null;
          if (this.selectedDestinationId) {
            this.loadDestinationPayloadConfig();
          }
        }
      } catch (err) {
        alert('Delete failed: ' + (err.response?.data?.error || err.message));
      }
    },
    applyTemplatePreset() {
      const p = this.payloadConfig.template_type;
      if (p !== 'KLHK_WQMS_ONLIMO') {
        this.payloadConfig.fill_missing_preset_fields = false;
      }
      if (p === 'KLHK_WQMS_ONLIMO') {
        this.payloadConfig.station_id_field = 'IDStasiun';
        this.payloadConfig.station_id_value = this.payloadConfig.station_id_value || '';
        this.payloadConfig.timestamp_field = 'Tanggal,Jam';
        this.payloadConfig.timestamp_format = 'CUSTOM';
        this.payloadConfig.timestamp_source = 'PERIOD_END';
        this.payloadConfig.timestamp_timezone = 'Asia/Jakarta';
        this.payloadConfig.numeric_precision_mode = 'NATURAL_2';
        this.payloadConfig.omit_interval_field = true;
        this.payloadConfig.structure = 'ONLIMO_DATA_OBJECT';
        this.payloadConfig.nested_data_key = 'data';
        this.payloadConfig.include_device_identity = false;
        this.payloadConfig.device_identity_mode = 'NONE';
        this.payloadConfig.conflict_resolution = 'PRIORITY_PRIMARY';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
        this.payloadConfig.include_metadata = false;
        this.payloadConfig.include_sensor_stat_flags = false;
        this.payloadConfig.fill_missing_preset_fields = true;

        for (const m of this.parameterMappings || []) {
          const wqmsAlias = this.resolveKLHKWQMSAlias(m.parameter_code, m.parameter_name);
          if (wqmsAlias) {
            m.external_name = wqmsAlias;
            m.enabled = true;
          } else {
            m.enabled = false;
          }
        }
      } else if (p === 'KLHK_ISPU_AQMS') {
        this.payloadConfig.station_id_field = 'id_stasiun';
        this.payloadConfig.station_id_value = this.payloadConfig.station_id_value || 'KABUPATEN_JOMBANG';
        this.payloadConfig.timestamp_field = 'waktu';
        this.payloadConfig.timestamp_format = 'CUSTOM'; // YYYY-MM-DD HH:mm:ss
        this.payloadConfig.timestamp_timezone = 'Asia/Jakarta';
        this.payloadConfig.omit_interval_field = true;
        this.payloadConfig.structure = 'FLAT';
        this.payloadConfig.include_device_identity = false;
        this.payloadConfig.device_identity_mode = 'NONE';
        this.payloadConfig.conflict_resolution = 'PRIORITY_PRIMARY';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
        this.payloadConfig.include_metadata = false;
        this.payloadConfig.include_sensor_stat_flags = true;
        this.payloadConfig.sensor_stat_mode = 'STATIC_1';

        const aliasMap = this.getPresetAliasMap('KLHK_ISPU_AQMS');
        for (const m of this.parameterMappings || []) {
          const lower = (m.parameter_code || '').toLowerCase().trim();
          if (aliasMap[lower]) {
            m.external_name = aliasMap[lower];
            m.enabled = true;
          }
        }
      } else if (p === 'MULTI_TICK_SENSOR') {
        this.payloadConfig.station_id_field = 'station_id';
        this.payloadConfig.station_id_value = this.payloadConfig.station_id_value || '3453';
        this.payloadConfig.module_id_value = this.payloadConfig.module_id_value || this.payloadConfig.station_id_value || '3453';
        this.payloadConfig.timestamp_field = 'created_at';
        this.payloadConfig.timestamp_format = 'CUSTOM'; // YYYY-MM-DD HH:mm:ss
        this.payloadConfig.omit_interval_field = false;
        this.payloadConfig.structure = 'MULTI_TICK_ARRAY';
        this.payloadConfig.include_device_identity = false;
        this.payloadConfig.device_identity_mode = 'NONE';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
        this.payloadConfig.include_metadata = false;
        this.payloadConfig.include_sensor_stat_flags = false;

        const aliasMap = this.getPresetAliasMap('MULTI_TICK_SENSOR');
        for (const m of this.parameterMappings || []) {
          const lower = (m.parameter_code || '').toLowerCase().trim();
          if (aliasMap[lower]) {
            m.external_name = aliasMap[lower];
            m.enabled = true;
          }
        }
      } else if (p === 'MULTI_DEVICE_FLAT') {
        this.payloadConfig.station_id_field = 'station_id';
        this.payloadConfig.timestamp_field = 'timestamp';
        this.payloadConfig.timestamp_format = 'ISO8601';
        this.payloadConfig.interval_field = 'interval_minutes';
        this.payloadConfig.omit_interval_field = false;
        this.payloadConfig.structure = 'FLAT';
        this.payloadConfig.nested_data_key = 'data';
        this.payloadConfig.include_device_identity = false;
        this.payloadConfig.device_identity_mode = 'NONE';
        this.payloadConfig.conflict_resolution = 'PREFIX';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
        this.payloadConfig.include_sensor_stat_flags = false;
      } else if (p === 'MULTI_DEVICE_NESTED') {
        this.payloadConfig.station_id_field = 'station_id';
        this.payloadConfig.timestamp_field = 'timestamp';
        this.payloadConfig.timestamp_format = 'ISO8601';
        this.payloadConfig.interval_field = 'interval_minutes';
        this.payloadConfig.omit_interval_field = false;
        this.payloadConfig.structure = 'NESTED_BY_DEVICE';
        this.payloadConfig.nested_data_key = 'data';
        this.payloadConfig.include_device_identity = true;
        this.payloadConfig.device_identity_mode = 'NESTED_BY_DEVICE';
        this.payloadConfig.conflict_resolution = 'FAIL_ON_CONFLICT';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
        this.payloadConfig.include_sensor_stat_flags = false;
      } else if (p === 'MULTI_DEVICE_ARRAY') {
        this.payloadConfig.station_id_field = 'station_id';
        this.payloadConfig.timestamp_field = 'timestamp';
        this.payloadConfig.timestamp_format = 'ISO8601';
        this.payloadConfig.interval_field = 'interval_minutes';
        this.payloadConfig.omit_interval_field = false;
        this.payloadConfig.structure = 'ARRAY_BY_DEVICE';
        this.payloadConfig.nested_data_key = 'devices';
        this.payloadConfig.include_device_identity = true;
        this.payloadConfig.device_identity_mode = 'ARRAY_BY_DEVICE';
        this.payloadConfig.conflict_resolution = 'FAIL_ON_CONFLICT';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
        this.payloadConfig.include_sensor_stat_flags = false;
      } else if (p === 'KLHK_SPARING_HOURLY' || p === 'KLHK_SPARING') {
        this.payloadConfig.station_id_field = 'uid';
        this.payloadConfig.station_id_value = this.payloadConfig.station_id_value || '';
        this.payloadConfig.timestamp_field = 'datetime';
        this.payloadConfig.timestamp_format = 'UNIX_SEC';
        this.payloadConfig.timestamp_source = 'PERIOD_START';
        this.payloadConfig.timestamp_timezone = '';
        this.payloadConfig.numeric_precision_mode = 'NATURAL_2';
        this.payloadConfig.omit_interval_field = true;
        this.payloadConfig.structure = p === 'KLHK_SPARING_HOURLY' ? 'SPARING_DATA_ARRAY' : 'FLAT';
        this.payloadConfig.nested_data_key = 'data';
        this.payloadConfig.include_device_identity = false;
        this.payloadConfig.device_identity_mode = 'NONE';
        this.payloadConfig.conflict_resolution = 'PRIORITY_PRIMARY';
        this.payloadConfig.include_units = false;
        this.payloadConfig.include_quality = false;
        this.payloadConfig.include_metadata = false;
        this.payloadConfig.include_sensor_stat_flags = false;

        const aliasMap = this.getPresetAliasMap(p);
        for (const m of this.parameterMappings || []) {
          const lower = (m.parameter_code || '').toLowerCase().trim();
          if (aliasMap[lower]) {
            m.external_name = aliasMap[lower];
            m.enabled = true;
          }
        }
      } else if (p === 'NESTED_SCADA') {
        this.payloadConfig.station_id_field = 'stationCode';
        this.payloadConfig.timestamp_field = 'datetime';
        this.payloadConfig.timestamp_format = 'UNIX_MS';
        this.payloadConfig.interval_field = 'periodMin';
        this.payloadConfig.omit_interval_field = false;
        this.payloadConfig.structure = 'NESTED';
        this.payloadConfig.nested_data_key = 'measurements';
        this.payloadConfig.include_device_identity = true;
        this.payloadConfig.device_identity_mode = 'FIELD';
        this.payloadConfig.conflict_resolution = 'PREFIX';
        this.payloadConfig.include_units = true;
        this.payloadConfig.include_quality = true;
        this.payloadConfig.include_sensor_stat_flags = false;
      } else {
        this.payloadConfig.structure = 'FLAT';
        this.payloadConfig.omit_interval_field = false;
        this.payloadConfig.include_sensor_stat_flags = false;
      }
      this.triggerLivePreview();
    },
    formatQueueIntervalBadge(item) {
      if (!item) return '-';
      const srcSec = Number(item.interval_seconds) || 0;
      const srcMin = srcSec > 0 ? Math.max(1, Math.round(srcSec / 60)) : 0;
      if (item.period_start && item.period_end && srcSec > 0) {
        const startMs = new Date(item.period_start).getTime();
        const endMs = new Date(item.period_end).getTime();
        if (!isNaN(startMs) && !isNaN(endMs) && endMs > startMs) {
          const spanSec = Math.round((endMs - startMs) / 1000);
          if (spanSec > srcSec * 1.5) {
            const bucketCount = Math.max(1, Math.round(spanSec / srcSec));
            const spanMin = Math.max(1, Math.round(spanSec / 60));
            return `${srcMin}m × ${bucketCount} (${spanMin}m)`;
          }
        }
      }
      return srcMin > 0 ? `${srcMin}m` : 'ALL';
    },
    formatDestinationSchedule(dest) {
      if (!dest) return 'All Intervals';
      const srcSec = Number(dest.interval_seconds) || 0;
      const dispSec = Number(dest.dispatch_interval_seconds) || 0;
      if (dispSec > 0) {
        const effSrc = srcSec > 0 ? srcSec : 120;
        const srcMin = Math.max(1, Math.round(effSrc / 60));
        const dispMin = Math.max(1, Math.round(dispSec / 60));
        const count = Math.max(1, Math.round(dispSec / effSrc));
        return `${srcMin}m → Kirim tiap ${dispMin}m (${count} data)`;
      }
      if (srcSec > 0) {
        return `${Math.max(1, Math.round(srcSec / 60))} min (Langsung)`;
      }
      return 'All Intervals';
    },
    formatAuthModeLabel(authType) {
      if (authType === 'LOGIN_BEARER') return '2-STEP LOGIN → BEARER';
      if (authType === 'SPARING_JWT_2STEP') return '2-STEP SECRET → JWT HS256';
      if (authType === 'SPARING_JWT_STATIC') return 'STATIC SECRET → JWT HS256';
      if (authType === 'KLHK_ONLIMO_BODY') return 'BODY APIKEY + SECRET (ONLIMO)';
      return authType || 'NONE';
    },
    formatDate(t) {
      if (!t) return '';
      try {
        const d = new Date(t);
        if (isNaN(d.getTime())) return '';
        return d.toLocaleDateString([], { day: '2-digit', month: 'short', year: 'numeric' });
      } catch (e) {
        return '';
      }
    },
    formatTime(t) {
      if (!t) return '-';
      try {
        const d = new Date(t);
        if (isNaN(d.getTime())) return '-';
        return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
      } catch (e) {
        return t;
      }
    },
    formatAggValue(val) {
      if (val === null || val === undefined || val === '') return 'null';
      const num = Number(val);
      if (isNaN(num)) return String(val);
      return num.toFixed(2);
    },
    formatPrettyJSON(str) {
      if (!str) return '';
      try {
        JSON.parse(str);
      } catch (e) {
        return str;
      }

      const preserveNaturalNumbers = /"uid"\s*:|"IDStasiun"\s*:/.test(str);
      let out = '';
      let depth = 0;
      const len = str.length;
      let i = 0;
      const skipWs = (idx) => {
        while (idx < len && /\s/.test(str[idx])) idx++;
        return idx;
      };

      while (i < len) {
        const ch = str[i];
        if (/\s/.test(ch)) {
          i++;
          continue;
        }
        if (ch === '"') {
          let s = '"';
          i++;
          while (i < len) {
            const c = str[i];
            s += c;
            i++;
            if (c === '\\' && i < len) {
              s += str[i];
              i++;
            } else if (c === '"') {
              break;
            }
          }
          out += s;
        } else if (ch === '{' || ch === '[') {
          const closeCh = ch === '{' ? '}' : ']';
          const nextIdx = skipWs(i + 1);
          if (nextIdx < len && str[nextIdx] === closeCh) {
            out += ch + closeCh;
            i = nextIdx + 1;
          } else {
            depth++;
            out += ch + '\n' + '  '.repeat(depth);
            i++;
          }
        } else if (ch === '}' || ch === ']') {
          depth = Math.max(0, depth - 1);
          out += '\n' + '  '.repeat(depth) + ch;
          i++;
        } else if (ch === ',') {
          out += ',\n' + '  '.repeat(depth);
          i++;
        } else if (ch === ':') {
          out += ': ';
          i++;
        } else if (ch === '-' || (ch >= '0' && ch <= '9')) {
          let numStr = '';
          while (i < len && /[-+0-9.eE]/.test(str[i])) {
            numStr += str[i];
            i++;
          }
          if (!preserveNaturalNumbers && (numStr.includes('.') || numStr.includes('e') || numStr.includes('E'))) {
            const parsedNum = Number(numStr);
            out += isNaN(parsedNum) ? numStr : parsedNum.toFixed(2);
          } else {
            out += numStr;
          }
        } else {
          out += ch;
          i++;
        }
      }
      return out;
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
