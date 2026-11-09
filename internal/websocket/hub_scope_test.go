package websocket

import (
	"testing"
)

func TestWebSocketClientScopeIsolation(t *testing.T) {
	tid1 := uint(1)
	tid2 := uint(2)

	superGlobal := &Client{
		role:           "SUPERADMIN",
		username:       "admin",
		userTenantID:   nil,
		activeTenantID: 0,
		activeSiteID:   0,
	}
	superScopedT2 := &Client{
		role:           "SUPERADMIN",
		username:       "admin",
		userTenantID:   nil,
		activeTenantID: 2,
		activeSiteID:   0,
	}
	tenantAUser := &Client{
		role:           "ADMIN",
		username:       "admin_t1",
		userTenantID:   &tid1,
		activeTenantID: 1,
		activeSiteID:   0,
	}
	tenantBOpRestrictedSite := &Client{
		role:           "OPERATOR",
		username:       "op_t2_s20",
		userTenantID:   &tid2,
		activeTenantID: 2,
		activeSiteID:   0,
		allowedSiteIDs: []uint{20},
	}

	// 1. Message for Tenant 2, Site 20
	if !superGlobal.matchesScope(2, 20) {
		t.Fatalf("expected global SUPERADMIN to receive Tenant 2 Site 20 message")
	}
	if !superScopedT2.matchesScope(2, 20) {
		t.Fatalf("expected Tenant-2-scoped SUPERADMIN to receive Tenant 2 Site 20 message")
	}
	if tenantAUser.matchesScope(2, 20) {
		t.Fatalf("SECURITY VIOLATION: Tenant A user must NOT receive Tenant 2 WS message")
	}
	if !tenantBOpRestrictedSite.matchesScope(2, 20) {
		t.Fatalf("expected Tenant B operator allowed on Site 20 to receive Site 20 message")
	}

	// 2. Message for Tenant 2, Site 21 (not in operator's allowedSiteIDs)
	if tenantBOpRestrictedSite.matchesScope(2, 21) {
		t.Fatalf("SECURITY VIOLATION: Site-20-restricted operator must NOT receive Site 21 WS message")
	}

	// 3. Message for Tenant 1, Site 10
	if !superGlobal.matchesScope(1, 10) {
		t.Fatalf("expected global SUPERADMIN to receive Tenant 1 message")
	}
	if superScopedT2.matchesScope(1, 10) {
		t.Fatalf("expected Tenant-2-scoped SUPERADMIN NOT to receive Tenant 1 message")
	}
	if !tenantAUser.matchesScope(1, 10) {
		t.Fatalf("expected Tenant A user to receive Tenant 1 message")
	}
	if tenantBOpRestrictedSite.matchesScope(1, 10) {
		t.Fatalf("SECURITY VIOLATION: Tenant B operator must NOT receive Tenant 1 message")
	}

	// 4. Verify extractMessageScope reads fallback map fields
	extTid, extSid := extractMessageScope(WSMessage{
		Type: "device.telemetry.received",
		Data: map[string]interface{}{
			"tenant_id": float64(2),
			"site_id":   float64(20),
		},
	})
	if extTid != 2 || extSid != 20 {
		t.Fatalf("expected extracted scope (2, 20), got (%d, %d)", extTid, extSid)
	}
}

