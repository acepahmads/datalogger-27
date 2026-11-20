const fs = require('fs');
const path = require('path');
const vm = require('vm');
const assert = require('assert');

const webSrc = path.join(__dirname, '..', 'src');
const viewCode = fs.readFileSync(path.join(webSrc, 'views', 'InboundDataView.vue'), 'utf8');
const routerCode = fs.readFileSync(path.join(webSrc, 'router', 'index.js'), 'utf8');
const sidebarCode = fs.readFileSync(path.join(webSrc, 'components', 'Sidebar.vue'), 'utf8');
const validatorRaw = fs.readFileSync(path.join(webSrc, 'utils', 'inboundValidator.js'), 'utf8');

// 1. Router registers /monitoring/inbound-data strictly for SUPERADMIN + Cloud Mode
assert(routerCode.includes("path: '/monitoring/inbound-data'"), 'Router must register /monitoring/inbound-data');
assert(routerCode.includes("meta: { roles: ['SUPERADMIN'], cloudOnly: true }"), 'InboundData route must enforce SUPERADMIN + cloudOnly');
assert(routerCode.includes('if (to.meta && to.meta.cloudOnly && !isCloudMode)'), 'Router guard must enforce cloudOnly restriction');

// 2. Sidebar shows Inbound Data link strictly for SUPERADMIN in Cloud Mode
assert(sidebarCode.includes('id="nav-inbound-data"'), 'Sidebar must expose #nav-inbound-data link');
assert(/v-if="isSuperAdmin && isCloudMode"\s+id="nav-inbound-data"/.test(sidebarCode), 'Sidebar Inbound Data link must be gated by isSuperAdmin && isCloudMode');

// 3. InboundDataView enforces role/mode guard + compact filters + 3-mode Raw JSON inspector + Validity Audit Tab
assert(viewCode.includes('if (!this.isSuperAdmin || !this.isCloudMode)'), 'InboundDataView must redirect non-Superadmin or Edge Mode users');
assert(viewCode.includes('/api/superadmin/inbound-logs'), 'InboundDataView must fetch from /api/superadmin/inbound-logs');
assert(!viewCode.includes('id="filter-inbound-tenant"'), 'InboundDataView must NOT render duplicate Tenant dropdown (uses top Header tenant selector)');
assert(!viewCode.includes('id="filter-inbound-logger"'), 'InboundDataView must NOT render Domain (AQMS/SPARING) dropdown');
assert(viewCode.includes('this.activeTenantId > 0'), 'InboundDataView must filter automatically by Header activeTenantId');
assert(viewCode.includes('id="filter-inbound-status"'), 'InboundDataView must include Status filter');
assert(viewCode.includes('id="filter-inbound-format"'), 'InboundDataView must include Format filter');
assert(viewCode.includes('id="filter-inbound-source"'), 'InboundDataView must include Source filter');
assert(viewCode.includes('id="filter-inbound-search"'), 'InboundDataView must include Search input');
assert(viewCode.includes('id="btn-inbound-view-pretty"') && viewCode.includes('id="btn-inbound-view-human"') && viewCode.includes('id="btn-inbound-view-raw"'), 'InboundDataView must provide Pretty / Waktu Terbaca / Raw JSON viewer switch');

// 3b. New Kevalidan Data Tab UI structure
assert(viewCode.includes('id="tab-inbound-stream"') && viewCode.includes('id="tab-inbound-validity"'), 'InboundDataView must expose Stream Raw JSON & Kevalidan Data tabs');
assert(viewCode.includes('id="panel-inbound-validity"'), 'InboundDataView must render Kevalidan Data panel');
assert(viewCode.includes('id="validity-kpi-strip"') && viewCode.includes('id="btn-validity-sparing-only"'), 'Kevalidan Data tab must render clickable KPI filter strip + SPARING (00-58) filter');
assert(viewCode.includes('id="sparing-minute-matrix"'), 'Kevalidan Data tab must render visual SPARING 30-slot minute matrix (00-58)');

// 4. Strict Light-Mode CSS Hygiene
const classRegex = /class="([^"]*)"/g;
let match;
while ((match = classRegex.exec(viewCode)) !== null) {
  const cls = match[1];
  const tokens = cls.split(/\s+/);
  if (tokens.includes('text-white')) {
    assert(
      tokens.includes('btn-text') || tokens.includes('badge-text'),
      `InboundDataView Light-Mode violation: class="${cls}" uses bare text-white without btn-text or badge-text`
    );
  }
  assert(
    !tokens.includes('bg-slate-900'),
    `InboundDataView Light-Mode violation: class="${cls}" uses bare bg-slate-900 without dark: prefix`
  );
}

// 5. Runtime Verification of inboundValidator.js (Format, Values, Completeness + SPARING 00-58)
const cjsValidator =
  validatorRaw.replace(/export\s+(const|function)\s+/g, '$1 ') +
  '\nmodule.exports = { EXPECTED_SPARING_MINUTES, parseInboundTimestamp, validateInboundLog };';
const valMod = { exports: {} };
vm.runInNewContext(cjsValidator, { module: valMod, exports: valMod.exports, Date, Math, Number, String, Boolean, Array, Set, Map, Object, RegExp, JSON });
const { EXPECTED_SPARING_MINUTES, validateInboundLog } = valMod.exports;

assert.strictEqual(EXPECTED_SPARING_MINUTES.length, 30, 'SPARING must expect exactly 30 even-minute slots');
assert.strictEqual(EXPECTED_SPARING_MINUTES[0], 0);
assert.strictEqual(EXPECTED_SPARING_MINUTES[29], 58);

// Base timestamp: 2026-11-20 09:00:00 +0700 (exact hour boundary)
const baseHourMs = Date.parse('2026-11-20T09:00:00+07:00');
const makeSparingSlot = (min, overrides = {}) => ({
  datetime: Math.floor((baseHourMs + min * 60000) / 1000),
  pH: 7.12,
  cod: 42.5,
  tss: 18.4,
  nh3n: 1.25,
  debit: 12.6,
  ...overrides,
});

// Case A: Complete SPARING 30x2m (00..58)
const completeSparingPayload = JSON.stringify({
  uid: '63fd626ce7defd222874166a',
  data: EXPECTED_SPARING_MINUTES.map(m => makeSparingSlot(m)),
});
const resComplete = validateInboundLog({
  id: 1,
  status: 'ACCEPTED',
  http_status: 200,
  logger_type: 'SPARING',
  detected_format: 'KLHK_SPARING',
  parameters_ingested: 150,
  parameters_skipped: 0,
  raw_payload: completeSparingPayload,
});
assert.strictEqual(resComplete.isSparing, true);
assert.strictEqual(resComplete.overallStatus, 'VALID', 'Complete 30/30 SPARING 00-58 must be VALID');
assert.strictEqual(resComplete.score, 100, 'Complete 30/30 SPARING 00-58 must score 100');
assert.strictEqual(resComplete.completeness.sparing.completeSlots, 30);
assert.strictEqual(resComplete.completeness.sparing.missingMinutes.length, 0);

// Case B: Missing SPARING minutes (:14 and :58 absent)
const missingSparingPayload = JSON.stringify({
  uid: '63fd626ce7defd222874166a',
  data: EXPECTED_SPARING_MINUTES.filter(m => m !== 14 && m !== 58).map(m => makeSparingSlot(m)),
});
const resMissing = validateInboundLog({
  id: 2,
  status: 'ACCEPTED',
  http_status: 200,
  logger_type: 'SPARING',
  detected_format: 'KLHK_SPARING',
  raw_payload: missingSparingPayload,
});
assert.strictEqual(resMissing.overallStatus, 'INVALID', 'SPARING missing :14 and :58 must be INVALID');
assert.strictEqual(resMissing.completeness.status, 'FAIL');
assert.deepStrictEqual(Array.from(resMissing.completeness.sparing.missingMinutes), [14, 58]);
assert.strictEqual(resMissing.completeness.sparing.presentSlots, 28);

// Case C: SPARING 30 slots present, but :56 and :58 have null values
const nullSparingPayload = JSON.stringify({
  uid: '63fd626ce7defd222874166a',
  data: EXPECTED_SPARING_MINUTES.map(m => (m >= 56 ? makeSparingSlot(m, { pH: null, cod: null }) : makeSparingSlot(m))),
});
const resPartialNull = validateInboundLog({
  id: 3,
  status: 'ACCEPTED',
  http_status: 200,
  logger_type: 'SPARING',
  detected_format: 'KLHK_SPARING',
  raw_payload: nullSparingPayload,
});
assert.strictEqual(resPartialNull.overallStatus, 'WARNING', 'SPARING with 30/30 slots present but null values must be WARNING');
assert.strictEqual(resPartialNull.completeness.sparing.presentSlots, 30);
assert.strictEqual(resPartialNull.completeness.sparing.completeSlots, 28);
assert.strictEqual(resPartialNull.completeness.sparing.partialSlots, 2);
assert.strictEqual(resPartialNull.values.nullCount, 4);

// Case D: SPARING physical range anomaly (pH = 18.5, cod = -5) + odd minute (:01)
const anomalySparingPayload = JSON.stringify({
  uid: '63fd626ce7defd222874166a',
  data: [
    makeSparingSlot(1, { pH: 18.5, cod: -5 }),
    ...EXPECTED_SPARING_MINUTES.slice(1).map(m => makeSparingSlot(m)),
  ],
});
const resAnomaly = validateInboundLog({
  id: 4,
  status: 'ACCEPTED',
  http_status: 200,
  logger_type: 'SPARING',
  detected_format: 'KLHK_SPARING',
  raw_payload: anomalySparingPayload,
});
assert.strictEqual(resAnomaly.overallStatus, 'INVALID');
assert.strictEqual(resAnomaly.values.status, 'FAIL');
assert.strictEqual(resAnomaly.values.outOfRangeCount, 2);
assert.deepStrictEqual(Array.from(resAnomaly.completeness.sparing.offGridMinutes), [1]);
assert.deepStrictEqual(Array.from(resAnomaly.completeness.sparing.missingMinutes), [0]);

// Case E: CBI Legacy multi-tick AQMS (exact screenshot shape with 1 skipped param -> WARNING)
const cbiMultiPayload = JSON.stringify({
  multi: [
    { tick: { module_id: '194', station_id: '197', sensor_name: 'o3', value: 66.36, created_at: '2026-11-20 10:00:00 +0700', updated_at: '2026-11-20 10:00:00 +0700' } },
    { tick: { module_id: '194', station_id: '197', sensor_name: 'pressure', value: 966.73, created_at: '2026-11-20 10:00:00 +0700', updated_at: '2026-11-20 10:00:00 +0700' } },
  ],
});
const resCbi = validateInboundLog({
  id: 5,
  status: 'ACCEPTED',
  http_status: 200,
  logger_type: 'AQMS',
  detected_format: 'CBI_LEGACY',
  parameters_ingested: 14,
  parameters_skipped: 1,
  raw_payload: cbiMultiPayload,
});
assert.strictEqual(resCbi.isSparing, false);
assert.strictEqual(resCbi.schemaLabel, 'CBI_LEGACY (multi[])');
assert.strictEqual(resCbi.format.status, 'PASS');
assert.strictEqual(resCbi.values.status, 'PASS');
assert.strictEqual(resCbi.overallStatus, 'WARNING', 'Skipped unmapped parameter should raise WARNING');

// Case F: Corrupted JSON
const resBroken = validateInboundLog({
  id: 6,
  status: 'REJECTED',
  http_status: 400,
  error_message: 'invalid character',
  raw_payload: '{"multi": [broken',
});
assert.strictEqual(resBroken.overallStatus, 'INVALID');
assert.strictEqual(resBroken.format.status, 'FAIL');
assert.strictEqual(resBroken.score, 0);

console.log('PASS: Cloud Mode Superadmin Inbound Data Monitor + Kevalidan Data Audit (Format, Nilai, Kelengkapan & SPARING 00-58) verified!');
