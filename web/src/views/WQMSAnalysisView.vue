<template>
  <div class="space-y-4 p-3 sm:p-5 max-w-[1920px] mx-auto min-h-screen text-slate-800 dark:text-slate-100 font-sans">
    <!-- Top Header & Controls -->
    <div class="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-slate-200 dark:border-[#20354B]">
      <div class="flex items-start space-x-3">
        <router-link
          to="/monitoring/dashboard-operator"
          class="mt-0.5 p-1.5 rounded-lg bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-[#20354B] transition"
          title="Kembali ke Dashboard Operator"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18" />
          </svg>
        </router-link>

        <div>
          <div class="flex flex-wrap items-center gap-2">
            <span class="w-2.5 h-2.5 rounded-full bg-sky-500 animate-pulse"></span>
            <h1 class="text-lg sm:text-xl font-bold tracking-tight text-slate-900 dark:text-white uppercase">
              STATUS MUTU AIR — INDEKS PENCEMARAN (IP) &amp; IKA
            </h1>
          </div>
          <div class="flex flex-wrap items-center gap-1.5 mt-1">
            <span class="px-2 py-0.5 rounded-full text-[10px] font-bold bg-sky-500/10 text-sky-600 dark:text-sky-400 border border-sky-500/30">
              ● WQMS
            </span>
            <span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
              PP NO. 22/2021
            </span>
            <span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
              KEPMEN LH 115/2003
            </span>
            <span class="px-2 py-0.5 rounded text-[10px] font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
              PERMEN LHK 27/2021
            </span>
          </div>
        </div>
      </div>

      <!-- Right Action Controls -->
      <div class="flex flex-wrap items-center gap-2">
        <!-- Water Class Selector -->
        <div class="flex items-center space-x-1.5 bg-white dark:bg-[#101B2B] px-2.5 py-1 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs">
          <span class="text-slate-500 font-medium">Kelas Air:</span>
          <select
            v-model="selectedWaterClass"
            @change="onWaterClassChange"
            :disabled="!canRecalculate || savingConfig"
            class="bg-transparent font-bold text-slate-800 dark:text-white focus:outline-none cursor-pointer"
          >
            <option value="CLASS_1">Kelas I (Air Baku Minum)</option>
            <option value="CLASS_2">Kelas II (Sungai / Rekreasi)</option>
            <option value="CLASS_3">Kelas III (Perikanan &amp; Peternakan)</option>
            <option value="CLASS_4">Kelas IV (Pengairan)</option>
          </select>
        </div>

        <!-- Mode Toggle -->
        <div class="flex items-center space-x-1 bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs">
          <button
            type="button"
            @click="setMode('REGULATORY')"
            class="px-2.5 py-1 rounded-md transition font-semibold"
            :class="calcMode === 'REGULATORY'
              ? 'bg-sky-600 text-white shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            📅 24-Jam (Regulasi)
          </button>
          <button
            type="button"
            @click="setMode('OPERATIONAL')"
            class="px-2.5 py-1 rounded-md transition font-semibold"
            :class="calcMode === 'OPERATIONAL'
              ? 'bg-sky-600 text-white shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            ⏱️ Per Agregasi ({{ aggIntervalLabel }})
          </button>
        </div>

        <button
          v-if="canManageMapping"
          type="button"
          @click="openMappingsModal"
          class="px-2.5 py-1.5 rounded-lg text-xs font-semibold transition border flex items-center space-x-1.5 bg-white dark:bg-[#101B2B] text-slate-700 dark:text-slate-200 border-slate-300 dark:border-[#20354B] hover:border-sky-500 hover:text-sky-600"
        >
          <span>⚙️</span>
          <span>Mappings &amp; Lab</span>
        </button>

        <button
          v-if="canRecalculate"
          type="button"
          @click="triggerRecalculate"
          :disabled="recalculating"
          class="px-2.5 py-1.5 rounded-lg text-xs font-semibold transition border flex items-center space-x-1.5 bg-white dark:bg-[#101B2B] text-slate-700 dark:text-slate-200 border-slate-300 dark:border-[#20354B] hover:border-sky-500 hover:text-sky-600 disabled:opacity-50"
        >
          <span>🔄</span>
          <span>{{ recalculating ? 'Menghitung...' : 'Recalculate' }}</span>
        </button>

        <button
          type="button"
          @click="exportFile('csv')"
          class="px-2.5 py-1.5 rounded-lg text-xs font-semibold transition border flex items-center space-x-1 bg-sky-600 hover:bg-sky-500 text-white border-sky-500 shadow-xs"
        >
          <span>⬇️ CSV</span>
        </button>

        <button
          type="button"
          @click="exportFile('excel')"
          class="px-2.5 py-1.5 rounded-lg text-xs font-semibold transition border flex items-center space-x-1 bg-emerald-600 hover:bg-emerald-500 text-white border-emerald-500 shadow-xs"
        >
          <span>📊 Excel</span>
        </button>
      </div>
    </div>

    <!-- ROW 1: DUAL HERO CARDS (IPj & SKOR IKA) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-4">
      <!-- Left Hero Card: Indeks Pencemaran (IPj) -->
      <div class="lg:col-span-6 p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs flex flex-col justify-between space-y-4">
        <div>
          <div class="flex items-center justify-between">
            <span class="text-xs font-extrabold uppercase tracking-wider text-slate-700 dark:text-slate-300">
              INDEKS PENCEMARAN SAAT INI — IPj {{ waterClassShort }}
            </span>
            <span class="text-[11px] font-semibold text-slate-400">
              {{ calcMode === 'OPERATIONAL' ? `Per Agregasi (${aggIntervalLabel})` : 'Rata-rata 24 Jam' }}
            </span>
          </div>

          <div class="flex items-center space-x-3 mt-3">
            <span class="text-4xl sm:text-5xl font-black tracking-tight tabular-nums text-slate-900 dark:text-white">
              {{ displayIP }}
            </span>
            <div>
              <span
                class="inline-block px-2.5 py-1 rounded-md text-xs font-extrabold uppercase tracking-wide"
                :class="ipStatusBadgeClass"
              >
                {{ ipCategoryTitle }}
              </span>
              <div class="text-xs text-slate-500 dark:text-slate-400 mt-1 font-medium">
                {{ ipRangeDescription }}
              </div>
            </div>
          </div>

          <!-- 4-Zone Horizontal IP Scale Gauge -->
          <div class="mt-5 pt-2">
            <div class="relative">
              <!-- Pointer Marker -->
              <div
                v-if="hasValidIP"
                class="absolute -top-3.5 transform -translate-x-1/2 transition-all duration-300"
                :style="{ left: ipMarkerLeftPct + '%' }"
              >
                <div class="w-0 h-0 border-l-[6px] border-r-[6px] border-t-[8px] border-l-transparent border-r-transparent border-t-slate-900 dark:border-t-white"></div>
              </div>

              <div class="grid grid-cols-4 h-3 rounded-full overflow-hidden">
                <div class="bg-emerald-500"></div>
                <div class="bg-amber-400"></div>
                <div class="bg-orange-500"></div>
                <div class="bg-rose-600"></div>
              </div>

              <div class="grid grid-cols-4 text-center text-[10.5px] mt-1.5 text-slate-600 dark:text-slate-400">
                <div>
                  <div class="font-bold text-slate-800 dark:text-slate-200">0 – 1.0</div>
                  <div>Memenuhi BM</div>
                </div>
                <div>
                  <div class="font-bold text-slate-800 dark:text-slate-200">1.0 – 5.0</div>
                  <div>Cemar Ringan</div>
                </div>
                <div>
                  <div class="font-bold text-slate-800 dark:text-slate-200">5.0 – 10</div>
                  <div>Cemar Sedang</div>
                </div>
                <div>
                  <div class="font-bold text-slate-800 dark:text-slate-200">&gt; 10</div>
                  <div>Cemar Berat</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="pt-3 border-t border-slate-100 dark:border-slate-800 text-xs text-slate-600 dark:text-slate-300 flex flex-wrap items-center gap-x-3 gap-y-1">
          <span>Parameter Dominan: <strong class="text-slate-900 dark:text-white">{{ dominantSummaryText }}</strong></span>
          <span>·</span>
          <span>Rasio Maks (M): <strong class="tabular-nums">{{ formatNum(activeIP && activeIP.max_ratio) }}</strong></span>
          <span>·</span>
          <span>Rasio Rata-rata (R): <strong class="tabular-nums">{{ formatNum(activeIP && activeIP.avg_ratio) }}</strong></span>
          <span>·</span>
          <span>Kelengkapan: <strong class="tabular-nums">{{ activeIP ? activeIP.completeness_percent : 0 }}% [{{ (activeIP && activeIP.calculation_status) || 'IDLE' }}]</strong></span>
        </div>
      </div>

      <!-- Right Hero Card: Skor IKA (Permen LHK 27/2021) -->
      <div class="lg:col-span-6 p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs flex flex-col justify-between">
        <div class="flex items-center justify-between mb-3">
          <span class="text-xs font-extrabold uppercase tracking-wider text-slate-700 dark:text-slate-300">
            INDEKS KUALITAS AIR — SKOR IKA {{ ikaSummary.window_days || 30 }} HARI
          </span>
          <span class="text-[11px] font-semibold text-slate-400">Permen LHK No. 27/2021</span>
        </div>

        <div class="flex flex-col sm:flex-row items-center gap-5 my-auto">
          <!-- Circular Gauge -->
          <div class="flex-shrink-0">
            <svg width="132" height="132" viewBox="0 0 132 132">
              <circle
                cx="66"
                cy="66"
                r="54"
                fill="none"
                class="stroke-slate-200 dark:stroke-slate-800"
                stroke-width="11"
              />
              <circle
                cx="66"
                cy="66"
                r="54"
                fill="none"
                :stroke="ikaRingColor"
                stroke-width="11"
                stroke-dasharray="339.3"
                :stroke-dashoffset="ikaRingDashOffset"
                stroke-linecap="round"
                transform="rotate(-90 66 66)"
                style="transition: stroke-dashoffset 0.4s ease;"
              />
              <text
                x="66"
                y="65"
                text-anchor="middle"
                class="fill-slate-900 dark:fill-white"
                font-size="27"
                font-weight="900"
              >{{ displayIKA }}</text>
              <text
                x="66"
                y="84"
                text-anchor="middle"
                class="fill-slate-500 dark:fill-slate-400"
                font-size="11.5"
                font-weight="700"
              >/ 100</text>
            </svg>
          </div>

          <!-- Weight Breakdown Bars -->
          <div class="flex-1 w-full space-y-2.5">
            <div class="flex items-center gap-2">
              <span class="px-2.5 py-0.5 rounded text-xs font-extrabold uppercase" :class="ikaBadgeClass">
                [{{ ikaSummary.category_label || 'Tidak Tersedia' }}]
              </span>
              <span class="text-xs text-slate-500 dark:text-slate-400">
                Basis: {{ ikaSummary.total_periods || 0 }} evaluasi
              </span>
            </div>

            <div v-for="bar in ikaWeightBars" :key="bar.label" class="space-y-1">
              <div class="flex items-center justify-between text-xs">
                <span class="flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-300">
                  <span class="w-2 h-2 rounded-full" :style="{ background: bar.color }"></span>
                  <span>{{ bar.label }}</span>
                </span>
                <span class="font-bold tabular-nums text-slate-800 dark:text-slate-100">
                  {{ bar.pct }}% ({{ bar.count }})
                </span>
              </div>
              <div class="h-1.5 w-full rounded-full bg-slate-100 dark:bg-slate-800 overflow-hidden">
                <div
                  class="h-full rounded-full transition-all duration-300"
                  :style="{ width: Math.max(bar.pct > 0 ? 4 : 0, bar.pct) + '%', background: bar.color }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ROW 2: PARAMETER EVALUATION TABLE -->
    <div class="p-4 sm:p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 class="text-sm font-bold text-slate-900 dark:text-white">
            Rincian Evaluasi Parameter terhadap Baku Mutu Air {{ waterClassShort }}
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            Rasio (Ci/Lij) &gt; 1.0 dikoreksi secara logaritmik: <code>1.0 + 5 · log10(Ci/Lij)</code> sesuai Kepmen LH No. 115 Tahun 2003.
          </p>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-xs border-collapse">
          <thead>
            <tr class="text-left text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800">
              <th class="py-2.5 px-3">Parameter</th>
              <th class="py-2.5 px-3">Sumber</th>
              <th class="py-2.5 px-3">Nilai Ukur (Ci)</th>
              <th class="py-2.5 px-3">Baku Mutu {{ waterClassShort }} (Lij)</th>
              <th class="py-2.5 px-3">Rasio Awal</th>
              <th class="py-2.5 px-3">Rasio (Ci/Lij) Baru</th>
              <th class="py-2.5 px-3">Evaluasi Rumus</th>
              <th class="py-2.5 px-3 text-right">Status Mutu</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="pr in sortedParameterResults"
              :key="pr.param_code"
              class="border-b border-slate-100 dark:border-slate-800/70 hover:bg-slate-50/70 dark:hover:bg-slate-800/40"
            >
              <td class="py-2.5 px-3 font-bold text-slate-900 dark:text-white">
                {{ formatParamFullName(pr.param_code) }}
              </td>
              <td class="py-2.5 px-3">
                <span
                  class="px-2 py-0.5 rounded text-[10.5px] font-semibold"
                  :class="pr.source_type === 'LAB'
                    ? 'bg-purple-500/15 text-purple-700 dark:text-purple-300'
                    : 'bg-sky-500/10 text-sky-700 dark:text-sky-300'"
                >
                  {{ pr.source_type === 'LAB' ? 'Input Lab Berkala' : 'Sensor Online' }}
                </span>
              </td>
              <td class="py-2.5 px-3 font-bold tabular-nums">
                {{ formatNum(pr.normalized_measured) }} {{ pr.normalized_unit }}
              </td>
              <td class="py-2.5 px-3 text-slate-600 dark:text-slate-300 font-medium">
                {{ pr.standard_label }}
              </td>
              <td class="py-2.5 px-3 tabular-nums text-slate-500">
                {{ formatNum(pr.raw_ratio) }}
              </td>
              <td class="py-2.5 px-3 font-extrabold tabular-nums">
                {{ formatNum(pr.adjusted_ratio) }}
                <span v-if="pr.is_dominant" class="text-amber-600 dark:text-amber-400 ml-1">(Maks)</span>
              </td>
              <td class="py-2.5 px-3 font-mono text-[11px] text-slate-500 dark:text-slate-400">
                {{ pr.formula_note || '-' }}
              </td>
              <td class="py-2.5 px-3 text-right">
                <span
                  class="inline-block px-2.5 py-0.5 rounded-full text-[11px] font-bold"
                  :class="pr.exceeds_standard
                    ? 'bg-amber-500/20 text-amber-800 dark:text-amber-300 border border-amber-500/40'
                    : 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30'"
                >
                  {{ pr.exceeds_standard ? (pr.is_dominant ? 'Melebihi BM (Dominan)' : 'Melebihi BM') : 'Memenuhi BM' }}
                </span>
              </td>
            </tr>
            <tr v-if="sortedParameterResults.length === 0">
              <td colspan="8" class="py-8 text-center text-slate-500 dark:text-slate-400">
                Belum ada parameter kualitas air yang memiliki data. Klik tombol <strong>⚙️ Mappings &amp; Lab</strong> di kanan atas untuk memetakan kanal sensor atau menambahkan hasil uji laboratorium.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- ROW 3: HISTORICAL TREND CHART -->
    <div class="p-4 sm:p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 class="text-sm font-bold text-slate-900 dark:text-white">
            Tren Historis Indeks Pencemaran (IP) — {{ trendRange.toUpperCase() }}
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            Garis hijau putus-putus menandakan batas Memenuhi Baku Mutu (IP = 1.0); garis kuning menandakan batas atas Cemar Ringan (IP = 5.0).
          </p>
        </div>

        <div class="flex items-center gap-1.5">
          <button
            v-if="canRecalculate"
            type="button"
            @click="runBackfill"
            :disabled="backfilling"
            class="px-2.5 py-1 rounded-lg text-xs font-semibold border border-slate-300 dark:border-slate-700 hover:border-sky-500 text-slate-600 dark:text-slate-300"
          >
            {{ backfilling ? 'Backfill...' : '⚡ Backfill Riwayat' }}
          </button>
          <button
            v-for="r in ['24h', '7d', '30d']"
            :key="r"
            type="button"
            @click="selectTrendRange(r)"
            class="px-2.5 py-1 rounded-lg text-xs font-semibold border transition"
            :class="trendRange === r
              ? 'bg-sky-600 text-white border-sky-500'
              : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700'"
          >
            {{ r.toUpperCase() }}
          </button>
        </div>
      </div>

      <div v-if="trendDots.length > 0" class="w-full overflow-x-auto">
        <svg viewBox="0 0 960 230" class="w-full h-56 min-w-[600px]">
          <defs>
            <linearGradient id="wqmsIpAreaGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#0EA5E9" stop-opacity="0.28" />
              <stop offset="100%" stop-color="#10B981" stop-opacity="0.02" />
            </linearGradient>
          </defs>

          <!-- Threshold lines -->
          <line x1="48" :y1="chartY(1.0)" x2="940" :y2="chartY(1.0)" stroke="#10B981" stroke-dasharray="5 4" stroke-width="1.3" />
          <text x="52" :y="Math.max(14, chartY(1.0) - 5)" fill="#10B981" font-size="10.5" font-weight="700">Batas Baku Mutu (IP = 1.0)</text>

          <line x1="48" :y1="chartY(5.0)" x2="940" :y2="chartY(5.0)" stroke="#F59E0B" stroke-dasharray="5 4" stroke-width="1.3" />
          <text x="52" :y="Math.max(14, chartY(5.0) - 5)" fill="#F59E0B" font-size="10.5" font-weight="700">Batas Cemar Ringan (IP = 5.0)</text>

          <!-- Area + Line -->
          <polygon v-if="trendAreaPoints" :points="trendAreaPoints" fill="url(#wqmsIpAreaGrad)" />
          <polyline
            fill="none"
            stroke="#0284C7"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
            :points="trendLinePoints"
          />
          <circle
            v-for="(pt, idx) in trendDots"
            :key="idx"
            :cx="pt.x"
            :cy="pt.y"
            r="3.2"
            :fill="pt.color"
          >
            <title>{{ pt.label }}</title>
          </circle>
        </svg>
      </div>
      <div v-else class="py-10 text-center text-xs text-slate-500 dark:text-slate-400">
        Belum ada titik riwayat tren pada rentang ini. Klik <strong>⚡ Backfill Riwayat</strong> atau <strong>🔄 Recalculate</strong> untuk membentuk histori dari tabel agregasi.
      </div>
    </div>

    <!-- ROW 4: HISTORICAL RECORDS TABLE -->
    <div class="p-4 sm:p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-bold text-slate-900 dark:text-white">Riwayat Perhitungan IP ({{ historyTotal }} Rekaman)</h2>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-xs border-collapse">
          <thead>
            <tr class="text-left text-[11px] uppercase tracking-wider text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800">
              <th class="py-2 px-3">Akhir Periode</th>
              <th class="py-2 px-3">Mode</th>
              <th class="py-2 px-3">Kelas Air</th>
              <th class="py-2 px-3">Nilai IP</th>
              <th class="py-2 px-3">Kategori Mutu</th>
              <th class="py-2 px-3">Dominan</th>
              <th class="py-2 px-3">M / R</th>
              <th class="py-2 px-3">Kelengkapan</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="row in historyRecords"
              :key="row.id"
              class="border-b border-slate-100 dark:border-slate-800/70"
            >
              <td class="py-2 px-3 font-mono">{{ formatDateTime(row.period_end) }}</td>
              <td class="py-2 px-3">{{ row.calculation_mode }}</td>
              <td class="py-2 px-3 font-semibold">{{ formatClassCode(row.water_class) }}</td>
              <td class="py-2 px-3 font-extrabold tabular-nums">{{ formatNum(row.overall_ip) }}</td>
              <td class="py-2 px-3 font-semibold">{{ formatIPCategory(row.overall_category) }}</td>
              <td class="py-2 px-3">{{ row.dominant_parameter || '-' }} ({{ formatNum(row.dominant_concentration) }} {{ row.dominant_unit }})</td>
              <td class="py-2 px-3 tabular-nums">{{ formatNum(row.max_ratio) }} / {{ formatNum(row.avg_ratio) }}</td>
              <td class="py-2 px-3 tabular-nums">{{ row.completeness_percent }}% ({{ row.calculation_status }})</td>
            </tr>
            <tr v-if="historyRecords.length === 0">
              <td colspan="8" class="py-6 text-center text-slate-400">Belum ada rekaman riwayat.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Mappings & Lab Modal (Shared capability) -->
    <div
      v-if="showMappingsModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-xs p-4"
      @click.self="showMappingsModal = false"
    >
      <div class="w-full max-w-3xl rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 shadow-2xl p-5 space-y-4 max-h-[88vh] overflow-y-auto">
        <div class="flex items-center justify-between pb-2 border-b border-slate-200 dark:border-slate-800">
          <div>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white">Pemetaan Parameter WQMS &amp; Input Uji Laboratorium</h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">Petakan sensor online atau tambahkan nilai uji lab berkala (misal Total Fosfat / Coliform).</p>
          </div>
          <button type="button" @click="showMappingsModal = false" class="px-2 py-1 text-slate-400 hover:text-slate-700 dark:hover:text-white">✕</button>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-12 gap-2.5 p-3.5 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-200 dark:border-slate-800 text-xs">
          <div class="sm:col-span-3">
            <label class="block text-[10px] font-bold uppercase text-slate-500 mb-1">Parameter Resmi</label>
            <select v-model="mappingForm.param_code" class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800">
              <option v-for="opt in officialParamOptions" :key="opt.code" :value="opt.code">{{ opt.label }}</option>
            </select>
          </div>
          <div class="sm:col-span-3">
            <label class="block text-[10px] font-bold uppercase text-slate-500 mb-1">Sumber Data</label>
            <select v-model="mappingForm.source_type" class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800">
              <option value="SENSOR">Sensor Online</option>
              <option value="LAB">Input Lab Berkala</option>
            </select>
          </div>
          <div v-if="mappingForm.source_type === 'SENSOR'" class="sm:col-span-4">
            <label class="block text-[10px] font-bold uppercase text-slate-500 mb-1">Kanal Sensor</label>
            <select v-model.number="mappingForm.parameter_id" class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800">
              <option :value="0">-- Pilih Parameter Sensor --</option>
              <option v-for="sp in availableSensorParameters" :key="sp.id" :value="sp.id">
                {{ sp.displayLabel }}
              </option>
            </select>
          </div>
          <div v-else class="sm:col-span-4">
            <label class="block text-[10px] font-bold uppercase text-slate-500 mb-1">Nilai Uji Lab</label>
            <input
              v-model.number="mappingForm.lab_value"
              type="number"
              step="any"
              placeholder="Contoh: 0.16"
              class="w-full px-2.5 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800"
            />
          </div>
          <div class="sm:col-span-2 flex items-end">
            <button
              type="button"
              @click="saveMapping"
              :disabled="savingMapping"
              class="w-full py-1.5 px-3 rounded-lg bg-sky-600 hover:bg-sky-500 text-white font-semibold disabled:opacity-50"
            >
              {{ savingMapping ? '...' : '+ Simpan' }}
            </button>
          </div>
        </div>

        <table class="w-full text-xs border-collapse">
          <thead>
            <tr class="text-left text-[10.5px] uppercase text-slate-500 border-b border-slate-200 dark:border-slate-800">
              <th class="py-2 px-2">Parameter</th>
              <th class="py-2 px-2">Sumber</th>
              <th class="py-2 px-2">Detail Kanal / Nilai</th>
              <th class="py-2 px-2 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="m in mappingsList" :key="m.id" class="border-b border-slate-100 dark:border-slate-800/60">
              <td class="py-2 px-2 font-bold">{{ formatParamFullName(m.param_code) }}</td>
              <td class="py-2 px-2">{{ m.source_type === 'LAB' ? 'Input Lab Berkala' : 'Sensor Online' }}</td>
              <td class="py-2 px-2">
                <template v-if="m.source_type === 'LAB'">
                  <strong>{{ m.lab_value }}</strong> {{ m.normalized_unit }}
                </template>
                <template v-else>
                  {{ getMappingSensorLabel(m) }}
                  <span
                    v-if="!isOperator && m.parameter && m.parameter.parameter_code && m.parameter.parameter_code !== m.parameter.parameter_name"
                    class="text-slate-400"
                  >
                    ({{ m.parameter.parameter_code }})
                  </span>
                </template>
              </td>
              <td class="py-2 px-2 text-right">
                <button type="button" @click="removeMapping(m.id)" class="text-rose-500 hover:text-rose-600 font-semibold">Hapus</button>
              </td>
            </tr>
            <tr v-if="mappingsList.length === 0">
              <td colspan="4" class="py-4 text-center text-slate-400">Belum ada pemetaan parameter.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'WQMSAnalysisView',
  data() {
    return {
      calcMode: 'REGULATORY',
      selectedWaterClass: 'CLASS_2',
      bundle: null,
      trendRange: '30d',
      trendRecords: [],
      historyRecords: [],
      historyTotal: 0,
      devicesList: [],
      mappingsList: [],
      showMappingsModal: false,
      recalculating: false,
      backfilling: false,
      savingConfig: false,
      savingMapping: false,
      mappingForm: {
        param_code: 'COD',
        source_type: 'SENSOR',
        parameter_id: 0,
        lab_value: null,
      },
      wsUnsub: null,
    };
  },
  computed: {
    userRole() {
      const u = this.$store && this.$store.state && this.$store.state.user;
      return String((u && (u.role_name || u.role)) || '').toUpperCase();
    },
    isOperator() {
      const path = (this.$route && this.$route.path) || '';
      if (path.includes('/operator') || path.includes('dashboard-operator')) {
        return true;
      }
      if (this.userRole === 'OPERATOR') {
        return true;
      }
      return Boolean(this.$store && this.$store.getters && this.$store.getters.isOperator);
    },
    canManageMapping() {
      return true;
    },
    canRecalculate() {
      return ['SUPERADMIN', 'ADMIN', 'ADMINISTRATOR', 'ENGINEER', 'OPERATOR'].includes(this.userRole);
    },
    activeIP() {
      if (!this.bundle) return null;
      if (this.calcMode === 'OPERATIONAL' && this.bundle.operational_ip_result) {
        return this.bundle.operational_ip_result;
      }
      return this.bundle.ip_result || this.bundle.operational_ip_result || null;
    },
    ikaSummary() {
      return (this.bundle && this.bundle.ika_summary) || {
        score: 0,
        category: 'TIDAK_TERSEDIA',
        category_label: 'Tidak Tersedia',
        window_days: 30,
        total_periods: 0,
        count_memenuhi: 0,
        pct_memenuhi: 0,
        count_cemar_ringan: 0,
        pct_cemar_ringan: 0,
        count_cemar_sedang: 0,
        pct_cemar_sedang: 0,
        count_cemar_berat: 0,
        pct_cemar_berat: 0,
      };
    },
    hasValidIP() {
      return Boolean(
        this.activeIP &&
        this.activeIP.active_parameters_count > 0 &&
        this.activeIP.overall_category !== 'TIDAK_TERSEDIA'
      );
    },
    displayIP() {
      if (!this.hasValidIP) return '—';
      return Number(this.activeIP.overall_ip || 0).toFixed(2);
    },
    displayIKA() {
      if (!this.ikaSummary || !this.ikaSummary.total_periods) return '—';
      return Number(this.ikaSummary.score || 0).toFixed(1);
    },
    aggIntervalLabel() {
      const sec = Number((this.bundle && this.bundle.aggregation_interval_sec) || 120);
      if (sec >= 3600) return `${Math.round(sec / 3600)} Jam`;
      if (sec >= 60) return `${Math.round(sec / 60)} Menit`;
      return `${sec}s`;
    },
    waterClassShort() {
      const m = { CLASS_1: 'KELAS I', CLASS_2: 'KELAS II', CLASS_3: 'KELAS III', CLASS_4: 'KELAS IV' };
      return m[this.selectedWaterClass] || 'KELAS II';
    },
    ipCategoryTitle() {
      if (!this.hasValidIP) return 'TIDAK TERSEDIA';
      switch (this.activeIP.overall_category) {
        case 'MEMENUHI_BAKU_MUTU': return '[MEMENUHI BAKU MUTU]';
        case 'CEMAR_RINGAN': return '[CEMAR RINGAN]';
        case 'CEMAR_SEDANG': return '[CEMAR SEDANG]';
        case 'CEMAR_BERAT': return '[CEMAR BERAT]';
        default: return 'TIDAK TERSEDIA';
      }
    },
    ipRangeDescription() {
      if (!this.hasValidIP) return 'Menunggu pemetaan parameter & data agregasi air';
      switch (this.activeIP.overall_category) {
        case 'MEMENUHI_BAKU_MUTU': return '0 ≤ IP ≤ 1.0 (Kondisi air baik / memenuhi kelas peruntukan)';
        case 'CEMAR_RINGAN': return '1.0 < IP ≤ 5.0 (Mulai tercemar ringan oleh parameter dominan)';
        case 'CEMAR_SEDANG': return '5.0 < IP ≤ 10.0 (Tercemar sedang)';
        case 'CEMAR_BERAT': return 'IP > 10.0 (Tercemar berat)';
        default: return '';
      }
    },
    ipStatusBadgeClass() {
      if (!this.hasValidIP) return 'bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-300';
      switch (this.activeIP.overall_category) {
        case 'MEMENUHI_BAKU_MUTU': return 'bg-emerald-500/20 text-emerald-800 dark:text-emerald-300 border border-emerald-500/40';
        case 'CEMAR_RINGAN': return 'bg-amber-400/30 text-amber-900 dark:text-amber-200 border border-amber-500/50';
        case 'CEMAR_SEDANG': return 'bg-orange-500/25 text-orange-800 dark:text-orange-200 border border-orange-500/40';
        case 'CEMAR_BERAT': return 'bg-rose-500/25 text-rose-800 dark:text-rose-200 border border-rose-500/40';
        default: return 'bg-slate-200 text-slate-600';
      }
    },
    ipMarkerLeftPct() {
      if (!this.hasValidIP) return 0;
      const ip = Math.max(0, Number(this.activeIP.overall_ip) || 0);
      if (ip <= 1.0) return (ip / 1.0) * 25.0;
      if (ip <= 5.0) return 25.0 + ((ip - 1.0) / 4.0) * 25.0;
      if (ip <= 10.0) return 50.0 + ((ip - 5.0) / 5.0) * 25.0;
      return Math.min(97.0, 75.0 + ((ip - 10.0) / 10.0) * 25.0);
    },
    dominantSummaryText() {
      if (!this.hasValidIP || !this.activeIP.dominant_parameter) return '-';
      return `${this.activeIP.dominant_parameter} (Ci/Lij = ${this.formatNum(this.activeIP.max_ratio)})`;
    },
    ikaRingColor() {
      const s = Number(this.ikaSummary.score) || 0;
      if (!this.ikaSummary.total_periods) return '#94A3B8';
      if (s >= 90) return '#06B6D4';
      if (s >= 70) return '#10B981';
      if (s >= 50) return '#0D9488';
      if (s >= 25) return '#F59E0B';
      return '#EF4444';
    },
    ikaRingDashOffset() {
      const circ = 339.3;
      if (!this.ikaSummary.total_periods) return circ;
      const pct = Math.max(0, Math.min(100, Number(this.ikaSummary.score) || 0)) / 100.0;
      return circ * (1 - pct);
    },
    ikaBadgeClass() {
      const s = Number(this.ikaSummary.score) || 0;
      if (!this.ikaSummary.total_periods) return 'bg-slate-200 dark:bg-slate-800 text-slate-500';
      if (s >= 70) return 'bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30';
      if (s >= 50) return 'bg-teal-500/20 text-teal-700 dark:text-teal-300 border border-teal-500/30';
      if (s >= 25) return 'bg-amber-500/20 text-amber-800 dark:text-amber-300 border border-amber-500/30';
      return 'bg-rose-500/20 text-rose-700 dark:text-rose-300 border border-rose-500/30';
    },
    ikaWeightBars() {
      return [
        { label: 'Memenuhi Baku Mutu (Bobot 70)', pct: this.ikaSummary.pct_memenuhi || 0, count: this.ikaSummary.count_memenuhi || 0, color: '#10B981' },
        { label: 'Cemar Ringan (Bobot 50)', pct: this.ikaSummary.pct_cemar_ringan || 0, count: this.ikaSummary.count_cemar_ringan || 0, color: '#F59E0B' },
        { label: 'Cemar Sedang (Bobot 30)', pct: this.ikaSummary.pct_cemar_sedang || 0, count: this.ikaSummary.count_cemar_sedang || 0, color: '#F97316' },
        { label: 'Cemar Berat (Bobot 10)', pct: this.ikaSummary.pct_cemar_berat || 0, count: this.ikaSummary.count_cemar_berat || 0, color: '#EF4444' },
      ];
    },
    sortedParameterResults() {
      const list = (this.activeIP && this.activeIP.parameter_results) || [];
      return [...list].sort((a, b) => {
        if (a.is_dominant && !b.is_dominant) return -1;
        if (!a.is_dominant && b.is_dominant) return 1;
        return (Number(b.adjusted_ratio) || 0) - (Number(a.adjusted_ratio) || 0);
      });
    },
    chartMaxVal() {
      let m = 5.5;
      for (const r of this.trendRecords) {
        const v = Number(r.overall_ip) || 0;
        if (v + 0.6 > m) m = v + 0.6;
      }
      return m;
    },
    trendDots() {
      const n = this.trendRecords.length;
      if (!n) return [];
      return this.trendRecords.map((r, i) => {
        const x = n === 1 ? 490 : 52 + (i / (n - 1)) * 880;
        const val = Number(r.overall_ip) || 0;
        const y = this.chartY(val);
        let color = '#10B981';
        if (val > 10) color = '#EF4444';
        else if (val > 5) color = '#F97316';
        else if (val > 1) color = '#F59E0B';
        return {
          x: Number(x.toFixed(1)),
          y,
          color,
          label: `${this.formatDateTime(r.period_end)} · IP ${val.toFixed(2)} (${this.formatIPCategory(r.overall_category)})`,
        };
      });
    },
    trendLinePoints() {
      return this.trendDots.map(d => `${d.x},${d.y}`).join(' ');
    },
    trendAreaPoints() {
      if (this.trendDots.length < 2) return '';
      const first = this.trendDots[0];
      const last = this.trendDots[this.trendDots.length - 1];
      return `${first.x},210 ${this.trendLinePoints} ${last.x},210`;
    },
    officialParamOptions() {
      return [
        { code: 'PH', label: 'pH (Derajat Keasaman)' },
        { code: 'DO', label: 'DO (Oksigen Terlarut)' },
        { code: 'COD', label: 'COD' },
        { code: 'BOD', label: 'BOD5' },
        { code: 'TSS', label: 'TSS (Padatan Tersuspensi)' },
        { code: 'NO3', label: 'Nitrat (NO3-N)' },
        { code: 'NH3', label: 'Amonia (NH3-N)' },
        { code: 'TDS', label: 'TDS (Padatan Terlarut)' },
        { code: 'PO4', label: 'Total Fosfat (PO4-P)' },
        { code: 'TURBIDITY', label: 'Kekeruhan (Turbidity)' },
        { code: 'FECAL_COLI', label: 'Fecal Coliform' },
      ];
    },
    availableSensorParameters() {
      const out = [];
      for (const dev of this.devicesList) {
        if (!dev || dev.deleted_at || dev.enabled === false || dev.enabled === 0) continue;
        const devStatus = String(dev.status || dev.administrative_status || '').toUpperCase();
        if (devStatus === 'DISABLED' || devStatus === 'INACTIVE') continue;
        if (!Array.isArray(dev.parameters)) continue;
        const devCode = dev.device_code || dev.code || dev.name || '';
        for (const p of dev.parameters) {
          if (!p || p.deleted_at || p.enabled === false || p.enabled === 0) continue;
          const paramName = p.parameter_name || p.param_name || p.name || p.parameter_code || '';
          const unit = (p.unit || '').trim();
          const unitSuffix = (unit && unit !== '-') ? ` (${unit})` : '';
          const baseLabel = `${paramName}${unitSuffix}`;
          out.push({
            id: p.id || p.param_id,
            device_code: devCode,
            parameter_name: paramName,
            unit,
            displayLabel: (this.isOperator || !devCode) ? baseLabel : `[${devCode}] ${baseLabel}`,
          });
        }
      }
      return out;
    },
  },
  mounted() {
    this.loadAll();
    this.fetchDevices();
    window.addEventListener('wqms-calculated-event', this.onWqmsCalculatedEvent);
  },
  beforeDestroy() {
    window.removeEventListener('wqms-calculated-event', this.onWqmsCalculatedEvent);
  },
  methods: {
    onWqmsCalculatedEvent() {
      this.loadAll();
    },
    getAuthHeaders() {
      const token = localStorage.getItem('datalogger_token') || localStorage.getItem('token') || (this.$store && this.$store.state && this.$store.state.token) || '';
      return token ? { Authorization: `Bearer ${token}` } : {};
    },
    async loadAll() {
      await Promise.all([
        this.fetchLatest(),
        this.fetchTrend(),
        this.fetchHistory(),
      ]);
    },
    async fetchDevices() {
      try {
        const res = await axios.get('/api/devices', { headers: this.getAuthHeaders() });
        const items = (res.data && res.data.data && (res.data.data.items || res.data.data)) || [];
        const rawList = Array.isArray(items) ? items : [];
        const enriched = await Promise.all(
          rawList.map(async (dev) => {
            if (!dev || dev.deleted_at || dev.enabled === false) return dev;
            const statusUpper = String(dev.status || '').toUpperCase();
            if (statusUpper === 'DISABLED' || statusUpper === 'INACTIVE') return dev;
            if (Array.isArray(dev.parameters) && dev.parameters.length > 0) return dev;
            try {
              const pRes = await axios.get(`/api/devices/${dev.id}/parameters`, { headers: this.getAuthHeaders() });
              const params = (pRes.data && (pRes.data.data || pRes.data)) || [];
              return { ...dev, parameters: Array.isArray(params) ? params : [] };
            } catch (_) {
              return dev;
            }
          })
        );
        this.devicesList = enriched;
      } catch (e) {}
    },
    async fetchLatest() {
      try {
        const res = await axios.get(`/api/wqms/latest?mode=${this.calcMode}`, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.data) {
          this.bundle = res.data.data;
          if (this.bundle.config && this.bundle.config.water_class) {
            this.selectedWaterClass = this.bundle.config.water_class;
          }
        }
      } catch (e) {}
    },
    async fetchTrend() {
      try {
        const res = await axios.get(`/api/wqms/trend?range=${this.trendRange}&mode=${this.calcMode}`, {
          headers: this.getAuthHeaders(),
        });
        this.trendRecords = (res.data && res.data.data) || [];
      } catch (e) {
        this.trendRecords = [];
      }
    },
    async fetchHistory() {
      try {
        const res = await axios.get(`/api/wqms/history?page=1&limit=30&mode=${this.calcMode}`, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.data) {
          this.historyRecords = res.data.data.records || [];
          this.historyTotal = res.data.data.total || 0;
        }
      } catch (e) {}
    },
    setMode(m) {
      this.calcMode = m;
      this.loadAll();
    },
    selectTrendRange(r) {
      this.trendRange = r;
      this.fetchTrend();
    },
    async onWaterClassChange() {
      this.savingConfig = true;
      try {
        const res = await axios.put(
          '/api/wqms/config',
          { water_class: this.selectedWaterClass },
          { headers: this.getAuthHeaders() }
        );
        if (res.data && res.data.data) {
          this.bundle = res.data.data;
        }
        await Promise.all([this.fetchTrend(), this.fetchHistory()]);
      } catch (e) {
      } finally {
        this.savingConfig = false;
      }
    },
    async triggerRecalculate() {
      this.recalculating = true;
      try {
        const res = await axios.post('/api/wqms/recalculate', {}, { headers: this.getAuthHeaders() });
        if (res.data && res.data.data) {
          this.bundle = res.data.data;
        }
        await Promise.all([this.fetchTrend(), this.fetchHistory()]);
      } catch (e) {
      } finally {
        this.recalculating = false;
      }
    },
    async runBackfill() {
      this.backfilling = true;
      try {
        await axios.post('/api/wqms/backfill', { hours: 168 }, { headers: this.getAuthHeaders() });
        await this.loadAll();
      } catch (e) {
      } finally {
        this.backfilling = false;
      }
    },
    async openMappingsModal() {
      this.showMappingsModal = true;
      await this.fetchMappings();
    },
    async fetchMappings() {
      try {
        const res = await axios.get('/api/wqms/mappings', { headers: this.getAuthHeaders() });
        this.mappingsList = (res.data && res.data.data) || [];
      } catch (e) {}
    },
    async saveMapping() {
      if (this.mappingForm.source_type === 'SENSOR' && !this.mappingForm.parameter_id) return;
      if (this.mappingForm.source_type === 'LAB' && (this.mappingForm.lab_value === null || this.mappingForm.lab_value === '')) return;
      this.savingMapping = true;
      try {
        await axios.post('/api/wqms/mappings', {
          param_code: this.mappingForm.param_code,
          source_type: this.mappingForm.source_type,
          parameter_id: this.mappingForm.source_type === 'SENSOR' ? this.mappingForm.parameter_id : null,
          lab_value: this.mappingForm.source_type === 'LAB' ? Number(this.mappingForm.lab_value) : null,
          enabled: true,
        }, { headers: this.getAuthHeaders() });
        await this.fetchMappings();
        await this.loadAll();
      } catch (e) {
      } finally {
        this.savingMapping = false;
      }
    },
    async removeMapping(id) {
      try {
        await axios.delete(`/api/wqms/mappings/${id}`, { headers: this.getAuthHeaders() });
        await this.fetchMappings();
        await this.loadAll();
      } catch (e) {}
    },
    async exportFile(type) {
      try {
        const res = await axios.get(`/api/wqms/export/${type}?mode=${this.calcMode}`, {
          headers: this.getAuthHeaders(),
          responseType: 'blob',
        });
        const blob = new Blob([res.data], { type: res.headers['content-type'] || 'text/csv' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        const ext = type === 'excel' ? 'xlsx' : 'csv';
        a.download = `wqms_ip_ika_${this.calcMode.toLowerCase()}_${new Date().toISOString().slice(0, 10)}.${ext}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      } catch (e) {}
    },
    chartY(val) {
      const top = 18;
      const bottom = 210;
      const ratio = Math.max(0, Math.min(1, val / this.chartMaxVal));
      return Number((bottom - ratio * (bottom - top)).toFixed(1));
    },
    formatNum(v) {
      if (v === null || v === undefined || isNaN(Number(v))) return '0.00';
      return Number(v).toFixed(2);
    },
    formatDateTime(ts) {
      if (!ts) return '-';
      const d = new Date(ts);
      return isNaN(d.getTime()) ? String(ts) : d.toLocaleString('id-ID');
    },
    formatClassCode(c) {
      const m = { CLASS_1: 'Kelas I', CLASS_2: 'Kelas II', CLASS_3: 'Kelas III', CLASS_4: 'Kelas IV' };
      return m[c] || c || 'Kelas II';
    },
    formatIPCategory(c) {
      const m = {
        MEMENUHI_BAKU_MUTU: 'Memenuhi BM',
        CEMAR_RINGAN: 'Cemar Ringan',
        CEMAR_SEDANG: 'Cemar Sedang',
        CEMAR_BERAT: 'Cemar Berat',
      };
      return m[c] || c || '-';
    },
    formatParamFullName(code) {
      const m = {
        PH: 'pH (Derajat Keasaman)',
        DO: 'DO (Oksigen Terlarut)',
        COD: 'COD',
        BOD: 'BOD5',
        TSS: 'TSS (Padatan Tersuspensi)',
        NO3: 'Nitrat (NO3-N)',
        NH3: 'Amonia (NH3-N)',
        TDS: 'TDS (Padatan Terlarut)',
        PO4: 'Total Fosfat (PO4-P)',
        TURBIDITY: 'Kekeruhan (Turbidity)',
        FECAL_COLI: 'Fecal Coliform',
      };
      return m[code] || code || '-';
    },
    getMappingSensorLabel(m) {
      if (m && m.parameter && (m.parameter.parameter_name || m.parameter.name)) {
        return m.parameter.parameter_name || m.parameter.name;
      }
      if (m && m.parameter_id) {
        const found = (this.availableSensorParameters || []).find(sp => Number(sp.id) === Number(m.parameter_id));
        if (found && found.parameter_name) return found.parameter_name;
      }
      if (this.isOperator) {
        return this.formatParamFullName(m && m.param_code);
      }
      return m && m.parameter_id ? `Parameter #${m.parameter_id}` : '-';
    },
  },
};
</script>
