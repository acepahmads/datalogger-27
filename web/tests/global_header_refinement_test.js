/**
 * Automated Verification Suite: Global Main Header Refinement
 * 
 * Verifies:
 * 1. Information and status strings (station, os/arch, app health, dev comm health, time) are removed from the visible top main panel.
 * 2. Operator Console button is converted to an icon-only button (no text label in button).
 * 3. Language button is converted to a switch icon with EN/ID slider toggle.
 * 4. Day/Night theme button is converted to a switch icon with Light/Dark slider toggle.
 * 5. Script methods toggleLanguage and toggleTheme are implemented.
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const headerPath = path.join(__dirname, '../src/components/Header.vue');
const headerContent = fs.readFileSync(headerPath, 'utf8');

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
console.log('RUNNING TEST SUITE: GLOBAL MAIN HEADER REFINEMENT');
console.log('===============================================================');

// 1. Center Status and Info Removal
runTest('1.1: Information and status (station, app status, comm health, time) removed from visible header template', () => {
  // Check template portion
  const templateMatch = headerContent.match(/<template>([\s\S]*?)<\/template>/);
  assert(templateMatch, 'Header must contain template block');
  const templateContent = templateMatch[1];

  assert(!templateContent.includes("$t('operator.station')"), 'Visible station label must be removed from header template');
  assert(!templateContent.includes("$t('operator.appStatus')"), 'Visible app status label must be removed from header template');
  assert(!templateContent.includes("$t('operator.commHealth')"), 'Visible comm health label must be removed from header template');
  assert(!templateContent.includes("$t('header.time')"), 'Visible time label must be removed from header template');
  assert(!templateContent.includes("systemStatus.os"), 'Visible os/arch info must be removed from header template');
});

// 2. Operator Console Icon Button
runTest('2.1: Operator Console button converted to icon-only button without text', () => {
  const templateMatch = headerContent.match(/<template>([\s\S]*?)<\/template>/);
  const templateContent = templateMatch[1];

  // Must link to /monitoring/operator
  assert(templateContent.includes('to="/monitoring/operator"'), 'Must maintain link to /monitoring/operator');
  // Must NOT contain text 'Operator Console' in visible button body
  assert(!templateContent.includes("{{ $t('navigation.operatorConsole') || 'Operator Console' }}</span>"), 'Must not have text span for Operator Console');
  // Must have svg icon
  const opBtnMatch = templateContent.match(/<router-link[^>]*to="\/monitoring\/operator"[\s\S]*?<\/router-link>/);
  assert(opBtnMatch, 'Operator console router link must exist');
  assert(opBtnMatch[0].includes('<svg'), 'Operator console button must render svg icon');
});

// 3. Language Switch Icon
runTest('3.1: Language control is a switch icon button with EN / ID toggle', () => {
  assert(headerContent.includes('toggleLanguage'), 'Header must implement toggleLanguage method');
  const templateMatch = headerContent.match(/<template>([\s\S]*?)<\/template>/);
  const templateContent = templateMatch[1];
  assert(templateContent.includes('@click="toggleLanguage"'), 'Language button must trigger toggleLanguage');
  assert(templateContent.includes('EN') && templateContent.includes('ID'), 'Language switch must contain EN and ID switch states');
});

// 4. Dark / Light Mode Switch Icon
runTest('4.1: Dark/Light theme control is a switch icon button', () => {
  assert(headerContent.includes('toggleTheme'), 'Header must implement toggleTheme method');
  assert(headerContent.includes('isDarkMode'), 'Header must calculate isDarkMode');
  const templateMatch = headerContent.match(/<template>([\s\S]*?)<\/template>/);
  const templateContent = templateMatch[1];
  assert(templateContent.includes('@click="toggleTheme"'), 'Theme button must trigger toggleTheme');
});

console.log('===============================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${totalTests - passedTests} FAILED (TOTAL: ${totalTests})`);
console.log('===============================================================');

if (passedTests !== totalTests) {
  process.exit(1);
}
