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
  assert(viewContent.includes('bento-dual-table-grid'), 'Must render side-by-side dual table grid');
});

check('7.2: Sensor cards, dual tables, Trend Studio, Docked Footer, and all operational modals preserve 100% functionality', () => {
  assert(viewContent.includes('openParamContextMenu'), 'Must support 3-dot & right-click context menu');
  assert(viewContent.includes('togglePin(param.param_id)'), 'Must support favorite pinning');
  assert(viewContent.includes('setChartMode('), 'Trend Studio must support Processed / Raw switch');
  assert(viewContent.includes('ref="scadaChartContainer"') && viewContent.includes('ref="trendPanelSvg"'), 'Trend Studio must preserve SVG chart refs');
  assert(viewContent.includes('@click="openExportModal"') && viewContent.includes('showExportModal'), 'Footer must open Export Data modal');
  assert(viewContent.includes('@click="openReportModal"'), 'Footer must open Operational Report modal');
  assert(viewContent.includes('audio-setup-anchor'), 'Footer must preserve .audio-setup-anchor');
  assert(viewContent.includes('showBrandingModal'), 'Must include Station Branding modal');
  assert(viewContent.includes('showTrendFullscreen'), 'Must include Fullscreen Trend Analyzer modal');
});

// 8. Space-Saving Icon Controls & Adjustable Trend Height (localStorage)
check('8.1: Parameter filter and Card/Table setup use compact icon-only buttons & popovers without horizontal scrollbars', () => {
  assert(viewContent.includes('class="param-ctrl-bar"'), 'Must render compact .param-ctrl-bar');
  assert(!viewContent.includes('class="cat-pills'), 'Must not render wide horizontal .cat-pills row');
  for (const btnId of ['btn-bento-param-filter', 'btn-bento-param-pin', 'btn-bento-view-cards', 'btn-bento-view-cat', 'btn-bento-view-2t', 'btn-bento-view-1t', 'btn-bento-card-setup']) {
    assert(viewContent.includes(`id="${btnId}"`), `Must render compact icon button #${btnId}`);
  }
  assert(viewContent.includes("toggleBentoMiniPopover('paramFilter')"), 'Must wire parameter filter popover');
  assert(viewContent.includes("toggleBentoMiniPopover('cardSetup')"), 'Must wire card & table setup popover');
});

check('8.2: Trend Studio header uses a single compact bar and space-saving Parameter Filter Popover', () => {
  assert(viewContent.includes('class="ts-compact-bar"'), 'Must render single-row .ts-compact-bar');
  assert(!viewContent.includes('class="series-list'), 'Must not render bulky horizontal .series-list chip row');
  assert(viewContent.includes('id="btn-bento-trend-filter"'), 'Must render #btn-bento-trend-filter trigger');
  assert(viewContent.includes("toggleBentoMiniPopover('trendFilter')"), 'Must wire trend parameter filter popover');
  assert(viewContent.includes('class="trend-param-grid custom-scroller"'), 'Must render 2-column compact parameter toggle grid inside popover');
  assert(viewContent.includes('v-model.trim="bentoTrendSearch"'), 'Must support quick parameter search in trend filter popover');
});

check('8.3: Trend Analysis height is user-adjustable via toolbar popover + drag handle and persisted in localStorage', () => {
  assert(viewContent.includes("localStorage.getItem('datalogger_bento_trend_height')"), 'Must read saved trend height from localStorage');
  assert(viewContent.includes("localStorage.setItem('datalogger_bento_trend_height'"), 'Must save trend height to localStorage');
  assert(viewContent.includes(':style="{ height: bentoTrendCanvasHeight + \'px\' }"'), 'Must bind scadaChartContainer height to bentoTrendCanvasHeight');
  assert(viewContent.includes('id="btn-bento-trend-height"') && viewContent.includes('id="input-bento-trend-height"'), 'Must render height adjuster icon button & range slider');
  assert(viewContent.includes('class="trend-resize-handle"') && viewContent.includes('startBentoTrendResize'), 'Must render interactive bottom drag-to-resize handle');
});

// 9. Direct Parameter Limit Integration with Edit Config (ParameterModal)
check('9.1: Parameter cards & tables read configured limits (warning_high, max_value, warning_low, min_value) directly from devicesList with 1-click Edit Config shortcut', () => {
  assert(viewContent.includes('bentoRawParamMap()'), 'Must build reactive bentoRawParamMap from devicesList');
  assert(viewContent.includes('resolveBentoParamLimits(param)'), 'Must resolve configured parameter limits before fallback');
  assert(viewContent.includes('raw.warning_high') && viewContent.includes('raw.max_value') && viewContent.includes('raw.warning_low') && viewContent.includes('raw.min_value'), 'Must read all 4 limit fields saved by ParameterModal');
  assert(viewContent.includes('openBentoQuickLimitEdit(param)'), 'Must support 1-click limit config modal from card/table limit label');
  assert(viewContent.includes('OUT_OF_HARD_RANGE'), 'Must distinguish online OUT_OF_HARD_RANGE limit exceedance from communication offline');
});

// 10. Strict Empty Branding Support & Clean Header (No Default Branding Text or Redundant "Semua Sensor Aktif" Pill)
check('10.1: Header renders only configured Customer & Maker branding and centered clock without default branding or "Semua Sensor Aktif" pill', () => {
  const headerSlice = viewContent.slice(0, viewContent.indexOf('<div class="h-divider">'));
  assert(!headerSlice.includes('PT. Centra Biotech Indonesia'), 'Header must not contain default PT. Centra Biotech Indonesia text');
  assert(!headerSlice.includes('CBI Industrial IoT · Edge Telemetry'), 'Header must not contain default CBI Industrial IoT · Edge Telemetry subtitle');
  assert(!headerSlice.includes('PT. Mitra Industri Enviro'), 'Header must not contain default PT. Mitra Industri Enviro text');
  assert(!headerSlice.includes('online-pill') && !headerSlice.includes('bentoFleetSummary'), 'Header must not render the redundant "Semua Sensor Aktif" online-pill badge');
  assert(headerSlice.includes('v-if="makerCompanyName || makerSubtitle"'), 'Maker info block must be hidden when both name and subtitle are empty');
  assert(headerSlice.includes('v-if="makerCompanyName"') && headerSlice.includes('v-if="makerSubtitle"'), 'Maker name and subtitle must each be conditionally rendered');
  assert(headerSlice.includes('v-if="customerCompanyName"'), 'Customer name must be hidden when empty');
});

// 11. Active Abnormality Alarms (Stuck Value, Spike, Zero Invalid, etc.) Surfaced Directly on Parameter Cards & Tables
check('11.1: Parameter cards & tables display active Stuck Value / abnormality alarms (badge, amber border, amber dot, tooltip, 1-click Acknowledge)', () => {
  assert(viewContent.includes('bentoParamActiveAlarmMap()'), 'Must build reactive bentoParamActiveAlarmMap from activeAlarmsList');
  assert(viewContent.includes('getBentoParamActiveAlarm(param)'), 'Must resolve active alarm details per parameter');
  assert(viewContent.includes("badgeLabel = '⚠ Stuck'"), 'Must display clear "⚠ Stuck" badge for DATA_ABNORMALITY_STUCK_VALUE');
  assert(viewContent.includes("badgeLabel = '⚠ Spike'"), 'Must display "⚠ Spike" badge for DATA_ABNORMALITY_SPIKE');
  assert(viewContent.includes('getBentoDotColor(param)'), 'Must update parameter status dot color to amber/rose when in alarm');
  assert(viewContent.includes('onBentoStatusBadgeClick(param)') && viewContent.includes('this.openAckModal(badge.alarm)'), 'Clicking an active alarm badge on a card/table must open Acknowledge modal');
  assert(viewContent.includes('is-clickable-alarm'), 'Active alarm badge must have interactive clickable styling');
});

// 12. Category-Grouped Side-by-Side Table View (CAT Mode — Up to 3 Compact Columns)
check('12.1: Category Table View (CAT) creates one table per active used category and arranges up to 3 compact columns side-by-side', () => {
  assert(viewContent.includes('id="btn-bento-view-cat"'), 'Must render #btn-bento-view-cat switcher icon button');
  assert(viewContent.includes("@click=\"setBentoDisplayMode('CAT')\""), 'Must support switching to CAT table mode');
  assert(viewContent.includes("savedSplit === 'CAT'"), 'Must restore CAT table mode from localStorage');
  assert(viewContent.includes('bentoCategoryTableSections()'), 'Must compute dynamic category table sections for active used categories');
  assert(viewContent.includes('bentoEffectiveTableGridCols()'), 'Must compute responsive table grid columns up to 3 side-by-side');
  assert(viewContent.includes("'is-tri-cols': bentoEffectiveTableGridCols === 3"), 'Must bind .is-tri-cols class when 3+ category tables are active');
  assert(viewContent.includes('.bento-dual-table-grid.is-tri-cols'), 'Must define CSS repeat(3, minmax(0, 1fr)) for 3 side-by-side category tables');
  assert(viewContent.includes('!col.isCategoryTable && !isCompactScreen && getParamCategory(param)'), 'Must omit redundant per-row category pill when already grouped by category table');
});

check('12.2: Parameter tables enforce strict column alignment (table-layout: fixed, colgroup, decimal-aligned reading + unit slot, dedicated uniform Status & Limit columns)', () => {
  assert(viewContent.includes('<colgroup>') && viewContent.includes('class="bt-col-val"') && viewContent.includes('class="bt-col-status"') && viewContent.includes('class="bt-col-limit"'), 'Must define explicit <colgroup> columns for Parameter, Nilai, Tren, Status, Limit, Update, and Aksi');
  assert(viewContent.includes('table-layout: fixed;'), 'Table must use table-layout: fixed for deterministic column geometry across tables');
  assert(viewContent.includes('class="bt-reading-grid"') && viewContent.includes('class="bt-unit-slot"'), 'Numeric readings must use 2-track grid so decimals align vertically regardless of unit length');
  assert(viewContent.includes('class="s-badge bt-status-pill"'), 'Status badges in tables must use uniform fixed-width .bt-status-pill');
  assert(viewContent.includes('class="bt-limit-bar"'), 'Limit pills must embed subtle progress fill bar without shifting column positions');
});

check('12.3: Parameter tables remain clean and readable on small/narrow screens without header overlap or collapsed parameter names', () => {
  assert(viewContent.includes('ref="bentoParamsWrap"') && viewContent.includes('measureBentoParamsWrap()'), 'Must measure actual .right-params-wrap container width');
  assert(viewContent.includes('new ResizeObserver'), 'Must observe container resize via ResizeObserver for sidebar/split-window changes');
  assert(viewContent.includes('availW < 796') && viewContent.includes('availW >= 1216'), 'Must step down table grid columns cleanly (3 -> 2 -> 1) before Parameter column gets squeezed');
  assert(viewContent.includes("'is-dense-tbl': isBentoDenseTableCard"), 'Must apply .is-dense-tbl compact column sizing when per-table card width < 500px');
  assert(viewContent.includes('min-width: 376px;') && viewContent.includes('min-width: 422px;'), 'Must enforce minimum table width so narrow mobile screens scroll cleanly instead of crushing Parameter column');
  assert(viewContent.includes('min-width: 44px;'), 'Parameter name (.bt-name) must retain minimum width so rows never collapse to colored dots only');
  const theadBlock = viewContent.slice(viewContent.indexOf('.bento-tbl thead th {'), viewContent.indexOf('.bento-tbl thead th {') + 380);
  assert(theadBlock.includes('overflow: hidden;') && theadBlock.includes('text-overflow: ellipsis;'), 'Table headers must clip overflow cleanly so PARAMETER and NILAI can never overlap');
});

check('12.4: Parameter tables and cards expand to full unclipped height when Trend Chart is hidden, with 1-click Hide/Show toggle on Trend Studio', () => {
  assert(viewContent.includes("'is-trend-hidden': !perfSettings.trendAnalysisEnabled"), 'Must bind .is-trend-hidden when trendAnalysisEnabled is false');
  assert(viewContent.includes('.bento-table-area.is-trend-hidden') && viewContent.includes('max-height: none !important;'), 'Must remove max-height clamp on parameter tables when trend chart is hidden so all wrapped tables are visible');
  assert(viewContent.includes('.sensors-scroll-area.is-trend-hidden'), 'Must also remove max-height clamp on parameter cards when trend chart is hidden');
  assert(viewContent.includes('bentoParamsScrollAreaStyle()'), 'Must dynamically scale parameter area max-height to available viewport height when trend chart is open');
  assert(viewContent.includes('id="btn-bento-trend-hide"') && viewContent.includes('id="btn-bento-trend-show"'), 'Must provide direct 1-click Hide & Show Trend buttons right on the Trend Studio bar');
});

// 13. Persistent Audio Mute Enforcement Across localStorage, Performance Settings & GlobalAlarmBanner
check('13.1: Audio Mute state persists reliably in localStorage and is never silently unmuted by applyPerformanceSettings or split user keys', () => {
  const audioAlertContent = fs.readFileSync(path.join(__dirname, '../src/utils/audioAlert.js'), 'utf8');
  const alarmPrefsContent = fs.readFileSync(path.join(__dirname, '../src/utils/alarmPreferences.js'), 'utf8');
  const globalBannerContent = fs.readFileSync(path.join(__dirname, '../src/components/GlobalAlarmBanner.vue'), 'utf8');
  const alarmModalContent = fs.readFileSync(path.join(__dirname, '../src/components/AlarmNotificationModal.vue'), 'utf8');

  assert(alarmPrefsContent.includes("ACTIVE_PREFS_MIRROR_KEY = 'datalogger.alarm_prefs.__active__'"), 'alarmPreferences.js must mirror active alarm preferences to canonical __active__ key to prevent user.id vs username split-brain');
  assert(alarmPrefsContent.includes('audioAlert.setMuteUntil(true, normalized.muteUntil)'), 'alarmPreferences.js must restore exact muteUntil expiration timestamp without resetting timer');
  assert(alarmPrefsContent.includes('applyAlarmPreferences(getStoredAlarmPreferences())'), 'alarmPreferences.js must hydrate singleton audioAlert immediately on module load');

  assert(audioAlertContent.includes('if (!this._isPreviewing && (!this.soundEnabled || this.checkMuteActive())) return;'), 'audioAlert.js playTone/playWarningAlert/playCriticalAlert must strictly block sound synthesis while muted unless previewing');
  assert(audioAlertContent.includes('setMuteUntil(muted, muteUntilTimestamp = null)'), 'audioAlert.js must implement setMuteUntil');

  assert(globalBannerContent.includes('saveAlarmPreferences('), 'GlobalAlarmBanner.vue toggleMute must persist mute state to localStorage');
  assert(alarmModalContent.includes('applyAlarmPreferences(stored)'), 'AlarmNotificationModal.vue close() must restore persisted preferences when dismissed without saving');

  assert(viewContent.includes('applyPerformanceSettings(changedKey = null, changedVal = null)'), 'DashboardOperatorView.vue must override applyPerformanceSettings');
  assert(!viewContent.includes('audioAlert.setMute(!soundOn)'), 'DashboardOperatorView.vue applyPerformanceSettings must NEVER overwrite isMuted with !soundOn');
  assert(viewContent.includes('toggleAudioMute()') && viewContent.includes('onAudioMuteDurationChange()'), 'DashboardOperatorView.vue must persist mute toggle and timer changes atomically to localStorage');
});

// 14. Category-Separated Parameter Cards Setup & Compact Zero-Scroll Optimization
check('14.1: Card Setup supports Category-Separated Parameter Cards (Kartu per Kategori) with collapsible category containers and localStorage persistence', () => {
  assert(viewContent.includes('id="btn-bento-view-cat-cards"'), 'Must render #btn-bento-view-cat-cards icon button in display switcher');
  assert(viewContent.includes("setBentoDisplayMode('cat-cards')") && viewContent.includes('setBentoCardGrouping('), 'Must support switching to category-separated cards from quick bar and Setup Kartu popover');
  assert(viewContent.includes("localStorage.getItem('datalogger_bento_card_group_cat')"), 'Must persist bentoCardGroupByCategory preference in localStorage');
  assert(viewContent.includes('bentoActiveCardSections()'), 'Must compute bentoActiveCardSections reusing active category resolution');
  assert(viewContent.includes('bento-cat-card-group') && viewContent.includes('bento-cat-card-head'), 'Must render dedicated category card group container and header');
  assert(viewContent.includes('toggleBentoCatCardCollapse('), 'Must support collapsing/expanding individual category card sections');
});

check('14.2: Category-Separated Cards use proportional multi-category row packing (up to 8-9 slots) and compact 56px card density to avoid vertical scrolling', () => {
  assert(viewContent.includes('bentoCategorySlotCols()'), 'Must compute responsive bentoCategorySlotCols up to 8-9 slots across');
  assert(viewContent.includes("'--cat-total-cols': bentoCategorySlotCols"), 'Must bind --cat-total-cols CSS variable on .bento-cat-cards-stack');
  assert(viewContent.includes("gridColumn: 'span ' + (sec.colSpan || bentoCategorySlotCols)"), 'Must span category groups proportionally so small categories share rows with larger categories');
  assert(viewContent.includes("gridTemplateColumns: 'repeat(' + (sec.innerCols || 1) + ', minmax(0, 1fr))'"), 'Must pack all cards inside each category group up to innerCols per row');
  assert(viewContent.includes('bentoCardGroupByCategory ? [3, 4, 5, 6, 7, 8] : [1, 2, 3, 4, 5]'), 'Setup Kartu popover must offer 3P..8P high-density column choices in category card mode');
  assert(viewContent.includes('.sensors-grid.is-inside-cat-group .s-card') && viewContent.includes('min-height: 56px;'), 'Must apply compact 56px high-density card styling inside category groups');
});

// 15. Sparkline & Trend Analysis Curve Parity (Adaptive Gap Tolerance + Monotone Cubic Spline)
check('15.1: Parameter card & table sparklines align with Trend Analysis (monotone cubic spline, non-BAD filtering, adaptive gap tolerance for 60s-90s cadence)', () => {
  assert(viewContent.includes('buildBentoSparklineGeometry(rawPoints, options = {})'), 'Must define buildBentoSparklineGeometry in DashboardOperatorView');
  assert(viewContent.includes('recomputeSparklineMeta(paramId)'), 'Must override recomputeSparklineMeta in DashboardOperatorView');
  assert(viewContent.includes('getSparkline(paramId)'), 'Must override getSparkline in DashboardOperatorView');
  assert(viewContent.includes('this.buildSplinePath(seg, smooth)') && viewContent.includes('this.buildSplineArea(seg, H, smooth)'), 'Must render card/table sparklines using the same monotone cubic spline interpolation as Trend Analysis');
  assert(viewContent.includes('medianDt * 3.5'), 'Must scale sparkline gap threshold adaptively by median sampling cadence so 60s-90s samples never fragment into dots');
  assert(viewContent.includes('trendCommDiagnostics()') && viewContent.includes('trendChartOutageMarkers()'), 'Must suppress misleading COMM LOST markers when plotted trend parameters are active via formula fallback');
});

console.log('========================================================================');
console.log(`SUMMARY: ${passed}/${total} TESTS PASSED`);
console.log('========================================================================');

if (passed !== total) {
  process.exit(1);
}
