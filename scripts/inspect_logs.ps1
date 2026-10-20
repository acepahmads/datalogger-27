$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }

$logs = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/logs?page=1&page_size=20" -Method Get -Headers $headers
Write-Host "Total logs in DB:" $logs.data.total
if ($logs.data.items) {
    $logs.data.items | Select-Object id, level, message, created_at | Format-Table -AutoSize
} elseif ($logs.data.logs) {
    $logs.data.logs | Select-Object id, level, message, created_at | Format-Table -AutoSize
} else {
    $logs.data | ConvertTo-Json -Depth 3
}
