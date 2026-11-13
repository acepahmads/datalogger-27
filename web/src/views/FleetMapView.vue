<template>
  <div
    id="superadmin-fleet-map-view"
    class="flex flex-col flex-1 min-h-0 h-full w-full gap-2 select-none font-sans text-slate-800 dark:text-slate-100"
  >
    <!-- ===================================================================== -->
    <!-- 1. DESKTOP 1-ROW UNIFIED COMMAND & KPI RIBBON (>= 768px)              -->
    <!-- ===================================================================== -->
    <div
      id="fleet-desktop-command-ribbon"
      class="hidden md:flex items-center justify-between gap-2 px-2.5 py-1.5 rounded-xl bg-white dark:bg-[#0E1626] border border-slate-200/90 dark:border-slate-800/90 shadow-xs flex-shrink-0 flex-wrap xl:flex-nowrap"
    >
      <!-- Left: Brand Title + Interactive KPI Filter Pills -->
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

      <!-- Right: Compact Search, Dropdown Filters & Map Utilities -->
      <div class="flex items-center gap-1.5 flex-wrap xl:flex-nowrap ml-auto">
        <!-- Search Input -->
        <div class="relative w-44 lg:w-52">
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

        <!-- Tenant Filter -->
        <select
          id="select-fleet-tenant"
          v-model="tenantFilter"
          class="h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 max-w-[150px] focus:outline-none focus:border-sky-500 cursor-pointer"
          title="Filter Tenant"
        >
          <option value="ALL">Semua Tenant ({{ tenants.length }})</option>
          <option v-for="t in tenants" :key="t.id" :value="String(t.id)">
            {{ t.name }} ({{ t.site_count }})
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

        <!-- Map Style Cycle Button -->
        <button
          type="button"
          id="btn-fleet-map-style"
          @click="cycleMapStyle"
          class="h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
          :title="'Tema Peta: ' + mapStyleLabel"
        >
          <span>🗺️</span>
          <span>{{ mapStyleLabel }}</span>
        </button>

        <!-- Fit Bounds Button -->
        <button
          type="button"
          id="btn-fleet-fit-bounds"
          @click="fitMapToVisibleSites"
          class="h-7 px-2 rounded-lg text-2xs font-semibold bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 inline-flex items-center gap-1 transition cursor-pointer whitespace-nowrap"
          title="Fit semua pin stasiun di layar"
        >
          <span>⛶</span>
          <span>Fit</span>
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
          <span class="hidden 2xl:inline font-mono text-3xs">{{ lastUpdatedShort }}</span>
        </button>

        <!-- Collapse / Expand Priority Panel -->
        <button
          type="button"
          @click="isRightDockCollapsed = !isRightDockCollapsed"
          class="h-7 px-2 rounded-lg text-2xs font-semibold border transition cursor-pointer whitespace-nowrap"
          :class="!isRightDockCollapsed
            ? 'bg-sky-600 text-white btn-text border-sky-600 shadow-2xs'
            : 'bg-slate-50 dark:bg-[#0B0F19] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-700/80'"
          title="Tampilkan / sembunyikan panel Prioritas Stasiun"
        >
          {{ isRightDockCollapsed ? '◂ Panel' : 'Panel ▸' }}
        </button>
      </div>
    </div>

    <!-- ===================================================================== -->
    <!-- 2. MOBILE / HP COMPACT TOP CONTROL BAR (< 768px)                      -->
    <!-- ===================================================================== -->
    <div
      id="fleet-mobile-header"
      class="flex md:hidden flex-col gap-1.5 p-2 rounded-xl bg-white dark:bg-[#0E1626] border border-slate-200/90 dark:border-slate-800/90 shadow-xs flex-shrink-0"
    >
      <!-- Row 1: Search + Filter Drawer Button + Segmented Switcher (Peta | Daftar) -->
      <div class="flex items-center gap-1.5">
        <div class="relative flex-1 min-w-0">
          <svg class="w-3.5 h-3.5 text-slate-400 absolute left-2.5 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <circle cx="11" cy="11" r="8" />
            <path stroke-linecap="round" d="m21 21-4.3-4.3" />
          </svg>
          <input
            v-model.trim="searchQuery"
            type="text"
            placeholder="Cari site / tenant..."
            class="w-full h-8 pl-7 pr-6 rounded-lg text-xs bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-800 dark:text-slate-100 placeholder:text-slate-400 focus:outline-none focus:border-sky-500"
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
          class="relative h-8 px-2.5 rounded-lg border text-xs font-bold inline-flex items-center justify-center gap-1 cursor-pointer flex-shrink-0"
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
          class="h-8 w-8 rounded-lg bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 flex items-center justify-center text-xs font-bold cursor-pointer flex-shrink-0"
          title="Segarkan Data"
        >
          <span :class="{ 'animate-spin': loading }">↻</span>
        </button>

        <!-- Segmented View Switcher: Peta | Daftar -->
        <div
          id="fleet-mobile-view-switcher"
          class="inline-flex items-center p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 flex-shrink-0"
        >
          <button
            type="button"
            id="btn-mobile-mode-map"
            @click="switchMobileView('map')"
            class="px-2.5 py-1 rounded-md text-2xs font-bold transition cursor-pointer"
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
            class="px-2.5 py-1 rounded-md text-2xs font-bold transition cursor-pointer"
            :class="mobileViewMode === 'list'
              ? 'bg-sky-600 text-white btn-text shadow-xs'
              : 'text-slate-600 dark:text-slate-400'"
          >
            📋 Daftar ({{ filteredSites.length }})
          </button>
        </div>
      </div>

      <!-- Row 2: Horizontally Scrollable KPI Filter Chips -->
      <div class="flex items-center gap-1.5 overflow-x-auto pb-0.5 fleet-no-scrollbar">
        <button
          v-for="pill in kpiPills"
          :key="'mob-' + pill.key"
          type="button"
          @click="setQuickFilter(pill.key)"
          class="inline-flex items-center gap-1 px-2 py-1 rounded-lg text-2xs font-bold border whitespace-nowrap flex-shrink-0 cursor-pointer"
          :class="quickFilter === pill.key ? pill.activeClass : pill.idleClass"
        >
          <span v-if="pill.dotClass" class="w-1.5 h-1.5 rounded-full" :class="pill.dotClass"></span>
          <span>{{ pill.label }}</span>
          <span class="font-mono text-3xs opacity-90">({{ pill.count }})</span>
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
    <!-- 3. MAIN SPLIT WORKSPACE (MAP CANVAS + DOCKED PRIORITY LIST)           -->
    <!-- ===================================================================== -->
    <div class="flex-1 min-h-0 flex gap-2.5 relative overflow-hidden">
      <!-- 3A. INTERACTIVE LEAFLET MAP PANE -->
      <div
        v-show="isDesktopViewport || mobileViewMode === 'map'"
        id="fleet-map-pane"
        class="flex-1 min-w-0 h-full relative rounded-2xl overflow-hidden border border-slate-200/90 dark:border-slate-800/90 bg-slate-100 dark:bg-[#0B0F19] shadow-xs"
      >
        <!-- Leaflet Container -->
        <div ref="fleetMapCanvas" id="fleet-leaflet-map" class="w-full h-full z-10"></div>

        <!-- Mobile Map Floating Utility Buttons (Fit & Style) -->
        <div class="flex md:hidden flex-col gap-1.5 absolute top-2.5 right-2.5 z-[450]">
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

        <!-- Compact Legend Pill (Desktop Bottom-Left) -->
        <div
          v-if="isTilesOnline"
          class="hidden sm:flex items-center gap-2.5 absolute bottom-2.5 left-2.5 z-[450] px-2.5 py-1 rounded-lg bg-white/90 dark:bg-[#0E1626]/90 backdrop-blur-xs border border-slate-200/80 dark:border-slate-700/80 text-3xs font-semibold text-slate-600 dark:text-slate-300 shadow-sm pointer-events-none"
        >
          <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-emerald-500"></span>Online</span>
          <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-amber-500"></span>Delayed / Warning</span>
          <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-rose-500"></span>Offline / Critical</span>
          <span class="inline-flex items-center gap-1"><span class="w-2 h-2 rounded-full bg-slate-400"></span>Nonaktif</span>
        </div>

        <!-- =============================================================== -->
        <!-- DESKTOP FLOATING SITE DETAIL CARD (>= 768px)                    -->
        <!-- =============================================================== -->
        <transition name="fade">
          <div
            v-if="selectedSite && isDesktopViewport"
            id="fleet-desktop-site-card"
            class="hidden md:flex flex-col absolute top-3 left-3 z-[500] w-[325px] lg:w-[350px] max-h-[calc(100%-24px)] rounded-2xl bg-white/95 dark:bg-[#0E1626]/95 backdrop-blur-md border border-slate-200 dark:border-slate-700/90 shadow-2xl overflow-hidden"
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
              <div class="grid grid-cols-3 gap-1.5">
                <button
                  type="button"
                  @click="openSiteModule(selectedSite, '/monitoring/devices')"
                  class="py-1 px-2 rounded-lg bg-white hover:bg-slate-100 dark:bg-slate-800/90 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 text-3xs font-bold transition cursor-pointer"
                >
                  Perangkat ({{ selectedSite.device_count }})
                </button>
                <button
                  type="button"
                  @click="openSiteModule(selectedSite, '/alarm/active')"
                  class="py-1 px-2 rounded-lg bg-white hover:bg-slate-100 dark:bg-slate-800/90 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200 text-3xs font-bold transition cursor-pointer"
                >
                  Detail Alarm
                </button>
                <button
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

      <!-- ================================================================= -->
      <!-- 3B. DESKTOP RIGHT PRIORITY STATION DOCK (>= 768px)                -->
      <!-- ================================================================= -->
      <aside
        v-if="!isRightDockCollapsed"
        id="fleet-priority-dock"
        class="hidden md:flex flex-col w-[305px] xl:w-[335px] flex-shrink-0 h-full rounded-2xl bg-white dark:bg-[#0E1626] border border-slate-200/90 dark:border-slate-800/90 shadow-xs overflow-hidden"
      >
        <!-- Dock Header + Sub-Tabs -->
        <div class="px-3 pt-2.5 pb-2 border-b border-slate-100 dark:border-slate-800/80 space-y-2 flex-shrink-0">
          <div class="flex items-center justify-between">
            <span class="text-xs font-extrabold text-slate-900 dark:text-white tracking-tight">Prioritas Stasiun</span>
            <span class="text-3xs font-mono text-slate-400">{{ dockDisplayedSites.length }} ditampilkan</span>
          </div>
          <div class="grid grid-cols-3 gap-1 p-0.5 rounded-lg bg-slate-100 dark:bg-[#0B0F19] border border-slate-200/70 dark:border-slate-800">
            <button
              type="button"
              @click="dockTab = 'ALL'"
              class="py-1 rounded-md text-3xs font-bold transition cursor-pointer"
              :class="dockTab === 'ALL' ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-white shadow-2xs' : 'text-slate-500 dark:text-slate-400'"
            >
              Semua ({{ filteredSites.length }})
            </button>
            <button
              type="button"
              @click="dockTab = 'ISSUES'"
              class="py-1 rounded-md text-3xs font-bold transition cursor-pointer"
              :class="dockTab === 'ISSUES' ? 'bg-rose-600 text-white btn-text shadow-2xs' : 'text-slate-500 dark:text-slate-400'"
            >
              Masalah ({{ issueFilteredSites.length }})
            </button>
            <button
              type="button"
              @click="dockTab = 'NOGPS'"
              class="py-1 rounded-md text-3xs font-bold transition cursor-pointer"
              :class="dockTab === 'NOGPS' ? 'bg-amber-600 text-white btn-text shadow-2xs' : 'text-slate-500 dark:text-slate-400'"
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
              <!-- Line 1: Site Name + Domain Badge -->
              <div class="flex items-center gap-1.5 min-w-0">
                <span class="text-xs font-bold text-slate-900 dark:text-white truncate">{{ site.site_name }}</span>
                <span class="px-1 py-0.2 rounded text-4xs font-mono font-bold border flex-shrink-0" :class="domainBadgeClass(site.monitoring_domain)">
                  {{ site.monitoring_domain }}
                </span>
              </div>
              <!-- Line 2: Tenant + Last Transmission -->
              <div class="flex items-center gap-1.5 text-3xs text-slate-500 dark:text-slate-400 mt-0.5 min-w-0">
                <span class="truncate font-medium">{{ site.tenant_name }}</span>
                <span>·</span>
                <span class="font-mono whitespace-nowrap" :class="deliveryTextClass(site.delivery_status)">
                  {{ formatRelativeAge(site) }}
                </span>
              </div>
            </div>

            <!-- Right Indicators & Quick Action -->
            <div class="flex items-center gap-1 flex-shrink-0">
              <span
                v-if="site.total_active_alarms > 0"
                class="px-1.5 py-0.5 rounded-full text-4xs font-mono font-extrabold text-white badge-text"
                :class="site.critical_alarm_count > 0 ? 'bg-rose-600' : 'bg-amber-500'"
                :title="site.total_active_alarms + ' alarm aktif'"
              >
                🚨 {{ site.total_active_alarms }}
              </span>
              <button
                v-if="!site.has_coordinates"
                type="button"
                @click.stop="openGpsModal(site)"
                class="px-1.5 py-0.5 rounded-md bg-amber-500/15 hover:bg-amber-500/25 text-amber-700 dark:text-amber-300 border border-amber-500/30 text-4xs font-bold cursor-pointer"
                title="Set Koordinat GPS Stasiun"
              >
                + GPS
              </button>
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
              v-else
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
    <!-- 4. MOBILE FILTER BOTTOM SHEET (< 768px)                               -->
    <!-- ===================================================================== -->
    <div
      v-if="showMobileFilterSheet"
      @click.self="showMobileFilterSheet = false"
      class="fixed inset-0 z-[80] flex items-end justify-center bg-black/55 backdrop-blur-xs md:hidden"
    >
      <div class="w-full rounded-t-3xl bg-white dark:bg-[#0E1626] border-t border-slate-200 dark:border-slate-800 p-4 space-y-3.5 shadow-2xl max-h-[82vh] overflow-y-auto">
        <div class="flex items-center justify-between">
          <h3 class="text-sm font-extrabold text-slate-900 dark:text-white">Filter Peta & Stasiun</h3>
          <button type="button" @click="showMobileFilterSheet = false" class="text-slate-400 text-sm font-bold cursor-pointer">✕</button>
        </div>

        <div class="space-y-2.5 text-xs">
          <div>
            <label class="block text-2xs font-bold text-slate-500 dark:text-slate-400 mb-1">Organisasi Tenant</label>
            <select
              v-model="tenantFilter"
              class="w-full h-9 px-2.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-100"
            >
              <option value="ALL">Semua Tenant ({{ tenants.length }})</option>
              <option v-for="t in tenants" :key="'m-t-' + t.id" :value="String(t.id)">
                {{ t.name }} ({{ t.site_count }} site)
              </option>
            </select>
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
              @pick-coordinates="onPickStationGps"
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

const INDONESIA_DEFAULT_CENTER = [-2.5489, 118.0149];
const INDONESIA_DEFAULT_ZOOM = 5;

// 100% Free, Zero-API-Key, Zero-Watermark Map Tile Providers (Esri Public ArcGIS Online & OpenStreetMap)
const TILE_LAYERS = {
  clean: {
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}',
    fallbackUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
    options: { maxZoom: 19 },
  },
  osm: {
    url: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
    options: { maxZoom: 19, className: 'fleet-tiles-osm' },
  },
  dark: {
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/Canvas/World_Dark_Gray_Base/MapServer/tile/{z}/{y}/{x}',
    overlayUrl: 'https://server.arcgisonline.com/ArcGIS/rest/services/Canvas/World_Dark_Gray_Reference/MapServer/tile/{z}/{y}/{x}',
    fallbackUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
    fallbackClassName: 'fleet-tiles-dark-osm',
    options: { maxZoom: 16 },
  },
  satellite: {
    url: 'https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}',
    overlayUrl: 'https://server.arcgisonline.com/ArcGIS/rest/services/Reference/World_Boundaries_and_Places/MapServer/tile/{z}/{y}/{x}',
    fallbackUrl: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
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

      // Filters
      quickFilter: 'ALL', // ALL | ONLINE | DELAYED | OFFLINE | ALARM | NOGPS
      searchQuery: '',
      tenantFilter: 'ALL',
      domainFilter: 'ALL',
      alarmFilter: 'ALL',
      dockTab: 'ALL', // ALL | ISSUES | NOGPS

      // Responsive & Layout State
      viewportWidth: typeof window !== 'undefined' ? window.innerWidth : 1280,
      mobileViewMode: 'map', // 'map' | 'list'
      showMobileFilterSheet: false,
      isRightDockCollapsed: false,

      // Map & Selection State
      map: null,
      tileLayer: null,
      tileOverlayLayer: null,
      markersLayer: null,
      markerBySiteId: {},
      hasInitialFitDone: false,
      isTilesOnline: true,
      mapStyle: 'auto', // 'auto' | 'clean' | 'osm' | 'dark' | 'satellite'
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
    isDesktopViewport() {
      return this.viewportWidth >= 768;
    },
    effectiveAppDark() {
      const pref = this.$store && this.$store.state && this.$store.state.uiPreferences;
      if (pref && pref.effectiveTheme) return pref.effectiveTheme === 'dark';
      return Boolean(this.$store && this.$store.state && this.$store.state.theme === 'dark');
    },
    resolvedTileTheme() {
      if (this.mapStyle === 'auto') {
        return this.effectiveAppDark ? 'dark' : 'clean';
      }
      return this.mapStyle;
    },
    mapStyleLabel() {
      const map = { auto: 'Auto', clean: 'Clean', osm: 'OSM', dark: 'Gelap', satellite: 'Satelit' };
      return map[this.mapStyle] || 'Auto';
    },
    mapStyleOptions() {
      return [
        { value: 'auto', label: 'Auto' },
        { value: 'clean', label: 'Clean' },
        { value: 'osm', label: 'OSM' },
        { value: 'dark', label: 'Gelap' },
        { value: 'satellite', label: 'Satelit' },
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
    hasActiveFilters() {
      return (
        this.quickFilter !== 'ALL' ||
        Boolean(this.searchQuery) ||
        this.tenantFilter !== 'ALL' ||
        this.domainFilter !== 'ALL' ||
        this.alarmFilter !== 'ALL'
      );
    },
    activeSecondaryFilterCount() {
      let n = 0;
      if (this.tenantFilter !== 'ALL') n += 1;
      if (this.domainFilter !== 'ALL') n += 1;
      if (this.alarmFilter !== 'ALL') n += 1;
      return n;
    },
    filteredSites() {
      const q = (this.searchQuery || '').toLowerCase();
      return (this.sites || []).filter((s) => {
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
    issueFilteredSites() {
      return this.filteredSites.filter(
        s => s.severity_state === 'CRITICAL' || s.severity_state === 'WARNING'
      );
    },
    noGpsFilteredSites() {
      return this.filteredSites.filter(s => !s.has_coordinates);
    },
    dockDisplayedSites() {
      if (this.dockTab === 'ISSUES') return this.issueFilteredSites;
      if (this.dockTab === 'NOGPS') return this.noGpsFilteredSites;
      return this.filteredSites;
    },
    selectedSite() {
      if (!this.selectedSiteId) return null;
      return (this.sites || []).find(s => Number(s.site_id) === Number(this.selectedSiteId)) || null;
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
    },
    mappedFilteredSites() {
      this.syncMapMarkers();
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
    if (!this.isSuperAdmin) {
      this.$router.replace({ path: '/access-denied', query: { from: this.$route.fullPath } });
      return;
    }
    if (typeof window !== 'undefined') {
      window.addEventListener('resize', this.onResize, { passive: true });
      window.addEventListener('alarm-event', this.scheduleRealtimeRefresh);
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
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', this.onResize);
      window.removeEventListener('alarm-event', this.scheduleRealtimeRefresh);
    }
    if (this.map) {
      this.map.remove();
      this.map = null;
    }
  },
  methods: {
    onResize() {
      if (typeof window === 'undefined') return;
      const prevDesktop = this.isDesktopViewport;
      this.viewportWidth = window.innerWidth;
      if (prevDesktop !== this.isDesktopViewport || this.mobileViewMode === 'map') {
        this.$nextTick(() => {
          if (this.map) this.map.invalidateSize();
        });
      }
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
      this.tenants = Array.isArray(payload.tenants) ? payload.tenants : [];
      this.sites = Array.isArray(payload.sites) ? payload.sites : [];
      this.errorMessage = '';

      this.$nextTick(() => {
        this.syncMapMarkers();
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
        const opt = { id: t.id, code: t.code, name: t.name, status: t.status || 'ACTIVE', site_count: 0, online_sites: 0, alarm_sites: 0 };
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
      });

      L.control.zoom({ position: 'bottomright' }).addTo(this.map);
      this.markersLayer = L.layerGroup().addTo(this.map);
      this.applyTileLayer();
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
      const cfg = TILE_LAYERS[this.resolvedTileTheme] || TILE_LAYERS.clean;
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
            const fbOpts = { maxZoom: 19 };
            if (cfg.fallbackClassName) fbOpts.className = cfg.fallbackClassName;
            this.tileLayer = attachLayer(cfg.fallbackUrl, fbOpts);
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

    buildPinIcon(site, isSelected) {
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

      const html = `
        <div class="fleet-pin-root" style="width:${size}px;height:${size}px;">
          ${pulseHtml}
          <div class="fleet-pin-core" style="width:${size}px;height:${size}px;background:${coreBg};border:2.5px solid ${borderCol};box-shadow:${shadowCss};">
            <span class="fleet-pin-letter">${domShort}</span>
          </div>
          ${alarmBadgeHtml}
        </div>
      `;

      return L.divIcon({
        className: 'fleet-leaflet-div-icon',
        html,
        iconSize: [size, size],
        iconAnchor: [size / 2, size / 2],
      });
    },

    syncMapMarkers() {
      if (!this.map || !this.markersLayer) return;
      this.markersLayer.clearLayers();
      this.markerBySiteId = {};

      this.mappedFilteredSites.forEach((site) => {
        const lat = Number(site.latitude);
        const lng = Number(site.longitude);
        if (!isValidStationCoordinate(lat, lng)) return;

        const isSelected = Number(this.selectedSiteId) === Number(site.site_id);
        const marker = L.marker([lat, lng], {
          icon: this.buildPinIcon(site, isSelected),
          zIndexOffset: isSelected ? 900 : site.severity_state === 'CRITICAL' ? 500 : site.severity_state === 'WARNING' ? 300 : 100,
        });

        const safeName = this.escapeHtml(site.site_name);
        const safeTenant = this.escapeHtml(site.tenant_name);
        const safeAge = this.escapeHtml(this.formatRelativeAge(site));
        marker.bindTooltip(
          `<div style="font-family:Inter,sans-serif;font-size:11px;line-height:1.35;"><strong>${safeName}</strong><br/><span style="opacity:0.82;">${safeTenant} · ${safeAge}</span></div>`,
          { direction: 'top', offset: [0, -12], opacity: 0.95 }
        );

        marker.on('click', () => {
          this.selectedSiteId = site.site_id;
          if (this.map) {
            const targetZoom = Math.max(this.map.getZoom(), 11);
            this.map.flyTo([lat, lng], targetZoom, { duration: 0.55 });
          }
        });

        marker.addTo(this.markersLayer);
        this.markerBySiteId[site.site_id] = marker;
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
      } else if (this.dockTab === 'NOGPS' && this.quickFilter !== 'NOGPS') {
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
      this.tenantFilter = 'ALL';
      this.domainFilter = 'ALL';
      this.alarmFilter = 'ALL';
      this.dockTab = 'ALL';
      this.$nextTick(() => this.fitMapToVisibleSites());
    },

    cycleMapStyle() {
      const order = ['auto', 'clean', 'osm', 'dark', 'satellite'];
      const idx = order.indexOf(this.mapStyle);
      this.mapStyle = order[(idx + 1) % order.length];
    },

    setMapStyle(val) {
      this.mapStyle = val;
    },

    async openSiteDashboard(site) {
      if (!site || this.switchingSiteId) return;
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

    openGpsModal(site) {
      if (!site) return;
      const hasValid = site.has_coordinates && isValidStationCoordinate(site.latitude, site.longitude);
      this.gpsForm = {
        site_id: site.site_id,
        site_name: site.site_name,
        tenant_name: site.tenant_name,
        latitude: hasValid ? Number(site.latitude) : -6.2,
        longitude: hasValid ? Number(site.longitude) : 106.816666,
        address: site.address || '',
      };
      this.showGpsModal = true;
    },

    onPickStationGps(coords) {
      if (!coords) return;
      if (Number.isFinite(Number(coords.latitude))) {
        this.gpsForm.latitude = Number(Number(coords.latitude).toFixed(6));
      }
      if (Number.isFinite(Number(coords.longitude))) {
        this.gpsForm.longitude = Number(Number(coords.longitude).toFixed(6));
      }
      if (coords.address && !this.gpsForm.address) {
        this.gpsForm.address = coords.address;
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
.fleet-tiles-osm {
  filter: saturate(0.88) contrast(1.02);
}
.fleet-tiles-dark-osm {
  filter: invert(100%) hue-rotate(180deg) brightness(93%) contrast(92%) saturate(80%);
}
</style>

