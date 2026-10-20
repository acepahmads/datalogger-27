<template>
  <div class="relative w-full flex flex-col space-y-2">
    <!-- Chart Header & Time Range Controls (Reference Design Target) -->
    <div class="flex flex-wrap items-start justify-between gap-3 pb-1">
      <div>
        <div class="flex items-center space-x-2">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white tracking-wide uppercase font-sans">
            {{ displayTitle }}
          </h3>
          <span
            class="px-1.5 py-0.5 rounded text-[9px] font-mono font-bold uppercase border"
            :class="isRegulatoryMode
              ? 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border-cyan-500/30'
              : 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border-blue-500/30'"
          >
            {{ isRegulatoryMode ? '24h Rolling' : 'Operational' }}
          </span>
        </div>
        <div
          @click="toggleBandOrientation"
          class="text-xs text-slate-500 dark:text-slate-400 font-normal mt-0.5 font-sans cursor-pointer hover:underline"
          title="Click to toggle between Regulatory Cartesian Scale and Reference Mockup Layout"
        >
          {{ displaySubtitle }}
          <span v-if="!isOverallMode" class="text-cyan-600 dark:text-cyan-400 font-medium ml-1">
            · {{ activePollutantCode }} ({{ activePollutantName || 'Pollutant Trend' }})
          </span>
        </div>
      </div>


      <!-- Right Side Controls: Parameter Selector, Capsule Range Switcher & Close Icon -->
      <div class="flex flex-wrap items-center gap-2">
        <!-- Parameter Selector Capsule Buttons (★ Dominant | PM2.5 | PM10 | SO2 | CO | O3 | NO2 | HC) -->
        <div class="flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B] shadow-xs overflow-x-auto no-scrollbar">
          <button
            v-for="p in availablePollutantOptions"
            :key="p.code"
            @click="selectPollutant(p.code)"
            class="px-2 py-1 rounded-md text-xs font-semibold transition cursor-pointer select-none whitespace-nowrap"
            :class="activePollutantCode === p.code
              ? 'bg-cyan-500 text-slate-950 font-bold shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
            :title="`Tampilkan grafik tren ${p.label}`"
          >
            {{ p.label }}
          </button>
        </div>

        <!-- Time Range Selector Capsule Buttons (24 Hours | 7 Days | 30 Days) -->
        <div class="flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B] shadow-xs">
          <button
            v-for="r in rangeOptions"
            :key="r.value"
            @click="selectRange(r.value)"
            class="px-2.5 py-1 rounded-md text-xs font-bold transition cursor-pointer select-none"
            :class="currentRange === r.value
              ? 'bg-cyan-500 text-slate-950 shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            {{ r.label }}
          </button>
        </div>

        <!-- Subtle Close / Reset Hover Button (Matching Reference Icon) -->
        <button
          @click="resetHover"
          class="p-1 rounded text-slate-400 hover:text-slate-700 dark:hover:text-white transition cursor-pointer"
          title="Reset hover selection"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Chart Canvas Container (Theme adaptive: Light Mode #F8FCFF border #B7D8EC, Dark Mode #0B1322) -->
    <div
      ref="chartContainer"
      class="relative w-full h-64 sm:h-72 rounded-xl overflow-hidden select-none transition-colors duration-200 shadow-md"
      :class="resolvedIsDark
        ? 'bg-[#0B1322] border border-slate-700/60'
        : 'bg-[#F8FCFF] border border-[#B7D8EC]'"
      @mousemove="handleMouseMove"
      @mouseleave="hoveredPoint = null"
    >
      <!-- Loading State -->
      <div
        v-if="loading"
        class="absolute inset-0 z-20 flex items-center justify-center backdrop-blur-xs transition-colors"
        :class="resolvedIsDark ? 'bg-slate-900/60' : 'bg-white/60'"
      >
        <div
          class="flex items-center space-x-2 text-xs font-mono px-3 py-1.5 rounded-lg border"
          :class="resolvedIsDark
            ? 'text-cyan-400 bg-slate-900/90 border-cyan-500/30'
            : 'text-cyan-700 bg-white/95 border-[#B7D8EC] shadow-sm'"
        >
          <div
            class="w-4 h-4 border-2 border-t-transparent rounded-full animate-spin"
            :class="resolvedIsDark ? 'border-cyan-400' : 'border-cyan-600'"
          ></div>
          <span>Loading historical trend...</span>
        </div>
      </div>

      <!-- Empty State -->
      <div
        v-else-if="!points || points.length === 0"
        class="absolute inset-0 flex flex-col items-center justify-center text-xs font-mono space-y-1"
        :class="resolvedIsDark ? 'text-slate-400' : 'text-slate-500'"
      >
        <span class="font-medium">No historical ISPU records available for selected range</span>
        <span class="text-3xs" :class="resolvedIsDark ? 'text-slate-500' : 'text-slate-400'">
          Historical trend will populate automatically from aggregation rollups
        </span>
      </div>

      <!-- SVG Chart -->
      <svg
        v-else
        class="w-full h-full block"
        :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
      >
        <defs>
          <!-- Neon Glow Filter for Bright Cyan Curve & Points -->
          <filter id="neonGlow" x="-30%" y="-30%" width="160%" height="160%">
            <feDropShadow
              dx="0"
              dy="0"
              stdDeviation="3.5"
              flood-color="#00F2FE"
              :flood-opacity="resolvedIsDark ? '0.9' : '0.0'"
            />
          </filter>

          <!-- 5 Category Bands (Theme Adaptive: Rich Saturated SCADA Colors in both themes) -->
          <linearGradient id="bandBaik" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="resolvedIsDark ? '#1B873F' : '#22C55E'" :stop-opacity="resolvedIsDark ? '0.92' : '0.90'" />
            <stop offset="100%" :stop-color="resolvedIsDark ? '#166534' : '#15803D'" :stop-opacity="resolvedIsDark ? '0.92' : '0.90'" />
          </linearGradient>

          <linearGradient id="bandSedang" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="resolvedIsDark ? '#1D63B8' : '#3B82F6'" :stop-opacity="resolvedIsDark ? '0.92' : '0.90'" />
            <stop offset="100%" :stop-color="resolvedIsDark ? '#1E40AF' : '#1D4ED8'" :stop-opacity="resolvedIsDark ? '0.92' : '0.90'" />
          </linearGradient>

          <linearGradient id="bandTidakSehat" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="resolvedIsDark ? '#C49000' : '#EAB308'" :stop-opacity="resolvedIsDark ? '0.92' : '0.90'" />
            <stop offset="100%" :stop-color="resolvedIsDark ? '#B45309' : '#CA8A04'" :stop-opacity="resolvedIsDark ? '0.92' : '0.90'" />
          </linearGradient>

          <linearGradient id="bandSangatTidakSehat" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="resolvedIsDark ? '#B91C1C' : '#EF4444'" :stop-opacity="resolvedIsDark ? '0.92' : '0.90'" />
            <stop offset="100%" :stop-color="resolvedIsDark ? '#991B1B' : '#B91C1C'" :stop-opacity="resolvedIsDark ? '0.92' : '0.90'" />
          </linearGradient>

          <linearGradient id="bandBerbahaya" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="resolvedIsDark ? '#4A0D15' : '#881337'" :stop-opacity="resolvedIsDark ? '0.95' : '0.92'" />
            <stop offset="100%" :stop-color="resolvedIsDark ? '#3B080F' : '#5C0820'" :stop-opacity="resolvedIsDark ? '0.95' : '0.92'" />
          </linearGradient>

          <linearGradient id="trendAreaGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="resolvedIsDark ? '#00F2FE' : '#0891B2'" stop-opacity="0.30" />
            <stop offset="100%" :stop-color="resolvedIsDark ? '#00F2FE' : '#0891B2'" stop-opacity="0.0" />
          </linearGradient>
        </defs>

        <!-- Plot Area Base Background (#FFFFFF in Light Mode, #080F20 in Dark Mode) -->
        <rect
          :x="padLeft"
          :y="padTop"
          :width="chartWidth - padLeft"
          :height="plotHeight"
          :fill="resolvedIsDark ? '#080F20' : '#FFFFFF'"
        />

        <!-- 5 Horizontal Background Category Bands -->
        <g class="category-bands">
          <rect
            v-for="b in computedBands"
            :key="b.key"
            :id="b.id"
            :x="padLeft"
            :y="b.y"
            :width="chartWidth - padLeft"
            :height="b.height"
            :fill="`url(#${b.id})`"
          />
        </g>

        <!-- Category Horizontal Boundary Lines -->
        <g :stroke="resolvedIsDark ? 'rgba(0, 0, 0, 0.45)' : 'rgba(0, 0, 0, 0.12)'" stroke-width="1.2">
          <line
            v-for="(b, idx) in computedBands.slice(0, 4)"
            :key="`line-${idx}`"
            :x1="padLeft"
            :y1="b.y + b.height"
            :x2="chartWidth"
            :y2="b.y + b.height"
          />
        </g>

        <!-- Category Labels Inside the Left Side of Each Band (High-Contrast White with Shadow) -->
        <g class="band-labels text-[11px] font-mono font-black select-none">
          <text
            v-for="b in computedBands"
            :key="`lbl-${b.key}`"
            v-if="b.height >= 14"
            :x="padLeft + 12"
            :y="b.y + b.height / 2 + 4"
            fill="#FFFFFF"
            letter-spacing="0.6"
            style="text-shadow: 0 1px 3px rgba(0,0,0,0.85); font-weight: 900;"
          >
            {{ b.label }}
          </text>
        </g>

        <!-- Left Y-Axis Numeric Labels (Theme Adaptive: #526779 in Light Mode, #94A3B8 in Dark Mode) -->
        <g class="text-[10px] font-mono select-none" text-anchor="end" :fill="resolvedIsDark ? '#94A3B8' : '#526779'">
          <text
            v-for="tick in yTicks"
            :key="`ytick-${tick.val}`"
            :x="padLeft - 6"
            :y="tick.y + 3.5"
          >
            {{ tick.val }}
          </text>
        </g>

        <!-- Bottom X-Axis Date/Time Labels (Theme Adaptive: #526779 in Light Mode, #94A3B8 in Dark Mode) -->
        <g class="text-[10px] font-mono select-none" text-anchor="middle" :fill="resolvedIsDark ? '#94A3B8' : '#526779'">
          <text
            v-for="(tick, idx) in xTicks"
            :key="`xtick-${idx}`"
            :x="tick.x"
            :y="padTop + plotHeight + 16"
          >
            {{ tick.label }}
          </text>
        </g>

        <!-- Underlay Halo for Curve to Pop Sharply Over Colored Category Bands -->
        <path
          :d="curvePath"
          fill="none"
          :stroke="resolvedIsDark ? 'rgba(8, 15, 32, 0.85)' : 'rgba(255, 255, 255, 0.95)'"
          :stroke-width="resolvedIsDark ? 4.5 : 4.8"
          stroke-linecap="round"
          stroke-linejoin="round"
          opacity="0.9"
        />

        <!-- Smooth Bright Cyan / Deep Teal Trend Curve -->
        <path
          :d="curvePath"
          fill="none"
          :stroke="resolvedIsDark ? '#00F2FE' : '#0891B2'"
          stroke-width="2.6"
          stroke-linecap="round"
          stroke-linejoin="round"
          filter="url(#neonGlow)"
        />

        <!-- Circular Data Point Markers -->
        <g class="data-markers">
          <g v-for="(p, idx) in points" :key="`pt-${idx}`">
            <template v-if="p.isAvailable && p.y !== null && shouldShowMarker(idx)">
              <!-- Ripple Effect on the Latest Point -->
              <circle
                v-if="idx === points.length - 1"
                :cx="p.x"
                :cy="p.y"
                r="7"
                :fill="resolvedIsDark ? '#00F2FE' : '#0891B2'"
                class="animate-ping opacity-35"
              />
              <!-- Core Circular Marker -->
              <circle
                :cx="p.x"
                :cy="p.y"
                r="3.5"
                :fill="resolvedIsDark ? '#00F2FE' : '#0891B2'"
                :stroke="resolvedIsDark ? '#080F20' : '#FFFFFF'"
                stroke-width="1.8"
                filter="url(#neonGlow)"
                class="transition-transform hover:scale-175 cursor-pointer shadow-xs"
              />
            </template>
          </g>
        </g>

        <!-- Hover Guideline & Crosshair Indicator -->
        <g v-if="hoveredPoint">
          <line
            :x1="hoveredPoint.x"
            :y1="padTop"
            :x2="hoveredPoint.x"
            :y2="padTop + plotHeight"
            :stroke="resolvedIsDark ? '#00F2FE' : '#0891B2'"
            stroke-width="1.2"
            stroke-dasharray="3,3"
            class="opacity-80"
          />
          <circle
            v-if="hoveredPoint.isAvailable && hoveredPoint.y !== null"
            :cx="hoveredPoint.x"
            :cy="hoveredPoint.y"
            r="6"
            :fill="resolvedIsDark ? '#00F2FE' : '#0891B2'"
            stroke="#FFFFFF"
            stroke-width="2"
            filter="url(#neonGlow)"
          />
        </g>
      </svg>

      <!-- Compact Adaptive Tooltip (Theme Adaptive: White in Light Mode, Slate in Dark Mode) -->
      <div
        v-if="hoveredPoint"
        class="absolute z-30 pointer-events-none p-3 rounded-xl backdrop-blur-md text-[11px] font-mono transform -translate-x-1/2 -translate-y-full transition-transform duration-75 select-none"
        :class="resolvedIsDark
          ? 'bg-[#0B1C2B]/95 text-white border border-[#244356] shadow-2xl'
          : 'bg-white/95 text-slate-800 border border-[#B7D8EC] shadow-xl'"
        :style="{
          left: `${tooltipLeft}px`,
          top: `${tooltipTop}px`
        }"
      >
        <div class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1.5 whitespace-nowrap">
          <span :class="resolvedIsDark ? 'text-slate-400' : 'text-[#526779]'">Mode:</span>
          <span class="font-bold" :class="resolvedIsDark ? 'text-cyan-400' : 'text-cyan-600'">
            {{ isRegulatoryMode ? '24-Jam Rolling Regulasi' : `Per Agregasi (${aggIntervalLabel})` }}
          </span>

          <span :class="resolvedIsDark ? 'text-slate-400' : 'text-[#526779]'">Window:</span>
          <span class="font-bold" :class="resolvedIsDark ? 'text-white' : 'text-slate-900'">
            {{ formatWindow(hoveredPoint.raw) }}
          </span>

          <span :class="resolvedIsDark ? 'text-slate-400' : 'text-[#526779]'">ISPU:</span>
          <span class="font-bold" :class="resolvedIsDark ? 'text-[#00F2FE]' : 'text-cyan-700'">
            <template v-if="hoveredPoint.isAvailable && hoveredPoint.value !== null">
              {{ Math.round(hoveredPoint.value) }} ({{ (hoveredPoint.category || '').replace('_', ' ') }})
            </template>
            <template v-else>
              <span :class="resolvedIsDark ? 'text-slate-400 font-normal' : 'text-slate-500 font-normal'">Tidak Tersedia / Missing</span>
            </template>
          </span>

          <span v-if="!isOverallMode" :class="resolvedIsDark ? 'text-slate-400' : 'text-[#526779]'">Concentration:</span>
          <span v-if="!isOverallMode" class="font-bold" :class="resolvedIsDark ? 'text-white' : 'text-slate-900'">
            <template v-if="hoveredPoint.isAvailable && hoveredPoint.concentration != null">
              {{ hoveredPoint.concentration.toFixed(1) }} {{ hoveredPoint.unit }}
            </template>
            <template v-else>
              <span :class="resolvedIsDark ? 'text-slate-400' : 'text-slate-500'">-</span>
            </template>
          </span>

          <span :class="resolvedIsDark ? 'text-slate-400' : 'text-[#526779]'">Dominant:</span>
          <span class="font-bold" :class="resolvedIsDark ? 'text-white' : 'text-slate-900'">
            {{ formatDominant(hoveredPoint.raw.dominant_pollutant || selectedPollutant) }}
          </span>

          <span :class="resolvedIsDark ? 'text-slate-400' : 'text-[#526779]'">Coverage:</span>
          <span class="font-bold" :class="resolvedIsDark ? 'text-white' : 'text-slate-900'">
            {{ Number(hoveredPoint.raw.completeness_percent || 0).toFixed(1) }}%
            <span class="text-4xs" :class="resolvedIsDark ? 'text-slate-400' : 'text-slate-500'">({{ isRegulatoryMode ? '≥75% valid' : 'bucket' }})</span>
          </span>

          <span :class="resolvedIsDark ? 'text-slate-400' : 'text-[#526779]'">Status:</span>
          <span
            class="font-bold px-1.5 py-0.5 rounded text-[10px]"
            :class="getStatusBadgeClass(hoveredPoint.raw.calculation_status)"
          >
            {{ hoveredPoint.raw.calculation_status || (isRegulatoryMode ? 'PROVISIONAL' : 'VALID') }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ISPUTrendChart',
  props: {
    title: { type: String, default: '' },
    records: { type: Array, default: () => [] },
    loading: { type: Boolean, default: false },
    initialRange: { type: String, default: '24h' },
    selectedPollutant: { type: String, default: 'OVERALL' },
    pollutantName: { type: String, default: '' },
    mode: { type: String, default: 'REGULATORY' },
    aggregationIntervalSeconds: { type: Number, default: 120 },
    isDark: { type: Boolean, default: null }
  },
  data() {
    return {
      currentRange: this.initialRange,
      internalSelectedPollutant: this.selectedPollutant || 'OVERALL',
      // Default to standard regulatory Cartesian scale; toggleable to reference mockup layout
      invertBands: false,
      isAdaptiveScale: false,
      domThemeIsDark: typeof document !== 'undefined' ? document.documentElement.classList.contains('dark') : true,
      rangeOptions: [
        { label: '24 Hours', value: '24h' },
        { label: '7 Days', value: '7d' },
        { label: '30 Days', value: '30d' }
      ],
      chartWidth: 700,
      chartHeight: 260,
      padTop: 14,
      padBottom: 28,
      padLeft: 38,
      padRight: 20,
      hoveredPoint: null,
      resizeObserver: null,
      themeObserver: null
    };
  },
  watch: {
    initialRange(newVal) {
      if (newVal) this.currentRange = newVal;
    },
    selectedPollutant(newVal) {
      if (newVal) this.internalSelectedPollutant = newVal;
    }
  },
  computed: {
    resolvedIsDark() {
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
      return this.domThemeIsDark;
    },
    isRegulatoryMode() {
      return this.mode === 'REGULATORY';
    },
    aggIntervalLabel() {
      if (this.aggregationIntervalSeconds % 60 === 0) {
        return `${this.aggregationIntervalSeconds / 60} Menit`;
      }
      return `${this.aggregationIntervalSeconds}s`;
    },
    activePollutantCode() {
      return this.internalSelectedPollutant || this.selectedPollutant || 'OVERALL';
    },
    availablePollutantOptions() {
      return [
        { code: 'OVERALL', label: '★ Dominant' },
        { code: 'PM25', label: 'PM2.5' },
        { code: 'PM10', label: 'PM10' },
        { code: 'SO2', label: 'SO2' },
        { code: 'CO', label: 'CO' },
        { code: 'O3', label: 'O3' },
        { code: 'NO2', label: 'NO2' },
        { code: 'HC', label: 'HC' }
      ];
    },
    activePollutantName() {
      if (this.pollutantName && this.activePollutantCode === this.selectedPollutant) {
        return this.pollutantName;
      }
      const names = {
        PM25: 'Partikulat PM2.5 (24h)',
        PM10: 'Partikulat PM10 (24h)',
        SO2: 'Sulfur Dioksida (24h)',
        CO: 'Karbon Monoksida (24h)',
        O3: 'Ozon (24h)',
        NO2: 'Nitrogen Dioksida (24h)',
        HC: 'Hidrokarbon (24h)'
      };
      return names[this.activePollutantCode] || '';
    },
    isOverallMode() {
      return !this.activePollutantCode || this.activePollutantCode === 'OVERALL';
    },
    displayTitle() {
      if (this.title) return this.title;
      if (this.mode === 'OPERATIONAL') {
        return `ISPU HISTORICAL TREND — PER AGREGASI (${this.aggIntervalLabel})`;
      }
      return 'ISPU HISTORICAL TREND — 24-JAM (REGULASI)';
    },
    displaySubtitle() {
      if (this.mode === 'OPERATIONAL') {
        return `Resolusi Bucket Operasional (${this.aggIntervalLabel}) · Nilai Agregasi Real-time`;
      }
      return 'Permen LHK No. 14 Tahun 2020 · Jendela Pengukuran 24 Jam (Cadence: 1 Jam)';
    },
    plotWidth() {
      return Math.max(10, this.chartWidth - this.padLeft - this.padRight);
    },
    plotHeight() {
      return Math.max(10, this.chartHeight - this.padTop - this.padBottom);
    },
    maxY() {
      if (!this.records || this.records.length === 0) return 350;
      let maxVal = 0;
      for (const p of this.extractedRawPoints) {
        if (p.value > maxVal) maxVal = p.value;
      }
      // Always support values above 300 (full ISPU range)
      return Math.max(350, Math.ceil((maxVal + 25) / 50) * 50);
    },
    computedBands() {
      // 5 regulatory bands: BAIK (0-50), SEDANG (51-100), TIDAK SEHAT (101-200), SANGAT TIDAK SEHAT (201-300), BERBAHAYA (>300)
      if (this.invertBands) {
        // Reference Mockup visual layout: Green at top, Blue 2nd, Yellow 3rd, Red 4th, Dark Red at bottom
        const h5 = this.plotHeight / 5;
        return [
          { key: 'baik', id: 'bandBaik', label: 'BAIK (0-50)', colorLight: '#064E3B', textColorLight: '#14532D', y: this.padTop, height: h5 },
          { key: 'sedang', id: 'bandSedang', label: 'SEDANG (51-100)', colorLight: '#1E3A8A', textColorLight: '#1E40AF', y: this.padTop + h5, height: h5 },
          { key: 'tidak_sehat', id: 'bandTidakSehat', label: 'TIDAK SEHAT (101-200)', colorLight: '#78350F', textColorLight: '#854D0E', y: this.padTop + 2 * h5, height: h5 },
          { key: 'sangat_tidak_sehat', id: 'bandSangatTidakSehat', label: 'SANGAT TIDAK SEHAT (201-300)', colorLight: '#7F1D1D', textColorLight: '#991B1B', y: this.padTop + 3 * h5, height: h5 },
          { key: 'berbahaya', id: 'bandBerbahaya', label: 'BERBAHAYA (>300)', colorLight: '#450A0A', textColorLight: '#701A28', y: this.padTop + 4 * h5, height: h5 }
        ];
      }

      // Standard regulatory Cartesian layout (0 at bottom, 350 at top)
      const y0 = this.getY(0);
      const y50 = this.getY(50);
      const y100 = this.getY(100);
      const y200 = this.getY(200);
      const y300 = this.getY(300);
      const yTop = this.padTop;

      return [
        {
          key: 'berbahaya',
          id: 'bandBerbahaya',
          label: 'BERBAHAYA (>300)',
          colorLight: '#450A0A',
          textColorLight: '#701A28',
          y: yTop,
          height: Math.max(0, y300 - yTop)
        },
        {
          key: 'sangat_tidak_sehat',
          id: 'bandSangatTidakSehat',
          label: 'SANGAT TIDAK SEHAT (201-300)',
          colorLight: '#7F1D1D',
          textColorLight: '#991B1B',
          y: y300,
          height: Math.max(0, y200 - y300)
        },
        {
          key: 'tidak_sehat',
          id: 'bandTidakSehat',
          label: 'TIDAK SEHAT (101-200)',
          colorLight: '#78350F',
          textColorLight: '#854D0E',
          y: y200,
          height: Math.max(0, y100 - y200)
        },
        {
          key: 'sedang',
          id: 'bandSedang',
          label: 'SEDANG (51-100)',
          colorLight: '#1E3A8A',
          textColorLight: '#1E40AF',
          y: y100,
          height: Math.max(0, y50 - y100)
        },
        {
          key: 'baik',
          id: 'bandBaik',
          label: 'BAIK (0-50)',
          colorLight: '#064E3B',
          textColorLight: '#14532D',
          y: y50,
          height: Math.max(0, y0 - y50)
        }
      ];
    },
    yTicks() {
      if (this.invertBands) {
        return [
          { val: 0, y: this.padTop },
          { val: 50, y: this.padTop + this.plotHeight * 0.2 },
          { val: 100, y: this.padTop + this.plotHeight * 0.4 },
          { val: 200, y: this.padTop + this.plotHeight * 0.6 },
          { val: 300, y: this.padTop + this.plotHeight * 0.8 },
          { val: this.maxY, y: this.padTop + this.plotHeight }
        ];
      }

      return [
        { val: 0, y: this.getY(0) },
        { val: 50, y: this.getY(50) },
        { val: 100, y: this.getY(100) },
        { val: 200, y: this.getY(200) },
        { val: 300, y: this.getY(300) },
        { val: this.maxY, y: this.padTop }
      ];
    },
    extractedRawPoints() {
      if (!this.records || this.records.length === 0) return [];

      let sourceRecords = this.records;
      if (this.isRegulatoryMode) {
        // Permen LHK 14/2020: Regulatory trend strictly represents hourly checkpoints.
        // Filter out any 2-minute or sub-hourly points so they never clutter the 24-hour chart.
        const hourlyRecords = sourceRecords.filter(r => new Date(r.period_end).getMinutes() === 0);
        if (hourlyRecords.length > 0) {
          sourceRecords = hourlyRecords;
        }
      }

      return sourceRecords.map(r => {
        if (this.isOverallMode) {
          const isAvail = r.overall_category !== 'TIDAK_TERSEDIA' && r.calculation_status !== 'INVALID';
          return {
            raw: r,
            isAvailable: isAvail,
            value: isAvail ? Number(r.overall_ispu || 0) : null,
            category: r.overall_category || 'TIDAK_TERSEDIA',
            concentration: 0,
            unit: 'µg/m³'
          };
        }
        const pr = (r.parameter_results || []).find(p => p.pollutant_code === this.activePollutantCode);
        if (pr) {
          const isAvail = pr.category !== 'TIDAK_TERSEDIA' &&
                          pr.category !== 'UNAVAILABLE' &&
                          (pr.valid_samples === undefined || pr.valid_samples > 0);
          return {
            raw: r,
            isAvailable: isAvail,
            value: isAvail ? Number(pr.calculated_ispu || 0) : null,
            category: pr.category || 'TIDAK_TERSEDIA',
            concentration: Number(pr.normalized_concentration || 0),
            unit: pr.normalized_unit || 'µg/m³'
          };
        }
        return {
          raw: r,
          isAvailable: false,
          value: null,
          category: 'TIDAK_TERSEDIA',
          concentration: 0,
          unit: 'µg/m³'
        };
      });
    },
    timeWindow() {
      let durationMs = 24 * 3600 * 1000;
      if (this.currentRange === '7d') durationMs = 7 * 86400 * 1000;
      else if (this.currentRange === '30d') durationMs = 30 * 86400 * 1000;

      let endMs = Date.now();
      if (this.records && this.records.length > 0) {
        const lastRecTime = new Date(this.records[this.records.length - 1].period_end).getTime();
        if (!isNaN(lastRecTime)) {
          endMs = Math.max(endMs, lastRecTime);
        }
      }
      const startMs = endMs - durationMs;
      return { startMs, endMs, durationMs };
    },
    points() {
      const raw = this.extractedRawPoints;
      if (!raw || raw.length === 0) return [];
      const { startMs, durationMs } = this.timeWindow;

      return raw.map((item) => {
        const timeMs = new Date(item.raw.period_end).getTime();
        let fraction = 0.5;
        if (durationMs > 0 && !isNaN(timeMs)) {
          fraction = (timeMs - startMs) / durationMs;
        }
        const clampedFraction = Math.max(0, Math.min(1, fraction));
        const x = this.padLeft + clampedFraction * this.plotWidth;
        const y = item.isAvailable && item.value !== null ? this.getY(item.value) : null;
        return {
          x,
          y,
          timeMs,
          isAvailable: item.isAvailable,
          value: item.value,
          category: item.category,
          concentration: item.concentration,
          unit: item.unit,
          raw: item.raw
        };
      });
    },
    curvePath() {
      if (!this.points || this.points.length <= 1) return '';

      // Break path across gaps where missing data exists:
      // In regulatory mode: for bi-daily (09:00 & 15:00) max gap is 19h (from 15:00 to 09:00 is 18h)
      // For hourly mode: gap > 2.5 hours
      // In operational mode: gap > 3 buckets
      const pCode = (this.selectedPollutant || '').toUpperCase();
      const isDailyCadence = this.isRegulatoryMode && !this.isOverallMode && pCode !== 'PM25';
      const maxGapMs = this.isRegulatoryMode
        ? (isDailyCadence ? 19 * 3600 * 1000 : 2.5 * 3600 * 1000)
        : Math.max(180, 3 * this.aggregationIntervalSeconds) * 1000;

      const segments = [];
      let currentSeg = [];

      for (let i = 0; i < this.points.length; i++) {
        const pt = this.points[i];
        if (!pt.isAvailable || pt.y === null) {
          if (currentSeg.length > 0) {
            segments.push(currentSeg);
            currentSeg = [];
          }
          continue;
        }
        if (currentSeg.length > 0) {
          const lastPt = currentSeg[currentSeg.length - 1];
          const gap = pt.timeMs - lastPt.timeMs;
          if (gap > maxGapMs) {
            segments.push(currentSeg);
            currentSeg = [];
          }
        }
        currentSeg.push(pt);
      }
      if (currentSeg.length > 0) segments.push(currentSeg);

      return segments
        .filter(seg => seg.length >= 2)
        .map(seg => this.generateSmoothPath(seg))
        .join(' ');
    },
    xTicks() {
      const { startMs, endMs, durationMs } = this.timeWindow;
      if (!durationMs) return [];

      const ticks = [];
      if (this.currentRange === '24h') {
        const stepMs = 4 * 3600 * 1000;
        const firstTickTime = Math.ceil(startMs / stepMs) * stepMs;
        for (let t = firstTickTime; t <= endMs; t += stepMs) {
          const frac = (t - startMs) / durationMs;
          if (frac >= 0 && frac <= 1) {
            ticks.push({
              x: this.padLeft + frac * this.plotWidth,
              label: this.formatTimeOnly(t)
            });
          }
        }
      } else if (this.currentRange === '7d') {
        const stepMs = 24 * 3600 * 1000;
        const firstTickTime = Math.ceil(startMs / stepMs) * stepMs;
        for (let t = firstTickTime; t <= endMs; t += stepMs) {
          const frac = (t - startMs) / durationMs;
          if (frac >= 0 && frac <= 1) {
            ticks.push({
              x: this.padLeft + frac * this.plotWidth,
              label: this.formatDateOnly(t)
            });
          }
        }
      } else {
        const stepMs = 5 * 24 * 3600 * 1000;
        const firstTickTime = Math.ceil(startMs / stepMs) * stepMs;
        for (let t = firstTickTime; t <= endMs; t += stepMs) {
          const frac = (t - startMs) / durationMs;
          if (frac >= 0 && frac <= 1) {
            ticks.push({
              x: this.padLeft + frac * this.plotWidth,
              label: this.formatDateOnly(t)
            });
          }
        }
      }
      return ticks;
    },
    tooltipLeft() {
      if (!this.hoveredPoint) return 0;
      const minX = 115;
      const maxX = Math.max(minX, this.chartWidth - 115);
      return Math.max(minX, Math.min(this.hoveredPoint.clientX, maxX));
    },
    tooltipTop() {
      if (!this.hoveredPoint) return 0;
      const cy = (this.hoveredPoint.clientY != null) ? this.hoveredPoint.clientY : (this.padTop + this.plotHeight / 2);
      return Math.max(14, cy - 14);
    }
  },
  mounted() {
    this.updateDimensions();
    if (window.ResizeObserver && this.$refs.chartContainer) {
      this.resizeObserver = new ResizeObserver(() => {
        this.updateDimensions();
      });
      this.resizeObserver.observe(this.$refs.chartContainer);
    }
    window.addEventListener('resize', this.updateDimensions);

    if (window.MutationObserver && typeof document !== 'undefined') {
      this.themeObserver = new MutationObserver(() => {
        this.domThemeIsDark = document.documentElement.classList.contains('dark');
      });
      this.themeObserver.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['class']
      });
    }
  },
  beforeDestroy() {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
    }
    if (this.themeObserver) {
      this.themeObserver.disconnect();
    }
    window.removeEventListener('resize', this.updateDimensions);
  },
  methods: {
    updateDimensions() {
      if (this.$refs.chartContainer) {
        const rect = this.$refs.chartContainer.getBoundingClientRect();
        if (rect.width > 0) {
          this.chartWidth = Math.round(rect.width);
        }
        if (rect.height > 0) {
          this.chartHeight = Math.round(rect.height);
        }
      }
    },
    toggleBandOrientation() {
      this.invertBands = !this.invertBands;
    },
    resetHover() {
      this.hoveredPoint = null;
    },
    selectRange(r) {
      this.currentRange = r;
      this.$emit('range-change', r);
    },
    selectPollutant(code) {
      if (this.internalSelectedPollutant === code && code !== 'OVERALL') {
        this.internalSelectedPollutant = 'OVERALL';
      } else {
        this.internalSelectedPollutant = code;
      }
      this.$emit('pollutant-change', this.internalSelectedPollutant);
    },
    getY(val) {
      const clamped = Math.max(0, Math.min(val, this.maxY));
      if (this.invertBands) {
        // Mockup layout: Green (0-50) at top, Black (>300) at bottom
        return this.padTop + this.plotHeight * (clamped / this.maxY);
      }
      // Standard Regulatory Cartesian: 0 at bottom, 350+ at top
      return this.padTop + this.plotHeight * (1 - clamped / this.maxY);
    },
    shouldShowMarker(idx) {
      if (this.isRegulatoryMode) return true; // Show all hourly points
      if (!this.points || this.points.length <= 30) return true;
      if (idx === 0 || idx === this.points.length - 1) return true;
      const step = Math.ceil(this.points.length / 20);
      return idx % step === 0;
    },
    formatDominant(code) {
      if (!code) return '-';
      if (code.toUpperCase() === 'PM25') return 'PM2.5';
      return code.toUpperCase();
    },
    formatTimeOnly(ms) {
      const d = new Date(ms);
      return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`;
    },
    formatDateOnly(ms) {
      const d = new Date(ms);
      const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      return `${d.getDate().toString().padStart(2, '0')} ${months[d.getMonth()]}`;
    },
    formatTickTime(isoStr) {
      if (!isoStr) return '';
      const d = new Date(isoStr);
      const day = d.getDate().toString().padStart(2, '0');
      const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      const month = months[d.getMonth()];
      const hour = d.getHours().toString().padStart(2, '0');
      const minute = d.getMinutes().toString().padStart(2, '0');
      return `${day} ${month} ${hour}:${minute}`;
    },
    formatWindow(raw) {
      if (!raw || !raw.period_start || !raw.period_end) return '-';
      const dur = this.isRegulatoryMode ? '24 Jam' : `${this.aggregationIntervalSeconds}s`;
      return `${this.formatTickTime(raw.period_start)} – ${this.formatTickTime(raw.period_end)} (${dur})`;
    },
    getStatusBadgeClass(status) {
      if (status === 'VALID') {
        return this.resolvedIsDark
          ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
          : 'bg-emerald-50 text-emerald-700 border border-emerald-300';
      }
      if (status === 'PROVISIONAL') {
        return this.resolvedIsDark
          ? 'bg-amber-500/20 text-amber-400 border border-amber-500/40'
          : 'bg-amber-50 text-amber-700 border border-amber-300';
      }
      return this.resolvedIsDark
        ? 'bg-rose-500/20 text-rose-400 border border-rose-500/40'
        : 'bg-rose-50 text-rose-700 border border-rose-300';
    },
    generateSmoothPath(pts) {
      if (pts.length <= 2) {
        return pts.reduce((acc, p, i) => (i === 0 ? `M ${p.x} ${p.y}` : `${acc} L ${p.x} ${p.y}`), '');
      }

      let d = `M ${pts[0].x} ${pts[0].y}`;
      for (let i = 0; i < pts.length - 1; i++) {
        const p0 = i > 0 ? pts[i - 1] : pts[i];
        const p1 = pts[i];
        const p2 = pts[i + 1];
        const p3 = i < pts.length - 2 ? pts[i + 2] : p2;

        const cp1x = p1.x + (p2.x - p0.x) * 0.18;
        const cp1y = p1.y + (p2.y - p0.y) * 0.18;
        const cp2x = p2.x - (p3.x - p1.x) * 0.18;
        const cp2y = p2.y - (p3.y - p1.y) * 0.18;

        d += ` C ${cp1x.toFixed(1)} ${cp1y.toFixed(1)}, ${cp2x.toFixed(1)} ${cp2y.toFixed(1)}, ${p2.x.toFixed(1)} ${p2.y.toFixed(1)}`;
      }
      return d;
    },
    handleMouseMove(e) {
      if (!this.points || this.points.length === 0) return;
      const rect = this.$refs.chartContainer.getBoundingClientRect();
      const mouseX = e.clientX - rect.left;

      let closest = this.points[0];
      let minDiff = Math.abs(this.points[0].x - mouseX);

      for (let i = 1; i < this.points.length; i++) {
        const diff = Math.abs(this.points[i].x - mouseX);
        if (diff < minDiff) {
          minDiff = diff;
          closest = this.points[i];
        }
      }

      this.hoveredPoint = {
        ...closest,
        clientX: closest.x,
        clientY: closest.y
      };
    }
  }
};
</script>

