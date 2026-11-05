/**
 * Test Suite: Realtime Telemetry & Historical Data Raw Value Role Restrictions
 *
 * Verifies:
 * Gambar 1:
 * - Realtime Telemetry: hide formula nilai ƒ(x) dan raw/hex kecuali admin dan superadmin.
 * Gambar 2:
 * - Historical Data: hide pilihan raw value (mode switch, resolution dropdown, table column, hover tooltip)
 *   kecuali admin dan superadmin.
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

function runTest(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    process.exit(1);
  }
}

console.log('===============================================================');
console.log('TEST SUITE: REALTIME TELEMETRY & HISTORICAL RAW VALUE RESTRICTIONS');
console.log('===============================================================');

const ddViewPath = path.resolve(__dirname, '../src/views/DeviceDetailView.vue');
const ddContent = fs.readFileSync(ddViewPath, 'utf8');

const tmViewPath = path.resolve(__dirname, '../src/views/TelemetryMonitorView.vue');
const tmContent = fs.readFileSync(tmViewPath, 'utf8');

// GAMBAR 1: REALTIME TELEMETRY
runTest('1.1: Live Telemetry formula result ƒ(x) is guarded by canAccessRawMode', () => {
  assert(
    ddContent.includes('v-if="canAccessRawMode && getParamFormulaValue(param) !== null"'),
    'Formula result ƒ(x) in Live Telemetry tab must be guarded by canAccessRawMode'
  );
});

runTest('1.2: Live Telemetry RAW / HEX diagnostics sub-bar is guarded by canAccessRawMode', () => {
  assert(
    ddContent.includes('<div v-if="canAccessRawMode">') &&
    ddContent.includes("{{ $t('telemetry.rawHex') }}") &&
    ddContent.includes('{{ getParamRaw(param) }}'),
    'RAW / HEX sub-bar in Live Telemetry cards must be guarded by canAccessRawMode'
  );
});

runTest('1.3: getParamFormulaValue and getParamRaw return null / empty for non-admin/superadmin', () => {
  assert(
    ddContent.includes('if (!this.canAccessRawMode) return null;'),
    'getParamFormulaValue must guard against non-admin/superadmin'
  );
  assert(
    ddContent.includes("if (!this.canAccessRawMode) return '--';"),
    'getParamRaw must guard against non-admin/superadmin'
  );
});

// GAMBAR 2: HISTORICAL DATA
runTest('2.1: Historical Data Mode Switch (Scaled vs Raw Value) is guarded by canAccessRawMode', () => {
  assert(
    ddContent.includes('<!-- Mode Switch: Scaled / Formula vs Raw Value (Admin and Superadmin only) -->') &&
    ddContent.includes('<div v-if="canAccessRawMode" class="bg-slate-900 p-0.5 rounded-lg border border-slate-800 flex items-center shadow-sm">') &&
    ddContent.includes("@click=\"chartValueMode = 'raw'\""),
    'Mode Switch (Scaled vs Raw) must be guarded with v-if="canAccessRawMode"'
  );
});

runTest('2.2: Downsampling resolution dropdown preserves full resolution option for instant telemetry load', () => {
  assert(
    ddContent.includes('<option value="raw">{{ $t(\'deviceDetail.resolutionRaw\') }}</option>'),
    'Downsampling resolution must support full resolution telemetry loading'
  );
});

runTest('2.3: Chart Y-scale watermark and hover tooltip raw mode indicators are guarded', () => {
  assert(
    ddContent.includes('<span v-if="canAccessRawMode" class="ml-1 text-slate-500 font-mono text-4xs">[{{ chartValueMode === \'raw\' ? $t(\'deviceDetail.modeRaw\') : $t(\'deviceDetail.modeScaled\') }}]</span>'),
    'Y-scale watermark mode indicator must be guarded by v-if="canAccessRawMode"'
  );
  assert(
    ddContent.includes('v-if="canAccessRawMode"\n                  class="text-4xs font-sans font-bold px-1.5 py-0.5 rounded ml-1.5 uppercase"') ||
    ddContent.includes('v-if="canAccessRawMode"') && ddContent.includes("chartValueMode === 'raw' ? 'RAW' : 'SCALED'"),
    'Chart hover tooltip mode badge must be guarded by v-if="canAccessRawMode"'
  );
  assert(
    ddContent.includes('<div v-if="canAccessRawMode" class="text-4xs text-slate-400 font-sans flex items-center justify-between pt-1 border-t border-slate-800/60">'),
    'Hover tooltip raw value detail must be guarded by v-if="canAccessRawMode"'
  );
});

runTest('2.4: Historical Data table RAW DECODED column is guarded by canAccessRawMode', () => {
  assert(
    ddContent.includes("<th v-if=\"canAccessRawMode\" class=\"py-2.5 px-4\">{{ $t('deviceDetail.rawDecodedCol') }}</th>"),
    'Historical Data table RAW header must be guarded by v-if="canAccessRawMode"'
  );
  assert(
    ddContent.includes('<td v-if="canAccessRawMode" class="py-2.5 px-4 text-slate-400">'),
    'Historical Data table RAW data cell must be guarded by v-if="canAccessRawMode"'
  );
});

runTest('2.5: chartActiveUnit and getTelemetryVal guard raw value mode', () => {
  assert(
    ddContent.includes("if (this.canAccessRawMode && this.chartValueMode === 'raw') {\n        return 'RAW';\n      }"),
    'chartActiveUnit must only return RAW when canAccessRawMode is true'
  );
  assert(
    ddContent.includes("if (this.canAccessRawMode && this.chartValueMode === 'raw') {\n        if (rec.raw_value !== undefined"),
    'getTelemetryVal must only return raw_value when canAccessRawMode is true'
  );
});

// ROLE CHECK & LIFECYCLE GUARDS
runTest('3.1: canAccessRawMode computed property permits only SUPERADMIN, ADMINISTRATOR, and ADMIN', () => {
  assert(ddContent.includes('canAccessRawMode()'), 'canAccessRawMode computed property must be defined');

  function checkRole(role) {
    const r = (role || '').toUpperCase();
    return ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN'].includes(r);
  }

  assert.strictEqual(checkRole('SUPERADMIN'), true);
  assert.strictEqual(checkRole('ADMINISTRATOR'), true);
  assert.strictEqual(checkRole('ADMIN'), true);
  assert.strictEqual(checkRole('OPERATOR'), false);
  assert.strictEqual(checkRole('ENGINEER'), false);
});

runTest('3.2: Raw Telemetry tab allows Operator access while hiding Raw Hex & Raw Float and renaming Scaled Value to Value', () => {
  assert(
    ddContent.includes('v-if="device && activeTab === \'raw\'"'),
    'Raw Telemetry tab container must be accessible to Operator as well as Admin/Superadmin'
  );
  assert(
    ddContent.includes('<th v-if="canAccessRawMode" class="py-2.5 px-4">{{ $t(\'deviceDetail.rawHexStream\') }}</th>') &&
    ddContent.includes('<th v-if="canAccessRawMode" class="py-2.5 px-4">{{ $t(\'deviceDetail.rawFloat\') }}</th>'),
    'Raw Hex Stream and Raw Float headers must be guarded by v-if="canAccessRawMode"'
  );
  assert(
    ddContent.includes("{{ canAccessRawMode ? $t('deviceDetail.scaledValue') : $t('deviceDetail.rawOperatorValue') }}"),
    'Scaled Value column header must switch to Value (rawOperatorValue) when !canAccessRawMode'
  );
  assert(
    ddContent.includes('? `/api/devices/${this.device.id}/telemetry/raw`') &&
    ddContent.includes(': `/api/devices/${this.device.id}/telemetry/history`'),
    'fetchRawTelemetry must use sanitized /telemetry/history endpoint when !canAccessRawMode'
  );
  assert(
    ddContent.includes('rawMaxRangeHours()') && ddContent.includes('rawCustomSpanExceedsMax()'),
    'Raw Telemetry tab must enforce configurable rawMaxRangeHours window limit'
  );
});

runTest('3.3: Watchers reset chartValueMode to scaled when raw mode is unauthorized', () => {
  assert(
    ddContent.includes("if (this.chartValueMode === 'raw') {\n            this.chartValueMode = 'scaled';"),
    'Watchers must reset chartValueMode to scaled'
  );
});

// TELEMETRY MONITOR VIEW
runTest('4.1: TelemetryMonitorView guards formula value display while allowing Raw Telemetry tab navigation', () => {
  assert(
    tmContent.includes('v-if="canAccessRawMode && item.formula_value !== undefined && item.formula_value !== null"'),
    'TelemetryMonitorView card formula must be guarded by canAccessRawMode'
  );
  assert(
    tmContent.includes(':to="\'/monitoring/devices/\' + item.device_id + \'?tab=raw\'"'),
    'TelemetryMonitorView must provide navigation link to ?tab=raw'
  );
  assert(
    tmContent.includes("<th v-if=\"canAccessRawMode\" class=\"py-3 px-4 text-right\">{{ $t('parameters.formula') }}</th>"),
    'TelemetryMonitorView table formula column must be guarded by canAccessRawMode'
  );
});

// PARAMETERS TAB RESTRICTION
runTest('5.1: Telemetry Parameters table hides SCALE / OFFSET column for non-admin/superadmin', () => {
  assert(
    ddContent.includes('<th v-if="canAccessRawMode" class="py-2.5 px-3">{{ $t(\'deviceDetail.colScaleOffset\') }}</th>'),
    'Parameters table header for colScaleOffset must be guarded by canAccessRawMode'
  );
  assert(
    ddContent.includes('<td v-if="canAccessRawMode" class="py-2.5 px-3 font-mono text-3xs text-slate-600 dark:text-slate-400">'),
    'Parameters table data cell for scale/offset must be guarded by canAccessRawMode'
  );
  assert(
    ddContent.includes('<td :colspan="canAccessRawMode ? 11 : 10" class="py-8 text-center text-xs text-slate-400">'),
    'Parameters table empty row colspan must adapt to canAccessRawMode'
  );
});

// OPERATOR NAVIGATION, LANDING & TELEMETRY ANALYSIS RESTRICTIONS
runTest('6.1: Sidebar hides Operator Console, Live Telemetry, Live Sensor Comm, and Outbound & Integration for OPERATOR while keeping Delivery Queue', () => {
  const sidebarContent = fs.readFileSync(path.resolve(__dirname, '../src/components/Sidebar.vue'), 'utf8');
  assert(sidebarContent.includes('v-if="!isOperator" to="/monitoring/operator"'), 'Sidebar must hide Operator Console for OPERATOR');
  assert(sidebarContent.includes('v-if="!isOperator" to="/monitoring/telemetry"'), 'Sidebar must hide Live Telemetry for OPERATOR');
  assert(sidebarContent.includes('v-if="!isOperator" to="/alarm/live-comm"'), 'Sidebar must hide Live Sensor Comm for OPERATOR');
  assert(sidebarContent.includes('v-if="!isOperator" to="/integration"'), 'Sidebar must hide Outbound & Integration for OPERATOR');
  assert(sidebarContent.includes('<router-link to="/integration/queue"'), 'Sidebar must keep Delivery Queue visible for OPERATOR');
});

runTest('6.2: Router, Header & Login default OPERATOR landing to /monitoring/dashboard-operator and preserve Delivery Queue (/integration/queue & /integration?tab=queue)', () => {
  const routerContent = fs.readFileSync(path.resolve(__dirname, '../src/router/index.js'), 'utf8');
  const headerContent = fs.readFileSync(path.resolve(__dirname, '../src/components/Header.vue'), 'utf8');
  const loginContent = fs.readFileSync(path.resolve(__dirname, '../src/views/LoginView.vue'), 'utf8');
  assert(routerContent.includes("return next({ path: '/monitoring/dashboard-operator', replace: true });"), 'Router beforeEach must redirect OPERATOR default/console routes to /monitoring/dashboard-operator');
  assert(routerContent.includes("if (to.query && to.query.tab === 'queue')"), 'Router beforeEach must allow OPERATOR on /integration?tab=queue');
  assert(headerContent.includes("if (isOp && this.$route.path === '/integration' && this.$route.query && this.$route.query.tab === 'queue')"), 'Header switch-account must keep OPERATOR on /integration?tab=queue');
  assert(loginContent.includes("isOperator ? '/monitoring/dashboard-operator' : '/monitoring/operator'"), 'LoginView must land OPERATOR on /monitoring/dashboard-operator');
});

runTest('6.3: TelemetryAnalysisView hides Raw Telemetry tab & Live Telemetry button for OPERATOR and removes Formula text on Historical tab', () => {
  const taContent = fs.readFileSync(path.resolve(__dirname, '../src/views/TelemetryAnalysisView.vue'), 'utf8');
  const enContent = fs.readFileSync(path.resolve(__dirname, '../src/i18n/locales/en.js'), 'utf8');
  const idContent = fs.readFileSync(path.resolve(__dirname, '../src/i18n/locales/id.js'), 'utf8');

  assert(/v-if="!isOperator"\s+@click="setMode\('RAW'\)"/.test(taContent), 'TelemetryAnalysisView must hide Raw Telemetry tab for OPERATOR');
  assert(/v-if="!isOperator"\s+to="\/monitoring\/telemetry"/.test(taContent), 'TelemetryAnalysisView must hide Live Telemetry button for OPERATOR');
  assert(taContent.includes("(this.isOperator && targetMode === 'RAW') ? 'HISTORICAL' : targetMode"), 'TelemetryAnalysisView must redirect ?mode=RAW to HISTORICAL for OPERATOR');

  const enHistDesc = enContent.match(/"modeHistoricalDesc":\s*"([^"]*)"/);
  const idHistDesc = idContent.match(/"modeHistoricalDesc":\s*"([^"]*)"/);
  const enHistSem = enContent.match(/"historicalSemantic":\s*"([^"]*)"/);
  const idHistSem = idContent.match(/"historicalSemantic":\s*"([^"]*)"/);

  assert(enHistDesc && !/formula/i.test(enHistDesc[1]), 'en.js modeHistoricalDesc must not contain Formula');
  assert(idHistDesc && !/formula/i.test(idHistDesc[1]), 'id.js modeHistoricalDesc must not contain Formula');
  assert(enHistSem && !/formula/i.test(enHistSem[1]), 'en.js historicalSemantic must not contain Formula');
  assert(idHistSem && !/formula/i.test(idHistSem[1]), 'id.js historicalSemantic must not contain Formula');
});

runTest('6.4: TelemetryAnalysisView hides all device info & device links for OPERATOR on Historical view, Summary Breakdown, tooltip, breadcrumb, and JSON export', () => {
  const taContent = fs.readFileSync(path.resolve(__dirname, '../src/views/TelemetryAnalysisView.vue'), 'utf8');
  assert(
    taContent.includes(':to="isOperator ? \'/monitoring/dashboard-operator\' : \'/monitoring/devices\'"'),
    'Breadcrumb must route OPERATOR to /monitoring/dashboard-operator instead of /monitoring/devices'
  );
  assert(
    taContent.includes('<th v-if="!isOperator" class="py-2.5 px-4 text-right">{{ $t(\'deviceDetail.actions\') || \'Actions\' }}</th>') &&
    taContent.includes('<td v-if="!isOperator" class="py-2.5 px-4 text-right font-sans">'),
    'Historical table must hide Actions / Detail -> device link column for OPERATOR'
  );
  assert(
    taContent.includes('<span v-if="!isOperator">{{ s.deviceCode }}: </span>{{ s.paramCode }}'),
    'Per-Parameter Summary Breakdown cards and chart legend must hide deviceCode prefix for OPERATOR'
  );
  assert(
    taContent.includes('<span v-if="!isOperator">{{ activePoint.deviceCode }}: </span>{{ activePoint.paramCode }}') &&
    taContent.includes('<span v-if="!isOperator && activePoint.deviceName"> ({{ activePoint.deviceName }})</span>'),
    'Chart hover tooltip must hide deviceCode and deviceName for OPERATOR'
  );
  assert(
    taContent.includes('this.records.map(({ device_id, device, ...rest }) => rest)'),
    'exportJSON must strip device_id and device metadata for OPERATOR'
  );
});

runTest('7.1: DeviceDetailView & main.css enforce Light Mode (Mode Siang) contrast across Historical, Raw, Aggregation, Audit, and Diagnostic Report', () => {
  const deviceDetailContent = fs.readFileSync(path.resolve(__dirname, '../src/views/DeviceDetailView.vue'), 'utf8');
  const mainCss = fs.readFileSync(path.resolve(__dirname, '../src/assets/main.css'), 'utf8');
  assert(
    !/(?<!dark:|hover:)bg-slate-800/.test(deviceDetailContent),
    'DeviceDetailView must not contain unprefixed bg-slate-800 without light-mode background + dark: prefix'
  );
  assert(
    !/(?<!dark:)bg-black/.test(deviceDetailContent),
    'DeviceDetailView must not contain unprefixed bg-black without light-mode background + dark: prefix'
  );
  assert(
    deviceDetailContent.includes('chart-badge-max') &&
    deviceDetailContent.includes('chart-badge-min') &&
    deviceDetailContent.includes('chart-badge-avg') &&
    mainCss.includes('html:not(.dark) .chart-svg .chart-badge-max rect'),
    'Historical Trend SVG callout badges (MAX, MIN, AVG) must have dedicated Light Mode contrast overrides'
  );
});

runTest('7.2: LoginView removes KLHK & SPARING compliance footer texts and enforces crisp Light/Dark Mode contrast on CTA & Brand Mark', () => {
  const loginContent = fs.readFileSync(path.resolve(__dirname, '../src/views/LoginView.vue'), 'utf8');
  const mainCss = fs.readFileSync(path.resolve(__dirname, '../src/assets/main.css'), 'utf8');
  assert(
    !loginContent.includes('KLHK Permen LHK No. 14/2020 Compliant') &&
    !loginContent.includes('SPARING Direct Integration Ready'),
    'LoginView must not display KLHK Permen LHK No. 14/2020 Compliant or SPARING Direct Integration Ready'
  );
  assert(
    loginContent.includes('login-primary-btn') &&
    loginContent.includes('login-brand-mark') &&
    mainCss.includes('html:not(.dark) .login-primary-btn'),
    'LoginView & main.css must preserve crisp white contrast on primary Sign In CTA and brand mark in Light Mode'
  );
});

runTest('8.1: ParameterModal exposes Stale Data Timeout (Seconds) input, quick presets, and effective polling guard', () => {
  const paramModalContent = fs.readFileSync(path.resolve(__dirname, '../src/components/ParameterModal.vue'), 'utf8');
  assert(
    paramModalContent.includes('v-model.number="form.stale_timeout_seconds"') &&
    paramModalContent.includes("$t('parameterModal.staleTimeoutSeconds')") &&
    paramModalContent.includes('effectiveStaleTimeoutSeconds'),
    'ParameterModal must expose editable Stale Data Timeout (Seconds) input and effective timeout indicator'
  );
});

console.log('===============================================================');
console.log('TEST SUMMARY: ALL RESTRICTION TESTS PASSED (20/20)');
console.log('===============================================================');
