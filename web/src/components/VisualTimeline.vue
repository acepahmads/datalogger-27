<template>
  <div class="bg-slate-900 border border-slate-800 rounded p-3">
    <div class="flex items-center justify-between border-b border-slate-800 pb-2 mb-2.5">
      <div class="flex items-center space-x-2">
        <span class="w-2 h-2 rounded-sm bg-blue-500"></span>
        <h3 class="text-xs font-bold text-slate-100 uppercase tracking-wider font-mono">
          Development Roadmap & Progress Timeline
        </h3>
      </div>
      <div class="flex items-center space-x-3 text-2xs font-mono">
        <span class="text-slate-400">Total Progress:</span>
        <span class="text-blue-400 font-bold">{{ progress.overall_percentage.toFixed(1) }}%</span>
      </div>
    </div>

    <!-- Timeline Phase Bars -->
    <div class="space-y-2">
      <div v-for="phase in progress.phases" :key="phase.id"
           class="p-2 rounded bg-slate-950/60 border border-slate-800/80 hover:border-slate-700 transition-colors">
        <div class="flex items-center justify-between text-2xs mb-1">
          <div class="flex items-center space-x-2">
            <span class="font-mono font-bold w-14 text-slate-400">Phase {{ phase.phase_number }}</span>
            <span class="font-medium text-slate-200">{{ phase.name.replace(/^Phase \d+ — /, '') }}</span>
          </div>
          <div class="flex items-center space-x-2">
            <span class="font-mono text-2xs text-slate-400">{{ phase.tasks ? phase.tasks.filter(t => t.status === 'DONE').length : 0 }}/{{ phase.tasks ? phase.tasks.length : 0 }} tasks</span>
            <!-- Status Badge -->
            <span v-if="phase.progress >= 100 || phase.status === 'COMPLETED'"
                  class="bg-emerald-950/80 text-emerald-400 border border-emerald-800/60 px-2 py-0.2 rounded font-mono font-semibold text-2xs">
              DONE
            </span>
            <span v-else-if="phase.progress > 0 || phase.status === 'WORKING'"
                  class="bg-blue-950/80 text-blue-400 border border-blue-800/60 px-2 py-0.2 rounded font-mono font-semibold text-2xs">
              {{ phase.progress.toFixed(0) }}%
            </span>
            <span v-else
                  class="bg-slate-900 text-slate-500 border border-slate-800 px-2 py-0.2 rounded font-mono text-2xs">
              0% PENDING
            </span>
          </div>
        </div>

        <!-- Progress Bar Track -->
        <div class="w-full bg-slate-900 h-1.5 rounded-full overflow-hidden flex">
          <div class="h-full transition-all duration-500 rounded-full"
               :class="phase.progress >= 100 ? 'bg-emerald-500' : phase.progress > 0 ? 'bg-blue-500' : 'bg-slate-700'"
               :style="{ width: phase.progress + '%' }"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'VisualTimeline',
  computed: {
    progress() {
      return this.$store.state.progress;
    },
  },
};
</script>
