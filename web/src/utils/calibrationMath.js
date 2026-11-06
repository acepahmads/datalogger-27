// Sensor Calibration & Multi-Point Regression Engine (1–10 Points)
// Supports AQMS (ISPU/Gas/Particulate), SPARING (Wastewater), and WQMS ONLIMO (Surface/River Water)

export const CALIBRATION_MODELS = [
  {
    value: 'LINEAR',
    shortLabel: 'Linier (y = ax + b)',
    description: 'Regresi kuadrat terkecil orde-1 (2–10 titik, termasuk Zero-Span)',
    minPoints: 2,
  },
  {
    value: 'QUADRATIC',
    shortLabel: 'Kuadratik Orde-2 (y = ax² + bx + c)',
    description: 'Polinomial orde-2 untuk sensor optik/elektroda non-linier (3–10 titik)',
    minPoints: 3,
  },
  {
    value: 'OFFSET_1PT',
    shortLabel: '1-Titik Offset (y = x + b)',
    description: 'Koreksi pergeseran nol / bias tetap (1–10 titik)',
    minPoints: 1,
  },
  {
    value: 'SPAN_1PT',
    shortLabel: '1-Titik Span (y = ax)',
    description: 'Koreksi faktor kemiringan / sensitivitas proporsional (1–10 titik)',
    minPoints: 1,
  },
];

export const CALIBRATION_DOMAINS = [
  { value: 'ALL', label: 'Semua Domain', badge: 'ALL' },
  { value: 'AQMS', label: 'AQMS (Udara & ISPU)', badge: 'AQMS' },
  { value: 'SPARING', label: 'SPARING (Air Limbah)', badge: 'SPARING' },
  { value: 'WQMS_ONLIMO', label: 'WQMS ONLIMO (Air Sungai)', badge: 'ONLIMO' },
];

export function normalizeMinR2(val) {
  const num = Number(val);
  if (!Number.isFinite(num) || num <= 0 || num > 1) return 0.995;
  return num < 0.5 ? 0.5 : num;
}

export function evaluateCalibrationCurve(enabled, modelType, c2, c1, c0, x) {
  const xv = Number(x);
  if (!enabled || !Number.isFinite(xv)) return xv;
  const m = String(modelType || 'LINEAR').toUpperCase();
  if (m === 'NONE') return xv;
  if (m === 'OFFSET_1PT') return xv + Number(c0 || 0);
  if (m === 'SPAN_1PT') return Number(c1 || 1) * xv;
  if (m === 'QUADRATIC') {
    return Number(c2 || 0) * xv * xv + Number(c1 || 0) * xv + Number(c0 || 0);
  }
  return Number(c1 || 1) * xv + Number(c0 || 0);
}

export function formatCalibrationEquation(enabled, modelType, c2, c1, c0) {
  if (!enabled) return 'y = x (Default)';
  const m = String(modelType || 'LINEAR').toUpperCase();
  const a2 = Number(c2 || 0);
  const a1 = Number(c1 || 0);
  const a0 = Number(c0 || 0);

  if (m === 'NONE') return 'y = x (Default)';
  if (m === 'OFFSET_1PT') {
    if (Math.abs(a0) < 1e-9) return 'y = x + 0.0000';
    return a0 >= 0 ? `y = x + ${a0.toFixed(4)}` : `y = x - ${Math.abs(a0).toFixed(4)}`;
  }
  if (m === 'SPAN_1PT') {
    return `y = ${a1.toFixed(4)}x`;
  }
  if (m === 'QUADRATIC') {
    const bSign = a1 < 0 ? '-' : '+';
    const cSign = a0 < 0 ? '-' : '+';
    return `y = ${a2.toFixed(5)}x² ${bSign} ${Math.abs(a1).toFixed(4)}x ${cSign} ${Math.abs(a0).toFixed(4)}`;
  }
  return a0 >= 0
    ? `y = ${a1.toFixed(4)}x + ${a0.toFixed(4)}`
    : `y = ${a1.toFixed(4)}x - ${Math.abs(a0).toFixed(4)}`;
}

function solveQuadraticOLS(xs, ys) {
  const n = xs.length;
  let mu = 0;
  for (let i = 0; i < n; i++) mu += xs[i];
  mu /= n;

  let maxAbs = 0;
  for (let i = 0; i < n; i++) {
    const d = Math.abs(xs[i] - mu);
    if (d > maxAbs) maxAbs = d;
  }
  if (maxAbs < 1e-12) {
    return { error: 'Nilai bacaan sensor (x) kurang bervariasi untuk regresi kuadratik.' };
  }
  const scale = maxAbs;

  let s0 = 0, s1 = 0, s2 = 0, s3 = 0, s4 = 0;
  let t0 = 0, t1 = 0, t2 = 0;
  for (let i = 0; i < n; i++) {
    const u = (xs[i] - mu) / scale;
    const u2 = u * u;
    const y = ys[i];
    s0 += 1;
    s1 += u;
    s2 += u2;
    s3 += u2 * u;
    s4 += u2 * u2;
    t0 += y;
    t1 += u * y;
    t2 += u2 * y;
  }

  const mat = [
    [s0, s1, s2, t0],
    [s1, s2, s3, t1],
    [s2, s3, s4, t2],
  ];

  for (let col = 0; col < 3; col++) {
    let pivot = col;
    let maxPivot = Math.abs(mat[col][col]);
    for (let row = col + 1; row < 3; row++) {
      const v = Math.abs(mat[row][col]);
      if (v > maxPivot) {
        maxPivot = v;
        pivot = row;
      }
    }
    if (maxPivot < 1e-11) {
      return { error: 'Minimal 3 nilai bacaan sensor (x) yang berbeda diperlukan untuk regresi kuadratik.' };
    }
    if (pivot !== col) {
      const tmp = mat[col];
      mat[col] = mat[pivot];
      mat[pivot] = tmp;
    }
    for (let row = col + 1; row < 3; row++) {
      const factor = mat[row][col] / mat[col][col];
      for (let k = col; k < 4; k++) {
        mat[row][k] -= factor * mat[col][k];
      }
    }
  }

  const sol = [0, 0, 0];
  for (let row = 2; row >= 0; row--) {
    let sum = mat[row][3];
    for (let k = row + 1; k < 3; k++) {
      sum -= mat[row][k] * sol[k];
    }
    sol[row] = sum / mat[row][row];
  }

  const [a0, a1, a2] = sol;
  const invS = 1 / scale;
  const invS2 = invS * invS;

  return {
    c2: a2 * invS2,
    c1: a1 * invS - 2 * a2 * mu * invS2,
    c0: a0 - a1 * mu * invS + a2 * mu * mu * invS2,
    error: null,
  };
}

export function calculateRegressionJS(modelType, inputPoints, minR2 = 0.995) {
  const list = Array.isArray(inputPoints) ? inputPoints : [];
  if (list.length < 1) {
    return { valid: false, error: 'Minimal 1 titik kalibrasi diperlukan.' };
  }
  if (list.length > 10) {
    return { valid: false, error: 'Maksimal 10 titik kalibrasi diperbolehkan.' };
  }

  const normModel = String(modelType || 'LINEAR').toUpperCase();
  const effMinR2 = normalizeMinR2(minR2);

  const points = [];
  const activeX = [];
  const activeY = [];

  for (let i = 0; i < list.length; i++) {
    const p = list[i];
    const rx = Number(p.raw_x);
    const ry = Number(p.ref_y);
    if (p.raw_x === '' || p.raw_x === null || p.ref_y === '' || p.ref_y === null || !Number.isFinite(rx) || !Number.isFinite(ry)) {
      return { valid: false, error: `Lengkapi angka bacaan sensor (x) dan standar referensi (y) pada baris #${i + 1}.` };
    }
    const copy = {
      index: i + 1,
      active: p.active !== false,
      label: p.label || '',
      raw_x: rx,
      ref_y: ry,
      pred_y: rx,
      error_abs: 0,
      error_pct: null,
    };
    points.push(copy);
    if (copy.active) {
      activeX.push(rx);
      activeY.push(ry);
    }
  }

  const n = activeX.length;
  if (n === 0) {
    return { valid: false, error: 'Minimal 1 titik standar harus dicentang sebagai titik Aktif.' };
  }

  let c2 = 0, c1 = 1, c0 = 0;
  let advisory = '';

  if (normModel === 'OFFSET_1PT') {
    let sumDiff = 0;
    for (let i = 0; i < n; i++) sumDiff += activeY[i] - activeX[i];
    c2 = 0;
    c1 = 1;
    c0 = sumDiff / n;
    if (n > 1) {
      advisory = 'Mode 1-Titik Offset menggunakan rata-rata selisih (y - x) dari titik aktif.';
    }
  } else if (normModel === 'SPAN_1PT') {
    let sumXX = 0, sumXY = 0;
    for (let i = 0; i < n; i++) {
      sumXX += activeX[i] * activeX[i];
      sumXY += activeX[i] * activeY[i];
    }
    if (sumXX < 1e-12) {
      return { valid: false, error: 'Bacaan sensor (x) tidak boleh 0 untuk kalibrasi 1-Titik Span.' };
    }
    c2 = 0;
    c1 = sumXY / sumXX;
    c0 = 0;
  } else if (normModel === 'QUADRATIC') {
    if (n < 3) {
      return { valid: false, error: 'Polinomial Kuadratik Orde-2 membutuhkan minimal 3 titik aktif.' };
    }
    const q = solveQuadraticOLS(activeX, activeY);
    if (q.error) return { valid: false, error: q.error };
    c2 = q.c2;
    c1 = q.c1;
    c0 = q.c0;
    if (n === 3) {
      advisory = 'Dengan tepat 3 titik, kurva kuadratik melewati seluruh titik (R² = 1). Disarankan ≥ 4 titik untuk menguji residu.';
    }
  } else {
    if (n < 2) {
      return { valid: false, error: 'Regresi Linier membutuhkan minimal 2 titik aktif (atau pilih model 1-Titik).' };
    }
    let meanX = 0, meanY = 0;
    for (let i = 0; i < n; i++) {
      meanX += activeX[i];
      meanY += activeY[i];
    }
    meanX /= n;
    meanY /= n;

    let sxx = 0, sxy = 0;
    for (let i = 0; i < n; i++) {
      const dx = activeX[i] - meanX;
      const dy = activeY[i] - meanY;
      sxx += dx * dx;
      sxy += dx * dy;
    }
    if (sxx < 1e-12) {
      return { valid: false, error: 'Seluruh bacaan sensor (x) bernilai sama; minimal 2 nilai x berbeda diperlukan.' };
    }
    c2 = 0;
    c1 = sxy / sxx;
    c0 = meanY - c1 * meanX;
  }

  if (Math.abs(c2) < 1e-15) c2 = 0;
  if (Math.abs(c1) < 1e-15) c1 = 0;
  if (Math.abs(c0) < 1e-15) c0 = 0;

  let meanY = 0;
  for (let i = 0; i < n; i++) meanY += activeY[i];
  meanY /= n;

  let sse = 0;
  let syy = 0;
  for (let i = 0; i < points.length; i++) {
    const pt = points[i];
    const pred = evaluateCalibrationCurve(true, normModel, c2, c1, c0, pt.raw_x);
    const errAbs = pred - pt.ref_y;
    pt.pred_y = pred;
    pt.error_abs = errAbs;
    pt.error_pct = Math.abs(pt.ref_y) >= 1e-6 ? (errAbs / Math.abs(pt.ref_y)) * 100.0 : null;

    if (pt.active) {
      const diff = pt.ref_y - pred;
      sse += diff * diff;
      const dy = pt.ref_y - meanY;
      syy += dy * dy;
    }
  }

  const rmse = Math.sqrt(sse / n);
  let r2 = null;
  let passedR2 = true;

  if (n >= 2 && normModel !== 'OFFSET_1PT' && normModel !== 'SPAN_1PT') {
    let val = 1.0;
    if (syy > 1e-12) {
      val = Math.min(1.0, 1.0 - sse / syy);
    } else if (sse > 1e-9) {
      val = 0.0;
    }
    r2 = val;
    passedR2 = val >= effMinR2;
  } else if (n >= 2 && syy > 1e-12) {
    r2 = Math.min(1.0, 1.0 - sse / syy);
    passedR2 = r2 >= effMinR2;
  }

  return {
    valid: true,
    error: null,
    model: normModel,
    c2,
    c1,
    c0,
    equation: formatCalibrationEquation(true, normModel, c2, c1, c0),
    active_count: n,
    total_count: points.length,
    r2,
    min_r2: effMinR2,
    passed_r2: passedR2,
    rmse,
    points,
    advisory_message: advisory,
  };
}

export function evaluateReadingStability(values = []) {
  const nums = (Array.isArray(values) ? values : []).map(Number).filter(Number.isFinite);
  if (nums.length < 2) {
    return {
      isStable: nums.length === 1,
      sigma: 0,
      mean: nums.length === 1 ? nums[0] : null,
      sampleCount: nums.length,
    };
  }
  const mean = nums.reduce((a, b) => a + b, 0) / nums.length;
  const variance = nums.reduce((acc, v) => acc + (v - mean) * (v - mean), 0) / nums.length;
  const sigma = Math.sqrt(variance);
  // Adaptive tolerance: max(0.05, 0.5% of magnitude) so both pH (0.05) and DHL (1413 uS/cm -> ~7 uS/cm) work accurately
  const threshold = Math.max(0.05, Math.abs(mean) * 0.005);
  return {
    isStable: sigma <= threshold,
    sigma,
    mean,
    sampleCount: nums.length,
  };
}

// Smart field calibration presets for AQMS, SPARING, and WQMS ONLIMO
export function getDomainStandardPresets(param) {
  const tok = `${(param && param.parameter_code) || ''} ${(param && param.parameter_name) || ''} ${(param && param.unit) || ''}`.toUpperCase();
  const presets = [];

  if (tok.includes('PH')) {
    presets.push(
      {
        id: 'ph_nist_3pt',
        label: 'Buffer pH NIST 3-Titik (4.01, 7.00, 10.01)',
        model: 'LINEAR',
        note: 'Larutan Buffer Standar NIST pH 4.01, 7.00, 10.01 @ 25°C',
        refs: [4.01, 7.00, 10.01],
      },
      {
        id: 'ph_acid_2pt',
        label: 'Buffer pH 2-Titik (4.01 & 7.00)',
        model: 'LINEAR',
        note: 'Larutan Buffer Standar pH 4.01 & 7.00',
        refs: [4.01, 7.00],
      }
    );
  }

  if (tok.includes('COD')) {
    presets.push(
      {
        id: 'cod_khp_5pt',
        label: 'CRM COD KHP 5-Titik (0–200 mg/L)',
        model: 'LINEAR',
        note: 'Larutan Standar KHP CRM (0, 25, 50, 100, 200 mg/L)',
        refs: [0, 25, 50, 100, 200],
      },
      {
        id: 'cod_khp_6pt_high',
        label: 'CRM COD Rentang Tinggi 6-Titik (0–1000 mg/L)',
        model: 'QUADRATIC',
        note: 'Larutan Standar KHP Rentang Lebar (0–1000 mg/L)',
        refs: [0, 50, 100, 250, 500, 1000],
      }
    );
  }

  if (tok.includes('TSS')) {
    presets.push(
      {
        id: 'tss_formazin_5pt',
        label: 'Standar TSS Formazin/Kaolin 5-Titik (0–500 mg/L)',
        model: 'LINEAR',
        note: 'Standar Suspensi TSS (0, 50, 100, 250, 500 mg/L)',
        refs: [0, 50, 100, 250, 500],
      },
      {
        id: 'tss_quad_6pt',
        label: 'TSS Optik Non-Linier 6-Titik (0–1000 mg/L)',
        model: 'QUADRATIC',
        note: 'Standar TSS Konsentrasi Tinggi Orde-2',
        refs: [0, 50, 150, 300, 600, 1000],
      }
    );
  }

  if (tok.includes('NH3') || tok.includes('AMMONIA') || tok.includes('AMONIA')) {
    presets.push({
      id: 'nh3_crm_5pt',
      label: 'CRM NH3-N Amonia 5-Titik (0–25 mg/L)',
      model: 'LINEAR',
      note: 'Larutan Standar NH4Cl CRM (0, 1, 5, 10, 25 mg/L)',
      refs: [0, 1, 5, 10, 25],
    });
  }

  if (tok.includes('DO') || tok.includes('OKSIGEN') || tok.includes('DISSOLVED')) {
    presets.push({
      id: 'do_onlimo_2pt',
      label: 'ONLIMO DO 2-Titik (Zero Na2SO3 & Air Jenuh 8.26 mg/L)',
      model: 'LINEAR',
      note: 'Zero Oksigen (Na2SO3) 0.00 mg/L & Air Jenuh Oksigen 8.26 mg/L @ 25°C',
      refs: [0.00, 8.26],
    });
  }

  if (tok.includes('DHL') || tok.includes('COND') || tok.includes('US/CM') || tok.includes('µS/CM')) {
    presets.push({
      id: 'dhl_kcl_4pt',
      label: 'ONLIMO Konduktivitas KCl 4-Titik (0–12880 µS/cm)',
      model: 'LINEAR',
      note: 'Larutan Standar KCl (Udara Kering 0, 84, 1413, 12880 µS/cm)',
      refs: [0, 84, 1413, 12880],
    });
  }

  if (tok.includes('TURB') || tok.includes('KEKERUHAN') || tok.includes('NTU')) {
    presets.push({
      id: 'ntu_formazin_5pt',
      label: 'ONLIMO Formazin Kekeruhan 5-Titik (0–800 NTU)',
      model: 'QUADRATIC',
      note: 'Standar Formazin Primer (0, 20, 100, 400, 800 NTU)',
      refs: [0, 20, 100, 400, 800],
    });
  }

  if (tok.includes('NO3') || tok.includes('NITRAT')) {
    presets.push({
      id: 'no3_crm_5pt',
      label: 'ONLIMO Nitrat (NO3-N) 5-Titik (0–20 mg/L)',
      model: 'LINEAR',
      note: 'Larutan Standar KNO3 CRM (0, 1, 5, 10, 20 mg/L)',
      refs: [0, 1, 5, 10, 20],
    });
  }

  if (tok.includes('ORP')) {
    presets.push({
      id: 'orp_zobell_1pt',
      label: 'ONLIMO Larutan Zobell ORP 1-Titik (+228 mV)',
      model: 'OFFSET_1PT',
      note: 'Larutan Standar ORP Zobell +228 mV @ 25°C',
      refs: [228.0],
    });
  }

  if (tok.includes('TEMP') || tok.includes('SUHU')) {
    presets.push({
      id: 'temp_nist_1pt',
      label: 'Termometer Referensi NIST 1-Titik Offset (25.0 °C)',
      model: 'OFFSET_1PT',
      note: 'Verifikasi Termometer Referensi Tersertifikasi',
      refs: [25.0],
    });
  }

  if (tok.includes('SO2') || tok.includes('NO2') || tok.includes('O3') || tok.includes('CO ') || tok.includes('HC') || tok.includes('PPB')) {
    presets.push(
      {
        id: 'aqms_gas_5pt',
        label: 'AQMS Dynamic Gas Dilution 5-Titik (Zero–80% Span)',
        model: 'LINEAR',
        note: 'Zero Air + Tabung Gas Span CRM via Dynamic Gas Calibrator (0, 100, 200, 300, 400)',
        refs: [0, 100, 200, 300, 400],
      },
      {
        id: 'aqms_gas_2pt',
        label: 'AQMS Gas Zero-Span 2-Titik (0 & 400)',
        model: 'LINEAR',
        note: 'Verifikasi Zero Air & Span Gas Standar',
        refs: [0, 400],
      }
    );
  }

  if (tok.includes('PM2.5') || tok.includes('PM10') || tok.includes('PM25') || tok.includes('TSP')) {
    presets.push(
      {
        id: 'aqms_pm_4pt',
        label: 'AQMS Korelasi BAM / Gravimetri 4-Titik (0–100 µg/m³)',
        model: 'LINEAR',
        note: 'Korelasi Colocated Sampler Referensi SNI/EPA',
        refs: [0, 25, 50, 100],
      },
      {
        id: 'aqms_flow_1pt',
        label: 'AQMS Audit Laju Alir Sampler 1-Titik Span (16.67 L/min)',
        model: 'SPAN_1PT',
        note: 'Kalibrasi Flow Meter Standar 16.67 L/min',
        refs: [16.67],
      }
    );
  }

  if (tok.includes('DEBIT') || tok.includes('FLOW')) {
    presets.push(
      {
        id: 'flow_zero_1pt',
        label: 'SPARING/ONLIMO Zero-Level Offset 1-Titik (0.00)',
        model: 'OFFSET_1PT',
        note: 'Koreksi Titik Nol Sensor Debit',
        refs: [0.00],
      },
      {
        id: 'flow_linear_3pt',
        label: 'Kalibrasi Laju Alir 3-Titik (0, 50, 100)',
        model: 'LINEAR',
        note: 'Verifikasi Laju Alir Multi-Titik',
        refs: [0, 50, 100],
      }
    );
  }

  return presets;
}

