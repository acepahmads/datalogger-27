<template>
  <div class="space-y-4 p-3 sm:p-5 max-w-[1920px] mx-auto min-h-screen text-slate-800 dark:text-slate-100">
    <!-- Breadcrumb & Top Bar -->
    <div class="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-slate-200 dark:border-[#20354B]">
      <div class="flex items-center space-x-3">
        <router-link
          to="/monitoring/operator"
          class="p-1.5 rounded-lg bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-[#20354B] transition"
          title="Back to Operator Command Center"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </router-link>
        <div>
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-400 animate-pulse"></span>
            <h1 class="text-lg sm:text-xl font-bold font-mono tracking-tight text-slate-900 dark:text-white uppercase">
              INDEKS STANDAR PENCEMAR UDARA (ISPU)
            </h1>
            <span class="px-2 py-0.5 rounded text-4xs font-mono font-bold bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30">
              PERMEN LHK NO. 14 TAHUN 2020
            </span>
          </div>
          <p class="text-3xs text-slate-500 dark:text-slate-400 font-mono mt-0.5">
            Authoritative regulatory air quality index calculation, continuous 24h data window verification, and compliance records.
          </p>
        </div>
      </div>

      <!-- Station Selector & Top Controls -->
      <div class="flex flex-wrap items-center gap-2">
        <!-- Station Dropdown -->
        <div class="flex items-center space-x-1.5 bg-slate-100 dark:bg-[#101B2B] px-2.5 py-1 rounded-lg border border-slate-200 dark:border-[#20354B] text-3xs font-mono">
          <span class="text-slate-500">Station:</span>
          <select
            v-model="selectedDeviceId"
            @change="handleDeviceChange"
            class="bg-transparent font-bold text-slate-800 dark:text-white focus:outline-hidden cursor-pointer"
          >
            <option
              v-for="d in devicesList"
              :key="d.id"
              :value="d.id"
              class="bg-white dark:bg-slate-900 text-slate-800 dark:text-white"
            >
              {{ d.device_code }} - {{ d.device_name }}
            </option>
          </select>
        </div>

        <!-- Mode Toggle: 24-Hour (Regulatory) vs Per Agregasi (Real-time Interval) -->
        <div class="flex items-center space-x-1 bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-3xs font-mono">
          <button
            @click="ispuMode = '24H'"
            class="px-2 py-1 rounded-md transition font-bold"
            :class="ispuMode === '24H'
              ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border border-cyan-500/40 shadow-xs'
              : 'text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200'"
            title="Perhitungan ISPU 24 Jam (Rata-rata Harian sesuai regulasi Permen LHK 14/2020)"
          >
            📅 24-Jam (Regulasi)
          </button>
          <button
            @click="ispuMode = 'AGGREGATION'"
            class="px-2 py-1 rounded-md transition font-bold"
            :class="ispuMode === 'AGGREGATION'
              ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border border-cyan-500/40 shadow-xs'
              : 'text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200'"
            title="Perhitungan ISPU Per Agregasi (Interval Rollup Terkini, misal 2 Menit)"
          >
            ⏱️ Per Agregasi (2 Menit)
          </button>
        </div>

        <!-- Parameter Mappings Button -->
        <button
          v-if="canManageMapping"
          @click="openMappingsModal"
          class="px-2.5 py-1.5 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1.5 bg-slate-100 dark:bg-[#101B2B] text-slate-700 dark:text-slate-300 border-slate-300 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-500"
          title="Configure station parameter mappings to official ISPU pollutants"
        >
          <span>⚙️</span>
          <span>Mappings</span>
        </button>

        <!-- Recalculate Button -->
        <button
          @click="triggerRecalculation"
          :disabled="recalculating"
          class="px-2.5 py-1.5 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1.5 bg-slate-100 dark:bg-[#101B2B] text-slate-700 dark:text-slate-300 border-slate-300 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-500"
          title="Trigger on-demand 24h recalculation from aggregation data"
        >
          <svg class="w-3 h-3" :class="{ 'animate-spin': recalculating }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>Recalculate</span>
        </button>

        <!-- Export CSV -->
        <button
          @click="exportData('csv')"
          class="px-2.5 py-1.5 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1 bg-cyan-600 hover:bg-cyan-500 text-white border-cyan-400 shadow-xs"
        >
          <span>⬇️ CSV</span>
        </button>

        <!-- Export Excel -->
        <button
          @click="exportData('excel')"
          class="px-2.5 py-1.5 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1 bg-emerald-600 hover:bg-emerald-500 text-white border-emerald-400 shadow-xs"
        >
          <span>📊 Excel</span>
        </button>
      </div>
    </div>

    <!-- Non-AQMS Warning Banner -->
    <div v-if="!isAqmsDomain" class="p-3 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-700 dark:text-amber-300 text-xs font-mono flex items-center space-x-2">
      <span>⚠️</span>
      <span>
        Active deployment monitoring domain is <strong>{{ currentMonitoringDomain }}</strong>. ISPU calculations are regulatory-valid strictly for <strong>AQMS</strong> (Air Quality Monitoring Station) installations.
      </span>
    </div>

    <!-- MAIN DASHBOARD CONTENT -->
    <div v-else class="space-y-4">
      <!-- Top Section: Station Overview KPI & 24h Gauge Card -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-3">
        <!-- Station Main ISPU Gauge Card (approx 5 cols) -->
        <div
          class="lg:col-span-5 p-4 rounded-2xl border transition-all duration-300 flex flex-col justify-between shadow-xs"
          :class="kpiCardClass"
        >
          <div>
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-2">
                <span class="text-3xs font-mono font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                  {{ isOverallMode ? `LATEST STATION ISPU (${ispuMode === 'AGGREGATION' ? 'PER AGREGASI' : '24-HOUR'})` : `LATEST ISPU — ${selectedPollutant}` }}
                </span>
                <button
                  v-if="!isOverallMode"
                  @click="selectPollutant('OVERALL')"
                  class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-cyan-500/20 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30 hover:bg-cyan-500/30"
                  title="Kembali ke ISPU Stasiun (Dominan)"
                >
                  ← Overall
                </button>
              </div>
              <span
                class="px-2 py-0.5 rounded text-4xs font-mono font-bold uppercase border"
                :class="statusBadgeClass"
              >
                {{ latestData?.calculation_status || 'PROVISIONAL' }}
              </span>
            </div>

            <!-- Pollutant Name Header if specific concentrate selected -->
            <div v-if="!isOverallMode" class="text-2xs font-mono font-semibold text-slate-500 dark:text-slate-400 mt-1">
              {{ selectedPollutantInfo?.name }}
            </div>

            <!-- Main Metric Display -->
            <div class="flex items-baseline space-x-3 mt-2.5">
              <span class="text-5xl sm:text-6xl font-black font-mono tracking-tight tabular-nums" :class="categoryTextClass">
                {{ displayISPU }}
              </span>
              <div>
                <div class="text-lg font-bold font-mono uppercase tracking-wider" :class="categoryTextClass">
                  {{ ispuCategory }}
                </div>
                <div class="text-3xs font-mono text-slate-500 dark:text-slate-400">
                  <template v-if="isOverallMode">
                    Dominant: <strong class="text-slate-900 dark:text-white">{{ dominantPollutant }}</strong>
                    <span v-if="latestData?.tied_pollutants" class="text-4xs text-amber-500 ml-1">
                      (Tied: {{ latestData.tied_pollutants }})
                    </span>
                  </template>
                  <template v-else>
                    Conc: <strong class="text-slate-900 dark:text-white">{{ activeConcentration.toFixed(1) }}</strong> µg/m³
                    <span v-if="selectedPollutantResult?.xb !== undefined" class="text-4xs text-slate-400 ml-1">
                      [Bp: {{ selectedPollutantResult.xb }} - {{ selectedPollutantResult.xa }}]
                    </span>
                  </template>
                </div>
              </div>
            </div>
          </div>

          <!-- Measurement Period & Completeness Bar -->
          <div class="mt-4 pt-3 border-t border-slate-200/60 dark:border-slate-800/80 space-y-2 text-3xs font-mono">
            <div class="flex items-center justify-between text-slate-500 dark:text-slate-400">
              <span>Measurement Window:</span>
              <span class="font-bold text-slate-800 dark:text-slate-200">{{ formattedPeriod }}</span>
            </div>
            <div>
              <div class="flex items-center justify-between text-slate-500 dark:text-slate-400 mb-1">
                <span>{{ isOverallMode ? 'Data Completeness (Pasal 7 ayat 3):' : 'Hourly Data Samples:' }}</span>
                <span class="font-bold" :class="completenessPercent >= 75 ? 'text-emerald-500' : 'text-amber-500'">
                  {{ isOverallMode ? `${completenessPercent.toFixed(1)}% (Req: ≥ 75%)` : `${selectedPollutantResult?.valid_samples || 0}/24h (${completenessPercent.toFixed(1)}%)` }}
                </span>
              </div>
              <div class="w-full h-1.5 bg-slate-200 dark:bg-slate-800 rounded-full overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-500"
                  :class="completenessPercent >= 75 ? 'bg-emerald-500' : 'bg-amber-500'"
                  :style="{ width: `${Math.min(100, completenessPercent)}%` }"
                ></div>
              </div>
            </div>
            <div class="flex items-center justify-between text-4xs text-slate-400 dark:text-slate-500 pt-1">
              <span>Data Source: Aggregation Rollups</span>
              <span>Timezone: {{ latestData?.timezone || 'Asia/Jakarta' }}</span>
            </div>
          </div>
        </div>

        <!-- Historical Trend Chart (approx 7 cols) -->
        <div class="lg:col-span-7 p-3.5 rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] shadow-xs flex flex-col justify-between">
          <ISPUTrendChart
            :records="trendRecords"
            :loading="trendLoading"
            :initial-range="trendRange"
            :selected-pollutant="selectedPollutant"
            :pollutant-name="selectedPollutantInfo?.name"
            title="ISPU HISTORICAL TREND CHART"
            @range-change="handleRangeChange"
          />
        </div>
      </div>

      <!-- Quick Concentrate Selector & 7 Official ISPU Pollutants Breakdown -->
      <div class="space-y-2.5">
        <div class="flex flex-wrap items-center justify-between gap-2 text-3xs font-mono">
          <div class="flex items-center space-x-2">
            <span class="w-2 h-2 rounded-full bg-cyan-500"></span>
            <span class="font-bold uppercase tracking-wider text-slate-700 dark:text-slate-200">
              OFFICIAL 7 PARAMETERS EVALUATION BREAKDOWN
            </span>
            <span class="text-slate-400 hidden sm:inline">(Permen LHK No. 14/2020 Lampiran I & II)</span>
          </div>

          <!-- Quick Selector Buttons -->
          <div class="flex flex-wrap items-center gap-1">
            <button
              @click="selectPollutant('OVERALL')"
              class="px-2 py-0.5 rounded text-4xs font-mono font-bold transition border flex items-center space-x-1"
              :class="selectedPollutant === 'OVERALL'
                ? 'bg-cyan-500 text-slate-950 border-cyan-400 shadow-xs ring-1 ring-cyan-400'
                : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B] hover:text-slate-900 dark:hover:text-white'"
            >
              <span>★ Overall</span>
            </button>
            <button
              v-for="p in sevenOfficialParameters"
              :key="p.code"
              @click="selectPollutant(p.code)"
              class="px-2 py-0.5 rounded text-4xs font-mono font-bold transition border"
              :class="selectedPollutant === p.code
                ? 'bg-cyan-500 text-slate-950 border-cyan-400 shadow-xs ring-1 ring-cyan-400'
                : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B] hover:text-slate-900 dark:hover:text-white'"
            >
              {{ p.code }}
            </button>
          </div>
        </div>

        <!-- 7 Cards Grid (Clickable to display concentrate on large card & chart) -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7 gap-2.5">
          <div
            v-for="p in sevenOfficialParameters"
            :key="p.code"
            @click="selectPollutant(p.code)"
            class="p-3 rounded-xl border flex flex-col justify-between transition cursor-pointer select-none shadow-xs"
            :class="getParameterCardClass(p.code)"
            :title="`Click to display ${p.code} ISPU on main card & trend chart`"
          >
            <div>
              <!-- Header: Code & Category Badge -->
              <div class="flex items-start justify-between">
                <div>
                  <div class="flex items-center space-x-1.5">
                    <span class="text-sm font-bold font-mono tracking-tight text-slate-900 dark:text-white">
                      {{ p.code }}
                    </span>
                    <span
                      v-if="selectedPollutant === p.code"
                      class="px-1 py-0.2 rounded text-[8px] font-mono font-bold bg-cyan-500 text-slate-950 uppercase"
                    >
                      Active
                    </span>
                  </div>
                  <div class="text-[10px] text-slate-400 truncate max-w-[110px]" :title="p.name">
                    {{ p.name }}
                  </div>
                </div>
                <span
                  class="px-1.5 py-0.2 rounded text-[9px] font-mono font-bold uppercase border"
                  :class="getParameterBadgeClass(p.code)"
                >
                  {{ getParameterDisplayCategory(p.code) }}
                </span>
              </div>

              <!-- Main ISPU Index & Concentration -->
              <div class="mt-2.5">
                <div class="flex items-baseline space-x-1">
                  <span class="text-2xl font-black font-mono tabular-nums text-slate-900 dark:text-white">
                    {{ getParameterDisplayISPU(p.code) }}
                  </span>
                  <span class="text-3xs font-mono text-slate-400">ISPU</span>
                </div>
                <div class="text-3xs font-mono text-slate-500 dark:text-slate-400 mt-0.5">
                  Conc: <strong class="text-slate-700 dark:text-slate-200">{{ getParameterDisplayConc(p.code) }}</strong> µg/m³
                </div>
              </div>
            </div>

            <!-- Footer: Breakpoints & Inspect Math -->
            <div class="mt-3 pt-2 border-t border-slate-200/60 dark:border-slate-800/80 flex items-center justify-between text-[10px] font-mono">
              <span class="text-slate-400">
                {{ getParameterResult(p.code)?.valid_samples || 0 }}/24h
              </span>
              <button
                @click.stop="openFormulaModal(p.code)"
                class="text-cyan-600 dark:text-cyan-400 hover:underline cursor-pointer"
                title="Inspect regulatory interpolation formula"
              >
                Inspect →
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Historical Records Table & Server Filters -->
      <div class="p-3.5 rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] shadow-xs space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <div class="flex items-center space-x-2">
            <span class="w-2 h-2 rounded-full bg-blue-500"></span>
            <h3 class="text-xs font-bold font-mono uppercase tracking-wider text-slate-800 dark:text-white">
              HISTORICAL ISPU CALCULATION RECORDS
            </h3>
            <span class="text-3xs text-slate-400 font-mono">
              ({{ historyTotal }} records found)
            </span>
          </div>

          <!-- Filters -->
          <div class="flex flex-wrap items-center gap-1.5 text-3xs font-mono">
            <!-- Category Filter -->
            <select
              v-model="filterCategory"
              @change="fetchHistory(1)"
              class="px-2 py-1 rounded-lg bg-slate-100 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-300 focus:outline-hidden"
            >
              <option value="">All Categories</option>
              <option value="BAIK">Baik</option>
              <option value="SEDANG">Sedang</option>
              <option value="TIDAK_SEHAT">Tidak Sehat</option>
              <option value="SANGAT_TIDAK_SEHAT">Sangat Tidak Sehat</option>
              <option value="BERBAHAYA">Berbahaya</option>
            </select>

            <!-- Status Filter -->
            <select
              v-model="filterStatus"
              @change="fetchHistory(1)"
              class="px-2 py-1 rounded-lg bg-slate-100 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-300 focus:outline-hidden"
            >
              <option value="">All Status</option>
              <option value="VALID">Valid (≥75%)</option>
              <option value="PROVISIONAL">Provisional (&lt;75%)</option>
              <option value="INVALID">Invalid</option>
            </select>

            <!-- Dominant Pollutant Filter -->
            <select
              v-model="filterDominant"
              @change="fetchHistory(1)"
              class="px-2 py-1 rounded-lg bg-slate-100 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-300 focus:outline-hidden"
            >
              <option value="">All Pollutants</option>
              <option v-for="p in sevenOfficialParameters" :key="p.code" :value="p.code">
                {{ p.code }}
              </option>
            </select>
          </div>
        </div>

        <!-- Table -->
        <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-[#20354B]">
          <table class="w-full text-left text-3xs font-mono select-none">
            <thead class="bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-b border-slate-200 dark:border-[#20354B]">
              <tr>
                <th class="py-2 px-3">Period (24h)</th>
                <th class="py-2 px-3">Source Aggregation</th>
                <th class="py-2 px-3">Calculated At</th>
                <th class="py-2 px-3" :class="{ 'text-cyan-400 font-bold': !isOverallMode }">
                  {{ isOverallMode ? 'ISPU' : `ISPU (${selectedPollutant})` }}
                </th>
                <th class="py-2 px-3" :class="{ 'text-cyan-400 font-bold': !isOverallMode }">
                  {{ isOverallMode ? 'Category' : `Category (${selectedPollutant})` }}
                </th>
                <th class="py-2 px-3">Dominant</th>
                <th class="py-2 px-3">Coverage</th>
                <th class="py-2 px-3">Status</th>
                <th
                  v-for="p in sevenOfficialParameters"
                  :key="p.code"
                  class="py-2 px-3 transition cursor-pointer"
                  :class="selectedPollutant === p.code ? 'bg-cyan-500/20 text-cyan-400 font-extrabold' : 'hover:text-cyan-400'"
                  @click="selectPollutant(p.code)"
                  :title="`Klik untuk fokus pada ${p.code}`"
                >
                  {{ p.code }}
                </th>
                <th class="py-2 px-3 text-right">Action</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
              <tr v-if="historyLoading">
                <td colspan="16" class="py-6 text-center text-slate-400">Loading records...</td>
              </tr>
              <tr v-else-if="!historyRecords || historyRecords.length === 0">
                <td colspan="16" class="py-6 text-center text-slate-400">No records found matching filters</td>
              </tr>
              <tr
                v-for="r in historyRecords"
                :key="r.id"
                class="hover:bg-slate-50 dark:hover:bg-[#101B2B]/60 transition cursor-pointer"
                @click="openDetailModal(r)"
              >
                <td class="py-2 px-3 font-bold text-slate-800 dark:text-slate-200">
                  <div>{{ formatDate(r.period_end) }}</div>
                  <div class="text-4xs text-slate-400 font-normal">24h: {{ formatShortTime(r.period_start) }} → {{ formatShortTime(r.period_end) }}</div>
                </td>
                <td class="py-2 px-3 font-mono">
                  <div class="font-bold text-cyan-600 dark:text-cyan-400">{{ r.source_bucket_identifier || '-' }}</div>
                  <div class="text-4xs text-slate-400 font-normal">{{ r.aggregation_interval_seconds || 120 }}s {{ r.aggregation_function || 'AVG' }}</div>
                </td>
                <td class="py-2 px-3 font-mono text-slate-500 dark:text-slate-400 whitespace-nowrap text-3xs">
                  {{ formatDateTime(r.calculated_at) }}
                </td>
                <!-- Main ISPU column (overall or selected pollutant) -->
                <td class="py-2 px-3 font-bold text-sm tabular-nums" :style="{ color: getRowDisplayColor(r) }">
                  {{ getRowDisplayISPU(r) }}
                </td>
                <!-- Main Category column (overall or selected pollutant) -->
                <td class="py-2 px-3">
                  <span class="px-1.5 py-0.2 rounded text-4xs font-bold border" :class="getCategoryPillClass(getRowDisplayCategory(r))">
                    {{ getRowDisplayCategory(r) }}
                  </span>
                </td>
                <td class="py-2 px-3 font-bold text-slate-700 dark:text-slate-300">
                  {{ r.dominant_pollutant }}
                </td>
                <td class="py-2 px-3 tabular-nums">
                  {{ (r.completeness_percent || 0).toFixed(1) }}%
                </td>
                <td class="py-2 px-3">
                  <span
                    class="px-1.5 py-0.2 rounded text-4xs font-bold"
                    :class="r.calculation_status === 'VALID' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-amber-500/20 text-amber-400'"
                  >
                    {{ r.calculation_status }}
                  </span>
                </td>
                <!-- 7 Pollutant Columns with Active Selection Highlight -->
                <td
                  v-for="p in sevenOfficialParameters"
                  :key="p.code"
                  class="py-2 px-3 tabular-nums transition"
                  :class="selectedPollutant === p.code ? 'bg-cyan-500/15 text-cyan-400 font-black' : ''"
                >
                  {{ getParamVal(r, p.code) }}
                </td>
                <td class="py-2 px-3 text-right">
                  <button
                    @click.stop="openDetailModal(r)"
                    class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 hover:bg-cyan-500/20 hover:text-cyan-400 transition"
                  >
                    Detail
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="flex items-center justify-between text-3xs font-mono text-slate-500 pt-1">
          <div>
            Page {{ historyPage }} of {{ totalPages || 1 }}
          </div>
          <div class="flex items-center space-x-1">
            <button
              :disabled="historyPage <= 1"
              @click="fetchHistory(historyPage - 1)"
              class="px-2.5 py-1 rounded bg-slate-100 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] disabled:opacity-30 hover:border-cyan-500 transition"
            >
              Prev
            </button>
            <button
              :disabled="historyPage >= totalPages"
              @click="fetchHistory(historyPage + 1)"
              class="px-2.5 py-1 rounded bg-slate-100 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] disabled:opacity-30 hover:border-cyan-500 transition"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- DETAIL / FORMULA INSPECTOR MODAL -->
    <div
      v-if="showDetailModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-3 bg-slate-950/75 backdrop-blur-sm"
      @click.self="showDetailModal = false"
    >
      <div class="w-full max-w-3xl bg-white dark:bg-[#0B1422] rounded-2xl border border-slate-200 dark:border-[#20354B] shadow-2xl p-5 space-y-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-[#20354B]">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-400"></span>
            <h3 class="text-sm font-bold font-mono text-slate-900 dark:text-white uppercase">
              ISPU CALCULATION REPRODUCTION & AUDIT DETAILS
            </h3>
          </div>
          <button @click="showDetailModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
        </div>

        <!-- Traceability & Provenance Information Card -->
        <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-[#070D18] border border-slate-200 dark:border-[#1E293B] font-mono text-xs space-y-2">
          <div class="text-2xs font-bold text-slate-400 uppercase tracking-wider flex items-center justify-between">
            <span class="text-cyan-600 dark:text-cyan-400 font-bold">Provenance & Traceability Metadata</span>
            <span class="px-2 py-0.5 rounded text-4xs font-bold bg-cyan-500/15 text-cyan-400 border border-cyan-500/30">
              {{ selectedRecord?.trigger_reason || 'AGGREGATION_COMPLETED' }}
            </span>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 text-3xs pt-1 border-t border-slate-200 dark:border-slate-800">
            <div>
              <span class="text-slate-400 block">Source Bucket:</span>
              <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ selectedRecord?.source_bucket_identifier || '-' }}</span>
            </div>
            <div>
              <span class="text-slate-400 block">Aggregation Setup:</span>
              <span class="font-bold text-slate-800 dark:text-slate-200">{{ selectedRecord?.aggregation_interval_seconds || 120 }}s · {{ selectedRecord?.aggregation_function || 'AVG' }}</span>
            </div>
            <div>
              <span class="text-slate-400 block">Calculation Period (24h):</span>
              <span class="font-bold text-slate-800 dark:text-slate-200">{{ formatDate(selectedRecord?.period_start) }} → {{ formatDate(selectedRecord?.period_end) }}</span>
            </div>
            <div>
              <span class="text-slate-400 block">Calculated At:</span>
              <span class="font-bold text-slate-800 dark:text-slate-200">{{ formatDateTime(selectedRecord?.calculated_at) }}</span>
            </div>
          </div>
        </div>

        <!-- Formula Box -->
        <div class="p-3.5 rounded-xl bg-slate-100 dark:bg-[#070D18] border border-slate-200 dark:border-[#1E293B] font-mono text-xs space-y-1.5">
          <div class="text-cyan-600 dark:text-cyan-400 font-bold">
            Mathematical Equation (Permen LHK No. 14 Tahun 2020 Lampiran I Persamaan 1):
          </div>
          <div class="text-slate-700 dark:text-slate-300 font-bold text-sm bg-white dark:bg-[#0B1422] p-2 rounded border border-slate-200 dark:border-slate-800">
            I = [ (Ia - Ib) / (Xa - Xb) ] × (Xx - Xb) + Ib
          </div>
          <div class="text-3xs text-slate-400">
            I = Calculated ISPU · Ia, Ib = Upper/Lower ISPU limits · Xa, Xb = Upper/Lower concentration limits · Xx = Measured concentration (µg/m³)
          </div>
        </div>

        <!-- Parameter Detail Table -->
        <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-[#20354B]">
          <table class="w-full text-left text-3xs font-mono">
            <thead class="bg-slate-100 dark:bg-[#101B2B] text-slate-500 border-b border-slate-200 dark:border-[#20354B]">
              <tr>
                <th class="p-2">Pollutant</th>
                <th class="p-2">Conc 24h</th>
                <th class="p-2">Conc Aggr</th>
                <th class="p-2">Buckets</th>
                <th class="p-2">Latest Bucket ID</th>
                <th class="p-2">Breakpoint [Xb, Xa]</th>
                <th class="p-2">Index [Ib, Ia]</th>
                <th class="p-2">Formula Evaluation</th>
                <th class="p-2">ISPU</th>
                <th class="p-2">Category</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-for="pr in selectedRecord?.parameter_results || []" :key="pr.id">
                <td class="p-2 font-bold text-slate-900 dark:text-white">{{ pr.pollutant_code }}</td>
                <td class="p-2 tabular-nums">{{ pr.normalized_concentration.toFixed(1) }} µg/m³</td>
                <td class="p-2 tabular-nums">{{ (pr.aggregation_concentration || 0).toFixed(1) }} µg/m³</td>
                <td class="p-2 tabular-nums">{{ pr.source_aggregation_count || pr.valid_samples || 0 }}</td>
                <td class="p-2 font-bold text-cyan-600 dark:text-cyan-400">{{ pr.source_latest_identifier || '-' }}</td>
                <td class="p-2 tabular-nums">[{{ pr.xb }}, {{ pr.xa }}]</td>
                <td class="p-2 tabular-nums">[{{ pr.ib }}, {{ pr.ia }}]</td>
                <td class="p-2 text-4xs text-slate-400">{{ pr.formula || '-' }}</td>
                <td class="p-2 font-bold text-sm tabular-nums" :style="{ color: getCategoryColor(pr.category) }">
                  {{ Math.round(pr.calculated_ispu) }}
                </td>
                <td class="p-2 font-bold">{{ pr.category }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="flex justify-end pt-2">
          <button
            @click="showDetailModal = false"
            class="px-4 py-1.5 rounded-lg bg-slate-200 dark:bg-slate-800 text-slate-800 dark:text-slate-200 text-xs font-mono font-bold hover:bg-slate-300 dark:hover:bg-slate-700"
          >
            Close
          </button>
        </div>
      </div>
    </div>

    <!-- PARAMETER MAPPINGS MODAL (Admin/Superadmin) -->
    <div
      v-if="showMappingsModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-3 bg-slate-950/75 backdrop-blur-sm"
      @click.self="showMappingsModal = false"
    >
      <div class="w-full max-w-2xl bg-white dark:bg-[#0B1422] rounded-2xl border border-slate-200 dark:border-[#20354B] shadow-2xl p-5 space-y-4">
        <div class="flex items-center justify-between pb-3 border-b border-slate-200 dark:border-[#20354B]">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-400"></span>
            <h3 class="text-sm font-bold font-mono text-slate-900 dark:text-white uppercase">
              STATION PARAMETER MAPPINGS (ISPU COMPLIANCE)
            </h3>
          </div>
          <button @click="showMappingsModal = false" class="text-slate-400 hover:text-white text-lg">✕</button>
        </div>

        <p class="text-3xs text-slate-400 font-mono">
          Explicitly map physical datalogger telemetry parameters to the seven official pollutants regulated by Permen LHK No. 14 Tahun 2020.
        </p>

        <!-- Current Mappings Table -->
        <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-[#20354B]">
          <table class="w-full text-left text-3xs font-mono">
            <thead class="bg-slate-100 dark:bg-[#101B2B] text-slate-500 border-b border-slate-200 dark:border-[#20354B]">
              <tr>
                <th class="p-2">Official Pollutant</th>
                <th class="p-2">Mapped Parameter</th>
                <th class="p-2">Unit</th>
                <th class="p-2">Factor</th>
                <th class="p-2">Min Coverage</th>
                <th class="p-2">Status</th>
                <th class="p-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-if="stationMappings.length === 0">
                <td colspan="7" class="p-4 text-center text-slate-400">No parameter mappings configured yet</td>
              </tr>
              <tr v-for="m in stationMappings" :key="m.id">
                <td class="p-2 font-bold text-cyan-400">{{ m.pollutant_code }}</td>
                <td class="p-2">{{ m.parameter?.parameter_name || m.parameter?.parameter_code || m.parameter_id }}</td>
                <td class="p-2">{{ m.source_unit }}</td>
                <td class="p-2">{{ m.conversion_factor }}</td>
                <td class="p-2">{{ m.min_completeness }}%</td>
                <td class="p-2">
                  <span class="px-1.5 py-0.2 rounded text-4xs font-bold" :class="m.enabled ? 'bg-emerald-500/20 text-emerald-400' : 'bg-slate-500/20 text-slate-400'">
                    {{ m.enabled ? 'ENABLED' : 'DISABLED' }}
                  </span>
                </td>
                <td class="p-2 text-right space-x-1">
                  <button @click="deleteMapping(m.id)" class="text-rose-500 hover:text-rose-400">Delete</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Add Mapping Form -->
        <div class="p-3 rounded-xl bg-slate-100 dark:bg-[#070D18] border border-slate-200 dark:border-[#1E293B] space-y-3 font-mono text-3xs">
          <div class="font-bold text-slate-800 dark:text-slate-200 uppercase">Add / Update Parameter Mapping</div>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
            <div>
              <label class="block text-slate-400 mb-1">Official Pollutant</label>
              <select v-model="newMapping.pollutant_code" class="w-full p-1.5 rounded bg-white dark:bg-[#0B1422] border border-slate-300 dark:border-slate-700">
                <option v-for="p in sevenOfficialParameters" :key="p.code" :value="p.code">
                  {{ p.code }} ({{ p.name }})
                </option>
              </select>
            </div>
            <div>
              <label class="block text-slate-400 mb-1">Datalogger Parameter</label>
              <select v-model="newMapping.parameter_id" class="w-full p-1.5 rounded bg-white dark:bg-[#0B1422] border border-slate-300 dark:border-slate-700">
                <option v-for="p in stationParameters" :key="p.id" :value="p.id">
                  {{ p.parameter_code }} - {{ p.parameter_name }} ({{ p.unit }})
                </option>
              </select>
            </div>
            <div>
              <label class="block text-slate-400 mb-1">Source Unit</label>
              <input v-model="newMapping.source_unit" class="w-full p-1.5 rounded bg-white dark:bg-[#0B1422] border border-slate-300 dark:border-slate-700" placeholder="µg/m³" />
            </div>
          </div>

          <div class="flex justify-end pt-1">
            <button
              @click="saveMapping"
              class="px-3 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white font-bold"
            >
              Save Mapping
            </button>
          </div>
        </div>

        <div class="flex justify-end">
          <button
            @click="showMappingsModal = false"
            class="px-4 py-1.5 rounded-lg bg-slate-200 dark:bg-slate-800 text-slate-800 dark:text-slate-200 text-xs font-mono font-bold"
          >
            Close
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import ISPUTrendChart from '@/components/ISPUTrendChart.vue';

export default {
  name: 'ISPUAnalysisView',
  components: {
    ISPUTrendChart
  },
  data() {
    return {
      selectedDeviceId: null,
      devicesList: [],
      latestData: null,
      trendRecords: [],
      trendLoading: false,
      trendRange: '24h',
      recalculating: false,

      // Calculation Mode: '24H' (regulatory rolling 24h) vs 'AGGREGATION' (instantaneous per 2-min rollup)
      ispuMode: '24H',

      // Selected Pollutant: 'OVERALL' or specific pollutant code ('PM25', 'PM10', etc.)
      selectedPollutant: 'OVERALL',

      // Seven Official ISPU Pollutants according to Permen LHK 14/2020
      sevenOfficialParameters: [
        { code: 'PM25', name: 'Partikulat PM2.5 (24h)' },
        { code: 'PM10', name: 'Partikulat PM10 (24h)' },
        { code: 'SO2', name: 'Sulfur Dioksida (24h)' },
        { code: 'CO', name: 'Karbon Monoksida (24h)' },
        { code: 'O3', name: 'Ozon (24h)' },
        { code: 'NO2', name: 'Nitrogen Dioksida (24h)' },
        { code: 'HC', name: 'Hidrokarbon (24h)' }
      ],

      // History
      historyRecords: [],
      historyTotal: 0,
      historyPage: 1,
      historyLimit: 15,
      historyLoading: false,
      filterCategory: '',
      filterStatus: '',
      filterDominant: '',

      // Modals
      showDetailModal: false,
      selectedRecord: null,
      showMappingsModal: false,
      stationMappings: [],
      stationParameters: [],
      newMapping: {
        pollutant_code: 'PM25',
        parameter_id: null,
        source_unit: 'µg/m³',
        conversion_factor: 1.0,
        min_completeness: 75.0,
        enabled: true
      },

      systemDomain: 'AQMS'
    };
  },
  computed: {
    isAqmsDomain() {
      return (this.systemDomain || 'AQMS').toUpperCase() === 'AQMS';
    },
    currentMonitoringDomain() {
      return this.systemDomain || 'AQMS';
    },
    isOverallMode() {
      return !this.selectedPollutant || this.selectedPollutant === 'OVERALL';
    },
    selectedPollutantInfo() {
      if (this.isOverallMode) return null;
      return this.sevenOfficialParameters.find(p => p.code === this.selectedPollutant) || null;
    },
    selectedPollutantResult() {
      if (this.isOverallMode) return null;
      return this.getParameterResult(this.selectedPollutant);
    },
    displayISPU() {
      if (this.isOverallMode) {
        if (!this.latestData) return '-';
        if (this.ispuMode === 'AGGREGATION') {
          return Math.round(this.latestData.aggregation_ispu || this.latestData.overall_ispu || 0);
        }
        return Math.round(this.latestData.overall_ispu || 0);
      }
      const pr = this.selectedPollutantResult;
      if (!pr) return '-';
      if (this.ispuMode === 'AGGREGATION') {
        return Math.round(pr.aggregation_ispu || pr.calculated_ispu || 0);
      }
      return Math.round(pr.calculated_ispu || 0);
    },
    ispuCategory() {
      if (this.isOverallMode) {
        if (!this.latestData) return 'TIDAK_TERSEDIA';
        if (this.ispuMode === 'AGGREGATION') {
          return this.latestData.aggregation_category || this.latestData.overall_category || 'TIDAK_TERSEDIA';
        }
        return this.latestData.overall_category || 'TIDAK_TERSEDIA';
      }
      const pr = this.selectedPollutantResult;
      if (!pr) return 'TIDAK_TERSEDIA';
      if (this.ispuMode === 'AGGREGATION') {
        return pr.aggregation_category || pr.category || 'TIDAK_TERSEDIA';
      }
      return pr.category || 'TIDAK_TERSEDIA';
    },
    dominantPollutant() {
      if (!this.latestData) return '-';
      if (this.ispuMode === 'AGGREGATION') {
        return this.latestData.aggregation_dominant || this.latestData.dominant_pollutant || '-';
      }
      return this.latestData.dominant_pollutant || '-';
    },
    activeConcentration() {
      const pr = this.selectedPollutantResult;
      if (!pr) return 0;
      if (this.ispuMode === 'AGGREGATION') {
        return pr.aggregation_concentration || pr.normalized_concentration || 0;
      }
      return pr.normalized_concentration || 0;
    },
    completenessPercent() {
      if (this.isOverallMode) {
        return (this.latestData && this.latestData.completeness_percent) || 0;
      }
      return (this.selectedPollutantResult && this.selectedPollutantResult.completeness_percent) || 0;
    },
    totalPages() {
      return Math.ceil(this.historyTotal / this.historyLimit);
    },
    formattedPeriod() {
      if (!this.latestData || !this.latestData.period_end) return '-';
      const end = new Date(this.latestData.period_end);
      const start = new Date(this.latestData.period_start || end.getTime() - 24 * 3600 * 1000);
      const fmt = d => d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });
      return `${fmt(start)} - ${fmt(end)} WIB`;
    },
    kpiCardClass() {
      switch (this.ispuCategory) {
        case 'BAIK': return 'bg-emerald-500/10 border-emerald-500/30';
        case 'SEDANG': return 'bg-blue-500/10 border-blue-500/30';
        case 'TIDAK_SEHAT': return 'bg-amber-500/10 border-amber-500/30';
        case 'SANGAT_TIDAK_SEHAT': return 'bg-rose-500/10 border-rose-500/30';
        case 'BERBAHAYA': return 'bg-purple-900/20 border-purple-500/40';
        default: return 'bg-white dark:bg-[#0B1422] border-slate-200 dark:border-[#20354B]';
      }
    },
    categoryTextClass() {
      switch (this.ispuCategory) {
        case 'BAIK': return 'text-emerald-600 dark:text-emerald-400';
        case 'SEDANG': return 'text-blue-600 dark:text-blue-400';
        case 'TIDAK_SEHAT': return 'text-amber-600 dark:text-amber-400';
        case 'SANGAT_TIDAK_SEHAT': return 'text-rose-600 dark:text-rose-400';
        case 'BERBAHAYA': return 'text-purple-600 dark:text-purple-400';
        default: return 'text-slate-600 dark:text-slate-400';
      }
    },
    statusBadgeClass() {
      const st = this.latestData?.calculation_status;
      if (st === 'VALID') return 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-300 border-emerald-500/40';
      if (st === 'PROVISIONAL') return 'bg-amber-500/20 text-amber-600 dark:text-amber-300 border-amber-500/40';
      return 'bg-rose-500/20 text-rose-600 dark:text-rose-300 border-rose-500/40';
    },
    canManageMapping() {
      return true;
    }
  },
  async mounted() {
    await this.fetchDomain();
    await this.fetchDevices();
    if (this.$route.query.device_id) {
      this.selectedDeviceId = Number(this.$route.query.device_id);
    }
    if (this.selectedDeviceId) {
      this.loadStationData();
    }
    window.addEventListener('ispu-calculated-event', this.handleISPUCalculated);
    window.addEventListener('aggregation-result-event', this.handleAggregationUpdated);
  },
  beforeDestroy() {
    window.removeEventListener('ispu-calculated-event', this.handleISPUCalculated);
    window.removeEventListener('aggregation-result-event', this.handleAggregationUpdated);
  },
  methods: {
    selectPollutant(code) {
      this.selectedPollutant = code;
    },
    async fetchDomain() {
      try {
        const res = await axios.get('/api/system/domain');
        if (res.data && res.data.data) {
          this.systemDomain = res.data.data.monitoring_domain || 'AQMS';
        }
      } catch (e) {
        this.systemDomain = 'AQMS';
      }
    },
    async fetchDevices() {
      try {
        const token = localStorage.getItem('token');
        const res = await axios.get('/api/devices', {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        const items = res.data?.data || res.data || [];
        this.devicesList = Array.isArray(items) ? items : [];
        if (!this.selectedDeviceId && this.devicesList.length > 0) {
          this.selectedDeviceId = this.devicesList[0].id;
        }
      } catch (err) {
        console.error('Failed to load devices:', err);
      }
    },
    handleDeviceChange() {
      this.$router.replace({ query: { ...this.$route.query, device_id: this.selectedDeviceId } });
      this.loadStationData();
    },
    async loadStationData() {
      if (!this.selectedDeviceId) return;
      await Promise.all([
        this.fetchLatestISPU(),
        this.fetchTrend(),
        this.fetchHistory(1)
      ]);
    },
    async fetchLatestISPU() {
      try {
        const token = localStorage.getItem('token');
        const res = await axios.get(`/api/ispu/latest?device_id=${this.selectedDeviceId}`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        if (res.data?.success && res.data?.data) {
          this.latestData = res.data.data;
        }
      } catch (err) {
        console.error('Failed to fetch latest ISPU:', err);
      }
    },
    async fetchTrend() {
      this.trendLoading = true;
      try {
        const token = localStorage.getItem('token');
        const res = await axios.get(`/api/ispu/trend?device_id=${this.selectedDeviceId}&range=${this.trendRange}`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        if (res.data?.success && res.data?.data) {
          this.trendRecords = res.data.data;
        }
      } catch (err) {
        console.error('Failed to fetch trend:', err);
      } finally {
        this.trendLoading = false;
      }
    },
    handleRangeChange(range) {
      this.trendRange = range;
      this.fetchTrend();
    },
    async fetchHistory(page = 1) {
      this.historyLoading = true;
      this.historyPage = page;
      try {
        const token = localStorage.getItem('token');
        let url = `/api/ispu/history?device_id=${this.selectedDeviceId}&page=${page}&limit=${this.historyLimit}`;
        if (this.filterCategory) url += `&category=${this.filterCategory}`;
        if (this.filterStatus) url += `&status=${this.filterStatus}`;
        if (this.filterDominant) url += `&dominant_pollutant=${this.filterDominant}`;

        const res = await axios.get(url, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        if (res.data?.success && res.data?.data) {
          this.historyRecords = res.data.data.items || [];
          this.historyTotal = res.data.data.total || 0;
        }
      } catch (err) {
        console.error('Failed to fetch history:', err);
      } finally {
        this.historyLoading = false;
      }
    },
    handleISPUCalculated(event) {
      const data = event.detail;
      if (data && data.device_id === this.selectedDeviceId) {
        this.latestData = data;
        this.fetchTrend();
        this.fetchHistory(this.historyPage);
      }
    },
    handleAggregationUpdated(event) {
      const data = event.detail;
      if (data && data.device_id === this.selectedDeviceId) {
        if (this.aggTimer) clearTimeout(this.aggTimer);
        this.aggTimer = setTimeout(() => {
          this.fetchLatestISPU();
        }, 1000);
      }
    },
    async triggerRecalculation() {
      if (!this.selectedDeviceId || this.recalculating) return;
      this.recalculating = true;
      try {
        const token = localStorage.getItem('token');
        await Promise.all([
          axios.post('/api/ispu/backfill', { device_id: this.selectedDeviceId, hours: 24 }, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          }),
          axios.post('/api/ispu/recalculate', { device_id: this.selectedDeviceId }, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          })
        ]);
        await this.loadStationData();
      } catch (err) {
        alert('Recalculation error: ' + (err.response?.data?.error || err.message));
      } finally {
        this.recalculating = false;
      }
    },
    getParameterResult(pollutantCode) {
      if (!this.latestData || !this.latestData.parameter_results) return null;
      return this.latestData.parameter_results.find(pr => pr.pollutant_code === pollutantCode);
    },
    getParameterDisplayISPU(pollutantCode) {
      const res = this.getParameterResult(pollutantCode);
      if (!res) return 0;
      if (this.ispuMode === 'AGGREGATION') {
        return Math.round(res.aggregation_ispu || res.calculated_ispu || 0);
      }
      return Math.round(res.calculated_ispu || 0);
    },
    getParameterDisplayConc(pollutantCode) {
      const res = this.getParameterResult(pollutantCode);
      if (!res) return '0.0';
      if (this.ispuMode === 'AGGREGATION') {
        return (res.aggregation_concentration || res.normalized_concentration || 0).toFixed(1);
      }
      return (res.normalized_concentration || 0).toFixed(1);
    },
    getParameterDisplayCategory(pollutantCode) {
      const res = this.getParameterResult(pollutantCode);
      if (!res) return 'UNAVAIL';
      if (this.ispuMode === 'AGGREGATION') {
        return res.aggregation_category || res.category || 'UNAVAIL';
      }
      return res.category || 'UNAVAIL';
    },
    getParameterCardClass(pollutantCode) {
      const isSelected = this.selectedPollutant === pollutantCode;
      const res = this.getParameterResult(pollutantCode);
      let baseClass = '';

      if (!res) {
        baseClass = 'bg-slate-50 dark:bg-[#070D18] border-slate-200 dark:border-[#1E293B] opacity-60';
      } else {
        const cat = this.getParameterDisplayCategory(pollutantCode);
        switch (cat) {
          case 'BAIK': baseClass = 'bg-emerald-500/5 border-emerald-500/30'; break;
          case 'SEDANG': baseClass = 'bg-blue-500/5 border-blue-500/30'; break;
          case 'TIDAK_SEHAT': baseClass = 'bg-amber-500/5 border-amber-500/30'; break;
          case 'SANGAT_TIDAK_SEHAT': baseClass = 'bg-rose-500/5 border-rose-500/30'; break;
          case 'BERBAHAYA': baseClass = 'bg-purple-900/15 border-purple-500/40'; break;
          default: baseClass = 'bg-slate-50 dark:bg-[#070D18] border-slate-200 dark:border-[#1E293B]'; break;
        }
      }

      if (isSelected) {
        return `${baseClass} ring-2 ring-cyan-400 border-cyan-400 dark:border-cyan-400 bg-cyan-500/15 dark:bg-cyan-500/20 shadow-lg shadow-cyan-500/20 scale-[1.02]`;
      }
      return `${baseClass} hover:border-cyan-500/50 hover:scale-[1.01]`;
    },
    getParameterBadgeClass(pollutantCode) {
      const cat = this.getParameterDisplayCategory(pollutantCode);
      return this.getCategoryPillClass(cat);
    },
    getCategoryPillClass(cat) {
      switch (cat) {
        case 'BAIK': return 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-300 border-emerald-500/40';
        case 'SEDANG': return 'bg-blue-500/20 text-blue-600 dark:text-blue-300 border-blue-500/40';
        case 'TIDAK_SEHAT': return 'bg-amber-500/20 text-amber-600 dark:text-amber-300 border-amber-500/40';
        case 'SANGAT_TIDAK_SEHAT': return 'bg-rose-500/20 text-rose-600 dark:text-rose-300 border-rose-500/40';
        case 'BERBAHAYA': return 'bg-purple-900/40 text-purple-300 border-purple-500/50';
        default: return 'bg-slate-200 dark:bg-slate-800 text-slate-500 border-slate-300 dark:border-slate-700';
      }
    },
    getCategoryColor(cat) {
      switch (cat) {
        case 'BAIK': return '#10B981';
        case 'SEDANG': return '#0EA5E9';
        case 'TIDAK_SEHAT': return '#F59E0B';
        case 'SANGAT_TIDAK_SEHAT': return '#F43F5E';
        case 'BERBAHAYA': return '#8B5CF6';
        default: return '#64748B';
      }
    },
    getParamVal(record, pollutantCode) {
      if (!record || !record.parameter_results) return '-';
      const pr = record.parameter_results.find(p => p.pollutant_code === pollutantCode);
      return pr ? Math.round(pr.calculated_ispu) : '-';
    },
    getParamCategory(record, pollutantCode) {
      if (!record || !record.parameter_results) return 'UNAVAIL';
      const pr = record.parameter_results.find(p => p.pollutant_code === pollutantCode);
      return pr ? pr.category : 'UNAVAIL';
    },
    getRowDisplayISPU(record) {
      if (!record) return '-';
      if (this.isOverallMode) return Math.round(record.overall_ispu);
      return this.getParamVal(record, this.selectedPollutant);
    },
    getRowDisplayCategory(record) {
      if (!record) return 'TIDAK_TERSEDIA';
      if (this.isOverallMode) return record.overall_category;
      return this.getParamCategory(record, this.selectedPollutant);
    },
    getRowDisplayColor(record) {
      return this.getCategoryColor(this.getRowDisplayCategory(record));
    },
    formatDate(isoStr) {
      if (!isoStr) return '-';
      const d = new Date(isoStr);
      return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });
    },
    formatDateTime(isoStr) {
      if (!isoStr) return '-';
      const d = new Date(isoStr);
      return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    formatShortTime(isoStr) {
      if (!isoStr) return '-';
      const d = new Date(isoStr);
      return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
    },
    openFormulaModal(pollutantCode) {
      this.selectedRecord = this.latestData;
      this.showDetailModal = true;
    },
    openDetailModal(record) {
      this.selectedRecord = record;
      this.showDetailModal = true;
    },
    async openMappingsModal() {
      this.showMappingsModal = true;
      try {
        const token = localStorage.getItem('token');
        const [mapRes, paramRes] = await Promise.all([
          axios.get(`/api/ispu/mappings?device_id=${this.selectedDeviceId}`, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          }),
          axios.get(`/api/devices/${this.selectedDeviceId}/parameters`, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          })
        ]);
        this.stationMappings = mapRes.data?.data || [];
        this.stationParameters = paramRes.data?.data || [];
        if (this.stationParameters.length > 0) {
          this.newMapping.parameter_id = this.stationParameters[0].id;
        }
      } catch (err) {
        console.error('Failed to load mappings:', err);
      }
    },
    async saveMapping() {
      if (!this.newMapping.parameter_id || !this.newMapping.pollutant_code) return;
      try {
        const token = localStorage.getItem('token');
        await axios.post('/api/ispu/mappings', {
          device_id: this.selectedDeviceId,
          parameter_id: this.newMapping.parameter_id,
          pollutant_code: this.newMapping.pollutant_code,
          source_unit: this.newMapping.source_unit || 'µg/m³',
          conversion_factor: Number(this.newMapping.conversion_factor) || 1.0,
          min_completeness: Number(this.newMapping.min_completeness) || 75.0,
          enabled: true
        }, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        await this.openMappingsModal();
        this.triggerRecalculation();
      } catch (err) {
        alert('Failed to save mapping: ' + (err.response?.data?.error || err.message));
      }
    },
    async deleteMapping(id) {
      if (!confirm('Are you sure you want to remove this ISPU parameter mapping?')) return;
      try {
        const token = localStorage.getItem('token');
        await axios.delete(`/api/ispu/mappings/${id}`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        await this.openMappingsModal();
        this.triggerRecalculation();
      } catch (err) {
        alert('Failed to delete mapping: ' + (err.response?.data?.error || err.message));
      }
    },
    exportData(type) {
      const token = localStorage.getItem('token');
      const url = `/api/ispu/export/${type}?device_id=${this.selectedDeviceId}`;
      const a = document.createElement('a');
      a.href = url + (token ? `&token=${token}` : '');
      a.setAttribute('download', `ispu_export_${Date.now()}.${type === 'excel' ? 'csv' : 'csv'}`);
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
    }
  }
};
</script>
