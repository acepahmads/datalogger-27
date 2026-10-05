import Vue from 'vue';
import Vuex from 'vuex';
import axios from 'axios';

Vue.use(Vuex);

const API_BASE = '/api';

export default new Vuex.Store({
  state: {
    token: localStorage.getItem('datalogger_token') || 'local-admin-token',
    user: {
      username: 'admin',
      role: 'Administrator',
      full_name: 'Lead System Engineer',
    },
    theme: localStorage.getItem('datalogger_theme') || 'dark',
    language: localStorage.getItem('datalogger_lang') || 'en',
    wsConnected: false,
    systemStatus: {
      cpu_percent: 0,
      ram_percent: 0,
      ram_used_mb: 0,
      ram_total_mb: 0,
      disk_percent: 0,
      disk_used_gb: 0,
      disk_total_gb: 0,
      uptime_human: '0m',
      os: 'Windows/Linux',
      arch: 'amd64/arm',
      hostname: 'localhost',
      app_version: 'v1.0.0-phase1',
      service_status: 'RUNNING',
      worker_status: 'READY',
      scheduler_status: 'ACTIVE',
      queue_status: 'OPTIMAL',
      database_status: 'HEALTHY',
      total_devices: 3,
      online_devices: 3,
      offline_devices: 0,
      error_devices: 0,
      last_device_comm: 'Just now',
      data_received_count: 32660,
      data_per_sec: 24.5,
      data_quality: 'EXCELLENT',
      active_alarms: 0,
      warning_alarms: 0,
      critical_alarms: 0,
      acknowledged_alarms: 0,
    },
    progress: {
      overall_percentage: 18.5,
      current_phase: 'Phase 1 — Foundation',
      current_phase_number: 1,
      current_subphase: 'Core Edge Runtime & UI Shell',
      current_task: 'Phase 1 Foundation Deliverables Completed',
      next_action: 'Proceed to Phase 2 Device Discovery & Modbus Driver',
      estimated_completion: 'Phase 1: Ready | Full System: Q4 2026',
      last_update: new Date().toISOString(),
      completed_tasks_count: 9,
      active_tasks_count: 14,
      blocked_tasks_count: 0,
      total_tasks_count: 110,
      phases: [],
    },
    tasks: [],
    devices: [],
    alarms: [],
    activities: [],
    ws: null,
  },
  mutations: {
    SET_TOKEN(state, token) {
      state.token = token;
      localStorage.setItem('datalogger_token', token);
    },
    SET_USER(state, user) {
      state.user = user;
    },
    SET_THEME(state, theme) {
      state.theme = theme;
      localStorage.setItem('datalogger_theme', theme);
      if (theme === 'dark') {
        document.documentElement.classList.add('dark');
      } else {
        document.documentElement.classList.remove('dark');
      }
    },
    SET_LANGUAGE(state, lang) {
      state.language = lang;
      localStorage.setItem('datalogger_lang', lang);
    },
    SET_WS_CONNECTED(state, connected) {
      state.wsConnected = connected;
    },
    SET_SYSTEM_STATUS(state, status) {
      state.systemStatus = { ...state.systemStatus, ...status };
    },
    SET_PROGRESS(state, progress) {
      state.progress = progress;
    },
    SET_TASKS(state, tasks) {
      state.tasks = tasks;
    },
    SET_DEVICES(state, devices) {
      state.devices = devices;
    },
    SET_ALARMS(state, alarms) {
      state.alarms = alarms;
    },
    SET_ACTIVITIES(state, activities) {
      state.activities = activities;
    },
    ADD_ACTIVITY(state, activity) {
      state.activities.unshift(activity);
      if (state.activities.length > 50) {
        state.activities.pop();
      }
    },
    UPDATE_TASK_IN_LIST(state, updatedTask) {
      const idx = state.tasks.findIndex(t => t.id === updatedTask.id);
      if (idx !== -1) {
        Vue.set(state.tasks, idx, { ...state.tasks[idx], ...updatedTask });
      }
    },
  },
  actions: {
    async fetchSystemStatus({ commit }) {
      try {
        const res = await axios.get(`${API_BASE}/system/status`);
        if (res.data && res.data.data) {
          commit('SET_SYSTEM_STATUS', res.data.data);
        }
      } catch (err) {
        console.error('Failed to fetch system status:', err);
      }
    },
    async fetchProgress({ commit }) {
      try {
        const res = await axios.get(`${API_BASE}/progress`);
        if (res.data && res.data.data) {
          commit('SET_PROGRESS', res.data.data);
        }
      } catch (err) {
        console.error('Failed to fetch progress:', err);
      }
    },
    async fetchTasks({ commit }, { phaseId = null, status = '' } = {}) {
      try {
        let url = `${API_BASE}/tasks`;
        const params = [];
        if (phaseId) params.push(`phase_id=${phaseId}`);
        if (status) params.push(`status=${status}`);
        if (params.length) url += `?${params.join('&')}`;

        const res = await axios.get(url);
        if (res.data && res.data.data) {
          commit('SET_TASKS', res.data.data);
        }
      } catch (err) {
        console.error('Failed to fetch tasks:', err);
      }
    },
    async fetchDevices({ commit }) {
      try {
        const res = await axios.get(`${API_BASE}/devices`);
        if (res.data && res.data.data) {
          commit('SET_DEVICES', res.data.data);
        }
      } catch (err) {
        console.error('Failed to fetch devices:', err);
      }
    },
    async fetchAlarms({ commit }) {
      try {
        const res = await axios.get(`${API_BASE}/alarms`);
        if (res.data && res.data.data) {
          commit('SET_ALARMS', res.data.data);
        }
      } catch (err) {
        console.error('Failed to fetch alarms:', err);
      }
    },
    async fetchActivity({ commit }) {
      try {
        const res = await axios.get(`${API_BASE}/activity?limit=30`);
        if (res.data && res.data.data) {
          commit('SET_ACTIVITIES', res.data.data);
        }
      } catch (err) {
        console.error('Failed to fetch activity:', err);
      }
    },
    async updateTask({ commit, dispatch }, { id, updates }) {
      try {
        const res = await axios.put(`${API_BASE}/tasks/${id}`, updates);
        if (res.data && res.data.data) {
          commit('UPDATE_TASK_IN_LIST', res.data.data);
          dispatch('fetchProgress');
          dispatch('fetchActivity');
          return res.data.data;
        }
      } catch (err) {
        console.error('Failed to update task:', err);
        throw err;
      }
    },
    async createTask({ dispatch }, taskData) {
      try {
        const res = await axios.post(`${API_BASE}/tasks`, taskData);
        if (res.data && res.data.data) {
          dispatch('fetchTasks');
          dispatch('fetchProgress');
          dispatch('fetchActivity');
          return res.data.data;
        }
      } catch (err) {
        console.error('Failed to create task:', err);
        throw err;
      }
    },
    async deleteTask({ dispatch }, id) {
      try {
        await axios.delete(`${API_BASE}/tasks/${id}`);
        dispatch('fetchTasks');
        dispatch('fetchProgress');
        dispatch('fetchActivity');
      } catch (err) {
        console.error('Failed to delete task:', err);
        throw err;
      }
    },
    async acknowledgeAlarm({ dispatch }, id) {
      try {
        await axios.put(`${API_BASE}/alarms/${id}/acknowledge`);
        dispatch('fetchAlarms');
        dispatch('fetchSystemStatus');
      } catch (err) {
        console.error('Failed to ack alarm:', err);
      }
    },
    initWebSocket({ commit, dispatch, state }) {
      if (state.ws) {
        try { state.ws.close(); } catch (e) {}
      }

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/ws`;

      let socket;
      try {
        socket = new WebSocket(wsUrl);
      } catch (e) {
        console.warn('WebSocket init error:', e);
        return;
      }

      state.ws = socket;

      socket.onopen = () => {
        commit('SET_WS_CONNECTED', true);
      };

      socket.onclose = () => {
        commit('SET_WS_CONNECTED', false);
        // Auto-reconnect after 3 seconds
        setTimeout(() => {
          dispatch('initWebSocket');
        }, 3000);
      };

      socket.onerror = () => {
        commit('SET_WS_CONNECTED', false);
      };

      socket.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data);
          if (msg.type === 'HEALTH_UPDATE' && msg.data) {
            commit('SET_SYSTEM_STATUS', {
              cpu_percent: msg.data.cpu_percent,
              ram_percent: msg.data.ram_percent,
              ram_used_mb: msg.data.ram_used_bytes / (1024 * 1024),
              ram_total_mb: msg.data.ram_total_bytes / (1024 * 1024),
              disk_percent: msg.data.disk_percent,
              uptime_human: msg.data.uptime_human,
              goroutines: msg.data.goroutines,
            });
          } else if (msg.type === 'TASK_UPDATE') {
            dispatch('fetchProgress');
            dispatch('fetchTasks');
          } else if (msg.type === 'ACTIVITY_LOG' && msg.data) {
            commit('ADD_ACTIVITY', msg.data);
          }
        } catch (e) {
          // ignore parse errors
        }
      };
    },
  },
});
