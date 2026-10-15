<template>
  <div v-if="isOpen && targetParam" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-4xl shadow-2xl overflow-hidden flex flex-col max-h-[92vh]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/90">
        <div class="flex items-center space-x-3">
          <div class="p-2 rounded-xl bg-purple-500/10 text-purple-400 border border-purple-500/20 flex-shrink-0">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="text-sm font-bold text-white tracking-wide">
                100 Data Terakhir (Historical Telemetry Log)
              </h2>
              <span class="text-3xs font-mono px-2 py-0.5 rounded bg-slate-800 text-purple-300 border border-slate-700">
                {{ targetParam.parameter_code || targetParam.param_code }}
              </span>
              <span class="text-3xs font-mono px-2 py-0.5 rounded bg-black/40 text-slate-300 border border-slate-700/60">
                {{ records.length }} / {{ total }} records
              </span>
            </div>
            <p class="text-3xs text-slate-400 font-mono mt-0.5">
              {{ targetParam.parameter_name || targetParam.param_name || targetParam.parameter_code }} · {{ targetDevice.device_code || '--' }}
            </p>
          </div>
        </div>
        <button @click="closeModal" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Quick Summary Stats Strip -->
      <div class="px-6 py-2.5 bg-[#080E18] border-b border-slate-800 flex flex-wrap items-center justify-between gap-3 text-3xs font-mono">
        <div class="flex items-center space-x-4">
          <div>
            <span class="text-slate-500 uppercase font-sans">Latest:</span>
            <span class="text-emerald-400 font-bold ml-1">{{ latestValueDisplay }}</span>
          </div>
          <div>
            <span class="text-slate-500 uppercase font-sans">Min:</span>
            <span class="text-cyan-300 font-bold ml-1">{{ summaryStats.min !== null ? summaryStats.min : '--' }}</span>
          </div>
          <div>
            <span class="text-slate-500 uppercase font-sans">Max:</span>
            <span class="text-cyan-300 font-bold ml-1">{{ summaryStats.max !== null ? summaryStats.max : '--' }}</span>
          </div>
          <div>
            <span class="text-slate-500 uppercase font-sans">Avg:</span>
            <span class="text-cyan-300 font-bold ml-1">{{ summaryStats.avg !== null ? summaryStats.avg : '--' }}</span>
          </div>
        </div>
        <div class="flex items-center space-x-2">
          <span class="px-1.5 py-0.5 rounded bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
            GOOD: {{ summaryStats.goodCount }}
          </span>
          <span v-if="summaryStats.staleCount" class="px-1.5 py-0.5 rounded bg-purple-500/15 text-purple-400 border border-purple-500/30">
            STALE: {{ summaryStats.staleCount }}
          </span>
          <span v-if="summaryStats.uncertainCount" class="px-1.5 py-0.5 rounded bg-amber-500/15 text-amber-400 border border-amber-500/30">
            UNCERTAIN: {{ summaryStats.uncertainCount }}
          </span>
          <span v-if="summaryStats.badCount" class="px-1.5 py-0.5 rounded bg-rose-500/15 text-rose-400 border border-rose-500/30">
            BAD: {{ summaryStats.badCount }}
          </span>
        </div>
      </div>

      <!-- Body / Table -->
      <div class="p-4 overflow-y-auto max-h-[calc(92vh-190px)] custom-scrollbar">
        <!-- Loading State -->
        <div v-if="loading" class="py-16 flex flex-col items-center justify-center space-y-3 bg-[#0B0F19]/50 rounded-xl border border-slate-800">
          <div class="w-8 h-8 border-2 border-purple-500/20 border-t-purple-500 rounded-full animate-spin"></div>
          <span class="text-xs text-slate-300 font-medium font-sans">Memuat 100 data terakhir...</span>
        </div>

        <!-- Error State -->
        <div v-else-if="error" class="p-6 text-center space-y-3 bg-rose-950/20 border border-rose-500/30 rounded-xl">
          <p class="text-xs text-rose-300 font-mono">{{ error }}</p>
          <button
            type="button"
            @click="fetchHistory"
            class="px-3 py-1.5 rounded-lg text-xs font-bold text-white bg-rose-600 hover:bg-rose-500 transition"
          >
            Coba Lagi
          </button>
        </div>

        <!-- Empty State -->
        <div v-else-if="records.length === 0" class="py-16 text-center text-slate-400 font-mono space-y-1">
          <p class="text-sm font-semibold">Belum ada riwayat data telemetri</p>
          <p class="text-3xs text-slate-500 font-sans">Data telemetri akan muncul di sini setelah sensor mengirimkan pembacaan.</p>
        </div>

        <!-- Records Table -->
        <div v-else class="rounded-xl border border-slate-800 overflow-hidden shadow-md">
          <table class="w-full text-left text-xs font-mono border-collapse bg-[#0B0F19]">
            <thead class="bg-[#080E18] text-3xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-800 sticky top-0 z-10">
              <tr>
                <th class="py-2.5 px-3 text-center w-12">#</th>
                <th class="py-2.5 px-3">Timestamp</th>
                <th class="py-2.5 px-3 text-right">Engineering Value</th>
                <th class="py-2.5 px-3 text-right">Raw PDU</th>
                <th class="py-2.5 px-3">Raw Hex</th>
                <th class="py-2.5 px-3 text-center">Quality</th>
                <th class="py-2.5 px-3">Quality Reason</th>
                <th class="py-2.5 px-3 text-center">Source</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/80 text-2xs">
              <tr
                v-for="(rec, idx) in records"
                :key="rec.id || ('rec-' + idx)"
                class="hover:bg-slate-800/40 transition-colors"
                :class="rec.quality === 'BAD' ? 'bg-rose-950/15' : (rec.quality === 'UNCERTAIN' ? 'bg-amber-950/15' : '')"
              >
                <!-- Row # -->
                <td class="py-2 px-3 text-center text-slate-500 font-mono text-3xs">
                  {{ idx + 1 }}
                </td>

                <!-- Timestamp -->
                <td class="py-2 px-3 whitespace-nowrap">
                  <div class="text-slate-200 font-bold">{{ formatTimestamp(rec.received_at || rec.timestamp) }}</div>
                  <div class="text-4xs text-slate-500 font-sans">{{ formatRelativeTime(rec.received_at || rec.timestamp) }}</div>
                </td>

                <!-- Engineering Value -->
                <td class="py-2 px-3 text-right whitespace-nowrap">
                  <span class="text-xs font-bold font-mono" :class="getValueColorClass(rec.quality)">
                    {{ formatRecordValue(rec.value) }}
                  </span>
                  <span class="text-3xs text-slate-400 ml-1 font-mono">{{ targetParam.unit || '' }}</span>
                  <span v-if="rec.is_held_value" class="ml-1 px-1 py-0.2 rounded bg-amber-500/15 text-amber-400 text-4xs font-bold" title="Held Value Active">
                    HELD
                  </span>
                </td>

                <!-- Raw Value -->
                <td class="py-2 px-3 text-right whitespace-nowrap text-slate-300 font-mono">
                  {{ rec.raw_value !== undefined && rec.raw_value !== null ? rec.raw_value : '--' }}
                </td>

                <!-- Raw Hex -->
                <td class="py-2 px-3 whitespace-nowrap text-3xs text-cyan-300 font-mono">
                  {{ rec.raw_hex || '--' }}
                </td>

                <!-- Quality Badge -->
                <td class="py-2 px-3 text-center whitespace-nowrap">
                  <span
                    class="px-2 py-0.5 rounded text-4xs font-bold uppercase font-mono"
                    :class="getQualityBadgeClass(rec.quality)"
                  >
                    {{ rec.quality || 'GOOD' }}
                  </span>
                </td>

                <!-- Reason / Flags -->
                <td class="py-2 px-3 whitespace-nowrap text-3xs font-mono" :class="rec.quality_reason && rec.quality_reason !== 'NONE' ? 'text-amber-300' : 'text-slate-500'">
                  {{ rec.quality_reason || 'NONE' }}
                </td>

                <!-- Source -->
                <td class="py-2 px-3 text-center whitespace-nowrap text-3xs text-slate-400 font-mono">
                  {{ rec.source || 'MODBUS_RTU' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-3 border-t border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
        <div>
          <button
            type="button"
            @click="exportCsv"
            :disabled="records.length === 0"
            class="px-3.5 py-1.5 rounded-xl text-xs font-semibold text-slate-200 hover:text-white bg-slate-800 hover:bg-slate-700 border border-slate-700 transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path>
            </svg>
            <span>Export CSV</span>
          </button>
        </div>
        <div class="flex items-center space-x-2">
          <button
            type="button"
            @click="fetchHistory"
            :disabled="loading"
            class="px-3.5 py-1.5 rounded-xl text-xs font-semibold text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 border border-slate-700 transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5 text-cyan-400" :class="loading ? 'animate-spin' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>Refresh</span>
          </button>
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-1.5 rounded-xl text-xs font-medium text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            {{ $t('common.close') || 'Close' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'ParameterHistoryModal',
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
    device: {
      type: Object,
      default: () => ({}),
    },
    parameter: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    return {
      loading: false,
      error: null,
      records: [],
      total: 0,
    };
  },
  computed: {
    targetDevice() {
      return this.device || {};
    },
    targetParam() {
      return this.parameter || null;
    },
    precision() {
      if (this.targetParam && this.targetParam.precision !== undefined) {
        return this.targetParam.precision;
      }
      return 2;
    },
    latestValueDisplay() {
      if (!this.records || this.records.length === 0) return '--';
      const first = this.records[0];
      return `${this.formatRecordValue(first.value)} ${this.targetParam.unit || ''}`;
    },
    summaryStats() {
      if (!this.records || this.records.length === 0) {
        return { min: null, max: null, avg: null, goodCount: 0, badCount: 0, uncertainCount: 0, staleCount: 0 };
      }
      let min = Infinity;
      let max = -Infinity;
      let sum = 0;
      let count = 0;
      let goodCount = 0;
      let badCount = 0;
      let uncertainCount = 0;
      let staleCount = 0;

      for (const r of this.records) {
        const q = (r.quality || 'GOOD').toUpperCase();
        if (q === 'GOOD') goodCount++;
        else if (q === 'BAD') badCount++;
        else if (q === 'UNCERTAIN') uncertainCount++;
        else if (q === 'STALE') staleCount++;

        if (r.value !== null && r.value !== undefined && !isNaN(r.value)) {
          const val = Number(r.value);
          if (val < min) min = val;
          if (val > max) max = val;
          sum += val;
          count++;
        }
      }

      return {
        min: count > 0 ? min.toFixed(this.precision) : null,
        max: count > 0 ? max.toFixed(this.precision) : null,
        avg: count > 0 ? (sum / count).toFixed(this.precision) : null,
        goodCount,
        badCount,
        uncertainCount,
        staleCount,
      };
    },
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.fetchHistory();
      } else {
        this.records = [];
        this.total = 0;
        this.error = null;
      }
    },
  },
  methods: {
    closeModal() {
      this.$emit('close');
    },
    async fetchHistory() {
      if (!this.targetDevice || !this.targetParam) return;
      const devId = this.targetDevice.id || this.targetDevice.device_id;
      const paramId = this.targetParam.id || this.targetParam.param_id;
      if (!devId || !paramId) return;

      this.loading = true;
      this.error = null;
      try {
        const res = await axios.get(`/api/devices/${devId}/telemetry/history`, {
          params: {
            parameter_id: paramId,
            page: 1,
            page_size: 100,
          },
        });
        if (res.data && res.data.data) {
          this.records = res.data.data.items || [];
          this.total = res.data.data.total || this.records.length;
        } else {
          this.records = [];
          this.total = 0;
        }
      } catch (err) {
        console.error('Failed to fetch 100 historical readings:', err);
        this.error = (err.response && err.response.data && err.response.data.error) || err.message || 'Gagal memuat riwayat data';
      } finally {
        this.loading = false;
      }
    },
    formatRecordValue(val) {
      if (val === null || val === undefined || isNaN(val)) return '--';
      return Number(val).toFixed(this.precision);
    },
    formatTimestamp(ts) {
      if (!ts) return '--';
      try {
        const d = new Date(ts);
        return d.toLocaleString('en-GB', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
        });
      } catch (e) {
        return ts;
      }
    },
    formatRelativeTime(ts) {
      if (!ts) return '';
      try {
        const diffMs = Date.now() - new Date(ts).getTime();
        const diffSec = Math.floor(diffMs / 1000);
        if (diffSec < 60) return `${diffSec}s ago`;
        const diffMin = Math.floor(diffSec / 60);
        if (diffMin < 60) return `${diffMin}m ago`;
        const diffHr = Math.floor(diffMin / 60);
        if (diffHr < 24) return `${diffHr}h ago`;
        const diffDays = Math.floor(diffHr / 24);
        return `${diffDays}d ago`;
      } catch (e) {
        return '';
      }
    },
    getValueColorClass(quality) {
      const q = (quality || 'GOOD').toUpperCase();
      if (q === 'BAD') return 'text-rose-400';
      if (q === 'UNCERTAIN') return 'text-amber-400';
      if (q === 'STALE') return 'text-purple-400';
      return 'text-emerald-400';
    },
    getQualityBadgeClass(quality) {
      const q = (quality || 'GOOD').toUpperCase();
      if (q === 'BAD') return 'bg-rose-500/20 text-rose-400 border border-rose-500/40';
      if (q === 'UNCERTAIN') return 'bg-amber-500/20 text-amber-400 border border-amber-500/40';
      if (q === 'STALE') return 'bg-purple-500/20 text-purple-400 border border-purple-500/40';
      return 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40';
    },
    exportCsv() {
      if (!this.records || this.records.length === 0) return;
      const headers = ['Record ID', 'Timestamp', 'Parameter Code', 'Parameter Name', 'Value', 'Unit', 'Raw Value', 'Raw Hex', 'Quality', 'Quality Reason', 'Source'];
      const rows = this.records.map((r, idx) => [
        r.id || (idx + 1),
        `"${this.formatTimestamp(r.received_at || r.timestamp)}"`,
        `"${this.targetParam.parameter_code || this.targetParam.param_code || ''}"`,
        `"${this.targetParam.parameter_name || this.targetParam.param_name || ''}"`,
        this.formatRecordValue(r.value),
        `"${this.targetParam.unit || ''}"`,
        r.raw_value !== undefined ? r.raw_value : '',
        `"${r.raw_hex || ''}"`,
        `"${r.quality || 'GOOD'}"`,
        `"${r.quality_reason || 'NONE'}"`,
        `"${r.source || 'MODBUS_RTU'}"`,
      ]);

      const csvContent = [headers.join(','), ...rows.map(e => e.join(','))].join('\n');
      const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      const filename = `telemetry_${this.targetParam.parameter_code || 'param'}_100_latest_${Date.now()}.csv`;
      link.setAttribute('href', url);
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    },
  },
};
</script>
