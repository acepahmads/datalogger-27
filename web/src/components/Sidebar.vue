<template>
  <aside
    :class="isMobileDrawer ? 'w-full flex flex-col border-r-0' : (isCollapsed ? 'hidden' : 'w-60 flex flex-col border-r border-slate-200 dark:border-slate-800/60')"
    class="bg-white dark:bg-[#0B0F19] flex-shrink-0 select-none text-slate-700 dark:text-slate-300 font-sans transition-all duration-150 h-full"
  >
    <!-- Brand Header -->
    <div class="px-5 py-4 border-b border-slate-200 dark:border-slate-800/40 flex items-start justify-between">
      <div class="min-w-0 pr-2">
        <div class="text-sm font-black text-slate-900 dark:text-white tracking-wider truncate">DATALOGGER</div>
        <div class="text-2xs text-slate-500 dark:text-slate-400 font-medium truncate">Analysis Application</div>
        <div
          :class="isMobileDrawer
            ? 'inline-flex items-center space-x-1.5 mt-2 px-2 py-0.5 rounded-full bg-emerald-500/10 border border-emerald-500/20'
            : 'flex items-center space-x-1.5 mt-2'"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
          <span class="text-3xs text-emerald-600 dark:text-emerald-400 font-medium tracking-wide">{{ isCloudMode ? 'Cloud Receiver Active' : $t('navigation.edgeSystem') }}</span>
        </div>
      </div>
      <!-- Setup / Auto-Hide Pin Icon Button (Desktop) -->
      <button
        v-if="!isMobileDrawer"
        type="button"
        id="btn-sidebar-autohide-setup"
        @click.stop="toggleAutoHide"
        class="p-1.5 rounded-lg border transition-all flex items-center justify-center mt-0.5 flex-shrink-0 cursor-pointer"
        :class="sidebarAutoHide
          ? 'bg-cyan-500/20 text-cyan-400 border-cyan-500/50 hover:bg-cyan-500/30 ring-1 ring-cyan-500/30'
          : 'bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 border-slate-200 dark:border-slate-700/60'"
        :title="sidebarAutoHide ? ($t('header.sidebarAutoHideOn') || 'Sidebar: Auto-Hide Enabled (Click to Pin)') : ($t('header.sidebarAutoHideOff') || 'Sidebar: Pinned (Click for Auto-Hide)')"
        aria-label="Toggle Sidebar Auto-Hide"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <path d="M9 3v18" stroke-linecap="round" />
          <path v-if="!sidebarAutoHide" stroke-linecap="round" stroke-linejoin="round" d="m14 9-3 3 3 3" />
          <path v-else stroke-linecap="round" stroke-linejoin="round" d="m12 9 3 3-3 3" />
        </svg>
      </button>
      <!-- Close Drawer Button (Mobile Slide-Over) -->
      <button
        v-else
        type="button"
        id="btn-sidebar-mobile-close"
        @click.stop="$emit('close-mobile')"
        class="w-7 h-7 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/70 dark:hover:bg-slate-800 text-slate-500 hover:text-slate-800 dark:text-slate-300 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 transition-colors flex items-center justify-center mt-0.5 flex-shrink-0 cursor-pointer"
        aria-label="Close Navigation Menu"
        title="Tutup Menu"
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>

    <!-- Navigation List -->
    <nav class="flex-1 overflow-y-auto px-3 py-4 space-y-6" @click="onNavClick">
      <!-- DEVELOPMENT (Strictly SUPERADMIN Only) -->
      <div v-if="isSuperAdmin">
        <div class="px-3 text-3xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2 font-sans">
          {{ $t('navigation.development') }}
        </div>
        <div class="space-y-1">
          <router-link to="/development" class="nav-item" :class="{ 'nav-active': $route.path === '/development' || $route.path === '/development/phases' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <rect x="3" y="3" width="7" height="7" rx="1"></rect>
              <rect x="14" y="3" width="7" height="7" rx="1"></rect>
              <rect x="14" y="14" width="7" height="7" rx="1"></rect>
              <rect x="3" y="14" width="7" height="7" rx="1"></rect>
            </svg>
            <span>{{ $t('navigation.overviewPhases') }}</span>
          </router-link>

          <router-link to="/development/tasks" class="nav-item" :class="{ 'nav-active': $route.path === '/development/tasks' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 11l3 3L22 4"></path>
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 12v7a2 2 0 01-2 2H5a2 2 0 01-2-2V5a2 2 0 012-2h11"></path>
            </svg>
            <span>{{ $t('navigation.tasksManager') }}</span>
            <span class="ml-auto text-3xs font-mono px-1.5 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
              {{ progress.total_tasks_count || 117 }}
            </span>
          </router-link>

          <router-link to="/development/activity" class="nav-item" :class="{ 'nav-active': $route.path === '/development/activity' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
            </svg>
            <span>{{ $t('navigation.activityLog') }}</span>
          </router-link>

          <router-link to="/development/release" class="nav-item" :class="{ 'nav-active': $route.path === '/development/release' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"></path>
              <polyline points="14 2 14 8 20 8"></polyline>
              <line x1="16" y1="13" x2="8" y2="13"></line>
              <line x1="16" y1="17" x2="8" y2="17"></line>
              <polyline points="10 9 9 9 8 9"></polyline>
            </svg>
            <span>{{ $t('navigation.releaseDocs') }}</span>
          </router-link>
        </div>
      </div>

      <!-- MONITORING -->
      <div>
        <div class="px-3 text-3xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2 font-sans flex items-center justify-between">
          <span>{{ $t('navigation.monitoring') }}</span>
        </div>
        <div class="space-y-1">
          <!-- Superadmin Global Fleet & Multi-Tenant Map Dashboard (STRICTLY SUPERADMIN ONLY) -->
          <router-link
            v-if="isSuperAdmin"
            id="nav-fleet-map"
            to="/monitoring/fleet-map"
            class="nav-item"
            :class="{ 'nav-active': $route.path === '/monitoring/fleet-map' || $route.path === '/fleet-map' }"
          >
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-emerald-500 dark:text-emerald-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <circle cx="12" cy="12" r="9" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.6 9h16.8M3.6 15h16.8M12 3a14.5 14.5 0 000 18M12 3a14.5 14.5 0 010 18" />
            </svg>
            <span class="font-bold text-emerald-600 dark:text-emerald-400">{{ $t('navigation.fleetMap') || 'Peta Tenant & Stasiun' }}</span>
            <span class="ml-auto text-4xs font-mono font-bold px-1.5 py-0.5 rounded-full bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30">
              MAP
            </span>
          </router-link>

          <!-- Operator Command Center / SCADA Console (Hidden for Operator role, and hidden for Admin & Engineer in Cloud mode) -->
          <router-link v-if="!isOperator && !((isAdmin || isEngineer) && isCloudMode)" to="/monitoring/operator" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/operator' || $route.path === '/operator' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-cyan-500 dark:text-cyan-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5m.75-9l3-3 2.25 2.25L15 6" />
            </svg>
            <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ $t('navigation.operatorConsole') || 'Operator Console' }}</span>
            <span class="ml-auto text-4xs font-mono font-bold px-1.5 py-0.5 rounded-full bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30">
              CBI
            </span>
          </router-link>

          <!-- Dashboard Operator (New Clean Asymmetric Bento Design) -->
          <router-link to="/monitoring/dashboard-operator" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/dashboard-operator' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-sky-500 dark:text-sky-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <rect x="3" y="3" width="7" height="9" rx="1.5" />
              <rect x="14" y="3" width="7" height="5" rx="1.5" />
              <rect x="14" y="12" width="7" height="9" rx="1.5" />
              <rect x="3" y="16" width="7" height="5" rx="1.5" />
            </svg>
            <span class="font-bold text-sky-600 dark:text-sky-400">{{ $t('navigation.dashboardOperator') || 'Dashboard Operator' }}</span>
          </router-link>

          <router-link v-if="!isOperator || !isCloudMode" to="/monitoring/system" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/system' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <rect x="4" y="4" width="16" height="16" rx="2"></rect>
              <rect x="9" y="9" width="6" height="6"></rect>
              <line x1="9" y1="1" x2="9" y2="4"></line>
              <line x1="15" y1="1" x2="15" y2="4"></line>
              <line x1="9" y1="20" x2="9" y2="23"></line>
              <line x1="15" y1="20" x2="15" y2="23"></line>
              <line x1="20" y1="9" x2="23" y2="9"></line>
              <line x1="20" y1="14" x2="23" y2="14"></line>
              <line x1="1" y1="9" x2="4" y2="9"></line>
              <line x1="1" y1="14" x2="4" y2="14"></line>
            </svg>
            <span>{{ $t('navigation.systemMonitoring') }}</span>
          </router-link>

          <router-link
            v-if="isSuperAdmin || !isCloudMode"
            to="/monitoring/devices"
            class="nav-item"
            :class="{ 'nav-active': $route.path === '/monitoring/devices' || $route.path === '/monitoring/parameters' }"
          >
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <circle cx="12" cy="12" r="2"></circle>
              <path d="M16.24 7.76a6 6 0 0 1 0 8.49m-8.48-.01a6 6 0 0 1 0-8.49m11.31-2.82a10 10 0 0 1 0 14.14m-14.14 0a10 10 0 0 1 0-14.14"></path>
            </svg>
            <span>{{ $t('navigation.devicesSensors') }}</span>
          </router-link>

          <router-link
            v-if="isSuperAdmin || (!isOperator && !isCloudMode)"
            to="/monitoring/calibration"
            class="nav-item"
            :class="{ 'nav-active': $route.path.startsWith('/monitoring/calibration') }"
          >
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-sky-500 dark:text-sky-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.85">
              <circle cx="12" cy="12" r="8"></circle>
              <circle cx="12" cy="12" r="3"></circle>
              <line x1="12" y1="2" x2="12" y2="4"></line>
              <line x1="12" y1="20" x2="12" y2="22"></line>
              <line x1="2" y1="12" x2="4" y2="12"></line>
              <line x1="20" y1="12" x2="22" y2="12"></line>
            </svg>
            <span class="font-semibold text-sky-600 dark:text-sky-300">{{ $t('navigation.sensorCalibration') || 'Kalibrasi Sensor' }}</span>
          </router-link>

          <router-link v-if="!isOperator" to="/monitoring/telemetry" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/telemetry' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-emerald-500 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline>
            </svg>
            <span class="font-medium text-emerald-600 dark:text-emerald-300">{{ $t('navigation.liveTelemetry') }}</span>
            <span class="ml-auto flex h-2 w-2 relative">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
          </router-link>

          <router-link to="/monitoring/analysis" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/analysis' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-blue-500 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 013 19.875v-6.75zM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V8.625zM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 01-1.125-1.125V4.125z" />
            </svg>
            <span class="font-medium text-slate-200">{{ $t('navigation.telemetryAnalysis') }}</span>
          </router-link>

          <!-- Telemetry Injector (SUPERADMIN ONLY) -->
          <router-link
            v-if="isSuperAdmin"
            to="/monitoring/telemetry-injection"
            class="nav-item"
            :class="{ 'nav-active': $route.path === '/monitoring/telemetry-injection' }"
          >
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-violet-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.85">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.428 15.428a2 2 0 00-1.022-.547l-2.387-.477a6 6 0 00-3.86.517l-.318.158a6 6 0 01-3.86.517L6.05 15.21a2 2 0 00-1.806.547M8 4h8l-1 1v5.172a2 2 0 00.586 1.414l5 5c1.26 1.26.367 3.414-1.415 3.414H4.828c-1.782 0-2.674-2.154-1.414-3.414l5-5A2 2 0 009 10.172V5L8 4z" />
            </svg>
            <span class="font-semibold text-violet-600 dark:text-violet-300">Telemetry Injector</span>
            <span class="ml-auto text-3xs font-mono px-1.5 py-0.2 rounded bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800 uppercase tracking-tighter">
              SA
            </span>
          </router-link>

          <!-- ISPU Compliance Module (Permen LHK No. 14 Tahun 2020) -->
          <router-link v-if="showIspuMenu" to="/monitoring/ispu" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/ispu' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-cyan-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 3v2.25m6.364.386l-1.591 1.591M21 12h-2.25m-.386 6.364l-1.591-1.591M12 18.75V21m-4.773-4.227l-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0z" />
            </svg>
            <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ $t('navigation.ispuAnalysis') || 'ISPU Analysis' }}</span>
            <span class="ml-auto text-4xs font-mono font-bold px-1.5 py-0.5 rounded-full bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30">
              P.14
            </span>
          </router-link>

          <!-- WQMS IP & IKA Compliance Module (Kepmen LH 115/2003, PP 22/2021 & Permen LHK 27/2021) -->
          <router-link v-if="showWqmsMenu" to="/monitoring/wqms" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/wqms' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-teal-500 dark:text-teal-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 2.25c-4.97 6.08-7.5 9.85-7.5 13.125a7.5 7.5 0 0015 0c0-3.275-2.53-7.045-7.5-13.125z" />
            </svg>
            <span class="font-bold text-teal-600 dark:text-teal-400">{{ $t('navigation.wqmsAnalysis') || 'WQMS Analysis (IP & IKA)' }}</span>
            <span class="ml-auto text-4xs font-mono font-bold px-1.5 py-0.5 rounded-full bg-teal-500/10 text-teal-600 dark:text-teal-400 border border-teal-500/30">
              PP.22
            </span>
          </router-link>

          <router-link v-if="!isOperator" to="/monitoring/reports" class="nav-item" :class="{ 'nav-active': $route.path === '/monitoring/reports' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
            </svg>
            <span>{{ $t('navigation.reports') || 'Reports' }}</span>
          </router-link>

          <router-link v-if="isSuperAdmin" to="/dashboard" class="nav-item" :class="{ 'nav-active': $route.path === '/dashboard' || $route.path === '/quality' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
            </svg>
            <span>{{ $t('common.quality') }}</span>
          </router-link>
        </div>
      </div>

      <!-- ALARMS -->
      <div v-if="!isOperator || !isCloudMode">
        <div class="px-3 text-3xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2 font-sans flex items-center justify-between">
          <span>{{ $t('navigation.alarmsEvents') }}</span>
          <span v-if="systemStatus.active_alarms > 0" class="w-1.5 h-1.5 rounded-full bg-rose-500 animate-ping"></span>
        </div>
        <div class="space-y-1">
          <router-link to="/alarm/active" class="nav-item" :class="{ 'nav-active': $route.path === '/alarm/active' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
            </svg>
            <span>{{ $t('navigation.activeAlarms') }}</span>
            <span v-if="systemStatus.active_alarms > 0" class="ml-auto text-3xs font-mono px-1.5 py-0.5 rounded-full bg-rose-100 dark:bg-rose-950 text-rose-600 dark:text-rose-400 border border-rose-300 dark:border-rose-800">
              {{ systemStatus.active_alarms }}
            </span>
          </router-link>

          <router-link to="/alarm/history" class="nav-item" :class="{ 'nav-active': $route.path === '/alarm/history' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <circle cx="12" cy="12" r="10"></circle>
              <polyline points="12 6 12 12 16 14"></polyline>
            </svg>
            <span>{{ $t('navigation.alarmHistory') }}</span>
          </router-link>

          <router-link to="/alarm/queue" class="nav-item" :class="{ 'nav-active': $route.path === '/alarm/queue' || $route.path === '/alarm/delivery' || ($route.path === '/alarm/active' && $route.query.tab === 'queue') }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"></path>
            </svg>
            <span>{{ $t('navigation.deliveryStatus') || 'Delivery Status' }}</span>
          </router-link>

          <router-link v-if="!isOperator && !((isAdmin || isEngineer) && isCloudMode)" to="/alarm/live-comm" class="nav-item" :class="{ 'nav-active': $route.path === '/alarm/live-comm' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
            </svg>
            <span>{{ $t('navigation.liveComm') || 'Live Sensor Comm' }}</span>
            <span v-if="liveCommActive" class="ml-auto flex h-2 w-2 relative" :title="$t('navigation.liveCommActiveTooltip') || 'Live Stream Active'">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
            </span>
          </router-link>

          <!-- System Activity Monitor (SUPERADMIN ONLY) -->
          <router-link
            v-if="isSuperAdmin"
            to="/alarm/activity-monitor"
            class="nav-item"
            :class="{ 'nav-active': $route.path === '/alarm/activity-monitor' }"
          >
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
            <span>{{ $t('navigation.activityMonitor') || 'System Activity Monitor' }}</span>
            <span class="ml-auto text-3xs font-mono px-1.5 py-0.2 rounded bg-indigo-50 dark:bg-indigo-950/60 text-indigo-600 dark:text-indigo-400 border border-indigo-200 dark:border-indigo-800 uppercase tracking-tighter">
              SA
            </span>
          </router-link>
        </div>
      </div>

      <!-- SYSTEM -->
      <div v-if="!isOperator || !isCloudMode">
        <div class="px-3 text-3xs font-semibold text-slate-400 dark:text-slate-500 uppercase tracking-widest mb-2 font-sans">
          {{ $t('navigation.systemAdmin') }}
        </div>
        <div class="space-y-1">
          <router-link v-if="!isOperator" to="/system/logs" class="nav-item" :class="{ 'nav-active': $route.path.startsWith('/system/logs') || $route.path.startsWith('/system/audit') || $route.path.startsWith('/system/communication') }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <polyline points="4 17 10 11 4 5"></polyline>
              <line x1="12" y1="19" x2="20" y2="19"></line>
            </svg>
            <span>{{ $t('navigation.systemLogs') }}</span>
          </router-link>

          <router-link v-if="!isOperator && !((isAdmin || isEngineer) && isCloudMode)" to="/administration/config" class="nav-item" :class="{ 'nav-active': $route.path === '/administration/config' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <circle cx="12" cy="12" r="3"></circle>
              <path stroke-linecap="round" stroke-linejoin="round" d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path>
            </svg>
            <span>{{ $t('navigation.configuration') }}</span>
          </router-link>

          <router-link v-if="!isOperator && !((isAdmin || isEngineer) && isCloudMode)" to="/administration/backup" class="nav-item" :class="{ 'nav-active': $route.path === '/administration/backup' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 7v10c0 2 1.5 3 3.5 3h9c2 0 3.5-1 3.5-3V7M4 7c0-2 1.5-3 3.5-3h9c2 0 3.5 1 3.5 3M4 7h16m-8 4v6m0 0l-3-3m3 3l3-3"></path>
            </svg>
            <span>{{ $t('navigation.backupRestore') }}</span>
          </router-link>

          <router-link v-if="!isOperator && !((isAdmin || isEngineer) && isCloudMode)" to="/administration/retention" class="nav-item" :class="{ 'nav-active': $route.path === '/administration/retention' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
            </svg>
            <span>{{ $t('navigation.retentionStorage') }}</span>
          </router-link>

          <router-link v-if="isSuperAdmin || isAdmin || (isEngineer && isCloudMode)" to="/administration/tenants" class="nav-item" :class="{ 'nav-active': $route.path.startsWith('/administration/tenants') }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-purple-500 dark:text-purple-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.85">
              <path stroke-linecap="round" stroke-linejoin="round" d="M3 21h18M5 21V7l8-4v18M19 21V11l-6-4M9 9v.01M9 12v.01M9 15v.01M9 18v.01" />
            </svg>
            <span class="font-semibold text-purple-600 dark:text-purple-300">Tenant &amp; Stasiun</span>
          </router-link>

          <router-link v-if="isSuperAdmin || isAdmin" to="/administration/users" class="nav-item" :class="{ 'nav-active': $route.path.startsWith('/administration/users') || $route.path.startsWith('/administration/roles') }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
              <circle cx="9" cy="7" r="4"></circle>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
            </svg>
            <span>{{ $t('navigation.usersRoles') || 'Users & RBAC' }}</span>
          </router-link>

          <router-link v-if="!isOperator" to="/integration" class="nav-item" :class="{ 'nav-active': $route.path === '/integration' && $route.query.tab !== 'queue' }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <circle cx="18" cy="5" r="3"></circle>
              <circle cx="6" cy="12" r="3"></circle>
              <circle cx="18" cy="19" r="3"></circle>
              <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line>
              <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>
            </svg>
            <span>{{ $t('navigation.integrations') || 'Outbound & Integration' }}</span>
          </router-link>

          <router-link v-if="!isOperator || !isCloudMode" to="/integration/queue" class="nav-item" :class="{ 'nav-active': $route.path === '/integration/queue' || ($route.path === '/integration' && $route.query.tab === 'queue') }">
            <span class="active-indicator"></span>
            <svg class="w-4 h-4 mr-2.5 text-slate-400 nav-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.75">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 10h16M4 14h16M4 18h16" />
            </svg>
            <span>{{ $t('navigation.outboundQueue') || 'Delivery Queue' }}</span>
          </router-link>
        </div>
      </div>
    </nav>

    <!-- Bottom Footer: Database Indicator (Hidden for Operator) & Logout Menu -->
    <div class="border-t border-slate-200 dark:border-slate-800/50 bg-slate-50 dark:bg-[#0A0E18]/60">
      <!-- Bottom Database Indicator (MariaDB Edge ● LOCAL) — Hidden for Operator Role -->
      <div v-if="!isOperator" class="px-4 pt-3.5 pb-3 border-b border-slate-200/80 dark:border-slate-800/50">
        <div class="flex items-center justify-between text-2xs mb-1.5">
          <span class="font-medium text-slate-700 dark:text-slate-200">{{ $t('navigation.edgeDatabase') }}</span>
          <span class="flex items-center text-emerald-600 dark:text-emerald-400 text-3xs font-semibold tracking-wide">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 mr-1"></span>
            LOCAL
          </span>
        </div>
        <div class="text-3xs text-slate-500 font-mono flex items-center justify-between">
          <span>127.0.0.1:3306</span>
          <span class="text-slate-600 dark:text-slate-400">10.4.x</span>
        </div>
      </div>

      <!-- Logout Menu Action -->
      <div class="p-3">
        <button
          type="button"
          id="btn-sidebar-logout"
          @click="handleSidebarLogout"
          class="w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs font-semibold text-rose-600 dark:text-rose-400 hover:bg-rose-50 dark:hover:bg-rose-500/10 border border-transparent hover:border-rose-200/80 dark:hover:border-rose-500/20 transition-colors cursor-pointer"
        >
          <span class="flex items-center">
            <svg class="w-4 h-4 mr-2.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="1.9">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
            <span>{{ $t('auth.logout') || 'Logout' }}</span>
          </span>
          <span v-if="currentUsername" class="text-3xs font-mono font-medium px-1.5 py-0.5 rounded bg-slate-200/70 dark:bg-slate-800 text-slate-600 dark:text-slate-400 truncate max-w-[96px]">
            @{{ currentUsername }}
          </span>
        </button>
      </div>
    </div>
  </aside>
</template>

<script>
export default {
  name: 'Sidebar',
  props: {
    isCollapsed: {
      type: Boolean,
      default: false,
    },
    isAutoHide: {
      type: Boolean,
      default: false,
    },
    isMobileDrawer: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      liveCommActive: false,
    };
  },
  computed: {
    userRole() {
      const u = this.$store.state.user || {};
      return (u.role || '').toUpperCase();
    },
    currentUsername() {
      const u = this.$store.state.user || {};
      return u.username || '';
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    isAdmin() {
      return this.userRole === 'ADMIN';
    },
    isEngineer() {
      return this.userRole === 'ENGINEER';
    },
    isOperator() {
      return this.userRole === 'OPERATOR';
    },
    sidebarAutoHide() {
      return !!((this.$store.state.uiPreferences && this.$store.state.uiPreferences.sidebarAutoHide) || this.$store.state.sidebarAutoHide);
    },
    progress() {
      return this.$store.state.progress;
    },
    systemStatus() {
      return this.$store.state.systemStatus;
    },
    monitoringDomain() {
      return String((this.systemStatus && this.systemStatus.monitoring_domain) || 'AQMS').toUpperCase();
    },
    isCloudMode() {
      return Boolean(
        (this.systemStatus && this.systemStatus.cloud_mode) ||
        this.monitoringDomain === 'CLOUD'
      );
    },
    showIspuMenu() {
      if (this.monitoringDomain === 'AQMS' || this.monitoringDomain === 'CUSTOM') {
        return true;
      }
      if (this.isCloudMode && this.monitoringDomain === 'CLOUD') {
        return this.isSuperAdmin;
      }
      return false;
    },
    showWqmsMenu() {
      if (this.monitoringDomain === 'WQMS' || this.monitoringDomain === 'ONLIMO' || this.monitoringDomain === 'CUSTOM') {
        return true;
      }
      if (this.isCloudMode && this.monitoringDomain === 'CLOUD') {
        return this.isSuperAdmin;
      }
      return false;
    },
  },
  methods: {
    toggleAutoHide() {
      this.$store.dispatch('toggleSidebarAutoHide');
      this.$emit('toggle-autohide', this.sidebarAutoHide);
    },
    onNavClick() {
      this.$emit('navigate');
    },
    async handleSidebarLogout() {
      this.$emit('navigate');
      if (this.$listeners && this.$listeners['request-logout']) {
        this.$emit('request-logout');
        return;
      }
      await this.$store.dispatch('logout');
      if (this.$route.path !== '/login') {
        this.$router.push('/login');
      }
    },
    onTraceStatus(e) {
      if (e && e.detail) {
        this.liveCommActive = !!e.detail.active;
      }
    },
    async checkLiveCommStatus() {
      const res = await this.$store.dispatch('fetchLiveCommStatus');
      if (res) {
        this.liveCommActive = !!res.active;
      }
    },
  },
  mounted() {
    this.checkLiveCommStatus();
    window.addEventListener('comm-trace-status', this.onTraceStatus);
  },
  beforeDestroy() {
    window.removeEventListener('comm-trace-status', this.onTraceStatus);
  },
};
</script>

<style scoped>
.nav-item {
  @apply relative flex items-center px-3 py-2 rounded-lg text-xs font-medium text-slate-600 dark:text-slate-400 hover:text-[#1685E5] dark:hover:text-slate-100 hover:bg-[#EAF5FF] dark:hover:bg-slate-800/40 transition-colors;
}
.nav-active {
  @apply bg-[#EAF5FF] dark:bg-blue-600/10 text-[#1685E5] dark:text-blue-400 font-semibold border border-[#B8D4EA]/80 dark:border-transparent shadow-sm;
}
.nav-active .nav-icon {
  @apply text-[#1685E5] dark:text-blue-400;
}
.active-indicator {
  @apply hidden;
}
.nav-active .active-indicator {
  @apply block absolute left-0 top-2 bottom-2 w-0.5 bg-[#1685E5] dark:bg-blue-500 rounded-r;
}
</style>
