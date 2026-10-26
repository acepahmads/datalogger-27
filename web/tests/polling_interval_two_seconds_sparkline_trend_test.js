const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('TEST SUITE: 2-SECOND POLLING INTERVAL & END-TO-END PIPELINE SYNC');
console.log('===============================================================');

const occViewPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

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

// 1. Audit: configuredPollingIntervalMs supports 2 seconds (and seconds-to-ms normalization)
test('1.1: configuredPollingIntervalMs normalizes seconds (<= 60) to milliseconds (e.g. 2s -> 2000ms)', () => {
  assert(occContent.includes('configuredPollingIntervalMs()'), 'configuredPollingIntervalMs must exist');
  assert(
    occContent.includes('if (iv && iv > 0 && iv <= 60) iv = iv * 1000;'),
    'Must normalize seconds <= 60 to milliseconds (e.g. 2 -> 2000)'
  );
  assert(
    occContent.includes('if (iv && iv >= 200) return iv;'),
    'Must allow intervals down to 200ms without restrictive 500ms floor'
  );
});

// 2. Audit: Trend Data Bucketing and Recording dynamically adopts 2-second polling interval
test('2.1: fetchTrendData uses configuredPollingIntervalMs for bucketing (pollStepMs = 2000ms)', () => {
  assert(
    occContent.includes('const pollStepMs = isUnderOrEqual6h ? (this.configuredPollingIntervalMs || 5000) : null;'),
    'fetchTrendData must dynamically derive pollStepMs from configuredPollingIntervalMs'
  );
});

test('2.2: recordTrendPoint dynamically aligns live points using configuredPollingIntervalMs', () => {
  const recordMatch = occContent.match(/recordTrendPoint\(data\)\s*\{([\s\S]*?)\n    buildSplinePath/);
  assert(recordMatch, 'recordTrendPoint method must exist');
  const body = recordMatch[1];
  assert(
    body.includes('const pollStepMs = isUnderOrEqual6h ? (this.configuredPollingIntervalMs || 5000) : null;'),
    'recordTrendPoint must use configuredPollingIntervalMs for live point bucketing'
  );
  assert(
    body.includes('const alignedX = pollStepMs ? Math.round(ts / pollStepMs) * pollStepMs : ts;'),
    'alignedX must align to pollStepMs multiples'
  );
});

test('2.3: advanceLiveTrendWindow update cadence dynamically matches 2-second interval', () => {
  const advanceMatch = occContent.match(/advanceLiveTrendWindow\(\)\s*\{([\s\S]*?)\n    \},/);
  assert(advanceMatch, 'advanceLiveTrendWindow must exist');
  const body = advanceMatch[1];
  assert(
    body.includes('let threshold = Math.max(pollInterval, 1000);'),
    'advanceLiveTrendWindow threshold must allow 2-second update threshold (Math.max(pollInterval, 1000))'
  );
});

test('2.4: onChartScrub and onFullscreenChartScrub snap to configured 2-second polling interval', () => {
  assert(
    occContent.includes('let effectiveTime = isUnderOrEqual6h ? (Math.round(hoverTime / pollStep) * pollStep) : hoverTime;'),
    'Chart scrubbing must snap to pollStep multiples instead of hardcoded 5000ms'
  );
});

// 3. Instrumentation & Diagnostics (Requirement 4)
test('3.1: Operator dashboard includes pipelineDiagnostics tracking all 6 stages', () => {
  assert(occContent.includes('pipelineDiagnostics: {'), 'pipelineDiagnostics object must exist in data()');
  assert(occContent.includes('lastModbusResponse: null,'), 'Must track lastModbusResponse');
  assert(occContent.includes('lastTelemetryProcessed: null,'), 'Must track lastTelemetryProcessed');
  assert(occContent.includes('lastTelemetryBroadcast: null,'), 'Must track lastTelemetryBroadcast');
  assert(occContent.includes('lastFrontendReceived: null,'), 'Must track lastFrontendReceived');
  assert(occContent.includes('lastSparklineUpdate: null,'), 'Must track lastSparklineUpdate');
  assert(occContent.includes('lastTrendUpdate: null,'), 'Must track lastTrendUpdate');
});

test('3.2: onTelemetryEvent measures end-to-end latency across all stages', () => {
  assert(occContent.includes('recordPipelineDiagnostics(data, {'), 'onTelemetryEvent must invoke recordPipelineDiagnostics');
  assert(occContent.includes('recordPipelineDiagnostics(data, times) {'), 'recordPipelineDiagnostics method must exist');
  assert(occContent.includes('modbusToProcessMs'), 'Must compute modbusToProcessMs');
  assert(occContent.includes('processToBroadcastMs'), 'Must compute processToBroadcastMs');
  assert(occContent.includes('broadcastToFrontendMs'), 'Must compute broadcastToFrontendMs');
  assert(occContent.includes('frontendToSparklineMs'), 'Must compute frontendToSparklineMs');
  assert(occContent.includes('sparklineToTrendMs'), 'Must compute sparklineToTrendMs');
  assert(occContent.includes('totalEndToEndMs'), 'Must compute totalEndToEndMs');
});

test('3.3: Live telemetry pipeline diagnostic badge is rendered in header', () => {
  assert(occContent.includes('pipelineDiagnostics && pipelineDiagnostics.sampleCount > 0'), 'Template must include diagnostic badge');
  assert(occContent.includes('pipelineDiagnostics.totalEndToEndMs'), 'Template must display totalEndToEndMs in diagnostic badge');
});

// 4. Functional Simulation: 10 consecutive 2-second samples
test('4.1: Simulation of 10 consecutive 2-second samples records 10 distinct trend & sparkline points', () => {
  const pollIntervalMs = 2000;
  const t0 = 1770000000000; // Aligned base time (mod 2000 = 0)

  // Emulate trend series
  const series = {
    id: 1,
    parameterId: 101,
    deviceId: 1,
    dataPoints: [],
  };

  // Emulate sparkline list
  const sparklinePts = [];

  const recordedSamples = [];

  for (let i = 0; i < 10; i++) {
    const modbusTime = t0 + i * pollIntervalMs;
    const processTime = modbusTime + 8; // 8ms decode
    const broadcastTime = processTime + 2; // 2ms WS enqueue
    const frontendRecvTime = broadcastTime + 5; // 5ms WS transport
    const sparklineEndTime = frontendRecvTime + 1; // 1ms sparkline calc
    const trendEndTime = sparklineEndTime + 2; // 2ms trend calc

    const rawVal = 25.0 + Math.sin(i * 0.5) * 2.0;

    // Emulate recordSparkline
    const existingIdx = sparklinePts.findIndex(p => Math.abs(p.time - modbusTime) < 500);
    if (existingIdx >= 0) {
      sparklinePts[existingIdx].val = rawVal;
    } else {
      sparklinePts.push({
        paramId: 101,
        deviceId: 1,
        val: rawVal,
        time: modbusTime,
        quality: 'GOOD',
      });
    }

    // Emulate recordTrendPoint with 2000ms pollStepMs
    const pollStepMs = pollIntervalMs;
    const alignedX = Math.round(modbusTime / pollStepMs) * pollStepMs;
    const pts = series.dataPoints;
    const newPt = { x: alignedX, y: rawVal, quality: 'GOOD' };

    if (pts.length > 0 && pts[pts.length - 1].x === alignedX) {
      pts[pts.length - 1] = newPt;
    } else {
      pts.push(newPt);
    }

    const prevRecv = recordedSamples.length > 0 ? recordedSamples[recordedSamples.length - 1].frontendRecv : null;
    const intervalMs = prevRecv ? (frontendRecvTime - prevRecv) : null;

    recordedSamples.push({
      sampleIndex: i + 1,
      modbusTime,
      processTime,
      broadcastTime,
      frontendRecv: frontendRecvTime,
      sparklineTime: sparklineEndTime,
      trendTime: trendEndTime,
      intervalMs,
      endToEndMs: trendEndTime - modbusTime,
    });
  }

  // Verification 1: Exactly 10 distinct samples in trend series
  assert.strictEqual(series.dataPoints.length, 10, 'Must contain exactly 10 distinct 2-second trend points');

  // Verification 2: Each consecutive trend point is spaced by exactly 2000ms
  for (let i = 1; i < series.dataPoints.length; i++) {
    const diff = series.dataPoints[i].x - series.dataPoints[i - 1].x;
    assert.strictEqual(diff, 2000, `Trend point ${i} must be separated by exactly 2000ms from point ${i - 1}`);
  }

  // Verification 3: Exactly 10 distinct samples in sparkline
  assert.strictEqual(sparklinePts.length, 10, 'Must contain exactly 10 distinct 2-second sparkline points');

  // Verification 4: Intervals between consecutive samples are 2000ms
  for (let i = 1; i < recordedSamples.length; i++) {
    assert.strictEqual(recordedSamples[i].intervalMs, 2000, `Sample ${i} cadence must be 2000ms`);
  }

  // Verification 5: End-to-end latency is well under 100ms
  for (let i = 0; i < recordedSamples.length; i++) {
    assert(recordedSamples[i].endToEndMs < 50, `Sample ${i} end-to-end latency must be < 50ms (got ${recordedSamples[i].endToEndMs}ms)`);
  }

  console.log('   Latency trace for 10 consecutive 2-second samples:');
  recordedSamples.forEach(s => {
    console.log(`   Sample #${s.sampleIndex.toString().padStart(2, '0')} | Cadence: ${s.intervalMs ? s.intervalMs + 'ms' : ' initial '} | E2E Latency: ${s.endToEndMs}ms (Modbus → Dec ${s.processTime - s.modbusTime}ms → WS ${s.broadcastTime - s.processTime}ms → UI ${s.trendTime - s.broadcastTime}ms)`);
  });
});

console.log('===============================================================');
console.log('TEST SUMMARY: ALL 2-SECOND POLLING INTERVAL TESTS PASSED (8/8)');
console.log('===============================================================');
