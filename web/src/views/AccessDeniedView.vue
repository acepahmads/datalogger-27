<template>
  <div class="min-h-[70vh] flex items-center justify-center px-4 font-sans select-none">
    <div class="max-w-lg w-full bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-6 sm:p-8 shadow-xl text-center space-y-5">
      <!-- Shield / Lock Graphic -->
      <div class="w-16 h-16 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-500 flex items-center justify-center mx-auto shadow-sm">
        <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
        </svg>
      </div>

      <!-- Title & Code -->
      <div class="space-y-1">
        <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-3xs font-mono font-bold bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/30">
          <span>HTTP 403</span>
          <span>&bull;</span>
          <span>FORBIDDEN</span>
        </div>
        <h1 class="text-xl sm:text-2xl font-bold tracking-tight text-slate-900 dark:text-white">
          {{ $t('accessDenied.title') || 'Access Denied' }}
        </h1>
        <p class="text-xs sm:text-sm text-slate-500 dark:text-slate-400">
          {{ $t('accessDenied.subtitle') || 'You do not have the required permissions to access this resource or feature.' }}
        </p>
      </div>

      <!-- Current Identity & Role Details -->
      <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19]/80 border border-slate-200 dark:border-slate-800/80 text-left space-y-2 text-xs">
        <div class="flex items-center justify-between">
          <span class="text-3xs uppercase tracking-wider font-semibold text-slate-500 dark:text-slate-400">
            {{ $t('accessDenied.currentUser') || 'Current Account' }}
          </span>
          <span class="font-mono font-bold text-slate-800 dark:text-slate-200">
            @{{ currentUser.username || 'unknown' }}
          </span>
        </div>
        <div class="flex items-center justify-between">
          <span class="text-3xs uppercase tracking-wider font-semibold text-slate-500 dark:text-slate-400">
            {{ $t('accessDenied.assignedRole') || 'Assigned Role' }}
          </span>
          <span class="inline-flex items-center px-2 py-0.5 rounded-full text-3xs font-semibold"
                :class="roleBadgeClass">
            {{ currentUser.role || 'GUEST' }}
          </span>
        </div>
        <div v-if="attemptedUrl" class="flex items-center justify-between pt-1 border-t border-slate-200 dark:border-slate-800">
          <span class="text-3xs uppercase tracking-wider font-semibold text-slate-500 dark:text-slate-400">
            {{ $t('accessDenied.requestedPath') || 'Target Path' }}
          </span>
          <span class="font-mono text-3xs text-rose-500 truncate max-w-[200px]" :title="attemptedUrl">
            {{ attemptedUrl }}
          </span>
        </div>
      </div>

      <!-- Policy Explanation -->
      <p class="text-2xs text-slate-500 dark:text-slate-400 leading-relaxed text-center px-2">
        {{ $t('accessDenied.policyNotice') || 'Under the Phase 7 Role-Based Access Control policy, Development tracking and raw telemetry inspection are strictly reserved for SUPERADMIN. Operational and engineering modules are provisioned based on explicit role capability.' }}
      </p>

      <!-- Navigation Actions -->
      <div class="pt-2 flex flex-col sm:flex-row items-center justify-center gap-3">
        <router-link
          to="/monitoring/operator"
          class="w-full sm:w-auto px-4 py-2.5 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-semibold shadow-sm transition-colors flex items-center justify-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5" />
          </svg>
          <span>{{ $t('navigation.operatorConsole') || 'Operator Console' }}</span>
        </router-link>

        <router-link
          v-if="isSuperAdmin"
          to="/dashboard"
          class="w-full sm:w-auto px-4 py-2.5 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 text-xs font-semibold transition-colors flex items-center justify-center gap-2"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
          </svg>
          <span>{{ $t('navigation.dashboard') || 'Overview Dashboard' }}</span>
        </router-link>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'AccessDeniedView',
  computed: {
    currentUser() {
      return this.$store.state.user || {};
    },
    isSuperAdmin() {
      const r = (this.currentUser.role || '').toUpperCase();
      return r === 'SUPERADMIN' || r === 'ADMINISTRATOR';
    },
    attemptedUrl() {
      return this.$route.query.from || '';
    },
    roleBadgeClass() {
      const r = (this.currentUser.role || '').toUpperCase();
      if (r === 'SUPERADMIN' || r === 'ADMINISTRATOR') {
        return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/30';
      }
      if (r === 'ADMIN') {
        return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/30';
      }
      if (r === 'ENGINEER') {
        return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/30';
      }
      if (r === 'OPERATOR') {
        return 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30';
      }
      return 'bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400';
    },
  },
};
</script>
