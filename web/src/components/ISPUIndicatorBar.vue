<template>
  <div v-if="isAqmsDomain" class="w-full flex flex-col space-y-1.5 transition-all">
    <!-- Compact SCADA ISPU Indicator Bar -->
    <div
      class="w-full px-3 py-2 rounded-xl border flex flex-wrap items-center justify-between gap-2.5 transition-all duration-300 shadow-xs"
      :class="barThemeClass"
    >
      <!-- Left: ISPU Metric & Category Badge -->
      <div class="flex items-center space-x-3">
        <!-- Glowing Pulse Dot & Main ISPU Value -->
        <div class="flex items-center space-x-2">
          <span class="w-2.5 h-2.5 rounded-full" :class="categoryDotClass"></span>
          <span class="text-3xs font-mono font-bold tracking-widest uppercase text-slate-500 dark:text-slate-400">
            ISPU:
          </span>
          <span class="text-xl sm:text-2xl font-bold font-mono tabular-nums leading-none tracking-tight" :class="categoryTextClass">
            {{ displayISPU }}
          </span>
        </div>

        <!-- Category Badge -->
        <span
          class="px-2 py-0.5 rounded-md text-3xs font-mono font-bold uppercase tracking-wider border shadow-xs"
          :class="categoryBadgeClass"
        >
          {{ ispuCategory }}
        </span>

        <!-- Dominant Pollutant Badge -->
        <div class="hidden sm:flex items-center space-x-1 text-3xs font-mono">
          <span class="text-slate-500 dark:text-slate-400">Dominant:</span>
          <span class="font-bold text-slate-900 dark:text-white px-1.5 py-0.5 rounded bg-slate-200/50 dark:bg-slate-800/80 border border-slate-300 dark:border-slate-700">
            {{ dominantPollutant || '-' }}
          </span>
        </div>
      </div>

      <!-- Center: Measurement Period & Validity -->
      <div class="flex items-center space-x-2 text-3xs font-mono text-slate-500 dark:text-slate-400">
        <!-- 24-Hour Period -->
        <span class="hidden md:inline">
          ⏱️ 24h Window: <strong class="text-slate-700 dark:text-slate-300">{{ formattedPeriod }}</strong>
        </span>

        <!-- Validity Badge -->
        <span
          class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold uppercase tracking-wider border"
          :class="validityBadgeClass"
          :title="validityTooltip"
        >
          {{ calculationStatus }}
          <span v-if="completenessPercent > 0">({{ completenessPercent.toFixed(0) }}%)</span>
        </span>

        <!-- Recalculated / Freshness indicator -->
        <span class="hidden lg:inline text-slate-400 dark:text-slate-500 text-4xs">
          · {{ lastUpdatedStr }}
        </span>
      </div>

      <!-- Right: Action Controls -->
      <div class="flex items-center space-x-1.5">
        <!-- Inline Trend Toggle -->
        <button
          @click="toggleTrend"
          class="px-2 py-1 rounded-lg text-3xs font-mono font-semibold transition border flex items-center space-x-1"
          :class="showTrend
            ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border-cyan-500/50 shadow-xs'
            : 'bg-white/80 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-300 dark:border-[#20354B] hover:border-cyan-500'"
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
          <span>Full ISPU →</span>
        </button>

        <!-- Recalculate Button -->
        <button
          v-if="canRecalculate"
          @click="triggerRecalculation"
          :disabled="recalculating"
          class="p-1 rounded-lg text-slate-500 hover:text-cyan-600 dark:hover:text-cyan-400 transition"
          :title="'Recalculate ISPU (Current 24h Window)'"
        >
          <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': recalculating }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Collapsible Embedded Trend Chart Drawer -->
    <div
      v-if="showTrend"
      class="p-3 rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] shadow-md transition-all"
    >
      <ISPUTrendChart
        :records="trendRecords"
        :loading="trendLoading"
        :initial-range="trendRange"
        title="ISPU 24-HOUR ROLLING TREND"
        @range-change="handleRangeChange"
      />
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
      trendRange: '24h',
      trendRecords: [],
      trendLoading: false,
      pollTimer: null
    };
  },
  computed: {
    displayISPU() {
      if (!this.ispuData || this.ispuData.overall_ispu === undefined) return '-';
      return Math.round(this.ispuData.overall_ispu);
    },
    ispuCategory() {
      return (this.ispuData && this.ispuData.overall_category) || 'TIDAK_TERSEDIA';
    },
    dominantPollutant() {
      return (this.ispuData && this.ispuData.dominant_pollutant) || '-';
    },
    calculationStatus() {
      return (this.ispuData && this.ispuData.calculation_status) || 'PROVISIONAL';
    },
    completenessPercent() {
      return (this.ispuData && this.ispuData.completeness_percent) || 0;
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
      switch (this.ispuCategory) {
        case 'BAIK':
          return 'bg-emerald-500/5 dark:bg-emerald-950/20 border-emerald-500/30';
        case 'SEDANG':
          return 'bg-blue-500/5 dark:bg-blue-950/20 border-blue-500/30';
        case 'TIDAK_SEHAT':
          return 'bg-amber-500/5 dark:bg-amber-950/20 border-amber-500/30';
        case 'SANGAT_TIDAK_SEHAT':
          return 'bg-rose-500/5 dark:bg-rose-950/20 border-rose-500/30';
        case 'BERBAHAYA':
          return 'bg-purple-900/10 dark:bg-[#1E1B4B]/40 border-purple-500/40';
        default:
          return 'bg-slate-100/80 dark:bg-[#0B1422] border-slate-300 dark:border-[#20354B]';
      }
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
        return 'Regulatory Valid: Data completeness meets or exceeds the 75% threshold (Permen LHK No. 14/2020)';
      }
      return 'Provisional calculation: Insufficient temporal coverage (< 75%) or partial data window';
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
          this.fetchLatestISPU();
          if (this.showTrend) {
            this.fetchTrend();
          }
        }
      }
    }
  },
  mounted() {
    this.pollTimer = setInterval(() => {
      if (this.deviceId && this.isAqmsDomain) {
        this.fetchLatestISPU();
      }
    }, 60000); // 1 minute poll
  },
  beforeDestroy() {
    if (this.pollTimer) {
      clearInterval(this.pollTimer);
    }
  },
  methods: {
    async fetchLatestISPU() {
      if (!this.deviceId) return;
      this.loading = true;
      try {
        const token = localStorage.getItem('token');
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
        const token = localStorage.getItem('token');
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
    toggleTrend() {
      this.showTrend = !this.showTrend;
      if (this.showTrend && this.trendRecords.length === 0) {
        this.fetchTrend();
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
        const token = localStorage.getItem('token');
        const res = await axios.post('/api/ispu/recalculate', {
          device_id: this.deviceId
        }, {
          headers: token ? { Authorization: `Bearer ${token}` } : {}
        });
        if (res.data && res.data.success && res.data.data) {
          this.ispuData = res.data.data;
          if (this.showTrend) {
            this.fetchTrend();
          }
        }
      } catch (err) {
        alert('ISPU recalculation error: ' + (err.response?.data?.error || err.message));
      } finally {
        this.recalculating = false;
      }
    }
  }
};
</script>
