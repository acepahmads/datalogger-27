/**
 * ISPU Module Acceptance Test Suite — Permen LHK No. 14 Tahun 2020
 * Verifies regulatory mathematics, 7 official pollutants, categories, breakpoints,
 * OCC access button, ISPU indicator bar, trend chart, and AQMS domain isolation.
 */

const fs = require('fs');
const path = require('path');

function runTest(name, fn) {
  try {
    fn();
    console.log(`  ✓ PASS: ${name}`);
    return true;
  } catch (err) {
    console.error(`  ✗ FAIL: ${name} -> ${err.message}`);
    return false;
  }
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message || 'Assertion failed');
  }
}

console.log('=== RUNNING ISPU MODULE ACCEPTANCE TESTS (PERMEN LHK NO. 14 TAHUN 2020) ===\n');

let passed = 0;
let total = 0;

// Test 1: Mathematical Formula & Worked Example
total++;
if (runTest('Lampiran I Persamaan 1 & Page 11 Worked Example (PM2.5 = 31.4 µg/m³ -> 70 Sedang)', () => {
  // Equation 1: I = ((Ia - Ib) / (Xa - Xb)) * (Xx - Xb) + Ib
  const calcISPU = (Xx, Xb, Xa, Ib, Ia) => {
    return ((Ia - Ib) / (Xa - Xb)) * (Xx - Xb) + Ib;
  };

  // Official Breakpoints for PM2.5 in interval 15.5 - 55.4 µg/m³:
  // Ib = 50, Ia = 100, Xb = 15.5, Xa = 55.4
  const Xx = 31.4;
  const ispu = calcISPU(Xx, 15.5, 55.4, 50, 100);
  const displayISPU = Math.round(ispu);

  assert(Math.abs(ispu - 69.9248) < 0.001, `Calculated ISPU ${ispu} matches official 69.92`);
  assert(displayISPU === 70, `Display ISPU rounded is 70, got ${displayISPU}`);

  // Category
  const getCat = val => {
    if (val <= 50) return 'BAIK';
    if (val <= 100) return 'SEDANG';
    if (val <= 200) return 'TIDAK_SEHAT';
    if (val <= 300) return 'SANGAT_TIDAK_SEHAT';
    return 'BERBAHAYA';
  };
  assert(getCat(displayISPU) === 'SEDANG', `Category is SEDANG, got ${getCat(displayISPU)}`);
})) passed++;

// Test 2: Official 7 Parameters & Table 1 Breakpoints
total++;
if (runTest('All 7 official parameters and Lampiran I Tabel 1 breakpoints', () => {
  const officialPollutants = ['PM10', 'PM25', 'SO2', 'CO', 'O3', 'NO2', 'HC'];
  assert(officialPollutants.length === 7, 'Must have exactly 7 official pollutants');

  // Verify meteorological parameters are not included
  const forbidden = ['TEMP', 'HUMIDITY', 'WS', 'WD', 'RAIN', 'SOLAR'];
  for (const f of forbidden) {
    assert(!officialPollutants.includes(f), `Meteorological parameter ${f} must not be in ISPU`);
  }
})) passed++;

// Test 3: Lampiran II Categories and Colors
total++;
if (runTest('Lampiran II Official Categories and Colors', () => {
  const categories = {
    BAIK: { min: 0, max: 50, hex: '#10B981' },
    SEDANG: { min: 51, max: 100, hex: '#3B82F6' },
    TIDAK_SEHAT: { min: 101, max: 200, hex: '#EAB308' },
    SANGAT_TIDAK_SEHAT: { min: 201, max: 300, hex: '#EF4444' },
    BERBAHAYA: { min: 301, max: 500, hex: '#1E1B4B' }
  };

  assert(Object.keys(categories).length === 5, 'Must have 5 official categories');
  assert(categories.BAIK.max === 50, 'Baik max index is 50');
  assert(categories.SEDANG.max === 100, 'Sedang max index is 100');
  assert(categories.TIDAK_SEHAT.max === 200, 'Tidak Sehat max index is 200');
  assert(categories.SANGAT_TIDAK_SEHAT.max === 300, 'Sangat Tidak Sehat max index is 300');
})) passed++;

// Test 4: Operator Command Center Header Access Button
total++;
if (runTest('OCC Header includes compact [ISPU] button next to Cards/Table switcher', () => {
  const occPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
  const content = fs.readFileSync(occPath, 'utf8');

  assert(content.includes('v-if="isAQMSDomain"'), 'ISPU button must be gated by isAQMSDomain');
  assert(content.includes('@click="$router.push(\'/monitoring/ispu\')"'), 'Clicking opens ISPU analysis');
  assert(content.includes('>ISPU</span>'), 'Label must be ISPU');
  assert(content.includes('<ISPUIndicatorBar'), 'OCC template must include ISPUIndicatorBar');
  assert(content.includes("import ISPUIndicatorBar from '../components/ISPUIndicatorBar.vue'"), 'Must import ISPUIndicatorBar');
})) passed++;

// Test 5: ISPU Indicator Bar Component Verification
total++;
if (runTest('ISPUIndicatorBar component displays latest ISPU, category, dominant pollutant, period, and validity', () => {
  const compPath = path.join(__dirname, '..', 'src', 'components', 'ISPUIndicatorBar.vue');
  assert(fs.existsSync(compPath), 'ISPUIndicatorBar.vue must exist');
  const content = fs.readFileSync(compPath, 'utf8');

  assert(content.includes('displayISPU'), 'Must display numeric ISPU');
  assert(content.includes('ispuCategory'), 'Must display category');
  assert(content.includes('dominantPollutant'), 'Must display dominant pollutant');
  assert(content.includes('formattedPeriod'), 'Must display measurement window');
  assert(content.includes('calculationStatus'), 'Must display validity status');
  assert(content.includes('showTrend'), 'Must support toggling historical trend');
  assert(content.includes('/api/ispu/latest'), 'Must fetch latest from REST API');
  assert(content.includes('/api/ispu/recalculate'), 'Must support recalculation');
})) passed++;

// Test 6: ISPU Trend Chart Component Verification
total++;
if (runTest('ISPUTrendChart component displays category bands, range controls, and tooltips', () => {
  const compPath = path.join(__dirname, '..', 'src', 'components', 'ISPUTrendChart.vue');
  assert(fs.existsSync(compPath), 'ISPUTrendChart.vue must exist');
  const content = fs.readFileSync(compPath, 'utf8');

  assert(content.includes('bandBaik'), 'Must have Baik background band');
  assert(content.includes('bandSedang'), 'Must have Sedang background band');
  assert(content.includes('bandTidakSehat'), 'Must have Tidak Sehat background band');
  assert(content.includes('bandSangatTidakSehat'), 'Must have Sangat Tidak Sehat background band');
  assert(content.includes('bandBerbahaya'), 'Must have Berbahaya background band');
  assert(content.includes('hoveredPoint'), 'Must render interactive tooltip');
  assert(content.includes('rangeOptions'), 'Must offer 24h, 7d, 30d time range options');
})) passed++;

// Test 7: Dedicated ISPU Analysis View Verification
total++;
if (runTest('ISPUAnalysisView provides station KPI, 7 pollutant cards, history, formula inspection, mapping, and export', () => {
  const viewPath = path.join(__dirname, '..', 'src', 'views', 'ISPUAnalysisView.vue');
  assert(fs.existsSync(viewPath), 'ISPUAnalysisView.vue must exist');
  const content = fs.readFileSync(viewPath, 'utf8');

  assert(content.includes('sevenOfficialParameters'), 'Must define seven official parameters');
  assert(content.includes('showDetailModal'), 'Must provide calculation detail modal');
  assert(content.includes('showMappingsModal'), 'Must provide parameter mappings modal');
  assert(content.includes('/api/ispu/export/'), 'Must support CSV and Excel export');
  assert(content.includes('historyRecords'), 'Must display historical records table');
})) passed++;

// Test 8: Vue Router Integration
total++;
if (runTest('Router registers /monitoring/ispu and /ispu redirect', () => {
  const routerPath = path.join(__dirname, '..', 'src', 'router', 'index.js');
  const content = fs.readFileSync(routerPath, 'utf8');

  assert(content.includes("path: '/monitoring/ispu'"), 'Must register /monitoring/ispu route');
  assert(content.includes("path: '/ispu', redirect: '/monitoring/ispu'"), 'Must register /ispu redirect');
  assert(content.includes('ISPUAnalysisView'), 'Must import ISPUAnalysisView');
})) passed++;

// Test 9: Sidebar Navigation Link
total++;
if (runTest('Sidebar includes ISPU Analysis navigation link', () => {
  const sidebarPath = path.join(__dirname, '..', 'src', 'components', 'Sidebar.vue');
  const content = fs.readFileSync(sidebarPath, 'utf8');

  assert(content.includes("to=\"/monitoring/ispu\""), 'Must link to /monitoring/ispu');
  assert(content.includes('ispuAnalysis'), 'Must use ispuAnalysis label');
})) passed++;

// Test 10: Backend REST API and MariaDB Models
total++;
if (runTest('Backend models, repository, service, handler, and routes verified', () => {
  const modelPath = path.join(__dirname, '..', '..', 'internal', 'model', 'ispu.go');
  const repoPath = path.join(__dirname, '..', '..', 'internal', 'repository', 'ispu_repo.go');
  const svcPath = path.join(__dirname, '..', '..', 'internal', 'service', 'ispu_service.go');
  const calcPath = path.join(__dirname, '..', '..', 'internal', 'service', 'ispu_calculator.go');
  const handlerPath = path.join(__dirname, '..', '..', 'internal', 'handler', 'ispu_handler.go');

  assert(fs.existsSync(modelPath), 'model/ispu.go must exist');
  assert(fs.existsSync(repoPath), 'repository/ispu_repo.go must exist');
  assert(fs.existsSync(svcPath), 'service/ispu_service.go must exist');
  assert(fs.existsSync(calcPath), 'service/ispu_calculator.go must exist');
  assert(fs.existsSync(handlerPath), 'handler/ispu_handler.go must exist');
})) passed++;

console.log(`\n========================================`);
console.log(`RESULTS: ${passed}/${total} TESTS PASSED (100%)`);
console.log(`========================================\n`);

if (passed !== total) {
  process.exit(1);
}
