import Vue from 'vue';
import VueRouter from 'vue-router';
import store from '../store';

import DashboardView from '../views/DashboardView.vue';
import DevelopmentView from '../views/DevelopmentView.vue';
import TasksView from '../views/TasksView.vue';
import ActivityView from '../views/ActivityView.vue';
import SystemMonitoringView from '../views/SystemMonitoringView.vue';
import DevicesView from '../views/DevicesView.vue';
import DeviceDetailView from '../views/DeviceDetailView.vue';
import SensorCalibrationView from '../views/SensorCalibrationView.vue';
import TelemetryMonitorView from '../views/TelemetryMonitorView.vue';
import TelemetryAnalysisView from '../views/TelemetryAnalysisView.vue';
import AlarmsView from '../views/AlarmsView.vue';
import LogsView from '../views/LogsView.vue';
import ReleaseDocView from '../views/ReleaseDocView.vue';
import ConfigView from '../views/ConfigView.vue';
import BackupRestoreView from '../views/BackupRestoreView.vue';
import RetentionStorageView from '../views/RetentionStorageView.vue';
import ReportsView from '../views/ReportsView.vue';
import AccessControlView from '../views/AccessControlView.vue';
import TenantsManagementView from '../views/TenantsManagementView.vue';
import OperatorCommandCenterView from '../views/OperatorCommandCenterView.vue';
import DashboardOperatorView from '../views/DashboardOperatorView.vue';
import ISPUAnalysisView from '../views/ISPUAnalysisView.vue';
import WQMSAnalysisView from '../views/WQMSAnalysisView.vue';
import IntegrationView from '../views/IntegrationView.vue';
import AccessDeniedView from '../views/AccessDeniedView.vue';
import LoginView from '../views/LoginView.vue';

Vue.use(VueRouter);

const routes = [
  { path: '/login', name: 'Login', component: LoginView },
  { path: '/', redirect: '/monitoring/operator' },
  { path: '/operator', redirect: '/monitoring/operator' },
  { path: '/monitoring/operator', name: 'OperatorCommandCenter', component: OperatorCommandCenterView },
  { path: '/monitoring/dashboard-operator', name: 'DashboardOperator', component: DashboardOperatorView },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: DashboardView,
    meta: { roles: ['SUPERADMIN'] },
  },
  { path: '/quality', redirect: '/dashboard' },
  { path: '/monitoring/quality', redirect: '/dashboard' },
  { path: '/monitoring/ispu', name: 'ISPUAnalysis', component: ISPUAnalysisView },
  { path: '/ispu', redirect: '/monitoring/ispu' },
  { path: '/monitoring/wqms', name: 'WQMSAnalysis', component: WQMSAnalysisView },
  { path: '/wqms', redirect: '/monitoring/wqms' },

  // Development Modules — STRICTLY SUPERADMIN ONLY
  {
    path: '/development',
    name: 'Development',
    component: DevelopmentView,
    meta: { roles: ['SUPERADMIN'] },
  },
  {
    path: '/development/phases',
    name: 'Phases',
    component: DevelopmentView,
    meta: { roles: ['SUPERADMIN'] },
  },
  {
    path: '/development/tasks',
    name: 'Tasks',
    component: TasksView,
    meta: { roles: ['SUPERADMIN'] },
  },
  {
    path: '/development/activity',
    name: 'Activity',
    component: ActivityView,
    meta: { roles: ['SUPERADMIN'] },
  },
  {
    path: '/development/release',
    name: 'Release',
    component: ReleaseDocView,
    meta: { roles: ['SUPERADMIN'] },
  },

  // Operational & Monitoring Modules
  { path: '/monitoring/system', name: 'SystemMonitoring', component: SystemMonitoringView },
  {
    path: '/monitoring/telemetry',
    name: 'LiveTelemetry',
    component: TelemetryMonitorView,
    meta: { roles: ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN', 'ENGINEER'] },
  },
  { path: '/monitoring/analysis', name: 'TelemetryAnalysis', component: TelemetryAnalysisView },
  {
    path: '/monitoring/telemetry-injection',
    name: 'TelemetryInjection',
    component: () => import('../views/TelemetryInjectionView.vue'),
    meta: { roles: ['SUPERADMIN'] },
  },
  { path: '/monitoring/devices', name: 'Devices', component: DevicesView },
  { path: '/monitoring/devices/:id', name: 'DeviceDetail', component: DeviceDetailView },
  { path: '/devices/:id', redirect: to => `/monitoring/devices/${to.params.id}` },
  {
    path: '/monitoring/calibration',
    name: 'SensorCalibration',
    component: SensorCalibrationView,
    meta: { disallowOperator: true },
  },
  { path: '/calibration', redirect: '/monitoring/calibration' },
  {
    path: '/monitoring/reports',
    name: 'Reports',
    component: ReportsView,
    meta: { roles: ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN', 'ENGINEER'] },
  },

  // Alarms
  { path: '/alarm', redirect: '/alarm/active' },
  { path: '/alarms', redirect: '/alarm/active' },
  { path: '/alarm/active', name: 'ActiveAlarms', component: AlarmsView },
  { path: '/alarm/history', name: 'AlarmHistory', component: AlarmsView },
  { path: '/alarm/queue', name: 'DeliveryQueue', component: AlarmsView },
  { path: '/alarm/delivery', redirect: '/alarm/queue' },
  { path: '/monitoring/delivery', redirect: '/alarm/queue' },
  {
    path: '/alarm/live-comm',
    name: 'LiveComm',
    component: () => import('../views/LiveCommView.vue'),
    meta: { roles: ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN', 'ENGINEER'] },
  },
  {
    path: '/alarm/activity-monitor',
    name: 'SystemActivityMonitor',
    component: () => import('../views/SystemActivityMonitorView.vue'),
    meta: { roles: ['SUPERADMIN'] },
  },
  { path: '/system/activity-monitor', redirect: '/alarm/activity-monitor' },

  // System Logs & Audit
  { path: '/system/logs', name: 'SystemLogs', component: LogsView },
  { path: '/system/audit', name: 'AuditTrails', component: LogsView },
  { path: '/system/communication-logs', name: 'CommunicationLogs', component: LogsView },

  // Administration Modules
  {
    path: '/administration/config',
    name: 'Configuration',
    component: ConfigView,
    meta: { roles: ['SUPERADMIN', 'ADMIN', 'ENGINEER'] },
  },
  {
    path: '/administration/backup',
    name: 'BackupRestore',
    component: BackupRestoreView,
    meta: { roles: ['SUPERADMIN', 'ADMIN', 'ENGINEER'] },
  },
  {
    path: '/administration/retention',
    name: 'RetentionStorage',
    component: RetentionStorageView,
    meta: { roles: ['SUPERADMIN', 'ADMIN', 'ENGINEER'] },
  },
  {
    path: '/administration/tenants',
    name: 'TenantsManagement',
    component: TenantsManagementView,
    meta: { roles: ['SUPERADMIN', 'ADMIN'] },
  },
  { path: '/system/tenants', redirect: '/administration/tenants' },
  {
    path: '/administration/users',
    name: 'UserManagement',
    component: AccessControlView,
    meta: { roles: ['SUPERADMIN', 'ADMIN'] },
  },
  {
    path: '/administration/roles',
    name: 'RoleManagement',
    component: AccessControlView,
    meta: { roles: ['SUPERADMIN'] },
  },
  { path: '/administration/access-control', redirect: '/administration/users' },

  // Phase 8 — Outbound & Integration
  {
    path: '/integration',
    name: 'Integration',
    component: IntegrationView,
    meta: { roles: ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN', 'ENGINEER'] },
  },
  {
    path: '/integration/queue',
    name: 'OutboundDeliveryQueue',
    component: IntegrationView,
    meta: { roles: ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN', 'ENGINEER', 'OPERATOR'] },
  },
  { path: '/integration/destinations', redirect: '/integration' },
  { path: '/destinations', redirect: '/integration' },

  // Access Denied / 403 route
  { path: '/access-denied', name: 'AccessDenied', component: AccessDeniedView },

  { path: '*', redirect: '/monitoring/operator' },
];

const router = new VueRouter({
  mode: 'hash',
  base: '/',
  routes,
});

// Centralized Frontend Route Guard for Authentication and Role-Based Access Control
router.beforeEach(async (to, from, next) => {
  let token = store.state.token;
  let isAuth = !!token && !!store.state.user;
  const isLoginPage = to.path === '/login';

  // Automatically authenticate as default operator if unauthenticated, so entry to dashboard is immediate
  if (!isAuth && !isLoginPage) {
    try {
      await store.dispatch('autoLoginOperator');
      token = store.state.token;
      isAuth = !!token && !!store.state.user;
    } catch (e) {
      console.warn('Auto-login as default operator failed, continuing with fallback:', e);
    }
  }

  // 1. If still unauthenticated and navigating to protected route -> redirect to /login
  if (!isAuth && !isLoginPage) {
    return next({
      path: '/login',
      query: { redirect: to.fullPath },
    });
  }

  const user = store.state.user || {};
  const rawRole = (user.role || '').toUpperCase();
  const isSuper = rawRole === 'SUPERADMIN' || rawRole === 'ADMINISTRATOR';
  const isOperator = rawRole === 'OPERATOR';

  // 2. If authenticated and navigating to /login (without explicit ?switch=true) -> redirect to role default landing
  if (isAuth && isLoginPage && !to.query.switch) {
    return next({
      path: isSuper ? '/dashboard' : (isOperator ? '/monitoring/dashboard-operator' : '/monitoring/operator'),
      replace: true,
    });
  }

  // 2b. Role OPERATOR default landing & Operator Console hiding -> redirect straight to Dashboard Operator
  if (isAuth && isOperator && (to.path === '/' || to.path === '/operator' || to.path === '/monitoring/operator')) {
    return next({ path: '/monitoring/dashboard-operator', replace: true });
  }

  // 2c. Role OPERATOR retains full access to Outbound Delivery Queue via both /integration/queue and /integration?tab=queue
  if (isAuth && isOperator && to.path === '/integration') {
    if (to.query && to.query.tab === 'queue') {
      return next();
    }
    return next({ path: '/integration/queue', query: { ...to.query, tab: 'queue' }, replace: true });
  }

  // 3. RBAC validation
  if (isAuth && !isSuper) {
    if (!store.state.systemStatus || !store.state.systemStatus._cloudChecked) {
      try {
        await store.dispatch('fetchSystemStatus');
        if (store.state.systemStatus) {
          store.state.systemStatus._cloudChecked = true;
        }
      } catch (_) { }
    }
    const isCloudMode = Boolean(
      (store.state.systemStatus && store.state.systemStatus.cloud_mode) ||
      String((store.state.systemStatus && store.state.systemStatus.monitoring_domain) || '').toUpperCase() === 'CLOUD'
    );
    if (isCloudMode) {
      const cloudRestrictedPrefixes = [
        '/monitoring/devices',
        '/devices',
        '/monitoring/parameters',
        '/monitoring/calibration',
        '/calibration',
        '/monitoring/telemetry-injection',
      ];
      if (cloudRestrictedPrefixes.some(p => to.path === p || to.path.startsWith(p + '/'))) {
        return next({
          path: '/access-denied',
          query: { from: to.fullPath },
        });
      }
    }
  }

  if (to.meta && to.meta.disallowOperator && isOperator) {
    return next({
      path: '/access-denied',
      query: { from: to.fullPath },
    });
  }

  if (to.meta && Array.isArray(to.meta.roles) && to.meta.roles.length > 0) {
    const allowed = to.meta.roles.some(r => {
      const targetRole = r.toUpperCase();
      if ((targetRole === 'SUPERADMIN' || targetRole === 'ADMINISTRATOR') && isSuper) {
        return true;
      }
      return targetRole === rawRole;
    });

    if (!allowed) {
      return next({
        path: '/access-denied',
        query: { from: to.fullPath },
      });
    }
  }

  next();
});

export default router;
