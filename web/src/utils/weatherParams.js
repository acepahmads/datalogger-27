/**
 * Weather & Meteorology parameter classifier and atmospheric state helpers
 * for Dashboard Operator's BentoWeatherHeroCard.
 */

const EXCLUDE_NON_AMBIENT_RE =
  /(^|[\s_\-.])(cpu|gpu|soc|board|pcb|box|panel|cabinet|shelter|enclosure|rack|indoor|internal|pump|motor|oil|water|wastewater|air[_\s-]?limbah|effluent|inlet|outlet|diff|differential|filter|sample|vacuum|exhaust|battery|charger|inverter)([\s_\-.]|$)/i;

const WEATHER_CATEGORY_RE = /(weather|meteorolog|cuaca|meteo|climate|iklim|aws)/i;

function normalizeText(param) {
  const code = String((param && (param.param_code || param.code)) || '').trim();
  const name = String((param && (param.param_name || param.name)) || '').trim();
  const unit = String((param && param.unit) || '').trim();
  return {
    code,
    name,
    unit,
    codeLower: code.toLowerCase(),
    nameLower: name.toLowerCase(),
    combined: `${code} ${name}`.toLowerCase(),
  };
}

function isWeatherCategoryParam(param, rawParamMap, categoriesById) {
  if (!param) return false;
  const catObj = param.category;
  if (catObj && WEATHER_CATEGORY_RE.test(`${catObj.name || ''} ${catObj.code || ''}`)) {
    return true;
  }
  const pid = param.param_id != null ? param.param_id : param.id;
  const raw = rawParamMap && pid != null ? rawParamMap[pid] : null;
  const catId =
    param.category_id != null && param.category_id !== ''
      ? Number(param.category_id)
      : raw && raw.category_id != null && raw.category_id !== ''
        ? Number(raw.category_id)
        : null;
  if (catId != null && categoriesById && categoriesById[catId]) {
    const c = categoriesById[catId];
    if (WEATHER_CATEGORY_RE.test(`${c.name || ''} ${c.code || ''}`)) {
      return true;
    }
  }
  return false;
}

function isOnlineGoodParam(param) {
  if (!param) return false;
  const q = String(param.quality || '').toUpperCase();
  const st = String(param.device_status || '').toUpperCase();
  return st !== 'OFFLINE' && q !== 'BAD';
}

function scoreSlot(param, slot, inWeatherCat) {
  const { codeLower, nameLower, combined, unit } = normalizeText(param);
  if (!combined.trim()) return -1;

  // Reject clearly non-ambient internal/water/shelter parameters
  if (EXCLUDE_NON_AMBIENT_RE.test(combined)) {
    return -1;
  }

  const unitLower = unit.toLowerCase();
  let score = -1;

  switch (slot) {
    case 'temperature': {
      const exactCode = /^(temp|temperature|t_amb|tamb|air_temp|ambient_temp|at|suhu|suhu_udara)$/.test(codeLower);
      const tokenMatch =
        /(^|[\s_\-.])(temp|temperature|suhu|air[_\s-]?temp|ambient[_\s-]?temp)([\s_\-.]|$)/.test(combined);
      if (exactCode) score = 100;
      else if (tokenMatch) score = 80;
      if (score > 0 && /(°c|degc|celcius|celsius|\bc\b|°f)/.test(unitLower)) score += 15;
      break;
    }
    case 'humidity': {
      const exactCode = /^(rh|hum|humidity|rel_hum|air_hum|kelembaban|kelembapan)$/.test(codeLower);
      const tokenMatch =
        /(^|[\s_\-.])(rh|hum|humidity|kelembaban|kelembapan|rel[_\s-]?hum)([\s_\-.]|$)/.test(combined);
      if (exactCode) score = 100;
      else if (tokenMatch) score = 80;
      if (score > 0 && /(%|rh)/.test(unitLower)) score += 15;
      break;
    }
    case 'windSpeed': {
      const exactCode = /^(ws|wspeed|w_speed|wind_speed|windspeed|wind_spd|kec_angin|kecepatan_angin)$/.test(codeLower);
      const tokenMatch =
        /(^|[\s_\-.])(wspeed|w[_\s-]?speed|wind[_\s-]?speed|windspeed|wind[_\s-]?spd|kecepatan[_\s-]?angin|kec[_\s-]?angin)([\s_\-.]|$)/.test(combined);
      if (exactCode) score = 100;
      else if (tokenMatch) score = 80;
      if (score > 0 && /(m\/s|km\/h|knot|kt|mph)/.test(unitLower)) score += 15;
      break;
    }
    case 'windAngle': {
      const exactCode = /^(wd|wdir|w_dir|wind_dir|winddir|wind_angle|windangle|w_angle|arah_angin|sudut_angin)$/.test(codeLower);
      const tokenMatch =
        /(^|[\s_\-.])(wdir|w[_\s-]?dir|wind[_\s-]?dir|winddir|wind[_\s-]?angle|windangle|w[_\s-]?angle|arah[_\s-]?angin|sudut[_\s-]?angin)([\s_\-.]|$)/.test(combined);
      if (exactCode) score = 100;
      else if (tokenMatch) score = 80;
      if (score > 0 && /(°|deg)/.test(unitLower)) score += 15;
      break;
    }
    case 'solarRadiation': {
      const exactCode = /^(sr|solar|solar_rad|solar_radiation|solarrad|radiasi|irradiance|intensity|pyrano|pyranometer)$/.test(codeLower);
      const tokenMatch =
        /(^|[\s_\-.])(solar[_\s-]?rad|solar[_\s-]?radiation|solar|radiasi|irradiance|pyranometer|intensity)([\s_\-.]|$)/.test(combined);
      if (exactCode) score = 100;
      else if (tokenMatch) score = 80;
      if (score > 0 && /(w\/m|watt|lux)/.test(unitLower)) score += 15;
      break;
    }
    case 'pressure': {
      const exactCode = /^(bp|baro|barometer|barometric|pressure|air_pressure|air_press|tekanan|tekanan_udara|atm)$/.test(codeLower);
      const tokenMatch =
        /(^|[\s_\-.])(baro|barometer|barometric|air[_\s-]?press|air[_\s-]?pressure|pressure|tekanan[_\s-]?udara|tekanan)([\s_\-.]|$)/.test(combined);
      if (exactCode) score = 100;
      else if (tokenMatch) score = 80;
      if (score > 0 && /(hpa|mbar|mb|kpa|pa|mmhg|atm)/.test(unitLower)) score += 15;
      break;
    }
    case 'rainFall': {
      const exactCode = /^(rain|rainfall|rain_fall|curah_hujan|hujan|precip|precipitation|tprecipitation)$/.test(codeLower);
      const tokenMatch =
        /(^|[\s_\-.])(rain[_\s-]?fall|rainfall|rain|curah[_\s-]?hujan|hujan|precip|precipitation|tprecipitation)([\s_\-.]|$)/.test(combined);
      if (exactCode) score = 100;
      else if (tokenMatch) score = 80;
      if (score > 0 && /(mm|inch|in\/h)/.test(unitLower)) score += 15;
      break;
    }
    default:
      return -1;
  }

  if (score < 0) return -1;
  if (inWeatherCat) score += 25;
  if (isOnlineGoodParam(param)) score += 5;
  if (nameLower.includes('ambient') || nameLower.includes('udara') || nameLower.includes('air')) {
    score += 4;
  }
  return score;
}

export function classifyWeatherParameters(allParameters = [], rawParamMap = null, categoriesList = []) {
  const list = Array.isArray(allParameters) ? allParameters : [];
  const categoriesById = {};
  if (Array.isArray(categoriesList)) {
    for (const c of categoriesList) {
      if (c && c.id != null) categoriesById[Number(c.id)] = c;
    }
  }

  const slots = [
    'temperature',
    'humidity',
    'windSpeed',
    'windAngle',
    'solarRadiation',
    'pressure',
    'rainFall',
  ];

  const picked = {
    temperature: null,
    humidity: null,
    windSpeed: null,
    windAngle: null,
    solarRadiation: null,
    pressure: null,
    rainFall: null,
  };
  const usedIds = new Set();

  for (const slot of slots) {
    let bestParam = null;
    let bestScore = -1;
    for (const p of list) {
      if (!p) continue;
      const pid = p.param_id != null ? p.param_id : p.id;
      if (pid != null && usedIds.has(pid)) continue;
      const inWeatherCat = isWeatherCategoryParam(p, rawParamMap, categoriesById);
      const sc = scoreSlot(p, slot, inWeatherCat);
      if (sc > bestScore) {
        bestScore = sc;
        bestParam = p;
      }
    }
    if (bestParam && bestScore > 0) {
      picked[slot] = bestParam;
      const pid = bestParam.param_id != null ? bestParam.param_id : bestParam.id;
      if (pid != null) usedIds.add(pid);
    }
  }

  const matchedParams = slots.map(s => picked[s]).filter(Boolean);
  const matchedIds = new Set(
    matchedParams
      .map(p => (p.param_id != null ? p.param_id : p.id))
      .filter(id => id != null)
  );

  return {
    ...picked,
    matchedParams,
    matchedIds,
    hasPrimaryClimate: Boolean(picked.temperature || picked.humidity),
    hasAnyWeather: matchedParams.length > 0,
    count: matchedParams.length,
  };
}

export function parseNumericValue(param) {
  if (!param || param.value == null || param.value === '') return null;
  const n = Number(param.value);
  return Number.isFinite(n) ? n : null;
}

const COMPASS_POINTS = [
  { short: 'N', idShort: 'U', label: 'Utara' },
  { short: 'NE', idShort: 'TL', label: 'Timur Laut' },
  { short: 'E', idShort: 'T', label: 'Timur' },
  { short: 'SE', idShort: 'TG', label: 'Tenggara' },
  { short: 'S', idShort: 'S', label: 'Selatan' },
  { short: 'SW', idShort: 'BD', label: 'Barat Daya' },
  { short: 'W', idShort: 'B', label: 'Barat' },
  { short: 'NW', idShort: 'BL', label: 'Barat Laut' },
];

export function getCardinalDirection(degrees) {
  const num = Number(degrees);
  if (!Number.isFinite(num)) {
    return { deg: 0, short: '-', idShort: '-', label: '—', valid: false };
  }
  const norm = ((num % 360) + 360) % 360;
  const idx = Math.round(norm / 45) % 8;
  const pt = COMPASS_POINTS[idx];
  return {
    deg: Math.round(norm),
    short: pt.short,
    idShort: pt.idShort,
    label: pt.label,
    valid: true,
  };
}

export function inferWeatherCondition({
  temperature = null,
  humidity = null,
  solarRadiation = null,
  rainFall = null,
  nowDate = null,
} = {}) {
  const dt = nowDate instanceof Date ? nowDate : new Date();
  const hour = dt.getHours();
  const isNightTime = hour >= 18 || hour < 6;

  const rain = Number.isFinite(Number(rainFall)) && rainFall !== null ? Number(rainFall) : null;
  const solar = Number.isFinite(Number(solarRadiation)) && solarRadiation !== null ? Number(solarRadiation) : null;
  const rh = Number.isFinite(Number(humidity)) && humidity !== null ? Number(humidity) : null;
  const temp = Number.isFinite(Number(temperature)) && temperature !== null ? Number(temperature) : null;

  if (rain !== null && rain > 0.05) {
    const heavy = rain >= 5;
    if (isNightTime) {
      return {
        key: 'night_rain',
        label: heavy ? 'Malam Hujan Deras' : 'Malam Hujan',
        isNightTime: true,
        themeTone: 'rain',
      };
    }
    return {
      key: heavy ? 'heavy_rain' : 'rain',
      label: heavy ? 'Hujan Deras' : 'Hujan Ringan',
      isNightTime: false,
      themeTone: 'rain',
    };
  }

  if (isNightTime) {
    if (rh !== null && rh >= 82) {
      return {
        key: 'night_cloudy',
        label: 'Malam Berawan',
        isNightTime: true,
        themeTone: 'night',
      };
    }
    return {
      key: 'night_clear',
      label: 'Malam Cerah',
      isNightTime: true,
      themeTone: 'night',
    };
  }

  if (solar !== null) {
    if (solar >= 450) {
      return {
        key: 'sunny',
        label: 'Cerah Terik',
        isNightTime: false,
        themeTone: 'sunny',
      };
    }
    if (solar >= 140) {
      return {
        key: 'partly_cloudy',
        label: 'Cerah Berawan',
        isNightTime: false,
        themeTone: 'partly',
      };
    }
    return {
      key: 'cloudy',
      label: 'Berawan',
      isNightTime: false,
      themeTone: 'cloudy',
    };
  }

  if (rh !== null && rh >= 85) {
    return {
      key: 'cloudy',
      label: 'Lembap Berawan',
      isNightTime: false,
      themeTone: 'cloudy',
    };
  }

  if (temp !== null && temp >= 30) {
    return {
      key: 'sunny',
      label: 'Cerah Hangat',
      isNightTime: false,
      themeTone: 'sunny',
    };
  }

  return {
    key: 'partly_cloudy',
    label: 'Cerah Berawan',
    isNightTime: false,
    themeTone: 'partly',
  };
}

export function getTempComfortMeta(tempVal) {
  const t = Number(tempVal);
  if (!Number.isFinite(t)) return { label: '—', tone: 'slate', pct: 0 };
  const pct = Math.max(4, Math.min(100, Math.round((t / 50) * 100)));
  if (t < 20) return { label: 'Dingin', tone: 'sky', pct };
  if (t < 27) return { label: 'Sejuk', tone: 'emerald', pct };
  if (t < 32) return { label: 'Hangat', tone: 'amber', pct };
  if (t < 36) return { label: 'Panas', tone: 'orange', pct };
  return { label: 'Terik', tone: 'rose', pct };
}

export function getHumidityComfortMeta(rhVal) {
  const rh = Number(rhVal);
  if (!Number.isFinite(rh)) return { label: '—', tone: 'slate', pct: 0 };
  const pct = Math.max(0, Math.min(100, Math.round(rh)));
  if (rh < 35) return { label: 'Kering', tone: 'amber', pct };
  if (rh <= 65) return { label: 'Ideal', tone: 'emerald', pct };
  if (rh <= 80) return { label: 'Lembap', tone: 'sky', pct };
  return { label: 'Tinggi', tone: 'indigo', pct };
}

export function getWindBeaufortLabel(wsVal) {
  const ws = Number(wsVal);
  if (!Number.isFinite(ws)) return '—';
  if (ws < 0.5) return 'Tenang';
  if (ws < 2.5) return 'Sepoi Lembut';
  if (ws < 5.5) return 'Sepoi Sedang';
  if (ws < 10.8) return 'Angin Segar';
  return 'Angin Kencang';
}
