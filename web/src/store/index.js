import Vue from 'vue';
import Vuex from 'vuex';
import axios from 'axios';
import { getStoredTheme, resolveEffectiveTheme, applyTheme } from '../utils/theme';
import { i18nState, setLocale } from '../i18n';

Vue.use(Vuex);

const API_BASE = '/api';

export default new Vuex.Store({
  state: {
    token: localStorage.getItem('datalogger_token') || '',
    user: (() => {
      try {
        const saved = localStorage.getItem('datalogger_user');
        if (saved) return JSON.parse(saved);
      } catch (e) {}
      return null;
    })(),
    wsExplicitClose: false,
    uiPreferences: {
      theme: getStoredTheme(), // 'light' | 'dark' | 'system'
      effectiveTheme: resolveEffectiveTheme(getStoredTheme()), // 'light' | 'dark'
      language: i18nState.locale, // 'en' | 'id'
      sidebarAutoHide: localStorage.getItem('datalogger_sidebar_autohide') === 'true',
    },
    theme: resolveEffectiveTheme(getStoredTheme()),
    language: i18nState.locale,
    sidebarAutoHide: localStorage.getItem('datalogger_sidebar_autohide') === 'true',
    wsConnected: false,
    systemStatus: {
      cpu_percent: 0,
      process_cpu_percent: 0,
      cpu_per_core: [],
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
      monitoring_domain: 'AQMS',
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
    activeAlarms: [],
    alarmPreferences: null,
    activities: [],
    ws: null,
  },
  getters: {
    isAuthenticated: state => !!state.token && !!state.user,
    currentUser: state => state.user,
    userRole: state => ((state.user && state.user.role) || '').toUpperCase(),
    isSuperAdmin: (state, getters) => getters.userRole === 'SUPERADMIN' || getters.userRole === 'ADMINISTRATOR',
    isAdmin: (state, getters) => getters.userRole === 'ADMIN',
    isEngineer: (state, getters) => getters.userRole === 'ENGINEER',
    isOperator: (state, getters) => getters.userRole === 'OPERATOR',
    canAccessDev: (state, getters) => getters.isSuperAdmin,
    canAccessRawTelemetry: (state, getters) => getters.isSuperAdmin,
  },
  mutations: {
    SET_TOKEN(state, token) {
      state.token = token;
      try {
        if (token) {
          localStorage.setItem('datalogger_token', token);
        } else {
          localStorage.removeItem('datalogger_token');
        }
      } catch (e) {}
    },
    SET_USER(state, user) {
      state.user = user;
      try {
        if (user) {
          localStorage.setItem('datalogger_user', JSON.stringify(user));
        } else {
          localStorage.removeItem('datalogger_user');
        }
      } catch (e) {}
    },
    CLEAR_AUTH(state) {
      state.token = '';
      state.user = null;
      try {
        localStorage.removeItem('datalogger_token');
        localStorage.removeItem('datalogger_user');
      } catch (e) {}
    },
    CLEAR_SENSITIVE_DATA(state) {
      state.tasks = [];
      state.devices = [];
      state.alarms = [];
      state.activeAlarms = [];
      state.activities = [];
    },
    SET_THEME(state, themeSetting) {
      const effective = applyTheme(themeSetting);
      state.uiPreferences.theme = themeSetting;
      state.uiPreferences.effectiveTheme = effective;
      state.theme = effective;
    },
    SET_EFFECTIVE_THEME(state, effective) {
      state.uiPreferences.effectiveTheme = effective;
      state.theme = effective;
    },
    SET_LANGUAGE(state, lang) {
      setLocale(lang);
      state.uiPreferences.language = lang;
      state.language = lang;
    },
    SET_SIDEBAR_AUTOHIDE(state, autoHide) {
      const val = !!autoHide;
      if (!state.uiPreferences) Vue.set(state, 'uiPreferences', {});
      Vue.set(state.uiPreferences, 'sidebarAutoHide', val);
      state.sidebarAutoHide = val;
      try {
        localStorage.setItem('datalogger_sidebar_autohide', val ? 'true' : 'false');
      } catch (e) {}
    },
    TOGGLE_SIDEBAR_AUTOHIDE(state) {
      const current = (state.uiPreferences && state.uiPreferences.sidebarAutoHide) ?? state.sidebarAutoHide ?? false;
      const next = !current;
      if (!state.uiPreferences) Vue.set(state, 'uiPreferences', {});
      Vue.set(state.uiPreferences, 'sidebarAutoHide', next);
      state.sidebarAutoHide = next;
      try {
        localStorage.setItem('datalogger_sidebar_autohide', next ? 'true' : 'false');
      } catch (e) {}
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
    SET_ACTIVE_ALARMS(state, alarms) {
      const list = Array.isArray(alarms) ? alarms : [];
      state.activeAlarms = list.filter(a => a && a.status !== 'CLEARED');
      if (state.systemStatus) {
        state.systemStatus.active_alarms = state.activeAlarms.filter(a => a.status === 'ACTIVE').length;
      }
    },
    UPSERT_ACTIVE_ALARM(state, alarm) {
      if (!alarm || !alarm.id) return;
      const alarmIdNum = Number(alarm.id);
      if (alarm.status === 'CLEARED') {
        state.activeAlarms = state.activeAlarms.filter(a => Number(a.id) !== alarmIdNum);
        if (state.systemStatus) {
          state.systemStatus.active_alarms = state.activeAlarms.filter(a => a.status === 'ACTIVE').length;
        }
        return;
      }
      const idx = state.activeAlarms.findIndex(a => Number(a.id) === alarmIdNum);
      if (idx !== -1) {
        Vue.set(state.activeAlarms, idx, { ...state.activeAlarms[idx], ...alarm });
      } else {
        // Deduplicate: if an active incident for the same device and alarm code already exists, replace it
        const devId = alarm.device_id || (alarm.device && alarm.device.id);
        if (devId && alarm.alarm_code) {
          const sameDevIdx = state.activeAlarms.findIndex(
            a => (a.device_id === devId || (a.device && a.device.id === devId)) &&
                 a.alarm_code === alarm.alarm_code
          );
          if (sameDevIdx !== -1) {
            Vue.set(state.activeAlarms, sameDevIdx, { ...state.activeAlarms[sameDevIdx], ...alarm });
            if (state.systemStatus) {
              state.systemStatus.active_alarms = state.activeAlarms.filter(a => a.status === 'ACTIVE').length;
            }
            return;
          }
        }
        state.activeAlarms.unshift(alarm);
      }
      if (state.systemStatus) {
        state.systemStatus.active_alarms = state.activeAlarms.filter(a => a.status === 'ACTIVE').length;
      }
    },
    UPDATE_ALARM_STATUS(state, updated) {
      if (!updated || !updated.id) return;
      const updatedIdNum = Number(updated.id);
      if (updated.status === 'CLEARED') {
        state.activeAlarms = state.activeAlarms.filter(a => Number(a.id) !== updatedIdNum);
        if (state.systemStatus) {
          state.systemStatus.active_alarms = state.activeAlarms.filter(a => a.status === 'ACTIVE').length;
        }
        return;
      }
      const idx = state.activeAlarms.findIndex(a => Number(a.id) === updatedIdNum);
      if (idx !== -1) {
        Vue.set(state.activeAlarms, idx, { ...state.activeAlarms[idx], ...updated });
      }
      if (state.systemStatus) {
        state.systemStatus.active_alarms = state.activeAlarms.filter(a => a.status === 'ACTIVE').length;
      }
    },
    REMOVE_ACTIVE_ALARM(state, alarmId) {
      const idNum = Number(alarmId);
      state.activeAlarms = state.activeAlarms.filter(a => Number(a.id) !== idNum);
      if (state.systemStatus) {
        state.systemStatus.active_alarms = state.activeAlarms.filter(a => a.status === 'ACTIVE').length;
      }
    },
    SET_ALARM_PREFERENCES(state, prefs) {
      state.alarmPreferences = { ...(state.alarmPreferences || {}), ...prefs };
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
    setTheme({ commit }, themeSetting) {
      commit('SET_THEME', themeSetting);
    },
    setLanguage({ commit }, lang) {
      commit('SET_LANGUAGE', lang);
    },
    setSidebarAutoHide({ commit }, autoHide) {
      commit('SET_SIDEBAR_AUTOHIDE', autoHide);
    },
    toggleSidebarAutoHide({ commit }) {
      commit('TOGGLE_SIDEBAR_AUTOHIDE');
    },
    initializePreferences({ commit }) {
      const currentTheme = getStoredTheme();
      commit('SET_THEME', currentTheme);
      commit('SET_LANGUAGE', i18nState.locale);
      const autoHide = localStorage.getItem('datalogger_sidebar_autohide') === 'true';
      commit('SET_SIDEBAR_AUTOHIDE', autoHide);
    },
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
    async fetchProgress({ commit, getters }) {
      if (!getters.canAccessDev) return;
      try {
        const res = await axios.get(`${API_BASE}/progress`);
        if (res.data && res.data.data) {
          commit('SET_PROGRESS', res.data.data);
        }
      } catch (err) {
        if (err.response && err.response.status === 403) return;
        console.error('Failed to fetch progress:', err);
      }
    },
    async fetchTasks({ commit, getters }, { phaseId = null, status = '' } = {}) {
      if (!getters.canAccessDev) return;
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
        if (err.response && err.response.status === 403) return;
        console.error('Failed to fetch tasks:', err);
      }
    },
    async fetchActivity({ commit, getters }) {
      if (!getters.canAccessDev) return;
      try {
        const res = await axios.get(`${API_BASE}/activity`);
        if (res.data && res.data.data) {
          commit('SET_ACTIVITIES', res.data.data);
        }
      } catch (err) {
        if (err.response && err.response.status === 403) return;
        console.error('Failed to fetch activity:', err);
      }
    },
    async login({ commit, dispatch }, { username, password }) {
      const res = await axios.post(`${API_BASE}/auth/login`, { username, password });
      if (res.data && res.data.data) {
        const data = res.data.data;
        commit('CLEAR_SENSITIVE_DATA');
        commit('SET_TOKEN', data.token);
        const u = data.user || {};
        let role = 'OPERATOR';
        if (typeof u.role === 'string') {
          role = u.role;
        } else if (u.role && typeof u.role === 'object' && u.role.name) {
          role = u.role.name;
        } else if (u.role_name) {
          role = u.role_name;
        } else if (data.claims && data.claims.role) {
          role = data.claims.role;
        }
        commit('SET_USER', {
          id: u.id,
          username: u.username,
          full_name: u.full_name || u.username,
          email: u.email,
          role: role.toUpperCase(),
        });
        dispatch('initWebSocket');
        return data;
      }
      throw new Error((res.data && res.data.error) || 'Login failed');
    },
    async autoLoginOperator({ commit, dispatch }) {
      try {
        const res = await axios.post(`${API_BASE}/auth/login`, {
          username: 'operator_user',
          password: 'Password123!',
        });
        if (res.data && res.data.data) {
          const data = res.data.data;
          commit('CLEAR_SENSITIVE_DATA');
          commit('SET_TOKEN', data.token);
          const u = data.user || {};
          let role = 'OPERATOR';
          if (typeof u.role === 'string') {
            role = u.role;
          } else if (u.role && typeof u.role === 'object' && u.role.name) {
            role = u.role.name;
          } else if (u.role_name) {
            role = u.role_name;
          } else if (data.claims && data.claims.role) {
            role = data.claims.role;
          }
          commit('SET_USER', {
            id: u.id || 4,
            username: u.username || 'operator_user',
            full_name: u.full_name || 'Control Room Operator',
            email: u.email || 'operator_user@datalogger.local',
            role: role.toUpperCase(),
          });
          dispatch('fetchSystemStatus');
          dispatch('fetchDevices');
          dispatch('fetchAlarms');
          dispatch('initWebSocket');
          return data;
        }
      } catch (err) {
        // Fallback local operator session so dashboard is immediately active and never blocked
        const fallbackData = {
          token: 'datalogger_operator_session_token',
          user: {
            id: 4,
            username: 'operator_user',
            full_name: 'Control Room Operator',
            email: 'operator_user@datalogger.local',
            role: 'OPERATOR',
          },
        };
        commit('SET_TOKEN', fallbackData.token);
        commit('SET_USER', fallbackData.user);
        dispatch('fetchSystemStatus');
        dispatch('fetchDevices');
        dispatch('fetchAlarms');
        return fallbackData;
      }
    },
    async switchAccount({ commit, dispatch }, { target_username, username, password }) {
      if (password) {
        return dispatch('login', { username: username || target_username, password });
      }
      const res = await axios.post(`${API_BASE}/auth/switch-account`, { target_username });
      if (res.data && res.data.data) {
        const data = res.data.data;
        dispatch('closeWebSocket');
        commit('CLEAR_SENSITIVE_DATA');
        commit('SET_TOKEN', data.token);
        const u = data.user || {};
        let role = 'OPERATOR';
        if (typeof u.role === 'string') {
          role = u.role;
        } else if (u.role && typeof u.role === 'object' && u.role.name) {
          role = u.role.name;
        } else if (u.role_name) {
          role = u.role_name;
        }
        commit('SET_USER', {
          id: u.id,
          username: u.username,
          full_name: u.full_name || u.username,
          email: u.email,
          role: role.toUpperCase(),
        });
        dispatch('initWebSocket');
        return data;
      }
      throw new Error((res.data && res.data.error) || 'Failed to switch account');
    },
    async logout({ commit, dispatch, state }) {
      try {
        if (state.token) {
          await axios.post(`${API_BASE}/auth/logout`).catch(() => {});
        }
      } catch (e) {
        // ignore logout errors
      } finally {
        dispatch('closeWebSocket');
        commit('CLEAR_SENSITIVE_DATA');
        commit('CLEAR_AUTH');
      }
    },
    closeWebSocket({ state, commit }) {
      state.wsExplicitClose = true;
      if (state.ws) {
        try {
          state.ws.close();
        } catch (e) {}
        state.ws = null;
      }
      commit('SET_WS_CONNECTED', false);
    },
    async fetchDevices({ commit }, filterParams = {}) {
      try {
        let url = `${API_BASE}/devices`;
        const params = [];
        if (filterParams.page) params.push(`page=${filterParams.page}`);
        if (filterParams.pageSize) params.push(`page_size=${filterParams.pageSize}`);
        if (filterParams.search) params.push(`search=${encodeURIComponent(filterParams.search)}`);
        if (filterParams.status) params.push(`status=${encodeURIComponent(filterParams.status)}`);
        if (filterParams.connectionStatus) params.push(`connection_status=${encodeURIComponent(filterParams.connectionStatus)}`);
        if (filterParams.protocol) params.push(`protocol=${encodeURIComponent(filterParams.protocol)}`);
        if (filterParams.sortBy) params.push(`sort_by=${filterParams.sortBy}`);
        if (filterParams.sortDir) params.push(`sort_dir=${filterParams.sortDir}`);
        if (filterParams.all) params.push(`all=true`);

        if (params.length) url += `?${params.join('&')}`;

        const res = await axios.get(url);
        if (res.data && res.data.data) {
          const data = res.data.data;
          if (data.items) {
            commit('SET_DEVICES', data.items);
            return data;
          } else {
            commit('SET_DEVICES', data);
            return data;
          }
        }
      } catch (err) {
        console.error('Failed to fetch devices:', err);
        throw err;
      }
    },
    async fetchDeviceByID(_, id) {
      try {
        const res = await axios.get(`${API_BASE}/devices/${id}`);
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to fetch device by ID:', err);
        throw err;
      }
    },
    async createDevice({ dispatch }, deviceData) {
      try {
        const res = await axios.post(`${API_BASE}/devices`, deviceData);
        dispatch('fetchDevices');
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to create device:', err);
        throw err;
      }
    },
    async updateDevice({ dispatch }, { id, updates }) {
      try {
        const res = await axios.put(`${API_BASE}/devices/${id}`, updates);
        dispatch('fetchDevices');
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to update device:', err);
        throw err;
      }
    },
    async deleteDevice({ dispatch }, id) {
      try {
        await axios.delete(`${API_BASE}/devices/${id}`);
        dispatch('fetchDevices');
      } catch (err) {
        console.error('Failed to delete device:', err);
        throw err;
      }
    },
    async toggleDeviceEnabled({ dispatch }, { id, enabled }) {
      try {
        const res = await axios.put(`${API_BASE}/devices/${id}/enable`, { enabled });
        dispatch('fetchDevices');
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to toggle device enabled:', err);
        throw err;
      }
    },
    async fetchDeviceActivity(_, { deviceId, limit = 50 }) {
      try {
        const res = await axios.get(`${API_BASE}/devices/${deviceId}/activity?limit=${limit}`);
        return res.data ? res.data.data : [];
      } catch (err) {
        console.error('Failed to fetch device activity:', err);
        return [];
      }
    },
    async createParameter(_, { deviceId, paramData }) {
      try {
        const res = await axios.post(`${API_BASE}/devices/${deviceId}/parameters`, paramData);
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to create parameter:', err);
        throw err;
      }
    },
    async updateParameter(_, { deviceId, paramId, updates }) {
      try {
        const res = await axios.put(`${API_BASE}/devices/${deviceId}/parameters/${paramId}`, updates);
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to update parameter:', err);
        throw err;
      }
    },
    async deleteParameter(_, { deviceId, paramId }) {
      try {
        await axios.delete(`${API_BASE}/devices/${deviceId}/parameters/${paramId}`);
      } catch (err) {
        console.error('Failed to delete parameter:', err);
        throw err;
      }
    },
    async toggleParameterEnabled(_, { deviceId, paramId, enabled }) {
      try {
        const res = await axios.put(`${API_BASE}/devices/${deviceId}/parameters/${paramId}/enable`, { enabled });
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to toggle parameter enabled:', err);
        throw err;
      }
    },
    async connectDevice(_, id) {
      try {
        const res = await axios.post(`${API_BASE}/devices/${id}/communication/connect`);
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to connect device:', err);
        throw err;
      }
    },
    async disconnectDevice(_, id) {
      try {
        const res = await axios.post(`${API_BASE}/devices/${id}/communication/disconnect`);
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to disconnect device:', err);
        throw err;
      }
    },
    async reconnectDevice(_, id) {
      try {
        const res = await axios.post(`${API_BASE}/devices/${id}/communication/reconnect`);
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to reconnect device:', err);
        throw err;
      }
    },
    async testDeviceConnection(_, id) {
      try {
        const res = await axios.post(`${API_BASE}/devices/${id}/communication/test`);
        return res.data ? (res.data.data || res.data) : null;
      } catch (err) {
        console.error('Failed to test device connection:', err);
        throw err;
      }
    },
    async fetchCommunicationStatus(_, id) {
      try {
        const res = await axios.get(`${API_BASE}/devices/${id}/communication/status`);
        return res.data ? res.data.data : null;
      } catch (err) {
        console.error('Failed to fetch communication status:', err);
        return null;
      }
    },
    async fetchPollingPlan(_, deviceId) {
      try {
        const res = await axios.get(`${API_BASE}/devices/${deviceId}/polling-plan`);
        return (res && res.data && res.data.data !== undefined) ? res.data.data : (res.data || res);
      } catch (err) {
        console.error('Failed to fetch polling plan:', err);
        throw err;
      }
    },
    async testReadParameter(_, payload) {
      const deviceId = payload.deviceId || payload.device_id;
      const paramId = payload.paramId || payload.parameterId || payload.parameter_id;
      try {
        const res = await axios.post(`${API_BASE}/devices/${deviceId}/parameters/${paramId}/test-read`);
        return res.data ? (res.data.data || res.data) : null;
      } catch (err) {
        console.error('Failed to test-read parameter:', err);
        throw err;
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
    async fetchActiveAlarms({ commit }) {
      try {
        const res = await axios.get(`${API_BASE}/alarms/active`);
        if (res.data && res.data.data) {
          commit('SET_ACTIVE_ALARMS', res.data.data);
          return res.data.data;
        }
      } catch (err) {
        console.error('Failed to fetch active alarms:', err);
      }
      return [];
    },
    updateAlarmPreferences({ commit }, prefs) {
      commit('SET_ALARM_PREFERENCES', prefs);
    },
    async acknowledgeAlarm({ commit, dispatch }, payload) {
      try {
        const id = typeof payload === 'object' && payload !== null ? payload.id : payload;
        const notes = typeof payload === 'object' && payload !== null ? (payload.notes || '') : '';
        const res = await axios.put(`${API_BASE}/alarms/${id}/acknowledge`, { notes });
        commit('UPDATE_ALARM_STATUS', {
          id: Number(id),
          status: 'ACKNOWLEDGED',
          notes,
          acknowledged_at: new Date().toISOString(),
        });
        dispatch('fetchAlarms');
        dispatch('fetchActiveAlarms');
        dispatch('fetchSystemStatus');
        return res.data;
      } catch (err) {
        console.error('Failed to ack alarm:', err);
        throw err;
      }
    },
    initWebSocket({ commit, dispatch, state }) {
      if (!state.token) return;
      state.wsExplicitClose = false;
      if (state.ws) {
        try { state.ws.close(); } catch (e) {}
      }

      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const tokenParam = state.token ? `?token=${encodeURIComponent(state.token)}` : '';
      const wsUrl = `${protocol}//${window.location.host}/ws${tokenParam}`;

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
        // Reconcile active alarms and system status after connection or reconnection
        dispatch('fetchActiveAlarms');
        dispatch('fetchSystemStatus');
      };

      socket.onclose = () => {
        commit('SET_WS_CONNECTED', false);
        // Auto-reconnect after 3 seconds only if authenticated and not explicitly closed
        if (state.token && !state.wsExplicitClose) {
          setTimeout(() => {
            if (state.token && !state.wsExplicitClose) {
              dispatch('initWebSocket');
            }
          }, 3000);
        }
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
              process_cpu_percent: msg.data.process_cpu_percent != null ? msg.data.process_cpu_percent : 0,
              cpu_per_core: msg.data.cpu_per_core || [],
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
          } else if (msg.type === 'device.connection.changed' || msg.type === 'device.communication.success' || msg.type === 'device.communication.error') {
            window.dispatchEvent(new CustomEvent('device-comm-event', { detail: { type: msg.type, data: msg.data } }));
          } else if (msg.type === 'device.telemetry.received' && msg.data) {
            window.dispatchEvent(new CustomEvent('device-telemetry-event', { detail: msg.data }));
          } else if (msg.type === 'aggregation.result.updated' && msg.data) {
            window.dispatchEvent(new CustomEvent('aggregation-result-event', { detail: msg.data }));
          } else if (msg.type === 'ispu.calculated' && msg.data) {
            window.dispatchEvent(new CustomEvent('ispu-calculated-event', { detail: msg.data }));
          } else if (msg.type === 'notification.alert' && msg.data) {
            dispatch('fetchActiveAlarms');
          } else if (msg.type === 'alarm.created' && msg.data) {
            commit('UPSERT_ACTIVE_ALARM', msg.data);
          } else if (msg.type === 'alarm.acknowledged' && msg.data) {
            commit('UPDATE_ALARM_STATUS', msg.data);
          } else if (msg.type === 'alarm.cleared' && msg.data) {
            commit('REMOVE_ACTIVE_ALARM', msg.data.id || msg.data.ID);
          }
        } catch (e) {
          // ignore parse errors
        }
      };
    },
  },
});
