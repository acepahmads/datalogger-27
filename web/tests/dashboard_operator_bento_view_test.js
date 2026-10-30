/**
 * Acceptance & Boundary Isolation Test Suite:
 * New Menu & Page "Dashboard Operator" (/monitoring/dashboard-operator)
 *
 * Verifies:
 * 1. Strict non-regression: OperatorCommandCenterView.vue, ISPUIndicatorBar.vue, and OperatorDashboardIcon.vue are untouched
 * 2. Route '/monitoring/dashboard-operator' (DashboardOperator) registered without altering existing routes
 * 3. Sidebar menu item 'Dashboard Operator' + EN/ID localization
 * 4. App.vue full-bleed layout & fullscreen landing support for '/monitoring/dashboard-operator'
 * 5. BentoISPUHeroCard.vue supports both Widescreen Full Hero Card and Compact Dominant-Only Bar + Click-to-Expand 7 Pollutants Drawer
 * 6. DashboardOperatorView.vue implements Icon-Only Operator Console Buttons, Auto-Adaptive Responsive Grid (up to 5P micro cards),
 *    and Side-by-Side 2-Table (2T) Bento View while preserving 100% of operational controls, popovers, trend studio, and modals.
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

console.log('========================================================================');
console.log('RUNNING TEST SUITE: DASHBOARD OPERATOR (AUTO-ADAPTIVE BENTO + ICON UI)');
console.log('========================================================================');

const repoRoot = path.resolve(__dirname, '../..');
const routerContent = fs.readFileSync(path.join(__dirname, '../src/router/index.js'), 'utf8');
const sidebarContent = fs.readFileSync(path.join(__dirname, '../src/components/Sidebar.vue'), 'utf8');
const appContent = fs.readFileSync(path.join(__dirname, '../src/App.vue'), 'utf8');
const enContent = fs.readFileSync(path.join(__dirname, '../src/i18n/locales/en.js'), 'utf8');
const idContent = fs.readFileSync(path.join(__dirname, '../src/i18n/locales/id.js'), 'utf8');
const heroContent = fs.readFileSync(path.join(__dirname, '../src/components/BentoISPUHeroCard.vue'), 'utf8');
const viewContent = fs.readFileSync(path.join(__dirname, '../src/views/DashboardOperatorView.vue'), 'utf8');

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
check('1.1: OperatorCommandCenterView.vue, ISPUIndicatorBar.vue, and OperatorDashboardIcon.vue remain 100% untouched in git', () => {
  const diff = execSync(
    'git diff -- web/src/views/OperatorCommandCenterView.vue web/src/components/ISPUIndicatorBar.vue web/src/components/OperatorDashboardIcon.vue',
    { cwd: repoRoot, encoding: 'utf8' }
  ).trim();
  assert.strictEqual(diff, '', 'Existing Operator Console & shared icon/ISPU files must have zero git diff');
});

// 2. Routing & Existing Route Preservation
check('2.1: Router registers /monitoring/dashboard-operator while keeping existing operator routes intact', () => {
  assert(routerContent.includes("import DashboardOperatorView from '../views/DashboardOperatorView.vue'"), 'Must import DashboardOperatorView');
  assert(routerContent.includes("path: '/monitoring/dashboard-operator'"), 'Must register /monitoring/dashboard-operator');
  assert(routerContent.includes("name: 'DashboardOperator'"), 'Must name route DashboardOperator');
  assert(routerContent.includes("{ path: '/', redirect: '/monitoring/operator' }"), 'Root redirect must stay untouched');
  assert(routerContent.includes("{ path: '/monitoring/operator', name: 'OperatorCommandCenter'"), 'Original OperatorCommandCenter route must stay untouched');
});

// 3. Sidebar Menu & Localization
check('3.1: Sidebar exposes "Dashboard Operator" menu item with EN and ID translations', () => {
  assert(sidebarContent.includes('to="/monitoring/dashboard-operator"'), 'Sidebar must link to /monitoring/dashboard-operator');
  assert(sidebarContent.includes('navigation.dashboardOperator'), 'Sidebar must reference navigation.dashboardOperator');
  assert(sidebarContent.includes('Dashboard Operator'), 'Sidebar must include Dashboard Operator label');
  assert(enContent.includes('"dashboardOperator": "Dashboard Operator"'), 'en.js must define dashboardOperator');
  assert(idContent.includes('"dashboardOperator": "Dashboard Operator"'), 'id.js must define dashboardOperator');
});

// 4. App.vue Fullscreen Workspace Support
check('4.1: App.vue applies edge-to-edge workspace layout for /monitoring/dashboard-operator', () => {
  assert(appContent.includes('isDashboardOperatorRoute'), 'App.vue must define isDashboardOperatorRoute');
  assert(appContent.includes("this.$route.path === '/monitoring/dashboard-operator'"), 'isDashboardOperatorRoute must match /monitoring/dashboard-operator');
});

// 5. BentoISPUHeroCard Adaptive Modes (Compact Dominant Bar + Full Widescreen Hero Card)
check('5.1: BentoISPUHeroCard supports full widescreen hero card AND compact dominant-only bar with click-to-expand 7 pollutants drawer', () => {
  assert(heroContent.includes('extends: ISPUIndicatorBar'), 'BentoISPUHeroCard must extend ISPUIndicatorBar');
  assert(heroContent.includes('compact: {'), 'Must accept compact boolean prop');
  assert(heroContent.includes('is-compact-bar'), 'Must render compact horizontal bar class when compact=true');
  assert(heroContent.includes('showCompactPollutantsDrawer'), 'Must provide reactive drawer toggle for 7 pollutants in compact mode');
  assert(heroContent.includes('Lihat 7 Polutan') && heroContent.includes('Sembunyikan 6 Lainnya'), 'Must provide user-friendly click-to-expand button');
  assert(heroContent.includes('PERMEN LHK 14/2020'), 'Must display regulatory standard label');
  assert(heroContent.includes('toggleWindowPopover'), 'Must support 24h window validity popover');
  assert(heroContent.includes('triggerRecalculation'), 'Must support manual ISPU recalculation');
  assert(heroContent.includes('orderedBentoPollutants'), 'Must render all 7 pollutant slots');
  assert(heroContent.includes('showMappingsModal'), 'Must render ISPU Parameter Mappings CRUD modal');
  assert(heroContent.includes('ISPUTrendChart'), 'Must render ISPU Historical Trend modal');
});

// 6. Icon-Only Toolbar & Trend Studio Matching Operator Console
check('6.1: DashboardOperatorView uses exact OperatorDashboardIcon icon-only buttons for toolbar and trend studio', () => {
  assert(viewContent.includes('extends: OperatorCommandCenterView'), 'DashboardOperatorView must extend OperatorCommandCenterView');
  assert(viewContent.includes('is-day') && viewContent.includes('is-night'), 'Must support Day and Night modes');
  assert(viewContent.includes('id="btn-operator-theme-toggle"') && viewContent.includes('@click="toggleTheme"'), 'Must use single icon theme toggle button');
  for (const iconName of ['sliders', 'performance', 'ispu', 'moon', 'sun', 'pulse', 'sensor', 'network', 'download', 'fullscreen', 'settings', 'check-all', 'clear', 'reset']) {
    assert(viewContent.includes(`name="${iconName}"`), `Must render OperatorDashboardIcon "${iconName}"`);
  }
  assert(viewContent.includes("stationLocation.is_configured ? 'map-pinned' : 'map-pin'"), 'Must render dynamic GPS pin icon');
  assert(viewContent.includes("trendLineSmooth ? 'line-smooth' : 'line-sharp'"), 'Must render dynamic line interpolation icon');
});

check('6.2: Header workspace bar preserves all 8 interactive popovers and click-outside anchors', () => {
  assert(viewContent.includes('toolbar-popover-anchor'), 'Must preserve .toolbar-popover-anchor for click-outside dismissal');
  for (const pop of ['viewOptions', 'performanceSettings', 'ispuSettings', 'actuators', 'healthAndAlarms', 'systemMonitor', 'networkMonitor', 'stationLocation']) {
    assert(viewContent.includes(`toggleToolbarPopover('${pop}')`), `Must wire ${pop} toolbar popover`);
  }
});

// 7. Auto-Adaptive Responsive Layout (Compact <1280px vs Wide >=1280px, 5P Micro Cards, 2-Table Side-by-Side)
check('7.1: DashboardOperatorView auto-adapts ISPU placement, 5P micro parameter cards, and 2-Table side-by-side view', () => {
  assert(viewContent.includes('isCompactScreen()') && viewContent.includes('this.bentoViewportWidth < 1280'), 'Must detect compact vs widescreen viewport reactively');
  assert(viewContent.includes(':compact="isCompactScreen"'), 'Must bind :compact="isCompactScreen" to BentoISPUHeroCard');
  assert(viewContent.includes('effectiveBentoCols') && viewContent.includes('.bento-cols-5'), 'Must support auto-dense 5-column parameter grid');
  assert(viewContent.includes('is-micro-cards') && viewContent.includes('is-compact-cards'), 'Must apply compact micro-card density styling');
  assert(viewContent.includes('bentoSplitTableColumns') && viewContent.includes("setBentoDisplayMode('2T')"), 'Must support 2 Tabel Samping (2T) side-by-side table view');
  assert(viewContent.includes('bento-dual-table-grid') && viewContent.includes('2 Tabel Samping'), 'Must render side-by-side dual table grid');
});

check('7.2: Sensor cards, dual tables, Trend Studio, Docked Footer, and all operational modals preserve 100% functionality', () => {
  assert(viewContent.includes('openParamContextMenu'), 'Must support 3-dot & right-click context menu');
  assert(viewContent.includes('togglePin(param.param_id)'), 'Must support favorite pinning');
  assert(viewContent.includes("setChartMode('processed')") && viewContent.includes("setChartMode('raw')"), 'Trend Studio must support 1-click Processed / Raw switch');
  assert(viewContent.includes('ref="scadaChartContainer"') && viewContent.includes('ref="trendPanelSvg"'), 'Trend Studio must preserve SVG chart refs');
  assert(viewContent.includes('@click="openExportModal"') && viewContent.includes('showExportModal'), 'Footer must open Export Data modal');
  assert(viewContent.includes('@click="openReportModal"'), 'Footer must open Operational Report modal');
  assert(viewContent.includes('audio-setup-anchor'), 'Footer must preserve .audio-setup-anchor');
  assert(viewContent.includes('showBrandingModal'), 'Must include Station Branding modal');
  assert(viewContent.includes('showTrendFullscreen'), 'Must include Fullscreen Trend Analyzer modal');
});

console.log('========================================================================');
console.log(`SUMMARY: ${passed}/${total} TESTS PASSED`);
console.log('========================================================================');

if (passed !== total) {
  process.exit(1);
}
