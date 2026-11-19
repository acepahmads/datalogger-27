const fs = require('fs');
const path = require('path');
const assert = require('assert');

const integrationViewPath = path.join(__dirname, '..', 'src', 'views', 'IntegrationView.vue');
const code = fs.readFileSync(integrationViewPath, 'utf8');

assert(code.includes('<option value="CUSTOM_SCHEMA">'), 'IntegrationView must offer CUSTOM_SCHEMA structure option');
assert(code.includes("setCustomSchemaMode('VISUAL_FIELDS')"), 'IntegrationView must include VISUAL_FIELDS mode toggle');
assert(code.includes("setCustomSchemaMode('JSON_TEMPLATE')"), 'IntegrationView must include JSON_TEMPLATE mode toggle');
assert(code.includes('custom_fields_json'), 'IntegrationView must bind custom_fields_json');
assert(code.includes('custom_template_json'), 'IntegrationView must bind custom_template_json');
assert(code.includes('convertVisualFieldsToJsonTemplate'), 'IntegrationView must allow converting Visual Fields to JSON Template');
assert(code.includes('insertCustomTemplateToken'), 'IntegrationView must provide clickable variable token insertion');
assert(code.includes('ALL_PARAMETERS'), 'IntegrationView Visual Builder must support ALL_PARAMETERS source');
assert(code.includes('RECORDS_VALUES_ARRAY'), 'IntegrationView Visual Builder must support RECORDS_VALUES_ARRAY source');

// 2-Mode Live JSON Preview Switch (Raw JSON vs Human-Readable Datetime)
assert(code.includes('id="btn-preview-mode-raw"'), 'IntegrationView must expose Raw JSON preview switch button');
assert(code.includes('id="btn-preview-mode-human"'), 'IntegrationView must expose Human-Readable Datetime preview switch button');
assert(code.includes('displayedPreviewJSON'), 'IntegrationView must render displayedPreviewJSON in live preview');
assert(code.includes('formatHumanPreviewTimestamp'), 'IntegrationView must format UNIX epoch timestamps into human-readable date/time');

// Historical Specific-Hour Preview & Test Delivery Selector (SPARING 30x2m complete hour testing)
assert(code.includes('id="select-preview-window-hour"'), 'IntegrationView must expose specific hour selector for bundled preview/test');
assert(code.includes('id="btn-pick-latest-complete-hour"'), 'IntegrationView must expose 1-click latest complete hour shortcut');
assert(code.includes('/preview-hours'), 'IntegrationView must fetch available DB hours via /preview-hours');
assert(code.includes('window_start: winStart'), 'IntegrationView must forward window_start to both preview and test endpoints');

console.log('PASS: Outbound Custom Schema UI (Visual Field Builder + JSON Template + 2-Mode Human Datetime Preview + Specific-Hour Test) verified!');

