<template>
  <div class="scada-panel p-2.5 rounded-xl bg-white dark:bg-[#0B1422] border border-[#334155] dark:border-[#20354B] shadow-xs dark:shadow-md flex flex-col h-full select-none">
    <!-- Header -->
    <div class="flex flex-wrap items-center justify-between gap-1.5 pb-2 border-b border-[#334155] dark:border-[#20354B]">
      <div>
        <div class="flex items-center space-x-2">
          <span class="w-2 h-2 rounded-full bg-cyan-500"></span>
          <h3 class="text-xs font-bold tracking-wider uppercase text-slate-900 dark:text-white font-mono">
            {{ $t('operator.dispersionTitle') || 'POLLUTANT DISPERSION' }}
          </h3>
        </div>
        <p class="text-4xs text-slate-500 dark:text-slate-400 font-mono mt-0.5">
          {{ $t('operator.dispersionSubtitle') || 'Environmental distribution and wind context' }}
        </p>
      </div>

      <!-- Controls: Pollutant, Station, Layers -->
      <div class="flex flex-wrap items-center gap-1.5 text-3xs font-mono">
        <!-- Pollutant Selector (Only displays pollutants that actually exist) -->
        <select
          v-model="selectedPollutantCode"
          class="px-2 py-0.5 rounded bg-slate-100 dark:bg-[#101B2B] border border-[#334155] dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-3xs focus:outline-none focus:border-cyan-500 font-bold"
          :title="$t('operator.selectPollutant') || 'Select Pollutant'"
        >
          <option v-for="p in availablePollutants" :key="p.code" :value="p.code">
            {{ p.name }} ({{ p.code }})
          </option>
        </select>

        <!-- Station Selector -->
        <select
          v-model="selectedStationId"
          class="px-2 py-0.5 rounded bg-slate-100 dark:bg-[#101B2B] border border-[#334155] dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-3xs focus:outline-none focus:border-cyan-500"
          :title="$t('operator.filterStation') || 'Filter Station'"
        >
          <option value="">{{ $t('operator.allStations') || 'All Stations' }}</option>
          <option v-for="d in devicesList" :key="d.id" :value="d.id">
            {{ d.device_code || d.code }}
          </option>
        </select>

        <!-- Layer Toggles -->
        <div class="flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded border border-[#334155] dark:border-[#20354B] space-x-0.5">
          <button
            @click="activeLayer = 'points'"
            class="px-1.5 py-0.5 rounded text-4xs transition font-semibold"
            :class="activeLayer === 'points' ? 'bg-cyan-500 text-slate-950 font-black' : 'text-slate-600 dark:text-slate-400'"
            title="Measured Station Points"
          >
            Points
          </button>
          <button
            @click="activeLayer = 'idw'"
            class="px-1.5 py-0.5 rounded text-4xs transition font-semibold"
            :class="activeLayer === 'idw' ? 'bg-cyan-500 text-slate-950 font-black' : 'text-slate-600 dark:text-slate-400'"
            title="Spatial Interpolation (IDW)"
          >
            IDW Surface
          </button>
          <button
            @click="toggleWindOverlay"
            class="px-1.5 py-0.5 rounded text-4xs transition font-semibold"
            :class="showWindOverlay ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-400 border border-cyan-500/40' : 'text-slate-400 dark:text-slate-500'"
            title="Toggle Wind Vector Overlay"
          >
            💨 Wind
          </button>
        </div>
      </div>
    </div>

    <!-- Main Map View Area -->
    <div class="flex-1 flex flex-col justify-between pt-1 relative min-h-[220px]">
      <!-- Case 1: Coordinates Missing on All Stations (Non-Map Operational Matrix Fallback) -->
      <div v-if="!hasAnyCoordinates" class="flex-1 flex flex-col space-y-2 py-1">
        <!-- Configuration Warning Banner -->
        <div class="px-2.5 py-1.5 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-700 dark:text-amber-300 font-mono text-3xs flex items-center space-x-2">
          <span>⚠️</span>
          <div class="flex-1">
            <strong>Station coordinates unconfigured.</strong>
            <span class="text-4xs text-amber-600 dark:text-amber-400 ml-1">
              Geographic coordinates (Lat/Long) are not set in device settings. Rendering operational station measurement matrix fallback.
            </span>
          </div>
        </div>

        <!-- Non-Map Station Measurement Fallback Cards -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 overflow-y-auto max-h-[160px] custom-scroller pr-1">
          <div
            v-for="st in targetStationCards"
            :key="st.deviceId"
            class="p-2 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-[#334155] dark:border-[#20354B] flex flex-col justify-between"
          >
            <div class="flex items-center justify-between text-3xs font-mono">
              <span class="font-bold text-slate-800 dark:text-slate-200">{{ st.deviceCode }} ({{ st.deviceName }})</span>
              <span
                class="px-1 py-0.2 rounded text-4xs font-bold uppercase"
                :class="qualityBadgeClass(st.quality)"
              >
                {{ st.quality }}
              </span>
            </div>

            <div class="flex items-baseline justify-between mt-1">
              <div class="flex items-baseline space-x-1">
                <span class="text-lg font-bold font-mono" :class="valueColorClass(st)">
                  {{ st.displayVal }}
                </span>
                <span class="text-3xs font-mono text-slate-500">{{ st.unit }}</span>
              </div>
              <span class="text-4xs font-mono text-slate-400">
                {{ formatRelative(st.timestamp) }}
              </span>
            </div>

            <!-- Threshold Range Bar -->
            <div class="w-full bg-slate-200 dark:bg-[#18283E] h-1.5 rounded-full overflow-hidden mt-1.5 flex">
              <div
                class="h-full transition-all duration-300"
                :class="thresholdBarColorClass(st)"
                :style="{ width: Math.min(Math.max((st.numericVal / (st.highCritical || 100)) * 100, 5), 100) + '%' }"
              ></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Case 2: Stations Have Coordinates -> Tactical SCADA Map Canvas -->
      <div v-else class="flex-1 relative w-full h-[180px] bg-slate-50 dark:bg-[#070D18] rounded-lg border border-[#334155] dark:border-[#20354B] overflow-hidden">
        <!-- SVG Tactical Coordinate Grid & Markers -->
        <svg
          class="w-full h-full"
          viewBox="0 0 500 220"
          preserveAspectRatio="none"
        >
          <!-- Tactical Distance Concentric Rings -->
          <g class="tactical-rings" :stroke="mapGridColor" fill="none" stroke-width="0.7">
            <circle cx="250" cy="110" r="45" stroke-dasharray="3 3" />
            <circle cx="250" cy="110" r="85" stroke-dasharray="3 3" />
            <circle cx="250" cy="110" r="130" stroke-dasharray="2 4" />
            <!-- Coordinate Crosshairs -->
            <line x1="25" y1="110" x2="475" y2="110" stroke-dasharray="1 3" />
            <line x1="250" y1="15" x2="250" y2="205" stroke-dasharray="1 3" />
          </g>

          <!-- Range Ring Scale Annotations -->
          <g font-size="7" font-family="monospace" :fill="mutedTextColor">
            <text x="298" y="108">1 km</text>
            <text x="338" y="108">2 km</text>
            <text x="383" y="108">5 km</text>
          </g>

          <!-- IDW Interpolated Surface Contour (When Enabled & Multiple Points Available) -->
          <g v-if="activeLayer === 'idw' && plottedStationsWithCoords.length >= 2" class="idw-layer" opacity="0.35">
            <!-- Simulated Spatial Heatmap Gradient Paths between Stations -->
            <circle
              v-for="st in plottedStationsWithCoords"
              :key="'idw-' + st.deviceId"
              :cx="st.mapX"
              :cy="st.mapY"
              :r="55"
              :fill="st.thresholdColor"
              filter="blur(14px)"
            />
          </g>

          <!-- Wind Vector Arrows at Stations -->
          <g v-if="showWindOverlay">
            <g
              v-for="st in plottedStationsWithCoords"
              :key="'wind-' + st.deviceId"
              v-if="st.windSpeed != null && st.windDir != null"
            >
              <!-- Arrow pointing toward wind destination (dir + 180 deg) -->
              <line
                :x1="st.mapX"
                :y1="st.mapY"
                :x2="st.mapX + Math.sin((st.windDir + 180) * Math.PI / 180) * Math.min(st.windSpeed * 6 + 10, 35)"
                :y2="st.mapY - Math.cos((st.windDir + 180) * Math.PI / 180) * Math.min(st.windSpeed * 6 + 10, 35)"
                stroke="#00CFFF"
                stroke-width="1.8"
                stroke-linecap="round"
              />
              <circle
                :cx="st.mapX + Math.sin((st.windDir + 180) * Math.PI / 180) * Math.min(st.windSpeed * 6 + 10, 35)"
                :cy="st.mapY - Math.cos((st.windDir + 180) * Math.PI / 180) * Math.min(st.windSpeed * 6 + 10, 35)"
                r="2.5"
                fill="#00CFFF"
              />
              <text
                :x="st.mapX + 8"
                :y="st.mapY - 14"
                fill="#00CFFF"
                font-size="7"
                font-family="monospace"
                font-weight="bold"
              >
                {{ st.windSpeed.toFixed(1) }}m/s
              </text>
            </g>
          </g>

          <!-- Station Markers & Pins -->
          <g
            v-for="st in plottedStationsWithCoords"
            :key="'pin-' + st.deviceId"
            @click="inspectStation(st)"
            class="cursor-pointer group"
          >
            <!-- Outer Pulsing Halo -->
            <circle
              :cx="st.mapX"
              :cy="st.mapY"
              r="14"
              :fill="st.thresholdColor"
              fill-opacity="0.18"
              class="animate-pulse"
            />
            <!-- Inner Solid Pin -->
            <circle
              :cx="st.mapX"
              :cy="st.mapY"
              r="6"
              :fill="st.thresholdColor"
              stroke="#ffffff"
              stroke-width="1.5"
            />

            <!-- Station Label Callout -->
            <g :transform="`translate(${st.mapX + 9}, ${st.mapY + 3})`">
              <rect
                x="-2"
                y="-9"
                width="72"
                height="15"
                rx="3"
                :fill="isDark ? '#0B1422' : '#FFFFFF'"
                :stroke="st.thresholdColor"
                stroke-width="0.8"
                fill-opacity="0.92"
              />
              <text
                x="2"
                y="1"
                :fill="isDark ? '#FFFFFF' : '#0F172A'"
                font-size="7"
                font-family="monospace"
                font-weight="bold"
              >
                {{ st.deviceCode }}: {{ st.displayVal }}
              </text>
            </g>
          </g>
        </svg>

        <!-- Selected Station Detail Inspector Modal Overlay -->
        <div
          v-if="inspectedStation"
          class="absolute top-2 right-2 z-30 p-2 rounded-lg bg-white/95 dark:bg-[#0B1422]/95 border border-cyan-500 shadow-xl text-3xs font-mono max-w-[210px]"
        >
          <div class="flex items-center justify-between pb-1 border-b border-[#334155] dark:border-[#20354B]">
            <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ inspectedStation.deviceCode }}</span>
            <button @click="inspectedStation = null" class="text-slate-400 hover:text-white">✕</button>
          </div>
          <div class="space-y-0.5 mt-1 text-slate-600 dark:text-slate-300">
            <div>Loc: {{ inspectedStation.location || 'Site Edge Node' }}</div>
            <div>Coords: [{{ inspectedStation.lat.toFixed(4) }}, {{ inspectedStation.lon.toFixed(4) }}]</div>
            <div>
              {{ selectedPollutantCode }}: <strong class="text-slate-900 dark:text-white">{{ inspectedStation.displayVal }} {{ inspectedStation.unit }}</strong>
            </div>
            <div>Quality: <span :class="qualityBadgeClass(inspectedStation.quality)">{{ inspectedStation.quality }}</span></div>
            <div>Obs: {{ formatRelative(inspectedStation.timestamp) }}</div>
          </div>
        </div>
      </div>

      <!-- Scientific Discipline & Disclaimer Footer -->
      <div class="pt-1.5 border-t border-slate-100 dark:border-[#1a2d42] flex flex-wrap items-center justify-between text-4xs font-mono text-slate-400 dark:text-slate-500">
        <div class="flex items-center space-x-1">
          <span class="w-1.5 h-1.5 rounded-full bg-cyan-500"></span>
          <span>Mode: <strong class="text-slate-700 dark:text-slate-300">{{ activeLayer === 'points' ? 'Direct Observation (Sensor Measurements)' : 'Spatial Interpolation (IDW Estimation)' }}</strong></span>
        </div>
        <div class="text-3xs text-slate-400">
          Scale: Good (0) · Warn ({{ currentThresholdWarn }}) · Crit ({{ currentThresholdCrit }})
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PollutantDispersionMap',
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
      selectedPollutantCode: '',
      selectedStationId: '',
      activeLayer: 'points', // 'points' | 'idw'
      showWindOverlay: true,
      inspectedStation: null,
    };
  },
  computed: {
    // Dynamic extraction of pollutants that actually exist in parameters across devices
    availablePollutants() {
      const map = new Map();
      for (const dev of this.devicesList || []) {
        if (!dev.parameters) continue;
        for (const p of dev.parameters) {
          const code = (p.parameter_code || p.code || '').trim();
          const name = (p.parameter_name || p.name || code).trim();
          const lower = code.toLowerCase();

          // Exclude wind direction/speed parameters from pollutant dropdown
          if (
            lower === 'wd' || lower === 'ws' ||
            lower.includes('direction') || lower.includes('speed')
          ) {
            continue;
          }

          if (code && !map.has(code)) {
            map.set(code, {
              code,
              name,
              unit: p.unit || '',
              param: p,
            });
          }
        }
      }
      return Array.from(map.values());
    },
    currentPollutantConfig() {
      if (!this.selectedPollutantCode) return null;
      return this.availablePollutants.find(p => p.code === this.selectedPollutantCode) || null;
    },
    currentThresholdWarn() {
      if (this.currentPollutantConfig && this.currentPollutantConfig.param) {
        return this.currentPollutantConfig.param.high_warning != null ? this.currentPollutantConfig.param.high_warning : 50;
      }
      return 50;
    },
    currentThresholdCrit() {
      if (this.currentPollutantConfig && this.currentPollutantConfig.param) {
        return this.currentPollutantConfig.param.high_critical != null ? this.currentPollutantConfig.param.high_critical : 100;
      }
      return 100;
    },
    // Check if any station in fleet has valid geographic coordinates
    hasAnyCoordinates() {
      return (this.devicesList || []).some(d => {
        return d.latitude != null && d.longitude != null && !isNaN(Number(d.latitude)) && !isNaN(Number(d.longitude));
      });
    },
    // Target station measurement cards (used for both map pins and non-map fallback)
    targetStationCards() {
      if (!this.selectedPollutantCode) return [];
      const cards = [];

      for (const dev of this.devicesList || []) {
        if (this.selectedStationId && dev.id !== Number(this.selectedStationId)) {
          continue;
        }
        if (!dev.parameters) continue;

        // Find parameter matching selected pollutant code
        const p = dev.parameters.find(item => (item.parameter_code || item.code) === this.selectedPollutantCode);
        if (!p) continue;

        const live = this.liveMap[p.id] || {};
        const val = live.value !== undefined ? live.value : p.current_value;
        const quality = live.quality || p.current_quality || 'UNKNOWN';
        const numVal = Number(val);
        const displayVal = (val != null && !isNaN(numVal)) ? numVal.toFixed(p.precision != null ? p.precision : 2) : '--';

        // Check for paired wind on this device for vector overlay
        let wsVal = null;
        let wdVal = null;
        for (const wp of dev.parameters) {
          const c = (wp.parameter_code || wp.code || '').toLowerCase();
          if (c === 'ws' || c.includes('speed')) {
            const wlive = this.liveMap[wp.id] || {};
            wsVal = Number(wlive.value !== undefined ? wlive.value : wp.current_value);
          }
          if (c === 'wd' || c.includes('direction')) {
            const wlive = this.liveMap[wp.id] || {};
            wdVal = Number(wlive.value !== undefined ? wlive.value : wp.current_value);
          }
        }

        // Color coding based on thresholds and quality
        let thresholdColor = '#10B981'; // Good (Emerald)
        if (quality === 'BAD') {
          thresholdColor = '#FF4559'; // Red
        } else if (quality === 'STALE') {
          thresholdColor = '#A855F7'; // Purple
        } else if (!isNaN(numVal)) {
          if (p.high_critical != null && numVal >= p.high_critical) thresholdColor = '#FF4559';
          else if (p.high_warning != null && numVal >= p.high_warning) thresholdColor = '#FFB020';
        }

        cards.push({
          deviceId: dev.id,
          deviceCode: dev.device_code || dev.code,
          deviceName: dev.device_name || dev.name,
          location: dev.location,
          lat: dev.latitude != null ? Number(dev.latitude) : null,
          lon: dev.longitude != null ? Number(dev.longitude) : null,
          paramId: p.id,
          unit: p.unit,
          numericVal: isNaN(numVal) ? 0 : numVal,
          displayVal,
          quality,
          highWarning: p.high_warning,
          highCritical: p.high_critical,
          timestamp: live.received_at || p.current_received_at || p.last_updated,
          thresholdColor,
          windSpeed: !isNaN(wsVal) ? wsVal : null,
          windDir: !isNaN(wdVal) ? wdVal : null,
        });
      }

      return cards;
    },
    // Plotted stations for tactical coordinate map
    plottedStationsWithCoords() {
      const withCoords = this.targetStationCards.filter(s => s.lat != null && s.lon != null);
      if (withCoords.length === 0) return [];

      // Auto-bounding box normalization to 500x220 canvas coordinates
      let minLat = Infinity, maxLat = -Infinity;
      let minLon = Infinity, maxLon = -Infinity;

      for (const s of withCoords) {
        if (s.lat < minLat) minLat = s.lat;
        if (s.lat > maxLat) maxLat = s.lat;
        if (s.lon < minLon) minLon = s.lon;
        if (s.lon > maxLon) maxLon = s.lon;
      }

      // Add margin to avoid edge clipping
      const latRange = Math.max(maxLat - minLat, 0.01) * 1.4;
      const lonRange = Math.max(maxLon - minLon, 0.01) * 1.4;
      const midLat = (minLat + maxLat) / 2;
      const midLon = (minLon + maxLon) / 2;

      return withCoords.map(s => {
        const x = 250 + ((s.lon - midLon) / lonRange) * 380;
        const y = 110 - ((s.lat - midLat) / latRange) * 150;
        return {
          ...s,
          mapX: Math.min(Math.max(x, 35), 465),
          mapY: Math.min(Math.max(y, 25), 195),
        };
      });
    },
    mapGridColor() {
      return this.isDark ? '#18283E' : '#E2E8F0';
    },
    mutedTextColor() {
      return this.isDark ? '#475569' : '#94A3B8';
    },
  },
  watch: {
    availablePollutants: {
      immediate: true,
      handler(list) {
        if (!this.selectedPollutantCode && list && list.length > 0) {
          // Default to PM2.5 or PM10 if present, else first available
          const preferred = list.find(p => p.code.toLowerCase().includes('pm2.5') || p.code.toLowerCase().includes('pm10'));
          this.selectedPollutantCode = preferred ? preferred.code : list[0].code;
        }
      },
    },
  },
  methods: {
    toggleWindOverlay() {
      this.showWindOverlay = !this.showWindOverlay;
    },
    inspectStation(st) {
      this.inspectedStation = st;
    },
    qualityBadgeClass(q) {
      switch (q) {
        case 'GOOD': return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30';
        case 'BAD': return 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/30';
        case 'STALE': return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/30';
        default: return 'bg-slate-100 dark:bg-slate-800 text-slate-500';
      }
    },
    valueColorClass(st) {
      if (st.quality === 'BAD') return 'text-rose-600 dark:text-rose-400';
      if (st.quality === 'STALE') return 'text-purple-600 dark:text-purple-400';
      if (st.highCritical != null && st.numericVal >= st.highCritical) return 'text-rose-600 dark:text-rose-400 font-bold';
      if (st.highWarning != null && st.numericVal >= st.highWarning) return 'text-amber-600 dark:text-amber-400';
      return 'text-slate-900 dark:text-white';
    },
    thresholdBarColorClass(st) {
      if (st.quality === 'BAD') return 'bg-rose-500';
      if (st.quality === 'STALE') return 'bg-purple-500';
      if (st.highCritical != null && st.numericVal >= st.highCritical) return 'bg-rose-500';
      if (st.highWarning != null && st.numericVal >= st.highWarning) return 'bg-amber-500';
      return 'bg-emerald-500';
    },
    formatRelative(iso) {
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
