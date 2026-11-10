<template>
  <div
    class="bento-dashboard-page select-none"
    :class="[
      isDark ? 'is-night dark-mode' : 'is-day light-mode',
      isCompactScreen ? 'is-compact-vp' : 'is-wide-vp',
      { 'no-perf-animations': perfSettings && !perfSettings.chartAnimationsEnabled }
    ]"
  >
    <!-- =================================================================== -->
    <!-- 1. UNIFIED CALM HEADER + ICON-ONLY OPERATOR CONSOLE TOOLBAR         -->
    <!-- =================================================================== -->
    <header class="header-shell">
      <div class="h-row1" :class="{ 'is-compact-row1': isCompactScreen }">
        <!-- Area 1 (Left): Customer Company Branding (Configured Values Only — No Default Text) -->
        <div
          class="brand-left cursor-pointer"
          @click="openBrandingModal('customer')"
          :title="(customerCompanyName ? customerCompanyName + ' · ' : '') + 'Klik untuk konfigurasi branding'"
        >
          <div v-if="customerLogoUrl" class="logo-img-wrap">
            <img
              :src="customerLogoUrl"
              :alt="customerCompanyName || 'Customer Logo'"
              class="h-6 max-w-[104px] w-auto object-contain rounded"
            />
          </div>
          <div v-else class="logo-badge">
            <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9" />
              <path d="M7.8 16.2c-2.3-2.3-2.3-6.1 0-8.5" />
              <circle cx="12" cy="12" r="2" />
              <path d="M16.2 7.8c2.3 2.3 2.3 6.1 0-8.5" />
              <path d="M19.1 4.9C23 8.8 23 15.1 19.1 19" />
            </svg>
          </div>

          <div
            v-if="customerCompanyName || stationTypeLabel || headerSecondaryLocationLabel"
            class="min-w-0 flex-1"
          >
            <div v-if="customerCompanyName" class="station-name">
              <span class="truncate">{{ customerCompanyName }}</span>
            </div>
            <div
              v-if="stationTypeLabel || headerSecondaryLocationLabel"
              class="station-loc truncate"
            >
              <span v-if="stationTypeLabel">Stasiun <strong>{{ stationTypeLabel }}</strong></span>
              <span v-if="headerSecondaryLocationLabel">
                <template v-if="stationTypeLabel"> · </template>{{ headerSecondaryLocationLabel }}
              </span>
            </div>
          </div>
        </div>

        <!-- Area 2 (Center): Calm Digital Clock -->
        <div v-if="currentTime" class="clock-center" :class="{ 'is-wide-center': !isCompactScreen }">
          <div class="clock-pill">
            <svg width="12.5" height="12.5" viewBox="0 0 24 24" fill="none" :stroke="isDark ? '#38BDF8' : '#0284C7'" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="10" />
              <polyline points="12 6 12 12 16 14" />
            </svg>
            <span>
              <template v-if="currentDayDate && bentoViewportWidth >= 920">{{ currentDayDate }} · </template>
              <strong>{{ currentTime }}</strong>
            </span>
          </div>
        </div>

        <!-- Area 3 (Right): Single Clean Maker Branding (Configured Values Only — No Default Text) -->
        <div
          class="brand-right cursor-pointer"
          @click="openBrandingModal('maker')"
          :title="(makerCompanyName ? makerCompanyName + ' · ' : '') + 'Klik untuk konfigurasi branding'"
        >
          <div v-if="makerCompanyName || makerSubtitle" class="maker-info hidden lg:block">
            <div v-if="makerCompanyName" class="maker-title truncate" :class="isCompactScreen ? 'max-w-[165px]' : 'max-w-[215px]'">
              {{ makerCompanyName }}
            </div>
            <div v-if="makerSubtitle" class="maker-sub truncate" :class="isCompactScreen ? 'max-w-[165px]' : 'max-w-[215px]'">
              {{ makerSubtitle }}
            </div>
          </div>

          <div v-if="makerLogoUrl" class="logo-img-wrap">
            <img
              :src="makerLogoUrl"
              :alt="makerCompanyName || 'Maker Logo'"
              class="h-6 max-w-[104px] w-auto object-contain rounded"
            />
          </div>
          <div v-else class="maker-logo">
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
            </svg>
          </div>
        </div>
      </div>

      <div class="h-divider"></div>

      <!-- Row 2: 100% Icon-Only Toolbar Matching Operator Console -->
      <div class="h-row2" ref="toolbarContainer">
        <!-- Left Icon Cluster: Sidebar Auto-Hide, Fullscreen, View/Filter, Performance, ISPU Settings, Day/Night -->
        <div class="tool-cluster flex-wrap">
          <!-- 1. Sidebar Auto-Hide (Icon Only) -->
          <button
            type="button"
            id="btn-operator-sidebar-autohide"
            @click.stop="toggleSidebarAutoHide"
            class="op-icon-btn"
            :class="{ 'is-active': sidebarAutoHide }"
            :title="sidebarAutoHide ? ($t('header.sidebarAutoHideOn') || 'Sidebar: Auto-Hide Aktif (Klik untuk Pin)') : ($t('header.sidebarAutoHideOff') || 'Sidebar: Dipin (Klik untuk Auto-Hide)')"
            aria-label="Toggle Sidebar Auto-Hide"
          >
            <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="2" />
              <path d="M9 3v18" />
              <path v-if="!sidebarAutoHide" d="m14 9-3 3 3 3" stroke-linecap="round" stroke-linejoin="round" />
              <path v-else d="m12 9 3 3-3 3" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>

          <!-- 2. Fullscreen / Focus Layout Toggle (Icon Only) -->
          <button
            type="button"
            @click="$emit('toggle-sidebar')"
            class="op-icon-btn text-sky-600 dark:text-sky-400"
            :class="{ 'is-active': isSidebarCollapsed }"
            :title="isSidebarCollapsed ? ($t('operator.exitFullWidth') || 'Keluar Mode Fokus Layar Penuh') : ($t('operator.fullWidth') || 'Mode Fokus Layar Penuh')"
            aria-label="Toggle Full Width"
          >
            <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path v-if="!isSidebarCollapsed" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
              <path v-else stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 4H5a1 1 0 00-1 1v4m0 0l5-5m6 0h4a1 1 0 011 1v4m0 0l-5-5M9 20H5a1 1 0 01-1-1v-4m0 0l5 5m6 0h4a1 1 0 001-1v-4m0 0l-5 5" />
            </svg>
          </button>

          <div class="op-v-div"></div>

          <!-- 3. View & Filter Options Popover (Icon Only + Corner Badge) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              id="btn-toolbar-view-options"
              @click.stop="toggleToolbarPopover('viewOptions')"
              class="op-icon-btn"
              :class="{
                'is-active': activeToolbarPopover === 'viewOptions',
                'is-amber': activeToolbarPopover !== 'viewOptions' && filterPinnedOnly,
                'is-cyan-soft': activeToolbarPopover !== 'viewOptions' && !filterPinnedOnly && selectedCategoryFilter !== 'all'
              }"
              :title="$t('operator.toolbarViewOptionsTooltip') || 'Opsi Tampilan & Filter (Kartu, Tabel, Kolom, Favorit, Kategori)'"
              aria-label="View & Filter Options"
            >
              <OperatorDashboardIcon name="sliders" :size="16" />
              <span
                v-if="filterPinnedOnly"
                class="op-corner-pill tone-amber"
                title="Filter favorit aktif"
              >★</span>
              <span
                v-else-if="selectedCategoryFilter !== 'all'"
                class="op-corner-dot dot-sky"
                title="Filter kategori aktif"
              ></span>
              <span
                v-else
                class="op-corner-pill tone-Density"
              >{{ viewMode === 'table' ? bentoTableModeShortBadge : bentoCardModeShortBadge }}</span>
            </button>
            <div
              v-if="activeToolbarPopover === 'viewOptions'"
              class="absolute left-0 top-full mt-2 z-50 w-[355px] sm:w-[376px] max-w-[calc(100vw-24px)] rounded-2xl shadow-xl p-4 text-left border border-slate-200/90 dark:border-slate-800/80 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-sans"
              @click.stop
            >
              <!-- Panel Header -->
              <div class="flex items-center justify-between pb-3 mb-3 border-b border-slate-100 dark:border-slate-800/80">
                <div class="flex items-center space-x-2.5">
                  <div class="w-8 h-8 rounded-xl flex items-center justify-center bg-[#EAF5FF] dark:bg-blue-600/15 border border-[#B8D4EA]/80 dark:border-blue-500/30 text-[#1685E5] dark:text-blue-400 flex-shrink-0">
                    <OperatorDashboardIcon name="sliders" :size="16" />
                  </div>
                  <div>
                    <span class="text-[13px] font-bold tracking-tight text-slate-900 dark:text-white block leading-tight">
                      {{ $t('operator.viewAndFilterOptions') || 'View & Filter Options' }}
                    </span>
                    <span class="text-3xs font-medium text-slate-500 dark:text-slate-400 block leading-snug mt-0.5">
                      Display mode, grid density &amp; parameter categories
                    </span>
                  </div>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- Body with tidy sections -->
              <div class="max-h-[calc(85vh-110px)] overflow-y-auto custom-scroller pr-0.5 space-y-3.5">
                <!-- 1. Display Mode Section -->
                <div>
                  <div class="text-3xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5 flex items-center justify-between">
                    <span>{{ $t('operator.displayMode') || 'Display Mode' }}</span>
                    <span class="text-3xs text-[#1685E5] dark:text-blue-400 font-semibold normal-case tracking-normal">{{ viewMode === 'cards' ? (bentoCardGroupByCategory ? 'Cards · Per Kategori' : 'Cards · Gabung') : 'Table' }}</span>
                  </div>
                  <div class="grid grid-cols-3 gap-1.5">
                    <button
                      type="button"
                      @click="setBentoDisplayMode('cards')"
                      class="px-2.5 py-2 rounded-lg border text-left transition-colors flex items-center space-x-1.5"
                      :class="viewMode === 'cards' && !bentoCardGroupByCategory
                        ? 'bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA] dark:border-blue-500/40 shadow-xs font-semibold'
                        : 'bg-slate-50/80 dark:bg-slate-900/60 border-slate-200/80 dark:border-slate-800 text-slate-600 dark:text-slate-300 font-medium hover:text-[#1685E5] dark:hover:text-slate-100 hover:bg-[#EAF5FF]/60 dark:hover:bg-slate-800/50'"
                    >
                      <OperatorDashboardIcon name="layout" :size="14" class="flex-shrink-0" :class="viewMode === 'cards' && !bentoCardGroupByCategory ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'" />
                      <span class="text-xs truncate flex-1">Kartu</span>
                      <span v-if="viewMode === 'cards' && !bentoCardGroupByCategory" class="text-[#1685E5] dark:text-blue-400 font-bold text-xs">✓</span>
                    </button>
                    <button
                      type="button"
                      @click="setBentoDisplayMode('cat-cards')"
                      class="px-2.5 py-2 rounded-lg border text-left transition-colors flex items-center space-x-1.5"
                      :class="viewMode === 'cards' && bentoCardGroupByCategory
                        ? 'bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA] dark:border-blue-500/40 shadow-xs font-semibold'
                        : 'bg-slate-50/80 dark:bg-slate-900/60 border-slate-200/80 dark:border-slate-800 text-slate-600 dark:text-slate-300 font-medium hover:text-[#1685E5] dark:hover:text-slate-100 hover:bg-[#EAF5FF]/60 dark:hover:bg-slate-800/50'"
                      title="Tampilkan kartu parameter terpisah sesuai kategori"
                    >
                      <OperatorDashboardIcon name="layout" :size="14" class="flex-shrink-0" :class="viewMode === 'cards' && bentoCardGroupByCategory ? 'text-[#1685E5] dark:text-blue-400' : 'text-emerald-500'" />
                      <span class="text-xs truncate flex-1">Kartu Kat.</span>
                      <span v-if="viewMode === 'cards' && bentoCardGroupByCategory" class="text-[#1685E5] dark:text-blue-400 font-bold text-xs">✓</span>
                    </button>
                    <button
                      type="button"
                      @click="viewMode = 'table'"
                      class="px-2.5 py-2 rounded-lg border text-left transition-colors flex items-center space-x-1.5"
                      :class="viewMode === 'table'
                        ? 'bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA] dark:border-blue-500/40 shadow-xs font-semibold'
                        : 'bg-slate-50/80 dark:bg-slate-900/60 border-slate-200/80 dark:border-slate-800 text-slate-600 dark:text-slate-300 font-medium hover:text-[#1685E5] dark:hover:text-slate-100 hover:bg-[#EAF5FF]/60 dark:hover:bg-slate-800/50'"
                    >
                      <OperatorDashboardIcon name="menu" :size="14" class="flex-shrink-0" :class="viewMode === 'table' ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'" />
                      <span class="text-xs truncate flex-1">Tabel</span>
                      <span v-if="viewMode === 'table'" class="text-[#1685E5] dark:text-blue-400 font-bold text-xs">✓</span>
                    </button>
                  </div>
                </div>

                <!-- 2. Panel Layout Density (1P - 4P Columns) -->
                <div>
                  <div class="text-3xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5 flex items-center justify-between">
                    <span>{{ $t('operator.panelLayout') || 'Panel Layout' }}</span>
                    <span class="text-3xs font-semibold normal-case tracking-normal" :class="viewMode === 'table' ? 'text-slate-400 italic' : 'text-[#1685E5] dark:text-blue-400'">
                      {{ viewMode === 'table' ? '(N/A in Table)' : gridColumns + 'P Columns' }}
                    </span>
                  </div>
                  <div
                    class="grid grid-cols-4 gap-1.5 transition-opacity"
                    :class="viewMode === 'table' ? 'opacity-40 pointer-events-none' : ''"
                  >
                    <button
                      type="button"
                      v-for="cnt in [1, 2, 3, 4]"
                      :key="'occ-panel-btn-' + cnt"
                      :id="'btn-toolbar-panel-' + cnt"
                      @click="setGridColumns(cnt)"
                      class="py-2 px-1.5 rounded-lg border text-center transition-colors flex flex-col items-center justify-center"
                      :class="gridColumns === cnt
                        ? 'bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA] dark:border-blue-500/40 shadow-xs font-semibold'
                        : 'bg-slate-50/80 dark:bg-slate-900/60 border-slate-200/80 dark:border-slate-800 text-slate-600 dark:text-slate-300 font-medium hover:text-[#1685E5] dark:hover:text-slate-100 hover:bg-[#EAF5FF]/60 dark:hover:bg-slate-800/50'"
                      :title="cnt + (cnt === 1 ? ' Column Grid (1P)' : ' Columns Grid (' + cnt + 'P)')"
                    >
                      <span class="text-xs font-bold tabular-nums leading-tight">{{ cnt }}P</span>
                      <span class="text-3xs opacity-80 leading-tight mt-0.5">{{ cnt === 1 ? '1 Col' : cnt + ' Cols' }}</span>
                    </button>
                  </div>
                </div>

                <!-- 3. Pinned Favorites Section -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center space-x-2">
                      <span class="w-6 h-6 rounded-lg bg-amber-50 dark:bg-amber-500/10 border border-amber-200/80 dark:border-amber-500/30 flex items-center justify-center text-amber-500 text-xs flex-shrink-0">★</span>
                      <div>
                        <div class="text-xs font-semibold text-slate-900 dark:text-white leading-tight">
                          {{ $t('operator.pinnedOnly') || 'Pinned Only' }}
                        </div>
                        <div class="text-3xs text-slate-500 dark:text-slate-400 leading-snug mt-0.5">
                          Filter CBI cards to favorites
                        </div>
                      </div>
                    </div>
                    <button
                      type="button"
                      @click="toggleFilterPinned"
                      class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                      :class="filterPinnedOnly ? 'bg-amber-500' : 'bg-slate-200 dark:bg-slate-700'"
                      role="switch"
                      :aria-checked="filterPinnedOnly"
                    >
                      <span
                        aria-hidden="true"
                        class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
                        :class="filterPinnedOnly ? 'translate-x-4' : 'translate-x-0'"
                      ></span>
                    </button>
                  </div>

                  <!-- Pinned items list if any -->
                  <div v-if="pinnedParametersList.length > 0" class="mt-2.5 pt-2.5 border-t border-slate-200/70 dark:border-slate-800">
                    <div class="flex items-center justify-between text-3xs font-medium text-slate-500 dark:text-slate-400 mb-1.5">
                      <span>Favorites ({{ pinnedParametersList.length }})</span>
                      <button
                        type="button"
                        @click="clearAllPinned"
                        class="text-rose-500 hover:text-rose-600 dark:hover:text-rose-400 font-semibold hover:underline"
                      >
                        {{ $t('operator.clearAllPinned') || 'Clear All' }}
                      </button>
                    </div>
                    <div class="max-h-28 overflow-y-auto custom-scroller space-y-1">
                      <div
                        v-for="param in pinnedParametersList"
                        :key="'pinned-item-' + param.param_id"
                        class="flex items-center justify-between px-2 py-1.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800"
                      >
                        <div class="min-w-0 flex-1 pr-1.5">
                          <div class="font-semibold text-slate-800 dark:text-slate-200 truncate text-2xs">
                            {{ param.param_name || param.param_code }}
                          </div>
                          <div class="text-3xs text-slate-500 dark:text-slate-400 flex items-center space-x-1 truncate">
                            <span v-if="!isOperator">{{ param.device_code }}</span>
                            <span v-if="param.unit">· {{ param.unit }}</span>
                          </div>
                        </div>
                        <button
                          type="button"
                          @click.stop="togglePin(param.param_id)"
                          class="p-1 rounded-md text-amber-400 hover:text-rose-500 transition hover:bg-slate-100 dark:hover:bg-slate-800 flex-shrink-0"
                          title="Unpin parameter"
                        >
                          <OperatorDashboardIcon name="pin" :size="13" class="text-amber-400" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 4. Parameter Categories Section -->
                <div>
                  <div class="text-3xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 mb-1.5 flex items-center justify-between">
                    <div class="flex items-center space-x-1.5">
                      <OperatorDashboardIcon name="filter" :size="12" class="text-[#1685E5] dark:text-blue-400" />
                      <span>{{ $t('operator.categories') || 'Parameter Categories' }}</span>
                    </div>
                    <span v-if="selectedCategoryFilter !== 'all'" class="text-3xs text-[#1685E5] dark:text-blue-400 font-semibold normal-case tracking-normal">Filtered</span>
                  </div>

                  <!-- Categories List -->
                  <div class="max-h-52 overflow-y-auto custom-scroller space-y-1 text-xs">
                    <!-- All Categories -->
                    <button
                      type="button"
                      @click="selectedCategoryFilter = 'all'"
                      class="w-full flex items-center justify-between px-2.5 py-2 rounded-lg transition-colors border text-left"
                      :class="selectedCategoryFilter === 'all'
                        ? 'bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA]/80 dark:border-blue-500/40 shadow-xs font-semibold'
                        : 'border-transparent text-slate-600 dark:text-slate-300 font-medium hover:bg-[#EAF5FF]/60 dark:hover:bg-slate-800/40 hover:text-[#1685E5] dark:hover:text-slate-100'"
                    >
                      <div class="flex items-center space-x-2.5 truncate">
                        <OperatorDashboardIcon name="menu" :size="14" class="flex-shrink-0" :class="selectedCategoryFilter === 'all' ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'" />
                        <span class="truncate">{{ $t('operator.allCategories') || 'All Categories' }}</span>
                      </div>
                      <div class="flex items-center space-x-1.5 flex-shrink-0">
                        <span
                          class="px-2 py-0.5 rounded-full text-3xs font-semibold tabular-nums"
                          :class="selectedCategoryFilter === 'all' ? 'bg-white/90 dark:bg-blue-500/20 text-[#1685E5] dark:text-blue-300' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400'"
                        >
                          {{ allParameters.length }}
                        </span>
                        <span v-if="selectedCategoryFilter === 'all'" class="text-[#1685E5] dark:text-blue-400 font-bold text-xs">✓</span>
                      </div>
                    </button>

                    <!-- Not Uncategorized (Categorized Only - Excludes uncategorized parameters) -->
                    <button
                      type="button"
                      id="btn-filter-not-uncategorized"
                      @click="selectedCategoryFilter = 'not_uncategorized'"
                      class="w-full flex items-center justify-between px-2.5 py-2 rounded-lg transition-colors border text-left group"
                      :class="selectedCategoryFilter === 'not_uncategorized'
                        ? 'bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-200 dark:border-emerald-500/40 shadow-xs font-semibold'
                        : 'border-transparent text-slate-600 dark:text-slate-300 font-medium hover:bg-[#EAF5FF]/60 dark:hover:bg-slate-800/40 hover:text-[#1685E5] dark:hover:text-slate-100'"
                      :title="$t('operator.notUncategorizedDesc') || 'Menampilkan semua parameter di luar Tanpa Kategori'"
                    >
                      <div class="flex items-center space-x-2.5 min-w-0 pr-2">
                        <OperatorDashboardIcon name="filter" :size="14" class="text-emerald-500 flex-shrink-0" />
                        <div class="min-w-0">
                          <span class="truncate block leading-snug" :class="selectedCategoryFilter === 'not_uncategorized' ? 'font-semibold' : 'font-medium'">
                            {{ $t('operator.notUncategorized') || 'Not Uncategorized' }}
                          </span>
                          <span class="text-3xs text-slate-500 dark:text-slate-400 block font-normal leading-snug truncate">
                            {{ $t('operator.notUncategorizedDesc') || 'Di luar uncategorized' }}
                          </span>
                        </div>
                      </div>
                      <div class="flex items-center space-x-1.5 flex-shrink-0">
                        <span
                          class="px-2 py-0.5 rounded-full text-3xs font-semibold tabular-nums"
                          :class="selectedCategoryFilter === 'not_uncategorized' ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400'"
                        >
                          {{ categorizedParamCount }}
                        </span>
                        <span v-if="selectedCategoryFilter === 'not_uncategorized'" class="text-emerald-600 dark:text-emerald-400 font-bold text-xs">✓</span>
                      </div>
                    </button>

                    <!-- Dynamic Category Tabs -->
                    <button
                      type="button"
                      v-for="cat in activeCategoriesList"
                      :key="cat.id"
                      @click="selectedCategoryFilter = cat.id"
                      class="w-full flex items-center justify-between px-2.5 py-2 rounded-lg transition-colors border text-left"
                      :class="selectedCategoryFilter === cat.id
                        ? getCategoryTheme(cat.color_key).activeTab + ' font-semibold'
                        : 'border-transparent text-slate-600 dark:text-slate-300 font-medium hover:bg-[#EAF5FF]/60 dark:hover:bg-slate-800/40 hover:text-[#1685E5] dark:hover:text-slate-100'"
                    >
                      <div class="flex items-center space-x-2.5 truncate pr-2">
                        <span class="w-2 h-2 rounded-full flex-shrink-0" :class="getCategoryTheme(cat.color_key).dot"></span>
                        <span class="truncate leading-snug">{{ cat.name }}</span>
                      </div>
                      <div class="flex items-center space-x-1.5 flex-shrink-0">
                        <span
                          class="px-2 py-0.5 rounded-full text-3xs font-semibold tabular-nums"
                          :class="selectedCategoryFilter === cat.id ? getCategoryTheme(cat.color_key).pill : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400'"
                        >
                          {{ getCategoryParamCount(cat.id) }}
                        </span>
                        <span v-if="selectedCategoryFilter === cat.id" class="text-[#1685E5] dark:text-blue-400 font-bold text-xs">✓</span>
                      </div>
                    </button>

                    <!-- Uncategorized Filter (All Parameters without category) -->
                    <button
                      type="button"
                      id="btn-filter-uncategorized"
                      @click="selectedCategoryFilter = 'uncategorized'"
                      class="w-full flex items-center justify-between px-2.5 py-2 rounded-lg transition-colors border text-left group"
                      :class="selectedCategoryFilter === 'uncategorized'
                        ? 'bg-[#EAF5FF] dark:bg-slate-800/80 text-[#1685E5] dark:text-white border-[#B8D4EA] dark:border-slate-600 shadow-xs font-semibold'
                        : 'border-transparent text-slate-600 dark:text-slate-300 font-medium hover:bg-[#EAF5FF]/60 dark:hover:bg-slate-800/40 hover:text-[#1685E5] dark:hover:text-slate-100'"
                      :title="$t('operator.uncategorizedDesc') || 'Semua parameter yang tidak masuk ke category'"
                    >
                      <div class="flex items-center space-x-2.5 min-w-0 pr-2">
                        <span class="w-2 h-2 rounded-full flex-shrink-0 bg-slate-400 dark:bg-slate-400 ring-2 ring-slate-300/60 dark:ring-slate-600/60"></span>
                        <div class="min-w-0">
                          <span class="truncate block leading-snug" :class="selectedCategoryFilter === 'uncategorized' ? 'font-semibold' : 'font-medium'">
                            {{ $t('operator.uncategorized') || 'Tanpa Kategori' }}
                          </span>
                          <span class="text-3xs text-slate-500 dark:text-slate-400 block font-normal leading-snug truncate">
                            {{ $t('operator.uncategorizedDesc') || 'Tidak masuk ke category' }}
                          </span>
                        </div>
                      </div>
                      <div class="flex items-center space-x-1.5 flex-shrink-0">
                        <span
                          class="px-2 py-0.5 rounded-full text-3xs font-semibold tabular-nums"
                          :class="selectedCategoryFilter === 'uncategorized' ? 'bg-white/90 dark:bg-slate-700 text-[#1685E5] dark:text-white' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400'"
                        >
                          {{ uncategorizedParamCount }}
                        </span>
                        <span v-if="selectedCategoryFilter === 'uncategorized'" class="text-[#1685E5] dark:text-blue-400 font-bold text-xs">✓</span>
                      </div>
                    </button>
                  </div>

                  <!-- Footer: Category Management Modal Trigger -->
                  <div v-if="canManageCategories" class="pt-2.5 mt-2.5 border-t border-slate-100 dark:border-slate-800">
                    <button
                      type="button"
                      @click="openCategoryManagementModal(); closeToolbarPopover()"
                      class="w-full py-2 px-3 rounded-xl text-xs font-semibold transition-colors border flex items-center justify-center space-x-1.5 bg-[#EAF5FF]/70 dark:bg-slate-900/70 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA]/80 dark:border-slate-700 hover:bg-[#EAF5FF] dark:hover:bg-blue-600/15"
                    >
                      <span>🏷️</span>
                      <span>{{ $t('operator.manageCategories') || 'Configure Categories & Layout' }}</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 4. Performance Settings Popover (Icon Only + Corner Dot) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              id="btn-operator-performance-settings"
              @click.stop="toggleToolbarPopover('performanceSettings')"
              class="op-icon-btn"
              :class="{
                'is-active': activeToolbarPopover === 'performanceSettings',
                'is-amber': activeToolbarPopover !== 'performanceSettings' && isPerformanceOptimizedActive
              }"
              title="Pengaturan Performa Dashboard (Animasi, Sparkline, Mode Hemat)"
              aria-label="Performance Settings"
            >
              <OperatorDashboardIcon name="performance" :size="16" />
              <span v-if="isPerformanceOptimizedActive" class="op-corner-dot dot-amber"></span>
            </button>
            <div
              v-if="activeToolbarPopover === 'performanceSettings'"
              class="absolute left-0 top-full mt-2 z-50 w-[420px] sm:w-[452px] max-w-[calc(100vw-24px)] rounded-2xl shadow-xl p-4 text-left border border-slate-200/90 dark:border-slate-800/80 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-sans max-h-[calc(100vh-160px)] overflow-y-auto custom-scroller"
              :style="{ maxHeight: bentoToolbarPopoverMaxHeight }"
              @click.stop
            >
              <!-- Panel Header -->
              <div class="flex items-center justify-between pb-3 mb-3 border-b border-slate-100 dark:border-slate-800/80">
                <div class="flex items-center space-x-2.5">
                  <div class="w-8 h-8 rounded-xl flex items-center justify-center bg-[#EAF5FF] dark:bg-blue-600/15 border border-[#B8D4EA]/80 dark:border-blue-500/30 text-[#1685E5] dark:text-blue-400 flex-shrink-0">
                    <OperatorDashboardIcon name="performance" :size="16" />
                  </div>
                  <div>
                    <div class="flex items-center space-x-2">
                      <span class="text-[13px] font-bold tracking-tight text-slate-900 dark:text-white leading-tight">
                        {{ $t('operator.perfSettingsTitle') || 'Performance Settings' }}
                      </span>
                      <span
                        v-if="isPerformanceOptimizedActive"
                        class="px-1.5 py-0.5 rounded-md text-3xs font-semibold bg-amber-50 dark:bg-amber-500/15 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-500/35 leading-none flex items-center space-x-1"
                      >
                        <span class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse"></span>
                        <span>OPTIMIZED</span>
                      </span>
                      <span
                        v-else
                        class="px-1.5 py-0.5 rounded-md text-3xs font-semibold bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200/80 dark:border-slate-700 leading-none"
                      >
                        STANDARD
                      </span>
                    </div>
                    <div class="text-3xs text-slate-500 dark:text-slate-400 mt-0.5 leading-snug font-medium">
                      {{ $t('operator.perfSettingsSubtitle') || 'Raspberry Pi & Low-Power Chromium Optimization' }}
                    </div>
                  </div>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- 1-Tap Quick Optimization Profiles -->
              <div class="mb-2.5 p-2.5 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800">
                <div class="flex items-center justify-between mb-1.5 px-0.5 text-3xs uppercase tracking-wider font-semibold text-slate-400 dark:text-slate-500">
                  <span>{{ $t('operator.perfPresetQuick') || 'Quick Presets' }}</span>
                  <span class="text-[#1685E5] dark:text-blue-400 font-medium normal-case tracking-normal">Tap to apply</span>
                </div>
                <div class="grid grid-cols-3 gap-1.5">
                  <button
                    type="button"
                    @click="setPerfPreset('standard')"
                    class="py-1.5 px-2 rounded-lg text-2xs font-semibold border transition-colors text-center cursor-pointer flex items-center justify-center space-x-1"
                    :class="currentPerfPreset === 'standard'
                      ? 'bg-[#1685E5] text-white btn-text border-[#1685E5] shadow-sm'
                      : 'bg-white dark:bg-[#0B0F19] text-slate-600 dark:text-slate-300 border-slate-200/90 dark:border-slate-800 hover:border-[#B8D4EA] hover:text-[#1685E5]'"
                    title="All features active for standard desktop workstations"
                  >
                    <span>⚡</span>
                    <span>{{ $t('operator.perfPresetStandard') || 'Standard' }}</span>
                  </button>
                  <button
                    type="button"
                    @click="setPerfPreset('balanced')"
                    class="py-1.5 px-2 rounded-lg text-2xs font-semibold border transition-colors text-center cursor-pointer flex items-center justify-center space-x-1"
                    :class="currentPerfPreset === 'balanced'
                      ? 'bg-[#1685E5] text-white btn-text border-[#1685E5] shadow-sm'
                      : 'bg-white dark:bg-[#0B0F19] text-slate-600 dark:text-slate-300 border-slate-200/90 dark:border-slate-800 hover:border-[#B8D4EA] hover:text-[#1685E5]'"
                    title="Balanced profile for Raspberry Pi 4 (sparklines off, full trend active)"
                  >
                    <span>⚖️</span>
                    <span>{{ $t('operator.perfPresetBalanced') || 'Balanced' }}</span>
                  </button>
                  <button
                    type="button"
                    @click="setPerfPreset('max_eco')"
                    class="py-1.5 px-2 rounded-lg text-2xs font-semibold border transition-colors text-center cursor-pointer flex items-center justify-center space-x-1"
                    :class="currentPerfPreset === 'max_eco'
                      ? 'bg-amber-500 text-white btn-text border-amber-500 shadow-sm'
                      : 'bg-white dark:bg-[#0B0F19] text-slate-600 dark:text-slate-300 border-slate-200/90 dark:border-slate-800 hover:border-amber-300 hover:text-amber-600'"
                    title="Maximum CPU savings for Raspberry Pi 3 or slow networks"
                  >
                    <span>🚀</span>
                    <span>{{ $t('operator.perfPresetMaxEco') || 'Max Eco' }}</span>
                  </button>
                </div>
              </div>

              <!-- Storage Badge Notice & Reset -->
              <div class="mb-3 px-3 py-2 rounded-xl bg-slate-50/70 dark:bg-slate-900/50 border border-slate-200/70 dark:border-slate-800 flex items-center justify-between text-3xs text-slate-500 dark:text-slate-400">
                <span class="flex items-center space-x-1.5 font-medium">
                  <svg class="w-3.5 h-3.5 text-[#1685E5] dark:text-blue-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2 1.5 3 3.5 3h9c2 0 3.5-1 3.5-3V7M4 7c0-2 1.5-3 3.5-3h9c2 0 3.5 1 3.5 3M4 7h16m-8 4v4m-2-2h4" />
                  </svg>
                  <span>{{ $t('operator.perfStorageNotice') || 'Saved in this browser only (localStorage)' }}</span>
                </span>
                <button
                  type="button"
                  @click="resetPerformanceSettings"
                  class="text-[#1685E5] dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 font-semibold text-3xs flex items-center space-x-1 py-1 px-2 rounded-lg hover:bg-[#EAF5FF] dark:hover:bg-blue-600/15 transition-colors cursor-pointer"
                >
                  <OperatorDashboardIcon name="reset" :size="11" />
                  <span>{{ $t('operator.perfResetDefaults') || 'Reset Defaults' }}</span>
                </button>
              </div>

              <!-- Content Cards -->
              <div class="space-y-2.5">
                <!-- 1. Heavy Graphics & Charting -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 space-y-2.5">
                  <div class="text-3xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                    <OperatorDashboardIcon name="chart" :size="13" class="text-[#1685E5] dark:text-blue-400" />
                    <span>Heavy Graphics &amp; Computations</span>
                  </div>

                  <!-- A. Sparklines -->
                  <div class="flex items-center justify-between pt-2 border-t border-slate-200/60 dark:border-slate-800">
                    <div class="pr-3 min-w-0 flex-1">
                      <div class="flex items-center space-x-1.5">
                        <span class="text-xs font-semibold text-slate-900 dark:text-white truncate">
                          {{ $t('operator.perfSparklines') || 'Parameter Card Sparklines' }}
                        </span>
                        <span
                          v-if="!perfSettings.parameterSparklinesEnabled"
                          class="px-1.5 py-0.5 rounded text-4xs font-semibold bg-amber-50 dark:bg-amber-500/15 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-500/30 leading-none shrink-0"
                        >
                          CPU SAVED
                        </span>
                      </div>
                      <div class="text-3xs text-slate-500 dark:text-slate-400 mt-0.5 leading-snug">
                        {{ $t('operator.perfSparklinesDesc') || 'Render mini SVG sparklines on parameter cards' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-3xs font-semibold tabular-nums" :class="perfSettings.parameterSparklinesEnabled ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'">
                        {{ perfSettings.parameterSparklinesEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('parameterSparklinesEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.parameterSparklinesEnabled ? 'bg-[#1685E5] dark:bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.parameterSparklinesEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.parameterSparklinesEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>

                  <!-- B. Trend Analysis Chart -->
                  <div class="flex items-center justify-between pt-2.5 border-t border-slate-200/60 dark:border-slate-800">
                    <div class="pr-3 min-w-0 flex-1">
                      <div class="flex items-center space-x-1.5">
                        <span class="text-xs font-semibold text-slate-900 dark:text-white truncate">
                          {{ $t('operator.perfTrend') || 'Trend Analysis Chart' }}
                        </span>
                        <span
                          v-if="!perfSettings.trendAnalysisEnabled"
                          class="px-1.5 py-0.5 rounded text-4xs font-semibold bg-amber-50 dark:bg-amber-500/15 text-amber-700 dark:text-amber-400 border border-amber-200 dark:border-amber-500/30 leading-none shrink-0"
                        >
                          UNMOUNTED
                        </span>
                      </div>
                      <div class="text-3xs text-slate-500 dark:text-slate-400 mt-0.5 leading-snug">
                        {{ $t('operator.perfTrendDesc') || 'Multi-parameter trend analysis panel and 1M/5S intervals' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-3xs font-semibold tabular-nums" :class="perfSettings.trendAnalysisEnabled ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'">
                        {{ perfSettings.trendAnalysisEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('trendAnalysisEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.trendAnalysisEnabled ? 'bg-[#1685E5] dark:bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.trendAnalysisEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.trendAnalysisEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- 2. Live Telemetry Cadence -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 space-y-2.5">
                  <div class="flex items-center justify-between">
                    <div class="text-3xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                      <OperatorDashboardIcon name="activity" :size="13" class="text-amber-500" />
                      <span>{{ $t('operator.perfLiveUpdates') || 'Live Dashboard Updates' }}</span>
                    </div>
                    <span class="text-3xs uppercase font-semibold text-[#1685E5] dark:text-blue-400">
                      {{ perfSettings.liveUpdateMode }}
                    </span>
                  </div>

                  <!-- Segmented Buttons -->
                  <div class="p-1 rounded-xl bg-slate-200/60 dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800 grid grid-cols-3 gap-1">
                    <button
                      type="button"
                      @click="setPerfSetting('liveUpdateMode', 'full')"
                      class="py-1.5 px-2 rounded-lg text-2xs font-semibold transition-colors text-center cursor-pointer"
                      :class="perfSettings.liveUpdateMode === 'full'
                        ? 'bg-[#1685E5] text-white btn-text shadow-xs'
                        : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      Full Live
                    </button>
                    <button
                      type="button"
                      @click="setPerfSetting('liveUpdateMode', 'essential')"
                      class="py-1.5 px-2 rounded-lg text-2xs font-semibold transition-colors text-center cursor-pointer"
                      :class="perfSettings.liveUpdateMode === 'essential'
                        ? 'bg-[#1685E5] text-white btn-text shadow-xs'
                        : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      Essential
                    </button>
                    <button
                      type="button"
                      @click="setPerfSetting('liveUpdateMode', 'reduced')"
                      class="py-1.5 px-2 rounded-lg text-2xs font-semibold transition-colors text-center cursor-pointer"
                      :class="perfSettings.liveUpdateMode === 'reduced'
                        ? 'bg-[#1685E5] text-white btn-text shadow-xs'
                        : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      Reduced (1s)
                    </button>
                  </div>

                  <!-- Description for selected mode -->
                  <div class="text-3xs text-slate-500 dark:text-slate-400 leading-snug px-0.5">
                    <span v-if="perfSettings.liveUpdateMode === 'full'">
                      {{ $t('operator.perfLiveFullDesc') || 'All live telemetry, sparklines, trends, and status updated continuously.' }}
                    </span>
                    <span v-else-if="perfSettings.liveUpdateMode === 'essential'">
                      {{ $t('operator.perfLiveEssentialDesc') || 'Only numeric sensor values and vital status update. Graphics skip processing.' }}
                    </span>
                    <span v-else-if="perfSettings.liveUpdateMode === 'reduced'">
                      {{ $t('operator.perfLiveReducedDesc') || 'Throttles UI visual renders to 1s batch intervals without altering sensor polling.' }}
                    </span>
                  </div>
                </div>

                <!-- 3. Alarm Display & Sound -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 space-y-2.5">
                  <div class="text-3xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                    <OperatorDashboardIcon name="bell" :size="13" class="text-rose-500" />
                    <span>{{ $t('operator.perfAlarms') || 'Alarm Display & Sound' }}</span>
                  </div>

                  <!-- Alarm Visual -->
                  <div class="flex items-center justify-between pt-2 border-t border-slate-200/60 dark:border-slate-800">
                    <div class="pr-3 min-w-0 flex-1">
                      <div class="text-xs font-semibold text-slate-900 dark:text-white truncate">
                        {{ $t('operator.perfAlarmVisual') || 'Alarm Visual (Flashing & Pulse)' }}
                      </div>
                      <div class="text-3xs text-slate-500 dark:text-slate-400 leading-snug mt-0.5">
                        {{ $t('operator.perfAlarmVisualDesc') || 'Flashing borders and pulse animations on cards during alarms' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-3xs font-semibold tabular-nums" :class="perfSettings.alarmVisualEnabled ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'">
                        {{ perfSettings.alarmVisualEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('alarmVisualEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.alarmVisualEnabled ? 'bg-[#1685E5] dark:bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.alarmVisualEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.alarmVisualEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>

                  <!-- Alarm Sound -->
                  <div class="flex items-center justify-between pt-2.5 border-t border-slate-200/60 dark:border-slate-800">
                    <div class="pr-3 min-w-0 flex-1">
                      <div class="text-xs font-semibold text-slate-900 dark:text-white truncate">
                        {{ $t('operator.perfAlarmSound') || 'Alarm Sound (Audio Buzzer)' }}
                      </div>
                      <div class="text-3xs text-slate-500 dark:text-slate-400 leading-snug mt-0.5">
                        {{ $t('operator.perfAlarmSoundDesc') || 'Audible chime when critical alarm incidents occur' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-3xs font-semibold tabular-nums" :class="perfSettings.alarmSoundEnabled ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'">
                        {{ perfSettings.alarmSoundEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('alarmSoundEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.alarmSoundEnabled ? 'bg-[#1685E5] dark:bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.alarmSoundEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.alarmSoundEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>

                  <!-- Alarm Safety Notice -->
                  <div class="mt-1 px-3 py-2 rounded-lg bg-amber-50 dark:bg-amber-500/10 border border-amber-200/80 dark:border-amber-500/25 flex items-start space-x-2 text-3xs text-amber-800 dark:text-amber-300 leading-snug">
                    <span class="text-amber-500 shrink-0 font-bold">⚠</span>
                    <span>{{ $t('operator.perfAlarmNotice') || 'Critical alarms remain logged on backend. Disabling visual or sound may delay operator awareness.' }}</span>
                  </div>
                </div>

                <!-- 4. Visual Polish & Animations -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 space-y-2.5">
                  <div class="text-3xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                    <svg class="w-3.5 h-3.5 text-[#1685E5] dark:text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
                    </svg>
                    <span>{{ $t('operator.perfVisualEffects') || 'Visual Effects' }}</span>
                  </div>

                  <!-- Tile Ingestion Blinks -->
                  <div class="flex items-center justify-between pt-2 border-t border-slate-200/60 dark:border-slate-800">
                    <div class="pr-3 min-w-0 flex-1">
                      <div class="text-xs font-semibold text-slate-900 dark:text-white truncate">
                        {{ $t('operator.perfCardBlinks') || 'Tile Ingestion Blinks' }}
                      </div>
                      <div class="text-3xs text-slate-500 dark:text-slate-400 leading-snug mt-0.5">
                        {{ $t('operator.perfCardBlinksDesc') || 'Cyan flash highlight on tiles when new telemetry packets arrive' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-3xs font-semibold tabular-nums" :class="perfSettings.cardBlinksEnabled ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'">
                        {{ perfSettings.cardBlinksEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('cardBlinksEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.cardBlinksEnabled ? 'bg-[#1685E5] dark:bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.cardBlinksEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.cardBlinksEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>

                  <!-- Chart Animations -->
                  <div class="flex items-center justify-between pt-2.5 border-t border-slate-200/60 dark:border-slate-800">
                    <div class="pr-3 min-w-0 flex-1">
                      <div class="text-xs font-semibold text-slate-900 dark:text-white truncate">
                        {{ $t('operator.perfChartAnimations') || 'Chart Path Animations' }}
                      </div>
                      <div class="text-3xs text-slate-500 dark:text-slate-400 leading-snug mt-0.5">
                        {{ $t('operator.perfChartAnimationsDesc') || 'Smooth transition effects on chart lines and curves' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-3xs font-semibold tabular-nums" :class="perfSettings.chartAnimationsEnabled ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'">
                        {{ perfSettings.chartAnimationsEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('chartAnimationsEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.chartAnimationsEnabled ? 'bg-[#1685E5] dark:bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.chartAnimationsEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.chartAnimationsEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 5. ISPU Settings Popover (Strictly AQMS Mode Only — Hidden in WQMS & SPARING) -->
          <div v-if="isAQMSDomain" class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('ispuSettings')"
              class="op-icon-btn"
              :class="{
                'is-emerald': activeToolbarPopover === 'ispuSettings',
                'is-cyan-soft': activeToolbarPopover !== 'ispuSettings' && isAQMSDomain && isIspuBarVisible
              }"
              :title="$t('operator.toolbarIspuSettingsTooltip') || 'Pengaturan & Status Indikator ISPU'"
              aria-label="ISPU Settings"
            >
              <OperatorDashboardIcon name="ispu" :size="16" />
              <span v-if="isAQMSDomain && isIspuBarVisible" class="op-corner-dot dot-emerald"></span>
            </button>
            <div
              v-if="activeToolbarPopover === 'ispuSettings'"
              class="absolute left-0 top-full mt-2 z-50 w-[385px] sm:w-[405px] max-w-[calc(100vw-24px)] rounded-2xl shadow-xl p-4 text-left border border-slate-200/90 dark:border-slate-800/80 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-sans"
              @click.stop
            >
              <div class="flex items-center justify-between pb-3 mb-3 border-b border-slate-100 dark:border-slate-800/80">
                <div class="flex items-center space-x-2.5">
                  <div class="w-8 h-8 rounded-xl flex items-center justify-center bg-emerald-50 dark:bg-emerald-500/15 border border-emerald-200/90 dark:border-emerald-500/30 text-emerald-600 dark:text-emerald-400 flex-shrink-0">
                    <OperatorDashboardIcon name="ispu" :size="16" />
                  </div>
                  <div>
                    <div class="flex items-center space-x-2">
                      <span class="text-[13px] font-bold tracking-tight text-slate-900 dark:text-white leading-tight">
                        {{ $t('operator.ispuSettings') || 'ISPU Settings' }}
                      </span>
                      <span class="px-1.5 py-0.5 rounded-md text-3xs font-semibold bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/30 leading-none">
                        AQMS
                      </span>
                    </div>
                    <div class="text-3xs text-slate-500 dark:text-slate-400 font-medium leading-snug mt-0.5">
                      Indeks Standar Pencemar Udara
                    </div>
                  </div>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- Unified Station Scope Row -->
              <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 mb-2.5 space-y-1">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-semibold text-slate-900 dark:text-white">
                    Lingkup Kalkulasi ISPU
                  </span>
                  <span class="px-2 py-0.5 rounded-md text-3xs font-semibold bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border border-[#B8D4EA]/80 dark:border-blue-500/30 leading-none">
                    UNIFIED
                  </span>
                </div>
                <div class="text-3xs text-slate-500 dark:text-slate-400 leading-relaxed">
                  Menggabungkan seluruh parameter resmi yang aktif dari semua device yang aktif di stasiun ini.
                </div>
              </div>

              <!-- Toggle Show/Hide Indicator Panel Row -->
              <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 mb-2.5">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-semibold text-slate-900 dark:text-white">
                    {{ $t('operator.showIspuBar') || 'Show ISPU Indicator Bar' }}
                  </span>
                  <button
                    type="button"
                    @click="toggleIspuBarVisibility"
                    class="relative inline-flex h-5 w-9 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                    :class="isIspuBarVisible ? 'bg-[#1685E5] dark:bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'"
                    role="switch"
                    :aria-checked="isIspuBarVisible"
                  >
                    <span
                      aria-hidden="true"
                      class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out"
                      :class="isIspuBarVisible ? 'translate-x-4' : 'translate-x-0'"
                    ></span>
                  </button>
                </div>
                <div class="text-3xs text-slate-500 dark:text-slate-400 mt-1 leading-snug">
                  {{ $t('operator.showIspuBarDesc') || 'Display real-time air quality index bar below parameters' }}
                </div>
              </div>

              <!-- Standard Info Card -->
              <div class="p-3 rounded-xl bg-slate-50/60 dark:bg-slate-900/40 border border-slate-200/75 dark:border-slate-800 mb-3 text-2xs space-y-2">
                <div class="flex items-center justify-between">
                  <span class="text-slate-500 dark:text-slate-400 font-medium">Standard</span>
                  <span class="font-semibold text-[#1685E5] dark:text-blue-400">Permen LHK No. 14 / 2020</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-500 dark:text-slate-400 font-medium">Interval</span>
                  <span class="font-semibold text-slate-800 dark:text-slate-200">24-Hour Rolling Average</span>
                </div>
                <div class="flex items-center justify-between">
                  <span class="text-slate-500 dark:text-slate-400 font-medium">Parameters</span>
                  <span class="font-medium text-slate-700 dark:text-slate-300 tabular-nums">PM2.5, PM10, SO2, NO2, O3, CO, HC</span>
                </div>
              </div>

              <!-- Navigation to Dedicated ISPU Analysis -->
              <button
                type="button"
                @click="navigateToRoute('/monitoring/ispu')"
                class="w-full py-2 px-3 rounded-xl text-xs font-semibold transition-colors border flex items-center justify-center space-x-2 bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA]/80 dark:border-blue-500/30 hover:bg-[#dceeff] dark:hover:bg-blue-600/25 shadow-xs"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                <span>{{ $t('operator.openIspuAnalysis') || 'Open ISPU Analysis & Reports →' }}</span>
              </button>
            </div>
          </div>

          <!-- 6. Day / Night Mode Toggle (Single Icon Button Matching Operator Console) -->
          <button
            type="button"
            id="btn-operator-theme-toggle"
            @click="toggleTheme"
            class="op-icon-btn"
            :class="isDark ? 'text-cyan-300' : 'text-amber-500'"
            :title="isDark ? ($t('operator.switchToDayMode') || 'Beralih ke Mode Siang (Day Mode)') : ($t('operator.switchToNightMode') || 'Beralih ke Mode Malam (Night Mode)')"
            :aria-label="isDark ? 'Switch to Day Mode' : 'Switch to Night Mode'"
          >
            <OperatorDashboardIcon v-if="isDark" name="moon" :size="16" />
            <OperatorDashboardIcon v-else name="sun" :size="16" />
          </button>

          <!-- Dismissible Active Filter Badges -->
          <div
            v-if="selectedCategoryFilter !== 'all' || filterPinnedOnly"
            class="flex items-center gap-1 ml-0.5 pl-1.5 border-l border-slate-200/80 dark:border-slate-700/60"
          >
            <button
              v-if="selectedCategoryFilter !== 'all'"
              type="button"
              @click="selectedCategoryFilter = 'all'"
              class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[9.5px] font-semibold bg-sky-500/15 text-sky-700 dark:text-sky-300 border border-sky-500/35 hover:bg-sky-500/25 cursor-pointer"
              title="Hapus filter kategori"
            >
              <span class="max-w-[130px] truncate">{{ selectedCategoryTitle }}</span>
              <span class="font-bold leading-none">×</span>
            </button>
            <button
              v-if="filterPinnedOnly"
              type="button"
              @click="toggleFilterPinned"
              class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[9.5px] font-semibold bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/35 hover:bg-amber-500/25 cursor-pointer"
              title="Hapus filter favorit"
            >
              <span>★ Favorit</span>
              <span class="font-bold leading-none">×</span>
            </button>
          </div>
        </div>

        <!-- Right Icon Cluster: Actuators, Health & Alarms, System, Network, Station GPS -->
        <div class="tool-cluster flex-shrink-0">
          <!-- 7. Actuators & Controls (Icon Only + Running Counter Badge) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              id="btn-toolbar-actuators"
              @click.stop="toggleToolbarPopover('actuators')"
              class="op-icon-btn"
              :class="{
                'is-amber': activeToolbarPopover === 'actuators',
                'is-emerald': activeToolbarPopover !== 'actuators' && activeActuatorsCount > 0
              }"
              title="Kontrol & Aktuator (Pompa, Katup Purge, Relay)"
              aria-label="Controls & Actuators"
            >
              <svg width="16" height="16" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
              </svg>
              <span
                v-if="activeActuatorsCount > 0"
                class="op-corner-pill tone-emerald"
                title="Aktuator aktif menyala"
              >{{ activeActuatorsCount }}</span>
            </button>
            <transition
              enter-active-class="transition duration-150 ease-out"
              enter-from-class="opacity-0 scale-95"
              enter-to-class="opacity-100 scale-100"
              leave-active-class="transition duration-100 ease-in"
              leave-from-class="opacity-100 scale-100"
              leave-to-class="opacity-0 scale-95"
            >
              <div
                v-if="activeToolbarPopover === 'actuators'"
                class="absolute right-0 top-full mt-2 w-96 max-w-[92vw] bg-white dark:bg-[#0B0F19] border border-slate-200/90 dark:border-slate-800/80 rounded-2xl shadow-xl z-50 p-4 space-y-3.5 select-none font-sans"
                @click.stop
              >
                <!-- Popover Header -->
                <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800/80 pb-3">
                  <div class="flex items-center space-x-2.5">
                    <span class="w-8 h-8 rounded-xl bg-amber-50 dark:bg-amber-500/15 border border-amber-200/80 dark:border-amber-500/30 text-amber-600 dark:text-amber-400 flex items-center justify-center text-xs font-bold">⚡</span>
                    <div>
                      <h3 class="text-[13px] font-bold tracking-tight text-slate-900 dark:text-white leading-tight">
                        Controls &amp; Actuators
                      </h3>
                      <span class="text-3xs font-medium text-slate-500 dark:text-slate-400 block leading-snug mt-0.5">
                        Relay, pump &amp; purge valve control
                      </span>
                    </div>
                  </div>
                  <span class="text-3xs font-semibold tabular-nums px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200/80 dark:border-slate-700">
                    {{ activeActuatorsCount }} Active / {{ operatorActuators.length }} Total
                  </span>
                </div>

                <!-- Actuator Cards (Reference Design Opsi B from user screenshot) -->
                <div v-if="loadingActuators" class="py-8 text-center text-xs text-slate-400">
                  Loading equipment status...
                </div>

                <div v-else-if="operatorActuators.length === 0" class="py-8 text-center text-slate-400 text-xs">
                  <p class="text-2xl mb-1">⚡</p>
                  <p class="font-bold text-slate-700 dark:text-slate-300">No Actuators Configured</p>
                  <p class="text-3xs mt-1 text-slate-500">Add equipment under Device Detail &gt; Controls & Actuators.</p>
                </div>

                <div v-else class="space-y-2.5 max-h-80 overflow-y-auto pr-0.5">
                  <div
                    v-for="act in operatorActuators"
                    :key="act.id"
                    class="p-3 rounded-lg border transition-all"
                    :class="act.current_state === 'ON'
                      ? 'bg-emerald-950/20 border-emerald-500/40 ring-1 ring-emerald-500/20'
                      : 'bg-slate-50 dark:bg-[#121E30] border-slate-200 dark:border-slate-800'"
                  >
                    <div class="flex items-start justify-between">
                      <div class="flex items-center space-x-2.5">
                        <div
                          class="w-8 h-8 rounded-lg flex items-center justify-center text-sm"
                          :class="act.current_state === 'ON'
                            ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                            : 'bg-slate-200 dark:bg-slate-800 text-slate-400'"
                        >
                          {{ getActuatorIcon(act.icon) }}
                        </div>
                        <div>
                          <div class="font-bold text-xs text-slate-800 dark:text-slate-100">{{ act.actuator_name }}</div>
                          <div class="text-3xs font-mono text-slate-400">
                            <span v-if="!isOperator">DEV: {{ act.device ? (act.device.device_code || act.device.code) : act.device_id }} &bull; </span>{{ act.control_type === 'COIL_FC05' ? 'Coil' : 'Reg' }} #{{ act.register_address }}
                          </div>
                        </div>
                      </div>

                      <!-- State Pill -->
                      <span
                        :class="act.current_state === 'ON'
                          ? 'bg-emerald-500/20 text-emerald-400 border-emerald-500/40 font-extrabold'
                          : 'bg-slate-200/60 dark:bg-slate-800 text-slate-500 border-slate-300 dark:border-slate-700 font-semibold'"
                        class="px-2 py-0.5 rounded-full text-3xs font-mono border flex items-center space-x-1"
                      >
                        <span class="w-1.5 h-1.5 rounded-full" :class="act.current_state === 'ON' ? 'bg-emerald-400 animate-pulse' : 'bg-slate-400'"></span>
                        <span>{{ act.current_state === 'ON' ? 'RUNNING (ON)' : 'STANDBY (OFF)' }}</span>
                      </span>
                    </div>

                    <!-- Bottom row: Countdown / Schedule + Manual Action Button -->
                    <div class="mt-3 flex items-center justify-between pt-2 border-t border-slate-100 dark:border-slate-800/80">
                      <div class="text-3xs font-mono text-slate-400">
                        <span v-if="act.current_state === 'ON' && getActuatorCountdown(act)" class="text-emerald-400 font-bold">
                          ⏱️ Auto-Off: {{ getActuatorCountdown(act) }} sisa
                        </span>
                        <span v-else-if="act.schedules && act.schedules.length > 0">
                          Jadwal: Setiap {{ formatActuatorDuration(act.schedules[0].interval_seconds) }}
                        </span>
                        <span v-else class="text-slate-500">
                          Manual Only
                        </span>
                      </div>

                      <!-- Action Button -->
                      <button
                        v-if="act.current_state === 'ON'"
                        type="button"
                        @click="toggleOperatorActuator(act, 'OFF')"
                        :disabled="togglingActuatorId === act.id"
                        class="py-1 px-3 rounded-md bg-rose-600 hover:bg-rose-500 active:bg-rose-700 text-white font-bold text-3xs shadow-sm transition-colors flex items-center space-x-1 disabled:opacity-50 cursor-pointer"
                      >
                        <span v-if="togglingActuatorId === act.id" class="w-2.5 h-2.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                        <span>Manual STOP</span>
                      </button>

                      <button
                        v-else
                        type="button"
                        @click="toggleOperatorActuator(act, 'ON')"
                        :disabled="togglingActuatorId === act.id"
                        class="py-1 px-3 rounded-md bg-emerald-600 hover:bg-emerald-500 active:bg-emerald-700 text-white font-bold text-3xs shadow-sm transition-colors flex items-center space-x-1 disabled:opacity-50 cursor-pointer"
                      >
                        <span v-if="togglingActuatorId === act.id" class="w-2.5 h-2.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></span>
                        <span>Manual START</span>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- Footer Link to Device Setup -->
                <div class="pt-2 border-t border-slate-100 dark:border-slate-800 flex justify-between items-center text-3xs">
                  <span class="text-slate-400">Automated cycle by scheduler</span>
                  <button
                    v-if="!isOperator"
                    type="button"
                    @click="navigateToRoute('/monitoring/devices')"
                    class="text-cyan-500 hover:text-cyan-400 font-bold hover:underline cursor-pointer"
                  >
                    Manage Equipment &rarr;
                  </button>
                </div>
              </div>
            </transition>
          </div>

          <!-- 8. Consolidated Station Health & Alarms (Icon Only + Count Badge / Dot) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('healthAndAlarms')"
              class="op-icon-btn"
              :class="{
                'is-rose': activeCriticalCount > 0,
                'is-amber': activeCriticalCount === 0 && (activeAlarmsList.length > 0 || summaryStats.offlineCount > 0 || isTelemetryStale),
                'is-active': activeToolbarPopover === 'healthAndAlarms' && activeAlarmsList.length === 0
              }"
              :title="healthAndAlarmsTooltip"
              aria-label="Alarms & Station Health Center"
            >
              <OperatorDashboardIcon name="pulse" :size="16" />
              <span
                v-if="activeAlarmsList.length > 0"
                class="op-corner-pill"
                :class="activeCriticalCount > 0 ? 'tone-rose' : 'tone-amber'"
              >{{ activeAlarmsList.length }}</span>
              <span
                v-else
                class="op-corner-dot"
                :class="summaryStats.offlineCount > 0 || isTelemetryStale ? 'dot-amber' : 'dot-emerald'"
              ></span>
            </button>
              <div
                v-if="activeToolbarPopover === 'healthAndAlarms'"
                class="absolute right-0 top-full mt-2 z-50 w-[410px] sm:w-[432px] max-w-[calc(100vw-24px)] rounded-2xl shadow-xl p-4 text-left border border-slate-200/90 dark:border-slate-800/80 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-sans"
                @click.stop
              >
                <!-- Panel Header -->
                <div class="flex items-center justify-between pb-3 mb-3 border-b border-slate-100 dark:border-slate-800/80">
                  <div class="flex items-center space-x-2.5">
                    <div
                      class="w-8 h-8 rounded-xl flex items-center justify-center border flex-shrink-0"
                      :class="activeCriticalCount > 0
                        ? 'bg-rose-50 dark:bg-rose-500/15 border-rose-200 dark:border-rose-500/40 text-rose-600 dark:text-rose-400'
                        : activeAlarmsList.length > 0
                          ? 'bg-amber-50 dark:bg-amber-500/15 border-amber-200 dark:border-amber-500/40 text-amber-600 dark:text-amber-400'
                          : 'bg-[#EAF5FF] dark:bg-blue-600/15 border-[#B8D4EA]/80 dark:border-blue-500/30 text-[#1685E5] dark:text-blue-400'"
                    >
                      <OperatorDashboardIcon name="pulse" :size="16" />
                    </div>
                    <div>
                      <div class="flex items-center space-x-2">
                        <h3 class="text-[13px] font-bold tracking-tight text-slate-900 dark:text-white leading-tight">
                          Health &amp; Notification Center
                        </h3>
                        <span
                          class="px-2 py-0.5 rounded-md text-3xs font-semibold uppercase tracking-wide border leading-none"
                          :class="activeCriticalCount > 0
                            ? 'bg-rose-50 dark:bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-200 dark:border-rose-500/30'
                            : activeAlarmsList.length > 0
                              ? 'bg-amber-50 dark:bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-500/30'
                              : 'bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30'"
                        >
                          {{ activeCriticalCount > 0 ? 'CRITICAL' : activeAlarmsList.length > 0 ? 'WARNING' : 'HEALTHY' }}
                        </span>
                      </div>
                      <div class="text-3xs font-medium text-slate-500 dark:text-slate-400 mt-0.5 leading-snug">
                        Alarms · Device Fleet · Telemetry Stream
                      </div>
                    </div>
                  </div>
                  <button
                    type="button"
                    @click="closeToolbarPopover"
                    class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                    title="Close"
                  >
                    <OperatorDashboardIcon name="close" :size="14" />
                  </button>
                </div>

                <!-- 3 Interactive Metric Cards acting as Tab Selectors -->
                <div class="grid grid-cols-3 gap-2 mb-3">
                  <!-- Tab 1: Alarms -->
                  <button
                    type="button"
                    @click="healthActiveTab = 'alarms'"
                    class="p-2 rounded-md border text-left transition flex flex-col justify-between"
                    :class="healthActiveTab === 'alarms'
                      ? (activeAlarmsList.length > 0
                          ? 'bg-rose-500/15 border-rose-500/50 text-rose-600 dark:text-rose-300 shadow-xs'
                          : 'bg-cyan-500/15 border-cyan-500/50 text-cyan-700 dark:text-cyan-300 shadow-xs')
                      : 'bg-slate-50 dark:bg-[#070E1A] border-slate-200/80 dark:border-[#1E3048] text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-700'"
                  >
                    <div class="flex items-center justify-between w-full">
                      <span class="text-[9px] uppercase font-bold tracking-wider">Alarms</span>
                      <span v-if="activeCriticalCount > 0" class="w-2 h-2 rounded-full bg-rose-500 animate-pulse"></span>
                      <span v-else-if="activeAlarmsList.length > 0" class="w-2 h-2 rounded-full bg-amber-500"></span>
                      <span v-else class="w-2 h-2 rounded-full bg-emerald-400"></span>
                    </div>
                    <div class="mt-1.5 flex items-baseline justify-between">
                      <span class="text-sm font-bold tabular-nums" :class="activeAlarmsList.length > 0 ? 'text-rose-500' : 'text-emerald-500 dark:text-emerald-400'">
                        {{ activeAlarmsList.length }}
                      </span>
                      <span
                        class="text-[8px] font-bold uppercase px-1 py-0.2 rounded"
                        :class="activeCriticalCount > 0 ? 'bg-rose-500/20 text-rose-400' : activeAlarmsList.length > 0 ? 'bg-amber-500/20 text-amber-400' : 'text-slate-400'"
                      >
                        {{ activeCriticalCount > 0 ? 'CRIT' : activeAlarmsList.length > 0 ? 'WARN' : 'NORMAL' }}
                      </span>
                    </div>
                  </button>

                  <!-- Tab 2: Devices -->
                  <button
                    type="button"
                    @click="healthActiveTab = 'devices'"
                    class="p-2 rounded-md border text-left transition flex flex-col justify-between"
                    :class="healthActiveTab === 'devices'
                      ? 'bg-cyan-500/15 border-cyan-500/50 text-cyan-700 dark:text-cyan-300 shadow-xs'
                      : 'bg-slate-50 dark:bg-[#070E1A] border-slate-200/80 dark:border-[#1E3048] text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-700'"
                  >
                    <div class="flex items-center justify-between w-full">
                      <span class="text-[9px] uppercase font-bold tracking-wider">Devices</span>
                      <span :class="summaryStats.offlineCount > 0 ? 'w-2 h-2 rounded-full bg-rose-500' : 'w-2 h-2 rounded-full bg-emerald-400'"></span>
                    </div>
                    <div class="mt-1.5 flex items-baseline justify-between">
                      <span class="text-sm font-bold tabular-nums text-slate-900 dark:text-white">
                        {{ summaryStats.onlineCount }}/{{ summaryStats.totalDevices }}
                      </span>
                      <span
                        class="text-[8px] font-bold uppercase px-1 py-0.2 rounded"
                        :class="summaryStats.offlineCount > 0 ? 'bg-rose-500/20 text-rose-400' : 'text-emerald-500 dark:text-emerald-400'"
                      >
                        {{ summaryStats.offlineCount > 0 ? summaryStats.offlineCount + ' Off' : 'Online' }}
                      </span>
                    </div>
                  </button>

                  <!-- Tab 3: Telemetry Stream -->
                  <button
                    type="button"
                    @click="healthActiveTab = 'telemetry'"
                    class="p-2 rounded-md border text-left transition flex flex-col justify-between"
                    :class="healthActiveTab === 'telemetry'
                      ? 'bg-cyan-500/15 border-cyan-500/50 text-cyan-700 dark:text-cyan-300 shadow-xs'
                      : 'bg-slate-50 dark:bg-[#070E1A] border-slate-200/80 dark:border-[#1E3048] text-slate-600 dark:text-slate-400 hover:border-slate-300 dark:hover:border-slate-700'"
                  >
                    <div class="flex items-center justify-between w-full">
                      <span class="text-[9px] uppercase font-bold tracking-wider">Stream</span>
                      <span :class="isTelemetryStale ? 'w-2 h-2 rounded-full bg-rose-500 animate-pulse' : 'w-2 h-2 rounded-full bg-emerald-400'"></span>
                    </div>
                    <div class="mt-1.5 flex items-baseline justify-between">
                      <span class="text-sm font-bold tabular-nums" :class="isTelemetryStale ? 'text-rose-500' : 'text-emerald-500 dark:text-emerald-400'">
                        {{ formatRelativeSeconds(lastTelemetrySecondsAgo) }}
                      </span>
                      <span
                        class="text-[8px] font-bold uppercase px-1 py-0.2 rounded"
                        :class="isTelemetryStale ? 'bg-rose-500/20 text-rose-400' : 'text-emerald-500 dark:text-emerald-400'"
                      >
                        {{ isTelemetryStale ? 'Stale' : 'Live' }}
                      </span>
                    </div>
                  </button>
                </div>

                <!-- TAB 1 CONTENT: ACTIVE ALARMS -->
                <div v-if="healthActiveTab === 'alarms'" class="space-y-2.5 animate-fadeIn">
                  <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048]">
                    <div class="flex items-center justify-between mb-2">
                      <div class="flex items-baseline space-x-2">
                        <span
                          class="text-2xl font-bold tabular-nums tracking-tight font-mono"
                          :class="activeCriticalCount > 0 ? 'text-rose-600 dark:text-rose-400' : activeAlarmsList.length > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'"
                        >
                          {{ activeAlarmsList.length }}
                        </span>
                        <span class="text-xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wide">
                          Active Incidents
                        </span>
                      </div>
                      <div class="flex items-center space-x-1.5 text-[9.5px]">
                        <span v-if="activeCriticalCount > 0" class="px-2 py-0.5 rounded font-bold bg-rose-500/15 text-rose-600 dark:text-rose-400 border border-rose-500/30">
                          {{ activeCriticalCount }} CRITICAL
                        </span>
                        <span v-if="activeWarningCount > 0" class="px-2 py-0.5 rounded font-bold bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30">
                          {{ activeWarningCount }} WARNING
                        </span>
                        <span v-if="activeAlarmsList.length === 0" class="px-2 py-0.5 rounded font-semibold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 flex items-center space-x-1">
                          <span>✓</span>
                          <span>Normal Limits</span>
                        </span>
                      </div>
                    </div>
                    <div class="text-[9.5px] text-slate-500 dark:text-slate-400 pt-2 border-t border-slate-200/50 dark:border-slate-800/80 leading-relaxed">
                      <span v-if="highestPriorityAlarm">
                        Top Priority: <strong class="text-rose-600 dark:text-rose-400">{{ getAlarmDisplayTitle(highestPriorityAlarm) }}</strong> ({{ highestPriorityAlarm.message }})
                      </span>
                      <span v-else class="text-slate-600 dark:text-slate-400">
                        All sensor measurements are currently within regulatory thresholds.
                      </span>
                    </div>
                  </div>

                  <!-- Active Alarms Incident List -->
                  <div>
                    <div class="text-[9px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5 flex items-center justify-between">
                      <span>Active Incidents</span>
                      <router-link
                        to="/alarm/history"
                        @click.native="closeToolbarPopover"
                        class="text-cyan-600 dark:text-cyan-400 hover:underline"
                      >
                        Full History →
                      </router-link>
                    </div>
                    <div v-if="activeAlarmsList.length > 0" class="max-h-48 overflow-y-auto custom-scroller space-y-1.5 pr-0.5">
                      <div
                        v-for="alm in activeAlarmsList.slice(0, 6)"
                        :key="'popover-alm-' + (alm.alarm_id || alm.id)"
                        class="p-2 rounded bg-slate-50/70 dark:bg-[#0E1726] border border-slate-200/70 dark:border-[#1E3048] text-[9.5px]"
                      >
                        <div class="flex items-center justify-between mb-1">
                          <span class="font-bold text-slate-800 dark:text-slate-200 truncate pr-1" :title="getAlarmDisplayTitle(alm)">
                            {{ getAlarmDisplayTitle(alm) }}
                          </span>
                          <span
                            class="px-1.5 py-0.2 rounded text-[7.5px] font-bold flex-shrink-0 uppercase"
                            :class="alm.severity === 'CRITICAL' ? 'bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/40' : 'bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/40'"
                          >
                            {{ alm.severity }}
                          </span>
                        </div>
                        <div class="text-[8.5px] text-slate-500 dark:text-slate-400 flex items-center justify-between gap-1.5">
                          <span v-if="alm.trigger_value != null" class="truncate">
                            Val: <strong class="text-slate-700 dark:text-slate-300 font-mono">{{ alm.trigger_value }}</strong>
                            <span v-if="alm.threshold_value != null"> (Limit: {{ alm.threshold_value }})</span>
                          </span>
                          <span v-else class="truncate pr-1 text-slate-600 dark:text-slate-400" :title="alm.message">
                            {{ alm.message || formatAlarmCode(alm.alarm_code) }}
                          </span>
                          <span class="opacity-80 flex-shrink-0 ml-auto whitespace-nowrap">{{ formatRelative(alm.triggered_at || alm.created_at) }}</span>
                        </div>
                      </div>
                    </div>
                    <div v-else class="py-3 px-3 text-center text-slate-400 dark:text-slate-500 text-[10px] bg-slate-50/50 dark:bg-[#0E1726]/50 rounded-lg border border-dashed border-slate-200 dark:border-[#20354B]">
                      No active alarm incidents detected.
                    </div>
                  </div>

                  <!-- Action Link -->
                  <router-link
                    to="/alarm/active"
                    @click.native="closeToolbarPopover"
                    class="w-full py-2 px-3 rounded-md text-[10.5px] font-semibold transition border flex items-center justify-center space-x-1.5"
                    :class="activeAlarmsList.length > 0
                      ? 'bg-rose-500/10 text-rose-700 dark:text-rose-300 border-rose-500/30 hover:bg-rose-500/20 hover:border-rose-400'
                      : 'bg-slate-50 hover:bg-slate-100 dark:bg-[#0E1726] dark:hover:bg-[#162538] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400'"
                  >
                    <span>🛡️</span>
                    <span>{{ activeAlarmsList.length > 0 ? 'Open Alarm Console & Acknowledge →' : 'View Alarm Management Console →' }}</span>
                  </router-link>
                </div>

                <!-- TAB 2 CONTENT: DEVICE FLEET STATUS -->
                <div v-else-if="healthActiveTab === 'devices'" class="space-y-2.5 animate-fadeIn">
                  <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048]">
                    <div class="flex items-center justify-between mb-2">
                      <div class="flex items-baseline space-x-2">
                        <span class="text-2xl font-bold text-slate-900 dark:text-white tabular-nums tracking-tight font-mono">
                          {{ summaryStats.onlineCount }}<span class="text-slate-400 text-lg">/{{ summaryStats.totalDevices }}</span>
                        </span>
                        <span class="text-xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wide">
                          Devices Online
                        </span>
                      </div>
                      <span
                        class="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider border leading-none"
                        :class="summaryStats.offlineCount > 0 ? 'bg-rose-500/15 text-rose-600 dark:text-rose-400 border-rose-500/30' : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border-emerald-500/30'"
                      >
                        {{ Math.round((summaryStats.onlineCount / (summaryStats.totalDevices || 1)) * 100) }}% Active
                      </span>
                    </div>
                    <div class="w-full h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all duration-300"
                        :class="summaryStats.offlineCount > 0 ? 'bg-amber-500' : 'bg-emerald-500'"
                        :style="{ width: Math.round((summaryStats.onlineCount / (summaryStats.totalDevices || 1)) * 100) + '%' }"
                      ></div>
                    </div>
                    <div class="flex items-center justify-between text-[9.5px] text-slate-500 dark:text-slate-400 mt-2 pt-2 border-t border-slate-200/50 dark:border-slate-800/80">
                      <span>Total Registered: <strong class="text-slate-700 dark:text-slate-300 font-mono">{{ summaryStats.totalDevices }} units</strong></span>
                      <span>{{ isPassiveOrCloudTelemetry ? 'Push Timeout:' : 'Cycle Rate:' }} <strong class="text-cyan-600 dark:text-cyan-400 font-mono">{{ isPassiveOrCloudTelemetry ? `${formatRelativeSeconds(configuredStaleTimeoutSeconds)} (${configuredStaleTimeoutSeconds}s)` : `${Math.round(configuredPollingIntervalMs / 1000)}s` }}</strong></span>
                    </div>
                  </div>

                  <!-- Connected Devices List -->
                  <div>
                    <div class="text-[9px] font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400 mb-1.5">
                      Connected Fleet Nodes
                    </div>
                    <div class="max-h-36 overflow-y-auto custom-scroller space-y-1.5 pr-0.5">
                      <div
                        v-for="dev in activeFleetDevices"
                        :key="'popover-dev-' + dev.id"
                        class="p-2 rounded bg-slate-50/70 dark:bg-[#0E1726] border border-slate-200/70 dark:border-[#1E3048] flex items-center justify-between text-[9.5px]"
                      >
                        <div class="flex items-center space-x-2 truncate pr-1">
                          <span
                            class="w-2 h-2 rounded-full flex-shrink-0"
                            :class="isDeviceActiveOnline(dev) ? 'bg-emerald-500' : 'bg-rose-500'"
                          ></span>
                          <span class="font-bold text-slate-800 dark:text-slate-200 truncate">{{ dev.device_code }}</span>
                          <span class="text-slate-400 truncate">({{ dev.device_name || 'Generic' }})</span>
                        </div>
                        <span
                          class="px-1.5 py-0.2 rounded text-[8px] font-bold flex-shrink-0"
                          :class="isDeviceActiveOnline(dev) ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20'"
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
                    class="w-full py-2 px-3 rounded-md text-[10.5px] font-semibold transition border flex items-center justify-center space-x-1.5 bg-slate-50 hover:bg-slate-100 dark:bg-[#0E1726] dark:hover:bg-[#162538] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400"
                  >
                    <span>⚙️</span>
                    <span>Configure Device Fleet & Telemetry →</span>
                  </button>
                </div>

                <!-- TAB 3 CONTENT: TELEMETRY STREAM DIAGNOSTICS -->
                <div v-else-if="healthActiveTab === 'telemetry'" class="space-y-2.5 animate-fadeIn">
                  <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048]">
                    <div class="flex items-center justify-between mb-2">
                      <div class="flex items-baseline space-x-2">
                        <span
                          class="text-2xl font-bold tabular-nums tracking-tight font-mono"
                          :class="isTelemetryStale ? 'text-rose-600 dark:text-rose-400' : 'text-slate-900 dark:text-white'"
                        >
                          {{ formatRelativeSeconds(lastTelemetrySecondsAgo) }}
                        </span>
                        <span class="text-xs text-slate-500 dark:text-slate-400 font-semibold uppercase tracking-wide">
                          {{ isPassiveOrCloudTelemetry ? 'Last Packet Age' : 'Packet Latency' }}
                        </span>
                      </div>
                      <span
                        class="px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider border leading-none"
                        :class="summaryStats.goodSensors > 0 ? 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border-emerald-500/30' : 'bg-rose-500/15 text-rose-600 dark:text-rose-400 border-rose-500/30'"
                      >
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
                    <div class="flex items-center justify-between text-[9.5px] text-slate-500 dark:text-slate-400 mt-2 pt-2 border-t border-slate-200/50 dark:border-slate-800/80">
                      <span>Last Packet: <strong class="text-slate-700 dark:text-slate-300 font-mono">{{ lastTelemetryTimeStr || 'Just now' }}</strong></span>
                      <span>Timed out: <strong :class="summaryStats.staleSensors > 0 ? 'text-amber-500 font-bold font-mono' : 'text-slate-700 dark:text-slate-300 font-mono'">{{ summaryStats.staleSensors }}</strong></span>
                    </div>
                  </div>

                  <!-- Stream Diagnostics Information -->
                  <div class="p-2.5 rounded bg-slate-50/70 dark:bg-[#0E1726] border border-slate-200/70 dark:border-[#1E3048] text-[9.5px] space-y-1.5">
                    <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                      <span class="text-slate-400">Stream Status:</span>
                      <span class="font-semibold" :class="isTelemetryStale ? 'text-rose-500' : 'text-emerald-500'">
                        {{ isTelemetryStale ? 'Delay Warning' : (isPassiveOrCloudTelemetry ? 'Active Cloud Push Stream' : 'Active High-Speed Acquisition') }}
                      </span>
                    </div>
                    <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                      <span class="text-slate-400">{{ isPassiveOrCloudTelemetry ? 'Receive Timeout:' : 'Polling Interval:' }}</span>
                      <span class="font-mono">{{ isPassiveOrCloudTelemetry ? `${formatRelativeSeconds(configuredStaleTimeoutSeconds)} (${configuredStaleTimeoutSeconds}s Timeout)` : `${Math.round(configuredPollingIntervalMs / 1000)}s Dynamic Cycle` }}</span>
                    </div>
                    <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                      <span class="text-slate-400">Total Monitored:</span>
                      <span class="font-bold text-cyan-600 dark:text-cyan-400 font-mono">{{ allParameters.length }} Parameters</span>
                    </div>
                  </div>

                  <!-- Action Link -->
                  <button
                    type="button"
                    @click="navigateToRoute('/monitoring/analysis')"
                    class="w-full py-2 px-3 rounded-md text-[10.5px] font-semibold transition border flex items-center justify-center space-x-1.5 bg-slate-50 hover:bg-slate-100 dark:bg-[#0E1726] dark:hover:bg-[#162538] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400"
                  >
                    <span>📊</span>
                    <span>Open Telemetry Analytics & Trends →</span>
                  </button>
                </div>
              </div>
          </div>

          <!-- 9. System Monitor (Icon Only + Dot) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('systemMonitor')"
              class="op-icon-btn"
              :class="{ 'is-active': activeToolbarPopover === 'systemMonitor' }"
              :title="($t('operator.toolbarSystemMonitorTooltip') || 'Monitor Sistem') + ': CPU ' + Math.round(systemStatus.cpu_percent || 0) + '% · RAM ' + Math.round(systemStatus.ram_percent || 0) + '% · Disk ' + Math.round(systemStatus.disk_percent || 0) + '%'"
              aria-label="System Monitor"
            >
              <OperatorDashboardIcon name="sensor" :size="16" />
              <span
                class="op-corner-dot"
                :class="(systemStatus.cpu_percent > 85 || systemStatus.ram_percent > 90) ? 'dot-amber' : 'dot-emerald'"
              ></span>
            </button>
              <div
                v-if="activeToolbarPopover === 'systemMonitor'"
                class="absolute right-0 top-full mt-2 z-50 w-[340px] sm:w-[360px] max-w-[calc(100vw-24px)] rounded-2xl shadow-xl p-4 text-left border border-slate-200/90 dark:border-slate-800/80 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-sans"
                @click.stop
              >
                <div class="flex items-center justify-between pb-3 mb-3 border-b border-slate-100 dark:border-slate-800/80">
                  <div class="flex items-center space-x-2.5">
                    <div class="w-8 h-8 rounded-xl flex items-center justify-center bg-[#EAF5FF] dark:bg-blue-600/15 border border-[#B8D4EA]/80 dark:border-blue-500/30 text-[#1685E5] dark:text-blue-400 flex-shrink-0">
                      <OperatorDashboardIcon name="sensor" :size="16" />
                    </div>
                    <div>
                      <h3 class="text-[13px] font-bold tracking-tight text-slate-900 dark:text-white leading-tight">
                        {{ $t('operator.summarySystem') || 'System Monitor' }}
                      </h3>
                      <div class="text-3xs font-medium text-slate-500 dark:text-slate-400 mt-0.5 leading-snug">
                        Hardware &amp; Node Health
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1.5">
                    <span class="text-3xs font-semibold px-2 py-0.5 rounded-md bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border border-emerald-200 dark:border-emerald-500/30 leading-none">
                      {{ systemStatus.service_status || 'RUNNING' }}
                    </span>
                    <button
                      type="button"
                      @click="closeToolbarPopover"
                      class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Main Hardware Gauges -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 mb-2.5 space-y-2.5">
                  <!-- CPU -->
                  <div>
                    <div class="flex justify-between items-center text-xs mb-1">
                      <span class="font-medium text-slate-600 dark:text-slate-300">CPU Utilization</span>
                      <span class="font-bold text-slate-900 dark:text-white tabular-nums">{{ Math.round(systemStatus.cpu_percent || 0) }}%</span>
                    </div>
                    <div class="w-full h-2 rounded-full bg-slate-200/80 dark:bg-slate-800 overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all duration-300"
                        :class="systemStatus.cpu_percent > 85 ? 'bg-rose-500' : systemStatus.cpu_percent > 70 ? 'bg-amber-500' : 'bg-[#1685E5] dark:bg-sky-500'"
                        :style="{ width: Math.round(systemStatus.cpu_percent || 0) + '%' }"
                      ></div>
                    </div>
                  </div>
                  <!-- RAM -->
                  <div>
                    <div class="flex justify-between items-center text-xs mb-1">
                      <span class="font-medium text-slate-600 dark:text-slate-300">RAM Utilization</span>
                      <span class="font-bold text-slate-900 dark:text-white tabular-nums">{{ Math.round(systemStatus.ram_percent || 0) }}%</span>
                    </div>
                    <div class="w-full h-2 rounded-full bg-slate-200/80 dark:bg-slate-800 overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all duration-300"
                        :class="systemStatus.ram_percent > 90 ? 'bg-rose-500' : systemStatus.ram_percent > 75 ? 'bg-amber-500' : 'bg-indigo-500'"
                        :style="{ width: Math.round(systemStatus.ram_percent || 0) + '%' }"
                      ></div>
                    </div>
                  </div>
                  <!-- Disk -->
                  <div>
                    <div class="flex justify-between items-center text-xs mb-1">
                      <span class="font-medium text-slate-600 dark:text-slate-300">Disk Storage</span>
                      <span class="font-bold text-slate-900 dark:text-white tabular-nums">{{ Math.round(systemStatus.disk_percent || 0) }}%</span>
                    </div>
                    <div class="w-full h-2 rounded-full bg-slate-200/80 dark:bg-slate-800 overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all duration-300"
                        :class="systemStatus.disk_percent > 90 ? 'bg-rose-500' : 'bg-emerald-500'"
                        :style="{ width: Math.round(systemStatus.disk_percent || 0) + '%' }"
                      ></div>
                    </div>
                  </div>
                </div>

                <!-- Database & Health Details -->
                <div class="p-3 rounded-xl bg-slate-50/60 dark:bg-slate-900/40 border border-slate-200/75 dark:border-slate-800 mb-3 text-2xs space-y-1.5">
                  <div class="flex items-center justify-between">
                    <span class="text-slate-500 dark:text-slate-400 font-medium">Database Latency</span>
                    <span class="font-semibold text-[#1685E5] dark:text-blue-400 tabular-nums">{{ systemStatus.db_latency_ms != null ? systemStatus.db_latency_ms + ' ms' : '0.8 ms' }}</span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="text-slate-500 dark:text-slate-400 font-medium">WAL Queue Status</span>
                    <span class="font-semibold text-emerald-600 dark:text-emerald-400">{{ systemStatus.queue_status || 'HEALTHY' }}</span>
                  </div>
                  <div class="flex items-center justify-between">
                    <span class="text-slate-500 dark:text-slate-400 font-medium">Station Node</span>
                    <span class="font-semibold text-slate-800 dark:text-slate-200">{{ systemStatus.hostname || 'Edge-Node-01' }}</span>
                  </div>
                </div>

                <!-- Action Link -->
                <button
                  type="button"
                  @click="navigateToRoute('/monitoring/system')"
                  class="w-full py-2 px-3 rounded-xl text-xs font-semibold transition-colors border flex items-center justify-center space-x-1.5 bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA]/80 dark:border-blue-500/30 hover:bg-[#dceeff] dark:hover:bg-blue-600/25 shadow-xs"
                >
                  <span>⚡</span>
                  <span>Open Full System Diagnostics →</span>
                </button>
              </div>
          </div>

          <!-- 10. Network & Outbound Monitor (Icon Only + Status Dot) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('networkMonitor')"
              class="op-icon-btn"
              :class="{ 'is-active': activeToolbarPopover === 'networkMonitor' }"
              :title="networkMonitorTooltip"
              aria-label="Network & Outbound Delivery Monitor"
            >
              <OperatorDashboardIcon name="network" :size="16" />
              <span class="op-corner-dot" :class="networkStatusDotClass"></span>
            </button>
              <div
                v-if="activeToolbarPopover === 'networkMonitor'"
                class="absolute right-0 top-full mt-2 z-50 w-[400px] sm:w-[424px] max-w-[calc(100vw-24px)] rounded-2xl shadow-xl p-4 text-left border border-slate-200/90 dark:border-slate-800/80 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-sans text-xs max-h-[calc(100vh-160px)] overflow-y-auto custom-scroller"
                :style="{ maxHeight: bentoToolbarPopoverMaxHeight }"
                @click.stop
              >
                <!-- Popover Header -->
                <div class="flex items-center justify-between pb-3 mb-3 border-b border-slate-100 dark:border-slate-800/80">
                  <div class="flex items-center space-x-2.5">
                    <div class="w-8 h-8 rounded-xl flex items-center justify-center bg-[#EAF5FF] dark:bg-blue-600/15 border border-[#B8D4EA]/80 dark:border-blue-500/30 text-[#1685E5] dark:text-blue-400 flex-shrink-0">
                      <OperatorDashboardIcon name="network" :size="16" />
                    </div>
                    <div>
                      <h3 class="text-[13px] font-bold tracking-tight text-slate-900 dark:text-white leading-tight">
                        Network &amp; Delivery Monitor
                      </h3>
                      <div class="text-3xs font-medium text-slate-500 dark:text-slate-400 mt-0.5 leading-snug">
                        LAN · Internet WAN · Outbound Links
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1">
                    <button
                      type="button"
                      @click="fetchNetworkStatus"
                      class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-[#1685E5] dark:hover:text-blue-400 hover:bg-[#EAF5FF] dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                      :class="{ 'animate-spin': isFetchingNetwork }"
                      title="Refresh Network Diagnostics"
                    >
                      <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67" />
                      </svg>
                    </button>
                    <button
                      type="button"
                      @click="closeToolbarPopover"
                      class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Group 1: Local Network Status (LAN) -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 mb-2.5">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-3xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center space-x-1.5">
                      <span>🖥️</span>
                      <span>Local Network (LAN)</span>
                    </span>
                    <span
                      class="text-3xs font-semibold px-2 py-0.5 rounded-md border leading-none"
                      :class="networkStatus.local_network.status === 'UP' ? 'bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30' : 'bg-rose-50 dark:bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-200 dark:border-rose-500/30'"
                    >
                      {{ networkStatus.local_network.status === 'UP' ? 'CONNECTED' : 'DOWN' }}
                    </span>
                  </div>
                  <div class="grid grid-cols-2 gap-2.5 text-2xs">
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">Primary IP &amp; Interface</span>
                      <span class="font-semibold text-slate-900 dark:text-slate-100 tabular-nums">
                        {{ networkStatus.local_network.primary_ip || '-' }}
                        <span v-if="networkStatus.local_network.primary_interface" class="text-slate-500 dark:text-slate-400 font-normal text-3xs">
                          ({{ networkStatus.local_network.primary_interface }})
                        </span>
                      </span>
                    </div>
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">Default Gateway</span>
                      <span class="font-semibold text-slate-900 dark:text-slate-100 tabular-nums">
                        {{ networkStatus.local_network.gateway || '-' }}
                      </span>
                    </div>
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">Gateway Status</span>
                      <span
                        class="font-semibold flex items-center space-x-1.5 tabular-nums"
                        :class="networkStatus.local_network.gateway_reachable ? 'text-emerald-600 dark:text-emerald-400' : 'text-rose-600 dark:text-rose-400'"
                      >
                        <span class="inline-block w-1.5 h-1.5 rounded-full" :class="networkStatus.local_network.gateway_reachable ? 'bg-emerald-500' : 'bg-rose-500'"></span>
                        <span>{{ networkStatus.local_network.gateway_reachable ? 'Reachable (' + formatLatency(networkStatus.local_network.gateway_latency_ms) + ')' : 'Unreachable' }}</span>
                      </span>
                    </div>
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">Active Interfaces</span>
                      <span class="font-semibold text-slate-700 dark:text-slate-300 tabular-nums">
                        {{ (networkStatus.local_network.interfaces || []).filter(i => i.is_up && i.type !== 'LOOPBACK').length }} interface(s) UP
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Group 2: Internet Connectivity Status (WAN) -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 mb-2.5">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-3xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center space-x-1.5">
                      <span>🌐</span>
                      <span>Internet Connectivity (WAN)</span>
                    </span>
                    <span
                      class="text-3xs font-semibold px-2 py-0.5 rounded-md border leading-none"
                      :class="networkStatus.internet.status === 'CONNECTED' ? 'bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30' : networkStatus.internet.status === 'DEGRADED' ? 'bg-amber-50 dark:bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-500/30' : 'bg-rose-50 dark:bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-200 dark:border-rose-500/30'"
                    >
                      {{ networkStatus.internet.status || 'UNKNOWN' }}
                    </span>
                  </div>
                  <div class="grid grid-cols-2 gap-2.5 text-2xs">
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">Probe Target</span>
                      <span class="font-semibold text-slate-800 dark:text-slate-200 tabular-nums">{{ networkStatus.internet.probe_target || '1.1.1.1:53' }}</span>
                    </div>
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">External Latency</span>
                      <span
                        class="font-semibold tabular-nums"
                        :class="networkStatus.internet.connected ? 'text-[#1685E5] dark:text-blue-400' : 'text-slate-400'"
                      >
                        {{ networkStatus.internet.connected ? formatLatency(networkStatus.internet.latency_ms) : 'Timeout' }}
                      </span>
                    </div>
                  </div>
                  <div v-if="!networkStatus.internet.connected" class="mt-2 pt-2 border-t border-slate-200/60 dark:border-slate-800 text-3xs font-medium text-amber-700 dark:text-amber-400">
                    ⚠️ WAN Offline: Data safely queued locally in autonomous offline storage.
                  </div>
                </div>

                <!-- Group 3: Outbound Delivery Monitor -->
                <div class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 mb-3">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-3xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400 flex items-center space-x-1.5">
                      <span>📤</span>
                      <span>Outbound Delivery Links ({{ networkStatus.outbound.total_links }})</span>
                    </span>
                    <span
                      class="text-3xs font-semibold px-2 py-0.5 rounded-md border leading-none"
                      :class="outboundStatusBadgeClass(networkStatus.outbound.overall_status)"
                    >
                      {{ networkStatus.outbound.overall_status || 'UNKNOWN' }}
                    </span>
                  </div>

                  <!-- Queue Summary Counters Bar -->
                  <div class="px-2.5 py-2 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800 mb-2.5 flex items-center justify-between text-3xs tabular-nums">
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 font-medium">Queue: </span>
                      <strong :class="((networkStatus.outbound.queue_summary.pending || 0) + (networkStatus.outbound.queue_summary.sending || 0) + (networkStatus.outbound.queue_summary.retrying || 0)) > 0 ? 'text-[#1685E5] dark:text-blue-400 font-bold' : 'text-slate-700 dark:text-slate-300 font-semibold'">
                        {{ (networkStatus.outbound.queue_summary.pending || 0) + (networkStatus.outbound.queue_summary.sending || 0) + (networkStatus.outbound.queue_summary.retrying || 0) }}
                      </strong>
                    </div>
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 font-medium">Delivered: </span>
                      <strong class="text-emerald-600 dark:text-emerald-400 font-bold">
                        {{ (networkStatus.outbound.queue_summary.delivered || 0) + (networkStatus.outbound.queue_summary.acknowledged || 0) }}
                      </strong>
                    </div>
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 font-medium">Retrying: </span>
                      <strong :class="(networkStatus.outbound.queue_summary.retrying || 0) > 0 ? 'text-amber-600 dark:text-amber-400 font-bold' : 'text-slate-500 dark:text-slate-400 font-semibold'">
                        {{ networkStatus.outbound.queue_summary.retrying || 0 }}
                      </strong>
                    </div>
                    <div>
                      <span class="text-slate-500 dark:text-slate-400 font-medium">Failed: </span>
                      <strong :class="(networkStatus.outbound.queue_summary.failed || 0) > 0 ? 'text-rose-600 dark:text-rose-400 font-bold' : 'text-slate-500 dark:text-slate-400 font-semibold'">
                        {{ networkStatus.outbound.queue_summary.failed || 0 }}
                      </strong>
                    </div>
                  </div>

                  <!-- Destinations List -->
                  <div v-if="networkStatus.outbound.destinations && networkStatus.outbound.destinations.length" class="space-y-2 max-h-48 overflow-y-auto custom-scroller pr-0.5">
                    <div
                      v-for="dest in networkStatus.outbound.destinations"
                      :key="dest.id"
                      class="p-2.5 rounded-xl bg-white dark:bg-[#0B0F19] border border-slate-200/90 dark:border-slate-800 space-y-1.5 text-2xs"
                    >
                      <div class="flex items-center justify-between">
                        <div class="flex items-center space-x-2 min-w-0">
                          <strong class="text-xs font-semibold text-slate-900 dark:text-white truncate">{{ dest.name }}</strong>
                          <span class="text-3xs font-medium px-1.5 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200/80 dark:border-slate-700 leading-none">
                            {{ dest.type }}
                          </span>
                        </div>
                        <span
                          class="text-3xs font-semibold px-2 py-0.5 rounded-md border leading-none flex-shrink-0"
                          :class="outboundStatusBadgeClass(dest.status)"
                        >
                          {{ dest.status }}
                        </span>
                      </div>

                      <div class="text-3xs text-slate-500 dark:text-slate-400 truncate" :title="dest.endpoint">
                        {{ dest.endpoint }}
                      </div>

                      <div class="flex items-center justify-between text-3xs pt-1.5 border-t border-slate-100 dark:border-slate-800/80 tabular-nums">
                        <span class="text-slate-600 dark:text-slate-300">
                          Success: <strong class="text-emerald-600 dark:text-emerald-400 font-semibold">{{ (dest.success_deliveries || 0).toLocaleString() }}</strong> ·
                          Fail: <strong class="text-rose-500 font-semibold">{{ (dest.failed_deliveries || 0).toLocaleString() }}</strong>
                          <span class="text-slate-400 ml-1">({{ (dest.success_rate || 0).toFixed(1) }}%)</span>
                        </span>
                        <span class="text-slate-500 dark:text-slate-400">
                          Queue: <strong class="text-slate-800 dark:text-slate-200 font-semibold">{{ dest.queue_count || 0 }}</strong>
                        </span>
                      </div>

                      <div v-if="dest.last_error" class="text-3xs text-rose-600 dark:text-rose-400 truncate bg-rose-50 dark:bg-rose-500/10 px-2 py-1 rounded-lg border border-rose-200 dark:border-rose-500/20" :title="dest.last_error">
                        ⚠️ {{ dest.last_error }}
                      </div>
                    </div>
                  </div>

                  <div v-else class="text-center py-3 text-2xs text-slate-400 italic">
                    No outbound destinations configured.
                  </div>
                </div>

                <!-- Action Button: Open Outbound Integration / Delivery Queue -->
                <button
                  type="button"
                  @click="navigateToRoute(canAccessOperationalReport ? '/integration' : '/integration/queue')"
                  class="w-full py-2 px-3 rounded-xl text-xs font-semibold transition-colors border flex items-center justify-center space-x-1.5 bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA]/80 dark:border-blue-500/30 hover:bg-[#dceeff] dark:hover:bg-blue-600/25 shadow-xs"
                >
                  <span>🔗</span>
                  <span>{{ canAccessOperationalReport ? 'Open Outbound Integration & Delivery (/integration) →' : 'Open Delivery Queue (/integration/queue) →' }}</span>
                </button>
              </div>
          </div>

          <!-- 11. Station GPS Location (Icon Only + Status Dot) -->
          <div class="relative toolbar-popover-anchor">
            <button
              type="button"
              @click.stop="toggleToolbarPopover('stationLocation')"
              class="op-icon-btn"
              :class="{ 'is-active': activeToolbarPopover === 'stationLocation' }"
              :title="stationLocationTooltip"
              aria-label="Station Location Monitor"
            >
              <OperatorDashboardIcon :name="stationLocation.is_configured ? 'map-pinned' : 'map-pin'" :size="16" />
              <span class="op-corner-dot" :class="stationLocationStatusDotClass"></span>
            </button>
              <div
                v-if="activeToolbarPopover === 'stationLocation'"
                class="absolute right-0 top-full mt-2 z-50 w-[420px] sm:w-[440px] max-w-[calc(100vw-24px)] rounded-2xl shadow-2xl text-left border border-slate-200/90 dark:border-slate-800/80 bg-white dark:bg-[#0B0F19] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-sans text-xs max-h-[calc(100vh-160px)] flex flex-col overflow-hidden"
                :style="{ maxHeight: bentoToolbarPopoverMaxHeight }"
                @click.stop
              >
                <!-- Pinned Popover Header (Never Clipped on Scroll) -->
                <div class="flex items-center justify-between px-4 py-2.5 border-b border-slate-100 dark:border-slate-800/80 bg-white dark:bg-[#0B0F19] shrink-0">
                  <div class="flex items-center space-x-2.5">
                    <div class="w-8 h-8 rounded-xl flex items-center justify-center bg-[#EAF5FF] dark:bg-blue-600/15 border border-[#B8D4EA]/80 dark:border-blue-500/30 text-[#1685E5] dark:text-blue-400 flex-shrink-0">
                      <OperatorDashboardIcon name="map-pin" :size="16" />
                    </div>
                    <div>
                      <h3 class="text-[13px] font-bold tracking-tight text-slate-900 dark:text-white leading-tight">
                        {{ $t('operator.stationLocation') || 'Station Location Monitor' }}
                      </h3>
                      <div class="text-3xs font-medium text-slate-500 dark:text-slate-400 mt-0.5 leading-snug">
                        {{ $t('operator.stationLocationSub') || 'Physical Geolocation & Coordinates' }}
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1">
                    <button
                      type="button"
                      @click="openFullScreenMap"
                      class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-[#1685E5] dark:hover:text-blue-400 hover:bg-[#EAF5FF] dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                      title="Expand to Full Screen Map"
                    >
                      <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <polyline points="15 3 21 3 21 9" />
                        <polyline points="9 21 3 21 3 15" />
                        <line x1="21" y1="3" x2="14" y2="10" />
                        <line x1="3" y1="21" x2="10" y2="14" />
                      </svg>
                    </button>
                    <button
                      type="button"
                      @click="fetchStationLocation"
                      class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-[#1685E5] dark:hover:text-blue-400 hover:bg-[#EAF5FF] dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                      :class="{ 'animate-spin': isFetchingStationLocation }"
                      title="Refresh Location"
                    >
                      <svg class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M21.5 2v6h-6M21.34 15.57a10 10 0 1 1-.57-8.38l5.67-5.67" />
                      </svg>
                    </button>
                    <button
                      type="button"
                      @click="closeToolbarPopover"
                      class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800/60 rounded-lg transition-colors"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Scrollable Popover Content Area -->
                <div class="flex-1 min-h-0 overflow-y-auto px-4 py-3 custom-scroller space-y-2.5">
                  <!-- Feedback Banner -->
                  <div
                    v-if="stationLocationFeedback.message"
                    class="p-2.5 rounded-xl text-2xs font-medium border flex items-center justify-between"
                    :class="stationLocationFeedback.type === 'error' ? 'bg-rose-50 dark:bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-200 dark:border-rose-500/30' : 'bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30'"
                  >
                    <span class="leading-snug pr-2">{{ stationLocationFeedback.message }}</span>
                    <button @click="stationLocationFeedback.message = ''" class="ml-1 opacity-70 hover:opacity-100 font-bold shrink-0">✕</button>
                  </div>

                  <!-- Station Summary Status Card -->
                  <div class="p-2.5 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800">
                    <div class="flex items-center justify-between gap-2 mb-1.5">
                      <span class="text-2xs font-bold uppercase tracking-wider text-slate-800 dark:text-slate-200 truncate">
                        {{ stationLocation.station_name || systemStatus.hostname || 'Edge Station Node' }}
                      </span>
                      <div class="flex items-center space-x-1.5 shrink-0">
                        <span
                          v-if="stationLocation.location_source === 'MODBUS_RTU'"
                          class="text-3xs font-semibold px-2 py-0.5 rounded-md border leading-none flex items-center space-x-1"
                          :class="modbusRtuGpsStatus.badgeClass"
                        >
                          <span class="w-1.5 h-1.5 rounded-full" :class="modbusRtuGpsStatus.dotClass"></span>
                          <span>{{ modbusRtuGpsStatus.badgeLabel }}</span>
                        </span>
                        <span
                          class="text-3xs font-semibold px-2 py-0.5 rounded-md border leading-none"
                          :class="stationLocation.is_configured ? 'bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30' : 'bg-amber-50 dark:bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-500/30'"
                        >
                          {{ stationLocation.is_configured ? 'CONFIGURED' : 'UNCONFIGURED' }}
                        </span>
                      </div>
                    </div>

                    <div class="grid grid-cols-2 gap-2 text-2xs">
                      <div>
                        <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">Coordinates (Lat, Lng)</span>
                        <span class="text-xs font-bold tabular-nums text-slate-900 dark:text-slate-100">
                          {{ stationLocation.is_configured && stationLocation.latitude !== null ? `${Number(stationLocation.latitude).toFixed(6)}, ${Number(stationLocation.longitude).toFixed(6)}` : 'Not Configured' }}
                        </span>
                      </div>
                      <div>
                        <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">Location Source</span>
                        <span class="text-xs font-semibold text-[#1685E5] dark:text-blue-400">
                          {{ formatLocationSourceLabel(stationLocation.location_source) }}
                          <span v-if="stationLocation.accuracy_meters" class="text-slate-500 dark:text-slate-400 text-3xs font-normal tabular-nums">
                            (±{{ stationLocation.accuracy_meters }}m)
                          </span>
                        </span>
                      </div>
                      <div v-if="stationLocation.address" class="col-span-2">
                        <span class="text-slate-500 dark:text-slate-400 block text-3xs font-medium mb-0.5">Installation Site / Address</span>
                        <span class="text-2xs text-slate-700 dark:text-slate-300 leading-snug line-clamp-1">
                          {{ stationLocation.address }}
                        </span>
                      </div>
                      <div class="col-span-2 flex items-center justify-between text-3xs text-slate-500 dark:text-slate-400 pt-1.5 border-t border-slate-200/60 dark:border-slate-800 tabular-nums">
                        <span>Updated: {{ formatLocationDate(stationLocation.updated_at) }}</span>
                        <span>By: <strong class="text-slate-700 dark:text-slate-300 font-semibold">{{ stationLocation.updated_by || 'system' }}</strong></span>
                      </div>
                    </div>
                  </div>

                  <!-- Mini Map Section -->
                  <div>
                    <div class="flex items-center justify-between mb-1.5">
                      <span class="text-3xs font-semibold text-slate-500 dark:text-slate-400 uppercase tracking-wider">
                        Interactive Station Map
                      </span>
                      <button
                        type="button"
                        @click="openFullScreenMap"
                        class="text-3xs font-semibold text-[#1685E5] dark:text-blue-400 hover:underline flex items-center space-x-1"
                      >
                        <span>Full Screen</span>
                        <svg class="w-3 h-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <polyline points="15 3 21 3 21 9" />
                          <polyline points="9 21 3 21 3 15" />
                          <line x1="21" y1="3" x2="14" y2="10" />
                          <line x1="3" y1="21" x2="10" y2="14" />
                        </svg>
                      </button>
                    </div>
                    <StationMapViewer
                      ref="miniMapViewer"
                      :latitude="candidateLocation.latitude !== null ? Number(candidateLocation.latitude) : (stationLocation.latitude !== null ? Number(stationLocation.latitude) : null)"
                      :longitude="candidateLocation.longitude !== null ? Number(candidateLocation.longitude) : (stationLocation.longitude !== null ? Number(stationLocation.longitude) : null)"
                      :accuracy-meters="candidateLocation.accuracy_meters !== null ? Number(candidateLocation.accuracy_meters) : (stationLocation.accuracy_meters !== null ? Number(stationLocation.accuracy_meters) : null)"
                      :editable="true"
                      :source="candidateLocation.location_source || stationLocation.location_source"
                      height="148px"
                      @coordinate-change="onMapCoordinatePicked"
                      @location-selected="onLocationSelected"
                    />
                    <div class="text-3xs text-slate-500 dark:text-slate-400 mt-1 flex items-center space-x-1.5 leading-snug">
                      <svg class="w-3.5 h-3.5 text-[#1685E5] dark:text-blue-400 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="12" cy="12" r="10" />
                        <line x1="12" y1="16" x2="12" y2="12" />
                        <line x1="12" y1="8" x2="12.01" y2="8" />
                      </svg>
                      <span>Click on the map, use search box, or drag the marker to pinpoint exact station coordinates.</span>
                    </div>
                  </div>

                  <!-- Three Location Sources Tabs: Manual, Internet IP, Modbus RTU -->
                  <div class="border-t border-slate-100 dark:border-slate-800/80 pt-2.5">
                    <div class="grid grid-cols-3 gap-1 p-1 rounded-xl bg-slate-100/90 dark:bg-slate-900 border border-slate-200/80 dark:border-slate-800 mb-2.5 text-2xs font-semibold">
                      <button
                        type="button"
                        @click="locationActiveTab = 'manual'"
                        class="py-1.5 px-2 rounded-lg transition-colors flex items-center justify-center space-x-1.5 cursor-pointer"
                        :class="locationActiveTab === 'manual' ? 'bg-[#1685E5] text-white btn-text shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-white/60 dark:hover:bg-slate-800/60'"
                      >
                        <svg class="w-3.5 h-3.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <path d="M12 20h9" />
                          <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z" />
                        </svg>
                        <span>Manual</span>
                      </button>
                      <button
                        type="button"
                        @click="locationActiveTab = 'ip'"
                        class="py-1.5 px-2 rounded-lg transition-colors flex items-center justify-center space-x-1.5 cursor-pointer"
                        :class="locationActiveTab === 'ip' ? 'bg-[#1685E5] text-white btn-text shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-white/60 dark:hover:bg-slate-800/60'"
                      >
                        <svg class="w-3.5 h-3.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <circle cx="12" cy="12" r="10" />
                          <line x1="2" y1="12" x2="22" y2="12" />
                          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                        </svg>
                        <span>Internet IP</span>
                      </button>
                      <button
                        type="button"
                        @click="locationActiveTab = 'modbus_rtu'"
                        class="py-1.5 px-2 rounded-lg transition-colors flex items-center justify-center space-x-1.5 cursor-pointer"
                        :class="locationActiveTab === 'modbus_rtu' ? 'bg-[#1685E5] text-white btn-text shadow-sm' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-white/60 dark:hover:bg-slate-800/60'"
                      >
                        <svg class="w-3.5 h-3.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9" />
                          <path d="M7.8 16.2c-2.3-2.3-2.3-6.1 0-8.5" />
                          <circle cx="12" cy="12" r="2" />
                          <path d="M16.2 7.8c2.3 2.3 2.3 6.1 0 8.5" />
                          <path d="M19.1 4.9C23 8.8 23 15.1 19.1 19" />
                        </svg>
                        <span>Modbus RTU</span>
                      </button>
                    </div>

                    <!-- Tab 1: Manual Input & Map Pin -->
                    <div v-if="locationActiveTab === 'manual'" class="space-y-2">
                      <!-- Quick Location Search Bar in Tab -->
                      <div class="p-2.5 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800">
                        <label class="text-3xs font-semibold text-slate-600 dark:text-slate-400 mb-1.5 flex items-center space-x-1.5">
                          <svg class="w-3.5 h-3.5 text-[#1685E5] dark:text-blue-400 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="11" cy="11" r="8" />
                            <line x1="21" y1="21" x2="16.65" y2="16.65" />
                          </svg>
                          <span>{{ $t('operator.quickLocationSearch') || 'Search Location (City, Address, or Coordinates)' }}</span>
                        </label>
                        <div class="flex items-center space-x-1.5">
                          <input
                            v-model="tabLocationSearchQuery"
                            type="text"
                            :placeholder="$t('operator.searchLocationPlaceholder') || 'Search city, address, or lat, lng... (Enter)'"
                            @keydown.enter.prevent="executeTabLocationSearch"
                            class="flex-1 px-2.5 py-1.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/90 dark:border-slate-800 text-slate-800 dark:text-slate-200 text-xs font-sans placeholder:text-slate-400 focus:outline-none focus:border-[#1685E5]"
                          />
                          <button
                            type="button"
                            @click="executeTabLocationSearch"
                            :disabled="isTabSearching"
                            class="px-3 py-1.5 rounded-lg text-xs font-semibold bg-[#1685E5] hover:bg-[#1273C9] text-white btn-text transition-colors flex items-center space-x-1 disabled:opacity-50 shrink-0"
                          >
                            <span v-if="isTabSearching" class="animate-spin inline-block w-3 h-3 border-2 border-white border-t-transparent rounded-full"></span>
                            <span v-else>Cari</span>
                          </button>
                        </div>
                      </div>

                      <div class="grid grid-cols-2 gap-2">
                        <div>
                          <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 mb-1">Latitude (-90 to 90)</label>
                          <input
                            v-model.number="candidateLocation.latitude"
                            type="number"
                            step="0.000001"
                            min="-90"
                            max="90"
                            placeholder="-6.208800"
                            class="w-full px-3 py-1.5 rounded-xl bg-slate-50/80 dark:bg-slate-900 border border-slate-200/90 dark:border-slate-800 text-slate-900 dark:text-slate-100 text-xs font-sans font-medium tabular-nums focus:outline-none focus:border-[#1685E5]"
                          />
                        </div>
                        <div>
                          <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 mb-1">Longitude (-180 to 180)</label>
                          <input
                            v-model.number="candidateLocation.longitude"
                            type="number"
                            step="0.000001"
                            min="-180"
                            max="180"
                            placeholder="106.845600"
                            class="w-full px-3 py-1.5 rounded-xl bg-slate-50/80 dark:bg-slate-900 border border-slate-200/90 dark:border-slate-800 text-slate-900 dark:text-slate-100 text-xs font-sans font-medium tabular-nums focus:outline-none focus:border-[#1685E5]"
                          />
                        </div>
                        <div class="col-span-2">
                          <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 mb-1">Site Name / Description (Optional)</label>
                          <input
                            v-model="candidateLocation.station_name"
                            type="text"
                            placeholder="Ambient Monitoring Station"
                            class="w-full px-3 py-1.5 rounded-xl bg-slate-50/80 dark:bg-slate-900 border border-slate-200/90 dark:border-slate-800 text-slate-900 dark:text-slate-100 text-xs font-sans focus:outline-none focus:border-[#1685E5]"
                          />
                        </div>
                      </div>
                    </div>

                    <!-- Tab 2: Internet Geolocation (IP-based) -->
                    <div v-else-if="locationActiveTab === 'ip'" class="space-y-2.5 text-2xs">
                      <div class="p-3 rounded-xl bg-amber-50/80 dark:bg-amber-500/10 border border-amber-200/80 dark:border-amber-500/25 space-y-1">
                        <div class="font-semibold text-amber-700 dark:text-amber-400 flex items-center space-x-1.5 text-xs">
                          <svg class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="12" cy="12" r="10" />
                            <line x1="12" y1="8" x2="12" y2="12" />
                            <line x1="12" y1="16" x2="12.01" y2="16" />
                          </svg>
                          <span>Internet IP Geolocation (Estimasi Area)</span>
                        </div>
                        <p class="text-3xs text-slate-600 dark:text-slate-400 leading-relaxed">
                          Lokasi berbasis IP publik merupakan estimasi level kota/ISP dan tidak seakurat modul GPS fisik. Sesuai kebijakan privasi edge, IP stasiun tidak dikirim ke pihak ketiga tanpa konfigurasi provider terverifikasi.
                        </p>
                      </div>

                      <button
                        type="button"
                        @click="checkIPGeolocationProvider"
                        :disabled="ipGeoState.isDetecting"
                        class="w-full py-2 px-3 rounded-xl text-xs font-semibold transition-colors border flex items-center justify-center space-x-1.5 bg-[#EAF5FF] dark:bg-blue-600/15 hover:bg-[#dceeff] dark:hover:bg-blue-600/25 text-[#1685E5] dark:text-blue-400 border-[#B8D4EA]/80 dark:border-blue-500/30"
                      >
                        <span v-if="ipGeoState.isDetecting" class="animate-spin inline-block w-3.5 h-3.5 border-2 border-[#1685E5] border-t-transparent rounded-full"></span>
                        <svg v-else class="w-3.5 h-3.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                          <circle cx="12" cy="12" r="10" />
                          <line x1="2" y1="12" x2="22" y2="12" />
                          <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
                        </svg>
                        <span>Periksa Provider IP Geolocation</span>
                      </button>

                      <div v-if="ipGeoState.status !== 'IDLE'" class="p-3 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 space-y-1">
                        <div class="flex items-center justify-between text-2xs">
                          <span class="text-slate-500 dark:text-slate-400 font-medium">Status Provider:</span>
                          <span class="font-bold text-amber-600 dark:text-amber-400">{{ ipGeoState.status }}</span>
                        </div>
                        <div class="text-3xs text-slate-500 dark:text-slate-400 leading-relaxed">
                          {{ ipGeoState.message || 'Provider IP Geolocation belum dikonfigurasi. Gunakan tab Manual atau GPS Modbus RTU.' }}
                        </div>
                      </div>
                    </div>

                    <!-- Tab 3: GPS Data via Modbus RTU -->
                    <div v-else-if="locationActiveTab === 'modbus_rtu'" class="space-y-2 text-2xs">
                      <!-- Live Modbus RTU Telemetry Status Card -->
                      <div class="p-2.5 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 space-y-2">
                        <div class="flex items-center justify-between">
                          <div class="flex items-center space-x-1.5">
                            <svg class="w-4 h-4 text-[#1685E5] dark:text-blue-400 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                              <path d="M4.9 19.1C1 15.2 1 8.8 4.9 4.9" />
                              <path d="M7.8 16.2c-2.3-2.3-2.3-6.1 0-8.5" />
                              <circle cx="12" cy="12" r="2" />
                              <path d="M16.2 7.8c2.3 2.3 2.3 6.1 0 8.5" />
                              <path d="M19.1 4.9C23 8.8 23 15.1 19.1 19" />
                            </svg>
                            <span class="text-xs font-bold text-slate-900 dark:text-white">GPS via Modbus RTU</span>
                          </div>
                          <span
                            class="text-3xs font-semibold px-2 py-0.5 rounded-md border leading-none flex items-center space-x-1"
                            :class="modbusRtuGpsStatus.badgeClass"
                          >
                            <span class="w-1.5 h-1.5 rounded-full" :class="modbusRtuGpsStatus.dotClass"></span>
                            <span>{{ modbusRtuGpsStatus.badgeLabel }}</span>
                          </span>
                        </div>

                        <div class="grid grid-cols-2 gap-2 pt-1.5 border-t border-slate-200/60 dark:border-slate-800">
                          <div class="p-2 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800">
                            <span class="block text-3xs font-medium text-slate-500 dark:text-slate-400">Live Latitude</span>
                            <span class="text-xs font-bold tabular-nums text-slate-900 dark:text-slate-100">
                              {{ modbusRtuGpsStatus.lat !== null ? modbusRtuGpsStatus.lat.toFixed(6) : '—' }}
                            </span>
                          </div>
                          <div class="p-2 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800">
                            <span class="block text-3xs font-medium text-slate-500 dark:text-slate-400">Live Longitude</span>
                            <span class="text-xs font-bold tabular-nums text-slate-900 dark:text-slate-100">
                              {{ modbusRtuGpsStatus.lng !== null ? modbusRtuGpsStatus.lng.toFixed(6) : '—' }}
                            </span>
                          </div>
                        </div>

                        <div class="flex items-center justify-between text-3xs text-slate-500 dark:text-slate-400 pt-0.5">
                          <span>{{ modbusRtuGpsStatus.reason }}</span>
                          <span v-if="modbusRtuGpsStatus.satellites !== null" class="font-semibold text-slate-700 dark:text-slate-300 tabular-nums">
                            Satelit / Fix: {{ modbusRtuGpsStatus.satellites }}
                          </span>
                        </div>
                      </div>

                      <!-- Single-Source GPS Module Card + Direct Parameter Setup -->
                      <div class="p-2.5 rounded-xl bg-slate-50/80 dark:bg-slate-900/60 border border-slate-200/80 dark:border-slate-800 space-y-2">
                        <div class="flex items-center justify-between gap-2">
                          <span class="text-3xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                            Sumber Modul GPS
                          </span>
                          <div class="flex items-center space-x-1.5">
                            <span v-if="effectiveModbusGpsMapping.isAutoDetected" class="text-4xs font-semibold px-1.5 py-0.5 rounded bg-[#EAF5FF] dark:bg-blue-600/15 text-[#1685E5] dark:text-blue-400">
                              Auto-Paired
                            </span>
                            <button
                              type="button"
                              @click="showAdvancedGpsRegisterMap = !showAdvancedGpsRegisterMap"
                              class="text-3xs font-semibold text-[#1685E5] dark:text-blue-400 hover:underline"
                            >
                              {{ showAdvancedGpsRegisterMap ? 'Tutup Detail' : 'Detail Register' }}
                            </button>
                          </div>
                        </div>

                        <!-- Single Source Selector (1 Modul GPS) for Admin / Engineer -->
                        <div v-if="canManageCategories">
                          <select
                            v-model="gpsModbusMapping.deviceId"
                            @change="onModbusGpsDeviceChange"
                            class="w-full px-2.5 py-1.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/90 dark:border-slate-800 text-slate-800 dark:text-slate-200 text-2xs font-sans focus:outline-none focus:border-[#1685E5]"
                          >
                            <option value="">Otomatis Deteksi Modul GPS (Semua Port RTU)</option>
                            <option v-for="dev in modbusRtuDevicesList" :key="'gps-dev-' + dev.id" :value="String(dev.id)">
                              {{ formatModbusGpsDeviceOption(dev) }}
                            </option>
                          </select>
                        </div>

                        <!-- Paired Register Readout / Empty-State Setup Helper -->
                        <div class="p-2 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800 space-y-1.5">
                          <div v-if="effectiveModbusGpsMapping.latParam && effectiveModbusGpsMapping.lngParam" class="flex items-center justify-between gap-2 text-3xs">
                            <div class="min-w-0 truncate text-slate-600 dark:text-slate-300">
                              <span class="text-slate-400">Register: </span>
                              <strong class="font-semibold text-slate-800 dark:text-slate-100">{{ effectiveModbusGpsMapping.latParam.param_name }}</strong>
                              <span class="text-slate-400"> &amp; </span>
                              <strong class="font-semibold text-slate-800 dark:text-slate-100">{{ effectiveModbusGpsMapping.lngParam.param_name }}</strong>
                            </div>
                            <button
                              v-if="canManageCategories"
                              type="button"
                              @click="openEditGpsParamModal(effectiveModbusGpsMapping.latParam)"
                              class="text-3xs font-semibold text-[#1685E5] dark:text-blue-400 hover:underline shrink-0"
                            >
                              Edit
                            </button>
                          </div>
                          <div v-else class="text-3xs text-slate-500 dark:text-slate-400 leading-relaxed">
                            Modul GPS Modbus RTU mengirim angka Latitude &amp; Longitude pada 2 register berbeda dalam 1 perangkat. Tambahkan parameter <strong class="text-slate-700 dark:text-slate-300">gps_lat</strong> &amp; <strong class="text-slate-700 dark:text-slate-300">gps_lng</strong> untuk mulai membaca koordinat.
                          </div>

                          <!-- Direct Parameter Setup Shortcut Buttons (Admin / Engineer) -->
                          <div v-if="canManageCategories" class="flex items-center gap-1.5 pt-1 border-t border-slate-100 dark:border-slate-800/80">
                            <button
                              type="button"
                              @click="openAddGpsParamModal"
                              class="flex-1 py-1.5 px-2.5 rounded-lg text-3xs font-semibold bg-[#EAF5FF] dark:bg-blue-600/15 hover:bg-[#dceeff] dark:hover:bg-blue-600/25 text-[#1685E5] dark:text-blue-400 border border-[#B8D4EA]/80 dark:border-blue-500/30 transition-colors flex items-center justify-center space-x-1"
                            >
                              <span>+ Tambah Parameter GPS</span>
                            </button>
                            <button
                              type="button"
                              @click="openDevicesPageForGps"
                              class="py-1.5 px-2.5 rounded-lg text-3xs font-semibold bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700 transition-colors shrink-0"
                            >
                              Perangkat &amp; Sensor →
                            </button>
                          </div>
                        </div>

                        <!-- Collapsible Manual Register Override (Only when custom parameter names are used) -->
                        <div v-if="showAdvancedGpsRegisterMap" class="grid grid-cols-2 gap-2 pt-1">
                          <div>
                            <label class="block text-3xs font-medium text-slate-600 dark:text-slate-400 mb-1">Register Latitude</label>
                            <select
                              v-model="gpsModbusMapping.latParamId"
                              @change="saveModbusGpsMapping"
                              class="w-full px-2 py-1.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/90 dark:border-slate-800 text-slate-800 dark:text-slate-200 text-2xs font-sans focus:outline-none focus:border-[#1685E5]"
                            >
                              <option value="">{{ effectiveModbusGpsMapping.latParam ? `Auto (${effectiveModbusGpsMapping.latParam.param_name})` : 'Otomatis...' }}</option>
                              <option v-for="p in modbusRtuCandidateParams" :key="'gps-lat-' + p.param_id" :value="String(p.param_id)">
                                {{ formatModbusGpsParamOption(p) }}
                              </option>
                            </select>
                          </div>
                          <div>
                            <label class="block text-3xs font-medium text-slate-600 dark:text-slate-400 mb-1">Register Longitude</label>
                            <select
                              v-model="gpsModbusMapping.lngParamId"
                              @change="saveModbusGpsMapping"
                              class="w-full px-2 py-1.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200/90 dark:border-slate-800 text-slate-800 dark:text-slate-200 text-2xs font-sans focus:outline-none focus:border-[#1685E5]"
                            >
                              <option value="">{{ effectiveModbusGpsMapping.lngParam ? `Auto (${effectiveModbusGpsMapping.lngParam.param_name})` : 'Otomatis...' }}</option>
                              <option v-for="p in modbusRtuCandidateParams" :key="'gps-lng-' + p.param_id" :value="String(p.param_id)">
                                {{ formatModbusGpsParamOption(p) }}
                              </option>
                            </select>
                          </div>
                        </div>

                        <div class="flex items-center justify-between pt-1.5 border-t border-slate-200/60 dark:border-slate-800">
                          <div class="pr-2">
                            <span class="block text-2xs font-semibold text-slate-800 dark:text-slate-200">Auto-Sync Stasiun Bergerak</span>
                            <span class="block text-3xs text-slate-500 dark:text-slate-400">Perbarui otomatis jika bergeser &gt;25m &amp; sinyal GPS Locked</span>
                          </div>
                          <button
                            type="button"
                            role="switch"
                            :aria-checked="gpsModbusMapping.autoSync ? 'true' : 'false'"
                            @click="toggleModbusGpsAutoSync"
                            class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                            :class="gpsModbusMapping.autoSync ? 'bg-[#1685E5] dark:bg-blue-500' : 'bg-slate-200 dark:bg-slate-700'"
                          >
                            <span
                              class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                              :class="gpsModbusMapping.autoSync ? 'translate-x-4' : 'translate-x-0'"
                            ></span>
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Pinned Bottom Action Bar (Always Framed, Never Clipped at Bottom) -->
                <div
                  v-if="locationActiveTab === 'manual' || locationActiveTab === 'modbus_rtu'"
                  class="px-4 py-2.5 border-t border-slate-200/80 dark:border-slate-800 bg-slate-50/95 dark:bg-slate-900/90 shrink-0"
                >
                  <div v-if="locationActiveTab === 'manual'" class="flex items-center space-x-2">
                    <button
                      type="button"
                      @click="saveStationLocation"
                      :disabled="isSavingStationLocation || candidateLocation.latitude == null || candidateLocation.longitude == null"
                      class="flex-1 py-2 px-3 rounded-xl text-xs font-semibold transition-colors border flex items-center justify-center space-x-1.5 bg-[#1685E5] hover:bg-[#1273C9] text-white btn-text border-[#1685E5] shadow-sm disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <span v-if="isSavingStationLocation" class="animate-spin inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full"></span>
                      <svg v-else class="w-3.5 h-3.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
                        <polyline points="17 21 17 13 7 13 7 21" />
                        <polyline points="7 3 7 8 15 8" />
                      </svg>
                      <span>Save Location</span>
                    </button>
                    <button
                      type="button"
                      @click="cancelLocationEdit"
                      class="py-2 px-3 rounded-xl text-xs font-semibold transition-colors border bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-700"
                      title="Revert to saved location"
                    >
                      Cancel
                    </button>
                    <button
                      type="button"
                      @click="resetStationLocation"
                      class="py-2 px-3 rounded-xl text-xs font-semibold transition-colors border bg-rose-50 dark:bg-rose-500/10 hover:bg-rose-100 dark:hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 border-rose-200 dark:border-rose-500/30"
                      title="Reset station coordinates to unconfigured state"
                    >
                      Reset
                    </button>
                  </div>

                  <div v-else-if="locationActiveTab === 'modbus_rtu'" class="flex items-center space-x-2">
                    <button
                      type="button"
                      @click="applyModbusGpsToCandidate(false)"
                      :disabled="modbusRtuGpsStatus.state !== 'LOCKED'"
                      class="py-2 px-3 rounded-xl text-xs font-semibold transition-colors border bg-white dark:bg-slate-800 text-slate-700 dark:text-slate-200 border-slate-200 dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-700 disabled:opacity-45 disabled:cursor-not-allowed"
                    >
                      Pratinjau Peta
                    </button>
                    <button
                      type="button"
                      @click="applyModbusGpsToCandidate(true)"
                      :disabled="isSavingStationLocation || modbusRtuGpsStatus.state !== 'LOCKED'"
                      class="flex-1 py-2 px-3 rounded-xl text-xs font-semibold transition-colors border flex items-center justify-center space-x-1.5 bg-[#1685E5] hover:bg-[#1273C9] text-white btn-text border-[#1685E5] shadow-sm disabled:opacity-45 disabled:cursor-not-allowed"
                    >
                      <span v-if="isSavingStationLocation" class="animate-spin inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full"></span>
                      <span>Simpan Lokasi Modbus</span>
                    </button>
                  </div>
                </div>
              </div>
          </div>
        </div>
      </div>
    </header>

    <!-- WQMS IP & IKA Compact Horizontal Strip (Active only when Monitoring Domain is WQMS) -->
    <BentoWQMSIndexBar
      v-if="isWQMSDomain && isIspuBarVisible"
      :is-dark="isDark"
      :devices-list="devicesList"
    />

    <!-- =================================================================== -->
    <!-- 2. AUTO-ADAPTIVE MIDDLE BENTO:                                      -->
    <!--    - Wide (>=1280px): Left Full ISPU Hero Card + Right Parameters   -->
    <!--    - Compact (<1280px): Top Dominant ISPU Bar + Full-Width Params   -->
    <!-- =================================================================== -->
    <section
      class="bento-mid"
      :class="{
        'bento-mid-single': !(isAQMSDomain && isIspuBarVisible) || isCompactScreen,
        'is-compact-mid': isCompactScreen,
        'is-trend-hidden': !perfSettings.trendAnalysisEnabled
      }"
    >
      <!-- ISPU Component: Automatically switches between Compact Bar & Full Left Hero Card -->
      <BentoISPUHeroCard
        v-if="isAQMSDomain && isIspuBarVisible"
        :device-id="0"
        :devices-list="devicesList"
        :is-dark="isDark"
        :is-aqms-domain="isAQMSDomain"
        :compact="isCompactScreen"
        @device-change="handleIspuDeviceChange"
      />

      <!-- Parameter Workspace: Compact Icon Filter + Icon Card/Table Setup + Dense Cards / Dual Table -->
      <div ref="bentoParamsWrap" class="right-params-wrap">
        <div
          class="param-ctrl-bar"
          :class="{ 'has-open-popover': bentoActiveMiniPopover === 'paramFilter' || bentoActiveMiniPopover === 'cardSetup' }"
        >
          <!-- Left: Icon-Only Filter Popover + Icon Pin Toggle + Compact Active Filter Indicator -->
          <div class="param-ctrl-left">
            <!-- 1. Category & Parameter Filter Icon Popover -->
            <div class="relative bento-mini-popover-anchor">
              <button
                id="btn-bento-param-filter"
                type="button"
                @click="toggleBentoMiniPopover('paramFilter')"
                class="op-icon-btn xs"
                :class="{
                  'is-active': bentoActiveMiniPopover === 'paramFilter' || selectedCategoryFilter !== 'all'
                }"
                :title="'Filter Kategori Parameter (' + activeBentoCategoryName + ')'"
                aria-label="Filter Parameter"
              >
                <OperatorDashboardIcon name="filter" :size="13" />
                <span v-if="selectedCategoryFilter !== 'all'" class="op-corner-dot dot-sky"></span>
              </button>

              <div
                v-if="bentoActiveMiniPopover === 'paramFilter'"
                class="bento-mini-popover align-left w-64"
              >
                <div class="bmp-head">
                  <span class="bmp-title">Filter Parameter</span>
                  <div class="flex items-center gap-1">
                    <button
                      v-if="selectedCategoryFilter !== 'all' || filterPinnedOnly"
                      type="button"
                      class="bmp-link-btn"
                      @click="selectBentoAllParams(); bentoActiveMiniPopover = null"
                    >
                      Reset
                    </button>
                    <button
                      v-if="canManageCategories"
                      type="button"
                      class="mini-icon-btn"
                      title="Kelola Kategori Parameter"
                      @click="openCategoryManagementModal(); bentoActiveMiniPopover = null"
                    >
                      <OperatorDashboardIcon name="settings" :size="12" />
                    </button>
                  </div>
                </div>

                <div class="bmp-list custom-scroller">
                  <button
                    type="button"
                    class="bmp-item"
                    :class="{ active: selectedCategoryFilter === 'all' && !filterPinnedOnly }"
                    @click="selectBentoAllParams(); bentoActiveMiniPopover = null"
                  >
                    <span class="bmp-item-left">
                      <span class="bmp-dot" style="background: #0284C7;"></span>
                      <span>Semua Parameter</span>
                    </span>
                    <span class="bmp-count">{{ allParameters.length }}</span>
                  </button>

                  <button
                    type="button"
                    class="bmp-item"
                    :class="{ active: filterPinnedOnly }"
                    @click="toggleFilterPinned(); bentoActiveMiniPopover = null"
                  >
                    <span class="bmp-item-left">
                      <svg width="11" height="11" viewBox="0 0 24 24" fill="#F59E0B" stroke="#F59E0B" stroke-width="1.8" stroke-linejoin="round">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                      </svg>
                      <span>Favorit Dipin</span>
                    </span>
                    <span class="bmp-count">{{ pinnedParametersList.length }}</span>
                  </button>

                  <div v-if="activeCategoriesList.length > 0" class="bmp-divider"></div>

                  <button
                    v-for="cat in activeCategoriesList"
                    :key="'bento-cat-opt-' + cat.id"
                    type="button"
                    class="bmp-item"
                    :class="{ active: !filterPinnedOnly && Number(selectedCategoryFilter) === Number(cat.id) }"
                    @click="selectBentoCategory(cat.id); bentoActiveMiniPopover = null"
                  >
                    <span class="bmp-item-left min-w-0">
                      <span class="bmp-dot" :style="{ background: cat.color || '#10B981' }"></span>
                      <span class="truncate">{{ cat.name }}</span>
                    </span>
                    <span class="bmp-count">{{ getBentoCategoryCount(cat.id) }}</span>
                  </button>
                </div>
              </div>
            </div>

            <!-- 2. 1-Click Pinned / Favorites Icon Button -->
            <button
              id="btn-bento-param-pin"
              type="button"
              @click="toggleFilterPinned"
              class="op-icon-btn xs"
              :class="{ 'is-amber': filterPinnedOnly }"
              :title="filterPinnedOnly ? 'Tampilkan Semua (Matikan Filter Favorit)' : ('Filter Parameter Favorit (' + pinnedParametersList.length + ')')"
              aria-label="Filter Favorit"
            >
              <OperatorDashboardIcon name="pin" :size="13" />
              <span v-if="pinnedParametersList.length > 0" class="op-corner-pill tone-amber">{{ pinnedParametersList.length }}</span>
            </button>

            <!-- Compact Channel Counter + Active Filter Dismissible Tag -->
            <span class="param-count-tag">
              <strong>{{ filteredParameters.length }}</strong><template v-if="filteredParameters.length !== allParameters.length">/{{ allParameters.length }}</template>
              <span>Kanal</span>
            </span>

            <button
              v-if="filterPinnedOnly"
              type="button"
              class="active-filter-chip amber"
              @click="toggleFilterPinned"
              title="Hapus filter favorit"
            >
              <span>★ Favorit</span>
              <span class="afc-x">×</span>
            </button>

            <button
              v-else-if="selectedCategoryFilter !== 'all'"
              type="button"
              class="active-filter-chip"
              @click="selectBentoAllParams"
              title="Hapus filter kategori"
            >
              <span class="truncate max-w-[140px]">{{ activeBentoCategoryName }}</span>
              <span class="afc-x">×</span>
            </button>
          </div>

          <!-- Right: Icon-Only Card & Table Display Switcher + Setup Popover -->
          <div class="param-ctrl-right flex-shrink-0">
            <div class="icon-seg-capsule" role="group" aria-label="Mode Tampilan Parameter">
              <!-- Cards Mode Icon (All Combined) -->
              <button
                id="btn-bento-view-cards"
                type="button"
                class="isc-btn"
                :class="{ active: viewMode === 'cards' && !bentoCardGroupByCategory }"
                @click="setBentoDisplayMode('cards')"
                :title="'Tampilan Kartu Gabung (' + (bentoAutoGrid ? 'Otomatis ' : '') + effectiveBentoCols + 'P)'"
                aria-label="Mode Kartu Gabung"
              >
                <OperatorDashboardIcon name="layout" :size="13" />
              </button>

              <!-- Category-Separated Cards Icon -->
              <button
                id="btn-bento-view-cat-cards"
                type="button"
                class="isc-btn"
                :class="{ active: viewMode === 'cards' && bentoCardGroupByCategory }"
                @click="setBentoDisplayMode('cat-cards')"
                :title="'Tampilan Kartu Terpisah per Kategori (' + bentoCategoryTableSections.length + ' Kategori · ' + effectiveBentoCols + 'P)'"
                aria-label="Mode Kartu per Kategori"
              >
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <line x1="2.5" y1="3.5" x2="10.5" y2="3.5" />
                  <rect x="2.5" y="6" width="8.5" height="5.2" rx="1.2" />
                  <rect x="13" y="6" width="8.5" height="5.2" rx="1.2" />
                  <line x1="2.5" y1="14.3" x2="10.5" y2="14.3" />
                  <rect x="2.5" y="16.8" width="8.5" height="5.2" rx="1.2" />
                  <rect x="13" y="16.8" width="8.5" height="5.2" rx="1.2" />
                </svg>
              </button>

              <!-- Category Tables Side-by-Side Icon (CAT) -->
              <button
                id="btn-bento-view-cat"
                type="button"
                class="isc-btn"
                :class="{ active: viewMode === 'table' && bentoTableSplitMode === 'CAT' }"
                @click="setBentoDisplayMode('CAT')"
                :title="'Tampilan Tabel per Kategori (' + bentoCategoryTableSections.length + ' Tabel)'"
                aria-label="Mode Tabel per Kategori"
              >
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="2" y="4" width="5.2" height="16" rx="1.2" />
                  <rect x="9.4" y="4" width="5.2" height="16" rx="1.2" />
                  <rect x="16.8" y="4" width="5.2" height="16" rx="1.2" />
                  <line x1="3.4" y1="8.5" x2="5.8" y2="8.5" />
                  <line x1="10.8" y1="8.5" x2="13.2" y2="8.5" />
                  <line x1="18.2" y1="8.5" x2="20.6" y2="8.5" />
                </svg>
              </button>

              <!-- 2-Table Side-by-Side Icon -->
              <button
                id="btn-bento-view-2t"
                type="button"
                class="isc-btn"
                :class="{ active: viewMode === 'table' && bentoTableSplitMode === '2T' }"
                @click="setBentoDisplayMode('2T')"
                title="Tampilan 2 Tabel ke Samping (2T)"
                aria-label="Mode 2 Tabel Samping"
              >
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <rect x="2.5" y="4" width="8" height="16" rx="1.8" />
                  <rect x="13.5" y="4" width="8" height="16" rx="1.8" />
                  <line x1="5" y1="8.5" x2="8" y2="8.5" />
                  <line x1="5" y1="12" x2="8" y2="12" />
                  <line x1="5" y1="15.5" x2="8" y2="15.5" />
                  <line x1="16" y1="8.5" x2="19" y2="8.5" />
                  <line x1="16" y1="12" x2="19" y2="12" />
                  <line x1="16" y1="15.5" x2="19" y2="15.5" />
                </svg>
              </button>

              <!-- 1-Table Full Width Icon -->
              <button
                id="btn-bento-view-1t"
                type="button"
                class="isc-btn"
                :class="{ active: viewMode === 'table' && bentoTableSplitMode === '1T' }"
                @click="setBentoDisplayMode('1T')"
                title="Tampilan 1 Tabel Penuh (1T)"
                aria-label="Mode 1 Tabel Penuh"
              >
                <OperatorDashboardIcon name="menu" :size="13" />
              </button>
            </div>

            <!-- Setup Card & Table Density Icon Popover -->
            <div class="relative bento-mini-popover-anchor">
              <button
                id="btn-bento-card-setup"
                type="button"
                @click="toggleBentoMiniPopover('cardSetup')"
                class="op-icon-btn xs"
                :class="{ 'is-active': bentoActiveMiniPopover === 'cardSetup' || !bentoAutoGrid || (viewMode === 'cards' && bentoCardGroupByCategory) }"
                :title="'Setup Kartu & Tabel (' + (viewMode === 'table' ? bentoTableModeShortBadge : ((bentoCardGroupByCategory ? 'Per Kategori · ' : '') + (bentoAutoGrid ? 'Auto ' : '') + effectiveBentoCols + 'P')) + ')'"
                aria-label="Setup Kartu dan Tabel"
              >
                <OperatorDashboardIcon name="sliders" :size="13" />
                <span class="op-corner-pill tone-Density">
                  {{ viewMode === 'table' ? bentoTableModeShortBadge : bentoCardModeShortBadge }}
                </span>
              </button>

              <div
                v-if="bentoActiveMiniPopover === 'cardSetup'"
                class="bento-mini-popover align-right w-80"
              >
                <div class="bmp-head">
                  <span class="bmp-title">Setup Kartu &amp; Tabel</span>
                  <span class="bmp-badge">{{ viewMode === 'table' ? (bentoTableSplitMode === 'CAT' ? ('Kategori · ' + bentoCategoryTableSections.length + ' Tabel') : ('Tabel ' + bentoTableSplitMode)) : ((bentoCardGroupByCategory ? 'Kat. Kartu · ' : '') + (bentoAutoGrid ? 'Auto · ' : 'Manual · ') + effectiveBentoCols + 'P') }}</span>
                </div>

                <div class="bmp-section">
                  <div class="bmp-label">Mode Tampilan</div>
                  <div class="bmp-seg-grid cols-5">
                    <button
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: viewMode === 'cards' && !bentoCardGroupByCategory }"
                      @click="setBentoDisplayMode('cards')"
                      title="Tampilan kartu parameter gabungan dalam 1 grid"
                    >
                      Kartu
                    </button>
                    <button
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: viewMode === 'cards' && bentoCardGroupByCategory }"
                      @click="setBentoDisplayMode('cat-cards')"
                      title="Tampilan kartu parameter terpisah sesuai kategori"
                    >
                      Kartu Kat.
                    </button>
                    <button
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: viewMode === 'table' && bentoTableSplitMode === 'CAT' }"
                      @click="setBentoDisplayMode('CAT')"
                      title="Tabel otomatis sebanyak kategori aktif yang terpakai"
                    >
                      Tabel Kat.
                    </button>
                    <button
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: viewMode === 'table' && bentoTableSplitMode === '2T' }"
                      @click="setBentoDisplayMode('2T')"
                    >
                      2 Tabel
                    </button>
                    <button
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: viewMode === 'table' && bentoTableSplitMode === '1T' }"
                      @click="setBentoDisplayMode('1T')"
                    >
                      1 Tabel
                    </button>
                  </div>
                </div>

                <div v-if="viewMode === 'cards'" class="bmp-section">
                  <div class="bmp-label">Susunan Kartu Parameter</div>
                  <div class="bmp-seg-grid cols-2">
                    <button
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: !bentoCardGroupByCategory }"
                      @click="setBentoCardGrouping(false)"
                    >
                      Gabung Semua
                    </button>
                    <button
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: bentoCardGroupByCategory }"
                      @click="setBentoCardGrouping(true)"
                    >
                      Pisah Kategori
                    </button>
                  </div>
                  <div
                    v-if="bentoCardGroupByCategory"
                    class="mt-1.5 flex items-center justify-between text-[9.5px] opacity-85"
                  >
                    <span>{{ bentoCategoryTableSections.length }} grup kartu kategori aktif</span>
                    <button
                      v-if="canManageCategories"
                      type="button"
                      class="bmp-link-btn"
                      @click="openCategoryManagementModal(); bentoActiveMiniPopover = null"
                    >
                      Kelola Kategori
                    </button>
                  </div>
                </div>

                <div
                  v-if="viewMode === 'table' && bentoTableSplitMode === 'CAT'"
                  class="bmp-section -mt-1 flex items-center justify-between text-[9.5px] opacity-85"
                >
                  <span>{{ bentoCategoryTableSections.length }} tabel kategori aktif</span>
                  <button
                    v-if="canManageCategories"
                    type="button"
                    class="bmp-link-btn"
                    @click="openCategoryManagementModal(); bentoActiveMiniPopover = null"
                  >
                    Kelola Kategori
                  </button>
                </div>

                <div class="bmp-section">
                  <div class="bmp-label">Kolom Kartu Parameter</div>
                  <div class="bmp-seg-grid" :class="bentoCardGroupByCategory ? 'cols-7' : 'cols-6'">
                    <button
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: viewMode === 'cards' && bentoAutoGrid }"
                      @click="setBentoDisplayMode('auto-cards')"
                      title="Otomatis menyesuaikan ukuran layar agar seluruh kategori muat tanpa scroll"
                    >
                      Auto
                    </button>
                    <button
                      v-for="col in (bentoCardGroupByCategory ? [3, 4, 5, 6, 7, 8] : [1, 2, 3, 4, 5])"
                      :key="'bento-col-setup-' + col"
                      type="button"
                      class="bmp-seg-btn"
                      :class="{ active: viewMode === 'cards' && !bentoAutoGrid && Number(gridColumns) === col }"
                      @click="setGridColumns(col)"
                    >
                      {{ col }}P
                    </button>
                  </div>
                </div>

                <div class="bmp-divider"></div>

                <label class="bmp-toggle-row">
                  <span>Indikator Kedip Data Baru</span>
                  <input
                    type="checkbox"
                    :checked="perfSettings.cardBlinksEnabled"
                    @change="setPerfSetting('cardBlinksEnabled', $event.target.checked)"
                  />
                </label>
              </div>
            </div>
          </div>
        </div>

        <!-- =============================================================== -->
        <!-- VIEW MODE A: AUTO-ADAPTIVE BENTO SENSOR CARDS                   -->
        <!-- =============================================================== -->
        <div
          v-if="viewMode === 'cards'"
          class="sensors-scroll-area custom-scroller"
          :class="{
            'is-compact-scroll': isCompactScreen,
            'is-trend-hidden': !perfSettings.trendAnalysisEnabled
          }"
          :style="bentoParamsScrollAreaStyle"
        >
          <div
            v-if="filteredParameters.length > 0"
            :class="{ 'bento-cat-cards-stack': bentoCardGroupByCategory }"
            :style="bentoCardGroupByCategory ? { '--cat-total-cols': bentoCategorySlotCols } : null"
          >
            <div
              v-for="sec in bentoActiveCardSections"
              :key="sec.key"
              :class="{ 'bento-cat-card-group': sec.isCategoryGroup }"
              :style="sec.isCategoryGroup ? { borderLeftColor: sec.accentColor || '#06B6D4', gridColumn: 'span ' + (sec.colSpan || bentoCategorySlotCols) } : null"
            >
              <!-- Category Section Header (shown when cards are separated by category) -->
              <div
                v-if="sec.isCategoryGroup"
                class="bento-cat-card-head"
                :class="{ 'is-collapsed-head': isBentoCatCardCollapsed(sec.key) }"
                @click="toggleBentoCatCardCollapse(sec.key)"
                :title="isBentoCatCardCollapsed(sec.key) ? 'Klik untuk menampilkan kartu kategori ini' : 'Klik untuk melipat kartu kategori ini'"
              >
                <div class="bento-cat-card-head-left">
                  <span
                    class="bento-cat-dot"
                    :style="{ background: sec.accentColor || '#06B6D4' }"
                  ></span>
                  <span class="bento-cat-card-title truncate" :title="sec.title">{{ sec.title }}</span>
                </div>
                <div class="bento-cat-card-head-right">
                  <span v-if="sec.faultCount > 0 && (sec.colSpan || 2) >= 2" class="s-badge badge-rose">{{ sec.faultCount }} Fault</span>
                  <span v-else-if="sec.warnCount > 0 && (sec.colSpan || 2) >= 2" class="s-badge badge-amber">{{ sec.warnCount }} Alarm</span>
                  <span class="bento-cat-count-pill">{{ (sec.colSpan || 2) === 1 ? sec.params.length : (sec.params.length + ' Kanal') }}</span>
                  <svg
                    width="11"
                    height="11"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2.2"
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    class="bento-cat-chevron"
                    :class="{ 'is-collapsed': isBentoCatCardCollapsed(sec.key) }"
                  >
                    <polyline points="6 9 12 15 18 9" />
                  </svg>
                </div>
              </div>

              <div
                v-show="!sec.isCategoryGroup || !isBentoCatCardCollapsed(sec.key)"
                class="sensors-grid"
                :class="[
                  sec.isCategoryGroup ? ('bento-cols-' + (sec.innerCols || effectiveBentoCols)) : ('bento-cols-' + effectiveBentoCols),
                  {
                    'is-micro-cards': isMicroCardDensity || sec.isCategoryGroup,
                    'is-compact-cards': isCompactScreen,
                    'is-inside-cat-group': sec.isCategoryGroup
                  }
                ]"
                :style="sec.isCategoryGroup ? { gridTemplateColumns: 'repeat(' + (sec.innerCols || 1) + ', minmax(0, 1fr))' } : null"
              >
                <div
                  v-for="param in sec.params"
                  :key="param.param_id"
                  @click="selectParamForTrend(param)"
                  @contextmenu.prevent="openParamContextMenu($event, param)"
                  class="s-card"
                  :class="[
                    { 's-card-selected': isParamSelectedForTrend(param) },
                    getBentoCardAlertClass(param),
                    perfSettings.cardBlinksEnabled && liveBlinks[param.param_id] ? 's-card-blink' : ''
                  ]"
                  :title="getBentoParamCardTitle(param)"
                >
                  <!-- Card Header -->
                  <div class="s-head">
                    <div class="s-title-wrap">
                      <span
                        class="s-dot"
                        :style="{ background: getBentoDotColor(param) }"
                      ></span>
                      <span class="s-name" :title="param.param_name || param.param_code">
                        {{ param.param_name || param.param_code }}
                      </span>
                    </div>

                    <div class="s-head-right">
                      <span v-if="!sec.isCategoryGroup && !isCompactScreen && getParamCategory(param)" class="s-cat">
                        {{ getParamCategoryShortLabel(param) }}
                      </span>
                  <span
                    class="s-badge"
                    :class="[getBentoStatusBadge(param).toneClass, { 'is-clickable-alarm': getBentoStatusBadge(param).isAlarm }]"
                    :title="getBentoStatusBadge(param).tooltip || ''"
                    @click="getBentoStatusBadge(param).isAlarm ? ($event.stopPropagation(), onBentoStatusBadgeClick(param)) : null"
                  >
                    {{ getBentoStatusBadge(param).label }}
                  </span>
                  <span class="s-card-quick-actions" :class="{ 'has-pinned-star': isPinned(param.param_id) }">
                    <button
                      type="button"
                      class="s-icon-btn"
                      :class="{ 'is-pinned-star': isPinned(param.param_id) }"
                      @click.stop="togglePin(param.param_id)"
                      :title="isPinned(param.param_id) ? 'Lepas Favorit' : 'Sematkan Favorit'"
                    >
                      <svg v-if="isPinned(param.param_id)" width="11" height="11" viewBox="0 0 24 24" fill="#F59E0B" stroke="#F59E0B" stroke-width="1.8" stroke-linejoin="round">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                      </svg>
                      <svg v-else width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round">
                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                      </svg>
                    </button>
                    <button
                      type="button"
                      class="s-icon-btn s-icon-menu-btn"
                      @click.stop="openParamContextMenu($event, param)"
                      title="Menu Aksi Parameter (Test Read, Konfigurasi, Riwayat)"
                    >
                      <svg width="10.5" height="10.5" viewBox="0 0 24 24" fill="currentColor">
                        <circle cx="12" cy="5" r="2" />
                        <circle cx="12" cy="12" r="2" />
                        <circle cx="12" cy="19" r="2" />
                      </svg>
                    </button>
                  </span>
                </div>
              </div>

              <!-- Card Body: Discrete Status OR Tabular Analog Reading + Sparkline -->
              <div v-if="isDiscreteStatusParam(param)" class="s-body">
                <div class="flex items-center gap-1.5 min-w-0">
                  <span
                    class="s-discrete-pill inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-bold text-[10.5px] uppercase"
                    :class="getDiscreteStatusInfo(param).pillClass"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="getDiscreteStatusInfo(param).dotClass"></span>
                    <span>{{ getDiscreteStatusInfo(param).label }}</span>
                  </span>
                  <span v-if="getDiscreteStatusInfo(param).durationText" class="s-discrete-dur text-[9.5px] opacity-75 truncate">
                    {{ getDiscreteStatusInfo(param).durationText }}
                  </span>
                </div>

                <div class="s-spark-wrap is-discrete-spark">
                  <svg
                    v-if="perfSettings.parameterSparklinesEnabled && getStatusTimelineBlocks(param).length > 0"
                    class="w-full h-full block"
                    viewBox="0 0 100 20"
                    preserveAspectRatio="none"
                  >
                    <rect
                      v-for="(blk, bIdx) in getStatusTimelineBlocks(param)"
                      :key="'bento-st-' + param.param_id + '-' + bIdx"
                      :x="blk.x"
                      :y="blk.y"
                      :width="blk.w"
                      :height="blk.h"
                      :rx="blk.rx"
                      :fill="blk.fill"
                      :fill-opacity="blk.opacity"
                    />
                  </svg>
                </div>
              </div>

              <div v-else class="s-body">
                <div class="s-reading min-w-0">
                  <span class="s-val" :class="{ 's-val-muted': formatParamValue(param) === '--' }">
                    {{ formatParamValue(param) }}
                  </span>
                  <span v-if="param.unit" class="s-unit">{{ param.unit }}</span>
                  <span v-if="param.is_held_value" class="s-held-tag">HELD</span>
                </div>

                <div class="s-spark-wrap">
                  <svg
                    v-if="perfSettings.parameterSparklinesEnabled && getSparkline(param).hasHistory"
                    class="s-spark-svg"
                    viewBox="0 0 100 20"
                    preserveAspectRatio="none"
                    style="overflow: visible"
                  >
                    <defs>
                      <linearGradient :id="'bentoSpGrad-' + param.param_id" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" :stop-color="getBentoCardColor(param)" :stop-opacity="isDark ? '0.26' : '0.18'" />
                        <stop offset="100%" :stop-color="getBentoCardColor(param)" stop-opacity="0" />
                      </linearGradient>
                    </defs>
                    <path
                      v-for="(areaD, aIdx) in getSparkline(param).areaPaths"
                      :key="'ba-' + param.param_id + '-' + aIdx"
                      :d="areaD"
                      :fill="'url(#bentoSpGrad-' + param.param_id + ')'"
                    />
                    <path
                      v-for="(strokeD, sIdx) in getSparkline(param).strokePaths"
                      :key="'bs-' + param.param_id + '-' + sIdx"
                      :d="strokeD"
                      fill="none"
                      :stroke="getBentoCardColor(param)"
                      stroke-width="1.9"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                    />
                  </svg>
                  <div v-else class="s-spark-empty"></div>
                </div>
              </div>

              <!-- Card Footer: Threshold Progress Bar + Adaptive Context -->
              <div class="s-foot">
                <div class="s-bar-track">
                  <div
                    class="s-bar-fill"
                    :style="{ width: getBentoCardProgress(param).pct + '%', background: getBentoCardColor(param) }"
                  ></div>
                </div>
                <div class="s-meta">
                  <button
                    type="button"
                    class="s-meta-limit-btn truncate"
                    :class="{ 'is-custom-limit': getBentoCardProgress(param).isConfiguredLimit }"
                    @click.stop="openBentoQuickLimitEdit(param)"
                    :title="getBentoCardProgress(param).limitTooltip"
                  >
                    {{ (sec.isCategoryGroup || isMicroCardDensity) ? getBentoCardProgress(param).shortMetaLabel : getBentoCardProgress(param).metaLabel }}
                  </button>
                  <span class="s-meta-pct">{{ getBentoCardProgress(param).rightLabel }}</span>
                </div>
              </div>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="bento-empty-cards">
            {{ ($te && $te('operator.noParamsMatch')) ? $t('operator.noParamsMatch') : 'Tidak ada parameter yang sesuai dengan filter saat ini.' }}
          </div>
        </div>

        <!-- =============================================================== -->
        <!-- VIEW MODE B: CATEGORY TABLES (UP TO 3 SIDE-BY-SIDE), 2T & 1T    -->
        <!-- =============================================================== -->
        <div
          v-else
          class="bento-table-area custom-scroller"
          :class="{
            'is-compact-scroll': isCompactScreen,
            'is-trend-hidden': !perfSettings.trendAnalysisEnabled
          }"
          :style="bentoParamsScrollAreaStyle"
        >
          <div
            v-if="filteredParameters.length > 0"
            class="bento-dual-table-grid"
            :class="{
              'is-dual-cols': bentoEffectiveTableGridCols === 2,
              'is-tri-cols': bentoEffectiveTableGridCols === 3,
              'is-dense-tbl': isBentoDenseTableCard
            }"
          >
            <div
              v-for="(col, colIdx) in bentoSplitTableColumns"
              :key="col.key || ('bento-tcol-' + colIdx)"
              class="bento-table-card"
              :style="col.isCategoryTable && col.accentColor ? { borderTop: '2px solid ' + col.accentColor } : null"
            >
              <div
                class="bento-tbl-subhead"
                :class="{ 'is-cat-subhead': col.isCategoryTable }"
              >
                <div class="bento-tbl-subhead-left">
                  <span
                    v-if="col.isCategoryTable"
                    class="bento-cat-dot"
                    :style="{ background: col.accentColor || '#06B6D4' }"
                  ></span>
                  <span class="truncate" :title="col.title">{{ col.title }}</span>
                </div>
                <div class="bento-tbl-subhead-right">
                  <span v-if="col.faultCount > 0" class="s-badge badge-rose">{{ col.faultCount }} Fault</span>
                  <span v-else-if="col.warnCount > 0" class="s-badge badge-amber">{{ col.warnCount }} Alarm</span>
                  <span>{{ col.params.length }} Kanal</span>
                </div>
              </div>

              <div class="bento-tbl-wrap custom-scroller">
                <table
                  class="bento-tbl"
                  :class="{ 'has-spark-col': showBentoTableSparklineCol }"
                >
                  <colgroup>
                    <col class="bt-col-param" />
                    <col class="bt-col-val" />
                    <col v-if="showBentoTableSparklineCol" class="bt-col-spark" />
                    <col class="bt-col-status" />
                    <col class="bt-col-limit" />
                    <col v-if="showBentoTableUpdateCol" class="bt-col-update" />
                    <col class="bt-col-act" />
                  </colgroup>
                  <thead>
                    <tr>
                      <th class="bt-th-param">Parameter</th>
                      <th class="bt-th-val">Nilai</th>
                      <th v-if="showBentoTableSparklineCol" class="bt-th-center">Tren</th>
                      <th class="bt-th-center">Status</th>
                      <th class="bt-th-center">Limit</th>
                      <th v-if="showBentoTableUpdateCol" class="bt-th-center">Update</th>
                      <th class="bt-th-center">Aksi</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="param in col.params"
                      :key="'btrow-' + param.param_id"
                      @click="selectParamForTrend(param)"
                      @contextmenu.prevent="openParamContextMenu($event, param)"
                      class="bento-tbl-row"
                      :class="[
                        { 'is-row-selected': isParamSelectedForTrend(param) },
                        getBentoRowAlertClass(param)
                      ]"
                      :title="getBentoParamCardTitle(param)"
                    >
                      <!-- 1. Parameter Name + Category (category pill omitted when already grouped by category) -->
                      <td class="bt-td-param">
                        <div class="bt-param-cell">
                          <span
                            class="s-dot"
                            :style="{ background: getBentoDotColor(param) }"
                          ></span>
                          <span class="bt-name" :title="param.param_name || param.param_code">
                            {{ param.param_name || param.param_code }}
                          </span>
                          <span v-if="!col.isCategoryTable && !isCompactScreen && getParamCategory(param)" class="s-cat flex-shrink-0">
                            {{ getParamCategoryShortLabel(param) }}
                          </span>
                        </div>
                      </td>

                      <!-- 2. Decimal-Aligned Value + Fixed Unit Track -->
                      <td class="bt-td-val">
                        <div v-if="isDiscreteStatusParam(param)" class="bt-reading-discrete">
                          <span
                            class="bt-discrete-pill"
                            :class="getDiscreteStatusInfo(param).pillClass"
                          >
                            <span class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="getDiscreteStatusInfo(param).dotClass"></span>
                            <span>{{ getDiscreteStatusInfo(param).label }}</span>
                          </span>
                        </div>
                        <div v-else class="bt-reading-grid">
                          <span class="bt-val" :class="{ 's-val-muted': formatParamValue(param) === '--' }">
                            {{ formatParamValue(param) }}
                          </span>
                          <span class="bt-unit-slot" :title="param.unit || ''">{{ param.unit || '' }}</span>
                        </div>
                      </td>

                      <!-- 3. Inline Mini Sparkline -->
                      <td v-if="showBentoTableSparklineCol" class="bt-td-center">
                        <div class="bt-spark-box mx-auto">
                          <template v-if="isDiscreteStatusParam(param)">
                            <svg
                              v-if="perfSettings.parameterSparklinesEnabled && getStatusTimelineBlocks(param).length > 0"
                              class="w-full h-full block"
                              viewBox="0 0 100 20"
                              preserveAspectRatio="none"
                            >
                              <rect
                                v-for="(blk, bIdx) in getStatusTimelineBlocks(param)"
                                :key="'btst-' + param.param_id + '-' + bIdx"
                                :x="blk.x"
                                :y="blk.y"
                                :width="blk.w"
                                :height="blk.h"
                                :rx="blk.rx"
                                :fill="blk.fill"
                                :fill-opacity="blk.opacity"
                              />
                            </svg>
                            <div v-else class="s-spark-empty"></div>
                          </template>
                          <template v-else>
                            <svg
                              v-if="perfSettings.parameterSparklinesEnabled && getSparkline(param).hasHistory"
                              class="w-full h-full block"
                              viewBox="0 0 100 20"
                              preserveAspectRatio="none"
                              style="overflow: visible"
                            >
                              <path
                                v-for="(strokeD, sIdx) in getSparkline(param).strokePaths"
                                :key="'btsp-' + param.param_id + '-' + sIdx"
                                :d="strokeD"
                                fill="none"
                                :stroke="getBentoCardColor(param)"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                              />
                            </svg>
                            <div v-else class="s-spark-empty"></div>
                          </template>
                        </div>
                      </td>

                      <!-- 4. Uniform Fixed-Width Status Pill -->
                      <td class="bt-td-center">
                        <span
                          class="s-badge bt-status-pill"
                          :class="[getBentoStatusBadge(param).toneClass, { 'is-clickable-alarm': getBentoStatusBadge(param).isAlarm }]"
                          :title="getBentoStatusBadge(param).tooltip || ''"
                          @click="getBentoStatusBadge(param).isAlarm ? ($event.stopPropagation(), onBentoStatusBadgeClick(param)) : null"
                        >
                          {{ getBentoStatusBadge(param).label }}
                        </span>
                      </td>

                      <!-- 5. Uniform Fixed-Width Configured Limit Pill (+ Subtle Bottom Fill Bar) -->
                      <td class="bt-td-center">
                        <span v-if="isDiscreteStatusParam(param)" class="bt-limit-dash" title="Kanal status digital">—</span>
                        <button
                          v-else-if="getBentoCardProgress(param).tableLimitLabel"
                          type="button"
                          class="bt-limit-btn"
                          :class="{ 'is-custom-limit': getBentoCardProgress(param).isConfiguredLimit }"
                          @click.stop="openBentoQuickLimitEdit(param)"
                          :title="getBentoCardProgress(param).limitTooltip"
                        >
                          <span class="bt-limit-text">{{ getBentoCardProgress(param).tableLimitLabel }}</span>
                          <span
                            v-if="getBentoCardProgress(param).pct > 0"
                            class="bt-limit-bar"
                            :style="{ width: getBentoCardProgress(param).pct + '%', background: getBentoCardColor(param) }"
                          ></span>
                        </button>
                        <button
                          v-else
                          type="button"
                          class="bt-limit-btn is-empty-limit"
                          @click.stop="openBentoQuickLimitEdit(param)"
                          :title="getBentoCardProgress(param).limitTooltip"
                        >
                          <span class="bt-limit-text">Set Limit</span>
                        </button>
                      </td>

                      <!-- 6. Freshness (Shown on 1T or Wide 2T) -->
                      <td v-if="showBentoTableUpdateCol" class="bt-td-center whitespace-nowrap">
                        <span class="bt-fresh">
                          {{ param.received_at ? formatRelative(param.received_at) : '-' }}
                        </span>
                      </td>

                      <!-- 7. Favorite + Context Menu Actions -->
                      <td class="bt-td-center whitespace-nowrap" @click.stop>
                        <div class="bt-act-group">
                          <button
                            type="button"
                            class="s-icon-btn bt-act-btn"
                            @click.stop="togglePin(param.param_id)"
                            :title="isPinned(param.param_id) ? 'Lepas Favorit' : 'Sematkan Favorit'"
                          >
                            <svg v-if="isPinned(param.param_id)" width="11" height="11" viewBox="0 0 24 24" fill="#F59E0B" stroke="#F59E0B" stroke-width="1.8" stroke-linejoin="round">
                              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                            </svg>
                            <svg v-else width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round">
                              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                            </svg>
                          </button>
                          <button
                            type="button"
                            class="s-icon-btn bt-act-btn"
                            @click.stop="openParamContextMenu($event, param)"
                            title="Menu Aksi Parameter"
                          >
                            <svg width="11" height="11" viewBox="0 0 24 24" fill="currentColor">
                              <circle cx="12" cy="5" r="2" />
                              <circle cx="12" cy="12" r="2" />
                              <circle cx="12" cy="19" r="2" />
                            </svg>
                          </button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>

          <div v-else class="bento-empty-cards">
            {{ ($te && $te('operator.noParamsMatch')) ? $t('operator.noParamsMatch') : 'Tidak ada parameter yang sesuai dengan filter saat ini.' }}
          </div>
        </div>
      </div>
    </section>

    <!-- =================================================================== -->
    <!-- 3. BOTTOM ZONE: COMPLETE TREND ANALYSIS STUDIO (COMPACT 1-ROW BAR)  -->
    <!-- =================================================================== -->
    <section v-if="perfSettings.trendAnalysisEnabled" class="trend-studio">
      <!-- Single Compact Control Bar: Zero Horizontal Scrollbar, Minimal Vertical Footprint -->
      <div
        class="ts-compact-bar"
        :class="{ 'has-open-popover': bentoActiveMiniPopover === 'trendFilter' || bentoActiveMiniPopover === 'trendHeight' || showTrendSettingsMenu || showTrendDownloadMenu || showLineSmoothMenu }"
      >
        <!-- Left Cluster: Concise Title, Compact Live/Diag Dot, Filter Parameter Popover, Chart Tool Icons -->
        <div class="ts-c-left">
          <span class="ts-title">Analisis Tren</span>

          <!-- Consolidated Live & Communication Badge (Hover for full diagnostic details + latency) -->
          <span
            class="live-badge"
            :class="!$store.state.wsConnected ? 'is-offline' : (trendCommDiagnostics.state === 'ONLINE' || trendCommDiagnostics.state === 'RECOVERED' ? 'is-live' : 'is-warn')"
            :title="($store.state.wsConnected ? 'WebSocket Aktif · ' : 'WebSocket Terputus · ') + (trendCommDiagnostics.badge || '') + (trendCommDiagnostics.message ? ' — ' + trendCommDiagnostics.message : '') + (pipelineDiagnostics && pipelineDiagnostics.sampleCount > 0 ? ' · Latency ' + pipelineDiagnostics.totalEndToEndMs + 'ms' : '')"
          >
            <span
              class="status-dot"
              :class="!$store.state.wsConnected ? 'dot-rose' : (trendCommDiagnostics.state === 'ONLINE' || trendCommDiagnostics.state === 'RECOVERED' ? 'dot-emerald' : 'dot-amber')"
            ></span>
            <span>{{ !$store.state.wsConnected ? 'OFFLINE' : (trendCommDiagnostics.state === 'ONLINE' || trendCommDiagnostics.state === 'RECOVERED' ? 'LIVE' : trendCommDiagnostics.badge) }}</span>
          </span>

          <button
            v-if="trendCommDiagnostics.state === 'API_ERROR'"
            type="button"
            @click="fetchTrendData(false)"
            class="op-icon-btn xs text-rose-500"
            title="Muat ulang data grafik"
          >
            <OperatorDashboardIcon name="reset" :size="12" />
          </button>

          <!-- Compact Parameter Filter Popover Trigger (Replaces bulky horizontal chip row) -->
          <div class="relative bento-mini-popover-anchor">
            <button
              id="btn-bento-trend-filter"
              type="button"
              @click="toggleBentoMiniPopover('trendFilter')"
              class="ts-filter-trigger"
              :class="{ active: bentoActiveMiniPopover === 'trendFilter' }"
              title="Filter & Pilih Parameter Grafik Tren"
              aria-label="Filter Parameter Grafik"
            >
              <OperatorDashboardIcon name="filter" :size="12" />
              <span v-if="previewTrendDots.length > 0" class="ts-mini-dots">
                <span
                  v-for="(col, dIdx) in previewTrendDots"
                  :key="'tdot-' + dIdx"
                  class="ts-mini-dot"
                  :style="{ background: col }"
                ></span>
              </span>
              <span class="ts-filter-cnt">{{ activeTrendSeries.length }}/{{ allParameters.length }}</span>
            </button>

            <!-- Compact Multi-Select Parameter Popover -->
            <div
              v-if="bentoActiveMiniPopover === 'trendFilter'"
              class="bento-mini-popover align-left w-80 sm:w-[356px]"
            >
              <div class="bmp-head">
                <span class="bmp-title">Parameter Grafik ({{ activeTrendSeries.length }}/{{ allParameters.length }})</span>
                <div class="flex items-center gap-1">
                  <button
                    type="button"
                    class="bmp-mini-act"
                    @click="selectAllEligibleTrendSeries(true)"
                    title="Plot Semua Parameter"
                  >
                    <OperatorDashboardIcon name="check-all" :size="11" />
                    <span>Semua</span>
                  </button>
                  <button
                    v-if="activeTrendSeries.length > 0"
                    type="button"
                    class="bmp-mini-act"
                    @click="clearAllTrendSeries"
                    title="Kosongkan Grafik"
                  >
                    <OperatorDashboardIcon name="clear" :size="11" />
                    <span>Clear</span>
                  </button>
                  <button
                    type="button"
                    class="bmp-mini-act"
                    @click="resetTrendSelectionToDefault"
                    title="Reset ke Default"
                  >
                    <OperatorDashboardIcon name="reset" :size="11" />
                  </button>
                </div>
              </div>

              <!-- Quick Search inside Trend Filter Popover -->
              <div class="bmp-search-wrap">
                <OperatorDashboardIcon name="search" :size="12" class="opacity-60 flex-shrink-0" />
                <input
                  v-model.trim="bentoTrendSearch"
                  type="text"
                  class="bmp-search-input"
                  placeholder="Cari parameter..."
                />
                <button
                  v-if="bentoTrendSearch"
                  type="button"
                  class="bmp-search-clear"
                  @click="bentoTrendSearch = ''"
                >
                  ×
                </button>
              </div>

              <!-- 2-Column Space-Saving Parameter Toggle Grid -->
              <div class="trend-param-grid custom-scroller">
                <button
                  v-for="p in filteredTrendPopoverParams"
                  :key="'tp-opt-' + p.param_id"
                  type="button"
                  class="tp-opt-btn"
                  :class="{ active: isParamSelectedForTrend(p) }"
                  @click="selectParamForTrend(p)"
                  :title="(p.param_name || p.param_code) + ' · ' + formatParamValue(p) + (p.unit ? ' ' + p.unit : '')"
                >
                  <span class="tp-opt-left min-w-0">
                    <span class="tp-dot" :style="{ background: getTrendParamDotColor(p) }"></span>
                    <span class="tp-name truncate">{{ p.param_name || p.param_code }}</span>
                  </span>
                  <span class="tp-val">{{ formatParamValue(p) }}</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Compact Icon-Only Chart Tools Capsule -->
          <div class="icon-seg-capsule">
            <!-- Chart Settings & Scale -->
            <div class="relative trend-settings-anchor">
              <button
                type="button"
                @click="bentoActiveMiniPopover = null; showTrendSettingsMenu = !showTrendSettingsMenu"
                class="isc-btn"
                :class="{ active: showTrendSettingsMenu }"
                title="Pengaturan Grafik & Skala (Dual-Axis / Norm % / Wind & Map)"
                aria-label="Chart Settings"
              >
                <OperatorDashboardIcon name="settings" :size="13" />
              </button>
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
                    <!-- 1. Mode: Processed vs Raw (Admin & Superadmin only) -->
                    <div v-if="canAccessRawTrendMode">
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
                        <span>{{ canAccessRawTrendMode ? '2. Scale' : '1. Scale' }}</span>
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

            <!-- Line Smooth / Sharp -->
            <div class="relative trend-line-smooth-anchor">
              <button
                type="button"
                @click.stop="toggleLineInterpolationMenu"
                class="isc-btn"
                :class="{ active: showLineSmoothMenu || trendLineSmooth }"
                :title="trendLineSmooth ? 'Mode Garis: Halus / Smooth' : 'Mode Garis: Runcing / Linear'"
                aria-label="Line Interpolation Mode"
              >
                <OperatorDashboardIcon :name="trendLineSmooth ? 'line-smooth' : 'line-sharp'" :size="13" />
              </button>
                <div
                  v-if="showLineSmoothMenu"
                  class="absolute left-0 top-full mt-1.5 z-50 w-56 rounded-lg shadow-2xl p-2.5 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#0E1726] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono text-[9.5px]"
                  @click.stop
                >
                  <div class="flex items-center justify-between pb-1.5 mb-2 border-b border-slate-100 dark:border-[#1E3048]">
                    <div class="flex items-center space-x-1.5 font-bold uppercase tracking-wider text-[10px] text-slate-900 dark:text-white">
                      <OperatorDashboardIcon :name="trendLineSmooth ? 'line-smooth' : 'line-sharp'" :size="13" class="text-cyan-500" />
                      <span>{{ $t('operator.lineStyleSetup') || 'Line Style Setup' }}</span>
                    </div>
                    <button
                      type="button"
                      @click="showLineSmoothMenu = false"
                      class="text-slate-400 hover:text-slate-600 dark:hover:text-white text-[11px] p-0.5 leading-none cursor-pointer"
                    >✕</button>
                  </div>

                  <div class="text-[8.5px] text-slate-500 dark:text-slate-400 mb-2 leading-tight">
                    {{ $t('operator.lineStyleDesc') || 'Pilih mode interpolasi garis grafik trend telemetry:' }}
                  </div>

                  <!-- Mode Selection Options -->
                  <div class="space-y-1.5 mb-2">
                    <!-- Option 1: Smooth / Lengkung -->
                    <button
                      type="button"
                      @click="setTrendLineSmooth(true)"
                      class="w-full p-2 rounded-md border flex items-center justify-between transition cursor-pointer text-left"
                      :class="trendLineSmooth
                        ? 'border-cyan-500/70 bg-cyan-500/15 text-cyan-700 dark:text-cyan-300 font-bold ring-1 ring-cyan-500/30'
                        : 'border-slate-200 dark:border-[#1E3048] bg-slate-50 dark:bg-[#121E30] text-slate-600 dark:text-slate-300 hover:border-slate-300 dark:hover:border-slate-600'"
                    >
                      <div class="flex items-center space-x-2 min-w-0">
                        <div class="w-6 h-6 rounded flex items-center justify-center shrink-0" :class="trendLineSmooth ? 'bg-cyan-500/20 text-cyan-500' : 'bg-slate-200/60 dark:bg-slate-800 text-slate-500'">
                          <OperatorDashboardIcon name="line-smooth" :size="14" />
                        </div>
                        <div class="min-w-0 truncate">
                          <div class="text-[10px] leading-tight">{{ $t('operator.lineSmoothTitle') || 'Smooth (Lengkung)' }}</div>
                          <div class="text-[8px] font-normal opacity-75 leading-tight">Monotone cubic spline</div>
                        </div>
                      </div>
                      <span v-if="trendLineSmooth" class="text-cyan-500 text-[10px]">✓</span>
                    </button>

                    <!-- Option 2: Runcing / Sharp / Linear -->
                    <button
                      type="button"
                      @click="setTrendLineSmooth(false)"
                      class="w-full p-2 rounded-md border flex items-center justify-between transition cursor-pointer text-left"
                      :class="!trendLineSmooth
                        ? 'border-cyan-500/70 bg-cyan-500/15 text-cyan-700 dark:text-cyan-300 font-bold ring-1 ring-cyan-500/30'
                        : 'border-slate-200 dark:border-[#1E3048] bg-slate-50 dark:bg-[#121E30] text-slate-600 dark:text-slate-300 hover:border-slate-300 dark:hover:border-slate-600'"
                    >
                      <div class="flex items-center space-x-2 min-w-0">
                        <div class="w-6 h-6 rounded flex items-center justify-center shrink-0" :class="!trendLineSmooth ? 'bg-cyan-500/20 text-cyan-500' : 'bg-slate-200/60 dark:bg-slate-800 text-slate-500'">
                          <OperatorDashboardIcon name="line-sharp" :size="14" />
                        </div>
                        <div class="min-w-0 truncate">
                          <div class="text-[10px] leading-tight">{{ $t('operator.lineSharpTitle') || 'Runcing (Linear)' }}</div>
                          <div class="text-[8px] font-normal opacity-75 leading-tight">Direct sharp segments</div>
                        </div>
                      </div>
                      <span v-if="!trendLineSmooth" class="text-cyan-500 text-[10px]">✓</span>
                    </button>
                  </div>

                  <!-- Quick 1-Click Toggle Action in footer -->
                  <div class="pt-1.5 border-t border-slate-100 dark:border-[#1E3048] flex items-center justify-between text-[8px] text-slate-400">
                    <span>{{ trendLineSmooth ? 'Active: Smooth' : 'Active: Runcing' }}</span>
                    <button
                      type="button"
                      @click="toggleTrendLineSmooth"
                      class="text-cyan-600 dark:text-cyan-400 font-bold hover:underline cursor-pointer"
                    >
                      Ganti ({{ trendLineSmooth ? '→ Runcing' : '→ Smooth' }})
                    </button>
                  </div>
                </div>
            </div>

            <!-- 1-Click Compact Processed / Raw Toggle -->
            <button
              v-if="canAccessRawTrendMode"
              type="button"
              class="isc-text-btn"
              :class="{ active: chartMode === 'raw' }"
              @click="setChartMode(chartMode === 'raw' ? 'processed' : 'raw')"
              :title="chartMode === 'raw' ? 'Mode Data Mentah Sensor (Klik untuk Processed)' : 'Mode Data Terkalibrasi / Processed (Klik untuk Raw)'"
            >
              {{ chartMode === 'raw' ? 'RAW' : 'PROC' }}
            </button>
          </div>

          <!-- Compact Spike Zoom Pill (only when outlier spike present) -->
          <button
            v-if="hasOutlierSpike && trendScaleMode === 'auto'"
            type="button"
            @click="toggleTrendSpikeZoom"
            class="diag-pill cursor-pointer"
            :title="activeSpikeSummary"
          >
            <span>⚡ {{ trendSpikeZoom === 'full' ? 'Focus' : 'Full' }}</span>
          </button>
        </div>

        <!-- Right Cluster: Subtle Dual-Axis Hint, Timeframes + Icon Period, Height Adjuster, Download, Fullscreen -->
        <div class="ts-c-right">
          <div v-if="activeTrendSeries.length > 0" class="axis-legend hidden xl:flex">
            <span v-if="trendScales.left" :style="{ color: trendScales.left.color }">
              L: {{ trendScales.left.label }}
            </span>
            <span v-if="trendScales.isDual && trendScales.right" :style="{ color: trendScales.right.color }">
              R: {{ trendScales.right.label }}
            </span>
          </div>

          <!-- Compact Timeframe Capsule + Icon-Only Custom Period -->
          <div class="capsule">
            <button
              v-for="tf in ['1m', '5m', '15m', '1h', '6h']"
              :key="'bento-tf-' + tf"
              type="button"
              class="cap-btn compact-tf"
              :class="{ active: currentTimeframe === tf && !fullscreenIsCustom }"
              @click="setTimeframe(tf)"
            >
              {{ tf }}
            </button>
            <button
              type="button"
              class="cap-btn compact-tf"
              :class="{ active: fullscreenIsCustom }"
              @click="openCustomTrendModal"
              title="Pilih & Unduh Grafik Periode Khusus"
              aria-label="Periode Khusus"
            >
              <OperatorDashboardIcon name="calendar" :size="12" />
            </button>
          </div>

          <!-- User-Adjustable Trend Height Icon + Popover (Saved to localStorage) -->
          <div class="relative bento-mini-popover-anchor">
            <button
              id="btn-bento-trend-height"
              type="button"
              @click="toggleBentoMiniPopover('trendHeight')"
              class="op-icon-btn xs"
              :class="{ 'is-active': bentoActiveMiniPopover === 'trendHeight' || isResizingBentoTrend }"
              :title="'Atur Tinggi Analisis Tren (' + bentoTrendCanvasHeight + 'px)'"
              aria-label="Atur Tinggi Grafik Tren"
            >
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="12" y1="3" x2="12" y2="21" />
                <polyline points="8 7 12 3 16 7" />
                <polyline points="8 17 12 21 16 17" />
                <line x1="4" y1="3" x2="20" y2="3" stroke-opacity="0.45" />
                <line x1="4" y1="21" x2="20" y2="21" stroke-opacity="0.45" />
              </svg>
            </button>

            <div
              v-if="bentoActiveMiniPopover === 'trendHeight'"
              class="bento-mini-popover align-right w-60"
            >
              <div class="bmp-head">
                <span class="bmp-title">Tinggi Grafik Tren</span>
                <span class="bmp-badge">{{ bentoTrendCanvasHeight }} px</span>
              </div>

              <div class="bmp-section">
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    class="bmp-step-btn"
                    @click="stepBentoTrendHeight(-25)"
                    title="Kurangi tinggi (-25px)"
                  >
                    −
                  </button>
                  <input
                    id="input-bento-trend-height"
                    type="range"
                    min="120"
                    max="560"
                    step="5"
                    :value="bentoTrendCanvasHeight"
                    @input="setBentoTrendHeight($event.target.value)"
                    class="trend-height-slider flex-1"
                  />
                  <button
                    type="button"
                    class="bmp-step-btn"
                    @click="stepBentoTrendHeight(25)"
                    title="Tambah tinggi (+25px)"
                  >
                    +
                  </button>
                </div>
              </div>

              <div class="bmp-section">
                <div class="bmp-label">Preset Ukuran</div>
                <div class="bmp-seg-grid cols-4">
                  <button
                    v-for="preset in [
                      { label: '140px', val: 140 },
                      { label: '205px', val: 205 },
                      { label: '300px', val: 300 },
                      { label: '420px', val: 420 }
                    ]"
                    :key="'th-p-' + preset.val"
                    type="button"
                    class="bmp-seg-btn"
                    :class="{ active: Math.abs(bentoTrendCanvasHeight - preset.val) <= 5 }"
                    @click="setBentoTrendHeight(preset.val)"
                  >
                    {{ preset.label }}
                  </button>
                </div>
              </div>

              <div class="bmp-hint">
                Tersimpan otomatis di browser. Anda juga dapat menarik garis bawah grafik.
              </div>
            </div>
          </div>

          <!-- Icon-Only Download Graphic / Data Dropdown -->
          <div class="relative trend-download-anchor">
            <button
              type="button"
              @click.stop="toggleTrendDownloadMenu"
              class="op-icon-btn xs"
              :class="{ 'is-active': showTrendDownloadMenu }"
              title="Unduh Grafik (PNG/SVG/PDF) atau Data (XLS/CSV/JSON)"
              aria-label="Download Chart or Data"
            >
              <OperatorDashboardIcon name="download" :size="13" />
            </button>
                <div
                  v-if="showTrendDownloadMenu"
                  @click.stop
                  class="absolute right-0 mt-1 w-56 bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#1E3048] rounded-xl shadow-2xl py-1.5 z-50 text-3xs font-mono select-none"
                >
                  <!-- Option 1: Custom Period Graphic Download -->
                  <button
                    type="button"
                    @click="openCustomTrendModal(); showTrendDownloadMenu = false"
                    class="w-full text-left px-3 py-1.5 hover:bg-cyan-500/10 dark:hover:bg-cyan-500/15 flex items-center space-x-2 text-cyan-600 dark:text-cyan-400 font-bold transition cursor-pointer border-b border-slate-100 dark:border-[#1E3048]"
                  >
                    <OperatorDashboardIcon name="calendar" :size="13" class="text-cyan-500" />
                    <span>Download Grafik Periode...</span>
                  </button>

                  <!-- Option 2: Current PNG Graphic -->
                  <button
                    type="button"
                    @click="downloadTrendChart('png'); showTrendDownloadMenu = false"
                    class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer"
                  >
                    <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
                    <span>Unduh Grafik (PNG)</span>
                  </button>

                  <!-- Option 3: Current SVG Graphic -->
                  <button
                    type="button"
                    @click="downloadTrendChart('svg'); showTrendDownloadMenu = false"
                    class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer"
                  >
                    <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
                    <span>Unduh Grafik (SVG)</span>
                  </button>

                  <!-- Option 4: Current PDF Report -->
                  <button
                    type="button"
                    @click="downloadTrendPDF(); showTrendDownloadMenu = false"
                    class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer"
                  >
                    <OperatorDashboardIcon name="report" :size="13" class="text-rose-500 dark:text-rose-400" />
                    <span>Unduh Laporan (PDF)</span>
                  </button>

                  <div class="my-1 border-t border-slate-100 dark:border-[#1E3048]"></div>

                  <!-- Option 5: Excel XLS -->
                  <button
                    type="button"
                    @click="downloadTrendXLS(); showTrendDownloadMenu = false"
                    class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer"
                  >
                    <OperatorDashboardIcon name="export" :size="13" class="text-emerald-500 dark:text-emerald-400" />
                    <span>Unduh Data Excel (XLS)</span>
                  </button>

                  <!-- Option 6: CSV Data -->
                  <button
                    type="button"
                    @click="downloadTrendCSV(); showTrendDownloadMenu = false"
                    class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer"
                  >
                    <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
                    <span>Data Telemetri (CSV)</span>
                  </button>

                  <!-- Option 7: JSON Data -->
                  <button
                    type="button"
                    @click="downloadTrendJSON(); showTrendDownloadMenu = false"
                    class="w-full text-left px-3 py-1.5 hover:bg-slate-100 dark:hover:bg-[#16253B] flex items-center space-x-2 text-slate-700 dark:text-slate-300 transition cursor-pointer"
                  >
                    <OperatorDashboardIcon name="download" :size="13" class="text-cyan-500" />
                    <span>Data Telemetri (JSON)</span>
                  </button>
                </div>
          </div>

          <!-- Icon-Only Expand Fullscreen Chart -->
          <button
            type="button"
            class="op-icon-btn xs"
            @click="openTrendFullscreen"
            title="Perbesar Grafik Layar Penuh"
            aria-label="Expand Fullscreen Chart"
          >
            <OperatorDashboardIcon name="fullscreen" :size="13" />
          </button>

          <!-- Icon-Only Hide / Minimize Trend Chart (Expands Parameter Area to Full View) -->
          <button
            id="btn-bento-trend-hide"
            type="button"
            class="op-icon-btn xs"
            @click="setPerfSetting('trendAnalysisEnabled', false)"
            title="Sembunyikan Grafik Tren (Tampilkan Semua Tabel/Kartu Penuh)"
            aria-label="Hide Trend Chart"
          >
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.1" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="6 9 12 15 18 9" />
              <line x1="5" y1="19" x2="19" y2="19" stroke-opacity="0.5" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Calm Breathable Dual-Axis SVG Chart (User-Adjustable Height persisted in localStorage) -->
      <div
        ref="scadaChartContainer"
        class="bento-chart-canvas"
        :style="{ height: bentoTrendCanvasHeight + 'px' }"
        @mousemove="onChartScrub"
        @mouseleave="scrubPoint = null"
      >
        <div v-if="isResizingBentoTrend" class="trend-resize-live-badge">
          Tinggi Grafik: {{ bentoTrendCanvasHeight }} px
        </div>
        <div v-if="isChartLoading" class="bento-chart-overlay">
          <span class="animate-spin w-4 h-4 border-2 border-sky-500 border-t-transparent rounded-full mr-2"></span>
          <span>Memuat data tren telemetri...</span>
        </div>

        <div v-else-if="activeTrendSeries.length === 0" class="bento-chart-empty">
          Klik tombol ikon <strong>Filter Parameter</strong> di kiri atas grafik atau klik kartu parameter untuk menampilkan tren.
        </div>

        <div v-else-if="trendCommDiagnostics.state === 'API_ERROR' && !hasAnyTrendData" class="bento-chart-empty text-rose-500">
          <div class="font-bold mb-1">Gagal Memuat Data Tren</div>
          <div class="text-xs opacity-80 mb-2">{{ trendCommDiagnostics.message }}</div>
          <button type="button" @click="fetchTrendData(false)" class="dock-btn cyan">↻ Coba Lagi</button>
        </div>

        <div v-else-if="trendCommDiagnostics.state === 'NO_DATA'" class="bento-chart-empty">
          Belum ada data historis pada rentang <strong>{{ currentTimeframe }}</strong>.
        </div>

        <svg
          v-else
          ref="trendPanelSvg"
          class="w-full h-full block"
          :viewBox="`0 0 ${trendChartDims.width} ${trendChartDims.height}`"
        >
          <defs>
            <linearGradient id="bentoPrimaryFill" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" :stop-color="primarySeriesColor" :stop-opacity="isDark ? '0.20' : '0.14'" />
              <stop offset="100%" :stop-color="primarySeriesColor" stop-opacity="0.00" />
            </linearGradient>
          </defs>

          <!-- Calm Horizontal Grid Lines & Dual-Axis Labels -->
          <g class="bento-grid-lines">
            <g v-for="(tick, idx) in trendYTicks" :key="'byt-' + idx">
              <line
                :x1="trendChartDims.padLeft"
                :y1="tick.y"
                :x2="trendChartDims.width - trendChartDims.padRight"
                :y2="tick.y"
                :stroke="isDark ? '#1A2740' : '#E8EEF5'"
                stroke-width="1"
              />
              <text
                :x="trendChartDims.padLeft - 8"
                :y="tick.y + 3.5"
                text-anchor="end"
                font-size="10.5"
                font-family="'Plus Jakarta Sans', 'Inter', sans-serif"
                font-weight="600"
                :fill="trendScales.isDual && trendScales.left ? trendScales.left.color : (isDark ? '#94A3B8' : '#64748B')"
              >{{ tick.label }}</text>
            </g>

            <template v-if="trendScales.isDual">
              <g v-for="(tick, idx) in trendRightYTicks" :key="'bryt-' + idx">
                <text
                  :x="trendChartDims.width - trendChartDims.padRight + 8"
                  :y="tick.y + 3.5"
                  text-anchor="start"
                  font-size="10.5"
                  font-family="'Plus Jakarta Sans', 'Inter', sans-serif"
                  font-weight="600"
                  :fill="trendScales.right ? trendScales.right.color : (isDark ? '#94A3B8' : '#64748B')"
                >{{ tick.label }}</text>
              </g>
            </template>
          </g>

          <!-- X-Axis Time Ticks -->
          <g class="bento-time-ticks">
            <text
              v-for="(tick, idx) in trendXTicks"
              :key="'bxt-' + idx"
              :x="tick.x"
              :y="trendChartDims.baseY + 16"
              text-anchor="middle"
              font-size="10.5"
              font-family="'Plus Jakarta Sans', 'Inter', sans-serif"
              font-weight="500"
              :fill="isDark ? '#64748B' : '#94A3B8'"
            >{{ tick.label }}</text>
          </g>

          <!-- Threshold Reference Lines -->
          <g v-if="trendThresholdLines.crit != null">
            <line
              :x1="trendChartDims.padLeft"
              :y1="trendThresholdLines.crit.y"
              :x2="trendChartDims.width - trendChartDims.padRight"
              :y2="trendThresholdLines.crit.y"
              stroke="#EF4444"
              stroke-width="1.2"
              stroke-dasharray="4 4"
              opacity="0.75"
            />
            <text
              :x="trendChartDims.width - trendChartDims.padRight - 4"
              :y="trendThresholdLines.crit.y - 4"
              fill="#EF4444"
              font-size="9.5"
              text-anchor="end"
              font-family="'Plus Jakarta Sans', sans-serif"
              font-weight="700"
            >{{ trendThresholdLines.crit.label }}</text>
          </g>
          <g v-if="trendThresholdLines.warn != null">
            <line
              :x1="trendChartDims.padLeft"
              :y1="trendThresholdLines.warn.y"
              :x2="trendChartDims.width - trendChartDims.padRight"
              :y2="trendThresholdLines.warn.y"
              stroke="#F59E0B"
              stroke-width="1.2"
              stroke-dasharray="4 4"
              opacity="0.75"
            />
            <text
              :x="trendChartDims.width - trendChartDims.padRight - 4"
              :y="trendThresholdLines.warn.y - 4"
              fill="#F59E0B"
              font-size="9.5"
              text-anchor="end"
              font-family="'Plus Jakarta Sans', sans-serif"
              font-weight="700"
            >{{ trendThresholdLines.warn.label }}</text>
          </g>

          <!-- Communication Outage & Recovery Markers -->
          <g v-if="trendChartOutageMarkers.length > 0">
            <g
              v-for="marker in trendChartOutageMarkers"
              :key="'bom-' + marker.id"
              class="cursor-pointer"
              @mouseenter="hoveredOutageMarker = marker"
              @mouseleave="hoveredOutageMarker = null"
            >
              <line
                :x1="marker.x"
                :y1="trendChartDims.padTop"
                :x2="marker.x"
                :y2="trendChartDims.baseY"
                :stroke="marker.color"
                stroke-width="1.5"
                stroke-dasharray="4 4"
                opacity="0.75"
              />
              <rect
                :x="Math.min(Math.max(marker.x - 38, trendChartDims.padLeft), trendChartDims.width - trendChartDims.padRight - 76)"
                :y="trendChartDims.padTop"
                width="76"
                height="15"
                rx="4"
                :fill="isDark ? '#111A2C' : '#FFFFFF'"
                :stroke="marker.color"
                stroke-width="1"
              />
              <text
                :x="Math.min(Math.max(marker.x, trendChartDims.padLeft + 38), trendChartDims.width - trendChartDims.padRight - 38)"
                :y="trendChartDims.padTop + 10.5"
                text-anchor="middle"
                font-size="8.5"
                font-family="'Plus Jakarta Sans', sans-serif"
                font-weight="700"
                :fill="marker.color"
              >{{ marker.label }}</text>
            </g>
          </g>

          <!-- Primary Series Soft Area Fill -->
          <path
            v-if="activeTrendSeries.length > 0 && generatePrimaryAreaPath()"
            :d="generatePrimaryAreaPath()"
            fill="url(#bentoPrimaryFill)"
          />

          <!-- Smooth Multi-Series Strokes -->
          <g v-for="(s, sIdx) in activeTrendSeries" :key="'bplot-' + s.id">
            <path
              :d="generateChartPath(s)"
              fill="none"
              :stroke="s.color"
              :stroke-width="hoveredTrendSeriesId === s.id ? 2.8 : (sIdx === 0 ? 2.5 : 2.2)"
              :opacity="hoveredTrendSeriesId && hoveredTrendSeriesId !== s.id ? 0.35 : 1"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="transition-opacity duration-150 cursor-pointer"
              @mouseenter="hoveredTrendSeriesId = s.id"
              @mouseleave="hoveredTrendSeriesId = null"
            />
          </g>

          <!-- Interactive Crosshair & Hover Dots -->
          <g v-if="scrubPoint">
            <line
              :x1="scrubPoint.x"
              :y1="trendChartDims.padTop"
              :x2="scrubPoint.x"
              :y2="trendChartDims.baseY"
              :stroke="isDark ? '#64748B' : '#94A3B8'"
              stroke-width="1.2"
              stroke-dasharray="4 4"
              opacity="0.7"
            />
            <g v-for="sp in scrubPoint.allSeries" :key="'bscrub-' + sp.seriesId">
              <circle
                v-if="sp.plotY != null"
                :cx="scrubPoint.x"
                :cy="sp.plotY"
                r="4"
                :fill="isDark ? '#111A2C' : '#FFFFFF'"
                :stroke="sp.color"
                stroke-width="2.4"
              />
            </g>
          </g>
        </svg>

        <!-- Clean Bento Floating Tooltip -->
        <div
          v-if="scrubPoint"
          class="bento-scrub-tooltip"
          :style="{ left: tooltipLeftPos + 'px', top: '10px' }"
        >
          <div class="bento-tt-head">
            <span>Pukul {{ scrubPoint.time }} · {{ scrubPoint.date }}</span>
            <span class="bento-tt-tf">{{ currentTimeframe }}</span>
          </div>
          <div v-if="scrubPoint.inDataGap" class="text-[10px] font-bold text-rose-500 mb-1.5">
            ⚠️ Jeda Komunikasi (Tidak ada sampel)
          </div>
          <div class="space-y-1">
            <div
              v-for="sp in scrubPoint.allSeries"
              :key="'btt-' + sp.seriesId"
              class="flex items-center justify-between gap-3 text-[11px]"
            >
              <div class="flex items-center gap-1.5 min-w-0">
                <span class="w-2 h-2 rounded-full flex-shrink-0" :style="{ background: sp.color }"></span>
                <span class="truncate font-semibold">{{ sp.paramName || sp.paramCode }}</span>
              </div>
              <span class="font-bold tabular-nums flex-shrink-0">
                {{ sp.val }} <span class="text-[10px] font-normal opacity-75">{{ sp.unit }}</span>
              </span>
            </div>
          </div>
        </div>

        <!-- Outage Marker Hover Tooltip -->
        <div
          v-if="hoveredOutageMarker"
          class="bento-scrub-tooltip z-40"
          :style="{ left: Math.min(Math.max(hoveredOutageMarker.x - 100, 10), trendChartDims.width - 220) + 'px', top: '14px' }"
        >
          <div class="font-bold text-xs mb-1" :class="hoveredOutageMarker.isLost ? 'text-rose-500' : 'text-emerald-500'">
            {{ hoveredOutageMarker.label }} · {{ hoveredOutageMarker.timeFormatted }}
          </div>
          <div class="text-[11px] opacity-85">{{ hoveredOutageMarker.reason }}</div>
        </div>
      </div>

      <!-- Interactive Drag Handle to Adjust Trend Analysis Height (Saved to localStorage) -->
      <div
        class="trend-resize-handle"
        :class="{ active: isResizingBentoTrend }"
        @mousedown.prevent="startBentoTrendResize"
        @touchstart.passive="startBentoTrendResizeTouch"
        @dblclick="setBentoTrendHeight(205)"
        :title="'Tarik atas/bawah untuk mengatur tinggi grafik (' + bentoTrendCanvasHeight + 'px) · Klik 2x untuk reset'"
      >
        <span class="trend-resize-grip"></span>
      </div>
    </section>

    <!-- Trend Analysis Hidden / Paused Slim Bar -->
    <section v-else class="trend-studio is-collapsed-trend">
      <div class="flex items-center gap-2 min-w-0">
        <span class="ts-title">Analisis Tren Telemetri (Disembunyikan)</span>
        <span class="text-[11px] opacity-70 hidden sm:inline truncate">
          · Area tabel &amp; kartu parameter ditampilkan penuh (grafik SVG dijeda)
        </span>
      </div>
      <button
        id="btn-bento-trend-show"
        type="button"
        @click="setPerfSetting('trendAnalysisEnabled', true)"
        class="dock-btn cyan flex-shrink-0"
      >
        <span>▶ Aktifkan Grafik Tren</span>
      </button>
    </section>

    <!-- Optional Environmental Visualizations (Wind Rose & Dispersion Map) -->
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

    <!-- =================================================================== -->
    <!-- 4. DOCKED BOTTOM QUICK ACTIONS & SHORTCUTS BAR                      -->
    <!-- =================================================================== -->
    <footer class="dock-footer">
      <div class="dock-left">
        <span class="dock-kicker hidden sm:inline">AKSI CEPAT:</span>

        <button type="button" @click="openExportModal" class="dock-btn cyan">
          <OperatorDashboardIcon name="export" :size="12" />
          <span>{{ $t('operator.exportData') || 'Export Data' }}</span>
        </button>

        <button
          v-if="canAccessOperationalReport"
          type="button"
          @click="openReportModal"
          class="dock-btn teal"
        >
          <OperatorDashboardIcon name="report" :size="12" />
          <span>{{ $t('operator.generateReport') || 'Operational Report' }}</span>
        </button>

        <router-link to="/alarm/history" class="dock-btn hidden md:inline-flex">
          <OperatorDashboardIcon name="history" :size="12" />
          <span>{{ $t('operator.alarmHistory') || 'Alarm History' }}</span>
        </router-link>
      </div>

      <div class="dock-right">
        <router-link v-if="isNonOperator" to="/monitoring/devices" class="dock-btn hidden lg:inline-flex">
          {{ $t('operator.deviceMonitoring') || 'Devices' }}
        </router-link>
        <router-link v-if="isNonOperator" to="/monitoring/analysis" class="dock-btn hidden lg:inline-flex">
          {{ $t('operator.telemetryAnalysis') || 'Analysis' }}
        </router-link>
        <router-link v-if="isNonOperator" to="/monitoring/system" class="dock-btn hidden lg:inline-flex">
          {{ $t('operator.systemHealthAction') || 'System Health' }}
        </router-link>

        <div class="relative audio-setup-anchor flex items-center">
          <button
            type="button"
            @click="toggleAudioSetupPopover"
            class="sound-pill"
            :class="{ 'is-muted': isAudioMuted || !isSoundEnabled }"
          >
            <OperatorDashboardIcon :name="isAudioMuted ? 'sound-muted' : (isSoundEnabled ? 'sound-on' : 'sound-off')" :size="12" />
            <span>{{ isAudioMuted ? 'MUTED' : (isSoundEnabled ? 'SOUND ON' : 'SOUND OFF') }}</span>
            <svg width="10.5" height="10.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </button>
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
    </footer>

    <!-- =================================================================== -->
    <!-- 5. ALL OPERATIONAL MODALS & CONTEXT MENUS (100% PRESERVED)          -->
    <!-- =================================================================== -->
    <!-- MODAL: DUAL ORGANIZATION BRANDING SETUP (CUSTOMER & MAKER) -->
    <div
      v-if="showBrandingModal || showLogoModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 dark:bg-slate-950/80 backdrop-blur-sm font-sans"
      @click.self="closeBrandingModal"
    >
      <div class="w-full max-w-lg rounded-2xl bg-white dark:bg-slate-900 border border-slate-200/90 dark:border-slate-800 shadow-[0_20px_50px_-12px_rgba(15,23,42,0.28)] dark:shadow-[0_24px_56px_-12px_rgba(0,0,0,0.75)] text-slate-800 dark:text-slate-100 overflow-hidden flex flex-col max-h-[92vh]">
        <!-- Modal Header -->
        <div class="px-5 py-3.5 bg-slate-50/90 dark:bg-slate-800/80 border-b border-slate-200/80 dark:border-slate-800 flex items-center justify-between shrink-0">
          <div class="flex items-center space-x-2.5">
            <div class="w-8 h-8 rounded-xl bg-blue-50 dark:bg-blue-950/60 border border-blue-200/70 dark:border-blue-800/60 flex items-center justify-center text-[#1685E5] dark:text-blue-400">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
              </svg>
            </div>
            <div>
              <h3 class="text-[13.5px] font-semibold text-slate-900 dark:text-slate-100 tracking-tight leading-tight">
                {{ $t('branding.title') || 'Organization & Company Branding' }}
              </h3>
              <p class="text-[11px] text-slate-500 dark:text-slate-400">Identitas instansi & tipe lokasi stasiun</p>
            </div>
          </div>
          <button
            type="button"
            @click="closeBrandingModal"
            class="p-1.5 rounded-lg text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <!-- Scrollable Body -->
        <div class="p-5 space-y-4 overflow-y-auto flex-1 min-h-0">
          <!-- Error Message Banner -->
          <div v-if="brandingErrorMsg || logoErrorMsg" class="p-2.5 rounded-xl bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800/60 text-rose-700 dark:text-rose-300 text-xs font-medium flex items-center space-x-2">
            <svg class="w-4 h-4 shrink-0 text-rose-500" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
            </svg>
            <span>{{ brandingErrorMsg || logoErrorMsg }}</span>
          </div>

          <!-- Tab Selection (Customer Left Header vs Maker Right Header) -->
          <div class="grid grid-cols-2 gap-1 p-1 bg-slate-100/90 dark:bg-slate-800/90 rounded-xl border border-slate-200/80 dark:border-slate-700/70">
            <button
              type="button"
              @click="activeBrandingTab = 'customer'"
              class="py-1.5 px-3 rounded-lg text-xs font-semibold transition-all flex items-center justify-center space-x-1.5"
              :class="activeBrandingTab === 'customer'
                ? 'bg-[#1685E5] text-white btn-text shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'"
            >
              <svg class="w-3.5 h-3.5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5" />
              </svg>
              <span>{{ $t('branding.customerTab') || 'Customer (Left Header)' }}</span>
            </button>
            <button
              type="button"
              @click="activeBrandingTab = 'maker'"
              class="py-1.5 px-3 rounded-lg text-xs font-semibold transition-all flex items-center justify-center space-x-1.5"
              :class="activeBrandingTab === 'maker'
                ? 'bg-[#1685E5] text-white btn-text shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'"
            >
              <svg class="w-3.5 h-3.5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
              <span>{{ $t('branding.makerTab') || 'Maker / Provider (Right Header)' }}</span>
            </button>
          </div>

          <!-- TAB 1: CUSTOMER COMPANY SETUP -->
          <div v-if="activeBrandingTab === 'customer'" class="space-y-3.5">
            <div class="space-y-1">
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                {{ $t('branding.customerName') || 'Customer Company Name (Appears Top-Left)' }}
              </label>
              <input
                v-model="modalCustomerName"
                type="text"
                maxlength="60"
                placeholder="e.g. Kementerian Lingkungan Hidup"
                class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/70 dark:bg-slate-800/80 text-slate-900 dark:text-slate-100 text-[13px] font-medium focus:ring-2 focus:ring-[#1685E5]/30 focus:border-[#1685E5] focus:outline-none transition"
              />
            </div>

            <!-- Station Type / Monitoring Domain: Manual vs Auto GPS Place Name -->
            <div class="space-y-2">
              <div class="flex items-center justify-between gap-2">
                <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                  {{ $t('branding.stationDomain') || 'Station Type / Monitoring Domain' }}
                </label>
                <div class="flex items-center space-x-1">
                  <button
                    type="button"
                    v-for="st in ['AQMS', 'WQMS', 'Sparing']"
                    :key="st"
                    @click="selectModalDomainPrefix(st)"
                    class="px-2 py-0.5 rounded-md text-[11px] font-semibold transition border cursor-pointer"
                    :class="isModalDomainChipActive(st)
                      ? 'bg-[#1685E5] text-white btn-text border-[#1685E5] shadow-2xs'
                      : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700 hover:border-[#1685E5]/50 hover:text-[#1685E5] dark:hover:text-blue-400'"
                  >
                    {{ st }}
                  </button>
                </div>
              </div>

              <!-- 2-Mode Selector: Isi Manual vs Otomatis Lokasi GPS -->
              <div class="grid grid-cols-2 gap-1 p-1 bg-slate-100/90 dark:bg-slate-800/90 rounded-xl border border-slate-200/80 dark:border-slate-700/70">
                <button
                  type="button"
                  @click="setModalStationTypeMode('MANUAL')"
                  class="py-1.5 px-2.5 rounded-lg text-[11.5px] font-semibold transition-all flex items-center justify-center space-x-1.5"
                  :class="modalStationTypeMode !== 'AUTO_GPS'
                    ? 'bg-white dark:bg-slate-900 text-[#1685E5] dark:text-blue-400 shadow-xs border border-slate-200/80 dark:border-slate-700'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'"
                >
                  <svg class="w-3.5 h-3.5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                  <span>Isi Manual</span>
                </button>
                <button
                  type="button"
                  @click="setModalStationTypeMode('AUTO_GPS')"
                  class="py-1.5 px-2.5 rounded-lg text-[11.5px] font-semibold transition-all flex items-center justify-center space-x-1.5"
                  :class="modalStationTypeMode === 'AUTO_GPS'
                    ? 'bg-white dark:bg-slate-900 text-[#1685E5] dark:text-blue-400 shadow-xs border border-slate-200/80 dark:border-slate-700'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-slate-100'"
                >
                  <svg class="w-3.5 h-3.5 shrink-0" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                  <span>Otomatis Lokasi GPS</span>
                </button>
              </div>

              <!-- Mode 1: Manual Input -->
              <div v-if="modalStationTypeMode !== 'AUTO_GPS'">
                <input
                  v-model="modalCustomerSubtitle"
                  type="text"
                  maxlength="80"
                  placeholder="Contoh: AQMS Kutawaringin Industrial Park"
                  class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/70 dark:bg-slate-800/80 text-slate-900 dark:text-slate-100 text-[13px] font-medium focus:ring-2 focus:ring-[#1685E5]/30 focus:border-[#1685E5] focus:outline-none transition"
                />
              </div>

              <!-- Mode 2: Automatic Place Name from GPS (Nama Tempat, NOT Lat/Long) -->
              <div
                v-else
                class="p-3 rounded-xl border border-blue-200/80 dark:border-blue-800/50 bg-blue-50/50 dark:bg-slate-800/70 space-y-2"
              >
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-center space-x-1.5 min-w-0">
                    <span class="w-2 h-2 rounded-full shrink-0" :class="effectiveModalPlaceName ? 'bg-emerald-500' : 'bg-amber-500'"></span>
                    <span class="text-[11px] font-semibold text-slate-700 dark:text-slate-200 truncate">
                      Sumber: {{ activeAutoGpsSourceBadge }}
                    </span>
                  </div>
                  <button
                    type="button"
                    @click="refreshAutoGpsPlaceName(true)"
                    :disabled="isResolvingAutoPlace"
                    class="px-2 py-0.5 rounded-md text-[11px] font-semibold bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700 text-[#1685E5] dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-slate-800 disabled:opacity-50 transition flex items-center space-x-1 shrink-0"
                  >
                    <svg class="w-3 h-3" :class="{ 'animate-spin': isResolvingAutoPlace }" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    <span>{{ isResolvingAutoPlace ? 'Mencari...' : 'Perbarui Lokasi' }}</span>
                  </button>
                </div>

                <div class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900/90 border border-slate-200/80 dark:border-slate-700/80 flex items-center justify-between gap-2">
                  <div class="min-w-0">
                    <div class="text-[10.5px] font-medium text-slate-400 dark:text-slate-500">Pratinjau Tampilan Header</div>
                    <div class="text-[13px] font-semibold text-slate-900 dark:text-slate-100 truncate">
                      Stasiun <span class="text-[#1685E5] dark:text-blue-400">{{ modalAutoStationTypePreview }}</span>
                    </div>
                  </div>
                </div>

                <p class="text-[11px] leading-relaxed text-slate-500 dark:text-slate-400">
                  <template v-if="effectiveModalPlaceName">
                    Menampilkan <strong>nama tempat/wilayah</strong> otomatis dari titik GPS (bukan angka koordinat Lat/Long).
                  </template>
                  <template v-else>
                    Nama wilayah belum tersedia. Pastikan lokasi stasiun atau GPS Modbus RTU sudah aktif pada kontrol GPS.
                  </template>
                </p>
              </div>
            </div>

            <!-- Customer Logo Live Preview (Dark & Light) -->
            <div class="space-y-1.5">
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                Customer Logo Preview (Header Scale)
              </label>
              <div class="grid grid-cols-2 gap-2.5">
                <div class="p-3 rounded-xl bg-[#0B1422] border border-[#20354B] flex flex-col items-center justify-center min-h-[68px]">
                  <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1.5">Dark Header</span>
                  <img
                    v-if="modalCustomerLogoPreviewUrl || customerLogoUrl"
                    :src="modalCustomerLogoPreviewUrl || customerLogoUrl"
                    alt="Customer Logo preview"
                    class="max-h-7 max-w-[120px] object-contain"
                  />
                  <span v-else class="text-[11px] text-slate-500">Default CBI Icon</span>
                </div>
                <div class="p-3 rounded-xl bg-white border border-slate-200 flex flex-col items-center justify-center min-h-[68px]">
                  <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1.5">Light Header</span>
                  <img
                    v-if="modalCustomerLogoPreviewUrl || customerLogoUrl"
                    :src="modalCustomerLogoPreviewUrl || customerLogoUrl"
                    alt="Customer Logo preview"
                    class="max-h-7 max-w-[120px] object-contain"
                  />
                  <span v-else class="text-[11px] text-slate-400">Default CBI Icon</span>
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
            <div class="flex items-center space-x-2 pt-0.5">
              <button
                type="button"
                @click="triggerCustomerFileInput"
                class="flex-1 py-2 px-3 rounded-xl bg-slate-100 hover:bg-slate-200/80 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 font-semibold text-xs flex items-center justify-center space-x-1.5 transition"
              >
                <svg class="w-3.5 h-3.5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                </svg>
                <span>{{ (modalCustomerLogoPreviewUrl || customerLogoUrl) ? 'Change Customer Logo...' : 'Upload Customer Logo...' }}</span>
              </button>
              <button
                v-if="modalCustomerLogoPreviewUrl || customerLogoUrl"
                type="button"
                @click="handleRemoveCustomerLogo"
                class="py-2 px-3 rounded-xl bg-rose-50 hover:bg-rose-100 dark:bg-rose-950/40 dark:hover:bg-rose-900/50 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-800/60 font-semibold text-xs flex items-center justify-center space-x-1 transition"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                <span>Remove</span>
              </button>
            </div>
          </div>

          <!-- TAB 2: MAKER COMPANY SETUP -->
          <div v-else-if="activeBrandingTab === 'maker'" class="space-y-3.5">
            <div class="space-y-1">
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                {{ $t('branding.makerName') || 'Maker / Provider Company Name (Appears Top-Right)' }}
              </label>
              <input
                v-model="modalMakerName"
                type="text"
                maxlength="60"
                placeholder="e.g. PT. Centra Biotech Indonesia"
                class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/70 dark:bg-slate-800/80 text-slate-900 dark:text-slate-100 text-[13px] font-medium focus:ring-2 focus:ring-[#1685E5]/30 focus:border-[#1685E5] focus:outline-none transition"
              />
            </div>

            <div class="space-y-1">
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                {{ $t('branding.makerSubtitle') || 'Maker Tagline / Subtitle' }}
              </label>
              <input
                v-model="modalMakerSubtitle"
                type="text"
                maxlength="60"
                placeholder="e.g. CBI Industrial IoT"
                class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-50/70 dark:bg-slate-800/80 text-slate-900 dark:text-slate-100 text-[13px] font-medium focus:ring-2 focus:ring-[#1685E5]/30 focus:border-[#1685E5] focus:outline-none transition"
              />
            </div>

            <!-- Maker Logo Live Preview (Dark & Light) -->
            <div class="space-y-1.5">
              <label class="block text-xs font-semibold text-slate-700 dark:text-slate-300">
                {{ $t('branding.makerLogoPreview') || 'Maker Logo Preview (Header Scale)' }}
              </label>
              <div class="grid grid-cols-2 gap-2.5">
                <div class="p-3 rounded-xl bg-[#0B1422] border border-[#20354B] flex flex-col items-center justify-center min-h-[68px]">
                  <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1.5">Dark Header</span>
                  <img
                    v-if="modalMakerLogoPreviewUrl || makerLogoUrl"
                    :src="modalMakerLogoPreviewUrl || makerLogoUrl"
                    alt="Maker Logo preview"
                    class="max-h-7 max-w-[120px] object-contain"
                  />
                  <span v-else class="text-[11px] text-slate-500">No logo configured</span>
                </div>
                <div class="p-3 rounded-xl bg-white border border-slate-200 flex flex-col items-center justify-center min-h-[68px]">
                  <span class="text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1.5">Light Header</span>
                  <img
                    v-if="modalMakerLogoPreviewUrl || makerLogoUrl"
                    :src="modalMakerLogoPreviewUrl || makerLogoUrl"
                    alt="Maker Logo preview"
                    class="max-h-7 max-w-[120px] object-contain"
                  />
                  <span v-else class="text-[11px] text-slate-400">No logo configured</span>
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
            <div class="flex items-center space-x-2 pt-0.5">
              <button
                type="button"
                @click="triggerMakerFileInput"
                class="flex-1 py-2 px-3 rounded-xl bg-slate-100 hover:bg-slate-200/80 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border border-slate-200 dark:border-slate-700 font-semibold text-xs flex items-center justify-center space-x-1.5 transition"
              >
                <svg class="w-3.5 h-3.5 text-slate-500 dark:text-slate-400" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
                </svg>
                <span>{{ (modalMakerLogoPreviewUrl || makerLogoUrl) ? ($t('branding.changeMakerLogo') || 'Change Maker Logo...') : ($t('branding.uploadMakerLogo') || 'Upload Maker Logo...') }}</span>
              </button>
              <button
                v-if="modalMakerLogoPreviewUrl || makerLogoUrl"
                type="button"
                @click="handleRemoveMakerLogo"
                class="py-2 px-3 rounded-xl bg-rose-50 hover:bg-rose-100 dark:bg-rose-950/40 dark:hover:bg-rose-900/50 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-800/60 font-semibold text-xs flex items-center justify-center space-x-1 transition"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
                <span>{{ $t('branding.remove') || 'Remove' }}</span>
              </button>
            </div>
          </div>

          <!-- System Admin Configuration Deep Link -->
          <div v-if="canManageCategories" class="px-3 py-2 rounded-xl bg-slate-50 dark:bg-slate-800/60 border border-slate-200/80 dark:border-slate-700/70 flex items-center justify-between text-xs">
            <span class="text-slate-500 dark:text-slate-400">Advanced settings available in System Admin:</span>
            <router-link
              to="/administration/config?tab=branding"
              @click.native="closeBrandingModal"
              class="text-[#1685E5] dark:text-blue-400 font-semibold hover:underline flex items-center space-x-1"
            >
              <span>System Admin Setup →</span>
            </router-link>
          </div>
        </div>

        <!-- Footer Actions -->
        <div class="px-5 py-3.5 bg-slate-50/90 dark:bg-slate-800/80 border-t border-slate-200/80 dark:border-slate-800 flex items-center justify-end space-x-2 shrink-0">
          <button
            type="button"
            @click="closeBrandingModal"
            class="px-3.5 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 text-xs font-semibold text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 transition"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            type="button"
            @click="saveBrandingFromModal"
            :disabled="brandingSaving || logoSaving"
            class="px-4 py-2 rounded-xl bg-[#1685E5] hover:bg-[#1273C9] disabled:opacity-50 text-white btn-text font-semibold text-xs shadow-xs transition"
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
            <div class="grid gap-2 font-mono text-2xs" :class="canAccessRawTrendMode ? 'grid-cols-3' : 'grid-cols-2'">
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
                v-if="canAccessRawTrendMode"
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

          <!-- Parameter Selection (Multi-parameter filter) -->
          <div>
            <div class="flex items-center justify-between mb-1.5">
              <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                Parameter Selection
                <span class="ml-1 px-1.5 py-0.5 rounded text-3xs font-mono font-bold"
                  :class="exportForm.selectedParamIds.length === 0
                    ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/30'
                    : 'bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30'">
                  {{ exportForm.selectedParamIds.length === 0 ? 'Semua Parameter' : `${exportForm.selectedParamIds.length} Dipilih` }}
                </span>
              </label>
              <div class="flex items-center space-x-2 text-3xs font-mono">
                <button
                  type="button"
                  @click="selectAllExportParams"
                  class="text-cyan-600 dark:text-cyan-400 hover:underline hover:text-cyan-500 font-semibold"
                >
                  Pilih Semua
                </button>
                <span class="text-slate-300 dark:text-slate-700">|</span>
                <button
                  type="button"
                  @click="clearAllExportParams"
                  class="text-slate-500 dark:text-slate-400 hover:underline hover:text-rose-400 font-semibold"
                >
                  Kosongkan
                </button>
              </div>
            </div>

            <!-- Scrollable Parameter Checkboxes Grid -->
            <div class="max-h-36 overflow-y-auto p-2 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] grid grid-cols-2 gap-1.5 custom-scrollbar">
              <div
                v-for="param in availableExportParameters"
                :key="'exp-p-' + param.id"
                @click="toggleExportParam(param.id)"
                class="flex items-center space-x-2 p-1.5 rounded cursor-pointer transition select-none text-2xs"
                :class="isParamExportSelected(param.id)
                  ? 'bg-cyan-500/15 text-cyan-700 dark:text-cyan-300 border border-cyan-500/40 font-semibold shadow-xs'
                  : 'bg-white/40 dark:bg-[#0d1624]/60 text-slate-700 dark:text-slate-300 border border-slate-200/80 dark:border-[#1a2c3f] hover:border-slate-300 dark:hover:border-slate-600'"
              >
                <input
                  type="checkbox"
                  :checked="isParamExportSelected(param.id)"
                  @click.stop="toggleExportParam(param.id)"
                  class="rounded text-cyan-500 focus:ring-0 focus:outline-none pointer-events-none w-3.5 h-3.5"
                />
                <span class="truncate flex-1 font-mono text-2xs">{{ param.name }}</span>
                <span v-if="param.unit" class="text-3xs text-slate-500 dark:text-slate-400 shrink-0 font-mono">{{ param.unit }}</span>
              </div>
              <div v-if="availableExportParameters.length === 0" class="col-span-2 text-center py-3 text-slate-400 text-3xs font-mono">
                Tidak ada parameter aktif ditemukan
              </div>
            </div>
            <p class="text-3xs text-slate-500 dark:text-slate-400 mt-1 font-sans">
              * Mendukung multi-parameter. Jika tidak ada yang dicentang, sistem otomatis mengekspor seluruh parameter.
            </p>
          </div>

          <!-- Format Selection -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1.5">
              {{ $t('operator.selectFormat') || 'Export Format' }}
            </label>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 font-mono text-2xs">
              <!-- Excel XLS / XLSX -->
              <label
                @click="exportForm.format = 'xlsx'"
                class="flex items-center space-x-2 p-2 rounded-lg border cursor-pointer transition"
                :class="exportForm.format === 'xlsx'
                  ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/50 shadow-xs'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B] hover:border-slate-300 dark:hover:border-slate-600'"
              >
                <input type="radio" value="xlsx" v-model="exportForm.format" class="text-emerald-500 focus:ring-0" />
                <div class="flex flex-col">
                  <span class="font-bold text-xs">Excel (XLS)</span>
                  <span class="text-3xs opacity-75">.xlsx Sheet</span>
                </div>
              </label>

              <!-- PDF Document -->
              <label
                @click="exportForm.format = 'pdf'"
                class="flex items-center space-x-2 p-2 rounded-lg border cursor-pointer transition"
                :class="exportForm.format === 'pdf'
                  ? 'bg-rose-500/15 text-rose-700 dark:text-rose-300 border-rose-500/50 shadow-xs'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B] hover:border-slate-300 dark:hover:border-slate-600'"
              >
                <input type="radio" value="pdf" v-model="exportForm.format" class="text-rose-500 focus:ring-0" />
                <div class="flex flex-col">
                  <span class="font-bold text-xs">PDF</span>
                  <span class="text-3xs opacity-75">.pdf Report</span>
                </div>
              </label>

              <!-- CSV Spreadsheet -->
              <label
                @click="exportForm.format = 'csv'"
                class="flex items-center space-x-2 p-2 rounded-lg border cursor-pointer transition"
                :class="exportForm.format === 'csv'
                  ? 'bg-cyan-500/15 text-cyan-700 dark:text-cyan-300 border-cyan-500/50 shadow-xs'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B] hover:border-slate-300 dark:hover:border-slate-600'"
              >
                <input type="radio" value="csv" v-model="exportForm.format" class="text-cyan-500 focus:ring-0" />
                <div class="flex flex-col">
                  <span class="font-bold text-xs">CSV</span>
                  <span class="text-3xs opacity-75">.csv Comma</span>
                </div>
              </label>

              <!-- JSON Records -->
              <label
                @click="exportForm.format = 'json'"
                class="flex items-center space-x-2 p-2 rounded-lg border cursor-pointer transition"
                :class="exportForm.format === 'json'
                  ? 'bg-indigo-500/15 text-indigo-700 dark:text-indigo-300 border-indigo-500/50 shadow-xs'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B] hover:border-slate-300 dark:hover:border-slate-600'"
              >
                <input type="radio" value="json" v-model="exportForm.format" class="text-indigo-500 focus:ring-0" />
                <div class="flex flex-col">
                  <span class="font-bold text-xs">JSON</span>
                  <span class="text-3xs opacity-75">.json Records</span>
                </div>
              </label>
            </div>
          </div>

          <!-- Export Limit Governance (Auto-Pagination Setup) -->
          <div>
            <div class="flex items-center justify-between mb-1.5">
              <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                {{ canConfigureExportLimit ? 'Batas Maksimal Data (Setup Auto-Pagination)' : 'Batas Unduhan Data' }}
              </label>
              <button
                v-if="canConfigureExportLimit"
                type="button"
                @click="saveAsSystemDefaultExportLimit"
                class="text-3xs text-cyan-600 dark:text-cyan-400 hover:underline font-semibold font-mono"
                title="Simpan batas ini sebagai setup default sistem untuk semua pengguna/operator"
              >
                💾 Simpan Setup Sistem
              </button>
            </div>

            <!-- If Role is Admin/Engineer (Can Setup) -->
            <div v-if="canConfigureExportLimit" class="space-y-1">
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <select
                  v-model.number="exportForm.maxRecords"
                  @change="onExportLimitChanged"
                  class="w-full px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] text-slate-900 dark:text-slate-200 text-xs focus:outline-none focus:border-cyan-500 font-mono"
                >
                  <option v-for="preset in exportLimitPresets" :key="'pres-' + preset.value" :value="preset.value">
                    {{ preset.label }}
                  </option>
                </select>
                <div class="flex items-center text-3xs text-slate-500 dark:text-slate-400 font-mono">
                  Auto-pagination aktif (@ 1.000 data/halaman).
                </div>
              </div>
              <p v-if="exportSettingsFeedback" class="text-3xs text-emerald-600 dark:text-emerald-400 font-mono">
                ✓ {{ exportSettingsFeedback }}
              </p>
            </div>

            <!-- If Role is Operator (Read-Only Bound to Setup) -->
            <div v-else class="p-2 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] flex items-center justify-between text-2xs font-mono">
              <div class="flex items-center space-x-2">
                <span class="text-amber-500 text-xs">🔒</span>
                <span class="text-slate-700 dark:text-slate-300">
                  Batas Sistem: <strong class="text-cyan-600 dark:text-cyan-400 font-bold">{{ currentExportLimitLabel }}</strong>
                </span>
              </div>
              <span class="text-3xs text-slate-400 dark:text-slate-500">
                Setup oleh Admin/Engineer
              </span>
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
            @click="initiateExportWithUsbCheck"
            :disabled="isExporting"
            class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50 flex items-center space-x-1.5"
          >
            <span v-if="isExporting" class="animate-spin w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full"></span>
            <span>{{ isExporting ? 'Generating...' : ($t('operator.downloadExport') || 'Download Export') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL: USB FLASHDISK NOT DETECTED NOTIFICATION -->
    <div
      v-if="showUsbWarningModal"
      class="fixed inset-0 flex items-center justify-center p-4 bg-slate-900/80 dark:bg-slate-950/90 backdrop-blur-sm animate-fade-in"
      style="z-index: 99999;"
    >
      <div class="w-full max-w-md rounded-2xl bg-white dark:bg-[#0B1422] border border-amber-500/50 p-5 shadow-2xl text-slate-900 dark:text-slate-100 space-y-4 font-mono text-xs">
        <div class="flex items-center space-x-3 border-b border-slate-200 dark:border-[#20354B] pb-3">
          <div class="w-10 h-10 rounded-xl bg-amber-500/15 border border-amber-500/40 flex items-center justify-center text-xl shrink-0">
            🔌
          </div>
          <div>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans">
              Tidak Ada USB Drive Terdeteksi
            </h3>
            <p class="text-3xs text-slate-500 dark:text-slate-400 font-sans mt-0.5">
              Peringatan Penyimpanan Eksternal
            </p>
          </div>
        </div>

        <div class="p-3.5 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-700/40 text-amber-900 dark:text-amber-200 text-xs font-sans leading-relaxed space-y-2">
          <div>
            Sistem tidak mendeteksi adanya USB flashdisk yang terpasang pada port fisik datalogger.
          </div>
          <p class="font-semibold text-slate-800 dark:text-slate-100">
            Apakah Anda ingin tetap melanjutkan ekspor data dan menyimpannya di drive lokal (Local Drive)?
          </p>
          <div class="pt-2 border-t border-amber-200/50 dark:border-amber-800/40 flex items-start space-x-2 text-3xs text-amber-700 dark:text-amber-300/90 font-mono">
            <span class="text-xs">📁</span>
            <span>Jendela File Explorer ("Save As") akan terbuka untuk memilih folder penyimpanan di komputer Anda.</span>
          </div>
        </div>

        <div class="flex flex-col sm:flex-row items-center justify-end gap-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
          <button
            @click="showUsbWarningModal = false"
            class="w-full sm:w-auto px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition"
          >
            Batal (Pasang Flashdisk)
          </button>
          <button
            @click="retryUsbDetection"
            :disabled="isRetryingUsb"
            class="w-full sm:w-auto px-3 py-1.5 rounded-lg border border-cyan-500/40 text-cyan-600 dark:text-cyan-400 hover:bg-cyan-500/10 text-xs font-semibold flex items-center justify-center space-x-1 transition"
          >
            <span v-if="isRetryingUsb" class="animate-spin w-3 h-3 border-2 border-cyan-400 border-t-transparent rounded-full mr-1"></span>
            <span>🔄 Cek Ulang USB</span>
          </button>
          <button
            @click="proceedExportToLocalDrive"
            class="w-full sm:w-auto px-4 py-1.5 rounded-lg bg-amber-600 hover:bg-amber-500 text-white font-bold text-xs shadow-md transition flex items-center justify-center space-x-1"
          >
            <span>💾 Lanjut ke Local Drive</span>
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

        <!-- Status / Progress message -->
        <div
          v-if="reportStatusMessage"
          class="p-2.5 rounded-lg bg-blue-500/10 border border-blue-500/30 text-blue-700 dark:text-blue-300 font-mono text-2xs flex items-center space-x-2"
        >
          <span v-if="isGeneratingReport" class="animate-spin w-3.5 h-3.5 border-2 border-blue-500 border-t-transparent rounded-full shrink-0"></span>
          <span>{{ reportStatusMessage }}</span>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
          <button
            @click="showReportModal = false"
            :disabled="isGeneratingReport"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white disabled:opacity-50 cursor-pointer"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            @click="executeReportGenerate"
            :disabled="isGeneratingReport"
            class="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-bold text-xs shadow-md transition disabled:opacity-50 flex items-center space-x-1.5 cursor-pointer"
          >
            <span v-if="isGeneratingReport" class="animate-spin w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full"></span>
            <span>{{ isGeneratingReport ? 'Generating...' : 'Generate & Download' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- ======================================================== -->
    <!-- MODAL: DOWNLOAD GRAFIK TELEMETRI (PERIODE KHUSUS)        -->
    <!-- ======================================================== -->
    <div
      v-if="showCustomTrendModal"
      class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-slate-900/70 dark:bg-slate-950/85 backdrop-blur-sm animate-fade-in"
      @keydown.esc="showCustomTrendModal = false"
    >
      <div class="w-full max-w-xl max-h-[92vh] flex flex-col rounded-2xl bg-white dark:bg-[#0B1422] border border-cyan-500/40 p-5 shadow-2xl text-slate-900 dark:text-slate-100 font-mono text-xs overflow-hidden">
        <!-- Header -->
        <div class="flex items-center justify-between border-b border-slate-200 dark:border-[#20354B] pb-3 shrink-0">
          <div class="flex items-center space-x-2.5">
            <span class="w-8 h-8 rounded-lg bg-cyan-500/15 border border-cyan-500/30 flex items-center justify-center text-base text-cyan-500 shrink-0">
              <OperatorDashboardIcon name="calendar" :size="16" />
            </span>
            <div>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white font-sans leading-tight">
                Download Grafik Telemetri (Periode Khusus)
              </h3>
              <p class="text-3xs text-slate-500 dark:text-slate-400 font-sans mt-0.5">
                Tentukan rentang tanggal/waktu, parameter sensor, dan format berkas grafik yang diinginkan
              </p>
            </div>
          </div>
          <button
            type="button"
            @click="showCustomTrendModal = false"
            class="text-slate-400 hover:text-slate-600 dark:hover:text-white transition p-1 cursor-pointer"
            title="Tutup (Esc)"
          >✕</button>
        </div>

        <!-- Form Body (Scrollable) -->
        <div class="space-y-3.5 font-sans text-xs py-3 overflow-y-auto pr-1">
          <!-- 1. Quick Presets -->
          <div>
            <div class="flex items-center justify-between mb-1.5">
              <label class="text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider font-mono">
                1. Preset Waktu Cepat
              </label>
              <span class="text-[9px] text-cyan-600 dark:text-cyan-400 font-mono">Klik untuk set waktu otomatis</span>
            </div>
            <div class="grid grid-cols-4 sm:grid-cols-7 gap-1.5 font-mono text-2xs">
              <button
                type="button"
                v-for="p in [
                  { id: '1h', label: '1 Jam' },
                  { id: '3h', label: '3 Jam' },
                  { id: '6h', label: '6 Jam' },
                  { id: '12h', label: '12 Jam' },
                  { id: 'today', label: 'Hari Ini' },
                  { id: '24h', label: '24 Jam' },
                  { id: '7d', label: '7 Hari' },
                ]"
                :key="'ctf-' + p.id"
                @click="applyCustomTrendPreset(p.id)"
                class="px-2 py-1.5 rounded-lg border font-semibold transition text-center cursor-pointer"
                :class="customTrendForm.preset === p.id
                  ? 'bg-cyan-500 text-slate-950 border-cyan-400 font-bold shadow-xs'
                  : 'bg-slate-50 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 border-slate-200 dark:border-[#20354B] hover:border-cyan-500/50'"
              >
                {{ p.label }}
              </button>
            </div>
          </div>

          <!-- 2. Custom Date Range Pickers -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1.5 font-mono">
              2. Rentang Tanggal &amp; Waktu
            </label>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
              <div class="p-2 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] space-y-1">
                <span class="text-[9px] text-slate-500 dark:text-slate-400 uppercase font-mono block">Waktu Mulai:</span>
                <input
                  type="datetime-local"
                  v-model="customTrendForm.startTime"
                  class="w-full bg-transparent text-slate-900 dark:text-white text-xs font-mono border-none focus:outline-none focus:ring-0 p-0"
                />
              </div>
              <div class="p-2 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] space-y-1">
                <span class="text-[9px] text-slate-500 dark:text-slate-400 uppercase font-mono block">Waktu Selesai:</span>
                <input
                  type="datetime-local"
                  v-model="customTrendForm.endTime"
                  class="w-full bg-transparent text-slate-900 dark:text-white text-xs font-mono border-none focus:outline-none focus:ring-0 p-0"
                />
              </div>
            </div>
          </div>

          <!-- 3. Parameter Selection -->
          <div>
            <div class="flex items-center justify-between mb-1.5 font-mono">
              <label class="text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                3. Parameter Ditampilkan ({{ customTrendForm.selectedParamIds.length }}/{{ allParameters.length }})
              </label>
              <div class="flex items-center space-x-2 text-[9.5px]">
                <button
                  type="button"
                  @click="toggleSelectAllCustomTrendParams"
                  class="text-cyan-600 dark:text-cyan-400 hover:underline cursor-pointer font-bold"
                >
                  {{ customTrendForm.selectedParamIds.length === allParameters.length ? 'Batal Semua' : 'Pilih Semua' }}
                </button>
                <span class="text-slate-400">·</span>
                <button
                  type="button"
                  @click="selectPlottedCustomTrendParams"
                  class="text-cyan-600 dark:text-cyan-400 hover:underline cursor-pointer font-bold"
                >
                  Sesuai Grafik Aktif
                </button>
              </div>
            </div>

            <!-- Parameters Grid / Checklist -->
            <div class="max-h-36 overflow-y-auto p-2 rounded-lg bg-slate-50 dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B] grid grid-cols-1 sm:grid-cols-2 gap-1.5 font-mono text-2xs">
              <label
                v-for="(p, pIdx) in allParameters"
                :key="'ctparam-' + p.param_id"
                class="flex items-center space-x-2 p-1 rounded hover:bg-slate-200/60 dark:hover:bg-[#16253B] transition cursor-pointer select-none"
              >
                <input
                  type="checkbox"
                  :value="p.param_id"
                  v-model="customTrendForm.selectedParamIds"
                  class="rounded text-cyan-600 focus:ring-cyan-500"
                />
                <span class="w-2 h-2 rounded-full shrink-0" :style="{ backgroundColor: getParamColorHelper(p, pIdx) }"></span>
                <span class="font-bold text-slate-900 dark:text-white truncate" :title="p.param_name">
                  {{ p.param_name }}
                </span>
                <span class="text-[8px] text-slate-500 dark:text-slate-400 ml-auto shrink-0 font-mono">
                  {{ p.unit || '' }}
                </span>
              </label>
            </div>
          </div>

          <!-- 4. Output Graphic Format -->
          <div>
            <label class="block text-3xs font-semibold text-slate-600 dark:text-slate-400 uppercase tracking-wider mb-1.5 font-mono">
              4. Format Grafik Berkas
            </label>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 font-mono text-2xs">
              <label
                class="flex flex-col items-center justify-center p-2.5 rounded-lg border cursor-pointer transition text-center space-y-1"
                :class="customTrendForm.format === 'png'
                  ? 'bg-cyan-500/15 border-cyan-500 text-cyan-700 dark:text-cyan-300 font-bold ring-1 ring-cyan-500/30'
                  : 'bg-slate-50 dark:bg-[#101B2B] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:border-slate-300'"
              >
                <input type="radio" value="png" v-model="customTrendForm.format" class="sr-only" />
                <span class="text-base">🖼️</span>
                <span class="text-[10px]">Gambar PNG</span>
                <span class="text-[8px] opacity-75">Resolusi Tinggi 2x</span>
              </label>

              <label
                class="flex flex-col items-center justify-center p-2.5 rounded-lg border cursor-pointer transition text-center space-y-1"
                :class="customTrendForm.format === 'svg'
                  ? 'bg-cyan-500/15 border-cyan-500 text-cyan-700 dark:text-cyan-300 font-bold ring-1 ring-cyan-500/30'
                  : 'bg-slate-50 dark:bg-[#101B2B] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:border-slate-300'"
              >
                <input type="radio" value="svg" v-model="customTrendForm.format" class="sr-only" />
                <span class="text-base">📐</span>
                <span class="text-[10px]">Vektor SVG</span>
                <span class="text-[8px] opacity-75">Vektor Skalabilitas</span>
              </label>

              <label
                class="flex flex-col items-center justify-center p-2.5 rounded-lg border cursor-pointer transition text-center space-y-1"
                :class="customTrendForm.format === 'pdf'
                  ? 'bg-cyan-500/15 border-cyan-500 text-cyan-700 dark:text-cyan-300 font-bold ring-1 ring-cyan-500/30'
                  : 'bg-slate-50 dark:bg-[#101B2B] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:border-slate-300'"
              >
                <input type="radio" value="pdf" v-model="customTrendForm.format" class="sr-only" />
                <span class="text-base">📑</span>
                <span class="text-[10px]">Dokumen PDF</span>
                <span class="text-[8px] opacity-75">Grafik &amp; Tabel Stats</span>
              </label>

              <label
                class="flex flex-col items-center justify-center p-2.5 rounded-lg border cursor-pointer transition text-center space-y-1"
                :class="customTrendForm.format === 'xlsx'
                  ? 'bg-emerald-500/15 border-emerald-500 text-emerald-700 dark:text-emerald-300 font-bold ring-1 ring-emerald-500/30'
                  : 'bg-slate-50 dark:bg-[#101B2B] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:border-slate-300'"
              >
                <input type="radio" value="xlsx" v-model="customTrendForm.format" class="sr-only" />
                <span class="text-base">📊</span>
                <span class="text-[10px]">Excel (XLS)</span>
                <span class="text-[8px] opacity-75">Data &amp; Ringkasan</span>
              </label>
            </div>
          </div>

          <!-- 5. Options -->
          <div class="pt-1">
            <label class="flex items-center space-x-2 text-2xs text-slate-700 dark:text-slate-300 font-mono cursor-pointer select-none">
              <input
                type="checkbox"
                v-model="customTrendForm.applyToMainChart"
                class="rounded text-cyan-600 focus:ring-cyan-500"
              />
              <span>Tampilkan juga rentang waktu ini pada grafik Trend utama di dashboard</span>
            </label>
          </div>

          <!-- Status Message / Spinner -->
          <div
            v-if="isCustomTrendExporting || customTrendStatusMsg"
            class="p-2.5 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-700 dark:text-cyan-300 font-mono text-2xs flex items-center space-x-2 animate-pulse"
          >
            <span v-if="isCustomTrendExporting" class="animate-spin w-3.5 h-3.5 border-2 border-cyan-500 border-t-transparent rounded-full shrink-0"></span>
            <span>{{ customTrendStatusMsg }}</span>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex items-center justify-end space-x-2 pt-3 border-t border-slate-200 dark:border-[#20354B] shrink-0 font-mono">
          <button
            type="button"
            @click="showCustomTrendModal = false"
            :disabled="isCustomTrendExporting"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-[#20354B] text-xs font-semibold text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition cursor-pointer"
          >
            {{ $t('common.cancel') || 'Batal' }}
          </button>
          <button
            type="button"
            @click="submitDownloadCustomTrendChart"
            :disabled="isCustomTrendExporting"
            class="px-4 py-1.5 rounded-lg bg-cyan-500 hover:bg-cyan-400 text-slate-950 font-bold text-xs shadow-md transition disabled:opacity-50 flex items-center space-x-1.5 cursor-pointer"
          >
            <span v-if="isCustomTrendExporting" class="animate-spin w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full"></span>
            <OperatorDashboardIcon v-else name="download" :size="13" />
            <span>Unduh Grafik Periode</span>
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
                <span>CBI TELEMETRY TREND ANALYZER</span>
                <span v-if="canAccessRawTrendMode" class="text-3xs font-semibold px-2 py-0.5 rounded-full bg-cyan-500/20 text-cyan-400 border border-cyan-500/30">
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
            <!-- Custom Period Graphic Modal -->
            <button
              @click="openCustomTrendModal"
              class="px-2.5 py-1.5 rounded-lg bg-cyan-500/15 hover:bg-cyan-500/25 border border-cyan-500/30 text-cyan-700 dark:text-cyan-300 text-3xs font-bold transition flex items-center space-x-1.5 cursor-pointer"
              title="Download Grafik Berdasarkan Periode Pilihan"
            >
              <OperatorDashboardIcon name="calendar" :size="13" class="text-cyan-500" />
              <span>Grafik Periode...</span>
            </button>

            <!-- PDF Report Download -->
            <button
              @click="downloadTrendPDF"
              class="px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#1E2E42] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-200 text-3xs font-bold transition flex items-center space-x-1.5 cursor-pointer"
              title="Download Laporan PDF Grafik & Statistik"
            >
              <OperatorDashboardIcon name="report" :size="13" class="text-rose-500 dark:text-rose-400" />
              <span>PDF</span>
            </button>

            <!-- Excel XLS Download -->
            <button
              @click="downloadTrendXLS"
              class="px-2.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#1E2E42] border border-slate-200 dark:border-[#20354B] text-slate-700 dark:text-slate-200 text-3xs font-bold transition flex items-center space-x-1.5 cursor-pointer"
              title="Download Data Telemetri Excel (.xlsx)"
            >
              <OperatorDashboardIcon name="export" :size="13" class="text-emerald-500 dark:text-emerald-400" />
              <span>Excel (XLS)</span>
            </button>

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

            <!-- Mode Selector (Admin & Superadmin only) -->
            <div v-if="canAccessRawTrendMode" class="flex items-center bg-white dark:bg-[#101B2B] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">
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

            <!-- Left Y-Axis Vertical Line (1.8px high contrast) -->
            <line
              :x1="fullscreenChartDims.padLeft"
              :y1="fullscreenChartDims.padTop"
              :x2="fullscreenChartDims.padLeft"
              :y2="fullscreenChartDims.baseY"
              :stroke="axisStroke"
              stroke-width="1.8"
            />

            <!-- Right Dual Y-Axis Vertical Line (1.8px, matching right series color) -->
            <line
              v-if="trendScales.isDual"
              :x1="fullscreenChartDims.width - fullscreenChartDims.padRight"
              :y1="fullscreenChartDims.padTop"
              :x2="fullscreenChartDims.width - fullscreenChartDims.padRight"
              :y2="fullscreenChartDims.baseY"
              :stroke="trendScales.right ? trendScales.right.color : axisStroke"
              stroke-width="1.8"
            />

            <!-- Y-Axis Grid Lines & Values -->
            <g class="grid-lines" :stroke="gridStroke" stroke-dasharray="3 3" stroke-width="0.8">
              <g v-for="(tick, idx) in fullscreenYTicks" :key="'fs-ytick-' + idx">
                <line :x1="fullscreenChartDims.padLeft" :y1="tick.y" :x2="fullscreenChartDims.width - fullscreenChartDims.padRight" :y2="tick.y" stroke-opacity="0.6" />
                <!-- Left Y-Axis Tick Protrusion (1.5px thick, 5px outward) -->
                <line :x1="fullscreenChartDims.padLeft - 5" :y1="tick.y" :x2="fullscreenChartDims.padLeft" :y2="tick.y" :stroke="axisStroke" stroke-width="1.5" stroke-dasharray="none" />
                <text
                  :x="fullscreenChartDims.padLeft - 8"
                  :y="tick.y + 3.5"
                  text-anchor="end"
                  font-size="11"
                  font-family="monospace"
                  font-weight="600"
                  :fill="isDark ? '#E2E8F0' : '#0F172A'"
                >{{ tick.label }}</text>
              </g>

              <!-- Dual-axis Right Ticks (when active series have incompatible units) -->
              <template v-if="trendScales.isDual">
                <g v-for="(tick, idx) in fullscreenRightYTicks" :key="'fs-rytick-' + idx">
                  <!-- Right Y-Axis Tick Protrusion (1.5px thick, 5px outward) -->
                  <line :x1="fullscreenChartDims.width - fullscreenChartDims.padRight" :y1="tick.y" :x2="fullscreenChartDims.width - fullscreenChartDims.padRight + 5" :y2="tick.y" :stroke="trendScales.right ? trendScales.right.color : axisStroke" stroke-width="1.5" stroke-dasharray="none" />
                  <text
                    :x="fullscreenChartDims.width - fullscreenChartDims.padRight + 8"
                    :y="tick.y + 3.5"
                    text-anchor="start"
                    font-size="11"
                    font-family="monospace"
                    font-weight="600"
                    :fill="trendScales.right ? trendScales.right.color : (isDark ? '#E2E8F0' : '#0F172A')"
                  >{{ tick.label }}</text>
                </g>
              </template>

              <!-- Base X-Axis Line (1.8px high contrast) -->
              <line :x1="fullscreenChartDims.padLeft" :y1="fullscreenChartDims.baseY" :x2="fullscreenChartDims.width - fullscreenChartDims.padRight" :y2="fullscreenChartDims.baseY" stroke-dasharray="none" :stroke="axisStroke" stroke-width="1.8" />
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
                ◀ {{ trendScales.left ? trendScales.left.label : '' }}
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
                {{ trendScales.right.label }} ▶
              </text>
            </g>

            <!-- Fullscreen Communication Outage & Recovery Event Markers -->
            <g v-if="fullscreenOutageMarkers.length > 0" class="fs-outage-markers">
              <g
                v-for="marker in fullscreenOutageMarkers"
                :key="marker.id"
                class="cursor-pointer"
                @mouseenter="hoveredOutageMarker = marker"
                @mouseleave="hoveredOutageMarker = null"
              >
                <line
                  :x1="marker.x"
                  :y1="fullscreenChartDims.padTop"
                  :x2="marker.x"
                  :y2="fullscreenChartDims.baseY"
                  :stroke="marker.color"
                  stroke-width="1.8"
                  stroke-dasharray="4 3"
                  opacity="0.85"
                />
                <circle
                  :cx="marker.x"
                  :cy="fullscreenChartDims.padTop + 4"
                  r="4"
                  :fill="marker.color"
                />
                <rect
                  :x="Math.min(Math.max(marker.x - 40, fullscreenChartDims.padLeft), fullscreenChartDims.width - fullscreenChartDims.padRight - 80)"
                  :y="fullscreenChartDims.padTop"
                  width="80"
                  height="16"
                  rx="3"
                  :fill="isDark ? '#0F172A' : '#FFFFFF'"
                  :stroke="marker.color"
                  stroke-width="1"
                />
                <text
                  :x="Math.min(Math.max(marker.x, fullscreenChartDims.padLeft + 40), fullscreenChartDims.width - fullscreenChartDims.padRight - 40)"
                  :y="fullscreenChartDims.padTop + 12"
                  text-anchor="middle"
                  font-size="9"
                  font-family="monospace"
                  font-weight="bold"
                  :fill="marker.color"
                >
                  {{ marker.label }}
                </text>
              </g>
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

            <!-- Communication Gap Indicator in Fullscreen Scrubbing Tooltip -->
            <div v-if="fullscreenScrubPoint.inDataGap" class="p-1.5 rounded bg-rose-500/15 border border-rose-500/30 text-rose-600 dark:text-rose-400 text-3xs font-bold mb-1.5 flex items-center space-x-1">
              <span>⚠️</span>
              <span>COMMUNICATION GAP (NO TELEMETRY)</span>
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
      :initial-categories="categoriesList"
      @close="isCategoryModalOpen = false"
      @saved="onCategoryConfigSaved"
      @category-created="handleBentoCategoryCreated"
      @refresh-devices="onCategoryConfigSaved"
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

    <!-- 3. Parameter Add / Edit Modal (Admin / Superadmin / Engineer) -->
    <ParameterModal
      v-if="editParamTarget.device && editParamTarget.deviceId"
      ref="bentoParamModal"
      :is-open="isEditParamModalOpen"
      :device-id="editParamTarget.deviceId"
      :param-to-edit="editParamTarget.parameter"
      :device="editParamTarget.device"
      :source-device="editParamTarget.device"
      :categories="categoriesList"
      @close="isEditParamModalOpen = false"
      @saved="handleParamSaved"
      @open-category-management="openCategoryManagementModal"
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

    <!-- 7. Station Location Full Screen Map Modal -->
    <div
      v-if="showFullScreenStationMap"
      class="fixed inset-0 z-[100] bg-slate-950/80 backdrop-blur-md flex flex-col p-3 md:p-5 select-none animate-fadeIn font-sans"
      @click.self="closeFullScreenMap"
    >
      <div class="flex-1 w-full max-w-7xl mx-auto flex flex-col bg-white dark:bg-[#0B0F19] rounded-2xl border border-slate-200/90 dark:border-slate-800 shadow-2xl overflow-hidden">
        <!-- Full Screen Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-slate-200/80 dark:border-slate-800 bg-slate-50/80 dark:bg-slate-900/60">
          <div class="flex items-center space-x-2.5">
            <div class="w-9 h-9 rounded-xl flex items-center justify-center bg-[#EAF5FF] dark:bg-blue-600/15 border border-[#B8D4EA]/80 dark:border-blue-500/30 text-[#1685E5] dark:text-blue-400">
              <OperatorDashboardIcon name="map-pinned" :size="18" />
            </div>
            <div>
              <h2 class="text-sm font-bold tracking-tight text-slate-900 dark:text-white leading-none">
                Station Location — Full Screen Map
              </h2>
              <div class="text-3xs font-medium text-slate-500 dark:text-slate-400 mt-1 tabular-nums">
                {{ stationLocation.station_name || 'Ambient Monitoring Station' }} ·
                <span class="text-[#1685E5] dark:text-blue-400 font-semibold">
                  {{ candidateLocation.latitude !== null ? `${Number(candidateLocation.latitude).toFixed(6)}, ${Number(candidateLocation.longitude).toFixed(6)}` : 'No coordinates selected' }}
                </span>
              </div>
            </div>
          </div>

          <div class="flex items-center space-x-2">
            <button
              type="button"
              @click="closeFullScreenMap"
              class="px-3 py-1.5 rounded-xl text-xs font-semibold border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors flex items-center space-x-1.5"
            >
              <span>✕ Exit Full Screen (Esc)</span>
            </button>
          </div>
        </div>

        <!-- Full Screen Map Container -->
        <div class="flex-1 relative w-full h-full min-h-[400px]">
          <StationMapViewer
            ref="fullScreenMapViewer"
            :latitude="candidateLocation.latitude !== null ? Number(candidateLocation.latitude) : (stationLocation.latitude !== null ? Number(stationLocation.latitude) : null)"
            :longitude="candidateLocation.longitude !== null ? Number(candidateLocation.longitude) : (stationLocation.longitude !== null ? Number(stationLocation.longitude) : null)"
            :accuracy-meters="candidateLocation.accuracy_meters !== null ? Number(candidateLocation.accuracy_meters) : (stationLocation.accuracy_meters !== null ? Number(stationLocation.accuracy_meters) : null)"
            :editable="true"
            :source="candidateLocation.location_source || stationLocation.location_source"
            height="100%"
            @coordinate-change="onMapCoordinatePicked"
            @location-selected="onLocationSelected"
          />
        </div>

        <!-- Full Screen Footer Control Bar -->
        <div class="px-4 py-3 border-t border-slate-200/80 dark:border-slate-800 bg-slate-50/80 dark:bg-slate-900/60 flex flex-wrap items-center justify-between gap-2 text-xs">
          <div class="flex items-center space-x-4 text-slate-600 dark:text-slate-300 tabular-nums">
            <div>
              <span class="text-slate-400">Lat: </span>
              <strong class="text-slate-900 dark:text-slate-100 font-semibold">{{ candidateLocation.latitude !== null ? Number(candidateLocation.latitude).toFixed(6) : '-' }}</strong>
            </div>
            <div>
              <span class="text-slate-400">Lng: </span>
              <strong class="text-slate-900 dark:text-slate-100 font-semibold">{{ candidateLocation.longitude !== null ? Number(candidateLocation.longitude).toFixed(6) : '-' }}</strong>
            </div>
            <div v-if="candidateLocation.accuracy_meters">
              <span class="text-slate-400">Accuracy: </span>
              <strong class="text-[#1685E5] dark:text-blue-400">±{{ candidateLocation.accuracy_meters }}m</strong>
            </div>
            <div>
              <span class="text-slate-400">Source: </span>
              <span class="font-semibold text-[#1685E5] dark:text-blue-400">{{ formatLocationSourceLabel(candidateLocation.location_source || 'MANUAL') }}</span>
            </div>
          </div>

          <div class="flex items-center space-x-2">
            <button
              type="button"
              @click="cancelLocationEdit"
              class="px-3.5 py-1.5 rounded-xl text-xs font-semibold border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 transition-colors"
            >
              Revert
            </button>
            <button
              type="button"
              @click="saveStationLocation"
              :disabled="isSavingStationLocation || candidateLocation.latitude == null || candidateLocation.longitude == null"
              class="px-4 py-1.5 rounded-xl text-xs font-semibold bg-[#1685E5] hover:bg-[#1273C9] text-white btn-text transition-colors disabled:opacity-50 disabled:cursor-not-allowed shadow-sm flex items-center space-x-1.5"
            >
              <span v-if="isSavingStationLocation" class="animate-spin inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full"></span>
              <span>Save Station Coordinates</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import OperatorCommandCenterView from './OperatorCommandCenterView.vue';
import BentoISPUHeroCard from '../components/BentoISPUHeroCard.vue';
import BentoWQMSIndexBar from '../components/BentoWQMSIndexBar.vue';
import audioAlert from '../utils/audioAlert';
import {
  getStoredAlarmPreferences,
  saveAlarmPreferences,
  applyAlarmPreferences,
} from '../utils/alarmPreferences';
import {
  saveCustomerBranding,
  saveMakerBranding,
  extractConcisePlaceName,
  composeAutoGpsStationType,
  reverseGeocodePlaceName,
  inferStationDomainPrefix,
} from '../utils/companyBranding';

const REGULATORY_REF_LIMITS = {
  PM10: 150,
  PM25: 55,
  'PM2.5': 55,
  SO2: 180,
  CO: 4000,
  O3: 120,
  NO2: 200,
  HC: 160,
  NMHC: 160,
};

export default {
  name: 'DashboardOperatorView',
  extends: OperatorCommandCenterView,
  components: {
    ...OperatorCommandCenterView.components,
    BentoISPUHeroCard,
    BentoWQMSIndexBar,
  },
  data() {
    let autoGrid = true;
    let cardGroupByCat = false;
    let splitMode = '2T';
    let savedTrendHeight = 205;
    let savedCustName = '';
    let savedCustSub = '';
    let savedStationTypeMode = 'MANUAL';
    let savedDomainPrefix = 'AQMS';
    let savedCachedPlaceName = '';
    let savedMakerName = '';
    let savedMakerSub = '';
    let savedGpsModbus = {
      deviceId: '',
      latParamId: '',
      lngParamId: '',
      satParamId: '',
      autoSync: false,
    };
    try {
      const savedAuto = localStorage.getItem('datalogger_bento_auto_grid');
      if (savedAuto === 'false') autoGrid = false;
      const savedCardGroup = localStorage.getItem('datalogger_bento_card_group_cat');
      if (savedCardGroup === 'true') cardGroupByCat = true;
      const savedSplit = localStorage.getItem('datalogger_bento_table_split');
      if (savedSplit === '1T' || savedSplit === '2T' || savedSplit === 'CAT') splitMode = savedSplit;
      const rawTrendH = parseInt(localStorage.getItem('datalogger_bento_trend_height'), 10);
      if (isFinite(rawTrendH) && rawTrendH >= 120 && rawTrendH <= 560) {
        savedTrendHeight = rawTrendH;
      }
      const rawBrand = localStorage.getItem('datalogger.organization_branding');
      if (rawBrand) {
        const parsedBrand = JSON.parse(rawBrand);
        if (parsedBrand && parsedBrand.customer) {
          savedCustName = parsedBrand.customer.name || '';
          savedCustSub = parsedBrand.customer.subtitle === 'Local Autonomous Edge' ? '' : (parsedBrand.customer.subtitle || '');
          savedStationTypeMode = parsedBrand.customer.stationTypeMode === 'AUTO_GPS' ? 'AUTO_GPS' : 'MANUAL';
          savedDomainPrefix = parsedBrand.customer.domainPrefix || inferStationDomainPrefix(savedCustSub, 'AQMS');
        }
        if (parsedBrand && parsedBrand.maker) {
          savedMakerName = parsedBrand.maker.name || '';
          savedMakerSub = parsedBrand.maker.subtitle || '';
        }
      }
      savedCachedPlaceName = extractConcisePlaceName(localStorage.getItem('datalogger_cached_gps_place_name') || '');
      const rawGpsModbus = localStorage.getItem('datalogger_gps_modbus_rtu_mapping');
      if (rawGpsModbus) {
        const parsedGpsModbus = JSON.parse(rawGpsModbus);
        if (parsedGpsModbus && typeof parsedGpsModbus === 'object') {
          savedGpsModbus = {
            deviceId: parsedGpsModbus.deviceId ? String(parsedGpsModbus.deviceId) : '',
            latParamId: parsedGpsModbus.latParamId ? String(parsedGpsModbus.latParamId) : '',
            lngParamId: parsedGpsModbus.lngParamId ? String(parsedGpsModbus.lngParamId) : '',
            satParamId: parsedGpsModbus.satParamId ? String(parsedGpsModbus.satParamId) : '',
            autoSync: Boolean(parsedGpsModbus.autoSync),
          };
        }
      }
    } catch (e) {}
    return {
      bentoViewportWidth: typeof window !== 'undefined' ? window.innerWidth : 1600,
      bentoViewportHeight: typeof window !== 'undefined' ? window.innerHeight : 900,
      bentoPopoverAnchorBottomPx: 140,
      bentoParamsWrapWidth: 0,
      bentoAutoGrid: autoGrid,
      bentoCardGroupByCategory: cardGroupByCat,
      bentoCollapsedCatCards: {},
      bentoTableSplitMode: splitMode,
      bentoActiveMiniPopover: null,
      bentoTrendSearch: '',
      bentoTrendCanvasHeight: savedTrendHeight,
      isResizingBentoTrend: false,
      customerCompanyName: savedCustName,
      customerSubtitle: savedCustSub,
      stationTypeMode: savedStationTypeMode,
      stationDomainPrefix: savedDomainPrefix,
      resolvedGpsPlaceName: savedCachedPlaceName,
      modalStationTypeMode: savedStationTypeMode,
      modalStationDomainPrefix: savedDomainPrefix,
      modalResolvedPlaceName: savedCachedPlaceName,
      isResolvingAutoPlace: false,
      lastResolvedGpsKey: '',
      makerCompanyName: savedMakerName,
      makerSubtitle: savedMakerSub,
      bentoAudioEngineMuted: audioAlert.checkMuteActive(),
      bentoAudioEngineSoundEnabled: audioAlert.soundEnabled !== false,
      gpsModbusMapping: savedGpsModbus,
      showAdvancedGpsRegisterMap: false,
      lastModbusAutoSyncMs: 0,
    };
  },
  computed: {
    canManageCategories() {
      const user = this.$store && this.$store.state ? this.$store.state.user : null;
      if (!user) return false;
      const rawRole = typeof user.role === 'string'
        ? user.role
        : ((user.role && user.role.name) || user.role_name || (user.Role && user.Role.name) || '');
      const normalized = String(rawRole).trim().toUpperCase();
      return Boolean(normalized && normalized !== 'OPERATOR');
    },
    activeMonitoringDomain() {
      const dom =
        (this.systemStatus && this.systemStatus.monitoring_domain) ||
        (this.$store && this.$store.state && this.$store.state.systemStatus && this.$store.state.systemStatus.monitoring_domain) ||
        'AQMS';
      return String(dom).trim().toUpperCase();
    },
    isAQMSDomain() {
      if (this.activeMonitoringDomain !== 'AQMS') {
        return false;
      }
      const checkStr = this.stationTypeMode === 'AUTO_GPS'
        ? String(this.stationDomainPrefix || 'AQMS').trim().toUpperCase()
        : String(this.customerSubtitle || '').trim().toUpperCase();
      if ((checkStr.includes('WQMS') || checkStr.includes('SPARING')) && !checkStr.includes('AQMS')) {
        return false;
      }
      return true;
    },
    activeAutoGpsCoords() {
      if (this.modbusRtuGpsStatus && this.modbusRtuGpsStatus.state === 'LOCKED') {
        return {
          lat: Number(this.modbusRtuGpsStatus.lat),
          lng: Number(this.modbusRtuGpsStatus.lng),
          source: 'MODBUS_RTU',
        };
      }
      if (
        this.stationLocation &&
        this.stationLocation.is_configured &&
        this.stationLocation.latitude !== null &&
        this.stationLocation.latitude !== undefined &&
        this.stationLocation.longitude !== null &&
        this.stationLocation.longitude !== undefined
      ) {
        const lat = Number(this.stationLocation.latitude);
        const lng = Number(this.stationLocation.longitude);
        if (Number.isFinite(lat) && Number.isFinite(lng) && !(Math.abs(lat) < 0.0001 && Math.abs(lng) < 0.0001)) {
          return {
            lat,
            lng,
            source: this.stationLocation.location_source || 'MANUAL',
          };
        }
      }
      return null;
    },
    activeAutoGpsSourceBadge() {
      if (!this.activeAutoGpsCoords) {
        return this.fallbackSavedPlaceName ? 'Alamat Stasiun Tersimpan' : 'Belum Dikonfigurasi';
      }
      if (this.activeAutoGpsCoords.source === 'MODBUS_RTU') {
        return 'GPS Modbus RTU';
      }
      return this.formatLocationSourceLabel(this.activeAutoGpsCoords.source);
    },
    fallbackSavedPlaceName() {
      const addr = this.stationLocation && this.stationLocation.address ? this.stationLocation.address : '';
      const stName = this.stationLocation && this.stationLocation.station_name ? this.stationLocation.station_name : '';
      return extractConcisePlaceName(addr, stName);
    },
    effectiveAutoPlaceName() {
      return extractConcisePlaceName(this.resolvedGpsPlaceName) || this.fallbackSavedPlaceName || '';
    },
    effectiveModalPlaceName() {
      return extractConcisePlaceName(this.modalResolvedPlaceName) || this.effectiveAutoPlaceName || '';
    },
    modalAutoStationTypePreview() {
      return composeAutoGpsStationType(this.modalStationDomainPrefix || 'AQMS', this.effectiveModalPlaceName);
    },
    stationTypeLabel() {
      if (this.stationTypeMode === 'AUTO_GPS') {
        return composeAutoGpsStationType(this.stationDomainPrefix || 'AQMS', this.effectiveAutoPlaceName);
      }
      if (this.customerSubtitle !== undefined && this.customerSubtitle !== null) {
        const trimmed = String(this.customerSubtitle).trim();
        if (trimmed !== '' && trimmed !== 'Local Autonomous Edge') {
          return trimmed;
        }
      }
      return '';
    },
    headerSecondaryLocationLabel() {
      if (!this.stationLocation) return '';
      if (this.stationTypeMode === 'AUTO_GPS') {
        const stName = String(this.stationLocation.station_name || '').trim();
        const cleanStName = extractConcisePlaceName(stName);
        if (!cleanStName) return '';
        const mainLower = String(this.stationTypeLabel || '').toLowerCase();
        if (mainLower.includes(cleanStName.toLowerCase())) return '';
        return cleanStName;
      }
      return this.stationLocation.station_name || this.stationLocation.address || '';
    },
    isWQMSDomain() {
      return this.activeMonitoringDomain === 'WQMS';
    },
    isCompactScreen() {
      return this.bentoViewportWidth < 1280;
    },
    bentoToolbarPopoverMaxHeight() {
      const vh = this.bentoViewportHeight || 900;
      const anchorBottom = this.bentoPopoverAnchorBottomPx || 140;
      const avail = Math.max(320, vh - anchorBottom - 20);
      return `${avail}px`;
    },
    modbusRtuDevicesList() {
      const devs = (this.devicesList || []).filter((d) => {
        if (d.enabled === false || d.enabled === 0) return false;
        const st = String(d.status || d.administrative_status || '').toUpperCase();
        return st !== 'DISABLED' && st !== 'INACTIVE';
      });
      const isRtu = (d) => {
        const proto = String(d.protocol || d.driver_type || d.connection_type || '').toUpperCase();
        const code = String(d.device_code || d.code || '').toUpperCase();
        return proto.includes('RTU') || proto.includes('RS485') || proto.includes('SERIAL') || code.startsWith('COM');
      };
      return devs.slice().sort((a, b) => {
        const aR = isRtu(a) ? 1 : 0;
        const bR = isRtu(b) ? 1 : 0;
        if (aR !== bR) return bR - aR;
        return String(a.device_name || a.name || a.device_code || '').localeCompare(String(b.device_name || b.name || b.device_code || ''));
      });
    },
    modbusRtuCandidateParams() {
      const mapping = this.gpsModbusMapping || {};
      let list = (this.allParameters || []).slice();
      if (mapping.deviceId) {
        list = list.filter(p => String(p.device_id) === String(mapping.deviceId));
      }
      const devProtoMap = {};
      for (const d of (this.devicesList || [])) {
        const proto = String(d.protocol || d.driver_type || d.connection_type || '').toUpperCase();
        const code = String(d.device_code || d.code || '').toUpperCase();
        devProtoMap[d.id] = proto.includes('RTU') || proto.includes('RS485') || proto.includes('SERIAL') || code.startsWith('COM');
      }
      const isGpsHint = (p) => {
        const text = `${p.param_code || ''} ${p.param_name || ''}`;
        return /(gps|lat|lintang|lng|lon|bujur|coord|sat)/i.test(text);
      };
      return list.sort((a, b) => {
        const aHint = isGpsHint(a) ? 1 : 0;
        const bHint = isGpsHint(b) ? 1 : 0;
        if (aHint !== bHint) return bHint - aHint;
        const aRtu = devProtoMap[a.device_id] ? 1 : 0;
        const bRtu = devProtoMap[b.device_id] ? 1 : 0;
        if (aRtu !== bRtu) return bRtu - aRtu;
        return String(a.param_name || '').localeCompare(String(b.param_name || ''));
      });
    },
    effectiveModbusGpsMapping() {
      const allParams = this.allParameters || [];
      const mapping = this.gpsModbusMapping || {};
      const scopedParams = mapping.deviceId
        ? allParams.filter(p => String(p.device_id) === String(mapping.deviceId))
        : allParams;

      const explicitLat = mapping.latParamId
        ? allParams.find(p => String(p.param_id) === String(mapping.latParamId)) || null
        : null;
      const explicitLng = mapping.lngParamId
        ? allParams.find(p => String(p.param_id) === String(mapping.lngParamId)) || null
        : null;
      const explicitSat = mapping.satParamId
        ? allParams.find(p => String(p.param_id) === String(mapping.satParamId)) || null
        : null;

      const latRegex = /(^|[_\-\s])(gps[_\-\s]?lat|latitude|lat|lintang)($|[_\-\s])/i;
      const lngRegex = /(^|[_\-\s])(gps[_\-\s]?(lng|lon|long)|longitude|lng|lon|bujur)($|[_\-\s])/i;
      const satRegex = /(^|[_\-\s])(gps[_\-\s]?(sat|fix)|satellites?|sats|gps_status|gnss_fix)($|[_\-\s])/i;

      const autoLat = explicitLat || scopedParams.find(p => latRegex.test(`${p.param_code || ''} ${p.param_name || ''}`)) || null;
      // Prefer pairing Longitude from the exact same GPS device as Latitude first
      const sameDevicePool = autoLat
        ? scopedParams.filter(p => p !== autoLat && String(p.device_id) === String(autoLat.device_id))
        : [];
      const autoLng = explicitLng
        || sameDevicePool.find(p => lngRegex.test(`${p.param_code || ''} ${p.param_name || ''}`))
        || scopedParams.find(p => p !== autoLat && lngRegex.test(`${p.param_code || ''} ${p.param_name || ''}`))
        || null;
      const autoSat = explicitSat
        || sameDevicePool.find(p => p !== autoLng && satRegex.test(`${p.param_code || ''} ${p.param_name || ''}`))
        || scopedParams.find(p => p !== autoLat && p !== autoLng && satRegex.test(`${p.param_code || ''} ${p.param_name || ''}`))
        || null;

      const isAutoDetected = Boolean(!explicitLat && !explicitLng && autoLat && autoLng);
      return {
        latParam: autoLat,
        lngParam: autoLng,
        satParam: autoSat,
        isAutoDetected,
      };
    },
    modbusRtuGpsStatus() {
      const { latParam, lngParam, satParam } = this.effectiveModbusGpsMapping;
      if (!latParam || !lngParam) {
        return {
          state: 'UNMAPPED',
          badgeLabel: 'NOT MAPPED',
          badgeClass: 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-700',
          dotClass: 'bg-slate-400',
          lat: null,
          lng: null,
          satellites: null,
          reason: 'Pilih parameter Latitude & Longitude dari data Modbus RTU.',
        };
      }

      const rawLat = latParam.value !== null && latParam.value !== undefined && latParam.value !== '' ? Number(latParam.value) : NaN;
      const rawLng = lngParam.value !== null && lngParam.value !== undefined && lngParam.value !== '' ? Number(lngParam.value) : NaN;
      const rawLatFinite = Number.isFinite(rawLat);
      const rawLngFinite = Number.isFinite(rawLng);
      const satVal = satParam && satParam.value !== null && satParam.value !== undefined && satParam.value !== '' && Number.isFinite(Number(satParam.value))
        ? Math.round(Number(satParam.value))
        : null;

      const isOfflineOrStale =
        this.isBentoParamCommOffline(latParam) ||
        this.isBentoParamCommOffline(lngParam) ||
        latParam.quality === 'BAD' ||
        lngParam.quality === 'BAD' ||
        latParam.quality === 'STALE' ||
        lngParam.quality === 'STALE';

      if (isOfflineOrStale) {
        return {
          state: 'OFFLINE',
          badgeLabel: 'COMM LOST / STALE',
          badgeClass: 'bg-rose-50 dark:bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-200 dark:border-rose-500/30',
          dotClass: 'bg-rose-500 animate-pulse',
          lat: rawLatFinite ? rawLat : null,
          lng: rawLngFinite ? rawLng : null,
          satellites: satVal,
          reason: 'Komunikasi Modbus RTU terputus atau stale. Koordinat tersimpan tetap aman.',
        };
      }

      const isNullIsland = rawLatFinite && rawLngFinite && Math.abs(rawLat) < 0.0001 && Math.abs(rawLng) < 0.0001;
      const isOutOfBounds = !rawLatFinite || !rawLngFinite || rawLat < -90 || rawLat > 90 || rawLng < -180 || rawLng > 180;
      const hasZeroSatFix = satVal !== null && satVal <= 0;

      if (isOutOfBounds || isNullIsland || hasZeroSatFix) {
        return {
          state: 'NO_FIX',
          badgeLabel: 'SEARCHING / NO FIX',
          badgeClass: 'bg-amber-50 dark:bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-200 dark:border-amber-500/30',
          dotClass: 'bg-amber-500 animate-pulse',
          lat: rawLatFinite ? rawLat : null,
          lng: rawLngFinite ? rawLng : null,
          satellites: satVal,
          reason: isNullIsland
            ? 'Modul GPS belum mengunci satelit (0, 0 ditolak agar lokasi stasiun tidak tertimpa).'
            : 'Menunggu sinyal satelit GPS yang valid (-90..90, -180..180).',
        };
      }

      return {
        state: 'LOCKED',
        badgeLabel: 'GPS LOCKED',
        badgeClass: 'bg-emerald-50 dark:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-200 dark:border-emerald-500/30',
        dotClass: 'bg-emerald-500',
        lat: Number(rawLat.toFixed(6)),
        lng: Number(rawLng.toFixed(6)),
        satellites: satVal,
        reason: 'Sinyal GPS Modbus RTU terkunci dan siap disinkronkan.',
      };
    },
    stationLocationStatusDotClass() {
      if (this.stationLocation && this.stationLocation.location_source === 'MODBUS_RTU') {
        if (this.modbusRtuGpsStatus.state === 'LOCKED') return 'dot-emerald';
        if (this.modbusRtuGpsStatus.state === 'OFFLINE') return 'dot-rose';
        return 'dot-amber';
      }
      if (this.stationLocation && this.stationLocation.is_configured) {
        return 'dot-emerald';
      }
      return 'dot-amber';
    },
    stationLocationTooltip() {
      if (this.stationLocation && this.stationLocation.is_configured && this.stationLocation.latitude !== null && this.stationLocation.longitude !== null) {
        const lat = Number(this.stationLocation.latitude).toFixed(4);
        const lng = Number(this.stationLocation.longitude).toFixed(4);
        const src = this.formatLocationSourceLabel(this.stationLocation.location_source);
        return `Station Location: [${lat}, ${lng}] (${src})`;
      }
      return 'Station Location: Unconfigured - Click to set coordinates';
    },
    activeBentoCategoryName() {
      if (this.filterPinnedOnly) return 'Favorit Dipin';
      if (this.selectedCategoryFilter === 'all' || this.selectedCategoryFilter == null) {
        return 'Semua Parameter';
      }
      const found = (this.activeCategoriesList || []).find(
        c => Number(c.id) === Number(this.selectedCategoryFilter)
      );
      return found ? found.name : 'Kategori';
    },
    filteredTrendPopoverParams() {
      const list = this.allParameters || [];
      const q = (this.bentoTrendSearch || '').toLowerCase();
      if (!q) return list;
      return list.filter(p => {
        const name = (p.param_name || '').toLowerCase();
        const code = (p.param_code || '').toLowerCase();
        const unit = (p.unit || '').toLowerCase();
        return name.includes(q) || code.includes(q) || unit.includes(q);
      });
    },
    previewTrendDots() {
      return (this.activeTrendSeries || [])
        .slice(0, 4)
        .map(s => s.color || '#0284C7');
    },
    effectiveBentoCols() {
      const w = this.bentoViewportWidth || 1600;
      const n = (this.filteredParameters && this.filteredParameters.length) || 0;
      if (!this.bentoAutoGrid) {
        const manual = Math.max(1, Math.min(5, Number(this.gridColumns) || 4));
        if (w < 600) return Math.min(manual, 2);
        if (w < 900) return Math.min(manual, 3);
        return manual;
      }
      if (w < 600) return 2;
      if (w < 900) return n >= 9 ? 4 : 3;
      if (w < 1280) {
        if (n >= 11) return 5;
        if (n >= 7) return 4;
        return 3;
      }
      // Wide screen (>= 1280px, alongside Left ISPU Hero Card)
      if (n >= 11) return 5;
      if (n >= 7) return 4;
      if (n >= 5) return 3;
      return n >= 2 ? 2 : 1;
    },
    isMicroCardDensity() {
      const n = (this.filteredParameters && this.filteredParameters.length) || 0;
      return this.isCompactScreen || n >= 11;
    },
    bentoTableModeShortBadge() {
      if (this.bentoTableSplitMode === 'CAT') return 'KAT';
      return this.bentoTableSplitMode || '2T';
    },
    bentoAvailableParamsWidth() {
      if (this.bentoParamsWrapWidth && this.bentoParamsWrapWidth > 100) {
        return this.bentoParamsWrapWidth;
      }
      const vw = this.bentoViewportWidth || 1600;
      const hasLeftIspu = Boolean(this.isAQMSDomain && this.isIspuBarVisible && !this.isCompactScreen);
      return Math.max(320, vw - (hasLeftIspu ? 284 : 28));
    },
    bentoCategoryTableSections() {
      const list = this.filteredParameters || [];
      if (!list.length) return [];

      const colorHexMap = {
        cyan: '#06B6D4',
        sky: '#0EA5E9',
        blue: '#3B82F6',
        indigo: '#6366F1',
        purple: '#8B5CF6',
        violet: '#8B5CF6',
        emerald: '#10B981',
        green: '#10B981',
        teal: '#14B8A6',
        amber: '#F59E0B',
        yellow: '#EAB308',
        orange: '#F97316',
        rose: '#F43F5E',
        red: '#EF4444',
        pink: '#EC4899',
        slate: '#64748B',
      };
      const resolveAccent = (cat) => {
        if (!cat) return '#64748B';
        if (cat.color && String(cat.color).startsWith('#')) return cat.color;
        const ck = String(cat.color_key || '').toLowerCase();
        if (colorHexMap[ck]) return colorHexMap[ck];
        const nm = String(cat.name || '').toLowerCase();
        if (/gas|pollutant/.test(nm)) return '#06B6D4';
        if (/particulate|pm/.test(nm)) return '#3B82F6';
        if (/weather|meteorolog/.test(nm)) return '#0EA5E9';
        if (/water|chemical|flow/.test(nm)) return '#14B8A6';
        if (/status/.test(nm)) return '#10B981';
        return '#64748B';
      };

      const buildColMeta = (key, title, accentColor, params, isUncategorized = false) => {
        let faultCount = 0;
        let warnCount = 0;
        for (const p of params) {
          const tone = this.getBentoStatusBadge(p).toneClass;
          if (tone === 'badge-rose') faultCount += 1;
          else if (tone === 'badge-amber') warnCount += 1;
        }
        return {
          key,
          title,
          accentColor,
          isCategoryTable: true,
          isUncategorized,
          faultCount,
          warnCount,
          params,
        };
      };

      const resolveParamCatId = (p) => {
        if (!p) return null;
        if (p.category_id != null && p.category_id !== '') return Number(p.category_id);
        if (p.category && p.category.id != null) return Number(p.category.id);
        const pid = p.param_id || p.id;
        const raw = (pid != null && this.bentoRawParamMap && this.bentoRawParamMap[pid]) || null;
        if (raw) {
          if (raw.category_id != null && raw.category_id !== '') return Number(raw.category_id);
          if (raw.category && raw.category.id != null) return Number(raw.category.id);
        }
        return null;
      };

      const sections = [];
      const usedIds = new Set();

      // 1. Primary domain categories ordered by display_order
      const orderedCats = [];
      const seenCatIds = new Set();
      for (const c of (this.activeCategoriesList || [])) {
        if (c && c.id != null && !seenCatIds.has(Number(c.id))) {
          seenCatIds.add(Number(c.id));
          orderedCats.push(c);
        }
      }
      // Also include any enabled cross-domain categories from categoriesList in display_order
      const allSortedCats = Array.isArray(this.categoriesList)
        ? [...this.categoriesList]
            .filter(c => c && c.id != null && c.enabled !== false)
            .sort((a, b) => (a.display_order || 0) - (b.display_order || 0))
        : [];
      for (const c of allSortedCats) {
        if (!seenCatIds.has(Number(c.id))) {
          seenCatIds.add(Number(c.id));
          orderedCats.push(c);
        }
      }

      for (const cat of orderedCats) {
        const catId = Number(cat.id);
        const catParams = list.filter(p => {
          const pid = p.param_id || p.id;
          if (usedIds.has(pid)) return false;
          return resolveParamCatId(p) === catId;
        });
        if (catParams.length > 0) {
          catParams.forEach(p => usedIds.add(p.param_id || p.id));
          sections.push(
            buildColMeta(
              'cat-' + catId,
              String(cat.name || cat.code || 'KATEGORI').toUpperCase(),
              resolveAccent(cat),
              catParams,
              false
            )
          );
        }
      }

      // 2. Fallback for embedded parameter category objects not present in categoriesList
      const extraGroups = new Map();
      for (const p of list) {
        const pid = p.param_id || p.id;
        if (usedIds.has(pid)) continue;
        const raw = (pid != null && this.bentoRawParamMap && this.bentoRawParamMap[pid]) || null;
        const cat = this.getParamCategory(p) || (raw ? this.getParamCategory(raw) : null);
        if (cat && (cat.id != null || cat.name)) {
          const gKey = cat.id != null ? ('id-' + cat.id) : ('nm-' + String(cat.name).toLowerCase());
          if (!extraGroups.has(gKey)) {
            extraGroups.set(gKey, { cat, params: [] });
          }
          extraGroups.get(gKey).params.push(p);
          usedIds.add(pid);
        }
      }
      for (const [gKey, grp] of extraGroups.entries()) {
        sections.push(
          buildColMeta(
            'extra-' + gKey,
            String(grp.cat.name || grp.cat.code || 'KATEGORI').toUpperCase(),
            resolveAccent(grp.cat),
            grp.params,
            false
          )
        );
      }

      // 3. Uncategorized parameters (only spawned if there are uncategorized parameters)
      const uncatParams = list.filter(p => !usedIds.has(p.param_id || p.id));
      if (uncatParams.length > 0) {
        sections.push(
          buildColMeta(
            'cat-uncategorized',
            sections.length > 0 ? 'TANPA KATEGORI' : 'PARAMETER TANPA KATEGORI',
            '#64748B',
            uncatParams,
            true
          )
        );
      }

      return sections;
    },
    bentoSplitTableColumns() {
      const list = this.filteredParameters || [];
      if (this.bentoTableSplitMode === 'CAT') {
        return this.bentoCategoryTableSections;
      }
      if (this.bentoTableSplitMode === '1T' || list.length <= 2) {
        return [
          {
            key: 'tbl-1t',
            title: 'DAFTAR PARAMETER TELEMETRI (' + list.length + ' KANAL)',
            params: list,
          },
        ];
      }
      const mid = Math.ceil(list.length / 2);
      return [
        {
          key: 'tbl-2t-1',
          title: 'KOLOM 1 · PARAMETER (#1–' + mid + ')',
          params: list.slice(0, mid),
        },
        {
          key: 'tbl-2t-2',
          title: 'KOLOM 2 · PARAMETER (#' + (mid + 1) + '–' + list.length + ')',
          params: list.slice(mid),
        },
      ];
    },
    bentoEffectiveTableGridCols() {
      const count = (this.bentoSplitTableColumns && this.bentoSplitTableColumns.length) || 1;
      const availW = this.bentoAvailableParamsWidth || 1200;
      // Require >= 796px container width for 2 side-by-side tables (~394px+ per table)
      if (count <= 1 || availW < 796) return 1;
      if (this.bentoTableSplitMode === '2T' || count === 2) return 2;
      // Require >= 1216px container width for 3 side-by-side category tables (~400px+ per table)
      if (this.bentoTableSplitMode === 'CAT' && count >= 3) {
        if (availW >= 1216) return 3;
        return 2;
      }
      return Math.min(2, count);
    },
    bentoPerTableCardWidth() {
      const cols = this.bentoEffectiveTableGridCols || 1;
      const availW = this.bentoAvailableParamsWidth || 1200;
      const gapTotal = (cols - 1) * 8;
      return Math.floor((availW - gapTotal) / cols);
    },
    isBentoDenseTableCard() {
      return this.bentoEffectiveTableGridCols >= 3 || this.bentoPerTableCardWidth < 500;
    },
    showBentoTableSparklineCol() {
      return this.bentoPerTableCardWidth >= 445;
    },
    showBentoTableUpdateCol() {
      if (this.bentoEffectiveTableGridCols >= 3) return false;
      return this.bentoPerTableCardWidth >= 560;
    },
    bentoParamsScrollAreaStyle() {
      if (!this.perfSettings || !this.perfSettings.trendAnalysisEnabled) {
        return null;
      }
      const isCatCards = this.viewMode === 'cards' && Boolean(this.bentoCardGroupByCategory);
      const isTableOrGrouped = this.viewMode === 'table' || isCatCards;
      const minH = isCatCards
        ? (this.isCompactScreen ? 340 : 385)
        : (isTableOrGrouped ? (this.isCompactScreen ? 325 : 365) : (this.isCompactScreen ? 268 : 332));
      const topIspuOffset = (this.isCompactScreen && this.isAQMSDomain && this.isIspuBarVisible) ? 62 : 0;
      const topWqmsOffset = (this.isWQMSDomain && this.isIspuBarVisible) ? 62 : 0;
      const chromeOffset = isCatCards ? 212 : 230;
      const reservedPx = Math.round((Number(this.bentoTrendCanvasHeight) || 205) + chromeOffset + topIspuOffset + topWqmsOffset);
      return {
        maxHeight: 'max(' + minH + 'px, calc(100vh - ' + reservedPx + 'px))',
      };
    },
    bentoCategorySlotCols() {
      const availW = this.bentoAvailableParamsWidth || this.bentoViewportWidth || 1280;
      if (!this.bentoAutoGrid) {
        const manual = Math.max(1, Math.min(8, Number(this.gridColumns) || 5));
        if (availW < 560) return Math.min(manual, 2);
        if (availW < 820) return Math.min(manual, 4);
        return manual;
      }
      if (availW >= 1460) return 9;
      if (availW >= 1050) return 8;
      if (availW >= 910) return 7;
      if (availW >= 770) return 6;
      if (availW >= 620) return 5;
      if (availW >= 460) return 4;
      return 2;
    },
    bentoCardModeShortBadge() {
      if (this.bentoCardGroupByCategory) {
        return 'K·' + this.bentoCategorySlotCols + 'P';
      }
      return this.effectiveBentoCols + 'P';
    },
    bentoActiveCardSections() {
      const list = this.filteredParameters || [];
      if (!list.length) return [];
      if (this.bentoCardGroupByCategory) {
        const totalCols = Math.max(2, this.bentoCategorySlotCols || 8);
        const rawItems = (this.bentoCategoryTableSections || []).map(sec => {
          const collapsed = this.isBentoCatCardCollapsed(sec.key);
          const count = Math.max(1, (sec.params && sec.params.length) || 1);
          const neededCols = collapsed ? Math.min(totalCols, 2) : Math.min(totalCols, count);
          const innerCols = collapsed ? 1 : Math.min(totalCols, count);
          return {
            ...sec,
            isCategoryGroup: true,
            collapsed,
            count,
            neededCols,
            colSpan: neededCols,
            innerCols,
          };
        });

        // First-fit proportional Bento row packing so small categories (e.g. STATUS: 1)
        // share a row side-by-side with larger categories (e.g. GAS: 7 -> 1 + 7 = 8 slots)
        const rows = [];
        for (const item of rawItems) {
          let placed = false;
          for (const r of rows) {
            if (r.used + item.neededCols <= totalCols) {
              r.items.push(item);
              r.used += item.neededCols;
              placed = true;
              break;
            }
          }
          if (!placed) {
            rows.push({ items: [item], used: item.neededCols });
          }
        }

        for (const r of rows) {
          const rem = totalCols - r.used;
          if (rem <= 0) continue;
          if (r.items.length === 1) {
            const single = r.items[0];
            if (!single.collapsed && single.count >= Math.ceil(totalCols * 0.65)) {
              // Align wide single-category row (e.g. 7 cards on 8-slot grid) flush right
              single.colSpan = totalCols;
            } else if (!single.collapsed && single.count === 1 && totalCols >= 4) {
              // Give a standalone 1-card row enough header breathing room without stretching full width
              single.colSpan = Math.min(totalCols, 2);
            }
          } else if (r.used >= Math.ceil(totalCols * 0.72)) {
            let widest = null;
            for (const it of r.items) {
              if (!it.collapsed && (!widest || it.colSpan >= widest.colSpan)) {
                widest = it;
              }
            }
            if (widest) {
              widest.colSpan += rem;
            }
          }
        }

        const packed = [];
        for (const r of rows) {
          for (const it of r.items) {
            packed.push(it);
          }
        }
        return packed;
      }
      return [
        {
          key: 'all-flat-cards',
          title: '',
          accentColor: '',
          isCategoryGroup: false,
          faultCount: 0,
          warnCount: 0,
          params: list,
        },
      ];
    },
    bentoRawParamMap() {
      const map = {};
      for (const dev of (this.devicesList || [])) {
        if (!dev || !Array.isArray(dev.parameters)) continue;
        for (const rp of dev.parameters) {
          const pid = rp.id || rp.param_id;
          if (pid != null) {
            map[pid] = rp;
          }
        }
      }
      return map;
    },
    bentoParamActiveAlarmMap() {
      const map = {};
      const byDevCode = {};
      const list = this.activeAlarmsList || [];
      if (!list.length) return map;

      const getPriorityScore = (a) => {
        if (!a) return 0;
        let score = String(a.severity || '').toUpperCase() === 'CRITICAL' ? 1000 : 500;
        const code = String(a.alarm_code || a.alarm_type || a.type || '').toUpperCase();
        if (code.includes('CRITICAL')) score += 90;
        else if (code.includes('THRESHOLD')) score += 80;
        else if (code.includes('BAD_QUALITY')) score += 70;
        else if (code.includes('STALE')) score += 60;
        else if (code.includes('STUCK')) score += 50;
        else if (code.includes('SPIKE')) score += 45;
        else if (code.includes('ZERO')) score += 40;
        return score;
      };

      const addEntry = (targetMap, key, alm) => {
        if (key == null || key === '') return;
        if (!targetMap[key]) {
          targetMap[key] = { primary: alm, all: [alm], count: 1 };
          return;
        }
        targetMap[key].all.push(alm);
        targetMap[key].count = targetMap[key].all.length;
        const curScore = getPriorityScore(targetMap[key].primary);
        const newScore = getPriorityScore(alm);
        if (newScore > curScore) {
          targetMap[key].primary = alm;
        } else if (newScore === curScore) {
          const curT = new Date(targetMap[key].primary.triggered_at || targetMap[key].primary.created_at || 0).getTime();
          const newT = new Date(alm.triggered_at || alm.created_at || 0).getTime();
          if (newT > curT) {
            targetMap[key].primary = alm;
          }
        }
      };

      for (const a of list) {
        if (!a || (a.status && a.status !== 'ACTIVE')) continue;
        const embP = a.parameter || a.Parameter || null;
        const pid = a.parameter_id != null
          ? Number(a.parameter_id)
          : (embP && (embP.id || embP.param_id) != null ? Number(embP.id || embP.param_id) : null);
        if (pid != null && !isNaN(pid)) {
          addEntry(map, pid, a);
        } else if (embP && (embP.parameter_code || embP.code)) {
          const devId = a.device_id != null
            ? Number(a.device_id)
            : (a.device && a.device.id != null ? Number(a.device.id) : 0);
          const pCode = String(embP.parameter_code || embP.code).toUpperCase();
          addEntry(byDevCode, devId + ':' + pCode, a);
        }
      }

      if (Object.keys(byDevCode).length > 0) {
        for (const p of (this.allParameters || [])) {
          const pid = Number(p.param_id || p.id);
          if (!pid || map[pid]) continue;
          const fbKey = Number(p.device_id || 0) + ':' + String(p.param_code || '').toUpperCase();
          if (byDevCode[fbKey]) {
            map[pid] = byDevCode[fbKey];
          }
        }
      }

      return map;
    },
    bentoFleetSummary() {
      const total = this.allParameters.length;
      if (total === 0) {
        return {
          label: 'Menunggu Kanal Sensor...',
          shortLabel: 'Menunggu...',
          toneClass: 'pill-amber',
          dotClass: 'dot-amber',
        };
      }
      const online = this.allParameters.filter(
        p => !this.isBentoParamCommOffline(p)
      ).length;
      if (online === total) {
        return {
          label: `Semua Sensor Aktif · ${online}/${total} Online`,
          shortLabel: `${online}/${total} Online`,
          toneClass: 'pill-green',
          dotClass: 'dot-emerald',
        };
      }
      if (online === 0) {
        return {
          label: `Komunikasi Terputus · 0/${total} Online`,
          shortLabel: `0/${total} Online`,
          toneClass: 'pill-rose',
          dotClass: 'dot-rose',
        };
      }
      return {
        label: `Sebagian Aktif · ${online}/${total} Sensor Online`,
        shortLabel: `${online}/${total} Online`,
        toneClass: 'pill-amber',
        dotClass: 'dot-amber',
      };
    },
    allParameters() {
      const baseList = (OperatorCommandCenterView.computed && OperatorCommandCenterView.computed.allParameters)
        ? OperatorCommandCenterView.computed.allParameters.call(this)
        : [];
      return baseList.map((item) => {
        const pid = item.id || item.param_id;
        const live = (pid != null && this.liveMap && this.liveMap[pid]) ? this.liveMap[pid] : {};
        const rawP = (pid != null && this.bentoRawParamMap && this.bentoRawParamMap[pid]) ? this.bentoRawParamMap[pid] : {};
        const formulaStr = String(item.formula || rawP.formula || live.formula || '').trim();
        const hasFormula = formulaStr !== '';
        const qFlags = String(live.quality_flags || rawP.current_quality_flags || '');
        const rawQuality = String(live.quality || rawP.current_quality || item.quality || 'UNKNOWN');
        const rawReason = String(live.quality_reason || rawP.current_quality_reason || item.quality_reason || '');

        let deviceStatus = item.device_status;
        let quality = item.quality;
        let qualityReason = item.quality_reason;
        let receivedAt = item.received_at;

        const hasValidNum = item.value !== null && item.value !== undefined && item.value !== '' && isFinite(Number(item.value));
        const devWasOffline = deviceStatus === 'OFFLINE' || deviceStatus === 'ERROR' || deviceStatus === 'DISCONNECTED' || deviceStatus === 'FAILED';

        if (hasFormula && hasValidNum && devWasOffline) {
          deviceStatus = 'ONLINE';
          quality = (rawQuality && rawQuality !== 'BAD' && rawQuality !== 'UNKNOWN') ? rawQuality : 'GOOD';
          qualityReason = (rawReason === 'COMMUNICATION_ERROR' || rawReason === 'DEVICE_OFFLINE' || rawReason === 'PORT_DISCONNECTED')
            ? 'NONE'
            : rawReason;
          receivedAt = live.received_at || rawP.current_received_at || receivedAt;
        }

        const isFormulaFallback = hasFormula && (
          qFlags.toLowerCase().includes('formula_fallback') ||
          rawReason === 'FORMULA_FALLBACK' ||
          devWasOffline
        );

        return {
          ...item,
          formula: formulaStr,
          device_status: deviceStatus,
          quality,
          quality_reason: qualityReason,
          quality_flags: qFlags,
          is_formula_fallback: isFormulaFallback,
          received_at: receivedAt,
        };
      });
    },
    unplottedParameters() {
      const plottedIds = new Set((this.activeTrendSeries || []).map(s => Number(s.parameterId)));
      return (this.allParameters || []).filter(
        p => !this.isDiscreteStatusParam(p) && !plottedIds.has(Number(p.param_id))
      );
    },
    trendCommDiagnostics() {
      const baseDiag = (OperatorCommandCenterView.computed && OperatorCommandCenterView.computed.trendCommDiagnostics)
        ? OperatorCommandCenterView.computed.trendCommDiagnostics.call(this)
        : { state: 'ONLINE', badge: 'LIVE NORMAL', hasHistory: true };
      if (!baseDiag || (baseDiag.state !== 'DEVICE_OFFLINE' && baseDiag.state !== 'PARTIAL_OFFLINE' && baseDiag.state !== 'SENSOR_TIMEOUT')) {
        return baseDiag;
      }
      const plottedParams = (this.activeTrendSeries || [])
        .map(s => (this.allParameters || []).find(p => Number(p.param_id || p.id) === Number(s.parameterId)))
        .filter(Boolean);
      if (plottedParams.length > 0 && plottedParams.every(p => !this.isBentoParamCommOffline(p))) {
        const anyFormula = plottedParams.some(p => this.isParamFormulaFallback(p));
        return {
          ...baseDiag,
          state: 'ONLINE',
          badge: anyFormula ? 'LIVE FORMULA*' : 'LIVE NORMAL',
          badgeClass: 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 font-medium',
          dotClass: 'bg-emerald-500',
          message: anyFormula
            ? `Nilai aktif via formula (*) · Data terakhir: ${baseDiag.lastTimeFormatted || '-'}`
            : `Komunikasi normal · Data terakhir: ${baseDiag.lastTimeFormatted || '-'}`,
        };
      }
      return baseDiag;
    },
    trendChartOutageMarkers() {
      const baseMarkers = (OperatorCommandCenterView.computed && OperatorCommandCenterView.computed.trendChartOutageMarkers)
        ? OperatorCommandCenterView.computed.trendChartOutageMarkers.call(this)
        : [];
      if (!baseMarkers.length) return baseMarkers;
      const plottedParams = (this.activeTrendSeries || [])
        .map(s => (this.allParameters || []).find(p => Number(p.param_id || p.id) === Number(s.parameterId)))
        .filter(Boolean);
      if (plottedParams.length > 0 && plottedParams.every(p => !this.isBentoParamCommOffline(p) && this.isParamFormulaFallback(p))) {
        return [];
      }
      return baseMarkers;
    },
    fullscreenOutageMarkers() {
      const baseMarkers = (OperatorCommandCenterView.computed && OperatorCommandCenterView.computed.fullscreenOutageMarkers)
        ? OperatorCommandCenterView.computed.fullscreenOutageMarkers.call(this)
        : [];
      if (!baseMarkers.length) return baseMarkers;
      const plottedParams = (this.activeTrendSeries || [])
        .map(s => (this.allParameters || []).find(p => Number(p.param_id || p.id) === Number(s.parameterId)))
        .filter(Boolean);
      if (plottedParams.length > 0 && plottedParams.every(p => !this.isBentoParamCommOffline(p) && this.isParamFormulaFallback(p))) {
        return [];
      }
      return baseMarkers;
    },
    isAudioMuted() {
      const prefs = this.$store && this.$store.state ? this.$store.state.alarmPreferences : null;
      if (prefs && prefs.isMuted !== undefined) {
        if (prefs.isMuted && prefs.muteUntil && Number(this.currentTimeMs || Date.now()) > Number(prefs.muteUntil)) {
          return false;
        }
        return !!prefs.isMuted;
      }
      return !!this.bentoAudioEngineMuted;
    },
    isSoundEnabled() {
      const prefs = this.$store && this.$store.state ? this.$store.state.alarmPreferences : null;
      const prefSound = prefs && prefs.soundEnabled !== undefined
        ? !!prefs.soundEnabled
        : !!this.bentoAudioEngineSoundEnabled;
      const perfSound = !this.perfSettings || this.perfSettings.alarmSoundEnabled !== false;
      return prefSound && perfSound;
    },
  },
  watch: {
    currentUserIdentifier(newId, oldId) {
      if (newId && newId !== oldId) {
        this.initAudioPreferences();
      }
    },
    isAQMSDomain(val) {
      if (!val && this.activeToolbarPopover === 'ispuSettings') {
        this.activeToolbarPopover = null;
      }
    },
    stationLocation: {
      deep: true,
      handler() {
        if (this.stationTypeMode === 'AUTO_GPS' || this.showBrandingModal) {
          this.refreshAutoGpsPlaceName(false);
        }
      },
    },
    modbusRtuGpsStatus(nextStatus) {
      if (
        nextStatus &&
        nextStatus.state === 'LOCKED' &&
        (this.stationTypeMode === 'AUTO_GPS' || (this.showBrandingModal && this.modalStationTypeMode === 'AUTO_GPS'))
      ) {
        const key = `${Number(nextStatus.lat).toFixed(3)},${Number(nextStatus.lng).toFixed(3)}`;
        if (key !== this.lastResolvedGpsKey) {
          this.refreshAutoGpsPlaceName(false);
        }
      }
    },
  },
  mounted() {
    this.initAudioPreferences();
    if (this.stationTypeMode === 'AUTO_GPS') {
      this.refreshAutoGpsPlaceName(false);
    }
    this._unsubBentoAudioState = audioAlert.onStateChange((state) => {
      if (!state) return;
      this.bentoAudioEngineMuted = !!state.isMuted;
      this.bentoAudioEngineSoundEnabled = state.soundEnabled !== false;
    });
    this.ensureBentoFontsLoaded();
    try {
      const savedBentoCols = parseInt(localStorage.getItem('datalogger_bento_grid_cols'), 10);
      if (savedBentoCols >= 1 && savedBentoCols <= 8) {
        this.gridColumns = savedBentoCols;
      }
    } catch (e) {}
    this.trendChartHeight = this.bentoTrendCanvasHeight;
    if (typeof window !== 'undefined') {
      this.bentoViewportWidth = window.innerWidth;
      this.bentoViewportHeight = window.innerHeight;
      window.addEventListener('resize', this.onBentoViewportResize, { passive: true });
    }
    this.measureBentoParamsWrap();
    this.measureBentoPopoverBounds();
    this.$nextTick(() => {
      this.measureBentoParamsWrap();
      this.measureBentoPopoverBounds();
      if (typeof ResizeObserver !== 'undefined' && this.$refs && this.$refs.bentoParamsWrap) {
        this._bentoParamsRO = new ResizeObserver(() => {
          this.measureBentoParamsWrap();
        });
        this._bentoParamsRO.observe(this.$refs.bentoParamsWrap);
      }
    });
  },
  beforeDestroy() {
    this.stopBentoTrendResize();
    if (this._unsubBentoAudioState) {
      this._unsubBentoAudioState();
      this._unsubBentoAudioState = null;
    }
    if (this._bentoParamsRO) {
      this._bentoParamsRO.disconnect();
      this._bentoParamsRO = null;
    }
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', this.onBentoViewportResize);
    }
  },
  methods: {
    openCategoryManagementModal() {
      if (!this.canManageCategories) return;
      this.isCategoryModalOpen = true;
    },
    handleBentoCategoryCreated(newCat) {
      if (this.$refs.bentoParamModal && this.$refs.bentoParamModal.selectNewlyCreatedCategory) {
        this.$refs.bentoParamModal.selectNewlyCreatedCategory(newCat);
      }
    },
    initAudioPreferences() {
      const prefs = getStoredAlarmPreferences(this.currentUserIdentifier);
      if (this.perfSettings && this.perfSettings.alarmSoundEnabled === false && prefs.soundEnabled !== false) {
        prefs.soundEnabled = false;
      }
      applyAlarmPreferences(prefs);
      this.$store.commit('SET_ALARM_PREFERENCES', prefs);
      this.audioVolume = prefs.volume !== undefined ? prefs.volume : 0.7;
      this.audioRepeatInterval = prefs.repeatIntervalSeconds !== undefined ? prefs.repeatIntervalSeconds : 5;
      this.audioSoundEnabled = prefs.soundEnabled !== undefined ? prefs.soundEnabled : true;
      this.audioMuteDuration = prefs.isMuted
        ? (prefs.muteDurationMinutes && prefs.muteDurationMinutes !== 0 ? prefs.muteDurationMinutes : -1)
        : 0;
      this.bentoAudioEngineMuted = audioAlert.checkMuteActive();
      this.bentoAudioEngineSoundEnabled = audioAlert.soundEnabled !== false;
    },
    toggleAudioMute() {
      audioAlert.unlockAudio().catch(() => {});
      const nextMuted = !this.isAudioMuted;
      const current = getStoredAlarmPreferences(this.currentUserIdentifier);
      const updated = saveAlarmPreferences(
        {
          ...current,
          isMuted: nextMuted,
          muteUntil: null,
          muteDurationMinutes: nextMuted ? -1 : 0,
        },
        this.currentUserIdentifier
      );
      this.$store.commit('SET_ALARM_PREFERENCES', updated);
      this.audioMuteDuration = updated.muteDurationMinutes;
      this.bentoAudioEngineMuted = !!updated.isMuted;
      if (!nextMuted && this.isSoundEnabled) {
        audioAlert.previewSound('WARNING');
      }
    },
    onAudioMuteDurationChange() {
      audioAlert.unlockAudio().catch(() => {});
      const dur = Number(this.audioMuteDuration);
      const nextMuted = dur !== 0;
      const muteUntil = dur > 0 ? (Date.now() + dur * 60 * 1000) : null;
      const current = getStoredAlarmPreferences(this.currentUserIdentifier);
      const updated = saveAlarmPreferences(
        {
          ...current,
          isMuted: nextMuted,
          muteDurationMinutes: nextMuted ? dur : 0,
          muteUntil,
        },
        this.currentUserIdentifier
      );
      this.$store.commit('SET_ALARM_PREFERENCES', updated);
      this.audioMuteDuration = updated.muteDurationMinutes;
      this.bentoAudioEngineMuted = !!updated.isMuted;
      if (!nextMuted && this.isSoundEnabled) {
        audioAlert.previewSound('WARNING');
      }
    },
    applyPerformanceSettings(changedKey = null, changedVal = null) {
      if (!this.perfSettings) return;
      // 1. Alarm Sound — preserve persisted mute state (never overwrite isMuted with !soundOn)
      if (changedKey === 'alarmSoundEnabled' || changedKey === null) {
        const soundOn = this.perfSettings.alarmSoundEnabled !== false;
        const stored = getStoredAlarmPreferences(this.currentUserIdentifier);
        if (changedKey === 'alarmSoundEnabled') {
          const updated = saveAlarmPreferences(
            { ...stored, soundEnabled: soundOn },
            this.currentUserIdentifier
          );
          this.$store.commit('SET_ALARM_PREFERENCES', updated);
          this.audioSoundEnabled = updated.soundEnabled;
        } else {
          const effective = soundOn ? stored : { ...stored, soundEnabled: false };
          applyAlarmPreferences(effective);
          this.$store.commit('SET_ALARM_PREFERENCES', effective);
          this.audioSoundEnabled = effective.soundEnabled !== false;
          this.audioMuteDuration = effective.isMuted
            ? (effective.muteDurationMinutes && effective.muteDurationMinutes !== 0 ? effective.muteDurationMinutes : -1)
            : 0;
        }
        this.bentoAudioEngineMuted = audioAlert.checkMuteActive();
        this.bentoAudioEngineSoundEnabled = audioAlert.soundEnabled !== false;
      }
      // 2. Trend Analysis
      if (changedKey === 'trendAnalysisEnabled') {
        if (changedVal === true) {
          this.fetchTrendData(true);
          this.startTrendTimers();
        } else {
          this.stopTrendTimers();
        }
      }
      // 3. Sparklines
      if (changedKey === 'parameterSparklinesEnabled' && changedVal === true) {
        this.fetchInitialSparklines();
      }
      // 4. Alarm Visual
      if (changedKey === 'alarmVisualEnabled' && changedVal === false) {
        this.isAlarmPulseActive = false;
        if (this.alarmPulseTimer) {
          clearTimeout(this.alarmPulseTimer);
          this.alarmPulseTimer = null;
        }
      }
    },
    measureBentoParamsWrap() {
      const el = this.$refs && this.$refs.bentoParamsWrap;
      if (el && el.clientWidth > 0) {
        this.bentoParamsWrapWidth = el.clientWidth;
      }
    },
    measureBentoPopoverBounds() {
      if (typeof window === 'undefined') return;
      this.bentoViewportHeight = window.innerHeight;
      if (this.$el && this.$el.querySelector) {
        const headerEl = this.$el.querySelector('.header-shell');
        if (headerEl) {
          const rect = headerEl.getBoundingClientRect();
          if (rect && rect.bottom > 0) {
            this.bentoPopoverAnchorBottomPx = Math.round(rect.bottom);
          }
        }
      }
    },
    onBentoViewportResize() {
      if (typeof window === 'undefined') return;
      this.bentoViewportWidth = window.innerWidth;
      this.bentoViewportHeight = window.innerHeight;
      this.measureBentoParamsWrap();
      this.measureBentoPopoverBounds();
    },
    onDocumentClick(e) {
      if (this.bentoActiveMiniPopover && e && e.target && !e.target.closest('.bento-mini-popover-anchor')) {
        this.bentoActiveMiniPopover = null;
      }
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.onDocumentClick) {
        OperatorCommandCenterView.methods.onDocumentClick.call(this, e);
      }
    },
    async initCompanyBranding() {
      let hadSavedLocal = false;
      try {
        hadSavedLocal = !!localStorage.getItem('datalogger.organization_branding');
      } catch (e) {}
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.initCompanyBranding) {
        await OperatorCommandCenterView.methods.initCompanyBranding.call(this);
      }
      try {
        const res = await fetch('/api/system/branding');
        if (res && res.ok) {
          const json = await res.json();
          const remote = (json && json.data) ? json.data : json;
          if (remote) {
            this.customerCompanyName = remote.customer_name || '';
            this.customerSubtitle = remote.customer_subtitle === 'Local Autonomous Edge' ? '' : (remote.customer_subtitle || '');
            this.stationTypeMode = remote.station_type_mode === 'AUTO_GPS' ? 'AUTO_GPS' : 'MANUAL';
            this.stationDomainPrefix = remote.station_domain_prefix || inferStationDomainPrefix(this.customerSubtitle, 'AQMS');
            this.makerCompanyName = remote.maker_name || '';
            this.makerSubtitle = remote.maker_subtitle || '';
            this.companyName = this.makerCompanyName;
            if (this.stationTypeMode === 'AUTO_GPS') {
              this.refreshAutoGpsPlaceName(false);
            }
          }
        } else if (!hadSavedLocal) {
          this.customerCompanyName = '';
          this.customerSubtitle = '';
          this.makerCompanyName = '';
          this.makerSubtitle = '';
          this.companyName = '';
        }
      } catch (e) {
        if (!hadSavedLocal) {
          this.customerCompanyName = '';
          this.customerSubtitle = '';
          this.makerCompanyName = '';
          this.makerSubtitle = '';
          this.companyName = '';
        }
      }
    },
    onOrganizationBrandingUpdated(e) {
      if (e && e.detail && e.detail.customer) {
        if (e.detail.customer.stationTypeMode !== undefined) {
          this.stationTypeMode = e.detail.customer.stationTypeMode === 'AUTO_GPS' ? 'AUTO_GPS' : 'MANUAL';
        }
        if (e.detail.customer.domainPrefix !== undefined) {
          this.stationDomainPrefix = e.detail.customer.domainPrefix || 'AQMS';
        }
      }
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.onOrganizationBrandingUpdated) {
        OperatorCommandCenterView.methods.onOrganizationBrandingUpdated.call(this, e);
      }
    },
    openBrandingModal(tab = 'customer') {
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.openBrandingModal) {
        OperatorCommandCenterView.methods.openBrandingModal.call(this, tab);
      }
      this.modalStationTypeMode = this.stationTypeMode === 'AUTO_GPS' ? 'AUTO_GPS' : 'MANUAL';
      this.modalStationDomainPrefix = this.stationDomainPrefix || inferStationDomainPrefix(this.customerSubtitle, 'AQMS');
      this.modalResolvedPlaceName = this.resolvedGpsPlaceName || '';
      if (this.modalStationTypeMode === 'AUTO_GPS' || this.activeAutoGpsCoords) {
        this.refreshAutoGpsPlaceName(false);
      }
    },
    isModalDomainChipActive(st) {
      if (this.modalStationTypeMode === 'AUTO_GPS') {
        return String(this.modalStationDomainPrefix || '').toLowerCase() === String(st).toLowerCase();
      }
      const sub = String(this.modalCustomerSubtitle || '').trim();
      if (sub.toLowerCase() === String(st).toLowerCase()) return true;
      return new RegExp(`^${st}\\b`, 'i').test(sub);
    },
    selectModalDomainPrefix(st) {
      this.modalStationDomainPrefix = st;
      if (this.modalStationTypeMode !== 'AUTO_GPS') {
        const cur = String(this.modalCustomerSubtitle || '').trim();
        if (!cur || /^(AQMS|WQMS|Sparing)$/i.test(cur)) {
          this.modalCustomerSubtitle = st;
        } else if (/^(AQMS|WQMS|Sparing)\b\s+/i.test(cur)) {
          this.modalCustomerSubtitle = cur.replace(/^(AQMS|WQMS|Sparing)\b\s+/i, `${st} `);
        } else {
          this.modalCustomerSubtitle = `${st} ${cur}`;
        }
      }
    },
    setModalStationTypeMode(mode) {
      this.modalStationTypeMode = mode === 'AUTO_GPS' ? 'AUTO_GPS' : 'MANUAL';
      if (this.modalStationTypeMode === 'AUTO_GPS') {
        if (!this.modalStationDomainPrefix) {
          this.modalStationDomainPrefix = inferStationDomainPrefix(this.modalCustomerSubtitle, 'AQMS');
        }
        this.refreshAutoGpsPlaceName(false);
      }
    },
    async refreshAutoGpsPlaceName(force = false) {
      const coords = this.activeAutoGpsCoords;
      if (!coords) {
        const saved = this.fallbackSavedPlaceName;
        if (saved) {
          this.resolvedGpsPlaceName = saved;
          this.modalResolvedPlaceName = saved;
        }
        return;
      }

      const key = `${coords.lat.toFixed(3)},${coords.lng.toFixed(3)}`;
      if (!force && key === this.lastResolvedGpsKey && this.resolvedGpsPlaceName) {
        this.modalResolvedPlaceName = this.resolvedGpsPlaceName;
        return;
      }

      this.isResolvingAutoPlace = true;
      try {
        const res = await reverseGeocodePlaceName(coords.lat, coords.lng);
        const cleanPlace = extractConcisePlaceName(res && res.placeName) || this.fallbackSavedPlaceName;
        if (cleanPlace) {
          this.resolvedGpsPlaceName = cleanPlace;
          this.modalResolvedPlaceName = cleanPlace;
          this.lastResolvedGpsKey = key;
          try {
            localStorage.setItem('datalogger_cached_gps_place_name', cleanPlace);
          } catch (e) {}
        }
      } catch (e) {
        const fallback = this.fallbackSavedPlaceName;
        if (fallback) {
          this.resolvedGpsPlaceName = fallback;
          this.modalResolvedPlaceName = fallback;
        }
      } finally {
        this.isResolvingAutoPlace = false;
      }
    },
    async saveBrandingFromModal() {
      this.brandingSaving = true;
      this.logoSaving = true;
      this.brandingErrorMsg = '';
      try {
        const nextMode = this.modalStationTypeMode === 'AUTO_GPS' ? 'AUTO_GPS' : 'MANUAL';
        const nextPrefix = this.modalStationDomainPrefix || inferStationDomainPrefix(this.modalCustomerSubtitle, 'AQMS');
        const subtitleToSave = nextMode === 'AUTO_GPS'
          ? composeAutoGpsStationType(nextPrefix, this.effectiveModalPlaceName)
          : this.modalCustomerSubtitle;

        await saveCustomerBranding({
          file: this.modalCustomerLogoFile,
          name: this.modalCustomerName,
          subtitle: subtitleToSave,
          stationTypeMode: nextMode,
          domainPrefix: nextPrefix,
        });

        await saveMakerBranding({
          file: this.modalMakerLogoFile,
          name: this.modalMakerName,
          subtitle: this.modalMakerSubtitle,
        });

        this.stationTypeMode = nextMode;
        this.stationDomainPrefix = nextPrefix;
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
    onDocumentContextMenu(e) {
      if (!this.paramContextMenu || !this.paramContextMenu.visible) return;
      if (
        e &&
        e.target &&
        !e.target.closest('.s-card') &&
        !e.target.closest('.bento-tbl-row') &&
        !e.target.closest('.scada-param-tile') &&
        !e.target.closest('.param-context-menu')
      ) {
        this.closeParamContextMenu();
      }
    },
    enrichParamForViewModal(raw) {
      if (!raw) return raw;
      return {
        ...raw,
        low_warning: raw.low_warning ?? raw.warning_low ?? null,
        high_warning: raw.high_warning ?? raw.warning_high ?? null,
        low_critical: raw.low_critical ?? raw.critical_low ?? raw.min_value ?? raw.low_limit ?? null,
        high_critical: raw.high_critical ?? raw.critical_high ?? raw.max_value ?? raw.high_limit ?? null,
      };
    },
    handleContextViewConfig() {
      const { device, fullParam, param } = this.paramContextMenu || {};
      this.closeParamContextMenu();
      if (!param) return;
      const raw = this.bentoRawParamMap[param.param_id || param.id] || fullParam || param;
      this.viewConfigTarget = {
        device: device || { id: param.device_id, device_code: param.device_code },
        parameter: this.enrichParamForViewModal(raw),
      };
      this.isViewConfigModalOpen = true;
    },
    openBentoQuickLimitEdit(param) {
      if (!param) return;
      const dev = (this.devicesList || []).find(d => Number(d.id) === Number(param.device_id)) || {
        id: param.device_id,
        device_code: param.device_code,
      };
      const raw = this.bentoRawParamMap[param.param_id || param.id] || param;
      if (this.canEditConfig) {
        this.editParamTarget = {
          deviceId: dev.id || param.device_id,
          device: dev,
          parameter: raw,
        };
        this.isEditParamModalOpen = true;
      } else {
        this.viewConfigTarget = {
          device: dev,
          parameter: this.enrichParamForViewModal(raw),
        };
        this.isViewConfigModalOpen = true;
      }
    },
    onDocumentKeyDown(e) {
      if (e && e.key === 'Escape' && this.bentoActiveMiniPopover) {
        this.bentoActiveMiniPopover = null;
      }
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.onDocumentKeyDown) {
        OperatorCommandCenterView.methods.onDocumentKeyDown.call(this, e);
      }
    },
    toggleToolbarPopover(popoverName) {
      this.bentoActiveMiniPopover = null;
      this.measureBentoPopoverBounds();
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.toggleToolbarPopover) {
        OperatorCommandCenterView.methods.toggleToolbarPopover.call(this, popoverName);
      }
      if (popoverName === 'stationLocation' && this.activeToolbarPopover === 'stationLocation') {
        if (this.locationActiveTab === 'gps') {
          this.locationActiveTab = 'modbus_rtu';
        } else if (this.stationLocation && this.stationLocation.location_source === 'MODBUS_RTU') {
          this.locationActiveTab = 'modbus_rtu';
        } else if (this.stationLocation && this.stationLocation.location_source === 'IP_GEOLOCATION') {
          this.locationActiveTab = 'ip';
        }
        this.$nextTick(() => {
          this.measureBentoPopoverBounds();
          if (this.$refs.miniMapViewer && this.$refs.miniMapViewer.invalidateSize) {
            this.$refs.miniMapViewer.invalidateSize();
          }
        });
      }
    },
    toggleTrendDownloadMenu() {
      this.bentoActiveMiniPopover = null;
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.toggleTrendDownloadMenu) {
        OperatorCommandCenterView.methods.toggleTrendDownloadMenu.call(this);
      }
    },
    toggleLineInterpolationMenu() {
      this.bentoActiveMiniPopover = null;
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.toggleLineInterpolationMenu) {
        OperatorCommandCenterView.methods.toggleLineInterpolationMenu.call(this);
      }
    },
    toggleBentoMiniPopover(name) {
      if (this.bentoActiveMiniPopover === name) {
        this.bentoActiveMiniPopover = null;
        return;
      }
      this.activeToolbarPopover = null;
      this.showTrendSettingsMenu = false;
      this.showTrendDownloadMenu = false;
      this.showLineSmoothMenu = false;
      this.showAudioSetupPopover = false;
      this.bentoActiveMiniPopover = name;
    },
    getBentoCategoryCount(catId) {
      return (this.allParameters || []).filter(p => Number(p.category_id) === Number(catId)).length;
    },
    getTrendParamDotColor(param) {
      if (!param) return '#0284C7';
      const plotted = (this.activeTrendSeries || []).find(
        s => Number(s.parameterId) === Number(param.param_id)
      );
      if (plotted && plotted.color) return plotted.color;
      return this.getBentoCardColor(param);
    },
    setBentoTrendHeight(val) {
      const parsed = Math.round(Number(val));
      if (!isFinite(parsed)) return;
      const clamped = Math.min(560, Math.max(120, parsed));
      this.bentoTrendCanvasHeight = clamped;
      this.trendChartHeight = clamped;
      try {
        localStorage.setItem('datalogger_bento_trend_height', String(clamped));
      } catch (e) {}
    },
    stepBentoTrendHeight(delta) {
      this.setBentoTrendHeight((this.bentoTrendCanvasHeight || 205) + Number(delta || 0));
    },
    startBentoTrendResize(e) {
      if (typeof window === 'undefined') return;
      this.stopBentoTrendResize();
      const startY = e.clientY;
      const startH = this.bentoTrendCanvasHeight || 205;
      this.isResizingBentoTrend = true;
      this._onBentoTrendMouseMove = (ev) => {
        const dy = ev.clientY - startY;
        this.setBentoTrendHeight(startH + dy);
      };
      this._onBentoTrendMouseUp = () => {
        this.stopBentoTrendResize();
      };
      window.addEventListener('mousemove', this._onBentoTrendMouseMove);
      window.addEventListener('mouseup', this._onBentoTrendMouseUp);
    },
    startBentoTrendResizeTouch(e) {
      if (typeof window === 'undefined' || !e.touches || !e.touches[0]) return;
      this.stopBentoTrendResize();
      const startY = e.touches[0].clientY;
      const startH = this.bentoTrendCanvasHeight || 205;
      this.isResizingBentoTrend = true;
      this._onBentoTrendTouchMove = (ev) => {
        if (!ev.touches || !ev.touches[0]) return;
        const dy = ev.touches[0].clientY - startY;
        this.setBentoTrendHeight(startH + dy);
      };
      this._onBentoTrendTouchEnd = () => {
        this.stopBentoTrendResize();
      };
      window.addEventListener('touchmove', this._onBentoTrendTouchMove, { passive: true });
      window.addEventListener('touchend', this._onBentoTrendTouchEnd);
    },
    stopBentoTrendResize() {
      this.isResizingBentoTrend = false;
      if (typeof window !== 'undefined') {
        if (this._onBentoTrendMouseMove) {
          window.removeEventListener('mousemove', this._onBentoTrendMouseMove);
          this._onBentoTrendMouseMove = null;
        }
        if (this._onBentoTrendMouseUp) {
          window.removeEventListener('mouseup', this._onBentoTrendMouseUp);
          this._onBentoTrendMouseUp = null;
        }
        if (this._onBentoTrendTouchMove) {
          window.removeEventListener('touchmove', this._onBentoTrendTouchMove);
          this._onBentoTrendTouchMove = null;
        }
        if (this._onBentoTrendTouchEnd) {
          window.removeEventListener('touchend', this._onBentoTrendTouchEnd);
          this._onBentoTrendTouchEnd = null;
        }
      }
    },
    ensureBentoFontsLoaded() {
      if (typeof document === 'undefined') return;
      const id = 'bento-jakarta-inter-fonts';
      if (!document.getElementById(id)) {
        const link = document.createElement('link');
        link.id = id;
        link.rel = 'stylesheet';
        link.href = 'https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Plus+Jakarta+Sans:wght@500;600;700;800&display=swap';
        document.head.appendChild(link);
      }
    },
    setBentoTheme(mode) {
      this.$store.dispatch('setTheme', mode);
    },
    setBentoDisplayMode(target) {
      if (target === 'cards') {
        this.viewMode = 'cards';
        this.bentoCardGroupByCategory = false;
        try {
          localStorage.setItem('datalogger_bento_card_group_cat', 'false');
        } catch (e) {}
        return;
      }
      if (target === 'cat-cards') {
        this.viewMode = 'cards';
        this.bentoCardGroupByCategory = true;
        try {
          localStorage.setItem('datalogger_bento_card_group_cat', 'true');
        } catch (e) {}
        return;
      }
      if (target === 'auto-cards') {
        this.viewMode = 'cards';
        this.bentoAutoGrid = true;
        try {
          localStorage.setItem('datalogger_bento_auto_grid', 'true');
        } catch (e) {}
        return;
      }
      if (target === 'CAT' || target === '2T' || target === '1T') {
        this.viewMode = 'table';
        this.bentoTableSplitMode = target;
        try {
          localStorage.setItem('datalogger_bento_table_split', target);
        } catch (e) {}
      }
    },
    setBentoCardGrouping(grouped) {
      const flag = Boolean(grouped);
      this.viewMode = 'cards';
      this.bentoCardGroupByCategory = flag;
      try {
        localStorage.setItem('datalogger_bento_card_group_cat', flag ? 'true' : 'false');
      } catch (e) {}
    },
    toggleBentoCatCardCollapse(secKey) {
      if (!secKey) return;
      const next = !this.bentoCollapsedCatCards[secKey];
      this.$set(this.bentoCollapsedCatCards, secKey, next);
    },
    isBentoCatCardCollapsed(secKey) {
      return Boolean(secKey && this.bentoCollapsedCatCards && this.bentoCollapsedCatCards[secKey]);
    },
    setGridColumns(cols) {
      const maxCols = this.bentoCardGroupByCategory ? 8 : 5;
      const count = Math.min(maxCols, Math.max(1, parseInt(cols, 10) || 4));
      const legacySafeCols = Math.min(5, count);
      this.bentoAutoGrid = false;
      this.viewMode = 'cards';
      this.gridColumns = count;
      this.panelCount = legacySafeCols;
      try {
        localStorage.setItem('datalogger_bento_auto_grid', 'false');
        localStorage.setItem('datalogger_bento_grid_cols', String(count));
        localStorage.setItem('operator_monitoring_grid_cols', String(legacySafeCols));
        localStorage.setItem('operator_console_grid_columns', String(legacySafeCols));
      } catch (e) {}
      if (typeof this.savePanelLayoutConfig === 'function') {
        this.savePanelLayoutConfig();
      }
    },
    selectBentoAllParams() {
      this.selectedCategoryFilter = 'all';
      if (this.filterPinnedOnly) {
        this.filterPinnedOnly = false;
      }
    },
    selectBentoCategory(catId) {
      if (this.filterPinnedOnly) {
        this.filterPinnedOnly = false;
      }
      if (Number(this.selectedCategoryFilter) === Number(catId)) {
        this.selectedCategoryFilter = 'all';
      } else {
        this.selectedCategoryFilter = catId;
      }
    },
    isBentoParamCommOffline(param) {
      if (!param) return true;
      const devOnline = param.device_status === 'ONLINE' || param.device_status === 'CONNECTED';
      if (!devOnline) return true;
      if (param.quality === 'BAD' && param.quality_reason !== 'OUT_OF_HARD_RANGE') {
        return true;
      }
      return false;
    },
    formatBentoLimitNum(num) {
      if (num === null || num === undefined || !isFinite(Number(num))) return '';
      return Number(num).toLocaleString('id-ID', { maximumFractionDigits: 2 });
    },
    resolveBentoParamLimits(param) {
      if (!param || this.isDiscreteStatusParam(param)) return null;
      const pid = param.param_id || param.id;
      const raw = (pid != null && this.bentoRawParamMap && this.bentoRawParamMap[pid]) || param;
      const toFinite = (v) => (v !== null && v !== undefined && v !== '' && isFinite(Number(v)) ? Number(v) : null);

      const warnHigh = toFinite(raw.warning_high ?? raw.critical_high ?? raw.high_warning ?? raw.high_critical);
      const warnLow = toFinite(raw.warning_low ?? raw.critical_low ?? raw.low_warning ?? raw.low_critical);
      const maxVal = toFinite(raw.max_value ?? raw.high_limit);
      const minVal = toFinite(raw.min_value ?? raw.low_limit);
      const zeroSuppress = (raw.zero_value_alarm_policy || param.zero_value_alarm_policy) === 'SUPPRESS_THRESHOLD';

      // 1. Direct Database Configuration from ParameterModal (Edit Config)
      if (warnHigh !== null || maxVal !== null || warnLow !== null || minVal !== null) {
        if (warnHigh !== null) {
          const lower = warnLow !== null ? warnLow : minVal;
          const hasLow = lower !== null && lower !== 0;
          const fmtUp = this.formatBentoLimitNum(warnHigh);
          const fmtLow = hasLow ? this.formatBentoLimitNum(lower) : '0';
          return {
            source: 'configured',
            upper: warnHigh,
            lower: lower !== null ? lower : 0,
            warnHigh,
            warnLow,
            hardMax: maxVal,
            hardMin: minVal,
            zeroSuppress,
            shortLabel: hasLow ? `${fmtLow}–${fmtUp}` : `Limit ${fmtUp}`,
            longLabel: hasLow ? `Batas ${fmtLow}–${fmtUp}` : `Limit ${fmtUp}`,
            detailNote: `Warning High: ${fmtUp}${maxVal !== null ? ' · Max Limit: ' + this.formatBentoLimitNum(maxVal) : ''}`,
          };
        }
        if (maxVal !== null) {
          const lower = warnLow !== null ? warnLow : minVal;
          const isRange = lower !== null;
          const fmtUp = this.formatBentoLimitNum(maxVal);
          const fmtLow = isRange ? this.formatBentoLimitNum(lower) : '0';
          return {
            source: 'configured',
            upper: maxVal,
            lower: isRange ? lower : 0,
            warnHigh: null,
            warnLow,
            hardMax: maxVal,
            hardMin: minVal,
            zeroSuppress,
            shortLabel: isRange ? `${fmtLow}–${fmtUp}` : `Limit ${fmtUp}`,
            longLabel: isRange ? `Rentang ${fmtLow}–${fmtUp}` : `Limit ${fmtUp}`,
            detailNote: isRange ? `Min–Max Limit: ${fmtLow} s/d ${fmtUp}` : `Max Limit: ${fmtUp}`,
          };
        }
        const onlyLow = warnLow !== null ? warnLow : minVal;
        const fmtOnlyLow = this.formatBentoLimitNum(onlyLow);
        return {
          source: 'configured',
          upper: null,
          lower: onlyLow,
          warnHigh: null,
          warnLow,
          hardMax: null,
          hardMin: minVal,
          zeroSuppress,
          shortLabel: `Min ${fmtOnlyLow}`,
          longLabel: `Batas Min ${fmtOnlyLow}`,
          detailNote: `Batas Minimum: ${fmtOnlyLow}`,
        };
      }

      // 2. Fallback Reference ONLY when parameter has no configured limits in Edit Config
      const keyStr = ((param.param_code || '') + ' ' + (param.param_name || '')).toUpperCase();
      let refUpper = null;
      if (keyStr.includes('PM2.5') || keyStr.includes('PM25') || keyStr.includes('PM 2.5')) refUpper = REGULATORY_REF_LIMITS.PM25;
      else if (keyStr.includes('PM10') || keyStr.includes('PM 10')) refUpper = REGULATORY_REF_LIMITS.PM10;
      else if (keyStr.includes('SO2') || keyStr.includes('SULFUR')) refUpper = REGULATORY_REF_LIMITS.SO2;
      else if (keyStr.includes('NO2') || keyStr.includes('NITROGEN')) refUpper = REGULATORY_REF_LIMITS.NO2;
      else if (keyStr.includes('O3') || keyStr.includes('OZON')) refUpper = REGULATORY_REF_LIMITS.O3;
      else if (keyStr.includes('CO') || keyStr.includes('KARBON')) refUpper = REGULATORY_REF_LIMITS.CO;
      else if (keyStr.includes('HC') || keyStr.includes('HIDROKARBON')) refUpper = REGULATORY_REF_LIMITS.HC;

      if (refUpper !== null) {
        const fmtRef = this.formatBentoLimitNum(refUpper);
        return {
          source: 'default',
          upper: refUpper,
          lower: 0,
          warnHigh: null,
          warnLow: null,
          hardMax: null,
          hardMin: null,
          zeroSuppress,
          shortLabel: `Limit ${fmtRef}`,
          longLabel: `Baku mutu ${fmtRef}`,
          detailNote: `Referensi default (${fmtRef})`,
        };
      }

      const unitStr = (param.unit || '').trim().toLowerCase();
      const isRainParam = keyStr.includes('RAIN') || keyStr.includes('HUJAN') || keyStr.includes('CURAH');
      const isTempUnit = unitStr === 'c' || unitStr === '°c' || unitStr === 'celcius' || unitStr === 'celsius';
      if (!isRainParam && (keyStr.includes('TEMP') || keyStr.includes('SUHU') || isTempUnit)) {
        return {
          source: 'default',
          upper: 50,
          lower: 0,
          warnHigh: null,
          warnLow: null,
          hardMax: null,
          hardMin: null,
          zeroSuppress,
          shortLabel: '0–50',
          longLabel: 'Rentang 0–50',
          detailNote: 'Rentang referensi suhu (0–50)',
        };
      }
      if (unitStr === '%' || keyStr.includes('HUMID') || keyStr.includes('LEMBAB')) {
        return {
          source: 'default',
          upper: 100,
          lower: 0,
          warnHigh: null,
          warnLow: null,
          hardMax: null,
          hardMin: null,
          zeroSuppress,
          shortLabel: '0–100',
          longLabel: 'Rentang 0–100',
          detailNote: 'Rentang referensi kelembaban (0–100%)',
        };
      }
      return null;
    },
    resolveRegulatoryLimit(param) {
      const lim = this.resolveBentoParamLimits(param);
      return lim && lim.upper !== null ? lim.upper : null;
    },
    getBentoParamActiveAlarm(param) {
      if (!param) return null;
      const pid = Number(param.param_id || param.id);
      const entry = (pid && this.bentoParamActiveAlarmMap && this.bentoParamActiveAlarmMap[pid]) || null;
      if (!entry || !entry.primary) return null;

      const alm = entry.primary;
      const code = String(alm.alarm_code || alm.alarm_type || alm.type || '').toUpperCase();
      const sev = String(alm.severity || 'WARNING').toUpperCase();
      const isCrit = sev === 'CRITICAL' || code.includes('CRITICAL');

      let badgeLabel = isCrit ? 'Kritis' : '⚠ Alarm';
      let shortMeta = isCrit ? 'Alarm Kritis' : 'Alarm Aktif';

      if (code.includes('STUCK')) {
        badgeLabel = '⚠ Stuck';
        shortMeta = 'Stuck';
      } else if (code.includes('SPIKE')) {
        badgeLabel = '⚠ Spike';
        shortMeta = 'Spike';
      } else if (code.includes('ZERO_INVALID') || code.includes('ZERO')) {
        badgeLabel = '⚠ Zero';
        shortMeta = 'Zero Invalid';
      } else if (code.includes('STALE')) {
        badgeLabel = '⚠ Stale';
        shortMeta = 'Stale';
      } else if (code.includes('BAD_QUALITY')) {
        badgeLabel = '⚠ Bad';
        shortMeta = 'Bad Data';
      } else if (code.includes('THRESHOLD_WARNING') || code === 'WARNING') {
        badgeLabel = 'Waspada';
        shortMeta = 'Warning Limit';
      } else if (code.includes('THRESHOLD_CRITICAL')) {
        badgeLabel = 'Kritis';
        shortMeta = 'Critical Limit';
      }

      const trigTime = alm.triggered_at || alm.created_at;
      const ageStr = trigTime ? this.formatRelative(trigTime) : '';
      const msgStr = alm.message || code || 'Active Alarm';
      const extraStr = entry.count > 1 ? ' (+' + (entry.count - 1) + ' alarm aktif lainnya)' : '';
      const tooltipText = 'Alarm Aktif [' + sev + ']: ' + msgStr + (ageStr ? ' · ' + ageStr : '') + extraStr + ' — Klik badge untuk Acknowledge';

      return {
        primary: alm,
        count: entry.count,
        all: entry.all,
        code,
        severity: isCrit ? 'CRITICAL' : 'WARNING',
        toneClass: isCrit ? 'badge-rose' : 'badge-amber',
        badgeLabel,
        shortMeta,
        ageStr,
        tooltipText,
      };
    },
    getBentoDotColor(param) {
      if (this.isBentoParamCommOffline(param)) return '#F43F5E';
      const tone = this.getBentoStatusBadge(param).toneClass;
      if (tone === 'badge-rose') return '#F43F5E';
      if (tone === 'badge-amber') return '#F59E0B';
      return '#10B981';
    },
    getBentoParamCardTitle(param) {
      const baseTitle = typeof this.getParamTimingTitle === 'function' ? this.getParamTimingTitle(param) : '';
      const activeInfo = this.getBentoParamActiveAlarm(param);
      if (!activeInfo) return baseTitle;
      return baseTitle ? (activeInfo.tooltipText + '\n' + baseTitle) : activeInfo.tooltipText;
    },
    onBentoStatusBadgeClick(param) {
      const badge = this.getBentoStatusBadge(param);
      if (badge && badge.isAlarm && badge.alarm && typeof this.openAckModal === 'function') {
        this.openAckModal(badge.alarm);
      }
    },
    getBentoCardAlertClass(param) {
      const tone = this.getBentoStatusBadge(param).toneClass;
      if (tone === 'badge-rose') return 's-card-fault';
      if (tone === 'badge-amber') return 's-card-warn';
      return '';
    },
    getBentoRowAlertClass(param) {
      const tone = this.getBentoStatusBadge(param).toneClass;
      if (tone === 'badge-rose') return 'is-row-fault';
      if (tone === 'badge-amber') return 'is-row-warn';
      return '';
    },
    deviceHasConfiguredFormula(dev) {
      if (!dev || !Array.isArray(dev.parameters)) return false;
      return dev.parameters.some(
        p => p && p.enabled !== false && p.enabled !== 0 && String(p.formula || '').trim() !== ''
      );
    },
    isParamFormulaFallback(param) {
      if (!param) return false;
      if (param.is_formula_fallback) return true;
      const formulaStr = String(param.formula || '').trim();
      if (!formulaStr) return false;
      const flags = String(param.quality_flags || '').toLowerCase();
      return flags.includes('formula_fallback') || param.quality_reason === 'FORMULA_FALLBACK';
    },
    getBentoStatusBadge(param) {
      const res = this.resolveCoreBentoStatusBadge(param);
      if (!param || this.isBentoParamCommOffline(param) || !this.isParamFormulaFallback(param)) {
        return res;
      }
      const baseLabel = String(res.label || 'Baik');
      const nextLabel = baseLabel.endsWith('*') ? baseLabel : `${baseLabel}*`;
      const fallbackNote = '(*) Sensor fisik tidak terhubung — nilai dihitung otomatis dari formula';
      const nextTooltip = res.tooltip ? `${res.tooltip} · ${fallbackNote}` : fallbackNote;
      return {
        ...res,
        label: nextLabel,
        tooltip: nextTooltip,
      };
    },
    resolveCoreBentoStatusBadge(param) {
      if (!param) return { label: 'Normal', toneClass: 'badge-green', isAlarm: false, alarm: null, tooltip: '' };
      const activeInfo = this.getBentoParamActiveAlarm(param);

      if (this.isBentoParamCommOffline(param)) {
        return {
          label: 'Offline',
          toneClass: 'badge-rose',
          isAlarm: Boolean(activeInfo),
          alarm: activeInfo ? activeInfo.primary : null,
          tooltip: activeInfo ? activeInfo.tooltipText : 'Sensor tidak terhubung / Offline',
        };
      }
      if (this.isDiscreteStatusParam(param)) {
        if (activeInfo && !activeInfo.code.includes('STUCK') && !activeInfo.code.includes('THRESHOLD')) {
          return {
            label: activeInfo.badgeLabel,
            toneClass: activeInfo.toneClass,
            isAlarm: true,
            alarm: activeInfo.primary,
            tooltip: activeInfo.tooltipText,
          };
        }
        const info = this.getDiscreteStatusInfo(param);
        return { label: info.label || 'Aktif', toneClass: 'badge-teal', isAlarm: false, alarm: null, tooltip: '' };
      }

      const limits = this.resolveBentoParamLimits(param);
      const val = Number(param.value);
      const hasVal = param.value !== null && param.value !== undefined && param.value !== '' && isFinite(val);
      const isZeroSuppressed = Boolean(limits && limits.zeroSuppress && hasVal && Math.abs(val) < 1e-9);

      if (
        param.alarm_status === 'CRITICAL' ||
        (activeInfo && activeInfo.severity === 'CRITICAL') ||
        (param.quality === 'BAD' && param.quality_reason === 'OUT_OF_HARD_RANGE') ||
        (hasVal && limits && (
          (limits.hardMax !== null && val > limits.hardMax) ||
          (!isZeroSuppressed && limits.hardMin !== null && val < limits.hardMin)
        ))
      ) {
        return {
          label: activeInfo && activeInfo.severity === 'CRITICAL' ? activeInfo.badgeLabel : 'Kritis',
          toneClass: 'badge-rose',
          isAlarm: Boolean(activeInfo),
          alarm: activeInfo ? activeInfo.primary : null,
          tooltip: activeInfo ? activeInfo.tooltipText : 'Melewati batas kritis parameter',
        };
      }

      const isThresholdWarn =
        param.alarm_status === 'WARNING' ||
        param.quality_reason === 'OUT_OF_WARNING_RANGE' ||
        (hasVal && limits && (
          (limits.warnHigh !== null && val >= limits.warnHigh) ||
          (!isZeroSuppressed && limits.warnLow !== null && val <= limits.warnLow)
        ));

      if (activeInfo) {
        return {
          label: isThresholdWarn && activeInfo.code.includes('THRESHOLD') ? 'Waspada' : activeInfo.badgeLabel,
          toneClass: activeInfo.toneClass,
          isAlarm: true,
          alarm: activeInfo.primary,
          tooltip: activeInfo.tooltipText,
        };
      }

      if (isThresholdWarn) {
        return {
          label: 'Waspada',
          toneClass: 'badge-amber',
          isAlarm: false,
          alarm: null,
          tooltip: 'Melewati batas peringatan (Warning Limit)',
        };
      }
      if (param.quality === 'STALE') {
        return {
          label: 'Stale',
          toneClass: 'badge-amber',
          isAlarm: false,
          alarm: null,
          tooltip: 'Data sensor belum diperbarui (Stale)',
        };
      }
      if (hasVal && limits && limits.upper !== null) {
        const base = limits.lower !== null ? limits.lower : 0;
        const span = limits.upper - base;
        if (span > 0) {
          const ratio = (val - base) / span;
          if (ratio >= 1.0) {
            return {
              label: limits.source === 'configured' ? 'Waspada' : 'Tinggi',
              toneClass: 'badge-amber',
              isAlarm: false,
              alarm: null,
              tooltip: '',
            };
          }
          if (ratio >= 0.5) {
            return { label: 'Sedang', toneClass: 'badge-blue', isAlarm: false, alarm: null, tooltip: '' };
          }
        }
      }
      return { label: 'Baik', toneClass: 'badge-green', isAlarm: false, alarm: null, tooltip: '' };
    },
    getBentoCardColor(param) {
      const dark = this.isDark;
      if (!param) return dark ? '#38BDF8' : '#0284C7';
      const badge = this.getBentoStatusBadge(param);
      if (badge.toneClass === 'badge-rose') {
        return '#F43F5E';
      }
      if (badge.toneClass === 'badge-amber') {
        return dark ? '#FBBF24' : '#D97706';
      }
      const plotted = (this.activeTrendSeries || []).find(s => Number(s.parameterId) === Number(param.param_id));
      if (plotted && plotted.color) {
        return plotted.color;
      }
      if (badge.toneClass === 'badge-blue') return dark ? '#38BDF8' : '#0284C7';
      if (badge.toneClass === 'badge-teal') return dark ? '#2DD4BF' : '#0D9488';
      const unitStr = (param.unit || '').trim().toLowerCase();
      if (unitStr === 'c' || unitStr === '°c') return dark ? '#FBBF24' : '#D97706';
      return dark ? '#34D399' : '#059669';
    },
    getBentoCardProgress(param) {
      if (!param) {
        return { pct: 0, metaLabel: '-', shortMetaLabel: '-', rightLabel: '', tableLimitLabel: '', isConfiguredLimit: false, limitTooltip: '' };
      }
      const editHint = this.canEditConfig ? 'Klik untuk mengatur limit parameter' : 'Klik untuk melihat konfigurasi limit';
      const rel = param.received_at ? this.formatRelative(param.received_at) : '';
      const activeInfo = this.getBentoParamActiveAlarm(param);
      const activeMetaSuffix = activeInfo
        ? (activeInfo.shortMeta + (activeInfo.ageStr ? ' ' + activeInfo.ageStr : ''))
        : rel;

      const limits = this.resolveBentoParamLimits(param);
      if (this.isBentoParamCommOffline(param)) {
        return {
          pct: 0,
          metaLabel: rel ? `Terakhir aktif · ${rel}` : 'Sensor tidak terhubung',
          shortMetaLabel: rel || 'Offline',
          rightLabel: 'Offline',
          tableLimitLabel: limits ? limits.shortLabel : '',
          isConfiguredLimit: Boolean(limits && limits.source === 'configured'),
          limitTooltip: `Status Offline${limits ? ' · ' + limits.detailNote : ''} · ${editHint}`,
        };
      }
      if (this.isDiscreteStatusParam(param)) {
        const info = this.getDiscreteStatusInfo(param);
        return {
          pct: info && info.isOn ? 100 : 18,
          metaLabel: rel ? `Status Kanal · ${rel}` : 'Status Kanal Aktif',
          shortMetaLabel: rel || 'Kanal Digital',
          rightLabel: info ? info.label : '',
          tableLimitLabel: '',
          isConfiguredLimit: false,
          limitTooltip: editHint,
        };
      }
      const val = Number(param.value);
      const hasVal = param.value !== null && param.value !== undefined && param.value !== '' && isFinite(val);

      if (limits && limits.upper !== null) {
        const base = limits.lower !== null ? limits.lower : 0;
        const span = limits.upper - base;
        const rawPct = hasVal && span > 0 ? Math.round(((val - base) / span) * 100) : 0;
        const barPct = hasVal ? Math.min(Math.max(rawPct, val > base ? 3 : 0), 100) : 0;
        const displayPct = hasVal ? Math.max(0, Math.min(rawPct, 999)) : 0;
        return {
          pct: barPct,
          metaLabel: activeMetaSuffix ? `${limits.longLabel} · ${activeMetaSuffix}` : limits.longLabel,
          shortMetaLabel: limits.shortLabel,
          rightLabel: hasVal ? `${displayPct}%` : '--',
          tableLimitLabel: limits.shortLabel,
          isConfiguredLimit: limits.source === 'configured',
          limitTooltip: `${limits.detailNote} (${limits.source === 'configured' ? 'Edit Config' : 'Default'}) · ${editHint}`,
        };
      }
      if (limits && limits.lower !== null) {
        return {
          pct: hasVal && val <= limits.lower ? 100 : 35,
          metaLabel: activeMetaSuffix ? `${limits.longLabel} · ${activeMetaSuffix}` : limits.longLabel,
          shortMetaLabel: limits.shortLabel,
          rightLabel: activeInfo ? activeInfo.shortMeta : (param.quality || 'GOOD'),
          tableLimitLabel: limits.shortLabel,
          isConfiguredLimit: true,
          limitTooltip: `${limits.detailNote} · ${editHint}`,
        };
      }
      return {
        pct: hasVal ? 35 : 0,
        metaLabel: activeMetaSuffix ? `Update · ${activeMetaSuffix}` : 'Online',
        shortMetaLabel: activeInfo ? activeInfo.shortMeta : (rel || 'Online'),
        rightLabel: activeInfo ? activeInfo.shortMeta : (param.quality || 'GOOD'),
        tableLimitLabel: '',
        isConfiguredLimit: false,
        limitTooltip: `Belum ada limit khusus · ${editHint}`,
      };
    },
    shouldFlattenBentoSparkline(param) {
      if (!param) return false;
      if (this.isBentoParamCommOffline(param)) return true;
      if (param.quality === 'BAD' && !param.is_held_value) return true;
      if (param.is_stale || param.quality === 'STALE') return true;
      if (param.value === null || param.value === undefined || param.value === '' || !isFinite(Number(param.value))) {
        return true;
      }
      const activeInfo = this.getBentoParamActiveAlarm(param);
      if (activeInfo && !this.isParamFormulaFallback(param)) {
        const c = String(activeInfo.code || '');
        if (
          c.includes('COMM') ||
          c.includes('OFFLINE') ||
          c.includes('DISCONNECT') ||
          c.includes('TIMEOUT') ||
          c.includes('BAD_QUALITY') ||
          c.includes('STALE') ||
          c.includes('STUCK') ||
          c.includes('ZERO')
        ) {
          return true;
        }
      }
      return false;
    },
    resolveBentoTrendMaxGapMs(timeSpan) {
      const pollStep = this.configuredPollingIntervalMs || 5000;
      const passiveGapMs = this.isPassiveOrCloudTelemetry ? (this.configuredStaleTimeoutSeconds || 300) * 1000 * 1.2 : 0;
      let baseGap = 4 * 3600 * 1000;
      if (pollStep <= 2500 && timeSpan <= 90 * 1000) baseGap = 6 * 1000;
      else if (pollStep <= 2500 && timeSpan <= 7 * 60 * 1000) baseGap = 10 * 1000;
      else if (timeSpan <= 90 * 1000) baseGap = 12 * 1000;
      else if (timeSpan <= 7 * 60 * 1000) baseGap = 20 * 1000;
      else if (timeSpan <= 20 * 60 * 1000) baseGap = 45 * 1000;
      else if (timeSpan <= 2 * 3600 * 1000) baseGap = 3 * 60 * 1000;
      else if (timeSpan <= 8 * 3600 * 1000) baseGap = 15 * 60 * 1000;
      else if (timeSpan <= 25 * 3600 * 1000) baseGap = 60 * 60 * 1000;
      return Math.max(baseGap, passiveGapMs);
    },
    buildBentoSparklineGeometry(rawPoints, options = {}) {
      const smooth = options.smooth !== undefined ? Boolean(options.smooth) : Boolean(this.trendLineSmooth);
      const W = 100;
      const H = 20;
      const topPad = 2.6;
      const bottomPad = 2.6;
      const plotH = H - topPad - bottomPad;

      if (!rawPoints || rawPoints.length === 0) {
        return {
          hasHistory: false,
          isFlat: false,
          validCount: 0,
          strokePaths: [],
          areaPaths: [],
          min: 0,
          max: 0,
          _smooth: smooth,
        };
      }

      if (rawPoints.length === 1) {
        const val = rawPoints[0].v;
        return {
          hasHistory: true,
          isFlat: false,
          validCount: 1,
          strokePaths: ['M 0,10 L 100,10'],
          areaPaths: ['M 0,20 L 0,10 L 100,10 L 100,20 Z'],
          min: Number(val.toFixed(2)),
          max: Number(val.toFixed(2)),
          _smooth: smooth,
        };
      }

      let min = Infinity;
      let max = -Infinity;
      const deltas = [];
      for (let i = 0; i < rawPoints.length; i++) {
        const pt = rawPoints[i];
        if (pt.v < min) min = pt.v;
        if (pt.v > max) max = pt.v;
        if (i > 0 && pt.t > rawPoints[i - 1].t) {
          deltas.push(pt.t - rawPoints[i - 1].t);
        }
      }

      let span = max - min;
      const isConstant = span < 1e-9;
      if (isConstant) {
        min -= 1;
        max += 1;
        span = 2;
      } else {
        const pad = Math.max(span * 0.18, 1e-4);
        min -= pad;
        max += pad;
        span = max - min;
      }

      let xStart = rawPoints[0].t;
      let xEnd = rawPoints[rawPoints.length - 1].t;
      if (options.windowStart && options.windowEnd && options.windowEnd > options.windowStart) {
        const winSpan = options.windowEnd - options.windowStart;
        const dataSpan = xEnd - xStart;
        if (dataSpan >= winSpan * 0.25) {
          xStart = options.windowStart;
          xEnd = options.windowEnd;
        }
      }
      const xSpan = xEnd - xStart;

      let medianDt = 0;
      if (deltas.length > 0) {
        deltas.sort((a, b) => a - b);
        medianDt = deltas[Math.floor(deltas.length / 2)];
      }
      const pollMs = this.configuredPollingIntervalMs || 5000;
      const passiveGapMs = this.isPassiveOrCloudTelemetry ? (this.configuredStaleTimeoutSeconds || 300) * 1000 * 1.2 : 0;
      const baseGapMs = Math.max(options.baseGapMs || 180000, passiveGapMs);
      const maxGapMs = Math.max(baseGapMs, pollMs * 6, medianDt * 3.5);

      const segments = [];
      let currentSeg = [];
      let prevT = 0;

      for (let i = 0; i < rawPoints.length; i++) {
        const pt = rawPoints[i];
        const x = xSpan > 0
          ? Math.min(Math.max(((pt.t - xStart) / xSpan) * W, 0), W)
          : (i / (rawPoints.length - 1)) * W;
        const norm = isConstant ? 0.5 : (pt.v - min) / span;
        const y = Math.min(Math.max((H - bottomPad) - norm * plotH, 1.2), H - 1.2);

        if (currentSeg.length > 0 && pt.t > 0 && prevT > 0 && (pt.t - prevT) > maxGapMs) {
          segments.push(currentSeg);
          currentSeg = [];
        }

        if (currentSeg.length > 0 && Math.abs(x - currentSeg[currentSeg.length - 1].x) < 0.22) {
          currentSeg[currentSeg.length - 1].y = y;
        } else {
          currentSeg.push({ x, y });
        }
        prevT = pt.t;
      }
      if (currentSeg.length > 0) {
        segments.push(currentSeg);
      }

      const strokePaths = [];
      const areaPaths = [];
      for (const seg of segments) {
        if (seg.length === 1) {
          const p = seg[0];
          const x1 = Math.max(0, Number((p.x - 1.2).toFixed(1)));
          const x2 = Math.min(W, Number((p.x + 1.2).toFixed(1)));
          const py = Number(p.y.toFixed(1));
          strokePaths.push(`M ${x1},${py} L ${x2},${py}`);
          areaPaths.push(`M ${x1},${H} L ${x1},${py} L ${x2},${py} L ${x2},${H} Z`);
        } else if (seg.length > 1) {
          const strokeD = this.buildSplinePath(seg, smooth);
          const areaD = this.buildSplineArea(seg, H, smooth);
          if (strokeD) strokePaths.push(strokeD);
          if (areaD) areaPaths.push(areaD);
        }
      }

      return {
        hasHistory: strokePaths.length > 0,
        isFlat: false,
        validCount: rawPoints.length,
        strokePaths,
        areaPaths,
        min: Number(min.toFixed(2)),
        max: Number(max.toFixed(2)),
        _smooth: smooth,
      };
    },
    getBentoSparkRenderCache() {
      if (!Object.prototype.hasOwnProperty.call(this, '_bentoSparkRenderCache') || !this._bentoSparkRenderCache) {
        Object.defineProperty(this, '_bentoSparkRenderCache', {
          value: Object.create(null),
          writable: true,
          configurable: true,
          enumerable: false,
        });
      }
      return this._bentoSparkRenderCache;
    },
    recomputeSparklineMeta(paramId) {
      // Never mutate reactive state here so template render getters stay 100% side-effect-free.
      const pid = Number(paramId);
      if (pid && this._bentoSparkRenderCache) {
        delete this._bentoSparkRenderCache[pid];
      }
    },
    getSparkline(paramId) {
      if (this.perfSettings && !this.perfSettings.parameterSparklinesEnabled) {
        return { hasHistory: false, isFlat: false, strokePaths: [], areaPaths: [] };
      }
      let param = null;
      let pid = null;
      if (typeof paramId === 'object' && paramId !== null) {
        param = paramId;
        pid = Number(param.param_id || param.id);
      } else if (paramId != null) {
        pid = Number(paramId);
        param = (this.allParameters || []).find(p => Number(p.param_id || p.id) === pid) || null;
      }
      if (!pid || isNaN(pid)) {
        return { hasHistory: false, isFlat: false, strokePaths: [], areaPaths: [] };
      }

      if (param && this.shouldFlattenBentoSparkline(param)) {
        return {
          hasHistory: true,
          isFlat: true,
          validCount: 2,
          strokePaths: ['M 0,10 L 100,10'],
          areaPaths: ['M 0,20 L 0,10 L 100,10 L 100,20 Z'],
          min: 0,
          max: 0,
        };
      }

      const smooth = Boolean(this.trendLineSmooth);
      const renderCache = this.getBentoSparkRenderCache();

      // 1. If parameter is actively plotted in Trend Analysis, mirror its visible series & spline 1:1
      const plottedSeries = (this.activeTrendSeries || []).find(
        s => Number(s.parameterId) === pid && Array.isArray(s.dataPoints) && s.dataPoints.length > 0
      );
      if (plottedSeries) {
        const sPts = plottedSeries.dataPoints;
        const bounds = this.trendTimeBounds || {};
        const tStart = Number(bounds.start || 0);
        const tEnd = Number(bounds.end || 0);
        const firstPt = sPts[0];
        const lastPt = sPts[sPts.length - 1];
        const pollBucket = Math.max(this.configuredPollingIntervalMs || 5000, 1000);
        const trendKey = `trend:${sPts.length}:${firstPt ? firstPt.x : 0}:${lastPt ? lastPt.x : 0}:${lastPt ? lastPt.y : 0}:${Math.floor(tStart / pollBucket)}:${Math.floor(tEnd / pollBucket)}:${smooth ? 1 : 0}`;
        const prevHit = renderCache[pid];
        if (prevHit && prevHit.key === trendKey) {
          return prevHit.val;
        }

        const winPts = [];
        const allValidPts = [];
        for (let i = 0; i < sPts.length; i++) {
          const p = sPts[i];
          if (!p || p.y === null || p.y === undefined || !isFinite(p.y) || p.quality === 'BAD') continue;
          const item = { t: Number(p.x || 0), v: Number(p.y) };
          allValidPts.push(item);
          if ((!tStart || item.t >= tStart) && (!tEnd || item.t <= tEnd)) {
            winPts.push(item);
          }
        }
        const chosenPts = winPts.length >= 2 ? winPts : allValidPts;
        if (chosenPts.length >= 2) {
          const spanMs = tEnd > tStart ? (tEnd - tStart) : Math.max(chosenPts[chosenPts.length - 1].t - chosenPts[0].t, 60000);
          const built = this.buildBentoSparklineGeometry(chosenPts, {
            windowStart: winPts.length >= 2 ? tStart : 0,
            windowEnd: winPts.length >= 2 ? tEnd : 0,
            baseGapMs: this.resolveBentoTrendMaxGapMs(spanMs),
            smooth,
          });
          renderCache[pid] = { key: trendKey, val: built };
          return built;
        }
      }

      // 2. Pure read of sparklinesMap[pid] (zero reactive mutations during template render!)
      const rawList = (this.sparklinesMap && this.sparklinesMap[pid]) || [];
      const rawLen = rawList.length;
      const firstRaw = rawLen > 0 ? rawList[0] : null;
      const lastRaw = rawLen > 0 ? rawList[rawLen - 1] : null;
      const liveVal = param && param.value !== null && param.value !== undefined && param.value !== '' && isFinite(Number(param.value))
        ? Number(param.value)
        : null;
      const sparkKey = `spark:${rawLen}:${firstRaw ? firstRaw.time : 0}:${lastRaw ? lastRaw.time : 0}:${lastRaw ? lastRaw.val : ''}:${lastRaw ? lastRaw.quality : ''}:${liveVal !== null ? liveVal : ''}:${smooth ? 1 : 0}`;
      const prevSparkHit = renderCache[pid];
      if (prevSparkHit && prevSparkHit.key === sparkKey) {
        return prevSparkHit.val;
      }

      const cleanPts = [];
      for (let i = 0; i < rawLen; i++) {
        const item = rawList[i];
        if (!item || item.val === null || item.val === undefined || !isFinite(item.val)) continue;
        if (item.quality === 'BAD') continue;
        cleanPts.push({
          t: Number(item.time || 0),
          v: Number(item.val),
        });
      }
      cleanPts.sort((a, b) => a.t - b.t);
      const deduped = [];
      for (let i = 0; i < cleanPts.length; i++) {
        if (deduped.length > 0 && cleanPts[i].t > 0 && Math.abs(cleanPts[i].t - deduped[deduped.length - 1].t) < 250) {
          deduped[deduped.length - 1] = cleanPts[i];
        } else {
          deduped.push(cleanPts[i]);
        }
      }

      let builtSpark;
      if (deduped.length > 0) {
        builtSpark = this.buildBentoSparklineGeometry(deduped, {
          baseGapMs: 180000,
          smooth,
        });
      } else if (liveVal !== null) {
        builtSpark = {
          hasHistory: true,
          isFlat: false,
          validCount: 1,
          strokePaths: ['M 0,10 L 100,10'],
          areaPaths: ['M 0,20 L 0,10 L 100,10 L 100,20 Z'],
          min: Number(liveVal.toFixed(2)),
          max: Number(liveVal.toFixed(2)),
          _smooth: smooth,
        };
      } else {
        builtSpark = { hasHistory: false, isFlat: false, strokePaths: [], areaPaths: [] };
      }

      renderCache[pid] = { key: sparkKey, val: builtSpark };
      return builtSpark;
    },
    onDeviceCommEvent(e) {
      const detail = e && e.detail;
      if (!detail) return;
      const type = detail.type;
      const data = detail.data || {};
      const devId = Number(data.device_id || data.id);
      const target = devId ? (this.devicesList || []).find(d => Number(d.id) === devId) : null;

      if (target && this.deviceHasConfiguredFormula(target)) {
        if (type === 'device.communication.error') {
          return;
        }
        if (type === 'device.connection.changed') {
          const status = data.connection_status || data.status;
          if (status === 'OFFLINE' || status === 'ERROR' || status === 'DISCONNECTED') {
            this.$set(target, 'connection_status', 'ONLINE');
            this.$set(target, 'status', 'ONLINE');
            this.staleCheckTimeMs = Date.now();
            (target.parameters || []).forEach((p) => {
              if (String(p.formula || '').trim() !== '') return;
              const prev = this.liveMap[p.id] || {};
              this.$set(this.liveMap, p.id, {
                ...prev,
                parameter_id: p.id,
                quality: 'BAD',
                quality_reason: data.reason || 'PORT_DISCONNECTED',
              });
            });
            return;
          }
        }
      }

      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.onDeviceCommEvent) {
        OperatorCommandCenterView.methods.onDeviceCommEvent.call(this, e);
      }
    },
    onTelemetryEvent(e) {
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.onTelemetryEvent) {
        OperatorCommandCenterView.methods.onTelemetryEvent.call(this, e);
      }
      const data = e && e.detail;
      if (data && data.device_id && data.quality === 'BAD') {
        const target = (this.devicesList || []).find(d => Number(d.id) === Number(data.device_id));
        if (target && this.deviceHasConfiguredFormula(target)) {
          this.$set(target, 'connection_status', 'ONLINE');
          this.$set(target, 'status', 'ONLINE');
        }
      }
      if (
        this.canManageCategories &&
        this.gpsModbusMapping &&
        this.gpsModbusMapping.autoSync &&
        !this.isSavingStationLocation &&
        this.modbusRtuGpsStatus &&
        this.modbusRtuGpsStatus.state === 'LOCKED'
      ) {
        const now = Date.now();
        if (now - (this.lastModbusAutoSyncMs || 0) >= 60000) {
          const nextLat = this.modbusRtuGpsStatus.lat;
          const nextLng = this.modbusRtuGpsStatus.lng;
          const prevLat = this.stationLocation && this.stationLocation.latitude !== null ? Number(this.stationLocation.latitude) : null;
          const prevLng = this.stationLocation && this.stationLocation.longitude !== null ? Number(this.stationLocation.longitude) : null;
          let movedMeters = Infinity;
          if (prevLat !== null && prevLng !== null && Number.isFinite(prevLat) && Number.isFinite(prevLng)) {
            const rad = Math.PI / 180;
            const dLat = (nextLat - prevLat) * 111320;
            const dLng = (nextLng - prevLng) * 111320 * Math.cos(((nextLat + prevLat) / 2) * rad);
            movedMeters = Math.sqrt(dLat * dLat + dLng * dLng);
          }
          if (!this.stationLocation.is_configured || movedMeters >= 25) {
            this.lastModbusAutoSyncMs = now;
            this.applyModbusGpsToCandidate(true);
          }
        }
      }
    },
    formatLocationSourceLabel(src) {
      const norm = String(src || '').trim().toUpperCase();
      if (norm === 'MODBUS_RTU') return 'GPS Modbus RTU';
      if (norm === 'IP_GEOLOCATION') return 'Internet IP';
      if (norm === 'MANUAL') return 'Manual';
      if (norm === 'GPS') return 'GPS Sensor';
      return 'Unconfigured';
    },
    formatModbusGpsParamOption(param) {
      if (!param) return '';
      const namePart = param.param_name && param.param_code && param.param_name !== param.param_code
        ? `${param.param_name} (${param.param_code})`
        : (param.param_name || param.param_code || `Param #${param.param_id}`);
      if (this.canManageCategories) {
        const devLabel = param.device_name || param.device_code || 'Modbus RTU';
        return `${devLabel} · ${namePart}`;
      }
      return namePart;
    },
    formatModbusGpsDeviceOption(dev) {
      if (!dev) return '';
      const code = dev.device_code || dev.code || '';
      const name = dev.device_name || dev.name || '';
      const proto = dev.protocol || dev.communication_protocol || 'MODBUS_RTU';
      if (name && code && name !== code) {
        return `${name} (${code} · ${proto})`;
      }
      return `${name || code || 'Device'} (${proto})`;
    },
    onModbusGpsDeviceChange() {
      this.gpsModbusMapping = {
        ...this.gpsModbusMapping,
        latParamId: '',
        lngParamId: '',
        satParamId: '',
      };
      this.saveModbusGpsMapping();
    },
    openDevicesPageForGps() {
      this.closeToolbarPopover();
      this.navigateToRoute('/monitoring/devices');
    },
    openAddGpsParamModal() {
      if (!this.canManageCategories) return;
      const selectedDevId = this.gpsModbusMapping && this.gpsModbusMapping.deviceId;
      const targetDev = (selectedDevId && (this.devicesList || []).find(d => String(d.id) === String(selectedDevId)))
        || (this.modbusRtuDevicesList && this.modbusRtuDevicesList[0])
        || (this.devicesList && this.devicesList[0])
        || null;
      if (!targetDev) {
        this.openDevicesPageForGps();
        return;
      }
      this.closeToolbarPopover();
      this.editParamTarget = {
        deviceId: targetDev.id,
        device: targetDev,
        parameter: null,
      };
      this.isEditParamModalOpen = true;
    },
    openEditGpsParamModal(param) {
      if (!param || !this.canManageCategories) return;
      this.closeToolbarPopover();
      this.openBentoQuickLimitEdit(param);
    },
    saveModbusGpsMapping() {
      try {
        localStorage.setItem('datalogger_gps_modbus_rtu_mapping', JSON.stringify(this.gpsModbusMapping));
      } catch (e) {}
    },
    toggleModbusGpsAutoSync() {
      this.gpsModbusMapping = {
        ...this.gpsModbusMapping,
        autoSync: !this.gpsModbusMapping.autoSync,
      };
      this.saveModbusGpsMapping();
    },
    async applyModbusGpsToCandidate(saveImmediately = false) {
      if (!this.modbusRtuGpsStatus || this.modbusRtuGpsStatus.state !== 'LOCKED') {
        this.stationLocationFeedback = {
          type: 'error',
          message: (this.modbusRtuGpsStatus && this.modbusRtuGpsStatus.reason) || 'Data GPS Modbus RTU belum terkunci / tidak valid.',
        };
        return;
      }
      const lat = this.modbusRtuGpsStatus.lat;
      const lng = this.modbusRtuGpsStatus.lng;
      this.candidateLocation.latitude = lat;
      this.candidateLocation.longitude = lng;
      this.candidateLocation.location_source = 'MODBUS_RTU';
      this.candidateLocation.accuracy_meters = null;

      try {
        const geo = await reverseGeocodePlaceName(lat, lng);
        if (geo && (geo.displayName || geo.placeName)) {
          this.candidateLocation.address = geo.displayName || geo.placeName;
          const cleanPlace = extractConcisePlaceName(geo.placeName || geo.displayName);
          if (cleanPlace) {
            this.resolvedGpsPlaceName = cleanPlace;
            this.modalResolvedPlaceName = cleanPlace;
            try {
              localStorage.setItem('datalogger_cached_gps_place_name', cleanPlace);
            } catch (e) {}
          }
        }
      } catch (e) {}

      this.$nextTick(() => {
        if (this.$refs.miniMapViewer) {
          this.$refs.miniMapViewer.setMarkerPosition(lat, lng);
          this.$refs.miniMapViewer.recenterMap();
        }
        if (this.$refs.fullScreenMapViewer) {
          this.$refs.fullScreenMapViewer.setMarkerPosition(lat, lng);
          this.$refs.fullScreenMapViewer.recenterMap();
        }
      });

      if (saveImmediately) {
        await this.saveStationLocation();
      } else {
        this.stationLocationFeedback = {
          type: 'success',
          message: `Pratinjau koordinat GPS Modbus RTU (${lat.toFixed(6)}, ${lng.toFixed(6)}) diterapkan pada peta.`,
        };
      }
    },
    async onMapCoordinatePicked({ lat, lng }) {
      if (OperatorCommandCenterView.methods && OperatorCommandCenterView.methods.onMapCoordinatePicked) {
        OperatorCommandCenterView.methods.onMapCoordinatePicked.call(this, { lat, lng });
      }
      try {
        const geo = await reverseGeocodePlaceName(lat, lng);
        if (geo && (geo.displayName || geo.placeName)) {
          this.candidateLocation.address = geo.displayName || geo.placeName;
        }
      } catch (e) {}
    },
    onTestReadCompleted(payload) {
      if (!payload || !payload.result || !payload.parameter) return;
      const hasFormulaVal = payload.result.formula_value !== undefined && payload.result.formula_value !== null;
      if (payload.result.success || hasFormulaVal) {
        const val = hasFormulaVal ? Number(payload.result.formula_value) : payload.result.decoded_value;
        if (val !== undefined && val !== null) {
          const pid = payload.parameter.id || payload.parameter.param_id;
          if (pid) {
            this.$set(this.liveMap, pid, {
              ...(this.liveMap[pid] || {}),
              value: val,
              processed_value: val,
              formula_value: hasFormulaVal ? val : payload.result.formula_value,
              quality: 'GOOD',
              quality_reason: payload.result.success ? 'TEST_READ_PASS' : 'FORMULA_FALLBACK',
              quality_flags: payload.result.success ? '' : 'formula_fallback',
              received_at: new Date().toISOString(),
            });
          }
        }
      }
    },
    getShortParamChipLabel(param) {
      if (!param) return '';
      const raw = (param.param_code || param.param_name || '').trim();
      const upper = raw.toUpperCase();
      if (upper.includes('PM25') || upper.includes('PM2.5') || upper.includes('PM 2.5')) return 'PM2.5';
      if (upper.includes('PM10') || upper.includes('PM 10')) return 'PM10';
      if (upper.includes('SO2') || upper.includes('SULFUR')) return 'SO₂';
      if (upper.includes('NO2') || upper.includes('NITROGEN')) return 'NO₂';
      if (upper.includes('O3') || upper.includes('OZON')) return 'O₃';
      if (upper.includes('KARBON') || upper === 'CO') return 'CO';
      if (upper.includes('HIDROKARBON') || upper === 'HC') return 'HC';
      if (raw.length > 12) return raw.slice(0, 12);
      return raw;
    },
  },
};
</script>

<style scoped>
.bento-dashboard-page {
  --b-bg: radial-gradient(1200px 520px at 14% -10%, rgba(14, 165, 233, 0.07), transparent 60%),
          radial-gradient(900px 460px at 88% 0%, rgba(16, 185, 129, 0.05), transparent 55%),
          #EEF3F8;
  --b-card: #FFFFFF;
  --b-subtle: #F8FAFC;
  --b-pill: #F1F5F9;
  --b-border: #E2E8F0;
  --b-border-soft: #EDF2F7;
  --b-shadow: 0 1px 2px rgba(15, 23, 42, 0.03), 0 6px 18px rgba(15, 23, 42, 0.03);
  --b-text-main: #0F172A;
  --b-text-sec: #334155;
  --b-text-muted: #64748B;
  --b-text-faint: #94A3B8;
  --b-track: #E9EEF5;
  --b-active-btn: #FFFFFF;
  --b-primary: #0284C7;
  --b-accent-blue: #0284C7;
  --b-tooltip-bg: rgba(255, 255, 255, 0.97);

  min-height: calc(100vh - 2px);
  margin: -4px -6px 0;
  padding: 9px 14px 42px;
  background: var(--b-bg);
  color: var(--b-text-main);
  font-family: 'Plus Jakarta Sans', 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  -webkit-font-smoothing: antialiased;
  font-variant-numeric: tabular-nums;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.bento-dashboard-page.is-compact-vp {
  padding: 7px 10px 40px;
  gap: 7px;
}

.bento-dashboard-page.is-night {
  --b-bg: radial-gradient(1100px 520px at 12% -10%, rgba(14, 165, 233, 0.11), transparent 60%),
          radial-gradient(900px 460px at 88% 0%, rgba(16, 185, 129, 0.07), transparent 55%),
          #090F1C;
  --b-card: #111A2C;
  --b-subtle: #152036;
  --b-pill: #172338;
  --b-border: rgba(148, 163, 184, 0.15);
  --b-border-soft: rgba(148, 163, 184, 0.08);
  --b-shadow: 0 8px 24px rgba(0, 0, 0, 0.28);
  --b-text-main: #F8FAFC;
  --b-text-sec: #CBD5E1;
  --b-text-muted: #94A3B8;
  --b-text-faint: #64748B;
  --b-track: #1B2942;
  --b-active-btn: #243450;
  --b-primary: #0284C7;
  --b-accent-blue: #38BDF8;
  --b-tooltip-bg: rgba(15, 23, 42, 0.96);
}

/* ========================================================================= */
/* 1. HEADER SHELL + ICON-ONLY TOOLBAR                                       */
/* ========================================================================= */
.header-shell {
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 13px;
  box-shadow: var(--b-shadow);
  padding: 7px 13px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  position: relative;
  z-index: 30;
}

.is-compact-vp .header-shell {
  padding: 6px 11px;
  gap: 5px;
}

.h-row1 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  min-height: 32px;
  position: relative;
}

.brand-left {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1 1 auto;
}

.logo-badge {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: linear-gradient(135deg, #0EA5E9, #0D9488);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #FFFFFF;
  box-shadow: 0 3px 10px rgba(14, 165, 233, 0.22);
  flex-shrink: 0;
}

.logo-img-wrap {
  height: 32px;
  padding: 2px 6px;
  border-radius: 8px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.station-name {
  font-size: 13.5px;
  font-weight: 800;
  color: var(--b-text-main);
  letter-spacing: -0.015em;
  display: flex;
  align-items: center;
  gap: 6px;
  line-height: 1.2;
}
.is-compact-vp .station-name {
  font-size: 12.8px;
}

.cbi-chip {
  font-size: 9.5px;
  font-weight: 700;
  padding: 1px 5.5px;
  border-radius: 5px;
  color: var(--b-accent-blue);
  background: rgba(14, 165, 233, 0.14);
  flex-shrink: 0;
}

.station-loc {
  font-size: 11px;
  color: var(--b-text-muted);
  font-weight: 500;
  margin-top: 1px;
}
.station-loc strong {
  color: var(--b-text-sec);
  font-weight: 700;
}

.clock-center {
  display: flex;
  align-items: center;
  gap: 7px;
  flex-shrink: 0;
}
.clock-center.is-wide-center {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
}

.online-pill {
  padding: 4px 10px;
  border-radius: 999px;
  font-size: 10.5px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 5.5px;
  white-space: nowrap;
  transition: opacity 0.15s ease;
}
.online-pill:hover { opacity: 0.9; }

.is-day .online-pill.pill-green { background: #ECFDF5; color: #047857; border: 1px solid #D1FAE5; }
.is-night .online-pill.pill-green { background: rgba(16, 185, 129, 0.14); color: #6EE7B7; border: 1px solid rgba(16, 185, 129, 0.26); }

.is-day .online-pill.pill-amber { background: #FFFBEB; color: #B45309; border: 1px solid #FDE68A; }
.is-night .online-pill.pill-amber { background: rgba(245, 158, 11, 0.15); color: #FCD34D; border: 1px solid rgba(245, 158, 11, 0.28); }

.is-day .online-pill.pill-rose { background: #FFF1F2; color: #BE123C; border: 1px solid #FECDD3; }
.is-night .online-pill.pill-rose { background: rgba(244, 63, 94, 0.15); color: #FDA4AF; border: 1px solid rgba(244, 63, 94, 0.30); }

.clock-pill {
  padding: 4px 11px;
  border-radius: 8px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border-soft);
  font-size: 11.5px;
  font-weight: 600;
  color: var(--b-text-sec);
  display: inline-flex;
  align-items: center;
  gap: 6px;
  white-space: nowrap;
}
.clock-pill strong {
  color: var(--b-text-main);
  font-weight: 700;
}

.brand-right {
  display: flex;
  align-items: center;
  gap: 9px;
  flex-shrink: 0;
}
.maker-info { text-align: right; }
.maker-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--b-text-main);
  line-height: 1.2;
}
.maker-sub {
  font-size: 10px;
  color: var(--b-text-muted);
  margin-top: 1px;
}
.maker-logo {
  width: 30px;
  height: 30px;
  border-radius: 8px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--b-accent-blue);
}

.h-divider {
  height: 1px;
  background: var(--b-border-soft);
}

.h-row2 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  min-height: 30px;
  gap: 6px;
}

.tool-cluster {
  display: flex;
  align-items: center;
  gap: 5px;
}

.op-v-div {
  width: 1px;
  height: 16px;
  background: var(--b-border);
  margin: 0 2px;
  flex-shrink: 0;
}

/* Exact Operator-Console-Style Icon-Only Button */
.op-icon-btn {
  width: 32px;
  height: 30px;
  border-radius: 8px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border);
  color: var(--b-text-sec);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  position: relative;
  cursor: pointer;
  transition: all 0.15s ease;
  flex-shrink: 0;
}
.is-compact-vp .op-icon-btn {
  width: 30px;
  height: 28px;
}
.op-icon-btn.sm {
  width: 28px;
  height: 27px;
  border-radius: 7px;
}
.op-icon-btn.xs,
.is-compact-vp .op-icon-btn.xs {
  width: 27px;
  height: 26px;
  border-radius: 7px;
}
.op-icon-btn:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}
.op-icon-btn.is-active {
  background: rgba(14, 165, 233, 0.15);
  border-color: var(--b-accent-blue);
  color: var(--b-accent-blue);
}
.op-icon-btn.is-cyan-soft {
  background: rgba(14, 165, 233, 0.09);
  border-color: rgba(14, 165, 233, 0.36);
  color: var(--b-accent-blue);
}
.op-icon-btn.is-emerald {
  background: rgba(16, 185, 129, 0.14);
  border-color: rgba(16, 185, 129, 0.45);
  color: #10B981;
}
.op-icon-btn.is-amber {
  background: rgba(245, 158, 11, 0.15);
  border-color: rgba(245, 158, 11, 0.45);
  color: #F59E0B;
}
.op-icon-btn.is-rose {
  background: rgba(244, 63, 94, 0.15);
  border-color: rgba(244, 63, 94, 0.50);
  color: #F43F5E;
}

/* Corner Badges & Status Dots on Icon Buttons */
.op-corner-pill {
  position: absolute;
  top: -5px;
  right: -5px;
  min-width: 15px;
  height: 14px;
  padding: 0 3.5px;
  border-radius: 999px;
  font-size: 8px;
  font-weight: 800;
  line-height: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1.5px solid var(--b-card);
  pointer-events: none;
}
.op-corner-pill.tone-Density {
  background: var(--b-pill);
  color: var(--b-accent-blue);
  border-color: var(--b-border);
}
.op-corner-pill.tone-emerald {
  background: #10B981;
  color: #FFFFFF;
}
.op-corner-pill.tone-amber {
  background: #F59E0B;
  color: #0F172A;
}
.op-corner-pill.tone-rose {
  background: #F43F5E;
  color: #FFFFFF;
}

.op-corner-dot {
  position: absolute;
  top: -2px;
  right: -2px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  border: 1.5px solid var(--b-card);
  pointer-events: none;
}
.dot-emerald { background: #10B981; }
.dot-amber { background: #F59E0B; }
.dot-rose { background: #F43F5E; }
.dot-sky { background: #38BDF8; }

.status-dot {
  width: 6.5px;
  height: 6.5px;
  border-radius: 50%;
  flex-shrink: 0;
}

.capsule {
  display: inline-flex;
  align-items: center;
  background: var(--b-pill);
  border: 1px solid var(--b-border);
  border-radius: 8px;
  padding: 2px;
  gap: 2px;
}

.cap-btn {
  height: 23px;
  padding: 0 8px;
  border-radius: 6px;
  font-size: 10.5px;
  font-weight: 600;
  color: var(--b-text-muted);
  display: inline-flex;
  align-items: center;
  gap: 4.5px;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.15s ease;
}
.cap-btn:hover { color: var(--b-text-main); }
.cap-btn.active {
  background: var(--b-active-btn);
  color: var(--b-text-main);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
}

/* ========================================================================= */
/* 2. AUTO-ADAPTIVE MIDDLE BENTO                                             */
/* ========================================================================= */
.bento-mid {
  display: grid;
  grid-template-columns: 300px minmax(0, 1fr);
  gap: 10px;
  align-items: stretch;
}
.bento-mid.bento-mid-single {
  grid-template-columns: minmax(0, 1fr);
}
.bento-mid.is-compact-mid {
  gap: 7px;
}
.bento-mid.is-trend-hidden {
  align-items: start;
}

.right-params-wrap {
  display: flex;
  flex-direction: column;
  gap: 7px;
  min-width: 0;
}

/* Scrollbar Utilities */
.no-scrollbar::-webkit-scrollbar {
  display: none;
}
.no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.custom-scroller::-webkit-scrollbar {
  width: 5px;
  height: 5px;
}
.custom-scroller::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scroller::-webkit-scrollbar-thumb {
  background: rgba(148, 163, 184, 0.35);
  border-radius: 999px;
}
.custom-scroller::-webkit-scrollbar-thumb:hover {
  background: rgba(148, 163, 184, 0.55);
}

/* Compact Parameter Icon Control Strip (Zero Horizontal Scrollbar) */
.param-ctrl-bar {
  min-height: 26px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  position: relative;
  z-index: 25;
}
.param-ctrl-bar.has-open-popover {
  z-index: 50;
}

.param-ctrl-left {
  display: flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
}

.param-ctrl-right {
  display: flex;
  align-items: center;
  gap: 5px;
}

.param-count-tag {
  font-size: 10.5px;
  font-weight: 600;
  color: var(--b-text-muted);
  display: inline-flex;
  align-items: center;
  gap: 3px;
  padding: 0 3px;
  white-space: nowrap;
}
.param-count-tag strong {
  font-weight: 800;
  color: var(--b-text-sec);
}

.active-filter-chip {
  height: 22px;
  padding: 0 7.5px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: rgba(14, 165, 233, 0.12);
  border: 1px solid rgba(14, 165, 233, 0.35);
  color: var(--b-accent-blue);
  cursor: pointer;
  transition: all 0.15s ease;
  min-width: 0;
}
.active-filter-chip.amber {
  background: rgba(245, 158, 11, 0.14);
  border-color: rgba(245, 158, 11, 0.40);
  color: #D97706;
}
.is-night .active-filter-chip.amber {
  color: #FCD34D;
}
.afc-x {
  font-size: 12px;
  line-height: 1;
  opacity: 0.75;
}
.active-filter-chip:hover .afc-x {
  opacity: 1;
}

/* Segmented Icon-Only Capsule */
.icon-seg-capsule {
  display: inline-flex;
  align-items: center;
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 7px;
  padding: 2px;
  gap: 2px;
}

.isc-btn {
  width: 24px;
  height: 22px;
  border-radius: 5px;
  color: var(--b-text-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
}
.isc-btn:hover {
  color: var(--b-text-main);
  background: var(--b-subtle);
}
.isc-btn.active {
  background: var(--b-primary);
  color: #FFFFFF;
}

.isc-text-btn {
  height: 22px;
  padding: 0 6px;
  border-radius: 5px;
  font-size: 9.5px;
  font-weight: 800;
  letter-spacing: 0.03em;
  color: var(--b-text-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
}
.isc-text-btn:hover {
  color: var(--b-text-main);
  background: var(--b-subtle);
}
.isc-text-btn.active {
  background: rgba(245, 158, 11, 0.16);
  color: #F59E0B;
}

/* Shared Bento Mini Popovers (Filter, Card/Table Setup, Trend Filter, Trend Height) */
.bento-mini-popover {
  position: absolute;
  top: calc(100% + 6px);
  z-index: 60;
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 11px;
  padding: 9px 10px;
  box-shadow: 0 12px 28px rgba(15, 23, 42, 0.18);
  color: var(--b-text-main);
}
.is-night .bento-mini-popover {
  box-shadow: 0 14px 32px rgba(0, 0, 0, 0.50);
}
.bento-mini-popover.align-left {
  left: 0;
}
.bento-mini-popover.align-right {
  right: 0;
}

.bmp-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  padding-bottom: 6px;
  margin-bottom: 6px;
  border-bottom: 1px solid var(--b-border-soft);
}
.bmp-title {
  font-size: 11px;
  font-weight: 800;
  color: var(--b-text-main);
}
.bmp-badge {
  font-size: 9.5px;
  font-weight: 700;
  padding: 1.5px 6px;
  border-radius: 999px;
  background: var(--b-pill);
  color: var(--b-accent-blue);
}
.bmp-link-btn {
  font-size: 10px;
  font-weight: 700;
  color: var(--b-accent-blue);
  padding: 1px 5px;
  border-radius: 4px;
  cursor: pointer;
}
.bmp-link-btn:hover {
  background: var(--b-pill);
}

.bmp-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 224px;
  overflow-y: auto;
}
.bmp-item {
  width: 100%;
  padding: 5px 7.5px;
  border-radius: 7px;
  font-size: 11px;
  font-weight: 600;
  color: var(--b-text-sec);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  cursor: pointer;
  transition: all 0.12s ease;
  text-align: left;
}
.bmp-item:hover {
  background: var(--b-subtle);
  color: var(--b-text-main);
}
.bmp-item.active {
  background: rgba(14, 165, 233, 0.12);
  color: var(--b-accent-blue);
  font-weight: 700;
}
.bmp-item-left {
  display: flex;
  align-items: center;
  gap: 6px;
}
.bmp-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}
.bmp-count {
  font-size: 9.5px;
  font-weight: 700;
  padding: 1px 5.5px;
  border-radius: 999px;
  background: var(--b-pill);
  color: var(--b-text-muted);
  flex-shrink: 0;
}
.bmp-item.active .bmp-count {
  background: var(--b-primary);
  color: #FFFFFF;
}

.bmp-divider {
  height: 1px;
  background: var(--b-border-soft);
  margin: 5px 0;
}

.bmp-section {
  margin-bottom: 7px;
}
.bmp-section:last-child {
  margin-bottom: 0;
}
.bmp-label {
  font-size: 9.8px;
  font-weight: 700;
  color: var(--b-text-muted);
  margin-bottom: 4px;
}
.bmp-seg-grid {
  display: grid;
  gap: 3.5px;
}
.bmp-seg-grid.cols-2 {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.bmp-seg-grid.cols-3 {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}
.bmp-seg-grid.cols-4 {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}
.bmp-seg-grid.cols-5 {
  grid-template-columns: repeat(5, minmax(0, 1fr));
}
.bmp-seg-grid.cols-6 {
  grid-template-columns: repeat(6, minmax(0, 1fr));
}
.bmp-seg-grid.cols-7 {
  grid-template-columns: repeat(7, minmax(0, 1fr));
}
.bmp-seg-btn {
  height: 24px;
  border-radius: 6px;
  font-size: 10px;
  font-weight: 700;
  background: var(--b-pill);
  border: 1px solid var(--b-border);
  color: var(--b-text-sec);
  cursor: pointer;
  transition: all 0.12s ease;
}
.bmp-seg-btn:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}
.bmp-seg-btn.active {
  background: var(--b-primary);
  border-color: var(--b-primary);
  color: #FFFFFF;
}

.bmp-toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 10.5px;
  font-weight: 600;
  color: var(--b-text-sec);
  cursor: pointer;
  padding: 2px 0;
}
.bmp-toggle-row input[type='checkbox'] {
  accent-color: #0284C7;
  cursor: pointer;
}
.bmp-hint {
  font-size: 9.5px;
  line-height: 1.35;
  color: var(--b-text-muted);
  margin-top: 4px;
}

/* ========================================================================= */
/* 2A. AUTO-ADAPTIVE PARAMETER CARDS (1P .. 5P + MICRO DENSITY)              */
/* ========================================================================= */
.sensors-scroll-area {
  max-height: 332px;
  overflow-y: auto;
  padding-right: 2px;
}
.sensors-scroll-area.is-compact-scroll {
  max-height: 268px;
}
.sensors-scroll-area.is-trend-hidden {
  max-height: none !important;
  overflow-y: visible !important;
  padding-right: 0;
}

.sensors-grid {
  display: grid;
  gap: 8px;
}
.sensors-grid.is-compact-cards {
  gap: 6px;
}
.sensors-grid.is-inside-cat-group {
  padding: 5px 6px 6px;
  gap: 5px;
}
.sensors-grid.is-compact-cards.is-inside-cat-group {
  padding: 4.5px 5.5px 5.5px;
  gap: 4.5px;
}

/* Category-Separated Card Group Containers (Proportional Multi-Category Bento Grid) */
.bento-cat-cards-stack {
  display: grid;
  grid-template-columns: repeat(var(--cat-total-cols, 8), minmax(0, 1fr));
  gap: 6px;
  align-items: start;
}
.bento-cat-card-group {
  min-width: 0;
  border: 1px solid var(--b-border);
  border-left: 3px solid #06B6D4;
  border-radius: 10px;
  overflow: hidden;
  transition: border-color 0.15s ease;
}
.is-day .bento-cat-card-group {
  background: rgba(248, 250, 252, 0.88);
}
.is-night .bento-cat-card-group {
  background: rgba(15, 23, 42, 0.44);
}
.bento-cat-card-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 5px;
  padding: 3.5px 7.5px;
  min-height: 23px;
  cursor: pointer;
  user-select: none;
  border-bottom: 1px solid var(--b-border-soft);
  transition: background 0.12s ease;
}
.bento-cat-card-head.is-collapsed-head {
  border-bottom: none;
}
.is-day .bento-cat-card-head {
  background: rgba(241, 245, 249, 0.92);
}
.is-night .bento-cat-card-head {
  background: rgba(30, 41, 59, 0.5);
}
.bento-cat-card-head:hover {
  filter: brightness(0.98);
}
.bento-cat-card-head-left {
  display: flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
}
.bento-cat-card-title {
  font-size: 9.4px;
  font-weight: 800;
  letter-spacing: 0.035em;
  text-transform: uppercase;
  color: var(--b-text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.bento-cat-card-head-right {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-shrink: 0;
}
.bento-cat-card-head-right .s-badge {
  font-size: 8px;
  padding: 0.5px 4.5px;
}
.bento-cat-count-pill {
  font-size: 8.6px;
  font-weight: 700;
  padding: 0.5px 5px;
  border-radius: 999px;
  background: var(--b-pill);
  color: var(--b-text-muted);
  border: 1px solid var(--b-border-soft);
}
.bento-cat-chevron {
  color: var(--b-text-muted);
  transition: transform 0.16s ease;
  flex-shrink: 0;
}
.bento-cat-chevron.is-collapsed {
  transform: rotate(-90deg);
}

.bento-cols-1 { grid-template-columns: minmax(0, 1fr); }
.bento-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.bento-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.bento-cols-4 { grid-template-columns: repeat(4, minmax(0, 1fr)); }
.bento-cols-5 { grid-template-columns: repeat(5, minmax(0, 1fr)); }
.bento-cols-6 { grid-template-columns: repeat(6, minmax(0, 1fr)); }
.bento-cols-7 { grid-template-columns: repeat(7, minmax(0, 1fr)); }
.bento-cols-8 { grid-template-columns: repeat(8, minmax(0, 1fr)); }
.bento-cols-9 { grid-template-columns: repeat(9, minmax(0, 1fr)); }

.s-card {
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 12px;
  box-shadow: var(--b-shadow);
  padding: 10px 12px;
  min-height: 118px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 6px;
  cursor: pointer;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

/* Dense 5-column cards on Widescreen (>=1280px) so 15 params fit cleanly in 3x5 */
.sensors-grid.is-micro-cards .s-card {
  padding: 8px 10px;
  min-height: 94px;
  border-radius: 11px;
  gap: 5px;
}

/* Ultra-tidy Micro Cards on Compact Screens (<1280px) */
.sensors-grid.is-compact-cards .s-card {
  padding: 6px 8.5px;
  min-height: 76px;
  border-radius: 10px;
  gap: 4px;
}

.s-card:hover {
  border-color: var(--b-accent-blue);
}
.is-day .s-card-selected {
  border-color: #38BDF8;
  box-shadow: 0 2px 10px rgba(2, 132, 199, 0.10);
}
.is-night .s-card-selected {
  border-color: rgba(56, 189, 248, 0.50);
}
.is-day .s-card-fault {
  border-color: rgba(244, 63, 94, 0.65) !important;
  background: linear-gradient(180deg, rgba(255, 228, 230, 0.36), #FFFFFF);
}
.is-night .s-card-fault {
  border-color: rgba(244, 63, 94, 0.58) !important;
  background: linear-gradient(180deg, rgba(244, 63, 94, 0.08), var(--b-card));
}
.is-day .s-card-warn {
  border-color: rgba(245, 158, 11, 0.68) !important;
  background: linear-gradient(180deg, rgba(254, 243, 199, 0.44), #FFFFFF);
}
.is-night .s-card-warn {
  border-color: rgba(245, 158, 11, 0.58) !important;
  background: linear-gradient(180deg, rgba(245, 158, 11, 0.08), var(--b-card));
}
.s-card-blink {
  box-shadow: 0 0 0 1px rgba(56, 189, 248, 0.35);
}

.s-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
}

.s-title-wrap {
  display: flex;
  align-items: center;
  gap: 5px;
  min-width: 0;
}

.s-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  flex-shrink: 0;
}

.s-name {
  font-size: 11.5px;
  font-weight: 700;
  color: var(--b-text-sec);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.sensors-grid.is-micro-cards .s-name { font-size: 11px; }
.sensors-grid.is-compact-cards .s-name { font-size: 10.4px; }

.s-head-right {
  display: flex;
  align-items: center;
  gap: 2.5px;
  flex-shrink: 0;
}

.s-cat {
  font-size: 8.8px;
  font-weight: 700;
  padding: 1px 4.5px;
  border-radius: 4px;
  background: var(--b-pill);
  color: var(--b-text-muted);
}

.s-badge {
  padding: 1.5px 6px;
  border-radius: 999px;
  font-size: 9px;
  font-weight: 700;
  white-space: nowrap;
}
.s-badge.is-clickable-alarm {
  cursor: pointer;
  border: 1px solid currentColor;
  transition: transform 0.12s ease, filter 0.12s ease;
}
.s-badge.is-clickable-alarm:hover {
  filter: brightness(1.08);
  transform: translateY(-0.5px);
}
.sensors-grid.is-compact-cards .s-badge {
  padding: 1px 5px;
  font-size: 8.5px;
}

.is-day .badge-green { background: #DCFCE7; color: #15803D; }
.is-night .badge-green { background: rgba(52, 211, 153, 0.15); color: #6EE7B7; }
.is-day .badge-blue { background: #E0F2FE; color: #0369A1; }
.is-night .badge-blue { background: rgba(56, 189, 248, 0.15); color: #7DD3FC; }
.is-day .badge-teal { background: #CCFBF1; color: #0F766E; }
.is-night .badge-teal { background: rgba(45, 212, 191, 0.15); color: #5EEAD4; }
.is-day .badge-amber { background: #FEF3C7; color: #B45309; }
.is-night .badge-amber { background: rgba(245, 158, 11, 0.16); color: #FCD34D; }
.is-day .badge-rose { background: #FFE4E6; color: #BE123C; }
.is-night .badge-rose { background: rgba(244, 63, 94, 0.16); color: #FDA4AF; }

.s-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 1.5px;
  border-radius: 4px;
  color: var(--b-text-faint);
  cursor: pointer;
  transition: color 0.15s ease, background 0.15s ease;
}
.s-icon-btn:hover {
  color: var(--b-text-main);
  background: var(--b-pill);
}

.s-body {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 6px;
}

.s-reading {
  display: flex;
  align-items: baseline;
  gap: 3.5px;
}

.s-val {
  font-size: 22px;
  font-weight: 800;
  color: var(--b-text-main);
  letter-spacing: -0.03em;
  line-height: 1;
}
.sensors-grid.is-micro-cards .s-val { font-size: 19.5px; }
.sensors-grid.is-compact-cards .s-val { font-size: 17px; }

.s-val-muted {
  color: var(--b-text-faint);
}

.s-unit {
  font-size: 11px;
  font-weight: 600;
  color: var(--b-text-muted);
}
.sensors-grid.is-micro-cards .s-unit { font-size: 10px; }
.sensors-grid.is-compact-cards .s-unit { font-size: 9.5px; }

.s-held-tag {
  font-size: 8.5px;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 4px;
  background: rgba(245, 158, 11, 0.16);
  color: #F59E0B;
}

.s-spark-wrap {
  width: 72px;
  height: 25px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.s-spark-wrap.is-discrete-spark {
  border-radius: 4px;
  overflow: hidden;
}
.sensors-grid.is-micro-cards .s-spark-wrap {
  width: 58px;
  height: 21px;
}
.sensors-grid.is-compact-cards .s-spark-wrap {
  width: 48px;
  height: 16px;
}
.s-spark-svg {
  width: 100%;
  height: 100%;
  display: block;
}
.s-spark-empty {
  width: 38px;
  height: 2px;
  border-radius: 999px;
  background: var(--b-track);
}

.s-foot {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.s-bar-track {
  height: 3.5px;
  border-radius: 999px;
  background: var(--b-track);
  overflow: hidden;
}
.sensors-grid.is-compact-cards .s-bar-track {
  height: 3px;
}

.s-bar-fill {
  height: 100%;
  border-radius: 999px;
  transition: width 0.3s ease;
}

.s-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  font-size: 10px;
  color: var(--b-text-muted);
}
.sensors-grid.is-compact-cards .s-meta {
  font-size: 9px;
}

.s-meta-limit-btn {
  font-size: inherit;
  font-weight: 500;
  color: var(--b-text-muted);
  text-align: left;
  padding: 0;
  background: transparent;
  border: none;
  cursor: pointer;
  transition: color 0.15s ease;
}
.s-meta-limit-btn:hover {
  color: var(--b-accent-blue);
  text-decoration: underline;
}
.s-meta-limit-btn.is-custom-limit {
  font-weight: 600;
  color: var(--b-text-sec);
}

.s-meta-pct {
  font-weight: 700;
  color: var(--b-text-sec);
  flex-shrink: 0;
}

.s-card-quick-actions {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

/* High-Density Compact Cards inside Category-Separated Containers (Zero-Scroll Optimization) */
.sensors-grid.is-inside-cat-group .s-card {
  padding: 5px 6.5px;
  min-height: 56px;
  border-radius: 8.5px;
  gap: 2.5px;
}
.sensors-grid.is-compact-cards.is-inside-cat-group .s-card {
  padding: 4.5px 6px;
  min-height: 53px;
  border-radius: 8px;
  gap: 2px;
}
.sensors-grid.is-inside-cat-group .s-head {
  position: relative;
  gap: 3px;
}
.sensors-grid.is-inside-cat-group .s-title-wrap {
  gap: 4px;
}
.sensors-grid.is-inside-cat-group .s-dot {
  width: 5px;
  height: 5px;
}
.sensors-grid.is-inside-cat-group .s-name {
  font-size: 9.8px;
}
.sensors-grid.is-inside-cat-group .s-badge {
  font-size: 7.8px;
  padding: 1px 4px;
}
.sensors-grid.is-inside-cat-group .s-body {
  gap: 4px;
}
.sensors-grid.is-inside-cat-group .s-reading {
  gap: 2.5px;
}
.sensors-grid.is-inside-cat-group .s-val {
  font-size: 15px;
}
.sensors-grid.is-inside-cat-group .s-unit {
  font-size: 8.6px;
}
.sensors-grid.is-inside-cat-group .s-spark-wrap {
  width: 36px;
  height: 13px;
  flex-shrink: 1;
  min-width: 20px;
}
.sensors-grid.is-inside-cat-group .s-spark-empty {
  width: 24px;
}
.sensors-grid.is-inside-cat-group .s-discrete-pill {
  font-size: 9px;
  padding: 1px 5px;
}
.sensors-grid.is-inside-cat-group .s-discrete-dur {
  font-size: 8.5px;
}
.sensors-grid.is-inside-cat-group .s-foot {
  gap: 2px;
}
.sensors-grid.is-inside-cat-group .s-bar-track {
  height: 2.5px;
}
.sensors-grid.is-inside-cat-group .s-meta {
  font-size: 8.4px;
  line-height: 1.15;
}
@media (hover: hover) {
  .sensors-grid.is-inside-cat-group .s-card-quick-actions:not(.has-pinned-star) {
    display: none;
    position: absolute;
    right: 0;
    top: 50%;
    transform: translateY(-50%);
    background: var(--b-card);
    padding: 0 1px 0 4px;
    border-radius: 4px;
    z-index: 2;
    box-shadow: -4px 0 6px var(--b-card);
  }
  .sensors-grid.is-inside-cat-group .s-card-quick-actions.has-pinned-star .s-icon-menu-btn {
    display: none;
  }
  .sensors-grid.is-inside-cat-group .s-card:hover .s-card-quick-actions {
    display: inline-flex;
  }
  .sensors-grid.is-inside-cat-group .s-card:hover .s-card-quick-actions.has-pinned-star .s-icon-menu-btn {
    display: inline-flex;
  }
}

/* ========================================================================= */
/* 2B. SIDE-BY-SIDE 2-TABLE & 1-TABLE BENTO VIEW                             */
/* ========================================================================= */
.bento-table-area {
  max-height: 360px;
  overflow-y: auto;
  padding-right: 2px;
}
.bento-table-area.is-compact-scroll {
  max-height: 325px;
}
.bento-table-area.is-trend-hidden {
  max-height: none !important;
  overflow-y: visible !important;
  padding-right: 0;
}

.bento-dual-table-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 8px;
  align-items: start;
}
.bento-dual-table-grid.is-dual-cols {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
.bento-dual-table-grid.is-tri-cols {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.bento-table-card {
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 12px;
  box-shadow: var(--b-shadow);
  overflow: hidden;
  min-width: 0;
}

.bento-tbl-subhead {
  height: 30px;
  padding: 0 11px;
  box-sizing: border-box;
  background: var(--b-subtle);
  border-bottom: 1px solid var(--b-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 9.8px;
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--b-text-muted);
}
.bento-tbl-subhead.is-cat-subhead {
  color: var(--b-text-main);
}
.is-tri-cols .bento-tbl-subhead,
.is-dense-tbl .bento-tbl-subhead {
  height: 29px;
  padding: 0 9px;
  font-size: 9.2px;
}
.bento-tbl-subhead-left {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}
.bento-tbl-subhead-right {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-shrink: 0;
  color: var(--b-text-muted);
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}
.bento-tbl-subhead-right .s-badge {
  height: 17px;
  padding: 0 5px;
  display: inline-flex;
  align-items: center;
  line-height: 1;
}
.bento-cat-dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  flex-shrink: 0;
}

.bento-tbl-wrap {
  width: 100%;
  overflow-x: auto;
}

.bento-tbl {
  width: 100%;
  min-width: 376px;
  table-layout: fixed;
  border-collapse: collapse;
  text-align: left;
}
.bento-tbl.has-spark-col {
  min-width: 422px;
}

/* Deterministic Column Widths Across All Side-by-Side Tables */
.bento-tbl col.bt-col-param {
  width: auto;
}
.bento-tbl col.bt-col-val {
  width: 118px;
}
.bento-tbl col.bt-col-spark {
  width: 58px;
}
.bento-tbl col.bt-col-status {
  width: 68px;
}
.bento-tbl col.bt-col-limit {
  width: 86px;
}
.bento-tbl col.bt-col-update {
  width: 74px;
}
.bento-tbl col.bt-col-act {
  width: 50px;
}

.is-tri-cols .bento-tbl col.bt-col-val,
.is-dense-tbl .bento-tbl col.bt-col-val {
  width: 102px;
}
.is-tri-cols .bento-tbl col.bt-col-spark,
.is-dense-tbl .bento-tbl col.bt-col-spark {
  width: 46px;
}
.is-tri-cols .bento-tbl col.bt-col-status,
.is-dense-tbl .bento-tbl col.bt-col-status {
  width: 60px;
}
.is-tri-cols .bento-tbl col.bt-col-limit,
.is-dense-tbl .bento-tbl col.bt-col-limit {
  width: 74px;
}
.is-tri-cols .bento-tbl col.bt-col-act,
.is-dense-tbl .bento-tbl col.bt-col-act {
  width: 42px;
}

.bento-tbl thead tr {
  height: 25px;
}
.bento-tbl thead th {
  padding: 0 8px;
  font-size: 9.2px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--b-text-muted);
  border-bottom: 1px solid var(--b-border-soft);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: middle;
  box-sizing: border-box;
}
.is-tri-cols .bento-tbl thead th,
.is-dense-tbl .bento-tbl thead th {
  padding: 0 5px;
  font-size: 8.6px;
}
.bento-tbl thead th.bt-th-param {
  text-align: left;
  padding-left: 10px;
}
.is-tri-cols .bento-tbl thead th.bt-th-param,
.is-dense-tbl .bento-tbl thead th.bt-th-param {
  padding-left: 8px;
}
.bento-tbl thead th.bt-th-val {
  text-align: right;
  padding-left: 4px;
  padding-right: 48px;
}
.is-tri-cols .bento-tbl thead th.bt-th-val,
.is-dense-tbl .bento-tbl thead th.bt-th-val {
  padding-left: 3px;
  padding-right: 38px;
}
.bento-tbl thead th.bt-th-center {
  text-align: center;
  padding-left: 3px;
  padding-right: 3px;
}

.bento-tbl tbody tr {
  height: 33px;
}
.is-compact-vp .bento-tbl tbody tr,
.is-tri-cols .bento-tbl tbody tr,
.is-dense-tbl .bento-tbl tbody tr {
  height: 31px;
}

.bento-tbl tbody td {
  padding: 0 6px;
  font-size: 11px;
  border-bottom: 1px solid var(--b-border-soft);
  vertical-align: middle;
  box-sizing: border-box;
}
.is-day .bento-tbl thead th:not(:last-child),
.is-day .bento-tbl tbody td:not(:last-child) {
  border-right: 1px solid rgba(15, 23, 42, 0.045);
}
.is-night .bento-tbl thead th:not(:last-child),
.is-night .bento-tbl tbody td:not(:last-child) {
  border-right: 1px solid rgba(148, 163, 184, 0.055);
}

.bento-tbl tbody td.bt-td-param {
  padding-left: 10px;
  padding-right: 6px;
}
.is-tri-cols .bento-tbl tbody td.bt-td-param,
.is-dense-tbl .bento-tbl tbody td.bt-td-param {
  padding-left: 8px;
  padding-right: 5px;
}
.bento-tbl tbody td.bt-td-val {
  padding: 0 8px;
}
.is-tri-cols .bento-tbl tbody td.bt-td-val,
.is-dense-tbl .bento-tbl tbody td.bt-td-val {
  padding: 0 5px;
}
.bento-tbl tbody td.bt-td-center {
  text-align: center;
  padding: 0 3px;
}

.bento-tbl tbody tr:last-child td {
  border-bottom: none;
}

.bento-tbl-row {
  cursor: pointer;
  transition: background 0.12s ease;
}
.bento-tbl-row:hover {
  background: var(--b-subtle);
}
.is-day .bento-tbl-row.is-row-selected {
  background: rgba(14, 165, 233, 0.07);
}
.is-night .bento-tbl-row.is-row-selected {
  background: rgba(56, 189, 248, 0.11);
}
.bento-tbl-row.is-row-fault {
  background: rgba(244, 63, 94, 0.08);
}
.bento-tbl-row.is-row-warn {
  background: rgba(245, 158, 11, 0.08);
}

.bt-param-cell {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  min-width: 0;
}
.is-tri-cols .bt-param-cell,
.is-dense-tbl .bt-param-cell {
  gap: 5px;
}
.bt-param-cell .s-dot {
  width: 6.5px;
  height: 6.5px;
  flex-shrink: 0;
}
.bt-name {
  flex: 1 1 auto;
  min-width: 44px;
  font-weight: 700;
  color: var(--b-text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.is-tri-cols .bt-name,
.is-dense-tbl .bt-name {
  font-size: 10.5px;
}

/* Decimal-Aligned Numeric Reading + Fixed Unit Slot */
.bt-reading-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 36px;
  column-gap: 4px;
  align-items: baseline;
  width: 100%;
}
.is-tri-cols .bt-reading-grid,
.is-dense-tbl .bt-reading-grid {
  grid-template-columns: minmax(0, 1fr) 30px;
  column-gap: 3px;
}
.bt-val {
  text-align: right;
  font-weight: 800;
  font-size: 11.8px;
  font-variant-numeric: tabular-nums;
  color: var(--b-text-main);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.is-compact-vp .bt-val,
.is-tri-cols .bt-val,
.is-dense-tbl .bt-val {
  font-size: 11px;
}
.bt-unit-slot {
  text-align: left;
  font-size: 9.2px;
  font-weight: 600;
  color: var(--b-text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.is-tri-cols .bt-unit-slot,
.is-dense-tbl .bt-unit-slot {
  font-size: 8.5px;
}

.bt-reading-discrete {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding-right: 40px;
  width: 100%;
  box-sizing: border-box;
}
.is-tri-cols .bt-reading-discrete,
.is-dense-tbl .bt-reading-discrete {
  padding-right: 33px;
}
.bt-discrete-pill {
  height: 19px;
  padding: 0 6px;
  border-radius: 5px;
  border-width: 1px;
  border-style: solid;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  font-size: 9px;
  font-weight: 800;
  letter-spacing: 0.02em;
  text-transform: uppercase;
  line-height: 1;
  white-space: nowrap;
  box-sizing: border-box;
}

.bt-spark-box {
  width: 44px;
  height: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.is-compact-vp .bt-spark-box {
  width: 38px;
  height: 13px;
}
.is-tri-cols .bt-spark-box,
.is-dense-tbl .bt-spark-box {
  width: 34px;
  height: 13px;
}

/* Uniform Fixed-Width Status & Limit Pills */
.bt-status-pill {
  width: 58px;
  height: 19px;
  padding: 0 3px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  line-height: 1;
  white-space: nowrap;
  box-sizing: border-box;
}
.is-tri-cols .bt-status-pill,
.is-dense-tbl .bt-status-pill {
  width: 52px;
  height: 18.5px;
  padding: 0 2px;
  font-size: 8.6px;
}

.bt-limit-btn {
  position: relative;
  width: 76px;
  height: 19.5px;
  padding: 0 4px;
  border-radius: 5px;
  background: var(--b-pill);
  border: 1px solid var(--b-border-soft);
  color: var(--b-text-muted);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  overflow: hidden;
  cursor: pointer;
  box-sizing: border-box;
  transition: border-color 0.15s ease, color 0.15s ease;
}
.is-tri-cols .bt-limit-btn,
.is-dense-tbl .bt-limit-btn {
  width: 66px;
  height: 18.5px;
  padding: 0 3px;
  font-size: 8.4px;
}
.bt-limit-btn.is-custom-limit {
  color: var(--b-accent-blue);
  border-color: rgba(14, 165, 233, 0.25);
}
.bt-limit-btn.is-empty-limit {
  opacity: 0.65;
  font-weight: 600;
}
.bt-limit-btn:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-accent-blue);
  opacity: 1;
}
.bt-limit-text {
  position: relative;
  z-index: 1;
  overflow: hidden;
  text-overflow: ellipsis;
}
.bt-limit-bar {
  position: absolute;
  left: 0;
  bottom: 0;
  height: 2px;
  border-radius: 0 2px 0 0;
  opacity: 0.85;
  pointer-events: none;
}
.bt-limit-dash {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 76px;
  height: 19px;
  font-size: 10px;
  font-weight: 600;
  color: var(--b-text-muted);
  opacity: 0.45;
}
.is-tri-cols .bt-limit-dash,
.is-dense-tbl .bt-limit-dash {
  width: 66px;
  height: 18.5px;
}

.bt-fresh {
  font-size: 9.5px;
  font-variant-numeric: tabular-nums;
  color: var(--b-text-muted);
}

.bt-act-group {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
}
.is-tri-cols .bt-act-group,
.is-dense-tbl .bt-act-group {
  gap: 1px;
}
.is-tri-cols .bt-act-btn,
.is-dense-tbl .bt-act-btn {
  width: 17px;
  height: 18px;
}

.bento-empty-cards {
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 13px;
  padding: 32px 20px;
  text-align: center;
  font-size: 12px;
  color: var(--b-text-muted);
}

/* ========================================================================= */
/* 3. BOTTOM TREND STUDIO                                                    */
/* ========================================================================= */
.trend-studio {
  flex: 0 0 auto;
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 13px;
  box-shadow: var(--b-shadow);
  padding: 7px 12px 4px;
  display: flex;
  flex-direction: column;
  gap: 5px;
  position: relative;
}
.is-compact-vp .trend-studio {
  padding: 6px 10px 4px;
}
.trend-studio.is-collapsed-trend {
  padding: 7px 12px;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.ts-compact-bar {
  min-height: 28px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  position: relative;
  z-index: 25;
}
.ts-compact-bar.has-open-popover {
  z-index: 50;
}

.ts-c-left,
.ts-c-right {
  display: flex;
  align-items: center;
  gap: 5px;
  flex-wrap: wrap;
}

.ts-title {
  font-size: 12.5px;
  font-weight: 800;
  color: var(--b-text-main);
  white-space: nowrap;
}

.live-badge {
  height: 22px;
  padding: 0 7px;
  border-radius: 999px;
  font-size: 9.5px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 4.5px;
  white-space: nowrap;
}
.is-day .live-badge.is-live { background: #ECFDF5; color: #047857; }
.is-night .live-badge.is-live { background: rgba(16, 185, 129, 0.14); color: #6EE7B7; }
.is-day .live-badge.is-warn { background: #FEF3C7; color: #B45309; }
.is-night .live-badge.is-warn { background: rgba(245, 158, 11, 0.16); color: #FCD34D; }
.is-day .live-badge.is-offline { background: #FFF1F2; color: #BE123C; }
.is-night .live-badge.is-offline { background: rgba(244, 63, 94, 0.16); color: #FDA4AF; }

.diag-pill {
  height: 22px;
  padding: 0 7px;
  border-radius: 6px;
  font-size: 9.8px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border-soft);
  color: var(--b-text-muted);
}

/* Compact Trend Parameter Filter Trigger & Popover Grid */
.ts-filter-trigger {
  height: 26px;
  padding: 0 7.5px;
  border-radius: 7px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border);
  color: var(--b-text-sec);
  font-size: 10px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.ts-filter-trigger:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}
.ts-filter-trigger.active {
  background: rgba(14, 165, 233, 0.14);
  border-color: var(--b-accent-blue);
  color: var(--b-accent-blue);
}

.ts-mini-dots {
  display: inline-flex;
  align-items: center;
  gap: 2.5px;
}
.ts-mini-dot {
  width: 5.5px;
  height: 5.5px;
  border-radius: 50%;
}
.ts-filter-cnt {
  font-weight: 800;
  font-size: 9.8px;
}

.bmp-mini-act {
  height: 21px;
  padding: 0 5.5px;
  border-radius: 5px;
  font-size: 9.5px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 3px;
  background: var(--b-pill);
  border: 1px solid var(--b-border);
  color: var(--b-text-sec);
  cursor: pointer;
  transition: all 0.12s ease;
}
.bmp-mini-act:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}

.bmp-search-wrap {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 8px;
  margin-bottom: 6px;
  border-radius: 7px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border);
}
.bmp-search-input {
  width: 100%;
  background: transparent;
  border: none;
  outline: none;
  font-size: 10.5px;
  color: var(--b-text-main);
}
.bmp-search-clear {
  font-size: 13px;
  line-height: 1;
  color: var(--b-text-muted);
  cursor: pointer;
}

.trend-param-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 4px;
  max-height: 224px;
  overflow-y: auto;
  padding-right: 2px;
}

.tp-opt-btn {
  padding: 5px 7px;
  border-radius: 7px;
  font-size: 10px;
  font-weight: 600;
  background: var(--b-subtle);
  border: 1px solid var(--b-border-soft);
  color: var(--b-text-muted);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 4px;
  cursor: pointer;
  transition: all 0.12s ease;
  text-align: left;
  min-width: 0;
}
.tp-opt-btn:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}
.tp-opt-btn.active {
  background: rgba(14, 165, 233, 0.11);
  border-color: rgba(14, 165, 233, 0.45);
  color: var(--b-text-main);
  font-weight: 700;
}
.tp-opt-left {
  display: flex;
  align-items: center;
  gap: 5px;
}
.tp-dot {
  width: 6.5px;
  height: 6.5px;
  border-radius: 50%;
  flex-shrink: 0;
}
.tp-opt-btn:not(.active) .tp-dot {
  opacity: 0.45;
}
.tp-val {
  font-size: 9.5px;
  font-weight: 700;
  color: var(--b-text-sec);
  flex-shrink: 0;
}

.mini-icon-btn {
  width: 22px;
  height: 21px;
  border-radius: 5px;
  color: var(--b-text-sec);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.15s ease;
}
.mini-icon-btn:hover {
  color: var(--b-text-main);
  background: var(--b-subtle);
}

.cap-btn.compact-tf {
  height: 22px;
  padding: 0 6.5px;
  font-size: 10px;
}

.axis-legend {
  font-size: 9.8px;
  font-weight: 700;
  gap: 8px;
  white-space: nowrap;
  flex-shrink: 0;
  padding-right: 2px;
}

/* Trend Height Slider, Stepper & Drag Handle */
.bmp-step-btn {
  width: 24px;
  height: 24px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 800;
  background: var(--b-pill);
  border: 1px solid var(--b-border);
  color: var(--b-text-sec);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
}
.bmp-step-btn:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}

.trend-height-slider {
  accent-color: #0284C7;
  height: 5px;
  cursor: pointer;
}

.bento-chart-canvas {
  width: 100%;
  min-height: 120px;
  max-height: 560px;
  position: relative;
  overflow: hidden;
  border-radius: 10px;
}

.trend-resize-live-badge {
  position: absolute;
  top: 8px;
  right: 10px;
  z-index: 25;
  padding: 2.5px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 800;
  background: rgba(2, 132, 199, 0.90);
  color: #FFFFFF;
  pointer-events: none;
  box-shadow: 0 4px 10px rgba(2, 132, 199, 0.25);
}

.trend-resize-handle {
  height: 9px;
  margin-top: -1px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: ns-resize;
  user-select: none;
  touch-action: none;
  border-radius: 6px;
  transition: background 0.15s ease;
}
.trend-resize-handle:hover,
.trend-resize-handle.active {
  background: rgba(14, 165, 233, 0.10);
}
.trend-resize-grip {
  width: 36px;
  height: 3.5px;
  border-radius: 999px;
  background: var(--b-border);
  transition: background 0.15s ease, width 0.15s ease;
}
.trend-resize-handle:hover .trend-resize-grip,
.trend-resize-handle.active .trend-resize-grip {
  width: 48px;
  background: var(--b-accent-blue);
}

.bento-chart-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  color: var(--b-accent-blue);
  background: var(--b-tooltip-bg);
  z-index: 20;
}

.bento-chart-empty {
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: var(--b-text-muted);
  text-align: center;
  padding: 16px;
}

.bento-scrub-tooltip {
  position: absolute;
  z-index: 30;
  pointer-events: none;
  background: var(--b-tooltip-bg);
  border: 1px solid var(--b-border);
  border-radius: 11px;
  padding: 9px 11px;
  min-width: 192px;
  max-width: 265px;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.16);
  color: var(--b-text-main);
}

.bento-tt-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 10px;
  font-weight: 700;
  color: var(--b-text-muted);
  padding-bottom: 5px;
  margin-bottom: 5px;
  border-bottom: 1px solid var(--b-border-soft);
}

.bento-tt-tf {
  padding: 1px 5px;
  border-radius: 4px;
  background: var(--b-pill);
  color: var(--b-accent-blue);
  font-size: 9px;
}

/* ========================================================================= */
/* 4. DOCKED FOOTER                                                          */
/* ========================================================================= */
.dock-footer {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  height: 32px;
  background: var(--b-card);
  border-top: 1px solid var(--b-border);
  padding: 0 14px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  z-index: 30;
}

.dock-left,
.dock-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.dock-kicker {
  font-weight: 700;
  color: var(--b-text-muted);
  font-size: 10px;
  margin-right: 2px;
}

.dock-btn {
  height: 22px;
  padding: 0 9.5px;
  border-radius: 6px;
  font-size: 10.5px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border);
  color: var(--b-text-sec);
  cursor: pointer;
  transition: all 0.15s ease;
  white-space: nowrap;
}
.dock-btn:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}
.dock-btn.cyan {
  background: #0284C7;
  border-color: #0284C7;
  color: #FFFFFF;
  font-weight: 700;
}
.dock-btn.cyan:hover { background: #0369A1; }

.dock-btn.teal {
  background: #0D9488;
  border-color: #0D9488;
  color: #FFFFFF;
  font-weight: 700;
}
.dock-btn.teal:hover { background: #0F766E; }

.sound-pill {
  height: 22px;
  padding: 0 8.5px;
  border-radius: 6px;
  font-size: 10px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 4.5px;
  cursor: pointer;
  transition: opacity 0.15s ease;
}
.is-day .sound-pill {
  background: #E0F2FE;
  color: #0369A1;
  border: 1px solid #BAE6FD;
}
.is-night .sound-pill {
  background: rgba(56, 189, 248, 0.14);
  color: #7DD3FC;
  border: 1px solid rgba(56, 189, 248, 0.28);
}
.is-day .sound-pill.is-muted {
  background: #FEF3C7;
  color: #B45309;
  border-color: #FDE68A;
}
.is-night .sound-pill.is-muted {
  background: rgba(245, 158, 11, 0.16);
  color: #FCD34D;
  border-color: rgba(245, 158, 11, 0.30);
}

.no-perf-animations * {
  transition: none !important;
  animation: none !important;
}
</style>
