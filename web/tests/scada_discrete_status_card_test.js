/**
 * Automated Verification Suite: SCADA Discrete Status Card & Flexible Status Labels
 *
 * Verifies:
 * 1. Parameters in Category 'Status' (even FLOAT32/UINT16) or BOOL/COIL are detected as discrete SCADA status cards.
 * 2. Analog environmental parameters (Gas, PM, Weather, Water) remain analog numeric cards.
 * 3. Value >= 1 renders active emerald LED badge ('NYALA' in ID / 'ON' in EN) and Value 0 renders neutral slate standby ('MATI' / 'OFF') — NOT alarm red.
 * 4. Alarm severity (CRITICAL / WARNING) overrides standby badge with ISA-101 rose/amber alert styling.
 * 5. Offline / BAD communication renders '--' safely without false 'MATI' or 'NYALA'.
 * 6. Equipment-aware auto labels (Valve -> TERBUKA/TERTUTUP, PLN/UPS -> PLN NORMAL/UPS BACKUP) + custom label override (status_on_label / status_off_label).
 * 7. Horizontal State Timeline Strip builds tall emerald blocks for ON runs and low slate blocks for OFF runs, and flatlines on fault/alarm.
 * 8. Contiguous state duration (e.g. '14m', '2j 5m') is computed accurately from sparklinesMap history.
 * 9. ParameterModal & Go backend persist status_on_label and status_off_label.
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const modalPath = path.join(__dirname, '../src/components/ParameterModal.vue');
const modalContent = fs.readFileSync(modalPath, 'utf8');

const goModelPath = path.join(__dirname, '../../internal/model/device.go');
const goModelContent = fs.readFileSync(goModelPath, 'utf8');

const goServicePath = path.join(__dirname, '../../internal/service/device_service.go');
const goServiceContent = fs.readFileSync(goServicePath, 'utf8');

let passed = 0;
let failed = 0;

function runTest(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    passed++;
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    failed++;
  }
}

function extractMethod(methodName) {
  const regex = new RegExp(`\\n\\s{4}${methodName}\\s*\\(([^)]*)\\)\\s*\\{([\\s\\S]*?)\\n\\s{4}\\},`);
  const match = occContent.match(regex);
  if (!match) {
    throw new Error(`Could not extract method ${methodName} from OperatorCommandCenterView.vue`);
  }
  const args = match[1].split(',').map(a => a.trim()).filter(Boolean);
  return new Function(...args, match[2]);
}

function createHarness(locale = 'id') {
  const harness = {
    $i18n: { locale },
    isDark: true,
    perfSettings: { parameterSparklinesEnabled: true },
    categoryMap: {
      1: { id: 1, code: 'gas_pollutants', name: 'Gas & Pollutants' },
      9: { id: 9, code: 'status', name: 'Status' },
    },
    sparklinesMap: {},
    isParamInAlarmOrFault(param) {
      if (!param) return false;
      if (param.device_status === 'OFFLINE' || param.quality === 'BAD' || param.quality === 'STALE') return true;
      if (param.alarm_status === 'CRITICAL' || param.alarm_status === 'WARNING') return true;
      return false;
    },
    getSparklineColor(param) {
      if (param.alarm_status === 'CRITICAL' || param.device_status === 'OFFLINE' || param.quality === 'BAD') return '#F43F5E';
      if (param.alarm_status === 'WARNING') return '#F59E0B';
      return '#22D3EE';
    },
  };

  harness.getParamCategory = extractMethod('getParamCategory').bind(harness);
  harness.formatParamValue = extractMethod('formatParamValue').bind(harness);
  harness.isDiscreteStatusParam = extractMethod('isDiscreteStatusParam').bind(harness);
  harness.resolveDiscreteStatusLabels = extractMethod('resolveDiscreteStatusLabels').bind(harness);
  harness.parseDiscreteBoolState = extractMethod('parseDiscreteBoolState').bind(harness);
  harness.formatStatusDurationCompact = extractMethod('formatStatusDurationCompact').bind(harness);
  harness.getDiscreteStatusInfo = extractMethod('getDiscreteStatusInfo').bind(harness);
  harness.getStatusTimelineBlocks = extractMethod('getStatusTimelineBlocks').bind(harness);

  return harness;
}

console.log('====================================================================');
console.log('RUNNING TEST SUITE: SCADA DISCRETE STATUS CARD & TIMELINE STRIP');
console.log('====================================================================\n');

runTest('1. Detects Status category (including FLOAT32 register) and BOOL/COIL while keeping analog sensors normal', () => {
  const h = createHarness('id');
  assert.strictEqual(h.isDiscreteStatusParam({ param_id: 1, param_code: 'pump on', data_type: 'FLOAT32', category_id: 9 }), true);
  assert.strictEqual(h.isDiscreteStatusParam({ param_id: 2, param_code: 'relay_1', data_type: 'BOOL', category_id: null }), true);
  assert.strictEqual(h.isDiscreteStatusParam({ param_id: 3, param_code: 'coil_01', data_type: 'UINT16', register_type: 'COIL', category_id: null }), true);
  assert.strictEqual(h.isDiscreteStatusParam({ param_id: 4, param_code: 'PM25', data_type: 'FLOAT32', category_id: 1 }), false);
});

runTest('2. Renders emerald NYALA for 1 and neutral slate MATI (not red) for 0 in Indonesian & ON/OFF in English', () => {
  const hId = createHarness('id');
  const pumpOn = { param_id: 10, param_code: 'pump on', param_name: 'PUMP ON', category_id: 9, value: 1.0, device_status: 'ONLINE', quality: 'GOOD', alarm_status: 'NORMAL' };
  const infoOn = hId.getDiscreteStatusInfo(pumpOn);
  assert.strictEqual(infoOn.valid, true);
  assert.strictEqual(infoOn.isOn, true);
  assert.strictEqual(infoOn.label, 'NYALA');
  assert(infoOn.pillClass.includes('emerald'), 'Active ON pill must be emerald');

  const pumpOff = { ...pumpOn, value: 0.0 };
  const infoOff = hId.getDiscreteStatusInfo(pumpOff);
  assert.strictEqual(infoOff.isOn, false);
  assert.strictEqual(infoOff.label, 'MATI');
  assert(infoOff.pillClass.includes('slate'), 'Standby OFF pill must be neutral slate');
  assert(!infoOff.pillClass.includes('rose') && !infoOff.pillClass.includes('red'), 'Normal standby OFF must NOT be alarm red');

  const hEn = createHarness('en');
  assert.strictEqual(hEn.getDiscreteStatusInfo(pumpOn).label, 'ON');
  assert.strictEqual(hEn.getDiscreteStatusInfo(pumpOff).label, 'OFF');
});

runTest('3. Alarm severity (CRITICAL / WARNING) overrides standby colors & offline/BAD displays safe "--"', () => {
  const h = createHarness('id');
  const faultPump = { param_id: 11, param_code: 'pump_1', category_id: 9, value: 0, device_status: 'ONLINE', quality: 'GOOD', alarm_status: 'CRITICAL' };
  const critInfo = h.getDiscreteStatusInfo(faultPump);
  assert.strictEqual(critInfo.label, 'MATI');
  assert(critInfo.pillClass.includes('rose'), 'CRITICAL alarm on pump must highlight rose');

  const offlinePump = { param_id: 12, param_code: 'pump_1', category_id: 9, value: 1, device_status: 'OFFLINE', quality: 'BAD', alarm_status: 'NORMAL' };
  const offInfo = h.getDiscreteStatusInfo(offlinePump);
  assert.strictEqual(offInfo.valid, false);
  assert.strictEqual(offInfo.label, '--');
});

runTest('4. Smart equipment heuristics (Valve, Door, PLN/UPS) and custom status_on_label / status_off_label overrides', () => {
  const h = createHarness('id');
  const valve = { param_id: 20, param_code: 'valve_inlet', param_name: 'Katup Inlet', category_id: 9, value: 1, device_status: 'ONLINE', quality: 'GOOD' };
  assert.strictEqual(h.getDiscreteStatusInfo(valve).label, 'TERBUKA');
  assert.strictEqual(h.getDiscreteStatusInfo({ ...valve, value: 0 }).label, 'TERTUTUP');

  const pln = { param_id: 21, param_code: 'pln_status', param_name: 'Power PLN', category_id: 9, value: 0, device_status: 'ONLINE', quality: 'GOOD' };
  assert.strictEqual(h.getDiscreteStatusInfo(pln).label, 'UPS BACKUP');

  const custom = {
    param_id: 22,
    param_code: 'genset_mode',
    category_id: 9,
    status_on_label: 'AUTO MODE',
    status_off_label: 'MANUAL MODE',
    value: 1,
    device_status: 'ONLINE',
    quality: 'GOOD',
  };
  assert.strictEqual(h.getDiscreteStatusInfo(custom).label, 'AUTO MODE');
  assert.strictEqual(h.getDiscreteStatusInfo({ ...custom, value: 0 }).label, 'MANUAL MODE');
});

runTest('5. State Timeline Strip merges contiguous ON/OFF runs with distinct height & color, and flatlines on fault', () => {
  const h = createHarness('id');
  const now = Date.now();
  h.sparklinesMap[30] = [
    { time: now - 25 * 60000, val: 0 },
    { time: now - 20 * 60000, val: 0 },
    { time: now - 15 * 60000, val: 1 },
    { time: now - 10 * 60000, val: 1 },
    { time: now - 5 * 60000, val: 1 },
  ];
  const pump = { param_id: 30, param_code: 'pump on', category_id: 9, value: 1, device_status: 'ONLINE', quality: 'GOOD', alarm_status: 'NORMAL' };
  const blocks = h.getStatusTimelineBlocks(pump);
  assert.strictEqual(blocks.length, 2, 'Should merge 2 OFF + 3 ON samples into 2 timeline runs');
  assert.strictEqual(blocks[0].h, 8.0, 'OFF run should be shorter height (8)');
  assert.strictEqual(blocks[1].h, 15.0, 'ON run should be taller height (15)');
  assert.strictEqual(blocks[1].fill, '#10B981', 'ON run should be emerald');

  const info = h.getDiscreteStatusInfo(pump);
  assert.strictEqual(info.durationText, '15m', 'Contiguous ON duration should reflect earliest ON transition (15m)');

  // Fault flatline check
  const faultBlocks = h.getStatusTimelineBlocks({ ...pump, device_status: 'OFFLINE', quality: 'BAD' });
  assert.strictEqual(faultBlocks.length, 1);
  assert.strictEqual(faultBlocks[0].w, 100);
  assert.strictEqual(faultBlocks[0].h, 3);
});

runTest('6. Template, ParameterModal, and Go backend support discrete status card & custom labels', () => {
  assert(occContent.includes('scada-status-pill'), 'OCC card template must render scada-status-pill');
  assert(occContent.includes('scada-status-timeline'), 'OCC card template must render scada-status-timeline');
  assert(modalContent.includes('form.status_on_label') && modalContent.includes('form.status_off_label'), 'ParameterModal must bind custom status labels');
  assert(modalContent.includes('applyStatusLabelPreset'), 'ParameterModal must provide equipment status presets');
  assert(goModelContent.includes('StatusOnLabel') && goModelContent.includes('StatusOffLabel'), 'Go Parameter model must include StatusOnLabel & StatusOffLabel');
  assert(goServiceContent.includes('param.StatusOnLabel') && goServiceContent.includes('param.StatusOffLabel'), 'DeviceService must persist StatusOnLabel & StatusOffLabel');
});

console.log('\n====================================================================');
console.log(`TEST SUMMARY: ${passed} passed, ${failed} failed (Total: ${passed + failed})`);
console.log('====================================================================');

if (failed > 0) process.exit(1);

