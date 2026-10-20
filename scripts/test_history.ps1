$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }

try {
    $res = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/telemetry/history?page=1&page_size=5" -Method Get -Headers $headers
    Write-Host "Success! Total:" $res.data.total
    $res.data.items | Select-Object id, device_id, parameter_id, value, quality, recorded_at | Format-Table -AutoSize
} catch {
    Write-Host "Error:" $_.Exception.Message
    if ($_.Exception.Response) {
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($stream)
        Write-Host "Response Body:" $reader.ReadToEnd()
    }
}
