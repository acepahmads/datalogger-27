package sysinfo

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// NetworkInterfaceInfo holds basic metadata for a network interface
type NetworkInterfaceInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"` // "ETHERNET", "WIFI", "LOOPBACK", "OTHER"
	IP        string `json:"ip"`
	Netmask   string `json:"netmask"`
	MAC       string `json:"mac"`
	IsUp      bool   `json:"is_up"`
	IsDefault bool   `json:"is_default"`
}

// LocalNetworkStatus holds local network state and gateway health
type LocalNetworkStatus struct {
	Status           string                 `json:"status"` // "UP", "DOWN"
	PrimaryIP        string                 `json:"primary_ip"`
	PrimaryInterface string                 `json:"primary_interface"`
	Gateway          string                 `json:"gateway"`
	GatewayReachable bool                   `json:"gateway_reachable"`
	GatewayLatencyMs float64                `json:"gateway_latency_ms"`
	Interfaces       []NetworkInterfaceInfo `json:"interfaces"`
	LastChecked      time.Time              `json:"last_checked"`
	Error            string                 `json:"error,omitempty"`
}

// InternetStatus holds external connectivity state and latency
type InternetStatus struct {
	Status      string    `json:"status"` // "CONNECTED", "DEGRADED", "DISCONNECTED"
	Connected   bool      `json:"connected"`
	ProbeTarget string    `json:"probe_target"`
	LatencyMs   float64   `json:"latency_ms"`
	LastChecked time.Time `json:"last_checked"`
	Error       string    `json:"error,omitempty"`
}

// NetworkDiagnostics combines local and internet diagnostics
type NetworkDiagnostics struct {
	LocalNetwork LocalNetworkStatus `json:"local_network"`
	Internet     InternetStatus     `json:"internet"`
	Timestamp    time.Time          `json:"timestamp"`
}

var (
	netCacheMu       sync.RWMutex
	cachedNetDiag    *NetworkDiagnostics
	lastNetDiagFetch time.Time
	isNetRefreshing  bool
	netCacheTTL      = 15 * time.Second
)

// SetNetworkCacheTTL allows configuring TTL (e.g. for testing)
func SetNetworkCacheTTL(ttl time.Duration) {
	netCacheMu.Lock()
	defer netCacheMu.Unlock()
	netCacheTTL = ttl
}

// InvalidateNetworkCache forces fresh probe on next call
func InvalidateNetworkCache() {
	netCacheMu.Lock()
	defer netCacheMu.Unlock()
	cachedNetDiag = nil
	lastNetDiagFetch = time.Time{}
}

// GetNetworkDiagnostics returns cached diagnostics or triggers a fast refresh
func GetNetworkDiagnostics() NetworkDiagnostics {
	netCacheMu.RLock()
	if cachedNetDiag != nil && time.Since(lastNetDiagFetch) < netCacheTTL {
		res := *cachedNetDiag
		netCacheMu.RUnlock()
		return res
	}
	stale := cachedNetDiag
	netCacheMu.RUnlock()

	netCacheMu.Lock()
	// Double-check after acquiring write lock
	if cachedNetDiag != nil && time.Since(lastNetDiagFetch) < netCacheTTL {
		res := *cachedNetDiag
		netCacheMu.Unlock()
		return res
	}

	// If already refreshing and we have a stale cache, return stale immediately (non-blocking)
	if isNetRefreshing && stale != nil {
		res := *stale
		netCacheMu.Unlock()
		return res
	}

	isNetRefreshing = true
	netCacheMu.Unlock()

	// Perform refresh synchronously within this goroutine with strict timeouts
	fresh := probeNetworkDiagnostics()

	netCacheMu.Lock()
	cachedNetDiag = &fresh
	lastNetDiagFetch = time.Now()
	isNetRefreshing = false
	netCacheMu.Unlock()

	return fresh
}

func probeNetworkDiagnostics() NetworkDiagnostics {
	now := time.Now().UTC()

	var wg sync.WaitGroup
	var local LocalNetworkStatus
	var internet InternetStatus

	wg.Add(2)
	go func() {
		defer wg.Done()
		local = DetectLocalNetwork()
	}()
	go func() {
		defer wg.Done()
		internet = DetectInternet()
	}()
	wg.Wait()

	return NetworkDiagnostics{
		LocalNetwork: local,
		Internet:     internet,
		Timestamp:    now,
	}
}

// DetectLocalNetwork detects interfaces, default gateway, and tests gateway reachability
func DetectLocalNetwork() LocalNetworkStatus {
	now := time.Now().UTC()
	status := LocalNetworkStatus{
		Status:      "DOWN",
		LastChecked: now,
		Interfaces:  make([]NetworkInterfaceInfo, 0),
	}

	// 1. Determine primary outbound IP using UDP connection to public IP
	// This does not send actual packets; the OS kernel selects outbound route & local IP
	primaryIP := ""
	if conn, err := net.DialTimeout("udp", "1.1.1.1:80", 250*time.Millisecond); err == nil {
		if localAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			primaryIP = localAddr.IP.String()
		}
		_ = conn.Close()
	}

	// 2. Discover default gateway
	gateway := detectDefaultGateway(primaryIP)
	status.Gateway = gateway

	// 3. Enumerate physical/virtual network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		status.Error = fmt.Sprintf("failed to read network interfaces: %v", err)
		return status
	}

	hasUpInterface := false
	for _, iface := range ifaces {
		isLoopback := (iface.Flags & net.FlagLoopback) != 0
		isUp := (iface.Flags & net.FlagUp) != 0

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue // Skip IPv6 for local telemetry LAN overview
			}

			netmask := net.IP(ipNet.Mask).String()
			ipStr := ip4.String()
			isDefault := (primaryIP != "" && ipStr == primaryIP)

			if isUp && !isLoopback {
				hasUpInterface = true
			}

			if isDefault {
				status.PrimaryIP = ipStr
				status.PrimaryInterface = iface.Name
			}

			status.Interfaces = append(status.Interfaces, NetworkInterfaceInfo{
				Name:      iface.Name,
				Type:      classifyInterfaceType(iface.Name),
				IP:        ipStr,
				Netmask:   netmask,
				MAC:       iface.HardwareAddr.String(),
				IsUp:      isUp,
				IsDefault: isDefault,
			})
		}
	}

	// If primary interface wasn't selected by outbound dial, pick first UP non-loopback interface
	if status.PrimaryIP == "" && len(status.Interfaces) > 0 {
		for _, info := range status.Interfaces {
			if info.IsUp && info.Type != "LOOPBACK" && !strings.HasPrefix(info.IP, "127.") {
				status.PrimaryIP = info.IP
				status.PrimaryInterface = info.Name
				break
			}
		}
	}

	if hasUpInterface || status.PrimaryIP != "" {
		status.Status = "UP"
	}

	// 4. Test Gateway Reachability with fast probe (< 400ms)
	if status.Gateway != "" {
		reachable, latMs := testGatewayReachability(status.Gateway)
		status.GatewayReachable = reachable
		status.GatewayLatencyMs = latMs
	}

	return status
}

// detectDefaultGateway finds the default route gateway on Linux or Windows
func detectDefaultGateway(primaryIP string) string {
	if runtime.GOOS == "linux" {
		// Read /proc/net/route directly (fastest, zero subprocess fork)
		if gw := parseLinuxProcNetRoute(); gw != "" {
			return gw
		}
	}

	if runtime.GOOS == "windows" {
		if gw := parseWindowsRoutePrint(); gw != "" {
			return gw
		}
	}

	// Fallback heuristic: if primary IP is known (e.g. 192.168.1.53), default router is usually .1
	if primaryIP != "" {
		ip := net.ParseIP(primaryIP).To4()
		if ip != nil {
			return fmt.Sprintf("%d.%d.%d.1", ip[0], ip[1], ip[2])
		}
	}

	return ""
}

// parseLinuxProcNetRoute parses /proc/net/route for the 00000000 destination
func parseLinuxProcNetRoute() string {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 && fields[1] == "00000000" {
			// Gateway is in little-endian hex
			gwHex := fields[2]
			if d, err := strconv.ParseUint(gwHex, 16, 32); err == nil && d != 0 {
				ip := net.IPv4(byte(d), byte(d>>8), byte(d>>16), byte(d>>24))
				return ip.String()
			}
		}
	}
	return ""
}

// parseWindowsRoutePrint parses `route print 0.0.0.0`
func parseWindowsRoutePrint() string {
	ctx, cancel := context.WithTimeout(context.Background(), 800*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(ctx, "route", "print", "0.0.0.0")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		fields := strings.Fields(line)
		// Active Routes format: 0.0.0.0 0.0.0.0 <Gateway> <Interface> <Metric>
		if len(fields) >= 4 && fields[0] == "0.0.0.0" && fields[1] == "0.0.0.0" {
			gw := fields[2]
			if net.ParseIP(gw) != nil && gw != "0.0.0.0" && !strings.HasPrefix(gw, "127.") {
				return gw
			}
		}
	}
	return ""
}

// testGatewayReachability probes gateway port or ICMP with strict 400ms timeout
func testGatewayReachability(gateway string) (bool, float64) {
	if gateway == "" {
		return false, 0
	}

	start := time.Now()
	// Test standard router ports: 80, 53, 443, 22
	// If any connects OR returns ECONNREFUSED ("connection refused"), gateway host is reachable!
	ports := []string{"80", "53", "443"}
	for _, port := range ports {
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(gateway, port), 250*time.Millisecond)
		lat := float64(time.Since(start).Microseconds()) / 1000.0
		if err == nil {
			_ = conn.Close()
			return true, lat
		}
		// If connection is refused, the host responded with RST packet -> gateway is UP!
		if strings.Contains(strings.ToLower(err.Error()), "refused") {
			return true, lat
		}
	}

	// Fallback to ping command with short timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 350*time.Millisecond)
	defer pingCancel()

	var pingCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		pingCmd = exec.CommandContext(pingCtx, "ping", "-n", "1", "-w", "300", gateway)
	} else {
		pingCmd = exec.CommandContext(pingCtx, "ping", "-c", "1", "-W", "1", gateway)
	}

	pingStart := time.Now()
	if err := pingCmd.Run(); err == nil {
		lat := float64(time.Since(pingStart).Microseconds()) / 1000.0
		return true, lat
	}

	return false, 0
}

// DetectInternet tests public internet probes (1.1.1.1:53, 8.8.8.8:53)
func DetectInternet() InternetStatus {
	now := time.Now().UTC()
	probes := []string{
		"1.1.1.1:53",
		"8.8.8.8:53",
		"208.67.222.222:53",
	}

	for _, target := range probes {
		start := time.Now()
		conn, err := net.DialTimeout("tcp", target, 1000*time.Millisecond)
		if err == nil {
			lat := float64(time.Since(start).Microseconds()) / 1000.0
			_ = conn.Close()

			status := "CONNECTED"
			if lat > 500.0 {
				status = "DEGRADED"
			}

			return InternetStatus{
				Status:      status,
				Connected:   true,
				ProbeTarget: target,
				LatencyMs:   lat,
				LastChecked: now,
			}
		}
	}

	return InternetStatus{
		Status:      "DISCONNECTED",
		Connected:   false,
		ProbeTarget: probes[0],
		LatencyMs:   0,
		LastChecked: now,
		Error:       "All internet probes unreachable (offline mode or WAN link down)",
	}
}

// classifyInterfaceType infers interface medium
func classifyInterfaceType(name string) string {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "loopback") || strings.HasPrefix(lower, "lo") {
		return "LOOPBACK"
	}
	if strings.Contains(lower, "wi-fi") || strings.Contains(lower, "wlan") ||
		strings.Contains(lower, "wireless") || strings.HasPrefix(lower, "wl") {
		return "WIFI"
	}
	if strings.Contains(lower, "ethernet") || strings.HasPrefix(lower, "eth") ||
		strings.HasPrefix(lower, "en") || strings.Contains(lower, "lan") {
		return "ETHERNET"
	}
	return "OTHER"
}

var sensitiveURLParamPattern = regexp.MustCompile(`(?i)(token|key|secret|password|auth|sig|bearer)=([^&#\s]+)`)

// SanitizeEndpoint masks basic auth credentials and sensitive query parameters
func SanitizeEndpoint(rawURL string) string {
	if strings.TrimSpace(rawURL) == "" {
		return ""
	}

	// Check if valid URL
	parsed, err := url.Parse(rawURL)
	if err == nil && parsed.Scheme != "" && parsed.Host != "" {
		// Mask userinfo password
		if parsed.User != nil {
			if _, hasPass := parsed.User.Password(); hasPass {
				parsed.User = url.UserPassword(parsed.User.Username(), "***")
			}
		}

		// Mask sensitive query params
		query := parsed.Query()
		modified := false
		for k := range query {
			lk := strings.ToLower(k)
			if strings.Contains(lk, "token") || strings.Contains(lk, "key") ||
				strings.Contains(lk, "secret") || strings.Contains(lk, "password") ||
				strings.Contains(lk, "auth") || strings.Contains(lk, "sig") {
				query.Set(k, "***")
				modified = true
			}
		}
		if modified {
			parsed.RawQuery = query.Encode()
		}
		res := parsed.String()
		res = strings.ReplaceAll(res, "%2A%2A%2A", "***")
		return res
	}

	// Fallback regex masking for non-standard endpoints
	return sensitiveURLParamPattern.ReplaceAllString(rawURL, `$1=***`)
}

var sensitiveErrPattern = regexp.MustCompile(`(?i)(bearer\s+[A-Za-z0-9_\-\.]+|(?:token|secret|password|key)\s*[:=]\s*["']?[^"'\s,;]+["']?)`)

// SanitizeError strips sensitive tokens from error strings
func SanitizeError(msg string) string {
	if strings.TrimSpace(msg) == "" {
		return ""
	}
	return sensitiveErrPattern.ReplaceAllString(msg, `[REDACTED]`)
}

// MapOutboundStatus maps destination metrics and status to canonical health state:
// "HEALTHY", "DEGRADED", "FAILED", "DISABLED", "UNKNOWN"
func MapOutboundStatus(enabled bool, rawStatus string, successDeliveries int64, failedDeliveries int64) string {
	if !enabled {
		return "DISABLED"
	}
	if rawStatus == "ERROR" {
		if successDeliveries > 0 {
			return "DEGRADED"
		}
		return "FAILED"
	}
	if failedDeliveries > 0 && successDeliveries == 0 {
		return "FAILED"
	}
	if failedDeliveries > 0 && successDeliveries > 0 {
		return "DEGRADED"
	}
	if successDeliveries > 0 && failedDeliveries == 0 {
		return "HEALTHY"
	}
	if rawStatus == "ACTIVE" {
		return "HEALTHY"
	}
	return "UNKNOWN"
}
