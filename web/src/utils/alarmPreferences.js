import audioAlert from './audioAlert';

const PREFS_KEY_PREFIX = 'datalogger.alarm_prefs.';
const ACTIVE_PREFS_MIRROR_KEY = 'datalogger.alarm_prefs.__active__';

export const DEFAULT_ALARM_PREFERENCES = {
  soundEnabled: true,
  bannerEnabled: true,
  volume: 0.7,
  repeatIntervalSeconds: 5, // 1, 3, 5, 10, 15, 30, 60, or 0 (once only)
  isMuted: false,
  muteDurationMinutes: 0, // 0 = unmuted, 5, 15, 30, 60, or -1 (indefinite)
  muteUntil: null,
};

export function getPreferencesKey(username) {
  const user = username !== undefined && username !== null && String(username).trim()
    ? String(username).trim()
    : 'default';
  return `${PREFS_KEY_PREFIX}${user}`;
}

function safeParsePrefs(raw) {
  if (!raw) return null;
  try {
    const parsed = JSON.parse(raw);
    return parsed && typeof parsed === 'object' ? parsed : null;
  } catch (e) {
    return null;
  }
}

function normalizeAlarmPreferences(parsed) {
  const merged = { ...DEFAULT_ALARM_PREFERENCES, ...(parsed || {}) };
  if (merged.isMuted) {
    if (merged.muteUntil && Date.now() > Number(merged.muteUntil)) {
      merged.isMuted = false;
      merged.muteUntil = null;
      merged.muteDurationMinutes = 0;
    } else if (!merged.muteUntil && (!merged.muteDurationMinutes || merged.muteDurationMinutes === 0)) {
      merged.muteDurationMinutes = -1;
    }
  } else {
    merged.muteUntil = null;
    merged.muteDurationMinutes = 0;
  }
  return merged;
}

export function getStoredAlarmPreferences(username) {
  if (typeof localStorage === 'undefined') {
    return { ...DEFAULT_ALARM_PREFERENCES };
  }
  try {
    const userKey = getPreferencesKey(username);
    const candidates = [];

    const directParsed = safeParsePrefs(localStorage.getItem(userKey));
    if (directParsed) candidates.push({ key: userKey, data: directParsed });

    const mirrorParsed = safeParsePrefs(localStorage.getItem(ACTIVE_PREFS_MIRROR_KEY));
    if (mirrorParsed) candidates.push({ key: ACTIVE_PREFS_MIRROR_KEY, data: mirrorParsed });

    // Legacy split-key migration (e.g. saved under numeric user.id "1" vs username "admin" vs "default")
    if (!mirrorParsed) {
      for (let i = 0; i < localStorage.length; i++) {
        const k = localStorage.key(i);
        if (k && k.startsWith(PREFS_KEY_PREFIX) && k !== userKey && k !== ACTIVE_PREFS_MIRROR_KEY) {
          const p = safeParsePrefs(localStorage.getItem(k));
          if (p) candidates.push({ key: k, data: p });
        }
      }
    }

    if (candidates.length > 0) {
      // Pick newest stamped entry; for legacy unstamped entries, prefer explicit muted / custom configs
      candidates.sort((a, b) => {
        const tA = Number(a.data._updatedAt) || 0;
        const tB = Number(b.data._updatedAt) || 0;
        if (tA !== tB) return tB - tA;
        const score = (entry) => (entry.data.isMuted ? 4 : 0) + (entry.data.soundEnabled === false ? 2 : 0) + (entry.key === userKey ? 1 : 0);
        return score(b) - score(a);
      });

      const chosen = normalizeAlarmPreferences(candidates[0].data);

      // Heal mirror & user key if needed so all components stay 100% synchronized
      if (!mirrorParsed || candidates[0].key !== userKey) {
        const healed = { ...chosen, _updatedAt: Number(candidates[0].data._updatedAt) || Date.now() };
        try {
          localStorage.setItem(userKey, JSON.stringify(healed));
          localStorage.setItem(ACTIVE_PREFS_MIRROR_KEY, JSON.stringify(healed));
        } catch (e) {}
      }

      return chosen;
    }
  } catch (e) {
    console.warn('[AlarmPreferences] Failed reading from localStorage:', e);
  }
  return { ...DEFAULT_ALARM_PREFERENCES };
}

export function saveAlarmPreferences(prefs, username) {
  const normalized = normalizeAlarmPreferences(prefs);
  const toSave = { ...normalized, _updatedAt: Date.now() };
  if (typeof localStorage !== 'undefined') {
    try {
      const key = getPreferencesKey(username);
      const serialized = JSON.stringify(toSave);
      localStorage.setItem(key, serialized);
      localStorage.setItem(ACTIVE_PREFS_MIRROR_KEY, serialized);
    } catch (e) {
      console.warn('[AlarmPreferences] Failed saving to localStorage:', e);
    }
  }
  applyAlarmPreferences(normalized);
  return normalized;
}

export function resetAlarmPreferences(username) {
  const defaults = { ...DEFAULT_ALARM_PREFERENCES };
  if (typeof localStorage !== 'undefined') {
    try {
      const key = getPreferencesKey(username);
      const stamped = JSON.stringify({ ...defaults, _updatedAt: Date.now() });
      localStorage.setItem(key, stamped);
      localStorage.setItem(ACTIVE_PREFS_MIRROR_KEY, stamped);
    } catch (e) {}
  }
  applyAlarmPreferences(defaults);
  return defaults;
}

export function applyAlarmPreferences(prefs) {
  if (!prefs) return;
  const normalized = normalizeAlarmPreferences(prefs);
  if (normalized.soundEnabled !== undefined) {
    audioAlert.setSoundEnabled(normalized.soundEnabled);
  }
  if (normalized.volume !== undefined) {
    audioAlert.setVolume(normalized.volume);
  }
  if (normalized.repeatIntervalSeconds !== undefined) {
    audioAlert.setRepeatInterval(normalized.repeatIntervalSeconds);
  }
  if (normalized.isMuted !== undefined) {
    if (normalized.isMuted && normalized.muteUntil) {
      audioAlert.setMuteUntil(true, normalized.muteUntil);
    } else if (normalized.isMuted && normalized.muteDurationMinutes && normalized.muteDurationMinutes > 0) {
      audioAlert.setMute(true, normalized.muteDurationMinutes);
    } else {
      audioAlert.setMute(!!normalized.isMuted, null);
    }
  }
}

// Immediately hydrate singleton audioAlert engine from localStorage on module load
if (typeof window !== 'undefined') {
  try {
    applyAlarmPreferences(getStoredAlarmPreferences());
  } catch (e) {}
}
