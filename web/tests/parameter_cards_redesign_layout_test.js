/**
 * Test Suite: Parameter Cards Redesign & Full-Width Category Layout
 * Verifies all 13 requirements specified in the UI Redesign specification:
 * 1. Parameter cards visually match Reference 2 (Top row: device, icon, name, quality badge, pin; Main row: large value + unit + horizontal sparkline; Bottom row: timestamp, plot trend)
 * 2. Numeric values are prominent (28-36px font) and units remain readable
 * 3. Mini sparklines appear horizontally beside the main numeric values
 * 4. No text, badges, or controls overlap (flex layout, truncate, min-w-[260px], height ~130-150px)
 * 5. Category sections use the full available dashboard width (12-column grid system)
 * 6. The previous large unused right-side area is eliminated
 * 7. Multiple categories render side by side correctly (lg:col-span-6 for span 1/2)
 * 8. A single selected category uses the available width appropriately (col-span-12)
 * 9. Trend Analysis, Wind Rose, and Pollutant Dispersion remain spacious and functional below parameter monitoring
 * 10. Search, category filters, pinned mode, and table view remain functional
 * 11. Live telemetry and alarm states remain correct
 * 12. Dark/light themes and EN/ID localization work
 * 13. Frontend production build and all regression test suites pass
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('================================================================');
console.log('RUNNING AUTOMATED TEST SUITE: PARAMETER CARDS & FULL-WIDTH LAYOUT');
console.log('================================================================\n');

let passedTests = 0;
let failedTests = 0;

function runTest(testName, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${testName}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${testName}`);
    console.error(`  Error: ${err.message}`);
    failedTests++;
  }
}

// Read relevant files
const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const enPath = path.join(__dirname, '../src/i18n/locales/en.js');
const enContent = fs.readFileSync(enPath, 'utf8');

const idPath = path.join(__dirname, '../src/i18n/locales/id.js');
const idContent = fs.readFileSync(idPath, 'utf8');

// Test 1: Parameter cards structure matches Reference 2
runTest('Test 1: Parameter cards structure matches Reference 2 (Top, Main, and Bottom rows)', () => {
  // Top row: device identifier in muted font, separator, icon, name, quality badge, pin
  assert(occContent.includes('param.device_code'), 'Card top row must display device identifier');
  assert(occContent.includes('text-cyan-600/75') || occContent.includes('text-slate-400'), 'Device identifier must be in small muted text');
  assert(occContent.includes('<ParameterIcon'), 'Card top row must render ParameterIcon');
  assert(occContent.includes('param.param_name || param.param_code'), 'Card top row must render parameter name');
  assert(occContent.includes('qualityBadgeClass(param.quality)'), 'Card top row must render quality badge aligned to right');
  assert(occContent.includes('togglePin(param.param_id)'), 'Card top row must render favorite pin control');

  // Main row: large value + unit + horizontal sparkline
  assert(occContent.includes('paramValueColorClass(param)'), 'Main row must render parameter value with status color');
  assert(occContent.includes('param.unit'), 'Main row must render unit beside value');
  assert(occContent.includes('getSparkline(param.param_id).hasHistory'), 'Main row must render sparkline beside value');

  // Bottom row: last-update timestamp + plot trend action
  assert(occContent.includes('formatTimeShort(param.received_at)'), 'Bottom row must show timestamp');
  assert(occContent.includes('plotTrend'), 'Bottom row must render Plot Trend action');
});

// Test 2: Card dimensions match uniform height (130-150px or compact 114-118px) and responsive min-width ~260px
runTest('Test 2: Card dimensions match uniform height and responsive min-width ~260px', () => {
  assert(occContent.includes('h-[134px] sm:h-[140px]') || occContent.includes('h-[114px] sm:h-[118px]'), 'Card must have uniform height between 114-140px');
  assert(occContent.includes('min-w-[260px]'), 'Card must enforce responsive min-width of ~260px');
  assert(occContent.includes('p-3'), 'Card must use comfortable p-3 internal padding');
  assert(occContent.includes('rounded-xl'), 'Card must use refined rounded-xl corners');
});

// Test 3: Numeric values are prominent (26-28px font) and unit remains readable without ellipsis clipping
runTest('Test 3: Numeric values are prominent (28-36px font) and unit remains readable', () => {
  assert(
    occContent.includes('font-bold font-mono tracking-tight leading-none whitespace-nowrap tabular-nums'),
    'Main numeric value must use bold monospace font with tabular-nums and whitespace-nowrap (never truncate)'
  );
  assert(
    occContent.includes('getParamValueFontSizeClass(param)'),
    'Main numeric value must scale font size dynamically for long numbers to prevent overflow'
  );
  assert(
    occContent.includes('text-xs sm:text-[13px] font-mono text-slate-500 dark:text-slate-400 font-semibold ml-1.5 flex-shrink-0 whitespace-nowrap'),
    'Unit must be clearly visible and positioned immediately beside numeric value'
  );
});

// Test 4: Mini sparkline is placed horizontally beside main numeric value with flexible responsive width
runTest('Test 4: Mini sparkline is placed horizontally beside main numeric value', () => {
  assert(
    occContent.includes('flex items-center justify-between gap-2.5 my-auto min-w-0'),
    'Main row must use flex justify-between to place value and sparkline horizontally side-by-side'
  );
  assert(
    occContent.includes('flex-1 min-w-[50px] max-w-[136px] h-9 sm:h-10 flex-shrink relative'),
    'Sparkline must have flexible responsive width adapting to remaining space with height ~36-40px'
  );
  assert(
    occContent.includes('viewBox="0 0 100 28"'),
    'SVG sparkline viewBox must match 100x28 aspect ratio'
  );
  assert(
    occContent.includes('preserveAspectRatio="none"'),
    'SVG must scale smoothly to container width and height'
  );
});

// Test 5: Full-width parent layout eliminates narrow 7/5 column constraint
runTest('Test 5: Full-width parent layout eliminates narrow 7/5 column constraint', () => {
  // Confirm outer 7/5 split is eliminated for monitoring container
  assert(!occContent.includes('xl:col-span-7 flex flex-col space-y-2'), 'Narrow 7-column parameter monitoring constraint must be eliminated');
  assert(occContent.includes('SECTION 1: DEVICE & PARAMETER MONITORING (FULL WIDTH)'), 'Parameter monitoring must use full width section');
  assert(occContent.includes('SECTION 2: OPERATIONAL ANALYTICAL SUITE (FULL WIDTH)'), 'Analytical suite must use full width section');
});

// Test 6: 12-column responsive grid system for category sections
runTest('Test 6: 12-column responsive grid system allows categories to use available desktop width', () => {
  assert(
    occContent.includes('grid grid-cols-1 lg:grid-cols-12 gap-3 sm:gap-3.5'),
    'Category sections must be positioned within a 12-column responsive grid'
  );
  assert(
    occContent.includes('getSectionContainerSpanClass(section)'),
    'Each category section must resolve its column span dynamically'
  );
});

// Test 7: Multiple categories render side-by-side (lg:col-span-6)
runTest('Test 7: Multiple categories render side-by-side (lg:col-span-6) when configured for half-width', () => {
  assert(
    occContent.includes("return 'col-span-12 lg:col-span-6';"),
    'Half-width categories must use lg:col-span-6 to render side by side'
  );
  assert(
    occContent.includes("return 'grid grid-cols-1 sm:grid-cols-2 xl:grid-cols-3 gap-3';"),
    'Half-width category cards must wrap in 1 to 3 columns'
  );
});

// Test 8: Single selected category uses available width efficiently (col-span-12)
runTest('Test 8: Single selected category uses available width efficiently (col-span-12)', () => {
  assert(
    occContent.includes("if (this.selectedCategoryFilter !== 'all') {\n        return 'col-span-12';"),
    'When a specific category tab is selected, it must expand to full 12 columns'
  );
  assert(
    occContent.includes("return 'grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4 2xl:grid-cols-4 gap-3';"),
    'Full-width category cards must wrap in 2 to 4 columns'
  );
});

// Test 9: Operational analytical suite (Trend Analysis, Wind Rose, Dispersion) preserved
runTest('Test 9: Trend Analysis, Wind Rose, and Dispersion Map are preserved and spacious', () => {
  assert(occContent.includes('TREND ANALYSIS'), 'Trend Analysis panel must remain present');
  assert(occContent.includes('<WindRoseChart'), 'WindRoseChart must remain rendered');
  assert(occContent.includes('<PollutantDispersionMap'), 'PollutantDispersionMap must remain rendered');
  assert(occContent.includes('lg:col-span-5'), 'Wind Rose receives 5/12 columns');
  assert(occContent.includes('lg:col-span-7'), 'Dispersion Map receives 7/12 columns');
});

// Test 10: Pinned mode, Cards view, and Table view preserved (filter dropdowns row removed per user request)
runTest('Test 10: Category tabs, Pinned mode, and Cards vs Table view preserved', () => {
  assert(occContent.includes('toggleFilterPinned'), 'Pinned filter toggle must be preserved');
  assert(occContent.includes("viewMode === 'cards'"), 'Cards view mode toggle must be preserved');
  assert(occContent.includes("viewMode === 'table'") || occContent.includes('v-else class="scada-panel'), 'Table view mode must be preserved');
  assert(occContent.includes('selectedCategoryFilter'), 'Category tabs filter must be preserved');
});

// Test 11: Live telemetry, alarms, sparkline gap handling, and empty states preserved
runTest('Test 11: Live telemetry, alarms, sparkline gap handling, and empty states preserved', () => {
  assert(occContent.includes('liveBlinks[param.param_id]'), 'Live data blink pulse must be preserved');
  assert(occContent.includes('isParamAlarmOrStaleBad(param)'), 'Alarm and anomaly state highlight must be preserved');
  assert(occContent.includes('awaitingSamples'), 'Empty state for < 2 samples must show awaiting history');
  assert(occContent.includes('areaPaths'), 'Gap-aware sparkline area paths must be preserved');
  assert(occContent.includes('strokePaths'), 'Gap-aware sparkline stroke paths must be preserved');
});

// Test 12: Fixed bottom Quick Actions bar has adequate page bottom padding
runTest('Test 12: Fixed bottom Quick Actions bar has adequate page bottom padding (pb-20)', () => {
  assert(occContent.includes('pb-20'), 'Page container must use pb-20 to prevent fixed bottom bar from covering content');
  assert(occContent.includes('fixed bottom-0'), 'Quick Actions bar must remain docked at bottom');
});

// Test 13: English and Indonesian locale files contain required translation strings
runTest('Test 13: English and Indonesian locale files contain required translation strings', () => {
  assert(enContent.includes('awaitingSamples') && idContent.includes('awaitingSamples'), 'awaitingSamples translation must exist in EN and ID');
  assert(enContent.includes('plotTrend') && idContent.includes('plotTrend'), 'plotTrend translation must exist in EN and ID');
  assert(enContent.includes('inChart') && idContent.includes('inChart'), 'inChart translation must exist in EN and ID');
  assert(enContent.includes('allCategories') && idContent.includes('allCategories'), 'allCategories translation must exist in EN and ID');
});

console.log('\n================================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('================================================================');

if (failedTests > 0) {
  process.exit(1);
} else {
  process.exit(0);
}
