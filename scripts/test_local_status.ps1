$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$token = $loginRes.data.token
$headers = @{ Authorization = "Bearer $token" }
$status = Invoke-RestMethod -Uri "http://localhost:8080/api/system/status" -Headers $headers
$status.data | Select-Object cpu_percent, process_cpu_percent, process_rss_bytes, process_rss_mb, go_heap_alloc_bytes, go_heap_sys_bytes, go_num_goroutine, ram_percent, uptime_human | Format-List
