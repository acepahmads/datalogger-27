/**
 * Acceptance test suite for:
 * 1. Default landing page is Operator Dashboard (/monitoring/operator).
 * 2. Unauthenticated visits automatically log in as default operator without showing login screen.
 * 3. Operator Dashboard defaults to full screen (sidebar collapsed & top panel hidden).
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

console.log('========================================================================');
console.log('RUNNING TEST SUITE: OPERATOR AUTO-LOGIN & FULLSCREEN LANDING');
console.log('========================================================================');

const routerFile = fs.readFileSync(path.join(__dirname, '../src/router/index.js'), 'utf8');
const storeFile = fs.readFileSync(path.join(__dirname, '../src/store/index.js'), 'utf8');
const appFile = fs.readFileSync(path.join(__dirname, '../src/App.vue'), 'utf8');
const mainFile = fs.readFileSync(path.join(__dirname, '../src/main.js'), 'utf8');

// 1. Verify root '/' and '*' redirect to '/monitoring/operator'
assert(
  routerFile.includes("{ path: '/', redirect: '/monitoring/operator' }"),
  'Root path / must redirect to /monitoring/operator'
);
console.log("✓ PASS: 1.1: Root path '/' redirects to '/monitoring/operator'");

assert(
  routerFile.includes("{ path: '*', redirect: '/monitoring/operator' }"),
  'Fallback wildcard path must redirect to /monitoring/operator'
);
console.log("✓ PASS: 1.2: Fallback wildcard path '*' redirects to '/monitoring/operator'");

// 2. Verify autoLoginOperator in store
assert(
  storeFile.includes('async autoLoginOperator({ commit, dispatch })'),
  'Store must provide autoLoginOperator action'
);
assert(
  storeFile.includes("username: 'operator_user'") && storeFile.includes("password: 'Password123!'"),
  'autoLoginOperator must use seeded default operator credentials'
);
console.log("✓ PASS: 2.1: Vuex store provides autoLoginOperator action with default operator credentials");

// 3. Verify router navigation guard automatically authenticates operator without showing login
assert(
  routerFile.includes("await store.dispatch('autoLoginOperator')"),
  'Navigation guard must call autoLoginOperator when unauthenticated'
);
assert(
  routerFile.includes("redirect: '/monitoring/operator'"),
  'Navigation guard must redirect to /monitoring/operator'
);
console.log("✓ PASS: 2.2: Router navigation guard bypasses manual login and auto-authenticates operator");

// 4. Verify main.js initializes auto-login on startup
assert(
  mainFile.includes("this.$store.dispatch('autoLoginOperator')"),
  'main.js must trigger autoLoginOperator when no token exists on startup'
);
console.log("✓ PASS: 2.3: main.js auto-initiates operator session and telemetry fetchers on boot");

// 5. Verify App.vue defaults to fullscreen for operator dashboard
assert(
  appFile.includes('isSidebarCollapsed: true') && appFile.includes('isTopPanelHidden: true'),
  'App.vue data must initialize isSidebarCollapsed: true and isTopPanelHidden: true'
);
assert(
  appFile.includes("to.path === '/monitoring/operator' || to.path === '/operator'") &&
  appFile.includes('this.isSidebarCollapsed = true') &&
  appFile.includes('this.isTopPanelHidden = true'),
  'App.vue route watcher must enforce fullscreen for operator dashboard'
);
console.log("✓ PASS: 3.1: Operator Dashboard defaults to full screen (sidebar collapsed & top panel hidden)");

console.log('========================================================================');
console.log('ALL AUTO-LOGIN & FULLSCREEN LANDING TESTS PASSED (100%)');
console.log('========================================================================');
