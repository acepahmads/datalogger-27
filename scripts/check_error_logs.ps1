$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }

$errLogs = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/logs?level=ERROR&page=1&page_size=10" -Method Get -Headers $headers
Write-Host "Total ERROR logs:" $errLogs.data.total
$errLogs.data.items | Select-Object id, level, message, created_at | Format-Table -AutoSize
