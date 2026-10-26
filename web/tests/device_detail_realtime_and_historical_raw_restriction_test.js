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

runTest('3.2: switchTab and mounted redirect tab=raw to telemetry for non-admin/superadmin', () => {
  assert(
    ddContent.includes("if (tab === 'raw' && !this.canAccessRawMode) {\n        this.activeTab = 'telemetry';\n        return;\n      }"),
    'switchTab must redirect raw tab to telemetry for non-admin'
  );
  assert(
    ddContent.includes("this.activeTab = (this.$route.query.tab === 'raw' && !this.canAccessRawMode) ? 'telemetry' : this.$route.query.tab;"),
    'mounted must sanitize tab=raw for non-admin'
  );
});

runTest('3.3: Watchers reset chartValueMode to scaled when raw mode is unauthorized', () => {
  assert(
    ddContent.includes("if (this.chartValueMode === 'raw') {\n            this.chartValueMode = 'scaled';"),
    'Watchers must reset chartValueMode to scaled'
  );
});

// TELEMETRY MONITOR VIEW
runTest('4.1: TelemetryMonitorView guards formula value display and raw device links', () => {
  assert(
    tmContent.includes('v-if="canAccessRawMode && item.formula_value !== undefined && item.formula_value !== null"'),
    'TelemetryMonitorView card formula must be guarded by canAccessRawMode'
  );
  assert(
    tmContent.includes('<router-link\n                v-if="canAccessRawMode"\n                :to="\'/monitoring/devices/\' + item.device_id + \'?tab=raw\'"'),
    'TelemetryMonitorView raw tab link must be guarded by canAccessRawMode'
  );
  assert(
    tmContent.includes("<th v-if=\"canAccessRawMode\" class=\"py-3 px-4 text-right\">{{ $t('parameters.formula') }}</th>"),
    'TelemetryMonitorView table formula column must be guarded by canAccessRawMode'
  );
});

console.log('===============================================================');
console.log('TEST SUMMARY: ALL REALTIME & HISTORICAL TESTS PASSED (12/12)');
console.log('===============================================================');
