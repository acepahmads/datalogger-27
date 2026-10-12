package notification

import (
	"encoding/json"
	"strings"
)

const MaskedSecret = "********"

var sensitiveKeys = []string{
	"password",
	"passwd",
	"token",
	"secret",
	"api_key",
	"apikey",
	"bot_token",
	"authorization",
	"private_key",
}

// IsSensitiveKey returns true if the key name indicates a secret
func IsSensitiveKey(key string) bool {
	k := strings.ToLower(strings.TrimSpace(key))
	for _, s := range sensitiveKeys {
		if strings.Contains(k, s) {
			return true
		}
	}
	return false
}

// SanitizeConfig parses config JSON and masks sensitive fields with "********"
func SanitizeConfig(configJSON string) map[string]interface{} {
	if strings.TrimSpace(configJSON) == "" {
		return make(map[string]interface{})
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &raw); err != nil {
		return map[string]interface{}{"raw": "invalid_json"}
	}

	sanitized := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		if IsSensitiveKey(k) {
			if strVal, ok := v.(string); ok && strVal != "" {
				sanitized[k] = MaskedSecret
			} else {
				sanitized[k] = ""
			}
		} else {
			sanitized[k] = v
		}
	}

	return sanitized
}

// MergeSecrets merges incoming config with existing stored config, preserving existing secrets if the incoming value is masked
func MergeSecrets(incomingJSON, existingJSON string) string {
	if strings.TrimSpace(incomingJSON) == "" {
		return existingJSON
	}

	var incoming map[string]interface{}
	if err := json.Unmarshal([]byte(incomingJSON), &incoming); err != nil {
		return incomingJSON
	}

	var existing map[string]interface{}
	if strings.TrimSpace(existingJSON) != "" {
		_ = json.Unmarshal([]byte(existingJSON), &existing)
	}
	if existing == nil {
		existing = make(map[string]interface{})
	}

	for k, v := range incoming {
		strVal, isStr := v.(string)
		if isStr && strVal == MaskedSecret && IsSensitiveKey(k) {
			// Retain existing secret
			if origVal, hasOrig := existing[k]; hasOrig {
				incoming[k] = origVal
			}
		}
	}

	mergedBytes, err := json.Marshal(incoming)
	if err != nil {
		return incomingJSON
	}
	return string(mergedBytes)
}
