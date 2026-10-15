<template>
  <div class="relative w-full flex flex-col space-y-2">
    <!-- Chart Header & Time Range Controls -->
    <div class="flex flex-wrap items-center justify-between gap-2 text-3xs font-mono">
      <div class="flex items-center space-x-2">
        <span class="w-2 h-2 rounded-full bg-cyan-400"></span>
        <span class="font-bold text-slate-700 dark:text-slate-200 uppercase tracking-wider">
          {{ title || 'HISTORICAL ISPU TREND' }}
        </span>
        <span class="text-slate-400 dark:text-slate-500">
          (Permen LHK No. 14/2020)
        </span>
      </div>

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

    <!-- Chart Container -->
    <div
      ref="chartContainer"
      class="relative w-full h-48 sm:h-56 bg-slate-50 dark:bg-[#070D18] rounded-xl border border-slate-200 dark:border-[#1E293B] overflow-hidden select-none"
      @mousemove="handleMouseMove"
      @mouseleave="hoveredPoint = null"
    >
      <!-- Loading State -->
      <div v-if="loading" class="absolute inset-0 z-20 flex items-center justify-center bg-slate-900/20 backdrop-blur-xs">
        <div class="w-5 h-5 border-2 border-cyan-500 border-t-transparent rounded-full animate-spin"></div>
      </div>

      <!-- Empty State -->
      <div v-else-if="!points || points.length === 0" class="absolute inset-0 flex flex-col items-center justify-center text-slate-400 text-xs font-mono">
        <span>No historical ISPU records available for selected range</span>
      </div>

      <!-- SVG Chart -->
      <svg
        v-else
        class="w-full h-full"
        :viewBox="`0 0 ${chartWidth} ${chartHeight}`"
        preserveAspectRatio="none"
      >
        <defs>
          <!-- Gradients for Category Bands -->
          <linearGradient id="bandBaik" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#10B981" stop-opacity="0.12" />
            <stop offset="100%" stop-color="#10B981" stop-opacity="0.04" />
          </linearGradient>
          <linearGradient id="bandSedang" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#3B82F6" stop-opacity="0.12" />
            <stop offset="100%" stop-color="#3B82F6" stop-opacity="0.04" />
          </linearGradient>
          <linearGradient id="bandTidakSehat" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#EAB308" stop-opacity="0.12" />
            <stop offset="100%" stop-color="#EAB308" stop-opacity="0.04" />
          </linearGradient>
          <linearGradient id="bandSangatTidakSehat" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#EF4444" stop-opacity="0.12" />
            <stop offset="100%" stop-color="#EF4444" stop-opacity="0.04" />
          </linearGradient>
          <linearGradient id="bandBerbahaya" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#1E1B4B" stop-opacity="0.25" />
            <stop offset="100%" stop-color="#1E1B4B" stop-opacity="0.08" />
          </linearGradient>
        </defs>

        <!-- Category Background Bands (Lampiran II Permen LHK 14/2020) -->
        <!-- 300+ : Berbahaya -->
        <rect
          v-if="getY(300) > padTop"
          :x="padLeft"
          :y="padTop"
          :width="plotWidth"
          :height="Math.max(0, getY(300) - padTop)"
          fill="url(#bandBerbahaya)"
        />
        <!-- 200 - 300: Sangat Tidak Sehat -->
        <rect
          :x="padLeft"
          :y="getY(300)"
          :width="plotWidth"
          :height="Math.max(0, getY(200) - getY(300))"
          fill="url(#bandSangatTidakSehat)"
        />
        <!-- 100 - 200: Tidak Sehat -->
        <rect
          :x="padLeft"
          :y="getY(200)"
          :width="plotWidth"
          :height="Math.max(0, getY(100) - getY(200))"
          fill="url(#bandTidakSehat)"
        />
        <!-- 50 - 100: Sedang -->
        <rect
          :x="padLeft"
          :y="getY(100)"
          :width="plotWidth"
          :height="Math.max(0, getY(50) - getY(100))"
          fill="url(#bandSedang)"
        />
        <!-- 0 - 50: Baik -->
        <rect
          :x="padLeft"
          :y="getY(50)"
          :width="plotWidth"
          :height="Math.max(0, getY(0) - getY(50))"
          fill="url(#bandBaik)"
        />

        <!-- Category Threshold Guide Lines -->
        <g stroke-dasharray="3,3" stroke-width="0.75" class="opacity-40">
          <line :x1="padLeft" :y1="getY(50)" :x2="padLeft + plotWidth" :y2="getY(50)" stroke="#10B981" />
          <line :x1="padLeft" :y1="getY(100)" :x2="padLeft + plotWidth" :y2="getY(100)" stroke="#3B82F6" />
          <line :x1="padLeft" :y1="getY(200)" :x2="padLeft + plotWidth" :y2="getY(200)" stroke="#EAB308" />
          <line :x1="padLeft" :y1="getY(300)" :x2="padLeft + plotWidth" :y2="getY(300)" stroke="#EF4444" />
        </g>

        <!-- Right Y-Axis Category Labels -->
        <g class="text-[9px] font-mono font-bold select-none" text-anchor="end">
          <text :x="padLeft + plotWidth - 4" :y="getY(25) + 3" fill="#10B981" class="opacity-70">BAIK (0-50)</text>
          <text :x="padLeft + plotWidth - 4" :y="getY(75) + 3" fill="#3B82F6" class="opacity-70">SEDANG (51-100)</text>
          <text :x="padLeft + plotWidth - 4" :y="getY(150) + 3" fill="#EAB308" class="opacity-70">TIDAK SEHAT (101-200)</text>
          <text :x="padLeft + plotWidth - 4" :y="getY(250) + 3" fill="#EF4444" class="opacity-70">SANGAT TIDAK SEHAT (201-300)</text>
          <text v-if="maxY > 300" :x="padLeft + plotWidth - 4" :y="padTop + 10" fill="#A855F7" class="opacity-70">BERBAHAYA (&gt;300)</text>
        </g>

        <!-- Left Y-Axis Numerical Labels -->
        <g class="text-[9px] font-mono select-none text-slate-400 dark:text-slate-500" text-anchor="end">
          <text :x="padLeft - 6" :y="getY(0) + 3">0</text>
          <text :x="padLeft - 6" :y="getY(50) + 3">50</text>
          <text :x="padLeft - 6" :y="getY(100) + 3">100</text>
          <text :x="padLeft - 6" :y="getY(200) + 3">200</text>
          <text :x="padLeft - 6" :y="getY(300) + 3">300</text>
        </g>

        <!-- Trend Line -->
        <path
          :d="trendPath"
          fill="none"
          stroke="#06B6D4"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
        />

        <!-- Data Point Circles -->
        <g v-for="(p, idx) in points" :key="idx">
          <circle
            :cx="p.x"
            :cy="p.y"
            r="3.5"
            :fill="getCategoryHex(p.raw.overall_category)"
            stroke="#0B1422"
            stroke-width="1.5"
            class="transition-all hover:scale-150 cursor-pointer"
          />
        </g>

        <!-- Hover Guideline -->
        <line
          v-if="hoveredPoint"
          :x1="hoveredPoint.x"
          :y1="padTop"
          :x2="hoveredPoint.x"
          :y2="padTop + plotHeight"
          stroke="#06B6D4"
          stroke-width="1"
          stroke-dasharray="2,2"
          class="opacity-70"
        />
      </svg>

      <!-- Tooltip -->
      <div
        v-if="hoveredPoint"
        class="absolute z-30 pointer-events-none p-2 rounded-lg bg-slate-900/95 text-white border border-cyan-500/50 shadow-xl backdrop-blur-md text-3xs font-mono transform -translate-x-1/2 -translate-y-full"
        :style="{ left: `${hoveredPoint.clientPercentX}%`, top: `${Math.max(10, hoveredPoint.clientPercentY - 10)}%` }"
      >
        <div class="flex items-center justify-between space-x-2 border-b border-slate-700 pb-1 mb-1">
          <span class="font-bold text-cyan-400">{{ formatTimestamp(hoveredPoint.raw.period_end) }}</span>
          <span
            class="px-1 py-0.2 rounded text-4xs font-bold"
            :class="hoveredPoint.raw.calculation_status === 'VALID' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-amber-500/20 text-amber-300'"
          >
            {{ hoveredPoint.raw.calculation_status }}
          </span>
        </div>
        <div class="flex items-baseline space-x-2">
          <span class="text-sm font-bold tabular-nums" :style="{ color: getCategoryHex(hoveredPoint.raw.overall_category) }">
            {{ Math.round(hoveredPoint.raw.overall_ispu) }}
          </span>
          <span class="font-bold" :style="{ color: getCategoryHex(hoveredPoint.raw.overall_category) }">
            {{ hoveredPoint.raw.overall_category }}
          </span>
        </div>
        <div class="text-slate-300 mt-0.5">
          <span>Dominant: <strong class="text-white">{{ hoveredPoint.raw.dominant_pollutant }}</strong></span>
          <span class="mx-1">·</span>
          <span>Coverage: {{ (hoveredPoint.raw.completeness_percent || 0).toFixed(1) }}%</span>
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
    initialRange: { type: String, default: '24h' }
  },
  data() {
    return {
      currentRange: this.initialRange,
      rangeOptions: [
        { label: '24H', value: '24h' },
        { label: '7D', value: '7d' },
        { label: '30D', value: '30d' }
      ],
      chartWidth: 600,
      chartHeight: 220,
      padTop: 15,
      padBottom: 25,
      padLeft: 30,
      padRight: 100,
      hoveredPoint: null
    };
  },
  computed: {
    plotWidth() {
      return this.chartWidth - this.padLeft - this.padRight;
    },
    plotHeight() {
      return this.chartHeight - this.padTop - this.padBottom;
    },
    maxY() {
      if (!this.records || this.records.length === 0) return 320;
      let maxVal = 0;
      for (const r of this.records) {
        if (r.overall_ispu > maxVal) maxVal = r.overall_ispu;
      }
      return Math.max(320, Math.ceil(maxVal / 50) * 50);
    },
    points() {
      if (!this.records || this.records.length === 0) return [];
      const n = this.records.length;
      return this.records.map((r, i) => {
        const x = n === 1 ? this.padLeft + this.plotWidth / 2 : this.padLeft + (i / (n - 1)) * this.plotWidth;
        const y = this.getY(r.overall_ispu);
        return {
          x,
          y,
          raw: r,
          percentX: (x / this.chartWidth) * 100,
          percentY: (y / this.chartHeight) * 100
        };
      });
    },
    trendPath() {
      if (this.points.length === 0) return '';
      if (this.points.length === 1) {
        return `M ${this.padLeft} ${this.points[0].y} L ${this.padLeft + this.plotWidth} ${this.points[0].y}`;
      }
      return this.points.reduce((acc, p, i) => {
        return i === 0 ? `M ${p.x} ${p.y}` : `${acc} L ${p.x} ${p.y}`;
      }, '');
    }
  },
  methods: {
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
        case 'SEDANG': return '#3B82F6';
        case 'TIDAK_SEHAT': return '#EAB308';
        case 'SANGAT_TIDAK_SEHAT': return '#EF4444';
        case 'BERBAHAYA': return '#A855F7';
        default: return '#64748B';
      }
    },
    formatTimestamp(isoStr) {
      if (!isoStr) return '-';
      const d = new Date(isoStr);
      return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit' });
    },
    handleMouseMove(e) {
      if (!this.points || this.points.length === 0) return;
      const rect = this.$refs.chartContainer.getBoundingClientRect();
      const mouseX = e.clientX - rect.left;
      const svgX = (mouseX / rect.width) * this.chartWidth;

      let closest = this.points[0];
      let minDiff = Math.abs(this.points[0].x - svgX);
      for (let i = 1; i < this.points.length; i++) {
        const diff = Math.abs(this.points[i].x - svgX);
        if (diff < minDiff) {
          minDiff = diff;
          closest = this.points[i];
        }
      }
      this.hoveredPoint = {
        ...closest,
        clientPercentX: (closest.x / this.chartWidth) * 100,
        clientPercentY: (closest.y / this.chartHeight) * 100
      };
    }
  }
};
</script>
