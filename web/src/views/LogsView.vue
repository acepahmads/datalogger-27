<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="bg-slate-900 border border-slate-800 rounded p-3 flex items-center justify-between">
      <div class="flex items-center space-x-2">
        <span class="w-2 h-2 rounded-sm bg-blue-500"></span>
        <h1 class="text-xs font-bold text-slate-100 uppercase tracking-wider font-mono">
          System Traces & Audit Trails
        </h1>
      </div>

      <!-- Tab Switcher -->
      <div class="flex items-center space-x-1 bg-slate-950 p-0.5 rounded border border-slate-800 text-2xs font-mono">
        <button @click="activeTab = 'logs'"
                class="px-2.5 py-1 rounded transition-colors"
                :class="activeTab === 'logs' ? 'bg-blue-600 text-white font-semibold' : 'text-slate-400 hover:text-white'">
          System Logs
        </button>
        <button @click="activeTab = 'audit'"
                class="px-2.5 py-1 rounded transition-colors"
                :class="activeTab === 'audit' ? 'bg-blue-600 text-white font-semibold' : 'text-slate-400 hover:text-white'">
          Audit Trails
        </button>
      </div>
    </div>

    <!-- Logs Table -->
    <div v-if="activeTab === 'logs'" class="bg-slate-900 border border-slate-800 rounded overflow-hidden">
      <div class="p-3 border-b border-slate-800 flex items-center justify-between bg-slate-950/60 text-xs font-mono">
        <span class="text-slate-300 font-bold uppercase tracking-wider">Engine Diagnostic Traces (logs/datalogger.log)</span>
        <button @click="fetchLogs" class="text-2xs text-blue-400 hover:underline">Refresh</button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs border-collapse font-mono">
          <thead>
            <tr class="border-b border-slate-800 bg-slate-950/80 text-2xs text-slate-400 uppercase tracking-wider">
              <th class="p-2.5 w-36">Timestamp</th>
              <th class="p-2.5 w-20">Level</th>
              <th class="p-2.5 w-24">Component</th>
              <th class="p-2.5 font-sans">Message</th>
              <th class="p-2.5 font-sans">Details</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-slate-300">
            <tr v-for="log in systemLogs" :key="log.id" class="hover:bg-slate-800/40 transition-colors">
              <td class="p-2.5 text-slate-500 text-2xs">{{ log.created_at | formatDate }}</td>
              <td class="p-2.5">
                <span class="px-1.5 py-0.2 rounded text-2xs font-bold"
                      :class="log.level === 'ERROR' ? 'bg-rose-950 text-rose-400' : log.level === 'WARN' ? 'bg-amber-950 text-amber-400' : 'bg-blue-950 text-blue-400'">
                  {{ log.level }}
                </span>
              </td>
              <td class="p-2.5 text-slate-300 text-2xs">{{ log.component }}</td>
              <td class="p-2.5 font-sans text-slate-100 text-xs">{{ log.message }}</td>
              <td class="p-2.5 font-sans text-slate-400 text-2xs">{{ log.details }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Audit Trails Table -->
    <div v-else class="bg-slate-900 border border-slate-800 rounded overflow-hidden">
      <div class="p-3 border-b border-slate-800 flex items-center justify-between bg-slate-950/60 text-xs font-mono">
        <span class="text-slate-300 font-bold uppercase tracking-wider">Security & Configuration Change Journal</span>
        <button @click="fetchAudit" class="text-2xs text-blue-400 hover:underline">Refresh</button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs border-collapse font-mono">
          <thead>
            <tr class="border-b border-slate-800 bg-slate-950/80 text-2xs text-slate-400 uppercase tracking-wider">
              <th class="p-2.5 w-36">Timestamp</th>
              <th class="p-2.5 w-24">User</th>
              <th class="p-2.5 w-28">Action</th>
              <th class="p-2.5 w-36">Resource</th>
              <th class="p-2.5 font-sans">Details</th>
              <th class="p-2.5 w-24">IP</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60 text-slate-300">
            <tr v-for="trail in auditTrails" :key="trail.id" class="hover:bg-slate-800/40 transition-colors">
              <td class="p-2.5 text-slate-500 text-2xs">{{ trail.created_at | formatDate }}</td>
              <td class="p-2.5 text-blue-400 font-semibold text-2xs">{{ trail.username }}</td>
              <td class="p-2.5 text-slate-200 text-2xs">{{ trail.action }}</td>
              <td class="p-2.5 text-slate-300 text-2xs">{{ trail.resource }}</td>
              <td class="p-2.5 font-sans text-slate-300 text-xs">{{ trail.details }}</td>
              <td class="p-2.5 text-slate-500 text-2xs">{{ trail.ip_address }}</td>
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
  name: 'LogsView',
  data() {
    return {
      activeTab: 'logs',
      systemLogs: [],
      auditTrails: [],
    };
  },
  mounted() {
    this.fetchLogs();
    this.fetchAudit();
  },
  methods: {
    async fetchLogs() {
      try {
        const res = await axios.get('/api/logs?limit=50');
        if (res.data && res.data.data) {
          this.systemLogs = res.data.data;
        }
      } catch (e) {}
    },
    async fetchAudit() {
      try {
        const res = await axios.get('/api/audit-trails?limit=50');
        if (res.data && res.data.data) {
          this.auditTrails = res.data.data;
        }
      } catch (e) {}
    },
  },
};
</script>
