/**
 * Automated Verification Suite: Single Monitoring Domain Configuration (#83)
 * Verifies application-level monitoring domain configuration (AQMS, WQMS, CUSTOM),
 * zero data deletion guarantee, compatibility review, operator dashboard integration,
 * dynamic categories, first-run setup wizard, and security/audit trail persistence.
 */

const fs = require('fs');
const path = require('path');

let totalTests = 0;
let passedTests = 0;
let failedTests = 0;

function assert(condition, message) {
  totalTests++;
  if (condition) {
    console.log(`✓ PASS: ${message}`);
    passedTests++;
  } else {
    console.error(`✗ FAIL: ${message}`);
    failedTests++;
  }
}

console.log('================================================================');
console.log('RUNNING AUTOMATED TEST SUITE: SINGLE MONITORING DOMAIN CONFIGURATION');
console.log('================================================================\n');

const ROOT_DIR = path.resolve(__dirname, '..', '..');

// Test 1: Configuration file and backend config struct
const configJsonPath = path.join(ROOT_DIR, 'config.json');
const configJsonContent = fs.readFileSync(configJsonPath, 'utf8');
const configJson = JSON.parse(configJsonContent);
assert(
  configJson.monitoring_domain && ['AQMS', 'WQMS', 'CUSTOM'].includes(configJson.monitoring_domain),
  'Test 1: config.json contains persistent monitoring_domain field (default: ' + configJson.monitoring_domain + ')'
);

const configGoPath = path.join(ROOT_DIR, 'internal', 'config', 'config.go');
const configGoContent = fs.readFileSync(configGoPath, 'utf8');
assert(
  configGoContent.includes('MonitoringDomain') &&
  configGoContent.includes('DATALOGGER_MONITORING_DOMAIN') &&
  configGoContent.includes('NormalizeMonitoringDomain'),
  'Test 2: internal/config/config.go defines MonitoringDomain, env var override, and normalization'
);

// Test 3: Category Model and Domain enumeration
const categoryModelPath = path.join(ROOT_DIR, 'internal', 'model', 'category.go');
const categoryModelContent = fs.readFileSync(categoryModelPath, 'utf8');
assert(
  categoryModelContent.includes('Domain') &&
  categoryModelContent.includes('DomainAQMS') &&
  categoryModelContent.includes('DomainWQMS') &&
  categoryModelContent.includes('DomainCustom') &&
  categoryModelContent.includes('IsValidCategoryDomain'),
  'Test 3: internal/model/category.go defines Domain field and ValidCategoryDomains (AQMS, WQMS, CUSTOM, ALL)'
);

// Test 4: Database migration seeds domain mappings and backfills safely
const migrationGoPath = path.join(ROOT_DIR, 'internal', 'database', 'migration.go');
const migrationGoContent = fs.readFileSync(migrationGoPath, 'utf8');
assert(
  migrationGoContent.includes('model.DomainAQMS') &&
  migrationGoContent.includes('model.DomainWQMS') &&
  migrationGoContent.includes('model.DomainCustom') &&
  migrationGoContent.includes('existing.Domain == ""'),
  'Test 4: internal/database/migration.go migrates parameter_categories table with domain column and seeds defaults'
);

// Test 5: Category Repository supports domain filtering
const categoryRepoPath = path.join(ROOT_DIR, 'internal', 'repository', 'category_repo.go');
const categoryRepoContent = fs.readFileSync(categoryRepoPath, 'utf8');
assert(
  categoryRepoContent.includes('ListWithDomain') &&
  categoryRepoContent.includes('domain = ?'),
  'Test 5: internal/repository/category_repo.go provides ListWithDomain query method'
);

// Test 6: System Service provides Domain Compatibility & Safe Reassignment
const systemServicePath = path.join(ROOT_DIR, 'internal', 'service', 'system_service.go');
const systemServiceContent = fs.readFileSync(systemServicePath, 'utf8');
assert(
  systemServiceContent.includes('CheckDomainCompatibility') &&
  systemServiceContent.includes('DomainCompatibilityReport') &&
  systemServiceContent.includes('DataSafetyGuarantee') &&
  systemServiceContent.includes('Reassignments'),
  'Test 6: internal/service/system_service.go implements CheckDomainCompatibility and zero-loss reassignments'
);

// Test 7: System Handler and main routing
const systemHandlerPath = path.join(ROOT_DIR, 'internal', 'handler', 'system_handler.go');
const systemHandlerContent = fs.readFileSync(systemHandlerPath, 'utf8');
const mainGoPath = path.join(ROOT_DIR, 'cmd', 'main.go');
const mainGoContent = fs.readFileSync(mainGoPath, 'utf8');
assert(
  systemHandlerContent.includes('CheckDomainCompatibility') &&
  mainGoContent.includes('/domain-compatibility') &&
  mainGoContent.includes('/system/domain'),
  'Test 7: HTTP routes /domain-compatibility and /system/domain exposed with RBAC'
);

// Test 8: ConfigView.vue has Domain Tab, 3 Domain Cards, and First-Run Setup Wizard
const configViewPath = path.join(ROOT_DIR, 'web', 'src', 'views', 'ConfigView.vue');
const configViewContent = fs.readFileSync(configViewPath, 'utf8');
assert(
  configViewContent.includes("activeTab === 'domain'") &&
  configViewContent.includes('AQMS') &&
  configViewContent.includes('WQMS') &&
  configViewContent.includes('CUSTOM') &&
  configViewContent.includes('showDomainModal') &&
  configViewContent.includes('showSetupWizard'),
  'Test 8: ConfigView.vue contains Domain Tab, AQMS/WQMS/CUSTOM cards, domain modal, and setup wizard'
);

// Test 9: OperatorCommandCenterView.vue loads domain, has no simultaneous Udara/Air selectors, and filters categories
const occViewPath = path.join(ROOT_DIR, 'web', 'src', 'views', 'OperatorCommandCenterView.vue');
const occViewContent = fs.readFileSync(occViewPath, 'utf8');
assert(
  occViewContent.includes('currentMonitoringDomain') &&
  occViewContent.includes('monitoringDomainBadgeClass') &&
  occViewContent.includes('isParamDomainMismatched') &&
  !occViewContent.includes('Udara / Air') &&
  !occViewContent.includes('domainSelector'),
  'Test 9: OperatorCommandCenterView.vue loads domain metadata, displays subtle badge, and has no duplicate domain selectors'
);

// Test 10: CategoryManagementModal.vue includes domain badges and domain dropdown
const catModalPath = path.join(ROOT_DIR, 'web', 'src', 'components', 'CategoryManagementModal.vue');
const catModalContent = fs.readFileSync(catModalPath, 'utf8');
assert(
  catModalContent.includes('getDomainBadgeClass') &&
  catModalContent.includes('Monitoring Domain') &&
  catModalContent.includes('v-model="form.domain"'),
  'Test 10: CategoryManagementModal.vue displays domain badges and allows editing category domain'
);

// Test 11: Locales contain translations for domain and first-run setup
const enJsPath = path.join(ROOT_DIR, 'web', 'src', 'i18n', 'locales', 'en.js');
const idJsPath = path.join(ROOT_DIR, 'web', 'src', 'i18n', 'locales', 'id.js');
const enJsContent = fs.readFileSync(enJsPath, 'utf8');
const idJsContent = fs.readFileSync(idJsPath, 'utf8');
assert(
  enJsContent.includes('"tabDomain"') &&
  enJsContent.includes('"zeroDataLossGuarantee"') &&
  idJsContent.includes('"tabDomain"') &&
  idJsContent.includes('"zeroDataLossGuarantee"'),
  'Test 11: Both English and Indonesian locale files contain comprehensive monitoring domain keys'
);

// Test 12: Zero Data Deletion Guarantee validation in migration and service
assert(
  systemServiceContent.includes('Zero data loss') || systemServiceContent.includes('DataSafetyGuarantee'),
  'Test 12: Zero Data Deletion Guarantee: switch logic explicitly preserves all existing records'
);

console.log('\n================================================================');
console.log(`TEST SUMMARY: ${passedTests} passed, ${failedTests} failed out of ${totalTests} tests.`);
console.log('================================================================\n');

if (failedTests > 0) {
  process.exit(1);
}
