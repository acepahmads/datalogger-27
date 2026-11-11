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

// 10. OperatorCommandCenterView.vue & DashboardOperatorView.vue: Cloud / EDGE_PUSH Stale Timeout & Gap Handling
const occViewCode = fs.readFileSync(path.join(srcDir, 'views', 'OperatorCommandCenterView.vue'), 'utf8');
const dashOpViewCode = fs.readFileSync(path.join(srcDir, 'views', 'DashboardOperatorView.vue'), 'utf8');
assert(occViewCode.includes('configuredStaleTimeoutSeconds'), 'OperatorCommandCenterView must compute configuredStaleTimeoutSeconds from setup timeout');
assert(occViewCode.includes('isPassiveOrCloudTelemetry'), 'OperatorCommandCenterView must detect passive/cloud telemetry mode');
assert(occViewCode.includes('getDeviceStaleTimeoutSeconds'), 'OperatorCommandCenterView must resolve per-device stale timeout in seconds');
assert(occViewCode.includes('Active Cloud Push Stream'), 'OperatorCommandCenterView stream diagnostics must display Active Cloud Push Stream');
assert(dashOpViewCode.includes('Active Cloud Push Stream'), 'DashboardOperatorView stream diagnostics must display Active Cloud Push Stream');
assert(dashOpViewCode.includes('configuredStaleTimeoutSeconds'), 'DashboardOperatorView must use configuredStaleTimeoutSeconds for stream diagnostics and bento sparklines');
assert(occViewCode.includes('isCloudMode()'), 'OperatorCommandCenterView must define isCloudMode computed property');
assert((occViewCode.match(/v-if="!isCloudMode"\s+class="relative toolbar-popover-anchor"/g) || []).length >= 2, 'OperatorCommandCenterView must hide Actuators and Network Monitor toolbar icons when isCloudMode is true');
assert((dashOpViewCode.match(/v-if="!isCloudMode"\s+class="relative toolbar-popover-anchor"/g) || []).length >= 2, 'DashboardOperatorView must hide Actuators and Network Monitor toolbar icons when isCloudMode is true');

// 11. Role RBAC on Organization Branding, Global Superadmin Maker Branding, System Logs Restriction, and Mandatory Cloud Login
const brandingUtilCode = fs.readFileSync(path.join(srcDir, 'utils', 'companyBranding.js'), 'utf8');
const loginViewCode = fs.readFileSync(path.join(srcDir, 'views', 'LoginView.vue'), 'utf8');
const logsViewCode = fs.readFileSync(path.join(srcDir, 'views', 'LogsView.vue'), 'utf8');

assert(occViewCode.includes('canManageCustomerBranding') && occViewCode.includes('canManageMakerBranding'), 'OperatorCommandCenterView must separate Customer vs Maker branding permissions');
assert(dashOpViewCode.includes('canManageCustomerBranding') && dashOpViewCode.includes('canManageMakerBranding'), 'DashboardOperatorView must separate Customer vs Maker branding permissions');
assert(configViewCode.includes(':disabled="!isSuperAdmin"'), 'ConfigView must disable Maker / Provider branding inputs for non-Superadmin');
assert(configViewCode.includes('resetCustomerBranding'), 'ConfigView must only reset Customer branding when non-Superadmin resets branding');
assert(brandingUtilCode.includes('buildBrandingRequestHeaders'), 'companyBranding.js must send Authorization, X-Tenant-ID, and X-Site-ID headers');

assert(sidebarCode.includes('<router-link v-if="!isOperator" to="/system/logs"'), 'Sidebar must hide /system/logs menu from OPERATOR role');
assert(/path:\s*'\/system\/logs'[\s\S]{0,180}disallowOperator:\s*true/.test(routerCode), 'Router must block OPERATOR on /system/logs');
assert(/path:\s*'\/system\/audit'[\s\S]{0,180}disallowOperator:\s*true/.test(routerCode), 'Router must block OPERATOR on /system/audit');
assert(/path:\s*'\/system\/communication-logs'[\s\S]{0,180}disallowOperator:\s*true/.test(routerCode), 'Router must block OPERATOR on /system/communication-logs');
assert(loginViewCode.includes("redirect.startsWith('/system/logs')"), 'LoginView must treat /system/logs as restricted for OPERATOR');
assert(logsViewCode.includes("this.$router.replace('/access-denied')"), 'LogsView must guard against direct OPERATOR mount');

assert(storeCode.includes("'X-Auto-Login': '1'"), 'autoLoginOperator must send X-Auto-Login header so backend blocks auto-login in Cloud mode');
assert(routerCode.includes('if (!isAuth && !isLoginPage && !isCloudMode)'), 'Router guard must never auto-login in Cloud mode');
assert(mainCode.includes('if (isCloudMode)'), 'main.js must enforce mandatory login and skip autoLoginOperator in Cloud mode');

// 12. LoginView Dark Mode Vue 2 Scoped CSS Compatibility & Cloud Mode Operator Delivery Queue Restriction
assert(!loginViewCode.includes(':global('), 'LoginView.vue must not use Vue 3 :global() inside Vue 2 <style scoped>');
assert(loginViewCode.includes('.login-dark .login-card'), 'LoginView.vue must define .login-dark .login-card');
assert(loginViewCode.includes('.login-dark .login-field'), 'LoginView.vue must define .login-dark .login-field');
assert(loginViewCode.includes('.login-dark .login-role-card'), 'LoginView.vue must define .login-dark .login-role-card');
assert(loginViewCode.includes('.login-dark .login-top-control'), 'LoginView.vue must define .login-dark .login-top-control');
assert(loginViewCode.includes('.login-dark .login-field:-webkit-autofill'), 'LoginView.vue must style dark mode browser autofill');

assert(sidebarCode.includes('v-if="!isOperator || !isCloudMode" to="/integration/queue"'), 'Sidebar must hide Delivery Queue for OPERATOR in Cloud mode');
assert(routerCode.includes("to.path === '/integration/queue'") && routerCode.includes('if (isCloudMode)'), 'Router must block OPERATOR from /integration/queue in Cloud mode');
assert(integrationViewCode.includes('if (this.isOperator && this.isCloudMode)'), 'IntegrationView must guard against OPERATOR mount in Cloud mode');

console.log('PASS: All Multi-Tenant, Multi-Site, and Hybrid Edge Push UI assertions verified!');
