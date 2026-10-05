<template>
  <div v-if="task" class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 select-none">
    <div class="bg-[#111827] border border-slate-700/80 rounded-2xl shadow-2xl max-w-2xl w-full max-h-[90vh] flex flex-col text-slate-200 text-xs font-sans">
      <!-- Modal Header -->
      <div class="px-5 py-4 border-b border-slate-800/80 flex items-center justify-between bg-[#0B0F19]/60 rounded-t-2xl">
        <div class="flex items-center space-x-2.5">
          <span class="font-mono text-3xs px-2.5 py-0.5 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/20 font-medium">
            Task #{{ task.id }}
          </span>
          <span class="text-3xs text-slate-400 font-sans font-medium">Phase {{ task.phase_id }}</span>
          <h2 class="text-sm font-bold text-white truncate max-w-md font-sans">{{ task.task_name }}</h2>
        </div>
        <button @click="$emit('close')" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-5 overflow-y-auto space-y-4 flex-1">
        <!-- Description -->
        <div>
          <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Task Description</label>
          <div class="p-3 rounded-xl bg-[#0B0F19] border border-slate-800/80 text-slate-300 font-sans leading-relaxed">
            {{ task.description || 'No detailed description provided.' }}
          </div>
        </div>

        <!-- Status & Progress Inputs -->
        <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
          <div>
            <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Status</label>
            <select v-model="editForm.status" class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans">
              <option value="PENDING">PENDING</option>
              <option value="WORKING">WORKING</option>
              <option value="TESTING">TESTING</option>
              <option value="DONE">DONE</option>
              <option value="BLOCKED">BLOCKED</option>
              <option value="FAILED">FAILED</option>
              <option value="WAITING_APPROVAL">WAITING_APPROVAL</option>
            </select>
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Progress ({{ editForm.progress }}%)</label>
            <input type="range" min="0" max="100" v-model.number="editForm.progress" class="w-full mt-2.5" />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Priority</label>
            <select v-model="editForm.priority" class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans">
              <option value="LOW">LOW</option>
              <option value="MEDIUM">MEDIUM</option>
              <option value="HIGH">HIGH</option>
              <option value="CRITICAL">CRITICAL</option>
            </select>
          </div>
        </div>

        <!-- Owner & Test Result -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div>
            <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Owner</label>
            <input type="text" v-model="editForm.owner" class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans" />
          </div>
          <div>
            <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Test Verification Result</label>
            <input type="text" v-model="editForm.test_result" placeholder="e.g. MariaDB migration verified, passed" class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans" />
          </div>
        </div>

        <!-- Notes -->
        <div>
          <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Engineering Notes</label>
          <textarea v-model="editForm.notes" rows="2" placeholder="Implementation details, architecture decisions..." class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans"></textarea>
        </div>

        <!-- Task Logs History -->
        <div v-if="task.logs && task.logs.length > 0">
          <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Activity Log History ({{ task.logs.length }})</label>
          <div class="space-y-1.5 max-h-36 overflow-y-auto pr-1">
            <div v-for="log in task.logs" :key="log.id" class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800/80 text-2xs flex items-start justify-between font-sans">
              <div>
                <span class="text-blue-400 font-semibold">[{{ log.user || 'system' }}]</span>
                <span class="text-slate-200 ml-1 font-medium">{{ log.action }}</span>
                <span class="text-slate-400 ml-2">{{ log.log }}</span>
              </div>
              <span class="text-3xs font-mono text-slate-500 flex-shrink-0 ml-2">{{ log.timestamp | formatDate }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-5 py-4 border-t border-slate-800/80 bg-[#0B0F19]/60 rounded-b-2xl flex items-center justify-between">
        <button @click="deleteTask" class="px-3 py-1.5 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 border border-rose-500/20 text-rose-400 font-sans text-xs transition-colors">
          Delete Task
        </button>
        <div class="flex items-center space-x-2.5">
          <button @click="$emit('close')" class="px-3.5 py-1.5 rounded-lg bg-slate-800/80 hover:bg-slate-700 border border-slate-700 text-slate-300 text-xs transition-colors">
            Cancel
          </button>
          <button @click="saveTask" :disabled="saving" class="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-semibold text-xs shadow-soft transition-colors flex items-center">
            <span v-if="saving" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
            Save Changes
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TaskModal',
  props: {
    task: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      saving: false,
      editForm: {
        status: 'PENDING',
        progress: 0,
        priority: 'MEDIUM',
        owner: '',
        test_result: '',
        notes: '',
      },
    };
  },
  watch: {
    task: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.editForm = {
            status: newVal.status || 'PENDING',
            progress: newVal.progress || 0,
            priority: newVal.priority || 'MEDIUM',
            owner: newVal.owner || 'Engineer',
            test_result: newVal.test_result || '',
            notes: newVal.notes || '',
          };
        }
      },
    },
  },
  methods: {
    async saveTask() {
      if (!this.task) return;
      this.saving = true;
      try {
        await this.$store.dispatch('updateTask', {
          id: this.task.id,
          updates: this.editForm,
        });
        this.$emit('saved');
        this.$emit('close');
      } catch (err) {
        alert('Failed to save task: ' + (err.response?.data?.error || err.message));
      } finally {
        this.saving = false;
      }
    },
    async deleteTask() {
      if (!confirm(`Are you sure you want to delete task "${this.task.task_name}"?`)) return;
      try {
        await this.$store.dispatch('deleteTask', this.task.id);
        this.$emit('saved');
        this.$emit('close');
      } catch (err) {
        alert('Failed to delete task: ' + (err.response?.data?.error || err.message));
      }
    },
  },
};
</script>
