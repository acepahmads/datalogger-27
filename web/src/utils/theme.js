const THEME_KEY = 'datalogger.theme';
const LEGACY_THEME_KEY = 'datalogger_theme';

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

let mediaQueryListener = null;

export function initSystemThemeWatcher(callback) {
  if (typeof window === 'undefined' || !window.matchMedia) return;

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
