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
import OperatorCommandCenterView from '../views/OperatorCommandCenterView.vue';
import ISPUAnalysisView from '../views/ISPUAnalysisView.vue';
import AccessDeniedView from '../views/AccessDeniedView.vue';

Vue.use(VueRouter);

const routes = [
  { path: '/', redirect: '/dashboard' },
  { path: '/operator', redirect: '/monitoring/operator' },
  { path: '/monitoring/operator', name: 'OperatorCommandCenter', component: OperatorCommandCenterView },
  { path: '/dashboard', name: 'Dashboard', component: DashboardView },
  { path: '/monitoring/ispu', name: 'ISPUAnalysis', component: ISPUAnalysisView },
  { path: '/ispu', redirect: '/monitoring/ispu' },

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
  { path: '/monitoring/telemetry', name: 'LiveTelemetry', component: TelemetryMonitorView },
  { path: '/monitoring/analysis', name: 'TelemetryAnalysis', component: TelemetryAnalysisView },
  { path: '/monitoring/devices', name: 'Devices', component: DevicesView },
  { path: '/monitoring/devices/:id', name: 'DeviceDetail', component: DeviceDetailView },
  { path: '/devices/:id', redirect: to => `/monitoring/devices/${to.params.id}` },
  { path: '/monitoring/parameters', name: 'Parameters', component: DevicesView },
  { path: '/monitoring/reports', name: 'Reports', component: ReportsView },

  // Alarms
  { path: '/alarm', redirect: '/alarm/active' },
  { path: '/alarms', redirect: '/alarm/active' },
  { path: '/alarm/active', name: 'ActiveAlarms', component: AlarmsView },
  { path: '/alarm/history', name: 'AlarmHistory', component: AlarmsView },
  { path: '/alarm/queue', name: 'DeliveryQueue', component: AlarmsView },
  { path: '/alarm/delivery', redirect: '/alarm/queue' },
  { path: '/monitoring/delivery', redirect: '/alarm/queue' },
  { path: '/alarm/channels', redirect: '/alarm/active?tab=channels' },

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

  // Access Denied / 403 route
  { path: '/access-denied', name: 'AccessDenied', component: AccessDeniedView },

  { path: '*', redirect: '/dashboard' },
];

const router = new VueRouter({
  mode: 'hash',
  base: '/',
  routes,
});

// Centralized Frontend Route Guard for Role-Based Access Control
router.beforeEach((to, from, next) => {
  const user = store.state.user || {};
  const rawRole = (user.role || '').toUpperCase();
  const isSuper = rawRole === 'SUPERADMIN' || rawRole === 'ADMINISTRATOR';

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
