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
        <!-- Primary View Tabs: Stream Raw JSON vs Kevalidan Data -->
        <div class="flex flex-wrap items-center justify-between gap-2 pb-2 border-b border-slate-200/70 dark:border-slate-800/80">
          <div class="inline-flex items-center p-0.5 rounded-xl bg-slate-200/75 dark:bg-[#0E1626] border border-slate-200/90 dark:border-slate-800 gap-0.5">
            <button
              id="tab-inbound-stream"
              type="button"
              @click="activeTab = 'stream'"
              class="px-3 py-1.5 rounded-lg text-2xs font-bold transition cursor-pointer flex items-center gap-1.5"
              :class="activeTab === 'stream'
                ? 'bg-cyan-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white'"
            >
              <span>📄 Stream Raw JSON</span>
              <span
                class="px-1.5 py-0.2 rounded-full text-4xs font-mono"
                :class="activeTab === 'stream' ? 'bg-white/20 text-white badge-text font-semibold' : 'bg-slate-300/70 dark:bg-slate-700 text-slate-700 dark:text-slate-200'"
              >{{ items.length }}</span>
            </button>
            <button
              id="tab-inbound-validity"
              type="button"
              @click="activeTab = 'validity'"
              class="px-3 py-1.5 rounded-lg text-2xs font-bold transition cursor-pointer flex items-center gap-1.5"
              :class="activeTab === 'validity'
                ? 'bg-emerald-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white'"
            >
              <span>🛡️ Kevalidan Data</span>
              <span
                v-if="validityStats.invalid > 0"
                class="px-1.5 py-0.2 rounded-full text-4xs font-mono font-extrabold bg-rose-500 text-white badge-text"
                title="Jumlah payload tidak valid pada halaman ini"
              >{{ validityStats.invalid }}</span>
              <span
                v-else-if="validityStats.warning > 0"
                class="px-1.5 py-0.2 rounded-full text-4xs font-mono font-extrabold bg-amber-500 text-white badge-text"
                title="Jumlah payload dengan peringatan pada halaman ini"
              >{{ validityStats.warning }}</span>
              <span
                v-else
                class="px-1.5 py-0.2 rounded-full text-4xs font-mono"
                :class="activeTab === 'validity' ? 'bg-white/20 text-white badge-text font-semibold' : 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300'"
              >{{ validityStats.avgScore }}%</span>
            </button>
          </div>

          <div v-if="activeTab === 'validity'" class="flex flex-wrap items-center gap-1.5 text-3xs font-mono">
            <span class="text-slate-500 dark:text-slate-400">Audit Format · Nilai · Kelengkapan (incl. SPARING 00–58)</span>
          </div>
        </div>

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

    <!-- =================================================================== -->
    <!-- TAB 1: SPLIT-PANE STREAM RAW JSON + STICKY INSPECTOR                -->
    <!-- =================================================================== -->
    <div v-if="activeTab === 'stream'" class="grid grid-cols-1 xl:grid-cols-12 gap-3 items-start">
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
                  <div v-if="auditById[item.id]" class="mt-0.5 flex justify-end">
                    <span
                      class="inline-flex items-center gap-0.5 px-1.5 py-0.2 rounded text-4xs font-bold border"
                      :class="validityBadgeClass(auditById[item.id].overallStatus)"
                      :title="'Skor validitas: ' + auditById[item.id].score + '% · ' + auditById[item.id].completeness.label"
                    >
                      <span>{{ validityShortIcon(auditById[item.id].overallStatus) }}</span>
                      <span>{{ auditById[item.id].isSparing ? auditById[item.id].completeness.label : (auditById[item.id].score + '%') }}</span>
                    </span>
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
                  <button
                    v-if="selectedAudit"
                    type="button"
                    @click="activeTab = 'validity'"
                    class="px-2 py-0.5 rounded-full text-4xs font-mono font-bold border transition cursor-pointer"
                    :class="validityBadgeClass(selectedAudit.overallStatus)"
                    title="Buka tab Kevalidan Data untuk detail audit format, nilai & kelengkapan"
                  >
                    {{ validityShortIcon(selectedAudit.overallStatus) }} Validitas {{ selectedAudit.score }}% →
                  </button>
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

    <!-- =================================================================== -->
    <!-- TAB 2: KEVALIDAN DATA (FORMAT, NILAI & KELENGKAPAN + SPARING 00-58) -->
    <!-- =================================================================== -->
    <div v-else id="panel-inbound-validity" class="space-y-3">
      <!-- Clickable Validity KPI & Filter Strip -->
      <div
        id="validity-kpi-strip"
        class="bg-white dark:bg-[#111827] border border-slate-200/90 dark:border-slate-800 rounded-2xl p-2.5 shadow-sm flex flex-wrap items-center justify-between gap-2"
      >
        <div class="flex flex-wrap items-center gap-1.5">
          <button
            type="button"
            @click="validityFilter = 'ALL'"
            class="px-2.5 py-1.5 rounded-xl border text-2xs font-bold transition cursor-pointer flex items-center gap-1.5"
            :class="validityFilter === 'ALL'
              ? 'bg-cyan-600 text-white btn-text border-cyan-600 shadow-2xs'
              : 'bg-slate-50 dark:bg-[#0B0F19] border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-200 hover:border-cyan-400'"
          >
            <span>Semua</span>
            <span class="font-mono font-extrabold">{{ validityStats.total }}</span>
            <span
              class="px-1.5 py-0.2 rounded-full text-4xs font-mono"
              :class="validityFilter === 'ALL' ? 'bg-white/20 text-white badge-text font-semibold' : 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-300'"
            >Avg {{ validityStats.avgScore }}%</span>
          </button>

          <button
            type="button"
            @click="validityFilter = 'VALID'"
            class="px-2.5 py-1.5 rounded-xl border text-2xs font-bold transition cursor-pointer flex items-center gap-1.5"
            :class="validityFilter === 'VALID'
              ? 'bg-emerald-600 text-white btn-text border-emerald-600 shadow-2xs'
              : 'bg-emerald-50/70 dark:bg-emerald-950/25 border-emerald-200/80 dark:border-emerald-800/60 text-emerald-700 dark:text-emerald-300 hover:border-emerald-400'"
          >
            <span>✓ Valid</span>
            <span class="font-mono font-extrabold">{{ validityStats.valid }}</span>
          </button>

          <button
            type="button"
            @click="validityFilter = 'WARNING'"
            class="px-2.5 py-1.5 rounded-xl border text-2xs font-bold transition cursor-pointer flex items-center gap-1.5"
            :class="validityFilter === 'WARNING'
              ? 'bg-amber-600 text-white btn-text border-amber-600 shadow-2xs'
              : 'bg-amber-50/70 dark:bg-amber-950/25 border-amber-200/80 dark:border-amber-800/60 text-amber-700 dark:text-amber-300 hover:border-amber-400'"
          >
            <span>⚠ Peringatan</span>
            <span class="font-mono font-extrabold">{{ validityStats.warning }}</span>
          </button>

          <button
            type="button"
            @click="validityFilter = 'INVALID'"
            class="px-2.5 py-1.5 rounded-xl border text-2xs font-bold transition cursor-pointer flex items-center gap-1.5"
            :class="validityFilter === 'INVALID'
              ? 'bg-rose-600 text-white btn-text border-rose-600 shadow-2xs'
              : 'bg-rose-50/70 dark:bg-rose-950/25 border-rose-200/80 dark:border-rose-800/60 text-rose-700 dark:text-rose-300 hover:border-rose-400'"
          >
            <span>✕ Tidak Valid</span>
            <span class="font-mono font-extrabold">{{ validityStats.invalid }}</span>
          </button>
        </div>

        <button
          id="btn-validity-sparing-only"
          type="button"
          @click="validitySparingOnly = !validitySparingOnly"
          class="px-2.5 py-1.5 rounded-xl border text-2xs font-bold transition cursor-pointer flex items-center gap-1.5"
          :class="validitySparingOnly
            ? 'bg-amber-600 text-white btn-text border-amber-600 shadow-2xs'
            : 'bg-slate-50 dark:bg-[#0B0F19] border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:border-amber-400'"
          title="Filter khusus payload SPARING (audit kelengkapan 30 slot menit 00–58)"
        >
          <span>⏱ Khusus SPARING (00–58)</span>
          <span
            class="px-1.5 py-0.2 rounded-full text-4xs font-mono"
            :class="validitySparingOnly ? 'bg-white/20 text-white badge-text font-semibold' : 'bg-amber-500/15 text-amber-700 dark:text-amber-300'"
          >{{ validityStats.sparingComplete }}/{{ validityStats.sparingTotal }}</span>
        </button>
      </div>

      <!-- Split-Pane Validity Audit Table + Sticky Detailed Inspector -->
      <div class="grid grid-cols-1 xl:grid-cols-12 gap-3 items-start">
        <!-- LEFT: Compact Audit Matrix Table -->
        <div class="xl:col-span-7 bg-white dark:bg-[#111827] border border-slate-200/90 dark:border-slate-800 rounded-2xl shadow-sm overflow-hidden">
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs">
              <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-4xs font-bold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
                <tr>
                  <th class="px-3 py-2.5">Waktu &amp; Tenant</th>
                  <th class="px-2 py-2.5">Format</th>
                  <th class="px-2 py-2.5">Nilai</th>
                  <th class="px-2 py-2.5">Kelengkapan</th>
                  <th class="px-3 py-2.5 text-right">Validitas</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
                <tr
                  v-for="row in filteredAuditedItems"
                  :key="'val-' + row.item.id"
                  @click="selectValidityRow(row.item)"
                  class="transition-colors cursor-pointer"
                  :class="selectedLog && Number(selectedLog.id) === Number(row.item.id)
                    ? 'bg-emerald-50/60 dark:bg-emerald-950/25'
                    : 'hover:bg-slate-50/80 dark:hover:bg-slate-800/40'"
                >
                  <!-- Waktu & Tenant -->
                  <td class="px-3 py-2 max-w-[220px]">
                    <div class="flex items-center gap-1.5">
                      <span class="font-mono text-2xs font-bold text-slate-800 dark:text-slate-100">{{ formatClock(row.item.received_at) }}</span>
                      <span
                        v-if="row.audit.isSparing"
                        class="px-1.5 py-0.2 rounded-full text-4xs font-mono font-bold bg-amber-500/12 text-amber-700 dark:text-amber-300 border border-amber-500/30"
                      >SPARING</span>
                      <span
                        v-else-if="row.item.logger_type"
                        class="px-1.5 py-0.2 rounded-full text-4xs font-mono font-bold border"
                        :class="domainBadgeClass(row.item.logger_type)"
                      >{{ row.item.logger_type }}</span>
                    </div>
                    <div class="font-bold text-2xs text-slate-900 dark:text-white truncate mt-0.5">
                      {{ row.item.tenant_name || (row.item.tenant_id ? ('Tenant #' + row.item.tenant_id) : 'Unauthenticated') }}
                    </div>
                    <div class="text-4xs font-mono text-slate-400 truncate">
                      {{ row.item.site_code || row.item.device_code || row.item.endpoint }}
                    </div>
                  </td>

                  <!-- Format Pillar -->
                  <td class="px-2 py-2 whitespace-nowrap">
                    <span
                      class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-4xs font-mono font-bold border"
                      :class="pillarBadgeClass(row.audit.format.status)"
                    >
                      <span>{{ pillarIcon(row.audit.format.status) }}</span>
                      <span>{{ row.audit.format.label }}</span>
                    </span>
                    <div class="text-4xs font-mono text-slate-400 truncate max-w-[110px] mt-0.5" :title="row.audit.schemaLabel">
                      {{ row.item.detected_format || row.audit.schemaLabel }}
                    </div>
                  </td>

                  <!-- Nilai Pillar -->
                  <td class="px-2 py-2 whitespace-nowrap">
                    <span
                      class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-4xs font-mono font-bold border"
                      :class="pillarBadgeClass(row.audit.values.status)"
                    >
                      <span>{{ pillarIcon(row.audit.values.status) }}</span>
                      <span>{{ row.audit.values.label }}</span>
                    </span>
                    <div class="text-4xs font-mono text-slate-400 mt-0.5">
                      {{ row.audit.values.parameterStats.length }} kanal
                    </div>
                  </td>

                  <!-- Kelengkapan Pillar (highlights SPARING 00-58 explicitly) -->
                  <td class="px-2 py-2 whitespace-nowrap">
                    <span
                      class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md text-4xs font-mono font-bold border"
                      :class="pillarBadgeClass(row.audit.completeness.status)"
                    >
                      <span>{{ pillarIcon(row.audit.completeness.status) }}</span>
                      <span>{{ row.audit.completeness.label }}</span>
                    </span>
                    <div
                      v-if="row.audit.isSparing && row.audit.completeness.sparing"
                      class="text-4xs font-mono mt-0.5"
                      :class="row.audit.completeness.sparing.missingMinutes.length > 0 ? 'text-rose-600 dark:text-rose-400 font-semibold' : 'text-slate-400'"
                    >
                      <template v-if="row.audit.completeness.sparing.missingMinutes.length > 0">
                        {{ row.audit.completeness.sparing.missingMinutes.length }} menit hilang
                      </template>
                      <template v-else-if="row.audit.completeness.sparing.partialSlots > 0">
                        {{ row.audit.completeness.sparing.partialSlots }} slot parsial/null
                      </template>
                      <template v-else>
                        00–58 lengkap penuh
                      </template>
                    </div>
                    <div v-else class="text-4xs font-mono text-slate-400 mt-0.5">
                      {{ row.audit.completeness.recordCount }} rec
                    </div>
                  </td>

                  <!-- Overall Verdict + Score -->
                  <td class="px-3 py-2 text-right whitespace-nowrap font-mono">
                    <span
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-4xs font-bold border"
                      :class="validityBadgeClass(row.audit.overallStatus)"
                    >
                      <span>{{ validityShortIcon(row.audit.overallStatus) }}</span>
                      <span>{{ validityLabel(row.audit.overallStatus) }}</span>
                    </span>
                    <div class="text-4xs font-bold text-slate-500 dark:text-slate-400 mt-0.5 tabular-nums">
                      Skor {{ row.audit.score }}%
                    </div>
                  </td>
                </tr>

                <tr v-if="!loading && filteredAuditedItems.length === 0">
                  <td colspan="5" class="px-4 py-12 text-center text-slate-500 dark:text-slate-400">
                    <div class="text-2xl mb-1">🛡️</div>
                    <div class="font-bold text-xs text-slate-700 dark:text-slate-300">Tidak ada data inbound yang cocok pada filter validitas ini</div>
                    <p class="text-2xs text-slate-400 mt-0.5">
                      Ubah filter status validitas atau rentang waktu di atas untuk melihat hasil audit lainnya.
                    </p>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Pagination Bar (shared with Tab 1) -->
          <div class="px-3.5 py-2.5 border-t border-slate-100 dark:border-slate-800/80 bg-slate-50/60 dark:bg-[#0B0F19]/50 flex flex-wrap items-center justify-between gap-2 text-2xs">
            <div class="text-slate-500 dark:text-slate-400 font-mono">
              Audit {{ filteredAuditedItems.length }} dari {{ items.length }} log halaman {{ page }}/{{ totalPages }}
            </div>
            <div class="flex items-center gap-1.5">
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

        <!-- RIGHT: Sticky Validity Audit Inspector (Desktop xl+) -->
        <div class="hidden xl:flex xl:col-span-5 bg-white dark:bg-[#111827] border border-slate-200/90 dark:border-slate-800 rounded-2xl shadow-sm overflow-hidden flex-col sticky top-3">
          <template v-if="selectedLog && selectedAudit">
            <!-- Verdict Header -->
            <div class="px-4 py-3 border-b border-slate-100 dark:border-slate-800 bg-slate-50/70 dark:bg-[#0B0F19]/60 space-y-2.5">
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <span
                      class="px-2 py-0.5 rounded-full text-4xs font-mono font-bold border"
                      :class="validityBadgeClass(selectedAudit.overallStatus)"
                    >
                      {{ validityShortIcon(selectedAudit.overallStatus) }} {{ validityLabel(selectedAudit.overallStatus) }} · {{ selectedAudit.score }}%
                    </span>
                    <span class="text-xs font-extrabold text-slate-900 dark:text-white truncate">
                      {{ selectedLog.tenant_name || (selectedLog.tenant_id ? ('Tenant #' + selectedLog.tenant_id) : 'Unauthenticated') }}
                    </span>
                  </div>
                  <div class="text-3xs font-mono text-slate-500 dark:text-slate-400 mt-1">
                    {{ formatFullTime(selectedLog.received_at) }} · {{ selectedAudit.schemaLabel }}
                  </div>
                </div>

                <button
                  type="button"
                  @click="activeTab = 'stream'"
                  class="px-2.5 py-1 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white btn-text text-3xs font-bold shadow-2xs transition cursor-pointer flex-shrink-0"
                >
                  Lihat Raw JSON →
                </button>
              </div>

              <!-- 3-Pillar Status Summary Grid -->
              <div class="grid grid-cols-3 gap-1.5">
                <div class="p-2 rounded-xl border" :class="pillarCardClass(selectedAudit.format.status)">
                  <div class="text-4xs font-bold uppercase tracking-wider opacity-75">1. Format</div>
                  <div class="text-xs font-extrabold mt-0.5">{{ pillarIcon(selectedAudit.format.status) }} {{ selectedAudit.format.label }}</div>
                  <div class="text-4xs font-mono truncate opacity-80 mt-0.5" :title="selectedAudit.format.summary">{{ selectedAudit.format.summary }}</div>
                </div>
                <div class="p-2 rounded-xl border" :class="pillarCardClass(selectedAudit.values.status)">
                  <div class="text-4xs font-bold uppercase tracking-wider opacity-75">2. Nilai</div>
                  <div class="text-xs font-extrabold mt-0.5">{{ pillarIcon(selectedAudit.values.status) }} {{ selectedAudit.values.label }}</div>
                  <div class="text-4xs font-mono truncate opacity-80 mt-0.5" :title="selectedAudit.values.summary">{{ selectedAudit.values.summary }}</div>
                </div>
                <div class="p-2 rounded-xl border" :class="pillarCardClass(selectedAudit.completeness.status)">
                  <div class="text-4xs font-bold uppercase tracking-wider opacity-75">3. Kelengkapan</div>
                  <div class="text-xs font-extrabold mt-0.5">{{ pillarIcon(selectedAudit.completeness.status) }} {{ selectedAudit.completeness.label }}</div>
                  <div class="text-4xs font-mono truncate opacity-80 mt-0.5" :title="selectedAudit.completeness.summary">{{ selectedAudit.completeness.summary }}</div>
                </div>
              </div>
            </div>

            <!-- Scrollable Audit Body -->
            <div class="p-3.5 space-y-3 overflow-y-auto max-h-[590px]">
              <!-- SPARING 30-Slot Minute Matrix (00 - 58) -->
              <div
                v-if="selectedAudit.isSparing && selectedAudit.completeness.sparing"
                id="sparing-minute-matrix"
                class="p-3 rounded-xl bg-slate-50/90 dark:bg-[#0B0F19] border border-slate-200/90 dark:border-slate-800 space-y-2"
              >
                <div class="flex items-center justify-between gap-2 flex-wrap">
                  <div>
                    <div class="text-2xs font-extrabold text-slate-900 dark:text-white flex items-center gap-1.5">
                      <span>⏱ Matrix Menit SPARING (00–58)</span>
                    </div>
                    <div class="text-4xs font-mono text-slate-500 dark:text-slate-400">
                      {{ selectedAudit.completeness.sparing.windowLabel }}
                    </div>
                  </div>
                  <span
                    class="px-2 py-0.5 rounded-full text-4xs font-mono font-bold border"
                    :class="pillarBadgeClass(selectedAudit.completeness.status)"
                  >
                    {{ selectedAudit.completeness.sparing.completeSlots }}/30 Penuh · {{ selectedAudit.completeness.sparing.missingMinutes.length }} Hilang
                  </span>
                </div>

                <!-- 10x3 Minute Slot Grid (:00 .. :58) -->
                <div class="grid grid-cols-10 gap-1">
                  <button
                    v-for="slot in selectedAudit.completeness.sparing.slots"
                    :key="'slot-' + slot.label"
                    type="button"
                    @click="selectedSparingMinute = slot.minute"
                    class="py-1 rounded-md text-4xs font-mono font-bold border text-center transition cursor-pointer"
                    :class="[
                      sparingSlotClass(slot),
                      selectedSparingMinute === slot.minute ? 'ring-2 ring-cyan-500 ring-offset-1 dark:ring-offset-slate-900' : ''
                    ]"
                    :title="sparingSlotTooltip(slot)"
                  >
                    :{{ slot.label }}
                  </button>
                </div>

                <!-- Legend -->
                <div class="flex flex-wrap items-center justify-between gap-2 text-4xs text-slate-500 dark:text-slate-400 pt-0.5">
                  <div class="flex items-center gap-2.5">
                    <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-xs bg-emerald-500"></span>Lengkap ({{ selectedAudit.completeness.sparing.completeSlots }})</span>
                    <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-xs bg-amber-500"></span>Null/Parsial ({{ selectedAudit.completeness.sparing.partialSlots }})</span>
                    <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-xs bg-rose-500"></span>Hilang ({{ selectedAudit.completeness.sparing.missingMinutes.length }})</span>
                  </div>
                  <span class="font-mono">Klik slot untuk detail</span>
                </div>

                <!-- Selected Minute Slot Detail Box -->
                <div
                  v-if="activeSparingSlot"
                  class="p-2 rounded-lg bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-700/80 text-3xs font-mono space-y-1"
                >
                  <div class="flex items-center justify-between gap-2">
                    <span class="font-bold text-slate-800 dark:text-slate-100">
                      Slot :{{ activeSparingSlot.label }}
                      <template v-if="activeSparingSlot.timestampText"> · {{ activeSparingSlot.timestampText }}</template>
                    </span>
                    <span
                      class="px-1.5 py-0.2 rounded text-4xs font-bold border"
                      :class="activeSparingSlot.state === 'COMPLETE'
                        ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30'
                        : activeSparingSlot.state === 'PARTIAL'
                          ? 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30'
                          : 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30'"
                    >
                      {{ activeSparingSlot.state === 'COMPLETE' ? 'LENGKAP' : activeSparingSlot.state === 'PARTIAL' ? 'PARSIAL / NULL' : 'TIDAK TERKIRIM' }}
                    </span>
                  </div>
                  <div v-if="activeSparingSlot.state === 'MISSING'" class="text-rose-600 dark:text-rose-400">
                    Slot menit :{{ activeSparingSlot.label }} tidak ditemukan pada array payload SPARING.
                  </div>
                  <div v-else class="flex flex-wrap gap-1.5 pt-0.5">
                    <span
                      v-for="(val, k) in activeSparingSlot.values"
                      :key="'sv-' + k"
                      class="px-1.5 py-0.5 rounded border text-4xs"
                      :class="val === null
                        ? 'bg-amber-50 dark:bg-amber-950/30 border-amber-300 dark:border-amber-700 text-amber-700 dark:text-amber-300 font-bold'
                        : activeSparingSlot.invalidKeys.includes(k)
                          ? 'bg-rose-50 dark:bg-rose-950/30 border-rose-300 dark:border-rose-700 text-rose-700 dark:text-rose-300 font-bold'
                          : 'bg-slate-50 dark:bg-slate-800/80 border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200'"
                    >
                      {{ k }}: <strong>{{ val === null ? 'null' : val }}</strong>
                    </span>
                  </div>
                </div>
              </div>

              <!-- Checklist Temuan (Format · Nilai · Kelengkapan) -->
              <div class="space-y-1.5">
                <div class="text-3xs font-extrabold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Hasil Evaluasi Format, Nilai &amp; Kelengkapan
                </div>
                <div class="space-y-1">
                  <div
                    v-for="(chk, cIdx) in combinedAuditChecks"
                    :key="'chk-' + cIdx"
                    class="px-2.5 py-1.5 rounded-xl border flex items-start gap-2 text-3xs"
                    :class="chk.level === 'FAIL'
                      ? 'bg-rose-50/80 dark:bg-rose-950/25 border-rose-200 dark:border-rose-800/60 text-rose-700 dark:text-rose-300'
                      : chk.level === 'WARN'
                        ? 'bg-amber-50/80 dark:bg-amber-950/25 border-amber-200 dark:border-amber-800/60 text-amber-700 dark:text-amber-300'
                        : 'bg-slate-50/70 dark:bg-[#0B0F19] border-slate-200/70 dark:border-slate-800 text-slate-600 dark:text-slate-300'"
                  >
                    <span class="font-bold flex-shrink-0 mt-0.2">{{ pillarIcon(chk.level) }}</span>
                    <div class="min-w-0 flex-1">
                      <div class="font-bold flex items-center gap-1.5 flex-wrap">
                        <span class="px-1 py-0.2 rounded text-4xs font-mono bg-white/70 dark:bg-slate-800/80 border border-slate-200/70 dark:border-slate-700">{{ chk.pillar }}</span>
                        <span>{{ chk.label }}</span>
                      </div>
                      <div class="text-4xs opacity-90 mt-0.5 break-words">{{ chk.detail }}</div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Compact Parameter Values Inspection Table -->
              <div v-if="selectedAudit.values.parameterStats.length > 0" class="space-y-1.5">
                <div class="text-3xs font-extrabold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  Inspeksi Nilai Parameter ({{ selectedAudit.values.parameterStats.length }} Kanal)
                </div>
                <div class="rounded-xl border border-slate-200 dark:border-slate-800 overflow-hidden">
                  <table class="w-full text-left text-3xs font-mono">
                    <thead class="bg-slate-50 dark:bg-[#0B0F19] text-4xs uppercase text-slate-500 border-b border-slate-200 dark:border-slate-800">
                      <tr>
                        <th class="px-2.5 py-1.5">Parameter</th>
                        <th class="px-2 py-1.5 text-right">Nilai / Rentang</th>
                        <th class="px-2 py-1.5 text-center">Null</th>
                        <th class="px-2.5 py-1.5 text-right">Status</th>
                      </tr>
                    </thead>
                    <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
                      <tr v-for="ps in selectedAudit.values.parameterStats" :key="'ps-' + ps.code">
                        <td class="px-2.5 py-1.5 font-bold text-slate-800 dark:text-slate-100">
                          {{ ps.code }}
                          <span class="text-4xs font-normal text-slate-400">({{ ps.count }}×)</span>
                        </td>
                        <td class="px-2 py-1.5 text-right text-slate-700 dark:text-slate-200 tabular-nums">
                          <template v-if="ps.min === null">—</template>
                          <template v-else-if="ps.count === 1 || ps.min === ps.max">{{ ps.lastValue }}</template>
                          <template v-else>{{ ps.min }} – {{ ps.max }}</template>
                        </td>
                        <td class="px-2 py-1.5 text-center tabular-nums" :class="ps.nullCount > 0 ? 'text-amber-600 dark:text-amber-400 font-bold' : 'text-slate-400'">
                          {{ ps.nullCount }}
                        </td>
                        <td class="px-2.5 py-1.5 text-right">
                          <span
                            class="px-1.5 py-0.2 rounded text-4xs font-bold border"
                            :class="validityBadgeClass(ps.status)"
                            :title="ps.issueText || 'Nilai valid'"
                          >
                            {{ ps.status === 'VALID' ? 'OK' : ps.status === 'WARNING' ? 'NULL' : 'ANOMALI' }}
                          </span>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </template>

          <div v-else class="p-10 text-center text-slate-400">
            <div class="text-3xl mb-2">🛡️</div>
            <div class="text-xs font-bold text-slate-600 dark:text-slate-300">Pilih Baris untuk Inspeksi Validitas</div>
            <p class="text-2xs mt-1">Klik salah satu log inbound di kiri untuk melihat detail format, nilai, dan matrix menit SPARING 00–58.</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Mobile / Tablet Validity Audit Modal (< xl) -->
    <div
      v-if="mobileValidityModalOpen && selectedLog && selectedAudit"
      @click.self="mobileValidityModalOpen = false"
      class="xl:hidden fixed inset-0 z-[75] flex items-center justify-center p-3 bg-slate-950/70 backdrop-blur-xs"
    >
      <div class="bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl w-full max-w-2xl max-h-[88vh] flex flex-col overflow-hidden">
        <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between gap-2">
          <div class="min-w-0">
            <div class="flex items-center gap-1.5 flex-wrap">
              <span class="px-2 py-0.5 rounded-full text-4xs font-mono font-bold border" :class="validityBadgeClass(selectedAudit.overallStatus)">
                {{ validityLabel(selectedAudit.overallStatus) }} · {{ selectedAudit.score }}%
              </span>
              <span class="font-extrabold text-xs text-slate-900 dark:text-white truncate">
                {{ selectedLog.tenant_name || ('Tenant #' + selectedLog.tenant_id) }}
              </span>
            </div>
            <div class="text-4xs font-mono text-slate-400 mt-0.5">
              {{ formatFullTime(selectedLog.received_at) }} · {{ selectedAudit.schemaLabel }}
            </div>
          </div>
          <div class="flex items-center gap-1.5">
            <button
              type="button"
              @click="mobileValidityModalOpen = false; activeTab = 'stream'; mobileInspectorOpen = true"
              class="px-2.5 py-1 rounded-lg bg-cyan-600 text-white btn-text text-3xs font-bold cursor-pointer"
            >
              Raw JSON
            </button>
            <button
              type="button"
              @click="mobileValidityModalOpen = false"
              class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 text-xs font-bold text-slate-600 dark:text-slate-300 cursor-pointer"
            >
              ✕
            </button>
          </div>
        </div>
        <div class="p-3.5 space-y-3 overflow-y-auto flex-1">
          <div v-if="selectedAudit.isSparing && selectedAudit.completeness.sparing" class="p-2.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 space-y-2">
            <div class="flex items-center justify-between text-2xs font-bold">
              <span>⏱ Menit SPARING (00–58)</span>
              <span class="font-mono">{{ selectedAudit.completeness.sparing.completeSlots }}/30 Penuh</span>
            </div>
            <div class="grid grid-cols-6 sm:grid-cols-10 gap-1">
              <span
                v-for="slot in selectedAudit.completeness.sparing.slots"
                :key="'mslot-' + slot.label"
                class="py-1 rounded-md text-4xs font-mono font-bold border text-center"
                :class="sparingSlotClass(slot)"
              >:{{ slot.label }}</span>
            </div>
          </div>

          <div class="space-y-1">
            <div
              v-for="(chk, idx) in combinedAuditChecks"
              :key="'mchk-' + idx"
              class="px-2.5 py-1.5 rounded-xl border text-3xs flex items-start gap-2"
              :class="chk.level === 'FAIL'
                ? 'bg-rose-50 dark:bg-rose-950/25 border-rose-200 dark:border-rose-800/60 text-rose-700 dark:text-rose-300'
                : chk.level === 'WARN'
                  ? 'bg-amber-50 dark:bg-amber-950/25 border-amber-200 dark:border-amber-800/60 text-amber-700 dark:text-amber-300'
                  : 'bg-slate-50 dark:bg-[#0B0F19] border-slate-200/70 dark:border-slate-800 text-slate-600 dark:text-slate-300'"
            >
              <span class="font-bold">{{ pillarIcon(chk.level) }}</span>
              <div>
                <div class="font-bold">{{ chk.pillar }} · {{ chk.label }}</div>
                <div class="text-4xs opacity-90">{{ chk.detail }}</div>
              </div>
            </div>
          </div>
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
import { validateInboundLog } from '../utils/inboundValidator';

export default {
  name: 'InboundDataView',
  data() {
    return {
      activeTab: 'stream',
      validityFilter: 'ALL',
      validitySparingOnly: false,
      selectedSparingMinute: null,
      mobileValidityModalOpen: false,

      loading: false,
      autoRefresh: true,
      refreshTimer: null,
      searchTimer: null,
      showPurgeMenu: false,
      mobileInspectorOpen: false,
      jsonViewMode: 'pretty',
      toastMessage: '',
      toastType: 'success',

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
        status: '',
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
    activeTenantId() {
      return Number((this.$store && this.$store.state && this.$store.state.activeTenantId) || 0);
    },
    totalPages() {
      return Math.max(1, Math.ceil((this.total || 0) / (this.pageSize || 25)));
    },
    hasActiveFilters() {
      return Boolean(
        this.filters.search ||
        this.filters.status ||
        this.filters.format ||
        this.filters.source ||
        this.activeTimePreset !== '24h'
      );
    },
    auditedItems() {
      return (this.items || []).map(item => ({
        item,
        audit: validateInboundLog(item),
      }));
    },
    auditById() {
      const map = {};
      for (const row of this.auditedItems) {
        map[row.item.id] = row.audit;
      }
      return map;
    },
    validityStats() {
      const rows = this.auditedItems;
      let valid = 0;
      let warning = 0;
      let invalid = 0;
      let scoreSum = 0;
      let sparingTotal = 0;
      let sparingComplete = 0;
      for (const r of rows) {
        scoreSum += r.audit.score;
        if (r.audit.overallStatus === 'VALID') valid += 1;
        else if (r.audit.overallStatus === 'WARNING') warning += 1;
        else invalid += 1;
        if (r.audit.isSparing) {
          sparingTotal += 1;
          const sp = r.audit.completeness && r.audit.completeness.sparing;
          if (sp && sp.completeSlots === 30 && sp.missingMinutes.length === 0) {
            sparingComplete += 1;
          }
        }
      }
      return {
        total: rows.length,
        valid,
        warning,
        invalid,
        avgScore: rows.length > 0 ? Math.round(scoreSum / rows.length) : 100,
        sparingTotal,
        sparingComplete,
      };
    },
    filteredAuditedItems() {
      return this.auditedItems.filter(row => {
        if (this.validitySparingOnly && !row.audit.isSparing) return false;
        if (this.validityFilter !== 'ALL' && row.audit.overallStatus !== this.validityFilter) return false;
        return true;
      });
    },
    selectedAudit() {
      if (!this.selectedLog) return null;
      return this.auditById[this.selectedLog.id] || validateInboundLog(this.selectedLog);
    },
    activeSparingSlot() {
      const sp = this.selectedAudit && this.selectedAudit.completeness && this.selectedAudit.completeness.sparing;
      if (!sp || !Array.isArray(sp.slots)) return null;
      if (this.selectedSparingMinute !== null) {
        const hit = sp.slots.find(s => s.minute === this.selectedSparingMinute);
        if (hit) return hit;
      }
      // Auto-focus first missing or partial slot for immediate troubleshooting, else slot :00
      return sp.slots.find(s => s.state === 'MISSING') || sp.slots.find(s => s.state === 'PARTIAL') || sp.slots[0] || null;
    },
    combinedAuditChecks() {
      const a = this.selectedAudit;
      if (!a) return [];
      const list = [
        ...(a.format.checks || []).map(c => ({ ...c, pillar: 'FORMAT' })),
        ...(a.values.checks || []).map(c => ({ ...c, pillar: 'NILAI' })),
        ...(a.completeness.checks || []).map(c => ({ ...c, pillar: 'KELENGKAPAN' })),
      ];
      const rank = { FAIL: 0, WARN: 1, PASS: 2 };
      return list.sort((x, y) => (rank[x.level] ?? 9) - (rank[y.level] ?? 9));
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
        if (this.activeTenantId > 0) params.tenant_id = this.activeTenantId;
        if (this.filters.status) params.status = this.filters.status;
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
        status: '',
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
      this.selectedSparingMinute = null;
      if (typeof window !== 'undefined' && window.innerWidth < 1280) {
        this.mobileInspectorOpen = true;
      }
    },
    selectValidityRow(item) {
      this.selectedLog = item;
      this.selectedSparingMinute = null;
      if (typeof window !== 'undefined' && window.innerWidth < 1280) {
        this.mobileValidityModalOpen = true;
      }
    },
    validityLabel(status) {
      if (status === 'VALID') return 'VALID';
      if (status === 'WARNING') return 'PERINGATAN';
      return 'TIDAK VALID';
    },
    validityShortIcon(status) {
      if (status === 'VALID') return '✓';
      if (status === 'WARNING') return '⚠';
      return '✕';
    },
    validityBadgeClass(status) {
      if (status === 'VALID') {
        return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30';
      }
      if (status === 'WARNING') {
        return 'bg-amber-500/12 text-amber-700 dark:text-amber-300 border-amber-500/35';
      }
      return 'bg-rose-500/12 text-rose-600 dark:text-rose-400 border-rose-500/35';
    },
    pillarIcon(status) {
      if (status === 'PASS') return '✓';
      if (status === 'WARN') return '⚠';
      return '✕';
    },
    pillarBadgeClass(status) {
      if (status === 'PASS') {
        return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30';
      }
      if (status === 'WARN') {
        return 'bg-amber-500/12 text-amber-700 dark:text-amber-300 border-amber-500/35';
      }
      return 'bg-rose-500/12 text-rose-600 dark:text-rose-400 border-rose-500/35';
    },
    pillarCardClass(status) {
      if (status === 'PASS') {
        return 'bg-emerald-50/70 dark:bg-emerald-950/25 border-emerald-200/80 dark:border-emerald-800/50 text-emerald-800 dark:text-emerald-200';
      }
      if (status === 'WARN') {
        return 'bg-amber-50/70 dark:bg-amber-950/25 border-amber-200/80 dark:border-amber-800/50 text-amber-800 dark:text-amber-200';
      }
      return 'bg-rose-50/70 dark:bg-rose-950/25 border-rose-200/80 dark:border-rose-800/50 text-rose-800 dark:text-rose-200';
    },
    sparingSlotClass(slot) {
      if (!slot || slot.state === 'MISSING') {
        return 'bg-rose-100 dark:bg-rose-950/45 border-rose-300 dark:border-rose-700 text-rose-700 dark:text-rose-300';
      }
      if (slot.state === 'PARTIAL') {
        return 'bg-amber-100 dark:bg-amber-950/45 border-amber-300 dark:border-amber-700 text-amber-800 dark:text-amber-300';
      }
      return 'bg-emerald-100 dark:bg-emerald-950/40 border-emerald-300 dark:border-emerald-700 text-emerald-800 dark:text-emerald-300';
    },
    sparingSlotTooltip(slot) {
      if (!slot) return '';
      if (slot.state === 'MISSING') return `Menit :${slot.label} — HILANG (tidak dikirim)`;
      if (slot.state === 'PARTIAL') {
        const parts = [];
        if (slot.nullKeys.length) parts.push('null: ' + slot.nullKeys.join(', '));
        if (slot.invalidKeys.length) parts.push('anomali: ' + slot.invalidKeys.join(', '));
        return `Menit :${slot.label} (${slot.timestampText}) — PARSIAL (${parts.join('; ')})`;
      }
      return `Menit :${slot.label} (${slot.timestampText}) — LENGKAP`;
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
