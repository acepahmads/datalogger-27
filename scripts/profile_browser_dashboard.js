const { spawn } = require('child_process');
const fs = require('fs');

async function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}

async function main() {
  const targetHost = process.argv[2] || 'http://192.168.1.53:8080';
  console.log(`Starting Chrome CDP profile against: ${targetHost}`);

  const profileDir = 'd:/cbi-project-src/datalogger-27/scratch/chrome-profile-bench';
  if (!fs.existsSync(profileDir)) {
    fs.mkdirSync(profileDir, { recursive: true });
  }

  const chrome = spawn('C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe', [
    '--headless=new',
    '--remote-debugging-port=9222',
    '--disable-gpu',
    `--user-data-dir=${profileDir}`,
    'about:blank'
  ]);

  let versionData = null;
  for (let i = 0; i < 40; i++) {
    try {
      const res = await fetch('http://127.0.0.1:9222/json/version');
      versionData = await res.json();
      if (versionData && versionData.webSocketDebuggerUrl) break;
    } catch (e) {
      await sleep(250);
    }
  }

  if (!versionData) {
    console.error('Could not connect to Chrome debugger');
    chrome.kill();
    process.exit(1);
  }

  // 1. Authenticate to get JWT token
  console.log('Authenticating...');
  const loginRes = await fetch(`${targetHost}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'Password123!' })
  });
  const loginData = await loginRes.json();
  const token = loginData?.data?.token;
  const user = loginData?.data?.user;
  console.log(`Login response: token present = ${!!token}, role = ${user?.role}`);

  // Create page
  const targetUrl = `${targetHost}/#/monitoring/operator`;
  const newTabRes = await fetch(`http://127.0.0.1:9222/json/new?${encodeURIComponent(targetUrl)}`, { method: 'PUT' });
  const tabData = await newTabRes.json();
  const ws = new WebSocket(tabData.webSocketDebuggerUrl);

  let msgId = 1;
  const pending = new Map();
  function send(method, params = {}) {
    return new Promise((resolve, reject) => {
      const id = msgId++;
      pending.set(id, { resolve, reject });
      ws.send(JSON.stringify({ id, method, params }));
    });
  }

  const networkRequests = [];
  const wsFrames = [];

  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.id && pending.has(data.id)) {
      const { resolve, reject } = pending.get(data.id);
      pending.delete(data.id);
      if (data.error) reject(data.error);
      else resolve(data.result);
      return;
    }

    if (data.method === 'Network.requestWillBeSent') {
      networkRequests.push({
        url: data.params.request.url,
        method: data.params.request.method,
        time: Date.now()
      });
    } else if (data.method === 'Network.webSocketFrameReceived') {
      wsFrames.push({
        payloadLength: data.params.response.payloadData.length,
        time: Date.now()
      });
    }
  };

  await new Promise(r => ws.onopen = r);

  await send('Page.enable');
  await send('Network.enable');
  await send('Performance.enable');
  await send('Runtime.enable');

  console.log(`Navigating to ${targetHost}/#/login...`);
  await send('Page.navigate', { url: `${targetHost}/#/login` });
  await sleep(2500);

  const loginEval = await send('Runtime.evaluate', {
    expression: `
      (async () => {
        let app = document.querySelector('#app')?.__vue__;
        if (!app) {
          const els = document.querySelectorAll('*');
          for (const el of els) {
            if (el.__vue__) { app = el.__vue__; break; }
          }
        }
        if (app && app.$store) {
          await app.$store.dispatch('login', { username: 'admin', password: 'Password123!' });
          app.$router.push('/monitoring/operator');
          return 'LOGGED_IN_AND_PUSHED';
        }
        return 'NO_VUE_APP';
      })()
    `,
    awaitPromise: true
  });
  console.log('Login action result:', loginEval.result?.value);

  await sleep(5000);

  const curUrl = await send('Runtime.evaluate', { expression: 'window.location.href' });
  const docTitle = await send('Runtime.evaluate', { expression: 'document.title' });
  const paramCardCount = await send('Runtime.evaluate', { expression: 'document.querySelectorAll(".scada-param-card, [data-param-id], .monitoring-tile").length' });
  console.log(`Current page URL: ${curUrl.result?.value} | Title: ${docTitle.result?.value}`);
  console.log(`Active monitoring tiles found: ${paramCardCount.result?.value}`);

  console.log('\n--- Commencing 30-second profiling session ---');
  const metricsStart = await send('Performance.getMetrics');
  const startTime = Date.now();

  // Monitor over 30 seconds
  const samples = [];
  for (let s = 1; s <= 6; s++) {
    await sleep(5000);
    const m = await send('Performance.getMetrics');
    const metricMap = {};
    for (const item of m.metrics) {
      metricMap[item.name] = item.value;
    }
    samples.push({
      timeSec: s * 5,
      nodes: metricMap.Nodes,
      jsHeapUsedMB: (metricMap.JSHeapUsedSize / (1024 * 1024)).toFixed(2),
      jsHeapTotalMB: (metricMap.JSHeapTotalSize / (1024 * 1024)).toFixed(2),
      layoutCount: metricMap.LayoutCount,
      recalcStyleCount: metricMap.RecalcStyleCount,
      taskDurationSec: metricMap.TaskDuration?.toFixed(3),
      scriptDurationSec: metricMap.ScriptDuration?.toFixed(3),
      layoutDurationSec: metricMap.LayoutDuration?.toFixed(3),
      recalcStyleDurationSec: metricMap.RecalcStyleDuration?.toFixed(3),
    });
    console.log(`[T+${s*5}s] Nodes: ${metricMap.Nodes}, JSHeap: ${(metricMap.JSHeapUsedSize / 1048576).toFixed(1)}MB, ScriptDur: ${metricMap.ScriptDuration?.toFixed(2)}s, RecalcStyles: ${metricMap.RecalcStyleCount}`);
  }

  const endTime = Date.now();
  const totalDurationSec = (endTime - startTime) / 1000;

  const mStartMap = {};
  for (const item of metricsStart.metrics) {
    mStartMap[item.name] = item.value;
  }
  const mEndMap = {};
  const metricsEnd = await send('Performance.getMetrics');
  for (const item of metricsEnd.metrics) {
    mEndMap[item.name] = item.value;
  }

  const deltaScriptSec = mEndMap.ScriptDuration - mStartMap.ScriptDuration;
  const deltaTaskSec = mEndMap.TaskDuration - mStartMap.TaskDuration;
  const deltaLayoutCount = mEndMap.LayoutCount - mStartMap.LayoutCount;
  const deltaRecalcStyle = mEndMap.RecalcStyleCount - mStartMap.RecalcStyleCount;
  const avgScriptCPUPercent = (deltaScriptSec / totalDurationSec) * 100;
  const avgTaskCPUPercent = (deltaTaskSec / totalDurationSec) * 100;

  // Filter network requests made during the 30s window
  const benchRequests = networkRequests.filter(r => r.time >= startTime);
  const benchWsFrames = wsFrames.filter(f => f.time >= startTime);

  // Group requests by endpoint
  const endpointCounts = {};
  for (const r of benchRequests) {
    const urlObj = new URL(r.url);
    const path = urlObj.pathname;
    endpointCounts[path] = (endpointCounts[path] || 0) + 1;
  }

  console.log('\n=============================================================');
  console.log('       BROWSER OPERATOR DASHBOARD PROFILE RESULTS           ');
  console.log('=============================================================');
  console.log(`Duration:                 ${totalDurationSec.toFixed(1)}s`);
  console.log(`DOM Nodes:                ${mEndMap.Nodes}`);
  console.log(`JS Heap Used:             ${(mEndMap.JSHeapUsedSize / 1048576).toFixed(2)} MB (Max: ${(mEndMap.JSHeapTotalSize / 1048576).toFixed(2)} MB)`);
  console.log(`Style Recalculations:     ${deltaRecalcStyle} (${(deltaRecalcStyle / totalDurationSec).toFixed(1)} /sec)`);
  console.log(`Layout Passes:            ${deltaLayoutCount} (${(deltaLayoutCount / totalDurationSec).toFixed(1)} /sec)`);
  console.log(`Script Execution Time:    ${deltaScriptSec.toFixed(2)}s (Equivalent to ${avgScriptCPUPercent.toFixed(1)}% single-thread core time)`);
  console.log(`Total Task Execution Time:${deltaTaskSec.toFixed(2)}s (Equivalent to ${avgTaskCPUPercent.toFixed(1)}% single-thread core time)`);
  console.log(`WebSocket Frames Received:${benchWsFrames.length} (${(benchWsFrames.length / totalDurationSec).toFixed(1)} frames/sec)`);
  console.log(`Total HTTP Requests:      ${benchRequests.length} (${(benchRequests.length / totalDurationSec).toFixed(2)} req/sec)`);
  console.log('\nHTTP Endpoints Polled:');
  for (const [path, count] of Object.entries(endpointCounts)) {
    console.log(`  ${path.padEnd(45)} : ${count} requests (${(count / totalDurationSec * 60).toFixed(1)} /min)`);
  }
  console.log('=============================================================');

  // Query backend state
  const beRes = await fetch(`${targetHost}/api/system/status`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  const beStatus = (await beRes.json()).data;
  console.log('\nBackend State During Browser Session:');
  console.log(`  Process CPU:            ${beStatus.process_cpu_percent?.toFixed(2)}%`);
  console.log(`  System CPU:             ${beStatus.cpu_percent?.toFixed(2)}%`);
  console.log(`  Process RSS:            ${beStatus.process_rss_mb?.toFixed(2)} MB`);
  console.log(`  Goroutines:             ${beStatus.goroutines}`);
  console.log(`  DB Latency:             ${beStatus.db_latency_ms} ms`);
  console.log(`  Data Received:          ${beStatus.data_received_count}`);
  console.log('=============================================================\n');

  ws.close();
  chrome.kill();
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
