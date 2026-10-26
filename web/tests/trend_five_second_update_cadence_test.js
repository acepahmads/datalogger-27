const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: TREND ANALYSIS 5-SECOND SENSOR POLLING CADENCE');
console.log('===============================================================');

const occViewPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

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

// 1. advanceLiveTrendWindow cadence
test('1.1: advanceLiveTrendWindow derives update threshold dynamically from sensor polling interval', () => {
  const methodMatch = occContent.match(/advanceLiveTrendWindow\(\)\s*\{([\s\S]*?)\n    \},/);
  assert(methodMatch, 'advanceLiveTrendWindow method must exist');
  const body = methodMatch[1];

  assert(body.includes('let threshold = Math.max(pollInterval, 1000);') || body.includes('let threshold = Math.max(pollInterval, 5000);'), 'Threshold must be derived dynamically from pollInterval');
});

test('1.2: advanceLiveTrendWindow aligns window boundaries strictly to sensor polling interval grid', () => {
  const methodMatch = occContent.match(/advanceLiveTrendWindow\(\)\s*\{([\s\S]*?)\n    \},/);
  const body = methodMatch[1];

  assert(body.includes('const alignedEnd = Math.floor(now / pollInterval) * pollInterval;'), 'Must snap trend window end to pollInterval multiples');
  assert(body.includes('start: alignedEnd - span'), 'start boundary must be alignedEnd - span');
  assert(body.includes('end: alignedEnd'), 'end boundary must be alignedEnd');
  assert(body.includes('if (alignedEnd !== this.trendTimeBounds.end)'), 'Must only re-assign trendTimeBounds when 5-second interval boundary shifts');
});

// 2. fetchTrendData alignment
test('2.1: fetchTrendData aligns live time bounds to polling interval boundary', () => {
  const fetchMatch = occContent.match(/async fetchTrendData\([\s\S]*?\{([\s\S]*?)\n    getTimeframeSpan/);
  assert(fetchMatch, 'fetchTrendData method must exist');
  const body = fetchMatch[1];

  assert(body.includes('const alignedEndMs = Math.floor(nowMs / pollInterval) * pollInterval;'), 'Must snap initial/auto-refresh fetch end time to polling interval');
  assert(body.includes('start: startTime.getTime(),\n          end: endTime.getTime(),'), 'Must set trendTimeBounds to snapped timestamps');
});

// 3. Memoization cache keys
test('3.1: generateChartPath memoizes path key with 5-second polling step', () => {
  assert(
    occContent.includes('${Math.floor(timeStart / pollStep)}_${Math.floor(timeEnd / pollStep)}'),
    'generateChartPath must cache by pollStep interval multiples instead of 1-second timestamps'
  );
});

test('3.2: generatePrimaryAreaPath memoizes path key with 5-second polling step', () => {
  const areaSection = occContent.substring(occContent.indexOf('generatePrimaryAreaPath('), occContent.indexOf('generateChartPolyline('));
  assert(
    areaSection.includes('${Math.floor(timeStart / pollStep)}_${Math.floor(timeEnd / pollStep)}'),
    'generatePrimaryAreaPath must cache by pollStep interval multiples'
  );
});

// 4. recordTrendPoint selective cache invalidation
test('4.1: recordTrendPoint only invalidates primary area cache for the primary series', () => {
  const recordMatch = occContent.match(/recordTrendPoint\(data\)\s*\{([\s\S]*?)\n    buildSplinePath/);
  assert(recordMatch, 'recordTrendPoint method must exist');
  const body = recordMatch[1];

  assert(
    body.includes('if (this.activeTrendSeries.length > 0 && this.activeTrendSeries[0].id === series.id)'),
    'Must only invalidate _cachedAreaKey for the primary series'
  );
});

console.log('===============================================================');
console.log('TEST SUMMARY: ALL 6 5-SECOND UPDATE CADENCE REQUIREMENTS PASSED');
console.log('===============================================================');
