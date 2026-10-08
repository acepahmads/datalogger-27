import Vue from 'vue';
import en from './locales/en';
import id from './locales/id';

const messages = {
  en,
  id,
};

const STORAGE_KEY = 'datalogger.language';
const LEGACY_STORAGE_KEY = 'datalogger_lang';

function getInitialLocale() {
  try {
    const saved = localStorage.getItem(STORAGE_KEY) || localStorage.getItem(LEGACY_STORAGE_KEY);
    if (saved && (saved === 'en' || saved === 'id')) {
      return saved;
    }
    // Browser language detection
    if (typeof navigator !== 'undefined' && navigator.language) {
      const browserLang = navigator.language.toLowerCase();
      if (browserLang.startsWith('id')) {
        return 'id';
      }
    }
  } catch (e) {
    console.warn('Failed to access localStorage for language preference:', e);
  }
  return 'en';
}

// Reactive i18n state using Vue 2 observable
export const i18nState = Vue.observable({
  locale: getInitialLocale(),
});

/**
 * Translate a dot-notated key to the current language
 * @param {string} path - e.g. 'navigation.devicesSensors'
 * @param {object} [params] - interpolation parameters e.g. { count: 10, name: 'Sensor A' }
 * @returns {string}
 */
export function t(path, params = {}) {
  const currentLocale = i18nState.locale;
  const currentDict = messages[currentLocale] || messages.en;

  const resolve = (obj, keyPath) => {
    return keyPath.split('.').reduce((acc, part) => (acc && acc[part] !== undefined ? acc[part] : null), obj);
  };

  let value = resolve(currentDict, path);

  // Fallback to English if missing in current language
  if (value === null && currentLocale !== 'en') {
    value = resolve(messages.en, path);
  }

  // Fallback to the path itself
  if (value === null || value === undefined) {
    return path;
  }

  // Parameter interpolation
  if (params && typeof params === 'object') {
    Object.keys(params).forEach((k) => {
      value = value.replace(new RegExp(`\\{${k}\\}`, 'g'), params[k]);
    });
  }

  return value;
}

/**
 * Set and persist active locale
 * @param {string} newLocale - 'en' or 'id'
 */
export function setLocale(newLocale) {
  if (newLocale !== 'en' && newLocale !== 'id') {
    console.warn(`Unsupported locale: ${newLocale}, falling back to 'en'`);
    newLocale = 'en';
  }
  i18nState.locale = newLocale;
  try {
    localStorage.setItem(STORAGE_KEY, newLocale);
    localStorage.setItem(LEGACY_STORAGE_KEY, newLocale);
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('lang', newLocale);
    }
  } catch (e) {
    console.warn('Failed to persist language preference:', e);
  }
}

export default {
  install(VueInstance) {
    VueInstance.prototype.$t = function(path, params) {
      return t(path, params);
    };
    VueInstance.prototype.$setLocale = setLocale;
    Object.defineProperty(VueInstance.prototype, '$locale', {
      get() {
        return i18nState.locale;
      },
    });
    VueInstance.filter('t', function(path, params) {
      return t(path, params);
    });
  },
};
