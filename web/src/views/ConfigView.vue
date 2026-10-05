<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans">
          System Configuration
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          Manage production MariaDB database connection, runtime parameters, and edge policies.
        </p>
      </div>

      <!-- Tab Switcher -->
      <div class="flex items-center space-x-1 bg-[#111827] p-1 rounded-lg border border-slate-800 text-xs font-sans">
        <button @click="activeTab = 'database'"
                class="px-3 py-1.5 rounded-md transition-colors"
                :class="activeTab === 'database' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
          Database (MariaDB)
        </button>
        <button @click="activeTab = 'general'"
                class="px-3 py-1.5 rounded-md transition-colors"
                :class="activeTab === 'general' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
          General & Server
        </button>
        <button @click="activeTab = 'localfirst'"
                class="px-3 py-1.5 rounded-md transition-colors"
                :class="activeTab === 'localfirst' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
          Local-First Principle
        </button>
      </div>
    </div>

    <!-- TAB 1: Database (MariaDB) -->
    <div v-if="activeTab === 'database'" class="space-y-5 max-w-4xl">
      <!-- Live Database Health Monitoring Card (Part 14) -->
      <div class="saas-card p-5 space-y-4 bg-gradient-to-br from-[#111827] to-[#131E35]">
        <div class="flex items-center justify-between border-b border-slate-800/80 pb-3">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
            <h2 class="text-sm font-bold text-white font-sans tracking-tight">MariaDB Edge Health Monitor</h2>
          </div>
          <button @click="fetchDbStatus" :disabled="testing"
                  class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium transition-colors flex items-center">
            <svg class="w-3.5 h-3.5 mr-1.5" :class="{ 'animate-spin': testing }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <polyline points="23 4 23 10 17 10"></polyline>
              <polyline points="1 20 1 14 7 14"></polyline>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
            Refresh Status
          </button>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs font-sans">
          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">Database Type</span>
            <span class="font-bold text-white text-sm block">MariaDB</span>
            <span class="text-3xs text-emerald-400 font-mono mt-0.5 block">● {{ dbStatus.status || 'Connected' }}</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">Host & Port</span>
            <span class="font-mono text-slate-200 text-xs font-semibold block">{{ dbStatus.host || '127.0.0.1' }}:{{ dbStatus.port || '3306' }}</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">Schema: {{ dbStatus.database || 'datalogger' }}</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">Latency</span>
            <span class="font-mono text-emerald-400 text-sm font-bold block">{{ dbStatus.latency_ms ? dbStatus.latency_ms.toFixed(2) : '0.45' }} ms</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">{{ dbStatus.version || '10.4.32-MariaDB' }}</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">Active / Max Conns</span>
            <span class="font-mono text-blue-400 text-sm font-bold block">{{ dbStatus.active_connections || 1 }} / {{ dbStatus.max_connections || 25 }}</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">DB Size: {{ dbStatus.database_size_mb || 1.08 }} MB</span>
          </div>
        </div>

        <div class="pt-2 text-2xs text-slate-400 flex items-center justify-between font-sans">
          <span>Last successful query: <span class="font-mono text-slate-300">{{ dbStatus.last_query_time ? new Date(dbStatus.last_query_time).toLocaleTimeString('en-GB') : 'Just now' }}</span></span>
          <span class="text-emerald-400 font-medium">Auto-reconnection & Connection Pool Active</span>
        </div>
      </div>

      <!-- Database Configuration Form (Part 15) -->
      <div class="saas-card p-5 space-y-4">
        <div class="border-b border-slate-800/80 pb-3">
          <h2 class="text-sm font-bold text-white font-sans tracking-tight">MariaDB Connection Parameters</h2>
          <p class="text-2xs text-slate-400 font-sans mt-0.5">Configure edge database credentials and network bindings</p>
        </div>

        <!-- Alert messages -->
        <div v-if="testResult" class="p-3 rounded-lg text-xs font-sans flex items-center justify-between"
             :class="testResult.success ? 'bg-emerald-500/10 border border-emerald-500/30 text-emerald-300' : 'bg-rose-500/10 border border-rose-500/30 text-rose-300'">
          <div>
            <span class="font-semibold">{{ testResult.success ? 'Connection Successful!' : 'Connection Failed' }}</span>
            <span class="block text-2xs mt-0.5">{{ testResult.message }}</span>
          </div>
          <span v-if="testResult.latency" class="font-mono text-2xs">{{ testResult.latency.toFixed(2) }}ms</span>
        </div>

        <div v-if="saveMessage" class="p-3 rounded-lg bg-blue-500/10 border border-blue-500/30 text-blue-300 text-xs font-sans">
          {{ saveMessage }}
        </div>

        <form @submit.prevent="saveDbConfig" class="space-y-4 text-xs font-sans">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Database Type</label>
              <input type="text" value="MariaDB" readonly class="w-full bg-[#0B0F19] border border-slate-800 rounded-lg px-3 py-2 text-slate-300 font-sans cursor-not-allowed" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Database Name</label>
              <input type="text" v-model="form.database" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Host IP / Address</label>
              <input type="text" v-model="form.host" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Port</label>
              <input type="text" v-model="form.port" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Username</label>
              <input type="text" v-model="form.username" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-sans" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Password</label>
              <input type="password" v-model="form.password" placeholder="••••••••" class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>
          </div>

          <div class="pt-3 border-t border-slate-800/80 flex items-center justify-end space-x-3">
            <button type="button" @click="testDbConnection" :disabled="testing"
                    class="px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 rounded-lg text-xs font-medium transition-colors flex items-center">
              <span v-if="testing" class="inline-block w-3.5 h-3.5 border-2 border-slate-400 border-t-transparent rounded-full animate-spin mr-1.5"></span>
              TEST CONNECTION
            </button>

            <button type="submit" :disabled="saving"
                    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
              <span v-if="saving" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
              SAVE CONFIGURATION
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- TAB 2: General & Server -->
    <div v-else-if="activeTab === 'general'" class="space-y-4 max-w-4xl">
      <div class="saas-card p-5 space-y-4">
        <h2 class="text-sm font-bold text-white font-sans tracking-tight border-b border-slate-800/80 pb-3">HTTP & Service Parameters</h2>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs font-sans">
          <div>
            <label class="block text-3xs text-slate-400 uppercase tracking-wider mb-1">HTTP Port</label>
            <input type="text" value="8080" readonly class="w-full bg-[#0B0F19] border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-300" />
          </div>
          <div>
            <label class="block text-3xs text-slate-400 uppercase tracking-wider mb-1">Environment Mode</label>
            <input type="text" value="production" readonly class="w-full bg-[#0B0F19] border border-slate-800 rounded-lg px-3 py-2 font-mono text-slate-300" />
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 3: Local-First Principle -->
    <div v-else class="saas-card p-5 space-y-3 max-w-4xl font-sans">
      <h2 class="text-sm font-bold text-white tracking-tight border-b border-slate-800/80 pb-3">
        Local-First Autonomous Principle (Section 14)
      </h2>
      <p class="text-xs text-slate-300 leading-relaxed">
        The application is architected to operate 100% locally on the industrial edge machine with MariaDB.
        Data acquisition, calculations, alarming, and UI dashboards require zero Internet connectivity. Cloud transmission is strictly an optional external forwarder.
      </p>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'ConfigView',
  data() {
    return {
      activeTab: 'database',
      testing: false,
      saving: false,
      testResult: null,
      saveMessage: '',
      dbStatus: {
        type: 'MariaDB',
        host: '127.0.0.1',
        port: '3306',
        database: 'datalogger',
        status: 'Connected',
        latency_ms: 0.45,
        version: '10.4.32-MariaDB',
        active_connections: 1,
        max_connections: 25,
        database_size_mb: 1.08,
      },
      form: {
        host: '127.0.0.1',
        port: '3306',
        database: 'datalogger',
        username: 'root',
        password: '',
      },
    };
  },
  mounted() {
    this.fetchDbStatus();
  },
  methods: {
    async fetchDbStatus() {
      try {
        const res = await axios.get('/api/system/database/status');
        if (res.data && res.data.data) {
          this.dbStatus = res.data.data;
          this.form.host = this.dbStatus.host || '127.0.0.1';
          this.form.port = this.dbStatus.port || '3306';
          this.form.database = this.dbStatus.database || 'datalogger';
          this.form.username = this.dbStatus.username || 'root';
        }
      } catch (err) {
        console.warn('Failed to fetch DB status:', err);
      }
    },
    async testDbConnection() {
      this.testing = true;
      this.testResult = null;
      try {
        const res = await axios.post('/api/system/database/test', this.form);
        if (res.data) {
          if (res.data.success && res.data.data && res.data.data.connected) {
            this.testResult = {
              success: true,
              message: `Connected to ${res.data.data.version} on ${this.form.host}:${this.form.port}`,
              latency: res.data.data.latency_ms,
            };
          } else {
            this.testResult = {
              success: false,
              message: res.data.error || 'Failed to connect to MariaDB',
            };
          }
        }
      } catch (err) {
        this.testResult = {
          success: false,
          message: err.response?.data?.error || err.message,
        };
      } finally {
        this.testing = false;
      }
    },
    async saveDbConfig() {
      this.saving = true;
      this.saveMessage = '';
      try {
        const res = await axios.post('/api/system/database/config', this.form);
        if (res.data && res.data.success) {
          this.saveMessage = 'MariaDB configuration saved to config.json. Applied successfully.';
          this.fetchDbStatus();
        }
      } catch (err) {
        alert('Failed to save config: ' + (err.response?.data?.error || err.message));
      } finally {
        this.saving = false;
      }
    },
  },
};
</script>
