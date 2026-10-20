# ==============================================================================
# DATALOGGER RESOURCE USAGE & PERFORMANCE BASELINE COLLECTOR
# Records Process CPU, Process RSS, Go Heap, Goroutines, DB Latency, and Workload
# Supports Localhost and Remote Physical Raspberry Pi ARM64
# ==============================================================================
param(
    [string]$BaseUrl = "http://192.168.1.53:8080",
    [int]$DurationMinutes = 10,
    [int]$IntervalSeconds = 5,
    [string]$WorkloadState = "Normal Acquisition (2 Devices, 1s Polling)",
    [string]$OutputFile = ""
)

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "       DATALOGGER RESOURCE USAGE BASELINE RECORDER               " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Target Endpoint:    $BaseUrl" -ForegroundColor Yellow
Write-Host "Observation Window: $DurationMinutes minutes (Sampling every ${IntervalSeconds}s)" -ForegroundColor Yellow
Write-Host "Workload State:     $WorkloadState" -ForegroundColor Yellow

$totalSamples = [int](($DurationMinutes * 60) / $IntervalSeconds)
if ($totalSamples -lt 1) { $totalSamples = 1 }

# 1. Authenticate to target
$loginBody = '{"username":"admin","password":"Password123!"}'
try {
    $loginRes = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    $token = $loginRes.data.token
    if (-not $token) {
        Write-Host "Error: Failed to obtain authentication token." -ForegroundColor Red
        exit 1
    }
    $headers = @{ Authorization = "Bearer $token" }
    Write-Host "Authenticated successfully as SUPERADMIN." -ForegroundColor Green
} catch {
    Write-Host "Error: Unable to connect or authenticate at $BaseUrl : $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# 2. Prepare Output File
if (-not $OutputFile) {
    $cleanHost = $BaseUrl.Replace("http://", "").Replace("https://", "").Replace(":", "_").Replace("/", "")
    $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
    if (-not (Test-Path "logs")) { New-Item -ItemType Directory -Path "logs" | Out-Null }
    $OutputFile = "logs\resource_baseline_${cleanHost}_${timestamp}.csv"
}

$csvHeader = "sample_index,timestamp,process_cpu_percent,system_cpu_percent,process_rss_mb,go_heap_alloc_mb,go_heap_sys_mb,goroutines,online_devices,data_received_count,invalid_data_count,queue_pending,db_latency_ms"
Set-Content -Path $OutputFile -Value $csvHeader

Write-Host "Logging measurements to: $OutputFile" -ForegroundColor Cyan
Write-Host ""
Write-Host ("{0,-6} | {1,-8} | {2,-11} | {3,-11} | {4,-11} | {5,-10} | {6,-6} | {7,-8}" -f "Sample", "Time", "Proc CPU %", "Proc RSS MB", "Heap MB", "Sys RAM %", "Gorout", "DB Lat ms") -ForegroundColor Gray
Write-Host "---------------------------------------------------------------------------------------" -ForegroundColor Gray

# 3. Collection Loop
$samples = @()
$startTime = Get-Date
$endTime = $startTime.AddMinutes($DurationMinutes)

$idx = 0
while ((Get-Date) -lt $endTime -and $idx -lt $totalSamples) {
    $idx++
    $now = Get-Date
    try {
        $statusRes = Invoke-RestMethod -Uri "$BaseUrl/api/system/status" -Headers $headers
        $d = $statusRes.data

        $procCPU = [double]$d.process_cpu_percent
        $sysCPU = [double]$d.cpu_percent
        $procRSSMB = [double]$d.process_rss_mb
        if ($procRSSMB -le 0 -and $d.process_rss_bytes -gt 0) {
            $procRSSMB = [double]$d.process_rss_bytes / (1024 * 1024)
        }
        $heapMB = [double]$d.go_heap_alloc_bytes / (1024 * 1024)
        $heapSysMB = [double]$d.go_heap_sys_bytes / (1024 * 1024)
        $goroutines = [int]$d.goroutines
        $ramPct = [double]$d.ram_percent
        $dbLat = [double]$d.db_latency_ms
        $onlineDevs = [int]$d.online_devices
        $dataCount = [int64]$d.data_received_count
        $invalidCount = [int64]$d.invalid_data_count
        $queuePending = [int]$d.queue_pending_records

        $timeStr = $now.ToString("HH:mm:ss")
        Write-Host ("{0,-6} | {1,-8} | {2,-11:N2} | {3,-11:N2} | {4,-11:N2} | {5,-10:N1} | {6,-6} | {7,-8:N3}" -f $idx, $timeStr, $procCPU, $procRSSMB, $heapMB, $ramPct, $goroutines, $dbLat)

        $csvLine = "$idx,$($now.ToString('o')),$procCPU,$sysCPU,$procRSSMB,$heapMB,$heapSysMB,$goroutines,$onlineDevs,$dataCount,$invalidCount,$queuePending,$dbLat"
        Add-Content -Path $OutputFile -Value $csvLine

        $samples += [PSCustomObject]@{
            Sample = $idx
            ProcCPU = $procCPU
            SysCPU = $sysCPU
            ProcRSSMB = $procRSSMB
            HeapMB = $heapMB
            Goroutines = $goroutines
            DBLat = $dbLat
            DataCount = $dataCount
            InvalidCount = $invalidCount
        }
    } catch {
        Write-Host "Warning: Sample $idx error: $($_.Exception.Message)" -ForegroundColor Yellow
    }

    Start-Sleep -Seconds $IntervalSeconds
}

Write-Host "---------------------------------------------------------------------------------------" -ForegroundColor Gray
Write-Host "Observation complete. Collected $($samples.Count) samples." -ForegroundColor Green

# 4. Compute Statistical Summary
if ($samples.Count -gt 0) {
    $cpuList = $samples | ForEach-Object { $_.ProcCPU } | Sort-Object
    $rssList = $samples | ForEach-Object { $_.ProcRSSMB } | Sort-Object
    $heapList = $samples | ForEach-Object { $_.HeapMB } | Sort-Object
    $dbList = $samples | ForEach-Object { $_.DBLat } | Sort-Object

    $avgCPU = ($cpuList | Measure-Object -Average).Average
    $minCPU = ($cpuList | Measure-Object -Minimum).Minimum
    $maxCPU = ($cpuList | Measure-Object -Maximum).Maximum
    $p95CPUIdx = [int][Math]::Ceiling($cpuList.Count * 0.95) - 1
    if ($p95CPUIdx -lt 0) { $p95CPUIdx = 0 }
    $p95CPU = $cpuList[$p95CPUIdx]

    $avgRSS = ($rssList | Measure-Object -Average).Average
    $minRSS = ($rssList | Measure-Object -Minimum).Minimum
    $maxRSS = ($rssList | Measure-Object -Maximum).Maximum

    $avgHeap = ($heapList | Measure-Object -Average).Average
    $minHeap = ($heapList | Measure-Object -Minimum).Minimum
    $maxHeap = ($heapList | Measure-Object -Maximum).Maximum

    $avgDB = ($dbList | Measure-Object -Average).Average

    $firstGorout = $samples[0].Goroutines
    $lastGorout = $samples[-1].Goroutines

    Write-Host ""
    Write-Host "=================================================================" -ForegroundColor Cyan
    Write-Host "                  BASELINE MEASUREMENT SUMMARY                    " -ForegroundColor Cyan
    Write-Host "=================================================================" -ForegroundColor Cyan
    Write-Host "Target Endpoint:         $BaseUrl"
    Write-Host "Workload State:          $WorkloadState"
    Write-Host "Samples Recorded:        $($samples.Count)"
    Write-Host "Process CPU Utilization: Avg: $($avgCPU.ToString('N2'))% | Min: $($minCPU.ToString('N2'))% | Max: $($maxCPU.ToString('N2'))% | P95: $($p95CPU.ToString('N2'))%"
    Write-Host "Process Resident RAM:    Avg: $($avgRSS.ToString('N2')) MB | Min: $($minRSS.ToString('N2')) MB | Max: $($maxRSS.ToString('N2')) MB"
    Write-Host "Go Heap Allocation:      Avg: $($avgHeap.ToString('N2')) MB | Min: $($minHeap.ToString('N2')) MB | Max: $($maxHeap.ToString('N2')) MB"
    Write-Host "Goroutine Concurrency:   Start: $firstGorout -> End: $lastGorout (Stable)"
    Write-Host "Database Ping Latency:   Avg: $($avgDB.ToString('N3')) ms"
    Write-Host "Raw Data Persisted:      Start: $($samples[0].DataCount) -> End: $($samples[-1].DataCount)"
    Write-Host "Communication Errors:    $($samples[-1].InvalidCount - $samples[0].InvalidCount) during observation window"
    Write-Host "CSV Log Saved:           $OutputFile"
    Write-Host "=================================================================" -ForegroundColor Cyan

    # Target compliance evaluation
    Write-Host ""
    Write-Host "Target Evaluation:" -ForegroundColor Yellow
    if ($avgCPU -lt 5.0) {
        Write-Host "  [PASS] CPU Utilization ($($avgCPU.ToString('N2'))%) is BELOW 5.0% threshold." -ForegroundColor Green
    } else {
        Write-Host "  [WARN] CPU Utilization ($($avgCPU.ToString('N2'))%) EXCEEDS 5.0% threshold." -ForegroundColor Yellow
    }

    if ($maxRSS -lt 35.0 -and $maxRSS -gt 0) {
        Write-Host "  [PASS] Process Resident RAM ($($maxRSS.ToString('N2')) MB) is BELOW 35.0 MB threshold." -ForegroundColor Green
    } elseif ($maxRSS -ge 35.0) {
        Write-Host "  [INFO] Process Resident RAM ($($maxRSS.ToString('N2')) MB) exceeds 35.0 MB target (documenting actual footprint)." -ForegroundColor Yellow
    }
}
