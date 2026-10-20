$loginBody = '{"username":"admin","password":"Password123!"}'

function Get-Phase10($baseUrl, $name) {
    Write-Host "=== Phase 10 on $name ($baseUrl) ==="
    try {
        $loginRes = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
        $token = $loginRes.data.token
        $headers = @{ Authorization = "Bearer $token" }

        $res = Invoke-RestMethod -Uri "$baseUrl/api/phases/10" -Headers $headers
        $p = $res.data
        Write-Host "Phase ID: $($p.id), Name: $($p.phase_name), Progress: $($p.progress)%, Status: $($p.status)"
        Write-Host "Tasks count: $($p.tasks.Count)"
        foreach ($t in $p.tasks) {
            Write-Host "--------------------------------------------------"
            Write-Host "Task #$($t.id): $($t.task_name) [$($t.progress)% - $($t.status)] (Assigned: $($t.assigned_to))"
            Write-Host "Description: $($t.description)"
            if ($t.notes) { Write-Host "Notes: $($t.notes)" }
            if ($t.test_result) { Write-Host "Test Result: $($t.test_result)" }
        }
    } catch {
        Write-Host "Error: $($_.Exception.Message)"
    }
}

Get-Phase10 "http://localhost:8080" "Local Windows Host"
Get-Phase10 "http://192.168.1.53:8080" "Remote Raspberry Pi 4"
