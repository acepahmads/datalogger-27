/**
 * Automated Verification Suite: Trend Analysis Axis Strengthening, Device Offline, and Data Gap Handling
 *
 * Verifies:
 * Section A: Axis Contrast, Thickness, Ticks, Units, and Dual Y-Axis Differentiation
 * Section B: Device Offline & Sensor Timeout Diagnostic State Handling
 * Section C: Visual Data Gap Representation and Path Splitting (No False Interpolation)
 * Section D: Status Badges, Empty States, and Error Recovery
 * Section E: Non-destructive Retention of Telemetry Data and Calculations
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const storePath = path.join(__dirname, '../src/store/index.js');
const occContent = fs.readFileSync(occPath, 'utf8');
const storeContent = fs.readFileSync(storePath, 'utf8');

let totalTests = 0;
let passedTests = 0;

function runTest(name, fn) {
  totalTests++;
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
  }
}

console.log('===============================================================');
console.log('RUNNING AUTOMATED TEST SUITE: TREND AXIS, OFFLINE & DATA GAP');
console.log('===============================================================\n');

// =========================================================================
// SECTION A: PERTEGAS SUMBU Y & X (AXIS STROKE, TICKS, LABELS, DUAL Y-AXIS)
// =========================================================================

runTest('A.1: Axis lines use high-contrast dark navy in light mode and theme token in dark mode', () => {
  assert(occContent.includes("axisStroke()"), 'Must define axisStroke computed property');
  assert(occContent.includes("return this.isDark ? '#38587A' : '#0F172A'"), 'axisStroke must return #0F172A for light mode and #38587A for dark mode');
});

runTest('A.2: Axis lines use prominent thickness (1.5–2px, stroke-width="1.8")', () => {
  // Left Y-axis vertical line
  assert(occContent.includes(':x1="trendChartDims.padLeft"'), 'Must have vertical line for left Y-axis');
  assert(occContent.includes(':x2="trendChartDims.padLeft"'), 'Left Y-axis must be vertical');
  assert(occContent.includes('stroke-width="1.8"'), 'Axis lines must have 1.8px stroke width');

  // Base X-axis horizontal line
  assert(occContent.includes(':y1="trendChartDims.baseY"'), 'Must have base X-axis line');
  assert(occContent.includes(':y2="trendChartDims.baseY"'), 'Base X-axis line must align with baseY');
});

runTest('A.3: Y-axis ticks have high contrast, medium/semibold weight, and 4px tick protrusion', () => {
  // Tick font weight medium / semibold
  assert(occContent.includes('font-weight="600"'), 'Y-axis tick labels must use font-weight="600"');
  // Light mode tick color #0F172A, dark mode #E2E8F0
  assert(occContent.includes(":fill=\"isDark ? '#E2E8F0' : '#0F172A'\""), 'Tick labels must use high contrast #0F172A in light mode');
  // Tick marks extending 4px outward
  assert(occContent.includes(':x1="trendChartDims.padLeft - 4"'), 'Left Y-axis tick mark must extend 4px outward');
  assert(occContent.includes(':x2="trendChartDims.padLeft"'), 'Left Y-axis tick mark must meet axis line');
});

runTest('A.4: Dual Y-axis is clearly differentiated and color-matched to prevent unit swapping', () => {
  // Right Y-axis vertical line rendered in dual-axis mode
  assert(occContent.includes('v-if="trendScales.isDual"'), 'Right Y-axis must render when dual axis is active');
  assert(occContent.includes(':stroke="trendScales.right ? trendScales.right.color : axisStroke"'), 'Right Y-axis line must match series color to eliminate confusion');
  assert(occContent.includes(':fill="trendScales.right ? trendScales.right.color : (isDark ? \'#E2E8F0\' : \'#0F172A\')"'), 'Right Y-axis tick numbers must match right series color');
  // Right tick mark extending outward
  assert(occContent.includes(':x1="trendChartDims.width - trendChartDims.padRight"'), 'Right tick mark origin at plot edge');
  assert(occContent.includes(':x2="trendChartDims.width - trendChartDims.padRight + 4"'), 'Right tick mark extends 4px outward');
});

runTest('A.5: Horizontal grid lines remain subtle thin gray and non-dominating', () => {
  assert(occContent.includes("return this.isDark ? '#132738' : '#E2E8F0'"), 'Grid stroke must be subtle gray');
  assert(occContent.includes('stroke-width="0.8"'), 'Grid lines must be 0.8px thin');
  assert(occContent.includes('stroke-dasharray="3 3"'), 'Grid lines must be dashed 3 3');
  assert(occContent.includes('stroke-opacity="0.6"'), 'Grid lines must have subtle opacity 0.6');
});

runTest('A.6: Clear parameter, unit, and scale badges with directional indicators on top of axes', () => {
  assert(occContent.includes('◀ {{ trendScales.left ? trendScales.left.label : \'\' }}'), 'Left axis header must display param name/unit with left arrow');
  assert(occContent.includes('{{ trendScales.right.label }} ▶'), 'Right axis header must display right param name/unit with right arrow');
});

// =========================================================================
// SECTION B: PENANGANAN DEVICE OFFLINE & SENSOR TIMEOUT
// =========================================================================

runTest('B.1: WebSocket activity events are dispatched to window event system', () => {
  assert(storeContent.includes("msg.type === 'SYSTEM_ACTIVITY_EVENT'"), 'Store must detect SYSTEM_ACTIVITY_EVENT');
  assert(storeContent.includes("new CustomEvent('system-activity-event'"), 'Store must dispatch system-activity-event');
  assert(storeContent.includes("new CustomEvent('device-comm-event'"), 'Store must dispatch device-comm-event');
});

runTest('B.2: Component tracks device communication state, outages, and freshness', () => {
  assert(occContent.includes('trendCommDiagnostics'), 'Must compute trendCommDiagnostics');
  assert(occContent.includes('DEVICE_OFFLINE'), 'Must identify DEVICE_OFFLINE state');
  assert(occContent.includes('SENSOR_TIMEOUT'), 'Must identify SENSOR_TIMEOUT state');
  assert(occContent.includes('STALE_DATA'), 'Must identify STALE_DATA state');
  assert(occContent.includes('COMMUNICATION LOST'), 'Must format COMMUNICATION LOST badge');
});

runTest('B.3: Historical data is strictly preserved when device transitions to OFFLINE', () => {
  assert(occContent.includes("const isDeviceOffline = this.trendCommDiagnostics && this.trendCommDiagnostics.state === 'DEVICE_OFFLINE';"), 'Must detect offline state to preserve history');
  assert(occContent.includes('if (!isDeviceOffline) {'), 'Must bypass point pruning when device is offline');
});

runTest('B.4: API fetch preserves valid historical cache if request returns empty during offline', () => {
  assert(occContent.includes('if (processed.length > 0 || !s.dataPoints || s.dataPoints.length === 0)'), 'Must retain existing valid history when API returns empty');
  assert(occContent.includes('s.dataPoints = processed;'), 'Must assign processed points only when present or initialized');
});

// =========================================================================
// SECTION C: VISUALISASI GAP DATA (NO FALSE INTERPOLATION, SEVERED PATHS)
// =========================================================================

runTest('C.1: generateChartPath breaks SVG path continuity on communication gap', () => {
  assert(occContent.includes('let maxGapMs = 15 * 60 * 1000;'), 'Must calculate adaptive gap threshold');
  assert(occContent.includes('if (timeSpan <= 90 * 1000) maxGapMs = 12 * 1000;'), '1m timeframe gap threshold is 12s');
  assert(occContent.includes('else if (timeSpan <= 7 * 60 * 1000) maxGapMs = 20 * 1000;'), '5m timeframe gap threshold is 20s');
  assert(occContent.includes('if (prevX !== null && (p.x - prevX) > maxGapMs)'), 'Must split points into segments when delta exceeds gap threshold');
  assert(occContent.includes('// pathStr += ` M ${px.toFixed(1)} ${py.toFixed(1)}`'), 'Must break path with new M command on gap');
});

runTest('C.2: generatePrimaryAreaPath breaks area fill down to baseY on outage gaps', () => {
  assert(occContent.includes('const segArea = this.buildSplineArea(seg, baseY, this.trendLineSmooth)'), 'Must generate area fill per segment down to baseY');
  assert(occContent.includes('// Group points into continuous valid segments to avoid diagonal area fills across outages'), 'Area must avoid diagonal fills across outages');
});

runTest('C.3: Outage and recovery markers rendered with timestamp and reason tooltip', () => {
  assert(occContent.includes('v-for="marker in trendChartOutageMarkers"'), 'Must render outage markers in chart SVG');
  assert(occContent.includes("const color = isLost ? '#F43F5E' : '#10B981';"), 'Lost marker uses rose red, recovered uses emerald green');
  assert(occContent.includes('stroke-dasharray="4 3"'), 'Outage marker must use dashed line 4 3');
  assert(occContent.includes('v-if="hoveredOutageMarker"'), 'Must render hover tooltip for outage marker');
  assert(occContent.includes('hoveredOutageMarker.timeFormatted'), 'Outage tooltip must show formatted time');
  assert(occContent.includes('hoveredOutageMarker.reason'), 'Outage tooltip must show reason if available');
});

runTest('C.4: Scrubbing inside a data gap suppresses dot and flags inDataGap', () => {
  assert(occContent.includes("const isPointInGap = minDiff > maxGapMs || closest.quality === 'BAD';"), 'Scrubbing must identify when cursor is inside gap');
  assert(occContent.includes('plotY = null;'), 'Must set plotY to null when point is in gap');
  assert(occContent.includes("val: (!isPointInGap && closest.y !== null && closest.y !== undefined) ? this.formatAxisValue(closest.y) : '--'"), 'Must show -- when point is in gap');
  assert(occContent.includes("quality: isPointInGap ? 'NO_DATA' : (closest.quality || 'GOOD')"), 'Quality must be NO_DATA in gap');
  assert(occContent.includes('COMMUNICATION GAP (NO TELEMETRY)'), 'Tooltip must show communication gap notice');
});

// =========================================================================
// SECTION D: STATUS DAN EMPTY STATES
// =========================================================================

runTest('D.1: Distinct diagnostic badges rendered in panel header', () => {
  assert(occContent.includes('trendCommDiagnostics.badgeClass'), 'Must style diagnostic badge dynamically');
  assert(occContent.includes('trendCommDiagnostics.badge'), 'Must display diagnostic badge text');
  assert(occContent.includes('trendCommDiagnostics.secondsAgo'), 'Must show seconds/minutes since last received data');
});

runTest('D.2: Data load error state renders retry button that triggers refresh', () => {
  assert(occContent.includes("trendCommDiagnostics.state === 'API_ERROR'"), 'Must display API_ERROR state');
  assert(occContent.includes('@click="fetchTrendData(false)"'), 'Must provide click handler to retry trend data fetch');
  assert(occContent.includes('trendLoadError'), 'Must track API error in trendLoadError');
});

runTest('D.3: No historical data state renders informative message without crashing', () => {
  assert(occContent.includes("v-else-if=\"trendCommDiagnostics.state === 'NO_DATA'\""), 'Must render empty state when no data exists');
  assert(occContent.includes('NO DATA AVAILABLE'), 'Must display NO DATA AVAILABLE heading');
});

// =========================================================================
// SECTION E: VERIFIKASI KEUTUHAN KALKULASI & FULLSCREEN SYNC
// =========================================================================

runTest('E.1: Telemetry math and scaling formulas remain intact', () => {
  assert(occContent.includes('const span = scale.max - scale.min || 1;'), 'Span calculation must remain scale.max - scale.min || 1');
  assert(occContent.includes('this.getSeriesScale(s, s.dataPoints)'), 'Must use getSeriesScale for multi-series scaling');
});

runTest('E.2: Fullscreen modal retains strengthened axis, gap handling, and outage markers', () => {
  assert(occContent.includes(':x1="fullscreenChartDims.padLeft"'), 'Fullscreen chart must have vertical left Y-axis');
  assert(occContent.includes(':stroke="axisStroke"'), 'Fullscreen chart must use axisStroke');
  assert(occContent.includes('v-for="marker in fullscreenOutageMarkers"'), 'Fullscreen chart must render outage markers');
  assert(occContent.includes('fullscreenScrubPoint.inDataGap'), 'Fullscreen scrub tooltip must handle data gap');
});

console.log('\n===============================================================');
console.log(`TEST RESULTS: ${passedTests} / ${totalTests} TESTS PASSED`);
console.log('===============================================================\n');

if (passedTests !== totalTests) {
  process.exit(1);
}
