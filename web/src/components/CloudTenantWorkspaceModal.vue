<template>
  <div
    v-if="visible"
    :id="mode === 'route-gate' ? 'cloud-tenant-route-gate' : 'cloud-tenant-workspace-modal'"
    :class="mode === 'route-gate'
      ? 'w-full h-full min-h-[calc(100vh-76px)] flex items-center justify-center p-3 sm:p-6 bg-slate-50/90 dark:bg-slate-950'
      : 'fixed inset-0 z-[95] flex items-center justify-center p-3 sm:p-5 bg-slate-900/70 backdrop-blur-xs animate-fade-in'"
    @click.self="onBackdropClick"
  >
    <div
      class="w-full max-w-3xl bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-800 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[88vh]"
    >
      <!-- Header -->
      <div class="px-5 py-4 border-b border-slate-100 dark:border-slate-800 bg-slate-50/70 dark:bg-[#0B0F19]/70 flex items-start justify-between gap-3">
        <div class="flex items-start gap-3 min-w-0">
          <div class="w-10 h-10 rounded-xl bg-blue-500/10 dark:bg-blue-400/15 border border-blue-500/25 text-blue-600 dark:text-blue-400 flex items-center justify-center text-lg shrink-0">
            🏢
          </div>
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="text-sm sm:text-base font-extrabold text-slate-900 dark:text-white leading-snug">
                <template v-if="mode === 'route-gate'">
                  Pilih Tenant / Stasiun untuk Membuka {{ targetRouteLabel || 'Halaman Operasional' }}
                </template>
                <template v-else-if="mode === 'post-login'">
                  Pilih Workspace Tenant / Stasiun Pantau
                </template>
                <template v-else>
                  Pindah Cepat Tenant / Stasiun Pantau
                </template>
              </h2>
              <span class="px-2 py-0.5 rounded-full text-4xs font-mono font-bold bg-purple-500/10 text-purple-700 dark:text-purple-300 border border-purple-500/30">
                CLOUD MULTI-TENANT ({{ enrichedTenants.length }})
              </span>
            </div>
            <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5 leading-relaxed">
              <template v-if="mode === 'route-gate'">
                Halaman <strong class="text-slate-700 dark:text-slate-200">{{ targetRouteLabel }}</strong> memuat telemetri real-time per stasiun. Pilih salah satu Tenant/Stasiun di bawah agar tampilan cepat dan tidak overload.
              </template>
              <template v-else-if="mode === 'post-login'">
                Pilih organisasi Tenant atau Stasiun yang ingin langsung dipantau di seluruh menu, atau lanjutkan ke Peta Global.
              </template>
              <template v-else>
                Cari dan pilih organisasi Tenant atau titik Stasiun Pantau yang ingin diaktifkan.
              </template>
            </p>
          </div>
        </div>

        <button
          v-if="mode !== 'route-gate'"
          type="button"
          @click="handleDismiss"
          class="w-7 h-7 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-500 dark:text-slate-300 flex items-center justify-center text-xs shrink-0 cursor-pointer transition-colors"
          title="Tutup"
        >
          ✕
        </button>
      </div>

      <!-- Search & Quick Filter Bar -->
      <div class="px-5 py-3 border-b border-slate-100 dark:border-slate-800/80 bg-white dark:bg-[#0E1626] space-y-2.5">
        <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2">
          <!-- Search Input -->
          <div class="relative flex-1">
            <svg class="w-4 h-4 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <circle cx="11" cy="11" r="8" />
              <path stroke-linecap="round" stroke-linejoin="round" d="m21 21-4.3-4.3" />
            </svg>
            <input
              ref="searchInput"
              id="cloud-workspace-search-input"
              v-model="searchQuery"
              type="text"
              placeholder="Cari nama tenant, kode slug, stasiun, atau tahun..."
              class="w-full pl-9 pr-8 py-2 rounded-xl text-xs bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500/30"
            />
            <button
              v-if="searchQuery"
              type="button"
              @click="searchQuery = ''"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-xs cursor-pointer"
            >
              ✕
            </button>
          </div>

          <!-- Project Year Filter -->
          <select
            v-if="yearOptions.length > 0"
            v-model="yearFilter"
            class="px-3 py-2 rounded-xl text-xs font-semibold bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 focus:outline-none focus:border-blue-500 cursor-pointer"
          >
            <option value="ALL">📅 Semua Tahun</option>
            <option v-for="yr in yearOptions" :key="'ws-yr-' + yr" :value="String(yr)">
              📅 Tahun {{ yr }}
            </option>
          </select>

          <!-- Domain Filter -->
          <select
            v-model="domainFilter"
            class="px-3 py-2 rounded-xl text-xs font-semibold bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-700/80 text-slate-700 dark:text-slate-200 focus:outline-none focus:border-blue-500 cursor-pointer"
          >
            <option value="ALL">📡 Semua Domain</option>
            <option value="AQMS">AQMS (Udara)</option>
            <option value="SPARING">SPARING (Limbah)</option>
            <option value="ONLIMO">ONLIMO / WQMS (Air)</option>
          </select>
        </div>

        <div class="flex items-center justify-between text-3xs text-slate-500 dark:text-slate-400">
          <span>Menampilkan <strong class="text-slate-700 dark:text-slate-200">{{ filteredTenants.length }}</strong> dari {{ enrichedTenants.length }} organisasi tenant</span>
          <button
            v-if="searchQuery || yearFilter !== 'ALL' || domainFilter !== 'ALL'"
            type="button"
            @click="resetFilters"
            class="text-blue-600 dark:text-blue-400 hover:underline font-semibold cursor-pointer"
          >
            Reset Filter
          </button>
        </div>
      </div>

      <!-- Tenant & Station List -->
      <div class="p-4 overflow-y-auto space-y-2 flex-1 min-h-[220px] bg-slate-50/40 dark:bg-[#0B0F19]/40">
        <div v-if="loadingContext" class="py-12 text-center text-xs text-slate-500 dark:text-slate-400">
          Memuat daftar Tenant &amp; Stasiun...
        </div>

        <div v-else-if="filteredTenants.length === 0" class="py-12 text-center text-xs text-slate-500 dark:text-slate-400 space-y-2">
          <div class="text-2xl">🔍</div>
          <div class="font-semibold text-slate-700 dark:text-slate-300">Tidak ada Tenant / Stasiun yang cocok dengan pencarian.</div>
        </div>

        <template v-else>
          <div
            v-for="item in filteredTenants"
            :key="'ws-ten-' + item.tenant.id"
            class="rounded-xl border transition-all bg-white dark:bg-[#111827]"
            :class="Number(activeTenantId) === Number(item.tenant.id)
              ? 'border-blue-500 ring-1 ring-blue-500/30'
              : 'border-slate-200/90 dark:border-slate-800 hover:border-blue-400 dark:hover:border-blue-500/60'"
          >
            <!-- Single-Site or Primary Tenant Row -->
            <div
              class="p-3 flex items-center justify-between gap-3 cursor-pointer"
              @click="selectWorkspace(item.tenant.id, item.sites.length === 1 ? item.sites[0].id : 0)"
            >
              <div class="min-w-0 flex-1">
                <div class="flex flex-wrap items-center gap-1.5">
                  <span class="font-bold text-xs text-slate-900 dark:text-white truncate">
                    {{ item.tenant.name }}
                  </span>
                  <span
                    v-if="item.primaryDomain"
                    class="px-1.5 py-0.2 rounded text-4xs font-mono font-extrabold border"
                    :class="domainBadgeClass(item.primaryDomain)"
                  >
                    {{ item.primaryDomain }}
                  </span>
                  <span
                    v-if="item.projectYear > 0"
                    class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border border-indigo-500/30"
                  >
                    {{ item.projectYear }}
                  </span>
                  <span
                    v-if="Number(activeTenantId) === Number(item.tenant.id)"
                    class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border border-emerald-500/30"
                  >
                    AKTIF
                  </span>
                </div>
                <div class="text-3xs text-slate-500 dark:text-slate-400 font-mono truncate mt-0.5">
                  <span>{{ item.tenant.code }}</span>
                  <span v-if="item.sites.length === 1"> &bull; 📍 {{ item.sites[0].name }}</span>
                  <span v-else-if="item.sites.length > 1"> &bull; 📍 {{ item.sites.length }} Stasiun Pantau</span>
                </div>
              </div>

              <button
                type="button"
                :disabled="selectingKey === `${item.tenant.id}_${item.sites.length === 1 ? item.sites[0].id : 0}`"
                @click.stop="selectWorkspace(item.tenant.id, item.sites.length === 1 ? item.sites[0].id : 0)"
                class="px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white btn-text text-3xs font-bold shadow-xs shrink-0 flex items-center gap-1 cursor-pointer disabled:opacity-50"
              >
                <span>{{ selectingKey === `${item.tenant.id}_${item.sites.length === 1 ? item.sites[0].id : 0}` ? 'Memuat...' : 'Pilih →' }}</span>
              </button>
            </div>

            <!-- Multi-Site Sub-Selector (only when tenant has > 1 site) -->
            <div
              v-if="item.sites.length > 1"
              class="px-3 pb-2.5 pt-1.5 border-t border-slate-100 dark:border-slate-800/80 flex flex-wrap items-center gap-1.5"
            >
              <span class="text-4xs font-semibold uppercase tracking-wider text-slate-400 mr-1">Pilih Stasiun:</span>
              <button
                v-for="st in item.sites"
                :key="'ws-site-' + st.id"
                type="button"
                @click.stop="selectWorkspace(item.tenant.id, st.id)"
                class="px-2 py-1 rounded-lg text-3xs font-semibold border transition-colors cursor-pointer"
                :class="Number(activeTenantId) === Number(item.tenant.id) && Number(activeSiteId) === Number(st.id)
                  ? 'bg-cyan-500/15 border-cyan-500/50 text-cyan-800 dark:text-cyan-300'
                  : 'bg-slate-50 hover:bg-slate-100 dark:bg-slate-800/80 dark:hover:bg-slate-800 border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-200'"
              >
                📍 {{ st.name }}
              </button>
            </div>
          </div>
        </template>
      </div>

      <!-- Footer -->
      <div class="px-5 py-3.5 border-t border-slate-100 dark:border-slate-800 bg-white dark:bg-[#0E1626] flex flex-col sm:flex-row sm:items-center justify-between gap-2.5">
        <!-- Post-Login Convenience Option -->
        <label
          v-if="mode === 'post-login'"
          class="inline-flex items-center gap-2 text-xs text-slate-600 dark:text-slate-300 cursor-pointer select-none"
        >
          <input
            v-model="openOperatorOnPostLoginSelect"
            type="checkbox"
            class="rounded text-blue-600 focus:ring-blue-500 h-4 w-4 border-slate-300 dark:border-slate-700 dark:bg-slate-800"
          />
          <span>Langsung buka <strong>Dashboard Operator</strong> stasiun terpilih</span>
        </label>
        <div v-else class="text-3xs text-slate-500 dark:text-slate-400">
          Tips: Anda dapat berpindah Tenant/Stasiun kapan saja melalui switcher di bagian atas layar.
        </div>

        <div class="flex flex-wrap items-center justify-end gap-2">
          <template v-if="mode === 'post-login'">
            <button
              type="button"
              id="btn-skip-to-global-fleet-map"
              @click="continueInGlobalMode"
              class="px-3.5 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-slate-100 hover:bg-slate-200/80 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs font-semibold cursor-pointer transition-colors"
            >
              🌐 Lewati — Buka Peta Global
            </button>
          </template>

          <template v-else-if="mode === 'route-gate'">
            <button
              type="button"
              @click="goToTenantsAdmin"
              class="px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white hover:bg-slate-100 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-200 text-xs font-semibold cursor-pointer transition-colors"
            >
              🏢 Manajemen Multi-Tenant
            </button>
            <button
              type="button"
              @click="goToFleetMap"
              class="px-3.5 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white btn-text text-xs font-bold shadow-xs cursor-pointer transition-colors"
            >
              🗺️ Buka Peta Global (Fleet Map)
            </button>
          </template>

          <template v-else>
            <button
              v-if="activeTenantId > 0"
              type="button"
              @click="switchToGlobalModeFromModal"
              class="px-3 py-2 rounded-xl border border-purple-200 dark:border-purple-700/60 bg-purple-50 hover:bg-purple-100 dark:bg-purple-950/40 dark:hover:bg-purple-900/50 text-purple-700 dark:text-purple-300 text-xs font-semibold cursor-pointer transition-colors"
            >
              🌐 Semua Tenant (Mode Global)
            </button>
            <button
              type="button"
              @click="handleDismiss"
              class="px-3.5 py-2 rounded-xl border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 text-xs font-semibold cursor-pointer"
            >
              Tutup
            </button>
          </template>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'CloudTenantWorkspaceModal',
  props: {
    visible: { type: Boolean, default: false },
    mode: {
      type: String,
      default: 'switcher', // 'post-login' | 'route-gate' | 'switcher'
    },
    targetRouteLabel: {
      type: String,
      default: '',
    },
    isCurrentRouteGated: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      searchQuery: '',
      yearFilter: 'ALL',
      domainFilter: 'ALL',
      selectingKey: '',
      loadingContext: false,
      openOperatorOnPostLoginSelect: true,
    };
  },
  computed: {
    activeTenantId() {
      return Number((this.$store && this.$store.state && this.$store.state.activeTenantId) || 0);
    },
    activeSiteId() {
      return Number((this.$store && this.$store.state && this.$store.state.activeSiteId) || 0);
    },
    rawTenants() {
      return (this.$store && this.$store.getters && this.$store.getters.availableTenants) || [];
    },
    allKnownSites() {
      const st = (this.$store && this.$store.state) || {};
      if (Array.isArray(st.globalSitesCatalog) && st.globalSitesCatalog.length > 0) {
        return st.globalSitesCatalog;
      }
      return Array.isArray(st.sites) ? st.sites : [];
    },
    enrichedTenants() {
      const sitesByTenant = {};
      for (const s of this.allKnownSites) {
        if (!s) continue;
        const tid = Number(s.tenant_id) || 0;
        if (!sitesByTenant[tid]) sitesByTenant[tid] = [];
        sitesByTenant[tid].push(s);
      }

      return this.rawTenants
        .filter(t => t && String(t.status || 'ACTIVE').toUpperCase() !== 'SUSPENDED')
        .map((t) => {
          const tSites = sitesByTenant[Number(t.id)] || [];
          const primarySite = tSites[0] || null;
          const dom = String((primarySite && primarySite.monitoring_domain) || t.monitoring_domain || '').toUpperCase();
          const yr = Number(t.project_year || (primarySite && primarySite.project_year) || 0);
          return {
            tenant: t,
            sites: tSites,
            primaryDomain: dom || '',
            projectYear: yr,
          };
        });
    },
    yearOptions() {
      const yrs = new Set();
      for (const item of this.enrichedTenants) {
        if (item.projectYear > 0) yrs.add(item.projectYear);
      }
      return Array.from(yrs).sort((a, b) => b - a);
    },
    filteredTenants() {
      const q = String(this.searchQuery || '').trim().toLowerCase();
      return this.enrichedTenants.filter((item) => {
        if (this.yearFilter !== 'ALL' && String(item.projectYear) !== String(this.yearFilter)) {
          return false;
        }
        if (this.domainFilter !== 'ALL') {
          const targetDom = this.domainFilter.toUpperCase();
          const matchesDom = item.primaryDomain === targetDom ||
            (targetDom === 'ONLIMO' && item.primaryDomain === 'WQMS') ||
            item.sites.some(s => {
              const sd = String(s.monitoring_domain || '').toUpperCase();
              return sd === targetDom || (targetDom === 'ONLIMO' && sd === 'WQMS');
            });
          if (!matchesDom) return false;
        }
        if (q) {
          const siteText = item.sites.map(s => `${s.name || ''} ${s.code || ''} ${s.address || ''}`).join(' ');
          const hay = `${item.tenant.name || ''} ${item.tenant.code || ''} ${item.tenant.customer_name || ''} ${item.projectYear || ''} ${item.primaryDomain || ''} ${siteText}`.toLowerCase();
          if (!hay.includes(q)) return false;
        }
        return true;
      });
    },
  },
  watch: {
    visible: {
      immediate: true,
      async handler(val) {
        if (!val) return;
        this.selectingKey = '';
        if (this.$store) {
          const needsTenants = this.rawTenants.length === 0;
          const st = this.$store.state || {};
          const needsGlobalSites = !Array.isArray(st.globalSitesCatalog) || st.globalSitesCatalog.length === 0;
          if (needsTenants || needsGlobalSites) {
            this.loadingContext = needsTenants;
            try {
              if (needsTenants) {
                await this.$store.dispatch('fetchTenantContext');
              }
              if (!Array.isArray(this.$store.state.globalSitesCatalog) || this.$store.state.globalSitesCatalog.length === 0) {
                await this.$store.dispatch('fetchGlobalSitesCatalog');
              }
            } finally {
              this.loadingContext = false;
            }
          }
        }
        this.$nextTick(() => {
          if (this.$refs.searchInput && typeof window !== 'undefined' && window.innerWidth >= 768) {
            this.$refs.searchInput.focus();
          }
        });
      },
    },
  },
  methods: {
    domainBadgeClass(dom) {
      const d = String(dom || '').toUpperCase();
      if (d === 'SPARING') return 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/30';
      if (d === 'ONLIMO' || d === 'WQMS') return 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-300 border-cyan-500/30';
      return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/30';
    },
    resetFilters() {
      this.searchQuery = '';
      this.yearFilter = 'ALL';
      this.domainFilter = 'ALL';
    },
    onBackdropClick() {
      if (this.mode !== 'route-gate') {
        this.handleDismiss();
      }
    },
    handleDismiss() {
      this.$emit('close');
    },
    continueInGlobalMode() {
      this.$emit('close');
      if (this.$route && this.$route.path !== '/monitoring/fleet-map' && this.isCurrentRouteGated) {
        this.$router.push('/monitoring/fleet-map').catch(() => {});
      }
    },
    goToFleetMap() {
      this.$router.push('/monitoring/fleet-map').catch(() => {});
    },
    goToTenantsAdmin() {
      this.$router.push('/administration/tenants').catch(() => {});
    },
    async switchToGlobalModeFromModal() {
      this.selectingKey = '0_0';
      try {
        await this.$store.dispatch('setActiveTenant', 0);
        this.$emit('close');
        if (this.isCurrentRouteGated) {
          this.$router.push('/monitoring/fleet-map').catch(() => {});
        }
      } finally {
        this.selectingKey = '';
      }
    },
    async selectWorkspace(tenantId, siteId = 0) {
      const tId = Number(tenantId) || 0;
      const sId = Number(siteId) || 0;
      if (tId <= 0) return;

      this.selectingKey = `${tId}_${sId}`;
      try {
        await this.$store.dispatch('setActiveTenantAndSite', {
          tenantId: tId,
          siteId: sId,
        });
        this.$emit('selected', { tenantId: tId, siteId: sId });
        this.$emit('close');

        if (
          this.mode === 'post-login' &&
          this.openOperatorOnPostLoginSelect &&
          this.$route &&
          (this.$route.path === '/monitoring/fleet-map' || this.$route.path === '/fleet-map')
        ) {
          this.$router.push('/monitoring/dashboard-operator').catch(() => {});
        }
      } finally {
        this.selectingKey = '';
      }
    },
  },
};
</script>
