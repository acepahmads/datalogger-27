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

    // Live continuous telemetry polling every 2 seconds
    setInterval(() => {
      this.$store.dispatch('fetchSystemStatus');
    }, 2000);
  },
  render: (h) => h(App),
}).$mount('#app');
