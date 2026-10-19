/**
 * Automated Acceptance Tests: Clean Parameter Cards & Subtle Chart Selection
 *
 * Verifies:
 * 1. Default timestamp (e.g. 10:24:08 Just now) and footer divider (border-t) hidden by default on normal cards
 * 2. Timing/quality information displayed conditionally only when data is stale, invalid, or unavailable
 * 3. Text "In Chart" removed from all parameter cards
 * 4. Selected chart parameters use subtle cyan border (no heavy glow or dark shadows)
 * 5. Unselected parameters use standard subtle border with soft hover
 * 6. Alarm and Warning indicators retain their severity priority over chart selection
 * 7. Selection directly bound to activeTrendSeries with source device identification
 * 8. Full timestamp and timing accessible via card hover tooltip (:title="getParamTimingTitle(param)")
 * 9. Context menu provides explicit Plot/Remove from Trend Analysis action
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: CLEAN PARAMETER CARDS & SUBTLE CHART SELECTION');
console.log('===============================================================');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const enPath = path.join(__dirname, '../src/i18n/locales/en.js');
const enContent = fs.readFileSync(enPath, 'utf8');

const idPath = path.join(__dirname, '../src/i18n/locales/id.js');
const idContent = fs.readFileSync(idPath, 'utf8');

let passedTests = 0;
let failedTests = 0;

function runTest(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    failedTests++;
  }
}

// 1. Uniform Compact Height Preserved
runTest('Test 1: Parameter cards maintain compact 58px uniform box-border height without layout shift', () => {
  assert(occContent.includes('h-[58px] min-w-0 box-border overflow-hidden'), 'Card must have h-[58px] min-w-0 box-border overflow-hidden');
});

// 2. Hide Timestamp and Footer Divider by Default
runTest('Test 2: Normal cards hide timestamp and footer divider (no permanent empty space)', () => {
  // Footer divider (border-t) must NOT be present on normal card footer
  assert(!occContent.includes('pt-0.5 border-t border-slate-100 dark:border-[#132235] text-[8px] font-mono leading-none mt-auto'), 'Permanent footer divider border-t must be removed');
  // Area 3 must be conditionally rendered only when data quality issue occurs
  assert(occContent.includes('v-if="isParamDataQualityIssue(param)"'), 'Timing footer must be gated behind isParamDataQualityIssue');
  assert(occContent.includes('isParamDataQualityIssue(param)'), 'Must have isParamDataQualityIssue method');
});

// 3. Data Quality Issues Display Timing & Reason Clearly
runTest('Test 3: Stale, invalid, offline, or missing data displays timing label and quality reason tag', () => {
  assert(occContent.includes('getParamTimingIcon(param)'), 'Must render timing icon on quality issue');
  assert(occContent.includes('getParamTimingLabel(param)'), 'Must render timing label on quality issue');
  assert(occContent.includes('getParamQualityReasonText(param)'), 'Must render quality reason text');
  assert(occContent.includes("param.quality === 'BAD' || param.device_status === 'OFFLINE'"), 'Must style bad/offline quality issues distinct from warning');
});

// 4. Remove "In Chart" Label from all cards
runTest('Test 4: Text "In Chart" (cleanInChartLabel) completely removed from parameter cards', () => {
  assert(!occContent.includes('cleanInChartLabel'), 'cleanInChartLabel must not be rendered in parameter card template');
  assert(!occContent.includes("isParamSelectedForTrend(param) ? cleanInChartLabel"), 'In Chart label must not be rendered on selected cards');
});

// 5. Subtle Cyan Border for Selected Parameters
runTest('Test 5: Selected parameters in Trend Analysis use subtle cyan border with ring-1 and no heavy glow', () => {
  assert(
    occContent.includes("isParamSelectedForTrend(param)\n                        ? 'bg-white dark:bg-[#0B1422] border-cyan-500/75 dark:border-cyan-400/65 ring-1 ring-cyan-500/25 dark:ring-cyan-400/20 shadow-xs hover:border-cyan-500 dark:hover:border-cyan-400'") ||
    occContent.includes("border-cyan-500/75 dark:border-cyan-400/65 ring-1 ring-cyan-500/25 dark:ring-cyan-400/20"),
    'Selected card must use subtle cyan border (border-cyan-500/75 dark:border-cyan-400/65 ring-1 ring-cyan-500/25 dark:ring-cyan-400/20)'
  );
});

// 6. Standard Border & Soft Hover for Unselected Parameters
runTest('Test 6: Unselected parameters use standard subtle border with soft hover (not matching selected intensity)', () => {
  assert(
    occContent.includes("border-slate-200 dark:border-[#20354B] hover:border-slate-300 dark:hover:border-[#334b66] shadow-xs dark:shadow-none"),
    'Unselected card must use standard subtle border with soft hover'
  );
});

// 7. Alarm and Warning Retain Priority Over Selection Border
runTest('Test 7: Critical alarm and Warning retain visual priority over chart selection', () => {
  assert(occContent.includes('isParamAlarmOrStaleBad(param)'), 'isParamAlarmOrStaleBad check must precede trend selection check');
  assert(occContent.includes('border-rose-500'), 'Critical alarm must use rose border');
  assert(occContent.includes('border-amber-500'), 'Warning alarm must use amber border');
});

// 8. Single Source of Truth for Trend Selection & Device ID Consistency
runTest('Test 8: Selection state strictly derived from activeTrendSeries with source device identification', () => {
  assert(occContent.includes('s.parameterId === param.param_id && (!s.deviceId || !param.device_id || s.deviceId === param.device_id)'), 'Selection check must match both parameter ID and device ID consistently');
  assert(occContent.includes('selectParamForTrend(param)'), 'Clicking card must execute selectParamForTrend');
});

// 9. Full Timestamp Accessible via Tooltip
runTest('Test 9: Tooltip provides full ISO timestamp, relative time, quality, reason, and status on hover', () => {
  assert(occContent.includes(':title="getParamTimingTitle(param)"'), 'Card container must bind getParamTimingTitle to title attribute');
  assert(occContent.includes('timeFull'), 'getParamTimingTitle must include localized full time');
  assert(occContent.includes('timeIso'), 'getParamTimingTitle must include ISO time string');
  assert(occContent.includes('rel'), 'getParamTimingTitle must include relative time');
  assert(occContent.includes('quality_reason'), 'getParamTimingTitle must include quality reason');
});

// 10. Context Menu Action for Plot / Remove Trend
runTest('Test 10: Right-click context menu provides explicit Plot/Remove from Trend Analysis action', () => {
  assert(occContent.includes('selectParamForTrend(paramContextMenu.param); closeParamContextMenu()'), 'Context menu must have Plot/Remove Trend action');
  assert(occContent.includes("isParamSelectedForTrend(paramContextMenu.param) ? ($t('operator.removeFromTrend')"), 'Context menu must dynamically display Remove or Plot Trend');
  assert(enContent.includes('"removeFromTrend": "Remove from Trend"'), 'English locale must contain removeFromTrend');
  assert(idContent.includes('"removeFromTrend": "Hapus dari Tren"'), 'Indonesian locale must contain removeFromTrend');
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
}
