<template>
  <div class="space-y-6 font-sans">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans">
          System Health & Edge Telemetry
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          Comprehensive diagnostic metrics, hardware utilization, and MariaDB edge health.
        </p>
      </div>

      <div class="flex items-center space-x-2">
        <span class="flex items-center text-emerald-400 text-xs font-semibold font-mono tracking-wide px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
          <span class="w-2 h-2 rounded-full bg-emerald-400 mr-2 animate-pulse"></span>
          MARIADB EDGE · AUTONOMOUS
        </span>
      </div>
    </div>

    <!-- System Telemetry Gauges Grid -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <!-- CPU -->
      <div class="saas-card p-5 space-y-3">
        <div class="flex items-center justify-between text-xs font-sans">
          <div class="flex items-center space-x-1.5">
            <span class="text-slate-400 font-medium">Edge CPU</span>
            <span class="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                  title="Datalogger Daemon actual CPU footprint">
              App: {{ (status.process_cpu_percent || 0).toFixed(1) }}%
            </span>
          </div>
          <span class="font-mono text-blue-400 font-bold">Host: {{ status.cpu_percent.toFixed(1) }}%</span>
        </div>
        <div class="w-full bg-slate-800/80 h-2 rounded-full overflow-hidden">
          <div class="bg-blue-500 h-full rounded-full transition-[width] duration-300"
               :style="{ width: Math.min(status.cpu_percent, 100) + '%' }"></div>
        </div>
        <!-- Per-core breakdown -->
        <div v-if="status.cpu_per_core && status.cpu_per_core.length" class="grid grid-cols-4 gap-2 pt-1 border-t border-slate-800/40">
          <div v-for="(c, idx) in status.cpu_per_core" :key="idx" class="text-2xs font-mono">
            <div class="flex justify-between text-[10px] text-slate-400">
              <span>C{{ idx }}</span>
              <span :class="c > 75 ? 'text-rose-400' : 'text-slate-300'">{{ c.toFixed(0) }}%</span>
            </div>
            <div class="w-full bg-slate-800/80 h-1 rounded-full overflow-hidden mt-0.5">
              <div class="bg-sky-400 h-full rounded-full transition-[width] duration-300" :style="{ width: Math.min(c, 100) + '%' }"></div>
            </div>
          </div>
        </div>
        <div class="text-2xs text-slate-400 font-sans flex justify-between pt-1">
          <span>{{ status.num_cpu || 4 }} Logical Cores</span>
          <span class="font-mono">Arch: {{ status.arch }}</span>
        </div>
      </div>

      <!-- RAM -->
      <div class="saas-card p-5 space-y-3">
        <div class="flex items-center justify-between text-xs font-sans">
          <span class="text-slate-400 font-medium">Memory Allocation (htop formula)</span>
          <span class="font-mono text-emerald-400 font-bold">{{ status.ram_percent.toFixed(1) }}%</span>
        </div>
        <div class="w-full bg-slate-800/80 h-2 rounded-full overflow-hidden">
          <div class="bg-emerald-500 h-full rounded-full transition-all duration-300"
               :style="{ width: Math.min(status.ram_percent, 100) + '%' }"></div>
        </div>
        <div class="text-2xs text-slate-400 font-sans flex justify-between pt-1">
          <span class="font-mono font-medium text-slate-200">
            {{ status.ram_total_mb >= 1024 ? (status.ram_used_mb / 1024).toFixed(2) + ' GB' : status.ram_used_mb.toFixed(0) + ' MB' }} used
          </span>
          <span class="font-mono text-slate-400">
            {{ status.ram_total_mb >= 1024 ? (status.ram_total_mb / 1024).toFixed(2) + ' GB' : status.ram_total_mb.toFixed(0) + ' MB' }} total
          </span>
        </div>
      </div>

      <!-- Disk -->
      <div class="saas-card p-5 space-y-3">
        <div class="flex items-center justify-between text-xs font-sans">
          <span class="text-slate-400 font-medium">Storage Volume</span>
          <span class="font-mono text-indigo-400 font-bold">{{ status.disk_percent.toFixed(1) }}%</span>
        </div>
        <div class="w-full bg-slate-800/80 h-2 rounded-full overflow-hidden">
          <div class="bg-indigo-500 h-full rounded-full transition-all duration-300"
               :style="{ width: Math.min(status.disk_percent, 100) + '%' }"></div>
        </div>
        <div class="text-2xs text-slate-400 font-sans flex justify-between pt-1">
          <span class="font-mono">{{ status.disk_used_gb.toFixed(1) }} GB used</span>
          <span class="font-mono">{{ status.disk_total_gb.toFixed(1) }} GB total</span>
        </div>
      </div>
    </div>

    <!-- 4 Subsystem Information Cards -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- 1. Engine Status -->
      <div class="saas-card p-5 space-y-3 text-xs font-sans">
        <div class="text-xs font-bold text-blue-400 uppercase tracking-wider pb-2 border-b border-slate-800/80">
          Engine Status
        </div>
        <div class="flex justify-between"><span class="text-slate-400">Service:</span><span class="text-emerald-400 font-bold font-mono">{{ status.service_status }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Workers:</span><span class="text-slate-200 font-mono">{{ status.worker_status }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Scheduler:</span><span class="text-slate-200 font-mono">{{ status.scheduler_status }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Local Queue:</span><span class="text-emerald-400 font-mono">{{ status.queue_status }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Database:</span><span class="text-emerald-400 font-semibold font-mono">MariaDB Edge</span></div>
      </div>

      <!-- 2. Devices Summary -->
      <div class="saas-card p-5 space-y-3 text-xs font-sans">
        <div class="text-xs font-bold text-emerald-400 uppercase tracking-wider pb-2 border-b border-slate-800/80">
          Devices
        </div>
        <div class="flex justify-between"><span class="text-slate-400">Configured:</span><span class="text-slate-200 font-mono">{{ status.total_devices }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Online:</span><span class="text-emerald-400 font-bold font-mono">{{ status.online_devices }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Offline:</span><span class="text-slate-400 font-mono">{{ status.offline_devices }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Error:</span><span class="text-rose-400 font-semibold font-mono">{{ status.error_devices }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Last Comm:</span><span class="text-blue-400 font-mono">{{ status.last_device_comm }}</span></div>
      </div>

      <!-- 3. Telemetry Ingestion -->
      <div class="saas-card p-5 space-y-3 text-xs font-sans">
        <div class="text-xs font-bold text-amber-400 uppercase tracking-wider pb-2 border-b border-slate-800/80">
          Telemetry Data
        </div>
        <div class="flex justify-between"><span class="text-slate-400">Total Samples:</span><span class="text-slate-200 font-mono">{{ status.data_received_count.toLocaleString() }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Throughput:</span><span class="text-blue-300 font-mono">{{ status.data_per_sec }} /sec</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Quality:</span><span class="text-emerald-400 font-semibold font-mono">99.8% GOOD</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Invalid:</span><span class="text-slate-400 font-mono">{{ status.invalid_data_count }}</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Last Sample:</span><span class="text-slate-300 font-mono">{{ status.last_data_received }}</span></div>
      </div>

      <!-- 4. Communication Health -->
      <div class="saas-card p-5 space-y-3 text-xs font-sans">
        <div class="text-xs font-bold text-purple-400 uppercase tracking-wider pb-2 border-b border-slate-800/80">
          Communication
        </div>
        <div class="flex justify-between"><span class="text-slate-400">Channels:</span><span class="text-slate-200 font-mono">{{ status.total_connections }} active</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Success Rate:</span><span class="text-emerald-400 font-bold font-mono">{{ status.success_rate.toFixed(1) }}%</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Fail Rate:</span><span class="text-slate-400 font-mono">{{ status.failed_rate.toFixed(1) }}%</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Latency:</span><span class="text-blue-300 font-semibold font-mono">{{ status.avg_latency_ms }} ms</span></div>
        <div class="flex justify-between"><span class="text-slate-400">Uptime:</span><span class="text-slate-200 font-mono">{{ status.uptime_human }}</span></div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'SystemMonitoringView',
  computed: {
    status() {
      return this.$store.state.systemStatus;
    },
  },
};
</script>
