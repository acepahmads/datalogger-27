<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/70 backdrop-blur-sm">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl max-w-2xl w-full overflow-hidden flex flex-col max-h-[90vh]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between bg-slate-50/50 dark:bg-slate-900/50">
        <div class="flex items-center space-x-3">
          <div class="w-9 h-9 rounded-lg bg-indigo-500/10 dark:bg-indigo-400/10 border border-indigo-500/20 text-indigo-600 dark:text-indigo-400 flex items-center justify-center font-bold text-sm">
            ⏱️
          </div>
          <div>
            <h3 class="font-bold text-slate-800 dark:text-slate-100 text-sm">
              Automation Schedulers
            </h3>
            <p class="text-xs text-slate-500 dark:text-slate-400">
              {{ actuator ? actuator.actuator_name : '' }} &bull; Recurring cycles & automatic shutoff
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="close"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1 rounded-md"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Content -->
      <div class="p-6 overflow-y-auto space-y-6">
        <!-- New Schedule Form Toggle -->
        <div v-if="!showForm" class="flex justify-between items-center">
          <span class="text-xs font-bold uppercase tracking-wider text-slate-500 font-mono">Active Routines</span>
          <button
            type="button"
            @click="openNewSchedule"
            class="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm flex items-center space-x-1.5 transition-colors"
          >
            <span>+</span>
            <span>Add Schedule Routine</span>
          </button>
        </div>

        <!-- Add / Edit Schedule Form -->
        <div v-if="showForm" class="p-4 rounded-xl border border-indigo-200 dark:border-indigo-900/40 bg-indigo-50/20 dark:bg-indigo-950/20 space-y-4">
          <div class="flex items-center justify-between border-b border-indigo-100 dark:border-indigo-900/40 pb-2">
            <span class="text-xs font-bold text-indigo-700 dark:text-indigo-400 uppercase font-mono">
              {{ editingScheduleId ? 'Edit Schedule Routine' : 'New Schedule Routine' }}
            </span>
            <label class="flex items-center space-x-2 text-xs font-semibold cursor-pointer text-slate-700 dark:text-slate-300">
              <input type="checkbox" v-model="form.enabled" class="rounded text-indigo-600 focus:ring-indigo-500 h-4 w-4 border-slate-300 dark:border-slate-700 dark:bg-slate-800" />
              <span>Enabled</span>
            </label>
          </div>

          <div class="space-y-3.5">
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 font-mono">
                ROUTINE NAME *
              </label>
              <input
                type="text"
                v-model="form.schedule_name"
                placeholder="e.g. 2-Hour Purge & Flush Cycle"
                required
                class="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
              />
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
              <div>
                <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 font-mono">
                  REPEAT INTERVAL (CYCLE)
                </label>
                <div class="flex space-x-2">
                  <input
                    type="number"
                    min="1"
                    v-model.number="intervalValue"
                    class="w-2/3 px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                  />
                  <select
                    v-model="intervalUnit"
                    class="w-1/3 px-2 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                  >
                    <option value="minutes">Minutes</option>
                    <option value="hours">Hours</option>
                    <option value="seconds">Seconds</option>
                  </select>
                </div>
              </div>

              <div>
                <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 font-mono">
                  ACTIVE ON DURATION
                </label>
                <div class="flex space-x-2">
                  <input
                    type="number"
                    min="1"
                    v-model.number="durationValue"
                    class="w-2/3 px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                  />
                  <select
                    v-model="durationUnit"
                    class="w-1/3 px-2 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                  >
                    <option value="minutes">Minutes</option>
                    <option value="seconds">Seconds</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- Days of Week Selector -->
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 font-mono">
                DAYS OF WEEK
              </label>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  v-for="d in daysList"
                  :key="d.id"
                  @click="toggleDay(d.id)"
                  :class="isDayActive(d.id) ? 'bg-indigo-600 text-white font-bold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700'"
                  class="px-2.5 py-1 text-2xs rounded-md transition-colors"
                >
                  {{ d.name }}
                </button>
              </div>
            </div>

            <div class="flex justify-end space-x-2 pt-2">
              <button
                type="button"
                @click="cancelForm"
                class="px-3 py-1.5 text-xs font-semibold text-slate-600 dark:text-slate-400"
              >
                Cancel
              </button>
              <button
                type="button"
                @click="saveSchedule"
                :disabled="saving"
                class="px-4 py-1.5 text-xs font-bold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors flex items-center space-x-1.5"
              >
                <span v-if="saving" class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                <span>{{ editingScheduleId ? 'Save Changes' : 'Create Routine' }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Schedule List -->
        <div v-if="loadingSchedules" class="py-6 text-center text-xs text-slate-400">
          Loading schedules...
        </div>

        <div v-else-if="schedules.length === 0 && !showForm" class="py-8 text-center text-slate-400 dark:text-slate-500 text-xs">
          <p class="text-xl mb-1">⏱️</p>
          <p>No automated routines configured yet for this actuator.</p>
          <p class="text-3xs mt-1 text-slate-400">Click "+ Add Schedule Routine" above to create periodic cycle.</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="s in schedules"
            :key="s.id"
            class="p-4 rounded-xl border border-slate-200/80 dark:border-slate-800 bg-white dark:bg-slate-800/40 hover:border-indigo-300 dark:hover:border-indigo-800 transition-colors flex items-center justify-between shadow-sm"
          >
            <div class="space-y-1">
              <div class="flex items-center space-x-2">
                <span class="font-bold text-xs text-slate-800 dark:text-slate-100">{{ s.schedule_name }}</span>
                <span
                  :class="s.enabled ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30' : 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400 border-slate-200 dark:border-slate-700'"
                  class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold border"
                >
                  {{ s.enabled ? 'ACTIVE' : 'PAUSED' }}
                </span>
              </div>
              <div class="text-slate-500 dark:text-slate-400 flex items-center space-x-3 text-3xs font-mono">
                <span>Repeat: {{ formatDuration(s.interval_seconds) }}</span>
                <span>&bull;</span>
                <span>Duration: {{ formatDuration(s.duration_seconds) }}</span>
                <span>&bull;</span>
                <span>Next: {{ formatDate(s.next_run_at) }}</span>
              </div>
            </div>

            <div class="flex items-center space-x-2">
              <button
                type="button"
                @click="editSchedule(s)"
                class="p-1.5 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                title="Edit Routine"
              >
                ✏️
              </button>
              <button
                type="button"
                @click="deleteSchedule(s.id)"
                class="p-1.5 text-slate-400 hover:text-rose-600 dark:hover:text-rose-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                title="Delete Routine"
              >
                🗑️
              </button>
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
  name: 'ActuatorScheduleModal',
  props: {
    isOpen: { type: Boolean, default: false },
    actuator: { type: Object, default: null }
  },
  data() {
    return {
      schedules: [],
      loadingSchedules: false,
      showForm: false,
      saving: false,
      editingScheduleId: null,
      intervalValue: 2,
      intervalUnit: 'hours',
      durationValue: 5,
      durationUnit: 'minutes',
      selectedDays: ['1', '2', '3', '4', '5', '6', '7'],
      daysList: [
        { id: '1', name: 'Mon' },
        { id: '2', name: 'Tue' },
        { id: '3', name: 'Wed' },
        { id: '4', name: 'Thu' },
        { id: '5', name: 'Fri' },
        { id: '6', name: 'Sat' },
        { id: '7', name: 'Sun' }
      ],
      form: {
        schedule_name: '',
        schedule_type: 'INTERVAL',
        interval_seconds: 7200,
        duration_seconds: 300,
        days_of_week: '1,2,3,4,5,6,7',
        enabled: true
      }
    };
  },
  watch: {
    isOpen(newVal) {
      if (newVal && this.actuator) {
        this.fetchSchedules();
      }
    }
  },
  methods: {
    close() {
      this.cancelForm();
      this.$emit('close');
    },
    async fetchSchedules() {
      if (!this.actuator) return;
      this.loadingSchedules = true;
      try {
        const res = await axios.get(`/api/actuators/${this.actuator.id}/schedules`);
        this.schedules = res.data.data || [];
      } catch (err) {
        console.error('Failed to fetch actuator schedules:', err);
      } finally {
        this.loadingSchedules = false;
      }
    },
    openNewSchedule() {
      this.editingScheduleId = null;
      this.form = {
        schedule_name: '',
        schedule_type: 'INTERVAL',
        interval_seconds: 7200,
        duration_seconds: 300,
        days_of_week: '1,2,3,4,5,6,7',
        enabled: true
      };
      this.intervalValue = 2;
      this.intervalUnit = 'hours';
      this.durationValue = 5;
      this.durationUnit = 'minutes';
      this.selectedDays = ['1', '2', '3', '4', '5', '6', '7'];
      this.showForm = true;
    },
    editSchedule(s) {
      this.editingScheduleId = s.id;
      this.form = { ...s };

      if (s.interval_seconds % 3600 === 0) {
        this.intervalValue = s.interval_seconds / 3600;
        this.intervalUnit = 'hours';
      } else if (s.interval_seconds % 60 === 0) {
        this.intervalValue = s.interval_seconds / 60;
        this.intervalUnit = 'minutes';
      } else {
        this.intervalValue = s.interval_seconds;
        this.intervalUnit = 'seconds';
      }

      if (s.duration_seconds % 60 === 0) {
        this.durationValue = s.duration_seconds / 60;
        this.durationUnit = 'minutes';
      } else {
        this.durationValue = s.duration_seconds;
        this.durationUnit = 'seconds';
      }

      this.selectedDays = s.days_of_week ? s.days_of_week.split(',') : ['1', '2', '3', '4', '5', '6', '7'];
      this.showForm = true;
    },
    cancelForm() {
      this.showForm = false;
      this.editingScheduleId = null;
    },
    toggleDay(d) {
      if (this.selectedDays.includes(d)) {
        if (this.selectedDays.length > 1) {
          this.selectedDays = this.selectedDays.filter(x => x !== d);
        }
      } else {
        this.selectedDays.push(d);
      }
    },
    isDayActive(d) {
      return this.selectedDays.includes(d);
    },
    formatDuration(sec) {
      if (!sec) return '0s';
      if (sec >= 3600) return `${(sec / 3600).toFixed(1)}h`;
      if (sec >= 60) return `${Math.round(sec / 60)}m`;
      return `${sec}s`;
    },
    formatDate(dStr) {
      if (!dStr) return 'Not scheduled';
      const d = new Date(dStr);
      return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    async saveSchedule() {
      if (!this.form.schedule_name || !this.form.schedule_name.trim()) {
        alert('Routine name is required');
        return;
      }
      this.saving = true;

      let intervalSec = Number(this.intervalValue);
      if (this.intervalUnit === 'hours') intervalSec *= 3600;
      else if (this.intervalUnit === 'minutes') intervalSec *= 60;

      let durationSec = Number(this.durationValue);
      if (this.durationUnit === 'minutes') durationSec *= 60;

      this.form.interval_seconds = intervalSec;
      this.form.duration_seconds = durationSec;
      this.form.days_of_week = this.selectedDays.join(',');

      try {
        if (this.editingScheduleId) {
          await axios.put(`/api/actuators/schedules/${this.editingScheduleId}`, this.form);
        } else {
          await axios.post(`/api/actuators/${this.actuator.id}/schedules`, this.form);
        }
        await this.fetchSchedules();
        this.$emit('updated');
        this.cancelForm();
      } catch (err) {
        const errData = err.response && err.response.data;
        alert('Failed to save schedule: ' + ((errData && (errData.error || errData.message)) || err.message));
      } finally {
        this.saving = false;
      }
    },
    async deleteSchedule(id) {
      if (!confirm('Are you sure you want to delete this scheduled routine?')) return;
      try {
        await axios.delete(`/api/actuators/schedules/${id}`);
        await this.fetchSchedules();
        this.$emit('updated');
      } catch (err) {
        const errData = err.response && err.response.data;
        alert('Failed to delete schedule: ' + ((errData && (errData.error || errData.message)) || err.message));
      }
    }
  }
};
</script>
