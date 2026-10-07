<template>
  <div class="space-y-5 font-sans">
    <!-- Top Filter Bar -->
    <div class="saas-card p-4 flex flex-col md:flex-row md:items-center justify-between gap-3">
      <!-- Search Input -->
      <div class="flex-1 max-w-md relative">
        <div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-slate-500">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
        </div>
        <input type="text" v-model="searchQuery" placeholder="Search tasks, deliverables, or owners..."
               class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg pl-9 pr-3 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-sans transition-colors" />
      </div>

      <!-- Filters & Action -->
      <div class="flex flex-wrap items-center gap-2.5 text-xs font-sans">
        <!-- Phase Filter -->
        <select v-model="filterPhase" class="bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-1.5 text-slate-200 focus:outline-none focus:border-blue-500 font-sans text-xs">
          <option value="">All Phases (1-10)</option>
          <option v-for="p in phases" :key="p.id" :value="p.id">Phase {{ p.phase_number }}: {{ p.name.replace(/^Phase \d+ — /, '') }}</option>
        </select>

        <!-- Status Filter -->
        <select v-model="filterStatus" class="bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-1.5 text-slate-200 focus:outline-none focus:border-blue-500 font-sans text-xs">
          <option value="">All Statuses</option>
          <option value="DONE">DONE</option>
          <option value="WORKING">WORKING</option>
          <option value="TESTING">TESTING</option>
          <option value="PENDING">PENDING</option>
          <option value="PLANNED">PLANNED</option>
          <option value="SUPERSEDED">SUPERSEDED</option>
          <option value="BLOCKED">BLOCKED</option>
          <option value="FAILED">FAILED</option>
          <option value="WAITING_APPROVAL">WAITING_APPROVAL</option>
        </select>

        <!-- New Task Button -->
        <button @click="showNewTaskModal = true"
                class="px-3.5 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
          + New Task
        </button>
      </div>
    </div>

    <!-- Task List Table in SaaS Card -->
    <div class="saas-card overflow-hidden">
      <div class="px-5 py-3.5 border-b border-slate-800/80 flex items-center justify-between">
        <div>
          <h2 class="text-sm font-bold text-white font-sans tracking-tight">
            Task Directory
          </h2>
          <span class="text-2xs text-slate-400 font-sans">Showing {{ filteredTasks.length }} of {{ tasks.length }} tasks across 10 phases</span>
        </div>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs border-collapse font-sans">
          <thead>
            <tr class="border-b border-slate-800/80 bg-[#0B0F19]/60 text-3xs font-semibold text-slate-400 uppercase tracking-wider font-sans">
              <th class="py-3 px-4 w-14 text-center">ID</th>
              <th class="py-3 px-4 w-28">Phase</th>
              <th class="py-3 px-4">Task Name & Scope</th>
              <th class="py-3 px-4 w-28">Status</th>
              <th class="py-3 px-4 w-32">Progress</th>
              <th class="py-3 px-4 w-24">Priority</th>
              <th class="py-3 px-4 w-32">Owner</th>
              <th class="py-3 px-4 w-40">Test Verification</th>
              <th class="py-3 px-4 w-16 text-center">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/50 text-slate-300">
            <tr v-for="task in filteredTasks" :key="task.id"
                class="hover:bg-slate-800/25 transition-colors cursor-pointer"
                @click="selectedTask = task">
              <td class="py-3 px-4 text-center font-mono text-3xs text-slate-500">#{{ task.id }}</td>
              <td class="py-3 px-4 font-mono text-2xs text-blue-400 font-medium">Phase {{ task.phase_id }}</td>
              <td class="py-3 px-4">
                <div class="font-semibold text-slate-100 font-sans text-xs">{{ task.task_name }}</div>
                <div class="text-2xs text-slate-400 line-clamp-1 mt-0.5 font-sans">{{ task.description }}</div>
              </td>
              <td class="py-3 px-4">
                <span class="px-2 py-0.5 rounded-md text-3xs font-semibold font-mono tracking-wide inline-block"
                      :class="getStatusPillClass(task.status)">
                  {{ task.status }}
                </span>
              </td>
              <td class="py-3 px-4">
                <div class="flex items-center space-x-2">
                  <div class="flex-1 bg-slate-800/80 h-1.5 rounded-full overflow-hidden">
                    <div class="h-full rounded-full transition-all duration-300"
                         :class="task.progress >= 100 ? 'bg-emerald-500' : 'bg-blue-500'"
                         :style="{ width: task.progress + '%' }"></div>
                  </div>
                  <span class="text-2xs font-semibold w-8 text-right font-sans"
                        :class="task.progress >= 100 ? 'text-emerald-400' : 'text-slate-300'">
                    {{ task.progress }}%
                  </span>
                </div>
              </td>
              <td class="py-3 px-4">
                <span class="text-2xs font-sans font-medium"
                      :class="task.priority === 'CRITICAL' ? 'text-rose-400 font-bold' : task.priority === 'HIGH' ? 'text-amber-400' : 'text-slate-400'">
                  {{ task.priority }}
                </span>
              </td>
              <td class="py-3 px-4 font-sans text-2xs text-slate-300 truncate max-w-[120px]">
                {{ task.owner || 'Engineer' }}
              </td>
              <td class="py-3 px-4 font-sans text-2xs text-slate-400 truncate max-w-[160px]">
                {{ task.test_result || 'Pending test' }}
              </td>
              <td class="py-3 px-4 text-center" @click.stop>
                <button @click="selectedTask = task"
                        class="p-1.5 rounded-md bg-slate-800/60 hover:bg-slate-700 text-blue-400 hover:text-white transition-colors"
                        title="Edit Task">
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                    <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                  </svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modals -->
    <TaskModal :task="selectedTask" @close="selectedTask = null" @saved="refresh" />
    <NewTaskModal :show="showNewTaskModal" @close="showNewTaskModal = false" @created="refresh" />
  </div>
</template>

<script>
import TaskModal from '../components/TaskModal.vue';
import NewTaskModal from '../components/NewTaskModal.vue';

export default {
  name: 'TasksView',
  components: {
    TaskModal,
    NewTaskModal,
  },
  data() {
    return {
      searchQuery: '',
      filterPhase: '',
      filterStatus: '',
      selectedTask: null,
      showNewTaskModal: false,
    };
  },
  computed: {
    tasks() {
      return this.$store.state.tasks;
    },
    phases() {
      return this.$store.state.progress.phases || [];
    },
    filteredTasks() {
      return this.tasks.filter((t) => {
        if (this.filterPhase && t.phase_id !== Number(this.filterPhase)) {
          return false;
        }
        if (this.filterStatus && t.status !== this.filterStatus) {
          return false;
        }
        if (this.searchQuery) {
          const q = this.searchQuery.toLowerCase();
          const matchName = t.task_name && t.task_name.toLowerCase().includes(q);
          const matchDesc = t.description && t.description.toLowerCase().includes(q);
          const matchOwner = t.owner && t.owner.toLowerCase().includes(q);
          const matchResult = t.test_result && t.test_result.toLowerCase().includes(q);
          if (!matchName && !matchDesc && !matchOwner && !matchResult) {
            return false;
          }
        }
        return true;
      });
    },
  },
  mounted() {
    this.refresh();
  },
  methods: {
    refresh() {
      this.$store.dispatch('fetchTasks');
      this.$store.dispatch('fetchProgress');
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
          return 'bg-slate-800/40 text-slate-400 border border-slate-700/30';
      }
    },
  },
};
</script>
