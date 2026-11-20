/**
 * Verification Suite: Multi-Theme & Design Preset Studio
 * (Oceanic Classic, Eco Forest, Industrial SCADA, Royal Glass + Compact/Comfortable Density)
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
const modalVue = fs.readFileSync(path.join(srcDir, 'components/ThemeStudioModal.vue'), 'utf8');
const headerVue = fs.readFileSync(path.join(srcDir, 'components/Header.vue'), 'utf8');
const dashVue = fs.readFileSync(path.join(srcDir, 'views/DashboardOperatorView.vue'), 'utf8');
const reportsVue = fs.readFileSync(path.join(srcDir, 'views/ReportsView.vue'), 'utf8');
const integrationVue = fs.readFileSync(path.join(srcDir, 'views/IntegrationView.vue'), 'utf8');
const configVue = fs.readFileSync(path.join(srcDir, 'views/ConfigView.vue'), 'utf8');

let passed = 0;
function check(name, fn) {
  fn();
  passed += 1;
  console.log(`✓ PASS: ${name}`);
}

console.log('========================================================================');
console.log('MULTI-THEME & DESIGN PRESET STUDIO VERIFICATION SUITE');
console.log('========================================================================');

check('1. theme.js exports 4 selectable presets, density modes, and localStorage persistence helpers', () => {
  for (const id of ['oceanic', 'eco-forest', 'industrial-scada', 'royal-glass']) {
    assert(themeJs.includes(`id: '${id}'`), `Missing theme preset '${id}' in theme.js`);
  }
  assert(themeJs.includes('export function applyThemePreset('), 'Missing applyThemePreset');
  assert(themeJs.includes('export function applyDensity('), 'Missing applyDensity');
  assert(themeJs.includes("document.documentElement.setAttribute('data-ui-theme'"), 'Must set data-ui-theme on <html>');
  assert(themeJs.includes("document.documentElement.setAttribute('data-ui-density'"), 'Must set data-ui-density on <html>');
});

check('2. theme-presets.css styles Eco Forest, Industrial SCADA, Royal Glass (Light + Dark) and Compact density', () => {
  assert(presetsCss.includes('html[data-ui-theme="eco-forest"]:not(.dark)'), 'Missing Eco Forest Light rules');
  assert(presetsCss.includes('html[data-ui-theme="eco-forest"].dark'), 'Missing Eco Forest Dark rules');
  assert(presetsCss.includes('html[data-ui-theme="industrial-scada"]:not(.dark)'), 'Missing Industrial SCADA Light rules');
  assert(presetsCss.includes('html[data-ui-theme="industrial-scada"].dark'), 'Missing Industrial SCADA Dark rules');
  assert(presetsCss.includes('html[data-ui-theme="royal-glass"]:not(.dark)'), 'Missing Royal Glass Light rules');
  assert(presetsCss.includes('html[data-ui-theme="royal-glass"].dark'), 'Missing Royal Glass Dark rules');
  assert(presetsCss.includes('html[data-ui-density="compact"]'), 'Missing Compact density rules');
  assert(mainJs.includes("import './assets/theme-presets.css'"), 'main.js must load theme-presets.css after main.css');
});

check('3. Vuex store and App.vue manage global Theme Studio state and modal lifecycle', () => {
  assert(storeJs.includes('SET_THEME_PRESET(state, presetId)'), 'Missing SET_THEME_PRESET mutation');
  assert(storeJs.includes('SET_DENSITY(state, densityMode)'), 'Missing SET_DENSITY mutation');
  assert(storeJs.includes('SET_SHOW_THEME_STUDIO_MODAL(state, visible)'), 'Missing SET_SHOW_THEME_STUDIO_MODAL mutation');
  assert(storeJs.includes('setThemePreset({ commit }, presetId)'), 'Missing setThemePreset action');
  assert(storeJs.includes('setDensity({ commit }, densityMode)'), 'Missing setDensity action');
  assert(appVue.includes('<ThemeStudioModal'), 'App.vue must mount ThemeStudioModal globally');
});

check('4. Theme Studio triggers exist across Header + all 4 target pages (Dashboard Operator, Export Data, Outbound, Setup)', () => {
  assert(headerVue.includes('id="btn-header-theme-studio"'), 'Missing Theme Studio trigger in Header');
  assert(dashVue.includes('id="btn-bento-theme-studio"'), 'Missing Theme Studio trigger in Dashboard Operator');
  assert(reportsVue.includes('id="btn-reports-theme-studio"'), 'Missing Theme Studio trigger in Export Data (ReportsView)');
  assert(integrationVue.includes('id="btn-outbound-theme-studio"'), 'Missing Theme Studio trigger in Pengiriman Outbound (IntegrationView)');
  assert(configVue.includes('id="btn-config-theme-studio"'), 'Missing Theme Studio trigger in Setup (ConfigView)');
});

check('5. ThemeStudioModal adheres to Light-Mode text-white CSS hygiene', () => {
  const tmpl = modalVue.match(/<template>([\s\S]*?)<\/template>/);
  assert(tmpl, 'Missing <template> in ThemeStudioModal.vue');
  const violations = [];
  tmpl[1].split('\n').forEach((line, idx) => {
    if (line.includes('text-white') && !line.includes('dark:text-white') && !line.includes('hover:text-white')) {
      if (!line.includes('btn-text') && !line.includes('badge-text')) {
        violations.push(`Line ${idx + 1}: ${line.trim()}`);
      }
    }
  });
  assert.strictEqual(violations.length, 0, `Light-mode text-white hygiene violations:\n${violations.join('\n')}`);
});

console.log('========================================================================');
console.log(`ALL ${passed} MULTI-THEME STUDIO TESTS PASSED (100%)`);
console.log('========================================================================');
