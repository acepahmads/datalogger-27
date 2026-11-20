/**
 * Inbound Telemetry Payload Validator (Format, Values & Completeness)
 * Supports:
 * - KLHK_SPARING 30x2-minute hourly bundles ({"uid": "...", "data": [{datetime, pH, cod, tss, nh3n, debit}, ...]})
 * - CBI_LEGACY SPARING hourly bundles ({"module_id": "...", "station_id": "...", "hourly_data": [...]})
 * - CBI_LEGACY multi-tick ({"multi": [{"tick": {module_id, station_id, sensor_name, value, created_at}}, ...]})
 * - STANDARD_V1 batch ({"records": [...]}) & single record ({"site_code", "device_code", "timestamp", "values"})
 * - MQTT webhook envelope ({"topic": "...", "payload": {...}})
 */

export const EXPECTED_SPARING_MINUTES = Array.from({ length: 30 }, (_, i) => i * 2);

const RESERVED_KEYS = new Set([
  'uid',
  'site_code',
  'station_code',
  'station_id',
  'id_stasiun',
  'idstasiun',
  'device_code',
  'device_name',
  'module_id',
  'code',
  'name',
  'logger_type',
  'domain',
  'monitoring_domain',
  'timestamp',
  'datetime',
  'waktu',
  'created_at',
  'updated_at',
  'period_start',
  'period_end',
  'tanggal',
  'jam',
  'interval_minutes',
  'interval',
  'quality',
  'status',
  'topic',
  'metadata',
  'source',
  'tenant_id',
  'site_id',
  'device_id',
]);

const SPARING_CORE_PARAM_RE = /^(ph|cod|tss|nh3n|nh3_n|amonia|ammonia|debit|flow)$/i;

const PHYSICAL_RULES = [
  { re: /^ph$/i, min: 0, max: 14, label: 'pH (0–14)' },
  { re: /^(rh|hum|humidity|kelembaban|kelembapan)$/i, min: 0, max: 100, label: 'Humidity (0–100%)' },
  { re: /^(temp|temperature|suhu|t_amb|tamb|air_temp)$/i, min: -40, max: 85, label: 'Temperature (-40–85°C)' },
  { re: /^(wd|wdir|wind_dir|wind_angle|arah_angin)$/i, min: 0, max: 360, label: 'Wind Angle (0–360°)' },
  { re: /^(ws|wspeed|wind_speed|kecepatan_angin)$/i, min: 0, max: 120, label: 'Wind Speed (0–120 m/s)' },
  { re: /^(bp|baro|pressure|air_pressure|tekanan)$/i, min: 300, max: 1200, label: 'Pressure (300–1200 hPa)' },
  { re: /^(cod|tss|nh3n|nh3_n|debit|flow|pm10|pm25|pm2_5|pm2\.5|so2|co|o3|no2|hc|nmhc|rain|rainfall|rain_fall|sr|solar|solar_radiation|intensity|do|tds|ec|turbidity|bod)$/i, min: 0, max: 1000000, label: 'Non-negatif (>= 0)' },
];

function pad2(n) {
  return String(n).padStart(2, '0');
}

function formatDateLocal(d) {
  if (!(d instanceof Date) || Number.isNaN(d.getTime())) return '-';
  return `${d.getFullYear()}-${pad2(d.getMonth() + 1)}-${pad2(d.getDate())} ${pad2(d.getHours())}:${pad2(d.getMinutes())}:${pad2(d.getSeconds())}`;
}

export function parseInboundTimestamp(raw) {
  if (raw === null || raw === undefined || raw === '') {
    return { valid: false, reason: 'Timestamp kosong' };
  }

  let date = null;
  if (typeof raw === 'number' && Number.isFinite(raw)) {
    if (raw <= 0) return { valid: false, reason: `Timestamp epoch tidak valid (${raw})` };
    const ms = raw > 1e12 ? raw : raw * 1000;
    date = new Date(ms);
  } else if (typeof raw === 'string') {
    const s = raw.trim();
    if (!s) return { valid: false, reason: 'Timestamp string kosong' };
    if (/^\d{10}$/.test(s)) {
      date = new Date(Number(s) * 1000);
    } else if (/^\d{13}$/.test(s)) {
      date = new Date(Number(s));
    } else {
      // Normalize legacy Go/SQL string: "2026-11-20 10:00:00 +0700" -> "2026-11-20T10:00:00+07:00"
      let norm = s;
      const legacyMatch = s.match(/^(\d{4}-\d{2}-\d{2})[ T](\d{2}:\d{2}:\d{2}(?:\.\d+)?)\s*([+-]\d{2}):?(\d{2})$/);
      if (legacyMatch) {
        norm = `${legacyMatch[1]}T${legacyMatch[2]}${legacyMatch[3]}:${legacyMatch[4]}`;
      } else if (/^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$/.test(s)) {
        norm = s.replace(' ', 'T');
      }
      date = new Date(norm);
    }
  } else {
    return { valid: false, reason: 'Tipe timestamp bukan string/angka' };
  }

  if (!date || Number.isNaN(date.getTime())) {
    return { valid: false, reason: `Format timestamp tidak dikenali (${String(raw)})` };
  }

  const year = date.getFullYear();
  if (year < 2000 || year > 2100) {
    return { valid: false, reason: `Tahun timestamp di luar batas wajar (${year})` };
  }

  const hourKey = `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(date.getDate())} ${pad2(date.getHours())}:00`;
  return {
    valid: true,
    date,
    ms: date.getTime(),
    minute: date.getMinutes(),
    second: date.getSeconds(),
    hourKey,
    formatted: formatDateLocal(date),
  };
}

function unwrapMQTTEnvelope(root) {
  if (!root || typeof root !== 'object' || Array.isArray(root)) return { unwrapped: root, wasMqtt: false };
  if (typeof root.topic === 'string' && root.topic.trim() && root.payload !== undefined) {
    if (root.payload && typeof root.payload === 'object' && !Array.isArray(root.payload)) {
      return { unwrapped: root.payload, wasMqtt: true };
    }
    if (typeof root.payload === 'string') {
      try {
        const parsed = JSON.parse(root.payload);
        if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
          return { unwrapped: parsed, wasMqtt: true };
        }
      } catch (_) {}
    }
  }
  return { unwrapped: root, wasMqtt: false };
}

function extractRawTimestamp(obj, fallback = null) {
  if (!obj || typeof obj !== 'object') return fallback;
  const keys = ['datetime', 'timestamp', 'waktu', 'created_at', 'period_start', 'period_end', 'updated_at'];
  for (const k of keys) {
    if (obj[k] !== undefined && obj[k] !== null && String(obj[k]).trim() !== '') {
      return obj[k];
    }
  }
  const tgl = obj.Tanggal || obj.tanggal;
  const jam = obj.Jam || obj.jam;
  if (tgl && jam) return `${String(tgl).trim()} ${String(jam).trim()}`;
  return fallback;
}

function isStatFlagKey(key) {
  return /^stat_/i.test(String(key || ''));
}

function checkPhysicalRange(code, numVal) {
  const cleanCode = String(code || '').trim();
  for (const rule of PHYSICAL_RULES) {
    if (rule.re.test(cleanCode)) {
      if (numVal < rule.min || numVal > rule.max) {
        return { ok: false, reason: `${cleanCode}=${numVal} di luar batas ${rule.label}` };
      }
      return { ok: true, reason: '' };
    }
  }
  return { ok: true, reason: '' };
}

function extractEntriesFromRecord(rec) {
  // Returns array of { code, rawValue, value, isNull, isInvalidType, outOfRange, rangeReason }
  const out = [];
  if (!rec || typeof rec !== 'object') return out;

  const inspectPair = (code, rawVal) => {
    const cleanCode = String(code || '').trim();
    if (!cleanCode || RESERVED_KEYS.has(cleanCode.toLowerCase()) || isStatFlagKey(cleanCode)) return;
    if (rawVal && typeof rawVal === 'object' && !Array.isArray(rawVal) && ('value' in rawVal)) {
      rawVal = rawVal.value;
    }
    if (rawVal === null || rawVal === undefined || rawVal === '') {
      out.push({ code: cleanCode, rawValue: rawVal, value: null, isNull: true, isInvalidType: false, outOfRange: false, rangeReason: '' });
      return;
    }
    const num = typeof rawVal === 'number' ? rawVal : Number(rawVal);
    if (typeof rawVal === 'boolean' || !Number.isFinite(num)) {
      out.push({ code: cleanCode, rawValue: rawVal, value: null, isNull: false, isInvalidType: true, outOfRange: false, rangeReason: `${cleanCode} bukan angka (${String(rawVal)})` });
      return;
    }
    const rng = checkPhysicalRange(cleanCode, num);
    out.push({
      code: cleanCode,
      rawValue: rawVal,
      value: num,
      isNull: false,
      isInvalidType: false,
      outOfRange: !rng.ok,
      rangeReason: rng.reason,
    });
  };

  let usedContainer = false;
  for (const containerKey of ['values', 'readings', 'parameters', 'data']) {
    const sub = rec[containerKey];
    if (sub && typeof sub === 'object') {
      usedContainer = true;
      if (Array.isArray(sub)) {
        for (const el of sub) {
          if (!el || typeof el !== 'object') continue;
          const code = el.code || el.parameter_code || el.sensor_name || el.name;
          if (code) inspectPair(code, el.value);
        }
      } else {
        for (const k of Object.keys(sub)) {
          inspectPair(k, sub[k]);
        }
      }
    }
  }

  if (!usedContainer) {
    for (const k of Object.keys(rec)) {
      const v = rec[k];
      if (v && typeof v === 'object' && !Array.isArray(v) && !('value' in v)) continue;
      inspectPair(k, v);
    }
  }

  return out;
}

function isCodeValueList(arr) {
  if (!Array.isArray(arr) || arr.length === 0) return false;
  const first = arr[0];
  return Boolean(first && typeof first === 'object' && (first.code || first.parameter_code));
}

export function validateInboundLog(item) {
  const issues = [];
  const formatChecks = [];
  const valueChecks = [];
  const completenessChecks = [];

  const log = item || {};
  const rawStr = String(log.raw_payload || '').trim();
  const httpAccepted = String(log.status || '').toUpperCase() === 'ACCEPTED';
  const loggerTypeUpper = String(log.logger_type || '').toUpperCase();
  const detectedFormatUpper = String(log.detected_format || '').toUpperCase();
  const skippedParams = Number(log.parameters_skipped || 0);

  // 1. Parse JSON
  let parsedRoot = null;
  let jsonParseErr = '';
  if (!rawStr) {
    jsonParseErr = 'Body payload kosong';
  } else {
    try {
      parsedRoot = JSON.parse(rawStr);
    } catch (e) {
      jsonParseErr = e && e.message ? e.message : 'JSON tidak valid';
    }
  }

  if (jsonParseErr || !parsedRoot || typeof parsedRoot !== 'object' || Array.isArray(parsedRoot)) {
    const msg = jsonParseErr || 'Root JSON harus berupa objek {...}';
    formatChecks.push({ ok: false, level: 'FAIL', label: 'Sintaks JSON', detail: msg });
    issues.push({ category: 'FORMAT', level: 'FAIL', message: `Format JSON rusak: ${msg}` });
    return {
      overallStatus: 'INVALID',
      score: 0,
      isSparing: loggerTypeUpper === 'SPARING' || detectedFormatUpper === 'KLHK_SPARING',
      schemaLabel: 'INVALID_JSON',
      format: { status: 'FAIL', label: 'JSON Rusak', summary: msg, checks: formatChecks },
      values: {
        status: 'FAIL',
        label: 'Tidak Terbaca',
        summary: 'Payload tidak dapat di-parse',
        totalValues: 0,
        validValues: 0,
        nullCount: 0,
        invalidTypeCount: 0,
        outOfRangeCount: 0,
        parameterStats: [],
        checks: [{ ok: false, level: 'FAIL', label: 'Pembacaan Nilai', detail: 'Tidak ada nilai terbaca karena JSON rusak' }],
      },
      completeness: {
        status: 'FAIL',
        label: 'Kosong',
        summary: 'Tidak ada record yang dapat diverifikasi',
        recordCount: 0,
        paramCount: 0,
        skippedParams,
        sparing: null,
        checks: [{ ok: false, level: 'FAIL', label: 'Kelengkapan Data', detail: 'Payload kosong atau rusak' }],
      },
      issues,
    };
  }

  formatChecks.push({ ok: true, level: 'PASS', label: 'Sintaks JSON', detail: 'Struktur JSON valid dan dapat di-parse' });

  const { unwrapped: root, wasMqtt } = unwrapMQTTEnvelope(parsedRoot);
  if (wasMqtt) {
    formatChecks.push({ ok: true, level: 'PASS', label: 'Envelope MQTT', detail: `Topic "${parsedRoot.topic}" berhasil dibuka` });
  }

  // Detect schema family & extract normalized records: [{ rawTS, entries, rawRow }]
  let schemaLabel = 'STANDARD_V1';
  let schemaKind = 'STANDARD_SINGLE';
  let records = [];
  const rootRawTS = extractRawTimestamp(root, null);

  if (Array.isArray(root.data) && !isCodeValueList(root.data)) {
    schemaLabel = 'KLHK_SPARING (data[])';
    schemaKind = 'SPARING_ARRAY';
    records = root.data.map(row => ({
      rawTS: extractRawTimestamp(row, rootRawTS),
      entries: extractEntriesFromRecord(row),
      rawRow: row,
    }));
  } else if (Array.isArray(root.hourly_data)) {
    schemaLabel = 'CBI_LEGACY (hourly_data[])';
    schemaKind = 'SPARING_ARRAY';
    records = root.hourly_data.map(row => ({
      rawTS: extractRawTimestamp(row, rootRawTS),
      entries: extractEntriesFromRecord(row),
      rawRow: row,
    }));
  } else if (Array.isArray(root.multi)) {
    schemaLabel = 'CBI_LEGACY (multi[])';
    schemaKind = 'CBI_MULTI';
    const byTs = new Map();
    for (const el of root.multi) {
      const tm = el && typeof el === 'object' && el.tick && typeof el.tick === 'object' ? el.tick : el;
      if (!tm || typeof tm !== 'object') continue;
      const code = String(tm.sensor_name || tm.parameter_code || tm.code || '').trim();
      const rawTS = extractRawTimestamp(tm, rootRawTS);
      const key = rawTS !== null && rawTS !== undefined ? String(rawTS) : '__no_ts__';
      if (!byTs.has(key)) {
        byTs.set(key, { rawTS, entries: [], rawRow: tm, seenCodes: new Set(), duplicateCodes: [] });
      }
      const bucket = byTs.get(key);
      if (code) {
        if (bucket.seenCodes.has(code.toLowerCase())) {
          bucket.duplicateCodes.push(code);
        }
        bucket.seenCodes.add(code.toLowerCase());
        const pseudo = {};
        pseudo[code] = tm.value;
        bucket.entries.push(...extractEntriesFromRecord(pseudo));
      }
    }
    records = Array.from(byTs.values());
  } else if (Array.isArray(root.records)) {
    schemaLabel = 'STANDARD_V1 (records[])';
    schemaKind = 'STANDARD_BATCH';
    records = root.records.map(row => ({
      rawTS: extractRawTimestamp(row, rootRawTS),
      entries: extractEntriesFromRecord(row),
      rawRow: row,
    }));
  } else {
    schemaLabel = 'STANDARD_V1 (Single)';
    schemaKind = 'STANDARD_SINGLE';
    records = [
      {
        rawTS: rootRawTS,
        entries: extractEntriesFromRecord(root),
        rawRow: root,
      },
    ];
  }

  // Identity field check
  const hasRootIdentity = Boolean(
    root.uid ||
    root.site_code ||
    root.station_code ||
    root.station_id ||
    root.module_id ||
    root.device_code ||
    log.site_code ||
    log.device_code ||
    (schemaKind === 'CBI_MULTI' && records.some(r => r.rawRow && (r.rawRow.station_id || r.rawRow.module_id)))
  );

  if (hasRootIdentity) {
    formatChecks.push({ ok: true, level: 'PASS', label: 'Identitas Stasiun/Device', detail: `Skema ${schemaLabel} memiliki identitas pengirim` });
  } else {
    formatChecks.push({ ok: false, level: 'WARN', label: 'Identitas Stasiun/Device', detail: 'Field uid/site_code/station_id tidak ditemukan pada root JSON' });
    issues.push({ category: 'FORMAT', level: 'WARN', message: 'Field identitas (uid / site_code / station_id) tidak eksplisit pada payload' });
  }

  // Ingestion acceptance check
  if (httpAccepted) {
    formatChecks.push({ ok: true, level: 'PASS', label: 'Status Receiver', detail: `HTTP ${log.http_status || 200} ACCEPTED` });
  } else {
    const errText = log.error_message || `HTTP ${log.http_status || 400} REJECTED`;
    formatChecks.push({ ok: false, level: 'FAIL', label: 'Status Receiver', detail: errText });
    issues.push({ category: 'FORMAT', level: 'FAIL', message: `Payload ditolak server: ${errText}` });
  }

  // Parse timestamps on every record
  let validTsCount = 0;
  let invalidTsCount = 0;
  const parsedRecords = records.map((rec, idx) => {
    const ts = parseInboundTimestamp(rec.rawTS);
    if (ts.valid) validTsCount += 1;
    else invalidTsCount += 1;
    return { ...rec, idx, ts };
  });

  if (records.length === 0) {
    formatChecks.push({ ok: false, level: 'FAIL', label: 'Array Record', detail: 'Payload tidak berisi record data' });
    issues.push({ category: 'FORMAT', level: 'FAIL', message: 'Array data/records kosong (0 record)' });
  } else if (invalidTsCount === 0) {
    formatChecks.push({ ok: true, level: 'PASS', label: 'Format Waktu (Timestamp)', detail: `${validTsCount}/${records.length} record memiliki timestamp valid` });
  } else {
    formatChecks.push({ ok: false, level: 'FAIL', label: 'Format Waktu (Timestamp)', detail: `${invalidTsCount} dari ${records.length} record tanpa/rusak timestamp` });
    issues.push({ category: 'FORMAT', level: 'FAIL', message: `${invalidTsCount} record memiliki timestamp kosong atau format tidak valid` });
  }

  // Determine whether this payload should be audited under SPARING 30x2-minute (00..58) rules
  const allParamCodes = new Set();
  for (const rec of parsedRecords) {
    for (const e of rec.entries) allParamCodes.add(e.code.toLowerCase());
  }
  let sparingCoreMatches = 0;
  for (const c of allParamCodes) {
    if (SPARING_CORE_PARAM_RE.test(c)) sparingCoreMatches += 1;
  }

  const isSparing =
    loggerTypeUpper === 'SPARING' ||
    detectedFormatUpper === 'KLHK_SPARING' ||
    schemaKind === 'SPARING_ARRAY' ||
    (records.length >= 20 && sparingCoreMatches >= 3);

  // 2. VALUE VALIDATION
  let totalValues = 0;
  let validValues = 0;
  let nullCount = 0;
  let invalidTypeCount = 0;
  let outOfRangeCount = 0;
  const rangeIssueSamples = [];
  const invalidTypeSamples = [];
  const paramStatMap = new Map();

  for (const rec of parsedRecords) {
    for (const e of rec.entries) {
      totalValues += 1;
      const key = e.code;
      if (!paramStatMap.has(key)) {
        paramStatMap.set(key, {
          code: key,
          count: 0,
          validCount: 0,
          nullCount: 0,
          invalidTypeCount: 0,
          outOfRangeCount: 0,
          min: null,
          max: null,
          lastValue: null,
          issueText: '',
        });
      }
      const st = paramStatMap.get(key);
      st.count += 1;
      if (e.isNull) {
        nullCount += 1;
        st.nullCount += 1;
      } else if (e.isInvalidType) {
        invalidTypeCount += 1;
        st.invalidTypeCount += 1;
        if (invalidTypeSamples.length < 4) invalidTypeSamples.push(e.rangeReason);
        if (!st.issueText) st.issueText = e.rangeReason;
      } else {
        st.lastValue = e.value;
        if (st.min === null || e.value < st.min) st.min = e.value;
        if (st.max === null || e.value > st.max) st.max = e.value;
        if (e.outOfRange) {
          outOfRangeCount += 1;
          st.outOfRangeCount += 1;
          if (rangeIssueSamples.length < 4) rangeIssueSamples.push(e.rangeReason);
          if (!st.issueText) st.issueText = e.rangeReason;
        } else {
          validValues += 1;
          st.validCount += 1;
        }
      }
    }
  }

  const parameterStats = Array.from(paramStatMap.values()).map(st => {
    let status = 'VALID';
    if (st.invalidTypeCount > 0 || st.outOfRangeCount > 0) status = 'INVALID';
    else if (st.nullCount > 0) status = 'WARNING';
    return { ...st, status };
  });

  if (totalValues === 0) {
    valueChecks.push({ ok: false, level: 'FAIL', label: 'Nilai Sensor', detail: 'Tidak ditemukan nilai sensor numerik pada payload' });
    issues.push({ category: 'NILAI', level: 'FAIL', message: 'Tidak ada nilai sensor pada payload' });
  } else {
    valueChecks.push({
      ok: invalidTypeCount === 0,
      level: invalidTypeCount === 0 ? 'PASS' : 'FAIL',
      label: 'Tipe Data Numerik',
      detail: invalidTypeCount === 0 ? `${totalValues - nullCount} nilai bertipe angka valid` : `${invalidTypeCount} nilai bukan angka: ${invalidTypeSamples.join(', ')}`,
    });
    if (invalidTypeCount > 0) {
      issues.push({ category: 'NILAI', level: 'FAIL', message: `${invalidTypeCount} nilai sensor bukan tipe numerik (${invalidTypeSamples.join(', ')})` });
    }

    valueChecks.push({
      ok: outOfRangeCount === 0,
      level: outOfRangeCount === 0 ? 'PASS' : 'FAIL',
      label: 'Kewajaran Rentang Fisik',
      detail: outOfRangeCount === 0 ? 'Seluruh nilai dalam rentang fisik wajar' : `${outOfRangeCount} nilai anomali: ${rangeIssueSamples.join(', ')}`,
    });
    if (outOfRangeCount > 0) {
      issues.push({ category: 'NILAI', level: 'FAIL', message: `${outOfRangeCount} nilai di luar rentang fisik wajar (${rangeIssueSamples.join(', ')})` });
    }

    if (nullCount > 0) {
      const allNull = validValues === 0;
      valueChecks.push({
        ok: false,
        level: allNull ? 'FAIL' : 'WARN',
        label: 'Nilai Kosong (null)',
        detail: `${nullCount} dari ${totalValues} field bernilai null/kosong`,
      });
      issues.push({
        category: 'NILAI',
        level: allNull ? 'FAIL' : 'WARN',
        message: `${nullCount} field parameter bernilai null/kosong (${Math.round((nullCount / totalValues) * 100)}%)`,
      });
    } else {
      valueChecks.push({
        ok: true,
        level: 'PASS',
        label: 'Nilai Kosong (null)',
        detail: 'Tidak ada field bernilai null',
      });
    }
  }

  // 3. COMPLETENESS VALIDATION (including SPARING 00..58 30-slot audit)
  let sparingAudit = null;

  if (isSparing) {
    const slotMap = new Map();
    const offGridMinutes = [];
    const hourSet = new Set();

    for (const rec of parsedRecords) {
      if (!rec.ts.valid) continue;
      hourSet.add(rec.ts.hourKey);
      const min = rec.ts.minute;
      if (min % 2 !== 0) {
        if (!offGridMinutes.includes(min)) offGridMinutes.push(min);
        continue;
      }
      if (!slotMap.has(min)) slotMap.set(min, []);
      slotMap.get(min).push(rec);
    }

    const missingMinutes = [];
    const duplicateMinutes = [];
    let presentSlots = 0;
    let completeSlots = 0;
    let partialSlots = 0;

    const slots = EXPECTED_SPARING_MINUTES.map(min => {
      const label = pad2(min);
      const bucket = slotMap.get(min) || [];
      if (bucket.length === 0) {
        missingMinutes.push(min);
        return {
          minute: min,
          label,
          state: 'MISSING',
          count: 0,
          timestampText: '',
          nullKeys: [],
          invalidKeys: [],
          values: {},
        };
      }
      presentSlots += 1;
      if (bucket.length > 1) duplicateMinutes.push(min);

      const primary = bucket[0];
      const nullKeys = [];
      const invalidKeys = [];
      const vals = {};
      for (const e of primary.entries) {
        if (e.isNull) {
          nullKeys.push(e.code);
          vals[e.code] = null;
        } else if (e.isInvalidType || e.outOfRange) {
          invalidKeys.push(e.code);
          vals[e.code] = e.rawValue;
        } else {
          vals[e.code] = e.value;
        }
      }

      const isComplete = primary.entries.length > 0 && nullKeys.length === 0 && invalidKeys.length === 0;
      if (isComplete) completeSlots += 1;
      else partialSlots += 1;

      return {
        minute: min,
        label,
        state: isComplete ? 'COMPLETE' : 'PARTIAL',
        count: bucket.length,
        timestampText: primary.ts.formatted,
        nullKeys,
        invalidKeys,
        values: vals,
      };
    });

    const hourBuckets = Array.from(hourSet).sort();
    const windowLabel = hourBuckets.length === 1
      ? `${hourBuckets[0]} – ${hourBuckets[0].slice(0, 13)}:58`
      : hourBuckets.length > 1
        ? `Lintas ${hourBuckets.length} jam (${hourBuckets[0]} .. ${hourBuckets[hourBuckets.length - 1]})`
        : 'Jam tidak terdeteksi';

    sparingAudit = {
      expectedSlots: 30,
      presentSlots,
      completeSlots,
      partialSlots,
      missingMinutes,
      duplicateMinutes,
      offGridMinutes: offGridMinutes.sort((a, b) => a - b),
      hourBuckets,
      windowLabel,
      slots,
    };

    if (missingMinutes.length === 0 && offGridMinutes.length === 0 && duplicateMinutes.length === 0) {
      completenessChecks.push({
        ok: true,
        level: 'PASS',
        label: 'Slot Menit SPARING (00–58)',
        detail: `Lengkap 30/30 slot menit genap (:00 s/d :58) pada jendela ${windowLabel}`,
      });
    } else {
      const missStr = missingMinutes.slice(0, 8).map(m => ':' + pad2(m)).join(', ') + (missingMinutes.length > 8 ? ` (+${missingMinutes.length - 8} lagi)` : '');
      if (missingMinutes.length > 0) {
        completenessChecks.push({
          ok: false,
          level: 'FAIL',
          label: 'Slot Menit SPARING (00–58)',
          detail: `Hanya ${presentSlots}/30 slot — hilang ${missingMinutes.length} menit: ${missStr}`,
        });
        issues.push({
          category: 'KELENGKAPAN',
          level: 'FAIL',
          message: `Data SPARING menit 00–58 tidak lengkap (${presentSlots}/30 slot; hilang: ${missStr})`,
        });
      }
      if (duplicateMinutes.length > 0) {
        const dupStr = duplicateMinutes.map(m => ':' + pad2(m)).join(', ');
        completenessChecks.push({
          ok: false,
          level: 'WARN',
          label: 'Duplikasi Slot Menit',
          detail: `Menit terduplikasi: ${dupStr}`,
        });
        issues.push({ category: 'KELENGKAPAN', level: 'WARN', message: `Terdapat slot menit SPARING duplikat: ${dupStr}` });
      }
      if (offGridMinutes.length > 0) {
        const oddStr = offGridMinutes.map(m => ':' + pad2(m)).join(', ');
        completenessChecks.push({
          ok: false,
          level: 'FAIL',
          label: 'Keselarasan Interval 2 Menit',
          detail: `Menit ganjil (di luar 00–58 genap): ${oddStr}`,
        });
        issues.push({ category: 'KELENGKAPAN', level: 'FAIL', message: `Menit SPARING tidak sejajar interval 2 menit (menit ganjil: ${oddStr})` });
      }
    }

    if (hourBuckets.length > 1) {
      completenessChecks.push({
        ok: false,
        level: 'FAIL',
        label: 'Konsistensi 1 Jam Window',
        detail: `Timestamp tersebar di ${hourBuckets.length} jam berbeda (${hourBuckets.join(', ')})`,
      });
      issues.push({ category: 'KELENGKAPAN', level: 'FAIL', message: `Batch SPARING mencampur ${hourBuckets.length} jam berbeda (${hourBuckets.join(', ')})` });
    } else if (hourBuckets.length === 1) {
      completenessChecks.push({
        ok: true,
        level: 'PASS',
        label: 'Konsistensi 1 Jam Window',
        detail: `Seluruh slot berada pada 1 jam yang sama (${hourBuckets[0]})`,
      });
    }

    if (partialSlots > 0) {
      completenessChecks.push({
        ok: false,
        level: 'WARN',
        label: 'Kelengkapan Nilai per Slot',
        detail: `${completeSlots}/30 slot terisi penuh · ${partialSlots} slot berisi null/parsial`,
      });
      issues.push({ category: 'KELENGKAPAN', level: 'WARN', message: `${partialSlots} slot menit SPARING hadir namun memiliki nilai null/parsial` });
    } else if (presentSlots === 30) {
      completenessChecks.push({
        ok: true,
        level: 'PASS',
        label: 'Kelengkapan Nilai per Slot',
        detail: 'Seluruh 30/30 slot memiliki nilai parameter penuh tanpa null',
      });
    }
  } else {
    // Non-SPARING completeness checks (AQMS / WQMS / ONLIMO / CBI Multi-Tick)
    const paramCount = parameterStats.length;
    completenessChecks.push({
      ok: paramCount > 0,
      level: paramCount > 0 ? 'PASS' : 'FAIL',
      label: 'Kelengkapan Kanal Sensor',
      detail: paramCount > 0 ? `${paramCount} parameter sensor terbaca dalam ${parsedRecords.length} record` : 'Tidak ada kanal sensor pada payload',
    });

    if (schemaKind === 'CBI_MULTI') {
      const dupList = parsedRecords.flatMap(r => r.duplicateCodes || []);
      if (dupList.length > 0) {
        completenessChecks.push({
          ok: false,
          level: 'WARN',
          label: 'Unik Sensor per Tick',
          detail: `Sensor duplikat pada timestamp yang sama: ${dupList.join(', ')}`,
        });
        issues.push({ category: 'KELENGKAPAN', level: 'WARN', message: `Terdapat duplikasi sensor dalam 1 tick (${dupList.join(', ')})` });
      }
    }
  }

  if (skippedParams > 0) {
    completenessChecks.push({
      ok: false,
      level: 'WARN',
      label: 'Pemetaan Parameter Cloud',
      detail: `${skippedParams} parameter dilewati (tidak terdaftar di stasiun tujuan)`,
    });
    issues.push({
      category: 'KELENGKAPAN',
      level: 'WARN',
      message: `${skippedParams} parameter dari payload tidak dipetakan/dilewati oleh Cloud Receiver`,
    });
  } else if (httpAccepted) {
    completenessChecks.push({
      ok: true,
      level: 'PASS',
      label: 'Pemetaan Parameter Cloud',
      detail: `${Number(log.parameters_ingested || validValues)} parameter diterima tanpa ada yang dilewati`,
    });
  }

  // Compute pillar statuses
  const reducePillarStatus = (checks) => {
    if (checks.some(c => c.level === 'FAIL')) return 'FAIL';
    if (checks.some(c => c.level === 'WARN')) return 'WARN';
    return 'PASS';
  };

  const formatStatus = reducePillarStatus(formatChecks);
  const valuesStatus = reducePillarStatus(valueChecks);
  const completenessStatus = reducePillarStatus(completenessChecks);

  // Score calculation (0-100): Format 30 pts + Values 35 pts + Completeness 35 pts
  let formatScore = formatStatus === 'PASS' ? 30 : formatStatus === 'WARN' ? 22 : 0;
  let valueScore = 35;
  if (totalValues === 0) {
    valueScore = 0;
  } else {
    const validRatio = validValues / totalValues;
    valueScore = Math.round(validRatio * 35);
    if (invalidTypeCount > 0 || outOfRangeCount > 0) {
      valueScore = Math.min(valueScore, 15);
    }
  }
  let compScore = 35;
  if (isSparing && sparingAudit) {
    const slotRatio = sparingAudit.completeSlots / 30;
    const presenceRatio = sparingAudit.presentSlots / 30;
    compScore = Math.round((slotRatio * 0.65 + presenceRatio * 0.35) * 35);
    if (sparingAudit.hourBuckets.length > 1 || sparingAudit.offGridMinutes.length > 0) {
      compScore = Math.max(0, compScore - 10);
    }
  } else if (completenessStatus === 'FAIL') {
    compScore = 0;
  } else if (completenessStatus === 'WARN') {
    compScore = 26;
  }
  if (skippedParams > 0) compScore = Math.max(0, compScore - 5);

  const score = Math.max(0, Math.min(100, formatScore + valueScore + compScore));

  let overallStatus = 'VALID';
  if (formatStatus === 'FAIL' || valuesStatus === 'FAIL' || completenessStatus === 'FAIL') {
    overallStatus = 'INVALID';
  } else if (formatStatus === 'WARN' || valuesStatus === 'WARN' || completenessStatus === 'WARN') {
    overallStatus = 'WARNING';
  }

  // Concise labels for table cells
  const formatLabel = formatStatus === 'PASS' ? 'Valid' : formatStatus === 'WARN' ? 'Peringatan' : 'Gagal';
  const valuesLabel =
    totalValues === 0
      ? 'Kosong'
      : invalidTypeCount + outOfRangeCount > 0
        ? `${invalidTypeCount + outOfRangeCount} Anomali`
        : nullCount > 0
          ? `${nullCount} Null`
          : `${validValues}/${totalValues} OK`;

  const completenessLabel = isSparing && sparingAudit
    ? `${sparingAudit.completeSlots}/30 (00–58)`
    : skippedParams > 0
      ? `${parameterStats.length}p · ${skippedParams} skip`
      : `${parameterStats.length} param`;

  return {
    overallStatus,
    score,
    isSparing,
    schemaLabel,
    format: {
      status: formatStatus,
      label: formatLabel,
      summary: `${schemaLabel} · ${validTsCount}/${records.length} ts`,
      checks: formatChecks,
    },
    values: {
      status: valuesStatus,
      label: valuesLabel,
      summary: `${validValues}/${totalValues} valid${nullCount ? `, ${nullCount} null` : ''}${outOfRangeCount ? `, ${outOfRangeCount} range` : ''}`,
      totalValues,
      validValues,
      nullCount,
      invalidTypeCount,
      outOfRangeCount,
      parameterStats,
      checks: valueChecks,
    },
    completeness: {
      status: completenessStatus,
      label: completenessLabel,
      summary: isSparing && sparingAudit
        ? `${sparingAudit.presentSlots}/30 slot hadir (${sparingAudit.completeSlots} penuh, ${sparingAudit.missingMinutes.length} hilang)`
        : `${parsedRecords.length} record · ${parameterStats.length} parameter`,
      recordCount: parsedRecords.length,
      paramCount: parameterStats.length,
      skippedParams,
      sparing: sparingAudit,
      checks: completenessChecks,
    },
    issues,
  };
}
