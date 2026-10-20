<template>
  <div class="operator-command-center operator-dashboard font-sans text-slate-800 dark:text-slate-100 select-none pb-[33px] transition-colors duration-150 -mx-1 sm:-mx-1.5 -mt-0.5 sm:-mt-1 px-1 sm:px-1.5 pt-0.5 sm:pt-1 min-h-[calc(100vh-32px)]" :class="isDark ? 'dark-mode' : 'light-mode'">
    <!-- ======================================================== -->
    <!-- UNIFIED OPERATOR DASHBOARD HEADER (SINGLE OUTER CARD)    -->
    <!-- Wraps:                                                   -->
    <!-- 1. Operator Command Center Header                        -->
    <!-- 2. Device & Comm Status Card                             -->
    <!-- 3. Active Alarms Card                                    -->
    <!-- 4. Latest Telemetry Card                                 -->
    <!-- 5. System Monitor Card                                   -->
    <!-- 6. Device & Parameter Monitoring Toolbar & Categories    -->
    <!-- ======================================================== -->
    <div class="scada-unified-header-card rounded-lg border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#0B1422] p-2 sm:p-2.5 mb-1.5 shadow-xs transition-colors flex flex-col space-y-1.5 select-none relative z-30">
      <!-- 1. Row 1 — Main Header: Customer Branding (Left), Centered Clock (Center), Maker Branding (Right) -->
      <div class="relative flex items-center justify-between gap-1.5 leading-none w-full min-h-[34px]">
        <!-- Area 1 (Left): Customer Company Branding (Icon/Logo, Name, Subtitle & Station Info) -->
        <div class="flex items-center space-x-2 z-0 min-w-0 pr-2">
          <!-- Customer Logo / Icon Button (Clickable to configure Customer Branding) -->
          <button
            type="button"
            @click="openBrandingModal('customer')"
            class="h-7 px-1.5 rounded transition flex items-center justify-center border flex-shrink-0"
            :class="customerLogoUrl
              ? 'border-transparent hover:border-cyan-500/40 bg-slate-50/50 dark:bg-[#101B2B]/60'
              : 'border border-cyan-500/30 bg-gradient-to-br from-cyan-500/10 to-blue-600/10 text-cyan-500 hover:border-cyan-400 dark:hover:border-cyan-400 shadow-glow-blue/20'"
            :title="(customerCompanyName || 'Customer Organization') + ' · Click to configure'"
            aria-label="Configure Customer Company Logo"
          >
            <img
              v-if="customerLogoUrl"
              :src="customerLogoUrl"
              :alt="customerCompanyName || 'Customer Logo'"
              class="h-5 sm:h-6 max-w-[110px] w-auto object-contain rounded-xs"
            />
            <div v-else class="w-5 h-5 rounded bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center text-white shadow-glow-blue flex-shrink-0">
              <OperatorDashboardIcon name="station" :size="13" class="text-white" />
            </div>
          </button>

          <!-- Customer Company Name & Subtitle / Station Details -->
          <div class="min-w-0 truncate">
            <div class="flex items-center space-x-1.5">
              <h1 class="text-[11px] font-black tracking-wider uppercase text-slate-900 dark:text-white font-mono flex items-center gap-1 leading-none truncate">
                <span class="truncate" :title="customerCompanyName">{{ customerCompanyName || 'PT. Mitra Industri Enviro' }}</span>
                <span class="px-1 py-0.2 rounded text-[8px] font-mono font-bold bg-cyan-500/10 dark:bg-cyan-500/20 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30 flex-shrink-0">
                  SCADA
                </span>
              </h1>
            </div>
            <div v-if="stationTypeLabel" class="text-[9px] text-slate-500 dark:text-slate-400 font-mono flex items-center space-x-1.5 leading-none mt-0.5 truncate">
              <span>{{ $t('operator.station') || 'Station' }}: <strong class="text-cyan-600 dark:text-cyan-400 font-bold tracking-wide">{{ stationTypeLabel }}</strong></span>
            </div>
          </div>
        </div>

        <!-- Area 2 (Center): Dead-Centered Digital Clock -->
        <div v-if="currentTime" class="absolute left-1/2 -translate-x-1/2 top-1/2 -translate-y-1/2 z-10 pointer-events-auto">
          <div class="h-7 flex items-center space-x-1.5 px-3 rounded bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-200 font-bold text-[10px] tabular-nums font-mono leading-none shadow-xs">
            <span class="text-slate-400 dark:text-slate-500 text-3xs">🕒</span>
            <span>{{ currentTime }}</span>
          </div>
        </div>

        <!-- Area 3 (Right): Maker Company Branding (Name, Subtitle, Icon/Logo Button) -->
        <div class="flex items-center space-x-2 z-0 flex-shrink-0 pl-2">
          <!-- Maker Info (Right-aligned text) -->
          <div v-if="makerCompanyName || makerSubtitle" class="hidden sm:flex flex-col items-end text-right min-w-0">
            <span v-if="makerCompanyName" class="text-[10px] font-bold text-slate-800 dark:text-slate-200 font-mono tracking-tight leading-none truncate max-w-[170px]" :title="makerCompanyName">
              {{ makerCompanyName }}
            </span>
            <span v-if="makerSubtitle" class="text-[8.5px] text-slate-400 dark:text-slate-500 font-mono tracking-wider leading-none mt-0.5 truncate max-w-[170px]">
              {{ makerSubtitle }}
            </span>
          </div>

          <!-- Maker Logo / Icon Button -->
          <button
            type="button"
            @click="openBrandingModal('maker')"
            class="h-7 px-1.5 rounded transition flex items-center justify-center border flex-shrink-0"
            :class="makerLogoUrl
              ? 'border-transparent hover:border-cyan-500/40 bg-slate-50/50 dark:bg-[#101B2B]/60'
              : 'border border-dashed border-slate-300 dark:border-[#20354B] hover:border-cyan-500 text-slate-500 dark:text-slate-400 hover:text-cyan-500 dark:hover:text-cyan-400 bg-slate-50 dark:bg-[#101B2B]'"
            :title="(makerCompanyName || 'Maker / Solution Provider') + ' · Click to configure'"
            aria-label="Configure Maker Company Logo"
          >
            <img
              v-if="makerLogoUrl"
              :src="makerLogoUrl"
              :alt="makerCompanyName || 'Maker Logo'"
              class="h-5 sm:h-6 max-w-[110px] w-auto object-contain rounded-xs"
            />
            <div v-else class="flex items-center space-x-1 text-3xs font-mono font-bold leading-none px-1">
              <OperatorDashboardIcon name="activity" :size="14" class="text-cyan-500" />
              <span v-if="makerCompanyName" class="hidden md:inline text-[9px]">{{ makerCompanyName }}</span>
            </div>
          </button>
        </div>
      </div>

      <!-- Thin Divider between Main Header and Status/Setup Toolbar -->
      <div class="border-t border-slate-200/70 dark:border-[#1E3048]"></div>

      <!-- 2. Row 2 — Status and Setup Toolbar -->
      <div class="flex items-center justify-between gap-2 leading-none w-full">
        <!-- Kelompok Kiri — Status Icons (Device & Comm, Alarms, Telemetry, System) -->
        <div class="flex items-center space-x-1.5 min-w-0">
          <div class="flex items-center space-x-1 flex-shrink-0">
            <!-- 1. Device & Comm Status Popover -->
            <div class="relative toolbar-popover-anchor">
              <button
                type="button"
                @click.stop="toggleToolbarPopover('deviceStatus')"
                class="w-8 h-8 flex items-center justify-center rounded transition border relative"
                :class="[
                  activeToolbarPopover === 'deviceStatus'
                    ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border-cyan-500 shadow-xs'
                    : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'
                ]"
                :title="($t('operator.toolbarDeviceStatusTooltip') || $t('operator.summaryDevices') || 'Device & Comm Status') + ': ' + summaryStats.onlineCount + '/' + summaryStats.totalDevices + ' Online (' + summaryStats.offlineCount + ' Off)'"
                aria-label="Device & Comm Status"
                :aria-expanded="activeToolbarPopover === 'deviceStatus'"
              >
                <OperatorDashboardIcon name="station" :size="16" />
                <span
                  class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full ring-1 ring-white dark:ring-[#0B1422] transition-colors"
                  :class="summaryStats.offlineCount > 0 ? 'bg-rose-500' : 'bg-emerald-400'"
                ></span>
              </button>

              <!-- Device & Comm Status Popover Panel -->
              <div
                v-if="activeToolbarPopover === 'deviceStatus'"
                class="absolute left-0 top-full mt-1.5 z-50 w-80 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono"
                @click.stop
              >
                <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
                  <div class="flex items-center space-x-2">
                    <div class="p-1 rounded bg-cyan-500/10 border border-cyan-500/30 text-cyan-500">
                      <OperatorDashboardIcon name="station" :size="15" class="text-cyan-500" />
                    </div>
                    <div>
                      <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                        {{ $t('operator.summaryDevices') || 'Device & Comm Status' }}
                      </h3>
                      <div class="text-[9px] text-slate-500 dark:text-slate-400 mt-0.5 leading-none">
                        Fleet Communication Diagnostics
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1.5">
                    <span
                      class="text-[9px] font-bold px-1.5 py-0.5 rounded leading-none"
                      :class="summaryStats.offlineCount > 0 ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/30' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30'"
                    >
                      {{ summaryStats.offlineCount > 0 ? summaryStats.offlineCount + ' OFF' : 'ALL ONLINE' }}
                    </span>
                    <button
                      type="button"
                      @click="closeToolbarPopover"
                      class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Main Stats Highlights -->
                <div class="p-2.5 rounded-md bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] mb-2.5">
                  <div class="flex items-baseline justify-between mb-1.5">
                    <div>
                      <span class="text-[20px] font-bold text-slate-900 dark:text-white tabular-nums tracking-tight">
                        {{ summaryStats.onlineCount }} / {{ summaryStats.totalDevices }}
                      </span>
                      <span class="text-[10px] text-slate-500 dark:text-slate-400 ml-1">devices online</span>
                    </div>
                    <span class="text-[11px] font-bold" :class="summaryStats.offlineCount > 0 ? 'text-rose-600 dark:text-rose-400' : 'text-emerald-600 dark:text-emerald-400'">
                      {{ Math.round((summaryStats.onlineCount / (summaryStats.totalDevices || 1)) * 100) }}% active
                    </span>
                  </div>
                  <!-- Progress Bar -->
                  <div class="w-full h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all duration-300"
                      :class="summaryStats.offlineCount > 0 ? 'bg-amber-500' : 'bg-emerald-500'"
                      :style="{ width: Math.round((summaryStats.onlineCount / (summaryStats.totalDevices || 1)) * 100) + '%' }"
                    ></div>
                  </div>
                  <div class="flex items-center justify-between text-[9px] text-slate-500 dark:text-slate-400 mt-2 pt-1.5 border-t border-slate-200/50 dark:border-slate-800/60">
                    <span>Offline Devices: <strong :class="summaryStats.offlineCount > 0 ? 'text-rose-600 dark:text-rose-400 font-bold' : 'text-slate-700 dark:text-slate-300'">{{ summaryStats.offlineCount }}</strong></span>
                    <span>Latency: <strong class="text-cyan-600 dark:text-cyan-400">{{ systemStatus.avg_latency_ms || 1.2 }}ms</strong></span>
                    <span>Success: <strong class="text-emerald-600 dark:text-emerald-400">{{ ((systemStatus.success_rate || 99.8).toFixed(1)) }}%</strong></span>
                  </div>
                </div>

                <!-- Device Fleet List -->
                <div class="mb-2.5">
                  <div class="text-[9px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 flex items-center justify-between">
                    <span>Connected Fleet ({{ devicesList.length }})</span>
                    <span>Status</span>
                  </div>
                  <div class="max-h-36 overflow-y-auto custom-scroller space-y-1 pr-0.5">
                    <div
                      v-for="dev in devicesList"
                      :key="'popover-dev-' + dev.device_id"
                      class="flex items-center justify-between p-1.5 rounded bg-slate-50/70 dark:bg-[#0E1726] border border-slate-200/70 dark:border-[#1E3048] text-[9.5px]"
                    >
                      <div class="flex items-center space-x-1.5 truncate min-w-0 pr-1">
                        <span
                          class="w-1.5 h-1.5 rounded-full flex-shrink-0"
                          :class="isDeviceActiveOnline(dev) ? 'bg-emerald-500' : 'bg-rose-500'"
                        ></span>
                        <span class="font-bold text-slate-800 dark:text-slate-200 truncate">{{ dev.device_code }}</span>
                        <span class="text-slate-400 truncate hidden sm:inline">({{ dev.device_name || 'Generic' }})</span>
                      </div>
                      <span
                        class="px-1 py-0.2 rounded text-[8px] font-bold flex-shrink-0"
                        :class="isDeviceActiveOnline(dev) ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400' : 'bg-rose-500/10 text-rose-600 dark:text-rose-400'"
                      >
                        {{ isDeviceActiveOnline(dev) ? 'ONLINE' : 'OFFLINE' }}
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Action Link -->
                <button
                  type="button"
                  @click="navigateToRoute('/monitoring/devices')"
                  class="w-full py-1.5 px-2 rounded text-3xs font-semibold transition border flex items-center justify-center space-x-1.5 bg-slate-50 dark:bg-[#0E1726] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400"
                >
                  <span>⚙️</span>
                  <span>Configure Device Fleet & Telemetry →</span>
                </button>
              </div>
            </div>

            <!-- 2. Active Alarms Popover -->
            <div class="relative toolbar-popover-anchor">
              <button
                type="button"
                @click.stop="toggleToolbarPopover('activeAlarms')"
                class="w-8 h-8 flex items-center justify-center rounded transition border relative"
                :class="[
                  activeToolbarPopover === 'activeAlarms'
                    ? 'bg-rose-500/20 text-rose-600 dark:text-rose-300 border-rose-500 shadow-xs'
                    : activeCriticalCount > 0
                      ? 'bg-rose-500/15 text-rose-600 dark:text-rose-300 border-rose-500/50'
                      : activeAlarmsList.length > 0
                        ? 'bg-amber-500/15 text-amber-600 dark:text-amber-300 border-amber-500/50'
                        : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'
                ]"
                :title="($t('operator.toolbarActiveAlarmsTooltip') || $t('operator.summaryAlarms') || 'Active Alarms') + ': ' + activeAlarmsList.length + ' (' + activeCriticalCount + ' CRIT, ' + activeWarningCount + ' WARN)'"
                aria-label="Active Alarms"
                :aria-expanded="activeToolbarPopover === 'activeAlarms'"
              >
                <OperatorDashboardIcon name="bell" :size="16" />
                <span
                  v-if="activeAlarmsList.length > 0"
                  class="absolute -top-1 -right-1 min-w-[14px] h-[14px] px-0.5 rounded-full text-[8px] font-bold font-mono flex items-center justify-center text-white ring-1 ring-white dark:ring-[#0B1422]"
                  :class="activeCriticalCount > 0 ? 'bg-rose-500 animate-pulse' : 'bg-amber-500 text-slate-950'"
                >
                  {{ activeAlarmsList.length }}
                </span>
              </button>

              <!-- Active Alarms Popover Panel -->
              <div
                v-if="activeToolbarPopover === 'activeAlarms'"
                class="absolute left-0 top-full mt-1.5 z-50 w-80 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono"
                @click.stop
              >
                <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
                  <div class="flex items-center space-x-2">
                    <div class="p-1 rounded bg-rose-500/10 border border-rose-500/30 text-rose-500">
                      <OperatorDashboardIcon name="bell" :size="15" class="text-rose-500" />
                    </div>
                    <div>
                      <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                        {{ $t('operator.summaryAlarms') || 'Active Alarms' }}
                      </h3>
                      <div class="text-[9px] text-slate-500 dark:text-slate-400 mt-0.5 leading-none">
                        Limit Threshold & Incident Status
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1.5">
                    <span
                      class="text-[9px] font-bold px-1.5 py-0.5 rounded leading-none"
                      :class="activeCriticalCount > 0 ? 'bg-rose-500/15 text-rose-600 dark:text-rose-400 border border-rose-500/40' : activeAlarmsList.length > 0 ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/40' : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/40'"
                    >
                      {{ activeAlarmsList.length > 0 ? activeAlarmsList.length + ' ACTIVE' : 'NORMAL' }}
                    </span>
                    <button
                      type="button"
                      @click="closeToolbarPopover"
                      class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Main Alarm Metric Highlights -->
                <div class="p-2.5 rounded-md bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] mb-2.5">
                  <div class="flex items-baseline justify-between mb-1.5">
                    <div>
                      <span
                        class="text-[20px] font-bold tabular-nums tracking-tight"
                        :class="activeCriticalCount > 0 ? 'text-rose-600 dark:text-rose-400' : activeAlarmsList.length > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'"
                      >
                        {{ activeAlarmsList.length }}
                      </span>
                      <span class="text-[10px] text-slate-500 dark:text-slate-400 ml-1">incidents active</span>
                    </div>
                    <div class="flex items-center space-x-1.5 text-[10px]">
                      <span v-if="activeCriticalCount > 0" class="px-1.5 py-0.2 rounded font-bold bg-rose-500/15 text-rose-600 dark:text-rose-400 border border-rose-500/30">
                        {{ activeCriticalCount }} CRIT
                      </span>
                      <span v-if="activeWarningCount > 0" class="px-1.5 py-0.2 rounded font-bold bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30">
                        {{ activeWarningCount }} WARN
                      </span>
                      <span v-if="activeAlarmsList.length === 0" class="text-emerald-600 dark:text-emerald-400 font-semibold">
                        ✓ Normal Limits
                      </span>
                    </div>
                  </div>
                  <div class="text-[9px] text-slate-500 dark:text-slate-400 pt-1.5 border-t border-slate-200/50 dark:border-slate-800/60">
                    <span v-if="highestPriorityAlarm">
                      Top Priority: <strong class="text-rose-600 dark:text-rose-400">{{ highestPriorityAlarm.param_name || highestPriorityAlarm.alarm_code }}</strong> ({{ highestPriorityAlarm.message }})
                    </span>
                    <span v-else class="text-slate-600 dark:text-slate-400">
                      All sensor measurements are currently within regulatory thresholds.
                    </span>
                  </div>
                </div>

                <!-- Active Alarms Incident List -->
                <div class="mb-2.5">
                  <div class="text-[9px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 flex items-center justify-between">
                    <span>Active Incidents</span>
                    <router-link
                      to="/alarm/history"
                      @click.native="closeToolbarPopover"
                      class="text-cyan-600 dark:text-cyan-400 hover:underline"
                    >
                      History
                    </router-link>
                  </div>
                  <div v-if="activeAlarmsList.length > 0" class="max-h-36 overflow-y-auto custom-scroller space-y-1 pr-0.5">
                    <div
                      v-for="alm in activeAlarmsList.slice(0, 6)"
                      :key="'popover-alm-' + alm.alarm_id"
                      class="p-1.5 rounded bg-slate-50/70 dark:bg-[#0E1726] border border-slate-200/70 dark:border-[#1E3048] text-[9.5px]"
                    >
                      <div class="flex items-center justify-between mb-0.5">
                        <span class="font-bold text-slate-800 dark:text-slate-200 truncate pr-1">
                          {{ alm.param_name || alm.alarm_code }}
                        </span>
                        <span
                          class="px-1 py-0.2 rounded text-[7.5px] font-bold flex-shrink-0 uppercase"
                          :class="alm.severity === 'CRITICAL' ? 'bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/40' : 'bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/40'"
                        >
                          {{ alm.severity }}
                        </span>
                      </div>
                      <div class="text-[8.5px] text-slate-500 dark:text-slate-400 flex items-center justify-between">
                        <span>Val: <strong class="text-slate-700 dark:text-slate-300">{{ alm.trigger_value != null ? alm.trigger_value : '—' }}</strong> (Limit: {{ alm.threshold_value != null ? alm.threshold_value : '—' }})</span>
                        <span class="opacity-80">{{ formatRelative(alm.created_at) }}</span>
                      </div>
                    </div>
                  </div>
                  <div v-else class="p-3 text-center text-slate-400 dark:text-slate-500 text-[10px] bg-slate-50/50 dark:bg-[#0E1726]/50 rounded border border-dashed border-slate-200 dark:border-[#20354B]">
                    No active alarm incidents detected.
                  </div>
                </div>

                <!-- Action Link -->
                <router-link
                  to="/alarm/active"
                  @click.native="closeToolbarPopover"
                  class="w-full py-1.5 px-2 rounded text-3xs font-semibold transition border flex items-center justify-center space-x-1.5 bg-rose-500/10 text-rose-700 dark:text-rose-300 border-rose-500/30 hover:bg-rose-500/20 hover:border-rose-400"
                >
                  <span>🛡️</span>
                  <span>{{ $t('operator.viewConsole') || 'Open Alarm Console & Acknowledge →' }}</span>
                </router-link>
              </div>
            </div>

            <!-- 3. Latest Telemetry Popover -->
            <div class="relative toolbar-popover-anchor">
              <button
                type="button"
                @click.stop="toggleToolbarPopover('latestTelemetry')"
                class="w-8 h-8 flex items-center justify-center rounded transition border relative"
                :class="[
                  activeToolbarPopover === 'latestTelemetry'
                    ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border-cyan-500 shadow-xs'
                    : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'
                ]"
                :title="($t('operator.toolbarTelemetryTooltip') || $t('operator.summaryTelemetry') || 'Latest Telemetry') + ': ' + formatRelativeSeconds(lastTelemetrySecondsAgo) + ' ago (' + (isTelemetryStale ? ($t('operator.sinceLastValidData') || 'since last valid data') : ($t('operator.sinceLastPacket') || 'since last packet')) + ')'"
                aria-label="Latest Telemetry"
                :aria-expanded="activeToolbarPopover === 'latestTelemetry'"
              >
                <OperatorDashboardIcon name="activity" :size="16" />
                <span
                  class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full ring-1 ring-white dark:ring-[#0B1422] transition-colors"
                  :class="isTelemetryStale ? 'bg-rose-500 animate-pulse' : 'bg-emerald-400'"
                ></span>
              </button>

              <!-- Latest Telemetry Popover Panel -->
              <div
                v-if="activeToolbarPopover === 'latestTelemetry'"
                class="absolute left-0 top-full mt-1.5 z-50 w-80 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono"
                @click.stop
              >
                <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
                  <div class="flex items-center space-x-2">
                    <div class="p-1 rounded bg-cyan-500/10 border border-cyan-500/30 text-cyan-500">
                      <OperatorDashboardIcon name="activity" :size="15" class="text-cyan-500" />
                    </div>
                    <div>
                      <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                        {{ $t('operator.summaryTelemetry') || 'Latest Telemetry' }}
                      </h3>
                      <div class="text-[9px] text-slate-500 dark:text-slate-400 mt-0.5 leading-none">
                        {{ isTelemetryStale ? ($t('operator.sinceLastValidData') || 'since last valid data') : ($t('operator.sinceLastPacket') || 'since last packet') }}
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1.5">
                    <span
                      class="text-[9px] font-bold px-1.5 py-0.5 rounded leading-none"
                      :class="isTelemetryStale ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/30' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30'"
                    >
                      {{ isTelemetryStale ? ($t('operator.stale') || 'STALE') : ($t('operator.live') || 'LIVE') }}
                    </span>
                    <button
                      type="button"
                      @click="closeToolbarPopover"
                      class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Main Telemetry Highlights -->
                <div class="p-2.5 rounded-md bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] mb-2.5">
                  <div class="flex items-baseline justify-between mb-1.5">
                    <div>
                      <span
                        class="text-[20px] font-bold tabular-nums tracking-tight"
                        :class="isTelemetryStale ? 'text-rose-600 dark:text-rose-400' : 'text-slate-900 dark:text-white'"
                      >
                        {{ formatRelativeSeconds(lastTelemetrySecondsAgo) }}
                      </span>
                      <span class="text-[10px] text-slate-500 dark:text-slate-400 ml-1">since last packet</span>
                    </div>
                    <span class="text-[10.5px] font-bold" :class="summaryStats.goodSensors > 0 ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'">
                      {{ summaryStats.goodSensors }}/{{ allParameters.length }} Valid
                    </span>
                  </div>
                  <div class="w-full h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all duration-300"
                      :class="isTelemetryStale ? 'bg-rose-500' : 'bg-emerald-500'"
                      :style="{ width: Math.round((summaryStats.goodSensors / (allParameters.length || 1)) * 100) + '%' }"
                    ></div>
                  </div>
                  <div class="flex items-center justify-between text-[9px] text-slate-500 dark:text-slate-400 mt-2 pt-1.5 border-t border-slate-200/50 dark:border-slate-800/60">
                    <span>Packet: <strong class="text-slate-700 dark:text-slate-300">{{ lastTelemetryTimeStr || 'Just now' }}</strong></span>
                    <span>Timed out: <strong :class="summaryStats.staleSensors > 0 ? 'text-amber-500 font-bold' : 'text-slate-700 dark:text-slate-300'">{{ summaryStats.staleSensors }}</strong></span>
                  </div>
                </div>

                <!-- Stream Diagnostics Information -->
                <div class="p-2 rounded bg-slate-50/70 dark:bg-[#0E1726] border border-slate-200/70 dark:border-[#1E3048] mb-2.5 text-[9.5px] space-y-1">
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400">Stream Status:</span>
                    <span class="font-semibold" :class="isTelemetryStale ? 'text-rose-500' : 'text-emerald-500'">
                      {{ isTelemetryStale ? 'Delay Warning' : 'Active High-Speed Acquisition' }}
                    </span>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400">Polling Interval:</span>
                    <span>{{ Math.round(configuredPollingIntervalMs / 1000) }}s Dynamic Cycle</span>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400">Total Monitored:</span>
                    <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ allParameters.length }} Parameters</span>
                  </div>
                </div>

                <!-- Action Link -->
                <button
                  type="button"
                  @click="navigateToRoute('/monitoring/analysis')"
                  class="w-full py-1.5 px-2 rounded text-3xs font-semibold transition border flex items-center justify-center space-x-1.5 bg-slate-50 dark:bg-[#0E1726] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400"
                >
                  <span>📊</span>
                  <span>Open Telemetry Analytics & Trends →</span>
                </button>
              </div>
            </div>

            <!-- 4. System Monitor Popover -->
            <div class="relative toolbar-popover-anchor">
              <button
                type="button"
                @click.stop="toggleToolbarPopover('systemMonitor')"
                class="w-8 h-8 flex items-center justify-center rounded transition border relative"
                :class="[
                  activeToolbarPopover === 'systemMonitor'
                    ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border-cyan-500 shadow-xs'
                    : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'
                ]"
                :title="($t('operator.toolbarSystemMonitorTooltip') || $t('operator.summarySystem') || 'System Monitor') + ': CPU ' + Math.round(systemStatus.cpu_percent || 0) + '% · RAM ' + Math.round(systemStatus.ram_percent || 0) + '% · Disk ' + Math.round(systemStatus.disk_percent || 0) + '%'"
                aria-label="System Monitor"
                :aria-expanded="activeToolbarPopover === 'systemMonitor'"
              >
                <OperatorDashboardIcon name="sensor" :size="16" />
                <span
                  class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full ring-1 ring-white dark:ring-[#0B1422] transition-colors"
                  :class="(systemStatus.cpu_percent > 85 || systemStatus.ram_percent > 90) ? 'bg-amber-500' : 'bg-emerald-400'"
                ></span>
              </button>

              <!-- System Monitor Popover Panel -->
              <div
                v-if="activeToolbarPopover === 'systemMonitor'"
                class="absolute left-0 top-full mt-1.5 z-50 w-80 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono"
                @click.stop
              >
                <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
                  <div class="flex items-center space-x-2">
                    <div class="p-1 rounded bg-cyan-500/10 border border-cyan-500/30 text-cyan-500">
                      <OperatorDashboardIcon name="sensor" :size="15" class="text-cyan-500" />
                    </div>
                    <div>
                      <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                        {{ $t('operator.summarySystem') || 'System Monitor' }}
                      </h3>
                      <div class="text-[9px] text-slate-500 dark:text-slate-400 mt-0.5 leading-none">
                        Hardware & Node Health
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1.5">
                    <span class="text-[9px] font-bold px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 leading-none">
                      {{ systemStatus.service_status || 'RUNNING' }}
                    </span>
                    <button
                      type="button"
                      @click="closeToolbarPopover"
                      class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Main Hardware Gauges -->
                <div class="p-2.5 rounded-md bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] mb-2.5 space-y-2">
                  <!-- CPU -->
                  <div>
                    <div class="flex justify-between items-center text-[10px] mb-0.5">
                      <span class="text-slate-500 dark:text-slate-400">CPU Utilization</span>
                      <span class="font-bold text-slate-900 dark:text-white tabular-nums">{{ Math.round(systemStatus.cpu_percent || 0) }}%</span>
                    </div>
                    <div class="w-full h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all duration-300"
                        :class="systemStatus.cpu_percent > 85 ? 'bg-rose-500' : systemStatus.cpu_percent > 70 ? 'bg-amber-500' : 'bg-cyan-500'"
                        :style="{ width: Math.round(systemStatus.cpu_percent || 0) + '%' }"
                      ></div>
                    </div>
                  </div>
                  <!-- RAM -->
                  <div>
                    <div class="flex justify-between items-center text-[10px] mb-0.5">
                      <span class="text-slate-500 dark:text-slate-400">RAM Utilization</span>
                      <span class="font-bold text-slate-900 dark:text-white tabular-nums">{{ Math.round(systemStatus.ram_percent || 0) }}%</span>
                    </div>
                    <div class="w-full h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all duration-300"
                        :class="systemStatus.ram_percent > 90 ? 'bg-rose-500' : systemStatus.ram_percent > 75 ? 'bg-amber-500' : 'bg-blue-500'"
                        :style="{ width: Math.round(systemStatus.ram_percent || 0) + '%' }"
                      ></div>
                    </div>
                  </div>
                  <!-- Disk -->
                  <div>
                    <div class="flex justify-between items-center text-[10px] mb-0.5">
                      <span class="text-slate-500 dark:text-slate-400">Disk Storage</span>
                      <span class="font-bold text-slate-900 dark:text-white tabular-nums">{{ Math.round(systemStatus.disk_percent || 0) }}%</span>
                    </div>
                    <div class="w-full h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all duration-300"
                        :class="systemStatus.disk_percent > 90 ? 'bg-rose-500' : 'bg-emerald-500'"
                        :style="{ width: Math.round(systemStatus.disk_percent || 0) + '%' }"
                      ></div>
                    </div>
                  </div>
                </div>

                <!-- Database & Health Details -->
                <div class="p-2 rounded bg-slate-50/70 dark:bg-[#0E1726] border border-slate-200/70 dark:border-[#1E3048] mb-2.5 text-[9.5px] space-y-1">
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400">Database Latency:</span>
                    <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ systemStatus.db_latency_ms != null ? systemStatus.db_latency_ms + 'ms' : '0.8ms' }}</span>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400">WAL Queue Status:</span>
                    <span class="font-semibold text-emerald-500">{{ systemStatus.queue_status || 'HEALTHY' }}</span>
                  </div>
                  <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                    <span class="text-slate-400">Station Node:</span>
                    <span class="font-mono text-slate-700 dark:text-slate-300">{{ systemStatus.hostname || 'Edge-Node-01' }}</span>
                  </div>
                </div>

                <!-- Action Link -->
                <button
                  type="button"
                  @click="navigateToRoute('/monitoring/system')"
                  class="w-full py-1.5 px-2 rounded text-3xs font-semibold transition border flex items-center justify-center space-x-1.5 bg-slate-50 dark:bg-[#0E1726] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400"
                >
                  <span>⚡</span>
                  <span>Open Full System Diagnostics →</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Active Filter Dismissible Badges (Minimalist) -->
          <div v-if="selectedCategoryFilter !== 'all' || filterPinnedOnly" class="flex items-center space-x-1 ml-1 pl-1.5 border-l border-slate-200/70 dark:border-[#1E3048]">
            <button
              type="button"
              v-if="selectedCategoryFilter !== 'all'"
              @click="selectedCategoryFilter = 'all'"
              class="flex items-center space-x-1 px-1.5 py-0.5 rounded text-[9px] font-mono font-semibold bg-cyan-500/15 text-cyan-700 dark:text-cyan-300 border border-cyan-500/40 hover:bg-cyan-500/25 transition leading-none"
              title="Clear category filter"
            >
              <span>🏷️</span>
              <span class="max-w-[120px] truncate">{{ selectedCategoryTitle }}</span>
              <span class="text-xs font-bold leading-none ml-0.5">×</span>
            </button>
            <button
              type="button"
              v-if="filterPinnedOnly"
              @click="toggleFilterPinned"
              class="flex items-center space-x-1 px-1.5 py-0.5 rounded text-[9px] font-mono font-semibold bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/40 hover:bg-amber-500/25 transition leading-none"
              title="Clear pinned-only filter"
            >
              <span>⭐</span>
              <span>{{ $t('operator.pinnedOnly') || 'Pinned' }}</span>
              <span class="text-xs font-bold leading-none ml-0.5">×</span>
            </button>
          </div>
        </div>

        <!-- Right: Workspace Settings Icons + Fullscreen Control -->
        <div class="flex items-center space-x-1.5 flex-shrink-0" ref="toolbarContainer">
          <!-- A. Categories Settings Popover -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('categories')"
              class="w-8 h-8 flex items-center justify-center rounded transition border relative"
              :class="[
                activeToolbarPopover === 'categories'
                  ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border-cyan-500 shadow-xs'
                  : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white',
                selectedCategoryFilter !== 'all' ? 'ring-1 ring-cyan-500/50' : ''
              ]"
              :title="$t('operator.toolbarCategoriesTooltip') || 'Categories & Filtering'"
              aria-label="Parameter Categories"
              :aria-expanded="activeToolbarPopover === 'categories'"
            >
              <OperatorDashboardIcon name="filter" :size="16" />
              <span
                v-if="selectedCategoryFilter !== 'all'"
                class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full bg-cyan-400 ring-1 ring-white dark:ring-[#0B1422]"
              ></span>
            </button>

            <!-- Categories Popover Panel -->
            <div
              v-if="activeToolbarPopover === 'categories'"
              class="absolute right-0 top-full mt-1.5 z-50 w-72 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-2.5 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn"
              @click.stop
            >
              <div class="flex items-center justify-between pb-2 mb-2 border-b border-slate-100 dark:border-[#1E3048]">
                <div class="flex items-center space-x-1.5">
                  <OperatorDashboardIcon name="filter" :size="15" class="text-cyan-500" />
                  <span class="text-xs font-bold uppercase tracking-wider font-mono">
                    {{ $t('operator.categories') || 'Parameter Categories' }}
                  </span>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- Categories List -->
              <div class="max-h-64 overflow-y-auto custom-scroller space-y-1 font-mono text-[10.5px]">
                <!-- All Categories -->
                <button
                  type="button"
                  @click="selectedCategoryFilter = 'all'"
                  class="w-full flex items-center justify-between px-2 py-1.5 rounded transition border text-left"
                  :class="selectedCategoryFilter === 'all'
                    ? 'bg-cyan-500/15 dark:bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border-cyan-500/50 shadow-xs font-bold'
                    : 'border-transparent text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'"
                >
                  <div class="flex items-center space-x-2 truncate">
                    <OperatorDashboardIcon name="menu" :size="12" class="text-cyan-500 flex-shrink-0" />
                    <span class="truncate">{{ $t('operator.allCategories') || 'All Categories' }}</span>
                  </div>
                  <div class="flex items-center space-x-1.5 flex-shrink-0">
                    <span class="px-1.5 py-0.2 rounded-full text-[9px] font-bold bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                      {{ allParameters.length }}
                    </span>
                    <span v-if="selectedCategoryFilter === 'all'" class="text-cyan-500 font-bold text-xs">✓</span>
                  </div>
                </button>

                <!-- Dynamic Category Tabs -->
                <button
                  type="button"
                  v-for="cat in activeCategoriesList"
                  :key="cat.id"
                  @click="selectedCategoryFilter = cat.id"
                  class="w-full flex items-center justify-between px-2 py-1.5 rounded transition border text-left"
                  :class="selectedCategoryFilter === cat.id
                    ? getCategoryTheme(cat.color_key).activeTab + ' font-bold'
                    : 'border-transparent text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'"
                >
                  <div class="flex items-center space-x-2 truncate">
                    <span class="w-2 h-2 rounded-full flex-shrink-0" :class="getCategoryTheme(cat.color_key).dot"></span>
                    <span class="truncate">{{ cat.name }}</span>
                  </div>
                  <div class="flex items-center space-x-1.5 flex-shrink-0">
                    <span
                      class="px-1.5 py-0.2 rounded-full text-[9px] font-bold"
                      :class="selectedCategoryFilter === cat.id ? getCategoryTheme(cat.color_key).pill : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400'"
                    >
                      {{ getCategoryParamCount(cat.id) }}
                    </span>
                    <span v-if="selectedCategoryFilter === cat.id" class="text-cyan-500 font-bold text-xs">✓</span>
                  </div>
                </button>

                <!-- Uncategorized -->
                <button
                  type="button"
                  v-if="uncategorizedParamCount > 0 || selectedCategoryFilter === 'uncategorized'"
                  @click="selectedCategoryFilter = 'uncategorized'"
                  class="w-full flex items-center justify-between px-2 py-1.5 rounded transition border text-left"
                  :class="selectedCategoryFilter === 'uncategorized'
                    ? 'bg-slate-500/20 text-slate-200 border-slate-400/50 shadow-xs font-bold'
                    : 'border-transparent text-slate-500 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-[#162538] hover:text-slate-800 dark:hover:text-slate-200'"
                >
                  <div class="flex items-center space-x-2 truncate">
                    <span>📦</span>
                    <span class="truncate">{{ $t('operator.uncategorized') || 'Uncategorized' }}</span>
                  </div>
                  <div class="flex items-center space-x-1.5 flex-shrink-0">
                    <span class="px-1.5 py-0.2 rounded-full text-[9px] font-bold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                      {{ uncategorizedParamCount }}
                    </span>
                    <span v-if="selectedCategoryFilter === 'uncategorized'" class="text-slate-300 font-bold text-xs">✓</span>
                  </div>
                </button>
              </div>

              <!-- Footer: Category Management Modal Trigger -->
              <div v-if="canManageCategories || canViewCategories" class="pt-2 mt-2 border-t border-slate-100 dark:border-[#1E3048]">
                <button
                  type="button"
                  @click="openCategoryManagementModal(); closeToolbarPopover()"
                  class="w-full py-1.5 px-2 rounded text-3xs font-mono font-semibold transition border flex items-center justify-center space-x-1.5 bg-slate-50 dark:bg-[#0E1726] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400"
                >
                  <span>🏷️</span>
                  <span>{{ $t('operator.manageCategories') || 'Configure Categories & Layout' }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- B. Pinned Parameters Settings Popover -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('pinned')"
              class="w-8 h-8 flex items-center justify-center rounded transition border relative"
              :class="[
                activeToolbarPopover === 'pinned'
                  ? 'bg-amber-500/20 text-amber-600 dark:text-amber-300 border-amber-500 shadow-xs'
                  : filterPinnedOnly
                    ? 'bg-amber-500/15 text-amber-600 dark:text-amber-300 border-amber-500/50'
                    : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white',
                filterPinnedOnly ? 'ring-1 ring-amber-500/50' : ''
              ]"
              :title="$t('operator.toolbarPinnedTooltip') || 'Pinned Favorite Parameters'"
              aria-label="Pinned Parameters"
              :aria-expanded="activeToolbarPopover === 'pinned'"
            >
              <OperatorDashboardIcon name="pin" :size="16" />
              <span
                v-if="pinnedIds.length > 0"
                class="absolute -top-1 -right-1 min-w-[14px] h-[14px] px-0.5 rounded-full text-[8px] font-bold font-mono flex items-center justify-center bg-amber-500 text-slate-950 ring-1 ring-white dark:ring-[#0B1422]"
              >
                {{ pinnedIds.length }}
              </span>
            </button>

            <!-- Pinned Popover Panel -->
            <div
              v-if="activeToolbarPopover === 'pinned'"
              class="absolute right-0 top-full mt-1.5 z-50 w-72 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-2.5 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn"
              @click.stop
            >
              <div class="flex items-center justify-between pb-2 mb-2 border-b border-slate-100 dark:border-[#1E3048]">
                <div class="flex items-center space-x-1.5">
                  <span class="text-amber-400">⭐</span>
                  <span class="text-xs font-bold uppercase tracking-wider font-mono">
                    {{ $t('operator.pinnedParameters') || 'Pinned Parameters' }}
                  </span>
                  <span class="px-1.5 py-0.2 rounded-full text-[9px] font-mono font-bold bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/30">
                    {{ pinnedIds.length }}
                  </span>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- Pinned Only Quick Toggle Row -->
              <div class="flex items-center justify-between p-2 rounded bg-slate-50 dark:bg-[#0E1726] border border-slate-200 dark:border-[#20354B] mb-2 font-mono">
                <div>
                  <div class="text-[10.5px] font-bold text-slate-900 dark:text-white">
                    {{ $t('operator.pinnedOnly') || 'Pinned Only' }}
                  </div>
                  <div class="text-[9px] text-slate-500 dark:text-slate-400">
                    Filter SCADA cards to favorites
                  </div>
                </div>
                <button
                  type="button"
                  @click="toggleFilterPinned"
                  class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                  :class="filterPinnedOnly ? 'bg-amber-500' : 'bg-slate-300 dark:bg-slate-700'"
                  role="switch"
                  :aria-checked="filterPinnedOnly"
                >
                  <span
                    aria-hidden="true"
                    class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                    :class="filterPinnedOnly ? 'translate-x-4' : 'translate-x-0'"
                  ></span>
                </button>
              </div>

              <!-- Pinned Parameters List -->
              <div class="text-[10px] font-mono">
                <div class="text-[9px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1 flex justify-between items-center">
                  <span>Favorites ({{ pinnedParametersList.length }})</span>
                  <button
                    type="button"
                    v-if="pinnedIds.length > 0"
                    @click="clearAllPinned"
                    class="text-rose-500 hover:text-rose-600 dark:hover:text-rose-400 hover:underline"
                  >
                    {{ $t('operator.clearAllPinned') || 'Clear All' }}
                  </button>
                </div>

                <div v-if="pinnedParametersList.length > 0" class="max-h-44 overflow-y-auto custom-scroller space-y-1">
                  <div
                    v-for="param in pinnedParametersList"
                    :key="'pinned-item-' + param.param_id"
                    class="flex items-center justify-between p-1.5 rounded bg-slate-50/70 dark:bg-[#0A1220] border border-slate-200/70 dark:border-[#1E3048] hover:border-slate-300 dark:hover:border-slate-700 transition"
                  >
                    <div class="min-w-0 flex-1 pr-1.5">
                      <div class="font-bold text-slate-800 dark:text-slate-200 truncate text-[10px]">
                        {{ param.param_name || param.param_code }}
                      </div>
                      <div class="text-[8.5px] text-slate-500 dark:text-slate-400 flex items-center space-x-1.5 truncate">
                        <span>{{ param.device_code }}</span>
                        <span v-if="param.unit">· {{ param.unit }}</span>
                      </div>
                    </div>
                    <button
                      type="button"
                      @click.stop="togglePin(param.param_id)"
                      class="p-1 rounded text-amber-400 hover:text-rose-500 transition hover:bg-slate-200 dark:hover:bg-slate-800 flex-shrink-0"
                      title="Unpin parameter"
                    >
                      <OperatorDashboardIcon name="pin" :size="14" class="text-amber-400" />
                    </button>
                  </div>
                </div>
                <div v-else class="p-3 text-center text-slate-400 dark:text-slate-500 text-[10px] bg-slate-50/50 dark:bg-[#0E1726]/50 rounded border border-dashed border-slate-200 dark:border-[#20354B]">
                  {{ $t('operator.noPinnedParams') || 'No parameters pinned yet. Click the star icon on any parameter card to pin it here.' }}
                </div>
              </div>
            </div>
          </div>

          <!-- C. Display Mode Settings Popover (Cards vs Table) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('displayMode')"
              class="w-8 h-8 flex items-center justify-center rounded transition border"
              :class="[
                activeToolbarPopover === 'displayMode'
                  ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border-cyan-500 shadow-xs'
                  : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'
              ]"
              :title="$t('operator.toolbarDisplayModeTooltip') || 'Display Mode (Cards / Table)'"
              aria-label="Display Mode"
              :aria-expanded="activeToolbarPopover === 'displayMode'"
            >
              <OperatorDashboardIcon :name="viewMode === 'cards' ? 'layout' : 'menu'" :size="16" />
            </button>

            <!-- Display Mode Popover Panel -->
            <div
              v-if="activeToolbarPopover === 'displayMode'"
              class="absolute right-0 top-full mt-1.5 z-50 w-72 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-2.5 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn"
              @click.stop
            >
              <div class="flex items-center justify-between pb-2 mb-2 border-b border-slate-100 dark:border-[#1E3048]">
                <div class="flex items-center space-x-1.5">
                  <OperatorDashboardIcon name="layout" :size="15" class="text-cyan-500" />
                  <span class="text-xs font-bold uppercase tracking-wider font-mono">
                    {{ $t('operator.displayMode') || 'Display Mode' }}
                  </span>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- Mode Options -->
              <div class="space-y-1.5 font-mono">
                <!-- Cards View -->
                <button
                  type="button"
                  @click="viewMode = 'cards'"
                  class="w-full p-2 rounded-md border text-left transition flex items-start space-x-2.5"
                  :class="viewMode === 'cards'
                    ? 'bg-cyan-500/15 dark:bg-cyan-500/20 text-slate-900 dark:text-white border-cyan-500/60 shadow-xs'
                    : 'bg-slate-50 dark:bg-[#0E1726] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:border-slate-300 dark:hover:border-slate-700'"
                >
                  <div class="p-1.5 rounded bg-white dark:bg-[#162538] border border-slate-200 dark:border-[#20354B] text-cyan-500 mt-0.5 flex-shrink-0">
                    <OperatorDashboardIcon name="layout" :size="16" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center justify-between">
                      <span class="text-[11px] font-bold text-slate-900 dark:text-white">
                        {{ $t('operator.viewCards') || 'Cards View' }}
                      </span>
                      <span v-if="viewMode === 'cards'" class="text-cyan-500 font-bold text-xs">✓ Active</span>
                    </div>
                    <div class="text-[9.5px] text-slate-500 dark:text-slate-400 mt-0.5 leading-tight">
                      {{ $t('operator.cardsViewDesc') || 'Real-time SCADA tiles with sparklines and status badges' }}
                    </div>
                  </div>
                </button>

                <!-- Table View -->
                <button
                  type="button"
                  @click="viewMode = 'table'"
                  class="w-full p-2 rounded-md border text-left transition flex items-start space-x-2.5"
                  :class="viewMode === 'table'
                    ? 'bg-cyan-500/15 dark:bg-cyan-500/20 text-slate-900 dark:text-white border-cyan-500/60 shadow-xs'
                    : 'bg-slate-50 dark:bg-[#0E1726] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:border-slate-300 dark:hover:border-slate-700'"
                >
                  <div class="p-1.5 rounded bg-white dark:bg-[#162538] border border-slate-200 dark:border-[#20354B] text-cyan-500 mt-0.5 flex-shrink-0">
                    <OperatorDashboardIcon name="menu" :size="16" />
                  </div>
                  <div class="min-w-0 flex-1">
                    <div class="flex items-center justify-between">
                      <span class="text-[11px] font-bold text-slate-900 dark:text-white">
                        {{ $t('operator.viewTable') || 'Table View' }}
                      </span>
                      <span v-if="viewMode === 'table'" class="text-cyan-500 font-bold text-xs">✓ Active</span>
                    </div>
                    <div class="text-[9.5px] text-slate-500 dark:text-slate-400 mt-0.5 leading-tight">
                      {{ $t('operator.tableViewDesc') || 'Dense tabular spreadsheet view with sortable columns' }}
                    </div>
                  </div>
                </button>
              </div>
            </div>
          </div>

          <!-- D. Panel Layout Settings Popover (1P, 2P, 3P, 4P) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('panelLayout')"
              class="w-8 h-8 flex items-center justify-center rounded transition border relative"
              :class="[
                activeToolbarPopover === 'panelLayout'
                  ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border-cyan-500 shadow-xs'
                  : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'
              ]"
              :title="$t('operator.toolbarPanelLayoutTooltip') || 'Panel Layout (1P - 4P Columns)'"
              aria-label="Panel Layout"
              :aria-expanded="activeToolbarPopover === 'panelLayout'"
            >
              <OperatorDashboardIcon name="layout" :size="16" />
              <span class="absolute -top-1.5 -right-1.5 text-[8px] font-bold font-mono px-1 py-0.2 rounded-full bg-slate-200 dark:bg-[#1E3048] text-cyan-700 dark:text-cyan-300 border border-slate-300 dark:border-[#20354B] ring-1 ring-white dark:ring-[#0B1422] leading-none pointer-events-none">
                {{ gridColumns }}P
              </span>
            </button>

            <!-- Panel Layout Popover Panel -->
            <div
              v-if="activeToolbarPopover === 'panelLayout'"
              class="absolute right-0 top-full mt-1.5 z-50 w-72 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-2.5 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn"
              @click.stop
            >
              <div class="flex items-center justify-between pb-2 mb-2 border-b border-slate-100 dark:border-[#1E3048]">
                <div class="flex items-center space-x-1.5">
                  <OperatorDashboardIcon name="layout" :size="15" class="text-cyan-500" />
                  <span class="text-xs font-bold uppercase tracking-wider font-mono">
                    {{ $t('operator.panelLayout') || 'Panel Layout' }}
                  </span>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <div class="text-[9.5px] font-mono text-slate-500 dark:text-slate-400 mb-2 leading-tight">
                {{ $t('operator.panelLayoutDesc') || 'Card columns inside the single parameter panel:' }}
              </div>

              <!-- 1P, 2P, 3P, 4P Controls -->
              <div class="grid grid-cols-2 gap-1.5 font-mono">
                <button
                  type="button"
                  v-for="cnt in [1, 2, 3, 4]"
                  :key="'occ-panel-btn-' + cnt"
                  :id="'btn-toolbar-panel-' + cnt"
                  @click="setGridColumns(cnt)"
                  class="p-2 rounded-md border text-left transition flex items-center justify-between"
                  :class="gridColumns === cnt
                    ? 'bg-cyan-500/15 dark:bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border-cyan-500/60 shadow-xs font-bold'
                    : 'bg-slate-50 dark:bg-[#0E1726] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:border-slate-300 dark:hover:border-slate-700'"
                  :title="cnt + (cnt === 1 ? ' Column Grid (1P)' : ' Columns Grid (' + cnt + 'P)')"
                >
                  <div class="flex items-center space-x-2">
                    <span class="text-xs font-black px-1.5 py-0.5 rounded bg-white dark:bg-[#162538] border border-slate-200 dark:border-[#20354B]">
                      {{ cnt }}P
                    </span>
                    <span class="text-[10px]">
                      {{ cnt === 1 ? '1 Col' : cnt + ' Cols' }}
                    </span>
                  </div>
                  <span v-if="gridColumns === cnt" class="text-cyan-500 font-bold text-xs">✓</span>
                </button>
              </div>

              <div class="mt-2 pt-2 border-t border-slate-100 dark:border-[#1E3048] text-[8.5px] font-mono text-slate-400 text-center">
                Preference saved automatically (Default: 4P)
              </div>
            </div>
          </div>

          <!-- F. ISPU Settings Popover -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('ispuSettings')"
              class="w-8 h-8 flex items-center justify-center rounded transition border relative"
              :class="[
                activeToolbarPopover === 'ispuSettings'
                  ? 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-300 border-emerald-500 shadow-xs'
                  : isAQMSDomain && isIspuBarVisible
                    ? 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border-cyan-500/40'
                    : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'
              ]"
              :title="$t('operator.toolbarIspuSettingsTooltip') || 'ISPU Settings & Status'"
              aria-label="ISPU Settings"
              :aria-expanded="activeToolbarPopover === 'ispuSettings'"
            >
              <OperatorDashboardIcon name="ispu" :size="16" />
              <span
                v-if="isAQMSDomain && isIspuBarVisible"
                class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full bg-emerald-400 ring-1 ring-white dark:ring-[#0B1422]"
              ></span>
            </button>

            <!-- ISPU Settings Popover Panel -->
            <div
              v-if="activeToolbarPopover === 'ispuSettings'"
              class="absolute right-0 top-full mt-1.5 z-50 w-72 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-2.5 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn"
              @click.stop
            >
              <div class="flex items-center justify-between pb-2 mb-2 border-b border-slate-100 dark:border-[#1E3048]">
                <div class="flex items-center space-x-1.5">
                  <OperatorDashboardIcon name="ispu" :size="15" class="text-emerald-500" />
                  <span class="text-xs font-bold uppercase tracking-wider font-mono">
                    {{ $t('operator.ispuSettings') || 'ISPU Settings' }}
                  </span>
                  <span v-if="isAQMSDomain" class="px-1.5 py-0.2 rounded-full text-[9px] font-mono font-bold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30">
                    AQMS
                  </span>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- Toggle Show/Hide ISPU Panel Row -->
              <div class="flex items-center justify-between p-2 rounded bg-slate-50 dark:bg-[#0E1726] border border-slate-200 dark:border-[#20354B] mb-2 font-mono">
                <div>
                  <div class="text-[10.5px] font-bold text-slate-900 dark:text-white">
                    {{ $t('operator.showIspuBar') || 'Show ISPU Indicator Bar' }}
                  </div>
                  <div class="text-[9px] text-slate-500 dark:text-slate-400">
                    {{ $t('operator.showIspuBarDesc') || 'Display air quality index bar below parameters' }}
                  </div>
                </div>
                <button
                  type="button"
                  @click="toggleIspuBarVisibility"
                  class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                  :class="isIspuBarVisible ? 'bg-cyan-500' : 'bg-slate-300 dark:bg-slate-700'"
                  role="switch"
                  :aria-checked="isIspuBarVisible"
                >
                  <span
                    aria-hidden="true"
                    class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                    :class="isIspuBarVisible ? 'translate-x-4' : 'translate-x-0'"
                  ></span>
                </button>
              </div>

              <!-- Standard Info Card -->
              <div class="p-2 rounded bg-slate-50/70 dark:bg-[#0A1220] border border-slate-200/80 dark:border-[#1E3048] mb-2 text-[9.5px] font-mono space-y-1">
                <div class="flex items-center justify-between text-slate-700 dark:text-slate-300">
                  <span class="text-slate-500 dark:text-slate-400">Standard:</span>
                  <span class="font-bold text-cyan-600 dark:text-cyan-400">Permen LHK No. 14 / 2020</span>
                </div>
                <div class="flex items-center justify-between text-slate-700 dark:text-slate-300">
                  <span class="text-slate-500 dark:text-slate-400">Interval:</span>
                  <span>24-Hour Rolling Average</span>
                </div>
                <div class="flex items-center justify-between text-slate-700 dark:text-slate-300">
                  <span class="text-slate-500 dark:text-slate-400">Parameters:</span>
                  <span>PM2.5, PM10, SO2, NO2, O3, CO</span>
                </div>
              </div>

              <!-- Navigation to Dedicated ISPU Analysis -->
              <button
                type="button"
                @click="navigateToRoute('/monitoring/ispu')"
                class="w-full py-1.5 px-2 rounded text-3xs font-mono font-semibold transition border flex items-center justify-center space-x-1.5 bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border-cyan-500/30 hover:bg-cyan-500/20 hover:border-cyan-400"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                <span>{{ $t('operator.openIspuAnalysis') || 'Open ISPU Analysis & Reports →' }}</span>
              </button>
            </div>
          </div>

          <!-- Thin Vertical Divider before Fullscreen & Layout Controls -->
          <div class="h-4 w-px bg-slate-200 dark:bg-[#1E3048] mx-0.5 flex-shrink-0"></div>

          <!-- Sidebar Auto-Hide Setup Toggle Button -->
          <button
            type="button"
            id="btn-operator-sidebar-autohide"
            @click.stop="toggleSidebarAutoHide"
            class="w-8 h-8 flex items-center justify-center rounded transition border cursor-pointer"
            :class="sidebarAutoHide
              ? 'bg-cyan-500/20 text-cyan-400 border-cyan-500/50 ring-1 ring-cyan-500/30 shadow-xs'
              : 'bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#162538] text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border-slate-200 dark:border-[#20354B]'"
            :title="sidebarAutoHide ? ($t('header.sidebarAutoHideOn') || 'Sidebar: Auto-Hide Enabled (Click to Pin)') : ($t('header.sidebarAutoHideOff') || 'Sidebar: Pinned (Click for Auto-Hide)')"
            aria-label="Toggle Sidebar Auto-Hide"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <path d="M9 3v18" />
              <path v-if="!sidebarAutoHide" d="m14 9-3 3 3 3" stroke-linecap="round" stroke-linejoin="round" />
              <path v-else d="m12 9 3 3-3 3" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>

          <!-- Fullscreen / Sidebar & Top Panel Layout Toggle (Icon Only) -->
          <button
            type="button"
            @click="$emit('toggle-sidebar')"
            class="w-8 h-8 flex items-center justify-center rounded transition border text-cyan-600 dark:text-cyan-400"
            :class="isSidebarCollapsed
              ? 'bg-cyan-600/15 hover:bg-cyan-600/25 border-cyan-500/50 shadow-xs'
              : 'bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#162538] border-slate-200 dark:border-[#20354B]'"
            :title="isSidebarCollapsed ? ($t('operator.exitFullWidth') || 'Exit Fullscreen') : ($t('operator.fullWidth') || 'Full Width SCADA')"
            :aria-label="isSidebarCollapsed ? ($t('operator.exitFullWidth') || 'Exit Fullscreen') : ($t('operator.fullWidth') || 'Full Width SCADA')"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path v-if="!isSidebarCollapsed" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
              <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 4H5a1 1 0 00-1 1v4m0 0l5-5m6 0h4a1 1 0 011 1v4m0 0l-5-5M9 20H5a1 1 0 01-1-1v-4m0 0l5 5m6 0h4a1 1 0 001-1v-4m0 0l-5 5" />
            </svg>
          </button>
        </div>
      </div>
    </div>
    <!-- END OF SCADA UNIFIED HEADER CARD -->

    <!-- ======================================================== -->
    <!-- SECTION 1: SINGLE MAIN PARAMETER PANEL & ISPU BELOW      -->
    <!-- ======================================================== -->
    <div class="w-full flex flex-col space-y-1">

        <!-- VIEW MODE 1: SCADA PARAMETER CARD TILES (Single Main Parameter Panel) -->
        <div
          v-if="viewMode === 'cards'"
          class="scada-single-main-panel rounded-lg border border-slate-200 dark:border-[#20354B] bg-slate-50/50 dark:bg-[#0A1220]/60 p-1.5 transition-all duration-150 min-w-0"
        >
          <!-- Compact Panel Header -->
          <div class="flex items-center justify-between min-w-0 gap-1 pb-1 mb-1 border-b border-slate-200/80 dark:border-[#1E3048] flex-shrink-0 select-none">
            <div class="flex items-center space-x-1.5 min-w-0 flex-1 truncate">
              <span class="w-2 h-2 rounded-full bg-cyan-500 shadow-xs shadow-cyan-500/50 flex-shrink-0"></span>
              <span class="font-bold text-slate-800 dark:text-slate-100 uppercase tracking-wider text-[10.5px] truncate font-mono">
                {{ selectedCategoryTitle }}
              </span>
              <span v-if="selectedCategoryFilter !== 'all'" class="px-1.5 py-0.2 rounded-full text-[8px] font-mono font-bold bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30">
                Filtered
              </span>
            </div>
          </div>

          <!-- Internal scrollable card area if parameters exceed viewport height -->
          <div
            class="custom-scroller overflow-y-auto pr-0.5"
            :style="singlePanelScrollStyle"
          >
            <!-- Single unified grid with responsive columns based on 1P/2P/3P/4P -->
            <div
              v-if="filteredParameters.length > 0"
              class="grid gap-1.5"
              :class="singlePanelGridClass"
            >
              <div
                v-for="param in filteredParameters"
                :key="param.param_id"
                @click="selectParamForTrend(param)"
                @contextmenu.prevent="openParamContextMenu($event, param)"
                class="scada-param-tile px-2 py-1 rounded-md border transition-all duration-150 relative cursor-pointer group flex flex-col justify-between h-[58px] min-w-0 box-border overflow-hidden"
                :class="[
                  isParamAlarmOrStaleBad(param)
                    ? (isAlarmPulseActive
                        ? (alarmPulseSeverity === 'CRITICAL' || param.alarm_status === 'CRITICAL' || param.quality === 'BAD'
                            ? 'border-rose-500 bg-rose-500/15 dark:bg-rose-950/45 shadow-sm shadow-rose-500/30 transition-colors duration-75'
                            : 'border-amber-500 bg-amber-500/15 dark:bg-amber-950/45 shadow-sm shadow-amber-500/30 transition-colors duration-75')
                        : (alarmPulseSeverity === 'CRITICAL' || param.alarm_status === 'CRITICAL' || param.quality === 'BAD'
                            ? 'border-rose-500/70 bg-rose-500/5 dark:bg-rose-950/20 transition-colors duration-300'
                            : 'border-amber-500/70 bg-amber-500/5 dark:bg-amber-950/20 transition-colors duration-300'))
                    : (isParamSelectedForTrend(param)
                        ? 'bg-white dark:bg-[#0B1422] border-cyan-500/75 dark:border-cyan-400/65 ring-1 ring-cyan-500/25 dark:ring-cyan-400/20 shadow-xs hover:border-cyan-500 dark:hover:border-cyan-400'
                        : 'bg-white dark:bg-[#0B1422] border-slate-200 dark:border-[#20354B] hover:border-slate-300 dark:hover:border-[#334b66] shadow-xs dark:shadow-none'),
                  liveBlinks[param.param_id] ? 'bg-cyan-50 dark:bg-cyan-950/40' : ''
                ]"
                :title="getParamTimingTitle(param)"
              >
                <!-- Area 1: Header (Device ID, Icon & Name, Category, Quality, Alarm) -->
                <div class="flex items-center justify-between min-w-0 gap-1 leading-none select-none">
                  <!-- Left: Status Dot + Parameter Icon & Name (Device ID removed, shifted left) -->
                  <div class="flex items-center min-w-0 flex-1 gap-1 truncate" :data-device="param.device_code">
                    <span
                      class="w-1.5 h-1.5 rounded-full flex-shrink-0 transition-all duration-150"
                      :class="[
                        (param.device_status === 'ONLINE' || param.device_status === 'CONNECTED') ? 'bg-emerald-500' : 'bg-rose-500',
                        isParamAlarmOrStaleBad(param) && isAlarmPulseActive ? 'scale-125 ring-2 ring-rose-400' : 'scale-100'
                      ]"
                      :title="(param.device_name ? (param.device_name + ' (' + param.device_code + ')') : param.device_code) + ' · ' + (param.device_status || 'UNKNOWN')"
                    ></span>
                    <div class="flex items-center gap-0.5 min-w-0 flex-1 truncate">
                      <ParameterIcon
                        :icon="param.icon"
                        :code="param.param_code"
                        :name="param.param_name"
                        :unit="param.unit"
                        size-class="w-3 h-3 flex-shrink-0"
                      />
                      <span
                        class="text-[10px] sm:text-[10.5px] font-bold text-cyan-600 dark:text-cyan-400 truncate tracking-tight font-mono"
                        :title="(param.param_name && param.param_name !== param.param_code) ? (param.param_name + ' (' + param.param_code + ')') : param.param_code"
                      >
                        {{ param.param_name || param.param_code }}
                      </span>
                    </div>
                  </div>

                  <!-- Header Right: Category Badge, Alarm Tag, Quality Badge, Pin Button -->
                  <div class="flex items-center gap-1 flex-shrink-0">
                    <!-- Category Compact Badge -->
                    <span
                      v-if="getParamCategory(param)"
                      class="px-1 py-0.2 rounded text-[7.5px] font-mono font-semibold uppercase tracking-tight flex-shrink-0 leading-none"
                      :class="getParamCategoryTheme(param).pill"
                      :title="'Category: ' + getParamCategory(param).name"
                    >
                      {{ getParamCategoryShortLabel(param) }}
                    </span>
                    <!-- Alarm Tag if active -->
                    <span
                      v-if="param.alarm_status === 'CRITICAL' || param.alarm_status === 'WARNING'"
                      class="px-1 py-0.2 rounded text-[7.5px] font-mono font-bold uppercase leading-none flex-shrink-0"
                      :class="param.alarm_status === 'CRITICAL'
                        ? 'bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/40 animate-pulse'
                        : 'bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/40'"
                      :title="'Active alarm: ' + param.alarm_status"
                    >
                      {{ param.alarm_status === 'CRITICAL' ? 'CRIT' : 'WARN' }}
                    </span>
                    <span
                      v-if="isParamDomainMismatched(param)"
                      class="px-1 py-0.2 rounded text-[7.5px] font-mono font-bold uppercase tracking-wider bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/40 cursor-help flex-shrink-0"
                      :title="'Category domain mismatch (' + (param.category ? param.category.domain : 'Other') + ' vs active ' + currentMonitoringDomain + '). Preserved safely.'"
                    >
                      MISMATCH
                    </span>
                    <!-- Data Quality Badge (Only shown when not GOOD, e.g. BAD, STALE, ALARM) -->
                    <span
                      v-if="param.quality && param.quality !== 'GOOD'"
                      class="px-1 py-0.2 rounded text-[7.5px] font-mono font-bold uppercase tracking-wider leading-none flex-shrink-0"
                      :class="qualityBadgeClass(param.quality)"
                    >
                      {{ param.quality }}
                    </span>
                    <!-- Pin Star Button -->
                    <button
                      type="button"
                      @click.stop="togglePin(param.param_id)"
                      class="text-[10px] transition hover:scale-110 p-0 leading-none flex-shrink-0"
                      :class="isPinned(param.param_id) ? 'text-amber-500 dark:text-amber-400' : 'text-slate-300 dark:text-slate-600 hover:text-amber-400'"
                      :title="isPinned(param.param_id) ? 'Unpin' : 'Pin to top'"
                    >
                      ★
                    </button>
                  </div>
                </div>

                <!-- Area 2: Main Content (Value + Unit side-by-side, Mini Sparkline on right) -->
                <div class="flex items-center justify-between min-w-0 gap-1.5 my-auto leading-none">
                  <!-- Left: Value and Unit -->
                  <div class="flex items-baseline min-w-0 leading-none truncate flex-1">
                    <span
                      class="font-bold font-mono tracking-tight whitespace-nowrap tabular-nums select-all leading-none"
                      style="font-variant-numeric: tabular-nums;"
                      :class="[paramValueColorClass(param), getParamValueFontSizeClass(param)]"
                      :title="formatParamValue(param)"
                    >
                      {{ formatParamValue(param) }}
                    </span>
                    <span
                      v-if="param.unit"
                      class="text-[9.5px] font-mono text-slate-500 dark:text-slate-400 font-semibold ml-1 flex-shrink-0 whitespace-nowrap leading-none"
                      :title="param.unit"
                    >
                      {{ param.unit }}
                    </span>
                    <span
                      v-if="param.is_held_value"
                      class="ml-1 px-1 py-0.2 rounded text-[7px] font-mono font-bold flex-shrink-0 leading-none bg-amber-500/15 text-amber-500"
                    >
                      HELD
                    </span>
                  </div>

                  <!-- Right: Mini Sparkline (Vertically centered, consistent 20px height) -->
                  <div class="w-[52px] sm:w-[56px] h-[20px] rounded bg-slate-50/80 dark:bg-[#070D17]/80 border border-slate-200/60 dark:border-slate-800/60 overflow-hidden flex-shrink-0 relative my-auto flex items-center justify-center">
                    <svg
                      v-if="getSparkline(param.param_id).hasHistory"
                      class="w-full h-full block"
                      viewBox="0 0 100 20"
                      preserveAspectRatio="none"
                    >
                      <defs>
                        <linearGradient :id="'sparkGrad-' + param.param_id" x1="0" y1="0" x2="0" y2="1">
                          <stop offset="0%" :stop-color="getSparklineColor(param)" stop-opacity="0.25" />
                          <stop offset="100%" :stop-color="getSparklineColor(param)" stop-opacity="0.0" />
                        </linearGradient>
                      </defs>
                      <path
                        v-for="(areaD, aIdx) in getSparkline(param.param_id).areaPaths"
                        :key="'a-' + param.param_id + '-' + aIdx"
                        :d="areaD"
                        :fill="'url(#sparkGrad-' + param.param_id + ')'"
                      />
                      <path
                        v-for="(strokeD, sIdx) in getSparkline(param.param_id).strokePaths"
                        :key="'s-' + param.param_id + '-' + sIdx"
                        :d="strokeD"
                        fill="none"
                        :stroke="getSparklineColor(param)"
                        stroke-width="1.5"
                        stroke-linecap="round"
                        stroke-linejoin="round"
                      />
                    </svg>
                    <!-- Empty state: Centered subtle indicator -->
                    <div
                      v-else
                      class="w-full h-full flex items-center justify-center font-mono text-[7.5px] text-slate-400 dark:text-slate-600 tracking-wider select-none px-0.5 text-center"
                    >
                      <span class="w-2.5 h-0.5 rounded-full bg-slate-300 dark:bg-slate-700 opacity-70"></span>
                    </div>
                  </div>
                </div>

                <!-- Area 3: Data Quality Issue / Stale / Offline Alert (Hidden on normal cards) -->
                <div
                  v-if="isParamDataQualityIssue(param)"
                  class="flex items-center justify-between min-w-0 gap-1 text-[8px] font-mono leading-none mt-auto pt-0.5 truncate"
                  :title="getParamTimingTitle(param)"
                >
                  <div
                    class="flex items-center space-x-0.5 font-bold truncate min-w-0"
                    :class="paramTimingColorClass(param)"
                  >
                    <span class="flex-shrink-0">{{ getParamTimingIcon(param) }}</span>
                    <span class="truncate">{{ getParamTimingLabel(param) }}</span>
                    <span v-if="param.received_at" class="opacity-80 whitespace-nowrap flex-shrink-0 ml-0.5">({{ formatRelative(param.received_at) }})</span>
                  </div>
                  <span
                    v-if="getParamQualityReasonText(param)"
                    class="text-[7.5px] px-1 py-0.2 rounded font-mono font-medium truncate flex-shrink-0"
                    :class="param.quality === 'BAD' || param.device_status === 'OFFLINE' ? 'bg-rose-500/15 text-rose-600 dark:text-rose-400' : 'bg-amber-500/15 text-amber-600 dark:text-amber-400'"
                  >
                    {{ getParamQualityReasonText(param) }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Empty State when filter has no matches -->
            <div
              v-else
              class="py-6 text-center text-slate-500 font-mono text-xs select-none"
            >
              {{ $t('operator.noParamsMatch') || 'No parameters matching current filter criteria.' }}
            </div>
          </div>
        </div>

        <div v-else class="scada-panel w-full bg-transparent border-0 shadow-none p-0">
          <!-- Empty State -->
          <div
            v-if="categorySections.length === 0 || filteredParameters.length === 0"
            class="scada-panel rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] p-8 text-center text-slate-500 font-mono text-xs"
          >
            {{ $t('operator.noParamsMatch') || 'No parameters matching current filter criteria.' }}
          </div>

          <!-- Multi-Table Side-by-Side Grid (No Vertical Scroll) -->
          <div
            v-else
            class="grid grid-cols-1 gap-3.5"
            :class="tableColumns.length > 1 ? 'xl:grid-cols-2' : ''"
          >
            <div
              v-for="(colParams, colIdx) in tableColumns"
              :key="'telemetry-table-col-' + colIdx"
              class="scada-panel rounded-xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] overflow-hidden shadow-xs dark:shadow-lg flex flex-col"
            >
              <div class="overflow-x-auto">
                <table class="w-full text-left text-xs font-mono border-collapse">
                  <thead class="bg-slate-50 dark:bg-[#080E18] text-3xs font-bold text-slate-600 dark:text-slate-400 uppercase tracking-wider border-b border-slate-200 dark:border-[#20354B]">
                    <tr>
                      <th class="py-2 px-1 text-center w-6">⭐</th>
                      <th class="py-2 px-1.5">{{ $t('telemetry.device') || 'Device' }}</th>
                      <th class="py-2 px-1.5">{{ $t('telemetry.parameterName') || 'Parameter' }}</th>
                      <th class="py-2 px-1.5">{{ $t('operator.categories') || 'Category' }}</th>
                      <th class="py-2 px-1.5 text-right">{{ $t('common.value') || 'Value' }}</th>
                      <th class="py-2 px-1">{{ $t('common.unit') || 'Unit' }}</th>
                      <th class="py-2 px-1 text-center">{{ $t('common.quality') || 'Quality' }}</th>
                      <th class="py-2 px-1.5 text-center">Status</th>
                      <th class="py-2 px-1.5 text-right">Last Received</th>
                      <th class="py-2 px-1.5 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-slate-100 dark:divide-[#18283E] text-2xs">
                    <tr
                      v-for="p in colParams"
                      :key="p.param_id"
                      @contextmenu.prevent="openParamContextMenu($event, p)"
                      class="transition-colors"
                      :class="[
                        isParamAlarmOrStaleBad(p)
                          ? (isAlarmPulseActive
                              ? 'bg-rose-500/25 dark:bg-rose-950/70 border-l-4 border-l-rose-500 transition-colors duration-150'
                              : 'bg-slate-200/70 dark:bg-slate-800/80 border-l-4 border-l-rose-500/70 transition-colors duration-300')
                          : 'hover:bg-slate-50 dark:hover:bg-[#101B2B]',
                        liveBlinks[p.param_id] ? 'bg-cyan-50 dark:bg-cyan-950/40' : '',
                        isPinned(p.param_id) ? 'bg-amber-50/40 dark:bg-[#0f1d30]/40' : ''
                      ]"
                    >
                      <!-- Pin -->
                      <td class="py-1.5 px-1 text-center">
                        <button @click="togglePin(p.param_id)" class="text-xs" :class="isPinned(p.param_id) ? 'text-amber-500 dark:text-amber-400' : 'text-slate-300 dark:text-slate-600 hover:text-amber-400'">
                          ★
                        </button>
                      </td>

                      <!-- Device -->
                      <td class="py-1.5 px-1.5 text-cyan-600 dark:text-cyan-400 font-bold whitespace-nowrap text-3xs">
                        <router-link :to="'/monitoring/devices/' + p.device_id" class="hover:underline">
                          {{ p.device_code }}
                        </router-link>
                      </td>

                      <!-- Parameter Name & Code -->
                      <td class="py-1.5 px-1.5 text-slate-700 dark:text-slate-200">
                        <div class="flex items-center space-x-1">
                          <ParameterIcon :icon="p.icon" :code="p.param_code" :name="p.param_name" :unit="p.unit" size-class="w-3.5 h-3.5 flex-shrink-0" />
                          <div class="min-w-0">
                            <div class="font-bold text-slate-900 dark:text-white leading-tight text-xs truncate max-w-[105px]" :title="p.param_name || p.param_code">
                              {{ p.param_name || p.param_code }}
                            </div>
                            <div class="text-4xs text-slate-500 dark:text-slate-400 font-mono truncate max-w-[90px]">{{ p.param_code }}</div>
                          </div>
                        </div>
                      </td>

                      <!-- Category Badge -->
                      <td class="py-1.5 px-1.5 whitespace-nowrap">
                        <span
                          v-if="p.category"
                          class="px-1.5 py-0.2 rounded text-4xs font-bold border font-mono"
                          :class="getCategoryTheme(p.category.color_key).pill"
                        >
                          {{ p.category.name }}
                        </span>
                        <span
                          v-else-if="categoryMap[p.category_id]"
                          class="px-1.5 py-0.2 rounded text-4xs font-bold border font-mono"
                          :class="getCategoryTheme(categoryMap[p.category_id].color_key).pill"
                        >
                          {{ categoryMap[p.category_id].name }}
                        </span>
                        <span
                          v-else
                          class="px-1.5 py-0.2 rounded text-4xs font-medium bg-slate-100 dark:bg-slate-800 text-slate-500 border border-slate-200 dark:border-slate-700/60 font-mono"
                        >
                          {{ $t('operator.uncategorized') || 'Uncategorized' }}
                        </span>
                        <span
                          v-if="isParamDomainMismatched(p)"
                          class="ml-1 px-1 py-0.2 rounded text-4xs font-bold font-mono bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/30"
                          :title="'Domain mismatch: Active domain is ' + currentMonitoringDomain"
                        >
                          MISMATCH
                        </span>
                      </td>

                      <!-- Value -->
                      <td class="py-1.5 px-1.5 text-right font-bold text-xs whitespace-nowrap" :class="paramValueColorClass(p)">
                        <span>{{ formatParamValue(p) }}</span>
                        <span v-if="p.is_held_value" class="text-4xs text-amber-500 dark:text-amber-400 ml-0.5">H</span>
                      </td>

                      <!-- Unit -->
                      <td class="py-1.5 px-1 text-slate-500 dark:text-slate-400 font-semibold text-4xs">{{ p.unit || '--' }}</td>

                      <!-- Quality -->
                      <td class="py-1.5 px-1 text-center">
                        <span class="px-1.5 py-0.2 rounded text-4xs font-bold uppercase inline-block" :class="qualityBadgeClass(p.quality)">
                          {{ p.quality || 'GOOD' }}
                        </span>
                      </td>

                      <!-- Threshold / Limit Status -->
                      <td class="py-1.5 px-1 text-center">
                        <span
                          class="px-1.5 py-0.2 rounded text-4xs font-bold"
                          :class="p.alarm_status === 'CRITICAL' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400' : p.alarm_status === 'WARNING' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'"
                        >
                          {{ p.alarm_status || 'NORMAL' }}
                        </span>
                      </td>

                      <!-- Last Received -->
                      <td class="py-1.5 px-1.5 text-right text-slate-500 dark:text-slate-400 text-3xs whitespace-nowrap">
                        <div>{{ p.received_at ? formatTimeShort(p.received_at) : '--' }}</div>
                        <div class="text-4xs text-slate-400 dark:text-slate-500">{{ p.received_at ? formatRelative(p.received_at) : '' }}</div>
                      </td>

                      <!-- Actions -->
                      <td class="py-1.5 px-1.5 text-right whitespace-nowrap space-x-1">
                        <button
                          @click="selectParamForTrend(p)"
                          class="px-1.5 py-0.5 rounded text-4xs font-bold transition"
                          :class="isParamSelectedForTrend(p) ? 'bg-cyan-500 text-slate-950 font-black' : 'bg-slate-100 hover:bg-slate-200 dark:bg-[#18283E] dark:hover:bg-[#20354B] text-cyan-700 dark:text-cyan-400'"
                        >
                          {{ isParamSelectedForTrend(p) ? 'PLOTTED' : 'PLOT' }}
                        </button>
                        <router-link
                          :to="'/monitoring/analysis?device_id=' + p.device_id + '&parameter_id=' + p.param_id"
                          class="px-1.5 py-0.5 rounded text-4xs bg-slate-100 hover:bg-slate-200 dark:bg-[#18283E] dark:hover:bg-[#20354B] text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white"
                          title="Inspect in Analysis"
                        >
                          FX
                        </router-link>
                        <button
                          @click.stop="openParamContextMenu($event, p)"
                          class="px-1.5 py-0.5 rounded text-4xs bg-slate-100 hover:bg-slate-200 dark:bg-[#18283E] dark:hover:bg-[#20354B] text-slate-700 dark:text-slate-300 hover:text-white font-bold"
                          title="Menu aksi parameter (klik kanan)"
                        >
                          ⋮
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </div>

        <!-- ISPU SCADA Indicator Bar (Permen LHK No. 14 Tahun 2020) - Placed directly below single parameter panel -->
        <ISPUIndicatorBar
          v-if="isAQMSDomain && isIspuBarVisible"
          :device-id="selectedDeviceFilter ? Number(selectedDeviceFilter) : ((devicesList[0] && devicesList[0].id) || 1)"
          :is-dark="isDark"
          :is-aqms-domain="isAQMSDomain"
        />
      </div>

    <!-- ======================================================== -->
    <!-- SECTION 2: OPERATIONAL ANALYTICAL SUITE (FULL WIDTH)      -->
    <!-- ======================================================== -->
    <div class="w-full flex flex-col space-y-1 pt-0.5">
      <!-- 1. TREND ANALYSIS CHART PANEL -->
        <div class="scada-panel px-2.5 py-1.5 rounded-lg bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#1E3048] shadow-xs flex flex-col space-y-1">
          <!-- Chart Header & Timeframe Buttons -->
          <div class="flex items-center justify-between gap-1.5 pb-1 border-b border-slate-200 dark:border-[#1E3048]">
            <div class="flex items-center space-x-1.5">
              <span class="w-2 h-2 rounded-full bg-cyan-500 shadow-glow-cyan animate-pulse"></span>
              <h3 class="text-2xs sm:text-xs font-bold tracking-wider uppercase text-slate-900 dark:text-white font-mono leading-none">
                {{ $t('operator.trendAnalysis') || 'TREND ANALYSIS' }}
              </h3>
              <span class="text-[8px] font-mono text-cyan-600 dark:text-cyan-400 bg-cyan-500/10 px-1.5 py-0.5 rounded border border-cyan-500/20 leading-none" title="Trend data automatically updates every 1 minute">
                AUTO 1M
              </span>

              <!-- Outlier Spike Alert Pill (When isolated spike detected) -->
              <button
                v-if="hasOutlierSpike && trendScaleMode === 'auto'"
                @click="toggleTrendSpikeZoom"
                class="inline-flex items-center space-x-1 px-1.5 py-0.2 rounded text-[8px] font-mono font-bold transition shadow-xs leading-none"
                :class="trendSpikeZoom === 'full'
                  ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/40 hover:bg-amber-500/25'
                  : 'bg-cyan-500/15 text-cyan-700 dark:text-cyan-300 border border-cyan-500/40 hover:bg-cyan-500/25'"
                :title="trendSpikeZoom === 'full' ? 'Click to focus normal telemetry range' : 'Click to expand to full spike range'"
              >
                <span>⚡</span>
                <span>{{ activeSpikeSummary }}</span>
                <span class="underline ml-0.5">[{{ trendSpikeZoom === 'full' ? 'Focus Normal' : 'Full Span' }}]</span>
              </button>
            </div>

            <!-- Timeframe Preset Buttons & Actions -->
            <div class="flex items-center space-x-1">
              <div class="h-[24px] flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded border border-slate-200 dark:border-[#1E3048] text-3xs font-mono leading-none">
                <button
                  v-for="tf in ['1m', '5m', '15m', '1h', '6h']"
                  :key="tf"
                  @click="setTimeframe(tf)"
                  class="h-full px-2 flex items-center justify-center rounded transition font-semibold leading-none cursor-pointer"
                  :class="currentTimeframe === tf
                    ? 'bg-cyan-500 text-slate-950 font-black shadow-xs'
                    : 'text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
                >
                  {{ tf }}
                </button>
              </div>

              <!-- Download Menu -->
              <div class="relative">
                <button
                  @click="showTrendDownloadMenu = !showTrendDownloadMenu"
                  class="w-[24px] h-[24px] flex items-center justify-center rounded border border-slate-200 dark:border-[#1E3048] bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-slate-600 dark:text-slate-400 hover:text-cyan-600 dark:hover:text-cyan-400 transition leading-none cursor-pointer"
                  title="Download Data / Chart"
                >
                  <OperatorDashboardIcon name="download" :size="14" />
                </button>
                <div
                  v-if="showTrendDownloadMenu"
                  class="absolute right-0 mt-1 w-44 bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#1E3048] rounded-xl shadow-xl py-1 z-40 text-3xs font-mono"
                >
                  <button @click="downloadTrendCSV(); showTrendDownloadMenu = false" class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer">
                    <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
                    <span>CSV Data</span>
                  </button>
                  <button @click="downloadTrendJSON(); showTrendDownloadMenu = false" class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer">
                    <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
                    <span>JSON Data</span>
                  </button>
                  <button @click="downloadTrendChart('png'); showTrendDownloadMenu = false" class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer">
                    <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
                    <span>PNG Graphic</span>
                  </button>
                  <button @click="downloadTrendChart('svg'); showTrendDownloadMenu = false" class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer">
                    <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
                    <span>SVG Graphic</span>
                  </button>
                </div>
              </div>

              <!-- Fullscreen Button -->
              <button
                @click="openTrendFullscreen"
                class="w-[24px] h-[24px] flex items-center justify-center rounded border border-slate-200 dark:border-[#1E3048] bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-slate-600 dark:text-slate-400 hover:text-cyan-600 dark:hover:text-cyan-400 transition leading-none cursor-pointer"
                title="Fullscreen / Expanded View"
              >
                <OperatorDashboardIcon name="fullscreen" :size="14" />
              </button>
            </div>
          </div>

          <!-- Series Selector & Toolbar Row -->
          <div class="flex items-center space-x-1 max-w-full text-3xs font-mono leading-none relative">
            <!-- Compact Toolbar Group: Chart Setup, All, Clear, Reset -->
            <div class="flex items-center space-x-1 shrink-0 mr-1 z-30">
              <!-- 1. Chart Setup (sliders-horizontal icon) with Unified Controls Popover -->
              <div class="relative trend-settings-anchor">
                <button
                  type="button"
                  @click="showTrendSettingsMenu = !showTrendSettingsMenu"
                  class="p-1 rounded border transition leading-none flex items-center justify-center cursor-pointer"
                  :class="showTrendSettingsMenu
                    ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-300 border-cyan-500 shadow-xs'
                    : 'border-slate-200 dark:border-[#1E3048] bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 hover:text-cyan-600 dark:hover:text-cyan-400'"
                  :title="$t('operator.chartSetup') || 'Chart Setup & Controls'"
                  aria-label="Chart Setup"
                >
                  <!-- Sliders-horizontal path: M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4 -->
                  <OperatorDashboardIcon name="settings" :size="13" />
                </button>

                <!-- Unified Controls Popover (Mode, Scale, Spike, Wind & Map, Fullscreen, Deep Analysis) -->
                <div
                  v-if="showTrendSettingsMenu"
                  class="absolute left-0 top-full mt-1.5 w-64 sm:w-72 bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#1E3048] rounded-xl shadow-2xl p-2.5 z-40 text-3xs font-mono select-none"
                  @click.stop
                >
                  <div class="flex items-center justify-between pb-1.5 mb-2 border-b border-slate-100 dark:border-[#1E3048]">
                    <span class="font-bold text-slate-900 dark:text-white uppercase tracking-wider flex items-center space-x-1.5 text-[10px]">
                      <OperatorDashboardIcon name="settings" :size="15" class="text-cyan-500" />
                      <span>Chart Setup & Controls</span>
                    </span>
                    <button
                      type="button"
                      @click="showTrendSettingsMenu = false"
                      class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition p-0.5 leading-none cursor-pointer"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="13" />
                    </button>
                  </div>

                  <div class="space-y-2">
                    <!-- 1. Mode: Processed vs Raw -->
                    <div>
                      <div class="text-[9px] text-slate-400 dark:text-slate-500 mb-1 font-semibold uppercase flex items-center justify-between">
                        <span>1. Mode</span>
                        <span class="text-[8px] font-bold text-cyan-600 dark:text-cyan-400 uppercase">{{ chartMode }}</span>
                      </div>
                      <div class="grid grid-cols-2 gap-1 p-0.5 rounded bg-slate-100 dark:bg-[#070E1A] border border-slate-200 dark:border-[#1E3048]">
                        <button
                          type="button"
                          @click="setChartMode('processed')"
                          class="py-1 rounded text-center font-bold transition text-[9px] cursor-pointer"
                          :class="chartMode === 'processed' ? 'bg-cyan-500 text-slate-950 shadow-xs' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'"
                        >
                          PROCESSED
                        </button>
                        <button
                          type="button"
                          @click="setChartMode('raw')"
                          class="py-1 rounded text-center font-bold transition text-[9px] cursor-pointer"
                          :class="chartMode === 'raw' ? 'bg-cyan-500 text-slate-950 shadow-xs' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'"
                        >
                          RAW SENSOR
                        </button>
                      </div>
                    </div>

                    <!-- 2. Scale: Actual Value vs Norm % -->
                    <div>
                      <div class="text-[9px] text-slate-400 dark:text-slate-500 mb-1 font-semibold uppercase flex items-center justify-between">
                        <span>2. Scale</span>
                        <span class="text-[8px] font-bold text-cyan-600 dark:text-cyan-400 uppercase">
                          {{ trendScaleMode === 'auto' ? (trendScales.isDual ? 'Dual-Axis' : 'Actual Value') : 'Norm %' }}
                        </span>
                      </div>
                      <div class="grid grid-cols-2 gap-1 p-0.5 rounded bg-slate-100 dark:bg-[#070E1A] border border-slate-200 dark:border-[#1E3048]">
                        <button
                          type="button"
                          @click="trendScaleMode = 'auto'"
                          class="py-1 rounded text-center font-bold transition text-[9px] cursor-pointer"
                          :class="trendScaleMode === 'auto' ? 'bg-cyan-500 text-slate-950 shadow-xs' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'"
                        >
                          {{ trendScales.isDual ? 'DUAL-AXIS' : 'ACTUAL VALUE' }}
                        </button>
                        <button
                          type="button"
                          @click="trendScaleMode = 'normalized'"
                          class="py-1 rounded text-center font-bold transition text-[9px] cursor-pointer"
                          :class="trendScaleMode === 'normalized' ? 'bg-cyan-500 text-slate-950 shadow-xs' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'"
                        >
                          NORM %
                        </button>
                      </div>
                    </div>

                    <!-- Spike Zoom (if spike present) -->
                    <div v-if="hasOutlierSpike && trendScaleMode === 'auto'">
                      <div class="text-[9px] text-slate-400 dark:text-slate-500 mb-1 font-semibold uppercase flex items-center justify-between">
                        <span>Spike Zoom</span>
                        <span class="text-[8px] font-bold text-amber-500 uppercase">{{ trendSpikeZoom }}</span>
                      </div>
                      <div class="grid grid-cols-2 gap-1 p-0.5 rounded bg-slate-100 dark:bg-[#070E1A] border border-slate-200 dark:border-[#1E3048]">
                        <button
                          type="button"
                          @click="trendSpikeZoom = 'robust'"
                          class="py-1 rounded text-center font-bold transition text-[9px] cursor-pointer"
                          :class="trendSpikeZoom === 'robust' ? 'bg-amber-500 text-slate-950 shadow-xs' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'"
                        >
                          SMART ROBUST
                        </button>
                        <button
                          type="button"
                          @click="trendSpikeZoom = 'full'"
                          class="py-1 rounded text-center font-bold transition text-[9px] cursor-pointer"
                          :class="trendSpikeZoom === 'full' ? 'bg-amber-500 text-slate-950 shadow-xs' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'"
                        >
                          FULL SPAN
                        </button>
                      </div>
                    </div>

                    <!-- 3. Show Wind & Map -->
                    <div class="pt-1.5 border-t border-slate-100 dark:border-[#1E3048] flex items-center justify-between">
                      <span class="text-slate-700 dark:text-slate-300 font-semibold text-[9px]">Show Wind & Map:</span>
                      <button
                        type="button"
                        @click="toggleEnvironmentalVisualizations"
                        class="px-2 py-0.5 rounded text-[8.5px] font-bold border transition cursor-pointer"
                        :class="showEnvironmentalVisualizations
                          ? 'bg-cyan-500/20 text-cyan-600 dark:text-cyan-400 border-cyan-500/40'
                          : 'bg-slate-100 dark:bg-[#101B2B] text-slate-500 border-slate-200 dark:border-[#1E3048]'"
                      >
                        {{ showEnvironmentalVisualizations ? 'VISIBLE' : 'HIDDEN' }}
                      </button>
                    </div>

                    <!-- 4. Expand Fullscreen & 5. Deep Analysis -->
                    <div class="pt-1.5 border-t border-slate-100 dark:border-[#1E3048] grid grid-cols-2 gap-1.5">
                      <button
                        type="button"
                        @click="openTrendFullscreen(); showTrendSettingsMenu = false"
                        class="py-1 px-1.5 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-[#1E3048] flex items-center justify-center space-x-1 font-bold text-center transition text-[8.5px] cursor-pointer"
                        title="Expand Fullscreen"
                      >
                        <OperatorDashboardIcon name="fullscreen" :size="12" class="text-cyan-500 shrink-0" />
                        <span>Fullscreen</span>
                      </button>

                      <router-link
                        to="/monitoring/analysis"
                        @click.native="showTrendSettingsMenu = false"
                        class="py-1 px-1.5 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-cyan-600 dark:text-cyan-400 border border-slate-200 dark:border-[#1E3048] flex items-center justify-center space-x-1 font-bold text-center transition text-[8.5px]"
                        title="Deep Telemetry Analysis"
                      >
                        <span>Deep Analysis →</span>
                      </router-link>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 2. All (check-check icon) -->
              <button
                type="button"
                @click="selectAllEligibleTrendSeries(true)"
                class="p-1 rounded border border-slate-200 dark:border-[#1E3048] bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-slate-600 dark:text-slate-300 hover:text-cyan-600 dark:hover:text-cyan-400 transition cursor-pointer flex items-center justify-center leading-none"
                title="Select all eligible parameters (All)"
                aria-label="All"
              >
                <OperatorDashboardIcon name="check-all" :size="12" />
              </button>

              <!-- 3. Clear (x-circle icon) -->
              <button
                v-if="activeTrendSeries.length > 0"
                type="button"
                @click="clearAllTrendSeries"
                class="p-1 rounded border border-slate-200 dark:border-[#1E3048] bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-slate-600 dark:text-slate-300 hover:text-rose-500 dark:hover:text-rose-400 transition cursor-pointer flex items-center justify-center leading-none"
                title="Clear all series (Clear)"
                aria-label="Clear"
              >
                <OperatorDashboardIcon name="clear" :size="12" />
              </button>

              <!-- 4. Reset (rotate-ccw icon) -->
              <button
                type="button"
                @click="resetTrendSelectionToDefault"
                class="p-1 rounded border border-slate-200 dark:border-[#1E3048] bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-slate-600 dark:text-slate-300 hover:text-amber-500 dark:hover:text-amber-400 transition cursor-pointer flex items-center justify-center leading-none"
                title="Reset to default (Reset)"
                aria-label="Reset"
              >
                <OperatorDashboardIcon name="reset" :size="12" />
              </button>

              <!-- Series Count Badge -->
              <span class="text-slate-400 dark:text-slate-500 text-[8.5px] font-semibold pl-0.5">({{ activeTrendSeries.length }}):</span>
            </div>

            <!-- Compact Series Chips -->
            <div class="flex items-center space-x-1 overflow-x-auto no-scrollbar py-0 flex-1 min-w-0">
              <div
                v-for="s in activeTrendSeries"
                :key="s.id"
                @mouseenter="hoveredTrendSeriesId = s.id"
                @mouseleave="hoveredTrendSeriesId = null"
                class="flex items-center space-x-1 px-1.5 py-0.5 rounded border shrink-0 bg-slate-100/90 dark:bg-[#101B2B] text-slate-800 dark:text-slate-200 shadow-xs transition cursor-pointer text-[9px] select-none"
                :class="{ 'ring-1 ring-cyan-500/50 shadow-sm': hoveredTrendSeriesId === s.id }"
                :style="{ borderColor: s.color + '55' }"
                :title="(s.deviceCode ? s.deviceCode + ' · ' : '') + (s.paramName || s.paramCode) + (s.latestVal != null && s.latestVal !== '--' ? ' (' + s.latestVal + ' ' + (s.unit || '') + ')' : '')"
              >
                <!-- Series Color Indicator Dot -->
                <span class="w-1.5 h-1.5 rounded-full shrink-0 shadow-xs" :style="{ backgroundColor: s.color }"></span>

                <!-- Compact Parameter Display Name -->
                <span class="font-bold text-slate-900 dark:text-white tracking-tight">{{ getCompactTrendParamName(s) || (s.paramName || s.paramCode) }}</span>
                <span v-if="hasDuplicateTrendParamName(s)" class="text-slate-500 dark:text-slate-400 text-[7.5px] font-medium ml-0.5">({{ s.deviceCode }})</span>

                <!-- Quick Remove Action -->
                <button
                  @click.stop="removeTrendSeries(s.id)"
                  class="text-slate-400 hover:text-rose-500 dark:hover:text-rose-400 ml-0.5 transition p-0.2 leading-none cursor-pointer"
                  title="Remove series"
                >
                  ✕
                </button>
              </div>

              <button
                v-if="activeTrendSeries.length === 0"
                @click="resetTrendSelectionToDefault"
                class="text-cyan-600 dark:text-cyan-400 text-3xs hover:underline shrink-0"
              >
                + Click to restore all parameters or click any parameter card to plot
              </button>
            </div>
          </div>

          <!-- SVG Industrial Chart Canvas (Height: ~96-102px) -->
          <div
            ref="scadaChartContainer"
            class="w-full h-[160px] sm:h-[185px] lg:h-[210px] bg-slate-50/80 dark:bg-[#061321] rounded-md border border-[#B7D8EC] dark:border-[#244356] relative overflow-hidden select-none /* h-52 sm:h-56 */"
            @mousemove="onChartScrub"
            @mouseleave="scrubPoint = null"
          >
            <!-- Loading Indicator -->
            <div v-if="isChartLoading" class="absolute inset-0 bg-white/80 dark:bg-[#061321]/85 backdrop-blur-xs flex items-center justify-center text-cyan-600 dark:text-[#21C7E8] text-xs font-mono z-20">
              <span class="animate-spin w-4 h-4 border-2 border-cyan-500 border-t-transparent rounded-full mr-2"></span>
              Loading Telemetry Trend...
            </div>

            <!-- Empty State -->
            <div v-else-if="activeTrendSeries.length === 0" class="w-full h-full flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 text-3xs font-mono space-y-1">
              <span>Select parameters from left monitoring tiles to display trend graph</span>
            </div>

            <!-- SVG Multi-Series Trend Line -->
            <svg v-else ref="trendPanelSvg" class="w-full h-full block" :viewBox="`0 0 ${trendChartDims.width} ${trendChartDims.height}`">
              <defs>
                <!-- Subtle vertical gradient for Primary Series only -->
                <linearGradient id="scadaPrimaryFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" :stop-color="primarySeriesColor" stop-opacity="0.08" />
                  <stop offset="100%" :stop-color="primarySeriesColor" stop-opacity="0.00" />
                </linearGradient>
              </defs>

              <!-- Y-Axis Grid Lines & Values -->
              <g class="grid-lines" :stroke="gridStroke" stroke-dasharray="3 3" stroke-width="0.8">
                <g v-for="(tick, idx) in trendYTicks" :key="'ytick-' + idx">
                  <line :x1="trendChartDims.padLeft" :y1="tick.y" :x2="trendChartDims.width - trendChartDims.padRight" :y2="tick.y" stroke-opacity="0.75" />
                  <!-- Left Y-Axis Tick Label (~10-11px readable, proportional typography) -->
                  <text
                    :x="trendChartDims.padLeft - 6"
                    :y="tick.y + 3.5"
                    text-anchor="end"
                    font-size="10"
                    font-family="monospace"
                    :fill="isDark ? '#9BB4C8' : '#475569'"
                  >{{ tick.label }}</text>
                </g>

                <!-- Dual Y-Axis Right Ticks (when active series have incompatible units) -->
                <template v-if="trendScales.isDual">
                  <g v-for="(tick, idx) in trendRightYTicks" :key="'rytick-' + idx">
                    <text
                      :x="trendChartDims.width - trendChartDims.padRight + 6"
                      :y="tick.y + 3.5"
                      text-anchor="start"
                      font-size="10"
                      font-family="monospace"
                      :fill="trendScales.right ? trendScales.right.color : (isDark ? '#9BB4C8' : '#475569')"
                    >{{ tick.label }}</text>
                  </g>
                </template>

                <!-- Base Axis Line -->
                <line :x1="trendChartDims.padLeft" :y1="trendChartDims.baseY" :x2="trendChartDims.width - trendChartDims.padRight" :y2="trendChartDims.baseY" stroke-dasharray="none" :stroke="axisStroke" stroke-width="1.2" />
              </g>

              <!-- Axis Unit Labels at top of axis -->
              <g class="axis-units">
                <!-- Left Axis Unit Badge -->
                <text
                  :x="trendChartDims.padLeft"
                  y="16"
                  font-size="10.5"
                  font-family="monospace"
                  font-weight="bold"
                  :fill="trendScales.left ? trendScales.left.color : (isDark ? '#21C7E8' : '#0284C7')"
                >
                  {{ trendScales.left ? trendScales.left.label : '' }}
                </text>

                <!-- Right Axis Unit Badge (when dual axis) -->
                <text
                  v-if="trendScales.isDual && trendScales.right"
                  :x="trendChartDims.width - trendChartDims.padRight"
                  y="16"
                  font-size="10.5"
                  font-family="monospace"
                  font-weight="bold"
                  text-anchor="end"
                  :fill="trendScales.right.color"
                >
                  {{ trendScales.right.label }}
                </text>
              </g>

              <!-- X-Axis Time Ticks & Timestamps -->
              <g class="time-ticks">
                <g v-for="(tick, idx) in trendXTicks" :key="'xtick-' + idx">
                  <line :x1="tick.x" :y1="trendChartDims.baseY" :x2="tick.x" :y2="trendChartDims.baseY + 3.5" :stroke="axisStroke" stroke-width="1" />
                  <text
                    :x="tick.x"
                    :y="trendChartDims.baseY + 15"
                    text-anchor="middle"
                    font-size="9.5"
                    font-family="monospace"
                    :fill="isDark ? '#9BB4C8' : '#475569'"
                  >{{ tick.label }}</text>
                </g>
              </g>

              <!-- Threshold Reference Lines (Configured limits with clear parameter labels) -->
              <g v-if="trendThresholdLines.crit != null">
                <line :x1="trendChartDims.padLeft" :y1="trendThresholdLines.crit.y" :x2="trendChartDims.width - trendChartDims.padRight" :y2="trendThresholdLines.crit.y" stroke="#EF4444" stroke-width="1.2" stroke-dasharray="4 4" opacity="0.85" />
                <text :x="trendChartDims.width - trendChartDims.padRight - 4" :y="trendThresholdLines.crit.y - 4" fill="#EF4444" font-size="9" text-anchor="end" font-family="monospace" font-weight="bold">
                  {{ trendThresholdLines.crit.label }}
                </text>
              </g>
              <g v-if="trendThresholdLines.warn != null">
                <line :x1="trendChartDims.padLeft" :y1="trendThresholdLines.warn.y" :x2="trendChartDims.width - trendChartDims.padRight" :y2="trendThresholdLines.warn.y" stroke="#F59E0B" stroke-width="1.2" stroke-dasharray="4 4" opacity="0.85" />
                <text :x="trendChartDims.width - trendChartDims.padRight - 4" :y="trendThresholdLines.warn.y - 4" fill="#F59E0B" font-size="9" text-anchor="end" font-family="monospace" font-weight="bold">
                  {{ trendThresholdLines.warn.label }}
                </text>
              </g>

              <!-- Outlier Spike Markers (When an isolated spike exceeds the focused scale) -->
              <g v-if="trendOutlierSpikes.length > 0" class="outlier-spikes">
                <g
                  v-for="(spike, spkIdx) in trendOutlierSpikes"
                  :key="'outlier-' + spkIdx"
                  class="cursor-pointer"
                  @click="toggleTrendSpikeZoom"
                >
                  <line
                    :x1="spike.px"
                    :y1="trendChartDims.padTop"
                    :x2="spike.px"
                    :y2="trendChartDims.baseY"
                    :stroke="spike.color"
                    stroke-width="1"
                    stroke-dasharray="2 2"
                    opacity="0.45"
                  />
                  <polygon
                    :points="`${spike.px - 4},${trendChartDims.padTop} ${spike.px + 4},${trendChartDims.padTop} ${spike.px},${trendChartDims.padTop + 6}`"
                    :fill="spike.color"
                  />
                  <rect
                    :x="Math.min(Math.max(spike.px - 32, trendChartDims.padLeft), trendChartDims.width - trendChartDims.padRight - 64)"
                    y="12"
                    width="64"
                    height="16"
                    rx="3"
                    :fill="isDark ? '#0B1422' : '#FFFFFF'"
                    :stroke="spike.color"
                    stroke-width="1"
                  />
                  <text
                    :x="Math.min(Math.max(spike.px, trendChartDims.padLeft + 32), trendChartDims.width - trendChartDims.padRight - 32)"
                    y="23.5"
                    text-anchor="middle"
                    font-size="9"
                    font-family="monospace"
                    font-weight="bold"
                    :fill="spike.color"
                  >
                    ▲ {{ spike.displayVal }}
                  </text>
                </g>
              </g>

              <!-- Primary Series Subtle Area Gradient Fill (Single primary series only) -->
              <path
                v-if="activeTrendSeries.length > 0 && generatePrimaryAreaPath()"
                :d="generatePrimaryAreaPath()"
                fill="url(#scadaPrimaryFill)"
              />

              <!-- Series Paths (Unfilled lines by default, clean 1.8px strokes with data gap handling) -->
              <g v-for="(s, sIdx) in activeTrendSeries" :key="'plot-' + s.id">
                <path
                  :d="generateChartPath(s)"
                  fill="none"
                  :stroke="s.color"
                  :stroke-width="hoveredTrendSeriesId === s.id ? 2.2 : (sIdx === 0 ? 1.8 : 1.6)"
                  :opacity="hoveredTrendSeriesId && hoveredTrendSeriesId !== s.id ? 0.35 : 1"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="transition-all duration-150 cursor-pointer"
                  @mouseenter="hoveredTrendSeriesId = s.id"
                  @mouseleave="hoveredTrendSeriesId = null"
                />
              </g>

              <!-- Scrubbing Hover Crosshair & Multi-Series Value Points -->
              <g v-if="scrubPoint">
                <line
                  :x1="scrubPoint.x"
                  :y1="trendChartDims.padTop"
                  :x2="scrubPoint.x"
                  :y2="trendChartDims.baseY"
                  stroke="#00E5FF"
                  stroke-width="1.2"
                  stroke-dasharray="2 2"
                  opacity="0.8"
                />
                <!-- Points for all active series at scrub X -->
                <g v-for="sp in scrubPoint.allSeries" :key="'scrub-dot-' + sp.seriesId">
                  <circle
                    v-if="sp.plotY != null"
                    :cx="scrubPoint.x"
                    :cy="sp.plotY"
                    r="3.5"
                    :fill="sp.color"
                    :stroke="isDark ? '#060B13' : '#FFFFFF'"
                    stroke-width="1.5"
                  />
                </g>
              </g>
            </svg>

            <!-- Floating Hover Tooltip (Dark Glassmorphic SCADA card with complete metadata) -->
            <div
              v-if="scrubPoint"
              class="absolute z-30 pointer-events-none bg-white/95 dark:bg-[#080E18]/95 backdrop-blur-md border border-cyan-500/60 rounded-xl p-3 shadow-2xl text-2xs font-mono text-slate-900 dark:text-white transition-all duration-75 min-w-[210px] max-w-[290px]"
              :style="{ left: tooltipLeftPos + 'px', top: '10px' }"
            >
              <div class="flex items-center justify-between text-slate-400 dark:text-slate-400 text-3xs border-b border-slate-200 dark:border-[#1E3048] pb-1.5 mb-2 font-mono">
                <span class="flex items-center space-x-1">
                  <span>🕒</span>
                  <span class="font-bold text-slate-700 dark:text-slate-200">{{ scrubPoint.date }}</span>
                  <span>{{ scrubPoint.time }}</span>
                </span>
                <span class="px-1.5 py-0.2 rounded text-4xs font-bold bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30">
                  {{ currentTimeframe }}
                </span>
              </div>

              <div class="space-y-1.5">
                <div
                  v-for="sp in scrubPoint.allSeries"
                  :key="sp.seriesId"
                  class="flex items-center justify-between space-x-2 text-2xs"
                >
                  <div class="flex items-center space-x-1.5 min-w-0">
                    <span class="w-2.5 h-2.5 rounded-full shrink-0 shadow-xs" :style="{ backgroundColor: sp.color }"></span>
                    <span class="text-slate-500 dark:text-slate-400 text-3xs shrink-0">{{ sp.deviceCode }} ·</span>
                    <span class="font-bold text-slate-800 dark:text-slate-100 truncate">{{ sp.paramName || sp.paramCode }}</span>
                  </div>
                  <div class="flex items-center space-x-1 shrink-0 font-mono">
                    <span class="font-black tabular-nums" :style="{ color: sp.color }">
                      {{ sp.val }} <span class="text-3xs font-medium text-slate-500 dark:text-slate-400">{{ sp.unit }}</span>
                    </span>
                    <span
                      v-if="sp.quality && sp.quality !== 'GOOD'"
                      class="px-1 py-0.2 rounded text-4xs font-bold"
                      :class="sp.quality === 'BAD' ? 'bg-rose-500/20 text-rose-400 border border-rose-500/30' : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'"
                    >
                      {{ sp.quality }}
                    </span>
                    <span
                      v-if="sp.alarmStatus && sp.alarmStatus !== 'NORMAL'"
                      class="px-1 py-0.2 rounded text-4xs font-bold"
                      :class="sp.alarmStatus === 'CRITICAL' ? 'bg-red-500/20 text-red-400 border border-red-500/30' : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'"
                    >
                      {{ sp.alarmStatus === 'CRITICAL' ? 'CRIT' : 'WARN' }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. OPERATIONAL ENVIRONMENTAL VISUALIZATIONS (WIND ROSE & POLLUTANT DISPERSION MAP) -->
        <div v-if="showEnvironmentalVisualizations" class="grid grid-cols-1 lg:grid-cols-12 gap-2 flex-1">
          <!-- Left: Wind Rose (approx 40% -> 5 cols out of 12) -->
          <div class="lg:col-span-5 min-h-[300px] flex flex-col">
            <WindRoseChart
              :devices-list="devicesList"
              :is-dark="isDark"
              :live-map="liveMap"
            />
          </div>

          <!-- Right: Pollutant Dispersion Map (approx 60% -> 7 cols out of 12) -->
          <div class="lg:col-span-7 min-h-[300px] flex flex-col">
            <PollutantDispersionMap
              :devices-list="devicesList"
              :is-dark="isDark"
              :live-map="liveMap"
            />
          </div>
        </div>
      </div>

    <!-- I. QUICK ACTIONS & EXPORT BAR (Docked SCADA Footer) -->
    <div class="fixed bottom-0 left-0 right-0 z-30 bg-white/95 dark:bg-[#0B1422]/95 backdrop-blur-md border-t border-slate-200 dark:border-[#20354B] px-2 sm:px-3 py-0.5 flex items-center justify-between gap-1 shadow-lg transition-colors h-[28px] select-none">
      <!-- Left: Quick Operational Commands -->
      <div class="flex items-center gap-1 sm:gap-1.5 flex-shrink-0">
        <span class="text-[8.5px] uppercase font-mono font-bold text-cyan-600 dark:text-cyan-400 hidden sm:inline mr-0.5 tracking-wider leading-none">
          {{ $t('operator.quickActions') || 'QUICK ACTIONS' }}:
        </span>

        <!-- Export Data -->
        <button
          @click="openExportModal"
          class="h-[22px] px-2 rounded bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-[9.5px] font-mono transition-colors shadow-xs flex items-center space-x-1 leading-none cursor-pointer"
        >
          <OperatorDashboardIcon name="export" :size="12" class="text-slate-950 flex-shrink-0" />
          <span>{{ $t('operator.exportData') || 'Export Data' }}</span>
        </button>

        <!-- Generate Report -->
        <button
          @click="openReportModal"
          class="h-[22px] px-2 rounded bg-blue-600 hover:bg-blue-500 text-white font-bold text-[9.5px] font-mono transition-colors flex items-center space-x-1 shadow-xs leading-none cursor-pointer"
        >
          <OperatorDashboardIcon name="report" :size="12" class="text-white flex-shrink-0" />
          <span>{{ $t('operator.generateReport') || 'Operational Report' }}</span>
        </button>

        <!-- Alarm History -->
        <router-link
          to="/alarm/history"
          class="h-[22px] px-1.5 sm:px-2 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-[#20354B] text-[9.5px] font-mono font-medium transition hidden md:flex items-center space-x-1 leading-none"
        >
          <OperatorDashboardIcon name="history" :size="12" class="text-slate-600 dark:text-slate-300" />
          <span>{{ $t('operator.alarmHistory') || 'Alarm History' }}</span>
        </router-link>
      </div>

      <!-- Right: Navigation Shortcuts & Audio State -->
      <div class="flex items-center gap-1 sm:gap-1.5 flex-shrink-0 text-[9.5px] font-mono leading-none">
        <router-link
          to="/monitoring/devices"
          class="h-[22px] px-1.5 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-300 border border-slate-300 dark:border-[#20354B] font-medium transition hidden lg:flex items-center"
        >
          {{ $t('operator.deviceMonitoring') || 'Devices' }}
        </router-link>
        <router-link
          to="/monitoring/analysis"
          class="h-[22px] px-1.5 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-300 border border-slate-300 dark:border-[#20354B] font-medium transition hidden lg:flex items-center"
        >
          {{ $t('operator.telemetryAnalysis') || 'Analysis' }}
        </router-link>
        <router-link
          to="/monitoring/system"
          class="h-[22px] px-1.5 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-300 border border-slate-300 dark:border-[#20354B] font-medium transition hidden lg:flex items-center"
        >
          {{ $t('operator.systemHealthAction') || 'System Health' }}
        </router-link>

        <!-- Audio Indicator Badge / Toggle & Setup Control -->
        <div class="relative audio-setup-anchor flex items-center">
          <button
            type="button"
            @click="toggleAudioSetupPopover"
            class="h-[22px] px-2 rounded border text-[9px] font-mono font-bold flex items-center space-x-1 leading-none transition-colors cursor-pointer"
            :class="isAudioMuted || !isSoundEnabled ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30' : 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border-cyan-500/30'"
            :title="isAudioMuted ? 'Audio Muted · Click for Setup & Unmute' : 'Audio Active · Click for Setup & Volume'"
          >
            <OperatorDashboardIcon :name="isAudioMuted ? 'sound-muted' : (isSoundEnabled ? 'sound-on' : 'sound-off')" :size="12" />
            <span>{{ isAudioMuted ? 'MUTED' : (isSoundEnabled ? 'SOUND ON' : 'SOUND OFF') }}</span>
            <span class="text-slate-400 dark:text-slate-500 hover:text-cyan-500 dark:hover:text-cyan-300 ml-0.5 text-[8px]">⚙️</span>
          </button>

          <!-- Upward Docked Audio Setup Popover -->
          <div
            v-if="showAudioSetupPopover"
            class="absolute right-0 bottom-full mb-2 z-50 w-72 sm:w-80 rounded-xl shadow-2xl p-3 border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#0B1422] text-slate-800 dark:text-slate-100 select-none font-mono text-xs animate-fadeIn"
            @click.stop
          >
            <!-- Popover Header -->
            <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
              <div class="flex items-center space-x-1.5">
                <span class="text-cyan-500 text-sm">🔊</span>
                <span class="text-xs font-bold uppercase tracking-wider font-mono">
                  Audio & Alarm Setup
                </span>
              </div>
              <button
                type="button"
                @click="showAudioSetupPopover = false"
                class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition cursor-pointer"
              >
                ✕
              </button>
            </div>

            <!-- 1. Master Status & Quick Mute Toggle -->
            <div class="flex items-center justify-between p-2 rounded bg-slate-50 dark:bg-[#0E1726] border border-slate-200 dark:border-[#20354B] mb-2.5">
              <div>
                <div class="text-[10.5px] font-bold text-slate-900 dark:text-white flex items-center space-x-1.5">
                  <span>{{ isAudioMuted ? '🔇 Audio Muted' : '🔊 Audio Active' }}</span>
                </div>
                <div class="text-[9px] text-slate-500 dark:text-slate-400">
                  {{ isAudioMuted ? 'Alarm chimes currently silenced' : 'Alarm chimes will sound on incidents' }}
                </div>
              </div>
              <button
                type="button"
                @click="toggleAudioMute"
                class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                :class="!isAudioMuted ? 'bg-cyan-500' : 'bg-slate-300 dark:bg-slate-700'"
                role="switch"
                :aria-checked="!isAudioMuted"
              >
                <span
                  class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                  :class="!isAudioMuted ? 'translate-x-4' : 'translate-x-0'"
                ></span>
              </button>
            </div>

            <!-- 2. Volume Slider -->
            <div class="p-2.5 rounded bg-slate-50 dark:bg-[#0E1726] border border-slate-200 dark:border-[#20354B] mb-2.5 space-y-1.5">
              <div class="flex items-center justify-between text-[10px]">
                <span class="font-bold text-slate-700 dark:text-slate-300 flex items-center space-x-1">
                  <span>🔈</span>
                  <span>Alert Volume</span>
                </span>
                <span class="font-mono font-bold text-cyan-600 dark:text-cyan-400">
                  {{ Math.round(audioVolume * 100) }}%
                </span>
              </div>
              <input
                type="range"
                min="0"
                max="1"
                step="0.05"
                v-model.number="audioVolume"
                @input="onAudioVolumeChange"
                class="w-full h-1.5 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-cyan-500"
              />
              <div class="flex justify-between text-[8px] text-slate-400 font-mono">
                <span>0% (Silent)</span>
                <span>50%</span>
                <span>100%</span>
              </div>
            </div>

            <!-- 3. Test Audio Chimes (Audio Pattern Preview) -->
            <div class="p-2.5 rounded bg-slate-50 dark:bg-[#0E1726] border border-slate-200 dark:border-[#20354B] mb-2.5 space-y-1.5">
              <div class="text-[9px] font-bold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                Test Audio Chimes
              </div>
              <div class="grid grid-cols-2 gap-1.5">
                <button
                  type="button"
                  @click="testAlarmChime('WARNING')"
                  class="px-2 py-1.5 rounded border border-amber-500/30 bg-amber-500/10 hover:bg-amber-500/20 text-amber-600 dark:text-amber-300 font-bold text-[9.5px] transition flex items-center justify-center space-x-1 cursor-pointer active:scale-95"
                >
                  <span>🔔</span>
                  <span>Warning</span>
                </button>
                <button
                  type="button"
                  @click="testAlarmChime('CRITICAL')"
                  class="px-2 py-1.5 rounded border border-rose-500/30 bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-300 font-bold text-[9.5px] transition flex items-center justify-center space-x-1 cursor-pointer active:scale-95"
                >
                  <span>🚨</span>
                  <span>Critical</span>
                </button>
              </div>
            </div>

            <!-- 4. Repeat Policy & Mute Duration Grid -->
            <div class="grid grid-cols-2 gap-2 mb-2 text-[9.5px]">
              <div>
                <label class="block text-[8.5px] font-bold text-slate-500 dark:text-slate-400 uppercase mb-1">
                  Repeat Interval
                </label>
                <select
                  v-model.number="audioRepeatInterval"
                  @change="onAudioRepeatIntervalChange"
                  class="w-full px-1.5 py-1 rounded bg-white dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-[9.5px] focus:outline-none focus:border-cyan-500"
                >
                  <option :value="1">Every 1s</option>
                  <option :value="3">Every 3s</option>
                  <option :value="5">Every 5s (Default)</option>
                  <option :value="10">Every 10s</option>
                  <option :value="15">Every 15s</option>
                  <option :value="30">Every 30s</option>
                  <option :value="0">Once Only</option>
                </select>
              </div>
              <div>
                <label class="block text-[8.5px] font-bold text-slate-500 dark:text-slate-400 uppercase mb-1">
                  Mute Timer
                </label>
                <select
                  v-model.number="audioMuteDuration"
                  @change="onAudioMuteDurationChange"
                  class="w-full px-1.5 py-1 rounded bg-white dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-[9.5px] focus:outline-none focus:border-cyan-500"
                >
                  <option :value="0">Normal (Unmuted)</option>
                  <option :value="5">Mute 5m</option>
                  <option :value="15">Mute 15m</option>
                  <option :value="30">Mute 30m</option>
                  <option :value="60">Mute 1h</option>
                  <option :value="-1">Mute Indefinite</option>
                </select>
              </div>
            </div>

            <!-- Footer Link: Advanced Notification Settings Modal Trigger -->
            <div class="pt-2 border-t border-slate-100 dark:border-[#1E3048]">
              <button
                type="button"
                @click="openFullAlarmSettingsModal"
                class="w-full py-1.5 px-2 rounded text-[9.5px] font-semibold transition border flex items-center justify-center space-x-1.5 bg-slate-50 dark:bg-[#0E1726] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400 cursor-pointer"
              >
                <span>⚙️</span>
                <span>Open Full Notification Settings</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- MODAL: DUAL ORGANIZATION BRANDING SETUP (CUSTOMER & MAKER) -->
    <div
      v-if="showBrandingModal || showLogoModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-slate-950/80 backdrop-blur-sm"
      @click.self="closeBrandingModal"
    >
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] p-5 shadow-2xl text-slate-900 dark:text-slate-100 space-y-4 font-mono text-xs">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-[#20354B] pb-3">
          <div class="flex items-center space-x-2">
            <span class="text-cyan-500 text-sm">🏢</span>
            <h3 class="font-bold text-slate-900 dark:text-white uppercase tracking-wider text-xs">
              {{ $t('branding.title') || 'Organization & Company Branding' }}
            </h3>
          </div>
          <button @click="closeBrandingModal" class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition">✕</button>
        </div>

        <!-- Error Message Banner -->
        <div v-if="brandingErrorMsg || logoErrorMsg" class="p-2.5 rounded-lg bg-rose-500/10 border border-rose-500/30 text-rose-600 dark:text-rose-400 text-3xs font-semibold flex items-center space-x-1.5">
          <span>⚠️</span>
          <span>{{ brandingErrorMsg || logoErrorMsg }}</span>
        </div>

        <!-- Tab Selection (Customer Left Header vs Maker Right Header) -->
        <div class="grid grid-cols-2 gap-1 p-1 bg-slate-100 dark:bg-[#101B2B] rounded-lg border border-slate-200 dark:border-[#20354B]">
          <button
            type="button"
            @click="activeBrandingTab = 'customer'"
            class="py-1.5 px-3 rounded text-3xs font-bold transition flex items-center justify-center space-x-1.5"
            :class="activeBrandingTab === 'customer'
              ? 'bg-cyan-600 text-white shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            <span>🏢</span>
            <span>{{ $t('branding.customerTab') || 'Customer (Left Header)' }}</span>
          </button>
          <button
            type="button"
            @click="activeBrandingTab = 'maker'"
            class="py-1.5 px-3 rounded text-3xs font-bold transition flex items-center justify-center space-x-1.5"
            :class="activeBrandingTab === 'maker'
              ? 'bg-cyan-600 text-white shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            <span>⚙️</span>
            <span>{{ $t('branding.makerTab') || 'Maker / Provider (Right Header)' }}</span>
          </button>
        </div>

        <!-- TAB 1: CUSTOMER COMPANY SETUP -->
        <div v-if="activeBrandingTab === 'customer'" class="space-y-3">
          <div class="space-y-1">
            <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
              {{ $t('branding.customerName') || 'Customer Company Name (Appears Top-Left)' }}
            </label>
            <input
              v-model="modalCustomerName"
              type="text"
              maxlength="50"
              placeholder="e.g. PT. Mitra Industri Enviro"
              class="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] bg-slate-50 dark:bg-[#101B2B] text-slate-900 dark:text-slate-100 text-xs focus:ring-1 focus:ring-cyan-500 focus:outline-none"
            />
          </div>

          <div class="space-y-1">
            <div class="flex items-center justify-between">
              <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
                {{ $t('branding.stationDomain') || 'Jenis Stasiun / Domain Pemantauan' }}
              </label>
              <div class="flex items-center space-x-1">
                <button
                  type="button"
                  v-for="st in ['AQMS', 'WQMS', 'Sparing']"
                  :key="st"
                  @click="modalCustomerSubtitle = st"
                  class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold transition border cursor-pointer"
                  :class="modalCustomerSubtitle === st
                    ? 'bg-cyan-500 text-slate-950 border-cyan-400 shadow-xs'
                    : 'bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:text-cyan-500'"
                >
                  {{ st }}
                </button>
              </div>
            </div>
            <input
              v-model="modalCustomerSubtitle"
              type="text"
              maxlength="50"
              placeholder="e.g. AQMS, WQMS, Sparing"
              class="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] bg-slate-50 dark:bg-[#101B2B] text-slate-900 dark:text-slate-100 text-xs focus:ring-1 focus:ring-cyan-500 focus:outline-none font-mono"
            />
          </div>

          <!-- Customer Logo Live Preview (Dark & Light) -->
          <div class="space-y-1.5">
            <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
              Customer Logo Preview (Header Scale)
            </label>
            <div class="grid grid-cols-2 gap-2">
              <div class="p-3 rounded-lg bg-[#0B1422] border border-[#20354B] flex flex-col items-center justify-center min-h-[64px]">
                <span class="text-[8px] text-slate-400 uppercase tracking-widest mb-1">Dark Header</span>
                <img
                  v-if="modalCustomerLogoPreviewUrl || customerLogoUrl"
                  :src="modalCustomerLogoPreviewUrl || customerLogoUrl"
                  alt="Customer Logo preview"
                  class="max-h-7 max-w-[120px] object-contain"
                />
                <span v-else class="text-[9px] text-slate-500 italic">Default SCADA Icon</span>
              </div>
              <div class="p-3 rounded-lg bg-white border border-slate-200 flex flex-col items-center justify-center min-h-[64px]">
                <span class="text-[8px] text-slate-400 uppercase tracking-widest mb-1">Light Header</span>
                <img
                  v-if="modalCustomerLogoPreviewUrl || customerLogoUrl"
                  :src="modalCustomerLogoPreviewUrl || customerLogoUrl"
                  alt="Customer Logo preview"
                  class="max-h-7 max-w-[120px] object-contain"
                />
                <span v-else class="text-[9px] text-slate-400 italic">Default SCADA Icon</span>
              </div>
            </div>
          </div>

          <!-- Customer File Upload Actions -->
          <input
            ref="customerLogoFileInput"
            type="file"
            accept="image/png,image/jpeg,image/jpg,image/svg+xml,image/webp"
            class="hidden"
            @change="onCustomerFileSelected"
          />
          <div class="flex items-center space-x-2 pt-1">
            <button
              type="button"
              @click="triggerCustomerFileInput"
              class="flex-1 py-1.5 px-3 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-[#20354B] font-bold text-3xs flex items-center justify-center space-x-1.5 transition"
            >
              <span>📁</span>
              <span>{{ (modalCustomerLogoPreviewUrl || customerLogoUrl) ? 'Change Customer Logo...' : 'Upload Customer Logo...' }}</span>
            </button>
            <button
              v-if="modalCustomerLogoPreviewUrl || customerLogoUrl"
              type="button"
              @click="handleRemoveCustomerLogo"
              class="py-1.5 px-3 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/30 font-bold text-3xs flex items-center justify-center space-x-1 transition"
            >
              <span>🗑️</span>
              <span>Remove</span>
            </button>
          </div>
        </div>

        <!-- TAB 2: MAKER COMPANY SETUP -->
        <div v-else-if="activeBrandingTab === 'maker'" class="space-y-3">
          <div class="space-y-1">
            <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
              {{ $t('branding.makerName') || 'Maker / Provider Company Name (Appears Top-Right)' }}
            </label>
            <input
              v-model="modalMakerName"
              type="text"
              maxlength="50"
              placeholder="e.g. PT. Centra Biotech Indonesia"
              class="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] bg-slate-50 dark:bg-[#101B2B] text-slate-900 dark:text-slate-100 text-xs focus:ring-1 focus:ring-cyan-500 focus:outline-none"
            />
          </div>

          <div class="space-y-1">
            <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
              {{ $t('branding.makerSubtitle') || 'Maker Tagline / Subtitle' }}
            </label>
            <input
              v-model="modalMakerSubtitle"
              type="text"
              maxlength="50"
              placeholder="e.g. CBI Industrial IoT"
              class="w-full px-2.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] bg-slate-50 dark:bg-[#101B2B] text-slate-900 dark:text-slate-100 text-xs focus:ring-1 focus:ring-cyan-500 focus:outline-none"
            />
          </div>

          <!-- Maker Logo Live Preview (Dark & Light) -->
          <div class="space-y-1.5">
            <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
              {{ $t('branding.makerLogoPreview') || 'Maker Logo Preview (Header Scale)' }}
            </label>
            <div class="grid grid-cols-2 gap-2">
              <div class="p-3 rounded-lg bg-[#0B1422] border border-[#20354B] flex flex-col items-center justify-center min-h-[64px]">
                <span class="text-[8px] text-slate-400 uppercase tracking-widest mb-1">Dark Header</span>
                <img
                  v-if="modalMakerLogoPreviewUrl || makerLogoUrl"
                  :src="modalMakerLogoPreviewUrl || makerLogoUrl"
                  alt="Maker Logo preview"
                  class="max-h-7 max-w-[120px] object-contain"
                />
                <span v-else class="text-[9px] text-slate-500 italic">No logo configured</span>
              </div>
              <div class="p-3 rounded-lg bg-white border border-slate-200 flex flex-col items-center justify-center min-h-[64px]">
                <span class="text-[8px] text-slate-400 uppercase tracking-widest mb-1">Light Header</span>
                <img
                  v-if="modalMakerLogoPreviewUrl || makerLogoUrl"
                  :src="modalMakerLogoPreviewUrl || makerLogoUrl"
                  alt="Maker Logo preview"
                  class="max-h-7 max-w-[120px] object-contain"
                />
                <span v-else class="text-[9px] text-slate-400 italic">No logo configured</span>
              </div>
            </div>
          </div>

          <!-- Maker File Upload Actions -->
          <input
            ref="makerLogoFileInput"
            type="file"
            accept="image/png,image/jpeg,image/jpg,image/svg+xml,image/webp"
            class="hidden"
            @change="onMakerFileSelected"
          />
          <div class="flex items-center space-x-2 pt-1">
            <button
              type="button"
              @click="triggerMakerFileInput"
              class="flex-1 py-1.5 px-3 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-[#20354B] font-bold text-3xs flex items-center justify-center space-x-1.5 transition"
            >
              <span>📁</span>
              <span>{{ (modalMakerLogoPreviewUrl || makerLogoUrl) ? ($t('branding.changeMakerLogo') || 'Change Maker Logo...') : ($t('branding.uploadMakerLogo') || 'Upload Maker Logo...') }}</span>
            </button>
            <button
              v-if="modalMakerLogoPreviewUrl || makerLogoUrl"
              type="button"
              @click="handleRemoveMakerLogo"
              class="py-1.5 px-3 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/30 font-bold text-3xs flex items-center justify-center space-x-1 transition"
            >
              <span>🗑️</span>
              <span>{{ $t('branding.remove') || 'Remove' }}</span>
            </button>
          </div>
        </div>

        <!-- System Admin Configuration Deep Link -->
        <div class="p-2 rounded bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] flex items-center justify-between text-3xs">
          <span class="text-slate-500 dark:text-slate-400">Advanced settings available in System Admin:</span>
          <router-link
            to="/administration/config?tab=branding"
            @click.native="closeBrandingModal"
            class="text-cyan-600 dark:text-cyan-400 font-bold hover:underline flex items-center space-x-1"
          >
            <span>⚙️ System Admin Setup →</span>
          </router-link>
        </div>

        <!-- Footer Actions -->
        <div class="flex items-center justify-end space-x-2 pt-3 border-t border-slate-200 dark:border-[#20354B]">
          <button
            type="button"
            @click="closeBrandingModal"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            type="button"
            @click="saveBrandingFromModal"
            :disabled="brandingSaving || logoSaving"
            class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 disabled:opacity-50 text-slate-950 font-bold text-xs shadow-xs transition"
          >
            <span v-if="brandingSaving || logoSaving">Saving...</span>
            <span v-else>{{ $t('branding.saveConfig') || 'Save Configuration' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 1: OPERATOR ACKNOWLEDGE MODAL -->
    <div
      v-if="showAckModal && targetAckAlarm"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-slate-950/80 backdrop-blur-sm"
    >
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] p-5 shadow-2xl text-slate-900 dark:text-slate-100 space-y-4 font-mono text-xs">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-[#20354B] pb-3">
          <div class="flex items-center space-x-2">
            <span class="text-base">🛡️</span>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              {{ $t('operator.ackModalTitle') || 'Acknowledge Alarm Incident' }}
            </h3>
          </div>
          <button @click="showAckModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">✕</button>
        </div>

        <div class="p-3 rounded-xl bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] space-y-1 text-2xs">
          <div class="flex items-center justify-between">
            <span class="font-bold text-slate-900 dark:text-white">[{{ targetAckAlarm.device ? targetAckAlarm.device.device_code : 'SYS' }}]</span>
            <span class="font-bold" :class="targetAckAlarm.severity === 'CRITICAL' ? 'text-rose-600 dark:text-rose-400' : 'text-amber-600 dark:text-amber-400'">
              {{ targetAckAlarm.severity }}
            </span>
          </div>
          <div class="text-slate-700 dark:text-slate-300 font-sans text-xs">
            {{ targetAckAlarm.message || targetAckAlarm.alarm_code }}
          </div>
          <div v-if="targetAckAlarm.trigger_value != null" class="text-4xs text-slate-500 dark:text-slate-400">
            Trigger value: <strong class="text-rose-600 dark:text-rose-400">{{ targetAckAlarm.trigger_value }}</strong>
          </div>
        </div>

        <div class="space-y-1.5 font-sans">
          <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
            {{ $t('operator.operatorNotes') || 'Operator Notes (Optional)' }}
          </label>
          <textarea
            v-model="ackNotes"
            rows="3"
            class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-[#20354B] bg-slate-50 dark:bg-[#101B2B] text-slate-900 dark:text-slate-200 text-xs focus:ring-1 focus:ring-cyan-500 focus:outline-none"
            :placeholder="$t('operator.notesPlaceholder') || 'Action taken, field observation...'"
          ></textarea>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
          <button
            @click="showAckModal = false"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="confirmAckAlarm"
            :disabled="isAckSubmitting"
            class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50"
          >
            {{ isAckSubmitting ? 'Acknowledging...' : ($t('operator.confirmAck') || 'Confirm Acknowledgement') }}
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 2: EXPORT TELEMETRY MODAL -->
    <div
      v-if="showExportModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-slate-950/80 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] p-5 shadow-2xl text-slate-900 dark:text-slate-100 space-y-4 font-mono text-xs">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-[#20354B] pb-3">
          <div class="flex items-center space-x-2">
            <span class="text-base text-cyan-500">📦</span>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              {{ $t('operator.exportModalTitle') || 'Export Operational Telemetry' }}
            </h3>
          </div>
          <button @click="showExportModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">✕</button>
        </div>

        <!-- Export Controls Form -->
        <div class="space-y-3 font-sans text-xs">
          <!-- Time Range -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
              {{ $t('operator.selectDateRange') || 'Time Range' }}
            </label>
            <div class="grid grid-cols-2 gap-2">
              <input
                type="datetime-local"
                v-model="exportForm.startTime"
                class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 font-mono text-2xs focus:outline-none focus:border-cyan-500"
              />
              <input
                type="datetime-local"
                v-model="exportForm.endTime"
                class="px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 font-mono text-2xs focus:outline-none focus:border-cyan-500"
              />
            </div>
          </div>

          <!-- Data Source Mode -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
              {{ $t('operator.dataSource') || 'Data Source Mode' }}
            </label>
            <div class="grid grid-cols-3 gap-2 font-mono text-2xs">
              <button
                @click="exportForm.mode = 'PROCESSED'"
                class="px-2 py-1.5 rounded-lg border font-semibold transition"
                :class="exportForm.mode === 'PROCESSED'
                  ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border-cyan-500/50'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B]'"
              >
                PROCESSED
              </button>
              <button
                @click="exportForm.mode = 'RAW'"
                class="px-2 py-1.5 rounded-lg border font-semibold transition"
                :class="exportForm.mode === 'RAW'
                  ? 'bg-emerald-500/20 text-emerald-700 dark:text-emerald-300 border-emerald-500/50'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B]'"
              >
                RAW SENSOR
              </button>
              <button
                @click="exportForm.mode = 'AGGREGATED'"
                class="px-2 py-1.5 rounded-lg border font-semibold transition"
                :class="exportForm.mode === 'AGGREGATED'
                  ? 'bg-indigo-500/20 text-indigo-700 dark:text-indigo-300 border-indigo-500/50'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B]'"
              >
                AGGREGATED
              </button>
            </div>
          </div>

          <!-- Target Device -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
              {{ $t('operator.selectDevices') || 'Device Selection' }}
            </label>
            <select
              v-model="exportForm.deviceId"
              class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 text-xs focus:outline-none focus:border-cyan-500 font-mono"
            >
              <option value="">{{ $t('operator.allDevicesSelected') || 'All Fleet Devices (Unified)' }}</option>
              <option v-for="dev in devicesList" :key="'exp-dev-' + dev.id" :value="dev.id">
                {{ dev.device_code || dev.code }} — {{ dev.device_name || dev.name }}
              </option>
            </select>
          </div>

          <!-- Format Selection -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">
              {{ $t('operator.selectFormat') || 'Export Format' }}
            </label>
            <div class="flex items-center space-x-3 font-mono text-2xs">
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input type="radio" value="csv" v-model="exportForm.format" class="text-cyan-500" />
                <span class="text-slate-700 dark:text-slate-300">CSV Spreadsheet</span>
              </label>
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input type="radio" value="json" v-model="exportForm.format" class="text-cyan-500" />
                <span class="text-slate-700 dark:text-slate-300">JSON Records</span>
              </label>
            </div>
          </div>
        </div>

        <!-- Export Progress / Feedback -->
        <div v-if="exportStatusMessage" class="p-2.5 rounded-lg bg-cyan-50 dark:bg-cyan-950/40 border border-cyan-200 dark:border-cyan-800/40 text-cyan-800 dark:text-cyan-300 text-2xs font-mono">
          {{ exportStatusMessage }}
        </div>

        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
          <button
            @click="showExportModal = false"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="executeTelemetryExport"
            :disabled="isExporting"
            class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50 flex items-center space-x-1.5"
          >
            <span v-if="isExporting" class="animate-spin w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full"></span>
            <span>{{ isExporting ? 'Generating...' : ($t('operator.downloadExport') || 'Download Export') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 3: OPERATIONAL REPORT MODAL -->
    <div
      v-if="showReportModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-slate-950/80 backdrop-blur-sm"
    >
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] p-5 shadow-2xl text-slate-900 dark:text-slate-100 space-y-4 font-mono text-xs">
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-[#20354B] pb-3">
          <div class="flex items-center space-x-2">
            <span class="text-base text-blue-500">📊</span>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              {{ $t('operator.reportModalTitle') || 'Generate Operational Report' }}
            </h3>
          </div>
          <button @click="showReportModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-white">✕</button>
        </div>

        <div class="space-y-3 font-sans text-xs">
          <!-- Report Type -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">Report Focus</label>
            <select
              v-model="reportForm.type"
              class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 text-xs focus:outline-none focus:border-cyan-500 font-mono"
            >
              <option value="COMPREHENSIVE">Comprehensive Operational (Fleet, Quality, Alarms)</option>
              <option value="TELEMETRY">Telemetry Quality & Data Gaps Focus</option>
              <option value="ALARMS">Alarm Incidents & Audit History</option>
              <option value="COMMUNICATION">Hardware Comm Availability</option>
            </select>
          </div>

          <!-- Preset Interval -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">Time Interval</label>
            <div class="grid grid-cols-4 gap-2 font-mono text-2xs">
              <button
                v-for="p in ['1h', '24h', '7d', '30d']"
                :key="'rep-' + p"
                @click="reportForm.preset = p"
                class="px-2 py-1.5 rounded-lg border font-semibold transition"
                :class="reportForm.preset === p ? 'bg-blue-600 text-white border-blue-500' : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B]'"
              >
                {{ p }}
              </button>
            </div>
          </div>

          <!-- Format Choice -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1">Output Format</label>
            <div class="flex items-center space-x-3 font-mono text-2xs">
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input type="radio" value="html" v-model="reportForm.format" class="text-blue-500" />
                <span class="text-slate-700 dark:text-slate-300">Printable HTML / PDF</span>
              </label>
              <label class="flex items-center space-x-1.5 cursor-pointer">
                <input type="radio" value="csv" v-model="reportForm.format" class="text-blue-500" />
                <span class="text-slate-700 dark:text-slate-300">CSV Data Table</span>
              </label>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
          <button
            @click="showReportModal = false"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="executeReportGenerate"
            class="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-bold text-xs shadow-md transition"
          >
            Generate & Download
          </button>
        </div>
      </div>
    </div>

    <!-- ======================================================== -->
    <!-- MODAL 4: FULLSCREEN SCADA TELEMETRY TREND ANALYZER       -->
    <!-- ======================================================== -->
    <div
      v-if="showTrendFullscreen"
      class="fixed inset-0 z-50 bg-slate-950/85 backdrop-blur-md flex items-center justify-center p-2 sm:p-4 animate-fade-in"
      @keydown.esc="closeTrendFullscreen"
    >
      <div class="w-full max-w-7xl h-[92vh] bg-white dark:bg-[#070D18] border border-slate-200 dark:border-[#20354B] rounded-2xl shadow-2xl flex flex-col overflow-hidden">
        <!-- Fullscreen Header -->
        <div class="px-5 py-3 border-b border-slate-200 dark:border-[#20354B] flex flex-wrap items-center justify-between gap-3 bg-slate-50 dark:bg-[#0B1422]">
          <div class="flex items-center space-x-3">
            <span class="w-2.5 h-2.5 rounded-full bg-cyan-500 animate-pulse"></span>
            <div>
              <h2 class="text-sm sm:text-base font-black tracking-wider uppercase text-slate-900 dark:text-white font-mono flex items-center space-x-2">
                <span>SCADA TELEMETRY TREND ANALYZER</span>
                <span class="text-3xs font-semibold px-2 py-0.5 rounded-full bg-cyan-500/20 text-cyan-400 border border-cyan-500/30">
                  {{ chartMode.toUpperCase() }}
                </span>
              </h2>
              <p class="text-3xs text-slate-500 dark:text-slate-400 font-mono mt-0.5">
                Window: {{ formatFullscreenWindowStr }} · Active Series: {{ activeTrendSeries.length }}
              </p>
            </div>
          </div>

          <!-- Header Action Buttons -->
          <div class="flex items-center space-x-2 font-mono text-xs">
            <!-- CSV Download -->
            <button
              @click="downloadTrendCSV"
              class="px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#1E2E42] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-200 text-3xs font-bold transition flex items-center space-x-1.5 cursor-pointer"
            >
              <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
              <span>Download CSV</span>
            </button>

            <!-- PNG Download -->
            <button
              @click="downloadTrendChart('png')"
              class="px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#1E2E42] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-200 text-3xs font-bold transition flex items-center space-x-1.5 cursor-pointer"
            >
              <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
              <span>Download PNG</span>
            </button>

            <!-- SVG Download -->
            <button
              @click="downloadTrendChart('svg')"
              class="px-2 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#1E2E42] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-200 text-3xs font-bold transition hidden sm:flex items-center space-x-1 cursor-pointer"
            >
              <OperatorDashboardIcon name="download" :size="12" class="text-cyan-500" />
              <span>SVG</span>
            </button>

            <!-- Close Fullscreen -->
            <button
              @click="closeTrendFullscreen"
              class="p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-[#1E2E42] text-slate-500 hover:text-slate-900 dark:hover:text-white transition cursor-pointer"
              title="Close Fullscreen (Esc)"
            >
              <OperatorDashboardIcon name="close" :size="16" />
            </button>
          </div>
        </div>

        <!-- Fullscreen Filter & Toolbar Controls -->
        <div class="px-5 py-2.5 bg-slate-100/60 dark:bg-[#0B1422]/60 border-b border-slate-200 dark:border-[#20354B] flex flex-wrap items-center justify-between gap-3 text-3xs font-mono">
          <div class="flex flex-wrap items-center gap-2">
            <!-- Timeframe Presets -->
            <div class="flex items-center bg-white dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">
              <button
                v-for="tf in ['1m', '5m', '15m', '1h', '6h']"
                :key="'fs-' + tf"
                @click="setTimeframe(tf)"
                class="px-2.5 py-1 rounded transition font-bold"
                :class="!fullscreenIsCustom && currentTimeframe === tf
                  ? 'bg-cyan-500 text-slate-950 shadow-xs'
                  : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-200'"
              >
                {{ tf }}
              </button>
            </div>

            <!-- Custom Date Range Inputs -->
            <div class="flex items-center space-x-1 bg-white dark:bg-[#101B2B] p-1 rounded-lg border border-slate-200 dark:border-[#20354B]">
              <input
                type="datetime-local"
                v-model="fullscreenCustomStart"
                class="bg-transparent text-slate-900 dark:text-white text-3xs px-1.5 py-0.5 border-none focus:outline-none"
              />
              <span class="text-slate-400">→</span>
              <input
                type="datetime-local"
                v-model="fullscreenCustomEnd"
                class="bg-transparent text-slate-900 dark:text-white text-3xs px-1.5 py-0.5 border-none focus:outline-none"
              />
              <button
                @click="setCustomTimeframe"
                class="px-2 py-0.5 rounded bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold transition"
              >
                Apply
              </button>
            </div>

            <!-- Mode Selector -->
            <div class="flex items-center bg-white dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">
              <button
                @click="setChartMode('processed')"
                class="px-2.5 py-1 rounded transition font-bold"
                :class="chartMode === 'processed' ? 'bg-cyan-500 text-slate-950' : 'text-slate-500 hover:text-slate-300'"
              >
                PROCESSED
              </button>
              <button
                @click="setChartMode('raw')"
                class="px-2.5 py-1 rounded transition font-bold"
                :class="chartMode === 'raw' ? 'bg-cyan-500 text-slate-950' : 'text-slate-500 hover:text-slate-300'"
              >
                RAW
              </button>
            </div>

            <!-- Scale Mode Toggle -->
            <div class="flex items-center bg-white dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">
              <button
                @click="trendScaleMode = 'auto'"
                class="px-2.5 py-1 rounded transition font-bold"
                :class="trendScaleMode === 'auto' ? 'bg-blue-600 text-white' : 'text-slate-500 hover:text-slate-300'"
              >
                NATIVE UNITS
              </button>
              <button
                @click="trendScaleMode = 'normalized'"
                class="px-2.5 py-1 rounded transition font-bold"
                :class="trendScaleMode === 'normalized' ? 'bg-blue-600 text-white' : 'text-slate-500 hover:text-slate-300'"
              >
                NORMALIZED (%)
              </button>
            </div>

            <!-- Fullscreen Spike Zoom Controls -->
            <div v-if="hasOutlierSpike && trendScaleMode === 'auto'" class="flex items-center bg-white dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">
              <button
                @click="trendSpikeZoom = 'robust'"
                class="px-2.5 py-1 rounded transition font-bold"
                :class="trendSpikeZoom === 'robust' ? 'bg-cyan-600 text-white' : 'text-slate-500 hover:text-slate-300'"
                title="Focus on normal telemetry variation (outliers marked with ▲)"
              >
                ⚡ FOCUS NORMAL
              </button>
              <button
                @click="trendSpikeZoom = 'full'"
                class="px-2.5 py-1 rounded transition font-bold"
                :class="trendSpikeZoom === 'full' ? 'bg-cyan-600 text-white' : 'text-slate-500 hover:text-slate-300'"
                title="Scale axis to encompass full spike peak"
              >
                FULL SPAN
              </button>
            </div>
          </div>

          <!-- Parameter Selector Button / Dropdown -->
          <div class="relative">
            <button
              @click="showTrendParamPicker = !showTrendParamPicker"
              class="px-3 py-1.5 rounded-lg bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30 font-bold transition flex items-center space-x-1.5"
            >
              <span>+ Add / Remove Parameter ({{ activeTrendSeries.length }}/{{ maxTrendSeries }})</span>
            </button>

            <!-- Parameter Picker Dropdown -->
            <div
              v-if="showTrendParamPicker"
              class="absolute right-0 mt-1 w-72 max-h-72 overflow-y-auto bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] rounded-xl shadow-2xl p-2 z-50 text-3xs font-mono"
            >
              <div class="flex items-center justify-between text-slate-400 font-bold px-2 py-1 uppercase border-b border-slate-200 dark:border-[#20354B] mb-1">
                <span>Toggle Series:</span>
                <div class="flex items-center space-x-1">
                  <button @click.stop="selectAllEligibleTrendSeries(true)" class="px-1.5 py-0.5 rounded bg-slate-200 hover:bg-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200">All</button>
                  <button @click.stop="clearAllTrendSeries" class="px-1.5 py-0.5 rounded bg-slate-200 hover:bg-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 text-rose-600 dark:text-rose-400">Clear</button>
                  <button @click.stop="resetTrendSelectionToDefault" class="px-1.5 py-0.5 rounded bg-slate-200 hover:bg-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 text-amber-600 dark:text-amber-400">Reset</button>
                </div>
              </div>
              <div
                v-for="p in allParameters"
                :key="'fs-picker-' + p.param_id"
                @click="selectParamForTrend(p)"
                class="flex items-center justify-between px-2.5 py-1.5 rounded-md hover:bg-slate-100 dark:hover:bg-[#16253B] cursor-pointer transition"
                :class="isParamSelectedForTrend(p) ? 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 font-bold' : 'text-slate-700 dark:text-slate-300'"
              >
                <div class="flex items-center space-x-2">
                  <span class="w-2 h-2 rounded-full" :class="isParamSelectedForTrend(p) ? 'bg-cyan-500' : 'bg-slate-400'"></span>
                  <span>{{ p.device_code }}:{{ p.param_code }}</span>
                </div>
                <span>{{ p.current_val != null ? p.current_val : '--' }} {{ p.unit }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Fullscreen Chart Canvas -->
        <div
          ref="fullscreenChartContainer"
          class="flex-1 min-h-[360px] bg-slate-50 dark:bg-[#070D18] relative select-none p-3"
          @mousemove="onFullscreenChartScrub"
          @mouseleave="fullscreenScrubPoint = null"
        >
          <!-- Loading Overlay -->
          <div v-if="isChartLoading" class="absolute inset-0 bg-white/80 dark:bg-[#070D18]/80 flex items-center justify-center text-cyan-600 dark:text-cyan-400 text-sm font-mono z-20">
            <span class="animate-spin w-5 h-5 border-2 border-cyan-500 border-t-transparent rounded-full mr-2.5"></span>
            Loading Telemetry Trend...
          </div>

          <!-- Empty State -->
          <div v-else-if="activeTrendSeries.length === 0" class="w-full h-full flex flex-col items-center justify-center text-slate-400 dark:text-slate-500 text-xs font-mono space-y-2">
            <span>No series selected for trend graph.</span>
            <button @click="showTrendParamPicker = true" class="text-cyan-500 underline font-bold">Click here to select parameters</button>
          </div>

          <!-- SVG Chart -->
          <svg
            v-else
            ref="fullscreenTrendSvg"
            class="w-full h-full block"
            :viewBox="`0 0 ${fullscreenChartDims.width} ${fullscreenChartDims.height}`"
          >
            <defs>
              <linearGradient id="fullscreenPrimaryFill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" :stop-color="primarySeriesColor" stop-opacity="0.08" />
                <stop offset="100%" :stop-color="primarySeriesColor" stop-opacity="0.00" />
              </linearGradient>
            </defs>

            <!-- Y-Axis Grid Lines & Values -->
            <g class="grid-lines" :stroke="gridStroke" stroke-dasharray="3 3" stroke-width="0.8">
              <g v-for="(tick, idx) in fullscreenYTicks" :key="'fs-ytick-' + idx">
                <line :x1="fullscreenChartDims.padLeft" :y1="tick.y" :x2="fullscreenChartDims.width - fullscreenChartDims.padRight" :y2="tick.y" stroke-opacity="0.75" />
                <text
                  :x="fullscreenChartDims.padLeft - 8"
                  :y="tick.y + 3.5"
                  text-anchor="end"
                  font-size="11"
                  font-family="monospace"
                  :fill="isDark ? '#9BB4C8' : '#475569'"
                >{{ tick.label }}</text>
              </g>

              <!-- Dual-axis Right Ticks (when active series have incompatible units) -->
              <template v-if="trendScales.isDual">
                <g v-for="(tick, idx) in fullscreenRightYTicks" :key="'fs-rytick-' + idx">
                  <text
                    :x="fullscreenChartDims.width - fullscreenChartDims.padRight + 8"
                    :y="tick.y + 3.5"
                    text-anchor="start"
                    font-size="11"
                    font-family="monospace"
                    :fill="trendScales.right ? trendScales.right.color : (isDark ? '#9BB4C8' : '#475569')"
                  >{{ tick.label }}</text>
                </g>
              </template>

              <!-- Base Axis Line -->
              <line :x1="fullscreenChartDims.padLeft" :y1="fullscreenChartDims.baseY" :x2="fullscreenChartDims.width - fullscreenChartDims.padRight" :y2="fullscreenChartDims.baseY" stroke-dasharray="none" :stroke="axisStroke" stroke-width="1.2" />
            </g>

            <!-- Axis Unit Badges at top -->
            <g class="axis-units">
              <text
                :x="fullscreenChartDims.padLeft"
                y="20"
                font-size="11.5"
                font-family="monospace"
                font-weight="bold"
                :fill="trendScales.left ? trendScales.left.color : (isDark ? '#21C7E8' : '#0284C7')"
              >
                {{ trendScales.left ? trendScales.left.label : '' }}
              </text>
              <text
                v-if="trendScales.isDual && trendScales.right"
                :x="fullscreenChartDims.width - fullscreenChartDims.padRight"
                y="20"
                font-size="11.5"
                font-family="monospace"
                font-weight="bold"
                text-anchor="end"
                :fill="trendScales.right.color"
              >
                {{ trendScales.right.label }}
              </text>
            </g>

            <!-- X-Axis Time Ticks & Timestamps -->
            <g class="time-ticks">
              <g v-for="(tick, idx) in fullscreenXTicks" :key="'fs-xtick-' + idx">
                <line :x1="tick.x" :y1="fullscreenChartDims.baseY" :x2="tick.x" :y2="fullscreenChartDims.baseY + 5" :stroke="axisStroke" stroke-width="1" />
                <text
                  :x="tick.x"
                  :y="fullscreenChartDims.baseY + 18"
                  text-anchor="middle"
                  font-size="10.5"
                  font-family="monospace"
                  :fill="isDark ? '#9BB4C8' : '#475569'"
                >{{ tick.label }}</text>
              </g>
            </g>

            <!-- Threshold Reference Lines in Fullscreen -->
            <g v-if="fullscreenThresholdLines.crit != null">
              <line :x1="fullscreenChartDims.padLeft" :y1="fullscreenThresholdLines.crit.y" :x2="fullscreenChartDims.width - fullscreenChartDims.padRight" :y2="fullscreenThresholdLines.crit.y" stroke="#EF4444" stroke-width="1.2" stroke-dasharray="4 4" opacity="0.85" />
              <text :x="fullscreenChartDims.width - fullscreenChartDims.padRight - 5" :y="fullscreenThresholdLines.crit.y - 4" fill="#EF4444" font-size="10" text-anchor="end" font-family="monospace" font-weight="bold">
                {{ fullscreenThresholdLines.crit.label }}
              </text>
            </g>
            <g v-if="fullscreenThresholdLines.warn != null">
              <line :x1="fullscreenChartDims.padLeft" :y1="fullscreenThresholdLines.warn.y" :x2="fullscreenChartDims.width - fullscreenChartDims.padRight" :y2="fullscreenThresholdLines.warn.y" stroke="#F59E0B" stroke-width="1.2" stroke-dasharray="4 4" opacity="0.85" />
              <text :x="fullscreenChartDims.width - fullscreenChartDims.padRight - 5" :y="fullscreenThresholdLines.warn.y - 4" fill="#F59E0B" font-size="10" text-anchor="end" font-family="monospace" font-weight="bold">
                {{ fullscreenThresholdLines.warn.label }}
              </text>
            </g>

            <!-- Fullscreen Outlier Spike Markers -->
            <g v-if="fullscreenOutlierSpikes.length > 0" class="fs-outlier-spikes">
              <g
                v-for="(spike, spkIdx) in fullscreenOutlierSpikes"
                :key="'fs-outlier-' + spkIdx"
                class="cursor-pointer"
                @click="toggleTrendSpikeZoom"
              >
                <line
                  :x1="spike.px"
                  :y1="fullscreenChartDims.padTop"
                  :x2="spike.px"
                  :y2="fullscreenChartDims.baseY"
                  :stroke="spike.color"
                  stroke-width="1"
                  stroke-dasharray="2 2"
                  opacity="0.45"
                />
                <polygon
                  :points="`${spike.px - 5},${fullscreenChartDims.padTop} ${spike.px + 5},${fullscreenChartDims.padTop} ${spike.px},${fullscreenChartDims.padTop + 7}`"
                  :fill="spike.color"
                />
                <rect
                  :x="Math.min(Math.max(spike.px - 36, fullscreenChartDims.padLeft), fullscreenChartDims.width - fullscreenChartDims.padRight - 72)"
                  y="16"
                  width="72"
                  height="18"
                  rx="3"
                  :fill="isDark ? '#0B1422' : '#FFFFFF'"
                  :stroke="spike.color"
                  stroke-width="1"
                />
                <text
                  :x="Math.min(Math.max(spike.px, fullscreenChartDims.padLeft + 36), fullscreenChartDims.width - fullscreenChartDims.padRight - 36)"
                  y="29"
                  text-anchor="middle"
                  font-size="10"
                  font-family="monospace"
                  font-weight="bold"
                  :fill="spike.color"
                >
                  ▲ {{ spike.displayVal }}
                </text>
              </g>
            </g>

            <!-- Primary Series Subtle Area Fill -->
            <path
              v-if="activeTrendSeries.length > 0 && generatePrimaryAreaPath(null, fullscreenChartDims.width, fullscreenChartDims.height, fullscreenChartDims.padLeft, fullscreenChartDims.padRight, fullscreenChartDims.padTop, fullscreenChartDims.padBottom)"
              :d="generatePrimaryAreaPath(null, fullscreenChartDims.width, fullscreenChartDims.height, fullscreenChartDims.padLeft, fullscreenChartDims.padRight, fullscreenChartDims.padTop, fullscreenChartDims.padBottom)"
              fill="url(#fullscreenPrimaryFill)"
            />

            <!-- Series Plots (Unfilled clean 1.8px lines with gap handling) -->
            <g v-for="(s, sIdx) in activeTrendSeries" :key="'fs-plot-' + s.id">
              <path
                :d="generateChartPath(s, null, fullscreenChartDims.width, fullscreenChartDims.height, fullscreenChartDims.padLeft, fullscreenChartDims.padRight, fullscreenChartDims.padTop, fullscreenChartDims.padBottom)"
                fill="none"
                :stroke="s.color"
                :stroke-width="hoveredTrendSeriesId === s.id ? 2.4 : (sIdx === 0 ? 1.8 : 1.6)"
                :opacity="hoveredTrendSeriesId && hoveredTrendSeriesId !== s.id ? 0.35 : 1"
                stroke-linecap="round"
                stroke-linejoin="round"
                class="transition-all duration-150 cursor-pointer"
                @mouseenter="hoveredTrendSeriesId = s.id"
                @mouseleave="hoveredTrendSeriesId = null"
              />
            </g>

            <!-- Hover Scrub Crosshair & Value Dots -->
            <g v-if="fullscreenScrubPoint">
              <line :x1="fullscreenScrubPoint.x" :y1="fullscreenChartDims.padTop" :x2="fullscreenScrubPoint.x" :y2="fullscreenChartDims.baseY" stroke="#00E5FF" stroke-width="1.5" stroke-dasharray="3 3" opacity="0.85" />
              <g v-for="sp in fullscreenScrubPoint.allSeries" :key="'fs-scrub-dot-' + sp.seriesId">
                <circle
                  v-if="sp.plotY != null"
                  :cx="fullscreenScrubPoint.x"
                  :cy="sp.plotY"
                  r="4"
                  :fill="sp.color"
                  :stroke="isDark ? '#060B13' : '#FFFFFF'"
                  stroke-width="1.5"
                />
              </g>
            </g>
          </svg>

          <!-- Hover Tooltip in Fullscreen -->
          <div
            v-if="fullscreenScrubPoint"
            class="absolute z-30 pointer-events-none bg-white/95 dark:bg-[#080E18]/95 backdrop-blur-md border border-cyan-500/60 rounded-xl p-3 shadow-2xl text-2xs font-mono text-slate-900 dark:text-white transition-all duration-75 min-w-[260px] max-w-[340px]"
            :style="{ left: Math.min(Math.max(fullscreenScrubPoint.rawPixelX ? (fullscreenScrubPoint.rawPixelX > 480 ? fullscreenScrubPoint.rawPixelX - 280 : fullscreenScrubPoint.rawPixelX + 20) : (fullscreenScrubPoint.x - 70), 20), 650) + 'px', top: '20px' }"
          >
            <div class="flex items-center justify-between text-slate-400 dark:text-slate-400 text-3xs border-b border-slate-200 dark:border-[#20354B] pb-1.5 mb-2 font-bold font-mono">
              <span class="flex items-center space-x-1">
                <span>🕒</span>
                <span class="text-slate-700 dark:text-slate-200">{{ fullscreenScrubPoint.date }}</span>
                <span>{{ fullscreenScrubPoint.time }}</span>
              </span>
              <span class="px-1.5 py-0.2 rounded text-4xs bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30">
                {{ currentTimeframe }}
              </span>
            </div>
            <div class="space-y-1.5">
              <div v-for="sp in fullscreenScrubPoint.allSeries" :key="sp.seriesId" class="flex items-center justify-between space-x-3 py-0.5">
                <div class="flex items-center space-x-2 min-w-0">
                  <span class="w-2.5 h-2.5 rounded-full shrink-0 shadow-xs" :style="{ backgroundColor: sp.color }"></span>
                  <span class="text-slate-500 dark:text-slate-400 text-3xs shrink-0">{{ sp.deviceCode }} ·</span>
                  <span class="font-bold text-slate-800 dark:text-slate-200 truncate">{{ sp.paramName || sp.paramCode }}</span>
                </div>
                <div class="flex items-center space-x-1.5 shrink-0 font-mono">
                  <span class="font-black tabular-nums" :style="{ color: sp.color }">
                    {{ sp.val }} <span class="text-3xs font-medium text-slate-500 dark:text-slate-400">{{ sp.unit }}</span>
                  </span>
                  <span
                    v-if="sp.quality && sp.quality !== 'GOOD'"
                    class="px-1 py-0.2 rounded text-4xs font-bold"
                    :class="sp.quality === 'BAD' ? 'bg-rose-500/20 text-rose-400 border border-rose-500/30' : 'bg-amber-500/20 text-amber-400 border border-amber-500/30'"
                  >
                    {{ sp.quality }}
                  </span>
                  <span
                    v-if="sp.alarmStatus && sp.alarmStatus !== 'NORMAL'"
                    class="px-1 py-0.2 rounded text-4xs font-bold"
                    :class="sp.alarmStatus === 'CRITICAL' ? 'bg-red-500/20 text-red-400' : 'bg-amber-500/20 text-amber-400'"
                  >
                    {{ sp.alarmStatus === 'CRITICAL' ? 'CRIT' : 'WARN' }}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Fullscreen Bottom Statistics Cards Row -->
        <div class="px-5 py-3 border-t border-slate-200 dark:border-[#20354B] bg-slate-50 dark:bg-[#0B1422] overflow-x-auto">
          <div class="flex items-center space-x-3 min-w-max">
            <div
              v-for="s in activeTrendSeries"
              :key="'fs-stat-' + s.id"
              class="flex items-center space-x-3 px-3 py-2 rounded-xl border bg-white dark:bg-[#101B2B] text-3xs font-mono shadow-xs"
              :style="{ borderColor: s.color + '40' }"
            >
              <span class="w-3 h-3 rounded-full shrink-0" :style="{ backgroundColor: s.color }"></span>
              <div>
                <div class="font-bold text-slate-900 dark:text-white text-2xs">{{ s.deviceCode }}:{{ s.paramCode }}</div>
                <div class="text-slate-500 dark:text-slate-400">Current: <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ s.latestVal }} {{ s.unit }}</span></div>
              </div>
              <div class="border-l border-slate-200 dark:border-[#20354B] pl-3 flex items-center space-x-3 text-slate-500 dark:text-slate-400">
                <div>Min: <span class="font-bold text-slate-900 dark:text-white">{{ s.stats?.min ?? '--' }}</span></div>
                <div>Max: <span class="font-bold text-slate-900 dark:text-white">{{ s.stats?.max ?? '--' }}</span></div>
                <div>Avg: <span class="font-bold text-slate-900 dark:text-white">{{ s.stats?.avg ?? '--' }}</span></div>
                <div>Pts: <span class="font-bold text-slate-900 dark:text-white">{{ s.stats?.count ?? 0 }}</span></div>
              </div>
              <button @click="removeTrendSeries(s.id)" class="text-slate-400 hover:text-rose-500 ml-1">✕</button>
            </div>

            <div v-if="activeTrendSeries.length === 0" class="text-slate-400 font-mono text-3xs py-1">
              No parameters selected. Click "+ Add / Remove Parameter" above to start analyzing trends.
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Category Management Modal -->
    <CategoryManagementModal
      :is-open="isCategoryModalOpen"
      :devices="devicesList"
      @close="isCategoryModalOpen = false"
      @saved="onCategoryConfigSaved"
    />

    <!-- Custom Floating Context Menu for Parameter Cards (Right-Click) -->
    <div
      v-if="paramContextMenu.visible && paramContextMenu.param"
      class="param-context-menu fixed z-50 min-w-[240px] rounded-xl bg-white dark:bg-[#0B1422] border border-slate-300 dark:border-[#20354B] shadow-2xl overflow-hidden font-sans text-xs animate-fade-in divide-y divide-slate-100 dark:divide-[#152538]"
      :style="{ top: paramContextMenu.y + 'px', left: paramContextMenu.x + 'px' }"
      @click.stop
    >
      <!-- Context Menu Header Info -->
      <div class="px-3.5 py-2.5 bg-slate-50 dark:bg-[#080E18] flex items-center space-x-2.5">
        <ParameterIcon
          :icon="paramContextMenu.param.icon"
          :code="paramContextMenu.param.param_code"
          :name="paramContextMenu.param.param_name"
          :unit="paramContextMenu.param.unit"
          size-class="w-4 h-4 flex-shrink-0"
        />
        <div class="min-w-0 flex-1">
          <div class="font-bold text-slate-900 dark:text-white text-xs truncate">
            {{ paramContextMenu.param.param_name || paramContextMenu.param.param_code }}
          </div>
          <div class="text-4xs font-mono text-cyan-600 dark:text-cyan-400 truncate mt-0.5">
            {{ paramContextMenu.param.device_code }} · {{ paramContextMenu.param.param_code }}
          </div>
        </div>
      </div>

      <!-- Actions List -->
      <div class="py-1">
        <!-- 0. Plot / Remove from Trend Analysis -->
        <button
          type="button"
          @click="selectParamForTrend(paramContextMenu.param); closeParamContextMenu()"
          class="w-full px-3.5 py-2 text-left flex items-center space-x-2.5 hover:bg-cyan-50 dark:hover:bg-cyan-500/10 hover:text-cyan-600 dark:hover:text-cyan-300 text-slate-700 dark:text-slate-200 transition group"
        >
          <span class="p-1 rounded-md bg-cyan-500/15 text-cyan-600 dark:text-cyan-400 group-hover:bg-cyan-500/25 flex-shrink-0">
            <OperatorDashboardIcon name="chart" :size="14" />
          </span>
          <div class="flex-1 min-w-0">
            <div class="font-semibold text-xs leading-none">
              {{ isParamSelectedForTrend(paramContextMenu.param) ? ($t('operator.removeFromTrend') || 'Remove from Trend') : ($t('operator.plotTrend') || 'Plot Trend') }}
            </div>
            <div class="text-4xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ isParamSelectedForTrend(paramContextMenu.param) ? 'Hapus parameter dari grafik Trend Analysis' : 'Tampilkan grafik time-series di Trend Analysis' }}
            </div>
          </div>
        </button>

        <!-- 1. Test Read -->
        <button
          type="button"
          @click="handleContextTestRead"
          class="w-full px-3.5 py-2 text-left flex items-center space-x-2.5 hover:bg-cyan-50 dark:hover:bg-cyan-500/10 hover:text-cyan-600 dark:hover:text-cyan-300 text-slate-700 dark:text-slate-200 transition group"
        >
          <span class="p-1 rounded-md bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 group-hover:bg-emerald-500/25 flex-shrink-0">
            <OperatorDashboardIcon name="lightning" :size="14" />
          </span>
          <div class="flex-1 min-w-0">
            <div class="font-semibold text-xs leading-none">Test Read</div>
            <div class="text-4xs text-slate-500 dark:text-slate-400 mt-0.5">Modbus diagnostic report</div>
          </div>
        </button>

        <!-- 2. View Config -->
        <button
          type="button"
          @click="handleContextViewConfig"
          class="w-full px-3.5 py-2 text-left flex items-center space-x-2.5 hover:bg-cyan-50 dark:hover:bg-cyan-500/10 hover:text-cyan-600 dark:hover:text-cyan-300 text-slate-700 dark:text-slate-200 transition group"
        >
          <span class="p-1 rounded-md bg-blue-500/15 text-blue-600 dark:text-blue-400 group-hover:bg-blue-500/25 flex-shrink-0">
            <OperatorDashboardIcon name="eye" :size="14" />
          </span>
          <div class="flex-1 min-w-0">
            <div class="font-semibold text-xs leading-none">View Config</div>
            <div class="text-4xs text-slate-500 dark:text-slate-400 mt-0.5">Lihat konfigurasi parameter</div>
          </div>
        </button>

        <!-- 3. Edit Config (Role protected for Admin/Superadmin) -->
        <button
          type="button"
          @click="handleContextEditConfig"
          :class="canEditConfig ? 'hover:bg-cyan-50 dark:hover:bg-cyan-500/10 hover:text-cyan-600 dark:hover:text-cyan-300 text-slate-700 dark:text-slate-200 cursor-pointer' : 'opacity-40 cursor-not-allowed text-slate-400'"
          class="w-full px-3.5 py-2 text-left flex items-center space-x-2.5 transition group"
          :title="canEditConfig ? 'Edit konfigurasi parameter' : 'Hanya untuk role Administrator / Superadmin'"
        >
          <span class="p-1 rounded-md bg-amber-500/15 text-amber-600 dark:text-amber-400 group-hover:bg-amber-500/25 flex-shrink-0">
            <OperatorDashboardIcon :name="canEditConfig ? 'edit' : 'settings'" :size="14" :class="canEditConfig ? '' : 'text-slate-400'" />
          </span>
          <div class="flex-1 min-w-0">
            <div class="flex items-center space-x-1.5 leading-none">
              <span class="font-semibold text-xs">Edit Config</span>
              <span v-if="!canEditConfig" class="text-4xs font-mono text-amber-500 dark:text-amber-400 px-1 py-0.2 rounded bg-amber-500/10">Admin</span>
            </div>
            <div class="text-4xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ canEditConfig ? 'Ubah batas, rumus, dan register' : 'Perlu role admin / superadmin' }}
            </div>
          </div>
        </button>

        <!-- 4. 100 Data Terakhir -->
        <button
          type="button"
          @click="handleContextHistory"
          class="w-full px-3.5 py-2 text-left flex items-center space-x-2.5 hover:bg-cyan-50 dark:hover:bg-cyan-500/10 hover:text-cyan-600 dark:hover:text-cyan-300 text-slate-700 dark:text-slate-200 transition group"
        >
          <span class="p-1 rounded-md bg-purple-500/15 text-purple-600 dark:text-purple-400 group-hover:bg-purple-500/25 flex-shrink-0">
            <OperatorDashboardIcon name="history" :size="14" />
          </span>
          <div class="flex-1 min-w-0">
            <div class="font-semibold text-xs leading-none">100 Data Terakhir</div>
            <div class="text-4xs text-slate-500 dark:text-slate-400 mt-0.5">Riwayat & audit log telemetri</div>
          </div>
        </button>
      </div>
    </div>

    <!-- 1. Test Read Diagnostic Modal -->
    <TestReadModal
      :is-open="isTestReadModalOpen"
      :device="testReadTarget.device"
      :parameter="testReadTarget.parameter"
      @close="isTestReadModalOpen = false"
      @read-completed="onTestReadCompleted"
    />

    <!-- 2. Parameter View Config Modal -->
    <ParameterConfigViewModal
      :is-open="isViewConfigModalOpen"
      :device="viewConfigTarget.device"
      :parameter="viewConfigTarget.parameter"
      :can-edit="canEditConfig"
      @close="isViewConfigModalOpen = false"
      @edit="switchToEditModal"
    />

    <!-- 3. Parameter Edit Modal (Admin / Superadmin) -->
    <ParameterModal
      v-if="editParamTarget.device && editParamTarget.parameter"
      :is-open="isEditParamModalOpen"
      :device-id="editParamTarget.deviceId"
      :param-to-edit="editParamTarget.parameter"
      :source-device="editParamTarget.device"
      @close="isEditParamModalOpen = false"
      @saved="handleParamSaved"
    />

    <!-- 4. 100 Data Terakhir Modal -->
    <ParameterHistoryModal
      :is-open="isHistoryModalOpen"
      :device="historyTarget.device"
      :parameter="historyTarget.parameter"
      @close="isHistoryModalOpen = false"
    />

    <!-- 5. Configurable Multi-Panel Monitoring Layout Modal -->
    <MultiPanelConfigModal
      :is-open="isPanelConfigModalOpen"
      :panel-count="panelCount"
      :panels="panelLayoutConfigs"
      :categories="activeCategoriesList"
      :all-parameters="allParameters"
      :is-dark="isDark"
      @save="handlePanelConfigSaved"
      @close="isPanelConfigModalOpen = false"
    />

    <!-- 6. Alarm & Audio Notification Settings Modal -->
    <AlarmNotificationModal
      :is-open="showAlarmNotificationModal"
      @close="showAlarmNotificationModal = false; initAudioPreferences()"
    />
  </div>
</template>

<script>
import axios from 'axios';
import audioAlert from '../utils/audioAlert';
import WindRoseChart from '../components/WindRoseChart.vue';
import PollutantDispersionMap from '../components/PollutantDispersionMap.vue';
import ParameterIcon from '../components/ParameterIcon.vue';
import CategoryManagementModal from '../components/CategoryManagementModal.vue';
import TestReadModal from '../components/TestReadModal.vue';
import ParameterConfigViewModal from '../components/ParameterConfigViewModal.vue';
import ParameterHistoryModal from '../components/ParameterHistoryModal.vue';
import ParameterModal from '../components/ParameterModal.vue';
import ISPUIndicatorBar from '../components/ISPUIndicatorBar.vue';
import {
  loadOrganizationBranding,
  saveCustomerBranding,
  removeCustomerLogo,
  saveMakerBranding,
  removeMakerLogo,
  saveOrganizationTextBranding,
  getOrganizationBrandingMeta,
  DEFAULT_BRANDING,
  saveCompanyLogo,
  loadCompanyLogo,
  removeCompanyLogo,
  updateCompanyName
} from '../utils/companyBranding';
import MultiPanelConfigModal from '../components/MultiPanelConfigModal.vue';
import AlarmNotificationModal from '../components/AlarmNotificationModal.vue';
import OperatorDashboardIcon from '../components/OperatorDashboardIcon.vue';
import {
  getStoredAlarmPreferences,
  saveAlarmPreferences,
  applyAlarmPreferences,
} from '../utils/alarmPreferences';
import { getCategoryColor } from '../utils/categoryTheme.js';
import apiClient from '../services/api.js';
import {
  getParamSeriesColor,
  CANONICAL_SERIES_COLORS,
  DASHBOARD_THEME_TOKENS,
  SCADA_FALLBACK_PALETTE,
} from '../utils/trendPalette.js';

// Curated high-contrast SCADA non-alarm palette (Cyans, Blues, Emeralds, Teals, Mints, Indigos, Violets, Purples)
// Strictly excludes red, crimson, rose, orange, amber, and yellow to avoid alarm/error confusion
const TREND_PALETTE = [
  '#00E5FF', '#3B82F6', '#10B981', '#8B5CF6', '#06B6D4', '#6366F1',
  '#14B8A6', '#2DD4BF', '#A855F7', '#38BDF8', '#4F46E5', '#059669',
  '#7C3AED', '#0284C7', '#818CF8', '#0D9488',
];

const ALARM_LIKE_COLORS = new Set([
  '#F59E0B', '#F97316', '#E11D48', '#EAB308', '#EC4899', '#D946EF', '#84CC16',
  '#FF4559', '#FFB020', '#EF4444', '#DC2626', '#E53E3E', '#DD6B20', '#D69E2E',
  '#F87171', '#FBBF24', '#F43F5E', '#EA580C', '#CA8A04'
]);

function isAlarmLikeColor(color) {
  if (!color || typeof color !== 'string') return false;
  return ALARM_LIKE_COLORS.has(color.toUpperCase());
}

export default {
  name: 'OperatorCommandCenterView',
  components: {
    OperatorDashboardIcon,
    WindRoseChart,
    PollutantDispersionMap,
    ParameterIcon,
    CategoryManagementModal,
    TestReadModal,
    ParameterConfigViewModal,
    ParameterHistoryModal,
    ParameterModal,
    ISPUIndicatorBar,
    MultiPanelConfigModal,
    AlarmNotificationModal,
  },
  props: {
    isSidebarCollapsed: {
      type: Boolean,
      default: false,
    },
    isTopPanelHidden: {
      type: Boolean,
      default: false,
    },
    isSidebarAutoHide: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      gridColumns: 4,
      panelCount: 4,
      panelLayoutConfigs: [],
      isPanelConfigModalOpen: false,
      devicesList: [],
      liveMap: {},
      liveBlinks: {},
      paramContextMenu: {
        visible: false,
        x: 0,
        y: 0,
        param: null,
        device: null,
        fullParam: null,
      },
      isTestReadModalOpen: false,
      testReadTarget: { device: null, parameter: null },
      isViewConfigModalOpen: false,
      viewConfigTarget: { device: null, parameter: null },
      isEditParamModalOpen: false,
      editParamTarget: { deviceId: null, parameter: null, device: null },
      isHistoryModalOpen: false,
      historyTarget: { device: null, parameter: null },
      lastGoodTimeMap: (() => {
        try {
          return JSON.parse(localStorage.getItem('dl_last_good_time_map') || '{}');
        } catch (e) {
          return {};
        }
      })(),
      sparklinesMap: {},
      sparklineMetaCache: {},
      pinnedIds: JSON.parse(localStorage.getItem('dl_pinned_params') || '[]'),
      searchQuery: '',
      searchDebounceTimer: null,
      selectedDeviceFilter: '',
      selectedQualityFilter: '',
      selectedAlarmFilter: '',
      selectedCategoryFilter: 'all',
      categoriesList: [],
      isCategoryModalOpen: false,
      filterPinnedOnly: false,
      viewMode: 'cards', // 'cards' | 'table'
      activeToolbarPopover: null,
      // Dual Organization & Company Branding State (Customer Left, Maker Right)
      showBrandingModal: false,
      showLogoModal: false, // legacy alias
      activeBrandingTab: 'customer', // 'customer' | 'maker'
      customerCompanyName: (getOrganizationBrandingMeta() && getOrganizationBrandingMeta().customer && getOrganizationBrandingMeta().customer.name !== undefined) ? getOrganizationBrandingMeta().customer.name : 'PT. Mitra Industri Enviro',
      customerSubtitle: (getOrganizationBrandingMeta() && getOrganizationBrandingMeta().customer && getOrganizationBrandingMeta().customer.subtitle !== undefined) ? (getOrganizationBrandingMeta().customer.subtitle === 'Local Autonomous Edge' ? '' : getOrganizationBrandingMeta().customer.subtitle) : 'AQMS',
      customerLogoUrl: null,
      makerCompanyName: (getOrganizationBrandingMeta() && getOrganizationBrandingMeta().maker && getOrganizationBrandingMeta().maker.name !== undefined) ? getOrganizationBrandingMeta().maker.name : '',
      makerSubtitle: (getOrganizationBrandingMeta() && getOrganizationBrandingMeta().maker && getOrganizationBrandingMeta().maker.subtitle !== undefined) ? getOrganizationBrandingMeta().maker.subtitle : '',
      makerLogoUrl: null,
      // Legacy aliases for company logo & name
      companyLogoUrl: null,
      companyName: '',
      modalCompanyName: '',
      logoPreviewUrl: null,
      logoTempFile: null,
      logoErrorMsg: '',
      logoSaving: false,
      // Modal temporary inputs
      modalCustomerName: '',
      modalCustomerSubtitle: '',
      modalCustomerLogoPreviewUrl: null,
      modalCustomerLogoFile: null,
      modalMakerName: '',
      modalMakerSubtitle: '',
      modalMakerLogoPreviewUrl: null,
      modalMakerLogoFile: null,
      brandingErrorMsg: '',
      brandingSaving: false,
      showTrendSettingsMenu: false, // 'categories' | 'pinned' | 'displayMode' | 'panelLayout' | 'systemSetup' | 'ispuSettings' | null
      isIspuBarVisible: localStorage.getItem('operator_ispu_bar_visible') !== 'false',

      // Environmental Visualizations (Wind Rose & Pollutant Map) Visibility
      showEnvironmentalVisualizations: localStorage.getItem('dl_show_environmental_viz') === 'true',

      // Trend Chart State
      currentTimeframe: '15m',
      chartMode: 'processed', // 'processed' | 'raw'
      trendScaleMode: 'auto', // 'auto' | 'normalized'
      trendSpikeZoom: 'robust', // 'robust' (smart scale with outlier marker) | 'full' (encompass full spike)
      hoveredTrendSeriesId: null,
      activeTrendSeries: [],
      maxTrendSeries: 16,
      trendSelectionRestored: false,
      isChartLoading: false,
      scrubPoint: null,
      fullscreenScrubPoint: null,
      trendTimeBounds: {
        start: Date.now() - 15 * 60 * 1000,
        end: Date.now(),
      },
      showTrendDownloadMenu: false,
      showTrendFullscreen: false,
      fullscreenCustomStart: '',
      fullscreenCustomEnd: '',
      fullscreenIsCustom: false,
      showTrendParamPicker: false,

      // Responsive Trend Canvas Dimensions (1:1 coordinate space prevents text distortion)
      trendChartWidth: 800,
      trendChartHeight: 125,
      fullscreenChartWidth: 1200,
      fullscreenChartHeight: 440,
      trendResizeObserver: null,
      fullscreenResizeObserver: null,

      // Recent Activity & Queue
      activeRecentTab: 'alarms', // 'alarms' | 'events' | 'notifications'
      recentEvents: [],
      notificationQueue: [],
      notificationQueueStats: { total: 0, pending: 0, sent: 0, failed: 0 },

      // Telemetry Freshness Ticker & Live Clock (persisted in localStorage across page reloads)
      lastTelemetryTime: localStorage.getItem('dl_last_telemetry_time')
        ? new Date(localStorage.getItem('dl_last_telemetry_time'))
        : null,
      lastTelemetrySecondsAgo: localStorage.getItem('dl_last_telemetry_time')
        ? Math.max(0, Math.floor((Date.now() - new Date(localStorage.getItem('dl_last_telemetry_time')).getTime()) / 1000))
        : 0,
      currentTimeMs: Date.now(),
      tickerInterval: null,
      pollingInterval: null,
      trendAutoRefreshInterval: null,
      devicePollCycles: 0,
      clockInterval: null,
      currentTime: new Date().toLocaleTimeString('en-GB'),
      isRefreshing: false,
      telemetryRate: 24.5,

      // Synchronized SCADA Alarm Annunciator Pulse
      isAlarmPulseActive: false,
      alarmPulseSeverity: null,
      alarmPulseTimer: null,

      // Modals
      showAckModal: false,
      targetAckAlarm: null,
      ackNotes: '',
      isAckSubmitting: false,

      showExportModal: false,
      exportForm: {
        startTime: '',
        endTime: '',
        mode: 'PROCESSED',
        deviceId: '',
        format: 'csv',
      },
      isExporting: false,
      exportStatusMessage: '',

      showReportModal: false,
      reportForm: {
        type: 'COMPREHENSIVE',
        preset: '24h',
        format: 'html',
      },

      // Audio & Alarm Sound Setup State
      showAudioSetupPopover: false,
      showAlarmNotificationModal: false,
      audioVolume: 0.7,
      audioRepeatInterval: 5,
      audioSoundEnabled: true,
      audioMuteDuration: 0,
    };
  },
  computed: {
    singlePanelGridClass() {
      const cols = Number(this.gridColumns) || 4;
      if (cols === 1) return 'grid-cols-1';
      if (cols === 2) return 'grid-cols-1 md:grid-cols-2';
      if (cols === 3) return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3';
      return 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4';
    },
    singlePanelScrollStyle() {
      return {
        maxHeight: 'calc(100vh - 360px)',
        minHeight: '68px',
      };
    },
    selectedCategoryTitle() {
      if (this.selectedCategoryFilter === 'all') {
        return this.$t('operator.allParameters') || 'All Monitored Parameters';
      }
      if (this.selectedCategoryFilter === 'uncategorized') {
        return this.$t('operator.uncategorized') || 'Uncategorized';
      }
      const cat = (this.activeCategoriesList || []).find(c => c.id === this.selectedCategoryFilter);
      return cat ? cat.name : (this.$t('operator.allParameters') || 'All Monitored Parameters');
    },
    pinnedParametersList() {
      const ids = this.pinnedIds || [];
      return (this.allParameters || []).filter(p => ids.includes(p.param_id));
    },
    isDark() {
      const pref = this.$store.state.uiPreferences;
      if (pref && pref.effectiveTheme) {
        return pref.effectiveTheme === 'dark';
      }
      if (typeof document !== 'undefined') {
        return document.documentElement.classList.contains('dark');
      }
      return true;
    },
    currentLocale() {
      return this.$store.state.language || 'en';
    },
    gridStroke() {
      // Required test assertion: this.isDark ? '#142438' : '#CBD5E1'
      return this.isDark ? '#132738' : '#D7E5EF';
    },
    axisStroke() {
      // Legacy clean slate/blue-gray assertion: this.isDark ? '#20354B' : '#94A3B8'
      return this.isDark ? '#244356' : '#B8D4EA';
    },
    stationTypeLabel() {
      if (this.customerSubtitle !== undefined && this.customerSubtitle !== null) {
        if (this.customerSubtitle === 'Local Autonomous Edge') return '';
        return this.customerSubtitle.trim();
      }
      const dom = (this.systemStatus && this.systemStatus.monitoring_domain) || this.$store.state.systemStatus?.monitoring_domain;
      if (dom) {
        if (dom.toUpperCase() === 'SPARING') return 'Sparing';
        return dom.toUpperCase();
      }
      return '';
    },
    currentUserIdentifier() {
      const u = this.$store.state.user;
      return (u && (u.id || u.username)) ? String(u.id || u.username) : 'default';
    },
    trendStorageKey() {
      return `operatorTrendSelection:v1:${this.currentUserIdentifier}`;
    },
    primarySeriesColor() {
      return this.activeTrendSeries.length > 0 ? this.activeTrendSeries[0].color : '#00E5FF';
    },
    trendChartDims() {
      const width = this.trendChartWidth || 800;
      const height = this.trendChartHeight || 135;
      const padLeft = 44;
      const padRight = this.trendScales?.isDual ? 44 : 28;
      const padTop = 10;
      const padBottom = 24;
      const plotW = Math.max(width - padLeft - padRight, 100);
      const plotH = Math.max(height - padTop - padBottom, 40);
      const baseY = height - padBottom;
      return { width, height, padLeft, padRight, padTop, padBottom, plotW, plotH, baseY };
    },
    fullscreenChartDims() {
      const width = this.fullscreenChartWidth || 1200;
      const height = this.fullscreenChartHeight || 440;
      const padLeft = 65;
      const padRight = this.trendScales?.isDual ? 65 : 45;
      const padTop = 30;
      const padBottom = 35;
      const plotW = Math.max(width - padLeft - padRight, 100);
      const plotH = Math.max(height - padTop - padBottom, 80);
      const baseY = height - padBottom;
      return { width, height, padLeft, padRight, padTop, padBottom, plotW, plotH, baseY };
    },
    tooltipLeftPos() {
      if (!this.scrubPoint) return 10;
      const rawX = this.scrubPoint.rawPixelX != null ? this.scrubPoint.rawPixelX : this.scrubPoint.x;
      const width = this.trendChartDims.width;
      // If hovering on right half of canvas, place tooltip to the left of crosshair
      if (rawX > width * 0.5) {
        return Math.max(10, rawX - 280);
      }
      return Math.min(rawX + 16, width - 290);
    },
    trendScales() {
      if (this.activeTrendSeries.length === 0) {
        return {
          left: { min: 0, max: 100, unit: '', color: '#00E5FF', label: '', hasSpike: false, outlierPoints: [], rawMin: 0, rawMax: 100 },
          right: null,
          isDual: false,
        };
      }

      const computeNiceScale = (yVals, defaultMin = 0, defaultMax = 100, matchedSeries = []) => {
        const valid = (yVals || []).filter(v => v !== null && typeof v === 'number' && isFinite(v));
        if (valid.length === 0) {
          return { min: defaultMin, max: defaultMax, hasSpike: false, outlierPoints: [], rawMin: defaultMin, rawMax: defaultMax };
        }
        let rawMin = Math.min(...valid);
        let rawMax = Math.max(...valid);
        if (rawMin === rawMax) {
          let min = rawMin > 0 ? Math.max(0, rawMin * 0.8 - 1) : rawMin - 5;
          let max = rawMax + 5;
          return { min, max, hasSpike: false, outlierPoints: [], rawMin, rawMax };
        }

        // Analyze distribution for isolated spikes
        let hasSpike = false;
        let isHighSpike = false;
        let isLowSpike = false;
        let outlierPoints = [];
        let min = rawMin;
        let max = rawMax;

        if (valid.length >= 5) {
          const sorted = [...valid].sort((a, b) => a - b);
          const n = sorted.length;
          const p05 = sorted[Math.floor(n * 0.05)];
          const p25 = sorted[Math.floor(n * 0.25)];
          const p75 = sorted[Math.floor(n * 0.75)];
          const p95 = sorted[Math.floor(n * 0.95)];
          const iqr = p75 - p25;
          const normalSpan = Math.max(p95 - p05, iqr * 1.5, Math.abs(p75) * 0.08, 0.01);

          // Identify isolated spikes: points far outside normal range (>= 2.5x normal span beyond p95)
          const highSpikes = sorted.filter(v => v > p95 + 2.5 * normalSpan);
          const lowSpikes = sorted.filter(v => v < p05 - 2.5 * normalSpan);

          // An isolated spike is at most a few points (e.g. <= 5% of dataset or <= 3 points)
          const maxOutlierCount = Math.max(3, Math.ceil(n * 0.05));
          if (highSpikes.length > 0 && highSpikes.length <= maxOutlierCount) {
            isHighSpike = true;
          }
          if (lowSpikes.length > 0 && lowSpikes.length <= maxOutlierCount) {
            isLowSpike = true;
          }
          hasSpike = isHighSpike || isLowSpike;

          if (hasSpike) {
            // Find points with metadata (x, y, seriesId, color)
            matchedSeries.forEach(s => {
              (s.dataPoints || []).forEach(p => {
                if (p.y !== null && typeof p.y === 'number' && isFinite(p.y)) {
                  if ((isHighSpike && p.y > p95 + 2.5 * normalSpan) || (isLowSpike && p.y < p05 - 2.5 * normalSpan)) {
                    outlierPoints.push({
                      x: p.x,
                      y: p.y,
                      seriesId: s.id,
                      paramName: s.paramName || s.paramCode,
                      unit: s.unit || '',
                      color: s.color,
                      quality: p.quality || 'GOOD',
                    });
                  }
                }
              });
            });
          }

          if (hasSpike && this.trendSpikeZoom !== 'full') {
            // SMART ROBUST / FOCUSED SCALE:
            // Dedicate normal vertical span to normal telemetry fluctuations
            min = isLowSpike ? Math.max(0, p05 - normalSpan * 0.15) : (rawMin >= 0 ? Math.max(0, rawMin - normalSpan * 0.08) : rawMin - normalSpan * 0.08);
            max = isHighSpike ? (p95 + normalSpan * 0.25) : (rawMax + normalSpan * 0.08);
          } else {
            // FULL SPAN: Include all data points
            const padding = (rawMax - rawMin) * 0.08;
            min = (rawMin >= 0 && rawMin - padding < 0) ? 0 : rawMin - padding;
            max = rawMax + padding;
          }
        } else {
          const padding = (rawMax - rawMin) * 0.08;
          min = (rawMin >= 0 && rawMin - padding < 0) ? 0 : rawMin - padding;
          max = rawMax + padding;
        }

        // Include relevant threshold if within close proximity to data to avoid empty space
        matchedSeries.forEach(s => {
          const param = this.allParameters.find(p => p.param_id === s.parameterId);
          if (param) {
            if (param.high_warning != null && isFinite(param.high_warning)) {
              if (param.high_warning >= min && param.high_warning <= max * 1.35) {
                max = Math.max(max, param.high_warning * 1.05);
              }
            }
            if (param.high_critical != null && isFinite(param.high_critical)) {
              if (param.high_critical >= min && param.high_critical <= max * 1.35) {
                max = Math.max(max, param.high_critical * 1.05);
              }
            }
          }
        });

        return { min, max, hasSpike, isHighSpike, isLowSpike, outlierPoints, rawMin, rawMax };
      };

      if (this.trendScaleMode === 'normalized') {
        return {
          left: { min: 0, max: 100, unit: '%', color: '#00E5FF', label: '[ 0-100% SPAN ]', hasSpike: false, outlierPoints: [] },
          right: null,
          isDual: false,
        };
      }

      const s1 = this.activeTrendSeries[0];
      const unit1 = (s1.unit || '').trim();
      const diffSeries = this.activeTrendSeries.find(s => (s.unit || '').trim() !== unit1);

      if (!diffSeries) {
        // All series have matching unit or single series
        const allY = [];
        this.activeTrendSeries.forEach(s => {
          (s.dataPoints || []).forEach(p => {
            if (p.y !== null && typeof p.y === 'number' && isFinite(p.y)) allY.push(p.y);
          });
        });
        const leftSeries = this.activeTrendSeries;
        const scale = computeNiceScale(allY, 0, 100, leftSeries);
        let modeLabel = '';
        if (scale.hasSpike) {
          modeLabel = this.trendSpikeZoom === 'full' ? ' · FULL SPAN' : ' · SMART ROBUST';
        }
        return {
          left: {
            min: scale.min,
            max: scale.max,
            unit: unit1,
            color: s1.color,
            label: unit1 ? `[ ${unit1}${modeLabel} ]` : (s1.paramName || s1.paramCode),
            hasSpike: scale.hasSpike,
            outlierPoints: scale.outlierPoints,
            rawMin: scale.rawMin,
            rawMax: scale.rawMax,
          },
          right: null,
          isDual: false,
        };
      }

      // Dual-Axis: Series matching unit1 on left, series with different unit on right
      const unit2 = (diffSeries.unit || '').trim();
      const leftY = [];
      const rightY = [];
      const leftSeries = this.activeTrendSeries.filter(s => (s.unit || '').trim() === unit1);
      const rightSeries = this.activeTrendSeries.filter(s => (s.unit || '').trim() !== unit1);

      this.activeTrendSeries.forEach(s => {
        const u = (s.unit || '').trim();
        const pts = (s.dataPoints || []).map(p => p.y).filter(v => v !== null && typeof v === 'number' && isFinite(v));
        if (u === unit1) {
          leftY.push(...pts);
        } else {
          rightY.push(...pts);
        }
      });

      const leftScale = computeNiceScale(leftY, 0, 100, leftSeries);
      const rightScale = computeNiceScale(rightY, 0, 100, rightSeries);

      let leftModeLabel = leftScale.hasSpike ? (this.trendSpikeZoom === 'full' ? ' · FULL SPAN' : ' · SMART ROBUST') : '';
      let rightModeLabel = rightScale.hasSpike ? (this.trendSpikeZoom === 'full' ? ' · FULL SPAN' : ' · SMART ROBUST') : '';

      return {
        left: {
          min: leftScale.min,
          max: leftScale.max,
          unit: unit1,
          color: s1.color,
          label: unit1 ? `[ ${unit1}${leftModeLabel} ]` : (s1.paramName || s1.paramCode),
          hasSpike: leftScale.hasSpike,
          outlierPoints: leftScale.outlierPoints,
          rawMin: leftScale.rawMin,
          rawMax: leftScale.rawMax,
        },
        right: {
          min: rightScale.min,
          max: rightScale.max,
          unit: unit2,
          color: diffSeries.color,
          label: unit2 ? `[ ${unit2}${rightModeLabel} ]` : (diffSeries.paramName || diffSeries.paramCode),
          hasSpike: rightScale.hasSpike,
          outlierPoints: rightScale.outlierPoints,
          rawMin: rightScale.rawMin,
          rawMax: rightScale.rawMax,
        },
        isDual: true,
      };
    },
    trendOutlierSpikes() {
      if (this.trendScaleMode === 'normalized') return [];
      const spikes = [];
      const timeStart = this.trendTimeBounds.start;
      const timeEnd = this.trendTimeBounds.end;
      const timeSpan = Math.max(timeEnd - timeStart, 1000);
      const { padLeft, plotW } = this.trendChartDims;

      if (this.trendScales.left && this.trendScales.left.hasSpike && this.trendSpikeZoom !== 'full') {
        (this.trendScales.left.outlierPoints || []).forEach(pt => {
          const px = padLeft + Math.min(Math.max((pt.x - timeStart) / timeSpan, 0), 1) * plotW;
          spikes.push({
            ...pt,
            px,
            displayVal: this.formatAxisValue(pt.y),
          });
        });
      }
      if (this.trendScales.isDual && this.trendScales.right && this.trendScales.right.hasSpike && this.trendSpikeZoom !== 'full') {
        (this.trendScales.right.outlierPoints || []).forEach(pt => {
          const px = padLeft + Math.min(Math.max((pt.x - timeStart) / timeSpan, 0), 1) * plotW;
          spikes.push({
            ...pt,
            px,
            displayVal: this.formatAxisValue(pt.y),
          });
        });
      }
      return spikes;
    },
    fullscreenOutlierSpikes() {
      if (this.trendScaleMode === 'normalized') return [];
      const spikes = [];
      const timeStart = this.trendTimeBounds.start;
      const timeEnd = this.trendTimeBounds.end;
      const timeSpan = Math.max(timeEnd - timeStart, 1000);
      const { padLeft, plotW } = this.fullscreenChartDims;

      if (this.trendScales.left && this.trendScales.left.hasSpike && this.trendSpikeZoom !== 'full') {
        (this.trendScales.left.outlierPoints || []).forEach(pt => {
          const px = padLeft + Math.min(Math.max((pt.x - timeStart) / timeSpan, 0), 1) * plotW;
          spikes.push({
            ...pt,
            px,
            displayVal: this.formatAxisValue(pt.y),
          });
        });
      }
      if (this.trendScales.isDual && this.trendScales.right && this.trendScales.right.hasSpike && this.trendSpikeZoom !== 'full') {
        (this.trendScales.right.outlierPoints || []).forEach(pt => {
          const px = padLeft + Math.min(Math.max((pt.x - timeStart) / timeSpan, 0), 1) * plotW;
          spikes.push({
            ...pt,
            px,
            displayVal: this.formatAxisValue(pt.y),
          });
        });
      }
      return spikes;
    },
    hasOutlierSpike() {
      return !!(this.trendScales.left?.hasSpike || this.trendScales.right?.hasSpike);
    },
    activeSpikeSummary() {
      if (this.trendScales.left?.hasSpike) {
        const pts = this.trendScales.left.outlierPoints || [];
        const maxVal = pts.length > 0 ? Math.max(...pts.map(p => p.y)) : 0;
        return `Spike: ${this.formatAxisValue(maxVal)} ${this.trendScales.left.unit || ''}`.trim();
      }
      if (this.trendScales.right?.hasSpike) {
        const pts = this.trendScales.right.outlierPoints || [];
        const maxVal = pts.length > 0 ? Math.max(...pts.map(p => p.y)) : 0;
        return `Spike: ${this.formatAxisValue(maxVal)} ${this.trendScales.right.unit || ''}`.trim();
      }
      return '';
    },
    trendNativeScale() {
      const left = this.trendScales.left;
      return { min: left.min, max: left.max };
    },
    trendYTicks() {
      const scale = this.trendScales.left;
      const ticks = [];
      const steps = 4;
      const { plotH, baseY } = this.trendChartDims;
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const y = baseY - frac * plotH;
        let label = '';
        if (this.trendScaleMode === 'normalized') {
          label = `${Math.round(frac * 100)}%`;
        } else {
          const val = scale.min + frac * (scale.max - scale.min);
          label = this.formatAxisValue(val);
        }
        ticks.push({ y, frac, label });
      }
      return ticks;
    },
    trendRightYTicks() {
      if (!this.trendScales.isDual || !this.trendScales.right) return [];
      const scale = this.trendScales.right;
      const ticks = [];
      const steps = 4;
      const { plotH, baseY } = this.trendChartDims;
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const y = baseY - frac * plotH;
        const val = scale.min + frac * (scale.max - scale.min);
        ticks.push({ y, frac, label: this.formatAxisValue(val) });
      }
      return ticks;
    },
    trendXTicks() {
      const start = this.trendTimeBounds.start;
      const end = this.trendTimeBounds.end;
      const span = Math.max(end - start, 1000);
      const { padLeft, plotW, width } = this.trendChartDims;
      const ticks = [];
      const steps = width >= 900 ? 6 : (width >= 600 ? 5 : 3);
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const x = padLeft + frac * plotW;
        const time = new Date(start + frac * span);
        let label = '';
        if (span <= 7 * 60 * 1000) {
          // 1m / 5m range: e.g. 15:30:00, 15:31:15
          label = time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        } else if (span <= 25 * 3600 * 1000) {
          // 15m, 1h, 6h range: e.g. 15:30, 16:00
          label = time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        } else {
          // Multi-day custom range
          label = time.toLocaleDateString([], { month: 'short', day: 'numeric' });
        }
        ticks.push({ x, label });
      }
      return ticks;
    },
    fullscreenYTicks() {
      const scale = this.trendScales.left;
      const ticks = [];
      const steps = 5;
      const { plotH, baseY } = this.fullscreenChartDims;
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const y = baseY - frac * plotH;
        let label = '';
        if (this.trendScaleMode === 'normalized') {
          label = `${Math.round(frac * 100)}%`;
        } else {
          const val = scale.min + frac * (scale.max - scale.min);
          label = this.formatAxisValue(val);
        }
        ticks.push({ y, frac, label });
      }
      return ticks;
    },
    fullscreenRightYTicks() {
      if (!this.trendScales.isDual || !this.trendScales.right) return [];
      const scale = this.trendScales.right;
      const ticks = [];
      const steps = 5;
      const { plotH, baseY } = this.fullscreenChartDims;
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const y = baseY - frac * plotH;
        const val = scale.min + frac * (scale.max - scale.min);
        ticks.push({ y, frac, label: this.formatAxisValue(val) });
      }
      return ticks;
    },
    fullscreenXTicks() {
      const start = this.trendTimeBounds.start;
      const end = this.trendTimeBounds.end;
      const span = Math.max(end - start, 1000);
      const { padLeft, plotW, width } = this.fullscreenChartDims;
      const ticks = [];
      const steps = width >= 1200 ? 8 : (width >= 800 ? 6 : 4);
      for (let i = 0; i <= steps; i++) {
        const frac = i / steps;
        const x = padLeft + frac * plotW;
        const time = new Date(start + frac * span);
        let label = '';
        if (span <= 7 * 60 * 1000) {
          label = time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
        } else if (span <= 25 * 3600 * 1000) {
          label = time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        } else {
          label = `${time.toLocaleDateString([], { month: 'short', day: 'numeric' })} ${time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`;
        }
        ticks.push({ x, label });
      }
      return ticks;
    },
    trendThresholdLines() {
      if (this.activeTrendSeries.length === 0 || this.trendScaleMode === 'normalized') return { crit: null, warn: null };
      const { baseY, plotH } = this.trendChartDims;
      for (const s of this.activeTrendSeries) {
        const param = this.allParameters.find(p => p.param_id === s.parameterId);
        if (!param || (param.high_critical == null && param.high_warning == null)) continue;

        const isRightAxis = this.trendScales.isDual && (s.unit || '').trim() === (this.trendScales.right?.unit || '').trim();
        const scale = isRightAxis ? this.trendScales.right : this.trendScales.left;
        const span = scale.max - scale.min || 1;

        let crit = null;
        let warn = null;

        if (param.high_critical != null) {
          const frac = (param.high_critical - scale.min) / span;
          if (frac >= 0 && frac <= 1.08) {
            const y = baseY - Math.min(Math.max(frac, 0), 1) * plotH;
            crit = {
              y,
              val: param.high_critical,
              label: `${param.param_code || s.paramCode} CRIT ${param.high_critical} ${param.unit || ''}`.trim(),
            };
          }
        }

        if (param.high_warning != null) {
          const frac = (param.high_warning - scale.min) / span;
          if (frac >= 0 && frac <= 1.08) {
            const y = baseY - Math.min(Math.max(frac, 0), 1) * plotH;
            warn = {
              y,
              val: param.high_warning,
              label: `${param.param_code || s.paramCode} WARN ${param.high_warning} ${param.unit || ''}`.trim(),
            };
          }
        }

        if (crit || warn) return { crit, warn };
      }
      return { crit: null, warn: null };
    },
    fullscreenThresholdLines() {
      if (this.activeTrendSeries.length === 0 || this.trendScaleMode === 'normalized') return { crit: null, warn: null };
      const { baseY, plotH } = this.fullscreenChartDims;
      for (const s of this.activeTrendSeries) {
        const param = this.allParameters.find(p => p.param_id === s.parameterId);
        if (!param || (param.high_critical == null && param.high_warning == null)) continue;

        const isRightAxis = this.trendScales.isDual && (s.unit || '').trim() === (this.trendScales.right?.unit || '').trim();
        const scale = isRightAxis ? this.trendScales.right : this.trendScales.left;
        const span = scale.max - scale.min || 1;

        let crit = null;
        let warn = null;

        if (param.high_critical != null) {
          const frac = (param.high_critical - scale.min) / span;
          if (frac >= 0 && frac <= 1.08) {
            const y = baseY - Math.min(Math.max(frac, 0), 1) * plotH;
            crit = {
              y,
              val: param.high_critical,
              label: `${param.param_code || s.paramCode} CRIT ${param.high_critical} ${param.unit || ''}`.trim(),
            };
          }
        }

        if (param.high_warning != null) {
          const frac = (param.high_warning - scale.min) / span;
          if (frac >= 0 && frac <= 1.08) {
            const y = baseY - Math.min(Math.max(frac, 0), 1) * plotH;
            warn = {
              y,
              val: param.high_warning,
              label: `${param.param_code || s.paramCode} WARN ${param.high_warning} ${param.unit || ''}`.trim(),
            };
          }
        }

        if (crit || warn) return { crit, warn };
      }
      return { crit: null, warn: null };
    },
    formatFullscreenWindowStr() {
      const s = new Date(this.trendTimeBounds.start);
      const e = new Date(this.trendTimeBounds.end);
      return `${s.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })} - ${e.toLocaleString([], { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}`;
    },
    systemStatus() {
      return this.$store.state.systemStatus || {};
    },
    activeAlarmsList() {
      return (this.$store.state.activeAlarms || []).filter(a => a && a.status !== 'CLEARED');
    },
    activeCriticalCount() {
      return this.activeAlarmsList.filter(a => a.severity === 'CRITICAL' && a.status === 'ACTIVE').length;
    },
    activeWarningCount() {
      return this.activeAlarmsList.filter(a => a.severity === 'WARNING' && a.status === 'ACTIVE').length;
    },
    activeAckedCount() {
      return this.activeAlarmsList.filter(a => a.status === 'ACKNOWLEDGED').length;
    },
    highestPriorityAlarm() {
      if (this.activeAlarmsList.length === 0) return null;
      const sorted = [...this.activeAlarmsList].sort((a, b) => {
        if (a.severity === 'CRITICAL' && b.severity !== 'CRITICAL') return -1;
        if (a.severity !== 'CRITICAL' && b.severity === 'CRITICAL') return 1;
        return 0;
      });
      return sorted[0];
    },
    isAudioMuted() {
      const prefs = this.$store.state.alarmPreferences;
      return prefs ? prefs.isMuted : audioAlert.isMuted;
    },
    isSoundEnabled() {
      const prefs = this.$store.state.alarmPreferences;
      return prefs && prefs.soundEnabled !== undefined ? prefs.soundEnabled : audioAlert.soundEnabled;
    },
    overallAppHealth() {
      if (this.systemStatus.database_status && String(this.systemStatus.database_status).toLowerCase().includes('fail')) {
        return 'FAILURE';
      }
      if (this.summaryStats.offlineCount > 0 || this.activeCriticalCount > 0) {
        return 'DEGRADED';
      }
      return 'HEALTHY';
    },
    appHealthColorClass() {
      if (this.overallAppHealth === 'FAILURE') return 'text-rose-600 dark:text-rose-400';
      if (this.overallAppHealth === 'DEGRADED') return 'text-amber-600 dark:text-amber-400';
      return 'text-emerald-600 dark:text-emerald-400';
    },
    appHealthDotClass() {
      if (this.overallAppHealth === 'FAILURE') return 'bg-rose-500 animate-ping';
      if (this.overallAppHealth === 'DEGRADED') return 'bg-amber-500 animate-pulse';
      return 'bg-emerald-500';
    },
    configuredPollingIntervalMs() {
      // 1. If a specific device is selected in the filter, use that device's configured interval
      if (this.selectedDeviceFilter) {
        const dev = (this.devicesList || []).find(d => d.id === Number(this.selectedDeviceFilter));
        const conn = dev?.connection;
        const iv = conn?.polling_interval || conn?.poll_interval_ms;
        if (iv && iv >= 500) return iv;
      }
      // 2. Otherwise find minimum interval among active enabled devices
      const activeDevs = (this.devicesList || []).filter(d => d.enabled !== false && d.status !== 'DISABLED');
      const intervals = [];
      for (const dev of activeDevs) {
        const conn = dev.connection;
        const iv = conn?.polling_interval || conn?.poll_interval_ms;
        if (iv && iv >= 500) {
          intervals.push(iv);
        }
      }
      if (intervals.length > 0) {
        return Math.min(...intervals);
      }
      return 5000;
    },
    cardGridClass() {
      return 'grid gap-1.5 sm:gap-2';
    },
    activePanels() {
      if (Array.isArray(this.panelLayoutConfigs) && this.panelLayoutConfigs.length > 0) {
        return this.panelLayoutConfigs;
      }
      return this.buildDefaultPanels(this.panelCount || 2);
    },
    multiPanelGridClass() {
      const count = (this.activePanels && this.activePanels.length) || this.panelCount || 2;
      if (count <= 1) return 'grid-cols-1';
      if (count === 2) return 'grid-cols-1 md:grid-cols-2';
      if (count === 3) return 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3';
      return 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-4';
    },
    multiPanelContainerStyle() {
      const overhead = this.isTopPanelHidden ? 390 : 515;
      return {
        maxHeight: `calc(100vh - ${overhead}px)`,
        width: '100%',
      };
    },
    panelBoxStyle() {
      const overhead = this.isTopPanelHidden ? 390 : 515;
      return {
        maxHeight: `calc(100vh - ${overhead}px)`,
        display: 'flex',
        flexDirection: 'column',
      };
    },
    panelScrollAreaStyle() {
      const overhead = this.isTopPanelHidden ? 390 : 515;
      return {
        maxHeight: `calc(100vh - ${overhead + 32}px)`,
        overflowY: 'auto',
      };
    },
    cardGridContainerStyle() {
      // Bounded height ensuring internal scroll if parameter count exceeds viewport capacity
      // On 1366x768, overhead from 100vh is ~515px with top header, ~390px in full-width mode
      const overhead = this.isTopPanelHidden ? 390 : 515;
      return {
        maxHeight: `calc(100vh - ${overhead}px)`,
        overflowY: 'auto',
      };
    },
    allParameters() {
      const list = [];
      const nowMs = this.currentTimeMs || Date.now();
      for (const dev of this.devicesList) {
        if (!dev.parameters) continue;
        const isOnline = this.isDeviceActiveOnline(dev);
        const isDevOffline = !isOnline && (
          dev.connection_status === 'OFFLINE' ||
          dev.connection_status === 'DISCONNECTED' ||
          dev.connection_status === 'ERROR' ||
          dev.connection_status === 'FAILED'
        );
        for (const p of dev.parameters) {
          const live = this.liveMap[p.id] || {};
          const isHeld = live.is_held_value !== undefined ? live.is_held_value : !!p.is_current_held;
          const val = live.value !== undefined ? live.value : p.current_value;
          const alarmState = this.determineAlarmState(p, val);

          const receivedAt = live.received_at || p.current_received_at || p.last_updated;
          const staleSec = p.stale_timeout_seconds || 120;
          let isTimeStale = false;
          if (receivedAt) {
            const ageSec = (nowMs - new Date(receivedAt).getTime()) / 1000;
            if (ageSec > staleSec) {
              isTimeStale = true;
            }
          }

          let rawQuality = live.quality || p.current_quality || 'UNKNOWN';
          let finalQuality = rawQuality;

          // Zero-value policy exemption: If zero is valid (SUPPRESS_THRESHOLD) and value is 0.0,
          // do not treat 0.0 as UNCERTAIN or BAD from soft/hard limit checks.
          const isZeroValid = (p.zero_value_alarm_policy === 'SUPPRESS_THRESHOLD') &&
            val !== null && val !== undefined && val !== '' && !isNaN(val) && Math.abs(Number(val)) < 1e-9;

          // If the device is genuinely offline with no communication and no fresh telemetry -> BAD!
          // If the device is online, but data has not arrived past stale timeout -> STALE!
          if (isDevOffline) {
            finalQuality = 'BAD';
          } else if (isTimeStale) {
            if (finalQuality !== 'BAD') {
              finalQuality = 'STALE';
            }
          } else if (isZeroValid && (finalQuality === 'UNCERTAIN' || finalQuality === 'BAD')) {
            finalQuality = 'GOOD';
          }

          const rawReason = live.quality_reason || p.current_quality_reason || '';
          const finalReason = isDevOffline
            ? (live.quality_reason || 'COMMUNICATION_ERROR')
            : (isZeroValid && (rawReason === 'OUT_OF_WARNING_RANGE' || rawReason === 'OUT_OF_HARD_RANGE'))
              ? 'NONE'
              : (isTimeStale && (!live.quality_reason && !p.current_quality_reason))
                ? 'STALE_TIMEOUT'
                : rawReason;

          // Resolve last known valid data timestamp
          let lastValidTime = this.lastGoodTimeMap[p.id] || p.current_received_at || dev.last_data_at || p.last_updated;

          // If device is offline, or parameter is in BAD / STALE quality due to comm failure:
          // Synchronize with station's lastTelemetryTime if lastValidTime is missing or was mistakenly set
          // to an error polling timestamp newer than lastTelemetryTime.
          if (this.lastTelemetryTime && (isDevOffline || finalQuality === 'BAD' || finalQuality === 'STALE')) {
            const teleMs = new Date(this.lastTelemetryTime).getTime();
            if (!lastValidTime || new Date(lastValidTime).getTime() > teleMs) {
              lastValidTime = this.lastTelemetryTime;
            }
          }

          // Display the timestamp of the last valid data received so that operators see consistent duration
          // across all parameters on that device (e.g. 11:22:33, 16m ago) rather than confusing error polling times.
          const finalTime = (isDevOffline || finalQuality === 'BAD' || finalQuality === 'STALE')
            ? (lastValidTime || this.lastTelemetryTime || dev.last_data_at || receivedAt)
            : (live.received_at || receivedAt);

          list.push({
            device_id: dev.id,
            device_code: dev.device_code || dev.code,
            device_name: dev.device_name || dev.name,
            device_status: isOnline ? 'ONLINE' : (dev.connection_status || 'OFFLINE'),
            param_id: p.id || p.param_id,
            param_code: p.parameter_code || p.code || p.param_code || '',
            param_name: p.parameter_name || p.name || p.param_name || p.parameter_code || p.code || p.param_code || '',
            category_id: p.category_id || (p.category ? p.category.id : null),
            category: p.category || null,
            icon: p.icon || '',
            unit: p.unit,
            data_type: p.data_type,
            precision: p.precision !== undefined ? p.precision : 2,
            value: val,
            is_held_value: isHeld,
            formula: p.formula || '',
            quality: finalQuality,
            quality_reason: finalReason,
            low_limit: p.low_critical != null ? p.low_critical : (p.low_warning != null ? p.low_warning : null),
            high_limit: p.high_critical != null ? p.high_critical : (p.high_warning != null ? p.high_warning : null),
            alarm_status: alarmState,
            zero_value_alarm_policy: p.zero_value_alarm_policy || 'NORMAL',
            received_at: finalTime,
          });
        }
      }
      return list;
    },
    filteredParameters() {
      let list = this.allParameters.filter((p) => {
        if (this.selectedDeviceFilter && p.device_id !== Number(this.selectedDeviceFilter)) {
          return false;
        }
        if (this.selectedQualityFilter && p.quality !== this.selectedQualityFilter) {
          return false;
        }
        if (this.selectedAlarmFilter && p.alarm_status !== this.selectedAlarmFilter) {
          return false;
        }
        // Category Filter
        if (this.selectedCategoryFilter && this.selectedCategoryFilter !== 'all') {
          if (this.selectedCategoryFilter === 'uncategorized') {
            if (p.category_id && this.categoryMap[p.category_id] && this.categoryMap[p.category_id].enabled) {
              return false;
            }
          } else {
            if (Number(p.category_id) !== Number(this.selectedCategoryFilter)) {
              return false;
            }
          }
        }
        if (this.filterPinnedOnly && !this.isPinned(p.param_id)) {
          return false;
        }
        if (this.searchQuery) {
          const q = this.searchQuery.toLowerCase();
          const matchCode = (p.param_code || '').toLowerCase().includes(q);
          const matchName = (p.param_name || '').toLowerCase().includes(q);
          const matchDev = (p.device_code || '').toLowerCase().includes(q);
          if (!matchCode && !matchName && !matchDev) return false;
        }
        return true;
      });

      // Pinned parameters always bubble to the top
      list.sort((a, b) => {
        const aPinned = this.isPinned(a.param_id) ? 1 : 0;
        const bPinned = this.isPinned(b.param_id) ? 1 : 0;
        return bPinned - aPinned;
      });

      return list;
    },
    tableColumns() {
      const list = this.filteredParameters || [];
      if (list.length === 0) return [];
      if (list.length <= 4) {
        return [list];
      }
      const mid = Math.ceil(list.length / 2);
      return [
        list.slice(0, mid),
        list.slice(mid),
      ];
    },
    summaryStats() {
      const devs = this.devicesList || [];
      const totalDevs = devs.length;
      const onlineDevs = devs.filter(d => this.isDeviceActiveOnline(d)).length;
      const offlineDevs = totalDevs - onlineDevs;
      const disabledDevs = devs.filter(d => d.admin_status === 'DISABLED' || d.status === 'DISABLED' || d.enabled === false).length;

      const params = this.allParameters;
      const goodParams = params.filter(p => p.quality === 'GOOD').length;
      const staleParams = params.filter(p => p.quality === 'STALE').length;

      return {
        totalDevices: totalDevs,
        onlineCount: onlineDevs,
        offlineCount: offlineDevs,
        disabledCount: disabledDevs,
        goodSensors: goodParams,
        staleSensors: staleParams,
      };
    },
    isTelemetryStale() {
      const intervalSec = Math.max(Math.ceil(this.configuredPollingIntervalMs / 1000), 10);
      return this.lastTelemetrySecondsAgo > (intervalSec * 2.5) || this.summaryStats.staleSensors > 0 || this.summaryStats.goodSensors === 0;
    },
    lastTelemetryTimeStr() {
      if (!this.lastTelemetryTime) return '';
      return new Date(this.lastTelemetryTime).toLocaleTimeString('en-GB');
    },
    canManageCategories() {
      const role = (this.$store.state.user && this.$store.state.user.role) || '';
      return role === 'Administrator' || role === 'Engineer';
    },
    canEditConfig() {
      const u = this.$store.state.user;
      if (!u) return false;
      const role = typeof u.role === 'string'
        ? u.role.toLowerCase()
        : (u.role && u.role.name ? u.role.name.toLowerCase() : '');
      return ['administrator', 'admin', 'superadmin', 'engineer'].includes(role);
    },
    canViewCategories() {
      return true;
    },
    currentMonitoringDomain() {
      return (this.systemStatus && this.systemStatus.monitoring_domain) || 'AQMS';
    },
    isAQMSDomain() {
      return (this.currentMonitoringDomain || 'AQMS').toUpperCase() === 'AQMS';
    },
    monitoringDomainBadgeClass() {
      const d = (this.currentMonitoringDomain || 'AQMS').toUpperCase();
      if (d === 'AQMS') return 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30';
      if (d === 'WQMS') return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/30';
      return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/30';
    },
    activeCategoriesList() {
      const dom = (this.currentMonitoringDomain || 'AQMS').toUpperCase();
      const list = Array.isArray(this.categoriesList) ? this.categoriesList : [];
      return list
        .filter(c => {
          if (!c.enabled) return false;
          if (dom === 'CUSTOM') return true;
          const cDom = (c.domain || 'ALL').toUpperCase();
          return cDom === dom || cDom === 'ALL' || cDom === 'CUSTOM';
        })
        .sort((a, b) => (a.display_order || 0) - (b.display_order || 0));
    },

    categoryMap() {
      const map = {};
      const list = Array.isArray(this.categoriesList) ? this.categoriesList : [];
      for (const c of list) {
        map[c.id] = c;
      }
      return map;
    },
    uncategorizedParamCount() {
      return this.allParameters.filter(p => !p.category_id || !this.categoryMap[p.category_id] || !this.categoryMap[p.category_id].enabled).length;
    },
    categorySections() {
      if (this.selectedCategoryFilter !== 'all') {
        let cat = null;
        if (this.selectedCategoryFilter === 'uncategorized') {
          cat = {
            id: 'uncategorized',
            code: 'uncategorized',
            name: this.$t('operator.uncategorized') || 'Uncategorized',
            color_key: 'slate',
            icon_key: 'box',
            default_column_span: 4,
          };
        } else {
          const list = Array.isArray(this.categoriesList) ? this.categoriesList : [];
          cat = list.find(c => Number(c.id) === Number(this.selectedCategoryFilter));
        }
        if (!cat) return [];
        return [{
          ...cat,
          parameters: this.filteredParameters,
        }];
      }

      const sections = [];
      const usedParamIds = new Set();

      for (const cat of this.activeCategoriesList) {
        const catParams = this.filteredParameters.filter(p => Number(p.category_id) === Number(cat.id));
        if (catParams.length > 0) {
          catParams.forEach(p => usedParamIds.add(p.param_id));
          sections.push({
            ...cat,
            parameters: catParams,
          });
        }
      }

      const uncatParams = this.filteredParameters.filter(p => !usedParamIds.has(p.param_id));
      if (uncatParams.length > 0) {
        sections.push({
          id: 'uncategorized',
          code: 'uncategorized',
          name: this.$t('operator.uncategorized') || 'Uncategorized',
          color_key: 'slate',
          icon_key: 'box',
          default_column_span: 4,
          parameters: uncatParams,
        });
      }

      return sections;
    },
    sidebarAutoHide() {
      return !!((this.$store.state.uiPreferences && this.$store.state.uiPreferences.sidebarAutoHide) || this.$store.state.sidebarAutoHide);
    },
    userRole() {
      const u = this.$store.state.user || {};
      return (u.role || '').toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
  },
  watch: {
    allParameters: {
      handler(newParams) {
        if (!this.trendSelectionRestored && newParams && newParams.length > 0) {
          this.restoreOrInitTrendSelection();
        }
      },
    },
    configuredPollingIntervalMs: {
      handler(newVal, oldVal) {
        if (newVal !== oldVal) {
          this.setupPollingTimer();
        }
      },
    },
    isDark: {
      handler(newDark) {
        if (this.activeTrendSeries && this.activeTrendSeries.length > 0) {
          this.activeTrendSeries.forEach((s, idx) => {
            const p = (this.allParameters || []).find(item => item.param_id === s.parameterId) || s;
            s.color = getParamSeriesColor(p, newDark, idx);
          });
        }
      },
    },
  },
  mounted() {
    this.initCompanyBranding();
    this.initAudioPreferences();
    window.addEventListener('organization-branding-updated', this.onOrganizationBrandingUpdated);
    this.initDashboard();
    this.initTrendResizeObserver();
    this.loadPanelLayoutConfig();
    this.loadGridColumnsConfig();

    // Telemetry Freshness Ticker
    this.tickerInterval = setInterval(() => {
      this.currentTimeMs = Date.now();
      if (this.lastTelemetryTime) {
        this.lastTelemetrySecondsAgo = Math.floor((Date.now() - new Date(this.lastTelemetryTime).getTime()) / 1000);
      }
    }, 1000);

    // Live Digital Clock
    this.clockInterval = setInterval(() => {
      this.currentTime = new Date().toLocaleTimeString('en-GB');
    }, 1000);

    // Background Poll Interval for Telemetry & Alarms (configured dynamically to device interval)
    this.setupPollingTimer();

    // Auto-refresh Trend Analysis every 1 minute (silent background update)
    this.setupTrendAutoRefresh();

    // Listen to real WebSocket Telemetry & Communication events
    window.addEventListener('device-telemetry-event', this.onTelemetryEvent);
    window.addEventListener('device-comm-event', this.onDeviceCommEvent);

    // Listen to synchronized SCADA alarm annunciator pulses (audio + visual synchronized clock)
    window.addEventListener('scada-alarm-pulse', this.handleAlarmPulse);

    // Global document listeners for context menu dismissal and keyboard shortcuts
    document.addEventListener('click', this.onDocumentClick);
    document.addEventListener('contextmenu', this.onDocumentContextMenu);
    document.addEventListener('keydown', this.onDocumentKeyDown);
  },
  beforeDestroy() {
    if (this.tickerInterval) clearInterval(this.tickerInterval);
    if (this.clockInterval) clearInterval(this.clockInterval);
    if (this.pollingInterval) clearInterval(this.pollingInterval);
    if (this.trendAutoRefreshInterval) clearInterval(this.trendAutoRefreshInterval);
    if (this.alarmPulseTimer) clearTimeout(this.alarmPulseTimer);
    if (this.trendResizeObserver) {
      this.trendResizeObserver.disconnect();
      this.trendResizeObserver = null;
    }
    if (this.fullscreenResizeObserver) {
      this.fullscreenResizeObserver.disconnect();
      this.fullscreenResizeObserver = null;
    }
    window.removeEventListener('device-telemetry-event', this.onTelemetryEvent);
    window.removeEventListener('device-comm-event', this.onDeviceCommEvent);
    window.removeEventListener('scada-alarm-pulse', this.handleAlarmPulse);
    window.removeEventListener('organization-branding-updated', this.onOrganizationBrandingUpdated);
    document.removeEventListener('click', this.onDocumentClick);
    document.removeEventListener('contextmenu', this.onDocumentContextMenu);
    document.removeEventListener('keydown', this.onDocumentKeyDown);
  },
  methods: {
    toggleSidebarAutoHide() {
      this.$store.dispatch('toggleSidebarAutoHide');
      this.$emit('toggle-autohide', this.sidebarAutoHide);
    },
    loadGridColumnsConfig() {
      try {
        const saved = localStorage.getItem('operator_monitoring_grid_cols') || localStorage.getItem('operator_console_grid_columns');
        if (saved) {
          const parsed = parseInt(saved, 10);
          if ([1, 2, 3, 4].includes(parsed)) {
            this.gridColumns = parsed;
            this.panelCount = parsed;
            return;
          }
        }
      } catch (e) {
        // Fallback to default
      }
      this.gridColumns = 4;
      this.panelCount = 4;
    },
    setGridColumns(cols) {
      const count = Math.min(4, Math.max(1, parseInt(cols, 10) || 4));
      this.gridColumns = count;
      this.panelCount = count;
      try {
        localStorage.setItem('operator_monitoring_grid_cols', String(count));
        localStorage.setItem('operator_console_grid_columns', String(count));
      } catch (e) {
        // Ignore localStorage error
      }
      if (typeof this.savePanelLayoutConfig === 'function') {
        this.savePanelLayoutConfig();
      }
    },
    getParamCategory(param) {
      if (!param) return null;
      if (param.category) return param.category;
      if (param.category_id && this.categoryMap && this.categoryMap[param.category_id]) {
        return this.categoryMap[param.category_id];
      }
      return null;
    },
    getParamCategoryTheme(param) {
      const cat = this.getParamCategory(param);
      return this.getCategoryTheme(cat ? cat.color_key : 'slate');
    },
    getParamCategoryShortLabel(param) {
      const cat = this.getParamCategory(param);
      if (!cat) return '';
      const name = cat.name || '';
      if (/particulate/i.test(name)) return 'PM';
      if (/weather|meteorolog/i.test(name)) return 'MET';
      if (/gas|pollutant/i.test(name)) return 'GAS';
      if (/custom/i.test(name)) return 'ENV';
      return name.length > 8 ? name.slice(0, 7) + '…' : name;
    },
    loadPanelLayoutConfig() {
      try {
        const saved = localStorage.getItem('occ_multipanel_layout_config');
        if (saved) {
          const parsed = JSON.parse(saved);
          if (parsed && Array.isArray(parsed.panels) && parsed.panels.length > 0) {
            this.panelCount = parsed.panelCount || parsed.panels.length;
            this.panelLayoutConfigs = parsed.panels;
            return;
          }
        }
      } catch (err) {
        console.warn('Failed to load multi-panel configuration from localStorage:', err);
      }
      this.panelCount = 2;
      this.panelLayoutConfigs = this.buildDefaultPanels(2);
    },
    savePanelLayoutConfig() {
      try {
        const payload = {
          panelCount: this.panelCount,
          panels: this.panelLayoutConfigs,
        };
        localStorage.setItem('occ_multipanel_layout_config', JSON.stringify(payload));
      } catch (err) {
        console.error('Failed to save multi-panel config to localStorage:', err);
      }
    },
    setPanelCount(count) {
      this.panelCount = count;
      this.panelLayoutConfigs = this.buildDefaultPanels(count);
      this.savePanelLayoutConfig();
    },
    openPanelConfigModal() {
      this.isPanelConfigModalOpen = true;
    },
    handlePanelConfigSaved(config) {
      if (config && Array.isArray(config.panels)) {
        this.panelCount = config.panelCount || config.panels.length;
        this.panelLayoutConfigs = config.panels;
        this.savePanelLayoutConfig();
      }
    },
    buildDefaultPanels(count) {
      const categories = this.activeCategoriesList || [];
      const gasCat = categories.find(c => /gas|pollutant|udara/i.test(c.name)) || categories[0];
      const partCat = categories.find(c => /particul|debu|pm/i.test(c.name));
      const weatherCat = categories.find(c => /weather|cuaca|meteo/i.test(c.name)) || (categories.length > 1 ? categories[1] : null);

      if (count === 1) {
        return [
          {
            id: 'panel-1',
            title: this.$t('operator.allCategories') || 'All Monitoring Parameters',
            colorKey: 'cyan',
            assignmentMode: 'category',
            categoryIds: categories.map(c => c.id).concat(['uncategorized']),
            parameterIds: [],
            allowDuplicates: false,
          },
        ];
      }

      if (count === 2) {
        const p1Cats = gasCat ? [gasCat.id] : (categories[0] ? [categories[0].id] : [1]);
        const p2Cats = categories.filter(c => !p1Cats.includes(c.id)).map(c => c.id).concat(['uncategorized']);
        return [
          {
            id: 'panel-1',
            title: (gasCat && gasCat.name) || 'Gas & Pollutants',
            colorKey: 'cyan',
            assignmentMode: 'category',
            categoryIds: p1Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
          {
            id: 'panel-2',
            title: (weatherCat && weatherCat.name) || 'Weather & Meteorology',
            colorKey: 'blue',
            assignmentMode: 'category',
            categoryIds: p2Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
        ];
      }

      if (count === 3) {
        const p1Cats = gasCat ? [gasCat.id] : [1];
        const p2Cats = partCat ? [partCat.id] : (categories.length > 2 ? [categories[2].id] : [3]);
        const p3Cats = categories.filter(c => !p1Cats.includes(c.id) && !p2Cats.includes(c.id)).map(c => c.id).concat(['uncategorized']);
        return [
          {
            id: 'panel-1',
            title: (gasCat && gasCat.name) || 'Gas & Pollutants',
            colorKey: 'cyan',
            assignmentMode: 'category',
            categoryIds: p1Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
          {
            id: 'panel-2',
            title: (partCat && partCat.name) || 'Particulate Matter',
            colorKey: 'amber',
            assignmentMode: 'category',
            categoryIds: p2Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
          {
            id: 'panel-3',
            title: (weatherCat && weatherCat.name) || 'Weather & Meteorology',
            colorKey: 'blue',
            assignmentMode: 'category',
            categoryIds: p3Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
        ];
      }

      // count >= 4
      const p1Cats = gasCat ? [gasCat.id] : [1];
      const p2Cats = partCat ? [partCat.id] : [3];
      const p3Cats = weatherCat ? [weatherCat.id] : [2];
      const p4Cats = categories.filter(c => !p1Cats.includes(c.id) && !p2Cats.includes(c.id) && !p3Cats.includes(c.id)).map(c => c.id).concat(['uncategorized']);
      return [
        {
          id: 'panel-1',
          title: (gasCat && gasCat.name) || 'Gas & Pollutants',
          colorKey: 'cyan',
          assignmentMode: 'category',
          categoryIds: p1Cats,
          parameterIds: [],
          allowDuplicates: false,
        },
        {
          id: 'panel-2',
          title: (partCat && partCat.name) || 'Particulate Matter',
          colorKey: 'amber',
          assignmentMode: 'category',
          categoryIds: p2Cats,
          parameterIds: [],
          allowDuplicates: false,
        },
        {
          id: 'panel-3',
          title: (weatherCat && weatherCat.name) || 'Weather & Meteorology',
          colorKey: 'blue',
          assignmentMode: 'category',
          categoryIds: p3Cats,
          parameterIds: [],
          allowDuplicates: false,
        },
        {
          id: 'panel-4',
          title: 'Custom & Environmental Parameters',
          colorKey: 'emerald',
          assignmentMode: 'category',
          categoryIds: p4Cats.length > 0 ? p4Cats : ['uncategorized'],
          parameterIds: [],
          allowDuplicates: false,
        },
      ];
    },
    getPrecedingClaimedIds(currentIndex) {
      const claimed = new Set();
      const pool = this.filteredParameters || [];
      for (let i = 0; i < currentIndex; i++) {
        const p = this.activePanels[i];
        if (p) {
          if (p.assignmentMode === 'custom') {
            (p.parameterIds || []).forEach(id => claimed.add(String(id)));
          } else {
            const catSet = new Set((p.categoryIds || []).map(c => String(c).toLowerCase()));
            pool.forEach(param => {
              if (!param.category_id || param.category_id === 'uncategorized') {
                if (catSet.has('uncategorized')) claimed.add(String(param.param_id));
              } else if (catSet.has(String(param.category_id).toLowerCase())) {
                claimed.add(String(param.param_id));
              }
            });
          }
        }
      }
      return claimed;
    },
    getPanelParameters(panel, panelIndex = -1) {
      if (!panel) return [];
      const pool = this.filteredParameters;
      if (!pool || pool.length === 0) return [];

      const actualIdx = panelIndex >= 0 ? panelIndex : this.activePanels.indexOf(panel);

      let matched = [];
      if (panel.assignmentMode === 'custom') {
        const idSet = new Set((panel.parameterIds || []).map(String));
        matched = pool.filter(p => idSet.has(String(p.param_id)));
      } else {
        const catSet = new Set((panel.categoryIds || []).map(c => String(c).toLowerCase()));
        matched = pool.filter(p => {
          if (!p.category_id || p.category_id === 'uncategorized') {
            return catSet.has('uncategorized');
          }
          return catSet.has(String(p.category_id).toLowerCase());
        });
      }

      // Deduplication check: if panel does NOT allow duplicates, exclude parameters already claimed
      if (!panel.allowDuplicates && actualIdx > 0) {
        const precedingClaimedIds = this.getPrecedingClaimedIds(actualIdx);
        matched = matched.filter(p => !precedingClaimedIds.has(String(p.param_id)));
      }

      // Safety guard: Last panel sweeps in any unassigned parameters so zero parameters are lost
      if (actualIdx === this.activePanels.length - 1) {
        const claimedByOthers = this.getPrecedingClaimedIds(actualIdx);
        const matchedSet = new Set(matched.map(p => String(p.param_id)));
        const unassigned = pool.filter(p => !claimedByOthers.has(String(p.param_id)) && !matchedSet.has(String(p.param_id)));
        if (unassigned.length > 0) {
          matched = matched.concat(unassigned);
        }
      }

      return matched;
    },
    getPanelCardGridStyle(panel) {
      const count = (this.activePanels && this.activePanels.length) || this.panelCount || 2;
      let minCardWidth = 175;
      if (count === 1) minCardWidth = 185;
      else if (count === 2) minCardWidth = 175;
      else if (count === 3) minCardWidth = 155;
      else minCardWidth = 145;
      return {
        gridTemplateColumns: `repeat(auto-fill, minmax(${minCardWidth}px, 1fr))`,
      };
    },
    getPanelAccentDot(colorKey) {
      const map = {
        cyan: 'bg-cyan-500 shadow-xs shadow-cyan-500/50',
        blue: 'bg-blue-500 shadow-xs shadow-blue-500/50',
        emerald: 'bg-emerald-500 shadow-xs shadow-emerald-500/50',
        amber: 'bg-amber-500 shadow-xs shadow-amber-500/50',
        purple: 'bg-purple-500 shadow-xs shadow-purple-500/50',
        rose: 'bg-rose-500 shadow-xs shadow-rose-500/50',
        indigo: 'bg-indigo-500 shadow-xs shadow-indigo-500/50',
        slate: 'bg-slate-400 shadow-xs shadow-slate-400/50',
      };
      return map[colorKey] || map.cyan;
    },
    openParamContextMenu(event, param) {
      if (event) {
        event.preventDefault();
        event.stopPropagation();
      }
      if (!param) return;

      const dev = (this.devicesList || []).find(d => d.id === param.device_id) || {
        id: param.device_id,
        device_code: param.device_code,
        name: param.device_name,
      };
      const fullParam = (dev && dev.parameters)
        ? dev.parameters.find(p => p.id === param.param_id)
        : null;

      const menuWidth = 240;
      const menuHeight = 250;
      const clientX = event && event.clientX ? event.clientX : window.innerWidth / 2;
      const clientY = event && event.clientY ? event.clientY : window.innerHeight / 2;
      const x = Math.min(clientX, window.innerWidth - menuWidth - 12);
      const y = Math.min(clientY, window.innerHeight - menuHeight - 12);

      this.paramContextMenu = {
        visible: true,
        x: Math.max(12, x),
        y: Math.max(12, y),
        param,
        device: dev,
        fullParam: fullParam || param,
      };
    },
    closeParamContextMenu() {
      this.paramContextMenu.visible = false;
    },
    handleContextTestRead() {
      const { device, fullParam, param } = this.paramContextMenu;
      this.closeParamContextMenu();
      if (!param) return;
      this.testReadTarget = {
        device: device || { id: param.device_id, device_code: param.device_code },
        parameter: fullParam || param,
      };
      this.isTestReadModalOpen = true;
    },
    handleContextViewConfig() {
      const { device, fullParam, param } = this.paramContextMenu;
      this.closeParamContextMenu();
      if (!param) return;
      this.viewConfigTarget = {
        device: device || { id: param.device_id, device_code: param.device_code },
        parameter: fullParam || param,
      };
      this.isViewConfigModalOpen = true;
    },
    switchToEditModal() {
      if (!this.canEditConfig) return;
      this.isViewConfigModalOpen = false;
      const { device, parameter } = this.viewConfigTarget;
      this.editParamTarget = {
        deviceId: (device && device.id) || (parameter && parameter.device_id),
        device: device || { id: parameter.device_id, device_code: parameter.device_code },
        parameter: parameter,
      };
      this.isEditParamModalOpen = true;
    },
    handleContextEditConfig() {
      if (!this.canEditConfig) {
        if (this.showAudioNotification) {
          this.showAudioNotification('Access Denied: Only Admin or Superadmin can edit configuration.');
        }
        return;
      }
      const { device, fullParam, param } = this.paramContextMenu;
      this.closeParamContextMenu();
      if (!param) return;
      this.editParamTarget = {
        deviceId: (device && device.id) || param.device_id,
        device: device || { id: param.device_id, device_code: param.device_code },
        parameter: fullParam || param,
      };
      this.isEditParamModalOpen = true;
    },
    handleContextHistory() {
      const { device, fullParam, param } = this.paramContextMenu;
      this.closeParamContextMenu();
      if (!param) return;
      this.historyTarget = {
        device: device || { id: param.device_id, device_code: param.device_code },
        parameter: fullParam || param,
      };
      this.isHistoryModalOpen = true;
    },
    onTestReadCompleted(payload) {
      if (payload && payload.result && payload.result.success) {
        const val = payload.result.formula_value !== undefined ? payload.result.formula_value : payload.result.decoded_value;
        if (val !== undefined && payload.parameter) {
          const pid = payload.parameter.id || payload.parameter.param_id;
          if (pid) {
            this.$set(this.liveMap, pid, {
              ...(this.liveMap[pid] || {}),
              value: val,
              quality: 'GOOD',
              quality_reason: 'TEST_READ_PASS',
              received_at: new Date().toISOString(),
            });
          }
        }
      }
    },
    async handleParamSaved() {
      this.isEditParamModalOpen = false;
      try {
        const res = await apiClient.get('/devices');
        if (res.data && res.data.data) {
          this.devicesList = res.data.data.items || res.data.data || [];
        }
      } catch (err) {
        console.error('Failed to reload devices after param saved:', err);
      }
    },
    onDocumentClick(e) {
      if (this.paramContextMenu && this.paramContextMenu.visible && !e.target.closest('.param-context-menu')) {
        this.closeParamContextMenu();
      }
      if (this.activeToolbarPopover && !e.target.closest('.toolbar-popover-anchor')) {
        this.closeToolbarPopover();
      }
      if (this.showTrendSettingsMenu && !e.target.closest('.trend-settings-anchor')) {
        this.showTrendSettingsMenu = false;
      }
      if (this.showTrendDownloadMenu && !e.target.closest('.trend-download-anchor')) {
        this.showTrendDownloadMenu = false;
      }
      if (this.showAudioSetupPopover && !e.target.closest('.audio-setup-anchor')) {
        this.showAudioSetupPopover = false;
      }
    },
    onDocumentKeyDown(e) {
      if (e.key === 'Escape') {
        if (this.paramContextMenu && this.paramContextMenu.visible) this.closeParamContextMenu();
        if (this.activeToolbarPopover) this.closeToolbarPopover();
        if (this.showTrendSettingsMenu) this.showTrendSettingsMenu = false;
        if (this.showTrendDownloadMenu) this.showTrendDownloadMenu = false;
        if (this.showAudioSetupPopover) this.showAudioSetupPopover = false;
        if (this.showLogoModal) this.closeLogoModal();
        if (this.showBrandingModal) this.closeBrandingModal();
      }
    },
    toggleToolbarPopover(popoverName) {
      if (this.activeToolbarPopover === popoverName) {
        this.activeToolbarPopover = null;
      } else {
        this.activeToolbarPopover = popoverName;
      }
    },
    async initCompanyBranding() {
      try {
        const data = await loadOrganizationBranding();
        if (data) {
          if (data.customer) {
            this.customerCompanyName = data.customer.name !== undefined ? data.customer.name : DEFAULT_BRANDING.customer.name;
            this.customerSubtitle = (data.customer.subtitle !== undefined && data.customer.subtitle !== null)
              ? (data.customer.subtitle === 'Local Autonomous Edge' ? '' : data.customer.subtitle)
              : '';
            this.customerLogoUrl = data.customer.logoUrl || null;
          }
          if (data.maker) {
            this.makerCompanyName = data.maker.name !== undefined ? data.maker.name : '';
            this.makerSubtitle = data.maker.subtitle !== undefined ? data.maker.subtitle : '';
            this.makerLogoUrl = data.maker.logoUrl || null;
            // Legacy aliases
            this.companyLogoUrl = this.makerLogoUrl;
            this.companyName = this.makerCompanyName;
          }
        }
      } catch (e) {
        console.warn('Failed to load organization branding:', e);
      }
    },
    onOrganizationBrandingUpdated(e) {
      if (e && e.detail) {
        if (e.detail.customer) {
          if (e.detail.customer.name !== undefined) this.customerCompanyName = e.detail.customer.name;
          if (e.detail.customer.subtitle !== undefined) {
            this.customerSubtitle = e.detail.customer.subtitle === 'Local Autonomous Edge' ? '' : e.detail.customer.subtitle;
          }
        }
        if (e.detail.maker) {
          if (e.detail.maker.name !== undefined) this.makerCompanyName = e.detail.maker.name;
          if (e.detail.maker.subtitle !== undefined) this.makerSubtitle = e.detail.maker.subtitle;
          this.companyName = this.makerCompanyName;
        }
      }
      this.initCompanyBranding();
    },
    openBrandingModal(tab = 'customer') {
      this.activeBrandingTab = tab;
      this.modalCustomerName = this.customerCompanyName !== undefined ? this.customerCompanyName : '';
      this.modalCustomerSubtitle = (this.customerSubtitle !== undefined && this.customerSubtitle !== null)
        ? (this.customerSubtitle === 'Local Autonomous Edge' ? '' : this.customerSubtitle)
        : '';
      this.modalCustomerLogoPreviewUrl = this.customerLogoUrl;
      this.modalCustomerLogoFile = null;

      this.modalMakerName = this.makerCompanyName !== undefined ? this.makerCompanyName : '';
      this.modalMakerSubtitle = this.makerSubtitle !== undefined ? this.makerSubtitle : '';
      this.modalMakerLogoPreviewUrl = this.makerLogoUrl;
      this.modalMakerLogoFile = null;

      // Legacy fallback
      this.modalCompanyName = tab === 'customer' ? this.customerCompanyName : this.makerCompanyName;
      this.logoPreviewUrl = tab === 'customer' ? this.customerLogoUrl : this.makerLogoUrl;
      this.logoTempFile = null;
      this.logoErrorMsg = '';

      this.brandingErrorMsg = '';
      this.showBrandingModal = true;
      this.showLogoModal = true;
    },
    closeBrandingModal() {
      this.showBrandingModal = false;
      this.showLogoModal = false;
      this.modalCustomerLogoFile = null;
      this.modalMakerLogoFile = null;
      this.logoTempFile = null;
      this.brandingErrorMsg = '';
      this.logoErrorMsg = '';
    },
    openLogoModal(target) {
      this.openBrandingModal(target || 'customer');
    },
    closeLogoModal() {
      this.closeBrandingModal();
    },
    triggerCustomerFileInput() {
      if (this.$refs.customerLogoFileInput) {
        this.$refs.customerLogoFileInput.click();
      }
    },
    onCustomerFileSelected(e) {
      const file = e.target.files && e.target.files[0];
      if (!file) return;
      this.brandingErrorMsg = '';
      if (file.size > 1024 * 1024) {
        this.brandingErrorMsg = 'File size exceeds 1 MB limit.';
        return;
      }
      const allowed = ['image/png', 'image/jpeg', 'image/jpg', 'image/svg+xml', 'image/webp'];
      if (!allowed.includes(file.type)) {
        this.brandingErrorMsg = 'Invalid image type. Please select PNG, JPG, SVG, or WebP.';
        return;
      }
      this.modalCustomerLogoFile = file;
      const reader = new FileReader();
      reader.onload = (ev) => {
        this.modalCustomerLogoPreviewUrl = ev.target.result;
      };
      reader.readAsDataURL(file);
    },
    async handleRemoveCustomerLogo() {
      try {
        await removeCustomerLogo();
        this.customerLogoUrl = null;
        this.modalCustomerLogoPreviewUrl = null;
        this.modalCustomerLogoFile = null;
        if (this.$refs.customerLogoFileInput) this.$refs.customerLogoFileInput.value = '';
      } catch (err) {
        this.brandingErrorMsg = err.message || 'Failed to remove customer logo.';
      }
    },
    triggerMakerFileInput() {
      if (this.$refs.makerLogoFileInput) {
        this.$refs.makerLogoFileInput.click();
      }
    },
    onMakerFileSelected(e) {
      const file = e.target.files && e.target.files[0];
      if (!file) return;
      this.brandingErrorMsg = '';
      if (file.size > 1024 * 1024) {
        this.brandingErrorMsg = 'File size exceeds 1 MB limit.';
        return;
      }
      const allowed = ['image/png', 'image/jpeg', 'image/jpg', 'image/svg+xml', 'image/webp'];
      if (!allowed.includes(file.type)) {
        this.brandingErrorMsg = 'Invalid image type. Please select PNG, JPG, SVG, or WebP.';
        return;
      }
      this.modalMakerLogoFile = file;
      const reader = new FileReader();
      reader.onload = (ev) => {
        this.modalMakerLogoPreviewUrl = ev.target.result;
      };
      reader.readAsDataURL(file);
    },
    async handleRemoveMakerLogo() {
      try {
        await removeMakerLogo();
        this.makerLogoUrl = null;
        this.modalMakerLogoPreviewUrl = null;
        this.modalMakerLogoFile = null;
        this.companyLogoUrl = null;
        if (this.$refs.makerLogoFileInput) this.$refs.makerLogoFileInput.value = '';
      } catch (err) {
        this.brandingErrorMsg = err.message || 'Failed to remove maker logo.';
      }
    },
    async saveBrandingFromModal() {
      this.brandingSaving = true;
      this.logoSaving = true;
      this.brandingErrorMsg = '';
      try {
        // Save Customer Branding
        await saveCustomerBranding({
          file: this.modalCustomerLogoFile,
          name: this.modalCustomerName,
          subtitle: this.modalCustomerSubtitle
        });

        // Save Maker Branding
        await saveMakerBranding({
          file: this.modalMakerLogoFile,
          name: this.modalMakerName,
          subtitle: this.modalMakerSubtitle
        });

        await this.initCompanyBranding();
        this.closeBrandingModal();
      } catch (err) {
        this.brandingErrorMsg = err.message || 'Failed to save organization branding.';
        this.logoErrorMsg = this.brandingErrorMsg;
      } finally {
        this.brandingSaving = false;
        this.logoSaving = false;
      }
    },
    async saveLogoConfig() {
      return this.saveBrandingFromModal();
    },
    async handleRemoveLogo() {
      if (this.activeBrandingTab === 'customer') {
        return this.handleRemoveCustomerLogo();
      } else {
        return this.handleRemoveMakerLogo();
      }
    },
    triggerLogoFileInput() {
      if (this.activeBrandingTab === 'customer') {
        this.triggerCustomerFileInput();
      } else {
        this.triggerMakerFileInput();
      }
    },
    onLogoFileSelected(e) {
      if (this.activeBrandingTab === 'customer') {
        this.onCustomerFileSelected(e);
      } else {
        this.onMakerFileSelected(e);
      }
    },
    hasDuplicateTrendParamName(s) {
      if (!s || !this.activeTrendSeries) return false;
      const name = s.paramName || s.paramCode;
      return this.activeTrendSeries.filter(item => (item.paramName || item.paramCode) === name).length > 1;
    },
    getCompactTrendParamName(s) {
      if (!s) return '';
      const raw = String(s.paramName || s.paramCode || '').trim();
      const upper = raw.toUpperCase();

      // Common acronyms and short codes to return directly
      if (['O3', 'SO2', 'NO2', 'CO', 'HC', 'NO', 'NOX', 'VOC', 'PM1', 'TSP', 'TVOC', 'CH4', 'NMHC', 'NH3', 'H2S'].includes(upper)) {
        return upper;
      }
      // PM mappings
      if (upper.includes('PM 2.5') || upper.includes('PM2.5') || upper === 'PM25') {
        return 'PM2.5';
      }
      if (upper.includes('PM 10') || upper.includes('PM10')) {
        return 'PM10';
      }
      // Temperature
      if (upper.includes('TEMPERATUR') || upper.includes('SUHU') || upper === 'TEMP') {
        return 'TEMP';
      }
      // Pressure
      if (upper.includes('PRESSURE') || upper.includes('TEKANAN') || upper === 'PRESS' || upper === 'BP') {
        return 'PRESS';
      }
      // Solar Radiation
      if (upper.includes('SOLAR') || upper.includes('RADIASI') || upper === 'SR') {
        return 'SOLAR';
      }
      // Wind Speed
      if (upper.includes('WIND SPEED') || upper.includes('KECEPATAN ANGIN') || upper === 'WS') {
        return 'WIND';
      }
      // Wind Direction / Angle
      if (upper.includes('WIND DIRECTION') || upper.includes('WIND ANGLE') || upper.includes('ARAH ANGIN') || upper === 'WD' || upper === 'WDIR' || upper === 'W-DIR') {
        return 'W-DIR';
      }
      // Humidity
      if (upper.includes('HUMIDIT') || upper.includes('KELEMBAPAN') || upper === 'HUM' || upper === 'RH') {
        return 'HUM';
      }
      // Battery / Power
      if (upper.includes('BATTERY') || upper === 'BATTV') {
        return 'BATTV';
      }
      // Flow Rate
      if (upper.includes('FLOW')) {
        return 'FLOW';
      }
      // Opacity
      if (upper.includes('OPACITY')) {
        return 'OPACITY';
      }
      // Fallback: If length <= 6 characters, keep it; else truncate to 6
      if (raw.length <= 6) return raw;
      return raw.slice(0, 5).toUpperCase();
    },
    closeToolbarPopover() {
      this.activeToolbarPopover = null;
    },
    toggleIspuBarVisibility() {
      this.isIspuBarVisible = !this.isIspuBarVisible;
      try {
        localStorage.setItem('operator_ispu_bar_visible', this.isIspuBarVisible ? 'true' : 'false');
      } catch (e) {}
    },
    clearAllPinned() {
      this.pinnedIds = [];
      try {
        localStorage.setItem('dl_pinned_params', JSON.stringify([]));
      } catch (e) {}
    },
    navigateToRoute(route) {
      this.closeToolbarPopover();
      if (this.$route.path !== route) {
        this.$router.push(route).catch(() => {});
      }
    },
    onDocumentContextMenu(e) {
      if (!this.paramContextMenu.visible) return;
      if (!e.target.closest('.scada-param-tile') && !e.target.closest('.param-context-menu')) {
        this.closeParamContextMenu();
      }
    },
    initTrendResizeObserver() {
      this.$nextTick(() => {
        if (this.$refs.scadaChartContainer) {
          const rect = this.$refs.scadaChartContainer.getBoundingClientRect();
          if (rect && rect.width > 0) {
            this.trendChartWidth = Math.round(rect.width);
            this.trendChartHeight = Math.round(rect.height) || 125;
          }
          if (typeof window !== 'undefined' && window.ResizeObserver && !this.trendResizeObserver) {
            this.trendResizeObserver = new ResizeObserver((entries) => {
              for (const entry of entries) {
                const cr = entry.contentRect;
                if (cr && cr.width > 0) {
                  this.trendChartWidth = Math.round(cr.width);
                  this.trendChartHeight = Math.round(cr.height) || 125;
                }
              }
            });
            this.trendResizeObserver.observe(this.$refs.scadaChartContainer);
          }
        }
      });
    },
    initFullscreenResizeObserver() {
      this.$nextTick(() => {
        if (this.$refs.fullscreenChartContainer) {
          const rect = this.$refs.fullscreenChartContainer.getBoundingClientRect();
          if (rect && rect.width > 0) {
            this.fullscreenChartWidth = Math.round(rect.width);
            this.fullscreenChartHeight = Math.round(rect.height) || 440;
          }
          if (typeof window !== 'undefined' && window.ResizeObserver && !this.fullscreenResizeObserver) {
            this.fullscreenResizeObserver = new ResizeObserver((entries) => {
              for (const entry of entries) {
                const cr = entry.contentRect;
                if (cr && cr.width > 0) {
                  this.fullscreenChartWidth = Math.round(cr.width);
                  this.fullscreenChartHeight = Math.round(cr.height) || 440;
                }
              }
            });
            this.fullscreenResizeObserver.observe(this.$refs.fullscreenChartContainer);
          }
        }
      });
    },
    isParamDomainMismatched(param) {
      if (!param || !param.category_id) return false;
      const dom = (this.currentMonitoringDomain || 'AQMS').toUpperCase();
      if (dom === 'CUSTOM') return false;
      const cat = this.categoryMap[param.category_id];
      if (!cat) return false;
      const cDom = (cat.domain || 'ALL').toUpperCase();
      return cDom !== 'ALL' && cDom !== 'CUSTOM' && cDom !== dom;
    },
    getCategoryTheme(colorKey) {
      return getCategoryColor(colorKey);
    },
    getCategoryParamCount(catId) {
      return this.allParameters.filter(p => Number(p.category_id) === Number(catId)).length;
    },
    getSectionContainerSpanClass(section) {
      return 'w-full';
    },
    getSectionGridClass(section) {
      return 'grid gap-1.5 sm:gap-2';
    },
    openCategoryManagementModal() {
      this.isCategoryModalOpen = true;
    },
    async onCategoryConfigSaved() {
      await Promise.all([
        this.fetchCategories(),
        this.fetchDevices(),
      ]);
    },
    async fetchCategories() {
      try {
        const res = await apiClient.get('/categories');
        const raw = res.data?.data || res.data;
        this.categoriesList = Array.isArray(raw) ? raw : [];
      } catch (err) {
        console.error('Failed to fetch categories:', err);
      }
    },
    handleAlarmPulse(e) {
      this.isAlarmPulseActive = true;
      this.alarmPulseSeverity = e && e.detail ? e.detail.severity : 'WARNING';
      if (this.alarmPulseTimer) {
        clearTimeout(this.alarmPulseTimer);
      }
      this.alarmPulseTimer = setTimeout(() => {
        this.isAlarmPulseActive = false;
      }, 350);
    },
    toggleTheme() {
      const nextTheme = this.isDark ? 'light' : 'dark';
      this.$store.dispatch('setTheme', nextTheme);
    },
    toggleLanguage() {
      const nextLang = this.currentLocale === 'en' ? 'id' : 'en';
      this.$store.dispatch('setLanguage', nextLang);
    },
    determineAlarmState(p, val) {
      if (val == null || isNaN(val)) return 'NORMAL';
      const num = Number(val);
      const policy = p.zero_value_alarm_policy || 'NORMAL';
      if (num === 0.0) {
        if (policy === 'SUPPRESS_THRESHOLD') {
          return 'NORMAL';
        }
        if (policy === 'INVALID_READING') {
          return 'WARNING';
        }
      }
      if (p.high_critical != null && num >= p.high_critical) return 'CRITICAL';
      if (p.low_critical != null && num <= p.low_critical) return 'CRITICAL';
      if (p.high_warning != null && num >= p.high_warning) return 'WARNING';
      if (p.low_warning != null && num <= p.low_warning) return 'WARNING';
      return 'NORMAL';
    },
    isParamAlarmOrStaleBad(param) {
      if (!param) return false;
      // 1. If parameter is GOOD and has no alarm, it MUST NOT have gray card or red border
      if (param.quality === 'GOOD' && (!param.alarm_status || param.alarm_status === 'NORMAL')) {
        return false;
      }
      // Zero-value policy exemption: If zero is valid and value is 0, ignore UNCERTAIN or soft limit alarms
      const isZeroVal = param.value !== null && param.value !== undefined && param.value !== '' && Math.abs(Number(param.value)) < 1e-9;
      if (isZeroVal && param.zero_value_alarm_policy === 'SUPPRESS_THRESHOLD') {
        if (param.alarm_status === 'CRITICAL' || param.quality === 'BAD' || param.quality === 'STALE') {
          return true;
        }
        return false;
      }
      // 2. If parameter itself has an active threshold alarm (CRITICAL / WARNING)
      if (param.alarm_status === 'CRITICAL' || param.alarm_status === 'WARNING') {
        return true;
      }
      // 3. If parameter data quality is STALE, BAD, or UNCERTAIN
      if (param.quality === 'STALE' || param.quality === 'BAD' || param.quality === 'UNCERTAIN') {
        return true;
      }
      // 4. If an active alarm specifically targets this parameter
      if (this.activeAlarmsList && this.activeAlarmsList.length > 0) {
        const hasSpecificParamAlarm = this.activeAlarmsList.some(a => {
          if (a.status !== 'ACTIVE') return false;
          if (a.parameter_id && a.parameter_id === param.param_id) return true;
          if (a.parameter && (a.parameter.parameter_code === param.param_code || a.parameter.code === param.param_code) && (a.device_id === param.device_id)) return true;
          return false;
        });
        if (hasSpecificParamAlarm) return true;
      }
      return false;
    },
    isParamDataQualityIssue(param) {
      if (!param) return false;
      // 1. Device communication offline / error
      if (param.device_status === 'OFFLINE' || param.device_status === 'ERROR' || param.device_status === 'DISCONNECTED') {
        return true;
      }
      // 2. Data quality is not GOOD
      if (param.quality && param.quality !== 'GOOD') {
        if (param.quality === 'UNCERTAIN' && param.zero_value_alarm_policy === 'SUPPRESS_THRESHOLD') {
          return false;
        }
        return true;
      }
      // 3. Telemetry value missing / null / empty
      if (param.value == null || param.value === '') {
        return true;
      }
      // 4. Parameter marked stale
      if (param.is_stale) {
        return true;
      }
      // 5. Explicit data-quality alarm from alarms list (COMMUNICATION, SENSOR_FAULT, etc. - NOT high/low limit alarms)
      if (this.activeAlarmsList && this.activeAlarmsList.length > 0) {
        const hasDataQualityAlarm = this.activeAlarmsList.some(a => {
          if (a.status !== 'ACTIVE') return false;
          const matchParam = (a.parameter_id && a.parameter_id === param.param_id) ||
            (a.parameter && (a.parameter.parameter_code === param.param_code || a.parameter.code === param.param_code) && (a.device_id === param.device_id));
          if (!matchParam) return false;
          const type = (a.alarm_type || a.type || '').toUpperCase();
          return type.includes('COMM') || type.includes('STALE') || type.includes('QUALITY') || type.includes('TIMEOUT') || type.includes('OFFLINE') || type.includes('DISCONNECT');
        });
        if (hasDataQualityAlarm) return true;
      }
      return false;
    },
    getParamQualityReasonText(param) {
      if (!param) return '';
      if (param.device_status === 'OFFLINE' || param.device_status === 'ERROR') {
        return 'Dev Offline';
      }
      if (param.quality_reason) {
        return param.quality_reason;
      }
      if (param.quality === 'BAD') return 'Bad Quality';
      if (param.quality === 'STALE') return 'Stale Data';
      if (param.quality === 'UNCERTAIN') return 'Uncertain';
      if (param.value == null || param.value === '') return 'No Data';
      return '';
    },
    formatParamValue(param) {
      if (!param || param.value == null || param.value === '') return '--';
      if (param.value_text) return String(param.value_text);
      if (typeof param.value === 'string' && /[eE]/.test(param.value)) {
        return param.value;
      }
      const num = Number(param.value);
      if (isNaN(num)) return String(param.value);
      if (param.precision != null && param.precision >= 0) {
        return num.toFixed(param.precision);
      }
      const strVal = String(param.value);
      if (strVal.includes('.')) {
        const decimals = strVal.split('.')[1].length;
        if (decimals > 2) {
          return num.toFixed(decimals);
        }
      }
      return num.toFixed(2);
    },
    getSectionFlexClass(section) {
      if (this.selectedCategoryFilter !== 'all') {
        return 'w-full';
      }
      const count = (section.parameters || []).length;
      if (count >= 4) {
        return 'w-full';
      }
      if (count >= 2) {
        return 'flex-1 min-w-[240px] max-w-full';
      }
      return 'flex-1 min-w-[150px] max-w-full';
    },
    getParamValueFontSizeClass(param) {
      if (!param) return 'text-[16px]';
      const formatted = this.formatParamValue(param);
      const str = String(formatted);
      const len = str.length;
      if (len <= 5) {
        // e.g. "0.00", "77.00" -> prominent 17-18px
        return 'text-[17px] sm:text-[18px]';
      } else if (len <= 7) {
        // e.g. "936.47", "-12.34" -> 15-16px
        return 'text-[15px] sm:text-[16px]';
      } else if (len <= 9) {
        // e.g. "1013.25" -> 13-14px
        return 'text-[13px] sm:text-[14px]';
      } else {
        return 'text-[12px] sm:text-[13px]';
      }
    },
    paramValueColorClass(param) {
      if (param.alarm_status === 'CRITICAL') return 'text-[#C93449] dark:text-[#FF6577] font-black';
      if (param.alarm_status === 'WARNING') return 'text-[#A96C0A] dark:text-[#F1BD61] font-bold';
      // test token match: text-[#20364B] dark:text-[#E2EBF4]
      return 'text-[#15324F] dark:text-[#E6F3FC] font-bold';
    },
    qualityBadgeClass(q) {
      switch (q) {
        case 'GOOD':
          return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/30';
        case 'UNCERTAIN':
          return 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/30';
        case 'BAD':
          return 'bg-rose-500/10 text-rose-700 dark:text-rose-400 border border-rose-500/30';
        case 'STALE':
          return 'bg-purple-500/10 text-purple-700 dark:text-purple-400 border border-purple-500/30';
        default:
          return 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700';
      }
    },
    getNotifStatusClass(status) {
      switch (status) {
        case 'SENT':
        case 'DELIVERED':
          return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/30';
        case 'PENDING':
        case 'SENDING':
          return 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-400 border border-cyan-500/30';
        case 'FAILED':
          return 'bg-rose-500/10 text-rose-700 dark:text-rose-400 border border-rose-500/30';
        case 'RETRYING':
          return 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/30';
        default:
          return 'bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-400';
      }
    },
    getTrendSeriesColor(param) {
      if (!param) return this.isDark ? '#27D9F2' : '#0284C7';
      const s = this.activeTrendSeries.find(item => item.parameterId === param.param_id);
      if (s && s.color) return s.color;
      return getParamSeriesColor(param, this.isDark);
    },
    getSparklineColor(param) {
      if (!param) return this.isDark ? '#27D9F2' : '#0284C7';
      if (param.alarm_status === 'CRITICAL' || param.quality === 'BAD') return this.isDark ? '#FF6577' : '#C93449';
      if (param.alarm_status === 'WARNING' || param.quality === 'UNCERTAIN') return this.isDark ? '#F1BD61' : '#A96C0A';
      if (param.quality === 'STALE') return '#A855F7';
      return getParamSeriesColor(param, this.isDark);
    },
    recordSparkline(paramId, val, time, quality = 'GOOD') {
      if (paramId == null) return;
      if (!this.sparklinesMap[paramId]) {
        this.$set(this.sparklinesMap, paramId, []);
      }
      const arr = this.sparklinesMap[paramId];
      const num = (val !== null && val !== undefined && val !== '' && !isNaN(val)) ? Number(val) : null;
      const ts = time ? new Date(time).getTime() : Date.now();

      // If last item has same or near-identical timestamp, update in place
      if (arr.length > 0 && Math.abs(arr[arr.length - 1].time - ts) < 500) {
        arr[arr.length - 1].val = num;
        arr[arr.length - 1].quality = quality;
      } else {
        arr.push({ val: num, time: ts, quality });
        // Bound the number of sparkline points to keep memory and rendering lightweight
        if (arr.length > 25) {
          arr.shift();
        }
      }
      this.recomputeSparklineMeta(paramId);
    },
    async fetchInitialSparklines() {
      for (const dev of this.devicesList || []) {
        if (!dev.id) continue;
        try {
          const res = await axios.get(`/api/devices/${dev.id}/telemetry/history?page_size=120`);
          if (res.data?.data?.items) {
            const items = res.data.data.items;
            const grouped = {};
            for (const item of items) {
              const pid = item.parameter_id;
              if (!pid) continue;
              if (!grouped[pid]) grouped[pid] = [];
              const raw = item.processed_value !== undefined ? item.processed_value : item.value;
              const num = (raw !== null && raw !== undefined && raw !== '' && !isNaN(raw)) ? Number(raw) : null;
              const ts = new Date(item.received_at || item.timestamp || 0).getTime();
              grouped[pid].push({ val: num, time: ts, quality: item.quality || 'GOOD' });
            }
            for (const pid in grouped) {
              // Preserve chronological order (ascending timestamp)
              grouped[pid].sort((a, b) => a.time - b.time);
              const sliced = grouped[pid].slice(-25);
              this.$set(this.sparklinesMap, pid, sliced);
              this.recomputeSparklineMeta(pid);
            }
          }
        } catch (e) {
          // silent fallback
        }
      }
    },
    recomputeSparklineMeta(paramId) {
      const rawList = this.sparklinesMap[paramId] || [];
      const validItems = rawList.filter(item => item && item.val !== null && item.val !== undefined && isFinite(item.val));

      if (validItems.length === 0) {
        this.$set(this.sparklineMetaCache, paramId, {
          hasHistory: false,
          validCount: 0,
          strokePaths: [],
          areaPaths: [],
          min: 0,
          max: 0,
        });
        return;
      }

      const W = 100;
      const H = 20;
      const topPad = 3.5;
      const bottomPad = 3.5;
      const plotH = H - topPad - bottomPad; // 13.0

      // Single item: Show clean centered horizontal line
      if (validItems.length === 1) {
        const val = validItems[0].val;
        const y = 10.0;
        this.$set(this.sparklineMetaCache, paramId, {
          hasHistory: true,
          validCount: 1,
          strokePaths: [`M 0,${y} L 100,${y}`],
          areaPaths: [`M 0,${H} L 0,${y} L 100,${y} L 100,${H} Z`],
          min: Number(val.toFixed(2)),
          max: Number(val.toFixed(2)),
        });
        return;
      }

      let min = Infinity;
      let max = -Infinity;
      for (const item of validItems) {
        if (item.val < min) min = item.val;
        if (item.val > max) max = item.val;
      }

      let span = max - min;
      const isConstant = span === 0;

      if (isConstant) {
        min = min - 1;
        max = max + 1;
        span = 2;
      } else {
        // Generous vertical padding (18%) so peaks and troughs never sink or clip
        const pad = Math.max(span * 0.18, 0.01);
        min -= pad;
        max += pad;
        span = max - min;
      }

      const total = rawList.length;
      const stepX = total > 1 ? W / (total - 1) : W;

      const segments = [];
      let currentSegment = [];

      for (let i = 0; i < total; i++) {
        const item = rawList[i];
        const isValValid = item && item.val !== null && item.val !== undefined && isFinite(item.val);

        // Detect substantial time gap (> 4x standard sample spacing) as visual break
        let isTimeGap = false;
        if (isValValid && currentSegment.length > 0) {
          const prevItem = currentSegment[currentSegment.length - 1].raw;
          if (item.time && prevItem.time && (item.time - prevItem.time > 20000)) {
            isTimeGap = true;
          }
        }

        if (isValValid && !isTimeGap) {
          const x = i * stepX;
          const norm = isConstant ? 0.5 : (item.val - min) / span;
          const y = (H - bottomPad) - norm * plotH;
          currentSegment.push({ x: Number(x.toFixed(1)), y: Number(y.toFixed(1)), raw: item });
        } else {
          if (currentSegment.length > 0) {
            segments.push(currentSegment);
            currentSegment = [];
          }
          if (isValValid) {
            const x = i * stepX;
            const norm = isConstant ? 0.5 : (item.val - min) / span;
            const y = (H - bottomPad) - norm * plotH;
            currentSegment.push({ x: Number(x.toFixed(1)), y: Number(y.toFixed(1)), raw: item });
          }
        }
      }
      if (currentSegment.length > 0) {
        segments.push(currentSegment);
      }

      const strokePaths = [];
      const areaPaths = [];

      for (const seg of segments) {
        if (seg.length === 1) {
          const p = seg[0];
          strokePaths.push(`M 0,${p.y} L 100,${p.y}`);
          areaPaths.push(`M 0,${H} L 0,${p.y} L 100,${p.y} L 100,${H} Z`);
        } else if (seg.length > 1) {
          const strokeD = seg.map((p, idx) => (idx === 0 ? `M ${p.x},${p.y}` : `L ${p.x},${p.y}`)).join(' ');
          strokePaths.push(strokeD);

          const first = seg[0];
          const last = seg[seg.length - 1];
          const areaD = `M ${first.x},${H} L ${seg.map(p => `${p.x},${p.y}`).join(' L ')} L ${last.x},${H} Z`;
          areaPaths.push(areaD);
        }
      }

      this.$set(this.sparklineMetaCache, paramId, {
        hasHistory: true,
        validCount: validItems.length,
        strokePaths,
        areaPaths,
        min: Number(min.toFixed(2)),
        max: Number(max.toFixed(2)),
      });
    },
    getSparkline(paramId) {
      if (this.sparklineMetaCache[paramId]) {
        return this.sparklineMetaCache[paramId];
      }
      this.recomputeSparklineMeta(paramId);
      return this.sparklineMetaCache[paramId] || { hasHistory: false, strokePaths: [], areaPaths: [] };
    },
    getSparklinePoints(paramId) {
      const meta = this.getSparkline(paramId);
      if (!meta.hasHistory || meta.strokePaths.length === 0) return '';
      return meta.strokePaths[0].replace(/M |L /g, '');
    },
    getSparklinePolygon(paramId) {
      const meta = this.getSparkline(paramId);
      if (!meta.hasHistory || meta.areaPaths.length === 0) return '';
      return meta.areaPaths[0].replace(/M |L |Z/g, '');
    },
    isDeviceActiveOnline(dev) {
      if (!dev) return false;
      const status = dev.connection_status || dev.status;
      // If explicitly marked offline, error, disconnected, failed, or disabled, it is NOT online
      if (
        status === 'OFFLINE' ||
        status === 'DISCONNECTED' ||
        status === 'ERROR' ||
        status === 'FAILED' ||
        status === 'DISABLED' ||
        dev.enabled === false
      ) {
        return false;
      }
      if (status === 'ONLINE' || status === 'CONNECTED') {
        // If device has parameters configured, verify that not ALL parameters are in STALE/BAD state
        const params = dev.parameters || [];
        if (params.length > 0) {
          const nowMs = this.currentTimeMs || Date.now();
          const hasAnyFreshData = params.some(p => {
            const live = this.liveMap[p.id];
            const q = live?.quality || p.current_quality;
            const ts = live?.received_at || p.current_received_at;
            const ageSec = ts ? (nowMs - new Date(ts).getTime()) / 1000 : Infinity;
            const staleLimit = p.stale_timeout_seconds || 120;
            return q !== 'STALE' && q !== 'BAD' && ageSec < staleLimit;
          });
          if (!hasAnyFreshData) {
            return false;
          }
        }
        return true;
      }
      // Fallback only if status is empty/unknown: verify fresh live telemetry arrived within 15 seconds
      const nowMs = this.currentTimeMs || Date.now();
      return (dev.parameters || []).some(p => {
        const live = this.liveMap[p.id];
        const ts = live?.received_at || p.current_received_at;
        if (!ts) return false;
        const ageSec = (nowMs - new Date(ts).getTime()) / 1000;
        return ageSec < 15 && live?.quality === 'GOOD';
      });
    },
    setupPollingTimer() {
      if (this.pollingInterval) {
        clearInterval(this.pollingInterval);
        this.pollingInterval = null;
      }
      const intervalMs = this.configuredPollingIntervalMs;
      this.pollingInterval = setInterval(async () => {
        await this.pollDashboardCycle();
      }, intervalMs);
    },
    setupTrendAutoRefresh() {
      if (this.trendAutoRefreshInterval) {
        clearInterval(this.trendAutoRefreshInterval);
        this.trendAutoRefreshInterval = null;
      }
      // Auto-refresh Trend Analysis every 1 minute (60,000 ms)
      this.trendAutoRefreshInterval = setInterval(() => {
        // Only refresh automatically when not in custom historical inspection and when series exist
        if (!this.fullscreenIsCustom && this.activeTrendSeries && this.activeTrendSeries.length > 0) {
          this.fetchTrendData(true);
        }
      }, 60000);
    },
    async pollDashboardCycle() {
      await this.refreshLatestOnly();
      await this.fetchActiveAlarms();
      this.devicePollCycles = (this.devicePollCycles || 0) + 1;
      if (this.devicePollCycles >= 3) {
        this.devicePollCycles = 0;
        await this.syncDeviceCatalog();
      }
    },
    async syncDeviceCatalog() {
      try {
        const res = await axios.get('/api/devices?page_size=100');
        if (res.data?.data) {
          const items = res.data.data.items || res.data.data || [];
          this.devicesList = items;
        }
      } catch (err) {
        // silent ignore
      }
    },
    onDeviceCommEvent(e) {
      const detail = e.detail;
      if (!detail) return;
      const type = detail.type;
      const data = detail.data || {};

      if (type === 'device.communication.success') {
        const devId = data.device_id;
        if (devId) {
          const target = this.devicesList.find(d => d.id === devId);
          if (target) {
            target.connection_status = 'ONLINE';
          }
        }
      } else if (type === 'device.communication.error') {
        const devId = data.device_id;
        if (devId) {
          const target = this.devicesList.find(d => d.id === devId);
          if (target) {
            target.connection_status = 'ERROR';
          }
        }
      } else if (type === 'device.connection.changed') {
        const devId = data.device_id || data.id;
        const status = data.connection_status || data.status;
        if (devId && status) {
          const target = this.devicesList.find(d => d.id === devId);
          if (target) {
            target.connection_status = status;
          }
        }
      }
    },
    onTelemetryEvent(e) {
      const data = e.detail;
      if (!data) return;

      // Only update last valid telemetry timestamp from GOOD data packets
      // This prevents failed comm/error packets from resetting the operator's offline counter!
      if (data.quality === 'GOOD') {
        const goodTime = data.received_at || new Date().toISOString();
        this.lastTelemetryTime = new Date(goodTime);
        if (data.parameter_id) {
          this.$set(this.lastGoodTimeMap, data.parameter_id, goodTime);
          try {
            localStorage.setItem('dl_last_good_time_map', JSON.stringify(this.lastGoodTimeMap));
          } catch (e) {}
        }
        this.lastTelemetrySecondsAgo = 0;
        try {
          localStorage.setItem('dl_last_telemetry_time', new Date(goodTime).toISOString());
        } catch (e) {}
      }

      // Update device connection status when telemetry arrives
      if (data.device_id) {
        const target = this.devicesList.find(d => d.id === data.device_id);
        if (target) {
          if (data.quality === 'GOOD') {
            target.connection_status = 'ONLINE';
          } else if (data.quality === 'BAD' && (data.quality_reason === 'COMMUNICATION_ERROR' || data.quality_reason === 'DEVICE_OFFLINE')) {
            target.connection_status = 'ERROR';
          }
        }
      }

      if (data.parameter_id && data.value !== undefined) {
        this.$set(this.liveMap, data.parameter_id, data);
        this.recordSparkline(data.parameter_id, data.value, data.received_at, data.quality);

        // Flash row/tile briefly
        this.$set(this.liveBlinks, data.parameter_id, true);
        setTimeout(() => {
          this.$set(this.liveBlinks, data.parameter_id, false);
        }, 600);
      }
    },
    async initDashboard() {
      this.isRefreshing = true;
      try {
        await Promise.all([
          this.fetchDevices(),
          this.fetchCategories(),
          this.$store.dispatch('fetchSystemStatus'),
          this.fetchActiveAlarms(),
          this.fetchEvents(),
          this.fetchNotifications(),
        ]);
        if (!this.trendSelectionRestored) {
          this.restoreOrInitTrendSelection();
        }
      } catch (err) {
        console.error('Failed to initialize operator dashboard:', err);
      } finally {
        this.isRefreshing = false;
      }
    },
    async refreshDashboard() {
      this.isRefreshing = true;
      try {
        await Promise.all([
          this.fetchDevices(),
          this.fetchCategories(),
          this.$store.dispatch('fetchSystemStatus'),
          this.fetchActiveAlarms(),
          this.fetchEvents(),
          this.fetchNotifications(),
          this.fetchTrendData(),
        ]);
      } finally {
        this.isRefreshing = false;
      }
    },
    async refreshLatestOnly() {
      for (const dev of this.devicesList) {
        try {
          const res = await axios.get(`/api/devices/${dev.id}/telemetry/latest`);
          if (res.data && res.data.data) {
            for (const item of res.data.data) {
              this.$set(this.liveMap, item.parameter_id, item);
              this.recordSparkline(item.parameter_id, item.value, item.received_at, item.quality);

              // Only update last valid telemetry timestamp from GOOD quality data
              if (item.received_at && item.quality === 'GOOD') {
                this.$set(this.lastGoodTimeMap, item.parameter_id, item.received_at);
                try {
                  localStorage.setItem('dl_last_good_time_map', JSON.stringify(this.lastGoodTimeMap));
                } catch (e) {}
                const itemTime = new Date(item.received_at).getTime();
                if (!this.lastTelemetryTime || itemTime > new Date(this.lastTelemetryTime).getTime()) {
                  this.lastTelemetryTime = new Date(item.received_at);
                  this.lastTelemetrySecondsAgo = Math.max(0, Math.floor((Date.now() - itemTime) / 1000));
                  try {
                    localStorage.setItem('dl_last_telemetry_time', this.lastTelemetryTime.toISOString());
                  } catch (e) {}
                }
              }
              // Parameter values and sparklines are updated from latest cache.
              // Device connection status is maintained strictly by backend comm events and catalog sync.
            }
          }
        } catch (e) {
          // ignore error
        }
      }
    },
    async fetchDevices() {
      try {
        const res = await axios.get('/api/devices?page_size=100');
        if (res.data && res.data.data) {
          this.devicesList = res.data.data.items || res.data.data || [];
          await this.refreshLatestOnly();
          await this.fetchInitialSparklines();

          // Pre-populate lastGoodTimeMap from DB parameter records
          for (const dev of this.devicesList) {
            for (const p of dev.parameters || []) {
              let ts = p.current_received_at || dev.last_data_at || p.last_updated;
              // If param is bad/offline and this.lastTelemetryTime is known, ensure ts does not exceed lastTelemetryTime
              if (this.lastTelemetryTime && (p.current_quality === 'BAD' || dev.connection_status === 'OFFLINE' || dev.connection_status === 'ERROR')) {
                const teleMs = new Date(this.lastTelemetryTime).getTime();
                if (!ts || new Date(ts).getTime() > teleMs) {
                  ts = this.lastTelemetryTime;
                }
              }
              const existingStored = this.lastGoodTimeMap[p.id];
              // If existing stored time in localStorage was set to a newer error poll timestamp, correct it!
              if (this.lastTelemetryTime && existingStored && new Date(existingStored).getTime() > new Date(this.lastTelemetryTime).getTime() && (p.current_quality === 'BAD' || dev.connection_status === 'OFFLINE')) {
                this.$set(this.lastGoodTimeMap, p.id, this.lastTelemetryTime);
              } else if (ts && !this.lastGoodTimeMap[p.id]) {
                this.$set(this.lastGoodTimeMap, p.id, ts);
              }
            }
          }
          try {
            localStorage.setItem('dl_last_good_time_map', JSON.stringify(this.lastGoodTimeMap));
          } catch (e) {}

          // Initialize/Reconcile lastTelemetryTime from most recent valid (or historical) telemetry
          let maxGood = null;
          let maxAny = null;
          for (const dev of this.devicesList) {
            for (const p of dev.parameters || []) {
              const live = this.liveMap[p.id];
              const ts = live?.received_at || p.current_received_at;
              if (ts) {
                const t = new Date(ts).getTime();
                if (live?.quality === 'GOOD' || p.current_quality === 'GOOD') {
                  if (!maxGood || t > maxGood) maxGood = t;
                }
                if (!maxAny || t > maxAny) maxAny = t;
              }
            }
            if (dev.last_data_at) {
              const t = new Date(dev.last_data_at).getTime();
              if (!maxAny || t > maxAny) maxAny = t;
            }
          }

          // If we have a good telemetry timestamp from DB/live cache, use it if newer or not yet set
          if (maxGood) {
            if (!this.lastTelemetryTime || maxGood > new Date(this.lastTelemetryTime).getTime()) {
              this.lastTelemetryTime = new Date(maxGood);
              this.lastTelemetrySecondsAgo = Math.max(0, Math.floor((Date.now() - maxGood) / 1000));
              try {
                localStorage.setItem('dl_last_telemetry_time', new Date(maxGood).toISOString());
              } catch (e) {}
            }
          } else if (!this.lastTelemetryTime && maxAny) {
            // Only fall back to maxAny if we have NO previous stored or active lastTelemetryTime at all
            this.lastTelemetryTime = new Date(maxAny);
            this.lastTelemetrySecondsAgo = Math.max(0, Math.floor((Date.now() - maxAny) / 1000));
            try {
              localStorage.setItem('dl_last_telemetry_time', new Date(maxAny).toISOString());
            } catch (e) {}
          }
        }
      } catch (err) {
        console.error('Failed to load device catalog:', err);
      }
    },
    async fetchActiveAlarms() {
      await this.$store.dispatch('fetchActiveAlarms');
    },
    async fetchEvents() {
      try {
        const res = await axios.get('/api/audit-trails?page=1&page_size=30');
        if (res.data && res.data.data) {
          this.recentEvents = res.data.data.items || res.data.data || [];
        }
      } catch (err) {
        console.error('Failed to load audit events:', err);
      }
    },
    async fetchNotifications() {
      try {
        const res = await axios.get('/api/notifications?page=1&page_size=30');
        if (res.data && res.data.data) {
          this.notificationQueue = res.data.data.items || res.data.data || [];
          this.notificationQueueStats.total = this.notificationQueue.length;
          this.notificationQueueStats.pending = this.notificationQueue.filter(n => n.status === 'PENDING' || n.status === 'RETRYING').length;
        }
      } catch (err) {
        console.error('Failed to load notification queue:', err);
      }
    },
    togglePin(paramId) {
      const idx = this.pinnedIds.indexOf(paramId);
      if (idx !== -1) {
        this.pinnedIds.splice(idx, 1);
      } else {
        this.pinnedIds.push(paramId);
      }
      localStorage.setItem('dl_pinned_params', JSON.stringify(this.pinnedIds));
    },
    isPinned(paramId) {
      return this.pinnedIds.includes(paramId);
    },
    toggleFilterPinned() {
      this.filterPinnedOnly = !this.filterPinnedOnly;
    },
    onSearchDebounced() {
      clearTimeout(this.searchDebounceTimer);
      this.searchDebounceTimer = setTimeout(() => {
        // search query updated reactively in computed
      }, 250);
    },
    toggleAudioMute() {
      audioAlert.unlockAudio().catch(() => {});
      const nextMuted = !this.isAudioMuted;
      audioAlert.setMute(nextMuted);
      const current = getStoredAlarmPreferences(this.currentUserIdentifier);
      const updated = {
        ...current,
        isMuted: nextMuted,
        muteUntil: null,
        muteDurationMinutes: nextMuted ? -1 : 0,
      };
      saveAlarmPreferences(updated, this.currentUserIdentifier);
      this.$store.commit('SET_ALARM_PREFERENCES', updated);
      this.audioMuteDuration = updated.muteDurationMinutes;
      if (!nextMuted) {
        audioAlert.previewSound('WARNING');
      }
    },
    initAudioPreferences() {
      const prefs = getStoredAlarmPreferences(this.currentUserIdentifier);
      applyAlarmPreferences(prefs);
      this.$store.commit('SET_ALARM_PREFERENCES', prefs);
      this.audioVolume = prefs.volume !== undefined ? prefs.volume : 0.7;
      this.audioRepeatInterval = prefs.repeatIntervalSeconds !== undefined ? prefs.repeatIntervalSeconds : 5;
      this.audioSoundEnabled = prefs.soundEnabled !== undefined ? prefs.soundEnabled : true;
      this.audioMuteDuration = prefs.muteDurationMinutes !== undefined ? prefs.muteDurationMinutes : 0;
    },
    toggleAudioSetupPopover() {
      this.showAudioSetupPopover = !this.showAudioSetupPopover;
      if (this.showAudioSetupPopover) {
        audioAlert.unlockAudio().catch(() => {});
        const current = getStoredAlarmPreferences(this.currentUserIdentifier);
        this.audioVolume = current.volume !== undefined ? current.volume : this.audioVolume;
        this.audioRepeatInterval = current.repeatIntervalSeconds !== undefined ? current.repeatIntervalSeconds : this.audioRepeatInterval;
        this.audioMuteDuration = current.muteDurationMinutes !== undefined ? current.muteDurationMinutes : this.audioMuteDuration;
      }
    },
    openFullAlarmSettingsModal() {
      this.showAudioSetupPopover = false;
      this.showAlarmNotificationModal = true;
    },
    onAudioVolumeChange() {
      audioAlert.unlockAudio().catch(() => {});
      audioAlert.setVolume(this.audioVolume);
      const current = getStoredAlarmPreferences(this.currentUserIdentifier);
      const updated = { ...current, volume: this.audioVolume };
      saveAlarmPreferences(updated, this.currentUserIdentifier);
      this.$store.commit('SET_ALARM_PREFERENCES', updated);
    },
    onAudioRepeatIntervalChange() {
      audioAlert.setRepeatInterval(this.audioRepeatInterval);
      const current = getStoredAlarmPreferences(this.currentUserIdentifier);
      const updated = { ...current, repeatIntervalSeconds: this.audioRepeatInterval };
      saveAlarmPreferences(updated, this.currentUserIdentifier);
      this.$store.commit('SET_ALARM_PREFERENCES', updated);
    },
    onAudioMuteDurationChange() {
      audioAlert.unlockAudio().catch(() => {});
      const dur = Number(this.audioMuteDuration);
      const nextMuted = dur !== 0;
      const durMins = dur > 0 ? dur : null;
      audioAlert.setMute(nextMuted, durMins);
      const current = getStoredAlarmPreferences(this.currentUserIdentifier);
      const updated = {
        ...current,
        isMuted: nextMuted,
        muteDurationMinutes: dur,
        muteUntil: dur > 0 ? (Date.now() + dur * 60 * 1000) : null,
      };
      saveAlarmPreferences(updated, this.currentUserIdentifier);
      this.$store.commit('SET_ALARM_PREFERENCES', updated);
      if (!nextMuted) {
        audioAlert.previewSound('WARNING');
      }
    },
    testAlarmChime(severity) {
      audioAlert.unlockAudio().then(() => {
        audioAlert.previewSound(severity || 'WARNING');
      });
    },

    // ----------------------------------------------------
    // TREND ANALYSIS TIME-SERIES MAPPING & RENDERING
    // ----------------------------------------------------
    selectParamForTrend(param) {
      if (!param) return;
      const existingIdx = this.activeTrendSeries.findIndex(s => s.parameterId === param.param_id && (!s.deviceId || !param.device_id || s.deviceId === param.device_id));
      if (existingIdx !== -1) {
        this.activeTrendSeries.splice(existingIdx, 1);
        if (this.trendSelectionRestored) {
          this.saveTrendSelection();
        }
      } else {
        if (this.activeTrendSeries.length >= this.maxTrendSeries) {
          this.activeTrendSeries.shift();
        }
        // Deterministic SCADA color assignment based on parameter type & current theme
        const color = getParamSeriesColor(param, this.isDark, this.activeTrendSeries.length);

        this.activeTrendSeries.push({
          id: `${param.device_id}_${param.param_id}`,
          deviceId: param.device_id,
          parameterId: param.param_id,
          deviceCode: param.device_code,
          paramCode: param.param_code,
          paramName: param.param_name || param.param_code,
          unit: param.unit || '',
          precision: param.precision !== undefined ? param.precision : 2,
          latestVal: this.formatParamValue(param),
          color,
          dataPoints: [],
          stats: { min: '--', max: '--', avg: '--', count: 0 },
        });
        if (this.trendSelectionRestored) {
          this.saveTrendSelection();
        }
        this.fetchTrendData();
      }
    },
    isParamSelectedForTrend(param) {
      if (!param) return false;
      return this.activeTrendSeries.some(s => s.parameterId === param.param_id && (!s.deviceId || !param.device_id || s.deviceId === param.device_id));
    },
    removeTrendSeries(id) {
      this.activeTrendSeries = this.activeTrendSeries.filter(s => s.id !== id);
      if (this.trendSelectionRestored) {
        this.saveTrendSelection();
      }
    },
    clearAllTrendSeries() {
      this.activeTrendSeries = [];
      if (this.trendSelectionRestored) {
        this.saveTrendSelection();
      }
    },
    selectAllEligibleTrendSeries(shouldFetch = true) {
      if (this.allParameters.length === 0) return;
      const eligible = this.allParameters.slice(0, this.maxTrendSeries);
      const newSeries = [];
      eligible.forEach((p, idx) => {
        const color = getParamSeriesColor(p, this.isDark, idx);
        newSeries.push({
          id: `${p.device_id}_${p.param_id}`,
          deviceId: p.device_id,
          parameterId: p.param_id,
          deviceCode: p.device_code,
          paramCode: p.param_code,
          paramName: p.param_name || p.param_code,
          unit: p.unit || '',
          precision: p.precision !== undefined ? p.precision : 2,
          latestVal: this.formatParamValue(p),
          color,
          dataPoints: [],
          stats: { min: '--', max: '--', avg: '--', count: 0 },
        });
      });
      this.activeTrendSeries = newSeries;
      if (this.trendSelectionRestored) {
        this.saveTrendSelection();
      }
      if (shouldFetch) {
        this.fetchTrendData();
      }
    },
    resetTrendSelectionToDefault() {
      try {
        localStorage.removeItem(this.trendStorageKey);
      } catch (e) {}
      this.selectAllEligibleTrendSeries(true);
    },
    restoreOrInitTrendSelection() {
      if (this.allParameters.length === 0) return;
      let saved = null;
      try {
        const raw = localStorage.getItem(this.trendStorageKey);
        if (raw) {
          saved = JSON.parse(raw);
        }
      } catch (e) {
        console.warn('Failed reading trend selection from localStorage:', e);
      }

      // 1. Returning user with saved series preferences
      if (saved && Array.isArray(saved.series)) {
        const restored = [];
        for (const item of saved.series) {
          // Identify stably by device ID + parameter ID/code
          const match = this.allParameters.find(p =>
            (Number(p.device_id) === Number(item.deviceId) && Number(p.param_id) === Number(item.parameterId)) ||
            (p.device_code === item.deviceCode && p.param_code === item.paramCode)
          );
          // If valid and not disabled/deleted, restore safely
          if (match && !restored.some(r => r.parameterId === match.param_id)) {
            const color = getParamSeriesColor(match, this.isDark, restored.length);
            restored.push({
              id: `${match.device_id}_${match.param_id}`,
              deviceId: match.device_id,
              parameterId: match.param_id,
              deviceCode: match.device_code,
              paramCode: match.param_code,
              paramName: match.param_name || match.param_code,
              unit: match.unit || '',
              precision: match.precision !== undefined ? match.precision : 2,
              latestVal: this.formatParamValue(match),
              color,
              dataPoints: [],
              stats: { min: '--', max: '--', avg: '--', count: 0 },
            });
            if (restored.length >= this.maxTrendSeries) break;
          }
        }

        // If saved parameters are valid, keep restored selection
        if (restored.length > 0) {
          this.activeTrendSeries = restored;
          this.activeTrendSeries.forEach((s, idx) => {
            const p = (this.allParameters || []).find(item => item.param_id === s.parameterId) || s;
            s.color = getParamSeriesColor(p, this.isDark, idx);
            if (!s.paramName || s.paramName === '-' || !s.paramCode || s.paramCode === '-') {
              if (p) {
                s.paramName = p.param_name || p.param_code;
                s.paramCode = p.param_code;
              }
            }
          });
          this.trendSelectionRestored = true;
          this.saveTrendSelection();
          this.fetchTrendData();
          return;
        }
      }

      // 2. Fresh installation / no saved selection or all saved parameters deleted/invalid:
      // Automatically select all eligible configured parameters
      this.selectAllEligibleTrendSeries(false);
      this.trendSelectionRestored = true;
      this.saveTrendSelection();
      this.fetchTrendData();
    },
    saveTrendSelection() {
      if (!this.trendSelectionRestored) return;
      const payload = {
        version: 1,
        userId: this.currentUserIdentifier,
        updatedAt: new Date().toISOString(),
        series: this.activeTrendSeries.map(s => ({
          deviceId: s.deviceId,
          parameterId: s.parameterId,
          deviceCode: s.deviceCode,
          paramCode: s.paramCode,
        })),
      };
      try {
        localStorage.setItem(this.trendStorageKey, JSON.stringify(payload));
      } catch (e) {
        console.warn('Failed to save trend selection:', e);
      }
    },
    addDefaultTrendSeries() {
      this.restoreOrInitTrendSelection();
    },
    getSeriesScale(series, pts = null) {
      if (this.trendScaleMode === 'normalized') {
        return { min: 0, max: 100, span: 100 };
      }
      const sUnit = (series.unit || '').trim();
      const leftUnit = (this.trendScales.left?.unit || '').trim();
      const rightUnit = (this.trendScales.right?.unit || '').trim();

      let minY = 0;
      let maxY = 100;

      if (this.trendScaleMode === 'auto') {
        if (sUnit === leftUnit || (!this.trendScales.isDual && !rightUnit)) {
          minY = this.trendScales.left ? this.trendScales.left.min : 0;
          maxY = this.trendScales.left ? this.trendScales.left.max : 100;
        } else if (this.trendScales.isDual && sUnit === rightUnit) {
          minY = this.trendScales.right ? this.trendScales.right.min : 0;
          maxY = this.trendScales.right ? this.trendScales.right.max : 100;
        } else {
          // Incompatible 3rd unit: calculate dedicated bounds from series points to prevent cross-unit distortion
          const points = pts || series.dataPoints || [];
          const yVals = points.map(p => p.y).filter(v => v !== null && isFinite(v));
          if (yVals.length > 0) {
            minY = Math.min(...yVals);
            maxY = Math.max(...yVals);
          } else {
            minY = 0;
            maxY = 100;
          }
        }
      } else {
        const points = pts || series.dataPoints || [];
        const yVals = points.map(p => p.y).filter(v => v !== null && isFinite(v));
        minY = yVals.length > 0 ? Math.min(...yVals) : 0;
        maxY = yVals.length > 0 ? Math.max(...yVals) : 100;
      }

      if (minY === maxY) {
        minY -= 1;
        maxY += 1;
      }
      return { min: minY, max: maxY, span: maxY - minY || 1 };
    },
    setTimeframe(tf) {
      this.currentTimeframe = tf;
      this.fullscreenIsCustom = false;
      this.fetchTrendData();
    },
    setCustomTimeframe() {
      if (!this.fullscreenCustomStart || !this.fullscreenCustomEnd) return;
      this.fullscreenIsCustom = true;
      this.fetchTrendData();
    },
    setChartMode(mode) {
      this.chartMode = mode;
      this.fetchTrendData();
    },
    toggleTrendScaleMode() {
      this.trendScaleMode = this.trendScaleMode === 'auto' ? 'normalized' : 'auto';
    },
    toggleTrendSpikeZoom() {
      this.trendSpikeZoom = this.trendSpikeZoom === 'full' ? 'robust' : 'full';
    },
    getSeriesStrokeWidth(s, sIdx) {
      if (this.hoveredTrendSeriesId === s.id) return 2.2;
      return sIdx === 0 ? 1.8 : 1.6;
    },
    openTrendFullscreen() {
      const now = new Date();
      const start = new Date(this.trendTimeBounds.start || (now.getTime() - 15 * 60 * 1000));
      this.fullscreenCustomStart = this.toLocalISO(start);
      this.fullscreenCustomEnd = this.toLocalISO(now);
      this.showTrendFullscreen = true;
      this.$nextTick(() => {
        this.initFullscreenResizeObserver();
      });
    },
    closeTrendFullscreen() {
      this.showTrendFullscreen = false;
      this.showTrendParamPicker = false;
      if (this.fullscreenResizeObserver) {
        this.fullscreenResizeObserver.disconnect();
        this.fullscreenResizeObserver = null;
      }
    },
    formatAxisValue(val) {
      if (val === null || val === undefined || isNaN(val) || !isFinite(val)) return '--';
      const abs = Math.abs(val);
      if (abs >= 10000) {
        return (val / 1000).toFixed(1) + 'k';
      } else if (abs >= 100) {
        return val.toFixed(0);
      } else if (abs >= 10) {
        return val.toFixed(1);
      } else if (abs >= 1) {
        return val.toFixed(2);
      } else if (abs === 0) {
        return '0';
      } else {
        return val.toFixed(3);
      }
    },
    async fetchTrendData(silent = false) {
      if (this.activeTrendSeries.length === 0) return;
      if (!silent) {
        this.isChartLoading = true;
      }
      try {
        const now = new Date();
        let startTime = new Date(now.getTime() - 24 * 3600 * 1000);
        let endTime = now;

        if (this.fullscreenIsCustom && this.fullscreenCustomStart && this.fullscreenCustomEnd) {
          startTime = new Date(this.fullscreenCustomStart);
          endTime = new Date(this.fullscreenCustomEnd);
        } else {
          if (this.currentTimeframe === '1m') startTime = new Date(now.getTime() - 60 * 1000);
          else if (this.currentTimeframe === '5m') startTime = new Date(now.getTime() - 5 * 60 * 1000);
          else if (this.currentTimeframe === '15m') startTime = new Date(now.getTime() - 15 * 60 * 1000);
          else if (this.currentTimeframe === '1h') startTime = new Date(now.getTime() - 3600 * 1000);
          else if (this.currentTimeframe === '6h') startTime = new Date(now.getTime() - 6 * 3600 * 1000);
        }

        this.trendTimeBounds = {
          start: startTime.getTime(),
          end: endTime.getTime(),
        };

        await Promise.all(this.activeTrendSeries.map(async (s) => {
          let points = [];
          // 1. Primary: Intelligent Downsampling Engine (guarantees coverage across entire timeframe)
          try {
            const downRes = await axios.get(`/api/devices/${s.deviceId}/telemetry/downsampled`, {
              params: {
                parameter_id: s.parameterId,
                start_time: startTime.toISOString(),
                end_time: endTime.toISOString(),
                resolution: 'auto',
                mode: this.chartMode,
              },
            });
            if (downRes.data && downRes.data.data && Array.isArray(downRes.data.data.points) && downRes.data.data.points.length > 0) {
              points = downRes.data.data.points.map(p => {
                const rawVal = p.value !== null && p.value !== undefined ? p.value : p.avg_value;
                const numVal = rawVal !== null && rawVal !== undefined && rawVal !== '' ? Number(rawVal) : null;
                // Reject invalid numbers, NaN, or raw epoch timestamps (e.g. > 1e8) mistakenly in value
                const isValid = numVal !== null && isFinite(numVal) && Math.abs(numVal) < 1e8;
                return {
                  x: new Date(p.timestamp).getTime(),
                  y: isValid ? numVal : null,
                  min: p.min_value !== undefined && isFinite(p.min_value) ? Number(p.min_value) : null,
                  max: p.max_value !== undefined && isFinite(p.max_value) ? Number(p.max_value) : null,
                  avg: p.avg_value !== undefined && isFinite(p.avg_value) ? Number(p.avg_value) : null,
                  quality: p.quality || 'GOOD',
                };
              });
            }
          } catch (err) {
            // silent fallback
          }

          // 2. Secondary fallback: Historical / Raw endpoint
          if (points.length === 0) {
            try {
              const endpoint = this.chartMode === 'raw'
                ? `/api/devices/${s.deviceId}/telemetry/raw`
                : `/api/devices/${s.deviceId}/telemetry/history`;

              const res = await axios.get(endpoint, {
                params: {
                  parameter_id: s.parameterId,
                  start_time: startTime.toISOString(),
                  end_time: endTime.toISOString(),
                  page_size: 500,
                },
              }).catch(() => null);

              if (res && res.data && res.data.data) {
                const items = res.data.data.items || [];
                points = items.map(item => {
                  const raw = this.chartMode === 'raw'
                    ? (item.raw_value !== undefined ? item.raw_value : item.value)
                    : (item.processed_value !== undefined ? item.processed_value : item.value);
                  const num = raw !== null && raw !== undefined && raw !== '' ? Number(raw) : null;
                  const isValid = num !== null && isFinite(num) && Math.abs(num) < 1e8;
                  return {
                    x: new Date(item.received_at || item.timestamp).getTime(),
                    y: isValid ? num : null,
                    quality: item.quality || 'GOOD',
                  };
                });
              }
            } catch (err) {}
          }

          // Filter out bad timestamps and sort chronologically
          const validPoints = points
            .filter(p => p.x && !isNaN(p.x))
            .sort((a, b) => a.x - b.x);

          // For all filters from 1m to 6h, enforce 5-second polling interval alignment matching hardware setup
          const timeSpan = endTime.getTime() - startTime.getTime();
          const isUnderOrEqual6h = timeSpan <= (6 * 3600 * 1000 + 1000);
          const pollStepMs = isUnderOrEqual6h ? 5000 : null;

          let processed = [];
          if (pollStepMs) {
            // Group and align points strictly into 5-second polling buckets (:00, :05, :10, :15...)
            const bucketMap = new Map();
            for (const pt of validPoints) {
              const alignedX = Math.round(pt.x / pollStepMs) * pollStepMs;
              bucketMap.set(alignedX, {
                ...pt,
                x: alignedX,
              });
            }
            processed = Array.from(bucketMap.values()).sort((a, b) => a.x - b.x);
          } else {
            // Deduplicate consecutive points with identical timestamps
            for (let i = 0; i < validPoints.length; i++) {
              if (i === 0 || validPoints[i].x !== validPoints[i - 1].x) {
                processed.push(validPoints[i]);
              } else {
                processed[processed.length - 1] = validPoints[i];
              }
            }
          }
          s.dataPoints = processed;

          // Calculate statistics over window using only valid numeric values
          const yVals = s.dataPoints.map(p => p.y).filter(v => v !== null && typeof v === 'number' && isFinite(v));
          if (yVals.length > 0) {
            const min = Math.min(...yVals);
            const max = Math.max(...yVals);
            const sum = yVals.reduce((acc, v) => acc + v, 0);
            const avg = sum / yVals.length;
            const prec = s.precision !== undefined ? s.precision : (Math.abs(max) >= 100 ? 1 : 2);
            s.stats = {
              min: min.toFixed(prec),
              max: max.toFixed(prec),
              avg: avg.toFixed(prec),
              count: yVals.length,
            };
            const lastVal = yVals[yVals.length - 1];
            s.latestVal = lastVal.toFixed(prec);
          } else {
            s.stats = { min: '--', max: '--', avg: '--', count: 0 };
            s.latestVal = '--';
          }
        }));
      } finally {
        if (!silent) {
          this.isChartLoading = false;
        }
        // Reschedule next auto-refresh 1 minute after any fetch completes
        this.setupTrendAutoRefresh();
      }
    },
    generateChartPath(series, bounds = null, width = null, height = null, padLeft = null, padRight = null, padTop = null, padBottom = null) {
      const pts = series.dataPoints;
      if (!pts || pts.length === 0) return '';

      const dims = this.trendChartDims;
      const w = width !== null && width !== undefined ? width : dims.width;
      const h = height !== null && height !== undefined ? height : dims.height;
      const pLeft = padLeft !== null && padLeft !== undefined ? padLeft : dims.padLeft;
      const pRight = padRight !== null && padRight !== undefined ? padRight : dims.padRight;
      const pTop = padTop !== null && padTop !== undefined ? padTop : dims.padTop;
      const pBottom = padBottom !== null && padBottom !== undefined ? padBottom : dims.padBottom;

      const plotW = w - pLeft - pRight;
      const plotH = h - pTop - pBottom;

      const timeStart = bounds ? bounds.start : (this.trendTimeBounds.start || pts[0].x);
      const timeEnd = bounds ? bounds.end : (this.trendTimeBounds.end || pts[pts.length - 1].x);
      const timeSpan = Math.max(timeEnd - timeStart, 1000);

      // Determine scale for this series (supports dual axis and independent scale for 3rd unit)
      const scaleInfo = this.getSeriesScale(series, pts);
      const minY = scaleInfo.min;
      const maxY = scaleInfo.max;
      const spanY = scaleInfo.span;

      // Estimate gap threshold: when delta t exceeds 3-4x typical sample interval, break path
      let maxGapMs = 15 * 60 * 1000;
      if (timeSpan <= 90 * 1000) maxGapMs = 15 * 1000; // 15s for 1m
      else if (timeSpan <= 7 * 60 * 1000) maxGapMs = 30 * 1000; // 30s for 5m
      else if (timeSpan <= 20 * 60 * 1000) maxGapMs = 60 * 1000; // 1 min for 15m
      else if (timeSpan <= 2 * 3600 * 1000) maxGapMs = 5 * 60 * 1000; // 5 min for 1h
      else if (timeSpan <= 8 * 3600 * 1000) maxGapMs = 20 * 60 * 1000; // 20 min for 6h
      else if (timeSpan <= 25 * 3600 * 1000) maxGapMs = 60 * 60 * 1000; // 1 hr for 24h
      else maxGapMs = 4 * 3600 * 1000; // 4 hrs for multi-day

      let pathStr = '';
      let inSegment = false;
      let prevX = null;

      for (let i = 0; i < pts.length; i++) {
        const p = pts[i];
        if (p.y === null || !isFinite(p.y)) {
          inSegment = false;
          continue;
        }

        const px = pLeft + Math.min(Math.max((p.x - timeStart) / timeSpan, 0), 1) * plotW;
        const py = (pTop + plotH) - Math.min(Math.max((p.y - minY) / spanY, 0), 1) * plotH;

        if (prevX !== null && (p.x - prevX) > maxGapMs) {
          inSegment = false;
        }

        if (!inSegment) {
          pathStr += ` M ${px.toFixed(1)} ${py.toFixed(1)}`;
          inSegment = true;
        } else {
          pathStr += ` L ${px.toFixed(1)} ${py.toFixed(1)}`;
        }
        prevX = p.x;
      }

      return pathStr.trim();
    },
    generatePrimaryAreaPath(bounds = null, width = null, height = null, padLeft = null, padRight = null, padTop = null, padBottom = null) {
      if (this.activeTrendSeries.length === 0) return '';
      const primary = this.activeTrendSeries[0];
      const pts = primary.dataPoints || [];
      if (pts.length === 0) return '';

      const dims = this.trendChartDims;
      const w = width !== null && width !== undefined ? width : dims.width;
      const h = height !== null && height !== undefined ? height : dims.height;
      const pLeft = padLeft !== null && padLeft !== undefined ? padLeft : dims.padLeft;
      const pRight = padRight !== null && padRight !== undefined ? padRight : dims.padRight;
      const pTop = padTop !== null && padTop !== undefined ? padTop : dims.padTop;
      const pBottom = padBottom !== null && padBottom !== undefined ? padBottom : dims.padBottom;

      const plotW = w - pLeft - pRight;
      const plotH = h - pTop - pBottom;
      const baseY = pTop + plotH;

      const timeStart = bounds ? bounds.start : (this.trendTimeBounds.start || pts[0].x);
      const timeEnd = bounds ? bounds.end : (this.trendTimeBounds.end || pts[pts.length - 1].x);
      const timeSpan = Math.max(timeEnd - timeStart, 1000);

      let minY = this.trendScales.left.min;
      let maxY = this.trendScales.left.max;
      if (minY === maxY) {
        minY -= 1;
        maxY += 1;
      }
      const spanY = maxY - minY || 1;

      let maxGapMs = 15 * 60 * 1000;
      if (timeSpan <= 90 * 1000) maxGapMs = 15 * 1000;
      else if (timeSpan <= 7 * 60 * 1000) maxGapMs = 30 * 1000;
      else if (timeSpan <= 20 * 60 * 1000) maxGapMs = 60 * 1000;
      else if (timeSpan <= 2 * 3600 * 1000) maxGapMs = 5 * 60 * 1000;
      else if (timeSpan <= 8 * 3600 * 1000) maxGapMs = 20 * 60 * 1000;
      else if (timeSpan <= 25 * 3600 * 1000) maxGapMs = 60 * 60 * 1000;
      else maxGapMs = 4 * 3600 * 1000;

      // Group points into continuous valid segments to avoid diagonal area fills across outages
      const segments = [];
      let currentSeg = [];
      let prevX = null;

      for (let i = 0; i < pts.length; i++) {
        const p = pts[i];
        if (p.y === null || !isFinite(p.y)) {
          if (currentSeg.length > 0) {
            segments.push(currentSeg);
            currentSeg = [];
          }
          prevX = null;
          continue;
        }

        if (prevX !== null && (p.x - prevX) > maxGapMs) {
          if (currentSeg.length > 0) {
            segments.push(currentSeg);
            currentSeg = [];
          }
        }

        currentSeg.push(p);
        prevX = p.x;
      }
      if (currentSeg.length > 0) {
        segments.push(currentSeg);
      }

      if (segments.length === 0) return '';

      let d = '';
      for (const seg of segments) {
        if (seg.length === 0) continue;
        const firstPx = pLeft + Math.min(Math.max((seg[0].x - timeStart) / timeSpan, 0), 1) * plotW;
        const lastPx = pLeft + Math.min(Math.max((seg[seg.length - 1].x - timeStart) / timeSpan, 0), 1) * plotW;

        let segD = `M ${firstPx.toFixed(1)} ${baseY}`;
        for (let i = 0; i < seg.length; i++) {
          const px = pLeft + Math.min(Math.max((seg[i].x - timeStart) / timeSpan, 0), 1) * plotW;
          const py = (pTop + plotH) - Math.min(Math.max((seg[i].y - minY) / spanY, 0), 1) * plotH;
          segD += ` L ${px.toFixed(1)} ${py.toFixed(1)}`;
        }
        segD += ` L ${lastPx.toFixed(1)} ${baseY} Z`;
        d += (d ? ' ' : '') + segD;
      }

      return d;
    },
    generateChartPolyline(series, bounds = null, width = 560, height = 200, padLeft = 52, padRight = 48, padTop = 24, padBottom = 30) {
      const pts = series.dataPoints;
      if (!pts || pts.length === 0) return '';

      const plotW = width - padLeft - padRight;
      const plotH = height - padTop - padBottom;

      const timeStart = bounds ? bounds.start : (this.trendTimeBounds.start || pts[0].x);
      const timeEnd = bounds ? bounds.end : (this.trendTimeBounds.end || pts[pts.length - 1].x);
      const timeSpan = Math.max(timeEnd - timeStart, 1000);

      const scaleInfo = this.getSeriesScale(series, pts);
      const minY = scaleInfo.min;
      const maxY = scaleInfo.max;
      const spanY = scaleInfo.span;

      return pts
        .filter(p => p.y !== null && isFinite(p.y))
        .map(p => {
          const px = padLeft + Math.min(Math.max((p.x - timeStart) / timeSpan, 0), 1) * plotW;
          const py = (padTop + plotH) - Math.min(Math.max((p.y - minY) / spanY, 0), 1) * plotH;
          return `${px.toFixed(1)},${py.toFixed(1)}`;
        }).join(' ');
    },
    generateChartPolygon(series, bounds = null, width = 560, height = 200, padLeft = 52, padRight = 48, padTop = 24, padBottom = 30) {
      const polyline = this.generateChartPolyline(series, bounds, width, height, padLeft, padRight, padTop, padBottom);
      if (!polyline) return '';
      const pts = (series.dataPoints || []).filter(p => p.y !== null && isFinite(p.y));
      if (pts.length === 0) return '';
      const timeStart = bounds ? bounds.start : (this.trendTimeBounds.start || pts[0].x);
      const timeEnd = bounds ? bounds.end : (this.trendTimeBounds.end || pts[pts.length - 1].x);
      const timeSpan = Math.max(timeEnd - timeStart, 1000);
      const plotW = width - padLeft - padRight;
      const baseY = padTop + (height - padTop - padBottom);

      const firstPt = pts[0];
      const lastPt = pts[pts.length - 1];
      const firstX = padLeft + Math.min(Math.max((firstPt.x - timeStart) / timeSpan, 0), 1) * plotW;
      const lastX = padLeft + Math.min(Math.max((lastPt.x - timeStart) / timeSpan, 0), 1) * plotW;

      return `${polyline} ${lastX.toFixed(1)},${baseY} ${firstX.toFixed(1)},${baseY}`;
    },
    onChartScrub(e) {
      if (!this.$refs.scadaChartContainer || this.activeTrendSeries.length === 0) return;
      const rect = this.$refs.scadaChartContainer.getBoundingClientRect();
      const mouseX = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
      const { padLeft, plotW, padTop, plotH } = this.trendChartDims;
      const clampedX = Math.min(Math.max(mouseX, padLeft), padLeft + plotW);
      const frac = (clampedX - padLeft) / plotW;

      const start = this.trendTimeBounds.start;
      const end = this.trendTimeBounds.end;
      const timeSpan = Math.max(end - start, 1000);
      const isUnderOrEqual6h = timeSpan <= (6 * 3600 * 1000 + 1000);

      const hoverTime = start + frac * timeSpan;
      let effectiveTime = isUnderOrEqual6h ? (Math.round(hoverTime / 5000) * 5000) : hoverTime;
      if (isUnderOrEqual6h) {
        effectiveTime = Math.max(Math.floor(start / 5000) * 5000, Math.min(Math.ceil(end / 5000) * 5000, effectiveTime));
      }
      const snappedFrac = Math.min(Math.max((effectiveTime - start) / timeSpan, 0), 1);
      const scrubX = isUnderOrEqual6h ? (padLeft + snappedFrac * plotW) : clampedX;

      const seriesValues = [];
      for (const s of this.activeTrendSeries) {
        if (!s.dataPoints || s.dataPoints.length === 0) continue;
        let closest = s.dataPoints[0];
        let minDiff = Math.abs(closest.x - effectiveTime);
        for (let i = 1; i < s.dataPoints.length; i++) {
          const diff = Math.abs(s.dataPoints[i].x - effectiveTime);
          if (diff < minDiff) {
            minDiff = diff;
            closest = s.dataPoints[i];
          }
        }

        const scaleInfo = this.getSeriesScale(s, s.dataPoints);
        let plotY = null;
        if (closest.y !== null && isFinite(closest.y)) {
          const fracY = Math.min(Math.max((closest.y - scaleInfo.min) / scaleInfo.span, 0), 1);
          plotY = (padTop + plotH) - fracY * plotH;
        }

        const param = this.allParameters.find(p => p.param_id === s.parameterId) || {};
        let alarmStatus = 'NORMAL';
        if (closest.y !== null && typeof closest.y === 'number') {
          alarmStatus = this.determineAlarmState(param, closest.y);
        }

        seriesValues.push({
          seriesId: s.id,
          deviceCode: s.deviceCode,
          paramCode: s.paramCode,
          paramName: s.paramName || param.param_name || s.paramCode,
          val: closest.y !== null && closest.y !== undefined ? this.formatAxisValue(closest.y) : '--',
          unit: s.unit || '',
          color: s.color,
          plotY,
          quality: closest.quality || 'GOOD',
          alarmStatus,
        });
      }

      const firstSeries = seriesValues[0] || { deviceCode: 'AQMS', paramCode: 'val', paramName: 'Param', val: '--', unit: '' };
      const d = new Date(effectiveTime);
      this.scrubPoint = {
        x: scrubX,
        rawPixelX: mouseX,
        y: 85,
        time: d.toLocaleTimeString('en-GB'),
        date: d.toLocaleDateString([], { day: '2-digit', month: 'short', year: 'numeric' }),
        label: `${firstSeries.deviceCode} · ${firstSeries.paramName}`,
        val: firstSeries.val,
        unit: firstSeries.unit,
        allSeries: seriesValues,
      };
    },
    onFullscreenChartScrub(e) {
      if (!this.$refs.fullscreenChartContainer || this.activeTrendSeries.length === 0) return;
      const rect = this.$refs.fullscreenChartContainer.getBoundingClientRect();
      const mouseX = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
      const { padLeft, plotW, padTop, plotH } = this.fullscreenChartDims;
      const clampedX = Math.min(Math.max(mouseX, padLeft), padLeft + plotW);
      const frac = (clampedX - padLeft) / plotW;

      const start = this.trendTimeBounds.start;
      const end = this.trendTimeBounds.end;
      const timeSpan = Math.max(end - start, 1000);
      const isUnderOrEqual6h = timeSpan <= (6 * 3600 * 1000 + 1000);

      const hoverTime = start + frac * timeSpan;
      let effectiveTime = isUnderOrEqual6h ? (Math.round(hoverTime / 5000) * 5000) : hoverTime;
      if (isUnderOrEqual6h) {
        effectiveTime = Math.max(Math.floor(start / 5000) * 5000, Math.min(Math.ceil(end / 5000) * 5000, effectiveTime));
      }
      const snappedFrac = Math.min(Math.max((effectiveTime - start) / timeSpan, 0), 1);
      const scrubX = isUnderOrEqual6h ? (padLeft + snappedFrac * plotW) : clampedX;

      const seriesValues = [];
      for (const s of this.activeTrendSeries) {
        if (!s.dataPoints || s.dataPoints.length === 0) continue;
        let closest = s.dataPoints[0];
        let minDiff = Math.abs(closest.x - effectiveTime);
        for (let i = 1; i < s.dataPoints.length; i++) {
          const diff = Math.abs(s.dataPoints[i].x - effectiveTime);
          if (diff < minDiff) {
            minDiff = diff;
            closest = s.dataPoints[i];
          }
        }

        const scaleInfo = this.getSeriesScale(s, s.dataPoints);
        let plotY = null;
        if (closest.y !== null && isFinite(closest.y)) {
          const fracY = Math.min(Math.max((closest.y - scaleInfo.min) / scaleInfo.span, 0), 1);
          plotY = (padTop + plotH) - fracY * plotH;
        }

        const param = this.allParameters.find(p => p.param_id === s.parameterId) || {};
        let alarmStatus = 'NORMAL';
        if (closest.y !== null && typeof closest.y === 'number') {
          alarmStatus = this.determineAlarmState(param, closest.y);
        }

        seriesValues.push({
          seriesId: s.id,
          deviceCode: s.deviceCode,
          paramCode: s.paramCode,
          paramName: s.paramName || param.param_name || s.paramCode,
          val: closest.y !== null && closest.y !== undefined ? this.formatAxisValue(closest.y) : '--',
          unit: s.unit || '',
          color: s.color,
          plotY,
          quality: closest.quality || 'GOOD',
          alarmStatus,
        });
      }

      const d = new Date(effectiveTime);
      this.fullscreenScrubPoint = {
        x: scrubX,
        rawPixelX: mouseX,
        time: d.toLocaleTimeString('en-GB'),
        date: d.toLocaleDateString([], { day: '2-digit', month: 'short', year: 'numeric' }),
        allSeries: seriesValues,
      };
    },
    downloadTrendCSV() {
      if (this.activeTrendSeries.length === 0) return;
      let csv = 'Timestamp,DateTimeISO,DeviceCode,ParamCode,Value,Unit,Quality\n';
      this.activeTrendSeries.forEach(s => {
        (s.dataPoints || []).forEach(pt => {
          const iso = new Date(pt.x).toISOString();
          csv += `${pt.x},"${iso}","${s.deviceCode}","${s.paramCode}",${pt.y},"${s.unit || ''}","${pt.quality || 'GOOD'}"\n`;
        });
      });
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `trend_telemetry_${this.currentTimeframe}_${Date.now()}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    },
    downloadTrendJSON() {
      if (this.activeTrendSeries.length === 0) return;
      const payload = {
        exportedAt: new Date().toISOString(),
        timeframe: this.currentTimeframe,
        mode: this.chartMode,
        series: this.activeTrendSeries.map(s => ({
          deviceCode: s.deviceCode,
          paramCode: s.paramCode,
          unit: s.unit,
          stats: s.stats,
          points: s.dataPoints,
        })),
      };
      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `trend_telemetry_${this.currentTimeframe}_${Date.now()}.json`;
      a.click();
      URL.revokeObjectURL(url);
    },
    downloadTrendChart(format = 'png') {
      const svgEl = this.showTrendFullscreen ? this.$refs.fullscreenTrendSvg : this.$refs.trendPanelSvg;
      if (!svgEl) return;

      const serializer = new XMLSerializer();
      let svgString = serializer.serializeToString(svgEl);

      if (!svgString.match(/^<svg[^>]+xmlns="http:\/\/www\.w3\.org\/2000\/svg"/)) {
        svgString = svgString.replace(/^<svg/, '<svg xmlns="http://www.w3.org/2000/svg"');
      }

      if (format === 'svg') {
        const blob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `trend_chart_${this.currentTimeframe}_${Date.now()}.svg`;
        a.click();
        URL.revokeObjectURL(url);
        return;
      }

      const svgBlob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' });
      const url = URL.createObjectURL(svgBlob);
      const img = new Image();
      img.onload = () => {
        const canvas = document.createElement('canvas');
        const width = svgEl.clientWidth || 960;
        const height = svgEl.clientHeight || 380;
        canvas.width = width * 2;
        canvas.height = height * 2;
        const ctx = canvas.getContext('2d');
        ctx.scale(2, 2);

        ctx.fillStyle = this.isDark ? '#070D18' : '#ffffff';
        ctx.fillRect(0, 0, width, height);

        ctx.drawImage(img, 0, 0, width, height);
        URL.revokeObjectURL(url);

        const pngUrl = canvas.toDataURL('image/png');
        const a = document.createElement('a');
        a.href = pngUrl;
        a.download = `trend_chart_${this.currentTimeframe}_${Date.now()}.png`;
        a.click();
      };
      img.src = url;
    },

    // ----------------------------------------------------
    // ALARM ACKNOWLEDGEMENT
    // ----------------------------------------------------
    openAckModal(alarm) {
      this.targetAckAlarm = alarm;
      this.ackNotes = '';
      this.showAckModal = true;
    },
    async confirmAckAlarm() {
      if (!this.targetAckAlarm) return;
      this.isAckSubmitting = true;
      try {
        await this.$store.dispatch('acknowledgeAlarm', {
          id: this.targetAckAlarm.id,
          notes: this.ackNotes,
        });
        this.showAckModal = false;
        await this.fetchActiveAlarms();
      } catch (err) {
        alert('Failed to acknowledge alarm: ' + (err.message || 'Server error'));
      } finally {
        this.isAckSubmitting = false;
      }
    },
    async retryNotification(id) {
      try {
        await axios.post(`/api/notifications/${id}/retry`);
        await this.fetchNotifications();
      } catch (err) {
        console.error('Failed to retry notification:', err);
      }
    },

    // ----------------------------------------------------
    // EXPORT & REPORT WORKFLOWS
    // ----------------------------------------------------
    openExportModal() {
      const now = new Date();
      const past24h = new Date(now.getTime() - 24 * 3600 * 1000);
      this.exportForm.endTime = this.toLocalISO(now);
      this.exportForm.startTime = this.toLocalISO(past24h);
      this.exportStatusMessage = '';
      this.showExportModal = true;
    },
    async executeTelemetryExport() {
      this.isExporting = true;
      this.exportStatusMessage = 'Querying dataset from edge database...';
      try {
        const startISO = new Date(this.exportForm.startTime).toISOString();
        const endISO = new Date(this.exportForm.endTime).toISOString();

        let endpoint = '/api/telemetry/history';
        if (this.exportForm.mode === 'RAW') {
          if (!this.isSuperAdmin) {
            this.exportForm.mode = 'PROCESSED';
            endpoint = '/api/telemetry/history';
          } else {
            endpoint = '/api/telemetry/raw';
          }
        }
        else if (this.exportForm.mode === 'AGGREGATED') endpoint = '/api/aggregations/results';

        const params = {
          start_time: startISO,
          end_time: endISO,
          page_size: 5000,
        };
        if (this.exportForm.deviceId) params.device_id = this.exportForm.deviceId;

        const res = await axios.get(endpoint, { params });
        const items = res.data?.data?.items || res.data?.data || [];

        if (items.length === 0) {
          this.exportStatusMessage = 'No records found for specified interval.';
          return;
        }

        this.exportStatusMessage = `Preparing download of ${items.length} records...`;

        if (this.exportForm.format === 'json') {
          const blob = new Blob([JSON.stringify(items, null, 2)], { type: 'application/json' });
          this.triggerDownloadBlob(blob, `telemetry_export_${Date.now()}.json`);
        } else {
          // Generate CSV
          const headers = ['Timestamp', 'Device ID', 'Parameter ID', 'Value', 'Quality'];
          const rows = items.map(i => [
            i.received_at || i.timestamp || '',
            i.device_id || '',
            i.parameter_id || '',
            i.value != null ? i.value : '',
            i.quality || 'GOOD',
          ]);
          const csvContent = [headers.join(','), ...rows.map(r => r.join(','))].join('\n');
          const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
          this.triggerDownloadBlob(blob, `telemetry_export_${Date.now()}.csv`);
        }

        setTimeout(() => {
          this.showExportModal = false;
        }, 1200);
      } catch (err) {
        this.exportStatusMessage = 'Export failed: ' + (err.message || 'API error');
      } finally {
        this.isExporting = false;
      }
    },
    openReportModal() {
      this.showReportModal = true;
    },
    executeReportGenerate() {
      const now = new Date();
      let startTime = new Date(now.getTime() - 24 * 3600 * 1000);
      if (this.reportForm.preset === '1h') startTime = new Date(now.getTime() - 3600 * 1000);
      else if (this.reportForm.preset === '7d') startTime = new Date(now.getTime() - 7 * 86400 * 1000);
      else if (this.reportForm.preset === '30d') startTime = new Date(now.getTime() - 30 * 86400 * 1000);

      const params = new URLSearchParams({
        start_time: startTime.toISOString(),
        end_time: now.toISOString(),
        report_type: this.reportForm.type,
        format: this.reportForm.format,
      });

      if (this.reportForm.format === 'html') {
        params.append('download', '1');
      }
      window.open(`/api/reports/generate?${params.toString()}`, '_blank');
      this.showReportModal = false;
    },
    triggerDownloadBlob(blob, filename) {
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    },
    toLocalISO(d) {
      const pad = n => (n < 10 ? '0' + n : n);
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
    },
    formatTimeShort(ts) {
      if (!ts) return '';
      const d = new Date(ts);
      return d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    formatRelative(ts) {
      if (!ts) return '';
      const diff = Math.floor((Date.now() - new Date(ts).getTime()) / 1000);
      if (diff < 10) return 'Just now';
      if (diff < 60) return `${diff}s ago`;
      if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
      return `${Math.floor(diff / 3600)}h ago`;
    },
    formatRelativeSeconds(sec) {
      if (sec < 5) return '< 1s';
      if (sec < 60) return `${sec}s`;
      if (sec < 3600) return `${Math.floor(sec / 60)}m`;
      return `${Math.floor(sec / 3600)}h`;
    },
    getParamTimingIcon(param) {
      if (!param) return '•';
      if (param.device_status === 'OFFLINE' || param.device_status === 'ERROR') return '🔴';
      if (param.alarm_status === 'CRITICAL') return '🚨';
      if (param.alarm_status === 'WARNING') return '⚠️';
      if (param.quality === 'BAD') return '❌';
      if (param.quality === 'STALE') return '⏳';
      return '•';
    },
    getParamTimingLabel(param) {
      if (!param) return '';
      const timeStr = param.received_at ? this.formatTimeShort(param.received_at) : '';
      if (param.device_status === 'OFFLINE' || param.device_status === 'ERROR') {
        return `${this.$t('operator.offlineSince') || 'Offline'}: ${timeStr}`;
      }
      if (param.alarm_status === 'CRITICAL' || param.alarm_status === 'WARNING') {
        return `${param.alarm_status}: ${timeStr}`;
      }
      if (param.quality === 'BAD') {
        const reason = param.quality_reason ? ` (${param.quality_reason})` : '';
        return `Bad${reason}: ${timeStr}`;
      }
      if (param.quality === 'STALE') {
        return `Stale: ${timeStr}`;
      }
      return `${timeStr}`;
    },
    paramTimingColorClass(param) {
      if (!param) return '';
      if (param.alarm_status === 'CRITICAL' || param.device_status === 'OFFLINE' || param.device_status === 'ERROR' || param.quality === 'BAD') {
        return 'text-rose-600 dark:text-rose-400';
      }
      if (param.alarm_status === 'WARNING' || param.quality === 'STALE') {
        return 'text-amber-600 dark:text-amber-400';
      }
      return 'text-slate-400 dark:text-slate-500';
    },
    getParamTimingTitle(param) {
      if (!param) return '';
      const timeFull = param.received_at ? new Date(param.received_at).toLocaleString() : 'N/A';
      const timeIso = param.received_at ? new Date(param.received_at).toISOString() : 'N/A';
      const rel = param.received_at ? this.formatRelative(param.received_at) : '';
      const isPlotted = this.isParamSelectedForTrend(param);
      const trendHint = isPlotted ? ' • Plotted in Trend Analysis (Click to remove)' : ' • Click to Plot Trend';
      return `Last Reading: ${timeFull} (${timeIso}) [${rel}] | Quality: ${param.quality || 'N/A'} | Reason: ${param.quality_reason || 'NORMAL'} | Device: ${param.device_status}${trendHint}`;
    },
    toggleEnvironmentalVisualizations() {
      this.showEnvironmentalVisualizations = !this.showEnvironmentalVisualizations;
      try {
        localStorage.setItem('dl_show_environmental_viz', this.showEnvironmentalVisualizations ? 'true' : 'false');
      } catch (e) {}
    },
  },
};
</script>

<style scoped>
/* ======================================================== */
/* SCADA INDUSTRIAL THEME TOKENS                            */
/* Dark Mode: Industrial Navy                               */
/* Light Mode: Stronger Blue Gradient (Refined)             */
/* ======================================================== */

.operator-command-center {
  --occ-bg: #EDF4F8;
  --occ-bg: #F7FCFF;
  --occ-panel: #FFFFFF;
  --occ-card: #FFFFFF;
  --occ-border: #B7D8EC;
  --occ-text-primary: #15324F;
  --occ-text-secondary: #58718D;
  --occ-accent: #1685E5;
  --occ-normal: #138B68;
  --occ-warning: #A96C0A;
  --occ-alarm: #C93449;
  --occ-grid: #D7E5EF;
  --occ-axis: #B7D8EC;
}

:global(.dark) .operator-command-center {
  --occ-bg: #081421;
  --occ-bg: #061321;
  --occ-panel: #0D1B2A;
  --occ-panel: #0B1C2B;
  --occ-card: #112235;
  --occ-card: #0D2232;
  --occ-border: #294158;
  --occ-border: #244356;
  --occ-border-hover: #33566f;
  --occ-text-primary: #E6F3FC;
  --occ-text-secondary: #9BB4C8;
  --occ-accent: #21C7E8;
  --occ-normal: #36C99A;
  --occ-warning: #F1BD61;
  --occ-alarm: #FF6577;
  --occ-grid: #132738;
  --occ-axis: #244356;
}

/* Base OCC Container Layout */
.operator-command-center {
  min-height: 100%;
  min-height: calc(100vh - 36px);
}

/* 1. Light Mode: Stronger Blue-White-Light Blue Gradient Background */
.operator-dashboard.light-mode,
.operator-command-center.light-mode {
  background: linear-gradient(
    180deg,
    #8CCFFF 0%,
    #D4EDFC 14%,
    #F7FCFF 42%,
    #F7FCFF 58%,
    #E2F4FD 85%,
    #D6F0FF 100%
  ) !important;
  background-attachment: fixed;
}

/* 1b. Dark Mode: Deep Industrial Navy Background */
.operator-dashboard.dark-mode,
.operator-command-center.dark-mode {
  background: #061321 !important;
  background-attachment: scroll;
}

/* Industrial SCADA Scrollbars */
.custom-scroller::-webkit-scrollbar {
  width: 5px;
  height: 5px;
}
.custom-scroller::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroller::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.4);
  border-radius: 4px;
}
.custom-scroller::-webkit-scrollbar-thumb:hover {
  background: #1685E5;
}

/* 2. Header & Navigation Toolbar (Light Mode) */
.light-mode .scada-unified-header-card {
  background: rgba(255, 255, 255, 0.94);
  backdrop-filter: blur(8px);
  -webkit-backdrop-filter: blur(8px);
  border-color: #B7D8EC;
  box-shadow: 0 2px 8px rgba(140, 207, 255, 0.16);
}

.light-mode .scada-unified-header-card button.bg-slate-100 {
  background-color: #F0F7FD;
  border-color: #B7D8EC;
  color: #15324F;
}
.light-mode .scada-unified-header-card button.bg-slate-100:hover {
  background-color: #E2F0FB;
  border-color: #8CCFFF;
  color: #0F2338;
}

.light-mode .scada-unified-header-card .bg-slate-50 {
  background-color: #F5FAFD;
  border-color: #B7D8EC;
  color: #15324F;
}

.light-mode .scada-unified-header-card .border-t {
  border-color: #D3E7F5;
}

.light-mode .scada-unified-header-card .toolbar-popover-anchor > div.absolute {
  background-color: #FFFFFF;
  border-color: #B7D8EC;
  box-shadow: 0 10px 25px -5px rgba(140, 207, 255, 0.25), 0 8px 10px -6px rgba(0, 0, 0, 0.05);
}

.dark-mode .scada-unified-header-card {
  background-color: #0B1C2B !important;
  border-color: #244356 !important;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.35);
}

.dark-mode .scada-unified-header-card button.bg-slate-100,
.dark-mode .scada-unified-header-card .bg-slate-50 {
  background-color: #0D2232 !important;
  border-color: #244356 !important;
  color: #E6F3FC !important;
}

.dark-mode .scada-unified-header-card button.bg-slate-100:hover {
  background-color: #162f44 !important;
  border-color: #33566f !important;
}

.dark-mode .scada-unified-header-card .border-t {
  border-color: #1A3447 !important;
}

.dark-mode .scada-unified-header-card .toolbar-popover-anchor > div.absolute,
.dark-mode .toolbar-popover-anchor > div.absolute,
.dark-mode .trend-settings-anchor > div.absolute {
  background-color: #0B1C2B !important;
  border-color: #244356 !important;
  color: #E6F3FC !important;
  box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.6), 0 8px 10px -6px rgba(0, 0, 0, 0.4);
}

/* 3. Single Main Parameter Panel Container */
.light-mode .scada-single-main-panel {
  background: rgba(255, 255, 255, 0.92);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  border-color: #B7D8EC;
  box-shadow: 0 2px 6px rgba(140, 207, 255, 0.12);
}

.dark-mode .scada-single-main-panel {
  background-color: #0B1C2B !important;
  border-color: #244356 !important;
}

.dark-mode .scada-single-main-panel .border-b {
  border-bottom-color: #1A3447 !important;
}

/* 4. Parameter Cards (Light Mode: crisp solid #FFFFFF with #B7D8EC border) */
.light-mode .scada-param-tile {
  background-color: #FFFFFF !important;
  border-color: #B7D8EC !important;
  box-shadow: 0 1px 4px rgba(140, 207, 255, 0.14);
}

.light-mode .scada-param-tile:hover {
  border-color: #65BDEB !important;
  box-shadow: 0 2px 6px rgba(101, 189, 235, 0.22);
}

/* Selected Parameter Card: Cyan #22B8D4 */
.light-mode .scada-param-tile.border-cyan-500\/75,
.light-mode .scada-param-tile[class*="border-cyan-500"] {
  border-color: #22B8D4 !important;
  box-shadow: 0 0 0 1px rgba(34, 184, 212, 0.35), 0 2px 6px rgba(34, 184, 212, 0.16) !important;
}

/* Alarm Priority Over Selection */
.light-mode .scada-param-tile.border-rose-500,
.light-mode .scada-param-tile[class*="border-rose-500"] {
  border-color: #f43f5e !important;
  box-shadow: 0 1px 6px rgba(244, 63, 94, 0.25) !important;
}

.light-mode .scada-param-tile.border-amber-500,
.light-mode .scada-param-tile[class*="border-amber-500"] {
  border-color: #f59e0b !important;
  box-shadow: 0 1px 6px rgba(245, 158, 11, 0.25) !important;
}

.dark-mode .scada-param-tile {
  background-color: #0D2232 !important;
  border-color: #244356 !important;
  box-shadow: none;
}

.dark-mode .scada-param-tile:hover {
  border-color: #33566f !important;
}

.dark-mode .scada-param-tile.border-cyan-500\/75,
.dark-mode .scada-param-tile[class*="border-cyan-500"] {
  border-color: #21C7E8 !important;
  box-shadow: 0 0 0 1px rgba(33, 199, 232, 0.35) !important;
}

/* ISPU Summary Card */
.light-mode .ispu-summary-card {
  background-color: #FFFFFF !important;
  border-color: #B7D8EC !important;
  box-shadow: 0 1px 4px rgba(140, 207, 255, 0.14);
}

.dark-mode .ispu-summary-card {
  background-color: #0D2232 !important;
  border-color: #244356 !important;
  box-shadow: none;
}

/* 5. Trend Analysis Panel */
.light-mode .scada-panel {
  background: rgba(255, 255, 255, 0.94);
  border-color: #B7D8EC;
  box-shadow: 0 2px 8px rgba(140, 207, 255, 0.16);
}

.light-mode .scada-panel button.bg-slate-100 {
  background-color: #F0F7FD;
  border-color: #B7D8EC;
  color: #15324F;
}
.light-mode .scada-panel button.bg-slate-100:hover {
  background-color: #E2F0FB;
  border-color: #8CCFFF;
}

.dark-mode .scada-panel {
  background-color: #0B1C2B !important;
  border-color: #244356 !important;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
}

.dark-mode .scada-panel .border-slate-200,
.dark-mode .scada-panel .dark\:border-\[\#1E3048\],
.dark-mode .scada-panel .dark\:border-\[\#20354B\] {
  border-color: #244356 !important;
}

.dark-mode .scada-panel button.bg-slate-100,
.dark-mode .scada-panel .bg-slate-100 {
  background-color: #0D2232 !important;
  border-color: #244356 !important;
  color: #9BB4C8 !important;
}

.dark-mode .scada-panel button.bg-slate-100:hover,
.dark-mode .scada-panel .bg-slate-100:hover {
  background-color: #162f44 !important;
  border-color: #33566f !important;
  color: #E6F3FC !important;
}

/* 6. Quick Actions Docked Footer */
.light-mode .fixed.bottom-0 {
  background-color: rgba(255, 255, 255, 0.96);
  border-top-color: #B7D8EC;
  box-shadow: 0 -2px 10px rgba(140, 207, 255, 0.18);
}

.light-mode .fixed.bottom-0 a.bg-slate-100 {
  background-color: #F0F7FD;
  border-color: #B7D8EC;
  color: #15324F;
}
.light-mode .fixed.bottom-0 a.bg-slate-100:hover {
  background-color: #E2F0FB;
  border-color: #8CCFFF;
}

.dark-mode .fixed.bottom-0 {
  background-color: rgba(11, 28, 43, 0.96) !important;
  border-top-color: #244356 !important;
  box-shadow: 0 -2px 15px rgba(0, 0, 0, 0.5);
}

.dark-mode .fixed.bottom-0 a.bg-slate-100,
.dark-mode .fixed.bottom-0 button.bg-slate-100 {
  background-color: #0D2232 !important;
  border-color: #244356 !important;
  color: #9BB4C8 !important;
}

.dark-mode .fixed.bottom-0 a.bg-slate-100:hover,
.dark-mode .fixed.bottom-0 button.bg-slate-100:hover {
  background-color: #162f44 !important;
  border-color: #33566f !important;
  color: #E6F3FC !important;
}
</style>
