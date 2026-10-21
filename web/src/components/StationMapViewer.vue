<template>
  <div class="station-map-viewer-wrapper relative w-full h-full overflow-hidden rounded-md border border-slate-200 dark:border-[#1E3048] bg-slate-100 dark:bg-[#070E1A] flex flex-col select-none" :style="{ minHeight: height || '180px' }">
    <!-- Map Canvas Element -->
    <div
      ref="mapContainer"
      class="map-container flex-1 w-full h-full z-10"
      :class="{ 'opacity-90': isTilesOnline, 'opacity-40 pointer-events-none': !isTilesOnline && hasCoordinates }"
    ></div>

    <!-- Offline / Tile Load Failure Fallback Overlay -->
    <div
      v-if="!isTilesOnline && hasCoordinates"
      class="absolute inset-0 z-20 flex flex-col items-center justify-center p-3 text-center pointer-events-auto bg-slate-900/60 backdrop-blur-xs font-mono"
    >
      <!-- SCADA Crosshair Radar Grid Schematic -->
      <div class="relative w-28 h-28 mb-2 rounded-full border border-cyan-500/40 flex items-center justify-center bg-cyan-950/20 shadow-inner">
        <div class="absolute inset-0 rounded-full border border-cyan-500/20 scale-75"></div>
        <div class="absolute inset-0 rounded-full border border-cyan-500/10 scale-50"></div>
        <div class="absolute w-full h-px bg-cyan-500/30"></div>
        <div class="absolute h-full w-px bg-cyan-500/30"></div>
        <!-- Pulsing Station Marker Icon in Center -->
        <div class="relative z-10 flex flex-col items-center text-cyan-400 animate-bounce">
          <svg class="w-7 h-7 drop-shadow-[0_0_8px_rgba(6,182,212,0.8)]" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
          </svg>
        </div>
      </div>

      <div class="text-[11px] font-bold text-white flex items-center space-x-1.5 mb-0.5">
        <span class="w-2 h-2 rounded-full bg-amber-400 animate-pulse"></span>
        <span>Local Schematic Fallback</span>
      </div>
      <div class="text-[9.5px] text-slate-300 max-w-xs leading-tight mb-2">
        Online map tiles unreachable (offline edge mode). Station coordinates remain verified & active.
      </div>
      <div class="px-2 py-1 rounded bg-black/50 border border-cyan-500/30 text-[10px] text-cyan-300 font-bold tabular-nums">
        Lat: {{ formattedLat }}, Lng: {{ formattedLng }}
      </div>
    </div>

    <!-- Unconfigured State Overlay -->
    <div
      v-else-if="!hasCoordinates"
      class="absolute inset-0 z-20 flex flex-col items-center justify-center p-3 text-center bg-slate-50/90 dark:bg-[#070E1A]/90 font-mono"
    >
      <div class="w-10 h-10 rounded-full bg-amber-500/10 border border-amber-500/30 text-amber-500 flex items-center justify-center mb-1.5">
        <svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
          <circle cx="12" cy="10" r="3" />
        </svg>
      </div>
      <div class="text-xs font-bold text-slate-800 dark:text-slate-200">
        {{ $t('operator.locationUnconfigured') || 'Location Unconfigured' }}
      </div>
      <p class="text-[9.5px] text-slate-500 dark:text-slate-400 mt-1 max-w-[220px]">
        {{ $t('operator.locationUnconfiguredDesc') || 'Set station coordinates using GPS, IP geolocation, or manual map pin.' }}
      </p>
    </div>

    <!-- Quick Floating Map Controls -->
    <div v-if="hasCoordinates" class="absolute top-2 right-2 z-30 flex flex-col space-y-1">
      <button
        type="button"
        @click="recenterMap"
        class="w-6 h-6 flex items-center justify-center rounded bg-white/90 dark:bg-[#101B2B]/90 hover:bg-cyan-500 hover:text-black text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-[#20354B] shadow-xs text-xs transition"
        title="Recenter Map to Station"
      >
        🎯
      </button>
      <button
        type="button"
        @click="zoomIn"
        class="w-6 h-6 flex items-center justify-center rounded bg-white/90 dark:bg-[#101B2B]/90 hover:bg-cyan-500 hover:text-black text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-[#20354B] shadow-xs text-xs font-bold transition"
        title="Zoom In"
      >
        +
      </button>
      <button
        type="button"
        @click="zoomOut"
        class="w-6 h-6 flex items-center justify-center rounded bg-white/90 dark:bg-[#101B2B]/90 hover:bg-cyan-500 hover:text-black text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-[#20354B] shadow-xs text-xs font-bold transition"
        title="Zoom Out"
      >
        -
      </button>
    </div>

    <!-- Tile Status Pill (Bottom Left) -->
    <div class="absolute bottom-1.5 left-2 z-30 flex items-center space-x-1.5 px-1.5 py-0.5 rounded bg-black/60 backdrop-blur-xs border border-white/10 text-[8.5px] font-mono text-slate-300">
      <span
        class="w-1.5 h-1.5 rounded-full"
        :class="isTilesOnline ? 'bg-emerald-400' : 'bg-amber-400'"
      ></span>
      <span>{{ isTilesOnline ? 'OSM Online Tiles' : 'Offline Mode' }}</span>
      <span v-if="accuracyMeters" class="text-cyan-400 ml-1">±{{ accuracyMeters }}m</span>
    </div>
  </div>
</template>

<script>
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';

export default {
  name: 'StationMapViewer',
  props: {
    latitude: {
      type: Number,
      default: null,
    },
    longitude: {
      type: Number,
      default: null,
    },
    accuracyMeters: {
      type: Number,
      default: null,
    },
    editable: {
      type: Boolean,
      default: false,
    },
    height: {
      type: String,
      default: '180px',
    },
    source: {
      type: String,
      default: 'UNCONFIGURED',
    },
  },
  data() {
    return {
      map: null,
      marker: null,
      accuracyCircle: null,
      tileLayer: null,
      isTilesOnline: true,
      tileErrorCount: 0,
      tileSuccessCount: 0,
    };
  },
  computed: {
    hasCoordinates() {
      return (
        typeof this.latitude === 'number' &&
        !isNaN(this.latitude) &&
        typeof this.longitude === 'number' &&
        !isNaN(this.longitude)
      );
    },
    formattedLat() {
      return this.latitude !== null && this.latitude !== undefined ? Number(this.latitude).toFixed(6) : '-';
    },
    formattedLng() {
      return this.longitude !== null && this.longitude !== undefined ? Number(this.longitude).toFixed(6) : '-';
    },
  },
  watch: {
    latitude() {
      this.updateMarkerPosition();
    },
    longitude() {
      this.updateMarkerPosition();
    },
    accuracyMeters() {
      this.updateAccuracyCircle();
    },
    editable(val) {
      if (this.marker) {
        if (val) {
          this.marker.dragging.enable();
        } else {
          this.marker.dragging.disable();
        }
      }
    },
  },
  mounted() {
    this.$nextTick(() => {
      this.initMap();
    });
  },
  beforeDestroy() {
    if (this.map) {
      this.map.off();
      this.map.remove();
      this.map = null;
    }
  },
  methods: {
    initMap() {
      if (!this.$refs.mapContainer) return;
      if (this.map) {
        this.map.remove();
        this.map = null;
      }

      const initialLat = this.hasCoordinates ? this.latitude : -6.2088;
      const initialLng = this.hasCoordinates ? this.longitude : 106.8456;
      const initialZoom = this.hasCoordinates ? 14 : 5;

      this.map = L.map(this.$refs.mapContainer, {
        center: [initialLat, initialLng],
        zoom: initialZoom,
        zoomControl: false,
        attributionControl: false,
      });

      // Add OpenStreetMap TileLayer with strict tile error listeners
      this.tileLayer = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        attribution: '&copy; OpenStreetMap',
      });

      this.tileLayer.on('tileerror', () => {
        this.tileErrorCount++;
        if (this.tileErrorCount >= 2 && this.tileSuccessCount === 0) {
          this.isTilesOnline = false;
        }
      });

      this.tileLayer.on('tileload', () => {
        this.tileSuccessCount++;
        this.isTilesOnline = true;
      });

      this.tileLayer.addTo(this.map);

      // Create Custom SCADA DivIcon Marker
      this.createMarker(initialLat, initialLng);

      // Click to pick location when editable
      this.map.on('click', (e) => {
        if (this.editable) {
          const { lat, lng } = e.latlng;
          this.$emit('coordinate-change', {
            lat: Number(lat.toFixed(6)),
            lng: Number(lng.toFixed(6)),
          });
        }
      });

      // Invalidate size once DOM stabilizes
      setTimeout(() => {
        if (this.map) {
          this.map.invalidateSize();
          if (this.hasCoordinates) {
            this.map.setView([this.latitude, this.longitude], this.map.getZoom() || 14);
          }
        }
      }, 150);
    },

    createMarker(lat, lng) {
      if (!this.map) return;

      const markerHtml = `
        <div class="relative flex items-center justify-center" style="width: 32px; height: 32px; transform: translate(-16px, -32px);">
          <div class="absolute -bottom-1 w-2.5 h-1.5 bg-black/40 rounded-full blur-[1px]"></div>
          <div class="relative z-10 flex flex-col items-center">
            <svg viewBox="0 0 24 24" width="30" height="30" fill="#06b6d4" stroke="#083344" stroke-width="1.2" class="drop-shadow-md">
              <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
            </svg>
          </div>
        </div>
      `;

      const customIcon = L.divIcon({
        className: 'scada-station-marker-icon',
        html: markerHtml,
        iconSize: [32, 32],
        iconAnchor: [16, 32],
      });

      this.marker = L.marker([lat, lng], {
        icon: customIcon,
        draggable: this.editable,
      }).addTo(this.map);

      this.marker.on('dragend', () => {
        if (this.editable && this.marker) {
          const pos = this.marker.getLatLng();
          this.$emit('coordinate-change', {
            lat: Number(pos.lat.toFixed(6)),
            lng: Number(pos.lng.toFixed(6)),
          });
        }
      });

      this.updateAccuracyCircle();
    },

    updateMarkerPosition() {
      if (!this.map) return;
      if (this.hasCoordinates) {
        if (!this.marker) {
          this.createMarker(this.latitude, this.longitude);
        } else {
          this.marker.setLatLng([this.latitude, this.longitude]);
        }
        this.updateAccuracyCircle();
        this.map.panTo([this.latitude, this.longitude]);
      }
    },

    updateAccuracyCircle() {
      if (!this.map) return;
      if (this.accuracyCircle) {
        this.map.removeLayer(this.accuracyCircle);
        this.accuracyCircle = null;
      }
      if (this.hasCoordinates && this.accuracyMeters && this.accuracyMeters > 0) {
        this.accuracyCircle = L.circle([this.latitude, this.longitude], {
          radius: this.accuracyMeters,
          color: '#06b6d4',
          fillColor: '#06b6d4',
          fillOpacity: 0.15,
          weight: 1.2,
          dashArray: '3 3',
        }).addTo(this.map);
      }
    },

    recenterMap() {
      if (this.map && this.hasCoordinates) {
        this.map.setView([this.latitude, this.longitude], 15);
      }
    },

    zoomIn() {
      if (this.map) this.map.zoomIn();
    },

    zoomOut() {
      if (this.map) this.map.zoomOut();
    },

    invalidateSize() {
      if (this.map) {
        this.$nextTick(() => {
          this.map.invalidateSize();
        });
      }
    },
  },
};
</script>

<style scoped>
.station-map-viewer-wrapper :deep(.leaflet-pane) {
  z-index: 10;
}
.station-map-viewer-wrapper :deep(.leaflet-top),
.station-map-viewer-wrapper :deep(.leaflet-bottom) {
  z-index: 15;
}
</style>
