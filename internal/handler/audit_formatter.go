package handler

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	reAlarmClearedManual  = regexp.MustCompile(`^Alarm\s+(\d+)\s+\(([^)]+)\)\s+cleared$`)
	reAlarmClearedAuto    = regexp.MustCompile(`^Alarm\s+(\d+)\s+\(([^)]+)\)\s+auto-cleared\s+upon\s+(.+)$`)
	reAlarmTriggered      = regexp.MustCompile(`^Alarm triggered:\s*Code=([^,]+),\s*Severity=([^,]+),\s*Message=(.+)$`)
	reAlarmAckWithNotes   = regexp.MustCompile(`^Alarm\s+(\d+)\s+\(([^)]+)\)\s+acknowledged by\s+(.+?)\s+\(Notes:\s*(.*)\)$`)
	reAlarmAckSimpleCode  = regexp.MustCompile(`^Alarm\s+(\d+)\s+\(([^)]+)\)\s+acknowledged by\s+(.+)$`)
	reAlarmAckSimple      = regexp.MustCompile(`^Alarm\s+(\d+)\s+acknowledged by\s+(.+)$`)
	reAlarmDeleted        = regexp.MustCompile(`^Alarm\s+(\d+)\s+\(([^,]+),\s*status=([^)]+)\)\s+deleted permanently by\s+(.+)$`)
	reAlarmPurged         = regexp.MustCompile(`^Purged\s+(\d+)\s+alarm history record\(s\) by\s+(.+)$`)
	reUserLogout          = regexp.MustCompile(`^User\s+'([^']+)'\s+logged out successfully$`)
	reUserSwitch          = regexp.MustCompile(`^Superadmin\s+'([^']+)'\s+switched session context to user\s+'([^']+)'\s+\(([^)]+)\)$`)
	reUserCreated         = regexp.MustCompile(`^Created user\s+'([^']+)'\s+with role\s+'([^']+)'$`)
	reUserUpdated         = regexp.MustCompile(`^Updated user profile for\s+'([^']+)'$`)
	reUserToggle          = regexp.MustCompile(`^(USER_ENABLE|USER_DISABLE)\s+status for user\s+'([^']+)'$`)
	reUserPwdReset        = regexp.MustCompile(`^Password reset for user\s+'([^']+)'$`)
	reUserDeleted         = regexp.MustCompile(`^Deleted user account\s+'([^']+)'$`)
	reRoleCreated         = regexp.MustCompile(`^Created new role\s+'([^']+)'$`)
	reRoleUpdated         = regexp.MustCompile(`^Updated metadata for role\s+'([^']+)'$`)
	reRolePermMatrix      = regexp.MustCompile(`^Updated permission matrix assignments for role\s+'([^']+)'$`)
	reRoleDeleted         = regexp.MustCompile(`^Deleted custom role\s+'([^']+)'$`)
	reCommFailed          = regexp.MustCompile(`^Communication failed:\s*(.+?)\.\s*Automatic background reconnect is active\.$`)
	reISPURecalc          = regexp.MustCompile(`^Manual ISPU recalculation triggered for period ending\s+([^\s.]+)\.\s*Result:\s*([^\s]+)\s*\(([^,]+),\s*([^)]+)\)$`)
	reISPUMappingSaved    = regexp.MustCompile(`^ISPU mapping saved:\s*Station=(\d+),\s*Param=(\d+),\s*Pollutant=([^,]+),\s*Unit=([^,]+),\s*Factor=([^,]+),\s*DailyStartHour=(\d+),\s*Enabled=(true|false)$`)
	reISPUMappingDeleted  = regexp.MustCompile(`^ISPU mapping deleted:\s*Station=(\d+),\s*Pollutant=(.+)$`)
	reBackupCreated       = regexp.MustCompile(`^Created\s+(\S+)\s+backup\s+(\S+)\s+\(size:\s*(\d+)\s*bytes`)
	reBackupDeleted       = regexp.MustCompile(`^Deleted backup\s+(\S+)\s+\((\d+)\s*bytes\)$`)
	reRestoreFailed       = regexp.MustCompile(`^Restore failed:\s*(.+)$`)
	reCatCreated          = regexp.MustCompile(`^Created parameter category\s+'([^']+)'\s+\(Code:\s*([^,]+)`)
	reCatUpdated          = regexp.MustCompile(`^Updated parameter category\s+'([^']+)'`)
	reCatToggled          = regexp.MustCompile(`^Toggled parameter category\s+'([^']+)'\s+enabled=(true|false)$`)
	reCatDeleted          = regexp.MustCompile(`^Deleted parameter category\s+'([^']+)'\s+\(Code:\s*([^)]+)\)`)
	reCatReordered        = regexp.MustCompile(`^Reordered\s+(\d+)\s+parameter categories$`)
	reCatAssigned         = regexp.MustCompile(`^Assigned\s+(\d+)\s+parameters to category\s+'([^']+)'$`)
	reRetPolicyCreated    = regexp.MustCompile(`^Created retention policy\s+'([^']+)'\s+for category\s+'([^']+)'$`)
	reRetPolicyUpdated    = regexp.MustCompile(`^Updated retention policy\s+'([^']+)'\s+\(enabled=(true|false),\s*retention_days=(\d+)\)$`)
	reRetPolicyToggled    = regexp.MustCompile(`^Toggled policy\s+(\S+)\s+enabled=(true|false)$`)
	reRetPruned           = regexp.MustCompile(`^Pruned\s+(\d+)\s+records from\s+(\S+)`)
	reStationLocUpdated   = regexp.MustCompile(`^Station location updated via\s+(\S+)\s+to Lat:\s*([^,]+),\s*Lng:\s*([^\s]+)\s+\(Accuracy:\s*([^)]+)\)$`)
	reStaleMsg            = regexp.MustCompile(`^Parameter\s+(.+?)\s+on device\s+(\S+)\s+has repeated STALE data\s+\((\d+)\s+consecutive stale samples\)$`)
	reStuckMsg            = regexp.MustCompile(`^Parameter\s+(.+?)\s+on device\s+(\S+)\s+has STUCK value:\s*([^\(]+?)\s*\((\d+)\s+identical readings over\s+([^)]+)\)$`)
	reSpikeMsg            = regexp.MustCompile(`^Parameter\s+(.+?)\s+on device\s+(\S+)\s+detected sudden SPIKE value:\s*(.+)$`)
	reQualityTailBracket  = regexp.MustCompile(`\s*\[Quality:[^\]]*\]\s*$`)
)

// FormatAuditTrailDetailsHuman converts raw technical or JSON audit trail details into
// clear, everyday Indonesian sentences understandable by non-technical users.
func FormatAuditTrailDetailsHuman(action, resource, rawDetails string) string {
	trimmed := strings.TrimSpace(rawDetails)
	if trimmed == "" {
		return humanFallbackByAction(action, resource)
	}

	// 1. JSON objects or arrays (including double-encoded before/after snapshots)
	if strings.HasPrefix(trimmed, "{") {
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &obj); err == nil {
			return formatAuditJSONObject(strings.ToUpper(strings.TrimSpace(action)), resource, obj)
		}
	}
	if strings.HasPrefix(trimmed, "[") {
		var arr []interface{}
		if err := json.Unmarshal([]byte(trimmed), &arr); err == nil {
			return formatAuditJSONArray(arr)
		}
	}

	// 2. Technical English / system strings -> natural Indonesian
	return formatAuditTechnicalText(strings.ToUpper(strings.TrimSpace(action)), resource, trimmed)
}

func formatAuditJSONObject(action, resource string, obj map[string]interface{}) string {
	// Check if this is a before/after diff payload
	_, hasBefore := obj["before"]
	_, hasAfter := obj["after"]
	if hasBefore || hasAfter {
		beforeMap := parseNestedJSONMap(obj["before"])
		afterMap := parseNestedJSONMap(obj["after"])
		return formatBeforeAfterDiff(action, resource, obj, beforeMap, afterMap)
	}

	switch action {
	case "CREATE_DEVICE":
		devLabel := buildDeviceSubject(obj)
		devType := strFromMap(obj, "type")
		if devType != "" {
			return fmt.Sprintf("Menambahkan perangkat baru %s (tipe: %s).", devLabel, devType)
		}
		return fmt.Sprintf("Menambahkan perangkat baru %s.", devLabel)

	case "CLONE_DEVICE":
		devLabel := buildDeviceSubject(obj)
		if srcID := intFromMap(obj, "cloned_from_device_id"); srcID > 0 {
			return fmt.Sprintf("Menyalin perangkat baru %s dari perangkat #%d.", devLabel, srcID)
		}
		return fmt.Sprintf("Menyalin perangkat baru %s.", devLabel)

	case "DELETE_DEVICE":
		devLabel := buildDeviceSubject(obj)
		return fmt.Sprintf("Menghapus perangkat %s.", devLabel)

	case "ENABLE_DEVICE":
		devLabel := buildDeviceSubject(obj)
		return fmt.Sprintf("Mengaktifkan perangkat %s.", devLabel)

	case "DISABLE_DEVICE":
		devLabel := buildDeviceSubject(obj)
		return fmt.Sprintf("Menonaktifkan perangkat %s.", devLabel)

	case "UPDATE_DEVICE":
		devLabel := buildDeviceSubject(obj)
		oldStatus := strFromMap(obj, "old_status")
		newStatus := strFromMap(obj, "new_status")
		if oldStatus != "" || newStatus != "" {
			return fmt.Sprintf("Mengubah status operasional perangkat %s dari %s menjadi %s.",
				devLabel, humanDeviceStatus(oldStatus), humanDeviceStatus(newStatus))
		}
		if en, ok := boolFromMap(obj, "enabled"); ok {
			if en {
				return fmt.Sprintf("Mengaktifkan perangkat %s.", devLabel)
			}
			return fmt.Sprintf("Menonaktifkan perangkat %s.", devLabel)
		}

	case "CHANGE_CONNECTION":
		devLabel := buildDeviceSubject(obj)
		proto := strFromMap(obj, "protocol")
		serialPort := strFromMap(obj, "serial_port")
		host := strFromMap(obj, "host")
		port := intFromMap(obj, "port")
		var connParts []string
		if proto != "" {
			connParts = append(connParts, "Protokol "+proto)
		}
		if serialPort != "" {
			connParts = append(connParts, "Port Serial "+serialPort)
		} else if host != "" {
			if port > 0 {
				connParts = append(connParts, fmt.Sprintf("Alamat %s:%d", host, port))
			} else {
				connParts = append(connParts, "Alamat "+host)
			}
		}
		if len(connParts) > 0 {
			return fmt.Sprintf("Mengubah pengaturan koneksi perangkat %s (%s).", devLabel, strings.Join(connParts, ", "))
		}
		return fmt.Sprintf("Mengubah pengaturan koneksi perangkat %s.", devLabel)

	case "CREATE_PARAMETER":
		paramLabel := buildParamSubject(obj)
		devSuffix := buildDeviceContextSuffix(obj)
		return fmt.Sprintf("Menambahkan parameter sensor baru %s%s.", paramLabel, devSuffix)

	case "CLONE_PARAMETER":
		paramLabel := buildParamSubject(obj)
		devSuffix := buildDeviceContextSuffix(obj)
		return fmt.Sprintf("Menyalin parameter sensor baru %s%s.", paramLabel, devSuffix)

	case "DELETE_PARAMETER":
		paramLabel := buildParamSubject(obj)
		devSuffix := buildDeviceContextSuffix(obj)
		return fmt.Sprintf("Menghapus parameter sensor %s%s.", paramLabel, devSuffix)

	case "UPDATE_PARAMETER":
		paramLabel := buildParamSubject(obj)
		devSuffix := buildDeviceContextSuffix(obj)
		if en, ok := boolFromMap(obj, "enabled"); ok && len(obj) <= 5 {
			if en {
				return fmt.Sprintf("Mengaktifkan pembacaan parameter sensor %s%s.", paramLabel, devSuffix)
			}
			return fmt.Sprintf("Menonaktifkan pembacaan parameter sensor %s%s.", paramLabel, devSuffix)
		}

	case "CREATE_AGGREGATION":
		name := strFromMap(obj, "name")
		fn := strFromMap(obj, "function")
		intervalSec := intFromMap(obj, "interval_seconds")
		if name != "" {
			var extras []string
			if fn != "" {
				extras = append(extras, "Metode "+humanAggFunction(fn))
			}
			if intervalSec > 0 {
				extras = append(extras, "Interval "+humanDurationSeconds(intervalSec))
			}
			if len(extras) > 0 {
				return fmt.Sprintf("Membuat aturan agregasi data '%s' (%s).", name, strings.Join(extras, ", "))
			}
			return fmt.Sprintf("Membuat aturan agregasi data '%s'.", name)
		}

	case "DELETE_AGGREGATION":
		name := strFromMap(obj, "name")
		if name != "" {
			return fmt.Sprintf("Menghapus aturan agregasi data '%s'.", name)
		}
	}

	// Generic clean human-readable fallback for any JSON object (never outputs braces or quotes)
	return flattenMapToHumanSentence(action, resource, obj)
}

func formatBeforeAfterDiff(action, resource string, outer, beforeMap, afterMap map[string]interface{}) string {
	mergedMeta := make(map[string]interface{})
	for k, v := range outer {
		if k != "before" && k != "after" {
			mergedMeta[k] = v
		}
	}
	for k, v := range beforeMap {
		mergedMeta[k] = v
	}
	for k, v := range afterMap {
		mergedMeta[k] = v
	}

	orderedKeys := []string{
		"parameter_name",
		"parameter_code",
		"device_name",
		"device_code",
		"name",
		"status",
		"enabled",
		"location",
		"scale",
		"offset",
		"min_value",
		"max_value",
		"warning_low",
		"warning_high",
		"quality_validation_enabled",
		"processing_enabled",
		"stale_timeout_seconds",
		"spike_detection_enabled",
		"spike_threshold",
		"spike_window_size",
		"function",
		"interval_seconds",
		"grace_period_seconds",
		"quality_policy",
		"timezone",
		"source_type",
	}

	ignoredKeys := map[string]bool{
		"id":            true,
		"created_at":    true,
		"updated_at":    true,
		"deleted_at":    true,
		"device_id":     true,
		"parameter_id":  true,
		"devices":       true,
		"parameters":    true,
		"device_ids":    true,
		"parameter_ids": true,
	}

	seen := make(map[string]bool)
	var allKeys []string
	for _, k := range orderedKeys {
		if _, inB := beforeMap[k]; inB {
			allKeys = append(allKeys, k)
			seen[k] = true
			continue
		}
		if _, inA := afterMap[k]; inA {
			allKeys = append(allKeys, k)
			seen[k] = true
		}
	}

	var extraKeys []string
	for k := range afterMap {
		if !seen[k] && !ignoredKeys[k] {
			extraKeys = append(extraKeys, k)
			seen[k] = true
		}
	}
	for k := range beforeMap {
		if !seen[k] && !ignoredKeys[k] {
			extraKeys = append(extraKeys, k)
			seen[k] = true
		}
	}
	sort.Strings(extraKeys)
	allKeys = append(allKeys, extraKeys...)

	var changes []string
	for _, k := range allKeys {
		bVal, hasB := beforeMap[k]
		aVal, hasA := afterMap[k]
		if !hasB && !hasA {
			continue
		}
		if jsonValuesEqual(bVal, aVal) {
			continue
		}
		label := humanFieldLabel(k)
		bEmpty := isNilOrEmptyValue(bVal)
		aEmpty := isNilOrEmptyValue(aVal)
		switch {
		case bEmpty && !aEmpty:
			changes = append(changes, fmt.Sprintf("%s diatur menjadi %s", label, humanFieldValue(k, aVal)))
		case !bEmpty && aEmpty:
			changes = append(changes, fmt.Sprintf("%s dihapus (sebelumnya %s)", label, humanFieldValue(k, bVal)))
		default:
			changes = append(changes, fmt.Sprintf("%s diubah dari %s menjadi %s", label, humanFieldValue(k, bVal), humanFieldValue(k, aVal)))
		}
	}

	var prefix string
	switch action {
	case "UPDATE_PARAMETER":
		paramLabel := buildParamSubject(mergedMeta)
		devSuffix := buildDeviceContextSuffix(mergedMeta)
		prefix = fmt.Sprintf("Memperbarui pengaturan parameter %s%s", paramLabel, devSuffix)
	case "UPDATE_PARAMETER_QUALITY":
		paramLabel := buildParamSubject(mergedMeta)
		devSuffix := buildDeviceContextSuffix(mergedMeta)
		prefix = fmt.Sprintf("Memperbarui batas & validasi kualitas data parameter %s%s", paramLabel, devSuffix)
	case "UPDATE_DEVICE":
		devLabel := buildDeviceSubject(mergedMeta)
		prefix = fmt.Sprintf("Memperbarui informasi perangkat %s", devLabel)
	case "UPDATE_AGGREGATION":
		aggName := strFromMap(mergedMeta, "name")
		if aggName != "" {
			prefix = fmt.Sprintf("Memperbarui aturan agregasi data '%s'", aggName)
		} else {
			prefix = "Memperbarui aturan agregasi data"
		}
	default:
		prefix = humanFallbackByAction(action, resource)
		prefix = strings.TrimSuffix(prefix, ".")
	}

	if len(changes) > 0 {
		return fmt.Sprintf("%s: %s.", prefix, strings.Join(changes, "; "))
	}

	// If no tracked snapshot fields differed (e.g., formula/register/unit updated), show concise readable state
	var summary []string
	if en, ok := boolFromMap(afterMap, "enabled"); ok {
		summary = append(summary, "Status "+humanBoolActive(en))
	}
	if sc, ok := afterMap["scale"]; ok && !isNilOrEmptyValue(sc) {
		summary = append(summary, "Skala "+humanFieldValue("scale", sc))
	}
	if off, ok := afterMap["offset"]; ok && !isNilOrEmptyValue(off) {
		summary = append(summary, "Offset "+humanFieldValue("offset", off))
	}
	if len(summary) > 0 {
		return fmt.Sprintf("%s (%s).", prefix, strings.Join(summary, ", "))
	}
	return prefix + "."
}

func formatAuditTechnicalText(action, resource, text string) string {
	// 1. Alarm Cleared (manual & auto)
	if m := reAlarmClearedManual.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Kondisi normal kembali: Alarm #%s (%s) telah selesai.", m[1], humanizeAlarmCode(m[2]))
	}
	if m := reAlarmClearedAuto.FindStringSubmatch(text); len(m) == 4 {
		id, code, reason := m[1], m[2], strings.TrimSpace(m[3])
		switch reason {
		case "recovery":
			return fmt.Sprintf("Kondisi normal kembali: Alarm #%s (%s) otomatis pulih setelah pembacaan sensor kembali normal.", id, humanizeAlarmCode(code))
		case "device disablement":
			return fmt.Sprintf("Alarm #%s (%s) ditutup otomatis karena perangkat dinonaktifkan.", id, humanizeAlarmCode(code))
		case "parameter disablement":
			return fmt.Sprintf("Alarm #%s (%s) ditutup otomatis karena parameter sensor dinonaktifkan.", id, humanizeAlarmCode(code))
		default:
			return fmt.Sprintf("Kondisi normal kembali: Alarm #%s (%s) otomatis selesai.", id, humanizeAlarmCode(code))
		}
	}

	// 2. Alarm Triggered
	if m := reAlarmTriggered.FindStringSubmatch(text); len(m) == 4 {
		code := strings.TrimSpace(m[1])
		sev := humanSeverity(strings.TrimSpace(m[2]))
		msg := humanizeAlarmMessage(strings.TrimSpace(m[3]))
		return fmt.Sprintf("Muncul alarm %s (%s): %s", sev, humanizeAlarmCode(code), msg)
	}

	// 3. Alarm Acknowledged / Deleted / Purged
	if m := reAlarmAckWithNotes.FindStringSubmatch(text); len(m) == 5 {
		notes := strings.TrimSpace(m[4])
		if notes != "" {
			return fmt.Sprintf("Alarm #%s (%s) telah dikonfirmasi oleh %s dengan catatan: %s", m[1], humanizeAlarmCode(m[2]), m[3], notes)
		}
		return fmt.Sprintf("Alarm #%s (%s) telah dikonfirmasi oleh %s.", m[1], humanizeAlarmCode(m[2]), m[3])
	}
	if m := reAlarmAckSimpleCode.FindStringSubmatch(text); len(m) == 4 {
		return fmt.Sprintf("Alarm #%s (%s) telah dikonfirmasi oleh %s.", m[1], humanizeAlarmCode(m[2]), m[3])
	}
	if m := reAlarmAckSimple.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Alarm #%s telah dikonfirmasi oleh %s.", m[1], m[2])
	}
	if m := reAlarmDeleted.FindStringSubmatch(text); len(m) == 5 {
		return fmt.Sprintf("Menghapus permanen riwayat alarm #%s (%s) oleh %s.", m[1], humanizeAlarmCode(m[2]), m[4])
	}
	if m := reAlarmPurged.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Menghapus seluruh riwayat alarm sebanyak %s data oleh %s.", m[1], m[2])
	}

	// 4. Authentication, Users & Roles
	if text == "User successfully logged in" {
		return "Pengguna berhasil masuk (login) ke dalam aplikasi."
	}
	if m := reUserLogout.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Pengguna '%s' keluar (logout) dari aplikasi.", m[1])
	}
	if m := reUserSwitch.FindStringSubmatch(text); len(m) == 4 {
		return fmt.Sprintf("Beralih sesi akun dari '%s' ke pengguna '%s' (Peran: %s).", m[1], m[2], m[3])
	}
	if m := reUserCreated.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Menambahkan akun pengguna baru '%s' dengan peran '%s'.", m[1], m[2])
	}
	if m := reUserUpdated.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Memperbarui profil akun pengguna '%s'.", m[1])
	}
	if m := reUserToggle.FindStringSubmatch(text); len(m) == 3 {
		if m[1] == "USER_ENABLE" {
			return fmt.Sprintf("Mengaktifkan akun pengguna '%s'.", m[2])
		}
		return fmt.Sprintf("Menonaktifkan akun pengguna '%s'.", m[2])
	}
	if m := reUserPwdReset.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Mengatur ulang kata sandi (reset password) untuk pengguna '%s'.", m[1])
	}
	if m := reUserDeleted.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Menghapus akun pengguna '%s'.", m[1])
	}
	if m := reRoleCreated.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Membuat peran hak akses baru '%s'.", m[1])
	}
	if m := reRoleUpdated.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Memperbarui informasi peran hak akses '%s'.", m[1])
	}
	if m := reRolePermMatrix.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Memperbarui daftar hak akses (izin fitur) untuk peran '%s'.", m[1])
	}
	if m := reRoleDeleted.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Menghapus peran hak akses '%s'.", m[1])
	}

	// 5. Device Communication
	if text == "Device communication link re-established; operational telemetry active." {
		return "Koneksi komunikasi perangkat berhasil terhubung kembali dan pembacaan data sensor aktif."
	}
	if m := reCommFailed.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Komunikasi perangkat terputus (%s). Sistem sedang mencoba menyambung ulang otomatis.", m[1])
	}

	// 6. ISPU, Categories, Backup & Retention, System Config
	if m := reISPURecalc.FindStringSubmatch(text); len(m) == 5 {
		return fmt.Sprintf("Menjalankan perhitungan ulang ISPU manual — Indeks: %s, Kategori: %s, Polutan dominan: %s.", m[2], m[3], m[4])
	}
	if m := reISPUMappingSaved.FindStringSubmatch(text); len(m) == 8 {
		stLabel := "Nonaktif"
		if m[7] == "true" {
			stLabel = "Aktif"
		}
		return fmt.Sprintf("Menyimpan pengaturan pemetaan polutan ISPU %s pada stasiun #%s (Satuan: %s, Faktor konversi: %s, Status: %s).", m[3], m[1], m[4], m[5], stLabel)
	}
	if m := reISPUMappingDeleted.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Menghapus pemetaan polutan ISPU %s pada stasiun #%s.", m[2], m[1])
	}
	if m := reBackupCreated.FindStringSubmatch(text); len(m) == 4 {
		return fmt.Sprintf("Membuat cadangan data sistem (%s) dengan kode %s.", strings.ToLower(m[1]), m[2])
	}
	if text == "Backup passed all integrity checksum and structural validations" {
		return "Pemeriksaan integritas cadangan data berhasil dan file cadangan dinyatakan valid."
	}
	if text == "Database and WAL restored and replayed successfully" {
		return "Pemulihan data sistem (restore database) berhasil diselesaikan."
	}
	if m := reRestoreFailed.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Pemulihan data sistem (restore) gagal: %s.", m[1])
	}
	if m := reBackupDeleted.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Menghapus file cadangan data %s.", m[1])
	}
	if m := reCatCreated.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Membuat kategori parameter baru '%s' (%s).", m[1], m[2])
	}
	if m := reCatUpdated.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Memperbarui pengaturan kategori parameter '%s'.", m[1])
	}
	if m := reCatToggled.FindStringSubmatch(text); len(m) == 3 {
		if m[2] == "true" {
			return fmt.Sprintf("Mengaktifkan kategori parameter '%s'.", m[1])
		}
		return fmt.Sprintf("Menonaktifkan kategori parameter '%s'.", m[1])
	}
	if m := reCatDeleted.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Menghapus kategori parameter '%s' (%s).", m[1], m[2])
	}
	if m := reCatReordered.FindStringSubmatch(text); len(m) == 2 {
		return fmt.Sprintf("Mengubah urutan tampilan %s kategori parameter.", m[1])
	}
	if m := reCatAssigned.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Memindahkan %s parameter ke dalam kategori '%s'.", m[1], m[2])
	}
	if m := reRetPolicyCreated.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Membuat aturan masa simpan data '%s' untuk kategori '%s'.", m[1], m[2])
	}
	if m := reRetPolicyUpdated.FindStringSubmatch(text); len(m) == 4 {
		st := "Nonaktif"
		if m[2] == "true" {
			st = "Aktif"
		}
		return fmt.Sprintf("Memperbarui aturan masa simpan '%s' (batas simpan %s hari, status %s).", m[1], m[3], st)
	}
	if m := reRetPolicyToggled.FindStringSubmatch(text); len(m) == 3 {
		if m[2] == "true" {
			return fmt.Sprintf("Mengaktifkan aturan masa simpan data '%s'.", m[1])
		}
		return fmt.Sprintf("Menonaktifkan aturan masa simpan data '%s'.", m[1])
	}
	if m := reRetPruned.FindStringSubmatch(text); len(m) == 3 {
		return fmt.Sprintf("Membersihkan otomatis %s data lama pada kategori %s sesuai aturan masa simpan.", m[1], m[2])
	}
	if text == "Configuration updated." {
		return "Memperbarui pengaturan konfigurasi sistem."
	}
	if strings.HasPrefix(text, "Configuration updated: ") {
		return "Memperbarui pengaturan konfigurasi sistem: " + strings.TrimPrefix(text, "Configuration updated: ")
	}
	if m := reStationLocUpdated.FindStringSubmatch(text); len(m) == 5 {
		return fmt.Sprintf("Memperbarui titik koordinat lokasi stasiun (%s) ke Lintang %s, Bujur %s (Akurasi %s).", m[1], m[2], m[3], m[4])
	}
	if text == "Task deleted by user" {
		return "Menghapus item tugas oleh pengguna."
	}

	return text
}

func humanizeAlarmMessage(msg string) string {
	cleaned := reQualityTailBracket.ReplaceAllString(msg, "")
	cleaned = strings.TrimSpace(cleaned)

	if m := reStaleMsg.FindStringSubmatch(cleaned); len(m) == 4 {
		return fmt.Sprintf("Data parameter %s pada perangkat %s tidak berubah/terhenti selama %s pembacaan berturut-turut.", m[1], m[2], m[3])
	}
	if m := reStuckMsg.FindStringSubmatch(cleaned); len(m) == 6 {
		return fmt.Sprintf("Nilai parameter %s pada perangkat %s tertahan di %s (%s pembacaan identik selama %s).", m[1], m[2], strings.TrimSpace(m[3]), m[4], m[5])
	}
	if m := reSpikeMsg.FindStringSubmatch(cleaned); len(m) == 4 {
		return fmt.Sprintf("Terdeteksi lonjakan nilai mendadak pada parameter %s (perangkat %s): %s.", m[1], m[2], m[3])
	}

	replacements := []struct{ old, new string }{
		{" dropped below warning low limit:", " turun di bawah batas peringatan minimum:"},
		{" dropped below critical low limit:", " turun di bawah batas kritis minimum:"},
		{" exceeded warning high limit:", " melebihi batas peringatan maksimum:"},
		{" exceeded critical high limit:", " melebihi batas kritis maksimum:"},
		{" escalated from warning to critical high limit:", " meningkat dari peringatan menjadi melewati batas kritis maksimum:"},
		{" escalated from warning to critical low limit:", " menurun dari peringatan menjadi melewati batas kritis minimum:"},
		{" de-escalated from critical to warning high limit:", " turun dari kritis ke batas peringatan maksimum:"},
		{" de-escalated from critical to warning low limit:", " naik dari kritis ke batas peringatan minimum:"},
		{"(Threshold:", "(Batas:"},
	}
	for _, r := range replacements {
		cleaned = strings.ReplaceAll(cleaned, r.old, r.new)
	}
	return cleaned
}

func humanizeAlarmCode(code string) string {
	c := strings.ToUpper(strings.TrimSpace(code))
	switch c {
	case "THRESHOLD_WARNING_LOW":
		return "nilai di bawah batas peringatan minimum"
	case "THRESHOLD_WARNING_HIGH":
		return "nilai melebihi batas peringatan maksimum"
	case "THRESHOLD_CRITICAL_LOW":
		return "nilai di bawah batas kritis minimum"
	case "THRESHOLD_CRITICAL_HIGH":
		return "nilai melebihi batas kritis maksimum"
	case "DATA_ABNORMALITY_STALE_DATA":
		return "data sensor tidak diperbarui/terhenti"
	case "DATA_ABNORMALITY_STUCK_VALUE":
		return "nilai pembacaan sensor tertahan sama"
	case "DATA_ABNORMALITY_SPIKE":
		return "lonjakan nilai sensor mendadak"
	case "DATA_ABNORMALITY_BAD_QUALITY":
		return "kualitas data sensor buruk"
	case "DATA_ABNORMALITY_ZERO_INVALID":
		return "nilai pembacaan nol tidak wajar"
	case "DEVICE_OFFLINE":
		return "perangkat tidak terhubung/offline"
	case "COMMUNICATION_ERROR":
		return "gangguan komunikasi perangkat"
	default:
		return strings.ToLower(strings.ReplaceAll(code, "_", " "))
	}
}

func humanSeverity(sev string) string {
	switch strings.ToUpper(strings.TrimSpace(sev)) {
	case "CRITICAL":
		return "Kritis"
	case "WARNING":
		return "Peringatan"
	case "INFO":
		return "Informasi"
	default:
		return sev
	}
}

func parseNestedJSONMap(val interface{}) map[string]interface{} {
	res := make(map[string]interface{})
	if val == nil {
		return res
	}
	switch v := val.(type) {
	case map[string]interface{}:
		return v
	case string:
		trimmed := strings.TrimSpace(v)
		if strings.HasPrefix(trimmed, "{") {
			_ = json.Unmarshal([]byte(trimmed), &res)
		}
	}
	return res
}

func buildParamSubject(m map[string]interface{}) string {
	name := strFromMap(m, "parameter_name")
	code := strFromMap(m, "parameter_code")
	switch {
	case name != "" && code != "" && !strings.EqualFold(name, code):
		return fmt.Sprintf("%s (%s)", name, code)
	case name != "":
		return name
	case code != "":
		return code
	}
	if pid := intFromMap(m, "parameter_id"); pid > 0 {
		return fmt.Sprintf("#%d", pid)
	}
	return "sensor"
}

func buildDeviceSubject(m map[string]interface{}) string {
	name := strFromMap(m, "device_name")
	code := strFromMap(m, "device_code")
	switch {
	case name != "" && code != "" && !strings.EqualFold(name, code):
		return fmt.Sprintf("%s (%s)", name, code)
	case name != "":
		return name
	case code != "":
		return code
	}
	if did := intFromMap(m, "device_id"); did > 0 {
		return fmt.Sprintf("#%d", did)
	}
	return "perangkat"
}

func buildDeviceContextSuffix(m map[string]interface{}) string {
	if code := strFromMap(m, "device_code"); code != "" {
		return fmt.Sprintf(" pada perangkat %s", code)
	}
	if did := intFromMap(m, "device_id"); did > 0 {
		return fmt.Sprintf(" pada perangkat #%d", did)
	}
	return ""
}

func humanFieldLabel(key string) string {
	switch key {
	case "parameter_name":
		return "Nama Parameter"
	case "parameter_code":
		return "Kode Parameter"
	case "device_name":
		return "Nama Perangkat"
	case "device_code":
		return "Kode Perangkat"
	case "device_id":
		return "ID Perangkat"
	case "parameter_id":
		return "ID Parameter"
	case "name":
		return "Nama"
	case "status":
		return "Status"
	case "old_status":
		return "Status Sebelumnya"
	case "new_status":
		return "Status Baru"
	case "enabled":
		return "Status Aktif"
	case "location":
		return "Lokasi"
	case "scale":
		return "Skala Kalibrasi"
	case "offset":
		return "Offset Kalibrasi"
	case "min_value":
		return "Batas Kritis Minimum"
	case "max_value":
		return "Batas Kritis Maksimum"
	case "warning_low":
		return "Batas Peringatan Minimum"
	case "warning_high":
		return "Batas Peringatan Maksimum"
	case "quality_validation_enabled":
		return "Validasi Kualitas Data"
	case "processing_enabled":
		return "Pemrosesan Data"
	case "stale_timeout_seconds":
		return "Batas Waktu Data Stale"
	case "spike_detection_enabled":
		return "Deteksi Lonjakan"
	case "spike_threshold":
		return "Ambang Lonjakan"
	case "spike_window_size":
		return "Sampel Deteksi Lonjakan"
	case "function":
		return "Metode Perhitungan"
	case "interval_seconds":
		return "Interval Waktu"
	case "grace_period_seconds":
		return "Masa Tenggang"
	case "quality_policy":
		return "Kebijakan Kualitas"
	case "timezone":
		return "Zona Waktu"
	case "source_type":
		return "Sumber Data"
	case "protocol":
		return "Protokol"
	case "connection_type":
		return "Tipe Koneksi"
	case "serial_port":
		return "Port Serial"
	case "host":
		return "Alamat Host"
	case "port":
		return "Port"
	case "status_on_label":
		return "Label Status Aktif"
	case "status_off_label":
		return "Label Status Nonaktif"
	default:
		clean := strings.ReplaceAll(key, "_", " ")
		if len(clean) > 0 {
			return strings.ToUpper(clean[:1]) + clean[1:]
		}
		return key
	}
}

func isSensitiveAuditKey(key string) bool {
	lk := strings.ToLower(strings.TrimSpace(key))
	for _, s := range []string{"password", "secret", "token", "api_key", "apikey", "bot_token", "auth_token", "bearer"} {
		if strings.Contains(lk, s) {
			return true
		}
	}
	return false
}

func humanFieldValue(key string, val interface{}) string {
	if isSensitiveAuditKey(key) && !isNilOrEmptyValue(val) {
		return "[REDACTED]"
	}
	if isNilOrEmptyValue(val) {
		return "kosong"
	}
	switch v := val.(type) {
	case bool:
		return humanBoolActive(v)
	case float64:
		if key == "interval_seconds" || key == "stale_timeout_seconds" || key == "grace_period_seconds" {
			return humanDurationSeconds(int(v))
		}
		return formatCleanFloat(v)
	case string:
		if key == "status" || key == "old_status" || key == "new_status" {
			return humanDeviceStatus(v)
		}
		if key == "function" {
			return humanAggFunction(v)
		}
		trimmed := strings.TrimSpace(v)
		if strings.HasPrefix(trimmed, "{") {
			var sub map[string]interface{}
			if err := json.Unmarshal([]byte(trimmed), &sub); err == nil {
				return flattenMapPairs(sub)
			}
		}
		return trimmed
	default:
		return fmt.Sprintf("%v", v)
	}
}

func humanBoolActive(v bool) string {
	if v {
		return "Aktif"
	}
	return "Nonaktif"
}

func humanDeviceStatus(st string) string {
	switch strings.ToUpper(strings.TrimSpace(st)) {
	case "ACTIVE":
		return "Aktif"
	case "INACTIVE":
		return "Tidak Aktif"
	case "MAINTENANCE":
		return "Pemeliharaan"
	case "DISABLED":
		return "Dinonaktifkan"
	default:
		if st == "" {
			return "-"
		}
		return st
	}
}

func humanAggFunction(fn string) string {
	switch strings.ToUpper(strings.TrimSpace(fn)) {
	case "AVG", "MEAN":
		return "Rata-rata"
	case "MIN":
		return "Nilai Minimum"
	case "MAX":
		return "Nilai Maksimum"
	case "SUM":
		return "Total Penjumlahan"
	case "LAST":
		return "Nilai Terakhir"
	default:
		return fn
	}
}

func humanDurationSeconds(sec int) string {
	if sec <= 0 {
		return "0 detik"
	}
	if sec%3600 == 0 {
		return fmt.Sprintf("%d jam", sec/3600)
	}
	if sec%60 == 0 {
		return fmt.Sprintf("%d menit", sec/60)
	}
	return fmt.Sprintf("%d detik", sec)
}

func formatCleanFloat(f float64) string {
	if math.Abs(f-math.Round(f)) < 1e-9 {
		return strconv.FormatInt(int64(math.Round(f)), 10)
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func isNilOrEmptyValue(v interface{}) bool {
	if v == nil {
		return true
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}

func jsonValuesEqual(a, b interface{}) bool {
	aEmpty := isNilOrEmptyValue(a)
	bEmpty := isNilOrEmptyValue(b)
	if aEmpty && bEmpty {
		return true
	}
	if aEmpty != bEmpty {
		return false
	}
	af, aNum := toFloat64(a)
	bf, bNum := toFloat64(b)
	if aNum && bNum {
		return math.Abs(af-bf) < 1e-9
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func toFloat64(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	}
	return 0, false
}

func strFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return strings.TrimSpace(s)
		}
		return strings.TrimSpace(fmt.Sprintf("%v", v))
	}
	return ""
}

func intFromMap(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok && v != nil {
		if f, isNum := toFloat64(v); isNum {
			return int(f)
		}
		if s, isStr := v.(string); isStr {
			if n, err := strconv.Atoi(strings.TrimSpace(s)); err == nil {
				return n
			}
		}
	}
	return 0
}

func boolFromMap(m map[string]interface{}, key string) (bool, bool) {
	if v, ok := m[key]; ok && v != nil {
		if b, isBool := v.(bool); isBool {
			return b, true
		}
	}
	return false, false
}

func flattenMapToHumanSentence(action, resource string, obj map[string]interface{}) string {
	pairs := flattenMapPairs(obj)
	prefix := strings.TrimSuffix(humanFallbackByAction(action, resource), ".")
	if pairs == "" {
		return prefix + "."
	}
	return fmt.Sprintf("%s (%s).", prefix, pairs)
}

func flattenMapPairs(obj map[string]interface{}) string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		if k == "created_at" || k == "updated_at" || k == "deleted_at" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		val := obj[k]
		if isNilOrEmptyValue(val) {
			continue
		}
		switch sub := val.(type) {
		case map[string]interface{}:
			nested := flattenMapPairs(sub)
			if nested != "" {
				parts = append(parts, fmt.Sprintf("%s: %s", humanFieldLabel(k), nested))
			}
		case []interface{}:
			arrTxt := formatAuditJSONArray(sub)
			if arrTxt != "" {
				parts = append(parts, fmt.Sprintf("%s: %s", humanFieldLabel(k), arrTxt))
			}
		default:
			parts = append(parts, fmt.Sprintf("%s: %s", humanFieldLabel(k), humanFieldValue(k, val)))
		}
	}
	return strings.Join(parts, ", ")
}

func formatAuditJSONArray(arr []interface{}) string {
	var items []string
	for _, el := range arr {
		switch v := el.(type) {
		case map[string]interface{}:
			if p := flattenMapPairs(v); p != "" {
				items = append(items, p)
			}
		default:
			s := strings.TrimSpace(fmt.Sprintf("%v", v))
			if s != "" {
				items = append(items, s)
			}
		}
	}
	return strings.Join(items, "; ")
}

func humanFallbackByAction(action, resource string) string {
	switch strings.ToUpper(strings.TrimSpace(action)) {
	case "LOGIN":
		return "Pengguna masuk ke dalam aplikasi."
	case "LOGOUT":
		return "Pengguna keluar dari aplikasi."
	case "UPDATE_PARAMETER":
		return "Memperbarui pengaturan parameter sensor."
	case "CREATE_PARAMETER":
		return "Menambahkan parameter sensor baru."
	case "DELETE_PARAMETER":
		return "Menghapus parameter sensor."
	case "UPDATE_DEVICE":
		return "Memperbarui pengaturan perangkat."
	case "CREATE_DEVICE":
		return "Menambahkan perangkat baru."
	case "DELETE_DEVICE":
		return "Menghapus perangkat."
	case "ALARM_CREATE":
		return "Alarm sistem terdeteksi."
	case "ALARM_CLEAR":
		return "Kondisi alarm telah pulih kembali normal."
	case "ALARM_ACKNOWLEDGE":
		return "Alarm telah dikonfirmasi oleh operator."
	default:
		if resource != "" {
			return fmt.Sprintf("Aktivitas %s pada %s.", strings.ToLower(strings.ReplaceAll(action, "_", " ")), resource)
		}
		return fmt.Sprintf("Aktivitas %s.", strings.ToLower(strings.ReplaceAll(action, "_", " ")))
	}
}
