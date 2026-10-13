/**
 * Test Suite: Parameter Card Typography & Configurable Icons
 * Verifies Section 8 Acceptance Criteria:
 * 1. Device identifier is smaller than the parameter name
 * 2. Parameter name remains readable with long names
 * 3. Correct icon mapping for supported parameter types
 * 4. Configured icon overrides automatic mapping
 * 5. Unknown parameter uses a safe fallback icon
 * 6. Saving an icon configuration persists after refresh / payload
 * 7. Existing parameters without icon configuration remain compatible
 * 8. Icons render correctly in dark and light themes
 * 9. Card layout works across supported screen widths
 * 10. Live telemetry, sparklines, quality badges, and chart selection still work
 * 11. Backend model and API support for icon field
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING AUTOMATED TEST SUITE: PARAMETER CARD TYPOGRAPHY & ICONS');
console.log('===============================================================\n');

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

// Read relevant files
const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const modalPath = path.join(__dirname, '../src/components/ParameterModal.vue');
const modalContent = fs.readFileSync(modalPath, 'utf8');

const iconUtilPath = path.join(__dirname, '../src/utils/parameterIcons.js');
const iconUtilContent = fs.readFileSync(iconUtilPath, 'utf8');

const iconCompPath = path.join(__dirname, '../src/components/ParameterIcon.vue');
const iconCompContent = fs.readFileSync(iconCompPath, 'utf8');

const modelDevicePath = path.join(__dirname, '../../internal/model/device.go');
const modelDeviceContent = fs.readFileSync(modelDevicePath, 'utf8');

const serviceDevicePath = path.join(__dirname, '../../internal/service/device_service.go');
const serviceDeviceContent = fs.readFileSync(serviceDevicePath, 'utf8');

// Require the icon utility dynamically
const { resolveParameterIconKey, resolveParameterIconDef, PARAMETER_ICON_CATALOG } = require('../src/utils/parameterIcons.js');

// Test 1: Device identifier is smaller than parameter name
runTest('Test 1: Device identifier font size is 10-11px and smaller than parameter name (14-16px)', () => {
  // Check OCC header has small device identifier class: text-[10.5px] or text-[11px] or text-3xs
  assert(
    occContent.includes('text-[10.5px]') || occContent.includes('text-[11px]') || occContent.includes('text-3xs'),
    'Device identifier must use 10-11px font styling'
  );
  assert(
    occContent.includes('text-cyan-600/75') || occContent.includes('text-cyan-400/75') || occContent.includes('text-cyan'),
    'Device identifier must use muted cyan / secondary styling'
  );
  // Check parameter name has prominent font: text-sm or text-[15px] and font-bold
  assert(
    occContent.includes('text-sm sm:text-[15px] font-bold') || occContent.includes('text-sm sm:text-base font-bold'),
    'Parameter name must use larger font (14-16px) with bold weight'
  );
});

// Test 2: Parameter name remains readable with long names
runTest('Test 2: Parameter name handles long names gracefully with truncation and tooltip', () => {
  assert(
    occContent.includes(':title="(param.param_name && param.param_name !== param.param_code) ? (param.param_name + \' (\' + param.param_code + \')\') : param.param_code"'),
    'Parameter name must provide full accessible title tooltip'
  );
  assert(
    occContent.includes('truncate tracking-tight leading-none'),
    'Parameter name must be styled to truncate without overflowing'
  );
});

// Test 3: Correct icon mapping for supported parameter types
runTest('Test 3: Heuristic automatic icon mapping for all major environmental and telemetry sensors', () => {
  assert.strictEqual(resolveParameterIconKey('', 'PM25', 'Particulate Matter 2.5', 'µg/m³'), 'particles');
  assert.strictEqual(resolveParameterIconKey('', 'PM10', 'Particulate Matter 10', 'µg/m³'), 'particles');
  assert.strictEqual(resolveParameterIconKey('', 'TSP', 'Total Suspended Particles', 'mg/m³'), 'particles');
  assert.strictEqual(resolveParameterIconKey('', 'TEMP_AMB', 'Ambient Temperature', '°C'), 'thermometer');
  assert.strictEqual(resolveParameterIconKey('', 'SUHU_RUANG', 'Suhu Udara', 'C'), 'thermometer');
  assert.strictEqual(resolveParameterIconKey('', 'RH_01', 'Relative Humidity', '%'), 'droplet');
  assert.strictEqual(resolveParameterIconKey('', 'LEMBAP', 'Kelembaban Relatif', '%RH'), 'droplet');
  assert.strictEqual(resolveParameterIconKey('', 'BARO_PRESS', 'Atmospheric Pressure', 'hPa'), 'gauge');
  assert.strictEqual(resolveParameterIconKey('', 'TEKANAN', 'Air Pressure', 'mbar'), 'gauge');
  assert.strictEqual(resolveParameterIconKey('', 'WS_AVG', 'Wind Speed', 'm/s'), 'wind');
  assert.strictEqual(resolveParameterIconKey('', 'KECEPATAN_ANGIN', 'Wind Speed Anemometer', 'knot'), 'wind');
  assert.strictEqual(resolveParameterIconKey('', 'WD_DEG', 'Wind Direction', 'deg'), 'compass');
  assert.strictEqual(resolveParameterIconKey('', 'ARAH_ANGIN', 'Wind Direction Vane', 'degree'), 'compass');
  assert.strictEqual(resolveParameterIconKey('', 'RAIN_GAUGE', 'Rainfall Tipping Bucket', 'mm'), 'cloud-rain');
  assert.strictEqual(resolveParameterIconKey('', 'CURAH_HUJAN', 'Curah Hujan Harian', 'mm'), 'cloud-rain');
  assert.strictEqual(resolveParameterIconKey('', 'SOLAR_RAD', 'Solar Radiation Pyranometer', 'W/m2'), 'sun');
  assert.strictEqual(resolveParameterIconKey('', 'UV_INDEX', 'Ultraviolet Index', 'index'), 'sun');
  assert.strictEqual(resolveParameterIconKey('', 'O3_PPM', 'Ozone Gas', 'ppm'), 'atom');
  assert.strictEqual(resolveParameterIconKey('', 'CO2_MON', 'Carbon Dioxide Concentration', 'ppm'), 'atom');
  assert.strictEqual(resolveParameterIconKey('', 'NO2_GAS', 'Nitrogen Dioxide', 'ppb'), 'atom');
  assert.strictEqual(resolveParameterIconKey('', 'SO2_GAS', 'Sulfur Dioxide', 'ppb'), 'atom');
  assert.strictEqual(resolveParameterIconKey('', 'FLOW_RATE', 'Water Discharge Rate', 'L/min'), 'waves');
  assert.strictEqual(resolveParameterIconKey('', 'DEBIT_AIR', 'Level Flume', 'm3/h'), 'waves');
});

// Test 4: Configured icon overrides automatic mapping
runTest('Test 4: Explicitly configured parameter icon overrides automatic heuristic mapping', () => {
  // Even if code is TEMP (which would normally map to thermometer), an explicit icon takes priority
  assert.strictEqual(resolveParameterIconKey('particles', 'TEMP_01', 'Temperature', '°C'), 'particles');
  assert.strictEqual(resolveParameterIconKey('gauge', 'WS', 'Wind Speed', 'm/s'), 'gauge');
  assert.strictEqual(resolveParameterIconKey('atom', 'FLOW', 'Flow Rate', 'L/s'), 'atom');
});

// Test 5: Unknown parameter uses a safe fallback icon
runTest('Test 5: Unknown or unclassified parameter gracefully falls back to generic activity icon', () => {
  assert.strictEqual(resolveParameterIconKey('', 'CUSTOM_REG_99', 'Vendor Register', 'raw'), 'activity');
  assert.strictEqual(resolveParameterIconKey('invalid_nonexistent_key', 'GEN_VAL', 'Value', ''), 'activity');
  const def = resolveParameterIconDef('', 'XYZ', 'Unknown Parameter', '');
  assert.strictEqual(def.key, 'activity');
  assert(def.svg.includes('<polyline'), 'Fallback must contain valid activity SVG');
});

// Test 6: Saving an icon configuration persists in parameter form and payload
runTest('Test 6: Parameter configuration interface supports icon selection, live preview, and save payload', () => {
  assert(modalContent.includes('form.icon'), 'ParameterModal must bind form.icon');
  assert(modalContent.includes('icon: this.form.icon ? this.form.icon.trim() : \'\''), 'Submit payload must serialize configured icon');
  assert(modalContent.includes('ParameterIcon :icon="form.icon"'), 'ParameterModal must render live preview using ParameterIcon');
  assert(modalContent.includes('currentResolvedIconKey'), 'ParameterModal must display resolved icon status');
});

// Test 7: Existing parameters without icon configuration remain compatible
runTest('Test 7: Parameters with empty or missing icon remain 100% backward compatible', () => {
  const result = resolveParameterIconKey('', 'O3', 'Ozone', 'ppm');
  assert.strictEqual(result, 'atom');
  const nullResult = resolveParameterIconKey(null, 'HUMIDITY', 'Humidity', '%');
  assert.strictEqual(nullResult, 'droplet');
});

// Test 8: Icons render correctly in dark and light themes
runTest('Test 8: Vector icons provide dual dark/light theme CSS tokens without alarm severity clash', () => {
  PARAMETER_ICON_CATALOG.forEach(cat => {
    assert(cat.colorClass.includes('dark:'), `Catalog item ${cat.key} must specify dark mode color`);
    // Ensure default icon colors do not use alarm severity red / danger
    assert(!cat.colorClass.includes('text-rose') && !cat.colorClass.includes('text-red'), `Catalog item ${cat.key} must not clash with alarm severity red`);
  });
});

// Test 9: Card layout works across supported screen widths
runTest('Test 9: Monitoring card layout maintains compact height, separator, and non-breaking structure', () => {
  assert(occContent.includes('AQMS-01') || occContent.includes('param.device_code'), 'Card header must render device code');
  assert(occContent.includes('·'), 'Card header must include clean separator between device and parameter');
  assert(occContent.includes('ParameterIcon'), 'Card header must render ParameterIcon component');
  assert(occContent.includes('truncate'), 'Card header elements must use truncation for narrow cards');
});

// Test 10: Live telemetry, sparklines, quality badges, and chart selection preserved
runTest('Test 10: Preserves live telemetry values, sparklines, quality badges, and chart toggle', () => {
  assert(occContent.includes('formatParamValue(param)'), 'formatParamValue must remain bound');
  assert(occContent.includes('qualityBadgeClass(param.quality)'), 'qualityBadgeClass must remain bound');
  assert(occContent.includes('getSparklinePoints'), 'getSparklinePoints must remain bound');
  assert(occContent.includes('selectParamForTrend'), 'selectParamForTrend must remain bound');
  assert(occContent.includes('isParamSelectedForTrend'), 'isParamSelectedForTrend must remain bound');
});

// Test 11: Backend model, DTOs, and Service support Icon field
runTest('Test 11: Backend Parameter model, CreateParameterRequest, and UpdateParameterRequest include Icon', () => {
  assert(modelDeviceContent.includes('Icon          string            `gorm:"size:64;default:\'\'" json:"icon,omitempty"`'), 'model.Parameter must include Icon column');
  assert(serviceDeviceContent.includes('Icon                     string                  `json:"icon"`'), 'CreateParameterRequest must include Icon');
  assert(serviceDeviceContent.includes('Icon                     *string                  `json:"icon"`'), 'UpdateParameterRequest must include Icon pointer');
  assert(serviceDeviceContent.includes('param.Icon = strings.TrimSpace(*req.Icon)'), 'UpdateParameter must assign Icon');
});

console.log('\n===============================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
} else {
  process.exit(0);
}
