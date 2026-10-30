/**
 * Acceptance & Boundary Isolation Test Suite:
 * New Menu & Page "Dashboard Operator" (/monitoring/dashboard-operator)
 *
 * Verifies:
 * 1. Strict non-regression: OperatorCommandCenterView.vue & ISPUIndicatorBar.vue are untouched
 * 2. Route '/monitoring/dashboard-operator' (DashboardOperator) registered without altering existing routes
 * 3. Sidebar menu item 'Dashboard Operator' + EN/ID localization
 * 4. App.vue full-bleed layout & fullscreen landing support for '/monitoring/dashboard-operator'
 * 5. BentoISPUHeroCard.vue extends ISPUIndicatorBar with full ISPU logic, validity popover, mappings & trend modals
 * 6. DashboardOperatorView.vue extends OperatorCommandCenterView with Clean Asymmetric Bento UI (Day & Night)
 *    preserving 100% of operational controls, popovers, trend studio, footer, context menus, and modals.
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

console.log('========================================================================');
console.log('RUNNING TEST SUITE: DASHBOARD OPERATOR (CLEAN ASYMMETRIC BENTO)');
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
check('1.1: OperatorCommandCenterView.vue and ISPUIndicatorBar.vue remain 100% untouched in git', () => {
  const diff = execSync(
    'git diff -- web/src/views/OperatorCommandCenterView.vue web/src/components/ISPUIndicatorBar.vue',
    { cwd: repoRoot, encoding: 'utf8' }
  ).trim();
  assert.strictEqual(diff, '', 'Existing OperatorCommandCenterView.vue and ISPUIndicatorBar.vue must have zero git diff');
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
  assert(sidebarContent.includes("navigation.dashboardOperator"), 'Sidebar must reference navigation.dashboardOperator');
  assert(sidebarContent.includes('Dashboard Operator'), 'Sidebar must include Dashboard Operator label');
  assert(enContent.includes('"dashboardOperator": "Dashboard Operator"'), 'en.js must define dashboardOperator');
  assert(idContent.includes('"dashboardOperator": "Dashboard Operator"'), 'id.js must define dashboardOperator');
});

// 4. App.vue Fullscreen Workspace Support
check('4.1: App.vue applies edge-to-edge workspace layout for /monitoring/dashboard-operator', () => {
  assert(appContent.includes('isDashboardOperatorRoute'), 'App.vue must define isDashboardOperatorRoute');
  assert(appContent.includes("this.$route.path === '/monitoring/dashboard-operator'"), 'isDashboardOperatorRoute must match /monitoring/dashboard-operator');
});

// 5. BentoISPUHeroCard Component
check('5.1: BentoISPUHeroCard extends ISPUIndicatorBar and provides ring gauge, 7 bars, validity popover & modals', () => {
  assert(heroContent.includes('extends: ISPUIndicatorBar'), 'BentoISPUHeroCard must extend ISPUIndicatorBar');
  assert(heroContent.includes('PERMEN LHK 14/2020'), 'Must display regulatory standard label');
  assert(heroContent.includes('toggleWindowPopover'), 'Must support 24h window validity popover');
  assert(heroContent.includes('triggerRecalculation'), 'Must support manual ISPU recalculation');
  assert(heroContent.includes('orderedBentoPollutants'), 'Must render all 7 pollutant slots');
  assert(heroContent.includes('showMappingsModal'), 'Must render ISPU Parameter Mappings CRUD modal');
  assert(heroContent.includes('ISPUTrendChart'), 'Must render ISPU Historical Trend modal');
});

// 6. DashboardOperatorView Complete Functional Parity & Bento Design
check('6.1: DashboardOperatorView extends OperatorCommandCenterView with Day/Night theme system', () => {
  assert(viewContent.includes('extends: OperatorCommandCenterView'), 'DashboardOperatorView must extend OperatorCommandCenterView');
  assert(viewContent.includes('BentoISPUHeroCard'), 'Must register and render BentoISPUHeroCard');
  assert(viewContent.includes('is-day') && viewContent.includes('is-night'), 'Must support Day (Siang) and Night (Malam) modes');
  assert(viewContent.includes("setBentoTheme('light')") && viewContent.includes("setBentoTheme('dark')"), 'Must wire Siang/Malam buttons to setBentoTheme');
});

check('6.2: Header workspace bar preserves all 8 interactive popovers and click-outside anchors', () => {
  assert(viewContent.includes('toolbar-popover-anchor'), 'Must preserve .toolbar-popover-anchor for click-outside dismissal');
  for (const pop of ['viewOptions', 'performanceSettings', 'ispuSettings', 'actuators', 'healthAndAlarms', 'systemMonitor', 'networkMonitor', 'stationLocation']) {
    assert(viewContent.includes(`toggleToolbarPopover('${pop}')`), `Must wire ${pop} toolbar popover`);
  }
});

check('6.3: Asymmetric Bento sensor grid & Trend Studio preserve all controls, raw/processed toggle, and dual-axis chart', () => {
  assert(viewContent.includes('openParamContextMenu'), 'Sensor cards must support 3-dot & right-click context menu');
  assert(viewContent.includes('togglePin(param.param_id)'), 'Sensor cards must support favorite pinning');
  assert(viewContent.includes("setChartMode('processed')") && viewContent.includes("setChartMode('raw')"), 'Trend Studio must support 1-click Processed / Raw Sensor switch');
  assert(viewContent.includes('trend-download-anchor'), 'Trend Studio must preserve .trend-download-anchor');
  assert(viewContent.includes('trend-settings-anchor'), 'Trend Studio must preserve .trend-settings-anchor');
  assert(viewContent.includes('trend-line-smooth-anchor'), 'Trend Studio must preserve .trend-line-smooth-anchor');
  assert(viewContent.includes('ref="scadaChartContainer"') && viewContent.includes('ref="trendPanelSvg"'), 'Trend Studio must preserve SVG chart refs');
});

check('6.4: Docked Footer & all operational modals are preserved', () => {
  assert(viewContent.includes('@click="openExportModal"') && viewContent.includes('showExportModal'), 'Footer must open Export Data modal');
  assert(viewContent.includes('@click="openReportModal"'), 'Footer must open Operational Report modal');
  assert(viewContent.includes('audio-setup-anchor'), 'Footer must preserve .audio-setup-anchor for sound setup');
  assert(viewContent.includes('showBrandingModal'), 'Must include Station Branding modal');
  assert(viewContent.includes('showTrendFullscreen'), 'Must include Fullscreen Trend Analyzer modal');
  assert(viewContent.includes('CategoryManagementModal'), 'Must include CategoryManagementModal');
  assert(viewContent.includes('AlarmNotificationModal'), 'Must include AlarmNotificationModal');
});

console.log('========================================================================');
console.log(`SUMMARY: ${passed}/${total} TESTS PASSED`);
console.log('========================================================================');

if (passed !== total) {
  process.exit(1);
}
