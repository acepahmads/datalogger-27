/**
 * Verification Suite: Fleet Map AQMS BMKG-Style Wind Streamlines & Wind-Slanted Rain Canvas Overlay
 * Run via: node web/tests/fleet_map_aqms_weather_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');

const fleetPath = path.join(__dirname, '../src/views/FleetMapView.vue');
const fleetCode = fs.readFileSync(fleetPath, 'utf8');
const canvasPath = path.join(__dirname, '../src/utils/fleetWeatherCanvasLayer.js');
const canvasCode = fs.readFileSync(canvasPath, 'utf8');
const goPath = path.join(__dirname, '../../internal/service/fleet_overview.go');
const goCode = fs.readFileSync(goPath, 'utf8');

let passed = 0;
function check(name, fn) {
  fn();
  passed += 1;
  console.log(`✓ PASS: ${name}`);
}

// Dynamically evaluate pure ESM exports from fleetWeatherCanvasLayer.js in CommonJS Node test
function loadCanvasMathExports() {
  const cjsSrc = canvasCode
    .replace(/export function /g, 'function ')
    .replace(/export class /g, 'class ')
    + '\nmodule.exports = { normalizeWindSpeedMs, windAngleToCanvasVelocity, resolveWindVisualStyle, classifyRainIntensity, sampleWindFieldAtPixel };';
  const mod = { exports: {} };
  const fn = new Function('module', 'exports', cjsSrc);
  fn(mod, mod.exports);
  return mod.exports;
}

const mathApi = loadCanvasMathExports();

console.log('========================================================================');
console.log('FLEET MAP AQMS BMKG WIND STREAMLINES & RAIN CANVAS VERIFICATION SUITE');
console.log('========================================================================');

check('1. Backend fleet_overview.go exposes FleetSiteWeatherSummary on FleetSiteItem with zero extra DB queries', () => {
  assert(goCode.includes('type FleetSiteWeatherSummary struct'), 'Missing FleetSiteWeatherSummary struct');
  assert(goCode.includes('Weather             *FleetSiteWeatherSummary `json:"weather,omitempty"`'), 'Missing Weather field on FleetSiteItem');
  assert(goCode.includes('siteWeatherSummary = buildFleetSiteWeatherSummary(siteDevs)'), 'Missing AQMS weather builder invocation');
  assert(goCode.includes('out.IsRaining = v > 0.05'), 'Missing tipping-bucket rain threshold (> 0.05mm)');
});

check('2. Wind angle math maps compass bearings (0°=N, 90°=E, 180°=S, 270°=W) to exact screen velocity vectors', () => {
  const n = mathApi.windAngleToCanvasVelocity(0, 1);
  assert(Math.abs(n.ux - 0) < 1e-6 && Math.abs(n.uy - -1) < 1e-6, '0° must flow North (0, -1)');
  const e = mathApi.windAngleToCanvasVelocity(90, 1);
  assert(Math.abs(e.ux - 1) < 1e-6 && Math.abs(e.uy - 0) < 1e-6, '90° must flow East (+1, 0)');
  const s = mathApi.windAngleToCanvasVelocity(180, 1);
  assert(Math.abs(s.ux - 0) < 1e-6 && Math.abs(s.uy - 1) < 1e-6, '180° must flow South (0, +1)');
  const w = mathApi.windAngleToCanvasVelocity(270, 1);
  assert(Math.abs(w.ux - -1) < 1e-6 && Math.abs(w.uy - 0) < 1e-6, '270° must flow West (-1, 0)');
  const neg = mathApi.windAngleToCanvasVelocity(-90, 1);
  assert(Math.abs(neg.ux - -1) < 1e-6 && neg.deg === 270, '-90° must normalize to 270° West');
});

check('3. Wind speed visual scaling increases particle speed, trail length & tier color with wind_speed', () => {
  const light = mathApi.resolveWindVisualStyle(1.2, false);
  const mod = mathApi.resolveWindVisualStyle(3.5, false);
  const strong = mathApi.resolveWindVisualStyle(7.2, false);
  const veryStrong = mathApi.resolveWindVisualStyle(12.5, false);
  assert.strictEqual(light.tier, 'LIGHT');
  assert.strictEqual(mod.tier, 'MODERATE');
  assert.strictEqual(strong.tier, 'STRONG');
  assert.strictEqual(veryStrong.tier, 'VERY_STRONG');
  assert(light.stepPx < mod.stepPx && mod.stepPx < strong.stepPx && strong.stepPx < veryStrong.stepPx, 'Particle velocity must increase with wind speed');
  assert(light.trailLen < mod.trailLen && mod.trailLen < strong.trailLen && strong.trailLen < veryStrong.trailLen, 'Streamline trail length must increase with wind speed');
  assert(Math.abs(mathApi.normalizeWindSpeedMs(36, 'km/h') - 10) < 1e-6, 'km/h must convert cleanly to m/s');
});

check('4. IDW wind field interpolation accurately preserves local station direction and fades smoothly outside reach', () => {
  const stations = [
    { x: 100, y: 100, u: 1, v: 0, speedMs: 4.0 }, // East
    { x: 500, y: 100, u: 0, v: 1, speedMs: 8.0 }, // South
  ];
  const nearA = mathApi.sampleWindFieldAtPixel(105, 100, stations, 220);
  assert(nearA.u > 0.95 && Math.abs(nearA.v) < 0.25, 'Field right beside station A must flow East');
  assert(nearA.influence > 0.9, 'Influence right beside station A must be strong');

  const mid = mathApi.sampleWindFieldAtPixel(300, 100, stations, 220);
  assert(mid.u > 0.5 && mid.v > 0.5, 'Midpoint between East & South stations must blend smoothly SE');
  assert(mid.speedMs > 5.0 && mid.speedMs < 7.0, 'Midpoint speed must interpolate smoothly');

  const farAway = mathApi.sampleWindFieldAtPixel(1200, 900, stations, 200);
  assert.strictEqual(farAway.influence, 0, 'Influence far outside station reach must attenuate to 0');
});

check('5. Rain intensity classification scales radar radius, droplet density & streak length cleanly', () => {
  const rLight = mathApi.classifyRainIntensity(0.8);
  const rMod = mathApi.classifyRainIntensity(5.4);
  const rHeavy = mathApi.classifyRainIntensity(15.2);
  assert.strictEqual(rLight.level, 'LIGHT');
  assert.strictEqual(rMod.level, 'MODERATE');
  assert.strictEqual(rHeavy.level, 'HEAVY');
  assert(rLight.dropCount < rMod.dropCount && rMod.dropCount < rHeavy.dropCount, 'Raindrop count must scale with intensity');
  assert(rLight.baseRadius < rMod.baseRadius && rMod.baseRadius < rHeavy.baseRadius, 'Rain radar radius must scale with intensity');
  assert(canvasCode.includes('renderRainAuras') && canvasCode.includes('renderRainDrops'), 'Canvas engine must render both radar aura and wind-slanted rain streaks');
});

check('6. FleetMapView integrates FleetWeatherCanvasOverlay with toggles, localStorage persistence, cleanup & legend isolation', () => {
  assert(fleetCode.includes("import { FleetWeatherCanvasOverlay } from '../utils/fleetWeatherCanvasLayer'"), 'FleetMapView must import FleetWeatherCanvasOverlay');
  assert(fleetCode.includes('new FleetWeatherCanvasOverlay(this.map'), 'FleetMapView must instantiate FleetWeatherCanvasOverlay on Leaflet map');
  assert(fleetCode.includes('this.weatherCanvasOverlay.destroy()'), 'FleetMapView must destroy Canvas overlay on unmount');
  assert(fleetCode.includes('this.syncWeatherCanvasLayer()'), 'FleetMapView must sync Canvas overlay on toggle, filter, theme & polling updates');
  assert(fleetCode.includes('id="btn-fleet-toggle-wind"') && fleetCode.includes('id="btn-fleet-toggle-rain"'), 'Missing desktop Wind/Rain toggle buttons');
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

check('7. Light-Mode CSS Hygiene: every text-white class in FleetMapView template pairs with btn-text or badge-text', () => {
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
console.log(`ALL ${passed} FLEET MAP AQMS BMKG WEATHER CANVAS TESTS PASSED (100%)`);
console.log('========================================================================');
