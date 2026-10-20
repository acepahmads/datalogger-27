<template>
  <span
    class="inline-flex items-center justify-center flex-shrink-0 transition-colors leading-none"
    :class="computedClasses"
    :style="computedStyle"
    :aria-hidden="true"
  >
    <svg
      :width="iconSize"
      :height="iconSize"
      viewBox="0 0 24 24"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      class="w-full h-full block"
    >
      <!-- ============================================== -->
      <!-- 1. HEADER & NAVIGATION ICONS                   -->
      <!-- ============================================== -->

      <!-- Station / Telemetry Mast -->
      <g v-if="name === 'station'">
        <!-- Mast frame -->
        <path d="M12 2v20M5 22h14M8 17l4-4 4 4M9 11l3-3 3 3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
        <!-- Signal beacon waves (duotone accent) -->
        <path d="M7 6a7 7 0 0 1 10 0M4 3a11 11 0 0 1 16 0" stroke="currentColor" stroke-opacity="0.35" stroke-width="1.8" stroke-linecap="round" />
        <circle cx="12" cy="7" r="1.8" fill="currentColor" />
      </g>

      <!-- Operator User -->
      <g v-else-if="name === 'operator'">
        <circle cx="12" cy="7" r="3.5" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.2" />
        <path d="M5 21a7 7 0 0 1 14 0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <!-- Headset / comm gear -->
        <path d="M18 10a6 6 0 0 0-12 0v2a1.5 1.5 0 0 0 1.5 1.5H8M18 12a1.5 1.5 0 0 1-1.5 1.5H16" stroke="currentColor" stroke-opacity="0.45" stroke-width="1.5" stroke-linecap="round" />
      </g>

      <!-- Notification Bell -->
      <g v-else-if="name === 'bell' || name === 'notifications'">
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" fill="currentColor" fill-opacity="0.18" />
        <path d="M13.73 21a2 2 0 0 1-3.46 0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
        <circle cx="18" cy="5" r="2" fill="currentColor" />
      </g>

      <!-- System Activity / Telemetry Pulse -->
      <g v-else-if="name === 'activity' || name === 'pulse'">
        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" />
        <circle cx="15" cy="21" r="1.5" fill="currentColor" fill-opacity="0.3" />
        <circle cx="9" cy="3" r="1.5" fill="currentColor" fill-opacity="0.3" />
      </g>

      <!-- Dashboard Menu -->
      <g v-else-if="name === 'menu'">
        <line x1="4" y1="6" x2="20" y2="6" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <line x1="4" y1="12" x2="20" y2="12" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <line x1="4" y1="18" x2="20" y2="18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <circle cx="4" cy="6" r="1.2" fill="currentColor" fill-opacity="0.4" />
        <circle cx="4" cy="12" r="1.2" fill="currentColor" fill-opacity="0.4" />
        <circle cx="4" cy="18" r="1.2" fill="currentColor" fill-opacity="0.4" />
      </g>

      <!-- Layout / Panels Grid -->
      <g v-else-if="name === 'layout' || name === 'grid'">
        <rect x="3" y="3" width="8" height="8" rx="2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.25" />
        <rect x="13" y="3" width="8" height="8" rx="2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.1" />
        <rect x="3" y="13" width="8" height="8" rx="2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.1" />
        <rect x="13" y="13" width="8" height="8" rx="2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.25" />
      </g>

      <!-- Fullscreen / Expand -->
      <g v-else-if="name === 'fullscreen'">
        <path d="M15 3h6v6M9 21H3v-6M21 9l-7 7M3 15l7-7M21 3l-7 7M3 21l7-7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
        <rect x="7" y="7" width="10" height="10" rx="1.5" stroke="currentColor" stroke-opacity="0.2" stroke-width="1.4" stroke-dasharray="2 2" fill="currentColor" fill-opacity="0.1" />
      </g>

      <!-- Filter / Funnel -->
      <g v-else-if="name === 'filter'">
        <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" fill="currentColor" fill-opacity="0.2" />
        <circle cx="12" cy="7" r="1.5" fill="currentColor" />
      </g>

      <!-- Search / Lens -->
      <g v-else-if="name === 'search'">
        <circle cx="11" cy="11" r="7" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.15" />
        <path d="m21 21-4.35-4.35" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
      </g>

      <!-- Pin / Bookmark -->
      <g v-else-if="name === 'pin'">
        <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" fill="currentColor" fill-opacity="0.25" />
      </g>

      <!-- Settings / Sliders -->
      <g v-else-if="name === 'settings' || name === 'sliders' || name === 'chart-setup'">
        <line x1="4" y1="7" x2="20" y2="7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <line x1="4" y1="17" x2="20" y2="17" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <circle cx="9" cy="7" r="2.8" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.25" />
        <circle cx="15" cy="17" r="2.8" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.25" />
      </g>

      <!-- ============================================== -->
      <!-- 2. PARAMETER SENSOR ICONS (DUOTONE SCADA)     -->
      <!-- ============================================== -->

      <!-- PM2.5 / Fine Particulates -->
      <g v-else-if="name === 'pm25' || name === 'pm2.5' || name === 'particles'">
        <circle cx="6" cy="8" r="2.2" fill="currentColor" />
        <circle cx="12" cy="6" r="1.6" fill="currentColor" fill-opacity="0.5" />
        <circle cx="18" cy="8" r="2.4" fill="currentColor" />
        <circle cx="7" cy="16" r="2" fill="currentColor" />
        <circle cx="13" cy="13" r="2.2" fill="currentColor" fill-opacity="0.6" />
        <circle cx="19" cy="15" r="1.5" fill="currentColor" />
        <circle cx="12" cy="19" r="1.8" fill="currentColor" fill-opacity="0.5" />
        <!-- Subtle aerosol stream curves -->
        <path d="M4 11c2.5-1 4.5 0 7-1s4.5-1 7 0" stroke="currentColor" stroke-opacity="0.3" stroke-width="1.2" stroke-linecap="round" />
      </g>

      <!-- PM10 / Coarse Particulates / Dust -->
      <g v-else-if="name === 'pm10'">
        <circle cx="7" cy="7" r="3.2" stroke="currentColor" stroke-width="1.6" fill="currentColor" fill-opacity="0.3" />
        <circle cx="16" cy="8" r="2.6" stroke="currentColor" stroke-width="1.6" fill="currentColor" fill-opacity="0.2" />
        <circle cx="10" cy="16" r="3" stroke="currentColor" stroke-width="1.6" fill="currentColor" fill-opacity="0.35" />
        <circle cx="18" cy="17" r="2" fill="currentColor" />
        <circle cx="12" cy="11" r="1.5" fill="currentColor" />
      </g>

      <!-- O3 (Ozone) / Triple Atom Ring -->
      <g v-else-if="name === 'o3' || name === 'ozone'">
        <circle cx="12" cy="6" r="2.4" fill="currentColor" />
        <circle cx="6" cy="16" r="2.4" fill="currentColor" />
        <circle cx="18" cy="16" r="2.4" fill="currentColor" />
        <path d="M12 6L6 16h12Z" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" fill="currentColor" fill-opacity="0.15" />
        <circle cx="12" cy="13" r="1.2" fill="currentColor" fill-opacity="0.4" />
      </g>

      <!-- SO2 / Sulfur Dioxide -->
      <g v-else-if="name === 'so2'">
        <!-- Central Sulfur -->
        <circle cx="12" cy="9" r="2.8" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.3" />
        <!-- Oxygen bounds -->
        <circle cx="6" cy="16" r="2.2" fill="currentColor" />
        <circle cx="18" cy="16" r="2.2" fill="currentColor" />
        <path d="M10 11l-3 3M14 11l3 3" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
      </g>

      <!-- NO2 / Nitrogen Dioxide -->
      <g v-else-if="name === 'no2'">
        <circle cx="12" cy="8" r="2.8" fill="currentColor" />
        <circle cx="6" cy="16" r="2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.25" />
        <circle cx="18" cy="16" r="2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.25" />
        <path d="M10.5 10L7.5 14M13.5 10l3 4" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </g>

      <!-- CO / Carbon Monoxide -->
      <g v-else-if="name === 'co'">
        <!-- Diatomic bond C=O -->
        <circle cx="7" cy="12" r="3.2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.2" />
        <circle cx="17" cy="12" r="3.2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.35" />
        <line x1="10.2" y1="10.5" x2="13.8" y2="10.5" stroke="currentColor" stroke-width="1.8" />
        <line x1="10.2" y1="13.5" x2="13.8" y2="13.5" stroke="currentColor" stroke-width="1.8" />
      </g>

      <!-- HC / Hydrocarbons / Gas Molecule -->
      <g v-else-if="name === 'hc' || name === 'atom' || name === 'gas'">
        <circle cx="12" cy="12" r="2.6" fill="currentColor" />
        <ellipse cx="12" cy="12" rx="9" ry="3.8" transform="rotate(45 12 12)" stroke="currentColor" stroke-width="1.6" fill="currentColor" fill-opacity="0.1" />
        <ellipse cx="12" cy="12" rx="9" ry="3.8" transform="rotate(-45 12 12)" stroke="currentColor" stroke-width="1.6" fill="currentColor" fill-opacity="0.1" />
        <circle cx="6" cy="6" r="1.4" fill="currentColor" fill-opacity="0.5" />
        <circle cx="18" cy="18" r="1.4" fill="currentColor" fill-opacity="0.5" />
      </g>

      <!-- Temperature / Thermometer -->
      <g v-else-if="name === 'temperature' || name === 'thermometer' || name === 'temp'">
        <path d="M14 14.76V4a2 2 0 0 0-4 0v10.76a4 4 0 1 0 4 0z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" fill="currentColor" fill-opacity="0.15" />
        <!-- Mercury bulb and column -->
        <circle cx="12" cy="17" r="2.2" fill="currentColor" />
        <line x1="12" y1="9" x2="12" y2="15" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <!-- Tick marks -->
        <line x1="16" y1="6" x2="18" y2="6" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-opacity="0.6" />
        <line x1="16" y1="9" x2="18" y2="9" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-opacity="0.6" />
      </g>

      <!-- Humidity / Droplet -->
      <g v-else-if="name === 'humidity' || name === 'droplet' || name === 'rh'">
        <path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" fill="currentColor" fill-opacity="0.25" />
        <!-- Inner moisture ripple -->
        <path d="M12 17a4 4 0 0 0 4-4" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.5" stroke-linecap="round" />
      </g>

      <!-- Air Pressure / Barometer Gauge -->
      <g v-else-if="name === 'pressure' || name === 'gauge' || name === 'baro'">
        <path d="M3.34 17a10 10 0 1 1 17.32 0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <line x1="12" y1="14" x2="15" y2="10" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <circle cx="12" cy="14" r="2" fill="currentColor" />
        <path d="M7 17h10" stroke="currentColor" stroke-opacity="0.3" stroke-width="1.5" stroke-dasharray="2 2" />
      </g>

      <!-- Wind Speed / Anemometer / Stream -->
      <g v-else-if="name === 'wind-speed' || name === 'wind'">
        <path d="M17.7 7.7A2.5 2.5 0 1 1 19.5 12H2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <path d="M12.7 3.7A2 2 0 1 1 14 7H2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <path d="M14.7 16.3A2 2 0 1 0 16 20H2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <circle cx="19.5" cy="12" r="1.5" fill="currentColor" fill-opacity="0.4" />
      </g>

      <!-- Wind Angle / Compass Rose -->
      <g v-else-if="name === 'wind-angle' || name === 'compass'">
        <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.1" />
        <polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" fill="currentColor" fill-opacity="0.4" />
        <circle cx="12" cy="12" r="1.5" fill="currentColor" />
      </g>

      <!-- Solar Radiation / Sun -->
      <g v-else-if="name === 'solar' || name === 'sun'">
        <circle cx="12" cy="12" r="4.5" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.3" />
        <path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </g>

      <!-- Rain Fall / Cloud & Rain -->
      <g v-else-if="name === 'rain' || name === 'cloud-rain'">
        <path d="M17.5 17H9a6 6 0 1 1 5.71-8h1.79a4 4 0 1 1 1 8Z" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" fill="currentColor" fill-opacity="0.2" />
        <path d="M8 19l-1 3M12 19l-1 3M16 19l-1 3" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </g>

      <!-- Waves / Water Flow -->
      <g v-else-if="name === 'waves'">
        <path d="M2 6c.6.5 1.2 1 2.5 1C7 7 7 5 9.5 5c2.6 0 2.4 2 5 2 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1M2 12c.6.5 1.2 1 2.5 1 2.5 0 2.5-2 5-2 2.6 0 2.4 2 5 2 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1M2 18c.6.5 1.2 1 2.5 1 2.5 0 2.5-2 5-2 2.6 0 2.4 2 5 2 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </g>

      <!-- Generic Sensor / Chip -->
      <g v-else-if="name === 'sensor'">
        <rect x="5" y="5" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.18" />
        <circle cx="12" cy="12" r="2.5" fill="currentColor" />
        <path d="M9 2v3M15 2v3M9 19v3M15 19v3M2 9h3M2 15h3M19 9h3M19 15h3" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" />
      </g>

      <!-- ============================================== -->
      <!-- 3. TREND ANALYSIS, ACTIONS & STATUS           -->
      <!-- ============================================== -->

      <!-- Chart / Multi-series Trend -->
      <g v-else-if="name === 'chart' || name === 'series'">
        <path d="M3 3v18h18" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
        <path d="M19 9l-5 5-4-4-5 5" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" />
        <path d="M19 9l-5 5-4-4-5 5v5h14z" fill="currentColor" fill-opacity="0.15" />
      </g>

      <!-- Reset / Counter-Clockwise Rotate -->
      <g v-else-if="name === 'reset'">
        <path d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
        <circle cx="12" cy="12" r="2" fill="currentColor" fill-opacity="0.25" />
      </g>

      <!-- Line Smooth (Bezier Wave Curve / Lengkung) -->
      <g v-else-if="name === 'line-smooth' || name === 'curve' || name === 'smooth'">
        <path d="M3 16C6 16 7 8 12 8S18 16 21 16" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" fill="none" />
        <circle cx="3" cy="16" r="2" fill="currentColor" />
        <circle cx="12" cy="8" r="2" fill="currentColor" />
        <circle cx="21" cy="16" r="2" fill="currentColor" />
      </g>

      <!-- Line Sharp / Linear (Runcing / Angular Segments) -->
      <g v-else-if="name === 'line-sharp' || name === 'linear' || name === 'runcing'">
        <path d="M3 16L12 8L21 16" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" fill="none" />
        <circle cx="3" cy="16" r="2" fill="currentColor" />
        <circle cx="12" cy="8" r="2" fill="currentColor" />
        <circle cx="21" cy="16" r="2" fill="currentColor" />
      </g>

      <!-- Clear / X-Circle -->
      <g v-else-if="name === 'clear'">
        <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.18" />
        <path d="M15 9l-6 6M9 9l6 6" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </g>

      <!-- Check All / Select All Eligible -->
      <g v-else-if="name === 'check-all'">
        <path d="M8 12l3 3 7-7M3 12l3 3 2-2" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" />
        <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.5" stroke-opacity="0.25" fill="currentColor" fill-opacity="0.1" />
      </g>

      <!-- Export Data -->
      <g v-else-if="name === 'export'">
        <path d="M4 16v1a3 3 0 0 0 3 3h10a3 3 0 0 0 3-3v-1M12 4v12M8 8l4-4 4 4" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" />
        <rect x="7" y="15" width="10" height="3" rx="1" fill="currentColor" fill-opacity="0.25" />
      </g>

      <!-- Download Tray -->
      <g v-else-if="name === 'download'">
        <path d="M4 16v1a3 3 0 0 0 3 3h10a3 3 0 0 0 3-3v-1M12 4v12M8 12l4 4 4-4" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round" />
        <rect x="7" y="16" width="10" height="3" rx="1" fill="currentColor" fill-opacity="0.25" />
      </g>

      <!-- Operational Report -->
      <g v-else-if="name === 'report'">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" fill="currentColor" fill-opacity="0.18" />
        <polyline points="14 2 14 8 20 8" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" />
        <!-- Report bar charts inside -->
        <line x1="8" y1="17" x2="8" y2="14" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <line x1="12" y1="17" x2="12" y2="12" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <line x1="16" y1="17" x2="16" y2="15" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
      </g>

      <!-- Alarm History / Clipboard Log -->
      <g v-else-if="name === 'history'">
        <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" fill="currentColor" fill-opacity="0.12" />
        <rect x="8" y="2" width="8" height="4" rx="1" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.3" />
        <line x1="8" y1="11" x2="16" y2="11" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
        <line x1="8" y1="15" x2="13" y2="15" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
      </g>

      <!-- ISPU / Air Quality Leaf Gauge -->
      <g v-else-if="name === 'ispu'">
        <path d="M12 2L4 7v6c0 5.25 3.4 10.15 8 11.45 4.6-1.3 8-6.2 8-11.45V7l-8-5z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" fill="currentColor" fill-opacity="0.2" />
        <!-- Environmental leaf stem -->
        <path d="M12 8c2 2 2 6-2 8M10 13c1-1 3-1 4 0" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" />
      </g>

      <!-- Sound On -->
      <g v-else-if="name === 'sound-on'">
        <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" fill="currentColor" fill-opacity="0.25" />
        <path d="M15.54 8.46a5 5 0 0 1 0 7.07M19.07 4.93a10 10 0 0 1 0 14.14" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </g>

      <!-- Sound Muted -->
      <g v-else-if="name === 'sound-muted'">
        <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" fill="currentColor" fill-opacity="0.2" />
        <line x1="23" y1="9" x2="17" y2="15" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
        <line x1="17" y1="9" x2="23" y2="15" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" />
      </g>

      <!-- Sound Off -->
      <g v-else-if="name === 'sound-off'">
        <polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" fill="currentColor" fill-opacity="0.15" />
        <line x1="2" y1="2" x2="22" y2="22" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" />
      </g>

      <!-- Eye / View Config -->
      <g v-else-if="name === 'eye'">
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" fill="currentColor" fill-opacity="0.15" />
        <circle cx="12" cy="12" r="3" stroke="currentColor" stroke-width="1.8" fill="currentColor" fill-opacity="0.3" />
      </g>

      <!-- Edit / Pencil -->
      <g v-else-if="name === 'edit'">
        <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
        <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round" fill="currentColor" fill-opacity="0.25" />
      </g>

      <!-- Lightning / Outlier Spike -->
      <g v-else-if="name === 'lightning' || name === 'spike'">
        <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" fill="currentColor" fill-opacity="0.35" />
      </g>

      <!-- Close / Cross -->
      <g v-else-if="name === 'close'">
        <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
      </g>

      <!-- Default Fallback: Activity / Pulse -->
      <g v-else>
        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
        <circle cx="12" cy="12" r="2" fill="currentColor" fill-opacity="0.25" />
      </g>
    </svg>
  </span>
</template>

<script>
export default {
  name: 'OperatorDashboardIcon',
  props: {
    name: {
      type: String,
      required: true,
      default: 'activity',
    },
    size: {
      type: [Number, String],
      default: 16,
    },
    sizeClass: {
      type: String,
      default: '',
    },
    color: {
      type: String,
      default: '',
    },
  },
  computed: {
    iconSize() {
      const s = Number(this.size);
      return !isNaN(s) && s > 0 ? s : 16;
    },
    computedClasses() {
      if (this.sizeClass) {
        return this.sizeClass;
      }
      return '';
    },
    computedStyle() {
      const style = {};
      if (this.color) {
        style.color = this.color;
      }
      if (!this.sizeClass && this.iconSize) {
        style.width = `${this.iconSize}px`;
        style.height = `${this.iconSize}px`;
      }
      return style;
    },
  },
};
</script>
