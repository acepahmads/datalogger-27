<template>
  <div
    id="superadmin-fleet-map-view"
    @click="showLayerPopover = false"
    class="flex flex-col flex-1 min-h-0 h-full w-full gap-2 select-none font-sans text-slate-800 dark:text-slate-100"
  >
    <!-- ===================================================================== -->
    <!-- 1. DESKTOP COMMAND & KPI RIBBON (>= 1360px wide desktop; hidden on tablet floating glass mode) -->
    <!-- ===================================================================== -->
    <div
      id="fleet-desktop-command-ribbon"
      class="hidden md:flex flex-col gap-1.5 px-2.5 py-1.5 rounded-xl bg-white dark:bg-[#0E1626] border border-slate-200/90 dark:border-slate-800/90 shadow-xs flex-shrink-0"
      :class="{ '!hidden': isTabletFloatingMode }"
    >
      <!-- Row 1: Brand Title + KPI Status Filter Pills (Left) & Map Action Utilities (Right) -->
      <div class="w-full flex items-center justify-between gap-1.5 flex-wrap">
        <div class="flex items-center gap-1.5 flex-wrap min-w-0">
          <div class="flex items-center gap-1.5 pr-2 mr-0.5 border-r border-slate-200 dark:border-slate-800">
            <span class="w-6 h-6 rounded-lg bg-sky-500/10 dark:bg-sky-500/15 border border-sky-500/25 text-sky-600 dark:text-sky-400 flex items-center justify-center flex-shrink-0">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <circle cx="12" cy="12" r="9" />
                <path stroke-linecap="round" stroke-linejoin="round" d="M3.6 9h16.8M3.6 15h16.8M12 3a14.5 14.5 0 000 18M12 3a14.5 14.5 0 010 18" />
              </svg>
            </span>
            <div class="leading-none">
              <div class="text-xs font-extrabold text-slate-900 dark:text-white tracking-tight whitespace-nowrap">Fleet Map</div>
              <div class="text-4xs font-mono text-slate-500 dark:text-slate-400 whitespace-nowrap">{{ summary.total_tenants }} Tenant</div>
            </div>
          </div>

          <!-- KPI Quick Filter Pills -->
          <button
            v-for="pill in kpiPills"
            :key="pill.key"
            type="button"
            :id="'btn-fleet-kpi-' + pill.key"
            @click="setQuickFilter(pill.key)"
            class="inline-flex items-center gap-1.5 px-2 py-1 rounded-lg text-2xs font-bold border transition-all cursor-pointer whitespace-nowrap"
            :class="quickFilter === pill.key ? pill.activeClass : pill.idleClass"
          >
            <span v-if="pill.dotClass" class="w-1.5 h-1.5 rounded-full flex-shrink-0" :class="pill.dotClass"></span>
            <span>{{ pill.label }}</span>
            <span class="font-mono text-3xs px-1 py-0.2 rounded tabular-nums" :class="quickFilter === pill.key ? 'bg-black/15 text-current' : 'bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300'">
              {{ pill.count }}
            </span>
          </button>
        </div>

        <!-- Right-Side Map Quick Actions (Never Clipped) -->
        <div class="flex items-center gap-1.5 ml-auto flex-shrink-0 flex-wrap">
          <!-- Pin Mode Switcher: Status Koneksi vs Indeks Lingkungan (ISPU / IP / IKA) -->
          <div
            id="fleet-pin-mode-switcher"
            class="inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800"
          >
            <button
              type="button"
              id="btn-fleet-pin-mode-status"
              @click="pinDisplayMode = 'status'"
              class="px-2 py-0.5 rounded-md text-2xs font-bold transition cursor-pointer whitespace-nowrap"
              :class="pinDisplayMode === 'status'
                ? 'bg-sky-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
              title="Warna lingkaran berdasarkan Status Koneksi & Alarm"
            >
              Status
            </button>
            <button
              type="button"
              id="btn-fleet-pin-mode-env"
              @click="pinDisplayMode = 'env'"
              class="px-2 py-0.5 rounded-md text-2xs font-bold transition cursor-pointer whitespace-nowrap"
              :class="pinDisplayMode === 'env'
                ? 'bg-emerald-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
              title="Lingkaran angka & warna berdasarkan Indeks ISPU (AQMS) dan IP/IKA (WQMS/ONLIMO)"
            >
              ISPU · IP/IKA
            </button>
          </div>

          <!-- Sub-Switch IP vs IKA for WQMS / ONLIMO when in Environmental Index Mode -->
          <div
            v-if="pinDisplayMode === 'env'"
            id="fleet-wqms-index-switcher"
            class="inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800"
            title="Pilih indeks lingkaran untuk stasiun WQMS / ONLIMO"
          >
            <button
              type="button"
              @click="wqmsIndexMode = 'ip'"
              class="px-1.5 py-0.5 rounded-md text-3xs font-extrabold font-mono transition cursor-pointer"
              :class="wqmsIndexMode === 'ip'
                ? 'bg-teal-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 dark:text-slate-400'"
            >
              IP
            </button>
            <button
              type="button"
              @click="wqmsIndexMode = 'ika'"
              class="px-1.5 py-0.5 rounded-md text-3xs font-extrabold font-mono transition cursor-pointer"
              :class="wqmsIndexMode === 'ika'
                ? 'bg-teal-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 dark:text-slate-400'"
            >
              IKA
            </button>
          </div>

          <!-- Fit Bounds Button -->
          <button
            type="button"
            id="btn-fleet-fit-bounds"
            @click="fitMapToVisibleSites"
            class="h-7 px-2.5 rounded-lg text-2xs font-semibold bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
            title="Fit semua pin stasiun di layar"
          >
            <span>⛶</span>
            <span>Fit Peta</span>
          </button>

          <!-- Fullscreen Map Button (Clean Wall-Display Mode Without Search/Filters) -->
          <button
            type="button"
            id="btn-fleet-fullscreen"
            @click="toggleMapFullscreen"
            class="h-7 px-2.5 rounded-lg text-2xs font-semibold bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
            title="Layar penuh peta tanpa bar filter pencarian (Esc untuk keluar)"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 8V4h4M20 8V4h-4M4 16v4h4M20 16v4h-4" />
            </svg>
            <span>Fullscreen</span>
          </button>

          <!-- Refresh Button -->
          <button
            type="button"
            id="btn-fleet-refresh"
            :disabled="loading"
            @click="fetchFleetOverview(true)"
            class="h-7 px-2 rounded-lg text-2xs font-semibold bg-sky-500/10 hover:bg-sky-500/20 border border-sky-500/30 text-sky-700 dark:text-sky-300 inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
            :title="'Terakhir diperbarui: ' + lastUpdatedLabel"
          >
            <span :class="{ 'animate-spin': loading }">↻</span>
            <span class="hidden xl:inline font-mono text-3xs">{{ lastUpdatedShort }}</span>
          </button>

          <!-- Collapse / Expand Priority Panel -->
          <button
            type="button"
            @click="isRightDockCollapsed = !isRightDockCollapsed"
            class="h-7 px-2.5 rounded-lg text-2xs font-semibold border transition cursor-pointer whitespace-nowrap"
            :class="!isRightDockCollapsed
              ? 'bg-sky-600 text-white btn-text border-sky-600 shadow-2xs'
              : 'bg-slate-50 dark:bg-[#0B0F19] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700/80'"
            title="Tampilkan / sembunyikan panel Prioritas Stasiun"
          >
            {{ isRightDockCollapsed ? '◂ Panel' : 'Panel ▸' }}
          </button>
        </div>
      </div>

      <!-- Row 2: Search + Dropdown Filters (Left) & Segmented Map Style Switcher (Right) -->
      <div class="w-full flex items-center justify-between gap-1.5 flex-wrap pt-1.5 border-t border-slate-100 dark:border-slate-800/80">
        <div class="flex items-center gap-1.5 flex-wrap min-w-0">
          <!-- Search Input -->
          <div class="relative w-48 lg:w-56">
            <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <circle cx="11" cy="11" r="8" />
              <path stroke-linecap="round" d="m21 21-4.3-4.3" />
            </svg>
            <input
              id="input-fleet-search"
              v-model.trim="searchQuery"
              type="text"
              placeholder="Cari site, tenant, kota..."
              class="w-full h-7 pl-7 pr-6 rounded-lg text-2xs bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-800 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:border-sky-500"
            />
            <button
              v-if="searchQuery"
              type="button"
              @click="searchQuery = ''"
              class="absolute right-1.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 text-2xs cursor-pointer"
              title="Hapus pencarian"
            >
              ✕
            </button>
          </div>

          <!-- Project Year Filter -->
          <select
            id="select-fleet-project-year"
            v-model="projectYearFilter"
            class="h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 focus:outline-none focus:border-sky-500 cursor-pointer"
            title="Filter Tahun Project"
          >
            <option value="ALL">Semua Tahun</option>
            <option v-for="yr in projectYearOptions" :key="'f-yr-' + yr.year" :value="String(yr.year)">
              📅 {{ yr.year }} ({{ yr.count }})
            </option>
            <option v-if="unsetProjectYearCount > 0" value="UNSET">
              Belum Diset ({{ unsetProjectYearCount }})
            </option>
          </select>

          <!-- Tenant Filter -->
          <select
            id="select-fleet-tenant"
            v-model="tenantFilter"
            class="h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 max-w-[175px] focus:outline-none focus:border-sky-500 cursor-pointer"
            title="Filter Tenant"
          >
            <option value="ALL">Semua Tenant ({{ displayedTenantOptions.length }})</option>
            <option v-for="t in displayedTenantOptions" :key="t.id" :value="String(t.id)">
              {{ t.name }}{{ t.project_year > 0 ? ` [${t.project_year}]` : '' }} ({{ t.site_count }})
            </option>
          </select>

          <!-- Domain Filter -->
          <select
            id="select-fleet-domain"
            v-model="domainFilter"
            class="h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 focus:outline-none focus:border-sky-500 cursor-pointer"
            title="Filter Domain Stasiun"
          >
            <option value="ALL">Semua Domain</option>
            <option value="AQMS">AQMS</option>
            <option value="WQMS">WQMS</option>
            <option value="SPARING">SPARING</option>
            <option value="ONLIMO">ONLIMO</option>
          </select>

          <!-- Alarm Severity Filter -->
          <select
            id="select-fleet-alarm"
            v-model="alarmFilter"
            class="h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 focus:outline-none focus:border-sky-500 cursor-pointer"
            title="Filter Alarm"
          >
            <option value="ALL">Semua Alarm</option>
            <option value="ANY_ALARM">Ada Alarm ({{ summary.sites_with_alarms }})</option>
            <option value="CRITICAL">Critical Saja</option>
            <option value="WARNING">Warning Saja</option>
            <option value="CLEAN">Normal / Tanpa Alarm</option>
          </select>

          <!-- Reset Filter Pill (Visible when any filter is active) -->
          <button
            v-if="hasActiveFilters"
            type="button"
            @click="resetAllFilters"
            class="h-7 px-2 rounded-lg text-2xs font-bold bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/30 transition cursor-pointer whitespace-nowrap"
            title="Reset semua filter"
          >
            Reset
          </button>
        </div>

        <!-- AQMS Weather Overlays (Angin & Hujan) + Map Theme Segmented Control -->
        <div class="flex items-center gap-1.5 ml-auto flex-shrink-0">
          <div
            v-if="isAqmsWeatherToggleVisible"
            id="fleet-aqms-weather-toggles"
            class="inline-flex items-center gap-1"
          >
            <button
              type="button"
              id="btn-fleet-toggle-wind"
              @click="showWindLayer = !showWindLayer"
              class="h-7 px-2 rounded-lg text-2xs font-bold border inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
              :class="showWindLayer
                ? 'bg-cyan-600 text-white btn-text border-cyan-600 shadow-2xs'
                : 'bg-slate-50 hover:bg-cyan-50 dark:bg-[#0B0F19] dark:hover:bg-cyan-950/40 text-slate-700 dark:text-slate-200 border-slate-200 dark:border-slate-700/80'"
              :title="showWindLayer ? 'Sembunyikan arah & kecepatan angin AQMS di peta' : 'Tampilkan arah & kecepatan angin AQMS langsung di peta'"
            >
              <span>🌬</span>
              <span>Angin</span>
              <span
                v-if="aqmsWindAvailableCount > 0"
                class="px-1 rounded text-4xs font-mono font-extrabold tabular-nums"
                :class="showWindLayer ? 'bg-white/20 text-white badge-text font-extrabold' : 'bg-cyan-500/15 text-cyan-700 dark:text-cyan-300'"
              >{{ aqmsWindAvailableCount }}</span>
            </button>
            <button
              type="button"
              id="btn-fleet-toggle-rain"
              @click="showRainLayer = !showRainLayer"
              class="h-7 px-2 rounded-lg text-2xs font-bold border inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
              :class="showRainLayer
                ? 'bg-blue-600 text-white btn-text border-blue-600 shadow-2xs'
                : 'bg-slate-50 hover:bg-blue-50 dark:bg-[#0B0F19] dark:hover:bg-blue-950/40 text-slate-700 dark:text-slate-200 border-slate-200 dark:border-slate-700/80'"
              :title="showRainLayer ? 'Sembunyikan overlay hujan AQMS di peta' : 'Tampilkan lokasi hujan AQMS langsung di peta'"
            >
              <span>🌧</span>
              <span>Hujan</span>
              <span
                v-if="aqmsRainingCount > 0"
                class="px-1 rounded text-4xs font-mono font-extrabold tabular-nums"
                :class="showRainLayer ? 'bg-white/20 text-white badge-text font-extrabold' : 'bg-blue-600 text-white badge-text font-extrabold animate-pulse'"
              >{{ aqmsRainingCount }}</span>
            </button>
          </div>

          <button
            type="button"
            id="btn-fleet-map-style"
            @click="cycleMapStyle"
            class="xl:hidden h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
            :title="'Tema Peta: ' + mapStyleLabel"
          >
            <span>🗺️</span>
            <span>{{ mapStyleLabel }}</span>
          </button>
          <div class="hidden xl:inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800">
            <button
              v-for="opt in mapStyleOptions"
              :key="'desk-style-' + opt.value"
              type="button"
              @click="setMapStyle(opt.value)"
              class="px-2 py-0.5 rounded-md text-3xs font-bold transition cursor-pointer whitespace-nowrap"
              :class="mapStyle === opt.value
                ? 'bg-sky-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
            >
              {{ opt.label }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- ===================================================================== -->
    <!-- 2. MOBILE / HP COMPACT TOP CONTROL BAR (< 768px)                      -->
    <!-- ===================================================================== -->
    <div
      id="fleet-mobile-header"
      class="flex md:hidden flex-col gap-1.5 p-1.5 rounded-2xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-md border border-slate-200/90 dark:border-slate-800/90 shadow-xs flex-shrink-0"
    >
      <!-- Row 1: Search + Filter Drawer Button + Refresh + Segmented Switcher (Peta | Daftar) -->
      <div class="flex items-center gap-1.5">
        <div class="relative flex-1 min-w-0">
          <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <circle cx="11" cy="11" r="8" />
            <path stroke-linecap="round" d="m21 21-4.3-4.3" />
          </svg>
          <input
            v-model.trim="searchQuery"
            type="text"
            placeholder="Cari stasiun / tenant..."
            class="w-full h-8 pl-7 pr-6 rounded-xl text-xs bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-800 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:border-sky-500"
          />
          <button
            v-if="searchQuery"
            type="button"
            @click="searchQuery = ''"
            class="absolute right-1.5 top-1/2 -translate-y-1/2 text-slate-400 text-xs"
          >
            ✕
          </button>
        </div>

        <!-- Mobile Filter Sheet Button -->
        <button
          id="btn-fleet-mobile-filter"
          type="button"
          @click="showMobileFilterSheet = true"
          class="relative h-8 px-2.5 rounded-xl border text-xs font-bold inline-flex items-center justify-center gap-1 cursor-pointer flex-shrink-0"
          :class="activeSecondaryFilterCount > 0
            ? 'bg-sky-600 text-white btn-text border-sky-600'
            : 'bg-slate-50 dark:bg-[#0B0F19] text-slate-700 dark:text-slate-200 border-slate-200 dark:border-slate-700/80'"
          title="Filter Tenant & Domain"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 4h18M6 12h12m-9 8h6" />
          </svg>
          <span v-if="activeSecondaryFilterCount > 0" class="text-3xs font-mono">{{ activeSecondaryFilterCount }}</span>
        </button>

        <!-- Mobile Refresh Button -->
        <button
          type="button"
          :disabled="loading"
          @click="fetchFleetOverview(true)"
          class="h-8 w-8 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 flex items-center justify-center text-xs font-bold cursor-pointer flex-shrink-0"
          title="Segarkan Data"
        >
          <span :class="{ 'animate-spin': loading }">↻</span>
        </button>

        <!-- Segmented View Switcher: Peta | Daftar -->
        <div
          id="fleet-mobile-view-switcher"
          class="inline-flex items-center p-0.5 rounded-xl bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 flex-shrink-0"
        >
          <button
            type="button"
            id="btn-mobile-mode-map"
            @click="switchMobileView('map')"
            class="px-2.5 py-1 rounded-lg text-2xs font-bold transition cursor-pointer"
            :class="mobileViewMode === 'map'
              ? 'bg-sky-600 text-white btn-text shadow-xs'
              : 'text-slate-600 dark:text-slate-400'"
          >
            🗺️ Peta
          </button>
          <button
            type="button"
            id="btn-mobile-mode-list"
            @click="switchMobileView('list')"
            class="px-2.5 py-1 rounded-lg text-2xs font-bold transition cursor-pointer"
            :class="mobileViewMode === 'list'
              ? 'bg-sky-600 text-white btn-text shadow-xs'
              : 'text-slate-600 dark:text-slate-400'"
          >
            📋 Daftar ({{ filteredSites.length }})
          </button>
        </div>
      </div>

      <!-- Row 2 (Only inline in Daftar mode; in Peta mode chips float directly on map!) -->
      <div v-if="mobileViewMode === 'list'" class="flex items-center gap-1.5 overflow-x-auto pb-0.5 pt-1 border-t border-slate-100 dark:border-slate-800/80 fleet-no-scrollbar">
        <button
          v-for="pill in kpiPills"
          :key="'mob-' + pill.key"
          type="button"
          @click="setQuickFilter(pill.key)"
          class="inline-flex items-center gap-1 px-2 py-1 rounded-full text-2xs font-bold border whitespace-nowrap flex-shrink-0 cursor-pointer"
          :class="quickFilter === pill.key ? pill.activeClass : pill.idleClass"
        >
          <span v-if="pill.dotClass" class="w-1.5 h-1.5 rounded-full" :class="pill.dotClass"></span>
          <span>{{ pill.label }}</span>
          <span class="font-mono text-3xs opacity-90">{{ pill.count }}</span>
        </button>
      </div>
    </div>

    <!-- Error Toast / Banner -->
    <div
      v-if="errorMessage"
      class="px-3 py-2 rounded-xl bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-800/70 text-rose-700 dark:text-rose-300 text-xs flex items-center justify-between flex-shrink-0"
    >
      <span>⚠️ {{ errorMessage }}</span>
      <button type="button" @click="errorMessage = ''" class="font-bold ml-2 cursor-pointer">✕</button>
    </div>

    <!-- ===================================================================== -->
    <!-- 3. MAIN SPLIT WORKSPACE (MAP CANVAS + RESIZABLE DOCKED PRIORITY LIST) -->
    <!-- ===================================================================== -->
    <div
      class="flex-1 min-h-0 flex gap-2.5 md:gap-0 relative overflow-hidden"
      :class="{ 'select-none cursor-col-resize': isResizingDock }"
    >
      <!-- 3A. INTERACTIVE LEAFLET MAP PANE -->
      <div
        v-show="isDesktopViewport || mobileViewMode === 'map' || isMapFullscreen"
        ref="fleetMapPane"
        id="fleet-map-pane"
        class="flex-1 min-w-0 h-full relative rounded-2xl overflow-hidden border border-slate-200/90 dark:border-slate-800/90 bg-slate-100 dark:bg-[#0B0F19] shadow-xs"
        :class="{ 'fleet-map-pane-fullscreen': isMapFullscreen }"
      >
        <!-- Leaflet Container -->
        <div ref="fleetMapCanvas" id="fleet-leaflet-map" class="w-full h-full z-10"></div>

        <!-- =============================================================== -->
        <!-- TABLET / COMPACT LANDSCAPE FLOATING GLASS HUD (Map-First UI)    -->
        <!-- =============================================================== -->
        <div
          v-if="isTabletFloatingMode && !isMapFullscreen"
          id="fleet-tablet-floating-hud"
          class="hidden md:flex items-start justify-between gap-2 absolute top-2.5 inset-x-2.5 z-[470] pointer-events-none"
        >
          <!-- Left: Floating Glass Search + Filter Pill -->
          <div class="pointer-events-auto flex items-center gap-1 p-1 rounded-full bg-white/90 dark:bg-[#0E1626]/90 backdrop-blur-md border border-slate-200/90 dark:border-slate-700/80 shadow-lg max-w-[260px] lg:max-w-[300px]">
            <div class="relative flex-1 min-w-0">
              <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <circle cx="11" cy="11" r="8" />
                <path stroke-linecap="round" d="m21 21-4.3-4.3" />
              </svg>
              <input
                v-model.trim="searchQuery"
                type="text"
                placeholder="Cari site, tenant, kota..."
                class="w-full h-7 pl-7 pr-5 rounded-full text-2xs bg-transparent text-slate-800 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none"
              />
              <button
                v-if="searchQuery"
                type="button"
                @click="searchQuery = ''"
                class="absolute right-1 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 text-3xs cursor-pointer"
                title="Hapus pencarian"
              >
                ✕
              </button>
            </div>
            <button
              id="btn-fleet-tablet-filter"
              type="button"
              @click="showMobileFilterSheet = true"
              class="h-7 px-2.5 rounded-full text-2xs font-bold inline-flex items-center gap-1 border transition cursor-pointer flex-shrink-0"
              :class="activeSecondaryFilterCount > 0
                ? 'bg-sky-600 text-white btn-text border-sky-600 shadow-xs'
                : 'bg-slate-100/90 hover:bg-slate-200/80 dark:bg-slate-800/90 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 border-slate-200/80 dark:border-slate-700'"
              title="Filter Tahun, Tenant, Domain & Alarm"
            >
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M3 5h18M6 12h12m-9 7h6" />
              </svg>
              <span>Filter</span>
              <span v-if="activeSecondaryFilterCount > 0" class="px-1 rounded-full bg-black/20 font-mono text-4xs">{{ activeSecondaryFilterCount }}</span>
            </button>
          </div>

          <!-- Center: Floating Glass Status Chips (Dual-Purpose Filter + Legend — Zero Duplicate Legend Below!) -->
          <div class="pointer-events-auto flex items-center gap-1 p-1 rounded-full bg-white/90 dark:bg-[#0E1626]/90 backdrop-blur-md border border-slate-200/90 dark:border-slate-700/80 shadow-lg overflow-x-auto fleet-no-scrollbar max-w-[44%]">
            <button
              v-for="pill in floatingStatusPills"
              :key="'flt-' + pill.key"
              type="button"
              @click="setQuickFilter(pill.key)"
              class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-2xs font-bold transition cursor-pointer whitespace-nowrap flex-shrink-0"
              :class="quickFilter === pill.key
                ? pill.activeClass
                : (pill.key === 'ALARM' && pill.count > 0
                  ? 'bg-rose-500/12 hover:bg-rose-500/20 text-rose-700 dark:text-rose-300'
                  : 'hover:bg-slate-100/90 dark:hover:bg-slate-800/80 text-slate-700 dark:text-slate-200')"
            >
              <span v-if="pill.dotClass" class="w-2 h-2 rounded-full flex-shrink-0" :class="pill.dotClass"></span>
              <span>{{ pill.label }}</span>
              <span class="font-mono text-3xs tabular-nums opacity-85">{{ pill.count }}</span>
            </button>
          </div>

          <!-- Right: Floating Glass Mode, Layer Popover, Fit, Fullscreen & Station Drawer Pill -->
          <div class="pointer-events-auto flex flex-col items-end gap-1.5 flex-shrink-0">
            <div class="flex items-center gap-1 p-1 rounded-full bg-white/90 dark:bg-[#0E1626]/90 backdrop-blur-md border border-slate-200/90 dark:border-slate-700/80 shadow-lg">
              <!-- Status | ISPU Segmented Switch -->
              <div class="inline-flex items-center p-0.5 rounded-full bg-slate-100/90 dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800">
                <button
                  type="button"
                  @click="pinDisplayMode = 'status'"
                  class="px-2 py-0.5 rounded-full text-3xs font-bold transition cursor-pointer whitespace-nowrap"
                  :class="pinDisplayMode === 'status' ? 'bg-sky-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-400'"
                >
                  Status
                </button>
                <button
                  type="button"
                  @click="pinDisplayMode = 'env'"
                  class="px-2 py-0.5 rounded-full text-3xs font-bold transition cursor-pointer whitespace-nowrap"
                  :class="pinDisplayMode === 'env' ? 'bg-emerald-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-400'"
                >
                  ISPU
                </button>
              </div>

              <!-- IP | IKA Sub-switch -->
              <div v-if="pinDisplayMode === 'env'" class="inline-flex items-center p-0.5 rounded-full bg-slate-100/90 dark:bg-[#0B0F19] border border-slate-200/80 dark:border-slate-800">
                <button
                  type="button"
                  @click="wqmsIndexMode = 'ip'"
                  class="px-1.5 py-0.5 rounded-full text-4xs font-extrabold font-mono cursor-pointer"
                  :class="wqmsIndexMode === 'ip' ? 'bg-teal-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-400'"
                >IP</button>
                <button
                  type="button"
                  @click="wqmsIndexMode = 'ika'"
                  class="px-1.5 py-0.5 rounded-full text-4xs font-extrabold font-mono cursor-pointer"
                  :class="wqmsIndexMode === 'ika' ? 'bg-teal-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-400'"
                >IKA</button>
              </div>

              <!-- Map Layer Popover Button -->
              <div class="relative">
                <button
                  id="btn-fleet-tablet-layers"
                  type="button"
                  @click.stop="showLayerPopover = !showLayerPopover"
                  class="relative w-7 h-7 rounded-full flex items-center justify-center text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer"
                  :class="{ 'bg-sky-500/15 text-sky-600 dark:text-sky-400': showLayerPopover || showWindLayer || showRainLayer }"
                  :title="'Layer Peta: ' + mapStyleLabel"
                >
                  <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 3 2 8.5 12 14l10-5.5L12 3Zm0 14.5-10-5.5m20 0-10 5.5m0 3.5-10-5.5m20 0-10 5.5" />
                  </svg>
                  <span
                    v-if="showWindLayer || showRainLayer"
                    class="absolute top-1 right-1 w-1.5 h-1.5 rounded-full bg-cyan-500"
                  ></span>
                </button>
                <div
                  v-if="showLayerPopover"
                  @click.stop
                  class="absolute right-0 top-full mt-1.5 w-44 p-1.5 rounded-xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-md border border-slate-200 dark:border-slate-700 shadow-2xl z-[520] space-y-0.5"
                >
                  <div v-if="isAqmsWeatherToggleVisible" class="pb-1.5 mb-1 border-b border-slate-100 dark:border-slate-800 space-y-1">
                    <div class="px-2 pt-0.5 text-4xs font-extrabold uppercase tracking-wider text-slate-400">
                      Cuaca AQMS
                    </div>
                    <button
                      type="button"
                      @click="showWindLayer = !showWindLayer"
                      class="w-full px-2.5 py-1.5 rounded-lg text-left text-2xs font-bold flex items-center justify-between transition cursor-pointer"
                      :class="showWindLayer
                        ? 'bg-cyan-600 text-white btn-text shadow-2xs'
                        : 'text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800'"
                    >
                      <span>🌬 Arah & Kec. Angin</span>
                      <span class="text-3xs font-mono">{{ showWindLayer ? 'ON' : 'OFF' }}</span>
                    </button>
                    <button
                      type="button"
                      @click="showRainLayer = !showRainLayer"
                      class="w-full px-2.5 py-1.5 rounded-lg text-left text-2xs font-bold flex items-center justify-between transition cursor-pointer"
                      :class="showRainLayer
                        ? 'bg-blue-600 text-white btn-text shadow-2xs'
                        : 'text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800'"
                    >
                      <span>🌧 Curah Hujan</span>
                      <span class="text-3xs font-mono">{{ showRainLayer ? 'ON' : 'OFF' }}</span>
                    </button>
                  </div>
                  <button
                    v-for="opt in mapStyleOptions"
                    :key="'tab-layer-' + opt.value"
                    type="button"
                    @click="setMapStyle(opt.value); showLayerPopover = false"
                    class="w-full px-2.5 py-1.5 rounded-lg text-left text-2xs font-bold flex items-center justify-between transition cursor-pointer"
                    :class="mapStyle === opt.value
                      ? 'bg-sky-600 text-white btn-text shadow-2xs'
                      : 'text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800'"
                  >
                    <span>{{ opt.label }}</span>
                    <span v-if="mapStyle === opt.value" class="text-3xs">✓</span>
                  </button>
                </div>
              </div>

              <!-- Fit Map -->
              <button
                type="button"
                @click="fitMapToVisibleSites"
                class="w-7 h-7 rounded-full flex items-center justify-center text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer text-xs"
                title="Fit semua pin stasiun"
              >
                ⛶
              </button>

              <!-- Refresh -->
              <button
                type="button"
                :disabled="loading"
                @click="fetchFleetOverview(true)"
                class="w-7 h-7 rounded-full flex items-center justify-center text-sky-600 dark:text-sky-400 hover:bg-sky-500/10 transition cursor-pointer text-xs font-bold"
                :title="'Segarkan · ' + lastUpdatedLabel"
              >
                <span :class="{ 'animate-spin': loading }">↻</span>
              </button>

              <!-- Fullscreen -->
              <button
                type="button"
                @click="toggleMapFullscreen"
                class="w-7 h-7 rounded-full flex items-center justify-center text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition cursor-pointer"
                title="Layar penuh peta"
              >
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M4 8V4h4M20 8V4h-4M4 16v4h4M20 16v4h-4" />
                </svg>
              </button>
            </div>

            <!-- Prominent Floating Station List Drawer Trigger -->
            <button
              id="btn-fleet-tablet-drawer-toggle"
              type="button"
              @click="isRightDockCollapsed = !isRightDockCollapsed"
              class="h-8 px-3.5 rounded-full text-2xs font-extrabold inline-flex items-center gap-1.5 shadow-lg transition cursor-pointer whitespace-nowrap border"
              :class="!isRightDockCollapsed
                ? 'bg-slate-800 text-white btn-text border-slate-700'
                : 'bg-sky-600 hover:bg-sky-500 text-white btn-text border-sky-500'"
              title="Buka / Tutup Daftar Prioritas Stasiun"
            >
              <span>{{ isRightDockCollapsed ? 'Daftar Stasiun' : 'Tutup Daftar' }}</span>
              <span v-if="isRightDockCollapsed && issueFilteredSites.length > 0" class="px-1.5 py-0.2 rounded-full bg-white/20 font-mono text-4xs">
                {{ issueFilteredSites.length }} Masalah
              </span>
              <span v-else-if="isRightDockCollapsed" class="px-1.5 py-0.2 rounded-full bg-white/20 font-mono text-4xs">
                {{ filteredSites.length }}
              </span>
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
                <path v-if="isRightDockCollapsed" stroke-linecap="round" stroke-linejoin="round" d="M4 6h16M4 12h16M4 18h10" />
                <path v-else stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
        </div>

        <!-- Mobile Floating Glass Status Chips (over top of map in Peta mode) -->
        <div
          v-if="!isDesktopViewport && mobileViewMode === 'map' && !isMapFullscreen"
          id="fleet-mobile-floating-chips"
          class="flex md:hidden items-center gap-1.5 absolute top-2 left-2 right-12 z-[450] overflow-x-auto fleet-no-scrollbar py-0.5"
        >
          <button
            v-for="pill in floatingStatusPills"
            :key="'mflt-' + pill.key"
            type="button"
            @click="setQuickFilter(pill.key)"
            class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full text-2xs font-bold border shadow-md backdrop-blur-md whitespace-nowrap flex-shrink-0 cursor-pointer"
            :class="quickFilter === pill.key
              ? pill.activeClass
              : 'bg-white/90 dark:bg-[#0E1626]/90 text-slate-700 dark:text-slate-200 border-slate-200/90 dark:border-slate-700/80'"
          >
            <span v-if="pill.dotClass" class="w-1.5 h-1.5 rounded-full" :class="pill.dotClass"></span>
            <span>{{ pill.label }}</span>
            <span class="font-mono text-3xs opacity-85">{{ pill.count }}</span>
          </button>
        </div>

        <!-- Mobile Bottom Sheet Peek Pill (in Peta mode when no site card is open) -->
        <div
          v-if="!isDesktopViewport && mobileViewMode === 'map' && !selectedSite && !isMapFullscreen"
          id="fleet-mobile-bottom-peek"
          @click="switchMobileView('list')"
          class="flex md:hidden flex-col gap-1 absolute inset-x-2.5 bottom-2.5 z-[460] px-3.5 py-2.5 rounded-2xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-md border border-slate-200/90 dark:border-slate-700/80 shadow-2xl cursor-pointer active:scale-[0.99] transition"
        >
          <div class="w-9 h-1 rounded-full bg-slate-300 dark:bg-slate-700 mx-auto mb-0.5"></div>
          <div class="flex items-center justify-between gap-2">
            <div class="flex items-center gap-1.5 text-xs font-extrabold text-slate-900 dark:text-white min-w-0">
              <span v-if="issueFilteredSites.length > 0" class="text-rose-500">▲</span>
              <span v-else class="text-emerald-500">●</span>
              <span class="truncate">{{ filteredSites.length }} Stasiun · {{ issueFilteredSites.length }} Masalah</span>
            </div>
            <span class="text-2xs font-bold text-sky-600 dark:text-sky-400 whitespace-nowrap">Lihat Daftar →</span>
          </div>
          <div v-if="dockDisplayedSites.length > 0" class="flex items-center justify-between gap-2 text-3xs text-slate-500 dark:text-slate-400">
            <span class="truncate font-semibold text-slate-700 dark:text-slate-300">{{ dockDisplayedSites[0].site_name }}</span>
            <span
              v-if="dockDisplayedSites[0].total_active_alarms > 0"
              class="px-1.5 py-0.2 rounded-full bg-rose-600 text-white badge-text font-mono text-4xs font-extrabold flex-shrink-0"
            >
              ● {{ dockDisplayedSites[0].total_active_alarms }}
            </span>
          </div>
        </div>

        <!-- Clean Fullscreen Top-Right Mini Controls (Zero Search / Zero Dropdown Filters) -->
        <div
          v-if="isMapFullscreen"
          id="fleet-fullscreen-hud"
          class="flex items-center gap-1.5 absolute top-3 right-3 z-[520] px-2 py-1.5 rounded-xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-md border border-slate-200 dark:border-slate-700 shadow-xl"
        >
          <!-- Compact Read-Only Live Status Summary -->
          <div class="hidden md:flex items-center gap-2 px-2 py-0.5 mr-0.5 border-r border-slate-200 dark:border-slate-700 text-2xs font-bold">
            <span class="text-slate-800 dark:text-white font-extrabold">🌐 Fleet Map</span>
            <span class="inline-flex items-center gap-1 text-emerald-600 dark:text-emerald-400 font-mono">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>{{ summary.online_sites || 0 }}
            </span>
            <span class="inline-flex items-center gap-1 text-amber-600 dark:text-amber-400 font-mono">
              <span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>{{ summary.delayed_sites || 0 }}
            </span>
            <span class="inline-flex items-center gap-1 text-rose-600 dark:text-rose-400 font-mono">
              <span class="w-1.5 h-1.5 rounded-full bg-rose-500"></span>{{ summary.offline_sites || 0 }}
            </span>
          </div>

          <!-- Pin Mode Switcher in Fullscreen (Status vs ISPU · IP/IKA) -->
          <div class="inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800">
            <button
              type="button"
              @click="pinDisplayMode = 'status'"
              class="px-2 py-0.5 rounded-md text-3xs font-bold transition cursor-pointer whitespace-nowrap"
              :class="pinDisplayMode === 'status'
                ? 'bg-sky-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 dark:text-slate-400'"
            >
              Status
            </button>
            <button
              type="button"
              @click="pinDisplayMode = 'env'"
              class="px-2 py-0.5 rounded-md text-3xs font-bold transition cursor-pointer whitespace-nowrap"
              :class="pinDisplayMode === 'env'
                ? 'bg-emerald-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 dark:text-slate-400'"
            >
              ISPU · IP/IKA
            </button>
          </div>
          <div
            v-if="pinDisplayMode === 'env'"
            class="inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800"
          >
            <button
              type="button"
              @click="wqmsIndexMode = 'ip'"
              class="px-1.5 py-0.5 rounded-md text-3xs font-extrabold font-mono transition cursor-pointer"
              :class="wqmsIndexMode === 'ip' ? 'bg-teal-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-400'"
            >
              IP
            </button>
            <button
              type="button"
              @click="wqmsIndexMode = 'ika'"
              class="px-1.5 py-0.5 rounded-md text-3xs font-extrabold font-mono transition cursor-pointer"
              :class="wqmsIndexMode === 'ika' ? 'bg-teal-600 text-white btn-text shadow-2xs' : 'text-slate-600 dark:text-slate-400'"
            >
              IKA
            </button>
          </div>

          <!-- AQMS Weather Overlay Toggles in Fullscreen -->
          <div
            v-if="isAqmsWeatherToggleVisible"
            class="inline-flex items-center gap-1"
          >
            <button
              type="button"
              @click="showWindLayer = !showWindLayer"
              class="h-7 px-2 rounded-lg text-3xs font-bold border inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
              :class="showWindLayer
                ? 'bg-cyan-600 text-white btn-text border-cyan-600 shadow-2xs'
                : 'bg-slate-50 dark:bg-[#0B0F19] border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200'"
              title="Toggle overlay Angin AQMS"
            >
              <span>🌬</span>
              <span class="hidden sm:inline">Angin</span>
            </button>
            <button
              type="button"
              @click="showRainLayer = !showRainLayer"
              class="h-7 px-2 rounded-lg text-3xs font-bold border inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
              :class="showRainLayer
                ? 'bg-blue-600 text-white btn-text border-blue-600 shadow-2xs'
                : 'bg-slate-50 dark:bg-[#0B0F19] border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200'"
              title="Toggle overlay Hujan AQMS"
            >
              <span>🌧</span>
              <span class="hidden sm:inline">Hujan</span>
            </button>
          </div>

          <!-- Theme Switcher in Fullscreen -->
          <div class="hidden sm:inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800">
            <button
              v-for="opt in mapStyleOptions"
              :key="'fs-style-' + opt.value"
              type="button"
              @click="setMapStyle(opt.value)"
              class="px-2 py-0.5 rounded-md text-3xs font-bold transition cursor-pointer whitespace-nowrap"
              :class="mapStyle === opt.value
                ? 'bg-sky-600 text-white btn-text shadow-2xs'
                : 'text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-200'"
            >
              {{ opt.label }}
            </button>
          </div>
          <button
            type="button"
            @click="cycleMapStyle"
            class="sm:hidden h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 cursor-pointer"
            :title="'Tema Peta: ' + mapStyleLabel"
          >
            🗺️ {{ mapStyleLabel }}
          </button>

          <!-- Fit & Refresh -->
          <button
            type="button"
            @click="fitMapToVisibleSites"
            class="h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
            title="Fit semua pin stasiun"
          >
            <span>⛶</span>
            <span class="hidden sm:inline">Fit</span>
          </button>
          <button
            type="button"
            :disabled="loading"
            @click="fetchFleetOverview(true)"
            class="h-7 px-2 rounded-lg text-2xs font-semibold bg-sky-500/10 hover:bg-sky-500/20 border border-sky-500/30 text-sky-700 dark:text-sky-300 inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
            :title="'Terakhir diperbarui: ' + lastUpdatedLabel"
          >
            <span :class="{ 'animate-spin': loading }">↻</span>
          </button>

          <!-- Exit Fullscreen Button -->
          <button
            type="button"
            id="btn-fleet-exit-fullscreen"
            @click="exitMapFullscreen"
            class="h-7 px-2.5 rounded-lg text-2xs font-extrabold bg-rose-600 hover:bg-rose-500 text-white btn-text inline-flex items-center gap-1 shadow-xs transition cursor-pointer whitespace-nowrap"
            title="Keluar dari layar penuh (Esc)"
          >
            <span>✕</span>
            <span>Keluar Fullscreen</span>
          </button>
        </div>

        <!-- Mobile Map Floating Utility Buttons (Mode, Fit, Fullscreen & Style when not in fullscreen) -->
        <div v-if="!isMapFullscreen" class="flex md:hidden flex-col gap-1.5 absolute top-2.5 right-2.5 z-[450]">
          <button
            type="button"
            @click="pinDisplayMode = pinDisplayMode === 'env' ? 'status' : 'env'"
            class="px-2 h-8 rounded-xl border shadow-md flex items-center justify-center text-3xs font-extrabold cursor-pointer"
            :class="pinDisplayMode === 'env'
              ? 'bg-emerald-600 text-white btn-text border-emerald-500'
              : 'bg-white/95 dark:bg-[#0E1626]/95 border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200'"
            title="Ganti Mode Pin: Status vs ISPU/IP/IKA"
          >
            {{ pinDisplayMode === 'env' ? 'ISPU' : 'Pin' }}
          </button>
          <button
            v-if="pinDisplayMode === 'env'"
            type="button"
            @click="wqmsIndexMode = wqmsIndexMode === 'ip' ? 'ika' : 'ip'"
            class="px-2 h-7 rounded-xl bg-teal-600 text-white btn-text border border-teal-500 shadow-md flex items-center justify-center text-3xs font-mono font-extrabold cursor-pointer"
            title="Ganti IP / IKA untuk WQMS & ONLIMO"
          >
            {{ wqmsIndexMode.toUpperCase() }}
          </button>
          <button
            v-if="isAqmsWeatherToggleVisible"
            type="button"
            @click="showWindLayer = !showWindLayer"
            class="w-8 h-8 rounded-xl border shadow-md flex items-center justify-center text-xs font-bold cursor-pointer"
            :class="showWindLayer
              ? 'bg-cyan-600 text-white btn-text border-cyan-500'
              : 'bg-white/95 dark:bg-[#0E1626]/95 border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200'"
            title="Toggle Arah & Kecepatan Angin AQMS"
          >
            🌬
          </button>
          <button
            v-if="isAqmsWeatherToggleVisible"
            type="button"
            @click="showRainLayer = !showRainLayer"
            class="relative w-8 h-8 rounded-xl border shadow-md flex items-center justify-center text-xs font-bold cursor-pointer"
            :class="showRainLayer
              ? 'bg-blue-600 text-white btn-text border-blue-500'
              : 'bg-white/95 dark:bg-[#0E1626]/95 border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200'"
            title="Toggle Curah Hujan AQMS"
          >
            🌧
            <span
              v-if="aqmsRainingCount > 0 && !showRainLayer"
              class="absolute -top-1 -right-1 w-3.5 h-3.5 rounded-full bg-blue-600 text-white badge-text text-4xs font-mono font-extrabold flex items-center justify-center"
            >{{ aqmsRainingCount }}</span>
          </button>
          <button
            type="button"
            @click="toggleMapFullscreen"
            class="w-8 h-8 rounded-xl bg-white/95 dark:bg-[#0E1626]/95 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 shadow-md flex items-center justify-center text-xs font-bold cursor-pointer"
            title="Fullscreen Peta"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 8V4h4M20 8V4h-4M4 16v4h4M20 16v4h-4" />
            </svg>
          </button>
          <button
            type="button"
            @click="fitMapToVisibleSites"
            class="w-8 h-8 rounded-xl bg-white/95 dark:bg-[#0E1626]/95 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 shadow-md flex items-center justify-center text-xs font-bold cursor-pointer"
            title="Fit Semua Pin"
          >
            ⛶
          </button>
          <button
            type="button"
            @click="cycleMapStyle"
            class="w-8 h-8 rounded-xl bg-white/95 dark:bg-[#0E1626]/95 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 shadow-md flex items-center justify-center text-xs cursor-pointer"
            :title="'Tema Peta: ' + mapStyleLabel"
          >
            🗺️
          </button>
        </div>

        <!-- Empty Mapped Sites Notice -->
        <div
          v-if="!loading && mappedFilteredSites.length === 0"
          class="absolute top-3 left-1/2 -translate-x-1/2 z-[450] px-3.5 py-2 rounded-xl bg-white/95 dark:bg-[#0E1626]/95 border border-slate-200 dark:border-slate-700 shadow-lg text-center max-w-sm"
        >
          <div class="text-xs font-bold text-slate-900 dark:text-white">Tidak ada pin stasiun pada filter ini</div>
          <div class="text-3xs text-slate-500 dark:text-slate-400 mt-0.5">
            <span v-if="noGpsFilteredSites.length > 0">
              Terdapat {{ noGpsFilteredSites.length }} stasiun belum memiliki koordinat GPS. Klik <strong>Atur GPS</strong> di daftar stasiun.
            </span>
            <span v-else>Coba ubah kata kunci pencarian atau tekan Reset Filter.</span>
          </div>
        </div>

        <!-- Offline Tiles Notice Pill -->
        <div
          v-if="!isTilesOnline"
          class="absolute bottom-2.5 left-2.5 z-[450] px-2.5 py-1 rounded-lg bg-amber-50/95 dark:bg-amber-950/90 text-amber-800 dark:text-amber-300 border border-amber-500/40 text-3xs font-semibold flex items-center gap-1.5 pointer-events-none shadow-sm"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse"></span>
          <span>Tile peta online tidak terjangkau (koordinat pin tetap akurat)</span>
        </div>

        <!-- Compact Legend Pill (ONLY in ISPU / IP / IKA mode — Status mode already shows color dots in top filter pills, zero duplicate info!) -->
        <div
          v-if="isTilesOnline && pinDisplayMode === 'env'"
          id="fleet-env-index-legend"
          class="hidden sm:flex flex-col gap-1 absolute bottom-2.5 left-2.5 z-[450] px-2.5 py-1.5 rounded-xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-xs border border-slate-200 dark:border-slate-700 text-4xs font-semibold text-slate-700 dark:text-slate-200 shadow-md pointer-events-none"
        >
          <div class="flex items-center gap-2 flex-wrap">
            <span class="font-extrabold text-slate-900 dark:text-white">ISPU (AQMS):</span>
            <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-emerald-500"></span>0–50 Baik</span>
            <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-sky-600"></span>51–100 Sedang</span>
            <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-amber-500"></span>101–200 Tdk Sehat</span>
            <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-rose-600"></span>201–300 Sgt Tdk Sehat</span>
            <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-slate-950 border border-rose-500"></span>&gt;300 Berbahaya</span>
          </div>
          <div class="flex items-center gap-2 flex-wrap pt-0.5 border-t border-slate-100 dark:border-slate-800">
            <span class="font-extrabold text-slate-900 dark:text-white">{{ wqmsIndexMode === 'ika' ? 'IKA (WQMS/ONLIMO):' : 'IP (WQMS/ONLIMO):' }}</span>
            <template v-if="wqmsIndexMode === 'ika'">
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-emerald-600"></span>90–100 Sgt Baik</span>
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-sky-600"></span>70–89 Baik</span>
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-amber-500"></span>50–69 Sedang</span>
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-orange-600"></span>25–49 Kurang</span>
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-rose-600"></span>&lt;25 Sgt Kurang</span>
            </template>
            <template v-else>
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-teal-600"></span>≤1.0 Memenuhi</span>
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-amber-500"></span>1–5 Cemar Ringan</span>
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-orange-600"></span>5–10 Cemar Sedang</span>
              <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-rose-600"></span>&gt;10 Cemar Berat</span>
            </template>
          </div>
        </div>

        <!-- Compact AQMS Weather Overlay Legend (ONLY when Angin or Hujan overlay is toggled ON) -->
        <div
          v-if="isTilesOnline && isAqmsWeatherToggleVisible && (showWindLayer || showRainLayer)"
          id="fleet-weather-overlay-legend"
          class="hidden sm:flex flex-col gap-1 absolute left-2.5 z-[450] px-2.5 py-1.5 rounded-xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-xs border border-slate-200 dark:border-slate-700 text-4xs font-semibold text-slate-700 dark:text-slate-200 shadow-md pointer-events-none"
          :class="pinDisplayMode === 'env' ? 'bottom-[74px]' : 'bottom-2.5'"
        >
          <div v-if="showWindLayer" class="flex items-center gap-2 flex-wrap">
            <span class="font-extrabold text-cyan-700 dark:text-cyan-300">🌬 Aliran Angin:</span>
            <span class="inline-flex items-center gap-1"><span class="w-2.5 h-1 rounded-full bg-sky-600"></span>&lt;2 Sepoi</span>
            <span class="inline-flex items-center gap-1"><span class="w-2.5 h-1 rounded-full bg-emerald-600"></span>2–5 Sedang</span>
            <span class="inline-flex items-center gap-1"><span class="w-2.5 h-1 rounded-full bg-amber-600"></span>5–10 Kencang</span>
            <span class="inline-flex items-center gap-1"><span class="w-2.5 h-1 rounded-full bg-rose-600"></span>&gt;10 m/s</span>
          </div>
          <div
            v-if="showRainLayer"
            class="flex items-center gap-2 flex-wrap"
            :class="{ 'pt-0.5 border-t border-slate-100 dark:border-slate-800': showWindLayer }"
          >
            <span class="font-extrabold text-blue-700 dark:text-blue-300">🌧 Hujan (Ikut Wind Angle):</span>
            <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-sky-400"></span>Ringan &lt;2</span>
            <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-blue-600"></span>Sedang 2–10</span>
            <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-indigo-700"></span>Lebat &gt;10 mm</span>
            <span class="font-mono font-extrabold text-blue-700 dark:text-blue-300">({{ aqmsRainingCount }} Aktif)</span>
          </div>
        </div>

        <!-- =============================================================== -->
        <!-- DESKTOP FLOATING SITE DETAIL CARD (>= 768px)                    -->
        <!-- =============================================================== -->
        <transition name="fade">
          <div
            v-if="selectedSite && isDesktopViewport"
            :key="'desk-site-card-' + selectedSite.site_id"
            id="fleet-desktop-site-card"
            class="hidden md:flex flex-col absolute left-3 z-[500] w-[325px] lg:w-[350px] rounded-2xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-md border border-slate-200 dark:border-slate-700/90 shadow-2xl overflow-hidden"
            :class="isTabletFloatingMode && !isMapFullscreen ? 'top-16 max-h-[calc(100%-74px)]' : 'top-3 max-h-[calc(100%-24px)]'"
          >
            <!-- Top Accent Bar -->
            <div class="h-1 w-full" :class="severityBarClass(selectedSite)"></div>

            <!-- Card Header -->
            <div class="p-3 pb-2.5 border-b border-slate-100 dark:border-slate-800/80 flex items-start justify-between gap-2">
              <div class="min-w-0">
                <div class="flex items-center gap-1.5 flex-wrap">
                  <span class="px-1.5 py-0.5 rounded text-4xs font-mono font-extrabold border" :class="domainBadgeClass(selectedSite.monitoring_domain)">
                    {{ selectedSite.monitoring_domain }}
                  </span>
                  <span v-if="selectedSite.project_year > 0" class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border border-indigo-500/30">
                    📅 {{ selectedSite.project_year }}
                  </span>
                  <span class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold border" :class="deliveryBadgeClass(selectedSite.delivery_status)">
                    {{ deliveryLabel(selectedSite.delivery_status) }}
                  </span>
                  <span v-if="!selectedSite.has_coordinates" class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/30">
                    NO-GPS
                  </span>
                </div>
                <h3 class="text-sm font-extrabold text-slate-900 dark:text-white mt-1 leading-snug break-words">
                  {{ selectedSite.site_name }}
                </h3>
                <div class="text-2xs font-semibold text-sky-600 dark:text-sky-400 truncate">
                  🏢 {{ selectedSite.tenant_name }}
                  <span class="font-mono text-3xs text-slate-400">({{ selectedSite.site_code }})</span>
                </div>
                <div v-if="selectedSite.address" class="text-3xs text-slate-500 dark:text-slate-400 truncate mt-0.5">
                  📍 {{ selectedSite.address }}
                </div>
              </div>
              <button
                type="button"
                @click="selectedSiteId = 0"
                class="w-6 h-6 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-500 dark:text-slate-300 flex items-center justify-center text-xs flex-shrink-0 cursor-pointer"
                title="Tutup kartu"
              >
                ✕
              </button>
            </div>

            <!-- Compact Telemetry & Device Metrics Grid -->
            <div class="grid grid-cols-3 gap-1.5 px-3 py-2 bg-slate-50/80 dark:bg-[#0B0F19]/60 border-b border-slate-100 dark:border-slate-800/80 text-center">
              <div class="px-1.5 py-1 rounded-lg bg-white dark:bg-[#111827] border border-slate-200/70 dark:border-slate-800">
                <div class="text-4xs uppercase tracking-wider text-slate-400 font-semibold">Kirim Terakhir</div>
                <div class="text-2xs font-extrabold text-slate-900 dark:text-white tabular-nums mt-0.5">
                  {{ formatRelativeAge(selectedSite) }}
                </div>
                <div class="text-4xs font-mono text-slate-400 truncate">
                  {{ formatClockTime(selectedSite.last_transmission_at) }}
                </div>
              </div>
              <div class="px-1.5 py-1 rounded-lg bg-white dark:bg-[#111827] border border-slate-200/70 dark:border-slate-800">
                <div class="text-4xs uppercase tracking-wider text-slate-400 font-semibold">Perangkat</div>
                <div class="text-2xs font-extrabold text-slate-900 dark:text-white tabular-nums mt-0.5">
                  {{ selectedSite.online_device_count }}/{{ selectedSite.enabled_device_count }} Online
                </div>
                <div class="text-4xs font-mono text-slate-400 truncate">
                  {{ selectedSite.primary_protocol || 'BELUM ADA' }}
                </div>
              </div>
              <div class="px-1.5 py-1 rounded-lg bg-white dark:bg-[#111827] border border-slate-200/70 dark:border-slate-800">
                <div class="text-4xs uppercase tracking-wider text-slate-400 font-semibold">Alarm Aktif</div>
                <div
                  class="text-2xs font-extrabold tabular-nums mt-0.5"
                  :class="selectedSite.total_active_alarms > 0 ? 'text-rose-600 dark:text-rose-400' : 'text-emerald-600 dark:text-emerald-400'"
                >
                  {{ selectedSite.total_active_alarms > 0 ? selectedSite.total_active_alarms + ' Aktif' : 'Normal' }}
                </div>
                <div class="text-4xs font-mono text-slate-400 truncate">
                  <span v-if="selectedSite.critical_alarm_count > 0">{{ selectedSite.critical_alarm_count }} Crit</span>
                  <span v-else-if="selectedSite.warning_alarm_count > 0">{{ selectedSite.warning_alarm_count }} Warn</span>
                  <span v-else>0 Gangguan</span>
                </div>
              </div>
            </div>

            <!-- Environmental Index Detail Section (AQMS ISPU or WQMS/ONLIMO IP & IKA) -->
            <div
              v-if="selectedSite.monitoring_domain === 'AQMS' && selectedSite.ispu && selectedSite.ispu.available"
              class="px-3 py-2 border-b border-slate-100 dark:border-slate-800/80 space-y-1.5"
            >
              <div class="p-2 rounded-xl border" :class="ispuBannerClass(selectedSite.ispu.category)">
                <div class="flex items-center justify-between gap-2">
                  <div class="flex items-baseline gap-1.5">
                    <span class="text-3xs font-bold uppercase opacity-80">ISPU</span>
                    <span class="text-lg font-extrabold font-mono leading-none tabular-nums">{{ selectedSite.ispu.score }}</span>
                  </div>
                  <span class="px-2 py-0.5 rounded-full text-4xs font-extrabold uppercase tracking-wider text-white badge-text" :class="ispuSolidBadgeClass(selectedSite.ispu.category)">
                    {{ selectedSite.ispu.category_label || selectedSite.ispu.category }}
                  </span>
                </div>
                <div class="text-3xs mt-1 opacity-90 flex items-center justify-between gap-1">
                  <span>Dominan: <strong>{{ selectedSite.ispu.dominant_pollutant || '-' }}</strong> <span v-if="selectedSite.ispu.dominant_concentration > 0">({{ selectedSite.ispu.dominant_concentration }} {{ selectedSite.ispu.dominant_unit || 'µg/m³' }})</span></span>
                  <span class="font-mono text-4xs opacity-75">Permen LHK 14/2020</span>
                </div>
              </div>
              <div v-if="selectedSiteIspuParams.length > 0" class="space-y-1 pt-0.5">
                <div
                  v-for="pm in selectedSiteIspuParams"
                  :key="'ispu-p-' + selectedSite.site_id + '-' + pm.code"
                  class="flex items-center gap-2 text-3xs"
                >
                  <span class="w-10 font-mono font-bold text-slate-600 dark:text-slate-300">{{ pm.code }}</span>
                  <div class="flex-1 h-1.5 rounded-full bg-slate-200 dark:bg-slate-800 overflow-hidden">
                    <div
                      class="h-full rounded-full transition-all"
                      :class="ispuSolidBadgeClass(pm.category)"
                      :style="{ width: Math.min(100, Math.max(6, (Number(pm.sub_index || 0) / 300) * 100)) + '%' }"
                    ></div>
                  </div>
                  <span class="w-8 text-right font-mono font-bold text-slate-800 dark:text-slate-200 tabular-nums">{{ pm.category === 'TIDAK_TERSEDIA' ? '--' : Math.round(pm.sub_index || 0) }}</span>
                </div>
              </div>
            </div>

            <div
              v-else-if="(selectedSite.monitoring_domain === 'WQMS' || selectedSite.monitoring_domain === 'ONLIMO') && selectedSite.wqms && selectedSite.wqms.available"
              class="px-3 py-2 border-b border-slate-100 dark:border-slate-800/80 space-y-1.5"
            >
              <div class="grid grid-cols-2 gap-1.5">
                <div class="p-2 rounded-xl border" :class="wqmsIPBannerClass(selectedSite.wqms.ip_category)">
                  <div class="text-4xs font-bold uppercase opacity-80">IP (Kepmen 115)</div>
                  <div class="text-base font-extrabold font-mono leading-tight tabular-nums mt-0.5">{{ Number(selectedSite.wqms.ip_score || 0).toFixed(2) }}</div>
                  <div class="mt-1">
                    <span class="px-1.5 py-0.2 rounded text-4xs font-extrabold text-white badge-text" :class="wqmsIPSolidBadgeClass(selectedSite.wqms.ip_category)">
                      {{ selectedSite.wqms.ip_category_label || selectedSite.wqms.ip_category }}
                    </span>
                  </div>
                  <div v-if="selectedSite.wqms.dominant_parameter" class="text-4xs mt-1 opacity-85 truncate">
                    Dominan: <strong>{{ selectedSite.wqms.dominant_parameter }}</strong>
                  </div>
                </div>
                <div class="p-2 rounded-xl border" :class="wqmsIKABannerClass(selectedSite.wqms.ika_category)">
                  <div class="text-4xs font-bold uppercase opacity-80">IKA (Permen 27)</div>
                  <div class="text-base font-extrabold font-mono leading-tight tabular-nums mt-0.5">{{ Math.round(selectedSite.wqms.ika_score || 0) }}<span class="text-3xs font-normal opacity-75">/100</span></div>
                  <div class="mt-1">
                    <span class="px-1.5 py-0.2 rounded text-4xs font-extrabold text-white badge-text" :class="wqmsIKASolidBadgeClass(selectedSite.wqms.ika_category)">
                      {{ selectedSite.wqms.ika_category_label || selectedSite.wqms.ika_category }}
                    </span>
                  </div>
                  <div class="text-4xs mt-1 opacity-85 truncate">
                    Mutu: <strong>{{ formatWaterClassLabel(selectedSite.wqms.water_class) }}</strong>
                  </div>
                </div>
              </div>
              <div v-if="selectedSiteWqmsParams.length > 0" class="flex items-center gap-1 flex-wrap pt-0.5">
                <span
                  v-for="wp in selectedSiteWqmsParams"
                  :key="'wq-p-' + selectedSite.site_id + '-' + wp.code"
                  class="px-1.5 py-0.5 rounded-md bg-slate-100 dark:bg-slate-800/90 border border-slate-200 dark:border-slate-700 text-4xs font-mono text-slate-700 dark:text-slate-200"
                >
                  <strong>{{ wp.code }}</strong> {{ wp.concentration }}{{ wp.unit && wp.unit !== 'pH' ? ' ' + wp.unit : '' }}
                </span>
              </div>
            </div>

            <!-- Active Alarms List (Scrollable if multiple) -->
            <div class="px-3 py-2 overflow-y-auto max-h-40 space-y-1.5">
              <template v-if="selectedSite.active_alarms && selectedSite.active_alarms.length > 0">
                <div class="text-4xs font-bold uppercase tracking-wider text-slate-400">Ringkasan Alarm Aktif</div>
                <div
                  v-for="al in selectedSite.active_alarms"
                  :key="al.id"
                  class="px-2 py-1.5 rounded-lg border text-2xs"
                  :class="al.severity === 'CRITICAL'
                    ? 'bg-rose-500/10 border-rose-500/30 text-rose-700 dark:text-rose-300'
                    : 'bg-amber-500/10 border-amber-500/30 text-amber-800 dark:text-amber-300'"
                >
                  <div class="flex items-center justify-between gap-1 font-bold">
                    <span class="truncate">
                      {{ al.severity === 'CRITICAL' ? '🚨' : '⚠️' }}
                      {{ al.parameter_name || al.device_code || al.alarm_code }}
                    </span>
                    <span class="font-mono text-4xs opacity-80 flex-shrink-0">{{ formatTimestampAgo(al.triggered_at) }}</span>
                  </div>
                  <div class="text-3xs opacity-90 line-clamp-2 mt-0.5">{{ al.message }}</div>
                </div>
              </template>
              <div
                v-else
                class="px-2.5 py-2 rounded-lg bg-emerald-500/10 border border-emerald-500/25 text-emerald-700 dark:text-emerald-300 text-2xs font-semibold flex items-center gap-1.5"
              >
                <span>✓</span>
                <span>Seluruh parameter stasiun beroperasi tanpa alarm aktif.</span>
              </div>
            </div>

            <!-- Action Footer -->
            <div class="p-2.5 pt-2 border-t border-slate-100 dark:border-slate-800/80 bg-slate-50/60 dark:bg-[#0B0F19]/40 flex flex-col gap-1.5">
              <button
                type="button"
                id="btn-fleet-open-dashboard"
                :disabled="switchingSiteId === selectedSite.site_id"
                @click="openSiteDashboard(selectedSite)"
                class="w-full py-2 px-3 rounded-xl bg-sky-600 hover:bg-sky-500 text-white btn-text text-xs font-extrabold shadow-sm transition flex items-center justify-center gap-1.5 cursor-pointer"
              >
                <span v-if="switchingSiteId === selectedSite.site_id">Membuka Stasiun...</span>
                <span v-else>Buka Dashboard Stasiun →</span>
              </button>
              <div class="grid gap-1.5" :class="canEditStationGps ? 'grid-cols-3' : 'grid-cols-2'">
                <button
                  type="button"
                  @click="openSiteModule(selectedSite, (!isSuperAdmin && isCloudMode) ? '/monitoring/telemetry' : '/monitoring/devices')"
                  class="py-1 px-2 rounded-lg bg-white hover:bg-slate-100 dark:bg-slate-800/90 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 text-3xs font-bold transition cursor-pointer"
                >
                  {{ (!isSuperAdmin && isCloudMode) ? `Live Data (${selectedSite.device_count})` : `Perangkat (${selectedSite.device_count})` }}
                </button>
                <button
                  v-if="selectedSite.monitoring_domain === 'AQMS'"
                  type="button"
                  @click="openSiteModule(selectedSite, '/monitoring/ispu')"
                  class="py-1 px-2 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/30 text-emerald-700 dark:text-emerald-300 text-3xs font-bold transition cursor-pointer"
                >
                  Analisis ISPU
                </button>
                <button
                  v-else-if="selectedSite.monitoring_domain === 'WQMS' || selectedSite.monitoring_domain === 'ONLIMO'"
                  type="button"
                  @click="openSiteModule(selectedSite, '/monitoring/wqms')"
                  class="py-1 px-2 rounded-lg bg-teal-500/10 hover:bg-teal-500/20 border border-teal-500/30 text-teal-700 dark:text-teal-300 text-3xs font-bold transition cursor-pointer"
                >
                  Analisis IP/IKA
                </button>
                <button
                  v-else
                  type="button"
                  @click="openSiteModule(selectedSite, '/alarm/active')"
                  class="py-1 px-2 rounded-lg bg-white hover:bg-slate-100 dark:bg-slate-800/90 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 text-3xs font-bold transition cursor-pointer"
                >
                  Detail Alarm
                </button>
                <button
                  v-if="canEditStationGps"
                  type="button"
                  @click="openGpsModal(selectedSite)"
                  class="py-1 px-2 rounded-lg bg-white hover:bg-slate-100 dark:bg-slate-800/90 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 text-sky-600 dark:text-sky-400 text-3xs font-bold transition cursor-pointer"
                >
                  📍 {{ selectedSite.has_coordinates ? 'Edit GPS' : 'Atur GPS' }}
                </button>
              </div>
            </div>
          </div>
        </transition>

        <!-- =============================================================== -->
        <!-- MOBILE DOCKED BOTTOM PREVIEW CARD (< 768px in Peta Mode)        -->
        <!-- =============================================================== -->
        <transition name="fade">
          <div
            v-if="selectedSite && !isDesktopViewport && mobileViewMode === 'map'"
            :key="'mob-site-card-' + selectedSite.site_id"
            id="fleet-mobile-bottom-card"
            class="flex md:hidden flex-col absolute inset-x-2 bottom-2 z-[500] rounded-2xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-md border border-slate-200 dark:border-slate-700 shadow-2xl overflow-hidden"
          >
            <div class="h-1 w-full" :class="severityBarClass(selectedSite)"></div>
            <div class="p-3 space-y-2">
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <div class="flex items-center gap-1.5 flex-wrap">
                    <span class="px-1.5 py-0.2 rounded text-4xs font-mono font-extrabold border" :class="domainBadgeClass(selectedSite.monitoring_domain)">
                      {{ selectedSite.monitoring_domain }}
                    </span>
                    <span v-if="selectedSite.project_year > 0" class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border border-indigo-500/30">
                      📅 {{ selectedSite.project_year }}
                    </span>
                    <span class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold border" :class="deliveryBadgeClass(selectedSite.delivery_status)">
                      {{ deliveryLabel(selectedSite.delivery_status) }}
                    </span>
                    <span class="text-3xs font-mono text-slate-500 dark:text-slate-400">
                      Kirim: <strong class="text-slate-800 dark:text-slate-200">{{ formatRelativeAge(selectedSite) }}</strong>
                    </span>
                  </div>
                  <div class="text-sm font-extrabold text-slate-900 dark:text-white truncate mt-0.5">
                    {{ selectedSite.site_name }}
                  </div>
                  <div class="text-2xs font-semibold text-sky-600 dark:text-sky-400 truncate">
                    {{ selectedSite.tenant_name }} · {{ selectedSite.online_device_count }}/{{ selectedSite.enabled_device_count }} Device Online
                  </div>
                </div>
                <button
                  type="button"
                  @click="selectedSiteId = 0"
                  class="w-7 h-7 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-500 dark:text-slate-300 flex items-center justify-center text-xs flex-shrink-0 cursor-pointer"
                >
                  ✕
                </button>
              </div>

              <!-- Mobile Compact Alarm Strip -->
              <div v-if="selectedSite.active_alarms && selectedSite.active_alarms.length > 0" class="flex items-center gap-1.5 overflow-x-auto fleet-no-scrollbar">
                <span
                  v-for="al in selectedSite.active_alarms.slice(0, 3)"
                  :key="'m-al-' + al.id"
                  class="px-2 py-0.5 rounded-md text-3xs font-bold border whitespace-nowrap flex-shrink-0"
                  :class="al.severity === 'CRITICAL'
                    ? 'bg-rose-500/10 border-rose-500/30 text-rose-700 dark:text-rose-300'
                    : 'bg-amber-500/10 border-amber-500/30 text-amber-800 dark:text-amber-300'"
                >
                  {{ al.severity === 'CRITICAL' ? '🚨' : '⚠️' }} {{ al.parameter_name || al.alarm_code }}
                </span>
              </div>

              <!-- Mobile Thumb Action Row -->
              <div class="flex items-center gap-2 pt-0.5">
                <button
                  type="button"
                  :disabled="switchingSiteId === selectedSite.site_id"
                  @click="openSiteDashboard(selectedSite)"
                  class="flex-1 py-2 px-3 rounded-xl bg-sky-600 hover:bg-sky-500 text-white btn-text text-xs font-extrabold shadow-xs text-center cursor-pointer"
                >
                  {{ switchingSiteId === selectedSite.site_id ? 'Membuka...' : 'Buka Dashboard Stasiun →' }}
                </button>
                <button
                  type="button"
                  @click="openSiteModule(selectedSite, '/alarm/active')"
                  class="py-2 px-2.5 rounded-xl bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 text-2xs font-bold cursor-pointer"
                >
                  Alarm ({{ selectedSite.total_active_alarms }})
                </button>
              </div>
            </div>
          </div>
        </transition>
      </div>

      <!-- Interactive Vertical Splitter Handle (Drag Left/Right to Resize Right Sidebar on Wide Desktop) -->
      <div
        v-if="!isRightDockCollapsed && !isMapFullscreen && !isTabletFloatingMode"
        id="fleet-dock-resize-handle"
        @mousedown.prevent="startDockResize"
        @touchstart.prevent="startDockResize"
        @dblclick="resetRightDockWidth"
        class="hidden md:flex items-center justify-center w-3 flex-shrink-0 cursor-col-resize group select-none"
        title="Geser kiri / kanan untuk mengatur lebar panel Prioritas Stasiun (Klik 2x untuk ukuran awal)"
      >
        <div
          class="w-1 rounded-full transition-all duration-150"
          :class="isResizingDock
            ? 'h-16 bg-sky-500 shadow-[0_0_8px_rgba(14,165,233,0.65)]'
            : 'h-10 bg-slate-300/90 dark:bg-slate-700/90 group-hover:h-14 group-hover:bg-sky-500'"
        ></div>
      </div>

      <!-- ================================================================= -->
      <!-- 3B. DESKTOP RIGHT PRIORITY STATION DOCK (Wide Split / Tablet Glass Drawer) -->
      <!-- ================================================================= -->
      <aside
        v-if="!isRightDockCollapsed"
        id="fleet-priority-dock"
        :style="{ width: (isTabletFloatingMode ? Math.min(340, effectiveRightDockWidth) : effectiveRightDockWidth) + 'px' }"
        class="hidden md:flex flex-col flex-shrink-0 rounded-2xl border border-slate-200/90 dark:border-slate-800/90 overflow-hidden"
        :class="isTabletFloatingMode
          ? 'absolute top-24 right-2.5 bottom-2.5 z-[510] bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-md shadow-2xl'
          : 'h-full bg-white dark:bg-[#0E1626] shadow-xs'"
      >
        <!-- Dock Header + Sub-Tabs (Semua | Normal | Masalah | No-GPS) -->
        <div class="px-3 pt-2.5 pb-2 border-b border-slate-100 dark:border-slate-800/80 space-y-2 flex-shrink-0">
          <div class="flex items-center justify-between gap-2">
            <span class="text-xs font-extrabold text-slate-900 dark:text-white tracking-tight">Prioritas Stasiun</span>
            <div class="flex items-center gap-1.5">
              <span class="text-3xs font-mono text-slate-400">{{ dockDisplayedSites.length }} ditampilkan</span>
              <button
                v-if="isTabletFloatingMode"
                type="button"
                @click="isRightDockCollapsed = true"
                class="w-6 h-6 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-500 dark:text-slate-300 flex items-center justify-center text-2xs font-bold cursor-pointer"
                title="Tutup Daftar Stasiun"
              >
                ✕
              </button>
            </div>
          </div>
          <div class="grid grid-cols-4 gap-1 p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200/70 dark:border-slate-800">
            <button
              type="button"
              id="btn-fleet-dock-tab-all"
              @click="dockTab = 'ALL'"
              class="py-1 px-1 rounded-md text-4xs font-bold transition cursor-pointer truncate"
              :class="dockTab === 'ALL' ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-white shadow-2xs' : 'text-slate-500 dark:text-slate-400'"
              :title="'Semua Stasiun (' + filteredSites.length + ')'"
            >
              Semua ({{ filteredSites.length }})
            </button>
            <button
              type="button"
              id="btn-fleet-dock-tab-normal"
              @click="dockTab = 'NORMAL'"
              class="py-1 px-1 rounded-md text-4xs font-bold transition cursor-pointer truncate"
              :class="dockTab === 'NORMAL' ? 'bg-emerald-600 text-white btn-text shadow-2xs' : 'text-slate-500 dark:text-slate-400'"
              :title="'Stasiun Normal / Online Tanpa Alarm (' + normalFilteredSites.length + ')'"
            >
              Normal ({{ normalFilteredSites.length }})
            </button>
            <button
              type="button"
              id="btn-fleet-dock-tab-issues"
              @click="dockTab = 'ISSUES'"
              class="py-1 px-1 rounded-md text-4xs font-bold transition cursor-pointer truncate"
              :class="dockTab === 'ISSUES' ? 'bg-rose-600 text-white btn-text shadow-2xs' : 'text-slate-500 dark:text-slate-400'"
              :title="'Stasiun Bermasalah / Alarm / Offline (' + issueFilteredSites.length + ')'"
            >
              Masalah ({{ issueFilteredSites.length }})
            </button>
            <button
              type="button"
              id="btn-fleet-dock-tab-nogps"
              @click="dockTab = 'NOGPS'"
              class="py-1 px-1 rounded-md text-4xs font-bold transition cursor-pointer truncate"
              :class="dockTab === 'NOGPS' ? 'bg-amber-600 text-white btn-text shadow-2xs' : 'text-slate-500 dark:text-slate-400'"
              :title="'Stasiun Belum Ada Koordinat GPS (' + noGpsFilteredSites.length + ')'"
            >
              No-GPS ({{ noGpsFilteredSites.length }})
            </button>
          </div>
        </div>

        <!-- Dense 52px Station Rows -->
        <div class="flex-1 min-h-0 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/60">
          <div
            v-for="site in dockDisplayedSites"
            :key="site.site_id"
            @click="focusSite(site)"
            class="group relative pl-3 pr-2.5 py-2 transition cursor-pointer flex items-center justify-between gap-2"
            :class="Number(selectedSiteId) === Number(site.site_id)
              ? 'bg-sky-50/90 dark:bg-sky-950/30'
              : 'hover:bg-slate-50 dark:hover:bg-slate-800/40'"
          >
            <!-- Left 3px Urgency Accent Bar -->
            <span class="absolute left-0 top-1 bottom-1 w-1 rounded-r" :class="severityBarClass(site)"></span>

            <div class="min-w-0 flex-1">
              <!-- Line 1: Site Name + Domain Badge + Project Year -->
              <div class="flex items-center gap-1.5 min-w-0">
                <span class="text-xs font-bold text-slate-900 dark:text-white truncate">{{ site.site_name }}</span>
                <span class="px-1 py-0.2 rounded text-4xs font-mono font-bold border flex-shrink-0" :class="domainBadgeClass(site.monitoring_domain)">
                  {{ site.monitoring_domain }}
                </span>
                <span v-if="site.project_year > 0" class="px-1 py-0.2 rounded text-4xs font-mono font-bold bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border border-indigo-500/25 flex-shrink-0">
                  {{ site.project_year }}
                </span>
              </div>
              <!-- Line 2: Tenant + Last Transmission (+ Inline Wind/Rain Chips when Overlay ON) -->
              <div class="flex items-center gap-1.5 text-3xs text-slate-500 dark:text-slate-400 mt-0.5 min-w-0 flex-wrap">
                <span class="truncate font-medium">{{ site.tenant_name }}</span>
                <span>·</span>
                <span class="font-mono whitespace-nowrap" :class="deliveryTextClass(site.delivery_status)">
                  {{ formatRelativeAge(site) }}
                </span>
                <span
                  v-if="showWindLayer && isSiteAqmsWeatherEligible(site) && site.weather && (site.weather.wind_speed != null || site.weather.wind_angle != null)"
                  class="px-1 py-0.2 rounded text-4xs font-mono font-bold bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border border-cyan-500/25 whitespace-nowrap"
                >
                  🌬 {{ site.weather.wind_speed != null ? Number(site.weather.wind_speed).toFixed(1) + ' ' + (site.weather.wind_speed_unit || 'm/s') : '' }}{{ site.weather.wind_angle != null ? (site.weather.wind_speed != null ? ' ' : '') + formatWindCompassShort(site.weather.wind_angle).idShort : '' }}
                </span>
                <span
                  v-if="showRainLayer && isSiteAqmsWeatherEligible(site) && site.weather && site.weather.is_raining"
                  class="px-1 py-0.2 rounded text-4xs font-mono font-bold bg-blue-500/15 text-blue-700 dark:text-blue-300 border border-blue-500/30 whitespace-nowrap"
                >
                  🌧 {{ Number(site.weather.rain_fall || 0).toFixed(1) }} {{ site.weather.rain_fall_unit || 'mm' }}
                </span>
              </div>
            </div>

            <!-- Right Indicators & Quick Action -->
            <div class="flex items-center gap-1 flex-shrink-0">
              <span
                v-if="pinDisplayMode === 'env' && getSiteEnvDockBadge(site)"
                class="w-6 h-6 rounded-full text-4xs font-mono font-extrabold text-white badge-text flex items-center justify-center shadow-2xs"
                :style="{ background: getSiteEnvDockBadge(site).bg }"
                :title="getSiteEnvDockBadge(site).tooltip"
              >
                {{ getSiteEnvDockBadge(site).label }}
              </span>
              <span
                v-if="site.total_active_alarms > 0"
                class="px-1.5 py-0.5 rounded-full text-4xs font-mono font-extrabold text-white badge-text"
                :class="site.critical_alarm_count > 0 ? 'bg-rose-600' : 'bg-amber-500'"
                :title="site.total_active_alarms + ' alarm aktif'"
              >
                🚨 {{ site.total_active_alarms }}
              </span>
              <button
                v-if="!site.has_coordinates && canEditStationGps"
                type="button"
                @click.stop="openGpsModal(site)"
                class="px-1.5 py-0.5 rounded-md bg-amber-500/15 hover:bg-amber-500/25 text-amber-700 dark:text-amber-300 border border-amber-500/30 text-4xs font-bold cursor-pointer"
                title="Set Koordinat GPS Stasiun"
              >
                + GPS
              </button>
              <span
                v-else-if="!site.has_coordinates"
                class="px-1.5 py-0.5 rounded-md bg-amber-500/10 text-amber-700 dark:text-amber-300 border border-amber-500/25 text-4xs font-mono font-bold"
              >
                NO-GPS
              </span>
              <button
                type="button"
                @click.stop="openSiteDashboard(site)"
                class="w-6 h-6 rounded-lg bg-slate-100 hover:bg-sky-600 dark:bg-slate-800 dark:hover:bg-sky-600 text-slate-600 hover:text-white dark:text-slate-300 btn-text flex items-center justify-center text-2xs font-bold transition cursor-pointer"
                title="Buka Dashboard Stasiun"
              >
                →
              </button>
            </div>
          </div>

          <!-- Empty State -->
          <div v-if="!loading && dockDisplayedSites.length === 0" class="p-6 text-center text-xs text-slate-400">
            Tidak ada stasiun yang cocok dengan filter aktif.
          </div>
        </div>
      </aside>

      <!-- ================================================================= -->
      <!-- 3C. MOBILE PRIORITY CARD LIST (< 768px in Daftar Mode)            -->
      <!-- ================================================================= -->
      <div
        v-if="!isDesktopViewport && mobileViewMode === 'list'"
        id="fleet-mobile-priority-list"
        class="flex md:hidden flex-col flex-1 min-h-0 overflow-y-auto gap-2 pb-2"
      >
        <div
          v-for="site in filteredSites"
          :key="'mob-card-' + site.site_id"
          class="relative pl-3.5 pr-3 py-2.5 rounded-xl bg-white dark:bg-[#0E1626] border border-slate-200/90 dark:border-slate-800/90 shadow-2xs space-y-2 flex-shrink-0"
        >
          <span class="absolute left-0 top-2 bottom-2 w-1 rounded-r" :class="severityBarClass(site)"></span>

          <div class="flex items-start justify-between gap-2">
            <div class="min-w-0">
              <div class="flex items-center gap-1.5 flex-wrap">
                <span class="px-1.5 py-0.2 rounded text-4xs font-mono font-extrabold border" :class="domainBadgeClass(site.monitoring_domain)">
                  {{ site.monitoring_domain }}
                </span>
                <span v-if="site.project_year > 0" class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border border-indigo-500/30">
                  📅 {{ site.project_year }}
                </span>
                <span class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold border" :class="deliveryBadgeClass(site.delivery_status)">
                  {{ deliveryLabel(site.delivery_status) }}
                </span>
                <span
                  v-if="site.total_active_alarms > 0"
                  class="px-1.5 py-0.2 rounded text-4xs font-mono font-extrabold text-white badge-text"
                  :class="site.critical_alarm_count > 0 ? 'bg-rose-600' : 'bg-amber-500'"
                >
                  🚨 {{ site.total_active_alarms }} Alarm
                </span>
              </div>
              <div class="text-xs font-extrabold text-slate-900 dark:text-white mt-1 leading-snug">
                {{ site.site_name }}
              </div>
              <div class="text-2xs font-medium text-slate-500 dark:text-slate-400 truncate">
                🏢 {{ site.tenant_name }} · {{ site.online_device_count }}/{{ site.enabled_device_count }} Device
              </div>
            </div>
            <div class="text-right flex-shrink-0">
              <div class="text-4xs uppercase text-slate-400 font-semibold">Kirim Terakhir</div>
              <div class="text-2xs font-mono font-bold" :class="deliveryTextClass(site.delivery_status)">
                {{ formatRelativeAge(site) }}
              </div>
            </div>
          </div>

          <!-- Top Alarm Message Preview (if any) -->
          <div
            v-if="site.active_alarms && site.active_alarms.length > 0"
            class="px-2 py-1 rounded-lg text-3xs truncate border"
            :class="site.critical_alarm_count > 0
              ? 'bg-rose-500/10 border-rose-500/25 text-rose-700 dark:text-rose-300'
              : 'bg-amber-500/10 border-amber-500/25 text-amber-800 dark:text-amber-300'"
          >
            {{ site.active_alarms[0].parameter_name || site.active_alarms[0].alarm_code }}: {{ site.active_alarms[0].message }}
          </div>

          <!-- Mobile Card Action Buttons -->
          <div class="flex items-center gap-2 pt-0.5">
            <button
              v-if="site.has_coordinates"
              type="button"
              @click="jumpMobileToMapPin(site)"
              class="px-3 py-1.5 rounded-lg bg-slate-100 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 text-2xs font-bold cursor-pointer"
            >
              📍 Peta
            </button>
            <button
              v-else-if="canEditStationGps"
              type="button"
              @click="openGpsModal(site)"
              class="px-3 py-1.5 rounded-lg bg-amber-500/15 border border-amber-500/30 text-amber-700 dark:text-amber-300 text-2xs font-bold cursor-pointer"
            >
              + Set GPS
            </button>
            <button
              type="button"
              :disabled="switchingSiteId === site.site_id"
              @click="openSiteDashboard(site)"
              class="flex-1 py-1.5 px-3 rounded-lg bg-sky-600 hover:bg-sky-500 text-white btn-text text-2xs font-extrabold text-center cursor-pointer"
            >
              {{ switchingSiteId === site.site_id ? 'Membuka...' : 'Buka Dashboard →' }}
            </button>
          </div>
        </div>

        <div v-if="!loading && filteredSites.length === 0" class="p-8 text-center text-xs text-slate-400">
          Tidak ada stasiun yang cocok dengan pencarian / filter saat ini.
        </div>
      </div>
    </div>

    <!-- ===================================================================== -->
    <!-- 4. FILTER SHEET / MODAL (Mobile Bottom Sheet + Tablet Floating Card)  -->
    <!-- ===================================================================== -->
    <div
      v-if="showMobileFilterSheet"
      @click.self="showMobileFilterSheet = false"
      class="fixed inset-0 z-[80] flex items-end md:items-center justify-center p-0 md:p-4 bg-black/55 backdrop-blur-xs"
    >
      <div class="w-full md:max-w-lg rounded-t-3xl md:rounded-2xl bg-white dark:bg-[#0E1626] border-t md:border border-slate-200 dark:border-slate-800 p-4 space-y-3.5 shadow-2xl max-h-[82vh] overflow-y-auto">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-extrabold text-slate-900 dark:text-white">Filter Peta & Stasiun</h3>
          <button type="button" @click="showMobileFilterSheet = false" class="text-slate-400 text-sm font-bold cursor-pointer">✕</button>
        </div>

        <div class="space-y-2.5 text-xs">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
            <div>
              <label class="block text-2xs font-bold text-slate-500 dark:text-slate-400 mb-1">Tahun Project</label>
              <select
                id="select-fleet-mobile-project-year"
                v-model="projectYearFilter"
                class="w-full h-9 px-2.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-100"
              >
                <option value="ALL">Semua Tahun</option>
                <option v-for="yr in projectYearOptions" :key="'m-yr-' + yr.year" :value="String(yr.year)">
                  📅 Tahun {{ yr.year }} ({{ yr.count }} site)
                </option>
                <option v-if="unsetProjectYearCount > 0" value="UNSET">
                  Belum Diset ({{ unsetProjectYearCount }} site)
                </option>
              </select>
            </div>
            <div>
              <label class="block text-2xs font-bold text-slate-500 dark:text-slate-400 mb-1">Organisasi Tenant</label>
              <select
                v-model="tenantFilter"
                class="w-full h-9 px-2.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-100"
              >
                <option value="ALL">Semua Tenant ({{ displayedTenantOptions.length }})</option>
                <option v-for="t in displayedTenantOptions" :key="'m-t-' + t.id" :value="String(t.id)">
                  {{ t.name }}{{ t.project_year > 0 ? ` [${t.project_year}]` : '' }} ({{ t.site_count }} site)
                </option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-2">
            <div>
              <label class="block text-2xs font-bold text-slate-500 dark:text-slate-400 mb-1">Domain Stasiun</label>
              <select
                v-model="domainFilter"
                class="w-full h-9 px-2.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-100"
              >
                <option value="ALL">Semua Domain</option>
                <option value="AQMS">AQMS</option>
                <option value="WQMS">WQMS</option>
                <option value="SPARING">SPARING</option>
                <option value="ONLIMO">ONLIMO</option>
              </select>
            </div>
            <div>
              <label class="block text-2xs font-bold text-slate-500 dark:text-slate-400 mb-1">Status Alarm</label>
              <select
                v-model="alarmFilter"
                class="w-full h-9 px-2.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-100"
              >
                <option value="ALL">Semua</option>
                <option value="ANY_ALARM">Ada Alarm</option>
                <option value="CRITICAL">Critical Saja</option>
                <option value="WARNING">Warning Saja</option>
                <option value="CLEAN">Tanpa Alarm</option>
              </select>
            </div>
          </div>

          <div>
            <label class="block text-2xs font-bold text-slate-500 dark:text-slate-400 mb-1">Tema Peta (100% Free)</label>
            <div class="grid grid-cols-5 gap-1">
              <button
                v-for="opt in mapStyleOptions"
                :key="opt.value"
                type="button"
                @click="setMapStyle(opt.value)"
                class="py-1.5 rounded-lg text-3xs font-bold border cursor-pointer"
                :class="mapStyle === opt.value
                  ? 'bg-sky-600 text-white btn-text border-sky-600'
                  : 'bg-slate-50 dark:bg-[#0B0F19] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700'"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-2 pt-2 border-t border-slate-100 dark:border-slate-800">
          <button
            type="button"
            @click="resetAllFilters"
            class="px-4 py-2 rounded-xl border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 text-xs font-bold cursor-pointer"
          >
            Reset
          </button>
          <button
            type="button"
            @click="showMobileFilterSheet = false"
            class="flex-1 py-2 rounded-xl bg-sky-600 text-white btn-text text-xs font-extrabold cursor-pointer"
          >
            Terapkan ({{ filteredSites.length }} Stasiun)
          </button>
        </div>
      </div>
    </div>

    <!-- ===================================================================== -->
    <!-- 5. INLINE QUICK SET / EDIT STATION GPS MODAL                          -->
    <!-- ===================================================================== -->
    <div
      v-if="showGpsModal"
      @click.self="showGpsModal = false"
      class="fixed inset-0 z-[90] flex items-center justify-center p-3 sm:p-4 bg-black/60 backdrop-blur-xs"
    >
      <div class="w-full max-w-xl rounded-2xl bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <div class="px-4 py-3 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between">
          <div>
            <h3 class="text-sm font-extrabold text-slate-900 dark:text-white">
              📍 Atur Koordinat GPS Stasiun
            </h3>
            <p class="text-2xs text-slate-500 dark:text-slate-400">
              {{ gpsForm.site_name }} · <span class="text-sky-600 dark:text-sky-400 font-semibold">{{ gpsForm.tenant_name }}</span>
            </p>
          </div>
          <button type="button" @click="showGpsModal = false" class="text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 text-sm font-bold cursor-pointer">✕</button>
        </div>

        <div class="p-4 space-y-3 overflow-y-auto">
          <div class="h-60 w-full rounded-xl overflow-hidden">
            <StationMapViewer
              :latitude="gpsPickerLat"
              :longitude="gpsPickerLng"
              :station-name="gpsForm.site_name"
              :editable="true"
              height="240px"
              @coordinate-change="onPickStationGps"
              @location-selected="onSelectStationLocation"
            />
          </div>

          <div class="grid grid-cols-2 gap-2.5">
            <div>
              <label class="block text-2xs font-bold text-slate-600 dark:text-slate-300 mb-1">Latitude</label>
              <input
                v-model.number="gpsForm.latitude"
                type="number"
                step="0.000001"
                placeholder="-6.200000"
                class="w-full h-8 px-2.5 rounded-lg text-xs font-mono bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white"
              />
            </div>
            <div>
              <label class="block text-2xs font-bold text-slate-600 dark:text-slate-300 mb-1">Longitude</label>
              <input
                v-model.number="gpsForm.longitude"
                type="number"
                step="0.000001"
                placeholder="106.816666"
                class="w-full h-8 px-2.5 rounded-lg text-xs font-mono bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white"
              />
            </div>
          </div>

          <div>
            <label class="block text-2xs font-bold text-slate-600 dark:text-slate-300 mb-1">Alamat / Kota Lokasi</label>
            <input
              v-model.trim="gpsForm.address"
              type="text"
              placeholder="Contoh: Jl. MH Thamrin, Jakarta Pusat"
              class="w-full h-8 px-2.5 rounded-lg text-xs bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-900 dark:text-white"
            />
          </div>
        </div>

        <div class="px-4 py-3 border-t border-slate-100 dark:border-slate-800 bg-slate-50/60 dark:bg-[#0B0F19]/40 flex items-center justify-between gap-2">
          <router-link
            to="/administration/tenants?tab=sites"
            class="text-2xs font-bold text-sky-600 dark:text-sky-400 hover:underline"
          >
            Kelola Semua Stasiun →
          </router-link>
          <div class="flex items-center gap-2">
            <button
              type="button"
              @click="showGpsModal = false"
              class="px-3 py-1.5 rounded-xl border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 text-xs font-semibold cursor-pointer"
            >
              Batal
            </button>
            <button
              type="button"
              :disabled="savingGps"
              @click="saveStationGps"
              class="px-4 py-1.5 rounded-xl bg-sky-600 hover:bg-sky-500 text-white btn-text text-xs font-extrabold shadow-xs cursor-pointer"
            >
              {{ savingGps ? 'Menyimpan...' : 'Simpan GPS' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import L from 'leaflet';
import 'leaflet/dist/leaflet.css';
import axios from 'axios';
import StationMapViewer from '../components/StationMapViewer.vue';
import { FleetWeatherCanvasOverlay } from '../utils/fleetWeatherCanvasLayer';

const INDONESIA_DEFAULT_CENTER = [-2.5489, 118.0149];
const INDONESIA_DEFAULT_ZOOM = 5;

// 100% Free, Zero-API-Key, Zero-Watermark Map Tile Providers (Esri Public ArcGIS Online & OpenStreetMap)
const TILE_PANE_FILTER_CLASSES = [
  'fleet-tiles-osm-vibrant',
  'fleet-tiles-topo-vibrant',
  'fleet-tiles-sat-crisp',
  'fleet-tiles-night-comfort',
  'fleet-tiles-navy-night',
];

const TILE_LAYERS = {
  osm: {
    url: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
    fallbackUrl: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}',
    paneFilterClass: 'fleet-tiles-osm-vibrant',
    fallbackPaneFilterClass: 'fleet-tiles-osm-vibrant',
    options: { maxZoom: 19 },
  },
  clean: {
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}',
    fallbackUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
    paneFilterClass: 'fleet-tiles-topo-vibrant',
    fallbackPaneFilterClass: 'fleet-tiles-osm-vibrant',
    options: { maxZoom: 19 },
  },
  night_clean: {
    url: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
    fallbackUrl: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}',
    paneFilterClass: 'fleet-tiles-night-comfort',
    fallbackPaneFilterClass: 'fleet-tiles-night-comfort',
    options: { maxZoom: 19 },
  },
  dark: {
    url: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
    fallbackUrl: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Street_Map/MapServer/tile/{z}/{y}/{x}',
    paneFilterClass: 'fleet-tiles-navy-night',
    fallbackPaneFilterClass: 'fleet-tiles-navy-night',
    options: { maxZoom: 19 },
  },
  satellite: {
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    overlayUrl: 'https://server.arcgisonline.com/ArcGIS/rest/services/Reference/World_Boundaries_and_Places/MapServer/tile/{z}/{y}/{x}',
    fallbackUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
    paneFilterClass: 'fleet-tiles-sat-crisp',
    fallbackPaneFilterClass: 'fleet-tiles-osm-vibrant',
    options: { maxZoom: 18 },
  },
};

export function isValidStationCoordinate(lat, lng) {
  if (lat === null || lat === undefined || lng === null || lng === undefined || lat === '' || lng === '') {
    return false;
  }
  const la = Number(lat);
  const ln = Number(lng);
  if (!Number.isFinite(la) || !Number.isFinite(ln)) return false;
  if (la < -90 || la > 90 || ln < -180 || ln > 180) return false;
  if (Math.abs(la) < 0.0001 && Math.abs(ln) < 0.0001) return false;
  return true;
}

export default {
  name: 'FleetMapView',
  components: {
    StationMapViewer,
  },
  data() {
    return {
      loading: false,
      errorMessage: '',
      generatedAt: null,
      nowTick: Date.now(),
      summary: {
        total_tenants: 0,
        active_tenants: 0,
        total_sites: 0,
        mapped_sites: 0,
        no_gps_sites: 0,
        online_sites: 0,
        delayed_sites: 0,
        offline_sites: 0,
        inactive_sites: 0,
        sites_with_alarms: 0,
        total_active_alarms: 0,
        critical_alarms: 0,
        warning_alarms: 0,
        total_devices: 0,
        online_devices: 0,
      },
      tenants: [],
      sites: [],
      projectYears: [],

      // Filters
      quickFilter: 'ALL', // ALL | ONLINE | DELAYED | OFFLINE | ALARM | NOGPS
      searchQuery: '',
      projectYearFilter: 'ALL',
      tenantFilter: 'ALL',
      domainFilter: 'ALL',
      alarmFilter: 'ALL',
      dockTab: 'ALL', // ALL | ISSUES | NOGPS

      // Responsive & Layout State
      viewportWidth: typeof window !== 'undefined' ? window.innerWidth : 1280,
      viewportHeight: typeof window !== 'undefined' ? window.innerHeight : 800,
      mobileViewMode: 'map', // 'map' | 'list'
      showMobileFilterSheet: false,
      showLayerPopover: false,
      isRightDockCollapsed: typeof window !== 'undefined'
        ? (window.innerWidth < 1360 || window.innerHeight <= 820)
        : false,
      rightDockWidth: (() => {
        try {
          const saved = Number(localStorage.getItem('datalogger_fleet_dock_width'));
          if (Number.isFinite(saved) && saved >= 240 && saved <= 720) return saved;
        } catch (_) {}
        return 335;
      })(),
      isResizingDock: false,
      dockResizeStartX: 0,
      dockResizeStartWidth: 335,
      dockResizeRaf: null,
      isMapFullscreen: false,

      // Map & Selection State
      map: null,
      tileLayer: null,
      tileOverlayLayer: null,
      markersLayer: null,
      weatherCanvasOverlay: null,
      markerBySiteId: {},
      hasInitialFitDone: false,
      isTilesOnline: true,
      mapStyle: 'auto', // 'auto' | 'osm' | 'clean' | 'satellite' | 'dark'
      pinDisplayMode: 'status', // 'status' | 'env'
      wqmsIndexMode: 'ip', // 'ip' | 'ika'
      showWindLayer: (() => {
        try {
          return localStorage.getItem('datalogger_fleet_show_wind') === '1';
        } catch (_) {
          return false;
        }
      })(),
      showRainLayer: (() => {
        try {
          return localStorage.getItem('datalogger_fleet_show_rain') === '1';
        } catch (_) {
          return false;
        }
      })(),
      spiderfiedClusterKey: '',
      selectedSiteId: 0,
      switchingSiteId: 0,

      // Quick Set GPS Modal State
      showGpsModal: false,
      savingGps: false,
      gpsForm: {
        site_id: 0,
        site_name: '',
        tenant_name: '',
        latitude: -6.2,
        longitude: 106.816666,
        address: '',
      },

      // Timers
      pollTimer: null,
      clockTimer: null,
      wsRefreshDebounce: null,
    };
  },
  computed: {
    isSuperAdmin() {
      return Boolean(this.$store && this.$store.getters && this.$store.getters.isSuperAdmin);
    },
    isCustomerService() {
      return Boolean(this.$store && this.$store.getters && this.$store.getters.isCustomerService);
    },
    canEditStationGps() {
      return this.isSuperAdmin && !this.isCustomerService;
    },
    isCloudMode() {
      const sys = (this.$store && this.$store.state && this.$store.state.systemStatus) || {};
      return Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
    },
    isDesktopViewport() {
      return this.viewportWidth >= 768;
    },
    isTabletFloatingMode() {
      return this.isDesktopViewport && (this.viewportWidth < 1360 || this.viewportHeight <= 820);
    },
    floatingStatusPills() {
      const keep = ['ALL', 'ONLINE', 'DELAYED', 'OFFLINE', 'ALARM'];
      if ((this.summary.no_gps_sites || 0) > 0) keep.push('NOGPS');
      return this.kpiPills.filter(p => keep.includes(p.key));
    },
    effectiveRightDockWidth() {
      const maxAllowed = Math.max(260, Math.min(720, Math.floor(this.viewportWidth * 0.58)));
      return Math.min(Math.max(240, Number(this.rightDockWidth) || 335), maxAllowed);
    },
    normalSitesTotalCount() {
      return (this.sites || []).filter(s => s.severity_state === 'NORMAL').length;
    },
    effectiveAppDark() {
      const pref = this.$store && this.$store.state && this.$store.state.uiPreferences;
      if (pref && pref.effectiveTheme) return pref.effectiveTheme === 'dark';
      return Boolean(this.$store && this.$store.state && this.$store.state.theme === 'dark');
    },
    resolvedTileTheme() {
      if (this.mapStyle === 'auto') {
        return this.effectiveAppDark ? 'night_clean' : 'osm';
      }
      return this.mapStyle;
    },
    mapStyleLabel() {
      const map = { auto: 'Auto', osm: 'Jalan', clean: 'Topo', satellite: 'Satelit', dark: 'Malam' };
      return map[this.mapStyle] || 'Auto';
    },
    mapStyleOptions() {
      return [
        { value: 'auto', label: 'Auto' },
        { value: 'osm', label: 'Jalan' },
        { value: 'clean', label: 'Topo' },
        { value: 'satellite', label: 'Satelit' },
        { value: 'dark', label: 'Malam' },
      ];
    },
    kpiPills() {
      return [
        {
          key: 'ALL',
          label: 'Semua',
          count: this.summary.total_sites || 0,
          dotClass: '',
          activeClass: 'bg-sky-600 text-white btn-text border-sky-600 shadow-2xs',
          idleClass: 'bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800/80 text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-800',
        },
        {
          key: 'NORMAL',
          label: 'Normal',
          count: this.normalSitesTotalCount,
          dotClass: 'bg-teal-500',
          activeClass: 'bg-teal-600 text-white btn-text border-teal-600 shadow-2xs',
          idleClass: 'bg-teal-500/5 hover:bg-teal-500/15 text-teal-700 dark:text-teal-300 border-teal-500/25',
        },
        {
          key: 'ONLINE',
          label: 'Online',
          count: this.summary.online_sites || 0,
          dotClass: 'bg-emerald-500',
          activeClass: 'bg-emerald-600 text-white btn-text border-emerald-600 shadow-2xs',
          idleClass: 'bg-emerald-500/5 hover:bg-emerald-500/15 text-emerald-700 dark:text-emerald-400 border-emerald-500/25',
        },
        {
          key: 'DELAYED',
          label: 'Delayed',
          count: this.summary.delayed_sites || 0,
          dotClass: 'bg-amber-500',
          activeClass: 'bg-amber-500 text-white btn-text border-amber-500 shadow-2xs',
          idleClass: 'bg-amber-500/5 hover:bg-amber-500/15 text-amber-700 dark:text-amber-400 border-amber-500/25',
        },
        {
          key: 'OFFLINE',
          label: 'Offline',
          count: this.summary.offline_sites || 0,
          dotClass: 'bg-rose-500',
          activeClass: 'bg-rose-600 text-white btn-text border-rose-600 shadow-2xs',
          idleClass: 'bg-rose-500/5 hover:bg-rose-500/15 text-rose-700 dark:text-rose-400 border-rose-500/25',
        },
        {
          key: 'ALARM',
          label: 'Alarm',
          count: this.summary.sites_with_alarms || 0,
          dotClass: 'bg-rose-500 animate-ping',
          activeClass: 'bg-rose-600 text-white btn-text border-rose-600 shadow-2xs',
          idleClass: 'bg-rose-500/5 hover:bg-rose-500/15 text-rose-700 dark:text-rose-300 border-rose-500/25',
        },
        {
          key: 'NOGPS',
          label: 'No-GPS',
          count: this.summary.no_gps_sites || 0,
          dotClass: 'bg-slate-400',
          activeClass: 'bg-indigo-600 text-white btn-text border-indigo-600 shadow-2xs',
          idleClass: 'bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-200 dark:border-slate-800',
        },
      ];
    },
    projectYearOptions() {
      const countByYear = {};
      (this.projectYears || []).forEach((y) => {
        const num = Number(y) || 0;
        if (num > 0 && countByYear[num] === undefined) countByYear[num] = 0;
      });
      (this.tenants || []).forEach((t) => {
        const num = Number(t.project_year) || 0;
        if (num > 0 && countByYear[num] === undefined) countByYear[num] = 0;
      });
      (this.sites || []).forEach((s) => {
        const num = Number(s.project_year) || 0;
        if (num > 0) {
          countByYear[num] = (countByYear[num] || 0) + 1;
        }
      });
      return Object.keys(countByYear)
        .map(y => Number(y))
        .filter(y => y > 0)
        .sort((a, b) => b - a)
        .map(year => ({ year, count: countByYear[year] || 0 }));
    },
    unsetProjectYearCount() {
      return (this.sites || []).filter(s => !(Number(s.project_year) > 0)).length;
    },
    displayedTenantOptions() {
      if (this.projectYearFilter === 'ALL') return this.tenants || [];
      const matchingTenantIds = new Set(
        (this.sites || [])
          .filter((s) => {
            const py = Number(s.project_year) || 0;
            return this.projectYearFilter === 'UNSET' ? py <= 0 : String(py) === String(this.projectYearFilter);
          })
          .map(s => Number(s.tenant_id))
      );
      return (this.tenants || []).filter((t) => {
        const tpy = Number(t.project_year) || 0;
        if (this.projectYearFilter === 'UNSET') {
          return tpy <= 0 || matchingTenantIds.has(Number(t.id));
        }
        return String(tpy) === String(this.projectYearFilter) || matchingTenantIds.has(Number(t.id));
      });
    },
    hasActiveFilters() {
      return (
        this.quickFilter !== 'ALL' ||
        Boolean(this.searchQuery) ||
        this.projectYearFilter !== 'ALL' ||
        this.tenantFilter !== 'ALL' ||
        this.domainFilter !== 'ALL' ||
        this.alarmFilter !== 'ALL'
      );
    },
    activeSecondaryFilterCount() {
      let n = 0;
      if (this.projectYearFilter !== 'ALL') n += 1;
      if (this.tenantFilter !== 'ALL') n += 1;
      if (this.domainFilter !== 'ALL') n += 1;
      if (this.alarmFilter !== 'ALL') n += 1;
      return n;
    },
    filteredSites() {
      const q = (this.searchQuery || '').toLowerCase();
      return (this.sites || []).filter((s) => {
        const py = Number(s.project_year) || 0;
        if (this.projectYearFilter === 'UNSET' && py > 0) {
          return false;
        }
        if (this.projectYearFilter !== 'ALL' && this.projectYearFilter !== 'UNSET' && String(py) !== String(this.projectYearFilter)) {
          return false;
        }
        if (this.tenantFilter !== 'ALL' && String(s.tenant_id) !== String(this.tenantFilter)) {
          return false;
        }
        if (this.domainFilter !== 'ALL' && String(s.monitoring_domain).toUpperCase() !== this.domainFilter) {
          return false;
        }
        if (this.alarmFilter === 'ANY_ALARM' && !(s.total_active_alarms > 0)) return false;
        if (this.alarmFilter === 'CRITICAL' && !(s.critical_alarm_count > 0)) return false;
        if (this.alarmFilter === 'WARNING' && !(s.warning_alarm_count > 0)) return false;
        if (this.alarmFilter === 'CLEAN' && s.total_active_alarms > 0) return false;

        if (this.quickFilter === 'NORMAL' && s.severity_state !== 'NORMAL') return false;
        if (this.quickFilter === 'ONLINE' && s.delivery_status !== 'ONLINE') return false;
        if (this.quickFilter === 'DELAYED' && s.delivery_status !== 'DELAYED') return false;
        if (this.quickFilter === 'OFFLINE' && s.delivery_status !== 'OFFLINE') return false;
        if (this.quickFilter === 'ALARM' && !(s.total_active_alarms > 0)) return false;
        if (this.quickFilter === 'NOGPS' && s.has_coordinates) return false;

        if (q) {
          const hay = [
            s.site_name,
            s.site_code,
            s.tenant_name,
            s.tenant_code,
            s.customer_name,
            s.address,
            s.monitoring_domain,
            s.primary_protocol,
            py > 0 ? String(py) : '',
          ]
            .filter(Boolean)
            .join(' ')
            .toLowerCase();
          if (!hay.includes(q)) return false;
        }
        return true;
      });
    },
    mappedFilteredSites() {
      return this.filteredSites.filter(
        s => s.has_coordinates && isValidStationCoordinate(s.latitude, s.longitude)
      );
    },
    normalFilteredSites() {
      return this.filteredSites.filter(s => s.severity_state === 'NORMAL');
    },
    issueFilteredSites() {
      return this.filteredSites.filter(
        s => s.severity_state === 'CRITICAL' || s.severity_state === 'WARNING'
      );
    },
    noGpsFilteredSites() {
      return this.filteredSites.filter(s => !s.has_coordinates);
    },
    dockDisplayedSites() {
      let base = this.filteredSites;
      if (this.dockTab === 'NORMAL') base = this.normalFilteredSites;
      else if (this.dockTab === 'ISSUES') base = this.issueFilteredSites;
      else if (this.dockTab === 'NOGPS') return this.noGpsFilteredSites;

      if (this.pinDisplayMode === 'env') {
        return [...base].sort((a, b) => {
          const ma = this.resolveSiteEnvPinMeta(a);
          const mb = this.resolveSiteEnvPinMeta(b);
          if (mb.urgencyRank !== ma.urgencyRank) return mb.urgencyRank - ma.urgencyRank;
          return (mb.numericScore || 0) - (ma.numericScore || 0);
        });
      }
      return base;
    },
    selectedSite() {
      if (!this.selectedSiteId) return null;
      return (this.sites || []).find(s => Number(s.site_id) === Number(this.selectedSiteId)) || null;
    },
    selectedSiteIspuParams() {
      const site = this.selectedSite;
      if (!site || !site.ispu || !Array.isArray(site.ispu.parameters)) return [];
      const byCode = new Map();
      for (const pm of site.ispu.parameters) {
        if (!pm || !pm.code) continue;
        const code = String(pm.code).trim().toUpperCase();
        if (!code) continue;
        const prev = byCode.get(code);
        if (!prev || Number(pm.sub_index || 0) > Number(prev.sub_index || 0)) {
          byCode.set(code, { ...pm, code });
        }
      }
      return Array.from(byCode.values())
        .sort((a, b) => Number(b.sub_index || 0) - Number(a.sub_index || 0))
        .slice(0, 7);
    },
    selectedSiteWqmsParams() {
      const site = this.selectedSite;
      if (!site || !site.wqms || !Array.isArray(site.wqms.parameters)) return [];
      const byCode = new Map();
      for (const wp of site.wqms.parameters) {
        if (!wp || !wp.code) continue;
        const code = String(wp.code).trim().toUpperCase();
        if (!code) continue;
        const prev = byCode.get(code);
        if (!prev || Number(wp.sub_index || 0) > Number(prev.sub_index || 0)) {
          byCode.set(code, { ...wp, code });
        }
      }
      return Array.from(byCode.values()).slice(0, 8);
    },
    isAqmsWeatherToggleVisible() {
      return this.domainFilter === 'ALL' || this.domainFilter === 'AQMS';
    },
    aqmsWindAvailableCount() {
      return (this.mappedFilteredSites || []).filter(
        s => this.isSiteAqmsWeatherEligible(s) && s.weather && (s.weather.wind_speed != null || s.weather.wind_angle != null)
      ).length;
    },
    aqmsRainingCount() {
      return (this.mappedFilteredSites || []).filter(
        s => this.isSiteAqmsWeatherEligible(s) && s.weather && s.weather.is_raining
      ).length;
    },
    gpsPickerLat() {
      return isValidStationCoordinate(this.gpsForm.latitude, this.gpsForm.longitude)
        ? Number(this.gpsForm.latitude)
        : null;
    },
    gpsPickerLng() {
      return isValidStationCoordinate(this.gpsForm.latitude, this.gpsForm.longitude)
        ? Number(this.gpsForm.longitude)
        : null;
    },
    lastUpdatedShort() {
      if (!this.generatedAt) return 'Live';
      const diff = Math.max(0, Math.floor((this.nowTick - new Date(this.generatedAt).getTime()) / 1000));
      if (diff < 5) return 'Baru saja';
      if (diff < 60) return `${diff}d lalu`;
      return `${Math.floor(diff / 60)}m lalu`;
    },
    lastUpdatedLabel() {
      if (!this.generatedAt) return '-';
      try {
        return new Date(this.generatedAt).toLocaleTimeString('id-ID');
      } catch (_) {
        return '-';
      }
    },
  },
  watch: {
    resolvedTileTheme() {
      this.applyTileLayer();
      this.syncWeatherCanvasLayer();
    },
    pinDisplayMode() {
      this.syncMapMarkers();
    },
    wqmsIndexMode() {
      if (this.pinDisplayMode === 'env') {
        this.syncMapMarkers();
      }
    },
    showWindLayer(val) {
      try {
        localStorage.setItem('datalogger_fleet_show_wind', val ? '1' : '0');
      } catch (_) {}
      this.syncMapMarkers();
      this.syncWeatherCanvasLayer();
    },
    showRainLayer(val) {
      try {
        localStorage.setItem('datalogger_fleet_show_rain', val ? '1' : '0');
      } catch (_) {}
      this.syncMapMarkers();
      this.syncWeatherCanvasLayer();
    },
    domainFilter() {
      this.syncWeatherCanvasLayer();
    },
    mappedFilteredSites() {
      this.syncMapMarkers();
      this.syncWeatherCanvasLayer();
    },
    projectYearFilter() {
      if (this.tenantFilter !== 'ALL') {
        const stillVisible = this.displayedTenantOptions.some(t => String(t.id) === String(this.tenantFilter));
        if (!stillVisible) {
          this.tenantFilter = 'ALL';
        }
      }
      this.$nextTick(() => {
        if (this.mappedFilteredSites.length > 0) {
          this.fitMapToVisibleSites();
        }
      });
    },
    selectedSiteId() {
      this.syncMapMarkers();
    },
    isRightDockCollapsed() {
      this.$nextTick(() => {
        if (this.map) this.map.invalidateSize();
      });
    },
  },
  mounted() {
    if (!this.isSuperAdmin && !this.isCustomerService) {
      this.$router.replace({ path: '/access-denied', query: { from: this.$route.fullPath } });
      return;
    }
    if (typeof window !== 'undefined') {
      window.addEventListener('resize', this.onResize, { passive: true });
      window.addEventListener('alarm-event', this.scheduleRealtimeRefresh);
      window.addEventListener('keydown', this.onGlobalKeydown);
    }
    if (typeof document !== 'undefined') {
      document.addEventListener('fullscreenchange', this.onNativeFullscreenChange);
      document.addEventListener('webkitfullscreenchange', this.onNativeFullscreenChange);
    }
    this.$nextTick(() => {
      this.initLeafletMap();
      this.fetchFleetOverview(false);
    });
    this.pollTimer = setInterval(() => {
      this.fetchFleetOverview(false);
    }, 15000);
    this.clockTimer = setInterval(() => {
      this.nowTick = Date.now();
    }, 5000);
  },
  beforeDestroy() {
    if (this.pollTimer) clearInterval(this.pollTimer);
    if (this.clockTimer) clearInterval(this.clockTimer);
    if (this.wsRefreshDebounce) clearTimeout(this.wsRefreshDebounce);
    this.stopDockResize();
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', this.onResize);
      window.removeEventListener('alarm-event', this.scheduleRealtimeRefresh);
      window.removeEventListener('keydown', this.onGlobalKeydown);
    }
    if (typeof document !== 'undefined') {
      document.removeEventListener('fullscreenchange', this.onNativeFullscreenChange);
      document.removeEventListener('webkitfullscreenchange', this.onNativeFullscreenChange);
    }
    if (this.weatherCanvasOverlay) {
      this.weatherCanvasOverlay.destroy();
      this.weatherCanvasOverlay = null;
    }
    if (this.map) {
      this.map.remove();
      this.map = null;
    }
  },
  methods: {
    onResize() {
      if (typeof window === 'undefined') return;
      const wasTabletFloating = this.isTabletFloatingMode;
      this.viewportWidth = window.innerWidth;
      this.viewportHeight = window.innerHeight;
      if (!wasTabletFloating && this.isTabletFloatingMode) {
        this.isRightDockCollapsed = true;
      }
      this.$nextTick(() => {
        if (this.map) this.map.invalidateSize();
      });
    },
    extractClientX(ev) {
      if (!ev) return null;
      if (typeof ev.clientX === 'number') return ev.clientX;
      if (ev.touches && ev.touches.length > 0) return ev.touches[0].clientX;
      if (ev.changedTouches && ev.changedTouches.length > 0) return ev.changedTouches[0].clientX;
      return null;
    },
    startDockResize(ev) {
      const x = this.extractClientX(ev);
      if (x === null || typeof window === 'undefined') return;
      this.isResizingDock = true;
      this.dockResizeStartX = x;
      this.dockResizeStartWidth = this.effectiveRightDockWidth;
      window.addEventListener('mousemove', this.onDockResizeMove);
      window.addEventListener('mouseup', this.stopDockResize);
      window.addEventListener('touchmove', this.onDockResizeMove, { passive: false });
      window.addEventListener('touchend', this.stopDockResize);
    },
    onDockResizeMove(ev) {
      if (!this.isResizingDock) return;
      const x = this.extractClientX(ev);
      if (x === null) return;
      if (ev.cancelable) ev.preventDefault();
      const deltaX = x - this.dockResizeStartX;
      const maxAllowed = Math.max(260, Math.min(720, Math.floor(this.viewportWidth * 0.58)));
      const nextWidth = Math.min(Math.max(240, Math.round(this.dockResizeStartWidth - deltaX)), maxAllowed);
      this.rightDockWidth = nextWidth;
      if (!this.dockResizeRaf && typeof window !== 'undefined' && window.requestAnimationFrame) {
        this.dockResizeRaf = window.requestAnimationFrame(() => {
          this.dockResizeRaf = null;
          if (this.map) this.map.invalidateSize({ pan: false });
        });
      }
    },
    stopDockResize() {
      if (this.dockResizeRaf && typeof window !== 'undefined' && window.cancelAnimationFrame) {
        window.cancelAnimationFrame(this.dockResizeRaf);
        this.dockResizeRaf = null;
      }
      if (typeof window !== 'undefined') {
        window.removeEventListener('mousemove', this.onDockResizeMove);
        window.removeEventListener('mouseup', this.stopDockResize);
        window.removeEventListener('touchmove', this.onDockResizeMove);
        window.removeEventListener('touchend', this.stopDockResize);
      }
      if (!this.isResizingDock) return;
      this.isResizingDock = false;
      try {
        localStorage.setItem('datalogger_fleet_dock_width', String(Math.round(this.rightDockWidth)));
      } catch (_) {}
      this.$nextTick(() => {
        if (this.map) this.map.invalidateSize({ pan: false });
      });
    },
    resetRightDockWidth() {
      this.rightDockWidth = 335;
      try {
        localStorage.setItem('datalogger_fleet_dock_width', '335');
      } catch (_) {}
      this.$nextTick(() => {
        if (this.map) this.map.invalidateSize({ pan: false });
      });
    },
    scheduleRealtimeRefresh() {
      if (this.wsRefreshDebounce) clearTimeout(this.wsRefreshDebounce);
      this.wsRefreshDebounce = setTimeout(() => {
        this.fetchFleetOverview(false);
      }, 1200);
    },
    applyOverviewPayload(payload) {
      if (!payload) return;
      this.generatedAt = payload.generated_at || new Date().toISOString();
      this.nowTick = Date.now();
      this.summary = { ...this.summary, ...(payload.summary || {}) };
      this.projectYears = Array.isArray(payload.project_years) ? payload.project_years : [];
      this.tenants = Array.isArray(payload.tenants) ? payload.tenants : [];
      this.sites = Array.isArray(payload.sites) ? payload.sites : [];
      this.errorMessage = '';

      this.$nextTick(() => {
        this.syncMapMarkers();
        this.syncWeatherCanvasLayer();
        if (!this.hasInitialFitDone && this.mappedFilteredSites.length > 0) {
          this.hasInitialFitDone = true;
          this.fitMapToVisibleSites();
        }
      });
    },
    async fetchFleetOverviewFallback() {
      const globalHeaders = { 'X-Tenant-ID': '0', 'X-Site-ID': '0' };
      const [tenantsRes, sitesRes, devsRes, alarmsRes] = await Promise.all([
        axios.get('/api/tenants', { headers: globalHeaders }),
        axios.get('/api/sites', { headers: globalHeaders }),
        axios.get('/api/devices', { params: { page_size: 500 }, headers: globalHeaders }),
        axios.get('/api/alarms/active', { headers: globalHeaders }).catch(() => ({ data: { data: [] } })),
      ]);

      const rawTenants = Array.isArray(tenantsRes.data && tenantsRes.data.data) ? tenantsRes.data.data : [];
      const rawSites = Array.isArray(sitesRes.data && sitesRes.data.data) ? sitesRes.data.data : [];
      const rawDevsPayload = devsRes.data && devsRes.data.data;
      const rawDevs = Array.isArray(rawDevsPayload)
        ? rawDevsPayload
        : Array.isArray(rawDevsPayload && rawDevsPayload.items)
        ? rawDevsPayload.items
        : [];
      const rawAlarmsPayload = alarmsRes.data && alarmsRes.data.data;
      const rawAlarms = Array.isArray(rawAlarmsPayload)
        ? rawAlarmsPayload
        : Array.isArray(rawAlarmsPayload && rawAlarmsPayload.items)
        ? rawAlarmsPayload.items
        : [];

      const nowMs = Date.now();
      const tenantMap = {};
      rawTenants.forEach((t) => {
        tenantMap[t.id] = t;
      });
      const devById = {};
      const devsBySite = {};
      rawDevs.forEach((d) => {
        devById[d.id] = d;
        const sid = d.site_id || 1;
        if (!devsBySite[sid]) devsBySite[sid] = [];
        devsBySite[sid].push(d);
      });

      const alarmsBySite = {};
      rawAlarms.forEach((al) => {
        if (!al || al.status === 'CLEARED') return;
        const dev = al.device_id ? devById[al.device_id] : al.device;
        if (dev && (dev.enabled === false || String(dev.status || '').toUpperCase() === 'DISABLED')) return;
        const sid = (dev && dev.site_id) || al.site_id || 1;
        if (!alarmsBySite[sid]) alarmsBySite[sid] = [];
        alarmsBySite[sid].push(al);
      });

      const summary = {
        total_tenants: rawTenants.length,
        active_tenants: rawTenants.filter(t => String(t.status || 'ACTIVE').toUpperCase() !== 'SUSPENDED').length,
        total_sites: 0,
        mapped_sites: 0,
        no_gps_sites: 0,
        online_sites: 0,
        delayed_sites: 0,
        offline_sites: 0,
        inactive_sites: 0,
        sites_with_alarms: 0,
        total_active_alarms: 0,
        critical_alarms: 0,
        warning_alarms: 0,
        total_devices: 0,
        online_devices: 0,
      };

      const tenantStats = {};
      const tenantOpts = rawTenants.map((t) => {
        const opt = {
          id: t.id,
          code: t.code,
          name: t.name,
          status: t.status || 'ACTIVE',
          project_year: Number(t.project_year) || 0,
          site_count: 0,
          online_sites: 0,
          alarm_sites: 0,
        };
        tenantStats[t.id] = opt;
        return opt;
      });

      const sites = [];
      rawSites.forEach((st) => {
        const ten = tenantMap[st.tenant_id] || { id: st.tenant_id, code: '', name: `Tenant #${st.tenant_id}`, status: 'ACTIVE' };
        const sDevs = devsBySite[st.id] || [];
        let lat = null;
        let lng = null;
        let hasCoords = false;
        let locSrc = st.location_source || 'UNCONFIGURED';

        if (isValidStationCoordinate(st.latitude, st.longitude)) {
          lat = Number(st.latitude);
          lng = Number(st.longitude);
          hasCoords = true;
        } else {
          const fbDev = sDevs.find(d => isValidStationCoordinate(d.latitude, d.longitude));
          if (fbDev) {
            lat = Number(fbDev.latitude);
            lng = Number(fbDev.longitude);
            hasCoords = true;
            locSrc = 'DEVICE_FALLBACK';
          }
        }

        let enabledDevs = 0;
        let onlineDevs = 0;
        let explicitOfflineDevs = 0;
        let latestTransMs = 0;
        let latestDataAt = null;
        let latestSeenAt = null;
        let primaryProto = '';
        let timeoutSec = 300;

        sDevs.forEach((d) => {
          const disabled = d.enabled === false || ['DISABLED', 'INACTIVE'].includes(String(d.status || '').toUpperCase());
          if (disabled) return;
          enabledDevs += 1;
          const proto = String((d.connection && d.connection.protocol) || d.device_type || '').toUpperCase();
          if (!primaryProto && proto) primaryProto = proto;
          const connT = Number(d.connection && d.connection.timeout) || 0;
          if (connT >= 10 && connT !== 1000 && connT <= 86400) timeoutSec = Math.max(timeoutSec, connT);

          const dDataMs = d.last_data_at ? new Date(d.last_data_at).getTime() : 0;
          const dSeenMs = (d.last_seen_at || d.last_communication) ? new Date(d.last_seen_at || d.last_communication).getTime() : 0;
          if (dDataMs > latestTransMs) {
            latestTransMs = dDataMs;
            latestDataAt = d.last_data_at;
          }
          if (dSeenMs > 0 && (!latestSeenAt || dSeenMs > new Date(latestSeenAt).getTime())) {
            latestSeenAt = d.last_seen_at || d.last_communication;
            if (!latestTransMs) latestTransMs = dSeenMs;
          }
          const connSt = String(d.connection_status || '').toUpperCase();
          if (connSt === 'OFFLINE' || connSt === 'ERROR') explicitOfflineDevs += 1;
          else if (connSt === 'ONLINE') onlineDevs += 1;
        });

        const lastTransIso = latestDataAt || latestSeenAt || null;
        const ageSec = latestTransMs > 0 ? Math.max(0, Math.floor((nowMs - latestTransMs) / 1000)) : null;
        let deliveryStatus = 'ONLINE';
        if (String(ten.status || '').toUpperCase() === 'SUSPENDED' || st.enabled === false || enabledDevs === 0) {
          deliveryStatus = 'INACTIVE';
        } else if (!lastTransIso) {
          deliveryStatus = onlineDevs > 0 ? 'DELAYED' : 'OFFLINE';
        } else if ((onlineDevs === 0 && explicitOfflineDevs === enabledDevs) || ageSec > Math.max(timeoutSec * 3, 1800)) {
          deliveryStatus = 'OFFLINE';
        } else if (ageSec > timeoutSec || (onlineDevs > 0 && onlineDevs < enabledDevs)) {
          deliveryStatus = 'DELAYED';
        }

        const sAlarms = alarmsBySite[st.id] || [];
        let critCount = 0;
        let warnCount = 0;
        const alarmItems = sAlarms.slice(0, 5).map((al) => {
          const sev = String(al.severity || 'WARNING').toUpperCase();
          if (sev === 'CRITICAL') critCount += 1;
          else warnCount += 1;
          return {
            id: al.id,
            alarm_code: al.alarm_code,
            severity: sev,
            status: al.status,
            device_code: (al.device && al.device.device_code) || '',
            parameter_name: (al.parameter && al.parameter.parameter_name) || '',
            message: al.message,
            triggered_at: al.triggered_at,
          };
        });
        if (sAlarms.length > 5) {
          sAlarms.slice(5).forEach((al) => {
            if (String(al.severity || '').toUpperCase() === 'CRITICAL') critCount += 1;
            else warnCount += 1;
          });
        }
        const totalAlarms = critCount + warnCount;
        let severityState = 'NORMAL';
        if (deliveryStatus === 'INACTIVE') severityState = 'INACTIVE';
        else if (critCount > 0 || deliveryStatus === 'OFFLINE') severityState = 'CRITICAL';
        else if (totalAlarms > 0 || deliveryStatus === 'DELAYED') severityState = 'WARNING';

        sites.push({
          site_id: st.id,
          site_code: st.code,
          site_name: st.name,
          tenant_id: ten.id,
          tenant_code: ten.code,
          tenant_name: ten.name,
          tenant_status: ten.status || 'ACTIVE',
          customer_name: ten.customer_name || '',
          project_year: Number(st.project_year) || Number(ten.project_year) || 0,
          monitoring_domain: String(st.monitoring_domain || 'AQMS').toUpperCase(),
          enabled: st.enabled !== false,
          address: st.address || '',
          timezone: st.timezone || 'Asia/Jakarta',
          has_coordinates: hasCoords,
          latitude: lat,
          longitude: lng,
          location_source: locSrc,
          device_count: sDevs.length,
          enabled_device_count: enabledDevs,
          online_device_count: onlineDevs,
          offline_device_count: Math.max(0, enabledDevs - onlineDevs),
          primary_protocol: primaryProto,
          effective_timeout_seconds: timeoutSec,
          last_data_at: latestDataAt,
          last_seen_at: latestSeenAt,
          last_transmission_at: lastTransIso,
          data_age_seconds: ageSec,
          delivery_status: deliveryStatus,
          severity_state: severityState,
          critical_alarm_count: critCount,
          warning_alarm_count: warnCount,
          total_active_alarms: totalAlarms,
          active_alarms: alarmItems,
          devices: [],
          weather: this.extractFallbackSiteWeather(sDevs, String(st.monitoring_domain || 'AQMS').toUpperCase()),
        });

        summary.total_sites += 1;
        if (hasCoords) summary.mapped_sites += 1;
        else summary.no_gps_sites += 1;
        if (deliveryStatus === 'ONLINE') summary.online_sites += 1;
        else if (deliveryStatus === 'DELAYED') summary.delayed_sites += 1;
        else if (deliveryStatus === 'OFFLINE') summary.offline_sites += 1;
        else summary.inactive_sites += 1;
        if (totalAlarms > 0) {
          summary.sites_with_alarms += 1;
          summary.total_active_alarms += totalAlarms;
          summary.critical_alarms += critCount;
          summary.warning_alarms += warnCount;
        }
        summary.total_devices += enabledDevs;
        summary.online_devices += onlineDevs;
        if (tenantStats[ten.id]) {
          tenantStats[ten.id].site_count += 1;
          if (deliveryStatus === 'ONLINE') tenantStats[ten.id].online_sites += 1;
          if (totalAlarms > 0) tenantStats[ten.id].alarm_sites += 1;
        }
      });

      const sevOrder = { CRITICAL: 0, WARNING: 1, NORMAL: 2, INACTIVE: 3 };
      sites.sort((a, b) => (sevOrder[a.severity_state] - sevOrder[b.severity_state]) || (b.critical_alarm_count - a.critical_alarm_count));

      return {
        generated_at: new Date().toISOString(),
        summary,
        tenants: tenantOpts,
        sites,
      };
    },
    async fetchFleetOverview(manual = false) {
      if (this.loading && !manual) return;
      this.loading = true;
      try {
        const res = await axios.get('/api/tenants/fleet-overview');
        const payload = res.data && res.data.data;
        if (payload) {
          this.applyOverviewPayload(payload);
        }
      } catch (err) {
        if (err && err.response && err.response.status === 404) {
          try {
            const fallbackPayload = await this.fetchFleetOverviewFallback();
            this.applyOverviewPayload(fallbackPayload);
            return;
          } catch (fbErr) {
            this.errorMessage =
              (fbErr.response && fbErr.response.data && (fbErr.response.data.error || fbErr.response.data.message)) ||
              fbErr.message ||
              'Gagal memuat ringkasan peta stasiun global.';
            return;
          }
        }
        this.errorMessage =
          (err.response && err.response.data && (err.response.data.error || err.response.data.message)) ||
          err.message ||
          'Gagal memuat ringkasan peta stasiun global.';
      } finally {
        this.loading = false;
      }
    },

    initLeafletMap() {
      const el = this.$refs.fleetMapCanvas;
      if (!el || this.map) return;

      this.map = L.map(el, {
        center: INDONESIA_DEFAULT_CENTER,
        zoom: INDONESIA_DEFAULT_ZOOM,
        zoomControl: false,
        attributionControl: false,
        fadeAnimation: true,
      });

      L.control.scale({ imperial: false, maxWidth: 100, position: 'bottomright' }).addTo(this.map);
      L.control.zoom({ position: 'bottomright' }).addTo(this.map);
      this.markersLayer = L.layerGroup().addTo(this.map);
      this.weatherCanvasOverlay = new FleetWeatherCanvasOverlay(this.map, {
        L,
        showWind: this.showWindLayer && this.isAqmsWeatherToggleVisible,
        showRain: this.showRainLayer && this.isAqmsWeatherToggleVisible,
        isDark: this.resolvedTileTheme === 'dark' || this.resolvedTileTheme === 'night_clean',
      });

      this.map.on('zoomstart', () => {
        this.spiderfiedClusterKey = '';
      });
      this.map.on('zoomend', () => {
        this.syncMapMarkers();
      });
      this.map.on('click', () => {
        if (this.spiderfiedClusterKey) {
          this.spiderfiedClusterKey = '';
          this.syncMapMarkers();
        }
      });

      this.applyTileLayer();
    },

    applyTilePaneFilter(filterClass) {
      if (!this.map) return;
      const tilePane = this.map.getPane('tilePane');
      if (!tilePane || !tilePane.classList) return;
      TILE_PANE_FILTER_CLASSES.forEach((cls) => tilePane.classList.remove(cls));
      if (filterClass) {
        tilePane.classList.add(filterClass);
      }
    },

    applyTileLayer() {
      if (!this.map) return;
      if (this.tileLayer) {
        this.map.removeLayer(this.tileLayer);
        this.tileLayer = null;
      }
      if (this.tileOverlayLayer) {
        this.map.removeLayer(this.tileOverlayLayer);
        this.tileOverlayLayer = null;
      }
      const cfg = TILE_LAYERS[this.resolvedTileTheme] || TILE_LAYERS.osm;
      this.applyTilePaneFilter(cfg.paneFilterClass || '');
      let fallbackAttempted = false;

      const attachLayer = (url, opts) => {
        const layer = L.tileLayer(url, opts);
        layer.on('load', () => {
          this.isTilesOnline = true;
        });
        layer.on('tileerror', () => {
          if (!fallbackAttempted && cfg.fallbackUrl && url !== cfg.fallbackUrl) {
            fallbackAttempted = true;
            if (this.map && this.tileLayer) {
              this.map.removeLayer(this.tileLayer);
            }
            if (this.map && this.tileOverlayLayer) {
              this.map.removeLayer(this.tileOverlayLayer);
              this.tileOverlayLayer = null;
            }
            this.applyTilePaneFilter(cfg.fallbackPaneFilterClass || cfg.paneFilterClass || '');
            this.tileLayer = attachLayer(cfg.fallbackUrl, { maxZoom: 19 });
            if (this.map) this.tileLayer.addTo(this.map);
          } else {
            this.isTilesOnline = false;
          }
        });
        return layer;
      };

      this.tileLayer = attachLayer(cfg.url, cfg.options);
      this.tileLayer.addTo(this.map);

      if (cfg.overlayUrl) {
        this.tileOverlayLayer = L.tileLayer(cfg.overlayUrl, {
          maxZoom: cfg.options && cfg.options.maxZoom ? cfg.options.maxZoom : 18,
          pane: 'tilePane',
        });
        this.tileOverlayLayer.addTo(this.map);
      }
    },

    resolveSiteEnvPinMeta(site) {
      const fallback = {
        available: false,
        label: '–',
        bg: '#64748b',
        haloBg: 'rgba(100, 116, 139, 0.22)',
        borderColor: '',
        pulse: false,
        urgencyRank: 0,
        numericScore: 0,
        clusterTone: 'INACTIVE',
        tooltipLine: 'Indeks belum tersedia',
      };
      if (!site) return fallback;
      const dom = String(site.monitoring_domain || 'AQMS').toUpperCase();

      if (dom === 'AQMS' && site.ispu && site.ispu.available) {
        const score = Math.round(Number(site.ispu.score) || 0);
        const cat = String(site.ispu.category || '').toUpperCase();
        const catLabel = site.ispu.category_label || cat;
        const domPoll = site.ispu.dominant_pollutant ? ` · ${site.ispu.dominant_pollutant}` : '';
        let bg = '#10b981';
        let haloBg = 'rgba(16, 185, 129, 0.28)';
        let borderColor = '';
        let pulse = false;
        let urgencyRank = 1;
        let clusterTone = 'NORMAL';

        if (cat === 'SEDANG') {
          bg = '#0284c7';
          haloBg = 'rgba(2, 132, 199, 0.30)';
          urgencyRank = 2;
          clusterTone = 'SEDANG';
        } else if (cat === 'TIDAK_SEHAT') {
          bg = '#d97706';
          haloBg = 'rgba(245, 158, 11, 0.34)';
          urgencyRank = 3;
          clusterTone = 'WARNING';
        } else if (cat === 'SANGAT_TIDAK_SEHAT') {
          bg = '#e11d48';
          haloBg = 'rgba(225, 29, 72, 0.38)';
          pulse = true;
          urgencyRank = 4;
          clusterTone = 'CRITICAL';
        } else if (cat === 'BERBAHAYA') {
          bg = '#0f172a';
          borderColor = '#ef4444';
          haloBg = 'rgba(239, 68, 68, 0.44)';
          pulse = true;
          urgencyRank = 5;
          clusterTone = 'HAZARDOUS';
        }

        return {
          available: true,
          label: String(score),
          bg,
          haloBg,
          borderColor,
          pulse,
          urgencyRank,
          numericScore: score,
          clusterTone,
          tooltipLine: `ISPU ${score} (${catLabel}${domPoll})`,
        };
      }

      if ((dom === 'WQMS' || dom === 'ONLIMO') && site.wqms && site.wqms.available) {
        if (this.wqmsIndexMode === 'ika') {
          const ika = Math.round(Number(site.wqms.ika_score) || 0);
          const cat = String(site.wqms.ika_category || '').toUpperCase();
          const catLabel = site.wqms.ika_category_label || cat;
          let bg = '#059669';
          let haloBg = 'rgba(16, 185, 129, 0.28)';
          let pulse = false;
          let urgencyRank = 1;
          let clusterTone = 'NORMAL';

          if (cat === 'BAIK') {
            bg = '#0284c7';
            haloBg = 'rgba(2, 132, 199, 0.30)';
            urgencyRank = 2;
            clusterTone = 'SEDANG';
          } else if (cat === 'SEDANG') {
            bg = '#d97706';
            haloBg = 'rgba(245, 158, 11, 0.34)';
            urgencyRank = 3;
            clusterTone = 'WARNING';
          } else if (cat === 'KURANG') {
            bg = '#ea580c';
            haloBg = 'rgba(234, 88, 12, 0.35)';
            urgencyRank = 4;
            clusterTone = 'ORANGE';
          } else if (cat === 'SANGAT_KURANG') {
            bg = '#e11d48';
            haloBg = 'rgba(225, 29, 72, 0.38)';
            pulse = true;
            urgencyRank = 5;
            clusterTone = 'CRITICAL';
          }

          return {
            available: true,
            label: String(ika),
            bg,
            haloBg,
            borderColor: '',
            pulse,
            urgencyRank,
            numericScore: 100 - ika,
            clusterTone,
            tooltipLine: `IKA ${ika}/100 (${catLabel}) · IP ${Number(site.wqms.ip_score || 0).toFixed(2)}`,
          };
        }

        const ip = Number(site.wqms.ip_score) || 0;
        const ipText = ip < 10 ? ip.toFixed(1) : String(Math.round(ip));
        const cat = String(site.wqms.ip_category || '').toUpperCase();
        const catLabel = site.wqms.ip_category_label || cat;
        const domParam = site.wqms.dominant_parameter ? ` · ${site.wqms.dominant_parameter}` : '';
        let bg = '#0d9488';
        let haloBg = 'rgba(13, 148, 136, 0.28)';
        let pulse = false;
        let urgencyRank = 1;
        let clusterTone = 'NORMAL';

        if (cat === 'CEMAR_RINGAN') {
          bg = '#d97706';
          haloBg = 'rgba(245, 158, 11, 0.34)';
          urgencyRank = 3;
          clusterTone = 'WARNING';
        } else if (cat === 'CEMAR_SEDANG') {
          bg = '#ea580c';
          haloBg = 'rgba(234, 88, 12, 0.35)';
          urgencyRank = 4;
          clusterTone = 'ORANGE';
        } else if (cat === 'CEMAR_BERAT') {
          bg = '#e11d48';
          haloBg = 'rgba(225, 29, 72, 0.38)';
          pulse = true;
          urgencyRank = 5;
          clusterTone = 'CRITICAL';
        }

        return {
          available: true,
          label: ipText,
          bg,
          haloBg,
          borderColor: '',
          pulse,
          urgencyRank,
          numericScore: ip * 20,
          clusterTone,
          tooltipLine: `IP ${ip.toFixed(2)} (${catLabel}${domParam}) · IKA ${Math.round(site.wqms.ika_score || 0)}`,
        };
      }

      return fallback;
    },

    isSiteAqmsWeatherEligible(site) {
      return Boolean(site && String(site.monitoring_domain || '').toUpperCase() === 'AQMS');
    },

    extractFallbackSiteWeather(siteDevs, domain) {
      if (domain !== 'AQMS' || !Array.isArray(siteDevs) || siteDevs.length === 0) return null;
      const excludeRe = /(^|[\s_\-.])(cpu|gpu|soc|board|pcb|box|panel|cabinet|shelter|enclosure|rack|indoor|internal|pump|motor|oil|water|wastewater|effluent|inlet|outlet|diff|filter|sample|vacuum|exhaust|battery)([\s_\-.]|$)/i;
      const wsRe = /(^|[\s_\-.])(ws|wspeed|w[_\s-]?speed|wind[_\s-]?speed|windspeed|kec[_\s-]?angin|kecepatan[_\s-]?angin)([\s_\-.]|$)/i;
      const wdRe = /(^|[\s_\-.])(wd|wdir|w[_\s-]?dir|wind[_\s-]?dir|winddir|wind[_\s-]?angle|windangle|arah[_\s-]?angin|sudut[_\s-]?angin)([\s_\-.]|$)/i;
      const rfRe = /(^|[\s_\-.])(rain|rainfall|rain[_\s-]?fall|curah[_\s-]?hujan|hujan|precip|precipitation)([\s_\-.]|$)/i;
      let windSpeed = null;
      let windSpeedUnit = 'm/s';
      let windAngle = null;
      let rainFall = null;
      let rainFallUnit = 'mm';
      for (const dev of siteDevs) {
        if (!dev || dev.enabled === false) continue;
        const params = Array.isArray(dev.parameters) ? dev.parameters : [];
        for (const p of params) {
          if (!p || p.enabled === false) continue;
          const rawVal = p.current_formula_value ?? p.current_processed_value ?? p.current_value ?? p.current_value_numeric;
          const num = Number(rawVal);
          if (rawVal == null || rawVal === '' || !Number.isFinite(num)) continue;
          const combined = `${p.parameter_code || p.code || ''} ${p.parameter_name || p.name || ''}`.trim();
          if (!combined || excludeRe.test(combined)) continue;
          if (windSpeed == null && wsRe.test(combined)) {
            windSpeed = Math.max(0, Math.round(num * 100) / 100);
            if (p.unit) windSpeedUnit = p.unit;
          } else if (windAngle == null && wdRe.test(combined)) {
            windAngle = Math.round((((num % 360) + 360) % 360) * 10) / 10;
          } else if (rainFall == null && rfRe.test(combined)) {
            rainFall = Math.max(0, Math.round(num * 100) / 100);
            if (p.unit) rainFallUnit = p.unit;
          }
        }
      }
      if (windSpeed == null && windAngle == null && rainFall == null) return null;
      return {
        available: true,
        wind_speed: windSpeed,
        wind_speed_unit: windSpeedUnit,
        wind_angle: windAngle,
        rain_fall: rainFall,
        rain_fall_unit: rainFallUnit,
        is_raining: rainFall != null && rainFall > 0.05,
      };
    },

    formatWindCompassShort(angle) {
      const num = Number(angle);
      if (!Number.isFinite(num)) {
        return { deg: 0, idShort: '-', label: '—', valid: false };
      }
      const norm = ((num % 360) + 360) % 360;
      const pts = [
        { idShort: 'U', label: 'Utara' },
        { idShort: 'TL', label: 'Timur Laut' },
        { idShort: 'T', label: 'Timur' },
        { idShort: 'TG', label: 'Tenggara' },
        { idShort: 'S', label: 'Selatan' },
        { idShort: 'BD', label: 'Barat Daya' },
        { idShort: 'B', label: 'Barat' },
        { idShort: 'BL', label: 'Barat Laut' },
      ];
      const pt = pts[Math.round(norm / 45) % 8];
      return { deg: Math.round(norm), idShort: pt.idShort, label: pt.label, valid: true };
    },

    resolveWindSpeedMeta(speed, unit = 'm/s') {
      const num = Number(speed);
      if (!Number.isFinite(num)) {
        return { color: '#0284c7', pillBg: 'rgba(2, 132, 199, 0.94)', ms: 0, label: 'Angin' };
      }
      const u = String(unit || 'm/s').toLowerCase();
      let ms = num;
      if (u.includes('km/h') || u.includes('kmh') || u.includes('kph')) ms = num / 3.6;
      else if (u.includes('knot') || u === 'kt') ms = num * 0.514444;
      else if (u.includes('mph')) ms = num * 0.44704;

      if (ms > 10) return { color: '#e11d48', pillBg: 'rgba(225, 29, 72, 0.95)', ms, label: 'Sangat Kencang' };
      if (ms >= 5) return { color: '#d97706', pillBg: 'rgba(217, 119, 6, 0.95)', ms, label: 'Kencang' };
      if (ms >= 2) return { color: '#059669', pillBg: 'rgba(5, 150, 105, 0.94)', ms, label: 'Sedang' };
      return { color: '#0284c7', pillBg: 'rgba(2, 132, 199, 0.94)', ms, label: 'Sepoi' };
    },

    buildWindArrowSvg(angle, color, pinSize) {
      const norm = Number.isFinite(Number(angle)) ? ((Number(angle) % 360) + 360) % 360 : 0;
      const orbitSize = pinSize + 28;
      return `
        <div class="fleet-wind-orbit" style="width:${orbitSize}px;height:${orbitSize}px;">
          <div class="fleet-wind-arrow-rot" style="transform:rotate(${norm}deg);">
            <svg viewBox="0 0 28 28" width="20" height="20" class="fleet-wind-arrow-svg" aria-hidden="true">
              <path d="M14 1.5 L21.5 16.5 L14 12.8 L6.5 16.5 Z" fill="${color}" stroke="#ffffff" stroke-width="2" stroke-linejoin="round" />
              <line x1="14" y1="12.5" x2="14" y2="22.5" stroke="${color}" stroke-width="2.6" stroke-linecap="round" />
            </svg>
          </div>
        </div>
      `;
    },

    syncWeatherCanvasLayer() {
      if (!this.weatherCanvasOverlay) return;
      const allowWeather = this.isAqmsWeatherToggleVisible;
      const isDarkMap =
        this.resolvedTileTheme === 'dark' ||
        this.resolvedTileTheme === 'night_clean' ||
        this.resolvedTileTheme === 'satellite';
      this.weatherCanvasOverlay.updateConfig({
        sites: this.mappedFilteredSites || [],
        showWind: Boolean(allowWeather && this.showWindLayer),
        showRain: Boolean(allowWeather && this.showRainLayer),
        isDark: isDarkMap,
      });
    },

    buildSiteWeatherOverlayHtml(site, pinSize) {
      if (!this.isSiteAqmsWeatherEligible(site) || !site.weather || !site.weather.available) {
        return { rippleHtml: '', arrowHtml: '', pillStripHtml: '' };
      }
      const w = site.weather;
      const pills = [];

      if (this.showWindLayer && (w.wind_speed != null || w.wind_angle != null)) {
        const wsMeta = this.resolveWindSpeedMeta(w.wind_speed, w.wind_speed_unit);
        const dir = this.formatWindCompassShort(w.wind_angle);
        const speedTxt =
          w.wind_speed != null
            ? `${Number(w.wind_speed).toFixed(1)} ${this.escapeHtml(w.wind_speed_unit || 'm/s')}`
            : '';
        const dirTxt = dir.valid ? dir.idShort : '';
        const combinedTxt = [speedTxt, dirTxt].filter(Boolean).join(' · ');
        if (combinedTxt) {
          pills.push(
            `<span class="fleet-weather-pill fleet-weather-pill-wind" style="background:${wsMeta.pillBg};">${combinedTxt}</span>`
          );
        }
      }

      if (this.showRainLayer && w.is_raining) {
        const rfVal = Number(w.rain_fall || 0).toFixed(1);
        const rfUnit = this.escapeHtml(w.rain_fall_unit || 'mm');
        pills.push(
          `<span class="fleet-weather-pill fleet-weather-pill-rain">🌧 ${rfVal} ${rfUnit}</span>`
        );
      }

      const pillStripHtml =
        pills.length > 0
          ? `<div class="fleet-weather-pill-strip" style="top:${pinSize + 4}px;">${pills.join('')}</div>`
          : '';

      return { rippleHtml: '', arrowHtml: '', pillStripHtml };
    },

    summarizeClusterWeather(sites) {
      let windCount = 0;
      let speedSum = 0;
      let speedCount = 0;
      let sinSum = 0;
      let cosSum = 0;
      let angleCount = 0;
      let windUnit = 'm/s';

      let rainingCount = 0;
      let maxRain = 0;
      let rainUnit = 'mm';

      for (const s of sites || []) {
        if (!this.isSiteAqmsWeatherEligible(s) || !s.weather || !s.weather.available) continue;
        const w = s.weather;
        if (w.wind_speed != null || w.wind_angle != null) {
          windCount += 1;
          if (w.wind_speed != null && Number.isFinite(Number(w.wind_speed))) {
            speedSum += Number(w.wind_speed);
            speedCount += 1;
            if (w.wind_speed_unit) windUnit = w.wind_speed_unit;
          }
          if (w.wind_angle != null && Number.isFinite(Number(w.wind_angle))) {
            const rad = (Number(w.wind_angle) * Math.PI) / 180;
            sinSum += Math.sin(rad);
            cosSum += Math.cos(rad);
            angleCount += 1;
          }
        }
        if (w.is_raining) {
          rainingCount += 1;
          const rf = Number(w.rain_fall) || 0;
          if (rf > maxRain) maxRain = rf;
          if (w.rain_fall_unit) rainUnit = w.rain_fall_unit;
        }
      }

      let avgAngle = null;
      if (angleCount > 0) {
        const deg = (Math.atan2(sinSum / angleCount, cosSum / angleCount) * 180) / Math.PI;
        avgAngle = ((deg % 360) + 360) % 360;
      }
      const avgSpeed = speedCount > 0 ? speedSum / speedCount : null;

      return {
        hasWind: windCount > 0 && (avgSpeed != null || avgAngle != null),
        windCount,
        avgSpeed,
        windUnit,
        avgAngle,
        hasRain: rainingCount > 0,
        rainingCount,
        maxRain,
        rainUnit,
      };
    },

    buildClusterWeatherOverlayHtml(group, outerSize) {
      if (!this.showWindLayer && !this.showRainLayer) {
        return { rippleHtml: '', arrowHtml: '', pillStripHtml: '' };
      }
      const cw = this.summarizeClusterWeather(group.sites);
      const pills = [];

      if (this.showWindLayer && cw.hasWind) {
        const wsMeta = this.resolveWindSpeedMeta(cw.avgSpeed, cw.windUnit);
        const dir = this.formatWindCompassShort(cw.avgAngle);
        const speedTxt =
          cw.avgSpeed != null
            ? `${Number(cw.avgSpeed).toFixed(1)} ${this.escapeHtml(cw.windUnit || 'm/s')}`
            : '';
        const dirTxt = dir.valid ? dir.idShort : '';
        const combinedTxt = [speedTxt, dirTxt].filter(Boolean).join(' · ');
        if (combinedTxt) {
          pills.push(
            `<span class="fleet-weather-pill fleet-weather-pill-wind" style="background:${wsMeta.pillBg};">${combinedTxt}</span>`
          );
        }
      }

      if (this.showRainLayer && cw.hasRain) {
        const rfTxt =
          cw.rainingCount > 1
            ? `🌧 ${cw.rainingCount} · ${Number(cw.maxRain).toFixed(1)}${this.escapeHtml(cw.rainUnit || 'mm')}`
            : `🌧 ${Number(cw.maxRain).toFixed(1)} ${this.escapeHtml(cw.rainUnit || 'mm')}`;
        pills.push(`<span class="fleet-weather-pill fleet-weather-pill-rain">${rfTxt}</span>`);
      }

      const pillStripHtml =
        pills.length > 0
          ? `<div class="fleet-weather-pill-strip" style="top:${outerSize + 3}px;">${pills.join('')}</div>`
          : '';

      return { rippleHtml: '', arrowHtml: '', pillStripHtml };
    },

    buildSiteWeatherTooltipHtml(site) {
      if (!this.isSiteAqmsWeatherEligible(site) || !site.weather || !site.weather.available) return '';
      const w = site.weather;
      const parts = [];
      if (w.wind_speed != null || w.wind_angle != null) {
        const dir = this.formatWindCompassShort(w.wind_angle);
        const wsStr = w.wind_speed != null ? `${Number(w.wind_speed).toFixed(1)} ${this.escapeHtml(w.wind_speed_unit || 'm/s')}` : '';
        const wdStr = dir.valid ? `${dir.deg}° (${dir.label})` : '';
        parts.push(`🌬 ${[wsStr, wdStr].filter(Boolean).join(' · ')}`);
      }
      if (w.rain_fall != null) {
        parts.push(
          w.is_raining
            ? `🌧 Hujan ${Number(w.rain_fall).toFixed(1)} ${this.escapeHtml(w.rain_fall_unit || 'mm')}`
            : `☀ Cerah (0 ${this.escapeHtml(w.rain_fall_unit || 'mm')})`
        );
      }
      if (parts.length === 0) return '';
      return `<br/><span style="font-size:10px;font-weight:700;color:#0369a1;">${parts.join(' &nbsp;|&nbsp; ')}</span>`;
    },

    buildClusterWeatherTooltipHtml(group) {
      const cw = this.summarizeClusterWeather(group.sites);
      const parts = [];
      if (cw.hasWind) {
        const dir = this.formatWindCompassShort(cw.avgAngle);
        const wsStr = cw.avgSpeed != null ? `Rata2 ${Number(cw.avgSpeed).toFixed(1)} ${this.escapeHtml(cw.windUnit || 'm/s')}` : '';
        const wdStr = dir.valid ? `${dir.deg}° (${dir.label})` : '';
        parts.push(`🌬 ${[wsStr, wdStr].filter(Boolean).join(' · ')}`);
      }
      if (cw.hasRain) {
        parts.push(`🌧 ${cw.rainingCount} Stasiun Hujan (maks ${Number(cw.maxRain).toFixed(1)} ${this.escapeHtml(cw.rainUnit || 'mm')})`);
      }
      if (parts.length === 0) return '';
      return `<br/><span style="font-size:10px;font-weight:700;color:#0369a1;">${parts.join(' &nbsp;|&nbsp; ')}</span>`;
    },

    buildPinIcon(site, isSelected) {
      if (this.pinDisplayMode === 'env') {
        const meta = this.resolveSiteEnvPinMeta(site);
        const size = isSelected ? 34 : 28;
        const borderCol = isSelected ? '#38bdf8' : meta.borderColor || '#ffffff';
        const shadowCss = isSelected
          ? `0 0 0 4px rgba(14, 165, 233, 0.45), 0 6px 16px rgba(15, 23, 42, 0.48)`
          : `0 0 0 3px ${meta.haloBg}, 0 3px 10px rgba(15, 23, 42, 0.40)`;
        const pulseHtml = meta.pulse
          ? `<span class="fleet-pin-pulse" style="background:${meta.haloBg};"></span>`
          : '';
        const fontSize = meta.label.length >= 4 ? '8.5px' : meta.label.length === 3 ? '9.5px' : '10.5px';
        const wxOverlay = this.buildSiteWeatherOverlayHtml(site, size);

        const html = `
          <div class="fleet-pin-root" style="width:${size}px;height:${size}px;">
            ${wxOverlay.rippleHtml}
            ${wxOverlay.arrowHtml}
            ${pulseHtml}
            <div class="fleet-pin-core" style="width:${size}px;height:${size}px;background:${meta.bg};border:2.5px solid ${borderCol};box-shadow:${shadowCss};">
              <span class="fleet-pin-score" style="font-size:${fontSize};">${meta.label}</span>
            </div>
            ${wxOverlay.pillStripHtml}
          </div>
        `;

        return L.divIcon({
          className: 'fleet-leaflet-div-icon',
          html,
          iconSize: [size, size],
          iconAnchor: [size / 2, size / 2],
        });
      }

      const sev = site.severity_state || 'NORMAL';
      let coreBg = '#10b981'; // emerald
      let ringBg = 'rgba(16, 185, 129, 0.28)';
      let pulseHtml = '';

      if (sev === 'CRITICAL') {
        coreBg = '#e11d48'; // rose-600
        ringBg = 'rgba(225, 29, 72, 0.35)';
        pulseHtml = `<span class="fleet-pin-pulse" style="background:${ringBg};"></span>`;
      } else if (sev === 'WARNING') {
        coreBg = '#f59e0b'; // amber-500
        ringBg = 'rgba(245, 158, 11, 0.30)';
      } else if (sev === 'INACTIVE') {
        coreBg = '#64748b'; // slate-500
        ringBg = 'rgba(100, 116, 139, 0.22)';
      }

      const size = isSelected ? 30 : 24;
      const borderCol = isSelected ? '#0ea5e9' : '#ffffff';
      const shadowCss = isSelected
        ? '0 0 0 4px rgba(14, 165, 233, 0.42), 0 6px 16px rgba(15, 23, 42, 0.45)'
        : '0 3px 10px rgba(15, 23, 42, 0.38)';

      const domShort =
        site.monitoring_domain === 'SPARING'
          ? 'S'
          : site.monitoring_domain === 'WQMS'
          ? 'W'
          : site.monitoring_domain === 'ONLIMO'
          ? 'O'
          : 'A';

      const alarmCount = Number(site.total_active_alarms) || 0;
      const badgeLabel = alarmCount > 9 ? '9+' : String(alarmCount);
      const alarmBadgeHtml =
        alarmCount > 0
          ? `<span class="fleet-pin-badge">${badgeLabel}</span>`
          : '';
      const wxOverlay = this.buildSiteWeatherOverlayHtml(site, size);

      const html = `
        <div class="fleet-pin-root" style="width:${size}px;height:${size}px;">
          ${wxOverlay.rippleHtml}
          ${wxOverlay.arrowHtml}
          ${pulseHtml}
          <div class="fleet-pin-core" style="width:${size}px;height:${size}px;background:${coreBg};border:2.5px solid ${borderCol};box-shadow:${shadowCss};">
            <span class="fleet-pin-letter">${domShort}</span>
          </div>
          ${alarmBadgeHtml}
          ${wxOverlay.pillStripHtml}
        </div>
      `;

      return L.divIcon({
        className: 'fleet-leaflet-div-icon',
        html,
        iconSize: [size, size],
        iconAnchor: [size / 2, size / 2],
      });
    },

    buildClusterIcon(group) {
      const count = group.sites.length;
      const sev = group.worstSeverity || 'NORMAL';
      let coreBg = '#059669'; // emerald-600
      let haloBg = 'rgba(16, 185, 129, 0.28)';
      let borderCol = '#ffffff';
      let pulseHtml = '';

      if (sev === 'HAZARDOUS') {
        coreBg = '#0f172a';
        borderCol = '#ef4444';
        haloBg = 'rgba(239, 68, 68, 0.40)';
        pulseHtml = `<span class="fleet-pin-pulse" style="background:${haloBg};"></span>`;
      } else if (sev === 'CRITICAL') {
        coreBg = '#e11d48'; // rose-600
        haloBg = 'rgba(225, 29, 72, 0.34)';
        pulseHtml = `<span class="fleet-pin-pulse" style="background:${haloBg};"></span>`;
      } else if (sev === 'ORANGE') {
        coreBg = '#ea580c';
        haloBg = 'rgba(234, 88, 12, 0.34)';
      } else if (sev === 'WARNING') {
        coreBg = '#d97706'; // amber-600
        haloBg = 'rgba(245, 158, 11, 0.32)';
      } else if (sev === 'SEDANG') {
        coreBg = '#0284c7'; // sky-600
        haloBg = 'rgba(2, 132, 199, 0.30)';
      } else if (sev === 'INACTIVE') {
        coreBg = '#64748b';
        haloBg = 'rgba(100, 116, 139, 0.26)';
      }

      const outerSize = count >= 10 ? 42 : 36;
      const coreSize = count >= 10 ? 32 : 28;
      const alarmCount = Number(group.totalAlarms) || 0;
      const alarmLabel = alarmCount > 99 ? '99+' : String(alarmCount);
      const alarmBadgeHtml =
        this.pinDisplayMode === 'status' && alarmCount > 0
          ? `<span class="fleet-pin-badge" title="${alarmCount} Alarm Aktif">!${alarmLabel}</span>`
          : '';
      const wxOverlay = this.buildClusterWeatherOverlayHtml(group, outerSize);

      const html = `
        <div class="fleet-cluster-root" style="width:${outerSize}px;height:${outerSize}px;background:${haloBg};">
          ${wxOverlay.rippleHtml}
          ${wxOverlay.arrowHtml}
          ${pulseHtml}
          <div class="fleet-cluster-core" style="width:${coreSize}px;height:${coreSize}px;background:${coreBg};border-color:${borderCol};">
            <span class="fleet-cluster-count">${count}</span>
          </div>
          ${alarmBadgeHtml}
          ${wxOverlay.pillStripHtml}
        </div>
      `;

      return L.divIcon({
        className: 'fleet-leaflet-div-icon',
        html,
        iconSize: [outerSize, outerSize],
        iconAnchor: [outerSize / 2, outerSize / 2],
      });
    },

    buildScreenClusters(validSites) {
      const isEnvMode = this.pinDisplayMode === 'env';
      const sevRank = { HAZARDOUS: 5, CRITICAL: 4, ORANGE: 3.5, WARNING: 3, SEDANG: 2, NORMAL: 1, INACTIVE: 0 };
      const getSiteTone = (s) =>
        isEnvMode ? this.resolveSiteEnvPinMeta(s).clusterTone : (s.severity_state || 'NORMAL');

      if (!this.map || validSites.length <= 1) {
        return validSites.map((s) => ({
          key: String(s.site_id),
          lat: Number(s.latitude),
          lng: Number(s.longitude),
          sites: [s],
          worstSeverity: getSiteTone(s),
          totalAlarms: Number(s.total_active_alarms) || 0,
        }));
      }

      const zoom = this.map.getZoom();
      let radiusPx = 46;
      if (zoom >= 14) radiusPx = 28;
      else if (zoom >= 11) radiusPx = 38;
      else if (zoom >= 8) radiusPx = 44;

      const sorted = [...validSites].sort((a, b) => Number(a.site_id) - Number(b.site_id));
      const clusters = [];

      sorted.forEach((site) => {
        const lat = Number(site.latitude);
        const lng = Number(site.longitude);
        const pt = this.map.latLngToLayerPoint([lat, lng]);

        let bestIdx = -1;
        let bestDist = radiusPx;
        for (let i = 0; i < clusters.length; i++) {
          const dx = pt.x - clusters[i].cx;
          const dy = pt.y - clusters[i].cy;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist <= bestDist) {
            bestDist = dist;
            bestIdx = i;
          }
        }

        const sSev = getSiteTone(site);
        const sAlarms = Number(site.total_active_alarms) || 0;

        if (bestIdx >= 0) {
          const c = clusters[bestIdx];
          const n = c.sites.length;
          c.cx = (c.cx * n + pt.x) / (n + 1);
          c.cy = (c.cy * n + pt.y) / (n + 1);
          c.lat = (c.lat * n + lat) / (n + 1);
          c.lng = (c.lng * n + lng) / (n + 1);
          c.sites.push(site);
          c.totalAlarms += sAlarms;
          if ((sevRank[sSev] || 0) > (sevRank[c.worstSeverity] || 0)) {
            c.worstSeverity = sSev;
          }
        } else {
          clusters.push({
            cx: pt.x,
            cy: pt.y,
            lat,
            lng,
            sites: [site],
            worstSeverity: sSev,
            totalAlarms: sAlarms,
          });
        }
      });

      clusters.forEach((c) => {
        c.key = c.sites
          .map(s => Number(s.site_id))
          .sort((a, b) => a - b)
          .join('-');
      });

      return clusters;
    },

    attachSingleStationMarker(site, lat, lng, zBoost = 0) {
      const isSelected = Number(this.selectedSiteId) === Number(site.site_id);
      const envMeta = this.pinDisplayMode === 'env' ? this.resolveSiteEnvPinMeta(site) : null;
      const baseZ = isSelected
        ? 950
        : envMeta
        ? 100 + envMeta.urgencyRank * 100
        : site.severity_state === 'CRITICAL'
        ? 500
        : site.severity_state === 'WARNING'
        ? 300
        : 100;

      const marker = L.marker([lat, lng], {
        icon: this.buildPinIcon(site, isSelected),
        zIndexOffset: baseZ + zBoost,
      });

      const safeName = this.escapeHtml(site.site_name);
      const safeTenant = this.escapeHtml(site.tenant_name);
      const safeAge = this.escapeHtml(this.formatRelativeAge(site));
      const yearSuffix = Number(site.project_year) > 0 ? ` · ${Number(site.project_year)}` : '';
      const envLineHtml =
        envMeta && envMeta.available
          ? `<br/><span style="font-weight:700;color:#0284c7;">${this.escapeHtml(envMeta.tooltipLine)}</span>`
          : '';
      const weatherLineHtml =
        this.showWindLayer || this.showRainLayer ? this.buildSiteWeatherTooltipHtml(site) : '';
      marker.bindTooltip(
        `<div style="font-family:Inter,sans-serif;font-size:11px;line-height:1.35;"><strong>${safeName}</strong><br/><span style="opacity:0.84;">${safeTenant}${yearSuffix} · ${safeAge}</span>${envLineHtml}${weatherLineHtml}</div>`,
        { direction: 'top', offset: [0, -14], opacity: 0.96 }
      );

      marker.on('click', (ev) => {
        L.DomEvent.stopPropagation(ev);
        this.selectedSiteId = site.site_id;
        if (this.map) {
          const targetZoom = Math.max(this.map.getZoom(), 11);
          this.map.flyTo([Number(site.latitude), Number(site.longitude)], targetZoom, { duration: 0.55 });
        }
      });

      marker.addTo(this.markersLayer);
      this.markerBySiteId[site.site_id] = marker;
    },

    syncMapMarkers() {
      if (!this.map || !this.markersLayer) return;
      this.markersLayer.clearLayers();
      this.markerBySiteId = {};

      const validSites = this.mappedFilteredSites.filter((s) =>
        isValidStationCoordinate(Number(s.latitude), Number(s.longitude))
      );
      if (validSites.length === 0) return;

      const zoom = this.map.getZoom();
      const groups = this.buildScreenClusters(validSites);

      groups.forEach((group) => {
        if (group.sites.length === 1) {
          const site = group.sites[0];
          this.attachSingleStationMarker(site, Number(site.latitude), Number(site.longitude), 0);
          return;
        }

        const containsSelected = group.sites.some(s => Number(s.site_id) === Number(this.selectedSiteId));
        const shouldSpiderfy =
          this.spiderfiedClusterKey === group.key || zoom >= 14 || containsSelected;

        if (shouldSpiderfy) {
          const hubCenter = [group.lat, group.lng];
          const centerPt = this.map.latLngToLayerPoint(hubCenter);
          const n = group.sites.length;
          const legPx = Math.min(76, 30 + n * 4.5);

          L.circleMarker(hubCenter, {
            radius: 4,
            color: '#ffffff',
            weight: 1.5,
            fillColor: '#0284c7',
            fillOpacity: 0.95,
          }).addTo(this.markersLayer);

          group.sites.forEach((site, idx) => {
            const angle = (2 * Math.PI * idx) / n - Math.PI / 2;
            const spiderPt = L.point(
              centerPt.x + Math.cos(angle) * legPx,
              centerPt.y + Math.sin(angle) * legPx
            );
            const spiderLatLng = this.map.layerPointToLatLng(spiderPt);

            L.polyline([hubCenter, [spiderLatLng.lat, spiderLatLng.lng]], {
              color: '#0284c7',
              weight: 1.8,
              opacity: 0.78,
              dashArray: '3, 3',
            }).addTo(this.markersLayer);

            this.attachSingleStationMarker(site, spiderLatLng.lat, spiderLatLng.lng, 220);
          });
          return;
        }

        // Render collapsed Cluster Badge
        const zRank =
          group.worstSeverity === 'CRITICAL'
            ? 700
            : group.worstSeverity === 'WARNING'
            ? 550
            : 400;

        const clusterMarker = L.marker([group.lat, group.lng], {
          icon: this.buildClusterIcon(group),
          zIndexOffset: zRank,
        });

        const onlineInCluster = group.sites.filter(s => s.delivery_status === 'ONLINE').length;
        const issueInCluster = group.sites.length - onlineInCluster;
        const previewNames = group.sites
          .slice(0, 4)
          .map(s => `• ${this.escapeHtml(s.site_name)}`)
          .join('<br/>');
        const moreLine =
          group.sites.length > 4
            ? `<br/><span style="opacity:0.75;">+${group.sites.length - 4} stasiun lainnya</span>`
            : '';

        const clusterWeatherHtml =
          this.showWindLayer || this.showRainLayer ? this.buildClusterWeatherTooltipHtml(group) : '';
        clusterMarker.bindTooltip(
          `<div style="font-family:Inter,sans-serif;font-size:11px;line-height:1.4;">` +
            `<strong>📍 ${group.sites.length} Stasiun Berdekatan</strong> ` +
            `<span style="opacity:0.85;">(${onlineInCluster} Online · ${issueInCluster} Perhatian)</span>${clusterWeatherHtml}<br/>` +
            `<span style="font-size:10px;opacity:0.9;">${previewNames}${moreLine}</span><br/>` +
            `<span style="font-size:10px;color:#0284c7;font-weight:700;">Klik untuk perbesar / mekarkan pin</span>` +
            `</div>`,
          { direction: 'top', offset: [0, -16], opacity: 0.97 }
        );

        clusterMarker.on('click', (ev) => {
          L.DomEvent.stopPropagation(ev);
          if (!this.map) return;

          const lats = group.sites.map(s => Number(s.latitude));
          const lngs = group.sites.map(s => Number(s.longitude));
          const latSpan = Math.max(...lats) - Math.min(...lats);
          const lngSpan = Math.max(...lngs) - Math.min(...lngs);

          // If stations span multiple cities at low zoom, zoom into their bounding box first;
          // if they are in the same city/spot or already zoomed in, immediately fan out (spiderfy)!
          if (this.map.getZoom() < 10 && (latSpan > 0.06 || lngSpan > 0.06)) {
            const bounds = L.latLngBounds(group.sites.map(s => [Number(s.latitude), Number(s.longitude)]));
            this.map.flyToBounds(bounds, { padding: [75, 75], maxZoom: 11, duration: 0.55 });
          } else {
            this.spiderfiedClusterKey = group.key;
            this.syncMapMarkers();
          }
        });

        clusterMarker.addTo(this.markersLayer);
      });
    },

    fitMapToVisibleSites() {
      if (!this.map) return;
      this.map.invalidateSize();
      const pts = this.mappedFilteredSites.map(s => [Number(s.latitude), Number(s.longitude)]);
      if (pts.length === 0) {
        this.map.setView(INDONESIA_DEFAULT_CENTER, INDONESIA_DEFAULT_ZOOM);
        return;
      }
      if (pts.length === 1) {
        this.map.flyTo(pts[0], 13, { duration: 0.6 });
        return;
      }
      const bounds = L.latLngBounds(pts);
      this.map.fitBounds(bounds, { padding: [44, 44], maxZoom: 13 });
    },

    focusSite(site) {
      if (!site) return;
      this.selectedSiteId = site.site_id;
      if (site.has_coordinates && isValidStationCoordinate(site.latitude, site.longitude) && this.map) {
        this.map.flyTo([Number(site.latitude), Number(site.longitude)], Math.max(this.map.getZoom(), 12), {
          duration: 0.6,
        });
      }
    },

    jumpMobileToMapPin(site) {
      this.selectedSiteId = site.site_id;
      this.mobileViewMode = 'map';
      this.$nextTick(() => {
        if (this.map) {
          this.map.invalidateSize();
          if (site.has_coordinates && isValidStationCoordinate(site.latitude, site.longitude)) {
            this.map.flyTo([Number(site.latitude), Number(site.longitude)], 13, { duration: 0.55 });
          }
        }
      });
    },

    switchMobileView(mode) {
      this.mobileViewMode = mode;
      if (mode === 'map') {
        this.$nextTick(() => {
          if (this.map) this.map.invalidateSize();
        });
      }
    },

    setQuickFilter(key) {
      this.quickFilter = this.quickFilter === key && key !== 'ALL' ? 'ALL' : key;
      if (this.quickFilter === 'NOGPS' && this.isDesktopViewport) {
        this.dockTab = 'NOGPS';
      } else if (this.quickFilter === 'NORMAL' && this.isDesktopViewport) {
        this.dockTab = 'NORMAL';
      } else if ((this.dockTab === 'NOGPS' && this.quickFilter !== 'NOGPS') || (this.dockTab === 'NORMAL' && this.quickFilter !== 'NORMAL')) {
        this.dockTab = 'ALL';
      }
      this.$nextTick(() => {
        if (this.mappedFilteredSites.length > 0 && this.quickFilter !== 'NOGPS') {
          this.fitMapToVisibleSites();
        }
      });
    },

    resetAllFilters() {
      this.quickFilter = 'ALL';
      this.searchQuery = '';
      this.projectYearFilter = 'ALL';
      this.tenantFilter = 'ALL';
      this.domainFilter = 'ALL';
      this.alarmFilter = 'ALL';
      this.dockTab = 'ALL';
      this.$nextTick(() => this.fitMapToVisibleSites());
    },

    cycleMapStyle() {
      const order = ['auto', 'osm', 'clean', 'satellite', 'dark'];
      const idx = order.indexOf(this.mapStyle);
      this.mapStyle = order[(idx + 1) % order.length];
    },

    setMapStyle(val) {
      this.mapStyle = val;
    },

    refreshMapLayoutAfterFullscreen() {
      this.$nextTick(() => {
        if (this.map) this.map.invalidateSize();
        setTimeout(() => {
          if (this.map) {
            this.map.invalidateSize();
            this.syncMapMarkers();
          }
        }, 180);
      });
    },

    async toggleMapFullscreen() {
      if (this.isMapFullscreen) {
        await this.exitMapFullscreen();
        return;
      }
      this.isMapFullscreen = true;
      const paneEl = this.$refs.fleetMapPane;
      if (paneEl && typeof document !== 'undefined') {
        try {
          if (paneEl.requestFullscreen) {
            await paneEl.requestFullscreen();
          } else if (paneEl.webkitRequestFullscreen) {
            paneEl.webkitRequestFullscreen();
          }
        } catch (_) {
          // Fallback CSS fixed fullscreen remains active cleanly
        }
      }
      this.refreshMapLayoutAfterFullscreen();
    },

    async exitMapFullscreen() {
      this.isMapFullscreen = false;
      if (typeof document !== 'undefined') {
        const fsEl = document.fullscreenElement || document.webkitFullscreenElement;
        if (fsEl) {
          try {
            if (document.exitFullscreen) {
              await document.exitFullscreen();
            } else if (document.webkitExitFullscreen) {
              document.webkitExitFullscreen();
            }
          } catch (_) {
            // ignore native fullscreen exit errors
          }
        }
      }
      this.refreshMapLayoutAfterFullscreen();
    },

    onNativeFullscreenChange() {
      if (typeof document === 'undefined') return;
      const fsEl = document.fullscreenElement || document.webkitFullscreenElement;
      const paneEl = this.$refs.fleetMapPane;
      if (!fsEl && this.isMapFullscreen) {
        this.isMapFullscreen = false;
        this.refreshMapLayoutAfterFullscreen();
      } else if (fsEl && paneEl && fsEl === paneEl && !this.isMapFullscreen) {
        this.isMapFullscreen = true;
        this.refreshMapLayoutAfterFullscreen();
      }
    },

    onGlobalKeydown(ev) {
      if (!ev) return;
      if (ev.key === 'Escape' && this.isMapFullscreen) {
        this.exitMapFullscreen();
      }
    },

    async openSiteDashboard(site) {
      if (!site || this.switchingSiteId) return;
      if (this.isMapFullscreen) {
        await this.exitMapFullscreen();
      }
      this.switchingSiteId = site.site_id;
      try {
        await this.$store.dispatch('setActiveTenantAndSite', {
          tenantId: site.tenant_id,
          siteId: site.site_id,
        });
        this.$router.push('/monitoring/dashboard-operator');
      } finally {
        this.switchingSiteId = 0;
      }
    },

    async openSiteModule(site, targetPath) {
      if (!site || this.switchingSiteId) return;
      if (this.isMapFullscreen) {
        await this.exitMapFullscreen();
      }
      this.switchingSiteId = site.site_id;
      try {
        await this.$store.dispatch('setActiveTenantAndSite', {
          tenantId: site.tenant_id,
          siteId: site.site_id,
        });
        this.$router.push(targetPath);
      } finally {
        this.switchingSiteId = 0;
      }
    },

    async openGpsModal(site) {
      if (!site || !this.canEditStationGps) return;
      if (this.isMapFullscreen) {
        await this.exitMapFullscreen();
      }
      const hasValid = site.has_coordinates && isValidStationCoordinate(site.latitude, site.longitude);
      this.gpsForm = {
        site_id: site.site_id,
        site_name: site.site_name,
        tenant_name: site.tenant_name,
        latitude: hasValid ? Number(site.latitude) : null,
        longitude: hasValid ? Number(site.longitude) : null,
        address: site.address || '',
      };
      this.showGpsModal = true;
    },

    async onPickStationGps(coords) {
      if (!coords) return;
      const rawLat = coords.lat !== undefined ? coords.lat : coords.latitude;
      const rawLng = coords.lng !== undefined ? coords.lng : coords.longitude;
      if (rawLat !== null && rawLat !== undefined && Number.isFinite(Number(rawLat))) {
        this.gpsForm.latitude = Number(Number(rawLat).toFixed(6));
      }
      if (rawLng !== null && rawLng !== undefined && Number.isFinite(Number(rawLng))) {
        this.gpsForm.longitude = Number(Number(rawLng).toFixed(6));
      }
      const addr = coords.displayName || coords.display_name || coords.address || '';
      if (addr && !this.gpsForm.address) {
        this.gpsForm.address = addr;
      } else if (coords.source === 'MAP' && isValidStationCoordinate(this.gpsForm.latitude, this.gpsForm.longitude)) {
        try {
          const res = await axios.get('/api/system/location/reverse', {
            params: { lat: this.gpsForm.latitude, lng: this.gpsForm.longitude },
          });
          const geo = res && res.data && res.data.data;
          const revAddr = (geo && (geo.display_name || geo.place_name)) || '';
          if (revAddr) {
            this.gpsForm.address = revAddr;
          }
        } catch (_) {
          // ignore reverse geocode network errors
        }
      }
    },

    onSelectStationLocation(loc) {
      if (!loc) return;
      const rawLat = loc.lat !== undefined ? loc.lat : loc.latitude;
      const rawLng = loc.lng !== undefined ? loc.lng : loc.longitude;
      if (rawLat !== null && rawLat !== undefined && Number.isFinite(Number(rawLat))) {
        this.gpsForm.latitude = Number(Number(rawLat).toFixed(6));
      }
      if (rawLng !== null && rawLng !== undefined && Number.isFinite(Number(rawLng))) {
        this.gpsForm.longitude = Number(Number(rawLng).toFixed(6));
      }
      const resolvedAddress = loc.displayName || loc.display_name || loc.name || '';
      if (resolvedAddress) {
        this.gpsForm.address = resolvedAddress;
      }
    },

    async saveStationGps() {
      if (!isValidStationCoordinate(this.gpsForm.latitude, this.gpsForm.longitude)) {
        this.errorMessage = 'Koordinat Latitude / Longitude tidak valid atau masih bernilai (0, 0).';
        return;
      }
      this.savingGps = true;
      try {
        await axios.put(`/api/sites/${this.gpsForm.site_id}`, {
          latitude: Number(this.gpsForm.latitude),
          longitude: Number(this.gpsForm.longitude),
          location_source: 'MANUAL',
          address: this.gpsForm.address || '',
        });
        const targetId = this.gpsForm.site_id;
        this.showGpsModal = false;
        await this.fetchFleetOverview(true);
        const updated = this.sites.find(s => Number(s.site_id) === Number(targetId));
        if (updated) {
          this.focusSite(updated);
        }
      } catch (err) {
        this.errorMessage =
          (err.response && err.response.data && err.response.data.error) ||
          err.message ||
          'Gagal menyimpan koordinat GPS stasiun.';
      } finally {
        this.savingGps = false;
      }
    },

    // Formatting & Badge Helpers
    formatRelativeAge(site) {
      if (!site) return '-';
      if (site.delivery_status === 'INACTIVE' && site.enabled_device_count === 0) {
        return '0 Device';
      }
      if (!site.last_transmission_at) {
        return 'Belum kirim';
      }
      const diffSec = Math.max(
        0,
        Math.floor((this.nowTick - new Date(site.last_transmission_at).getTime()) / 1000)
      );
      if (diffSec < 60) return `${diffSec}d lalu`;
      if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m lalu`;
      if (diffSec < 86400) {
        const h = Math.floor(diffSec / 3600);
        const m = Math.floor((diffSec % 3600) / 60);
        return m > 0 ? `${h}j ${m}m lalu` : `${h}j lalu`;
      }
      return `${Math.floor(diffSec / 86400)}hri lalu`;
    },

    formatTimestampAgo(ts) {
      if (!ts) return '-';
      const diffSec = Math.max(0, Math.floor((this.nowTick - new Date(ts).getTime()) / 1000));
      if (diffSec < 60) return `${diffSec}d lalu`;
      if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m lalu`;
      if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}j lalu`;
      return `${Math.floor(diffSec / 86400)}hri lalu`;
    },

    formatClockTime(ts) {
      if (!ts) return '--:--';
      try {
        return new Date(ts).toLocaleTimeString('id-ID', {
          hour: '2-digit',
          minute: '2-digit',
          second: '2-digit',
        });
      } catch (_) {
        return '--:--';
      }
    },

    deliveryLabel(status) {
      const map = {
        ONLINE: 'ONLINE',
        DELAYED: 'DELAYED',
        OFFLINE: 'OFFLINE',
        INACTIVE: 'NONAKTIF',
      };
      return map[status] || status || 'UNKNOWN';
    },

    severityBarClass(site) {
      if (!site) return 'bg-slate-400';
      if (site.severity_state === 'CRITICAL') return 'bg-rose-500';
      if (site.severity_state === 'WARNING') return 'bg-amber-500';
      if (site.severity_state === 'INACTIVE') return 'bg-slate-400';
      return 'bg-emerald-500';
    },

    domainBadgeClass(domain) {
      const d = String(domain || '').toUpperCase();
      if (d === 'WQMS' || d === 'ONLIMO') {
        return 'bg-teal-500/10 text-teal-700 dark:text-teal-300 border-teal-500/30';
      }
      if (d === 'SPARING') {
        return 'bg-violet-500/10 text-violet-700 dark:text-violet-300 border-violet-500/30';
      }
      return 'bg-sky-500/10 text-sky-700 dark:text-sky-300 border-sky-500/30';
    },

    deliveryBadgeClass(status) {
      if (status === 'ONLINE') {
        return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30';
      }
      if (status === 'DELAYED') {
        return 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30';
      }
      if (status === 'OFFLINE') {
        return 'bg-rose-500/10 text-rose-700 dark:text-rose-300 border-rose-500/30';
      }
      return 'bg-slate-500/10 text-slate-600 dark:text-slate-400 border-slate-500/25';
    },

    deliveryTextClass(status) {
      if (status === 'ONLINE') return 'text-emerald-600 dark:text-emerald-400 font-semibold';
      if (status === 'DELAYED') return 'text-amber-600 dark:text-amber-400 font-semibold';
      if (status === 'OFFLINE') return 'text-rose-600 dark:text-rose-400 font-bold';
      return 'text-slate-400';
    },

    getSiteEnvDockBadge(site) {
      const meta = this.resolveSiteEnvPinMeta(site);
      if (!meta || !meta.available) return null;
      return {
        label: meta.label,
        bg: meta.bg,
        tooltip: meta.tooltipLine,
      };
    },

    formatWaterClassLabel(wc) {
      const map = { CLASS_1: 'Kelas I', CLASS_2: 'Kelas II', CLASS_3: 'Kelas III', CLASS_4: 'Kelas IV' };
      return map[String(wc || '').toUpperCase()] || 'Kelas II';
    },

    ispuBannerClass(cat) {
      const c = String(cat || '').toUpperCase();
      if (c === 'BERBAHAYA' || c === 'SANGAT_TIDAK_SEHAT') return 'bg-rose-500/10 border-rose-500/35 text-rose-800 dark:text-rose-200';
      if (c === 'TIDAK_SEHAT') return 'bg-amber-500/10 border-amber-500/35 text-amber-800 dark:text-amber-200';
      if (c === 'SEDANG') return 'bg-sky-500/10 border-sky-500/35 text-sky-800 dark:text-sky-200';
      return 'bg-emerald-500/10 border-emerald-500/35 text-emerald-800 dark:text-emerald-200';
    },

    ispuSolidBadgeClass(cat) {
      const c = String(cat || '').toUpperCase();
      if (c === 'BERBAHAYA') return 'bg-slate-950 border border-rose-500';
      if (c === 'SANGAT_TIDAK_SEHAT') return 'bg-rose-600';
      if (c === 'TIDAK_SEHAT') return 'bg-amber-500';
      if (c === 'SEDANG') return 'bg-sky-600';
      return 'bg-emerald-600';
    },

    wqmsIPBannerClass(cat) {
      const c = String(cat || '').toUpperCase();
      if (c === 'CEMAR_BERAT') return 'bg-rose-500/10 border-rose-500/35 text-rose-800 dark:text-rose-200';
      if (c === 'CEMAR_SEDANG') return 'bg-orange-500/10 border-orange-500/35 text-orange-800 dark:text-orange-200';
      if (c === 'CEMAR_RINGAN') return 'bg-amber-500/10 border-amber-500/35 text-amber-800 dark:text-amber-200';
      return 'bg-teal-500/10 border-teal-500/35 text-teal-800 dark:text-teal-200';
    },

    wqmsIPSolidBadgeClass(cat) {
      const c = String(cat || '').toUpperCase();
      if (c === 'CEMAR_BERAT') return 'bg-rose-600';
      if (c === 'CEMAR_SEDANG') return 'bg-orange-600';
      if (c === 'CEMAR_RINGAN') return 'bg-amber-500';
      return 'bg-teal-600';
    },

    wqmsIKABannerClass(cat) {
      const c = String(cat || '').toUpperCase();
      if (c === 'SANGAT_KURANG') return 'bg-rose-500/10 border-rose-500/35 text-rose-800 dark:text-rose-200';
      if (c === 'KURANG') return 'bg-orange-500/10 border-orange-500/35 text-orange-800 dark:text-orange-200';
      if (c === 'SEDANG') return 'bg-amber-500/10 border-amber-500/35 text-amber-800 dark:text-amber-200';
      if (c === 'BAIK') return 'bg-sky-500/10 border-sky-500/35 text-sky-800 dark:text-sky-200';
      return 'bg-emerald-500/10 border-emerald-500/35 text-emerald-800 dark:text-emerald-200';
    },

    wqmsIKASolidBadgeClass(cat) {
      const c = String(cat || '').toUpperCase();
      if (c === 'SANGAT_KURANG') return 'bg-rose-600';
      if (c === 'KURANG') return 'bg-orange-600';
      if (c === 'SEDANG') return 'bg-amber-500';
      if (c === 'BAIK') return 'bg-sky-600';
      return 'bg-emerald-600';
    },

    escapeHtml(raw) {
      return String(raw || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;');
    },
  },
};
</script>

<style>
.fleet-leaflet-div-icon {
  background: transparent !important;
  border: none !important;
}
.fleet-pin-root {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}
.fleet-pin-core {
  position: relative;
  z-index: 2;
  border-radius: 9999px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.15s ease;
}
.fleet-pin-core:hover {
  transform: scale(1.12);
}
.fleet-pin-letter {
  color: #ffffff;
  font-size: 10px;
  font-weight: 800;
  font-family: Inter, system-ui, sans-serif;
  line-height: 1;
}
.fleet-pin-score {
  color: #ffffff;
  font-weight: 800;
  font-family: JetBrains Mono, Inter, system-ui, sans-serif;
  letter-spacing: -0.03em;
  line-height: 1;
}
.fleet-pin-pulse {
  position: absolute;
  inset: -6px;
  border-radius: 9999px;
  z-index: 1;
  animation: fleetPinPulseAnim 1.8s cubic-bezier(0, 0, 0.2, 1) infinite;
}
.fleet-pin-badge {
  position: absolute;
  top: -6px;
  right: -7px;
  z-index: 3;
  min-width: 15px;
  height: 15px;
  padding: 0 3.5px;
  border-radius: 9999px;
  background: #be123c;
  color: #ffffff;
  border: 1.5px solid #ffffff;
  font-size: 9px;
  font-weight: 800;
  font-family: JetBrains Mono, monospace;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: 1;
}
@keyframes fleetPinPulseAnim {
  0% {
    transform: scale(0.8);
    opacity: 0.9;
  }
  80%,
  100% {
    transform: scale(1.65);
    opacity: 0;
  }
}
.fleet-no-scrollbar::-webkit-scrollbar {
  display: none;
}
.fleet-no-scrollbar {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
.fleet-cluster-root {
  position: relative;
  border-radius: 9999px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: transform 0.16s ease-out;
}
.fleet-cluster-root:hover {
  transform: scale(1.1);
}
.fleet-cluster-core {
  position: relative;
  z-index: 2;
  border-radius: 9999px;
  border: 2.5px solid #ffffff;
  box-shadow: 0 4px 12px rgba(15, 23, 42, 0.42);
  display: flex;
  align-items: center;
  justify-content: center;
}
.fleet-cluster-count {
  color: #ffffff;
  font-size: 11px;
  font-weight: 800;
  font-family: JetBrains Mono, Inter, system-ui, sans-serif;
  line-height: 1;
}
/* AQMS Wind Vector Arrow & Rain Ripple Map Overlays (Visible at Overview & Detail Zoom) */
.fleet-wind-orbit {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  z-index: 4;
}
.fleet-wind-arrow-rot {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  transform-origin: center center;
}
.fleet-wind-arrow-svg {
  margin-top: -11px;
  filter: drop-shadow(0 1.5px 2.5px rgba(15, 23, 42, 0.55));
}
.fleet-rain-ripple {
  position: absolute;
  inset: -10px;
  border-radius: 9999px;
  border: 2.5px solid rgba(37, 99, 235, 0.82);
  background: radial-gradient(circle, rgba(59, 130, 246, 0.24) 0%, rgba(37, 99, 235, 0.04) 70%);
  pointer-events: none;
  z-index: 1;
  animation: fleetRainRippleAnim 2.2s cubic-bezier(0.18, 0.72, 0.25, 1) infinite;
}
.fleet-rain-ripple-delay {
  animation-delay: 1.1s;
}
@keyframes fleetRainRippleAnim {
  0% {
    transform: scale(0.72);
    opacity: 0.95;
  }
  85%,
  100% {
    transform: scale(1.85);
    opacity: 0;
  }
}
.fleet-weather-pill-strip {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  display: inline-flex;
  align-items: center;
  gap: 3px;
  white-space: nowrap;
  pointer-events: none;
  z-index: 5;
}
.fleet-weather-pill {
  padding: 1.5px 5px;
  border-radius: 9999px;
  color: #ffffff;
  font-size: 9.5px;
  font-weight: 800;
  font-family: JetBrains Mono, Inter, system-ui, sans-serif;
  line-height: 1.15;
  letter-spacing: -0.01em;
  border: 1.5px solid rgba(255, 255, 255, 0.95);
  box-shadow: 0 2px 6px rgba(15, 23, 42, 0.38);
}
.fleet-weather-pill-rain {
  background: rgba(37, 99, 235, 0.95);
}
.fleet-map-pane-fullscreen {
  position: fixed !important;
  inset: 0 !important;
  z-index: 9999 !important;
  width: 100vw !important;
  height: 100vh !important;
  border-radius: 0 !important;
  border: 0 !important;
}
/* Eliminate Leaflet 1.9.4 plus-lighter white grid lines & apply color filters once on tilePane */
#fleet-leaflet-map .leaflet-tile-container img,
#fleet-leaflet-map img.leaflet-tile {
  max-width: none !important;
  max-height: none !important;
  width: 256.5px !important;
  height: 256.5px !important;
  mix-blend-mode: normal !important;
  filter: none !important;
  margin: 0 !important;
  padding: 0 !important;
  border: 0 !important;
  outline: 1px solid transparent;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
}
#fleet-leaflet-map .leaflet-tile-container {
  will-change: transform;
}
.fleet-tiles-osm-vibrant {
  filter: saturate(1.16) contrast(1.04) brightness(0.99);
}
.fleet-tiles-topo-vibrant {
  filter: saturate(1.52) contrast(1.09) brightness(0.98);
}
.fleet-tiles-sat-crisp {
  filter: saturate(1.14) contrast(1.06);
}
.fleet-tiles-night-comfort {
  filter: invert(90%) hue-rotate(192deg) saturate(142%) brightness(98%) contrast(96%);
}
.fleet-tiles-navy-night {
  filter: invert(92%) hue-rotate(194deg) saturate(152%) brightness(94%) contrast(98%);
}
#fleet-leaflet-map.leaflet-container {
  background: #aad3df;
}
.dark #fleet-leaflet-map.leaflet-container {
  background: #183657;
}
#fleet-leaflet-map .leaflet-control-scale-line {
  background: rgba(255, 255, 255, 0.86);
  border-color: #475569;
  color: #1e293b;
  font-size: 9px;
  font-weight: 700;
  font-family: JetBrains Mono, monospace;
  border-radius: 4px;
  padding: 1px 5px;
}
.dark #fleet-leaflet-map .leaflet-control-scale-line {
  background: rgba(14, 22, 38, 0.88);
  border-color: #64748b;
  color: #e2e8f0;
}
.dark #fleet-leaflet-map .leaflet-control-zoom {
  border: 1px solid rgba(51, 65, 85, 0.9) !important;
  box-shadow: 0 4px 12px rgba(2, 6, 23, 0.45) !important;
}
.dark #fleet-leaflet-map .leaflet-control-zoom a {
  background-color: rgba(14, 22, 38, 0.92) !important;
  color: #e2e8f0 !important;
  border-bottom-color: rgba(51, 65, 85, 0.85) !important;
}
.dark #fleet-leaflet-map .leaflet-control-zoom a:hover {
  background-color: rgba(30, 41, 59, 0.96) !important;
  color: #38bdf8 !important;
}
</style>

