<template>
  <div v-if="isAqmsDomain" class="w-full flex flex-col space-y-1.5 transition-all">
    <!-- Compact Single-Row SCADA ISPU Indicator Bar -->
    <div
      class="ispu-summary-card w-full px-2.5 py-1 rounded-lg border border-slate-200 dark:border-[#244356] bg-white dark:bg-[#0D2232] flex flex-wrap lg:flex-nowrap items-center justify-between gap-x-2 gap-y-1 transition-colors duration-150 shadow-xs"
    >
      <!-- Left: ISPU Metric, Category, Dominant Pollutant & Parameter Pill -->
      <div class="flex items-center space-x-2 sm:space-x-2.5 flex-shrink-0">
        <!-- Interactive ISPU Metric & Category with Popover -->
        <div
          class="relative ispu-popover-anchor cursor-pointer flex items-center space-x-1.5"
          @click="toggleWindowPopover"
          @mouseenter="showWindowPopover = true"
          @mouseleave="showWindowPopover = false"
          title="Click or hover to view 24-Hour ISPU Window & Regulatory Calculation Details"
        >
          <!-- Pulse Dot, Label & Main ISPU Value -->
          <div class="flex items-center space-x-1.5 flex-shrink-0">
            <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="categoryDotClass"></span>
            <span class="text-3xs font-mono font-bold tracking-wider uppercase text-slate-500 dark:text-slate-400">
              ISPU:
            </span>
            <span class="text-base sm:text-lg font-bold font-mono tabular-nums leading-none tracking-tight" :class="categoryTextClass">
              {{ displayISPU }}
            </span>
          </div>

          <!-- Category Badge -->
          <span
            class="px-1.5 py-0.5 rounded text-4xs sm:text-3xs font-mono font-bold uppercase tracking-wider border shadow-xs whitespace-nowrap flex-shrink-0"
            :class="categoryBadgeClass"
          >
            {{ ispuCategoryDisplay }}
          </span>

          <!-- 24h Window & Validity Popover -->
          <div
            v-if="showWindowPopover"
            class="absolute left-0 top-full mt-1.5 z-50 w-80 max-w-[calc(100vw-32px)] rounded-lg shadow-2xl p-3 bg-white dark:bg-[#0B1C2B] border border-slate-200 dark:border-[#244356] text-slate-800 dark:text-[#E6F3FC] font-mono text-[10px] animate-fadeIn select-none text-left"
            @click.stop
          >
            <!-- Popover Header -->
            <div class="flex items-center justify-between pb-2 mb-2 border-b border-slate-100 dark:border-[#244356]">
              <div class="flex items-center space-x-2">
                <span class="w-2.5 h-2.5 rounded-full flex-shrink-0" :class="categoryDotClass"></span>
                <div>
                  <h4 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-[#E6F3FC] leading-none">
                    24-Hour ISPU Window
                  </h4>
                  <div class="text-[9px] text-slate-500 dark:text-[#9BB4C8] mt-0.5 leading-none">
                    Permen LHK No. 14 / 2020 Standard
                  </div>
                </div>
              </div>
              <span
                class="px-1.5 py-0.5 rounded text-[8.5px] font-bold leading-none uppercase"
                :class="validityBadgeClass"
              >
                {{ calculationStatus }}
              </span>
            </div>

            <!-- Detail Rows -->
            <div class="space-y-2">
              <div class="p-2 rounded bg-slate-50 dark:bg-[#061321] border border-slate-200/80 dark:border-[#244356] space-y-1 text-[9.5px]">
                <div class="flex items-start justify-between">
                  <span class="text-slate-500 dark:text-[#9BB4C8]">Time Window:</span>
                  <span class="font-bold text-right text-slate-900 dark:text-[#E6F3FC]">{{ formattedPeriod }}</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-500 dark:text-[#9BB4C8]">Timezone:</span>
                  <span class="font-bold text-slate-700 dark:text-slate-300">WIB (UTC+7)</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-500 dark:text-[#9BB4C8]">Data Coverage:</span>
                  <span class="font-bold" :class="completenessPercent >= 75 ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-500'">
                    {{ completenessPercent.toFixed(1) }}% {{ completenessPercent >= 75 ? '(Meets ≥75% rule)' : '(Provisional <75%)' }}
                  </span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-500 dark:text-[#9BB4C8]">Dominant Pollutant:</span>
                  <span class="font-bold text-cyan-600 dark:text-[#21C7E8]">{{ dominantPollutant || '-' }} (Index: {{ dominantISPU || '-' }})</span>
                </div>
              </div>

              <!-- Regulatory Validity Status & Reason -->
              <div class="text-[9px] text-slate-600 dark:text-[#9BB4C8] leading-snug p-2 rounded bg-slate-50/70 dark:bg-[#061321] border border-slate-200/60 dark:border-[#244356]">
                <div class="font-bold text-slate-700 dark:text-slate-300 mb-0.5">Calculation Assessment:</div>
                <div>{{ validityTooltip }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Dominant Pollutant (when OVERALL) or Selected Parameter Detail Pill -->
        <div v-if="selectedDisplayParam === 'OVERALL'" class="flex items-center space-x-1 text-3xs font-mono flex-shrink-0">
          <span class="text-slate-500 dark:text-[#9BB4C8] hidden xl:inline">Dominant:</span>
          <span class="font-bold text-slate-900 dark:text-[#E6F3FC] px-1.5 py-0.5 rounded bg-slate-200/50 dark:bg-[#0D2232] border border-slate-300 dark:border-[#244356] whitespace-nowrap">
            {{ dominantPollutant || '-' }}
          </span>
        </div>
        <div v-else class="flex items-center space-x-1 text-3xs font-mono flex-shrink-0">
          <span class="font-bold text-cyan-700 dark:text-[#21C7E8] px-1.5 py-0.5 rounded bg-cyan-500/15 border border-cyan-500/30 whitespace-nowrap">
            {{ selectedParamLabel }}
          </span>
          <span class="text-4xs text-amber-500 font-mono hidden xl:inline" :title="'Polutan kritis stasiun saat ini adalah ' + dominantPollutant">
            (⭐ Kritis: {{ dominantPollutant }} [{{ dominantISPU }}])
          </span>
          <button
            @click="selectDisplayParameter('OVERALL')"
            class="text-4xs text-slate-400 hover:text-cyan-500 underline ml-1 cursor-pointer"
            title="Kembali ke nilai ISPU Dominan/Kritis"
          >
            ✕ Reset
          </button>
        </div>

        <!-- Expandable Parameter Breakdown Pill -->
        <button
          @click="toggleBreakdown"
          class="flex items-center space-x-1 px-1.5 py-0.5 rounded text-3xs font-mono font-bold transition border flex-shrink-0"
          :class="showBreakdown
            ? 'bg-cyan-500/20 text-cyan-700 dark:text-[#21C7E8] border-cyan-500/50 shadow-xs'
            : 'bg-slate-200/60 dark:bg-[#0D2232] text-slate-700 dark:text-slate-300 border-slate-300 dark:border-[#244356] hover:border-cyan-500'"
          title="Inspect All 7 Official ISPU Parameters"
        >
          <span>📊</span>
          <span class="hidden md:inline">Params:</span>
          <span class="text-cyan-600 dark:text-[#21C7E8] font-extrabold">{{ activeParametersCount }}/7</span>
          <svg class="w-3 h-3 transition-transform" :class="{ 'rotate-180': showBreakdown }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>

      <!-- Center: Data Status & Validity (Compact without long 24h period text) -->
      <div class="flex items-center space-x-2 text-3xs font-mono text-slate-500 dark:text-[#9BB4C8] min-w-0 flex-shrink">
        <!-- Validity Badge & Coverage Percentage (Click/Tap to view 24h window details) -->
        <button
          type="button"
          @click="toggleWindowPopover"
          class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold uppercase tracking-wider border whitespace-nowrap flex-shrink-0 transition hover:opacity-90 cursor-pointer"
          :class="validityBadgeClass"
          :title="'24h: ' + formattedPeriod + ' · ' + validityTooltip"
        >
          {{ calculationStatus }}
          <span v-if="completenessPercent > 0" class="ml-0.5 font-semibold">({{ completenessPercent.toFixed(0) }}%)</span>
        </button>

        <!-- Recalculated / Freshness indicator -->
        <span v-if="lastUpdatedStr" class="hidden sm:inline text-slate-400 dark:text-[#9BB4C8] text-4xs whitespace-nowrap flex-shrink-0">
          · {{ lastUpdatedStr }}
        </span>
      </div>

      <!-- Right: Action Controls -->
      <div class="flex items-center space-x-1.5 flex-shrink-0">
        <!-- Setup Mappings Button -->
        <button
          @click="openMappingsModal"
          class="px-2 py-1 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1 bg-white/80 dark:bg-[#0D2232] text-slate-600 dark:text-[#9BB4C8] border-slate-300 dark:border-[#244356] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-[#21C7E8]"
          title="Configure & Setup ISPU Parameter Mappings"
        >
          <span>⚙️</span>
          <span class="hidden xl:inline">Mappings</span>
        </button>

        <!-- Inline Trend Toggle -->
        <button
          @click="toggleTrend"
          class="px-2 py-1 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1"
          :class="showTrend
            ? 'bg-cyan-500/20 text-cyan-700 dark:text-[#21C7E8] border-cyan-500/50 shadow-xs'
            : 'bg-white/80 dark:bg-[#0D2232] text-slate-600 dark:text-[#9BB4C8] border-slate-300 dark:border-[#244356] hover:border-cyan-500'"
          title="Toggle Historical ISPU Trend"
        >
          <span>📈</span>
          <span class="hidden sm:inline">{{ showTrend ? 'Hide Trend' : 'Trend' }}</span>
        </button>

        <!-- Full Analysis Page Link -->
        <button
          @click="navigateToAnalysis"
          class="px-2.5 py-1 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1 bg-cyan-600 hover:bg-cyan-500 text-white border-cyan-400 shadow-xs"
          title="Open Dedicated ISPU Analysis & Official Breakdowns"
        >
          <span class="whitespace-nowrap">Full ISPU →</span>
        </button>
      </div>
    </div>

    <!-- Collapsible 7 Official Parameters SCADA Tray -->
    <div
      v-if="showBreakdown"
      class="p-2.5 rounded-xl bg-white dark:bg-[#0B1C2B] border border-slate-200 dark:border-[#244356] shadow-md transition-all"
    >
      <div class="flex items-center justify-between mb-2 pb-1.5 border-b border-slate-100 dark:border-slate-800 text-3xs font-mono">
        <div class="flex items-center space-x-2">
          <span class="w-2 h-2 rounded-full bg-cyan-500"></span>
          <span class="font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider">
            All 7 Official ISPU Parameters Breakdown (Permen LHK No. 14/2020)
          </span>
        </div>
        <div class="flex items-center space-x-2">
          <span class="text-slate-500 dark:text-slate-400 text-4xs">
            Dominant parameter determines overall ISPU
          </span>
          <button
            v-if="selectedDisplayParam !== 'OVERALL'"
            @click="selectDisplayParameter('OVERALL')"
            class="px-1.5 py-0.5 rounded text-4xs font-bold bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30 hover:bg-amber-500/25 transition cursor-pointer"
            title="Kembali ke nilai ISPU Dominan / Kritis Stasiun"
          >
            ★ Reset ke Dominant
          </button>
          <button
            @click="openMappingsModal"
            class="text-4xs text-cyan-600 dark:text-cyan-400 hover:underline flex items-center space-x-0.5"
          >
            <span>⚙️ Edit Mappings</span>
          </button>
        </div>
      </div>

      <!-- 7 Cards Grid -->
      <div class="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-2">
        <div
          v-for="pollutant in sevenOfficialParameters"
          :key="pollutant.code"
          @click="selectDisplayParameter(pollutant.code)"
          class="p-2 rounded-lg border transition-all flex flex-col justify-between cursor-pointer hover:border-cyan-400 select-none overflow-hidden"
          :class="[
            getCardClass(pollutant.code),
            selectedDisplayParam === pollutant.code ? 'ring-2 ring-cyan-400 shadow-md bg-cyan-500/10' : ''
          ]"
          :title="`Klik untuk menampilkan ISPU ${pollutant.code} pada bar utama`"
        >
          <!-- Card Header -->
          <div class="flex items-center justify-between text-3xs font-mono mb-1">
            <span class="font-bold text-slate-900 dark:text-white">{{ pollutant.code }}</span>
            <span
              v-if="selectedDisplayParam === pollutant.code"
              class="px-1 py-0.2 rounded text-4xs font-bold bg-cyan-500 text-slate-950 uppercase"
            >
              Active
            </span>
            <span
              v-else-if="isDominant(pollutant.code)"
              class="px-1 py-0.2 rounded text-4xs font-bold bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/40"
              title="Dominant Pollutant"
            >
              ⭐ Dominant
            </span>
            <span
              v-else-if="!getParameterResult(pollutant.code)"
              class="px-1 py-0.2 rounded text-4xs bg-slate-200 dark:bg-slate-800 text-slate-500 border border-slate-300 dark:border-slate-700"
            >
              Unmapped
            </span>
          </div>

          <!-- Metric (Fit tightly without overflowing card boundaries) -->
          <div class="flex items-center justify-between gap-1 my-0.5 min-w-0">
            <span class="text-base sm:text-lg font-bold font-mono tabular-nums leading-none shrink-0" :class="getTextClass(pollutant.code)">
              {{ getISPUValue(pollutant.code) }}
            </span>
            <span
              v-if="getParameterResult(pollutant.code)"
              class="px-1 py-0.5 rounded text-[8.5px] font-bold uppercase tracking-tight border text-center leading-none max-w-full truncate shrink"
              :class="getBadgeClass(pollutant.code)"
              :title="formatCategoryLabel(getParameterResult(pollutant.code).category)"
            >
              {{ formatCategoryLabel(getParameterResult(pollutant.code).category) }}
            </span>
          </div>

          <!-- Concentration & Formula hint -->
          <div class="text-4xs font-mono text-slate-500 dark:text-slate-400 truncate mt-1">
            <span v-if="getParameterResult(pollutant.code)">
              {{ getParameterResult(pollutant.code).concentration.toFixed(1) }} {{ getParameterResult(pollutant.code).source_unit || 'µg/m³' }}
            </span>
            <span v-else class="italic text-slate-400">
              No mapped sensor
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Collapsible Embedded Trend Chart Drawer -->
    <div
      v-if="showTrend"
      class="p-3 rounded-xl bg-white dark:bg-[#0B1C2B] border border-slate-200 dark:border-[#244356] shadow-md transition-all"
    >
      <ISPUTrendChart
        :records="trendRecords"
        :loading="trendLoading"
        :initial-range="trendRange"
        :selected-pollutant="selectedDisplayParam"
        :pollutant-name="selectedPollutantName"
        :is-dark="effectiveIsDark"
        title="ISPU 24-HOUR ROLLING TREND"
        @range-change="handleRangeChange"
        @pollutant-change="selectDisplayParameter"
      />
    </div>

    <!-- Built-in Parameter Mapping Modal (accessible directly in OCC) -->
    <div
      v-if="showMappingsModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4"
      @click.self="showMappingsModal = false"
    >
      <div class="w-full max-w-2xl bg-white dark:bg-[#0B1C2B] border border-slate-200 dark:border-[#244356] rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <!-- Header -->
        <div class="px-4 py-3 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <span class="text-lg">⚙️</span>
            <div>
              <h3 class="text-sm font-bold font-mono text-slate-900 dark:text-white">
                ISPU Parameter Mapping Setup
              </h3>
              <p class="text-4xs font-mono text-slate-500 dark:text-slate-400">
                Map physical/virtual station telemetry channels to official Permen LHK 14/2020 pollutants
              </p>
            </div>
          </div>
          <button
            @click="showMappingsModal = false"
            class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-lg leading-none"
          >
            ✕
          </button>
        </div>

        <!-- Body -->
        <div class="p-4 space-y-4 overflow-y-auto custom-scroller">
          <!-- Current Mappings Table -->
          <div>
            <h4 class="text-3xs font-mono font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 mb-2">
              Active Mappings for this Station
            </h4>
            <div class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden">
              <table class="w-full text-left text-3xs font-mono">
                <thead class="bg-slate-50 dark:bg-[#101B2B] text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800">
                  <tr>
                    <th class="p-2">Official Pollutant</th>
                    <th class="p-2">Source Sensor / Channel</th>
                    <th class="p-2">Source Unit</th>
                    <th class="p-2">Factor</th>
                    <th class="p-2">Jam Mulai</th>
                    <th class="p-2">Min Completeness</th>
                    <th class="p-2 text-right">Action</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
                  <tr v-for="m in stationMappings" :key="m.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30">
                    <td class="p-2 font-bold text-slate-900 dark:text-white">
                      {{ m.pollutant_code }}
                    </td>
                    <td class="p-2 text-cyan-600 dark:text-cyan-400">
                      {{ m.parameter ? (m.parameter.name || m.parameter.parameter_name) : `ID ${m.parameter_id}` }}
                    </td>
                    <td class="p-2 text-slate-500 dark:text-slate-400">{{ m.source_unit }}</td>
                    <td class="p-2 text-slate-500 dark:text-slate-400">{{ m.conversion_factor }}</td>
                    <td class="p-2 text-cyan-600 dark:text-cyan-400 font-bold font-mono">{{ formatStartHour(m.daily_start_hour) }}</td>
                    <td class="p-2 text-slate-500 dark:text-slate-400">{{ m.min_completeness }}%</td>
                    <td class="p-2 text-right">
                      <button
                        @click="deleteMapping(m.id)"
                        class="text-rose-500 hover:text-rose-700 font-bold px-1.5 py-0.5 rounded hover:bg-rose-500/10 transition"
                        title="Remove Mapping"
                      >
                        🗑️
                      </button>
                    </td>
                  </tr>
                  <tr v-if="stationMappings.length === 0">
                    <td colspan="7" class="p-3 text-center text-slate-400 italic">
                      No parameter mappings configured yet. Add one below.
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Add / Update Mapping Form -->
          <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-slate-800 space-y-3">
            <h4 class="text-3xs font-mono font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">
              Add or Map Parameter
            </h4>
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-2 text-3xs font-mono">
              <div>
                <label class="block text-slate-500 dark:text-slate-400 mb-1">Official Pollutant</label>
                <select
                  v-model="newMapping.pollutant_code"
                  class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-[#0B1422] text-slate-900 dark:text-white"
                >
                  <option v-for="p in sevenOfficialParameters" :key="p.code" :value="p.code">
                    {{ p.code }} ({{ p.name }})
                  </option>
                </select>
              </div>

              <div>
                <label class="block text-slate-500 dark:text-slate-400 mb-1">Source Sensor / Channel</label>
                <select
                  v-model="newMapping.parameter_id"
                  class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-[#0B1422] text-slate-900 dark:text-white"
                >
                  <option v-for="sp in stationParameters" :key="sp.id" :value="sp.id">
                    {{ sp.name || sp.parameter_name }} ({{ sp.parameter_code || sp.code }})
                  </option>
                </select>
              </div>

              <div>
                <label class="block text-slate-500 dark:text-slate-400 mb-1">Source Unit</label>
                <input
                  type="text"
                  v-model="newMapping.source_unit"
                  placeholder="µg/m³ or ppm"
                  class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-[#0B1422] text-slate-900 dark:text-white"
                />
              </div>

              <div>
                <label class="block text-slate-500 dark:text-slate-400 mb-1">Conversion Factor</label>
                <input
                  type="number"
                  step="any"
                  v-model="newMapping.conversion_factor"
                  placeholder="1.0"
                  class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-[#0B1422] text-slate-900 dark:text-white"
                />
              </div>

              <div>
                <label class="block text-slate-500 dark:text-slate-400 mb-1">Jam Mulai Harian</label>
                <select
                  v-model.number="newMapping.daily_start_hour"
                  class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-[#0B1422] text-slate-900 dark:text-white font-mono"
                  title="Jam acuan dimulainya siklus perhitungan harian 24 jam"
                >
                  <option v-for="h in 24" :key="h-1" :value="h-1">
                    {{ (h-1 < 10 ? '0' : '') + (h-1) }}:00 WIB {{ (h-1 === 9 ? '(09:00 Pagi)' : (h-1 === 15 ? '(15:00 Sore)' : (h-1 === 0 ? '(00:00 Malam)' : ''))) }}
                  </option>
                </select>
              </div>
            </div>

            <div class="flex justify-end">
              <button
                @click="saveMapping"
                class="px-3 py-1.5 rounded-lg text-3xs font-mono font-bold bg-cyan-600 hover:bg-cyan-500 text-white shadow-xs transition"
              >
                + Save Mapping
              </button>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-4 py-2.5 border-t border-slate-200 dark:border-slate-800 flex justify-end">
          <button
            @click="showMappingsModal = false"
            class="px-3 py-1.5 rounded-lg text-3xs font-mono font-semibold bg-slate-200 dark:bg-slate-800 text-slate-700 dark:text-slate-300 hover:bg-slate-300 transition"
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
import ISPUTrendChart from './ISPUTrendChart.vue';

export default {
  name: 'ISPUIndicatorBar',
  components: {
    ISPUTrendChart
  },
  props: {
    deviceId: { type: Number, default: 0 },
    isDark: { type: Boolean, default: true },
    isAqmsDomain: { type: Boolean, default: true }
  },
  data() {
    return {
      ispuData: null,
      loading: false,
      recalculating: false,
      showTrend: false,
      showBreakdown: false,
      trendRange: '24h',
      trendRecords: [],
      trendLoading: false,
      pollTimer: null,
      showWindowPopover: false,

      // Seven official parameters defined in Permen LHK 14/2020
      sevenOfficialParameters: [
        { code: 'PM25', name: 'Partikulat PM2.5 (24h)' },
        { code: 'PM10', name: 'Partikulat PM10 (24h)' },
        { code: 'SO2', name: 'Sulfur Dioksida (24h)' },
        { code: 'CO', name: 'Karbon Monoksida (24h)' },
        { code: 'O3', name: 'Ozon (24h)' },
        { code: 'NO2', name: 'Nitrogen Dioksida (24h)' },
        { code: 'HC', name: 'Hidrokarbon (24h)' }
      ],

      // Selected Display Parameter: 'OVERALL' (Default) or specific pollutant code ('PM25', 'CO', etc.)
      selectedDisplayParam: 'OVERALL',

      // Mapping Modal State
      showMappingsModal: false,
      stationMappings: [],
      stationParameters: [],
      newMapping: {
        pollutant_code: 'PM25',
        parameter_id: null,
        source_unit: 'µg/m³',
        conversion_factor: 1.0,
        min_completeness: 75.0,
        daily_start_hour: 0,
        enabled: true
      }
    };
  },
  computed: {
    effectiveIsDark() {
      if (typeof this.isDark === 'boolean') {
        return this.isDark;
      }
      const pref = this.$store && this.$store.state && this.$store.state.uiPreferences;
      if (pref && pref.effectiveTheme) {
        return pref.effectiveTheme === 'dark';
      }
      if (this.$store && this.$store.state && this.$store.state.theme) {
        return this.$store.state.theme === 'dark';
      }
      if (typeof document !== 'undefined') {
        return document.documentElement.classList.contains('dark');
      }
      return true;
    },
    selectedParamResult() {
      if (this.selectedDisplayParam === 'OVERALL') return null;
      if (!this.ispuData || !this.ispuData.parameter_results) return null;
      return this.ispuData.parameter_results.find(p => p.pollutant_code === this.selectedDisplayParam);
    },
    displayISPU() {
      if (this.selectedDisplayParam === 'OVERALL') {
        if (!this.ispuData || this.ispuData.overall_ispu === undefined) return '-';
        return Math.round(this.ispuData.overall_ispu);
      }
      const pr = this.selectedParamResult;
      if (!pr || pr.valid_samples === 0 || pr.category === 'TIDAK_TERSEDIA' || pr.calculated_ispu === undefined) {
        return '-';
      }
      return Math.round(pr.calculated_ispu);
    },
    ispuCategory() {
      if (this.selectedDisplayParam === 'OVERALL') {
        return (this.ispuData && this.ispuData.overall_category) || 'TIDAK_TERSEDIA';
      }
      const pr = this.selectedParamResult;
      return (pr && pr.category) || 'TIDAK_TERSEDIA';
    },
    ispuCategoryDisplay() {
      const cat = this.ispuCategory;
      if (!cat) return '-';
      return cat.replace(/_/g, ' ');
    },
    dominantPollutant() {
      return (this.ispuData && this.ispuData.dominant_pollutant) || '-';
    },
    dominantISPU() {
      if (!this.ispuData || this.ispuData.overall_ispu === undefined) return '-';
      return Math.round(this.ispuData.overall_ispu);
    },
    selectedParamLabel() {
      const pr = this.selectedParamResult;
      if (!pr) return `${this.selectedDisplayParam} (Unmapped)`;
      const conc = pr.concentration !== undefined ? pr.concentration.toFixed(1) : (pr.normalized_concentration !== undefined ? pr.normalized_concentration.toFixed(1) : '-');
      return `${this.selectedDisplayParam}: ${conc} ${pr.source_unit || 'µg/m³'}`;
    },
    selectedPollutantName() {
      if (!this.selectedDisplayParam || this.selectedDisplayParam === 'OVERALL') return '';
      const p = this.sevenOfficialParameters.find(item => item.code === this.selectedDisplayParam);
      return p ? p.name : '';
    },
    calculationStatus() {
      return (this.ispuData && this.ispuData.calculation_status) || 'PROVISIONAL';
    },
    completenessPercent() {
      return (this.ispuData && this.ispuData.completeness_percent) || 0;
    },
    activeParametersCount() {
      if (!this.ispuData || !this.ispuData.parameter_results) return 0;
      return this.ispuData.parameter_results.filter(p => p.calculated_ispu !== undefined).length;
    },
    formattedPeriod() {
      if (!this.ispuData || !this.ispuData.period_end) return '-';
      const end = new Date(this.ispuData.period_end);
      const start = new Date(this.ispuData.period_start || end.getTime() - 24 * 3600 * 1000);
      const fmt = d => d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });
      return `${fmt(start)} - ${fmt(end)}`;
    },
    lastUpdatedStr() {
      if (!this.ispuData || !this.ispuData.calculated_at) return '';
      const diffMs = Date.now() - new Date(this.ispuData.calculated_at).getTime();
      const mins = Math.floor(diffMs / 60000);
      if (mins < 1) return 'Just now';
      if (mins < 60) return `${mins}m ago`;
      const hrs = Math.floor(mins / 60);
      return `${hrs}h ago`;
    },
    barThemeClass() {
      // Unified with SCADA card background and border; categories and status shown via semantic badges
      return '';
    },
    categoryDotClass() {
      switch (this.ispuCategory) {
        case 'BAIK': return 'bg-emerald-500 animate-pulse';
        case 'SEDANG': return 'bg-blue-500 animate-pulse';
        case 'TIDAK_SEHAT': return 'bg-amber-500 animate-pulse';
        case 'SANGAT_TIDAK_SEHAT': return 'bg-rose-500 animate-pulse';
        case 'BERBAHAYA': return 'bg-purple-500 animate-pulse';
        default: return 'bg-slate-400';
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
    categoryBadgeClass() {
      switch (this.ispuCategory) {
        case 'BAIK':
          return 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/40';
        case 'SEDANG':
          return 'bg-blue-500/15 text-blue-700 dark:text-blue-300 border-blue-500/40';
        case 'TIDAK_SEHAT':
          return 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/40';
        case 'SANGAT_TIDAK_SEHAT':
          return 'bg-rose-500/15 text-rose-700 dark:text-rose-300 border-rose-500/40';
        case 'BERBAHAYA':
          return 'bg-purple-900/40 text-purple-300 border-purple-500/50';
        default:
          return 'bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-300 dark:border-slate-700';
      }
    },
    validityBadgeClass() {
      switch (this.calculationStatus) {
        case 'VALID':
          return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30';
        case 'PROVISIONAL':
          return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30';
        default:
          return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30';
      }
    },
    validityTooltip() {
      if (this.calculationStatus === 'VALID') {
        return 'Regulatory Valid: Kelengkapan data pemantauan ≥ 75% sesuai Permen LHK No. 14/2020 Pasal 7 ayat (3)';
      }
      if (this.calculationStatus === 'PROVISIONAL') {
        return 'Kalkulasi Provisori: Cakupan data < 75% (belum memenuhi batas penjaminan mutu penuh Permen LHK No. 14/2020 Pasal 7 ayat 3)';
      }
      return 'Kalkulasi Tidak Valid: Data telemetri tidak tersedia atau terjadi error sensor (0% coverage)';
    },
    canRecalculate() {
      return true;
    }
  },
  watch: {
    deviceId: {
      immediate: true,
      handler(newId) {
        if (newId && this.isAqmsDomain) {
          try {
            const saved = localStorage.getItem(`scada_ispu_selected_param_${newId}`);
            this.selectedDisplayParam = saved || 'OVERALL';
          } catch (e) {}
          this.fetchLatestISPU();
          if (this.showTrend) {
            this.fetchTrend();
          }
        }
      }
    }
  },
  mounted() {
    try {
      const saved = localStorage.getItem(`scada_ispu_selected_param_${this.deviceId}`);
      if (saved) {
        this.selectedDisplayParam = saved;
      }
    } catch (e) {}

    // 1. Polling fallback every 30 seconds
    this.pollTimer = setInterval(() => {
      if (this.deviceId && this.isAqmsDomain) {
        this.fetchLatestISPU();
      }
    }, 30000);

    // 2. Real-time WebSocket Listeners
    window.addEventListener('ispu-calculated-event', this.handleISPUCalculated);
    window.addEventListener('aggregation-result-event', this.handleAggregationUpdated);
    document.addEventListener('click', this.handleDocumentClick);
    document.addEventListener('keydown', this.handleDocumentKeydown);
  },
  beforeDestroy() {
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
    }
    window.removeEventListener('ispu-calculated-event', this.handleISPUCalculated);
    window.removeEventListener('aggregation-result-event', this.handleAggregationUpdated);
    document.removeEventListener('click', this.handleDocumentClick);
    document.removeEventListener('keydown', this.handleDocumentKeydown);
  },
  methods: {
    toggleWindowPopover(e) {
      if (e) e.stopPropagation();
      this.showWindowPopover = !this.showWindowPopover;
    },
    closeWindowPopover() {
      this.showWindowPopover = false;
    },
    handleDocumentClick(e) {
      if (this.showWindowPopover) {
        const anchor = this.$el && this.$el.querySelector('.ispu-popover-anchor');
        if (anchor && !anchor.contains(e.target)) {
          this.closeWindowPopover();
        }
      }
    },
    handleDocumentKeydown(e) {
      if (e.key === 'Escape' && this.showWindowPopover) {
        this.closeWindowPopover();
      }
    },
    handleISPUCalculated(event) {
      const data = event.detail;
      if (data && data.device_id === this.deviceId) {
        this.ispuData = data;
        if (this.showTrend) {
          this.fetchTrend();
        }
      }
    },
    handleAggregationUpdated(event) {
      const data = event.detail;
      if (data && data.device_id === this.deviceId) {
        // Debounced latest fetch
        if (this.aggTimeout) clearTimeout(this.aggTimeout);
        this.aggTimeout = setTimeout(() => {
          this.fetchLatestISPU();
        }, 1000);
      }
    },
    getParameterResult(pollutantCode) {
      if (!this.ispuData || !this.ispuData.parameter_results) return null;
      return this.ispuData.parameter_results.find(p => p.pollutant_code === pollutantCode);
    },
    isDominant(pollutantCode) {
      return this.dominantPollutant === pollutantCode;
    },
    getISPUValue(pollutantCode) {
      const pr = this.getParameterResult(pollutantCode);
      return pr ? Math.round(pr.calculated_ispu) : '-';
    },
    getCardClass(pollutantCode) {
      const pr = this.getParameterResult(pollutantCode);
      if (!pr) return 'bg-slate-50 dark:bg-[#101B2B]/60 border-slate-200 dark:border-slate-800 opacity-60';
      if (this.isDominant(pollutantCode)) return 'bg-amber-500/10 dark:bg-amber-950/20 border-amber-500/50 shadow-xs';
      switch (pr.category) {
        case 'BAIK': return 'bg-emerald-500/5 dark:bg-emerald-950/15 border-emerald-500/30';
        case 'SEDANG': return 'bg-blue-500/5 dark:bg-blue-950/15 border-blue-500/30';
        case 'TIDAK_SEHAT': return 'bg-amber-500/5 dark:bg-amber-950/15 border-amber-500/30';
        case 'SANGAT_TIDAK_SEHAT': return 'bg-rose-500/5 dark:bg-rose-950/15 border-rose-500/30';
        case 'BERBAHAYA': return 'bg-purple-900/20 border-purple-500/40';
        default: return 'bg-slate-50 dark:bg-[#101B2B] border-slate-200 dark:border-slate-800';
      }
    },
    getTextClass(pollutantCode) {
      const pr = this.getParameterResult(pollutantCode);
      if (!pr) return 'text-slate-400';
      switch (pr.category) {
        case 'BAIK': return 'text-emerald-600 dark:text-emerald-400';
        case 'SEDANG': return 'text-blue-600 dark:text-blue-400';
        case 'TIDAK_SEHAT': return 'text-amber-600 dark:text-amber-400';
        case 'SANGAT_TIDAK_SEHAT': return 'text-rose-600 dark:text-rose-400';
        case 'BERBAHAYA': return 'text-purple-600 dark:text-purple-400';
        default: return 'text-slate-600 dark:text-slate-400';
      }
    },
    getBadgeClass(pollutantCode) {
      const pr = this.getParameterResult(pollutantCode);
      if (!pr) return 'bg-slate-200 text-slate-500 border-slate-300';
      switch (pr.category) {
        case 'BAIK': return 'bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 border-emerald-500/40';
        case 'SEDANG': return 'bg-blue-500/20 text-blue-700 dark:text-blue-300 border-blue-500/40';
        case 'TIDAK_SEHAT': return 'bg-amber-500/20 text-amber-700 dark:text-amber-300 border-amber-500/40';
        case 'SANGAT_TIDAK_SEHAT': return 'bg-rose-500/20 text-rose-700 dark:text-rose-300 border-rose-500/40';
        case 'BERBAHAYA': return 'bg-purple-900/40 text-purple-300 border-purple-500/50';
        default: return 'bg-slate-200 text-slate-500 border-slate-300';
      }
    },
    formatCategoryLabel(cat) {
      if (!cat) return '';
      switch (cat) {
        case 'BAIK': return 'BAIK';
        case 'SEDANG': return 'SEDANG';
        case 'TIDAK_SEHAT': return 'TIDAK SEHAT';
        case 'SANGAT_TIDAK_SEHAT': return 'SGT TDK SEHAT';
        case 'BERBAHAYA': return 'BERBAHAYA';
        case 'TIDAK_TERSEDIA': return 'N/A';
        default: return cat.replace(/_/g, ' ');
      }
    },
    getAuthToken() {
      return localStorage.getItem('datalogger_token') || localStorage.getItem('token') || (this.$store?.state?.token) || '';
    },
    toggleBreakdown() {
      this.showBreakdown = !this.showBreakdown;
    },
    toggleTrend() {
      this.showTrend = !this.showTrend;
      if (this.showTrend && this.trendRecords.length === 0) {
        this.fetchTrend();
      }
    },
    async fetchLatestISPU() {
      if (!this.deviceId) return;
      this.loading = true;
      try {
        const token = this.getAuthToken();
        const res = await axios.get(`/api/ispu/latest?device_id=${this.deviceId}`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        if (res.data && res.data.success && res.data.data) {
          this.ispuData = res.data.data;
        }
      } catch (err) {
        // Silently preserve stale or previous state
      } finally {
        this.loading = false;
      }
    },
    async fetchTrend() {
      if (!this.deviceId) return;
      this.trendLoading = true;
      try {
        const token = this.getAuthToken();
        const res = await axios.get(`/api/ispu/trend?device_id=${this.deviceId}&range=${this.trendRange}`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        if (res.data && res.data.success && res.data.data) {
          this.trendRecords = res.data.data;
        }
      } catch (err) {
        // Handle error
      } finally {
        this.trendLoading = false;
      }
    },
    handleRangeChange(range) {
      this.trendRange = range;
      this.fetchTrend();
    },
    navigateToAnalysis() {
      this.$router.push({
        path: '/monitoring/ispu',
        query: { device_id: this.deviceId }
      });
    },
    async triggerRecalculation() {
      if (!this.deviceId || this.recalculating) return;
      this.recalculating = true;
      try {
        const token = this.getAuthToken();
        // Trigger both backfill and recalculate for immediate rich history
        await Promise.all([
          axios.post('/api/ispu/backfill', { device_id: this.deviceId, hours: 24 }, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          }),
          axios.post('/api/ispu/recalculate', { device_id: this.deviceId }, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          })
        ]);
        await Promise.all([this.fetchLatestISPU(), this.fetchTrend()]);
      } catch (err) {
        alert('ISPU recalculation error: ' + (err.response?.data?.error || err.message));
      } finally {
        this.recalculating = false;
      }
    },
    async openMappingsModal() {
      this.showMappingsModal = true;
      try {
        const token = this.getAuthToken();
        const [mapRes, devRes] = await Promise.all([
          axios.get(`/api/ispu/mappings?device_id=${this.deviceId}`, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          }),
          axios.get(`/api/devices`, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          })
        ]);
        this.stationMappings = mapRes.data?.data || [];
        
        // Find parameters for this device
        const devices = devRes.data?.data?.items || devRes.data?.data || [];
        const currentDev = devices.find(d => d.id === this.deviceId) || devices[0];
        this.stationParameters = currentDev?.parameters || [];
        if (this.stationParameters.length > 0 && !this.newMapping.parameter_id) {
          this.newMapping.parameter_id = this.stationParameters[0].id;
        }
      } catch (err) {
        console.error('Failed to load mappings:', err);
      }
    },
    selectDisplayParameter(code) {
      if (this.selectedDisplayParam === code && code !== 'OVERALL') {
        this.selectedDisplayParam = 'OVERALL';
      } else {
        this.selectedDisplayParam = code;
      }
      try {
        localStorage.setItem(`scada_ispu_selected_param_${this.deviceId}`, this.selectedDisplayParam);
      } catch (e) {}
    },
    formatStartHour(h) {
      const hour = Number(h) || 0;
      return `${hour < 10 ? '0' : ''}${hour}:00 WIB`;
    },
    async saveMapping() {
      if (!this.newMapping.parameter_id || !this.newMapping.pollutant_code) return;
      try {
        const token = this.getAuthToken();
        await axios.post('/api/ispu/mappings', {
          device_id: this.deviceId,
          parameter_id: this.newMapping.parameter_id,
          pollutant_code: this.newMapping.pollutant_code,
          source_unit: this.newMapping.source_unit || 'µg/m³',
          conversion_factor: Number(this.newMapping.conversion_factor) || 1.0,
          min_completeness: Number(this.newMapping.min_completeness) || 75.0,
          daily_start_hour: Number(this.newMapping.daily_start_hour) || 0,
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
        const token = this.getAuthToken();
        await axios.delete(`/api/ispu/mappings/${id}`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        await this.openMappingsModal();
        this.triggerRecalculation();
      } catch (err) {
        alert('Failed to delete mapping: ' + (err.response?.data?.error || err.message));
      }
    }
  }
};
</script>
