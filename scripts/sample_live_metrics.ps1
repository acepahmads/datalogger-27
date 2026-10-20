$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }

Write-Host "Collecting 10 consecutive live telemetry samples at 3s intervals from Raspberry Pi..."
Write-Host ("{0,-8} | {1,-10} | {2,-11} | {3,-10} | {4,-11} | {5,-10} | {6,-6} | {7,-8} | {8,-12}" -f "Time", "Sys CPU %", "Proc CPU %", "Sys RAM %", "Proc RSS MB", "Heap (MB)", "Gorout", "DB Lat", "Data Count")
Write-Host ("-" * 100)

for ($i = 1; $i -le 10; $i++) {
    $s = (Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/system/status" -Headers $headers).data
    $heapMB = [math]::Round($s.go_heap_alloc_bytes / 1MB, 2)
    $rssMB = [math]::Round($s.process_rss_mb, 2)
    $sysCpu = [math]::Round($s.cpu_percent, 2)
    $procCpu = [math]::Round($s.process_cpu_percent, 2)
    $sysRam = [math]::Round($s.ram_percent, 2)
    $timeStr = (Get-Date).ToString("HH:mm:ss")
    
    Write-Host ("{0,-8} | {1,-10} | {2,-11} | {3,-10} | {4,-11} | {5,-10} | {6,-6} | {7,-8} | {8,-12}" -f $timeStr, $sysCpu, $procCpu, $sysRam, $rssMB, $heapMB, $s.go_num_goroutine, $s.db_latency_ms, $s.data_received_count)
    if ($i -lt 10) { Start-Sleep -Seconds 3 }
}
