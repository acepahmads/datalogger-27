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
        precision: 2,
        scale: 1.0,
        offset: 0.0,
        enabled: true,
      },
    };
  },
  computed: {
    isEditing() {
      return !!this.paramToEdit;
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
            precision: newVal.precision !== undefined ? newVal.precision : 2,
            scale: newVal.scale !== undefined ? newVal.scale : (newVal.scale_factor || 1.0),
            offset: newVal.offset !== undefined ? newVal.offset : 0.0,
            enabled: newVal.enabled !== undefined ? newVal.enabled : true,
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
        precision: 2,
        scale: 1.0,
        offset: 0.0,
        enabled: true,
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
