const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING AUTOMATED TEST SUITE: OCC TREND SELECTION, FRESHNESS & ZERO-VALUE ALARM POLICY');
console.log('===============================================================');

let passedTests = 0;
let failedTests = 0;

function runTest(testName, testFn) {
  try {
    testFn();
    console.log(`✓ PASS: ${testName}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${testName}`);
    console.error(`  Error: ${err.message}`);
    failedTests++;
  }
}

// Read relevant project files
const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const modalPath = path.join(__dirname, '../src/components/ParameterModal.vue');
const modalContent = fs.readFileSync(modalPath, 'utf8');

const enLocalePath = path.join(__dirname, '../src/i18n/locales/en.js');
const enLocaleContent = fs.readFileSync(enLocalePath, 'utf8');

const idLocalePath = path.join(__dirname, '../src/i18n/locales/id.js');
const idLocaleContent = fs.readFileSync(idLocalePath, 'utf8');

const modelDevicePath = path.join(__dirname, '../../internal/model/device.go');
const modelDeviceContent = fs.readFileSync(modelDevicePath, 'utf8');

const migrationPath = path.join(__dirname, '../../internal/database/migration.go');
const migrationContent = fs.readFileSync(migrationPath, 'utf8');

const serviceDevicePath = path.join(__dirname, '../../internal/service/device_service.go');
const serviceDeviceContent = fs.readFileSync(serviceDevicePath, 'utf8');

const alarmEvaluatorPath = path.join(__dirname, '../../internal/service/alarm_evaluator.go');
const alarmEvaluatorContent = fs.readFileSync(alarmEvaluatorPath, 'utf8');

// -------------------------------------------------------------
// REQUIREMENT 1: TREND ANALYSIS — SELECT ALL PARAMETERS BY DEFAULT
// -------------------------------------------------------------

runTest('1.1: Fresh visit selects all eligible parameters up to maxTrendSeries (not just 2)', () => {
  assert(occContent.includes('selectAllEligibleTrendSeries'), 'Must have selectAllEligibleTrendSeries method');
  assert(!occContent.includes('const firstTwo = this.allParameters.slice(0, 2);'), 'Must NOT limit default to first two parameters');
  assert(occContent.includes('this.allParameters.slice(0, this.maxTrendSeries)'), 'Must slice all eligible parameters up to maxTrendSeries');
});

runTest('1.2: Stable parameter identity uses device ID + parameter ID/code', () => {
  assert(occContent.includes("id: `${p.device_id}_${p.param_id}`"), 'Must compose stable series ID from device_id and param_id');
  assert(occContent.includes('Number(p.device_id) === Number(item.deviceId) && Number(p.param_id) === Number(item.parameterId)'), 'Must restore by matching device_id and param_id');
});

runTest('1.3: Expanded high-contrast SCADA palette supports up to 16 distinct series', () => {
  assert(occContent.includes("maxTrendSeries: 16"), 'maxTrendSeries must be configured to 16');
  // Check first 6 colors preserved for backward compatibility
  assert(occContent.includes("'#00E5FF', '#3B82F6', '#10B981', '#F59E0B', '#EC4899', '#8B5CF6'"), 'Must preserve the first 6 SCADA colors');
  // Check expanded colors
  assert(occContent.includes("'#14B8A6'"), 'Must include expanded color Teal');
  assert(occContent.includes("'#F97316'"), 'Must include expanded color Orange');
  assert(occContent.includes("'#E11D48'"), 'Must include expanded color Rose');
});

runTest('1.4: Incompatible units support dual-axis or independent scale for 3rd unit', () => {
  assert(occContent.includes('getSeriesScale'), 'Must provide getSeriesScale helper');
  assert(occContent.includes('this.trendScales.left'), 'Must support left axis scale');
  assert(occContent.includes('this.trendScales.right'), 'Must support right axis scale');
});

runTest('1.5: Parallel telemetry trend fetching with Promise.all for fast load', () => {
  assert(occContent.includes('await Promise.all(this.activeTrendSeries.map(async (s) =>'), 'Must parallelize activeTrendSeries requests');
});

// -------------------------------------------------------------
// REQUIREMENT 2: PERSIST USER SERIES SELECTION IN LOCALSTORAGE
// -------------------------------------------------------------

runTest('2.1: Versioned storage key isolated by user (operatorTrendSelection:v1:${userId})', () => {
  assert(occContent.includes('operatorTrendSelection:v1:'), 'Storage key must be versioned operatorTrendSelection:v1');
  assert(occContent.includes('currentUserIdentifier'), 'Must isolate key per user identifier');
});

runTest('2.2: Restoration waits until metadata is loaded to prevent race conditions', () => {
  assert(occContent.includes('trendSelectionRestored'), 'Must have trendSelectionRestored guard');
  assert(occContent.includes('if (!this.trendSelectionRestored) return;'), 'saveTrendSelection must not run before restoration');
  assert(occContent.includes('if (this.allParameters.length === 0) return;'), 'restoreOrInitTrendSelection must wait for parameters');
});

runTest('2.3: Deleted or disabled parameters in saved selection are skipped safely', () => {
  assert(occContent.includes('if (match && !restored.some(r => r.parameterId === match.param_id))'), 'Must only restore valid matching parameters');
  assert(occContent.includes('if (restored.length > 0)'), 'Must verify restored count > 0 before applying');
});

runTest('2.4: Empty invalid saved selection falls back to all eligible parameters', () => {
  assert(occContent.includes('this.selectAllEligibleTrendSeries(false)'), 'Must fallback to all eligible if saved entries invalid');
});

runTest('2.5: User selection actions update localStorage (selectParamForTrend, removeTrendSeries, clearAllTrendSeries)', () => {
  assert(occContent.includes('clearAllTrendSeries'), 'Must have clearAllTrendSeries action');
  assert(occContent.includes('resetTrendSelectionToDefault'), 'Must have resetTrendSelectionToDefault action');
  assert(occContent.includes('localStorage.removeItem(this.trendStorageKey)'), 'Reset must clear storage key and re-select defaults');
});

runTest('2.6: Compact Reset, All, and Clear controls available in UI', () => {
  assert(occContent.includes('@click="selectAllEligibleTrendSeries(true)"'), 'Must have All button');
  assert(occContent.includes('@click="clearAllTrendSeries"'), 'Must have Clear button');
  assert(occContent.includes('@click="resetTrendSelectionToDefault"'), 'Must have Reset button');
});

// -------------------------------------------------------------
// REQUIREMENT 3: PARAMETER CARD — CORRECT LAST UPDATE / FRESHNESS DISPLAY
// -------------------------------------------------------------

runTest('3.1: formatRelative eliminates flickering for 5-second polling cycle (diff < 10 returns Just now)', () => {
  assert(occContent.includes('if (diff < 10) return \'Just now\';'), 'diff < 10 must return Just now to eliminate 5s flicker');
  assert(occContent.includes('if (diff < 60) return `${diff}s ago`;'), 'diff < 60 must return seconds ago');
  assert(occContent.includes('if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;'), 'diff < 3600 must return minutes ago');
});

runTest('3.2: Failed polls or offline state preserve last successful reading timestamp (lastGoodTimeMap)', () => {
  assert(occContent.includes('this.lastGoodTimeMap[p.id]'), 'allParameters must use lastGoodTimeMap for last valid reading');
  assert(occContent.includes("isDevOffline || finalQuality === 'BAD' || finalQuality === 'STALE'"), 'Offline or bad parameters must display lastValidTime');
});

runTest('3.3: Tooltip preserves full source timestamp including ISO format and quality', () => {
  assert(occContent.includes('getParamTimingTitle(param)'), 'Must have getParamTimingTitle');
  assert(occContent.includes('timeIso'), 'Tooltip title must include ISO timestamp');
  assert(occContent.includes('new Date(param.received_at).toISOString()'), 'Must format ISO timestamp for source trace');
});

// -------------------------------------------------------------
// REQUIREMENT 4: ZERO-VALUE ALARM POLICY — CONFIGURABLE PER PARAMETER
// -------------------------------------------------------------

runTest('4.1: Backend ZeroValueAlarmPolicy enum defined in model with NORMAL default', () => {
  assert(modelDeviceContent.includes('ZeroValuePolicyNormal            ZeroValueAlarmPolicy = "NORMAL"'), 'Model must define NORMAL policy');
  assert(modelDeviceContent.includes('ZeroValuePolicySuppressThreshold ZeroValueAlarmPolicy = "SUPPRESS_THRESHOLD"'), 'Model must define SUPPRESS_THRESHOLD policy');
  assert(modelDeviceContent.includes('ZeroValuePolicyInvalidReading    ZeroValueAlarmPolicy = "INVALID_READING"'), 'Model must define INVALID_READING policy');
  assert(modelDeviceContent.includes('ZeroValueAlarmPolicy ZeroValueAlarmPolicy'), 'Parameter struct must contain ZeroValueAlarmPolicy');
  assert(modelDeviceContent.includes('p.ZeroValueAlarmPolicy = ZeroValuePolicyNormal'), 'BeforeSave must default empty policy to NORMAL');
});

runTest('4.2: Database migration ensures zero_value_alarm_policy column and backfills NORMAL', () => {
  assert(migrationContent.includes('zero_value_alarm_policy'), 'Migration must check for zero_value_alarm_policy column');
  assert(migrationContent.includes('AddColumn(&model.Parameter{}, "zero_value_alarm_policy")'), 'Migration must add column if missing');
  assert(migrationContent.includes('zero_value_alarm_policy IS NULL OR zero_value_alarm_policy = \'\''), 'Migration must backfill existing empty values');
});

runTest('4.3: DeviceService includes ZeroValueAlarmPolicy in Create and Update Parameter requests', () => {
  assert(/ZeroValueAlarmPolicy\s+\*model\.ZeroValueAlarmPolicy\s+`json:"zero_value_alarm_policy"`/.test(serviceDeviceContent), 'UpdateParameterRequest must have ZeroValueAlarmPolicy pointer');
  assert(serviceDeviceContent.includes('param.ZeroValueAlarmPolicy = *req.ZeroValueAlarmPolicy'), 'UpdateParameter must assign policy');
});

runTest('4.4: AlarmEvaluator implements SUPPRESS_THRESHOLD and INVALID_READING', () => {
  assert(alarmEvaluatorContent.includes('AlarmCodeAbnormalityZeroInvalid = "DATA_ABNORMALITY_ZERO_INVALID"'), 'Must define DATA_ABNORMALITY_ZERO_INVALID code');
  assert(alarmEvaluatorContent.includes('model.ZeroValuePolicySuppressThreshold'), 'AlarmEvaluator must evaluate SUPPRESS_THRESHOLD');
  assert(alarmEvaluatorContent.includes('model.ZeroValuePolicyInvalidReading'), 'AlarmEvaluator must evaluate INVALID_READING');
  assert(alarmEvaluatorContent.includes('Threshold alarms suppressed for parameter'), 'Must log debug record when threshold alarm is suppressed');
});

runTest('4.5: ParameterModal supports zero_value_alarm_policy selection and save', () => {
  assert(modalContent.includes('form.zero_value_alarm_policy'), 'ParameterModal must have form.zero_value_alarm_policy');
  assert(modalContent.includes('value="NORMAL"'), 'Modal must have NORMAL option');
  assert(modalContent.includes('value="SUPPRESS_THRESHOLD"'), 'Modal must have SUPPRESS_THRESHOLD option');
  assert(modalContent.includes('value="INVALID_READING"'), 'Modal must have INVALID_READING option');
  assert(modalContent.includes('zero_value_alarm_policy: this.form.zero_value_alarm_policy || \'NORMAL\''), 'Payload must include zero_value_alarm_policy');
});

runTest('4.6: determineAlarmState in OCC handles zero-value alarm policy', () => {
  assert(occContent.includes("policy === 'SUPPRESS_THRESHOLD'"), 'determineAlarmState must check SUPPRESS_THRESHOLD');
  assert(occContent.includes("policy === 'INVALID_READING'"), 'determineAlarmState must check INVALID_READING');
});

runTest('4.7: i18n locales contain zero-value alarm policy translations in EN and ID', () => {
  assert(enLocaleContent.includes('zeroValueAlarmPolicy'), 'EN must have zeroValueAlarmPolicy key');
  assert(enLocaleContent.includes('zeroPolicyNormal'), 'EN must have zeroPolicyNormal key');
  assert(enLocaleContent.includes('zeroPolicySuppress'), 'EN must have zeroPolicySuppress key');
  assert(enLocaleContent.includes('zeroPolicyInvalid'), 'EN must have zeroPolicyInvalid key');

  assert(idLocaleContent.includes('zeroValueAlarmPolicy'), 'ID must have zeroValueAlarmPolicy key');
  assert(idLocaleContent.includes('zeroPolicyNormal'), 'ID must have zeroPolicyNormal key');
  assert(idLocaleContent.includes('zeroPolicySuppress'), 'ID must have zeroPolicySuppress key');
  assert(idLocaleContent.includes('zeroPolicyInvalid'), 'ID must have zeroPolicyInvalid key');
});

console.log('\n===============================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
} else {
  process.exit(0);
}
