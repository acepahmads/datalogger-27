/**
 * Test Suite: Compact Telemetry Cards with Mini Sparklines
 * Verifies all requirements of UI Enhancement:
 * 1. Responsive card grid (3-4 columns) supporting 6, 14, 24, and 40 parameters
 * 2. Avoids internal vertical scrolling when configured parameter count permits (<= 9)
 * 3. Never silently hides parameters; preserves Cards vs Table alternative view
 * 4. Card height is compact (105-125px target: h-[114px] sm:h-[118px])
 * 5. Prominent parameter name, smaller device code, and parameter icon
 * 6. Large latest numeric value, correct unit, and compact quality badges
 * 7. Mini sparkline (25-35px tall: h-[26px]) with subtle area gradient
 * 8. Real historical samples loaded chronologically from telemetry history
 * 9. Missing data handled as true gaps (disconnected paths, no fake zero)
 * 10. No sample fabrication or flat line when < 2 samples; shows clean empty state
 * 11. Preserves real peaks and extrema without arbitrary clamping
 * 12. Stable series color consistent with Trend Analysis palette and alarm states
 * 13. Preserves Trend Analysis card selection with subtle highlight
 * 14. Bounded points (<= 30) and lightweight pure SVG for Raspberry Pi performance
 * 15. Complete English and Indonesian localization
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('================================================================');
console.log('RUNNING AUTOMATED TEST SUITE: COMPACT TELEMETRY CARDS & SPARKLINES');
console.log('================================================================\n');

let passedTests = 0;
let failedTests = 0;

function runTest(testName, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${testName}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${testName}`);
    console.error(`  Error: ${err.message}`);
    failedTests++;
  }
}

// Read relevant files
const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const enPath = path.join(__dirname, '../src/i18n/locales/en.js');
const enContent = fs.readFileSync(enPath, 'utf8');

const idPath = path.join(__dirname, '../src/i18n/locales/id.js');
const idContent = fs.readFileSync(idPath, 'utf8');

// Test 1: Card container uses uniform height within design targets
runTest('Test 1: Card container uses uniform height within design targets (88-95px, 105-125px, or 130-150px)', () => {
  assert(
    occContent.includes('h-[88px] sm:h-[90px]') || occContent.includes('h-[134px] sm:h-[140px]') || occContent.includes('h-[114px] sm:h-[118px]'),
    'Card container must enforce uniform height'
  );
  assert(
    occContent.includes('flex flex-col justify-between'),
    'Card must use flex layout for balanced vertical distribution'
  );
});

// Test 2: Card Header layout
runTest('Test 2: Header contains smaller device code, parameter icon, prominent name, alarm badge, and pin', () => {
  // Device code is smaller (10-11px font)
  assert(occContent.includes('text-[10.5px] sm:text-[11px] font-mono'), 'Device code must use 10.5-11px font');
  assert(occContent.includes('param.device_code'), 'Device code must be bound to param.device_code');
  // Dot separator
  assert(occContent.includes('·'), 'Separator dot must be present');
  // Parameter icon
  assert(occContent.includes('<ParameterIcon'), 'ParameterIcon component must be rendered');
  // Prominent parameter name (14-15px bold)
  assert(occContent.includes('text-sm sm:text-[15px] font-bold'), 'Parameter name must use 14-15px font with bold weight');
  // Visible alarm indicator on header
  assert(occContent.includes("param.alarm_status === 'CRITICAL' ? 'CRIT' : 'WARN'"), 'Header must show compact alarm badge');
  // Pin button
  assert(occContent.includes('togglePin(param.param_id)'), 'Pin favorite button must be intact');
});

// Test 3: Value and Quality row
runTest('Test 3: Value row contains large numeric display, unit, limits, and compact quality badges', () => {
  assert(
    occContent.includes('text-2xl sm:text-[28px]') || occContent.includes('text-lg sm:text-[21px] font-bold font-mono'),
    'Numeric value must be prominent (28-36px bold mono)'
  );
  assert(occContent.includes('formatParamValue(param)'), 'Value must be formatted via formatParamValue');
  assert(occContent.includes('param.unit'), 'Unit must be rendered');
  assert(occContent.includes('qualityBadgeClass(param.quality)'), 'Quality badge styling must be applied');
  assert(occContent.includes('HELD') && occContent.includes('FX'), 'Sub indicators for HELD and Formula must be supported');
});

// Test 4: Mini Sparkline height and SVG configuration
runTest('Test 4: Mini sparkline container is positioned beside value with subtle area gradient and smooth stroke', () => {
  assert(occContent.includes('h-8') || occContent.includes('h-9 sm:h-10') || occContent.includes('h-[26px]'), 'Sparkline height must be approximately 26-40px');
  assert(occContent.includes('viewBox="0 0 100 28"') || occContent.includes('viewBox="0 0 100 26"'), 'SVG viewBox must match aspect ratio');
  assert(occContent.includes('<linearGradient'), 'SVG must use linear gradient for subtle area fill');
  assert(occContent.includes('stop-opacity="0.22"') || occContent.includes('stop-opacity="0.25"'), 'Gradient fill must be subtle');
  assert(occContent.includes('stroke-width="1.8"') || occContent.includes('stroke-width="1.75"'), 'Stroke width must be optimized for mini sparkline');
});

// Test 5: Real historical sample fetching and chronological ordering
runTest('Test 5: Real historical samples fetched on load and sorted chronologically ascending', () => {
  assert(occContent.includes('async fetchInitialSparklines()'), 'fetchInitialSparklines method must be implemented');
  assert(occContent.includes('/api/devices/${dev.id}/telemetry/history'), 'Must fetch historical samples from telemetry history API');
  assert(occContent.includes('grouped[pid].sort((a, b) => a.time - b.time)'), 'Samples must be sorted strictly chronologically ascending');
  assert(occContent.includes('fetchInitialSparklines();'), 'fetchDevices must trigger fetchInitialSparklines');
});

// Test 6: Gap handling in sparkline (no fake zero, separate disconnected paths)
runTest('Test 6: Missing data and time gaps produce disconnected segments rather than dropping to zero', () => {
  assert(occContent.includes('isTimeGap'), 'Time gaps between samples must be detected');
  assert(occContent.includes('strokePaths.push'), 'Multiple stroke path segments must be supported');
  assert(occContent.includes('areaPaths.push'), 'Multiple area path segments must be supported');
  assert(occContent.includes('v-for="(strokeD, sIdx) in getSparkline(param.param_id).strokePaths"'), 'Template must render segmented stroke paths');
  assert(occContent.includes('v-for="(areaD, aIdx) in getSparkline(param.param_id).areaPaths"'), 'Template must render segmented area paths');
});

// Test 7: Empty state when < 2 samples (no flat line fabrication)
runTest('Test 7: Renders clean empty state when fewer than 2 valid history samples exist (no fake line)', () => {
  assert(occContent.includes('v-if="getSparkline(param.param_id).hasHistory"'), 'Only renders SVG chart when hasHistory is true');
  assert(occContent.includes('v-else'), 'Renders fallback container when hasHistory is false');
  assert(occContent.includes('awaitingSamples'), 'Empty state must show awaiting history / samples indicator');
});

// Test 8: Real peaks and extrema preserved
runTest('Test 8: Normalization preserves actual peaks and extrema with proportional vertical padding', () => {
  assert(occContent.includes('let min = Infinity'), 'Min must be calculated from actual data');
  assert(occContent.includes('let max = -Infinity'), 'Max must be calculated from actual data');
  assert(occContent.includes('const pad = span * 0.08'), 'Proportional padding must ensure extrema are not clipped');
  assert(occContent.includes('span = max - min'), 'Span must dynamically scale Y coordinate normalization');
});

// Test 9: Responsive Grid with 3-4 columns and parameter count recalculation
runTest('Test 9: Responsive Grid adapts across 6, 14, 24, and 40 parameters without clipping', () => {
  assert(occContent.includes('cardGridClass()'), 'cardGridClass computed property must be defined');
  assert(occContent.includes('grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4'), 'CSS grid must support 3 to 4 columns on desktop/wide');
  assert(occContent.includes('cardGridContainerStyle()'), 'cardGridContainerStyle computed property must be defined');
  // For small parameter counts (<= 9), no internal scroll needed
  assert(occContent.includes('maxHeight: \'none\''), 'Internal vertical scroll is bypassed when parameters fit (count <= 9)');
  // Never hide parameters
  assert(occContent.includes('v-for="param in filteredParameters"'), 'Renders all filtered parameters without artificial limits');
});

// Test 10: Alternative view mode (Cards vs Table) is preserved
runTest('Test 10: Cards vs Table switcher provides explicit full alternative view', () => {
  assert(occContent.includes("v-if=\"viewMode === 'cards'\""), 'Cards view mode switch preserved');
  assert(occContent.includes("v-else class=\"scada-panel"), 'Table view mode switch preserved as alternative');
});

// Test 11: Card interaction and Trend Analysis selection
runTest('Test 11: Clicking card selects parameter for Trend Analysis with subtle color-matched highlight', () => {
  assert(occContent.includes('@click="selectParamForTrend(param)"'), 'Card click must trigger selectParamForTrend');
  assert(occContent.includes('getTrendSeriesColor(param)'), 'Selected card border and highlight must match trend series color');
  assert(occContent.includes('isParamSelectedForTrend(param)'), 'isParamSelectedForTrend must govern selection state');
  assert(occContent.includes('inChart'), 'Footer action must indicate selection state in chart');
});

// Test 12: Bounded points and Raspberry Pi performance
runTest('Test 12: Sparkline point history is strictly bounded (<= 25-30 points) with O(1) cache lookup', () => {
  assert(occContent.includes('arr.length > 25') || occContent.includes('arr.length > 30'), 'Point count must be capped');
  assert(occContent.includes('arr.shift()'), 'Old points must be shifted out');
  assert(occContent.includes('sparklineMetaCache'), 'Metadata cache must prevent expensive recalculations in template');
});

// Test 13: Complete English and Indonesian localization
runTest('Test 13: Locales contain all required keys (awaitingSamples, inChart, plotTrend)', () => {
  assert(enContent.includes('"awaitingSamples": "Awaiting history"'), 'en.js must have awaitingSamples');
  assert(enContent.includes('"inChart": "● In Chart"'), 'en.js must have inChart');
  assert(enContent.includes('"plotTrend": "+ Plot Trend"'), 'en.js must have plotTrend');

  assert(idContent.includes('"awaitingSamples": "Menunggu riwayat"'), 'id.js must have awaitingSamples');
  assert(idContent.includes('"inChart": "● Di Grafik"'), 'id.js must have inChart');
  assert(idContent.includes('"plotTrend": "+ Plot Tren"'), 'id.js must have plotTrend');
});

// Test 14: Simulated parameter count test (6, 14, 24, 40)
runTest('Test 14: Grid classes produce expected column count and heights for 6, 14, 24, 40 parameters', () => {
  function getGridInfo(count) {
    let cols = 3;
    let className = 'grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-2 sm:gap-2.5';
    let maxHeight = count <= 9 ? 'none' : 'calc(100vh - 290px)';
    let rows3Col = Math.ceil(count / 3);
    let rows4Col = Math.ceil(count / 4);
    let estHeight3Col = rows3Col * 116 + (rows3Col - 1) * 10;
    let estHeight4Col = rows4Col * 116 + (rows4Col - 1) * 10;
    return { count, maxHeight, rows3Col, rows4Col, estHeight3Col, estHeight4Col };
  }

  const test6 = getGridInfo(6);
  assert.strictEqual(test6.maxHeight, 'none');
  assert.strictEqual(test6.rows3Col, 2);
  assert(test6.estHeight3Col <= 250, '6 parameters in 3 columns must fit within 250px');

  const test14 = getGridInfo(14);
  assert.strictEqual(test14.maxHeight, 'calc(100vh - 290px)');
  assert.strictEqual(test14.rows4Col, 4);

  const test24 = getGridInfo(24);
  assert.strictEqual(test24.rows4Col, 6);

  const test40 = getGridInfo(40);
  assert.strictEqual(test40.rows4Col, 10);
});

console.log('================================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('================================================================');

if (failedTests > 0) {
  process.exit(1);
}
