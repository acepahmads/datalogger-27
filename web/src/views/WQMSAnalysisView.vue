<template>
  <div class="space-y-3.5 sm:space-y-4 p-2.5 sm:p-5 max-w-[1920px] mx-auto min-h-screen text-slate-800 dark:text-slate-100 font-sans">
    <!-- MOBILE UNIFIED HEADER & CONTROL CARD (< 768px) -->
    <div
      id="wqms-mobile-control-card"
      class="md:hidden p-3 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs space-y-2.5"
    >
      <!-- Row 1: Back + Title + WQMS Pill -->
      <div class="flex items-center justify-between gap-2">
        <div class="flex items-center gap-2.5 min-w-0">
          <router-link
            to="/monitoring/dashboard-operator"
            class="w-8 h-8 rounded-xl bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-[#20354B] flex items-center justify-center shrink-0 transition"
            title="Kembali ke Dashboard Operator"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.2" d="M15 19l-7-7 7-7" />
            </svg>
          </router-link>
          <div class="min-w-0">
            <div class="flex items-center gap-1.5">
              <span class="w-2 h-2 rounded-full bg-sky-500 animate-pulse shrink-0"></span>
              <h1 class="text-[15px] font-extrabold tracking-tight text-slate-900 dark:text-white truncate">
                Mutu Air (IP &amp; IKA)
              </h1>
            </div>
            <p class="text-[10px] text-slate-500 dark:text-slate-400 truncate">
              PP 22/2021 · Kepmen 115/2003 · Permen 27/2021
            </p>
          </div>
        </div>
        <span class="px-2 py-0.5 rounded-full text-[9.5px] font-bold bg-sky-500/10 text-sky-600 dark:text-sky-400 border border-sky-500/30 shrink-0 whitespace-nowrap">
          ● WQMS
        </span>
      </div>

      <!-- Row 2: Full-Width Water Class Selector -->
      <div class="flex items-center justify-between gap-2 bg-slate-50 dark:bg-[#101B2B] px-3 py-1.5 rounded-xl border border-slate-200 dark:border-[#20354B] text-xs">
        <span class="text-slate-500 dark:text-slate-400 font-semibold shrink-0 text-[11px]">💧 Kelas Air:</span>
        <select
          v-model="selectedWaterClass"
          @change="onWaterClassChange"
          :disabled="!canRecalculate || savingConfig"
          class="min-w-0 flex-1 text-right bg-transparent font-bold text-slate-900 dark:text-white focus:outline-none cursor-pointer truncate text-[11.5px]"
        >
          <option value="CLASS_1">Kelas I (Air Baku Minum)</option>
          <option value="CLASS_2">Kelas II (Sungai / Rekreasi)</option>
          <option value="CLASS_3">Kelas III (Perikanan &amp; Ternak)</option>
          <option value="CLASS_4">Kelas IV (Pengairan)</option>
        </select>
      </div>

      <!-- Row 3: Symmetrical 2-Column Mode Switcher -->
      <div class="grid grid-cols-2 gap-1 p-1 rounded-xl bg-slate-100 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-[11px]">
        <button
          type="button"
          @click="setMode('REGULATORY')"
          class="py-1.5 px-2 rounded-lg transition font-bold text-center truncate"
          :class="calcMode === 'REGULATORY'
            ? 'bg-sky-600 text-white shadow-xs'
            : 'text-slate-600 dark:text-slate-400'"
        >
          📅 24-Jam (Regulasi)
        </button>
        <button
          type="button"
          @click="setMode('OPERATIONAL')"
          class="py-1.5 px-2 rounded-lg transition font-bold text-center truncate"
          :class="calcMode === 'OPERATIONAL'
            ? 'bg-sky-600 text-white shadow-xs'
            : 'text-slate-600 dark:text-slate-400'"
        >
          ⏱️ Agregasi ({{ aggIntervalLabel }})
        </button>
      </div>

      <!-- Row 4: Symmetrical Equal-Column Action Bar -->
      <div class="grid gap-1.5" :class="mobileActionGridCols">
        <button
          v-if="canManageMapping"
          type="button"
          @click="openMappingsModal"
          class="h-8 px-1.5 rounded-xl text-[11px] font-bold transition border flex items-center justify-center gap-1 bg-slate-50 dark:bg-[#101B2B] text-slate-700 dark:text-slate-200 border-slate-200 dark:border-[#20354B] active:scale-[0.98]"
        >
          <span>⚙️</span>
          <span class="truncate">Map/Lab</span>
        </button>
        <button
          v-if="canRecalculate"
          type="button"
          @click="triggerRecalculate"
          :disabled="recalculating"
          class="h-8 px-1.5 rounded-xl text-[11px] font-bold transition border flex items-center justify-center gap-1 bg-slate-50 dark:bg-[#101B2B] text-slate-700 dark:text-slate-200 border-slate-200 dark:border-[#20354B] disabled:opacity-50 active:scale-[0.98]"
        >
          <span :class="{ 'animate-spin': recalculating }">🔄</span>
          <span class="truncate">{{ recalculating ? '...' : 'Hitung' }}</span>
        </button>
        <button
          type="button"
          @click="exportFile('csv')"
          class="h-8 px-1.5 rounded-xl text-[11px] font-bold transition border flex items-center justify-center gap-1 bg-sky-600 hover:bg-sky-500 text-white border-sky-500 shadow-xs active:scale-[0.98]"
        >
          <span>⬇️</span>
          <span>CSV</span>
        </button>
        <button
          type="button"
          @click="exportFile('excel')"
          class="h-8 px-1.5 rounded-xl text-[11px] font-bold transition border flex items-center justify-center gap-1 bg-emerald-600 hover:bg-emerald-500 text-white border-emerald-500 shadow-xs active:scale-[0.98]"
        >
          <span>📊</span>
          <span>Excel</span>
        </button>
      </div>
    </div>

    <!-- DESKTOP Top Header & Controls (>= 768px — 100% Preserved) -->
    <div class="hidden md:flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-slate-200 dark:border-[#20354B]">
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
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-3.5 sm:gap-4">
      <!-- Left Hero Card: Indeks Pencemaran (IPj) -->
      <div class="lg:col-span-6 min-w-0 p-3.5 sm:p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs flex flex-col justify-between space-y-3.5 sm:space-y-4">
        <div>
          <div class="flex items-center justify-between gap-2">
            <span class="text-[11px] sm:text-xs font-extrabold uppercase tracking-wider text-slate-700 dark:text-slate-300 truncate">
              <span class="sm:hidden">IPj SAAT INI — {{ waterClassShort }}</span>
              <span class="hidden sm:inline">INDEKS PENCEMARAN SAAT INI — IPj {{ waterClassShort }}</span>
            </span>
            <span class="px-2 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800/80 text-[10px] sm:text-[11px] font-semibold text-slate-500 dark:text-slate-400 shrink-0">
              {{ calcMode === 'OPERATIONAL' ? `Agregasi (${aggIntervalLabel})` : 'Rata-rata 24 Jam' }}
            </span>
          </div>

          <div class="flex items-center space-x-3 mt-2.5 sm:mt-3">
            <span class="text-4xl sm:text-5xl font-black tracking-tight tabular-nums leading-none text-slate-900 dark:text-white">
              {{ displayIP }}
            </span>
            <div class="min-w-0">
              <span
                class="inline-block px-2.5 py-0.5 sm:py-1 rounded-md text-[11px] sm:text-xs font-extrabold uppercase tracking-wide"
                :class="ipStatusBadgeClass"
              >
                <span class="sm:hidden">{{ ipCategoryCleanLabel }}</span>
                <span class="hidden sm:inline">{{ ipCategoryTitle }}</span>
              </span>
              <div class="text-[11px] sm:text-xs text-slate-500 dark:text-slate-400 mt-1 font-medium leading-snug">
                {{ ipRangeDescription }}
              </div>
            </div>
          </div>

          <!-- 4-Zone Horizontal IP Scale Gauge -->
          <div class="mt-4 sm:mt-5 pt-2">
            <div class="relative">
              <!-- Pointer Marker -->
              <div
                v-if="hasValidIP"
                class="absolute -top-3.5 transform -translate-x-1/2 transition-all duration-300"
                :style="{ left: ipMarkerLeftPct + '%' }"
              >
                <div class="w-0 h-0 border-l-[6px] border-r-[6px] border-t-[8px] border-l-transparent border-r-transparent border-t-slate-900 dark:border-t-white"></div>
              </div>

              <div class="grid grid-cols-4 h-2.5 sm:h-3 rounded-full overflow-hidden gap-0.5 bg-slate-200/60 dark:bg-slate-800">
                <div class="bg-emerald-500"></div>
                <div class="bg-amber-400"></div>
                <div class="bg-orange-500"></div>
                <div class="bg-rose-600"></div>
              </div>

              <div class="grid grid-cols-4 text-center text-[9.5px] sm:text-[10.5px] mt-1.5 text-slate-600 dark:text-slate-400 leading-tight gap-0.5">
                <div>
                  <div class="font-bold text-slate-800 dark:text-slate-200">0 – 1.0</div>
                  <div class="sm:hidden">Memenuhi</div>
                  <div class="hidden sm:block">Memenuhi BM</div>
                </div>
                <div>
                  <div class="font-bold text-slate-800 dark:text-slate-200">1.0 – 5.0</div>
                  <div class="sm:hidden">Ringan</div>
                  <div class="hidden sm:block">Cemar Ringan</div>
                </div>
                <div>
                  <div class="font-bold text-slate-800 dark:text-slate-200">5.0 – 10</div>
                  <div class="sm:hidden">Sedang</div>
                  <div class="hidden sm:block">Cemar Sedang</div>
                </div>
                <div>
                  <div class="font-bold text-slate-800 dark:text-slate-200">&gt; 10</div>
                  <div class="sm:hidden">Berat</div>
                  <div class="hidden sm:block">Cemar Berat</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Mobile 2x2 Mini-Bento KPI Grid (< 640px) -->
        <div class="grid grid-cols-2 gap-2 pt-2.5 border-t border-slate-100 dark:border-slate-800 sm:hidden text-[11px]">
          <div class="p-2 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-100 dark:border-slate-800/80 min-w-0">
            <div class="text-[9.5px] uppercase font-bold text-slate-400">Dominan</div>
            <div class="font-extrabold text-slate-900 dark:text-white truncate mt-0.5">{{ dominantSummaryText }}</div>
          </div>
          <div class="p-2 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-100 dark:border-slate-800/80 min-w-0">
            <div class="text-[9.5px] uppercase font-bold text-slate-400">Rasio Maks / Rata</div>
            <div class="font-extrabold font-mono tabular-nums text-slate-900 dark:text-white mt-0.5">
              {{ formatNum(activeIP && activeIP.max_ratio) }} / {{ formatNum(activeIP && activeIP.avg_ratio) }}
            </div>
          </div>
          <div class="p-2 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-100 dark:border-slate-800/80 min-w-0">
            <div class="text-[9.5px] uppercase font-bold text-slate-400">Kelengkapan</div>
            <div class="font-extrabold font-mono tabular-nums text-slate-900 dark:text-white mt-0.5">
              {{ activeIP ? activeIP.completeness_percent : 0 }}%
            </div>
          </div>
          <div class="p-2 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-100 dark:border-slate-800/80 min-w-0">
            <div class="text-[9.5px] uppercase font-bold text-slate-400">Status Kalkulasi</div>
            <div class="font-extrabold font-mono text-sky-600 dark:text-sky-400 truncate mt-0.5">
              {{ (activeIP && activeIP.calculation_status) || 'IDLE' }}
            </div>
          </div>
        </div>

        <!-- Desktop Footer (>= 640px) -->
        <div class="hidden sm:flex pt-3 border-t border-slate-100 dark:border-slate-800 text-xs text-slate-600 dark:text-slate-300 flex-wrap items-center gap-x-3 gap-y-1">
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
      <div class="lg:col-span-6 min-w-0 p-3.5 sm:p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs flex flex-col justify-between">
        <div class="flex items-center justify-between gap-2 mb-3">
          <span class="text-[11px] sm:text-xs font-extrabold uppercase tracking-wider text-slate-700 dark:text-slate-300 truncate">
            <span class="sm:hidden">SKOR IKA — {{ ikaSummary.window_days || 30 }} HARI</span>
            <span class="hidden sm:inline">INDEKS KUALITAS AIR — SKOR IKA {{ ikaSummary.window_days || 30 }} HARI</span>
          </span>
          <span class="text-[10px] sm:text-[11px] font-semibold text-slate-400 shrink-0">Permen LHK 27/2021</span>
        </div>

        <!-- MOBILE Compact Side-by-Side Ring + Summary (< 640px) -->
        <div class="sm:hidden space-y-3">
          <div class="flex items-center gap-3.5 p-2.5 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-100 dark:border-slate-800/80">
            <svg viewBox="0 0 132 132" class="w-20 h-20 flex-shrink-0">
              <circle cx="66" cy="66" r="54" fill="none" class="stroke-slate-200 dark:stroke-slate-800" stroke-width="12" />
              <circle
                cx="66"
                cy="66"
                r="54"
                fill="none"
                :stroke="ikaRingColor"
                stroke-width="12"
                stroke-dasharray="339.3"
                :stroke-dashoffset="ikaRingDashOffset"
                stroke-linecap="round"
                transform="rotate(-90 66 66)"
                style="transition: stroke-dashoffset 0.4s ease;"
              />
              <text x="66" y="65" text-anchor="middle" class="fill-slate-900 dark:fill-white" font-size="28" font-weight="900">{{ displayIKA }}</text>
              <text x="66" y="84" text-anchor="middle" class="fill-slate-500 dark:fill-slate-400" font-size="12" font-weight="700">/ 100</text>
            </svg>
            <div class="min-w-0 space-y-1">
              <span class="inline-block px-2.5 py-0.5 rounded-md text-[11px] font-extrabold uppercase" :class="ikaBadgeClass">
                {{ ikaSummary.category_label || 'Tidak Tersedia' }}
              </span>
              <div class="text-xs font-bold text-slate-800 dark:text-slate-100">
                Basis: {{ ikaSummary.total_periods || 0 }} evaluasi
              </div>
              <div class="text-[10.5px] text-slate-500 dark:text-slate-400 leading-snug">
                Frekuensi status mutu air {{ ikaSummary.window_days || 30 }} hari terakhir
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <div v-for="bar in ikaWeightBars" :key="'mob-' + bar.label" class="space-y-1">
              <div class="flex items-center justify-between text-[11px]">
                <span class="flex items-center gap-1.5 font-medium text-slate-700 dark:text-slate-300 truncate">
                  <span class="w-2 h-2 rounded-full shrink-0" :style="{ background: bar.color }"></span>
                  <span class="truncate">{{ bar.label }}</span>
                </span>
                <span class="font-bold tabular-nums text-slate-800 dark:text-slate-100 shrink-0 ml-2">
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

        <!-- DESKTOP Ring + Weight Bars (>= 640px — 100% Preserved) -->
        <div class="hidden sm:flex flex-row items-center gap-5 my-auto">
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

    <!-- ROW 2: PARAMETER EVALUATION SECTION -->
    <div class="p-3.5 sm:p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs space-y-3">
      <div class="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h2 class="text-xs sm:text-sm font-bold text-slate-900 dark:text-white">
            Rincian Evaluasi Parameter terhadap Baku Mutu Air {{ waterClassShort }}
          </h2>
          <p class="text-[11px] sm:text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Rasio (Ci/Lij) &gt; 1.0 dikoreksi secara logaritmik: <code>1.0 + 5 · log10(Ci/Lij)</code> sesuai Kepmen LH No. 115 Tahun 2003.
          </p>
        </div>
      </div>

      <!-- MOBILE PARAMETER EVALUATION CARDS (< 768px) -->
      <div class="md:hidden space-y-2.5">
        <div
          v-if="sortedParameterResults.length === 0"
          class="p-5 rounded-xl border border-dashed border-slate-200 dark:border-slate-800 text-center space-y-2.5"
        >
          <p class="text-xs text-slate-500 dark:text-slate-400 leading-relaxed">
            Belum ada parameter kualitas air yang memiliki data pada kelas ini.
          </p>
          <button
            v-if="canManageMapping"
            type="button"
            @click="openMappingsModal"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl bg-sky-600 hover:bg-sky-500 text-white text-xs font-bold shadow-xs"
          >
            <span>⚙️</span>
            <span>Petakan Sensor / Input Lab</span>
          </button>
        </div>

        <div
          v-for="pr in sortedParameterResults"
          :key="'mob-pr-' + pr.param_code"
          class="p-3 rounded-xl border space-y-2"
          :class="pr.exceeds_standard
            ? 'bg-amber-500/5 border-amber-500/35'
            : 'bg-slate-50/70 dark:bg-slate-900/60 border-slate-200 dark:border-slate-800'"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <div class="text-xs font-extrabold text-slate-900 dark:text-white truncate">
                {{ formatParamFullName(pr.param_code) }}
              </div>
              <span
                class="inline-block mt-0.5 px-1.5 py-0.2 rounded text-[9.5px] font-semibold"
                :class="pr.source_type === 'LAB'
                  ? 'bg-purple-500/15 text-purple-700 dark:text-purple-300'
                  : 'bg-sky-500/10 text-sky-700 dark:text-sky-300'"
              >
                {{ pr.source_type === 'LAB' ? 'Input Lab Berkala' : 'Sensor Online' }}
              </span>
            </div>
            <span
              class="px-2 py-0.5 rounded-full text-[10px] font-bold shrink-0"
              :class="pr.exceeds_standard
                ? 'bg-amber-500/20 text-amber-800 dark:text-amber-300 border border-amber-500/40'
                : 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30'"
            >
              {{ pr.exceeds_standard ? (pr.is_dominant ? 'Melebihi BM ★' : 'Melebihi BM') : 'Memenuhi BM' }}
            </span>
          </div>

          <div class="grid grid-cols-3 gap-1.5 text-[11px]">
            <div class="p-2 rounded-lg bg-white dark:bg-[#0F172A] border border-slate-200/70 dark:border-slate-800">
              <div class="text-[9px] uppercase font-bold text-slate-400">Nilai Ukur (Ci)</div>
              <div class="font-extrabold tabular-nums text-slate-900 dark:text-white mt-0.5 truncate">
                {{ formatNum(pr.normalized_measured) }} <span class="font-normal text-[10px] text-slate-500">{{ pr.normalized_unit }}</span>
              </div>
            </div>
            <div class="p-2 rounded-lg bg-white dark:bg-[#0F172A] border border-slate-200/70 dark:border-slate-800">
              <div class="text-[9px] uppercase font-bold text-slate-400">Baku Mutu (Lij)</div>
              <div class="font-bold text-slate-700 dark:text-slate-200 mt-0.5 truncate">
                {{ pr.standard_label }}
              </div>
            </div>
            <div class="p-2 rounded-lg bg-white dark:bg-[#0F172A] border border-slate-200/70 dark:border-slate-800">
              <div class="text-[9px] uppercase font-bold text-slate-400">Rasio (Ci/Lij)</div>
              <div class="font-extrabold tabular-nums mt-0.5 truncate" :class="pr.exceeds_standard ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'">
                {{ formatNum(pr.adjusted_ratio) }}
                <span v-if="pr.is_dominant" class="text-[9px]">(Maks)</span>
              </div>
            </div>
          </div>

          <!-- Ratio Progress Indicator & Formula Note -->
          <div class="space-y-1">
            <div class="h-1.5 w-full rounded-full bg-slate-200/80 dark:bg-slate-800 overflow-hidden">
              <div
                class="h-full rounded-full transition-all duration-300"
                :class="pr.exceeds_standard ? 'bg-amber-500' : 'bg-emerald-500'"
                :style="{ width: `${Math.min(100, Math.max(4, (Number(pr.adjusted_ratio || 0) / 2.0) * 100))}%` }"
              ></div>
            </div>
            <div v-if="pr.formula_note" class="text-[10px] font-mono text-slate-500 dark:text-slate-400 truncate">
              Rumus: {{ pr.formula_note }}
            </div>
          </div>
        </div>
      </div>

      <!-- DESKTOP PARAMETER EVALUATION TABLE (>= 768px — 100% Preserved) -->
      <div class="hidden md:block overflow-x-auto">
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
    <div class="p-3.5 sm:p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs space-y-3">
      <div class="flex flex-col sm:flex-row sm:flex-wrap sm:items-center justify-between gap-2.5">
        <div>
          <h2 class="text-xs sm:text-sm font-bold text-slate-900 dark:text-white">
            Tren Historis Indeks Pencemaran (IP) — {{ trendRange.toUpperCase() }}
          </h2>
          <p class="text-[11px] sm:text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            Garis hijau putus-putus: Baku Mutu (IP = 1.0) · Garis kuning: Batas Cemar Ringan (IP = 5.0).
          </p>
        </div>

        <div class="flex items-center justify-between sm:justify-end gap-1.5 w-full sm:w-auto">
          <button
            v-if="canRecalculate"
            type="button"
            @click="runBackfill"
            :disabled="backfilling"
            class="px-2.5 py-1.5 sm:py-1 rounded-xl sm:rounded-lg text-[11px] sm:text-xs font-semibold border border-slate-300 dark:border-slate-700 hover:border-sky-500 text-slate-600 dark:text-slate-300 shrink-0"
          >
            {{ backfilling ? 'Backfill...' : '⚡ Backfill Riwayat' }}
          </button>
          <div class="grid grid-cols-3 sm:flex items-center gap-0.5 bg-slate-100 dark:bg-slate-800/80 p-0.5 rounded-xl sm:rounded-lg border border-slate-200 dark:border-slate-700 flex-1 sm:flex-initial max-w-[200px] sm:max-w-none">
            <button
              v-for="r in ['24h', '7d', '30d']"
              :key="r"
              type="button"
              @click="selectTrendRange(r)"
              class="px-2.5 py-1 rounded-lg sm:rounded-md text-[11px] sm:text-xs font-bold transition text-center"
              :class="trendRange === r
                ? 'bg-sky-600 text-white shadow-xs'
                : 'text-slate-600 dark:text-slate-300'"
            >
              {{ r.toUpperCase() }}
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="trendDots.length > 0"
        ref="ipChartContainer"
        class="relative w-full h-52 sm:h-56 rounded-xl bg-slate-50/60 dark:bg-[#0B1322] border border-slate-200/70 dark:border-slate-800 overflow-hidden select-none"
        @mousemove="handleIpChartMove"
        @touchstart.passive="handleIpChartTouch"
        @touchmove.passive="handleIpChartTouch"
        @mouseleave="hoveredIpPoint = null"
      >
        <svg :viewBox="`0 0 ${ipChartWidth} 230`" class="w-full h-full block">
          <defs>
            <linearGradient id="wqmsIpAreaGrad" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#0EA5E9" stop-opacity="0.28" />
              <stop offset="100%" stop-color="#10B981" stop-opacity="0.02" />
            </linearGradient>
          </defs>

          <!-- Left Y-axis baseline & labels -->
          <g class="text-[9.5px] font-mono fill-slate-400" text-anchor="end">
            <text x="32" y="213">0</text>
            <text x="32" :y="chartY(1.0) + 3">1.0</text>
            <text x="32" :y="chartY(5.0) + 3">5.0</text>
          </g>

          <!-- Threshold lines -->
          <line x1="38" :y1="chartY(1.0)" :x2="ipChartWidth - 14" :y2="chartY(1.0)" stroke="#10B981" stroke-dasharray="5 4" stroke-width="1.3" />
          <text x="42" :y="Math.max(14, chartY(1.0) - 5)" fill="#10B981" font-size="10" font-weight="700">
            {{ ipChartWidth < 420 ? 'BM (IP = 1.0)' : 'Batas Baku Mutu (IP = 1.0)' }}
          </text>

          <line x1="38" :y1="chartY(5.0)" :x2="ipChartWidth - 14" :y2="chartY(5.0)" stroke="#F59E0B" stroke-dasharray="5 4" stroke-width="1.3" />
          <text x="42" :y="Math.max(14, chartY(5.0) - 5)" fill="#F59E0B" font-size="10" font-weight="700">
            {{ ipChartWidth < 420 ? 'Cemar Ringan (5.0)' : 'Batas Cemar Ringan (IP = 5.0)' }}
          </text>

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

          <!-- Interactive Touch / Hover Crosshair -->
          <g v-if="hoveredIpPoint">
            <line
              :x1="hoveredIpPoint.x"
              y1="14"
              :x2="hoveredIpPoint.x"
              y2="210"
              stroke="#0EA5E9"
              stroke-width="1.2"
              stroke-dasharray="3 3"
            />
            <circle
              :cx="hoveredIpPoint.x"
              :cy="hoveredIpPoint.y"
              r="5.5"
              :fill="hoveredIpPoint.color"
              stroke="#FFFFFF"
              stroke-width="2"
            />
          </g>
        </svg>

        <!-- Floating Adaptive Tooltip -->
        <div
          v-if="hoveredIpPoint"
          class="absolute top-2.5 left-2.5 right-2.5 sm:right-auto sm:max-w-xs z-20 pointer-events-none px-3 py-2 rounded-xl bg-white/95 dark:bg-[#0F172A]/95 border border-slate-200 dark:border-slate-700 shadow-lg backdrop-blur-xs text-[11px] flex items-center justify-between sm:block gap-2"
        >
          <div class="font-mono text-[10px] text-slate-500 dark:text-slate-400 truncate">
            {{ hoveredIpPoint.timeLabel }}
          </div>
          <div class="flex items-center gap-2 shrink-0 sm:mt-0.5">
            <span class="font-extrabold font-mono text-slate-900 dark:text-white">
              IP {{ hoveredIpPoint.ipFormatted }}
            </span>
            <span class="px-1.5 py-0.2 rounded text-[9.5px] font-bold" :style="{ background: hoveredIpPoint.color + '22', color: hoveredIpPoint.color }">
              {{ hoveredIpPoint.categoryLabel }}
            </span>
          </div>
        </div>
      </div>
      <div v-else class="py-10 text-center text-xs text-slate-500 dark:text-slate-400">
        Belum ada titik riwayat tren pada rentang ini. Klik <strong>⚡ Backfill Riwayat</strong> atau <strong>🔄 Hitung</strong> untuk membentuk histori dari tabel agregasi.
      </div>
    </div>

    <!-- ROW 4: HISTORICAL RECORDS SECTION -->
    <div class="p-3.5 sm:p-5 rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-[#1E2F48] shadow-xs space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-xs sm:text-sm font-bold text-slate-900 dark:text-white">Riwayat Perhitungan IP ({{ historyTotal }} Rekaman)</h2>
      </div>

      <!-- MOBILE IP HISTORY CARDS (< 768px) -->
      <div class="md:hidden space-y-2">
        <div v-if="historyRecords.length === 0" class="py-6 text-center text-xs text-slate-400">
          Belum ada rekaman riwayat perhitungan IP.
        </div>
        <div
          v-for="row in historyRecords"
          :key="'mob-ip-row-' + row.id"
          class="p-3 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/60 space-y-2 text-xs"
        >
          <div class="flex items-center justify-between gap-2">
            <span class="font-mono text-[11px] font-bold text-slate-700 dark:text-slate-200 truncate">
              {{ formatDateTime(row.period_end) }}
            </span>
            <span class="px-2 py-0.5 rounded bg-slate-200/70 dark:bg-slate-800 text-[10px] font-bold text-slate-600 dark:text-slate-300 shrink-0">
              {{ formatClassCode(row.water_class) }} · {{ row.calculation_mode === 'OPERATIONAL' ? 'Agregasi' : '24 Jam' }}
            </span>
          </div>
          <div class="flex items-center justify-between gap-2 pt-1.5 border-t border-slate-200/60 dark:border-slate-800/80">
            <div class="flex items-center gap-2.5 min-w-0">
              <span class="text-2xl font-black font-mono tabular-nums text-slate-900 dark:text-white leading-none">
                {{ formatNum(row.overall_ip) }}
              </span>
              <div class="min-w-0">
                <span class="inline-block px-2 py-0.5 rounded text-[10px] font-bold" :class="getHistoryCategoryBadgeClass(row.overall_category)">
                  {{ formatIPCategory(row.overall_category) }}
                </span>
                <div class="text-[10.5px] text-slate-500 dark:text-slate-400 mt-0.5 truncate">
                  Dominan: <strong class="text-slate-800 dark:text-slate-200">{{ row.dominant_parameter || '-' }}</strong>
                </div>
              </div>
            </div>
            <div class="text-right text-[10px] text-slate-500 dark:text-slate-400 shrink-0 font-mono tabular-nums">
              <div>M/R: <strong class="text-slate-700 dark:text-slate-200">{{ formatNum(row.max_ratio) }}/{{ formatNum(row.avg_ratio) }}</strong></div>
              <div>Data: <strong class="text-slate-700 dark:text-slate-200">{{ row.completeness_percent }}%</strong></div>
            </div>
          </div>
        </div>
      </div>

      <!-- DESKTOP HISTORICAL RECORDS TABLE (>= 768px — 100% Preserved) -->
      <div class="hidden md:block overflow-x-auto">
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
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-slate-900/60 backdrop-blur-xs p-0 sm:p-4"
      @click.self="showMappingsModal = false"
    >
      <div class="w-full max-w-3xl rounded-t-2xl sm:rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 shadow-2xl p-4 sm:p-5 space-y-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between gap-2 pb-2 border-b border-slate-200 dark:border-slate-800">
          <div class="min-w-0">
            <h3 class="text-xs sm:text-sm font-bold text-slate-900 dark:text-white">Pemetaan Parameter WQMS &amp; Input Uji Laboratorium</h3>
            <p class="text-[11px] sm:text-xs text-slate-500 dark:text-slate-400">Petakan sensor online atau tambahkan nilai uji lab berkala (misal Total Fosfat / Coliform).</p>
          </div>
          <button type="button" @click="showMappingsModal = false" class="w-7 h-7 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-500 hover:text-slate-800 dark:hover:text-white flex items-center justify-center shrink-0">✕</button>
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
              class="w-full py-2 sm:py-1.5 px-3 rounded-lg bg-sky-600 hover:bg-sky-500 text-white font-semibold disabled:opacity-50"
            >
              {{ savingMapping ? '...' : '+ Simpan' }}
            </button>
          </div>
        </div>

        <!-- Mobile Mappings Cards (< 640px) -->
        <div class="sm:hidden space-y-2">
          <div v-if="mappingsList.length === 0" class="py-4 text-center text-xs text-slate-400">
            Belum ada pemetaan parameter.
          </div>
          <div
            v-for="m in mappingsList"
            :key="'mob-wmap-' + m.id"
            class="p-2.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/60 flex items-center justify-between gap-2 text-xs"
          >
            <div class="min-w-0">
              <div class="font-bold text-slate-900 dark:text-white truncate">{{ formatParamFullName(m.param_code) }}</div>
              <div class="text-[10.5px] text-slate-500 dark:text-slate-400 truncate mt-0.5">
                <span class="font-semibold">{{ m.source_type === 'LAB' ? 'Uji Lab:' : 'Sensor:' }}</span>
                <template v-if="m.source_type === 'LAB'">
                  <strong class="ml-1 text-slate-800 dark:text-slate-200">{{ m.lab_value }}</strong> {{ m.normalized_unit }}
                </template>
                <template v-else>
                  <span class="ml-1">{{ getMappingSensorLabel(m) }}</span>
                  <span
                    v-if="!isOperator && m.parameter && m.parameter.parameter_code && m.parameter.parameter_code !== m.parameter.parameter_name"
                    class="text-slate-400 ml-1"
                  >
                    ({{ m.parameter.parameter_code }})
                  </span>
                </template>
              </div>
            </div>
            <button type="button" @click="removeMapping(m.id)" class="px-2.5 py-1 rounded-lg bg-rose-500/10 text-rose-500 font-bold text-[11px] shrink-0">
              Hapus
            </button>
          </div>
        </div>

        <!-- Desktop Mappings Table (>= 640px) -->
        <table class="hidden sm:table w-full text-xs border-collapse">
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
      ipChartWidth: 960,
      hoveredIpPoint: null,
      ipChartResizeObserver: null,
    };
  },
  computed: {
    mobileActionGridCols() {
      let cols = 2; // CSV + Excel
      if (this.canManageMapping) cols++;
      if (this.canRecalculate) cols++;
      if (cols === 4) return 'grid-cols-4';
      if (cols === 3) return 'grid-cols-3';
      return 'grid-cols-2';
    },
    ipCategoryCleanLabel() {
      if (!this.hasValidIP) return 'Tidak Tersedia';
      switch (this.activeIP.overall_category) {
        case 'MEMENUHI_BAKU_MUTU': return 'Memenuhi Baku Mutu';
        case 'CEMAR_RINGAN': return 'Cemar Ringan';
        case 'CEMAR_SEDANG': return 'Cemar Sedang';
        case 'CEMAR_BERAT': return 'Cemar Berat';
        default: return 'Tidak Tersedia';
      }
    },
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
      const padL = 42;
      const padR = 18;
      const plotW = Math.max(80, this.ipChartWidth - padL - padR);
      return this.trendRecords.map((r, i) => {
        const x = n === 1 ? padL + plotW / 2 : padL + (i / (n - 1)) * plotW;
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
          ipFormatted: val.toFixed(2),
          categoryLabel: this.formatIPCategory(r.overall_category),
          timeLabel: this.formatDateTime(r.period_end),
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
    this.$nextTick(() => this.observeIpChartWidth());
    window.addEventListener('resize', this.updateIpChartWidth);
    window.addEventListener('wqms-calculated-event', this.onWqmsCalculatedEvent);
    window.addEventListener('tenant-context-changed', this.onTenantContextChanged);
  },
  beforeDestroy() {
    if (this.ipChartResizeObserver) {
      this.ipChartResizeObserver.disconnect();
    }
    window.removeEventListener('resize', this.updateIpChartWidth);
    window.removeEventListener('wqms-calculated-event', this.onWqmsCalculatedEvent);
    window.removeEventListener('tenant-context-changed', this.onTenantContextChanged);
  },
  watch: {
    'trendDots.length'() {
      this.$nextTick(() => this.observeIpChartWidth());
    },
  },
  methods: {
    observeIpChartWidth() {
      this.updateIpChartWidth();
      if (typeof window !== 'undefined' && window.ResizeObserver && this.$refs.ipChartContainer) {
        if (this.ipChartResizeObserver) this.ipChartResizeObserver.disconnect();
        this.ipChartResizeObserver = new ResizeObserver(() => {
          this.updateIpChartWidth();
        });
        this.ipChartResizeObserver.observe(this.$refs.ipChartContainer);
      }
    },
    updateIpChartWidth() {
      const el = this.$refs.ipChartContainer;
      if (el) {
        const w = Math.round(el.getBoundingClientRect().width);
        if (w > 120) this.ipChartWidth = w;
      }
    },
    handleIpChartTouch(e) {
      if (!e.touches || e.touches.length === 0) return;
      const t = e.touches[0];
      this.handleIpChartMove({ clientX: t.clientX });
    },
    handleIpChartMove(e) {
      if (!this.trendDots || this.trendDots.length === 0 || !this.$refs.ipChartContainer) return;
      const rect = this.$refs.ipChartContainer.getBoundingClientRect();
      const relX = e.clientX - rect.left;
      let closest = this.trendDots[0];
      let minDiff = Math.abs(closest.x - relX);
      for (let i = 1; i < this.trendDots.length; i++) {
        const d = Math.abs(this.trendDots[i].x - relX);
        if (d < minDiff) {
          minDiff = d;
          closest = this.trendDots[i];
        }
      }
      this.hoveredIpPoint = closest;
    },
    getHistoryCategoryBadgeClass(cat) {
      switch (cat) {
        case 'MEMENUHI_BAKU_MUTU': return 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30';
        case 'CEMAR_RINGAN': return 'bg-amber-400/20 text-amber-800 dark:text-amber-300 border border-amber-500/35';
        case 'CEMAR_SEDANG': return 'bg-orange-500/20 text-orange-800 dark:text-orange-300 border border-orange-500/35';
        case 'CEMAR_BERAT': return 'bg-rose-500/20 text-rose-700 dark:text-rose-300 border border-rose-500/35';
        default: return 'bg-slate-200 dark:bg-slate-800 text-slate-500';
      }
    },
    onTenantContextChanged() {
      this.loadAll();
      this.fetchDevices();
    },
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
