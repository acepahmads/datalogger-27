/**
 * Acceptance test suite for Global Visibility Config of Operator Console & Development Menu:
 * - Default hidden (show_operator_console: false, show_development_menu: false)
 * - Configurable globally via System Configuration -> Server & Runtime & config.json
 * - Protected files remain untouched
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const rootDir = path.resolve(__dirname, '../..');
const srcDir = path.resolve(__dirname, '../src');

let passed = 0;
function check(name, fn) {
  fn();
  passed++;
  console.log(`✓ PASS: ${name}`);
}

check('1. Protected files remain untouched', () => {
  const diff = execSync(
    'git diff -- web/src/views/OperatorCommandCenterView.vue web/src/components/ISPUIndicatorBar.vue web/src/components/OperatorDashboardIcon.vue web/src/views/DashboardOperatorView.vue',
    { cwd: rootDir, encoding: 'utf8' }
  );
  assert.strictEqual(diff.trim(), '', 'Protected files must not be modified');
});

check('2. config.json & Vuex store default show_operator_console and show_development_menu to false', () => {
  const cfgJson = JSON.parse(fs.readFileSync(path.join(rootDir, 'config.json'), 'utf8'));
  assert.strictEqual(cfgJson.show_operator_console, false, 'config.json must default show_operator_console to false');
  assert.strictEqual(cfgJson.show_development_menu, false, 'config.json must default show_development_menu to false');

  const storeSrc = fs.readFileSync(path.join(srcDir, 'store/index.js'), 'utf8');
  assert(storeSrc.includes('show_operator_console: false'), 'store must default show_operator_console to false');
  assert(storeSrc.includes('show_development_menu: false'), 'store must default show_development_menu to false');
  assert(storeSrc.includes('canAccessDev: (state, getters) => getters.isSuperAdmin && getters.showDevelopmentMenu'), 'canAccessDev must check showDevelopmentMenu');
});

check('3. Sidebar & Header gate Operator Console and Development Menu behind global config', () => {
  const sidebarSrc = fs.readFileSync(path.join(srcDir, 'components/Sidebar.vue'), 'utf8');
  assert(sidebarSrc.includes('v-if="isSuperAdmin && showDevelopmentMenuConfig"'), 'Sidebar must gate Development menu with showDevelopmentMenuConfig');
  assert(sidebarSrc.includes('v-if="showOperatorConsoleConfig && !isOperator'), 'Sidebar must gate Operator Console with showOperatorConsoleConfig');

  const headerSrc = fs.readFileSync(path.join(srcDir, 'components/Header.vue'), 'utf8');
  assert(headerSrc.includes('v-if="showOperatorConsoleConfig && !isOperator'), 'Header quick button must gate Operator Console with showOperatorConsoleConfig');
});

check('4. Router guard redirects hidden Operator Console to Dashboard Operator and hidden Development to Fleet Map', () => {
  const routerSrc = fs.readFileSync(path.join(srcDir, 'router/index.js'), 'utf8');
  assert(routerSrc.includes('const showOperatorConsole = Boolean(store.state.systemStatus && store.state.systemStatus.show_operator_console === true);'), 'Router must read show_operator_console');
  assert(routerSrc.includes('const showDevelopmentMenu = Boolean(store.state.systemStatus && store.state.systemStatus.show_development_menu === true);'), 'Router must read show_development_menu');
  assert(routerSrc.includes('const usesDashboardOperatorLanding = !showOperatorConsole ||'), 'Router must redirect to /monitoring/dashboard-operator when showOperatorConsole is false');
  assert(routerSrc.includes("if (isAuth && isSuper && !showDevelopmentMenu && (to.path === '/development' || to.path.startsWith('/development/')))"), 'Router must redirect Superadmin away from /development when hidden');
});

check('5. ConfigView exposes global visibility checkboxes and syncs Vuex systemStatus', () => {
  const cfgViewSrc = fs.readFileSync(path.join(srcDir, 'views/ConfigView.vue'), 'utf8');
  assert(cfgViewSrc.includes('id="cfg-show-operator-console"') && cfgViewSrc.includes('v-model="form.show_operator_console"'), 'ConfigView must include Operator Console toggle');
  assert(cfgViewSrc.includes('id="cfg-show-development-menu"') && cfgViewSrc.includes('v-model="form.show_development_menu"'), 'ConfigView must include Development Menu toggle');
});

console.log(`\nALL ${passed} GLOBAL MENU VISIBILITY TESTS PASSED!`);

