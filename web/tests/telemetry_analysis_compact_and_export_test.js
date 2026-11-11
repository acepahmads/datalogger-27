const assert = require('assert');
const fs = require('fs');
const path = require('path');

const viewPath = path.join(__dirname, '../src/views/TelemetryAnalysisView.vue');
const exportUtilPath = path.join(__dirname, '../src/utils/telemetryAnalysisExport.js');
const enPath = path.join(__dirname, '../src/i18n/locales/en.js');
const idPath = path.join(__dirname, '../src/i18n/locales/id.js');

const viewCode = fs.readFileSync(viewPath, 'utf8');
const exportCode = fs.readFileSync(exportUtilPath, 'utf8');
const enCode = fs.readFileSync(enPath, 'utf8');
const idCode = fs.readFileSync(idPath, 'utf8');

console.log('Running Telemetry Analysis Compact Workbench & Export Suite...');

// 1. Compact Single-Viewport Workbench Layout Checks
assert(viewCode.includes('compact-telemetry-workbench'), 'Must apply compact-telemetry-workbench root container');
assert(viewCode.includes("setWorkspaceLayout('SPLIT')"), 'Must provide Split layout toggle');
assert(viewCode.includes("setWorkspaceLayout('CHART')"), 'Must provide Chart layout toggle');
assert(viewCode.includes("setWorkspaceLayout('TABLE')"), 'Must provide Table layout toggle');
assert(viewCode.includes('availableTimePresets'), 'Must provide 1-click quick time preset pills');
assert(viewCode.includes('activeSingleSeriesStats'), 'Must provide inline monospace single-series KPI ribbon');
assert(viewCode.includes('compact-analysis-table'), 'Must apply compact sticky-header data table styling');

// 2. Data Export (XLS, PDF, CSV, JSON) & Chart Export (PNG, JPG) Wiring
assert(viewCode.includes('@click="exportXLS"'), 'Must wire XLS export button');
assert(viewCode.includes('@click="exportPDF"'), 'Must wire PDF export button');
assert(viewCode.includes('@click="exportCSV"'), 'Must preserve CSV export button');
assert(viewCode.includes('@click="exportJSON"'), 'Must preserve JSON export button');
assert(viewCode.includes("@click=\"exportChart('png')\""), 'Must wire PNG chart image export button');
assert(viewCode.includes("@click=\"exportChart('jpg')\""), 'Must wire JPG chart image export button');
assert(viewCode.includes('ref="chartSvg"'), 'Trend SVG must expose chartSvg ref for PNG/JPG/PDF serialization');

// 3. Export Utility Verification (SheetJS XLS, jsPDF + autoTable PDF, 2x Canvas PNG/JPG, RBAC)
assert(exportCode.includes("import * as XLSX from 'xlsx'"), 'Export utility must import SheetJS xlsx');
assert(exportCode.includes("import { jsPDF } from 'jspdf'"), 'Export utility must import jsPDF');
assert(exportCode.includes("import autoTable from 'jspdf-autotable'"), 'Export utility must import jspdf-autotable');
assert(exportCode.includes("Data_Telemetri") && exportCode.includes("Ringkasan_Statistik"), 'XLS export must generate Data_Telemetri and Ringkasan_Statistik sheets');
assert(exportCode.includes("'image/jpeg'") && exportCode.includes("'image/png'"), 'Chart image renderer must support both PNG and JPEG');
assert(exportCode.includes('!isOperator'), 'Export utility must strip Device and Raw columns for OPERATOR');

// 4. Operator RBAC Preservation Checks (6.3 & 6.4 Parity)
assert(/v-if="!isOperator"\s+@click="setMode\('RAW'\)"/.test(viewCode), 'Must hide Raw mode tab for OPERATOR');
assert(/v-if="!isOperator"\s+to="\/monitoring\/telemetry"/.test(viewCode), 'Must hide Live Telemetry button for OPERATOR');
assert(viewCode.includes("(this.isOperator && targetMode === 'RAW') ? 'HISTORICAL' : targetMode"), 'Must redirect ?mode=RAW to HISTORICAL for OPERATOR');
assert(viewCode.includes(':to="isOperator ? \'/monitoring/dashboard-operator\' : \'/monitoring/devices\'"'), 'Breadcrumb must route OPERATOR to /monitoring/dashboard-operator');
assert(viewCode.includes('<span v-if="!isOperator">{{ s.deviceCode }}: </span>{{ s.paramCode }}'), 'Per-series matrix must hide deviceCode for OPERATOR');
assert(viewCode.includes('<span v-if="!isOperator">{{ activePoint.deviceCode }}: </span>{{ activePoint.paramCode }}'), 'Chart tooltip must hide deviceCode for OPERATOR');
assert(viewCode.includes('this.records.map(({ device_id, device, ...rest }) => rest)'), 'exportJSON must strip device metadata for OPERATOR');

// 5. Bilingual i18n Keys
['exportXls', 'exportPdf', 'exportChartPng', 'exportChartJpg', 'layoutSplit', 'perSeriesStats'].forEach((k) => {
  assert(enCode.includes(`"${k}":`), `en.js must define analysis.${k}`);
  assert(idCode.includes(`"${k}":`), `id.js must define analysis.${k}`);
});

console.log('PASS: All Telemetry Analysis Compact Workbench & Export (XLS, PDF, PNG, JPG) checks passed!');
