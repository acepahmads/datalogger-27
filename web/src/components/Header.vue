<template>
  <header class="relative bg-white/95 dark:bg-[#0B0F19]/95 border-b border-slate-200 dark:border-slate-800/60 text-slate-800 dark:text-slate-200 px-3.5 sm:px-4 py-1.5 flex items-center justify-between sticky top-0 z-40 select-none transition-colors duration-150">
    <!-- Backdrop Blur Layer (Isolated sibling so header does not form a containing block for modals) -->
    <div class="absolute inset-0 backdrop-blur-md -z-10 pointer-events-none"></div>
    <!-- Left: Logo & App Title -->
    <div class="flex items-center space-x-2.5">
      <div class="w-7 h-7 rounded-lg bg-gradient-to-br from-blue-600 to-indigo-700 flex items-center justify-center shadow-soft text-white font-bold text-xs tracking-wider">
        DL
      </div>
      <div class="flex items-center space-x-2">
        <h1 class="text-xs font-bold tracking-tight text-slate-900 dark:text-white uppercase font-sans">
          {{ $t('header.appTitle') }}
        </h1>
        <span class="bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20 text-3xs px-1.5 py-0.5 rounded-full font-mono font-medium">
          {{ $t('header.version') }}
        </span>
      </div>
    </div>

    <!-- Right: Controls & User -->
    <div class="flex items-center space-x-1.5 sm:space-x-2">
      <!-- 0. Sidebar Auto-Hide Setup Icon Button -->
      <button
        type="button"
        id="btn-sidebar-autohide-toggle"
        @click.stop="toggleSidebarAutoHide"
        class="relative w-7 h-7 rounded-lg border transition-colors flex items-center justify-center shadow-xs cursor-pointer"
        :class="sidebarAutoHide
          ? 'bg-cyan-500/20 text-cyan-400 border-cyan-500/50 ring-1 ring-cyan-500/30'
          : 'bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border-slate-200 dark:border-slate-700/60'"
        :title="sidebarAutoHide ? ($t('header.sidebarAutoHideOn') || 'Sidebar: Auto-Hide Enabled (Click to Pin)') : ($t('header.sidebarAutoHideOff') || 'Sidebar: Pinned (Click for Auto-Hide)')"
        aria-label="Toggle Sidebar Auto-Hide"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <path d="M9 3v18" />
          <path v-if="!sidebarAutoHide" d="m14 9-3 3 3 3" stroke-linecap="round" stroke-linejoin="round" />
          <path v-else d="m12 9 3 3-3 3" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
      </button>

      <!-- 1. Operator Console Quick Icon Button -->
      <router-link
        to="/monitoring/operator"
        class="relative w-7 h-7 rounded-lg border transition-colors flex items-center justify-center shadow-xs"
        :class="$route.path.includes('/operator')
          ? 'bg-cyan-500/20 text-cyan-400 border-cyan-500/50 ring-1 ring-cyan-500/30'
          : 'bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-cyan-600 dark:text-cyan-400 hover:text-cyan-500 dark:hover:text-cyan-300 border-slate-200 dark:border-slate-700/60'"
        :title="$t('navigation.operatorConsole') || 'Operator Console'"
        aria-label="Operator Console"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5m.75-9l3-3 2.25 2.25L15 6" />
        </svg>
        <!-- Pulsing live indicator dot -->
        <span class="absolute top-0.5 right-0.5 flex h-1.5 w-1.5">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-cyan-400 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-1.5 w-1.5 bg-cyan-500"></span>
        </span>
      </router-link>

      <!-- 2. Notification Bell with Unread / Active Alarms Badge -->
      <router-link
        to="/alarm/active"
        class="relative w-7 h-7 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-600 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 transition-colors flex items-center justify-center shadow-xs"
        :title="$t('navigation.alarmsEvents') || 'Active Alarms'"
        aria-label="Active Alarms"
      >
        <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
        </svg>
        <span
          v-if="activeAlarmsCount > 0"
          class="absolute -top-1 -right-1 px-1.5 py-0.2 rounded-full text-4xs font-mono font-bold bg-rose-500 text-white shadow-xs animate-pulse"
        >
          {{ activeAlarmsCount }}
        </span>
      </router-link>

      <!-- 3. Language Switch Icon Button (EN / ID Slide Switch) -->
      <button
        type="button"
        @click="toggleLanguage"
        class="h-7 px-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 flex items-center space-x-1 transition-all shadow-xs"
        :title="$t('header.language') + ': ' + (language === 'id' ? 'Bahasa Indonesia (Click to switch to English)' : 'English (Click to switch to Bahasa Indonesia)')"
        aria-label="Toggle Language"
      >
        <svg class="w-3.5 h-3.5 text-cyan-600 dark:text-cyan-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
          <circle cx="12" cy="12" r="10" />
          <path stroke-linecap="round" stroke-linejoin="round" d="M2 12h20M12 2a15.3 15.3 0 014 10 15.3 15.3 0 01-4 10 15.3 15.3 0 01-4-10 15.3 15.3 0 014-10z" />
        </svg>
        <div class="relative w-7 h-3.5 bg-slate-200 dark:bg-slate-700 rounded-full flex items-center p-0.5 transition-colors">
          <span
            class="text-[7px] font-bold font-mono text-center flex-1 z-10 transition-colors"
            :class="language === 'en' ? 'text-cyan-700 dark:text-cyan-300 font-extrabold' : 'text-slate-400 dark:text-slate-500'"
          >EN</span>
          <span
            class="text-[7px] font-bold font-mono text-center flex-1 z-10 transition-colors"
            :class="language === 'id' ? 'text-cyan-700 dark:text-cyan-300 font-extrabold' : 'text-slate-400 dark:text-slate-500'"
          >ID</span>
          <div
            class="absolute top-0.5 bottom-0.5 w-[13px] bg-white dark:bg-cyan-500 rounded-full shadow-xs transition-transform duration-200"
            :class="language === 'id' ? 'translate-x-[13px]' : 'translate-x-0'"
          ></div>
        </div>
      </button>

      <!-- 4. Theme Switch Icon Button (Light / Dark Slide Switch) -->
      <button
        type="button"
        @click="toggleTheme"
        class="h-7 px-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800/60 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300 hover:text-slate-900 dark:hover:text-white border border-slate-200 dark:border-slate-700/60 flex items-center space-x-1 transition-all shadow-xs"
        :title="$t('header.appearance') + ': ' + (isDarkMode ? $t('header.dark') : $t('header.light')) + ' (Click to toggle)'"
        aria-label="Toggle Dark / Light Theme"
      >
        <span v-if="!isDarkMode" class="text-amber-500 text-2xs leading-none flex-shrink-0">☀</span>
        <span v-else class="text-blue-400 text-2xs leading-none flex-shrink-0">🌙</span>
        <div class="relative w-7 h-3.5 bg-slate-200 dark:bg-slate-700 rounded-full flex items-center p-0.5 transition-colors">
          <span class="text-[7px] text-center flex-1 z-10 leading-none">☀</span>
          <span class="text-[7px] text-center flex-1 z-10 leading-none">🌙</span>
          <div
            class="absolute top-0.5 bottom-0.5 w-[13px] bg-white dark:bg-blue-500 rounded-full shadow-xs transition-transform duration-200"
            :class="isDarkMode ? 'translate-x-[13px]' : 'translate-x-0'"
          ></div>
        </div>
      </button>

      <!-- 5. User Profile Badge & Quick Role Switcher -->
      <div class="relative pl-2.5 border-l border-slate-200 dark:border-slate-800" ref="userDropdown">
        <button
          type="button"
          @click="isUserMenuOpen = !isUserMenuOpen"
          class="flex items-center space-x-1.5 p-0.5 sm:p-1 rounded-xl hover:bg-slate-100 dark:hover:bg-slate-800/60 transition-colors cursor-pointer text-left"
          title="Switch User Role / Account"
        >
          <div class="w-6.5 h-6.5 rounded-full flex items-center justify-center text-3xs font-bold font-sans ring-1"
               :class="roleAvatarClass">
            {{ user.username ? user.username.charAt(0).toUpperCase() : 'A' }}
          </div>
          <div class="hidden sm:block text-left font-sans">
            <div class="text-xs font-bold text-slate-800 dark:text-slate-100 leading-tight">@{{ user.username }}</div>
            <div class="inline-flex items-center gap-1 px-1.5 py-0.2 rounded-full text-4xs font-mono font-bold mt-0.5"
                 :class="roleBadgeClass">
              <span class="w-1.5 h-1.5 rounded-full" :class="roleDotClass"></span>
              {{ userRole }}
            </div>
          </div>
          <svg class="w-3.5 h-3.5 text-slate-400 ml-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        <!-- Dropdown Menu -->
        <div
          v-if="isUserMenuOpen"
          class="absolute right-0 mt-2 w-72 bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-xl z-50 p-2 text-xs font-sans animate-fade-in"
        >
          <!-- Active Session Info Card -->
          <div class="px-3 py-2.5 border-b border-slate-100 dark:border-slate-800">
            <div class="text-2xs text-slate-400 font-semibold uppercase tracking-wider flex items-center justify-between">
              <span>{{ $t('auth.activeSession') || 'Active Session' }}</span>
              <span class="w-1.5 h-1.5 rounded-full" :class="roleDotClass"></span>
            </div>
            <div class="font-bold text-slate-900 dark:text-white mt-0.5 truncate">{{ user.full_name || user.username }}</div>
            <div class="text-3xs font-mono text-slate-400">@{{ user.username }} &bull; {{ userRole }}</div>
          </div>

          <!-- Standard Account Actions -->
          <div class="py-1 border-b border-slate-100 dark:border-slate-800 space-y-0.5">
            <!-- 1. Profile / Account -->
            <button
              type="button"
              @click="openProfileModal"
              class="w-full text-left px-3 py-2 rounded-xl flex items-center gap-2.5 hover:bg-slate-100 dark:hover:bg-slate-800/80 text-slate-700 dark:text-slate-200 transition-colors"
            >
              <svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
              </svg>
              <div>
                <div class="font-semibold text-xs leading-none">{{ $t('auth.profile') || 'Profile / Account' }}</div>
                <div class="text-4xs text-slate-400 mt-0.5">View identity & capabilities</div>
              </div>
            </button>

            <!-- 2. Switch Account / Login as Another User -->
            <button
              type="button"
              @click="openSwitchModal"
              class="w-full text-left px-3 py-2 rounded-xl flex items-center gap-2.5 hover:bg-slate-100 dark:hover:bg-slate-800/80 text-slate-700 dark:text-slate-200 transition-colors"
            >
              <svg class="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
              <div>
                <div class="font-semibold text-xs leading-none">{{ $t('auth.switchAccount') || 'Switch Account' }}</div>
                <div class="text-4xs text-slate-400 mt-0.5">{{ $t('auth.loginAsAnotherUser') || 'Login as another user' }}</div>
              </div>
            </button>
          </div>

          <!-- Quick Role Switch (Strictly SUPERADMIN in Dev/Testing environments only) -->
          <div v-if="isSuperAdmin" class="py-1 border-b border-slate-100 dark:border-slate-800">
            <div class="px-3 py-1 flex items-center justify-between">
              <span class="text-4xs font-mono font-bold uppercase tracking-wider text-purple-600 dark:text-purple-400">
                DEV / TESTING ONLY
              </span>
              <span class="text-4xs font-mono px-1 rounded bg-purple-500/10 text-purple-600 dark:text-purple-400 font-bold border border-purple-500/20">
                SUPERADMIN
              </span>
            </div>

            <!-- Role 1: SUPERADMIN -->
            <button
              @click="quickSwitchAccount('admin', 'Password123!')"
              class="w-full text-left px-3 py-1.5 rounded-xl flex items-center justify-between transition-colors"
              :class="userRole === 'SUPERADMIN' ? 'bg-purple-500/10 text-purple-600 dark:text-purple-400 font-bold' : 'hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300'"
            >
              <div>
                <div class="font-semibold text-xs flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full bg-purple-500"></span>
                  SUPERADMIN
                </div>
                <div class="text-4xs text-slate-400 font-mono">@admin (Full Dev + Raw Access)</div>
              </div>
              <span v-if="userRole === 'SUPERADMIN'" class="text-xs">&check;</span>
            </button>

            <!-- Role 2: ADMIN -->
            <button
              @click="quickSwitchAccount('admin_user', 'Password123!')"
              class="w-full text-left px-3 py-1.5 rounded-xl flex items-center justify-between transition-colors"
              :class="userRole === 'ADMIN' ? 'bg-blue-500/10 text-blue-600 dark:text-blue-400 font-bold' : 'hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300'"
            >
              <div>
                <div class="font-semibold text-xs flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full bg-blue-500"></span>
                  ADMIN
                </div>
                <div class="text-4xs text-slate-400 font-mono">@admin_user (Ops Admin, No Dev/Raw)</div>
              </div>
              <span v-if="userRole === 'ADMIN'" class="text-xs">&check;</span>
            </button>

            <!-- Role 3: ENGINEER -->
            <button
              @click="quickSwitchAccount('engineer_user', 'Password123!')"
              class="w-full text-left px-3 py-1.5 rounded-xl flex items-center justify-between transition-colors"
              :class="userRole === 'ENGINEER' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 font-bold' : 'hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300'"
            >
              <div>
                <div class="font-semibold text-xs flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full bg-amber-500"></span>
                  ENGINEER
                </div>
                <div class="text-4xs text-slate-400 font-mono">@engineer_user (Sensors/Live Comm, No Dev)</div>
              </div>
              <span v-if="userRole === 'ENGINEER'" class="text-xs">&check;</span>
            </button>

            <!-- Role 4: OPERATOR -->
            <button
              @click="quickSwitchAccount('operator_user', 'Password123!')"
              class="w-full text-left px-3 py-1.5 rounded-xl flex items-center justify-between transition-colors"
              :class="userRole === 'OPERATOR' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 font-bold' : 'hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300'"
            >
              <div>
                <div class="font-semibold text-xs flex items-center gap-1.5">
                  <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
                  OPERATOR
                </div>
                <div class="text-4xs text-slate-400 font-mono">@operator_user (Console/Alarms, No Dev/Raw)</div>
              </div>
              <span v-if="userRole === 'OPERATOR'" class="text-xs">&check;</span>
            </button>
          </div>

          <!-- Logout Button -->
          <div class="pt-1">
            <button
              type="button"
              @click="openLogoutConfirmModal"
              class="w-full text-left px-3 py-2 rounded-xl flex items-center gap-2.5 text-rose-600 dark:text-rose-400 hover:bg-rose-500/10 transition-colors font-semibold"
            >
              <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              <span>{{ $t('auth.logout') || 'Logout' }}</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 1. Profile Modal -->
    <div
      v-if="showProfileModal"
      @click.self="showProfileModal = false"
      class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-fade-in"
    >
      <div class="bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 rounded-3xl w-full max-w-md p-6 shadow-2xl space-y-4 m-auto max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800">
          <div class="flex items-center gap-2.5">
            <div class="w-9 h-9 rounded-2xl flex items-center justify-center text-xs font-bold ring-1" :class="roleAvatarClass">
              {{ (user.username || 'U').charAt(0).toUpperCase() }}
            </div>
            <div>
              <h2 class="text-sm font-bold text-slate-900 dark:text-white">{{ $t('auth.profileTitle') || 'User Account Profile' }}</h2>
              <p class="text-3xs text-slate-400">Authenticated RBAC session</p>
            </div>
          </div>
          <button @click="showProfileModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-lg leading-none">&times;</button>
        </div>

        <div class="space-y-2.5 text-xs">
          <div class="flex items-center justify-between p-3 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-100 dark:border-slate-800/80">
            <span class="text-slate-500">{{ $t('auth.fullName') || 'Full Name' }}</span>
            <span class="font-bold text-slate-900 dark:text-white">{{ user.full_name || user.username }}</span>
          </div>
          <div class="flex items-center justify-between p-3 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-100 dark:border-slate-800/80">
            <span class="text-slate-500">{{ $t('auth.username') || 'Username' }}</span>
            <span class="font-mono font-semibold text-slate-900 dark:text-white">@{{ user.username }}</span>
          </div>
          <div class="flex items-center justify-between p-3 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-100 dark:border-slate-800/80">
            <span class="text-slate-500">{{ $t('auth.email') || 'Email' }}</span>
            <span class="font-mono text-slate-700 dark:text-slate-300">{{ user.email || 'N/A' }}</span>
          </div>
          <div class="flex items-center justify-between p-3 rounded-2xl bg-slate-50 dark:bg-slate-900/60 border border-slate-100 dark:border-slate-800/80">
            <span class="text-slate-500">{{ $t('auth.activeRole') || 'Assigned Role' }}</span>
            <span class="px-2.5 py-0.5 rounded-full text-3xs font-mono font-bold" :class="roleBadgeClass">
              {{ userRole }}
            </span>
          </div>
        </div>

        <div class="pt-3 flex items-center justify-end gap-2 border-t border-slate-100 dark:border-slate-800">
          <button
            @click="showProfileModal = false; openSwitchModal();"
            class="px-3.5 py-2 rounded-xl border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-semibold"
          >
            {{ $t('auth.switchAccount') || 'Switch Account' }}
          </button>
          <button
            @click="showProfileModal = false"
            class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold shadow-xs"
          >
            Close
          </button>
        </div>
      </div>
    </div>

    <!-- 2. Switch Account Modal -->
    <div
      v-if="showSwitchModal"
      @click.self="showSwitchModal = false"
      class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-fade-in"
    >
      <div class="bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 rounded-3xl w-full max-w-md p-6 shadow-2xl space-y-4 m-auto max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800">
          <div>
            <h2 class="text-sm font-bold text-slate-900 dark:text-white">{{ $t('auth.switchAccount') || 'Switch Account' }}</h2>
            <p class="text-3xs text-slate-400">{{ $t('auth.switchAccountSubtitle') || 'Enter credentials for target account' }}</p>
          </div>
          <button @click="showSwitchModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-lg leading-none">&times;</button>
        </div>

        <div v-if="switchForm.error" class="p-3 rounded-2xl bg-rose-500/10 border border-rose-500/20 text-rose-500 text-xs font-medium">
          {{ switchForm.error }}
        </div>

        <form @submit.prevent="performSwitch" class="space-y-3 text-xs">
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">{{ $t('auth.username') || 'Username' }}</label>
            <input
              type="text"
              v-model="switchForm.username"
              required
              placeholder="Username"
              class="w-full bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white font-mono focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">{{ $t('auth.password') || 'Password' }}</label>
            <input
              type="password"
              v-model="switchForm.password"
              required
              placeholder="Password"
              class="w-full bg-slate-50 dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white font-mono focus:ring-1 focus:ring-blue-500"
            />
          </div>

          <div class="pt-3 flex items-center justify-end gap-2 border-t border-slate-100 dark:border-slate-800">
            <button
              type="button"
              @click="showSwitchModal = false"
              class="px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-semibold"
            >
              {{ $t('auth.cancel') || 'Cancel' }}
            </button>
            <button
              type="submit"
              :disabled="switchForm.loading"
              class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold shadow-xs disabled:opacity-50 flex items-center gap-1.5"
            >
              <svg v-if="switchForm.loading" class="animate-spin h-3.5 w-3.5 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
              </svg>
              <span>{{ switchForm.loading ? ($t('auth.signingIn') || 'Authenticating...') : ($t('auth.switchAccount') || 'Sign In & Switch') }}</span>
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- 3. Logout Confirm Modal -->
    <div
      v-if="showLogoutConfirmModal"
      @click.self="showLogoutConfirmModal = false"
      class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 backdrop-blur-xs animate-fade-in"
    >
      <div class="bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 rounded-3xl w-full max-w-sm p-6 shadow-2xl space-y-4 text-center m-auto max-h-[90vh] overflow-y-auto">
        <div class="w-12 h-12 rounded-2xl bg-rose-500/10 text-rose-500 flex items-center justify-center mx-auto">
          <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </div>

        <div>
          <h2 class="text-base font-bold text-slate-900 dark:text-white">{{ $t('auth.logoutConfirmTitle') || 'Confirm Session Logout' }}</h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
            {{ $t('auth.logoutConfirmMsg') || 'Are you sure you want to end your active session? You will be redirected to the sign-in screen.' }}
          </p>
        </div>

        <div class="pt-3 flex items-center justify-center gap-2 border-t border-slate-100 dark:border-slate-800">
          <button
            type="button"
            @click="showLogoutConfirmModal = false"
            class="px-4 py-2 rounded-xl border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-semibold"
          >
            {{ $t('auth.cancel') || 'Cancel' }}
          </button>
          <button
            type="button"
            @click="handleLogout"
            class="px-4 py-2 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold shadow-xs"
          >
            {{ $t('auth.confirmLogout') || 'Yes, Logout' }}
          </button>
        </div>
      </div>
    </div>
  </header>
</template>

<script>
export default {
  name: 'Header',
  data() {
    return {
      currentTime: new Date().toLocaleTimeString('en-GB'),
      clockInterval: null,
      isLangMenuOpen: false,
      isThemeMenuOpen: false,
      isUserMenuOpen: false,
      showProfileModal: false,
      showSwitchModal: false,
      showLogoutConfirmModal: false,
      switchForm: {
        username: '',
        password: '',
        error: '',
        loading: false,
      },
    };
  },
  computed: {
    systemStatus() {
      return this.$store.state.systemStatus;
    },
    user() {
      return this.$store.state.user || {};
    },
    userRole() {
      return ((this.user && this.user.role) || 'OPERATOR').toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    roleBadgeClass() {
      if (this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR') {
        return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/30';
      }
      if (this.userRole === 'ADMIN') {
        return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/30';
      }
      if (this.userRole === 'ENGINEER') {
        return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/30';
      }
      return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30';
    },
    roleDotClass() {
      if (this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR') return 'bg-purple-500';
      if (this.userRole === 'ADMIN') return 'bg-blue-500';
      if (this.userRole === 'ENGINEER') return 'bg-amber-500';
      return 'bg-emerald-500';
    },
    roleAvatarClass() {
      if (this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR') {
        return 'bg-purple-600/20 text-purple-600 dark:text-purple-400 ring-purple-500/40';
      }
      if (this.userRole === 'ADMIN') {
        return 'bg-blue-600/20 text-blue-600 dark:text-blue-400 ring-blue-500/40';
      }
      if (this.userRole === 'ENGINEER') {
        return 'bg-amber-600/20 text-amber-600 dark:text-amber-400 ring-amber-500/40';
      }
      return 'bg-emerald-600/20 text-emerald-600 dark:text-emerald-400 ring-emerald-500/40';
    },
    sidebarAutoHide() {
      return !!((this.$store.state.uiPreferences && this.$store.state.uiPreferences.sidebarAutoHide) || this.$store.state.sidebarAutoHide);
    },
    themeSetting() {
      return (this.$store.state.uiPreferences && this.$store.state.uiPreferences.theme) || 'system';
    },
    isDarkMode() {
      const eff = this.$store.state.uiPreferences && this.$store.state.uiPreferences.effectiveTheme;
      if (eff) return eff === 'dark';
      return this.themeSetting === 'dark' || (this.themeSetting === 'system' && typeof window !== 'undefined' && window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches);
    },
    language() {
      return (this.$store.state.uiPreferences && this.$store.state.uiPreferences.language) || 'en';
    },
    currentLanguageLabel() {
      return this.language === 'id' ? 'ID' : 'EN';
    },
    currentThemeLabel() {
      if (this.themeSetting === 'light') return this.$t('header.light');
      if (this.themeSetting === 'dark') return this.$t('header.dark');
      return this.$t('header.system');
    },
    activeAlarmsCount() {
      const active = this.$store.state.activeAlarms || [];
      return active.filter(a => a && a.status === 'ACTIVE').length;
    },
    overallAppStatus() {
      const s = this.systemStatus;
      if (!s) return 'HEALTHY';
      if (s.database_status && typeof s.database_status === 'string' && s.database_status.toLowerCase().includes('fail')) {
        return 'FAILURE';
      }
      if (s.offline_devices > 0 || this.activeAlarmsCount > 0 || (s.critical_alarms && s.critical_alarms > 0) || (this.systemStatus && this.systemStatus.online_devices === 0)) {
        return 'DEGRADED';
      }
      return 'HEALTHY';
    },
    appStatusBadgeClass() {
      if (this.overallAppStatus === 'FAILURE') {
        return 'bg-rose-500/20 text-rose-500 dark:text-rose-400 border-rose-500/40';
      }
      if (this.overallAppStatus === 'DEGRADED') {
        return 'bg-amber-500/20 text-amber-600 dark:text-amber-400 border-amber-500/40';
      }
      return 'bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 border-emerald-500/40';
    },
  },
  mounted() {
    this.clockInterval = setInterval(() => {
      this.currentTime = new Date().toLocaleTimeString('en-GB');
    }, 1000);
    document.addEventListener('click', this.handleOutsideClick);
  },
  beforeDestroy() {
    if (this.clockInterval) {
      clearInterval(this.clockInterval);
    }
    document.removeEventListener('click', this.handleOutsideClick);
  },
  methods: {
    toggleSidebarAutoHide() {
      this.$store.dispatch('toggleSidebarAutoHide');
      this.$emit('toggle-autohide', this.sidebarAutoHide);
    },
    toggleLanguage() {
      const next = this.language === 'id' ? 'en' : 'id';
      this.$store.dispatch('setLanguage', next);
    },
    toggleTheme() {
      const next = this.isDarkMode ? 'light' : 'dark';
      this.$store.dispatch('setTheme', next);
    },
    selectLanguage(lang) {
      this.$store.dispatch('setLanguage', lang);
      this.isLangMenuOpen = false;
    },
    selectTheme(theme) {
      this.$store.dispatch('setTheme', theme);
      this.isThemeMenuOpen = false;
    },
    openProfileModal() {
      this.isUserMenuOpen = false;
      this.showProfileModal = true;
    },
    openSwitchModal() {
      this.isUserMenuOpen = false;
      this.switchForm = { username: '', password: '', error: '', loading: false };
      this.showSwitchModal = true;
    },
    openLogoutConfirmModal() {
      this.isUserMenuOpen = false;
      this.showLogoutConfirmModal = true;
    },
    async handleLogout() {
      this.showLogoutConfirmModal = false;
      this.isUserMenuOpen = false;
      await this.$store.dispatch('logout');
      if (this.$route.path !== '/login') {
        this.$router.push('/login');
      }
    },
    async performSwitch() {
      this.switchForm.loading = true;
      this.switchForm.error = '';
      try {
        await this.$store.dispatch('login', {
          username: this.switchForm.username.trim(),
          password: this.switchForm.password,
        });
        this.showSwitchModal = false;
        this.redirectAfterRoleChange();
      } catch (err) {
        console.error('Switch error:', err);
        this.switchForm.error = err.response?.data?.error || err.message || 'Failed to switch account';
      } finally {
        this.switchForm.loading = false;
      }
    },
    async quickSwitchAccount(username, password) {
      this.isUserMenuOpen = false;
      try {
        await this.$store.dispatch('switchAccount', {
          target_username: username,
          username,
          password,
        });
        this.redirectAfterRoleChange();
      } catch (err) {
        console.error('Failed to switch account:', err);
        alert('Failed to switch user account: ' + (err.response?.data?.error || err.message || 'Error'));
      }
    },
    redirectAfterRoleChange() {
      const meta = this.$route.meta || {};
      const roles = meta.roles || [];
      const currentRole = (this.$store.state.user?.role || '').toUpperCase();
      const isSuper = currentRole === 'SUPERADMIN' || currentRole === 'ADMINISTRATOR';

      let allowed = true;
      if (roles.length > 0) {
        allowed = roles.some(r => {
          const target = r.toUpperCase();
          if ((target === 'SUPERADMIN' || target === 'ADMINISTRATOR') && isSuper) return true;
          return target === currentRole;
        });
      }

      if (!allowed) {
        if (isSuper) {
          this.$router.push('/dashboard');
        } else {
          this.$router.push('/monitoring/operator');
        }
      } else {
        // Re-trigger router refresh
        this.$router.replace({ path: this.$route.path, query: { ...this.$route.query, _r: Date.now() } }).catch(() => {});
      }
    },
    handleOutsideClick(event) {
      if (this.$refs.userDropdown && !this.$refs.userDropdown.contains(event.target)) {
        this.isUserMenuOpen = false;
      }
    },
  },
};
</script>
