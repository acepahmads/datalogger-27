<template>
  <div class="space-y-6">
    <!-- Header / Toolbar -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center space-x-2.5">
          <h2 class="text-base font-bold text-slate-800 dark:text-slate-100">Controls & Actuators</h2>
          <span class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-blue-50 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400 border border-blue-200 dark:border-blue-500/30">
            {{ actuators.length }} Configured
          </span>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          Direct control of pumps, purging valves, and relays with automated interval schedules and auto-off watchdogs.
        </p>
      </div>

      <div class="flex items-center space-x-2.5 self-start md:self-auto">
        <button
          type="button"
          @click="showLogsModal = true"
          class="px-3 py-2 text-xs font-semibold rounded-lg bg-slate-100 hover:bg-slate-200/70 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors flex items-center space-x-1.5"
        >
          <span>📜</span>
          <span>Execution Logs</span>
        </button>

        <button
          type="button"
          @click="openAddModal"
          class="px-4 py-2 text-xs font-bold rounded-lg bg-blue-600 hover:bg-blue-500 active:bg-blue-700 text-white shadow-sm transition-all flex items-center space-x-1.5"
        >
          <span>+</span>
          <span>Add Actuator</span>
        </button>
      </div>
    </div>

    <!-- Actuators Grid -->
    <div v-if="loading" class="py-16 text-center text-xs text-slate-400">
      Loading actuators...
    </div>

    <div v-else-if="actuators.length === 0" class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-12 text-center text-slate-400 dark:text-slate-500">
      <p class="text-3xl mb-2">⚡</p>
      <h3 class="text-sm font-bold text-slate-700 dark:text-slate-300">No Actuators Configured</h3>
      <p class="text-xs mt-1 max-w-md mx-auto">
        This device does not have any control actuators configured yet. Add an actuator to control pumps, valves, or relays via Modbus FC05/FC06.
      </p>
      <button
        type="button"
        @click="openAddModal"
        class="mt-4 px-4 py-2 text-xs font-bold rounded-lg bg-blue-600 hover:bg-blue-500 text-white shadow-sm inline-flex items-center space-x-1.5"
      >
        <span>+ Add First Actuator</span>
      </button>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="act in actuators"
        :key="act.id"
        class="bg-white dark:bg-slate-900 border rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-all flex flex-col justify-between"
        :class="act.current_state === 'ON' ? 'border-emerald-500/50 dark:border-emerald-500/40 ring-1 ring-emerald-500/20' : 'border-slate-200 dark:border-slate-800'"
      >
        <!-- Card Header -->
        <div class="p-5 border-b border-slate-100 dark:border-slate-800/80 bg-slate-50/30 dark:bg-slate-800/30">
          <div class="flex items-start justify-between">
            <div class="flex items-center space-x-3">
              <div
                class="w-10 h-10 rounded-xl flex items-center justify-center text-lg"
                :class="act.current_state === 'ON' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20' : 'bg-slate-100 dark:bg-slate-800 text-slate-500 border border-slate-200 dark:border-slate-700'"
              >
                {{ getEquipmentIcon(act.icon) }}
              </div>
              <div>
                <h3 class="font-bold text-xs text-slate-800 dark:text-slate-100">{{ act.actuator_name }}</h3>
                <span class="text-3xs font-mono text-slate-400 dark:text-slate-500 block">{{ act.actuator_code }}</span>
              </div>
            </div>

            <!-- State Badge -->
            <span
              :class="act.current_state === 'ON' ? 'bg-emerald-500 text-white shadow-sm font-extrabold' : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700 font-bold'"
              class="px-2.5 py-1 rounded-full text-3xs font-mono tracking-wider flex items-center space-x-1.5"
            >
              <span class="w-1.5 h-1.5 rounded-full" :class="act.current_state === 'ON' ? 'bg-white animate-pulse' : 'bg-slate-400'"></span>
              <span>{{ act.current_state === 'ON' ? 'RUNNING (ON)' : 'STANDBY (OFF)' }}</span>
            </span>
          </div>

          <!-- Description if any -->
          <p v-if="act.description" class="text-3xs text-slate-500 dark:text-slate-400 mt-2.5 line-clamp-2">
            {{ act.description }}
          </p>

          <!-- Auto-Off Countdown Timer (if running and due) -->
          <div v-if="act.current_state === 'ON' && getCountdown(act)" class="mt-3 p-2 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800 flex items-center justify-between text-3xs font-mono">
            <span class="text-emerald-700 dark:text-emerald-400 flex items-center space-x-1 font-sans font-medium">
              <span>⏱️ Auto-Off Countdown:</span>
            </span>
            <span class="font-bold text-emerald-800 dark:text-emerald-300">
              {{ getCountdown(act) }} remaining
            </span>
          </div>
        </div>

        <!-- Technical Parameters Details -->
        <div class="p-4 space-y-2 text-2xs font-mono text-slate-500 dark:text-slate-400 bg-white dark:bg-slate-900">
          <div class="flex justify-between items-center">
            <span>Modbus Function:</span>
            <span class="font-bold text-slate-700 dark:text-slate-300">{{ act.control_type }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span>Slave & Address:</span>
            <span class="font-bold text-slate-700 dark:text-slate-300">Slave #{{ act.slave_id }} &bull; Offset {{ act.register_address }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span>Watchdog Interlock:</span>
            <span class="text-slate-700 dark:text-slate-300">
              {{ act.safety_watchdog_active ? `${Math.round(act.safety_timeout_seconds / 60)}m Max` : 'Disabled' }}
            </span>
          </div>
          <div class="flex justify-between items-center">
            <span>Automation Routines:</span>
            <span class="text-indigo-600 dark:text-indigo-400 font-bold">
              {{ act.schedules ? act.schedules.length : 0 }} active
            </span>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="p-4 border-t border-slate-100 dark:border-slate-800/80 bg-slate-50/50 dark:bg-slate-800/40 flex flex-col space-y-2.5">
          <!-- Manual Toggle Button -->
          <button
            v-if="act.current_state === 'ON'"
            type="button"
            @click="toggleActuation(act, 'OFF')"
            :disabled="togglingId === act.id"
            class="w-full py-2 px-3 rounded-lg bg-rose-600 hover:bg-rose-500 active:bg-rose-700 text-white font-bold text-xs shadow-sm transition-colors flex items-center justify-center space-x-1.5 disabled:opacity-50"
          >
            <span v-if="togglingId === act.id" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span>Manual STOP</span>
          </button>

          <button
            v-else
            type="button"
            @click="toggleActuation(act, 'ON')"
            :disabled="togglingId === act.id"
            class="w-full py-2 px-3 rounded-lg bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white font-bold text-xs shadow-sm transition-colors flex items-center justify-center space-x-1.5 disabled:opacity-50"
          >
            <span v-if="togglingId === act.id" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span>Manual START</span>
          </button>

          <!-- Utilities Row -->
          <div class="grid grid-cols-4 gap-1.5 pt-1">
            <button
              type="button"
              @click="openTestModal(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-center justify-center space-x-1"
              title="Test Write"
            >
              <span>⚡</span>
              <span>Test</span>
            </button>

            <button
              type="button"
              @click="openScheduleModal(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-indigo-600 dark:text-indigo-400 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-center justify-center space-x-1"
              title="Manage Automation Schedules"
            >
              <span>⏱️</span>
              <span>Sched</span>
            </button>

            <button
              type="button"
              @click="openEditModal(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-center justify-center space-x-1"
              title="Edit Actuator Settings"
            >
              <span>✏️</span>
              <span>Edit</span>
            </button>

            <button
              type="button"
              @click="deleteActuator(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors flex items-center justify-center"
              title="Delete Actuator"
            >
              <span>🗑️</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <ActuatorModal
      :isOpen="showModal"
      :actuator="selectedActuator"
      :deviceId="deviceId"
      :deviceCode="deviceCode"
      :deviceSlaveId="deviceSlaveId"
      :parameters="parameters"
      @close="showModal = false"
      @saved="fetchActuators"
    />

    <ActuatorTestWriteModal
      :isOpen="showTestModal"
      :actuator="selectedActuator"
      :deviceId="deviceId"
      @close="showTestModal = false"
      @tested="fetchActuators"
    />

    <ActuatorScheduleModal
      :isOpen="showScheduleModal"
      :actuator="selectedActuator"
      @close="showScheduleModal = false"
      @updated="fetchActuators"
    />

    <ActuatorLogsModal
      :isOpen="showLogsModal"
      :deviceId="deviceId"
      @close="showLogsModal = false"
    />
  </div>
</template>

<script>
import axios from 'axios';
import ActuatorModal from './ActuatorModal.vue';
import ActuatorTestWriteModal from './ActuatorTestWriteModal.vue';
import ActuatorScheduleModal from './ActuatorScheduleModal.vue';
import ActuatorLogsModal from './ActuatorLogsModal.vue';

export default {
  name: 'ActuatorTab',
  components: {
    ActuatorModal,
    ActuatorTestWriteModal,
    ActuatorScheduleModal,
    ActuatorLogsModal
  },
  props: {
    deviceId: { type: [Number, String], required: true },
    deviceCode: { type: String, default: '' },
    deviceSlaveId: { type: Number, default: 1 },
    parameters: { type: Array, default: () => [] }
  },
  data() {
    return {
      actuators: [],
      loading: false,
      togglingId: null,
      showModal: false,
      showTestModal: false,
      showScheduleModal: false,
      showLogsModal: false,
      selectedActuator: null,
      timerInterval: null,
      now: Date.now()
    };
  },
  watch: {
    deviceId() {
      this.fetchActuators();
    }
  },
  mounted() {
    this.fetchActuators();
    this.timerInterval = setInterval(() => {
      this.now = Date.now();
    }, 1000);
    window.addEventListener('actuator-state-event', this.onActuatorStateEvent);
  },
  beforeDestroy() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
    }
    window.removeEventListener('actuator-state-event', this.onActuatorStateEvent);
  },
  methods: {
    async fetchActuators() {
      this.loading = true;
      try {
        const res = await axios.get(`/api/devices/${this.deviceId}/actuators`);
        this.actuators = res.data.data || [];
      } catch (err) {
        console.error('Failed to fetch device actuators:', err);
      } finally {
        this.loading = false;
      }
    },
    onActuatorStateEvent(e) {
      const data = e.detail;
      if (!data || Number(data.device_id) !== Number(this.deviceId)) return;
      const idx = this.actuators.findIndex(a => a.id === data.actuator_id);
      if (idx !== -1) {
        this.$set(this.actuators[idx], 'current_state', data.state);
        this.$set(this.actuators[idx], 'auto_off_due_at', data.auto_off_due_at || null);
      } else {
        this.fetchActuators();
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
    getCountdown(act) {
      if (!act.auto_off_due_at) return null;
      const due = new Date(act.auto_off_due_at).getTime();
      const diff = Math.max(0, Math.floor((due - this.now) / 1000));
      if (diff <= 0) return '00:00';
      const m = Math.floor(diff / 60).toString().padStart(2, '0');
      const s = (diff % 60).toString().padStart(2, '0');
      return `${m}:${s}`;
    },
    openAddModal() {
      this.selectedActuator = null;
      this.showModal = true;
    },
    openEditModal(act) {
      this.selectedActuator = act;
      this.showModal = true;
    },
    openTestModal(act) {
      this.selectedActuator = act;
      this.showTestModal = true;
    },
    openScheduleModal(act) {
      this.selectedActuator = act;
      this.showScheduleModal = true;
    },
    async toggleActuation(act, action) {
      this.togglingId = act.id;
      try {
        await axios.post(`/api/actuators/${act.id}/actuate`, {
          action: action
        });
        await this.fetchActuators();
      } catch (err) {
        const errData = err.response && err.response.data;
        alert(`Failed to actuate ${act.actuator_name}: ` + ((errData && (errData.error || errData.message)) || err.message));
      } finally {
        this.togglingId = null;
      }
    },
    async deleteActuator(act) {
      if (!confirm(`Are you sure you want to delete actuator "${act.actuator_name}"? All associated schedules and logs will also be removed.`)) return;
      try {
        await axios.delete(`/api/actuators/${act.id}`);
        await this.fetchActuators();
      } catch (err) {
        const errData = err.response && err.response.data;
        alert('Failed to delete actuator: ' + ((errData && (errData.error || errData.message)) || err.message));
      }
    }
  }
};
</script>
