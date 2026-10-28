<template>
  <div v-if="isAqmsDomain" class="w-full flex flex-col space-y-1.5 transition-all">
    <!-- Compact Single-Row SCADA ISPU Indicator Bar -->
    <div
      class="ispu-summary-card w-full px-2.5 py-1 rounded-lg border border-[#1E3A8A] dark:border-[#244356] bg-white dark:bg-[#0D2232] flex flex-wrap lg:flex-nowrap items-center justify-between gap-x-2 gap-y-1 transition-colors duration-150 shadow-xs"
    >
      <!-- Left: ISPU Metric, Category, Dominant Pollutant, Parameter Pill & Validity -->
      <div class="flex items-center space-x-2 sm:space-x-2.5 flex-wrap gap-y-1 flex-shrink-0">
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
          class="flex items-center space-x-1 px-1.5 py-0.5 rounded text-3xs font-mono font-bold transition border flex-shrink-0 cursor-pointer"
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

        <!-- Status Validitas ISPU (PROVISIONAL/VALID) & Freshness: Berdampingan di Samping Kanan Params 14/7 -->
        <div class="flex items-center space-x-1.5 text-3xs font-mono text-slate-500 dark:text-[#9BB4C8] flex-shrink-0">
          <button
            type="button"
            @click="toggleWindowPopover"
            class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold uppercase tracking-wider border whitespace-nowrap flex-shrink-0 transition hover:opacity-90 cursor-pointer shadow-2xs"
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
      </div>

      <!-- Right: Action Controls -->
      <div class="flex items-center space-x-1.5 flex-shrink-0">
        <!-- Setup Mappings Button -->
        <button
          @click="openMappingsModal"
          class="px-2 py-1 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1 bg-white/80 dark:bg-[#0D2232] text-slate-600 dark:text-[#9BB4C8] border-slate-300 dark:border-[#244356] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-[#21C7E8] cursor-pointer"
          title="Configure & Setup ISPU Parameter Mappings"
        >
          <span>⚙️</span>
          <span class="hidden xl:inline">Mappings</span>
        </button>

        <!-- Inline Trend Toggle -->
        <button
          @click="toggleTrend"
          class="px-2 py-1 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1 cursor-pointer"
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
          class="px-2.5 py-1 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1 bg-cyan-600 hover:bg-cyan-500 text-white border-cyan-400 shadow-xs cursor-pointer"
          title="Open Dedicated ISPU Analysis & Official Breakdowns"
        >
          <span class="whitespace-nowrap">Full ISPU →</span>
        </button>
      </div>
    </div>

    <!-- Collapsible 7 Official Parameters SCADA Tray -->
    <div
      v-if="showBreakdown"
      class="ispu-breakdown-tray p-2.5 rounded-xl bg-white dark:bg-[#0B1C2B] border border-[#1E3A8A] dark:border-[#244356] shadow-md transition-all"
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
          class="ispu-breakdown-tile p-2 sm:p-2.5 rounded-lg border transition-all flex flex-col justify-between cursor-pointer hover:scale-[1.01] select-none overflow-hidden"
          :class="[
            getCardClass(pollutant.code),
            selectedDisplayParam === pollutant.code ? '!border-cyan-500/50 dark:!border-cyan-400/40' : ''
          ]"
          :title="`Klik untuk menampilkan ISPU ${pollutant.code} pada bar utama`"
        >
          <!-- Card Header -->
          <div class="flex items-center justify-between text-3xs font-mono mb-1.5">
            <span class="font-extrabold text-xs tracking-tight text-slate-900 dark:text-white">{{ pollutant.code }}</span>
            <span
              v-if="selectedDisplayParam === pollutant.code"
              class="px-1.5 py-0.5 rounded text-4xs font-extrabold bg-cyan-600 text-white uppercase shadow-2xs"
            >
              Active
            </span>
            <span
              v-else-if="isDominant(pollutant.code)"
              class="px-1.5 py-0.5 rounded text-4xs font-extrabold bg-amber-500 text-white dark:bg-amber-400 dark:text-slate-950 border border-amber-600 dark:border-amber-300 shadow-2xs"
              title="Dominant Pollutant"
            >
              ⭐ Dominant
            </span>
            <span
              v-else-if="!getParameterResult(pollutant.code)"
              class="px-1 py-0.5 rounded text-4xs font-semibold bg-slate-200 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-300 dark:border-slate-700"
            >
              Unmapped
            </span>
          </div>

          <!-- Metric (Fit tightly without overflowing card boundaries) -->
          <div class="flex items-baseline justify-between gap-1 my-1 min-w-0">
            <span class="text-xl sm:text-2xl font-black font-mono tabular-nums leading-none shrink-0" :class="getTextClass(pollutant.code)">
              {{ getISPUValue(pollutant.code) }}
            </span>
            <span
              v-if="getParameterResult(pollutant.code)"
              class="px-1.5 py-0.5 rounded text-[9px] font-black uppercase tracking-wider border text-center leading-none max-w-full truncate shrink shadow-2xs"
              :class="getBadgeClass(pollutant.code)"
              :title="formatCategoryLabel(getParameterResult(pollutant.code).category)"
            >
              {{ formatCategoryLabel(getParameterResult(pollutant.code).category) }}
            </span>
          </div>

          <!-- Concentration & Formula hint -->
          <div class="text-4xs font-mono font-bold text-slate-700 dark:text-slate-300 truncate mt-1">
            <span v-if="getParameterResult(pollutant.code)">
              {{ getParameterResult(pollutant.code).concentration.toFixed(1) }} {{ getParameterResult(pollutant.code).source_unit || 'µg/m³' }}
            </span>
            <span v-else class="italic text-slate-400 dark:text-slate-500">
              No mapped sensor
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Collapsible Embedded Trend Chart Drawer -->
    <div
      v-if="showTrend"
      class="ispu-trend-drawer p-3 rounded-xl bg-white dark:bg-[#0B1C2B] border border-[#1E3A8A] dark:border-[#244356] shadow-md transition-all"
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
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-xs p-3 sm:p-4 animate-fadeIn"
      @click.self="showMappingsModal = false"
    >
      <div class="w-full max-w-4xl bg-white dark:bg-[#0B1C2B] border border-slate-200 dark:border-[#244356] rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[92vh]">
        <!-- Header -->
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
                <h3 class="text-sm font-bold font-mono text-slate-900 dark:text-[#E6F3FC] tracking-wide">
                  ISPU Parameter Mapping Setup
                </h3>
                <span class="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-cyan-500/10 text-cyan-600 dark:text-[#21C7E8] border border-cyan-500/30">
                  Permen LHK 14/2020
                </span>
              </div>
              <p class="text-3xs font-mono text-slate-500 dark:text-[#9BB4C8] truncate">
                Map physical/virtual station telemetry channels to official Permen LHK 14/2020 pollutants
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2 flex-shrink-0">
            <span class="text-3xs font-mono px-2 py-0.5 rounded bg-slate-200/60 dark:bg-[#102436] text-slate-600 dark:text-[#9BB4C8] border border-slate-300 dark:border-[#244356]">
              {{ uniqueMappedPollutantsCount }}/7 Mapped
            </span>
            <button
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

        <!-- Body -->
        <div class="p-4 sm:p-5 space-y-4 overflow-y-auto custom-scroller flex-1">
          <!-- 7 Official Permen LHK 14/2020 Pollutants Checklist / Quick Navigator -->
          <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#07131F] border border-slate-200 dark:border-[#1E3A52]">
            <div class="flex items-center justify-between mb-2">
              <span class="text-4xs font-mono font-bold uppercase tracking-wider text-slate-500 dark:text-[#9BB4C8] flex items-center space-x-1.5">
                <span>📋</span>
                <span>Standar 7 Polutan Permen LHK No. 14 / 2020</span>
              </span>
              <span class="text-4xs font-mono" :class="uniqueMappedPollutantsCount === 7 ? 'text-emerald-500 font-bold' : 'text-amber-500'">
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
                :title="isPollutantMapped(p.code) ? `${p.code} sudah dipetakan. Klik untuk edit.` : `${p.code} belum dipetakan. Klik untuk petakan sensor.`"
              >
                <div class="flex items-center justify-between">
                  <span class="text-xs font-mono font-bold">{{ p.code }}</span>
                  <span class="text-3xs">{{ isPollutantMapped(p.code) ? '✓' : '—' }}</span>
                </div>
                <span class="text-[9px] truncate font-sans text-slate-400 dark:text-slate-400">{{ p.name.split(' ')[0] }}</span>
              </button>
            </div>
          </div>

          <!-- Current Mappings Table Section -->
          <div>
            <div class="flex items-center justify-between mb-2">
              <h4 class="text-3xs font-mono font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center space-x-1.5">
                <span class="w-2 h-2 rounded-full bg-cyan-500"></span>
                <span>Active ISPU Parameter Mappings ({{ stationMappings.length }})</span>
              </h4>
              <span class="text-4xs font-mono text-slate-400">
                Klik baris atau tombol ✏️ untuk mengedit mapping
              </span>
            </div>

            <div class="border border-slate-200 dark:border-[#1E3A52] rounded-xl overflow-hidden bg-white dark:bg-[#07131F] shadow-xs">
              <div class="overflow-x-auto custom-scroller">
                <table class="w-full text-left text-3xs font-mono">
                  <thead class="bg-slate-100/80 dark:bg-[#091827] text-slate-500 dark:text-[#9BB4C8] border-b border-slate-200 dark:border-[#1E3A52] text-4xs uppercase tracking-wider select-none">
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
                      class="transition-colors cursor-pointer group"
                      :class="editingMappingId === m.id
                        ? 'bg-amber-500/10 dark:bg-amber-950/20'
                        : 'hover:bg-slate-50/80 dark:hover:bg-[#0E2336]/60'"
                      @click="editMapping(m)"
                    >
                      <!-- Official Pollutant -->
                      <td class="py-2 px-3 whitespace-nowrap">
                        <div class="flex items-center space-x-1.5">
                          <span class="px-2 py-0.5 rounded text-xs font-mono font-bold bg-cyan-500/10 text-cyan-700 dark:text-[#21C7E8] border border-cyan-500/30">
                            {{ m.pollutant_code }}
                          </span>
                          <span
                            v-if="isDuplicatePollutant(m.pollutant_code, m.id)"
                            class="px-1.5 py-0.5 rounded text-[8.5px] font-bold bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/40"
                            title="Terdapat lebih dari satu kanal untuk polutan ini"
                          >
                            Duplikat
                          </span>
                        </div>
                        <div class="text-[9.5px] text-slate-400 dark:text-slate-400 mt-0.5 font-sans truncate max-w-[150px]">
                          {{ getPollutantName(m.pollutant_code) }}
                        </div>
                      </td>

                      <!-- Source Sensor / Channel -->
                      <td class="py-2 px-3">
                        <div class="flex items-center space-x-1.5 font-bold text-slate-800 dark:text-[#E6F3FC]">
                          <span class="text-cyan-500 text-3xs">📡</span>
                          <span class="truncate max-w-[200px]" :title="getMappingChannelName(m)">
                            {{ getMappingChannelName(m) }}
                          </span>
                        </div>
                        <div v-if="!isOperator && getMappingDeviceName(m)" class="text-[9.5px] text-slate-500 dark:text-[#9BB4C8] font-mono mt-0.5">
                          Device: {{ getMappingDeviceName(m) }}
                        </div>
                      </td>

                      <!-- Source Unit -->
                      <td class="py-2 px-2.5 text-center whitespace-nowrap">
                        <span class="px-1.5 py-0.5 rounded text-4xs font-mono bg-slate-100 dark:bg-[#0E2336] text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-[#1E3A52]">
                          {{ m.source_unit || 'µg/m³' }}
                        </span>
                      </td>

                      <!-- Factor -->
                      <td class="py-2 px-2.5 text-center whitespace-nowrap text-slate-600 dark:text-slate-300 tabular-nums">
                        {{ m.conversion_factor }}
                      </td>

                      <!-- Jam Mulai -->
                      <td class="py-2 px-3 text-center whitespace-nowrap">
                        <span class="inline-flex items-center space-x-1 px-1.5 py-0.5 rounded text-4xs font-mono font-bold text-cyan-700 dark:text-[#21C7E8] bg-cyan-500/10 border border-cyan-500/20">
                          <span>🕒</span>
                          <span>{{ formatStartHour(m.daily_start_hour) }}</span>
                        </span>
                      </td>

                      <!-- Min Completeness -->
                      <td class="py-2 px-2.5 text-center whitespace-nowrap text-slate-600 dark:text-slate-300 tabular-nums">
                        {{ m.min_completeness || 75 }}%
                      </td>

                      <!-- Actions -->
                      <td class="py-2 px-3 text-right whitespace-nowrap" @click.stop>
                        <div class="flex items-center justify-end space-x-1">
                          <button
                            type="button"
                            @click="editMapping(m)"
                            class="p-1 rounded text-slate-400 hover:text-cyan-600 dark:hover:text-[#21C7E8] hover:bg-cyan-500/10 transition cursor-pointer"
                            title="Edit Mapping ini"
                          >
                            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                            </svg>
                          </button>
                          <button
                            type="button"
                            @click="deleteMapping(m.id)"
                            class="p-1 rounded text-rose-500 hover:text-rose-700 dark:hover:text-rose-400 hover:bg-rose-500/10 transition cursor-pointer"
                            title="Hapus Mapping ini"
                          >
                            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                          </button>
                        </div>
                      </td>
                    </tr>
                    <tr v-if="stationMappings.length === 0">
                      <td colspan="7" class="py-8 text-center text-slate-400 italic">
                        Belum ada parameter mapping yang dikonfigurasi untuk stasiun ini. Tambahkan parameter di bawah.
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <!-- Add / Update Mapping Form (Modernized Card) -->
          <div id="ispu-mapping-form" class="p-4 rounded-xl bg-slate-50 dark:bg-[#0D2232] border border-slate-200 dark:border-[#1E3A52] space-y-3.5 shadow-xs transition-all">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-2">
                <span class="w-2 h-2 rounded-full" :class="editingMappingId ? 'bg-amber-400 animate-pulse' : 'bg-cyan-500'"></span>
                <h4 class="text-3xs font-mono font-bold uppercase tracking-wider" :class="editingMappingId ? 'text-amber-500 dark:text-amber-400' : 'text-slate-700 dark:text-slate-300'">
                  {{ editingMappingId ? `Edit Mapping: ${newMapping.pollutant_code}` : 'Tambah / Konfigurasi Mapping Polutan' }}
                </h4>
              </div>
              <div class="flex items-center space-x-2">
                <span class="text-4xs text-slate-400 font-mono">
                  Sensor Channels: {{ stationParameters.length }} available
                </span>
                <button
                  v-if="editingMappingId"
                  type="button"
                  @click="cancelEdit"
                  class="text-4xs font-mono text-slate-400 hover:text-rose-400 underline cursor-pointer"
                >
                  ✕ Batal Edit
                </button>
              </div>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-2.5 text-3xs font-mono">
              <!-- Official Pollutant -->
              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">
                  Official Pollutant <span class="text-rose-500">*</span>
                </label>
                <select
                  v-model="newMapping.pollutant_code"
                  @change="onPollutantChange"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC] focus:ring-1 focus:ring-cyan-500 focus:border-cyan-500"
                >
                  <option v-for="p in sevenOfficialParameters" :key="p.code" :value="p.code">
                    {{ p.code }} ({{ p.name }})
                  </option>
                </select>
              </div>

              <!-- Source Sensor / Channel -->
              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">
                  Source Sensor / Channel <span class="text-rose-500">*</span>
                </label>
                <select
                  v-model="newMapping.parameter_id"
                  @change="onParameterChange"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC] focus:ring-1 focus:ring-cyan-500 focus:border-cyan-500"
                >
                  <option v-if="stationParameters.length === 0" :value="null" disabled>
                    -- Tidak ada sensor channel ditemukan --
                  </option>
                  <option v-for="sp in stationParameters" :key="sp.id" :value="sp.id">
                    {{ sp.displayName || (sp.name || sp.parameter_name) }}
                  </option>
                </select>
              </div>

              <!-- Source Unit -->
              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">
                  Source Unit
                </label>
                <input
                  type="text"
                  v-model="newMapping.source_unit"
                  placeholder="µg/m³ or ppm"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC] focus:ring-1 focus:ring-cyan-500 focus:border-cyan-500"
                />
              </div>

              <!-- Conversion Factor -->
              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">
                  Conversion Factor
                </label>
                <input
                  type="number"
                  step="any"
                  v-model="newMapping.conversion_factor"
                  placeholder="1.0"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC] focus:ring-1 focus:ring-cyan-500 focus:border-cyan-500"
                />
              </div>

              <!-- Jam Mulai Harian -->
              <div>
                <label class="block text-slate-500 dark:text-[#9BB4C8] mb-1 font-bold">
                  Jam Mulai Harian
                </label>
                <select
                  v-model.number="newMapping.daily_start_hour"
                  class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-[#244356] bg-white dark:bg-[#081521] text-slate-900 dark:text-[#E6F3FC] font-mono focus:ring-1 focus:ring-cyan-500 focus:border-cyan-500"
                  title="Jam acuan dimulainya siklus perhitungan harian 24 jam"
                >
                  <option v-for="h in 24" :key="h-1" :value="h-1">
                    {{ (h-1 < 10 ? '0' : '') + (h-1) }}:00 WIB {{ (h-1 === 9 ? '(09:00 Pagi)' : (h-1 === 15 ? '(15:00 Sore)' : (h-1 === 0 ? '(00:00 Malam)' : ''))) }}
                  </option>
                </select>
              </div>
            </div>

            <!-- Submit Button & State -->
            <div class="flex items-center justify-between pt-1">
              <div class="text-4xs text-slate-500 dark:text-[#9BB4C8] font-mono">
                Formula otomatis menormalkan konsentrasi sensor ke skala baku mutu Permen LHK 14/2020
              </div>
              <div class="flex items-center space-x-2">
                <button
                  v-if="editingMappingId"
                  type="button"
                  @click="cancelEdit"
                  class="px-3 py-1.5 rounded-lg text-3xs font-mono font-semibold bg-slate-200 dark:bg-[#1E3A52] text-slate-700 dark:text-slate-300 hover:bg-slate-300 transition cursor-pointer"
                >
                  Batal
                </button>
                <button
                  type="button"
                  @click="saveMapping"
                  :disabled="savingMapping || !newMapping.parameter_id"
                  class="px-4 py-1.5 rounded-lg text-3xs font-mono font-bold bg-cyan-600 hover:bg-cyan-500 disabled:opacity-50 text-white shadow-xs transition flex items-center space-x-1.5 cursor-pointer"
                >
                  <span v-if="savingMapping" class="animate-spin text-3xs">⏳</span>
                  <span v-else>{{ editingMappingId ? '💾' : '➕' }}</span>
                  <span>{{ editingMappingId ? 'Perbarui Mapping' : 'Simpan Mapping' }}</span>
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-[#1E3A52] bg-slate-50 dark:bg-[#081521] flex items-center justify-between flex-shrink-0">
          <div class="text-4xs font-mono text-slate-500 dark:text-[#9BB4C8] flex items-center space-x-1.5">
            <span class="text-cyan-500">ℹ️</span>
            <span>Perhitungan ISPU 24-jam di-update otomatis oleh background engine stasiun</span>
          </div>
          <button
            @click="showMappingsModal = false"
            class="px-4 py-1.5 rounded-lg text-3xs font-mono font-semibold bg-slate-200 dark:bg-[#102436] text-slate-700 dark:text-slate-300 hover:bg-slate-300 dark:hover:bg-[#1A3348] transition cursor-pointer"
          >
            Tutup
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
    devicesList: { type: Array, default: () => [] },
    isDark: { type: Boolean, default: true },
    isAqmsDomain: { type: Boolean, default: true }
  },
  data() {
    return {
      selectedDeviceId: (() => {
        try {
          const saved = localStorage.getItem('scada_ispu_selected_device_id');
          if (saved) return Number(saved);
        } catch (e) {}
        return null;
      })(),
      internalDevicesList: [],
      ispuData: null,
      loading: false,
      recalculating: false,
      showTrend: false,
      showBreakdown: (() => {
        try {
          const savedDev = localStorage.getItem('scada_ispu_selected_device_id');
          const devId = savedDev ? Number(savedDev) : this.deviceId;
          if (devId) {
            const devSaved = localStorage.getItem(`scada_ispu_show_breakdown_${devId}`);
            if (devSaved !== null) return devSaved === 'true';
          }
          const saved = localStorage.getItem('scada_ispu_show_breakdown');
          if (saved !== null) return saved === 'true';
        } catch (e) {}
        return false;
      })(),
      trendRange: '24h',
      trendRecords: [],
      trendLoading: false,
      pollTimer: null,
      aggTimeout: null,
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
      editingMappingId: null,
      savingMapping: false,
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
    isOperator() {
      if (this.$route && this.$route.path && this.$route.path.includes('/operator')) {
        return true;
      }
      return (this.$store && this.$store.getters && this.$store.getters.isOperator) || false;
    },
    uniqueMappedPollutantsCount() {
      const set = new Set();
      (this.stationMappings || []).forEach(m => {
        if (m.pollutant_code && m.enabled !== false) set.add(m.pollutant_code);
      });
      return set.size;
    },
    availableDevices() {
      if (this.devicesList && this.devicesList.length > 0) {
        return this.devicesList;
      }
      return this.internalDevicesList || [];
    },
    effectiveDeviceId() {
      if (this.selectedDeviceId) return Number(this.selectedDeviceId);
      if (this.deviceId) return Number(this.deviceId);
      return 0;
    },
    currentDeviceName() {
      const dev = this.availableDevices.find(d => Number(d.id) === Number(this.effectiveDeviceId));
      if (!dev) return '';
      const code = dev.device_code || dev.code || '';
      const name = dev.device_name || dev.name || '';
      if (code && name) return `${code} - ${name}`;
      return code || name || `Device #${this.effectiveDeviceId}`;
    },
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
      const validCodes = new Set();
      (this.ispuData.parameter_results || []).forEach(p => {
        if (p.calculated_ispu !== undefined && p.category && p.category !== 'TIDAK_TERSEDIA') {
          const code = (p.pollutant_code || '').toUpperCase();
          if (this.sevenOfficialParameters.some(sp => sp.code === code)) {
            validCodes.add(code);
          }
        }
      });
      return validCodes.size;
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
        if (newId && (!this.selectedDeviceId || this.selectedDeviceId === 0)) {
          this.selectedDeviceId = newId;
        }
      }
    },
    effectiveDeviceId: {
      immediate: true,
      handler(newId) {
        if (newId && this.isAqmsDomain) {
          try {
            const saved = localStorage.getItem(`scada_ispu_selected_param_${newId}`);
            this.selectedDisplayParam = saved || 'OVERALL';
            const devSaved = localStorage.getItem(`scada_ispu_show_breakdown_${newId}`);
            if (devSaved !== null) {
              this.showBreakdown = devSaved === 'true';
            }
          } catch (e) {}
          this.fetchLatestISPU();
          if (this.showTrend) {
            this.fetchTrend();
          }
        }
      }
    },
    showBreakdown(newVal) {
      try {
        const valStr = newVal ? 'true' : 'false';
        localStorage.setItem('scada_ispu_show_breakdown', valStr);
        if (this.effectiveDeviceId) {
          localStorage.setItem(`scada_ispu_show_breakdown_${this.effectiveDeviceId}`, valStr);
        }
      } catch (e) {}
    },
    'newMapping.parameter_id': function(newId) {
      if (!newId || !this.stationParameters) return;
      const p = this.stationParameters.find(item => item.id === newId);
      if (p && p.unit && (!this.newMapping.source_unit || this.newMapping.source_unit === 'µg/m³')) {
        this.newMapping.source_unit = p.unit;
      }
    }
  },
  mounted() {
    try {
      const savedDev = localStorage.getItem('scada_ispu_selected_device_id');
      if (savedDev) {
        this.selectedDeviceId = Number(savedDev);
      }
      const saved = localStorage.getItem(`scada_ispu_selected_param_${this.effectiveDeviceId}`);
      if (saved) {
        this.selectedDisplayParam = saved;
      }
      if (this.effectiveDeviceId) {
        const devSaved = localStorage.getItem(`scada_ispu_show_breakdown_${this.effectiveDeviceId}`);
        if (devSaved !== null) {
          this.showBreakdown = devSaved === 'true';
        } else {
          const globalSaved = localStorage.getItem('scada_ispu_show_breakdown');
          if (globalSaved !== null) {
            this.showBreakdown = globalSaved === 'true';
          }
        }
      }
    } catch (e) {}

    // Fetch devices list if empty
    this.fetchDevicesList();

    // 1. Polling fallback every 30 seconds
    this.pollTimer = setInterval(() => {
      if (this.effectiveDeviceId && this.isAqmsDomain) {
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
      if (data && (!this.effectiveDeviceId || Number(data.device_id) === Number(this.effectiveDeviceId) || Number(data.device_id) === 0)) {
        this.ispuData = data;
        if (this.showTrend) {
          this.fetchTrend();
        }
      }
    },
    handleAggregationUpdated(event) {
      const data = event.detail;
      if (data && (!this.effectiveDeviceId || Number(data.device_id) === Number(this.effectiveDeviceId) || Number(data.device_id) === 0)) {
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
      if (!pr) {
        return 'bg-slate-100/60 dark:bg-[#101B2B] border-slate-200/80 dark:border-slate-800 opacity-60';
      }
      if (this.isDominant(pollutantCode)) {
        return 'bg-amber-500/10 dark:bg-amber-950/25 border-amber-500/35 dark:border-amber-400/30';
      }
      switch (pr.category) {
        case 'BAIK':
          return 'bg-emerald-500/10 dark:bg-emerald-950/20 border-emerald-500/25 dark:border-emerald-500/20';
        case 'SEDANG':
          return 'bg-blue-500/10 dark:bg-blue-950/20 border-blue-500/25 dark:border-blue-500/20';
        case 'TIDAK_SEHAT':
          return 'bg-amber-500/10 dark:bg-amber-950/20 border-amber-500/25 dark:border-amber-500/20';
        case 'SANGAT_TIDAK_SEHAT':
          return 'bg-rose-500/10 dark:bg-rose-950/20 border-rose-500/25 dark:border-rose-500/20';
        case 'BERBAHAYA':
          return 'bg-purple-500/10 dark:bg-purple-950/20 border-purple-500/25 dark:border-purple-400/20';
        default:
          return 'bg-slate-100/60 dark:bg-[#101B2B] border-slate-200 dark:border-[#20354B]';
      }
    },
    getTextClass(pollutantCode) {
      const pr = this.getParameterResult(pollutantCode);
      if (!pr) return 'text-slate-400 dark:text-slate-500';
      switch (pr.category) {
        case 'BAIK': return 'text-emerald-700 dark:text-emerald-300';
        case 'SEDANG': return 'text-blue-700 dark:text-blue-300';
        case 'TIDAK_SEHAT': return 'text-amber-800 dark:text-amber-300';
        case 'SANGAT_TIDAK_SEHAT': return 'text-rose-700 dark:text-rose-300';
        case 'BERBAHAYA': return 'text-purple-800 dark:text-purple-300';
        default: return 'text-slate-700 dark:text-slate-300';
      }
    },
    getBadgeClass(pollutantCode) {
      const pr = this.getParameterResult(pollutantCode);
      if (!pr) return 'bg-slate-200 text-slate-600 border border-slate-300';
      switch (pr.category) {
        case 'BAIK':
          return 'bg-emerald-600 text-white dark:bg-emerald-500 dark:text-slate-950 border border-emerald-700 dark:border-emerald-400';
        case 'SEDANG':
          return 'bg-blue-600 text-white dark:bg-blue-500 dark:text-slate-950 border border-blue-700 dark:border-blue-400';
        case 'TIDAK_SEHAT':
          return 'bg-amber-500 text-slate-950 dark:bg-amber-400 dark:text-slate-950 border border-amber-600 dark:border-amber-300 font-black';
        case 'SANGAT_TIDAK_SEHAT':
          return 'bg-rose-600 text-white dark:bg-rose-500 dark:text-white border border-rose-700 dark:border-rose-400';
        case 'BERBAHAYA':
          return 'bg-purple-800 text-white dark:bg-purple-500 dark:text-slate-950 border border-purple-900 dark:border-purple-300';
        default:
          return 'bg-slate-200 text-slate-600 border border-slate-300';
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
    handleDeviceChange(newId) {
      const id = Number(newId);
      if (!id || id === this.selectedDeviceId) return;
      this.selectedDeviceId = id;
      try {
        localStorage.setItem('scada_ispu_selected_device_id', String(id));
        const savedParam = localStorage.getItem(`scada_ispu_selected_param_${id}`);
        this.selectedDisplayParam = savedParam || 'OVERALL';
        const devSaved = localStorage.getItem(`scada_ispu_show_breakdown_${id}`);
        if (devSaved !== null) {
          this.showBreakdown = devSaved === 'true';
        }
      } catch (e) {}
      this.$emit('device-change', id);
      this.fetchLatestISPU();
      if (this.showTrend) {
        this.fetchTrend();
      }
      if (this.showMappingsModal) {
        this.loadMappingsForDevice(id);
      }
    },
    async fetchDevicesList() {
      try {
        const token = this.getAuthToken();
        const res = await axios.get('/api/devices', {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        const items = res.data?.data?.items || res.data?.data || [];
        this.internalDevicesList = Array.isArray(items) ? items : [];
        if (!this.selectedDeviceId && this.internalDevicesList.length > 0) {
          this.selectedDeviceId = this.internalDevicesList[0].id;
        }
      } catch (err) {
        // silent ignore
      }
    },
    toggleBreakdown() {
      this.showBreakdown = !this.showBreakdown;
      try {
        const valStr = this.showBreakdown ? 'true' : 'false';
        localStorage.setItem('scada_ispu_show_breakdown', valStr);
        if (this.effectiveDeviceId) {
          localStorage.setItem(`scada_ispu_show_breakdown_${this.effectiveDeviceId}`, valStr);
        }
      } catch (e) {}
    },
    toggleTrend() {
      this.showTrend = !this.showTrend;
      if (this.showTrend && this.trendRecords.length === 0) {
        this.fetchTrend();
      }
    },
    async fetchLatestISPU() {
      this.loading = true;
      try {
        const token = this.getAuthToken();
        const url = this.effectiveDeviceId ? `/api/ispu/latest?device_id=${this.effectiveDeviceId}` : '/api/ispu/latest';
        const res = await axios.get(url, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        if (res.data && res.data.success && res.data.data) {
          this.ispuData = res.data.data;
        } else {
          this.ispuData = null;
        }
      } catch (err) {
        // Silently preserve stale or previous state
      } finally {
        this.loading = false;
      }
    },
    async fetchTrend() {
      this.trendLoading = true;
      try {
        const token = this.getAuthToken();
        const url = this.effectiveDeviceId
          ? `/api/ispu/trend?device_id=${this.effectiveDeviceId}&range=${this.trendRange}`
          : `/api/ispu/trend?range=${this.trendRange}`;
        const res = await axios.get(url, {
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
        query: { device_id: this.effectiveDeviceId }
      });
    },
    async triggerRecalculation() {
      if (!this.effectiveDeviceId || this.recalculating) return;
      this.recalculating = true;
      try {
        const token = this.getAuthToken();
        // Trigger both backfill and recalculate for immediate rich history
        await Promise.all([
          axios.post('/api/ispu/backfill', { device_id: this.effectiveDeviceId, hours: 24 }, {
            headers: token ? { Authorization: `Bearer ${token}` } : {}
          }),
          axios.post('/api/ispu/recalculate', { device_id: this.effectiveDeviceId }, {
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
      this.resetNewMapping();
      this.showMappingsModal = true;
      await this.loadAllMappings();
    },
    async loadAllMappings() {
      try {
        const token = this.getAuthToken();
        const headers = token ? { Authorization: `Bearer ${token}` } : {};

        // 1. Fetch devices (prefer availableDevices from props, fallback to API)
        let devs = (this.availableDevices && this.availableDevices.length > 0) ? this.availableDevices : [];
        if (devs.length === 0) {
          try {
            const devRes = await axios.get('/api/devices', { headers, params: { page_size: 100 } });
            devs = devRes.data?.data?.items || devRes.data?.data || [];
            this.internalDevicesList = devs;
          } catch (e) {
            console.error('Failed to load devices list:', e);
          }
        }

        // Active devices filter: enabled !== false and not explicitly DISABLED / INACTIVE
        const isDevActive = (d) => {
          if (!d) return false;
          if (d.enabled === false) return false;
          const st = (d.status || d.administrative_status || '').toUpperCase();
          if (st === 'DISABLED' || st === 'INACTIVE') return false;
          return true;
        };
        let activeDevs = (devs || []).filter(isDevActive);
        if (activeDevs.length === 0 && devs.length > 0) {
          activeDevs = devs.filter(d => d.enabled !== false);
        }
        if (activeDevs.length === 0 && devs.length > 0) {
          activeDevs = devs;
        }

        // 2. Fetch parameters for all active devices
        const allParams = [];
        await Promise.all(activeDevs.map(async (d) => {
          try {
            const pRes = await axios.get(`/api/devices/${d.id}/parameters`, { headers });
            const pList = pRes.data?.data?.items || pRes.data?.data || [];
            pList.forEach(p => {
              if (p.enabled !== false) {
                const devCode = d.device_code || d.code || `DEV-${d.id}`;
                const paramName = p.parameter_name || p.name || p.parameter_code || p.code || `Param ${p.id}`;
                const paramCode = p.parameter_code || p.code || '';
                const unitVal = p.unit || p.source_unit || p.parameter_unit || 'µg/m³';
                const baseDisplayName = `${paramName}${paramCode && paramCode !== paramName ? ' (' + paramCode + ')' : ''}`;
                allParams.push({
                  ...p,
                  deviceId: d.id,
                  device_id: d.id,
                  deviceCode: devCode,
                  deviceName: d.device_name || d.name,
                  unit: unitVal,
                  displayName: (this.isOperator || !devCode) ? baseDisplayName : `[${devCode}] ${baseDisplayName}`
                });
              }
            });
          } catch (e) {
            console.error('Failed to load parameters for device', d.id, e);
          }
        }));
        this.stationParameters = allParams;
        if (this.stationParameters.length > 0) {
          if (!this.newMapping.parameter_id || !this.stationParameters.some(p => p.id === this.newMapping.parameter_id)) {
            this.newMapping.parameter_id = this.stationParameters[0].id;
            this.onParameterChange();
          }
        } else {
          this.newMapping.parameter_id = null;
        }

        // 3. Fetch mappings: try unified endpoint first, fallback to per-device if empty/error
        let mappings = [];
        try {
          const mapRes = await axios.get('/api/ispu/mappings', { headers });
          if (mapRes.data?.success && Array.isArray(mapRes.data?.data)) {
            mappings = mapRes.data.data;
          }
        } catch (e) {
          // If unified endpoint throws error (e.g. backend expecting device_id), fall through to per-device
        }

        // Fallback: if mappings are empty, query per active device
        if (mappings.length === 0 && activeDevs.length > 0) {
          const perDevResults = await Promise.all(activeDevs.map(async (d) => {
            try {
              const res = await axios.get(`/api/ispu/mappings?device_id=${d.id}`, { headers });
              return res.data?.data || [];
            } catch (err) {
              return [];
            }
          }));
          const seen = new Set();
          for (const list of perDevResults) {
            for (const m of list) {
              if (!seen.has(m.id)) {
                seen.add(m.id);
                mappings.push(m);
              }
            }
          }
        }

        // Ensure each mapping has device information attached
        mappings.forEach(m => {
          if (!m.device && m.device_id) {
            const d = activeDevs.find(dev => Number(dev.id) === Number(m.device_id));
            if (d) m.device = d;
          }
        });

        this.stationMappings = mappings;
      } catch (err) {
        console.error('Failed to load ISPU mappings:', err);
      }
    },
    selectDisplayParameter(code) {
      if (this.selectedDisplayParam === code && code !== 'OVERALL') {
        this.selectedDisplayParam = 'OVERALL';
      } else {
        this.selectedDisplayParam = code;
      }
      try {
        localStorage.setItem(`scada_ispu_selected_param`, this.selectedDisplayParam);
      } catch (e) {}
    },
    formatStartHour(h) {
      const hour = Number(h) || 0;
      return `${hour < 10 ? '0' : ''}${hour}:00 WIB`;
    },
    selectPollutantToMap(code) {
      this.newMapping.pollutant_code = code;
      this.onPollutantChange();
      const existing = (this.stationMappings || []).find(m => m.pollutant_code === code);
      if (existing) {
        this.editMapping(existing);
      } else {
        this.editingMappingId = null;
      }
    },
    onPollutantChange() {
      const unitMap = {
        'PM25': 'µg/m³',
        'PM10': 'µg/m³',
        'SO2': 'µg/m³',
        'CO': 'µg/m³',
        'O3': 'µg/m³',
        'NO2': 'µg/m³',
        'HC': 'µg/m³'
      };
      if (!this.newMapping.source_unit || this.newMapping.source_unit === 'µg/m³') {
        this.newMapping.source_unit = unitMap[this.newMapping.pollutant_code] || 'µg/m³';
      }
    },
    onParameterChange() {
      const selected = this.stationParameters.find(p => p.id === this.newMapping.parameter_id);
      if (selected && selected.unit) {
        this.newMapping.source_unit = selected.unit;
      }
    },
    editMapping(m) {
      this.editingMappingId = m.id;
      this.newMapping = {
        pollutant_code: m.pollutant_code,
        parameter_id: m.parameter_id,
        source_unit: m.source_unit || 'µg/m³',
        conversion_factor: m.conversion_factor !== undefined ? m.conversion_factor : 1.0,
        min_completeness: m.min_completeness !== undefined ? m.min_completeness : 75.0,
        daily_start_hour: m.daily_start_hour !== undefined ? m.daily_start_hour : 0,
        enabled: m.enabled !== false
      };
      this.$nextTick(() => {
        const formEl = document.getElementById('ispu-mapping-form');
        if (formEl) formEl.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
      });
    },
    cancelEdit() {
      this.resetNewMapping();
    },
    resetNewMapping() {
      this.editingMappingId = null;
      this.newMapping = {
        pollutant_code: 'PM25',
        parameter_id: this.stationParameters.length > 0 ? this.stationParameters[0].id : null,
        source_unit: 'µg/m³',
        conversion_factor: 1.0,
        min_completeness: 75.0,
        daily_start_hour: 0,
        enabled: true
      };
      this.onPollutantChange();
    },
    isPollutantMapped(code) {
      return (this.stationMappings || []).some(m => m.pollutant_code === code && m.enabled !== false);
    },
    getPollutantName(code) {
      const found = this.sevenOfficialParameters.find(p => p.code === code);
      return found ? found.name : code;
    },
    getMappingChannelName(m) {
      if (m.parameter) {
        return m.parameter.parameter_name || m.parameter.name || m.parameter.parameter_code || m.parameter.code || `Param ${m.parameter_id}`;
      }
      const p = this.stationParameters.find(sp => sp.id === m.parameter_id);
      if (p) {
        return p.parameter_name || p.name || p.displayName || `Param ${m.parameter_id}`;
      }
      return m.pollutant_code || `Param #${m.parameter_id}`;
    },
    getMappingDeviceName(m) {
      if (m.device) {
        return m.device.device_code || m.device.code || m.device.name || '';
      }
      const p = this.stationParameters.find(sp => sp.id === m.parameter_id);
      if (p && p.deviceCode) {
        return p.deviceCode;
      }
      return m.device_id ? `DEV-${m.device_id}` : '';
    },
    isDuplicatePollutant(code, id) {
      return (this.stationMappings || []).filter(m => m.pollutant_code === code).length > 1;
    },
    async saveMapping() {
      if (!this.newMapping.parameter_id || !this.newMapping.pollutant_code) return;
      const selectedParam = this.stationParameters.find(p => p.id === this.newMapping.parameter_id);
      const devId = selectedParam ? selectedParam.deviceId : 0;
      this.savingMapping = true;
      try {
        const token = this.getAuthToken();
        const payload = {
          id: this.editingMappingId || undefined,
          device_id: devId,
          parameter_id: this.newMapping.parameter_id,
          pollutant_code: this.newMapping.pollutant_code,
          source_unit: this.newMapping.source_unit || 'µg/m³',
          conversion_factor: Number(this.newMapping.conversion_factor) || 1.0,
          min_completeness: Number(this.newMapping.min_completeness) || 75.0,
          daily_start_hour: Number(this.newMapping.daily_start_hour) || 0,
          enabled: true
        };
        await axios.post('/api/ispu/mappings', payload, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        this.resetNewMapping();
        await this.loadAllMappings();
        await this.fetchLatestISPU();
      } catch (err) {
        console.error('Failed to save mapping:', err);
        alert('Gagal menyimpan parameter mapping: ' + (err.response?.data?.error || err.message));
      } finally {
        this.savingMapping = false;
      }
    },
    async deleteMapping(id) {
      if (!confirm('Hapus konfigurasi mapping polutan ini?')) return;
      try {
        const token = this.getAuthToken();
        await axios.delete(`/api/ispu/mappings/${id}`, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        if (this.editingMappingId === id) {
          this.resetNewMapping();
        }
        await this.loadAllMappings();
        await this.fetchLatestISPU();
      } catch (err) {
        console.error('Failed to delete mapping:', err);
        alert('Gagal menghapus mapping: ' + (err.response?.data?.error || err.message));
      }
    }
  }
};
</script>

<style scoped>
.ispu-breakdown-tile {
  border-width: 1px !important;
  box-shadow: none !important;
}

/* Match other main SCADA cards: #1E3A8A dark blue border in light mode, #244356 in dark mode */
.ispu-summary-card,
.ispu-breakdown-tray,
.ispu-trend-drawer {
  border-color: #1E3A8A !important;
}

:not(.dark) .ispu-summary-card,
:not(.dark) .ispu-breakdown-tray,
:not(.dark) .ispu-trend-drawer,
:global(.light-mode) .ispu-summary-card,
:global(.light-mode) .ispu-breakdown-tray,
:global(.light-mode) .ispu-trend-drawer {
  border-color: #1E3A8A !important;
  box-shadow: 0 1px 4px rgba(30, 58, 138, 0.08);
}

:global(.dark) .ispu-summary-card,
:global(.dark-mode) .ispu-summary-card {
  background-color: #0D2232 !important;
  border-color: #244356 !important;
  box-shadow: none;
}

:global(.dark) .ispu-breakdown-tray,
:global(.dark-mode) .ispu-breakdown-tray,
:global(.dark) .ispu-trend-drawer,
:global(.dark-mode) .ispu-trend-drawer {
  background-color: #0B1C2B !important;
  border-color: #244356 !important;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
}
</style>
