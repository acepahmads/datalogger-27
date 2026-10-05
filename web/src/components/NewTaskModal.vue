<template>
  <div v-if="show" class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 select-none">
    <div class="bg-[#111827] border border-slate-700/80 rounded-2xl shadow-2xl max-w-xl w-full flex flex-col text-slate-200 text-xs font-sans">
      <div class="px-5 py-4 border-b border-slate-800/80 flex items-center justify-between bg-[#0B0F19]/60 rounded-t-2xl">
        <h2 class="text-sm font-bold text-white font-sans tracking-tight">Create Development Task</h2>
        <button @click="$emit('close')" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition-colors">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>

      <form @submit.prevent="createTask" class="p-5 space-y-4">
        <div>
          <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Development Phase *</label>
          <select v-model.number="form.phase_id" required class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans">
            <option v-for="p in phases" :key="p.id" :value="p.id">
              Phase {{ p.phase_number }}: {{ p.name.replace(/^Phase \d+ — /, '') }}
            </option>
          </select>
        </div>

        <div>
          <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Task Name *</label>
          <input type="text" v-model="form.task_name" required placeholder="e.g. Modbus RTU Frame Check" class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans" />
        </div>

        <div>
          <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Scope & Description</label>
          <textarea v-model="form.description" rows="3" placeholder="Technical deliverables, protocol details, or test plans..." class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans"></textarea>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Initial Status</label>
            <select v-model="form.status" class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans">
              <option value="PENDING">PENDING</option>
              <option value="WORKING">WORKING</option>
              <option value="TESTING">TESTING</option>
              <option value="DONE">DONE</option>
            </select>
          </div>
          <div>
            <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Priority</label>
            <select v-model="form.priority" class="w-full bg-[#0B0F19] border border-slate-700/80 rounded-lg px-3 py-2 text-xs text-white focus:outline-none focus:border-blue-500 font-sans">
              <option value="LOW">LOW</option>
              <option value="MEDIUM">MEDIUM</option>
              <option value="HIGH">HIGH</option>
              <option value="CRITICAL">CRITICAL</option>
            </select>
          </div>
        </div>

        <div class="pt-3 border-t border-slate-800/80 flex items-center justify-end space-x-2.5">
          <button type="button" @click="$emit('close')" class="px-3.5 py-1.5 rounded-lg bg-slate-800/80 hover:bg-slate-700 border border-slate-700 text-slate-300 text-xs transition-colors">
            Cancel
          </button>
          <button type="submit" :disabled="loading" class="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-semibold text-xs shadow-soft transition-colors flex items-center">
            <span v-if="loading" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
            Create Task
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NewTaskModal',
  props: {
    show: {
      type: Boolean,
      default: false,
    },
    defaultPhaseId: {
      type: Number,
      default: 1,
    },
  },
  data() {
    return {
      loading: false,
      form: {
        phase_id: 1,
        task_name: '',
        description: '',
        status: 'PENDING',
        priority: 'MEDIUM',
        progress: 0,
        owner: 'Engineer',
      },
    };
  },
  computed: {
    phases() {
      return this.$store.state.progress.phases || [];
    },
  },
  watch: {
    defaultPhaseId(val) {
      if (val) this.form.phase_id = val;
    },
  },
  methods: {
    async createTask() {
      this.loading = true;
      try {
        await this.$store.dispatch('createTask', this.form);
        this.form.task_name = '';
        this.form.description = '';
        this.$emit('created');
        this.$emit('close');
      } catch (err) {
        alert('Failed to create task: ' + (err.response?.data?.error || err.message));
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
