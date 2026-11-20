/**
 * Verification Suite: Centralized Global Multi-Theme & Multi-Layout Configuration
 * (Stored globally on server in config.json + configured from System Configuration > Tampilan & Desain)
 * Run via: node web/tests/multi_theme_studio_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const srcDir = path.join(__dirname, '../src');
const themeJs = fs.readFileSync(path.join(srcDir, 'utils/theme.js'), 'utf8');
const presetsCss = fs.readFileSync(path.join(srcDir, 'assets/theme-presets.css'), 'utf8');
const mainJs = fs.readFileSync(path.join(srcDir, 'main.js'), 'utf8');
const storeJs = fs.readFileSync(path.join(srcDir, 'store/index.js'), 'utf8');
const appVue = fs.readFileSync(path.join(srcDir, 'App.vue'), 'utf8');
const headerVue = fs.readFileSync(path.join(srcDir, 'components/Header.vue'), 'utf8');
const dashVue = fs.readFileSync(path.join(srcDir, 'views/DashboardOperatorView.vue'), 'utf8');
const reportsVue = fs.readFileSync(path.join(srcDir, 'views/ReportsView.vue'), 'utf8');
const integrationVue = fs.readFileSync(path.join(srcDir, 'views/IntegrationView.vue'), 'utf8');
const configVue = fs.readFileSync(path.join(srcDir, 'views/ConfigView.vue'), 'utf8');
const goConfig = fs.readFileSync(path.join(__dirname, '../../internal/config/config.go'), 'utf8');
const goSysSvc = fs.readFileSync(path.join(__dirname, '../../internal/service/system_service.go'), 'utf8');

let passed = 0;
function check(name, fn) {
  fn();
  passed += 1;
  console.log(`✓ PASS: ${name}`);
}

console.log('========================================================================');
console.log('GLOBAL MULTI-THEME & MULTI-LAYOUT STUDIO VERIFICATION SUITE');
console.log('========================================================================');

check('1. Backend persists global UIThemePreset, UILayout, UIDensity & UILightingMode in config.go and broadcasts via system_service.go', () => {
  assert(goConfig.includes('UIThemePreset       string `json:"ui_theme_preset"`'), 'Missing UIThemePreset in config.go');
  assert(goConfig.includes('UILayout            string `json:"ui_layout"`'), 'Missing UILayout in config.go');
  assert(goConfig.includes('UIDensity           string `json:"ui_density"`'), 'Missing UIDensity in config.go');
  assert(goConfig.includes('UILightingMode      string `json:"ui_lighting_mode"`'), 'Missing UILightingMode in config.go');
  assert(goSysSvc.includes('normalizeUIThemePreset') && goSysSvc.includes('normalizeUILayout'), 'Missing global appearance normalizers in system_service.go');
});

check('2. theme.js & theme-presets.css support 5 distinct themes (Light + Dark), Bento --b-primary overrides, 3 layouts & compact density', () => {
  for (const id of ['oceanic', 'eco-forest', 'industrial-scada', 'royal-glass', 'warm-terracotta']) {
    assert(themeJs.includes(`id: '${id}'`), `Missing theme preset '${id}'`);
  }
  for (const id of ['classic', 'split', 'executive']) {
    assert(themeJs.includes(`id: '${id}'`), `Missing layout preset '${id}'`);
  }
  assert(presetsCss.includes('html[data-ui-theme="eco-forest"]:not(.dark)'), 'Missing Eco Forest Light rules');
  assert(presetsCss.includes('html[data-ui-theme="industrial-scada"].dark'), 'Missing Industrial SCADA Dark rules');
  assert(presetsCss.includes('html[data-ui-theme="royal-glass"]:not(.dark)'), 'Missing Royal Glass Light rules');
  assert(presetsCss.includes('html[data-ui-theme="warm-terracotta"]:not(.dark)'), 'Missing Warm Terracotta Light rules');
  assert(presetsCss.includes('--b-primary: #0F766E') && presetsCss.includes('--b-primary: #C2410C') && presetsCss.includes('--b-primary: #4F46E5'), 'Themes must override Dashboard Operator --b-primary tokens');
  assert(presetsCss.includes('html[data-ui-layout="split"] .bento-dashboard-page'), 'Missing Split-Screen Dashboard Operator CSS');
  assert(presetsCss.includes('html[data-ui-layout="executive"] .bento-dashboard-page'), 'Missing Executive Spotlight Dashboard Operator CSS');
  assert(mainJs.includes("import './assets/theme-presets.css'"), 'main.js must load theme-presets.css');
});

check('3. Vuex store syncs global server appearance from /api/system/status and saves via PUT /api/system/config', () => {
  assert(storeJs.includes('async saveGlobalAppearance('), 'Missing saveGlobalAppearance action');
  assert(storeJs.includes('axios.put(`${API_BASE}/system/config`, payload)'), 'saveGlobalAppearance must persist to /api/system/config');
  assert(storeJs.includes('data.ui_theme_preset') && storeJs.includes('data.ui_layout'), 'fetchSystemStatus must sync server global appearance');
});

check('4. Setup is centralized in ConfigView (Tab Tampilan & Desain Global) — scattered per-page theme buttons removed', () => {
  assert(configVue.includes('id="tab-btn-global-appearance"'), 'ConfigView must expose global appearance tab');
  assert(configVue.includes('id="tab-panel-global-appearance"'), 'ConfigView must render global appearance panel');
  assert(configVue.includes('id="btn-global-appearance-save"'), 'ConfigView must have Save Global Appearance button');
  assert(!headerVue.includes('id="btn-header-theme-studio"'), 'Header must not duplicate per-page theme popup button');
  assert(!dashVue.includes('id="btn-bento-theme-studio"'), 'Dashboard Operator must not duplicate per-page theme popup button');
  assert(!reportsVue.includes('id="btn-reports-theme-studio"'), 'ReportsView must not duplicate per-page theme popup button');
  assert(!integrationVue.includes('id="btn-outbound-theme-studio"'), 'IntegrationView must not duplicate per-page theme popup button');
});

check('5. Multi-Layout structural views remain wired in App, ReportsView, IntegrationView & ConfigView', () => {
  assert(appVue.includes('id="executive-top-nav-pill"'), 'App.vue must render Executive Top-Nav Pill Bar');
  assert(reportsVue.includes('id="reports-executive-quick-gallery"') && reportsVue.includes('reports-split-filter-rail'), 'ReportsView must support Split Rail & Executive Quick Gallery');
  assert(integrationVue.includes('id="outbound-split-inspector-dock"') && integrationVue.includes('id="outbound-executive-health-card"'), 'IntegrationView must support Split Inspector & Executive SLA Card');
  assert(configVue.includes('id="config-split-vertical-rail"'), 'ConfigView must support Split Vertical Section Rail');
});

console.log('========================================================================');
console.log(`ALL ${passed} GLOBAL MULTI-THEME & LAYOUT TESTS PASSED (100%)`);
console.log('========================================================================');
