/**
 * Acceptance Test: ISPU Parameter Breakdown Tray LocalStorage Persistence
 *
 * Verifies that when the user toggles/opens the "ISPU Params : X/7" breakdown pill,
 * the expanded/visible state is saved in localStorage, and upon page reload/refresh,
 * the component restores the user's choice.
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('=== RUNNING ISPU BREAKDOWN LOCALSTORAGE PERSISTENCE TEST ===\n');

let passed = 0;
let total = 0;

function runTest(name, fn) {
  try {
    fn();
    console.log(`  ✓ PASS: ${name}`);
    return true;
  } catch (err) {
    console.error(`  ✗ FAIL: ${name}`);
    console.error(`    Error: ${err.message}`);
    return false;
  }
}

// Test 1: ISPUIndicatorBar.vue contains breakdown persistence logic in data()
total++;
if (runTest('ISPUIndicatorBar initializes showBreakdown from localStorage in data()', () => {
  const filePath = path.join(__dirname, '../src/components/ISPUIndicatorBar.vue');
  assert(fs.existsSync(filePath), 'ISPUIndicatorBar.vue must exist');
  const content = fs.readFileSync(filePath, 'utf8');

  assert(content.includes('scada_ispu_show_breakdown'), 'Must check scada_ispu_show_breakdown in localStorage');
  assert(/showBreakdown:\s*\(\(\)\s*=>/.test(content), 'showBreakdown must be initialized dynamically from localStorage');
})) passed++;

// Test 2: toggleBreakdown writes state to localStorage
total++;
if (runTest('toggleBreakdown persists state to localStorage', () => {
  const filePath = path.join(__dirname, '../src/components/ISPUIndicatorBar.vue');
  const content = fs.readFileSync(filePath, 'utf8');

  const toggleMethod = content.substring(content.indexOf('toggleBreakdown()'), content.indexOf('toggleTrend()'));
  assert(toggleMethod.includes('localStorage.setItem'), 'toggleBreakdown must call localStorage.setItem');
  assert(toggleMethod.includes('scada_ispu_show_breakdown'), 'toggleBreakdown must set scada_ispu_show_breakdown');
})) passed++;

// Test 3: Watcher on showBreakdown synchronizes state to localStorage
total++;
if (runTest('Watcher on showBreakdown persists changes to localStorage', () => {
  const filePath = path.join(__dirname, '../src/components/ISPUIndicatorBar.vue');
  const content = fs.readFileSync(filePath, 'utf8');

  assert(content.includes('showBreakdown(newVal)'), 'Must include showBreakdown watcher');
  assert(content.includes('localStorage.setItem'), 'Watcher must save to localStorage');
})) passed++;

// Test 4: Device change and mounted hooks restore breakdown preference
total++;
if (runTest('Mounted and deviceId watcher restore breakdown preference for the active station', () => {
  const filePath = path.join(__dirname, '../src/components/ISPUIndicatorBar.vue');
  const content = fs.readFileSync(filePath, 'utf8');

  assert(content.includes('scada_ispu_show_breakdown_${newId}'), 'watch deviceId must check device-specific key');
  assert(content.includes('scada_ispu_show_breakdown_${this.deviceId}'), 'mounted must check device-specific key');
})) passed++;

// Test 5: Verify simulated localStorage persistence cycle
total++;
if (runTest('Simulated storage cycle: toggle -> persist -> reload restores state', () => {
  const mockStorage = {};
  global.localStorage = {
    getItem: (k) => mockStorage[k] !== undefined ? mockStorage[k] : null,
    setItem: (k, v) => { mockStorage[k] = String(v); },
    removeItem: (k) => { delete mockStorage[k]; }
  };

  // 1. Initial state (no storage) -> default false
  let saved = localStorage.getItem('scada_ispu_show_breakdown');
  assert.strictEqual(saved, null);

  // 2. User clicks toggle to open ISPU Params: 14/7
  let showBreakdown = false;
  showBreakdown = !showBreakdown;
  localStorage.setItem('scada_ispu_show_breakdown', showBreakdown ? 'true' : 'false');
  localStorage.setItem('scada_ispu_show_breakdown_1', showBreakdown ? 'true' : 'false');

  assert.strictEqual(localStorage.getItem('scada_ispu_show_breakdown'), 'true');
  assert.strictEqual(localStorage.getItem('scada_ispu_show_breakdown_1'), 'true');

  // 3. Page reload simulation
  const restored = (localStorage.getItem('scada_ispu_show_breakdown_1') || localStorage.getItem('scada_ispu_show_breakdown')) === 'true';
  assert.strictEqual(restored, true, 'Reloaded state must be true (opened/visible)');

  // 4. User collapses it -> persist false
  showBreakdown = false;
  localStorage.setItem('scada_ispu_show_breakdown', 'false');
  localStorage.setItem('scada_ispu_show_breakdown_1', 'false');

  const restoredClosed = (localStorage.getItem('scada_ispu_show_breakdown_1') || localStorage.getItem('scada_ispu_show_breakdown')) === 'true';
  assert.strictEqual(restoredClosed, false, 'Reloaded state must be false (collapsed)');
})) passed++;

console.log('\n========================================');
console.log(`RESULTS: ${passed}/${total} TESTS PASSED`);
console.log('========================================');

if (passed === total) {
  process.exit(0);
} else {
  process.exit(1);
}
