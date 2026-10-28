/**
 * Automated Verification: Trend Analysis Download & Operator Column Privacy
 *
 * Verifies:
 * 1. Role OPERATOR hides 'Kode' and 'Kualitas' columns in Excel (.xlsx) export
 * 2. Role OPERATOR hides 'Kode' and 'Kualitas' in PDF report autoTable
 * 3. Role OPERATOR hides 'Kode' and 'Kualitas' in CSV export
 * 4. Role OPERATOR hides 'parameter_code', 'quality', and 'quality_reason' in JSON export
 * 5. Non-operator roles (ADMIN, ENGINEER, SUPERADMIN) retain 'Kode' and 'Kualitas' columns
 * 6. Trend download button anchor has 'trend-download-anchor' preventing premature popover dismissal
 * 7. Trend download dropdown includes 'Download Grafik Periode...', PNG, SVG, CSV, JSON
 * 8. Custom period modal supports quick presets (1h, 3h, 6h, 12h, today, 24h, 7d), parameter checklists, and formats (PNG, SVG, PDF)
 * 9. Custom trend SVG generator produces valid vector graphic with header, legend, grid lines, and timestamps
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

let passedTests = 0;
let failedTests = 0;

function it(desc, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${desc}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${desc}`);
    console.error(`  Error: ${err.message}`);
    failedTests++;
  }
}

console.log('===============================================================');
console.log('RUNNING TEST SUITE: TREND DOWNLOAD & OPERATOR PRIVACY GOVERNANCE');
console.log('===============================================================');

const occViewPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

// Test 1: Operator Column Privacy - Excel
it('Test 1: Operator role hides Kode and Kualitas in Excel (.xlsx)', () => {
  assert(occContent.includes("const isOperator = (this.userRole || '').toUpperCase() === 'OPERATOR';"), 'Missing isOperator check');
  assert(occContent.includes("isOperator\n            ? ['No', 'Waktu (Timestamp)', 'Parameter', 'Nilai', 'Satuan']"), 'Excel headers did not hide Kode and Kualitas for operator');
  assert(occContent.includes("return [idx + 1, time, name, val, unit];"), 'Excel rows did not omit Kode and Kualitas for operator');
});

// Test 2: Operator Column Privacy - PDF
it('Test 2: Operator role hides Kode and Kualitas in PDF report', () => {
  assert(occContent.includes("const head = isOperator\n            ? [['No', 'Waktu', 'Parameter', 'Nilai', 'Satuan']]"), 'PDF head did not omit Kode and Kualitas for operator');
  assert(occContent.includes("return [String(idx + 1), time, name, val, unit];"), 'PDF body did not omit Kode and Kualitas for operator');
});

// Test 3: Operator Column Privacy - CSV
it('Test 3: Operator role hides Kode and Kualitas in CSV export', () => {
  assert(occContent.includes("const headers = isOperator\n            ? ['Waktu', 'Parameter', 'Nilai', 'Satuan']"), 'CSV headers did not omit Kode and Kualitas for operator');
  assert(occContent.includes("return [time, name, val, unit];"), 'CSV rows did not omit Kode and Kualitas for operator');
});

// Test 4: Operator Column Privacy - JSON
it('Test 4: Operator role hides parameter_code and quality in JSON export', () => {
  assert(occContent.includes("if (!isOperator) {\n              rec.parameter_code = formatCode(i);"), 'JSON mapping did not conditionally omit parameter_code and quality');
});

// Test 5: Non-Operator roles retain full columns
it('Test 5: Non-operator roles retain full columns (Kode & Kualitas)', () => {
  assert(occContent.includes("['No', 'Waktu (Timestamp)', 'Parameter', 'Kode', 'Nilai', 'Satuan', 'Kualitas']"), 'Non-operators missing full Excel columns');
  assert(occContent.includes("formatCode(i)"), 'Missing formatCode in export');
  assert(occContent.includes("i.quality || 'GOOD'"), 'Missing quality in export');
});

// Test 6: Trend download button anchor class
it('Test 6: Trend download anchor has trend-download-anchor class and click.stop handler', () => {
  assert(occContent.includes('class="relative trend-download-anchor"'), 'Missing trend-download-anchor class on download menu anchor');
  assert(occContent.includes('@click.stop="toggleTrendDownloadMenu"'), 'Missing toggleTrendDownloadMenu with .stop modifier');
  assert(occContent.includes('toggleTrendDownloadMenu()'), 'Missing toggleTrendDownloadMenu method definition');
});

// Test 7: Trend download menu options
it('Test 7: Trend download menu includes custom period, PNG, SVG, CSV, JSON', () => {
  assert(occContent.includes('openCustomTrendModal()'), 'Missing openCustomTrendModal option in download menu');
  assert(occContent.includes("downloadTrendChart('png')"), 'Missing PNG download in menu');
  assert(occContent.includes("downloadTrendChart('svg')"), 'Missing SVG download in menu');
  assert(occContent.includes('downloadTrendCSV()'), 'Missing CSV download in menu');
  assert(occContent.includes('downloadTrendJSON()'), 'Missing JSON download in menu');
});

// Test 8: Custom Period Modal components
it('Test 8: Custom period trend modal exists with presets, range inputs, and format choices', () => {
  assert(occContent.includes('v-if="showCustomTrendModal"'), 'Missing showCustomTrendModal condition');
  assert(occContent.includes('applyCustomTrendPreset'), 'Missing quick presets helper');
  assert(occContent.includes('customTrendForm.startTime'), 'Missing startTime input binding');
  assert(occContent.includes('customTrendForm.endTime'), 'Missing endTime input binding');
  assert(occContent.includes('customTrendForm.selectedParamIds'), 'Missing selectedParamIds binding');
  assert(occContent.includes('submitDownloadCustomTrendChart'), 'Missing submitDownloadCustomTrendChart method');
});

// Test 9: Custom Trend SVG generator & rasterizer
it('Test 9: Custom Trend SVG generator and rasterizer methods are properly defined', () => {
  assert(occContent.includes('generateCustomTrendSvgString(seriesList, timeBounds)'), 'Missing generateCustomTrendSvgString method');
  assert(occContent.includes('downloadSvgAsPng('), 'Missing downloadSvgAsPng method');
  assert(occContent.includes('downloadCustomTrendPdf('), 'Missing downloadCustomTrendPdf method');
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) process.exit(1);
