<template>
  <div class="grid grid-cols-2 lg:grid-cols-4 gap-3.5">
    <!-- 1. Overall Progress -->
    <div class="saas-kpi-card group">
      <div class="flex items-center justify-between">
        <span class="kpi-label">Overall Progress</span>
        <div class="w-7 h-7 rounded-lg bg-blue-500/10 flex items-center justify-center text-blue-400">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"></path>
          </svg>
        </div>
      </div>
      <div class="mt-2.5 flex items-baseline space-x-1.5">
        <span class="text-2xl font-bold font-sans tracking-tight text-white">
          {{ progress.overall_percentage ? progress.overall_percentage.toFixed(1) : '20.5' }}%
        </span>
      </div>
      <div class="mt-1 text-2xs text-slate-400">
        Phase 1 Foundation Complete
      </div>
    </div>

    <!-- 2. Completed Tasks -->
    <div class="saas-kpi-card group">
      <div class="flex items-center justify-between">
        <span class="kpi-label">Completed Tasks</span>
        <div class="w-7 h-7 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-400">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
        </div>
      </div>
      <div class="mt-2.5 flex items-baseline space-x-1.5">
        <span class="text-2xl font-bold font-sans tracking-tight text-white">
          {{ progress.completed_tasks_count || 14 }}
        </span>
        <span class="text-xs text-slate-500 font-sans">
          / {{ progress.total_tasks_count || 117 }} tasks
        </span>
      </div>
      <div class="mt-1 text-2xs text-emerald-400/90 font-medium">
        Verified & tested
      </div>
    </div>

    <!-- 3. Current Phase -->
    <div class="saas-kpi-card group">
      <div class="flex items-center justify-between">
        <span class="kpi-label">Current Phase</span>
        <div class="w-7 h-7 rounded-lg bg-indigo-500/10 flex items-center justify-center text-indigo-400">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
          </svg>
        </div>
      </div>
      <div class="mt-2.5 flex items-baseline space-x-1.5">
        <span class="text-2xl font-bold font-sans tracking-tight text-white">
          Phase {{ progress.current_phase_number || 1 }}
        </span>
      </div>
      <div class="mt-1 text-2xs text-slate-400 truncate">
        {{ progress.current_subphase || 'Core Edge Architecture' }}
      </div>
    </div>

    <!-- 4. Current Task -->
    <div class="saas-kpi-card group">
      <div class="flex items-center justify-between">
        <span class="kpi-label">Current Task</span>
        <div class="w-7 h-7 rounded-lg bg-amber-500/10 flex items-center justify-center text-amber-400">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
        </div>
      </div>
      <div class="mt-2.5 truncate">
        <span class="text-lg font-bold font-sans tracking-tight text-slate-100 truncate block">
          {{ progress.current_task || 'UI Redesign & MariaDB' }}
        </span>
      </div>
      <div class="mt-1 text-2xs text-blue-400/90 truncate">
        In Active Development
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'KpiCards',
  computed: {
    progress() {
      return this.$store.state.progress;
    },
    systemStatus() {
      return this.$store.state.systemStatus;
    },
  },
};
</script>

<style scoped>
.saas-kpi-card {
  @apply bg-[#111827]/80 backdrop-blur-sm border border-slate-800/80 rounded-xl p-4 transition-all duration-200 hover:border-slate-700/80 hover:shadow-soft;
}
.kpi-label {
  @apply text-2xs font-medium uppercase tracking-wider text-slate-400 font-sans;
}
</style>
