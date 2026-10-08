import Vue from 'vue';
import App from './App.vue';
import router from './router';
import store from './store';
import axios from 'axios';
import './assets/main.css';

Vue.config.productionTip = false;

// Setup Axios authorization header
axios.interceptors.request.use((config) => {
  const token = store.state.token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor for automatic token renewal
let isRefreshing = false;
let failedQueue = [];

const processQueue = (error, token = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

axios.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;
    if (
      error.response &&
      error.response.status === 401 &&
      originalRequest &&
      !originalRequest._retry &&
      !originalRequest.url.includes('/auth/login')
    ) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return axios(originalRequest);
          })
          .catch((err) => Promise.reject(err));
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const res = await axios.post('/api/auth/login', {
          username: 'admin',
          password: 'admin123',
        });
        const newToken = res.data && res.data.data && res.data.data.token;
        if (newToken) {
          store.commit('SET_TOKEN', newToken);
          processQueue(null, newToken);
          originalRequest.headers.Authorization = `Bearer ${newToken}`;
          return axios(originalRequest);
        }
      } catch (loginErr) {
        processQueue(loginErr, null);
        return Promise.reject(loginErr);
      } finally {
        isRefreshing = false;
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
    // Apply initial theme
    const theme = this.$store.state.theme;
    if (theme === 'dark') {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }

    // Initialize initial telemetry & WebSocket
    this.$store.dispatch('fetchSystemStatus');
    this.$store.dispatch('fetchProgress');
    this.$store.dispatch('fetchTasks');
    this.$store.dispatch('fetchDevices');
    this.$store.dispatch('fetchAlarms');
    this.$store.dispatch('fetchActivity');
    this.$store.dispatch('initWebSocket');

    // Fallback telemetry polling: only poll via HTTP if WebSocket is disconnected
    setInterval(() => {
      if (!this.$store.state.wsConnected) {
        this.$store.dispatch('fetchSystemStatus');
      }
    }, 3000);
  },
  render: (h) => h(App),
}).$mount('#app');
