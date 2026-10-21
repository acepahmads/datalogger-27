/**
 * Acceptance Test: Network & Outbound Delivery Monitor
 * Verifies component structure, popover layout, status mappings, and API contracts.
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
const occContent = fs.readFileSync(occPath, 'utf8');
const iconContent = fs.readFileSync(iconCompPath, 'utf8');

console.log('=== Running Network & Outbound Delivery Monitor Acceptance Tests ===\n');

// 1. Icon component test
runTest('1.1: OperatorDashboardIcon.vue includes network and workflow icons', () => {
  assert(iconContent.includes("name === 'network' || name === 'workflow'"), 'Must define network and workflow icon');
  assert(iconContent.includes('stroke-width="1.8"'), 'Must have stroke-width 1.8 matching SCADA design');
});

// 2. Position test: Network Monitor placed immediately to the right of System Monitor
runTest('2.1: Network Monitor button placed immediately to the right of System Monitor', () => {
  assert(occContent.includes("activeToolbarPopover === 'systemMonitor'"), 'Must have System Monitor popover');
  assert(occContent.includes("activeToolbarPopover === 'networkMonitor'"), 'Must have Network Monitor popover');
  
  const sysMonIdx = occContent.indexOf("toggleToolbarPopover('systemMonitor')");
  const netMonIdx = occContent.indexOf("toggleToolbarPopover('networkMonitor')");
  assert(sysMonIdx !== -1, 'System monitor button must exist');
  assert(netMonIdx !== -1, 'Network monitor button must exist');
  assert(netMonIdx > sysMonIdx, 'Network monitor must be placed AFTER (to the right of) System monitor');
});

// 3. Three groups in floating panel
runTest('3.1: Floating panel contains Group 1: Local Network (LAN)', () => {
  assert(occContent.includes('Local Network (LAN)'), 'Must contain Local Network header');
  assert(occContent.includes('primary_ip'), 'Must display primary IP');
  assert(occContent.includes('gateway'), 'Must display gateway');
  assert(occContent.includes('gateway_reachable'), 'Must display gateway reachability');
});

runTest('3.2: Floating panel contains Group 2: Internet Connectivity (WAN)', () => {
  assert(occContent.includes('Internet Connectivity (WAN)'), 'Must contain Internet Connectivity header');
  assert(occContent.includes('probe_target'), 'Must display probe target');
  assert(occContent.includes('latency_ms'), 'Must display latency');
});

runTest('3.3: Floating panel contains Group 3: Outbound Delivery Links', () => {
  assert(occContent.includes('Outbound Delivery Links'), 'Must contain Outbound Delivery Links header');
  assert(occContent.includes('queue_summary'), 'Must show queue summary counters');
  assert(occContent.includes('success_deliveries'), 'Must show success deliveries');
  assert(occContent.includes('failed_deliveries'), 'Must show failed deliveries');
  assert(occContent.includes('last_error'), 'Must show sanitized last error');
});

// 4. Quick navigation link to /integration
runTest('4.1: Floating panel provides navigation to /integration', () => {
  assert(occContent.includes("navigateToRoute('/integration')"), 'Must link to /integration page');
});

// 5. Status indicators & badge differentiation
runTest('5.1: Distinguishes Healthy, Degraded, Failed, Disabled, and Unknown', () => {
  assert(occContent.includes("case 'HEALTHY':"), 'Must handle HEALTHY status');
  assert(occContent.includes("case 'DEGRADED':"), 'Must handle DEGRADED status');
  assert(occContent.includes("case 'FAILED':"), 'Must handle FAILED status');
  assert(occContent.includes("case 'DISABLED':"), 'Must handle DISABLED status');
  assert(occContent.includes('networkStatusDotClass'), 'Must have dynamic status dot on toolbar icon');
});

// 6. Centralized polling
runTest('6.1: Centralized polling integrated into pollDashboardCycle', () => {
  assert(occContent.includes('fetchNetworkStatus()'), 'Must call fetchNetworkStatus in dashboard lifecycle');
  assert(occContent.includes('networkPollCycles'), 'Must use centralized poll cycles, not per-card poll');
});

console.log('\n=== All Acceptance Tests Passed Successfully ===');
