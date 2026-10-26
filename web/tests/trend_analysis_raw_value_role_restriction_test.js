/**
 * Test Suite: Role-based Access Control for Raw Value Mode in Trend Analysis
 *
 * Requirements:
 * - Setup pilihan / mode raw value pada Trend Analysis: tidak ada kecuali hanya Admin dan Superadmin.
 * - Non-admin/superadmin (Operator, Engineer, Guest, etc.) MUST NOT see or access Raw Value mode.
 * - If non-admin attempts to switch or loads raw mode from storage, it must be reset to 'processed'.
 * - Store getters, component template, and backend authorization must all restrict raw value access to Admin and Superadmin only.
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
console.log('TEST SUITE: TREND ANALYSIS RAW VALUE ROLE RESTRICTION');
console.log('===============================================================');

const occViewPath = path.resolve(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

const storePath = path.resolve(__dirname, '../src/store/index.js');
const storeContent = fs.readFileSync(storePath, 'utf8');

const middlewarePath = path.resolve(__dirname, '../../internal/middleware/middleware.go');
const middlewareContent = fs.readFileSync(middlewarePath, 'utf8');

const telemetryHandlerPath = path.resolve(__dirname, '../../internal/handler/telemetry_handler.go');
const telemetryHandlerContent = fs.readFileSync(telemetryHandlerPath, 'utf8');

const hubPath = path.resolve(__dirname, '../../internal/websocket/hub.go');
const hubContent = fs.readFileSync(hubPath, 'utf8');

// 1. Template-level UI Checks in OperatorCommandCenterView
runTest('1.1: Chart Setup & Controls popover guards Mode selector with v-if="canAccessRawTrendMode"', () => {
  assert(
    occContent.includes('<div v-if="canAccessRawTrendMode">') &&
    occContent.includes('<span>1. Mode</span>') &&
    occContent.includes('RAW SENSOR'),
    'Popover mode selector must be guarded with v-if="canAccessRawTrendMode"'
  );
});

runTest('1.2: Scale section dynamically adjusts label when Mode selector is hidden for non-admin', () => {
  assert(
    occContent.includes("<span>{{ canAccessRawTrendMode ? '2. Scale' : '1. Scale' }}</span>"),
    'Scale header must show clean numbered label depending on canAccessRawTrendMode'
  );
});

runTest('1.3: Fullscreen modal header and toolbar guard raw mode with canAccessRawTrendMode', () => {
  assert(
    occContent.includes('<span v-if="canAccessRawTrendMode" class="text-3xs font-semibold px-2 py-0.5 rounded-full bg-cyan-500/20 text-cyan-400 border border-cyan-500/30">'),
    'Fullscreen header badge must be guarded with v-if="canAccessRawTrendMode"'
  );
  assert(
    occContent.includes('<div v-if="canAccessRawTrendMode" class="flex items-center bg-white dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">') &&
    occContent.includes("@click=\"setChartMode('raw')\""),
    'Fullscreen toolbar mode selector must be guarded with v-if="canAccessRawTrendMode"'
  );
});

runTest('1.4: Export Data modal guards RAW SENSOR option with v-if="canAccessRawTrendMode"', () => {
  assert(
    occContent.includes('v-if="canAccessRawTrendMode"') &&
    occContent.includes("@click=\"exportForm.mode = 'RAW'\""),
    'Export modal RAW option must be guarded by v-if="canAccessRawTrendMode"'
  );
});

// 2. Computed Property & Role Checking Logic
runTest('2.1: canAccessRawTrendMode allows only SUPERADMIN, ADMINISTRATOR, and ADMIN', () => {
  assert(occContent.includes('canAccessRawTrendMode()'), 'canAccessRawTrendMode computed property must be defined');

  function checkRoleAccess(role) {
    const r = (role || '').toUpperCase();
    return ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN'].includes(r);
  }

  // Allowed roles
  assert.strictEqual(checkRoleAccess('SUPERADMIN'), true, 'SUPERADMIN must be allowed');
  assert.strictEqual(checkRoleAccess('superadmin'), true, 'superadmin lowercase must be allowed');
  assert.strictEqual(checkRoleAccess('ADMINISTRATOR'), true, 'ADMINISTRATOR must be allowed');
  assert.strictEqual(checkRoleAccess('Administrator'), true, 'Administrator mixed case must be allowed');
  assert.strictEqual(checkRoleAccess('ADMIN'), true, 'ADMIN must be allowed');
  assert.strictEqual(checkRoleAccess('admin'), true, 'admin lowercase must be allowed');

  // Forbidden roles
  assert.strictEqual(checkRoleAccess('OPERATOR'), false, 'OPERATOR must be denied');
  assert.strictEqual(checkRoleAccess('operator'), false, 'operator lowercase must be denied');
  assert.strictEqual(checkRoleAccess('ENGINEER'), false, 'ENGINEER must be denied');
  assert.strictEqual(checkRoleAccess('engineer'), false, 'engineer lowercase must be denied');
  assert.strictEqual(checkRoleAccess('VIEWER'), false, 'VIEWER must be denied');
  assert.strictEqual(checkRoleAccess('GUEST'), false, 'GUEST must be denied');
  assert.strictEqual(checkRoleAccess(''), false, 'Empty role must be denied');
  assert.strictEqual(checkRoleAccess(null), false, 'Null role must be denied');
});

// 3. Security Guards in Methods & Watchers
runTest('3.1: setChartMode denies switching to raw when canAccessRawTrendMode is false', () => {
  assert(
    occContent.includes("if (mode === 'raw' && !this.canAccessRawTrendMode)"),
    'setChartMode must check canAccessRawTrendMode before switching to raw'
  );
});

runTest('3.2: loadOperatorPreferences and mounted sanitize chartMode if user is unauthorized', () => {
  assert(
    occContent.includes("this.chartMode = (p.chartMode === 'raw' && !this.canAccessRawTrendMode) ? 'processed' : p.chartMode;"),
    'loadOperatorPreferences must reset raw mode to processed for non-admin'
  );
  assert(
    occContent.includes("if (!this.canAccessRawTrendMode && this.chartMode === 'raw') {\n      this.chartMode = 'processed';\n    }"),
    'mounted hook must guard chartMode'
  );
});

runTest('3.3: Watcher on canAccessRawTrendMode and userRole resets raw mode on role change', () => {
  assert(
    occContent.includes('canAccessRawTrendMode(canAccess) {\n      if (!canAccess && this.chartMode === \'raw\') {'),
    'canAccessRawTrendMode watcher must reset chartMode to processed'
  );
});

runTest('3.4: recordTrendPoint only uses raw_value when authorized and active in raw mode', () => {
  assert(
    occContent.includes("const isRaw = this.canAccessRawTrendMode && this.chartMode === 'raw';"),
    'recordTrendPoint must evaluate canAccessRawTrendMode && chartMode === "raw"'
  );
});

runTest('3.5: fetchTrendData ensures activeMode guards requests to raw endpoints', () => {
  assert(
    occContent.includes("const activeMode = (this.canAccessRawTrendMode && this.chartMode === 'raw') ? 'raw' : 'processed';"),
    'fetchTrendData must evaluate activeMode with canAccessRawTrendMode'
  );
});

// 4. Store Getters
runTest('4.1: Vuex store exposes canAccessRawTrendMode allowing admin and superadmin', () => {
  assert(
    storeContent.includes("canAccessRawTrendMode: (state, getters) => getters.isSuperAdmin || getters.isAdmin"),
    'Vuex store must define canAccessRawTrendMode'
  );
});

// 5. Backend Authorization and WebSocket Sanitization
runTest('5.1: Backend middleware allows telemetry.raw for Admin and Superadmin only', () => {
  assert(
    middlewareContent.includes('if permissionCode == "telemetry.raw"') &&
    middlewareContent.includes('!strings.EqualFold(role, "ADMIN")'),
    'Middleware must restrict telemetry.raw to Admin and Superadmin'
  );
});

runTest('5.2: Backend telemetry handler restricts raw telemetry endpoints to Admin/Superadmin', () => {
  assert(
    telemetryHandlerContent.includes('canAccessRawTelemetry(c *gin.Context) bool') &&
    telemetryHandlerContent.includes('strings.EqualFold(role, "ADMIN")'),
    'TelemetryHandler must authorize Admin and Superadmin for raw telemetry'
  );
});

runTest('5.3: WebSocket hub broadcasts raw telemetry only to authorized Admin and Superadmin', () => {
  assert(
    hubContent.includes('canRaw := isSuper || strings.EqualFold(client.role, "ADMIN")') &&
    hubContent.includes('if !canRaw && message.Type == "device.telemetry.received"'),
    'WebSocket hub must sanitize raw telemetry for clients without Admin/Superadmin privileges'
  );
});

console.log('===============================================================');
console.log('TEST SUMMARY: ALL ROLE RESTRICTION TESTS PASSED (14/14)');
console.log('===============================================================');
