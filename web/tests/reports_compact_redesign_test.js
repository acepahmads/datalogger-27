/**
 * Automated Verification: Compact Single-Viewport Operational Reports Redesign
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

const reportsViewPath = path.join(__dirname, '../src/views/ReportsView.vue');
const reportsExportPath = path.join(__dirname, '../src/utils/reportsExport.js');
const reportServicePath = path.join(__dirname, '../../internal/service/report_service.go');
const enLocalePath = path.join(__dirname, '../src/i18n/locales/en.js');
const idLocalePath = path.join(__dirname, '../src/i18n/locales/id.js');

const viewSrc = fs.readFileSync(reportsViewPath, 'utf8');
const exportSrc = fs.readFileSync(reportsExportPath, 'utf8');
const serviceSrc = fs.readFileSync(reportServicePath, 'utf8');
const enSrc = fs.readFileSync(enLocalePath, 'utf8');
const idSrc = fs.readFileSync(idLocalePath, 'utf8');

let passed = 0;
let failed = 0;

function test(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    passed++;
  } catch (err) {
    console.error(`✗ FAIL: ${name}\n  -> ${err.message}`);
    failed++;
  }
}

test('1. Removes dev #79 badge and unifies command, tabs, filters, and complete export suite', () => {
  assert(!viewSrc.includes('#79'), 'Dev #79 badge must be removed from ReportsView.vue');
  assert(viewSrc.includes('id="btn-print-report"'), 'Missing Print button');
  assert(viewSrc.includes('id="btn-export-pdf"'), 'Missing PDF export button');
  assert(viewSrc.includes('id="btn-export-xls"'), 'Missing Excel .XLS export button');
  assert(viewSrc.includes('id="btn-export-csv"'), 'Missing CSV export button');
  assert(viewSrc.includes('id="btn-download-html"'), 'Missing Download HTML button');
  assert(viewSrc.includes('id="btn-fetch-report"'), 'Missing Compile/Refresh button');
  assert(viewSrc.includes('v-model.trim="searchQuery"'), 'Missing inline quick search filter');
});

test('2. Implements 4 segmented report focus tabs & asymmetric 8/4 Split Bento layout', () => {
  assert(viewSrc.includes("value: 'COMPREHENSIVE'"), 'Missing COMPREHENSIVE tab');
  assert(viewSrc.includes("value: 'TELEMETRY'"), 'Missing TELEMETRY tab');
  assert(viewSrc.includes("value: 'COMMUNICATION'"), 'Missing COMMUNICATION tab');
  assert(viewSrc.includes("value: 'ALARMS'"), 'Missing ALARMS tab');
  assert(viewSrc.includes("lg:col-span-8"), 'Missing 65% (8/12) left column in Comprehensive Bento');
  assert(viewSrc.includes("lg:col-span-4"), 'Missing 35% (4/12) right audit stack in Comprehensive Bento');
});

test('3. Renders Visual Range Bar (Min-Avg-Max) and First -> Last (Delta) metrics', () => {
  assert(viewSrc.includes('computeVisualRangeStyle(p)'), 'Missing computeVisualRangeStyle');
  assert(viewSrc.includes('hasFirstLast(p)'), 'Missing hasFirstLast helper');
  assert(viewSrc.includes('formatDelta(p)'), 'Missing formatDelta helper');
  assert(!viewSrc.includes('this.reportType.toLowerCase()'), 'Undefined this.reportType bug must be fixed');
  assert(viewSrc.includes("String(this.filters.report_type || 'comprehensive').toLowerCase()"), 'Export filename should use filters.report_type');
});

test('4. Enforces Vue 2 scoped dark mode without :global()', () => {
  assert(viewSrc.includes(":class=\"{ 'reports-dark': isDarkMode }\""), 'Missing reports-dark root class binding');
  assert(!viewSrc.includes(':global('), 'Vue 2 scoped CSS must not contain :global()');
  assert(viewSrc.includes('.reports-dark .reports-card'), 'Missing dark mode card styling');
});

test('5. Backend ReportService populates FirstValue and LastValue boundaries', () => {
  assert(serviceSrc.includes('MIN(id) as first_id, MAX(id) as last_id'), 'ReportService should select first_id and last_id');
  assert(serviceSrc.includes('sum.FirstValue ='), 'ReportService should populate FirstValue');
  assert(serviceSrc.includes('sum.LastValue ='), 'ReportService should populate LastValue');
});

test('6. Multi-sheet Excel (.xlsx) and Landscape PDF export helpers exist', () => {
  assert(exportSrc.includes('export function exportOperationalReportXLS'), 'Missing exportOperationalReportXLS');
  assert(exportSrc.includes('export function exportOperationalReportPDF'), 'Missing exportOperationalReportPDF');
  assert(exportSrc.includes('Ringkasan_Eksekutif'), 'Missing Executive Summary sheet');
  assert(exportSrc.includes('Statistik_Parameter'), 'Missing Parameter Telemetry sheet');
  assert(exportSrc.includes('Audit_Komunikasi_Fleet'), 'Missing Fleet Audit sheet');
  assert(exportSrc.includes('Log_Alarm_Operasional'), 'Missing Alarms Audit sheet');
});

test('7. Bilingual locale keys exist in both en.js and id.js', () => {
  const keys = ['exportXls', 'exportPdf', 'tabComprehensive', 'tabTelemetry', 'tabComm', 'tabAlarms', 'visualRange', 'firstLastDelta'];
  for (const k of keys) {
    assert(enSrc.includes(`"${k}"`), `Missing reports.${k} in en.js`);
    assert(idSrc.includes(`"${k}"`), `Missing reports.${k} in id.js`);
  }
});

console.log(`\nReports Compact Redesign Test Summary: ${passed} passed, ${failed} failed`);
if (failed > 0) process.exit(1);

