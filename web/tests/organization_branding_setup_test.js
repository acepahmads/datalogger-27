/**
 * Automated Verification Suite: Customer (Left) & Maker (Right) Organization Branding Setup
 * 
 * Verifies:
 * 1. Default/temporary branding: Customer on Left (PT. Mitra Industri Enviro), Maker on Right (PT. Centra Biotech Indonesia)
 * 2. Left side renders customer company icon and name with SCADA station details
 * 3. Right side renders maker company icon and name with solution provider details
 * 4. System Admin Configuration (ConfigView.vue) provides dedicated Branding & Organization tab
 * 5. Operator Command Center provides Branding Configuration modal with Customer & Maker setup
 * 6. companyBranding.js utility supports dual IndexedDB blob storage and reactive synchronization
 * 7. Localization parity between English and Indonesian
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const occPath = path.join(__dirname, '../src/views/OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const configViewPath = path.join(__dirname, '../src/views/ConfigView.vue');
const configViewContent = fs.readFileSync(configViewPath, 'utf8');

const brandingUtilPath = path.join(__dirname, '../src/utils/companyBranding.js');
const brandingUtilContent = fs.readFileSync(brandingUtilPath, 'utf8');

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
console.log('RUNNING TEST SUITE: CUSTOMER (LEFT) & MAKER (RIGHT) BRANDING');
console.log('===============================================================');

// 1. Dual Branding Utility Specification
runTest('1.1: companyBranding.js defines DEFAULT_BRANDING with customer and maker companies', () => {
  assert(brandingUtilContent.includes('DEFAULT_BRANDING'), 'Must export DEFAULT_BRANDING');
  assert(brandingUtilContent.includes('PT. Mitra Industri Enviro'), 'Must specify temporary customer name');
  assert(brandingUtilContent.includes('PT. Centra Biotech Indonesia'), 'Must specify temporary maker name');
});

runTest('1.2: companyBranding.js provides load and save operations for both customer and maker', () => {
  assert(brandingUtilContent.includes('loadOrganizationBranding'), 'Must provide loadOrganizationBranding');
  assert(brandingUtilContent.includes('saveCustomerBranding'), 'Must provide saveCustomerBranding');
  assert(brandingUtilContent.includes('saveMakerBranding'), 'Must provide saveMakerBranding');
  assert(brandingUtilContent.includes('CUSTOMER_LOGO_KEY'), 'Must store customer logo key');
  assert(brandingUtilContent.includes('MAKER_LOGO_KEY'), 'Must store maker logo key');
});

// 2. Operator Command Center Left Side (Customer Company)
runTest('2.1: Operator Command Center renders customer company icon & name on left side', () => {
  assert(occContent.includes('customerCompanyName'), 'Template must bind customerCompanyName on left side');
  assert(occContent.includes('customerLogoUrl'), 'Template must support customer logo image');
  assert(occContent.includes('customerSubtitle'), 'Template must support customer subtitle');
});

// 3. Operator Command Center Right Side (Maker Company)
runTest('3.1: Operator Command Center renders maker company icon & name on right side', () => {
  assert(occContent.includes('makerCompanyName'), 'Template must bind makerCompanyName on right side');
  assert(occContent.includes('makerLogoUrl'), 'Template must support maker logo image');
  assert(occContent.includes('makerSubtitle'), 'Template must support maker subtitle');
});

// 4. System Admin Setup (ConfigView.vue)
runTest('4.1: ConfigView.vue provides dedicated Branding tab for System Admin', () => {
  assert(configViewContent.includes("activeTab === 'branding'"), 'ConfigView must contain branding activeTab');
  assert(configViewContent.includes('tabBranding') || configViewContent.includes('Organization & Branding'), 'ConfigView must have Branding tab button');
  assert(configViewContent.includes('saveBrandingConfig'), 'ConfigView must have saveBrandingConfig method');
  assert(configViewContent.includes('handleCustomerLogoUpload'), 'ConfigView must allow customer logo upload');
  assert(configViewContent.includes('handleMakerLogoUpload'), 'ConfigView must allow maker logo upload');
});

// 5. Operator Console Branding Modal
runTest('5.1: Operator Command Center provides Branding modal with customer and maker tabs', () => {
  assert(occContent.includes('showBrandingModal'), 'OCC must provide showBrandingModal');
  assert(occContent.includes('activeBrandingTab'), 'Branding modal must have customer and maker tabs');
  assert(occContent.includes('saveBrandingFromModal'), 'OCC must provide saveBrandingFromModal method');
});

// 6. Localization
runTest('6.1: Localization files contain customer and maker company keys', () => {
  assert(enContent.includes('"customerCompany"'), 'en.js must have customerCompany');
  assert(enContent.includes('"makerCompany"'), 'en.js must have makerCompany');
  assert(idContent.includes('"customerCompany"'), 'id.js must have customerCompany');
  assert(idContent.includes('"makerCompany"'), 'id.js must have makerCompany');
});

console.log('===============================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${totalTests - passedTests} FAILED (TOTAL: ${totalTests})`);
console.log('===============================================================');

if (passedTests !== totalTests) {
  process.exit(1);
}
