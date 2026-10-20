# ==============================================================================
# PHASE 10.1 DEVELOPMENT PROGRESS RECONCILIATION
# Updates tasks #105, #106, #107 on local host and remote Raspberry Pi 4
# ==============================================================================

$loginBody = '{"username":"admin","password":"Password123!"}'

function Reconcile-Phase10($baseUrl, $name) {
    Write-Host "=================================================================" -ForegroundColor Cyan
    Write-Host "Reconciling Phase 10.1 on $name ($baseUrl)" -ForegroundColor Cyan
    Write-Host "=================================================================" -ForegroundColor Cyan

    try {
        $loginRes = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
        $token = $loginRes.data.token
        $headers = @{ Authorization = "Bearer $token" }

        # Get existing tasks
        $res = Invoke-RestMethod -Uri "$baseUrl/api/phases/10" -Headers $headers
        $tasks = $res.data.tasks

        $updates = @(
            @{
                ID = 105
                Name = "CPU Optimization"
                Status = "WORKING"
                Progress = 60.0
                TestResult = "IN PROGRESS: CPU baseline confirmed < 5.0% on ARM64; full multi-workload endurance sign-off pending."
                Notes = "Continuous 1s CPU sampler implemented in pkg/sysinfo/sysinfo.go; process CPU normalized across cores. Measured on live Raspberry Pi 4 ARM64 edge host (cbi): 3.13%-3.24% CPU under live 2-device acquisition (below 5.0% threshold). Local host: 0.24% CPU."
            },
            @{
                ID = 106
                Name = "RAM Optimization"
                Status = "WORKING"
                Progress = 50.0
                TestResult = "IN PROGRESS: Process RSS metrics instrumented and tested; local Windows host measured at 28.66 MB RSS; physical ARM64 deployment pending."
                Notes = "Process resident set size (ProcessRSSBytes, ProcessRSSMB) and Go heap metrics (HeapAlloc, HeapSys, NumGoroutine) instrumented in pkg/sysinfo/sysinfo.go and /api/system/status. Unit tests passing 100%. Measured local process RSS: 28.66 MB avg / 28.95 MB max (below 35.0 MB threshold). Linux ARM64 binary compiled in bin/datalogger-linux-arm64."
            },
            @{
                ID = 107
                Name = "Database Optimization"
                Status = "WORKING"
                Progress = 85.0
                TestResult = "PASSED: Query plans confirmed using composite indexes, sub-millisecond latencies, and zero lock contention."
                Notes = "Connection pooling verified (MaxOpen=25, MaxIdle=10, MaxLifetime=5m, MaxIdleTime=2m). EXPLAIN query plans evaluated on 46,861 raw records and 12,866 aggregations: index idx_raw_dev_rec (0.22ms), idx_raw_rec (0.29ms), idx_agg_res_dev_param (0.22ms). Batch commit write latency: 0.23 ms/packet. Ping latency: 0.176ms (local), 0.124ms (RPi). Zero lock contention (WaitCount=0)."
            }
        )

        foreach ($u in $updates) {
            $t = $tasks | Where-Object { $_.id -eq $u.ID -or $_.task_name -eq $u.Name } | Select-Object -First 1
            if ($t) {
                $targetID = $t.id
                $body = @{
                    task_name = $t.task_name
                    description = $t.description
                    status = $u.Status
                    progress = $u.Progress
                    priority = $t.priority
                    notes = $u.Notes
                    test_result = $u.TestResult
                } | ConvertTo-Json

                $updateRes = Invoke-RestMethod -Uri "$baseUrl/api/tasks/$targetID" -Method Put -Body $body -ContentType "application/json" -Headers $headers
                Write-Host "Updated Task #$targetID ($($u.Name)): Status=$($u.Status), Progress=$($u.Progress)%" -ForegroundColor Green
            } else {
                Write-Host "Warning: Task $($u.Name) not found in Phase 10 on $name" -ForegroundColor Yellow
            }
        }

        # Query updated Phase 10
        $resUpdated = Invoke-RestMethod -Uri "$baseUrl/api/phases/10" -Headers $headers
        $p = $resUpdated.data
        Write-Host ""
        Write-Host "Updated Phase 10 Aggregate: Progress=$($p.progress)%, Status=$($p.status)" -ForegroundColor Cyan

        # Query all phases to verify Phase 8 & 9 untouched
        $allPhasesRes = Invoke-RestMethod -Uri "$baseUrl/api/phases" -Headers $headers
        $p8_10 = $allPhasesRes.data | Where-Object { $_.id -in 8, 9, 10 } | Select-Object id, phase_name, progress, status, @{N='tasks';E={$_.tasks.Count}}
        $p8_10 | Format-Table -AutoSize
    } catch {
        Write-Host "Error reconciling on $name : $($_.Exception.Message)" -ForegroundColor Red
    }
}

Reconcile-Phase10 "http://localhost:8080" "Local Development Host"
Reconcile-Phase10 "http://192.168.1.53:8080" "Remote Raspberry Pi 4"
