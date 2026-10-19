/**
 * Automated Verification Suite: Sparkline Vertical Centering, Compact Series Selector & Unified Trend Controls
 * 
 * Verifies:
 * 1. Sparkline vertical centering at y=10 for constant/zero values, consistent h-[20px] box, no bottom cut-off
 * 2. getCompactTrendParamName maps parameter names to concise standard acronyms (PM2.5, PM10, TEMP, PRESS, SOLAR, WIND, W-DIR, HUM, etc.)
 * 3. Series chips are compact with smaller font, reduced padding, and tooltip with full parameter details
 * 4. All, Clear, and Reset are icon-only without text, placed immediately to the right of Chart Setup in one toolbar
 * 5. Chart Setup popover consolidates Mode (Processed/Raw), Scale (Norm%/Actual), Fullscreen, Deep Analysis, and Wind & Map
 * 6. Bottom chart footer is uncluttered with minimal SCADA status line
 * 7. Translations exist in English and Indonesian
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const enPath = path.join(__dirname, '../src/i18n/locales/en.js');
const enContent = fs.readFileSync(enPath, 'utf8');

const idPath = path.join(__dirname, '../src/i18n/locales/id.js');
const idContent = fs.readFileSync(idPath, 'utf8');

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
console.log('RUNNING TEST SUITE: SPARKLINE CENTERING & COMPACT TREND SELECTOR');
console.log('===============================================================');

// -------------------------------------------------------------
// REQUIREMENT 1: PARAMETER CARDS — SPARKLINE VERTICAL CENTERING & CONSISTENT HEIGHT
// -------------------------------------------------------------

runTest('1.1: Sparkline container uses consistent 20px height and vertical centering', () => {
  assert(occContent.includes('h-[20px]'), 'Sparkline container must use consistent h-[20px]');
  assert(occContent.includes('viewBox="0 0 100 20"'), 'Sparkline SVG must use viewBox="0 0 100 20"');
  assert(occContent.includes('my-auto'), 'Sparkline container must use my-auto for vertical alignment');
});

runTest('1.2: recomputeSparklineMeta calculates H=20 and centers constant/zero flat lines at y=10.0', () => {
  assert(occContent.includes('const H = 20;'), 'recomputeSparklineMeta must compute for H=20');
  assert(occContent.includes('const W = 100;'), 'recomputeSparklineMeta must compute for W=100');
  assert(occContent.includes('const isConstant = span === 0;'), 'Must detect constant values (including 0)');
  assert(occContent.includes('norm = isConstant ? 0.5 : (item.val - min) / span;'), 'Constant/zero values must map to norm=0.5 (exact center)');
  assert(occContent.includes('M 0,${y} L 100,${y}'), 'Single or constant points must draw centered flat horizontal line');
});

runTest('1.3: recomputeSparklineMeta provides generous vertical padding so varying peaks never sink or clip', () => {
  assert(occContent.includes('Math.max(span * 0.18, 0.01)'), 'Must use 18% span padding so peaks/troughs do not clip or sink');
});

// -------------------------------------------------------------
// REQUIREMENT 2: COMPACT SERIES SELECTOR — SHORTENED NAMES & TOOLTIPS
// -------------------------------------------------------------

runTest('2.1: getCompactTrendParamName helper maps all user-specified names', () => {
  assert(occContent.includes('getCompactTrendParamName(s)'), 'Must have getCompactTrendParamName helper method');

  // Extract function body and evaluate
  const match = occContent.match(/getCompactTrendParamName\(s\)\s*\{([\s\S]*?)\n\s*\},/);
  assert(match, 'Must be able to extract getCompactTrendParamName');
  const fn = new Function('s', match[1]);

  assert.strictEqual(fn({ paramName: 'PM 2.5' }), 'PM2.5');
  assert.strictEqual(fn({ paramName: 'PM 10' }), 'PM10');
  assert.strictEqual(fn({ paramName: 'TEMPERATURE' }), 'TEMP');
  assert.strictEqual(fn({ paramName: 'Air Pressure' }), 'PRESS');
  assert.strictEqual(fn({ paramName: 'Solar Radiation' }), 'SOLAR');
  assert.strictEqual(fn({ paramName: 'Wind Speed' }), 'WIND');
  assert.strictEqual(fn({ paramName: 'Wind Angle' }), 'W-DIR');
  assert.strictEqual(fn({ paramName: 'Humidity' }), 'HUM');
  assert.strictEqual(fn({ paramName: 'O3' }), 'O3');
  assert.strictEqual(fn({ paramName: 'SO2' }), 'SO2');
  assert.strictEqual(fn({ paramName: 'NO2' }), 'NO2');
  assert.strictEqual(fn({ paramName: 'CO' }), 'CO');
  assert.strictEqual(fn({ paramName: 'HC' }), 'HC');
});

runTest('2.2: Series chips use compact typography, reduced padding, and tooltip with full parameter details', () => {
  assert(occContent.includes('getCompactTrendParamName(s) || (s.paramName || s.paramCode)'), 'Chip must display compact name');
  assert(occContent.includes('px-1.5 py-0.5'), 'Chip must use reduced horizontal padding px-1.5 py-0.5');
  assert(occContent.includes('text-[9px]'), 'Chip must use compact text-[9px]');
  assert(occContent.includes(':title="(s.deviceCode ? s.deviceCode + \' · \' : \'\') + (s.paramName || s.paramCode)'), 'Chip must have comprehensive tooltip');
});

// -------------------------------------------------------------
// REQUIREMENT 3: ICON-ONLY ALL, CLEAR, RESET, AND CHART SETUP IN ONE TOOLBAR
// -------------------------------------------------------------

runTest('3.1: All, Clear, Reset, and Chart Setup are icon-only without text buttons', () => {
  // Check that the toolbar buttons inside Series Selector Bar do not contain text like >All</button>
  const seriesSelectorMatch = occContent.match(/<!-- Compact Toolbar Group: Chart Setup, All, Clear, Reset -->([\s\S]*?)<!-- Compact Series Chips -->/);
  assert(seriesSelectorMatch, 'Must have Compact Toolbar Group');
  const toolbarHtml = seriesSelectorMatch[1];

  assert(!toolbarHtml.match(/>\s*All\s*<\/button>/i), 'All button must be icon-only');
  assert(!toolbarHtml.match(/>\s*Clear\s*<\/button>/i), 'Clear button must be icon-only');
  assert(!toolbarHtml.match(/>\s*Reset\s*<\/button>/i), 'Reset button must be icon-only');
});

runTest('3.2: All, Clear, and Reset are placed immediately to the right of Chart Setup', () => {
  const seriesSelectorMatch = occContent.match(/<!-- Compact Toolbar Group: Chart Setup, All, Clear, Reset -->([\s\S]*?)<!-- Compact Series Chips -->/);
  const toolbarHtml = seriesSelectorMatch[1];

  const chartSetupIdx = toolbarHtml.indexOf('aria-label="Chart Setup"');
  const allIdx = toolbarHtml.indexOf('@click="selectAllEligibleTrendSeries(true)"');
  const clearIdx = toolbarHtml.indexOf('@click="clearAllTrendSeries"');
  const resetIdx = toolbarHtml.indexOf('@click="resetTrendSelectionToDefault"');

  assert(chartSetupIdx !== -1, 'Chart Setup icon must exist');
  assert(allIdx !== -1, 'All icon must exist');
  assert(clearIdx !== -1, 'Clear icon must exist');
  assert(resetIdx !== -1, 'Reset icon must exist');

  assert(chartSetupIdx < allIdx, 'All icon must be placed to the right of Chart Setup');
  assert(allIdx < clearIdx, 'Clear icon must be placed to the right of All icon');
  assert(clearIdx < resetIdx, 'Reset icon must be placed to the right of Clear icon');
});

runTest('3.3: Every icon has descriptive tooltip and recognizable SVG path', () => {
  assert(occContent.includes('title="Select all eligible parameters (All)"'), 'All button must have tooltip');
  assert(occContent.includes('title="Clear all series (Clear)"'), 'Clear button must have tooltip');
  assert(occContent.includes('title="Reset to default (Reset)"'), 'Reset button must have tooltip');
  assert(occContent.includes('M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4'), 'Chart Setup must use sliders-horizontal icon');
});

// -------------------------------------------------------------
// REQUIREMENT 4: UNIFIED CONTROLS IN ONE MENU (MODE, SCALE, FULLSCREEN, DEEP ANALYSIS, WIND & MAP)
// -------------------------------------------------------------

runTest('4.1: Unified settings menu popover contains all 5 required controls', () => {
  assert(occContent.includes('showTrendSettingsMenu'), 'Must have showTrendSettingsMenu popover');
  // 1. Mode: Processed vs Raw
  assert(occContent.includes('@click="setChartMode(\'processed\')"'), 'Menu must contain Processed mode button');
  assert(occContent.includes('@click="setChartMode(\'raw\')"'), 'Menu must contain Raw mode button');
  // 2. Scale: Norm % vs Actual Value
  assert(occContent.includes('@click="trendScaleMode = \'auto\'"'), 'Menu must contain Actual Value / Dual-Axis scale button');
  assert(occContent.includes('@click="trendScaleMode = \'normalized\'"'), 'Menu must contain Norm % scale button');
  // 3. Show Wind & Map
  assert(occContent.includes('@click="toggleEnvironmentalVisualizations"'), 'Menu must contain Show Wind & Map toggle');
  // 4. Expand Fullscreen
  assert(occContent.includes('@click="openTrendFullscreen(); showTrendSettingsMenu = false"'), 'Menu must contain Expand Fullscreen button');
  // 5. Deep Analysis
  assert(occContent.includes('to="/monitoring/analysis"'), 'Menu must contain Deep Analysis route link');
});

runTest('4.2: Menu options display active state on current selection', () => {
  assert(occContent.includes("chartMode === 'processed' ? 'bg-cyan-500 text-slate-950 shadow-xs'"), 'Processed mode must show active state');
  assert(occContent.includes("chartMode === 'raw' ? 'bg-cyan-500 text-slate-950 shadow-xs'"), 'Raw mode must show active state');
  assert(occContent.includes("trendScaleMode === 'auto' ? 'bg-cyan-500 text-slate-950 shadow-xs'"), 'Auto scale must show active state');
  assert(occContent.includes("trendScaleMode === 'normalized' ? 'bg-cyan-500 text-slate-950 shadow-xs'"), 'Norm scale must show active state');
});

runTest('4.3: Bottom footer row is uncluttered with minimal SCADA status line', () => {
  assert(occContent.includes('<!-- Chart Minimal Status & Quick Controls Anchor -->'), 'Chart footer must be minimal');
  assert(occContent.includes('LIVE · {{ chartMode.toUpperCase() }}'), 'Must display live mode status');
});

// -------------------------------------------------------------
// REQUIREMENT 5: LOCALIZATION
// -------------------------------------------------------------

runTest('5.1: chartSetup translation exists in English and Indonesian', () => {
  assert(enContent.includes('"chartSetup"'), 'en.js must have chartSetup translation');
  assert(idContent.includes('"chartSetup"'), 'id.js must have chartSetup translation');
});

console.log('===============================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${totalTests - passedTests} FAILED (TOTAL: ${totalTests})`);
console.log('===============================================================');

if (passedTests !== totalTests) {
  process.exit(1);
}
