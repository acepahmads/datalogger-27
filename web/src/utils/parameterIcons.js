/**
 * Parameter Icons System for Datalogger Analysis Application
 * 
 * Provides:
 * 1. Safe predefined vector icon catalog
 * 2. Automatic mapping based on parameter code, name, and unit
 * 3. Graceful fallback for unknown parameters
 * 4. Distinct subtle visual accents (separated from alarm severity colors)
 */

export const PARAMETER_ICON_CATALOG = [
  {
    key: 'particles',
    label: 'Particles / Aerosol (PM2.5, PM10, Dust)',
    colorClass: 'text-sky-500 dark:text-sky-400',
    bgClass: 'bg-sky-500/10 dark:bg-sky-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><circle cx="6" cy="8" r="1.8" fill="currentColor" stroke="none"/><circle cx="11" cy="5" r="1.2" fill="currentColor" stroke="none"/><circle cx="17" cy="7" r="2.2" fill="currentColor" stroke="none"/><circle cx="6" cy="16" r="2" fill="currentColor" stroke="none"/><circle cx="13" cy="13" r="1.8" fill="currentColor" stroke="none"/><circle cx="19" cy="14" r="1.2" fill="currentColor" stroke="none"/><circle cx="12" cy="19" r="1.5" fill="currentColor" stroke="none"/><circle cx="18" cy="19" r="1.8" fill="currentColor" stroke="none"/></svg>`,
  },
  {
    key: 'atom',
    label: 'Gas / Molecule (O₃, CO, CO₂, NO₂, SO₂)',
    colorClass: 'text-purple-500 dark:text-purple-400',
    bgClass: 'bg-purple-500/10 dark:bg-purple-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><circle cx="12" cy="12" r="2.2" fill="currentColor"/><ellipse cx="12" cy="12" rx="9" ry="3.5" transform="rotate(45 12 12)" stroke="currentColor" stroke-width="1.8"/><ellipse cx="12" cy="12" rx="9" ry="3.5" transform="rotate(-45 12 12)" stroke="currentColor" stroke-width="1.8"/></svg>`,
  },
  {
    key: 'thermometer',
    label: 'Temperature (Temp, Ambient, Suhu)',
    colorClass: 'text-amber-600 dark:text-amber-400',
    bgClass: 'bg-amber-500/10 dark:bg-amber-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><path d="M14 14.76V3.5a2.5 2.5 0 0 0-5 0v11.26a4.5 4.5 0 1 0 5 0z"/><circle cx="12" cy="17" r="1.8" fill="currentColor" stroke="none"/></svg>`,
  },
  {
    key: 'droplet',
    label: 'Humidity / Moisture (RH, Water, Kelembaban)',
    colorClass: 'text-blue-500 dark:text-blue-400',
    bgClass: 'bg-blue-500/10 dark:bg-blue-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><path d="M12 2.69l5.66 5.66a8 8 0 1 1-11.31 0z"/></svg>`,
  },
  {
    key: 'gauge',
    label: 'Pressure (Air Pressure, Baro, Tekanan)',
    colorClass: 'text-teal-600 dark:text-teal-400',
    bgClass: 'bg-teal-500/10 dark:bg-teal-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><path d="M12 14l3-3"/><path d="M3.34 17a10 10 0 1 1 17.32 0"/><circle cx="12" cy="14" r="1.5" fill="currentColor"/></svg>`,
  },
  {
    key: 'wind',
    label: 'Wind Speed (Velocity, Kecepatan Angin)',
    colorClass: 'text-cyan-600 dark:text-cyan-400',
    bgClass: 'bg-cyan-500/10 dark:bg-cyan-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><path d="M17.7 7.7A2.5 2.5 0 1 1 19.5 12H2"/><path d="M12.7 3.7A2 2 0 1 1 14 7H2"/><path d="M14.7 16.3A2 2 0 1 0 16 20H2"/></svg>`,
  },
  {
    key: 'compass',
    label: 'Wind Direction (Compass, Arah Angin)',
    colorClass: 'text-emerald-600 dark:text-emerald-400',
    bgClass: 'bg-emerald-500/10 dark:bg-emerald-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><circle cx="12" cy="12" r="10"/><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76" stroke-width="1.8" fill="currentColor" fill-opacity="0.25"/></svg>`,
  },
  {
    key: 'cloud-rain',
    label: 'Rainfall / Precipitation (Curah Hujan)',
    colorClass: 'text-indigo-500 dark:text-indigo-400',
    bgClass: 'bg-indigo-500/10 dark:bg-indigo-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><path d="M17.5 19H9a7 7 0 1 1 6.71-9h1.79a4.5 4.5 0 1 1 0 9Z"/><path d="M8 19v3M12 19v3M16 19v3"/></svg>`,
  },
  {
    key: 'sun',
    label: 'Solar Radiation / UV (Cahaya, Radiasi)',
    colorClass: 'text-amber-500 dark:text-amber-300',
    bgClass: 'bg-amber-500/10 dark:bg-amber-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41"/></svg>`,
  },
  {
    key: 'waves',
    label: 'Flow Rate / Water Level (Debit, Level)',
    colorClass: 'text-cyan-500 dark:text-cyan-300',
    bgClass: 'bg-cyan-500/10 dark:bg-cyan-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><path d="M2 6c.6.5 1.2 1 2.5 1C7 7 7 5 9.5 5c2.6 0 2.4 2 5 2 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1M2 12c.6.5 1.2 1 2.5 1 2.5 0 2.5-2 5-2 2.6 0 2.4 2 5 2 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1M2 18c.6.5 1.2 1 2.5 1 2.5 0 2.5-2 5-2 2.6 0 2.4 2 5 2 2.5 0 2.5-2 5-2 1.3 0 1.9.5 2.5 1"/></svg>`,
  },
  {
    key: 'activity',
    label: 'Generic Sensor / Activity (Signal, Volt, Amps, Default)',
    colorClass: 'text-slate-500 dark:text-slate-400',
    bgClass: 'bg-slate-500/10 dark:bg-slate-500/20',
    svg: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="w-full h-full"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>`,
  },
];

const CATALOG_MAP = PARAMETER_ICON_CATALOG.reduce((acc, item) => {
  acc[item.key] = item;
  return acc;
}, {});

/**
 * Resolves the appropriate icon key based on:
 * 1. Explicitly configured parameter icon
 * 2. Automatic mapping based on parameter code, name, and unit
 * 3. Generic fallback icon ('activity')
 * 
 * @param {string} configuredIcon - Explicit icon key from database
 * @param {string} code - Parameter code (e.g. 'PM25', 'TEMP_01')
 * @param {string} name - Parameter name (e.g. 'Particulate Matter 2.5')
 * @param {string} unit - Parameter unit (e.g. 'µg/m³', '°C')
 * @returns {string} Safe icon key
 */
export function resolveParameterIconKey(configuredIcon, code = '', name = '', unit = '') {
  // 1. Explicitly configured parameter icon
  if (configuredIcon && CATALOG_MAP[configuredIcon]) {
    return configuredIcon;
  }

  // Combine search text for heuristic pattern matching
  const text = `${code} ${name} ${unit}`.toUpperCase();

  // 2. Automatic mapping based on sensor type patterns
  // Wind Direction must precede general Wind Speed
  if (/(?:WIND.*DIR|DIR.*WIND|ARAH.*ANGIN|\bWD\b|DEGREE|COMPASS)/i.test(text)) {
    return 'compass';
  }
  // Wind Speed
  if (/(?:WIND.*SPEED|SPEED.*WIND|KECEPATAN.*ANGIN|\bWS\b|KNOT|M\/S|ANEMO)/i.test(text)) {
    return 'wind';
  }
  // Particulate matter
  if (/(?:PM(?:2\.?5|10|1\.?0)|PARTICLE|DUST|AEROSOL|TSP|DEBU)/i.test(text)) {
    return 'particles';
  }
  // Temperature
  if (/(?:TEMP|SUHU|CELSIUS|FAHRENHEIT|°C|°F)/i.test(text)) {
    return 'thermometer';
  }
  // Humidity / Moisture
  if (/(?:HUM|RH|LEMBAP|KELEMBABAN|MOIST)/i.test(text)) {
    return 'droplet';
  }
  // Pressure
  if (/(?:PRESS|BARO|TEKANAN|HPA|MBAR|KPA|PSI|ATM)/i.test(text)) {
    return 'gauge';
  }
  // Rainfall
  if (/(?:RAIN|HUJAN|PRECIP|CURAH)/i.test(text)) {
    return 'cloud-rain';
  }
  // Solar radiation / UV
  if (/(?:SOLAR|SUN|RADIAS|RAD|UV|LUX|ILLUM|W\/M2|WM2)/i.test(text)) {
    return 'sun';
  }
  // Gas / Molecules
  if (/(?:O3|OZONE|CO2|CO|NO2|NOX|SO2|H2S|NH3|VOC|CH4|GAS|CHEMICAL)/i.test(text)) {
    return 'atom';
  }
  // Flow / Water Level
  if (/(?:FLOW|DEBIT|LEVEL|WATER|AIR_FLOW|WAVE|GPM|L\/MIN)/i.test(text)) {
    return 'waves';
  }

  // 3. Generic fallback
  return 'activity';
}

/**
 * Gets the icon definition object by key
 * @param {string} key 
 * @returns {object} Icon object containing key, label, svg, colorClass
 */
export function getParameterIconDefinition(key) {
  return CATALOG_MAP[key] || CATALOG_MAP['activity'];
}

/**
 * Resolves full icon definition from parameter attributes
 */
export function resolveParameterIconDef(configuredIcon, code, name, unit) {
  const key = resolveParameterIconKey(configuredIcon, code, name, unit);
  return getParameterIconDefinition(key);
}
