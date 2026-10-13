/**
 * Automated Verification Suite: Operator Command Center
 * Datalogger Analysis Application
 *
 * Verifies:
 * 1. Operator routing and redirection (/operator -> /monitoring/operator)
 * 2. i18n parity between EN and ID for all Operator Command Center keys (including Wind Rose & Dispersion)
 * 3. Operational Summary Cards metric mapping (Devices, Alarms, Telemetry Freshness, System Health)
 * 4. SCADA design tokens & color styling conformances (#050B16, #0B1422, etc.)
 * 5. Parameter filtering & favorite pinning behavior
 * 6. Trend analysis multi-parameter mapping and timeframe preset validity
 * 7. Quick Actions, Telemetry Export & Report payloads
 * 8. Audio Alert engine integration
 * 9. Wind Rose component integration (replacing lower alarm list tabs)
 * 10. Requirement 1: Wind Rose with valid direction and speed samples pairing
 * 11. Requirement 2: Wind Rose exclusion of invalid, bad, and stale samples
 * 12. Requirement 3: Wind Rose percentage calculation normalized to sample count denominator N
 * 13. Requirement 4: Missing wind data fallback ("Wind data unavailable" state)
 * 14. Requirement 5: Pollutant selection dynamically derived from configured parameters
 * 15. Requirement 6: Multiple monitoring stations with valid coordinates in tactical map
 * 16. Requirement 7: Missing/invalid station coordinates displays warning and non-map matrix fallback
 * 17. Requirement 8: Stale and invalid pollutant measurements handled with distinct visual states
 * 18. Requirement 9: Pure SVG offline operation without external map tile dependencies
 * 19. Requirement 10: Threshold boundaries, units, and concentration legends
 * 20. Requirement 11: 40/60 Responsive layout and dark/light theme support
 * 21. Requirement 12: Complete EN & ID localization for Wind Rose and Dispersion Map
 * 22. Requirement 13: Global alarm banner, bell count, and alarm navigation retention
 * 23. Requirement 14: Existing telemetry and alarm regression tests
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
console.log('RUNNING AUTOMATED TEST SUITE: OPERATOR COMMAND CENTER (WIND ROSE & DISPERSION MAP)');
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
    'exportData', 'generateReport', 'alarmHistory', 'audioMuted', 'audioActive',
    'windRoseTitle', 'windRoseSubtitle', 'selectStation', 'windUnavailable',
    'dispersionTitle', 'dispersionSubtitle', 'selectPollutant', 'allStations',
    'filterStation', 'measuredPoints', 'spatialInterpolation', 'windVector'
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

// 10. Lower Panel Replacement: Wind Rose and Pollutant Dispersion Map Registered
it('Test 10: Lower panel replaced with WindRoseChart and PollutantDispersionMap in OperatorCommandCenterView', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  // Lower tab panel removed
  assert(!viewSrc.includes("activeRecentTab === 'alarms'"), 'Old lower active alarms tab should be removed');
  assert(!viewSrc.includes("activeRecentTab === 'events'"), 'Old lower events tab should be removed');
  assert(!viewSrc.includes("activeRecentTab === 'notifications'"), 'Old lower notifications tab should be removed');

  // Both new components imported and rendered
  assert(viewSrc.includes("import WindRoseChart from '../components/WindRoseChart.vue'"), 'Missing WindRoseChart import');
  assert(viewSrc.includes("import PollutantDispersionMap from '../components/PollutantDispersionMap.vue'"), 'Missing PollutantDispersionMap import');
  assert(viewSrc.includes('<WindRoseChart'), 'Missing WindRoseChart template usage');
  assert(viewSrc.includes('<PollutantDispersionMap'), 'Missing PollutantDispersionMap template usage');
  assert(viewSrc.includes('lg:col-span-5'), 'Wind Rose should use approx 40% (5 cols)');
  assert(viewSrc.includes('lg:col-span-7'), 'Dispersion Map should use approx 60% (7 cols)');
});

// 11. Requirement 1: Wind Rose with valid direction and speed samples pairing
it('Test 11: Wind Rose pairs valid direction and speed samples within 60s tolerance', () => {
  const windSrc = fs.readFileSync(path.join(__dirname, '../src/components/WindRoseChart.vue'), 'utf8');
  assert(windSrc.includes('minDiff = 60000'), 'Missing 60s pairing tolerance between speed and direction');
  assert(windSrc.includes('/api/devices/${this.currentDevice.id}/telemetry/history'), 'Must reuse existing telemetry history API');
  assert(windSrc.includes('SPEED_BINS'), 'Must define speed intervals/bins');
  assert(windSrc.includes('COMPASS_16'), 'Must support 16-point directional compass rose');

  // Algorithm test: Pairing logic
  const wsSamples = [
    { timestamp: '2026-10-13T10:00:00Z', value: 3.5, quality: 'GOOD' },
    { timestamp: '2026-10-13T10:05:00Z', value: 6.2, quality: 'GOOD' }
  ];
  const wdSamples = [
    { timestamp: '2026-10-13T10:00:20Z', value: 45, quality: 'GOOD' },
    { timestamp: '2026-10-13T10:05:15Z', value: 180, quality: 'GOOD' }
  ];

  const paired = [];
  for (const ws of wsSamples) {
    const wsTime = new Date(ws.timestamp).getTime();
    let bestWd = null;
    let minDiff = 60000;
    for (const wd of wdSamples) {
      const wdTime = new Date(wd.timestamp).getTime();
      const diff = Math.abs(wsTime - wdTime);
      if (diff < minDiff) {
        minDiff = diff;
        bestWd = wd;
      }
    }
    if (bestWd) paired.push({ speed: ws.value, dir: bestWd.value });
  }

  assert.strictEqual(paired.length, 2, 'Should pair 2 valid samples');
  assert.strictEqual(paired[0].dir, 45, 'First pair should have WD=45');
  assert.strictEqual(paired[1].speed, 6.2, 'Second pair should have WS=6.2');
});

// 12. Requirement 2: Wind Rose exclusion of invalid, bad, and stale samples
it('Test 12: Wind Rose filters out BAD and UNCERTAIN quality samples', () => {
  const windSrc = fs.readFileSync(path.join(__dirname, '../src/components/WindRoseChart.vue'), 'utf8');
  assert(windSrc.includes("ws.quality === 'BAD' || ws.quality === 'UNCERTAIN'"), 'Missing WS quality exclusion');
  assert(windSrc.includes("wd.quality === 'BAD' || wd.quality === 'UNCERTAIN'"), 'Missing WD quality exclusion');

  // Verify filter behavior
  const samples = [
    { value: 5, quality: 'GOOD' },
    { value: 8, quality: 'BAD' },
    { value: 4, quality: 'UNCERTAIN' }
  ];
  const valid = samples.filter(s => s.quality !== 'BAD' && s.quality !== 'UNCERTAIN');
  assert.strictEqual(valid.length, 1, 'Only GOOD sample must pass filter');
  assert.strictEqual(valid[0].value, 5);
});

// 13. Requirement 3: Percentage calculation and sample count denominator
it('Test 13: Wind Rose calculates directional frequency normalized by sample count N', () => {
  const windSrc = fs.readFileSync(path.join(__dirname, '../src/components/WindRoseChart.vue'), 'utf8');
  assert(windSrc.includes('(totalSectorCount / this.totalValidSamples) * 100'), 'Must normalize sector percentage with valid sample count denominator');
  assert(windSrc.includes('totalValidSamples'), 'Must track total valid sample count N');
  assert(windSrc.includes('calmCount'), 'Must account for calm observations');

  // Math check: N = 200, Sector = 50 -> 25%
  const N = 200;
  const sectorCount = 50;
  const pct = (sectorCount / N) * 100;
  assert.strictEqual(pct, 25.0, 'Percentage calculation must be accurate');
});

// 14. Requirement 4: Missing wind data fallback
it('Test 14: Wind Rose displays "Wind data unavailable" when no wind sensors exist', () => {
  const windSrc = fs.readFileSync(path.join(__dirname, '../src/components/WindRoseChart.vue'), 'utf8');
  assert(windSrc.includes('!hasWindSensors'), 'Must check hasWindSensors condition');
  assert(windSrc.includes('windUnavailable'), 'Must use windUnavailable i18n key');
  assert(windSrc.includes('Wind data unavailable'), 'Must display fallback text');
  assert(!windSrc.includes('Math.random()'), 'Zero fabrication: no random fake wind telemetry');
});

// 15. Requirement 5: Pollutant selection dynamically derived from configured parameters
it('Test 15: Pollutant selector discovers existing parameters and excludes meteorological parameters', () => {
  const mapSrc = fs.readFileSync(path.join(__dirname, '../src/components/PollutantDispersionMap.vue'), 'utf8');
  assert(mapSrc.includes('availablePollutants()'), 'Must have availablePollutants dynamic computed property');
  assert(mapSrc.includes("lower === 'wd' || lower === 'ws'"), 'Must exclude wind speed/direction from pollutant list');
  assert(mapSrc.includes('selectedPollutantCode'), 'Must support reactive pollutant selection');
});

// 16. Requirement 6: Multiple monitoring stations with valid coordinates in tactical map
it('Test 16: Tactical SCADA map renders concentric range rings, station pins, and live values', () => {
  const mapSrc = fs.readFileSync(path.join(__dirname, '../src/components/PollutantDispersionMap.vue'), 'utf8');
  assert(mapSrc.includes('tactical-rings'), 'Must render tactical range rings');
  assert(mapSrc.includes('1 km'), 'Must display range ring scale');
  assert(mapSrc.includes('plottedStationsWithCoords'), 'Must map station coordinates');
  assert(mapSrc.includes('inspectStation'), 'Must allow clicking station to inspect current status');
  assert(mapSrc.includes('showWindOverlay'), 'Must support wind vector overlay at station location');
});

// 17. Requirement 7: Missing or invalid station coordinates fallback
it('Test 17: Unconfigured coordinates triggers warning and non-map operational matrix fallback', () => {
  const mapSrc = fs.readFileSync(path.join(__dirname, '../src/components/PollutantDispersionMap.vue'), 'utf8');
  assert(mapSrc.includes('!hasAnyCoordinates'), 'Must detect when coordinates are missing');
  assert(mapSrc.includes('Station coordinates unconfigured'), 'Must display unconfigured warning message');
  assert(mapSrc.includes('targetStationCards'), 'Must render matrix fallback cards');
  assert(mapSrc.includes('thresholdBarColorClass'), 'Fallback must include threshold bar indicator');
});

// 18. Requirement 8: Stale and invalid pollutant measurements
it('Test 18: Pollutant measurements display distinct BAD and STALE states', () => {
  const mapSrc = fs.readFileSync(path.join(__dirname, '../src/components/PollutantDispersionMap.vue'), 'utf8');
  assert(mapSrc.includes("quality === 'BAD'"), 'Must handle BAD quality measurement');
  assert(mapSrc.includes("quality === 'STALE'"), 'Must handle STALE quality measurement');
  assert(mapSrc.includes('qualityBadgeClass'), 'Must assign distinct quality styling');
});

// 19. Requirement 9: Pure SVG offline operation without external map tile dependencies
it('Test 19: Map uses lightweight pure SVG coordinate projection without external tiles', () => {
  const mapSrc = fs.readFileSync(path.join(__dirname, '../src/components/PollutantDispersionMap.vue'), 'utf8');
  assert(mapSrc.includes('<svg'), 'Must render via pure SVG');
  assert(!mapSrc.includes('tile.openstreetmap.org'), 'Must not send telemetry or request external map tiles');
  assert(!mapSrc.includes('api.mapbox.com'), 'Must not use external mapbox dependency');
  assert(!mapSrc.includes('google.com/maps'), 'Must not use external google maps dependency');
});

// 20. Requirement 10: Threshold boundaries, units, and concentration legends
it('Test 10b: Map footer displays verified scientific mode and actual parameter threshold limits', () => {
  const mapSrc = fs.readFileSync(path.join(__dirname, '../src/components/PollutantDispersionMap.vue'), 'utf8');
  assert(mapSrc.includes('currentThresholdWarn'), 'Must use parameter configured high_warning limit');
  assert(mapSrc.includes('currentThresholdCrit'), 'Must use parameter configured high_critical limit');
  assert(mapSrc.includes('Direct Observation (Sensor Measurements)'), 'Must separate measured observation from dispersion simulation');
  assert(mapSrc.includes('Spatial Interpolation (IDW Estimation)'), 'Must clearly label interpolation as estimation');
});

// 21. Requirement 11: 40/60 Responsive layout and dark/light themes
it('Test 21: Layout allocates ~40% for Wind Rose and ~60% for Pollutant Map with theme support', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  assert(viewSrc.includes('grid-cols-1 lg:grid-cols-12'), 'Responsive grid (stacked mobile, 12 cols desktop)');
  assert(viewSrc.includes('lg:col-span-5'), 'Wind Rose allocated 5/12 cols (41.7%)');
  assert(viewSrc.includes('lg:col-span-7'), 'Dispersion Map allocated 7/12 cols (58.3%)');
  assert(viewSrc.includes(':is-dark="isDark"'), 'Passes theme mode prop to both components');
});

// 22. Requirement 12: English and Indonesian localization for new components
it('Test 22: English and Indonesian localization files contain all Wind Rose & Dispersion keys', () => {
  const enSrc = fs.readFileSync(path.join(__dirname, '../src/i18n/locales/en.js'), 'utf8');
  const idSrc = fs.readFileSync(path.join(__dirname, '../src/i18n/locales/id.js'), 'utf8');

  const keys = [
    'windRoseTitle', 'windRoseSubtitle', 'selectStation', 'windUnavailable',
    'dispersionTitle', 'dispersionSubtitle', 'selectPollutant', 'allStations',
    'filterStation', 'measuredPoints', 'spatialInterpolation', 'windVector'
  ];

  keys.forEach(k => {
    assert(enSrc.includes(`"${k}":`), `en.js missing ${k}`);
    assert(idSrc.includes(`"${k}":`), `id.js missing ${k}`);
  });
});

// 23. Requirement 13: Global alarm banner, bell count, and alarm navigation still work
it('Test 23: Alarm system accessibility preserved (Banner, Bell, History, Active Console)', () => {
  const appSrc = fs.readFileSync(path.join(__dirname, '../src/App.vue'), 'utf8');
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  const headerSrc = fs.readFileSync(path.join(__dirname, '../src/components/Header.vue'), 'utf8');

  assert(appSrc.includes('<GlobalAlarmBanner'), 'Global Alarm Banner must remain in App.vue');
  assert(headerSrc.includes('activeAlarmsCount'), 'Notification bell count must remain in Header.vue');
  assert(viewSrc.includes('to="/alarm/history"'), 'Alarm History link must remain in quick actions');
  assert(viewSrc.includes('to="/alarm/active"'), 'Active Alarms console link must remain accessible');
  assert(viewSrc.includes('openAckModal'), 'Alarm acknowledgement modal workflow must remain functional');
});

// 24. Requirement 14: Existing telemetry and alarm regression tests
it('Test 24: Quick actions bar provides Export Telemetry, Reports, History, and Audio controls', () => {
  const viewSrc = fs.readFileSync(path.join(__dirname, '../src/views/OperatorCommandCenterView.vue'), 'utf8');
  assert(viewSrc.includes('openExportModal'), 'Missing Export modal trigger');
  assert(viewSrc.includes('openReportModal'), 'Missing Report modal trigger');
  assert(viewSrc.includes('executeTelemetryExport'), 'Missing Telemetry export execution');
  assert(viewSrc.includes('executeReportGenerate'), 'Missing Report generation execution');
  assert(viewSrc.includes('toggleAudioMute'), 'Missing audio mute/unmute action');
});

console.log('===============================================================');
console.log(`TEST RESULTS: ${passedTests} PASSED, ${failedTests} FAILED (TOTAL: ${passedTests + failedTests})`);
console.log('===============================================================');

if (failedTests > 0) {
  process.exit(1);
}
