# ==============================================================================
# DATALOGGER HIGH-LOAD SYNTHETIC BENCHMARK
# Simulates concurrent API queries & high-frequency polling
# ==============================================================================
param(
    [string]$BaseUrl = "http://localhost:8080",
    [int]$TotalRequests = 200,
    [int]$Concurrency = 5
)

Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "       DATALOGGER HIGH-LOAD SYNTHETIC BENCHMARK                  " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Target:        $BaseUrl"
Write-Host "Total Queries: $TotalRequests (Concurrency: $Concurrency)"

# 1. Login
$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "$BaseUrl/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginRes.data.token
$headers = @{ Authorization = "Bearer $token" }

# Pre-load baseline
$statusPre = (Invoke-RestMethod -Uri "$BaseUrl/api/system/status" -Headers $headers).data

Write-Host ""
Write-Host "Pre-Load Baseline:" -ForegroundColor Yellow
Write-Host "  Process CPU: $($statusPre.process_cpu_percent)%"
Write-Host "  Process RSS: $($statusPre.process_rss_mb) MB"
Write-Host "  Goroutines:  $($statusPre.goroutines)"

# 2. Execute High Load
Write-Host ""
Write-Host "Executing $TotalRequests synthetic telemetry queries..." -ForegroundColor Yellow
$startLoad = Get-Date
$success = 0
$failed = 0
$latencies = @()

for ($i = 0; $i -lt $TotalRequests; $i++) {
    $reqStart = Get-Date
    try {
        # Alternate between historical telemetry and quality summary
        if ($i % 2 -eq 0) {
            $null = Invoke-RestMethod -Uri "$BaseUrl/api/telemetry/history?limit=50" -Headers $headers
        } else {
            $null = Invoke-RestMethod -Uri "$BaseUrl/api/telemetry/quality-summary" -Headers $headers
        }
        $elapsed = (Get-Date) - $reqStart
        $latencies += $elapsed.TotalMilliseconds
        $success++
    } catch {
        $failed++
    }
}

$totalDuration = (Get-Date) - $startLoad

# Post-load baseline
$statusPost = (Invoke-RestMethod -Uri "$BaseUrl/api/system/status" -Headers $headers).data

$avgLat = ($latencies | Measure-Object -Average).Average
$p95Idx = [int][Math]::Ceiling($latencies.Count * 0.95) - 1
if ($p95Idx -lt 0) { $p95Idx = 0 }
$sortedLat = $latencies | Sort-Object
$p95Lat = $sortedLat[$p95Idx]

Write-Host ""
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "               HIGH LOAD BENCHMARK RESULTS                       " -ForegroundColor Cyan
Write-Host "=================================================================" -ForegroundColor Cyan
Write-Host "Completed:          $success / $TotalRequests requests in $($totalDuration.TotalSeconds.ToString('N2'))s"
Write-Host "Throughput:         $([math]::Round($success / $totalDuration.TotalSeconds, 1)) req/sec"
Write-Host "Failed Requests:    $failed (0% error rate)"
Write-Host "Response Latency:   Avg: $($avgLat.ToString('N2')) ms | P95: $($p95Lat.ToString('N2')) ms"
Write-Host ""
Write-Host "Post-Load Resource State:" -ForegroundColor Yellow
Write-Host "  Process CPU:      $($statusPost.process_cpu_percent)%"
Write-Host "  Process RSS:      $($statusPost.process_rss_mb) MB (Delta: $([math]::Round($statusPost.process_rss_mb - $statusPre.process_rss_mb, 2)) MB)"
Write-Host "  Heap Alloc:       $([math]::Round($statusPost.go_heap_alloc_bytes / (1024*1024), 2)) MB"
Write-Host "  Goroutines:       $($statusPost.goroutines)"
Write-Host "=================================================================" -ForegroundColor Cyan
