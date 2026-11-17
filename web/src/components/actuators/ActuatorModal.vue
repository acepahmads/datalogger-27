<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/70 backdrop-blur-sm">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl max-w-3xl w-full overflow-hidden flex flex-col max-h-[92vh]">
      <!-- Header -->
      <div class="px-6 py-5 border-b border-slate-100 dark:border-slate-800 flex flex-wrap items-center justify-between gap-3 bg-slate-50/50 dark:bg-slate-900/50">
        <div class="flex items-center space-x-3.5">
          <div class="w-10 h-10 rounded-xl bg-blue-500/10 dark:bg-blue-400/10 border border-blue-500/20 text-blue-600 dark:text-blue-400 flex items-center justify-center text-lg">
            ⚡
          </div>
          <div>
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="text-base font-bold text-slate-900 dark:text-slate-100">
                {{ isEditing ? 'Edit Actuator / Control Equipment' : (isFromPaste ? 'Paste Actuator Configuration' : 'Add Actuator / Control Equipment') }}
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

        <div class="flex items-center space-x-2">
          <button
            type="button"
            @click="copyCurrentFormConfig"
            class="px-2.5 py-1.5 rounded-lg text-3xs font-semibold border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-700 transition-colors flex items-center space-x-1"
            title="Copy konfigurasi form ini ke clipboard"
          >
            <span>{{ copiedBanner ? '✓' : '📋' }}</span>
            <span>{{ copiedBanner ? 'Copied!' : 'Copy Config' }}</span>
          </button>

          <button
            type="button"
            @click="pasteIntoCurrentForm"
            class="px-2.5 py-1.5 rounded-lg text-3xs font-semibold border border-indigo-200 dark:border-indigo-500/30 bg-indigo-50 dark:bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 hover:bg-indigo-100 dark:hover:bg-indigo-500/20 transition-colors flex items-center space-x-1"
            title="Paste konfigurasi actuator dari clipboard"
          >
            <span>📥</span>
            <span>Paste Config</span>
          </button>

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
      </div>

      <!-- Form Body -->
      <form ref="formBody" @submit.prevent="handleSubmit" class="p-6 overflow-y-auto space-y-5">
        <!-- Pasted Info & Schedule Clone Toggle -->
        <div
          v-if="!isEditing && (isFromPaste || copiedSchedules.length > 0)"
          class="p-3.5 rounded-xl bg-indigo-50/80 dark:bg-indigo-950/30 border border-indigo-200 dark:border-indigo-800/60 text-xs text-indigo-800 dark:text-indigo-200 space-y-2"
        >
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center space-x-2 font-semibold">
              <span>📋</span>
              <span>Konfigurasi hasil Paste{{ pasteSourceInfo ? ` (${pasteSourceInfo})` : '' }} — silakan sesuaikan bila perlu lalu klik Save Actuator.</span>
            </div>
          </div>
          <label
            v-if="copiedSchedules.length > 0"
            class="inline-flex items-center space-x-2 pt-1 cursor-pointer font-semibold text-indigo-700 dark:text-indigo-300"
          >
            <input
              type="checkbox"
              v-model="cloneCopiedSchedules"
              class="rounded text-indigo-600 focus:ring-indigo-500 h-4 w-4 border-indigo-300 dark:border-indigo-700 dark:bg-slate-800"
            />
            <span>Ikut salin {{ copiedSchedules.length }} jadwal otomatis (Automation Schedules)</span>
          </label>
        </div>

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
            <div class="flex items-center space-x-1 text-3xs font-mono text-slate-500 dark:text-slate-400 bg-slate-100 dark:bg-slate-800 px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700">
              <span>Preview:</span>
              <span class="text-blue-600 dark:text-blue-400 font-bold ml-1">{{ getEquipmentIcon(form.icon) }} {{ form.actuator_name || 'Actuator' }}</span>
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
                  {{ p.parameter_name || p.name }} ({{ p.parameter_code || p.code }})
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

        <!-- Inline Footer Error Alert (visible even when scrolled down) -->
        <div v-if="errorMessage" class="p-3 rounded-lg bg-rose-50 dark:bg-rose-500/10 border border-rose-200 dark:border-rose-500/20 text-xs text-rose-600 dark:text-rose-400 font-medium flex items-center space-x-2">
          <span>⚠️</span>
          <span>{{ errorMessage }}</span>
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

const ACTUATOR_CLIPBOARD_KEY = 'datalogger_actuator_clipboard_v1';

export default {
  name: 'ActuatorModal',
  components: { ActuatorTestWriteModal },
  props: {
    isOpen: { type: Boolean, default: false },
    actuator: { type: Object, default: null },
    draftActuator: { type: Object, default: null },
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
      copiedBanner: false,
      isFromPaste: false,
      pasteSourceInfo: '',
      copiedSchedules: [],
      cloneCopiedSchedules: true,
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
    resolveFeedbackParameterId(cfg) {
      if (!cfg) return null;
      if (cfg.feedback_parameter_id && Array.isArray(this.parameters)) {
        const byId = this.parameters.find(p => Number(p.id) === Number(cfg.feedback_parameter_id));
        if (byId) return Number(byId.id);
      }
      if (cfg.feedback_parameter_code && Array.isArray(this.parameters)) {
        const clean = String(cfg.feedback_parameter_code).trim().toLowerCase();
        const byCode = this.parameters.find(
          p => String(p.parameter_code || p.code || '').trim().toLowerCase() === clean
        );
        if (byCode) return Number(byCode.id);
      }
      return null;
    },
    resetForm() {
      this.errorMessage = '';
      this.copiedBanner = false;
      this.isFromPaste = false;
      this.pasteSourceInfo = '';
      this.copiedSchedules = [];
      this.cloneCopiedSchedules = true;

      if (this.actuator && this.actuator.id) {
        this.form = { ...this.actuator };
      } else if (this.draftActuator) {
        this.isFromPaste = true;
        this.pasteSourceInfo = this.draftActuator._paste_source || '';
        this.copiedSchedules = Array.isArray(this.draftActuator._copied_schedules)
          ? this.draftActuator._copied_schedules
          : [];
        this.form = {
          id: null,
          actuator_code: this.draftActuator.actuator_code || '',
          actuator_name: this.draftActuator.actuator_name || '',
          description: this.draftActuator.description || '',
          icon: this.draftActuator.icon || 'pump',
          control_type: this.draftActuator.control_type || 'COIL_FC05',
          slave_id: Number(this.draftActuator.slave_id) || this.deviceSlaveId || 1,
          register_address: Number(this.draftActuator.register_address) || 0,
          on_payload: this.draftActuator.on_payload !== undefined ? Number(this.draftActuator.on_payload) : 65280,
          off_payload: this.draftActuator.off_payload !== undefined ? Number(this.draftActuator.off_payload) : 0,
          safety_timeout_seconds: Number(this.draftActuator.safety_timeout_seconds) || 900,
          safety_watchdog_active: this.draftActuator.safety_watchdog_active !== false,
          feedback_parameter_id: this.resolveFeedbackParameterId(this.draftActuator),
          enabled: this.draftActuator.enabled !== false
        };
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
    copyCurrentFormConfig() {
      const matchedParam = Array.isArray(this.parameters)
        ? this.parameters.find(p => Number(p.id) === Number(this.form.feedback_parameter_id))
        : null;
      const schedules = (this.actuator && Array.isArray(this.actuator.schedules))
        ? this.actuator.schedules.map(s => ({
            schedule_name: s.schedule_name || '',
            schedule_type: s.schedule_type || 'INTERVAL',
            interval_seconds: Number(s.interval_seconds) || 7200,
            duration_seconds: Number(s.duration_seconds) || 300,
            minute_of_hour: s.minute_of_hour || '',
            start_time_str: s.start_time_str || '',
            end_time_str: s.end_time_str || '',
            cron_expression: s.cron_expression || '',
            days_of_week: s.days_of_week || '1,2,3,4,5,6,7',
            enabled: s.enabled !== false
          }))
        : [];

      const packet = {
        _type: 'datalogger_actuator_config_v1',
        copied_at: new Date().toISOString(),
        source_device_id: Number(this.deviceId) || 0,
        source_device_code: this.deviceCode || '',
        config: {
          actuator_code: String(this.form.actuator_code || '').trim(),
          actuator_name: String(this.form.actuator_name || '').trim(),
          description: String(this.form.description || '').trim(),
          icon: this.form.icon || 'pump',
          control_type: this.form.control_type || 'COIL_FC05',
          slave_id: Number(this.form.slave_id) || 1,
          register_address: Number(this.form.register_address) || 0,
          on_payload: this.form.on_payload === '' || this.form.on_payload === null || this.form.on_payload === undefined ? 0 : Number(this.form.on_payload),
          off_payload: this.form.off_payload === '' || this.form.off_payload === null || this.form.off_payload === undefined ? 0 : Number(this.form.off_payload),
          safety_timeout_seconds: Number(this.form.safety_timeout_seconds) || 900,
          safety_watchdog_active: Boolean(this.form.safety_watchdog_active),
          feedback_parameter_id: this.form.feedback_parameter_id ? Number(this.form.feedback_parameter_id) : null,
          feedback_parameter_code: matchedParam ? (matchedParam.parameter_code || matchedParam.code || '') : '',
          enabled: Boolean(this.form.enabled)
        },
        schedules
      };

      const jsonStr = JSON.stringify(packet, null, 2);
      try {
        localStorage.setItem(ACTUATOR_CLIPBOARD_KEY, jsonStr);
      } catch (e) { /* ignore storage quota */ }
      if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
        navigator.clipboard.writeText(jsonStr).catch(() => {});
      }
      this.copiedBanner = true;
      this.$emit('copied', packet);
      setTimeout(() => {
        this.copiedBanner = false;
      }, 2500);
    },
    async pasteIntoCurrentForm() {
      let packet = null;
      if (navigator.clipboard && typeof navigator.clipboard.readText === 'function') {
        try {
          const clipText = await navigator.clipboard.readText();
          if (clipText && clipText.includes('datalogger_actuator_config_v1')) {
            packet = JSON.parse(clipText);
          }
        } catch (e) { /* fallback to localStorage */ }
      }
      if (!packet) {
        try {
          const stored = localStorage.getItem(ACTUATOR_CLIPBOARD_KEY);
          if (stored) packet = JSON.parse(stored);
        } catch (e) { /* ignore */ }
      }
      if (!packet || !packet.config) {
        this.errorMessage = 'Belum ada konfigurasi actuator yang di-copy. Klik tombol Copy pada actuator terlebih dahulu.';
        return;
      }

      const cfg = packet.config;
      this.errorMessage = '';
      this.isFromPaste = true;
      this.pasteSourceInfo = packet.source_device_code ? `DEV: ${packet.source_device_code}` : (cfg.actuator_code || '');
      this.copiedSchedules = Array.isArray(packet.schedules) ? packet.schedules : [];

      if (!this.isEditing) {
        if (!this.form.actuator_code) this.form.actuator_code = cfg.actuator_code || '';
        if (!this.form.actuator_name) this.form.actuator_name = cfg.actuator_name || '';
      }
      this.form.description = cfg.description || '';
      this.form.icon = cfg.icon || 'pump';
      this.form.control_type = cfg.control_type || 'COIL_FC05';
      this.form.slave_id = Number(cfg.slave_id) || this.deviceSlaveId || 1;
      this.form.register_address = Number(cfg.register_address) || 0;
      this.form.on_payload = cfg.on_payload !== undefined ? Number(cfg.on_payload) : 65280;
      this.form.off_payload = cfg.off_payload !== undefined ? Number(cfg.off_payload) : 0;
      this.form.safety_timeout_seconds = Number(cfg.safety_timeout_seconds) || 900;
      this.form.safety_watchdog_active = cfg.safety_watchdog_active !== false;
      this.form.feedback_parameter_id = this.resolveFeedbackParameterId(cfg);
      this.form.enabled = cfg.enabled !== false;
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

      const payload = {
        actuator_code: String(this.form.actuator_code || '').trim(),
        actuator_name: String(this.form.actuator_name || '').trim(),
        description: String(this.form.description || '').trim(),
        icon: this.form.icon || 'pump',
        control_type: this.form.control_type || 'COIL_FC05',
        slave_id: Number(this.form.slave_id) || 1,
        register_address: Number(this.form.register_address) || 0,
        on_payload: this.form.on_payload === '' || this.form.on_payload === null || this.form.on_payload === undefined ? 0 : Number(this.form.on_payload),
        off_payload: this.form.off_payload === '' || this.form.off_payload === null || this.form.off_payload === undefined ? 0 : Number(this.form.off_payload),
        safety_timeout_seconds: Number(this.form.safety_timeout_seconds) || 900,
        safety_watchdog_active: Boolean(this.form.safety_watchdog_active),
        feedback_parameter_id: this.form.feedback_parameter_id ? Number(this.form.feedback_parameter_id) : null,
        enabled: Boolean(this.form.enabled)
      };

      try {
        let res;
        if (this.isEditing) {
          res = await axios.put(`/api/actuators/${this.form.id}`, payload);
        } else {
          res = await axios.post(`/api/devices/${this.deviceId}/actuators`, payload);
          const created = res.data.data || res.data;
          if (created && created.id && this.cloneCopiedSchedules && Array.isArray(this.copiedSchedules) && this.copiedSchedules.length > 0) {
            for (const sched of this.copiedSchedules) {
              try {
                await axios.post(`/api/actuators/${created.id}/schedules`, {
                  schedule_name: sched.schedule_name || 'Auto Schedule',
                  schedule_type: sched.schedule_type || 'INTERVAL',
                  interval_seconds: Number(sched.interval_seconds) || 7200,
                  duration_seconds: Number(sched.duration_seconds) || 300,
                  minute_of_hour: sched.minute_of_hour || '',
                  start_time_str: sched.start_time_str || '',
                  end_time_str: sched.end_time_str || '',
                  cron_expression: sched.cron_expression || '',
                  days_of_week: sched.days_of_week || '1,2,3,4,5,6,7',
                  enabled: sched.enabled !== false
                });
              } catch (schedErr) {
                console.warn('Failed to clone schedule during actuator paste:', schedErr);
              }
            }
          }
        }
        this.$emit('saved', res.data.data || res.data);
        this.close();
      } catch (err) {
        const errData = err.response && err.response.data;
        if (err.response && err.response.status === 404) {
          this.errorMessage = 'Endpoint /api/devices/' + this.deviceId + '/actuators tidak ditemukan (404). Pastikan binary backend di server sudah di-update dan service sudah di-restart.';
        } else {
          this.errorMessage = (errData && (errData.error || errData.message)) || err.message;
        }
      } finally {
        this.saving = false;
      }
    }
  }
};
</script>
