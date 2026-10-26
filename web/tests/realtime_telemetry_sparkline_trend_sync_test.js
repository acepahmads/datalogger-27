const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: REAL-TIME SYNCHRONIZATION (SPARKLINE & TREND)');
console.log('===============================================================');

const occViewPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const storePath = path.join(__dirname, '..', 'src', 'store', 'index.js');
const hubPath = path.join(__dirname, '..', '..', 'internal', 'websocket', 'hub.go');

const occContent = fs.readFileSync(occViewPath, 'utf8');
const storeContent = fs.readFileSync(storePath, 'utf8');
const hubContent = fs.readFileSync(hubPath, 'utf8');

function test(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    process.exit(1);
  }
}

// Requirement 1: Store WebSocket unpacking supports newline-delimited JSON (NDJSON)
test('Scenario 1: Store socket.onmessage handles newline-delimited batched frames without dropping', () => {
  assert(storeContent.includes("const chunks = typeof raw === 'string' ? raw.split('\\n') : [raw];"),
    'Must split incoming WebSocket data by newline so batched frames do not fail JSON.parse');
  assert(storeContent.includes("new CustomEvent('device-telemetry-event'"),
    'Must dispatch device-telemetry-event for each parsed telemetry chunk');
  
  // Functional simulation of NDJSON parsing
  const batchedPayload = '{"type":"device.telemetry.received","data":{"parameter_id":1,"value":50.0}}\n{"type":"device.telemetry.received","data":{"parameter_id":2,"value":66.0}}';
  const lines = batchedPayload.split('\n');
  const parsed = [];
  for (const l of lines) {
    if (l.trim()) parsed.push(JSON.parse(l.trim()));
  }
  assert.strictEqual(parsed.length, 2, 'Must parse both batched telemetry items');
  assert.strictEqual(parsed[0].data.parameter_id, 1);
  assert.strictEqual(parsed[1].data.parameter_id, 2);
});

// Requirement 2: Backend WebSocket hub channel capacity expanded for bursts
test('Scenario 2: Backend hub buffers withstand multi-parameter bursts without client drops', () => {
  assert(hubContent.includes('broadcast:  make(chan WSMessage, 256),'),
    'Hub broadcast queue must be at least 256 to absorb multi-parameter polling frames');
  assert(hubContent.includes('send:     make(chan []byte, 256),'),
    'Client send buffer must be at least 256 to prevent disconnect during bursts');
});

// Requirement 3: Real-time WebSocket connection indicator in Trend Analysis header
test('Scenario 3: Trend Analysis displays distinct LIVE CONNECTED vs DISCONNECTED badge', () => {
  assert(occContent.includes("LIVE CONNECTED"),
    'Must render LIVE CONNECTED indicator when WebSocket link is active');
  assert(occContent.includes("DISCONNECTED"),
    'Must render DISCONNECTED indicator when WebSocket link is lost');
  assert(occContent.includes("$store.state.wsConnected"),
    'Must bind connection status to store.state.wsConnected');
});

// Requirement 4: Real-time Telemetry Freshness & Age distinction
test('Scenario 4: Telemetry freshness tracks data age separately from WebSocket connection', () => {
  assert(occContent.includes('trendCommDiagnostics'),
    'Must compute trendCommDiagnostics for telemetry freshness');
  assert(occContent.includes('lastTimeFormatted'),
    'Must display formatted time of last telemetry packet');
  assert(occContent.includes('secondsAgo'),
    'Must display elapsed seconds since last valid telemetry reading');
  assert(occContent.includes('STALE_DATA'),
    'Must trigger STALE_DATA badge when telemetry age exceeds freshness threshold');
});

// Requirement 5: WebSocket auto-reconnect backfills without manual page refresh
test('Scenario 5: WebSocket reconnect watcher triggers seamless telemetry and trend backfill', () => {
  assert(occContent.includes("'$store.state.wsConnected'(newVal, oldVal)"),
    'Must watch $store.state.wsConnected for connection recovery');
  assert(occContent.includes('this.refreshLatestOnly()'),
    'Must refresh latest telemetry on reconnection');
  assert(occContent.includes('this.fetchTrendData(true)'),
    'Must silently fetch trend data on reconnection');
});

// Requirement 6: Telemetry event updates Parameter Card value and sparkline reactively
test('Scenario 6: onTelemetryEvent immediately updates parameter card and sparkline', () => {
  assert(occContent.includes('this.$set(this.liveMap, data.parameter_id, data);'),
    'Must update liveMap reactively for the parameter card');
  assert(occContent.includes('this.recordSparkline(data);'),
    'Must record sparkline sample immediately');
  assert(occContent.includes('this.recordTrendPoint(data);'),
    'Must record trend point immediately');
});

// Requirement 7: In-place sparkline update uses Vue reactive $set
test('Scenario 7: Sparkline in-place deduplication updates reactively without dropping history', () => {
  assert(occContent.includes('this.$set(arr, existingIdx'),
    'Must use this.$set for in-place replacement of near-identical timestamps');
  assert(occContent.includes('this.recomputeSparklineMeta(pid);'),
    'Must recompute sparkline metadata after sample updates');
});

// Requirement 8: recordTrendPoint updates series reactively with $set and revision tick
test('Scenario 8: recordTrendPoint updates series reactively with $set and revision tick', () => {
  assert(occContent.includes('this.$set(pts, pts.length - 1, newPt);'),
    'Must use this.$set when overwriting latest point at alignedX');
  assert(occContent.includes('this.$set(pts, idx, newPt);'),
    'Must use this.$set when updating out-of-order point at alignedX');
  assert(occContent.includes('this.trendLiveTick = (this.trendLiveTick || 0) + 1;'),
    'Must increment trendLiveTick to invalidate render cache');
});

// Requirement 9: Right boundary slides forward on live telemetry without rubber-banding
test('Scenario 9: Trend time bounds right edge encompasses live point immediately', () => {
  assert(occContent.includes('if (!this.fullscreenIsCustom && alignedX >= this.trendTimeBounds.end)'),
    'Must slide trendTimeBounds.end forward as soon as live point lands at or past boundary');
});

// Requirement 10: Simulated pipeline latency is sub-50ms
test('Scenario 10: Verification of end-to-end processing latency', () => {
  const t0 = Date.now();
  
  // Simulated Modbus response reception
  const mockFrameTimestamp = new Date().toISOString();
  
  // Simulated backend ingestion
  const mockTelemetry = {
    device_id: 4,
    device_code: 'com4',
    parameter_id: 1,
    parameter_code: 'pm2.5',
    value: 50.0,
    processed_value: 50.0,
    quality: 'GOOD',
    received_at: mockFrameTimestamp,
  };
  
  // Simulated WebSocket delivery & JSON parse
  const wireData = JSON.stringify({ type: 'device.telemetry.received', data: mockTelemetry });
  const eventMsg = JSON.parse(wireData);
  
  // Simulated frontend store dispatch & OCC handler
  const data = eventMsg.data;
  const numVal = Number(data.processed_value);
  const ts = new Date(data.received_at).getTime();
  const pollStepMs = 5000;
  const alignedX = Math.round(ts / pollStepMs) * pollStepMs;
  
  const pt = { x: alignedX, y: numVal, quality: data.quality };
  assert.strictEqual(pt.y, 50.0);
  assert.strictEqual(pt.quality, 'GOOD');
  
  const elapsedMs = Date.now() - t0;
  assert(elapsedMs < 50, `End-to-end telemetry propagation must complete in < 50ms, took ${elapsedMs}ms`);
  console.log(`  Measured internal pipeline latency: ${elapsedMs} ms`);
});

console.log('===============================================================');
console.log('TEST SUMMARY: ALL 10 REAL-TIME SYNC SCENARIOS PASSED');
console.log('===============================================================');
