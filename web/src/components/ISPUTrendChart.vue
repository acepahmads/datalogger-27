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
            · {{ selectedPollutant }} ({{ pollutantName || 'Pollutant Trend' }})
          </span>
        </div>
      </div>


      <!-- Right Side Controls: Compact Capsule Range Switcher & Close Icon -->
      <div class="flex items-center space-x-2">
        <!-- Time Range Selector Capsule Buttons (24 Hours | 7 Days | 30 Days) -->
        <div class="flex items-center bg-slate-200/80 dark:bg-[#1E293B]/80 p-0.5 rounded-lg border border-slate-300 dark:border-slate-700/70 shadow-xs">
          <button
            v-for="r in rangeOptions"
            :key="r.value"
            @click="selectRange(r.value)"
            class="px-2.5 py-1 rounded-md text-xs font-medium transition cursor-pointer select-none"
            :class="currentRange === r.value
              ? 'bg-slate-700 dark:bg-[#334155] text-white shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            {{ r.label }}
          </button>
        </div>

        <!-- Subtle Close / Reset Hover Button (Matching Reference Icon) -->
        <button
          @click="resetHover"
          class="p-1 rounded text-slate-400 hover:text-slate-700 dark:hover:text-white transition"
          title="Reset hover selection"
        >
          <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Chart Canvas Container -->
    <div
      ref="chartContainer"
      class="relative w-full h-64 sm:h-72 bg-[#0B1322] rounded-xl border border-slate-700/60 overflow-hidden select-none shadow-2xl"
      @mousemove="handleMouseMove"
      @mouseleave="hoveredPoint = null"
    >
      <!-- Loading State -->
      <div v-if="loading" class="absolute inset-0 z-20 flex items-center justify-center bg-slate-900/60 backdrop-blur-xs">
        <div class="flex items-center space-x-2 text-xs font-mono text-cyan-400 bg-slate-900/90 px-3 py-1.5 rounded-lg border border-cyan-500/30">
          <div class="w-4 h-4 border-2 border-cyan-400 border-t-transparent rounded-full animate-spin"></div>
          <span>Loading historical trend...</span>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else-if="!points || points.length === 0" class="absolute inset-0 flex flex-col items-center justify-center text-slate-400 text-xs font-mono space-y-1">
        <span>No historical ISPU records available for selected range</span>
        <span class="text-3xs text-slate-500">Historical trend will populate automatically from aggregation rollups</span>
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
            <feDropShadow dx="0" dy="0" stdDeviation="3.5" flood-color="#00F2FE" flood-opacity="0.9" />
          </filter>

          <!-- Subtle linear gradients for solid / blended bands matching reference image -->
          <linearGradient id="bandBaik" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#1B873F" stop-opacity="0.95" />
            <stop offset="100%" stop-color="#166534" stop-opacity="0.95" />
          </linearGradient>

          <linearGradient id="bandSedang" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#1D63B8" stop-opacity="0.95" />
            <stop offset="100%" stop-color="#1E40AF" stop-opacity="0.95" />
          </linearGradient>

          <linearGradient id="bandTidakSehat" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#C49000" stop-opacity="0.95" />
            <stop offset="100%" stop-color="#B45309" stop-opacity="0.95" />
          </linearGradient>

          <linearGradient id="bandSangatTidakSehat" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#B91C1C" stop-opacity="0.95" />
            <stop offset="100%" stop-color="#991B1B" stop-opacity="0.95" />
          </linearGradient>

          <linearGradient id="bandBerbahaya" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#0B132B" stop-opacity="0.98" />
            <stop offset="100%" stop-color="#0A1128" stop-opacity="0.98" />
          </linearGradient>

          <linearGradient id="trendAreaGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#00F2FE" stop-opacity="0.30" />
            <stop offset="100%" stop-color="#00F2FE" stop-opacity="0.0" />
          </linearGradient>
        </defs>

        <!-- 5 Horizontal Full-Width Background Category Bands (Solid / Subtly Blended) -->
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

        <!-- Category Clean Boundary Lines -->
        <g stroke="rgba(0, 0, 0, 0.45)" stroke-width="1.2">
          <line
            v-for="(b, idx) in computedBands.slice(0, 4)"
            :key="`line-${idx}`"
            :x1="padLeft"
            :y1="b.y + b.height"
            :x2="chartWidth"
            :y2="b.y + b.height"
          />
        </g>

        <!-- Category Labels Inside the Left Side of Each Band (Readable White Text) -->
        <g class="band-labels text-[11px] font-mono font-bold select-none" fill="#FFFFFF">
          <text
            v-for="b in computedBands"
            :key="`lbl-${b.key}`"
            :x="padLeft + 12"
            :y="b.y + b.height / 2 + 4"
            letter-spacing="0.5"
          >
            {{ b.label }}
          </text>
        </g>

        <!-- Left Y-Axis Numeric Labels (Muted, Compact) -->
        <g class="text-[10px] font-mono select-none" text-anchor="end" fill="#94A3B8">
          <text
            v-for="tick in yTicks"
            :key="`ytick-${tick.val}`"
            :x="padLeft - 6"
            :y="tick.y + 3.5"
          >
            {{ tick.val }}
          </text>
        </g>

        <!-- Bottom X-Axis Date/Time Labels (Evenly Spaced, No Overlap) -->
        <g class="text-[10px] font-mono select-none" text-anchor="middle" fill="#94A3B8">
          <text
            v-for="(tick, idx) in xTicks"
            :key="`xtick-${idx}`"
            :x="tick.x"
            :y="padTop + plotHeight + 16"
          >
            {{ tick.label }}
          </text>
        </g>

        <!-- Smooth Bright Cyan / Turquoise Trend Curve with Glow -->
        <path
          :d="curvePath"
          fill="none"
          stroke="#00F2FE"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          filter="url(#neonGlow)"
        />

        <!-- Circular Data Point Markers (Intelligently spaced to prevent overlapping clutter) -->
        <g class="data-markers">
          <g v-for="(p, idx) in points" :key="`pt-${idx}`">
            <template v-if="p.isAvailable && p.y !== null && shouldShowMarker(idx)">
              <!-- Ripple Effect on the Latest Point -->
              <circle
                v-if="idx === points.length - 1"
                :cx="p.x"
                :cy="p.y"
                r="7"
                fill="#00F2FE"
                class="animate-ping opacity-35"
              />
              <!-- Core Circular Marker (Cyan with White Ring, Matching Reference Design) -->
              <circle
                :cx="p.x"
                :cy="p.y"
                r="3.5"
                fill="#00F2FE"
                stroke="#FFFFFF"
                stroke-width="1.5"
                class="transition-transform hover:scale-175 cursor-pointer"
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
            stroke="#00F2FE"
            stroke-width="1.2"
            stroke-dasharray="3,3"
            class="opacity-80"
          />
          <circle
            v-if="hoveredPoint.isAvailable && hoveredPoint.y !== null"
            :cx="hoveredPoint.x"
            :cy="hoveredPoint.y"
            r="6"
            fill="#00F2FE"
            stroke="#FFFFFF"
            stroke-width="2"
            filter="url(#neonGlow)"
          />
        </g>
      </svg>

      <!-- Compact Dark Glassmorphic Tooltip (Matches Reference Design) -->
      <div
        v-if="hoveredPoint"
        class="absolute z-30 pointer-events-none p-3 rounded-xl bg-[#0F172A]/95 text-white border border-slate-700/80 shadow-2xl backdrop-blur-md text-[11px] font-mono transform -translate-x-1/2 -translate-y-full transition-transform duration-75"
        :style="{
          left: `${tooltipLeft}px`,
          top: `${tooltipTop}px`
        }"
      >
        <div class="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1.5 whitespace-nowrap">
          <span class="text-slate-400">Mode:</span>
          <span class="font-bold text-cyan-400">
            {{ isRegulatoryMode ? '24-Jam Rolling Regulasi' : `Per Agregasi (${aggIntervalLabel})` }}
          </span>

          <span class="text-slate-400">Window:</span>
          <span class="font-bold text-white">
            {{ formatWindow(hoveredPoint.raw) }}
          </span>

          <span class="text-slate-400">ISPU:</span>
          <span class="font-bold text-[#00F2FE]">
            <template v-if="hoveredPoint.isAvailable && hoveredPoint.value !== null">
              {{ Math.round(hoveredPoint.value) }} ({{ (hoveredPoint.category || '').replace('_', ' ') }})
            </template>
            <template v-else>
              <span class="text-slate-400 font-normal">Tidak Tersedia / Missing</span>
            </template>
          </span>

          <span v-if="!isOverallMode" class="text-slate-400">Concentration:</span>
          <span v-if="!isOverallMode" class="font-bold text-white">
            <template v-if="hoveredPoint.isAvailable && hoveredPoint.concentration != null">
              {{ hoveredPoint.concentration.toFixed(1) }} {{ hoveredPoint.unit }}
            </template>
            <template v-else>
              <span class="text-slate-400">-</span>
            </template>
          </span>

          <span class="text-slate-400">Dominant:</span>
          <span class="font-bold text-white">
            {{ formatDominant(hoveredPoint.raw.dominant_pollutant || selectedPollutant) }}
          </span>

          <span class="text-slate-400">Coverage:</span>
          <span class="font-bold text-white">
            {{ Number(hoveredPoint.raw.completeness_percent || 0).toFixed(1) }}%
            <span class="text-4xs text-slate-400">({{ isRegulatoryMode ? '≥75% valid' : 'bucket' }})</span>
          </span>

          <span class="text-slate-400">Status:</span>
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
    aggregationIntervalSeconds: { type: Number, default: 120 }
  },
  data() {
    return {
      currentRange: this.initialRange,
      // Default to standard regulatory Cartesian scale; toggleable to reference mockup layout
      invertBands: false,
      isAdaptiveScale: false,
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
      resizeObserver: null
    };
  },
  watch: {
    initialRange(newVal) {
      if (newVal) this.currentRange = newVal;
    }
  },
  computed: {
    isRegulatoryMode() {
      return this.mode === 'REGULATORY';
    },
    aggIntervalLabel() {
      if (this.aggregationIntervalSeconds % 60 === 0) {
        return `${this.aggregationIntervalSeconds / 60} Menit`;
      }
      return `${this.aggregationIntervalSeconds}s`;
    },
    isOverallMode() {
      return !this.selectedPollutant || this.selectedPollutant === 'OVERALL';
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
      if (this.selectedPollutant && this.selectedPollutant !== 'OVERALL' && this.selectedPollutant !== 'ALL') {
        const pCode = this.selectedPollutant.toUpperCase();
        if (pCode === 'PM25') {
          return 'Permen LHK No. 14/2020 Pasal 10 ayat (1)a · Jendela 24 Jam (Pelaporan Setiap Jam)';
        }
        return 'Permen LHK No. 14/2020 Pasal 10 ayat (1)b & (2) · Jendela 24 Jam (Jadwal 09.00 & 15.00 WIB / 1 Jam jika >100)';
      }
      return 'Permen LHK No. 14 Tahun 2020 · Jendela Pengukuran 24 Jam (Cadence: 1 Jam · Pasal 10 ayat 1a)';
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
        // Reference Mockup visual layout: Green at top, Blue 2nd, Yellow 3rd, Red 4th, Dark Navy at bottom
        const h5 = this.plotHeight / 5;
        return [
          { key: 'baik', id: 'bandBaik', label: 'BAIK (0-50)', y: this.padTop, height: h5 },
          { key: 'sedang', id: 'bandSedang', label: 'SEDANG (51-100)', y: this.padTop + h5, height: h5 },
          { key: 'tidak_sehat', id: 'bandTidakSehat', label: 'TIDAK SEHAT (101-200)', y: this.padTop + 2 * h5, height: h5 },
          { key: 'sangat_tidak_sehat', id: 'bandSangatTidakSehat', label: 'SANGAT TIDAK SEHAT (201-300)', y: this.padTop + 3 * h5, height: h5 },
          { key: 'berbahaya', id: 'bandBerbahaya', label: 'BERBAHAYA (>300)', y: this.padTop + 4 * h5, height: h5 }
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
          y: yTop,
          height: Math.max(0, y300 - yTop)
        },
        {
          key: 'sangat_tidak_sehat',
          id: 'bandSangatTidakSehat',
          label: 'SANGAT TIDAK SEHAT (201-300)',
          y: y300,
          height: Math.max(0, y200 - y300)
        },
        {
          key: 'tidak_sehat',
          id: 'bandTidakSehat',
          label: 'TIDAK SEHAT (101-200)',
          y: y200,
          height: Math.max(0, y100 - y200)
        },
        {
          key: 'sedang',
          id: 'bandSedang',
          label: 'SEDANG (51-100)',
          y: y100,
          height: Math.max(0, y50 - y100)
        },
        {
          key: 'baik',
          id: 'bandBaik',
          label: 'BAIK (0-50)',
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
      if (this.isRegulatoryMode && !this.isOverallMode) {
        const pCode = (this.selectedPollutant || '').toUpperCase();
        if (pCode !== 'PM25') {
          // Permen LHK 14/2020 Pasal 10 ayat (1)b & ayat (2):
          // Cadence for gases/PM10: 09:00 & 15:00 WIB, or hourly when pollutant or station ISPU > 100
          sourceRecords = this.records.filter((r, idx) => {
            if (idx === this.records.length - 1) return true; // always show latest point
            const d = new Date(r.period_end);
            const hour = d.getHours();
            if (hour === 9 || hour === 15) return true;
            const pr = (r.parameter_results || []).find(p => p.pollutant_code === this.selectedPollutant);
            const val = pr ? pr.calculated_ispu : r.overall_ispu;
            return val > 100;
          });
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
        const pr = (r.parameter_results || []).find(p => p.pollutant_code === this.selectedPollutant);
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
  },
  beforeDestroy() {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
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
      if (status === 'VALID') return 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40';
      if (status === 'PROVISIONAL') return 'bg-amber-500/20 text-amber-400 border border-amber-500/40';
      return 'bg-rose-500/20 text-rose-400 border border-rose-500/40';
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

