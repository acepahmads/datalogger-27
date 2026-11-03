<template>
  <div
    class="bento-wqms-strip font-sans"
    :class="effectiveIsDark ? 'is-night' : 'is-day'"
  >
    <!-- ================================================================= -->
    <!-- ROW 1: COMPACT SINGLE-LINE IP & IKA STRIP                         -->
    <!-- ================================================================= -->
    <div class="wqms-strip-main">
      <!-- Left: IP (Indeks Pencemaran) Ring + Dominant Summary -->
      <div class="wqms-metric-group min-w-0">
        <svg
          width="42"
          height="42"
          viewBox="0 0 46 46"
          class="flex-shrink-0 cursor-pointer"
          @click="showWindowPopover = !showWindowPopover"
          :title="'Indeks Pencemaran (Kepmen LH 115/2003) · Klik untuk detail evaluasi'"
        >
          <circle
            cx="23"
            cy="23"
            r="19"
            fill="none"
            :stroke="effectiveIsDark ? '#1E293B' : '#E2E8F0'"
            stroke-width="4.5"
          />
          <circle
            cx="23"
            cy="23"
            r="19"
            fill="none"
            :stroke="ipGaugeColor"
            stroke-width="4.5"
            stroke-dasharray="119.4"
            :stroke-dashoffset="ipGaugeDashOffset"
            stroke-linecap="round"
            transform="rotate(-90 23 23)"
            style="transition: stroke-dashoffset 0.35s ease, stroke 0.25s ease;"
          />
          <text
            x="23"
            y="26.5"
            text-anchor="middle"
            :fill="effectiveIsDark ? '#F8FAFC' : '#0F172A'"
            :font-size="String(displayIP).length >= 5 ? '9.5' : '10.8'"
            font-weight="800"
          >{{ displayIP }}</text>
        </svg>

        <div class="min-w-0">
          <div class="flex flex-wrap items-center gap-1.5">
            <span class="wqms-kicker">IP DOMINAN</span>
            <span class="wqms-pill" :class="ipPillToneClass">
              <template v-if="hasValidIP">
                {{ dominantPrettyLabel }}: {{ dominantConcFormatted }} · IP {{ displayIP }} ({{ ipCategoryReadable }})
              </template>
              <template v-else>
                Belum Ada Data Parameter Air ({{ waterClassLabelShort }})
              </template>
            </span>
          </div>
          <div class="wqms-subline truncate">
            <template v-if="hasValidIP">
              Rasio Maks (Ci/Lij): <strong>{{ formatRatio(ipResult.max_ratio) }}</strong> ·
              {{ waterClassLabelShort }} (PP 22/2021) · Update {{ relativeUpdatedLabel }}
            </template>
            <template v-else>
              Petakan sensor air (pH, DO, COD, BOD, TSS, Nitrat, Amonia) melalui tombol Mappings
            </template>
          </div>
        </div>
      </div>

      <div class="wqms-v-divider hidden xl:block"></div>

      <!-- Middle: IKA (Indeks Kualitas Air) Ring + 30d Score Summary -->
      <div class="wqms-metric-group min-w-0">
        <svg
          width="42"
          height="42"
          viewBox="0 0 46 46"
          class="flex-shrink-0 cursor-pointer"
          @click="showWindowPopover = !showWindowPopover"
          :title="'Skor IKA (Permen LHK 27/2021) · Klik untuk detail bobot status mutu'"
        >
          <circle
            cx="23"
            cy="23"
            r="19"
            fill="none"
            :stroke="effectiveIsDark ? '#1E293B' : '#E2E8F0'"
            stroke-width="4.5"
          />
          <circle
            cx="23"
            cy="23"
            r="19"
            fill="none"
            :stroke="ikaGaugeColor"
            stroke-width="4.5"
            stroke-dasharray="119.4"
            :stroke-dashoffset="ikaGaugeDashOffset"
            stroke-linecap="round"
            transform="rotate(-90 23 23)"
            style="transition: stroke-dashoffset 0.35s ease, stroke 0.25s ease;"
          />
          <text
            x="23"
            y="26.5"
            text-anchor="middle"
            :fill="effectiveIsDark ? '#F8FAFC' : '#0F172A'"
            font-size="10.8"
            font-weight="800"
          >{{ displayIKA }}</text>
        </svg>

        <div class="min-w-0 relative">
          <div class="flex flex-wrap items-center gap-1.5">
            <span class="wqms-kicker">SKOR IKA ({{ ikaSummary.window_days || 30 }}H)</span>
            <span class="wqms-pill" :class="ikaPillToneClass">
              {{ hasValidIKA ? `${displayIKA} / 100 (${ikaSummary.category_label || 'Sedang'})` : 'Belum Tersedia' }}
            </span>
            <button
              type="button"
              class="wqms-valid-chip"
              :class="validityChipClass"
              @click.stop="showWindowPopover = !showWindowPopover"
              title="Klik untuk detail jendela evaluasi IP & IKA"
            >
              {{ activeCountLabel }} {{ calculationStatusLabel }}<span v-if="completenessPct > 0"> · {{ Math.round(completenessPct) }}%</span>
            </button>
          </div>
          <div class="wqms-subline truncate">
            <template v-if="hasValidIKA">
              {{ ikaSummary.pct_memenuhi }}% Memenuhi BM · {{ ikaSummary.pct_cemar_ringan }}% Cemar Ringan
              <template v-if="ikaSummary.pct_cemar_sedang > 0"> · {{ ikaSummary.pct_cemar_sedang }}% Sedang</template>
              <template v-if="ikaSummary.pct_cemar_berat > 0"> · {{ ikaSummary.pct_cemar_berat }}% Berat</template>
              (Permen LHK 27/2021)
            </template>
            <template v-else>
              Bobot Permen LHK 27/2021: Baik (70) · Ringan (50) · Sedang (30) · Berat (10)
            </template>
          </div>

          <!-- Detail Popover -->
          <div
            v-if="showWindowPopover"
            class="absolute left-0 top-full mt-2 z-50 w-80 max-w-[calc(100vw-32px)] rounded-xl shadow-2xl p-3.5 bg-white dark:bg-[#111A2C] border border-slate-200 dark:border-slate-700/80 text-slate-800 dark:text-slate-100 text-[11px]"
            @click.stop
          >
            <div class="flex items-center justify-between pb-2 mb-2 border-b border-slate-100 dark:border-slate-800">
              <div>
                <div class="font-bold text-xs text-slate-900 dark:text-white">Evaluasi Mutu Air (IP &amp; IKA)</div>
                <div class="text-[10px] text-slate-500 dark:text-slate-400">PP 22/2021 · Kepmen LH 115/2003 · Permen LHK 27/2021</div>
              </div>
              <button type="button" @click="showWindowPopover = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 px-1">✕</button>
            </div>
            <div class="space-y-1.5 text-[10.5px]">
              <div class="flex justify-between">
                <span class="text-slate-500">Kelas Air Aktif:</span>
                <span class="font-bold">{{ waterClassLabelFull }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">Rumus IPj:</span>
                <span class="font-mono font-semibold">√((M² + R²) / 2)</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">Rasio Maks (M) / Rata (R):</span>
                <span class="font-bold">{{ formatRatio(ipResult && ipResult.max_ratio) }} / {{ formatRatio(ipResult && ipResult.avg_ratio) }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-500">Periode Sampel IKA:</span>
                <span class="font-bold">{{ ikaSummary.total_periods || 0 }} periode ({{ ikaSummary.window_days || 30 }} hari)</span>
              </div>
            </div>
            <div v-if="canRecalculate" class="pt-2.5 mt-2 border-t border-slate-100 dark:border-slate-800 flex justify-end">
              <button
                type="button"
                @click="triggerRecalculate"
                :disabled="recalculating"
                class="px-2.5 py-1 rounded-md text-[10.5px] font-semibold bg-sky-600 hover:bg-sky-500 text-white disabled:opacity-50 transition"
              >
                {{ recalculating ? 'Menghitung...' : '↻ Hitung Ulang IP & IKA' }}
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Right: Action Buttons -->
      <div class="wqms-actions">
        <button
          type="button"
          class="wqms-btn wqms-btn-expand"
          @click="toggleParamsDrawer"
        >
          <span>{{ showParamsDrawer ? '▴ Sembunyikan' : '▾ Lihat' }} {{ paramCards.length || '' }} Parameter</span>
        </button>

        <select
          v-model="selectedWaterClass"
          @change="onWaterClassChange"
          :disabled="!canRecalculate || savingClass"
          class="wqms-select"
          title="Pilih Baku Mutu Kelas Air (Lampiran VI PP No. 22 Tahun 2021)"
        >
          <option value="CLASS_1">Kelas I</option>
          <option value="CLASS_2">Kelas II</option>
          <option value="CLASS_3">Kelas III</option>
          <option value="CLASS_4">Kelas IV</option>
        </select>

        <button
          v-if="canManageMapping"
          type="button"
          class="wqms-btn"
          @click="openMappingsModal"
        >
          Mappings
        </button>

        <button
          type="button"
          class="wqms-btn"
          @click="openTrendModal"
        >
          Trend
        </button>

        <button
          type="button"
          class="wqms-btn wqms-btn-primary"
          @click="$router.push('/monitoring/wqms')"
        >
          <span>Full IP &amp; IKA</span>
          <span>›</span>
        </button>
      </div>
    </div>

    <!-- ================================================================= -->
    <!-- ROW 2: EXPANDABLE PARAMETER RATIO SUB-CARDS                       -->
    <!-- ================================================================= -->
    <div v-if="showParamsDrawer" class="wqms-subcards-grid">
      <template v-if="paramCards.length > 0">
        <div
          v-for="item in paramCards"
          :key="item.param_code"
          class="wqms-subcard"
          :class="{
            'is-dominant': item.is_dominant,
            'is-exceeded': item.exceeds_standard
          }"
        >
          <div class="flex items-center justify-between gap-1">
            <span class="text-[11px] font-bold tracking-tight truncate">
              {{ formatParamPretty(item.param_code) }}<span v-if="item.is_dominant" class="text-amber-500 ml-0.5">★</span>
              <span v-if="item.source_type === 'LAB'" class="text-[9px] font-semibold px-1 py-0.2 rounded bg-purple-500/15 text-purple-600 dark:text-purple-300 ml-1">LAB</span>
            </span>
            <strong
              class="text-xs font-extrabold tabular-nums"
              :style="{ color: getRatioColor(item.adjusted_ratio) }"
            >
              {{ formatRatio(item.adjusted_ratio) }}
            </strong>
          </div>
          <div class="wqms-ratio-track">
            <div
              class="wqms-ratio-fill"
              :style="{
                width: getRatioBarWidthPct(item.adjusted_ratio) + '%',
                background: getRatioColor(item.adjusted_ratio)
              }"
            ></div>
          </div>
          <div class="text-[10px] text-slate-500 dark:text-slate-400 truncate tabular-nums">
            {{ formatValue(item.normalized_measured) }} {{ item.normalized_unit }} (BM: {{ item.standard_label }})
          </div>
        </div>
      </template>
      <div v-else class="col-span-full text-center py-2 text-xs text-slate-500 dark:text-slate-400">
        Belum ada parameter kualitas air yang aktif. Klik <strong>Mappings</strong> untuk memetakan sensor atau mengisi nilai uji lab.
      </div>
    </div>

    <!-- ================================================================= -->
    <!-- QUICK TREND MODAL                                                 -->
    <!-- ================================================================= -->
    <div
      v-if="showTrendModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-xs p-4"
      @click.self="showTrendModal = false"
    >
      <div class="w-full max-w-3xl rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 shadow-2xl p-4 space-y-3 text-slate-800 dark:text-slate-100">
        <div class="flex items-center justify-between pb-2 border-b border-slate-200 dark:border-slate-800">
          <div>
            <h3 class="text-sm font-bold">Tren Indeks Pencemaran (IP) — {{ waterClassLabelFull }}</h3>
            <p class="text-[11px] text-slate-500 dark:text-slate-400">Batas Memenuhi Baku Mutu (IP ≤ 1.0) &amp; Cemar Ringan (IP ≤ 5.0)</p>
          </div>
          <div class="flex items-center gap-1.5">
            <button
              v-for="r in ['24h', '7d', '30d']"
              :key="r"
              type="button"
              @click="changeTrendRange(r)"
              class="px-2 py-1 rounded text-xs font-semibold border transition"
              :class="trendRange === r
                ? 'bg-sky-600 text-white border-sky-500'
                : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700'"
            >
              {{ r.toUpperCase() }}
            </button>
            <button type="button" @click="showTrendModal = false" class="ml-2 px-2 py-1 text-slate-400 hover:text-slate-700 dark:hover:text-white">✕</button>
          </div>
        </div>

        <div v-if="trendPoints.length > 0" class="w-full overflow-hidden">
          <svg viewBox="0 0 720 210" class="w-full h-52">
            <!-- Threshold reference lines -->
            <line x1="42" :y1="trendY(1.0)" x2="705" :y2="trendY(1.0)" stroke="#10B981" stroke-dasharray="4 4" stroke-width="1.2" />
            <text x="46" :y="Math.max(12, trendY(1.0) - 4)" fill="#10B981" font-size="9.5" font-weight="700">BM = 1.0 (Memenuhi)</text>
            <line x1="42" :y1="trendY(5.0)" x2="705" :y2="trendY(5.0)" stroke="#F59E0B" stroke-dasharray="4 4" stroke-width="1.2" />
            <text x="46" :y="Math.max(12, trendY(5.0) - 4)" fill="#F59E0B" font-size="9.5" font-weight="700">IP = 5.0 (Batas Cemar Ringan)</text>

            <!-- Polyline -->
            <polyline
              fill="none"
              stroke="#0284C7"
              stroke-width="2.4"
              stroke-linecap="round"
              stroke-linejoin="round"
              :points="trendPolylinePoints"
            />
            <circle
              v-for="(pt, idx) in trendSampledDots"
              :key="idx"
              :cx="pt.x"
              :cy="pt.y"
              r="3"
              :fill="getIPColorByValue(pt.val)"
            />
          </svg>
        </div>
        <div v-else class="py-10 text-center text-xs text-slate-500">
          Belum cukup riwayat titik perhitungan IP pada rentang waktu ini.
        </div>
      </div>
    </div>

    <!-- ================================================================= -->
    <!-- MAPPINGS & LAB INPUT MODAL                                        -->
    <!-- ================================================================= -->
    <div
      v-if="showMappingsModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-slate-900/60 backdrop-blur-xs p-4"
      @click.self="showMappingsModal = false"
    >
      <div class="w-full max-w-3xl rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 shadow-2xl p-4 space-y-3 text-slate-800 dark:text-slate-100 max-h-[88vh] overflow-y-auto">
        <div class="flex items-center justify-between pb-2 border-b border-slate-200 dark:border-slate-800">
          <div>
            <h3 class="text-sm font-bold">Pemetaan Parameter WQMS (Sensor Online &amp; Input Uji Lab)</h3>
            <p class="text-[11px] text-slate-500 dark:text-slate-400">Hubungkan kanal sensor stasiun atau masukkan hasil uji laboratorium berkala untuk kalkulasi IP &amp; IKA.</p>
          </div>
          <button type="button" @click="showMappingsModal = false" class="px-2 py-1 text-slate-400 hover:text-slate-700 dark:hover:text-white">✕</button>
        </div>

        <!-- Add / Update Form -->
        <div class="grid grid-cols-1 sm:grid-cols-12 gap-2 p-3 rounded-xl bg-slate-50 dark:bg-slate-900/70 border border-slate-200 dark:border-slate-800 text-xs">
          <div class="sm:col-span-3">
            <label class="block text-[10px] font-bold uppercase text-slate-500 mb-1">Parameter Resmi</label>
            <select v-model="mappingForm.param_code" class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800">
              <option v-for="opt in officialParamOptions" :key="opt.code" :value="opt.code">{{ opt.label }}</option>
            </select>
          </div>
          <div class="sm:col-span-3">
            <label class="block text-[10px] font-bold uppercase text-slate-500 mb-1">Sumber Data</label>
            <select v-model="mappingForm.source_type" class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800">
              <option value="SENSOR">Sensor Online</option>
              <option value="LAB">Input Lab Berkala</option>
            </select>
          </div>
          <div v-if="mappingForm.source_type === 'SENSOR'" class="sm:col-span-4">
            <label class="block text-[10px] font-bold uppercase text-slate-500 mb-1">Kanal Sensor</label>
            <select v-model.number="mappingForm.parameter_id" class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800">
              <option :value="0">-- Pilih Parameter Sensor --</option>
              <option v-for="sp in allAvailableSensorParams" :key="sp.id" :value="sp.id">
                {{ sp.displayLabel }}
              </option>
            </select>
          </div>
          <div v-else class="sm:col-span-4">
            <label class="block text-[10px] font-bold uppercase text-slate-500 mb-1">Nilai Uji Lab</label>
            <input
              v-model.number="mappingForm.lab_value"
              type="number"
              step="any"
              placeholder="Contoh: 0.16"
              class="w-full px-2 py-1.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-800"
            />
          </div>
          <div class="sm:col-span-2 flex items-end">
            <button
              type="button"
              @click="saveMappingItem"
              :disabled="savingMapping"
              class="w-full py-1.5 px-3 rounded-lg bg-sky-600 hover:bg-sky-500 text-white font-semibold disabled:opacity-50 transition"
            >
              {{ savingMapping ? '...' : '+ Simpan' }}
            </button>
          </div>
        </div>

        <!-- Existing Mappings Table -->
        <div class="overflow-x-auto">
          <table class="w-full text-xs border-collapse">
            <thead>
              <tr class="text-left text-[10px] uppercase text-slate-500 border-b border-slate-200 dark:border-slate-800">
                <th class="py-2 px-2">Parameter</th>
                <th class="py-2 px-2">Sumber</th>
                <th class="py-2 px-2">Sensor / Nilai Lab</th>
                <th class="py-2 px-2 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="m in mappingsList"
                :key="m.id"
                class="border-b border-slate-100 dark:border-slate-800/60"
              >
                <td class="py-2 px-2 font-bold">{{ formatParamPretty(m.param_code) }}</td>
                <td class="py-2 px-2">
                  <span
                    class="px-1.5 py-0.5 rounded text-[10px] font-semibold"
                    :class="m.source_type === 'LAB' ? 'bg-purple-500/15 text-purple-600 dark:text-purple-300' : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-300'"
                  >
                    {{ m.source_type === 'LAB' ? 'Input Lab' : 'Sensor Online' }}
                  </span>
                </td>
                <td class="py-2 px-2">
                  <template v-if="m.source_type === 'LAB'">
                    <strong>{{ m.lab_value }}</strong> {{ m.normalized_unit }}
                  </template>
                  <template v-else>
                    {{ getMappingSensorLabel(m) }}
                    <span
                      v-if="!isOperator && m.parameter && m.parameter.parameter_code && m.parameter.parameter_code !== m.parameter.parameter_name"
                      class="text-slate-400"
                    >
                      ({{ m.parameter.parameter_code }})
                    </span>
                  </template>
                </td>
                <td class="py-2 px-2 text-right">
                  <button
                    type="button"
                    @click="deleteMappingItem(m.id)"
                    class="text-rose-500 hover:text-rose-600 font-semibold text-[11px]"
                  >
                    Hapus
                  </button>
                </td>
              </tr>
              <tr v-if="mappingsList.length === 0">
                <td colspan="4" class="py-4 text-center text-slate-400">Belum ada pemetaan parameter.</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'BentoWQMSIndexBar',
  props: {
    isDark: {
      type: Boolean,
      default: false,
    },
    devicesList: {
      type: Array,
      default: () => [],
    },
  },
  data() {
    let savedExpanded = false;
    try {
      savedExpanded = localStorage.getItem('datalogger_wqms_show_params') === 'true';
    } catch (e) {}
    return {
      bundle: null,
      selectedWaterClass: 'CLASS_2',
      showParamsDrawer: savedExpanded,
      showWindowPopover: false,
      showTrendModal: false,
      showMappingsModal: false,
      recalculating: false,
      savingClass: false,
      savingMapping: false,
      trendRange: '30d',
      trendPoints: [],
      mappingsList: [],
      mappingForm: {
        param_code: 'COD',
        source_type: 'SENSOR',
        parameter_id: 0,
        lab_value: null,
      },
      pollTimer: null,
      wsUnsub: null,
    };
  },
  computed: {
    effectiveIsDark() {
      return Boolean(this.isDark);
    },
    userRole() {
      const u = this.$store && this.$store.state && this.$store.state.user;
      return String((u && (u.role_name || u.role)) || '').toUpperCase();
    },
    isOperator() {
      const path = (this.$route && this.$route.path) || '';
      if (path.includes('/operator') || path.includes('dashboard-operator')) {
        return true;
      }
      if (this.userRole === 'OPERATOR') {
        return true;
      }
      return Boolean(this.$store && this.$store.getters && this.$store.getters.isOperator);
    },
    canManageMapping() {
      return true;
    },
    canRecalculate() {
      return ['SUPERADMIN', 'ADMIN', 'ADMINISTRATOR', 'ENGINEER', 'OPERATOR'].includes(this.userRole);
    },
    ipResult() {
      if (!this.bundle) return null;
      return this.bundle.operational_ip_result || this.bundle.ip_result || null;
    },
    ikaSummary() {
      return (this.bundle && this.bundle.ika_summary) || {
        score: 0,
        category: 'TIDAK_TERSEDIA',
        category_label: 'Tidak Tersedia',
        window_days: 30,
        total_periods: 0,
        pct_memenuhi: 0,
        pct_cemar_ringan: 0,
        pct_cemar_sedang: 0,
        pct_cemar_berat: 0,
      };
    },
    hasValidIP() {
      return Boolean(
        this.ipResult &&
        this.ipResult.active_parameters_count > 0 &&
        this.ipResult.overall_category &&
        this.ipResult.overall_category !== 'TIDAK_TERSEDIA'
      );
    },
    hasValidIKA() {
      return Boolean(this.ikaSummary && this.ikaSummary.total_periods > 0 && this.ikaSummary.category !== 'TIDAK_TERSEDIA');
    },
    displayIP() {
      if (!this.hasValidIP) return '—';
      return Number(this.ipResult.overall_ip || 0).toFixed(2);
    },
    displayIKA() {
      if (!this.hasValidIKA) return '—';
      return Number(this.ikaSummary.score || 0).toFixed(1);
    },
    ipGaugeColor() {
      if (!this.hasValidIP) return '#94A3B8';
      return this.getIPColorByCategory(this.ipResult.overall_category);
    },
    ipGaugeDashOffset() {
      const circ = 119.4;
      if (!this.hasValidIP) return circ;
      const ip = Math.max(0, Math.min(12, Number(this.ipResult.overall_ip) || 0));
      const pct = Math.min(1, ip / 10.0);
      return circ * (1 - Math.max(0.08, pct));
    },
    ikaGaugeColor() {
      if (!this.hasValidIKA) return '#94A3B8';
      const s = Number(this.ikaSummary.score) || 0;
      if (s >= 90) return '#06B6D4';
      if (s >= 70) return '#10B981';
      if (s >= 50) return '#0D9488';
      if (s >= 25) return '#F59E0B';
      return '#EF4444';
    },
    ikaGaugeDashOffset() {
      const circ = 119.4;
      if (!this.hasValidIKA) return circ;
      const pct = Math.max(0, Math.min(100, Number(this.ikaSummary.score) || 0)) / 100.0;
      return circ * (1 - pct);
    },
    ipCategoryReadable() {
      if (!this.ipResult) return 'Belum Tersedia';
      switch (this.ipResult.overall_category) {
        case 'MEMENUHI_BAKU_MUTU': return 'Memenuhi BM';
        case 'CEMAR_RINGAN': return 'Cemar Ringan';
        case 'CEMAR_SEDANG': return 'Cemar Sedang';
        case 'CEMAR_BERAT': return 'Cemar Berat';
        default: return 'Belum Tersedia';
      }
    },
    ipPillToneClass() {
      if (!this.hasValidIP) return 'tone-muted';
      switch (this.ipResult.overall_category) {
        case 'MEMENUHI_BAKU_MUTU': return 'tone-emerald';
        case 'CEMAR_RINGAN': return 'tone-amber';
        case 'CEMAR_SEDANG': return 'tone-orange';
        case 'CEMAR_BERAT': return 'tone-rose';
        default: return 'tone-muted';
      }
    },
    ikaPillToneClass() {
      if (!this.hasValidIKA) return 'tone-muted';
      const s = Number(this.ikaSummary.score) || 0;
      if (s >= 70) return 'tone-emerald';
      if (s >= 50) return 'tone-teal';
      if (s >= 25) return 'tone-amber';
      return 'tone-rose';
    },
    dominantPrettyLabel() {
      return this.formatParamPretty(this.ipResult && this.ipResult.dominant_parameter);
    },
    dominantConcFormatted() {
      if (!this.ipResult) return '-';
      const v = Number(this.ipResult.dominant_concentration || 0);
      const u = this.ipResult.dominant_unit || '';
      return `${v.toFixed(2)} ${u}`.trim();
    },
    waterClassLabelShort() {
      const map = { CLASS_1: 'Kelas I', CLASS_2: 'Kelas II', CLASS_3: 'Kelas III', CLASS_4: 'Kelas IV' };
      return map[this.selectedWaterClass] || 'Kelas II';
    },
    waterClassLabelFull() {
      const map = {
        CLASS_1: 'Kelas I (Air Baku Air Minum)',
        CLASS_2: 'Kelas II (Sungai / Rekreasi Air)',
        CLASS_3: 'Kelas III (Perikanan & Peternakan)',
        CLASS_4: 'Kelas IV (Pengairan Pertanaman)',
      };
      return map[this.selectedWaterClass] || 'Kelas II';
    },
    activeCountLabel() {
      if (!this.ipResult) return '0/0';
      const active = Number(this.ipResult.active_parameters_count || 0);
      const total = active + Number(this.ipResult.missing_parameters_count || 0);
      return `${active}/${total || active}`;
    },
    calculationStatusLabel() {
      return (this.ipResult && this.ipResult.calculation_status) || 'IDLE';
    },
    completenessPct() {
      return Number((this.ipResult && this.ipResult.completeness_percent) || 0);
    },
    validityChipClass() {
      const st = this.calculationStatusLabel;
      if (st === 'VALID') return 'chip-valid';
      if (st === 'PROVISIONAL' || st === 'INCOMPLETE') return 'chip-warn';
      return 'chip-muted';
    },
    relativeUpdatedLabel() {
      const ts = this.ipResult && (this.ipResult.calculated_at || this.ipResult.period_end);
      if (!ts) return '-';
      const diffSec = Math.max(0, Math.floor((Date.now() - new Date(ts).getTime()) / 1000));
      if (diffSec < 60) return `${diffSec}s ago`;
      if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
      return `${Math.floor(diffSec / 3600)}h ago`;
    },
    paramCards() {
      const list = (this.ipResult && this.ipResult.parameter_results) || [];
      return [...list].sort((a, b) => {
        if (a.is_dominant && !b.is_dominant) return -1;
        if (!a.is_dominant && b.is_dominant) return 1;
        return (Number(b.adjusted_ratio) || 0) - (Number(a.adjusted_ratio) || 0);
      });
    },
    officialParamOptions() {
      return [
        { code: 'PH', label: 'pH (Derajat Keasaman)' },
        { code: 'DO', label: 'DO (Oksigen Terlarut)' },
        { code: 'COD', label: 'COD' },
        { code: 'BOD', label: 'BOD5' },
        { code: 'TSS', label: 'TSS (Padatan Tersuspensi)' },
        { code: 'NO3', label: 'Nitrat (NO3-N)' },
        { code: 'NH3', label: 'Amonia (NH3-N)' },
        { code: 'TDS', label: 'TDS (Padatan Terlarut)' },
        { code: 'PO4', label: 'Total Fosfat (PO4-P)' },
        { code: 'TURBIDITY', label: 'Kekeruhan (Turbidity)' },
        { code: 'FECAL_COLI', label: 'Fecal Coliform' },
      ];
    },
    allAvailableSensorParams() {
      const out = [];
      for (const dev of (this.devicesList || [])) {
        if (!dev || dev.deleted_at || dev.enabled === false || dev.enabled === 0) continue;
        const devStatus = String(dev.status || dev.administrative_status || '').toUpperCase();
        if (devStatus === 'DISABLED' || devStatus === 'INACTIVE') continue;
        if (!Array.isArray(dev.parameters)) continue;
        const devCode = dev.device_code || dev.code || dev.name || '';
        for (const p of dev.parameters) {
          if (!p || p.deleted_at || p.enabled === false || p.enabled === 0) continue;
          const paramName = p.parameter_name || p.param_name || p.name || p.parameter_code || '';
          const unit = (p.unit || '').trim();
          const unitSuffix = (unit && unit !== '-') ? ` (${unit})` : '';
          const baseLabel = `${paramName}${unitSuffix}`;
          out.push({
            id: p.id || p.param_id,
            device_code: devCode,
            parameter_name: paramName,
            unit,
            displayLabel: (this.isOperator || !devCode) ? baseLabel : `[${devCode}] ${baseLabel}`,
          });
        }
      }
      return out;
    },
    trendMaxVal() {
      let m = 5.5;
      for (const p of this.trendPoints) {
        const v = Number(p.overall_ip) || 0;
        if (v + 0.5 > m) m = v + 0.5;
      }
      return m;
    },
    trendSampledDots() {
      const n = this.trendPoints.length;
      if (!n) return [];
      return this.trendPoints.map((r, i) => {
        const x = n === 1 ? 370 : 48 + (i / (n - 1)) * 650;
        const val = Number(r.overall_ip) || 0;
        return { x: Number(x.toFixed(1)), y: this.trendY(val), val };
      });
    },
    trendPolylinePoints() {
      return this.trendSampledDots.map(d => `${d.x},${d.y}`).join(' ');
    },
  },
  mounted() {
    this.fetchLatest();
    this.pollTimer = setInterval(() => {
      this.fetchLatest();
    }, 30000);
    window.addEventListener('wqms-calculated-event', this.onWqmsCalculatedEvent);
    document.addEventListener('click', this.handleOutsideClick);
  },
  beforeDestroy() {
    if (this.pollTimer) clearInterval(this.pollTimer);
    window.removeEventListener('wqms-calculated-event', this.onWqmsCalculatedEvent);
    document.removeEventListener('click', this.handleOutsideClick);
  },
  methods: {
    onWqmsCalculatedEvent() {
      this.fetchLatest();
    },
    getAuthHeaders() {
      const token = localStorage.getItem('datalogger_token') || localStorage.getItem('token') || (this.$store && this.$store.state && this.$store.state.token) || '';
      return token ? { Authorization: `Bearer ${token}` } : {};
    },
    handleOutsideClick() {
      if (this.showWindowPopover) this.showWindowPopover = false;
    },
    toggleParamsDrawer() {
      this.showParamsDrawer = !this.showParamsDrawer;
      try {
        localStorage.setItem('datalogger_wqms_show_params', this.showParamsDrawer ? 'true' : 'false');
      } catch (e) {}
    },
    async fetchLatest() {
      try {
        const res = await axios.get('/api/wqms/latest?mode=OPERATIONAL', {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.data) {
          this.bundle = res.data.data;
          if (this.bundle.config && this.bundle.config.water_class) {
            this.selectedWaterClass = this.bundle.config.water_class;
          }
        }
      } catch (e) {}
    },
    async onWaterClassChange() {
      this.savingClass = true;
      try {
        const res = await axios.put(
          '/api/wqms/config',
          { water_class: this.selectedWaterClass },
          { headers: this.getAuthHeaders() }
        );
        if (res.data && res.data.data) {
          this.bundle = res.data.data;
        }
      } catch (e) {
      } finally {
        this.savingClass = false;
      }
    },
    async triggerRecalculate() {
      this.recalculating = true;
      try {
        const res = await axios.post(
          '/api/wqms/recalculate',
          {},
          { headers: this.getAuthHeaders() }
        );
        if (res.data && res.data.data) {
          this.bundle = res.data.data;
        }
      } catch (e) {
      } finally {
        this.recalculating = false;
      }
    },
    async openTrendModal() {
      this.showTrendModal = true;
      await this.fetchTrend();
    },
    async changeTrendRange(r) {
      this.trendRange = r;
      await this.fetchTrend();
    },
    async fetchTrend() {
      try {
        const res = await axios.get(`/api/wqms/trend?range=${this.trendRange}&mode=REGULATORY`, {
          headers: this.getAuthHeaders(),
        });
        this.trendPoints = (res.data && res.data.data) || [];
      } catch (e) {
        this.trendPoints = [];
      }
    },
    trendY(val) {
      const top = 16;
      const bottom = 192;
      const ratio = Math.max(0, Math.min(1, val / this.trendMaxVal));
      return Number((bottom - ratio * (bottom - top)).toFixed(1));
    },
    async openMappingsModal() {
      this.showMappingsModal = true;
      await this.fetchMappings();
    },
    async fetchMappings() {
      try {
        const res = await axios.get('/api/wqms/mappings', { headers: this.getAuthHeaders() });
        this.mappingsList = (res.data && res.data.data) || [];
      } catch (e) {
        this.mappingsList = [];
      }
    },
    async saveMappingItem() {
      if (this.mappingForm.source_type === 'SENSOR' && !this.mappingForm.parameter_id) return;
      if (this.mappingForm.source_type === 'LAB' && (this.mappingForm.lab_value === null || this.mappingForm.lab_value === '')) return;
      this.savingMapping = true;
      try {
        const payload = {
          param_code: this.mappingForm.param_code,
          source_type: this.mappingForm.source_type,
          parameter_id: this.mappingForm.source_type === 'SENSOR' ? this.mappingForm.parameter_id : null,
          lab_value: this.mappingForm.source_type === 'LAB' ? Number(this.mappingForm.lab_value) : null,
          enabled: true,
        };
        await axios.post('/api/wqms/mappings', payload, { headers: this.getAuthHeaders() });
        await this.fetchMappings();
        await this.fetchLatest();
      } catch (e) {
      } finally {
        this.savingMapping = false;
      }
    },
    async deleteMappingItem(id) {
      try {
        await axios.delete(`/api/wqms/mappings/${id}`, { headers: this.getAuthHeaders() });
        await this.fetchMappings();
        await this.fetchLatest();
      } catch (e) {}
    },
    formatParamPretty(code) {
      const map = {
        PH: 'pH',
        DO: 'DO',
        COD: 'COD',
        BOD: 'BOD',
        TSS: 'TSS',
        NO3: 'NO3-N',
        NH3: 'NH3-N',
        TDS: 'TDS',
        PO4: 'PO4-P',
        TURBIDITY: 'Turbidity',
        FECAL_COLI: 'F.Coliform',
      };
      return map[code] || code || '-';
    },
    getMappingSensorLabel(m) {
      if (m && m.parameter && (m.parameter.parameter_name || m.parameter.name)) {
        return m.parameter.parameter_name || m.parameter.name;
      }
      if (m && m.parameter_id) {
        const found = (this.allAvailableSensorParams || []).find(sp => Number(sp.id) === Number(m.parameter_id));
        if (found && found.parameter_name) return found.parameter_name;
      }
      if (this.isOperator) {
        return this.formatParamPretty(m && m.param_code);
      }
      return m && m.parameter_id ? `Param #${m.parameter_id}` : '-';
    },
    formatRatio(v) {
      if (v === null || v === undefined || isNaN(Number(v))) return '0.00';
      return Number(v).toFixed(2);
    },
    formatValue(v) {
      if (v === null || v === undefined || isNaN(Number(v))) return '0.00';
      return Number(v).toFixed(2);
    },
    getIPColorByCategory(cat) {
      switch (cat) {
        case 'MEMENUHI_BAKU_MUTU': return '#10B981';
        case 'CEMAR_RINGAN': return '#F59E0B';
        case 'CEMAR_SEDANG': return '#F97316';
        case 'CEMAR_BERAT': return '#EF4444';
        default: return '#94A3B8';
      }
    },
    getIPColorByValue(v) {
      if (v <= 1.0) return '#10B981';
      if (v <= 5.0) return '#F59E0B';
      if (v <= 10.0) return '#F97316';
      return '#EF4444';
    },
    getRatioColor(r) {
      const v = Number(r) || 0;
      if (v <= 1.0) return '#10B981';
      if (v <= 3.0) return '#F59E0B';
      return '#EF4444';
    },
    getRatioBarWidthPct(r) {
      const v = Math.max(0, Number(r) || 0);
      return Math.min(100, Math.max(6, Math.round((v / 2.6) * 100)));
    },
  },
};
</script>

<style scoped>
.bento-wqms-strip {
  border-radius: 12px;
  padding: 8px 12px;
  border: 1px solid #E2E8F0;
  background: #FFFFFF;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.04);
}
.bento-wqms-strip.is-night {
  background: #0D1626;
  border-color: #1E3048;
  color: #F1F5F9;
}

.wqms-strip-main {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px 14px;
}

.wqms-metric-group {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1 1 290px;
}

.wqms-v-divider {
  width: 1px;
  height: 32px;
  background: #E2E8F0;
}
.is-night .wqms-v-divider {
  background: #1E3048;
}

.wqms-kicker {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.04em;
  color: #475569;
  text-transform: uppercase;
}
.is-night .wqms-kicker {
  color: #94A3B8;
}

.wqms-subline {
  font-size: 10.5px;
  color: #64748B;
  margin-top: 2px;
}
.is-night .wqms-subline {
  color: #94A3B8;
}

.wqms-pill {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 10.5px;
  font-weight: 700;
  line-height: 1.3;
}
.tone-emerald {
  background: rgba(16, 185, 129, 0.14);
  color: #047857;
  border: 1px solid rgba(16, 185, 129, 0.35);
}
.is-night .tone-emerald {
  color: #34D399;
}
.tone-teal {
  background: rgba(20, 184, 166, 0.14);
  color: #0F766E;
  border: 1px solid rgba(20, 184, 166, 0.35);
}
.is-night .tone-teal {
  color: #2DD4BF;
}
.tone-amber {
  background: rgba(245, 158, 11, 0.16);
  color: #B45309;
  border: 1px solid rgba(245, 158, 11, 0.4);
}
.is-night .tone-amber {
  color: #FBBF24;
}
.tone-orange {
  background: rgba(249, 115, 22, 0.16);
  color: #C2410C;
  border: 1px solid rgba(249, 115, 22, 0.4);
}
.is-night .tone-orange {
  color: #FB923C;
}
.tone-rose {
  background: rgba(239, 68, 68, 0.16);
  color: #B91C1C;
  border: 1px solid rgba(239, 68, 68, 0.4);
}
.is-night .tone-rose {
  color: #F87171;
}
.tone-muted {
  background: rgba(148, 163, 184, 0.14);
  color: #64748B;
  border: 1px solid rgba(148, 163, 184, 0.3);
}

.wqms-valid-chip {
  padding: 2px 7px;
  border-radius: 999px;
  font-size: 9.5px;
  font-weight: 700;
  cursor: pointer;
}
.chip-valid {
  background: rgba(16, 185, 129, 0.14);
  color: #059669;
  border: 1px solid rgba(16, 185, 129, 0.3);
}
.is-night .chip-valid {
  color: #34D399;
}
.chip-warn {
  background: rgba(245, 158, 11, 0.14);
  color: #D97706;
  border: 1px solid rgba(245, 158, 11, 0.3);
}
.chip-muted {
  background: rgba(148, 163, 184, 0.14);
  color: #64748B;
  border: 1px solid rgba(148, 163, 184, 0.3);
}

.wqms-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
}

.wqms-btn,
.wqms-select {
  height: 28px;
  padding: 0 10px;
  border-radius: 7px;
  font-size: 11px;
  font-weight: 600;
  border: 1px solid #CBD5E1;
  background: #F8FAFC;
  color: #334155;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.wqms-btn:hover,
.wqms-select:hover {
  border-color: #0284C7;
  color: #0284C7;
}
.is-night .wqms-btn,
.is-night .wqms-select {
  background: #132036;
  border-color: #243956;
  color: #E2E8F0;
}
.wqms-btn-expand {
  background: rgba(14, 165, 233, 0.1);
  border-color: rgba(14, 165, 233, 0.35);
  color: #0369A1;
}
.is-night .wqms-btn-expand {
  color: #38BDF8;
}
.wqms-btn-primary {
  background: #0284C7;
  border-color: #0284C7;
  color: #FFFFFF !important;
}
.wqms-btn-primary:hover {
  background: #0369A1;
}

.wqms-subcards-grid {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #F1F5F9;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(145px, 1fr));
  gap: 7px;
}
.is-night .wqms-subcards-grid {
  border-top-color: #1B2B44;
}

.wqms-subcard {
  padding: 7px 9px;
  border-radius: 9px;
  border: 1px solid #E2E8F0;
  background: #F8FAFC;
}
.is-night .wqms-subcard {
  background: #111D31;
  border-color: #1F324D;
}
.wqms-subcard.is-dominant {
  border-color: #F59E0B;
  background: rgba(245, 158, 11, 0.06);
}

.wqms-ratio-track {
  height: 4px;
  border-radius: 999px;
  background: #E2E8F0;
  overflow: hidden;
  margin: 5px 0;
}
.is-night .wqms-ratio-track {
  background: #1E293B;
}
.wqms-ratio-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s ease;
}
</style>
