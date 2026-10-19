/**
 * Automated Verification Suite: Sidebar Auto-Hide with Setup Icon
 * 
 * Verifies:
 * 1. Store state and mutations for sidebarAutoHide persistence
 * 2. Sidebar brand header setup / pin icon toggle button
 * 3. App.vue edge hover trigger strip and slide-in drawer layout
 * 4. Header.vue sidebar auto-hide icon button and reactive toggle
 * 5. OperatorCommandCenterView.vue toolbar sidebar auto-hide icon button
 * 6. Localization keys in English and Indonesian
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const storePath = path.join(__dirname, '../src/store/index.js');
const storeContent = fs.readFileSync(storePath, 'utf8');

const sidebarPath = path.join(__dirname, '../src/components/Sidebar.vue');
const sidebarContent = fs.readFileSync(sidebarPath, 'utf8');

const appPath = path.join(__dirname, '../src/App.vue');
const appContent = fs.readFileSync(appPath, 'utf8');

const headerPath = path.join(__dirname, '../src/components/Header.vue');
const headerContent = fs.readFileSync(headerPath, 'utf8');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const enPath = path.join(__dirname, '../src/i18n/locales/en.js');
const enContent = fs.readFileSync(enPath, 'utf8');

const idPath = path.join(__dirname, '../src/i18n/locales/id.js');
const idContent = fs.readFileSync(idPath, 'utf8');

let totalTests = 0;
let passedTests = 0;

function runTest(name, fn) {
  totalTests++;
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
    passedTests++;
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
  }
}

console.log('===============================================================');
console.log('RUNNING TEST SUITE: SIDEBAR AUTO-HIDE VIA SETUP ICON');
console.log('===============================================================');

// 1. Vuex Store State & Persistence
runTest('1.1: Store initializes sidebarAutoHide in uiPreferences with localStorage persistence', () => {
  assert(storeContent.includes("sidebarAutoHide: localStorage.getItem('datalogger_sidebar_autohide') === 'true'"), 'Store must read sidebarAutoHide from localStorage');
  assert(storeContent.includes('SET_SIDEBAR_AUTOHIDE'), 'Store must define SET_SIDEBAR_AUTOHIDE mutation');
  assert(storeContent.includes('TOGGLE_SIDEBAR_AUTOHIDE'), 'Store must define TOGGLE_SIDEBAR_AUTOHIDE mutation');
  assert(storeContent.includes("localStorage.setItem('datalogger_sidebar_autohide'"), 'Store must persist to datalogger_sidebar_autohide');
  assert(storeContent.includes('setSidebarAutoHide'), 'Store must have setSidebarAutoHide action');
  assert(storeContent.includes('toggleSidebarAutoHide'), 'Store must have toggleSidebarAutoHide action');
});

// 2. Sidebar.vue Setup Icon & Event Handling
runTest('2.1: Sidebar.vue provides setup / pin button in brand header with toggle-autohide', () => {
  assert(sidebarContent.includes('@click="toggleAutoHide"'), 'Sidebar header must have toggleAutoHide button');
  assert(sidebarContent.includes('toggleSidebarAutoHide'), 'Sidebar method must dispatch toggleSidebarAutoHide');
  assert(sidebarContent.includes("this.$emit('toggle-autohide')"), 'Sidebar method must emit toggle-autohide');
  assert(sidebarContent.includes('isCollapsed'), 'Sidebar must preserve isCollapsed prop compatibility');
  assert(sidebarContent.includes('sidebarAutoHide'), 'Sidebar must have sidebarAutoHide computed property');
});

// 3. App.vue Auto-Hide Layout, Hover Trigger & Sliding Drawer
runTest('3.1: App.vue provides normal pinned sidebar and auto-hide edge trigger drawer', () => {
  assert(appContent.includes('v-if="!sidebarAutoHide"'), 'App.vue must render normal pinned sidebar when auto-hide is off');
  assert(appContent.includes('v-if="sidebarAutoHide"'), 'App.vue must render edge hover trigger when auto-hide is on');
  assert(appContent.includes('@mouseenter="openSidebarHover"'), 'App.vue must open drawer on edge hover');
  assert(appContent.includes('sidebar-slide'), 'App.vue must have sidebar-slide transition for drawer');
  assert(appContent.includes('@mouseleave="scheduleCloseSidebarHover"'), 'App.vue must close drawer on mouse leave');
  assert(appContent.includes('.sidebar-slide-enter'), 'App.vue must define sidebar-slide CSS animation');
});

// 4. Header.vue Setup Icon
runTest('4.1: Header.vue includes Sidebar Auto-Hide setup icon button and toggle handler', () => {
  assert(headerContent.includes('id="btn-sidebar-autohide-toggle"'), 'Header must contain btn-sidebar-autohide-toggle');
  assert(headerContent.includes('@click="toggleSidebarAutoHide"'), 'Header button must call toggleSidebarAutoHide');
  assert(headerContent.includes('sidebarAutoHide()'), 'Header must provide sidebarAutoHide computed property');
  assert(headerContent.includes('toggleSidebarAutoHide()'), 'Header must provide toggleSidebarAutoHide method');
});

// 5. OperatorCommandCenterView.vue Setup Icon
runTest('5.1: OperatorCommandCenterView.vue provides Sidebar Auto-Hide setup button in toolbar', () => {
  assert(occContent.includes('id="btn-operator-sidebar-autohide"'), 'Operator view must contain btn-operator-sidebar-autohide');
  assert(occContent.includes('@click="toggleSidebarAutoHide"'), 'Operator button must call toggleSidebarAutoHide');
  assert(occContent.includes('sidebarAutoHide()'), 'Operator view must provide sidebarAutoHide computed property');
  assert(occContent.includes('toggleSidebarAutoHide()'), 'Operator view must provide toggleSidebarAutoHide method');
});

// 6. Localization
runTest('6.1: English and Indonesian localization files define sidebar auto-hide and edgeSystem keys', () => {
  assert(enContent.includes('"sidebarAutoHideOn"'), 'en.js must define sidebarAutoHideOn');
  assert(enContent.includes('"sidebarAutoHideOff"'), 'en.js must define sidebarAutoHideOff');
  assert(enContent.includes('"sidebarHoverTitle"'), 'en.js must define sidebarHoverTitle');
  assert(enContent.includes('"edgeSystem"'), 'en.js must define edgeSystem');

  assert(idContent.includes('"sidebarAutoHideOn"'), 'id.js must define sidebarAutoHideOn');
  assert(idContent.includes('"sidebarAutoHideOff"'), 'id.js must define sidebarAutoHideOff');
  assert(idContent.includes('"sidebarHoverTitle"'), 'id.js must define sidebarHoverTitle');
  assert(idContent.includes('"edgeSystem"'), 'id.js must define edgeSystem');
});

console.log('===============================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${totalTests - passedTests} FAILED (TOTAL: ${totalTests})`);
console.log('===============================================================');

if (passedTests !== totalTests) {
  process.exit(1);
}
