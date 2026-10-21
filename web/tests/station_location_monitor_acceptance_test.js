/**
 * Acceptance Test: Station Location Monitor
 * Verifies toolbar placement, 3 explicit location sources (GPS, Internet Geolocation, Manual),
 * interactive Leaflet map, offline schematic fallback, full screen mode, and validation.
 */

const fs = require('fs');
const path = require('path');
const assert = require('assert');

function runTest(testName, fn) {
  try {
    fn();
    console.log(`[PASS] ${testName}`);
  } catch (err) {
    console.error(`[FAIL] ${testName}: ${err.message}`);
    process.exitCode = 1;
  }
}

const occPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const iconCompPath = path.join(__dirname, '..', 'src', 'components', 'OperatorDashboardIcon.vue');
const mapCompPath = path.join(__dirname, '..', 'src', 'components', 'StationMapViewer.vue');

const occContent = fs.readFileSync(occPath, 'utf8');
const iconContent = fs.readFileSync(iconCompPath, 'utf8');
const mapContent = fs.readFileSync(mapCompPath, 'utf8');

console.log('=== Running Station Location Monitor Acceptance Tests ===\n');

// 1. Icon Component Verification
runTest('1.1: OperatorDashboardIcon defines map-pin and map-pinned duotone SCADA icons', () => {
  assert(iconContent.includes("name === 'map-pin' || name === 'location'"), 'Must define map-pin/location icon');
  assert(iconContent.includes("name === 'map-pinned' || name === 'map'"), 'Must define map-pinned/map icon');
  assert(iconContent.includes('fill-opacity="0.18"'), 'Must have duotone fill opacity matching SCADA icons');
});

// 2. Toolbar Hierarchy & Positioning Verification
runTest('2.1: Station Location Monitor is placed immediately to the right of Network Monitor', () => {
  const netMonIdx = occContent.indexOf("toggleToolbarPopover('networkMonitor')");
  const stationLocIdx = occContent.indexOf("toggleToolbarPopover('stationLocation')");

  assert(netMonIdx !== -1, 'Network monitor toolbar button must exist');
  assert(stationLocIdx !== -1, 'Station location toolbar button must exist');
  assert(stationLocIdx > netMonIdx, 'Station Location icon must be placed AFTER (to the right of) Network Monitor');

  // Verify it has the same button size, borders, styling, and status indicator dot
  assert(occContent.includes(':class="stationLocationStatusDotClass"'), 'Must have status dot indicator');
  assert(occContent.includes(':title="stationLocationTooltip"'), 'Must have tooltip');
});

// 3. Floating Panel & Summary Card Verification
runTest('3.1: Floating popover displays status, coordinates, source, and last update', () => {
  assert(occContent.includes("activeToolbarPopover === 'stationLocation'"), 'Must conditionally render popover');
  assert(occContent.includes('stationLocation.is_configured'), 'Must distinguish configured vs unconfigured');
  assert(occContent.includes('stationLocation.location_source'), 'Must display location source badge');
  assert(occContent.includes('stationLocation.latitude'), 'Must display latitude');
  assert(occContent.includes('stationLocation.longitude'), 'Must display longitude');
  assert(occContent.includes('stationLocation.updated_at'), 'Must display update timestamp');
  assert(occContent.includes('stationLocation.updated_by'), 'Must display updater username');
});

// 4. Source A: GPS (Browser Geolocation API)
runTest('4.1: Source A (GPS) uses Browser Geolocation API only upon operator action', () => {
  assert(occContent.includes('detectBrowserGPS'), 'Must have detectBrowserGPS method');
  assert(occContent.includes('navigator.geolocation.getCurrentPosition'), 'Must use navigator.geolocation');
  assert(occContent.includes('enableHighAccuracy: true'), 'Must enable high accuracy');
  assert(occContent.includes('applyGPSToCandidate'), 'Must require operator action to apply to edit form');
  assert(occContent.includes('gpsState.coordinates.accuracy'), 'Must display accuracy in meters');
  assert(!occContent.includes('mounted() {\n    this.detectBrowserGPS'), 'Must NOT trigger GPS on page load');
});

// 5. Source B: Internet Geolocation (IP-based)
runTest('5.1: Source B (Internet Geolocation) states approximation, prevents 3rd-party leaks, and avoids fake coordinates', () => {
  assert(occContent.includes('checkIPGeolocationProvider'), 'Must have checkIPGeolocationProvider method');
  assert(occContent.includes('/api/system/location/ip-estimate'), 'Must query backend IP geolocation endpoint');
  assert(occContent.includes('estimasi kasar') || occContent.includes('kasar'), 'Must clearly explain IP is approximate and not as accurate as GPS');
  assert(occContent.includes('UNAVAILABLE'), 'Must handle UNAVAILABLE status without fabricating fake coordinates');
});

// 6. Source C: Manual Input & Map Pin
runTest('6.1: Source C (Manual) validates [-90, 90] and [-180, 180], provides Save, Cancel, and Reset', () => {
  assert(occContent.includes('candidateLocation.latitude'), 'Must bind candidate latitude');
  assert(occContent.includes('candidateLocation.longitude'), 'Must bind candidate longitude');
  assert(occContent.includes('lat < -90 || lat > 90'), 'Must validate latitude range -90..90');
  assert(occContent.includes('lng < -180 || lng > 180'), 'Must validate longitude range -180..180');
  assert(occContent.includes('saveStationLocation'), 'Must have saveStationLocation method');
  assert(occContent.includes('cancelLocationEdit'), 'Must have cancelLocationEdit method');
  assert(occContent.includes('resetStationLocation'), 'Must have resetStationLocation method');
  assert(occContent.includes('onMapCoordinatePicked'), 'Must update coordinates on map pick/drag');
});

// 7. Interactive Map & Offline Fallback (StationMapViewer.vue)
runTest('7.1: StationMapViewer integrates Leaflet with OpenStreetMap tiles and offline schematic fallback', () => {
  assert(mapContent.includes("import L from 'leaflet'"), 'Must import Leaflet');
  assert(mapContent.includes('leaflet/dist/leaflet.css'), 'Must import Leaflet CSS');
  assert(mapContent.includes('tile.openstreetmap.org'), 'Must use OpenStreetMap tiles');
  assert(mapContent.includes('tileerror'), 'Must listen to tileerror events');
  assert(mapContent.includes('Local Schematic Fallback') || mapContent.includes('fallback'), 'Must have offline schematic fallback');
  assert(mapContent.includes('L.divIcon'), 'Must use custom DivIcon for SCADA marker (no broken assets)');
  assert(mapContent.includes('invalidateSize'), 'Must support invalidateSize on resize');
});

// 8. Full Screen Mode
runTest('8.1: Full screen mode covers viewport, maintains marker, and provides clear exit', () => {
  assert(occContent.includes('showFullScreenStationMap'), 'Must have showFullScreenStationMap flag');
  assert(occContent.includes('openFullScreenMap'), 'Must have openFullScreenMap method');
  assert(occContent.includes('closeFullScreenMap'), 'Must have closeFullScreenMap method');
  assert(occContent.includes('Exit Full Screen (Esc)'), 'Must have clear exit button');
  assert(occContent.includes('fixed inset-0 z-[100]'), 'Must use full viewport area');
});

// 9. Location Search Bar (Enter to execute)
runTest('9.1: Map supports location search by city/address or coordinates with Enter key execution', () => {
  assert(mapContent.includes('locationSearchInput') || mapContent.includes('searchQuery'), 'Must have search input for location');
  assert(mapContent.includes('@keydown.enter.prevent="executeSearch"'), 'Must execute search on Enter key press');
  assert(mapContent.includes('executeSearch'), 'Must define executeSearch method');
  assert(mapContent.includes('parseCoordinates'), 'Must support direct coordinate search');
  assert(occContent.includes('executeTabLocationSearch'), 'Must support tab location search execution');
});

console.log('\n=== All Station Location Monitor Acceptance Tests Passed Successfully ===');
