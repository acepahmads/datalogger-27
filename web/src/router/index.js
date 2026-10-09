import Vue from 'vue';
import VueRouter from 'vue-router';

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

Vue.use(VueRouter);

const routes = [
  { path: '/', redirect: '/dashboard' },
  { path: '/dashboard', name: 'Dashboard', component: DashboardView },
  { path: '/development', name: 'Development', component: DevelopmentView },
  { path: '/development/phases', name: 'Phases', component: DevelopmentView },
  { path: '/development/tasks', name: 'Tasks', component: TasksView },
  { path: '/development/activity', name: 'Activity', component: ActivityView },
  { path: '/development/release', name: 'Release', component: ReleaseDocView },
  { path: '/monitoring/system', name: 'SystemMonitoring', component: SystemMonitoringView },
  { path: '/monitoring/telemetry', name: 'LiveTelemetry', component: TelemetryMonitorView },
  { path: '/monitoring/analysis', name: 'TelemetryAnalysis', component: TelemetryAnalysisView },
  { path: '/monitoring/devices', name: 'Devices', component: DevicesView },
  { path: '/monitoring/devices/:id', name: 'DeviceDetail', component: DeviceDetailView },
  { path: '/monitoring/parameters', name: 'Parameters', component: DevicesView },
  { path: '/alarm/active', name: 'ActiveAlarms', component: AlarmsView },
  { path: '/alarm/history', name: 'AlarmHistory', component: AlarmsView },
  { path: '/system/logs', name: 'SystemLogs', component: LogsView },
  { path: '/system/audit', name: 'AuditTrails', component: LogsView },
  { path: '/administration/config', name: 'Configuration', component: ConfigView },
  { path: '/administration/backup', name: 'BackupRestore', component: BackupRestoreView },
  { path: '/administration/retention', name: 'RetentionStorage', component: RetentionStorageView },
  { path: '*', redirect: '/dashboard' },
];

const router = new VueRouter({
  mode: 'hash',
  base: '/',
  routes,
});

export default router;
