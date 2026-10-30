const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('========================================================================');
console.log('RUNNING TEST SUITE: OPERATOR DASHBOARD PERSISTENCE & CONSOLIDATED BELL');
console.log('========================================================================');

const occViewPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

function test(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    process.exit(1);
  }
}

// 1. Preference persistence
test('1.1: viewMode is initialized from localStorage (cards vs table)', () => {
  assert(occContent.includes("viewMode: (() => {"), 'viewMode must have localStorage initializer function');
  assert(occContent.includes("localStorage.getItem('dl_operator_preferences')"), 'Must read dl_operator_preferences');
  assert(occContent.includes("localStorage.getItem('dl_view_mode')"), 'Must fallback to dl_view_mode');
});

test('1.2: gridColumns is initialized from localStorage (1P - 4P)', () => {
  assert(occContent.includes("gridColumns: (() => {"), 'gridColumns must have localStorage initializer function');
});

test('1.3: filterPinnedOnly and selectedCategoryFilter are initialized from localStorage', () => {
  assert(occContent.includes("filterPinnedOnly: (() => {"), 'filterPinnedOnly must have localStorage initializer');
  assert(occContent.includes("selectedCategoryFilter: (() => {"), 'selectedCategoryFilter must have localStorage initializer');
});

test('1.4: Trend timeframe and chart mode are initialized from localStorage', () => {
  assert(occContent.includes("currentTimeframe: (() => {"), 'currentTimeframe must have localStorage initializer');
  assert(occContent.includes("chartMode: (() => {"), 'chartMode must have localStorage initializer');
});

test('1.5: saveOperatorPreferences saves all dashboard setups to localStorage', () => {
  assert(occContent.includes('saveOperatorPreferences() {'), 'saveOperatorPreferences method must exist');
  assert(occContent.includes("localStorage.setItem('dl_operator_preferences'"), 'Must persist dl_operator_preferences');
  assert(occContent.includes("localStorage.setItem('dl_view_mode'"), 'Must persist dl_view_mode for compatibility');
});

test('1.6: loadOperatorPreferences restores all dashboard setups', () => {
  assert(occContent.includes('loadOperatorPreferences() {'), 'loadOperatorPreferences method must exist');
  assert(occContent.includes('this.loadOperatorPreferences();'), 'mounted() must call loadOperatorPreferences()');
});

test('1.7: Vue watchers persist setup changes automatically', () => {
  assert(occContent.includes('viewMode() {\n      this.saveOperatorPreferences();'), 'viewMode watcher must exist');
  assert(occContent.includes('gridColumns() {\n      this.saveOperatorPreferences();'), 'gridColumns watcher must exist');
  assert(occContent.includes('filterPinnedOnly() {\n      this.saveOperatorPreferences();'), 'filterPinnedOnly watcher must exist');
  assert(occContent.includes('selectedCategoryFilter() {\n      this.saveOperatorPreferences();'), 'selectedCategoryFilter watcher must exist');
  assert(occContent.includes('currentTimeframe() {\n      this.saveOperatorPreferences();'), 'currentTimeframe watcher must exist');
});

// 2. Consolidated Bell Icon
test('2.1: The 3 status icons (station, bell, activity) are consolidated into single bell icon', () => {
  // Ensure the old separate station and activity toolbar buttons are removed from top bar
  assert(!occContent.includes("@click.stop=\"toggleToolbarPopover('deviceStatus')\""), 'Old deviceStatus button must be removed');
  assert(!occContent.includes("@click.stop=\"toggleToolbarPopover('latestTelemetry')\""), 'Old latestTelemetry button must be removed');
  assert(occContent.includes("@click.stop=\"toggleToolbarPopover('healthAndAlarms')\""), 'New consolidated healthAndAlarms button must exist');
});

test('2.2: Consolidated station health button displays distinct pulse/heartbeat icon (no duplicate bell vs top header) and dynamic indicator badge/dot', () => {
  assert(occContent.includes('aria-label="Alarms & Station Health Center"'), 'Aria label for consolidated health center must exist');
  assert(occContent.includes(':title="healthAndAlarmsTooltip"'), 'Tooltip must use healthAndAlarmsTooltip');
  assert(occContent.includes('<OperatorDashboardIcon name="pulse" :size="16" />'), 'Distinct pulse/heartbeat icon must be rendered on operator toolbar');
});

test('2.4: Top header bell, global alarm banner, and operator health center use synchronized operational alarm filtering', () => {
  const storeContent = fs.readFileSync(path.join(__dirname, '..', 'src', 'store', 'index.js'), 'utf8');
  const headerContent = fs.readFileSync(path.join(__dirname, '..', 'src', 'components', 'Header.vue'), 'utf8');
  const bannerContent = fs.readFileSync(path.join(__dirname, '..', 'src', 'components', 'GlobalAlarmBanner.vue'), 'utf8');

  assert(storeContent.includes('operationalActiveAlarms'), 'Store must expose operationalActiveAlarms getter');
  assert(storeContent.includes('unacknowledgedActiveAlarms'), 'Store must expose unacknowledgedActiveAlarms getter');
  assert(headerContent.includes('unacknowledgedActiveAlarms'), 'Header bell must use synchronized unacknowledgedActiveAlarms getter');
  assert(bannerContent.includes('operationalActiveAlarms'), 'GlobalAlarmBanner must use synchronized operationalActiveAlarms getter');
  assert(occContent.includes("if (!a || (a.status && a.status !== 'ACTIVE')) return false;"), 'Operator activeAlarmsList must align on ACTIVE status');
});

test('2.3: Popover contains 3 interactive tabs for Alarms, Devices, and Telemetry', () => {
  assert(occContent.includes("@click=\"healthActiveTab = 'alarms'\""), 'Alarms tab selector must exist');
  assert(occContent.includes("@click=\"healthActiveTab = 'devices'\""), 'Devices tab selector must exist');
  assert(occContent.includes("@click=\"healthActiveTab = 'telemetry'\""), 'Telemetry tab selector must exist');
  assert(occContent.includes("v-if=\"healthActiveTab === 'alarms'\""), 'Alarms tab content must exist');
  assert(occContent.includes("v-else-if=\"healthActiveTab === 'devices'\""), 'Devices tab content must exist');
  assert(occContent.includes("v-else-if=\"healthActiveTab === 'telemetry'\""), 'Telemetry tab content must exist');
});

console.log('========================================================================');
console.log('ALL TESTS PASSED SUCCESSFULLY!');
console.log('========================================================================');
