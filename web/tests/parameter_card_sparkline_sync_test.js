/**
 * Automated Verification Suite: Parameter Card Sparkline Data Synchronization
 *
 * Verifies:
 * 1. Telemetry baru memperbarui sparkline parameter yang benar.
 * 2. Parameter lain tidak terpengaruh (isolasi parameter).
 * 3. Timestamp tidak berurutan (out-of-order) ditangani dengan benar (chronological sort).
 * 4. Sampel duplikat tidak menimbulkan garis atau titik berulang (in-place update).
 * 5. Sparkline tidak terus bergerak ketika data berhenti (tidak ada fake ticks).
 * 6. Device offline mempertahankan histori tanpa membuat sampel baru.
 * 7. Sparkline diperbarui kembali setelah recovery tanpa merusak histori sebelumnya.
 * 8. Navigasi halaman tidak menggandakan event listener (mounted & beforeDestroy cleanup).
 * 9. Buffer tetap terbatas saat aplikasi berjalan lama (bounded FIFO max 40 samples).
 * 10. Data sparkline sesuai dengan sampel parameter yang sama di Trend Analysis (syncTrendPointsToSparkline).
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

console.log('===============================================================');
console.log('RUNNING TEST SUITE: PARAMETER CARD SPARKLINE DATA SYNCHRONIZATION');
console.log('===============================================================\n');

// Helper to create a test harness simulating OperatorCommandCenterView sparkline methods
function createHarness() {
  const harness = {
    devicesList: [
      { id: 1, device_code: 'AQMS_01', connection_status: 'ONLINE' },
      { id: 2, device_code: 'MET_01', connection_status: 'ONLINE' },
    ],
    sparklinesMap: {},
    sparklineMetaCache: {},
    configuredPollingIntervalMs: 5000,
    isDark: true,
    $set(target, key, val) {
      target[key] = val;
    },
  };

  // Extract normalizeTelemetryData
  const normMatch = occContent.match(/normalizeTelemetryData\(data\)\s*\{([\s\S]*?)\n\s*\},/);
  assert(normMatch, 'Must find normalizeTelemetryData in OCCView');
  harness.normalizeTelemetryData = new Function('data', normMatch[1]).bind(harness);

  // Extract recordSparkline
  const recMatch = occContent.match(/recordSparkline\(paramOrData,\s*val,\s*time,\s*quality\s*=\s*'GOOD',\s*deviceId\s*=\s*null\)\s*\{([\s\S]*?)\n\s*\},/);
  assert(recMatch, 'Must find recordSparkline in OCCView');
  harness.recordSparkline = new Function('paramOrData', 'val', 'time', 'quality', 'deviceId', recMatch[1]).bind(harness);

  // Extract syncTrendPointsToSparkline
  const syncMatch = occContent.match(/syncTrendPointsToSparkline\(paramId,\s*deviceId,\s*points\)\s*\{([\s\S]*?)\n\s*\},/);
  assert(syncMatch, 'Must find syncTrendPointsToSparkline in OCCView');
  harness.syncTrendPointsToSparkline = new Function('paramId', 'deviceId', 'points', syncMatch[1]).bind(harness);

  // Extract recomputeSparklineMeta
  const recomputeMatch = occContent.match(/recomputeSparklineMeta\(paramId\)\s*\{([\s\S]*?)\n\s*\},/);
  assert(recomputeMatch, 'Must find recomputeSparklineMeta in OCCView');
  harness.recomputeSparklineMeta = new Function('paramId', recomputeMatch[1]).bind(harness);

  // Extract getSparkline
  const getMatch = occContent.match(/getSparkline\(paramId\)\s*\{([\s\S]*?)\n\s*\},/);
  assert(getMatch, 'Must find getSparkline in OCCView');
  harness.getSparkline = new Function('paramId', getMatch[1]).bind(harness);

  return harness;
}

// -------------------------------------------------------------
// SCENARIO 1: TELEMETRI BARU MEMPERBARUI SPARKLINE PARAMETER YANG BENAR
// -------------------------------------------------------------
runTest('Scenario 1: Telemetry baru memperbarui sparkline parameter yang benar dengan processed_value', () => {
  const h = createHarness();
  const baseTime = 1700000000000;

  // Incoming packet has both raw value (245) and calibrated processed_value (24.5)
  h.recordSparkline({
    device_id: 1,
    parameter_id: 101,
    value: 245,
    processed_value: 24.5,
    received_at: new Date(baseTime).toISOString(),
    quality: 'GOOD',
  });

  const samples = h.sparklinesMap[101];
  assert(samples && samples.length === 1, 'Must create 1 sample in sparklinesMap for parameter 101');
  assert.strictEqual(samples[0].val, 24.5, 'Must record calibrated processed_value (24.5), NOT raw register value (245)');
  assert.strictEqual(samples[0].paramId, 101, 'Sample must have correct parameter ID');
  assert.strictEqual(samples[0].deviceId, 1, 'Sample must have correct device ID');

  const meta = h.getSparkline(101);
  assert(meta.hasHistory, 'Meta must mark hasHistory: true');
  assert(meta.strokePaths.length > 0, 'Must generate SVG stroke path');
});

// -------------------------------------------------------------
// SCENARIO 2: PARAMETER LAIN TIDAK TERPENGARUH
// -------------------------------------------------------------
runTest('Scenario 2: Pembaruan satu parameter tidak mengubah sparkline parameter lain', () => {
  const h = createHarness();
  const baseTime = 1700000000000;

  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 25.0, received_at: new Date(baseTime).toISOString() });
  h.recordSparkline({ device_id: 1, parameter_id: 102, processed_value: 60.0, received_at: new Date(baseTime).toISOString() });

  // Update only 101 with new reading
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 26.5, received_at: new Date(baseTime + 5000).toISOString() });

  assert.strictEqual(h.sparklinesMap[101].length, 2, 'Param 101 should have 2 points');
  assert.strictEqual(h.sparklinesMap[102].length, 1, 'Param 102 must remain untouched with exactly 1 point');
  assert.strictEqual(h.sparklinesMap[102][0].val, 60.0, 'Param 102 value must not be modified');
});

// -------------------------------------------------------------
// SCENARIO 3: TIMESTAMP TIDAK BERURUTAN (OUT-OF-ORDER) DITANGANI DENGAN BENAR
// -------------------------------------------------------------
runTest('Scenario 3: Timestamp tidak berurutan diurutkan secara kronologis (ascending)', () => {
  const h = createHarness();
  const t0 = 1700000000000;

  // Insert point at t + 20s
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 30.0, time: t0 + 20000 });
  // Insert point at t + 0s (out of order, earlier timestamp)
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 10.0, time: t0 });
  // Insert point at t + 10s (in between)
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 20.0, time: t0 + 10000 });

  const samples = h.sparklinesMap[101];
  assert.strictEqual(samples.length, 3, 'Must contain all 3 points');
  assert.strictEqual(samples[0].time, t0, 'Index 0 must be earliest timestamp');
  assert.strictEqual(samples[0].val, 10.0, 'Index 0 must have value 10.0');
  assert.strictEqual(samples[1].time, t0 + 10000, 'Index 1 must be middle timestamp');
  assert.strictEqual(samples[2].time, t0 + 20000, 'Index 2 must be latest timestamp');
});

// -------------------------------------------------------------
// SCENARIO 4: SAMPEL DUPLIKAT TIDAK MENIMBULKAN TITIK BERULANG
// -------------------------------------------------------------
runTest('Scenario 4: Sampel duplikat (polling ulang timestamp yang sama) diupdate in-place', () => {
  const h = createHarness();
  const t0 = 1700000000000;

  // First poll
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 24.0, time: t0, quality: 'GOOD' });
  // Repeated poll returning same reading and timestamp
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 24.2, time: t0, quality: 'GOOD' });

  const samples = h.sparklinesMap[101];
  assert.strictEqual(samples.length, 1, 'Must NOT append duplicate point for identical timestamp');
  assert.strictEqual(samples[0].val, 24.2, 'Should update existing point in-place with latest valid reading');
});

// -------------------------------------------------------------
// SCENARIO 5: SPARKLINE TIDAK TERUS BERGERAK KETIKA DATA BERHENTI
// -------------------------------------------------------------
runTest('Scenario 5: Sparkline tidak membuat sampel sintetis ketika data berhenti', () => {
  const h = createHarness();
  const t0 = 1700000000000;

  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 50.0, time: t0 });
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 52.0, time: t0 + 5000 });

  // Simulate multiple timer ticks or polls with no new data arriving (same timestamp)
  for (let i = 0; i < 5; i++) {
    h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 52.0, time: t0 + 5000 });
  }

  assert.strictEqual(h.sparklinesMap[101].length, 2, 'Must remain at exactly 2 points without synthetic growth');
  assert.strictEqual(h.sparklinesMap[101][1].time, t0 + 5000, 'Last timestamp must not drift forward');
});

// -------------------------------------------------------------
// SCENARIO 6: DEVICE OFFLINE MEMPERTAHANKAN HISTORI TANPA MEMBUAT SAMPEL BARU
// -------------------------------------------------------------
runTest('Scenario 6: Device offline mempertahankan histori valid terakhir tanpa membuat sampel baru', () => {
  const h = createHarness();
  const t0 = 1700000000000;

  // Valid telemetry while online
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 25.0, time: t0 });
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 26.0, time: t0 + 5000 });
  assert.strictEqual(h.sparklinesMap[101].length, 2);

  // Device transitions to OFFLINE
  h.devicesList[0].connection_status = 'OFFLINE';

  // Offline event / error poll attempt
  h.recordSparkline({
    device_id: 1,
    parameter_id: 101,
    value: null,
    time: t0 + 10000,
    quality: 'BAD',
  });

  const samples = h.sparklinesMap[101];
  assert.strictEqual(samples.length, 2, 'History must be preserved with exactly 2 valid points');
  assert.strictEqual(samples[1].val, 26.0, 'Last valid point must remain 26.0');

  const meta = h.getSparkline(101);
  assert(meta.hasHistory, 'Sparkline must retain valid historical visualization while offline');
});

// -------------------------------------------------------------
// SCENARIO 7: SPARKLINE DIPERBARUI KEMBALI SETELAH RECOVERY
// -------------------------------------------------------------
runTest('Scenario 7: Sparkline melanjutkan sampel valid setelah komunikasi pulih (recovery)', () => {
  const h = createHarness();
  const t0 = 1700000000000;

  // 1. Online points
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 25.0, time: t0 });
  h.recordSparkline({ device_id: 1, parameter_id: 101, processed_value: 26.0, time: t0 + 5000 });

  // 2. Offline outage for 60 seconds
  h.devicesList[0].connection_status = 'OFFLINE';
  h.recordSparkline({ device_id: 1, parameter_id: 101, value: null, time: t0 + 10000, quality: 'BAD' });

  // 3. Recovery: Device comes online and sends fresh valid packet at t0 + 65s
  h.devicesList[0].connection_status = 'ONLINE';
  h.recordSparkline({
    device_id: 1,
    parameter_id: 101,
    processed_value: 27.5,
    time: t0 + 65000,
    quality: 'GOOD',
  });

  const samples = h.sparklinesMap[101];
  assert.strictEqual(samples.length, 3, 'Must contain 2 pre-outage points + 1 post-recovery point');
  assert.strictEqual(samples[2].val, 27.5, 'New point must be recorded accurately');

  // Verify time gap was detected between t0+5s and t0+65s (60s > 20s threshold)
  const meta = h.getSparkline(101);
  assert(meta.hasHistory, 'Must have history');
  assert(meta.strokePaths.length >= 2, 'Must break sparkline into separate segments across outage gap');
});

// -------------------------------------------------------------
// SCENARIO 8: NAVIGASI HALAMAN TIDAK MENGGANDAKAN EVENT LISTENER
// -------------------------------------------------------------
runTest('Scenario 8: Event listeners terdaftar di mounted dan dibersihkan di beforeDestroy', () => {
  assert(occContent.includes("window.addEventListener('device-telemetry-event', this.onTelemetryEvent);"), 'Must register device-telemetry-event listener');
  assert(occContent.includes("window.removeEventListener('device-telemetry-event', this.onTelemetryEvent);"), 'Must unregister device-telemetry-event listener in beforeDestroy');
  assert(occContent.includes("window.addEventListener('device-comm-event', this.onDeviceCommEvent);"), 'Must register device-comm-event listener');
  assert(occContent.includes("window.removeEventListener('device-comm-event', this.onDeviceCommEvent);"), 'Must unregister device-comm-event listener in beforeDestroy');
});

// -------------------------------------------------------------
// SCENARIO 9: BUFFER TETAP TERBATAS SAAT APLIKASI BERJALAN LAMA
// -------------------------------------------------------------
runTest('Scenario 9: Buffer sparkline dibatasi maksimum 40 sampel untuk mencegah memory leak', () => {
  const h = createHarness();
  const t0 = 1700000000000;

  // Feed 100 consecutive live points
  for (let i = 0; i < 100; i++) {
    h.recordSparkline({
      device_id: 1,
      parameter_id: 101,
      processed_value: 20 + (i % 10),
      time: t0 + i * 5000,
    });
  }

  const samples = h.sparklinesMap[101];
  assert(samples.length <= 40, `Buffer size must not exceed 40 (got: ${samples.length})`);
  assert.strictEqual(samples.length, 40, 'Buffer must maintain exactly 40 most recent points');
  assert.strictEqual(samples[samples.length - 1].time, t0 + 99 * 5000, 'Last point must be newest timestamp');
});

// -------------------------------------------------------------
// SCENARIO 10: DATA SPARKLINE SESUAI DENGAN TREND ANALYSIS
// -------------------------------------------------------------
runTest('Scenario 10: Data sparkline sinkron dengan sampel parameter yang sama di Trend Analysis', () => {
  const h = createHarness();
  const t0 = 1700000000000;

  // Simulated points returned by Trend Analysis downsampled endpoint for parameter 101
  const trendPoints = [
    { x: t0, y: 15.2, quality: 'GOOD' },
    { x: t0 + 5000, y: 15.8, quality: 'GOOD' },
    { x: t0 + 10000, y: 16.1, quality: 'GOOD' },
    { x: t0 + 15000, y: 16.5, quality: 'GOOD' },
  ];

  // Synchronize Trend Analysis points to sparkline
  h.syncTrendPointsToSparkline(101, 1, trendPoints);

  const samples = h.sparklinesMap[101];
  assert.strictEqual(samples.length, 4, 'Sparkline buffer must have 4 synchronized points');
  for (let i = 0; i < trendPoints.length; i++) {
    assert.strictEqual(samples[i].time, trendPoints[i].x, `Point ${i} timestamp must match Trend Analysis`);
    assert.strictEqual(samples[i].val, trendPoints[i].y, `Point ${i} value must match Trend Analysis`);
  }

  // Verify fetchTrendData invokes syncTrendPointsToSparkline
  assert(occContent.includes('this.syncTrendPointsToSparkline(s.parameterId, s.deviceId, s.dataPoints);'), 'fetchTrendData must sync verified telemetry points to sparkline');
});

console.log('\n===============================================================');
console.log(`TEST RESULTS: ${passedTests} / ${totalTests} TESTS PASSED`);
console.log('===============================================================\n');

if (passedTests !== totalTests) {
  process.exit(1);
}
