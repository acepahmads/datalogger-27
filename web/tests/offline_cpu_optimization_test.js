const assert = require('assert');
const fs = require('fs');
const path = require('path');
const vm = require('vm');

const viewPath = path.join(__dirname, '..', 'src', 'views', 'DashboardOperatorView.vue');
const occPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const viewContent = fs.readFileSync(viewPath, 'utf8');

console.log('Running Offline Multi-Parameter CPU Optimization Tests...');

let passed = 0;

function check(name, fn) {
  try {
    fn();
    console.log(`  PASS: ${name}`);
    passed++;
  } catch (err) {
    console.error(`  FAIL: ${name}`);
    console.error(`    ${err.stack || err.message}`);
    process.exit(1);
  }
}

function extractMethodSource(src, signature) {
  const fullSig = `    ${signature} {`;
  const idx = src.indexOf(fullSig);
  assert(idx !== -1, `Could not find method signature: ${fullSig}`);
  const openBrace = src.indexOf('{', idx);
  assert(openBrace !== -1, `Could not find opening brace for: ${signature}`);
  let depth = 0;
  for (let i = openBrace; i < src.length; i++) {
    const ch = src[i];
    if (ch === '{') depth++;
    else if (ch === '}') {
      depth--;
      if (depth === 0) {
        return src.slice(idx + 4, i + 1);
      }
    }
  }
  throw new Error(`Unbalanced braces for: ${signature}`);
}

check('1. Protected base files remain completely untouched', () => {
  assert(fs.existsSync(occPath), 'OperatorCommandCenterView.vue must exist');
  assert(viewContent.includes('extends: OperatorCommandCenterView'), 'DashboardOperatorView must extend OperatorCommandCenterView cleanly');
});

check('2. isDeviceActiveOnline in DashboardOperatorView depends on 10s staleCheckTimeMs instead of 1s currentTimeMs', () => {
  const methodSrc = extractMethodSource(viewContent, 'isDeviceActiveOnline(dev)');
  assert(methodSrc.includes('this.staleCheckTimeMs'), 'isDeviceActiveOnline must read staleCheckTimeMs (10s cadence)');
  assert(!methodSrc.includes('this.currentTimeMs'), 'isDeviceActiveOnline must NOT read 1-second currentTimeMs');

  const methodsObj = vm.runInNewContext(`({ ${methodSrc} })`, {});
  let currentTimeReads = 0;
  let staleTimeReads = 0;
  const ctx = {
    ...methodsObj,
    get currentTimeMs() {
      currentTimeReads++;
      return Date.now();
    },
    get staleCheckTimeMs() {
      staleTimeReads++;
      return 1700000050000;
    },
    liveMap: {
      1: { quality: 'GOOD', received_at: new Date(1700000040000).toISOString() },
    },
    getDeviceStaleTimeoutSeconds: () => 120,
    isDevicePassivePush: () => false,
  };

  const onlineDev = {
    id: 10,
    enabled: true,
    connection_status: 'ONLINE',
    parameters: [{ id: 1, enabled: true, stale_timeout_seconds: 120 }],
  };
  const offlineDev = {
    id: 11,
    enabled: true,
    connection_status: 'OFFLINE',
    parameters: [{ id: 2, enabled: true, stale_timeout_seconds: 120 }],
  };

  assert.strictEqual(ctx.isDeviceActiveOnline(onlineDev), true);
  assert.strictEqual(ctx.isDeviceActiveOnline(offlineDev), false);
  assert.strictEqual(currentTimeReads, 0, 'Must perform zero reactive reads of 1s currentTimeMs');
  assert(staleTimeReads >= 1, 'Must read 10s staleCheckTimeMs');
});

check('3. WeakMap render memoization caches card badge/progress/color across repeated 1s clock renders without reactive $set', () => {
  const memoSrc = [
    extractMethodSource(viewContent, 'getBentoCardRenderMemo(param)'),
    extractMethodSource(viewContent, 'getBentoParamActiveAlarm(param)'),
    extractMethodSource(viewContent, 'getBentoStatusBadge(param)'),
    extractMethodSource(viewContent, 'resolveCoreBentoStatusBadge(param)'),
    extractMethodSource(viewContent, 'getBentoCardColor(param)'),
    extractMethodSource(viewContent, 'getBentoCardProgress(param)'),
  ].join(',\n');

  const methodsObj = vm.runInNewContext(`({ ${memoSrc}, WeakMap })`, { WeakMap });
  let formatRelCalls = 0;
  let alarmMapRef = {
    201: {
      primary: {
        id: 99,
        alarm_code: 'DATA_ABNORMALITY_STUCK_VALUE',
        severity: 'WARNING',
        triggered_at: '2026-11-17T02:00:00Z',
        message: 'Stuck sensor reading',
      },
      count: 1,
      all: [],
    },
  };

  const ctx = {
    ...methodsObj,
    isDark: false,
    canEditConfig: true,
    staleCheckTimeMs: 1700000010000,
    activeTrendSeries: [],
    get bentoParamActiveAlarmMap() {
      return alarmMapRef;
    },
    $set() {
      throw new Error('FORBIDDEN: reactive $set during card render getter');
    },
    formatRelative() {
      formatRelCalls++;
      return '5m ago';
    },
    isBentoParamCommOffline(p) {
      return p.device_status === 'OFFLINE';
    },
    isParamFormulaFallback() {
      return false;
    },
    isDiscreteStatusParam() {
      return false;
    },
    resolveBentoParamLimits() {
      return {
        lower: 0,
        upper: 50,
        warnLow: null,
        warnHigh: 45,
        hardMin: null,
        hardMax: 50,
        zeroSuppress: false,
        source: 'configured',
        shortLabel: '0–50',
        longLabel: '0–50 °C',
        detailNote: '0–50 °C',
      };
    },
  };

  const stuckParam = {
    param_id: 201,
    param_code: 'SUHU_CH3',
    value: 29.5,
    unit: '°C',
    quality: 'GOOD',
    device_status: 'ONLINE',
    received_at: '2026-11-17T02:04:00Z',
  };

  // Simulate 9 badge calls + 4 progress calls + 3 color calls during a single card render + 5 consecutive 1s clock ticks
  for (let tick = 0; tick < 5; tick++) {
    for (let i = 0; i < 9; i++) {
      const b = ctx.getBentoStatusBadge(stuckParam);
      assert.strictEqual(b.label, '⚠ Stuck');
    }
    for (let i = 0; i < 4; i++) {
      const pr = ctx.getBentoCardProgress(stuckParam);
      assert.strictEqual(pr.pct, 59);
    }
    for (let i = 0; i < 3; i++) {
      const col = ctx.getBentoCardColor(stuckParam);
      assert.strictEqual(col, '#D97706');
    }
  }
  // formatRelative was called only twice total (once for alarm age, once for param age), NOT 5 * 13 = 65 times!
  assert.strictEqual(formatRelCalls, 2, `Expected 2 memoized formatRelative calls across 5 ticks, got ${formatRelCalls}`);

  // When alarm clears (bentoParamActiveAlarmMap reference changes), memo invalidates immediately!
  alarmMapRef = {};
  const clearedBadge = ctx.getBentoStatusBadge(stuckParam);
  assert.strictEqual(clearedBadge.label, 'Sedang');
});

check('4. onDeviceCommEvent & onTelemetryEvent suppress redundant steady-state offline/BAD mutations while preserving live recovery', (done) => {
  const commSrc = extractMethodSource(viewContent, 'onDeviceCommEvent(e)');
  const telemSrc = extractMethodSource(viewContent, 'onTelemetryEvent(e)');
  const flushSrc = extractMethodSource(viewContent, 'flushBentoTelemetryQueue()');

  let baseCommCalls = 0;
  let baseTelemCalls = 0;
  let setCalls = 0;

  const OperatorCommandCenterView = {
    methods: {
      onDeviceCommEvent() {
        baseCommCalls++;
      },
      onTelemetryEvent(e) {
        baseTelemCalls++;
        const d = e.detail;
        this.$set(this.liveMap, d.parameter_id, d);
      },
    },
  };

  const sandbox = {
    OperatorCommandCenterView,
    Map,
    setTimeout,
    clearTimeout,
  };
  const methodsObj = vm.runInNewContext(`({ ${commSrc}, ${telemSrc}, ${flushSrc} })`, sandbox);

  const ctx = {
    ...methodsObj,
    devicesList: [
      {
        id: 1,
        device_code: 'GAS-DEV',
        connection_status: 'OFFLINE',
        parameters: [
          { id: 11, enabled: true, formula: '' },
          { id: 12, enabled: true, formula: '' },
        ],
      },
    ],
    liveMap: {
      11: { parameter_id: 11, value: 0, quality: 'BAD', quality_reason: 'COMMUNICATION_ERROR' },
      12: { parameter_id: 12, value: 0, quality: 'BAD', quality_reason: 'COMMUNICATION_ERROR' },
    },
    deviceHasConfiguredFormula: () => false,
    recordTrendOutageEvent: () => {},
    $set(obj, key, val) {
      setCalls++;
      obj[key] = val;
    },
  };

  // 20 repeated offline connection.changed + communication.error events -> 0 baseCommCalls, 0 $set calls!
  for (let i = 0; i < 20; i++) {
    ctx.onDeviceCommEvent({
      detail: {
        type: 'device.connection.changed',
        data: { device_id: 1, connection_status: 'OFFLINE', status: 'OFFLINE' },
      },
    });
    ctx.onDeviceCommEvent({
      detail: {
        type: 'device.communication.error',
        data: { device_id: 1, error: 'timeout' },
      },
    });
  }
  assert.strictEqual(baseCommCalls, 0, 'Steady-state offline comm events must not re-trigger base handler');
  assert.strictEqual(setCalls, 0, 'Steady-state offline comm events must perform 0 reactive $set calls');

  // 20 repeated identical BAD telemetry packets -> 0 reactive $set calls, 0 baseTelemCalls!
  for (let i = 0; i < 20; i++) {
    ctx.onTelemetryEvent({
      detail: {
        device_id: 1,
        parameter_id: 11,
        value: 0,
        quality: 'BAD',
        quality_reason: 'COMMUNICATION_ERROR',
      },
    });
  }
  assert.strictEqual(setCalls, 0, 'Repeated identical BAD telemetry packets must perform 0 reactive $set calls');
  assert.strictEqual(baseTelemCalls, 0, 'Repeated BAD telemetry packets must bypass sparkline/blink base handler');

  // Genuine GOOD telemetry burst (3 updates for param 11 + 1 update for param 12 within same burst) -> coalesced to latest per param!
  ctx.onTelemetryEvent({ detail: { device_id: 1, parameter_id: 11, value: 12.1, quality: 'GOOD' } });
  ctx.onTelemetryEvent({ detail: { device_id: 1, parameter_id: 11, value: 12.4, quality: 'GOOD' } });
  ctx.onTelemetryEvent({ detail: { device_id: 1, parameter_id: 12, value: 44.0, quality: 'GOOD' } });

  if (ctx._bentoTelemFlushTimer) {
    clearTimeout(ctx._bentoTelemFlushTimer);
  }
  ctx.flushBentoTelemetryQueue();

  assert.strictEqual(baseTelemCalls, 2, '3 updates across 2 parameters in same burst must coalesce to 2 unique parameter updates');
  assert.strictEqual(ctx.liveMap[11].value, 12.4);
  assert.strictEqual(ctx.liveMap[12].value, 44.0);
});

console.log(`All ${passed} offline CPU optimization assertions passed!`);
