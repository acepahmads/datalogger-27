<template>
  <div class="space-y-6">
    <!-- Top Action / Breadcrumb Bar -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-slate-900 dark:text-white font-sans">
          {{ $t('dashboard.title') }}
        </h1>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5 font-sans">
          {{ $t('dashboard.subtitle') }}
        </p>
      </div>

      <div class="flex items-center space-x-2">
        <button @click="showNewTaskModal = true"
                class="px-3 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
          <svg class="w-3.5 h-3.5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          {{ $t('dashboard.addTask') }}
        </button>
        <button @click="refreshAll" :disabled="refreshing"
                class="p-2 bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700/60 text-slate-700 dark:text-slate-300 rounded-lg text-xs transition-colors"
                :title="$t('dashboard.refresh')">
          <svg class="w-4 h-4" :class="{ 'animate-spin': refreshing }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <polyline points="23 4 23 10 17 10"></polyline>
            <polyline points="1 20 1 14 7 14"></polyline>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
          </svg>
        </button>
      </div>
    </div>

    <!-- 4 Primary KPI Cards -->
    <KpiCards />

    <!-- Architecture Pipeline Status -->
    <div class="saas-card p-4 space-y-3">
      <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800/60 pb-2.5">
        <div class="flex items-center space-x-2">
          <span class="w-2 h-2 rounded-full bg-blue-500"></span>
          <h2 class="text-xs font-bold text-slate-800 dark:text-white uppercase tracking-wider font-sans">
            {{ $t('dashboard.architecturePipeline') }}
          </h2>
        </div>
        <span class="text-3xs font-mono text-emerald-600 dark:text-emerald-400 flex items-center">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5 animate-pulse"></span>
          {{ $t('dashboard.fullPipelineAutonomous') }}
        </span>
      </div>

      <ArchitecturePipeline />
    </div>

    <!-- Main Content: Dual-Column (Development Progress Left, Runtime Edge Monitoring Right) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-5">
      <!-- Left 7 cols: Development Snapshot & Quick Tasks -->
      <div class="lg:col-span-7 space-y-5">
        <!-- Development Summary Card -->
        <div class="saas-card p-5 space-y-4">
          <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800/60 pb-3">
            <div>
              <h2 class="text-sm font-bold text-slate-800 dark:text-white font-sans tracking-tight">{{ $t('dashboard.activeMilestone') }}</h2>
              <span class="text-2xs text-blue-600 dark:text-blue-400 font-sans font-medium">{{ progress.current_phase }}</span>
            </div>
            <router-link to="/development" class="text-2xs text-slate-500 dark:text-slate-400 hover:text-blue-600 dark:hover:text-white font-sans flex items-center">
              {{ $t('dashboard.viewAllPhases') }} &rarr;
            </router-link>
          </div>

          <!-- Progress Bar & Details -->
          <div class="space-y-2">
            <div class="flex justify-between text-xs font-sans">
              <span class="text-slate-600 dark:text-slate-300 font-medium">{{ $t('dashboard.phaseProgress') }}</span>
              <span class="text-slate-900 dark:text-white font-bold">{{ progress.overall_percentage.toFixed(1) }}%</span>
            </div>
            <div class="w-full bg-slate-200 dark:bg-slate-800/80 h-2 rounded-full overflow-hidden">
              <div class="bg-gradient-to-r from-blue-600 to-cyan-500 h-full rounded-full transition-all duration-500"
                   :style="{ width: progress.overall_percentage + '%' }"></div>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3 text-xs pt-1">
            <div class="p-3 rounded-lg bg-slate-50 dark:bg-[#0B0F19]/60 border border-slate-200 dark:border-slate-800/80">
              <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-sans block mb-1">{{ $t('dashboard.activeSubPhase') }}</span>
              <span class="text-xs font-semibold text-slate-800 dark:text-slate-100 font-sans truncate block">{{ progress.current_subphase }}</span>
            </div>
            <div class="p-3 rounded-lg bg-slate-50 dark:bg-[#0B0F19]/60 border border-slate-200 dark:border-slate-800/80">
              <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-sans block mb-1">{{ $t('dashboard.activeTask') }}</span>
              <span class="text-xs font-semibold text-emerald-600 dark:text-emerald-400 font-sans truncate block">{{ progress.current_task }}</span>
            </div>
          </div>
        </div>

        <!-- Recent Activities Feed -->
        <div class="saas-card p-5 space-y-3">
          <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800/60 pb-3">
            <h2 class="text-sm font-bold text-slate-800 dark:text-white font-sans tracking-tight">{{ $t('dashboard.recentActivity') }}</h2>
            <router-link to="/development/activity" class="text-2xs text-blue-600 dark:text-blue-400 hover:text-blue-500 dark:hover:text-blue-300 font-sans font-medium">
              {{ $t('dashboard.viewAllActivity') }} &rarr;
            </router-link>
          </div>

          <div class="space-y-2 max-h-60 overflow-y-auto pr-1">
            <div v-for="act in recentActivities" :key="act.id"
                 class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0B0F19]/60 border border-slate-200 dark:border-slate-800/80 flex items-start justify-between text-xs">
              <div>
                <div class="flex items-center space-x-2">
                  <span class="text-2xs font-semibold text-blue-600 dark:text-blue-400 font-sans">{{ act.phase }}</span>
                  <span class="text-slate-400 dark:text-slate-600">·</span>
                  <span class="font-medium text-slate-800 dark:text-slate-200 font-sans">{{ act.task }}</span>
                </div>
                <p class="text-2xs text-slate-500 dark:text-slate-400 font-sans mt-0.5">{{ act.log }}</p>
              </div>
              <span class="text-3xs text-slate-400 dark:text-slate-500 font-mono flex-shrink-0 ml-2">{{ act.timestamp | formatDate }}</span>
            </div>
            <div v-if="!recentActivities || recentActivities.length === 0" class="text-center py-6 text-2xs text-slate-400">
              {{ $t('dashboard.noActivity') }}
            </div>
          </div>
        </div>
      </div>

      <!-- Right 5 cols: Edge Runtime & MariaDB Monitoring -->
      <div class="lg:col-span-5 space-y-5">
        <!-- Hardware Diagnostics Card -->
        <div class="saas-card p-5 space-y-4">
          <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800/60 pb-3">
            <div class="flex items-center space-x-2">
              <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
              <h2 class="text-sm font-bold text-slate-800 dark:text-white font-sans tracking-tight">{{ $t('dashboard.edgeRuntimeTelemetry') }}</h2>
            </div>
            <span class="text-3xs font-mono text-slate-500 dark:text-slate-400">127.0.0.1:8080</span>
          </div>

          <!-- Gauges -->
          <div class="space-y-3 font-sans text-xs">
            <!-- CPU Breakdown: Datalogger App vs Total Pi Host -->
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <div class="flex items-center space-x-1.5">
                  <span class="text-slate-600 dark:text-slate-400 font-medium">{{ $t('dashboard.cpuLoad') }}</span>
                  <!-- App Process CPU Badge -->
                  <span class="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/20"
                        title="Datalogger Daemon actual CPU footprint">
                    App: {{ (systemStatus.process_cpu_percent || 0).toFixed(1) }}%
                  </span>
                </div>
                <div class="text-right">
                  <span class="font-mono text-slate-700 dark:text-slate-200 font-medium text-xs">Host: {{ systemStatus.cpu_percent.toFixed(1) }}%</span>
                  <span class="text-3xs text-slate-400 dark:text-slate-500 ml-1">({{ systemStatus.num_cpu || 4 }} Cores)</span>
                </div>
              </div>
              <div class="w-full bg-slate-200 dark:bg-slate-800/80 h-1.5 rounded-full overflow-hidden">
                <div class="bg-blue-500 h-full rounded-full transition-[width] duration-300"
                     :style="{ width: Math.min(systemStatus.cpu_percent, 100) + '%' }"></div>
              </div>

              <!-- Per-core breakdown matching htop -->
              <div v-if="systemStatus.cpu_per_core && systemStatus.cpu_per_core.length" class="grid grid-cols-4 gap-1.5 mt-2 pt-1 border-t border-slate-200 dark:border-slate-800/40">
                <div v-for="(c, idx) in systemStatus.cpu_per_core" :key="idx" class="text-2xs">
                  <div class="flex justify-between font-mono text-[10px] text-slate-500 dark:text-slate-400 mb-0.5">
                    <span>Core {{ idx }}</span>
                    <span :class="c > 75 ? 'text-rose-600 dark:text-rose-400' : c > 50 ? 'text-amber-600 dark:text-amber-400' : 'text-slate-700 dark:text-slate-300'">{{ c.toFixed(0) }}%</span>
                  </div>
                  <div class="w-full bg-slate-200 dark:bg-slate-800/80 h-1 rounded-full overflow-hidden">
                    <div class="bg-sky-500 h-full rounded-full transition-[width] duration-300" :style="{ width: Math.min(c, 100) + '%' }"></div>
                  </div>
                </div>
              </div>
            </div>

            <!-- RAM -->
            <div>
              <div class="flex justify-between mb-1">
                <span class="text-slate-600 dark:text-slate-400">{{ $t('dashboard.memory') }}</span>
                <span class="font-mono text-slate-700 dark:text-slate-200 font-medium">
                  {{ systemStatus.ram_total_mb >= 1024 ? (systemStatus.ram_used_mb / 1024).toFixed(2) + ' GB' : systemStatus.ram_used_mb.toFixed(0) + ' MB' }} / {{ systemStatus.ram_total_mb >= 1024 ? (systemStatus.ram_total_mb / 1024).toFixed(2) + ' GB' : systemStatus.ram_total_mb.toFixed(0) + ' MB' }}
                </span>
              </div>
              <div class="w-full bg-slate-200 dark:bg-slate-800/80 h-1.5 rounded-full overflow-hidden">
                <div class="bg-emerald-500 h-full rounded-full transition-all duration-300"
                     :style="{ width: Math.min(systemStatus.ram_percent, 100) + '%' }"></div>
              </div>
            </div>

            <!-- Disk -->
            <div>
              <div class="flex justify-between mb-1">
                <span class="text-slate-600 dark:text-slate-400">{{ $t('dashboard.disk') }}</span>
                <span class="font-mono text-slate-700 dark:text-slate-200 font-medium">{{ systemStatus.disk_used_gb.toFixed(1) }} GB / {{ systemStatus.disk_total_gb.toFixed(1) }} GB</span>
              </div>
              <div class="w-full bg-slate-200 dark:bg-slate-800/80 h-1.5 rounded-full overflow-hidden">
                <div class="bg-indigo-500 h-full rounded-full transition-all duration-300"
                     :style="{ width: Math.min(systemStatus.disk_percent, 100) + '%' }"></div>
              </div>
            </div>
          </div>

          <!-- MariaDB Runtime Status Card -->
          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19]/80 border border-slate-200 dark:border-slate-800/90 space-y-2">
            <div class="flex items-center justify-between">
              <span class="text-2xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 font-sans">{{ $t('navigation.edgeDatabase') }}</span>
              <span class="flex items-center text-emerald-600 dark:text-emerald-400 text-3xs font-semibold font-mono tracking-wide">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1.5"></span>
                MariaDB Edge · {{ $t('status.healthy') }}
              </span>
            </div>
            <div class="grid grid-cols-2 gap-2 text-2xs font-sans pt-1">
              <div>
                <span class="text-slate-400 dark:text-slate-500 block">{{ $t('dashboard.hostPort') }}:</span>
                <span class="font-mono text-slate-700 dark:text-slate-200">127.0.0.1:3306</span>
              </div>
              <div>
                <span class="text-slate-400 dark:text-slate-500 block">{{ $t('dashboard.database') }}:</span>
                <span class="font-mono text-slate-700 dark:text-slate-200">datalogger</span>
              </div>
              <div>
                <span class="text-slate-400 dark:text-slate-500 block">{{ $t('dashboard.version') }}:</span>
                <span class="font-mono text-blue-600 dark:text-blue-400">10.4.32-MariaDB</span>
              </div>
              <div>
                <span class="text-slate-400 dark:text-slate-500 block">{{ $t('dashboard.latency') }}:</span>
                <span class="font-mono text-emerald-600 dark:text-emerald-400">&lt; 1 ms</span>
              </div>
            </div>
          </div>

          <!-- Subsystem Badges -->
          <div class="grid grid-cols-2 gap-2 text-xs pt-1">
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0B0F19]/60 border border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
              <span class="text-2xs text-slate-600 dark:text-slate-400">WebSocket</span>
              <span class="text-3xs font-semibold text-emerald-600 dark:text-emerald-400 font-mono">{{ $t('status.connected') }}</span>
            </div>
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0B0F19]/60 border border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
              <span class="text-2xs text-slate-600 dark:text-slate-400">{{ $t('navigation.devicesSensors') }}</span>
              <span class="text-3xs font-semibold text-blue-600 dark:text-blue-400 font-mono">{{ systemStatus.online_devices }} {{ $t('status.online') }}</span>
            </div>
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0B0F19]/60 border border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
              <span class="text-2xs text-slate-600 dark:text-slate-400">{{ $t('dashboard.throughput') }}</span>
              <span class="text-3xs font-semibold text-emerald-600 dark:text-emerald-400 font-mono">24.5 /s</span>
            </div>
            <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0B0F19]/60 border border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
              <span class="text-2xs text-slate-600 dark:text-slate-400">{{ $t('navigation.activeAlarms') }}</span>
              <span class="text-3xs font-semibold text-slate-500 dark:text-slate-300 font-mono">{{ systemStatus.active_alarms }} {{ $t('status.active') }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <NewTaskModal :show="showNewTaskModal" @close="showNewTaskModal = false" @created="refreshAll" />
  </div>
</template>

<script>
import KpiCards from '../components/KpiCards.vue';
import ArchitecturePipeline from '../components/ArchitecturePipeline.vue';
import NewTaskModal from '../components/NewTaskModal.vue';

export default {
  name: 'DashboardView',
  components: {
    KpiCards,
    ArchitecturePipeline,
    NewTaskModal,
  },
  data() {
    return {
      refreshing: false,
      showNewTaskModal: false,
    };
  },
  computed: {
    progress() {
      return this.$store.state.progress;
    },
    systemStatus() {
      return this.$store.state.systemStatus;
    },
    recentActivities() {
      return this.$store.state.activities.slice(0, 5);
    },
  },
  methods: {
    async refreshAll() {
      this.refreshing = true;
      try {
        await Promise.all([
          this.$store.dispatch('fetchSystemStatus'),
          this.$store.dispatch('fetchProgress'),
          this.$store.dispatch('fetchTasks'),
          this.$store.dispatch('fetchActivity'),
        ]);
      } finally {
        setTimeout(() => {
          this.refreshing = false;
        }, 300);
      }
    },
  },
};
</script>
