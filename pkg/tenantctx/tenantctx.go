package tenantctx

import (
	"context"
	"strconv"
	"strings"
)

const DefaultTenantID uint = 1
const DefaultSiteID uint = 1

type scopeKey struct{}

// Scope represents the resolved Multi-Tenant & Multi-Site execution context for an HTTP/WS request.
type Scope struct {
	Authenticated  bool   `json:"authenticated"`
	IsSuperAdmin   bool   `json:"is_super_admin"`
	UserID         uint   `json:"user_id"`
	Username       string `json:"username"`
	Role           string `json:"role"`
	UserTenantID   *uint  `json:"user_tenant_id,omitempty"`
	TenantID       uint   `json:"tenant_id"` // 0 = Global (all tenants, SUPERADMIN only), >0 = specific tenant
	SiteID         uint   `json:"site_id"`   // 0 = All allowed sites within tenant, >0 = specific site
	AllowedSiteIDs []uint `json:"allowed_site_ids,omitempty"`
}

// WithScope attaches a resolved tenant Scope to a standard context.Context.
func WithScope(ctx context.Context, s Scope) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, scopeKey{}, s)
}

// FromContext extracts the resolved Scope from context.Context (returns zero Scope if absent).
func FromContext(ctx context.Context) Scope {
	if ctx == nil {
		return Scope{}
	}
	if s, ok := ctx.Value(scopeKey{}).(Scope); ok {
		return s
	}
	return Scope{}
}

// WriteTenantID returns the TenantID to stamp on newly created entities.
// If TenantID > 0, returns TenantID; if UserTenantID != nil && *UserTenantID > 0, returns *UserTenantID;
// otherwise defaults to DefaultTenantID (1).
func (s Scope) WriteTenantID(explicit uint) uint {
	if !s.IsSuperAdmin && s.UserTenantID != nil && *s.UserTenantID > 0 {
		return *s.UserTenantID
	}
	if explicit > 0 {
		return explicit
	}
	if s.TenantID > 0 {
		return s.TenantID
	}
	if s.UserTenantID != nil && *s.UserTenantID > 0 {
		return *s.UserTenantID
	}
	return DefaultTenantID
}

// WriteSiteID returns the SiteID to stamp on newly created site-scoped entities.
func (s Scope) WriteSiteID(explicit uint) uint {
	if len(s.AllowedSiteIDs) > 0 {
		if explicit > 0 && s.IsSiteAllowed(explicit) {
			return explicit
		}
		if s.SiteID > 0 && s.IsSiteAllowed(s.SiteID) {
			return s.SiteID
		}
		return s.AllowedSiteIDs[0]
	}
	if explicit > 0 {
		return explicit
	}
	if s.SiteID > 0 {
		return s.SiteID
	}
	return 0
}

// IsTenantEnforced returns true when queries MUST be restricted to s.TenantID.
func (s Scope) IsTenantEnforced() bool {
	return s.TenantID > 0
}

// CanAccessTenant checks whether this scope is permitted to read/modify a record owned by recordTenantID.
func (s Scope) CanAccessTenant(recordTenantID uint) bool {
	if s.IsSuperAdmin {
		if s.TenantID > 0 && recordTenantID > 0 {
			return s.TenantID == recordTenantID
		}
		return true
	}
	if s.UserTenantID != nil && *s.UserTenantID > 0 {
		return recordTenantID == *s.UserTenantID
	}
	if s.TenantID > 0 {
		return recordTenantID == s.TenantID
	}
	return true
}

// IsSiteAllowed checks whether a specific siteID is permitted under AllowedSiteIDs (empty AllowedSiteIDs means all sites in tenant).
func (s Scope) IsSiteAllowed(siteID uint) bool {
	if siteID == 0 || len(s.AllowedSiteIDs) == 0 || s.IsSuperAdmin {
		return true
	}
	for _, id := range s.AllowedSiteIDs {
		if id == siteID {
			return true
		}
	}
	return false
}

// CanAccessRecord checks both TenantID and SiteID access for a target entity.
func (s Scope) CanAccessRecord(recordTenantID, recordSiteID uint) bool {
	if !s.CanAccessTenant(recordTenantID) {
		return false
	}
	if recordSiteID > 0 && !s.IsSiteAllowed(recordSiteID) {
		return false
	}
	return true
}

// ParseCSVUintIDs parses a comma-separated string like "1,3,5" into a deduplicated []uint slice.
func ParseCSVUintIDs(csv string) []uint {
	csv = strings.TrimSpace(csv)
	if csv == "" {
		return nil
	}
	parts := strings.Split(csv, ",")
	seen := make(map[uint]bool, len(parts))
	out := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, err := strconv.ParseUint(p, 10, 32); err == nil && v > 0 {
			uid := uint(v)
			if !seen[uid] {
				seen[uid] = true
				out = append(out, uid)
			}
		}
	}
	return out
}

// FormatCSVUintIDs formats a []uint slice into a normalized comma-separated string (e.g. "1,3,5").
func FormatCSVUintIDs(ids []uint) string {
	if len(ids) == 0 {
		return ""
	}
	seen := make(map[uint]bool, len(ids))
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		if id > 0 && !seen[id] {
			seen[id] = true
			parts = append(parts, strconv.FormatUint(uint64(id), 10))
		}
	}
	return strings.Join(parts, ",")
}

