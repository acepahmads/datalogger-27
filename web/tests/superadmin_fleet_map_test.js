/**
 * Automated Verification Test for Superadmin Ultra-Compact Global Fleet & Tenant Map Dashboard
 * (Desktop 1-Row Ribbon + Split Map/Dock & Mobile HP Mode Switcher)
 * Run via: node web/tests/superadmin_fleet_map_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const rootDir = path.join(__dirname, '../..');
const srcDir = path.join(__dirname, '../src');

const fleetViewCode = fs.readFileSync(path.join(srcDir, 'views/FleetMapView.vue'), 'utf8');
const routerCode = fs.readFileSync(path.join(srcDir, 'router/index.js'), 'utf8');
const sidebarCode = fs.readFileSync(path.join(srcDir, 'components/Sidebar.vue'), 'utf8');
const headerCode = fs.readFileSync(path.join(srcDir, 'components/Header.vue'), 'utf8');
const appCode = fs.readFileSync(path.join(srcDir, 'App.vue'), 'utf8');
const storeCode = fs.readFileSync(path.join(srcDir, 'store/index.js'), 'utf8');
const loginViewCode = fs.readFileSync(path.join(srcDir, 'views/LoginView.vue'), 'utf8');
const mainGoCode = fs.readFileSync(path.join(rootDir, 'cmd/main.go'), 'utf8');
const tenantHandlerCode = fs.readFileSync(path.join(rootDir, 'internal/handler/tenant_handler.go'), 'utf8');
const fleetServiceCode = fs.readFileSync(path.join(rootDir, 'internal/service/fleet_overview.go'), 'utf8');

// 1. Backend Endpoint & Strict SUPERADMIN RBAC + Null-Island Guard
assert(
  mainGoCode.includes('tenantsGroup.GET("/fleet-overview", middleware.RequireRole("SUPERADMIN", "ADMINISTRATOR"), tenantHandler.GetFleetOverview)'),
  'cmd/main.go must register GET /api/tenants/fleet-overview restricted to SUPERADMIN'
);
assert(
  tenantHandlerCode.includes('func (h *TenantHandler) GetFleetOverview') && tenantHandlerCode.includes('if !scope.IsSuperAdmin'),
  'TenantHandler.GetFleetOverview must enforce strict scope.IsSuperAdmin guard'
);
assert(
  fleetServiceCode.includes('func isValidFleetCoordinate') && fleetServiceCode.includes('math.Abs(la) < 0.0001 && math.Abs(ln) < 0.0001'),
  'fleet_overview.go must guard against (0, 0) Null-Island coordinates'
);
assert(
  fleetServiceCode.includes('DEVICE_FALLBACK'),
  'fleet_overview.go must fall back to Device coordinates when Site coordinates are unconfigured'
);

// 2. Router, Sidebar, Header, Vuex & App Shell Integration
assert(
  /path:\s*'\/monitoring\/fleet-map'[\s\S]{0,180}roles:\s*\['SUPERADMIN'\]/.test(routerCode),
  'Router must register /monitoring/fleet-map strictly restricted to SUPERADMIN'
);
assert(
  routerCode.includes("path: '/fleet-map', redirect: '/monitoring/fleet-map'"),
  'Router must provide /fleet-map shortcut redirect'
);
assert(
  sidebarCode.includes('v-if="isSuperAdmin"') && sidebarCode.includes('to="/monitoring/fleet-map"'),
  'Sidebar must expose Fleet & Tenant Map strictly for SUPERADMIN'
);
assert(
  headerCode.includes('id="header-fleet-map-shortcut"') && headerCode.includes('to="/monitoring/fleet-map"'),
  'Header must expose 1-click Fleet Map shortcut next to global Tenant Switcher for SUPERADMIN'
);
assert(
  storeCode.includes('async setActiveTenantAndSite('),
  'Vuex store must provide atomic setActiveTenantAndSite action for 1-click station context switching'
);
assert(
  appCode.includes('isFleetMapRoute') && appCode.includes('overflow-hidden flex flex-col min-h-0'),
  'App.vue must grant /monitoring/fleet-map a zero-double-scroll full-height flex viewport'
);
assert(
  loginViewCode.includes("redirect.startsWith('/monitoring/fleet-map')"),
  'LoginView must protect /monitoring/fleet-map redirects for non-Superadmin logins'
);

// 3. Desktop Ultra-Compact Layout Assertions
assert(fleetViewCode.includes('id="fleet-desktop-command-ribbon"'), 'FleetMapView must render 1-row desktop command & KPI ribbon');
assert(fleetViewCode.includes('id="input-fleet-search"'), 'FleetMapView must include compact station/tenant search input');
assert(fleetViewCode.includes('id="select-fleet-tenant"'), 'FleetMapView must include Tenant dropdown filter');
assert(fleetViewCode.includes('id="select-fleet-domain"'), 'FleetMapView must include Domain dropdown filter (AQMS/WQMS/SPARING)');
assert(fleetViewCode.includes('id="select-fleet-alarm"'), 'FleetMapView must include Alarm filter dropdown');
assert(fleetViewCode.includes('id="fleet-leaflet-map"'), 'FleetMapView must render Leaflet interactive map canvas');
assert(fleetViewCode.includes('id="fleet-desktop-site-card"'), 'FleetMapView must render floating desktop site preview card');
assert(fleetViewCode.includes('id="fleet-priority-dock"'), 'FleetMapView must render docked Prioritas Stasiun panel');
assert(fleetViewCode.includes("axios.get('/api/tenants/fleet-overview')"), 'FleetMapView must fetch /api/tenants/fleet-overview');
assert(fleetViewCode.includes("this.$store.dispatch('setActiveTenantAndSite'"), 'FleetMapView must switch tenant & site context before opening station dashboard');
assert(fleetViewCode.includes('Math.abs(la) < 0.0001 && Math.abs(ln) < 0.0001'), 'FleetMapView frontend must also enforce Null-Island (0, 0) guard');

// 4. Mobile / HP Mode (< 768px) Assertions
assert(fleetViewCode.includes('id="fleet-mobile-header"'), 'FleetMapView must render dedicated compact Mobile HP header');
assert(fleetViewCode.includes('id="fleet-mobile-view-switcher"'), 'FleetMapView must render Mobile Segmented View Switcher (Peta | Daftar)');
assert(fleetViewCode.includes('id="btn-mobile-mode-map"') && fleetViewCode.includes('id="btn-mobile-mode-list"'), 'FleetMapView must include Peta and Daftar mode buttons on mobile');
assert(fleetViewCode.includes('id="fleet-mobile-bottom-card"'), 'FleetMapView must render thumb-friendly Mobile Docked Bottom Preview Card');
assert(fleetViewCode.includes('id="fleet-mobile-priority-list"'), 'FleetMapView must render Mobile Priority Card List in Daftar mode');
assert(fleetViewCode.includes('showMobileFilterSheet'), 'FleetMapView must include Mobile Filter Bottom Sheet');

// 5. Light-Mode CSS Hygiene Check (No bare text-white without btn-text or badge-text)
const classRegex = /class="([^"]*)"/g;
let match;
while ((match = classRegex.exec(fleetViewCode)) !== null) {
  const cls = match[1];
  const tokens = cls.split(/\s+/);
  if (tokens.includes('text-white')) {
    assert(
      tokens.includes('btn-text') || tokens.includes('badge-text'),
      `FleetMapView Light-Mode violation: class="${cls}" uses bare text-white without btn-text or badge-text`
    );
  }
}

console.log('✅ ALL SUPERADMIN FLEET & TENANT MAP TESTS PASSED (Desktop Ultra-Compact + Mobile HP Mode + RBAC + Light/Dark Hygiene)!');

