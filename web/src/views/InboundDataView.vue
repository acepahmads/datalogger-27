<template>
  <div class="space-y-3 font-sans text-slate-800 dark:text-slate-200">
    <!-- Unified Compact Command Header + Inline KPI Pills -->
    <div class="bg-white dark:bg-[#111827] border border-slate-200/90 dark:border-slate-800 rounded-2xl shadow-sm overflow-hidden">
      <div class="px-4 py-3 flex flex-col xl:flex-row xl:items-center justify-between gap-3">
        <!-- Identity -->
        <div class="flex items-center gap-3 min-w-0">
          <div class="w-9 h-9 rounded-xl bg-gradient-to-br from-emerald-500 to-cyan-600 text-white badge-text shadow-sm flex items-center justify-center flex-shrink-0">
            <svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
            </svg>
          </div>
          <div class="min-w-0">
            <h1 class="text-sm sm:text-base font-extrabold text-slate-900 dark:text-white tracking-tight flex items-center gap-2 flex-wrap">
              <span class="truncate">Inbound Data Monitor</span>
              <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-4xs font-mono font-bold bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border border-emerald-500/25">
                CLOUD RECEIVER · SUPERADMIN
              </span>
            </h1>
            <p class="hidden sm:block text-2xs text-slate-500 dark:text-slate-400 truncate mt-0.5">
              Stream audit &amp; inspeksi payload Raw JSON masuk dari seluruh tenant (HTTP JSON, Legacy CBI, SPARING 30×2m &amp; MQTT).
            </p>
          </div>
        </div>

        <!-- Inline KPI Pills + Live Actions -->
        <div class="flex flex-wrap items-center gap-2">
          <div class="grid grid-cols-4 gap-1.5 sm:flex sm:items-center sm:gap-1.5">
            <div class="px-2.5 py-1.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800 flex items-center gap-1.5">
              <span class="w-2 h-2 rounded-full bg-cyan-500 hidden sm:inline-block flex-shrink-0"></span>
              <div class="leading-none">
                <div class="text-4xs font-semibold uppercase tracking-wider text-slate-400">Total</div>
                <div class="text-sm font-extrabold font-mono text-slate-900 dark:text-white tabular-nums mt-0.5">{{ formatNum(summary.total_requests) }}</div>
              </div>
            </div>
            <div class="px-2.5 py-1.5 rounded-xl bg-emerald-50/70 dark:bg-emerald-950/25 border border-emerald-200/70 dark:border-emerald-800/50 flex items-center gap-1.5">
              <span class="w-2 h-2 rounded-full bg-emerald-500 hidden sm:inline-block flex-shrink-0"></span>
              <div class="leading-none">
                <div class="text-4xs font-semibold uppercase tracking-wider text-emerald-700/80 dark:text-emerald-400/80">Accepted</div>
                <div class="text-sm font-extrabold font-mono text-emerald-700 dark:text-emerald-400 tabular-nums mt-0.5">{{ formatNum(summary.accepted_count) }}</div>
              </div>
            </div>
            <div class="px-2.5 py-1.5 rounded-xl bg-rose-50/70 dark:bg-rose-950/25 border border-rose-200/70 dark:border-rose-800/50 flex items-center gap-1.5">
              <span class="w-2 h-2 rounded-full bg-rose-500 hidden sm:inline-block flex-shrink-0"></span>
              <div class="leading-none">
                <div class="text-4xs font-semibold uppercase tracking-wider text-rose-700/80 dark:text-rose-400/80">Rejected</div>
                <div class="text-sm font-extrabold font-mono text-rose-600 dark:text-rose-400 tabular-nums mt-0.5">{{ formatNum(summary.rejected_count) }}</div>
              </div>
            </div>
            <div class="px-2.5 py-1.5 rounded-xl bg-purple-50/70 dark:bg-purple-950/25 border border-purple-200/70 dark:border-purple-800/50 flex items-center gap-1.5">
              <span class="w-2 h-2 rounded-full bg-purple-500 hidden sm:inline-block flex-shrink-0"></span>
              <div class="leading-none">
                <div class="text-4xs font-semibold uppercase tracking-wider text-purple-700/80 dark:text-purple-400/80">Tenant</div>
                <div class="text-sm font-extrabold font-mono text-purple-700 dark:text-purple-400 tabular-nums mt-0.5">{{ formatNum(summary.active_tenants) }}</div>
              </div>
            </div>
          </div>

          <div class="flex items-center gap-1.5">
            <button
              id="btn-inbound-auto-refresh"
              type="button"
              @click="toggleAutoRefresh"
              class="px-2.5 py-1.5 rounded-xl text-2xs font-mono font-bold border transition cursor-pointer flex items-center gap-1.5"
              :class="autoRefresh
                ? 'bg-emerald-500/12 text-emerald-700 dark:text-emerald-300 border-emerald-500/35'
                : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-700'"
            >
              <span class="w-2 h-2 rounded-full" :class="autoRefresh ? 'bg-emerald-500 animate-pulse' : 'bg-slate-400'"></span>
              <span>{{ autoRefresh ? 'LIVE 5s' : 'PAUSED' }}</span>
            </button>

            <button
              id="btn-inbound-refresh"
              type="button"
              @click="fetchLogs(false)"
              :disabled="loading"
              class="px-2.5 py-1.5 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-white btn-text text-2xs font-bold shadow-2xs transition cursor-pointer disabled:opacity-50"
            >
              {{ loading ? '…' : '↻ Refresh' }}
            </button>

            <div class="relative">
              <button
                id="btn-inbound-purge-menu"
                type="button"
                @click="showPurgeMenu = !showPurgeMenu"
                class="px-2.5 py-1.5 rounded-xl bg-rose-50 hover:bg-rose-100 dark:bg-rose-950/40 dark:hover:bg-rose-900/50 text-rose-600 dark:text-rose-400 border border-rose-200/80 dark:border-rose-800/60 text-2xs font-bold transition cursor-pointer"
                title="Bersihkan log inbound lama"
              >
                🗑️
              </button>
              <div
                v-if="showPurgeMenu"
                class="absolute right-0 top-full mt-1.5 w-52 rounded-xl bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-700 shadow-xl z-40 py-1 text-2xs"
              >
                <button
                  type="button"
                  @click="purgeLogs(false, 7)"
                  class="w-full px-3 py-2 text-left hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-200 font-medium cursor-pointer"
                >
                  🧹 Hapus Log &gt; 7 Hari
                </button>
                <button
                  type="button"
                  @click="purgeLogs(false, 1)"
                  class="w-full px-3 py-2 text-left hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-200 font-medium cursor-pointer"
                >
                  🧹 Hapus Log &gt; 24 Jam
                </button>
                <button
                  type="button"
                  @click="purgeLogs(true, 0)"
                  class="w-full px-3 py-2 text-left hover:bg-rose-50 dark:hover:bg-rose-950/40 text-rose-600 dark:text-rose-400 font-bold border-t border-slate-100 dark:border-slate-800 cursor-pointer"
                >
                  ⚠️ Bersihkan Semua Log
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Compact Multi-Facet Filter Bar -->
      <div class="px-4 py-2.5 border-t border-slate-100 dark:border-slate-800/80 bg-slate-50/70 dark:bg-[#0B0F19]/60 space-y-2">
        <div class="flex flex-wrap items-center gap-2">
          <!-- Search -->
          <div class="relative flex-1 min-w-[200px]">
            <svg class="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <circle cx="11" cy="11" r="8" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-4.35-4.35" />
            </svg>
            <input
              id="filter-inbound-search"
              v-model="filters.search"
              @input="debouncedFetch"
              type="text"
              placeholder="Cari tenant, site_code, device, IP, error, atau isi JSON..."
              class="w-full pl-8 pr-3 py-1.5 rounded-xl bg-white dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-2xs text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-cyan-500/30 focus:border-cyan-500"
            />
          </div>

          <!-- Tenant Filter -->
          <select
            id="filter-inbound-tenant"
            v-model="filters.tenant_id"
            @change="onFilterChange"
            class="px-2.5 py-1.5 rounded-xl bg-white dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-2xs font-semibold text-slate-700 dark:text-slate-200 cursor-pointer max-w-[200px]"
          >
            <option value="">🏢 Semua Tenant ({{ tenants.length }})</option>
            <option v-for="t in tenants" :key="t.id" :value="String(t.id)">
              {{ t.name }}
            </option>
          </select>

          <!-- Status Filter -->
          <select
            id="filter-inbound-status"
            v-model="filters.status"
            @change="onFilterChange"
            class="px-2.5 py-1.5 rounded-xl bg-white dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-2xs font-semibold text-slate-700 dark:text-slate-200 cursor-pointer"
          >
            <option value="">Semua Status</option>
            <option value="ACCEPTED">✓ ACCEPTED</option>
            <option value="REJECTED">✕ REJECTED</option>
          </select>

          <!-- Logger Domain -->
          <select
            id="filter-inbound-logger"
            v-model="filters.logger_type"
            @change="onFilterChange"
            class="px-2.5 py-1.5 rounded-xl bg-white dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-2xs font-semibold text-slate-700 dark:text-slate-200 cursor-pointer"
          >
            <option value="">Semua Domain</option>
            <option value="AQMS">AQMS</option>
            <option value="SPARING">SPARING</option>
            <option value="ONLIMO">ONLIMO</option>
            <option value="WQMS">WQMS</option>
          </select>

          <!-- Format -->
          <select
            id="filter-inbound-format"
            v-model="filters.format"
            @change="onFilterChange"
            class="px-2.5 py-1.5 rounded-xl bg-white dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-2xs font-semibold text-slate-700 dark:text-slate-200 cursor-pointer"
          >
            <option value="">Semua Format</option>
            <option value="STANDARD_V1">STANDARD_V1</option>
            <option value="CBI_LEGACY">CBI_LEGACY</option>
            <option value="KLHK_SPARING">KLHK_SPARING</option>
          </select>

          <!-- Source -->
          <select
            id="filter-inbound-source"
            v-model="filters.source"
            @change="onFilterChange"
            class="px-2.5 py-1.5 rounded-xl bg-white dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-2xs font-semibold text-slate-700 dark:text-slate-200 cursor-pointer"
          >
            <option value="">HTTP &amp; MQTT</option>
            <option value="EDGE_PUSH">HTTP Push</option>
            <option value="MQTT">MQTT</option>
          </select>
        </div>

        <!-- Quick Time Presets + Active Filter Reset -->
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex flex-wrap items-center gap-1">
            <span class="text-4xs font-bold uppercase tracking-wider text-slate-400 mr-1">Waktu:</span>
            <button
              v-for="p in timePresets"
              :key="p.id"
              type="button"
              @click="selectTimePreset(p.id)"
              class="px-2 py-0.5 rounded-lg text-3xs font-semibold border transition cursor-pointer"
              :class="activeTimePreset === p.id
                ? 'bg-cyan-600 text-white btn-text border-cyan-600 shadow-2xs'
                : 'bg-white dark:bg-slate-800/80 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:border-cyan-400'"
            >
              {{ p.label }}
            </button>
          </div>

          <div class="flex items-center gap-2 text-3xs text-slate-500">
            <span v-if="summary.total_records > 0 || summary.total_params > 0" class="font-mono hidden sm:inline">
              {{ formatNum(summary.total_records) }} rec · {{ formatNum(summary.total_params) }} param tersimpan
            </span>
            <button
              v-if="hasActiveFilters"
              type="button"
              @click="resetFilters"
              class="px-2 py-0.5 rounded-lg bg-slate-200/80 hover:bg-slate-300/80 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 font-bold cursor-pointer"
            >
              ✕ Reset Filter
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Feedback Toast -->
    <div
      v-if="toastMessage"
      class="px-3.5 py-2 rounded-xl text-xs font-semibold flex items-center justify-between shadow-xs border"
      :class="toastType === 'error'
        ? 'bg-rose-500/10 border-rose-500/30 text-rose-600 dark:text-rose-400'
        : 'bg-emerald-500/10 border-emerald-500/30 text-emerald-700 dark:text-emerald-400'"
    >
      <span>{{ toastMessage }}</span>
      <button type="button" @click="toastMessage = ''" class="text-sm leading-none cursor-pointer">&times;</button>
    </div>

    <!-- Split-Pane Master (Compact Table) + Detail (Sticky Raw JSON Inspector) -->
    <div class="grid grid-cols-1 xl:grid-cols-12 gap-3 items-start">
      <!-- LEFT: Compact Stream Table -->
      <div class="xl:col-span-7 bg-white dark:bg-[#111827] border border-slate-200/90 dark:border-slate-800 rounded-2xl shadow-sm overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-4xs font-bold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-3 py-2.5">Waktu</th>
                <th class="px-2.5 py-2.5">Tenant &amp; Stasiun</th>
                <th class="px-2.5 py-2.5">Domain &amp; Format</th>
                <th class="px-2.5 py-2.5">Status</th>
                <th class="px-3 py-2.5 text-right">Beban</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
              <tr
                v-for="item in items"
                :key="item.id"
                @click="selectLogItem(item)"
                class="transition-colors cursor-pointer"
                :class="selectedLog && Number(selectedLog.id) === Number(item.id)
                  ? 'bg-cyan-50/70 dark:bg-cyan-950/30'
                  : 'hover:bg-slate-50/80 dark:hover:bg-slate-800/40'"
              >
                <!-- Time -->
                <td class="px-3 py-2 whitespace-nowrap font-mono">
                  <div class="text-2xs font-bold text-slate-800 dark:text-slate-100">{{ formatClock(item.received_at) }}</div>
                  <div class="text-4xs text-slate-400">{{ formatShortDate(item.received_at) }} · {{ formatRelative(item.received_at) }}</div>
                </td>

                <!-- Tenant & Site / Device -->
                <td class="px-2.5 py-2 max-w-[230px]">
                  <div class="font-bold text-2xs text-slate-900 dark:text-white truncate flex items-center gap-1.5">
                    <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="item.status === 'ACCEPTED' ? 'bg-emerald-500' : 'bg-rose-500'"></span>
                    <span class="truncate">{{ item.tenant_name || (item.tenant_id ? ('Tenant #' + item.tenant_id) : 'Unauthenticated') }}</span>
                  </div>
                  <div class="flex items-center gap-1 mt-0.5 flex-wrap">
                    <span v-if="item.site_code" class="px-1.5 py-0.2 rounded bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 font-mono text-4xs truncate max-w-[130px]">
                      {{ item.site_code }}
                    </span>
                    <span v-if="item.device_code && item.device_code !== item.site_code" class="px-1.5 py-0.2 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 font-mono text-4xs truncate max-w-[110px]">
                      {{ item.device_code }}
                    </span>
                  </div>
                  <div v-if="item.status !== 'ACCEPTED' && item.error_message" class="text-4xs text-rose-600 dark:text-rose-400 truncate mt-0.5">
                    {{ item.error_message }}
                  </div>
                </td>

                <!-- Domain & Format -->
                <td class="px-2.5 py-2 whitespace-nowrap">
                  <div class="flex items-center gap-1 flex-wrap">
                    <span
                      v-if="item.logger_type"
                      class="px-1.5 py-0.2 rounded-full text-4xs font-mono font-bold border"
                      :class="domainBadgeClass(item.logger_type)"
                    >
                      {{ item.logger_type }}
                    </span>
                    <span class="px-1.5 py-0.2 rounded text-4xs font-mono font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                      {{ item.detected_format || item.source || 'HTTP' }}
                    </span>
                  </div>
                  <div class="text-4xs font-mono text-slate-400 truncate max-w-[160px] mt-0.5" :title="item.endpoint">
                    {{ item.endpoint }}
                  </div>
                </td>

                <!-- Status -->
                <td class="px-2.5 py-2 whitespace-nowrap">
                  <span
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-4xs font-mono font-bold border"
                    :class="item.status === 'ACCEPTED'
                      ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border-emerald-500/30'
                      : 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30'"
                  >
                    <span>{{ item.http_status || (item.status === 'ACCEPTED' ? 200 : 400) }}</span>
                    <span>{{ item.status }}</span>
                  </span>
                  <div class="text-4xs font-mono text-slate-400 mt-0.5">
                    {{ item.latency_ms }}ms · {{ item.client_ip || '-' }}
                  </div>
                </td>

                <!-- Load / Size -->
                <td class="px-3 py-2 text-right whitespace-nowrap font-mono">
                  <div class="text-2xs font-bold text-slate-800 dark:text-slate-200">{{ formatBytes(item.payload_bytes) }}</div>
                  <div class="text-4xs text-slate-400">
                    <template v-if="item.status === 'ACCEPTED'">
                      {{ item.records_processed }}r · {{ item.parameters_ingested }}p
                    </template>
                    <template v-else>
                      JSON →
                    </template>
                  </div>
                </td>
              </tr>

              <tr v-if="!loading && items.length === 0">
                <td colspan="5" class="px-4 py-12 text-center text-slate-500 dark:text-slate-400">
                  <div class="text-2xl mb-1">📡</div>
                  <div class="font-bold text-xs text-slate-700 dark:text-slate-300">Belum ada payload inbound pada filter ini</div>
                  <p class="text-2xs text-slate-400 mt-0.5">
                    Setiap push JSON ke <code class="font-mono">/api/v1/ingest/telemetry</code>, <code class="font-mono">/api/v1/ticks/create_multi</code>, atau <code class="font-mono">/api/v1/ingest/mqtt</code> akan tercatat otomatis di sini.
                  </p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div class="px-3.5 py-2.5 border-t border-slate-100 dark:border-slate-800/80 bg-slate-50/60 dark:bg-[#0B0F19]/50 flex flex-wrap items-center justify-between gap-2 text-2xs">
          <div class="text-slate-500 dark:text-slate-400 font-mono">
            Menampilkan {{ items.length }} dari {{ formatNum(total) }} log · Hal {{ page }}/{{ totalPages }}
          </div>
          <div class="flex items-center gap-1.5">
            <select
              v-model.number="pageSize"
              @change="onFilterChange"
              class="px-2 py-1 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-3xs font-mono cursor-pointer"
            >
              <option :value="25">25 / hal</option>
              <option :value="50">50 / hal</option>
              <option :value="100">100 / hal</option>
            </select>
            <button
              type="button"
              :disabled="page <= 1"
              @click="goToPage(page - 1)"
              class="px-2.5 py-1 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 font-semibold cursor-pointer disabled:opacity-40"
            >
              ← Prev
            </button>
            <button
              type="button"
              :disabled="page >= totalPages"
              @click="goToPage(page + 1)"
              class="px-2.5 py-1 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 font-semibold cursor-pointer disabled:opacity-40"
            >
              Next →
            </button>
          </div>
        </div>
      </div>

      <!-- RIGHT: Sticky Raw JSON Payload Inspector (Desktop xl+) -->
      <div class="hidden xl:flex xl:col-span-5 bg-white dark:bg-[#111827] border border-slate-200/90 dark:border-slate-800 rounded-2xl shadow-sm overflow-hidden flex-col sticky top-3">
        <template v-if="selectedLog">
          <!-- Inspector Header -->
          <div class="px-4 py-3 border-b border-slate-100 dark:border-slate-800 bg-slate-50/70 dark:bg-[#0B0F19]/60 space-y-2">
            <div class="flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span
                    class="px-2 py-0.5 rounded-full text-4xs font-mono font-bold border"
                    :class="selectedLog.status === 'ACCEPTED'
                      ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/30'
                      : 'bg-rose-500/15 text-rose-600 dark:text-rose-300 border-rose-500/30'"
                  >
                    HTTP {{ selectedLog.http_status }} · {{ selectedLog.status }}
                  </span>
                  <span class="text-xs font-extrabold text-slate-900 dark:text-white truncate">
                    {{ selectedLog.tenant_name || (selectedLog.tenant_id ? ('Tenant #' + selectedLog.tenant_id) : 'Unauthenticated') }}
                  </span>
                </div>
                <div class="text-3xs font-mono text-slate-500 dark:text-slate-400 mt-1 flex flex-wrap items-center gap-x-2 gap-y-0.5">
                  <span>{{ formatFullTime(selectedLog.received_at) }}</span>
                  <span>· {{ selectedLog.endpoint }}</span>
                  <span>· {{ formatBytes(selectedLog.payload_bytes) }}</span>
                  <span>· {{ selectedLog.latency_ms }}ms</span>
                </div>
              </div>
            </div>

            <!-- Compact Metadata Pills -->
            <div class="flex flex-wrap items-center gap-1 text-4xs font-mono">
              <span v-if="selectedLog.site_code" class="px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border border-cyan-500/20">
                site: {{ selectedLog.site_code }}
              </span>
              <span v-if="selectedLog.device_code" class="px-1.5 py-0.5 rounded bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border border-indigo-500/20">
                dev: {{ selectedLog.device_code }}
              </span>
              <span v-if="selectedLog.key_prefix" class="px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border border-emerald-500/20">
                key: {{ selectedLog.key_prefix }}
              </span>
              <span v-if="selectedLog.client_ip" class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">
                ip: {{ selectedLog.client_ip }}
              </span>
              <span v-if="selectedLog.status === 'ACCEPTED'" class="px-1.5 py-0.5 rounded bg-purple-500/10 text-purple-700 dark:text-purple-300">
                {{ selectedLog.records_processed }} rec · {{ selectedLog.parameters_ingested }} in / {{ selectedLog.parameters_skipped }} skip
              </span>
            </div>

            <!-- Error Callout if Rejected -->
            <div
              v-if="selectedLog.error_message"
              class="p-2 rounded-xl bg-rose-50 dark:bg-rose-950/35 border border-rose-200 dark:border-rose-800/60 text-3xs text-rose-700 dark:text-rose-300 font-mono break-all"
            >
              ⚠️ {{ selectedLog.error_message }}
            </div>

            <!-- Viewer Mode Switch + Copy/Download -->
            <div class="flex flex-wrap items-center justify-between gap-2 pt-1">
              <div class="inline-flex items-center p-0.5 rounded-lg bg-slate-200/70 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-4xs font-bold">
                <button
                  id="btn-inbound-view-pretty"
                  type="button"
                  @click="jsonViewMode = 'pretty'"
                  class="px-2 py-0.5 rounded-md transition cursor-pointer"
                  :class="jsonViewMode === 'pretty' ? 'bg-emerald-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-300'"
                >
                  Pretty JSON
                </button>
                <button
                  id="btn-inbound-view-human"
                  type="button"
                  @click="jsonViewMode = 'human'"
                  class="px-2 py-0.5 rounded-md transition cursor-pointer"
                  :class="jsonViewMode === 'human' ? 'bg-cyan-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-300'"
                >
                  🕒 Waktu Terbaca
                </button>
                <button
                  id="btn-inbound-view-raw"
                  type="button"
                  @click="jsonViewMode = 'raw'"
                  class="px-2 py-0.5 rounded-md transition cursor-pointer"
                  :class="jsonViewMode === 'raw' ? 'bg-purple-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-300'"
                >
                  Raw
                </button>
              </div>

              <div class="flex items-center gap-1.5">
                <button
                  type="button"
                  @click="copySelectedJson"
                  class="px-2.5 py-1 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-3xs font-bold text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition cursor-pointer"
                >
                  📋 Copy
                </button>
                <button
                  type="button"
                  @click="downloadSelectedJson"
                  class="px-2.5 py-1 rounded-lg bg-teal-600 hover:bg-teal-500 text-white btn-text text-3xs font-bold transition cursor-pointer"
                >
                  ⬇️ .json
                </button>
              </div>
            </div>
          </div>

          <!-- Raw JSON Terminal Viewer -->
          <div class="p-3 bg-slate-950 flex-1">
            <pre class="text-emerald-300 font-mono text-2xs overflow-auto max-h-[560px] leading-relaxed whitespace-pre-wrap break-all select-all">{{ formattedSelectedPayload }}</pre>
          </div>
        </template>

        <div v-else class="p-10 text-center text-slate-400">
          <div class="text-3xl mb-2">🔍</div>
          <div class="text-xs font-bold text-slate-600 dark:text-slate-300">Pilih Baris Inbound di Sebelah Kiri</div>
          <p class="text-2xs mt-1">Klik salah satu log untuk menginspeksi payload Raw JSON lengkapnya di sini.</p>
        </div>
      </div>
    </div>

    <!-- Mobile / Tablet Raw JSON Inspector Modal (< xl) -->
    <div
      v-if="mobileInspectorOpen && selectedLog"
      @click.self="mobileInspectorOpen = false"
      class="xl:hidden fixed inset-0 z-[75] flex items-center justify-center p-3 bg-slate-950/70 backdrop-blur-xs"
    >
      <div class="bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[88vh] flex flex-col overflow-hidden">
        <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between gap-2">
          <div class="min-w-0">
            <div class="font-extrabold text-xs text-slate-900 dark:text-white truncate">
              {{ selectedLog.tenant_name || ('Tenant #' + selectedLog.tenant_id) }} · {{ selectedLog.site_code || selectedLog.endpoint }}
            </div>
            <div class="text-4xs font-mono text-slate-400">
              {{ formatFullTime(selectedLog.received_at) }} · HTTP {{ selectedLog.http_status }} · {{ formatBytes(selectedLog.payload_bytes) }}
            </div>
          </div>
          <button
            type="button"
            @click="mobileInspectorOpen = false"
            class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 text-xs font-bold text-slate-600 dark:text-slate-300 cursor-pointer"
          >
            ✕
          </button>
        </div>
        <div class="px-4 py-2 bg-slate-50 dark:bg-[#0B0F19] border-b border-slate-200 dark:border-slate-800 flex flex-wrap items-center justify-between gap-2">
          <div class="inline-flex items-center p-0.5 rounded-lg bg-slate-200/70 dark:bg-slate-800 text-4xs font-bold">
            <button type="button" @click="jsonViewMode = 'pretty'" class="px-2 py-0.5 rounded-md cursor-pointer" :class="jsonViewMode === 'pretty' ? 'bg-emerald-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-300'">Pretty</button>
            <button type="button" @click="jsonViewMode = 'human'" class="px-2 py-0.5 rounded-md cursor-pointer" :class="jsonViewMode === 'human' ? 'bg-cyan-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-300'">🕒 Waktu</button>
            <button type="button" @click="jsonViewMode = 'raw'" class="px-2 py-0.5 rounded-md cursor-pointer" :class="jsonViewMode === 'raw' ? 'bg-purple-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-300'">Raw</button>
          </div>
          <div class="flex items-center gap-1.5">
            <button type="button" @click="copySelectedJson" class="px-2.5 py-1 rounded-lg bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-3xs font-bold cursor-pointer">📋 Copy</button>
            <button type="button" @click="downloadSelectedJson" class="px-2.5 py-1 rounded-lg bg-teal-600 text-white btn-text text-3xs font-bold cursor-pointer">⬇️ .json</button>
          </div>
        </div>
        <div class="p-3 bg-slate-950 flex-1 overflow-auto">
          <pre class="text-emerald-300 font-mono text-2xs leading-relaxed whitespace-pre-wrap break-all select-all">{{ formattedSelectedPayload }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'InboundDataView',
  data() {
    return {
      loading: false,
      autoRefresh: true,
      refreshTimer: null,
      searchTimer: null,
      showPurgeMenu: false,
      mobileInspectorOpen: false,
      jsonViewMode: 'pretty',
      toastMessage: '',
      toastType: 'success',

      tenants: [],
      items: [],
      total: 0,
      page: 1,
      pageSize: 25,
      selectedLog: null,
      summary: {
        total_requests: 0,
        accepted_count: 0,
        rejected_count: 0,
        total_records: 0,
        total_params: 0,
        active_tenants: 0,
        last_received_at: null,
      },

      activeTimePreset: '24h',
      timePresets: [
        { id: '15m', label: '15 Menit' },
        { id: '1h', label: '1 Jam' },
        { id: '6h', label: '6 Jam' },
        { id: '24h', label: '24 Jam' },
        { id: '7d', label: '7 Hari' },
        { id: 'all', label: 'Semua' },
      ],
      filters: {
        search: '',
        tenant_id: '',
        status: '',
        logger_type: '',
        format: '',
        source: '',
      },
    };
  },
  computed: {
    userRole() {
      const u = (this.$store && this.$store.state && this.$store.state.user) || {};
      const r = typeof u.role === 'string' ? u.role : ((u.role && u.role.name) || '');
      return String(r || '').trim().toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    isCloudMode() {
      const sys = (this.$store && this.$store.state && this.$store.state.systemStatus) || {};
      return Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
    },
    totalPages() {
      return Math.max(1, Math.ceil((this.total || 0) / (this.pageSize || 25)));
    },
    hasActiveFilters() {
      return Boolean(
        this.filters.search ||
        this.filters.tenant_id ||
        this.filters.status ||
        this.filters.logger_type ||
        this.filters.format ||
        this.filters.source ||
        this.activeTimePreset !== '24h'
      );
    },
    formattedSelectedPayload() {
      if (!this.selectedLog) return '';
      const raw = String(this.selectedLog.raw_payload || '');
      if (!raw) return '// (Empty body)';
      if (this.jsonViewMode === 'raw') {
        return raw;
      }
      let pretty = raw;
      try {
        pretty = JSON.stringify(JSON.parse(raw), null, 2);
      } catch (_) {
        pretty = raw;
      }
      if (this.jsonViewMode !== 'human') {
        return pretty;
      }
      const timeKeys = new Set([
        'datetime', 'timestamp', 'waktu', 'created_at', 'updated_at',
        'period_start', 'period_end', 'time', 'date', 'ts', 'epoch',
      ]);
      let out = pretty.replace(
        /("([^"]+)"\s*:\s*)(?:"?(\d{10}|\d{13})"?)(\s*[,}\r\n])/g,
        (fullMatch, prefix, keyName, digits, suffix) => {
          if (!timeKeys.has(String(keyName || '').toLowerCase())) return fullMatch;
          const num = Number(digits);
          if (!Number.isFinite(num)) return fullMatch;
          const ms = digits.length === 10 ? num * 1000 : num;
          const formatted = this.formatFullTime(new Date(ms));
          return `${prefix}"${formatted} (${digits})"${suffix}`;
        }
      );
      out = out.replace(
        /("([^"]+)"\s*:\s*)"(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2}))"(\s*[,}\r\n])/g,
        (fullMatch, prefix, keyName, isoVal, suffix) => {
          if (!timeKeys.has(String(keyName || '').toLowerCase())) return fullMatch;
          return `${prefix}"${this.formatFullTime(new Date(isoVal))}"${suffix}`;
        }
      );
      return out;
    },
  },
  mounted() {
    if (!this.isSuperAdmin || !this.isCloudMode) {
      this.$router.replace({ path: '/access-denied', query: { from: this.$route.fullPath } });
      return;
    }
    this.loadTenants();
    this.fetchLogs(false);
    this.startAutoRefresh();
  },
  beforeDestroy() {
    this.stopAutoRefresh();
    if (this.searchTimer) clearTimeout(this.searchTimer);
  },
  methods: {
    showToast(msg, type = 'success') {
      this.toastMessage = msg;
      this.toastType = type;
      setTimeout(() => {
        if (this.toastMessage === msg) this.toastMessage = '';
      }, 4500);
    },
    startAutoRefresh() {
      this.stopAutoRefresh();
      if (!this.autoRefresh) return;
      this.refreshTimer = setInterval(() => {
        if (!this.loading) this.fetchLogs(true);
      }, 5000);
    },
    stopAutoRefresh() {
      if (this.refreshTimer) {
        clearInterval(this.refreshTimer);
        this.refreshTimer = null;
      }
    },
    toggleAutoRefresh() {
      this.autoRefresh = !this.autoRefresh;
      if (this.autoRefresh) {
        this.fetchLogs(true);
        this.startAutoRefresh();
      } else {
        this.stopAutoRefresh();
      }
    },
    async loadTenants() {
      try {
        const res = await axios.get('/api/tenants');
        if (res.data && Array.isArray(res.data.data)) {
          this.tenants = res.data.data;
        }
      } catch (_) {
        this.tenants = [];
      }
    },
    computeStartTimeISO() {
      const now = Date.now();
      const mapMs = {
        '15m': 15 * 60 * 1000,
        '1h': 3600 * 1000,
        '6h': 6 * 3600 * 1000,
        '24h': 24 * 3600 * 1000,
        '7d': 7 * 24 * 3600 * 1000,
      };
      const dur = mapMs[this.activeTimePreset];
      return dur ? new Date(now - dur).toISOString() : '';
    },
    async fetchLogs(silent = false) {
      if (!silent) this.loading = true;
      try {
        const params = {
          page: this.page,
          page_size: this.pageSize,
        };
        if (this.filters.tenant_id) params.tenant_id = this.filters.tenant_id;
        if (this.filters.status) params.status = this.filters.status;
        if (this.filters.logger_type) params.logger_type = this.filters.logger_type;
        if (this.filters.format) params.format = this.filters.format;
        if (this.filters.source) params.source = this.filters.source;
        if (this.filters.search && this.filters.search.trim()) params.search = this.filters.search.trim();
        const startISO = this.computeStartTimeISO();
        if (startISO) params.start_time = startISO;

        const res = await axios.get('/api/superadmin/inbound-logs', { params });
        const d = (res.data && res.data.data) || {};
        this.items = Array.isArray(d.items) ? d.items : [];
        this.total = Number(d.total || 0);
        if (d.summary) {
          this.summary = d.summary;
        }
        if (this.items.length > 0) {
          if (!this.selectedLog) {
            this.selectedLog = this.items[0];
          } else {
            const refreshed = this.items.find(i => Number(i.id) === Number(this.selectedLog.id));
            if (refreshed) this.selectedLog = refreshed;
          }
        }
      } catch (err) {
        if (!silent) {
          const msg = (err.response && err.response.data && (err.response.data.error || err.response.data.message)) || err.message;
          this.showToast(`Gagal memuat inbound data: ${msg}`, 'error');
        }
      } finally {
        if (!silent) this.loading = false;
      }
    },
    debouncedFetch() {
      if (this.searchTimer) clearTimeout(this.searchTimer);
      this.searchTimer = setTimeout(() => {
        this.page = 1;
        this.fetchLogs(false);
      }, 300);
    },
    onFilterChange() {
      this.page = 1;
      this.fetchLogs(false);
    },
    selectTimePreset(id) {
      this.activeTimePreset = id;
      this.page = 1;
      this.fetchLogs(false);
    },
    resetFilters() {
      this.filters = {
        search: '',
        tenant_id: '',
        status: '',
        logger_type: '',
        format: '',
        source: '',
      };
      this.activeTimePreset = '24h';
      this.page = 1;
      this.fetchLogs(false);
    },
    goToPage(nextPage) {
      if (nextPage < 1 || nextPage > this.totalPages) return;
      this.page = nextPage;
      this.fetchLogs(false);
    },
    selectLogItem(item) {
      this.selectedLog = item;
      if (typeof window !== 'undefined' && window.innerWidth < 1280) {
        this.mobileInspectorOpen = true;
      }
    },
    async purgeLogs(clearAll, days) {
      this.showPurgeMenu = false;
      const label = clearAll
        ? 'SEMUA log inbound raw JSON'
        : `log inbound lebih lama dari ${days} hari`;
      if (!confirm(`Hapus ${label}? Tindakan ini tidak mempengaruhi data telemetri sensor yang sudah tersimpan.`)) {
        return;
      }
      try {
        const params = clearAll ? { all: '1' } : { older_than_days: days };
        const res = await axios.delete('/api/superadmin/inbound-logs', { params });
        const deleted = Number((res.data && res.data.data && res.data.data.deleted) || 0);
        this.showToast(`Berhasil menghapus ${this.formatNum(deleted)} baris log inbound.`);
        if (clearAll) this.selectedLog = null;
        this.page = 1;
        await this.fetchLogs(false);
      } catch (err) {
        const msg = (err.response && err.response.data && (err.response.data.error || err.response.data.message)) || err.message;
        this.showToast(`Gagal membersihkan log: ${msg}`, 'error');
      }
    },
    async copySelectedJson() {
      if (!this.selectedLog) return;
      const text = this.jsonViewMode === 'raw' ? String(this.selectedLog.raw_payload || '') : this.formattedSelectedPayload;
      try {
        await navigator.clipboard.writeText(text);
        this.showToast('Payload JSON berhasil disalin ke clipboard.');
      } catch (_) {
        this.showToast('Gagal menyalin ke clipboard.', 'error');
      }
    },
    downloadSelectedJson() {
      if (!this.selectedLog) return;
      const raw = String(this.selectedLog.raw_payload || '');
      let content = raw;
      try {
        content = JSON.stringify(JSON.parse(raw), null, 2);
      } catch (_) {
        content = raw;
      }
      const blob = new Blob([content], { type: 'application/json;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      const code = (this.selectedLog.site_code || this.selectedLog.device_code || 'inbound').replace(/[^a-zA-Z0-9_-]/g, '_');
      link.setAttribute('href', url);
      link.setAttribute('download', `inbound_${code}_id${this.selectedLog.id}.json`);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    },
    domainBadgeClass(domain) {
      const d = String(domain || '').toUpperCase();
      if (d === 'SPARING') return 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30';
      if (d === 'ONLIMO') return 'bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/30';
      if (d === 'WQMS') return 'bg-teal-500/10 text-teal-700 dark:text-teal-300 border-teal-500/30';
      return 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border-cyan-500/30';
    },
    formatNum(n) {
      return Number(n || 0).toLocaleString('id-ID');
    },
    formatBytes(b) {
      const n = Number(b || 0);
      if (n < 1024) return `${n} B`;
      if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`;
      return `${(n / (1024 * 1024)).toFixed(2)} MB`;
    },
    formatClock(ts) {
      if (!ts) return '-';
      const d = new Date(ts);
      if (Number.isNaN(d.getTime())) return '-';
      return d.toLocaleTimeString('en-GB', { hour12: false });
    },
    formatShortDate(ts) {
      if (!ts) return '';
      const d = new Date(ts);
      if (Number.isNaN(d.getTime())) return '';
      return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short' });
    },
    formatFullTime(ts) {
      if (!ts) return '-';
      const d = ts instanceof Date ? ts : new Date(ts);
      if (Number.isNaN(d.getTime())) return String(ts);
      const pad = (v) => String(v).padStart(2, '0');
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
    },
    formatRelative(ts) {
      if (!ts) return '';
      const diffSec = Math.max(0, Math.floor((Date.now() - new Date(ts).getTime()) / 1000));
      if (diffSec < 60) return `${diffSec}s`;
      if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m`;
      if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}j`;
      return `${Math.floor(diffSec / 86400)}h`;
    },
  },
};
</script>
