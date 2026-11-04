<template>
  <div v-if="isSuperAdmin" class="p-4 md:p-6 space-y-5 max-w-[1600px] mx-auto font-sans">
    <!-- Top Header -->
    <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
      <div>
        <div class="flex items-center space-x-2.5">
          <h1 class="text-lg md:text-xl font-black text-slate-900 dark:text-white tracking-tight">
            Penyuntikan Telemetri &amp; Agregasi Otomatis
          </h1>
          <span class="px-2 py-0.5 rounded-md text-3xs font-mono font-bold uppercase tracking-wider bg-indigo-500/15 text-indigo-600 dark:text-indigo-300 border border-indigo-500/30">
            SUPERADMIN ONLY
          </span>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          Isi celah data kosong/<code class="font-mono">null</code> dengan rentang nilai realistis (Raw, Hex Modbus, Processed), hitung ulang seluruh bucket agregasi, dan perbarui payload antrean Outbound secara otomatis.
        </p>
      </div>

      <div class="flex items-center space-x-2">
        <router-link
          :to="selectedDeviceId ? ('/monitoring/analysis?device_id=' + selectedDeviceId) : '/monitoring/analysis'"
          class="px-3 py-1.5 rounded-xl bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <span>📊 Buka Telemetry Analysis</span>
        </router-link>
        <router-link
          to="/integration/queue"
          class="px-3 py-1.5 rounded-xl bg-white dark:bg-slate-800 hover:bg-slate-50 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <span>📤 Cek Delivery Queue</span>
        </router-link>
      </div>
    </div>

    <!-- Stealth & Zero-Leak Assurance Banner -->
    <div class="p-3.5 rounded-2xl bg-indigo-50/80 dark:bg-indigo-950/30 border border-indigo-200 dark:border-indigo-800/60 flex items-start space-x-3 text-xs">
      <div class="w-7 h-7 rounded-xl bg-indigo-500/15 border border-indigo-500/30 flex items-center justify-center flex-shrink-0 mt-0.5">
        <span class="text-indigo-600 dark:text-indigo-300 font-mono font-bold text-2xs">SA</span>
      </div>
      <div class="space-y-0.5 text-slate-700 dark:text-slate-300 leading-relaxed">
        <div class="font-bold text-indigo-900 dark:text-indigo-200">
          Mode Penanda Khusus Superadmin (Zero-Leak Protection Aktif)
        </div>
        <div class="text-2xs text-slate-600 dark:text-slate-400">
          Data yang disuntikkan menyimpan <code class="font-mono">RawValue</code>, frame register <code class="font-mono">RawHex</code> sesuai tipe data/endianness sensor, <code class="font-mono">ProcessedValue</code>, dan <code class="font-mono">Quality=GOOD</code> tanpa meninggalkan jejak di Audit Trail maupun System Log umum. Peran <strong>Admin, Engineer, Operator</strong>, Laporan PDF/CSV, serta payload <strong>KLHK SPARING / ONLIMO / ISPU</strong> melihatnya sebagai telemetri normal — sementara akun <strong>Superadmin</strong> tetap melihat lencana <span class="px-1.5 py-0.2 rounded bg-violet-500/15 text-violet-600 dark:text-violet-300 border border-violet-500/30 font-mono font-bold text-4xs">INJ</span> dan dapat melakukan <strong>Rollback 1-Klik</strong> kapan saja.
        </div>
      </div>
    </div>

    <!-- Toast / Feedback Banner -->
    <div
      v-if="feedbackMessage"
      class="p-3.5 rounded-xl border flex items-center justify-between text-xs font-semibold"
      :class="feedbackType === 'error'
        ? 'bg-rose-50 dark:bg-rose-950/50 border-rose-300 dark:border-rose-800 text-rose-700 dark:text-rose-200'
        : 'bg-emerald-50 dark:bg-emerald-950/50 border-emerald-300 dark:border-emerald-800 text-emerald-700 dark:text-emerald-200'"
    >
      <span>{{ feedbackMessage }}</span>
      <button type="button" @click="feedbackMessage = ''" class="ml-3 text-slate-400 hover:text-white font-mono">✕</button>
    </div>

    <!-- STEP 1: Stasiun, Rentang Waktu & Konfigurasi Generator -->
    <div class="saas-card p-4 md:p-5 rounded-2xl bg-white dark:bg-[#0F172A]/90 border border-slate-200 dark:border-slate-800 space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-200 dark:border-slate-800 pb-3">
        <div class="flex items-center space-x-2">
          <span class="w-5 h-5 rounded-full bg-blue-600 text-white text-3xs font-bold flex items-center justify-center">1</span>
          <h2 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white">
            Target Device, Rentang Waktu &amp; Strategi Generator
          </h2>
        </div>
        <div v-if="deviceHints" class="flex flex-wrap items-center gap-1.5 text-3xs font-mono">
          <span class="px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
            Protokol: {{ deviceHints.protocol }}
          </span>
          <span class="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-300 border border-emerald-500/30">
            Agregasi Aktif: {{ deviceHints.active_agg_intervals && deviceHints.active_agg_intervals.length ? deviceHints.active_agg_intervals.map(formatIntervalLabel).join(', ') : 'Belum Ada' }}
          </span>
          <button
            type="button"
            @click="showUnixHelper = !showUnixHelper"
            class="px-2 py-0.5 rounded bg-amber-500/10 hover:bg-amber-500/20 text-amber-600 dark:text-amber-300 border border-amber-500/30 transition cursor-pointer"
          >
            🕒 {{ showUnixHelper ? 'Tutup Konverter Epoch' : 'Bantu dari Unix Epoch SPARING' }}
          </button>
        </div>
      </div>

      <!-- Optional SPARING Unix Epoch Helper -->
      <div v-if="showUnixHelper" class="p-3 rounded-xl bg-amber-50/70 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-800/50 flex flex-wrap items-center gap-3 text-xs">
        <span class="text-2xs font-semibold text-amber-800 dark:text-amber-200">
          Tempel <code class="font-mono">datetime</code> Unix Epoch dari JSON SPARING (misal <code class="font-mono">1793777880</code>):
        </span>
        <input
          type="number"
          v-model="unixHelperInput"
          placeholder="Contoh: 1793777880"
          class="px-2.5 py-1 rounded-lg bg-white dark:bg-[#0B0F19] border border-amber-300 dark:border-amber-700 text-xs font-mono text-slate-900 dark:text-white w-44 focus:outline-none"
        />
        <button
          type="button"
          @click="applyUnixEpochExactBucket"
          class="px-3 py-1 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-2xs font-bold transition cursor-pointer"
        >
          Setel 30 Bucket mulai Epoch Pertama Ini (-2m s/d +58m)
        </button>
        <button
          type="button"
          @click="applyUnixEpochToHourWindow"
          class="px-3 py-1 rounded-lg bg-slate-200 dark:bg-slate-800 hover:bg-slate-300 dark:hover:bg-slate-700 text-slate-800 dark:text-slate-200 text-2xs font-semibold transition cursor-pointer"
        >
          Setel Jam Kalender (:00 s/d :00)
        </button>
      </div>

      <!-- Main Form Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-3.5">
        <!-- Device Selector -->
        <div class="space-y-1">
          <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
            Device / Stasiun Sensor
          </label>
          <select
            v-model="selectedDeviceId"
            @change="onDeviceChange"
            class="w-full px-3 py-2 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs text-slate-900 dark:text-white font-semibold focus:outline-none focus:border-blue-500"
          >
            <option value="" disabled>-- Pilih Device --</option>
            <option v-for="d in devices" :key="d.id" :value="d.id">
              {{ d.device_code }} — {{ d.device_name }} {{ d.enabled === false ? '(Disabled)' : '' }}
            </option>
          </select>
        </div>

        <!-- Start Time -->
        <div class="space-y-1">
          <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
            Waktu Mulai (Lokal)
          </label>
          <input
            type="datetime-local"
            step="1"
            v-model="form.startTime"
            class="w-full px-3 py-1.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs font-mono text-slate-900 dark:text-white focus:outline-none focus:border-blue-500"
          />
        </div>

        <!-- End Time -->
        <div class="space-y-1">
          <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
            Waktu Selesai (Lokal)
          </label>
          <input
            type="datetime-local"
            step="1"
            v-model="form.endTime"
            class="w-full px-3 py-1.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs font-mono text-slate-900 dark:text-white focus:outline-none focus:border-blue-500"
          />
        </div>

        <!-- Sample Step Seconds -->
        <div class="space-y-1">
          <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
            Interval Sampel Raw
          </label>
          <select
            v-model.number="form.sampleStepSeconds"
            class="w-full px-3 py-2 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
          >
            <option :value="10">Setiap 10 detik (6 sampel / menit)</option>
            <option :value="15">Setiap 15 detik (4 sampel / menit)</option>
            <option :value="30">Setiap 30 detik (4 sampel / bucket 2m — Rekomendasi)</option>
            <option :value="60">Setiap 60 detik (1 sampel / menit)</option>
            <option :value="120">Setiap 120 detik (1 sampel / 2 menit)</option>
            <option :value="300">Setiap 300 detik (1 sampel / 5 menit)</option>
          </select>
        </div>
      </div>

      <!-- Quick Time Window Buttons -->
      <div class="flex flex-wrap items-center gap-2 pt-0.5">
        <span class="text-3xs font-bold uppercase tracking-wider text-slate-400 mr-1">Preset Cepat:</span>
        <button type="button" @click="applyWindowPreset('prev_hour')" class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-blue-600 hover:text-white text-slate-700 dark:text-slate-300 text-3xs font-semibold transition cursor-pointer">
          Jam Penuh Sebelumnya (30 Bucket SPARING)
        </button>
        <button type="button" @click="applyWindowPreset('curr_hour')" class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-blue-600 hover:text-white text-slate-700 dark:text-slate-300 text-3xs font-semibold transition cursor-pointer">
          Jam Berjalan (:00 s/d Sekarang)
        </button>
        <button type="button" @click="applyWindowPreset('last_1h')" class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-blue-600 hover:text-white text-slate-700 dark:text-slate-300 text-3xs font-semibold transition cursor-pointer">
          1 Jam Terakhir
        </button>
        <button type="button" @click="applyWindowPreset('last_3h')" class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-blue-600 hover:text-white text-slate-700 dark:text-slate-300 text-3xs font-semibold transition cursor-pointer">
          3 Jam Terakhir
        </button>
        <button type="button" @click="applyWindowPreset('last_6h')" class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-blue-600 hover:text-white text-slate-700 dark:text-slate-300 text-3xs font-semibold transition cursor-pointer">
          6 Jam Terakhir
        </button>
        <button type="button" @click="applyWindowPreset('today')" class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-blue-600 hover:text-white text-slate-700 dark:text-slate-300 text-3xs font-semibold transition cursor-pointer">
          Hari Ini (00:00 – Sekarang)
        </button>
      </div>

      <!-- Strategy Row -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3.5 pt-2 border-t border-slate-100 dark:border-slate-800/80">
        <!-- Conflict Mode -->
        <div class="space-y-1">
          <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
            Mode Penanganan Data Eksisting
          </label>
          <select
            v-model="form.conflictMode"
            class="w-full px-3 py-2 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500"
          >
            <option value="FILL_MISSING_AND_NULL">Hanya Isi Kosong / Null / Error (Data Valid Aman)</option>
            <option value="OVERWRITE_ALL">Timpa Semua Data pada Rentang Waktu Ini (Overwrite)</option>
          </select>
        </div>

        <!-- Random Mode -->
        <div class="space-y-1">
          <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
            Distribusi Nilai di Dalam Range
          </label>
          <select
            v-model="form.randomMode"
            class="w-full px-3 py-2 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500"
          >
            <option value="SMOOTH_WALK">Natural Smooth Walk (Tren Halus seperti Sensor Asli)</option>
            <option value="UNIFORM">Uniform Random (Acak Bebas Merata Min–Max)</option>
          </select>
        </div>

        <!-- Value Target -->
        <div class="space-y-1">
          <label class="block text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
            Basis Angka Min–Max yang Diinput
          </label>
          <select
            v-model="form.valueTarget"
            class="w-full px-3 py-2 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500"
          >
            <option value="PROCESSED">Processed / Engineering Value (Hitung Balik Raw Otomatis)</option>
            <option value="RAW">Raw Register Value (Hitung Scale, Offset &amp; Formula Otomatis)</option>
          </select>
        </div>
      </div>

      <!-- Checkboxes & Internal Note -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 items-center pt-1">
        <div class="lg:col-span-7 flex flex-wrap items-center gap-4 text-xs">
          <label class="inline-flex items-center space-x-2 cursor-pointer select-none">
            <input type="checkbox" v-model="form.refreshOutbound" class="rounded border-slate-400 text-blue-600 focus:ring-blue-500" />
            <span class="text-2xs text-slate-700 dark:text-slate-300 font-medium">
              Perbarui payload Outbound <code class="font-mono">PENDING/RETRYING/FAILED</code> yang tumpang tindih
            </span>
          </label>
          <label class="inline-flex items-center space-x-2 cursor-pointer select-none">
            <input type="checkbox" v-model="form.enqueueMissingOutbound" class="rounded border-slate-400 text-blue-600 focus:ring-blue-500" />
            <span class="text-2xs text-slate-700 dark:text-slate-300 font-medium">
              Buat antrean Outbound baru jika jam tersebut belum masuk Queue
            </span>
          </label>
        </div>
        <div class="lg:col-span-5">
          <input
            type="text"
            v-model="form.note"
            maxlength="240"
            placeholder="Catatan internal Superadmin (opsional, tidak tampil ke role lain)..."
            class="w-full px-3 py-1.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:border-blue-500"
          />
        </div>
      </div>
    </div>

    <!-- STEP 2: Tabel Parameter & Rentang Nilai -->
    <div class="saas-card rounded-2xl bg-white dark:bg-[#0F172A]/90 border border-slate-200 dark:border-slate-800 overflow-hidden">
      <div class="p-4 border-b border-slate-200 dark:border-slate-800 flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div class="flex items-center space-x-2">
          <span class="w-5 h-5 rounded-full bg-blue-600 text-white text-3xs font-bold flex items-center justify-center">2</span>
          <div>
            <h2 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white">
              Pilih Parameter &amp; Rentang Nilai ({{ selectedParamCount }} / {{ paramRows.length }} dipilih)
            </h2>
            <p class="text-3xs text-slate-500 dark:text-slate-400">
              Isi <code class="font-mono">Min = Max</code> jika ingin menyuntikkan angka konstan (misal <code class="font-mono">debit: 0.85 .. 0.85</code>), atau isi rentang untuk fluktuasi acak alami.
            </p>
          </div>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <button
            type="button"
            @click="selectAllParams(true)"
            class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-3xs font-semibold transition cursor-pointer"
          >
            Pilih Semua
          </button>
          <button
            type="button"
            @click="selectAllParams(false)"
            class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-3xs font-semibold transition cursor-pointer"
          >
            Reset Pilihan
          </button>
          <button
            type="button"
            @click="applySuggestedRangesAll"
            class="px-2.5 py-1 rounded-lg bg-indigo-500/15 hover:bg-indigo-500/25 text-indigo-600 dark:text-indigo-300 border border-indigo-500/30 text-3xs font-bold transition cursor-pointer"
          >
            ✨ Auto-Fill Semua dari Saran Sensor (±3%)
          </button>
        </div>
      </div>

      <div v-if="hintsLoading" class="p-8 text-center text-xs text-slate-400">
        Memuat daftar parameter &amp; definisi agregasi...
      </div>
      <div v-else-if="paramRows.length === 0" class="p-8 text-center text-xs text-slate-400">
        Pilih device yang memiliki parameter aktif terlebih dahulu.
      </div>
      <div v-else class="overflow-x-auto">
        <table class="w-full text-left text-xs font-sans">
          <thead class="bg-slate-50 dark:bg-[#0B0F19] text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800">
            <tr>
              <th class="py-2.5 px-4 w-10">Pilih</th>
              <th class="py-2.5 px-3">Parameter</th>
              <th class="py-2.5 px-3">Tipe &amp; Skala</th>
              <th class="py-2.5 px-3">Agregasi Aktif</th>
              <th class="py-2.5 px-3 text-right">Pembacaan Terakhir</th>
              <th class="py-2.5 px-3 w-36">Nilai Minimum</th>
              <th class="py-2.5 px-3 w-36">Nilai Maksimum</th>
              <th class="py-2.5 px-4 text-right">Bantuan Cepat</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-800/60">
            <tr
              v-for="row in paramRows"
              :key="row.parameter_id"
              class="transition"
              :class="row.selected ? 'bg-blue-50/40 dark:bg-blue-950/15' : 'opacity-65 hover:opacity-100'"
            >
              <td class="py-2.5 px-4">
                <input
                  type="checkbox"
                  v-model="row.selected"
                  class="rounded border-slate-400 text-blue-600 focus:ring-blue-500 cursor-pointer"
                />
              </td>
              <td class="py-2.5 px-3">
                <div class="font-bold text-slate-900 dark:text-white font-mono">
                  {{ row.parameter_code }}
                  <span v-if="row.unit" class="ml-1 text-3xs font-normal text-slate-500 dark:text-slate-400">({{ row.unit }})</span>
                </div>
                <div class="text-3xs text-slate-500 dark:text-slate-400">{{ row.parameter_name }}</div>
              </td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-600 dark:text-slate-300">
                <div>{{ row.data_type }} · ×{{ row.scale }} <span v-if="row.offset">+ {{ row.offset }}</span></div>
                <div v-if="row.formula" class="text-indigo-500 dark:text-indigo-300 truncate max-w-[160px]" :title="row.formula">
                  ƒ(x): {{ row.formula }}
                </div>
              </td>
              <td class="py-2.5 px-3">
                <div v-if="row.aggregation_intervals && row.aggregation_intervals.length" class="flex flex-wrap gap-1">
                  <span
                    v-for="iv in row.aggregation_intervals"
                    :key="iv"
                    class="px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-300 border border-emerald-500/25 text-4xs font-mono font-bold"
                  >
                    {{ formatIntervalLabel(iv) }}
                  </span>
                </div>
                <span v-else class="text-3xs text-amber-500 font-mono">Hanya Raw</span>
              </td>
              <td class="py-2.5 px-3 text-right font-mono">
                <div v-if="row.current_processed_value !== null && row.current_processed_value !== undefined" class="font-bold text-slate-900 dark:text-white">
                  {{ row.current_processed_value }} <span class="text-3xs font-normal text-slate-400">{{ row.unit }}</span>
                </div>
                <div v-else class="text-slate-400">--</div>
                <div v-if="row.current_quality" class="text-4xs" :class="row.current_quality === 'GOOD' ? 'text-emerald-500' : 'text-amber-500'">
                  {{ row.current_quality }}
                </div>
              </td>
              <td class="py-2.5 px-3">
                <input
                  type="number"
                  step="any"
                  v-model.number="row.minValue"
                  :disabled="!row.selected"
                  class="w-full px-2.5 py-1.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs font-mono text-slate-900 dark:text-white disabled:opacity-40 focus:outline-none focus:border-blue-500"
                />
              </td>
              <td class="py-2.5 px-3">
                <input
                  type="number"
                  step="any"
                  v-model.number="row.maxValue"
                  :disabled="!row.selected"
                  class="w-full px-2.5 py-1.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700 text-xs font-mono text-slate-900 dark:text-white disabled:opacity-40 focus:outline-none focus:border-blue-500"
                />
              </td>
              <td class="py-2.5 px-4 text-right space-x-1">
                <button
                  type="button"
                  @click="applyRowSuggestion(row)"
                  class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-3xs font-mono text-slate-700 dark:text-slate-300 transition cursor-pointer"
                  title="Isi rentang saran otomatis"
                >
                  {{ row.suggested_min }}..{{ row.suggested_max }}
                </button>
                <button
                  type="button"
                  @click="setRowConstant(row)"
                  class="px-2 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-3xs font-mono text-slate-600 dark:text-slate-400 transition cursor-pointer"
                  title="Samakan Max = Min (Konstan)"
                >
                  =Konstan
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Action Footer -->
      <div class="p-4 bg-slate-50/80 dark:bg-[#0B0F19]/80 border-t border-slate-200 dark:border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div class="text-2xs text-slate-500 dark:text-slate-400">
          Mode aktif: <strong class="text-slate-800 dark:text-slate-200">{{ form.conflictMode === 'OVERWRITE_ALL' ? 'TIMPA SEMUA (OVERWRITE)' : 'ISI KOSONG / NULL SAJA (AMAN)' }}</strong>
          · Generator <strong class="text-slate-800 dark:text-slate-200">{{ form.randomMode }}</strong>
        </div>
        <div class="flex items-center space-x-2.5">
          <button
            type="button"
            @click="runPreview"
            :disabled="previewLoading || executeLoading || selectedParamCount === 0"
            class="px-4 py-2 rounded-xl bg-slate-200 dark:bg-slate-800 hover:bg-slate-300 dark:hover:bg-slate-700 disabled:opacity-40 text-slate-900 dark:text-white text-xs font-bold transition cursor-pointer flex items-center space-x-1.5"
          >
            <span v-if="previewLoading" class="animate-spin w-3.5 h-3.5 border-2 border-current border-t-transparent rounded-full"></span>
            <span>🔍 Simulasi / Preview</span>
          </button>
          <button
            type="button"
            @click="confirmAndExecute"
            :disabled="executeLoading || previewLoading || selectedParamCount === 0"
            class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 disabled:opacity-40 text-white text-xs font-bold shadow-sm transition cursor-pointer flex items-center space-x-1.5"
          >
            <span v-if="executeLoading" class="animate-spin w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full"></span>
            <span>⚡ Suntik Data &amp; Hitung Agregasi</span>
          </button>
        </div>
      </div>
    </div>

    <!-- STEP 3: Hasil Simulasi / Preview atau Ringkasan Eksekusi -->
    <div
      v-if="previewResult"
      class="saas-card p-4 md:p-5 rounded-2xl bg-white dark:bg-[#0F172A]/90 border border-blue-300/70 dark:border-blue-800/60 space-y-4"
    >
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-200 dark:border-slate-800 pb-3">
        <div class="flex items-center space-x-2">
          <span class="px-2 py-0.5 rounded bg-blue-500/15 text-blue-600 dark:text-blue-300 border border-blue-500/30 text-3xs font-mono font-bold">
            {{ lastExecutedBatch ? ('TERSIMPAN · ' + lastExecutedBatch.batch_code) : 'SIMULASI DRY-RUN (BELUM DISIMPAN)' }}
          </span>
          <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white">
            Rincian Perhitungan Sampel &amp; Agregasi
          </h3>
        </div>
        <button type="button" @click="previewResult = null; lastExecutedBatch = null" class="text-3xs text-slate-400 hover:text-white font-mono">
          Tutup Preview ✕
        </button>
      </div>

      <!-- Summary KPIs -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800">
          <div class="text-3xs uppercase font-semibold text-slate-500">Sampel Akan Disuntik</div>
          <div class="text-xl font-mono font-bold text-blue-600 dark:text-blue-400 mt-0.5">{{ previewResult.total_planned_samples }}</div>
          <div class="text-4xs text-slate-400 font-mono">dari {{ previewResult.total_candidate_slots }} slot kandidat</div>
        </div>
        <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800">
          <div class="text-3xs uppercase font-semibold text-emerald-600 dark:text-emerald-400">Data Valid Dipertahankan</div>
          <div class="text-xl font-mono font-bold text-emerald-600 dark:text-emerald-400 mt-0.5">{{ previewResult.total_skipped_samples }}</div>
          <div class="text-4xs text-slate-400 font-mono">tidak ditimpa (aman)</div>
        </div>
        <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800">
          <div class="text-3xs uppercase font-semibold text-indigo-600 dark:text-indigo-400">Bucket Agregasi Dihitung</div>
          <div class="text-xl font-mono font-bold text-indigo-600 dark:text-indigo-400 mt-0.5">
            {{ lastExecutedBatch ? lastExecutedBatch.aggregated_buckets : previewResult.estimated_agg_buckets }}
          </div>
          <div class="text-4xs text-slate-400 font-mono">RAW &amp; CUSTOMER_PROCESSED</div>
        </div>
        <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800">
          <div class="text-3xs uppercase font-semibold text-amber-600 dark:text-amber-400">Antrean Outbound Segar</div>
          <div class="text-xl font-mono font-bold text-amber-600 dark:text-amber-400 mt-0.5">
            {{ lastExecutedBatch ? lastExecutedBatch.refreshed_outbound : (form.refreshOutbound ? 'Auto' : 'Off') }}
          </div>
          <div class="text-4xs text-slate-400 font-mono">batch queue diperbarui</div>
        </div>
      </div>

      <!-- Per-Parameter Breakdown -->
      <div class="overflow-x-auto">
        <table class="w-full text-left text-2xs font-mono">
          <thead class="bg-slate-50 dark:bg-[#0B0F19] text-3xs uppercase text-slate-500 border-b border-slate-200 dark:border-slate-800">
            <tr>
              <th class="py-2 px-3">Parameter</th>
              <th class="py-2 px-3 text-right">Disuntik</th>
              <th class="py-2 px-3 text-right">Dilewati (Valid)</th>
              <th class="py-2 px-3 text-right">Valid Eksisting</th>
              <th class="py-2 px-3 text-right">Error/Null Eksisting</th>
              <th class="py-2 px-3 text-right">Processed (Min..Max / Avg)</th>
              <th class="py-2 px-3 text-right">Raw Register (Min..Max)</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-800/60">
            <tr v-for="ps in previewResult.parameter_summaries" :key="ps.parameter_id">
              <td class="py-2 px-3 font-bold text-slate-900 dark:text-white">{{ ps.parameter_code }} <span class="text-slate-400 font-normal">{{ ps.unit }}</span></td>
              <td class="py-2 px-3 text-right font-bold text-blue-600 dark:text-blue-400">{{ ps.planned_samples }}</td>
              <td class="py-2 px-3 text-right text-emerald-600 dark:text-emerald-400">{{ ps.skipped_samples }}</td>
              <td class="py-2 px-3 text-right text-slate-500">{{ ps.existing_valid_in_window }}</td>
              <td class="py-2 px-3 text-right text-rose-500">{{ ps.existing_bad_in_window }}</td>
              <td class="py-2 px-3 text-right text-slate-800 dark:text-slate-200">
                <span v-if="ps.planned_samples > 0">{{ ps.generated_min_processed }} .. {{ ps.generated_max_processed }} (avg {{ ps.generated_avg_processed }})</span>
                <span v-else class="text-slate-400">--</span>
              </td>
              <td class="py-2 px-3 text-right text-slate-500">
                <span v-if="ps.planned_samples > 0">{{ ps.generated_min_raw }} .. {{ ps.generated_max_raw }}</span>
                <span v-else>--</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Sample Preview Rows -->
      <div v-if="previewResult.sample_preview && previewResult.sample_preview.length > 0" class="space-y-2">
        <button
          type="button"
          @click="showSamplePreviewRows = !showSamplePreviewRows"
          class="text-3xs font-mono font-semibold text-blue-600 dark:text-blue-400 hover:underline cursor-pointer"
        >
          {{ showSamplePreviewRows ? '▼ Sembunyikan Contoh Baris Raw & Hex' : ('▶ Lihat Contoh ' + previewResult.sample_preview.length + ' Baris Sampel Raw & Hex yang Dibangkitkan') }}
        </button>
        <div v-if="showSamplePreviewRows" class="max-h-60 overflow-y-auto rounded-xl border border-slate-200 dark:border-slate-800">
          <table class="w-full text-left text-2xs font-mono">
            <thead class="bg-slate-50 dark:bg-[#0B0F19] text-3xs uppercase text-slate-500 sticky top-0">
              <tr>
                <th class="py-1.5 px-3">Waktu (Lokal)</th>
                <th class="py-1.5 px-3">Parameter</th>
                <th class="py-1.5 px-3 text-right">Processed</th>
                <th class="py-1.5 px-3 text-right">Raw</th>
                <th class="py-1.5 px-3">Modbus Raw Hex</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 dark:divide-slate-800/60">
              <tr v-for="(sp, idx) in previewResult.sample_preview" :key="idx">
                <td class="py-1.5 px-3 text-slate-600 dark:text-slate-300">{{ formatDateTime(sp.received_at) }}</td>
                <td class="py-1.5 px-3 font-bold text-slate-900 dark:text-white">{{ sp.parameter_code }}</td>
                <td class="py-1.5 px-3 text-right font-bold text-blue-600 dark:text-blue-400">{{ sp.processed_value }} {{ sp.unit }}</td>
                <td class="py-1.5 px-3 text-right text-slate-500">{{ sp.raw_value }}</td>
                <td class="py-1.5 px-3 text-emerald-600 dark:text-emerald-400 tracking-wider">{{ sp.raw_hex }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- STEP 4: Riwayat Batch Suntikan & 1-Klik Rollback -->
    <div class="saas-card rounded-2xl bg-white dark:bg-[#0F172A]/90 border border-slate-200 dark:border-slate-800 overflow-hidden">
      <div class="p-4 border-b border-slate-200 dark:border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
        <div>
          <h2 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white">
            Riwayat Batch Suntikan Superadmin ({{ batchesTotal }})
          </h2>
          <p class="text-3xs text-slate-500 dark:text-slate-400">
            Tabel ini khusus internal Superadmin. Klik <strong>Rollback</strong> untuk mencabut seluruh titik sampel batch dan menghitung ulang agregasi ke kondisi semula.
          </p>
        </div>
        <div class="flex items-center space-x-2">
          <label class="inline-flex items-center space-x-1.5 text-2xs text-slate-600 dark:text-slate-300 cursor-pointer">
            <input type="checkbox" v-model="filterBatchesByDevice" @change="fetchBatches" class="rounded border-slate-400 text-blue-600" />
            <span>Hanya Device Terpilih</span>
          </label>
          <button
            type="button"
            @click="fetchBatches"
            class="px-3 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-xs font-semibold text-slate-700 dark:text-slate-200 transition cursor-pointer"
          >
            ↻ Refresh
          </button>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs font-sans">
          <thead class="bg-slate-50 dark:bg-[#0B0F19] text-3xs font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 border-b border-slate-200 dark:border-slate-800">
            <tr>
              <th class="py-2.5 px-4">Kode Batch</th>
              <th class="py-2.5 px-3">Device</th>
              <th class="py-2.5 px-3">Rentang Waktu</th>
              <th class="py-2.5 px-3">Parameter &amp; Range</th>
              <th class="py-2.5 px-3 text-right">Raw / Agregasi / Queue</th>
              <th class="py-2.5 px-3 text-center">Status</th>
              <th class="py-2.5 px-4 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-slate-800/60 text-2xs">
            <tr v-for="b in batches" :key="b.id" class="hover:bg-slate-50 dark:hover:bg-slate-800/30 transition">
              <td class="py-2.5 px-4 font-mono">
                <div class="font-bold text-indigo-600 dark:text-indigo-300">{{ b.batch_code }}</div>
                <div class="text-4xs text-slate-400">{{ formatDateTime(b.created_at) }} · @{{ b.created_by }}</div>
              </td>
              <td class="py-2.5 px-3 font-mono">
                <div class="font-bold text-slate-900 dark:text-white">{{ b.device_code }}</div>
                <div class="text-4xs text-slate-400">{{ b.device_name }}</div>
              </td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-700 dark:text-slate-300">
                <div>{{ formatDateTime(b.start_time) }}</div>
                <div class="text-slate-400">s/d {{ formatDateTime(b.end_time) }} ({{ b.sample_step_seconds }}s)</div>
              </td>
              <td class="py-2.5 px-3 max-w-xs">
                <div class="font-mono text-3xs text-slate-700 dark:text-slate-200 truncate" :title="b.parameters_summary">
                  {{ b.parameters_summary }}
                </div>
                <div class="text-4xs text-slate-400">
                  {{ b.conflict_mode }} · {{ b.random_mode }} · {{ b.value_target }}
                  <span v-if="b.note" class="text-amber-500 ml-1">· "{{ b.note }}"</span>
                </div>
              </td>
              <td class="py-2.5 px-3 text-right font-mono text-3xs">
                <div><strong class="text-slate-900 dark:text-white">{{ b.injected_raw_count }}</strong> raw</div>
                <div class="text-slate-400">{{ b.aggregated_buckets }} agg · {{ b.refreshed_outbound }} out</div>
              </td>
              <td class="py-2.5 px-3 text-center font-mono">
                <span
                  class="px-2 py-0.5 rounded text-4xs font-bold"
                  :class="b.status === 'ACTIVE'
                    ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300 border border-emerald-500/30'
                    : 'bg-slate-200 dark:bg-slate-800 text-slate-500 border border-slate-300 dark:border-slate-700'"
                >
                  {{ b.status }}
                </span>
              </td>
              <td class="py-2.5 px-4 text-right">
                <button
                  v-if="b.status === 'ACTIVE'"
                  type="button"
                  @click="openRollbackModal(b)"
                  :disabled="rollbackLoadingId === b.id"
                  class="px-2.5 py-1 rounded-lg bg-rose-500/15 hover:bg-rose-600 text-rose-600 dark:text-rose-300 hover:text-white border border-rose-500/30 text-3xs font-bold transition cursor-pointer"
                >
                  {{ rollbackLoadingId === b.id ? 'Menghapus...' : '↩ Rollback' }}
                </button>
                <span v-else class="text-4xs font-mono text-slate-400">
                  Oleh @{{ b.rolled_back_by || 'superadmin' }}
                </span>
              </td>
            </tr>
            <tr v-if="batches.length === 0">
              <td colspan="7" class="py-8 text-center text-xs text-slate-400">
                Belum ada riwayat batch suntikan telemetri.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Rollback Confirmation Modal -->
    <div
      v-if="batchToRollback"
      class="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4"
    >
      <div class="bg-white dark:bg-[#111827] border border-rose-300 dark:border-rose-900/70 rounded-2xl w-full max-w-md p-5 space-y-4 shadow-2xl">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white">
            Rollback Batch Suntikan {{ batchToRollback.batch_code }}?
          </h3>
          <button type="button" @click="batchToRollback = null" class="text-slate-400 hover:text-white">✕</button>
        </div>
        <p class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed">
          Seluruh <strong>{{ batchToRollback.injected_raw_count }} baris RawData</strong> dari batch ini akan dihapus berdasarkan primary key secara presisi (data asli sensor tidak tersentuh), bucket agregasi kosong dibersihkan, dan payload antrean Outbound dihitung ulang otomatis.
        </p>
        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-slate-800">
          <button
            type="button"
            @click="batchToRollback = null"
            class="px-3.5 py-1.5 rounded-xl bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold"
          >
            Batal
          </button>
          <button
            type="button"
            @click="executeRollback"
            :disabled="rollbackLoadingId !== null"
            class="px-4 py-1.5 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold transition cursor-pointer"
          >
            Ya, Rollback Sekarang
          </button>
        </div>
      </div>
    </div>
  </div>

  <!-- Fallback if non-superadmin somehow mounts view -->
  <div v-else class="p-10 text-center text-sm text-rose-500 font-semibold">
    403 — Halaman ini hanya dapat diakses oleh Superadmin.
  </div>
</template>

<script>
import api from '../services/api';

function toLocalDatetimeInputValue(dateObj) {
  const d = new Date(dateObj);
  const pad = n => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
}

export default {
  name: 'TelemetryInjectionView',
  data() {
    const now = new Date();
    const prevHourEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours(), 0, 0);
    const prevHourStart = new Date(prevHourEnd.getTime() - 3600 * 1000);

    return {
      devices: [],
      selectedDeviceId: '',
      deviceHints: null,
      hintsLoading: false,
      paramRows: [],

      showUnixHelper: false,
      unixHelperInput: '',

      form: {
        startTime: toLocalDatetimeInputValue(prevHourStart),
        endTime: toLocalDatetimeInputValue(prevHourEnd),
        sampleStepSeconds: 30,
        conflictMode: 'FILL_MISSING_AND_NULL',
        randomMode: 'SMOOTH_WALK',
        valueTarget: 'PROCESSED',
        refreshOutbound: true,
        enqueueMissingOutbound: true,
        note: '',
      },

      previewLoading: false,
      executeLoading: false,
      previewResult: null,
      lastExecutedBatch: null,
      showSamplePreviewRows: false,

      batches: [],
      batchesTotal: 0,
      filterBatchesByDevice: false,
      batchToRollback: null,
      rollbackLoadingId: null,

      feedbackMessage: '',
      feedbackType: 'success',
    };
  },
  computed: {
    userRole() {
      const u = (this.$store && this.$store.state && this.$store.state.user) || {};
      return (u.role || '').toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    selectedParamCount() {
      return this.paramRows.filter(r => r.selected).length;
    },
  },
  mounted() {
    if (!this.isSuperAdmin) {
      this.$router.replace('/access-denied');
      return;
    }
    this.loadDevices();
    this.fetchBatches();
  },
  methods: {
    showFeedback(msg, type = 'success') {
      this.feedbackMessage = msg;
      this.feedbackType = type;
    },
    formatIntervalLabel(sec) {
      const n = Number(sec) || 0;
      if (n >= 3600 && n % 3600 === 0) return `${n / 3600}h`;
      if (n >= 60 && n % 60 === 0) return `${n / 60}m`;
      return `${n}s`;
    },
    formatDateTime(iso) {
      if (!iso) return '--';
      const d = new Date(iso);
      if (Number.isNaN(d.getTime())) return iso;
      return d.toLocaleString('id-ID', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    },
    async loadDevices() {
      try {
        const res = await api.get('/devices');
        const list = (res.data && res.data.data) || res.data || [];
        this.devices = Array.isArray(list) ? list : (list.items || []);

        const qDev = Number(this.$route.query.device_id);
        if (qDev && this.devices.some(d => d.id === qDev)) {
          this.selectedDeviceId = qDev;
        } else {
          const firstActive = this.devices.find(d => d.enabled !== false) || this.devices[0];
          if (firstActive) this.selectedDeviceId = firstActive.id;
        }
        if (this.selectedDeviceId) {
          await this.onDeviceChange();
        }
      } catch (err) {
        this.showFeedback('Gagal memuat daftar device: ' + ((err.response && err.response.data && err.response.data.message) || err.message), 'error');
      }
    },
    async onDeviceChange() {
      if (!this.selectedDeviceId) return;
      this.hintsLoading = true;
      this.previewResult = null;
      this.lastExecutedBatch = null;
      try {
        const res = await api.get(`/superadmin/telemetry-inject/devices/${this.selectedDeviceId}/hints`);
        const hints = (res.data && res.data.data) || res.data;
        this.deviceHints = hints;
        this.paramRows = (hints.parameters || []).map(p => ({
          ...p,
          selected: true,
          minValue: p.suggested_min,
          maxValue: p.suggested_max,
        }));
        if (this.filterBatchesByDevice) {
          this.fetchBatches();
        }
      } catch (err) {
        this.showFeedback('Gagal memuat parameter device: ' + ((err.response && err.response.data && err.response.data.message) || err.message), 'error');
      } finally {
        this.hintsLoading = false;
      }
    },
    applyWindowPreset(preset) {
      const now = new Date();
      let start;
      let end = now;

      if (preset === 'prev_hour') {
        end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours(), 0, 0);
        start = new Date(end.getTime() - 3600 * 1000);
      } else if (preset === 'curr_hour') {
        start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours(), 0, 0);
        end = now;
      } else if (preset === 'last_1h') {
        start = new Date(now.getTime() - 3600 * 1000);
      } else if (preset === 'last_3h') {
        start = new Date(now.getTime() - 3 * 3600 * 1000);
      } else if (preset === 'last_6h') {
        start = new Date(now.getTime() - 6 * 3600 * 1000);
      } else if (preset === 'today') {
        start = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0);
      } else {
        return;
      }

      this.form.startTime = toLocalDatetimeInputValue(start);
      this.form.endTime = toLocalDatetimeInputValue(end);
    },
    applyUnixEpochToHourWindow() {
      const raw = Number(this.unixHelperInput);
      if (!raw || raw < 100000000) {
        this.showFeedback('Masukkan Unix Epoch detik yang valid (contoh: 1793777880)', 'error');
        return;
      }
      const sec = raw > 1e12 ? Math.floor(raw / 1000) : Math.floor(raw);
      const hourStartSec = Math.floor((sec - 1) / 3600) * 3600;
      const hourEndSec = hourStartSec + 3600;
      this.form.startTime = toLocalDatetimeInputValue(new Date(hourStartSec * 1000));
      this.form.endTime = toLocalDatetimeInputValue(new Date(hourEndSec * 1000));
      this.showFeedback('Rentang waktu disetel ke jam kalender penuh (:00 s/d :00) untuk epoch ' + sec, 'success');
    },
    applyUnixEpochExactBucket() {
      const raw = Number(this.unixHelperInput);
      if (!raw || raw < 100000000) {
        this.showFeedback('Masukkan Unix Epoch detik yang valid (contoh: 1793777880)', 'error');
        return;
      }
      const sec = raw > 1e12 ? Math.floor(raw / 1000) : Math.floor(raw);
      const startSec = sec - 120;
      this.form.startTime = toLocalDatetimeInputValue(new Date(startSec * 1000));
      this.form.endTime = toLocalDatetimeInputValue(new Date((startSec + 3600) * 1000));
      this.showFeedback('Rentang waktu disetel tepat 30 bucket SPARING (mulai bucket pertama ' + sec + ')', 'success');
    },
    selectAllParams(flag) {
      this.paramRows.forEach(r => {
        r.selected = flag;
      });
    },
    applySuggestedRangesAll() {
      this.paramRows.forEach(r => {
        r.minValue = r.suggested_min;
        r.maxValue = r.suggested_max;
      });
    },
    applyRowSuggestion(row) {
      row.selected = true;
      row.minValue = row.suggested_min;
      row.maxValue = row.suggested_max;
    },
    setRowConstant(row) {
      row.selected = true;
      row.maxValue = row.minValue;
    },
    buildRequestPayload() {
      const start = new Date(this.form.startTime);
      const end = new Date(this.form.endTime);
      if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
        throw new Error('Format Waktu Mulai atau Waktu Selesai belum lengkap/valid.');
      }
      const selected = this.paramRows
        .filter(r => r.selected)
        .map(r => ({
          parameter_id: r.parameter_id,
          min_value: Number(r.minValue),
          max_value: Number(r.maxValue),
        }));
      if (selected.length === 0) {
        throw new Error('Pilih minimal 1 parameter terlebih dahulu.');
      }
      return {
        device_id: Number(this.selectedDeviceId),
        start_time: start.toISOString(),
        end_time: end.toISOString(),
        sample_step_seconds: Number(this.form.sampleStepSeconds) || 30,
        conflict_mode: this.form.conflictMode,
        random_mode: this.form.randomMode,
        value_target: this.form.valueTarget,
        refresh_outbound: !!this.form.refreshOutbound,
        enqueue_missing_outbound: !!this.form.enqueueMissingOutbound,
        note: this.form.note || '',
        parameters: selected,
      };
    },
    async runPreview() {
      this.feedbackMessage = '';
      let payload;
      try {
        payload = this.buildRequestPayload();
      } catch (e) {
        this.showFeedback(e.message, 'error');
        return;
      }

      this.previewLoading = true;
      try {
        const res = await api.post('/superadmin/telemetry-inject/preview', payload);
        this.previewResult = (res.data && res.data.data) || res.data;
        this.lastExecutedBatch = null;
      } catch (err) {
        this.showFeedback('Simulasi gagal: ' + ((err.response && err.response.data && err.response.data.message) || err.message), 'error');
      } finally {
        this.previewLoading = false;
      }
    },
    async confirmAndExecute() {
      this.feedbackMessage = '';
      let payload;
      try {
        payload = this.buildRequestPayload();
      } catch (e) {
        this.showFeedback(e.message, 'error');
        return;
      }

      if (this.form.conflictMode === 'OVERWRITE_ALL') {
        const ok = window.confirm('Mode OVERWRITE_ALL dipilih: data eksisting pada rentang waktu dan parameter terpilih akan ditimpa. Lanjutkan penyuntikan?');
        if (!ok) return;
      }

      this.executeLoading = true;
      try {
        const res = await api.post('/superadmin/telemetry-inject/execute', payload);
        const data = (res.data && res.data.data) || res.data;
        this.lastExecutedBatch = data.batch;
        this.previewResult = data.preview;
        this.showFeedback(
          `Berhasil menyuntikkan ${data.batch.injected_raw_count} sampel (${data.batch.batch_code}), menghitung ulang ${data.batch.aggregated_buckets} bucket agregasi, dan menyegarkan ${data.batch.refreshed_outbound} batch Outbound.`,
          'success'
        );
        await this.fetchBatches();
      } catch (err) {
        this.showFeedback('Eksekusi gagal: ' + ((err.response && err.response.data && err.response.data.message) || err.message), 'error');
      } finally {
        this.executeLoading = false;
      }
    },
    async fetchBatches() {
      try {
        const params = { page: 1, page_size: 25 };
        if (this.filterBatchesByDevice && this.selectedDeviceId) {
          params.device_id = this.selectedDeviceId;
        }
        const res = await api.get('/superadmin/telemetry-inject/batches', { params });
        const payload = (res.data && res.data.data) || res.data || {};
        this.batches = payload.items || [];
        this.batchesTotal = payload.total || this.batches.length;
      } catch (_) {
        // Ignore silent fetch error on background refresh
      }
    },
    openRollbackModal(batch) {
      this.batchToRollback = batch;
    },
    async executeRollback() {
      if (!this.batchToRollback) return;
      const target = this.batchToRollback;
      this.rollbackLoadingId = target.id;
      try {
        const res = await api.post(`/superadmin/telemetry-inject/batches/${target.id}/rollback`);
        const out = (res.data && res.data.data) || res.data || {};
        this.batchToRollback = null;
        this.showFeedback(
          `Batch ${target.batch_code} berhasil di-rollback (${out.deleted_raw_count || 0} raw dihapus, agregasi & antrean Outbound telah disinkronkan ulang).`,
          'success'
        );
        await this.fetchBatches();
      } catch (err) {
        this.showFeedback('Rollback gagal: ' + ((err.response && err.response.data && err.response.data.message) || err.message), 'error');
      } finally {
        this.rollbackLoadingId = null;
      }
    },
  },
};
</script>
