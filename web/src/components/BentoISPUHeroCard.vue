<template>
  <div
    v-if="isAqmsDomain"
    class="bento-hero-card"
    :class="effectiveIsDark ? 'is-night' : 'is-day'"
  >
    <!-- Top Row: Kicker + Clickable 24h Validity Chip & Popover -->
    <div class="hero-top">
      <span class="hero-kicker">ISPU · PERMEN LHK 14/2020</span>

      <div class="relative ispu-popover-anchor">
        <button
          type="button"
          class="valid-chip"
          :class="validityToneClass"
          @click="toggleWindowPopover"
          :title="'24h: ' + formattedPeriod + ' · ' + validityTooltip"
        >
          <span>{{ activeParametersCount }}/7 · {{ calculationStatus }}</span>
          <span v-if="completenessPercent > 0" class="ml-0.5">{{ Math.round(completenessPercent) }}%</span>
        </button>

        <!-- 24-Hour ISPU Window & Regulatory Calculation Popover -->
        <div
          v-if="showWindowPopover"
          class="absolute right-0 top-full mt-2 z-50 w-80 max-w-[calc(100vw-32px)] rounded-xl shadow-2xl p-3.5 bg-white dark:bg-[#111A2C] border border-slate-200 dark:border-slate-700/70 text-slate-800 dark:text-slate-100 text-[11px] select-none text-left"
          @click.stop
        >
          <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-slate-800">
            <div class="flex items-center space-x-2">
              <span class="w-2.5 h-2.5 rounded-full flex-shrink-0" :class="categoryDotClass"></span>
              <div>
                <h4 class="text-xs font-bold text-slate-900 dark:text-white leading-none">
                  Jendela ISPU 24 Jam
                </h4>
                <div class="text-[10px] text-slate-500 dark:text-slate-400 mt-0.5 leading-none">
                  Standar Permen LHK No. 14 / 2020
                </div>
              </div>
            </div>
            <button
              type="button"
              @click="closeWindowPopover"
              class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-xs px-1 cursor-pointer"
            >
              ✕
            </button>
          </div>

          <div class="space-y-2">
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0C1424] border border-slate-200/80 dark:border-slate-800 space-y-1.5 text-[10.5px]">
              <div class="flex items-start justify-between gap-2">
                <span class="text-slate-500 dark:text-slate-400">Rentang Waktu:</span>
                <span class="font-bold text-right text-slate-900 dark:text-white">{{ formattedPeriod }}</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-500 dark:text-slate-400">Zona Waktu:</span>
                <span class="font-bold text-slate-700 dark:text-slate-300">WIB (UTC+7)</span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-500 dark:text-slate-400">Cakupan Data:</span>
                <span class="font-bold" :class="completenessPercent >= 75 ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-500'">
                  {{ completenessPercent.toFixed(1) }}% {{ completenessPercent >= 75 ? '(≥75% Valid)' : '(<75% Provisori)' }}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-slate-500 dark:text-slate-400">Polutan Dominan:</span>
                <span class="font-bold text-sky-600 dark:text-sky-400">{{ formatPollutantCodePretty(dominantPollutant) }} ({{ dominantISPU }})</span>
              </div>
            </div>

            <div class="text-[10px] text-slate-600 dark:text-slate-400 leading-relaxed p-2.5 rounded-lg bg-slate-50/70 dark:bg-[#0C1424] border border-slate-200/60 dark:border-slate-800">
              <div class="font-bold text-slate-800 dark:text-slate-200 mb-0.5">Evaluasi Kalkulasi:</div>
              <div>{{ validityTooltip }}</div>
            </div>

            <div v-if="canRecalculate && effectiveDeviceId" class="pt-1 flex justify-end">
              <button
                type="button"
                @click="triggerRecalculation"
                :disabled="recalculating"
                class="px-2.5 py-1 rounded-md text-[10.5px] font-semibold bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white transition cursor-pointer"
              >
                {{ recalculating ? 'Menghitung ulang...' : '↻ Hitung Ulang ISPU' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Middle Gauge Row: Circular Ring Gauge + Dominant Pill & Status Description -->
    <div class="hero-gauge-row">
      <svg width="84" height="84" viewBox="0 0 96 96" class="flex-shrink-0 cursor-pointer" @click="toggleWindowPopover">
        <circle cx="48" cy="48" r="39" fill="none" :stroke="effectiveIsDark ? '#1B2942' : '#E9EEF5'" stroke-width="8" />
        <circle
          cx="48"
          cy="48"
          r="39"
          fill="none"
          :stroke="heroGaugeColor"
          stroke-width="8"
          stroke-dasharray="245"
          :stroke-dashoffset="heroGaugeDashOffset"
          stroke-linecap="round"
          transform="rotate(-90 48 48)"
          style="transition: stroke-dashoffset 0.35s ease, stroke 0.25s ease;"
        />
        <text
          x="48"
          y="49"
          text-anchor="middle"
          :fill="effectiveIsDark ? '#F8FAFC' : '#0F172A'"
          :font-size="String(displayISPU).length >= 3 ? '22' : '26'"
          font-weight="800"
          font-family="'Plus Jakarta Sans', 'Inter', sans-serif"
        >{{ displayISPU }}</text>
        <text
          x="48"
          y="63"
          text-anchor="middle"
          :fill="effectiveIsDark ? '#94A3B8' : '#64748B'"
          font-size="9"
          font-weight="700"
          font-family="'Plus Jakarta Sans', 'Inter', sans-serif"
        >{{ bentoShortCategory }}</text>
      </svg>

      <div class="min-w-0 flex-1">
        <div v-if="selectedDisplayParam === 'OVERALL'" class="flex items-center flex-wrap gap-1 mb-1">
          <span class="hero-status-pill" :class="heroStatusPillTone">
            Dominan: {{ formatPollutantCodePretty(dominantPollutant) }}{{ dominantConcentrationSuffix }}
          </span>
        </div>
        <div v-else class="flex items-center flex-wrap gap-1.5 mb-1">
          <span class="hero-status-pill tone-blue">
            {{ selectedParamLabel }}
          </span>
          <button
            type="button"
            @click="selectDisplayParameter('OVERALL')"
            class="text-[10px] font-bold text-slate-400 hover:text-sky-500 underline cursor-pointer"
            title="Kembali ke nilai ISPU Dominan Stasiun"
          >
            Reset
          </button>
        </div>

        <p class="hero-desc">
          {{ heroHumanDescription }}
          <span v-if="lastUpdatedStr">
            Diperbarui <strong>{{ lastUpdatedStr }}</strong>.
          </span>
        </p>
      </div>
    </div>

    <!-- All 7 Official ISPU Parameters Breakdown List -->
    <div class="mini-ispu-list">
      <div
        v-for="item in orderedBentoPollutants"
        :key="item.code"
        @click="selectDisplayParameter(item.code)"
        class="mini-ispu-row"
        :class="{ 'row-selected': selectedDisplayParam === item.code }"
        :title="getPollutantRowTitle(item)"
      >
        <span
          class="mini-ispu-lbl truncate"
          :class="{ 'is-dom': isDominant(item.code) || selectedDisplayParam === item.code }"
        >
          {{ item.pretty }}{{ isDominant(item.code) ? ' · Dom' : '' }}
        </span>
        <div class="mini-ispu-track">
          <div
            class="mini-ispu-fill"
            :style="{ width: getBarWidthPct(item.code) + '%', background: getBarFillColor(item.code) }"
          ></div>
        </div>
        <span class="mini-ispu-num">{{ getISPUValue(item.code) }}</span>
      </div>
    </div>

    <!-- Bottom Action Buttons: Mappings, Trend ISPU, Full ISPU -->
    <div class="hero-actions">
      <button
        type="button"
        class="h-act-btn"
        @click="openMappingsModal"
        title="Konfigurasi Pemetaan Parameter ISPU"
      >
        <span>Mappings</span>
      </button>

      <button
        type="button"
        class="h-act-btn"
        :class="{ 'is-active': showTrend }"
        @click="toggleTrend"
        title="Lihat Grafik Tren ISPU 24 Jam"
      >
        <span>Trend ISPU</span>
      </button>

      <button
        type="button"
        class="h-act-btn primary"
        @click="navigateToAnalysis"
        title="Buka Halaman Analisis ISPU Lengkap"
      >
        <span>Full ISPU</span>
        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.3" stroke-linecap="round" stroke-linejoin="round">
          <polyline points="9 18 15 12 9 6" />
        </svg>
      </button>
    </div>

    <!-- Modal: 24-Hour Rolling ISPU Trend Chart (when Trend ISPU is clicked) -->
    <div
      v-if="showTrend"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 dark:bg-black/75 backdrop-blur-xs p-3 sm:p-5"
      @click.self="showTrend = false"
    >
      <div class="w-full max-w-4xl rounded-2xl bg-white dark:bg-[#111A2C] border border-slate-200 dark:border-slate-700/70 shadow-2xl p-4 sm:p-5">
        <div class="flex items-center justify-between pb-3 mb-3 border-b border-slate-200 dark:border-slate-800">
          <div>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white">
              Tren Bergulir ISPU 24 Jam
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              Standar Permen LHK No. 14 Tahun 2020
            </p>
          </div>
          <button
            type="button"
            @click="showTrend = false"
            class="px-2.5 py-1 rounded-lg border border-slate-200 dark:border-slate-700 text-xs font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 cursor-pointer"
          >
            Tutup ✕
          </button>
        </div>
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
    </div>

    <!-- Built-in Parameter Mapping Modal (100% preserved from ISPUIndicatorBar) -->
    <div
      v-if="showMappingsModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xs p-3 sm:p-4 animate-fadeIn"
      @click.self="showMappingsModal = false"
    >
      <div class="w-full max-w-4xl bg-white dark:bg-[#0B1C2B] border border-slate-200 dark:border-[#244356] rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[92vh]">
        <div class="px-5 py-3.5 border-b border-slate-200 dark:border-[#1E3A52] bg-slate-50 dark:bg-[#081521] flex items-center justify-between gap-3 flex-shrink-0">
          <div class="flex items-center space-x-3 min-w-0">
            <div class="w-8 h-8 rounded-lg bg-cyan-500/10 dark:bg-cyan-500/20 border border-cyan-500/30 flex items-center justify-center text-cyan-600 dark:text-[#21C7E8] flex-shrink-0">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </div>
            <div class="min-w-0">
              <div class="flex items-center space-x-2">
                <h3 class="text-sm font-bold text-slate-900 dark:text-[#E6F3FC] tracking-wide">
                  ISPU Parameter Mapping Setup
                </h3>
                <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-cyan-500/10 text-cyan-600 dark:text-[#21C7E8] border border-cyan-500/30">
                  Permen LHK 14/2020
                </span>
              </div>
              <p class="text-xs text-slate-500 dark:text-[#9BB4C8] truncate">
                Petakan kanal sensor stasiun ke 7 polutan resmi Permen LHK No. 14/2020
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2 flex-shrink-0">
            <span class="text-xs px-2 py-0.5 rounded bg-slate-200/60 dark:bg-[#102436] text-slate-600 dark:text-[#9BB4C8] border border-slate-300 dark:border-[#244356]">
              {{ uniqueMappedPollutantsCount }}/7 Mapped
            </span>
            <button
              type="button"
              @click="showMappingsModal = false"
              class="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-white hover:bg-slate-200/60 dark:hover:bg-[#1A3348] transition cursor-pointer"
              title="Close (Esc)"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <div class="p-4 sm:p-5 space-y-4 overflow-y-auto custom-scroller flex-1">
          <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#07131F] border border-slate-200 dark:border-[#1E3A52]">
            <div class="flex items-center justify-between mb-2">
              <span class="text-[11px] font-bold uppercase tracking-wider text-slate-500 dark:text-[#9BB4C8]">
                Standar 7 Polutan Permen LHK No. 14 / 2020
              </span>
              <span class="text-[11px]" :class="uniqueMappedPollutantsCount === 7 ? 'text-emerald-500 font-bold' : 'text-amber-500'">
                {{ uniqueMappedPollutantsCount === 7 ? '✓ Semua 7 Polutan Telah Dipetakan' : `${7 - uniqueMappedPollutantsCount} Belum Dikonfigurasi` }}
              </span>
            </div>
            <div class="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-7 gap-1.5">
              <button
                v-for="p in sevenOfficialParameters"
                :key="p.code"
                type="button"
                @click="selectPollutantToMap(p.code)"
                class="px-2 py-1.5 rounded-lg border text-left flex flex-col justify-between transition cursor-pointer"
                :class="isPollutantMapped(p.code)
                  ? 'bg-emerald-500/10 dark:bg-emerald-950/20 border-emerald-500/40 text-emerald-700 dark:text-emerald-300 hover:border-emerald-400'
                  : 'bg-slate-100/70 dark:bg-[#0E2336]/40 border-slate-300 dark:border-[#244356] text-slate-500 dark:text-slate-400 hover:border-cyan-500'"
              >
                <div class="flex items-center justify-between">
                  <span class="text-xs font-bold">{{ p.code }}</span>
                  <span class="text-xs">{{ isPollutantMapped(p.code) ? '✓' : '—' }}</span>
                </div>
                <span class="text-[9.5px] truncate text-slate-400">{{ p.name.split(' ')[0] }}</span>
              </button>
            </div>
          </div>

          <div>
            <div class="flex items-center justify-between mb-2">
              <h4 class="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">
                Active ISPU Parameter Mappings ({{ stationMappings.length }})
              </h4>
              <span class="text-[10.5px] text-slate-400">
                Klik baris untuk mengedit mapping
              </span>
            </div>

            <div class="border border-slate-200 dark:border-[#1E3A52] rounded-xl overflow-hidden bg-white dark:bg-[#07131F] shadow-xs">
              <div class="overflow-x-auto custom-scroller">
                <table class="w-full text-left text-xs">
                  <thead class="bg-slate-100/80 dark:bg-[#091827] text-slate-500 dark:text-[#9BB4C8] border-b border-slate-200 dark:border-[#1E3A52] text-[10px] uppercase tracking-wider select-none">
                    <tr>
                      <th class="py-2.5 px-3">Official Pollutant</th>
                      <th class="py-2.5 px-3">Source Sensor / Channel</th>
                      <th class="py-2.5 px-2.5 text-center">Unit</th>
                      <th class="py-2.5 px-2.5 text-center">Factor</th>
                      <th class="py-2.5 px-3 text-center">Jam Mulai</th>
                      <th class="py-2.5 px-2.5 text-center">Min Data</th>
                      <th class="py-2.5 px-3 text-right">Action</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-[#12283A]">
                    <tr
                      v-for="m in stationMappings"
                      :key="m.id"
                      class="transition-colors cursor-pointer"
                      :class="editingMappingId === m.id
                        ? 'bg-amber-500/10 dark:bg-amber-950/20'
                        : 'hover:bg-slate-50/80 dark:hover:bg-[#0E2336]/60'"
                      @click="editMapping(m)"
                    >
                      <td class="py-2 px-3 whitespace-nowrap">
                        <div class="flex items-center space-x-1.5">
                          <span class="px-2 py-0.5 rounded text-xs font-bold bg-cyan-500/10 text-cyan-700 dark:text-[#21C7E8] border border-cyan-500/30">
                            {{ m.pollutant_code }}
                          </span>
                          <span
                            v-if="isDuplicatePollutant(m.pollutant_code, m.id)"
                            class="px-1.5 py-0.5 rounded text-[9px] font-bold bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/40"
                          >
                            Duplikat
                          </span>
                        </div>
                        <div class="text-[10px] text-slate-400 mt-0.5 truncate max-w-[150px]">
                          {{ getPollutantName(m.pollutant_code) }}
                        </div>
                      </td>
                      <td class="py-2 px-3">
                        <div class="font-bold text-slate-800 dark:text-[#E6F3FC] truncate max-w-[200px]">
                          {{ getMappingChannelName(m) }}
                        </div>
                        <div v-if="!isOperator && getMappingDeviceName(m)" class="text-[10px] text-slate-500 dark:text-[#9BB4C8] mt-0.5">
                          Device: {{ getMappingDeviceName(m) }}
                        </div>
                      </td>
                      <td class="py-2 px-2.5 text-center whitespace-nowrap">
                        <span class="px-1.5 py-0.5 rounded text-[10.5px] bg-slate-100 dark:bg-[#0E2336] text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-[#1E3A52]">
                          {{ m.source_unit || 'µg/m³' }}
                        </span>
                      </td>
                      <td class="py-2 px-2.5 text-center whitespace-nowrap text-slate-600 dark:text-slate-300 tabular-nums">
                        {{ m.conversion_factor }}
                      </td>
                      <td class="py-2 px-3 text-center whitespace-nowrap">
                        <span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10.5px] font-bold text-cyan-700 dark:text-[#21C7E8] bg-cyan-500/10 border border-cyan-500/20">
                          {{ formatStartHour(m.daily_start_hour) }}
                        </span>
                      </td>
                      <td class="py-2 px-2.5 text-center whitespace-nowrap text-slate-600 dark:text-slate-300 tabular-nums">
                        {{ m.min_completeness || 75 }}%
                      </td>
                      <td class="py-2 px-3 text-right whitespace-nowrap" @click.stop>
                        <div class="flex items-center justify-end space-x-1">
                          <button
                            type="button"
                            @click="editMapping(m)"
                            class="p-1 rounded text-slate-400 hover:text-cyan-600 dark:hover:text-[#21C7E8] hover:bg-cyan-500/10 transition cursor-pointer"
                            title="Edit Mapping"
                          >
                            ✎
                          </button>
                          <button
                            type="button"
                            @click="deleteMapping(m.id)"
                            class="p-1 rounded text-rose-500 hover:text-rose-700 dark:hover:text-rose-400 hover:bg-rose-500/10 transition cursor-pointer"
                            title="Hapus Mapping"
                          >
                            ✕
                          </button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="stationMappings.length === 0">
                      <td colspan="7" class="py-8 text-center text-slate-400 italic">
                        Belum ada parameter mapping yang dikonfigurasi untuk stasiun ini.
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <div id="ispu-mapping-form" class="p-4 rounded-xl bg-slate-50 dark:bg-[#0D2232] border border-slate-200 dark:border-[#1E3A52] space-y-3.5 shadow-xs">
            <div class="flex items-center justify-between">
              <h4 class="text-xs font-bold uppercase tracking-wider" :class="editingMappingId ? 'text-amber-500 dark:text-amber-400' : 'text-slate-700 dark:text-slate-300'">
                {{ editingMappingId ? `Edit Mapping: ${newMapping.pollutant_code}` : 'Tambah / Konfigurasi Mapping Polutan' }}
              </h4>
              <button
                v-if="editingMappingId"
                type="button"
                @click="cancelEdit"
                class="text-xs text-slate-400 hover:text-rose-400 underline cursor-pointer"
              >
                ✕ Batal Edit
              </button>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-2.5 text-xs">
              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">Official Pollutant *</label>
                <select
                  v-model="newMapping.pollutant_code"
                  @change="onPollutantChange"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC]"
                >
                  <option v-for="p in sevenOfficialParameters" :key="p.code" :value="p.code">
                    {{ p.code }} ({{ p.name }})
                  </option>
                </select>
              </div>

              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">Source Sensor / Channel *</label>
                <select
                  v-model="newMapping.parameter_id"
                  @change="onParameterChange"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC]"
                >
                  <option v-if="stationParameters.length === 0" :value="null" disabled>
                    -- Tidak ada sensor channel --
                  </option>
                  <option v-for="sp in stationParameters" :key="sp.id" :value="sp.id">
                    {{ sp.displayName || (sp.name || sp.parameter_name) }}
                  </option>
                </select>
              </div>

              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">Source Unit</label>
                <input
                  type="text"
                  v-model="newMapping.source_unit"
                  placeholder="µg/m³"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC]"
                />
              </div>

              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">Conversion Factor</label>
                <input
                  type="number"
                  step="any"
                  v-model="newMapping.conversion_factor"
                  placeholder="1.0"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC]"
                />
              </div>

              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">Jam Mulai Harian</label>
                <select
                  v-model.number="newMapping.daily_start_hour"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC]"
                >
                  <option v-for="h in 24" :key="h-1" :value="h-1">
                    {{ (h-1 < 10 ? '0' : '') + (h-1) }}:00 WIB
                  </option>
                </select>
              </div>
            </div>

            <div class="flex items-center justify-between pt-1">
              <div class="text-[11px] text-slate-500 dark:text-[#9BB4C8]">
                Formula otomatis menormalkan konsentrasi ke baku mutu Permen LHK 14/2020
              </div>
              <div class="flex items-center space-x-2">
                <button
                  v-if="editingMappingId"
                  type="button"
                  @click="cancelEdit"
                  class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-200 dark:bg-[#1E3A52] text-slate-700 dark:text-slate-300 cursor-pointer"
                >
                  Batal
                </button>
                <button
                  type="button"
                  @click="saveMapping"
                  :disabled="savingMapping || !newMapping.parameter_id"
                  class="px-4 py-1.5 rounded-lg text-xs font-bold bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white shadow-xs transition cursor-pointer"
                >
                  {{ savingMapping ? 'Menyimpan...' : (editingMappingId ? 'Perbarui Mapping' : 'Simpan Mapping') }}
                </button>
              </div>
            </div>
          </div>
        </div>

        <div class="px-5 py-3 border-t border-slate-200 dark:border-[#1E3A52] bg-slate-50 dark:bg-[#081521] flex items-center justify-between flex-shrink-0">
          <div class="text-xs text-slate-500 dark:text-[#9BB4C8]">
            Perhitungan ISPU 24 jam diperbarui otomatis oleh background engine stasiun
          </div>
          <button
            type="button"
            @click="showMappingsModal = false"
            class="px-4 py-1.5 rounded-lg text-xs font-semibold bg-slate-200 dark:bg-[#102436] text-slate-700 dark:text-slate-300 hover:bg-slate-300 dark:hover:bg-[#1A3348] transition cursor-pointer"
          >
            Tutup
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import ISPUIndicatorBar from './ISPUIndicatorBar.vue';

export default {
  name: 'BentoISPUHeroCard',
  extends: ISPUIndicatorBar,
  computed: {
    orderedBentoPollutants() {
      return [
        { code: 'PM10', pretty: 'PM10' },
        { code: 'PM25', pretty: 'PM2.5' },
        { code: 'O3', pretty: 'O₃' },
        { code: 'SO2', pretty: 'SO₂' },
        { code: 'CO', pretty: 'CO' },
        { code: 'NO2', pretty: 'NO₂' },
        { code: 'HC', pretty: 'HC' },
      ];
    },
    bentoShortCategory() {
      const cat = this.ispuCategory;
      switch (cat) {
        case 'BAIK': return 'BAIK';
        case 'SEDANG': return 'SEDANG';
        case 'TIDAK_SEHAT': return 'TDK SEHAT';
        case 'SANGAT_TIDAK_SEHAT': return 'SGT TDK SEHAT';
        case 'BERBAHAYA': return 'BERBAHAYA';
        default: return 'BELUM ADA';
      }
    },
    heroGaugeColor() {
      const dark = this.effectiveIsDark;
      switch (this.ispuCategory) {
        case 'BAIK': return dark ? '#34D399' : '#059669';
        case 'SEDANG': return dark ? '#38BDF8' : '#0284C7';
        case 'TIDAK_SEHAT': return dark ? '#FBBF24' : '#D97706';
        case 'SANGAT_TIDAK_SEHAT': return '#F43F5E';
        case 'BERBAHAYA': return dark ? '#A78BFA' : '#7C3AED';
        default: return dark ? '#475569' : '#94A3B8';
      }
    },
    heroGaugeDashOffset() {
      const circumference = 245;
      const num = Number(this.displayISPU);
      if (!isFinite(num) || num <= 0) {
        return circumference;
      }
      // Smoothly map 0..200 ISPU onto ring (clamped at 96% so rounded cap remains crisp)
      const ratio = Math.min(Math.max(num / 150, 0.06), 0.96);
      return Math.round(circumference * (1 - ratio));
    },
    validityToneClass() {
      if (this.calculationStatus === 'VALID') return 'tone-green';
      if (this.calculationStatus === 'PROVISIONAL') return 'tone-amber';
      return 'tone-slate';
    },
    heroStatusPillTone() {
      switch (this.ispuCategory) {
        case 'BAIK': return 'tone-green';
        case 'SEDANG': return 'tone-blue';
        case 'TIDAK_SEHAT': return 'tone-amber';
        case 'SANGAT_TIDAK_SEHAT':
        case 'BERBAHAYA':
          return 'tone-rose';
        default:
          return 'tone-slate';
      }
    },
    dominantConcentrationSuffix() {
      if (!this.dominantPollutant || this.dominantPollutant === '-') return '';
      const pr = this.getParameterResult(this.dominantPollutant);
      if (!pr || pr.concentration === undefined || pr.concentration === null) return '';
      const val = Number(pr.concentration);
      if (!isFinite(val)) return '';
      return ` (${val.toFixed(1)} ${pr.source_unit || 'µg/m³'})`;
    },
    heroHumanDescription() {
      switch (this.ispuCategory) {
        case 'BAIK':
          return 'Mutu udara sangat baik & aman untuk seluruh aktivitas.';
        case 'SEDANG':
          return 'Udara aman untuk operasional stasiun & aktivitas umum.';
        case 'TIDAK_SEHAT':
          return 'Waspada peningkatan konsentrasi polutan udara ambien.';
        case 'SANGAT_TIDAK_SEHAT':
          return 'Kualitas udara sangat tidak sehat, batasi paparan luar.';
        case 'BERBAHAYA':
          return 'Kondisi udara berbahaya, aktifkan prosedur tanggap darurat.';
        default:
          return 'Menunggu kecukupan sampel sensor untuk kalkulasi ISPU.';
      }
    },
  },
  methods: {
    formatPollutantCodePretty(code) {
      if (!code || code === '-') return '-';
      const map = {
        PM25: 'PM2.5',
        PM10: 'PM10',
        SO2: 'SO₂',
        NO2: 'NO₂',
        O3: 'O₃',
        CO: 'CO',
        HC: 'HC',
      };
      return map[code] || code;
    },
    getBarWidthPct(code) {
      const pr = this.getParameterResult(code);
      if (!pr || pr.calculated_ispu === undefined || pr.category === 'TIDAK_TERSEDIA') return 0;
      const val = Number(pr.calculated_ispu);
      if (!isFinite(val) || val <= 0) return 0;
      return Math.min(Math.max(Math.round((val / 150) * 100), 5), 100);
    },
    getBarFillColor(code) {
      const dark = this.effectiveIsDark;
      const pr = this.getParameterResult(code);
      if (!pr || pr.category === 'TIDAK_TERSEDIA') {
        return dark ? '#334155' : '#CBD5E1';
      }
      if (this.isDominant(code) && pr.category === 'BAIK') {
        return dark ? '#2DD4BF' : '#0D9488';
      }
      switch (pr.category) {
        case 'BAIK': return dark ? '#34D399' : '#059669';
        case 'SEDANG': return dark ? '#38BDF8' : '#0284C7';
        case 'TIDAK_SEHAT': return dark ? '#FBBF24' : '#D97706';
        case 'SANGAT_TIDAK_SEHAT': return '#F43F5E';
        case 'BERBAHAYA': return dark ? '#A78BFA' : '#7C3AED';
        default: return dark ? '#38BDF8' : '#0284C7';
      }
    },
    getPollutantRowTitle(item) {
      const pr = this.getParameterResult(item.code);
      if (!pr) {
        return `${item.pretty}: Belum dipetakan ke sensor. Klik Mappings untuk mengatur.`;
      }
      const conc = pr.concentration != null ? `${Number(pr.concentration).toFixed(2)} ${pr.source_unit || 'µg/m³'}` : '-';
      return `Klik untuk menampilkan ISPU ${item.pretty} · Konsentrasi 24j: ${conc}`;
    },
  },
};
</script>

<style scoped>
.bento-hero-card {
  border-radius: 14px;
  padding: 13px 15px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 10px;
  height: 100%;
  min-height: 332px;
  font-family: 'Plus Jakarta Sans', 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  font-variant-numeric: tabular-nums;
  transition: background 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.bento-hero-card.is-day {
  background: linear-gradient(160deg, #FFFFFF 0%, #F4FAFF 100%);
  border: 1px solid #E2E8F0;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03), 0 6px 18px rgba(15, 23, 42, 0.03);
  color: #0F172A;
}

.bento-hero-card.is-night {
  background: linear-gradient(160deg, #132038 0%, #0E1829 100%);
  border: 1px solid rgba(148, 163, 184, 0.14);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.28);
  color: #F8FAFC;
}

.hero-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.hero-kicker {
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.015em;
  white-space: nowrap;
}
.is-day .hero-kicker { color: #334155; }
.is-night .hero-kicker { color: #CBD5E1; }

.valid-chip {
  font-size: 10px;
  font-weight: 700;
  padding: 2.5px 7.5px;
  border-radius: 6px;
  white-space: nowrap;
  cursor: pointer;
  border: 1px solid transparent;
  transition: opacity 0.15s ease;
}
.valid-chip:hover { opacity: 0.9; }

.is-day .valid-chip.tone-green { background: #DCFCE7; color: #15803D; border-color: #BBF7D0; }
.is-night .valid-chip.tone-green { background: rgba(16, 185, 129, 0.15); color: #6EE7B7; border-color: rgba(16, 185, 129, 0.28); }

.is-day .valid-chip.tone-amber { background: #FEF3C7; color: #B45309; border-color: #FDE68A; }
.is-night .valid-chip.tone-amber { background: rgba(245, 158, 11, 0.16); color: #FCD34D; border-color: rgba(245, 158, 11, 0.28); }

.is-day .valid-chip.tone-slate { background: #F1F5F9; color: #475569; border-color: #E2E8F0; }
.is-night .valid-chip.tone-slate { background: rgba(148, 163, 184, 0.14); color: #CBD5E1; border-color: rgba(148, 163, 184, 0.22); }

.hero-gauge-row {
  display: flex;
  align-items: center;
  gap: 13px;
}

.hero-status-pill {
  display: inline-block;
  padding: 3px 9px;
  border-radius: 999px;
  font-size: 11px;
  font-weight: 700;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.is-day .hero-status-pill.tone-blue { background: #E0F2FE; color: #0369A1; }
.is-night .hero-status-pill.tone-blue { background: rgba(56, 189, 248, 0.16); color: #7DD3FC; }

.is-day .hero-status-pill.tone-green { background: #DCFCE7; color: #15803D; }
.is-night .hero-status-pill.tone-green { background: rgba(52, 211, 153, 0.16); color: #6EE7B7; }

.is-day .hero-status-pill.tone-amber { background: #FEF3C7; color: #B45309; }
.is-night .hero-status-pill.tone-amber { background: rgba(251, 191, 36, 0.16); color: #FCD34D; }

.is-day .hero-status-pill.tone-rose { background: #FFE4E6; color: #BE123C; }
.is-night .hero-status-pill.tone-rose { background: rgba(244, 63, 94, 0.18); color: #FDA4AF; }

.is-day .hero-status-pill.tone-slate { background: #F1F5F9; color: #475569; }
.is-night .hero-status-pill.tone-slate { background: rgba(148, 163, 184, 0.14); color: #CBD5E1; }

.hero-desc {
  font-size: 11.5px;
  line-height: 1.38;
  margin: 0;
}
.is-day .hero-desc { color: #64748B; }
.is-night .hero-desc { color: #94A3B8; }
.is-day .hero-desc strong { color: #334155; }
.is-night .hero-desc strong { color: #E2E8F0; }

.mini-ispu-list {
  display: flex;
  flex-direction: column;
  gap: 3.5px;
  padding: 6px 0;
}
.is-day .mini-ispu-list {
  border-top: 1px solid #EDF2F7;
  border-bottom: 1px solid #EDF2F7;
}
.is-night .mini-ispu-list {
  border-top: 1px solid rgba(148, 163, 184, 0.09);
  border-bottom: 1px solid rgba(148, 163, 184, 0.09);
}

.mini-ispu-row {
  display: grid;
  grid-template-columns: 68px 1fr 26px;
  align-items: center;
  gap: 8px;
  font-size: 10.5px;
  padding: 1.5px 4px;
  border-radius: 5px;
  cursor: pointer;
  transition: background 0.15s ease;
}
.is-day .mini-ispu-row:hover,
.is-day .mini-ispu-row.row-selected {
  background: rgba(2, 132, 199, 0.06);
}
.is-night .mini-ispu-row:hover,
.is-night .mini-ispu-row.row-selected {
  background: rgba(56, 189, 248, 0.09);
}

.mini-ispu-lbl {
  font-weight: 600;
}
.is-day .mini-ispu-lbl { color: #64748B; }
.is-night .mini-ispu-lbl { color: #94A3B8; }
.is-day .mini-ispu-lbl.is-dom { color: #0284C7; font-weight: 700; }
.is-night .mini-ispu-lbl.is-dom { color: #38BDF8; font-weight: 700; }

.mini-ispu-track {
  height: 5px;
  border-radius: 999px;
  overflow: hidden;
}
.is-day .mini-ispu-track { background: #E9EEF5; }
.is-night .mini-ispu-track { background: #1B2942; }

.mini-ispu-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s ease;
}

.mini-ispu-num {
  text-align: right;
  font-weight: 700;
}
.is-day .mini-ispu-num { color: #0F172A; }
.is-night .mini-ispu-num { color: #F8FAFC; }

.hero-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.h-act-btn {
  flex: 1;
  height: 27px;
  border-radius: 7px;
  font-size: 11px;
  font-weight: 600;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.is-day .h-act-btn {
  background: #FFFFFF;
  border: 1px solid #E2E8F0;
  color: #334155;
}
.is-day .h-act-btn:hover,
.is-day .h-act-btn.is-active {
  border-color: #0284C7;
  color: #0284C7;
}
.is-night .h-act-btn {
  background: #111A2C;
  border: 1px solid rgba(148, 163, 184, 0.16);
  color: #CBD5E1;
}
.is-night .h-act-btn:hover,
.is-night .h-act-btn.is-active {
  border-color: #38BDF8;
  color: #38BDF8;
}

.h-act-btn.primary {
  background: #0284C7 !important;
  border-color: #0284C7 !important;
  color: #FFFFFF !important;
  font-weight: 700;
}
.h-act-btn.primary:hover {
  background: #0369A1 !important;
}
</style>

