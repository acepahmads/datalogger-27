<template>
  <div
    class="station-map-viewer-wrapper relative w-full h-full overflow-hidden rounded-xl border border-slate-200/90 dark:border-slate-800 bg-slate-100 dark:bg-[#0B0F19] flex flex-col select-none font-sans"
    :class="{ 'is-editable': editable }"
    :style="{ minHeight: height || '180px' }"
  >
    <!-- Map Canvas Element -->
    <div
      ref="mapContainer"
      class="map-container flex-1 w-full h-full z-10"
      :class="{ 'opacity-95': isTilesOnline, 'opacity-40 pointer-events-none': !isTilesOnline && hasCoordinates }"
    ></div>

    <!-- Offline / Tile Load Failure Fallback Overlay -->
    <div
      v-if="!isTilesOnline && hasCoordinates"
      class="absolute inset-0 z-20 flex flex-col items-center justify-center p-3 text-center pointer-events-auto bg-slate-900/65 backdrop-blur-xs font-sans"
    >
      <!-- SCADA Crosshair Radar Grid Schematic -->
      <div class="relative w-24 h-24 mb-2 rounded-full border border-blue-400/40 flex items-center justify-center bg-blue-950/25 shadow-inner">
        <div class="absolute inset-0 rounded-full border border-blue-400/20 scale-75"></div>
        <div class="absolute inset-0 rounded-full border border-blue-400/10 scale-50"></div>
        <div class="absolute w-full h-px bg-blue-400/30"></div>
        <div class="absolute h-full w-px bg-blue-400/30"></div>
        <!-- Pulsing Station Marker Icon in Center -->
        <div class="relative z-10 flex flex-col items-center text-blue-400 animate-bounce">
          <svg class="w-6 h-6 drop-shadow-[0_0_8px_rgba(22,133,229,0.8)]" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
          </svg>
        </div>
      </div>

      <div class="text-xs font-bold text-white badge-text flex items-center space-x-1.5 mb-0.5">
        <span class="w-2 h-2 rounded-full bg-amber-400 animate-pulse"></span>
        <span>Local Schematic Fallback</span>
      </div>
      <div class="text-3xs text-slate-300 max-w-xs leading-snug mb-2">
        Online map tiles unreachable (offline edge mode). Station coordinates remain verified &amp; active.
      </div>
      <div class="px-2.5 py-1 rounded-lg bg-slate-950/70 border border-blue-400/30 text-2xs text-blue-300 font-semibold tabular-nums">
        Lat: {{ formattedLat }}, Lng: {{ formattedLng }}
      </div>
    </div>

    <!-- Floating Location Search Bar (Top Left, Press Enter to Execute) -->
    <div
      v-if="editable"
      class="absolute top-2.5 left-2.5 z-30 flex flex-col w-72 sm:w-80 md:w-96 max-w-[calc(100%-56px)] font-sans"
    >
      <div class="relative flex items-center shadow-md rounded-xl bg-white/95 dark:bg-[#0B0F19]/95 border border-slate-200/90 dark:border-slate-700/80 focus-within:border-[#1685E5] focus-within:ring-2 focus-within:ring-[#1685E5]/20 transition">
        <button
          type="button"
          @click.stop="executeSearch(true)"
          class="pl-2.5 pr-1 py-1.5 text-slate-400 hover:text-[#1685E5] dark:text-slate-500 dark:hover:text-blue-400 text-xs flex items-center justify-center cursor-pointer"
          title="Cari Lokasi"
        >
          <svg v-if="!isSearching" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <span v-else class="animate-spin inline-block w-3.5 h-3.5 border-2 border-[#1685E5] border-t-transparent rounded-full"></span>
        </button>
        <input
          ref="locationSearchInput"
          v-model="searchQuery"
          type="text"
          :placeholder="$t('operator.searchLocationPlaceholder') || 'Search city, address, or lat, lng... (Enter)'"
          class="w-full py-1.5 pl-1 pr-7 text-xs font-sans bg-transparent text-slate-800 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none"
          @input="onSearchInput"
          @keydown.enter.prevent="executeSearch(true)"
          @keydown.esc="clearSearch"
        />
        <button
          v-if="searchQuery"
          type="button"
          @click.stop="clearSearch"
          class="absolute right-2 text-slate-400 hover:text-slate-700 dark:hover:text-white text-xs cursor-pointer"
          title="Clear search"
        >
          ✕
        </button>
      </div>

      <!-- Dropdown Results List -->
      <div
        v-if="searchResults && searchResults.length > 0"
        @click.stop
        class="mt-1.5 bg-white dark:bg-[#0B0F19] border border-slate-200/90 dark:border-slate-700/80 rounded-xl shadow-xl overflow-hidden max-h-52 overflow-y-auto text-xs font-sans divide-y divide-slate-100 dark:divide-slate-800"
      >
        <div
          v-for="(res, idx) in searchResults"
          :key="idx"
          @mousedown.prevent
          @click.stop="selectSearchResult(res)"
          class="p-2.5 hover:bg-[#EAF5FF]/70 dark:hover:bg-blue-600/15 cursor-pointer transition flex items-start space-x-2"
        >
          <svg class="w-3.5 h-3.5 text-[#1685E5] dark:text-blue-400 mt-0.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
            <circle cx="12" cy="10" r="3" />
          </svg>
          <div class="flex-1 min-w-0">
            <div class="font-semibold text-slate-900 dark:text-slate-100 truncate">{{ res.name || res.display_name }}</div>
            <div class="text-3xs text-slate-500 dark:text-slate-400 truncate">{{ res.display_name }}</div>
            <div class="text-3xs text-[#1685E5] dark:text-blue-400 font-semibold tabular-nums mt-0.5">
              Lat: {{ Number(res.lat).toFixed(6) }}, Lng: {{ Number(res.lon).toFixed(6) }}
            </div>
          </div>
        </div>
      </div>

      <!-- Search Notice (e.g. Not Found) -->
      <div
        v-if="searchNotice"
        class="mt-1.5 p-2 rounded-xl bg-rose-50 dark:bg-rose-500/15 border border-rose-200 dark:border-rose-500/30 text-rose-700 dark:text-rose-400 text-3xs font-medium flex items-center justify-between"
      >
        <span>{{ searchNotice }}</span>
        <button @click="searchNotice = ''" class="ml-1 opacity-70 hover:opacity-100 font-bold">✕</button>
      </div>
    </div>

    <!-- Non-blocking Map Pinning Mode Hint Pill -->
    <div
      v-if="!hasCoordinates && !searchResults.length && !searchNotice"
      class="absolute top-12 left-2.5 z-20 pointer-events-none max-w-max flex items-center space-x-1.5 px-2.5 py-1 rounded-lg bg-slate-900/85 text-amber-300 border border-amber-500/40 text-3xs font-sans font-medium shadow-md backdrop-blur-xs"
    >
      <span class="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse"></span>
      <span>{{ $t('operator.clickMapToPick') || 'Klik pada peta untuk menentukan koordinat stasiun' }}</span>
    </div>

    <!-- Quick Floating Map Controls -->
    <div v-if="hasCoordinates" class="absolute top-2.5 right-2.5 z-30 flex flex-col space-y-1">
      <button
        type="button"
        @click="recenterMap"
        class="w-7 h-7 flex items-center justify-center rounded-lg bg-white/95 dark:bg-[#0B0F19]/95 hover:bg-[#EAF5FF] hover:text-[#1685E5] dark:hover:bg-blue-600/20 dark:hover:text-blue-400 text-slate-700 dark:text-slate-200 border border-slate-200/90 dark:border-slate-700/80 shadow-sm transition-colors"
        title="Recenter Map to Station"
      >
        <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="12" cy="12" r="7" />
          <circle cx="12" cy="12" r="2" />
          <line x1="12" y1="2" x2="12" y2="5" />
          <line x1="12" y1="19" x2="12" y2="22" />
          <line x1="2" y1="12" x2="5" y2="12" />
          <line x1="19" y1="12" x2="22" y2="12" />
        </svg>
      </button>
      <button
        type="button"
        @click="zoomIn"
        class="w-7 h-7 flex items-center justify-center rounded-lg bg-white/95 dark:bg-[#0B0F19]/95 hover:bg-[#EAF5FF] hover:text-[#1685E5] dark:hover:bg-blue-600/20 dark:hover:text-blue-400 text-slate-700 dark:text-slate-200 border border-slate-200/90 dark:border-slate-700/80 shadow-sm text-sm font-semibold transition-colors leading-none"
        title="Zoom In"
      >
        +
      </button>
      <button
        type="button"
        @click="zoomOut"
        class="w-7 h-7 flex items-center justify-center rounded-lg bg-white/95 dark:bg-[#0B0F19]/95 hover:bg-[#EAF5FF] hover:text-[#1685E5] dark:hover:bg-blue-600/20 dark:hover:text-blue-400 text-slate-700 dark:text-slate-200 border border-slate-200/90 dark:border-slate-700/80 shadow-sm text-sm font-semibold transition-colors leading-none"
        title="Zoom Out"
      >
        −
      </button>
    </div>

    <!-- Tile Status Pill (Bottom Left) -->
    <div class="absolute bottom-2 left-2.5 z-30 flex items-center space-x-1.5 px-2 py-0.5 rounded-lg bg-slate-900/75 backdrop-blur-xs border border-white/15 text-3xs font-sans font-medium text-white badge-text pointer-events-none shadow-sm">
      <span
        class="w-1.5 h-1.5 rounded-full"
        :class="isTilesOnline ? 'bg-emerald-400' : 'bg-amber-400'"
      ></span>
      <span>{{ isTilesOnline ? 'OSM Online Tiles' : 'Offline Mode' }}</span>
      <span v-if="accuracyMeters" class="text-sky-300 ml-1 tabular-nums">±{{ accuracyMeters }}m</span>
    </div>
  </div>
</template>

<script>
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import axios from 'axios';

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
      searchQuery: '',
      searchResults: [],
      isSearching: false,
      searchNotice: '',
      searchDebounceTimer: null,
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
    if (this.searchDebounceTimer) {
      clearTimeout(this.searchDebounceTimer);
      this.searchDebounceTimer = null;
    }
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
        if (this.searchResults && this.searchResults.length > 0) {
          this.searchResults = [];
        }
        if (this.editable) {
          const lat = Number(e.latlng.lat.toFixed(6));
          const lng = Number(e.latlng.lng.toFixed(6));
          this.setMarkerPosition(lat, lng);
          this.$emit('coordinate-change', { lat, lng, source: 'MAP' });
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
          <div class="absolute bottom-0 w-3 h-1.5 bg-black/35 rounded-full blur-[1px]"></div>
          <div class="relative z-10 flex flex-col items-center">
            <svg viewBox="0 0 24 24" width="32" height="32" fill="#1685E5" stroke="#ffffff" stroke-width="1.5" class="drop-shadow-md">
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
          this.$emit('coordinate-change', { lat, lng, source: 'MAP' });
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

    parseCoordinates(input) {
      if (!input) return null;
      const cleaned = input.trim().replace(/[;]/g, ',');
      const parts = cleaned.includes(',') ? cleaned.split(',') : cleaned.split(/\s+/);
      if (parts.length === 2) {
        const lat = parseFloat(parts[0].trim());
        const lng = parseFloat(parts[1].trim());
        if (!isNaN(lat) && !isNaN(lng) && lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180) {
          return { lat, lng };
        }
      }
      return null;
    },

    clearSearch() {
      if (this.searchDebounceTimer) {
        clearTimeout(this.searchDebounceTimer);
        this.searchDebounceTimer = null;
      }
      this.searchQuery = '';
      this.searchResults = [];
      this.searchNotice = '';
    },

    onSearchInput() {
      if (this.searchDebounceTimer) {
        clearTimeout(this.searchDebounceTimer);
        this.searchDebounceTimer = null;
      }
      this.searchNotice = '';
      const q = (this.searchQuery || '').trim();
      if (q.length < 3) {
        this.searchResults = [];
        return;
      }
      this.searchDebounceTimer = setTimeout(() => {
        this.executeSearch(false);
      }, 400);
    },

    async executeSearch(autoApplyTop = true) {
      if (this.searchDebounceTimer) {
        clearTimeout(this.searchDebounceTimer);
        this.searchDebounceTimer = null;
      }
      const q = (this.searchQuery || '').trim();
      if (!q) return;

      this.searchNotice = '';
      if (autoApplyTop) {
        this.searchResults = [];
      }

      // 1. Direct Coordinate Parsing
      const directCoords = this.parseCoordinates(q);
      if (directCoords) {
        this.applySearchedLocation({
          lat: directCoords.lat,
          lon: directCoords.lng,
          name: `Coordinates (${directCoords.lat.toFixed(6)}, ${directCoords.lng.toFixed(6)})`,
          display_name: `Lat: ${directCoords.lat.toFixed(6)}, Lng: ${directCoords.lng.toFixed(6)}`,
        });
        return;
      }

      // 2. Geocoding Search via Backend Proxy with Fallback to Nominatim
      this.isSearching = true;
      try {
        let results = [];
        try {
          const res = await axios.get('/api/system/location/search', { params: { q } });
          if (res.data && Array.isArray(res.data.data)) {
            results = res.data.data;
          }
        } catch (backendErr) {
          console.warn('Backend location search failed, attempting direct geocoder:', backendErr);
        }

        if (!results || results.length === 0) {
          try {
            const directRes = await axios.get(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(q)}&limit=5&addressdetails=1`);
            if (Array.isArray(directRes.data)) {
              results = directRes.data.map((item) => ({
                name: item.name || (item.display_name ? item.display_name.split(',')[0] : q),
                display_name: item.display_name,
                lat: parseFloat(item.lat),
                lon: parseFloat(item.lon),
                type: item.type,
              }));
            }
          } catch (directErr) {
            console.warn('Direct geocoder failed:', directErr);
          }
        }

        if (results && results.length > 0) {
          this.searchResults = results;
          if (autoApplyTop) {
            this.applySearchedLocation(results[0], false);
          }
        } else if (autoApplyTop) {
          this.searchResults = [];
          this.searchNotice = 'Lokasi tidak ditemukan. Masukkan nama kota/alamat spesifik atau koordinat (lat, lng).';
        }
      } catch (err) {
        if (autoApplyTop) {
          this.searchNotice = 'Gagal melakukan pencarian lokasi.';
        }
      } finally {
        this.isSearching = false;
      }
    },

    selectSearchResult(item) {
      if (!item) return;
      this.searchQuery = item.name || item.display_name || this.searchQuery;
      this.applySearchedLocation(item, true);
      this.searchResults = [];
    },

    applySearchedLocation(item, closeDropdown = true) {
      const lat = Number(Number(item.lat).toFixed(6));
      const lng = Number(Number(item.lon).toFixed(6));
      this.setMarkerPosition(lat, lng);
      if (this.map) {
        this.map.setView([lat, lng], 15);
      }
      if (closeDropdown) {
        this.searchResults = [];
      }
      this.$emit('coordinate-change', { lat, lng, source: 'SEARCH' });
      this.$emit('location-selected', {
        lat,
        lng,
        name: item.name,
        displayName: item.display_name,
      });
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
