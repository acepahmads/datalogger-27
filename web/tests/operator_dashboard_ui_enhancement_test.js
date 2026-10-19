const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: OPERATOR DASHBOARD UI ENHANCEMENT');
console.log('(MULTICOLOR SVG ICONS & SCADA THEME COLOR SPECIFICATION)');
console.log('===============================================================');

function runTest(name, fn) {
  try {
    fn();
    console.log(`✓ PASS: ${name}`);
  } catch (err) {
    console.error(`✗ FAIL: ${name}`);
    console.error(`  Error: ${err.message}`);
    process.exit(1);
  }
}

const occPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occPath, 'utf8');

const palettePath = path.join(__dirname, '..', 'src', 'utils', 'trendPalette.js');
const paletteContent = fs.readFileSync(palettePath, 'utf8');

const iconCompPath = path.join(__dirname, '..', 'src', 'components', 'OperatorDashboardIcon.vue');
const iconCompContent = fs.readFileSync(iconCompPath, 'utf8');

const appPath = path.join(__dirname, '..', 'src', 'App.vue');
const appContent = fs.readFileSync(appPath, 'utf8');

// -------------------------------------------------------------
// REQUIREMENT 1: MULTICOLOR / DUOTONE SVG ICONS
// -------------------------------------------------------------

runTest('1.1: OperatorDashboardIcon.vue exists with duotone/multicolor SVG definitions', () => {
  assert(fs.existsSync(iconCompPath), 'OperatorDashboardIcon.vue must exist');
  assert(iconCompContent.includes('name: \'OperatorDashboardIcon\''), 'Must have OperatorDashboardIcon component name');
  assert(iconCompContent.includes('fill-opacity='), 'Must have translucent fill layers for duotone accents');
  assert(iconCompContent.includes('name === \'station\''), 'Must have station icon');
  assert(iconCompContent.includes('name === \'chart\''), 'Must have chart icon');
  assert(iconCompContent.includes('name === \'export\''), 'Must have export icon');
  assert(iconCompContent.includes('name === \'report\''), 'Must have report icon');
  assert(iconCompContent.includes('name === \'history\''), 'Must have history icon');
  assert(iconCompContent.includes('name === \'sound-on\''), 'Must have sound-on icon');
  assert(iconCompContent.includes('name === \'sound-muted\''), 'Must have sound-muted icon');
});

runTest('1.2: OperatorCommandCenterView.vue registers and uses OperatorDashboardIcon', () => {
  assert(occContent.includes("import OperatorDashboardIcon from '../components/OperatorDashboardIcon.vue'"), 'Must import OperatorDashboardIcon');
  assert(occContent.includes('OperatorDashboardIcon,'), 'Must register OperatorDashboardIcon in components');
  assert(occContent.includes('<OperatorDashboardIcon'), 'Must render OperatorDashboardIcon in template');
});

// -------------------------------------------------------------
// REQUIREMENT 2: OPTIMIZED DARK AND LIGHT THEME TOKENS
// -------------------------------------------------------------

runTest('2.1: trendPalette.js exports DASHBOARD_THEME_TOKENS matching Industrial Navy & Soft Sky Blue', () => {
  assert(paletteContent.includes('DASHBOARD_THEME_TOKENS'), 'Must export DASHBOARD_THEME_TOKENS');
  // Dark mode
  assert(paletteContent.includes("bg: '#081421'"), 'Dark background must be #081421');
  assert(paletteContent.includes("panel: '#0D1B2A'"), 'Dark panel must be #0D1B2A');
  assert(paletteContent.includes("card: '#112235'"), 'Dark card must be #112235');
  assert(paletteContent.includes("border: '#294158'"), 'Dark border must be #294158');
  assert(paletteContent.includes("textPrimary: '#E2EBF4'"), 'Dark primary text must be #E2EBF4');
  assert(paletteContent.includes("textSecondary: '#91A6BA'"), 'Dark secondary text must be #91A6BA');
  assert(paletteContent.includes("accent: '#35C8DA'"), 'Dark accent must be #35C8DA');
  assert(paletteContent.includes("normal: '#36C99A'"), 'Dark normal/good must be #36C99A');
  assert(paletteContent.includes("warning: '#F1BD61'"), 'Dark warning must be #F1BD61');
  assert(paletteContent.includes("alarm: '#FF6577'"), 'Dark alarm must be #FF6577');
  
  // Light mode
  assert(paletteContent.includes("bg: '#EDF4F8'"), 'Light background must be #EDF4F8');
  assert(paletteContent.includes("panel: '#F5F9FC'"), 'Light panel must be #F5F9FC');
  assert(paletteContent.includes("card: '#FFFFFF'"), 'Light card must be #FFFFFF');
  assert(paletteContent.includes("border: '#C8DCE8'"), 'Light border must be #C8DCE8');
  assert(paletteContent.includes("textPrimary: '#20364B'"), 'Light primary text must be #20364B');
  assert(paletteContent.includes("textSecondary: '#647D91'"), 'Light secondary text must be #647D91');
  assert(paletteContent.includes("accent: '#009DB8'"), 'Light accent must be #009DB8');
  assert(paletteContent.includes("normal: '#138B68'"), 'Light normal/good must be #138B68');
  assert(paletteContent.includes("warning: '#A96C0A'"), 'Light warning must be #A96C0A');
  assert(paletteContent.includes("alarm: '#C93449'"), 'Light alarm must be #C93449');
});

runTest('2.2: App.vue scopes Operator route background to #EDF4F8 / #081421', () => {
  assert(appContent.includes("isOperatorRoute"), 'Must check isOperatorRoute');
  assert(appContent.includes("bg-[#EDF4F8] dark:bg-[#081421]"), 'App.vue must style operator route background');
});

runTest('2.3: OperatorCommandCenterView.vue defines theme CSS variables in scoped styles', () => {
  assert(occContent.includes('--occ-bg: #EDF4F8;'), 'Must have --occ-bg in light mode');
  assert(occContent.includes('--occ-bg: #081421;'), 'Must have --occ-bg in dark mode');
  assert(occContent.includes('--occ-panel: #0D1B2A;'), 'Must have --occ-panel in dark mode');
  assert(occContent.includes('--occ-card: #112235;'), 'Must have --occ-card in dark mode');
  assert(occContent.includes('--occ-border: #294158;'), 'Must have --occ-border in dark mode');
});

// -------------------------------------------------------------
// REQUIREMENT 3: EXACT PER-PARAMETER CANONICAL TREND COLOR PALETTE
// -------------------------------------------------------------

runTest('3.1: All 14 canonical parameter colors defined for both Dark and Light modes', () => {
  const expectedMappings = [
    { key: 'pm25', dark: '#27D9F2', light: '#0284C7' },
    { key: 'pm10', dark: '#5794FF', light: '#2563EB' },
    { key: 'o3', dark: '#B493FF', light: '#7C3AED' },
    { key: 'so2', dark: '#36D6CE', light: '#0D9488' },
    { key: 'no2', dark: '#FF83B5', light: '#E11D48' },
    { key: 'co', dark: '#43D9A3', light: '#059669' },
    { key: 'hc', dark: '#D4E66D', light: '#65A30D' },
    { key: 'temperature', dark: '#FFBB66', light: '#D97706' },
    { key: 'humidity', dark: '#5CA9FF', light: '#1D4ED8' },
    { key: 'pressure', dark: '#A8B8CC', light: '#475569' },
    { key: 'wind_speed', dark: '#50C7E8', light: '#0891B2' },
    { key: 'wind_angle', dark: '#C1A0FF', light: '#6D28D9' },
    { key: 'solar_radiation', dark: '#FFD166', light: '#B45309' },
    { key: 'rain_fall', dark: '#7DC8FF', light: '#4338CA' },
  ];

  expectedMappings.forEach(({ key, dark, light }) => {
    assert(paletteContent.includes(dark), `Canonical color ${dark} for ${key} must exist in trendPalette`);
    assert(paletteContent.includes(light), `Canonical color ${light} for ${key} must exist in trendPalette`);
  });
});

runTest('3.2: getParamSeriesColor helper handles theme toggling dynamically', () => {
  assert(paletteContent.includes('export function getParamSeriesColor(param, isDark = true)'), 'Must export getParamSeriesColor helper');
  assert(occContent.includes('getParamSeriesColor(param, this.isDark)'), 'OCC must invoke getParamSeriesColor with isDark state');
});

runTest('3.3: Series chart lines have 100% opacity when active without dimming', () => {
  assert(occContent.includes(':opacity="hoveredTrendSeriesId && hoveredTrendSeriesId !== s.id ? 0.35 : 1"'), 'Trend line stroke must have 100% opacity (1)');
});

// -------------------------------------------------------------
// REQUIREMENT 4: READABILITY & SPARKLINE COLOR COORDINATION
// -------------------------------------------------------------

runTest('4.1: Parameter telemetry values use high-contrast text tokens', () => {
  assert(occContent.includes("paramValueColorClass(param)"), 'Must have paramValueColorClass method');
  assert(occContent.includes("text-[#20364B] dark:text-[#E2EBF4]"), 'Normal values must use high-contrast primary text tokens');
  assert(occContent.includes("text-[#C93449] dark:text-[#FF6577]"), 'Critical values must use high-contrast alarm tokens');
  assert(occContent.includes("text-[#A96C0A] dark:text-[#F1BD61]"), 'Warning values must use high-contrast warning tokens');
});

runTest('4.2: Sparklines coordinate color with canonical parameter colors and alarm states', () => {
  assert(occContent.includes('getSparklineColor(param)'), 'Must invoke getSparklineColor');
  assert(occContent.includes(':stop-color="getSparklineColor(param)"'), 'Sparkline gradient must bind to getSparklineColor');
  assert(occContent.includes(':stroke="getSparklineColor(param)"'), 'Sparkline stroke must bind to getSparklineColor');
});

console.log('===============================================================');
console.log('TEST SUMMARY: ALL OPERATOR DASHBOARD UI ENHANCEMENT TESTS PASSED!');
console.log('===============================================================');
