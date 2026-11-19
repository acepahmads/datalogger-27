const fs = require('fs');
const path = require('path');
const assert = require('assert');

const webSrc = path.join(__dirname, '..', 'src');
const dashboardCode = fs.readFileSync(path.join(webSrc, 'views', 'DashboardOperatorView.vue'), 'utf8');
const downloadModalCode = fs.readFileSync(path.join(webSrc, 'components', 'ParameterDataDownloadModal.vue'), 'utf8');

// 1. DashboardOperatorView integrates ParameterDataDownloadModal, ActuatorTestWriteModal, and ActuatorModal
assert(dashboardCode.includes("import ParameterDataDownloadModal from '../components/ParameterDataDownloadModal.vue';"), 'DashboardOperatorView must import ParameterDataDownloadModal');
assert(dashboardCode.includes("import ActuatorModal from '../components/actuators/ActuatorModal.vue';"), 'DashboardOperatorView must import ActuatorModal');
assert(dashboardCode.includes("import ActuatorTestWriteModal from '../components/actuators/ActuatorTestWriteModal.vue';"), 'DashboardOperatorView must import ActuatorTestWriteModal');
assert(dashboardCode.includes('<ParameterDataDownloadModal'), 'DashboardOperatorView must mount ParameterDataDownloadModal');
assert(dashboardCode.includes('<ActuatorTestWriteModal'), 'DashboardOperatorView must mount ActuatorTestWriteModal');
assert(dashboardCode.includes('<ActuatorModal'), 'DashboardOperatorView must mount ActuatorModal');

// 2. Right-click context menu exposes Download Data for every parameter card/row
assert(dashboardCode.includes('id="ctx-btn-download-param-data"'), 'Context menu must include Download Data button');
assert(dashboardCode.includes('@click="handleContextDownloadData"'), 'Download Data button must invoke handleContextDownloadData');

// 3. Right-click context menu exposes Test Actuator, Edit Actuator, and + Tambah Actuator for linked / STATUS parameters (non-Operator roles only)
assert(dashboardCode.includes('v-if="canManageLinkedActuator && (contextMenuLinkedActuator || isContextMenuStatusParam)"'), 'Actuator context menu actions must appear on linked feedback parameters and STATUS parameters for non-Operator roles');
assert(dashboardCode.includes('id="ctx-btn-test-actuator"'), 'Context menu must include Test Actuator button');
assert(dashboardCode.includes('@click="handleContextTestActuator(contextMenuEffectiveActuator)"'), 'Test Actuator button must invoke handleContextTestActuator');
assert(dashboardCode.includes('id="ctx-btn-edit-actuator"'), 'Context menu must include Edit Actuator button');
assert(dashboardCode.includes('@click="handleContextEditActuator(contextMenuEffectiveActuator)"'), 'Edit Actuator button must invoke handleContextEditActuator');
assert(dashboardCode.includes('id="ctx-btn-add-actuator"') && dashboardCode.includes('@click="handleContextAddActuator"'), 'Context menu must include + Tambah Actuator shortcut for unlinked STATUS parameters');
assert(dashboardCode.includes('canManageLinkedActuator() {') && dashboardCode.includes('this.isNonOperator && this.canEditConfig'), 'canManageLinkedActuator must enforce non-Operator + canEditConfig');
assert(dashboardCode.includes('fetchActuatorsForDevice(deviceId)'), 'DashboardOperatorView must fetch per-device actuators so disabled/unlinked actuators are also detected');

// 4. Actuator feedback parameter mapping & status pill on both card and table views
assert(dashboardCode.includes('actuatorByFeedbackParamMap()'), 'DashboardOperatorView must build actuatorByFeedbackParamMap from feedback_parameter_id');
assert(dashboardCode.includes('feedback_parameter_id'), 'DashboardOperatorView must link actuators by feedback_parameter_id');
assert(dashboardCode.includes('getLinkedActuatorForParam(param)'), 'DashboardOperatorView must expose getLinkedActuatorForParam helper');
const pillMatches = dashboardCode.match(/v-if="getLinkedActuatorForParam\(param\)"/g) || [];
assert(pillMatches.length >= 2, `Expected linked actuator pill on both card and table views, found ${pillMatches.length}`);

// 5. ParameterDataDownloadModal supports time range filter, presets, formats, count preview & RBAC sanitization
assert(downloadModalCode.includes('id="param-dl-start"') && downloadModalCode.includes('id="param-dl-end"'), 'Download modal must have start/end datetime-local inputs');
assert(downloadModalCode.includes("applyPreset('24h')"), 'Download modal must default to 24h preset on open');
assert(downloadModalCode.includes('start_time: this.parsedStartDate.toISOString()'), 'Download modal must send RFC3339 start_time');
assert(downloadModalCode.includes('end_time: this.parsedEndDate.toISOString()'), 'Download modal must send RFC3339 end_time');
assert(downloadModalCode.includes('/api/devices/${this.targetDeviceId}/telemetry/history'), 'Download modal must query device telemetry history API');
assert(downloadModalCode.includes('triggerCsvDownload') && downloadModalCode.includes('triggerXlsDownload') && downloadModalCode.includes('triggerJsonDownload'), 'Download modal must support CSV, XLS, and JSON export');
assert(downloadModalCode.includes('if (!this.isOperator)'), 'Download modal must omit raw PDU/hex fields for Operator role');

// 6. Light-Mode CSS hygiene on ParameterDataDownloadModal.vue
const classRegex = /class="([^"]*)"/g;
let match;
while ((match = classRegex.exec(downloadModalCode)) !== null) {
  const cls = match[1];
  const tokens = cls.split(/\s+/);
  if (tokens.includes('text-white')) {
    assert(
      tokens.includes('btn-text') || tokens.includes('badge-text'),
      `ParameterDataDownloadModal Light-Mode violation: class="${cls}" uses bare text-white without btn-text or badge-text`
    );
  }
  assert(
    !tokens.includes('bg-slate-900'),
    `ParameterDataDownloadModal Light-Mode violation: class="${cls}" uses bare bg-slate-900 without dark: prefix`
  );
}

console.log('PASS: Dashboard Operator linked-actuator context menu (Test/Edit) & time-range Download Data modal verified!');
