const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: OPERATOR CONSOLE SINGLE PARAMETER PANEL + ISPU BELOW');
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

const ispuBarPath = path.join(__dirname, '../src/components/ISPUIndicatorBar.vue');
const ispuBarContent = fs.readFileSync(ispuBarPath, 'utf8');

const enLocalePath = path.join(__dirname, '../src/i18n/locales/en.js');
const enContent = fs.readFileSync(enLocalePath, 'utf8');

const idLocalePath = path.join(__dirname, '../src/i18n/locales/id.js');
const idContent = fs.readFileSync(idLocalePath, 'utf8');

// Test 1: Single unified parameter container
runTest('Test 1: Exactly one single main parameter container contains all visible parameter cards', () => {
  assert(occContent.includes('scada-single-main-panel'), 'Template must have unified scada-single-main-panel container');
  // Must NOT have multiple separate category panels loop for cards view
  assert(!occContent.includes('v-for="section in categorySections" :key="\'category-section-\' + section.id"'), 'Must not render separate category containers');
  assert(occContent.includes('v-for="param in filteredParameters"'), 'Cards must iterate over single filteredParameters list');
});

// Test 2: Category filters available in toolbar and filter inside the single panel
runTest('Test 2: Category filters available in toolbar and filter inside single main panel', () => {
  assert(occContent.includes('selectedCategoryFilter = \'all\''), 'Must have All Categories filter');
  assert(occContent.includes('v-for="cat in activeCategoriesList"'), 'Must render dynamic category filter tabs');
  assert(occContent.includes('selectedCategoryFilter = cat.id'), 'Must filter by category on tab click');
  assert(occContent.includes('selectedCategoryTitle'), 'Must dynamically display active category title');
});

// Test 3: Configurable Grid 1P, 2P, 3P, 4P controls inside single panel
runTest('Test 3: 1P, 2P, 3P, 4P controls change grid columns inside single panel', () => {
  assert(occContent.includes("btn-toolbar-panel-' + cnt"), 'Must have 1P-4P toolbar buttons');
  assert(occContent.includes('setGridColumns(cnt)'), 'Toolbar buttons must call setGridColumns');
  assert(occContent.includes('singlePanelGridClass'), 'Must have singlePanelGridClass computed property');
  assert(occContent.includes('gridColumns === cnt'), 'Active button highlights current grid columns');
  assert(occContent.includes('grid-cols-1'), 'Must include 1 column layout for 1P');
  assert(occContent.includes('md:grid-cols-2'), 'Must include 2 column layout for 2P');
  assert(occContent.includes('lg:grid-cols-3'), 'Must include 3 column layout for 3P');
  assert(occContent.includes('lg:grid-cols-4'), 'Must include 4 column layout for 4P');
});

// Test 4: Default 4P on desktop & localStorage persistence
runTest('Test 4: Default 4P on desktop & localStorage preference persistence', () => {
  assert(occContent.includes('operator_monitoring_grid_cols'), 'Must use localStorage key operator_monitoring_grid_cols');
  assert(occContent.includes('loadGridColumnsConfig'), 'Must have loadGridColumnsConfig method');
  assert(occContent.includes('setGridColumns'), 'Must have setGridColumns method');
  assert(occContent.includes('this.loadGridColumnsConfig()'), 'Must load grid preference on mounted()');
  assert(occContent.includes('gridColumns: 4'), 'Default grid columns must be 4');
});

// Test 5: Compact Parameter Cards specification
runTest('Test 5: Parameter cards display name, device ID, unit, status, sparkline, and timestamp within 58px', () => {
  assert(occContent.includes('h-[58px] min-w-0 box-border'), 'Cards must retain compact 58px height');
  assert(occContent.includes('param.device_code'), 'Card must display device ID / code');
  assert(occContent.includes('param.param_name || param.param_code'), 'Card must display parameter name');
  assert(occContent.includes('formatParamValue(param)'), 'Card must display formatted sensor value');
  assert(occContent.includes('param.unit'), 'Card must display engineering unit');
  assert(occContent.includes('getSparkline('), 'Card must render sparkline');
  assert(occContent.includes('getParamTimingLabel(param)'), 'Card footer must display timestamp / duration');
  assert(occContent.includes('selectParamForTrend('), 'Card must have Plot Trend action');
  assert(occContent.includes('getParamCategoryShortLabel(param)'), 'Card header must display category badge');
});

// Test 6: Layout order: Toolbar -> Single Parameter Panel -> ISPU Panel -> Trend Analysis
runTest('Test 6: Layout order strictly satisfies Toolbar -> Single Parameter Panel -> ISPU Panel -> Trend Analysis', () => {
  const panelIdx = occContent.indexOf('scada-single-main-panel');
  const ispuIdx = occContent.indexOf('<ISPUIndicatorBar');
  const trendIdx = occContent.indexOf('SECTION 2: OPERATIONAL ANALYTICAL SUITE');

  assert(panelIdx !== -1, 'scada-single-main-panel must exist');
  assert(ispuIdx !== -1, '<ISPUIndicatorBar must exist');
  assert(trendIdx !== -1, 'Trend section must exist');

  assert(panelIdx < ispuIdx, 'Parameter Panel must appear BEFORE ISPU Panel');
  assert(ispuIdx < trendIdx, 'ISPU Panel must appear BEFORE Trend Analysis');
});

// Test 7: Preserves ISPU functionality without separate columns
runTest('Test 7: ISPU component provides full compliance metrics and inline trend drawer', () => {
  assert(ispuBarContent.includes('displayISPU'), 'ISPU must show current ISPU value');
  assert(ispuBarContent.includes('ispuCategoryDisplay'), 'ISPU must show category');
  assert(ispuBarContent.includes('dominantPollutant'), 'ISPU must show dominant pollutant');
  assert(ispuBarContent.includes('formattedPeriod'), 'ISPU must show regulatory 24h window');
  assert(ispuBarContent.includes('showTrend'), 'ISPU must support collapsible inline trend');
  assert(ispuBarContent.includes('navigateToAnalysis'), 'ISPU must provide navigation to Full ISPU');
});

// Test 8: Fit-to-screen layout optimizations
runTest('Test 8: Fit-to-screen budget: internal scrolling on overflow, no CSS transform scaling', () => {
  assert(occContent.includes('singlePanelScrollStyle'), 'Single panel must support responsive internal scrolling');
  assert(!occContent.includes('transform: scale('), 'Must not use CSS transform scaling to simulate fit-to-screen');
  assert(occContent.includes('min-w-0'), 'Must use min-w-0 to prevent flexbox clipping');
});

// Test 9: Cards View and Table View preserved
runTest('Test 9: Both Cards View and Table View are preserved with switcher controls', () => {
  assert(occContent.includes("viewMode === 'cards'"), 'Cards view mode supported');
  assert(occContent.includes("viewMode === 'table'"), 'Table view mode supported');
  assert(occContent.includes('@click="viewMode = \'cards\'"'), 'Cards switch button exists');
  assert(occContent.includes('@click="viewMode = \'table\'"'), 'Table switch button exists');
});

// Test 10: Localization & Preserved controls
runTest('Test 10: Localization and existing monitoring controls preserved', () => {
  assert(occContent.includes('filterPinnedOnly'), 'Pinned filter toggle preserved');
  assert(occContent.includes('isAQMSDomain'), 'AQMS domain indicator preserved');
  assert(occContent.includes('allParameters.length'), 'Parameter count indicator preserved');
  assert(enContent.includes('parameterMonitoring'), 'EN translation preserved');
  assert(idContent.includes('parameterMonitoring'), 'ID translation preserved');
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
}
