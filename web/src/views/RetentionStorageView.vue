<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-slate-900 dark:text-white font-sans flex items-center gap-2.5">
          <svg class="w-5 h-5 text-indigo-500 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
          </svg>
          {{ $t('retention.title') }}
        </h1>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5 font-sans">
          {{ $t('retention.subtitle') }}
        </p>
      </div>

      <!-- Action Buttons & Tabs -->
      <div class="flex items-center space-x-2">
        <button @click="triggerSweep"
                :disabled="isSweeping || (storage && storage.active_housekeeping_running)"
                class="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
          <svg class="w-3.5 h-3.5 mr-1.5" :class="{ 'animate-spin': isSweeping }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
          </svg>
          {{ isSweeping ? $t('retention.runningHousekeeping') : $t('retention.runHousekeeping') }}
        </button>

        <div class="flex items-center space-x-1 bg-white dark:bg-[#111827] p-1 rounded-lg border border-slate-200 dark:border-slate-800 text-xs font-sans shadow-sm">
          <button @click="activeTab = 'policies'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'policies' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('retention.tabPolicies') }}
          </button>
          <button @click="activeTab = 'storage'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'storage' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('retention.tabStorage') }}
          </button>
          <button @click="activeTab = 'history'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'history' ? 'bg-indigo-600 text-white font-semibold shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
            {{ $t('retention.tabHistory') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Active Housekeeping Banner -->
    <div v-if="storage && storage.active_housekeeping_running"
         class="p-3.5 rounded-xl bg-gradient-to-r from-indigo-950/70 to-blue-950/60 border border-indigo-500/30 flex items-center justify-between text-xs text-white">
      <div class="flex items-center space-x-3">
        <span class="flex h-2.5 w-2.5 relative">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-indigo-400 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-indigo-500"></span>
        </span>
        <span class="font-semibold tracking-wide font-sans">
          HOUSEKEEPING RETENTION CLEANUP IN PROGRESS...
        </span>
      </div>
      <span class="font-mono text-indigo-300 text-2xs">Non-overlapping lock active</span>
    </div>

    <!-- Storage Capacity Warning Banner (if WARNING or CRITICAL) -->
    <div v-if="storage && (storage.capacity_status === 'WARNING' || storage.capacity_status === 'CRITICAL')"
         class="p-3.5 rounded-xl border flex items-center justify-between text-xs"
         :class="storage.capacity_status === 'CRITICAL' ? 'bg-rose-950/40 border-rose-500/40 text-rose-300' : 'bg-amber-950/40 border-amber-500/40 text-amber-300'">
      <div class="flex items-center space-x-3">
        <svg class="w-5 h-5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
        </svg>
        <div>
          <span class="font-bold uppercase tracking-wider block font-sans">
            STORAGE DISK {{ storage.capacity_status }}: {{ storage.filesystem_used_percent.toFixed(1) }}% UTILIZATION
          </span>
          <span class="text-2xs opacity-90 block mt-0.5 font-sans">
            Only {{ formatBytes(storage.filesystem_free_bytes) }} available. Enable configured retention policies or run housekeeping cleanup to avoid disk pressure.
          </span>
        </div>
      </div>
      <button @click="activeTab = 'storage'" class="underline text-xs font-semibold hover:opacity-80">
        Inspect
      </button>
    </div>

    <!-- Error Alert Banner -->
    <div v-if="errorStorage || errorPolicies" class="p-3.5 rounded-xl bg-rose-500/10 border border-rose-500/30 flex items-center justify-between text-xs text-rose-300">
      <div class="flex items-center space-x-2">
        <svg class="w-4 h-4 text-rose-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
        </svg>
        <span>{{ errorStorage || errorPolicies }}</span>
      </div>
      <button @click="refreshAll" class="px-2.5 py-1 rounded bg-rose-600 hover:bg-rose-500 text-white text-2xs font-semibold transition-colors">
        {{ $t('common.retry') || 'Retry' }}
      </button>
    </div>

    <!-- Overview Metric KPI Cards -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
      <!-- Total Database Size -->
      <div class="p-4 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-sm space-y-1">
        <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-sans block">{{ $t('retention.totalDatabase') }}</span>
        <div class="text-lg font-bold text-slate-900 dark:text-white font-mono">
          <span v-if="loadingStorage" class="animate-pulse text-slate-400">Loading...</span>
          <span v-else-if="storage">{{ storage.total_database_size_mb.toFixed(2) }} MB</span>
          <span v-else class="text-slate-400">N/A</span>
        </div>
        <span class="text-3xs text-slate-500 dark:text-slate-400 block font-sans">
          <span v-if="loadingStorage">...</span>
          <span v-else-if="storage">{{ storage.tables ? storage.tables.length : 0 }} Monitored Tables</span>
          <span v-else>--</span>
        </span>
      </div>

      <!-- Capacity Status -->
      <div class="p-4 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-sm space-y-1">
        <div class="flex items-center justify-between">
          <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-sans">{{ $t('retention.diskCapacity') }}</span>
          <span class="w-2 h-2 rounded-full"
                :class="storage && storage.capacity_status === 'HEALTHY' ? 'bg-emerald-400' : (storage && storage.capacity_status === 'WARNING' ? 'bg-amber-400' : (storage && storage.capacity_status === 'CRITICAL' ? 'bg-rose-400' : 'bg-slate-400'))"></span>
        </div>
        <div class="text-sm font-bold font-sans flex items-center space-x-1.5"
             :class="storage && storage.capacity_status === 'HEALTHY' ? 'text-emerald-500 dark:text-emerald-400' : (storage && storage.capacity_status === 'WARNING' ? 'text-amber-500 dark:text-amber-400' : (storage && storage.capacity_status === 'CRITICAL' ? 'text-rose-500 dark:text-rose-400' : 'text-slate-400'))">
          <span v-if="loadingStorage" class="animate-pulse">Loading...</span>
          <template v-else-if="storage && storage.filesystem_total_bytes > 0">
            <span>{{ storage.capacity_status }}</span>
            <span class="text-xs text-slate-500 dark:text-slate-400 font-mono font-normal">({{ storage.filesystem_used_percent.toFixed(1) }}%)</span>
          </template>
          <span v-else>N/A</span>
        </div>
        <div class="w-full bg-slate-100 dark:bg-slate-800 rounded-full h-1 mt-1 overflow-hidden">
          <div class="h-1 rounded-full transition-all duration-300"
               :class="storage && storage.capacity_status === 'HEALTHY' ? 'bg-emerald-500' : (storage && storage.capacity_status === 'WARNING' ? 'bg-amber-500' : 'bg-rose-500')"
               :style="{ width: (storage && storage.filesystem_total_bytes > 0 ? Math.min(100, storage.filesystem_used_percent) : 0) + '%' }"></div>
        </div>
      </div>

      <!-- WAL Queue Spool -->
      <div class="p-4 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-sm space-y-1">
        <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-sans block">{{ $t('retention.walSpool') }}</span>
        <div class="text-lg font-bold text-indigo-500 dark:text-indigo-400 font-mono">
          <span v-if="loadingStorage" class="animate-pulse text-slate-400">Loading...</span>
          <span v-else-if="storage">{{ formatBytes(storage.wal_queue_size_bytes) }}</span>
          <span v-else class="text-slate-400">N/A</span>
        </div>
        <span class="text-3xs text-slate-500 dark:text-slate-400 block font-sans">
          <span v-if="loadingStorage">...</span>
          <span v-else-if="storage">{{ storage.wal_queue_pending_records }} Pending records</span>
          <span v-else>--</span>
        </span>
      </div>

      <!-- Backup Storage -->
      <div class="p-4 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 shadow-sm space-y-1">
        <span class="text-3xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-sans block">{{ $t('retention.backupStorage') }}</span>
        <div class="text-lg font-bold text-blue-500 dark:text-blue-400 font-mono">
          <span v-if="loadingStorage" class="animate-pulse text-slate-400">Loading...</span>
          <span v-else-if="storage">{{ formatBytes(storage.backup_storage_size_bytes) }}</span>
          <span v-else class="text-slate-400">N/A</span>
        </div>
        <span class="text-3xs text-slate-500 dark:text-slate-400 block font-sans">
          <span v-if="loadingStorage">...</span>
          <span v-else-if="storage">{{ storage.backup_archives_count }} Archives</span>
          <span v-else>--</span>
        </span>
      </div>
    </div>

    <!-- TAB 1: RETENTION POLICIES -->
    <div v-if="activeTab === 'policies'" class="space-y-4">
      <div class="rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
        <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <svg class="w-4 h-4 text-indigo-500 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"></path>
            </svg>
            <h2 class="text-sm font-semibold text-slate-900 dark:text-white font-sans">
              Configured Data Lifecycle Policies
            </h2>
          </div>
          <div class="flex items-center space-x-3 text-2xs text-slate-500 dark:text-slate-400 font-sans">
            <span v-if="loadingPolicies" class="animate-pulse">Loading policies...</span>
            <span v-else>{{ policies.length }} Policies Defined • Destructive Actions Require Explicit Authorization</span>
            <button @click="fetchPolicies" class="text-indigo-500 hover:text-indigo-400 font-medium">
              Refresh
            </button>
          </div>
        </div>

        <div class="overflow-x-auto">
          <!-- Loading state -->
          <div v-if="loadingPolicies" class="py-12 text-center text-slate-400 text-xs font-sans">
            <svg class="animate-spin w-5 h-5 mx-auto mb-2 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" stroke-width="4" stroke="currentColor" stroke-dasharray="32" stroke-linecap="round"></circle>
            </svg>
            Loading retention policies...
          </div>
          <!-- Error state -->
          <div v-else-if="errorPolicies" class="py-10 text-center text-xs font-sans space-y-2">
            <div class="text-rose-400 font-medium">{{ errorPolicies }}</div>
            <button @click="fetchPolicies" class="px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-2xs font-semibold transition-colors">
              Retry Loading Policies
            </button>
          </div>
          <!-- Empty state -->
          <div v-else-if="policies.length === 0" class="py-12 text-center text-xs font-sans text-slate-400 space-y-2">
            <p>No retention policies found in database.</p>
            <button @click="triggerSweep" class="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-2xs font-semibold transition-colors">
              Initialize Default Policies
            </button>
          </div>
          <!-- Table when loaded -->
          <table v-else class="w-full text-left text-xs font-sans">
            <thead class="bg-slate-50 dark:bg-slate-900/50 text-slate-500 dark:text-slate-400 text-3xs uppercase tracking-wider border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-5 py-3">Category & Policy</th>
                <th class="px-4 py-3">Retention Window</th>
                <th class="px-4 py-3">Safety Requirements</th>
                <th class="px-4 py-3">Housekeeping Schedule</th>
                <th class="px-4 py-3">State</th>
                <th class="px-4 py-3">Last Result</th>
                <th class="px-5 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 text-slate-700 dark:text-slate-200">
              <tr v-for="policy in policies" :key="policy.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                <!-- Category & Policy Name -->
                <td class="px-5 py-3.5">
                  <div class="flex items-center space-x-2">
                    <span class="px-2 py-0.5 rounded text-3xs font-mono font-semibold"
                          :class="getCategoryBadgeClass(policy.category)">
                      {{ policy.category }}
                    </span>
                  </div>
                  <div class="font-medium text-slate-900 dark:text-white mt-1">{{ policy.name }}</div>
                  <div class="text-3xs text-slate-500 dark:text-slate-400 truncate max-w-xs">{{ policy.description }}</div>
                </td>

                <!-- Retention Window -->
                <td class="px-4 py-3.5">
                  <div class="font-mono font-semibold text-slate-900 dark:text-white">{{ policy.retention_days }} Days</div>
                  <div class="text-3xs text-slate-500 dark:text-slate-400 font-mono">
                    Min age: {{ policy.minimum_age_hours }}h • Buffer: {{ policy.protected_period_days }}d
                  </div>
                </td>

                <!-- Safety Requirements -->
                <td class="px-4 py-3.5">
                  <div class="flex flex-col space-y-1">
                    <span v-if="policy.require_backup" class="inline-flex items-center text-3xs text-blue-600 dark:text-blue-400 font-medium">
                      <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
                      </svg>
                      Verified Backup
                    </span>
                    <span v-else class="text-3xs text-slate-400 dark:text-slate-500">No backup gate</span>

                    <span v-if="policy.require_rollup" class="inline-flex items-center text-3xs text-purple-600 dark:text-purple-400 font-medium">
                      <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"></path>
                      </svg>
                      Verified Rollup
                    </span>
                  </div>
                </td>

                <!-- Schedule -->
                <td class="px-4 py-3.5">
                  <div class="font-mono text-2xs">{{ policy.schedule_time }} UTC</div>
                  <div class="text-3xs text-slate-500 dark:text-slate-400">Every {{ policy.schedule_interval_hours }}h</div>
                </td>

                <!-- State Toggle -->
                <td class="px-4 py-3.5">
                  <button @click="togglePolicy(policy)"
                          class="relative inline-flex flex-shrink-0 h-5 w-9 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none"
                          :class="policy.enabled ? 'bg-emerald-500' : 'bg-slate-300 dark:bg-slate-700'">
                    <span class="inline-block h-4 w-4 rounded-full bg-white shadow transform transition ease-in-out duration-200"
                          :class="policy.enabled ? 'translate-x-4' : 'translate-x-0'"></span>
                  </button>
                  <div class="text-3xs font-semibold mt-0.5" :class="policy.enabled ? 'text-emerald-500' : 'text-slate-400'">
                    {{ policy.enabled ? $t('retention.enabled') : $t('retention.disabled') }}
                  </div>
                </td>

                <!-- Last Result -->
                <td class="px-4 py-3.5">
                  <div class="flex items-center space-x-1.5">
                    <span class="px-1.5 py-0.5 rounded text-3xs font-mono font-semibold"
                          :class="getResultBadgeClass(policy.last_execution_result)">
                      {{ policy.last_execution_result || 'NONE' }}
                    </span>
                  </div>
                  <div v-if="policy.last_execution_time" class="text-3xs text-slate-500 dark:text-slate-400 font-mono mt-0.5">
                    {{ formatDate(policy.last_execution_time) }}
                  </div>
                  <div v-if="policy.last_deleted_count > 0" class="text-3xs text-slate-600 dark:text-slate-300 font-mono">
                    {{ policy.last_deleted_count }} records ({{ policy.last_duration_ms }}ms)
                  </div>
                </td>

                <!-- Actions -->
                <td class="px-5 py-3.5 text-right space-x-1.5 whitespace-nowrap">
                  <button @click="openDryRun(policy)"
                          title="Simulate Dry Run"
                          class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 text-2xs font-medium transition-colors">
                    {{ $t('retention.dryRun') }}
                  </button>
                  <button @click="openEditModal(policy)"
                          title="Edit Policy"
                          class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 text-2xs font-medium transition-colors">
                    {{ $t('common.edit') }}
                  </button>
                  <button @click="openConfirmModal(policy)"
                          title="Execute Cleanup"
                          :disabled="!policy.enabled"
                          class="px-2.5 py-1 rounded bg-rose-600 hover:bg-rose-500 disabled:opacity-30 disabled:hover:bg-rose-600 text-white text-2xs font-semibold shadow-sm transition-colors">
                    {{ $t('retention.cleanNow') }}
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 2: STORAGE BREAKDOWN -->
    <div v-if="activeTab === 'storage'" class="space-y-6">
      <!-- Monitored Tables Breakdown -->
      <div class="rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
        <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <svg class="w-4 h-4 text-indigo-500 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 7v10c0 2 1.5 3 3.5 3h9c2 0 3.5-1 3.5-3V7M4 7c0-2 1.5-3 3.5-3h9c2 0 3.5 1 3.5 3M4 7h16m-8 4v6m0 0l-3-3m3 3l3-3"></path>
            </svg>
            <h2 class="text-sm font-semibold text-slate-900 dark:text-white font-sans">
              MariaDB Database Table Storage Footprint
            </h2>
          </div>
          <span class="text-2xs text-slate-500 dark:text-slate-400 font-mono">
            DB: {{ storage ? storage.database_name + ' (' + storage.database_type + ')' : '...' }}
          </span>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs font-sans">
            <thead class="bg-slate-50 dark:bg-slate-900/50 text-slate-500 dark:text-slate-400 text-3xs uppercase tracking-wider border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-5 py-3">Table Name</th>
                <th class="px-4 py-3">Domain Category</th>
                <th class="px-4 py-3 text-right">Row Count</th>
                <th class="px-4 py-3 text-right">Data Size</th>
                <th class="px-4 py-3 text-right">Index Size</th>
                <th class="px-4 py-3 text-right">Total Size</th>
                <th class="px-5 py-3 text-center">Measurement</th>
              </tr>
            </thead>
            <tbody v-if="loadingStorage">
              <tr>
                <td colspan="7" class="px-5 py-8 text-center text-xs text-slate-400">
                  <svg class="animate-spin w-5 h-5 mx-auto mb-2 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <circle cx="12" cy="12" r="10" stroke-width="4" stroke="currentColor" stroke-dasharray="32" stroke-linecap="round"></circle>
                  </svg>
                  Loading table storage metrics...
                </td>
              </tr>
            </tbody>
            <tbody v-else-if="errorStorage">
              <tr>
                <td colspan="7" class="px-5 py-6 text-center text-xs text-rose-400 font-sans space-y-2">
                  <div>{{ errorStorage }}</div>
                  <button @click="fetchStorage" class="px-3 py-1 bg-indigo-600 hover:bg-indigo-500 text-white rounded text-2xs font-semibold">
                    Retry Loading Storage
                  </button>
                </td>
              </tr>
            </tbody>
            <tbody v-else-if="storage && storage.tables && storage.tables.length" class="divide-y divide-slate-100 dark:divide-slate-800/60 text-slate-700 dark:text-slate-200 font-mono">
              <tr v-for="tbl in storage.tables" :key="tbl.table_name" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-5 py-3 font-semibold text-slate-900 dark:text-white font-sans">{{ tbl.table_name }}</td>
                <td class="px-4 py-3 text-2xs text-slate-500 dark:text-slate-400 font-sans">{{ tbl.category }}</td>
                <td class="px-4 py-3 text-right">{{ tbl.row_count.toLocaleString() }}</td>
                <td class="px-4 py-3 text-right text-slate-500 dark:text-slate-400">{{ formatBytes(tbl.data_size_bytes) }}</td>
                <td class="px-4 py-3 text-right text-slate-500 dark:text-slate-400">{{ formatBytes(tbl.index_size_bytes) }}</td>
                <td class="px-4 py-3 text-right font-bold text-indigo-600 dark:text-indigo-400">{{ tbl.total_size_mb.toFixed(2) }} MB</td>
                <td class="px-5 py-3 text-center">
                  <span class="px-2 py-0.5 rounded text-3xs font-sans font-medium"
                        :class="tbl.is_estimated ? 'bg-amber-100 dark:bg-amber-950/60 text-amber-700 dark:text-amber-400 border border-amber-300 dark:border-amber-800' : 'bg-emerald-100 dark:bg-emerald-950/60 text-emerald-700 dark:text-emerald-400 border border-emerald-300 dark:border-emerald-800'">
                    {{ tbl.is_estimated ? $t('retention.estimated') : $t('retention.measured') }}
                  </span>
                </td>
              </tr>
            </tbody>
            <tbody v-else>
              <tr>
                <td colspan="7" class="px-5 py-8 text-center text-xs text-slate-400 dark:text-slate-500">
                  No database table storage metrics available.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- TAB 3: EXECUTION HISTORY -->
    <div v-if="activeTab === 'history'" class="space-y-4">
      <div class="rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 overflow-hidden shadow-sm">
        <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800/80 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <svg class="w-4 h-4 text-indigo-500 dark:text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <h2 class="text-sm font-semibold text-slate-900 dark:text-white font-sans">
              Retention Execution & Dry-Run Audit History
            </h2>
          </div>
          <span class="text-2xs text-slate-500 dark:text-slate-400 font-sans">
            Page {{ historyPage }} • Total {{ historyTotal }} Operations
          </span>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs font-sans">
            <thead class="bg-slate-50 dark:bg-slate-900/50 text-slate-500 dark:text-slate-400 text-3xs uppercase tracking-wider border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-5 py-3">Timestamp</th>
                <th class="px-4 py-3">Policy / Category</th>
                <th class="px-4 py-3">Type</th>
                <th class="px-4 py-3">Status</th>
                <th class="px-4 py-3">Impact</th>
                <th class="px-4 py-3">Evidence & Blocking</th>
                <th class="px-4 py-3">Duration</th>
                <th class="px-5 py-3">Initiated By</th>
              </tr>
            </thead>
            <tbody v-if="loadingHistory">
              <tr>
                <td colspan="8" class="px-5 py-8 text-center text-xs text-slate-400">
                  <svg class="animate-spin w-5 h-5 mx-auto mb-2 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <circle cx="12" cy="12" r="10" stroke-width="4" stroke="currentColor" stroke-dasharray="32" stroke-linecap="round"></circle>
                  </svg>
                  Loading audit history...
                </td>
              </tr>
            </tbody>
            <tbody v-else-if="errorHistory">
              <tr>
                <td colspan="8" class="px-5 py-6 text-center text-xs text-rose-400 font-sans space-y-2">
                  <div>{{ errorHistory }}</div>
                  <button @click="fetchHistory" class="px-3 py-1 bg-indigo-600 hover:bg-indigo-500 text-white rounded text-2xs font-semibold">
                    Retry Loading History
                  </button>
                </td>
              </tr>
            </tbody>
            <tbody v-else-if="historyLogs && historyLogs.length" class="divide-y divide-slate-100 dark:divide-slate-800/60 text-slate-700 dark:text-slate-200">
              <tr v-for="log in historyLogs" :key="log.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30 transition-colors">
                <td class="px-5 py-3 font-mono text-2xs whitespace-nowrap">{{ formatDate(log.created_at) }}</td>
                <td class="px-4 py-3">
                  <div class="font-medium text-slate-900 dark:text-white font-sans">{{ log.policy_id }}</div>
                  <div class="text-3xs font-mono text-slate-500 dark:text-slate-400">{{ log.category }}</div>
                </td>
                <td class="px-4 py-3">
                  <span class="px-2 py-0.5 rounded text-3xs font-mono font-semibold"
                        :class="log.execution_type === 'DRY_RUN' ? 'bg-blue-100 dark:bg-blue-950/60 text-blue-700 dark:text-blue-400 border border-blue-300 dark:border-blue-800' : 'bg-purple-100 dark:bg-purple-950/60 text-purple-700 dark:text-purple-400 border border-purple-300 dark:border-purple-800'">
                    {{ log.execution_type }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="px-2 py-0.5 rounded text-3xs font-mono font-semibold"
                        :class="getResultBadgeClass(log.status)">
                    {{ log.status }}
                  </span>
                </td>
                <td class="px-4 py-3 font-mono text-2xs">
                  <div v-if="log.execution_type === 'DRY_RUN'">
                    {{ log.candidate_count }} candidates ({{ formatBytes(log.bytes_recovered_estimate) }})
                  </div>
                  <div v-else class="font-semibold text-slate-900 dark:text-white">
                    {{ log.deleted_count }} deleted
                  </div>
                </td>
                <td class="px-4 py-3 text-3xs max-w-xs truncate" :title="log.blocking_reasons || log.error_message || log.backup_evidence">
                  <div v-if="log.blocking_reasons" class="text-rose-500 dark:text-rose-400 truncate">{{ log.blocking_reasons }}</div>
                  <div v-else-if="log.error_message" class="text-rose-500 dark:text-rose-400 truncate">{{ log.error_message }}</div>
                  <div v-else-if="log.backup_evidence" class="text-emerald-600 dark:text-emerald-400 truncate">{{ log.backup_evidence }}</div>
                  <div v-else class="text-slate-400 dark:text-slate-500">-</div>
                </td>
                <td class="px-4 py-3 font-mono text-2xs">{{ log.duration_ms }}ms</td>
                <td class="px-5 py-3 text-2xs font-mono text-slate-500 dark:text-slate-400">{{ log.initiated_by }}</td>
              </tr>
            </tbody>
            <tbody v-else>
              <tr>
                <td colspan="8" class="px-5 py-8 text-center text-xs text-slate-400 dark:text-slate-500">
                  {{ $t('retention.noHistory') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- History Pagination -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-slate-800/80 flex items-center justify-between text-xs text-slate-500 dark:text-slate-400">
          <button @click="changeHistoryPage(historyPage - 1)"
                  :disabled="historyPage <= 1"
                  class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 disabled:opacity-40 text-2xs font-medium">
            Previous
          </button>
          <span class="font-mono text-2xs">Page {{ historyPage }}</span>
          <button @click="changeHistoryPage(historyPage + 1)"
                  :disabled="historyLogs.length < historyPageSize"
                  class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 disabled:opacity-40 text-2xs font-medium">
            Next
          </button>
        </div>
      </div>
    </div>

    <!-- DRY RUN MODAL -->
    <div v-if="showDryRunModal && dryRunData"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl space-y-4 p-6">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <div class="flex items-center space-x-2">
            <svg class="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
              <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
            </svg>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              {{ $t('retention.dryRunTitle') }}
            </h3>
          </div>
          <button @click="showDryRunModal = false" class="text-slate-400 hover:text-white">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
          </button>
        </div>

        <div class="space-y-3 text-xs font-sans">
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-1.5 font-mono">
            <div class="flex justify-between">
              <span class="text-slate-500 dark:text-slate-400 font-sans">Policy:</span>
              <span class="font-bold text-slate-900 dark:text-white font-sans">{{ dryRunData.policy_name }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500 dark:text-slate-400 font-sans">Data Category:</span>
              <span class="text-indigo-400">{{ dryRunData.category }}</span>
            </div>
            <div class="flex justify-between">
              <span class="text-slate-500 dark:text-slate-400 font-sans">Cutoff Boundary:</span>
              <span class="text-slate-200">{{ formatDate(dryRunData.cutoff_timestamp) }}</span>
            </div>
          </div>

          <!-- Candidates & Impact -->
          <div class="grid grid-cols-2 gap-3 font-mono">
            <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
              <span class="text-3xs uppercase text-slate-400 block font-sans">Candidate Records</span>
              <span class="text-base font-bold text-slate-900 dark:text-white">{{ dryRunData.candidate_count.toLocaleString() }}</span>
            </div>
            <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800">
              <span class="text-3xs uppercase text-slate-400 block font-sans">Est. Recovery</span>
              <span class="text-base font-bold text-emerald-500 dark:text-emerald-400">{{ dryRunData.estimated_mb.toFixed(2) }} MB</span>
            </div>
          </div>

          <!-- Safety Checks Status -->
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 space-y-2">
            <!-- Backup check -->
            <div class="flex items-start space-x-2">
              <span class="w-4 h-4 rounded-full flex items-center justify-center text-3xs font-bold flex-shrink-0 mt-0.5"
                    :class="dryRunData.backup_safety_verified ? 'bg-emerald-500 text-white' : 'bg-rose-500 text-white'">
                {{ dryRunData.backup_safety_verified ? '✓' : '✗' }}
              </span>
              <div>
                <span class="font-semibold block text-slate-900 dark:text-white">Backup Safety</span>
                <span class="text-3xs text-slate-500 dark:text-slate-400 block">{{ dryRunData.backup_evidence }}</span>
              </div>
            </div>

            <!-- Rollup check -->
            <div class="flex items-start space-x-2">
              <span class="w-4 h-4 rounded-full flex items-center justify-center text-3xs font-bold flex-shrink-0 mt-0.5"
                    :class="dryRunData.rollup_safety_verified ? 'bg-emerald-500 text-white' : 'bg-rose-500 text-white'">
                {{ dryRunData.rollup_safety_verified ? '✓' : '✗' }}
              </span>
              <div>
                <span class="font-semibold block text-slate-900 dark:text-white">Rollup Downsampling Safety</span>
                <span class="text-3xs text-slate-500 dark:text-slate-400 block">{{ dryRunData.rollup_evidence }}</span>
              </div>
            </div>
          </div>

          <!-- Blocking Reasons Alert (if any) -->
          <div v-if="dryRunData.blocking_reasons && dryRunData.blocking_reasons.length"
               class="p-3 rounded-lg bg-rose-950/40 border border-rose-500/40 text-rose-300 text-2xs space-y-1">
            <span class="font-bold block uppercase tracking-wider font-sans">Execution Blocked:</span>
            <ul class="list-disc list-inside space-y-0.5 font-mono">
              <li v-for="(reason, idx) in dryRunData.blocking_reasons" :key="idx">{{ reason }}</li>
            </ul>
          </div>
        </div>

        <div class="flex items-center justify-end space-x-2 border-t border-slate-200 dark:border-slate-800 pt-3">
          <button @click="showDryRunModal = false"
                  class="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
            {{ $t('common.close') }}
          </button>
          <button @click="proceedFromDryRunToClean"
                  :disabled="!dryRunData.can_execute"
                  class="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 disabled:opacity-40 text-white text-xs font-semibold shadow-sm transition-colors">
            {{ $t('retention.clean') }}
          </button>
        </div>
      </div>
    </div>

    <!-- EXPLICIT CONFIRMATION CLEANUP MODAL -->
    <div v-if="showConfirmModal && selectedPolicy"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-md overflow-hidden shadow-2xl space-y-4 p-6">
        <div class="flex items-center space-x-3 text-rose-500">
          <div class="w-9 h-9 rounded-full bg-rose-100 dark:bg-rose-950/80 flex items-center justify-center">
            <svg class="w-5 h-5 text-rose-600 dark:text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
            </svg>
          </div>
          <div>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              {{ $t('retention.confirmCleanupTitle') }}
            </h3>
            <span class="text-3xs text-slate-500 dark:text-slate-400 font-sans block">Destructive Bounded Deletion</span>
          </div>
        </div>

        <p class="text-xs text-slate-600 dark:text-slate-300 font-sans">
          {{ $t('retention.confirmCleanupDesc') }}
        </p>

        <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-2xs space-y-1 font-mono">
          <div><span class="text-slate-400">Target Category:</span> <span class="text-indigo-400 font-bold">{{ selectedPolicy.category }}</span></div>
          <div><span class="text-slate-400">Retention Days:</span> <span class="text-slate-200">{{ selectedPolicy.retention_days }} days</span></div>
          <div><span class="text-slate-400">Batch Size:</span> <span class="text-slate-200">{{ selectedPolicy.batch_size }} records / chunk</span></div>
        </div>

        <div class="flex items-start space-x-2 pt-1">
          <input type="checkbox" id="confirmExplicit" v-model="confirmExplicitCheck"
                 class="mt-0.5 rounded border-slate-300 dark:border-slate-700 text-indigo-600 focus:ring-indigo-500">
          <label for="confirmExplicit" class="text-2xs text-slate-600 dark:text-slate-300 cursor-pointer select-none">
            {{ $t('retention.explicitConfirmCheck') }}
          </label>
        </div>

        <div class="flex items-center justify-end space-x-2 border-t border-slate-200 dark:border-slate-800 pt-3">
          <button @click="showConfirmModal = false"
                  class="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
            {{ $t('common.cancel') }}
          </button>
          <button @click="executeCleanup"
                  :disabled="!confirmExplicitCheck || isCleaning"
                  class="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 disabled:opacity-40 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
            <span v-if="isCleaning" class="animate-spin w-3 h-3 mr-1.5 border-2 border-white border-t-transparent rounded-full"></span>
            {{ isCleaning ? $t('retention.cleaning') : $t('retention.confirmCleanupBtn') }}
          </button>
        </div>
      </div>
    </div>

    <!-- EDIT POLICY MODAL -->
    <div v-if="showEditModal && editingPolicy"
         class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl space-y-4 p-6">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
            {{ $t('retention.editPolicy') }}: {{ editingPolicy.name }}
          </h3>
          <button @click="showEditModal = false" class="text-slate-400 hover:text-white">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
          </button>
        </div>

        <form @submit.prevent="savePolicyConfig" class="space-y-3 text-xs font-sans">
          <div>
            <label class="block text-slate-500 dark:text-slate-400 mb-1">Policy Name</label>
            <input v-model="editingPolicy.name" required
                   class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white">
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('retention.retentionDays') }}</label>
              <input type="number" min="1" v-model.number="editingPolicy.retention_days" required
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono">
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('retention.minimumAge') }}</label>
              <input type="number" min="1" v-model.number="editingPolicy.minimum_age_hours" required
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono">
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('retention.protectedPeriod') }}</label>
              <input type="number" min="0" v-model.number="editingPolicy.protected_period_days"
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono">
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('retention.batchSize') }}</label>
              <input type="number" min="10" max="5000" v-model.number="editingPolicy.batch_size" required
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono">
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">{{ $t('retention.scheduleTime') }}</label>
              <input type="text" placeholder="03:30" v-model="editingPolicy.schedule_time"
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono">
            </div>
            <div>
              <label class="block text-slate-500 dark:text-slate-400 mb-1">Interval (Hours)</label>
              <input type="number" min="1" v-model.number="editingPolicy.schedule_interval_hours"
                     class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 text-slate-900 dark:text-white font-mono">
            </div>
          </div>

          <!-- Toggles for Safety Gates -->
          <div class="space-y-2 pt-2 border-t border-slate-200 dark:border-slate-800">
            <div class="flex items-center justify-between">
              <span class="text-slate-600 dark:text-slate-300 font-medium">Require Verified Backup</span>
              <input type="checkbox" v-model="editingPolicy.require_backup"
                     class="rounded border-slate-300 dark:border-slate-700 text-indigo-600 focus:ring-indigo-500">
            </div>
            <div v-if="editingPolicy.category === 'RAW_TELEMETRY'" class="flex items-center justify-between">
              <span class="text-slate-600 dark:text-slate-300 font-medium">Require Verified Rollup (Aggregations)</span>
              <input type="checkbox" v-model="editingPolicy.require_rollup"
                     class="rounded border-slate-300 dark:border-slate-700 text-indigo-600 focus:ring-indigo-500">
            </div>
            <div class="flex items-center justify-between">
              <span class="text-slate-600 dark:text-slate-300 font-medium">Policy Enabled</span>
              <input type="checkbox" v-model="editingPolicy.enabled"
                     class="rounded border-slate-300 dark:border-slate-700 text-indigo-600 focus:ring-indigo-500">
            </div>
          </div>

          <div class="flex items-center justify-end space-x-2 border-t border-slate-200 dark:border-slate-800 pt-3">
            <button type="button" @click="showEditModal = false"
                    class="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 text-xs font-semibold">
              {{ $t('common.cancel') }}
            </button>
            <button type="submit" :disabled="isSaving"
                    class="px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm transition-colors">
              {{ isSaving ? $t('retention.saving') : $t('retention.savePolicy') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'RetentionStorageView',
  data() {
    return {
      activeTab: 'policies',
      policies: [],
      storage: null,
      historyLogs: [],
      historyTotal: 0,
      historyPage: 1,
      historyPageSize: 15,
      isSweeping: false,
      isCleaning: false,
      isSaving: false,

      // Loading and Error states
      loadingPolicies: false,
      loadingStorage: false,
      loadingHistory: false,
      errorPolicies: null,
      errorStorage: null,
      errorHistory: null,

      // Modals
      showDryRunModal: false,
      dryRunData: null,
      showConfirmModal: false,
      selectedPolicy: null,
      confirmExplicitCheck: false,
      showEditModal: false,
      editingPolicy: null,
    };
  },
  mounted() {
    this.refreshAll();
  },
  methods: {
    getAuthHeaders() {
      const token = (this.$store && this.$store.state.token) ||
                    localStorage.getItem('datalogger_token') ||
                    localStorage.getItem('token');
      return token ? { Authorization: `Bearer ${token}` } : {};
    },
    refreshAll() {
      this.fetchPolicies();
      this.fetchStorage();
      this.fetchHistory();
    },
    async fetchPolicies() {
      this.loadingPolicies = true;
      this.errorPolicies = null;
      try {
        const res = await axios.get('/api/retention/policies', {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.success) {
          this.policies = res.data.data || [];
        } else {
          this.errorPolicies = (res.data && res.data.error) || 'Failed to load retention policies.';
        }
      } catch (err) {
        console.error('Failed fetching retention policies:', err);
        const errMsg = err.response && err.response.data && err.response.data.error
          ? err.response.data.error
          : (err.message || 'Unknown network error');
        const status = err.response ? `[HTTP ${err.response.status}] ` : '';
        this.errorPolicies = `${status}${errMsg}`;
      } finally {
        this.loadingPolicies = false;
      }
    },
    async fetchStorage() {
      this.loadingStorage = true;
      this.errorStorage = null;
      try {
        const res = await axios.get('/api/retention/storage', {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.success) {
          this.storage = res.data.data;
        } else {
          this.errorStorage = (res.data && res.data.error) || 'Failed to load storage metrics.';
        }
      } catch (err) {
        console.error('Failed fetching storage overview:', err);
        const errMsg = err.response && err.response.data && err.response.data.error
          ? err.response.data.error
          : (err.message || 'Unknown network error');
        const status = err.response ? `[HTTP ${err.response.status}] ` : '';
        this.errorStorage = `${status}${errMsg}`;
      } finally {
        this.loadingStorage = false;
      }
    },
    async fetchHistory() {
      this.loadingHistory = true;
      this.errorHistory = null;
      try {
        const res = await axios.get(`/api/retention/history?page=${this.historyPage}&page_size=${this.historyPageSize}`, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.success) {
          this.historyLogs = res.data.data.items || [];
          this.historyTotal = res.data.data.total || 0;
        } else {
          this.errorHistory = (res.data && res.data.error) || 'Failed to load retention history.';
        }
      } catch (err) {
        console.error('Failed fetching retention history:', err);
        const errMsg = err.response && err.response.data && err.response.data.error
          ? err.response.data.error
          : (err.message || 'Unknown network error');
        const status = err.response ? `[HTTP ${err.response.status}] ` : '';
        this.errorHistory = `${status}${errMsg}`;
      } finally {
        this.loadingHistory = false;
      }
    },
    changeHistoryPage(page) {
      if (page < 1) return;
      this.historyPage = page;
      this.fetchHistory();
    },
    async togglePolicy(policy) {
      try {
        const nextState = !policy.enabled;
        const res = await axios.put(`/api/retention/policies/${policy.id}/toggle`, { enabled: nextState }, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.success) {
          policy.enabled = nextState;
        } else {
          alert('Failed to toggle policy: ' + ((res.data && res.data.error) || 'Unknown error'));
        }
      } catch (err) {
        const errMsg = err.response && err.response.data && err.response.data.error ? err.response.data.error : err.message;
        alert(`Error toggling policy: ${errMsg}`);
      }
    },
    async triggerSweep() {
      this.isSweeping = true;
      try {
        const res = await axios.post('/api/retention/housekeeping/trigger', {}, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.success) {
          await this.refreshAll();
        } else {
          alert('Trigger failed: ' + ((res.data && res.data.error) || 'Unknown error'));
        }
      } catch (err) {
        const errMsg = err.response && err.response.data && err.response.data.error ? err.response.data.error : err.message;
        alert(`Error triggering housekeeping: ${errMsg}`);
      } finally {
        this.isSweeping = false;
      }
    },
    async openDryRun(policy) {
      try {
        const res = await axios.post(`/api/retention/policies/${policy.id}/dry-run`, {}, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.success) {
          this.dryRunData = res.data.data;
          this.selectedPolicy = policy;
          this.showDryRunModal = true;
        } else {
          alert('Dry run simulation failed: ' + ((res.data && res.data.error) || 'Unknown error'));
        }
      } catch (err) {
        const errMsg = err.response && err.response.data && err.response.data.error ? err.response.data.error : err.message;
        alert(`Dry run error: ${errMsg}`);
      }
    },
    proceedFromDryRunToClean() {
      this.showDryRunModal = false;
      this.openConfirmModal(this.selectedPolicy);
    },
    openConfirmModal(policy) {
      this.selectedPolicy = policy;
      this.confirmExplicitCheck = false;
      this.showConfirmModal = true;
    },
    async executeCleanup() {
      if (!this.confirmExplicitCheck) return;
      this.isCleaning = true;
      try {
        const res = await axios.post(`/api/retention/policies/${this.selectedPolicy.id}/execute`, { confirm: true }, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.success) {
          this.showConfirmModal = false;
          await this.refreshAll();
        } else {
          alert('Execution Blocked: ' + (res.data ? res.data.error : 'Unknown'));
        }
      } catch (err) {
        const errMsg = err.response && err.response.data && err.response.data.error ? err.response.data.error : err.message;
        alert(`Cleanup error: ${errMsg}`);
      } finally {
        this.isCleaning = false;
      }
    },
    openEditModal(policy) {
      this.editingPolicy = JSON.parse(JSON.stringify(policy));
      this.showEditModal = true;
    },
    async savePolicyConfig() {
      this.isSaving = true;
      try {
        const res = await axios.put(`/api/retention/policies/${this.editingPolicy.id}`, this.editingPolicy, {
          headers: this.getAuthHeaders(),
        });
        if (res.data && res.data.success) {
          this.showEditModal = false;
          await this.fetchPolicies();
        } else {
          alert('Failed to save policy: ' + ((res.data && res.data.error) || 'Unknown error'));
        }
      } catch (err) {
        const errMsg = err.response && err.response.data && err.response.data.error ? err.response.data.error : err.message;
        alert(`Save policy error: ${errMsg}`);
      } finally {
        this.isSaving = false;
      }
    },
    getCategoryBadgeClass(category) {
      switch (category) {
        case 'RAW_TELEMETRY':
          return 'bg-blue-100 dark:bg-blue-950/80 text-blue-700 dark:text-blue-400 border border-blue-300 dark:border-blue-800';
        case 'INTERNAL_AGGREGATIONS':
          return 'bg-purple-100 dark:bg-purple-950/80 text-purple-700 dark:text-purple-400 border border-purple-300 dark:border-purple-800';
        case 'CUSTOMER_AGGREGATIONS':
          return 'bg-emerald-100 dark:bg-emerald-950/80 text-emerald-700 dark:text-emerald-400 border border-emerald-300 dark:border-emerald-800';
        case 'CLEARED_ALARMS':
          return 'bg-rose-100 dark:bg-rose-950/80 text-rose-700 dark:text-rose-400 border border-rose-300 dark:border-rose-800';
        case 'COMMUNICATION_LOGS':
        case 'SYSTEM_LOGS':
          return 'bg-amber-100 dark:bg-amber-950/80 text-amber-700 dark:text-amber-400 border border-amber-300 dark:border-amber-800';
        case 'AUDIT_TRAILS':
          return 'bg-cyan-100 dark:bg-cyan-950/80 text-cyan-700 dark:text-cyan-400 border border-cyan-300 dark:border-cyan-800';
        case 'BACKUP_ARCHIVES':
          return 'bg-teal-100 dark:bg-teal-950/80 text-teal-700 dark:text-teal-400 border border-teal-300 dark:border-teal-800';
        default:
          return 'bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300';
      }
    },
    getResultBadgeClass(result) {
      switch (result) {
        case 'SUCCESS':
        case 'COMPLETED':
          return 'bg-emerald-100 dark:bg-emerald-950 text-emerald-700 dark:text-emerald-400 border border-emerald-300 dark:border-emerald-800';
        case 'BLOCKED':
          return 'bg-amber-100 dark:bg-amber-950 text-amber-700 dark:text-amber-400 border border-amber-300 dark:border-amber-800';
        case 'FAILED':
          return 'bg-rose-100 dark:bg-rose-950 text-rose-700 dark:text-rose-400 border border-rose-300 dark:border-rose-800';
        case 'RUNNING':
          return 'bg-blue-100 dark:bg-blue-950 text-blue-700 dark:text-blue-400 border border-blue-300 dark:border-blue-800';
        default:
          return 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400';
      }
    },
    formatBytes(bytes) {
      if (bytes === undefined || bytes === null || isNaN(bytes)) return '0 B';
      if (bytes === 0) return '0 B';
      const k = 1024;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    },
    formatDate(dateStr) {
      if (!dateStr) return '-';
      const d = new Date(dateStr);
      if (isNaN(d.getTime())) return '-';
      return d.toLocaleString('en-GB', {
        year: 'numeric',
        month: 'short',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      });
    },
  },
};
</script>
