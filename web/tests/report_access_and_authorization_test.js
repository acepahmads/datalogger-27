/**
 * Automated Verification: Operational Report Access Control & Authorization Fix
 *
 * Verifies:
 * 1. Generate Report button in OperatorCommandCenterView has v-if="canAccessOperationalReport"
 * 2. canAccessOperationalReport computed property returns false for role OPERATOR and true for non-operator roles
 * 3. openReportModal and executeReportGenerate guard against role OPERATOR
 * 4. executeReportGenerate uses authenticated axios.get with blob response handling
 * 5. ReportsView exportReport uses authenticated axios.get with blob response handling
 * 6. Sidebar.vue hides /monitoring/reports for Operator role via v-if="!isOperator"
 * 7. Router index.js enforces RBAC meta roles on /monitoring/reports excluding OPERATOR
 * 8. Backend middleware JWTAuth supports token query parameter fallback
 * 9. Backend middleware opAllowed does not grant report.generate to OPERATOR
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

let passedTests = 0;
let failedTests = 0;

function it(desc, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${desc}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${desc}`);
    console.error(`  Error: ${err.message}`);
    failedTests++;
  }
}

console.log('===================================================================');
console.log('RUNNING TEST SUITE: OPERATIONAL REPORT ACCESS & AUTH GOVERNANCE');
console.log('===================================================================');

const occViewPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

const reportsViewPath = path.join(__dirname, '../src/views/ReportsView.vue');
const reportsContent = fs.readFileSync(reportsViewPath, 'utf8');

const sidebarPath = path.join(__dirname, '../src/components/Sidebar.vue');
const sidebarContent = fs.readFileSync(sidebarPath, 'utf8');

const routerPath = path.join(__dirname, '../src/router/index.js');
const routerContent = fs.readFileSync(routerPath, 'utf8');

const middlewarePath = path.join(__dirname, '../../internal/middleware/middleware.go');
const middlewareContent = fs.readFileSync(middlewarePath, 'utf8');

// Test 1: Button guarded by v-if="canAccessOperationalReport"
it('Test 1: Operational Report button in OCC is guarded by v-if="canAccessOperationalReport"', () => {
  assert(occContent.includes('v-if="canAccessOperationalReport"'), 'Missing v-if="canAccessOperationalReport" on button');
  assert(occContent.includes('@click="openReportModal"'), 'Missing openReportModal handler');
});

// Test 2: canAccessOperationalReport computed property
it('Test 2: canAccessOperationalReport excludes OPERATOR role and permits others', () => {
  assert(occContent.includes('canAccessOperationalReport() {'), 'Missing canAccessOperationalReport computed property');
  assert(occContent.includes("return role !== 'OPERATOR' && role !== '';"), 'canAccessOperationalReport logic mismatch');
});

// Test 3: Method guards against Operator
it('Test 3: openReportModal and executeReportGenerate guard against OPERATOR', () => {
  assert(occContent.includes('if (!this.canAccessOperationalReport) {'), 'Missing method guard in OCC');
});

// Test 4: OCC executeReportGenerate uses axios with blob handling and auth token
it('Test 4: OCC executeReportGenerate uses authenticated axios and blob object URLs', () => {
  assert(occContent.includes("const res = await axios.get('/api/reports/generate'"), 'Missing axios.get for report generation');
  assert(occContent.includes("responseType: 'blob'"), 'Missing blob responseType in OCC report generator');
  assert(occContent.includes('URL.createObjectURL(blob)'), 'Missing blob URL handling in OCC');
});

// Test 5: ReportsView exportReport uses authenticated axios and blob object URLs
it('Test 5: ReportsView exportReport uses authenticated axios and blob object URLs', () => {
  assert(reportsContent.includes("const res = await axios.get('/api/reports/generate'"), 'Missing axios.get in ReportsView export');
  assert(reportsContent.includes("responseType: 'blob'"), 'Missing blob responseType in ReportsView export');
  assert(reportsContent.includes('URL.createObjectURL(blob)'), 'Missing blob URL handling in ReportsView');
});

// Test 6: Sidebar hides /monitoring/reports for Operator
it('Test 6: Sidebar.vue hides /monitoring/reports for Operator role', () => {
  assert(sidebarContent.includes('v-if="!isOperator" to="/monitoring/reports"'), 'Sidebar must hide reports link for Operator');
});

// Test 7: Router guards /monitoring/reports with meta roles
it('Test 7: Router guards /monitoring/reports with meta roles excluding OPERATOR', () => {
  assert(routerContent.includes("path: '/monitoring/reports'"), 'Missing /monitoring/reports in router');
  assert(routerContent.includes("meta: { roles: ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN', 'ENGINEER'] }"), 'Router meta roles missing on /monitoring/reports');
});

// Test 8: Backend JWTAuth supports token query parameter
it('Test 8: Backend JWTAuth middleware supports token query parameter fallback', () => {
  assert(middlewareContent.includes('tokenString = c.Query("token")'), 'JWTAuth must check c.Query("token") as fallback');
});

// Test 9: Backend opAllowed does not grant report.generate to OPERATOR
it('Test 9: Backend opAllowed does not grant report.generate to OPERATOR', () => {
  const opSection = middlewareContent.slice(middlewareContent.indexOf('OPERATOR fallback'));
  assert(!opSection.includes('"report.generate": true'), 'OPERATOR must not have report.generate permission in middleware');
  assert(opSection.includes('"report.view":       true'), 'OPERATOR should still have report.view');
});

// Test 10: Device Monitoring, Telemetry Analysis, and System Health shortcut buttons have v-if="isNonOperator"
it('Test 10: Device Monitoring, Telemetry Analysis, and System Health buttons have v-if="isNonOperator"', () => {
  assert(occContent.includes('v-if="isNonOperator"\n          to="/monitoring/devices"'), 'Device Monitoring button must have v-if="isNonOperator"');
  assert(occContent.includes('v-if="isNonOperator"\n          to="/monitoring/analysis"'), 'Telemetry Analysis button must have v-if="isNonOperator"');
  assert(occContent.includes('v-if="isNonOperator"\n          to="/monitoring/system"'), 'System Health button must have v-if="isNonOperator"');
});

// Test 11: isNonOperator computed property defined
it('Test 11: isNonOperator computed property defined to exclude OPERATOR', () => {
  assert(occContent.includes('isNonOperator() {'), 'Missing isNonOperator computed property');
  assert(occContent.includes("return role !== 'OPERATOR' && role !== '';"), 'isNonOperator logic mismatch');
});

console.log('===================================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===================================================================');

if (failedTests > 0) process.exit(1);
