/**
 * Automated Test Suite: Operator Dashboard Performance Optimization Settings
 * Verifies localStorage persistence, UI toggle panel, component unmounting,
 * and functional safety under low-power / Raspberry Pi environments.
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

let passedTests = 0;
let failedTests = 0;

function runTest(desc, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${desc}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${desc}`);
    console.error(`  Error: ${err.message}`);
    failedTests++;
  }
}

const occPath = path.resolve(__dirname, '../src/views/OperatorCommandCenterView.vue');
const enPath = path.resolve(__dirname, '../src/i18n/locales/en.js');
const idPath = path.resolve(__dirname, '../src/i18n/locales/id.js');

const occContent = fs.readFileSync(occPath, 'utf8');
const enContent = fs.readFileSync(enPath, 'utf8');
const idContent = fs.readFileSync(idPath, 'utf8');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: OPERATOR DASHBOARD PERFORMANCE SETTINGS');
console.log('===============================================================');

// 1. LOCALSTORAGE KEY AND DEFAULT DATA STRUCTURE
runTest('1.1: Versioned localStorage key defined as datalogger_operator_performance_v1', () => {
  assert(occContent.includes("const PERF_STORAGE_KEY = 'datalogger_operator_performance_v1';"), 'Must define versioned key PERF_STORAGE_KEY');
});

runTest('1.2: Default settings structure contains all required properties', () => {
  assert(occContent.includes('parameterSparklinesEnabled: true'), 'Must default parameterSparklinesEnabled to true');
  assert(occContent.includes('trendAnalysisEnabled: true'), 'Must default trendAnalysisEnabled to true');
  assert(occContent.includes('alarmVisualEnabled: true'), 'Must default alarmVisualEnabled to true');
  assert(occContent.includes('alarmSoundEnabled: true'), 'Must default alarmSoundEnabled to true');
  assert(occContent.includes("liveUpdateMode: 'full'"), 'Must default liveUpdateMode to full');
  assert(occContent.includes('chartAnimationsEnabled: true'), 'Must default chartAnimationsEnabled to true');
  assert(occContent.includes('cardBlinksEnabled: true'), 'Must default cardBlinksEnabled to true');
});

runTest('1.3: loadPerformanceSettings safely handles missing or corrupt localStorage', () => {
  assert(occContent.includes('function loadPerformanceSettings()'), 'loadPerformanceSettings function must exist');
  assert(occContent.includes('localStorage.getItem(PERF_STORAGE_KEY)'), 'Must read from localStorage with PERF_STORAGE_KEY');
  assert(occContent.includes('JSON.parse(raw)'), 'Must parse JSON');
  assert(occContent.includes('catch (e)'), 'Must have try-catch error safety');
});

runTest('1.4: savePerformanceSettings persists JSON to localStorage', () => {
  assert(occContent.includes('function savePerformanceSettings('), 'savePerformanceSettings function must exist');
  assert(occContent.includes('localStorage.setItem(PERF_STORAGE_KEY, JSON.stringify('), 'Must set item to localStorage');
});

// 2. UI CONTROLS & POPOVER PANEL
runTest('2.1: Performance settings toolbar button exists with ID and accessibility attributes', () => {
  assert(occContent.includes('id="btn-operator-performance-settings"'), 'Toolbar button must have ID btn-operator-performance-settings');
  assert(occContent.includes("toggleToolbarPopover('performanceSettings')"), 'Button must toggle performanceSettings popover');
  assert(occContent.includes("activeToolbarPopover === 'performanceSettings'"), 'Popover panel must be bound to activeToolbarPopover');
});

runTest('2.2: Popover panel provides Reset to Defaults action and storage notice', () => {
  assert(occContent.includes('resetPerformanceSettings'), 'Must have resetPerformanceSettings method call');
  assert(occContent.includes('perfStorageNotice'), 'Must display browser-only storage notice');
});

runTest('2.3: Toggle controls exist for Sparklines, Trend Analysis, Visual Alarms, Sound, Blinks, Animations', () => {
  assert(occContent.includes("togglePerfSetting('parameterSparklinesEnabled')"), 'Must have sparkline toggle');
  assert(occContent.includes("togglePerfSetting('trendAnalysisEnabled')"), 'Must have trend toggle');
  assert(occContent.includes("togglePerfSetting('alarmVisualEnabled')"), 'Must have visual alarm toggle');
  assert(occContent.includes("togglePerfSetting('alarmSoundEnabled')"), 'Must have sound alarm toggle');
  assert(occContent.includes("togglePerfSetting('cardBlinksEnabled')"), 'Must have tile blinks toggle');
  assert(occContent.includes("togglePerfSetting('chartAnimationsEnabled')"), 'Must have chart animations toggle');
});

runTest('2.4: Live Dashboard Updates mode selector provides Full, Essential, and Reduced', () => {
  assert(occContent.includes("setPerfSetting('liveUpdateMode', 'full')"), 'Must provide Full live mode');
  assert(occContent.includes("setPerfSetting('liveUpdateMode', 'essential')"), 'Must provide Essential live mode');
  assert(occContent.includes("setPerfSetting('liveUpdateMode', 'reduced')"), 'Must provide Reduced live mode');
});

runTest('2.5: Operator Dashboard displays visual indicator dot when performance optimization is active', () => {
  assert(occContent.includes('isPerformanceOptimizedActive'), 'Must have computed isPerformanceOptimizedActive');
  assert(occContent.includes('OPTIMIZED'), 'Must show OPTIMIZED pill badge when active');
});

// 3. COMPONENT BYPASS & CPU OPTIMIZATIONS
runTest('3.1: Parameter Card Sparklines SVG is gated by perfSettings.parameterSparklinesEnabled', () => {
  assert(occContent.includes('v-if="perfSettings.parameterSparklinesEnabled && getSparkline(param).hasHistory"'), 'SVG sparkline must be gated with parameterSparklinesEnabled');
  assert(occContent.includes('if (this.perfSettings && (!this.perfSettings.parameterSparklinesEnabled'), 'recordSparkline must exit early when disabled');
  assert(occContent.includes('if (this.perfSettings && !this.perfSettings.parameterSparklinesEnabled)'), 'recomputeSparklineMeta and fetchInitialSparklines must exit early when disabled');
});

runTest('3.2: Trend Analysis Chart is unmounted when disabled and replaced with paused placeholder', () => {
  assert(occContent.includes('v-if="perfSettings.trendAnalysisEnabled" class="scada-panel'), 'Trend panel scada-panel must be conditionally mounted with v-if');
  assert(occContent.includes('TREND ANALYSIS DISABLED PLACEHOLDER'), 'Must provide disabled placeholder banner');
  assert(occContent.includes("setPerfSetting('trendAnalysisEnabled', true)"), 'Placeholder banner must allow quick re-enabling');
  assert(occContent.includes('if (this.perfSettings && !this.perfSettings.trendAnalysisEnabled) return;'), 'fetchTrendData and advanceLiveTrendWindow must exit early when disabled');
  assert(occContent.includes('if (this.perfSettings && !this.perfSettings.trendAnalysisEnabled)'), 'setupTrendAutoRefresh must not start interval when disabled');
});

runTest('3.3: Visual alarm flashing and audio sound are separated safely', () => {
  assert(occContent.includes('perfSettings.alarmVisualEnabled && isAlarmPulseActive'), 'Card and table alarm pulse must check alarmVisualEnabled');
  assert(occContent.includes('audioAlert.setSoundEnabled(soundOn)'), 'Alarm sound toggle must route through audioAlert.setSoundEnabled');
  assert(occContent.includes('perfAlarmNotice'), 'Must display alarm safety warning to operator');
});

runTest('3.4: Essential Live mode skips graphics calculations and renders while maintaining liveMap telemetry', () => {
  assert(occContent.includes("mode === 'essential'"), 'onTelemetryEvent must check essential mode');
  assert(occContent.includes('this.$set(this.liveMap, data.parameter_id, data)'), 'liveMap must always be updated regardless of mode');
});

runTest('3.5: Reduced Live mode throttles visual packet processing to 1s batch interval', () => {
  assert(occContent.includes("mode === 'reduced'"), 'onTelemetryEvent must handle reduced mode');
  assert(occContent.includes('this._lastReducedProcessTime'), 'Must track last reduced processing timestamp');
});

runTest('3.6: Chart animations toggle binds no-perf-animations and disables transitions in CSS', () => {
  assert(occContent.includes('no-perf-animations'), 'Root container must bind no-perf-animations class');
  assert(occContent.includes('.no-perf-animations'), 'CSS must define .no-perf-animations rule');
  assert(occContent.includes('animation: none !important;'), 'Must disable CSS animations in no-perf-animations');
});

// 4. FUNCTIONAL & COMMUNICATION INTEGRITY
runTest('4.1: WebSocket connection is strictly preserved and never disconnected', () => {
  assert(occContent.includes("window.addEventListener('device-telemetry-event', this.onTelemetryEvent)"), 'WebSocket telemetry listener must remain registered');
  assert(!occContent.includes('socket.close()') && !occContent.includes('ws.close()'), 'Must not close WebSocket connection');
});

// 5. LOCALIZATION COVERAGE
runTest('5.1: Localization files (en.js and id.js) contain all performance settings translation keys', () => {
  const requiredKeys = [
    'perfSettingsTitle',
    'perfSettingsSubtitle',
    'perfSettingsTooltip',
    'perfStorageNotice',
    'perfResetDefaults',
    'perfSparklines',
    'perfSparklinesDesc',
    'perfSparklinesOn',
    'perfSparklinesOff',
    'perfTrend',
    'perfTrendDesc',
    'perfTrendOn',
    'perfTrendOff',
    'perfAlarms',
    'perfAlarmVisual',
    'perfAlarmVisualDesc',
    'perfAlarmSound',
    'perfAlarmSoundDesc',
    'perfAlarmNotice',
    'perfLiveUpdates',
    'perfLiveFull',
    'perfLiveFullDesc',
    'perfLiveEssential',
    'perfLiveEssentialDesc',
    'perfLiveReduced',
    'perfLiveReducedDesc',
    'perfVisualEffects',
    'perfCardBlinks',
    'perfCardBlinksDesc',
    'perfChartAnimations',
    'perfChartAnimationsDesc',
  ];

  for (const k of requiredKeys) {
    assert(enContent.includes(`"${k}":`), `en.js missing ${k}`);
    assert(idContent.includes(`"${k}":`), `id.js missing ${k}`);
  }
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
} else {
  process.exit(0);
}

