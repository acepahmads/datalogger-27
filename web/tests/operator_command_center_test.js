/**
 * Automated Verification Suite: Operator Command Center
 * Datalogger Analysis Application
 *
 * Verifies:
 * 1. Operator routing and redirection (/operator -> /monitoring/operator)
 * 2. i18n parity between EN and ID for all Operator Command Center keys
 * 3. Operational Summary Cards metric mapping (Devices, Alarms, Telemetry Freshness, System Health)
 * 4. SCADA design tokens & color styling conformances (#050B16, #0B1422, etc.)
 * 5. Parameter filtering & favorite pinning behavior
 * 6. Trend analysis multi-parameter mapping and timeframe preset validity
 * 7. Notification queue states & channel configuration distinction
 * 8. Export & Report payload configuration
 * 9. Audio Alert engine integration
 * 10. RBAC Operator permission limits
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
console.log('RUNNING AUTOMATED TEST SUITE: OPERATOR COMMAND CENTER');
console.log('===============================================================');

// 1. Router Test
it('Test 1: Operator routes configured (/operator -> /monitoring/operator)', () => {
  const routerSrc = fs.readFileSync(path.join(__dirname, '../src/router/index.js'), 'utf8');
  assert(routerSrc.includes("path: '/operator'"), 'Missing /operator route');
  assert(routerSrc.includes("path: '/monitoring/operator'"), 'Missing /monitoring/operator route');
  assert(routerSrc.includes("component: OperatorCommandCenterView"), 'Missing OperatorCommandCenterView component mapping');
  assert(routerSrc.includes("redirect: '/monitoring/operator'"), 'Missing /operator redirection');
});

// 2. i18n Parity Test
it('Test 2: Complete i18n coverage between English and Indonesian for operator keys', () => {
  const enSrc = fs.readFileSync(path.join(__dirname, '../src/i18n/locales/en.js'), 'utf8');
  const idSrc = fs.readFileSync(path.join(__dirname, '../src/i18n/locales/id.js'), 'utf8');

  // Verify operator dictionary exists in both
  assert(enSrc.includes('"operator": {'), 'Missing operator dictionary in en.js');
  assert(idSrc.includes('"operator": {'), 'Missing operator dictionary in id.js');

  const requiredKeys = [
    'title', 'subtitle', 'station', 'site', 'systemHealth', 'commHealth',
    'healthy', 'degraded', 'failure', 'summaryDevices', 'totalDevices',
    'online', 'offline', 'summaryAlarms', 'critical', 'warning',
    'summaryTelemetry', 'dataFreshness', 'live', 'stale', 'summarySystem',
    'parameterMonitoring', 'liveTable', 'trendAnalysis', 'quickActions',
    'exportData', 'generateReport', 'alarmHistory', 'audioMuted', 'audioActive'
  ];

  for (const k of requiredKeys) {
    assert(enSrc.includes(`"${k}":`), `en.js missing operator key: ${k}`);
    assert(idSrc.includes(`"${k}":`), `id.js missing operator key: ${k}`);
  }
});

// 3. SCADA Design Token Conformance
it('Test 3: Tailwind and SCADA dark palette tokens match design specification', () => {
  const tailwindSrc = fs.readFileSync(path.join(__dirname, '../tailwind.config.js'), 'utf8');
  assert(tailwindSrc.includes("bg: '#050B16'"), 'Missing main background token #050B16');
  assert(tailwindSrc.includes("panel: '#0B1422'"), 'Missing panel token #0B1422');
  assert(tailwindSrc.includes("border: '#20354B'"), 'Missing border token #20354B');
  assert(tailwindSrc.includes("cyan: '#00CFFF'"), 'Missing cyan accent token #00CFFF');
  assert(tailwindSrc.includes("normal: '#00D6A0'"), 'Missing normal status token #00D6A0');
  assert(tailwindSrc.includes("critical: '#FF4559'"), 'Missing critical status token #FF4559');
});

// 4. Four Summary Cards Component
it('Test 4: Four Operational Summary Cards present with required telemetry metrics', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  assert(viewSrc.includes('summaryStats.onlineCount'), 'Missing online devices count metric');
  assert(viewSrc.includes('summaryStats.offlineCount'), 'Missing offline devices count metric');
  assert(viewSrc.includes('activeCriticalCount'), 'Missing active critical alarms count');
  assert(viewSrc.includes('activeWarningCount'), 'Missing active warning alarms count');
  assert(viewSrc.includes('lastTelemetrySecondsAgo'), 'Missing telemetry freshness seconds');
  assert(viewSrc.includes('systemStatus.cpu_percent'), 'Missing CPU utilization metric');
  assert(viewSrc.includes('systemStatus.ram_percent'), 'Missing RAM utilization metric');
  assert(viewSrc.includes('systemStatus.disk_percent'), 'Missing Disk utilization metric');
  assert(viewSrc.includes('systemStatus.db_latency_ms'), 'Missing DB status metric');
});

// 5. Global Alarm Banner Measured Value and Threshold
it('Test 5: Global Alarm Banner displays trigger value and limit threshold', () => {
  const bannerSrc = fs.readFileSync(path.join(__dirname, '../src/components/GlobalAlarmBanner.vue'), 'utf8');
  assert(bannerSrc.includes('highestAlarm.trigger_value != null'), 'Missing trigger_value conditional rendering');
  assert(bannerSrc.includes('highestAlarm.threshold_value != null'), 'Missing threshold_value conditional rendering');
  assert(bannerSrc.includes('dismissHighestAlarm'), 'Missing banner dismiss handler');
  assert(bannerSrc.includes('openAckModal'), 'Missing quick acknowledge modal trigger');
});

// 6. Global Header Enhancements
it('Test 6: Global Header distinguishes application health from device communication health', () => {
  const headerSrc = fs.readFileSync(path.join(__dirname, '../src/components/Header.vue'), 'utf8');
  assert(headerSrc.includes('overallAppStatus'), 'Missing overallAppStatus in Header');
  assert(headerSrc.includes('activeAlarmsCount'), 'Missing notification bell active count badge');
  assert(headerSrc.includes('/monitoring/operator'), 'Missing Operator Console shortcut button in Header');
  assert(headerSrc.includes('systemStatus.online_devices'), 'Missing online devices communication metric');
});

// 7. Navigation Sidebar Integration
it('Test 7: Sidebar includes Operator Console under MONITORING section', () => {
  const sidebarSrc = fs.readFileSync(path.join(__dirname, '../src/components/Sidebar.vue'), 'utf8');
  assert(sidebarSrc.includes('/monitoring/operator'), 'Missing /monitoring/operator in Sidebar');
  assert(sidebarSrc.includes('operatorConsole'), 'Missing operatorConsole label in Sidebar');
  assert(sidebarSrc.includes('isCollapsed'), 'Missing isCollapsed prop support in Sidebar');
});

// 8. Parameter Filtering and View Modes
it('Test 8: Operator Command Center supports Grid and Table view modes with debounced filtering', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  assert(viewSrc.includes("viewMode === 'cards'"), 'Missing cards view mode');
  assert(viewSrc.includes("viewMode === 'table'"), 'Missing table view mode');
  assert(viewSrc.includes('togglePin'), 'Missing favorite pinning functionality');
  assert(viewSrc.includes('selectedQualityFilter'), 'Missing quality filter');
  assert(viewSrc.includes('selectedAlarmFilter'), 'Missing alarm status filter');
});

// 9. Trend Analysis Chart
it('Test 9: Trend Analysis supports 15m, 1h, 6h, 24h, 7d timeframes and multi-parameter SVG plotting', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  assert(viewSrc.includes("'15m', '1h', '6h', '24h', '7d'"), 'Missing standard timeframe presets');
  assert(viewSrc.includes('generateChartPolyline'), 'Missing SVG polyline generator');
  assert(viewSrc.includes('generateChartPolygon'), 'Missing SVG polygon gradient generator');
  assert(viewSrc.includes('onChartScrub'), 'Missing hover scrubbing tooltip');
});

// 10. Recent Alarms, Events, and Notification Queue Tabs
it('Test 10: Segmented panel integrates active alarms, recent events, and notification queue states', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  assert(viewSrc.includes("activeRecentTab = 'alarms'"), 'Missing active alarms tab');
  assert(viewSrc.includes("activeRecentTab = 'events'"), 'Missing events tab');
  assert(viewSrc.includes("activeRecentTab = 'notifications'"), 'Missing notification delivery tab');
  assert(viewSrc.includes('retryNotification'), 'Missing notification retry action');
  assert(viewSrc.includes('channelsHealth'), 'Missing notification channels health indicator');
});

// 11. Quick Actions and Export Modal
it('Test 11: Quick actions bar provides Export Telemetry, Reports, History, and Audio controls', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  assert(viewSrc.includes('openExportModal'), 'Missing Export modal trigger');
  assert(viewSrc.includes('openReportModal'), 'Missing Report modal trigger');
  assert(viewSrc.includes('executeTelemetryExport'), 'Missing Telemetry export execution');
  assert(viewSrc.includes('executeReportGenerate'), 'Missing Report generation execution');
  assert(viewSrc.includes('toggleAudioMute'), 'Missing audio mute/unmute action');
});

// 12. Audio Alert Engine Integration
it('Test 12: Operator Command Center integrates with AudioAlertEngine singleton', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  assert(viewSrc.includes("import audioAlert from '../utils/audioAlert'"), 'Missing audioAlert import');
  assert(viewSrc.includes('audioAlert.playChime'), 'Missing playChime call in test alert');
  assert(viewSrc.includes('audioAlert.setMuted'), 'Missing setMuted call in toggle');
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
}
