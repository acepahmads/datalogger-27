<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/70 backdrop-blur-sm">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-2xl max-w-lg w-full overflow-hidden flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-900/50">
        <div class="flex items-center space-x-3">
          <div class="w-9 h-9 rounded-lg bg-amber-500/10 dark:bg-amber-400/10 border border-amber-500/20 text-amber-600 dark:text-amber-400 flex items-center justify-center font-mono font-bold text-sm">
            ⚡
          </div>
          <div>
            <h3 class="font-bold text-slate-800 dark:text-slate-100 text-sm">Test Write Modbus</h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              {{ actuator ? actuator.actuator_name : 'Direct Register Probe' }}
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="close"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 transition-colors p-1 rounded-md"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 space-y-5 overflow-y-auto">
        <!-- Target Info Pill -->
        <div class="bg-slate-50 dark:bg-slate-800/60 rounded-lg p-3 border border-slate-200/80 dark:border-slate-700/60 flex flex-wrap items-center justify-between text-xs gap-2">
          <div>
            <span class="text-slate-400">Function:</span>
            <span class="font-mono font-bold ml-1 text-slate-700 dark:text-slate-200">{{ targetControlType }}</span>
          </div>
          <div>
            <span class="text-slate-400">Slave ID:</span>
            <span class="font-mono font-bold ml-1 text-slate-700 dark:text-slate-200">#{{ targetSlaveID }}</span>
          </div>
          <div>
            <span class="text-slate-400">Address:</span>
            <span class="font-mono font-bold ml-1 text-slate-700 dark:text-slate-200">{{ targetAddress }}</span>
          </div>
        </div>

        <!-- Action Selection -->
        <div>
          <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300 mb-2">
            Select Test Command
          </label>
          <div class="grid grid-cols-2 gap-3">
            <button
              type="button"
              @click="executeTest('ON')"
              :disabled="loading"
              class="flex items-center justify-center space-x-2 py-2.5 px-4 rounded-lg bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white font-semibold text-xs shadow-sm transition-all disabled:opacity-50"
            >
              <span v-if="loading && currentAction === 'ON'" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
              <span>Test Write ON</span>
              <span class="text-emerald-200 text-3xs font-mono">({{ formatHex(targetOnPayload) }})</span>
            </button>

            <button
              type="button"
              @click="executeTest('OFF')"
              :disabled="loading"
              class="flex items-center justify-center space-x-2 py-2.5 px-4 rounded-lg bg-rose-600 hover:bg-rose-500 active:bg-rose-700 text-white font-semibold text-xs shadow-sm transition-all disabled:opacity-50"
            >
              <span v-if="loading && currentAction === 'OFF'" class="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
              <span>Test Write OFF</span>
              <span class="text-rose-200 text-3xs font-mono">({{ formatHex(targetOffPayload) }})</span>
            </button>
          </div>
        </div>

        <!-- Result Console -->
        <div v-if="result" class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">Response Verification</span>
            <span
              :class="result.success ? 'bg-emerald-50 text-emerald-700 border-emerald-200 dark:bg-emerald-500/10 dark:text-emerald-400 dark:border-emerald-500/30' : 'bg-rose-50 text-rose-700 border-rose-200 dark:bg-rose-500/10 dark:text-rose-400 dark:border-rose-500/30'"
              class="px-2.5 py-0.5 rounded-full text-xs font-bold border flex items-center space-x-1"
            >
              <span class="w-1.5 h-1.5 rounded-full" :class="result.success ? 'bg-emerald-500' : 'bg-rose-500'"></span>
              <span>{{ result.success ? 'ECHO VERIFIED' : 'WRITE FAILED' }}</span>
            </span>
          </div>

          <div class="bg-slate-900 dark:bg-slate-950 text-slate-200 rounded-lg p-3.5 font-mono text-xs space-y-2 border border-slate-800">
            <div class="flex justify-between border-b border-slate-800 pb-1.5 text-slate-400">
              <span>Roundtrip Latency</span>
              <span class="text-amber-400 font-bold">{{ result.latency_ms }} ms</span>
            </div>

            <div class="space-y-1">
              <div class="text-slate-400 text-3xs uppercase">TX Request Frame (Hex)</div>
              <div class="text-sky-400 break-all select-all">{{ formatFrameHex(result.raw_request_hex) || '--' }}</div>
            </div>

            <div class="space-y-1">
              <div class="text-slate-400 text-3xs uppercase">RX Response Echo (Hex)</div>
              <div :class="result.success ? 'text-emerald-400' : 'text-rose-400'" class="break-all select-all">
                {{ formatFrameHex(result.raw_response_hex) || (result.error_message ? 'NO RESPONSE / TIMEOUT' : '--') }}
              </div>
            </div>

            <div v-if="result.error_message" class="text-rose-400 text-3xs pt-1 border-t border-slate-800/80">
              Error: {{ result.error_message }}
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-3 border-t border-slate-100 dark:border-slate-800 flex justify-end bg-slate-50/50 dark:bg-slate-900/50">
        <button
          type="button"
          @click="close"
          class="px-4 py-2 text-xs font-semibold rounded-lg text-slate-700 dark:text-slate-200 hover:bg-slate-200/60 dark:hover:bg-slate-800 transition-colors"
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
  name: 'ActuatorTestWriteModal',
  props: {
    isOpen: { type: Boolean, default: false },
    actuator: { type: Object, default: null },
    deviceId: { type: [Number, String], default: null },
    rawConfig: { type: Object, default: null }
  },
  data() {
    return {
      loading: false,
      currentAction: '',
      result: null
    };
  },
  computed: {
    targetControlType() {
      if (this.actuator) return this.actuator.control_type;
      if (this.rawConfig) return this.rawConfig.control_type;
      return 'COIL_FC05';
    },
    targetSlaveID() {
      if (this.actuator) return this.actuator.slave_id;
      if (this.rawConfig) return this.rawConfig.slave_id;
      return 1;
    },
    targetAddress() {
      if (this.actuator) return this.actuator.register_address;
      if (this.rawConfig) return this.rawConfig.register_address;
      return 0;
    },
    targetOnPayload() {
      if (this.actuator) return this.actuator.on_payload;
      if (this.rawConfig) return this.rawConfig.on_payload;
      return this.targetControlType === 'COIL_FC05' ? 0xFF00 : 1;
    },
    targetOffPayload() {
      if (this.actuator) return this.actuator.off_payload;
      if (this.rawConfig) return this.rawConfig.off_payload;
      return 0;
    }
  },
  methods: {
    close() {
      this.result = null;
      this.$emit('close');
    },
    formatHex(val) {
      if (val === undefined || val === null) return '0x0000';
      return '0x' + Number(val).toString(16).toUpperCase().padStart(4, '0');
    },
    formatFrameHex(hexStr) {
      if (!hexStr) return '';
      const chunks = hexStr.match(/.{1,2}/g);
      return chunks ? chunks.join(' ').toUpperCase() : hexStr;
    },
    async executeTest(action) {
      this.loading = true;
      this.currentAction = action;
      this.result = null;

      try {
        let res;
        if (this.actuator && this.actuator.id) {
          res = await axios.post(`/api/actuators/${this.actuator.id}/test-write`, {
            action: action
          });
        } else {
          res = await axios.post('/api/actuators/test-write', {
            device_id: Number(this.deviceId),
            control_type: this.targetControlType,
            slave_id: Number(this.targetSlaveID),
            address: Number(this.targetAddress),
            value: action === 'ON' ? Number(this.targetOnPayload) : Number(this.targetOffPayload),
            action: action
          });
        }
        this.result = res.data.data;
        this.$emit('tested', this.result);
      } catch (err) {
        const errData = err.response && err.response.data;
        this.result = {
          success: false,
          latency_ms: 0,
          error_message: (errData && (errData.error || errData.message)) || err.message
        };
      } finally {
        this.loading = false;
      }
    }
  }
};
</script>
