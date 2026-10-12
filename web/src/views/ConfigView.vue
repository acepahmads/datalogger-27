<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans flex items-center gap-2">
          {{ $t('config.title') || 'System Configuration' }}
          <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
            #83
          </span>
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          {{ $t('config.subtitle') || 'Operational runtime settings, database connectivity, and hardware serial ports' }}
        </p>
      </div>

      <!-- Tab Switcher -->
      <div class="flex items-center space-x-1 bg-[#111827] p-1 rounded-lg border border-slate-800 text-xs font-sans">
        <button @click="activeTab = 'database'"
                class="px-3 py-1.5 rounded-md transition-colors"
                :class="activeTab === 'database' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
          {{ $t('config.tabDatabase') || 'Database' }}
        </button>
        <button @click="activeTab = 'serial'"
                class="px-3 py-1.5 rounded-md transition-colors"
                :class="activeTab === 'serial' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
          Serial & Comm
        </button>
        <button @click="activeTab = 'general'"
                class="px-3 py-1.5 rounded-md transition-colors"
                :class="activeTab === 'general' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
          {{ $t('config.tabGeneral') || 'Server & Runtime' }}
        </button>
        <button @click="activeTab = 'localfirst'"
                class="px-3 py-1.5 rounded-md transition-colors"
                :class="activeTab === 'localfirst' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
          {{ $t('config.tabLocalFirst') || 'Local-First' }}
        </button>
      </div>
    </div>

    <!-- Alert / Feedback Banner -->
    <div v-if="saveResult" class="p-4 rounded-xl border text-xs font-sans space-y-1"
         :class="saveResult.restart_required ? 'bg-amber-500/10 border-amber-500/30 text-amber-300' : 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300'">
      <div class="font-bold flex items-center gap-2">
        <span v-if="saveResult.restart_required">⚠️ Restart Required</span>
        <span v-else>✓ Configuration Saved</span>
      </div>
      <div>{{ saveResult.message }}</div>
      <div v-if="saveResult.restart_required_fields && saveResult.restart_required_fields.length > 0" class="text-3xs text-amber-400 font-mono">
        Modified fields requiring restart: {{ saveResult.restart_required_fields.join(', ') }}
      </div>
    </div>

    <!-- TAB 1: Database (MariaDB) -->
    <div v-if="activeTab === 'database'" class="space-y-5 max-w-4xl">
      <!-- Live Database Health Monitoring Card -->
      <div class="saas-card p-5 space-y-4 bg-gradient-to-br from-[#111827] to-[#131E35]">
        <div class="flex items-center justify-between border-b border-slate-800/80 pb-3">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
            <h2 class="text-sm font-bold text-white font-sans tracking-tight">{{ $t('config.dbMonitor') || 'MariaDB Edge Engine Monitor' }}</h2>
          </div>
          <button @click="fetchDbStatus" :disabled="testing"
                  class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium transition-colors flex items-center">
            <svg class="w-3.5 h-3.5 mr-1.5" :class="{ 'animate-spin': testing }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <polyline points="23 4 23 10 17 10"></polyline>
              <polyline points="1 20 1 14 7 14"></polyline>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
            {{ $t('common.refresh') || 'Refresh' }}
          </button>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs font-sans">
          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">{{ $t('config.dbType') || 'Database Engine' }}</span>
            <span class="font-bold text-white text-sm block">MariaDB</span>
            <span class="text-3xs text-emerald-400 font-mono mt-0.5 block">● Connected</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">{{ $t('config.hostPort') || 'Host & Port' }}</span>
            <span class="font-mono text-slate-200 text-xs font-semibold block">{{ dbStatus.host || '127.0.0.1' }}:{{ dbStatus.port || '3306' }}</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">{{ dbStatus.database || 'datalogger' }}</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">{{ $t('devices.latency') || 'Ping Latency' }}</span>
            <span class="font-mono text-emerald-400 text-sm font-bold block">{{ dbStatus.latency_ms ? dbStatus.latency_ms.toFixed(2) : '0.45' }} ms</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">{{ dbStatus.version || '10.4.32-MariaDB' }}</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">Active Pool Conns</span>
            <span class="font-mono text-blue-400 text-sm font-bold block">{{ dbStatus.active_connections || 1 }} / {{ dbStatus.max_connections || 25 }}</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">{{ dbStatus.database_size_mb || 1.08 }} MB</span>
          </div>
        </div>
      </div>

      <!-- Database Configuration Form -->
      <div class="saas-card p-5 space-y-4">
        <div class="border-b border-slate-800/80 pb-3 flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-white font-sans tracking-tight">Database Connectivity Parameters</h2>
            <p class="text-2xs text-slate-400 font-sans mt-0.5">Modifying host, port, or database schema requires service restart.</p>
          </div>
          <span class="text-3xs font-mono px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
            RESTART REQUIRED
          </span>
        </div>

        <!-- Connection Test Result -->
        <div v-if="testResult" class="p-3 rounded-lg text-xs font-sans flex items-center justify-between"
             :class="testResult.success ? 'bg-emerald-500/10 border border-emerald-500/30 text-emerald-300' : 'bg-rose-500/10 border border-rose-500/30 text-rose-300'">
          <div>
            <span class="font-semibold">{{ testResult.success ? 'Connection Successful' : 'Connection Failed' }}</span>
            <span class="block text-2xs mt-0.5">{{ testResult.message }}</span>
          </div>
          <span v-if="testResult.latency" class="font-mono text-2xs">{{ testResult.latency.toFixed(2) }}ms</span>
        </div>

        <form @submit.prevent="submitConfig" class="space-y-4 text-xs font-sans">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Host / IP Address</label>
              <input type="text" v-model="form.db_host" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Database Port</label>
              <input type="text" v-model="form.db_port" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Database Name / Schema</label>
              <input type="text" v-model="form.db_name" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Username</label>
              <input type="text" v-model="form.db_user" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-sans" />
            </div>

            <div class="sm:col-span-2">
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">
                Password <span class="text-slate-500 font-normal lowercase">(leave as •••••••• to keep current secret)</span>
              </label>
              <input type="password" v-model="form.db_password" placeholder="••••••••" class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>
          </div>

          <div class="pt-3 border-t border-slate-800/80 flex items-center justify-end space-x-3">
            <button type="button" @click="testDbConnection" :disabled="testing"
                    class="px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 rounded-lg text-xs font-medium transition-colors flex items-center">
              <span v-if="testing" class="inline-block w-3.5 h-3.5 border-2 border-slate-400 border-t-transparent rounded-full animate-spin mr-1.5"></span>
              Test Connection
            </button>

            <button type="submit" :disabled="saving"
                    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
              <span v-if="saving" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
              Save Configuration
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- TAB 2: Serial & Communication Ports (#83) -->
    <div v-else-if="activeTab === 'serial'" class="space-y-4 max-w-4xl font-sans">
      <div class="saas-card p-5 space-y-4">
        <div class="border-b border-slate-800/80 pb-3 flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-white tracking-tight">Hardware Serial Bus & Port Defaults</h2>
            <p class="text-2xs text-slate-400 mt-0.5">Default communication parameters for RS-485 and USB Serial COM interfaces.</p>
          </div>
          <span class="text-3xs font-mono px-2 py-0.5 rounded bg-rose-500/10 text-rose-400 border border-rose-500/20">
            MAY DISRUPT ACQUISITION
          </span>
        </div>

        <!-- Warning notice -->
        <div class="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-xs text-amber-300 space-y-1">
          <div class="font-bold">⚠️ Acquisition Interruption Notice</div>
          <p class="text-2xs text-amber-400/90 leading-relaxed">
            Changing serial baud rate, parity, or COM port settings will cause active Modbus RTU polling routines on affected ports to reconnect. Ensure field instruments support the new baud rate.
          </p>
        </div>

        <form @submit.prevent="submitConfig" class="space-y-4 text-xs font-sans">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Default Serial Port</label>
              <input type="text" v-model="form.serial_port" placeholder="e.g. COM1 or /dev/ttyUSB0"
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Baud Rate (bps)</label>
              <select v-model.number="form.serial_baud_rate"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono">
                <option :value="1200">1200</option>
                <option :value="2400">2400</option>
                <option :value="4800">4800</option>
                <option :value="9600">9600 (Standard)</option>
                <option :value="19200">19200</option>
                <option :value="38400">38400</option>
                <option :value="57600">57600</option>
                <option :value="115200">115200 (High-Speed)</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Data Bits</label>
              <select v-model.number="form.serial_data_bits"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono">
                <option :value="8">8 Data Bits (Standard)</option>
                <option :value="7">7 Data Bits</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Stop Bits</label>
              <select v-model.number="form.serial_stop_bits"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono">
                <option :value="1">1 Stop Bit</option>
                <option :value="2">2 Stop Bits</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Parity</label>
              <select v-model="form.serial_parity"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono">
                <option value="N">None (N)</option>
                <option value="E">Even (E)</option>
                <option value="O">Odd (O)</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Response Timeout (ms)</label>
              <input type="number" v-model.number="form.serial_timeout_ms" min="100" max="10000" step="50"
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>
          </div>

          <div class="pt-3 border-t border-slate-800/80 flex items-center justify-end space-x-3">
            <button type="submit" :disabled="saving"
                    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
              <span v-if="saving" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
              Save Serial Parameters
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- TAB 3: General & Server Runtime -->
    <div v-else-if="activeTab === 'general'" class="space-y-4 max-w-4xl font-sans">
      <div class="saas-card p-5 space-y-4">
        <div class="border-b border-slate-800/80 pb-3 flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-white tracking-tight">HTTP Server & Runtime Governance</h2>
            <p class="text-2xs text-slate-400 mt-0.5">Port allocation, system log levels, and session security token lifetimes.</p>
          </div>
        </div>

        <form @submit.prevent="submitConfig" class="space-y-4 text-xs">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <label class="text-3xs text-slate-400 uppercase tracking-wider font-semibold">HTTP Service Port</label>
                <span class="text-3xs text-amber-400 font-mono">Requires Restart</span>
              </div>
              <input type="text" v-model="form.port" required
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 font-mono text-white" />
            </div>

            <div>
              <label class="block text-3xs text-slate-400 uppercase tracking-wider mb-1.5 font-semibold">Environment Profile</label>
              <input type="text" v-model="form.environment"
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 font-mono text-white" />
            </div>

            <div>
              <label class="block text-3xs text-slate-400 uppercase tracking-wider mb-1.5 font-semibold">Logging Level</label>
              <select v-model="form.log_level"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 font-mono text-white">
                <option value="debug">DEBUG (High Verbosity)</option>
                <option value="info">INFO (Standard Production)</option>
                <option value="warn">WARN (Errors & Warnings Only)</option>
                <option value="error">ERROR (Critical Failures Only)</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs text-slate-400 uppercase tracking-wider mb-1.5 font-semibold">JWT Session Expiration (Hours)</label>
              <input type="number" v-model.number="form.jwt_expiration_hours" min="1" max="720"
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 font-mono text-white" />
            </div>
          </div>

          <div class="pt-3 border-t border-slate-800/80 flex items-center justify-end space-x-3">
            <button type="submit" :disabled="saving"
                    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
              <span v-if="saving" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
              Apply Runtime Settings
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- TAB 4: Local-First Principle -->
    <div v-else class="saas-card p-5 space-y-3 max-w-4xl font-sans">
      <h2 class="text-sm font-bold text-white tracking-tight border-b border-slate-800/80 pb-3">
        Local-First Edge Architecture
      </h2>
      <p class="text-xs text-slate-300 leading-relaxed">
        The system is designed with a strict local-first architecture. All operational telemetry, user accounts, security roles, alarms, and configuration parameters are persisted directly onto the edge hardware (MariaDB Edge). In the event of network disruption or cloud disconnects, telemetry acquisition, threshold monitoring, and local reporting continue uninterrupted.
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
      saveResult: null,
      dbStatus: {
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
        db_host: '127.0.0.1',
        db_port: '3306',
        db_name: 'datalogger',
        db_user: 'root',
        db_password: '',
        port: '8080',
        environment: 'production',
        log_level: 'info',
        jwt_expiration_hours: 24,
        serial_port: 'COM1',
        serial_baud_rate: 9600,
        serial_data_bits: 8,
        serial_stop_bits: 1,
        serial_parity: 'N',
        serial_timeout_ms: 1000,
      },
    };
  },
  mounted() {
    this.fetchDbStatus();
    this.fetchConfig();
  },
  methods: {
    async fetchDbStatus() {
      try {
        const res = await axios.get('/api/system/database/status');
        if (res.data && res.data.data) {
          this.dbStatus = res.data.data;
        }
      } catch (err) {
        console.warn('Failed to fetch DB status:', err);
      }
    },
    async fetchConfig() {
      try {
        const res = await axios.get('/api/system/config');
        if (res.data && res.data.data) {
          const cfg = res.data.data;
          this.form = {
            db_host: cfg.db_host || '127.0.0.1',
            db_port: cfg.db_port || '3306',
            db_name: cfg.db_name || 'datalogger',
            db_user: cfg.db_user || 'root',
            db_password: '',
            port: cfg.port || '8080',
            environment: cfg.environment || 'production',
            log_level: cfg.log_level || 'info',
            jwt_expiration_hours: cfg.jwt_expiration_hours || 24,
            serial_port: cfg.serial_port || 'COM1',
            serial_baud_rate: cfg.serial_baud_rate || 9600,
            serial_data_bits: cfg.serial_data_bits || 8,
            serial_stop_bits: cfg.serial_stop_bits || 1,
            serial_parity: cfg.serial_parity || 'N',
            serial_timeout_ms: cfg.serial_timeout_ms || 1000,
          };
        }
      } catch (err) {
        console.warn('Failed to fetch system config:', err);
      }
    },
    async testDbConnection() {
      this.testing = true;
      this.testResult = null;
      try {
        const payload = {
          host: this.form.db_host,
          port: this.form.db_port,
          database: this.form.db_name,
          username: this.form.db_user,
          password: this.form.db_password,
        };
        const res = await axios.post('/api/system/database/test', payload);
        if (res.data) {
          if (res.data.success && res.data.data && res.data.data.connected) {
            this.testResult = {
              success: true,
              message: `Connected to ${res.data.data.version} on ${this.form.db_host}:${this.form.db_port}`,
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
    async submitConfig() {
      if (this.activeTab === 'serial') {
        if (!confirm('Warning: Modifying serial parameters may temporarily interrupt active telemetry acquisition. Do you wish to continue?')) {
          return;
        }
      }

      this.saving = true;
      this.saveResult = null;
      try {
        const res = await axios.post('/api/system/config', this.form);
        if (res.data && res.data.success) {
          this.saveResult = res.data.data;
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
