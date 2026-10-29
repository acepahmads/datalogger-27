<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-slate-900/70 backdrop-blur-sm">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl max-w-4xl w-full overflow-hidden flex flex-col max-h-[94vh]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between bg-slate-50/70 dark:bg-slate-900/70">
        <div class="flex items-center space-x-3.5">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-500/20 to-blue-500/20 border border-indigo-500/30 text-indigo-600 dark:text-indigo-400 flex items-center justify-center text-lg shadow-sm">
            ⏱️
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h3 class="font-bold text-slate-900 dark:text-slate-100 text-sm">
                Automation Schedulers & Execution Log
              </h3>
              <span v-if="actuator" class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-indigo-50 text-indigo-700 dark:bg-indigo-500/10 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-500/20">
                {{ actuator.actuator_name }} ({{ actuator.actuator_code }})
              </span>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
              Kelola waktu eksekusi otomatis (menit tertentu, harian, interval) dan pantau riwayat eksekusi.
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="close"
          class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
        >
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Main Body -->
      <div class="p-5 sm:p-6 overflow-y-auto space-y-6">
        <!-- FORM SECTION: Add / Edit Routine (Optimasi Tampilan Gambar 1) -->
        <div class="rounded-2xl border transition-all shadow-sm" :class="showForm ? 'border-indigo-300 dark:border-indigo-800 bg-gradient-to-b from-indigo-50/40 to-white dark:from-indigo-950/20 dark:to-slate-900 p-5' : 'border-slate-200 dark:border-slate-800 bg-slate-50/40 dark:bg-slate-800/20 p-4'">
          <!-- Form Header & Toggle Button -->
          <div class="flex items-center justify-between" :class="showForm ? 'border-b border-indigo-100 dark:border-indigo-900/40 pb-3 mb-4' : ''">
            <div class="flex items-center space-x-2">
              <span class="text-xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider font-mono flex items-center space-x-1.5">
                <span class="text-indigo-600 dark:text-indigo-400">⚡</span>
                <span>{{ showForm ? (editingScheduleId ? 'Edit Jadwal Rutinitas' : 'Tambah Jadwal Rutinitas Baru') : 'Konfigurasi Jadwal Otomatis' }}</span>
              </span>
              <span v-if="!showForm" class="text-3xs text-slate-400 font-mono">
                ({{ schedules.length }} jadwal terdaftar)
              </span>
            </div>

            <div class="flex items-center space-x-3">
              <label v-if="showForm" class="flex items-center space-x-2 text-xs font-semibold cursor-pointer text-slate-700 dark:text-slate-300">
                <input type="checkbox" v-model="form.enabled" class="rounded text-indigo-600 focus:ring-indigo-500 h-4 w-4 border-slate-300 dark:border-slate-700 dark:bg-slate-800" />
                <span>Aktif</span>
              </label>

              <button
                v-if="!showForm"
                type="button"
                @click="openNewSchedule"
                class="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white text-xs font-bold shadow-sm flex items-center space-x-1.5 transition-all"
              >
                <span>+</span>
                <span>Tambah Jadwal Baru</span>
              </button>
            </div>
          </div>

          <!-- Expanded Form Fields -->
          <div v-if="showForm" class="space-y-4.5">
            <!-- 1. Routine Name -->
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-400 mb-1 font-mono">
                NAMA JADWAL / RUTINITAS <span class="text-rose-500">*</span>
              </label>
              <input
                type="text"
                v-model="form.schedule_name"
                placeholder="e.g. Sampling Air Bersih Menit 45"
                required
                class="w-full px-3.5 py-2 text-xs rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30 font-medium"
              />
            </div>

            <!-- 2. Schedule Timing Mode Selector (Modern SCADA Segmented Tabs) -->
            <div>
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-400 mb-1.5 font-mono">
                PILIH MODE PENENTUAN WAKTU <span class="text-rose-500">*</span>
              </label>
              <div class="grid grid-cols-2 md:grid-cols-4 gap-2">
                <!-- Mode 1: HOURLY_MINUTE -->
                <button
                  type="button"
                  @click="setScheduleType('HOURLY_MINUTE')"
                  :class="form.schedule_type === 'HOURLY_MINUTE' ? 'bg-indigo-600 text-white shadow-sm ring-2 ring-indigo-500/30 border-transparent' : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 hover:border-slate-300 border-slate-200 dark:border-slate-800'"
                  class="p-2.5 rounded-xl border text-left transition-all flex flex-col justify-between"
                >
                  <div class="flex items-center space-x-1.5 text-xs font-bold">
                    <span>⏱️</span>
                    <span>Menit Tiap Jam</span>
                  </div>
                  <span class="text-3xs mt-1 block" :class="form.schedule_type === 'HOURLY_MINUTE' ? 'text-indigo-100' : 'text-slate-400'">
                    Misal: Menit ke-45 setiap jam
                  </span>
                </button>

                <!-- Mode 2: TIME_OF_DAY -->
                <button
                  type="button"
                  @click="setScheduleType('TIME_OF_DAY')"
                  :class="form.schedule_type === 'TIME_OF_DAY' ? 'bg-indigo-600 text-white shadow-sm ring-2 ring-indigo-500/30 border-transparent' : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 hover:border-slate-300 border-slate-200 dark:border-slate-800'"
                  class="p-2.5 rounded-xl border text-left transition-all flex flex-col justify-between"
                >
                  <div class="flex items-center space-x-1.5 text-xs font-bold">
                    <span>📅</span>
                    <span>Jam Spesifik</span>
                  </div>
                  <span class="text-3xs mt-1 block" :class="form.schedule_type === 'TIME_OF_DAY' ? 'text-indigo-100' : 'text-slate-400'">
                    Misal: Tepat jam 08:00 harian
                  </span>
                </button>

                <!-- Mode 3: INTERVAL -->
                <button
                  type="button"
                  @click="setScheduleType('INTERVAL')"
                  :class="form.schedule_type === 'INTERVAL' ? 'bg-indigo-600 text-white shadow-sm ring-2 ring-indigo-500/30 border-transparent' : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 hover:border-slate-300 border-slate-200 dark:border-slate-800'"
                  class="p-2.5 rounded-xl border text-left transition-all flex flex-col justify-between"
                >
                  <div class="flex items-center space-x-1.5 text-xs font-bold">
                    <span>🔄</span>
                    <span>Siklus Interval</span>
                  </div>
                  <span class="text-3xs mt-1 block" :class="form.schedule_type === 'INTERVAL' ? 'text-indigo-100' : 'text-slate-400'">
                    Misal: Berulang tiap 15m / 2 jam
                  </span>
                </button>

                <!-- Mode 4: CRON -->
                <button
                  type="button"
                  @click="setScheduleType('CRON')"
                  :class="form.schedule_type === 'CRON' ? 'bg-indigo-600 text-white shadow-sm ring-2 ring-indigo-500/30 border-transparent' : 'bg-white dark:bg-slate-900 text-slate-700 dark:text-slate-300 hover:border-slate-300 border-slate-200 dark:border-slate-800'"
                  class="p-2.5 rounded-xl border text-left transition-all flex flex-col justify-between"
                >
                  <div class="flex items-center space-x-1.5 text-xs font-bold">
                    <span>⚙️</span>
                    <span>Cron Kustom</span>
                  </div>
                  <span class="text-3xs mt-1 block" :class="form.schedule_type === 'CRON' ? 'text-indigo-100' : 'text-slate-400'">
                    Format cron ekspresi fleksibel
                  </span>
                </button>
              </div>
            </div>

            <!-- Dynamic Setting Box based on Selected Mode -->
            <!-- Mode 1 Panel: Menit Tiap Jam -->
            <div v-if="form.schedule_type === 'HOURLY_MINUTE'" class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-3 shadow-xs">
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-1">
                <label class="block text-2xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 font-mono">
                  MENIT KE- DALAM SETIAP JAM (0 - 59) <span class="text-rose-500">*</span>
                </label>
                <span class="text-3xs text-slate-400 font-mono">Bisa satu menit (45) atau beberapa (15, 45)</span>
              </div>
              <input
                type="text"
                v-model="form.minute_of_hour"
                placeholder="e.g. 45 atau 0, 30 atau 15, 45"
                required
                class="w-full px-3.5 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/80 text-slate-800 dark:text-slate-100 font-mono font-bold focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
              />
              <!-- Presets -->
              <div class="flex flex-wrap gap-1.5 items-center">
                <span class="text-3xs text-slate-400 font-medium">Pilihan Cepat:</span>
                <button
                  type="button"
                  v-for="p in minutePresets"
                  :key="p.label"
                  @click="form.minute_of_hour = p.val"
                  :class="form.minute_of_hour === p.val ? 'bg-indigo-600 text-white font-bold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700'"
                  class="px-2.5 py-1 rounded-md text-3xs font-mono transition-colors"
                >
                  {{ p.label }}
                </button>
              </div>
              <!-- Human Descriptive Note -->
              <div class="text-2xs text-indigo-700 dark:text-indigo-300 bg-indigo-50/80 dark:bg-indigo-950/40 p-2.5 rounded-lg border border-indigo-100 dark:border-indigo-900/40 flex items-start space-x-2">
                <span>💡</span>
                <span>
                  Aktuator akan menyala otomatis <strong>setiap jam tepat pada menit ke-{{ form.minute_of_hour || '45' }}</strong>
                  (misal: 01:{{ form.minute_of_hour || '45' }}, 02:{{ form.minute_of_hour || '45' }}, 03:{{ form.minute_of_hour || '45' }}, dst) selama {{ durationValue }} {{ durationUnit }}.
                </span>
              </div>
            </div>

            <!-- Mode 2 Panel: Jam Spesifik Harian -->
            <div v-if="form.schedule_type === 'TIME_OF_DAY'" class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-3 shadow-xs">
              <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-1">
                <label class="block text-2xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 font-mono">
                  WAKTU EKSEKUSI HARIAN (JAM:MENIT FORMAT 24-JAM) <span class="text-rose-500">*</span>
                </label>
                <span class="text-3xs text-slate-400 font-mono">Contoh: 08:00 atau 08:00, 16:30</span>
              </div>
              <input
                type="text"
                v-model="form.start_time_str"
                placeholder="e.g. 08:00"
                required
                class="w-full px-3.5 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/80 text-slate-800 dark:text-slate-100 font-mono font-bold focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
              />
              <div class="flex flex-wrap gap-1.5 items-center">
                <span class="text-3xs text-slate-400 font-medium">Pilihan Cepat:</span>
                <button
                  type="button"
                  v-for="p in timeOfDayPresets"
                  :key="p.label"
                  @click="form.start_time_str = p.val"
                  :class="form.start_time_str === p.val ? 'bg-indigo-600 text-white font-bold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700'"
                  class="px-2.5 py-1 rounded-md text-3xs font-mono transition-colors"
                >
                  {{ p.label }}
                </button>
              </div>
              <div class="text-2xs text-indigo-700 dark:text-indigo-300 bg-indigo-50/80 dark:bg-indigo-950/40 p-2.5 rounded-lg border border-indigo-100 dark:border-indigo-900/40 flex items-start space-x-2">
                <span>💡</span>
                <span>
                  Aktuator akan menyala otomatis tepat pukul <strong>{{ form.start_time_str || '08:00' }}</strong> pada setiap hari yang dipilih.
                </span>
              </div>
            </div>

            <!-- Mode 3 Panel: Interval Siklus -->
            <div v-if="form.schedule_type === 'INTERVAL'" class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-3 shadow-xs">
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 font-mono">
                INTERVAL PENGULANGAN (CYCLE) <span class="text-rose-500">*</span>
              </label>
              <div class="flex space-x-2.5">
                <input
                  type="number"
                  min="1"
                  v-model.number="intervalValue"
                  class="w-2/3 px-3.5 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/80 text-slate-800 dark:text-slate-100 font-mono font-bold focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                />
                <select
                  v-model="intervalUnit"
                  class="w-1/3 px-3 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/80 text-slate-800 dark:text-slate-100 font-medium focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                >
                  <option value="minutes">Menit (Minutes)</option>
                  <option value="hours">Jam (Hours)</option>
                  <option value="seconds">Detik (Seconds)</option>
                </select>
              </div>
              <div class="flex flex-wrap gap-1.5 items-center">
                <span class="text-3xs text-slate-400 font-medium">Pilihan Cepat:</span>
                <button
                  type="button"
                  v-for="p in intervalPresets"
                  :key="p.label"
                  @click="setIntervalPreset(p.val, p.unit)"
                  class="px-2.5 py-1 rounded-md text-3xs font-mono bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700 transition-colors"
                >
                  {{ p.label }}
                </button>
              </div>
            </div>

            <!-- Mode 4 Panel: Cron Kustom -->
            <div v-if="form.schedule_type === 'CRON'" class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-3 shadow-xs">
              <label class="block text-2xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 font-mono">
                EKSPRESI CRON (5-FIELD: MIN HOUR DOM MONTH DOW) <span class="text-rose-500">*</span>
              </label>
              <input
                type="text"
                v-model="form.cron_expression"
                placeholder="e.g. 45 * * * *"
                required
                class="w-full px-3.5 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50/50 dark:bg-slate-800/80 text-slate-800 dark:text-slate-100 font-mono font-bold focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
              />
              <div class="flex flex-wrap gap-1.5 items-center">
                <span class="text-3xs text-slate-400 font-medium">Pilihan Cepat:</span>
                <button
                  type="button"
                  v-for="p in cronPresets"
                  :key="p.label"
                  @click="form.cron_expression = p.val"
                  :class="form.cron_expression === p.val ? 'bg-indigo-600 text-white font-bold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 border border-slate-200 dark:border-slate-700'"
                  class="px-2.5 py-1 rounded-md text-3xs font-mono transition-colors"
                >
                  {{ p.label }}
                </button>
              </div>
            </div>

            <!-- 3. Active ON Duration (Lama Menyala) -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-1">
              <div>
                <div class="flex items-center justify-between mb-1">
                  <label class="block text-2xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-400 font-mono">
                    DURASI MENYALA (ACTIVE ON) <span class="text-rose-500">*</span>
                  </label>
                  <span class="text-3xs text-slate-400 font-mono">Otomatis mati setelah durasi ini</span>
                </div>
                <div class="flex space-x-2">
                  <input
                    type="number"
                    min="1"
                    v-model.number="durationValue"
                    class="w-2/3 px-3.5 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 font-mono font-bold focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                  />
                  <select
                    v-model="durationUnit"
                    class="w-1/3 px-2 py-2 text-xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                  >
                    <option value="minutes">Menit</option>
                    <option value="seconds">Detik</option>
                  </select>
                </div>
                <div class="flex flex-wrap gap-1.5 items-center pt-2">
                  <span class="text-3xs text-slate-400 font-medium">Pilihan Durasi:</span>
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

              <!-- 4. Days of Week -->
              <div>
                <div class="flex items-center justify-between mb-1">
                  <label class="block text-2xs font-bold uppercase tracking-wider text-slate-600 dark:text-slate-400 font-mono">
                    HARI AKTIF (DAYS OF WEEK)
                  </label>
                  <div class="flex space-x-1.5 text-3xs font-medium">
                    <button type="button" @click="setAllDays" class="text-indigo-600 dark:text-indigo-400 hover:underline">Semua Hari</button>
                    <span class="text-slate-300 dark:text-slate-700">&bull;</span>
                    <button type="button" @click="setWeekdaysOnly" class="text-indigo-600 dark:text-indigo-400 hover:underline">Sen-Jum</button>
                  </div>
                </div>
                <div class="flex flex-wrap gap-1.5 pt-1">
                  <button
                    type="button"
                    v-for="d in daysList"
                    :key="d.id"
                    @click="toggleDay(d.id)"
                    :class="isDayActive(d.id) ? 'bg-indigo-600 text-white font-bold shadow-xs' : 'bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700'"
                    class="px-3 py-1.5 text-2xs rounded-lg transition-colors"
                  >
                    {{ d.name }}
                  </button>
                </div>
              </div>
            </div>

            <!-- Form Actions -->
            <div class="flex items-center justify-end space-x-2.5 pt-3 border-t border-slate-200/80 dark:border-slate-800">
              <button
                type="button"
                @click="cancelForm"
                class="px-4 py-2 text-xs font-semibold rounded-lg text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
              >
                Batal
              </button>
              <button
                type="button"
                @click="saveSchedule"
                :disabled="saving"
                class="px-5 py-2 text-xs font-bold rounded-lg bg-indigo-600 hover:bg-indigo-500 active:bg-indigo-700 text-white shadow-sm transition-all flex items-center space-x-1.5 disabled:opacity-50"
              >
                <span v-if="saving" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                <span>{{ editingScheduleId ? 'Simpan Perubahan' : 'Buat Jadwal' }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- TABLE SECTION: Table View Kapan Aja Eksekusi & Bisa Disortir (Gambar 2) -->
        <div class="space-y-3.5">
          <!-- Table Controls Toolbar -->
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <!-- Tabs: Jadwal Rutinitas vs Riwayat Eksekusi Log -->
            <div class="flex items-center space-x-1 p-1 bg-slate-100 dark:bg-slate-800/80 rounded-xl border border-slate-200 dark:border-slate-700/80 self-start">
              <button
                type="button"
                @click="activeViewTab = 'schedules'"
                :class="activeViewTab === 'schedules' ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 font-bold shadow-xs' : 'text-slate-600 dark:text-slate-400 font-medium hover:text-slate-900'"
                class="px-3.5 py-1.5 text-xs rounded-lg transition-all flex items-center space-x-1.5"
              >
                <span>⏱️</span>
                <span>Tabel Jadwal Rutinitas</span>
                <span class="px-1.5 py-0.2 rounded-full text-3xs font-mono font-bold bg-indigo-50 text-indigo-700 dark:bg-indigo-500/20 dark:text-indigo-300">
                  {{ schedules.length }}
                </span>
              </button>

              <button
                type="button"
                @click="switchToLogsTab"
                :class="activeViewTab === 'logs' ? 'bg-white dark:bg-slate-900 text-indigo-600 dark:text-indigo-400 font-bold shadow-xs' : 'text-slate-600 dark:text-slate-400 font-medium hover:text-slate-900'"
                class="px-3.5 py-1.5 text-xs rounded-lg transition-all flex items-center space-x-1.5"
              >
                <span>📜</span>
                <span>Riwayat Eksekusi (History)</span>
                <span v-if="logs.length" class="px-1.5 py-0.2 rounded-full text-3xs font-mono font-bold bg-slate-200 text-slate-700 dark:bg-slate-700 dark:text-slate-300">
                  {{ logs.length }}
                </span>
              </button>
            </div>

            <!-- Search & Filter Options -->
            <div class="flex items-center space-x-2 text-xs">
              <div class="relative w-48 sm:w-56">
                <input
                  type="text"
                  v-model="searchQuery"
                  placeholder="Cari jadwal / waktu..."
                  class="w-full pl-8 pr-3 py-1.5 text-2xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-indigo-500/30"
                />
                <span class="absolute left-2.5 top-2 text-slate-400 text-2xs">🔍</span>
              </div>

              <select
                v-if="activeViewTab === 'schedules'"
                v-model="statusFilter"
                class="px-2.5 py-1.5 text-2xs rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-100 focus:outline-none"
              >
                <option value="ALL">Semua Status</option>
                <option value="ACTIVE">Aktif Saja</option>
                <option value="PAUSED">Jeda Saja</option>
              </select>

              <button
                v-if="activeViewTab === 'logs'"
                type="button"
                @click="fetchLogs"
                class="p-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-600 dark:text-slate-300 hover:bg-slate-50 transition-colors"
                title="Refresh Riwayat Eksekusi"
              >
                🔄
              </button>
            </div>
          </div>

          <!-- TAB 1: TABEL JADWAL RUTINITAS (BISA SORTIR & LIHAT KAPAN AJA) -->
          <div v-if="activeViewTab === 'schedules'" class="overflow-x-auto border border-slate-200 dark:border-slate-800 rounded-xl bg-white dark:bg-slate-900 shadow-sm">
            <div v-if="loadingSchedules" class="py-12 text-center text-xs text-slate-400">
              Memuat data jadwal...
            </div>

            <div v-else-if="sortedSchedules.length === 0" class="py-12 text-center text-slate-400 dark:text-slate-500 text-xs space-y-1">
              <p class="text-2xl mb-1">⏱️</p>
              <p class="font-medium text-slate-600 dark:text-slate-300">Belum ada jadwal yang cocok.</p>
              <p class="text-3xs">Klik tombol "+ Tambah Jadwal Baru" di atas untuk menambahkan jadwal otomatis.</p>
            </div>

            <table v-else class="w-full text-left text-xs font-sans">
              <thead class="bg-slate-50/80 dark:bg-slate-800/60 text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800 text-3xs uppercase font-mono tracking-wider select-none">
                <tr>
                  <th @click="changeSort('schedule_name')" class="py-3 px-3.5 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Nama Rutinitas</span>
                      <span class="text-2xs">{{ getSortIcon('schedule_name') }}</span>
                    </div>
                  </th>
                  <th @click="changeSort('schedule_type')" class="py-3 px-3 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Mode</span>
                      <span class="text-2xs">{{ getSortIcon('schedule_type') }}</span>
                    </div>
                  </th>
                  <th class="py-3 px-3.5">
                    <span>Kapan Dieksekusi</span>
                  </th>
                  <th @click="changeSort('duration_seconds')" class="py-3 px-3 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Durasi ON</span>
                      <span class="text-2xs">{{ getSortIcon('duration_seconds') }}</span>
                    </div>
                  </th>
                  <th class="py-3 px-3">
                    <span>Hari Aktif</span>
                  </th>
                  <th @click="changeSort('next_run_at')" class="py-3 px-3.5 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Next Run (Eksekusi Berikutnya)</span>
                      <span class="text-2xs text-indigo-600 dark:text-indigo-400 font-bold">{{ getSortIcon('next_run_at') }}</span>
                    </div>
                  </th>
                  <th @click="changeSort('last_status')" class="py-3 px-3 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Last Status</span>
                      <span class="text-2xs">{{ getSortIcon('last_status') }}</span>
                    </div>
                  </th>
                  <th @click="changeSort('enabled')" class="py-3 px-3 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Status</span>
                      <span class="text-2xs">{{ getSortIcon('enabled') }}</span>
                    </div>
                  </th>
                  <th class="py-3 px-3 text-right">
                    <span>Aksi</span>
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800 text-2xs">
                <tr
                  v-for="s in sortedSchedules"
                  :key="s.id"
                  class="hover:bg-indigo-50/30 dark:hover:bg-slate-800/40 transition-colors"
                >
                  <!-- Routine Name -->
                  <td class="py-2.5 px-3.5 font-bold text-slate-800 dark:text-slate-100 whitespace-nowrap">
                    {{ s.schedule_name }}
                  </td>

                  <!-- Mode Badge -->
                  <td class="py-2.5 px-3 whitespace-nowrap">
                    <span
                      class="px-2 py-0.5 rounded text-3xs font-mono font-bold"
                      :class="getModeBadgeClass(s.schedule_type)"
                    >
                      {{ getModeBadgeLabel(s) }}
                    </span>
                  </td>

                  <!-- Kapan Dieksekusi -->
                  <td class="py-2.5 px-3.5 font-mono text-slate-700 dark:text-slate-300 font-semibold whitespace-nowrap">
                    {{ getTimingSummary(s) }}
                  </td>

                  <!-- Durasi ON -->
                  <td class="py-2.5 px-3 font-mono text-slate-600 dark:text-slate-300 whitespace-nowrap">
                    {{ formatDuration(s.duration_seconds) }}
                  </td>

                  <!-- Hari Aktif -->
                  <td class="py-2.5 px-3 font-mono text-3xs text-slate-500 whitespace-nowrap">
                    {{ formatDaysSummary(s.days_of_week) }}
                  </td>

                  <!-- Next Run Time (Highlight) -->
                  <td class="py-2.5 px-3.5 font-mono whitespace-nowrap">
                    <div class="flex items-center space-x-1.5 font-bold text-indigo-600 dark:text-indigo-400">
                      <span>⏰</span>
                      <span>{{ formatFullDateTime(s.next_run_at) }}</span>
                    </div>
                  </td>

                  <!-- Last Status -->
                  <td class="py-2.5 px-3 font-mono text-3xs whitespace-nowrap">
                    <span
                      v-if="s.last_status"
                      :class="s.last_status === 'SUCCESS' ? 'text-emerald-600 font-bold' : (s.last_status === 'FAILED' ? 'text-rose-600 font-bold' : 'text-slate-400')"
                    >
                      {{ s.last_status }}
                    </span>
                    <span v-else class="text-slate-400">--</span>
                  </td>

                  <!-- Status Toggle -->
                  <td class="py-2.5 px-3 whitespace-nowrap">
                    <button
                      type="button"
                      @click="toggleScheduleActive(s)"
                      :class="s.enabled ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30' : 'bg-slate-100 text-slate-500 dark:bg-slate-800 dark:text-slate-400 border-slate-200 dark:border-slate-700'"
                      class="px-2.5 py-0.5 rounded-full text-3xs font-mono font-bold border transition-colors hover:opacity-80"
                      title="Klik untuk mengubah status aktif/jeda"
                    >
                      {{ s.enabled ? '● AKTIF' : '○ JEDA' }}
                    </button>
                  </td>

                  <!-- Action Buttons -->
                  <td class="py-2.5 px-3 text-right whitespace-nowrap">
                    <div class="flex items-center justify-end space-x-1">
                      <button
                        type="button"
                        @click="editSchedule(s)"
                        class="p-1.5 text-slate-400 hover:text-indigo-600 dark:hover:text-indigo-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                        title="Edit Rutinitas"
                      >
                        ✏️
                      </button>
                      <button
                        type="button"
                        @click="deleteSchedule(s.id)"
                        class="p-1.5 text-slate-400 hover:text-rose-600 dark:hover:text-rose-400 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                        title="Hapus Rutinitas"
                      >
                        🗑️
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- TAB 2: TABEL RIWAYAT EKSEKUSI (BISA SORTIR KAPAN AJA EKSEKUSI BERJALAN) -->
          <div v-if="activeViewTab === 'logs'" class="overflow-x-auto border border-slate-200 dark:border-slate-800 rounded-xl bg-white dark:bg-slate-900 shadow-sm">
            <div v-if="loadingLogs" class="py-12 text-center text-xs text-slate-400">
              Memuat riwayat eksekusi...
            </div>

            <div v-else-if="sortedLogs.length === 0" class="py-12 text-center text-slate-400 dark:text-slate-500 text-xs">
              <p class="text-2xl mb-1">📜</p>
              <p class="font-medium text-slate-600 dark:text-slate-300">Belum ada catatan riwayat eksekusi untuk aktuator ini.</p>
            </div>

            <table v-else class="w-full text-left text-xs font-sans">
              <thead class="bg-slate-50/80 dark:bg-slate-800/60 text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800 text-3xs uppercase font-mono tracking-wider select-none">
                <tr>
                  <th @click="changeLogSort('executed_at')" class="py-3 px-3.5 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Waktu Eksekusi (Executed At)</span>
                      <span class="text-2xs text-indigo-600 font-bold">{{ getLogSortIcon('executed_at') }}</span>
                    </div>
                  </th>
                  <th @click="changeLogSort('action')" class="py-3 px-3 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Perintah (Action)</span>
                      <span class="text-2xs">{{ getLogSortIcon('action') }}</span>
                    </div>
                  </th>
                  <th @click="changeLogSort('triggered_by')" class="py-3 px-3 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Pemicu (Triggered By)</span>
                      <span class="text-2xs">{{ getLogSortIcon('triggered_by') }}</span>
                    </div>
                  </th>
                  <th class="py-3 px-3 font-mono">Modbus PDU</th>
                  <th class="py-3 px-3 font-mono">Nilai Ditulis</th>
                  <th @click="changeLogSort('latency_ms')" class="py-3 px-3 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Latency</span>
                      <span class="text-2xs">{{ getLogSortIcon('latency_ms') }}</span>
                    </div>
                  </th>
                  <th @click="changeLogSort('status')" class="py-3 px-3 cursor-pointer hover:text-indigo-600 transition-colors">
                    <div class="flex items-center space-x-1">
                      <span>Status</span>
                      <span class="text-2xs">{{ getLogSortIcon('status') }}</span>
                    </div>
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800 text-2xs font-mono">
                <tr
                  v-for="log in sortedLogs"
                  :key="log.id"
                  class="hover:bg-slate-50/50 dark:hover:bg-slate-800/40 transition-colors"
                >
                  <td class="py-2.5 px-3.5 font-bold text-slate-700 dark:text-slate-300 whitespace-nowrap">
                    {{ formatFullDateTime(log.executed_at) }}
                  </td>
                  <td class="py-2.5 px-3 whitespace-nowrap">
                    <span
                      :class="log.action === 'ON' ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30' : 'bg-rose-50 text-rose-700 dark:bg-rose-500/10 dark:text-rose-400 border-rose-200 dark:border-rose-500/30'"
                      class="px-2 py-0.5 rounded-full text-3xs font-bold border"
                    >
                      {{ log.action }}
                    </span>
                  </td>
                  <td class="py-2.5 px-3 font-sans font-medium text-slate-800 dark:text-slate-200 whitespace-nowrap">
                    {{ log.triggered_by }}
                  </td>
                  <td class="py-2.5 px-3 text-slate-500 whitespace-nowrap">
                    FC0{{ log.function_code }} @ #{{ log.address }} (Slave #{{ log.slave_id }})
                  </td>
                  <td class="py-2.5 px-3 text-slate-700 dark:text-slate-300 whitespace-nowrap">
                    {{ formatHex(log.payload_written) }} ({{ log.payload_written }})
                  </td>
                  <td class="py-2.5 px-3 text-amber-600 dark:text-amber-400 whitespace-nowrap font-bold">
                    {{ log.latency_ms }} ms
                  </td>
                  <td class="py-2.5 px-3 whitespace-nowrap">
                    <span
                      :class="log.status === 'SUCCESS' ? 'text-emerald-600 dark:text-emerald-400 font-bold' : 'text-rose-600 dark:text-rose-400 font-bold'"
                      class="text-3xs"
                    >
                      {{ log.status }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-3 border-t border-slate-100 dark:border-slate-800 flex items-center justify-between bg-slate-50/60 dark:bg-slate-900/60">
        <span class="text-3xs text-slate-400 font-mono">
          Sinkronisasi otomatis &bull; Ticker interval 5 detik
        </span>
        <button
          type="button"
          @click="close"
          class="px-5 py-2 text-xs font-semibold rounded-lg bg-slate-200 hover:bg-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors"
        >
          Tutup
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
      logs: [],
      loadingSchedules: false,
      loadingLogs: false,
      showForm: false,
      saving: false,
      editingScheduleId: null,
      activeViewTab: 'schedules', // 'schedules' or 'logs'
      searchQuery: '',
      statusFilter: 'ALL',

      // Sorting for Schedules
      sortBy: 'next_run_at',
      sortDir: 'asc',

      // Sorting for Logs
      logSortBy: 'executed_at',
      logSortDir: 'desc',

      intervalValue: 2,
      intervalUnit: 'hours',
      durationValue: 5,
      durationUnit: 'minutes',
      selectedDays: ['1', '2', '3', '4', '5', '6', '7'],
      daysList: [
        { id: '1', name: 'Sen' },
        { id: '2', name: 'Sel' },
        { id: '3', name: 'Rab' },
        { id: '4', name: 'Kam' },
        { id: '5', name: 'Jum' },
        { id: '6', name: 'Sab' },
        { id: '7', name: 'Min' }
      ],
      minutePresets: [
        { label: ':00 (Awal Jam)', val: '0' },
        { label: ':15', val: '15' },
        { label: ':30 (Tengah Jam)', val: '30' },
        { label: ':45', val: '45' },
        { label: ':55 (Sebelum Ganti Jam)', val: '55' },
        { label: ':15, :45 (2x/Jam)', val: '15, 45' }
      ],
      timeOfDayPresets: [
        { label: '06:00 (Pagi)', val: '06:00' },
        { label: '08:00 (Awal Shift)', val: '08:00' },
        { label: '12:00 (Siang)', val: '12:00' },
        { label: '17:00 (Sore)', val: '17:00' },
        { label: '08:00, 16:30 (2x/Hari)', val: '08:00, 16:30' }
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
  computed: {
    sortedSchedules() {
      let list = [...this.schedules];

      // Filter search
      if (this.searchQuery && this.searchQuery.trim()) {
        const q = this.searchQuery.toLowerCase().trim();
        list = list.filter(s =>
          (s.schedule_name && s.schedule_name.toLowerCase().includes(q)) ||
          (s.schedule_type && s.schedule_type.toLowerCase().includes(q)) ||
          (s.minute_of_hour && s.minute_of_hour.includes(q)) ||
          (s.start_time_str && s.start_time_str.includes(q))
        );
      }

      // Filter status
      if (this.statusFilter !== 'ALL') {
        const want = this.statusFilter === 'ACTIVE';
        list = list.filter(s => s.enabled === want);
      }

      // Sort
      list.sort((a, b) => {
        let vA = a[this.sortBy];
        let vB = b[this.sortBy];

        if (this.sortBy === 'next_run_at' || this.sortBy === 'last_run_at') {
          vA = vA ? new Date(vA).getTime() : (this.sortDir === 'asc' ? Infinity : -Infinity);
          vB = vB ? new Date(vB).getTime() : (this.sortDir === 'asc' ? Infinity : -Infinity);
        } else if (typeof vA === 'string') {
          vA = vA.toLowerCase();
          vB = (vB || '').toLowerCase();
        }

        if (vA < vB) return this.sortDir === 'asc' ? -1 : 1;
        if (vA > vB) return this.sortDir === 'asc' ? 1 : -1;
        return 0;
      });

      return list;
    },
    sortedLogs() {
      let list = [...this.logs];

      if (this.searchQuery && this.searchQuery.trim()) {
        const q = this.searchQuery.toLowerCase().trim();
        list = list.filter(l =>
          (l.triggered_by && l.triggered_by.toLowerCase().includes(q)) ||
          (l.action && l.action.toLowerCase().includes(q)) ||
          (l.status && l.status.toLowerCase().includes(q))
        );
      }

      list.sort((a, b) => {
        let vA = a[this.logSortBy];
        let vB = b[this.logSortBy];

        if (this.logSortBy === 'executed_at') {
          vA = vA ? new Date(vA).getTime() : 0;
          vB = vB ? new Date(vB).getTime() : 0;
        } else if (typeof vA === 'string') {
          vA = vA.toLowerCase();
          vB = (vB || '').toLowerCase();
        }

        if (vA < vB) return this.logSortDir === 'asc' ? -1 : 1;
        if (vA > vB) return this.logSortDir === 'asc' ? 1 : -1;
        return 0;
      });

      return list;
    }
  },
  watch: {
    isOpen(newVal) {
      if (newVal && this.actuator) {
        this.fetchSchedules();
        this.fetchLogs();
      }
    }
  },
  methods: {
    close() {
      this.cancelForm();
      this.$emit('close');
    },
    switchToLogsTab() {
      this.activeViewTab = 'logs';
      if (!this.logs.length) {
        this.fetchLogs();
      }
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
    async fetchLogs() {
      if (!this.actuator) return;
      this.loadingLogs = true;
      try {
        const res = await axios.get('/api/actuators/logs', {
          params: { actuator_id: this.actuator.id, limit: 100 }
        });
        this.logs = res.data.data || [];
      } catch (err) {
        console.error('Failed to fetch actuator logs:', err);
      } finally {
        this.loadingLogs = false;
      }
    },
    changeSort(col) {
      if (this.sortBy === col) {
        this.sortDir = this.sortDir === 'asc' ? 'desc' : 'asc';
      } else {
        this.sortBy = col;
        this.sortDir = 'asc';
      }
    },
    getSortIcon(col) {
      if (this.sortBy !== col) return '↕';
      return this.sortDir === 'asc' ? '▲' : '▼';
    },
    changeLogSort(col) {
      if (this.logSortBy === col) {
        this.logSortDir = this.logSortDir === 'asc' ? 'desc' : 'asc';
      } else {
        this.logSortBy = col;
        this.logSortDir = 'desc';
      }
    },
    getLogSortIcon(col) {
      if (this.logSortBy !== col) return '↕';
      return this.logSortDir === 'asc' ? '▲' : '▼';
    },
    setScheduleType(type) {
      this.form.schedule_type = type;
      if (!this.form.schedule_name || this.form.schedule_name.startsWith('Sampling') || this.form.schedule_name.startsWith('Jadwal') || this.form.schedule_name.startsWith('Cycle')) {
        if (type === 'HOURLY_MINUTE') {
          this.form.schedule_name = `Sampling Menit ${this.form.minute_of_hour || '45'}`;
        } else if (type === 'TIME_OF_DAY') {
          this.form.schedule_name = `Jadwal Harian Pukul ${this.form.start_time_str || '08:00'}`;
        } else if (type === 'INTERVAL') {
          this.form.schedule_name = `Siklus ${this.intervalValue}-${this.intervalUnit}`;
        } else if (type === 'CRON') {
          this.form.schedule_name = `Jadwal Cron Kustom`;
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
        schedule_name: 'Sampling Menit 45',
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
      if (sec >= 3600) return `${(sec / 3600).toFixed(1)} jam`;
      if (sec >= 60) return `${Math.round(sec / 60)} menit`;
      return `${sec} detik`;
    },
    formatFullDateTime(dStr) {
      if (!dStr) return '--';
      const d = new Date(dStr);
      return d.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    formatHex(val) {
      if (val === undefined || val === null) return '0x0000';
      return '0x' + Number(val).toString(16).toUpperCase().padStart(4, '0');
    },
    formatDaysSummary(daysStr) {
      if (!daysStr || daysStr === '1,2,3,4,5,6,7') return 'Tiap Hari';
      if (daysStr === '1,2,3,4,5') return 'Sen-Jum';
      if (daysStr === '6,7') return 'Sab-Min';
      return daysStr.split(',').length + ' hari';
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
          return `Tiap jam m:${s.minute_of_hour || '0'}`;
        case 'TIME_OF_DAY':
          return `Pukul ${s.start_time_str || '08:00'}`;
        case 'CRON':
          return `Cron: ${s.cron_expression || '45 * * * *'}`;
        case 'INTERVAL':
        default:
          return `Tiap ${this.formatDuration(s.interval_seconds)}`;
      }
    },
    async toggleScheduleActive(s) {
      try {
        await axios.put(`/api/actuators/schedules/${s.id}`, {
          ...s,
          enabled: !s.enabled
        });
        await this.fetchSchedules();
        this.$emit('updated');
      } catch (err) {
        const errData = err.response && err.response.data;
        alert('Gagal mengubah status jadwal: ' + ((errData && (errData.error || errData.message)) || err.message));
      }
    },
    async saveSchedule() {
      if (!this.form.schedule_name || !this.form.schedule_name.trim()) {
        alert('Nama jadwal/rutinitas wajib diisi');
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
        alert('Gagal menyimpan jadwal: ' + ((errData && (errData.error || errData.message)) || err.message));
      } finally {
        this.saving = false;
      }
    },
    async deleteSchedule(id) {
      if (!confirm('Apakah Anda yakin ingin menghapus jadwal otomatis ini?')) return;
      try {
        await axios.delete(`/api/actuators/schedules/${id}`);
        await this.fetchSchedules();
        this.$emit('updated');
      } catch (err) {
        const errData = err.response && err.response.data;
        alert('Gagal menghapus jadwal: ' + ((errData && (errData.error || errData.message)) || err.message));
      }
    }
  }
};
</script>