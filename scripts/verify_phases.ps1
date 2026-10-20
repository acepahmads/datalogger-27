$loginBody = '{"username":"admin","password":"Password123!"}'

function Test-Phases($baseUrl, $name) {
    Write-Host "=== $name ($baseUrl) ==="
    try {
        $loginRes = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
        $token = $loginRes.data.token
        $headers = @{ Authorization = "Bearer $token" }

        # Get phases
        $phasesRes = Invoke-RestMethod -Uri "$baseUrl/api/phases" -Headers $headers
        $phases = $phasesRes.data
        $p8_10 = $phases | Where-Object { $_.id -in 8, 9, 10 } | Select-Object id, phase_name, progress, status, @{N='task_count';E={$_.tasks.Count}}
        $p8_10 | Format-Table -AutoSize

        # Get Phase 9 tasks
        Write-Host "Phase 9 Tasks on ${name}:"
        $p9 = $phases | Where-Object { $_.id -eq 9 }
        $tasks = $p9.tasks | Select-Object id, task_name, progress, status, assigned_to
        $tasks | Format-Table -AutoSize

        # Get progress endpoint
        $progRes = Invoke-RestMethod -Uri "$baseUrl/api/progress" -Headers $headers
        $progPhases = $progRes.data.phases
        $prog9 = $progPhases | Where-Object { $_.id -eq 9 }
        Write-Host "api/progress Phase 9: progress=$($prog9.progress), status=$($prog9.status), total_phases=$($progRes.data.total_phases)"
        Write-Host ""
    } catch {
        Write-Host "Error on $name : $($_.Exception.Message)"
    }
}

Test-Phases "http://localhost:8080" "Local Windows Host"
Test-Phases "http://192.168.1.53:8080" "Remote Raspberry Pi 4"
