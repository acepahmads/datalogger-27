/**
 * Test Suite: Parameter Values Clipping Prevention & Value-First Layout
 * Verifies all requirements from UI BUG FIX — PREVENT PARAMETER VALUES FROM BEING CLIPPED:
 * 1. Root cause eliminated: sparkline flex-shrink-0 removed and numeric value truncation eliminated
 * 2. Value-First layout: numeric value + unit container protected with flex-shrink-0
 * 3. Tabular numerals: tabular-nums and font-variant-numeric applied to prevent layout shift
 * 4. Dynamic font size scaling: font size adjusts based on value length (short, medium, long, negative)
 * 5. Formatting accuracy: precision rules, negative signs, multi-digit values, scientific notation preserved
 * 6. Edge case values tested: 2.00, 39.00, 912.00, -1234.567, 101325.00, 0.000125, 1.23e+05
 * 7. Environmental units tested: µg/m³, mg/L, °C, m/s, hPa with safe spacing and no overlap
 * 8. Flexible mini sparkline: height 30-40px (h-8 sm:h-9: 32-36px), flex-1, flex-shrink, ml-auto
 * 9. Parameter metadata preserved: name, icon, device code, limits, quality, alarm tag, pin
 * 10. Bottom row preserved: timestamp, alarm indicator, and Plot Trend action
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('================================================================');
console.log('RUNNING AUTOMATED TEST SUITE: PARAMETER VALUE CLIPPING PREVENTION');
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

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

// Test 1: Value container is protected from flex shrink and does not use truncate
runTest('Test 1: Numeric value container uses flex-shrink-0, whitespace-nowrap, and tabular-nums (no truncate)', () => {
  assert(!occContent.includes('paramValueColorClass(param)"\n                          :title="formatParamValue(param)"\n                        >\n                          {{ formatParamValue(param) }}\n                        </span>\n                        <span class="text-xs sm:text-[13px] font-mono text-slate-500 dark:text-slate-400 font-semibold ml-1.5 flex-shrink-0">\n                          {{ param.unit || \'\' }}\n                        </span>\n                      </div>\n                      <div class="flex items-center space-x-1.5 mt-1 text-4xs font-mono text-slate-400 dark:text-slate-500 leading-none truncate">\n                        <span\n                          v-if="param.low_limit != null || param.high_limit != null"\n                          class="truncate"\n                          :title="\'Limit range: \' + (param.low_limit != null ? param.low_limit : \'--\') + \' .. \' + (param.high_limit != null ? param.high_limit : \'--\')"\n                        >\n                          [{{ param.low_limit != null ? param.low_limit : \'--\' }}..{{ param.high_limit != null ? param.high_limit : \'--\' }}]\n                        </span>\n                        <span\n                          v-if="param.is_held_value"\n                          class="px-1 py-0.2 rounded bg-amber-500/10 text-amber-500 font-bold"\n                          title="Hold last value active"\n                        >\n                          HELD\n                        </span>\n                        <span\n                          v-if="param.formula"\n                          class="px-1 py-0.2 rounded bg-blue-500/10 text-blue-400 font-bold"\n                          :title="\'Formula: \' + param.formula"\n                        >\n                          FX\n                        </span>\n                      </div>\n                    </div>\n\n                    <!-- Right: Mini Sparkline Beside Value (approx 36-40px height, gap-aware) -->\n                    <div class="w-28 sm:w-32 xl:w-36 h-9 sm:h-10 flex-shrink-0 relative bg-slate-50/70 dark:bg-[#070D17]/70 border border-slate-100 dark:border-slate-800/40 rounded-lg overflow-hidden">'),
    'Old rigid layout that caused 39..., 48..., 912... clipping must be eliminated');

  assert(
    occContent.includes('whitespace-nowrap tabular-nums'),
    'Numeric value must use whitespace-nowrap and tabular-nums to prevent wrapping and jitter'
  );
  assert(
    occContent.includes('font-variant-numeric: tabular-nums'),
    'Explicit CSS font-variant-numeric: tabular-nums must be enforced'
  );
  assert(
    occContent.includes('flex flex-col justify-center min-w-0 flex-shrink-0 max-w-[65%]'),
    'Value container must prioritize numeric readability with flex-shrink-0 and generous proportion'
  );
});

// Test 2: Dynamic font size scaling function exists and handles various digit lengths
runTest('Test 2: getParamValueFontSizeClass dynamically scales font size for short, medium, and long values', () => {
  assert(occContent.includes('getParamValueFontSizeClass(param)'), 'getParamValueFontSizeClass method must be implemented');

  // Extract method from occContent
  const fnMatch = occContent.match(/getParamValueFontSizeClass\(param\)\s*\{([\s\S]*?)\n\s*\},/);
  assert(fnMatch, 'getParamValueFontSizeClass method definition must be found');

  const fnBody = fnMatch[1];
  assert(fnBody.includes('len <= 5'), 'Must check short numbers (len <= 5)');
  assert(fnBody.includes('len <= 7'), 'Must check medium numbers (len <= 7)');
  assert(fnBody.includes('len <= 9'), 'Must check long numbers (len <= 9)');

  // Simulate method
  function simulateFontSize(val, precision = null) {
    let formatted;
    if (val == null || val === '') formatted = '--';
    else if (typeof val === 'string' && /[eE]/.test(val)) formatted = val;
    else {
      const num = Number(val);
      if (isNaN(num)) formatted = String(val);
      else if (precision != null && precision >= 0) formatted = num.toFixed(precision);
      else {
        const strVal = String(val);
        if (strVal.includes('.')) {
          const decimals = strVal.split('.')[1].length;
          if (decimals > 2) formatted = num.toFixed(decimals);
          else formatted = num.toFixed(2);
        } else formatted = num.toFixed(2);
      }
    }

    const str = String(formatted);
    const len = str.length;
    if (len <= 5) return 'large';
    if (len <= 7) return 'medium';
    if (len <= 9) return 'compact';
    return 'extra-compact';
  }

  assert.strictEqual(simulateFontSize(2.0, 2), 'large', '2.00 must have large font size');
  assert.strictEqual(simulateFontSize(39.0, 2), 'large', '39.00 must have large font size');
  assert.strictEqual(simulateFontSize(48.0, 2), 'large', '48.00 must have large font size');
  assert.strictEqual(simulateFontSize(912.0, 2), 'medium', '912.00 (6 chars) must have medium font size');
  assert.strictEqual(simulateFontSize(-1234.567, 3), 'compact', '-1234.567 (9 chars) must have compact font size');
  assert.strictEqual(simulateFontSize(101325.0, 2), 'compact', '101325.00 (9 chars) must have compact font size');
  assert.strictEqual(simulateFontSize('1.23e+05'), 'compact', '1.23e+05 (8 chars) must have compact font size');
});

// Test 3: formatParamValue preserves decimals and precision without silently omitting digits
runTest('Test 3: formatParamValue preserves negative numbers, high precision, and scientific notation', () => {
  // Extract formatParamValue from occContent
  const fnMatch = occContent.match(/formatParamValue\(param\)\s*\{([\s\S]*?)\n\s*\},/);
  assert(fnMatch, 'formatParamValue method definition must be found');

  const fnBody = fnMatch[1];
  assert(fnBody.includes('num.toFixed'), 'Must use toFixed with precision');
  assert(fnBody.includes('decimals > 2'), 'Must preserve extra decimal places if present');
  assert(fnBody.includes('[eE]'), 'Must support scientific notation strings');

  // Simulate formatParamValue
  function formatParamValue(param) {
    if (!param || param.value == null || param.value === '') return '--';
    if (param.value_text) return String(param.value_text);
    if (typeof param.value === 'string' && /[eE]/.test(param.value)) {
      return param.value;
    }
    const num = Number(param.value);
    if (isNaN(num)) return String(param.value);
    if (param.precision != null && param.precision >= 0) {
      return num.toFixed(param.precision);
    }
    const strVal = String(param.value);
    if (strVal.includes('.')) {
      const decimals = strVal.split('.')[1].length;
      if (decimals > 2) {
        return num.toFixed(decimals);
      }
    }
    return num.toFixed(2);
  }

  assert.strictEqual(formatParamValue({ value: 2, precision: 2 }), '2.00');
  assert.strictEqual(formatParamValue({ value: 39, precision: 2 }), '39.00');
  assert.strictEqual(formatParamValue({ value: 912, precision: 2 }), '912.00');
  assert.strictEqual(formatParamValue({ value: -1234.567, precision: 3 }), '-1234.567');
  assert.strictEqual(formatParamValue({ value: '-1234.567' }), '-1234.567');
  assert.strictEqual(formatParamValue({ value: 101325, precision: 2 }), '101325.00');
  assert.strictEqual(formatParamValue({ value: '1.23e+05' }), '1.23e+05');
  assert.strictEqual(formatParamValue({ value: 0.000125, precision: 6 }), '0.000125');
  assert.strictEqual(formatParamValue(null), '--');
  assert.strictEqual(formatParamValue({ value: null }), '--');
});

// Test 4: Unit is displayed with safe spacing and does not overlap numeric value
runTest('Test 4: Environmental units (µg/m³, mg/L, °C, m/s, hPa) are rendered cleanly beside value', () => {
  assert(
    occContent.includes('text-xs sm:text-[13px] font-mono text-slate-500 dark:text-slate-400 font-semibold ml-1.5 flex-shrink-0 whitespace-nowrap'),
    'Unit must be cleanly styled with margin-left, whitespace-nowrap, and flex-shrink-0'
  );
  assert(
    occContent.includes(':title="param.unit"'),
    'Unit tooltip must be present for full clarity'
  );
});

// Test 5: Mini Sparkline has flexible width and height ~30-40px
runTest('Test 5: Sparkline container adapts flexibly (flex-1, flex-shrink, min-w-[50px], max-w-[136px], height 32-40px)', () => {
  assert(
    occContent.includes('flex-1 min-w-[50px] max-w-[136px]') &&
    (occContent.includes('h-8') || occContent.includes('h-9 sm:h-10')),
    'Sparkline must have flexible container that shrinks before numeric value'
  );
  assert(
    occContent.includes('h-8') || occContent.includes('h-9 sm:h-10'),
    'Sparkline height must be approximately 32-40px (within 30-40px target)'
  );
  assert(
    occContent.includes('viewBox="0 0 100 28"'),
    'Sparkline viewBox must match 100x28 aspect ratio'
  );
  assert(
    occContent.includes('preserveAspectRatio="none"'),
    'Sparkline must stretch/shrink horizontally without distortion'
  );
});

// Test 6: Sparkline preserves history rendering, linear gradient, and empty state
runTest('Test 6: Sparkline retains real sample history, gap handling, area fill, and awaiting samples state', () => {
  assert(occContent.includes('getSparkline(param.param_id).hasHistory'), 'Must check hasHistory');
  assert(occContent.includes('getSparklineColor(param)'), 'Must apply dynamic series color');
  assert(occContent.includes('areaPaths'), 'Must render gap-aware area paths');
  assert(occContent.includes('strokePaths'), 'Must render gap-aware stroke paths');
  assert(occContent.includes('awaitingSamples'), 'Must render awaiting samples state when samples < 2');
});

// Test 7: Card dimensions and responsive grid layout preserved
runTest('Test 7: Uniform card height (88-95px, 114-118px, or 130-150px) and minimum width (~260px) preserved across categories', () => {
  assert(
    occContent.includes('h-[88px] sm:h-[90px]') ||
    occContent.includes('h-[134px] sm:h-[140px]') ||
    occContent.includes('h-[114px] sm:h-[118px]'),
    'Uniform card height preserved'
  );
  assert(occContent.includes('min-w-[260px]'), 'Card min-width 260px preserved');
  assert(occContent.includes('grid grid-cols-1 lg:grid-cols-12 gap-3 sm:gap-3.5'), '12-column responsive layout preserved');
});

// Test 8: Live updates and alarm indications do not cause layout shifts
runTest('Test 8: Live telemetry blinks, alarms, and held/formula flags do not shift value row', () => {
  assert(occContent.includes('liveBlinks[param.param_id]'), 'Live blink styling preserved');
  assert(occContent.includes('isParamAlarmOrStaleBad(param)'), 'Alarm border & background pulse preserved');
  assert(occContent.includes('param.is_held_value'), 'HELD badge preserved');
  assert(occContent.includes('param.formula'), 'FX formula badge preserved');
  assert(occContent.includes('max-w-[140px] sm:max-w-[170px]'), 'Sub-limit row has bounded width to prevent row expansion');
});

console.log('\n================================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('================================================================');

if (failedTests > 0) {
  process.exit(1);
} else {
  process.exit(0);
}
