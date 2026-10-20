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

// Setup Axios authorization header
axios.interceptors.request.use((config) => {
  const token = store.state.token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
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

      // Don't intercept auth login/logout endpoint errors
      if (!url.includes('/auth/login') && !url.includes('/auth/logout')) {
        // Token expired or invalid: purge session & redirect to login
        store.commit('CLEAR_SENSITIVE_DATA');
        store.commit('CLEAR_AUTH');
        store.dispatch('closeWebSocket');

        if (router.currentRoute.path !== '/login') {
          router.replace({
            path: '/login',
            query: { redirect: router.currentRoute.fullPath },
          }).catch(() => {});
        }
      }
    }
    return Promise.reject(error);
  }
);

// Format date helper filter
Vue.filter('formatDate', function(value) {
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

    // Initialize initial telemetry & WebSocket only if authenticated
    if (this.$store.state.token) {
      this.$store.dispatch('fetchSystemStatus');
      this.$store.dispatch('fetchDevices');
      this.$store.dispatch('fetchAlarms');
      if (this.$store.getters.isSuperAdmin) {
        this.$store.dispatch('fetchProgress');
        this.$store.dispatch('fetchTasks');
        this.$store.dispatch('fetchActivity');
      }
      this.$store.dispatch('initWebSocket');
    }

    // Fallback telemetry polling: only poll via HTTP if WebSocket is disconnected and authenticated
    setInterval(() => {
      if (this.$store.state.token && !this.$store.state.wsConnected) {
        this.$store.dispatch('fetchSystemStatus');
      }
    }, 5000);
  },
  render: (h) => h(App),
}).$mount('#app');
