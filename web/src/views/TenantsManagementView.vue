<template>
  <div class="space-y-4 font-sans text-slate-800 dark:text-slate-200">
    <!-- Top Header & Tab Switcher -->
    <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm flex flex-col lg:flex-row lg:items-center justify-between gap-4">
      <div class="flex items-center space-x-3">
        <div class="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 text-purple-600 dark:text-purple-400 flex items-center justify-center flex-shrink-0">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3 21h18M5 21V7l8-4v18M19 21V11l-6-4M9 9v.01M9 12v.01M9 15v.01M9 18v.01" />
          </svg>
        </div>
        <div>
          <h1 class="text-base font-bold text-slate-900 dark:text-white tracking-tight flex items-center gap-2 flex-wrap">
            <span>Manajemen Multi-Tenant, Stasiun &amp; Hybrid Edge</span>
            <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20">
              {{ isSuperAdmin ? 'GLOBAL SUPERADMIN' : (isEngineer ? 'TENANT ENGINEER' : 'TENANT ADMIN') }}
            </span>
          </h1>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            Kelola organisasi tenant, stasiun pantau (AQMS / SPARING / ONLIMO), token Bearer API Receiver, serta modul akuisisi Sensor Fisik &amp; MQTT.
          </p>
        </div>
      </div>

      <!-- Tab Switcher -->
      <div class="flex flex-wrap items-center p-1 bg-slate-100 dark:bg-[#0B0F19] rounded-xl border border-slate-200 dark:border-slate-800 text-xs font-medium gap-1">
        <button
          id="tab-tenants"
          type="button"
          @click="activeTab = 'tenants'"
          class="px-3 py-1.5 rounded-lg transition-all flex items-center gap-1.5 cursor-pointer"
          :class="activeTab === 'tenants' ? 'bg-white dark:bg-slate-800 text-purple-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
        >
          <span>🏢 {{ isSuperAdmin ? `Tenant (${tenants.length})` : 'Profil Organisasi' }}</span>
        </button>

        <button
          id="tab-sites"
          type="button"
          @click="activeTab = 'sites'"
          class="px-3 py-1.5 rounded-lg transition-all flex items-center gap-1.5 cursor-pointer"
          :class="activeTab === 'sites' ? 'bg-white dark:bg-slate-800 text-cyan-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
        >
          <span>📍 Stasiun Pantau ({{ sites.length }})</span>
        </button>

        <button
          id="tab-ingest-keys"
          type="button"
          @click="activeTab = 'ingest'"
          class="px-3 py-1.5 rounded-lg transition-all flex items-center gap-1.5 cursor-pointer"
          :class="activeTab === 'ingest' ? 'bg-white dark:bg-slate-800 text-emerald-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
        >
          <span>🔑 API &amp; Bearer Tokens ({{ ingestKeys.length }})</span>
        </button>
      </div>
    </div>

    <!-- Active Workspace Banner for SUPERADMIN -->
    <div
      v-if="isSuperAdmin"
      class="p-3 rounded-xl border flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs"
      :class="activeTenantId > 0
        ? 'bg-purple-50 dark:bg-purple-950/30 border-purple-200 dark:border-purple-800/60 text-purple-900 dark:text-purple-200'
        : 'bg-slate-50 dark:bg-slate-900/60 border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300'"
    >
      <div class="flex items-center gap-2">
        <span class="text-sm">{{ activeTenantId > 0 ? '🏢' : '🌐' }}</span>
        <span v-if="activeTenantId > 0">
          Workspace Aktif: <strong>{{ activeTenantName }}</strong>
          <span v-if="activeSiteId > 0"> &bull; Stasiun: <strong>{{ activeSiteName }}</strong></span>
        </span>
        <span v-else>
          Mode <strong>Global Semua Tenant</strong> aktif — Anda melihat seluruh organisasi, stasiun, dan kunci Edge lintas tenant.
        </span>
      </div>
      <div class="flex items-center gap-2" v-if="activeTenantId > 0 || activeSiteId > 0">
        <button
          type="button"
          @click="resetToGlobalScope"
          class="px-2.5 py-1 rounded-lg bg-white dark:bg-slate-800 border border-purple-300 dark:border-purple-700 text-purple-700 dark:text-purple-300 font-semibold text-2xs hover:bg-purple-100/60 dark:hover:bg-slate-700 transition cursor-pointer"
        >
          🌐 Kembali ke Global Semua Tenant
        </button>
      </div>
    </div>

    <!-- Toast Feedback -->
    <div
      v-if="toastMessage"
      class="p-3 rounded-xl text-xs font-semibold flex items-center justify-between transition-all"
      :class="toastType === 'error'
        ? 'bg-rose-500/10 border border-rose-500/20 text-rose-600 dark:text-rose-400'
        : 'bg-emerald-500/10 border border-emerald-500/20 text-emerald-700 dark:text-emerald-400'"
    >
      <span>{{ toastMessage }}</span>
      <button type="button" @click="toastMessage = ''" class="hover:opacity-75 text-sm leading-none cursor-pointer">&times;</button>
    </div>

    <!-- =================================================================== -->
    <!-- TAB 1: TENANTS (SUPERADMIN CRUD / TENANT ADMIN SELF-SERVICE)        -->
    <!-- =================================================================== -->
    <div v-if="activeTab === 'tenants'" class="space-y-4">
      <!-- 1A. SUPERADMIN FULL TENANT LIST & CRUD -->
      <template v-if="isSuperAdmin">
        <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
          <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4">
            <div class="text-3xs font-semibold uppercase tracking-wider text-slate-500">Total Tenant</div>
            <div class="text-xl font-bold font-mono text-slate-900 dark:text-white mt-1">{{ tenants.length }}</div>
          </div>
          <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4">
            <div class="text-3xs font-semibold uppercase tracking-wider text-slate-500">Tenant Aktif</div>
            <div class="text-xl font-bold font-mono text-emerald-600 dark:text-emerald-400 mt-1">
              {{ tenants.filter(t => t.status === 'ACTIVE').length }}
            </div>
          </div>
          <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4">
            <div class="text-3xs font-semibold uppercase tracking-wider text-slate-500">Total Stasiun</div>
            <div class="text-xl font-bold font-mono text-cyan-600 dark:text-cyan-400 mt-1">{{ totalSitesAcrossTenants }}</div>
          </div>
          <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4">
            <div class="text-3xs font-semibold uppercase tracking-wider text-slate-500">Total Perangkat</div>
            <div class="text-xl font-bold font-mono text-blue-600 dark:text-blue-400 mt-1">{{ totalDevicesAcrossTenants }}</div>
          </div>
        </div>

        <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
          <input
            v-model="tenantSearch"
            type="text"
            placeholder="Cari nama organisasi atau kode slug tenant..."
            class="w-full sm:max-w-sm bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3.5 py-2 text-xs text-slate-800 dark:text-slate-200 focus:outline-none focus:border-purple-500"
          />
          <button
            id="btn-create-tenant"
            type="button"
            @click="openCreateTenantModal"
            class="px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white btn-text text-xs font-bold shadow-xs transition cursor-pointer flex items-center gap-1.5 self-start sm:self-auto"
          >
            <span>+ Tambah Tenant Baru</span>
          </button>
        </div>

        <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
          <div class="overflow-x-auto">
            <table class="w-full text-left text-xs min-w-[920px]">
              <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
                <tr>
                  <th class="px-4 py-3">Organisasi / Tenant</th>
                  <th class="px-4 py-3">Kode</th>
                  <th class="px-4 py-3">Status</th>
                  <th class="px-4 py-3">Stasiun / Kuota</th>
                  <th class="px-4 py-3">Perangkat / Kuota</th>
                  <th class="px-4 py-3">Pengguna</th>
                  <th class="px-4 py-3">Kontak PIC</th>
                  <th class="px-4 py-3 text-right">Aksi</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
                <tr v-for="t in filteredTenants" :key="t.id" class="hover:bg-slate-50/60 dark:hover:bg-slate-800/30">
                  <td class="px-4 py-3">
                    <div class="font-bold text-slate-900 dark:text-white flex items-center gap-2">
                      <span>{{ t.name }}</span>
                      <span
                        v-if="Number(activeTenantId) === Number(t.id)"
                        class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold bg-purple-600 text-white badge-text"
                      >
                        AKTIF
                      </span>
                    </div>
                    <div class="text-3xs text-slate-500 truncate max-w-[260px]">
                      {{ t.customer_name || t.description || '-' }}
                    </div>
                  </td>
                  <td class="px-4 py-3 font-mono text-2xs text-purple-600 dark:text-purple-400 font-semibold">
                    {{ t.code }}
                  </td>
                  <td class="px-4 py-3">
                    <span
                      class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-3xs font-bold"
                      :class="t.status === 'ACTIVE'
                        ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20'
                        : 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border border-rose-500/20'"
                    >
                      {{ t.status }}
                    </span>
                  </td>
                  <td class="px-4 py-3 font-mono text-2xs">
                    {{ t.site_count || 0 }} / {{ t.max_sites > 0 ? t.max_sites : '∞' }}
                  </td>
                  <td class="px-4 py-3 font-mono text-2xs">
                    {{ t.device_count || 0 }} / {{ t.max_devices > 0 ? t.max_devices : '∞' }}
                  </td>
                  <td class="px-4 py-3 font-mono text-2xs">
                    {{ t.user_count || 0 }}
                  </td>
                  <td class="px-4 py-3 text-2xs text-slate-600 dark:text-slate-400">
                    <div>{{ t.contact_name || '-' }}</div>
                    <div v-if="t.contact_email" class="text-3xs font-mono text-slate-400">{{ t.contact_email }}</div>
                  </td>
                  <td class="px-4 py-3 text-right whitespace-nowrap">
                    <div class="inline-flex items-center gap-1.5">
                      <button
                        type="button"
                        @click="enterTenantWorkspace(t)"
                        class="px-2.5 py-1 rounded-lg bg-purple-50 hover:bg-purple-100 dark:bg-purple-950/50 dark:hover:bg-purple-900/60 text-purple-700 dark:text-purple-300 border border-purple-200 dark:border-purple-800/60 text-3xs font-semibold cursor-pointer"
                      >
                        Masuk Workspace
                      </button>
                      <button
                        type="button"
                        @click="openEditTenantModal(t)"
                        class="px-2.5 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 text-3xs font-semibold cursor-pointer"
                      >
                        Edit
                      </button>
                      <button
                        v-if="Number(t.id) !== 1"
                        type="button"
                        @click="confirmDeleteTenant(t)"
                        class="px-2.5 py-1 rounded-lg bg-rose-50 hover:bg-rose-100 dark:bg-rose-950/40 dark:hover:bg-rose-900/50 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-800/60 text-3xs font-semibold cursor-pointer"
                      >
                        Hapus
                      </button>
                    </div>
                  </td>
                </tr>
                <tr v-if="filteredTenants.length === 0">
                  <td colspan="8" class="px-4 py-8 text-center text-slate-500">Belum ada tenant yang sesuai filter.</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </template>

      <!-- 1B. TENANT ADMIN SELF-SERVICE ORGANIZATION & BRANDING -->
      <template v-else>
        <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 shadow-sm space-y-5">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-100 dark:border-slate-800 pb-4">
            <div>
              <h2 class="text-sm font-bold text-slate-900 dark:text-white">Profil Organisasi &amp; Branding Tenant</h2>
              <p class="text-2xs text-slate-500 mt-0.5">
                Perbarui identitas organisasi, kontak PIC, dan nama/logo customer yang tampil di layar monitoring tenant Anda.
              </p>
            </div>
            <div class="flex items-center gap-2 text-2xs font-mono">
              <span class="px-2.5 py-1 rounded-lg bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300">
                Kode: <strong>{{ selfTenantForm.code || '-' }}</strong>
              </span>
              <span class="px-2.5 py-1 rounded-lg bg-cyan-50 dark:bg-cyan-950/40 text-cyan-700 dark:text-cyan-300 border border-cyan-200 dark:border-cyan-800/50">
                Kuota Stasiun: {{ sites.length }} / {{ selfTenantForm.max_sites > 0 ? selfTenantForm.max_sites : '∞' }}
              </span>
            </div>
          </div>

          <form @submit.prevent="saveSelfTenantProfile" class="space-y-4 text-xs">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Nama Organisasi *</label>
                <input
                  v-model="selfTenantForm.name"
                  type="text"
                  required
                  class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
                />
              </div>
              <div>
                <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Nama Tampilan Customer (Header SCADA)</label>
                <input
                  v-model="selfTenantForm.customer_name"
                  type="text"
                  placeholder="Misal: PT Industri Hijau Nusantara"
                  class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
                />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Subjudul Customer</label>
                <input
                  v-model="selfTenantForm.customer_subtitle"
                  type="text"
                  placeholder="Environmental Compliance Division"
                  class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
                />
              </div>
              <div>
                <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Deskripsi</label>
                <input
                  v-model="selfTenantForm.description"
                  type="text"
                  class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
                />
              </div>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Nama Kontak PIC</label>
                <input
                  v-model="selfTenantForm.contact_name"
                  type="text"
                  class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
                />
              </div>
              <div>
                <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Email Kontak</label>
                <input
                  v-model="selfTenantForm.contact_email"
                  type="email"
                  class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
                />
              </div>
              <div>
                <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Telepon / WhatsApp</label>
                <input
                  v-model="selfTenantForm.contact_phone"
                  type="text"
                  class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
                />
              </div>
            </div>

            <div class="flex items-center justify-end pt-3 border-t border-slate-100 dark:border-slate-800">
              <button
                type="submit"
                :disabled="savingTenant"
                class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white btn-text text-xs font-bold shadow-xs transition cursor-pointer disabled:opacity-50"
              >
                {{ savingTenant ? 'Menyimpan...' : 'Simpan Profil Organisasi' }}
              </button>
            </div>
          </form>
        </div>
      </template>
    </div>

    <!-- =================================================================== -->
    <!-- TAB 2: SITES / STASIUN PANTAU                                       -->
    <!-- =================================================================== -->
    <div v-if="activeTab === 'sites'" class="space-y-4">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-sm font-bold text-slate-900 dark:text-white">Daftar Stasiun / Titik Pantau</h2>
          <p class="text-2xs text-slate-500">
            Setiap stasiun memiliki domain datalogger (AQMS / SPARING / ONLIMO), koordinat GPS, dan perangkat serta parameter manual terisolasi.
          </p>
        </div>
        <button
          id="btn-create-site"
          type="button"
          @click="openCreateSiteModal"
          class="px-4 py-2 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-white btn-text text-xs font-bold shadow-xs transition cursor-pointer self-start sm:self-auto"
        >
          + Tambah Stasiun Pantau
        </button>
      </div>

      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs min-w-[880px]">
            <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-4 py-3">Nama Stasiun</th>
                <th class="px-4 py-3">Kode</th>
                <th v-if="isSuperAdmin" class="px-4 py-3">Tenant</th>
                <th class="px-4 py-3">Domain</th>
                <th class="px-4 py-3">Koordinat GPS</th>
                <th class="px-4 py-3">Perangkat</th>
                <th class="px-4 py-3">Status</th>
                <th class="px-4 py-3 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
              <tr v-for="s in sites" :key="s.id" class="hover:bg-slate-50/60 dark:hover:bg-slate-800/30">
                <td class="px-4 py-3">
                  <div class="font-bold text-slate-900 dark:text-white flex items-center gap-2">
                    <span>{{ s.name }}</span>
                    <span
                      v-if="Number(activeSiteId) === Number(s.id)"
                      class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold bg-cyan-600 text-white badge-text"
                    >
                      AKTIF
                    </span>
                  </div>
                  <div class="text-3xs text-slate-500 truncate max-w-[240px]">{{ s.address || s.timezone || 'Asia/Jakarta' }}</div>
                </td>
                <td class="px-4 py-3">
                  <div class="font-mono text-2xs text-cyan-700 dark:text-cyan-400 font-semibold">{{ s.code }}</div>
                  <div v-if="s.legacy_station_id" class="mt-0.5 inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-700 dark:text-amber-300 border border-amber-500/25 font-mono text-4xs font-semibold">
                    station_id: {{ s.legacy_station_id }}
                  </div>
                </td>
                <td v-if="isSuperAdmin" class="px-4 py-3 text-2xs font-medium">{{ tenantNameById(s.tenant_id) }}</td>
                <td class="px-4 py-3">
                  <div class="flex flex-wrap items-center gap-1">
                    <span
                      class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold border"
                      :class="domainBadgeClass(s.monitoring_domain)"
                    >
                      {{ s.monitoring_domain || 'AQMS' }}
                    </span>
                    <span
                      class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold border"
                      :class="ingestFormatBadgeClass(s.ingest_format)"
                      :title="ingestFormatHint(s.ingest_format)"
                    >
                      {{ ingestFormatShortLabel(s.ingest_format) }}
                    </span>
                  </div>
                </td>
                <td class="px-4 py-3 font-mono text-3xs text-slate-600 dark:text-slate-400">
                  <span v-if="s.latitude || s.longitude">{{ Number(s.latitude || 0).toFixed(5) }}, {{ Number(s.longitude || 0).toFixed(5) }}</span>
                  <span v-else class="text-slate-400">Belum diset</span>
                </td>
                <td class="px-4 py-3 font-mono text-2xs">{{ s.device_count || 0 }}</td>
                <td class="px-4 py-3">
                  <span
                    class="px-2 py-0.5 rounded-full text-3xs font-bold"
                    :class="s.enabled !== false
                      ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20'
                      : 'bg-slate-200 dark:bg-slate-800 text-slate-500'"
                  >
                    {{ s.enabled !== false ? 'ENABLED' : 'DISABLED' }}
                  </span>
                </td>
                <td class="px-4 py-3 text-right whitespace-nowrap">
                  <div class="inline-flex items-center gap-1.5">
                    <button
                      type="button"
                      @click="focusSite(s)"
                      class="px-2.5 py-1 rounded-lg bg-cyan-50 hover:bg-cyan-100 dark:bg-cyan-950/50 dark:hover:bg-cyan-900/60 text-cyan-700 dark:text-cyan-300 border border-cyan-200 dark:border-cyan-800/60 text-3xs font-semibold cursor-pointer"
                    >
                      Pilih Stasiun
                    </button>
                    <button
                      type="button"
                      @click="openEditSiteModal(s)"
                      class="px-2.5 py-1 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 text-3xs font-semibold cursor-pointer"
                    >
                      Edit
                    </button>
                    <button
                      type="button"
                      @click="confirmDeleteSite(s)"
                      class="px-2.5 py-1 rounded-lg bg-rose-50 hover:bg-rose-100 dark:bg-rose-950/40 dark:hover:bg-rose-900/50 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-800/60 text-3xs font-semibold cursor-pointer"
                    >
                      Hapus
                    </button>
                  </div>
                </td>
              </tr>
              <tr v-if="sites.length === 0">
                <td :colspan="isSuperAdmin ? 8 : 7" class="px-4 py-8 text-center text-slate-500">
                  Belum ada stasiun pantau terdaftar pada scope ini.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- =================================================================== -->
    <!-- TAB 3: API BEARER TOKENS, ACQUISITION MODULES & PAYLOAD GENERATOR   -->
    <!-- =================================================================== -->
    <div v-if="activeTab === 'ingest'" class="space-y-4">
      <!-- Configurable Telemetry Acquisition Modules (SUPERADMIN) -->
      <div v-if="isSuperAdmin" class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm space-y-3">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-100 dark:border-slate-800 pb-3">
          <div>
            <h2 class="text-sm font-bold text-slate-900 dark:text-white">Konfigurasi Modul Akuisisi Data Server</h2>
            <p class="text-2xs text-slate-500">
              Atur modul pembacaan sensor fisik (Modbus/Serial) dan penerimaan MQTT tanpa mematikan penerima utama HTTP JSON + Bearer Token.
            </p>
          </div>
          <span class="px-2.5 py-1 rounded-lg text-3xs font-mono font-bold bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 border border-emerald-500/20 self-start sm:self-auto">
            HTTP JSON + BEARER: SELALU AKTIF
          </span>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
          <!-- Physical Sensor Polling Toggle -->
          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 flex items-center justify-between gap-3">
            <div>
              <div class="font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <span>🔌 Modul Pembacaan Sensor Fisik (Modbus RTU / TCP / Serial)</span>
              </div>
              <p class="text-3xs text-slate-500 mt-0.5 leading-relaxed">
                Nonaktifkan pada server pusat cloud murni agar hanya menerima data push API/MQTT, atau aktifkan untuk hybrid gateway.
              </p>
            </div>
            <button
              id="btn-toggle-sensor-polling"
              type="button"
              :disabled="savingModules || acqModules.cloud_mode"
              @click="toggleAcquisitionModule('sensor_polling_enabled')"
              class="px-3 py-1.5 rounded-xl font-mono text-3xs font-bold border transition flex-shrink-0"
              :class="acqModules.cloud_mode
                ? 'bg-sky-500/15 text-sky-700 dark:text-sky-300 border-sky-500/40 cursor-not-allowed'
                : acqModules.sensor_polling_enabled
                  ? 'bg-emerald-600 text-white btn-text border-emerald-600 cursor-pointer'
                  : 'bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-300 dark:border-slate-700 cursor-pointer'"
            >
              {{ acqModules.cloud_mode ? 'OFF (CLOUD MODE)' : (acqModules.sensor_polling_enabled ? 'ENABLED' : 'PAUSED') }}
            </button>
          </div>

          <!-- MQTT Receiver Toggle -->
          <div class="p-3.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 flex items-center justify-between gap-3">
            <div>
              <div class="font-bold text-slate-900 dark:text-white flex items-center gap-2">
                <span>📡 Modul Penerimaan Telemetri MQTT (<code class="font-mono text-3xs">/api/v1/ingest/mqtt</code>)</span>
              </div>
              <p class="text-3xs text-slate-500 mt-0.5 leading-relaxed">
                Izinkan admin/integrator meneruskan pesan dari Broker MQTT (<code class="font-mono">{{ acqModules.mqtt_topic_prefix || 'datalogger/ingest/+' }}</code>) dengan Bearer Token.
              </p>
            </div>
            <button
              id="btn-toggle-mqtt-ingest"
              type="button"
              :disabled="savingModules"
              @click="toggleAcquisitionModule('mqtt_ingest_enabled')"
              class="px-3 py-1.5 rounded-xl font-mono text-3xs font-bold border transition cursor-pointer flex-shrink-0"
              :class="acqModules.mqtt_ingest_enabled
                ? 'bg-cyan-600 text-white btn-text border-cyan-600'
                : 'bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400 border-slate-300 dark:border-slate-700'"
            >
              {{ acqModules.mqtt_ingest_enabled ? 'ENABLED' : 'DISABLED' }}
            </button>
          </div>
        </div>
      </div>

      <!-- API Bearer Keys Header -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <div>
          <h2 class="text-sm font-bold text-slate-900 dark:text-white">Telemetry Receiver Bearer Tokens &amp; Switch Format Penerimaan</h2>
          <p class="text-2xs text-slate-500">
            Autentikasi via header <code class="font-mono text-emerald-600 dark:text-emerald-400">Authorization: Bearer ...</code> pada endpoint standar <code class="font-mono">/api/v1/ingest/telemetry</code> maupun endpoint Cloud lama <code class="font-mono">/api/v1/ticks/create_multi</code>. Pilih format penerimaan langsung per token.
          </p>
        </div>
        <button
          id="btn-create-ingest-key"
          type="button"
          @click="openCreateKeyModal"
          class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white btn-text text-xs font-bold shadow-xs transition cursor-pointer self-start sm:self-auto"
        >
          + Buat Bearer Token Baru
        </button>
      </div>

      <!-- Newly Minted Key One-Time Banner -->
      <div
        v-if="newlyCreatedPlainKey"
        class="p-4 rounded-2xl bg-amber-50 dark:bg-amber-950/30 border-2 border-amber-300 dark:border-amber-700/70 space-y-2.5"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <div class="text-xs font-bold text-amber-900 dark:text-amber-200">
              ⚠️ Simpan Bearer Token Ini Sekarang (Hanya Ditampilkan Sekali!)
            </div>
            <p class="text-2xs text-amber-800 dark:text-amber-300/90 mt-0.5">
              Server hanya menyimpan hash SHA-256. Gunakan pada header <code class="font-mono font-bold">Authorization: Bearer {{ newlyCreatedPlainKey }}</code>:
            </p>
          </div>
          <button type="button" @click="newlyCreatedPlainKey = ''" class="text-amber-700 dark:text-amber-300 text-xs underline cursor-pointer">
            Tutup
          </button>
        </div>
        <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-2">
          <input
            type="text"
            readonly
            :value="newlyCreatedPlainKey"
            class="flex-1 font-mono text-xs px-3 py-2 rounded-xl bg-white dark:bg-[#0B0F19] border border-amber-300 dark:border-amber-700 text-slate-900 dark:text-emerald-300 font-bold"
          />
          <button
            type="button"
            @click="copyPlainKey(newlyCreatedPlainKey)"
            class="px-4 py-2 rounded-xl bg-amber-600 hover:bg-amber-500 text-white btn-text text-xs font-bold cursor-pointer"
          >
            📋 Salin Bearer Token
          </button>
        </div>
      </div>

      <!-- Keys Table -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs min-w-[860px]">
            <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-4 py-3">Nama Token</th>
                <th class="px-4 py-3">Bearer Prefix</th>
                <th v-if="isSuperAdmin" class="px-4 py-3">Tenant</th>
                <th class="px-4 py-3">Target Stasiun</th>
                <th class="px-4 py-3">Switch Format Penerimaan</th>
                <th class="px-4 py-3">Pemakaian Terakhir</th>
                <th class="px-4 py-3">Status</th>
                <th class="px-4 py-3 text-right">Aksi</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60">
              <tr v-for="k in ingestKeys" :key="k.id" class="hover:bg-slate-50/60 dark:hover:bg-slate-800/30">
                <td class="px-4 py-3 font-bold text-slate-900 dark:text-white">{{ k.name }}</td>
                <td class="px-4 py-3 font-mono text-2xs text-emerald-700 dark:text-emerald-400">Bearer {{ k.key_prefix }}••••••</td>
                <td v-if="isSuperAdmin" class="px-4 py-3 text-2xs">{{ tenantNameById(k.tenant_id) }}</td>
                <td class="px-4 py-3 text-2xs">
                  {{ k.site_id ? siteNameById(k.site_id) : 'Semua Stasiun Tenant (site_code / station_id)' }}
                </td>
                <td class="px-4 py-3">
                  <div class="flex flex-col gap-1 max-w-[235px]">
                    <select
                      :value="normalizeIngestFormat(k.ingest_format)"
                      @change="updateIngestKeyFormat(k, $event.target.value)"
                      class="w-full rounded-lg px-2 py-1 text-3xs font-mono font-bold border cursor-pointer focus:outline-none"
                      :class="ingestFormatBadgeClass(k.ingest_format)"
                    >
                      <option value="AUTO">AUTO · Hybrid (Baru + Lama)</option>
                      <option value="STANDARD_V1">STANDARD_V1 · Format Baru</option>
                      <option value="CBI_LEGACY">CBI_LEGACY · create_multi</option>
                    </select>
                    <span class="text-4xs font-mono text-slate-400">STRICT MANUAL PARAMS</span>
                  </div>
                </td>
                <td class="px-4 py-3 font-mono text-3xs text-slate-500">
                  <div v-if="k.last_used_at">{{ formatDateTime(k.last_used_at) }} ({{ k.usage_count || 0 }}x)</div>
                  <div v-else>Belum pernah dipakai</div>
                </td>
                <td class="px-4 py-3">
                  <button
                    type="button"
                    @click="toggleIngestKey(k)"
                    class="px-2 py-0.5 rounded-full text-3xs font-bold cursor-pointer border"
                    :class="k.enabled
                      ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/30'
                      : 'bg-rose-500/10 text-rose-600 dark:text-rose-400 border-rose-500/30'"
                  >
                    {{ k.enabled ? 'ACTIVE' : 'DISABLED' }}
                  </button>
                </td>
                <td class="px-4 py-3 text-right">
                  <button
                    type="button"
                    @click="revokeIngestKey(k)"
                    class="px-2.5 py-1 rounded-lg bg-rose-50 hover:bg-rose-100 dark:bg-rose-950/40 dark:hover:bg-rose-900/50 text-rose-600 dark:text-rose-400 border border-rose-200 dark:border-rose-800/60 text-3xs font-semibold cursor-pointer"
                  >
                    Cabut Token
                  </button>
                </td>
              </tr>
              <tr v-if="ingestKeys.length === 0">
                <td :colspan="isSuperAdmin ? 8 : 7" class="px-4 py-8 text-center text-slate-500">
                  Belum ada Bearer Token dibuat.
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Interactive Standardized JSON Payload & Bearer Token Builder -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm space-y-4 text-xs">
        <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 border-b border-slate-100 dark:border-slate-800 pb-3">
          <div>
            <div class="font-bold text-sm text-slate-900 dark:text-white flex items-center gap-2">
              <span>🧾 Generator Payload JSON Standar &amp; Bearer Token (AQMS · SPARING · ONLIMO)</span>
            </div>
            <p class="text-2xs text-slate-500 mt-0.5">
              Pilih jenis datalogger, stasiun, dan perangkat manual untuk menghasilkan contoh JSON siap pakai yang terhubung langsung ke Dashboard Operator.
            </p>
          </div>
          <div class="flex items-center gap-1.5 flex-wrap">
            <button
              v-for="lt in ['AQMS', 'SPARING', 'ONLIMO']"
              :key="'lt-btn-' + lt"
              type="button"
              @click="selectPayloadLoggerType(lt)"
              class="px-3 py-1.5 rounded-lg font-mono text-2xs font-bold border transition cursor-pointer"
              :class="payloadBuilder.loggerType === lt
                ? 'bg-emerald-600 text-white border-emerald-600 btn-text shadow-2xs'
                : 'bg-slate-50 dark:bg-[#0B0F19] text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-800 hover:border-emerald-400'"
            >
              {{ lt }}
            </button>
          </div>
        </div>

        <!-- Station & Registered Device Selector -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Stasiun Target (site_code)</label>
            <select
              v-model.number="payloadBuilder.siteId"
              @change="onPayloadSiteChange"
              class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-2xs text-slate-900 dark:text-white"
            >
              <option :value="0">-- Contoh Template Standar ({{ payloadBuilder.loggerType }}) --</option>
              <option v-for="s in sites" :key="'pb-site-' + s.id" :value="s.id">
                {{ s.code }} — {{ s.name }} ({{ s.monitoring_domain || 'AQMS' }})
              </option>
            </select>
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Perangkat Terdaftar Manual (device_code)</label>
            <select
              v-model.number="payloadBuilder.deviceId"
              @change="syncSelectedBuilderDeviceTimeout"
              class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-2xs text-slate-900 dark:text-white"
            >
              <option :value="0">-- Gunakan Contoh Kode Device --</option>
              <option v-for="d in filteredDevicesForBuilder" :key="'pb-dev-' + d.id" :value="d.id">
                {{ d.device_code || d.code }} ({{ (d.parameters || []).filter(p => p.enabled !== false).length }} parameter manual)
              </option>
            </select>
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Mode Contoh Payload</label>
            <div class="flex flex-wrap items-center p-1 bg-slate-100 dark:bg-[#0B0F19] rounded-xl border border-slate-200 dark:border-slate-800 gap-1">
              <button
                v-for="m in effectivePayloadFormatTabs"
                :key="m.id"
                type="button"
                @click="payloadBuilder.mode = m.id"
                class="flex-1 min-w-fit px-2 py-1 rounded-lg text-3xs font-bold transition cursor-pointer text-center whitespace-nowrap"
                :class="payloadBuilder.mode === m.id
                  ? 'bg-white dark:bg-slate-800 text-emerald-700 dark:text-emerald-300 shadow-2xs'
                  : 'text-slate-500 hover:text-slate-800 dark:hover:text-white'"
              >
                {{ m.label }}
              </button>
            </div>
          </div>
        </div>

        <!-- Direct Offline / Communication Timeout Config for Selected Tenant Device -->
        <div
          v-if="selectedBuilderDevice"
          class="p-3.5 rounded-xl bg-sky-50/70 dark:bg-sky-950/20 border border-sky-200 dark:border-sky-800/60 space-y-2.5"
        >
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div>
              <div class="text-xs font-bold text-sky-950 dark:text-sky-200 flex items-center gap-1.5">
                <span>⏱️ Batas Waktu Terima Data / Offline Timeout:</span>
                <code class="font-mono text-sky-700 dark:text-sky-300">{{ selectedBuilderDevice.device_code || selectedBuilderDevice.code }}</code>
              </div>
              <p class="text-3xs text-sky-800/90 dark:text-sky-300/80 mt-0.5">
                Tentukan jeda maksimal jika data belum diterima dari tenant sebelum perangkat &amp; kartu parameter dinyatakan <strong>Offline</strong> (<code class="font-mono">COMMUNICATION_TIMEOUT</code>).
              </p>
            </div>
            <div class="flex items-center gap-2">
              <input
                v-model.number="deviceTimeoutSeconds"
                type="number"
                min="30"
                max="86400"
                step="30"
                class="w-24 px-2.5 py-1.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-sky-300 dark:border-sky-700 font-mono text-xs font-bold text-slate-900 dark:text-white"
              />
              <span class="text-3xs font-mono text-sky-800 dark:text-sky-300 whitespace-nowrap">
                detik (={{ Math.round((Number(deviceTimeoutSeconds) || 300) / 60 * 10) / 10 }}m)
              </span>
              <button
                type="button"
                :disabled="savingDeviceTimeout"
                @click="saveBuilderDeviceTimeout"
                class="px-3 py-1.5 rounded-lg bg-sky-600 hover:bg-sky-500 text-white btn-text text-3xs font-bold cursor-pointer whitespace-nowrap"
              >
                {{ savingDeviceTimeout ? 'Menyimpan...' : 'Simpan Timeout' }}
              </button>
            </div>
          </div>
          <div class="flex flex-wrap items-center gap-1.5">
            <span class="text-3xs font-semibold text-sky-800 dark:text-sky-300 mr-1">Preset Cepat:</span>
            <button
              v-for="pr in [
                { sec: 180, label: '3 Menit (180s)' },
                { sec: 300, label: '5 Menit (300s)' },
                { sec: 600, label: '10 Menit (600s)' },
                { sec: 900, label: '15 Menit (900s)' },
                { sec: 1800, label: '30 Menit (1800s)' },
                { sec: 3900, label: '65 Menit (3900s)' }
              ]"
              :key="'dev-to-' + pr.sec"
              type="button"
              @click="deviceTimeoutSeconds = pr.sec"
              class="px-2 py-1 rounded-md text-3xs font-mono font-semibold border transition cursor-pointer"
              :class="Number(deviceTimeoutSeconds) === pr.sec
                ? 'bg-sky-600 text-white btn-text border-sky-600'
                : 'bg-white dark:bg-[#0B0F19] text-slate-700 dark:text-slate-300 border-sky-200 dark:border-sky-800'"
            >
              {{ pr.label }}
            </button>
          </div>
          <p
            v-if="payloadBuilder.loggerType === 'SPARING'"
            class="text-3xs text-sky-900 dark:text-sky-200 bg-sky-100/80 dark:bg-sky-900/30 border border-sky-200 dark:border-sky-800/70 rounded-lg px-2.5 py-1.5"
          >
            💡 <strong>Pengiriman SPARING Per Jam (30 Data × 2 Menit):</strong> Pilih preset <strong>65 Menit (3900s)</strong> lalu klik <strong>Simpan Timeout</strong> agar status perangkat tetap <strong>ONLINE</strong> di antara siklus pengiriman tiap jam.
          </p>

          <!-- Quick Legacy CBI IDs Mapping (station_id & module_id) -->
          <div class="pt-2.5 border-t border-sky-200/80 dark:border-sky-800/60 flex flex-col lg:flex-row lg:items-center justify-between gap-2.5">
            <div>
              <div class="text-2xs font-bold text-sky-950 dark:text-sky-200">
                🔗 Mapping ID Cloud Lama (<code class="font-mono">create_multi</code>)
              </div>
              <p class="text-3xs text-sky-800/90 dark:text-sky-300/80">
                Cocokkan <code class="font-mono">station_id</code> &amp; <code class="font-mono">module_id</code> dari JSON Cloud lama (misal <code class="font-mono">202</code> &amp; <code class="font-mono">199</code>) ke stasiun &amp; perangkat ini.
              </p>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <div v-if="selectedBuilderSite" class="flex items-center gap-1">
                <span class="text-3xs font-mono font-semibold text-sky-900 dark:text-sky-300">station_id:</span>
                <input
                  v-model.trim="siteLegacyStationId"
                  type="text"
                  placeholder="202"
                  class="w-20 px-2 py-1 rounded-lg bg-white dark:bg-[#0B0F19] border border-sky-300 dark:border-sky-700 font-mono text-xs text-slate-900 dark:text-white"
                />
              </div>
              <div class="flex items-center gap-1">
                <span class="text-3xs font-mono font-semibold text-sky-900 dark:text-sky-300">module_id:</span>
                <input
                  v-model.trim="deviceLegacyModuleId"
                  type="text"
                  placeholder="199"
                  class="w-20 px-2 py-1 rounded-lg bg-white dark:bg-[#0B0F19] border border-sky-300 dark:border-sky-700 font-mono text-xs text-slate-900 dark:text-white"
                />
              </div>
              <button
                type="button"
                :disabled="savingLegacyMapping"
                @click="saveBuilderLegacyMapping"
                class="px-3 py-1.5 rounded-lg bg-amber-600 hover:bg-amber-500 text-white btn-text text-3xs font-bold cursor-pointer whitespace-nowrap"
              >
                {{ savingLegacyMapping ? 'Menyimpan...' : 'Simpan ID Cloud Lama' }}
              </button>
            </div>
          </div>
        </div>

        <!-- Registered Parameter Chips Notice -->
        <div class="p-3 rounded-xl border flex flex-col sm:flex-row sm:items-center justify-between gap-2"
          :class="selectedBuilderDeviceParams.length > 0
            ? 'bg-emerald-50/70 dark:bg-emerald-950/20 border-emerald-200 dark:border-emerald-800/50 text-emerald-900 dark:text-emerald-200'
            : 'bg-slate-50 dark:bg-[#0B0F19] border-slate-200 dark:border-slate-800 text-slate-600 dark:text-slate-400'"
        >
          <div class="text-2xs">
            <template v-if="selectedBuilderDeviceParams.length > 0">
              <strong>Menggunakan {{ selectedBuilderDeviceParams.length }} Parameter Manual Aktual</strong> pada perangkat
              <code class="font-mono font-bold">{{ selectedBuilderDevice.device_code || selectedBuilderDevice.code }}</code>:
              <span class="inline-flex flex-wrap gap-1 ml-1">
                <code
                  v-for="p in selectedBuilderDeviceParams"
                  :key="'chip-' + p.id"
                  class="px-1.5 py-0.5 rounded bg-white dark:bg-slate-900 border border-emerald-300/70 dark:border-emerald-700/60 font-mono text-3xs"
                >
                  {{ p.parameter_code || p.code }}{{ p.unit ? ` (${p.unit})` : '' }}
                </code>
              </span>
            </template>
            <template v-else>
              Menampilkan <strong>Template Parameter Standar {{ payloadBuilder.loggerType }}</strong>. Buat Device (<code class="font-mono">EDGE_PUSH</code> / <code class="font-mono">MQTT</code>) &amp; Parameter secara manual di menu <strong>Devices</strong> agar key disesuaikan otomatis.
            </template>
          </div>
          <div class="flex items-center gap-1.5 flex-shrink-0">
            <button
              v-if="payloadBuilder.mode !== 'curl' && payloadBuilder.mode !== 'mqtt'"
              type="button"
              @click="copyPlainKey(generatedActiveCurlCommand)"
              class="px-3 py-1.5 rounded-xl bg-sky-600 hover:bg-sky-500 text-white btn-text text-3xs font-bold cursor-pointer"
            >
              ⚡ Salin cURL
            </button>
            <button
              type="button"
              @click="copyPlainKey(generatedPayloadSnippet)"
              class="px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white btn-text text-3xs font-bold cursor-pointer"
            >
              📋 Salin Payload
            </button>
          </div>
        </div>

        <!-- Generated Code Preview -->
        <div class="rounded-xl bg-slate-950 text-emerald-300 border border-slate-800 p-3.5 overflow-x-auto">
          <div class="flex flex-wrap items-center justify-between gap-2 text-3xs font-mono text-slate-400 pb-2 mb-2 border-b border-slate-800">
            <span v-if="payloadBuilder.mode === 'cbi_multi_tick' || payloadBuilder.mode === 'cbi_hourly'">
              POST {{ activeHttpIngestEndpoint }} <span class="text-slate-500">(juga kompatibel di /api/v1/ingest/telemetry)</span> &bull; Authorization: Bearer {{ previewBearerToken }}
            </span>
            <span v-else-if="payloadBuilder.mode !== 'mqtt'">
              POST {{ centralIngestEndpoint }} &bull; Authorization: Bearer {{ previewBearerToken }}
            </span>
            <span v-else>
              POST {{ centralMqttIngestEndpoint }} &bull; Topic: datalogger/ingest/{{ resolvedBuilderSiteCode }}/{{ resolvedBuilderDeviceCode }}
            </span>
            <span class="text-emerald-400 font-bold">{{ payloadBuilder.loggerType }} &bull; {{ payloadBuilder.mode === 'cbi_multi_tick' || payloadBuilder.mode === 'cbi_hourly' ? 'CBI_LEGACY' : 'STANDARD_V1' }}</span>
          </div>
          <pre class="text-2xs font-mono leading-relaxed whitespace-pre">{{ generatedPayloadSnippet }}</pre>
        </div>
      </div>
    </div>

    <!-- =================================================================== -->
    <!-- MODAL 1: CREATE / EDIT TENANT (SUPERADMIN)                          -->
    <!-- =================================================================== -->
    <div v-if="showTenantModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-xl p-5 shadow-2xl space-y-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-sm text-slate-900 dark:text-white">
            {{ editingTenant ? `Edit Tenant: ${editingTenant.name}` : 'Tambah Organisasi Tenant Baru' }}
          </h3>
          <button type="button" @click="showTenantModal = false" class="text-slate-400 hover:text-slate-700 dark:hover:text-white text-lg cursor-pointer">&times;</button>
        </div>

        <form @submit.prevent="submitTenantModal" class="space-y-3 text-xs">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Nama Organisasi *</label>
              <input
                v-model="tenantForm.name"
                type="text"
                required
                placeholder="PT Kawasan Industri Nusantara"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Kode Slug Tenant *</label>
              <input
                v-model="tenantForm.code"
                type="text"
                required
                :disabled="editingTenant && Number(editingTenant.id) === 1"
                placeholder="kin-nusantara"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-slate-900 dark:text-white disabled:opacity-60"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Status</label>
              <select
                v-model="tenantForm.status"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
              >
                <option value="ACTIVE">ACTIVE</option>
                <option value="SUSPENDED">SUSPENDED</option>
              </select>
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Kuota Stasiun (0 = ∞)</label>
              <input
                v-model.number="tenantForm.max_sites"
                type="number"
                min="0"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-slate-900 dark:text-white"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Kuota Device (0 = ∞)</label>
              <input
                v-model.number="tenantForm.max_devices"
                type="number"
                min="0"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-slate-900 dark:text-white"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Nama Branding Customer</label>
              <input
                v-model="tenantForm.customer_name"
                type="text"
                placeholder="Nama header SCADA untuk tenant ini"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Subjudul Branding</label>
              <input
                v-model="tenantForm.customer_subtitle"
                type="text"
                placeholder="Environmental Monitoring Network"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Kontak PIC</label>
              <input
                v-model="tenantForm.contact_name"
                type="text"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Email PIC</label>
              <input
                v-model="tenantForm.contact_email"
                type="email"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Telepon PIC</label>
              <input
                v-model="tenantForm.contact_phone"
                type="text"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
              />
            </div>
          </div>

          <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
            <button type="button" @click="showTenantModal = false" class="px-3.5 py-2 rounded-xl bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 cursor-pointer">
              Batal
            </button>
            <button type="submit" :disabled="savingTenant" class="px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white btn-text font-bold cursor-pointer">
              {{ savingTenant ? 'Menyimpan...' : (editingTenant ? 'Simpan Perubahan' : 'Buat Tenant') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- =================================================================== -->
    <!-- MODAL 2: CREATE / EDIT SITE                                         -->
    <!-- =================================================================== -->
    <div v-if="showSiteModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg p-5 shadow-2xl space-y-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-sm text-slate-900 dark:text-white">
            {{ editingSite ? `Edit Stasiun: ${editingSite.name}` : 'Tambah Stasiun / Titik Pantau Baru' }}
          </h3>
          <button type="button" @click="showSiteModal = false" class="text-slate-400 hover:text-slate-700 dark:hover:text-white text-lg cursor-pointer">&times;</button>
        </div>

        <form @submit.prevent="submitSiteModal" class="space-y-3 text-xs">
          <div v-if="isSuperAdmin && !editingSite">
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Pemilik Organisasi Tenant *</label>
            <select
              v-model.number="siteForm.tenant_id"
              required
              class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
            >
              <option v-for="t in tenants" :key="'st-t-' + t.id" :value="t.id">{{ t.name }} ({{ t.code }})</option>
            </select>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Nama Stasiun *</label>
              <input
                v-model="siteForm.name"
                type="text"
                required
                placeholder="Stasiun SPARING Outlet IPAL #1"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Kode Stasiun *</label>
              <input
                v-model="siteForm.code"
                type="text"
                required
                placeholder="sparing-ipal-01"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-slate-900 dark:text-white"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Domain Kepatuhan *</label>
              <select
                v-model="siteForm.monitoring_domain"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-slate-900 dark:text-white"
              >
                <option value="AQMS">AQMS (Kualitas Udara Ambien &amp; ISPU)</option>
                <option value="SPARING">SPARING (Pemantauan Air Limbah Industri)</option>
                <option value="ONLIMO">ONLIMO (Pemantauan Kualitas Air Sungai / Permukaan)</option>
                <option value="WQMS">WQMS (Water Quality Umum)</option>
              </select>
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Zona Waktu</label>
              <select
                v-model="siteForm.timezone"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-slate-900 dark:text-white"
              >
                <option value="Asia/Jakarta">Asia/Jakarta (WIB)</option>
                <option value="Asia/Makassar">Asia/Makassar (WITA)</option>
                <option value="Asia/Jayapura">Asia/Jayapura (WIT)</option>
                <option value="UTC">UTC</option>
              </select>
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3 p-3 rounded-xl bg-slate-50/80 dark:bg-[#0B0F19]/70 border border-slate-200 dark:border-slate-800">
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Switch Format Penerimaan Data</label>
              <select
                v-model="siteForm.ingest_format"
                class="w-full bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-700 rounded-xl px-3 py-2 font-mono text-2xs text-slate-900 dark:text-white"
              >
                <option value="AUTO">AUTO — Hybrid (Baru &amp; Cloud Lama)</option>
                <option value="STANDARD_V1">STANDARD_V1 — Hanya Format Baru</option>
                <option value="CBI_LEGACY">CBI_LEGACY — Hanya Cloud Lama (create_multi)</option>
              </select>
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Legacy Station ID (station_id)</label>
              <input
                v-model.trim="siteForm.legacy_station_id"
                type="text"
                placeholder="Contoh: 202 (opsional)"
                class="w-full bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-700 rounded-xl px-3 py-2 font-mono text-2xs text-slate-900 dark:text-white"
              />
            </div>
          </div>

          <div class="space-y-1.5">
            <label class="block text-3xs font-semibold uppercase text-slate-500">Cari Lokasi / Klik Peta GPS Stasiun</label>
            <div class="h-52 w-full rounded-xl overflow-hidden">
              <StationMapViewer
                :latitude="siteFormPickerLat"
                :longitude="siteFormPickerLng"
                :editable="true"
                height="208px"
                @coordinate-change="onSiteMapCoordinatePicked"
                @location-selected="onSiteLocationSelected"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Latitude</label>
              <input
                v-model.number="siteForm.latitude"
                type="number"
                step="0.000001"
                placeholder="-6.2088"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-slate-900 dark:text-white"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Longitude</label>
              <input
                v-model.number="siteForm.longitude"
                type="number"
                step="0.000001"
                placeholder="106.8456"
                class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-slate-900 dark:text-white"
              />
            </div>
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Alamat Lokasi</label>
            <input
              v-model="siteForm.address"
              type="text"
              placeholder="Kawasan Industri Blok C-12"
              class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
            />
          </div>

          <label class="flex items-center gap-2 pt-1 cursor-pointer">
            <input type="checkbox" v-model="siteForm.enabled" class="rounded text-cyan-600" />
            <span class="text-xs text-slate-700 dark:text-slate-300">Stasiun aktif beroperasi</span>
          </label>

          <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
            <button type="button" @click="showSiteModal = false" class="px-3.5 py-2 rounded-xl bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 cursor-pointer">
              Batal
            </button>
            <button type="submit" :disabled="savingSite" class="px-4 py-2 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-white btn-text font-bold cursor-pointer">
              {{ savingSite ? 'Menyimpan...' : (editingSite ? 'Simpan Stasiun' : 'Tambah Stasiun') }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- =================================================================== -->
    <!-- MODAL 3: CREATE EDGE INGEST API KEY                                 -->
    <!-- =================================================================== -->
    <div v-if="showCreateKeyModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-lg p-5 shadow-2xl space-y-4 max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-sm text-slate-900 dark:text-white">Buat Bearer Token &amp; Pilih Format Penerimaan</h3>
          <button type="button" @click="showCreateKeyModal = false" class="text-slate-400 hover:text-slate-700 dark:hover:text-white text-lg cursor-pointer">&times;</button>
        </div>

        <form @submit.prevent="submitCreateKey" class="space-y-3 text-xs">
          <div v-if="isSuperAdmin">
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Tenant Target *</label>
            <select
              v-model.number="keyForm.tenant_id"
              required
              class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
            >
              <option v-for="t in tenants" :key="'key-t-' + t.id" :value="t.id">{{ t.name }} ({{ t.code }})</option>
            </select>
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Nama Kunci / Unit Edge *</label>
            <input
              v-model="keyForm.name"
              type="text"
              required
              placeholder="Misal: Logger AQMS Stasiun 202 (Cloud Lama)"
              class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
            />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Kunci ke Stasiun Tertentu (Opsional)</label>
            <select
              v-model.number="keyForm.site_id"
              class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-900 dark:text-white"
            >
              <option :value="0">Semua Stasiun di Tenant (Dinamis via site_code / station_id)</option>
              <option v-for="s in filteredSitesForKeyModal" :key="'key-s-' + s.id" :value="s.id">
                {{ s.name }} ({{ s.code }}{{ s.legacy_station_id ? ` · station_id:${s.legacy_station_id}` : '' }})
              </option>
            </select>
          </div>

          <!-- Switch Format Penerimaan Data -->
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1.5">Switch Format Penerimaan Data *</label>
            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2">
              <button
                v-for="opt in ingestFormatOptions"
                :key="'fmt-opt-' + opt.value"
                type="button"
                @click="keyForm.ingest_format = opt.value"
                class="p-2.5 rounded-xl border text-left transition cursor-pointer space-y-0.5"
                :class="keyForm.ingest_format === opt.value
                  ? 'bg-emerald-600 text-white btn-text border-emerald-600 shadow-xs'
                  : 'bg-slate-50 dark:bg-[#0B0F19] text-slate-700 dark:text-slate-300 border-slate-200 dark:border-slate-800 hover:border-emerald-400'"
              >
                <div class="font-mono text-2xs font-bold">{{ opt.shortTitle }}</div>
                <div class="text-4xs opacity-90 leading-snug">{{ opt.desc }}</div>
              </button>
            </div>
          </div>

          <!-- Optional Custom Legacy Bearer Token -->
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">
              Token Bearer Eksisting Cloud Lama (Opsional)
            </label>
            <input
              v-model.trim="keyForm.custom_token"
              type="text"
              placeholder="Kosongkan untuk generate otomatis dlg_live_..., atau tempel token lama"
              class="w-full bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 font-mono text-2xs text-slate-900 dark:text-white"
            />
            <p class="text-3xs text-slate-500 mt-1">
              Jika datalogger lapangan sudah memiliki <code class="font-mono">Authorization: Bearer ...</code> dari Cloud lama, tempel tokennya di sini agar alat langsung terhubung tanpa re-konfigurasi.
            </p>
          </div>

          <div class="p-3 rounded-xl bg-emerald-50/70 dark:bg-emerald-950/20 border border-emerald-200 dark:border-emerald-800/50 space-y-1">
            <div class="font-semibold text-emerald-900 dark:text-emerald-200 flex items-center justify-between">
              <span>Registrasi Device &amp; Parameter Manual + Auto-Alias</span>
              <span class="text-4xs font-mono px-1.5 py-0.5 rounded bg-emerald-100 dark:bg-emerald-500/20 text-emerald-800 dark:text-emerald-300 font-bold">STRICT SCHEMA</span>
            </div>
            <div class="text-3xs text-emerald-800/90 dark:text-emerald-300/80 leading-relaxed">
              Parameter tidak dibuat otomatis. Pada mode <code class="font-mono">AUTO</code> / <code class="font-mono">CBI_LEGACY</code>, nama sensor lama (<code class="font-mono">temp</code>, <code class="font-mono">pressure</code>, <code class="font-mono">wspeed</code>, <code class="font-mono">intensity</code>, <code class="font-mono">tprecipitation</code>, <code class="font-mono">wind_angle</code>, <code class="font-mono">pm25</code>) otomatis dicocokkan ke parameter terdaftar.
            </div>
          </div>

          <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
            <button type="button" @click="showCreateKeyModal = false" class="px-3.5 py-2 rounded-xl bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-300 cursor-pointer">
              Batal
            </button>
            <button type="submit" :disabled="savingKey" class="px-4 py-2 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white btn-text font-bold cursor-pointer">
              {{ savingKey ? 'Menyimpan...' : (keyForm.custom_token ? 'Simpan Token Eksisting' : 'Generate Bearer Token') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import StationMapViewer from '../components/StationMapViewer.vue';

export default {
  name: 'TenantsManagementView',
  components: {
    StationMapViewer,
  },
  data() {
    return {
      activeTab: 'tenants',
      toastMessage: '',
      toastType: 'success',
      tenantSearch: '',

      tenants: [],
      sites: [],
      ingestKeys: [],
      registeredDevices: [],

      savingTenant: false,
      savingSite: false,
      savingKey: false,
      savingModules: false,
      savingDeviceTimeout: false,
      savingLegacyMapping: false,
      deviceTimeoutSeconds: 300,
      deviceLegacyModuleId: '',
      siteLegacyStationId: '',

      ingestFormatOptions: [
        {
          value: 'AUTO',
          shortTitle: 'AUTO (Hybrid)',
          desc: 'Terima Format Baru & Cloud Lama create_multi otomatis.',
        },
        {
          value: 'STANDARD_V1',
          shortTitle: 'STANDARD_V1',
          desc: 'Hanya terima JSON format baru (values / records).',
        },
        {
          value: 'CBI_LEGACY',
          shortTitle: 'CBI_LEGACY',
          desc: 'Hanya terima format Cloud lama (multi[tick] / hourly_data).',
        },
      ],

      acqModules: {
        sensor_polling_enabled: true,
        mqtt_ingest_enabled: true,
        mqtt_broker_url: '',
        mqtt_topic_prefix: 'datalogger/ingest/+',
      },

      payloadBuilder: {
        loggerType: 'AQMS',
        siteId: 0,
        deviceId: 0,
        mode: 'single',
      },
      payloadFormatTabs: [
        { id: 'single', label: 'JSON Real-Time' },
        { id: 'batch', label: 'JSON Batch' },
        { id: 'cbi_multi_tick', label: 'Cloud Lama (create_multi)' },
        { id: 'curl', label: 'cURL Bearer' },
        { id: 'mqtt', label: 'MQTT' },
      ],

      showTenantModal: false,
      editingTenant: null,
      tenantForm: {
        name: '',
        code: '',
        description: '',
        status: 'ACTIVE',
        max_sites: 0,
        max_devices: 0,
        customer_name: '',
        customer_subtitle: '',
        contact_name: '',
        contact_email: '',
        contact_phone: '',
      },

      selfTenantForm: {
        id: 0,
        name: '',
        code: '',
        description: '',
        max_sites: 0,
        max_devices: 0,
        customer_name: '',
        customer_subtitle: '',
        contact_name: '',
        contact_email: '',
        contact_phone: '',
      },

      showSiteModal: false,
      editingSite: null,
      siteForm: {
        tenant_id: 1,
        name: '',
        code: '',
        monitoring_domain: 'AQMS',
        ingest_format: 'AUTO',
        legacy_station_id: '',
        latitude: 0,
        longitude: 0,
        address: '',
        timezone: 'Asia/Jakarta',
        enabled: true,
      },

      showCreateKeyModal: false,
      newlyCreatedPlainKey: '',
      keyForm: {
        tenant_id: 1,
        site_id: 0,
        name: '',
        ingest_format: 'AUTO',
        custom_token: '',
        auto_provision: false,
      },
    };
  },
  computed: {
    userRole() {
      return ((this.$store.state.user && this.$store.state.user.role) || '').toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    isEngineer() {
      return this.userRole === 'ENGINEER';
    },
    isCloudMode() {
      const sys = (this.$store && this.$store.state && this.$store.state.systemStatus) || {};
      return Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
    },
    activeTenantId() {
      return Number(this.$store.state.activeTenantId || 0);
    },
    activeSiteId() {
      return Number(this.$store.state.activeSiteId || 0);
    },
    activeTenantName() {
      const found = this.tenants.find(t => Number(t.id) === this.activeTenantId);
      return found ? found.name : `Tenant #${this.activeTenantId}`;
    },
    activeSiteName() {
      const found = this.sites.find(s => Number(s.id) === this.activeSiteId);
      return found ? found.name : `Site #${this.activeSiteId}`;
    },
    siteFormPickerLat() {
      const la = Number(this.siteForm.latitude);
      const ln = Number(this.siteForm.longitude);
      if (!Number.isFinite(la) || !Number.isFinite(ln) || (Math.abs(la) < 0.0001 && Math.abs(ln) < 0.0001)) {
        return null;
      }
      return la;
    },
    siteFormPickerLng() {
      const la = Number(this.siteForm.latitude);
      const ln = Number(this.siteForm.longitude);
      if (!Number.isFinite(la) || !Number.isFinite(ln) || (Math.abs(la) < 0.0001 && Math.abs(ln) < 0.0001)) {
        return null;
      }
      return ln;
    },
    filteredTenants() {
      const q = (this.tenantSearch || '').trim().toLowerCase();
      if (!q) return this.tenants;
      return this.tenants.filter(t =>
        (t.name || '').toLowerCase().includes(q) ||
        (t.code || '').toLowerCase().includes(q) ||
        (t.customer_name || '').toLowerCase().includes(q)
      );
    },
    totalSitesAcrossTenants() {
      return this.tenants.reduce((acc, t) => acc + Number(t.site_count || 0), 0);
    },
    totalDevicesAcrossTenants() {
      return this.tenants.reduce((acc, t) => acc + Number(t.device_count || 0), 0);
    },
    filteredSitesForKeyModal() {
      if (!this.isSuperAdmin) return this.sites;
      const tid = Number(this.keyForm.tenant_id || 0);
      if (!tid) return this.sites;
      return this.sites.filter(s => Number(s.tenant_id) === tid);
    },
    centralIngestEndpoint() {
      const origin = typeof window !== 'undefined' && window.location ? window.location.origin : 'https://central-datalogger.example.com';
      return `${origin}/api/v1/ingest/telemetry`;
    },
    centralLegacyCreateMultiEndpoint() {
      const origin = typeof window !== 'undefined' && window.location ? window.location.origin : 'https://central-datalogger.example.com';
      if (this.payloadBuilder.mode === 'cbi_hourly') {
        return `${origin}/api/v1/sparing_logs/create_multi`;
      }
      return `${origin}/api/v1/ticks/create_multi`;
    },
    activeHttpIngestEndpoint() {
      if (this.payloadBuilder.mode === 'cbi_multi_tick' || this.payloadBuilder.mode === 'cbi_hourly') {
        return this.centralLegacyCreateMultiEndpoint;
      }
      return this.centralIngestEndpoint;
    },
    centralMqttIngestEndpoint() {
      const origin = typeof window !== 'undefined' && window.location ? window.location.origin : 'https://central-datalogger.example.com';
      return `${origin}/api/v1/ingest/mqtt`;
    },
    previewBearerToken() {
      return this.newlyCreatedPlainKey || 'dlg_live_YOUR_SECRET_BEARER_TOKEN';
    },
    selectedBuilderSite() {
      return this.sites.find(s => Number(s.id) === Number(this.payloadBuilder.siteId)) || null;
    },
    filteredDevicesForBuilder() {
      if (!this.payloadBuilder.siteId) return this.registeredDevices;
      return this.registeredDevices.filter(d => Number(d.site_id) === Number(this.payloadBuilder.siteId));
    },
    selectedBuilderDevice() {
      return this.filteredDevicesForBuilder.find(d => Number(d.id) === Number(this.payloadBuilder.deviceId)) || null;
    },
    selectedBuilderDeviceParams() {
      if (!this.selectedBuilderDevice || !Array.isArray(this.selectedBuilderDevice.parameters)) return [];
      return this.selectedBuilderDevice.parameters.filter(p => p.enabled !== false && (p.parameter_code || p.code));
    },
    resolvedBuilderSiteCode() {
      if (this.selectedBuilderSite && this.selectedBuilderSite.code) return this.selectedBuilderSite.code;
      const lt = String(this.payloadBuilder.loggerType || 'AQMS').toLowerCase();
      return `stasiun-${lt}-01`;
    },
    resolvedBuilderDeviceCode() {
      if (this.selectedBuilderDevice) {
        return this.selectedBuilderDevice.device_code || this.selectedBuilderDevice.code || 'LOGGER-01';
      }
      return `${this.payloadBuilder.loggerType}-LOGGER-01`;
    },
    effectivePayloadFormatTabs() {
      if (String(this.payloadBuilder.loggerType || '').toUpperCase() === 'SPARING') {
        return [
          { id: 'single', label: 'Per 2m' },
          { id: 'batch', label: '30×2m Standar' },
          { id: 'cbi_multi_tick', label: 'multi[tick] Lama' },
          { id: 'cbi_hourly', label: '30×2m CBI' },
          { id: 'klhk_hourly', label: '30×2m KLHK' },
          { id: 'curl', label: 'cURL' },
          { id: 'mqtt', label: 'MQTT' },
        ];
      }
      return this.payloadFormatTabs;
    },
    generatedPayloadObject() {
      const lt = String(this.payloadBuilder.loggerType || 'AQMS').toUpperCase();
      const siteCode = this.resolvedBuilderSiteCode;
      const deviceCode = this.resolvedBuilderDeviceCode;

      const sampleForCode = (rawCode) => {
        const k = String(rawCode || '').toLowerCase().replace(/[^a-z0-9]/g, '');
        const map = {
          pm25: 34.5, pm10: 62.1, so2: 12.4, co: 1150.0, o3: 41.2, no2: 18.9, hc: 8.4,
          ws: 2.8, wd: 185.0, temp: 28.4, temperature: 28.4, temp_panel: 28.1, temppanel: 28.1, humidity: 71.2, rh: 71.2, pressure: 1009.5, sr: 540.0, rain: 0.0,
          ph: 7.03, cod: 85.54, bod: 16.4, tss: 19.21, nh3n: 0.42, nh4: 0.48, no3: 2.85,
          do: 6.45, tds: 312.0, turbidity: 14.8, debit: 0.23, flow: 0.23, waterlevel: 1.42, level: 1.42,
        };
        return map[k] !== undefined ? map[k] : 15.25;
      };

      let values1 = {};
      let values2 = {};
      if (this.selectedBuilderDeviceParams.length > 0) {
        this.selectedBuilderDeviceParams.forEach((p) => {
          const code = p.parameter_code || p.code;
          const v = sampleForCode(code);
          values1[code] = v;
          values2[code] = Number((v * 1.03).toFixed(2));
        });
      } else if (lt === 'SPARING') {
        values1 = { ph: 7.03, cod: 85.54, tss: 19.21, nh3n: 0.42, debit: 0.23 };
        values2 = { ph: 7.05, cod: 86.12, tss: 19.48, nh3n: 0.44, debit: 0.23 };
      } else if (lt === 'ONLIMO') {
        values1 = { ph: 7.05, do: 6.45, bod: 14.2, cod: 32.6, tss: 24.0, nh3n: 0.31, no3: 2.85, tds: 312.0, turbidity: 14.8, temperature: 27.6, water_level: 1.42 };
        values2 = { ph: 7.09, do: 6.52, bod: 14.6, cod: 33.4, tss: 24.8, nh3n: 0.33, no3: 2.91, tds: 315.0, turbidity: 15.1, temperature: 27.8, water_level: 1.44 };
      } else {
        values1 = { pm25: 34.5, pm10: 62.1, so2: 12.4, co: 1150.0, o3: 41.2, no2: 18.9, hc: 8.4, ws: 2.8, wd: 185.0, temperature: 28.4, humidity: 71.2, pressure: 1009.5, sr: 540.0, rain: 0.0 };
        values2 = { pm25: 35.2, pm10: 63.4, so2: 12.8, co: 1165.0, o3: 42.0, no2: 19.3, hc: 8.6, ws: 3.0, wd: 188.0, temperature: 28.5, humidity: 70.8, pressure: 1009.4, sr: 548.0, rain: 0.0 };
      }

      if (lt === 'SPARING' && ['batch', 'cbi_hourly', 'klhk_hourly'].includes(this.payloadBuilder.mode)) {
        const baseUnix = 1794445200; // 2026-11-12 08:00:00 +0700
        const pad2 = (n) => String(n).padStart(2, '0');

        if (this.payloadBuilder.mode === 'cbi_hourly') {
          const hourlyData = [];
          for (let i = 0; i < 30; i += 1) {
            const mm = pad2(i * 2);
            const tsStr = `2026-11-12 08:${mm}:00 +0700`;
            const factor = 1 + ((i % 5) * 0.004);
            const row = {};
            Object.keys(values1).forEach((k) => {
              row[k] = Number((values1[k] * factor).toFixed(2));
            });
            row.created_at = tsStr;
            row.updated_at = tsStr;
            hourlyData.push(row);
          }
          return {
            module_id: String((this.selectedBuilderDevice && (this.selectedBuilderDevice.legacy_module_id || this.selectedBuilderDevice.id)) || '4'),
            station_id: String((this.selectedBuilderSite && (this.selectedBuilderSite.legacy_station_id || this.selectedBuilderSite.id)) || '3'),
            hourly_data: hourlyData,
          };
        }

        if (this.payloadBuilder.mode === 'klhk_hourly') {
          const dataRows = [];
          for (let i = 0; i < 30; i += 1) {
            const factor = 1 + ((i % 5) * 0.004);
            const row = { datetime: baseUnix + (i * 120) };
            Object.keys(values1).forEach((k) => {
              row[k] = Number((values1[k] * factor).toFixed(2));
            });
            dataRows.push(row);
          }
          return {
            uid: siteCode,
            data: dataRows,
          };
        }

        const stdRecords = [];
        for (let i = 0; i < 30; i += 1) {
          const mm = pad2(i * 2);
          const factor = 1 + ((i % 5) * 0.004);
          const slotVals = {};
          Object.keys(values1).forEach((k) => {
            slotVals[k] = Number((values1[k] * factor).toFixed(2));
          });
          stdRecords.push({
            timestamp: `2026-11-12T08:${mm}:00+07:00`,
            values: slotVals,
          });
        }
        return {
          logger_type: lt,
          site_code: siteCode,
          device_code: deviceCode,
          records: stdRecords,
        };
      }

      const nowIso = '2026-11-13T10:08:38+07:00';
      const prevIso = '2026-11-13T10:03:38+07:00';

      if (this.payloadBuilder.mode === 'cbi_multi_tick') {
        const rawModId = (this.selectedBuilderDevice && (this.selectedBuilderDevice.legacy_module_id || this.selectedBuilderDevice.id)) || 199;
        const rawStaId = (this.selectedBuilderSite && (this.selectedBuilderSite.legacy_station_id || this.selectedBuilderSite.id)) || 202;
        const parsedModId = /^\d+$/.test(String(rawModId)) ? Number(rawModId) : rawModId;
        const parsedStaId = /^\d+$/.test(String(rawStaId)) ? Number(rawStaId) : rawStaId;

        const toLegacySensorName = (code) => {
          const norm = String(code || '').trim().toLowerCase();
          const legacyMap = {
            'pm2.5': 'pm25',
            'pm_25': 'pm25',
            temperature: 'temp',
            'air-pressure': 'pressure',
            air_pressure: 'pressure',
            'w-speed': 'wspeed',
            ws: 'wspeed',
            'solar-radiation': 'intensity',
            sr: 'intensity',
            'rain-fall': 'tprecipitation',
            rainfall: 'tprecipitation',
            rain: 'tprecipitation',
            'wind-angle': 'wind_angle',
            wd: 'wind_angle',
          };
          return legacyMap[norm] || code;
        };

        let tickEntries = [];
        if (this.selectedBuilderDeviceParams.length > 0) {
          tickEntries = Object.keys(values1).map((k) => ({
            sensor_name: toLegacySensorName(k),
            value: values1[k],
          }));
        } else if (lt === 'AQMS') {
          tickEntries = [
            { sensor_name: 'no2', value: 14.236 },
            { sensor_name: 'so2', value: 5.673 },
            { sensor_name: 'o3', value: 70.782 },
            { sensor_name: 'co', value: 586.191 },
            { sensor_name: 'pm25', value: 34.811 },
            { sensor_name: 'pm10', value: 66.176 },
            { sensor_name: 'temp', value: 32.941 },
            { sensor_name: 'humidity', value: 63.874 },
            { sensor_name: 'pressure', value: 1007.5 },
            { sensor_name: 'wspeed', value: 3.543 },
            { sensor_name: 'intensity', value: 813.654 },
            { sensor_name: 'tprecipitation', value: 0 },
            { sensor_name: 'hc', value: 4.686 },
            { sensor_name: 'wind_angle', value: 237.4 },
            { sensor_name: 'rain_intensity', value: 0 },
          ];
        } else {
          tickEntries = Object.keys(values1).map((k) => ({
            sensor_name: k,
            value: values1[k],
          }));
        }

        return {
          multi: tickEntries.map((entry) => ({
            tick: {
              module_id: parsedModId,
              station_id: parsedStaId,
              sensor_name: entry.sensor_name,
              value: entry.value,
              created_at: nowIso,
              updated_at: nowIso,
            },
          })),
        };
      }

      if (this.payloadBuilder.mode === 'batch') {
        return {
          logger_type: lt,
          site_code: siteCode,
          device_code: deviceCode,
          records: [
            { timestamp: prevIso, values: values1 },
            { timestamp: nowIso, values: values2 },
          ],
        };
      }

      return {
        logger_type: lt,
        site_code: siteCode,
        device_code: deviceCode,
        timestamp: nowIso,
        values: values1,
      };
    },
    generatedActiveCurlCommand() {
      const jsonPretty = JSON.stringify(this.generatedPayloadObject, null, 2);
      return `curl -X POST "${this.activeHttpIngestEndpoint}" \\\n  -H "Authorization: Bearer ${this.previewBearerToken}" \\\n  -H "Content-Type: application/json" \\\n  -d '${jsonPretty}'`;
    },
    generatedPayloadSnippet() {
      const jsonPretty = JSON.stringify(this.generatedPayloadObject, null, 2);
      if (this.payloadBuilder.mode === 'curl') {
        return `curl -X POST "${this.centralIngestEndpoint}" \\\n  -H "Authorization: Bearer ${this.previewBearerToken}" \\\n  -H "Content-Type: application/json" \\\n  -d '${jsonPretty}'`;
      }
      if (this.payloadBuilder.mode === 'mqtt') {
        const topic = `datalogger/ingest/${this.resolvedBuilderSiteCode}/${this.resolvedBuilderDeviceCode}`;
        const webhookBody = JSON.stringify({
          topic,
          payload: this.generatedPayloadObject,
        }, null, 2);
        return `# MQTT Broker Topic: ${topic}\n# Webhook / Bridge Forwarding ke Server:\ncurl -X POST "${this.centralMqttIngestEndpoint}" \\\n  -H "Authorization: Bearer ${this.previewBearerToken}" \\\n  -H "Content-Type: application/json" \\\n  -d '${webhookBody}'`;
      }
      return jsonPretty;
    },
  },
  mounted() {
    if (this.isEngineer && !this.isCloudMode) {
      this.$router.replace({ path: '/access-denied', query: { from: this.$route.fullPath } });
      return;
    }
    if (this.$route && this.$route.query && this.$route.query.tab) {
      const t = String(this.$route.query.tab);
      if (['tenants', 'sites', 'ingest'].includes(t)) {
        this.activeTab = t;
      }
    }
    this.loadAll();
  },
  methods: {
    showToast(msg, type = 'success') {
      this.toastMessage = msg;
      this.toastType = type;
      setTimeout(() => {
        if (this.toastMessage === msg) this.toastMessage = '';
      }, 5000);
    },
    async loadAll() {
      await Promise.all([
        this.loadTenantsOrProfile(),
        this.loadSites(),
        this.loadIngestKeys(),
        this.loadRegisteredDevices(),
        this.loadAcquisitionModules(),
      ]);
      this.$store.dispatch('fetchTenantContext');
    },
    async loadRegisteredDevices() {
      try {
        const res = await axios.get('/api/devices', { params: { page_size: 200 } });
        const raw = res.data && res.data.data;
        if (Array.isArray(raw)) {
          this.registeredDevices = raw;
        } else if (raw && Array.isArray(raw.items)) {
          this.registeredDevices = raw.items;
        }
        if (!this.payloadBuilder.deviceId && this.registeredDevices.length > 0) {
          this.payloadBuilder.deviceId = this.registeredDevices[0].id;
          if (!this.payloadBuilder.siteId && this.registeredDevices[0].site_id) {
            this.payloadBuilder.siteId = this.registeredDevices[0].site_id;
          }
        }
        this.syncSelectedBuilderDeviceTimeout();
      } catch (_) {
        this.registeredDevices = [];
      }
    },
    syncSelectedBuilderDeviceTimeout() {
      const site = this.selectedBuilderSite;
      this.siteLegacyStationId = (site && site.legacy_station_id) || '';
      const dev = this.selectedBuilderDevice;
      if (!dev) {
        this.deviceTimeoutSeconds = 300;
        this.deviceLegacyModuleId = '';
        return;
      }
      this.deviceLegacyModuleId = dev.legacy_module_id || '';
      const connTimeout = Number(dev.connection && (dev.connection.timeout || dev.connection.timeout_ms)) || 0;
      if (connTimeout >= 10 && connTimeout !== 1000) {
        this.deviceTimeoutSeconds = connTimeout;
        return;
      }
      const params = Array.isArray(dev.parameters) ? dev.parameters : [];
      const maxParamStale = params.reduce((mx, p) => Math.max(mx, Number(p.stale_timeout_seconds) || 0), 0);
      this.deviceTimeoutSeconds = maxParamStale >= 180 ? maxParamStale : 300;
    },
    async saveBuilderLegacyMapping() {
      const dev = this.selectedBuilderDevice;
      const site = this.selectedBuilderSite;
      if (!dev && !site) return;
      this.savingLegacyMapping = true;
      try {
        if (site && site.id) {
          await axios.put(`/api/sites/${site.id}`, {
            legacy_station_id: (this.siteLegacyStationId || '').trim(),
          });
        }
        if (dev && dev.id) {
          await axios.put(`/api/devices/${dev.id}`, {
            legacy_module_id: (this.deviceLegacyModuleId || '').trim(),
          });
        }
        await Promise.all([
          this.loadSites(),
          this.loadRegisteredDevices(),
          this.$store.dispatch('fetchDevices'),
        ]);
        this.showToast('Mapping ID Cloud lama (station_id & module_id) berhasil disimpan.');
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      } finally {
        this.savingLegacyMapping = false;
      }
    },
    async saveBuilderDeviceTimeout() {
      const dev = this.selectedBuilderDevice;
      if (!dev || !dev.id) return;
      const sec = Math.max(30, Number(this.deviceTimeoutSeconds) || 300);
      this.deviceTimeoutSeconds = sec;
      this.savingDeviceTimeout = true;
      try {
        const conn = dev.connection || {};
        const proto = conn.protocol || dev.device_type || 'EDGE_PUSH';
        await axios.put(`/api/devices/${dev.id}/connection`, {
          protocol: proto,
          connection_type: conn.connection_type || (proto === 'MQTT' ? 'MQTT_BROKER' : 'HTTPS_PUSH'),
          host: conn.host || '',
          port: conn.port || (proto === 'MQTT' ? 1883 : 443),
          timeout: sec,
          retry_count: conn.retry_count || 3,
          polling_interval: conn.polling_interval || 1000,
          enabled: conn.enabled !== false,
        });
        await this.loadRegisteredDevices();
        await Promise.all([
          this.$store.dispatch('fetchDevices'),
          this.$store.dispatch('fetchActiveAlarms'),
        ]);
        const devCode = dev.device_code || dev.code;
        this.showToast(`Timeout penerimaan data untuk ${devCode} berhasil diatur ke ${sec} detik (${Math.round(sec / 60 * 10) / 10} menit).`);
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      } finally {
        this.savingDeviceTimeout = false;
      }
    },
    async loadAcquisitionModules() {
      if (!this.isSuperAdmin) return;
      try {
        const res = await axios.get('/api/system/config');
        const d = res.data && res.data.data;
        if (d) {
          this.acqModules = {
            cloud_mode: Boolean(d.cloud_mode || String(d.monitoring_domain || '').toUpperCase() === 'CLOUD'),
            sensor_polling_enabled: !d.cloud_mode && d.sensor_polling_enabled !== false,
            mqtt_ingest_enabled: d.mqtt_ingest_enabled !== false,
            mqtt_broker_url: d.mqtt_broker_url || '',
            mqtt_topic_prefix: d.mqtt_topic_prefix || 'datalogger/ingest/+',
          };
        }
      } catch (_) {
        // non-blocking
      }
    },
    async toggleAcquisitionModule(field) {
      if (!this.isSuperAdmin) return;
      this.savingModules = true;
      const nextVal = !this.acqModules[field];
      try {
        await axios.put('/api/system/config', { [field]: nextVal });
        this.acqModules[field] = nextVal;
        const label = field === 'sensor_polling_enabled' ? 'Modul Pembacaan Sensor Fisik' : 'Modul Penerimaan MQTT';
        this.showToast(`${label} berhasil diubah menjadi ${nextVal ? 'AKTIF' : 'NONAKTIF'}.`);
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      } finally {
        this.savingModules = false;
      }
    },
    selectPayloadLoggerType(lt) {
      this.payloadBuilder.loggerType = lt;
      if (lt !== 'SPARING' && ['cbi_hourly', 'klhk_hourly'].includes(this.payloadBuilder.mode)) {
        this.payloadBuilder.mode = 'batch';
      }
      const matchingSite = this.sites.find(s => String(s.monitoring_domain || '').toUpperCase() === lt);
      if (matchingSite) {
        this.payloadBuilder.siteId = matchingSite.id;
        this.onPayloadSiteChange();
      } else {
        this.payloadBuilder.siteId = 0;
        this.payloadBuilder.deviceId = 0;
      }
    },
    onPayloadSiteChange() {
      const site = this.selectedBuilderSite;
      if (site && site.monitoring_domain) {
        const dom = String(site.monitoring_domain).toUpperCase();
        if (['AQMS', 'SPARING', 'ONLIMO'].includes(dom)) {
          this.payloadBuilder.loggerType = dom;
        } else if (dom === 'WQMS') {
          this.payloadBuilder.loggerType = 'ONLIMO';
        }
      }
      if (this.payloadBuilder.loggerType !== 'SPARING' && ['cbi_hourly', 'klhk_hourly'].includes(this.payloadBuilder.mode)) {
        this.payloadBuilder.mode = 'batch';
      }
      const devs = this.filteredDevicesForBuilder;
      this.payloadBuilder.deviceId = devs.length > 0 ? devs[0].id : 0;
      this.syncSelectedBuilderDeviceTimeout();
    },
    async loadTenantsOrProfile() {
      try {
        if (this.isSuperAdmin) {
          const res = await axios.get('/api/tenants');
          if (res.data && res.data.data) {
            this.tenants = Array.isArray(res.data.data) ? res.data.data : [];
          }
        } else {
          const res = await axios.get('/api/tenants/current');
          if (res.data && res.data.data) {
            const t = res.data.data;
            this.selfTenantForm = {
              id: t.id,
              name: t.name || '',
              code: t.code || '',
              description: t.description || '',
              max_sites: t.max_sites || 0,
              max_devices: t.max_devices || 0,
              customer_name: t.customer_name || '',
              customer_subtitle: t.customer_subtitle || '',
              contact_name: t.contact_name || '',
              contact_email: t.contact_email || '',
              contact_phone: t.contact_phone || '',
            };
          }
        }
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      }
    },
    async loadSites() {
      try {
        const res = await axios.get('/api/sites');
        if (res.data && res.data.data) {
          this.sites = Array.isArray(res.data.data) ? res.data.data : [];
        }
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      }
    },
    async loadIngestKeys() {
      try {
        const res = await axios.get('/api/ingest-keys');
        if (res.data && res.data.data) {
          this.ingestKeys = Array.isArray(res.data.data) ? res.data.data : [];
        }
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      }
    },
    tenantNameById(id) {
      const found = this.tenants.find(t => Number(t.id) === Number(id));
      return found ? found.name : `Tenant #${id}`;
    },
    siteNameById(id) {
      const found = this.sites.find(s => Number(s.id) === Number(id));
      return found ? `${found.name} (${found.code})` : `Site #${id}`;
    },
    domainBadgeClass(domain) {
      const d = String(domain || 'AQMS').toUpperCase();
      if (d === 'SPARING') return 'bg-amber-500/10 text-amber-700 dark:text-amber-400 border-amber-500/30';
      if (d === 'ONLIMO') return 'bg-blue-500/10 text-blue-700 dark:text-blue-400 border-blue-500/30';
      if (d === 'WQMS') return 'bg-teal-500/10 text-teal-700 dark:text-teal-400 border-teal-500/30';
      return 'bg-cyan-500/10 text-cyan-700 dark:text-cyan-400 border-cyan-500/30';
    },
    normalizeIngestFormat(fmt) {
      const v = String(fmt || '').trim().toUpperCase();
      if (v === 'STANDARD_V1' || v === 'CBI_LEGACY') return v;
      return 'AUTO';
    },
    ingestFormatShortLabel(fmt) {
      const v = this.normalizeIngestFormat(fmt);
      if (v === 'CBI_LEGACY') return 'create_multi';
      if (v === 'STANDARD_V1') return 'STANDARD_V1';
      return 'AUTO';
    },
    ingestFormatHint(fmt) {
      const v = this.normalizeIngestFormat(fmt);
      if (v === 'CBI_LEGACY') return 'Hanya menerima format Cloud lama (multi[tick] / hourly_data)';
      if (v === 'STANDARD_V1') return 'Hanya menerima format standar baru (values / records)';
      return 'Menerima format standar baru dan Cloud lama create_multi sekaligus';
    },
    ingestFormatBadgeClass(fmt) {
      const v = this.normalizeIngestFormat(fmt);
      if (v === 'CBI_LEGACY') return 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/35';
      if (v === 'STANDARD_V1') return 'bg-sky-500/10 text-sky-700 dark:text-sky-300 border-sky-500/35';
      return 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/35';
    },
    formatDateTime(val) {
      if (!val) return '-';
      return new Date(val).toLocaleString('en-GB', { hour12: false });
    },
    async enterTenantWorkspace(tenant) {
      await this.$store.dispatch('setActiveTenant', tenant.id);
      this.showToast(`Workspace aktif dialihkan ke '${tenant.name}'.`);
    },
    async resetToGlobalScope() {
      await this.$store.dispatch('setActiveTenant', 0);
      this.showToast('Kembali ke mode Global Semua Tenant.');
    },
    async focusSite(site) {
      if (this.isSuperAdmin && Number(this.activeTenantId) !== Number(site.tenant_id)) {
        await this.$store.dispatch('setActiveTenant', site.tenant_id);
      }
      await this.$store.dispatch('setActiveSite', site.id);
      this.showToast(`Stasiun aktif difokuskan ke '${site.name}'.`);
    },
    openCreateTenantModal() {
      this.editingTenant = null;
      this.tenantForm = {
        name: '',
        code: '',
        description: '',
        status: 'ACTIVE',
        max_sites: 0,
        max_devices: 0,
        customer_name: '',
        customer_subtitle: '',
        contact_name: '',
        contact_email: '',
        contact_phone: '',
      };
      this.showTenantModal = true;
    },
    openEditTenantModal(t) {
      this.editingTenant = t;
      this.tenantForm = {
        name: t.name || '',
        code: t.code || '',
        description: t.description || '',
        status: t.status || 'ACTIVE',
        max_sites: t.max_sites || 0,
        max_devices: t.max_devices || 0,
        customer_name: t.customer_name || '',
        customer_subtitle: t.customer_subtitle || '',
        contact_name: t.contact_name || '',
        contact_email: t.contact_email || '',
        contact_phone: t.contact_phone || '',
      };
      this.showTenantModal = true;
    },
    async submitTenantModal() {
      this.savingTenant = true;
      try {
        if (this.editingTenant) {
          await axios.put(`/api/tenants/${this.editingTenant.id}`, this.tenantForm);
          this.showToast(`Tenant '${this.tenantForm.name}' berhasil diperbarui.`);
        } else {
          await axios.post('/api/tenants', this.tenantForm);
          this.showToast(`Tenant '${this.tenantForm.name}' berhasil dibuat beserta stasiun & kategori default.`);
        }
        this.showTenantModal = false;
        await this.loadAll();
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      } finally {
        this.savingTenant = false;
      }
    },
    async saveSelfTenantProfile() {
      this.savingTenant = true;
      try {
        await axios.put('/api/tenants/current', this.selfTenantForm);
        this.showToast('Profil organisasi & branding tenant berhasil disimpan.');
        await this.loadAll();
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      } finally {
        this.savingTenant = false;
      }
    },
    async confirmDeleteTenant(t) {
      if (!confirm(`Hapus organisasi tenant '${t.name}'? Pastikan seluruh perangkat di tenant ini sudah dikosongkan terlebih dahulu.`)) {
        return;
      }
      try {
        await axios.delete(`/api/tenants/${t.id}`);
        this.showToast(`Tenant '${t.name}' berhasil dihapus.`);
        if (Number(this.activeTenantId) === Number(t.id)) {
          await this.$store.dispatch('setActiveTenant', 0);
        }
        await this.loadAll();
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      }
    },
    openCreateSiteModal() {
      this.editingSite = null;
      const defaultTid = this.activeTenantId || (this.tenants[0] && this.tenants[0].id) || 1;
      this.siteForm = {
        tenant_id: defaultTid,
        name: '',
        code: '',
        monitoring_domain: 'AQMS',
        ingest_format: 'AUTO',
        legacy_station_id: '',
        latitude: 0,
        longitude: 0,
        address: '',
        timezone: 'Asia/Jakarta',
        enabled: true,
      };
      this.showSiteModal = true;
    },
    openEditSiteModal(s) {
      this.editingSite = s;
      this.siteForm = {
        tenant_id: s.tenant_id,
        name: s.name || '',
        code: s.code || '',
        monitoring_domain: s.monitoring_domain || 'AQMS',
        ingest_format: this.normalizeIngestFormat(s.ingest_format),
        legacy_station_id: s.legacy_station_id || '',
        latitude: s.latitude || 0,
        longitude: s.longitude || 0,
        address: s.address || '',
        timezone: s.timezone || 'Asia/Jakarta',
        enabled: s.enabled !== false,
      };
      this.showSiteModal = true;
    },
    async onSiteMapCoordinatePicked(coords) {
      if (!coords) return;
      const rawLat = coords.lat !== undefined ? coords.lat : coords.latitude;
      const rawLng = coords.lng !== undefined ? coords.lng : coords.longitude;
      if (rawLat !== null && rawLat !== undefined && Number.isFinite(Number(rawLat))) {
        this.siteForm.latitude = Number(Number(rawLat).toFixed(6));
      }
      if (rawLng !== null && rawLng !== undefined && Number.isFinite(Number(rawLng))) {
        this.siteForm.longitude = Number(Number(rawLng).toFixed(6));
      }
      if (coords.source === 'MAP' && Number.isFinite(Number(this.siteForm.latitude)) && Number.isFinite(Number(this.siteForm.longitude))) {
        try {
          const res = await axios.get('/api/system/location/reverse', {
            params: { lat: this.siteForm.latitude, lng: this.siteForm.longitude },
          });
          const geo = res && res.data && res.data.data;
          const revAddr = (geo && (geo.display_name || geo.place_name)) || '';
          if (revAddr) {
            this.siteForm.address = revAddr;
          }
        } catch (_) {
          // ignore reverse geocode network errors
        }
      }
    },
    onSiteLocationSelected(loc) {
      if (!loc) return;
      this.onSiteMapCoordinatePicked(loc);
      const resolvedAddress = loc.displayName || loc.display_name || loc.name || '';
      if (resolvedAddress) {
        this.siteForm.address = resolvedAddress;
      }
    },
    async submitSiteModal() {
      this.savingSite = true;
      try {
        if (this.editingSite) {
          await axios.put(`/api/sites/${this.editingSite.id}`, this.siteForm);
          this.showToast(`Stasiun '${this.siteForm.name}' berhasil diperbarui.`);
        } else {
          await axios.post('/api/sites', this.siteForm);
          this.showToast(`Stasiun '${this.siteForm.name}' berhasil ditambahkan.`);
        }
        this.showSiteModal = false;
        await this.loadAll();
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      } finally {
        this.savingSite = false;
      }
    },
    async confirmDeleteSite(s) {
      if (!confirm(`Hapus stasiun '${s.name}'? Stasiun yang masih memiliki perangkat aktif tidak dapat dihapus.`)) {
        return;
      }
      try {
        await axios.delete(`/api/sites/${s.id}`);
        this.showToast(`Stasiun '${s.name}' berhasil dihapus.`);
        if (Number(this.activeSiteId) === Number(s.id)) {
          await this.$store.dispatch('setActiveSite', 0);
        }
        await this.loadAll();
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      }
    },
    openCreateKeyModal() {
      const defaultTid = this.activeTenantId || (this.tenants[0] && this.tenants[0].id) || 1;
      const currentSite = this.sites.find(s => Number(s.id) === Number(this.activeSiteId));
      this.keyForm = {
        tenant_id: defaultTid,
        site_id: this.activeSiteId || 0,
        name: '',
        ingest_format: currentSite ? this.normalizeIngestFormat(currentSite.ingest_format) : 'AUTO',
        custom_token: '',
        auto_provision: false,
      };
      this.showCreateKeyModal = true;
    },
    async submitCreateKey() {
      this.savingKey = true;
      try {
        const payload = {
          tenant_id: this.keyForm.tenant_id,
          name: this.keyForm.name,
          ingest_format: this.normalizeIngestFormat(this.keyForm.ingest_format),
          custom_token: (this.keyForm.custom_token || '').trim(),
          auto_provision: false,
          site_id: this.keyForm.site_id > 0 ? this.keyForm.site_id : null,
        };
        const res = await axios.post('/api/ingest-keys', payload);
        if (res.data && res.data.data) {
          this.newlyCreatedPlainKey = res.data.data.plain_key || '';
          this.showCreateKeyModal = false;
          this.showToast('Bearer Token berhasil disimpan! Salin token sebelum menutup banner.');
          await this.loadIngestKeys();
        }
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      } finally {
        this.savingKey = false;
      }
    },
    async updateIngestKeyFormat(k, nextFormat) {
      const normalized = this.normalizeIngestFormat(nextFormat);
      try {
        await axios.put(`/api/ingest-keys/${k.id}`, { ingest_format: normalized });
        await this.loadIngestKeys();
        this.showToast(`Format penerimaan untuk token '${k.name}' dialihkan ke ${normalized}.`);
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
        await this.loadIngestKeys();
      }
    },
    async toggleIngestKey(k) {
      try {
        await axios.put(`/api/ingest-keys/${k.id}/enable`, { enabled: !k.enabled });
        await this.loadIngestKeys();
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      }
    },
    async revokeIngestKey(k) {
      if (!confirm(`Cabut dan hapus permanen kunci '${k.name}' (${k.key_prefix}...)? Unit Edge yang memakai kunci ini tidak dapat push lagi.`)) {
        return;
      }
      try {
        await axios.delete(`/api/ingest-keys/${k.id}`);
        this.showToast(`Kunci '${k.name}' telah dicabut.`);
        await this.loadIngestKeys();
      } catch (err) {
        this.showToast((err.response && err.response.data && err.response.data.error) || err.message, 'error');
      }
    },
    async copyPlainKey(val) {
      try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
          await navigator.clipboard.writeText(val);
          this.showToast('API Key berhasil disalin ke clipboard!');
        }
      } catch (e) {
        this.showToast('Salin manual dari kotak teks di atas.', 'error');
      }
    },
  },
};
</script>

