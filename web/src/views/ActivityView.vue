<template>
  <div class="space-y-5 font-sans">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans">
          {{ $t('activity.title') }}
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          {{ $t('activity.subtitle') }}
        </p>
      </div>

      <button @click="refreshActivity" class="px-3.5 py-1.5 bg-slate-800/80 hover:bg-slate-700 text-slate-200 rounded-lg text-xs font-medium border border-slate-700/60 transition-colors">
        {{ $t('activity.refreshFeed') }}
      </button>
    </div>

    <!-- Timeline Container -->
    <div class="saas-card p-6">
      <div v-if="activities.length === 0" class="py-8 text-center text-slate-500 font-sans text-xs">
        {{ $t('activity.noActivity') }}
      </div>
      <div v-else class="relative pl-6 border-l border-slate-800 space-y-6">
        <div v-for="item in activities" :key="item.id" class="relative group">
          <!-- Soft Timeline Node -->
          <div class="absolute -left-[31px] top-1.5 w-3.5 h-3.5 rounded-full bg-[#111827] border-2 border-blue-500 group-hover:scale-110 transition-transform"></div>

          <!-- Card Content -->
          <div class="p-4 rounded-xl bg-[#0B0F19]/70 border border-slate-800/80 hover:border-slate-700 transition-colors">
            <div class="flex flex-wrap items-center justify-between text-2xs mb-1.5 gap-2">
              <div class="flex items-center space-x-2 font-sans">
                <span class="font-semibold text-blue-400">{{ item.phase }}</span>
                <span class="text-slate-600">·</span>
                <span class="text-slate-400">{{ item.user }}</span>
              </div>
              <div class="flex items-center space-x-2">
                <span class="text-3xs font-mono text-slate-500">{{ item.timestamp | formatDate }}</span>
                <span class="px-2 py-0.5 rounded text-3xs font-mono font-semibold"
                      :class="item.result === 'PASSED' || item.result === 'DONE' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-blue-500/10 text-blue-400 border border-blue-500/20'">
                  {{ item.result || 'OK' }}
                </span>
              </div>
            </div>

            <div class="text-sm font-semibold text-slate-100 font-sans">
              {{ item.task }}
            </div>

            <div class="text-xs text-slate-400 mt-1 font-sans">
              <span class="text-slate-300 font-medium">{{ item.action }}:</span> {{ item.log }}
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ActivityView',
  computed: {
    activities() {
      return this.$store.state.activities;
    },
  },
  mounted() {
    this.refreshActivity();
  },
  methods: {
    refreshActivity() {
      this.$store.dispatch('fetchActivity');
    },
  },
};
</script>
