$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }
$status = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/system/status" -Headers $headers
$status.data | ConvertTo-Json -Depth 5
