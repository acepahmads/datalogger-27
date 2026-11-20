<template>
  <div
    v-if="weather.hasAnyWeather"
    id="bento-weather-hero-card"
    class="bento-weather-root"
    :class="[
      isDark ? 'is-night' : 'is-day',
      'tone-' + condition.themeTone,
      compact ? 'is-compact-weather-bar' : 'bento-weather-card',
      { 'no-anim': !animationsEnabled }
    ]"
  >
    <!-- =================================================================== -->
    <!-- MODE A: COMPACT HORIZONTAL WEATHER STRIP (< 1280px TABLET / MOBILE) -->
    <!-- =================================================================== -->
    <template v-if="compact">
      <div class="cw-row">
        <!-- Left: Dynamic Weather Icon + Condition + Primary Temp & Humidity -->
        <div class="cw-primary">
          <div class="cw-icon-wrap" :title="condition.label">
            <svg width="36" height="36" viewBox="0 0 64 64" fill="none" class="w-icon-svg" aria-hidden="true">
              <defs>
                <radialGradient :id="uid + '-sunGlow'" cx="50%" cy="50%" r="50%">
                  <stop offset="0%" stop-color="#FDE047" stop-opacity="0.65" />
                  <stop offset="65%" stop-color="#F59E0B" stop-opacity="0.18" />
                  <stop offset="100%" stop-color="#F59E0B" stop-opacity="0" />
                </radialGradient>
                <linearGradient :id="uid + '-sunFill'" x1="18" y1="12" x2="48" y2="48" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" stop-color="#FEF08A" />
                  <stop offset="52%" stop-color="#FBBF24" />
                  <stop offset="100%" stop-color="#F97316" />
                </linearGradient>
                <linearGradient :id="uid + '-cloudFill'" x1="16" y1="24" x2="50" y2="54" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" :stop-color="isDark ? '#E2E8F0' : '#FFFFFF'" />
                  <stop offset="100%" :stop-color="isDark ? '#94A3B8' : '#CBD5E1'" />
                </linearGradient>
                <linearGradient :id="uid + '-moonFill'" x1="18" y1="12" x2="46" y2="48" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" stop-color="#FEF9C3" />
                  <stop offset="100%" stop-color="#F59E0B" />
                </linearGradient>
              </defs>

              <!-- SUNNY -->
              <g v-if="condition.key === 'sunny'">
                <circle cx="32" cy="32" r="28" :fill="'url(#' + uid + '-sunGlow)'" />
                <g stroke="#F59E0B" stroke-width="2.8" stroke-linecap="round" opacity="0.88" class="w-sun-rays">
                  <line x1="32" y1="7" x2="32" y2="12" />
                  <line x1="32" y1="52" x2="32" y2="57" />
                  <line x1="7" y1="32" x2="12" y2="32" />
                  <line x1="52" y1="32" x2="57" y2="32" />
                  <line x1="14.3" y1="14.3" x2="17.8" y2="17.8" />
                  <line x1="46.2" y1="46.2" x2="49.7" y2="49.7" />
                  <line x1="49.7" y1="14.3" x2="46.2" y2="17.8" />
                  <line x1="17.8" y1="46.2" x2="14.3" y2="49.7" />
                </g>
                <circle cx="32" cy="32" r="14.5" :fill="'url(#' + uid + '-sunFill)'" />
              </g>

              <!-- PARTLY CLOUDY -->
              <g v-else-if="condition.key === 'partly_cloudy'">
                <circle cx="24" cy="24" r="17" :fill="'url(#' + uid + '-sunGlow)'" />
                <circle cx="24" cy="24" r="10.5" :fill="'url(#' + uid + '-sunFill)'" />
                <path
                  d="M20 48H45C50.5 48 55 43.7 55 38.4C55 33.4 51.1 29.3 46.2 28.9C44.5 23.3 39.2 19.2 33 19.2C25.9 19.2 20.1 24.5 19.4 31.3C15.1 32 11.8 35.7 11.8 40.1C11.8 44.5 15.5 48 20 48Z"
                  :fill="'url(#' + uid + '-cloudFill)'"
                  :stroke="isDark ? 'rgba(255,255,255,0.22)' : 'rgba(148,163,184,0.35)'"
                  stroke-width="1"
                />
              </g>

              <!-- RAIN / NIGHT RAIN -->
              <g v-else-if="condition.key === 'rain' || condition.key === 'heavy_rain' || condition.key === 'night_rain'">
                <path
                  d="M18 41H46C51.2 41 55.4 36.9 55.4 31.8C55.4 27 51.6 23.1 46.9 22.7C45.3 17.3 40.2 13.4 34.2 13.4C27.4 13.4 21.8 18.4 21.1 25C17 25.7 13.8 29.2 13.8 33.4C13.8 37.6 15.6 41 18 41Z"
                  :fill="'url(#' + uid + '-cloudFill)'"
                  :stroke="isDark ? 'rgba(148,163,184,0.35)' : 'rgba(100,116,139,0.28)'"
                  stroke-width="1"
                />
                <g stroke="#0EA5E9" stroke-width="2.6" stroke-linecap="round" class="w-rain-drops">
                  <line x1="23" y1="45" x2="20" y2="54" />
                  <line x1="33" y1="46" x2="30" y2="56" />
                  <line x1="43" y1="45" x2="40" y2="54" />
                </g>
              </g>

              <!-- NIGHT CLEAR / NIGHT CLOUDY -->
              <g v-else-if="condition.key === 'night_clear' || condition.key === 'night_cloudy'">
                <path
                  d="M36.5 13C27.5 14 21 21.8 22 30.8C23 39.8 30.8 46.3 39.8 45.3C43.2 44.9 46.2 43.4 48.6 41.2C42.1 42.1 35.8 37.9 34.5 31.4C33.2 24.9 37 18.5 43.2 16.5C41.1 14.3 38.9 13.2 36.5 13Z"
                  :fill="'url(#' + uid + '-moonFill)'"
                />
                <circle cx="48" cy="18" r="1.8" fill="#FEF08A" />
                <circle cx="17" cy="20" r="1.3" fill="#BAE6FD" />
                <path
                  v-if="condition.key === 'night_cloudy'"
                  d="M18 52H46C50.4 52 54 48.5 54 44.2C54 40.2 50.8 36.9 46.8 36.6C45.5 32.3 41.4 29.2 36.6 29.2C31.1 29.2 26.6 33.2 26 38.5C22.6 39 20 41.9 20 45.4C20 49 16 52 18 52Z"
                  :fill="'url(#' + uid + '-cloudFill)'"
                />
              </g>

              <!-- CLOUDY DEFAULT -->
              <g v-else>
                <path
                  d="M18 46H46C51.2 46 55.4 41.9 55.4 36.8C55.4 32 51.6 28.1 46.9 27.7C45.3 22.3 40.2 18.4 34.2 18.4C27.4 18.4 21.8 23.4 21.1 30C17 30.7 13.8 34.2 13.8 38.4C13.8 42.6 15.6 46 18 46Z"
                  :fill="'url(#' + uid + '-cloudFill)'"
                  :stroke="isDark ? 'rgba(255,255,255,0.2)' : 'rgba(148,163,184,0.35)'"
                  stroke-width="1"
                />
              </g>
            </svg>
          </div>

          <!-- Compact Temp -->
          <button
            v-if="weather.temperature"
            type="button"
            class="cw-metric-btn"
            :class="{ 'is-selected': isSelected(weather.temperature) }"
            @click="onSelectParam(weather.temperature)"
            @contextmenu.prevent="onContextParam($event, weather.temperature)"
            :title="buildParamTooltip(weather.temperature)"
          >
            <span class="cw-kicker">TEMP</span>
            <span class="cw-val-row">
              <strong class="cw-temp-val">{{ formatVal(weather.temperature, 1) }}</strong>
              <span class="cw-unit">{{ resolveUnit(weather.temperature, '°C') }}</span>
            </span>
          </button>

          <div v-if="weather.temperature && weather.humidity" class="cw-vdiv"></div>

          <!-- Compact Humidity -->
          <button
            v-if="weather.humidity"
            type="button"
            class="cw-metric-btn"
            :class="{ 'is-selected': isSelected(weather.humidity) }"
            @click="onSelectParam(weather.humidity)"
            @contextmenu.prevent="onContextParam($event, weather.humidity)"
            :title="buildParamTooltip(weather.humidity)"
          >
            <span class="cw-kicker">HUMIDITY</span>
            <span class="cw-val-row">
              <strong class="cw-hum-val">{{ formatVal(weather.humidity, 1) }}</strong>
              <span class="cw-unit">{{ resolveUnit(weather.humidity, '%RH') }}</span>
            </span>
          </button>

          <span class="cw-cond-pill">{{ condition.label }}</span>
        </div>

        <!-- Middle/Right: Mini Wind Compass + Secondary Chips + Actions -->
        <div class="cw-extras">
          <button
            v-if="weather.windSpeed || weather.windAngle"
            type="button"
            class="cw-wind-chip"
            :class="{ 'is-selected': isSelected(weather.windSpeed) || isSelected(weather.windAngle) }"
            @click="onSelectParam(weather.windSpeed || weather.windAngle)"
            @contextmenu.prevent="onContextParam($event, weather.windSpeed || weather.windAngle)"
            :title="windTooltip"
          >
            <svg width="20" height="20" viewBox="0 0 36 36" fill="none" class="flex-shrink-0">
              <circle cx="18" cy="18" r="15" :stroke="isDark ? 'rgba(148,163,184,0.3)' : '#CBD5E1'" stroke-width="1.8" />
              <g :style="compactCompassNeedleStyle" class="w-compass-needle">
                <polygon points="18,6 21.5,18 18,15.5 14.5,18" fill="#0EA5E9" />
                <polygon points="18,30 14.5,18 18,20.5 21.5,18" :fill="isDark ? '#64748B' : '#94A3B8'" />
              </g>
              <circle cx="18" cy="18" r="2" fill="#0EA5E9" />
            </svg>
            <span v-if="weather.windSpeed" class="tabular-nums font-extrabold">{{ formatVal(weather.windSpeed, 2) }} <small>{{ resolveUnit(weather.windSpeed, 'm/s') }}</small></span>
            <span v-if="weather.windAngle && windCardinal.valid" class="cw-dir-tag">{{ windCardinal.deg }}° {{ windCardinal.short }}</span>
          </button>

          <button
            v-for="chip in secondaryChips"
            :key="'cw-' + chip.key"
            type="button"
            class="cw-mini-chip"
            :class="[chip.toneClass, { 'is-selected': isSelected(chip.param) }]"
            @click="onSelectParam(chip.param)"
            @contextmenu.prevent="onContextParam($event, chip.param)"
            :title="buildParamTooltip(chip.param)"
          >
            <span class="cw-chip-lbl">{{ chip.shortLabel }}</span>
            <strong class="tabular-nums">{{ chip.formatted }}</strong>
            <span class="cw-chip-unit">{{ chip.unit }}</span>
          </button>

          <button
            type="button"
            class="cw-trend-btn"
            @click="onSelectAllWeatherTrend"
            title="Tampilkan semua parameter cuaca pada grafik Analisis Tren"
          >
            <span>Tren Cuaca</span>
          </button>
        </div>
      </div>
    </template>

    <!-- =================================================================== -->
    <!-- MODE B: WIDESCREEN FULL RIGHT BENTO WEATHER HERO CARD (>= 1280px)   -->
    <!-- =================================================================== -->
    <template v-else>
      <!-- 1. Atmospheric Sky Header: Condition + Dynamic 3D-Style Weather SVG + Temperature Focus -->
      <div class="wh-sky-band" :class="'sky-' + condition.themeTone">
        <div class="wh-top-row">
          <span class="wh-kicker">WEATHER · METEO</span>
          <span class="wh-cond-badge" :class="'cond-' + condition.themeTone">
            <span class="wh-live-dot"></span>
            <span>{{ condition.label }}</span>
          </span>
        </div>

        <div class="wh-hero-stage">
          <!-- Multi-layered Gradient SVG Weather Illustration -->
          <div class="wh-art-wrap" :title="condition.label">
            <svg width="64" height="64" viewBox="0 0 64 64" fill="none" class="w-icon-svg" aria-hidden="true">
              <defs>
                <radialGradient :id="uid + '-fsunGlow'" cx="50%" cy="50%" r="50%">
                  <stop offset="0%" stop-color="#FDE047" stop-opacity="0.72" />
                  <stop offset="62%" stop-color="#F59E0B" stop-opacity="0.22" />
                  <stop offset="100%" stop-color="#F59E0B" stop-opacity="0" />
                </radialGradient>
                <linearGradient :id="uid + '-fsunFill'" x1="16" y1="10" x2="48" y2="50" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" stop-color="#FEF9C3" />
                  <stop offset="48%" stop-color="#FBBF24" />
                  <stop offset="100%" stop-color="#EA580C" />
                </linearGradient>
                <linearGradient :id="uid + '-fcloudFill'" x1="14" y1="20" x2="52" y2="54" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" :stop-color="isDark ? '#F1F5F9' : '#FFFFFF'" />
                  <stop offset="65%" :stop-color="isDark ? '#CBD5E1' : '#EFF6FF'" />
                  <stop offset="100%" :stop-color="isDark ? '#94A3B8' : '#CBD5E1'" />
                </linearGradient>
                <linearGradient :id="uid + '-fstormCloud'" x1="14" y1="14" x2="52" y2="44" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" :stop-color="isDark ? '#CBD5E1' : '#E2E8F0'" />
                  <stop offset="100%" :stop-color="isDark ? '#64748B' : '#94A3B8'" />
                </linearGradient>
                <linearGradient :id="uid + '-fmoonFill'" x1="18" y1="10" x2="48" y2="48" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" stop-color="#FEF9C3" />
                  <stop offset="60%" stop-color="#FDE047" />
                  <stop offset="100%" stop-color="#F59E0B" />
                </linearGradient>
                <filter :id="uid + '-softShadow'" x="-20%" y="-20%" width="140%" height="140%">
                  <feDropShadow dx="0" dy="3" stdDeviation="3" flood-color="#0F172A" :flood-opacity="isDark ? 0.4 : 0.12" />
                </filter>
              </defs>

              <!-- SUNNY -->
              <g v-if="condition.key === 'sunny'">
                <circle cx="32" cy="32" r="29" :fill="'url(#' + uid + '-fsunGlow)'" />
                <g stroke="#F59E0B" stroke-width="2.8" stroke-linecap="round" opacity="0.9" class="w-sun-rays">
                  <line x1="32" y1="6" x2="32" y2="11.5" />
                  <line x1="32" y1="52.5" x2="32" y2="58" />
                  <line x1="6" y1="32" x2="11.5" y2="32" />
                  <line x1="52.5" y1="32" x2="58" y2="32" />
                  <line x1="13.6" y1="13.6" x2="17.5" y2="17.5" />
                  <line x1="46.5" y1="46.5" x2="50.4" y2="50.4" />
                  <line x1="50.4" y1="13.6" x2="46.5" y2="17.5" />
                  <line x1="17.5" y1="46.5" x2="13.6" y2="50.4" />
                </g>
                <circle cx="32" cy="32" r="14.5" :fill="'url(#' + uid + '-fsunFill)'" :filter="'url(#' + uid + '-softShadow)'" />
                <circle cx="27" cy="27" r="4.5" fill="#FFFFFF" opacity="0.35" />
              </g>

              <!-- PARTLY CLOUDY -->
              <g v-else-if="condition.key === 'partly_cloudy'">
                <circle cx="24" cy="23" r="19" :fill="'url(#' + uid + '-fsunGlow)'" />
                <g stroke="#F59E0B" stroke-width="2.2" stroke-linecap="round" opacity="0.85" class="w-sun-rays" style="transform-origin: 24px 23px;">
                  <line x1="24" y1="6" x2="24" y2="9.5" />
                  <line x1="7" y1="23" x2="10.5" y2="23" />
                  <line x1="12" y1="11" x2="14.6" y2="13.6" />
                  <line x1="36" y1="11" x2="33.4" y2="13.6" />
                </g>
                <circle cx="24" cy="23" r="10.8" :fill="'url(#' + uid + '-fsunFill)'" />
                <path
                  d="M19 49H46C51.6 49 56 44.6 56 39.2C56 34.1 52 30 47 29.6C45.3 23.8 39.9 19.6 33.5 19.6C26.2 19.6 20.3 25 19.6 32C15.1 32.7 11.8 36.5 11.8 41C11.8 45.5 15 49 19 49Z"
                  :fill="'url(#' + uid + '-fcloudFill)'"
                  :filter="'url(#' + uid + '-softShadow)'"
                />
              </g>

              <!-- RAIN / HEAVY RAIN / NIGHT RAIN -->
              <g v-else-if="condition.key === 'rain' || condition.key === 'heavy_rain' || condition.key === 'night_rain'">
                <path
                  v-if="condition.key === 'night_rain'"
                  d="M44 10C39.2 10.6 35.8 14.9 36.4 19.8C37 24.6 41.4 28 46.2 27.4C48 27.2 49.6 26.4 50.9 25.2C47.4 25.7 44 23.4 43.3 19.9C42.6 16.4 44.7 13 48 11.9C46.8 10.7 45.4 10.1 44 10Z"
                  :fill="'url(#' + uid + '-fmoonFill)'"
                />
                <path
                  d="M17 41H46C51.4 41 55.6 36.8 55.6 31.6C55.6 26.7 51.8 22.8 47 22.4C45.4 16.9 40.3 12.9 34.2 12.9C27.3 12.9 21.6 18 20.9 24.7C16.7 25.4 13.5 29 13.5 33.3C13.5 37.6 14.9 41 17 41Z"
                  :fill="'url(#' + uid + '-fstormCloud)'"
                  :filter="'url(#' + uid + '-softShadow)'"
                />
                <g stroke="#0EA5E9" stroke-width="2.6" stroke-linecap="round" class="w-rain-drops">
                  <line x1="21" y1="45" x2="17.5" y2="55" />
                  <line x1="31" y1="46" x2="27.5" y2="57" />
                  <line x1="41" y1="45" x2="37.5" y2="55" />
                  <line v-if="condition.key === 'heavy_rain'" x1="49" y1="45" x2="45.5" y2="54" />
                </g>
              </g>

              <!-- NIGHT CLEAR / NIGHT CLOUDY -->
              <g v-else-if="condition.key === 'night_clear' || condition.key === 'night_cloudy'">
                <circle cx="32" cy="30" r="24" :fill="'url(#' + uid + '-fsunGlow)'" opacity="0.45" />
                <path
                  d="M35.5 11.5C26 12.6 19.2 21 20.3 30.5C21.4 40 29.8 46.8 39.3 45.7C42.9 45.3 46.1 43.7 48.6 41.4C41.8 42.3 35.1 37.9 33.7 31C32.3 24.1 36.4 17.4 42.9 15.2C40.7 12.9 38.2 11.7 35.5 11.5Z"
                  :fill="'url(#' + uid + '-fmoonFill)'"
                  :filter="'url(#' + uid + '-softShadow)'"
                />
                <path d="M49 14L50.1 16.9L53 18L50.1 19.1L49 22L47.9 19.1L45 18L47.9 16.9L49 14Z" fill="#FEF08A" />
                <circle cx="15" cy="18" r="1.5" fill="#BAE6FD" opacity="0.9" />
                <circle cx="52" cy="31" r="1.2" fill="#E0F2FE" opacity="0.8" />
                <path
                  v-if="condition.key === 'night_cloudy'"
                  d="M17 54H47C51.6 54 55.2 50.4 55.2 46C55.2 41.9 52 38.5 47.9 38.2C46.5 33.7 42.3 30.5 37.4 30.5C31.7 30.5 27 34.7 26.4 40.2C22.9 40.8 20.2 43.7 20.2 47.3C20.2 50.9 15 54 17 54Z"
                  :fill="'url(#' + uid + '-fcloudFill)'"
                  :filter="'url(#' + uid + '-softShadow)'"
                />
              </g>

              <!-- CLOUDY -->
              <g v-else>
                <path
                  d="M23 36H48C52.6 36 56.2 32.4 56.2 28C56.2 23.9 53 20.5 48.9 20.2C47.5 15.7 43.3 12.5 38.4 12.5C32.7 12.5 28 16.7 27.4 22.2C23.9 22.8 21.2 25.7 21.2 29.3C21.2 32.9 21 36 23 36Z"
                  :fill="'url(#' + uid + '-fstormCloud)'"
                  opacity="0.7"
                />
                <path
                  d="M16 49H45C50.4 49 54.6 44.8 54.6 39.6C54.6 34.7 50.8 30.8 46 30.4C44.4 24.9 39.3 20.9 33.2 20.9C26.3 20.9 20.6 26 19.9 32.7C15.7 33.4 12.5 37 12.5 41.3C12.5 45.6 13.9 49 16 49Z"
                  :fill="'url(#' + uid + '-fcloudFill)'"
                  :filter="'url(#' + uid + '-softShadow)'"
                />
              </g>
            </svg>
          </div>

          <!-- Primary Temperature Focal Readout -->
          <div
            v-if="weather.temperature"
            class="wh-temp-focal"
            :class="{ 'is-selected': isSelected(weather.temperature) }"
            @click="onSelectParam(weather.temperature)"
            @contextmenu.prevent="onContextParam($event, weather.temperature)"
            :title="buildParamTooltip(weather.temperature)"
          >
            <div class="wh-focal-lbl-row">
              <span class="wh-focal-lbl">TEMPERATUR</span>
              <span class="wh-comfort-tag" :class="'ct-' + tempComfort.tone">{{ tempComfort.label }}</span>
            </div>
            <div class="wh-temp-num-row">
              <span class="wh-temp-big tabular-nums">{{ formatVal(weather.temperature, 1) }}</span>
              <span class="wh-temp-unit">{{ resolveUnit(weather.temperature, '°C') }}</span>
            </div>
            <div class="wh-temp-bar-track">
              <div class="wh-temp-bar-fill" :style="{ width: tempComfort.pct + '%' }"></div>
            </div>
          </div>

          <!-- Fallback when station has Humidity but no Temperature -->
          <div
            v-else-if="weather.humidity"
            class="wh-temp-focal"
            :class="{ 'is-selected': isSelected(weather.humidity) }"
            @click="onSelectParam(weather.humidity)"
            @contextmenu.prevent="onContextParam($event, weather.humidity)"
            :title="buildParamTooltip(weather.humidity)"
          >
            <div class="wh-focal-lbl-row">
              <span class="wh-focal-lbl">HUMIDITY</span>
              <span class="wh-comfort-tag" :class="'ct-' + humidityComfort.tone">{{ humidityComfort.label }}</span>
            </div>
            <div class="wh-temp-num-row">
              <span class="wh-temp-big wh-hum-big tabular-nums">{{ formatVal(weather.humidity, 1) }}</span>
              <span class="wh-temp-unit wh-hum-unit">{{ resolveUnit(weather.humidity, '%RH') }}</span>
            </div>
            <div class="wh-temp-bar-track">
              <div class="wh-hum-bar-fill" :style="{ width: humidityComfort.pct + '%' }"></div>
            </div>
          </div>

          <!-- Fallback when station only has wind/rain/solar/pressure -->
          <div v-else class="wh-temp-focal">
            <div class="wh-focal-lbl-row">
              <span class="wh-focal-lbl">STASIUN CUACA</span>
            </div>
            <div class="text-xs font-bold leading-snug opacity-85">
              {{ weather.count }} Sensor Meteorologi Aktif
            </div>
          </div>
        </div>
      </div>

      <!-- 2. Dedicated Humidity Showcase Tile (Aqua Droplet + Circular Gauge Ring) -->
      <div
        v-if="weather.humidity && weather.temperature"
        class="wh-hum-tile"
        :class="{ 'is-selected': isSelected(weather.humidity) }"
        @click="onSelectParam(weather.humidity)"
        @contextmenu.prevent="onContextParam($event, weather.humidity)"
        :title="buildParamTooltip(weather.humidity)"
      >
        <div class="wh-hum-left">
          <div class="wh-drop-badge">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
              <defs>
                <linearGradient :id="uid + '-dropGrad'" x1="12" y1="2" x2="12" y2="22" gradientUnits="userSpaceOnUse">
                  <stop offset="0%" stop-color="#38BDF8" />
                  <stop offset="100%" stop-color="#0284C7" />
                </linearGradient>
              </defs>
              <path
                d="M12 2.5C12 2.5 5.5 10.2 5.5 14.8C5.5 18.4 8.4 21.3 12 21.3C15.6 21.3 18.5 18.4 18.5 14.8C18.5 10.2 12 2.5 12 2.5Z"
                :fill="'url(#' + uid + '-dropGrad)'"
              />
              <path d="M9.5 15.2C9.5 16.9 10.7 18.1 12.2 18.2" stroke="#E0F2FE" stroke-width="1.6" stroke-linecap="round" opacity="0.85" />
            </svg>
          </div>
          <div class="min-w-0">
            <div class="wh-hum-top">
              <span class="wh-sub-kicker">HUMIDITY</span>
              <span class="wh-comfort-tag" :class="'ct-' + humidityComfort.tone">{{ humidityComfort.label }}</span>
            </div>
            <div class="wh-hum-val-row">
              <strong class="wh-hum-num tabular-nums">{{ formatVal(weather.humidity, 1) }}</strong>
              <span class="wh-hum-u">{{ resolveUnit(weather.humidity, '%RH') }}</span>
            </div>
          </div>
        </div>

        <!-- Precision Humidity Circular Ring Gauge -->
        <svg width="42" height="42" viewBox="0 0 44 44" class="flex-shrink-0">
          <defs>
            <linearGradient :id="uid + '-rhRing'" x1="0" y1="0" x2="44" y2="44" gradientUnits="userSpaceOnUse">
              <stop offset="0%" stop-color="#06B6D4" />
              <stop offset="100%" stop-color="#0284C7" />
            </linearGradient>
          </defs>
          <circle cx="22" cy="22" r="17.5" fill="none" :stroke="isDark ? '#1B2D48' : '#DBEAFE'" stroke-width="4.2" />
          <circle
            cx="22"
            cy="22"
            r="17.5"
            fill="none"
            :stroke="'url(#' + uid + '-rhRing)'"
            stroke-width="4.2"
            stroke-dasharray="110"
            :stroke-dashoffset="humidityRingOffset"
            stroke-linecap="round"
            transform="rotate(-90 22 22)"
            style="transition: stroke-dashoffset 0.45s ease;"
          />
          <text
            x="22"
            y="25.5"
            text-anchor="middle"
            :fill="isDark ? '#E0F2FE' : '#0369A1'"
            font-size="10"
            font-weight="800"
            font-family="'Plus Jakarta Sans', 'Inter', sans-serif"
          >{{ humidityComfort.pct }}%</text>
        </svg>
      </div>

      <!-- 3. Precision Rotating SVG Wind Compass Banner -->
      <div
        v-if="weather.windSpeed || weather.windAngle"
        class="wh-wind-band"
        :class="{ 'is-selected': isSelected(weather.windSpeed) || isSelected(weather.windAngle) }"
        @click="onSelectParam(weather.windSpeed || weather.windAngle)"
        @contextmenu.prevent="onContextParam($event, weather.windSpeed || weather.windAngle)"
        :title="windTooltip"
      >
        <svg width="44" height="44" viewBox="0 0 48 48" fill="none" class="flex-shrink-0">
          <circle cx="24" cy="24" r="21" :fill="isDark ? '#0D192B' : '#F8FAFC'" :stroke="isDark ? 'rgba(148,163,184,0.22)' : '#D6E2EE'" stroke-width="1.4" />
          <circle cx="24" cy="24" r="16.5" fill="none" :stroke="isDark ? 'rgba(56,189,248,0.16)' : 'rgba(2,132,199,0.14)'" stroke-width="1" stroke-dasharray="2 3" />
          <text x="24" y="9.5" text-anchor="middle" :fill="isDark ? '#38BDF8' : '#0284C7'" font-size="6.8" font-weight="800">N</text>
          <text x="39.5" y="26.3" text-anchor="middle" :fill="isDark ? '#64748B' : '#94A3B8'" font-size="6" font-weight="700">E</text>
          <text x="24" y="42.5" text-anchor="middle" :fill="isDark ? '#64748B' : '#94A3B8'" font-size="6" font-weight="700">S</text>
          <text x="8.5" y="26.3" text-anchor="middle" :fill="isDark ? '#64748B' : '#94A3B8'" font-size="6" font-weight="700">W</text>
          <!-- Rotating Wind Needle -->
          <g :style="compassNeedleStyle" class="w-compass-needle">
            <polygon points="24,10.5 27.6,24 24,21.2 20.4,24" fill="#F97316" />
            <polygon points="24,37.5 20.4,24 24,26.8 27.6,24" :fill="isDark ? '#475569' : '#94A3B8'" />
          </g>
          <circle cx="24" cy="24" r="2.4" :fill="isDark ? '#F8FAFC' : '#0F172A'" />
        </svg>

        <div class="min-w-0 flex-1">
          <div class="wh-wind-top">
            <span class="wh-sub-kicker">ANGIN</span>
            <span class="wh-wind-beaufort">{{ windBeaufort }}</span>
          </div>
          <div class="wh-wind-main">
            <strong v-if="weather.windSpeed" class="wh-wind-spd tabular-nums">
              {{ formatVal(weather.windSpeed, 2) }}
              <span>{{ resolveUnit(weather.windSpeed, 'm/s') }}</span>
            </strong>
            <span v-if="weather.windAngle && windCardinal.valid" class="wh-wind-deg-pill tabular-nums">
              {{ windCardinal.deg }}° {{ windCardinal.short }}
            </span>
          </div>
          <div v-if="weather.windAngle && windCardinal.valid" class="wh-wind-sub truncate">
            Arah {{ windCardinal.label }}
          </div>
        </div>
      </div>

      <!-- 4. Secondary Meteorology Sensor Chips (Solar Radiation, Air Pressure, Rain Fall) -->
      <div v-if="secondaryChips.length > 0" class="wh-chips-grid" :class="'chips-n-' + secondaryChips.length">
        <button
          v-for="chip in secondaryChips"
          :key="'wh-' + chip.key"
          type="button"
          class="wh-meteo-chip"
          :class="[chip.toneClass, { 'is-selected': isSelected(chip.param) }]"
          @click="onSelectParam(chip.param)"
          @contextmenu.prevent="onContextParam($event, chip.param)"
          :title="buildParamTooltip(chip.param)"
        >
          <div class="wh-chip-top">
            <!-- Solar SVG -->
            <svg v-if="chip.key === 'solar'" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#F59E0B" stroke-width="2.2" stroke-linecap="round">
              <circle cx="12" cy="12" r="4.5" fill="#FDE047" />
              <line x1="12" y1="2" x2="12" y2="4.5" />
              <line x1="12" y1="19.5" x2="12" y2="22" />
              <line x1="2" y1="12" x2="4.5" y2="12" />
              <line x1="19.5" y1="12" x2="22" y2="12" />
              <line x1="4.9" y1="4.9" x2="6.7" y2="6.7" />
              <line x1="17.3" y1="17.3" x2="19.1" y2="19.1" />
              <line x1="19.1" y1="4.9" x2="17.3" y2="6.7" />
              <line x1="6.7" y1="17.3" x2="4.9" y2="19.1" />
            </svg>
            <!-- Pressure SVG -->
            <svg v-else-if="chip.key === 'pressure'" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#6366F1" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="9" />
              <path d="M12 12L16.5 8.5" />
              <circle cx="12" cy="12" r="1.5" fill="#6366F1" />
            </svg>
            <!-- Rain SVG -->
            <svg v-else width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#0EA5E9" stroke-width="2.1" stroke-linecap="round" stroke-linejoin="round">
              <path d="M16 13a4 4 0 0 0 .5-7.97A5.5 5.5 0 0 0 6 6.5a3.5 3.5 0 0 0 .5 6.96" />
              <line x1="8.5" y1="16.5" x2="7.5" y2="20" />
              <line x1="12.5" y1="17" x2="11.5" y2="20.5" />
              <line x1="16.5" y1="16.5" x2="15.5" y2="20" />
            </svg>
            <span class="wh-chip-name truncate">{{ chip.shortLabel }}</span>
          </div>
          <div class="wh-chip-val tabular-nums">{{ chip.formatted }}</div>
          <div class="wh-chip-unit">{{ chip.unit }}</div>
        </button>
      </div>

      <!-- 5. Bottom Footer Actions (Matches BentoISPUHeroCard .hero-actions height & feel) -->
      <div class="wh-footer-actions">
        <button
          type="button"
          class="wh-act-btn primary"
          @click="onSelectAllWeatherTrend"
          title="Fokuskan grafik Analisis Tren pada semua parameter cuaca"
        >
          <svg width="11.5" height="11.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.3" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
          </svg>
          <span>Tren Cuaca</span>
        </button>

        <button
          v-if="canToggleDedup"
          type="button"
          class="wh-act-btn"
          :class="{ 'is-active': !hideDuplicatesInGrid }"
          @click="$emit('toggle-dedup')"
          :title="hideDuplicatesInGrid ? 'Tampilkan juga kartu/baris sensor cuaca di panel tengah' : 'Sembunyikan duplikat sensor cuaca di panel tengah agar lebih ringkas'"
        >
          <span>{{ hideDuplicatesInGrid ? 'Grid: Ringkas' : 'Grid: Tampil' }}</span>
        </button>
      </div>
    </template>
  </div>
</template>

<script>
import {
  parseNumericValue,
  getCardinalDirection,
  inferWeatherCondition,
  getTempComfortMeta,
  getHumidityComfortMeta,
  getWindBeaufortLabel,
} from '../utils/weatherParams';

let weatherCardUidCounter = 0;

export default {
  name: 'BentoWeatherHeroCard',
  props: {
    weather: {
      type: Object,
      required: true,
    },
    isDark: {
      type: Boolean,
      default: false,
    },
    compact: {
      type: Boolean,
      default: false,
    },
    animationsEnabled: {
      type: Boolean,
      default: true,
    },
    selectedParamIds: {
      type: Array,
      default: () => [],
    },
    hideDuplicatesInGrid: {
      type: Boolean,
      default: true,
    },
    canToggleDedup: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    weatherCardUidCounter += 1;
    return {
      uid: 'bwh-' + weatherCardUidCounter,
      clockTick: Date.now(),
      clockTimer: null,
    };
  },
  mounted() {
    this.clockTimer = setInterval(() => {
      this.clockTick = Date.now();
    }, 60000);
  },
  beforeDestroy() {
    if (this.clockTimer) {
      clearInterval(this.clockTimer);
      this.clockTimer = null;
    }
  },
  computed: {
    selectedIdSet() {
      return new Set((this.selectedParamIds || []).map(id => Number(id)));
    },
    tempVal() {
      return parseNumericValue(this.weather && this.weather.temperature);
    },
    humidityVal() {
      return parseNumericValue(this.weather && this.weather.humidity);
    },
    windSpeedVal() {
      return parseNumericValue(this.weather && this.weather.windSpeed);
    },
    windAngleVal() {
      return parseNumericValue(this.weather && this.weather.windAngle);
    },
    solarVal() {
      return parseNumericValue(this.weather && this.weather.solarRadiation);
    },
    pressureVal() {
      return parseNumericValue(this.weather && this.weather.pressure);
    },
    rainVal() {
      return parseNumericValue(this.weather && this.weather.rainFall);
    },
    condition() {
      return inferWeatherCondition({
        temperature: this.tempVal,
        humidity: this.humidityVal,
        solarRadiation: this.solarVal,
        rainFall: this.rainVal,
        nowDate: new Date(this.clockTick),
      });
    },
    tempComfort() {
      return getTempComfortMeta(this.tempVal);
    },
    humidityComfort() {
      return getHumidityComfortMeta(this.humidityVal);
    },
    humidityRingOffset() {
      const pct = Math.max(0, Math.min(100, this.humidityComfort.pct || 0));
      return (110 - (pct / 100) * 110).toFixed(1);
    },
    windCardinal() {
      return getCardinalDirection(this.windAngleVal);
    },
    windBeaufort() {
      return getWindBeaufortLabel(this.windSpeedVal);
    },
    compassNeedleStyle() {
      const deg = this.windCardinal.valid ? this.windCardinal.deg : 0;
      return {
        transform: `rotate(${deg}deg)`,
        transformOrigin: '24px 24px',
      };
    },
    compactCompassNeedleStyle() {
      const deg = this.windCardinal.valid ? this.windCardinal.deg : 0;
      return {
        transform: `rotate(${deg}deg)`,
        transformOrigin: '18px 18px',
      };
    },
    windTooltip() {
      const parts = [];
      if (this.weather.windSpeed) {
        parts.push(`Kecepatan Angin: ${this.formatVal(this.weather.windSpeed, 2)} ${this.resolveUnit(this.weather.windSpeed, 'm/s')}`);
      }
      if (this.weather.windAngle && this.windCardinal.valid) {
        parts.push(`Arah Angin: ${this.windCardinal.deg}° (${this.windCardinal.label})`);
      }
      parts.push('Klik: Tren · Klik Kanan: Opsi');
      return parts.join(' · ');
    },
    secondaryChips() {
      const chips = [];
      if (this.weather.solarRadiation) {
        chips.push({
          key: 'solar',
          shortLabel: 'Solar',
          param: this.weather.solarRadiation,
          formatted: this.formatVal(this.weather.solarRadiation, this.solarVal !== null && this.solarVal >= 100 ? 0 : 1),
          unit: this.resolveUnit(this.weather.solarRadiation, 'W/m²'),
          toneClass: 'chip-solar',
        });
      }
      if (this.weather.pressure) {
        chips.push({
          key: 'pressure',
          shortLabel: 'Tekanan',
          param: this.weather.pressure,
          formatted: this.formatVal(this.weather.pressure, this.pressureVal !== null && this.pressureVal >= 100 ? 0 : 1),
          unit: this.resolveUnit(this.weather.pressure, 'hPa'),
          toneClass: 'chip-pressure',
        });
      }
      if (this.weather.rainFall) {
        chips.push({
          key: 'rain',
          shortLabel: 'Hujan',
          param: this.weather.rainFall,
          formatted: this.formatVal(this.weather.rainFall, 2),
          unit: this.resolveUnit(this.weather.rainFall, 'mm'),
          toneClass: 'chip-rain',
        });
      }
      return chips;
    },
  },
  methods: {
    isSelected(param) {
      if (!param) return false;
      const pid = Number(param.param_id != null ? param.param_id : param.id);
      return Number.isFinite(pid) && this.selectedIdSet.has(pid);
    },
    formatVal(param, decimals = 1) {
      const n = parseNumericValue(param);
      if (n === null) return '—';
      return n.toFixed(decimals);
    },
    resolveUnit(param, fallback) {
      const u = param && param.unit ? String(param.unit).trim() : '';
      return u || fallback;
    },
    buildParamTooltip(param) {
      if (!param) return '';
      const name = param.param_name || param.param_code || 'Weather Sensor';
      const val = this.formatVal(param, 2);
      const unit = param.unit ? ` ${param.unit}` : '';
      const qual = param.quality ? ` (${param.quality})` : '';
      return `${name}: ${val}${unit}${qual} · Klik untuk Tren · Klik Kanan untuk Download/Opsi`;
    },
    onSelectParam(param) {
      if (!param) return;
      this.$emit('select-param', param);
    },
    onContextParam(event, param) {
      if (!param) return;
      this.$emit('context-param', { event, param });
    },
    onSelectAllWeatherTrend() {
      this.$emit('focus-weather-trend', this.weather.matchedParams || []);
    },
  },
};
</script>

<style scoped>
.bento-weather-root {
  font-family: 'Plus Jakarta Sans', 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
  position: relative;
  min-width: 0;
}

.bento-weather-card {
  border-radius: 13px;
  padding: 11px 12px 10px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 8px;
  height: 100%;
  min-height: 310px;
}

.bento-weather-root.is-day {
  background: linear-gradient(165deg, #FFFFFF 0%, #F4FAFF 100%);
  border: 1px solid #DDE5EF;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.03), 0 6px 18px rgba(15, 23, 42, 0.03);
  color: #0F172A;
}

.bento-weather-root.is-night {
  background: linear-gradient(165deg, #132038 0%, #0E1829 100%);
  border: 1px solid rgba(148, 163, 184, 0.15);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.28);
  color: #F8FAFC;
}

/* ===================================================== */
/* 1. ATMOSPHERIC SKY HEADER BAND                        */
/* ===================================================== */
.wh-sky-band {
  border-radius: 11px;
  padding: 9px 10px 9px;
  display: flex;
  flex-direction: column;
  gap: 7px;
  transition: background 0.3s ease;
}

.is-day .wh-sky-band.sky-sunny {
  background: linear-gradient(145deg, #E8F5FF 0%, #F2F9FF 56%, #FFF8E7 100%);
  border: 1px solid #D3E8FA;
}
.is-day .wh-sky-band.sky-partly {
  background: linear-gradient(145deg, #E6F4FF 0%, #F3F8FD 65%, #FFFBEB 100%);
  border: 1px solid #D4E6F6;
}
.is-day .wh-sky-band.sky-cloudy {
  background: linear-gradient(145deg, #EDF2F8 0%, #F5F8FC 100%);
  border: 1px solid #DCE4EE;
}
.is-day .wh-sky-band.sky-rain {
  background: linear-gradient(145deg, #E2F1FA 0%, #EDF5FB 100%);
  border: 1px solid #C6E2F5;
}
.is-day .wh-sky-band.sky-night {
  background: linear-gradient(145deg, #1E293B 0%, #0F172A 100%);
  border: 1px solid #334155;
  color: #F8FAFC;
}

.is-night .wh-sky-band.sky-sunny,
.is-night .wh-sky-band.sky-partly {
  background: linear-gradient(145deg, rgba(14, 165, 233, 0.18) 0%, rgba(245, 158, 11, 0.11) 100%);
  border: 1px solid rgba(56, 189, 248, 0.24);
}
.is-night .wh-sky-band.sky-cloudy {
  background: linear-gradient(145deg, rgba(100, 116, 139, 0.2) 0%, rgba(30, 41, 59, 0.55) 100%);
  border: 1px solid rgba(148, 163, 184, 0.2);
}
.is-night .wh-sky-band.sky-rain {
  background: linear-gradient(145deg, rgba(14, 165, 233, 0.22) 0%, rgba(30, 58, 138, 0.34) 100%);
  border: 1px solid rgba(56, 189, 248, 0.26);
}
.is-night .wh-sky-band.sky-night {
  background: linear-gradient(145deg, rgba(30, 58, 138, 0.32) 0%, rgba(15, 23, 42, 0.72) 100%);
  border: 1px solid rgba(99, 102, 241, 0.28);
}

.wh-top-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.wh-kicker {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.04em;
  white-space: nowrap;
}
.is-day .wh-kicker { color: #475569; }
.is-day .sky-night .wh-kicker { color: #CBD5E1; }
.is-night .wh-kicker { color: #94A3B8; }

.wh-cond-badge {
  display: inline-flex;
  align-items: center;
  gap: 4.5px;
  padding: 2px 7.5px;
  border-radius: 999px;
  font-size: 9.5px;
  font-weight: 700;
  white-space: nowrap;
  border: 1px solid transparent;
}
.is-day .wh-cond-badge.cond-sunny { background: #FEF3C7; color: #B45309; border-color: #FDE68A; }
.is-day .wh-cond-badge.cond-partly { background: #E0F2FE; color: #0369A1; border-color: #BAE6FD; }
.is-day .wh-cond-badge.cond-cloudy { background: #E2E8F0; color: #334155; border-color: #CBD5E1; }
.is-day .wh-cond-badge.cond-rain { background: #DBEAFE; color: #1D4ED8; border-color: #BFDBFE; }
.is-day .wh-cond-badge.cond-night { background: rgba(99, 102, 241, 0.22); color: #E0E7FF; border-color: rgba(129, 140, 248, 0.4); }

.is-night .wh-cond-badge.cond-sunny { background: rgba(245, 158, 11, 0.18); color: #FCD34D; border-color: rgba(245, 158, 11, 0.35); }
.is-night .wh-cond-badge.cond-partly { background: rgba(56, 189, 248, 0.16); color: #38BDF8; border-color: rgba(56, 189, 248, 0.32); }
.is-night .wh-cond-badge.cond-cloudy { background: rgba(148, 163, 184, 0.16); color: #CBD5E1; border-color: rgba(148, 163, 184, 0.3); }
.is-night .wh-cond-badge.cond-rain { background: rgba(14, 165, 233, 0.2); color: #7DD3FC; border-color: rgba(56, 189, 248, 0.38); }
.is-night .wh-cond-badge.cond-night { background: rgba(99, 102, 241, 0.22); color: #C7D2FE; border-color: rgba(129, 140, 248, 0.36); }

.wh-live-dot {
  width: 5.5px;
  height: 5.5px;
  border-radius: 999px;
  background: currentColor;
  animation: whPulse 2.2s ease-in-out infinite;
}
.no-anim .wh-live-dot,
.no-anim .w-sun-rays,
.no-anim .w-rain-drops {
  animation: none !important;
}

@keyframes whPulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.4; transform: scale(0.8); }
}

.wh-hero-stage {
  display: flex;
  align-items: center;
  gap: 10px;
}

.wh-art-wrap {
  width: 62px;
  height: 62px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.wh-temp-focal {
  flex: 1;
  min-width: 0;
  padding: 4px 6px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s ease, box-shadow 0.15s ease;
}
.is-day .wh-temp-focal:hover { background: rgba(255, 255, 255, 0.55); }
.is-day .sky-night .wh-temp-focal:hover { background: rgba(255, 255, 255, 0.08); }
.is-night .wh-temp-focal:hover { background: rgba(255, 255, 255, 0.06); }
.wh-temp-focal.is-selected {
  box-shadow: inset 0 0 0 1.5px #0EA5E9;
}

.wh-focal-lbl-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  margin-bottom: 1px;
}

.wh-focal-lbl {
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.05em;
  opacity: 0.72;
}

.wh-temp-num-row {
  display: flex;
  align-items: baseline;
  gap: 3px;
  line-height: 1.05;
}

.wh-temp-big {
  font-size: 28px;
  font-weight: 800;
  letter-spacing: -0.03em;
}
.is-day .wh-temp-big { color: #0F172A; }
.is-day .sky-night .wh-temp-big { color: #FFFFFF; }
.is-night .wh-temp-big { color: #F8FAFC; }

.wh-temp-unit {
  font-size: 13px;
  font-weight: 800;
  color: #EA580C;
}
.is-night .wh-temp-unit,
.is-day .sky-night .wh-temp-unit {
  color: #FBBF24;
}

.wh-hum-unit {
  color: #0284C7 !important;
}
.is-night .wh-hum-unit {
  color: #38BDF8 !important;
}

.wh-temp-bar-track {
  margin-top: 5px;
  height: 4px;
  border-radius: 999px;
  overflow: hidden;
  background: rgba(148, 163, 184, 0.24);
}
.wh-temp-bar-fill {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #38BDF8 0%, #F59E0B 55%, #EF4444 100%);
  transition: width 0.35s ease;
}
.wh-hum-bar-fill {
  height: 100%;
  border-radius: 999px;
  background: linear-gradient(90deg, #06B6D4 0%, #0284C7 100%);
  transition: width 0.35s ease;
}

/* Comfort Tag Pills */
.wh-comfort-tag {
  font-size: 8.5px;
  font-weight: 800;
  padding: 1px 5.5px;
  border-radius: 999px;
  line-height: 1.3;
  white-space: nowrap;
}
.is-day .wh-comfort-tag.ct-sky { background: #E0F2FE; color: #0369A1; }
.is-day .wh-comfort-tag.ct-emerald { background: #DCFCE7; color: #15803D; }
.is-day .wh-comfort-tag.ct-amber { background: #FEF3C7; color: #B45309; }
.is-day .wh-comfort-tag.ct-orange { background: #FFEDD5; color: #C2410C; }
.is-day .wh-comfort-tag.ct-rose { background: #FFE4E6; color: #BE123C; }
.is-day .wh-comfort-tag.ct-indigo { background: #E0E7FF; color: #4338CA; }
.is-day .wh-comfort-tag.ct-slate { background: #F1F5F9; color: #475569; }

.is-night .wh-comfort-tag.ct-sky,
.is-day .sky-night .wh-comfort-tag.ct-sky { background: rgba(56, 189, 248, 0.2); color: #7DD3FC; }
.is-night .wh-comfort-tag.ct-emerald,
.is-day .sky-night .wh-comfort-tag.ct-emerald { background: rgba(16, 185, 129, 0.2); color: #6EE7B7; }
.is-night .wh-comfort-tag.ct-amber,
.is-day .sky-night .wh-comfort-tag.ct-amber { background: rgba(245, 158, 11, 0.2); color: #FCD34D; }
.is-night .wh-comfort-tag.ct-orange,
.is-day .sky-night .wh-comfort-tag.ct-orange { background: rgba(249, 115, 22, 0.22); color: #FDBA74; }
.is-night .wh-comfort-tag.ct-rose,
.is-day .sky-night .wh-comfort-tag.ct-rose { background: rgba(244, 63, 94, 0.22); color: #FDA4AF; }
.is-night .wh-comfort-tag.ct-indigo,
.is-day .sky-night .wh-comfort-tag.ct-indigo { background: rgba(99, 102, 241, 0.24); color: #C7D2FE; }
.is-night .wh-comfort-tag.ct-slate,
.is-day .sky-night .wh-comfort-tag.ct-slate { background: rgba(148, 163, 184, 0.2); color: #CBD5E1; }

/* ===================================================== */
/* 2. HUMIDITY SHOWCASE TILE                             */
/* ===================================================== */
.wh-hum-tile {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 7.5px 10px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.16s ease;
}
.is-day .wh-hum-tile {
  background: linear-gradient(135deg, #ECFEFF 0%, #F0F9FF 100%);
  border: 1px solid #CFEAF8;
}
.is-day .wh-hum-tile:hover {
  border-color: #7DD3FC;
  box-shadow: 0 2px 8px rgba(2, 132, 199, 0.08);
}
.is-night .wh-hum-tile {
  background: linear-gradient(135deg, rgba(6, 182, 212, 0.12) 0%, rgba(14, 165, 233, 0.08) 100%);
  border: 1px solid rgba(56, 189, 248, 0.22);
}
.is-night .wh-hum-tile:hover {
  border-color: rgba(56, 189, 248, 0.45);
}
.wh-hum-tile.is-selected {
  border-color: #0EA5E9 !important;
  box-shadow: 0 0 0 1px rgba(14, 165, 233, 0.35);
}

.wh-hum-left {
  display: flex;
  align-items: center;
  gap: 9px;
  min-width: 0;
  flex: 1;
}

.wh-drop-badge {
  width: 34px;
  height: 34px;
  border-radius: 9px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.is-day .wh-drop-badge {
  background: rgba(255, 255, 255, 0.85);
  border: 1px solid #BAE6FD;
}
.is-night .wh-drop-badge {
  background: rgba(14, 165, 233, 0.15);
  border: 1px solid rgba(56, 189, 248, 0.25);
}

.wh-hum-top {
  display: flex;
  align-items: center;
  gap: 5px;
}

.wh-sub-kicker {
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.05em;
}
.is-day .wh-sub-kicker { color: #64748B; }
.is-night .wh-sub-kicker { color: #94A3B8; }

.wh-hum-val-row {
  display: flex;
  align-items: baseline;
  gap: 3px;
  line-height: 1.1;
  margin-top: 1px;
}

.wh-hum-num {
  font-size: 20px;
  font-weight: 800;
  letter-spacing: -0.02em;
}
.is-day .wh-hum-num { color: #0C4A6E; }
.is-night .wh-hum-num { color: #E0F2FE; }

.wh-hum-u {
  font-size: 10.5px;
  font-weight: 700;
}
.is-day .wh-hum-u { color: #0284C7; }
.is-night .wh-hum-u { color: #38BDF8; }

/* ===================================================== */
/* 3. ROTATING WIND COMPASS BANNER                       */
/* ===================================================== */
.wh-wind-band {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 9.5px;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.16s ease;
}
.is-day .wh-wind-band {
  background: #F8FBFF;
  border: 1px solid #E2ECF5;
}
.is-day .wh-wind-band:hover {
  border-color: #93C5FD;
}
.is-night .wh-wind-band {
  background: rgba(15, 23, 42, 0.55);
  border: 1px solid rgba(148, 163, 184, 0.14);
}
.is-night .wh-wind-band:hover {
  border-color: rgba(56, 189, 248, 0.4);
}
.wh-wind-band.is-selected {
  border-color: #0EA5E9 !important;
  box-shadow: 0 0 0 1px rgba(14, 165, 233, 0.35);
}

.w-compass-needle {
  transition: transform 0.6s cubic-bezier(0.22, 1, 0.36, 1);
}

.wh-wind-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
}

.wh-wind-beaufort {
  font-size: 9px;
  font-weight: 700;
}
.is-day .wh-wind-beaufort { color: #0284C7; }
.is-night .wh-wind-beaufort { color: #38BDF8; }

.wh-wind-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  margin-top: 1px;
}

.wh-wind-spd {
  font-size: 14.5px;
  font-weight: 800;
  line-height: 1.15;
}
.wh-wind-spd span {
  font-size: 10px;
  font-weight: 700;
  opacity: 0.72;
}
.is-day .wh-wind-spd { color: #0F172A; }
.is-night .wh-wind-spd { color: #F8FAFC; }

.wh-wind-deg-pill {
  font-size: 9.5px;
  font-weight: 800;
  padding: 1.5px 6px;
  border-radius: 6px;
  white-space: nowrap;
}
.is-day .wh-wind-deg-pill {
  background: #FFF7ED;
  border: 1px solid #FED7AA;
  color: #C2410C;
}
.is-night .wh-wind-deg-pill {
  background: rgba(249, 115, 22, 0.16);
  border: 1px solid rgba(249, 115, 22, 0.34);
  color: #FDBA74;
}

.wh-wind-sub {
  font-size: 9.5px;
  font-weight: 600;
  margin-top: 1px;
}
.is-day .wh-wind-sub { color: #64748B; }
.is-night .wh-wind-sub { color: #94A3B8; }

/* ===================================================== */
/* 4. SECONDARY METEO CHIPS (SOLAR, PRESSURE, RAIN)      */
/* ===================================================== */
.wh-chips-grid {
  display: grid;
  gap: 5.5px;
}
.wh-chips-grid.chips-n-1 { grid-template-columns: 1fr; }
.wh-chips-grid.chips-n-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.wh-chips-grid.chips-n-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); }

.wh-meteo-chip {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1.5px;
  padding: 6px 7px;
  border-radius: 9px;
  text-align: left;
  cursor: pointer;
  min-width: 0;
  transition: all 0.15s ease;
}
.is-day .wh-meteo-chip {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
}
.is-day .wh-meteo-chip:hover {
  background: #F0F9FF;
  border-color: #BAE6FD;
  transform: translateY(-1px);
}
.is-night .wh-meteo-chip {
  background: #142238;
  border: 1px solid rgba(148, 163, 184, 0.14);
}
.is-night .wh-meteo-chip:hover {
  border-color: rgba(56, 189, 248, 0.4);
  transform: translateY(-1px);
}
.wh-meteo-chip.is-selected {
  border-color: #0EA5E9 !important;
  box-shadow: 0 0 0 1px rgba(14, 165, 233, 0.35);
}

.wh-chip-top {
  display: flex;
  align-items: center;
  gap: 3.5px;
  width: 100%;
  min-width: 0;
}

.wh-chip-name {
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.02em;
}
.is-day .wh-chip-name { color: #64748B; }
.is-night .wh-chip-name { color: #94A3B8; }

.wh-chip-val {
  font-size: 13px;
  font-weight: 800;
  line-height: 1.15;
  margin-top: 1px;
}
.is-day .wh-chip-val { color: #0F172A; }
.is-night .wh-chip-val { color: #F8FAFC; }

.wh-chip-unit {
  font-size: 8.5px;
  font-weight: 700;
  opacity: 0.72;
}
.is-day .wh-chip-unit { color: #475569; }
.is-night .wh-chip-unit { color: #94A3B8; }

/* ===================================================== */
/* 5. FOOTER ACTION BUTTONS                              */
/* ===================================================== */
.wh-footer-actions {
  display: flex;
  align-items: center;
  gap: 5px;
  padding-top: 2px;
}

.wh-act-btn {
  flex: 1;
  height: 27px;
  border-radius: 7px;
  font-size: 10px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  cursor: pointer;
  white-space: nowrap;
  padding: 0 8px;
  transition: all 0.15s ease;
}
.is-day .wh-act-btn {
  background: #F8FAFC;
  border: 1px solid #DDE5EF;
  color: #334155;
}
.is-day .wh-act-btn:hover,
.is-day .wh-act-btn.is-active {
  border-color: #0284C7;
  color: #0284C7;
  background: #F0F9FF;
}
.is-night .wh-act-btn {
  background: #162238;
  border: 1px solid rgba(148, 163, 184, 0.16);
  color: #CBD5E1;
}
.is-night .wh-act-btn:hover,
.is-night .wh-act-btn.is-active {
  border-color: #38BDF8;
  color: #38BDF8;
}
.wh-act-btn.primary {
  background: #0284C7 !important;
  border-color: #0284C7 !important;
  color: #FFFFFF !important;
}
.wh-act-btn.primary:hover {
  background: #0369A1 !important;
}

/* ===================================================== */
/* COMPACT HORIZONTAL WEATHER BAR (< 1280px)             */
/* ===================================================== */
.is-compact-weather-bar {
  border-radius: 12px;
  padding: 6px 11px;
}

.cw-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
}

.cw-primary {
  display: flex;
  align-items: center;
  gap: 9px;
  flex-wrap: wrap;
  min-width: 0;
}

.cw-icon-wrap {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.cw-metric-btn {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  padding: 2px 6px;
  border-radius: 7px;
  border: 1px solid transparent;
  background: transparent;
  cursor: pointer;
  transition: all 0.15s ease;
}
.is-day .cw-metric-btn:hover { background: #F1F5F9; }
.is-night .cw-metric-btn:hover { background: rgba(255, 255, 255, 0.06); }
.cw-metric-btn.is-selected {
  border-color: #0EA5E9;
}

.cw-kicker {
  font-size: 8.5px;
  font-weight: 800;
  letter-spacing: 0.05em;
  opacity: 0.65;
  line-height: 1;
}

.cw-val-row {
  display: flex;
  align-items: baseline;
  gap: 2.5px;
  line-height: 1.15;
}

.cw-temp-val {
  font-size: 16px;
  font-weight: 800;
}
.is-day .cw-temp-val { color: #EA580C; }
.is-night .cw-temp-val { color: #FBBF24; }

.cw-hum-val {
  font-size: 16px;
  font-weight: 800;
}
.is-day .cw-hum-val { color: #0284C7; }
.is-night .cw-hum-val { color: #38BDF8; }

.cw-unit {
  font-size: 9.5px;
  font-weight: 700;
  opacity: 0.72;
}

.cw-vdiv {
  width: 1px;
  height: 24px;
  background: rgba(148, 163, 184, 0.25);
}

.cw-cond-pill {
  font-size: 10px;
  font-weight: 700;
  padding: 2.5px 8px;
  border-radius: 999px;
  white-space: nowrap;
}
.is-day .cw-cond-pill { background: #E0F2FE; color: #0369A1; border: 1px solid #BAE6FD; }
.is-night .cw-cond-pill { background: rgba(56, 189, 248, 0.15); color: #7DD3FC; border: 1px solid rgba(56, 189, 248, 0.28); }

.cw-extras {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.cw-wind-chip {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 26px;
  padding: 0 8px;
  border-radius: 8px;
  font-size: 10.5px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.is-day .cw-wind-chip { background: #F8FAFC; border: 1px solid #DDE5EF; color: #0F172A; }
.is-night .cw-wind-chip { background: #162238; border: 1px solid rgba(148, 163, 184, 0.18); color: #F8FAFC; }
.cw-wind-chip.is-selected { border-color: #0EA5E9; }

.cw-dir-tag {
  font-size: 9.5px;
  font-weight: 800;
  padding: 1px 5px;
  border-radius: 5px;
  color: #EA580C;
  background: rgba(249, 115, 22, 0.12);
}
.is-night .cw-dir-tag { color: #FDBA74; }

.cw-mini-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  height: 26px;
  padding: 0 7.5px;
  border-radius: 8px;
  font-size: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
}
.is-day .cw-mini-chip { background: #F8FAFC; border: 1px solid #E2E8F0; color: #0F172A; }
.is-night .cw-mini-chip { background: #162238; border: 1px solid rgba(148, 163, 184, 0.16); color: #F8FAFC; }
.cw-mini-chip.is-selected { border-color: #0EA5E9; }

.cw-chip-lbl {
  font-size: 9px;
  font-weight: 700;
  opacity: 0.68;
}
.cw-chip-unit {
  font-size: 8.5px;
  opacity: 0.68;
}

.cw-trend-btn {
  height: 26px;
  padding: 0 9px;
  border-radius: 7px;
  font-size: 10px;
  font-weight: 700;
  background: #0284C7;
  border: 1px solid #0284C7;
  color: #FFFFFF;
  cursor: pointer;
  white-space: nowrap;
  transition: background 0.15s ease;
}
.cw-trend-btn:hover {
  background: #0369A1;
}

/* Subtle SVG Animations */
.w-sun-rays {
  transform-origin: 32px 32px;
  animation: wSlowSpin 28s linear infinite;
}
@keyframes wSlowSpin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.w-rain-drops {
  animation: wRainFall 1.3s ease-in-out infinite;
}
@keyframes wRainFall {
  0% { opacity: 0.45; transform: translateY(-1px); }
  50% { opacity: 1; transform: translateY(2px); }
  100% { opacity: 0.45; transform: translateY(-1px); }
}

@media (max-width: 767px) {
  .cw-mini-chip {
    display: none;
  }
}
</style>
