<template>
  <div class="space-y-6 font-sans">
    <!-- Header Section -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center space-x-2 text-xs text-slate-400 mb-1">
          <router-link to="/monitoring/devices" class="hover:text-blue-400 flex items-center space-x-1">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
            <span>Monitoring</span>
          </router-link>
          <span>/</span>
          <span class="text-emerald-400 font-semibold">Live Telemetry & Sensors</span>
        </div>
        <div class="flex items-center space-x-3">
          <h1 class="text-xl font-bold tracking-tight text-white flex items-center space-x-2">
            <span>Incoming Sensor Telemetry</span>
            <span class="relative flex h-3 w-3">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
            </span>
          </h1>
          <span class="px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-3xs font-mono font-bold">
            Realtime Ingestion Active
          </span>
        </div>
        <p class="text-xs text-slate-400 mt-1">
          Unified real-time telemetry stream across all connected Modbus RTU/TCP devices and field sensors.
        </p>
      </div>

      <!-- Quick Actions -->
      <div class="flex items-center space-x-2.5">
        <button
          @click="fetchTelemetry"
          :disabled="loading"
          class="px-3.5 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition flex items-center space-x-2 disabled:opacity-50"
        >
          <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>Refresh Data</span>
        </button>

        <router-link
          to="/monitoring/devices"
          class="px-3.5 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition flex items-center space-x-1.5 shadow-lg shadow-blue-600/20"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="2"></circle>
            <path d="M16.24 7.76a6 6 0 0 1 0 8.49m-8.48-.01a6 6 0 0 1 0-8.49m11.31-2.82a10 10 0 0 1 0 14.14m-14.14 0a10 10 0 0 1 0-14.14"></path>
          </svg>
          <span>Manage Devices</span>
        </router-link>
      </div>
    </div>

    <!-- Summary KPI Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="saas-card p-4 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-400">Connected Devices</div>
        <div class="flex items-baseline space-x-2">
          <span class="text-2xl font-mono font-bold text-white">{{ stats.activeDevices }}</span>
          <span class="text-2xs text-slate-400">/ {{ stats.totalDevices }} configured</span>
        </div>
        <div class="text-3xs text-emerald-400 font-mono">Modbus engines running</div>
      </div>

      <div class="saas-card p-4 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-400">Total Monitored Sensors</div>
        <div class="flex items-baseline space-x-2">
          <span class="text-2xl font-mono font-bold text-blue-400">{{ parametersList.length }}</span>
          <span class="text-2xs text-slate-400">parameters</span>
        </div>
        <div class="text-3xs text-slate-400">Auto-polled periodically</div>
      </div>

      <div class="saas-card p-4 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-400">Signal Quality (Good)</div>
        <div class="flex items-baseline space-x-2">
          <span class="text-2xl font-mono font-bold text-emerald-400">{{ stats.goodSensors }}</span>
          <span class="text-2xs text-slate-400">healthy</span>
        </div>
        <div class="text-3xs text-emerald-400">CRC & Modbus valid</div>
      </div>

      <div class="saas-card p-4 space-y-1 bg-[#0F172A]/70 border border-slate-800">
        <div class="text-3xs uppercase tracking-wider font-semibold text-slate-400">Last Telemetry Ingest</div>
        <div class="flex items-baseline space-x-2">
          <span class="text-lg font-mono font-bold text-slate-200">{{ lastPacketTime ? formatRelative(lastPacketTime) : 'Listening...' }}</span>
        </div>
        <div class="text-3xs text-slate-400 font-mono truncate">{{ lastPacketTime || 'Awaiting incoming packet' }}</div>
      </div>
    </div>

    <!-- Filters & Controls Bar -->
    <div class="p-4 bg-[#0F172A]/90 border border-slate-800 rounded-2xl flex flex-col md:flex-row gap-4 items-stretch md:items-center justify-between">
      <div class="flex flex-wrap items-center gap-3">
        <!-- Search -->
        <div class="relative min-w-[220px]">
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search parameter, code, unit..."
            class="w-full pl-9 pr-4 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
          />
          <svg class="w-4 h-4 text-slate-500 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>

        <!-- Filter Device -->
        <div>
          <select
            v-model="selectedDeviceId"
            class="px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500"
          >
            <option value="">All Devices ({{ devices.length }})</option>
            <option v-for="d in devices" :key="d.id" :value="d.id">
              {{ d.device_code || d.code }} - {{ d.device_name || d.name }}
            </option>
          </select>
        </div>

        <!-- Filter Quality -->
        <div>
          <select
            v-model="selectedQuality"
            class="px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
          >
            <option value="">All Qualities</option>
            <option value="GOOD">GOOD Quality Only</option>
            <option value="BAD">BAD / Error Only</option>
            <option value="UNCERTAIN">UNCERTAIN Only</option>
          </select>
        </div>
      </div>

      <!-- View Switcher -->
      <div class="flex items-center space-x-2 self-end md:self-auto">
        <span class="text-3xs text-slate-400 uppercase font-semibold mr-1">Display:</span>
        <button
          @click="viewMode = 'grid'"
          :class="viewMode === 'grid' ? 'bg-blue-600 text-white' : 'bg-slate-800 text-slate-400 hover:text-white'"
          class="p-2 rounded-lg transition"
          title="Sensor Cards Grid"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <rect x="3" y="3" width="7" height="7" rx="1"></rect>
            <rect x="14" y="3" width="7" height="7" rx="1"></rect>
            <rect x="14" y="14" width="7" height="7" rx="1"></rect>
            <rect x="3" y="14" width="7" height="7" rx="1"></rect>
          </svg>
        </button>
        <button
          @click="viewMode = 'table'"
          :class="viewMode === 'table' ? 'bg-blue-600 text-white' : 'bg-slate-800 text-slate-400 hover:text-white'"
          class="p-2 rounded-lg transition"
          title="Telemetry Data Table"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <line x1="3" y1="6" x2="21" y2="6"></line>
            <line x1="3" y1="12" x2="21" y2="12"></line>
            <line x1="3" y1="18" x2="21" y2="18"></line>
          </svg>
        </button>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="filteredParameters.length === 0" class="saas-card p-12 text-center space-y-3 bg-[#0F172A]/50 border border-slate-800">
      <div class="w-12 h-12 rounded-full bg-slate-800 flex items-center justify-center mx-auto text-slate-400">
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="9"></circle>
          <line x1="12" y1="8" x2="12" y2="12"></line>
          <line x1="12" y1="16" x2="12.01" y2="16"></line>
        </svg>
      </div>
      <h3 class="text-sm font-bold text-white">No incoming sensor parameters match filters</h3>
      <p class="text-xs text-slate-400 max-w-md mx-auto">
        Verify that devices are registered and parameters have been configured in Devices & Sensors.
      </p>
      <div class="pt-2">
        <router-link to="/monitoring/devices" class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold inline-block">
          Open Devices Configuration
        </router-link>
      </div>
    </div>

    <!-- VIEW 1: SENSOR CARDS GRID -->
    <div v-else-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
      <div
        v-for="item in filteredParameters"
        :key="item.param_id"
        :class="[
          liveBlinks[item.param_id] ? 'border-emerald-500 shadow-lg shadow-emerald-500/10' : 'border-slate-800/80',
          item.quality === 'BAD' ? 'bg-rose-950/10' : 'bg-[#0F172A]'
        ]"
        class="saas-card p-4 space-y-3 rounded-2xl border transition-all duration-300 hover:border-slate-700 flex flex-col justify-between"
      >
        <div>
          <!-- Card Header: Device + Parameter Code -->
          <div class="flex items-start justify-between gap-2">
            <div>
              <div class="flex items-center space-x-1.5">
                <span class="w-1.5 h-1.5 rounded-full" :class="item.device_status === 'CONNECTED' ? 'bg-emerald-400' : 'bg-slate-500'"></span>
                <router-link :to="'/monitoring/devices/' + item.device_id" class="text-3xs font-mono font-bold text-slate-400 hover:text-blue-400 uppercase tracking-wide">
                  {{ item.device_code }}
                </router-link>
              </div>
              <h3 class="text-xs font-bold text-white mt-0.5 truncate" :title="item.param_name">
                {{ item.param_name }}
              </h3>
            </div>

            <!-- Quality Badge -->
            <span
              :class="qualityBadgeClass(item.quality)"
              class="px-2 py-0.5 rounded text-3xs font-mono font-bold uppercase flex items-center space-x-1 flex-shrink-0"
            >
              <span class="w-1 h-1 rounded-full" :class="item.quality === 'GOOD' ? 'bg-emerald-400' : 'bg-rose-400'"></span>
              <span>{{ item.quality || 'WAITING' }}</span>
            </span>
          </div>

          <!-- Parameter Code Tag -->
          <div class="mt-1 flex items-center space-x-2">
            <span class="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 font-mono text-3xs">
              {{ item.param_code }}
            </span>
            <span v-if="item.data_type" class="text-3xs font-mono text-slate-500">
              {{ item.data_type }}
            </span>
          </div>

          <!-- Primary Metric Value -->
          <div class="mt-4 p-3 bg-[#0B0F19] rounded-xl border border-slate-800/80 flex items-baseline justify-between">
            <div class="text-2xl font-mono font-extrabold tracking-tight" :class="valueColorClass(item)">
              {{ formatValue(item) }}
            </div>
            <div class="text-xs font-bold text-slate-400 font-mono">
              {{ item.unit || '--' }}
            </div>
          </div>
        </div>

        <!-- Card Footer -->
        <div class="pt-2 border-t border-slate-800/80 space-y-2">
          <div class="flex items-center justify-between text-3xs font-mono text-slate-400">
            <span>Last packet:</span>
            <span class="text-slate-300">{{ item.received_at ? formatRelative(item.received_at) : 'No data yet' }}</span>
          </div>

          <!-- Quick Navigation Actions -->
          <div class="flex items-center justify-between gap-1 pt-1">
            <router-link
              :to="'/monitoring/devices/' + item.device_id + '?tab=telemetry'"
              class="text-3xs text-emerald-400 hover:text-emerald-300 font-semibold px-2 py-1 rounded bg-emerald-950/30 border border-emerald-800/40 hover:bg-emerald-900/40 transition"
            >
              Live Telemetry
            </router-link>
            <div class="flex items-center space-x-1">
              <router-link
                :to="'/monitoring/devices/' + item.device_id + '?tab=history'"
                class="text-3xs text-blue-400 hover:text-blue-300 font-medium px-2 py-1 rounded bg-blue-950/30 border border-blue-800/40 hover:bg-blue-900/40 transition"
              >
                Chart
              </router-link>
              <router-link
                :to="'/monitoring/devices/' + item.device_id + '?tab=raw'"
                class="text-3xs text-slate-400 hover:text-slate-200 font-medium px-2 py-1 rounded bg-slate-800/60 hover:bg-slate-700 transition"
              >
                Raw
              </router-link>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- VIEW 2: TELEMETRY DATA TABLE -->
    <div v-else class="saas-card overflow-hidden border border-slate-800 bg-[#0F172A] rounded-2xl">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse text-xs">
          <thead>
            <tr class="border-b border-slate-800 bg-[#0B0F19]/60 text-slate-400 text-3xs font-semibold uppercase tracking-wider">
              <th class="py-3 px-4">Device</th>
              <th class="py-3 px-4">Parameter Name</th>
              <th class="py-3 px-4">Code</th>
              <th class="py-3 px-4 text-right">Current Value</th>
              <th class="py-3 px-4">Unit</th>
              <th class="py-3 px-4 text-center">Quality</th>
              <th class="py-3 px-4">Last Timestamp</th>
              <th class="py-3 px-4 text-right">Inspect</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60">
            <tr
              v-for="item in filteredParameters"
              :key="item.param_id"
              :class="liveBlinks[item.param_id] ? 'bg-emerald-500/10' : 'hover:bg-slate-800/30'"
              class="transition-colors"
            >
              <!-- Device -->
              <td class="py-3 px-4 font-mono font-medium text-slate-200">
                <router-link :to="'/monitoring/devices/' + item.device_id" class="text-blue-400 hover:underline">
                  {{ item.device_code }}
                </router-link>
                <div class="text-3xs text-slate-400 font-sans truncate max-w-[140px]">{{ item.device_name }}</div>
              </td>

              <!-- Parameter Name -->
              <td class="py-3 px-4 font-medium text-white">
                {{ item.param_name }}
              </td>

              <!-- Code -->
              <td class="py-3 px-4 font-mono text-3xs text-slate-300">
                <span class="px-1.5 py-0.5 rounded bg-slate-800">
                  {{ item.param_code }}
                </span>
              </td>

              <!-- Value -->
              <td class="py-3 px-4 text-right font-mono font-bold text-sm" :class="valueColorClass(item)">
                {{ formatValue(item) }}
              </td>

              <!-- Unit -->
              <td class="py-3 px-4 font-mono text-2xs text-slate-400">
                {{ item.unit || '--' }}
              </td>

              <!-- Quality -->
              <td class="py-3 px-4 text-center">
                <span
                  :class="qualityBadgeClass(item.quality)"
                  class="px-2 py-0.5 rounded text-3xs font-mono font-bold uppercase inline-block"
                >
                  {{ item.quality || 'WAITING' }}
                </span>
              </td>

              <!-- Timestamp -->
              <td class="py-3 px-4 font-mono text-3xs text-slate-400">
                {{ item.received_at ? formatRelative(item.received_at) : '--' }}
                <div class="text-slate-500">{{ item.received_at ? formatTimestamp(item.received_at) : '' }}</div>
              </td>

              <!-- Actions -->
              <td class="py-3 px-4 text-right space-x-1.5">
                <router-link
                  :to="'/monitoring/devices/' + item.device_id + '?tab=telemetry'"
                  class="text-emerald-400 hover:text-emerald-300 text-3xs font-semibold px-2 py-1 rounded bg-emerald-950/40 border border-emerald-800/40"
                >
                  Live
                </router-link>
                <router-link
                  :to="'/monitoring/devices/' + item.device_id + '?tab=history'"
                  class="text-blue-400 hover:text-blue-300 text-3xs font-semibold px-2 py-1 rounded bg-blue-950/40 border border-blue-800/40"
                >
                  History
                </router-link>
              </td>
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
  name: 'TelemetryMonitorView',
  data() {
    return {
      devices: [],
      liveMap: {},
      liveBlinks: {},
      loading: false,
      searchQuery: '',
      selectedDeviceId: '',
      selectedQuality: '',
      viewMode: 'grid',
      lastPacketTime: null,
      pollTimer: null,
    };
  },
  computed: {
    parametersList() {
      const list = [];
      for (const dev of this.devices) {
        if (!dev.parameters) continue;
        for (const p of dev.parameters) {
          const live = this.liveMap[p.id] || {};
          list.push({
            device_id: dev.id,
            device_code: dev.device_code || dev.code,
            device_name: dev.device_name || dev.name,
            device_status: dev.connection_status,
            param_id: p.id,
            param_code: p.parameter_code || p.code,
            param_name: p.parameter_name || p.name,
            unit: p.unit,
            data_type: p.data_type,
            precision: p.precision !== undefined ? p.precision : 2,
            value: live.value !== undefined ? live.value : p.current_value,
            value_text: live.value_text,
            quality: live.quality || p.current_quality || 'UNKNOWN',
            received_at: live.received_at || p.current_received_at || p.last_updated,
          });
        }
      }
      return list;
    },
    filteredParameters() {
      return this.parametersList.filter(item => {
        if (this.selectedDeviceId && item.device_id !== parseInt(this.selectedDeviceId, 10)) {
          return false;
        }
        if (this.selectedQuality && item.quality !== this.selectedQuality) {
          return false;
        }
        if (this.searchQuery) {
          const q = this.searchQuery.toLowerCase();
          const matchName = item.param_name && item.param_name.toLowerCase().includes(q);
          const matchCode = item.param_code && item.param_code.toLowerCase().includes(q);
          const matchDevice = item.device_code && item.device_code.toLowerCase().includes(q);
          const matchUnit = item.unit && item.unit.toLowerCase().includes(q);
          if (!matchName && !matchCode && !matchDevice && !matchUnit) return false;
        }
        return true;
      });
    },
    stats() {
      const activeDevices = this.devices.filter(d => d.connection_status === 'CONNECTED').length;
      const goodSensors = this.parametersList.filter(p => p.quality === 'GOOD').length;
      return {
        totalDevices: this.devices.length,
        activeDevices,
        goodSensors,
      };
    },
  },
  mounted() {
    this.fetchTelemetry();
    window.addEventListener('device-telemetry-event', this.onTelemetryEvent);
    this.pollTimer = setInterval(this.fetchLatestTelemetryAll, 10000);
  },
  beforeDestroy() {
    window.removeEventListener('device-telemetry-event', this.onTelemetryEvent);
    if (this.pollTimer) clearInterval(this.pollTimer);
  },
  methods: {
    async fetchTelemetry() {
      this.loading = true;
      try {
        const res = await axios.get('/api/devices?limit=100');
        this.devices = res.data.data.items || [];
        await this.fetchLatestTelemetryAll();
      } catch (err) {
        console.error('Failed to load devices/telemetry:', err);
      } finally {
        this.loading = false;
      }
    },
    async fetchLatestTelemetryAll() {
      for (const dev of this.devices) {
        try {
          const res = await axios.get(`/api/devices/${dev.id}/telemetry/latest`);
          if (res.data && res.data.data) {
            for (const item of res.data.data) {
              this.$set(this.liveMap, item.parameter_id, item);
              if (item.received_at && (!this.lastPacketTime || item.received_at > this.lastPacketTime)) {
                this.lastPacketTime = item.received_at;
              }
            }
          }
        } catch (e) {
          // ignore error per device
        }
      }
    },
    onTelemetryEvent(e) {
      if (!e.detail) return;
      const data = e.detail;
      this.$set(this.liveMap, data.parameter_id, data);
      this.$set(this.liveBlinks, data.parameter_id, true);
      this.lastPacketTime = data.received_at;
      setTimeout(() => {
        this.$set(this.liveBlinks, data.parameter_id, false);
      }, 1200);
    },
    formatValue(item) {
      if (item.value_text) return item.value_text;
      if (item.value !== null && item.value !== undefined) {
        return Number(item.value).toFixed(item.precision || 2);
      }
      return '--';
    },
    valueColorClass(item) {
      if (item.quality === 'BAD') return 'text-rose-400';
      if (item.quality === 'UNCERTAIN') return 'text-amber-400';
      if (item.value !== null && item.value !== undefined) return 'text-emerald-400';
      return 'text-slate-400';
    },
    qualityBadgeClass(quality) {
      switch (quality) {
        case 'GOOD': return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
        case 'BAD': return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
        case 'UNCERTAIN': return 'bg-amber-500/10 text-amber-400 border border-amber-500/20';
        default: return 'bg-slate-500/10 text-slate-400 border border-slate-500/20';
      }
    },
    formatTimestamp(ts) {
      if (!ts) return '--';
      try {
        const d = new Date(ts);
        return d.toLocaleTimeString('id-ID', { hour12: false });
      } catch (e) {
        return ts;
      }
    },
    formatRelative(ts) {
      if (!ts) return '--';
      try {
        const d = new Date(ts);
        const diff = Math.floor((new Date() - d) / 1000);
        if (diff < 2) return 'just now';
        if (diff < 60) return `${diff}s ago`;
        if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
        return `${Math.floor(diff / 3600)}h ago`;
      } catch (e) {
        return ts;
      }
    },
  },
};
</script>
