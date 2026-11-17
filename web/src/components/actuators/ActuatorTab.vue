<template>
  <div class="space-y-5">
    <!-- Header / Toolbar -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-5 shadow-sm flex flex-col lg:flex-row lg:items-center justify-between gap-4">
      <div>
        <div class="flex flex-wrap items-center gap-2.5">
          <h2 class="text-base font-bold text-slate-800 dark:text-slate-100">Controls & Actuators</h2>
          <span class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-blue-50 text-blue-700 dark:bg-blue-500/10 dark:text-blue-400 border border-blue-200 dark:border-blue-500/30">
            {{ actuators.length }} Configured
          </span>
        </div>
        <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
          Direct control of pumps, purging valves, and relays with automated interval schedules and auto-off watchdogs.
        </p>
      </div>

      <div class="flex flex-wrap items-center gap-2.5 self-start lg:self-auto">
        <!-- View Mode Toggle (Cards / List) -->
        <div
          class="inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700"
          role="group"
          aria-label="Actuator View Mode"
        >
          <button
            type="button"
            @click="setViewMode('grid')"
            class="px-2.5 py-1.5 rounded-md text-xs font-semibold transition-all flex items-center space-x-1.5"
            :class="viewMode === 'grid'
              ? 'bg-white dark:bg-slate-900 text-blue-600 dark:text-blue-400 shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200'"
            title="Mode Kartu (Grid)"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zm10 0a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
            </svg>
            <span>Cards</span>
          </button>

          <button
            type="button"
            @click="setViewMode('list')"
            class="px-2.5 py-1.5 rounded-md text-xs font-semibold transition-all flex items-center space-x-1.5"
            :class="viewMode === 'list'
              ? 'bg-white dark:bg-slate-900 text-blue-600 dark:text-blue-400 shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-800 dark:hover:text-slate-200'"
            title="Mode Daftar (List)"
          >
            <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h16" />
            </svg>
            <span>List</span>
          </button>
        </div>

        <button
          type="button"
          @click="showLogsModal = true"
          class="px-3 py-2 text-xs font-semibold rounded-lg bg-slate-100 hover:bg-slate-200/70 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors flex items-center space-x-1.5"
        >
          <span>📜</span>
          <span>Execution Logs</span>
        </button>

        <!-- Paste Actuator Button -->
        <button
          type="button"
          @click="openPasteModal"
          class="px-3.5 py-2 text-xs font-semibold rounded-lg border transition-all flex items-center space-x-1.5"
          :class="hasClipboardActuator
            ? 'bg-indigo-50 hover:bg-indigo-100 dark:bg-indigo-500/15 dark:hover:bg-indigo-500/25 border-indigo-200 dark:border-indigo-500/40 text-indigo-700 dark:text-indigo-300'
            : 'bg-slate-50 hover:bg-slate-100 dark:bg-slate-800/70 dark:hover:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300'"
          :title="clipboardSummaryTitle"
        >
          <span>📥</span>
          <span>Paste Actuator</span>
          <span
            v-if="clipboardActuatorCode"
            class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-indigo-500/15 dark:bg-indigo-400/20 text-indigo-700 dark:text-indigo-200 max-w-[110px] truncate"
          >
            {{ clipboardActuatorCode }}
          </span>
        </button>

        <button
          type="button"
          @click="openAddModal"
          class="px-4 py-2 text-xs font-bold rounded-lg bg-blue-600 hover:bg-blue-500 active:bg-blue-700 text-white shadow-sm transition-all flex items-center space-x-1.5"
        >
          <span>+</span>
          <span>Add Actuator</span>
        </button>
      </div>
    </div>

    <!-- Copy / Paste Feedback Banner -->
    <div
      v-if="clipboardNotice"
      class="px-4 py-2.5 rounded-xl bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800/60 text-xs text-emerald-800 dark:text-emerald-300 flex items-center justify-between gap-3"
    >
      <div class="flex items-center space-x-2">
        <span>✅</span>
        <span>{{ clipboardNotice }}</span>
      </div>
      <button
        type="button"
        @click="clipboardNotice = ''"
        class="text-emerald-600 dark:text-emerald-400 hover:opacity-75 text-xs font-bold px-1.5"
      >
        ✕
      </button>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="py-16 text-center text-xs text-slate-400">
      Loading actuators...
    </div>

    <!-- Empty State -->
    <div v-else-if="actuators.length === 0" class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-12 text-center text-slate-400 dark:text-slate-500">
      <p class="text-3xl mb-2">⚡</p>
      <h3 class="text-sm font-bold text-slate-700 dark:text-slate-300">No Actuators Configured</h3>
      <p class="text-xs mt-1 max-w-md mx-auto">
        This device does not have any control actuators configured yet. Add an actuator or paste a copied actuator configuration.
      </p>
      <div class="mt-4 flex flex-wrap items-center justify-center gap-2.5">
        <button
          v-if="hasClipboardActuator"
          type="button"
          @click="openPasteModal"
          class="px-4 py-2 text-xs font-bold rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white shadow-sm inline-flex items-center space-x-1.5"
        >
          <span>📥</span>
          <span>Paste Copied Actuator ({{ clipboardActuatorCode }})</span>
        </button>
        <button
          type="button"
          @click="openAddModal"
          class="px-4 py-2 text-xs font-bold rounded-lg bg-blue-600 hover:bg-blue-500 text-white shadow-sm inline-flex items-center space-x-1.5"
        >
          <span>+ Add First Actuator</span>
        </button>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- MODE 1: CARDS / GRID VIEW                  -->
    <!-- ========================================== -->
    <div v-else-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      <div
        v-for="act in actuators"
        :key="act.id"
        class="bg-white dark:bg-slate-900 border rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-all flex flex-col justify-between"
        :class="act.current_state === 'ON' ? 'border-emerald-500/50 dark:border-emerald-500/40 ring-1 ring-emerald-500/20' : 'border-slate-200 dark:border-slate-800'"
      >
        <!-- Card Header -->
        <div class="p-5 border-b border-slate-100 dark:border-slate-800/80 bg-slate-50/30 dark:bg-slate-800/30">
          <div class="flex items-start justify-between gap-2">
            <div class="flex items-center space-x-3 min-w-0">
              <div
                class="w-10 h-10 rounded-xl flex items-center justify-center text-lg shrink-0"
                :class="act.current_state === 'ON' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20' : 'bg-slate-100 dark:bg-slate-800 text-slate-500 border border-slate-200 dark:border-slate-700'"
              >
                {{ getEquipmentIcon(act.icon) }}
              </div>
              <div class="min-w-0">
                <div class="flex items-center space-x-1.5">
                  <h3 class="font-bold text-xs text-slate-800 dark:text-slate-100 truncate">{{ act.actuator_name }}</h3>
                  <span
                    v-if="act.enabled === false"
                    class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/30"
                  >
                    DISABLED
                  </span>
                </div>
                <span class="text-3xs font-mono text-slate-400 dark:text-slate-500 block truncate">{{ act.actuator_code }}</span>
              </div>
            </div>

            <!-- State Badge -->
            <span
              :class="act.current_state === 'ON' ? 'bg-emerald-500 text-white shadow-sm font-extrabold' : 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-slate-700 font-bold'"
              class="px-2.5 py-1 rounded-full text-3xs font-mono tracking-wider flex items-center space-x-1.5 shrink-0"
            >
              <span class="w-1.5 h-1.5 rounded-full" :class="act.current_state === 'ON' ? 'bg-white animate-pulse' : 'bg-slate-400'"></span>
              <span>{{ act.current_state === 'ON' ? 'RUNNING (ON)' : 'STANDBY (OFF)' }}</span>
            </span>
          </div>

          <!-- Description if any -->
          <p v-if="act.description" class="text-3xs text-slate-500 dark:text-slate-400 mt-2.5 line-clamp-2">
            {{ act.description }}
          </p>

          <!-- Auto-Off Countdown Timer (if running and due) -->
          <div v-if="act.current_state === 'ON' && getCountdown(act)" class="mt-3 p-2 rounded-lg bg-emerald-50 dark:bg-emerald-950/30 border border-emerald-200 dark:border-emerald-800 flex items-center justify-between text-3xs font-mono">
            <span class="text-emerald-700 dark:text-emerald-400 flex items-center space-x-1 font-sans font-medium">
              <span>⏱️ Auto-Off Countdown:</span>
            </span>
            <span class="font-bold text-emerald-800 dark:text-emerald-300">
              {{ getCountdown(act) }} remaining
            </span>
          </div>
        </div>

        <!-- Technical Parameters Details -->
        <div class="p-4 space-y-2 text-2xs font-mono text-slate-500 dark:text-slate-400 bg-white dark:bg-slate-900">
          <div class="flex justify-between items-center">
            <span>Modbus Function:</span>
            <span class="font-bold text-slate-700 dark:text-slate-300">{{ act.control_type }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span>Slave & Address:</span>
            <span class="font-bold text-slate-700 dark:text-slate-300">Slave #{{ act.slave_id }} &bull; Offset {{ act.register_address }}</span>
          </div>
          <div class="flex justify-between items-center">
            <span>Watchdog Interlock:</span>
            <span class="text-slate-700 dark:text-slate-300">
              {{ act.safety_watchdog_active ? `${Math.round(act.safety_timeout_seconds / 60)}m Max` : 'Disabled' }}
            </span>
          </div>
          <div class="flex justify-between items-center">
            <span>Automation Routines:</span>
            <span class="text-indigo-600 dark:text-indigo-400 font-bold">
              {{ act.schedules ? act.schedules.length : 0 }} active
            </span>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="p-4 border-t border-slate-100 dark:border-slate-800/80 bg-slate-50/50 dark:bg-slate-800/40 flex flex-col space-y-2.5">
          <!-- Manual Toggle Button -->
          <button
            v-if="act.current_state === 'ON'"
            type="button"
            @click="toggleActuation(act, 'OFF')"
            :disabled="togglingId === act.id"
            class="w-full py-2 px-3 rounded-lg bg-rose-600 hover:bg-rose-500 active:bg-rose-700 text-white font-bold text-xs shadow-sm transition-colors flex items-center justify-center space-x-1.5 disabled:opacity-50"
          >
            <span v-if="togglingId === act.id" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span>Manual STOP</span>
          </button>

          <button
            v-else
            type="button"
            @click="toggleActuation(act, 'ON')"
            :disabled="togglingId === act.id"
            class="w-full py-2 px-3 rounded-lg bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white font-bold text-xs shadow-sm transition-colors flex items-center justify-center space-x-1.5 disabled:opacity-50"
          >
            <span v-if="togglingId === act.id" class="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
            <span>Manual START</span>
          </button>

          <!-- Utilities Row -->
          <div class="grid grid-cols-5 gap-1.5 pt-1">
            <button
              type="button"
              @click="openTestModal(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-center justify-center space-x-1"
              title="Test Write"
            >
              <span>⚡</span>
              <span>Test</span>
            </button>

            <button
              type="button"
              @click="openScheduleModal(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-indigo-600 dark:text-indigo-400 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-center justify-center space-x-1"
              title="Manage Automation Schedules"
            >
              <span>⏱️</span>
              <span>Sched</span>
            </button>

            <button
              type="button"
              @click="copyActuatorConfig(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-cyan-700 dark:text-cyan-400 hover:bg-cyan-50 dark:hover:bg-slate-700 transition-colors flex items-center justify-center space-x-1"
              title="Copy Actuator Configuration"
            >
              <span>📋</span>
              <span>Copy</span>
            </button>

            <button
              type="button"
              @click="openEditModal(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors flex items-center justify-center space-x-1"
              title="Edit Actuator Settings"
            >
              <span>✏️</span>
              <span>Edit</span>
            </button>

            <button
              type="button"
              @click="deleteActuator(act)"
              class="py-1 px-1.5 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors flex items-center justify-center"
              title="Delete Actuator"
            >
              <span>🗑️</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ========================================== -->
    <!-- MODE 2: LIST / TABLE VIEW                  -->
    <!-- ========================================== -->
    <div v-else class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-sm overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="border-b border-slate-200 dark:border-slate-800 bg-slate-50/80 dark:bg-slate-800/50 text-3xs font-mono font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
              <th class="py-3 px-4">Actuator</th>
              <th class="py-3 px-3">State</th>
              <th class="py-3 px-3">Modbus Function & Target</th>
              <th class="py-3 px-3">Payload (ON / OFF)</th>
              <th class="py-3 px-3">Watchdog & Schedule</th>
              <th class="py-3 px-4 text-right">Controls & Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-800/80 text-xs">
            <tr
              v-for="act in actuators"
              :key="act.id"
              class="hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition-colors"
              :class="act.current_state === 'ON' ? 'bg-emerald-50/30 dark:bg-emerald-950/10' : ''"
            >
              <!-- Actuator Identity -->
              <td class="py-3 px-4">
                <div class="flex items-center space-x-3">
                  <div
                    class="w-8 h-8 rounded-lg flex items-center justify-center text-sm shrink-0"
                    :class="act.current_state === 'ON' ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30' : 'bg-slate-100 dark:bg-slate-800 text-slate-500 border border-slate-200 dark:border-slate-700'"
                  >
                    {{ getEquipmentIcon(act.icon) }}
                  </div>
                  <div class="min-w-0">
                    <div class="flex items-center space-x-1.5">
                      <span class="font-bold text-slate-800 dark:text-slate-100 truncate">{{ act.actuator_name }}</span>
                      <span
                        v-if="act.enabled === false"
                        class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/30"
                      >
                        DISABLED
                      </span>
                    </div>
                    <div class="text-3xs font-mono text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                      <span>{{ act.actuator_code }}</span>
                      <span v-if="act.description" class="text-slate-400 dark:text-slate-500 truncate max-w-[180px]" :title="act.description">
                        &bull; {{ act.description }}
                      </span>
                    </div>
                  </div>
                </div>
              </td>

              <!-- State -->
              <td class="py-3 px-3 whitespace-nowrap">
                <div class="flex flex-col items-start space-y-1">
                  <span
                    :class="act.current_state === 'ON' ? 'bg-emerald-500 text-white shadow-xs font-extrabold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700 font-bold'"
                    class="px-2 py-0.5 rounded-full text-3xs font-mono tracking-wider inline-flex items-center space-x-1.5"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="act.current_state === 'ON' ? 'bg-white animate-pulse' : 'bg-slate-400'"></span>
                    <span>{{ act.current_state === 'ON' ? 'RUNNING (ON)' : 'STANDBY (OFF)' }}</span>
                  </span>
                  <span
                    v-if="act.current_state === 'ON' && getCountdown(act)"
                    class="text-4xs font-mono font-bold text-emerald-700 dark:text-emerald-400"
                  >
                    ⏱️ {{ getCountdown(act) }} left
                  </span>
                </div>
              </td>

              <!-- Modbus Function & Target -->
              <td class="py-3 px-3 font-mono whitespace-nowrap">
                <div class="text-2xs font-bold text-slate-700 dark:text-slate-200">{{ act.control_type }}</div>
                <div class="text-3xs text-slate-500 dark:text-slate-400">Slave #{{ act.slave_id }} &bull; Offset {{ act.register_address }}</div>
              </td>

              <!-- Payload ON/OFF -->
              <td class="py-3 px-3 font-mono whitespace-nowrap text-3xs">
                <span class="px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border border-emerald-500/25 font-bold mr-1">
                  ON: {{ act.on_payload }}
                </span>
                <span class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700 font-bold">
                  OFF: {{ act.off_payload }}
                </span>
              </td>

              <!-- Watchdog & Schedules -->
              <td class="py-3 px-3 font-mono whitespace-nowrap text-3xs">
                <div class="text-slate-700 dark:text-slate-300">
                  🛡️ {{ act.safety_watchdog_active ? `${Math.round(act.safety_timeout_seconds / 60)}m Max` : 'Disabled' }}
                </div>
                <button
                  type="button"
                  @click="openScheduleModal(act)"
                  class="text-indigo-600 dark:text-indigo-400 hover:underline font-bold mt-0.5 inline-flex items-center space-x-1"
                >
                  <span>⏱️ {{ act.schedules ? act.schedules.length : 0 }} schedule(s)</span>
                </button>
              </td>

              <!-- Actions -->
              <td class="py-3 px-4 whitespace-nowrap text-right">
                <div class="inline-flex items-center justify-end gap-1.5">
                  <button
                    v-if="act.current_state === 'ON'"
                    type="button"
                    @click="toggleActuation(act, 'OFF')"
                    :disabled="togglingId === act.id"
                    class="py-1 px-2.5 rounded-md bg-rose-600 hover:bg-rose-500 text-white font-bold text-3xs shadow-xs transition-colors inline-flex items-center space-x-1 disabled:opacity-50"
                  >
                    <span v-if="togglingId === act.id" class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                    <span>STOP</span>
                  </button>
                  <button
                    v-else
                    type="button"
                    @click="toggleActuation(act, 'ON')"
                    :disabled="togglingId === act.id"
                    class="py-1 px-2.5 rounded-md bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-3xs shadow-xs transition-colors inline-flex items-center space-x-1 disabled:opacity-50"
                  >
                    <span v-if="togglingId === act.id" class="w-3 h-3 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                    <span>START</span>
                  </button>

                  <button
                    type="button"
                    @click="openTestModal(act)"
                    class="py-1 px-2 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                    title="Test Write"
                  >
                    ⚡ Test
                  </button>

                  <button
                    type="button"
                    @click="openScheduleModal(act)"
                    class="py-1 px-2 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-indigo-600 dark:text-indigo-400 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                    title="Manage Automation Schedules"
                  >
                    ⏱️ Sched
                  </button>

                  <button
                    type="button"
                    @click="copyActuatorConfig(act)"
                    class="py-1 px-2 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-cyan-700 dark:text-cyan-400 hover:bg-cyan-50 dark:hover:bg-slate-700 transition-colors"
                    title="Copy Actuator Configuration"
                  >
                    📋 Copy
                  </button>

                  <button
                    type="button"
                    @click="openEditModal(act)"
                    class="py-1 px-2 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-700 transition-colors"
                    title="Edit Actuator Settings"
                  >
                    ✏️ Edit
                  </button>

                  <button
                    type="button"
                    @click="deleteActuator(act)"
                    class="py-1 px-2 text-3xs font-semibold rounded-md bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-950/30 transition-colors"
                    title="Delete Actuator"
                  >
                    🗑️
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Modals -->
    <ActuatorModal
      :isOpen="showModal"
      :actuator="selectedActuator"
      :draftActuator="draftActuator"
      :deviceId="deviceId"
      :deviceCode="deviceCode"
      :deviceSlaveId="deviceSlaveId"
      :parameters="parameters"
      @close="closeActuatorModal"
      @copied="onModalCopied"
      @saved="fetchActuators"
    />

    <ActuatorTestWriteModal
      :isOpen="showTestModal"
      :actuator="selectedActuator"
      :deviceId="deviceId"
      @close="showTestModal = false"
      @tested="fetchActuators"
    />

    <ActuatorScheduleModal
      :isOpen="showScheduleModal"
      :actuator="selectedActuator"
      @close="showScheduleModal = false"
      @updated="fetchActuators"
    />

    <ActuatorLogsModal
      :isOpen="showLogsModal"
      :deviceId="deviceId"
      @close="showLogsModal = false"
    />
  </div>
</template>

<script>
import axios from 'axios';
import ActuatorModal from './ActuatorModal.vue';
import ActuatorTestWriteModal from './ActuatorTestWriteModal.vue';
import ActuatorScheduleModal from './ActuatorScheduleModal.vue';
import ActuatorLogsModal from './ActuatorLogsModal.vue';

const ACTUATOR_VIEW_MODE_KEY = 'datalogger_actuator_view_mode';
const ACTUATOR_CLIPBOARD_KEY = 'datalogger_actuator_clipboard_v1';

export default {
  name: 'ActuatorTab',
  components: {
    ActuatorModal,
    ActuatorTestWriteModal,
    ActuatorScheduleModal,
    ActuatorLogsModal
  },
  props: {
    deviceId: { type: [Number, String], required: true },
    deviceCode: { type: String, default: '' },
    deviceSlaveId: { type: Number, default: 1 },
    parameters: { type: Array, default: () => [] }
  },
  data() {
    let initialViewMode = 'grid';
    try {
      const savedMode = localStorage.getItem(ACTUATOR_VIEW_MODE_KEY);
      if (savedMode === 'list' || savedMode === 'grid') {
        initialViewMode = savedMode;
      }
    } catch (e) { /* ignore */ }

    return {
      actuators: [],
      loading: false,
      togglingId: null,
      viewMode: initialViewMode,
      clipboardPacket: null,
      clipboardNotice: '',
      showModal: false,
      showTestModal: false,
      showScheduleModal: false,
      showLogsModal: false,
      selectedActuator: null,
      draftActuator: null,
      timerInterval: null,
      now: Date.now()
    };
  },
  computed: {
    hasClipboardActuator() {
      return !!(this.clipboardPacket && this.clipboardPacket.config);
    },
    clipboardActuatorCode() {
      if (!this.hasClipboardActuator) return '';
      return this.clipboardPacket.config.actuator_code || this.clipboardPacket.config.actuator_name || '';
    },
    clipboardSummaryTitle() {
      if (!this.hasClipboardActuator) {
        return 'Paste konfigurasi actuator yang sudah di-copy dari perangkat ini atau perangkat lain';
      }
      const cfg = this.clipboardPacket.config;
      const src = this.clipboardPacket.source_device_code ? ` dari ${this.clipboardPacket.source_device_code}` : '';
      return `Paste '${cfg.actuator_name || cfg.actuator_code}'${src}`;
    }
  },
  watch: {
    deviceId() {
      this.syncClipboardState();
      this.fetchActuators();
    }
  },
  mounted() {
    this.syncClipboardState();
    this.fetchActuators();
    this.timerInterval = setInterval(() => {
      this.now = Date.now();
    }, 1000);
    window.addEventListener('actuator-state-event', this.onActuatorStateEvent);
    window.addEventListener('storage', this.syncClipboardState);
  },
  beforeDestroy() {
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
    }
    window.removeEventListener('actuator-state-event', this.onActuatorStateEvent);
    window.removeEventListener('storage', this.syncClipboardState);
  },
  methods: {
    setViewMode(mode) {
      this.viewMode = mode === 'list' ? 'list' : 'grid';
      try {
        localStorage.setItem(ACTUATOR_VIEW_MODE_KEY, this.viewMode);
      } catch (e) { /* ignore */ }
    },
    syncClipboardState() {
      try {
        const raw = localStorage.getItem(ACTUATOR_CLIPBOARD_KEY);
        if (!raw) {
          this.clipboardPacket = null;
          return;
        }
        const parsed = JSON.parse(raw);
        if (parsed && parsed._type === 'datalogger_actuator_config_v1' && parsed.config) {
          this.clipboardPacket = parsed;
        } else {
          this.clipboardPacket = null;
        }
      } catch (e) {
        this.clipboardPacket = null;
      }
    },
    generateUniqueCode(baseCode) {
      const clean = String(baseCode || 'ACTUATOR_01').trim();
      const existingCodes = new Set(
        (this.actuators || []).map(a => String(a.actuator_code || '').trim().toLowerCase())
      );
      if (!existingCodes.has(clean.toLowerCase())) {
        return clean;
      }
      const match = clean.match(/^(.*?)(\d+)$/);
      if (match) {
        const prefix = match[1];
        let num = parseInt(match[2], 10) + 1;
        const width = match[2].length;
        while (num < 10000) {
          const candidate = `${prefix}${String(num).padStart(width, '0')}`;
          if (!existingCodes.has(candidate.toLowerCase())) return candidate;
          num++;
        }
      }
      let idx = 2;
      while (idx < 1000) {
        const candidate = `${clean}_${idx}`;
        if (!existingCodes.has(candidate.toLowerCase())) return candidate;
        idx++;
      }
      return `${clean}_copy`;
    },
    generateUniqueName(baseName, codeChanged) {
      const clean = String(baseName || 'Actuator').trim();
      if (!codeChanged) return clean;
      const match = clean.match(/^(.*?)(\d+)$/);
      if (match) {
        return `${match[1]}${parseInt(match[2], 10) + 1}`;
      }
      return `${clean} (Copy)`;
    },
    copyActuatorConfig(act) {
      if (!act) return;
      const matchedParam = Array.isArray(this.parameters)
        ? this.parameters.find(p => Number(p.id) === Number(act.feedback_parameter_id))
        : null;
      const schedules = Array.isArray(act.schedules)
        ? act.schedules.map(s => ({
            schedule_name: s.schedule_name || '',
            schedule_type: s.schedule_type || 'INTERVAL',
            interval_seconds: Number(s.interval_seconds) || 7200,
            duration_seconds: Number(s.duration_seconds) || 300,
            minute_of_hour: s.minute_of_hour || '',
            start_time_str: s.start_time_str || '',
            end_time_str: s.end_time_str || '',
            cron_expression: s.cron_expression || '',
            days_of_week: s.days_of_week || '1,2,3,4,5,6,7',
            enabled: s.enabled !== false
          }))
        : [];

      const packet = {
        _type: 'datalogger_actuator_config_v1',
        copied_at: new Date().toISOString(),
        source_device_id: Number(this.deviceId) || 0,
        source_device_code: this.deviceCode || '',
        config: {
          actuator_code: act.actuator_code || '',
          actuator_name: act.actuator_name || '',
          description: act.description || '',
          icon: act.icon || 'pump',
          control_type: act.control_type || 'COIL_FC05',
          slave_id: Number(act.slave_id) || 1,
          register_address: Number(act.register_address) || 0,
          on_payload: act.on_payload !== undefined ? Number(act.on_payload) : 65280,
          off_payload: act.off_payload !== undefined ? Number(act.off_payload) : 0,
          safety_timeout_seconds: Number(act.safety_timeout_seconds) || 900,
          safety_watchdog_active: act.safety_watchdog_active !== false,
          feedback_parameter_id: act.feedback_parameter_id ? Number(act.feedback_parameter_id) : null,
          feedback_parameter_code: matchedParam ? (matchedParam.parameter_code || matchedParam.code || '') : '',
          enabled: act.enabled !== false
        },
        schedules
      };

      const jsonStr = JSON.stringify(packet, null, 2);
      try {
        localStorage.setItem(ACTUATOR_CLIPBOARD_KEY, jsonStr);
      } catch (e) { /* ignore */ }
      if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
        navigator.clipboard.writeText(jsonStr).catch(() => {});
      }
      this.clipboardPacket = packet;
      const schedInfo = schedules.length > 0 ? ` beserta ${schedules.length} jadwal` : '';
      this.clipboardNotice = `Konfigurasi "${act.actuator_name}" (${act.actuator_code})${schedInfo} berhasil di-copy. Klik "Paste Actuator" di perangkat ini atau perangkat lain.`;
    },
    onModalCopied(packet) {
      this.clipboardPacket = packet;
    },
    async openPasteModal() {
      let packet = this.clipboardPacket;
      if (navigator.clipboard && typeof navigator.clipboard.readText === 'function') {
        try {
          const clipText = await navigator.clipboard.readText();
          if (clipText && clipText.includes('datalogger_actuator_config_v1')) {
            const parsed = JSON.parse(clipText);
            if (parsed && parsed.config) {
              packet = parsed;
              this.clipboardPacket = parsed;
              try {
                localStorage.setItem(ACTUATOR_CLIPBOARD_KEY, clipText);
              } catch (e) { /* ignore */ }
            }
          }
        } catch (e) { /* fallback to localStorage */ }
      }
      if (!packet) {
        this.syncClipboardState();
        packet = this.clipboardPacket;
      }
      if (!packet || !packet.config) {
        alert('Belum ada konfigurasi actuator yang di-copy. Klik tombol "Copy" pada salah satu actuator terlebih dahulu.');
        return;
      }

      const cfg = packet.config;
      const uniqueCode = this.generateUniqueCode(cfg.actuator_code);
      const codeChanged = uniqueCode.toLowerCase() !== String(cfg.actuator_code || '').trim().toLowerCase();
      const uniqueName = this.generateUniqueName(cfg.actuator_name, codeChanged);

      this.selectedActuator = null;
      this.draftActuator = {
        ...cfg,
        actuator_code: uniqueCode,
        actuator_name: uniqueName,
        _paste_source: packet.source_device_code ? `DEV: ${packet.source_device_code}` : (cfg.actuator_code || ''),
        _copied_schedules: Array.isArray(packet.schedules) ? packet.schedules : []
      };
      this.showModal = true;
    },
    closeActuatorModal() {
      this.showModal = false;
      this.draftActuator = null;
    },
    async fetchActuators() {
      this.loading = true;
      try {
        const res = await axios.get(`/api/devices/${this.deviceId}/actuators`);
        this.actuators = res.data.data || [];
      } catch (err) {
        console.error('Failed to fetch device actuators:', err);
      } finally {
        this.loading = false;
      }
    },
    onActuatorStateEvent(e) {
      const data = e.detail;
      if (!data || Number(data.device_id) !== Number(this.deviceId)) return;
      const idx = this.actuators.findIndex(a => a.id === data.actuator_id);
      if (idx !== -1) {
        this.$set(this.actuators[idx], 'current_state', data.state);
        this.$set(this.actuators[idx], 'auto_off_due_at', data.auto_off_due_at || null);
      } else {
        this.fetchActuators();
      }
    },
    getEquipmentIcon(type) {
      const map = {
        pump: '💧',
        valve: '💨',
        fan: '🌀',
        heater: '🔥',
        relay: '⚡',
        light: '💡',
        motor: '⚙️'
      };
      return map[type] || '⚡';
    },
    getCountdown(act) {
      if (!act.auto_off_due_at) return null;
      const due = new Date(act.auto_off_due_at).getTime();
      const diff = Math.max(0, Math.floor((due - this.now) / 1000));
      if (diff <= 0) return '00:00';
      const m = Math.floor(diff / 60).toString().padStart(2, '0');
      const s = (diff % 60).toString().padStart(2, '0');
      return `${m}:${s}`;
    },
    openAddModal() {
      this.selectedActuator = null;
      this.draftActuator = null;
      this.showModal = true;
    },
    openEditModal(act) {
      this.selectedActuator = act;
      this.draftActuator = null;
      this.showModal = true;
    },
    openTestModal(act) {
      this.selectedActuator = act;
      this.showTestModal = true;
    },
    openScheduleModal(act) {
      this.selectedActuator = act;
      this.showScheduleModal = true;
    },
    async toggleActuation(act, action) {
      this.togglingId = act.id;
      try {
        await axios.post(`/api/actuators/${act.id}/actuate`, {
          action: action
        });
        await this.fetchActuators();
      } catch (err) {
        const errData = err.response && err.response.data;
        alert(`Failed to actuate ${act.actuator_name}: ` + ((errData && (errData.error || errData.message)) || err.message));
      } finally {
        this.togglingId = null;
      }
    },
    async deleteActuator(act) {
      if (!confirm(`Are you sure you want to delete actuator "${act.actuator_name}"? All associated schedules and logs will also be removed.`)) return;
      try {
        await axios.delete(`/api/actuators/${act.id}`);
        await this.fetchActuators();
      } catch (err) {
        const errData = err.response && err.response.data;
        alert('Failed to delete actuator: ' + ((errData && (errData.error || errData.message)) || err.message));
      }
    }
  }
};
</script>
