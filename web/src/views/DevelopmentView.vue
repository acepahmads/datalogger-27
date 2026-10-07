<template>
  <div class="space-y-6">
    <!-- Header: Development Lifecycle -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans">
          Development Lifecycle
        </h1>
        <p class="text-xs text-slate-400 mt-1 font-sans">
          Single source of truth for complete hardware and software delivery.
        </p>
      </div>

      <div class="flex items-center space-x-2.5">
        <button @click="showNewTaskModal = true"
                class="px-3.5 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
          <svg class="w-3.5 h-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          Add Task
        </button>
        <button @click="refreshAll" :disabled="refreshing"
                class="p-2 bg-slate-800/60 hover:bg-slate-800 border border-slate-700/60 text-slate-300 rounded-lg text-xs transition-colors"
                title="Refresh">
          <svg class="w-4 h-4" :class="{ 'animate-spin': refreshing }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <polyline points="23 4 23 10 17 10"></polyline>
            <polyline points="1 20 1 14 7 14"></polyline>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
          </svg>
        </button>
      </div>
    </div>

    <!-- KPI Cards (Overall Progress, Completed, Current Phase, Current Task) -->
    <KpiCards />

    <!-- Current Status Hero Card (Visually dominant Current Phase, no rigid 4 boxes) -->
    <div class="saas-card p-5 relative overflow-hidden bg-gradient-to-r from-[#111827] via-[#131E35] to-[#111827]">
      <div class="absolute right-0 top-0 bottom-0 w-1/3 bg-radial from-blue-600/10 to-transparent pointer-events-none"></div>

      <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-center">
        <!-- Visually Dominant Left Section: Active Phase -->
        <div class="lg:col-span-6 space-y-2">
          <div class="flex items-center space-x-2">
            <span class="text-3xs font-bold uppercase tracking-widest text-blue-400 font-mono">ACTIVE MILESTONE</span>
            <span class="w-1.5 h-1.5 rounded-full bg-blue-500 animate-pulse"></span>
          </div>
          <div class="text-2xl font-extrabold text-white tracking-tight font-sans">
            {{ activePhaseName }}
          </div>
          <p class="text-xs text-slate-300 leading-relaxed max-w-xl font-sans">
            {{ activePhaseDescription }}
          </p>
        </div>

        <!-- Right Section: Current Task, Next Action & Timestamp in clean hierarchy -->
        <div class="lg:col-span-6 grid grid-cols-1 sm:grid-cols-2 gap-4 border-t lg:border-t-0 lg:border-l border-slate-800/80 pt-4 lg:pt-0 lg:pl-6 text-xs font-sans">
          <!-- Current Task -->
          <div class="space-y-1">
            <span class="text-3xs font-medium uppercase tracking-wider text-slate-400 block">CURRENT TASK</span>
            <div class="text-sm font-semibold text-slate-100 flex items-center space-x-1.5">
              <span class="w-2 h-2 rounded-full bg-amber-400 flex-shrink-0"></span>
              <span class="truncate">{{ progress.current_task || 'UI Redesign & MariaDB' }}</span>
            </div>
            <span class="text-2xs text-blue-400 block">{{ progress.current_subphase || 'Edge Runtime' }}</span>
          </div>

          <!-- Next Action -->
          <div class="space-y-1">
            <span class="text-3xs font-medium uppercase tracking-wider text-slate-400 block">NEXT ACTION</span>
            <div class="text-xs text-slate-200 leading-snug">
              {{ progress.next_action || 'Complete Phase 1 deployment and verify MariaDB connectivity' }}
            </div>
          </div>

          <!-- Last Update -->
          <div class="sm:col-span-2 pt-2 border-t border-slate-800/50 flex items-center justify-between text-2xs text-slate-400 font-sans">
            <div class="flex items-center space-x-2">
              <span>Last update:</span>
              <span class="font-mono text-slate-300">{{ progress.last_update }}</span>
            </div>
            <span class="text-emerald-400 font-medium">MariaDB Edge Synchronized</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 10 Phases Development Roadmap Cards (Expandable Accordion) -->
    <div class="space-y-3">
      <div class="flex items-center justify-between">
        <h2 class="text-sm font-bold text-slate-200 tracking-tight font-sans">
          Engineering Roadmap (10 Phases)
        </h2>
        <span class="text-2xs text-slate-400">Click to expand tasks & deliverables</span>
      </div>

      <!-- Phase Cards List -->
      <div class="space-y-3">
        <div v-for="phase in progress.phases" :key="phase.id"
             class="saas-card overflow-hidden transition-all duration-200"
             :class="{ 'ring-1 ring-blue-500/30': openPhases[phase.id] }">
          
          <!-- Phase Card Header (Clickable) -->
          <div @click="togglePhase(phase.id)"
               class="p-4 cursor-pointer hover:bg-slate-800/20 transition-colors flex flex-col md:flex-row md:items-center justify-between gap-3 select-none">
            
            <div class="flex items-start space-x-3.5 flex-1">
              <!-- Phase Number Pill -->
              <span class="px-2.5 py-1 rounded-md text-2xs font-mono font-bold tracking-wider"
                    :class="getPhaseBadgeClass(phase)">
                PHASE 0{{ phase.phase_number }}
              </span>

              <!-- Name & Description -->
              <div class="space-y-0.5">
                <div class="text-sm font-bold text-white tracking-tight font-sans">
                  {{ phase.name.replace(/^Phase \d+ — /, '') }}
                </div>
                <p class="text-2xs text-slate-400 font-sans line-clamp-1">
                  {{ phase.description }}
                </p>
              </div>
            </div>

            <!-- Right: Task Completion, Progress Bar, Status Badge & Chevron -->
            <div class="flex items-center space-x-4 self-end md:self-center font-sans">
              <!-- Task Count -->
              <span class="text-2xs text-slate-400 font-sans hidden sm:inline-block">
                {{ getCompletedTasksCount(phase) }} / {{ getActiveTasksCount(phase) }} active tasks
              </span>

              <!-- Thin subtle Progress Bar -->
              <div class="w-24 sm:w-32 bg-slate-800/80 h-1.5 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all duration-500"
                     :class="getProgressBarColor(phase)"
                     :style="{ width: phase.progress + '%' }"></div>
              </div>

              <!-- Percentage -->
              <span class="text-2xs font-bold w-10 text-right font-sans"
                    :class="phase.progress >= 100 ? 'text-emerald-400' : phase.progress > 0 ? 'text-blue-400' : 'text-slate-500'">
                {{ phase.progress.toFixed(0) }}%
              </span>

              <!-- Status Pill -->
              <span class="px-2 py-0.5 rounded-md text-3xs font-semibold tracking-wider font-mono uppercase"
                    :class="getStatusPillClass(phase.status)">
                {{ phase.status === 'COMPLETED' ? 'DONE' : phase.status }}
              </span>

              <!-- Chevron -->
              <svg class="w-4 h-4 text-slate-400 transform transition-transform duration-200"
                   :class="{ 'rotate-180': openPhases[phase.id] }"
                   fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <polyline points="6 9 12 15 18 9"></polyline>
              </svg>
            </div>
          </div>

          <!-- Phase Details Accordion Content -->
          <div v-if="openPhases[phase.id]" class="px-4 pb-4 pt-3 border-t border-slate-800/60 bg-[#0E1524]/40 space-y-4">
            
            <!-- Subphase Breakdown (Hierarchy for phases with subphases like Phase 2) -->
            <div v-if="phase.subphases && phase.subphases.length" class="space-y-2.5">
              <div class="flex items-center justify-between">
                <span class="text-2xs font-semibold text-slate-300 uppercase tracking-wider font-sans flex items-center space-x-1.5">
                  <span class="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
                  <span>Subphase Verification & Delivery Breakdown</span>
                </span>
                <span class="text-3xs text-emerald-400 font-mono">Normalized Hierarchy</span>
              </div>

              <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-2.5">
                <div v-for="sp in phase.subphases" :key="sp.id"
                     @click="toggleSubphaseFilter(sp.id)"
                     class="p-3 rounded-lg bg-[#0B0F19]/80 border transition-all cursor-pointer"
                     :class="selectedSubphaseId === sp.id ? 'border-blue-500 ring-1 ring-blue-500/40 bg-blue-950/20' : 'border-slate-800/80 hover:border-slate-700'">
                  <div class="flex items-center justify-between mb-1.5">
                    <span class="text-3xs font-mono font-bold px-1.5 py-0.5 rounded"
                          :class="getSubphaseBadgeClass(sp)">
                      {{ sp.status || (sp.progress >= 100 ? 'DONE' : 'WORKING') }}
                    </span>
                    <span class="text-2xs font-bold font-mono"
                          :class="sp.progress >= 100 ? 'text-emerald-400' : sp.progress > 0 ? 'text-blue-400' : 'text-slate-400'">
                      {{ sp.progress.toFixed(0) }}%
                    </span>
                  </div>
                  <div class="text-xs font-bold text-white font-sans truncate">{{ sp.name }}</div>
                  <div class="text-2xs text-slate-400 font-sans mt-0.5 line-clamp-1">{{ sp.description }}</div>
                  <div class="mt-2 pt-1.5 border-t border-slate-800/60 flex items-center justify-between text-3xs font-mono">
                    <span class="text-slate-400">{{ sp.acceptance_criteria || 'VERIFIED' }}</span>
                    <span class="text-blue-400 hover:underline">
                      {{ getSubphaseTasks(phase, sp.id).length }} tasks
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Task Header & Filter Bar -->
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pt-2 border-t border-slate-800/50">
              <div class="flex items-center space-x-2">
                <span class="text-2xs font-semibold text-slate-400 uppercase tracking-wider font-sans">
                  Tasks & Deliverables ({{ getDisplayedTasks(phase).length }})
                </span>
                <span v-if="selectedSubphaseId" class="text-3xs text-blue-400 font-sans bg-blue-500/10 px-2 py-0.5 rounded border border-blue-500/20 flex items-center">
                  Subphase Filter Active
                  <button @click.stop="selectedSubphaseId = null" class="ml-1.5 text-slate-400 hover:text-white" title="Clear filter">&times;</button>
                </span>
              </div>
              <div class="flex items-center space-x-2">
                <button v-if="selectedSubphaseId" @click.stop="selectedSubphaseId = null"
                        class="text-3xs text-slate-400 hover:text-white font-sans">
                  Show All Tasks
                </button>
                <button @click.stop="openNewTaskForPhase(phase.id)"
                        class="text-2xs text-blue-400 hover:text-blue-300 font-medium font-sans flex items-center">
                  + Add Task to Phase {{ phase.phase_number }}
                </button>
              </div>
            </div>

            <!-- Tasks Grid in Modern SaaS Cards -->
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2.5">
              <div v-for="task in getDisplayedTasks(phase)" :key="task.id"
                   @click="selectedTask = task"
                   class="p-3 rounded-lg bg-[#111827] border border-slate-800/80 hover:border-blue-500/50 hover:shadow-soft cursor-pointer transition-all duration-150 flex flex-col justify-between"
                   :class="{'opacity-60 bg-slate-900/50': task.status === 'SUPERSEDED'}">
                <div>
                  <div class="flex items-center justify-between mb-1.5">
                    <span class="text-3xs font-mono text-slate-500">#{{ task.id }}</span>
                    <span class="px-1.5 py-0.5 rounded text-3xs font-semibold font-mono tracking-wide"
                          :class="getStatusPillClass(task.status)">
                      {{ task.status }}
                    </span>
                  </div>
                  <div class="text-xs font-semibold text-slate-100 font-sans">{{ task.task_name }}</div>
                  <div class="text-2xs text-slate-400 mt-1 font-sans line-clamp-2">{{ task.description }}</div>
                </div>

                <div class="mt-3 pt-2 border-t border-slate-800/60 flex items-center justify-between text-2xs">
                  <span class="text-slate-400 font-sans truncate max-w-[120px]">{{ task.owner || 'Engineer' }}</span>
                  <span class="font-sans font-medium" :class="task.progress >= 100 ? 'text-emerald-400' : task.progress > 0 ? 'text-blue-400' : 'text-slate-500'">
                    {{ task.progress }}%
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Development Activity Timeline Component -->
    <div class="saas-card p-5 space-y-4">
      <div class="flex items-center justify-between border-b border-slate-800/60 pb-3">
        <div>
          <h2 class="text-sm font-bold text-white font-sans tracking-tight">Milestone Timeline</h2>
          <p class="text-2xs text-slate-400 font-sans mt-0.5">Chronological record of verified deliverables</p>
        </div>
        <router-link to="/development/activity" class="text-2xs text-blue-400 hover:text-blue-300 font-sans font-medium">
          View all activity &rarr;
        </router-link>
      </div>

      <div class="space-y-4 relative pl-5 border-l border-slate-800">
        <!-- Milestone 1 -->
        <div class="relative">
          <div class="absolute -left-[25px] top-1 w-3 h-3 rounded-full bg-emerald-500 ring-4 ring-[#0B0F19]"></div>
          <div class="text-xs font-semibold text-white font-sans">Phase 1 Foundation Completed</div>
          <div class="text-2xs text-slate-400 font-sans mt-0.5">Go backend, Vue 2 UI, MariaDB schema, and cross-platform architecture deployed.</div>
          <div class="text-3xs text-slate-500 font-mono mt-1">2026-10-05 · Verified (17/17 PASS)</div>
        </div>

        <!-- Milestone 2 -->
        <div class="relative">
          <div class="absolute -left-[25px] top-1 w-3 h-3 rounded-full bg-emerald-500 ring-4 ring-[#0B0F19]"></div>
          <div class="text-xs font-semibold text-white font-sans">Phase 2 Device & Communication Completed</div>
          <div class="text-2xs text-slate-400 font-sans mt-0.5">Phase 2.1 Device Management, Phase 2.2 Modbus RTU/TCP Engine, and Phase 2.3 Hardening verified.</div>
          <div class="text-3xs text-emerald-400 font-mono mt-1">2026-10-07 · Verified (81/81 PASS)</div>
        </div>

        <!-- Milestone 3 -->
        <div class="relative">
          <div class="absolute -left-[25px] top-1 w-3 h-3 rounded-full bg-blue-500 ring-4 ring-[#0B0F19]"></div>
          <div class="text-xs font-semibold text-white font-sans">Phase 3 Data Engine (Next Milestone)</div>
          <div class="text-2xs text-slate-400 font-sans mt-0.5">Raw data acquisition, scaling, calibration formulas, and moving average aggregations.</div>
          <div class="text-3xs text-blue-400 font-mono mt-1">Planned Roadmap</div>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <TaskModal :task="selectedTask" @close="selectedTask = null" @saved="refreshAll" />
    <NewTaskModal :show="showNewTaskModal" :default-phase-id="newTaskPhaseId" @close="showNewTaskModal = false" @created="refreshAll" />
  </div>
</template>

<script>
import KpiCards from '../components/KpiCards.vue';
import TaskModal from '../components/TaskModal.vue';
import NewTaskModal from '../components/NewTaskModal.vue';

export default {
  name: 'DevelopmentView',
  components: {
    KpiCards,
    TaskModal,
    NewTaskModal,
  },
  data() {
    return {
      refreshing: false,
      selectedTask: null,
      showNewTaskModal: false,
      newTaskPhaseId: 1,
      selectedSubphaseId: null,
      openPhases: { 1: true, 2: true }, // Open Phase 1 & 2 by default
    };
  },
  computed: {
    progress() {
      return this.$store.state.progress;
    },
    activePhaseName() {
      if (this.progress.current_phase) {
        return this.progress.current_phase;
      }
      return 'Phase 2 — Device & Communication';
    },
    activePhaseDescription() {
      return 'Device registration, connection profiles, Modbus RTU & TCP communication engines, concurrency isolation, and production hardening.';
    },
  },
  methods: {
    togglePhase(id) {
      this.$set(this.openPhases, id, !this.openPhases[id]);
      this.selectedSubphaseId = null;
    },
    toggleSubphaseFilter(spId) {
      if (this.selectedSubphaseId === spId) {
        this.selectedSubphaseId = null;
      } else {
        this.selectedSubphaseId = spId;
      }
    },
    getSubphaseTasks(phase, subphaseId) {
      if (!phase || !phase.tasks) return [];
      return phase.tasks.filter((t) => t.subphase_id === subphaseId);
    },
    getDisplayedTasks(phase) {
      if (!phase || !phase.tasks) return [];
      if (this.selectedSubphaseId) {
        return phase.tasks.filter((t) => t.subphase_id === this.selectedSubphaseId);
      }
      return phase.tasks;
    },
    openNewTaskForPhase(phaseId) {
      this.newTaskPhaseId = phaseId;
      this.showNewTaskModal = true;
    },
    getCompletedTasksCount(phase) {
      if (!phase || !phase.tasks) return 0;
      return phase.tasks.filter((t) => t.status === 'DONE').length;
    },
    getActiveTasksCount(phase) {
      if (!phase || !phase.tasks) return 0;
      return phase.tasks.filter((t) => t.status !== 'SUPERSEDED' && t.status !== 'PLANNED').length;
    },
    async refreshAll() {
      this.refreshing = true;
      try {
        await Promise.all([
          this.$store.dispatch('fetchProgress'),
          this.$store.dispatch('fetchTasks'),
        ]);
      } finally {
        setTimeout(() => {
          this.refreshing = false;
        }, 300);
      }
    },
    getPhaseBadgeClass(phase) {
      if (phase.progress >= 100 || phase.status === 'COMPLETED') {
        return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
      }
      if (phase.progress > 0 || phase.status === 'WORKING') {
        return 'bg-blue-500/10 text-blue-400 border border-blue-500/20';
      }
      return 'bg-slate-800/40 text-slate-400 border border-slate-700/30';
    },
    getSubphaseBadgeClass(sp) {
      if (sp.status === 'DONE' || sp.progress >= 100) {
        return 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30';
      }
      if (sp.status === 'PLANNED') {
        return 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/30';
      }
      if (sp.progress > 0 || sp.status === 'WORKING') {
        return 'bg-blue-500/20 text-blue-400 border border-blue-500/30';
      }
      return 'bg-slate-800 text-slate-400 border border-slate-700';
    },
    getProgressBarColor(phase) {
      if (phase.progress >= 100) return 'bg-emerald-500';
      if (phase.progress > 0) return 'bg-blue-500';
      return 'bg-slate-700';
    },
    getStatusPillClass(status) {
      switch (status) {
        case 'DONE':
        case 'COMPLETED':
          return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
        case 'WORKING':
          return 'bg-blue-500/10 text-blue-400 border border-blue-500/20';
        case 'TESTING':
          return 'bg-purple-500/10 text-purple-400 border border-purple-500/20';
        case 'BLOCKED':
        case 'FAILED':
          return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
        case 'WAITING_APPROVAL':
          return 'bg-amber-500/10 text-amber-400 border border-amber-500/20';
        case 'PLANNED':
          return 'bg-cyan-500/10 text-cyan-400 border border-cyan-500/20';
        case 'SUPERSEDED':
          return 'bg-slate-700/30 text-slate-400 border border-slate-600/40';
        case 'PENDING':
        default:
          return 'bg-slate-800/50 text-slate-400 border border-slate-700/40';
      }
    },
  },
};
</script>
