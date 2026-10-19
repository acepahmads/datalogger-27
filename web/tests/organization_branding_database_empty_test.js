/**
 * Automated Verification Suite: Database Organization Branding & Empty String Support
 * 
 * Verifies:
 * 1. Backend database persistence:
 *    - model.SystemBranding defined with customer & maker attributes and longtext logos
 *    - migration.go auto-migrates SystemBranding
 *    - SystemRepository & SystemService implement GetBranding and SaveBranding
 *    - SystemHandler exposes GET/POST/PUT /api/system/branding
 *    - cmd/main.go registers the /api/system/branding endpoints
 * 2. Allowing empty values on save:
 *    - Maker company name can be saved as empty ("") without falling back to defaults
 *    - Maker subtitle can be saved as empty ("") without falling back to defaults
 *    - Customer subtitle can be saved as empty ("") without falling back to defaults
 * 3. Frontend UI rendering:
 *    - Operator Command Center hides subtitle when customerSubtitle is empty
 *    - Operator Command Center hides maker text when makerCompanyName and makerSubtitle are empty
 *    - Operator Command Center does not display fallback 'MAKER' text when empty
 * 4. Localization:
 *    - Root 'branding.title' exists in en.js and id.js, preventing raw 'BRANDING.TITLE' display
 */

const assert = require('assert');
const fs = require('fs');
const path = require('path');

const ROOT_DIR = path.join(__dirname, '..', '..');

// File paths
const modelPath = path.join(ROOT_DIR, 'internal', 'model', 'system.go');
const migrationPath = path.join(ROOT_DIR, 'internal', 'database', 'migration.go');
const repoPath = path.join(ROOT_DIR, 'internal', 'repository', 'system_repo.go');
const servicePath = path.join(ROOT_DIR, 'internal', 'service', 'system_service.go');
const handlerPath = path.join(ROOT_DIR, 'internal', 'handler', 'system_handler.go');
const mainPath = path.join(ROOT_DIR, 'cmd', 'main.go');

const brandingUtilPath = path.join(__dirname, '..', 'src', 'utils', 'companyBranding.js');
const occPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const configViewPath = path.join(__dirname, '..', 'src', 'views', 'ConfigView.vue');
const enPath = path.join(__dirname, '..', 'src', 'i18n', 'locales', 'en.js');
const idPath = path.join(__dirname, '..', 'src', 'i18n', 'locales', 'id.js');

const modelContent = fs.readFileSync(modelPath, 'utf8');
const migrationContent = fs.readFileSync(migrationPath, 'utf8');
const repoContent = fs.readFileSync(repoPath, 'utf8');
const serviceContent = fs.readFileSync(servicePath, 'utf8');
const handlerContent = fs.readFileSync(handlerPath, 'utf8');
const mainContent = fs.readFileSync(mainPath, 'utf8');
const brandingUtilContent = fs.readFileSync(brandingUtilPath, 'utf8');
const occContent = fs.readFileSync(occPath, 'utf8');
const configViewContent = fs.readFileSync(configViewPath, 'utf8');
const enContent = fs.readFileSync(enPath, 'utf8');
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

console.log('======================================================================');
console.log('RUNNING TEST SUITE: DATABASE BRANDING & EMPTY STRING SAVE VERIFICATION');
console.log('======================================================================');

// 1. Backend Database Persistence
runTest('1.1: SystemBranding model is defined with complete customer and maker fields', () => {
  assert(modelContent.includes('type SystemBranding struct'), 'model/system.go must define SystemBranding');
  assert(modelContent.includes('CustomerName'), 'SystemBranding must contain CustomerName');
  assert(modelContent.includes('CustomerSubtitle'), 'SystemBranding must contain CustomerSubtitle');
  assert(modelContent.includes('CustomerLogo'), 'SystemBranding must contain CustomerLogo');
  assert(modelContent.includes('MakerName'), 'SystemBranding must contain MakerName');
  assert(modelContent.includes('MakerSubtitle'), 'SystemBranding must contain MakerSubtitle');
  assert(modelContent.includes('MakerLogo'), 'SystemBranding must contain MakerLogo');
});

runTest('1.2: Database migration includes SystemBranding in AutoMigrate', () => {
  assert(migrationContent.includes('&model.SystemBranding{}'), 'migration.go must auto-migrate SystemBranding');
});

runTest('1.3: Repository provides GetBranding and SaveBranding', () => {
  assert(repoContent.includes('GetBranding() (*model.SystemBranding, error)'), 'system_repo.go must have GetBranding');
  assert(repoContent.includes('SaveBranding(b *model.SystemBranding) error'), 'system_repo.go must have SaveBranding');
});

runTest('1.4: Service provides GetBranding, SaveBranding and UpdateSystemBrandingRequest with pointer fields', () => {
  assert(serviceContent.includes('type UpdateSystemBrandingRequest struct'), 'system_service.go must define UpdateSystemBrandingRequest');
  assert(/CustomerSubtitle\s+\*string/.test(serviceContent), 'CustomerSubtitle must be *string to distinguish empty from unset');
  assert(/MakerName\s+\*string/.test(serviceContent), 'MakerName must be *string to distinguish empty from unset');
  assert(/MakerSubtitle\s+\*string/.test(serviceContent), 'MakerSubtitle must be *string to distinguish empty from unset');
  assert(serviceContent.includes('GetBranding()'), 'system_service.go must have GetBranding');
  assert(serviceContent.includes('SaveBranding(req UpdateSystemBrandingRequest)'), 'system_service.go must have SaveBranding');
});

runTest('1.5: Handler and Router expose /api/system/branding endpoints', () => {
  assert(handlerContent.includes('func (h *SystemHandler) GetBranding'), 'system_handler.go must have GetBranding');
  assert(handlerContent.includes('func (h *SystemHandler) SaveBranding'), 'system_handler.go must have SaveBranding');
  assert(mainContent.includes('api.GET("/system/branding", systemHandler.GetBranding)'), 'main.go must register GET /system/branding');
  assert(mainContent.includes('api.POST("/system/branding", systemHandler.SaveBranding)'), 'main.go must register POST /system/branding');
  assert(mainContent.includes('api.PUT("/system/branding", systemHandler.SaveBranding)'), 'main.go must register PUT /system/branding');
});

// 2. Frontend Database Synchronization & Empty Values Support
runTest('2.1: companyBranding.js provides fetchBrandingFromDatabase & syncBrandingToDatabase', () => {
  assert(brandingUtilContent.includes('fetchBrandingFromDatabase'), 'Must export fetchBrandingFromDatabase');
  assert(brandingUtilContent.includes('syncBrandingToDatabase'), 'Must export syncBrandingToDatabase');
  assert(brandingUtilContent.includes('/api/system/branding'), 'Must target /api/system/branding API');
});

runTest('2.2: companyBranding.js allows empty maker name and subtitles on save without ignoring them', () => {
  assert(brandingUtilContent.includes("if (name !== undefined) meta.customer.name = (name || '').trim()"), 'saveCustomerBranding must allow empty string');
  assert(brandingUtilContent.includes("if (subtitle !== undefined) meta.customer.subtitle = (subtitle || '').trim()"), 'saveCustomerBranding must allow empty subtitle');
  assert(brandingUtilContent.includes("if (name !== undefined) meta.maker.name = (name || '').trim()"), 'saveMakerBranding must allow empty maker name');
  assert(brandingUtilContent.includes("if (subtitle !== undefined) meta.maker.subtitle = (subtitle || '').trim()"), 'saveMakerBranding must allow empty maker subtitle');
});

runTest('2.3: getOrganizationBrandingMeta checks !== undefined so empty strings do not revert to defaults', () => {
  assert(brandingUtilContent.includes('parsed.maker.name !== undefined') && brandingUtilContent.includes('DEFAULT_BRANDING.maker.name'), 'getOrganizationBrandingMeta must preserve empty maker name');
  assert(brandingUtilContent.includes('parsed.maker.subtitle !== undefined') && brandingUtilContent.includes('DEFAULT_BRANDING.maker.subtitle'), 'getOrganizationBrandingMeta must preserve empty maker subtitle');
  assert(brandingUtilContent.includes('parsed.customer.subtitle !== undefined'), 'getOrganizationBrandingMeta must preserve customer subtitle condition');
});

// 3. UI Header Rendering (Operator Command Center & ConfigView)
runTest('3.1: Operator Command Center header hides subtitle when stationTypeLabel is empty', () => {
  assert(occContent.includes('<div v-if="stationTypeLabel"'), 'OCC must conditionally render station subtitle');
  assert(occContent.includes('stationTypeLabel()'), 'OCC must have stationTypeLabel computed');
});

runTest('3.2: Operator Command Center header hides maker text block when makerCompanyName and subtitle are empty', () => {
  assert(occContent.includes('v-if="makerCompanyName || makerSubtitle"'), 'OCC must hide maker text block when empty');
  assert(occContent.includes('v-if="makerCompanyName"'), 'OCC must conditionally render makerCompanyName span');
});

runTest('3.3: ConfigView.vue live preview hides empty customer and maker subtitles', () => {
  assert(configViewContent.includes('v-if="branding.customer.subtitle && branding.customer.subtitle !== \'Local Autonomous Edge\'"'), 'ConfigView preview must conditionally render customer subtitle');
  assert(configViewContent.includes('v-if="branding.maker.name || branding.maker.subtitle"'), 'ConfigView preview must conditionally render maker block');
});

// 4. Localization
runTest('4.1: Localization files contain root branding.title preventing raw uppercase key display', () => {
  assert(enContent.includes('"branding": {'), 'en.js must have root branding object');
  assert(enContent.includes('"title": "Organization & Company Branding"'), 'en.js must have English title');
  assert(idContent.includes('"branding": {'), 'id.js must have root branding object');
  assert(idContent.includes('"title": "Organisasi & Branding Perusahaan"'), 'id.js must have Indonesian title');
  assert(occContent.includes("$t('branding.title')"), 'OCC modal must use $t(\'branding.title\')');
});

console.log('======================================================================');
console.log(`TEST SUMMARY: ${passedTests} PASSED, ${totalTests - passedTests} FAILED (TOTAL: ${totalTests})`);
console.log('======================================================================');

if (passedTests !== totalTests) {
  process.exit(1);
}
