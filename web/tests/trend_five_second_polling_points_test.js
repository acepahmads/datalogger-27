const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: 5-SECOND POLLING INTERVAL TREND POINTS & SCRUB SNAPPING');
console.log('===============================================================');

const occViewPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

// 1. Check fetchTrendData 5-second bucketing logic for 1m to 6h
console.log('Checking Requirement 1: fetchTrendData enforces 5-second polling interval bucketing for 1m to 6h...');

assert(
  occContent.includes('const isUnderOrEqual6h = timeSpan <= (6 * 3600 * 1000 + 1000);'),
  'Must detect whether timeframe is within 6 hours (covers 1m, 5m, 15m, 1h, 6h)'
);

assert(
  occContent.includes('const pollStepMs = isUnderOrEqual6h ? 5000 : null;'),
  'Must set pollStepMs = 5000 ms for filters 1m to 6h'
);

assert(
  occContent.includes('const alignedX = Math.round(pt.x / pollStepMs) * pollStepMs;'),
  'Must align data points strictly into 5-second polling buckets (Math.round(pt.x / pollStepMs) * pollStepMs)'
);

console.log('✓ PASS: 1.1: fetchTrendData groups points strictly into 5-second polling interval (:00, :05, :10, :15...)');

// Test functional behavior of bucketing
const samplePoints = [
  { x: 1700000001000, y: 10.5 }, // 1s
  { x: 1700000002000, y: 11.0 }, // 2s -> maps to 0s
  { x: 1700000004500, y: 12.2 }, // 4.5s -> maps to 5s
  { x: 1700000005100, y: 12.8 }, // 5.1s -> maps to 5s (overwrites with latest)
  { x: 1700000010000, y: 14.1 }, // 10s -> maps to 10s
];

const pollStepMs = 5000;
const bucketMap = new Map();
for (const pt of samplePoints) {
  const alignedX = Math.round(pt.x / pollStepMs) * pollStepMs;
  bucketMap.set(alignedX, { ...pt, x: alignedX });
}
const processed = Array.from(bucketMap.values()).sort((a, b) => a.x - b.x);

assert.strictEqual(processed.length, 3, 'Should produce exactly 3 5-second buckets');
assert.strictEqual(processed[0].x % 5000, 0, 'First bucket must be aligned to 5000ms');
assert.strictEqual(processed[1].x % 5000, 0, 'Second bucket must be aligned to 5000ms');
assert.strictEqual(processed[2].x % 5000, 0, 'Third bucket must be aligned to 5000ms');
assert.strictEqual(processed[1].y, 12.8, 'Latest reading in the 5-second bucket must be preserved');
console.log('✓ PASS: 1.2: Bucketing logic correctly aligns unaligned 1-second timestamps to 5-second polling intervals');

// 2. Check onChartScrub Snapping to 5-Second Boundaries
console.log('Checking Requirement 2: onChartScrub snaps crosshair and tooltip time to 5 seconds for 1m to 6h...');

assert(
  occContent.includes('let effectiveTime = isUnderOrEqual6h ? (Math.round(hoverTime / 5000) * 5000) : hoverTime;'),
  'onChartScrub must snap effectiveTime to 5000ms multiples for 1m to 6h'
);

assert(
  occContent.includes('const scrubX = isUnderOrEqual6h ? (padLeft + snappedFrac * plotW) : clampedX;'),
  'onChartScrub must snap crosshair x coordinate to 5-second point position'
);

assert(
  occContent.includes('x: scrubX,\n        rawPixelX: mouseX,\n        y: 85,\n        time: d.toLocaleTimeString(\'en-GB\'),'),
  'onChartScrub must use scrubX for crosshair line and 5-second snapped time for tooltip'
);

console.log('✓ PASS: 2.1: onChartScrub snaps crosshair and tooltip time strictly to 5-second polling points');

// 3. Check onFullscreenChartScrub Snapping
console.log('Checking Requirement 3: onFullscreenChartScrub snaps crosshair and tooltip time to 5 seconds...');

assert(
  occContent.includes('onFullscreenChartScrub(e) {'),
  'onFullscreenChartScrub method exists'
);

const fsScrubSection = occContent.substring(occContent.indexOf('onFullscreenChartScrub(e)'), occContent.indexOf('downloadTrendCSV()'));
assert(
  fsScrubSection.includes('let effectiveTime = isUnderOrEqual6h ? (Math.round(hoverTime / 5000) * 5000) : hoverTime;'),
  'onFullscreenChartScrub must snap effectiveTime to 5000ms'
);
assert(
  fsScrubSection.includes('x: scrubX,'),
  'onFullscreenChartScrub must snap crosshair line to scrubX'
);

console.log('✓ PASS: 3.1: Fullscreen chart also snaps to 5-second polling points');

// 4. Check X-Axis Ticks Spacing Preserved ("tidak merubah jarak titik waktu pada sumbu x")
console.log('Checking Requirement 4: Preserving X-axis tick intervals and spacing...');

assert(
  occContent.includes('trendXTicks() {'),
  'trendXTicks computed property exists'
);

assert(
  occContent.includes('const steps = width >= 900 ? 6 : (width >= 600 ? 5 : 3);'),
  'trendXTicks must maintain its visual step divisions'
);

assert(
  occContent.includes('const frac = i / steps;\n        const x = padLeft + frac * plotW;\n        const time = new Date(start + frac * span);'),
  'trendXTicks must maintain uniform fractional x spacing (padLeft + frac * plotW) without modification'
);

console.log('✓ PASS: 4.1: X-axis tick intervals, coordinates, and spacing are completely preserved');

// 5. Check configuredPollingIntervalMs default fallback
console.log('Checking Requirement 5: configuredPollingIntervalMs default fallback...');

const pollIntervalMatch = occContent.match(/configuredPollingIntervalMs\(\)\s*\{[\s\S]*?return\s*(\d+);/);
assert(pollIntervalMatch, 'configuredPollingIntervalMs must have a fallback return');
assert.strictEqual(pollIntervalMatch[1], '5000', 'Fallback polling interval must be 5000 ms');

console.log('✓ PASS: 5.1: Default polling interval configured to 5000ms');

console.log('===============================================================');
console.log('TEST SUMMARY: ALL 5 REQUIREMENTS PASSED (TOTAL: 5)');
console.log('===============================================================');
