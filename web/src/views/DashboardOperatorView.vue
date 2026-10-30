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
        <!-- Area 1 (Left): Customer Company Branding (Anti-Truncate) -->
        <div
          class="brand-left cursor-pointer"
          @click="openBrandingModal('customer')"
          :title="(customerCompanyName || 'Customer Organization') + ' · Klik untuk konfigurasi branding'"
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
              <path d="M16.2 7.8c2.3 2.3 2.3 6.1 0 8.5" />
              <path d="M19.1 4.9C23 8.8 23 15.1 19.1 19" />
            </svg>
          </div>

          <div class="min-w-0 flex-1">
            <div class="station-name">
              <span class="truncate">{{ customerCompanyName || 'PT. Mitra Industri Enviro' }}</span>
              <span class="cbi-chip">CBI</span>
            </div>
            <div class="station-loc truncate">
              <span>Stasiun <strong>{{ stationTypeLabel || currentMonitoringDomain || 'AQMS' }}</strong></span>
              <span v-if="stationLocation && (stationLocation.station_name || stationLocation.address)">
                · {{ stationLocation.station_name || stationLocation.address }}
              </span>
            </div>
          </div>
        </div>

        <!-- Area 2 (Center): Adaptive Sensor Summary Pill + Calm Digital Clock -->
        <div class="clock-center" :class="{ 'is-wide-center': !isCompactScreen }">
          <button
            type="button"
            class="online-pill cursor-pointer"
            :class="bentoFleetSummary.toneClass"
            @click.stop="toggleToolbarPopover('healthAndAlarms')"
            :title="bentoFleetSummary.label + ' · Klik untuk membuka Pusat Kesehatan Stasiun & Alarm'"
          >
            <span class="status-dot" :class="bentoFleetSummary.dotClass"></span>
            <span>{{ isCompactScreen ? bentoFleetSummary.shortLabel : bentoFleetSummary.label }}</span>
          </button>

          <div v-if="currentTime" class="clock-pill">
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

        <!-- Area 3 (Right): Single Clean Maker Branding -->
        <div
          class="brand-right cursor-pointer"
          @click="openBrandingModal('maker')"
          :title="(makerCompanyName || 'PT. Centra Biotech Indonesia') + ' · Klik untuk konfigurasi branding'"
        >
          <div class="maker-info hidden lg:block">
            <div class="maker-title truncate" :class="isCompactScreen ? 'max-w-[165px]' : 'max-w-[215px]'">
              {{ makerCompanyName || 'PT. Centra Biotech Indonesia' }}
            </div>
            <div class="maker-sub truncate" :class="isCompactScreen ? 'max-w-[165px]' : 'max-w-[215px]'">
              {{ makerSubtitle || 'CBI Industrial IoT · Edge Telemetry' }}
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
              >{{ viewMode === 'table' ? bentoTableSplitMode : (effectiveBentoCols + 'P') }}</span>
            </button>
            <div
              v-if="activeToolbarPopover === 'viewOptions'"
              class="absolute left-0 top-full mt-1.5 z-50 w-[350px] sm:w-[370px] max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn"
              @click.stop
            >
              <!-- Panel Header -->
              <div class="flex items-center justify-between pb-2 mb-2 border-b border-slate-100 dark:border-[#1E3048]">
                <div class="flex items-center space-x-1.5">
                  <OperatorDashboardIcon name="sliders" :size="15" class="text-cyan-500" />
                  <span class="text-xs font-bold uppercase tracking-wider font-mono block leading-tight">
                    {{ $t('operator.viewAndFilterOptions') || 'View & Filter Options' }}
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

              <!-- Body with tidy sections -->
              <div class="max-h-[calc(85vh-100px)] overflow-y-auto custom-scroller pr-0.5 space-y-2.5">
                <!-- 1. Display Mode Section -->
                <div>
                  <div class="text-[9.5px] font-bold uppercase tracking-wider font-mono text-slate-500 dark:text-slate-400 mb-1.5 flex items-center justify-between">
                    <span>{{ $t('operator.displayMode') || 'Display Mode' }}</span>
                    <span class="text-[9px] text-cyan-500 font-semibold">{{ viewMode === 'cards' ? 'Cards' : 'Table' }}</span>
                  </div>
                  <div class="grid grid-cols-2 gap-1.5 font-mono">
                    <button
                      type="button"
                      @click="viewMode = 'cards'"
                      class="p-1.5 rounded-md border text-left transition flex items-center space-x-2"
                      :class="viewMode === 'cards'
                        ? 'bg-cyan-500/15 dark:bg-cyan-500/20 text-slate-900 dark:text-white border-cyan-500/60 shadow-xs font-bold'
                        : 'bg-slate-50 dark:bg-[#0E1726] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      <OperatorDashboardIcon name="layout" :size="14" class="text-cyan-500 flex-shrink-0" />
                      <span class="text-[10px] truncate flex-1">{{ $t('operator.viewCards') || 'Cards View' }}</span>
                      <span v-if="viewMode === 'cards'" class="text-cyan-500 font-bold text-xs">✓</span>
                    </button>
                    <button
                      type="button"
                      @click="viewMode = 'table'"
                      class="p-1.5 rounded-md border text-left transition flex items-center space-x-2"
                      :class="viewMode === 'table'
                        ? 'bg-cyan-500/15 dark:bg-cyan-500/20 text-slate-900 dark:text-white border-cyan-500/60 shadow-xs font-bold'
                        : 'bg-slate-50 dark:bg-[#0E1726] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      <OperatorDashboardIcon name="menu" :size="14" class="text-cyan-500 flex-shrink-0" />
                      <span class="text-[10px] truncate flex-1">{{ $t('operator.viewTable') || 'Table View' }}</span>
                      <span v-if="viewMode === 'table'" class="text-cyan-500 font-bold text-xs">✓</span>
                    </button>
                  </div>
                </div>

                <!-- 2. Panel Layout Density (1P - 4P Columns) -->
                <div>
                  <div class="text-[9.5px] font-bold uppercase tracking-wider font-mono text-slate-500 dark:text-slate-400 mb-1.5 flex items-center justify-between">
                    <span>{{ $t('operator.panelLayout') || 'Panel Layout' }}</span>
                    <span class="text-[9px] font-semibold" :class="viewMode === 'table' ? 'text-slate-400 italic' : 'text-cyan-500'">
                      {{ viewMode === 'table' ? '(N/A in Table)' : gridColumns + 'P Columns' }}
                    </span>
                  </div>
                  <div
                    class="grid grid-cols-4 gap-1 font-mono transition-opacity"
                    :class="viewMode === 'table' ? 'opacity-40 pointer-events-none' : ''"
                  >
                    <button
                      type="button"
                      v-for="cnt in [1, 2, 3, 4]"
                      :key="'occ-panel-btn-' + cnt"
                      :id="'btn-toolbar-panel-' + cnt"
                      @click="setGridColumns(cnt)"
                      class="p-1.5 rounded border text-center transition flex flex-col items-center justify-center"
                      :class="gridColumns === cnt
                        ? 'bg-cyan-500/15 dark:bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border-cyan-500/60 shadow-xs font-bold'
                        : 'bg-slate-50 dark:bg-[#0E1726] border-slate-200 dark:border-[#20354B] text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:border-slate-300 dark:hover:border-slate-700'"
                      :title="cnt + (cnt === 1 ? ' Column Grid (1P)' : ' Columns Grid (' + cnt + 'P)')"
                    >
                      <span class="text-xs font-black">{{ cnt }}P</span>
                      <span class="text-[8px] opacity-75">{{ cnt === 1 ? '1 Col' : cnt + ' Cols' }}</span>
                    </button>
                  </div>
                </div>

                <!-- 3. Pinned Favorites Section -->
                <div class="p-2 rounded bg-slate-50 dark:bg-[#0E1726] border border-slate-200 dark:border-[#20354B] font-mono">
                  <div class="flex items-center justify-between">
                    <div class="flex items-center space-x-1.5">
                      <span class="text-amber-400 text-xs">⭐</span>
                      <div>
                        <div class="text-[10px] font-bold text-slate-900 dark:text-white leading-tight">
                          {{ $t('operator.pinnedOnly') || 'Pinned Only' }}
                        </div>
                        <div class="text-[8.5px] text-slate-500 dark:text-slate-400">
                          Filter CBI cards to favorites
                        </div>
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

                  <!-- Pinned items list if any -->
                  <div v-if="pinnedParametersList.length > 0" class="mt-2 pt-2 border-t border-slate-200/70 dark:border-[#1E3048]">
                    <div class="flex items-center justify-between text-[9px] text-slate-500 dark:text-slate-400 mb-1">
                      <span>Favorites ({{ pinnedParametersList.length }})</span>
                      <button
                        type="button"
                        @click="clearAllPinned"
                        class="text-rose-500 hover:text-rose-600 dark:hover:text-rose-400 hover:underline"
                      >
                        {{ $t('operator.clearAllPinned') || 'Clear All' }}
                      </button>
                    </div>
                    <div class="max-h-28 overflow-y-auto custom-scroller space-y-1">
                      <div
                        v-for="param in pinnedParametersList"
                        :key="'pinned-item-' + param.param_id"
                        class="flex items-center justify-between p-1 rounded bg-white dark:bg-[#101B2B] border border-slate-200 dark:border-[#20354B]"
                      >
                        <div class="min-w-0 flex-1 pr-1.5">
                          <div class="font-bold text-slate-800 dark:text-slate-200 truncate text-[9.5px]">
                            {{ param.param_name || param.param_code }}
                          </div>
                          <div class="text-[8px] text-slate-500 dark:text-slate-400 flex items-center space-x-1 truncate">
                            <span v-if="!isOperator">{{ param.device_code }}</span>
                            <span v-if="param.unit">· {{ param.unit }}</span>
                          </div>
                        </div>
                        <button
                          type="button"
                          @click.stop="togglePin(param.param_id)"
                          class="p-0.5 rounded text-amber-400 hover:text-rose-500 transition hover:bg-slate-100 dark:hover:bg-slate-800 flex-shrink-0"
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
                  <div class="text-[9.5px] font-bold uppercase tracking-wider font-mono text-slate-500 dark:text-slate-400 mb-1.5 flex items-center justify-between">
                    <div class="flex items-center space-x-1.5">
                      <OperatorDashboardIcon name="filter" :size="12" class="text-cyan-500" />
                      <span>{{ $t('operator.categories') || 'Parameter Categories' }}</span>
                    </div>
                    <span v-if="selectedCategoryFilter !== 'all'" class="text-[9px] text-cyan-500 font-semibold">Filtered</span>
                  </div>

                  <!-- Categories List -->
                  <div class="max-h-48 overflow-y-auto custom-scroller space-y-1 font-mono text-[10.5px]">
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

                    <!-- Not Uncategorized (Categorized Only - Excludes uncategorized parameters) -->
                    <button
                      type="button"
                      id="btn-filter-not-uncategorized"
                      @click="selectedCategoryFilter = 'not_uncategorized'"
                      class="w-full flex items-center justify-between px-2 py-1.5 rounded transition border text-left group"
                      :class="selectedCategoryFilter === 'not_uncategorized'
                        ? 'bg-emerald-500/15 dark:bg-emerald-500/20 text-emerald-800 dark:text-emerald-300 border-emerald-500/50 shadow-xs font-bold ring-1 ring-emerald-500/30'
                        : 'border-transparent text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'"
                      :title="$t('operator.notUncategorizedDesc') || 'Menampilkan semua parameter di luar Tanpa Kategori'"
                    >
                      <div class="flex items-center space-x-2 truncate">
                        <OperatorDashboardIcon name="filter" :size="12" class="text-emerald-500 flex-shrink-0" />
                        <div class="truncate">
                          <span class="truncate block leading-tight font-medium" :class="selectedCategoryFilter === 'not_uncategorized' ? 'font-bold' : ''">
                            {{ $t('operator.notUncategorized') || 'Not Uncategorized' }}
                          </span>
                          <span class="text-[8.5px] text-slate-500 dark:text-slate-400 block font-normal leading-none mt-0.5">
                            {{ $t('operator.notUncategorizedDesc') || 'Di luar uncategorized' }}
                          </span>
                        </div>
                      </div>
                      <div class="flex items-center space-x-1.5 flex-shrink-0">
                        <span
                          class="px-1.5 py-0.2 rounded-full text-[9px] font-bold"
                          :class="selectedCategoryFilter === 'not_uncategorized' ? 'bg-emerald-500/20 text-emerald-700 dark:text-emerald-300' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400'"
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

                    <!-- Uncategorized Filter (All Parameters without category) -->
                    <button
                      type="button"
                      id="btn-filter-uncategorized"
                      @click="selectedCategoryFilter = 'uncategorized'"
                      class="w-full flex items-center justify-between px-2 py-1.5 rounded transition border text-left group"
                      :class="selectedCategoryFilter === 'uncategorized'
                        ? 'bg-slate-200/90 dark:bg-slate-700/60 text-slate-900 dark:text-white border-slate-400 dark:border-slate-500 shadow-xs font-bold ring-1 ring-slate-400/40'
                        : 'border-transparent text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-[#162538] hover:text-slate-900 dark:hover:text-white'"
                      :title="$t('operator.uncategorizedDesc') || 'Semua parameter yang tidak masuk ke category'"
                    >
                      <div class="flex items-center space-x-2 truncate">
                        <span class="w-2 h-2 rounded-full flex-shrink-0 bg-slate-400 dark:bg-slate-400 ring-1 ring-slate-400/50"></span>
                        <div class="truncate">
                          <span class="truncate block leading-tight font-medium" :class="selectedCategoryFilter === 'uncategorized' ? 'font-bold' : ''">
                            {{ $t('operator.uncategorized') || 'Tanpa Kategori' }}
                          </span>
                          <span class="text-[8.5px] text-slate-500 dark:text-slate-400 block font-normal leading-none mt-0.5">
                            {{ $t('operator.uncategorizedDesc') || 'Tidak masuk ke category' }}
                          </span>
                        </div>
                      </div>
                      <div class="flex items-center space-x-1.5 flex-shrink-0">
                        <span
                          class="px-1.5 py-0.2 rounded-full text-[9px] font-bold"
                          :class="selectedCategoryFilter === 'uncategorized' ? 'bg-slate-300 dark:bg-slate-600 text-slate-900 dark:text-white' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400'"
                        >
                          {{ uncategorizedParamCount }}
                        </span>
                        <span v-if="selectedCategoryFilter === 'uncategorized'" class="text-cyan-600 dark:text-cyan-400 font-bold text-xs">✓</span>
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
              class="absolute left-0 top-full mt-1.5 z-50 w-[420px] sm:w-[460px] max-w-[calc(100vw-24px)] rounded-xl shadow-2xl p-3.5 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#0E1726] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono max-h-[88vh] overflow-y-auto custom-scroller"
              @click.stop
            >
              <!-- Panel Header -->
              <div class="flex items-center justify-between pb-2.5 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
                <div class="flex items-center space-x-2.5">
                  <div class="w-7 h-7 rounded-lg flex items-center justify-center bg-cyan-500/15 border border-cyan-500/40 text-cyan-600 dark:text-cyan-400 flex-shrink-0 shadow-xs">
                    <OperatorDashboardIcon name="performance" :size="16" />
                  </div>
                  <div>
                    <div class="flex items-center space-x-2">
                      <span class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                        {{ $t('operator.perfSettingsTitle') || 'Performance Settings' }}
                      </span>
                      <span
                        v-if="isPerformanceOptimizedActive"
                        class="px-1.5 py-0.5 rounded text-[8px] font-bold bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/35 leading-none flex items-center space-x-1"
                      >
                        <span class="w-1.5 h-1.5 rounded-full bg-amber-400 animate-pulse"></span>
                        <span>OPTIMIZED</span>
                      </span>
                      <span
                        v-else
                        class="px-1.5 py-0.5 rounded text-[8px] font-semibold bg-slate-100 dark:bg-[#162538] text-slate-500 dark:text-slate-400 border border-slate-200 dark:border-[#1E3048] leading-none"
                      >
                        STANDARD
                      </span>
                    </div>
                    <div class="text-[9.5px] text-slate-500 dark:text-slate-400 mt-1 leading-none font-normal">
                      {{ $t('operator.perfSettingsSubtitle') || 'Raspberry Pi & Low-Power Chromium Optimization' }}
                    </div>
                  </div>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="w-7 h-7 flex items-center justify-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-[#18283E] rounded-lg transition"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- 1-Tap Quick Optimization Profiles -->
              <div class="mb-2.5 p-1.5 rounded-lg bg-slate-50 dark:bg-[#070D17] border border-slate-200 dark:border-[#1E3048]">
                <div class="flex items-center justify-between mb-1 px-1 text-[9px] uppercase tracking-wider font-bold text-slate-500 dark:text-slate-400">
                  <span>{{ $t('operator.perfPresetQuick') || 'Quick Presets' }}</span>
                  <span class="text-cyan-600 dark:text-cyan-400 font-semibold lowercase">tap to apply</span>
                </div>
                <div class="grid grid-cols-3 gap-1">
                  <button
                    type="button"
                    @click="setPerfPreset('standard')"
                    class="py-1 px-1.5 rounded text-[9.5px] font-bold border transition text-center cursor-pointer flex items-center justify-center space-x-1"
                    :class="currentPerfPreset === 'standard'
                      ? 'bg-cyan-500 text-slate-950 border-cyan-400 shadow-xs font-black'
                      : 'bg-white dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-slate-400'"
                    title="All features active for standard desktop workstations"
                  >
                    <span>⚡</span>
                    <span>{{ $t('operator.perfPresetStandard') || 'Standard' }}</span>
                  </button>
                  <button
                    type="button"
                    @click="setPerfPreset('balanced')"
                    class="py-1 px-1.5 rounded text-[9.5px] font-bold border transition text-center cursor-pointer flex items-center justify-center space-x-1"
                    :class="currentPerfPreset === 'balanced'
                      ? 'bg-cyan-500 text-slate-950 border-cyan-400 shadow-xs font-black'
                      : 'bg-white dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-slate-400'"
                    title="Balanced profile for Raspberry Pi 4 (sparklines off, full trend active)"
                  >
                    <span>⚖️</span>
                    <span>{{ $t('operator.perfPresetBalanced') || 'Balanced' }}</span>
                  </button>
                  <button
                    type="button"
                    @click="setPerfPreset('max_eco')"
                    class="py-1 px-1.5 rounded text-[9.5px] font-bold border transition text-center cursor-pointer flex items-center justify-center space-x-1"
                    :class="currentPerfPreset === 'max_eco'
                      ? 'bg-amber-400 text-slate-950 border-amber-300 shadow-xs font-black'
                      : 'bg-white dark:bg-[#101B2B] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-slate-400'"
                    title="Maximum CPU savings for Raspberry Pi 3 or slow networks"
                  >
                    <span>🚀</span>
                    <span>{{ $t('operator.perfPresetMaxEco') || 'Max Eco' }}</span>
                  </button>
                </div>
              </div>

              <!-- Storage Badge Notice & Reset -->
              <div class="mb-2.5 px-2.5 py-1.5 rounded-lg bg-slate-50 dark:bg-[#070D17] border border-slate-200/70 dark:border-[#1E3048] flex items-center justify-between text-[9px] text-slate-500 dark:text-slate-400">
                <span class="flex items-center space-x-1.5">
                  <svg class="w-3.5 h-3.5 text-cyan-500 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 7v10c0 2 1.5 3 3.5 3h9c2 0 3.5-1 3.5-3V7M4 7c0-2 1.5-3 3.5-3h9c2 0 3.5 1 3.5 3M4 7h16m-8 4v4m-2-2h4" />
                  </svg>
                  <span class="font-mono">{{ $t('operator.perfStorageNotice') || 'Saved in this browser only (localStorage)' }}</span>
                </span>
                <button
                  type="button"
                  @click="resetPerformanceSettings"
                  class="text-cyan-600 dark:text-cyan-400 hover:text-cyan-500 font-bold text-[9px] uppercase tracking-wider flex items-center space-x-1 py-0.5 px-1.5 rounded hover:bg-cyan-500/10 transition cursor-pointer"
                >
                  <OperatorDashboardIcon name="reset" :size="10" />
                  <span>{{ $t('operator.perfResetDefaults') || 'Reset Defaults' }}</span>
                </button>
              </div>

              <!-- Content Cards -->
              <div class="space-y-2">
                <!-- 1. Heavy Graphics & Charting -->
                <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#070D17] border border-slate-200 dark:border-[#1E3048] space-y-2.5">
                  <div class="text-[9.5px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                    <OperatorDashboardIcon name="chart" :size="12" class="text-cyan-500" />
                    <span>Heavy Graphics & Computations</span>
                  </div>

                  <!-- A. Sparklines -->
                  <div class="flex items-center justify-between pt-1 border-t border-slate-200/60 dark:border-[#18283E]">
                    <div class="pr-2 min-w-0 flex-1">
                      <div class="flex items-center space-x-1.5">
                        <span class="text-[11px] font-bold text-slate-900 dark:text-white truncate">
                          {{ $t('operator.perfSparklines') || 'Parameter Card Sparklines' }}
                        </span>
                        <span
                          v-if="!perfSettings.parameterSparklinesEnabled"
                          class="px-1 py-0.2 rounded text-[7.5px] font-bold bg-amber-500/15 text-amber-500 border border-amber-500/30 leading-none shrink-0"
                        >
                          CPU SAVED
                        </span>
                      </div>
                      <div class="text-[9px] text-slate-500 dark:text-slate-400 mt-0.5 leading-tight">
                        {{ $t('operator.perfSparklinesDesc') || 'Render mini SVG sparklines on parameter cards' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-[9px] font-mono font-bold" :class="perfSettings.parameterSparklinesEnabled ? 'text-cyan-600 dark:text-cyan-400' : 'text-slate-400'">
                        {{ perfSettings.parameterSparklinesEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('parameterSparklinesEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.parameterSparklinesEnabled ? 'bg-cyan-500 shadow-glow-cyan-xs' : 'bg-slate-300 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.parameterSparklinesEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.parameterSparklinesEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>

                  <!-- B. Trend Analysis Chart -->
                  <div class="flex items-center justify-between pt-2 border-t border-slate-200/60 dark:border-[#18283E]">
                    <div class="pr-2 min-w-0 flex-1">
                      <div class="flex items-center space-x-1.5">
                        <span class="text-[11px] font-bold text-slate-900 dark:text-white truncate">
                          {{ $t('operator.perfTrend') || 'Trend Analysis Chart' }}
                        </span>
                        <span
                          v-if="!perfSettings.trendAnalysisEnabled"
                          class="px-1 py-0.2 rounded text-[7.5px] font-bold bg-amber-500/15 text-amber-500 border border-amber-500/30 leading-none shrink-0"
                        >
                          UNMOUNTED
                        </span>
                      </div>
                      <div class="text-[9px] text-slate-500 dark:text-slate-400 mt-0.5 leading-tight">
                        {{ $t('operator.perfTrendDesc') || 'Multi-parameter trend analysis panel and 1M/5S intervals' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-[9px] font-mono font-bold" :class="perfSettings.trendAnalysisEnabled ? 'text-cyan-600 dark:text-cyan-400' : 'text-slate-400'">
                        {{ perfSettings.trendAnalysisEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('trendAnalysisEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.trendAnalysisEnabled ? 'bg-cyan-500 shadow-glow-cyan-xs' : 'bg-slate-300 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.trendAnalysisEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.trendAnalysisEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>
                </div>

                <!-- 2. Live Telemetry Cadence -->
                <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#070D17] border border-slate-200 dark:border-[#1E3048] space-y-2">
                  <div class="flex items-center justify-between">
                    <div class="text-[9.5px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                      <OperatorDashboardIcon name="activity" :size="12" class="text-amber-500" />
                      <span>{{ $t('operator.perfLiveUpdates') || 'Live Dashboard Updates' }}</span>
                    </div>
                    <span class="text-[8.5px] font-mono uppercase font-bold text-cyan-600 dark:text-cyan-400">
                      {{ perfSettings.liveUpdateMode }}
                    </span>
                  </div>

                  <!-- Segmented Buttons -->
                  <div class="p-1 rounded-lg bg-slate-200/70 dark:bg-[#101B2B] border border-slate-200 dark:border-[#1E3048] grid grid-cols-3 gap-1">
                    <button
                      type="button"
                      @click="setPerfSetting('liveUpdateMode', 'full')"
                      class="py-1 px-1.5 rounded-md text-[9.5px] font-bold transition text-center cursor-pointer"
                      :class="perfSettings.liveUpdateMode === 'full'
                        ? 'bg-cyan-500 text-slate-950 shadow-xs font-black'
                        : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      Full Live
                    </button>
                    <button
                      type="button"
                      @click="setPerfSetting('liveUpdateMode', 'essential')"
                      class="py-1 px-1.5 rounded-md text-[9.5px] font-bold transition text-center cursor-pointer"
                      :class="perfSettings.liveUpdateMode === 'essential'
                        ? 'bg-cyan-500 text-slate-950 shadow-xs font-black'
                        : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      Essential
                    </button>
                    <button
                      type="button"
                      @click="setPerfSetting('liveUpdateMode', 'reduced')"
                      class="py-1 px-1.5 rounded-md text-[9.5px] font-bold transition text-center cursor-pointer"
                      :class="perfSettings.liveUpdateMode === 'reduced'
                        ? 'bg-cyan-500 text-slate-950 shadow-xs font-black'
                        : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      Reduced (1s)
                    </button>
                  </div>

                  <!-- Description for selected mode -->
                  <div class="text-[8.5px] text-slate-500 dark:text-slate-400 leading-tight px-0.5">
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
                <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#070D17] border border-slate-200 dark:border-[#1E3048] space-y-2">
                  <div class="text-[9.5px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                    <OperatorDashboardIcon name="bell" :size="12" class="text-rose-400" />
                    <span>{{ $t('operator.perfAlarms') || 'Alarm Display & Sound' }}</span>
                  </div>

                  <!-- Alarm Visual -->
                  <div class="flex items-center justify-between pt-1 border-t border-slate-200/60 dark:border-[#18283E]">
                    <div class="pr-2 min-w-0 flex-1">
                      <div class="text-[10.5px] font-semibold text-slate-800 dark:text-slate-200 truncate">
                        {{ $t('operator.perfAlarmVisual') || 'Alarm Visual (Flashing & Pulse)' }}
                      </div>
                      <div class="text-[8.5px] text-slate-500 dark:text-slate-400 leading-tight mt-0.5">
                        {{ $t('operator.perfAlarmVisualDesc') || 'Flashing borders and pulse animations on cards during alarms' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-[9px] font-mono font-bold" :class="perfSettings.alarmVisualEnabled ? 'text-cyan-600 dark:text-cyan-400' : 'text-slate-400'">
                        {{ perfSettings.alarmVisualEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('alarmVisualEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.alarmVisualEnabled ? 'bg-cyan-500 shadow-glow-cyan-xs' : 'bg-slate-300 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.alarmVisualEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.alarmVisualEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>

                  <!-- Alarm Sound -->
                  <div class="flex items-center justify-between pt-2 border-t border-slate-200/60 dark:border-[#18283E]">
                    <div class="pr-2 min-w-0 flex-1">
                      <div class="text-[10.5px] font-semibold text-slate-800 dark:text-slate-200 truncate">
                        {{ $t('operator.perfAlarmSound') || 'Alarm Sound (Audio Buzzer)' }}
                      </div>
                      <div class="text-[8.5px] text-slate-500 dark:text-slate-400 leading-tight mt-0.5">
                        {{ $t('operator.perfAlarmSoundDesc') || 'Audible chime when critical alarm incidents occur' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-[9px] font-mono font-bold" :class="perfSettings.alarmSoundEnabled ? 'text-cyan-600 dark:text-cyan-400' : 'text-slate-400'">
                        {{ perfSettings.alarmSoundEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('alarmSoundEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.alarmSoundEnabled ? 'bg-cyan-500 shadow-glow-cyan-xs' : 'bg-slate-300 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.alarmSoundEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.alarmSoundEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>

                  <!-- Alarm Safety Notice -->
                  <div class="mt-1 px-2.5 py-1.5 rounded bg-amber-500/10 border border-amber-500/25 flex items-start space-x-2 text-[8.5px] text-amber-700 dark:text-amber-300 leading-tight">
                    <span class="text-amber-500 shrink-0 font-bold">⚠</span>
                    <span>{{ $t('operator.perfAlarmNotice') || 'Critical alarms remain logged on backend. Disabling visual or sound may delay operator awareness.' }}</span>
                  </div>
                </div>

                <!-- 4. Visual Polish & Animations -->
                <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#070D17] border border-slate-200 dark:border-[#1E3048] space-y-2">
                  <div class="text-[9.5px] font-bold uppercase tracking-wider text-slate-400 dark:text-slate-500 flex items-center space-x-1.5">
                    <svg class="w-3 h-3 text-cyan-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
                    </svg>
                    <span>{{ $t('operator.perfVisualEffects') || 'Visual Effects' }}</span>
                  </div>

                  <!-- Tile Ingestion Blinks -->
                  <div class="flex items-center justify-between pt-1 border-t border-slate-200/60 dark:border-[#18283E]">
                    <div class="pr-2 min-w-0 flex-1">
                      <div class="text-[10.5px] font-semibold text-slate-800 dark:text-slate-200 truncate">
                        {{ $t('operator.perfCardBlinks') || 'Tile Ingestion Blinks' }}
                      </div>
                      <div class="text-[8.5px] text-slate-500 dark:text-slate-400 leading-tight mt-0.5">
                        {{ $t('operator.perfCardBlinksDesc') || 'Cyan flash highlight on tiles when new telemetry packets arrive' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-[9px] font-mono font-bold" :class="perfSettings.cardBlinksEnabled ? 'text-cyan-600 dark:text-cyan-400' : 'text-slate-400'">
                        {{ perfSettings.cardBlinksEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('cardBlinksEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.cardBlinksEnabled ? 'bg-cyan-500 shadow-glow-cyan-xs' : 'bg-slate-300 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.cardBlinksEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.cardBlinksEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>

                  <!-- Chart Animations -->
                  <div class="flex items-center justify-between pt-2 border-t border-slate-200/60 dark:border-[#18283E]">
                    <div class="pr-2 min-w-0 flex-1">
                      <div class="text-[10.5px] font-semibold text-slate-800 dark:text-slate-200 truncate">
                        {{ $t('operator.perfChartAnimations') || 'Chart Path Animations' }}
                      </div>
                      <div class="text-[8.5px] text-slate-500 dark:text-slate-400 leading-tight mt-0.5">
                        {{ $t('operator.perfChartAnimationsDesc') || 'Smooth transition effects on chart lines and curves' }}
                      </div>
                    </div>
                    <div class="flex items-center space-x-2 shrink-0">
                      <span class="text-[9px] font-mono font-bold" :class="perfSettings.chartAnimationsEnabled ? 'text-cyan-600 dark:text-cyan-400' : 'text-slate-400'">
                        {{ perfSettings.chartAnimationsEnabled ? 'ON' : 'OFF' }}
                      </span>
                      <button
                        type="button"
                        @click="togglePerfSetting('chartAnimationsEnabled')"
                        class="relative inline-flex h-5 w-9 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                        :class="perfSettings.chartAnimationsEnabled ? 'bg-cyan-500 shadow-glow-cyan-xs' : 'bg-slate-300 dark:bg-slate-700'"
                        role="switch"
                        :aria-checked="perfSettings.chartAnimationsEnabled"
                      >
                        <span
                          class="pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow-md ring-0 transition duration-200 ease-in-out"
                          :class="perfSettings.chartAnimationsEnabled ? 'translate-x-4' : 'translate-x-0'"
                        ></span>
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 5. ISPU Settings Popover (Icon Only + Emerald Dot) -->
          <div class="relative toolbar-popover-anchor">
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
              class="absolute left-0 top-full mt-1.5 z-50 w-[390px] sm:w-[410px] max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono"
              @click.stop
            >
              <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
                <div class="flex items-center space-x-2">
                  <div class="w-6 h-6 rounded flex items-center justify-center bg-emerald-500/10 border border-emerald-500/30 text-emerald-500 flex-shrink-0">
                    <OperatorDashboardIcon name="ispu" :size="14" />
                  </div>
                  <div class="flex items-center space-x-1.5">
                    <span class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                      {{ $t('operator.ispuSettings') || 'ISPU Settings' }}
                    </span>
                    <span v-if="isAQMSDomain" class="px-1.5 py-0.2 rounded text-[8px] font-bold bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 leading-none">
                      AQMS
                    </span>
                  </div>
                </div>
                <button
                  type="button"
                  @click="closeToolbarPopover"
                  class="w-6 h-6 flex items-center justify-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-[#18283E] rounded transition"
                  title="Close"
                >
                  <OperatorDashboardIcon name="close" :size="14" />
                </button>
              </div>

              <!-- Unified Station Scope Row -->
              <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0E1726] border border-slate-200 dark:border-[#20354B] mb-2.5 space-y-1 font-mono">
                <div class="flex items-center justify-between text-[11px]">
                  <span class="font-bold text-slate-900 dark:text-white">
                    Lingkup Kalkulasi ISPU
                  </span>
                  <span class="text-4xs text-cyan-600 dark:text-cyan-400 font-bold">
                    UNIFIED
                  </span>
                </div>
                <div class="text-[9px] text-slate-500 dark:text-slate-400 leading-tight">
                  Menggabungkan seluruh parameter resmi yang aktif dari semua device yang aktif di stasiun ini.
                </div>
              </div>

              <!-- Toggle Show/Hide ISPU Panel Row -->
              <div class="p-2.5 rounded-lg bg-slate-50 dark:bg-[#0E1726] border border-slate-200 dark:border-[#20354B] mb-2.5">
                <div class="flex items-center justify-between">
                  <span class="text-[11px] font-bold text-slate-900 dark:text-white">
                    {{ $t('operator.showIspuBar') || 'Show ISPU Indicator Bar' }}
                  </span>
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
                <div class="text-[9.5px] text-slate-500 dark:text-slate-400 mt-1.5 leading-relaxed">
                  {{ $t('operator.showIspuBarDesc') || 'Display air quality index bar below parameters' }}
                </div>
              </div>

              <!-- Standard Info Card -->
              <div class="p-2.5 rounded-lg bg-slate-50/70 dark:bg-[#0A1220] border border-slate-200/80 dark:border-[#1E3048] mb-2.5 text-[9.5px] space-y-1.5">
                <div class="flex items-center justify-between text-slate-700 dark:text-slate-300">
                  <span class="text-slate-400">Standard:</span>
                  <span class="font-bold text-cyan-600 dark:text-cyan-400">Permen LHK No. 14 / 2020</span>
                </div>
                <div class="flex items-center justify-between text-slate-700 dark:text-slate-300">
                  <span class="text-slate-400">Interval:</span>
                  <span class="font-semibold text-slate-700 dark:text-slate-300">24-Hour Rolling Average</span>
                </div>
                <div class="flex items-center justify-between text-slate-700 dark:text-slate-300">
                  <span class="text-slate-400">Parameters:</span>
                  <span class="font-mono text-[9px] text-slate-700 dark:text-slate-300">PM2.5, PM10, SO2, NO2, O3, CO</span>
                </div>
              </div>

              <!-- Navigation to Dedicated ISPU Analysis -->
              <button
                type="button"
                @click="navigateToRoute('/monitoring/ispu')"
                class="w-full py-2 px-3 rounded-md text-[10.5px] font-semibold transition border flex items-center justify-center space-x-2 bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border-cyan-500/30 hover:bg-cyan-500/20 hover:border-cyan-400"
              >
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
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
                class="absolute right-0 top-full mt-2 w-96 max-w-[92vw] bg-white dark:bg-[#0E1726] border border-slate-200 dark:border-[#1E3048] rounded-xl shadow-2xl z-50 p-4 space-y-3.5 select-none"
                @click.stop
              >
                <!-- Popover Header -->
                <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-2.5">
                  <div class="flex items-center space-x-2">
                    <span class="w-6 h-6 rounded-md bg-amber-500/10 text-amber-500 flex items-center justify-center text-xs font-bold font-mono">⚡</span>
                    <h3 class="text-xs font-bold uppercase tracking-wider text-slate-800 dark:text-slate-100 font-mono">
                      Controls & Actuators
                    </h3>
                  </div>
                  <span class="text-3xs font-mono px-2 py-0.5 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
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
                class="absolute right-0 top-full mt-1.5 z-50 w-[410px] sm:w-[430px] max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3.5 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono"
                @click.stop
              >
                <!-- Panel Header -->
                <div class="flex items-center justify-between pb-2.5 mb-3 border-b border-slate-200/80 dark:border-[#1E3048]">
                  <div class="flex items-center space-x-2.5">
                    <div
                      class="w-7 h-7 rounded flex items-center justify-center border flex-shrink-0"
                      :class="activeCriticalCount > 0
                        ? 'bg-rose-500/15 border-rose-500/40 text-rose-500'
                        : activeAlarmsList.length > 0
                          ? 'bg-amber-500/15 border-amber-500/40 text-amber-400'
                          : 'bg-cyan-500/15 border-cyan-500/40 text-cyan-400'"
                    >
                      <OperatorDashboardIcon name="pulse" :size="15" />
                    </div>
                    <div>
                      <div class="flex items-center space-x-2">
                        <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                          Health & Notification Center
                        </h3>
                        <span
                          class="px-1.5 py-0.5 rounded text-[8px] font-bold uppercase tracking-wider border leading-none"
                          :class="activeCriticalCount > 0
                            ? 'bg-rose-500/15 text-rose-600 dark:text-rose-400 border-rose-500/30'
                            : activeAlarmsList.length > 0
                              ? 'bg-amber-500/15 text-amber-600 dark:text-amber-400 border-amber-500/30'
                              : 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border-emerald-500/30'"
                        >
                          {{ activeCriticalCount > 0 ? 'CRITICAL' : activeAlarmsList.length > 0 ? 'WARNING' : 'HEALTHY' }}
                        </span>
                      </div>
                      <div class="text-[9.5px] text-slate-500 dark:text-slate-400 mt-1 leading-none">
                        Alarms · Device Fleet · Telemetry Stream
                      </div>
                    </div>
                  </div>
                  <button
                    type="button"
                    @click="closeToolbarPopover"
                    class="w-6 h-6 flex items-center justify-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-[#18283E] rounded transition"
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
                      <span>Cycle Rate: <strong class="text-cyan-600 dark:text-cyan-400 font-mono">{{ Math.round(configuredPollingIntervalMs / 1000) }}s</strong></span>
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
                          Packet Latency
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
                        {{ isTelemetryStale ? 'Delay Warning' : 'Active High-Speed Acquisition' }}
                      </span>
                    </div>
                    <div class="flex items-center justify-between text-slate-600 dark:text-slate-300">
                      <span class="text-slate-400">Polling Interval:</span>
                      <span class="font-mono">{{ Math.round(configuredPollingIntervalMs / 1000) }}s Dynamic Cycle</span>
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
                class="absolute right-0 top-full mt-1.5 z-50 w-80 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono"
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
                class="absolute right-0 top-full mt-1.5 z-50 w-96 max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono text-xs max-h-[85vh] overflow-y-auto"
                @click.stop
              >
                <!-- Popover Header -->
                <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
                  <div class="flex items-center space-x-2">
                    <div class="p-1 rounded bg-cyan-500/10 border border-cyan-500/30 text-cyan-500">
                      <OperatorDashboardIcon name="network" :size="15" class="text-cyan-500" />
                    </div>
                    <div>
                      <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                        Network & Delivery Monitor
                      </h3>
                      <div class="text-[9px] text-slate-500 dark:text-slate-400 mt-0.5 leading-none">
                        LAN · Internet WAN · Outbound Links
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1.5">
                    <button
                      type="button"
                      @click="fetchNetworkStatus"
                      class="text-slate-400 hover:text-cyan-500 p-0.5 rounded transition"
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
                      class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Group 1: Local Network Status (LAN) -->
                <div class="p-2.5 rounded-md bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] mb-2">
                  <div class="flex items-center justify-between mb-1.5">
                    <span class="text-[10px] font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center space-x-1">
                      <span>🖥️</span>
                      <span>Local Network (LAN)</span>
                    </span>
                    <span
                      class="text-[9px] font-bold px-1.5 py-0.5 rounded border leading-none"
                      :class="networkStatus.local_network.status === 'UP' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30' : 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30'"
                    >
                      {{ networkStatus.local_network.status === 'UP' ? 'CONNECTED' : 'DOWN' }}
                    </span>
                  </div>
                  <div class="grid grid-cols-2 gap-1.5 text-[9.5px]">
                    <div>
                      <span class="text-slate-400 block text-[9px]">Primary IP & Interface</span>
                      <span class="font-bold text-slate-800 dark:text-slate-200">
                        {{ networkStatus.local_network.primary_ip || '-' }}
                        <span v-if="networkStatus.local_network.primary_interface" class="text-slate-400 text-[8.5px]">
                          ({{ networkStatus.local_network.primary_interface }})
                        </span>
                      </span>
                    </div>
                    <div>
                      <span class="text-slate-400 block text-[9px]">Default Gateway</span>
                      <span class="font-bold text-slate-800 dark:text-slate-200">
                        {{ networkStatus.local_network.gateway || '-' }}
                      </span>
                    </div>
                    <div>
                      <span class="text-slate-400 block text-[9px]">Gateway Status</span>
                      <span
                        class="font-semibold flex items-center space-x-1"
                        :class="networkStatus.local_network.gateway_reachable ? 'text-emerald-500' : 'text-rose-500'"
                      >
                        <span class="inline-block w-1.5 h-1.5 rounded-full" :class="networkStatus.local_network.gateway_reachable ? 'bg-emerald-500' : 'bg-rose-500'"></span>
                        <span>{{ networkStatus.local_network.gateway_reachable ? 'Reachable (' + formatLatency(networkStatus.local_network.gateway_latency_ms) + ')' : 'Unreachable' }}</span>
                      </span>
                    </div>
                    <div>
                      <span class="text-slate-400 block text-[9px]">Active Interfaces</span>
                      <span class="font-medium text-slate-700 dark:text-slate-300">
                        {{ (networkStatus.local_network.interfaces || []).filter(i => i.is_up && i.type !== 'LOOPBACK').length }} interface(s) UP
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Group 2: Internet Connectivity Status (WAN) -->
                <div class="p-2.5 rounded-md bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] mb-2">
                  <div class="flex items-center justify-between mb-1.5">
                    <span class="text-[10px] font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center space-x-1">
                      <span>🌐</span>
                      <span>Internet Connectivity (WAN)</span>
                    </span>
                    <span
                      class="text-[9px] font-bold px-1.5 py-0.5 rounded border leading-none"
                      :class="networkStatus.internet.status === 'CONNECTED' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30' : networkStatus.internet.status === 'DEGRADED' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30' : 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30'"
                    >
                      {{ networkStatus.internet.status || 'UNKNOWN' }}
                    </span>
                  </div>
                  <div class="grid grid-cols-2 gap-1.5 text-[9.5px]">
                    <div>
                      <span class="text-slate-400 block text-[9px]">Probe Target</span>
                      <span class="font-mono text-slate-700 dark:text-slate-300">{{ networkStatus.internet.probe_target || '1.1.1.1:53' }}</span>
                    </div>
                    <div>
                      <span class="text-slate-400 block text-[9px]">External Latency</span>
                      <span
                        class="font-bold tabular-nums"
                        :class="networkStatus.internet.connected ? 'text-cyan-600 dark:text-cyan-400' : 'text-slate-400'"
                      >
                        {{ networkStatus.internet.connected ? formatLatency(networkStatus.internet.latency_ms) : 'Timeout' }}
                      </span>
                    </div>
                  </div>
                  <div v-if="!networkStatus.internet.connected" class="mt-1.5 pt-1.5 border-t border-slate-200/50 dark:border-slate-800/60 text-[9px] text-amber-600 dark:text-amber-400">
                    ⚠️ WAN Offline: Data safely queued locally in autonomous offline storage.
                  </div>
                </div>

                <!-- Group 3: Outbound Delivery Monitor -->
                <div class="p-2.5 rounded-md bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] mb-2.5">
                  <div class="flex items-center justify-between mb-2">
                    <span class="text-[10px] font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center space-x-1">
                      <span>📤</span>
                      <span>Outbound Delivery Links ({{ networkStatus.outbound.total_links }})</span>
                    </span>
                    <span
                      class="text-[9px] font-bold px-1.5 py-0.5 rounded border leading-none"
                      :class="outboundStatusBadgeClass(networkStatus.outbound.overall_status)"
                    >
                      {{ networkStatus.outbound.overall_status || 'UNKNOWN' }}
                    </span>
                  </div>

                  <!-- Queue Summary Counters Bar -->
                  <div class="p-1.5 rounded bg-slate-100 dark:bg-[#0E1726] border border-slate-200/70 dark:border-[#1E3048] mb-2 flex items-center justify-between text-[9px]">
                    <div>
                      <span class="text-slate-400">Queue: </span>
                      <strong :class="((networkStatus.outbound.queue_summary.pending || 0) + (networkStatus.outbound.queue_summary.sending || 0) + (networkStatus.outbound.queue_summary.retrying || 0)) > 0 ? 'text-cyan-500 font-bold' : 'text-slate-700 dark:text-slate-300'">
                        {{ (networkStatus.outbound.queue_summary.pending || 0) + (networkStatus.outbound.queue_summary.sending || 0) + (networkStatus.outbound.queue_summary.retrying || 0) }}
                      </strong>
                    </div>
                    <div>
                      <span class="text-slate-400">Delivered: </span>
                      <strong class="text-emerald-600 dark:text-emerald-400">
                        {{ (networkStatus.outbound.queue_summary.delivered || 0) + (networkStatus.outbound.queue_summary.acknowledged || 0) }}
                      </strong>
                    </div>
                    <div>
                      <span class="text-slate-400">Retrying: </span>
                      <strong :class="(networkStatus.outbound.queue_summary.retrying || 0) > 0 ? 'text-amber-500 font-bold' : 'text-slate-400'">
                        {{ networkStatus.outbound.queue_summary.retrying || 0 }}
                      </strong>
                    </div>
                    <div>
                      <span class="text-slate-400">Failed: </span>
                      <strong :class="(networkStatus.outbound.queue_summary.failed || 0) > 0 ? 'text-rose-500 font-bold' : 'text-slate-400'">
                        {{ networkStatus.outbound.queue_summary.failed || 0 }}
                      </strong>
                    </div>
                  </div>

                  <!-- Destinations List -->
                  <div v-if="networkStatus.outbound.destinations && networkStatus.outbound.destinations.length" class="space-y-1.5 max-h-48 overflow-y-auto pr-0.5">
                    <div
                      v-for="dest in networkStatus.outbound.destinations"
                      :key="dest.id"
                      class="p-2 rounded bg-white dark:bg-[#0A1220] border border-slate-200/90 dark:border-[#1E3048] space-y-1 text-[9.5px]"
                    >
                      <div class="flex items-center justify-between">
                        <div class="flex items-center space-x-1.5 min-w-0">
                          <strong class="text-slate-900 dark:text-white truncate">{{ dest.name }}</strong>
                          <span class="text-[8.5px] px-1 py-0.2 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
                            {{ dest.type }}
                          </span>
                        </div>
                        <span
                          class="text-[8.5px] font-bold px-1.5 py-0.5 rounded border leading-none flex-shrink-0"
                          :class="outboundStatusBadgeClass(dest.status)"
                        >
                          {{ dest.status }}
                        </span>
                      </div>

                      <div class="font-mono text-[8.5px] text-slate-500 dark:text-slate-400 truncate" :title="dest.endpoint">
                        {{ dest.endpoint }}
                      </div>

                      <div class="flex items-center justify-between text-[9px] pt-1 border-t border-slate-100 dark:border-slate-800/80">
                        <span>
                          Success: <strong class="text-emerald-600 dark:text-emerald-400">{{ (dest.success_deliveries || 0).toLocaleString() }}</strong> ·
                          Fail: <strong class="text-rose-500">{{ (dest.failed_deliveries || 0).toLocaleString() }}</strong>
                          <span class="text-slate-400 text-[8px]">({{ (dest.success_rate || 0).toFixed(1) }}%)</span>
                        </span>
                        <span class="text-slate-400">
                          Queue: <strong class="text-slate-700 dark:text-slate-300">{{ dest.queue_count || 0 }}</strong>
                        </span>
                      </div>

                      <div v-if="dest.last_error" class="text-[8px] text-rose-500 dark:text-rose-400 truncate bg-rose-500/5 p-1 rounded border border-rose-500/20" :title="dest.last_error">
                        ⚠️ {{ dest.last_error }}
                      </div>
                    </div>
                  </div>

                  <div v-else class="text-center py-3 text-[10px] text-slate-400 italic">
                    No outbound destinations configured.
                  </div>
                </div>

                <!-- Action Button: Open Outbound Integration -->
                <button
                  type="button"
                  @click="navigateToRoute('/integration')"
                  class="w-full py-1.5 px-2 rounded text-3xs font-semibold transition border flex items-center justify-center space-x-1.5 bg-slate-50 dark:bg-[#0E1726] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:border-cyan-500 hover:text-cyan-600 dark:hover:text-cyan-400"
                >
                  <span>🔗</span>
                  <span>Open Outbound Integration & Delivery (/integration) →</span>
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
                class="absolute right-0 top-full mt-1.5 z-50 w-[420px] max-w-[calc(100vw-24px)] rounded-lg shadow-2xl p-3 text-left border border-slate-200 dark:border-[#20354B] bg-white dark:bg-[#101B2B] text-slate-800 dark:text-slate-100 select-none animate-fadeIn font-mono text-xs max-h-[88vh] overflow-y-auto"
                @click.stop
              >
                <!-- Popover Header -->
                <div class="flex items-center justify-between pb-2 mb-2.5 border-b border-slate-100 dark:border-[#1E3048]">
                  <div class="flex items-center space-x-2">
                    <div class="p-1 rounded bg-cyan-500/10 border border-cyan-500/30 text-cyan-500">
                      <OperatorDashboardIcon name="map-pin" :size="15" class="text-cyan-500" />
                    </div>
                    <div>
                      <h3 class="text-xs font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                        {{ $t('operator.stationLocation') || 'Station Location Monitor' }}
                      </h3>
                      <div class="text-[9px] text-slate-500 dark:text-slate-400 mt-0.5 leading-none">
                        {{ $t('operator.stationLocationSub') || 'Physical Geolocation & Coordinates' }}
                      </div>
                    </div>
                  </div>
                  <div class="flex items-center space-x-1.5">
                    <button
                      type="button"
                      @click="openFullScreenMap"
                      class="text-slate-400 hover:text-cyan-500 p-0.5 rounded transition flex items-center"
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
                      class="text-slate-400 hover:text-cyan-500 p-0.5 rounded transition"
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
                      class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 p-0.5 rounded transition"
                      title="Close"
                    >
                      <OperatorDashboardIcon name="close" :size="14" />
                    </button>
                  </div>
                </div>

                <!-- Feedback Banner -->
                <div
                  v-if="stationLocationFeedback.message"
                  class="p-2 mb-2 rounded text-[9.5px] border flex items-center justify-between"
                  :class="stationLocationFeedback.type === 'error' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30' : 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30'"
                >
                  <span>{{ stationLocationFeedback.message }}</span>
                  <button @click="stationLocationFeedback.message = ''" class="ml-1 opacity-70 hover:opacity-100">✕</button>
                </div>

                <!-- Station Summary Status Card -->
                <div class="p-2.5 rounded-md bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] mb-2">
                  <div class="flex items-center justify-between mb-1.5">
                    <div class="flex items-center space-x-1.5">
                      <span class="text-[10px] font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">
                        {{ stationLocation.station_name || systemStatus.hostname || 'Edge Station Node' }}
                      </span>
                    </div>
                    <span
                      class="text-[9px] font-bold px-1.5 py-0.5 rounded border leading-none"
                      :class="stationLocation.is_configured ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30' : 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border-amber-500/30'"
                    >
                      {{ stationLocation.is_configured ? 'CONFIGURED' : 'UNCONFIGURED' }}
                    </span>
                  </div>

                  <div class="grid grid-cols-2 gap-2 text-[9.5px]">
                    <div>
                      <span class="text-slate-400 block text-[9px]">Coordinates (Lat, Lng)</span>
                      <span class="font-bold tabular-nums text-slate-800 dark:text-slate-200">
                        {{ stationLocation.is_configured && stationLocation.latitude !== null ? `${Number(stationLocation.latitude).toFixed(6)}, ${Number(stationLocation.longitude).toFixed(6)}` : 'Not Configured' }}
                      </span>
                    </div>
                    <div>
                      <span class="text-slate-400 block text-[9px]">Location Source</span>
                      <span class="font-semibold text-cyan-600 dark:text-cyan-400">
                        {{ stationLocation.location_source || 'UNCONFIGURED' }}
                        <span v-if="stationLocation.accuracy_meters" class="text-slate-400 text-[8.5px]">
                          (±{{ stationLocation.accuracy_meters }}m)
                        </span>
                      </span>
                    </div>
                    <div v-if="stationLocation.address" class="col-span-2">
                      <span class="text-slate-400 block text-[9px]">Installation Site / Address</span>
                      <span class="text-slate-700 dark:text-slate-300 truncate block">
                        {{ stationLocation.address }}
                      </span>
                    </div>
                    <div class="col-span-2 flex items-center justify-between text-[8.5px] text-slate-400 pt-1 border-t border-slate-200/50 dark:border-slate-800/60">
                      <span>Updated: {{ formatLocationDate(stationLocation.updated_at) }}</span>
                      <span>By: <strong class="text-slate-600 dark:text-slate-300">{{ stationLocation.updated_by || 'system' }}</strong></span>
                    </div>
                  </div>
                </div>

                <!-- Mini Map Section -->
                <div class="mb-2.5">
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-[9.5px] font-bold text-slate-600 dark:text-slate-400 uppercase tracking-wider">
                      Interactive Station Map
                    </span>
                    <button
                      type="button"
                      @click="openFullScreenMap"
                      class="text-[9px] text-cyan-600 dark:text-cyan-400 hover:underline flex items-center space-x-1"
                    >
                      <span>Full Screen ⛶</span>
                    </button>
                  </div>
                  <StationMapViewer
                    ref="miniMapViewer"
                    :latitude="candidateLocation.latitude !== null ? Number(candidateLocation.latitude) : (stationLocation.latitude !== null ? Number(stationLocation.latitude) : null)"
                    :longitude="candidateLocation.longitude !== null ? Number(candidateLocation.longitude) : (stationLocation.longitude !== null ? Number(stationLocation.longitude) : null)"
                    :accuracy-meters="candidateLocation.accuracy_meters !== null ? Number(candidateLocation.accuracy_meters) : (stationLocation.accuracy_meters !== null ? Number(stationLocation.accuracy_meters) : null)"
                    :editable="true"
                    :source="candidateLocation.location_source || stationLocation.location_source"
                    height="170px"
                    @coordinate-change="onMapCoordinatePicked"
                    @location-selected="onLocationSelected"
                  />
                  <div class="text-[8.5px] text-slate-400 mt-1 italic">
                    💡 Click on the map, use search box, or drag the marker to pinpoint exact station coordinates.
                  </div>
                </div>

                <!-- Three Location Sources Tabs -->
                <div class="border-t border-slate-200 dark:border-[#1E3048] pt-2 mb-2">
                  <div class="flex rounded bg-slate-100 dark:bg-[#0A1220] p-0.5 border border-slate-200 dark:border-[#1E3048] mb-2 text-[9px] font-bold">
                    <button
                      type="button"
                      @click="locationActiveTab = 'manual'"
                      class="flex-1 py-1 rounded transition text-center"
                      :class="locationActiveTab === 'manual' ? 'bg-cyan-500 text-slate-950 font-black shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      ✏️ Manual & Pin
                    </button>
                    <button
                      type="button"
                      @click="locationActiveTab = 'gps'"
                      class="flex-1 py-1 rounded transition text-center"
                      :class="locationActiveTab === 'gps' ? 'bg-cyan-500 text-slate-950 font-black shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      📡 GPS (Browser)
                    </button>
                    <button
                      type="button"
                      @click="locationActiveTab = 'ip'"
                      class="flex-1 py-1 rounded transition text-center"
                      :class="locationActiveTab === 'ip' ? 'bg-cyan-500 text-slate-950 font-black shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                    >
                      🌐 Internet IP
                    </button>
                  </div>

                  <!-- Tab 1: Manual Input & Map Pin -->
                  <div v-if="locationActiveTab === 'manual'" class="space-y-2">
                    <!-- Quick Location Search Bar in Tab -->
                    <div class="p-1.5 rounded bg-slate-100 dark:bg-[#070E1A] border border-slate-200 dark:border-[#1E3048]">
                      <label class="block text-[8.5px] font-bold text-slate-600 dark:text-slate-400 mb-1">
                        🔍 {{ $t('operator.quickLocationSearch') || 'Pencarian Lokasi (Nama Kota / Alamat / Koordinat)' }}
                      </label>
                      <div class="flex items-center space-x-1">
                        <input
                          v-model="tabLocationSearchQuery"
                          type="text"
                          :placeholder="$t('operator.searchLocationPlaceholder') || 'Contoh: Bandung, Cimahi, atau -6.917, 107.619 (Tekan Enter)'"
                          @keydown.enter.prevent="executeTabLocationSearch"
                          class="flex-1 px-2 py-1 rounded bg-white dark:bg-[#0E1726] border border-slate-300 dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-3xs font-mono focus:outline-none focus:border-cyan-500"
                        />
                        <button
                          type="button"
                          @click="executeTabLocationSearch"
                          :disabled="isTabSearching"
                          class="px-2.5 py-1 rounded text-3xs font-bold bg-cyan-600 hover:bg-cyan-500 text-white transition flex items-center space-x-1 disabled:opacity-50"
                        >
                          <span v-if="isTabSearching" class="animate-spin">⏳</span>
                          <span v-else>Cari</span>
                        </button>
                      </div>
                    </div>

                    <div class="grid grid-cols-2 gap-2 text-[9px]">
                      <div>
                        <label class="block text-slate-500 dark:text-slate-400 mb-0.5 font-bold">Latitude (-90 to 90)</label>
                        <input
                          v-model.number="candidateLocation.latitude"
                          type="number"
                          step="0.000001"
                          min="-90"
                          max="90"
                          placeholder="-6.208800"
                          class="w-full px-2 py-1 rounded bg-slate-50 dark:bg-[#070E1A] border border-slate-300 dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-3xs font-mono focus:outline-none focus:border-cyan-500"
                        />
                      </div>
                      <div>
                        <label class="block text-slate-500 dark:text-slate-400 mb-0.5 font-bold">Longitude (-180 to 180)</label>
                        <input
                          v-model.number="candidateLocation.longitude"
                          type="number"
                          step="0.000001"
                          min="-180"
                          max="180"
                          placeholder="106.845600"
                          class="w-full px-2 py-1 rounded bg-slate-50 dark:bg-[#070E1A] border border-slate-300 dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-3xs font-mono focus:outline-none focus:border-cyan-500"
                        />
                      </div>
                      <div class="col-span-2">
                        <label class="block text-slate-500 dark:text-slate-400 mb-0.5">Site Name / Description (Optional)</label>
                        <input
                          v-model="candidateLocation.station_name"
                          type="text"
                          placeholder="Station Central Jakarta AQMS"
                          class="w-full px-2 py-1 rounded bg-slate-50 dark:bg-[#070E1A] border border-slate-300 dark:border-[#20354B] text-slate-800 dark:text-slate-200 text-3xs font-mono focus:outline-none focus:border-cyan-500"
                        />
                      </div>
                    </div>

                    <!-- Action Buttons: Save, Cancel, Reset -->
                    <div class="flex items-center space-x-1.5 pt-1">
                      <button
                        type="button"
                        @click="saveStationLocation"
                        :disabled="isSavingStationLocation || candidateLocation.latitude == null || candidateLocation.longitude == null"
                        class="flex-1 py-1.5 px-2 rounded text-3xs font-bold transition border flex items-center justify-center space-x-1 bg-cyan-600 hover:bg-cyan-500 text-white border-cyan-500 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        <span v-if="isSavingStationLocation" class="animate-spin">⏳</span>
                        <span>💾 Save Location</span>
                      </button>
                      <button
                        type="button"
                        @click="cancelLocationEdit"
                        class="py-1.5 px-2.5 rounded text-3xs font-semibold transition border bg-slate-100 dark:bg-[#0E1726] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538]"
                        title="Revert to saved location"
                      >
                        Cancel
                      </button>
                      <button
                        type="button"
                        @click="resetStationLocation"
                        class="py-1.5 px-2 rounded text-3xs font-semibold transition border bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 border-rose-500/30"
                        title="Reset station coordinates to unconfigured state"
                      >
                        Reset
                      </button>
                    </div>
                  </div>

                  <!-- Tab 2: GPS (Browser Geolocation API) -->
                  <div v-else-if="locationActiveTab === 'gps'" class="space-y-2 text-[9.5px]">
                    <div class="p-2 rounded bg-cyan-500/5 border border-cyan-500/20 text-slate-600 dark:text-slate-300 space-y-1">
                      <div class="font-bold text-cyan-600 dark:text-cyan-400 flex items-center space-x-1">
                        <span>📡</span>
                        <span>Hardware / Browser GPS Geolocation</span>
                      </div>
                      <p class="text-[9px] text-slate-500 dark:text-slate-400 leading-relaxed">
                        Menggunakan Browser Geolocation API dengan izin operator. Hanya dijalankan saat tombol ditekan (tidak melakukan polling latar belakang).
                      </p>
                    </div>

                    <button
                      type="button"
                      @click="detectBrowserGPS"
                      :disabled="gpsState.isDetecting"
                      class="w-full py-1.5 px-2 rounded text-3xs font-bold transition border flex items-center justify-center space-x-1.5 bg-cyan-600/15 hover:bg-cyan-600/25 text-cyan-700 dark:text-cyan-300 border-cyan-500/40"
                    >
                      <span v-if="gpsState.isDetecting" class="animate-spin">⏳</span>
                      <span>🛰️ {{ gpsState.isDetecting ? 'Detecting GPS Satellite / Device...' : 'Ambil Koordinat dari GPS Browser' }}</span>
                    </button>

                    <!-- GPS Result Card -->
                    <div v-if="gpsState.coordinates" class="p-2 rounded bg-slate-50 dark:bg-[#070E1A] border border-cyan-500/30 space-y-1">
                      <div class="flex items-center justify-between text-[9px] font-bold text-emerald-500">
                        <span>✓ Koordinat GPS Berhasil Diperoleh</span>
                        <span>±{{ gpsState.coordinates.accuracy }}m</span>
                      </div>
                      <div class="grid grid-cols-2 gap-1 font-mono text-[9px]">
                        <div>Lat: <strong class="text-slate-800 dark:text-slate-200">{{ gpsState.coordinates.lat }}</strong></div>
                        <div>Lng: <strong class="text-slate-800 dark:text-slate-200">{{ gpsState.coordinates.lng }}</strong></div>
                      </div>
                      <button
                        type="button"
                        @click="applyGPSToCandidate"
                        class="w-full mt-1.5 py-1 px-2 rounded text-3xs font-bold transition bg-cyan-600 hover:bg-cyan-500 text-white"
                      >
                        Gunakan untuk Form Edit (Pratinjau) →
                      </button>
                    </div>

                    <div v-if="gpsState.error" class="p-2 rounded bg-rose-500/10 border border-rose-500/30 text-rose-500 text-[9px]">
                      ⚠️ {{ gpsState.error }}
                    </div>
                  </div>

                  <!-- Tab 3: Internet Geolocation (IP-based) -->
                  <div v-else-if="locationActiveTab === 'ip'" class="space-y-2 text-[9.5px]">
                    <div class="p-2 rounded bg-amber-500/10 border border-amber-500/20 text-slate-600 dark:text-slate-300 space-y-1">
                      <div class="font-bold text-amber-600 dark:text-amber-400 flex items-center space-x-1">
                        <span>⚠️</span>
                        <span>Internet IP Geolocation (Perkiraan Kasar)</span>
                      </div>
                      <p class="text-[9px] text-slate-500 dark:text-slate-400 leading-relaxed">
                        Lokasi berbasis IP hanya merupakan estimasi kasar level kota/ISP dan tidak seakurat GPS. Sesuai kebijakan CBI privat, data IP tidak dikirim ke pihak ketiga tanpa konfigurasi provider terverifikasi.
                      </p>
                    </div>

                    <button
                      type="button"
                      @click="checkIPGeolocationProvider"
                      :disabled="ipGeoState.isDetecting"
                      class="w-full py-1.5 px-2 rounded text-3xs font-bold transition border flex items-center justify-center space-x-1.5 bg-slate-100 dark:bg-[#0E1726] hover:bg-slate-200 dark:hover:bg-[#162538] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-[#20354B]"
                    >
                      <span v-if="ipGeoState.isDetecting" class="animate-spin">⏳</span>
                      <span>🌐 Periksa Provider IP Geolocation</span>
                    </button>

                    <div v-if="ipGeoState.status !== 'IDLE'" class="p-2 rounded bg-slate-50 dark:bg-[#070E1A] border border-slate-200/80 dark:border-[#1E3048] space-y-1">
                      <div class="flex items-center justify-between text-[9px]">
                        <span class="text-slate-400">Status Provider:</span>
                        <span class="font-bold text-amber-500">{{ ipGeoState.status }}</span>
                      </div>
                      <div class="text-[8.5px] text-slate-500 dark:text-slate-400">
                        {{ ipGeoState.message || 'Provider IP Geolocation tidak dikonfigurasi / tidak tersedia. Koordinat tidak dapat diestimasi secara otomatis.' }}
                      </div>
                    </div>
                  </div>
                </div>
              </div>
          </div>
        </div>
      </div>
    </header>

    <!-- =================================================================== -->
    <!-- 2. AUTO-ADAPTIVE MIDDLE BENTO:                                      -->
    <!--    - Wide (>=1280px): Left Full ISPU Hero Card + Right Parameters   -->
    <!--    - Compact (<1280px): Top Dominant ISPU Bar + Full-Width Params   -->
    <!-- =================================================================== -->
    <section
      class="bento-mid"
      :class="{
        'bento-mid-single': !(isAQMSDomain && isIspuBarVisible) || isCompactScreen,
        'is-compact-mid': isCompactScreen
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

      <!-- Parameter Workspace: Category Pills + Quick View Switcher + Dense Cards / Dual Table -->
      <div class="right-params-wrap">
        <div class="cat-bar">
          <div class="cat-pills no-scrollbar">
            <button
              type="button"
              class="cat-pill"
              :class="{ active: selectedCategoryFilter === 'all' && !filterPinnedOnly }"
              @click="selectBentoAllParams"
            >
              <span>Semua ({{ allParameters.length }})</span>
            </button>

            <button
              type="button"
              class="cat-pill"
              :class="{ active: filterPinnedOnly }"
              @click="toggleFilterPinned"
            >
              <svg width="11" height="11" viewBox="0 0 24 24" fill="#F59E0B" stroke="#F59E0B" stroke-width="1.8" stroke-linejoin="round">
                <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
              </svg>
              <span>Favorit ({{ pinnedParametersList.length }})</span>
            </button>

            <button
              v-for="cat in activeCategoriesList"
              :key="'bento-cat-' + cat.id"
              type="button"
              class="cat-pill"
              :class="{ active: !filterPinnedOnly && Number(selectedCategoryFilter) === Number(cat.id) }"
              @click="selectBentoCategory(cat.id)"
            >
              <span>{{ cat.name }}</span>
            </button>
          </div>

          <!-- Right Quick View Mode Switcher: Auto Cards vs 2-Table Side-by-Side vs 1-Table -->
          <div class="view-quick-capsule flex-shrink-0">
            <button
              type="button"
              class="vq-btn"
              :class="{ active: viewMode === 'cards' }"
              @click="setBentoDisplayMode('auto-cards')"
              title="Tampilan Kartu Otomatis (Menyesuaikan lebar layar & jumlah sensor)"
            >
              <OperatorDashboardIcon name="layout" :size="11" />
              <span>{{ bentoAutoGrid ? ('Kartu Otomatis (' + effectiveBentoCols + 'P)') : ('Kartu (' + effectiveBentoCols + 'P)') }}</span>
            </button>
            <button
              type="button"
              class="vq-btn"
              :class="{ active: viewMode === 'table' && bentoTableSplitMode === '2T' }"
              @click="setBentoDisplayMode('2T')"
              title="Tampilan 2 Tabel ke Samping (Hemat Ruang Vertikal)"
            >
              <OperatorDashboardIcon name="menu" :size="11" />
              <span>2 Tabel Samping</span>
            </button>
            <button
              type="button"
              class="vq-btn hidden sm:inline-flex"
              :class="{ active: viewMode === 'table' && bentoTableSplitMode === '1T' }"
              @click="setBentoDisplayMode('1T')"
              title="Tampilan 1 Tabel Penuh"
            >
              <span>1 Tabel</span>
            </button>
          </div>
        </div>

        <!-- =============================================================== -->
        <!-- VIEW MODE A: AUTO-ADAPTIVE BENTO SENSOR CARDS                   -->
        <!-- =============================================================== -->
        <div
          v-if="viewMode === 'cards'"
          class="sensors-scroll-area custom-scroller"
          :class="{ 'is-compact-scroll': isCompactScreen }"
        >
          <div
            v-if="filteredParameters.length > 0"
            class="sensors-grid"
            :class="[
              'bento-cols-' + effectiveBentoCols,
              {
                'is-micro-cards': isMicroCardDensity,
                'is-compact-cards': isCompactScreen
              }
            ]"
          >
            <div
              v-for="param in filteredParameters"
              :key="param.param_id"
              @click="selectParamForTrend(param)"
              @contextmenu.prevent="openParamContextMenu($event, param)"
              class="s-card"
              :class="[
                { 's-card-selected': isParamSelectedForTrend(param) },
                isParamAlarmOrStaleBad(param) ? (param.alarm_status === 'CRITICAL' || param.quality === 'BAD' ? 's-card-fault' : 's-card-warn') : '',
                perfSettings.cardBlinksEnabled && liveBlinks[param.param_id] ? 's-card-blink' : ''
              ]"
              :title="getParamTimingTitle(param)"
            >
              <!-- Card Header -->
              <div class="s-head">
                <div class="s-title-wrap">
                  <span
                    class="s-dot"
                    :style="{ background: (param.device_status === 'ONLINE' || param.device_status === 'CONNECTED') && param.quality !== 'BAD' ? '#10B981' : '#F43F5E' }"
                  ></span>
                  <span class="s-name" :title="param.param_name || param.param_code">
                    {{ param.param_name || param.param_code }}
                  </span>
                </div>

                <div class="s-head-right">
                  <span v-if="!isCompactScreen && getParamCategory(param)" class="s-cat">
                    {{ getParamCategoryShortLabel(param) }}
                  </span>
                  <span class="s-badge" :class="getBentoStatusBadge(param).toneClass">
                    {{ getBentoStatusBadge(param).label }}
                  </span>
                  <button
                    type="button"
                    class="s-icon-btn"
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
                    class="s-icon-btn"
                    @click.stop="openParamContextMenu($event, param)"
                    title="Menu Aksi Parameter (Test Read, Konfigurasi, Riwayat)"
                  >
                    <svg width="10.5" height="10.5" viewBox="0 0 24 24" fill="currentColor">
                      <circle cx="12" cy="5" r="2" />
                      <circle cx="12" cy="12" r="2" />
                      <circle cx="12" cy="19" r="2" />
                    </svg>
                  </button>
                </div>
              </div>

              <!-- Card Body: Discrete Status OR Tabular Analog Reading + Sparkline -->
              <div v-if="isDiscreteStatusParam(param)" class="s-body">
                <div class="flex items-center gap-1.5 min-w-0">
                  <span
                    class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md border font-bold text-[10.5px] uppercase"
                    :class="getDiscreteStatusInfo(param).pillClass"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="getDiscreteStatusInfo(param).dotClass"></span>
                    <span>{{ getDiscreteStatusInfo(param).label }}</span>
                  </span>
                  <span v-if="getDiscreteStatusInfo(param).durationText" class="text-[9.5px] opacity-75 truncate">
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
                  <span class="truncate">{{ isMicroCardDensity ? getBentoCardProgress(param).shortMetaLabel : getBentoCardProgress(param).metaLabel }}</span>
                  <span class="s-meta-pct">{{ getBentoCardProgress(param).rightLabel }}</span>
                </div>
              </div>
            </div>
          </div>

          <div v-else class="bento-empty-cards">
            {{ $t('operator.noParamsMatch') || 'Tidak ada parameter yang sesuai dengan filter saat ini.' }}
          </div>
        </div>

        <!-- =============================================================== -->
        <!-- VIEW MODE B: SIDE-BY-SIDE 2-TABLE & 1-TABLE BENTO VIEW          -->
        <!-- =============================================================== -->
        <div
          v-else
          class="bento-table-area custom-scroller"
          :class="{ 'is-compact-scroll': isCompactScreen }"
        >
          <div
            v-if="filteredParameters.length > 0"
            class="bento-dual-table-grid"
            :class="{ 'is-dual-cols': bentoSplitTableColumns.length === 2 }"
          >
            <div
              v-for="(col, colIdx) in bentoSplitTableColumns"
              :key="'bento-tcol-' + colIdx"
              class="bento-table-card"
            >
              <div class="bento-tbl-subhead">
                <span>{{ col.title }}</span>
                <span>{{ col.params.length }} Kanal</span>
              </div>

              <div class="overflow-x-auto custom-scroller">
                <table class="bento-tbl">
                  <thead>
                    <tr>
                      <th>Parameter</th>
                      <th class="text-right">Nilai Bacaan</th>
                      <th class="text-center">Tren</th>
                      <th>Status &amp; Limit</th>
                      <th v-if="!isCompactScreen || bentoSplitTableColumns.length === 1">Update</th>
                      <th class="text-right">Aksi</th>
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
                        isParamAlarmOrStaleBad(param) ? (param.alarm_status === 'CRITICAL' || param.quality === 'BAD' ? 'is-row-fault' : 'is-row-warn') : ''
                      ]"
                      :title="getParamTimingTitle(param)"
                    >
                      <!-- 1. Parameter Name + Category -->
                      <td>
                        <div class="bt-param-cell">
                          <span
                            class="s-dot"
                            :style="{ background: (param.device_status === 'ONLINE' || param.device_status === 'CONNECTED') && param.quality !== 'BAD' ? '#10B981' : '#F43F5E' }"
                          ></span>
                          <span class="bt-name truncate" :title="param.param_name || param.param_code">
                            {{ param.param_name || param.param_code }}
                          </span>
                          <span v-if="!isCompactScreen && getParamCategory(param)" class="s-cat">
                            {{ getParamCategoryShortLabel(param) }}
                          </span>
                        </div>
                      </td>

                      <!-- 2. Value + Unit -->
                      <td class="text-right whitespace-nowrap">
                        <template v-if="isDiscreteStatusParam(param)">
                          <span
                            class="inline-flex items-center gap-1 px-2 py-0.5 rounded border font-bold text-[9.5px] uppercase"
                            :class="getDiscreteStatusInfo(param).pillClass"
                          >
                            <span class="w-1.5 h-1.5 rounded-full" :class="getDiscreteStatusInfo(param).dotClass"></span>
                            <span>{{ getDiscreteStatusInfo(param).label }}</span>
                          </span>
                        </template>
                        <template v-else>
                          <span class="bt-val" :class="{ 's-val-muted': formatParamValue(param) === '--' }">
                            {{ formatParamValue(param) }}
                          </span>
                          <span v-if="param.unit" class="bt-unit">{{ param.unit }}</span>
                        </template>
                      </td>

                      <!-- 3. Inline Mini Sparkline -->
                      <td class="text-center">
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

                      <!-- 4. Status Pill + Mini Threshold Bar -->
                      <td>
                        <div class="bt-status-cell">
                          <span class="s-badge" :class="getBentoStatusBadge(param).toneClass">
                            {{ getBentoStatusBadge(param).label }}
                          </span>
                          <div class="bt-mini-track">
                            <div
                              class="bt-mini-fill"
                              :style="{ width: getBentoCardProgress(param).pct + '%', background: getBentoCardColor(param) }"
                            ></div>
                          </div>
                          <span v-if="!isCompactScreen" class="bt-pct">
                            {{ getBentoCardProgress(param).rightLabel }}
                          </span>
                        </div>
                      </td>

                      <!-- 5. Freshness (Shown on Wide or 1-Table) -->
                      <td v-if="!isCompactScreen || bentoSplitTableColumns.length === 1" class="whitespace-nowrap">
                        <span class="bt-fresh">
                          {{ param.received_at ? formatRelative(param.received_at) : '-' }}
                        </span>
                      </td>

                      <!-- 6. Favorite + Context Menu Actions -->
                      <td class="text-right whitespace-nowrap" @click.stop>
                        <div class="inline-flex items-center justify-end gap-0.5">
                          <button
                            type="button"
                            class="s-icon-btn"
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
                            class="s-icon-btn"
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
            {{ $t('operator.noParamsMatch') || 'Tidak ada parameter yang sesuai dengan filter saat ini.' }}
          </div>
        </div>
      </div>
    </section>

    <!-- =================================================================== -->
    <!-- 3. BOTTOM ZONE: COMPLETE TREND ANALYSIS STUDIO (ICON-ONLY TOOLS)    -->
    <!-- =================================================================== -->
    <section v-if="perfSettings.trendAnalysisEnabled" class="trend-studio">
      <!-- Row 1: Title, Live Diagnostics, Processed/Raw Switch, Timeframes, Icon Download & Fullscreen -->
      <div class="ts-row1">
        <div class="ts-left flex-wrap">
          <span class="ts-title">Analisis Tren Telemetri</span>

          <span
            class="live-badge"
            :class="$store.state.wsConnected ? 'is-live' : 'is-offline'"
            :title="$store.state.wsConnected ? 'Koneksi WebSocket aktif · Auto-update 1 menit' : 'Koneksi WebSocket terputus'"
          >
            <span class="status-dot" :class="$store.state.wsConnected ? 'dot-emerald' : 'dot-rose'"></span>
            <span>{{ $store.state.wsConnected ? (isCompactScreen ? 'LIVE · 1M' : 'LIVE CONNECTED · AUTO 1M') : 'DISCONNECTED' }}</span>
          </span>

          <span class="diag-pill" :title="trendCommDiagnostics.message">
            <span
              class="w-1.5 h-1.5 rounded-full flex-shrink-0"
              :style="{ background: trendCommDiagnostics.state === 'ONLINE' || trendCommDiagnostics.state === 'RECOVERED' ? '#10B981' : (trendCommDiagnostics.state === 'DEVICE_OFFLINE' || trendCommDiagnostics.state === 'PARTIAL_OFFLINE' || trendCommDiagnostics.state === 'API_ERROR' ? '#F43F5E' : '#F59E0B') }"
            ></span>
            <span>{{ trendCommDiagnostics.badge }}</span>
            <span v-if="!isCompactScreen && trendCommDiagnostics.lastTimeFormatted && trendCommDiagnostics.lastTimeFormatted !== '-'" class="opacity-80">
              · {{ trendCommDiagnostics.lastTimeFormatted }}<template v-if="trendCommDiagnostics.secondsAgo != null"> ({{ trendCommDiagnostics.secondsAgo }}s lalu)</template>
            </span>
          </span>

          <span
            v-if="pipelineDiagnostics && pipelineDiagnostics.sampleCount > 0"
            class="diag-pill diag-pill-accent hidden md:inline-flex"
            :title="'End-to-end latency: ' + pipelineDiagnostics.totalEndToEndMs + 'ms'"
          >
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linejoin="round">
              <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2" />
            </svg>
            <span>{{ (pipelineDiagnostics.lastIntervalMs ? (pipelineDiagnostics.lastIntervalMs / 1000).toFixed(1) : (configuredPollingIntervalMs / 1000).toFixed(0)) }}s · {{ pipelineDiagnostics.totalEndToEndMs }}ms</span>
          </span>

          <button
            v-if="trendCommDiagnostics.state === 'API_ERROR'"
            type="button"
            @click="fetchTrendData(false)"
            class="op-icon-btn sm text-rose-500"
            title="Muat ulang data grafik"
          >
            <OperatorDashboardIcon name="reset" :size="13" />
          </button>

          <button
            v-if="hasOutlierSpike && trendScaleMode === 'auto'"
            type="button"
            @click="toggleTrendSpikeZoom"
            class="diag-pill cursor-pointer"
          >
            <span>⚡ {{ activeSpikeSummary }} [{{ trendSpikeZoom === 'full' ? 'Focus Normal' : 'Full Span' }}]</span>
          </button>

          <!-- Direct 1-Click Processed / Raw Sensor Switch -->
          <div v-if="canAccessRawTrendMode" class="capsule">
            <button
              type="button"
              class="cap-btn"
              :class="{ active: chartMode === 'processed' }"
              @click="setChartMode('processed')"
            >
              Processed
            </button>
            <button
              type="button"
              class="cap-btn"
              :class="{ active: chartMode === 'raw' }"
              @click="setChartMode('raw')"
            >
              Raw
            </button>
          </div>
        </div>

        <div class="ts-right flex-wrap">
          <!-- Timeframe Presets + Periode -->
          <div class="capsule">
            <button
              v-for="tf in ['1m', '5m', '15m', '1h', '6h']"
              :key="'bento-tf-' + tf"
              type="button"
              class="cap-btn"
              :class="{ active: currentTimeframe === tf && !fullscreenIsCustom }"
              @click="setTimeframe(tf)"
            >
              {{ tf }}
            </button>
            <button
              type="button"
              class="cap-btn"
              :class="{ active: fullscreenIsCustom }"
              @click="openCustomTrendModal"
              title="Pilih & Unduh Grafik Periode Khusus"
            >
              <OperatorDashboardIcon name="calendar" :size="11" />
              <span>Periode</span>
            </button>
          </div>

          <!-- Icon-Only Download Graphic / Data Dropdown (Matching Operator Console) -->
          <div class="relative trend-download-anchor">
            <button
              type="button"
              @click.stop="toggleTrendDownloadMenu"
              class="op-icon-btn sm"
              :class="{ 'is-active': showTrendDownloadMenu }"
              title="Unduh Grafik (PNG/SVG/PDF) atau Data (XLS/CSV/JSON)"
              aria-label="Download Chart or Data"
            >
              <OperatorDashboardIcon name="download" :size="14" />
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

          <!-- Icon-Only Expand Fullscreen Chart (Matching Operator Console) -->
          <button
            type="button"
            class="op-icon-btn sm"
            @click="openTrendFullscreen"
            title="Perbesar Grafik Layar Penuh"
            aria-label="Expand Fullscreen Chart"
          >
            <OperatorDashboardIcon name="fullscreen" :size="14" />
          </button>
        </div>
      </div>

      <!-- Row 2: 5 Icon-Only Chart Tools + Plotted & Quick-Add Parameter Chips + Dual-Axis Legend -->
      <div class="ts-row2">
        <div class="ts-row2-left min-w-0 flex-1">
          <div class="mini-tool-group flex-shrink-0">
            <!-- 1. Chart Setup & Skala (Icon Only) -->
            <div class="relative trend-settings-anchor">
              <button
                type="button"
                @click="showTrendSettingsMenu = !showTrendSettingsMenu"
                class="mini-icon-btn"
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

            <!-- 2. Select All (Icon Only) -->
            <button
              type="button"
              @click="selectAllEligibleTrendSeries(true)"
              class="mini-icon-btn"
              title="Plot Semua Parameter"
              aria-label="Select All Parameters"
            >
              <OperatorDashboardIcon name="check-all" :size="13" />
            </button>

            <!-- 3. Clear (Icon Only) -->
            <button
              v-if="activeTrendSeries.length > 0"
              type="button"
              @click="clearAllTrendSeries"
              class="mini-icon-btn"
              title="Kosongkan Pilihan Grafik"
              aria-label="Clear Plotted Series"
            >
              <OperatorDashboardIcon name="clear" :size="13" />
            </button>

            <!-- 4. Reset (Icon Only) -->
            <button
              type="button"
              @click="resetTrendSelectionToDefault"
              class="mini-icon-btn"
              title="Reset ke Pilihan Default"
              aria-label="Reset Default Series"
            >
              <OperatorDashboardIcon name="reset" :size="13" />
            </button>

            <!-- 5. Smooth / Sharp Line Setup (Icon Only) -->
            <div class="relative trend-line-smooth-anchor">
              <button
                type="button"
                @click.stop="toggleLineInterpolationMenu"
                class="mini-icon-btn"
                :class="{ active: showLineSmoothMenu || trendLineSmooth }"
                :title="trendLineSmooth ? 'Mode Garis: Halus / Smooth (Klik untuk ubah)' : 'Mode Garis: Runcing / Linear (Klik untuk ubah)'"
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
          </div>

          <!-- Active Plotted Series Chips + Unplotted Quick-Add Chips -->
          <div class="series-list no-scrollbar">
            <div
              v-for="s in activeTrendSeries"
              :key="'bento-chip-' + s.id"
              @mouseenter="hoveredTrendSeriesId = s.id"
              @mouseleave="hoveredTrendSeriesId = null"
              class="s-chip"
              :class="{ 's-chip-hover': hoveredTrendSeriesId === s.id }"
              :title="(s.paramName || s.paramCode) + (s.latestVal != null && s.latestVal !== '--' ? ' (' + s.latestVal + ' ' + (s.unit || '') + ')' : '')"
            >
              <span class="c-dot" :style="{ background: s.color }"></span>
              <span>{{ getCompactTrendParamName(s) || (s.paramName || s.paramCode) }}</span>
              <strong v-if="s.latestVal != null && s.latestVal !== '--'">{{ s.latestVal }}</strong>
              <button
                type="button"
                @click.stop="removeTrendSeries(s.id)"
                class="chip-remove"
                title="Hapus dari grafik"
              >
                ×
              </button>
            </div>

            <button
              v-for="up in unplottedParameters"
              :key="'bento-unplot-' + up.param_id"
              type="button"
              @click="selectParamForTrend(up)"
              class="s-chip off"
              :title="'Klik untuk menambahkan ' + (up.param_name || up.param_code) + ' ke grafik'"
            >
              <span class="c-dot" :style="{ background: getBentoCardColor(up) }"></span>
              <span>+ {{ getShortParamChipLabel(up) }}</span>
            </button>
          </div>
        </div>

        <!-- Dual-Axis Legend -->
        <div v-if="activeTrendSeries.length > 0" class="axis-legend hidden lg:flex">
          <span v-if="trendScales.left" :style="{ color: trendScales.left.color }">
            Sumbu Kiri: {{ trendScales.left.label }}
          </span>
          <span v-if="trendScales.isDual && trendScales.right" :style="{ color: trendScales.right.color }">
            Sumbu Kanan: {{ trendScales.right.label }}
          </span>
        </div>
      </div>

      <!-- Calm Breathable Dual-Axis SVG Chart -->
      <div
        ref="scadaChartContainer"
        class="bento-chart-canvas"
        :class="{ 'is-compact-canvas': isCompactScreen }"
        @mousemove="onChartScrub"
        @mouseleave="scrubPoint = null"
      >
        <div v-if="isChartLoading" class="bento-chart-overlay">
          <span class="animate-spin w-4 h-4 border-2 border-sky-500 border-t-transparent rounded-full mr-2"></span>
          <span>Memuat data tren telemetri...</span>
        </div>

        <div v-else-if="activeTrendSeries.length === 0" class="bento-chart-empty">
          Pilih kartu/baris parameter di atas atau klik chip <strong>+ Parameter</strong> untuk menampilkan grafik tren.
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
    </section>

    <!-- Trend Analysis Paused Placeholder -->
    <section v-else class="trend-studio !py-3 !flex-row items-center justify-between">
      <div>
        <div class="ts-title">Analisis Tren Telemetri (Dijeda)</div>
        <div class="text-xs opacity-70">Mode hemat performa aktif — grafik SVG dinonaktifkan sementara.</div>
      </div>
      <button type="button" @click="setPerfSetting('trendAnalysisEnabled', true)" class="dock-btn cyan">
        ▶ Aktifkan Grafik Tren
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
                <span v-else class="text-[9px] text-slate-500 italic">Default CBI Icon</span>
              </div>
              <div class="p-3 rounded-lg bg-white border border-slate-200 flex flex-col items-center justify-center min-h-[64px]">
                <span class="text-[8px] text-slate-400 uppercase tracking-widest mb-1">Light Header</span>
                <img
                  v-if="modalCustomerLogoPreviewUrl || customerLogoUrl"
                  :src="modalCustomerLogoPreviewUrl || customerLogoUrl"
                  alt="Customer Logo preview"
                  class="max-h-7 max-w-[120px] object-contain"
                />
                <span v-else class="text-[9px] text-slate-400 italic">Default CBI Icon</span>
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
      :categories="categoriesList"
      @close="isEditParamModalOpen = false"
      @saved="handleParamSaved"
      @open-category-management="isCategoryModalOpen = true"
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
      class="fixed inset-0 z-[100] bg-slate-950/85 backdrop-blur-md flex flex-col p-3 md:p-5 select-none animate-fadeIn font-mono"
      @click.self="closeFullScreenMap"
    >
      <div class="flex-1 w-full max-w-7xl mx-auto flex flex-col bg-white dark:bg-[#0B1422] rounded-xl border border-slate-200 dark:border-[#20354B] shadow-2xl overflow-hidden">
        <!-- Full Screen Header -->
        <div class="flex items-center justify-between px-4 py-3 border-b border-slate-200 dark:border-[#1E3048] bg-slate-50 dark:bg-[#070E1A]">
          <div class="flex items-center space-x-2.5">
            <div class="p-1.5 rounded-lg bg-cyan-500/10 border border-cyan-500/30 text-cyan-500">
              <OperatorDashboardIcon name="map-pinned" :size="20" class="text-cyan-500" />
            </div>
            <div>
              <h2 class="text-sm font-bold uppercase tracking-wider text-slate-900 dark:text-white leading-none">
                Station Location — Full Screen View
              </h2>
              <div class="text-3xs text-slate-500 dark:text-slate-400 mt-1">
                {{ stationLocation.station_name || 'Ambient Monitoring Station' }} ·
                <span class="text-cyan-500 font-bold">
                  {{ candidateLocation.latitude !== null ? `${Number(candidateLocation.latitude).toFixed(6)}, ${Number(candidateLocation.longitude).toFixed(6)}` : 'No coordinates selected' }}
                </span>
              </div>
            </div>
          </div>

          <div class="flex items-center space-x-2">
            <button
              type="button"
              @click="closeFullScreenMap"
              class="px-2.5 py-1 rounded text-2xs font-bold border border-slate-300 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] text-slate-700 dark:text-slate-300 transition flex items-center space-x-1"
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
        <div class="px-4 py-2.5 border-t border-slate-200 dark:border-[#1E3048] bg-slate-50 dark:bg-[#070E1A] flex flex-wrap items-center justify-between gap-2 text-2xs">
          <div class="flex items-center space-x-3 text-slate-600 dark:text-slate-300">
            <div>
              <span class="text-slate-400">Lat: </span>
              <strong class="text-slate-800 dark:text-slate-100 font-mono">{{ candidateLocation.latitude !== null ? Number(candidateLocation.latitude).toFixed(6) : '-' }}</strong>
            </div>
            <div>
              <span class="text-slate-400">Lng: </span>
              <strong class="text-slate-800 dark:text-slate-100 font-mono">{{ candidateLocation.longitude !== null ? Number(candidateLocation.longitude).toFixed(6) : '-' }}</strong>
            </div>
            <div v-if="candidateLocation.accuracy_meters">
              <span class="text-slate-400">Accuracy: </span>
              <strong class="text-cyan-500">±{{ candidateLocation.accuracy_meters }}m</strong>
            </div>
            <div>
              <span class="text-slate-400">Source: </span>
              <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ candidateLocation.location_source || 'MANUAL' }}</span>
            </div>
          </div>

          <div class="flex items-center space-x-2">
            <button
              type="button"
              @click="cancelLocationEdit"
              class="px-3 py-1.5 rounded text-3xs font-semibold border border-slate-300 dark:border-[#20354B] hover:bg-slate-200 dark:hover:bg-[#162538] text-slate-700 dark:text-slate-300 transition"
            >
              Revert
            </button>
            <button
              type="button"
              @click="saveStationLocation"
              :disabled="isSavingStationLocation || candidateLocation.latitude == null || candidateLocation.longitude == null"
              class="px-4 py-1.5 rounded text-3xs font-bold bg-cyan-600 hover:bg-cyan-500 text-white transition disabled:opacity-50 disabled:cursor-not-allowed shadow-sm"
            >
              <span v-if="isSavingStationLocation" class="animate-spin mr-1">⏳</span>
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
  },
  data() {
    let autoGrid = true;
    let splitMode = '2T';
    try {
      const savedAuto = localStorage.getItem('datalogger_bento_auto_grid');
      if (savedAuto === 'false') autoGrid = false;
      const savedSplit = localStorage.getItem('datalogger_bento_table_split');
      if (savedSplit === '1T' || savedSplit === '2T') splitMode = savedSplit;
    } catch (e) {}
    return {
      bentoViewportWidth: typeof window !== 'undefined' ? window.innerWidth : 1600,
      bentoAutoGrid: autoGrid,
      bentoTableSplitMode: splitMode,
    };
  },
  computed: {
    isCompactScreen() {
      return this.bentoViewportWidth < 1280;
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
    bentoSplitTableColumns() {
      const list = this.filteredParameters || [];
      if (this.bentoTableSplitMode === '1T' || list.length <= 2) {
        return [
          {
            title: `DAFTAR PARAMETER TELEMETRI (${list.length} KANAL)`,
            params: list,
          },
        ];
      }
      const mid = Math.ceil(list.length / 2);
      return [
        {
          title: `KOLOM 1 · PARAMETER (#1–${mid})`,
          params: list.slice(0, mid),
        },
        {
          title: `KOLOM 2 · PARAMETER (#${mid + 1}–${list.length})`,
          params: list.slice(mid),
        },
      ];
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
        p => (p.device_status === 'ONLINE' || p.device_status === 'CONNECTED') && p.quality !== 'BAD'
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
    unplottedParameters() {
      const plottedIds = new Set((this.activeTrendSeries || []).map(s => Number(s.parameterId)));
      return (this.allParameters || []).filter(
        p => !this.isDiscreteStatusParam(p) && !plottedIds.has(Number(p.param_id))
      );
    },
  },
  mounted() {
    this.ensureBentoFontsLoaded();
    if (typeof window !== 'undefined') {
      this.bentoViewportWidth = window.innerWidth;
      window.addEventListener('resize', this.onBentoViewportResize, { passive: true });
    }
  },
  beforeDestroy() {
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', this.onBentoViewportResize);
    }
  },
  methods: {
    onBentoViewportResize() {
      if (typeof window === 'undefined') return;
      this.bentoViewportWidth = window.innerWidth;
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
      if (target === 'auto-cards') {
        this.viewMode = 'cards';
        this.bentoAutoGrid = true;
        try {
          localStorage.setItem('datalogger_bento_auto_grid', 'true');
        } catch (e) {}
        return;
      }
      if (target === '2T' || target === '1T') {
        this.viewMode = 'table';
        this.bentoTableSplitMode = target;
        try {
          localStorage.setItem('datalogger_bento_table_split', target);
        } catch (e) {}
      }
    },
    setGridColumns(cols) {
      const count = Math.min(5, Math.max(1, parseInt(cols, 10) || 4));
      this.bentoAutoGrid = false;
      this.viewMode = 'cards';
      this.gridColumns = count;
      this.panelCount = count;
      try {
        localStorage.setItem('datalogger_bento_auto_grid', 'false');
        localStorage.setItem('operator_monitoring_grid_cols', String(count));
        localStorage.setItem('operator_console_grid_columns', String(count));
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
    resolveRegulatoryLimit(param) {
      if (!param) return null;
      if (param.high_limit != null && isFinite(param.high_limit) && Number(param.high_limit) > 0) {
        return Number(param.high_limit);
      }
      const keyStr = ((param.param_code || '') + ' ' + (param.param_name || '')).toUpperCase();
      if (keyStr.includes('PM2.5') || keyStr.includes('PM25') || keyStr.includes('PM 2.5')) return REGULATORY_REF_LIMITS.PM25;
      if (keyStr.includes('PM10') || keyStr.includes('PM 10')) return REGULATORY_REF_LIMITS.PM10;
      if (keyStr.includes('SO2') || keyStr.includes('SULFUR')) return REGULATORY_REF_LIMITS.SO2;
      if (keyStr.includes('NO2') || keyStr.includes('NITROGEN')) return REGULATORY_REF_LIMITS.NO2;
      if (keyStr.includes('O3') || keyStr.includes('OZON')) return REGULATORY_REF_LIMITS.O3;
      if (keyStr.includes('CO') || keyStr.includes('KARBON')) return REGULATORY_REF_LIMITS.CO;
      if (keyStr.includes('HC') || keyStr.includes('HIDROKARBON')) return REGULATORY_REF_LIMITS.HC;
      const unitStr = (param.unit || '').toLowerCase();
      if (unitStr.includes('c') || keyStr.includes('TEMP') || keyStr.includes('SUHU')) return 50;
      if (unitStr === '%' || keyStr.includes('HUMID') || keyStr.includes('LEMBAB')) return 100;
      return null;
    },
    getBentoStatusBadge(param) {
      if (!param) return { label: 'Normal', toneClass: 'badge-green' };
      const isOnline = param.device_status === 'ONLINE' || param.device_status === 'CONNECTED';
      if (!isOnline || param.quality === 'BAD') {
        return { label: 'Offline', toneClass: 'badge-rose' };
      }
      if (param.alarm_status === 'CRITICAL') {
        return { label: 'Kritis', toneClass: 'badge-rose' };
      }
      if (param.alarm_status === 'WARNING') {
        return { label: 'Waspada', toneClass: 'badge-amber' };
      }
      if (param.quality === 'STALE') {
        return { label: 'Stale', toneClass: 'badge-amber' };
      }
      if (this.isDiscreteStatusParam(param)) {
        const info = this.getDiscreteStatusInfo(param);
        return { label: info.label || 'Aktif', toneClass: 'badge-teal' };
      }
      const limit = this.resolveRegulatoryLimit(param);
      const val = Number(param.value);
      if (limit && isFinite(val) && limit > 0) {
        const ratio = val / limit;
        if (ratio >= 1.0) return { label: 'Tinggi', toneClass: 'badge-amber' };
        if (ratio >= 0.5) return { label: 'Sedang', toneClass: 'badge-blue' };
      }
      return { label: 'Baik', toneClass: 'badge-green' };
    },
    getBentoCardColor(param) {
      const dark = this.isDark;
      if (!param) return dark ? '#38BDF8' : '#0284C7';
      const isOnline = param.device_status === 'ONLINE' || param.device_status === 'CONNECTED';
      if (!isOnline || param.quality === 'BAD' || param.alarm_status === 'CRITICAL') {
        return '#F43F5E';
      }
      if (param.alarm_status === 'WARNING' || param.quality === 'STALE') {
        return dark ? '#FBBF24' : '#D97706';
      }
      const plotted = (this.activeTrendSeries || []).find(s => Number(s.parameterId) === Number(param.param_id));
      if (plotted && plotted.color) {
        return plotted.color;
      }
      const badge = this.getBentoStatusBadge(param);
      if (badge.toneClass === 'badge-blue') return dark ? '#38BDF8' : '#0284C7';
      if (badge.toneClass === 'badge-teal') return dark ? '#2DD4BF' : '#0D9488';
      const unitStr = (param.unit || '').toLowerCase();
      if (unitStr.includes('c')) return dark ? '#FBBF24' : '#D97706';
      return dark ? '#34D399' : '#059669';
    },
    getBentoCardProgress(param) {
      if (!param) return { pct: 0, metaLabel: '-', shortMetaLabel: '-', rightLabel: '' };
      const isOnline = (param.device_status === 'ONLINE' || param.device_status === 'CONNECTED') && param.quality !== 'BAD';
      const rel = param.received_at ? this.formatRelative(param.received_at) : '';
      if (!isOnline) {
        return {
          pct: 0,
          metaLabel: rel ? `Terakhir aktif · ${rel}` : 'Sensor tidak terhubung',
          shortMetaLabel: rel || 'Offline',
          rightLabel: 'Offline',
        };
      }
      if (this.isDiscreteStatusParam(param)) {
        const info = this.getDiscreteStatusInfo(param);
        return {
          pct: info && info.isOn ? 100 : 18,
          metaLabel: rel ? `Status Kanal · ${rel}` : 'Status Kanal Aktif',
          shortMetaLabel: rel || 'Kanal Digital',
          rightLabel: info ? info.label : '',
        };
      }
      const limit = this.resolveRegulatoryLimit(param);
      const val = Number(param.value);
      if (limit && isFinite(val) && limit > 0) {
        const pct = Math.min(Math.max(Math.round((val / limit) * 100), 3), 100);
        const unitStr = (param.unit || '').toLowerCase();
        const isMeteo = unitStr.includes('c') || unitStr === '%' || unitStr.includes('hpa') || unitStr.includes('m/s');
        const prefix = isMeteo ? `Rentang 0–${limit}` : `Baku mutu ${limit.toLocaleString('id-ID')}`;
        const shortPrefix = isMeteo ? `0–${limit}` : `Limit ${limit.toLocaleString('id-ID')}`;
        return {
          pct,
          metaLabel: rel ? `${prefix} · ${rel}` : prefix,
          shortMetaLabel: shortPrefix,
          rightLabel: `${pct}%`,
        };
      }
      return {
        pct: isFinite(val) ? 35 : 0,
        metaLabel: rel ? `Update · ${rel}` : 'Online',
        shortMetaLabel: rel || 'Online',
        rightLabel: param.quality || 'GOOD',
      };
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

.right-params-wrap {
  display: flex;
  flex-direction: column;
  gap: 7px;
  min-width: 0;
}

.cat-bar {
  min-height: 26px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.cat-pills {
  display: flex;
  align-items: center;
  gap: 4.5px;
  overflow-x: auto;
  min-width: 0;
}

.cat-pill {
  padding: 3px 9.5px;
  border-radius: 999px;
  font-size: 10.8px;
  font-weight: 600;
  color: var(--b-text-muted);
  background: var(--b-card);
  border: 1px solid var(--b-border);
  display: inline-flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.15s ease;
}
.cat-pill:hover {
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}
.cat-pill.active {
  background: var(--b-primary);
  border-color: var(--b-primary);
  color: #FFFFFF;
}

/* Quick View Mode Switcher Capsule */
.view-quick-capsule {
  display: inline-flex;
  align-items: center;
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 8px;
  padding: 2px;
  gap: 2px;
}

.vq-btn {
  height: 22px;
  padding: 0 7.5px;
  border-radius: 6px;
  font-size: 10.2px;
  font-weight: 700;
  color: var(--b-text-muted);
  display: inline-flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
  cursor: pointer;
  transition: all 0.15s ease;
}
.vq-btn:hover {
  color: var(--b-text-main);
}
.vq-btn.active {
  background: var(--b-primary);
  color: #FFFFFF;
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

.sensors-grid {
  display: grid;
  gap: 8px;
}
.sensors-grid.is-compact-cards {
  gap: 6px;
}

.bento-cols-1 { grid-template-columns: minmax(0, 1fr); }
.bento-cols-2 { grid-template-columns: repeat(2, minmax(0, 1fr)); }
.bento-cols-3 { grid-template-columns: repeat(3, minmax(0, 1fr)); }
.bento-cols-4 { grid-template-columns: repeat(4, minmax(0, 1fr)); }
.bento-cols-5 { grid-template-columns: repeat(5, minmax(0, 1fr)); }

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
.s-card-fault {
  border-color: rgba(244, 63, 94, 0.55) !important;
}
.s-card-warn {
  border-color: rgba(245, 158, 11, 0.55) !important;
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

.s-meta-pct {
  font-weight: 700;
  color: var(--b-text-sec);
  flex-shrink: 0;
}

/* ========================================================================= */
/* 2B. SIDE-BY-SIDE 2-TABLE & 1-TABLE BENTO VIEW                             */
/* ========================================================================= */
.bento-table-area {
  max-height: 332px;
  overflow-y: auto;
  padding-right: 2px;
}
.bento-table-area.is-compact-scroll {
  max-height: 274px;
}

.bento-dual-table-grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: 8px;
  align-items: start;
}
@media (min-width: 768px) {
  .bento-dual-table-grid.is-dual-cols {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

.bento-table-card {
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 12px;
  box-shadow: var(--b-shadow);
  overflow: hidden;
}

.bento-tbl-subhead {
  padding: 5.5px 11px;
  background: var(--b-subtle);
  border-bottom: 1px solid var(--b-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 9.8px;
  font-weight: 800;
  letter-spacing: 0.04em;
  color: var(--b-text-muted);
}

.bento-tbl {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}
.bento-tbl thead th {
  padding: 5px 9px;
  font-size: 9.5px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.03em;
  color: var(--b-text-muted);
  border-bottom: 1px solid var(--b-border-soft);
  white-space: nowrap;
}
.bento-tbl tbody td {
  padding: 5px 9px;
  font-size: 11px;
  border-bottom: 1px solid var(--b-border-soft);
  vertical-align: middle;
}
.is-compact-vp .bento-tbl tbody td {
  padding: 4px 7.5px;
  font-size: 10.5px;
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
  min-width: 0;
}
.bt-name {
  font-weight: 700;
  color: var(--b-text-main);
  max-width: 140px;
}
.is-compact-vp .bt-name {
  max-width: 105px;
}

.bt-val {
  font-weight: 800;
  font-size: 12px;
  color: var(--b-text-main);
}
.is-compact-vp .bt-val {
  font-size: 11.2px;
}
.bt-unit {
  font-size: 9.8px;
  font-weight: 600;
  color: var(--b-text-muted);
  margin-left: 3px;
}

.bt-spark-box {
  width: 50px;
  height: 15px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.is-compact-vp .bt-spark-box {
  width: 42px;
  height: 13px;
}

.bt-status-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}
.bt-mini-track {
  width: 38px;
  height: 3.5px;
  border-radius: 999px;
  background: var(--b-track);
  overflow: hidden;
  flex-shrink: 0;
}
.is-compact-vp .bt-mini-track {
  width: 28px;
}
.bt-mini-fill {
  height: 100%;
  border-radius: 999px;
}
.bt-pct {
  font-size: 9.5px;
  font-weight: 700;
  color: var(--b-text-muted);
}
.bt-fresh {
  font-size: 9.8px;
  color: var(--b-text-muted);
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
  flex: 1;
  background: var(--b-card);
  border: 1px solid var(--b-border);
  border-radius: 13px;
  box-shadow: var(--b-shadow);
  padding: 9px 14px 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 248px;
}
.is-compact-vp .trend-studio {
  padding: 8px 11px 7px;
  min-height: 222px;
}

.ts-row1 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 6px;
  flex-wrap: wrap;
}

.ts-left,
.ts-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.ts-title {
  font-size: 13px;
  font-weight: 800;
  color: var(--b-text-main);
}

.live-badge {
  padding: 2.5px 8px;
  border-radius: 999px;
  font-size: 10px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.is-day .live-badge.is-live { background: #ECFDF5; color: #047857; }
.is-night .live-badge.is-live { background: rgba(16, 185, 129, 0.14); color: #6EE7B7; }
.is-day .live-badge.is-offline { background: #FFF1F2; color: #BE123C; }
.is-night .live-badge.is-offline { background: rgba(244, 63, 94, 0.16); color: #FDA4AF; }

.diag-pill {
  padding: 2.5px 7.5px;
  border-radius: 6px;
  font-size: 10px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border-soft);
  color: var(--b-text-muted);
}
.diag-pill-accent {
  color: var(--b-accent-blue);
}

.ts-row2 {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.ts-row2-left {
  display: flex;
  align-items: center;
  gap: 7px;
}

.mini-tool-group {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 2px;
  border-radius: 7px;
  background: var(--b-pill);
  border: 1px solid var(--b-border);
}

.mini-icon-btn {
  width: 24px;
  height: 22px;
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
.mini-icon-btn.active {
  background: var(--b-active-btn);
  color: var(--b-accent-blue);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
}

.series-list {
  display: flex;
  align-items: center;
  gap: 4.5px;
  overflow-x: auto;
  min-width: 0;
}

.s-chip {
  padding: 2.5px 8px;
  border-radius: 999px;
  font-size: 10.5px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 4.5px;
  background: var(--b-subtle);
  border: 1px solid var(--b-border);
  color: var(--b-text-main);
  white-space: nowrap;
  flex-shrink: 0;
  transition: border-color 0.15s ease;
}
.s-chip-hover {
  border-color: var(--b-accent-blue);
}
.s-chip.off {
  background: transparent;
  color: var(--b-text-muted);
  opacity: 0.84;
  cursor: pointer;
}
.s-chip.off:hover {
  opacity: 1;
  border-color: var(--b-accent-blue);
  color: var(--b-text-main);
}

.c-dot {
  width: 6.5px;
  height: 6.5px;
  border-radius: 50%;
  flex-shrink: 0;
}

.chip-remove {
  color: var(--b-text-faint);
  font-size: 12px;
  line-height: 1;
  margin-left: 1px;
  cursor: pointer;
}
.chip-remove:hover {
  color: #F43F5E;
}

.axis-legend {
  font-size: 10px;
  font-weight: 600;
  gap: 11px;
  white-space: nowrap;
  flex-shrink: 0;
}

.bento-chart-canvas {
  width: 100%;
  height: 205px;
  position: relative;
  overflow: hidden;
  border-radius: 10px;
}
.bento-chart-canvas.is-compact-canvas {
  height: 175px;
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
