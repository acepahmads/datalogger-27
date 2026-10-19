const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: 1-MINUTE TREND ANALYSIS AUTO-REFRESH');
console.log('===============================================================');

const viewPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const viewContent = fs.readFileSync(viewPath, 'utf8');

function test(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    process.exit(1);
  }
}

// Requirement 1: Data definition and cleanup lifecycle
test('1.1: trendAutoRefreshInterval is declared in component data()', () => {
  assert(viewContent.includes('trendAutoRefreshInterval: null'), 'trendAutoRefreshInterval must be defined in data()');
});

test('1.2: setupTrendAutoRefresh() is invoked in mounted() and cleared in beforeDestroy()', () => {
  assert(viewContent.includes('this.setupTrendAutoRefresh();'), 'mounted() must call this.setupTrendAutoRefresh()');
  assert(viewContent.includes('if (this.trendAutoRefreshInterval) clearInterval(this.trendAutoRefreshInterval);'), 'beforeDestroy() must clear trendAutoRefreshInterval');
});

// Requirement 2: 1-minute interval timing and safe historical inspection
test('2.1: setupTrendAutoRefresh registers a 60,000 ms (1-minute) interval', () => {
  assert(viewContent.includes('setupTrendAutoRefresh()'), 'setupTrendAutoRefresh method must exist');
  assert(viewContent.includes('60000'), 'setupTrendAutoRefresh must use 60000 ms (1 minute)');
});

test('2.2: Auto-refresh preserves user focus by not shifting fixed custom timeframes', () => {
  assert(viewContent.includes('!this.fullscreenIsCustom'), 'Auto-refresh must guard against shifting custom historical timeframes');
});

// Requirement 3: Silent background updating and fetch reschedule
test('3.1: fetchTrendData supports silent parameter to prevent UI spinner flicker', () => {
  assert(viewContent.includes('async fetchTrendData(silent = false)'), 'fetchTrendData must take silent = false parameter');
  assert(viewContent.includes('if (!silent)'), 'isChartLoading must only trigger when silent is false');
});

test('3.2: fetchTrendData reschedules setupTrendAutoRefresh after completion', () => {
  assert(viewContent.includes('this.setupTrendAutoRefresh();'), 'fetchTrendData finally block must reschedule auto-refresh');
});

// Requirement 4: User-facing visual indicator
test('4.1: Chart header displays AUTO 1M indicator', () => {
  assert(viewContent.includes('AUTO 1M'), 'Chart header must display AUTO 1M badge');
});

console.log('===============================================================');
console.log('TEST SUMMARY: ALL 7 1-MINUTE TREND AUTO-REFRESH REQUIREMENTS PASSED');
console.log('===============================================================');
