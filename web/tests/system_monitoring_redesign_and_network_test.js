/**
 * Acceptance Test: System Monitoring Compact 4-Tier Bento Redesign, Network (Jaringan) Diagnostics,
 * and Cloud-Mode Operator Access Restriction.
 */
const fs = require('fs');
const path = require('path');
const assert = require('assert');

const rootDir = path.resolve(__dirname, '..', '..');
const srcDir = path.join(rootDir, 'web', 'src');

const viewCode = fs.readFileSync(path.join(srcDir, 'views', 'SystemMonitoringView.vue'), 'utf8');
const sidebarCode = fs.readFileSync(path.join(srcDir, 'components', 'Sidebar.vue'), 'utf8');
const routerCode = fs.readFileSync(path.join(srcDir, 'router', 'index.js'), 'utf8');
const storeCode = fs.readFileSync(path.join(srcDir, 'store', 'index.js'), 'utf8');
const enCode = fs.readFileSync(path.join(srcDir, 'i18n', 'locales', 'en.js'), 'utf8');
const idCode = fs.readFileSync(path.join(srcDir, 'i18n', 'locales', 'id.js'), 'utf8');
const sysinfoGo = fs.readFileSync(path.join(rootDir, 'pkg', 'sysinfo', 'sysinfo.go'), 'utf8');
const modelGo = fs.readFileSync(path.join(rootDir, 'internal', 'model', 'system.go'), 'utf8');
const serviceGo = fs.readFileSync(path.join(rootDir, 'internal', 'service', 'system_service.go'), 'utf8');

console.log('=== Running System Monitoring Redesign & Network (Jaringan) Tests ===');

// 1. Cloud Mode OPERATOR Restriction
assert(sidebarCode.includes('v-if="!isOperator || !isCloudMode" to="/monitoring/system"'), 'Sidebar must hide /monitoring/system for OPERATOR in Cloud mode');
assert(routerCode.includes("to.path === '/monitoring/system'") && routerCode.includes('isAuth && isOperator && isCloudMode'), 'Router must block Cloud OPERATOR from /monitoring/system');
assert(viewCode.includes('enforceCloudOperatorGuard()') && viewCode.includes('if (this.isOperator && this.isCloudMode)'), 'SystemMonitoringView must enforce Cloud OPERATOR guard');

// 2. 4-Tier Compact Bento Layout & Vue 2 Scoped CSS Safety
assert(!viewCode.includes(':global('), 'SystemMonitoringView must not use Vue 3 :global() inside Vue 2 <style scoped>');
assert(viewCode.includes('.sysmon-dark .sysmon-pill-emerald'), 'SystemMonitoringView must support .sysmon-dark scoped dark mode pills');
assert(viewCode.includes('xl:grid-cols-4'), 'SystemMonitoringView must align Hardware+Network and Subsystem cards in 4-column Bento grid');
for (const key of [
  'hostServiceTitle',
  'cpuCardTitle',
  'memoryCardTitle',
  'storageCardTitle',
  'networkCardTitle',
  'engineServicesTitle',
  'stationFleetTitle',
  'telemetryIngestionTitle',
  'communicationApiTitle',
  'timelineTitle',
]) {
  assert(viewCode.includes(`systemMonitoring.${key}`), `SystemMonitoringView must render ${key}`);
  assert(enCode.includes(`"${key}":`), `en.js must define systemMonitoring.${key}`);
  assert(idCode.includes(`"${key}":`), `id.js must define systemMonitoring.${key}`);
}

// 3. Network (Jaringan) Live Diagnostics & RX/TX Bandwidth Integration
assert(viewCode.includes("axios.get('/api/system/network')"), 'SystemMonitoringView must fetch live LAN/WAN/Outbound diagnostics from /api/system/network');
assert(viewCode.includes('net_rx_bytes_per_sec') && viewCode.includes('net_tx_bytes_per_sec'), 'SystemMonitoringView must display live RX & TX network rates');
assert(viewCode.includes('displayPrimaryIp') && viewCode.includes('displayGateway') && viewCode.includes('wanSummaryText'), 'SystemMonitoringView must display LAN IP, Gateway, and WAN Internet status');
assert(viewCode.includes('showNetworkModal') && viewCode.includes('sanitizedInterfaces') && viewCode.includes('outboundDestinations'), 'SystemMonitoringView must provide detailed NIC & Outbound modal');
assert(viewCode.includes('canViewHostPrivateNetwork'), 'SystemMonitoringView must protect host private LAN IP/MAC for non-Superadmin in Cloud mode');
assert(viewCode.includes("seriesNetwork") && viewCode.includes("netKbps"), '5-minute live timeline must include Network KB/s series');

// 4. Backend & Store Network Counter Wiring
assert(storeCode.includes('net_rx_bytes_per_sec') && storeCode.includes('net_tx_bytes_per_sec'), 'Vuex store must initialize network rate fields');
assert(sysinfoGo.includes('NetRxBytesPerSec') && sysinfoGo.includes('sampleNetworkCounters'), 'pkg/sysinfo/sysinfo.go must sample real-time NIC RX/TX rates');
assert(modelGo.includes('net_rx_bytes_per_sec') && modelGo.includes('net_tx_bytes_per_sec'), 'SystemStatusOverview model must expose net_rx_bytes_per_sec and net_tx_bytes_per_sec');
assert(serviceGo.includes('NetRxBytesPerSec:          info.NetRxBytesPerSec'), 'SystemService must map NetRxBytesPerSec into overview');

console.log('PASS: All System Monitoring Redesign, Network (Jaringan), and Cloud RBAC assertions passed!');

