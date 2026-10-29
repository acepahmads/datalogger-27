<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/70 backdrop-blur-sm">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl max-w-2xl w-full overflow-hidden flex flex-col max-h-[92vh]">
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
              {{ actuator ? actuator.actuator_name : '' }} &bull; Custom timing routines & automatic shutoff
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
        <div v-if="showForm" class="p-4.5 rounded-xl border border-indigo-200 dark:border-indigo-900/40 bg-indigo-50/20 dark:bg-indigo-950/20 space-y-4">
          <div class="flex items-center justify-between border-b border-indigo-100 dark:border-indigo-900/40 pb-2.5">
            <span class="text-xs font-bold text-indigo-700 dark:text-indigo-400 uppercase font-mono flex items-center space-x-1.5">
              <span>⏱️</span>
              <span>{{ editingScheduleId ? 'Edit Schedule Routine' : 'New Schedule Routine' }}</span>
            </span>
            <label class="flex items-center space-x-2 text-xs font-semibold cursor-pointer text-slate-700 dark:text-slate-300">
              <input type="checkbox" v-model="form.enabled" class="rounded text-indigo-600 focus:ring-indigo-500 h-4 w-4 border-slate-300 dark:border-slate-700 dark:bg-slate-800" />
              <span>Enabled</span>
            </label>
          </div>

          <div class="space-y-4">
            <!-- Routine Name -->
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 font-mono">
                ROUTINE NAME *
              </label>
              <input
                type="text"
                v-model="form.schedule_name"
                placeholder="e.g. Hourly Air Purge on Minute 45"
                required
                class="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
              />
            </div>

            <!-- Schedule Mode Selector (Segmented Tabs) -->
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 font-mono">
                SCHEDULE TIMING MODE *
              </label>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-1.5 p-1 bg-slate-100 dark:bg-slate-800/80 rounded-xl border border-slate-200/80 dark:border-slate-700">
                <button
                  type="button"
                  @click="setScheduleType('HOURLY_MINUTE')"
                  :class="form.schedule_type === 'HOURLY_MINUTE' ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 shadow-xs font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 font-medium'"
                  class="py-1.5 px-2 text-2xs rounded-lg transition-all flex items-center justify-center space-x-1"
                >
                  <span>⏱️</span>
                  <span>Menit Tiap Jam</span>
                </button>

                <button
                  type="button"
                  @click="setScheduleType('TIME_OF_DAY')"
                  :class="form.schedule_type === 'TIME_OF_DAY' ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 shadow-xs font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 font-medium'"
                  class="py-1.5 px-2 text-2xs rounded-lg transition-all flex items-center justify-center space-x-1"
                >
                  <span>📅</span>
                  <span>Jam Spesifik</span>
                </button>

                <button
                  type="button"
                  @click="setScheduleType('INTERVAL')"
                  :class="form.schedule_type === 'INTERVAL' ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 shadow-xs font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 font-medium'"
                  class="py-1.5 px-2 text-2xs rounded-lg transition-all flex items-center justify-center space-x-1"
                >
                  <span>🔄</span>
                  <span>Siklus Interval</span>
                </button>

                <button
                  type="button"
                  @click="setScheduleType('CRON')"
                  :class="form.schedule_type === 'CRON' ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 shadow-xs font-bold' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 font-medium'"
                  class="py-1.5 px-2 text-2xs rounded-lg transition-all flex items-center justify-center space-x-1"
                >
                  <span>⚙️</span>
                  <span>Cron Kustom</span>
                </button>
              </div>
            </div>

            <!-- Dynamic Config Block 1: HOURLY_MINUTE -->
            <div v-if="form.schedule_type === 'HOURLY_MINUTE'" class="p-3.5 rounded-lg bg-white dark:bg-slate-900 border border-indigo-200/80 dark:border-indigo-900/50 space-y-2.5">
              <div class="flex items-center justify-between">
                <label class="block text-2xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-300 font-mono">
                  MENIT KE- DALAM SATU JAM (0 - 59) *
                </label>
                <span class="text-3xs text-slate-400 font-mono">Bisa multiple, misal: 15, 45</span>
              </div>
              <input
                type="text"
                v-model="form.minute_of_hour"
                placeholder="e.g. 45 atau 0, 30 atau 15, 45"
                required
                class="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/30 font-bold"
              />
              <!-- Preset Chips -->
              <div class="flex flex-wrap gap-1.5 items-center pt-1">
                <span class="text-3xs text-slate-400 font-medium">Quick Presets:</span>
                <button
                  type="button"
                  v-for="p in minutePresets"
                  :key="p.label"
                  @click="form.minute_of_hour = p.val"
                  :class="form.minute_of_hour === p.val ? 'bg-indigo-600 text-white font-bold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700'"
                  class="px-2 py-0.5 rounded text-3xs font-mono transition-colors"
                >
                  {{ p.label }}
                </button>
              </div>
              <!-- Explanatory Box -->
              <div class="text-2xs text-indigo-700 dark:text-indigo-300 bg-indigo-50/60 dark:bg-indigo-950/40 p-2.5 rounded-lg border border-indigo-100 dark:border-indigo-900/40 flex items-start space-x-2">
                <span>💡</span>
                <span>
                  Aktuator akan otomatis menyala <strong>setiap jam tepat pada menit ke-{{ form.minute_of_hour || '0' }}</strong>
                  (contoh: 01:{{ form.minute_of_hour || '00' }}, 02:{{ form.minute_of_hour || '00' }}, 03:{{ form.minute_of_hour || '00' }}, dst).
                </span>
              </div>
            </div>

            <!-- Dynamic Config Block 2: TIME_OF_DAY -->
            <div v-if="form.schedule_type === 'TIME_OF_DAY'" class="p-3.5 rounded-lg bg-white dark:bg-slate-900 border border-indigo-200/80 dark:border-indigo-900/50 space-y-2.5">
              <div class="flex items-center justify-between">
                <label class="block text-2xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-300 font-mono">
                  WAKTU EKSEKUSI HARIAN (JAM:MENIT) *
                </label>
                <span class="text-3xs text-slate-400 font-mono">Format 24 Jam (HH:mm)</span>
              </div>
              <input
                type="text"
                v-model="form.start_time_str"
                placeholder="e.g. 08:00 atau 08:00, 16:30"
                required
                class="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/30 font-bold"
              />
              <!-- Preset Chips -->
              <div class="flex flex-wrap gap-1.5 items-center pt-1">
                <span class="text-3xs text-slate-400 font-medium">Quick Presets:</span>
                <button
                  type="button"
                  v-for="p in timeOfDayPresets"
                  :key="p.label"
                  @click="form.start_time_str = p.val"
                  :class="form.start_time_str === p.val ? 'bg-indigo-600 text-white font-bold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700'"
                  class="px-2 py-0.5 rounded text-3xs font-mono transition-colors"
                >
                  {{ p.label }}
                </button>
              </div>
              <!-- Explanatory Box -->
              <div class="text-2xs text-indigo-700 dark:text-indigo-300 bg-indigo-50/60 dark:bg-indigo-950/40 p-2.5 rounded-lg border border-indigo-100 dark:border-indigo-900/40 flex items-start space-x-2">
                <span>💡</span>
                <span>
                  Aktuator akan otomatis menyala pada jam <strong>{{ form.start_time_str || '08:00' }}</strong> setiap hari yang aktif.
                </span>
              </div>
            </div>

            <!-- Dynamic Config Block 3: INTERVAL -->
            <div v-if="form.schedule_type === 'INTERVAL'" class="p-3.5 rounded-lg bg-white dark:bg-slate-900 border border-indigo-200/80 dark:border-indigo-900/50 space-y-2.5">
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-300 font-mono">
                INTERVAL PENGULANGAN (REPEAT CYCLE) *
              </label>
              <div class="flex space-x-2">
                <input
                  type="number"
                  min="1"
                  v-model.number="intervalValue"
                  class="w-2/3 px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/30 font-bold"
                />
                <select
                  v-model="intervalUnit"
                  class="w-1/3 px-2 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                >
                  <option value="minutes">Minutes</option>
                  <option value="hours">Hours</option>
                  <option value="seconds">Seconds</option>
                </select>
              </div>
              <!-- Preset Chips -->
              <div class="flex flex-wrap gap-1.5 items-center pt-1">
                <span class="text-3xs text-slate-400 font-medium">Quick Presets:</span>
                <button
                  type="button"
                  v-for="p in intervalPresets"
                  :key="p.label"
                  @click="setIntervalPreset(p.val, p.unit)"
                  class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700 transition-colors"
                >
                  {{ p.label }}
                </button>
              </div>
            </div>

            <!-- Dynamic Config Block 4: CRON -->
            <div v-if="form.schedule_type === 'CRON'" class="p-3.5 rounded-lg bg-white dark:bg-slate-900 border border-indigo-200/80 dark:border-indigo-900/50 space-y-2.5">
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-300 font-mono">
                CRON EXPRESSION (5-FIELD: MIN HOUR DOM MONTH DOW) *
              </label>
              <input
                type="text"
                v-model="form.cron_expression"
                placeholder="e.g. 45 * * * * atau */15 * * * *"
                required
                class="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/30 font-bold"
              />
              <div class="flex flex-wrap gap-1.5 items-center pt-1">
                <span class="text-3xs text-slate-400 font-medium">Quick Presets:</span>
                <button
                  type="button"
                  v-for="p in cronPresets"
                  :key="p.label"
                  @click="form.cron_expression = p.val"
                  :class="form.cron_expression === p.val ? 'bg-indigo-600 text-white font-bold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700'"
                  class="px-2 py-0.5 rounded text-3xs font-mono transition-colors"
                >
                  {{ p.label }}
                </button>
              </div>
            </div>

            <!-- Active ON Duration (Shared across all modes) -->
            <div>
              <div class="flex items-center justify-between mb-1">
                <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 font-mono">
                  ACTIVE ON DURATION (LAMA MENYALA) *
                </label>
                <span class="text-3xs text-slate-400 font-mono">Otomatis mati setelah durasi ini</span>
              </div>
              <div class="flex space-x-2">
                <input
                  type="number"
                  min="1"
                  v-model.number="durationValue"
                  class="w-2/3 px-3 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono focus:outline-none focus:ring-2 focus:ring-indigo-500/30 font-bold"
                />
                <select
                  v-model="durationUnit"
                  class="w-1/3 px-2 py-1.5 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                >
                  <option value="minutes">Minutes</option>
                  <option value="seconds">Seconds</option>
                </select>
              </div>
              <div class="flex flex-wrap gap-1.5 items-center pt-1.5">
                <span class="text-3xs text-slate-400 font-medium">Quick Durations:</span>
                <button
                  type="button"
                  v-for="d in durationPresets"
                  :key="d.label"
                  @click="setDurationPreset(d.val, d.unit)"
                  class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700 transition-colors"
                >
                  {{ d.label }}
                </button>
              </div>
            </div>

            <!-- Days of Week Selector -->
            <div>
              <div class="flex items-center justify-between mb-1">
                <label class="block text-2xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 font-mono">
                  DAYS OF WEEK (HARI AKTIF)
                </label>
                <div class="flex space-x-1.5">
                  <button type="button" @click="setAllDays" class="text-3xs text-indigo-600 dark:text-indigo-400 hover:underline">Semua Hari</button>
                  <span class="text-3xs text-slate-300 dark:text-slate-700">&bull;</span>
                  <button type="button" @click="setWeekdaysOnly" class="text-3xs text-indigo-600 dark:text-indigo-400 hover:underline">Sen-Jum</button>
                </div>
              </div>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  v-for="d in daysList"
                  :key="d.id"
                  @click="toggleDay(d.id)"
                  :class="isDayActive(d.id) ? 'bg-indigo-600 text-white font-bold shadow-xs' : 'bg-white dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700'"
                  class="px-2.5 py-1 text-2xs rounded-md transition-colors"
                >
                  {{ d.name }}
                </button>
              </div>
            </div>

            <!-- Footer Actions -->
            <div class="flex justify-end space-x-2 pt-2 border-t border-indigo-100 dark:border-indigo-900/40">
              <button
                type="button"
                @click="cancelForm"
                class="px-3 py-1.5 text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-800 transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                @click="saveSchedule"
                :disabled="saving"
                class="px-4 py-1.5 text-xs font-bold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition-colors flex items-center space-x-1.5 shadow-sm"
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
          <p class="text-3xs mt-1 text-slate-400">Click "+ Add Schedule Routine" above to configure hourly minute, daily, or cyclic routines.</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="s in schedules"
            :key="s.id"
            class="p-4 rounded-xl border border-slate-200/80 dark:border-slate-800 bg-white dark:bg-slate-800/40 hover:border-indigo-300 dark:hover:border-indigo-800 transition-colors flex items-center justify-between shadow-sm"
          >
            <div class="space-y-1.5">
              <div class="flex items-center space-x-2">
                <span class="font-bold text-xs text-slate-800 dark:text-slate-100">{{ s.schedule_name }}</span>

                <!-- Mode Tag Badge -->
                <span
                  class="px-2 py-0.5 rounded text-3xs font-mono font-bold"
                  :class="getModeBadgeClass(s.schedule_type)"
                >
                  {{ getModeBadgeLabel(s) }}
                </span>

                <!-- Active / Paused Badge -->
                <span
                  :class="s.enabled ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30' : 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400 border-slate-200 dark:border-slate-700'"
                  class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold border"
                >
                  {{ s.enabled ? 'ACTIVE' : 'PAUSED' }}
                </span>
              </div>

              <!-- Human Timing Description -->
              <div class="text-slate-500 dark:text-slate-400 flex flex-wrap items-center gap-x-3 gap-y-1 text-3xs font-mono">
                <span>{{ getTimingSummary(s) }}</span>
                <span>&bull;</span>
                <span>Durasi: {{ formatDuration(s.duration_seconds) }}</span>
                <span>&bull;</span>
                <span class="text-indigo-600 dark:text-indigo-400 font-semibold">Next: {{ formatDate(s.next_run_at) }}</span>
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
      minutePresets: [
        { label: ':00 (Awal Jam)', val: '0' },
        { label: ':15', val: '15' },
        { label: ':30 (Setengah Jam)', val: '30' },
        { label: ':45', val: '45' },
        { label: ':55 (Sebelum Ganti Jam)', val: '55' },
        { label: ':15, :45 (Dua Kali/Jam)', val: '15, 45' }
      ],
      timeOfDayPresets: [
        { label: '06:00 (Pagi)', val: '06:00' },
        { label: '08:00 (Awal Shift)', val: '08:00' },
        { label: '12:00 (Siang)', val: '12:00' },
        { label: '17:00 (Sore)', val: '17:00' },
        { label: '08:00, 16:30 (Dua Kali)', val: '08:00, 16:30' }
      ],
      intervalPresets: [
        { label: '15 Menit', val: 15, unit: 'minutes' },
        { label: '30 Menit', val: 30, unit: 'minutes' },
        { label: '1 Jam', val: 1, unit: 'hours' },
        { label: '2 Jam', val: 2, unit: 'hours' },
        { label: '6 Jam', val: 6, unit: 'hours' }
      ],
      cronPresets: [
        { label: '45 * * * * (Tiap jam m:45)', val: '45 * * * *' },
        { label: '0 8,17 * * * (Jam 8 & 17)', val: '0 8,17 * * *' },
        { label: '*/15 * * * * (Tiap 15m)', val: '*/15 * * * *' },
        { label: '0 8 * * 1-5 (Jam 8 Sen-Jum)', val: '0 8 * * 1-5' }
      ],
      durationPresets: [
        { label: '30 Detik', val: 30, unit: 'seconds' },
        { label: '1 Menit', val: 1, unit: 'minutes' },
        { label: '5 Menit', val: 5, unit: 'minutes' },
        { label: '10 Menit', val: 10, unit: 'minutes' },
        { label: '15 Menit', val: 15, unit: 'minutes' }
      ],
      form: {
        schedule_name: '',
        schedule_type: 'HOURLY_MINUTE',
        minute_of_hour: '45',
        start_time_str: '08:00',
        cron_expression: '45 * * * *',
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
    setScheduleType(type) {
      this.form.schedule_type = type;
      if (!this.form.schedule_name || this.form.schedule_name.startsWith('Hourly') || this.form.schedule_name.startsWith('Daily') || this.form.schedule_name.startsWith('Cycle')) {
        if (type === 'HOURLY_MINUTE') {
          this.form.schedule_name = `Hourly Run on Minute ${this.form.minute_of_hour || '45'}`;
        } else if (type === 'TIME_OF_DAY') {
          this.form.schedule_name = `Daily Run at ${this.form.start_time_str || '08:00'}`;
        } else if (type === 'INTERVAL') {
          this.form.schedule_name = `${this.intervalValue}-${this.intervalUnit} Cycle`;
        } else if (type === 'CRON') {
          this.form.schedule_name = `Custom Cron Routine`;
        }
      }
    },
    setIntervalPreset(val, unit) {
      this.intervalValue = val;
      this.intervalUnit = unit;
    },
    setDurationPreset(val, unit) {
      this.durationValue = val;
      this.durationUnit = unit;
    },
    setAllDays() {
      this.selectedDays = ['1', '2', '3', '4', '5', '6', '7'];
    },
    setWeekdaysOnly() {
      this.selectedDays = ['1', '2', '3', '4', '5'];
    },
    openNewSchedule() {
      this.editingScheduleId = null;
      this.form = {
        schedule_name: 'Hourly Run on Minute 45',
        schedule_type: 'HOURLY_MINUTE',
        minute_of_hour: '45',
        start_time_str: '08:00',
        cron_expression: '45 * * * *',
        interval_seconds: 3600,
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
      this.form = {
        ...s,
        schedule_type: s.schedule_type || 'INTERVAL',
        minute_of_hour: s.minute_of_hour || '45',
        start_time_str: s.start_time_str || '08:00',
        cron_expression: s.cron_expression || '45 * * * *'
      };

      if (s.interval_seconds % 3600 === 0) {
        this.intervalValue = s.interval_seconds / 3600;
        this.intervalUnit = 'hours';
      } else if (s.interval_seconds % 60 === 0) {
        this.intervalValue = s.interval_seconds / 60;
        this.intervalUnit = 'minutes';
      } else {
        this.intervalValue = s.interval_seconds || 15;
        this.intervalUnit = 'seconds';
      }

      if (s.duration_seconds % 60 === 0) {
        this.durationValue = s.duration_seconds / 60;
        this.durationUnit = 'minutes';
      } else {
        this.durationValue = s.duration_seconds || 60;
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
    getModeBadgeClass(type) {
      switch (type) {
        case 'HOURLY_MINUTE':
          return 'bg-indigo-50 text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-500/30';
        case 'TIME_OF_DAY':
          return 'bg-sky-50 text-sky-700 dark:bg-sky-500/10 dark:text-sky-400 border border-sky-200 dark:border-sky-500/30';
        case 'CRON':
          return 'bg-amber-50 text-amber-700 dark:bg-amber-500/10 dark:text-amber-400 border border-amber-200 dark:border-amber-500/30';
        case 'INTERVAL':
        default:
          return 'bg-purple-50 text-purple-700 dark:bg-purple-500/10 dark:text-purple-400 border border-purple-200 dark:border-purple-500/30';
      }
    },
    getModeBadgeLabel(s) {
      switch (s.schedule_type) {
        case 'HOURLY_MINUTE':
          return `HOURLY :${s.minute_of_hour || '0'}`;
        case 'TIME_OF_DAY':
          return `DAILY ${s.start_time_str || '08:00'}`;
        case 'CRON':
          return `CRON`;
        case 'INTERVAL':
        default:
          return `CYCLE ${this.formatDuration(s.interval_seconds)}`;
      }
    },
    getTimingSummary(s) {
      switch (s.schedule_type) {
        case 'HOURLY_MINUTE':
          return `Setiap jam di menit :${s.minute_of_hour || '0'}`;
        case 'TIME_OF_DAY':
          return `Harian pukul ${s.start_time_str || '08:00'}`;
        case 'CRON':
          return `Cron: ${s.cron_expression || '45 * * * *'}`;
        case 'INTERVAL':
        default:
          return `Siklus: tiap ${this.formatDuration(s.interval_seconds)}`;
      }
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

      const payload = {
        schedule_name: this.form.schedule_name.trim(),
        schedule_type: this.form.schedule_type || 'HOURLY_MINUTE',
        minute_of_hour: String(this.form.minute_of_hour || '45').trim(),
        start_time_str: String(this.form.start_time_str || '08:00').trim(),
        cron_expression: String(this.form.cron_expression || '45 * * * *').trim(),
        interval_seconds: intervalSec,
        duration_seconds: durationSec,
        days_of_week: this.selectedDays.join(','),
        enabled: Boolean(this.form.enabled)
      };

      try {
        if (this.editingScheduleId) {
          await axios.put(`/api/actuators/schedules/${this.editingScheduleId}`, payload);
        } else {
          await axios.post(`/api/actuators/${this.actuator.id}/schedules`, payload);
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