$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginRes.data.token
$headers = @{ Authorization = "Bearer $token" }

Write-Host "=== 1. SYSTEM HEALTH ==="
$health = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/system/health" -Method Get
$health.data | ConvertTo-Json

Write-Host "=== 2. SYSTEM STATUS ==="
$status = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/system/status" -Method Get -Headers $headers
$status.data | ConvertTo-Json -Depth 4

Write-Host "=== 3. DATABASE STATUS ==="
$db = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/system/database/status" -Method Get -Headers $headers
$db.data | ConvertTo-Json -Depth 4

Write-Host "=== 4. DEVICES ==="
$devices = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/devices" -Method Get -Headers $headers
$devices.data | ForEach-Object {
    [PSCustomObject]@{
        id = $_.id
        name = $_.name
        protocol = $_.protocol
        is_enabled = $_.is_enabled
        status = $_.status
        device_type = $_.device_type
        port_name = $_.port_name
    }
} | Format-Table -AutoSize

Write-Host "=== 5. TELEMETRY LATEST & COUNTS ==="
$raw = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/telemetry/raw?page=1&page_size=5" -Method Get -Headers $headers
Write-Host ("Total Raw Records: " + $raw.data.total)
$raw.data.records | Select-Object id, device_id, parameter_id, value, quality, recorded_at | Format-Table -AutoSize

$hist = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/telemetry/history?page=1&page_size=5" -Method Get -Headers $headers
Write-Host ("Total Historical Records: " + $hist.data.total)

$agg = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/aggregations/results?page=1&limit=5" -Method Get -Headers $headers
Write-Host ("Total Aggregation Results: " + $agg.data.total)

Write-Host "=== 6. LOGS (RECENT) ==="
$logs = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/logs?page=1&limit=10" -Method Get -Headers $headers
$logs.data.logs | Select-Object id, level, message, created_at | Format-Table -AutoSize

Write-Host "=== 7. PHASE 10 TASKS ==="
$tasks = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/tasks?phase_id=10" -Method Get -Headers $headers
$tasks.data | Select-Object id, task_number, name, status, progress, assigned_to | Sort-Object id | Format-Table -AutoSize

Write-Host "=== 8. PHASE PROGRESS ==="
$phases = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/phases" -Method Get -Headers $headers
$phases.data | Select-Object id, phase_number, name, status, progress | Sort-Object id | Format-Table -AutoSize
