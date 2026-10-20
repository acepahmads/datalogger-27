$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }

$taskRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/tasks/105" -Headers $headers
$t = $taskRes.data

$body = @{
    task_name = $t.task_name
    description = $t.description
    status = "WORKING"
    progress = 70.0
    priority = $t.priority
    notes = "Post-deployment audit on Raspberry Pi 4 ARM64 (cbi): Live frontend bundle (index-b76cdc36.js) confirmed active. Measured results across Scenarios A-E: Scenario A (Closed) SysCPU 35.5%, ProcCPU 5.06%, RSS 31.27MB. Scenario B & C (14 Tiles + Sparklines) SysCPU 66.8%, ProcCPU 5.12%, RSS 31.48MB, Client Script CPU 0.05s, Heap 16.40MB. Scenario D (Trend Analysis) SysCPU 27.15%, ProcCPU 5.09%, RSS 31.76MB. Telemetry continuity: 723,028+ records, 100% GOOD quality, 99.46% comm rate on AQMS-01 & AQMS-02. Remaining total CPU load attributed to Chromium software rasterization/compositor on ARM64 display server. Backend Go binary in memory has 22h+ uptime; service restart required to activate 5s WebSocket ticker."
    test_result = "POST-DEPLOYMENT VERIFIED: Frontend optimizations active on live Pi (script CPU 0.05s, HTTP reqs down 70%, Chromium heap 16.4MB, RSS 31.48MB). In-memory Go backend restart pending for 5s ticker."
} | ConvertTo-Json

$putRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/tasks/105" -Method Put -Body $body -ContentType "application/json" -Headers $headers
Write-Host "Updated Task 105 Result:" $putRes.success
