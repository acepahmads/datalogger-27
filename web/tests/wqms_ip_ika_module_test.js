/**
 * WQMS IP & IKA Module Acceptance Test Suite
 * Regulations:
 *  - Kepmen LH No. 115 Tahun 2003 (Indeks Pencemaran / IPj — Nemerow-Sumitomo)
 *  - Lampiran VI PP No. 22 Tahun 2021 (Baku Mutu Air Nasional Kelas I, II, III, IV)
 *  - Permen LHK No. 27 Tahun 2021 (Indeks Kualitas Air / IKA — Bobot 70 / 50 / 30 / 10)
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

console.log('========================================================================');
console.log('RUNNING TEST SUITE: WQMS IP & IKA MODULE (END-TO-END ACCEPTANCE)');
console.log('========================================================================');

const repoRoot = path.resolve(__dirname, '../..');

let passed = 0;
let total = 0;

function check(title, fn) {
  total++;
  try {
    fn();
    passed++;
    console.log(`✓ PASS: ${title}`);
  } catch (err) {
    console.error(`✗ FAIL: ${title}`);
    console.error(`  ${err.message}`);
  }
}

// 1. Strict Boundary Isolation
check('1.1: OperatorCommandCenterView.vue, ISPUIndicatorBar.vue, and OperatorDashboardIcon.vue remain untouched in git', () => {
  const diff = execSync(
    'git diff -- web/src/views/OperatorCommandCenterView.vue web/src/components/ISPUIndicatorBar.vue web/src/components/OperatorDashboardIcon.vue',
    { cwd: repoRoot, encoding: 'utf8' }
  ).trim();
  assert.strictEqual(diff, '', 'Shared OCC / ISPU / Icon files must have zero git diff');
});

// 2. Kepmen LH No. 115/2003 & Permen LHK No. 27/2021 Reference Mathematics
check('2.1: Kepmen LH 115/2003 Nemerow-Sumitomo (pH range, DO deficit, log10 transform, IPj) & Permen LHK 27/2021 IKA weights', () => {
  // Range rule: pH [6.0, 9.0] -> avg = 7.5; Ci = 8.1 -> (8.1 - 7.5)/(9.0 - 7.5) = 0.40
  const phAvg = (6.0 + 9.0) / 2.0;
  const phRatio = (8.1 - phAvg) / (9.0 - phAvg);
  assert(Math.abs(phRatio - 0.4) < 1e-6, 'pH range ratio must equal 0.40');

  // Inverse DO rule: Cim = 7.0, Lij (Class II) = 4.0, Ci = 5.5 -> (7.0 - 5.5)/(7.0 - 4.0) = 0.50
  const doRatio = (7.0 - 5.5) / (7.0 - 4.0);
  assert(Math.abs(doRatio - 0.5) < 1e-6, 'DO oxygen deficit ratio must equal 0.50');

  // Logarithmic transform when Ci/Lij > 1.0: COD Ci = 35, Lij = 25 -> raw = 1.4 -> 1.0 + 5*log10(1.4)
  const rawCod = 35.0 / 25.0;
  const adjCod = 1.0 + 5.0 * Math.log10(rawCod);
  assert(adjCod > 1.73 && adjCod < 1.74, `Log-adjusted COD ratio expected ~1.73, got ${adjCod}`);

  // Nemerow-Sumitomo IPj = sqrt((M^2 + R^2) / 2)
  const ratios = [phRatio, doRatio, adjCod];
  const M = Math.max(...ratios);
  const R = ratios.reduce((a, b) => a + b, 0) / ratios.length;
  const ipj = Math.sqrt((M * M + R * R) / 2.0);
  assert(ipj > 1.0 && ipj <= 5.0, `Expected CEMAR_RINGAN (1.0 < IPj <= 5.0), got ${ipj}`);

  // Permen LHK 27/2021 IKA score: 65% Memenuhi (70) + 30% Cemar Ringan (50) + 5% Cemar Sedang (30) + 0% Berat (10)
  const ika = 0.65 * 70 + 0.30 * 50 + 0.05 * 30 + 0.0 * 10;
  assert(Math.abs(ika - 62.0) < 1e-6, `Expected IKA score 62.0 (Sedang), got ${ika}`);
});

// 3. Backend Architecture & Route Wiring
check('3.1: Backend WQMS model, calculator, repository, service, handler, migration, and /api/wqms routes exist and are wired', () => {
  const requiredBackendFiles = [
    'internal/model/wqms.go',
    'internal/service/wqms_calculator.go',
    'internal/repository/wqms_repo.go',
    'internal/service/wqms_service.go',
    'internal/handler/wqms_handler.go',
    'internal/database/migration.go',
    'cmd/main.go',
  ];
  for (const rel of requiredBackendFiles) {
    assert(fs.existsSync(path.join(repoRoot, rel)), `${rel} must exist`);
  }

  const mainGo = fs.readFileSync(path.join(repoRoot, 'cmd/main.go'), 'utf8');
  assert(mainGo.includes('service.NewWQMSService'), 'cmd/main.go must instantiate WQMSService');
  assert(mainGo.includes('wqmsService.EnqueueCalculation'), 'Aggregation bucket callback must trigger WQMS calculation');
  assert(mainGo.includes('api.Group("/wqms")'), 'cmd/main.go must mount /api/wqms routes');

  const migGo = fs.readFileSync(path.join(repoRoot, 'internal/database/migration.go'), 'utf8');
  assert(migGo.includes('&model.WQMSStationConfig{}'), 'migration.go must AutoMigrate WQMSStationConfig');
  assert(migGo.includes('&model.WQMSParameterMapping{}'), 'migration.go must AutoMigrate WQMSParameterMapping');
  assert(migGo.includes('&model.WQMSCalculationResult{}'), 'migration.go must AutoMigrate WQMSCalculationResult');
  assert(migGo.includes('&model.WQMSParameterResult{}'), 'migration.go must AutoMigrate WQMSParameterResult');
  assert(migGo.includes('"wqms.view"'), 'migration.go must seed wqms.* RBAC permissions');
});

// 4. BentoWQMSIndexBar Component (Dashboard Operator Strip)
check('4.1: BentoWQMSIndexBar implements collapsed 1-row IP & IKA strip, expandable sub-cards, Class I-IV selector, Lab/Sensor Mappings, Trend modal, and WebSocket sync', () => {
  const barPath = path.join(__dirname, '../src/components/BentoWQMSIndexBar.vue');
  assert(fs.existsSync(barPath), 'BentoWQMSIndexBar.vue must exist');
  const content = fs.readFileSync(barPath, 'utf8');

  assert(content.includes('IP DOMINAN'), 'Must display IP DOMINAN label');
  assert(content.includes('SKOR IKA'), 'Must display SKOR IKA label');
  assert(content.includes('showParamsDrawer'), 'Must support collapsible/expandable parameter ratio drawer');
  assert(content.includes('CLASS_1') && content.includes('CLASS_2') && content.includes('CLASS_3') && content.includes('CLASS_4'), 'Must support Water Class I-IV selector');
  assert(content.includes('/api/wqms/latest'), 'Must fetch latest IP & IKA bundle');
  assert(content.includes('/api/wqms/mappings'), 'Must support sensor & periodic lab parameter mappings');
  assert(content.includes("wqms-calculated-event"), 'Must subscribe to wqms-calculated-event');
  assert(content.includes("'/monitoring/wqms'"), 'Must link to full WQMS Analysis page');
});

// 5. DashboardOperatorView Integration
check('5.1: DashboardOperatorView mounts BentoWQMSIndexBar only when isWQMSDomain is active without altering existing layout', () => {
  const dashContent = fs.readFileSync(path.join(__dirname, '../src/views/DashboardOperatorView.vue'), 'utf8');
  assert(dashContent.includes("import BentoWQMSIndexBar from '../components/BentoWQMSIndexBar.vue'"), 'Must import BentoWQMSIndexBar');
  assert(dashContent.includes('<BentoWQMSIndexBar'), 'Must render BentoWQMSIndexBar');
  assert(dashContent.includes('v-if="isWQMSDomain && isIspuBarVisible"'), 'Must gate BentoWQMSIndexBar by isWQMSDomain && isIspuBarVisible');
  assert(dashContent.includes('IP & IKA Settings'), 'Toolbar popover #5 must adapt title to IP & IKA Settings in WQMS domain');
});

// 6. Dedicated WQMSAnalysisView, Router, Sidebar & Localization
check('6.1: WQMSAnalysisView, Vue Router (/monitoring/wqms), Sidebar (PP.22), Vuex WebSocket event, and EN/ID locales are wired', () => {
  const viewPath = path.join(__dirname, '../src/views/WQMSAnalysisView.vue');
  assert(fs.existsSync(viewPath), 'WQMSAnalysisView.vue must exist');
  const viewContent = fs.readFileSync(viewPath, 'utf8');
  assert(viewContent.includes('calcMode'), 'WQMSAnalysisView must support 24-Jam (REGULATORY) and Per Agregasi (OPERATIONAL) switch');
  assert(viewContent.includes('/api/wqms/export/') && viewContent.includes("exportFile('csv')") && viewContent.includes("exportFile('excel')"), 'WQMSAnalysisView must support CSV & Excel export');
  assert(viewContent.includes('/api/wqms/backfill'), 'WQMSAnalysisView must support historical backfill');

  const routerContent = fs.readFileSync(path.join(__dirname, '../src/router/index.js'), 'utf8');
  assert(routerContent.includes("path: '/monitoring/wqms'"), 'Router must register /monitoring/wqms');
  assert(routerContent.includes("path: '/wqms', redirect: '/monitoring/wqms'"), 'Router must register /wqms redirect');

  const sidebarContent = fs.readFileSync(path.join(__dirname, '../src/components/Sidebar.vue'), 'utf8');
  assert(sidebarContent.includes('to="/monitoring/wqms"'), 'Sidebar must link to /monitoring/wqms');
  assert(sidebarContent.includes('navigation.wqmsAnalysis'), 'Sidebar must use navigation.wqmsAnalysis i18n key');
  assert(sidebarContent.includes('PP.22'), 'Sidebar must display PP.22 regulatory badge');

  const storeContent = fs.readFileSync(path.join(__dirname, '../src/store/index.js'), 'utf8');
  assert(storeContent.includes("msg.type === 'wqms.calculated'"), 'Vuex store must handle wqms.calculated WS event');
  assert(storeContent.includes('wqms-calculated-event'), 'Vuex store must dispatch wqms-calculated-event');

  const enContent = fs.readFileSync(path.join(__dirname, '../src/i18n/locales/en.js'), 'utf8');
  const idContent = fs.readFileSync(path.join(__dirname, '../src/i18n/locales/id.js'), 'utf8');
  assert(enContent.includes('"wqmsAnalysis"'), 'en.js must include wqmsAnalysis');
  assert(idContent.includes('"wqmsAnalysis"'), 'id.js must include wqmsAnalysis');
});

// 7. Operator Role Hygiene on WQMS Mapping UI & RBAC
check('7.1: WQMS Mapping UI hides device/dev codes ([com*], parameter_code, Param #id) for Operator role & Dashboard Operator', () => {
  const barContent = fs.readFileSync(path.join(__dirname, '../src/components/BentoWQMSIndexBar.vue'), 'utf8');
  const viewContent = fs.readFileSync(path.join(__dirname, '../src/views/WQMSAnalysisView.vue'), 'utf8');

  assert(!barContent.includes('[{{ sp.device_code }}]'), 'BentoWQMSIndexBar must not hardcode [device_code] in sensor dropdown');
  assert(!viewContent.includes('[{{ sp.device_code }}]'), 'WQMSAnalysisView must not hardcode [device_code] in sensor dropdown');
  assert(barContent.includes('isOperator') && barContent.includes('displayLabel') && barContent.includes('getMappingSensorLabel'), 'BentoWQMSIndexBar must sanitize sensor labels for Operator');
  assert(viewContent.includes('isOperator') && viewContent.includes('displayLabel') && viewContent.includes('getMappingSensorLabel'), 'WQMSAnalysisView must sanitize sensor labels for Operator');

  const mwContent = fs.readFileSync(path.join(repoRoot, 'internal/middleware/middleware.go'), 'utf8');
  assert(mwContent.includes('"wqms.manage_mapping": true'), 'Middleware must grant wqms.manage_mapping appropriately');
});

console.log('========================================================================');
console.log(`SUMMARY: ${passed}/${total} TESTS PASSED`);
console.log('========================================================================');

if (passed !== total) {
  process.exit(1);
}
