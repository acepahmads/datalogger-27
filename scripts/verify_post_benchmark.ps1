$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }

$health = (Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/system/health").data
Write-Host "=== POST-BENCHMARK HEALTH ==="
Write-Host ("Process CPU:   " + [math]::Round($health.process_cpu_percent, 2) + "%")
Write-Host ("Process RSS:   " + [math]::Round($health.process_rss_mb, 2) + " MB (" + $health.process_rss_bytes + " bytes)")
Write-Host ("Go Heap Alloc: " + [math]::Round($health.go_heap_alloc_bytes / 1MB, 2) + " MB")
Write-Host ("Go Heap Sys:   " + [math]::Round($health.go_heap_sys_bytes / 1MB, 2) + " MB")
Write-Host ("Goroutines:    " + $health.go_num_goroutine)
Write-Host ("System CPU:    " + [math]::Round($health.cpu_percent, 2) + "%")
Write-Host ("System RAM:    " + [math]::Round($health.ram_percent, 2) + "%")
Write-Host ("Uptime:        " + $health.uptime)

$status = (Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/system/status" -Headers $headers).data
Write-Host "`n=== ACQUISITION STATUS ==="
Write-Host ("Service Status:    " + $status.service_status)
Write-Host ("Total Devices:     " + $status.total_devices)
Write-Host ("Last Device Comm:  " + $status.last_device_comm)
Write-Host ("Last Data Received:" + $status.last_data_received)
Write-Host ("Data Received Count:" + $status.data_received_count)
Write-Host ("DB Status:         " + $status.database_status)
Write-Host ("DB Latency:        " + $status.db_latency_ms + " ms")
Write-Host ("Queue Pending:     " + $status.queue_pending_records)
Write-Host ("Active Alarms:     " + $status.active_alarms)

$raw = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/telemetry/raw?page=1&page_size=1" -Headers $headers
Write-Host ("Total Raw Records: " + $raw.data.total)

$agg = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/aggregations/results?page=1&limit=1" -Headers $headers
Write-Host ("Total Aggregations:" + $agg.data.total)
