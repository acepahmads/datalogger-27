<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/70 backdrop-blur-sm">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl max-w-3xl w-full overflow-hidden flex flex-col max-h-[92vh]">
      <!-- Header -->
      <div class="px-6 py-5 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-900/50">
        <div class="flex items-center space-x-3.5">
          <div class="w-10 h-10 rounded-xl bg-blue-500/10 dark:bg-blue-400/10 border border-blue-500/20 text-blue-600 dark:text-blue-400 flex items-center justify-center text-lg">
            ⚡
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="text-base font-bold text-slate-900 dark:text-slate-100">
                {{ isEditing ? 'Edit Actuator / Control Equipment' : 'Add Actuator / Control Equipment' }}
              </h2>
              <span v-if="deviceCode" class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
                DEV: {{ deviceCode }}
              </span>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Configure Modbus coil/holding register mapping, auto-off watchdog, and safety limits.
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="close"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1.5 rounded-lg transition-colors"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Form Body -->
      <form @submit.prevent="handleSubmit" class="p-6 overflow-y-auto space-y-5">
        <!-- Error Alert -->
        <div v-if="errorMessage" class="p-3.5 rounded-lg bg-rose-50 dark:bg-rose-500/10 border border-rose-200 dark:border-rose-500/20 text-xs text-rose-600 dark:text-rose-400 font-medium flex items-center space-x-2">
          <span>⚠️</span>
          <span>{{ errorMessage }}</span>
        </div>

        <!-- Section 1: ACTUATOR IDENTITY & PROFILE -->
        <div class="p-4 rounded-xl border border-slate-200/80 dark:border-slate-800 bg-white dark:bg-slate-800/40 shadow-sm space-y-4">
          <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-2.5">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-300 font-mono flex items-center space-x-1.5">
              <span class="text-blue-500">#</span>
              <span>ACTUATOR IDENTITY & PROFILE</span>
            </span>
            <label class="flex items-center space-x-2 text-xs font-semibold cursor-pointer text-slate-700 dark:text-slate-300">
              <input type="checkbox" v-model="form.enabled" class="rounded text-blue-600 focus:ring-blue-500 h-4 w-4 border-slate-300 dark:border-slate-700 dark:bg-slate-800" />
              <span>Active</span>
            </label>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                ACTUATOR CODE <span class="text-rose-500">*</span>
              </label>
              <input
                type="text"
                v-model="form.actuator_code"
                placeholder="e.g. PUMP_01, PURGE_VALVE"
                required
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              />
            </div>
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                ACTUATOR NAME <span class="text-rose-500">*</span>
              </label>
              <input
                type="text"
                v-model="form.actuator_name"
                placeholder="e.g. Water Sampling Pump 01"
                required
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              />
            </div>
          </div>

          <div>
            <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
              DESCRIPTION
            </label>
            <input
              type="text"
              v-model="form.description"
              placeholder="Optional engineering notes, e.g. Submersible pump located in inlet pit..."
              class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
            />
          </div>
        </div>

        <!-- Section 2: CATEGORIES & LAYOUT -->
        <div class="p-4 rounded-xl border border-slate-200/80 dark:border-slate-800 bg-white dark:bg-slate-800/40 shadow-sm space-y-4">
          <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-2.5">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-300 font-mono flex items-center space-x-1.5">
              <span class="text-amber-500">🏷️</span>
              <span>EQUIPMENT TYPE & ICON</span>
            </span>
            <div class="flex items-center space-x-1 text-3xs font-mono text-slate-500 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
              <span>Preview:</span>
              <span class="text-blue-500 font-bold ml-1">{{ getEquipmentIcon(form.icon) }} {{ form.actuator_name || 'Actuator' }}</span>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                EQUIPMENT TYPE / ICON
              </label>
              <select
                v-model="form.icon"
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              >
                <option value="pump">💧 Pump (Water / Fluid)</option>
                <option value="valve">💨 Valve (Solenoid / Air Purge)</option>
                <option value="fan">🌀 Fan / Blower (Exhaust / Ventilation)</option>
                <option value="heater">🔥 Heater (Chamber / Line)</option>
                <option value="relay">⚡ Relay / Switch</option>
                <option value="light">💡 Light / Beacon Indicator</option>
                <option value="motor">⚙️ Motor Drive</option>
              </select>
            </div>
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                FEEDBACK PARAMETER (OPTIONAL)
              </label>
              <select
                v-model="form.feedback_parameter_id"
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              >
                <option :value="null">-- None (Logical State Only) --</option>
                <option v-for="p in parameters" :key="p.id" :value="p.id">
                  {{ p.name }} ({{ p.code }})
                </option>
              </select>
            </div>
          </div>
        </div>

        <!-- Section 3: MODBUS COMMUNICATION & CONTROL CONFIGURATION -->
        <div class="p-4 rounded-xl border border-slate-200/80 dark:border-slate-800 bg-white dark:bg-slate-800/40 shadow-sm space-y-4">
          <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-2.5">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-300 font-mono flex items-center space-x-1.5">
              <span class="text-amber-500">⚡</span>
              <span>MODBUS COMMUNICATION & CONTROL CONFIGURATION</span>
            </span>
            <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-amber-50 dark:bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-500/20">
              MODBUS &bull; SLAVE #{{ form.slave_id }}
            </span>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                CONTROL TYPE / FUNCTION CODE <span class="text-rose-500">*</span>
              </label>
              <select
                v-model="form.control_type"
                @change="handleControlTypeChange"
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              >
                <option value="COIL_FC05">Coil (FC 05 - Write Single Coil)</option>
                <option value="HOLDING_REGISTER_FC06">Holding Register (FC 06 - Write Single Register)</option>
                <option value="MULTIPLE_COILS_FC15">Multiple Coils (FC 15)</option>
                <option value="MULTIPLE_REGISTERS_FC16">Multiple Registers (FC 16)</option>
              </select>
            </div>

            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                SLAVE ID / UNIT ID <span class="text-slate-400 text-3xs">(1-247)</span>
              </label>
              <input
                type="number"
                min="1"
                max="247"
                v-model.number="form.slave_id"
                required
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              />
            </div>

            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                REGISTER ADDRESS / OFFSET <span class="text-rose-500">*</span>
              </label>
              <input
                type="number"
                min="0"
                max="65535"
                v-model.number="form.register_address"
                required
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-1 border-t border-slate-100 dark:border-slate-800">
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 font-mono">
                COMMAND "ON" PAYLOAD
              </label>
              <input
                type="number"
                v-model.number="form.on_payload"
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              />
              <span class="text-3xs text-slate-400 font-mono mt-1 block">
                {{ form.control_type === 'COIL_FC05' ? 'Standard FC05: 65280 (0xFF00)' : 'Standard FC06: 1' }}
              </span>
            </div>

            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 font-mono">
                COMMAND "OFF" PAYLOAD
              </label>
              <input
                type="number"
                v-model.number="form.off_payload"
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-blue-500/30 focus:border-blue-500"
              />
              <span class="text-3xs text-slate-400 font-mono mt-1 block">
                Standard: 0 (0x0000)
              </span>
            </div>
          </div>
        </div>

        <!-- Section 4: SAFETY INTERLOCK & AUTO-OFF WATCHDOG -->
        <div class="p-4 rounded-xl border border-slate-200/80 dark:border-slate-800 bg-white dark:bg-slate-800/40 shadow-sm space-y-4">
          <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-2.5">
            <span class="text-xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-300 font-mono flex items-center space-x-1.5">
              <span class="text-emerald-500">🛡️</span>
              <span>SAFETY INTERLOCK & AUTO-OFF WATCHDOG</span>
            </span>
            <label class="flex items-center space-x-2 text-xs font-semibold cursor-pointer text-slate-700 dark:text-slate-300">
              <input type="checkbox" v-model="form.safety_watchdog_active" class="rounded text-emerald-600 focus:ring-emerald-500 h-4 w-4 border-slate-300 dark:border-slate-700 dark:bg-slate-800" />
              <span>Watchdog Active</span>
            </label>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                MAX CONTINUOUS RUNTIME (SECONDS)
              </label>
              <input
                type="number"
                min="10"
                max="86400"
                v-model.number="form.safety_timeout_seconds"
                class="w-full px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-emerald-500/30 focus:border-emerald-500"
              />
              <span class="text-3xs text-slate-400 mt-1 block">
                ≈ {{ formatMinutes(form.safety_timeout_seconds) }} minutes.
              </span>
            </div>
            <div class="flex items-center text-xs text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-900/50 p-3 rounded-lg border border-slate-200/60 dark:border-slate-800">
              Automatically turns OFF the equipment if left running beyond this limit to prevent motor overheating, overfill, or pipe burst.
            </div>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="pt-4 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between">
          <button
            type="button"
            @click="openTestModal"
            class="px-3.5 py-2 text-xs font-semibold rounded-lg bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20 hover:bg-amber-500/20 transition-colors flex items-center space-x-1.5"
          >
            <span>⚡</span>
            <span>Test Connection & Write</span>
          </button>

          <div class="flex items-center space-x-2.5">
            <button
              type="button"
              @click="close"
              class="px-4 py-2 text-xs font-semibold rounded-lg text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="saving"
              class="px-5 py-2 text-xs font-bold rounded-lg bg-blue-600 hover:bg-blue-500 active:bg-blue-700 text-white shadow-sm transition-all disabled:opacity-50 flex items-center space-x-1.5"
            >
              <span v-if="saving" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
              <span>{{ isEditing ? 'Update Actuator' : 'Save Actuator' }}</span>
            </button>
          </div>
        </div>
      </form>
    </div>

    <!-- Inline Test Write Modal -->
    <ActuatorTestWriteModal
      :isOpen="showTestModal"
      :actuator="isEditing ? form : null"
      :deviceId="deviceId"
      :rawConfig="!isEditing ? form : null"
      @close="showTestModal = false"
    />
  </div>
</template>

<script>
import axios from 'axios';
import ActuatorTestWriteModal from './ActuatorTestWriteModal.vue';

export default {
  name: 'ActuatorModal',
  components: { ActuatorTestWriteModal },
  props: {
    isOpen: { type: Boolean, default: false },
    actuator: { type: Object, default: null },
    deviceId: { type: [Number, String], required: true },
    deviceCode: { type: String, default: '' },
    deviceSlaveId: { type: Number, default: 1 },
    parameters: { type: Array, default: () => [] }
  },
  data() {
    return {
      saving: false,
      errorMessage: '',
      showTestModal: false,
      form: {
        id: null,
        actuator_code: '',
        actuator_name: '',
        description: '',
        icon: 'pump',
        control_type: 'COIL_FC05',
        slave_id: 1,
        register_address: 0,
        on_payload: 65280,
        off_payload: 0,
        safety_timeout_seconds: 900,
        safety_watchdog_active: true,
        feedback_parameter_id: null,
        enabled: true
      }
    };
  },
  computed: {
    isEditing() {
      return !!(this.actuator && this.actuator.id);
    }
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.resetForm();
      }
    }
  },
  methods: {
    resetForm() {
      this.errorMessage = '';
      if (this.actuator && this.actuator.id) {
        this.form = { ...this.actuator };
      } else {
        this.form = {
          id: null,
          actuator_code: '',
          actuator_name: '',
          description: '',
          icon: 'pump',
          control_type: 'COIL_FC05',
          slave_id: this.deviceSlaveId || 1,
          register_address: 0,
          on_payload: 65280,
          off_payload: 0,
          safety_timeout_seconds: 900,
          safety_watchdog_active: true,
          feedback_parameter_id: null,
          enabled: true
        };
      }
    },
    handleControlTypeChange() {
      if (this.form.control_type === 'COIL_FC05') {
        this.form.on_payload = 65280; // 0xFF00
        this.form.off_payload = 0;
      } else if (this.form.control_type === 'HOLDING_REGISTER_FC06') {
        this.form.on_payload = 1;
        this.form.off_payload = 0;
      }
    },
    getEquipmentIcon(type) {
      const map = {
        pump: '💧',
        valve: '💨',
        fan: '🌀',
        heater: '🔥',
        relay: '⚡',
        light: '💡',
        motor: '⚙️'
      };
      return map[type] || '⚡';
    },
    formatMinutes(sec) {
      if (!sec) return 0;
      return (sec / 60).toFixed(1);
    },
    openTestModal() {
      this.showTestModal = true;
    },
    close() {
      this.$emit('close');
    },
    async handleSubmit() {
      this.saving = true;
      this.errorMessage = '';

      try {
        let res;
        if (this.isEditing) {
          res = await axios.put(`/api/actuators/${this.form.id}`, this.form);
        } else {
          res = await axios.post(`/api/devices/${this.deviceId}/actuators`, this.form);
        }
        this.$emit('saved', res.data.data || res.data);
        this.close();
      } catch (err) {
        const errData = err.response && err.response.data;
        this.errorMessage = (errData && (errData.error || errData.message)) || err.message;
      } finally {
        this.saving = false;
      }
    }
  }
};
</script>
