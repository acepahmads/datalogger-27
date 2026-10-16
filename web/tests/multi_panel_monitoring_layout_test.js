const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING AUTOMATED TEST SUITE: CONFIGURABLE MULTI-PANEL PARAMETER LAYOUT');
console.log('===============================================================');

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

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const modalPath = path.join(__dirname, '../src/components/MultiPanelConfigModal.vue');
const modalContent = fs.readFileSync(modalPath, 'utf8');

const enLocalePath = path.join(__dirname, '../src/i18n/locales/en.js');
const enContent = fs.readFileSync(enLocalePath, 'utf8');

const idLocalePath = path.join(__dirname, '../src/i18n/locales/id.js');
const idContent = fs.readFileSync(idLocalePath, 'utf8');

// Test 1: MultiPanelConfigModal component existence & template syntax
runTest('Test 1: MultiPanelConfigModal component is created and registered in OCC', () => {
  assert(fs.existsSync(modalPath), 'MultiPanelConfigModal.vue must exist');
  assert(occContent.includes("import MultiPanelConfigModal from '../components/MultiPanelConfigModal.vue'"), 'Must import MultiPanelConfigModal');
  assert(occContent.includes('MultiPanelConfigModal,'), 'Must register MultiPanelConfigModal in components');
  assert(occContent.includes('<MultiPanelConfigModal'), 'Must mount MultiPanelConfigModal in template');
});

// Test 2: Configurable panel count (1, 2, 3, 4 panels) controls in toolbar
runTest('Test 2: Configurable panel count (1P, 2P, 3P, 4P) preset controls and modal trigger in OCC toolbar', () => {
  assert(occContent.includes("btn-toolbar-panel-' + cnt"), 'Toolbar must have dynamic button for panel counts');
  assert(occContent.includes('setPanelCount(cnt)'), 'Toolbar must invoke setPanelCount');
  assert(occContent.includes('btn-open-panel-config'), 'Toolbar must have button to open panel configuration modal');
});

// Test 3: Horizontal responsive layout with CSS Grid & min-width: 0
runTest('Test 3: Horizontal responsive layout uses CSS Grid with proportional columns and min-width: 0', () => {
  assert(occContent.includes('multiPanelGridClass'), 'Must have multiPanelGridClass computed property');
  assert(occContent.includes('grid-cols-1'), 'Must include 1 column layout');
  assert(occContent.includes('md:grid-cols-2'), 'Must include 2 column layout for medium screens');
  assert(occContent.includes('lg:grid-cols-3'), 'Must include 3 column layout for large screens');
  assert(occContent.includes('lg:grid-cols-4'), 'Must include 4 column layout for extra large screens');
  assert(occContent.includes('scada-monitoring-panel flex flex-col min-w-0'), 'Panels must use min-w-0 to prevent overflowing container');
});

// Test 4: Panel structure: configurable title, parameter count badge, compact card grid, internal scroll
runTest('Test 4: Panel layout has configurable title, count badge, compact card grid, and internal scroll area', () => {
  assert(occContent.includes('{{ panel.title }}'), 'Panel must display dynamic configurable title');
  assert(occContent.includes('getPanelParameters(panel, pIdx).length'), 'Panel must display parameter count badge');
  assert(occContent.includes('panelScrollAreaStyle'), 'Panel must have internal scroll area style');
  assert(occContent.includes('custom-scroller overflow-y-auto flex-1'), 'Panel must scroll internally only when content exceeds available height');
  assert(occContent.includes('getPanelCardGridStyle(panel)'), 'Panel must use compact dynamic card grid');
});

// Test 5: Parameter Assignment logic: by category, custom parameter list, and deduplication
runTest('Test 5: Assignment logic supports categories, individual parameters, and deduplication', () => {
  assert(occContent.includes('getPanelParameters('), 'OCC must have getPanelParameters method');
  assert(occContent.includes('getPrecedingClaimedIds('), 'OCC must have getPrecedingClaimedIds method to prevent accidental duplicates');
  assert(occContent.includes('allowDuplicates'), 'Panel config must support optional explicit allowDuplicates flag');
});

// Test 6: Zero parameter loss guard: newly registered parameters and unassigned parameters swept into last panel
runTest('Test 6: Safety guard ensures newly registered or unassigned parameters are never lost silently', () => {
  assert(occContent.includes('actualIdx === this.activePanels.length - 1'), 'Last panel must sweep in unassigned parameters');
  assert(modalContent.includes('assignUnassignedToLastPanel'), 'Modal must provide auto-assign button for unassigned parameters');
  assert(modalContent.includes('unassignedParamsList'), 'Modal must track unassigned parameters list');
});

// Test 7: Persistence: localStorage preference saving and retrieval with safe defaults
runTest('Test 7: Panel layout preferences persist to localStorage with safe default', () => {
  assert(occContent.includes('occ_multipanel_layout_config'), 'Must use localStorage key occ_multipanel_layout_config');
  assert(occContent.includes('loadPanelLayoutConfig'), 'Must have loadPanelLayoutConfig method');
  assert(occContent.includes('savePanelLayoutConfig'), 'Must have savePanelLayoutConfig method');
  assert(occContent.includes('this.loadPanelLayoutConfig()'), 'Must load configuration on mounted()');
});

// Test 8: MultiPanelConfigModal features: add/remove, rename, reorder, assign, duplicate toggle
runTest('Test 8: MultiPanelConfigModal supports add, remove, rename, reorder left/right, and mode toggle', () => {
  assert(modalContent.includes('addNewPanel'), 'Modal must support adding new panels');
  assert(modalContent.includes('removePanel'), 'Modal must support removing panels');
  assert(modalContent.includes('movePanel('), 'Modal must support moving panels');
  assert(modalContent.includes('resetToDefaults'), 'Modal must support resetting to smart defaults');
  assert(modalContent.includes('v-model="panel.title"'), 'Modal must allow editing panel title');
  assert(modalContent.includes('panel.assignmentMode'), 'Modal must support category vs individual parameter assignment');
});

// Test 9: Complete i18n coverage in English and Indonesian locales
runTest('Test 9: English and Indonesian locale files contain all required multi-panel translation keys', () => {
  const requiredKeys = [
    'multiPanel',
    'panelCount',
    'panels',
    'panelsConfig',
    'addPanel',
    'removePanel',
    'panelTitle',
    'panelColor',
    'assignmentMode',
    'byCategory',
    'byParameter',
    'assignedCategories',
    'assignedParameters',
    'unassignedNotice',
    'autoAssignUnassigned',
    'resetPanels',
    'moveLeft',
    'moveRight',
    'noParamsInPanel',
    'preset1P',
    'preset2P',
    'preset3P',
    'preset4P'
  ];

  for (const key of requiredKeys) {
    const hasEn = enContent.includes(`"${key}":`) || enContent.includes(`${key}:`);
    const hasId = idContent.includes(`"${key}":`) || idContent.includes(`${key}:`);
    assert(hasEn, `en.js missing operator.${key}`);
    assert(hasId, `id.js missing operator.${key}`);
  }
});

// Test 10: Preserved features: card compact dimensions (~58px), sparkline, telemetry freshness, Plot Trend
runTest('Test 10: Preserves compact 58px card, sparklines, live telemetry, and Plot Trend actions', () => {
  assert(occContent.includes('h-[58px] min-w-0 box-border'), 'Cards must retain compact 58px height');
  assert(occContent.includes('getSparkline('), 'Cards must retain mini sparklines');
  assert(occContent.includes('selectParamForTrend('), 'Cards must retain Plot Trend action');
  assert(occContent.includes('formatParamValue('), 'Cards must retain numeric value formatting');
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
}
