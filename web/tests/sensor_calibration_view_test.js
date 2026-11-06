/**
 * Automated Frontend Verification Test for Sensor Calibration & Multi-Point Regression (1-10 Points)
 * Run via: node web/tests/sensor_calibration_view_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const mathSrcPath = path.join(__dirname, '../src/utils/calibrationMath.js');
const viewSrcPath = path.join(__dirname, '../src/views/SensorCalibrationView.vue');
const sidebarPath = path.join(__dirname, '../src/components/Sidebar.vue');
const routerPath = path.join(__dirname, '../src/router/index.js');

// Load ES module calibrationMath.js by keeping local bindings + attaching to exports
const rawMathCode = fs.readFileSync(mathSrcPath, 'utf8');
const exportedNames = [];
const cjsMathCode = rawMathCode
  .replace(/export\s+const\s+(\w+)\s*=/g, (_, name) => {
    exportedNames.push(name);
    return `const ${name} =`;
  })
  .replace(/export\s+function\s+(\w+)/g, (_, name) => {
    exportedNames.push(name);
    return `function ${name}`;
  }) + '\n' + exportedNames.map(n => `exports.${n} = ${n};`).join('\n');

const calMath = {};
const fn = new Function('exports', cjsMathCode);
fn(calMath);

// 1. Verify 1-Point Offset & 1-Point Span
const resOffset = calMath.calculateRegressionJS('OFFSET_1PT', [
  { index: 1, active: true, label: 'Buffer pH 7.00', raw_x: 6.85, ref_y: 7.00 },
], 0.995);
assert.strictEqual(resOffset.valid, true, '1-Point Offset should be valid');
assert(Math.abs(resOffset.c1 - 1.0) < 1e-6 && Math.abs(resOffset.c0 - 0.15) < 1e-6, '1-Point Offset equation mismatch');
assert.strictEqual(resOffset.r2, null, '1-Point calibration R2 should be null');
assert.strictEqual(resOffset.passed_r2, true, '1-Point calibration passed_r2 should be true');

// 2. Verify 5-Point & 10-Point Linear Regression + Per-Parameter MinR2
const pts10 = Array.from({ length: 10 }, (_, i) => {
  const ref_y = i * 100;
  const raw_x = (ref_y - 2.0) / 1.025;
  return { index: i + 1, active: true, label: `Point ${i + 1}`, raw_x, ref_y };
});
const res10 = calMath.calculateRegressionJS('LINEAR', pts10, 0.999);
assert.strictEqual(res10.valid, true, '10-Point Linear should be valid');
assert.strictEqual(res10.active_count, 10, '10-Point Linear active count should be 10');
assert(Math.abs(res10.c1 - 1.025) < 1e-5 && Math.abs(res10.c0 - 2.0) < 1e-5, '10-Point Linear slope/offset mismatch');
assert(res10.r2 > 0.99999 && res10.passed_r2 === true, '10-Point Linear should pass R2 >= 0.999');

// 3. Verify Quadratic Order-2 Regression with Unchecked Outlier
const quadFn = x => 0.0015 * x * x + 0.95 * x + 0.4;
const quadPts = [
  { index: 1, active: true, raw_x: 0, ref_y: quadFn(0) },
  { index: 2, active: true, raw_x: 50, ref_y: quadFn(50) },
  { index: 3, active: false, raw_x: 100, ref_y: 888 }, // unchecked outlier
  { index: 4, active: true, raw_x: 150, ref_y: quadFn(150) },
  { index: 5, active: true, raw_x: 250, ref_y: quadFn(250) },
  { index: 6, active: true, raw_x: 400, ref_y: quadFn(400) },
];
const resQuad = calMath.calculateRegressionJS('QUADRATIC', quadPts, 0.998);
assert.strictEqual(resQuad.valid, true, 'Quadratic regression should be valid');
assert.strictEqual(resQuad.active_count, 5, 'Quadratic active count should exclude unchecked outlier');
assert(Math.abs(resQuad.c2 - 0.0015) < 1e-6 && Math.abs(resQuad.c1 - 0.95) < 1e-5, 'Quadratic coefficients mismatch');
assert.strictEqual(resQuad.passed_r2, true, 'Quadratic R2 should pass 0.998');

// 4. Verify Smart Standard Presets for AQMS, SPARING, and WQMS_ONLIMO
const presetAQMS = calMath.getDomainStandardPresets({ parameter_code: 'SO2', parameter_name: 'Sulfur Dioxide', unit: 'µg/m³' });
const presetSPARING = calMath.getDomainStandardPresets({ parameter_code: 'COD', parameter_name: 'COD', unit: 'mg/L' });
const presetONLIMO = calMath.getDomainStandardPresets({ parameter_code: 'DO', parameter_name: 'Dissolved Oxygen', unit: 'mg/L' });
assert(presetAQMS.length >= 2 && presetAQMS[0].refs.length === 5, 'AQMS SO2 preset should include 5-point gas dilution');
assert(presetSPARING.length >= 2 && presetSPARING[0].refs.length === 5, 'SPARING COD preset should include 5-point KHP CRM');
assert(presetONLIMO.length >= 1 && presetONLIMO[0].refs.length === 2, 'WQMS ONLIMO DO preset should include Zero & Air-Saturated points');

// 5. Verify SensorCalibrationView.vue Light-Mode `text-white` compliance, global domain inheritance, & no device/domain selector clutter
const viewCode = fs.readFileSync(viewSrcPath, 'utf8');
assert(viewCode.includes('draftMinR2'), 'SensorCalibrationView must include per-parameter draftMinR2 state');
assert(viewCode.includes('globalMonitoringDomain'), 'SensorCalibrationView must inherit globalMonitoringDomain from global config');
assert(!viewCode.includes('select-calibration-device'), 'SensorCalibrationView must NOT show a device selector dropdown');
assert(!viewCode.includes('onSelectDomainFilter'), 'SensorCalibrationView must NOT show local domain filter pills');
assert(!viewCode.includes('saveParameterDomainOnly'), 'SensorCalibrationView must NOT show a per-parameter domain selector');

// Check every class attribute containing `text-white` (without `dark:text-white`) has `btn-text` or `badge-text`
const classRegex = /class="([^"]*)"/g;
let m;
while ((m = classRegex.exec(viewCode)) !== null) {
  const cls = m[1];
  const tokens = cls.split(/\s+/);
  if (tokens.includes('text-white')) {
    assert(
      tokens.includes('btn-text') || tokens.includes('badge-text'),
      `Light Mode CSS violation in SensorCalibrationView.vue: class="${cls}" uses text-white without btn-text or badge-text`
    );
  }
}

// 6. Verify Sidebar & Router RBAC (Accessible to all roles EXCEPT Operator)
const sidebarCode = fs.readFileSync(sidebarPath, 'utf8');
assert(sidebarCode.includes('/monitoring/calibration'), 'Sidebar must link to /monitoring/calibration');
assert(sidebarCode.includes('!isOperator'), 'Sidebar calibration link must be hidden for OPERATOR');

const routerCode = fs.readFileSync(routerPath, 'utf8');
assert(routerCode.includes("path: '/monitoring/calibration'"), 'Router must register /monitoring/calibration');
assert(routerCode.includes('disallowOperator: true'), 'Router must enforce disallowOperator: true for /monitoring/calibration');

console.log('PASS: All Sensor Calibration (1-10 Points, AQMS/SPARING/WQMS ONLIMO, Per-Parameter Min R2, RBAC & Light Mode) checks succeeded.');
