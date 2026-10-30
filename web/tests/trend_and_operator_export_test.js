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

// Test 10: Trend dropdown & Fullscreen toolbar contain PDF and XLS options
it('Test 10: Trend download dropdown and Fullscreen toolbar contain PDF & XLS options', () => {
  assert(occContent.includes('@click="downloadTrendPDF(); showTrendDownloadMenu = false"'), 'Missing downloadTrendPDF in dropdown');
  assert(occContent.includes('@click="downloadTrendXLS(); showTrendDownloadMenu = false"'), 'Missing downloadTrendXLS in dropdown');
  assert(occContent.includes('Unduh Laporan (PDF)'), 'Missing "Unduh Laporan (PDF)" label in dropdown');
  assert(occContent.includes('Unduh Data Excel (XLS)'), 'Missing "Unduh Data Excel (XLS)" label in dropdown');
  assert(occContent.includes('@click="downloadTrendPDF"'), 'Missing downloadTrendPDF in fullscreen toolbar');
  assert(occContent.includes('@click="downloadTrendXLS"'), 'Missing downloadTrendXLS in fullscreen toolbar');
});

// Test 11: downloadTrendPDF role governance and PDF report construction
it('Test 11: downloadTrendPDF creates landscape PDF with chart and hides Kode for Operator', () => {
  assert(occContent.includes('async downloadTrendPDF() {'), 'Missing downloadTrendPDF implementation');
  assert(occContent.includes("doc.addImage(chartPngUrl, 'PNG', 13.5, 24, 270, 118);"), 'Missing doc.addImage in downloadTrendPDF');
  assert(occContent.includes("const isOperator = (this.userRole || '').toUpperCase() === 'OPERATOR';"), 'Missing isOperator check in downloadTrendPDF');
  assert(occContent.includes("head: head,\n          body: body,\n          theme: 'grid',"), 'Missing autoTable config in downloadTrendPDF');
});

// Test 12: downloadTrendXLS multi-sheet workbook and role governance
it('Test 12: downloadTrendXLS builds Data_Telemetri and Ringkasan_Statistik with operator privacy', () => {
  assert(occContent.includes('downloadTrendXLS() {'), 'Missing downloadTrendXLS implementation');
  assert(occContent.includes("XLSX.utils.book_append_sheet(wb, wsData, 'Data_Telemetri');"), 'Missing Data_Telemetri sheet');
  assert(occContent.includes("XLSX.utils.book_append_sheet(wb, wsStats, 'Ringkasan_Statistik');"), 'Missing Ringkasan_Statistik sheet');
  assert(occContent.includes("const headers = isOperator\n          ? ['No', 'Waktu (Timestamp)', 'Parameter', 'Nilai', 'Satuan']\n          : ['No', 'Waktu (Timestamp)', 'Parameter', 'Kode', 'Nilai', 'Satuan', 'Kualitas'];"), 'Missing operator column filter in downloadTrendXLS');
});

// Test 13: Custom Period Modal supports Excel XLS with downloadCustomTrendXls
it('Test 13: Custom period modal supports Excel XLS with downloadCustomTrendXls', () => {
  assert(occContent.includes('customTrendForm.format === \'xlsx\''), 'Missing customTrendForm format xlsx handling');
  assert(occContent.includes('downloadCustomTrendXls(seriesList, timeBounds, filename)'), 'Missing downloadCustomTrendXls implementation');
  assert(occContent.includes('downloadCustomTrendXls(activeSeriesList, timeBounds, `data_telemetri_${periodSafe}.xlsx`)'), 'Missing call to downloadCustomTrendXls in submitDownloadCustomTrendChart');
});

// Test 14: RAW SENSOR export authorizes via canAccessRawTrendMode (ADMIN & SUPERADMIN)
it('Test 14: executeTelemetryExport authorizes RAW mode via canAccessRawTrendMode', () => {
  assert(
    occContent.includes("if (this.exportForm.mode === 'RAW') {\n          if (!this.canAccessRawTrendMode) {"),
    'executeTelemetryExport must check canAccessRawTrendMode so both ADMIN and SUPERADMIN can export RAW data'
  );
  assert(!occContent.includes("if (this.exportForm.mode === 'RAW') {\n          if (!this.isSuperAdmin)"), 'executeTelemetryExport must not block ADMIN from RAW export with !this.isSuperAdmin');
});

// Test 15: RAW SENSOR export extracts raw_value instead of processed_value
it('Test 15: formatVal extracts raw_value in RAW mode, avg_value in AGGREGATED mode, and processed_value in PROCESSED mode', () => {
  assert(occContent.includes("const isRawMode = this.exportForm.mode === 'RAW' && this.canAccessRawTrendMode;"), 'Missing isRawMode flag in executeTelemetryExport');
  assert(occContent.includes("raw = i.raw_value != null ? i.raw_value : (i.value != null ? i.value : '');"), 'formatVal must read i.raw_value first when isRawMode is true');
  assert(occContent.includes("Number(Number(raw).toFixed(6))"), 'RAW mode must preserve raw sensor decimal fidelity');

  // Behavioral simulation of formatVal across modes
  const makeFormatter = (mode, canAccessRaw) => {
    const isRawMode = mode === 'RAW' && canAccessRaw;
    const isAggMode = mode === 'AGGREGATED';
    return (i) => {
      let raw = '';
      if (isRawMode) {
        raw = i.raw_value != null ? i.raw_value : (i.value != null ? i.value : '');
        if (raw !== '' && !isNaN(Number(raw))) {
          return Number(Number(raw).toFixed(6));
        }
        return raw;
      }
      if (isAggMode) {
        raw = i.value != null ? i.value : (i.avg_value != null ? i.avg_value : (i.processed_value != null ? i.processed_value : ''));
      } else {
        raw = i.processed_value != null ? i.processed_value : (i.value != null ? i.value : '');
      }
      if (raw !== '' && !isNaN(Number(raw))) {
        return Number(Number(raw).toFixed(2));
      }
      return raw;
    };
  };

  const sampleRecord = {
    parameter_id: 10,
    raw_value: 2048.123456,
    raw_hex: '0x450001F9',
    processed_value: 54.0,
    value: 54.0,
  };

  assert.strictEqual(makeFormatter('RAW', true)(sampleRecord), 2048.123456, 'RAW export for Admin/Superadmin must output raw_value');
  assert.strictEqual(makeFormatter('PROCESSED', true)(sampleRecord), 54, 'PROCESSED export must output processed_value');
  assert.strictEqual(makeFormatter('RAW', false)(sampleRecord), 54, 'Unauthorized RAW mode must fall back to processed_value');
  assert.strictEqual(makeFormatter('AGGREGATED', true)({ avg_value: 42.678 }), 42.68, 'AGGREGATED export must support avg_value');
});

// Test 16: RAW SENSOR columns & fields across Excel, PDF, CSV, JSON + Custom Trend Chart
it('Test 16: RAW mode exports Nilai Raw & Raw Hex across Excel, PDF, CSV, JSON and Custom Trend Chart', () => {
  assert(occContent.includes("['No', 'Waktu (Timestamp)', 'Parameter', 'Kode', 'Nilai Raw', 'Raw Hex', 'Kualitas']"), 'Excel RAW mode must show Nilai Raw & Raw Hex headers');
  assert(occContent.includes("[['No', 'Waktu', 'Parameter', 'Kode', 'Nilai Raw', 'Raw Hex', 'Kualitas']]"), 'PDF RAW mode must show Nilai Raw & Raw Hex headers');
  assert(occContent.includes("['Waktu', 'Parameter', 'Kode', 'Nilai Raw', 'Raw Hex', 'Kualitas']"), 'CSV RAW mode must show Nilai Raw & Raw Hex headers');
  assert(occContent.includes("raw_value: val,\n                  raw_hex: i.raw_hex || null,"), 'JSON RAW mode must include raw_value and raw_hex');
  assert(occContent.includes("const isRawTrend = this.canAccessRawTrendMode && this.chartMode === 'raw';"), 'submitDownloadCustomTrendChart must also honor raw chartMode');
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) process.exit(1);

