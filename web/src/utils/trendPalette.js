/**
 * Trend Analysis & SCADA Industrial Palette Engine
 * 
 * Provides:
 * 1. Exact canonical per-parameter color mappings for Dark Mode & Light Mode
 * 2. Deterministic hashing for unmapped/custom sensors so colors stay stable
 * 3. Theme tokens for OCC Dashboard (Industrial Navy Dark & Soft Sky Blue Light)
 * 4. High-contrast stroke colors with 100% opacity, clean grid and axis tones
 */

// 1. CANONICAL PARAMETER COLOR SPECIFICATION
export const CANONICAL_SERIES_COLORS = {
  pm25: {
    dark: '#27D9F2',
    light: '#0284C7',
    label: 'PM2.5',
    category: 'particulate',
  },
  pm10: {
    dark: '#5794FF',
    light: '#2563EB',
    label: 'PM10',
    category: 'particulate',
  },
  o3: {
    dark: '#B493FF',
    light: '#7C3AED',
    label: 'O₃',
    category: 'gas',
  },
  so2: {
    dark: '#36D6CE',
    light: '#0D9488',
    label: 'SO₂',
    category: 'gas',
  },
  no2: {
    dark: '#FF83B5',
    light: '#E11D48',
    label: 'NO₂',
    category: 'gas',
  },
  co: {
    dark: '#43D9A3',
    light: '#059669',
    label: 'CO',
    category: 'gas',
  },
  hc: {
    dark: '#D4E66D',
    light: '#65A30D',
    label: 'HC',
    category: 'gas',
  },
  temperature: {
    dark: '#FFBB66',
    light: '#D97706',
    label: 'Temperature',
    category: 'weather',
  },
  humidity: {
    dark: '#5CA9FF',
    light: '#1D4ED8',
    label: 'Humidity',
    category: 'weather',
  },
  pressure: {
    dark: '#A8B8CC',
    light: '#475569',
    label: 'Air Pressure',
    category: 'weather',
  },
  wind_speed: {
    dark: '#50C7E8',
    light: '#0891B2',
    label: 'Wind Speed',
    category: 'weather',
  },
  wind_angle: {
    dark: '#C1A0FF',
    light: '#6D28D9',
    label: 'Wind Angle',
    category: 'weather',
  },
  solar_radiation: {
    dark: '#FFD166',
    light: '#B45309',
    label: 'Solar Radiation',
    category: 'weather',
  },
  rain_fall: {
    dark: '#7DC8FF',
    light: '#4338CA',
    label: 'Rain Fall',
    category: 'weather',
  },
};

// 2. CURATED SCADA GENERAL PALETTE (for custom / additional parameters)
export const SCADA_FALLBACK_PALETTE = {
  dark: [
    '#27D9F2', '#5794FF', '#43D9A3', '#B493FF', '#50C7E8', '#FFBB66',
    '#36D6CE', '#FF83B5', '#D4E66D', '#5CA9FF', '#C1A0FF', '#7DC8FF',
    '#00E5FF', '#3B82F6', '#10B981', '#8B5CF6', '#F59E0B', '#EC4899',
  ],
  light: [
    '#0284C7', '#2563EB', '#059669', '#7C3AED', '#0891B2', '#D97706',
    '#0D9488', '#E11D48', '#65A30D', '#1D4ED8', '#6D28D9', '#4338CA',
    '#0369A1', '#1E40AF', '#047857', '#6D28D9', '#B45309', '#BE185D',
  ],
};

// 3. THEME TOKENS FOR OPERATOR DASHBOARD
export const DASHBOARD_THEME_TOKENS = {
  dark: {
    bg: '#061321', // legacy reference: bg: '#081421'
    panel: '#0B1C2B', // legacy reference: panel: '#0D1B2A'
    card: '#0D2232', // legacy reference: card: '#112235'
    border: '#244356', // legacy reference: border: '#294158'
    textPrimary: '#E6F3FC', // legacy reference: textPrimary: '#E2EBF4'
    textSecondary: '#9BB4C8', // legacy reference: textSecondary: '#91A6BA'
    accent: '#21C7E8', // legacy reference: accent: '#35C8DA'
    normal: '#36C99A',
    warning: '#F1BD61',
    alarm: '#FF6577',
    grid: '#132738',
    axis: '#244356',
  },
  light: {
    bg: 'linear-gradient(180deg, #E6EFF7 0%, #EDF3F9 12%, #F4F7FB 40%, #F8FAFD 100%)',
    bgTop: '#E6EFF7',
    bgCenter: '#F4F7FB',
    bgBottom: '#F8FAFD',
    panel: '#FFFFFF',
    card: '#FFFFFF',
    border: '#CBD5E1',
    textPrimary: '#0F172A',
    textSecondary: '#475569',
    accent: '#0284C7',
    normal: '#059669',
    warning: '#D97706',
    alarm: '#E11D48',
    grid: '#E2E8F0',
    axis: '#CBD5E1',
  },
};

/**
 * Resolves a parameter code/name to a canonical key in CANONICAL_SERIES_COLORS
 */
export function resolveCanonicalParamKey(code = '', name = '', unit = '') {
  const c = String(code || '').trim().toLowerCase();
  const n = String(name || '').trim().toLowerCase();
  const u = String(unit || '').trim().toLowerCase();

  // Particulates
  if (c.includes('pm2.5') || c.includes('pm25') || c === 'pm2_5' || n.includes('pm2.5') || n.includes('pm 2.5') || n.includes('particulate matter 2.5')) {
    return 'pm25';
  }
  if (c.includes('pm10') || n.includes('pm10') || n.includes('pm 10') || n.includes('particulate matter 10')) {
    return 'pm10';
  }

  // Gases
  if (c === 'o3' || c.startsWith('o3_') || n.includes('ozone') || n.includes('o3') || n.includes('o₃')) {
    return 'o3';
  }
  if (c === 'so2' || c.startsWith('so2_') || n.includes('sulfur') || n.includes('so2') || n.includes('so₂')) {
    return 'so2';
  }
  if (c === 'no2' || c.startsWith('no2_') || n.includes('nitrogen') || n.includes('no2') || n.includes('no₂')) {
    return 'no2';
  }
  if ((c === 'co' || c.startsWith('co_')) && !c.includes('co2') && !n.includes('dioxide')) {
    return 'co';
  }
  if (c.includes('hc') || c.includes('nmhc') || c.includes('thc') || c.includes('ch4') || n.includes('hydrocarbon') || n.includes('methane')) {
    return 'hc';
  }

  // Meteorology / Weather
  if (c.includes('temp') || c.includes('suhu') || n.includes('temperature') || n.includes('suhu') || u === '°c' || u === 'c') {
    return 'temperature';
  }
  if (c.includes('hum') || c.includes('rh') || c.includes('lembap') || n.includes('humidity') || n.includes('kelembaban') || u === '%rh') {
    return 'humidity';
  }
  if (c.includes('press') || c.includes('baro') || c.includes('tekanan') || n.includes('pressure') || n.includes('tekanan') || u === 'hpa' || u === 'mbar') {
    return 'pressure';
  }
  if (c.includes('ws') || c.includes('wind_speed') || c.includes('kecepatan') || n.includes('wind speed') || n.includes('kecepatan angin')) {
    return 'wind_speed';
  }
  if (c.includes('wd') || c.includes('wind_dir') || c.includes('wind_angle') || c.includes('arah') || n.includes('wind direction') || n.includes('arah angin')) {
    return 'wind_angle';
  }
  if (c.includes('solar') || c.includes('rad') || c.includes('uv') || n.includes('solar') || n.includes('radiasi') || u === 'w/m2') {
    return 'solar_radiation';
  }
  if (c.includes('rain') || c.includes('hujan') || c.includes('precip') || n.includes('rainfall') || n.includes('curah hujan') || u === 'mm') {
    return 'rain_fall';
  }

  return null;
}

/**
 * Returns deterministic high-contrast series color for a given parameter
 * @param {Object} param - { param_code, param_name, param_id, unit }
 * @param {boolean} isDark - Dark mode active
 * @param {number} fallbackIndex - index in active series list
 */
export function getParamSeriesColor(param, isDark = true) {
  const fallbackIndex = arguments[2] || 0;
  if (!param) return isDark ? '#27D9F2' : '#0284C7';

  const key = resolveCanonicalParamKey(
    param.param_code || param.paramCode || param.code,
    param.param_name || param.paramName || param.name,
    param.unit
  );

  const mode = isDark ? 'dark' : 'light';

  if (key && CANONICAL_SERIES_COLORS[key]) {
    return CANONICAL_SERIES_COLORS[key][mode];
  }

  // Fallback: Deterministic selection based on parameter ID or code string hash
  const rawId = String(param.param_id || param.parameterId || param.id || param.param_code || param.paramCode || fallbackIndex);
  let hash = 0;
  for (let i = 0; i < rawId.length; i++) {
    hash = ((hash << 5) - hash) + rawId.charCodeAt(i);
    hash |= 0;
  }
  const palette = SCADA_FALLBACK_PALETTE[mode];
  const idx = Math.abs(hash + fallbackIndex) % palette.length;
  return palette[idx];
}
