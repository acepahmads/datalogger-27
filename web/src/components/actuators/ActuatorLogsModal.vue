<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/70 backdrop-blur-sm">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl max-w-4xl w-full overflow-hidden flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-900/50">
        <div class="flex items-center space-x-3">
          <div class="w-9 h-9 rounded-lg bg-slate-500/10 dark:bg-slate-400/10 border border-slate-500/20 text-slate-600 dark:text-slate-300 flex items-center justify-center text-sm font-bold font-mono">
            📜
          </div>
          <div>
            <h3 class="font-bold text-slate-800 dark:text-slate-100 text-sm">
              Actuator Command Audit Logs
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              Audit trail of manual overrides, scheduled cycles, and watchdog actions
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="close"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1 rounded-md"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 overflow-y-auto space-y-4">
        <div v-if="loading" class="py-12 text-center text-xs text-slate-400">
          Loading audit logs...
        </div>

        <div v-else-if="logs.length === 0" class="py-12 text-center text-slate-400 dark:text-slate-500 text-xs">
          <p class="text-xl mb-1">📜</p>
          <p>No actuation logs recorded yet.</p>
        </div>

        <div v-else class="overflow-x-auto border border-slate-200 dark:border-slate-800 rounded-xl">
          <table class="w-full text-left text-xs font-sans">
            <thead class="bg-slate-50 dark:bg-slate-800/60 text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800 text-3xs uppercase font-mono tracking-wider">
              <tr>
                <th class="py-2.5 px-3">Executed At</th>
                <th class="py-2.5 px-3">Action</th>
                <th class="py-2.5 px-3">Triggered By</th>
                <th class="py-2.5 px-3">Modbus PDU</th>
                <th class="py-2.5 px-3">Value</th>
                <th class="py-2.5 px-3">Latency</th>
                <th class="py-2.5 px-3">Status</th>
                <th class="py-2.5 px-3 text-right">Raw Hex</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-2xs">
              <tr v-for="log in logs" :key="log.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40">
                <td class="py-2 px-3 text-slate-600 dark:text-slate-300 whitespace-nowrap">
                  {{ formatTime(log.executed_at) }}
                </td>
                <td class="py-2 px-3 whitespace-nowrap font-bold">
                  <span
                    :class="log.action === 'ON' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30' : 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400 border-slate-200 dark:border-slate-700'"
                    class="px-2 py-0.5 rounded-full text-3xs border"
                  >
                    {{ log.action }}
                  </span>
                </td>
                <td class="py-2 px-3 text-slate-700 dark:text-slate-200 font-sans font-medium whitespace-nowrap">
                  {{ log.triggered_by }}
                </td>
                <td class="py-2 px-3 text-slate-500 whitespace-nowrap">
                  FC0{{ log.function_code }} @ #{{ log.address }} (Slave #{{ log.slave_id }})
                </td>
                <td class="py-2 px-3 text-slate-600 dark:text-slate-300 whitespace-nowrap">
                  {{ formatHex(log.payload_written) }}
                </td>
                <td class="py-2 px-3 text-amber-600 dark:text-amber-400 whitespace-nowrap">
                  {{ log.latency_ms }} ms
                </td>
                <td class="py-2 px-3 whitespace-nowrap">
                  <span
                    :class="log.status === 'SUCCESS' ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'"
                    class="font-bold text-3xs"
                  >
                    {{ log.status }}
                  </span>
                </td>
                <td class="py-2 px-3 text-right whitespace-nowrap">
                  <span
                    :title="`TX: ${log.raw_request_hex || '--'}\nRX: ${log.raw_response_hex || '--'}`"
                    class="cursor-pointer text-blue-600 dark:text-blue-400 hover:underline text-3xs"
                  >
                    Inspect
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-3 border-t border-slate-100 dark:border-slate-800 flex justify-between items-center bg-slate-50/50 dark:bg-slate-900/50">
        <span class="text-3xs text-slate-400 font-mono">Showing up to 100 recent actuation events</span>
        <button
          type="button"
          @click="close"
          class="px-4 py-1.5 text-xs font-semibold rounded-lg text-slate-700 dark:text-slate-200 hover:bg-slate-200/60 dark:hover:bg-slate-800 transition-colors"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'ActuatorLogsModal',
  props: {
    isOpen: { type: Boolean, default: false },
    actuatorId: { type: [Number, String], default: null },
    deviceId: { type: [Number, String], default: null }
  },
  data() {
    return {
      logs: [],
      loading: false
    };
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.fetchLogs();
      }
    }
  },
  methods: {
    close() {
      this.$emit('close');
    },
    formatTime(tStr) {
      if (!tStr) return '--';
      const d = new Date(tStr);
      return d.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    formatHex(val) {
      if (val === undefined || val === null) return '0x0000';
      return '0x' + Number(val).toString(16).toUpperCase().padStart(4, '0');
    },
    async fetchLogs() {
      this.loading = true;
      try {
        const params = { limit: 100 };
        if (this.actuatorId) params.actuator_id = this.actuatorId;
        if (this.deviceId) params.device_id = this.deviceId;

        const res = await axios.get('/api/actuators/logs', { params });
        this.logs = res.data.data || [];
      } catch (err) {
        console.error('Failed to fetch actuation logs:', err);
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>
