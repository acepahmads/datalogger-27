const { spawn } = require('child_process');
const fs = require('fs');

async function sleep(ms) {
  return new Promise(r => setTimeout(r, ms));
}

async function getPiHealth(host) {
  const res = await fetch(`${host}/api/system/health`);
  return (await res.json()).data;
}

async function getPiStatus(host, token) {
  const res = await fetch(`${host}/api/system/status`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  return (await res.json()).data;
}

async function main() {
  const targetHost = process.argv[2] || 'http://192.168.1.53:8080';
  console.log('=================================================================');
  console.log(` PHASE 10: RASPBERRY PI MULTI-SCENARIO VERIFICATION BENCHMARK`);
  console.log(` Target Endpoint: ${targetHost}`);
  console.log('=================================================================');

  // Authenticate
  const loginRes = await fetch(`${targetHost}/api/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'admin', password: 'Password123!' })
  });
  const loginData = await loginRes.json();
  const token = loginData?.data?.token;
  if (!token) {
    console.error('Authentication failed');
    process.exit(1);
  }
  console.log(`Authenticated as ${loginData.data.user.role}.`);

  // SCENARIO A: Closed Dashboard (Backend + MariaDB only)
  console.log('\n>>> SCENARIO A: Dashboard Closed (Backend + MariaDB baseline, 30s) <<<');
  const aSamples = [];
  for (let i = 0; i < 6; i++) {
    await sleep(5000);
    const s = await getPiStatus(targetHost, token);
    aSamples.push(s);
    process.stdout.write(`  [${(i+1)*5}s] SysCPU: ${s.cpu_percent?.toFixed(1)}%, ProcCPU: ${s.process_cpu_percent?.toFixed(2)}%, RSS: ${s.process_rss_mb?.toFixed(2)}MB, DBLat: ${s.db_latency_ms}ms\n`);
  }
  const avgSysCpuA = aSamples.reduce((a, b) => a + (b.cpu_percent || 0), 0) / aSamples.length;
  const avgProcCpuA = aSamples.reduce((a, b) => a + (b.process_cpu_percent || 0), 0) / aSamples.length;
  const avgRssA = aSamples.reduce((a, b) => a + (b.process_rss_mb || 0), 0) / aSamples.length;

  // Launch Chromium via CDP for Scenarios B, C, D
  const profileDir = 'd:/cbi-project-src/datalogger-27/scratch/chrome-profile-scenarios';
  if (!fs.existsSync(profileDir)) fs.mkdirSync(profileDir, { recursive: true });

  const chrome = spawn('C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe', [
    '--headless=new',
    '--remote-debugging-port=9223',
    '--disable-gpu',
    `--user-data-dir=${profileDir}`,
    'about:blank'
  ]);

  let versionData = null;
  for (let i = 0; i < 40; i++) {
    try {
      const res = await fetch('http://127.0.0.1:9223/json/version');
      versionData = await res.json();
      if (versionData && versionData.webSocketDebuggerUrl) break;
    } catch (e) {
      await sleep(250);
    }
  }

  if (!versionData) {
    console.error('Could not connect to Chrome on port 9223');
    chrome.kill();
    process.exit(1);
  }

  const targetUrl = `${targetHost}/#/login`;
  const newTabRes = await fetch(`http://127.0.0.1:9223/json/new?${encodeURIComponent(targetUrl)}`, { method: 'PUT' });
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

  let networkRequests = [];
  let wsFrames = [];
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
      networkRequests.push({ url: data.params.request.url, time: Date.now() });
    } else if (data.method === 'Network.webSocketFrameReceived') {
      wsFrames.push({ time: Date.now(), length: data.params.response.payloadData.length });
    }
  };

  await new Promise(r => ws.onopen = r);
  await send('Page.enable');
  await send('Network.enable');
  await send('Performance.enable');
  await send('Runtime.enable');

  console.log(`Navigating to ${targetHost}/#/login...`);
  await send('Page.navigate', { url: `${targetHost}/#/login` });
  await sleep(3000);

  // Authenticate in Vue app
  await send('Runtime.evaluate', {
    expression: `
      (async () => {
        let app = document.querySelector('#app')?.__vue__;
        if (!app) {
          for (const el of document.querySelectorAll('*')) {
            if (el.__vue__) { app = el.__vue__; break; }
          }
        }
        if (app && app.$store) {
          await app.$store.dispatch('login', { username: 'admin', password: 'Password123!' });
          app.$router.push('/monitoring/operator');
          return 'OK';
        }
        return 'NO_APP';
      })()
    `,
    awaitPromise: true
  });

  await sleep(4000);

  // SCENARIO B & C: Operator Dashboard with parameter cards & sparklines
  console.log('\n>>> SCENARIO B & C: Operator Dashboard Open with Active Cards & Sparklines (30s) <<<');
  const tileCount = await send('Runtime.evaluate', {
    expression: `document.querySelectorAll('.scada-param-tile').length`
  });
  console.log(`Active SCADA Parameter Tiles Rendered: ${tileCount.result?.value}`);

  networkRequests = [];
  wsFrames = [];
  const startPerfBC = await send('Performance.getMetrics');
  const t0BC = Date.now();

  const bcSamples = [];
  for (let i = 0; i < 6; i++) {
    await sleep(5000);
    const s = await getPiStatus(targetHost, token);
    bcSamples.push(s);
    process.stdout.write(`  [${(i+1)*5}s] SysCPU: ${s.cpu_percent?.toFixed(1)}%, ProcCPU: ${s.process_cpu_percent?.toFixed(2)}%, RSS: ${s.process_rss_mb?.toFixed(2)}MB, DBLat: ${s.db_latency_ms}ms, Goroutines: ${s.goroutines}\n`);
  }
  const endPerfBC = await send('Performance.getMetrics');
  const t1BC = Date.now();
  const durBC = (t1BC - t0BC) / 1000;

  const mStartMapBC = {};
  for (const item of startPerfBC.metrics) mStartMapBC[item.name] = item.value;
  const mEndMapBC = {};
  for (const item of endPerfBC.metrics) mEndMapBC[item.name] = item.value;

  const scriptDurBC = mEndMapBC.ScriptDuration - mStartMapBC.ScriptDuration;
  const taskDurBC = mEndMapBC.TaskDuration - mStartMapBC.TaskDuration;
  const recalcsBC = mEndMapBC.RecalcStyleCount - mStartMapBC.RecalcStyleCount;
  const layoutsBC = mEndMapBC.LayoutCount - mStartMapBC.LayoutCount;

  // SCENARIO D: Trend Analysis panel enabled and scrubbed
  console.log('\n>>> SCENARIO D: Trend Analysis Enabled (Multi-Series Trend Rendering, 30s) <<<');
  // Click first 3 parameter tiles to select trend series
  await send('Runtime.evaluate', {
    expression: `
      (() => {
        const tiles = document.querySelectorAll('.scada-param-tile');
        for (let i = 0; i < Math.min(3, tiles.length); i++) {
          tiles[i].click();
        }
      })()
    `
  });
  await sleep(1500);

  networkRequests = [];
  wsFrames = [];
  const startPerfD = await send('Performance.getMetrics');
  const t0D = Date.now();

  const dSamples = [];
  for (let i = 0; i < 6; i++) {
    await sleep(5000);
    const s = await getPiStatus(targetHost, token);
    dSamples.push(s);
    process.stdout.write(`  [${(i+1)*5}s] SysCPU: ${s.cpu_percent?.toFixed(1)}%, ProcCPU: ${s.process_cpu_percent?.toFixed(2)}%, RSS: ${s.process_rss_mb?.toFixed(2)}MB, DBLat: ${s.db_latency_ms}ms\n`);
  }
  const endPerfD = await send('Performance.getMetrics');
  const t1D = Date.now();
  const durD = (t1D - t0D) / 1000;

  const mStartMapD = {};
  for (const item of startPerfD.metrics) mStartMapD[item.name] = item.value;
  const mEndMapD = {};
  for (const item of endPerfD.metrics) mEndMapD[item.name] = item.value;

  const scriptDurD = mEndMapD.ScriptDuration - mStartMapD.ScriptDuration;
  const taskDurD = mEndMapD.TaskDuration - mStartMapD.TaskDuration;

  // SCENARIO E: Telemetry continuity check across both Modbus devices
  console.log('\n>>> SCENARIO E: Modbus Telemetry Continuity & Raw Record Audit <<<');
  const rawRes = await fetch(`${targetHost}/api/telemetry/raw?page=1&page_size=10`, {
    headers: { Authorization: `Bearer ${token}` }
  });
  const rawData = await rawRes.json();
  const totalRawRecords = rawData.data?.total || 0;

  const latestStatus = await getPiStatus(targetHost, token);

  console.log('\n=================================================================');
  console.log('             FINAL COMPREHENSIVE BENCHMARK RESULTS               ');
  console.log('=================================================================');
  console.log(`Scenario A (Closed Dashboard):`);
  console.log(`  System CPU:             ${avgSysCpuA.toFixed(2)}%`);
  console.log(`  Process CPU:            ${avgProcCpuA.toFixed(2)}%`);
  console.log(`  Process RSS:            ${avgRssA.toFixed(2)} MB`);
  console.log(`\nScenario B & C (Operator Cards & Sparklines Active):`);
  console.log(`  Tiles Rendered:         ${tileCount.result?.value}`);
  console.log(`  System CPU:             ${(bcSamples.reduce((a,b)=>a+b.cpu_percent,0)/bcSamples.length).toFixed(2)}%`);
  console.log(`  Process CPU:            ${(bcSamples.reduce((a,b)=>a+b.process_cpu_percent,0)/bcSamples.length).toFixed(2)}%`);
  console.log(`  Process RSS:            ${(bcSamples.reduce((a,b)=>a+b.process_rss_mb,0)/bcSamples.length).toFixed(2)} MB`);
  console.log(`  Client JS Script Time:  ${scriptDurBC.toFixed(2)}s (${(scriptDurBC/durBC*100).toFixed(1)}% single-core)`);
  console.log(`  Client Total Task Time: ${taskDurBC.toFixed(2)}s (${(taskDurBC/durBC*100).toFixed(1)}% single-core)`);
  console.log(`  Style Recalculations:   ${recalcsBC} (${(recalcsBC/durBC).toFixed(1)} /s)`);
  console.log(`  Layout Passes:          ${layoutsBC} (${(layoutsBC/durBC).toFixed(1)} /s)`);
  console.log(`  JS Heap Used:           ${(mEndMapBC.JSHeapUsedSize/1048576).toFixed(2)} MB`);
  console.log(`  HTTP Request Rate:      ${(networkRequests.length/durBC*60).toFixed(1)} req/min`);
  console.log(`  WS Frame Rate:          ${(wsFrames.length/durBC).toFixed(1)} frames/sec`);
  console.log(`\nScenario D (Trend Analysis Active):`);
  console.log(`  System CPU:             ${(dSamples.reduce((a,b)=>a+b.cpu_percent,0)/dSamples.length).toFixed(2)}%`);
  console.log(`  Process CPU:            ${(dSamples.reduce((a,b)=>a+b.process_cpu_percent,0)/dSamples.length).toFixed(2)}%`);
  console.log(`  Process RSS:            ${(dSamples.reduce((a,b)=>a+b.process_rss_mb,0)/dSamples.length).toFixed(2)} MB`);
  console.log(`  Client JS Script Time:  ${scriptDurD.toFixed(2)}s`);
  console.log(`  Client Total Task Time: ${taskDurD.toFixed(2)}s`);
  console.log(`\nScenario E (Telemetry Continuity & Integrity):`);
  console.log(`  Total Devices:          ${latestStatus.total_devices}`);
  console.log(`  Total Raw Records:      ${totalRawRecords}`);
  console.log(`  Data Quality:           ${latestStatus.data_quality}`);
  console.log(`  Comm Success Rate:      ${latestStatus.success_rate?.toFixed(2)}%`);
  console.log(`  Last Comm Time:         ${latestStatus.last_device_comm}`);
  console.log(`  Last Data Received:     ${latestStatus.last_data_received}`);
  console.log(`  DB Latency:             ${latestStatus.db_latency_ms} ms`);
  console.log('=================================================================\n');

  ws.close();
  chrome.kill();
}

main().catch(err => {
  console.error(err);
  process.exit(1);
});
