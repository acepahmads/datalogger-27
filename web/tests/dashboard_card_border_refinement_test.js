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

// Test 1: Parameter Card when selected for Trend Analysis uses delicate subtle border and no shadow
total++;
if (runTest('1.1: Parameter card selected state in OCC template uses delicate border without ring-1', () => {
  assert(occContent.includes("isParamSelectedForTrend(param)\n                        ? 'bg-white dark:bg-[#0B1422] border-cyan-500/40 dark:border-cyan-400/35"),
    'Must use border-cyan-500/40 dark:border-cyan-400/35 without ring-1');
  assert(!occContent.includes('ring-1 ring-cyan-500/25'),
    'Template parameter card must not have ring-1 ring-cyan-500/25');
})) passed++;

total++;
if (runTest('1.2: Parameter card CSS in OCC uses delicate rgba cyan and box-shadow: none', () => {
  assert(occContent.includes('border-color: rgba(6, 182, 212, 0.45) !important;\n  box-shadow: none !important;'),
    'Light mode selected card must have rgba(6, 182, 212, 0.45) with box-shadow: none');
  assert(occContent.includes('border-color: rgba(33, 199, 232, 0.35) !important;\n  box-shadow: none !important;'),
    'Dark mode selected card must have rgba(33, 199, 232, 0.35) with box-shadow: none');
})) passed++;

// Test 2: ISPU Breakdown Cards border refinement
total++;
if (runTest('2.1: ISPU Breakdown Cards use delicate 1px border instead of thick border-2', () => {
  assert(!ispuContent.includes('rounded-xl border-2'),
    'Cards must not use border-2');
  assert(ispuContent.includes('class="ispu-breakdown-tile p-2 sm:p-2.5 rounded-lg border'),
    'Cards must use ispu-breakdown-tile with 1px border');
  assert(ispuContent.includes('.ispu-breakdown-tile {\n  border-width: 1px !important;\n  box-shadow: none !important;\n}'),
    'Must have scoped style enforcing 1px border and box-shadow: none');
})) passed++;

total++;
if (runTest('2.2: Active ISPU card uses delicate !border-cyan-500/50 without ring-2 or shadow-md', () => {
  assert(!ispuContent.includes("ring-2 ring-cyan-400"),
    'Must not have ring-2 ring-cyan-400 on active ISPU card');
  assert(ispuContent.includes("selectedDisplayParam === pollutant.code ? '!border-cyan-500/50 dark:!border-cyan-400/40' : ''"),
    'Active ISPU card must use delicate !border-cyan-500/50 dark:!border-cyan-400/40');
})) passed++;

total++;
if (runTest('2.3: Dominant and category ISPU cards use delicate 20-35% opacity borders without ring-2', () => {
  assert(!ispuContent.includes("ring-2 ring-amber-400"),
    'Must not have ring-2 ring-amber-400 on dominant ISPU card');
  assert(ispuContent.includes("border-amber-500/35 dark:border-amber-400/30"),
    'Dominant ISPU card must have delicate border-amber-500/35');
  assert(ispuContent.includes("border-emerald-500/25 dark:border-emerald-500/20"),
    'Category card must have delicate 20-25% opacity border');
})) passed++;

console.log('===============================================================');
console.log(`TEST SUMMARY: ${passed}/${total} TESTS PASSED`);
console.log('===============================================================');

if (passed !== total) {
  process.exit(1);
}
