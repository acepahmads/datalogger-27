/**
 * Automated Verification Suite:
 * 1. Configurable Active / Inactive Values for Discrete Status Parameters
 *    (ParameterModal.vue, DashboardOperatorView.vue, ParameterConfigViewModal.vue, Go backend)
 * 2. Actuator Configuration Copy/Paste + Cards/List Mode
 *    (ActuatorTab.vue, ActuatorModal.vue)
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const dashPath = path.join(__dirname, '../src/views/DashboardOperatorView.vue');
const dashContent = fs.readFileSync(dashPath, 'utf8').replace(/\r\n/g, '\n');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8').replace(/\r\n/g, '\n');

const paramModalPath = path.join(__dirname, '../src/components/ParameterModal.vue');
const paramModalContent = fs.readFileSync(paramModalPath, 'utf8').replace(/\r\n/g, '\n');

const actTabPath = path.join(__dirname, '../src/components/actuators/ActuatorTab.vue');
const actTabContent = fs.readFileSync(actTabPath, 'utf8').replace(/\r\n/g, '\n');

const actModalPath = path.join(__dirname, '../src/components/actuators/ActuatorModal.vue');
const actModalContent = fs.readFileSync(actModalPath, 'utf8').replace(/\r\n/g, '\n');

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

function extractMethodFromSource(src, methodName) {
  const regex = new RegExp(`\\n\\s{4}${methodName}\\s*\\(([^)]*)\\)\\s*\\{([\\s\\S]*?)\\n\\s{4}\\},`);
  const match = src.match(regex);
  if (!match) {
    throw new Error(`Could not extract method ${methodName}`);
  }
  const args = match[1].split(',').map(a => a.trim()).filter(Boolean);
  return new Function(...args, match[2]);
}

function createDashboardHarness(locale = 'id') {
  const h = {
    $i18n: { locale },
    isDark: true,
    perfSettings: { parameterSparklinesEnabled: true },
    categoryMap: {
      9: { id: 9, code: 'status', name: 'Status' },
    },
    sparklinesMap: {},
    isParamInAlarmOrFault(param) {
      if (!param) return false;
      if (param.device_status === 'OFFLINE' || param.quality === 'BAD' || param.quality === 'STALE') return true;
      if (param.alarm_status === 'CRITICAL' || param.alarm_status === 'WARNING') return true;
      return false;
    },
    getSparklineColor() {
      return '#F43F5E';
    },
  };

  h.getParamCategory = extractMethodFromSource(occContent, 'getParamCategory').bind(h);
  h.formatParamValue = extractMethodFromSource(occContent, 'formatParamValue').bind(h);
  h.resolveDiscreteStatusLabels = extractMethodFromSource(occContent, 'resolveDiscreteStatusLabels').bind(h);
  h.formatStatusDurationCompact = extractMethodFromSource(occContent, 'formatStatusDurationCompact').bind(h);

  h.isDiscreteStatusParam = extractMethodFromSource(dashContent, 'isDiscreteStatusParam').bind(h);
  h.evaluateDiscreteParamState = extractMethodFromSource(dashContent, 'evaluateDiscreteParamState').bind(h);
  h.getDiscreteStatusInfo = extractMethodFromSource(dashContent, 'getDiscreteStatusInfo').bind(h);
  h.getStatusTimelineBlocks = extractMethodFromSource(dashContent, 'getStatusTimelineBlocks').bind(h);

  return h;
}

console.log('====================================================================');
console.log('RUNNING TEST SUITE: CUSTOM STATUS VALUES & ACTUATOR COPY/LIST MODE');
console.log('====================================================================\n');

runTest('1. Default discrete status values preserve Active >= 1 and Inactive = 0', () => {
  const h = createDashboardHarness('id');
  const p = {
    param_id: 101,
    param_code: 'sampling',
    param_name: 'Sampling',
    category_id: 9,
    status_on_label: 'RUNNING',
    status_off_label: 'STOPPED',
    status_on_value: null,
    status_off_value: null,
    value: 1,
    device_status: 'ONLINE',
    quality: 'GOOD',
    alarm_status: 'NORMAL',
  };
  assert.strictEqual(h.getDiscreteStatusInfo(p).isOn, true);
  assert.strictEqual(h.getDiscreteStatusInfo(p).label, 'RUNNING');
  assert.strictEqual(h.getDiscreteStatusInfo({ ...p, value: 0 }).isOn, false);
  assert.strictEqual(h.getDiscreteStatusInfo({ ...p, value: 0 }).label, 'STOPPED');
});

runTest('2. Inverted Active-Low (status_on_value=0, status_off_value=1) evaluates 0 as RUNNING and 1 as STOPPED with matching timeline & duration', () => {
  const h = createDashboardHarness('id');
  const now = Date.now();
  h.sparklinesMap[102] = [
    { time: now - 20 * 60000, val: 1 }, // OFF in Active-Low
    { time: now - 15 * 60000, val: 1 }, // OFF in Active-Low
    { time: now - 10 * 60000, val: 0 }, // ON in Active-Low
    { time: now - 5 * 60000, val: 0 },  // ON in Active-Low
  ];
  const activeLowParam = {
    param_id: 102,
    param_code: 'sampling',
    param_name: 'Sampling',
    category_id: 9,
    status_on_label: 'RUNNING',
    status_off_label: 'STOPPED',
    status_on_value: 0,
    status_off_value: 1,
    value: 0,
    device_status: 'ONLINE',
    quality: 'GOOD',
    alarm_status: 'NORMAL',
  };

  const infoOn = h.getDiscreteStatusInfo(activeLowParam);
  assert.strictEqual(infoOn.isOn, true, 'Value 0 should be Active (RUNNING) when status_on_value=0');
  assert.strictEqual(infoOn.label, 'RUNNING');
  assert(infoOn.pillClass.includes('emerald'), 'Active-Low value 0 should render emerald badge');
  assert.strictEqual(infoOn.durationText, '10m', 'Contiguous Active-Low ON duration should be 10m');

  const blocks = h.getStatusTimelineBlocks(activeLowParam);
  assert.strictEqual(blocks.length, 2);
  assert.strictEqual(blocks[0].h, 8.0, 'Initial val=1 samples should render as OFF block (height 8)');
  assert.strictEqual(blocks[1].h, 15.0, 'Recent val=0 samples should render as ON block (height 15)');

  const infoOff = h.getDiscreteStatusInfo({ ...activeLowParam, value: 1 });
  assert.strictEqual(infoOff.isOn, false, 'Value 1 should be Inactive (STOPPED) when status_off_value=1');
  assert.strictEqual(infoOff.label, 'STOPPED');
});

runTest('3. Custom Modbus register status codes (e.g. 255 = ON, 0 = OFF or 2 = RUNNING, 5 = STOPPED) evaluate accurately', () => {
  const h = createDashboardHarness('id');
  const modbusParam = {
    param_id: 103,
    param_code: 'valve_state',
    category_id: 9,
    status_on_label: 'TERBUKA',
    status_off_label: 'TERTUTUP',
    status_on_value: 255,
    status_off_value: 0,
    value: 255,
    device_status: 'ONLINE',
    quality: 'GOOD',
    alarm_status: 'NORMAL',
  };
  assert.strictEqual(h.getDiscreteStatusInfo(modbusParam).isOn, true);
  assert.strictEqual(h.getDiscreteStatusInfo(modbusParam).label, 'TERBUKA');
  assert.strictEqual(h.getDiscreteStatusInfo({ ...modbusParam, value: 0 }).isOn, false);
  assert.strictEqual(h.getDiscreteStatusInfo({ ...modbusParam, value: 0 }).label, 'TERTUTUP');
});

runTest('4. ParameterModal binds status_on_value & status_off_value, swap helper, and validation against identical values', () => {
  assert(paramModalContent.includes('v-model="form.status_on_value"'), 'ParameterModal must bind form.status_on_value');
  assert(paramModalContent.includes('v-model="form.status_off_value"'), 'ParameterModal must bind form.status_off_value');
  assert(paramModalContent.includes('swapStatusValues'), 'ParameterModal must include swapStatusValues helper');
  assert(paramModalContent.includes('statusValueWarning'), 'ParameterModal must validate identical active/inactive values');
  assert(paramModalContent.includes('clear_status_on_value'), 'ParameterModal submitForm must include clear_status_on_value');
  assert(paramModalContent.includes('clear_status_off_value'), 'ParameterModal submitForm must include clear_status_off_value');
});

runTest('5. ActuatorTab provides Cards/List view mode toggle persisted in localStorage and Copy/Paste workflow', () => {
  assert(actTabContent.includes("setViewMode('grid')") && actTabContent.includes("setViewMode('list')"), 'ActuatorTab must provide Cards and List view mode buttons');
  assert(actTabContent.includes('datalogger_actuator_view_mode'), 'ActuatorTab must persist view mode in localStorage');
  assert(actTabContent.includes('<table'), 'ActuatorTab must render a structured table in List mode');
  assert(actTabContent.includes('copyActuatorConfig(act)'), 'ActuatorTab must provide Copy button on cards and list rows');
  assert(actTabContent.includes('openPasteModal'), 'ActuatorTab must provide Paste Actuator action');
  assert(actTabContent.includes('datalogger_actuator_clipboard_v1'), 'ActuatorTab & ActuatorModal must share clipboard storage key');

  // Verify smart unique code generator behavior directly
  const genUniqueCode = extractMethodFromSource(actTabContent, 'generateUniqueCode');
  // Cross-device paste (no collision): keeps exact code!
  const crossDevCtx = { actuators: [{ actuator_code: 'other_pump' }] };
  assert.strictEqual(genUniqueCode.call(crossDevCtx, 'act_sampling'), 'act_sampling');
  // Same-device paste (collision): increments cleanly!
  const sameDevCtx = { actuators: [{ actuator_code: 'act_sampling' }] };
  assert.strictEqual(genUniqueCode.call(sameDevCtx, 'act_sampling'), 'act_sampling_2');
  const numberedCtx = { actuators: [{ actuator_code: 'PUMP_01' }] };
  assert.strictEqual(genUniqueCode.call(numberedCtx, 'PUMP_01'), 'PUMP_02');
});

runTest('6. ActuatorModal supports draftActuator prefill, in-modal Copy/Paste, cross-device feedback parameter matching, and schedule cloning', () => {
  assert(actModalContent.includes('draftActuator'), 'ActuatorModal must accept draftActuator prop');
  assert(actModalContent.includes('copyCurrentFormConfig'), 'ActuatorModal must support copying config from modal');
  assert(actModalContent.includes('pasteIntoCurrentForm'), 'ActuatorModal must support pasting config inside modal');
  assert(actModalContent.includes('cloneCopiedSchedules'), 'ActuatorModal must support cloning schedules when pasting');
  assert(actModalContent.includes('/schedules'), 'ActuatorModal handleSubmit must create cloned schedules via POST /api/actuators/:id/schedules');

  const resolveFeedback = extractMethodFromSource(actModalContent, 'resolveFeedbackParameterId');
  const targetDeviceModalCtx = {
    parameters: [
      { id: 77, parameter_code: 'sampling', parameter_name: 'Sampling' },
    ],
  };
  // Source device had feedback_parameter_id: 12 ('sampling'), target device has id: 77 ('sampling')
  assert.strictEqual(
    resolveFeedback.call(targetDeviceModalCtx, { feedback_parameter_id: 12, feedback_parameter_code: 'sampling' }),
    77,
    'Should automatically map feedback parameter by parameter_code when pasting onto another device'
  );
});

console.log('\n====================================================================');
console.log(`TEST SUMMARY: ${passed} passed, ${failed} failed (Total: ${passed + failed})`);
console.log('====================================================================');

if (failed > 0) process.exit(1);
