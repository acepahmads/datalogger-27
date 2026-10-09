<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans flex items-center gap-2.5">
          <svg class="w-5 h-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M4 7v10c0 2 1.5 3 3.5 3h9c2 0 3.5-1 3.5-3V7M4 7c0-2 1.5-3 3.5-3h9c2 0 3.5 1 3.5 3M4 7h16m-8 4v6m0 0l-3-3m3 3l3-3"></path>
          </svg>
          {{ $t('backup.title') }}
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          {{ $t('backup.subtitle') }}
        </p>
      </div>

      <!-- Action Buttons & Tabs -->
      <div class="flex items-center space-x-2">
        <button @click="showCreateModal = true"
                class="px-3.5 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
          <svg class="w-3.5 h-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.5">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          {{ $t('backup.createBackup') }}
        </button>

        <div class="flex items-center space-x-1 bg-[#111827] p-1 rounded-lg border border-slate-800 text-xs font-sans">
          <button @click="activeTab = 'catalog'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'catalog' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
            {{ $t('backup.tabCatalog') }}
          </button>
          <button @click="activeTab = 'schedule'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'schedule' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
            {{ $t('backup.tabSchedule') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Active Job Progress Banner (if any running job) -->
    <div v-if="activeJob && (activeJob.status === 'RUNNING' || activeJob.status === 'PENDING')"
         class="saas-card p-4 bg-gradient-to-r from-blue-950/60 to-indigo-950/50 border border-blue-500/30 flex items-center justify-between">
      <div class="flex items-center space-x-3">
        <span class="flex h-3 w-3 relative">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-3 w-3 bg-blue-500"></span>
        </span>
        <div>
          <span class="text-xs font-bold text-white uppercase tracking-wider block">
            {{ activeJob.type }} OPERATION IN PROGRESS ({{ activeJob.stage }})
          </span>
          <span class="text-2xs text-slate-300 font-sans block mt-0.5">
            Job ID: <code class="font-mono text-blue-300">{{ activeJob.id }}</code> • Started: {{ activeJob.started_at | formatDate }}
          </span>
        </div>
      </div>
      <div class="text-right">
        <span class="text-xs font-bold font-mono text-blue-400">{{ activeJob.progress || 0 }}%</span>
        <div class="w-32 bg-slate-800 rounded-full h-1.5 mt-1 overflow-hidden">
          <div class="bg-blue-500 h-1.5 rounded-full transition-all duration-300" :style="{ width: (activeJob.progress || 10) + '%' }"></div>
        </div>
      </div>
    </div>

    <!-- TAB 1: Backup Catalog & Overview -->
    <div v-if="activeTab === 'catalog'" class="space-y-6">
      <!-- Overview Metric Cards -->
      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
        <!-- Subsystem Health -->
        <div class="saas-card p-4 space-y-1">
          <div class="flex items-center justify-between">
            <span class="text-3xs uppercase tracking-wider text-slate-400 font-sans">{{ $t('backup.healthStatus') }}</span>
            <span class="w-2 h-2 rounded-full" :class="statusData.subsystem_healthy ? 'bg-emerald-400' : 'bg-rose-400'"></span>
          </div>
          <div class="text-sm font-bold text-white font-sans flex items-center space-x-1.5">
            <span :class="statusData.subsystem_healthy ? 'text-emerald-400' : 'text-rose-400'">
              {{ statusData.subsystem_healthy ? 'HEALTHY' : 'DEGRADED' }}
            </span>
          </div>
          <span class="text-3xs text-slate-400 block font-mono">
            MariaDB + WAL Spool
          </span>
        </div>

        <!-- Total Archives -->
        <div class="saas-card p-4 space-y-1">
          <span class="text-3xs uppercase tracking-wider text-slate-400 font-sans block">{{ $t('backup.totalBackups') }}</span>
          <div class="text-lg font-bold text-white font-mono">
            {{ statusData.total_backups || backupsList.length || 0 }}
          </div>
          <span class="text-3xs text-slate-400 block font-sans">
            {{ validCount }} Validated
          </span>
        </div>

        <!-- Storage Used -->
        <div class="saas-card p-4 space-y-1">
          <span class="text-3xs uppercase tracking-wider text-slate-400 font-sans block">{{ $t('backup.storageUsed') }}</span>
          <div class="text-lg font-bold text-blue-400 font-mono">
            {{ formatBytes(statusData.storage_used_bytes || 0) }}
          </div>
          <span class="text-3xs text-slate-400 block font-sans">
            Free: {{ formatBytes(statusData.free_disk_bytes || 0) }}
          </span>
        </div>

        <!-- Last Backup -->
        <div class="saas-card p-4 space-y-1">
          <span class="text-3xs uppercase tracking-wider text-slate-400 font-sans block">{{ $t('backup.lastBackup') }}</span>
          <div class="text-xs font-semibold text-slate-200 font-mono truncate">
            {{ statusData.last_successful_backup ? formatDate(statusData.last_successful_backup) : 'Never' }}
          </div>
          <span class="text-3xs text-slate-400 block font-sans">
            Next: {{ statusData.next_scheduled_run ? formatDate(statusData.next_scheduled_run) : 'Disabled' }}
          </span>
        </div>
      </div>

      <!-- Catalog Table -->
      <div class="saas-card overflow-hidden">
        <div class="p-4 border-b border-slate-800 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <h2 class="text-sm font-bold text-white font-sans">{{ $t('backup.tabCatalog') }}</h2>
            <span class="text-3xs font-mono px-2 py-0.5 rounded-full bg-slate-800 text-slate-300 border border-slate-700">
              {{ totalCount }} items
            </span>
          </div>

          <div class="flex items-center space-x-2">
            <button @click="fetchBackups" :disabled="loading"
                    class="p-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs transition-colors flex items-center"
                    title="Refresh">
              <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <polyline points="23 4 23 10 17 10"></polyline>
                <polyline points="1 20 1 14 7 14"></polyline>
                <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
              </svg>
            </button>
          </div>
        </div>

        <!-- Table View -->
        <div class="overflow-x-auto">
          <table class="w-full text-left border-collapse text-xs font-sans">
            <thead>
              <tr class="border-b border-slate-800 text-slate-400 text-3xs uppercase tracking-wider bg-[#0B0F19]/50">
                <th class="py-2.5 px-4 font-semibold">{{ $t('backup.backupId') }}</th>
                <th class="py-2.5 px-3 font-semibold">{{ $t('backup.createdAt') }}</th>
                <th class="py-2.5 px-3 font-semibold">{{ $t('backup.type') }}</th>
                <th class="py-2.5 px-3 font-semibold">{{ $t('backup.size') }}</th>
                <th class="py-2.5 px-3 font-semibold">{{ $t('backup.validationStatus') }}</th>
                <th class="py-2.5 px-3 font-semibold">{{ $t('backup.consistency') }}</th>
                <th class="py-2.5 px-3 font-semibold">{{ $t('backup.restoreCompat') }}</th>
                <th class="py-2.5 px-3 font-semibold text-right">{{ $t('backup.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60">
              <tr v-if="loading && backupsList.length === 0">
                <td colspan="8" class="text-center py-8 text-slate-400">
                  <div class="flex items-center justify-center space-x-2">
                    <svg class="animate-spin w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <circle cx="12" cy="12" r="10" stroke-width="4" stroke="currentColor" stroke-dasharray="32" stroke-linecap="round"></circle>
                    </svg>
                    <span>Loading backup catalog...</span>
                  </div>
                </td>
              </tr>
              <tr v-else-if="backupsList.length === 0">
                <td colspan="8" class="text-center py-8 text-slate-500">
                  {{ $t('backup.noBackups') }}
                </td>
              </tr>
              <tr v-for="b in backupsList" :key="b.backup_id" class="hover:bg-slate-800/30 transition-colors">
                <!-- Backup ID -->
                <td class="py-2.5 px-4 font-mono text-slate-200">
                  <div class="flex items-center space-x-1.5">
                    <span class="w-1.5 h-1.5 rounded-full" :class="getBackupStatusDot(b.status)"></span>
                    <span class="font-semibold text-xs">{{ b.backup_id }}</span>
                  </div>
                  <span v-if="b.notes" class="text-3xs text-slate-400 block truncate max-w-xs mt-0.5">{{ b.notes }}</span>
                </td>

                <!-- Created At -->
                <td class="py-2.5 px-3 text-slate-300 font-mono text-2xs whitespace-nowrap">
                  {{ b.created_at | formatDate }}
                </td>

                <!-- Type Badge -->
                <td class="py-2.5 px-3 whitespace-nowrap">
                  <span class="px-2 py-0.5 rounded text-3xs font-semibold uppercase tracking-wider"
                        :class="getBackupTypeBadge(b.backup_type)">
                    {{ b.backup_type }}
                  </span>
                </td>

                <!-- Size -->
                <td class="py-2.5 px-3 font-mono text-slate-300 text-2xs whitespace-nowrap">
                  {{ formatBytes(b.total_size_bytes) }}
                </td>

                <!-- Validation Status -->
                <td class="py-2.5 px-3 whitespace-nowrap">
                  <span class="px-2 py-0.5 rounded text-3xs font-semibold"
                        :class="getValidationBadge(b.validation_status)">
                    {{ b.validation_status }}
                  </span>
                </td>

                <!-- Consistency Status -->
                <td class="py-2.5 px-3 whitespace-nowrap">
                  <span class="px-2 py-0.5 rounded text-3xs font-semibold"
                        :class="b.consistency_status === 'CONSISTENT' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'">
                    {{ b.consistency_status || 'CONSISTENT' }}
                  </span>
                </td>

                <!-- Restore Compatibility -->
                <td class="py-2.5 px-3 whitespace-nowrap">
                  <span class="px-2 py-0.5 rounded text-3xs font-semibold"
                        :class="b.is_compatible ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-400 border border-rose-500/20'">
                    {{ b.is_compatible ? 'COMPATIBLE' : 'INCOMPATIBLE' }}
                  </span>
                </td>

                <!-- Actions Menu -->
                <td class="py-2.5 px-3 text-right whitespace-nowrap">
                  <div class="flex items-center justify-end space-x-1.5">
                    <!-- View Manifest Details -->
                    <button @click="openDetails(b)"
                            class="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-white transition-colors"
                            title="View Manifest">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"></path>
                      </svg>
                    </button>

                    <!-- Validate Integrity -->
                    <button @click="validateBackup(b.backup_id)" :disabled="validatingId === b.backup_id"
                            class="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-emerald-400 transition-colors"
                            title="Validate Integrity">
                      <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': validatingId === b.backup_id }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                      </svg>
                    </button>

                    <!-- Restore Action -->
                    <button @click="prepareRestore(b.backup_id)"
                            class="px-2 py-1 rounded bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 border border-amber-500/20 text-3xs font-semibold transition-colors flex items-center"
                            title="Restore Database">
                      <svg class="w-3 h-3 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <polyline points="1 4 1 10 7 10"></polyline>
                        <path d="M3.51 15a9 9 0 1 0 2.13-9.36L1 10"></path>
                      </svg>
                      {{ $t('backup.restore') }}
                    </button>

                    <!-- Download Archive -->
                    <button @click="downloadBackup(b.backup_id)"
                            class="p-1 rounded hover:bg-slate-700 text-slate-400 hover:text-blue-400 transition-colors"
                            title="Download .tar.gz">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"></path>
                      </svg>
                    </button>

                    <!-- Delete Backup -->
                    <button @click="deleteBackup(b.backup_id)"
                            class="p-1 rounded hover:bg-rose-950 text-slate-400 hover:text-rose-400 transition-colors"
                            title="Delete Backup">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <polyline points="3 6 5 6 21 6"></polyline>
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Bar -->
        <div class="p-3 border-t border-slate-800 flex items-center justify-between text-2xs text-slate-400 font-sans">
          <span>Showing page {{ page }} of {{ totalPages || 1 }}</span>
          <div class="flex items-center space-x-1">
            <button @click="changePage(page - 1)" :disabled="page <= 1"
                    class="px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition-colors">
              Prev
            </button>
            <button @click="changePage(page + 1)" :disabled="page >= totalPages"
                    class="px-2.5 py-1 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition-colors">
              Next
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 2: Schedule & Retention Policy Form -->
    <div v-if="activeTab === 'schedule'" class="space-y-6 max-w-3xl">
      <div class="saas-card p-5 space-y-5">
        <div class="border-b border-slate-800 pb-3">
          <h2 class="text-sm font-bold text-white font-sans">{{ $t('backup.scheduleEnabled') }}</h2>
          <p class="text-2xs text-slate-400 font-sans mt-0.5">
            Configure automated background snapshots and edge archive retention limits.
          </p>
        </div>

        <!-- Success Alert -->
        <div v-if="scheduleMessage" class="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 text-xs">
          {{ scheduleMessage }}
        </div>

        <div class="space-y-4 text-xs font-sans">
          <!-- Enable Toggle -->
          <div class="flex items-center justify-between p-3 rounded-lg bg-[#0B0F19] border border-slate-800">
            <div>
              <span class="font-semibold text-white block">{{ $t('backup.scheduleEnabled') }}</span>
              <span class="text-3xs text-slate-400 block mt-0.5">Execute backups periodically without manual trigger</span>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" v-model="scheduleForm.enabled" class="sr-only peer">
              <div class="w-10 h-5 bg-slate-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>

          <!-- Interval Hours -->
          <div class="space-y-1">
            <label class="text-slate-300 font-semibold block">{{ $t('backup.scheduleInterval') }}</label>
            <input type="number" v-model.number="scheduleForm.interval_hours" min="1" max="168"
                   class="w-full px-3 py-2 rounded-lg bg-[#0B0F19] border border-slate-800 text-white font-mono focus:border-blue-500 focus:outline-none">
            <span class="text-3xs text-slate-500 block">Default 24 hours (Daily)</span>
          </div>

          <!-- Time of Day UTC -->
          <div class="space-y-1">
            <label class="text-slate-300 font-semibold block">{{ $t('backup.scheduleTime') }}</label>
            <input type="text" v-model="scheduleForm.schedule_time" placeholder="02:00"
                   class="w-full px-3 py-2 rounded-lg bg-[#0B0F19] border border-slate-800 text-white font-mono focus:border-blue-500 focus:outline-none">
            <span class="text-3xs text-slate-500 block">24-hour format HH:MM UTC (e.g. 02:00)</span>
          </div>

          <!-- Max Keep Archives -->
          <div class="space-y-1">
            <label class="text-slate-300 font-semibold block">{{ $t('backup.maxKeep') }}</label>
            <input type="number" v-model.number="scheduleForm.max_keep" min="1" max="50"
                   class="w-full px-3 py-2 rounded-lg bg-[#0B0F19] border border-slate-800 text-white font-mono focus:border-blue-500 focus:outline-none">
            <span class="text-3xs text-slate-500 block">Oldest completed backups will be pruned beyond this threshold</span>
          </div>

          <!-- Min Free Space Threshold -->
          <div class="space-y-1">
            <label class="text-slate-300 font-semibold block">{{ $t('backup.minFreeSpace') }}</label>
            <input type="number" v-model.number="scheduleForm.min_free_space_mb" min="10" max="10000"
                   class="w-full px-3 py-2 rounded-lg bg-[#0B0F19] border border-slate-800 text-white font-mono focus:border-blue-500 focus:outline-none">
            <span class="text-3xs text-slate-500 block">Backups will abort safely if free disk space is lower than this margin</span>
          </div>
        </div>

        <div class="pt-4 border-t border-slate-800 flex justify-end">
          <button @click="saveScheduleConfig" :disabled="savingSchedule"
                  class="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-semibold text-xs transition-colors flex items-center">
            <svg v-if="savingSchedule" class="animate-spin w-3.5 h-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" stroke-width="4" stroke="currentColor" stroke-dasharray="32" stroke-linecap="round"></circle>
            </svg>
            {{ savingSchedule ? $t('backup.savingSchedule') : $t('backup.saveSchedule') }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 1: Create Backup Confirmation -->
    <div v-if="showCreateModal" class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-[#111827] border border-slate-800 rounded-xl p-5 max-w-md w-full space-y-4 shadow-2xl">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-white font-sans flex items-center space-x-2">
            <svg class="w-4 h-4 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 7H5a2 2 0 00-2 2v9a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-3m-1 4l-3 3m0 0l-3-3m3 3V4"></path>
            </svg>
            <span>{{ $t('backup.confirmCreateTitle') }}</span>
          </h3>
          <button @click="showCreateModal = false" class="text-slate-400 hover:text-white">&times;</button>
        </div>

        <p class="text-xs text-slate-300 font-sans leading-relaxed">
          {{ $t('backup.confirmCreateDesc') }}
        </p>

        <div class="space-y-1">
          <label class="text-3xs uppercase tracking-wider text-slate-400 font-sans block">Backup Note / Reason (Optional)</label>
          <input type="text" v-model="createNotes" placeholder="e.g. Pre-maintenance snapshot"
                 class="w-full px-3 py-1.5 rounded-lg bg-[#0B0F19] border border-slate-800 text-white text-xs font-sans focus:border-blue-500 focus:outline-none">
        </div>

        <div class="flex items-center justify-end space-x-2 pt-3 border-t border-slate-800">
          <button @click="showCreateModal = false"
                  class="px-3.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium transition-colors">
            {{ $t('backup.cancel') }}
          </button>
          <button @click="triggerCreateBackup" :disabled="creating"
                  class="px-3.5 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
            <svg v-if="creating" class="animate-spin w-3.5 h-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" stroke-width="4" stroke="currentColor" stroke-dasharray="32" stroke-linecap="round"></circle>
            </svg>
            {{ creating ? $t('backup.creatingBackup') : $t('backup.confirm') }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 2: Manifest Details -->
    <div v-if="showDetailsModal && selectedBackup" class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-[#111827] border border-slate-800 rounded-xl p-5 max-w-2xl w-full space-y-4 shadow-2xl max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-white font-sans flex items-center space-x-2">
            <svg class="w-4 h-4 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
            </svg>
            <span>{{ $t('backup.manifestDetails') }} — {{ selectedBackup.backup_id }}</span>
          </h3>
          <button @click="showDetailsModal = false" class="text-slate-400 hover:text-white">&times;</button>
        </div>

        <!-- Manifest Data Grid -->
        <div v-if="selectedManifest" class="space-y-4 text-xs font-sans">
          <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
            <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800">
              <span class="text-3xs uppercase tracking-wider text-slate-500 block">Backup Version</span>
              <span class="font-mono text-slate-200 text-xs font-bold">{{ selectedManifest.backup_version }}</span>
            </div>
            <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800">
              <span class="text-3xs uppercase tracking-wider text-slate-500 block">App Version</span>
              <span class="font-mono text-slate-200 text-xs font-bold">{{ selectedManifest.app_version }}</span>
            </div>
            <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800">
              <span class="text-3xs uppercase tracking-wider text-slate-500 block">Database Engine</span>
              <span class="font-mono text-emerald-400 text-xs font-bold truncate">{{ selectedManifest.database ? selectedManifest.database.engine : 'MariaDB' }}</span>
            </div>
            <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800">
              <span class="text-3xs uppercase tracking-wider text-slate-500 block">Dump Method</span>
              <span class="font-mono text-blue-400 text-xs font-bold">{{ selectedManifest.database ? selectedManifest.database.dump_method : 'Logical' }}</span>
            </div>
            <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800">
              <span class="text-3xs uppercase tracking-wider text-slate-500 block">Total DB Rows</span>
              <span class="font-mono text-slate-200 text-xs font-bold">{{ selectedManifest.database ? selectedManifest.database.total_rows : 0 }}</span>
            </div>
            <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800">
              <span class="text-3xs uppercase tracking-wider text-slate-500 block">Pending WAL Records</span>
              <span class="font-mono text-purple-400 text-xs font-bold">{{ selectedManifest.wal ? selectedManifest.wal.pending_records : 0 }}</span>
            </div>
          </div>

          <!-- Table Rows Breakdown -->
          <div v-if="selectedManifest.database && selectedManifest.database.table_rows" class="space-y-2">
            <span class="font-semibold text-slate-300 block">{{ $t('backup.allTables') }}</span>
            <div class="grid grid-cols-2 sm:grid-cols-3 gap-1.5 p-2 rounded-lg bg-[#0B0F19] border border-slate-800 text-2xs font-mono">
              <div v-for="(rows, tbl) in selectedManifest.database.table_rows" :key="tbl" class="flex justify-between px-2 py-1 bg-slate-900/50 rounded">
                <span class="text-slate-400">{{ tbl }}:</span>
                <span class="text-blue-300 font-semibold">{{ rows }}</span>
              </div>
            </div>
          </div>

          <!-- Artifacts and Checksums -->
          <div v-if="selectedManifest.database && selectedManifest.database.artifact" class="space-y-2">
            <span class="font-semibold text-slate-300 block">Database Dump Artifact</span>
            <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800 text-2xs font-mono space-y-1">
              <div class="flex justify-between">
                <span class="text-slate-400">File:</span>
                <span class="text-white">{{ selectedManifest.database.artifact.path }} ({{ formatBytes(selectedManifest.database.artifact.size_bytes) }})</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-400">SHA-256:</span>
                <span class="text-emerald-400 truncate max-w-sm" :title="selectedManifest.database.artifact.sha256">{{ selectedManifest.database.artifact.sha256 }}</span>
              </div>
            </div>
          </div>

          <!-- WAL Segments -->
          <div v-if="selectedManifest.wal && selectedManifest.wal.segments && selectedManifest.wal.segments.length" class="space-y-2">
            <span class="font-semibold text-slate-300 block">{{ $t('backup.walSegments') }} ({{ selectedManifest.wal.segments.length }})</span>
            <div class="space-y-1 max-h-32 overflow-y-auto">
              <div v-for="seg in selectedManifest.wal.segments" :key="seg.path"
                   class="p-2 rounded bg-[#0B0F19] border border-slate-800 text-3xs font-mono flex items-center justify-between">
                <span class="text-slate-300">{{ seg.path }} ({{ formatBytes(seg.size_bytes) }})</span>
                <span class="text-slate-500">CRC-32: 0x{{ seg.crc32.toString(16).toUpperCase() }}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="flex justify-end pt-3 border-t border-slate-800">
          <button @click="showDetailsModal = false"
                  class="px-4 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold transition-colors">
            {{ $t('backup.close') }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 3: Restore Preview & Confirmation -->
    <div v-if="showRestoreModal && restorePreview" class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div class="bg-[#111827] border border-rose-500/40 rounded-xl p-5 max-w-lg w-full space-y-4 shadow-2xl">
        <div class="flex items-center justify-between border-b border-slate-800 pb-3">
          <h3 class="text-sm font-bold text-rose-400 font-sans flex items-center space-x-2">
            <svg class="w-4 h-4 text-rose-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
            </svg>
            <span>{{ $t('backup.confirmRestoreTitle') }}</span>
          </h3>
          <button @click="showRestoreModal = false" class="text-slate-400 hover:text-white">&times;</button>
        </div>

        <!-- Warning Alert -->
        <div class="p-3 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-300 text-xs font-sans leading-relaxed">
          {{ $t('backup.confirmRestoreWarning') }}
        </div>

        <!-- Impact Preview Comparison -->
        <div class="space-y-3 text-xs font-sans">
          <div class="grid grid-cols-2 gap-3 p-3 rounded-lg bg-[#0B0F19] border border-slate-800">
            <div>
              <span class="text-3xs uppercase tracking-wider text-slate-500 block">Backup Source</span>
              <span class="font-mono text-white text-xs font-bold">{{ restorePreview.backup_id }}</span>
              <span class="text-3xs text-slate-400 font-mono block mt-0.5">Rows: {{ restorePreview.total_rows }}</span>
            </div>
            <div>
              <span class="text-3xs uppercase tracking-wider text-slate-500 block">Target Active DB</span>
              <span class="font-mono text-emerald-400 text-xs font-bold">{{ restorePreview.target_database || 'datalogger' }}</span>
              <span class="text-3xs text-slate-400 font-mono block mt-0.5">WAL Records: {{ restorePreview.pending_wal_records || 0 }}</span>
            </div>
          </div>

          <!-- Warnings List -->
          <div v-if="restorePreview.warnings && restorePreview.warnings.length" class="space-y-1">
            <span class="text-3xs uppercase tracking-wider text-amber-400 font-semibold block">Important Warnings:</span>
            <ul class="space-y-0.5 text-2xs text-amber-300/90 list-disc list-inside">
              <li v-for="(w, idx) in restorePreview.warnings" :key="idx">{{ w }}</li>
            </ul>
          </div>

          <!-- Safety Backup Checkbox -->
          <label class="flex items-center space-x-2.5 p-2.5 rounded-lg bg-slate-900 border border-slate-800 cursor-pointer">
            <input type="checkbox" v-model="restoreSafetyBackup" class="rounded border-slate-700 text-blue-600 focus:ring-0">
            <span class="text-slate-300 text-2xs font-medium">{{ $t('backup.createSafetyBackup') }}</span>
          </label>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-3 border-t border-slate-800">
          <button @click="showRestoreModal = false"
                  class="px-3.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium transition-colors">
            {{ $t('backup.cancel') }}
          </button>
          <button @click="executeRestore" :disabled="restoring"
                  class="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold shadow-sm transition-colors flex items-center">
            <svg v-if="restoring" class="animate-spin w-3.5 h-3.5 mr-1.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" stroke-width="4" stroke="currentColor" stroke-dasharray="32" stroke-linecap="round"></circle>
            </svg>
            {{ restoring ? $t('backup.restoring') : $t('backup.confirmRestoreButton') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'BackupRestoreView',
  data() {
    return {
      activeTab: 'catalog',
      loading: false,
      creating: false,
      restoring: false,
      savingSchedule: false,
      validatingId: null,
      page: 1,
      pageSize: 15,
      totalCount: 0,
      backupsList: [],
      statusData: {
        subsystem_healthy: true,
        total_backups: 0,
        storage_used_bytes: 0,
        free_disk_bytes: 0,
        last_successful_backup: null,
        next_scheduled_run: null,
      },
      activeJob: null,
      jobPollInterval: null,

      // Modals
      showCreateModal: false,
      createNotes: '',

      showDetailsModal: false,
      selectedBackup: null,
      selectedManifest: null,

      showRestoreModal: false,
      restoreTargetId: null,
      restorePreview: null,
      restoreSafetyBackup: true,

      // Schedule Form
      scheduleForm: {
        enabled: false,
        interval_hours: 24,
        schedule_time: '02:00',
        max_keep: 10,
        min_free_space_mb: 100,
      },
      scheduleMessage: '',
    };
  },
  computed: {
    totalPages() {
      return Math.ceil(this.totalCount / this.pageSize) || 1;
    },
    validCount() {
      return this.backupsList.filter(b => b.validation_status === 'VALID').length;
    },
  },
  mounted() {
    this.fetchBackups();
    this.fetchStatus();
    this.fetchSchedule();
    this.startJobPolling();
  },
  beforeDestroy() {
    if (this.jobPollInterval) {
      clearInterval(this.jobPollInterval);
    }
  },
  methods: {
    formatBytes(bytes) {
      if (!bytes || bytes === 0) return '0 B';
      const k = 1024;
      const dm = 2;
      const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
      const i = Math.floor(Math.log(bytes) / Math.log(k));
      return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
    },
    formatDate(dateStr) {
      if (!dateStr) return '';
      const d = new Date(dateStr);
      return d.toLocaleDateString('en-GB', {
        day: '2-digit',
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    },
    getBackupStatusDot(status) {
      switch (status) {
        case 'COMPLETED':
          return 'bg-emerald-400';
        case 'FAILED':
          return 'bg-rose-400';
        case 'IN_PROGRESS':
          return 'bg-blue-400 animate-pulse';
        default:
          return 'bg-slate-400';
      }
    },
    getBackupTypeBadge(type) {
      switch (type) {
        case 'MANUAL':
          return 'bg-blue-500/10 text-blue-400 border border-blue-500/20';
        case 'SCHEDULED':
          return 'bg-purple-500/10 text-purple-400 border border-purple-500/20';
        case 'SAFETY':
          return 'bg-amber-500/10 text-amber-400 border border-amber-500/20';
        default:
          return 'bg-slate-800 text-slate-300';
      }
    },
    getValidationBadge(status) {
      switch (status) {
        case 'VALID':
          return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
        case 'INVALID':
          return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
        default:
          return 'bg-slate-800 text-slate-400 border border-slate-700';
      }
    },
    async fetchBackups() {
      this.loading = true;
      try {
        const res = await axios.get('/api/backups', {
          params: { page: this.page, page_size: this.pageSize },
        });
        if (res.data && res.data.data) {
          this.backupsList = res.data.data.items || [];
          this.totalCount = res.data.data.total || 0;
        }
      } catch (err) {
        console.warn('Failed fetching backups:', err);
      } finally {
        this.loading = false;
      }
    },
    async fetchStatus() {
      try {
        const res = await axios.get('/api/backups/status');
        if (res.data && res.data.data) {
          this.statusData = res.data.data;
          if (res.data.data.active_job) {
            this.activeJob = res.data.data.active_job;
          }
        }
      } catch (err) {
        console.warn('Failed fetching backup status:', err);
      }
    },
    async fetchSchedule() {
      try {
        const res = await axios.get('/api/backups/schedule');
        if (res.data && res.data.data) {
          const cfg = res.data.data;
          this.scheduleForm = {
            enabled: cfg.enabled || false,
            interval_hours: cfg.interval_hours || 24,
            schedule_time: cfg.schedule_time || '02:00',
            max_keep: cfg.max_keep || 10,
            min_free_space_mb: cfg.min_free_space_mb || 100,
          };
        }
      } catch (err) {
        console.warn('Failed fetching schedule config:', err);
      }
    },
    startJobPolling() {
      this.jobPollInterval = setInterval(async () => {
        if (this.activeJob && (this.activeJob.status === 'RUNNING' || this.activeJob.status === 'PENDING')) {
          try {
            const res = await axios.get(`/api/backups/jobs/${this.activeJob.id}`);
            if (res.data && res.data.data) {
              this.activeJob = res.data.data;
              if (this.activeJob.status === 'COMPLETED' || this.activeJob.status === 'FAILED') {
                this.fetchBackups();
                this.fetchStatus();
              }
            }
          } catch (err) {
            // Job may have finished
            this.fetchBackups();
            this.fetchStatus();
          }
        } else {
          this.fetchStatus();
        }
      }, 3000);
    },
    changePage(newPage) {
      if (newPage >= 1 && newPage <= this.totalPages) {
        this.page = newPage;
        this.fetchBackups();
      }
    },
    async triggerCreateBackup() {
      this.creating = true;
      try {
        const res = await axios.post('/api/backups', {
          type: 'MANUAL',
          description: this.createNotes,
          async: true,
        });
        if (res.data && res.data.data) {
          const jobId = res.data.data.job_id;
          this.activeJob = {
            id: jobId,
            type: 'BACKUP',
            status: 'RUNNING',
            stage: 'INITIALIZING',
            started_at: new Date().toISOString(),
          };
        }
        this.showCreateModal = false;
        this.createNotes = '';
        this.fetchBackups();
      } catch (err) {
        alert('Failed starting backup: ' + (err.response?.data?.error || err.message));
      } finally {
        this.creating = false;
      }
    },
    async openDetails(backup) {
      this.selectedBackup = backup;
      this.selectedManifest = null;
      this.showDetailsModal = true;
      try {
        const res = await axios.get(`/api/backups/${backup.backup_id}`);
        if (res.data && res.data.data) {
          this.selectedManifest = res.data.data.manifest;
        }
      } catch (err) {
        console.warn('Failed loading manifest details:', err);
      }
    },
    async validateBackup(backupId) {
      this.validatingId = backupId;
      try {
        const res = await axios.post(`/api/backups/${backupId}/validate`);
        if (res.data && res.data.success) {
          this.fetchBackups();
        }
      } catch (err) {
        alert('Validation failed: ' + (err.response?.data?.error || err.message));
        this.fetchBackups();
      } finally {
        this.validatingId = null;
      }
    },
    async prepareRestore(backupId) {
      this.restoreTargetId = backupId;
      this.restorePreview = null;
      try {
        const res = await axios.get(`/api/backups/${backupId}/restore-preview`);
        if (res.data && res.data.data) {
          this.restorePreview = res.data.data;
          this.showRestoreModal = true;
        }
      } catch (err) {
        alert('Cannot preview restore: ' + (err.response?.data?.error || err.message));
      }
    },
    async executeRestore() {
      if (!this.restoreTargetId) return;
      this.restoring = true;
      try {
        const res = await axios.post(`/api/backups/${this.restoreTargetId}/restore`, {
          confirm: true,
          create_safety_backup: this.restoreSafetyBackup,
        });
        if (res.data && res.data.data) {
          const jobId = res.data.data.job_id;
          this.activeJob = {
            id: jobId,
            type: 'RESTORE',
            status: 'RUNNING',
            stage: 'PREPARING',
            started_at: new Date().toISOString(),
          };
        }
        this.showRestoreModal = false;
      } catch (err) {
        alert('Restore operation failed: ' + (err.response?.data?.error || err.message));
      } finally {
        this.restoring = false;
      }
    },
    downloadBackup(backupId) {
      const url = `/api/backups/${backupId}/download`;
      window.open(url, '_blank');
    },
    async deleteBackup(backupId) {
      if (!confirm(`Are you sure you want to permanently delete backup archive ${backupId}?`)) {
        return;
      }
      try {
        await axios.delete(`/api/backups/${backupId}`);
        this.fetchBackups();
        this.fetchStatus();
      } catch (err) {
        alert('Failed deleting backup: ' + (err.response?.data?.error || err.message));
      }
    },
    async saveScheduleConfig() {
      this.savingSchedule = true;
      this.scheduleMessage = '';
      try {
        const res = await axios.put('/api/backups/schedule', this.scheduleForm);
        if (res.data && res.data.success) {
          this.scheduleMessage = 'Schedule configuration saved successfully.';
          this.fetchStatus();
        }
      } catch (err) {
        alert('Failed saving schedule: ' + (err.response?.data?.error || err.message));
      } finally {
        this.savingSchedule = false;
      }
    },
  },
};
</script>

<style scoped>
.saas-card {
  background: #111827;
  border: 1px solid rgba(51, 65, 85, 0.4);
  border-radius: 0.75rem;
}
</style>
