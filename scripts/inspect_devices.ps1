$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }

$devices = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/devices" -Method Get -Headers $headers
$devices | ConvertTo-Json -Depth 4
