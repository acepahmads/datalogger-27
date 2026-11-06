<template>
  <div class="space-y-4 p-3 sm:p-5 max-w-[1920px] mx-auto min-h-screen text-slate-800 dark:text-slate-100 font-sans">
    <!-- Top Header & Mode Switch -->
    <div class="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-slate-200 dark:border-[#20354B]">
      <div>
        <div class="flex flex-wrap items-center gap-2">
          <span class="w-2.5 h-2.5 rounded-full bg-sky-500 animate-pulse"></span>
          <h1 class="text-lg sm:text-xl font-bold tracking-tight text-slate-900 dark:text-white">
            Kalibrasi Sensor &amp; Regresi Multi-Titik
          </h1>
          <span class="px-2 py-0.5 rounded-full text-2xs font-bold bg-sky-500/10 text-sky-700 dark:text-sky-300 border border-sky-500/30">
            1–10 Titik Standar
          </span>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
          Kalibrasi presisi regresi linier, polinomial kuadratik orde-2, &amp; 1-titik untuk kepatuhan
          <strong class="text-slate-700 dark:text-slate-200">AQMS</strong>,
          <strong class="text-slate-700 dark:text-slate-200">SPARING</strong>, dan
          <strong class="text-slate-700 dark:text-slate-200">WQMS ONLIMO</strong>.
        </p>
      </div>

      <!-- Right Controls: Domain Filter & View Tabs -->
      <div class="flex flex-wrap items-center gap-2">
        <!-- Domain Filter Pills -->
        <div class="flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-2xs">
          <button
            v-for="dom in domainOptions"
            :key="dom.value"
            type="button"
            @click="onSelectDomainFilter(dom.value)"
            class="px-2.5 py-1 rounded-md font-semibold transition cursor-pointer"
            :class="selectedDomainFilter === dom.value
              ? 'bg-sky-600 text-white btn-text shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            {{ dom.badge === 'ALL' ? 'Semua' : dom.badge }}
          </button>
        </div>

        <!-- Workspace vs Full History Tab -->
        <div class="flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs">
          <button
            type="button"
            id="tab-calibration-workspace"
            @click="activeTab = 'workspace'"
            class="px-3 py-1.5 rounded-md font-semibold transition flex items-center gap-1.5 cursor-pointer"
            :class="activeTab === 'workspace'
              ? 'bg-blue-600 text-white btn-text shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            <span>🎯 Kerja Kalibrasi</span>
          </button>
          <button
            type="button"
            id="tab-calibration-history"
            @click="switchToHistoryTab"
            class="px-3 py-1.5 rounded-md font-semibold transition flex items-center gap-1.5 cursor-pointer"
            :class="activeTab === 'history'
              ? 'bg-blue-600 text-white btn-text shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            <span>🕒 Riwayat &amp; Log Audit ({{ totalLogs }})</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Notification Toast Banner -->
    <div
      v-if="banner.message"
      class="px-3.5 py-2.5 rounded-xl border flex items-center justify-between text-xs font-medium transition"
      :class="banner.type === 'error'
        ? 'bg-rose-50 dark:bg-rose-950/40 text-rose-700 dark:text-rose-300 border-rose-200 dark:border-rose-800/60'
        : 'bg-emerald-50 dark:bg-emerald-950/40 text-emerald-700 dark:text-emerald-300 border-emerald-200 dark:border-emerald-800/60'"
    >
      <span>{{ banner.message }}</span>
      <button type="button" @click="banner.message = ''" class="ml-3 text-slate-400 hover:text-slate-700 dark:hover:text-white cursor-pointer">✕</button>
    </div>

    <!-- ========================================================= -->
    <!-- TAB 1: KERJA KALIBRASI                                    -->
    <!-- ========================================================= -->
    <div v-if="activeTab === 'workspace'" class="space-y-4">
      <!-- Top 3-Card Control & Live Readout Strip -->
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-3.5">
        <!-- Card 1: Device, Parameter & Session Guard -->
        <div class="saas-card p-3.5 lg:col-span-5 flex flex-col justify-between space-y-2.5">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
            <div>
              <label class="block text-2xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Perangkat Sensor</label>
              <select
                id="select-calibration-device"
                v-model="selectedDeviceId"
                @change="onDeviceChange"
                class="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#244356] bg-white dark:bg-[#0D2232] text-xs font-semibold text-slate-800 dark:text-slate-100 focus:outline-none"
              >
                <option :value="0">Semua Perangkat ({{ deviceOptions.length }})</option>
                <option v-for="dev in deviceOptions" :key="dev.device_id" :value="dev.device_id">
                  {{ dev.device_name }} ({{ dev.device_code }} · {{ dev.protocol }})
                </option>
              </select>
            </div>

            <div>
              <label class="block text-2xs font-semibold text-slate-500 dark:text-slate-400 mb-1">Parameter Kalibrasi</label>
              <select
                id="select-calibration-parameter"
                v-model="selectedParameterId"
                @change="onParameterChange"
                class="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#244356] bg-white dark:bg-[#0D2232] text-xs font-semibold text-slate-800 dark:text-slate-100 focus:outline-none"
              >
                <option v-if="filteredParameters.length === 0" :value="0">-- Tidak ada parameter analog --</option>
                <option v-for="p in filteredParameters" :key="p.parameter_id" :value="p.parameter_id">
                  [{{ domainShortBadge(p.effective_domain) }}] {{ p.parameter_name }} ({{ p.unit || '-' }}) {{ p.calibration_enabled ? '✓' : '' }}
                </option>
              </select>
            </div>
          </div>

          <!-- Domain Tag Assignment & Calibration Session Toggle -->
          <div class="pt-2 border-t border-slate-100 dark:border-[#1E3648] flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2">
              <button
                type="button"
                id="btn-toggle-calibration-session"
                :disabled="!currentParam || togglingSession"
                @click="toggleCalibrationSession"
                class="relative inline-flex h-6 w-11 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none disabled:opacity-50"
                :class="currentParam && currentParam.calibration_session_active ? 'bg-amber-500' : 'bg-slate-300 dark:bg-slate-700'"
                :title="currentParam && currentParam.calibration_session_active ? 'Nonaktifkan Mode Sesi Kalibrasi' : 'Aktifkan Mode Sesi Kalibrasi'"
              >
                <span
                  class="pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                  :class="currentParam && currentParam.calibration_session_active ? 'translate-x-5' : 'translate-x-0'"
                ></span>
              </button>
              <div>
                <div class="text-xs font-bold flex items-center gap-1.5">
                  <span>Mode Sesi Kalibrasi:</span>
                  <span
                    :class="currentParam && currentParam.calibration_session_active
                      ? 'text-amber-600 dark:text-amber-400'
                      : 'text-slate-500 dark:text-slate-400'"
                  >
                    {{ currentParam && currentParam.calibration_session_active ? 'AKTIF' : 'NONAKTIF' }}
                  </span>
                </div>
                <div class="text-2xs text-slate-500 dark:text-slate-400">
                  Alarm &amp; agregasi kepatuhan AQMS / SPARING / ONLIMO ditahan sementara
                </div>
              </div>
            </div>

            <!-- Parameter Compliance Domain Selector -->
            <div v-if="currentParam" class="flex items-center gap-1.5 text-2xs">
              <span class="text-slate-500 dark:text-slate-400">Domain:</span>
              <select
                v-model="draftDomain"
                @change="saveParameterDomainOnly"
                class="px-2 py-1 rounded-md border border-slate-200 dark:border-[#244356] bg-slate-50 dark:bg-[#0D2232] font-bold text-slate-700 dark:text-slate-200 text-2xs focus:outline-none"
              >
                <option value="AQMS">AQMS</option>
                <option value="SPARING">SPARING</option>
                <option value="WQMS_ONLIMO">WQMS ONLIMO</option>
              </select>
            </div>
          </div>
        </div>

        <!-- Card 2: Live Raw / Uncalibrated Sensor Reading (x) + Stability -->
        <div class="saas-card p-3.5 lg:col-span-4 flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold text-slate-600 dark:text-slate-300">Bacaan Mentah Sensor (x)</span>
            <button
              type="button"
              id="btn-poll-live-sensor"
              :disabled="!currentParam || pollingLive"
              @click="pollLiveSensor(true)"
              class="px-2 py-0.5 rounded-md text-2xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 transition cursor-pointer disabled:opacity-50"
              title="Baca langsung register sensor saat ini"
            >
              {{ pollingLive ? 'Membaca...' : '🔄 Baca Live' }}
            </button>
          </div>

          <div class="my-1.5 flex items-baseline gap-2">
            <span class="text-2xl sm:text-3xl font-extrabold tracking-tight tabular-nums text-slate-900 dark:text-white">
              {{ formattedLiveX }}
            </span>
            <span class="text-sm font-bold text-slate-500 dark:text-slate-400">
              {{ (currentParam && currentParam.unit) || '' }}
            </span>
          </div>

          <div class="flex items-center justify-between gap-2 pt-1">
            <span
              v-if="liveX !== null"
              class="px-2 py-0.5 rounded-full text-2xs font-bold tabular-nums inline-flex items-center gap-1"
              :class="stabilityInfo.isStable
                ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30'
                : 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/30'"
            >
              <span>● {{ stabilityInfo.isStable ? 'STABIL' : 'FLUKTUATIF' }}</span>
              <span>(σ = {{ stabilityInfo.sigma.toFixed(3) }})</span>
            </span>
            <span v-else class="text-2xs text-slate-400">Menunggu pembacaan telemetri...</span>

            <!-- Mini Sparkline SVG -->
            <svg v-if="sparklinePath" class="w-24 h-6 overflow-visible" viewBox="0 0 96 24">
              <path
                :d="sparklinePath"
                fill="none"
                stroke="#0284C7"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </div>
        </div>

        <!-- Card 3: Simulated & Active Calibrated Output (ŷ) -->
        <div class="saas-card p-3.5 lg:col-span-3 flex flex-col justify-between">
          <div class="flex items-center justify-between">
            <span class="text-xs font-semibold text-slate-600 dark:text-slate-300">Hasil Terkalibrasi (ŷ)</span>
            <span
              v-if="currentParam && currentParam.calibration_enabled"
              class="px-2 py-0.5 rounded text-3xs font-bold bg-sky-500/15 text-sky-700 dark:text-sky-300 border border-sky-500/30"
            >
              KURVA AKTIF
            </span>
            <span
              v-else
              class="px-2 py-0.5 rounded text-3xs font-semibold bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700"
            >
              DEFAULT (y = x)
            </span>
          </div>

          <div class="my-1.5 flex items-baseline gap-2">
            <span class="text-2xl sm:text-3xl font-extrabold tracking-tight tabular-nums text-sky-600 dark:text-sky-400">
              {{ formattedSimulatedY }}
            </span>
            <span class="text-sm font-bold text-sky-600/80 dark:text-sky-400/80">
              {{ (currentParam && currentParam.unit) || '' }}
            </span>
          </div>

          <div class="space-y-0.5 text-2xs text-slate-500 dark:text-slate-400 truncate">
            <div class="font-medium text-slate-700 dark:text-slate-200 truncate" :title="activeSavedEquation">
              Tersimpan: <span class="font-mono">{{ activeSavedEquation }}</span>
            </div>
            <div v-if="currentParam" class="text-3xs text-slate-400">
              Register: Reg {{ currentParam.register_address }} (Scale ×{{ currentParam.register_scale }} + {{ currentParam.register_offset }} tetap utuh)
            </div>
          </div>
        </div>
      </div>

      <!-- Main 2-Column Calibration Workspace -->
      <div class="grid grid-cols-1 xl:grid-cols-12 gap-4">
        <!-- LEFT COLUMN: 1–10 Point Calibration Table -->
        <div class="saas-card p-4 xl:col-span-7 space-y-3.5">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <div>
              <h2 class="text-sm font-bold text-slate-900 dark:text-white">
                Tabel Titik Standar Kalibrasi ({{ activePointsCount }} Aktif / {{ points.length }} dari Maks 10 Titik)
              </h2>
            </div>

            <!-- Domain Smart Preset Dropdown -->
            <div v-if="domainStandardPresets.length > 0" class="flex items-center gap-1.5">
              <select
                v-model="selectedStandardPresetId"
                @change="applyDomainStandardPreset"
                class="px-2.5 py-1 rounded-lg border border-sky-300 dark:border-sky-700 bg-sky-50/70 dark:bg-sky-950/30 text-2xs font-semibold text-sky-800 dark:text-sky-300 focus:outline-none cursor-pointer"
              >
                <option value="">🧪 Pilih Preset Larutan / Gas Standar...</option>
                <option v-for="pr in domainStandardPresets" :key="pr.id" :value="pr.id">
                  {{ pr.label }}
                </option>
              </select>
            </div>
          </div>

          <!-- Regression Model Selector Pills -->
          <div class="flex flex-wrap items-center gap-1.5">
            <span class="text-xs font-semibold text-slate-500 dark:text-slate-400 mr-1">Model:</span>
            <button
              v-for="m in modelOptions"
              :key="m.value"
              type="button"
              @click="selectModel(m.value)"
              class="px-2.5 py-1.5 rounded-lg text-xs font-semibold border transition cursor-pointer"
              :class="draftModel === m.value
                ? 'bg-blue-600 text-white btn-text border-blue-600 shadow-xs'
                : 'bg-slate-50 dark:bg-[#0D2232] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#244356] hover:bg-slate-100 dark:hover:bg-[#132d42]'"
              :title="m.description"
            >
              {{ m.shortLabel }}
            </button>
          </div>

          <!-- Add Point & Quick Point Count Presets (1..10) -->
          <div class="flex flex-wrap items-center justify-between gap-2 pt-1">
            <button
              type="button"
              id="btn-add-calibration-point"
              :disabled="points.length >= 10"
              @click="addPoint"
              class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-[#13293D] dark:hover:bg-[#1B3954] text-slate-800 dark:text-slate-100 border border-slate-200 dark:border-[#2A4B63] transition cursor-pointer disabled:opacity-40"
            >
              + Tambah Titik (Maks 10)
            </button>

            <div class="flex flex-wrap items-center gap-1 text-2xs">
              <span class="text-slate-500 dark:text-slate-400 mr-1">Preset Jumlah:</span>
              <button
                v-for="cnt in [1, 2, 3, 5, 10]"
                :key="cnt"
                type="button"
                @click="setPointCountPreset(cnt)"
                class="px-2 py-1 rounded border text-2xs font-semibold transition cursor-pointer"
                :class="points.length === cnt
                  ? 'bg-sky-500/15 text-sky-700 dark:text-sky-300 border-sky-400/50'
                  : 'bg-white dark:bg-[#0D2232] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#244356] hover:bg-slate-50'"
              >
                {{ cnt === 2 ? '2 Zero-Span' : cnt + ' Titik' }}
              </button>
            </div>
          </div>

          <!-- Points Data Table -->
          <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-[#244356]">
            <table class="w-full text-left border-collapse text-xs">
              <thead>
                <tr class="bg-slate-100 dark:bg-[#102436] text-slate-700 dark:text-slate-300 border-b border-slate-200 dark:border-[#244356] text-2xs uppercase tracking-wider">
                  <th class="py-2 px-2.5 w-8 text-center">#</th>
                  <th class="py-2 px-2 w-12 text-center" title="Matikan centang untuk mengabaikan titik outlier">Aktif</th>
                  <th class="py-2 px-2.5">Bacaan Sensor (x)</th>
                  <th class="py-2 px-2 w-28 text-center">Aksi Live</th>
                  <th class="py-2 px-2.5">Standar Referensi CRM (y)</th>
                  <th class="py-2 px-2.5 text-right">Estimasi (ŷ)</th>
                  <th class="py-2 px-2.5 text-right">Error</th>
                  <th class="py-2 px-2 w-8 text-center"></th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200 dark:divide-[#1E3648]">
                <tr
                  v-for="(pt, idx) in points"
                  :key="idx"
                  :class="!pt.active
                    ? 'bg-amber-50/70 dark:bg-amber-950/20 text-slate-400'
                    : (activeCaptureRow === idx ? 'bg-sky-50/80 dark:bg-sky-950/25' : 'bg-white dark:bg-[#0B1C2B]')"
                >
                  <td class="py-2 px-2.5 text-center font-bold tabular-nums">{{ idx + 1 }}</td>
                  <td class="py-2 px-2 text-center">
                    <input
                      type="checkbox"
                      v-model="pt.active"
                      class="w-4 h-4 rounded text-sky-600 focus:ring-sky-500 cursor-pointer"
                      :title="pt.active ? 'Titik aktif dalam regresi' : 'Titik diabaikan (Outlier)'"
                    />
                  </td>
                  <td class="py-1.5 px-2.5">
                    <input
                      type="number"
                      step="any"
                      v-model.number="pt.raw_x"
                      @focus="activeCaptureRow = idx"
                      placeholder="0.00"
                      class="w-full px-2 py-1 rounded-md border border-slate-200 dark:border-[#2A4B63] bg-white dark:bg-[#0D2232] text-xs font-mono tabular-nums text-right focus:outline-none"
                    />
                  </td>
                  <td class="py-1.5 px-2 text-center">
                    <button
                      type="button"
                      @click="captureLiveToRow(idx)"
                      class="px-2 py-1 rounded-md text-2xs font-semibold bg-amber-50 hover:bg-amber-100 dark:bg-amber-500/15 dark:hover:bg-amber-500/25 text-amber-800 dark:text-amber-300 border border-amber-300/70 dark:border-amber-500/30 transition cursor-pointer whitespace-nowrap"
                      title="Salin bacaan mentah sensor live (x) ke baris ini"
                    >
                      ⚡ Ambil Live
                    </button>
                  </td>
                  <td class="py-1.5 px-2.5">
                    <input
                      type="number"
                      step="any"
                      v-model.number="pt.ref_y"
                      placeholder="Standar CRM"
                      class="w-full px-2 py-1 rounded-md border border-slate-200 dark:border-[#2A4B63] bg-white dark:bg-[#0D2232] text-xs font-mono tabular-nums text-right focus:outline-none"
                    />
                  </td>
                  <td class="py-2 px-2.5 text-right font-mono tabular-nums font-semibold">
                    <span v-if="!pt.active" class="text-2xs font-sans italic text-amber-600 dark:text-amber-400">Outlier</span>
                    <span v-else>{{ formatRowPred(idx) }}</span>
                  </td>
                  <td class="py-2 px-2.5 text-right font-mono tabular-nums text-2xs font-semibold" :class="rowErrorClass(idx)">
                    {{ formatRowError(idx) }}
                  </td>
                  <td class="py-1.5 px-2 text-center">
                    <button
                      v-if="points.length > 1"
                      type="button"
                      @click="removePoint(idx)"
                      class="text-slate-400 hover:text-rose-600 dark:hover:text-rose-400 font-bold px-1 cursor-pointer"
                      title="Hapus titik ini"
                    >
                      ×
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Standard CRM / Span Gas Certificate Notes -->
          <div class="space-y-2 pt-1">
            <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
              Catatan Standar / No. Lot CRM / Tabung Gas Span:
            </label>
            <input
              type="text"
              id="input-calibration-standard-note"
              v-model="draftStandardNote"
              maxlength="240"
              placeholder="Contoh: Larutan Standar KHP Merck Lot #HC492810 / Gas Span SO2 400 ppb NIST"
              class="w-full px-3 py-2 rounded-lg border border-slate-200 dark:border-[#244356] bg-white dark:bg-[#0D2232] text-xs text-slate-800 dark:text-slate-100 focus:outline-none"
            />
            <label v-if="currentParam && currentParam.calibration_session_active" class="inline-flex items-center gap-2 text-2xs text-slate-600 dark:text-slate-300 cursor-pointer">
              <input type="checkbox" v-model="autoEndSessionOnApply" class="rounded text-sky-600" />
              <span>Nonaktifkan otomatis <strong>Mode Sesi Kalibrasi</strong> setelah kurva berhasil disimpan</span>
            </label>
          </div>
        </div>

        <!-- RIGHT COLUMN: Regression Curve SVG Chart, Per-Parameter Min R2 & Apply Controls -->
        <div class="saas-card p-4 xl:col-span-5 flex flex-col justify-between space-y-3.5">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <h2 class="text-sm font-bold text-slate-900 dark:text-white">
                Kurva Regresi &amp; Evaluasi Kelulusan
              </h2>
              <span v-if="currentParam" class="text-2xs font-semibold px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
                {{ domainFullLabel(draftDomain || currentParam.effective_domain) }}
              </span>
            </div>

            <!-- Interactive SVG Regression Curve Chart -->
            <div class="rounded-xl border border-slate-200 dark:border-[#244356] bg-slate-50/60 dark:bg-[#081724] p-2.5">
              <svg class="w-full h-60 chart-svg select-none" viewBox="0 0 420 240">
                <!-- Grid Lines -->
                <g class="grid-lines" stroke="#CBD5E1" stroke-opacity="0.45" stroke-width="1">
                  <line v-for="gy in chartGeometry.yTicks" :key="'gy-' + gy.pos" x1="54" :y1="gy.pos" x2="398" :y2="gy.pos" />
                  <line v-for="gx in chartGeometry.xTicks" :key="'gx-' + gx.pos" :x1="gx.pos" y1="18" :x2="gx.pos" y2="198" />
                </g>

                <!-- Axes -->
                <line x1="54" y1="18" x2="54" y2="198" stroke="#64748B" stroke-width="1.25" />
                <line x1="54" y1="198" x2="398" y2="198" stroke="#64748B" stroke-width="1.25" />

                <!-- 1:1 Ideal Reference Dashed Line (y = x) -->
                <line
                  v-if="chartGeometry.idealLine"
                  :x1="chartGeometry.idealLine.x1"
                  :y1="chartGeometry.idealLine.y1"
                  :x2="chartGeometry.idealLine.x2"
                  :y2="chartGeometry.idealLine.y2"
                  stroke="#94A3B8"
                  stroke-width="1.5"
                  stroke-dasharray="5 4"
                />

                <!-- Fitted Regression Curve Path -->
                <path
                  v-if="chartGeometry.curvePath"
                  :d="chartGeometry.curvePath"
                  fill="none"
                  stroke="#0284C7"
                  stroke-width="2.6"
                  stroke-linecap="round"
                />

                <!-- Live Current Reading Marker on Curve -->
                <g v-if="chartGeometry.liveMarker">
                  <circle
                    :cx="chartGeometry.liveMarker.cx"
                    :cy="chartGeometry.liveMarker.cy"
                    r="6"
                    fill="#10B981"
                    fill-opacity="0.22"
                  />
                  <circle
                    :cx="chartGeometry.liveMarker.cx"
                    :cy="chartGeometry.liveMarker.cy"
                    r="3.5"
                    fill="#10B981"
                    stroke="#FFFFFF"
                    stroke-width="1.2"
                  />
                </g>

                <!-- Calibration Standard Points (Active & Outlier) -->
                <g v-for="dot in chartGeometry.dots" :key="'dot-' + dot.index">
                  <circle
                    :cx="dot.cx"
                    :cy="dot.cy"
                    :r="dot.active ? 5 : 4.5"
                    :fill="dot.active ? '#0284C7' : '#F59E0B'"
                    :fill-opacity="dot.active ? 1 : 0.25"
                    :stroke="dot.active ? '#FFFFFF' : '#D97706'"
                    stroke-width="1.75"
                    class="chart-data-dot"
                  >
                    <title>Titik #{{ dot.index }}: x={{ dot.rawX }}, y={{ dot.refY }} {{ dot.active ? '' : '(Outlier)' }}</title>
                  </circle>
                </g>

                <!-- Y-Axis Tick Labels -->
                <g font-size="10" fill="#475569" text-anchor="end">
                  <text v-for="gy in chartGeometry.yTicks" :key="'ly-' + gy.pos" x="48" :y="gy.pos + 3.5">
                    {{ gy.label }}
                  </text>
                </g>

                <!-- X-Axis Tick Labels -->
                <g font-size="10" fill="#475569" text-anchor="middle">
                  <text v-for="gx in chartGeometry.xTicks" :key="'lx-' + gx.pos" :x="gx.pos" y="213">
                    {{ gx.label }}
                  </text>
                </g>

                <!-- Axis Titles -->
                <text x="226" y="232" font-size="10.5" font-weight="600" fill="#475569" text-anchor="middle">
                  Sumbu X: Bacaan Sensor {{ currentParam && currentParam.unit ? '(' + currentParam.unit + ')' : '' }}
                </text>
                <text
                  x="14"
                  y="108"
                  font-size="10.5"
                  font-weight="600"
                  fill="#475569"
                  text-anchor="middle"
                  transform="rotate(-90 14 108)"
                >
                  Sumbu Y: Standar Referensi
                </text>
              </svg>
            </div>

            <!-- Per-Parameter Minimum R² Setting (User Requirement #2: diatur per parameter) -->
            <div class="p-3 rounded-xl border border-slate-200 dark:border-[#244356] bg-slate-50/80 dark:bg-[#0D2232] space-y-2">
              <div class="flex flex-wrap items-center justify-between gap-2">
                <div>
                  <div class="text-xs font-bold text-slate-800 dark:text-slate-100">
                    Batas Lulus Skor R² Parameter Ini
                  </div>
                  <div class="text-2xs text-slate-500 dark:text-slate-400">
                    Disimpan khusus per parameter (default 0.9950)
                  </div>
                </div>

                <div class="flex items-center gap-1.5">
                  <input
                    type="number"
                    id="input-param-min-r2"
                    v-model.number="draftMinR2"
                    min="0.5"
                    max="1.0"
                    step="0.0005"
                    class="w-24 px-2 py-1 rounded-lg border border-slate-300 dark:border-[#2A4B63] bg-white dark:bg-[#081724] text-xs font-mono font-bold tabular-nums text-right focus:outline-none"
                  />
                  <button
                    type="button"
                    id="btn-save-param-min-r2"
                    :disabled="!currentParam || savingConfig"
                    @click="saveParameterMinR2"
                    class="px-2.5 py-1 rounded-lg text-2xs font-semibold bg-slate-200 hover:bg-slate-300 dark:bg-slate-700 dark:hover:bg-slate-600 text-slate-800 dark:text-slate-100 transition cursor-pointer disabled:opacity-50"
                  >
                    {{ savingConfig ? '...' : 'Simpan Batas' }}
                  </button>
                </div>
              </div>

              <div class="flex flex-wrap items-center gap-1 text-2xs">
                <span class="text-slate-500 dark:text-slate-400 mr-1">Preset R²:</span>
                <button
                  v-for="presetR2 in [0.99, 0.995, 0.998, 0.999]"
                  :key="presetR2"
                  type="button"
                  @click="draftMinR2 = presetR2"
                  class="px-2 py-0.5 rounded border font-mono transition cursor-pointer"
                  :class="Math.abs(Number(draftMinR2) - presetR2) < 0.00005
                    ? 'bg-sky-500/15 text-sky-700 dark:text-sky-300 border-sky-400/50 font-bold'
                    : 'bg-white dark:bg-[#081724] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#244356]'"
                >
                  ≥ {{ presetR2.toFixed(3) }}
                </button>
              </div>
            </div>

            <!-- Regression Equation & Pass/Fail Result Card -->
            <div class="p-3.5 rounded-xl border border-slate-200 dark:border-[#244356] bg-white dark:bg-[#0D2232] space-y-2.5">
              <div v-if="!regressionPreview.valid" class="text-xs font-medium text-amber-700 dark:text-amber-300">
                ⚠ {{ regressionPreview.error }}
              </div>
              <template v-else>
                <div class="flex flex-wrap items-start justify-between gap-2">
                  <div>
                    <div class="text-base sm:text-lg font-extrabold font-mono text-slate-900 dark:text-white tracking-tight">
                      {{ regressionPreview.equation }}
                    </div>
                    <div class="flex flex-wrap items-center gap-2 mt-1.5">
                      <template v-if="regressionPreview.r2 !== null">
                        <span class="text-xs font-mono font-bold text-slate-700 dark:text-slate-200">
                          R² = {{ regressionPreview.r2.toFixed(5) }}
                        </span>
                        <span
                          class="px-2 py-0.5 rounded-full text-2xs font-bold"
                          :class="regressionPreview.passed_r2
                            ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30'
                            : 'bg-rose-500/15 text-rose-700 dark:text-rose-300 border border-rose-500/30'"
                        >
                          {{ regressionPreview.passed_r2
                            ? '✓ LULUS (≥ ' + regressionPreview.min_r2.toFixed(4) + ')'
                            : '✕ TIDAK LULUS (< ' + regressionPreview.min_r2.toFixed(4) + ')' }}
                        </span>
                      </template>
                      <span
                        v-else
                        class="px-2 py-0.5 rounded-full text-2xs font-bold bg-sky-500/15 text-sky-700 dark:text-sky-300 border border-sky-500/30"
                      >
                        ✓ KALIBRASI 1-TITIK SIAP
                      </span>
                    </div>
                  </div>

                  <div class="text-right text-2xs font-mono space-y-0.5 text-slate-600 dark:text-slate-300">
                    <div v-if="draftModel === 'QUADRATIC'">Orde-2 (a): {{ regressionPreview.c2.toFixed(5) }}</div>
                    <div>Slope: {{ regressionPreview.c1.toFixed(4) }}</div>
                    <div>Offset: {{ regressionPreview.c0.toFixed(4) }}</div>
                    <div>RMSE: {{ regressionPreview.rmse.toFixed(4) }} {{ (currentParam && currentParam.unit) || '' }}</div>
                  </div>
                </div>

                <div v-if="regressionPreview.advisory_message" class="text-2xs text-sky-700 dark:text-sky-300 bg-sky-50 dark:bg-sky-950/30 px-2.5 py-1.5 rounded-lg border border-sky-200 dark:border-sky-800/50">
                  ℹ {{ regressionPreview.advisory_message }}
                </div>

                <!-- Explicit Force Override Option if R2 < MinR2 -->
                <label
                  v-if="regressionPreview.r2 !== null && !regressionPreview.passed_r2"
                  class="flex items-center gap-2 text-2xs text-rose-700 dark:text-rose-300 bg-rose-50 dark:bg-rose-950/30 p-2 rounded-lg border border-rose-200 dark:border-rose-800/50 cursor-pointer"
                >
                  <input type="checkbox" v-model="forceOverrideLowR2" class="rounded text-rose-600" />
                  <span>Tetap terapkan kurva meskipun R² di bawah batas lulus (wajib sertakan alasan di catatan CRM)</span>
                </label>
              </template>
            </div>
          </div>

          <!-- Bottom Action Buttons -->
          <div class="flex flex-wrap items-center justify-between gap-2 pt-2 border-t border-slate-100 dark:border-[#1E3648]">
            <button
              type="button"
              id="btn-reset-calibration"
              :disabled="!currentParam || applyingCalibration"
              @click="resetToDefaultCurve"
              class="px-3.5 py-2 rounded-lg text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 transition cursor-pointer disabled:opacity-50"
            >
              Reset ke Default (y = x)
            </button>

            <button
              type="button"
              id="btn-apply-calibration"
              :disabled="!canApplyCalibration || applyingCalibration"
              @click="applyCalibration"
              class="px-4 py-2 rounded-lg text-xs font-bold bg-blue-600 hover:bg-blue-700 text-white btn-text shadow-sm transition cursor-pointer disabled:opacity-50"
            >
              {{ applyingCalibration ? 'Menyimpan...' : '✓ Simpan & Terapkan Kalibrasi' }}
            </button>
          </div>
        </div>
      </div>

      <!-- Compact Audit History Preview Below Workspace -->
      <div class="saas-card p-4 space-y-3">
        <div class="flex flex-wrap items-center justify-between gap-2">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white">
            Riwayat &amp; Audit Log Kalibrasi Terkini
          </h3>
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="exportAllLogsCSV"
              :disabled="auditLogs.length === 0"
              class="px-2.5 py-1 rounded-lg text-2xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 transition cursor-pointer disabled:opacity-40"
            >
              📊 Export CSV
            </button>
            <button
              type="button"
              @click="switchToHistoryTab"
              class="text-xs font-semibold text-sky-600 dark:text-sky-400 hover:underline cursor-pointer"
            >
              Lihat Semua ({{ totalLogs }}) →
            </button>
          </div>
        </div>

        <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-[#244356]">
          <table class="w-full text-left border-collapse text-xs">
            <thead>
              <tr class="bg-slate-100 dark:bg-[#102436] text-slate-700 dark:text-slate-300 border-b border-slate-200 dark:border-[#244356] text-2xs uppercase tracking-wider">
                <th class="py-2 px-3">Waktu</th>
                <th class="py-2 px-2.5">Domain</th>
                <th class="py-2 px-3">Parameter</th>
                <th class="py-2 px-2.5">Metode &amp; Titik</th>
                <th class="py-2 px-3">Persamaan Baru</th>
                <th class="py-2 px-2.5 text-center">Skor R²</th>
                <th class="py-2 px-3">Teknisi (Role)</th>
                <th class="py-2 px-3">Catatan CRM / Gas</th>
                <th class="py-2 px-3 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200 dark:divide-[#1E3648]">
              <tr v-if="auditLogs.length === 0">
                <td colspan="9" class="py-6 text-center text-xs text-slate-400">
                  Belum ada riwayat kalibrasi tercatat.
                </td>
              </tr>
              <tr v-for="log in auditLogs.slice(0, 6)" :key="log.id" class="bg-white dark:bg-[#0B1C2B]">
                <td class="py-2 px-3 whitespace-nowrap tabular-nums text-slate-600 dark:text-slate-300">{{ formatTimestamp(log.created_at) }}</td>
                <td class="py-2 px-2.5">
                  <span class="px-2 py-0.5 rounded text-3xs font-bold border" :class="domainBadgeClass(log.domain)">
                    {{ domainShortBadge(log.domain) }}
                  </span>
                </td>
                <td class="py-2 px-3 font-semibold text-slate-800 dark:text-slate-100">
                  {{ log.parameter_name }} <span v-if="log.unit" class="text-slate-400 font-normal">({{ log.unit }})</span>
                </td>
                <td class="py-2 px-2.5 whitespace-nowrap">{{ formatLogMethod(log) }}</td>
                <td class="py-2 px-3 font-mono font-semibold text-slate-800 dark:text-slate-200">{{ log.equation }}</td>
                <td class="py-2 px-2.5 text-center">
                  <span
                    v-if="log.r2 !== null && log.r2 !== undefined"
                    class="px-2 py-0.5 rounded text-2xs font-mono font-bold"
                    :class="log.passed_r2
                      ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30'
                      : 'bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/30'"
                  >
                    {{ Number(log.r2).toFixed(4) }}
                  </span>
                  <span v-else class="px-2 py-0.5 rounded text-2xs font-mono bg-sky-500/10 text-sky-700 dark:text-sky-300 border border-sky-500/20">
                    {{ log.action === 'RESET' ? 'Default' : '1-Titik' }}
                  </span>
                </td>
                <td class="py-2 px-3 whitespace-nowrap">{{ log.performed_by }} <span class="text-2xs text-slate-400">({{ log.performed_role }})</span></td>
                <td class="py-2 px-3 max-w-[220px] truncate" :title="log.standard_note">{{ log.standard_note || '-' }}</td>
                <td class="py-2 px-3 text-right whitespace-nowrap space-x-1">
                  <button
                    type="button"
                    @click="printCalibrationCertificate(log)"
                    class="px-2 py-0.5 rounded border border-slate-200 dark:border-slate-700 bg-slate-50 hover:bg-slate-100 dark:bg-slate-800 text-2xs font-semibold cursor-pointer"
                    title="Cetak / Simpan PDF Berita Acara Kalibrasi"
                  >
                    📄 PDF
                  </button>
                  <button
                    type="button"
                    @click="rollbackCalibrationLog(log)"
                    class="px-2 py-0.5 rounded border border-amber-300/70 dark:border-amber-600/40 bg-amber-50 hover:bg-amber-100 dark:bg-amber-500/15 text-amber-800 dark:text-amber-300 text-2xs font-semibold cursor-pointer"
                    title="Kembalikan kurva kalibrasi dari riwayat ini"
                  >
                    ↩ Rollback
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ========================================================= -->
    <!-- TAB 2: RIWAYAT & LOG AUDIT PENUH                          -->
    <!-- ========================================================= -->
    <div v-else class="saas-card p-4 space-y-3.5">
      <div class="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 class="text-base font-bold text-slate-900 dark:text-white">
            Jejak Audit &amp; Berita Acara Kalibrasi Sensor
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            Rekaman lengkap koefisien kurva lama vs baru, titik standar CRM/Gas, skor R² terhadap batas parameter, dan rollback 1-klik.
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <label class="inline-flex items-center gap-1.5 text-xs text-slate-600 dark:text-slate-300 cursor-pointer">
            <input type="checkbox" v-model="historyOnlySelectedParam" @change="fetchAuditLogs" class="rounded text-sky-600" />
            <span>Hanya Parameter Terpilih</span>
          </label>
          <button
            type="button"
            @click="exportAllLogsCSV"
            :disabled="auditLogs.length === 0"
            class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 transition cursor-pointer disabled:opacity-40"
          >
            📊 Export CSV Riwayat
          </button>
          <button
            type="button"
            @click="fetchAuditLogs"
            class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-blue-600 hover:bg-blue-700 text-white btn-text transition cursor-pointer"
          >
            🔄 Segarkan
          </button>
        </div>
      </div>

      <div class="overflow-x-auto rounded-xl border border-slate-200 dark:border-[#244356]">
        <table class="w-full text-left border-collapse text-xs">
          <thead>
            <tr class="bg-slate-100 dark:bg-[#102436] text-slate-700 dark:text-slate-300 border-b border-slate-200 dark:border-[#244356] text-2xs uppercase tracking-wider">
              <th class="py-2.5 px-3">Waktu</th>
              <th class="py-2.5 px-2.5">Domain</th>
              <th class="py-2.5 px-3">Perangkat &amp; Parameter</th>
              <th class="py-2.5 px-2.5">Aksi / Metode</th>
              <th class="py-2.5 px-3">Persamaan Baru (vs Lama)</th>
              <th class="py-2.5 px-2.5 text-center">R² (Min)</th>
              <th class="py-2.5 px-3">Teknisi</th>
              <th class="py-2.5 px-3">Catatan Standar CRM</th>
              <th class="py-2.5 px-3 text-right">Aksi</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-200 dark:divide-[#1E3648]">
            <tr v-if="auditLogs.length === 0">
              <td colspan="9" class="py-8 text-center text-xs text-slate-400">
                Tidak ada log kalibrasi untuk filter yang dipilih.
              </td>
            </tr>
            <tr v-for="log in auditLogs" :key="'full-' + log.id" class="bg-white dark:bg-[#0B1C2B]">
              <td class="py-2.5 px-3 whitespace-nowrap tabular-nums">{{ formatTimestamp(log.created_at) }}</td>
              <td class="py-2.5 px-2.5">
                <span class="px-2 py-0.5 rounded text-3xs font-bold border" :class="domainBadgeClass(log.domain)">
                  {{ domainShortBadge(log.domain) }}
                </span>
              </td>
              <td class="py-2.5 px-3">
                <div class="font-bold text-slate-800 dark:text-slate-100">{{ log.parameter_name }} ({{ log.unit || '-' }})</div>
                <div class="text-2xs text-slate-400">{{ log.device_name }} · {{ log.parameter_code }}</div>
              </td>
              <td class="py-2.5 px-2.5 whitespace-nowrap">
                <div class="font-semibold">{{ formatLogMethod(log) }}</div>
                <div class="text-2xs text-slate-400">{{ log.action }}</div>
              </td>
              <td class="py-2.5 px-3 font-mono">
                <div class="font-bold text-slate-900 dark:text-white">{{ log.equation }}</div>
                <div v-if="log.prev_equation" class="text-2xs text-slate-400">Sebelumnya: {{ log.prev_equation }}</div>
              </td>
              <td class="py-2.5 px-2.5 text-center font-mono">
                <div v-if="log.r2 !== null && log.r2 !== undefined" class="font-bold" :class="log.passed_r2 ? 'text-emerald-600 dark:text-emerald-400' : 'text-amber-600 dark:text-amber-400'">
                  {{ Number(log.r2).toFixed(5) }}
                </div>
                <div v-else class="text-slate-400">—</div>
                <div class="text-3xs text-slate-400">Min {{ Number(log.min_r2 || 0.995).toFixed(3) }}</div>
              </td>
              <td class="py-2.5 px-3 whitespace-nowrap">
                <div class="font-semibold">{{ log.performed_by }}</div>
                <div class="text-2xs text-slate-400">{{ log.performed_role }}</div>
              </td>
              <td class="py-2.5 px-3 max-w-[260px] break-words">{{ log.standard_note || '-' }}</td>
              <td class="py-2.5 px-3 text-right whitespace-nowrap space-x-1">
                <button
                  type="button"
                  @click="printCalibrationCertificate(log)"
                  class="px-2 py-1 rounded border border-slate-200 dark:border-slate-700 bg-slate-50 hover:bg-slate-100 dark:bg-slate-800 text-2xs font-semibold cursor-pointer"
                >
                  📄 PDF
                </button>
                <button
                  type="button"
                  @click="exportSingleLogCSV(log)"
                  class="px-2 py-1 rounded border border-slate-200 dark:border-slate-700 bg-slate-50 hover:bg-slate-100 dark:bg-slate-800 text-2xs font-semibold cursor-pointer"
                >
                  📊 CSV
                </button>
                <button
                  type="button"
                  @click="rollbackCalibrationLog(log)"
                  class="px-2 py-1 rounded border border-amber-300/70 dark:border-amber-600/40 bg-amber-50 hover:bg-amber-100 dark:bg-amber-500/15 text-amber-800 dark:text-amber-300 text-2xs font-semibold cursor-pointer"
                >
                  ↩ Rollback
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import api from '../services/api';
import {
  CALIBRATION_MODELS,
  CALIBRATION_DOMAINS,
  calculateRegressionJS,
  evaluateCalibrationCurve,
  evaluateReadingStability,
  getDomainStandardPresets,
  normalizeMinR2,
} from '../utils/calibrationMath';

export default {
  name: 'SensorCalibrationView',
  data() {
    return {
      activeTab: 'workspace',
      domainOptions: CALIBRATION_DOMAINS,
      modelOptions: CALIBRATION_MODELS,
      selectedDomainFilter: 'ALL',
      allParameters: [],
      stationDomain: 'SPARING',

      selectedDeviceId: 0,
      selectedParameterId: 0,
      activeCaptureRow: 0,

      draftModel: 'LINEAR',
      draftDomain: 'SPARING',
      draftMinR2: 0.995,
      draftStandardNote: '',
      selectedStandardPresetId: '',
      autoEndSessionOnApply: true,
      forceOverrideLowR2: false,

      points: [
        { active: true, raw_x: 0, ref_y: 0 },
        { active: true, raw_x: 25, ref_y: 25 },
        { active: true, raw_x: 50, ref_y: 50 },
        { active: true, raw_x: 100, ref_y: 100 },
        { active: true, raw_x: 200, ref_y: 200 },
      ],

      liveX: null,
      liveCalibratedSaved: null,
      liveHistorySamples: [],
      pollingLive: false,
      liveTimer: null,

      togglingSession: false,
      savingConfig: false,
      applyingCalibration: false,

      auditLogs: [],
      totalLogs: 0,
      historyOnlySelectedParam: false,

      banner: {
        type: 'success',
        message: '',
      },
    };
  },
  computed: {
    deviceOptions() {
      const map = new Map();
      for (const p of this.allParameters) {
        if (!map.has(p.device_id)) {
          map.set(p.device_id, {
            device_id: p.device_id,
            device_code: p.device_code,
            device_name: p.device_name,
            protocol: p.protocol,
          });
        }
      }
      return Array.from(map.values());
    },
    filteredParameters() {
      return this.allParameters.filter(p => {
        if (this.selectedDeviceId > 0 && Number(p.device_id) !== Number(this.selectedDeviceId)) {
          return false;
        }
        return true;
      });
    },
    currentParam() {
      return this.allParameters.find(p => Number(p.parameter_id) === Number(this.selectedParameterId)) || null;
    },
    domainStandardPresets() {
      return getDomainStandardPresets(this.currentParam);
    },
    activePointsCount() {
      return this.points.filter(p => p.active !== false).length;
    },
    regressionPreview() {
      return calculateRegressionJS(this.draftModel, this.points, this.draftMinR2);
    },
    stabilityInfo() {
      return evaluateReadingStability(this.liveHistorySamples);
    },
    formattedLiveX() {
      if (this.liveX === null || this.liveX === undefined || !Number.isFinite(Number(this.liveX))) {
        return '—';
      }
      const prec = this.currentParam && this.currentParam.precision >= 0 ? Math.min(4, Math.max(2, this.currentParam.precision)) : 2;
      return Number(this.liveX).toFixed(prec);
    },
    formattedSimulatedY() {
      if (this.liveX === null || this.liveX === undefined || !Number.isFinite(Number(this.liveX))) {
        if (this.liveCalibratedSaved !== null && Number.isFinite(Number(this.liveCalibratedSaved))) {
          return Number(this.liveCalibratedSaved).toFixed(2);
        }
        return '—';
      }
      const prec = this.currentParam && this.currentParam.precision >= 0 ? Math.min(4, Math.max(2, this.currentParam.precision)) : 2;
      if (this.regressionPreview.valid) {
        const yHat = evaluateCalibrationCurve(
          true,
          this.regressionPreview.model,
          this.regressionPreview.c2,
          this.regressionPreview.c1,
          this.regressionPreview.c0,
          this.liveX
        );
        return Number(yHat).toFixed(prec);
      }
      if (this.currentParam && this.currentParam.calibration_enabled) {
        const ySaved = evaluateCalibrationCurve(
          true,
          this.currentParam.calibration_model,
          this.currentParam.cal_c2,
          this.currentParam.cal_c1,
          this.currentParam.cal_c0,
          this.liveX
        );
        return Number(ySaved).toFixed(prec);
      }
      return Number(this.liveX).toFixed(prec);
    },
    activeSavedEquation() {
      if (!this.currentParam) return 'y = x (Default)';
      return this.currentParam.cal_equation || 'y = x (Default)';
    },
    canApplyCalibration() {
      if (!this.currentParam || !this.regressionPreview.valid) return false;
      if (!this.regressionPreview.passed_r2 && !this.forceOverrideLowR2) return false;
      return true;
    },
    sparklinePath() {
      const vals = this.liveHistorySamples.filter(Number.isFinite);
      if (vals.length < 2) return '';
      const min = Math.min(...vals);
      const max = Math.max(...vals);
      const span = max - min || 1;
      const w = 96;
      const h = 20;
      return vals
        .map((v, idx) => {
          const x = (idx / (vals.length - 1)) * w;
          const y = 2 + h - ((v - min) / span) * h;
          return `${idx === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`;
        })
        .join(' ');
    },
    chartGeometry() {
      const padL = 54, padR = 398, padT = 18, padB = 198;
      const width = padR - padL;
      const height = padB - padT;

      const validPts = this.points
        .map((p, idx) => ({
          index: idx + 1,
          active: p.active !== false,
          rawX: Number(p.raw_x),
          refY: Number(p.ref_y),
        }))
        .filter(p => Number.isFinite(p.rawX) && Number.isFinite(p.refY));

      let minX = 0, maxX = 100, minY = 0, maxY = 100;
      if (validPts.length > 0) {
        minX = Math.min(...validPts.map(p => p.rawX));
        maxX = Math.max(...validPts.map(p => p.rawX));
        minY = Math.min(...validPts.map(p => p.refY));
        maxY = Math.max(...validPts.map(p => p.refY));
      }
      if (this.liveX !== null && Number.isFinite(Number(this.liveX))) {
        minX = Math.min(minX, Number(this.liveX));
        maxX = Math.max(maxX, Number(this.liveX));
      }
      if (Math.abs(maxX - minX) < 1e-6) {
        minX -= 5;
        maxX += 5;
      } else {
        const padX = (maxX - minX) * 0.08;
        minX = minX >= 0 && minX - padX < 0 ? 0 : minX - padX;
        maxX += padX;
      }
      if (Math.abs(maxY - minY) < 1e-6) {
        minY -= 5;
        maxY += 5;
      } else {
        const padY = (maxY - minY) * 0.08;
        minY = minY >= 0 && minY - padY < 0 ? 0 : minY - padY;
        maxY += padY;
      }

      const toSvgX = x => padL + ((x - minX) / (maxX - minX)) * width;
      const toSvgY = y => padB - ((y - minY) / (maxY - minY)) * height;
      const clampSvgY = y => Math.max(padT - 4, Math.min(padB + 4, toSvgY(y)));

      const xTicks = [];
      const yTicks = [];
      for (let i = 0; i <= 4; i++) {
        const xv = minX + ((maxX - minX) * i) / 4;
        const yv = minY + ((maxY - minY) * i) / 4;
        xTicks.push({ pos: padL + (width * i) / 4, label: Math.abs(xv) >= 100 ? xv.toFixed(0) : xv.toFixed(1) });
        yTicks.push({ pos: padB - (height * i) / 4, label: Math.abs(yv) >= 100 ? yv.toFixed(0) : yv.toFixed(1) });
      }

      const idealStart = Math.max(minX, minY);
      const idealEnd = Math.min(maxX, maxY);
      const idealLine = idealEnd > idealStart
        ? {
            x1: toSvgX(idealStart),
            y1: clampSvgY(idealStart),
            x2: toSvgX(idealEnd),
            y2: clampSvgY(idealEnd),
          }
        : null;

      let curvePath = '';
      if (this.regressionPreview.valid) {
        const steps = 40;
        const segs = [];
        for (let s = 0; s <= steps; s++) {
          const x = minX + ((maxX - minX) * s) / steps;
          const y = evaluateCalibrationCurve(
            true,
            this.regressionPreview.model,
            this.regressionPreview.c2,
            this.regressionPreview.c1,
            this.regressionPreview.c0,
            x
          );
          segs.push(`${s === 0 ? 'M' : 'L'}${toSvgX(x).toFixed(1)},${clampSvgY(y).toFixed(1)}`);
        }
        curvePath = segs.join(' ');
      }

      let liveMarker = null;
      if (this.liveX !== null && Number.isFinite(Number(this.liveX)) && this.regressionPreview.valid) {
        const lx = Number(this.liveX);
        const ly = evaluateCalibrationCurve(
          true,
          this.regressionPreview.model,
          this.regressionPreview.c2,
          this.regressionPreview.c1,
          this.regressionPreview.c0,
          lx
        );
        liveMarker = {
          cx: toSvgX(lx),
          cy: clampSvgY(ly),
        };
      }

      const dots = validPts.map(pt => ({
        ...pt,
        cx: toSvgX(pt.rawX),
        cy: clampSvgY(pt.refY),
      }));

      return { xTicks, yTicks, idealLine, curvePath, dots, liveMarker };
    },
  },
  mounted() {
    this.initializePage();
    this.liveTimer = setInterval(() => {
      if (this.activeTab === 'workspace' && this.selectedParameterId > 0) {
        this.pollLiveSensor(false);
      }
    }, 2500);
  },
  beforeDestroy() {
    if (this.liveTimer) {
      clearInterval(this.liveTimer);
      this.liveTimer = null;
    }
  },
  methods: {
    showBanner(message, type = 'success') {
      this.banner = { message, type };
      if (type === 'success') {
        setTimeout(() => {
          if (this.banner.message === message) {
            this.banner.message = '';
          }
        }, 5000);
      }
    },
    domainShortBadge(d) {
      const norm = String(d || '').toUpperCase();
      if (norm === 'AQMS') return 'AQMS';
      if (norm === 'WQMS_ONLIMO' || norm === 'WQMS' || norm === 'ONLIMO') return 'ONLIMO';
      return 'SPARING';
    },
    domainFullLabel(d) {
      const norm = String(d || '').toUpperCase();
      if (norm === 'AQMS') return 'AQMS · Udara & ISPU';
      if (norm === 'WQMS_ONLIMO' || norm === 'WQMS') return 'WQMS ONLIMO · Mutu Air';
      return 'SPARING · Air Limbah';
    },
    domainBadgeClass(d) {
      const norm = String(d || '').toUpperCase();
      if (norm === 'AQMS') {
        return 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border-cyan-500/30';
      }
      if (norm === 'WQMS_ONLIMO' || norm === 'WQMS') {
        return 'bg-teal-500/10 text-teal-700 dark:text-teal-300 border-teal-500/30';
      }
      return 'bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border-indigo-500/30';
    },
    async initializePage() {
      const qDev = Number(this.$route.query.device_id || 0);
      const qParam = Number(this.$route.query.parameter_id || 0);
      if (qDev > 0) this.selectedDeviceId = qDev;
      await this.fetchParameters(qParam);
      await this.fetchAuditLogs();
    },
    async onSelectDomainFilter(domainVal) {
      this.selectedDomainFilter = domainVal;
      await this.fetchParameters(this.selectedParameterId);
      await this.fetchAuditLogs();
    },
    async fetchParameters(preferredParamId = 0) {
      try {
        const params = {};
        if (this.selectedDomainFilter && this.selectedDomainFilter !== 'ALL') {
          params.domain = this.selectedDomainFilter;
        }
        const res = await api.get('/calibrations/parameters', { params });
        const payload = (res.data && res.data.data) || {};
        this.allParameters = Array.isArray(payload.items) ? payload.items : [];
        if (payload.station_domain) {
          this.stationDomain = payload.station_domain;
        }

        if (this.selectedDeviceId > 0 && !this.allParameters.some(p => Number(p.device_id) === Number(this.selectedDeviceId))) {
          this.selectedDeviceId = 0;
        }

        const list = this.filteredParameters;
        if (list.length === 0) {
          this.selectedParameterId = 0;
          this.liveX = null;
          return;
        }

        const targetId = preferredParamId || this.selectedParameterId;
        const found = list.find(p => Number(p.parameter_id) === Number(targetId));
        this.selectedParameterId = found ? found.parameter_id : list[0].parameter_id;
        this.hydrateWorkspaceFromCurrentParam();
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.message) || 'Gagal memuat daftar parameter kalibrasi.';
        this.showBanner(msg, 'error');
      }
    },
    onDeviceChange() {
      const list = this.filteredParameters;
      this.selectedParameterId = list.length > 0 ? list[0].parameter_id : 0;
      this.hydrateWorkspaceFromCurrentParam();
    },
    onParameterChange() {
      this.hydrateWorkspaceFromCurrentParam();
      if (this.historyOnlySelectedParam) {
        this.fetchAuditLogs();
      }
    },
    hydrateWorkspaceFromCurrentParam() {
      const p = this.currentParam;
      this.selectedStandardPresetId = '';
      this.forceOverrideLowR2 = false;
      this.liveHistorySamples = [];
      if (!p) {
        this.liveX = null;
        this.liveCalibratedSaved = null;
        return;
      }

      this.draftDomain = p.effective_domain || 'SPARING';
      this.draftMinR2 = normalizeMinR2(p.cal_min_r2);
      this.draftStandardNote = p.cal_standard_note || '';

      const savedModel = String(p.calibration_model || 'LINEAR').toUpperCase();
      this.draftModel = ['LINEAR', 'QUADRATIC', 'OFFSET_1PT', 'SPAN_1PT'].includes(savedModel) ? savedModel : 'LINEAR';

      if (Array.isArray(p.cal_points) && p.cal_points.length >= 1) {
        this.points = p.cal_points.slice(0, 10).map(pt => ({
          active: pt.active !== false,
          label: pt.label || '',
          raw_x: Number(pt.raw_x),
          ref_y: Number(pt.ref_y),
        }));
      } else {
        // Default clean 5-point template or parameter-smart preset
        const presets = getDomainStandardPresets(p);
        if (presets.length > 0) {
          this.draftModel = presets[0].model;
          this.points = presets[0].refs.map(r => ({ active: true, raw_x: r, ref_y: r }));
        } else {
          this.draftModel = 'LINEAR';
          this.points = [0, 25, 50, 100, 200].map(v => ({ active: true, raw_x: v, ref_y: v }));
        }
      }

      if (p.live_uncalibrated_value !== undefined && p.live_uncalibrated_value !== null) {
        this.liveX = Number(p.live_uncalibrated_value);
        this.liveHistorySamples = [this.liveX];
      } else {
        this.liveX = null;
      }
      this.liveCalibratedSaved = p.live_calibrated_value !== undefined ? p.live_calibrated_value : null;
      this.pollLiveSensor(false);
    },
    async pollLiveSensor(pollHardware = false) {
      if (!this.currentParam) return;
      if (pollHardware) this.pollingLive = true;
      try {
        const res = await api.get(`/calibrations/parameters/${this.currentParam.parameter_id}/live`, {
          params: pollHardware ? { poll: 'true' } : {},
        });
        const d = (res.data && res.data.data) || {};
        if (d.uncalibrated_value !== null && d.uncalibrated_value !== undefined && Number.isFinite(Number(d.uncalibrated_value))) {
          const val = Number(d.uncalibrated_value);
          this.liveX = val;
          this.liveHistorySamples = [...this.liveHistorySamples.slice(-15), val];
        }
        if (d.calibrated_value !== null && d.calibrated_value !== undefined) {
          this.liveCalibratedSaved = Number(d.calibrated_value);
        }
      } catch (e) {
        // Silent during background poll; show banner only if manual click
        if (pollHardware) {
          this.showBanner('Gagal membaca sensor secara langsung. Pastikan koneksi perangkat aktif.', 'error');
        }
      } finally {
        if (pollHardware) this.pollingLive = false;
      }
    },
    selectModel(m) {
      this.draftModel = m;
      if (m === 'QUADRATIC' && this.points.length < 3) {
        while (this.points.length < 4) {
          const nextVal = this.points.length * 25;
          this.points.push({ active: true, raw_x: nextVal, ref_y: nextVal });
        }
      } else if (m === 'LINEAR' && this.points.length < 2) {
        this.points.push({ active: true, raw_x: 100, ref_y: 100 });
      }
    },
    addPoint() {
      if (this.points.length >= 10) return;
      const last = this.points[this.points.length - 1];
      const nextRef = last && Number.isFinite(Number(last.ref_y)) ? Number((Number(last.ref_y) + 25).toFixed(2)) : this.points.length * 25;
      this.points.push({ active: true, raw_x: nextRef, ref_y: nextRef });
    },
    removePoint(idx) {
      if (this.points.length <= 1) return;
      this.points.splice(idx, 1);
    },
    setPointCountPreset(count) {
      if (count === 1 && (this.draftModel === 'LINEAR' || this.draftModel === 'QUADRATIC')) {
        this.draftModel = 'OFFSET_1PT';
      } else if (count === 2 && this.draftModel === 'QUADRATIC') {
        this.draftModel = 'LINEAR';
      } else if (count >= 2 && (this.draftModel === 'OFFSET_1PT' || this.draftModel === 'SPAN_1PT')) {
        this.draftModel = 'LINEAR';
      }

      const maxSpan = 100;
      const nextPts = [];
      for (let i = 0; i < count; i++) {
        const existing = this.points[i];
        if (existing) {
          nextPts.push({ ...existing, active: true });
        } else {
          const v = count === 1 ? 25 : Number(((maxSpan * i) / (count - 1)).toFixed(2));
          nextPts.push({ active: true, raw_x: v, ref_y: v });
        }
      }
      this.points = nextPts;
    },
    applyDomainStandardPreset() {
      const preset = this.domainStandardPresets.find(x => x.id === this.selectedStandardPresetId);
      if (!preset) return;
      this.draftModel = preset.model;
      if (!this.draftStandardNote) {
        this.draftStandardNote = preset.note;
      }
      this.points = preset.refs.slice(0, 10).map((refVal, idx) => {
        const prevX = this.points[idx] && Number.isFinite(Number(this.points[idx].raw_x)) ? Number(this.points[idx].raw_x) : refVal;
        return {
          active: true,
          raw_x: prevX,
          ref_y: refVal,
        };
      });
    },
    captureLiveToRow(idx) {
      this.activeCaptureRow = idx;
      if (this.liveX === null || !Number.isFinite(Number(this.liveX))) {
        this.showBanner('Belum ada bacaan mentah sensor live (x). Klik "Baca Live" atau ketik manual.', 'error');
        return;
      }
      const prec = this.currentParam && this.currentParam.precision >= 0 ? Math.min(4, Math.max(2, this.currentParam.precision)) : 2;
      this.$set(this.points[idx], 'raw_x', Number(Number(this.liveX).toFixed(prec)));
    },
    formatRowPred(idx) {
      if (!this.regressionPreview.valid || !this.regressionPreview.points[idx]) return '—';
      return Number(this.regressionPreview.points[idx].pred_y).toFixed(2);
    },
    formatRowError(idx) {
      if (!this.regressionPreview.valid || !this.regressionPreview.points[idx]) return '—';
      const pt = this.regressionPreview.points[idx];
      if (pt.error_pct !== null && pt.error_pct !== undefined) {
        const sign = pt.error_pct > 0 ? '+' : '';
        return `${sign}${pt.error_pct.toFixed(2)}%`;
      }
      const sign = pt.error_abs > 0 ? '+' : '';
      const unit = (this.currentParam && this.currentParam.unit) || '';
      return `${sign}${pt.error_abs.toFixed(2)} ${unit}`.trim();
    },
    rowErrorClass(idx) {
      if (!this.regressionPreview.valid || !this.regressionPreview.points[idx]) return 'text-slate-400';
      const pt = this.regressionPreview.points[idx];
      if (!pt.active) return 'text-amber-600 dark:text-amber-400';
      const pctAbs = pt.error_pct !== null ? Math.abs(pt.error_pct) : Math.abs(pt.error_abs);
      if (pctAbs <= 2.0) return 'text-emerald-600 dark:text-emerald-400';
      if (pctAbs <= 5.0) return 'text-amber-600 dark:text-amber-400';
      return 'text-rose-600 dark:text-rose-400';
    },
    replaceUpdatedParameterInList(updatedDto) {
      if (!updatedDto) return;
      const idx = this.allParameters.findIndex(p => Number(p.parameter_id) === Number(updatedDto.parameter_id));
      if (idx >= 0) {
        this.$set(this.allParameters, idx, updatedDto);
      }
    },
    async toggleCalibrationSession() {
      if (!this.currentParam) return;
      this.togglingSession = true;
      const desired = !this.currentParam.calibration_session_active;
      try {
        const res = await api.put(`/calibrations/parameters/${this.currentParam.parameter_id}/session`, {
          active: desired,
        });
        const dto = res.data && res.data.data;
        this.replaceUpdatedParameterInList(dto);
        this.showBanner(
          desired
            ? `Mode Sesi Kalibrasi AKTIF untuk ${this.currentParam.parameter_name}. Alarm & agregasi kepatuhan ditahan sementara.`
            : `Mode Sesi Kalibrasi NONAKTIF untuk ${this.currentParam.parameter_name}. Pemantauan normal dilanjutkan.`
        );
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.message) || 'Gagal mengubah status Mode Sesi Kalibrasi.';
        this.showBanner(msg, 'error');
      } finally {
        this.togglingSession = false;
      }
    },
    async saveParameterMinR2() {
      if (!this.currentParam) return;
      const v = Number(this.draftMinR2);
      if (!Number.isFinite(v) || v < 0.5 || v > 1.0) {
        this.showBanner('Batas minimum R² harus berada di rentang 0.5000 hingga 1.0000.', 'error');
        return;
      }
      this.savingConfig = true;
      try {
        const res = await api.put(`/calibrations/parameters/${this.currentParam.parameter_id}/config`, {
          cal_min_r2: v,
          domain: this.draftDomain,
        });
        this.replaceUpdatedParameterInList(res.data && res.data.data);
        this.showBanner(`Batas lulus R² untuk ${this.currentParam.parameter_name} disimpan (≥ ${v.toFixed(4)}).`);
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.message) || 'Gagal menyimpan batas minimum R².';
        this.showBanner(msg, 'error');
      } finally {
        this.savingConfig = false;
      }
    },
    async saveParameterDomainOnly() {
      if (!this.currentParam) return;
      try {
        const res = await api.put(`/calibrations/parameters/${this.currentParam.parameter_id}/config`, {
          domain: this.draftDomain,
        });
        this.replaceUpdatedParameterInList(res.data && res.data.data);
      } catch (e) {
        // ignore
      }
    },
    async applyCalibration() {
      if (!this.currentParam || !this.regressionPreview.valid) return;
      if (!this.regressionPreview.passed_r2 && this.forceOverrideLowR2 && !String(this.draftStandardNote || '').trim()) {
        this.showBanner('Harap isi alasan teknis pada kolom Catatan Standar sebelum menerapkan kurva dengan R² di bawah batas.', 'error');
        return;
      }
      this.applyingCalibration = true;
      try {
        const res = await api.post(`/calibrations/parameters/${this.currentParam.parameter_id}/apply`, {
          model: this.draftModel,
          domain: this.draftDomain,
          points: this.points,
          min_r2: Number(this.draftMinR2),
          standard_note: this.draftStandardNote,
          end_session: Boolean(this.autoEndSessionOnApply && this.currentParam.calibration_session_active),
          force_override: Boolean(this.forceOverrideLowR2),
        });
        const payload = (res.data && res.data.data) || {};
        this.replaceUpdatedParameterInList(payload.parameter);
        this.forceOverrideLowR2 = false;
        await this.fetchAuditLogs();
        this.showBanner(`Kalibrasi ${this.currentParam.parameter_name} berhasil diterapkan: ${payload.parameter.cal_equation}`);
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.message) || 'Gagal menyimpan kurva kalibrasi.';
        this.showBanner(msg, 'error');
      } finally {
        this.applyingCalibration = false;
      }
    },
    async resetToDefaultCurve() {
      if (!this.currentParam) return;
      if (!window.confirm(`Reset kurva kalibrasi "${this.currentParam.parameter_name}" ke default (y = x)?\n\nCatatan: Skala & Offset register Modbus dasar tidak akan berubah.`)) {
        return;
      }
      this.applyingCalibration = true;
      try {
        const res = await api.post(`/calibrations/parameters/${this.currentParam.parameter_id}/reset`, {
          reason_note: this.draftStandardNote || 'Reset manual ke kurva default (y = x)',
        });
        const payload = (res.data && res.data.data) || {};
        this.replaceUpdatedParameterInList(payload.parameter);
        await this.fetchAuditLogs();
        this.showBanner(`Kurva kalibrasi ${this.currentParam.parameter_name} telah dikembalikan ke default (y = x).`);
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.message) || 'Gagal mereset kurva kalibrasi.';
        this.showBanner(msg, 'error');
      } finally {
        this.applyingCalibration = false;
      }
    },
    switchToHistoryTab() {
      this.activeTab = 'history';
      this.fetchAuditLogs();
    },
    async fetchAuditLogs() {
      try {
        const params = { page: 1, page_size: 100 };
        if (this.selectedDomainFilter && this.selectedDomainFilter !== 'ALL') {
          params.domain = this.selectedDomainFilter;
        }
        if (this.historyOnlySelectedParam && this.selectedParameterId > 0) {
          params.parameter_id = this.selectedParameterId;
        }
        const res = await api.get('/calibrations/logs', { params });
        const payload = (res.data && res.data.data) || {};
        this.auditLogs = Array.isArray(payload.items) ? payload.items : [];
        this.totalLogs = Number(payload.total || this.auditLogs.length);
      } catch (e) {
        // non-blocking
      }
    },
    async rollbackCalibrationLog(log) {
      if (!log) return;
      if (!window.confirm(`Rollback kalibrasi parameter "${log.parameter_name}" berdasarkan log #${log.id}?`)) {
        return;
      }
      try {
        const res = await api.post(`/calibrations/logs/${log.id}/rollback`);
        const payload = (res.data && res.data.data) || {};
        this.replaceUpdatedParameterInList(payload.parameter);
        if (this.currentParam && Number(this.currentParam.parameter_id) === Number(log.parameter_id)) {
          this.hydrateWorkspaceFromCurrentParam();
        }
        await this.fetchAuditLogs();
        this.showBanner(`Rollback kalibrasi ${log.parameter_name} berhasil diterapkan (${payload.parameter.cal_equation}).`);
      } catch (err) {
        const msg = (err.response && err.response.data && err.response.data.message) || 'Gagal melakukan rollback kalibrasi.';
        this.showBanner(msg, 'error');
      }
    },
    formatTimestamp(ts) {
      if (!ts) return '-';
      const d = new Date(ts);
      if (Number.isNaN(d.getTime())) return String(ts);
      return d.toLocaleString('id-ID', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    },
    formatLogMethod(log) {
      if (!log) return '-';
      if (log.action === 'RESET') return 'Reset Default';
      const m = String(log.calibration_model || '').toUpperCase();
      const label = m === 'QUADRATIC' ? 'Kuadratik' : m === 'OFFSET_1PT' ? '1-Titik Offset' : m === 'SPAN_1PT' ? '1-Titik Span' : 'Linier';
      return log.point_count > 0 ? `${label} · ${log.point_count} Titik` : label;
    },
    downloadCSVContent(filename, csvContent) {
      const blob = new Blob(['\uFEFF' + csvContent], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    },
    exportAllLogsCSV() {
      const headers = ['ID', 'Waktu', 'Domain', 'Perangkat', 'Parameter', 'Satuan', 'Aksi', 'Model', 'Jumlah Titik', 'Persamaan Baru', 'Persamaan Sebelumnya', 'R2', 'Min R2', 'Lulus R2', 'RMSE', 'Teknisi', 'Role', 'Catatan Standar'];
      const rows = this.auditLogs.map(l => [
        l.id,
        this.formatTimestamp(l.created_at),
        l.domain,
        l.device_name,
        l.parameter_name,
        l.unit || '',
        l.action,
        l.calibration_model,
        l.point_count,
        l.equation,
        l.prev_equation || '',
        l.r2 !== null && l.r2 !== undefined ? Number(l.r2).toFixed(5) : '',
        Number(l.min_r2 || 0.995).toFixed(4),
        l.passed_r2 ? 'YA' : 'TIDAK',
        l.rmse !== null && l.rmse !== undefined ? Number(l.rmse).toFixed(4) : '',
        l.performed_by,
        l.performed_role,
        String(l.standard_note || '').replace(/"/g, '""'),
      ]);
      const csv = [headers, ...rows].map(r => r.map(v => `"${v}"`).join(',')).join('\n');
      this.downloadCSVContent(`riwayat_kalibrasi_sensor_${Date.now()}.csv`, csv);
    },
    exportSingleLogCSV(log) {
      let pts = [];
      try {
        pts = log.points_json ? JSON.parse(log.points_json) : [];
      } catch (e) {
        pts = [];
      }
      const meta = [
        ['Berita Acara Kalibrasi Sensor', log.domain || 'SPARING'],
        ['Waktu Kalibrasi', this.formatTimestamp(log.created_at)],
        ['Perangkat', `${log.device_name} (${log.device_code})`],
        ['Parameter', `${log.parameter_name} (${log.unit || '-'})`],
        ['Metode & Titik', this.formatLogMethod(log)],
        ['Persamaan Baru', log.equation],
        ['Skor R2', log.r2 !== null && log.r2 !== undefined ? Number(log.r2).toFixed(5) : 'N/A'],
        ['Batas Lulus R2 Parameter', Number(log.min_r2 || 0.995).toFixed(4)],
        ['Teknisi', `${log.performed_by} (${log.performed_role})`],
        ['Catatan CRM / Gas Standar', log.standard_note || '-'],
        [],
        ['Titik #', 'Status', 'Bacaan Sensor (x)', 'Standar Referensi (y)', 'Estimasi Kurva (y_hat)', 'Error Absolut', 'Error (%)'],
      ];
      const ptRows = pts.map(p => [
        p.index,
        p.active ? 'AKTIF' : 'OUTLIER',
        p.raw_x,
        p.ref_y,
        Number(p.pred_y || 0).toFixed(4),
        Number(p.error_abs || 0).toFixed(4),
        p.error_pct !== null && p.error_pct !== undefined ? Number(p.error_pct).toFixed(2) + '%' : '-',
      ]);
      const csv = [...meta, ...ptRows].map(r => r.map(v => `"${v}"`).join(',')).join('\n');
      this.downloadCSVContent(`kalibrasi_${log.parameter_code}_log_${log.id}.csv`, csv);
    },
    printCalibrationCertificate(log) {
      let pts = [];
      try {
        pts = log.points_json ? JSON.parse(log.points_json) : [];
      } catch (e) {
        pts = [];
      }
      const rowsHtml = pts.length > 0
        ? pts.map(p => `
            <tr>
              <td style="text-align:center">${p.index}</td>
              <td style="text-align:center">${p.active ? 'Aktif' : 'Outlier'}</td>
              <td style="text-align:right">${Number(p.raw_x).toFixed(3)}</td>
              <td style="text-align:right">${Number(p.ref_y).toFixed(3)}</td>
              <td style="text-align:right">${Number(p.pred_y || 0).toFixed(3)}</td>
              <td style="text-align:right">${p.error_pct !== null && p.error_pct !== undefined ? Number(p.error_pct).toFixed(2) + '%' : Number(p.error_abs || 0).toFixed(3)}</td>
            </tr>`).join('')
        : '<tr><td colspan="6" style="text-align:center">Reset ke kurva default (y = x)</td></tr>';

      const win = window.open('', '_blank', 'width=880,height=720');
      if (!win) return;
      win.document.write(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8" />
  <title>Berita Acara Kalibrasi - ${log.parameter_name}</title>
  <style>
    body { font-family: Arial, sans-serif; color: #0f172a; padding: 28px; font-size: 12px; }
    h1 { font-size: 18px; margin: 0 0 4px 0; }
    .sub { color: #475569; margin-bottom: 18px; }
    .meta { width: 100%; border-collapse: collapse; margin-bottom: 18px; }
    .meta td { padding: 6px 8px; border: 1px solid #cbd5e1; }
    .meta td.lbl { background: #f1f5f9; font-weight: bold; width: 28%; }
    table.pts { width: 100%; border-collapse: collapse; margin-bottom: 24px; }
    table.pts th, table.pts td { border: 1px solid #cbd5e1; padding: 6px 8px; }
    table.pts th { background: #e2e8f0; }
    .sig { display: flex; justify-content: space-between; margin-top: 40px; }
    .sig-box { width: 42%; text-align: center; }
    .sig-line { margin-top: 64px; border-top: 1px solid #334155; padding-top: 4px; font-weight: bold; }
  </style>
</head>
<body>
  <h1>BERITA ACARA KALIBRASI SENSOR (${this.domainShortBadge(log.domain)})</h1>
  <div class="sub">Sistem Akuisisi Data Datalogger — Kepatuhan AQMS / SPARING / WQMS ONLIMO</div>
  <table class="meta">
    <tr><td class="lbl">Waktu Pelaksanaan</td><td>${this.formatTimestamp(log.created_at)}</td><td class="lbl">Domain Kepatuhan</td><td>${this.domainFullLabel(log.domain)}</td></tr>
    <tr><td class="lbl">Perangkat</td><td>${log.device_name} (${log.device_code})</td><td class="lbl">Parameter</td><td>${log.parameter_name} (${log.unit || '-'})</td></tr>
    <tr><td class="lbl">Metode &amp; Titik</td><td>${this.formatLogMethod(log)}</td><td class="lbl">Teknisi Pelaksana</td><td>${log.performed_by} (${log.performed_role})</td></tr>
    <tr><td class="lbl">Persamaan Baru</td><td><strong>${log.equation}</strong></td><td class="lbl">Persamaan Sebelumnya</td><td>${log.prev_equation || '-'}</td></tr>
    <tr><td class="lbl">Skor Determinasi (R²)</td><td>${log.r2 !== null && log.r2 !== undefined ? Number(log.r2).toFixed(5) + (log.passed_r2 ? ' (LULUS)' : ' (DI BAWAH BATAS)') : 'Mode 1-Titik / Default'}</td><td class="lbl">Batas Minimum R² Parameter</td><td>≥ ${Number(log.min_r2 || 0.995).toFixed(4)}</td></tr>
    <tr><td class="lbl">Catatan Standar CRM / Gas</td><td colspan="3">${log.standard_note || '-'}</td></tr>
  </table>
  <h3>Tabel Titik Pengujian Standar (1–10 Titik)</h3>
  <table class="pts">
    <thead>
      <tr><th>#</th><th>Status</th><th>Bacaan Sensor (x)</th><th>Standar Referensi (y)</th><th>Estimasi Kurva (ŷ)</th><th>Simpangan / Error</th></tr>
    </thead>
    <tbody>${rowsHtml}</tbody>
  </table>
  <div class="sig">
    <div class="sig-box">Dibuat / Dikalibrasi Oleh,<div class="sig-line">${log.performed_by} (${log.performed_role})</div></div>
    <div class="sig-box">Diperiksa / Disetujui Oleh,<div class="sig-line">Penanggung Jawab Stasiun</div></div>
  </div>
  <script>window.onload = function() { window.print(); };<\/script>
</body>
</html>`);
      win.document.close();
    },
  },
};
</script>

