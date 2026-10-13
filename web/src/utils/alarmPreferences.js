import audioAlert from './audioAlert';

const PREFS_KEY_PREFIX = 'datalogger.alarm_prefs.';

export const DEFAULT_ALARM_PREFERENCES = {
  soundEnabled: true,
  bannerEnabled: true,
  volume: 0.7,
  repeatIntervalSeconds: 30, // 1, 3, 5, 10, 15, 30, 60, or 0 (once only)
  isMuted: false,
  muteDurationMinutes: 0, // 0 = unmuted, 5, 15, 30, 60, or -1 (indefinite)
  muteUntil: null,
};

export function getPreferencesKey(username) {
  const user = username && typeof username === 'string' && username.trim() ? username.trim() : 'default';
  return `${PREFS_KEY_PREFIX}${user}`;
}

export function getStoredAlarmPreferences(username) {
  try {
    const key = getPreferencesKey(username);
    const raw = localStorage.getItem(key);
    if (raw) {
      const parsed = JSON.parse(raw);
      // Check if temporary mute has expired
      if (parsed.isMuted && parsed.muteUntil && Date.now() > parsed.muteUntil) {
        parsed.isMuted = false;
        parsed.muteUntil = null;
        parsed.muteDurationMinutes = 0;
      }
      return { ...DEFAULT_ALARM_PREFERENCES, ...parsed };
    }
  } catch (e) {
    console.warn('[AlarmPreferences] Failed reading from localStorage:', e);
  }
  return { ...DEFAULT_ALARM_PREFERENCES };
}

export function saveAlarmPreferences(prefs, username) {
  try {
    const key = getPreferencesKey(username);
    const toSave = { ...DEFAULT_ALARM_PREFERENCES, ...prefs };
    localStorage.setItem(key, JSON.stringify(toSave));
  } catch (e) {
    console.warn('[AlarmPreferences] Failed saving to localStorage:', e);
  }
  applyAlarmPreferences(prefs);
  return prefs;
}

export function resetAlarmPreferences(username) {
  try {
    const key = getPreferencesKey(username);
    localStorage.removeItem(key);
  } catch (e) {}
  const defaults = { ...DEFAULT_ALARM_PREFERENCES };
  applyAlarmPreferences(defaults);
  return defaults;
}

export function applyAlarmPreferences(prefs) {
  if (!prefs) return;
  if (prefs.soundEnabled !== undefined) {
    audioAlert.setSoundEnabled(prefs.soundEnabled);
  }
  if (prefs.volume !== undefined) {
    audioAlert.setVolume(prefs.volume);
  }
  if (prefs.repeatIntervalSeconds !== undefined) {
    audioAlert.setRepeatInterval(prefs.repeatIntervalSeconds);
  }
  if (prefs.isMuted !== undefined) {
    let dur = null;
    if (prefs.muteDurationMinutes && prefs.muteDurationMinutes > 0) {
      dur = prefs.muteDurationMinutes;
    }
    audioAlert.setMute(prefs.isMuted, dur);
  }
}
