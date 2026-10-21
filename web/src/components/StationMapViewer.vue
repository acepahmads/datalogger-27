<template>
  <div
    class="station-map-viewer-wrapper relative w-full h-full overflow-hidden rounded-md border border-slate-200 dark:border-[#1E3048] bg-slate-100 dark:bg-[#070E1A] flex flex-col select-none"
    :class="{ 'is-editable': editable }"
    :style="{ minHeight: height || '180px' }"
  >
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

    <!-- Non-blocking Map Pinning Mode Hint Pill (Allows direct clicking through to map) -->
    <div
      v-if="!hasCoordinates"
      class="absolute top-2 left-2 right-12 z-30 pointer-events-none flex items-center space-x-1.5 px-2.5 py-1 rounded bg-slate-900/85 text-amber-300 border border-amber-500/40 text-[9.5px] font-mono shadow-md backdrop-blur-xs"
    >
      <span class="animate-pulse">📍</span>
      <span>{{ $t('operator.clickMapToPick') || 'Klik pada peta untuk menentukan koordinat stasiun' }}</span>
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
    <div class="absolute bottom-1.5 left-2 z-30 flex items-center space-x-1.5 px-1.5 py-0.5 rounded bg-black/60 backdrop-blur-xs border border-white/10 text-[8.5px] font-mono text-slate-300 pointer-events-none">
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

      const initialLat = this.hasCoordinates ? this.latitude : -6.9175;
      const initialLng = this.hasCoordinates ? this.longitude : 107.6191;
      const initialZoom = this.hasCoordinates ? 15 : 12;

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

      // Create Custom SCADA DivIcon Marker if station has coordinates
      if (this.hasCoordinates) {
        this.createMarker(this.latitude, this.longitude);
      }

      // Click on map to pinpoint exact station location when editable
      this.map.on('click', (e) => {
        if (this.editable) {
          const lat = Number(e.latlng.lat.toFixed(6));
          const lng = Number(e.latlng.lng.toFixed(6));
          this.setMarkerPosition(lat, lng);
          this.$emit('coordinate-change', { lat, lng });
        }
      });

      // Invalidate size once DOM stabilizes
      setTimeout(() => {
        if (this.map) {
          this.map.invalidateSize();
          if (this.hasCoordinates) {
            this.map.setView([this.latitude, this.longitude], this.map.getZoom() || 15);
          }
        }
      }, 150);
    },

    createMarker(lat, lng) {
      if (!this.map) return;

      const markerHtml = `
        <div class="relative flex items-center justify-center w-8 h-8 select-none">
          <div class="absolute bottom-0 w-3 h-1.5 bg-black/40 rounded-full blur-[1px]"></div>
          <div class="relative z-10 flex flex-col items-center">
            <svg viewBox="0 0 24 24" width="32" height="32" fill="#06b6d4" stroke="#083344" stroke-width="1.2" class="drop-shadow-md">
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
          const lat = Number(pos.lat.toFixed(6));
          const lng = Number(pos.lng.toFixed(6));
          this.updateAccuracyCircle();
          this.$emit('coordinate-change', { lat, lng });
        }
      });

      this.updateAccuracyCircle();
    },

    setMarkerPosition(lat, lng) {
      if (!this.map) return;
      if (!this.marker) {
        this.createMarker(lat, lng);
      } else {
        this.marker.setLatLng([lat, lng]);
      }
      this.updateAccuracyCircle();
    },

    updateMarkerPosition() {
      if (!this.map) return;
      if (this.hasCoordinates) {
        this.setMarkerPosition(this.latitude, this.longitude);
        if (this.map.getBounds() && !this.map.getBounds().contains([this.latitude, this.longitude])) {
          this.map.panTo([this.latitude, this.longitude]);
        }
      } else {
        if (this.marker) {
          this.map.removeLayer(this.marker);
          this.marker = null;
        }
        if (this.accuracyCircle) {
          this.map.removeLayer(this.accuracyCircle);
          this.accuracyCircle = null;
        }
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
.station-map-viewer-wrapper.is-editable :deep(.leaflet-grab) {
  cursor: crosshair !important;
}
.station-map-viewer-wrapper.is-editable :deep(.leaflet-marker-icon) {
  cursor: grab !important;
}
.station-map-viewer-wrapper.is-editable :deep(.leaflet-marker-icon:active) {
  cursor: grabbing !important;
}
</style>
