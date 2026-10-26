const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('TEST SUITE: OPERATOR DASHBOARD PARAMETER & ISPU CARD BORDERS');
console.log('===============================================================');

function runTest(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    return true;
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    return false;
  }
}

let passed = 0;
let total = 0;

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const ispuPath = path.join(__dirname, '../src/components/ISPUIndicatorBar.vue');

const occContent = fs.readFileSync(occPath, 'utf8');
const ispuContent = fs.readFileSync(ispuPath, 'utf8');

// Test 1: Parameter Card when selected for Trend Analysis does NOT have ring-1 or double-stroke shadow
total++;
if (runTest('1.1: Parameter card selected state in OCC template uses clean 1px border without ring-1', () => {
  assert(occContent.includes("isParamSelectedForTrend(param)\n                        ? 'bg-white dark:bg-[#0B1422] border-cyan-500 dark:border-cyan-400"),
    'Must use border-cyan-500 dark:border-cyan-400 without ring-1');
  assert(!occContent.includes('ring-1 ring-cyan-500/25'),
    'Template must not have ring-1 ring-cyan-500/25');
})) passed++;

total++;
if (runTest('1.2: Parameter card CSS in OCC does not simulate thick 2px border with 0 0 0 1px shadow', () => {
  // Check light mode
  assert(!occContent.includes('box-shadow: 0 0 0 1px rgba(34, 184, 212'),
    'Light mode selected card must not have 0 0 0 1px ring shadow');
  // Check dark mode
  assert(!occContent.includes('box-shadow: 0 0 0 1px rgba(33, 199, 232'),
    'Dark mode selected card must not have 0 0 0 1px ring shadow');
  assert(occContent.includes('.light-mode .scada-param-tile.border-cyan-500'),
    'Light mode must have rule for .border-cyan-500');
  assert(occContent.includes('.dark-mode .scada-param-tile.border-cyan-500'),
    'Dark mode must have rule for .border-cyan-500');
})) passed++;

// Test 2: ISPU Breakdown Cards border refinement
total++;
if (runTest('2.1: ISPU Breakdown Cards use 1px border instead of thick border-2', () => {
  assert(!ispuContent.includes('rounded-xl border-2 transition-all flex flex-col justify-between cursor-pointer'),
    'Cards must not use border-2');
  assert(ispuContent.includes('rounded-xl border transition-all flex flex-col justify-between cursor-pointer'),
    'Cards must use border (1px)');
})) passed++;

total++;
if (runTest('2.2: Active ISPU card does not use heavy ring-2 ring-cyan-400', () => {
  assert(!ispuContent.includes("ring-2 ring-cyan-400 dark:ring-cyan-300"),
    'Must not have ring-2 ring-cyan-400 on active ISPU card');
  assert(ispuContent.includes("selectedDisplayParam === pollutant.code ? '!border-cyan-500 dark:!border-cyan-400 shadow-md' : ''"),
    'Active ISPU card must use crisp !border-cyan-500 with shadow-md');
})) passed++;

total++;
if (runTest('2.3: Dominant ISPU card does not use heavy ring-2 ring-amber-400/50', () => {
  assert(!ispuContent.includes("ring-2 ring-amber-400/50"),
    'Must not have ring-2 ring-amber-400/50 on dominant ISPU card');
  assert(ispuContent.includes("border-amber-500 dark:border-amber-400 shadow-sm"),
    'Dominant ISPU card must have clean border-amber-500 without ring-2');
})) passed++;

console.log('===============================================================');
console.log(`TEST SUMMARY: ${passed}/${total} TESTS PASSED`);
console.log('===============================================================');

if (passed !== total) {
  process.exit(1);
}
