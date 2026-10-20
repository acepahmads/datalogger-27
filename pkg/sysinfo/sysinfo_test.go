package sysinfo

import (
	"math"
	"runtime"
	"testing"
	"time"
)

func TestParseVmRSS(t *testing.T) {
	mockStatus := []byte(`
Name:	datalogger
Umask:	0022
State:	S (sleeping)
Tgid:	1234
Ngid:	0
Pid:	1234
PPid:	1
TracerPid:	0
Uid:	1000	1000	1000	1000
Gid:	1000	1000	1000	1000
FDSize:	64
Groups:	1000 
VmPeak:	   35840 kB
VmSize:	   32768 kB
VmLck:	       0 kB
VmPin:	       0 kB
VmHWM:	   28672 kB
VmRSS:	   26624 kB
RssAnon:	   18432 kB
RssFile:	    8192 kB
RssShmem:	       0 kB
VmData:	   12288 kB
VmStk:	     136 kB
VmExe:	    8192 kB
VmLib:	    4096 kB
VmPTE:	      72 kB
VmSwap:	       0 kB
Threads:	12
`)

	bytes, ok := ParseVmRSS(mockStatus)
	if !ok {
		t.Fatalf("expected ParseVmRSS to succeed, got ok=false")
	}

	expectedBytes := uint64(26624 * 1024)
	if bytes != expectedBytes {
		t.Errorf("expected %d bytes, got %d bytes", expectedBytes, bytes)
	}

	// Test missing VmRSS
	emptyStatus := []byte("Name:\ttest\nPid:\t123\n")
	_, okEmpty := ParseVmRSS(emptyStatus)
	if okEmpty {
		t.Errorf("expected ParseVmRSS to return false on empty status")
	}

	// Test malformed VmRSS line
	malformed := []byte("VmRSS: not_a_number kB\n")
	_, okMalformed := ParseVmRSS(malformed)
	if okMalformed {
		t.Errorf("expected ParseVmRSS to return false on malformed number")
	}
}

func TestGetInfo_ProcessMetrics(t *testing.T) {
	info := GetInfo()

	// 1. Process CPU
	if info.ProcessCPUPercent < 0 {
		t.Errorf("ProcessCPUPercent must be >= 0, got %f", info.ProcessCPUPercent)
	}

	// 2. Process RSS
	if info.ProcessRSSBytes == 0 {
		t.Errorf("ProcessRSSBytes must be > 0")
	}
	if info.ProcessRSSMB <= 0 {
		t.Errorf("ProcessRSSMB must be > 0, got %f", info.ProcessRSSMB)
	}

	// Verify MB conversion accuracy
	expectedMB := float64(info.ProcessRSSBytes) / (1024 * 1024)
	if math.Abs(info.ProcessRSSMB-expectedMB) > 0.0001 {
		t.Errorf("ProcessRSSMB conversion mismatch: got %f, expected %f", info.ProcessRSSMB, expectedMB)
	}

	// 3. Go Runtime Metrics
	if info.GoHeapAllocBytes == 0 {
		t.Errorf("GoHeapAllocBytes must be > 0")
	}
	if info.GoHeapSysBytes == 0 {
		t.Errorf("GoHeapSysBytes must be > 0")
	}
	if info.GoHeapAllocBytes > info.GoHeapSysBytes {
		t.Errorf("GoHeapAllocBytes (%d) should not exceed GoHeapSysBytes (%d)", info.GoHeapAllocBytes, info.GoHeapSysBytes)
	}
	if info.GoNumGoroutine < 1 {
		t.Errorf("GoNumGoroutine must be >= 1, got %d", info.GoNumGoroutine)
	}
	if info.Goroutines != info.GoNumGoroutine {
		t.Errorf("Goroutines (%d) and GoNumGoroutine (%d) should match", info.Goroutines, info.GoNumGoroutine)
	}

	// 4. System RAM separation (System RAM should be significantly larger than process RSS)
	if info.RAMTotalBytes > 0 && info.ProcessRSSBytes > info.RAMTotalBytes {
		t.Errorf("Process RSS (%d) cannot exceed total system RAM (%d)", info.ProcessRSSBytes, info.RAMTotalBytes)
	}

	t.Logf("Verified Process Metrics on %s/%s:", runtime.GOOS, runtime.GOARCH)
	t.Logf("  Process RSS:       %.2f MB (%d bytes)", info.ProcessRSSMB, info.ProcessRSSBytes)
	t.Logf("  Go Heap Alloc:     %.2f MB (%d bytes)", float64(info.GoHeapAllocBytes)/1048576, info.GoHeapAllocBytes)
	t.Logf("  Go Heap Sys:       %.2f MB (%d bytes)", float64(info.GoHeapSysBytes)/1048576, info.GoHeapSysBytes)
	t.Logf("  Goroutines:        %d", info.GoNumGoroutine)
	t.Logf("  Process CPU:       %.2f%%", info.ProcessCPUPercent)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{45 * time.Second, "00m 45s"},
		{2*time.Minute + 15*time.Second, "02m 15s"},
		{3*time.Hour + 20*time.Minute + 5*time.Second, "03h 20m 05s"},
		{26*time.Hour + 30*time.Minute, "1d 02h 30m"},
	}

	for _, tc := range tests {
		actual := FormatDuration(tc.d)
		if actual != tc.expected {
			t.Errorf("FormatDuration(%v) = %q; expected %q", tc.d, actual, tc.expected)
		}
	}
}
