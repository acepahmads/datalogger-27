<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-[80] flex items-center justify-center p-4 bg-slate-950/65 backdrop-blur-xs font-sans"
    @click.self="closeModal"
  >
    <div class="bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden flex flex-col max-h-[92vh]">
      <!-- Header -->
      <div class="px-5 py-4 border-b border-slate-100 dark:border-slate-800 bg-slate-50/70 dark:bg-[#0B111E] flex items-center justify-between gap-3">
        <div class="flex items-center gap-3 min-w-0">
          <div class="w-9 h-9 rounded-xl bg-teal-500/15 border border-teal-500/30 text-teal-600 dark:text-teal-400 flex items-center justify-center flex-shrink-0">
            <svg class="w-4.5 h-4.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v2a2 2 0 002 2h12a2 2 0 002-2v-2M7 10l5 5 5-5M12 15V3" />
            </svg>
          </div>
          <div class="min-w-0">
            <h3 class="text-sm font-extrabold text-slate-900 dark:text-white truncate">
              Download Data Parameter
            </h3>
            <p class="text-2xs font-mono text-cyan-700 dark:text-cyan-400 truncate mt-0.5">
              {{ paramDisplayName }}
              <span v-if="paramCode"> ({{ paramCode }})</span>
              <span v-if="paramUnit"> · {{ paramUnit }}</span>
              <span v-if="deviceCode" class="text-slate-400"> · {{ deviceCode }}</span>
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="closeModal"
          :disabled="downloading"
          class="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-200/60 dark:hover:bg-slate-800 transition cursor-pointer disabled:opacity-40"
          aria-label="Close"
        >
          <svg class="w-4.5 h-4.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Body -->
      <div class="p-5 space-y-4 overflow-y-auto text-xs">
        <!-- Quick Time Presets -->
        <div>
          <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5">
            Preset Rentang Waktu
          </label>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="preset in presets"
              :key="preset.id"
              type="button"
              @click="applyPreset(preset.id)"
              :disabled="downloading"
              class="px-2.5 py-1.5 rounded-lg text-2xs font-semibold border transition cursor-pointer disabled:opacity-50"
              :class="activePreset === preset.id
                ? 'bg-teal-600 text-white btn-text border-teal-600 shadow-2xs'
                : 'bg-slate-50 dark:bg-slate-800/70 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:border-teal-400 dark:hover:border-teal-500'"
            >
              {{ preset.label }}
            </button>
          </div>
        </div>

        <!-- Custom Start / End DateTime -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <label for="param-dl-start" class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1">
              Waktu Mulai
            </label>
            <input
              id="param-dl-start"
              v-model="startLocal"
              type="datetime-local"
              :disabled="downloading"
              @change="onManualDateChange"
              class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 font-mono text-2xs focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-500"
            />
          </div>
          <div>
            <label for="param-dl-end" class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1">
              Waktu Selesai
            </label>
            <input
              id="param-dl-end"
              v-model="endLocal"
              type="datetime-local"
              :disabled="downloading"
              @change="onManualDateChange"
              class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 font-mono text-2xs focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-500"
            />
          </div>
        </div>

        <!-- Quality Filter & File Format -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <label for="param-dl-quality" class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1">
              Filter Kualitas Data
            </label>
            <select
              id="param-dl-quality"
              v-model="qualityFilter"
              :disabled="downloading"
              @change="fetchCountPreview"
              class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 text-2xs font-medium focus:outline-none focus:ring-2 focus:ring-teal-500/30 focus:border-teal-500"
            >
              <option value="">Semua Kualitas</option>
              <option value="GOOD">GOOD (Valid)</option>
              <option value="UNCERTAIN">UNCERTAIN</option>
              <option value="BAD">BAD</option>
              <option value="STALE">STALE</option>
            </select>
          </div>
          <div>
            <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1">
              Format File
            </label>
            <div class="grid grid-cols-3 gap-1.5">
              <button
                v-for="fmt in formatOptions"
                :key="fmt.id"
                type="button"
                @click="fileFormat = fmt.id"
                :disabled="downloading"
                class="py-2 px-2 rounded-xl text-2xs font-bold font-mono border transition cursor-pointer disabled:opacity-50"
                :class="fileFormat === fmt.id
                  ? 'bg-cyan-600 text-white btn-text border-cyan-600 shadow-2xs'
                  : 'bg-slate-50 dark:bg-slate-800/70 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:border-cyan-400'"
              >
                {{ fmt.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- Live Count Preview Strip -->
        <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800 flex items-center justify-between gap-2">
          <div class="flex items-center gap-2 min-w-0">
            <span
              class="w-2 h-2 rounded-full flex-shrink-0"
              :class="previewLoading ? 'bg-amber-400 animate-pulse' : (previewTotal > 0 ? 'bg-emerald-500' : 'bg-slate-400')"
            ></span>
            <span v-if="previewLoading" class="text-2xs text-slate-500 dark:text-slate-400">
              Menghitung jumlah data pada rentang waktu...
            </span>
            <span v-else-if="dateRangeError" class="text-2xs font-semibold text-rose-600 dark:text-rose-400">
              {{ dateRangeError }}
            </span>
            <span v-else class="text-2xs text-slate-600 dark:text-slate-300">
              Ditemukan <strong class="font-mono text-slate-900 dark:text-white">{{ formatNumber(previewTotal) }}</strong> baris data
              <span v-if="previewTotal > maxDownloadRows" class="text-amber-600 dark:text-amber-400">
                (maks. {{ formatNumber(maxDownloadRows) }} baris per unduhan)
              </span>
            </span>
          </div>
          <button
            type="button"
            @click="fetchCountPreview"
            :disabled="previewLoading || downloading || Boolean(dateRangeError)"
            class="px-2 py-1 rounded-lg text-3xs font-semibold border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer disabled:opacity-40 flex-shrink-0"
          >
            Cek Ulang
          </button>
        </div>

        <!-- Download Progress Bar -->
        <div v-if="downloading" class="space-y-1.5">
          <div class="flex items-center justify-between text-2xs font-mono text-teal-700 dark:text-teal-300">
            <span>Mengunduh data...</span>
            <span>{{ formatNumber(downloadedCount) }} / {{ formatNumber(downloadTargetTotal || downloadedCount) }}</span>
          </div>
          <div class="w-full h-2 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
            <div
              class="h-full bg-gradient-to-r from-teal-500 to-cyan-500 transition-all duration-200"
              :style="{ width: progressPercent + '%' }"
            ></div>
          </div>
        </div>

        <!-- Error / Success Feedback -->
        <div
          v-if="errorMsg"
          class="p-3 rounded-xl bg-rose-50 dark:bg-rose-500/10 border border-rose-200 dark:border-rose-500/30 text-2xs text-rose-700 dark:text-rose-300 font-medium"
        >
          {{ errorMsg }}
        </div>
        <div
          v-if="successMsg"
          class="p-3 rounded-xl bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-200 dark:border-emerald-500/30 text-2xs text-emerald-700 dark:text-emerald-300 font-medium"
        >
          {{ successMsg }}
        </div>
      </div>

      <!-- Footer -->
      <div class="px-5 py-3.5 border-t border-slate-100 dark:border-slate-800 bg-slate-50/70 dark:bg-[#0B111E] flex items-center justify-end gap-2">
        <button
          type="button"
          @click="closeModal"
          :disabled="downloading"
          class="px-4 py-2 rounded-xl border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-semibold transition cursor-pointer disabled:opacity-40"
        >
          Tutup
        </button>
        <button
          id="btn-confirm-download-param-data"
          type="button"
          @click="executeDownload"
          :disabled="downloading || Boolean(dateRangeError) || (!previewLoading && previewTotal === 0)"
          class="px-4 py-2 rounded-xl bg-teal-600 hover:bg-teal-500 text-white btn-text text-xs font-bold shadow-sm transition cursor-pointer disabled:opacity-45 disabled:cursor-not-allowed flex items-center gap-2"
        >
          <span v-if="downloading" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
          <span>{{ downloading ? 'Memproses...' : `Download ${fileFormat.toUpperCase()}` }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

const MAX_DOWNLOAD_ROWS = 50000;
const PAGE_SIZE = 1000;

function toLocalDateTimeInput(dateObj) {
  const d = dateObj instanceof Date ? dateObj : new Date(dateObj);
  if (Number.isNaN(d.getTime())) return '';
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function formatFilenameStamp(dateObj) {
  const d = dateObj instanceof Date ? dateObj : new Date(dateObj);
  if (Number.isNaN(d.getTime())) return 'range';
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}-${pad(d.getHours())}${pad(d.getMinutes())}`;
}

function escapeCsvCell(val) {
  const str = val === null || val === undefined ? '' : String(val);
  if (/[",\r\n]/.test(str)) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

function escapeHtmlCell(val) {
  return String(val === null || val === undefined ? '' : val)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

export default {
  name: 'ParameterDataDownloadModal',
  props: {
    isOpen: { type: Boolean, default: false },
    device: { type: Object, default: () => ({}) },
    parameter: { type: Object, default: null },
  },
  data() {
    return {
      activePreset: '24h',
      startLocal: '',
      endLocal: '',
      qualityFilter: '',
      fileFormat: 'csv',
      previewLoading: false,
      previewTotal: 0,
      downloading: false,
      downloadedCount: 0,
      downloadTargetTotal: 0,
      errorMsg: '',
      successMsg: '',
      maxDownloadRows: MAX_DOWNLOAD_ROWS,
      presets: [
        { id: '1h', label: '1 Jam' },
        { id: '6h', label: '6 Jam' },
        { id: '24h', label: '24 Jam' },
        { id: 'today', label: 'Hari Ini' },
        { id: '7d', label: '7 Hari' },
        { id: '30d', label: '30 Hari' },
      ],
      formatOptions: [
        { id: 'csv', label: 'CSV' },
        { id: 'xls', label: 'Excel' },
        { id: 'json', label: 'JSON' },
      ],
    };
  },
  computed: {
    userRole() {
      const u = (this.$store && this.$store.state && this.$store.state.user) || {};
      const r = typeof u.role === 'string' ? u.role : ((u.role && u.role.name) || '');
      return String(r || '').trim().toUpperCase();
    },
    isOperator() {
      return !this.userRole || this.userRole === 'OPERATOR';
    },
    targetDeviceId() {
      return Number(
        (this.device && (this.device.id || this.device.device_id)) ||
        (this.parameter && this.parameter.device_id) ||
        0
      );
    },
    targetParamId() {
      return Number((this.parameter && (this.parameter.id || this.parameter.param_id)) || 0);
    },
    deviceCode() {
      return (this.device && (this.device.device_code || this.device.code)) ||
        (this.parameter && this.parameter.device_code) || '';
    },
    paramCode() {
      return (this.parameter && (this.parameter.parameter_code || this.parameter.param_code || this.parameter.code)) || '';
    },
    paramDisplayName() {
      return (this.parameter && (this.parameter.parameter_name || this.parameter.param_name || this.parameter.name)) ||
        this.paramCode || 'Parameter';
    },
    paramUnit() {
      return (this.parameter && this.parameter.unit) || '';
    },
    precision() {
      const p = this.parameter && this.parameter.precision;
      return Number.isFinite(Number(p)) ? Number(p) : 2;
    },
    parsedStartDate() {
      return this.startLocal ? new Date(this.startLocal) : null;
    },
    parsedEndDate() {
      return this.endLocal ? new Date(this.endLocal) : null;
    },
    dateRangeError() {
      if (!this.parsedStartDate || !this.parsedEndDate) return 'Pilih waktu mulai dan selesai.';
      if (Number.isNaN(this.parsedStartDate.getTime()) || Number.isNaN(this.parsedEndDate.getTime())) {
        return 'Format tanggal/waktu tidak valid.';
      }
      if (this.parsedEndDate.getTime() <= this.parsedStartDate.getTime()) {
        return 'Waktu selesai harus lebih besar dari waktu mulai.';
      }
      return '';
    },
    progressPercent() {
      if (!this.downloadTargetTotal) return this.downloading ? 30 : 0;
      return Math.min(100, Math.round((this.downloadedCount / this.downloadTargetTotal) * 100));
    },
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.errorMsg = '';
        this.successMsg = '';
        this.downloadedCount = 0;
        this.downloadTargetTotal = 0;
        this.qualityFilter = '';
        this.applyPreset('24h');
      }
    },
  },
  methods: {
    closeModal() {
      if (this.downloading) return;
      this.$emit('close');
    },
    formatNumber(n) {
      return Number(n || 0).toLocaleString('id-ID');
    },
    formatTimestamp(ts) {
      if (!ts) return '';
      const d = new Date(ts);
      if (Number.isNaN(d.getTime())) return String(ts);
      const pad = (v) => String(v).padStart(2, '0');
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
    },
    formatValue(val) {
      if (val === null || val === undefined || Number.isNaN(Number(val))) return '';
      return Number(val).toFixed(this.precision);
    },
    applyPreset(presetId) {
      this.activePreset = presetId;
      const now = new Date();
      let start = new Date(now.getTime() - 24 * 3600 * 1000);
      if (presetId === '1h') {
        start = new Date(now.getTime() - 1 * 3600 * 1000);
      } else if (presetId === '6h') {
        start = new Date(now.getTime() - 6 * 3600 * 1000);
      } else if (presetId === '24h') {
        start = new Date(now.getTime() - 24 * 3600 * 1000);
      } else if (presetId === 'today') {
        start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0, 0);
      } else if (presetId === '7d') {
        start = new Date(now.getTime() - 7 * 24 * 3600 * 1000);
      } else if (presetId === '30d') {
        start = new Date(now.getTime() - 30 * 24 * 3600 * 1000);
      }
      this.startLocal = toLocalDateTimeInput(start);
      this.endLocal = toLocalDateTimeInput(now);
      this.errorMsg = '';
      this.successMsg = '';
      this.fetchCountPreview();
    },
    onManualDateChange() {
      this.activePreset = 'custom';
      this.errorMsg = '';
      this.successMsg = '';
      if (!this.dateRangeError) {
        this.fetchCountPreview();
      } else {
        this.previewTotal = 0;
      }
    },
    buildQueryParams(page, pageSize, skipCount = false) {
      const params = {
        parameter_id: this.targetParamId,
        start_time: this.parsedStartDate.toISOString(),
        end_time: this.parsedEndDate.toISOString(),
        page,
        page_size: pageSize,
      };
      if (this.qualityFilter) {
        params.quality = this.qualityFilter;
      }
      if (skipCount) {
        params.skip_count = '1';
      }
      return params;
    },
    async fetchCountPreview() {
      if (this.dateRangeError || !this.targetDeviceId || !this.targetParamId) {
        this.previewTotal = 0;
        return;
      }
      this.previewLoading = true;
      this.errorMsg = '';
      try {
        const res = await axios.get(`/api/devices/${this.targetDeviceId}/telemetry/history`, {
          params: this.buildQueryParams(1, 1, false),
        });
        const data = (res.data && res.data.data) || {};
        this.previewTotal = Number(data.total || (Array.isArray(data.items) ? data.items.length : 0));
      } catch (err) {
        this.previewTotal = 0;
        const msg = (err.response && err.response.data && (err.response.data.error || err.response.data.message)) || err.message;
        this.errorMsg = `Gagal memeriksa jumlah data: ${msg}`;
      } finally {
        this.previewLoading = false;
      }
    },
    async executeDownload() {
      if (this.downloading || this.dateRangeError || !this.targetDeviceId || !this.targetParamId) return;
      this.downloading = true;
      this.errorMsg = '';
      this.successMsg = '';
      this.downloadedCount = 0;
      this.downloadTargetTotal = Math.min(this.previewTotal || PAGE_SIZE, MAX_DOWNLOAD_ROWS);

      try {
        const allItems = [];
        let page = 1;
        let totalReported = this.previewTotal || 0;

        while (allItems.length < MAX_DOWNLOAD_ROWS) {
          const res = await axios.get(`/api/devices/${this.targetDeviceId}/telemetry/history`, {
            params: this.buildQueryParams(page, PAGE_SIZE, page > 1),
          });
          const data = (res.data && res.data.data) || {};
          const items = Array.isArray(data.items) ? data.items : [];
          if (page === 1 && Number(data.total) > 0) {
            totalReported = Number(data.total);
            this.downloadTargetTotal = Math.min(totalReported, MAX_DOWNLOAD_ROWS);
          }
          if (items.length === 0) break;
          for (const item of items) {
            if (allItems.length >= MAX_DOWNLOAD_ROWS) break;
            allItems.push(item);
          }
          this.downloadedCount = allItems.length;
          if (items.length < PAGE_SIZE) break;
          if (totalReported > 0 && allItems.length >= totalReported) break;
          page += 1;
        }

        if (allItems.length === 0) {
          this.errorMsg = 'Tidak ada data telemetri pada rentang waktu yang dipilih.';
          return;
        }

        const safeDev = String(this.deviceCode || `dev${this.targetDeviceId}`).replace(/[^a-zA-Z0-9_-]/g, '_');
        const safeParam = String(this.paramCode || `param${this.targetParamId}`).replace(/[^a-zA-Z0-9_-]/g, '_');
        const stampStart = formatFilenameStamp(this.parsedStartDate);
        const stampEnd = formatFilenameStamp(this.parsedEndDate);
        const baseName = `telemetry_${safeDev}_${safeParam}_${stampStart}_to_${stampEnd}`;

        if (this.fileFormat === 'json') {
          this.triggerJsonDownload(allItems, `${baseName}.json`);
        } else if (this.fileFormat === 'xls') {
          this.triggerXlsDownload(allItems, `${baseName}.xls`);
        } else {
          this.triggerCsvDownload(allItems, `${baseName}.csv`);
        }

        this.successMsg = `Berhasil mengunduh ${this.formatNumber(allItems.length)} baris data (${this.fileFormat.toUpperCase()}).`;
      } catch (err) {
        const msg = (err.response && err.response.data && (err.response.data.error || err.response.data.message)) || err.message;
        this.errorMsg = `Gagal mengunduh data: ${msg}`;
      } finally {
        this.downloading = false;
      }
    },
    buildExportTable(records) {
      const headers = this.isOperator
        ? ['Record ID', 'Timestamp', 'Device Code', 'Parameter Code', 'Parameter Name', 'Value', 'Unit', 'Quality', 'Quality Reason', 'Source']
        : ['Record ID', 'Timestamp', 'Device Code', 'Parameter Code', 'Parameter Name', 'Value', 'Unit', 'Raw Value', 'Raw Hex', 'Quality', 'Quality Reason', 'Source'];

      const rows = records.map((r, idx) => {
        const isBad = String(r.quality || '').toUpperCase() === 'BAD';
        const row = [
          r.id || (idx + 1),
          this.formatTimestamp(r.received_at || r.timestamp),
          this.deviceCode,
          this.paramCode,
          this.paramDisplayName,
          isBad ? '' : this.formatValue(r.value),
          this.paramUnit,
        ];
        if (!this.isOperator) {
          row.push(r.raw_value !== undefined && r.raw_value !== null ? r.raw_value : '');
          row.push(r.raw_hex || '');
        }
        row.push(r.quality || 'GOOD');
        row.push(r.quality_reason || 'NONE');
        row.push(r.source || '');
        return row;
      });

      return { headers, rows };
    },
    triggerBlobDownload(blob, filename) {
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.setAttribute('href', url);
      link.setAttribute('download', filename);
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      setTimeout(() => URL.revokeObjectURL(url), 1000);
    },
    triggerCsvDownload(records, filename) {
      const { headers, rows } = this.buildExportTable(records);
      const csvLines = [
        headers.map(escapeCsvCell).join(','),
        ...rows.map(r => r.map(escapeCsvCell).join(',')),
      ];
      const blob = new Blob(['\uFEFF' + csvLines.join('\n')], { type: 'text/csv;charset=utf-8;' });
      this.triggerBlobDownload(blob, filename);
    },
    triggerXlsDownload(records, filename) {
      const { headers, rows } = this.buildExportTable(records);
      const thead = `<tr>${headers.map(h => `<th style="background:#0f766e;color:#ffffff;font-weight:bold;padding:6px;">${escapeHtmlCell(h)}</th>`).join('')}</tr>`;
      const tbody = rows.map(r => `<tr>${r.map(c => `<td style="padding:4px 6px;border:1px solid #cbd5e1;">${escapeHtmlCell(c)}</td>`).join('')}</tr>`).join('');
      const html = `<!DOCTYPE html><html><head><meta charset="utf-8" /></head><body><table border="1">${thead}${tbody}</table></body></html>`;
      const blob = new Blob(['\uFEFF' + html], { type: 'application/vnd.ms-excel;charset=utf-8;' });
      this.triggerBlobDownload(blob, filename);
    },
    triggerJsonDownload(records, filename) {
      const cleanRecords = records.map((r) => {
        const entry = {
          id: r.id,
          timestamp: r.received_at || r.timestamp,
          value: String(r.quality || '').toUpperCase() === 'BAD' ? null : r.value,
          unit: this.paramUnit,
          quality: r.quality || 'GOOD',
          quality_reason: r.quality_reason || 'NONE',
          source: r.source || '',
        };
        if (!this.isOperator) {
          entry.raw_value = r.raw_value;
          entry.raw_hex = r.raw_hex || '';
        }
        return entry;
      });
      const payload = {
        device_code: this.deviceCode,
        parameter_code: this.paramCode,
        parameter_name: this.paramDisplayName,
        unit: this.paramUnit,
        start_time: this.parsedStartDate.toISOString(),
        end_time: this.parsedEndDate.toISOString(),
        quality_filter: this.qualityFilter || 'ALL',
        total_records: cleanRecords.length,
        records: cleanRecords,
      };
      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json;charset=utf-8;' });
      this.triggerBlobDownload(blob, filename);
    },
  },
};
</script>
