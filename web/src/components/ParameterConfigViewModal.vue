<template>
  <div v-if="isOpen && targetParam" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-3xl shadow-2xl overflow-hidden flex flex-col max-h-[92vh]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/90">
        <div class="flex items-center space-x-3">
          <div class="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 flex-shrink-0">
            <ParameterIcon
              :icon="targetParam.icon"
              :code="targetParam.parameter_code || targetParam.param_code"
              :name="targetParam.parameter_name || targetParam.param_name"
              :unit="targetParam.unit"
              size-class="w-5 h-5"
            />
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="text-sm font-bold text-white tracking-wide">
                {{ targetParam.parameter_name || targetParam.param_name || targetParam.parameter_code }}
              </h2>
              <span class="text-3xs font-mono px-2 py-0.5 rounded bg-slate-800 text-cyan-300 border border-slate-700">
                {{ targetParam.parameter_code || targetParam.param_code }}
              </span>
              <span
                class="text-4xs font-mono font-bold px-1.5 py-0.5 rounded uppercase"
                :class="isParamActive ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30' : 'bg-slate-800 text-slate-400 border border-slate-700'"
              >
                {{ isParamActive ? 'ACTIVE' : 'DISABLED' }}
              </span>
            </div>
            <p class="text-3xs text-slate-400 font-mono mt-0.5 flex items-center space-x-1.5">
              <span>Device: <strong class="text-slate-200">{{ targetDevice.device_code || targetDevice.code || '--' }}</strong> ({{ targetDevice.name || targetDevice.device_name || 'Station Node' }})</span>
              <span>·</span>
              <span>Register: <strong class="text-slate-200">#{{ registerAddress }}</strong></span>
            </p>
          </div>
        </div>
        <button @click="closeModal" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Body -->
      <div class="p-6 space-y-4 text-xs font-sans overflow-y-auto max-h-[calc(92vh-130px)] custom-scrollbar">
        <!-- Section 1: Identification & General Properties -->
        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-3">
          <div class="flex items-center space-x-2">
            <span class="w-2 h-2 rounded-full bg-cyan-400"></span>
            <h3 class="text-xs font-bold text-white uppercase tracking-wider font-sans">
              General & Identification
            </h3>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-3 text-3xs font-mono">
            <div>
              <span class="text-slate-500 uppercase block font-sans">Parameter Code</span>
              <span class="text-slate-200 font-bold mt-0.5 block truncate">{{ targetParam.parameter_code || targetParam.param_code || '--' }}</span>
            </div>
            <div>
              <span class="text-slate-500 uppercase block font-sans">Display Name</span>
              <span class="text-slate-200 font-bold mt-0.5 block truncate">{{ targetParam.parameter_name || targetParam.param_name || '--' }}</span>
            </div>
            <div>
              <span class="text-slate-500 uppercase block font-sans">Category</span>
              <span class="text-blue-300 font-semibold mt-0.5 block truncate">{{ categoryName }}</span>
            </div>
            <div>
              <span class="text-slate-500 uppercase block font-sans">Measurement Unit</span>
              <span class="text-slate-200 font-bold mt-0.5 block">{{ targetParam.unit || '(None)' }}</span>
            </div>
            <div>
              <span class="text-slate-500 uppercase block font-sans">Precision</span>
              <span class="text-slate-200 mt-0.5 block">{{ targetParam.precision !== undefined ? targetParam.precision : 2 }} Decimals</span>
            </div>
            <div>
              <span class="text-slate-500 uppercase block font-sans">Icon Identifier</span>
              <span class="text-slate-300 mt-0.5 block">{{ targetParam.icon || 'Auto-detected' }}</span>
            </div>
            <div>
              <span class="text-slate-500 uppercase block font-sans">Poll State</span>
              <span class="mt-0.5 block font-bold" :class="isParamActive ? 'text-emerald-400' : 'text-slate-400'">
                {{ isParamActive ? 'Acquisition Enabled' : 'Acquisition Paused' }}
              </span>
            </div>
            <div>
              <span class="text-slate-500 uppercase block font-sans">Device Interface</span>
              <span class="text-slate-300 mt-0.5 block truncate" :title="configuredPort">
                {{ configuredPort }}
              </span>
            </div>
          </div>
        </div>

        <!-- Section 2: Modbus Registers & Conversion -->
        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-3">
          <div class="flex items-center space-x-2">
            <span class="w-2 h-2 rounded-full bg-blue-400"></span>
            <h3 class="text-xs font-bold text-white uppercase tracking-wider font-sans">
              Modbus Protocol & Data Mapping
            </h3>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-3 gap-3 text-3xs font-mono">
            <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
              <span class="text-slate-400 font-sans uppercase block">Register Address</span>
              <span class="text-sm font-bold text-emerald-400 mt-1 block">#{{ registerAddress }}</span>
              <span class="text-slate-500 text-4xs">PDU Address: {{ pduAddress }}</span>
            </div>
            <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
              <span class="text-slate-400 font-sans uppercase block">Register Type</span>
              <span class="text-xs font-bold text-cyan-300 mt-1 block">{{ registerType }}</span>
              <span class="text-slate-500 text-4xs">{{ registerTypeDescription }}</span>
            </div>
            <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
              <span class="text-slate-400 font-sans uppercase block">Data Type & Byte Order</span>
              <span class="text-xs font-bold text-blue-300 mt-1 block">{{ dataType }} [{{ byteOrder }}]</span>
              <span class="text-slate-500 text-4xs">{{ dataTypeSize }}</span>
            </div>
            <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
              <span class="text-slate-400 font-sans uppercase block">Scale Multiplier</span>
              <span class="text-xs font-bold text-slate-200 mt-1 block">×{{ scaleFactor }}</span>
              <span class="text-slate-500 text-4xs">Multiplied before offset</span>
            </div>
            <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
              <span class="text-slate-400 font-sans uppercase block">Offset Adder</span>
              <span class="text-xs font-bold text-slate-200 mt-1 block">+{{ offsetAdder }}</span>
              <span class="text-slate-500 text-4xs">Added after scale</span>
            </div>
            <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
              <span class="text-slate-400 font-sans uppercase block">Linear Formula</span>
              <span class="text-xs font-bold text-purple-300 mt-1 block truncate">y = (x × {{ scaleFactor }}) + {{ offsetAdder }}</span>
              <span class="text-slate-500 text-4xs">Engineering value</span>
            </div>
          </div>
        </div>

        <!-- Section 3: Alarm Thresholds & Warning Limits -->
        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-3">
          <div class="flex items-center space-x-2">
            <span class="w-2 h-2 rounded-full bg-amber-400"></span>
            <h3 class="text-xs font-bold text-white uppercase tracking-wider font-sans">
              Thresholds & Warning Limits
            </h3>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-2.5 text-3xs font-mono">
            <div class="p-2 rounded-lg bg-rose-950/20 border border-rose-500/30">
              <span class="text-rose-400 font-sans uppercase block font-semibold">Critical Low</span>
              <span class="text-xs font-bold text-rose-300 mt-1 block">{{ lowCritical !== null ? lowCritical : '--' }} {{ targetParam.unit }}</span>
            </div>
            <div class="p-2 rounded-lg bg-amber-950/20 border border-amber-500/30">
              <span class="text-amber-400 font-sans uppercase block font-semibold">Warning Low</span>
              <span class="text-xs font-bold text-amber-300 mt-1 block">{{ lowWarning !== null ? lowWarning : '--' }} {{ targetParam.unit }}</span>
            </div>
            <div class="p-2 rounded-lg bg-amber-950/20 border border-amber-500/30">
              <span class="text-amber-400 font-sans uppercase block font-semibold">Warning High</span>
              <span class="text-xs font-bold text-amber-300 mt-1 block">{{ highWarning !== null ? highWarning : '--' }} {{ targetParam.unit }}</span>
            </div>
            <div class="p-2 rounded-lg bg-rose-950/20 border border-rose-500/30">
              <span class="text-rose-400 font-sans uppercase block font-semibold">Critical High</span>
              <span class="text-xs font-bold text-rose-300 mt-1 block">{{ highCritical !== null ? highCritical : '--' }} {{ targetParam.unit }}</span>
            </div>
          </div>

          <!-- Limit Visual Strip -->
          <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800 text-4xs font-mono space-y-1.5">
            <div class="flex items-center justify-between text-slate-400">
              <span>[CRIT LOW: {{ lowCritical !== null ? lowCritical : 'None' }}]</span>
              <span>[WARN LOW: {{ lowWarning !== null ? lowWarning : 'None' }}]</span>
              <span class="text-emerald-400 font-bold">NORMAL OPERATIONAL RANGE</span>
              <span>[WARN HIGH: {{ highWarning !== null ? highWarning : 'None' }}]</span>
              <span>[CRIT HIGH: {{ highCritical !== null ? highCritical : 'None' }}]</span>
            </div>
            <div class="h-2 w-full rounded-full bg-slate-800 flex overflow-hidden">
              <div class="w-1/6 bg-rose-500/40" title="Critical Low Zone"></div>
              <div class="w-1/6 bg-amber-500/40" title="Warning Low Zone"></div>
              <div class="w-2/6 bg-emerald-500/40" title="Normal Healthy Zone"></div>
              <div class="w-1/6 bg-amber-500/40" title="Warning High Zone"></div>
              <div class="w-1/6 bg-rose-500/40" title="Critical High Zone"></div>
            </div>
          </div>
        </div>

        <!-- Section 4: Signal Policies & Advanced Math -->
        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-3">
          <div class="flex items-center space-x-2">
            <span class="w-2 h-2 rounded-full bg-emerald-400"></span>
            <h3 class="text-xs font-bold text-white uppercase tracking-wider font-sans">
              Signal Policies & Advanced Math
            </h3>
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 text-3xs font-mono">
            <!-- Zero-Value Alarm Policy -->
            <div class="p-3 bg-black/40 rounded-lg border border-slate-800 space-y-1.5">
              <div class="flex items-center justify-between">
                <span class="text-slate-400 font-sans uppercase font-semibold">Zero-Value Alarm Policy</span>
                <span
                  class="px-2 py-0.5 rounded font-bold uppercase text-4xs"
                  :class="zeroPolicyBadgeClass"
                >
                  {{ zeroPolicy }}
                </span>
              </div>
              <p class="text-slate-300 font-sans text-2xs leading-relaxed">
                {{ zeroPolicyDescription }}
              </p>
            </div>

            <!-- Transformation Formula -->
            <div class="p-3 bg-black/40 rounded-lg border border-slate-800 space-y-1.5">
              <div class="flex items-center justify-between">
                <span class="text-slate-400 font-sans uppercase font-semibold">Formula Transform (ƒ(x))</span>
                <span class="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20 font-bold uppercase text-4xs">
                  {{ formula ? 'ACTIVE' : 'LINEAR ONLY' }}
                </span>
              </div>
              <div class="font-mono text-xs font-bold text-blue-300 truncate" :title="formula || 'None'">
                {{ formula ? ('ƒ(x) = ' + formula) : 'Linear conversion only: y = x × scale + offset' }}
              </div>
            </div>

            <!-- Hold Last Good Value -->
            <div class="p-3 bg-black/40 rounded-lg border border-slate-800 space-y-1.5">
              <div class="flex items-center justify-between">
                <span class="text-slate-400 font-sans uppercase font-semibold">Hold Last Good Value</span>
                <span class="px-2 py-0.5 rounded font-bold uppercase text-4xs" :class="holdLastValueDuration > 0 ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20' : 'bg-slate-800 text-slate-400'">
                  {{ holdLastValueDuration > 0 ? (holdLastValueDuration + 's') : 'DISABLED' }}
                </span>
              </div>
              <p class="text-slate-300 font-sans text-2xs leading-relaxed">
                {{ holdLastValueDuration > 0
                  ? `Suppresses intermittent invalid spikes for up to ${holdLastValueDuration} seconds before reporting BAD quality.`
                  : 'Disabled. Instant invalid quality upon reading errors.' }}
              </p>
            </div>

            <!-- Stale Timeout & Spike Anomaly -->
            <div class="p-3 bg-black/40 rounded-lg border border-slate-800 space-y-1.5">
              <div class="flex items-center justify-between">
                <span class="text-slate-400 font-sans uppercase font-semibold">Stale Timeout & Anomaly</span>
                <span class="px-2 py-0.5 rounded bg-purple-500/10 text-purple-400 border border-purple-500/20 font-bold uppercase text-4xs">
                  {{ staleTimeoutSec }}s ({{ Math.round(staleTimeoutSec / 60 * 10) / 10 }}m)
                </span>
              </div>
              <p class="text-slate-300 font-sans text-2xs leading-relaxed">
                Marks parameter STALE after {{ staleTimeoutSec }}s without updates.
                Spike detection: <strong :class="isSpikeDetectionActive ? 'text-emerald-400' : 'text-slate-500'">{{ isSpikeDetectionActive ? 'Enabled' : 'Disabled' }}</strong>.
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-3 border-t border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
        <div>
          <span class="text-3xs text-slate-500 font-mono">
            Param ID #{{ targetParam.id || targetParam.param_id }} · Device ID #{{ targetDevice.id || targetDevice.device_id }}
          </span>
        </div>
        <div class="flex items-center space-x-2">
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-1.5 rounded-xl text-xs font-medium text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            {{ $t('common.close') || 'Close' }}
          </button>
          <button
            v-if="canEdit"
            type="button"
            @click="$emit('edit')"
            class="px-4 py-1.5 rounded-xl text-xs font-bold text-white bg-blue-600 hover:bg-blue-500 transition flex items-center space-x-1.5 shadow-lg shadow-blue-900/30"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
            </svg>
            <span>Edit Configuration</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import ParameterIcon from './ParameterIcon.vue';

export default {
  name: 'ParameterConfigViewModal',
  components: {
    ParameterIcon,
  },
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
    canEdit: {
      type: Boolean,
      default: false,
    },
  },
  computed: {
    targetDevice() {
      return this.device || {};
    },
    targetParam() {
      return this.parameter || null;
    },
    rawParam() {
      if (!this.targetParam) return {};
      return this.targetParam._rawParam || this.targetParam;
    },
    isParamActive() {
      if (!this.targetParam) return false;
      return this.rawParam.enabled !== false;
    },
    categoryName() {
      if (this.targetParam && this.targetParam.category && this.targetParam.category.name) {
        return this.targetParam.category.name;
      }
      return 'Uncategorized';
    },
    configuredPort() {
      if (this.targetDevice.connection) {
        return this.targetDevice.connection.serial_port ||
          (this.targetDevice.connection.host ? (this.targetDevice.connection.host + ':' + this.targetDevice.connection.port) : '--');
      }
      return '--';
    },
    registerAddress() {
      return this.rawParam.register_address !== undefined ? this.rawParam.register_address : (this.targetParam ? this.targetParam.register_address : '--');
    },
    pduAddress() {
      const reg = Number(this.registerAddress);
      if (isNaN(reg)) return '--';
      return Math.max(0, reg - 1);
    },
    registerType() {
      return this.rawParam.register_type || (this.targetParam && this.targetParam.register_type) || 'HOLDING_REGISTER';
    },
    registerTypeDescription() {
      const t = this.registerType;
      if (t === 'HOLDING_REGISTER') return 'Read/Write 16-bit word (FC 03)';
      if (t === 'INPUT_REGISTER') return 'Read-only 16-bit analog (FC 04)';
      if (t === 'COIL') return 'Discrete digital output (FC 01)';
      if (t === 'DISCRETE_INPUT') return 'Discrete digital input (FC 02)';
      return 'Standard register';
    },
    dataType() {
      return this.rawParam.data_type || (this.targetParam && this.targetParam.data_type) || 'FLOAT32';
    },
    dataTypeSize() {
      const dt = this.dataType.toUpperCase();
      if (dt.includes('64')) return '64-bit (4 Modbus words)';
      if (dt.includes('32')) return '32-bit (2 Modbus words)';
      return '16-bit (1 Modbus word)';
    },
    byteOrder() {
      return this.rawParam.byte_order ||
        (this.targetParam && this.targetParam.byte_order) ||
        (this.targetDevice.connection && this.targetDevice.connection.byte_order) ||
        'ABCD';
    },
    scaleFactor() {
      if (this.rawParam.scale !== undefined) return this.rawParam.scale;
      if (this.targetParam && this.targetParam.scale !== undefined) return this.targetParam.scale;
      return 1;
    },
    offsetAdder() {
      if (this.rawParam.offset !== undefined) return this.rawParam.offset;
      if (this.targetParam && this.targetParam.offset !== undefined) return this.targetParam.offset;
      return 0;
    },
    lowCritical() {
      if (this.rawParam.low_critical !== undefined && this.rawParam.low_critical !== null) return this.rawParam.low_critical;
      if (this.targetParam && this.targetParam.low_critical !== undefined && this.targetParam.low_critical !== null) return this.targetParam.low_critical;
      if (this.targetParam && this.targetParam.low_limit !== undefined && this.targetParam.low_limit !== null) return this.targetParam.low_limit;
      return null;
    },
    highCritical() {
      if (this.rawParam.high_critical !== undefined && this.rawParam.high_critical !== null) return this.rawParam.high_critical;
      if (this.targetParam && this.targetParam.high_critical !== undefined && this.targetParam.high_critical !== null) return this.targetParam.high_critical;
      if (this.targetParam && this.targetParam.high_limit !== undefined && this.targetParam.high_limit !== null) return this.targetParam.high_limit;
      return null;
    },
    lowWarning() {
      if (this.rawParam.low_warning !== undefined && this.rawParam.low_warning !== null) return this.rawParam.low_warning;
      if (this.targetParam && this.targetParam.low_warning !== undefined && this.targetParam.low_warning !== null) return this.targetParam.low_warning;
      return null;
    },
    highWarning() {
      if (this.rawParam.high_warning !== undefined && this.rawParam.high_warning !== null) return this.rawParam.high_warning;
      if (this.targetParam && this.targetParam.high_warning !== undefined && this.targetParam.high_warning !== null) return this.targetParam.high_warning;
      return null;
    },
    zeroPolicy() {
      return this.rawParam.zero_value_alarm_policy || (this.targetParam && this.targetParam.zero_value_alarm_policy) || 'NORMAL';
    },
    zeroPolicyBadgeClass() {
      const p = this.zeroPolicy;
      if (p === 'SUPPRESS_THRESHOLD') return 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30';
      if (p === 'INVALID_READING') return 'bg-amber-500/15 text-amber-400 border border-amber-500/30';
      return 'bg-slate-800 text-slate-300 border border-slate-700';
    },
    zeroPolicyDescription() {
      const p = this.zeroPolicy;
      if (p === 'SUPPRESS_THRESHOLD') {
        return 'Zero is a valid physical measurement (e.g. rain/solar/wind/gas). When sensor reads 0.0, low alarm thresholds are suppressed and quality stays GOOD.';
      }
      if (p === 'INVALID_READING') {
        return 'Zero represents a disconnected wire or power loss. When sensor reads 0.0, reading is marked UNCERTAIN/BAD.';
      }
      return 'Standard threshold evaluation. 0.0 is tested normally against configured warning and critical limits.';
    },
    formula() {
      return this.rawParam.formula || (this.targetParam && this.targetParam.formula) || '';
    },
    holdLastValueDuration() {
      return this.rawParam.hold_last_value_seconds || (this.targetParam && this.targetParam.hold_last_value_seconds) || 0;
    },
    staleTimeoutSec() {
      return this.rawParam.stale_timeout_seconds || (this.targetParam && this.targetParam.stale_timeout_seconds) || 120;
    },
    isSpikeDetectionActive() {
      return !!(this.rawParam.spike_detection_enabled || this.rawParam.anomaly_detection_enabled || (this.targetParam && this.targetParam.spike_detection_enabled));
    },
  },
  methods: {
    closeModal() {
      this.$emit('close');
    },
  },
};
</script>
