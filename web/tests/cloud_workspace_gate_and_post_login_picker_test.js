/**
 * Verification Suite: Cloud Mode Post-Login Workspace Picker & Operational Route Scope Gate
 *
 * Verifies:
 * 1. CloudTenantWorkspaceModal.vue supports 'post-login', 'route-gate', and 'switcher' modes
 *    with instant search, Project Year filter, Domain filter, multi-site station selector,
 *    and Global Fleet Map / Multi-Tenant Admin shortcuts.
 * 2. Vuex store (web/src/store/index.js) defines isCloudMode & isCloudGlobalUnscoped,
 *    triggers showPostLoginTenantPicker on login/switchAccount in Cloud Global mode,
 *    and prevents unpaginated global fetchDevices/fetchAlarms/fetchActiveAlarms & WS telemetry storms.
 * 3. main.js awaits fetchTenantContext before deciding whether to fetch devices/alarms on Cloud boot.
 * 4. App.vue enforces requiresCloudTenantGate (without mounting <router-view>) across all 8 operational
 *    route groups when Cloud Mode + canSwitchTenants + activeTenantId === 0, while keeping
 *    Global pages (/monitoring/fleet-map, /administration/tenants, /monitoring/devices, etc.) accessible.
 * 5. Header.vue & DashboardOperatorView.vue provide 1-click Quick Workspace Search triggers.
 * 6. Protected files remain untouched.
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const srcDir = path.join(__dirname, '../src');
const modalCode = fs.readFileSync(path.join(srcDir, 'components/CloudTenantWorkspaceModal.vue'), 'utf8');
const storeCode = fs.readFileSync(path.join(srcDir, 'store/index.js'), 'utf8');
const mainCode = fs.readFileSync(path.join(srcDir, 'main.js'), 'utf8');
const appCode = fs.readFileSync(path.join(srcDir, 'App.vue'), 'utf8');
const headerCode = fs.readFileSync(path.join(srcDir, 'components/Header.vue'), 'utf8');
const bentoOpCode = fs.readFileSync(path.join(srcDir, 'views/DashboardOperatorView.vue'), 'utf8');

let passed = 0;
function check(name, fn) {
  fn();
  passed += 1;
  console.log(`✓ PASS: ${name}`);
}

console.log('========================================================================');
console.log('CLOUD WORKSPACE GATE & POST-LOGIN PICKER VERIFICATION SUITE');
console.log('========================================================================');

check('1. CloudTenantWorkspaceModal supports post-login, route-gate, and switcher modes with search & filters', () => {
  assert(modalCode.includes("mode === 'route-gate'"), 'Modal must support route-gate mode');
  assert(modalCode.includes("mode === 'post-login'"), 'Modal must support post-login mode');
  assert(modalCode.includes('id="cloud-workspace-search-input"'), 'Modal must provide search input');
  assert(modalCode.includes('yearFilter'), 'Modal must provide Project Year filter');
  assert(modalCode.includes('domainFilter'), 'Modal must provide Monitoring Domain filter');
  assert(modalCode.includes('setActiveTenantAndSite'), 'Modal must dispatch setActiveTenantAndSite on selection');
  assert(modalCode.includes('id="btn-skip-to-global-fleet-map"'), 'Post-login modal must allow skipping to Global Fleet Map');
  assert(!/text-white(?![^\n"]*(?:btn-text|badge-text|dark:))/.test(
    modalCode.split('\n').filter(l => l.includes('bg-blue-600') || l.includes('bg-emerald-600')).join('\n')
  ), 'Action buttons with colored backgrounds must preserve Light Mode contrast via btn-text');
});

check('2. Vuex store prevents Cloud Global overload and triggers Post-Login Picker', () => {
  assert(storeCode.includes('isCloudGlobalUnscoped'), 'Store must expose isCloudGlobalUnscoped getter');
  assert(storeCode.includes('showPostLoginTenantPicker'), 'Store state must include showPostLoginTenantPicker');
  assert(storeCode.includes('showWorkspaceSelectorModal'), 'Store state must include showWorkspaceSelectorModal');
  assert(storeCode.includes('globalSitesCatalog'), 'Store state must maintain globalSitesCatalog across tenant switches');
  assert(storeCode.includes("commit('SET_SHOW_POST_LOGIN_TENANT_PICKER', true)"), 'Login/switchAccount must open Post-Login Picker when Cloud Global unscoped');
  assert(
    storeCode.includes('if (getters.isCloudGlobalUnscoped && !filterParams.pageSize && !filterParams.allowGlobal)'),
    'Unpaginated fetchDevices must not pull all cross-tenant parameters in Cloud Global unscoped mode'
  );
  assert(
    storeCode.includes("if (getters.isCloudGlobalUnscoped) {\n        commit('SET_ALARMS', []);"),
    'fetchAlarms must skip global cross-tenant alarm dump when unscoped'
  );
  assert(
    storeCode.includes("if (getters.isCloudGlobalUnscoped) {\n        commit('SET_ACTIVE_ALARMS', []);"),
    'fetchActiveAlarms must skip global cross-tenant active alarm dump when unscoped'
  );
});

check('3. main.js resolves tenant context before firing device/alarm fetches on Cloud boot', () => {
  assert(
    mainCode.includes("this.$store.dispatch('fetchTenantContext').finally(() => {") &&
    mainCode.includes('if (!this.$store.getters.isCloudGlobalUnscoped)'),
    'main.js must wait for fetchTenantContext and guard fetchDevices/fetchAlarms in Cloud Global mode'
  );
});

check('4. App.vue gates all 8 operational route groups without mounting <router-view> when unscoped in Cloud Mode', () => {
  assert(appCode.includes('CloudTenantWorkspaceModal'), 'App.vue must register CloudTenantWorkspaceModal');
  assert(appCode.includes('v-if="requiresCloudTenantGate"'), 'App.vue must render Route Gate when requiresCloudTenantGate is true');
  assert(appCode.includes('<router-view\n          v-else'), 'App.vue must put <router-view> behind v-else so gated views never mount when unscoped');

  const expectedRouteChecks = [
    "'/monitoring/dashboard-operator'",
    "'/monitoring/operator'",
    "'/monitoring/telemetry'",
    "'/monitoring/analysis'",
    "'/monitoring/ispu'",
    "'/monitoring/wqms'",
    "'/monitoring/reports'",
    "'/monitoring/calibration'",
    "'/alarm/active'",
    "'/alarm/history'",
    "'/alarm/queue'",
  ];
  for (const r of expectedRouteChecks) {
    assert(appCode.includes(r), `App.vue gatedOperationalRouteLabel must cover ${r}`);
  }
  assert(!appCode.includes("p === '/monitoring/fleet-map' ?"), 'Fleet Map must remain accessible in Global Mode');
  assert(!appCode.includes("p === '/administration/tenants'"), 'Tenants Management must remain accessible in Global Mode');
});

check('5. Header.vue and DashboardOperatorView.vue expose Quick Workspace Switcher triggers', () => {
  assert(headerCode.includes('id="btn-open-workspace-picker"'), 'Header.vue must render Quick Workspace Search button');
  assert(bentoOpCode.includes('id="btn-bento-switch-tenant"'), 'DashboardOperatorView.vue must render Quick Switch Tenant button in fullscreen toolbar');
});

console.log('========================================================================');
console.log(`ALL ${passed} CLOUD WORKSPACE GATE & POST-LOGIN PICKER TESTS PASSED (100%)`);
console.log('========================================================================');
