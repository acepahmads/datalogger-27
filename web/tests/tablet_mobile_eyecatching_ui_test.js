/**
 * Verification Suite: Eye-Catching Tablet & Mobile UI (Fleet Map + Dashboard Operator + Header)
 * Run via: node web/tests/tablet_mobile_eyecatching_ui_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const srcDir = path.join(__dirname, '../src');
const fleetCode = fs.readFileSync(path.join(srcDir, 'views/FleetMapView.vue'), 'utf8');
const heroCode = fs.readFileSync(path.join(srcDir, 'components/BentoISPUHeroCard.vue'), 'utf8');
const dashCode = fs.readFileSync(path.join(srcDir, 'views/DashboardOperatorView.vue'), 'utf8');
const headerCode = fs.readFileSync(path.join(srcDir, 'components/Header.vue'), 'utf8');

let passed = 0;
function check(name, fn) {
  fn();
  passed += 1;
  console.log(`✓ PASS: ${name}`);
}

console.log('========================================================================');
console.log('TABLET & MOBILE EYE-CATCHING UI VERIFICATION SUITE');
console.log('========================================================================');

check('1. FleetMapView eliminates duplicate Status legend and keeps ISPU/IP-IKA legend only in env mode', () => {
  assert(
    !fleetCode.includes("v-if=\"isTilesOnline && pinDisplayMode === 'status'\""),
    'FleetMapView must NOT render duplicate bottom-left Status legend when top KPI pills already act as status legend'
  );
  assert(
    fleetCode.includes('id="fleet-env-index-legend"') && fleetCode.includes("v-if=\"isTilesOnline && pinDisplayMode === 'env'\""),
    'FleetMapView must keep environmental index color scale legend exclusively in ISPU/IP-IKA (env) mode'
  );
});

check('2. FleetMapView implements Map-First Floating Glass HUD, Layer Popover, Slide-Over Drawer & Mobile Bottom Peek', () => {
  assert(fleetCode.includes('isTabletFloatingMode()'), 'FleetMapView must detect tablet / compact landscape floating mode');
  assert(fleetCode.includes('id="fleet-tablet-floating-hud"'), 'FleetMapView must render Tablet Floating Glass HUD');
  assert(fleetCode.includes('id="btn-fleet-tablet-filter"'), 'Tablet HUD must integrate Filter modal button');
  assert(fleetCode.includes('id="btn-fleet-tablet-layers"') && fleetCode.includes('showLayerPopover'), 'Tablet HUD must consolidate map layers into single Layer Popover');
  assert(fleetCode.includes('id="btn-fleet-tablet-drawer-toggle"'), 'Tablet HUD must render floating Daftar Stasiun drawer trigger');
  assert(fleetCode.includes('id="fleet-mobile-floating-chips"') && fleetCode.includes('id="fleet-mobile-bottom-peek"'), 'Mobile Peta mode must float status chips and render rounded bottom peek card');
});

check('3. BentoISPUHeroCard renders inline 7-pollutant vertical mini-bar strip in compact tablet landscape mode', () => {
  assert(heroCode.includes('id="bento-ispu-compact-minibars"'), 'BentoISPUHeroCard compact mode must render inline 7-pollutant mini-bars strip');
  assert(heroCode.includes('.compact-mini-bars-strip'), 'BentoISPUHeroCard must style .compact-mini-bars-strip for tablet landscape');
});

check('4. DashboardOperatorView enables compact horizontal ISPU + collapsible Trend Bar/Sheet + unclipped sensors on Tablet & Mobile', () => {
  assert(dashCode.includes('useCompactTrendSheet()'), 'DashboardOperatorView must share collapsible Trend Bar/Sheet across mobile and compact tablet');
  assert(dashCode.includes('this.bentoViewportWidth < 1400 && this.bentoViewportHeight <= 780'), 'isCompactScreen must activate on short tablet/compact landscape screens');
  assert(dashCode.includes('bento-dashboard-page.has-compact-trend-dock .bento-table-area'), 'Tablet compact trend dock must unclip sensor table/card scroll height');
  assert(dashCode.includes('v-if="useCompactTrendSheet && !mobileTrendSheetOpen"'), 'Docked Analisis Tren pill bar must render for both tablet compact and mobile');
});

check('5. Header hides verbose app title on tablet (< 1280px) to prevent Tenant/Site switcher clipping', () => {
  assert(headerCode.includes('class="hidden xl:flex items-center space-x-2 flex-shrink-0"'), 'Header must hide long title text below xl (1280px) so tablet workspace switchers have room');
  assert(headerCode.includes("!(availableSites.length === 1 && $route.path === '/monitoring/fleet-map')"), 'Header must avoid duplicating identical single-site dropdown while on Fleet Map');
});

console.log('========================================================================');
console.log(`ALL ${passed} TABLET & MOBILE EYE-CATCHING UI TESTS PASSED (100%)`);
console.log('========================================================================');
