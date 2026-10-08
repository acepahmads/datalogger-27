<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
        <div>
          <h2 class="text-base font-bold text-white tracking-wide">
            {{ isEditing ? $t('parameterModal.titleEdit') : $t('parameterModal.titleAdd') }}
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            {{ isEditing ? $t('parameterModal.subtitleEdit') : $t('parameterModal.subtitleAdd') }}
          </p>
        </div>
        <button @click="closeModal" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Modal Body -->
      <form @submit.prevent="submitForm" class="flex-1 overflow-y-auto p-6 space-y-4">
        <!-- Error Alert -->
        <div v-if="error" class="p-3 bg-rose-500/10 border border-rose-500/30 rounded-xl text-xs text-rose-400 flex items-center space-x-2">
          <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>{{ error }}</span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Parameter Code -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.code') }}
            </label>
            <input
              v-model="form.parameter_code"
              type="text"
              required
              :placeholder="$t('parameterModal.placeholders.code')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>

          <!-- Parameter Name -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.name') }}
            </label>
            <input
              v-model="form.parameter_name"
              type="text"
              required
              :placeholder="$t('parameterModal.placeholders.name')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <!-- Data Type -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.dataType') }}
            </label>
            <select
              v-model="form.data_type"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="FLOAT32">FLOAT32 (Single Precision)</option>
              <option value="FLOAT64">FLOAT64 (Double Precision)</option>
              <option value="INT16">INT16 (Signed 16-bit)</option>
              <option value="INT32">INT32 (Signed 32-bit)</option>
              <option value="UINT16">UINT16 (Unsigned 16-bit)</option>
              <option value="UINT32">UINT32 (Unsigned 32-bit)</option>
              <option value="BOOLEAN">BOOLEAN (Digital State)</option>
              <option value="STRING">STRING (ASCII)</option>
            </select>
          </div>

          <!-- Unit -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.unit') }}
            </label>
            <input
              v-model="form.unit"
              type="text"
              :placeholder="$t('parameterModal.placeholders.unit')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <!-- Decimal Precision -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.precision') }}
            </label>
            <input
              v-model.number="form.precision"
              type="number"
              min="0"
              max="6"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <!-- Register Address -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.addressOffset') }}
            </label>
            <input
              v-model.number="form.register_address"
              type="number"
              placeholder="e.g. 30001, 40001"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>

          <!-- Register Type -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.registerType') }}
            </label>
            <select
              v-model="form.register_type"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="HOLDING_REGISTER">Holding Register (4x)</option>
              <option value="INPUT_REGISTER">Input Register (3x)</option>
              <option value="COIL">Coil (0x)</option>
              <option value="DISCRETE_INPUT">Discrete Input (1x)</option>
              <option value="VARIABLE">Virtual / Internal Variable</option>
            </select>
          </div>

          <!-- Byte Order Override -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.byteOrderOverride') }}
            </label>
            <select
              v-model="form.byte_order"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="">{{ $t('parameterModal.defaultDeviceConn') }}</option>
              <option value="ABCD">ABCD (Big Endian)</option>
              <option value="CDAB">CDAB (Word Swap)</option>
              <option value="BADC">BADC (Byte Swap)</option>
              <option value="DCBA">DCBA (Little Endian)</option>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Scale Factor -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.scaleFactor') }}
            </label>
            <input
              v-model.number="form.scale"
              type="number"
              step="0.001"
              placeholder="1.0"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>

          <!-- Offset -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.offsetAdder') }}
            </label>
            <input
              v-model.number="form.offset"
              type="number"
              step="0.01"
              placeholder="0.0"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Min Value -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.minLimit') }}
            </label>
            <input
              v-model.number="form.min_value"
              type="number"
              step="any"
              placeholder="e.g. 0.0"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>

          <!-- Max Value -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.maxLimit') }}
            </label>
            <input
              v-model.number="form.max_value"
              type="number"
              step="any"
              placeholder="e.g. 500.0"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>
        </div>

        <!-- Warning Limits (Soft Limits: Phase 3.2 UNCERTAIN Quality) -->
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label class="block text-2xs font-semibold text-amber-400/90 uppercase tracking-wider mb-1.5 flex items-center justify-between">
              <span>{{ $t('parameterModal.warningLow') }}</span>
              <span class="text-3xs text-amber-500/80 font-mono">UNCERTAIN</span>
            </label>
            <input
              v-model.number="form.warning_low"
              type="number"
              step="any"
              :placeholder="$t('parameterModal.placeholders.warningLow')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-amber-500/30 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-amber-400 font-mono"
            />
          </div>

          <div>
            <label class="block text-2xs font-semibold text-amber-400/90 uppercase tracking-wider mb-1.5 flex items-center justify-between">
              <span>{{ $t('parameterModal.warningHigh') }}</span>
              <span class="text-3xs text-amber-500/80 font-mono">UNCERTAIN</span>
            </label>
            <input
              v-model.number="form.warning_high"
              type="number"
              step="any"
              :placeholder="$t('parameterModal.placeholders.warningHigh')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-amber-500/30 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-amber-400 font-mono"
            />
          </div>
        </div>

        <!-- Mathematical Formula Evaluation (Python-style eval) -->
        <div class="p-3.5 bg-slate-900/60 rounded-xl border border-slate-800 space-y-2.5">
          <div class="flex items-center justify-between">
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.formula') }}
            </label>
            <span class="text-3xs text-blue-400 font-mono">eval(x, raw)</span>
          </div>
          <input
            v-model="form.formula"
            type="text"
            :placeholder="$t('parameterModal.formulaPlaceholder')"
            class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
          />
          <p class="text-3xs text-slate-400 leading-relaxed">
            {{ $t('parameterModal.formulaDesc') }}
          </p>

          <!-- Quick Presets -->
          <div class="flex items-center space-x-1.5 flex-wrap gap-y-1.5 pt-1">
            <span class="text-3xs text-slate-400 mr-1">{{ $t('parameterModal.presets') }}:</span>
            <button
              type="button"
              @click="form.formula = 'x * 1.8 + 32'"
              class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              °C➔°F (x*1.8+32)
            </button>
            <button
              type="button"
              @click="form.formula = '(raw - 4.0) * (100.0 / 16.0)'"
              class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              4-20mA➔%
            </button>
            <button
              type="button"
              @click="form.formula = 'round(x, 2)'"
              class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              round(x, 2)
            </button>
            <button
              type="button"
              @click="form.formula = 'x / 1000'"
              class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              W➔kW (x/1000)
            </button>
          </div>

          <!-- Live Preview Banner -->
          <div v-if="form.formula && form.formula.trim()" class="p-2 rounded-lg bg-[#0B0F19] border border-slate-800 flex items-center justify-between text-2xs font-mono">
            <span class="text-slate-400">{{ $t('parameterModal.formulaPreview') }} (x={{ previewSampleX }}):</span>
            <span v-if="formulaPreview && formulaPreview.valid" class="font-bold text-emerald-400">
              ➔ {{ formulaPreview.value }} {{ form.unit }}
            </span>
            <span v-else class="text-rose-400 font-semibold">
              ⚠️ {{ (formulaPreview && formulaPreview.error) || 'Invalid expression' }}
            </span>
          </div>
        </div>

        <!-- Anomaly & Timeout Hold-Last-Good-Value Protection -->
        <div class="p-3.5 bg-slate-900/60 rounded-xl border border-slate-800 space-y-3">
          <label class="flex items-center space-x-3 cursor-pointer">
            <input
              v-model="form.hold_last_value_enabled"
              type="checkbox"
              class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
            />
            <div>
              <span class="text-xs font-semibold text-slate-200">{{ $t('parameterModal.holdLastValueTitle') }}</span>
              <span class="block text-3xs text-slate-400 mt-0.5">{{ $t('parameterModal.holdLastValueDesc') }}</span>
            </div>
          </label>

          <!-- Duration Input (Conditional when enabled) -->
          <div v-if="form.hold_last_value_enabled" class="pl-7 pt-2 border-t border-slate-800/80 space-y-1.5">
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.holdDurationSeconds') }}
            </label>
            <div class="flex items-center space-x-3">
              <input
                v-model.number="form.hold_last_value_seconds"
                type="number"
                min="10"
                max="86400"
                step="10"
                placeholder="120"
                class="w-32 px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
              />
              <span class="text-3xs text-slate-400 font-mono">
                ({{ Math.round((form.hold_last_value_seconds || 120) / 60 * 10) / 10 }} min)
              </span>
            </div>
            <p class="text-3xs text-amber-400/90 leading-relaxed">
              {{ $t('parameterModal.holdDurationHelp') }}
            </p>
          </div>
        </div>

        <!-- Phase 3.2 Data Quality & Processing Settings -->
        <div class="p-3.5 bg-slate-900/60 rounded-xl border border-slate-800 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-2xs font-bold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.qualitySettingsTitle') }}
            </span>
            <span class="text-3xs text-blue-400 font-mono">Phase 3.2</span>
          </div>

          <!-- Quality Validation Toggle -->
          <label class="flex items-center space-x-3 cursor-pointer">
            <input
              v-model="form.quality_validation_enabled"
              type="checkbox"
              class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
            />
            <div>
              <span class="text-xs font-semibold text-slate-200">{{ $t('parameterModal.qualityValidationEnabled') }}</span>
              <span class="block text-3xs text-slate-400">{{ $t('parameterModal.qualityValidationDesc') }}</span>
            </div>
          </label>

          <!-- Stale Timeout Input -->
          <div class="pt-2 border-t border-slate-800/80 space-y-1.5">
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.staleTimeoutSeconds') }}
            </label>
            <div class="flex items-center space-x-3">
              <input
                v-model.number="form.stale_timeout_seconds"
                type="number"
                min="10"
                max="86400"
                step="10"
                placeholder="120"
                class="w-32 px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
              />
              <span class="text-3xs text-slate-400 font-mono">
                ({{ Math.round((form.stale_timeout_seconds || 120) / 60 * 10) / 10 }} min)
              </span>
            </div>
            <p class="text-3xs text-slate-400 leading-relaxed">
              {{ $t('parameterModal.staleTimeoutHelp') }}
            </p>
          </div>

          <!-- Spike Detection Toggle & Config -->
          <div class="pt-2 border-t border-slate-800/80 space-y-2">
            <label class="flex items-center space-x-3 cursor-pointer">
              <input
                v-model="form.spike_detection_enabled"
                type="checkbox"
                class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
              />
              <div>
                <span class="text-xs font-semibold text-slate-200">{{ $t('parameterModal.spikeDetectionEnabled') }}</span>
                <span class="block text-3xs text-slate-400">{{ $t('parameterModal.spikeDetectionDesc') }}</span>
              </div>
            </label>

            <div v-if="form.spike_detection_enabled" class="pl-7 grid grid-cols-1 md:grid-cols-2 gap-3 pt-1">
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                  {{ $t('parameterModal.spikeThreshold') }}
                </label>
                <input
                  v-model.number="form.spike_threshold"
                  type="number"
                  step="any"
                  placeholder="e.g. 20.0"
                  class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
                />
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                  {{ $t('parameterModal.spikeWindowSize') }}
                </label>
                <input
                  v-model.number="form.spike_window_size"
                  type="number"
                  min="2"
                  max="10"
                  placeholder="3"
                  class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Description -->
        <div>
          <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
            {{ $t('common.description') }}
          </label>
          <input
            v-model="form.description"
            type="text"
            :placeholder="$t('parameterModal.placeholders.description')"
            class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
          />
        </div>

        <!-- Enabled Switch -->
        <div>
          <label class="flex items-center space-x-3 cursor-pointer p-2.5 rounded-xl bg-[#0B0F19] border border-slate-800">
            <input
              v-model="form.enabled"
              type="checkbox"
              class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
            />
            <div>
              <span class="text-xs font-semibold text-slate-200">{{ $t('parameterModal.paramActive') }}</span>
              <span class="block text-3xs text-slate-400">{{ $t('parameterModal.collectChannel') }}</span>
            </div>
          </label>
        </div>

        <!-- Modal Footer -->
        <div class="pt-4 border-t border-slate-800 flex items-center justify-between">
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-2 rounded-xl text-xs font-medium text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            {{ $t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="loading"
            class="px-5 py-2 rounded-xl text-xs font-bold text-white bg-blue-600 hover:bg-blue-500 transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <span v-if="loading" class="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
            <span>{{ isEditing ? $t('parameterModal.updateParam') : $t('parameterModal.addParam') }}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ParameterModal',
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
    deviceId: {
      type: [Number, String],
      required: true,
    },
    paramToEdit: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      loading: false,
      error: null,
      form: {
        parameter_code: '',
        parameter_name: '',
        data_type: 'FLOAT32',
        unit: '',
        description: '',
        register_address: 0,
        register_type: 'HOLDING_REGISTER',
        byte_order: '',
        min_value: null,
        max_value: null,
        warning_low: null,
        warning_high: null,
        quality_validation_enabled: true,
        processing_enabled: true,
        stale_timeout_seconds: 120,
        spike_detection_enabled: false,
        spike_threshold: 0.0,
        spike_window_size: 3,
        precision: 2,
        scale: 1.0,
        offset: 0.0,
        enabled: true,
        formula: '',
        hold_last_value_enabled: false,
        hold_last_value_seconds: 120,
      },
    };
  },
  computed: {
    isEditing() {
      return !!this.paramToEdit;
    },
    previewSampleX() {
      if (this.paramToEdit && this.paramToEdit.current_value !== null && this.paramToEdit.current_value !== undefined) {
        return this.paramToEdit.current_value;
      }
      return this.form.min_value !== null && this.form.min_value !== undefined ? this.form.min_value + 10 : 25.0;
    },
    formulaPreview() {
      if (!this.form.formula || !this.form.formula.trim()) return null;
      try {
        const formStr = this.form.formula.trim();
        const x = this.previewSampleX;
        const raw = 100.0;
        const expr = formStr
          .replace(/\b(pi)\b/gi, 'Math.PI')
          .replace(/\b(e)\b/gi, 'Math.E')
          .replace(/\b(abs)\b/gi, 'Math.abs')
          .replace(/\b(round)\b/gi, 'Math.round')
          .replace(/\b(floor)\b/gi, 'Math.floor')
          .replace(/\b(ceil)\b/gi, 'Math.ceil')
          .replace(/\b(sqrt)\b/gi, 'Math.sqrt')
          .replace(/\b(min)\b/gi, 'Math.min')
          .replace(/\b(max)\b/gi, 'Math.max')
          .replace(/\b(log|ln)\b/gi, 'Math.log')
          .replace(/\b(exp)\b/gi, 'Math.exp')
          .replace(/\b(sin)\b/gi, 'Math.sin')
          .replace(/\b(cos)\b/gi, 'Math.cos')
          .replace(/\b(tan)\b/gi, 'Math.tan')
          .replace(/\^/g, '**')
          .replace(/\b(x|val|value)\b/gi, `(${x})`)
          .replace(/\b(raw)\b/gi, `(${raw})`);

        if (/[^0-9\.\+\-\*\/\(\)\,\sMathPIEabseroundfloceilsqrtminxlgexpstnc]/.test(expr)) {
          return { valid: false, error: 'Invalid syntax or unsupported characters' };
        }
        const fn = new Function(`return (${expr})`);
        const res = fn();
        if (typeof res === 'number' && !isNaN(res) && isFinite(res)) {
          return { valid: true, value: Number(res.toFixed(4)) };
        }
        return { valid: false, error: 'Result is NaN or infinite' };
      } catch (e) {
        return { valid: false, error: e.message };
      }
    },
  },
  watch: {
    paramToEdit: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.form = {
            parameter_code: newVal.parameter_code || newVal.code || '',
            parameter_name: newVal.parameter_name || newVal.name || '',
            data_type: newVal.data_type || 'FLOAT32',
            unit: newVal.unit || '',
            description: newVal.description || '',
            register_address: newVal.register_address || 0,
            register_type: newVal.register_type || 'HOLDING_REGISTER',
            byte_order: newVal.byte_order || '',
            min_value: newVal.min_value !== undefined ? newVal.min_value : (newVal.low_limit || null),
            max_value: newVal.max_value !== undefined ? newVal.max_value : (newVal.high_limit || null),
            warning_low: newVal.warning_low !== undefined ? newVal.warning_low : null,
            warning_high: newVal.warning_high !== undefined ? newVal.warning_high : null,
            quality_validation_enabled: newVal.quality_validation_enabled !== undefined ? newVal.quality_validation_enabled : true,
            processing_enabled: newVal.processing_enabled !== undefined ? newVal.processing_enabled : true,
            stale_timeout_seconds: newVal.stale_timeout_seconds || 120,
            spike_detection_enabled: newVal.spike_detection_enabled !== undefined ? newVal.spike_detection_enabled : false,
            spike_threshold: newVal.spike_threshold || 0.0,
            spike_window_size: newVal.spike_window_size || 3,
            precision: newVal.precision !== undefined ? newVal.precision : 2,
            scale: newVal.scale !== undefined ? newVal.scale : (newVal.scale_factor || 1.0),
            offset: newVal.offset !== undefined ? newVal.offset : 0.0,
            enabled: newVal.enabled !== undefined ? newVal.enabled : true,
            formula: newVal.formula || '',
            hold_last_value_enabled: newVal.hold_last_value_enabled !== undefined ? newVal.hold_last_value_enabled : false,
            hold_last_value_seconds: newVal.hold_last_value_seconds || 120,
          };
        } else {
          this.resetForm();
        }
      },
    },
  },
  methods: {
    closeModal() {
      this.error = null;
      this.$emit('close');
    },
    resetForm() {
      this.form = {
        parameter_code: '',
        parameter_name: '',
        data_type: 'FLOAT32',
        unit: '',
        description: '',
        register_address: 0,
        register_type: 'HOLDING_REGISTER',
        byte_order: '',
        min_value: null,
        max_value: null,
        warning_low: null,
        warning_high: null,
        quality_validation_enabled: true,
        processing_enabled: true,
        stale_timeout_seconds: 120,
        spike_detection_enabled: false,
        spike_threshold: 0.0,
        spike_window_size: 3,
        precision: 2,
        scale: 1.0,
        offset: 0.0,
        enabled: true,
        formula: '',
        hold_last_value_enabled: false,
        hold_last_value_seconds: 120,
      };
    },
    async submitForm() {
      this.loading = true;
      this.error = null;

      try {
        if (this.isEditing) {
          await this.$store.dispatch('updateParameter', {
            deviceId: this.deviceId,
            paramId: this.paramToEdit.id,
            updates: this.form,
          });
        } else {
          await this.$store.dispatch('createParameter', {
            deviceId: this.deviceId,
            paramData: this.form,
          });
        }
        this.$emit('saved');
        this.closeModal();
      } catch (err) {
        this.error = (err.response && err.response.data && err.response.data.error) || err.message || 'Operation failed';
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
