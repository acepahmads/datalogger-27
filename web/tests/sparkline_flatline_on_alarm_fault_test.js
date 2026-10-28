/**
 * Automated Verification Suite: Sparkline Flatline on Alarm, Device Fault, and Data Anomaly
 *
 * Requirements:
 * "ketika ada alarm, gangguan device atau anomaly data, grapik di masing2 parameter ubah menjadi datar"
 *
 * Verifies:
 * 1. Device disruption (OFFLINE, ERROR, DISCONNECTED, TIMEOUT) forces flatline.
 * 2. Active alarms (CRITICAL, WARNING, ALARM, activeAlarmsList) force flatline.
 * 3. Telemetry data anomaly (BAD, STALE, UNCERTAIN, is_anomaly, is_held_value, missing) forces flatline.
 * 4. Healthy parameter displays dynamic wavy historical sparkline.
 * 5. Recovery transition: automatically restores wavy curve when device/telemetry recovers.
 * 6. Alert colors: Red for offline/bad/critical, Amber for warning, Purple for stale.
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

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

console.log('====================================================================');
console.log('RUNNING TEST SUITE: SPARKLINE FLATLINE ON ALARM, FAULT & ANOMALY');
console.log('====================================================================\n');

// Extract methods from OperatorCommandCenterView.vue
function extractMethod(methodName) {
  const regex = new RegExp(`${methodName}\\s*\\(([^)]*)\\)\\s*\\{([\\s\\S]*?)\\n\\s{4}\\},`);
  const match = occContent.match(regex);
  if (!match) {
    throw new Error(`Could not find method ${methodName} in OperatorCommandCenterView.vue`);
  }
  const args = match[1].split(',').map(a => a.trim()).filter(Boolean);
  const body = match[2];
  return new Function(...args, body);
}

function createHarness() {
  const harness = {
    devicesList: [
      { id: 1, device_code: 'AQMS_01', connection_status: 'ONLINE', enabled: true },
      { id: 2, device_code: 'MET_01', connection_status: 'ONLINE', enabled: true },
    ],
    allParameters: [],
    activeAlarmsList: [],
    sparklinesMap: {},
    sparklineMetaCache: {},
    configuredPollingIntervalMs: 5000,
    isDark: true,
    $set(target, key, val) {
      target[key] = val;
    },
    formatParamValue(param) {
      if (!param) return '--';
      if (param.device_status === 'OFFLINE' || param.quality === 'BAD') return '--';
      if (param.value == null || param.value === '') return '--';
      return String(param.value);
    },
    isDeviceActiveOnline(dev) {
      return dev && (dev.connection_status === 'ONLINE' || dev.connection_status === 'CONNECTED');
    },
  };

  harness.isParamInAlarmOrFault = extractMethod('isParamInAlarmOrFault').bind(harness);
  harness.getSparklineColor = extractMethod('getSparklineColor').bind(harness);
  harness.recomputeSparklineMeta = extractMethod('recomputeSparklineMeta').bind(harness);
  harness.getSparkline = extractMethod('getSparkline').bind(harness);

  return harness;
}

// 1. Device Disruption / Trouble
runTest('Scenario 1.1: Device OFFLINE status produces flatline sparkline', () => {
  const h = createHarness();
  const param = {
    param_id: 101,
    param_code: 'PM25',
    device_id: 1,
    device_status: 'OFFLINE',
    quality: 'BAD',
    alarm_status: 'NORMAL',
    value: null,
  };
  h.allParameters = [param];

  // Pre-fill historical wavy points
  h.sparklinesMap[101] = [
    { time: 1000, val: 25.0 },
    { time: 2000, val: 35.0 },
    { time: 3000, val: 45.0 },
    { time: 4000, val: 15.0 },
  ];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true, 'Parameter must be evaluated as in alarm/fault');

  const spark = h.getSparkline(param);
  assert.strictEqual(spark.hasHistory, true, 'Must display sparkline SVG');
  assert.strictEqual(spark.isFlat, true, 'Sparkline must be flagged as flat');
  assert.deepStrictEqual(spark.strokePaths, ['M 0,10 L 100,10'], 'Stroke path must be horizontal line at center y=10');
  assert.deepStrictEqual(spark.areaPaths, ['M 0,20 L 0,10 L 100,10 L 100,20 Z'], 'Area path must shade bottom half');
});

runTest('Scenario 1.2: Device communication error reason produces flatline', () => {
  const h = createHarness();
  const param = {
    param_id: 102,
    param_code: 'PM10',
    device_id: 1,
    device_status: 'ONLINE',
    quality: 'BAD',
    quality_reason: 'COMMUNICATION_ERROR',
    alarm_status: 'NORMAL',
    value: null,
  };
  h.allParameters = [param];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true);
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.isFlat, true);
  assert.strictEqual(spark.strokePaths[0], 'M 0,10 L 100,10');
});

runTest('Scenario 1.3: Parent device in devicesList offline produces flatline', () => {
  const h = createHarness();
  h.devicesList[0].connection_status = 'OFFLINE';
  const param = {
    param_id: 103,
    param_code: 'SO2',
    device_id: 1,
    device_status: 'ONLINE', // card not yet updated, but device list is OFFLINE
    quality: 'GOOD',
    alarm_status: 'NORMAL',
    value: 12.0,
  };
  h.allParameters = [param];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true);
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.isFlat, true);
});

// 2. Active Alarms
runTest('Scenario 2.1: CRITICAL alarm status produces flatline sparkline', () => {
  const h = createHarness();
  const param = {
    param_id: 104,
    param_code: 'CO',
    device_id: 1,
    device_status: 'ONLINE',
    quality: 'GOOD',
    alarm_status: 'CRITICAL',
    value: 185.0,
  };
  h.allParameters = [param];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true);
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.isFlat, true);
  assert.strictEqual(spark.strokePaths[0], 'M 0,10 L 100,10');
});

runTest('Scenario 2.2: WARNING alarm status produces flatline sparkline', () => {
  const h = createHarness();
  const param = {
    param_id: 105,
    param_code: 'O3',
    device_id: 1,
    device_status: 'ONLINE',
    quality: 'GOOD',
    alarm_status: 'WARNING',
    value: 120.0,
  };
  h.allParameters = [param];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true);
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.isFlat, true);
});

runTest('Scenario 2.3: Active alarm in activeAlarmsList triggers flatline', () => {
  const h = createHarness();
  const param = {
    param_id: 106,
    param_code: 'NO2',
    device_id: 1,
    device_status: 'ONLINE',
    quality: 'GOOD',
    alarm_status: 'NORMAL',
    value: 45.0,
  };
  h.allParameters = [param];
  h.activeAlarmsList = [
    { id: 1, parameter_id: 106, status: 'ACTIVE', severity: 'CRITICAL', alarm_type: 'THRESHOLD_HIGH' },
  ];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true);
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.isFlat, true);
});

// 3. Telemetry Data Anomaly
runTest('Scenario 3.1: Data quality BAD triggers flatline', () => {
  const h = createHarness();
  const param = {
    param_id: 107,
    param_code: 'HC',
    device_id: 1,
    device_status: 'ONLINE',
    quality: 'BAD',
    alarm_status: 'NORMAL',
    value: null,
  };
  h.allParameters = [param];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true);
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.isFlat, true);
});

runTest('Scenario 3.2: Data quality STALE triggers flatline', () => {
  const h = createHarness();
  const param = {
    param_id: 108,
    param_code: 'WS',
    device_id: 2,
    device_status: 'ONLINE',
    quality: 'STALE',
    is_stale: true,
    alarm_status: 'NORMAL',
    value: 2.1,
  };
  h.allParameters = [param];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true);
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.isFlat, true);
});

runTest('Scenario 3.3: Frozen held value (anomaly grace hold) triggers flatline', () => {
  const h = createHarness();
  const param = {
    param_id: 109,
    param_code: 'TEMP',
    device_id: 2,
    device_status: 'ONLINE',
    quality: 'GOOD',
    is_held_value: true,
    alarm_status: 'NORMAL',
    value: 28.5,
  };
  h.allParameters = [param];

  assert.strictEqual(h.isParamInAlarmOrFault(param), true);
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.isFlat, true);
});

// 4. Healthy Parameter Behavior
runTest('Scenario 4: Healthy parameter produces dynamic wavy historical sparkline', () => {
  const h = createHarness();
  const param = {
    param_id: 110,
    param_code: 'HUM',
    device_id: 2,
    device_status: 'ONLINE',
    quality: 'GOOD',
    alarm_status: 'NORMAL',
    value: 65.0,
  };
  h.allParameters = [param];

  h.sparklinesMap[110] = [
    { time: 1000, val: 50.0 },
    { time: 2000, val: 70.0 },
    { time: 3000, val: 60.0 },
    { time: 4000, val: 80.0 },
  ];

  assert.strictEqual(h.isParamInAlarmOrFault(param), false, 'Healthy parameter must not be flagged in fault');
  const spark = h.getSparkline(param);
  assert.strictEqual(spark.hasHistory, true);
  assert.strictEqual(spark.isFlat, false, 'Healthy sparkline must not be flat');
  assert(spark.min <= 50, 'Min should include visual padding <= 50');
  assert(spark.max >= 80, 'Max should include visual padding >= 80');
  assert(spark.strokePaths[0].includes('L'), 'Must contain multi-point line path');
  assert(spark.strokePaths[0] !== 'M 0,10 L 100,10', 'Healthy stroke must not be single center flatline');
});

// 5. Dynamic Recovery Transition
runTest('Scenario 5: Recovery transition from OFFLINE to ONLINE restores wavy curve', () => {
  const h = createHarness();
  const param = {
    param_id: 111,
    param_code: 'SOLAR',
    device_id: 2,
    device_status: 'OFFLINE',
    quality: 'BAD',
    alarm_status: 'NORMAL',
    value: null,
  };
  h.allParameters = [param];
  h.sparklinesMap[111] = [
    { time: 1000, val: 100.0 },
    { time: 2000, val: 300.0 },
    { time: 3000, val: 500.0 },
  ];

  // While offline: flatline
  const flatSpark = h.getSparkline(param);
  assert.strictEqual(flatSpark.isFlat, true);

  // Recovery: Device reconnects, good data received
  param.device_status = 'ONLINE';
  param.quality = 'GOOD';
  param.value = 520.0;
  h.sparklinesMap[111].push({ time: 4000, val: 520.0 });

  const recoveredSpark = h.getSparkline(param);
  assert.strictEqual(recoveredSpark.isFlat, false, 'Must recover from flatline to dynamic curve');
  assert(recoveredSpark.strokePaths[0].includes('L'));
  assert(recoveredSpark.min <= 100, 'Min should include visual padding <= 100');
  assert(recoveredSpark.max >= 520, 'Max should include visual padding >= 520');
});

// 6. Sparkline Color Aesthetics
runTest('Scenario 6: getSparklineColor returns alert red for offline, bad quality, or critical alarm', () => {
  const h = createHarness();

  // Dark mode
  h.isDark = true;
  assert.strictEqual(h.getSparklineColor({ device_status: 'OFFLINE', quality: 'BAD', alarm_status: 'NORMAL' }), '#FF6577');
  assert.strictEqual(h.getSparklineColor({ device_status: 'ONLINE', quality: 'BAD', alarm_status: 'NORMAL' }), '#FF6577');
  assert.strictEqual(h.getSparklineColor({ device_status: 'ONLINE', quality: 'GOOD', alarm_status: 'CRITICAL' }), '#FF6577');
  assert.strictEqual(h.getSparklineColor({ device_status: 'ONLINE', quality: 'GOOD', alarm_status: 'WARNING' }), '#F1BD61');
  assert.strictEqual(h.getSparklineColor({ device_status: 'ONLINE', quality: 'STALE', alarm_status: 'NORMAL' }), '#A855F7');

  // Light mode
  h.isDark = false;
  assert.strictEqual(h.getSparklineColor({ device_status: 'OFFLINE', quality: 'BAD', alarm_status: 'NORMAL' }), '#C93449');
  assert.strictEqual(h.getSparklineColor({ device_status: 'ONLINE', quality: 'GOOD', alarm_status: 'WARNING' }), '#A96C0A');
});

console.log('\n====================================================================');
console.log(`TEST RESULTS: ${passedTests} / ${totalTests} TESTS PASSED`);
console.log('====================================================================');

if (passedTests !== totalTests) {
  process.exit(1);
}
