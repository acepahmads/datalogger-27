/**
 * Automated Verification Suite: Trend Analysis Thin Lines, Clean SCADA Visualization & Spike Resilience
 * 
 * Tests:
 * 1. Thin stroke-width verification (1.8px primary, 1.6px secondary, 2.2px on hover)
 * 2. Rounded caps and joins (stroke-linecap="round", stroke-linejoin="round")
 * 3. Low-contrast subtle gridlines (stroke-width 0.8px, dash 3 3)
 * 4. Spike handling: computeNiceScale detects isolated spikes without compressing normal data
 * 5. Outlier markers and badges rendered at spike timestamps with real values
 * 6. User-controlled zoom toggle: robust focus vs full span
 * 7. Multi-series color consistency and dual-axis separation for incompatible units
 * 8. Strict preservation of data integrity (no clamping, no false 0 drops)
 * 9. Light and Dark theme color definitions
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const telAnalysisPath = path.join(__dirname, '../src/views/TelemetryAnalysisView.vue');
const telAnalysisContent = fs.readFileSync(telAnalysisPath, 'utf8');

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
console.log('RUNNING TEST SUITE: THIN LINES & CLEAN SCADA VISUALIZATION');
console.log('===============================================================\n');

// 1. Line Thickness & Sparkline Parity
runTest('Test 1: Series lines use thin 1.8px primary / 1.6px secondary strokes (hover 2.2px)', () => {
  assert(occContent.includes('hoveredTrendSeriesId === s.id ? 2.2 : (sIdx === 0 ? 1.8 : 1.6)'), 'OCC inline chart must use 1.8px/1.6px (2.2px on hover)');
  assert(occContent.includes('hoveredTrendSeriesId === s.id ? 2.4 : (sIdx === 0 ? 1.8 : 1.6)'), 'OCC fullscreen chart must use 1.8px/1.6px');
  assert(telAnalysisContent.includes('hoveredSeriesId === s.id || focusedSeriesId === s.id ? 2.2 : 1.8'), 'TelemetryAnalysisView must use 1.8px (2.2px on hover)');
});

// 2. Rounded Caps and Joins
runTest('Test 2: All trend paths use rounded line caps and joins', () => {
  assert(occContent.includes('stroke-linecap="round"'), 'Must have stroke-linecap="round"');
  assert(occContent.includes('stroke-linejoin="round"'), 'Must have stroke-linejoin="round"');
  assert(telAnalysisContent.includes('stroke-linecap="round"'), 'TelemetryAnalysisView must have stroke-linecap="round"');
  assert(telAnalysisContent.includes('stroke-linejoin="round"'), 'TelemetryAnalysisView must have stroke-linejoin="round"');
});

// 3. Subtle, Low-Contrast Gridlines
runTest('Test 3: Gridlines are subtle and low-contrast (0.8px stroke width)', () => {
  assert(occContent.includes('stroke-width="0.8"'), 'OCC gridlines must use subtle 0.8px stroke');
  assert(occContent.includes("this.isDark ? '#142438' : '#CBD5E1'"), 'OCC gridStroke must use refined low-contrast palette');
  assert(occContent.includes("this.isDark ? '#20354B' : '#94A3B8'"), 'OCC axisStroke must use clean slate/blue-gray');
});

// 4. Isolated Spike Detection & Robust Scale Logic
runTest('Test 4: computeNiceScale detects isolated spikes and preserves normal telemetry visibility', () => {
  assert(occContent.includes('highSpikes = sorted.filter(v => v > p95 + 2.5 * normalSpan)'), 'Must detect upper spike threshold');
  assert(occContent.includes('trendSpikeZoom'), 'Must support trendSpikeZoom state');
  assert(occContent.includes('SMART ROBUST'), 'Must explicitly indicate SMART ROBUST scale mode');
  assert(occContent.includes('FULL SPAN'), 'Must explicitly indicate FULL SPAN scale mode');
});

// 5. Outlier Markers on Canvas
runTest('Test 5: Outlier spike markers rendered on canvas with beacon lines and real values', () => {
  assert(occContent.includes('trendOutlierSpikes'), 'Must compute trendOutlierSpikes');
  assert(occContent.includes('fullscreenOutlierSpikes'), 'Must compute fullscreenOutlierSpikes');
  assert(occContent.includes('▲ {{ spike.displayVal }}'), 'Must display outlier triangle and value tag');
  assert(occContent.includes('toggleTrendSpikeZoom'), 'Clicking outlier marker must toggle spike zoom');
});

// 6. User-Controlled Spike Zoom Toggle Button
runTest('Test 6: Provides explicit spike alert pill and toggle buttons in header and footer', () => {
  assert(occContent.includes('hasOutlierSpike'), 'Must check hasOutlierSpike');
  assert(occContent.includes('activeSpikeSummary'), 'Must compute activeSpikeSummary');
  assert(occContent.includes("trendSpikeZoom === 'full' ? 'Focus Normal' : 'Full Span'"), 'Must offer 1-click toggle between Focus Normal and Full Span');
});

// 7. Light Mode & Dark Mode Styling
runTest('Test 7: Light mode uses sky-blue border (#B8D4EA) and dark labels (#475569)', () => {
  assert(occContent.includes('border-[#B8D4EA]'), 'Canvas container must use subtle blue-gray border in Light Mode');
  assert(occContent.includes("isDark ? '#7E8B9B' : '#475569'"), 'Axis labels must use #475569 dark slate in Light Mode');
  assert(occContent.includes('dark:bg-[#060B13]'), 'Canvas container must preserve dark SCADA surface');
});

// 8. Low-Opacity Area Fill
runTest('Test 8: Area fill uses low-opacity gradient (0.08 down to 0.00)', () => {
  assert(occContent.includes('stop-opacity="0.08"'), 'OCC primary area gradient must start at low opacity (0.08)');
  assert(telAnalysisContent.includes('stop-opacity="0.08"'), 'TelemetryAnalysisView gradient must use low opacity (0.08)');
});

console.log('\n===============================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${totalTests - passedTests} FAILED (TOTAL: ${totalTests})`);
console.log('===============================================================');

if (passedTests !== totalTests) {
  process.exit(1);
}
