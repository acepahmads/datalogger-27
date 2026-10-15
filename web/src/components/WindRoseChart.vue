<template>
  <div class="scada-panel p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border border-[#334155] dark:border-[#20354B] shadow-xs dark:shadow-md flex flex-col h-full select-none">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-1.5 pb-2 border-b border-[#334155] dark:border-[#20354B]">
      <div>
        <div class="flex items-center space-x-2">
          <span class="w-2 h-2 rounded-full bg-cyan-500"></span>
          <h3 class="text-xs font-bold tracking-wider uppercase text-slate-900 dark:text-white font-mono">
            {{ $t('operator.windRoseTitle') || 'WIND ROSE' }}
          </h3>
        </div>
        <p class="text-4xs text-slate-500 dark:text-slate-400 font-mono mt-0.5">
          {{ $t('operator.windRoseSubtitle') || 'Wind direction and speed distribution' }}
        </p>
      </div>

      <!-- Controls: Station & Timeframe -->
      <div class="flex items-center space-x-1.5 text-3xs font-mono">
        <!-- Station Dropdown -->
        <select
          v-model="selectedDeviceId"
          @change="onDeviceChange"
          class="px-2 py-0.5 rounded bg-slate-100 dark:bg-[#101B2B] border border-[#334155] dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-3xs focus:outline-none focus:border-cyan-500"
          :title="$t('operator.selectStation') || 'Select Station'"
        >
          <option v-for="d in availableDevices" :key="d.id" :value="d.id">
            {{ d.device_code || d.code }}
          </option>
        </select>

        <!-- Timeframe Preset Buttons -->
        <div class="flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded border border-[#334155] dark:border-[#20354B]">
          <button
            v-for="tf in ['1h', '6h', '24h', '7d']"
            :key="tf"
            @click="setTimeframe(tf)"
            class="px-1.5 py-0.5 rounded transition font-semibold"
            :class="timeframe === tf
              ? 'bg-cyan-500 text-slate-950 font-black shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
          >
            {{ tf }}
          </button>
        </div>

        <!-- Sync Button -->
        <button
          @click="fetchWindData"
          :disabled="isLoading"
          class="p-1 rounded bg-slate-100 dark:bg-[#101B2B] hover:bg-slate-200 dark:hover:bg-[#18283E] text-slate-600 dark:text-slate-300 transition"
          :title="$t('common.refresh') || 'Refresh'"
        >
          <svg class="w-3 h-3" :class="{ 'animate-spin': isLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col justify-between pt-1 relative min-h-[220px]">
      <!-- Loading State -->
      <div v-if="isLoading" class="absolute inset-0 bg-white/70 dark:bg-[#0B1422]/80 z-20 flex items-center justify-center text-cyan-600 dark:text-cyan-400 text-3xs font-mono">
        <span class="animate-spin w-4 h-4 border-2 border-cyan-500 border-t-transparent rounded-full mr-2"></span>
        Loading Wind Observations...
      </div>

      <!-- State A: No Wind Data / No Wind Sensors Configured -->
      <div
        v-if="!hasWindSensors"
        class="flex-1 flex flex-col items-center justify-center p-3 text-center font-mono space-y-1.5"
      >
        <span class="text-2xl">💨</span>
        <div class="text-xs font-bold text-slate-700 dark:text-slate-300">
          {{ $t('operator.windUnavailable') || 'Wind data unavailable' }}
        </div>
        <p class="text-4xs text-slate-500 dark:text-slate-400 max-w-[220px]">
          No paired wind direction (WD) and speed (WS) sensors detected on {{ currentDeviceCode }}. Select another station with meteorological sensors.
        </p>
      </div>

      <!-- State B: No Observations in Timeframe -->
      <div
        v-else-if="totalValidSamples === 0"
        class="flex-1 flex flex-col items-center justify-center p-3 text-center font-mono space-y-1"
      >
        <span class="text-slate-400 text-lg">⚠️</span>
        <div class="text-3xs font-bold text-slate-700 dark:text-slate-300">
          No valid wind telemetry in selected timeframe ({{ timeframe }})
        </div>
        <p class="text-4xs text-slate-500 dark:text-slate-400">
          Samples may be offline or flagged as invalid by quality filters.
        </p>
      </div>

      <!-- State C: Polar Wind Rose Chart -->
      <div v-else class="flex-1 flex flex-col sm:flex-row items-center justify-between gap-2 relative">
        <!-- SVG Polar Wind Rose Container -->
        <div class="relative w-full max-w-[220px] aspect-square flex items-center justify-center mx-auto">
          <svg viewBox="0 0 320 320" class="w-full h-full overflow-visible">
            <!-- Concentric Frequency Grid Rings -->
            <g class="frequency-rings" :stroke="gridColor" fill="none" stroke-width="0.8">
              <circle
                v-for="(r, idx) in ringRadii"
                :key="'ring-' + idx"
                :cx="center"
                :cy="center"
                :r="r"
                stroke-dasharray="2 2"
              />
              <!-- Outer Boundary Circle -->
              <circle :cx="center" :cy="center" :r="maxRadius" stroke-width="1.2" :stroke="axisColor" />
              <!-- Calm Center Circle -->
              <circle :cx="center" :cy="center" :r="calmRadius" fill="rgba(148, 163, 184, 0.2)" :stroke="axisColor" />
            </g>

            <!-- Radial Compass Direction Axes -->
            <g class="compass-axes" :stroke="gridColor" stroke-width="0.7">
              <line
                v-for="deg in [0, 45, 90, 135, 180, 225, 270, 315]"
                :key="'axis-' + deg"
                :x1="center"
                :y1="center"
                :x2="center + maxRadius * Math.sin(deg * Math.PI / 180)"
                :y2="center - maxRadius * Math.cos(deg * Math.PI / 180)"
              />
            </g>

            <!-- Polar Petals (16 Compass Sectors Stacked by Speed Bin) -->
            <g class="petals">
              <g
                v-for="sector in sectorSlices"
                :key="'sector-' + sector.dirIndex"
                @mouseenter="hoveredSector = sector"
                @mouseleave="hoveredSector = null"
                class="cursor-pointer transition-opacity hover:opacity-90"
              >
                <path
                  v-for="(bin, bIdx) in sector.bins"
                  :key="'bin-' + bIdx"
                  :d="bin.pathD"
                  :fill="bin.color"
                  fill-opacity="0.85"
                  :stroke="bin.color"
                  stroke-width="0.5"
                />
              </g>
            </g>

            <!-- Calm Center Percentage Text -->
            <text
              :x="center"
              :y="center + 3"
              text-anchor="middle"
              :fill="textColor"
              font-size="9"
              font-family="monospace"
              font-weight="bold"
            >
              {{ calmPercentage.toFixed(1) }}%
            </text>

            <!-- Frequency Ring Labels (Radial %) -->
            <g class="ring-labels" font-size="7" font-family="monospace" :fill="mutedTextColor" text-anchor="middle">
              <text
                v-for="(pct, idx) in ringPercentages"
                :key="'lbl-' + idx"
                :x="center"
                :y="center - ringRadii[idx] + 8"
              >
                {{ pct }}%
              </text>
            </g>

            <!-- Compass Cardinal Labels (N, NE, E, SE, S, SW, W, NW) -->
            <g class="cardinal-labels" font-size="9" font-family="monospace" font-weight="bold" :fill="cardinalColor" text-anchor="middle">
              <text :x="center" :y="center - maxRadius - 6">N</text>
              <text :x="center + (maxRadius + 10) * 0.707" :y="center - (maxRadius + 10) * 0.707 + 3">NE</text>
              <text :x="center + maxRadius + 12" :y="center + 3">E</text>
              <text :x="center + (maxRadius + 10) * 0.707" :y="center + (maxRadius + 10) * 0.707 + 7">SE</text>
              <text :x="center" :y="center + maxRadius + 12">S</text>
              <text :x="center - (maxRadius + 10) * 0.707" :y="center + (maxRadius + 10) * 0.707 + 7">SW</text>
              <text :x="center - maxRadius - 12" :y="center + 3">W</text>
              <text :x="center - (maxRadius + 10) * 0.707" :y="center - (maxRadius + 10) * 0.707 + 3">NW</text>
            </g>
          </svg>

          <!-- Floating Hover Tooltip -->
          <div
            v-if="hoveredSector"
            class="absolute z-30 pointer-events-none bg-slate-900/95 text-white border border-cyan-500 rounded px-2 py-1 shadow-xl text-4xs font-mono"
            style="bottom: 0px; left: 50%; transform: translateX(-50%);"
          >
            <div class="font-bold text-cyan-400">Sector {{ hoveredSector.label }} ({{ hoveredSector.degRange }})</div>
            <div>Freq: {{ hoveredSector.totalPercent.toFixed(1) }}% ({{ hoveredSector.totalCount }} obs)</div>
            <div class="text-3xs text-slate-300">Dominant: {{ hoveredSector.dominantSpeedBin }}</div>
          </div>
        </div>

        <!-- Legend & Speed Bins Side Panel -->
        <div class="flex flex-col justify-center space-y-1 text-4xs font-mono pr-1 min-w-[105px]">
          <div class="text-slate-400 dark:text-slate-500 uppercase font-semibold tracking-wider text-4xs mb-0.5">
            Speed ({{ speedUnit }})
          </div>

          <div
            v-for="b in speedBins"
            :key="b.label"
            class="flex items-center space-x-1.5"
          >
            <span class="w-2.5 h-2 rounded-xs flex-shrink-0" :style="{ backgroundColor: b.color }"></span>
            <span class="text-slate-700 dark:text-slate-300 font-semibold">{{ b.label }}</span>
          </div>

          <!-- Total Sample Metrics -->
          <div class="pt-1.5 mt-1 border-t border-slate-100 dark:border-[#1a2d42] text-slate-500 dark:text-slate-400 space-y-0.5">
            <div>Valid N: <strong class="text-slate-800 dark:text-slate-200">{{ totalValidSamples }}</strong></div>
            <div>Calm: <strong class="text-slate-800 dark:text-slate-200">{{ calmPercentage.toFixed(1) }}%</strong></div>
            <div v-if="lastObservationTime" class="truncate" :title="lastObservationTime">
              Obs: <span class="text-slate-700 dark:text-slate-300">{{ formatRelativeTime(lastObservationTime) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer Note on Convention -->
      <div class="pt-1.5 border-t border-slate-100 dark:border-[#1a2d42] flex items-center justify-between text-4xs font-mono text-slate-400 dark:text-slate-500">
        <span>Convention: <strong class="text-cyan-600 dark:text-cyan-400">Wind Origin (Blowing FROM)</strong></span>
        <span>Denominator: Valid N ({{ totalValidSamples }})</span>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

// 16 Compass Direction Labels (Clockwise from North)
const COMPASS_16 = [
  'N', 'NNE', 'NE', 'ENE',
  'E', 'ESE', 'SE', 'SSE',
  'S', 'SSW', 'SW', 'WSW',
  'W', 'WNW', 'NW', 'NNW'
];

// Standard Industrial Wind Speed Bins
const SPEED_BINS = [
  { min: 0.0, max: 0.5, label: '< 0.5 (Calm)', color: '#94A3B8' },
  { min: 0.5, max: 2.0, label: '0.5 - 2.0', color: '#00CFFF' },
  { min: 2.0, max: 4.0, label: '2.0 - 4.0', color: '#0284c7' },
  { min: 4.0, max: 6.0, label: '4.0 - 6.0', color: '#00D6A0' },
  { min: 6.0, max: 8.0, label: '6.0 - 8.0', color: '#FFB020' },
  { min: 8.0, max: 999, label: '> 8.0', color: '#FF4559' },
];

export default {
  name: 'WindRoseChart',
  props: {
    devicesList: {
      type: Array,
      default: () => [],
    },
    isDark: {
      type: Boolean,
      default: true,
    },
    liveMap: {
      type: Object,
      default: () => ({}),
    },
  },
  data() {
    return {
      selectedDeviceId: null,
      timeframe: '24h',
      isLoading: false,
      windSpeedParam: null,
      windDirParam: null,
      speedUnit: 'm/s',
      pairedSamples: [], // [{ speed, dir, time, quality }]
      hoveredSector: null,
      lastObservationTime: null,
      center: 160,
      calmRadius: 18,
      maxRadius: 110,
    };
  },
  computed: {
    availableDevices() {
      return (this.devicesList || []).filter(d => d.parameters && d.parameters.length > 0);
    },
    currentDevice() {
      if (!this.selectedDeviceId) return null;
      return this.availableDevices.find(d => d.id === Number(this.selectedDeviceId)) || null;
    },
    currentDeviceCode() {
      return this.currentDevice ? (this.currentDevice.device_code || this.currentDevice.code) : 'None';
    },
    hasWindSensors() {
      return !!(this.windSpeedParam && this.windDirParam);
    },
    speedBins() {
      return SPEED_BINS;
    },
    totalValidSamples() {
      return this.pairedSamples.length;
    },
    calmCount() {
      return this.pairedSamples.filter(s => s.speed < 0.5).length;
    },
    calmPercentage() {
      if (this.totalValidSamples === 0) return 0;
      return (this.calmCount / this.totalValidSamples) * 100;
    },
    // Polar Grid Styling Tokens based on Theme
    gridColor() {
      return this.isDark ? '#1C2C40' : '#E2E8F0';
    },
    axisColor() {
      return this.isDark ? '#2E4460' : '#CBD5E1';
    },
    textColor() {
      return this.isDark ? '#F1F5F9' : '#0F172A';
    },
    mutedTextColor() {
      return this.isDark ? '#64748B' : '#94A3B8';
    },
    cardinalColor() {
      return this.isDark ? '#00CFFF' : '#0284c7';
    },
    // Dynamic Frequency Scale Rings
    ringPercentages() {
      return [5, 10, 15, 20, 25];
    },
    ringRadii() {
      const availRadius = this.maxRadius - this.calmRadius;
      return this.ringPercentages.map(pct => {
        return this.calmRadius + (pct / 25) * availRadius;
      });
    },
    // Calculated Slices for the 16 Compass Sectors
    sectorSlices() {
      if (this.totalValidSamples === 0) return [];

      const sectorAngle = 360 / 16; // 22.5 deg
      const availRadius = this.maxRadius - this.calmRadius;

      // Group paired samples into 16 compass sectors (excluding calm samples from directional petals)
      const nonCalm = this.pairedSamples.filter(s => s.speed >= 0.5);

      return COMPASS_16.map((label, dirIdx) => {
        // Sector angle center in degrees: 0=N, 22.5=NNE, 45=NE, etc.
        const centerDeg = dirIdx * sectorAngle;
        const halfAngle = sectorAngle / 2;

        // Collect samples whose wind direction falls into this sector
        const sectorSamples = nonCalm.filter(s => {
          let diff = Math.abs(s.dir - centerDeg);
          if (diff > 180) diff = 360 - diff;
          return diff < halfAngle;
        });

        const totalSectorCount = sectorSamples.length;
        const totalPercent = this.totalValidSamples > 0
          ? (totalSectorCount / this.totalValidSamples) * 100
          : 0;

        // Build stacked bins from lowest speed to highest
        let currentR = this.calmRadius;
        const bins = [];
        let dominantBinLabel = '--';
        let maxBinCount = -1;

        // Speed bins starting from index 1 (since index 0 is calm)
        for (let b = 1; b < SPEED_BINS.length; b++) {
          const binConfig = SPEED_BINS[b];
          const binSamples = sectorSamples.filter(s => s.speed >= binConfig.min && s.speed < binConfig.max);
          const binCount = binSamples.length;

          if (binCount > maxBinCount) {
            maxBinCount = binCount;
            dominantBinLabel = binConfig.label;
          }

          if (binCount > 0) {
            const binPercent = (binCount / this.totalValidSamples) * 100;
            // Radius increment proportional to percentage (scaled to 25% max scale)
            const rDelta = (binPercent / 25) * availRadius;
            const rInner = currentR;
            const rOuter = Math.min(this.maxRadius + 15, currentR + rDelta);

            const pathD = this.createSectorPath(
              this.center,
              this.center,
              rInner,
              rOuter,
              centerDeg - halfAngle + 1,
              centerDeg + halfAngle - 1
            );

            bins.push({
              pathD,
              color: binConfig.color,
              count: binCount,
              percent: binPercent,
              label: binConfig.label,
            });

            currentR = rOuter;
          }
        }

        const degMin = ((centerDeg - halfAngle + 360) % 360).toFixed(1);
        const degMax = ((centerDeg + halfAngle) % 360).toFixed(1);

        return {
          dirIndex: dirIdx,
          label,
          centerDeg,
          degRange: `${degMin}° - ${degMax}°`,
          totalCount: totalSectorCount,
          totalPercent,
          dominantSpeedBin: dominantBinLabel,
          bins,
        };
      });
    },
  },
  watch: {
    devicesList: {
      immediate: true,
      handler(newDevs) {
        if (!this.selectedDeviceId && newDevs && newDevs.length > 0) {
          // Find first device that has wind sensors
          const devWithWind = newDevs.find(d => this.findWindParamsInDevice(d).hasWind);
          this.selectedDeviceId = devWithWind ? devWithWind.id : newDevs[0].id;
        }
        this.inspectSelectedDeviceSensors();
      },
    },
  },
  mounted() {
    this.inspectSelectedDeviceSensors();
  },
  methods: {
    findWindParamsInDevice(dev) {
      if (!dev || !dev.parameters) return { hasWind: false, wd: null, ws: null };
      let wd = null;
      let ws = null;

      for (const p of dev.parameters) {
        const code = (p.parameter_code || p.code || '').toLowerCase();
        const name = (p.parameter_name || p.name || '').toLowerCase();

        // Direction detection
        if (
          code === 'wd' || code === 'wind_dir' || code === 'wind_direction' ||
          name.includes('wind dir') || name.includes('wind direction') || name.includes('arah angin')
        ) {
          wd = p;
        }

        // Speed detection
        if (
          code === 'ws' || code === 'wind_speed' || code === 'speed' ||
          name.includes('wind speed') || name.includes('kecepatan angin')
        ) {
          ws = p;
        }
      }

      return {
        hasWind: !!(wd && ws),
        wd,
        ws,
      };
    },
    inspectSelectedDeviceSensors() {
      if (!this.currentDevice) return;
      const detected = this.findWindParamsInDevice(this.currentDevice);
      this.windDirParam = detected.wd;
      this.windSpeedParam = detected.ws;
      if (this.windSpeedParam && this.windSpeedParam.unit) {
        this.speedUnit = this.windSpeedParam.unit;
      }

      if (this.hasWindSensors) {
        this.fetchWindData();
      } else {
        this.pairedSamples = [];
      }
    },
    onDeviceChange() {
      this.inspectSelectedDeviceSensors();
    },
    setTimeframe(tf) {
      this.timeframe = tf;
      if (this.hasWindSensors) {
        this.fetchWindData();
      }
    },
    async fetchWindData() {
      if (!this.currentDevice || !this.hasWindSensors) return;
      this.isLoading = true;

      try {
        const now = new Date();
        let startTime = new Date(now.getTime() - 24 * 3600 * 1000);
        if (this.timeframe === '1h') startTime = new Date(now.getTime() - 3600 * 1000);
        else if (this.timeframe === '6h') startTime = new Date(now.getTime() - 6 * 3600 * 1000);
        else if (this.timeframe === '7d') startTime = new Date(now.getTime() - 7 * 86400 * 1000);

        // Fetch historical observations for Wind Speed and Direction
        const [wsRes, wdRes] = await Promise.all([
          axios.get(`/api/devices/${this.currentDevice.id}/telemetry/history`, {
            params: {
              parameter_id: this.windSpeedParam.id,
              start_time: startTime.toISOString(),
              end_time: now.toISOString(),
              page_size: 200,
            },
          }).catch(() => null),
          axios.get(`/api/devices/${this.currentDevice.id}/telemetry/history`, {
            params: {
              parameter_id: this.windDirParam.id,
              start_time: startTime.toISOString(),
              end_time: now.toISOString(),
              page_size: 200,
            },
          }).catch(() => null),
        ]);

        const wsItems = (wsRes && wsRes.data && wsRes.data.data && wsRes.data.data.items) || [];
        const wdItems = (wdRes && wdRes.data && wdRes.data.data && wdRes.data.data.items) || [];

        // Pair observations by timestamp (within 60s tolerance)
        // Exclude samples with BAD or UNCERTAIN quality per rules
        const paired = [];
        let latestTime = null;

        for (const ws of wsItems) {
          if (ws.quality === 'BAD' || ws.quality === 'UNCERTAIN') continue;
          const wsTime = new Date(ws.received_at || ws.timestamp).getTime();
          const speedVal = Number(ws.value);
          if (isNaN(speedVal) || speedVal < 0) continue;

          // Find closest matching WD sample
          let bestWd = null;
          let minDiff = 60000; // 60s max tolerance
          for (const wd of wdItems) {
            if (wd.quality === 'BAD' || wd.quality === 'UNCERTAIN') continue;
            const wdTime = new Date(wd.received_at || wd.timestamp).getTime();
            const diff = Math.abs(wsTime - wdTime);
            if (diff < minDiff) {
              minDiff = diff;
              bestWd = wd;
            }
          }

          if (bestWd) {
            let dirVal = Number(bestWd.value);
            if (!isNaN(dirVal) && dirVal >= 0 && dirVal <= 360) {
              // Normalize direction modulo 360
              dirVal = dirVal % 360;
              paired.push({
                speed: speedVal,
                dir: dirVal,
                time: ws.received_at || ws.timestamp,
              });

              if (!latestTime || new Date(ws.received_at).getTime() > new Date(latestTime).getTime()) {
                latestTime = ws.received_at || ws.timestamp;
              }
            }
          }
        }

        this.pairedSamples = paired;
        this.lastObservationTime = latestTime;
      } catch (err) {
        console.error('Failed to load wind history:', err);
      } finally {
        this.isLoading = false;
      }
    },
    // Helper to generate SVG Annular Sector path string (donut wedge)
    createSectorPath(cx, cy, rInner, rOuter, startAngleDeg, endAngleDeg) {
      // Convert to math angles (0 is North / up)
      const toRad = deg => ((deg - 90) * Math.PI) / 180;
      const startRad = toRad(startAngleDeg);
      const endRad = toRad(endAngleDeg);

      const x1 = cx + rOuter * Math.cos(startRad);
      const y1 = cy + rOuter * Math.sin(startRad);
      const x2 = cx + rOuter * Math.cos(endRad);
      const y2 = cy + rOuter * Math.sin(endRad);

      const x3 = cx + rInner * Math.cos(endRad);
      const y3 = cy + rInner * Math.sin(endRad);
      const x4 = cx + rInner * Math.cos(startRad);
      const y4 = cy + rInner * Math.sin(startRad);

      const arcSweep = endAngleDeg - startAngleDeg <= 180 ? 0 : 1;

      return `M ${x1} ${y1} A ${rOuter} ${rOuter} 0 ${arcSweep} 1 ${x2} ${y2} L ${x3} ${y3} A ${rInner} ${rInner} 0 ${arcSweep} 0 ${x4} ${y4} Z`;
    },
    formatRelativeTime(iso) {
      if (!iso) return '--';
      const sec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
      if (sec < 5) return 'just now';
      if (sec < 60) return `${sec}s ago`;
      if (sec < 3600) return `${Math.floor(sec / 60)}m ago`;
      return `${Math.floor(sec / 3600)}h ago`;
    },
  },
};
</script>

<style scoped>
.custom-scroller::-webkit-scrollbar {
  width: 4px;
}
.custom-scroller::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.3);
  border-radius: 4px;
}

/* SCADA Light Mode Border Enforcement */
html:not(.dark) .scada-panel {
  border-color: #334155;
}
</style>
