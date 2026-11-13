/**
 * Automated Verification Suite: Mobile (< 768px) Redesign for ISPU Analysis & WQMS (IP & IKA) Pages
 *
 * Verifies:
 * 1. Header.vue narrow-phone anti-clip layout (px-2 sm:px-4, max-w-[140px] sm:max-w-[215px]).
 * 2. ISPUTrendChart.vue mobile overflow fix (w-full sm:w-auto max-w-full min-w-0 overflow-x-auto no-scrollbar),
 *    symmetrical 3-col range switcher, conditional reset hover button, and touch scrubbing (@touchstart.passive / @touchmove.passive).
 * 3. ISPUAnalysisView.vue unified mobile control card (#ispu-mobile-control-card), smart back route,
 *    human-readable category/pollutant labels, 5-zone ISPU scale bar, 2-column Bento 7-parameter grid,
 *    mobile audit history cards, and 100% preserved desktop (>= 768px) layout.
 * 4. WQMSAnalysisView.vue unified mobile control card (#wqms-mobile-control-card), symmetrical 4-col actions,
 *    2x2 mini-Bento IPj KPI grid, compact side-by-side IKA ring gauge, mobile parameter evaluation cards,
 *    responsive touch-scrubbable IP trend chart (no forced min-w-[600px]), and mobile history cards.
 * 5. Zero git diff on protected files.
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const repoRoot = path.resolve(__dirname, '../..');
const headerCode = fs.readFileSync(path.join(__dirname, '../src/components/Header.vue'), 'utf8');
const ispuChartCode = fs.readFileSync(path.join(__dirname, '../src/components/ISPUTrendChart.vue'), 'utf8');
const ispuViewCode = fs.readFileSync(path.join(__dirname, '../src/views/ISPUAnalysisView.vue'), 'utf8');
const wqmsViewCode = fs.readFileSync(path.join(__dirname, '../src/views/WQMSAnalysisView.vue'), 'utf8');

let total = 0;
let passed = 0;

function check(name, fn) {
  total++;
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    passed++;
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  ${err.message}`);
  }
}

console.log('========================================================================');
console.log('RUNNING SUITE: MOBILE ISPU & WQMS (IP & IKA) REDESIGN VERIFICATION');
console.log('========================================================================');

check('1.1: Header.vue prevents right-edge avatar clipping on narrow 320px-360px phones', () => {
  assert(headerCode.includes('px-2 sm:px-4'), 'Header must use compact px-2 on mobile and sm:px-4 on tablet/desktop');
  assert(headerCode.includes('max-w-[140px] sm:max-w-[215px]'), 'Site switcher must clamp to 140px on narrow phones');
});

check('2.1: ISPUTrendChart.vue fixes pollutant capsule overflow & adds touch scrubbing', () => {
  assert(
    ispuChartCode.includes('w-full sm:w-auto max-w-full min-w-0 overflow-x-auto no-scrollbar'),
    'Pollutant capsule strip must constrain width with w-full max-w-full min-w-0 overflow-x-auto'
  );
  assert(ispuChartCode.includes('grid grid-cols-3 sm:flex'), 'Time range switcher must use 3 equal columns on mobile');
  assert(ispuChartCode.includes('@touchstart.passive="handleTouchMove"'), 'Chart must support @touchstart.passive');
  assert(ispuChartCode.includes('@touchmove.passive="handleTouchMove"'), 'Chart must support @touchmove.passive');
  assert(ispuChartCode.includes('handleTouchMove(e)'), 'Chart must implement handleTouchMove method');
});

check('3.1: ISPUAnalysisView.vue provides unified mobile control card, 5-zone scale, 2-col Bento grid & mobile audit cards', () => {
  assert(ispuViewCode.includes('id="ispu-mobile-control-card"'), 'Must render #ispu-mobile-control-card for < 768px');
  assert(ispuViewCode.includes('hidden md:flex'), 'Desktop header must remain preserved via hidden md:flex');
  assert(ispuViewCode.includes('backRoutePath'), 'Back link must route Operators cleanly to /monitoring/dashboard-operator');
  assert(ispuViewCode.includes('formatCategoryLabel('), 'Hero & history must display human-readable category labels without raw underscores');
  assert(ispuViewCode.includes('ispuScaleMarkerPct'), 'Hero card must display 5-zone ISPU spectrum scale marker');
  assert(ispuViewCode.includes('@pollutant-change="selectPollutant"'), 'ISPUTrendChart pollutant selection must sync with parent view');
  assert(ispuViewCode.includes('grid-cols-2 sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-7'), '7 official parameters must render in a 2-col Bento grid on mobile');
  assert(ispuViewCode.includes('md:hidden space-y-2'), 'Must provide mobile ISPU audit history cards');
  assert(ispuViewCode.includes('hidden md:block overflow-x-auto'), 'Desktop 16-col history table must remain preserved on >= 768px');
});

check('4.1: WQMSAnalysisView.vue provides unified mobile control card, compact IP/IKA hero cards, parameter cards & touch IP chart', () => {
  assert(wqmsViewCode.includes('id="wqms-mobile-control-card"'), 'Must render #wqms-mobile-control-card for < 768px');
  assert(wqmsViewCode.includes('mobileActionGridCols'), 'Mobile action buttons must sit in symmetrical equal columns');
  assert(wqmsViewCode.includes('ipCategoryCleanLabel'), 'Mobile IP badge must render clean category label');
  assert(!wqmsViewCode.includes('min-w-[600px]'), 'WQMS IP trend chart must not force horizontal scrolling via min-w-[600px]');
  assert(wqmsViewCode.includes('@touchstart.passive="handleIpChartTouch"'), 'WQMS IP trend chart must support touch scrubbing');
  assert(wqmsViewCode.includes('hoveredIpPoint'), 'WQMS IP trend chart must provide interactive floating tooltip');
  assert(wqmsViewCode.includes('hidden md:block overflow-x-auto'), 'Desktop WQMS tables must remain preserved on >= 768px');
});

check('5.1: Protected legacy files remain 100% untouched in git diff and DashboardOperatorView.vue is un-truncated', () => {
  const protectedFiles = [
    'web/src/views/OperatorCommandCenterView.vue',
    'web/src/components/ISPUIndicatorBar.vue',
    'web/src/components/OperatorDashboardIcon.vue',
  ];
  for (const rel of protectedFiles) {
    const out = execSync(`git diff -- "${rel}"`, { cwd: repoRoot, encoding: 'utf8' }).trim();
    assert.strictEqual(out, '', `Protected file ${rel} must not be modified`);
  }
  const dashOpStat = fs.statSync(path.join(repoRoot, 'web/src/views/DashboardOperatorView.vue'));
  assert(dashOpStat.size > 570000, 'DashboardOperatorView.vue must remain full size (> 570 KB)');
});

console.log('========================================================================');
console.log(`SUMMARY: ${passed}/${total} TESTS PASSED`);
console.log('========================================================================');

if (passed !== total) {
  process.exit(1);
}

