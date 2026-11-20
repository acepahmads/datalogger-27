const THEME_KEY = 'datalogger.theme';
const LEGACY_THEME_KEY = 'datalogger_theme';
const THEME_PRESET_KEY = 'datalogger.themePreset';
const DENSITY_KEY = 'datalogger.uiDensity';

export const THEME_PRESETS = [
  {
    id: 'oceanic',
    name: 'Oceanic Classic',
    badge: 'Default',
    description: 'Sky Blue & Slate bersih standar SaaS monitoring.',
    recommendedFor: 'Umum · Semua Domain',
    swatches: ['#0284C7', '#38BDF8', '#E2E8F0', '#0B1C2B'],
    previewBgLight: 'linear-gradient(135deg, #E6EFF7 0%, #F8FAFD 100%)',
    previewBgDark: 'linear-gradient(135deg, #061321 0%, #0D2232 100%)',
    previewAccent: '#0284C7',
    previewRadius: '10px',
  },
  {
    id: 'eco-forest',
    name: 'Eco Forest',
    badge: 'DLH · AQMS',
    description: 'Ivory Sage & Emerald Teal lembut bernuansa lingkungan.',
    recommendedFor: 'DLH / KLHK · AQMS & ONLIMO',
    swatches: ['#0F766E', '#10B981', '#D4E3D9', '#0C1E18'],
    previewBgLight: 'linear-gradient(135deg, #EBF3EB 0%, #F7FAF6 100%)',
    previewBgDark: 'linear-gradient(135deg, #071410 0%, #10261F 100%)',
    previewAccent: '#0F766E',
    previewRadius: '14px',
  },
  {
    id: 'industrial-scada',
    name: 'Industrial SCADA',
    badge: 'Control Room',
    description: 'Sudut presisi, kontras tinggi Amber-Cyan & angka monospace.',
    recommendedFor: 'Pabrik 24/7 · SPARING & Outbound',
    swatches: ['#06B6D4', '#F59E0B', '#263143', '#0A0D12'],
    previewBgLight: 'linear-gradient(135deg, #E3E9F0 0%, #F1F5F9 100%)',
    previewBgDark: 'linear-gradient(135deg, #0A0D12 0%, #151C28 100%)',
    previewAccent: '#F59E0B',
    previewRadius: '4px',
  },
  {
    id: 'royal-glass',
    name: 'Royal Glass',
    badge: 'Executive',
    description: 'Pearl Indigo & Frosted Glassmorphism modern melayang.',
    recommendedFor: 'Presentasi & Command Center',
    swatches: ['#4F46E5', '#0EA5E9', '#DDD6FE', '#0C0A1D'],
    previewBgLight: 'linear-gradient(135deg, #EDE9FE 0%, #E0F2FE 100%)',
    previewBgDark: 'linear-gradient(135deg, #0C0A1D 0%, #1A1740 100%)',
    previewAccent: '#4F46E5',
    previewRadius: '16px',
  },
];

export const VALID_THEME_PRESET_IDS = THEME_PRESETS.map(p => p.id);
export const VALID_DENSITY_MODES = ['comfortable', 'compact'];

export function getStoredTheme() {
  try {
    const saved = localStorage.getItem(THEME_KEY) || localStorage.getItem(LEGACY_THEME_KEY);
    if (saved && (saved === 'light' || saved === 'dark' || saved === 'system')) {
      return saved;
    }
  } catch (e) {
    console.warn('Failed to access localStorage for theme:', e);
  }
  return 'system';
}

export function getStoredThemePreset() {
  try {
    const saved = localStorage.getItem(THEME_PRESET_KEY);
    if (saved && VALID_THEME_PRESET_IDS.includes(saved)) {
      return saved;
    }
  } catch (e) {
    console.warn('Failed to access localStorage for theme preset:', e);
  }
  return 'oceanic';
}

export function getStoredDensity() {
  try {
    const saved = localStorage.getItem(DENSITY_KEY);
    if (saved && VALID_DENSITY_MODES.includes(saved)) {
      return saved;
    }
  } catch (e) {
    console.warn('Failed to access localStorage for ui density:', e);
  }
  return 'comfortable';
}

export function getSystemTheme() {
  if (typeof window !== 'undefined' && window.matchMedia) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return 'dark'; // Industrial default fallback
}

export function resolveEffectiveTheme(themeSetting) {
  if (themeSetting === 'system') {
    return getSystemTheme();
  }
  return themeSetting === 'light' ? 'light' : 'dark';
}

export function applyTheme(themeSetting) {
  const effective = resolveEffectiveTheme(themeSetting);
  if (typeof document !== 'undefined') {
    if (effective === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }
  try {
    localStorage.setItem(THEME_KEY, themeSetting);
    localStorage.setItem(LEGACY_THEME_KEY, themeSetting);
  } catch (e) {
    console.warn('Failed to persist theme:', e);
  }
  return effective;
}

export function applyThemePreset(presetId) {
  const normalized = VALID_THEME_PRESET_IDS.includes(presetId) ? presetId : 'oceanic';
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-ui-theme', normalized);
  }
  try {
    localStorage.setItem(THEME_PRESET_KEY, normalized);
  } catch (e) {
    console.warn('Failed to persist theme preset:', e);
  }
  return normalized;
}

export function applyDensity(densityMode) {
  const normalized = VALID_DENSITY_MODES.includes(densityMode) ? densityMode : 'comfortable';
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-ui-density', normalized);
  }
  try {
    localStorage.setItem(DENSITY_KEY, normalized);
  } catch (e) {
    console.warn('Failed to persist ui density:', e);
  }
  return normalized;
}

export function getThemePresetMeta(presetId) {
  return THEME_PRESETS.find(p => p.id === presetId) || THEME_PRESETS[0];
}

let mediaQueryListener = null;

export function initSystemThemeWatcher(callback) {
  if (typeof window === 'undefined' || !window.matchMedia) return;

  // Ensure preset and density attributes are always synced on boot
  applyThemePreset(getStoredThemePreset());
  applyDensity(getStoredDensity());

  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handler = (e) => {
    const currentTheme = getStoredTheme();
    if (currentTheme === 'system') {
      const effective = e.matches ? 'dark' : 'light';
      if (effective === 'dark') {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
      if (typeof callback === 'function') {
        callback(effective);
      }
    }
  };

  if (mediaQuery.addEventListener) {
    mediaQuery.addEventListener('change', handler);
  } else if (mediaQuery.addListener) {
    mediaQuery.addListener(handler);
  }
  mediaQueryListener = handler;
}
