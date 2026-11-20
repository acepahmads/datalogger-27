/**
 * Verification Suite: Dashboard Operator Weather Hero Card (Right Bento Column & Compact Strip)
 * Run via: node web/tests/dashboard_operator_weather_card_test.js
 */
const assert = require('assert');
const fs = require('fs');
const path = require('path');
const vm = require('vm');
const { execSync } = require('child_process');

const repoRoot = path.resolve(__dirname, '../..');
const weatherUtilRaw = fs.readFileSync(path.join(__dirname, '../src/utils/weatherParams.js'), 'utf8');
const weatherCompCode = fs.readFileSync(path.join(__dirname, '../src/components/BentoWeatherHeroCard.vue'), 'utf8');
const dashViewCode = fs.readFileSync(path.join(__dirname, '../src/views/DashboardOperatorView.vue'), 'utf8');

// Load ESM weatherParams.js in Node VM for direct runtime assertions
const cjsWeatherCode =
  weatherUtilRaw.replace(/export\s+function\s+/g, 'function ') +
  '\nmodule.exports = { classifyWeatherParameters, parseNumericValue, getCardinalDirection, inferWeatherCondition, getTempComfortMeta, getHumidityComfortMeta, getWindBeaufortLabel };';
const weatherMod = { exports: {} };
vm.runInNewContext(cjsWeatherCode, { module: weatherMod, exports: weatherMod.exports, Date, Math, Number, String, Boolean, Array, Set, Object, RegExp });
const {
  classifyWeatherParameters,
  getCardinalDirection,
  inferWeatherCondition,
  getTempComfortMeta,
  getHumidityComfortMeta,
  getWindBeaufortLabel,
} = weatherMod.exports;

let passed = 0;
function check(name, fn) {
  fn();
  passed += 1;
  console.log(`✓ PASS: ${name}`);
}

console.log('========================================================================');
console.log('DASHBOARD OPERATOR WEATHER HERO CARD VERIFICATION SUITE');
console.log('========================================================================');

check('1. Standing protected files remain 100% untouched in git', () => {
  const diff = execSync(
    'git diff -- web/src/views/OperatorCommandCenterView.vue web/src/components/ISPUIndicatorBar.vue web/src/components/OperatorDashboardIcon.vue',
    { cwd: repoRoot, encoding: 'utf8' }
  ).trim();
  assert.strictEqual(diff, '', 'Protected files must have zero git diff');
});

check('2. weatherParams classifies full AQMS weather suite & excludes pollutants/shelter/water sensors', () => {
  const categories = [
    { id: 1, name: 'GAS & POLLUTANTS' },
    { id: 2, name: 'PARTICULATE MATTER' },
    { id: 3, name: 'WEATHER / METEOROLOGY' },
  ];
  const params = [
    { param_id: 1, param_code: 'co', param_name: 'CO', unit: 'µg/m³', value: 114.77, quality: 'GOOD', category_id: 1 },
    { param_id: 2, param_code: 'no2', param_name: 'NO2', unit: 'µg/m³', value: 7.9, quality: 'GOOD', category_id: 1 },
    { param_id: 3, param_code: 'pm10', param_name: 'PM10', unit: 'µg/m³', value: 19.27, quality: 'GOOD', category_id: 2 },
    { param_id: 4, param_code: 'pm25', param_name: 'PM2.5', unit: 'µg/m³', value: 10.92, quality: 'GOOD', category_id: 2 },
    { param_id: 5, param_code: 'shelter_temp', param_name: 'Shelter Temp', unit: '°C', value: 24.5, quality: 'GOOD', category_id: 1 },
    { param_id: 6, param_code: 'pump_pressure', param_name: 'Pump Pressure', unit: 'kPa', value: 95.1, quality: 'GOOD', category_id: 1 },
    { param_id: 10, param_code: 'humidity', param_name: 'Humidity', unit: '%', value: 61.78, quality: 'GOOD', category_id: 3 },
    { param_id: 11, param_code: 'solar_radiation', param_name: 'Solar Radiation', unit: 'W/m²', value: 875.27, quality: 'GOOD', category_id: 3 },
    { param_id: 12, param_code: 'air_pressure', param_name: 'Air Pressure', unit: 'hPa', value: 1008.92, quality: 'GOOD', category_id: 3 },
    { param_id: 13, param_code: 'rain_fall', param_name: 'Rain Fall', unit: 'mm', value: 0.0, quality: 'GOOD', category_id: 3 },
    { param_id: 14, param_code: 'temperature', param_name: 'Temperature', unit: '°C', value: 32.8, quality: 'GOOD', category_id: 3 },
    { param_id: 15, param_code: 'wind_angle', param_name: 'Wind Angle', unit: '°', value: 274.26, quality: 'GOOD', category_id: 3 },
    { param_id: 16, param_code: 'wind_speed', param_name: 'Wind Speed', unit: 'm/s', value: 1.11, quality: 'GOOD', category_id: 3 },
  ];

  const res = classifyWeatherParameters(params, null, categories);
  assert.strictEqual(res.hasAnyWeather, true);
  assert.strictEqual(res.hasPrimaryClimate, true);
  assert.strictEqual(res.count, 7);
  assert.strictEqual(res.temperature.param_id, 14, 'Must pick ambient Temperature, never shelter_temp');
  assert.strictEqual(res.humidity.param_id, 10);
  assert.strictEqual(res.solarRadiation.param_id, 11);
  assert.strictEqual(res.pressure.param_id, 12, 'Must pick Air Pressure, never pump_pressure');
  assert.strictEqual(res.rainFall.param_id, 13);
  assert.strictEqual(res.windAngle.param_id, 15);
  assert.strictEqual(res.windSpeed.param_id, 16);
  assert.strictEqual(res.matchedIds.has(5), false, 'shelter_temp must NOT be matched as weather');
  assert.strictEqual(res.matchedIds.has(6), false, 'pump_pressure must NOT be matched as weather');
});

check('3. weatherParams auto-hides on pure SPARING / WQMS stations without outdoor weather sensors', () => {
  const sparingParams = [
    { param_id: 1, param_code: 'ph', param_name: 'pH', unit: '', value: 7.1 },
    { param_id: 2, param_code: 'cod', param_name: 'COD', unit: 'mg/L', value: 22.5 },
    { param_id: 3, param_code: 'tss', param_name: 'TSS', unit: 'mg/L', value: 11.4 },
    { param_id: 4, param_code: 'nh3n', param_name: 'NH3-N', unit: 'mg/L', value: 1.2 },
    { param_id: 5, param_code: 'debit', param_name: 'Debit', unit: 'm³/h', value: 14.0 },
    { param_id: 6, param_code: 'water_temp', param_name: 'Water Temp', unit: '°C', value: 28.4 },
  ];
  const res = classifyWeatherParameters(sparingParams);
  assert.strictEqual(res.hasAnyWeather, false, 'Pure SPARING station must return hasAnyWeather=false');
  assert.strictEqual(res.count, 0);
});

check('4. Cardinal direction, atmospheric condition inference, and comfort helpers work accurately', () => {
  const west = getCardinalDirection(274.26);
  assert.strictEqual(west.deg, 274);
  assert.strictEqual(west.short, 'W');
  assert.strictEqual(west.label, 'Barat');

  const north = getCardinalDirection(359);
  assert.strictEqual(north.short, 'N');
  assert.strictEqual(north.label, 'Utara');

  const sunnyDay = inferWeatherCondition({ temperature: 32.8, humidity: 61.78, solarRadiation: 875.27, rainFall: 0, nowDate: new Date(2026, 10, 20, 11, 0) });
  assert.strictEqual(sunnyDay.key, 'sunny');
  assert.strictEqual(sunnyDay.label, 'Cerah Terik');

  const rainyDay = inferWeatherCondition({ temperature: 25.0, humidity: 88, solarRadiation: 120, rainFall: 2.4, nowDate: new Date(2026, 10, 20, 14, 0) });
  assert.strictEqual(rainyDay.key, 'rain');

  const clearNight = inferWeatherCondition({ temperature: 24.5, humidity: 70, solarRadiation: 0, rainFall: 0, nowDate: new Date(2026, 10, 20, 21, 30) });
  assert.strictEqual(clearNight.key, 'night_clear');

  assert.strictEqual(getTempComfortMeta(32.8).label, 'Panas');
  assert.strictEqual(getHumidityComfortMeta(61.78).label, 'Ideal');
  assert.strictEqual(getWindBeaufortLabel(1.11), 'Sepoi Lembut');
});

check('5. BentoWeatherHeroCard supports Full Right Bento Hero + Compact Horizontal Strip + Light-Mode CSS hygiene', () => {
  assert(weatherCompCode.includes('id="bento-weather-hero-card"'), 'Must define #bento-weather-hero-card');
  assert(weatherCompCode.includes('is-compact-weather-bar') && weatherCompCode.includes('bento-weather-card'), 'Must support both compact and full hero modes');
  assert(weatherCompCode.includes('wh-sky-band') && weatherCompCode.includes('wh-hum-tile') && weatherCompCode.includes('wh-wind-band'), 'Must render Atmospheric Sky band, Humidity showcase tile, and Wind Compass banner');
  assert(weatherCompCode.includes('compassNeedleStyle') && weatherCompCode.includes('compactCompassNeedleStyle'), 'Must rotate wind compass needles cleanly in both full and compact views');
  assert(weatherCompCode.includes("this.$emit('select-param'") && weatherCompCode.includes("this.$emit('context-param'"), 'Must emit trend selection and context menu events');
  assert(weatherCompCode.includes("this.$emit('focus-weather-trend'"), 'Must emit 1-click weather trend focus event');

  // Light-mode CSS hygiene check
  const templateSection = weatherCompCode.slice(0, weatherCompCode.indexOf('</template>'));
  assert(!/(?<!dark:)bg-slate-900\b/.test(templateSection), 'Must not use bare bg-slate-900 without dark: prefix');
  assert(!/(?<!dark:)text-white\b(?![^"']*(btn-text|badge-text))/.test(templateSection), 'Must not use bare text-white without btn-text/badge-text');
});

check('6. DashboardOperatorView integrates symmetric 3-column Bento layout, compact ordering, toggles & localStorage persistence', () => {
  assert(dashViewCode.includes("import BentoWeatherHeroCard from '../components/BentoWeatherHeroCard.vue'"), 'Must import BentoWeatherHeroCard');
  assert(dashViewCode.includes('<BentoWeatherHeroCard'), 'Must render BentoWeatherHeroCard inside .bento-mid');
  assert(dashViewCode.includes("'bento-mid-both'") && dashViewCode.includes('.bento-mid.bento-mid-both'), 'Must style 3-column symmetric Bento grid (ISPU + Params + Weather)');
  assert(dashViewCode.includes("'bento-mid-weather-only'") && dashViewCode.includes('.bento-mid.bento-mid-weather-only'), 'Must style 2-column Weather-only grid when ISPU is hidden');
  assert(dashViewCode.includes("'bento-mid-compact-weather-right'") && dashViewCode.includes('.bento-mid.is-compact-mid.bento-mid-compact-weather-right'), 'Must dock full vertical Weather card on right of parameters at medium/tablet landscape resolutions (e.g. 1024x768)');
  assert(dashViewCode.includes('isWeatherCompactStrip()') && dashViewCode.includes(':compact="isWeatherCompactStrip"'), 'Must only collapse Weather card to horizontal strip below 900px');
  assert(dashViewCode.includes('.bento-mid.is-compact-mid > .bento-weather-root'), 'Must position compact Weather strip above parameters on narrow tablet portrait/mobile');
  assert(dashViewCode.includes('id="btn-bento-weather-toggle"'), 'Must render quick 1-click Weather toggle icon button');
  assert(dashViewCode.includes("localStorage.getItem('datalogger_bento_show_weather_card')") && dashViewCode.includes("localStorage.getItem('datalogger_bento_hide_dup_weather')"), 'Must persist weather visibility & deduplication preferences in localStorage');
  assert(dashViewCode.includes('shouldDeduplicateWeatherInGrid()') && dashViewCode.includes('nonWeatherCount > 0'), 'Must protect pure AWS stations from ending up with an empty center grid');
  assert(dashViewCode.includes('focusBentoWeatherTrend('), 'Must implement focusBentoWeatherTrend handler');
  for (const methodSig of ['getBentoCardRenderMemo(param) {', 'resolveBentoParamLimits(param) {', 'getBentoParamActiveAlarm(param) {', 'getBentoCardAlertClass(param) {', 'onBentoStatusBadgeClick(param) {']) {
    assert(dashViewCode.includes(methodSig), `DashboardOperatorView must preserve method definition: ${methodSig}`);
  }
});

console.log('========================================================================');
console.log(`ALL ${passed} WEATHER HERO CARD TESTS PASSED (100%)`);
console.log('========================================================================');
