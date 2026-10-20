$loginBody = '{"username":"admin","password":"Password123!"}'
$loginRes = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
$headers = @{ Authorization = "Bearer " + $loginRes.data.token }

$devices = Invoke-RestMethod -Uri "http://192.168.1.53:8080/api/devices" -Method Get -Headers $headers
Write-Host "Total devices in DB:" $devices.data.total
foreach ($d in $devices.data.items) {
    Write-Host ("`n=======================================================")
    Write-Host ("DEVICE [" + $d.id + "] " + $d.name + " | Protocol: " + $d.protocol + " | Status: " + $d.status + " | Enabled: " + $d.is_enabled + " | Type: " + $d.device_type)
    Write-Host ("Serial: " + $d.serial_config.port_name + " @ " + $d.serial_config.baud_rate + " baud, SlaveID: " + $d.modbus_config.slave_id)
    Write-Host ("Parameters (" + $d.parameters.Count + "):")
    foreach ($p in $d.parameters) {
        Write-Host ("  - Param [" + $p.id + "] " + $p.code + " (" + $p.name + "): val=" + $p.current_value + " " + $p.unit + " | Quality=" + $p.current_quality + " | LastUpdated=" + $p.last_updated)
    }
}
