<template>
  <div class="relative w-full flex flex-col space-y-2.5">
    <!-- Chart Header & Controls -->
    <div class="flex flex-wrap items-center justify-between gap-2 text-3xs font-mono">
      <div class="flex items-center space-x-2">
        <span class="w-2.5 h-2.5 rounded-full" :class="pollutantDotClass"></span>
        <div>
          <span class="font-bold text-slate-800 dark:text-slate-100 uppercase tracking-wider">
            {{ chartDisplayTitle }}
          </span>
          <span v-if="pollutantSubtitle" class="text-slate-400 dark:text-slate-500 ml-1.5 hidden sm:inline">
            ({{ pollutantSubtitle }})
          </span>
        </div>
      </div>

      <div class="flex flex-wrap items-center gap-1.5">
        <!-- Scale Mode Toggle: Full Scale (0-350 All 5 Bands) vs Focus (Zoomed) -->
        <button
          @click="toggleScaleMode"
          class="px-2 py-0.5 rounded text-4xs font-mono font-semibold transition border flex items-center space-x-1"
          :class="isAdaptiveScale
            ? 'bg-amber-500/20 text-amber-600 dark:text-amber-300 border-amber-500/40 shadow-xs'
            : 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border-cyan-500/40 shadow-xs'"
          :title="isAdaptiveScale ? 'Adaptive Scale: Zoomed to data range' : 'Full Scale: 0 to 350+ (All 5 Regulatory Bands)'"
        >
          <span>{{ isAdaptiveScale ? '⚡ Focus (Zoom)' : '📏 Full (0-350 Bands)' }}</span>
        </button>

        <!-- Time Range Selector -->
        <div class="flex items-center space-x-1 bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">
          <button
            v-for="r in rangeOptions"
            :key="r.value"
            @click="selectRange(r.value)"
            class="px-2 py-0.5 rounded text-4xs font-mono font-semibold transition"
            :class="currentRange === r.value
              ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border border-cyan-500/40 shadow-xs'
              : 'text-slate-500 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200'"
          >
            {{ r.label }}
          </button>
        </div>
      </div>
    </div>

    <!-- Chart Canvas Container -->
    <div
      ref="chartContainer"
      class="relative w-full h-56 sm:h-68 bg-slate-50 dark:bg-[#070D18] rounded-xl border border-slate-200/90 dark:border-[#1E293B] overflow-hidden select-none shadow-inner"
      @mousemove="handleMouseMove"
      @mouseleave="hoveredPoint = null"
    >
      <!-- Loading State -->
      <div v-if="loading" class="absolute inset-0 z-20 flex items-center justify-center bg-slate-900/30 backdrop-blur-xs">
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
          <!-- Gradient for Active Trend Area Fill -->
          <linearGradient id="trendAreaGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="activeCurveColor" stop-opacity="0.35" />
            <stop offset="60%" :stop-color="activeCurveColor" stop-opacity="0.10" />
            <stop offset="100%" :stop-color="activeCurveColor" stop-opacity="0.0" />
          </linearGradient>

          <!-- Neon Glow Filter -->
          <filter id="neonGlow" x="-20%" y="-20%" width="140%" height="140%">
            <feDropShadow dx="0" dy="1" stdDeviation="2.5" :flood-color="activeCurveColor" flood-opacity="0.65" />
          </filter>

          <!-- Vivid Band Gradients for all 5 Permen LHK 14/2020 Categories -->
          <!-- 1. BAIK (0 - 50): Emerald -->
          <linearGradient id="bandBaikGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#10B981" stop-opacity="0.20" />
            <stop offset="100%" stop-color="#10B981" stop-opacity="0.07" />
          </linearGradient>

          <!-- 2. SEDANG (51 - 100): Electric Cyan / Sky Blue -->
          <linearGradient id="bandSedangGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#0EA5E9" stop-opacity="0.20" />
            <stop offset="100%" stop-color="#0EA5E9" stop-opacity="0.07" />
          </linearGradient>

          <!-- 3. TIDAK SEHAT (101 - 200): Luminous Warm Amber -->
          <linearGradient id="bandTidakSehatGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#F59E0B" stop-opacity="0.22" />
            <stop offset="100%" stop-color="#F59E0B" stop-opacity="0.08" />
          </linearGradient>

          <!-- 4. SANGAT TIDAK SEHAT (201 - 300): Vivid Crimson Coral -->
          <linearGradient id="bandSangatTidakSehatGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#F43F5E" stop-opacity="0.25" />
            <stop offset="100%" stop-color="#F43F5E" stop-opacity="0.09" />
          </linearGradient>

          <!-- 5. BERBAHAYA (> 300): Deep Neon Violet Purple -->
          <linearGradient id="bandBerbahayaGrad" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#A855F7" stop-opacity="0.30" />
            <stop offset="100%" stop-color="#A855F7" stop-opacity="0.12" />
          </linearGradient>
        </defs>

        <!-- Category Regulatory Background Bands (Permen LHK No. 14/2020 Lampiran II) -->
        <!-- Zone 5: Berbahaya (>300) -->
        <rect
          v-if="maxY > 300"
          :x="padLeft"
          :y="padTop"
          :width="plotWidth"
          :height="Math.max(0, getY(300) - padTop)"
          fill="url(#bandBerbahayaGrad)"
        />
        <!-- Zone 4: Sangat Tidak Sehat (201 - 300) -->
        <rect
          v-if="maxY > 200"
          :x="padLeft"
          :y="Math.max(padTop, getY(300))"
          :width="plotWidth"
          :height="Math.max(0, getY(200) - Math.max(padTop, getY(300)))"
          fill="url(#bandSangatTidakSehatGrad)"
        />
        <!-- Zone 3: Tidak Sehat (101 - 200) -->
        <rect
          v-if="maxY > 100"
          :x="padLeft"
          :y="Math.max(padTop, getY(200))"
          :width="plotWidth"
          :height="Math.max(0, getY(100) - Math.max(padTop, getY(200)))"
          fill="url(#bandTidakSehatGrad)"
        />
        <!-- Zone 2: Sedang (51 - 100) -->
        <rect
          v-if="maxY > 50"
          :x="padLeft"
          :y="Math.max(padTop, getY(100))"
          :width="plotWidth"
          :height="Math.max(0, getY(50) - Math.max(padTop, getY(100)))"
          fill="url(#bandSedangGrad)"
        />
        <!-- Zone 1: Baik (0 - 50) -->
        <rect
          :x="padLeft"
          :y="Math.max(padTop, getY(50))"
          :width="plotWidth"
          :height="Math.max(0, getY(0) - Math.max(padTop, getY(50)))"
          fill="url(#bandBaikGrad)"
        />

        <!-- Category Threshold Boundary Lines (Dashed, High-Visibility) -->
        <g stroke-dasharray="4,4" stroke-width="1.2" class="opacity-70">
          <line v-if="getY(50) >= padTop" :x1="padLeft" :y1="getY(50)" :x2="padLeft + plotWidth" :y2="getY(50)" stroke="#10B981" />
          <line v-if="getY(100) >= padTop" :x1="padLeft" :y1="getY(100)" :x2="padLeft + plotWidth" :y2="getY(100)" stroke="#0EA5E9" />
          <line v-if="getY(200) >= padTop" :x1="padLeft" :y1="getY(200)" :x2="padLeft + plotWidth" :y2="getY(200)" stroke="#F59E0B" />
          <line v-if="getY(300) >= padTop" :x1="padLeft" :y1="getY(300)" :x2="padLeft + plotWidth" :y2="getY(300)" stroke="#F43F5E" />
        </g>

        <!-- Right Side Floating Regulatory Band Badges (High Contrast in Night Mode) -->
        <g class="text-[9px] font-mono font-extrabold select-none" text-anchor="end">
          <!-- BERBAHAYA (>300) -->
          <g v-if="maxY > 300">
            <rect
              :x="padLeft + plotWidth - 116"
              :y="Math.max(padTop + 2, getY(300) - 20)"
              width="114"
              height="16"
              rx="4"
              fill="#0B1422"
              fill-opacity="0.9"
              stroke="#A855F7"
              stroke-width="1"
            />
            <text
              :x="padLeft + plotWidth - 6"
              :y="Math.max(padTop + 2, getY(300) - 20) + 11.5"
              fill="#D8B4FE"
            >
              BERBAHAYA (&gt;300)
            </text>
          </g>

          <!-- SANGAT TIDAK SEHAT (201-300) -->
          <g v-if="maxY > 200">
            <rect
              :x="padLeft + plotWidth - 162"
              :y="getY(250) - 8"
              width="160"
              height="16"
              rx="4"
              fill="#0B1422"
              fill-opacity="0.9"
              stroke="#F43F5E"
              stroke-width="1"
            />
            <text
              :x="padLeft + plotWidth - 6"
              :y="getY(250) + 3.5"
              fill="#FDA4AF"
            >
              SANGAT TIDAK SEHAT (201-300)
            </text>
          </g>

          <!-- TIDAK SEHAT (101-200) -->
          <g v-if="maxY > 100">
            <rect
              :x="padLeft + plotWidth - 128"
              :y="getY(150) - 8"
              width="126"
              height="16"
              rx="4"
              fill="#0B1422"
              fill-opacity="0.9"
              stroke="#F59E0B"
              stroke-width="1"
            />
            <text
              :x="padLeft + plotWidth - 6"
              :y="getY(150) + 3.5"
              fill="#FDE047"
            >
              TIDAK SEHAT (101-200)
            </text>
          </g>

          <!-- SEDANG (51-100) -->
          <g v-if="maxY > 50">
            <rect
              :x="padLeft + plotWidth - 98"
              :y="getY(75) - 8"
              width="96"
              height="16"
              rx="4"
              fill="#0B1422"
              fill-opacity="0.9"
              stroke="#0EA5E9"
              stroke-width="1"
            />
            <text
              :x="padLeft + plotWidth - 6"
              :y="getY(75) + 3.5"
              fill="#7DD3FC"
            >
              SEDANG (51-100)
            </text>
          </g>

          <!-- BAIK (0-50) -->
          <g>
            <rect
              :x="padLeft + plotWidth - 76"
              :y="getY(25) - 8"
              width="74"
              height="16"
              rx="4"
              fill="#0B1422"
              fill-opacity="0.9"
              stroke="#10B981"
              stroke-width="1"
            />
            <text
              :x="padLeft + plotWidth - 6"
              :y="getY(25) + 3.5"
              fill="#6EE7B7"
            >
              BAIK (0-50)
            </text>
          </g>
        </g>

        <!-- Left Y-Axis Numerical Scale Labels (Crisp Bright Text in Night Mode) -->
        <g class="text-[9px] font-mono font-bold select-none" text-anchor="end" fill="#CBD5E1">
          <text :x="padLeft - 6" :y="getY(0) + 3">0</text>
          <text v-if="getY(50) >= padTop" :x="padLeft - 6" :y="getY(50) + 3">50</text>
          <text v-if="getY(100) >= padTop" :x="padLeft - 6" :y="getY(100) + 3">100</text>
          <text v-if="getY(200) >= padTop" :x="padLeft - 6" :y="getY(200) + 3">200</text>
          <text v-if="getY(300) >= padTop" :x="padLeft - 6" :y="getY(300) + 3">300</text>
          <text v-if="maxY > 300" :x="padLeft - 6" :y="padTop + 6">{{ maxY }}</text>
        </g>

        <!-- Gradient Area Fill Under Curve -->
        <path
          v-if="areaPath"
          :d="areaPath"
          fill="url(#trendAreaGradient)"
        />

        <!-- Smooth Trend Path with Neon Glow -->
        <path
          :d="curvePath"
          fill="none"
          :stroke="activeCurveColor"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          filter="url(#neonGlow)"
        />

        <!-- Data Point Markers -->
        <g v-for="(p, idx) in points" :key="idx">
          <!-- Active Pulsing Ripple on the Latest Data Point -->
          <circle
            v-if="idx === points.length - 1"
            :cx="p.x"
            :cy="p.y"
            r="8"
            :fill="p.color"
            class="animate-ping opacity-35"
          />
          <!-- Core Data Dot -->
          <circle
            :cx="p.x"
            :cy="p.y"
            :r="idx === points.length - 1 ? 4.5 : 3"
            :fill="p.color"
            stroke="#070D18"
            :stroke-width="idx === points.length - 1 ? 2 : 1.2"
            class="transition-all hover:scale-175 cursor-pointer"
          />
        </g>

        <!-- Hover Guideline and Crosshair -->
        <g v-if="hoveredPoint">
          <line
            :x1="hoveredPoint.x"
            :y1="padTop"
            :x2="hoveredPoint.x"
            :y2="padTop + plotHeight"
            :stroke="activeCurveColor"
            stroke-width="1.2"
            stroke-dasharray="2,2"
            class="opacity-75"
          />
          <circle
            :cx="hoveredPoint.x"
            :cy="hoveredPoint.y"
            r="6"
            fill="none"
            :stroke="activeCurveColor"
            stroke-width="2"
            class="animate-pulse"
          />
        </g>
      </svg>

      <!-- Glassmorphism Floating Tooltip -->
      <div
        v-if="hoveredPoint"
        class="absolute z-30 pointer-events-none p-2.5 rounded-xl bg-slate-900/95 text-white border border-cyan-500/40 shadow-2xl backdrop-blur-md text-3xs font-mono transform -translate-x-1/2 -translate-y-full transition-transform duration-75"
        :style="{
          left: `${hoveredPoint.clientX}px`,
          top: `${Math.max(12, hoveredPoint.clientY - 12)}px`
        }"
      >
        <div class="flex items-center justify-between space-x-3 border-b border-slate-700/80 pb-1.5 mb-1.5">
          <span class="font-bold text-cyan-400">{{ formatTimestamp(hoveredPoint.raw.period_end) }}</span>
          <span
            class="px-1.5 py-0.2 rounded text-4xs font-bold uppercase border"
            :class="hoveredPoint.statusClass"
          >
            {{ hoveredPoint.raw.calculation_status }}
          </span>
        </div>

        <div class="flex items-baseline space-x-2">
          <span class="text-base font-black tabular-nums" :style="{ color: hoveredPoint.color }">
            {{ Math.round(hoveredPoint.value) }}
          </span>
          <span class="text-3xs font-bold uppercase tracking-wider" :style="{ color: hoveredPoint.color }">
            {{ hoveredPoint.category }}
          </span>
        </div>

        <div class="text-slate-300 mt-1 space-y-0.5">
          <div v-if="isOverallMode">
            <span>Dominant: <strong class="text-white">{{ hoveredPoint.raw.dominant_pollutant }}</strong></span>
            <span class="mx-1">·</span>
            <span>Coverage: {{ (hoveredPoint.raw.completeness_percent || 0).toFixed(1) }}%</span>
          </div>
          <div v-else>
            <span>Parameter: <strong class="text-white">{{ selectedPollutant }}</strong></span>
            <span class="mx-1">·</span>
            <span>Conc: <strong class="text-white">{{ hoveredPoint.concentration.toFixed(1) }}</strong> {{ hoveredPoint.unit }}</span>
          </div>
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
    pollutantName: { type: String, default: '' }
  },
  data() {
    return {
      currentRange: this.initialRange,
      // Default to Full Regulatory Scale (0-350) so all 5 categories (Baik to Berbahaya) are visible
      isAdaptiveScale: false,
      rangeOptions: [
        { label: '24H', value: '24h' },
        { label: '7D', value: '7d' },
        { label: '30D', value: '30d' }
      ],
      chartWidth: 700,
      chartHeight: 250,
      padTop: 18,
      padBottom: 22,
      padLeft: 32,
      padRight: 20, // Reduced from 110 to 20 so data curve flows all the way to the right edge!
      hoveredPoint: null,
      resizeObserver: null
    };
  },
  computed: {
    isOverallMode() {
      return !this.selectedPollutant || this.selectedPollutant === 'OVERALL';
    },
    chartDisplayTitle() {
      if (this.isOverallMode) {
        return this.title || 'HISTORICAL ISPU TREND & REGULATORY BANDS';
      }
      return `HISTORICAL ISPU TREND — ${this.selectedPollutant}`;
    },
    pollutantSubtitle() {
      if (this.isOverallMode) return 'Permen LHK No. 14/2020';
      return this.pollutantName || 'Official Pollutant Trend';
    },
    pollutantDotClass() {
      if (this.isOverallMode) return 'bg-cyan-400';
      return 'bg-amber-400';
    },
    plotWidth() {
      return Math.max(100, this.chartWidth - this.padLeft - this.padRight);
    },
    plotHeight() {
      return Math.max(80, this.chartHeight - this.padTop - this.padBottom);
    },
    activeCurveColor() {
      if (this.points.length === 0) return '#06B6D4';
      const last = this.points[this.points.length - 1];
      return last.color || '#06B6D4';
    },
    maxY() {
      if (!this.records || this.records.length === 0) return 350;
      let maxVal = 0;
      for (const p of this.extractedRawPoints) {
        if (p.value > maxVal) maxVal = p.value;
      }

      if (!this.isAdaptiveScale) {
        // Full Scale: Always show all 5 bands (0 to 350+)
        return Math.max(350, Math.ceil((maxVal + 25) / 50) * 50);
      }

      // Adaptive Scale: Focus on data range with padding
      if (maxVal <= 80) return 120;
      if (maxVal <= 130) return 160;
      if (maxVal <= 180) return 220;
      return Math.max(350, Math.ceil((maxVal + 25) / 50) * 50);
    },
    extractedRawPoints() {
      if (!this.records || this.records.length === 0) return [];
      return this.records.map(r => {
        if (this.isOverallMode) {
          return {
            raw: r,
            value: Number(r.overall_ispu || 0),
            category: r.overall_category || 'TIDAK_TERSEDIA',
            concentration: 0,
            unit: 'µg/m³'
          };
        }
        // Extract specific pollutant breakdown
        const pr = (r.parameter_results || []).find(p => p.pollutant_code === this.selectedPollutant);
        if (pr) {
          return {
            raw: r,
            value: Number(pr.calculated_ispu || 0),
            category: pr.category || 'TIDAK_TERSEDIA',
            concentration: Number(pr.normalized_concentration || 0),
            unit: pr.normalized_unit || 'µg/m³'
          };
        }
        return {
          raw: r,
          value: 0,
          category: 'TIDAK_TERSEDIA',
          concentration: 0,
          unit: 'µg/m³'
        };
      });
    },
    points() {
      const raw = this.extractedRawPoints;
      if (!raw || raw.length === 0) return [];
      const n = raw.length;

      return raw.map((item, i) => {
        const x = n === 1
          ? this.padLeft + this.plotWidth / 2
          : this.padLeft + (i / (n - 1)) * this.plotWidth;
        const y = this.getY(item.value);
        const color = this.getCategoryHex(item.category);
        const statusClass = item.raw.calculation_status === 'VALID'
          ? 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40'
          : 'bg-amber-500/20 text-amber-300 border-amber-500/40';

        return {
          x,
          y,
          value: item.value,
          category: item.category,
          concentration: item.concentration,
          unit: item.unit,
          color,
          statusClass,
          raw: item.raw
        };
      });
    },
    curvePath() {
      if (this.points.length === 0) return '';
      if (this.points.length === 1) {
        return `M ${this.padLeft} ${this.points[0].y} L ${this.padLeft + this.plotWidth} ${this.points[0].y}`;
      }
      return this.generateSmoothPath(this.points);
    },
    areaPath() {
      if (this.points.length === 0) return '';
      const basePath = this.curvePath;
      const lastX = this.points[this.points.length - 1].x;
      const firstX = this.points[0].x;
      const bottomY = this.padTop + this.plotHeight;
      return `${basePath} L ${lastX} ${bottomY} L ${firstX} ${bottomY} Z`;
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
    toggleScaleMode() {
      this.isAdaptiveScale = !this.isAdaptiveScale;
    },
    selectRange(r) {
      this.currentRange = r;
      this.$emit('range-change', r);
    },
    getY(val) {
      const clamped = Math.max(0, Math.min(val, this.maxY));
      return this.padTop + this.plotHeight * (1 - clamped / this.maxY);
    },
    getCategoryHex(cat) {
      switch (cat) {
        case 'BAIK': return '#10B981';
        case 'SEDANG': return '#0EA5E9';
        case 'TIDAK_SEHAT': return '#F59E0B';
        case 'SANGAT_TIDAK_SEHAT': return '#F43F5E';
        case 'BERBAHAYA': return '#A855F7';
        default: return '#64748B';
      }
    },
    formatTimestamp(isoStr) {
      if (!isoStr) return '-';
      const d = new Date(isoStr);
      return d.toLocaleDateString('id-ID', {
        day: '2-digit',
        month: 'short',
        hour: '2-digit',
        minute: '2-digit'
      });
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
