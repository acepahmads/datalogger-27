<template>
  <div class="space-y-6 font-sans">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans">
          Alarm Console & Notifications
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          Threshold monitoring, abnormality detection, and operator acknowledgement workflows.
        </p>
      </div>

      <div class="flex items-center space-x-2">
        <span class="px-3 py-1.5 rounded-lg text-xs font-semibold font-mono"
              :class="activeCount > 0 ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20' : 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'">
          {{ activeCount }} ACTIVE ALARMS
        </span>
      </div>
    </div>

    <!-- Alarm Table in saas-card -->
    <div class="saas-card overflow-hidden">
      <div class="px-5 py-3.5 border-b border-slate-800/80 flex items-center justify-between">
        <h2 class="text-sm font-bold text-white font-sans tracking-tight">Active & Historical Incidents</h2>
        <button @click="refresh" class="text-2xs text-blue-400 hover:text-blue-300 font-sans font-medium">Refresh</button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs border-collapse font-sans">
          <thead>
            <tr class="border-b border-slate-800/80 bg-[#0B0F19]/60 text-3xs font-semibold text-slate-400 uppercase tracking-wider font-sans">
              <th class="py-3 px-4 w-28">Severity</th>
              <th class="py-3 px-4 w-32">Alarm Code</th>
              <th class="py-3 px-4">Message</th>
              <th class="py-3 px-4 w-28">Device</th>
              <th class="py-3 px-4 w-40">Triggered At</th>
              <th class="py-3 px-4 w-28">Status</th>
              <th class="py-3 px-4 w-28 text-center">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/50 text-slate-300">
            <tr v-if="alarms.length === 0">
              <td colspan="7" class="py-8 text-center text-slate-500 font-sans">
                No active alarms detected. All industrial devices operating normally.
              </td>
            </tr>
            <tr v-for="alarm in alarms" :key="alarm.id" class="hover:bg-slate-800/25 transition-colors">
              <td class="py-3 px-4">
                <span class="px-2 py-0.5 rounded-md text-3xs font-bold font-mono tracking-wider"
                      :class="alarm.severity === 'CRITICAL' ? 'bg-rose-500/10 text-rose-400 border border-rose-500/20' : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'">
                  {{ alarm.severity }}
                </span>
              </td>
              <td class="py-3 px-4 text-slate-200 font-semibold font-mono text-2xs">{{ alarm.alarm_code }}</td>
              <td class="py-3 px-4 text-xs text-slate-100 font-sans">{{ alarm.message }}</td>
              <td class="py-3 px-4 text-blue-400 font-mono text-2xs">{{ alarm.device ? alarm.device.code : 'System' }}</td>
              <td class="py-3 px-4 text-slate-400 font-mono text-3xs">{{ alarm.triggered_at | formatDate }}</td>
              <td class="py-3 px-4">
                <span class="px-2 py-0.5 rounded-md text-3xs font-semibold font-mono"
                      :class="alarm.status === 'ACTIVE' ? 'bg-rose-500/10 text-rose-400' : 'bg-slate-800 text-slate-400'">
                  {{ alarm.status }}
                </span>
              </td>
              <td class="py-3 px-4 text-center">
                <button v-if="alarm.status === 'ACTIVE'"
                        @click="ackAlarm(alarm.id)"
                        class="px-2.5 py-1 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 hover:text-white rounded-lg text-2xs font-sans transition-colors">
                  Acknowledge
                </button>
                <span v-else class="text-slate-500 text-3xs font-mono">Acked</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AlarmsView',
  computed: {
    alarms() {
      return this.$store.state.alarms;
    },
    activeCount() {
      return this.alarms.filter((a) => a.status === 'ACTIVE').length;
    },
  },
  mounted() {
    this.refresh();
  },
  methods: {
    refresh() {
      this.$store.dispatch('fetchAlarms');
    },
    async ackAlarm(id) {
      await this.$store.dispatch('acknowledgeAlarm', id);
    },
  },
};
</script>
