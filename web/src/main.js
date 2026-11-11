import Vue from 'vue';
import App from './App.vue';
import router from './router';
import store from './store';
import i18n from './i18n';
import { initSystemThemeWatcher } from './utils/theme';
import axios from 'axios';
import './assets/main.css';

Vue.config.productionTip = false;
Vue.use(i18n);

// Setup Axios authorization and Multi-Tenant / Multi-Site scope headers
axios.interceptors.request.use((config) => {
  const token = store.state.token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  const activeTenantId = Number(store.state.activeTenantId || 0);
  if (activeTenantId > 0) {
    config.headers['X-Tenant-ID'] = String(activeTenantId);
  }
  const activeSiteId = Number(store.state.activeSiteId || 0);
  if (activeSiteId > 0) {
    config.headers['X-Site-ID'] = String(activeSiteId);
  }
  return config;
});

// Response interceptor for authentication handling
axios.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      const originalRequest = error.config || {};
      const url = originalRequest.url || '';

      // Don't intercept auth login/logout endpoint errors or when user explicitly navigated to /login
      const onLoginScreen = (router.currentRoute && router.currentRoute.path === '/login') ||
        (typeof window !== 'undefined' && (window.location.hash || '').startsWith('#/login'));
      if (!url.includes('/auth/login') && !url.includes('/auth/logout') && !url.includes('/auth/me') && !onLoginScreen) {
        const sys = store.state.systemStatus || {};
        const isCloudMode = Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
        if (isCloudMode) {
          store.dispatch('closeWebSocket');
          store.commit('CLEAR_SENSITIVE_DATA');
          store.commit('CLEAR_AUTH');
          router.replace({
            path: '/login',
            query: router.currentRoute && router.currentRoute.fullPath && router.currentRoute.path !== '/login'
              ? { redirect: router.currentRoute.fullPath }
              : undefined,
          }).catch(() => { });
        } else {
          // Silently re-authenticate as default operator rather than dumping user to login screen (Edge mode)
          store.dispatch('autoLoginOperator').catch(() => { });
        }
      }
    }
    return Promise.reject(error);
  }
);

// Format date helper filter
Vue.filter('formatDate', function (value) {
  if (!value) return '';
  const date = new Date(value);
  return date.toLocaleString('en-GB', {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    day: '2-digit',
    month: 'short',
    year: 'numeric',
  });
});

new Vue({
  router,
  store,
  created() {
    // Initialize preferences and theme watcher
    this.$store.dispatch('initializePreferences');
    initSystemThemeWatcher((effective) => {
      this.$store.commit('SET_EFFECTIVE_THEME', effective);
    });

    // Check Cloud vs Edge mode first before deciding startup authentication flow
    const isLoginEntry = typeof window !== 'undefined' && (window.location.hash || '').startsWith('#/login');
    this.$store.dispatch('fetchSystemStatus').finally(() => {
      const sys = this.$store.state.systemStatus || {};
      const isCloudMode = Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');

      if (isCloudMode) {
        // Cloud mode strictly requires explicit login — never auto-login as guest/operator
        if (this.$store.state.token === 'datalogger_operator_session_token' || this.$store.state.authSource !== 'manual') {
          this.$store.commit('CLEAR_AUTH');
        }
        if (this.$store.state.token && this.$store.state.user) {
          this.$store.dispatch('fetchTenantContext');
          this.$store.dispatch('fetchDevices');
          this.$store.dispatch('fetchAlarms');
          if (this.$store.getters.isSuperAdmin) {
            this.$store.dispatch('fetchProgress');
            this.$store.dispatch('fetchTasks');
            this.$store.dispatch('fetchActivity');
          }
          this.$store.dispatch('initWebSocket');
        } else if (!isLoginEntry && (!this.$route || this.$route.path !== '/login')) {
          this.$router.replace({
            path: '/login',
            query: this.$route && this.$route.fullPath && this.$route.path !== '/login'
              ? { redirect: this.$route.fullPath }
              : undefined,
          }).catch(() => { });
        }
      } else if (!this.$store.state.token && !isLoginEntry) {
        this.$store.dispatch('autoLoginOperator').then(() => {
          this.$store.dispatch('fetchSystemStatus');
          this.$store.dispatch('fetchDevices');
          this.$store.dispatch('fetchAlarms');
          this.$store.dispatch('initWebSocket');
        }).catch(() => { });
      } else {
        this.$store.dispatch('fetchTenantContext');
        this.$store.dispatch('fetchDevices');
        this.$store.dispatch('fetchAlarms');
        if (this.$store.getters.isSuperAdmin) {
          this.$store.dispatch('fetchProgress');
          this.$store.dispatch('fetchTasks');
          this.$store.dispatch('fetchActivity');
        }
        this.$store.dispatch('initWebSocket');
      }
    });

    // Fallback telemetry polling: only poll via HTTP if WebSocket is disconnected and authenticated
    setInterval(() => {
      if (this.$store.state.token && !this.$store.state.wsConnected) {
        this.$store.dispatch('fetchSystemStatus');
      }
    }, 5000);
  },
  render: (h) => h(App),
}).$mount('#app');
