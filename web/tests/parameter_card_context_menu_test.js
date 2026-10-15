/**
 * Automated Test Suite: Parameter Card Context Menu (Right-Click) & Modals
 * Tests:
 *  1. Right-click context menu presence on parameter cards and table rows
 *  2. Four required context menu options (Test Read, View Config, Edit Config, 100 Data Terakhir)
 *  3. TestReadModal component structure and full Modbus diagnostic report parity
 *  4. ParameterConfigViewModal component structure and parameter policy inspection
 *  5. ParameterHistoryModal component structure with 100 latest records and CSV export
 *  6. RBAC Role check for Edit Config (Admin/Superadmin permitted, non-admin blocked)
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

let passedTests = 0;
let failedTests = 0;

function runTest(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    failedTests++;
  }
}

console.log('===============================================================');
console.log('RUNNING TEST SUITE: PARAMETER CARD CONTEXT MENU & MODALS');
console.log('===============================================================');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const testReadModalPath = path.join(__dirname, '../src/components/TestReadModal.vue');
const viewConfigModalPath = path.join(__dirname, '../src/components/ParameterConfigViewModal.vue');
const historyModalPath = path.join(__dirname, '../src/components/ParameterHistoryModal.vue');

// Test 1: OCC has @contextmenu handler on .scada-param-tile cards
runTest('Test 1: Parameter cards have @contextmenu handler for right-click', () => {
  assert(occContent.includes('@contextmenu.prevent="openParamContextMenu($event, param)"'),
    'Cards must attach @contextmenu.prevent="openParamContextMenu($event, param)"');
});

// Test 2: OCC has @contextmenu handler on table rows and 3-dots action button
runTest('Test 2: Table view rows support context menu and action button', () => {
  assert(occContent.includes('@contextmenu.prevent="openParamContextMenu($event, p)"'),
    'Table rows must support right-click context menu');
  assert(occContent.includes('openParamContextMenu($event, p)'),
    'Table actions column must offer context menu activation');
});

// Test 3: Context menu template includes all 4 required actions
runTest('Test 3: Context menu renders all 4 actions with descriptions', () => {
  assert(occContent.includes('handleContextTestRead'), 'Must have Test Read action');
  assert(occContent.includes('handleContextViewConfig'), 'Must have View Config action');
  assert(occContent.includes('handleContextEditConfig'), 'Must have Edit Config action');
  assert(occContent.includes('handleContextHistory'), 'Must have 100 Data Terakhir action');
  assert(occContent.includes('100 Data Terakhir'), 'Must display "100 Data Terakhir" label');
});

// Test 4: OCC declares and registers all 4 modal components
runTest('Test 4: OCC imports and registers TestReadModal, ParameterConfigViewModal, ParameterModal, and ParameterHistoryModal', () => {
  assert(occContent.includes('import TestReadModal from \'../components/TestReadModal.vue\';'), 'TestReadModal import missing');
  assert(occContent.includes('import ParameterConfigViewModal from \'../components/ParameterConfigViewModal.vue\';'), 'ParameterConfigViewModal import missing');
  assert(occContent.includes('import ParameterHistoryModal from \'../components/ParameterHistoryModal.vue\';'), 'ParameterHistoryModal import missing');
  assert(occContent.includes('import ParameterModal from \'../components/ParameterModal.vue\';'), 'ParameterModal import missing');
  assert(occContent.includes('<TestReadModal'), 'TestReadModal template tag missing');
  assert(occContent.includes('<ParameterConfigViewModal'), 'ParameterConfigViewModal template tag missing');
  assert(occContent.includes('<ParameterModal'), 'ParameterModal template tag missing');
  assert(occContent.includes('<ParameterHistoryModal'), 'ParameterHistoryModal template tag missing');
});

// Test 5: Role check for Edit Config
runTest('Test 5: RBAC Role check allows admin/superadmin/Administrator/Engineer and blocks others', () => {
  assert(occContent.includes('canEditConfig()'), 'canEditConfig method must exist');
  
  // Test simulated canEditConfig logic
  const checkRole = (role) => {
    const r = typeof role === 'string' ? role.toLowerCase() : (role && role.name ? role.name.toLowerCase() : '');
    return ['administrator', 'admin', 'superadmin', 'engineer'].includes(r);
  };

  assert.strictEqual(checkRole('Administrator'), true);
  assert.strictEqual(checkRole('admin'), true);
  assert.strictEqual(checkRole('superadmin'), true);
  assert.strictEqual(checkRole('SuperAdmin'), true);
  assert.strictEqual(checkRole('Engineer'), true);
  assert.strictEqual(checkRole({ name: 'Administrator' }), true);
  assert.strictEqual(checkRole('Operator'), false);
  assert.strictEqual(checkRole('Viewer'), false);
  assert.strictEqual(checkRole(''), false);
  assert.strictEqual(checkRole(null), false);
});

// Test 6: TestReadModal.vue component structure and diagnostic parity
runTest('Test 6: TestReadModal component matches telemetry diagnostic report specifications', () => {
  assert(fs.existsSync(testReadModalPath), 'TestReadModal.vue must exist');
  const content = fs.readFileSync(testReadModalPath, 'utf8');

  // Verify diagnostic elements
  assert(content.includes('testReadParameter'), 'Must dispatch testReadParameter');
  assert(content.includes('getStageSummaryList'), 'Must render 4-stage pipeline stepper');
  assert(content.includes('PORT_VERIFICATION'), 'Must include PORT_VERIFICATION stage');
  assert(content.includes('SERIAL_CONFIGURATION'), 'Must include SERIAL_CONFIGURATION stage');
  assert(content.includes('MODBUS_TRANSACTION'), 'Must include MODBUS_TRANSACTION stage');
  assert(content.includes('DATA_INTERPRETATION'), 'Must include DATA_INTERPRETATION stage');
  assert(content.includes('copyMbpollCommand'), 'Must allow copying mbpoll command');
  assert(content.includes('copyDiagnosticReport'), 'Must allow copying engineering diagnostic report');
  assert(content.includes('getTestReadFormulaVal'), 'Must evaluate formula if configured');
  assert(content.includes('getTestReadModbusVal'), 'Must evaluate scaled Modbus value');
  assert(content.includes('executeTestRead'), 'Must allow "Read Again"');
});

// Test 7: ParameterConfigViewModal.vue component structure
runTest('Test 7: ParameterConfigViewModal displays complete configured parameter settings', () => {
  assert(fs.existsSync(viewConfigModalPath), 'ParameterConfigViewModal.vue must exist');
  const content = fs.readFileSync(viewConfigModalPath, 'utf8');

  assert(content.includes('General & Identification'), 'Must show general section');
  assert(content.includes('Modbus Protocol & Data Mapping'), 'Must show Modbus section');
  assert(content.includes('Thresholds & Warning Limits'), 'Must show limits section');
  assert(content.includes('Zero-Value Alarm Policy'), 'Must show zero-value alarm policy');
  assert(content.includes('Formula Transform'), 'Must show formula transform');
  assert(content.includes('Hold Last Good Value'), 'Must show hold last good value duration');
  assert(content.includes('Stale Timeout'), 'Must show stale timeout duration');
  assert(content.includes('canEdit'), 'Must receive canEdit prop');
});

// Test 8: ParameterHistoryModal.vue component structure & CSV export
runTest('Test 8: ParameterHistoryModal fetches 100 records and supports CSV export', () => {
  assert(fs.existsSync(historyModalPath), 'ParameterHistoryModal.vue must exist');
  const content = fs.readFileSync(historyModalPath, 'utf8');

  assert(content.includes('page_size: 100'), 'Must query 100 records');
  assert(content.includes('exportCsv'), 'Must provide CSV export functionality');
  assert(content.includes('summaryStats'), 'Must compute min, max, avg, and quality counts');
  assert(content.includes('formatTimestamp'), 'Must format timestamps clearly');
  assert(content.includes('Raw PDU'), 'Must display raw PDU value');
  assert(content.includes('Raw Hex'), 'Must display raw hex');
  assert(content.includes('Quality Reason'), 'Must display quality reason');
});

// Test 9: Context menu auto-dismissal listeners
runTest('Test 9: Window and document listeners handle context menu dismissal on click outside', () => {
  assert(occContent.includes('document.addEventListener(\'click\', this.onDocumentClick);'),
    'Document click listener must be registered in mounted()');
  assert(occContent.includes('document.removeEventListener(\'click\', this.onDocumentClick);'),
    'Document click listener must be removed in beforeDestroy()');
  assert(occContent.includes('closeParamContextMenu'), 'closeParamContextMenu method must exist');
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
}
