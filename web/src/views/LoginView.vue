<template>
  <div
    class="login-shell min-h-screen w-full flex flex-col justify-between relative overflow-hidden font-sans transition-colors duration-200"
    :class="isDarkMode ? 'login-dark' : 'login-light'"
  >
    <!-- Architectural Engineering Grid & Ambient Glows -->
    <div class="login-grid-pattern absolute inset-0 pointer-events-none"></div>
    <div class="login-orb login-orb-top pointer-events-none"></div>
    <div class="login-orb login-orb-bottom pointer-events-none"></div>

    <!-- Top Navigation Bar: System Status, Platform Title, Language & Theme -->
    <header class="login-topbar w-full z-10">
      <div class="w-full max-w-6xl mx-auto px-5 sm:px-8 py-3.5 flex items-center justify-between gap-3">
        <div class="flex flex-wrap items-center gap-2.5 min-w-0">
          <!-- Online Status Pill -->
          <span class="login-status-pill inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-[10.5px] font-mono font-bold tracking-wide">
            <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
            <span>{{ $t('auth.systemOnline') || 'SYSTEM ONLINE' }}</span>
          </span>

          <!-- Compact Telemetry Mark + Platform Title -->
          <div class="flex items-center gap-2 min-w-0">
            <div class="login-brand-mark hidden sm:flex w-6 h-6 rounded-md items-center justify-center shrink-0">
              <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
            </div>
            <span class="login-platform-title text-xs sm:text-[13px] font-extrabold tracking-tight uppercase truncate">
              {{ $t('auth.platformHeaderTitle') || 'DATALOGGER TELEMETRY PLATFORM' }}
            </span>
          </div>

          <!-- Dynamic Mode Badge -->
          <span class="login-mode-badge px-2 py-0.5 rounded-md text-[10px] font-mono font-bold uppercase tracking-wider">
            {{ isCloudMode ? 'CLOUD' : 'EDGE' }}
          </span>
        </div>

        <!-- Language & Theme Controls -->
        <div class="flex items-center space-x-2 shrink-0">
          <!-- Segmented Language Switcher -->
          <button
            type="button"
            @click="toggleLanguage"
            class="login-top-control h-8 px-1.5 rounded-xl flex items-center gap-1 text-xs font-semibold transition-all cursor-pointer"
            :title="$t('header.toggleLang') || 'Switch Language'"
            aria-label="Switch Language"
          >
            <span
              class="px-2 py-0.5 rounded-lg text-[10.5px] font-mono font-bold transition-all"
              :class="currentLanguage === 'en' ? 'login-seg-active' : 'login-seg-idle'"
            >EN</span>
            <span class="login-seg-sep text-[10px]">|</span>
            <span
              class="px-2 py-0.5 rounded-lg text-[10.5px] font-mono font-bold transition-all"
              :class="currentLanguage === 'id' ? 'login-seg-active' : 'login-seg-idle'"
            >ID</span>
          </button>

          <!-- Clean Vector Day / Night Switcher -->
          <button
            type="button"
            @click="toggleTheme"
            class="login-top-control w-8 h-8 rounded-xl flex items-center justify-center transition-all cursor-pointer"
            :title="$t('header.toggleTheme') || 'Switch Theme'"
            aria-label="Switch Theme"
          >
            <!-- Moon Icon in Night Mode -->
            <svg v-if="isDarkMode" class="w-4 h-4 text-sky-300" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
            </svg>
            <!-- Sun Icon in Day Mode -->
            <svg v-else class="w-4 h-4 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <circle cx="12" cy="12" r="4" />
              <path stroke-linecap="round" d="M12 2v2m0 16v2M4.93 4.93l1.41 1.41m11.32 11.32l1.41 1.41M2 12h2m16 0h2M4.93 19.07l1.41-1.41m11.32-11.32l1.41-1.41" />
            </svg>
          </button>
        </div>
      </div>
    </header>

    <!-- Main Centered Authentication Card -->
    <main class="w-full max-w-[436px] mx-auto px-4 py-6 z-10 flex flex-col justify-center">
      <div class="login-card relative rounded-2xl overflow-hidden">
        <!-- Top Gradient Accent Bar -->
        <div class="h-1 w-full bg-gradient-to-r from-sky-500 via-blue-600 to-indigo-600"></div>

        <div class="px-6 py-7 sm:px-8 sm:py-8">
          <!-- Official Cakrawala Bima Instrument Brand Header -->
          <div class="flex flex-col items-center pb-5 mb-5 border-b login-divider">
            <img
              :src="isDarkMode ? cbiLogoNight : cbiLogoDay"
              alt="Cakrawala Bima Instrument"
              class="cbi-brand-logo h-11 sm:h-12 w-auto object-contain select-none"
              draggable="false"
            />
          </div>

          <!-- Access Control Title Block -->
          <div class="text-center mb-5">
            <div class="login-access-pill inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider mb-2.5">
              <span class="w-1.5 h-1.5 rounded-full bg-sky-500 animate-pulse"></span>
              <span>{{ $t('auth.systemAccess') || 'SYSTEM ACCESS CONTROL' }}</span>
            </div>
            <h1 class="login-heading text-lg sm:text-[19px] font-bold tracking-tight leading-snug">
              {{ $t('auth.signInTitle') || 'Sign In to Datalogger' }}
            </h1>
            <p class="login-subtitle text-xs mt-1 leading-relaxed max-w-[310px] mx-auto">
              {{ $t('auth.signInSubtitle') || 'Enter your operational credentials to access telemetry and monitoring systems' }}
            </p>
          </div>

          <!-- Error Feedback Banner -->
          <div
            v-if="errorMessage"
            class="login-error-banner mb-4 p-3 rounded-xl text-xs flex items-start gap-2.5 animate-shake"
          >
            <svg class="w-4 h-4 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" stroke-width="2" />
              <line x1="12" y1="8" x2="12" y2="12" stroke-width="2" />
              <line x1="12" y1="16" x2="12.01" y2="16" stroke-width="2" />
            </svg>
            <div class="flex-1 font-medium leading-relaxed">{{ errorMessage }}</div>
            <button type="button" @click="errorMessage = ''" class="opacity-80 hover:opacity-100 font-bold">&times;</button>
          </div>

          <!-- Credentials Form -->
          <form @submit.prevent="handleLogin" class="space-y-4">
            <!-- Username Input -->
            <div>
              <label for="login-username" class="login-label block text-[10.5px] font-bold uppercase tracking-wider mb-1.5">
                {{ $t('auth.username') || 'Username' }}
              </label>
              <div class="login-input-group relative">
                <span class="login-input-icon absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                  </svg>
                </span>
                <input
                  id="login-username"
                  type="text"
                  v-model="form.username"
                  required
                  autocomplete="username"
                  :placeholder="$t('auth.usernamePlaceholder') || 'Enter your registered username'"
                  class="login-field w-full pl-10 pr-4 py-2.5 rounded-xl text-xs font-mono transition-all"
                />
              </div>
            </div>

            <!-- Password Input -->
            <div>
              <label for="login-password" class="login-label block text-[10.5px] font-bold uppercase tracking-wider mb-1.5">
                {{ $t('auth.password') || 'Password' }}
              </label>
              <div class="login-input-group relative">
                <span class="login-input-icon absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
                  </svg>
                </span>
                <input
                  id="login-password"
                  :type="showPassword ? 'text' : 'password'"
                  v-model="form.password"
                  required
                  autocomplete="current-password"
                  :placeholder="$t('auth.passwordPlaceholder') || 'Enter your password'"
                  class="login-field w-full pl-10 pr-10 py-2.5 rounded-xl text-xs font-mono transition-all"
                />
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="login-eye-btn absolute inset-y-0 right-0 pr-3.5 flex items-center transition-colors cursor-pointer"
                  title="Toggle password visibility"
                >
                  <svg v-if="!showPassword" class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    <path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                  </svg>
                  <svg v-else class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l18 18" />
                  </svg>
                </button>
              </div>
            </div>

            <!-- Submit CTA -->
            <div class="pt-1.5">
              <button
                id="btn-login-submit"
                type="submit"
                :disabled="isLoading"
                class="login-primary-btn btn-text w-full py-2.5 px-4 font-bold text-xs rounded-xl transition-all transform active:scale-[0.99] disabled:opacity-55 disabled:pointer-events-none flex items-center justify-center space-x-2 cursor-pointer"
              >
                <svg v-if="isLoading" class="animate-spin -ml-1 mr-1.5 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
                </svg>
                <span class="btn-text">{{ isLoading ? ($t('auth.signingIn') || 'Authenticating...') : ($t('auth.signInButton') || 'Sign In') }}</span>
                <svg v-if="!isLoading" class="w-4 h-4 ml-1 btn-text" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.2">
                  <path stroke-linecap="round" stroke-linejoin="round" d="M14 5l7 7m0 0l-7 7m7-7H3" />
                </svg>
              </button>
            </div>
          </form>

          <!-- Minimal Production Security Note (No User Shortcuts) -->
          <div class="login-security-note mt-5 pt-4 border-t login-divider flex items-center justify-center gap-1.5 text-[11px]">
            <svg class="w-3.5 h-3.5 shrink-0 opacity-80" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
              <path d="M7 11V7a5 5 0 0110 0v4"></path>
            </svg>
            <span>{{ $t('auth.securitySessionNote') || 'Authorized RBAC Session • Encrypted Telemetry Access' }}</span>
          </div>
        </div>
      </div>
    </main>

    <!-- Official Brand & System Footer -->
    <footer class="w-full max-w-6xl mx-auto px-6 pb-5 pt-2 text-center z-10">
      <div class="login-footer-text inline-flex flex-wrap items-center justify-center gap-2 text-[11px]">
        <span class="font-semibold">&copy; Cakrawala Bima Instrument</span>
        <span class="opacity-50">&bull;</span>
        <span>{{ $t('auth.platformSubtitle') || 'Industrial Telemetry & Environmental Monitoring Platform' }}</span>
      </div>
    </footer>
  </div>
</template>

<script>
import cbiLogoDay from '../assets/cbi-logo-day.png';
import cbiLogoNight from '../assets/cbi-logo-night.png';

export default {
  name: 'LoginView',
  data() {
    return {
      cbiLogoDay,
      cbiLogoNight,
      domDark: false,
      themeObserver: null,
      form: {
        username: '',
        password: '',
      },
      showPassword: false,
      isLoading: false,
      errorMessage: '',
    };
  },
  computed: {
    isDarkMode() {
      if (this.$store && this.$store.state) {
        if (this.$store.state.uiPreferences && this.$store.state.uiPreferences.effectiveTheme) {
          return this.$store.state.uiPreferences.effectiveTheme === 'dark';
        }
        if (this.$store.state.theme) {
          return this.$store.state.theme === 'dark';
        }
      }
      return this.domDark;
    },
    isCloudMode() {
      const sys = (this.$store && this.$store.state && this.$store.state.systemStatus) || {};
      return Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
    },
    currentLanguage() {
      return this.$store.state.uiPreferences?.language || 'id';
    },
  },
  mounted() {
    this.syncDomDark();
    if (typeof MutationObserver !== 'undefined' && typeof document !== 'undefined') {
      this.themeObserver = new MutationObserver(() => this.syncDomDark());
      this.themeObserver.observe(document.documentElement, {
        attributes: true,
        attributeFilter: ['class'],
      });
    }
  },
  beforeDestroy() {
    if (this.themeObserver) {
      this.themeObserver.disconnect();
      this.themeObserver = null;
    }
  },
  methods: {
    syncDomDark() {
      if (typeof document !== 'undefined' && document.documentElement) {
        this.domDark = document.documentElement.classList.contains('dark');
      }
    },
    toggleLanguage() {
      const next = this.currentLanguage === 'id' ? 'en' : 'id';
      this.$store.dispatch('setLanguage', next);
    },
    toggleTheme() {
      const next = this.isDarkMode ? 'light' : 'dark';
      this.$store.dispatch('setTheme', next);
      this.syncDomDark();
    },
    async handleLogin() {
      this.isLoading = true;
      this.errorMessage = '';
      try {
        await this.$store.dispatch('login', {
          username: this.form.username.trim(),
          password: this.form.password,
        });

        // Determine destination redirect
        const redirect = this.$route.query.redirect;
        const user = this.$store.state.user || {};
        const role = (user.role || '').toUpperCase();

        const isSuper = role === 'SUPERADMIN' || role === 'ADMINISTRATOR';
        const isAdmin = role === 'ADMIN';
        const isEngineer = role === 'ENGINEER';
        const isOperator = role === 'OPERATOR';
        const sys = (this.$store && this.$store.state && this.$store.state.systemStatus) || {};
        const isCloudMode = Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
        const defaultLanding = isSuper
          ? '/dashboard'
          : ((isOperator || ((isAdmin || isEngineer) && isCloudMode)) ? '/monitoring/dashboard-operator' : '/monitoring/operator');

        if (redirect && redirect !== '/login') {
          // Check if destination is restricted for the logged-in role
          const isDevRoute = redirect.startsWith('/development');
          const isUsersRoute = redirect.startsWith('/administration/users');
          const isRolesRoute = redirect.startsWith('/administration/roles');
          const isTenantsRoute = redirect.startsWith('/administration/tenants');
          const isQualityRoute = redirect === '/dashboard' || redirect.startsWith('/dashboard') || redirect === '/quality' || redirect.startsWith('/monitoring/quality');
          const isCloudAdminRestrictedRoute =
            isCloudMode && (
              redirect === '/' ||
              redirect === '/operator' ||
              redirect.startsWith('/monitoring/operator') ||
              redirect.startsWith('/alarm/live-comm') ||
              redirect.startsWith('/administration/config') ||
              redirect.startsWith('/administration/backup') ||
              redirect.startsWith('/administration/retention')
            );
          const isOperatorRestrictedRoute =
            redirect === '/' ||
            redirect === '/operator' ||
            redirect.startsWith('/monitoring/operator') ||
            redirect.startsWith('/monitoring/telemetry') ||
            redirect.startsWith('/monitoring/reports') ||
            redirect.startsWith('/alarm/live-comm') ||
            redirect.startsWith('/system/logs') ||
            redirect.startsWith('/system/audit') ||
            redirect.startsWith('/system/communication') ||
            redirect.startsWith('/administration') ||
            redirect === '/integration' ||
            (isCloudMode && (
              redirect === '/monitoring/system' ||
              redirect.startsWith('/monitoring/system/') ||
              redirect === '/integration/queue' ||
              redirect.startsWith('/integration/')
            )) ||
            redirect.startsWith('/integration/destinations') ||
            redirect === '/destinations';

          if (isOperator && isOperatorRestrictedRoute) {
            this.$router.push('/monitoring/dashboard-operator');
            return;
          }
          if ((isAdmin || isEngineer) && isCloudAdminRestrictedRoute) {
            this.$router.push('/monitoring/dashboard-operator');
            return;
          }
          if (isTenantsRoute && !isSuper && !isAdmin && !(isEngineer && isCloudMode)) {
            this.$router.push(defaultLanding);
            return;
          }
          if (isQualityRoute && !isSuper) {
            this.$router.push(defaultLanding);
            return;
          }
          if (isDevRoute && !isSuper) {
            this.$router.push(defaultLanding);
            return;
          }
          if (isUsersRoute && !isSuper && role !== 'ADMIN') {
            this.$router.push(defaultLanding);
            return;
          }
          if (isRolesRoute && !isSuper) {
            this.$router.push(defaultLanding);
            return;
          }

          this.$router.push(redirect);
        } else {
          this.$router.push(defaultLanding);
        }
      } catch (err) {
        console.error('Login error:', err);
        const backendErr = err.response?.data?.error || err.message;
        if (err.response?.status === 401 || (backendErr && backendErr.toLowerCase().includes('invalid username'))) {
          this.errorMessage = this.$t('auth.invalidCredentials') || 'Invalid username or password. Please verify your credentials.';
        } else {
          this.errorMessage = backendErr || 'Authentication failed. Please check network connection.';
        }
      } finally {
        this.isLoading = false;
      }
    },
  },
};
</script>

<style scoped>
/* =========================================================================
   DAY MODE (MODE SIANG — CLEAN ENTERPRISE LIGHT)
   ========================================================================= */
.login-shell {
  background: linear-gradient(145deg, #F8FAFC 0%, #EFF6FF 50%, #F1F5F9 100%);
  color: #0F172A;
}

.login-topbar {
  background: rgba(255, 255, 255, 0.72);
  border-bottom: 1px solid rgba(203, 213, 225, 0.75);
  backdrop-filter: blur(10px);
}

.login-status-pill {
  background: rgba(16, 185, 129, 0.12);
  color: #047857;
  border: 1px solid rgba(16, 185, 129, 0.28);
}

.login-platform-title {
  color: #0F172A;
}

.login-mode-badge {
  background: #F1F5F9;
  color: #334155;
  border: 1px solid #CBD5E1;
}

.login-grid-pattern {
  background-image:
    linear-gradient(to right, rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(15, 23, 42, 0.035) 1px, transparent 1px);
  background-size: 28px 28px;
  mask-image: radial-gradient(ellipse at center, rgba(0, 0, 0, 0.85) 25%, transparent 80%);
  -webkit-mask-image: radial-gradient(ellipse at center, rgba(0, 0, 0, 0.85) 25%, transparent 80%);
}

.login-orb {
  position: absolute;
  width: 28rem;
  height: 28rem;
  border-radius: 9999px;
  filter: blur(76px);
}

.login-orb-top {
  top: -10rem;
  right: -8rem;
  background: rgba(14, 165, 233, 0.13);
}

.login-orb-bottom {
  bottom: -10rem;
  left: -8rem;
  background: rgba(37, 99, 235, 0.11);
}

.login-brand-mark {
  background: linear-gradient(135deg, #0284C7 0%, #2563EB 100%);
  box-shadow: 0 4px 10px -2px rgba(2, 132, 199, 0.35);
  color: #FFFFFF !important;
}

.login-top-control {
  background: rgba(255, 255, 255, 0.94);
  border: 1px solid #CBD5E1;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
}

.login-top-control:hover {
  border-color: #3B82F6;
  background: #FFFFFF;
}

.login-seg-active {
  background: #EFF6FF;
  color: #1D4ED8;
  box-shadow: inset 0 0 0 1px #BFDBFE;
}

.login-seg-idle {
  color: #64748B;
}

.login-seg-sep {
  color: #CBD5E1;
}

.login-card {
  background: rgba(255, 255, 255, 0.98);
  border: 1px solid #E2E8F0;
  box-shadow:
    0 24px 48px -12px rgba(15, 23, 42, 0.11),
    0 4px 12px -2px rgba(15, 23, 42, 0.05);
  backdrop-filter: blur(12px);
}

.login-divider {
  border-color: #E2E8F0;
}

.login-access-pill {
  background: #EFF6FF;
  color: #1D4ED8;
  border: 1px solid #BFDBFE;
}

.login-heading {
  color: #0F172A !important;
}

.login-subtitle {
  color: #64748B;
}

.login-label {
  color: #334155;
}

.login-error-banner {
  background: #FFF1F2;
  border: 1px solid #FECDD3;
  color: #BE123C;
}

.login-input-icon {
  color: #64748B;
  transition: color 0.15s ease;
}

.login-input-group:focus-within .login-input-icon {
  color: #0284C7;
}

.login-eye-btn {
  color: #64748B;
}

.login-eye-btn:hover {
  color: #0F172A;
}

.login-field {
  background-color: #FFFFFF !important;
  border: 1px solid #CBD5E1 !important;
  color: #0F172A !important;
}

.login-field::placeholder {
  color: #94A3B8 !important;
}

.login-field:focus {
  background-color: #FFFFFF !important;
  border-color: #0284C7 !important;
  box-shadow: 0 0 0 3px rgba(2, 132, 199, 0.15) !important;
  outline: none;
}

.login-light .login-field:-webkit-autofill,
.login-light .login-field:-webkit-autofill:hover,
.login-light .login-field:-webkit-autofill:focus,
.login-light .login-field:-webkit-autofill:active {
  -webkit-text-fill-color: #0F172A !important;
  -webkit-box-shadow: 0 0 0 1000px #FFFFFF inset !important;
  box-shadow: 0 0 0 1000px #FFFFFF inset !important;
  border: 1px solid #CBD5E1 !important;
  caret-color: #0F172A;
  transition: background-color 5000s ease-in-out 0s;
}

.login-primary-btn {
  background: linear-gradient(90deg, #0284C7 0%, #2563EB 100%);
  color: #FFFFFF !important;
  box-shadow:
    0 10px 20px -6px rgba(37, 99, 235, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.login-primary-btn:hover {
  background: linear-gradient(90deg, #0369A1 0%, #1D4ED8 100%);
  box-shadow:
    0 12px 24px -6px rgba(29, 78, 216, 0.52),
    inset 0 1px 0 rgba(255, 255, 255, 0.25);
}

.login-security-note {
  color: #64748B;
}

.login-footer-text {
  color: #64748B;
}

/* =========================================================================
   NIGHT MODE (MODE MALAM — HIGH-CONTRAST INDUSTRIAL DARK)
   ========================================================================= */
.login-shell.login-dark {
  background: radial-gradient(circle at 50% 0%, #0D1B30 0%, #070D19 58%, #040811 100%);
  color: #F8FAFC;
}

.login-dark .login-topbar {
  background: rgba(7, 13, 25, 0.78);
  border-bottom-color: rgba(30, 41, 59, 0.85);
}

.login-dark .login-status-pill {
  background: rgba(16, 185, 129, 0.14);
  color: #34D399;
  border-color: rgba(16, 185, 129, 0.32);
}

.login-dark .login-platform-title {
  color: #F8FAFC;
}

.login-dark .login-mode-badge {
  background: rgba(15, 23, 42, 0.9);
  color: #CBD5E1;
  border-color: #334155;
}

.login-dark .login-grid-pattern {
  background-image:
    linear-gradient(to right, rgba(56, 189, 248, 0.045) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(56, 189, 248, 0.045) 1px, transparent 1px);
}

.login-dark .login-orb-top {
  background: rgba(14, 165, 233, 0.16);
}

.login-dark .login-orb-bottom {
  background: rgba(37, 99, 235, 0.16);
}

.login-dark .login-top-control {
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(51, 65, 85, 0.9);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.45);
}

.login-dark .login-top-control:hover {
  border-color: #38BDF8;
  background: #0F172A;
}

.login-dark .login-seg-active {
  background: rgba(56, 189, 248, 0.18);
  color: #38BDF8;
  box-shadow: inset 0 0 0 1px rgba(56, 189, 248, 0.35);
}

.login-dark .login-seg-idle {
  color: #94A3B8;
}

.login-dark .login-seg-sep {
  color: #334155;
}

.login-dark .login-card {
  background: rgba(15, 23, 42, 0.96);
  border: 1px solid rgba(51, 65, 85, 0.85);
  box-shadow:
    0 28px 56px -12px rgba(0, 0, 0, 0.65),
    0 0 0 1px rgba(56, 189, 248, 0.08);
}

.login-dark .login-divider {
  border-color: rgba(51, 65, 85, 0.75);
}

.login-dark .login-access-pill {
  background: rgba(14, 165, 233, 0.14);
  color: #38BDF8;
  border-color: rgba(56, 189, 248, 0.3);
}

.login-dark .login-heading {
  color: #F8FAFC !important;
}

.login-dark .login-subtitle {
  color: #94A3B8;
}

.login-dark .login-label {
  color: #CBD5E1;
}

.login-dark .login-error-banner {
  background: rgba(244, 63, 94, 0.12);
  border-color: rgba(244, 63, 94, 0.32);
  color: #FDA4AF;
}

.login-dark .login-input-icon {
  color: #94A3B8;
}

.login-dark .login-input-group:focus-within .login-input-icon {
  color: #38BDF8;
}

.login-dark .login-eye-btn {
  color: #94A3B8;
}

.login-dark .login-eye-btn:hover {
  color: #F8FAFC;
}

.login-dark .login-field {
  background-color: #070D18 !important;
  border: 1px solid #1E293B !important;
  color: #F8FAFC !important;
}

.login-dark .login-field::placeholder {
  color: #64748B !important;
}

.login-dark .login-field:focus {
  background-color: #050912 !important;
  border-color: #38BDF8 !important;
  box-shadow: 0 0 0 3px rgba(56, 189, 248, 0.18) !important;
  outline: none;
}

.login-dark .login-field:-webkit-autofill,
.login-dark .login-field:-webkit-autofill:hover,
.login-dark .login-field:-webkit-autofill:focus,
.login-dark .login-field:-webkit-autofill:active {
  -webkit-text-fill-color: #F8FAFC !important;
  -webkit-box-shadow: 0 0 0 1000px #070D18 inset !important;
  box-shadow: 0 0 0 1000px #070D18 inset !important;
  border: 1px solid #1E293B !important;
  caret-color: #F8FAFC;
  transition: background-color 5000s ease-in-out 0s;
}

.login-dark .login-primary-btn {
  background: linear-gradient(90deg, #0284C7 0%, #2563EB 100%);
  box-shadow:
    0 10px 24px -6px rgba(2, 132, 199, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.login-dark .login-primary-btn:hover {
  background: linear-gradient(90deg, #0EA5E9 0%, #3B82F6 100%);
}

.login-dark .login-security-note {
  color: #94A3B8;
}

.login-dark .login-footer-text {
  color: #94A3B8;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  20%, 60% { transform: translateX(-4px); }
  40%, 80% { transform: translateX(4px); }
}
.animate-shake {
  animation: shake 0.4s ease-in-out;
}
</style>
