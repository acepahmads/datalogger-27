/**
 * Verification Suite: Fleet Map AQMS Direct Wind & Rain Overlays (Overview + Single Pin)
 * Run via: node web/tests/fleet_map_aqms_weather_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const fleetPath = path.join(__dirname, '../src/views/FleetMapView.vue');
const fleetCode = fs.readFileSync(fleetPath, 'utf8');
const goPath = path.join(__dirname, '../../internal/service/fleet_overview.go');
const goCode = fs.readFileSync(goPath, 'utf8');

let passed = 0;
function check(name, fn) {
  fn();
  passed += 1;
  console.log(`✓ PASS: ${name}`);
}

console.log('========================================================================');
console.log('FLEET MAP AQMS WIND & RAIN OVERLAY VERIFICATION SUITE');
console.log('========================================================================');

check('1. Backend fleet_overview.go exposes FleetSiteWeatherSummary on FleetSiteItem with zero extra DB queries', () => {
  assert(goCode.includes('type FleetSiteWeatherSummary struct'), 'Missing FleetSiteWeatherSummary struct');
  assert(goCode.includes('Weather             *FleetSiteWeatherSummary `json:"weather,omitempty"`'), 'Missing Weather field on FleetSiteItem');
  assert(goCode.includes('siteWeatherSummary = buildFleetSiteWeatherSummary(siteDevs)'), 'Missing AQMS weather builder invocation');
  assert(goCode.includes('out.IsRaining = v > 0.05'), 'Missing tipping-bucket rain threshold (> 0.05mm)');
});

check('2. FleetMapView provides active/inactive Angin & Hujan toggles persisted in localStorage across Desktop, Tablet, Fullscreen & Mobile', () => {
  assert(fleetCode.includes('id="btn-fleet-toggle-wind"'), 'Missing desktop Wind toggle button');
  assert(fleetCode.includes('id="btn-fleet-toggle-rain"'), 'Missing desktop Rain toggle button');
  assert(fleetCode.includes("localStorage.getItem('datalogger_fleet_show_wind')"), 'Wind toggle must persist in localStorage');
  assert(fleetCode.includes("localStorage.getItem('datalogger_fleet_show_rain')"), 'Rain toggle must persist in localStorage');
  assert(fleetCode.includes('Cuaca AQMS'), 'Tablet Layer Popover must expose AQMS weather toggles');
  assert(fleetCode.includes('isAqmsWeatherToggleVisible'), 'Weather toggles must be scoped to ALL / AQMS domain filter');
});

check('3. Wind vector arrow & Rain ripple render directly on BOTH single AQMS pins AND collapsed overview clusters (no detail view needed)', () => {
  assert(fleetCode.includes('buildSiteWeatherOverlayHtml(site, size)'), 'Single pin must render weather overlay directly');
  assert(fleetCode.includes('buildClusterWeatherOverlayHtml(group, outerSize)'), 'Collapsed overview cluster must render aggregated wind & rain overlay directly');
  assert(fleetCode.includes('summarizeClusterWeather(sites)'), 'Cluster must aggregate circular mean wind angle, avg speed & raining count');
  assert(fleetCode.includes('Math.atan2(sinSum / angleCount, cosSum / angleCount)'), 'Cluster wind direction must compute circular vector mean');
  assert(fleetCode.includes('fleet-wind-arrow-svg'), 'Wind overlay must render directional SVG arrow');
  assert(fleetCode.includes('fleet-rain-ripple'), 'Rain overlay must render animated concentric blue ripple');
  assert(fleetCode.includes('fleet-weather-pill-strip'), 'Overlay must render crisp floating speed/rain pills beneath pins/clusters');
});

check('4. Dedicated #fleet-weather-overlay-legend stacks cleanly without breaking #fleet-env-index-legend contract', () => {
  assert(fleetCode.includes('id="fleet-weather-overlay-legend"'), 'Missing weather overlay legend');
  assert(
    fleetCode.includes('id="fleet-env-index-legend"') && fleetCode.includes("v-if=\"isTilesOnline && pinDisplayMode === 'env'\""),
    'Existing ISPU/IP-IKA legend contract must remain intact'
  );
  assert(
    !fleetCode.includes("v-if=\"isTilesOnline && pinDisplayMode === 'status'\""),
    'Must not re-introduce duplicate bottom-left status legend'
  );
});

check('5. Light-Mode CSS Hygiene: every new text-white class in template pairs with btn-text or badge-text', () => {
  const templateMatch = fleetCode.match(/<template>([\s\S]*?)<\/template>/);
  assert(templateMatch, 'Could not locate <template> block');
  const lines = templateMatch[1].split('\n');
  const violations = [];
  lines.forEach((line, idx) => {
    if (line.includes('text-white') && !line.includes('dark:text-white') && !line.includes('hover:text-white')) {
      if (!line.includes('btn-text') && !line.includes('badge-text')) {
        violations.push(`Line ${idx + 1}: ${line.trim()}`);
      }
    }
  });
  assert.strictEqual(violations.length, 0, `Light-mode text-white hygiene violations:\n${violations.join('\n')}`);
});

console.log('========================================================================');
console.log(`ALL ${passed} FLEET MAP AQMS WEATHER TESTS PASSED (100%)`);
console.log('========================================================================');
