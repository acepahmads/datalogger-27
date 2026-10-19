const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: STATION SETUP (AQMS/WQMS/SPARING) & CLEAN CARD REFINEMENT');
console.log('===============================================================');

const occViewPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

const configViewPath = path.join(__dirname, '..', 'src', 'views', 'ConfigView.vue');
const configContent = fs.readFileSync(configViewPath, 'utf8');

const brandingUtilPath = path.join(__dirname, '..', 'src', 'utils', 'companyBranding.js');
const brandingUtilContent = fs.readFileSync(brandingUtilPath, 'utf8');

// 1. Station Setup & Removal of Local Autonomous Edge
console.log('Checking Requirement 1: Station setup (AQMS, WQMS, Sparing) & Removal of Local Autonomous Edge...');

// Top panel header station line binds stationTypeLabel
assert(
  occContent.includes("{{ $t('operator.station') || 'Station' }}: <strong class=\"text-cyan-600 dark:text-cyan-400 font-bold tracking-wide\">{{ stationTypeLabel }}</strong>"),
  'OCC header must render stationTypeLabel without hardcoded cbi'
);

// Header station line must NOT contain Local Autonomous Edge
const occHeaderSection = occContent.slice(0, 1000);
assert(
  !occHeaderSection.includes('Local Autonomous Edge'),
  'OCC top header must NOT contain Local Autonomous Edge'
);

// stationTypeLabel computed property exists and handles AQMS, WQMS, Sparing
assert(occContent.includes('stationTypeLabel()'), 'OCC must provide stationTypeLabel computed property');
assert(occContent.includes("if (dom.toUpperCase() === 'SPARING') return 'Sparing';"), 'stationTypeLabel must format SPARING as Sparing');

// Customer branding setup modal provides quick preset buttons for AQMS, WQMS, Sparing
assert(occContent.includes("['AQMS', 'WQMS', 'Sparing']"), 'Customer branding modal must provide AQMS, WQMS, Sparing preset buttons');
assert(configContent.includes("['AQMS', 'WQMS', 'Sparing']"), 'ConfigView customer branding must provide AQMS, WQMS, Sparing preset buttons');

// ConfigView live preview does NOT contain Local Autonomous Edge
assert(!configContent.includes('Station: cbi ·'), 'ConfigView live preview must not render Station: cbi');

console.log('✓ PASS: 1.1: Station label adapts to AQMS, WQMS, Sparing and Local Autonomous Edge is removed');

// 2. Parameter Card Device Code Removed, Icon & Name Shifted Left with Cyan Font
console.log('Checking Requirement 2: Parameter card layout & cyan typography...');

// Status dot followed immediately by ParameterIcon and parameter name
assert(
  occContent.includes('<div class="flex items-center gap-0.5 min-w-0 flex-1 truncate">\n                      <ParameterIcon'),
  'ParameterIcon and name must be shifted directly to the left beside status dot'
);

// Parameter name styled with cyan font matching former device name
assert(
  occContent.includes('text-[10px] sm:text-[10.5px] font-bold text-cyan-600 dark:text-cyan-400 truncate tracking-tight font-mono'),
  'Parameter name must use cyan typography matching former device code style'
);

// Visible text {{ param.device_code }} is NOT present in the card header
const cardHeaderChunkMatch = occContent.match(/<!-- Area 1: Header[\s\S]*?<!-- Area 2: Main Content/);
assert(cardHeaderChunkMatch, 'Must find Area 1 Header');
const cardHeaderChunk = cardHeaderChunkMatch[0];
assert(
  !cardHeaderChunk.includes('{{ param.device_code }}'),
  'Card header must not render visible {{ param.device_code }} text'
);

console.log('✓ PASS: 2.1: Device code removed, icon and parameter name shifted left, and styled with cyan font');

// 3. Removal of GOOD Quality Badge (Only show when alarm or abnormal quality)
console.log('Checking Requirement 3: Removal of GOOD badge...');

assert(
  cardHeaderChunk.includes("v-if=\"param.quality && param.quality !== 'GOOD'\""),
  'Quality badge must only render when quality is not GOOD'
);
assert(
  !cardHeaderChunk.includes("{{ param.quality || 'GOOD' }}"),
  'Quality badge must not display fallback GOOD text'
);

console.log('✓ PASS: 3.1: GOOD text is removed from normal card; only alarms and non-GOOD states show badges');

// 4. Removal of FX badge from value row
console.log('Checking Requirement 4: Removal of FX badge...');

const cardValueChunkMatch = occContent.match(/<!-- Area 2: Main Content[\s\S]*?<!-- Right: Mini Sparkline/);
assert(cardValueChunkMatch, 'Must find Area 2 Main Content');
const cardValueChunk = cardValueChunkMatch[0];
assert(
  !cardValueChunk.includes("'FX'"),
  'Card value row must not render FX badge'
);
assert(
  !cardValueChunk.includes('param.formula'),
  'Card value row must not render formula badge'
);

console.log('✓ PASS: 4.1: FX badge completely removed from value row');

console.log('===============================================================');
console.log('TEST SUMMARY: ALL 4 REQUIREMENTS PASSED (TOTAL: 4)');
console.log('===============================================================');
