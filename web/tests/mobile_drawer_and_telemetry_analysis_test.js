/**
 * Automated Verification Suite: Mobile Slide-Over Menu Drawer & Symmetrical Telemetry Analysis (< 768px)
 * Run via: node web/tests/mobile_drawer_and_telemetry_analysis_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const appCode = fs.readFileSync(path.join(__dirname, '../src/App.vue'), 'utf8');
const sidebarCode = fs.readFileSync(path.join(__dirname, '../src/components/Sidebar.vue'), 'utf8');
const headerCode = fs.readFileSync(path.join(__dirname, '../src/components/Header.vue'), 'utf8');
const taCode = fs.readFileSync(path.join(__dirname, '../src/views/TelemetryAnalysisView.vue'), 'utf8');

let total = 0;
let passed = 0;

function check(name, fn) {
  total++;
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    passed++;
  } catch (err) {
    console.error(`✗ FAIL: ${name}\n  ${err.message}`);
  }
}

console.log('========================================================================');
console.log('RUNNING TEST SUITE: MOBILE MENU DRAWER & TELEMETRY ANALYSIS (< 768px)');
console.log('========================================================================');

check('1.1: App.vue renders Sidebar as full-height z-[70] overlay drawer on < 768px instead of inline flex column', () => {
  assert(appCode.includes('isMobileViewport()') && appCode.includes('this.viewportWidth < 768'), 'App.vue must detect mobile viewport (< 768px)');
  assert(appCode.includes('isMobileDrawerOpen'), 'App.vue must track isMobileDrawerOpen state');
  assert(appCode.includes('id="mobile-sidebar-drawer"') && appCode.includes('z-[70]'), 'Mobile drawer must render at z-[70] above bottom docked bars (z-46)');
  assert(appCode.includes('z-[65]') && appCode.includes('@click="closeMobileDrawer"'), 'Mobile backdrop must render at z-[65] and close drawer on tap');
  assert(appCode.includes('<div v-if="!sidebarAutoHide" class="hidden md:flex h-full flex-shrink-0">'), 'Inline pinned sidebar must be hidden on < 768px (hidden md:flex) so <main> never gets crushed');
  assert(/if \(this\.isMobileViewport\)\s*\{\s*this\.isMobileDrawerOpen = !this\.isMobileDrawerOpen;/.test(appCode), 'toggleSidebar() on < 768px must toggle mobile overlay drawer cleanly');
});

check('1.2: Sidebar.vue supports isMobileDrawer mode with dedicated Close (X) button while keeping desktop Auto-Hide intact', () => {
  assert(sidebarCode.includes('isMobileDrawer: {'), 'Sidebar.vue must accept isMobileDrawer prop');
  assert(sidebarCode.includes('id="btn-sidebar-mobile-close"') && sidebarCode.includes("$emit('close-mobile')"), 'Sidebar.vue must render mobile close button emitting close-mobile');
  assert(sidebarCode.includes('v-if="!isMobileDrawer"') && sidebarCode.includes('id="btn-sidebar-autohide-setup"'), 'Desktop auto-hide setup button must stay intact when !isMobileDrawer');
});

check('2.1: Header.vue provides mobile Hamburger trigger and single-pill station switcher with zero horizontal overflow', () => {
  assert(headerCode.includes('id="btn-mobile-hamburger"') && headerCode.includes('md:hidden'), 'Header.vue must expose md:hidden mobile hamburger button');
  assert(headerCode.includes(":class=\"availableSites.length > 0 ? 'hidden md:flex' : 'flex'\""), 'Header.vue must hide static tenant badge on mobile when site switcher is shown');
  assert(headerCode.includes('id="header-site-switcher"') && headerCode.includes('min-w-0 flex-1 sm:flex-initial'), 'Site switcher must shrink gracefully with min-w-0 flex-1 on mobile');
  assert(headerCode.includes('sm:hidden w-full text-left px-3 py-2') && headerCode.includes('ID • Indonesia'), 'User dropdown must expose mobile language switch row on < 640px');
});

check('3.1: TelemetryAnalysisView.vue unifies top controls into symmetrical rows and 2-col filter grid on < 768px', () => {
  assert(taCode.includes('@media (max-width: 767px)'), 'TelemetryAnalysisView.vue must scope mobile layout inside @media (max-width: 767px)');
  const mobileCss = taCode.slice(taCode.lastIndexOf('@media (max-width: 767px)'));
  assert(mobileCss.includes('.ta-bar1-card + .ta-bar2-card'), 'Mobile CSS must seamlessly join Bar 1 and Bar 2 into one unified card');
  assert(mobileCss.includes('.ta-filter-grid') && mobileCss.includes('grid-template-columns: repeat(2, minmax(0, 1fr))'), 'Mobile filters must form a symmetrical 2-column grid');
  assert(mobileCss.includes('.ta-filter-grid > select:last-child:nth-child(odd)'), 'Odd trailing filter select must span full width so no half-empty ragged row appears');
  assert(taCode.includes('sm:hidden text-slate-400 dark:text-slate-600">&middot;</span>'), 'Mobile KPI strip must use clean middle-dot separator without hanging pipe');
});

check('3.2: TelemetryAnalysisView.vue keeps TELEMETRY TREND above-the-fold with touch scrubbing and interactive series chips', () => {
  assert(taCode.includes("h-[215px] sm:h-60"), 'Trend chart must use proportional 215px height on mobile while keeping h-60 on desktop');
  assert(taCode.includes('@touchstart.passive="onChartTouchMove"') && taCode.includes('@touchmove.passive="onChartTouchMove"'), 'Trend SVG must support touch scrubbing on mobile');
  assert(taCode.includes('onChartTouchMove(e)') && taCode.includes('isTouch: true'), 'onChartTouchMove must translate touch coordinates for tooltip scrubbing');
  assert(taCode.includes("mob-chip-"), 'Mobile chart card must render interactive parameter legend chips');
});

console.log('========================================================================');
console.log(`SUMMARY: ${passed}/${total} TESTS PASSED`);
console.log('========================================================================');

if (passed !== total) {
  process.exit(1);
}
