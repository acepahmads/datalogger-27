<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans flex items-center gap-2">
          {{ $t('config.title') || 'System Configuration' }}
          <span class="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-3xs font-mono font-bold"
                :class="currentDomainBadgeClass">
            ● {{ form.monitoring_domain || 'AQMS' }}
          </span>
          <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
            #83
          </span>
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          {{ $t('config.subtitle') || 'Operational runtime settings, application monitoring domain, and edge connectivity' }}
        </p>
      </div>

      <!-- Tab Switcher & First-Run Setup Button -->
      <div class="flex flex-wrap items-center gap-2">
        <button @click="openSetupWizard"
                class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-cyan-400 border border-cyan-500/30 text-xs font-semibold flex items-center space-x-1.5 transition">
          <span>🚀</span>
          <span>{{ $t('config.firstRunSetup') || 'First-Run Setup' }}</span>
        </button>

        <div class="flex items-center space-x-1 bg-[#111827] p-1 rounded-lg border border-slate-800 text-xs font-sans">
          <button @click="activeTab = 'domain'"
                  class="px-3 py-1.5 rounded-md transition-colors flex items-center space-x-1.5"
                  :class="activeTab === 'domain' ? 'bg-cyan-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
            <span>🌍</span>
            <span>{{ $t('config.tabDomain') || 'Monitoring Domain' }}</span>
          </button>
          <button @click="activeTab = 'branding'"
                  id="tab-btn-branding"
                  class="px-3 py-1.5 rounded-md transition-colors flex items-center space-x-1.5"
                  :class="activeTab === 'branding' ? 'bg-cyan-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
            <span>🏢</span>
            <span>{{ $t('config.tabBranding') || 'Organization & Branding' }}</span>
          </button>
          <button @click="activeTab = 'database'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'database' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
            {{ $t('config.tabDatabase') || 'Database' }}
          </button>
          <button @click="activeTab = 'serial'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'serial' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
            Serial & Comm
          </button>
          <button @click="activeTab = 'general'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'general' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
            {{ $t('config.tabGeneral') || 'Server & Runtime' }}
          </button>
          <button @click="activeTab = 'localfirst'"
                  class="px-3 py-1.5 rounded-md transition-colors"
                  :class="activeTab === 'localfirst' ? 'bg-blue-600 text-white font-semibold shadow-sm' : 'text-slate-400 hover:text-white'">
            {{ $t('config.tabLocalFirst') || 'Local-First' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Alert / Feedback Banner -->
    <div v-if="saveResult" class="p-4 rounded-xl border text-xs font-sans space-y-1"
         :class="saveResult.restart_required ? 'bg-amber-500/10 border-amber-500/30 text-amber-300' : 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300'">
      <div class="font-bold flex items-center gap-2">
        <span v-if="saveResult.restart_required">⚠️ Restart Required</span>
        <span v-else>✓ Configuration Saved</span>
      </div>
      <div>{{ saveResult.message }}</div>
      <div v-if="saveResult.restart_required_fields && saveResult.restart_required_fields.length > 0" class="text-3xs text-amber-400 font-mono">
        Modified fields requiring restart: {{ saveResult.restart_required_fields.join(', ') }}
      </div>
    </div>

    <!-- TAB 0: Monitoring Domain (#83 Single Monitoring Domain Configuration) -->
    <div v-if="activeTab === 'domain'" class="space-y-5 max-w-5xl font-sans">
      <!-- Active Domain Overview Card -->
      <div class="saas-card p-5 space-y-4 bg-gradient-to-br from-[#111827] to-[#131E35] border border-cyan-500/30">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800/80 pb-3">
          <div class="flex items-center space-x-3">
            <span class="p-2.5 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 text-lg">
              🌍
            </span>
            <div>
              <div class="flex items-center space-x-2">
                <h2 class="text-sm font-bold text-white tracking-tight">{{ $t('config.domain.title') || 'Application Monitoring Domain' }}</h2>
                <span class="text-3xs font-mono font-bold px-2 py-0.5 rounded-full" :class="currentDomainBadgeClass">
                  ● ACTIVE: {{ form.monitoring_domain || 'AQMS' }}
                </span>
              </div>
              <p class="text-2xs text-slate-400 mt-0.5">
                {{ $t('config.domain.subtitle') || 'Select the primary operational monitoring domain for this datalogger deployment.' }}
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <span class="text-3xs font-mono px-2.5 py-1 rounded-lg bg-emerald-500/10 text-emerald-300 border border-emerald-500/30 font-semibold flex items-center gap-1">
              <span>🛡️</span>
              <span>Zero Data Deletion Guarantee</span>
            </span>
          </div>
        </div>

        <p class="text-xs text-slate-300 leading-relaxed">
          A single application installation represents one monitoring deployment. The selected domain governs default category presentation on the Operator Command Center while strictly preserving all device mappings, historical telemetry, alarms, and analytical records.
        </p>
      </div>

      <!-- Domain Selection Cards Grid -->
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <!-- 1. AQMS Card -->
        <div class="p-4 rounded-xl border transition-all flex flex-col justify-between space-y-4"
             :class="form.monitoring_domain === 'AQMS'
               ? 'bg-cyan-950/20 border-cyan-500/50 shadow-lg shadow-cyan-950/30 ring-1 ring-cyan-500/30'
               : 'bg-[#111827] border-slate-800 hover:border-slate-700'">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-2xl p-2 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400">💨</span>
              <span v-if="form.monitoring_domain === 'AQMS'" class="text-4xs font-mono font-bold px-2 py-0.5 rounded-full bg-cyan-500/20 text-cyan-300 border border-cyan-500/40">
                ACTIVE DOMAIN
              </span>
              <span v-else class="text-4xs font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700">
                READY
              </span>
            </div>

            <div>
              <h3 class="text-sm font-bold text-white font-sans flex items-center gap-1.5">
                AQMS
                <span class="text-2xs text-slate-400 font-normal">Air Quality</span>
              </h3>
              <p class="text-3xs text-slate-400 mt-1 leading-normal">
                Ambient air quality, toxic gases, particulate matter (PM2.5, PM10), and meteorological stations.
              </p>
            </div>

            <div class="space-y-1.5 pt-2 border-t border-slate-800/80">
              <span class="text-4xs uppercase tracking-wider text-slate-400 font-bold block">Relevant Categories:</span>
              <div class="flex flex-wrap gap-1 text-3xs font-mono">
                <span class="px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">Gas & Pollutants</span>
                <span class="px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-300 border border-amber-500/20">Particulate Matter</span>
                <span class="px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-300 border border-blue-500/20">Meteorology</span>
              </div>
            </div>
          </div>

          <div class="pt-2">
            <button v-if="form.monitoring_domain === 'AQMS'" disabled
                    class="w-full py-1.5 px-3 rounded-lg bg-cyan-500/20 text-cyan-300 border border-cyan-500/30 text-xs font-semibold cursor-default text-center">
              ✓ Active Domain
            </button>
            <button v-else @click="initiateDomainSwitch('AQMS')"
                    class="w-full py-1.5 px-3 rounded-lg bg-slate-800 hover:bg-cyan-600 hover:text-white text-slate-200 border border-slate-700 text-xs font-semibold transition text-center flex items-center justify-center gap-1">
              <span>Switch to AQMS</span>
              <span>→</span>
            </button>
          </div>
        </div>

        <!-- 2. WQMS Card -->
        <div class="p-4 rounded-xl border transition-all flex flex-col justify-between space-y-4"
             :class="form.monitoring_domain === 'WQMS'
               ? 'bg-blue-950/20 border-blue-500/50 shadow-lg shadow-blue-950/30 ring-1 ring-blue-500/30'
               : 'bg-[#111827] border-slate-800 hover:border-slate-700'">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-2xl p-2 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400">💧</span>
              <span v-if="form.monitoring_domain === 'WQMS'" class="text-4xs font-mono font-bold px-2 py-0.5 rounded-full bg-blue-500/20 text-blue-300 border border-blue-500/40">
                ACTIVE DOMAIN
              </span>
              <span v-else class="text-4xs font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700">
                READY
              </span>
            </div>

            <div>
              <h3 class="text-sm font-bold text-white font-sans flex items-center gap-1.5">
                WQMS
                <span class="text-2xs text-slate-400 font-normal">Water Quality</span>
              </h3>
              <p class="text-3xs text-slate-400 mt-1 leading-normal">
                Surface water, industrial effluent, wastewater treatment (WWTP/IPAL), and chemical hydrology.
              </p>
            </div>

            <div class="space-y-1.5 pt-2 border-t border-slate-800/80">
              <span class="text-4xs uppercase tracking-wider text-slate-400 font-bold block">Relevant Categories:</span>
              <div class="flex flex-wrap gap-1 text-3xs font-mono">
                <span class="px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-300 border border-blue-500/20">Water Quality</span>
                <span class="px-1.5 py-0.5 rounded bg-teal-500/10 text-teal-300 border border-teal-500/20">Physical Water</span>
                <span class="px-1.5 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/20">Chemical Water</span>
                <span class="px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">Flow & Level</span>
              </div>
            </div>
          </div>

          <div class="pt-2">
            <button v-if="form.monitoring_domain === 'WQMS'" disabled
                    class="w-full py-1.5 px-3 rounded-lg bg-blue-500/20 text-blue-300 border border-blue-500/30 text-xs font-semibold cursor-default text-center">
              ✓ Active Domain
            </button>
            <button v-else @click="initiateDomainSwitch('WQMS')"
                    class="w-full py-1.5 px-3 rounded-lg bg-slate-800 hover:bg-blue-600 hover:text-white text-slate-200 border border-slate-700 text-xs font-semibold transition text-center flex items-center justify-center gap-1">
              <span>Switch to WQMS</span>
              <span>→</span>
            </button>
          </div>
        </div>

        <!-- 3. SPARING Card -->
        <div class="p-4 rounded-xl border transition-all flex flex-col justify-between space-y-4"
             :class="form.monitoring_domain === 'SPARING'
               ? 'bg-teal-950/20 border-teal-500/50 shadow-lg shadow-teal-950/30 ring-1 ring-teal-500/30'
               : 'bg-[#111827] border-slate-800 hover:border-slate-700'">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-2xl p-2 rounded-xl bg-teal-500/10 border border-teal-500/20 text-teal-400">🌊</span>
              <span v-if="form.monitoring_domain === 'SPARING'" class="text-4xs font-mono font-bold px-2 py-0.5 rounded-full bg-teal-500/20 text-teal-300 border border-teal-500/40">
                ACTIVE DOMAIN
              </span>
              <span v-else class="text-4xs font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700">
                READY
              </span>
            </div>

            <div>
              <h3 class="text-sm font-bold text-white font-sans flex items-center gap-1.5">
                SPARING
                <span class="text-2xs text-slate-400 font-normal">Air Limbah KLHK</span>
              </h3>
              <p class="text-3xs text-slate-400 mt-1 leading-normal">
                Sistem Pemantauan Kualitas Air Limbah Terus Menerus dan Dalam Jaringan sesuai regulasi KLHK.
              </p>
            </div>

            <div class="space-y-1.5 pt-2 border-t border-slate-800/80">
              <span class="text-4xs uppercase tracking-wider text-slate-400 font-bold block">Relevant Categories:</span>
              <div class="flex flex-wrap gap-1 text-3xs font-mono">
                <span class="px-1.5 py-0.5 rounded bg-teal-500/10 text-teal-300 border border-teal-500/20">Wastewater</span>
                <span class="px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">Flow & Debit</span>
                <span class="px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-300 border border-blue-500/20">Effluent pH & COD</span>
              </div>
            </div>
          </div>

          <div class="pt-2">
            <button v-if="form.monitoring_domain === 'SPARING'" disabled
                    class="w-full py-1.5 px-3 rounded-lg bg-teal-500/20 text-teal-300 border border-teal-500/30 text-xs font-semibold cursor-default text-center">
              ✓ Active Domain
            </button>
            <button v-else @click="initiateDomainSwitch('SPARING')"
                    class="w-full py-1.5 px-3 rounded-lg bg-slate-800 hover:bg-teal-600 hover:text-white text-slate-200 border border-slate-700 text-xs font-semibold transition text-center flex items-center justify-center gap-1">
              <span>Switch to SPARING</span>
              <span>→</span>
            </button>
          </div>
        </div>

        <!-- 4. CUSTOM Card -->
        <div class="p-4 rounded-xl border transition-all flex flex-col justify-between space-y-4"
             :class="form.monitoring_domain === 'CUSTOM'
               ? 'bg-purple-950/20 border-purple-500/50 shadow-lg shadow-purple-950/30 ring-1 ring-purple-500/30'
               : 'bg-[#111827] border-slate-800 hover:border-slate-700'">
          <div class="space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-2xl p-2 rounded-xl bg-purple-500/10 border border-purple-500/20 text-purple-400">⚙️</span>
              <span v-if="form.monitoring_domain === 'CUSTOM'" class="text-4xs font-mono font-bold px-2 py-0.5 rounded-full bg-purple-500/20 text-purple-300 border border-purple-500/40">
                ACTIVE DOMAIN
              </span>
              <span v-else class="text-4xs font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700">
                READY
              </span>
            </div>

            <div>
              <h3 class="text-sm font-bold text-white font-sans flex items-center gap-1.5">
                CUSTOM
                <span class="text-2xs text-slate-400 font-normal">Multi-Discipline</span>
              </h3>
              <p class="text-3xs text-slate-400 mt-1 leading-normal">
                Configurable environmental, energy, vibration, or multi-domain industrial sensor telemetry.
              </p>
            </div>

            <div class="space-y-1.5 pt-2 border-t border-slate-800/80">
              <span class="text-4xs uppercase tracking-wider text-slate-400 font-bold block">Relevant Categories:</span>
              <div class="flex flex-wrap gap-1 text-3xs font-mono">
                <span class="px-1.5 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/20">All Categories Enabled</span>
                <span class="px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">Custom Taxonomies</span>
              </div>
            </div>
          </div>

          <div class="pt-2">
            <button v-if="form.monitoring_domain === 'CUSTOM'" disabled
                    class="w-full py-1.5 px-3 rounded-lg bg-purple-500/20 text-purple-300 border border-purple-500/30 text-xs font-semibold cursor-default text-center">
              ✓ Active Domain
            </button>
            <button v-else @click="initiateDomainSwitch('CUSTOM')"
                    class="w-full py-1.5 px-3 rounded-lg bg-slate-800 hover:bg-purple-600 hover:text-white text-slate-200 border border-slate-700 text-xs font-semibold transition text-center flex items-center justify-center gap-1">
              <span>Switch to CUSTOM</span>
              <span>→</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB: Organization & Branding Setup (#Admin Customer & Maker Branding) -->
    <div v-else-if="activeTab === 'branding'" class="space-y-5 max-w-5xl font-sans">
      <!-- Overview Card -->
      <div class="saas-card p-5 space-y-4 bg-gradient-to-br from-[#111827] to-[#131E35] border border-cyan-500/30">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800/80 pb-3">
          <div class="flex items-center space-x-3">
            <span class="p-2.5 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 text-lg">
              🏢
            </span>
            <div>
              <div class="flex items-center space-x-2">
                <h2 class="text-sm font-bold text-white tracking-tight">{{ $t('config.branding.title') || 'Organization & Branding Configuration' }}</h2>
                <span class="text-3xs font-mono font-bold px-2 py-0.5 rounded-full bg-cyan-500/20 text-cyan-300 border border-cyan-500/40">
                  SYSTEM ADMIN SETUP
                </span>
              </div>
              <p class="text-2xs text-slate-400 mt-0.5">
                {{ $t('config.branding.subtitle') || 'Customize customer client identity (left header) and maker developer branding (right header)' }}
              </p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <button
              type="button"
              @click="resetBrandingToDefault"
              class="px-3 py-1.5 rounded-lg border border-slate-700 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition"
            >
              {{ $t('config.branding.resetButton') || 'Reset to Default' }}
            </button>
            <button
              type="button"
              @click="saveBrandingConfig"
              :disabled="brandingSaving"
              class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 text-xs font-bold transition shadow-sm disabled:opacity-50"
            >
              {{ brandingSaving ? 'Saving...' : ($t('config.branding.saveButton') || 'Save All Branding') }}
            </button>
          </div>
        </div>

        <p class="text-xs text-slate-300 leading-relaxed">
          Nama dan icon perusahaan customer ditampilkan di sebelah kiri header Operator Dashboard, sedangkan nama dan icon perusahaan pembuat (vendor) ditampilkan di sebelah kanan. Pengaturan ini tersimpan di sistem lokal dan langsung tersinkronisasi ke seluruh dashboard operator.
        </p>

        <!-- Feedback Alert -->
        <div v-if="brandingSuccessMsg" class="p-3 rounded-lg bg-emerald-500/15 border border-emerald-500/40 text-emerald-300 text-xs font-mono flex items-center space-x-2">
          <span>✓</span>
          <span>{{ brandingSuccessMsg }}</span>
        </div>
        <div v-if="brandingErrorMsg" class="p-3 rounded-lg bg-rose-500/15 border border-rose-500/40 text-rose-300 text-xs font-mono flex items-center space-x-2">
          <span>⚠️</span>
          <span>{{ brandingErrorMsg }}</span>
        </div>
      </div>

      <!-- Live Dashboard Header Preview -->
      <div class="p-4 rounded-xl border border-slate-800 bg-[#0B1422] space-y-2">
        <div class="flex items-center justify-between text-2xs font-mono text-slate-400 uppercase tracking-wider">
          <span>{{ $t('config.branding.preview') || 'Live Operator Header Preview' }}</span>
          <span class="text-cyan-400">Scale: 100%</span>
        </div>
        <div class="p-2.5 rounded-lg border border-[#20354B] bg-[#0E1726] flex items-center justify-between gap-2 select-none">
          <!-- Customer Side (Left) -->
          <div class="flex items-center space-x-2 min-w-0">
            <div class="h-7 px-1.5 rounded border border-[#20354B] bg-[#101B2B] flex items-center justify-center shrink-0">
              <img v-if="customerLogoPreview" :src="customerLogoPreview" class="h-5 max-w-[80px] object-contain rounded-xs" alt="Customer Logo" />
              <div v-else class="w-5 h-5 rounded bg-gradient-to-br from-emerald-500 to-teal-600 flex items-center justify-center text-white text-3xs font-bold">
                🌿
              </div>
            </div>
            <div class="leading-none">
              <div class="text-[11px] font-bold text-white font-mono flex items-center gap-1">
                <span>{{ branding.customer.name || 'PT. Mitra Industri Enviro' }}</span>
                <span class="px-1 py-0.2 rounded text-[7.5px] bg-cyan-500/20 text-cyan-300 border border-cyan-500/30">SCADA</span>
              </div>
              <div v-if="branding.customer.subtitle && branding.customer.subtitle !== 'Local Autonomous Edge'" class="text-[9px] text-slate-400 font-mono mt-0.5">
                Station: <strong class="text-cyan-400">{{ branding.customer.subtitle }}</strong>
              </div>
            </div>
          </div>

          <!-- Centered Clock -->
          <div class="px-2.5 py-1 rounded bg-[#101B2B] border border-[#20354B] text-[9.5px] font-mono font-bold text-slate-300">
            🕒 12:00:00
          </div>

          <!-- Maker Side (Right) -->
          <div class="flex items-center space-x-2 shrink-0 text-right">
            <div v-if="branding.maker.name || branding.maker.subtitle" class="leading-none hidden sm:block">
              <div v-if="branding.maker.name" class="text-[10px] font-bold text-white font-mono uppercase tracking-tight">
                {{ branding.maker.name }}
              </div>
              <div v-if="branding.maker.subtitle" class="text-[8px] text-slate-400 font-mono mt-0.5">
                {{ branding.maker.subtitle }}
              </div>
            </div>
            <div class="h-7 px-1.5 rounded border border-[#20354B] bg-[#101B2B] flex items-center justify-center shrink-0">
              <img v-if="makerLogoPreview" :src="makerLogoPreview" class="h-5 max-w-[80px] object-contain rounded-xs" alt="Maker Logo" />
              <div v-else class="w-5 h-5 rounded bg-gradient-to-br from-cyan-500 to-blue-600 flex items-center justify-center text-white text-3xs font-bold">
                ⚡
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Two Setup Columns: Customer (Left) vs Maker (Right) -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <!-- 1. Customer Organization (Left Header) -->
        <div class="p-5 rounded-xl border border-slate-800 bg-[#111827] space-y-4">
          <div class="flex items-center space-x-2.5 border-b border-slate-800 pb-3">
            <div class="w-7 h-7 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center text-sm font-bold">
              🏢
            </div>
            <div>
              <h3 class="text-xs font-bold uppercase tracking-wider text-white">
                {{ $t('config.branding.customerSection') || 'Perusahaan Customer (Header Kiri)' }}
              </h3>
              <p class="text-3xs text-slate-400">Identitas klien / pemilik fasilitas pemantauan</p>
            </div>
          </div>

          <!-- Customer Company Name -->
          <div class="space-y-1">
            <label class="block text-3xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('config.branding.customerName') || 'Nama Perusahaan Customer' }}
            </label>
            <input
              v-model="branding.customer.name"
              type="text"
              class="w-full px-3 py-2 rounded-lg bg-[#0B1422] border border-slate-700 text-white font-mono text-xs focus:outline-none focus:border-cyan-500"
              placeholder="e.g. PT. Mitra Industri Enviro"
            />
          </div>

          <!-- Customer Subtitle / Station Type -->
          <div class="space-y-1">
            <div class="flex items-center justify-between">
              <label class="block text-3xs font-semibold text-slate-300 uppercase tracking-wider">
                Jenis Stasiun / Domain Pemantauan
              </label>
              <div class="flex items-center space-x-1">
                <button
                  type="button"
                  v-for="st in ['AQMS', 'WQMS', 'Sparing']"
                  :key="st"
                  @click="branding.customer.subtitle = st"
                  class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold transition border cursor-pointer"
                  :class="branding.customer.subtitle === st
                    ? 'bg-cyan-500 text-slate-950 border-cyan-400 shadow-xs'
                    : 'bg-slate-800 text-slate-300 border-slate-700 hover:text-cyan-400'"
                >
                  {{ st }}
                </button>
              </div>
            </div>
            <input
              v-model="branding.customer.subtitle"
              type="text"
              class="w-full px-3 py-2 rounded-lg bg-[#0B1422] border border-slate-700 text-white font-mono text-xs focus:outline-none focus:border-cyan-500"
              placeholder="e.g. AQMS, WQMS, Sparing"
            />
          </div>

          <!-- Customer Logo Upload & Preview -->
          <div class="space-y-2">
            <label class="block text-3xs font-semibold text-slate-300 uppercase tracking-wider">
              Logo Perusahaan Customer
            </label>
            <div class="flex items-center space-x-3">
              <div class="w-24 h-14 rounded-lg bg-[#0B1422] border border-slate-700 flex items-center justify-center p-1 overflow-hidden">
                <img v-if="customerLogoPreview" :src="customerLogoPreview" class="max-h-full max-w-full object-contain" alt="Customer Logo Preview" />
                <span v-else class="text-4xs text-slate-500 font-mono text-center">No Logo (Default Icon)</span>
              </div>
              <div class="flex flex-col space-y-1.5 flex-1">
                <label class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-3xs font-mono font-bold cursor-pointer transition text-center flex items-center justify-center space-x-1">
                  <span>📁</span>
                  <span>{{ customerLogoPreview ? 'Ganti Logo...' : 'Upload Logo...' }}</span>
                  <input type="file" accept="image/png,image/jpeg,image/svg+xml,image/webp" class="hidden" @change="handleCustomerLogoUpload" />
                </label>
                <button
                  v-if="customerLogoPreview"
                  type="button"
                  @click="deleteCustomerLogo"
                  class="px-2 py-1 rounded bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/30 text-4xs font-mono transition text-center"
                >
                  Hapus Logo (Gunakan Icon)
                </button>
              </div>
            </div>
          </div>
        </div>

        <!-- 2. Maker Company (Right Header) -->
        <div class="p-5 rounded-xl border border-slate-800 bg-[#111827] space-y-4">
          <div class="flex items-center space-x-2.5 border-b border-slate-800 pb-3">
            <div class="w-7 h-7 rounded-lg bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 flex items-center justify-center text-sm font-bold">
              ⚙️
            </div>
            <div>
              <h3 class="text-xs font-bold uppercase tracking-wider text-white">
                {{ $t('config.branding.makerSection') || 'Perusahaan Pembuat (Header Kanan)' }}
              </h3>
              <p class="text-3xs text-slate-400">Identitas pengembang sistem & vendor perangkat keras</p>
            </div>
          </div>

          <!-- Maker Company Name -->
          <div class="space-y-1">
            <label class="block text-3xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('config.branding.makerName') || 'Nama Perusahaan Pembuat' }}
            </label>
            <input
              v-model="branding.maker.name"
              type="text"
              class="w-full px-3 py-2 rounded-lg bg-[#0B1422] border border-slate-700 text-white font-mono text-xs focus:outline-none focus:border-cyan-500"
              placeholder="e.g. PT. Centra Biotech Indonesia"
            />
          </div>

          <!-- Maker Subtitle / Tagline -->
          <div class="space-y-1">
            <label class="block text-3xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('config.branding.makerSubtitle') || 'Tagline / Subtitle Pembuat' }}
            </label>
            <input
              v-model="branding.maker.subtitle"
              type="text"
              class="w-full px-3 py-2 rounded-lg bg-[#0B1422] border border-slate-700 text-white font-mono text-xs focus:outline-none focus:border-cyan-500"
              placeholder="e.g. CBI Industrial IoT"
            />
          </div>

          <!-- Maker Logo Upload & Preview -->
          <div class="space-y-2">
            <label class="block text-3xs font-semibold text-slate-300 uppercase tracking-wider">
              Logo Perusahaan Pembuat
            </label>
            <div class="flex items-center space-x-3">
              <div class="w-24 h-14 rounded-lg bg-[#0B1422] border border-slate-700 flex items-center justify-center p-1 overflow-hidden">
                <img v-if="makerLogoPreview" :src="makerLogoPreview" class="max-h-full max-w-full object-contain" alt="Maker Logo Preview" />
                <span v-else class="text-4xs text-slate-500 font-mono text-center">No Logo (Default Icon)</span>
              </div>
              <div class="flex flex-col space-y-1.5 flex-1">
                <label class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-3xs font-mono font-bold cursor-pointer transition text-center flex items-center justify-center space-x-1">
                  <span>📁</span>
                  <span>{{ makerLogoPreview ? 'Ganti Logo...' : 'Upload Logo...' }}</span>
                  <input type="file" accept="image/png,image/jpeg,image/svg+xml,image/webp" class="hidden" @change="handleMakerLogoUpload" />
                </label>
                <button
                  v-if="makerLogoPreview"
                  type="button"
                  @click="deleteMakerLogo"
                  class="px-2 py-1 rounded bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/30 text-4xs font-mono transition text-center"
                >
                  Hapus Logo (Gunakan Icon)
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- TAB 1: Database (MariaDB) -->
    <div v-else-if="activeTab === 'database'" class="space-y-5 max-w-4xl">
      <!-- Live Database Health Monitoring Card -->
      <div class="saas-card p-5 space-y-4 bg-gradient-to-br from-[#111827] to-[#131E35]">
        <div class="flex items-center justify-between border-b border-slate-800/80 pb-3">
          <div class="flex items-center space-x-2">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
            <h2 class="text-sm font-bold text-white font-sans tracking-tight">{{ $t('config.dbMonitor') || 'MariaDB Edge Engine Monitor' }}</h2>
          </div>
          <button @click="fetchDbStatus" :disabled="testing"
                  class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-medium transition-colors flex items-center">
            <svg class="w-3.5 h-3.5 mr-1.5" :class="{ 'animate-spin': testing }" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <polyline points="23 4 23 10 17 10"></polyline>
              <polyline points="1 20 1 14 7 14"></polyline>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
            </svg>
            {{ $t('common.refresh') || 'Refresh' }}
          </button>
        </div>

        <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs font-sans">
          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">{{ $t('config.dbType') || 'Database Engine' }}</span>
            <span class="font-bold text-white text-sm block">MariaDB</span>
            <span class="text-3xs text-emerald-400 font-mono mt-0.5 block">● Connected</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">{{ $t('config.hostPort') || 'Host & Port' }}</span>
            <span class="font-mono text-slate-200 text-xs font-semibold block">{{ dbStatus.host || '127.0.0.1' }}:{{ dbStatus.port || '3306' }}</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">{{ dbStatus.database || 'datalogger' }}</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">{{ $t('devices.latency') || 'Ping Latency' }}</span>
            <span class="font-mono text-emerald-400 text-sm font-bold block">{{ dbStatus.latency_ms ? dbStatus.latency_ms.toFixed(2) : '0.45' }} ms</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">{{ dbStatus.version || '10.4.32-MariaDB' }}</span>
          </div>

          <div class="p-3 rounded-lg bg-[#0B0F19]/80 border border-slate-800/80">
            <span class="text-3xs uppercase tracking-wider text-slate-400 block mb-1">Active Pool Conns</span>
            <span class="font-mono text-blue-400 text-sm font-bold block">{{ dbStatus.active_connections || 1 }} / {{ dbStatus.max_connections || 25 }}</span>
            <span class="text-3xs text-slate-400 font-mono mt-0.5 block">{{ dbStatus.database_size_mb || 1.08 }} MB</span>
          </div>
        </div>
      </div>

      <!-- Database Configuration Form -->
      <div class="saas-card p-5 space-y-4">
        <div class="border-b border-slate-800/80 pb-3 flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-white font-sans tracking-tight">Database Connectivity Parameters</h2>
            <p class="text-2xs text-slate-400 font-sans mt-0.5">Modifying host, port, or database schema requires service restart.</p>
          </div>
          <span class="text-3xs font-mono px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
            RESTART REQUIRED
          </span>
        </div>

        <!-- Connection Test Result -->
        <div v-if="testResult" class="p-3 rounded-lg text-xs font-sans flex items-center justify-between"
             :class="testResult.success ? 'bg-emerald-500/10 border border-emerald-500/30 text-emerald-300' : 'bg-rose-500/10 border border-rose-500/30 text-rose-300'">
          <div>
            <span class="font-semibold">{{ testResult.success ? 'Connection Successful' : 'Connection Failed' }}</span>
            <span class="block text-2xs mt-0.5">{{ testResult.message }}</span>
          </div>
          <span v-if="testResult.latency" class="font-mono text-2xs">{{ testResult.latency.toFixed(2) }}ms</span>
        </div>

        <form @submit.prevent="submitConfig" class="space-y-4 text-xs font-sans">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Host / IP Address</label>
              <input type="text" v-model="form.db_host" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Database Port</label>
              <input type="text" v-model="form.db_port" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Database Name / Schema</label>
              <input type="text" v-model="form.db_name" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">Username</label>
              <input type="text" v-model="form.db_user" required class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-sans" />
            </div>

            <div class="sm:col-span-2">
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5 font-sans">
                Password <span class="text-slate-500 font-normal lowercase">(leave as •••••••• to keep current secret)</span>
              </label>
              <input type="password" v-model="form.db_password" placeholder="••••••••" class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>
          </div>

          <div class="pt-3 border-t border-slate-800/80 flex items-center justify-end space-x-3">
            <button type="button" @click="testDbConnection" :disabled="testing"
                    class="px-4 py-2 bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 rounded-lg text-xs font-medium transition-colors flex items-center">
              <span v-if="testing" class="inline-block w-3.5 h-3.5 border-2 border-slate-400 border-t-transparent rounded-full animate-spin mr-1.5"></span>
              Test Connection
            </button>

            <button type="submit" :disabled="saving"
                    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
              <span v-if="saving" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
              Save Configuration
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- TAB 2: Serial & Communication Ports (#83) -->
    <div v-else-if="activeTab === 'serial'" class="space-y-4 max-w-4xl font-sans">
      <div class="saas-card p-5 space-y-4">
        <div class="border-b border-slate-800/80 pb-3 flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-white tracking-tight">Hardware Serial Bus & Port Defaults</h2>
            <p class="text-2xs text-slate-400 mt-0.5">Default communication parameters for RS-485 and USB Serial COM interfaces.</p>
          </div>
          <span class="text-3xs font-mono px-2 py-0.5 rounded bg-rose-500/10 text-rose-400 border border-rose-500/20">
            MAY DISRUPT ACQUISITION
          </span>
        </div>

        <!-- Warning notice -->
        <div class="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-xs text-amber-300 space-y-1">
          <div class="font-bold">⚠️ Acquisition Interruption Notice</div>
          <p class="text-2xs text-amber-400/90 leading-relaxed">
            Changing serial baud rate, parity, or COM port settings will cause active Modbus RTU polling routines on affected ports to reconnect. Ensure field instruments support the new baud rate.
          </p>
        </div>

        <form @submit.prevent="submitConfig" class="space-y-4 text-xs font-sans">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Default Serial Port</label>
              <input type="text" v-model="form.serial_port" placeholder="e.g. COM1 or /dev/ttyUSB0"
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Baud Rate (bps)</label>
              <select v-model.number="form.serial_baud_rate"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono">
                <option :value="1200">1200</option>
                <option :value="2400">2400</option>
                <option :value="4800">4800</option>
                <option :value="9600">9600 (Standard)</option>
                <option :value="19200">19200</option>
                <option :value="38400">38400</option>
                <option :value="57600">57600</option>
                <option :value="115200">115200 (High-Speed)</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Data Bits</label>
              <select v-model.number="form.serial_data_bits"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono">
                <option :value="8">8 Data Bits (Standard)</option>
                <option :value="7">7 Data Bits</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Stop Bits</label>
              <select v-model.number="form.serial_stop_bits"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono">
                <option :value="1">1 Stop Bit</option>
                <option :value="2">2 Stop Bits</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Parity</label>
              <select v-model="form.serial_parity"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono">
                <option value="N">None (N)</option>
                <option value="E">Even (E)</option>
                <option value="O">Odd (O)</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs font-semibold uppercase tracking-wider text-slate-400 mb-1.5">Response Timeout (ms)</label>
              <input type="number" v-model.number="form.serial_timeout_ms" min="100" max="10000" step="50"
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 text-white font-mono" />
            </div>
          </div>

          <div class="pt-3 border-t border-slate-800/80 flex items-center justify-end space-x-3">
            <button type="submit" :disabled="saving"
                    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
              <span v-if="saving" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
              Save Serial Parameters
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- TAB 3: General & Server Runtime -->
    <div v-else-if="activeTab === 'general'" class="space-y-4 max-w-4xl font-sans">
      <div class="saas-card p-5 space-y-4">
        <div class="border-b border-slate-800/80 pb-3 flex items-center justify-between">
          <div>
            <h2 class="text-sm font-bold text-white tracking-tight">HTTP Server & Runtime Governance</h2>
            <p class="text-2xs text-slate-400 mt-0.5">Port allocation, system log levels, and session security token lifetimes.</p>
          </div>
        </div>

        <form @submit.prevent="submitConfig" class="space-y-4 text-xs">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <label class="text-3xs text-slate-400 uppercase tracking-wider font-semibold">HTTP Service Port</label>
                <span class="text-3xs text-amber-400 font-mono">Requires Restart</span>
              </div>
              <input type="text" v-model="form.port" required
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 font-mono text-white" />
            </div>

            <div>
              <label class="block text-3xs text-slate-400 uppercase tracking-wider mb-1.5 font-semibold">Environment Profile</label>
              <input type="text" v-model="form.environment"
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 font-mono text-white" />
            </div>

            <div>
              <label class="block text-3xs text-slate-400 uppercase tracking-wider mb-1.5 font-semibold">Logging Level</label>
              <select v-model="form.log_level"
                      class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 font-mono text-white">
                <option value="debug">DEBUG (High Verbosity)</option>
                <option value="info">INFO (Standard Production)</option>
                <option value="warn">WARN (Errors & Warnings Only)</option>
                <option value="error">ERROR (Critical Failures Only)</option>
              </select>
            </div>

            <div>
              <label class="block text-3xs text-slate-400 uppercase tracking-wider mb-1.5 font-semibold">JWT Session Expiration (Hours)</label>
              <input type="number" v-model.number="form.jwt_expiration_hours" min="1" max="720"
                     class="w-full bg-[#0B0F19] border border-slate-700/80 focus:border-blue-500 rounded-lg px-3 py-2 font-mono text-white" />
            </div>
          </div>

          <div class="pt-3 border-t border-slate-800/80 flex items-center justify-end space-x-3">
            <button type="submit" :disabled="saving"
                    class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-xs font-semibold shadow-soft transition-colors flex items-center">
              <span v-if="saving" class="inline-block w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin mr-1.5"></span>
              Apply Runtime Settings
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- TAB 4: Local-First Principle -->
    <div v-else class="saas-card p-5 space-y-3 max-w-4xl font-sans">
      <h2 class="text-sm font-bold text-white tracking-tight border-b border-slate-800/80 pb-3">
        Local-First Edge Architecture
      </h2>
      <p class="text-xs text-slate-300 leading-relaxed">
        The system is designed with a strict local-first architecture. All operational telemetry, user accounts, security roles, alarms, and configuration parameters are persisted directly onto the edge hardware (MariaDB Edge). In the event of network disruption or cloud disconnects, telemetry acquisition, threshold monitoring, and local reporting continue uninterrupted.
      </p>
    </div>

    <!-- MODAL 1: Monitoring Domain Transition Validation & Confirmation Modal -->
    <div v-if="showDomainModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
      <div class="bg-[#0B1422] border border-[#20354B] rounded-2xl w-full max-w-2xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
        <!-- Modal Header -->
        <div class="px-6 py-4 border-b border-[#20354B] flex items-center justify-between bg-[#080E18]">
          <div class="flex items-center space-x-2.5">
            <div class="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 text-base">
              🔄
            </div>
            <div>
              <h2 class="text-sm font-bold text-white tracking-wide font-mono uppercase">
                {{ $t('config.domain.confirmModalTitle') || 'Confirm Monitoring Domain Transition' }}
              </h2>
              <p class="text-3xs text-slate-400 mt-0.5 font-mono">
                Transitioning primary domain from <strong class="text-cyan-400">{{ form.monitoring_domain || 'AQMS' }}</strong> ➔ <strong class="text-emerald-400">{{ targetDomainForModal }}</strong>
              </p>
            </div>
          </div>
          <button @click="showDomainModal = false" class="text-slate-400 hover:text-white p-1 rounded-lg">✕</button>
        </div>

        <!-- Modal Body -->
        <div class="flex-1 overflow-y-auto p-6 space-y-4 custom-scroller text-xs">
          <!-- Zero Data Deletion Guarantee Banner -->
          <div class="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 space-y-1">
            <div class="font-bold flex items-center gap-1.5">
              <span>🛡️</span>
              <span>Zero Data Deletion Guarantee</span>
            </div>
            <p class="text-2xs text-emerald-300/90 leading-relaxed">
              {{ $t('config.domain.zeroDataLossGuarantee') || 'Changing the monitoring domain will NEVER delete existing devices, telemetry records, aggregations, alarms, or audit history.' }}
            </p>
          </div>

          <!-- Compatibility Summary Stats -->
          <div v-if="compatibilityReport" class="grid grid-cols-3 gap-3 font-mono text-center">
            <div class="p-3 rounded-xl bg-[#080E18] border border-slate-800">
              <span class="text-3xs text-slate-400 block uppercase">Total Parameters</span>
              <span class="text-base font-bold text-white mt-0.5 block">{{ compatibilityReport.total_parameters }}</span>
            </div>
            <div class="p-3 rounded-xl bg-[#080E18] border border-emerald-500/30">
              <span class="text-3xs text-emerald-400 block uppercase">Compatible</span>
              <span class="text-base font-bold text-emerald-400 mt-0.5 block">{{ compatibilityReport.compatible_parameters }}</span>
            </div>
            <div class="p-3 rounded-xl bg-[#080E18] border"
                 :class="compatibilityReport.incompatible_parameters > 0 ? 'border-amber-500/30' : 'border-slate-800'">
              <span class="text-3xs block uppercase" :class="compatibilityReport.incompatible_parameters > 0 ? 'text-amber-400' : 'text-slate-400'">
                Incompatible
              </span>
              <span class="text-base font-bold mt-0.5 block" :class="compatibilityReport.incompatible_parameters > 0 ? 'text-amber-400' : 'text-slate-300'">
                {{ compatibilityReport.incompatible_parameters }}
              </span>
            </div>
          </div>

          <!-- Incompatible Parameters Review & Controlled Reassignment -->
          <div v-if="compatibilityReport && compatibilityReport.incompatible_parameters > 0" class="space-y-2">
            <div class="p-3 rounded-xl bg-amber-500/10 border border-amber-500/20 text-2xs text-amber-300 space-y-1">
              <div class="font-bold flex items-center gap-1">
                <span>⚠️</span>
                <span>Controlled Reassignment Review</span>
              </div>
              <p class="text-amber-300/90 leading-relaxed">
                {{ $t('config.domain.incompatibleNotice') || 'The following parameters belong to categories outside the target domain. They will remain fully accessible under \'Uncategorized\' or you may reassign them below:' }}
              </p>
            </div>

            <div class="space-y-2 max-h-56 overflow-y-auto custom-scroller pr-1">
              <div v-for="inc in compatibilityReport.incompatibilities" :key="inc.parameter_id"
                   class="p-2.5 rounded-lg bg-[#080E18] border border-slate-800 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-2xs font-mono">
                <div>
                  <div class="font-bold text-white">{{ inc.parameter_name }} <span class="text-slate-400 font-normal">({{ inc.parameter_code }})</span></div>
                  <div class="text-3xs text-slate-400">Current Category: <span class="text-cyan-400">{{ inc.category_name }}</span> ({{ inc.category_domain }})</div>
                </div>

                <div class="flex items-center space-x-2">
                  <select v-model="parameterReassignments[inc.parameter_id]"
                          class="bg-[#101B2B] border border-slate-700 rounded-lg px-2.5 py-1 text-white text-3xs focus:outline-none focus:border-cyan-500">
                    <option :value="0">Keep as Uncategorized (Safely Preserved)</option>
                    <option v-for="cat in availableTargetCategories" :key="cat.id" :value="cat.id">
                      Reassign: {{ cat.name }}
                    </option>
                  </select>
                </div>
              </div>
            </div>
          </div>

          <div v-else-if="compatibilityReport && compatibilityReport.incompatible_parameters === 0" class="p-3 rounded-xl bg-cyan-500/10 border border-cyan-500/30 text-cyan-300 text-2xs flex items-center gap-2">
            <span>✓</span>
            <span>All active parameters and categories are 100% compatible with the {{ targetDomainForModal }} domain.</span>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="px-6 py-3 border-t border-[#20354B] bg-[#080E18] flex items-center justify-end space-x-3">
          <button @click="showDomainModal = false" :disabled="switchingDomain"
                  class="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition">
            {{ $t('config.domain.cancelButton') || 'Cancel' }}
          </button>
          <button @click="executeDomainSwitch" :disabled="switchingDomain"
                  class="px-4 py-2 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 text-xs font-bold transition flex items-center gap-1.5 shadow-md shadow-cyan-900/20">
            <span v-if="switchingDomain" class="inline-block w-3.5 h-3.5 border-2 border-slate-950 border-t-transparent rounded-full animate-spin mr-1"></span>
            <span>{{ $t('config.domain.confirmButton') || 'Apply Domain Transition' }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- MODAL 2: First-Run Setup Wizard Modal -->
    <div v-if="showSetupWizard" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
      <div class="bg-[#0B1422] border border-[#20354B] rounded-2xl w-full max-w-xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
        <!-- Wizard Header -->
        <div class="px-6 py-4 border-b border-[#20354B] flex items-center justify-between bg-[#080E18]">
          <div class="flex items-center space-x-2.5">
            <div class="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20 text-base">
              🚀
            </div>
            <div>
              <h2 class="text-sm font-bold text-white tracking-wide font-mono uppercase">
                {{ $t('config.firstRun.title') || 'Application First-Run Setup Wizard' }}
              </h2>
              <p class="text-3xs text-slate-400 mt-0.5">
                {{ $t('config.firstRun.subtitle') || 'Configure initial monitoring domain and station identity for this edge installation.' }}
              </p>
            </div>
          </div>
          <button @click="showSetupWizard = false" class="text-slate-400 hover:text-white p-1 rounded-lg">✕</button>
        </div>

        <!-- Wizard Step Indicators -->
        <div class="px-6 py-2.5 bg-[#0e192a] border-b border-[#20354B] flex items-center justify-between text-3xs font-mono text-slate-400">
          <span :class="wizardStep >= 1 ? 'text-cyan-400 font-bold' : ''">1. Monitoring Domain</span>
          <span>→</span>
          <span :class="wizardStep >= 2 ? 'text-cyan-400 font-bold' : ''">2. Edge Identity</span>
          <span>→</span>
          <span :class="wizardStep >= 3 ? 'text-cyan-400 font-bold' : ''">3. Review & Save</span>
        </div>

        <!-- Wizard Body -->
        <div class="flex-1 overflow-y-auto p-6 space-y-4 custom-scroller text-xs">
          <!-- Step 1: Domain Selection -->
          <div v-if="wizardStep === 1" class="space-y-3">
            <div class="text-2xs text-slate-300 font-semibold">
              Select the primary monitoring domain for this station:
            </div>

            <div class="grid grid-cols-1 gap-2.5">
              <label v-for="d in [
                       { code: 'AQMS', name: 'Air Quality Monitoring System (AQMS)', icon: '💨', desc: 'Gas pollutants, PM2.5, PM10, meteorology' },
                       { code: 'WQMS', name: 'Water Quality Monitoring System (WQMS)', icon: '💧', desc: 'Effluent, wastewater, pH, DO, turbidity, hydrology' },
                       { code: 'CUSTOM', name: 'Custom Multi-Domain Deployment', icon: '⚙️', desc: 'All categories enabled, user-defined taxonomies' },
                     ]"
                     :key="d.code"
                     class="p-3 rounded-xl border cursor-pointer flex items-center justify-between transition"
                     :class="wizardDomain === d.code ? 'bg-cyan-950/30 border-cyan-500/60 ring-1 ring-cyan-500/40' : 'bg-[#080E18] border-slate-800 hover:border-slate-700'">
                <div class="flex items-center space-x-3">
                  <input type="radio" v-model="wizardDomain" :value="d.code" class="text-cyan-500 focus:ring-0" />
                  <div>
                    <div class="font-bold text-white flex items-center gap-1.5">
                      <span>{{ d.icon }}</span>
                      <span>{{ d.name }}</span>
                    </div>
                    <div class="text-3xs text-slate-400 mt-0.5">{{ d.desc }}</div>
                  </div>
                </div>
                <span class="text-3xs font-mono font-bold px-2 py-0.5 rounded"
                      :class="wizardDomain === d.code ? 'bg-cyan-500/20 text-cyan-300' : 'bg-slate-800 text-slate-400'">
                  {{ d.code }}
                </span>
              </label>
            </div>
          </div>

          <!-- Step 2: Station Identity -->
          <div v-else-if="wizardStep === 2" class="space-y-3">
            <div class="text-2xs text-slate-300 font-semibold">
              Configure station and site identification:
            </div>

            <div class="space-y-3 font-mono text-2xs">
              <div>
                <label class="block text-slate-400 mb-1 uppercase font-semibold">Station / Host Identifier</label>
                <input type="text" v-model="wizardStationName"
                       class="w-full bg-[#080E18] border border-slate-700 rounded-lg px-3 py-2 text-white" />
              </div>
              <div>
                <label class="block text-slate-400 mb-1 uppercase font-semibold">Deployment Site / Location</label>
                <input type="text" v-model="wizardSiteName"
                       class="w-full bg-[#080E18] border border-slate-700 rounded-lg px-3 py-2 text-white" />
              </div>
              <div>
                <label class="block text-slate-400 mb-1 uppercase font-semibold">Operating Organization</label>
                <input type="text" v-model="wizardOrganization"
                       class="w-full bg-[#080E18] border border-slate-700 rounded-lg px-3 py-2 text-white" />
              </div>
            </div>
          </div>

          <!-- Step 3: Review & Summary -->
          <div v-else-if="wizardStep === 3" class="space-y-3 font-mono text-2xs">
            <div class="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-300 space-y-1">
              <div class="font-bold">✓ Ready to Activate Application Profile</div>
              <p class="text-3xs text-emerald-300/90 leading-relaxed">
                The setup wizard will persist this monitoring domain to the edge configuration without disrupting historical records.
              </p>
            </div>

            <div class="p-3.5 rounded-xl bg-[#080E18] border border-slate-800 space-y-2">
              <div class="flex justify-between py-1 border-b border-slate-800/80">
                <span class="text-slate-400">Target Monitoring Domain:</span>
                <span class="font-bold text-cyan-400">{{ wizardDomain }}</span>
              </div>
              <div class="flex justify-between py-1 border-b border-slate-800/80">
                <span class="text-slate-400">Station Identifier:</span>
                <span class="text-white">{{ wizardStationName }}</span>
              </div>
              <div class="flex justify-between py-1 border-b border-slate-800/80">
                <span class="text-slate-400">Deployment Location:</span>
                <span class="text-white">{{ wizardSiteName }}</span>
              </div>
              <div class="flex justify-between py-1">
                <span class="text-slate-400">Operating Organization:</span>
                <span class="text-white">{{ wizardOrganization }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Wizard Footer -->
        <div class="px-6 py-3 border-t border-[#20354B] bg-[#080E18] flex items-center justify-between">
          <button v-if="wizardStep > 1" @click="wizardStep--"
                  class="px-3 py-1.5 rounded-lg bg-slate-800 text-slate-300 text-xs font-semibold">
            ← Back
          </button>
          <div v-else></div>

          <div class="flex items-center space-x-2">
            <button @click="showSetupWizard = false"
                    class="px-3 py-1.5 rounded-lg bg-slate-800 text-slate-400 text-xs hover:text-white">
              Cancel
            </button>
            <button v-if="wizardStep < 3" @click="wizardStep++"
                    class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-xs transition">
              Next →
            </button>
            <button v-else @click="applyWizardSetup" :disabled="saving"
                    class="px-4 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-xs transition flex items-center gap-1.5">
              <span v-if="saving" class="inline-block w-3 h-3 border-2 border-slate-950 border-t-transparent rounded-full animate-spin"></span>
              <span>Complete Setup & Launch</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import {
  loadOrganizationBranding,
  saveCustomerBranding,
  removeCustomerLogo,
  saveMakerBranding,
  removeMakerLogo,
  resetOrganizationBranding
} from '../utils/companyBranding';

export default {
  name: 'ConfigView',
  data() {
    return {
      activeTab: 'domain',
      branding: {
        customer: { name: '', subtitle: '', logoUrl: null },
        maker: { name: '', subtitle: '', logoUrl: null }
      },
      customerLogoFile: null,
      customerLogoPreview: null,
      makerLogoFile: null,
      makerLogoPreview: null,
      brandingSaving: false,
      brandingSuccessMsg: '',
      brandingErrorMsg: '',
      testing: false,
      saving: false,
      testResult: null,
      saveResult: null,
      selectedTargetDomain: '',
      targetDomainForModal: '',
      showDomainModal: false,
      switchingDomain: false,
      compatibilityReport: null,
      parameterReassignments: {},
      availableTargetCategories: [],
      showSetupWizard: false,
      wizardStep: 1,
      wizardDomain: 'AQMS',
      wizardStationName: 'Edge-Node-01',
      wizardSiteName: 'Local Autonomous Edge',
      wizardOrganization: 'Environmental Monitoring Division',
      dbStatus: {
        host: '127.0.0.1',
        port: '3306',
        database: 'datalogger',
        status: 'Connected',
        latency_ms: 0.45,
        version: '10.4.32-MariaDB',
        active_connections: 1,
        max_connections: 25,
        database_size_mb: 1.08,
      },
      form: {
        monitoring_domain: 'AQMS',
        db_host: '127.0.0.1',
        db_port: '3306',
        db_name: 'datalogger',
        db_user: 'root',
        db_password: '',
        port: '8080',
        environment: 'production',
        log_level: 'info',
        jwt_expiration_hours: 24,
        serial_port: 'COM1',
        serial_baud_rate: 9600,
        serial_data_bits: 8,
        serial_stop_bits: 1,
        serial_parity: 'N',
        serial_timeout_ms: 1000,
      },
    };
  },
  computed: {
    currentDomainBadgeClass() {
      const d = (this.form.monitoring_domain || 'AQMS').toUpperCase();
      if (d === 'AQMS') return 'bg-cyan-500/10 text-cyan-400 border border-cyan-500/30';
      if (d === 'WQMS') return 'bg-blue-500/10 text-blue-400 border border-blue-500/30';
      if (d === 'SPARING') return 'bg-teal-500/10 text-teal-400 border border-teal-500/30';
      return 'bg-purple-500/10 text-purple-400 border border-purple-500/30';
    },
  },
  mounted() {
    if (this.$route.query && this.$route.query.tab) {
      this.activeTab = this.$route.query.tab;
    }
    this.fetchDbStatus();
    this.fetchConfig();
    this.loadBrandingConfig();
  },
  methods: {
    async loadBrandingConfig() {
      try {
        const data = await loadOrganizationBranding();
        this.branding = data;
        this.customerLogoPreview = data.customer.logoUrl;
        this.makerLogoPreview = data.maker.logoUrl;
      } catch (e) {
        console.warn('Could not load branding:', e);
      }
    },
    handleCustomerLogoUpload(e) {
      const file = e.target.files && e.target.files[0];
      if (!file) return;
      this.customerLogoFile = file;
      this.customerLogoPreview = URL.createObjectURL(file);
    },
    async deleteCustomerLogo() {
      await removeCustomerLogo();
      this.customerLogoFile = null;
      this.customerLogoPreview = null;
      this.branding.customer.logoUrl = null;
      this.branding.customer.hasLogo = false;
    },
    handleMakerLogoUpload(e) {
      const file = e.target.files && e.target.files[0];
      if (!file) return;
      this.makerLogoFile = file;
      this.makerLogoPreview = URL.createObjectURL(file);
    },
    async deleteMakerLogo() {
      await removeMakerLogo();
      this.makerLogoFile = null;
      this.makerLogoPreview = null;
      this.branding.maker.logoUrl = null;
      this.branding.maker.hasLogo = false;
    },
    async saveBrandingConfig() {
      this.brandingSaving = true;
      this.brandingSuccessMsg = '';
      this.brandingErrorMsg = '';
      try {
        await saveCustomerBranding({
          file: this.customerLogoFile,
          name: this.branding.customer.name,
          subtitle: this.branding.customer.subtitle
        });
        await saveMakerBranding({
          file: this.makerLogoFile,
          name: this.branding.maker.name,
          subtitle: this.branding.maker.subtitle
        });
        this.customerLogoFile = null;
        this.makerLogoFile = null;
        await this.loadBrandingConfig();
        this.brandingSuccessMsg = 'Organization & Branding updated successfully!';
        setTimeout(() => { this.brandingSuccessMsg = ''; }, 4000);
      } catch (err) {
        this.brandingErrorMsg = err.message || 'Failed to save branding.';
      } finally {
        this.brandingSaving = false;
      }
    },
    async resetBrandingToDefault() {
      if (!confirm('Reset company branding to standard defaults?')) return;
      await resetOrganizationBranding();
      this.customerLogoFile = null;
      this.makerLogoFile = null;
      await this.loadBrandingConfig();
      this.brandingSuccessMsg = 'Branding reset to default.';
      setTimeout(() => { this.brandingSuccessMsg = ''; }, 4000);
    },
    openSetupWizard() {
      this.wizardStep = 1;
      this.wizardDomain = this.form.monitoring_domain || 'AQMS';
      this.showSetupWizard = true;
    },
    async initiateDomainSwitch(targetDomain) {
      if (targetDomain === this.form.monitoring_domain) return;
      this.targetDomainForModal = targetDomain;
      this.selectedTargetDomain = targetDomain;
      this.compatibilityReport = null;
      this.parameterReassignments = {};
      this.availableTargetCategories = [];
      try {
        const res = await axios.get(`/api/system/config/domain-compatibility?target_domain=${targetDomain}`);
        if (res.data && res.data.data) {
          this.compatibilityReport = res.data.data;
          this.availableTargetCategories = res.data.data.available_categories || [];
          if (this.compatibilityReport.incompatibilities) {
            this.compatibilityReport.incompatibilities.forEach(inc => {
              this.parameterReassignments[inc.parameter_id] = 0;
            });
          }
        }
      } catch (err) {
        console.warn('Failed to fetch domain compatibility:', err);
      }
      this.showDomainModal = true;
    },
    async executeDomainSwitch() {
      this.switchingDomain = true;
      try {
        const reassignmentsArray = [];
        Object.keys(this.parameterReassignments).forEach(paramIdStr => {
          const targetCatId = this.parameterReassignments[paramIdStr];
          reassignmentsArray.push({
            parameter_id: Number(paramIdStr),
            target_category_id: Number(targetCatId),
          });
        });

        const payload = {
          ...this.form,
          monitoring_domain: this.targetDomainForModal,
          reassignments: reassignmentsArray,
        };

        const res = await axios.post('/api/system/config', payload);
        if (res.data && res.data.success) {
          this.form.monitoring_domain = this.targetDomainForModal;
          this.$store.commit('SET_SYSTEM_STATUS', {
            monitoring_domain: this.targetDomainForModal,
          });
          this.saveResult = {
            restart_required: false,
            message: `Monitoring domain successfully transitioned to ${this.targetDomainForModal}. Zero data deletion guaranteed.`,
          };
          this.showDomainModal = false;
          await this.fetchConfig();
        }
      } catch (err) {
        alert('Failed to switch monitoring domain: ' + (err.response?.data?.error || err.message));
      } finally {
        this.switchingDomain = false;
      }
    },
    async applyWizardSetup() {
      this.saving = true;
      try {
        const payload = {
          ...this.form,
          monitoring_domain: this.wizardDomain,
        };
        const res = await axios.post('/api/system/config', payload);
        if (res.data && res.data.success) {
          this.form.monitoring_domain = this.wizardDomain;
          this.$store.commit('SET_SYSTEM_STATUS', {
            monitoring_domain: this.wizardDomain,
          });
          this.saveResult = {
            restart_required: false,
            message: `Initial deployment setup completed. Active domain: ${this.wizardDomain}.`,
          };
          this.showSetupWizard = false;
          await this.fetchConfig();
        }
      } catch (err) {
        alert('Failed to complete setup: ' + (err.response?.data?.error || err.message));
      } finally {
        this.saving = false;
      }
    },
    async fetchDbStatus() {
      try {
        const res = await axios.get('/api/system/database/status');
        if (res.data && res.data.data) {
          this.dbStatus = res.data.data;
        }
      } catch (err) {
        console.warn('Failed to fetch DB status:', err);
      }
    },
    async fetchConfig() {
      try {
        const res = await axios.get('/api/system/config');
        if (res.data && res.data.data) {
          const cfg = res.data.data;
          this.form = {
            monitoring_domain: cfg.monitoring_domain || 'AQMS',
            db_host: cfg.db_host || '127.0.0.1',
            db_port: cfg.db_port || '3306',
            db_name: cfg.db_name || 'datalogger',
            db_user: cfg.db_user || 'root',
            db_password: '',
            port: cfg.port || '8080',
            environment: cfg.environment || 'production',
            log_level: cfg.log_level || 'info',
            jwt_expiration_hours: cfg.jwt_expiration_hours || 24,
            serial_port: cfg.serial_port || 'COM1',
            serial_baud_rate: cfg.serial_baud_rate || 9600,
            serial_data_bits: cfg.serial_data_bits || 8,
            serial_stop_bits: cfg.serial_stop_bits || 1,
            serial_parity: cfg.serial_parity || 'N',
            serial_timeout_ms: cfg.serial_timeout_ms || 1000,
          };
        }
      } catch (err) {
        console.warn('Failed to fetch system config:', err);
      }
    },
    async testDbConnection() {
      this.testing = true;
      this.testResult = null;
      try {
        const payload = {
          host: this.form.db_host,
          port: this.form.db_port,
          database: this.form.db_name,
          username: this.form.db_user,
          password: this.form.db_password,
        };
        const res = await axios.post('/api/system/database/test', payload);
        if (res.data) {
          if (res.data.success && res.data.data && res.data.data.connected) {
            this.testResult = {
              success: true,
              message: `Connected to ${res.data.data.version} on ${this.form.db_host}:${this.form.db_port}`,
              latency: res.data.data.latency_ms,
            };
          } else {
            this.testResult = {
              success: false,
              message: res.data.error || 'Failed to connect to MariaDB',
            };
          }
        }
      } catch (err) {
        this.testResult = {
          success: false,
          message: err.response?.data?.error || err.message,
        };
      } finally {
        this.testing = false;
      }
    },
    async submitConfig() {
      if (this.activeTab === 'serial') {
        if (!confirm('Warning: Modifying serial parameters may temporarily interrupt active telemetry acquisition. Do you wish to continue?')) {
          return;
        }
      }

      this.saving = true;
      this.saveResult = null;
      try {
        const res = await axios.post('/api/system/config', this.form);
        if (res.data && res.data.success) {
          this.saveResult = res.data.data;
          this.fetchDbStatus();
        }
      } catch (err) {
        alert('Failed to save config: ' + (err.response?.data?.error || err.message));
      } finally {
        this.saving = false;
      }
    },
  },
};
</script>

