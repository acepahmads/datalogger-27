const fs = require('fs');
const path = require('path');
const assert = require('assert');

const webSrc = path.join(__dirname, '..', 'src');
const viewCode = fs.readFileSync(path.join(webSrc, 'views', 'InboundDataView.vue'), 'utf8');
const routerCode = fs.readFileSync(path.join(webSrc, 'router', 'index.js'), 'utf8');
const sidebarCode = fs.readFileSync(path.join(webSrc, 'components', 'Sidebar.vue'), 'utf8');

// 1. Router registers /monitoring/inbound-data strictly for SUPERADMIN + Cloud Mode
assert(routerCode.includes("path: '/monitoring/inbound-data'"), 'Router must register /monitoring/inbound-data');
assert(routerCode.includes("meta: { roles: ['SUPERADMIN'], cloudOnly: true }"), 'InboundData route must enforce SUPERADMIN + cloudOnly');
assert(routerCode.includes('if (to.meta && to.meta.cloudOnly && !isCloudMode)'), 'Router guard must enforce cloudOnly restriction');

// 2. Sidebar shows Inbound Data link strictly for SUPERADMIN in Cloud Mode
assert(sidebarCode.includes('id="nav-inbound-data"'), 'Sidebar must expose #nav-inbound-data link');
assert(/v-if="isSuperAdmin && isCloudMode"\s+id="nav-inbound-data"/.test(sidebarCode), 'Sidebar Inbound Data link must be gated by isSuperAdmin && isCloudMode');

// 3. InboundDataView enforces role/mode guard + compact filters (uses Header tenant switcher, no duplicate tenant/domain dropdowns) + 3-mode Raw JSON inspector
assert(viewCode.includes('if (!this.isSuperAdmin || !this.isCloudMode)'), 'InboundDataView must redirect non-Superadmin or Edge Mode users');
assert(viewCode.includes('/api/superadmin/inbound-logs'), 'InboundDataView must fetch from /api/superadmin/inbound-logs');
assert(!viewCode.includes('id="filter-inbound-tenant"'), 'InboundDataView must NOT render duplicate Tenant dropdown (uses top Header tenant selector)');
assert(!viewCode.includes('id="filter-inbound-logger"'), 'InboundDataView must NOT render Domain (AQMS/SPARING) dropdown');
assert(viewCode.includes('this.activeTenantId > 0'), 'InboundDataView must filter automatically by Header activeTenantId');
assert(viewCode.includes('id="filter-inbound-status"'), 'InboundDataView must include Status filter');
assert(viewCode.includes('id="filter-inbound-format"'), 'InboundDataView must include Format filter');
assert(viewCode.includes('id="filter-inbound-source"'), 'InboundDataView must include Source filter');
assert(viewCode.includes('id="filter-inbound-search"'), 'InboundDataView must include Search input');
assert(viewCode.includes('id="btn-inbound-view-pretty"') && viewCode.includes('id="btn-inbound-view-human"') && viewCode.includes('id="btn-inbound-view-raw"'), 'InboundDataView must provide Pretty / Waktu Terbaca / Raw JSON viewer switch');

// 4. Strict Light-Mode CSS Hygiene
const classRegex = /class="([^"]*)"/g;
let match;
while ((match = classRegex.exec(viewCode)) !== null) {
  const cls = match[1];
  const tokens = cls.split(/\s+/);
  if (tokens.includes('text-white')) {
    assert(
      tokens.includes('btn-text') || tokens.includes('badge-text'),
      `InboundDataView Light-Mode violation: class="${cls}" uses bare text-white without btn-text or badge-text`
    );
  }
  assert(
    !tokens.includes('bg-slate-900'),
    `InboundDataView Light-Mode violation: class="${cls}" uses bare bg-slate-900 without dark: prefix`
  );
}

console.log('PASS: Cloud Mode Superadmin Inbound Data Monitor (RBAC + Multi-Tenant Filters + Raw JSON Inspector + Light-Mode Hygiene) verified!');
