/**
 * Automated Frontend Verification Test for Multi-Tenant & Multi-Site Hybrid Architecture
 * Run via: node web/tests/multi_tenant_ui_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const srcDir = path.join(__dirname, '../src');

const storeCode = fs.readFileSync(path.join(srcDir, 'store/index.js'), 'utf8');
const mainCode = fs.readFileSync(path.join(srcDir, 'main.js'), 'utf8');
const appCode = fs.readFileSync(path.join(srcDir, 'App.vue'), 'utf8');
const headerCode = fs.readFileSync(path.join(srcDir, 'components/Header.vue'), 'utf8');
const sidebarCode = fs.readFileSync(path.join(srcDir, 'components/Sidebar.vue'), 'utf8');
const routerCode = fs.readFileSync(path.join(srcDir, 'router/index.js'), 'utf8');
const tenantsViewCode = fs.readFileSync(path.join(srcDir, 'views/TenantsManagementView.vue'), 'utf8');
const accessControlCode = fs.readFileSync(path.join(srcDir, 'views/AccessControlView.vue'), 'utf8');
const deviceModalCode = fs.readFileSync(path.join(srcDir, 'components/DeviceModal.vue'), 'utf8');
const devicesViewCode = fs.readFileSync(path.join(srcDir, 'views/DevicesView.vue'), 'utf8');
const integrationViewCode = fs.readFileSync(path.join(srcDir, 'views/IntegrationView.vue'), 'utf8');
const configViewCode = fs.readFileSync(path.join(srcDir, 'views/ConfigView.vue'), 'utf8');

// 1. Vuex Store: Tenant & Site State, Actions, and WebSocket SET_SCOPE Handshake
assert(storeCode.includes('activeTenantId'), 'Vuex store must define activeTenantId state');
assert(storeCode.includes('activeSiteId'), 'Vuex store must define activeSiteId state');
assert(storeCode.includes('contextRevision'), 'Vuex store must define contextRevision state');
assert(storeCode.includes('fetchTenantContext'), 'Vuex store must define fetchTenantContext action');
assert(storeCode.includes('setActiveTenant'), 'Vuex store must define setActiveTenant action');
assert(storeCode.includes('setActiveSite'), 'Vuex store must define setActiveSite action');
assert(storeCode.includes('SET_SCOPE'), 'Vuex store must send SET_SCOPE handshake over WebSocket');

// 2. Axios Interceptor: X-Tenant-ID and X-Site-ID Header Injection
assert(mainCode.includes("'X-Tenant-ID'"), 'main.js Axios interceptor must inject X-Tenant-ID header');
assert(mainCode.includes("'X-Site-ID'"), 'main.js Axios interceptor must inject X-Site-ID header');

// 3. App.vue: Reactive Router-View Key on Tenant/Site Context Switch
assert(appCode.includes('contextRevision'), 'App.vue must key router-view by contextRevision for zero-touch workspace refresh');

// 4. Header.vue: Global SUPERADMIN Tenant Switcher + Site Switcher
assert(headerCode.includes('header-tenant-switcher'), 'Header.vue must render #header-tenant-switcher for SUPERADMIN');
assert(headerCode.includes('header-site-switcher'), 'Header.vue must render #header-site-switcher for multi-site switching');
assert(headerCode.includes('Semua Tenant (Global)'), 'Header.vue must allow SUPERADMIN to view Semua Tenant (Global)');
assert(headerCode.includes('Semua Stasiun'), 'Header.vue must allow selecting Semua Stasiun');

// 5. Router, Header & Sidebar: SUPERADMIN & Scoped ADMIN /administration/tenants + Cloud Mode Menu Restrictions
assert(routerCode.includes("path: '/administration/tenants'"), 'Router must register /administration/tenants');
assert(routerCode.includes('TenantsManagementView'), 'Router must wire TenantsManagementView');
assert(
  /path:\s*'\/administration\/tenants'[\s\S]{0,160}roles:\s*\['SUPERADMIN',\s*'ADMIN'\]/.test(routerCode),
  'Router must allow SUPERADMIN and scoped ADMIN on /administration/tenants'
);
assert(routerCode.includes("'/monitoring/devices'") && routerCode.includes('cloudRestrictedPrefixes'), 'Router must block Cloud-restricted paths for non-SUPERADMIN');
assert(sidebarCode.includes('v-if="isSuperAdmin || isAdmin" to="/administration/tenants"'), 'Sidebar must allow Tenant & Stasiun for SUPERADMIN and scoped ADMIN');
assert(sidebarCode.includes('isSuperAdmin || !isCloudMode'), 'Sidebar must hide Devices & Sensors in Cloud mode for non-Superadmin');
assert(sidebarCode.includes('isSuperAdmin || (!isOperator && !isCloudMode)'), 'Sidebar must hide Sensor Calibration in Cloud mode for non-Superadmin');
assert(/v-if="isSuperAdmin \|\| isAdmin"\s+to="\/administration\/tenants"/.test(headerCode), 'Header must allow Tenants shortcut for SUPERADMIN and scoped ADMIN');
assert(configViewCode.includes("initiateDomainSwitch('CLOUD')"), 'ConfigView must offer CLOUD configuration mode');
assert(configViewCode.includes('<!-- Configurable Telemetry Acquisition Modules (SUPERADMIN Only) -->'), 'ConfigView must restrict sensor polling & MQTT module toggles to SUPERADMIN');
assert(deviceModalCode.includes('input-edge-push-timeout-seconds'), 'DeviceModal must expose EDGE_PUSH / MQTT offline timeout input in seconds');
assert(tenantsViewCode.includes('saveBuilderDeviceTimeout'), 'TenantsManagementView must allow configuring device offline timeout in seconds');

// 6. TenantsManagementView.vue: 3 Tabs + Standardized Bearer Token, AQMS/SPARING/ONLIMO Payload Generator, Configurable Modules & Light-Mode CSS Compliance
assert(tenantsViewCode.includes("activeTab = 'tenants'"), 'TenantsManagementView must include tenants tab');
assert(tenantsViewCode.includes("activeTab = 'sites'"), 'TenantsManagementView must include sites tab');
assert(tenantsViewCode.includes("activeTab = 'ingest'"), 'TenantsManagementView must include ingest keys tab');
assert(tenantsViewCode.includes('/api/v1/ingest/telemetry'), 'TenantsManagementView must document /api/v1/ingest/telemetry endpoint');
assert(tenantsViewCode.includes('/api/v1/ingest/mqtt'), 'TenantsManagementView must document /api/v1/ingest/mqtt endpoint');
assert(tenantsViewCode.includes('Authorization: Bearer'), 'TenantsManagementView must document Authorization: Bearer token header');
assert(tenantsViewCode.includes("'AQMS', 'SPARING', 'ONLIMO'"), 'TenantsManagementView must support AQMS, SPARING, and ONLIMO logger payload templates');
assert(tenantsViewCode.includes('sensor_polling_enabled'), 'TenantsManagementView must allow toggling sensor_polling_enabled');
assert(tenantsViewCode.includes('mqtt_ingest_enabled'), 'TenantsManagementView must allow toggling mqtt_ingest_enabled');
assert(!tenantsViewCode.includes('v-model="keyForm.auto_provision"'), 'TenantsManagementView must enforce strict manual parameter registration (no auto_provision checkbox)');

const classRegex = /class="([^"]*)"/g;
let match;
while ((match = classRegex.exec(tenantsViewCode)) !== null) {
  const cls = match[1];
  const tokens = cls.split(/\s+/);
  if (tokens.includes('text-white')) {
    assert(
      tokens.includes('btn-text') || tokens.includes('badge-text'),
      `TenantsManagementView Light-Mode violation: class="${cls}" uses bare text-white without btn-text or badge-text`
    );
  }
}

// 7. AccessControlView.vue: Scoped User Tenant Selection & Allowed Site Multi-Select
assert(accessControlCode.includes('tenant_id'), 'AccessControlView must support tenant_id assignment');
assert(accessControlCode.includes('allowed_site_ids'), 'AccessControlView must support allowed_site_ids station restriction');

// 8. DeviceModal.vue & DevicesView.vue: Site Assignment & EDGE_PUSH Protocol
assert(deviceModalCode.includes('site_id'), 'DeviceModal must support site_id selection');
assert(deviceModalCode.includes('EDGE_PUSH'), 'DeviceModal must support EDGE_PUSH protocol');
assert(devicesViewCode.includes('EDGE_PUSH'), 'DevicesView must support EDGE_PUSH protocol badge/filter');

// 9. IntegrationView.vue: Edge-to-Central Outbound X-Ingest-Key Support
assert(integrationViewCode.includes('INGEST_KEY'), 'IntegrationView must support INGEST_KEY outbound authentication');

console.log('PASS: All Multi-Tenant, Multi-Site, and Hybrid Edge Push UI assertions verified!');
