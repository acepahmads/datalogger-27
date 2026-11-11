<template>
  <div
    class="login-shell min-h-screen w-full flex flex-col justify-between relative overflow-hidden font-sans transition-colors duration-200"
    :class="isDarkMode ? 'login-dark' : 'login-light'"
  >
    <!-- Architectural Engineering Grid & Ambient Glows -->
    <div class="login-grid-pattern absolute inset-0 pointer-events-none"></div>
    <div class="login-orb login-orb-top pointer-events-none"></div>
    <div class="login-orb login-orb-bottom pointer-events-none"></div>

    <!-- Top Navigation Bar for Brand Identity, Language & Theme -->
    <header class="w-full max-w-6xl mx-auto px-5 sm:px-8 pt-5 pb-3 flex items-center justify-between z-10">
      <div class="flex items-center space-x-3">
        <div class="login-brand-mark w-10 h-10 rounded-xl flex items-center justify-center text-white shrink-0">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2.1">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
        </div>
        <div>
          <div class="text-xs font-extrabold tracking-tight uppercase text-slate-900 dark:text-white flex items-center gap-2">
            <span>DATALOGGER-27</span>
            <span class="px-2 py-0.5 rounded-md text-4xs font-mono font-bold bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-500/15 dark:text-sky-300 dark:border-blue-500/30">
              EDGE &bull; CBI
            </span>
          </div>
          <div class="text-3xs font-medium text-slate-600 dark:text-slate-400 mt-0.5">
            {{ $t('auth.platformSubtitle') || 'Industrial Telemetry & Edge Monitoring Platform' }}
          </div>
        </div>
      </div>

      <!-- Language & Theme Controls -->
      <div class="flex items-center space-x-2">
        <!-- Segmented Language Switcher -->
        <button
          type="button"
          @click="toggleLanguage"
          class="login-top-control h-8 px-1.5 rounded-xl flex items-center gap-1.5 text-xs font-semibold transition-all cursor-pointer"
          :title="$t('header.toggleLang') || 'Switch Language'"
          aria-label="Switch Language"
        >
          <span
            class="px-2 py-0.5 rounded-lg text-3xs font-mono font-bold transition-all"
            :class="currentLanguage === 'en' ? 'login-seg-active' : 'text-slate-500 dark:text-slate-400'"
          >EN</span>
          <span
            class="px-2 py-0.5 rounded-lg text-3xs font-mono font-bold transition-all"
            :class="currentLanguage === 'id' ? 'login-seg-active' : 'text-slate-500 dark:text-slate-400'"
          >ID</span>
        </button>

        <!-- Clean Vector Theme Switcher -->
        <button
          type="button"
          @click="toggleTheme"
          class="login-top-control w-8 h-8 rounded-xl flex items-center justify-center transition-all cursor-pointer"
          :title="$t('header.toggleTheme') || 'Switch Theme'"
          aria-label="Switch Theme"
        >
          <!-- Moon Icon when in Dark Mode (click to switch to Light) -->
          <svg v-if="isDarkMode" class="w-4 h-4 text-sky-300" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
          </svg>
          <!-- Sun Icon when in Light Mode (click to switch to Dark) -->
          <svg v-else class="w-4 h-4 text-amber-500" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <circle cx="12" cy="12" r="4" />
            <path stroke-linecap="round" d="M12 2v2m0 16v2M4.93 4.93l1.41 1.41m11.32 11.32l1.41 1.41M2 12h2m16 0h2M4.93 19.07l1.41-1.41m11.32-11.32l1.41-1.41" />
          </svg>
        </button>
      </div>
    </header>

    <!-- Main Form Center Card -->
    <main class="w-full max-w-[440px] mx-auto px-4 py-6 z-10 flex flex-col justify-center">
      <div class="login-card relative rounded-3xl overflow-hidden">
        <!-- Top Accent Hairline -->
        <div class="h-1 w-full bg-gradient-to-r from-blue-600 via-sky-500 to-indigo-600"></div>

        <div class="p-6 sm:p-8">
          <!-- Header in Card -->
          <div class="text-center mb-5">
            <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-4xs font-bold uppercase tracking-wider bg-blue-50 text-blue-700 border border-blue-200 dark:bg-blue-500/12 dark:text-sky-300 dark:border-blue-500/25 mb-2.5">
              <span class="w-1.5 h-1.5 rounded-full bg-blue-600 dark:bg-sky-400 animate-pulse"></span>
              <span>{{ $t('auth.systemAccess') || 'SYSTEM ACCESS CONTROL' }}</span>
            </div>
            <h1 class="text-base sm:text-[17px] font-bold text-slate-900 dark:text-white tracking-tight leading-snug">
              {{ $t('auth.signInTitle') || 'Sign In to Datalogger' }}
            </h1>
            <p class="text-2xs text-slate-500 dark:text-slate-400 mt-1 leading-relaxed max-w-[290px] mx-auto">
              {{ $t('auth.signInSubtitle') || 'Enter your operational credentials to access telemetry and monitoring systems' }}
            </p>
          </div>

          <!-- Error Feedback Toast -->
          <div
            v-if="errorMessage"
            class="mb-5 p-3 rounded-xl bg-rose-50 border border-rose-200 text-rose-700 dark:bg-rose-500/10 dark:border-rose-500/30 dark:text-rose-300 text-xs flex items-start gap-2.5 animate-shake"
          >
            <svg class="w-4 h-4 flex-shrink-0 mt-0.5 text-rose-600 dark:text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <circle cx="12" cy="12" r="10" stroke-width="2" />
              <line x1="12" y1="8" x2="12" y2="12" stroke-width="2" />
              <line x1="12" y1="16" x2="12.01" y2="16" stroke-width="2" />
            </svg>
            <div class="flex-1 font-medium leading-relaxed">{{ errorMessage }}</div>
            <button type="button" @click="errorMessage = ''" class="text-rose-500 hover:text-rose-700 dark:text-rose-400 dark:hover:text-rose-200 font-bold">&times;</button>
          </div>

          <!-- Credentials Form -->
          <form @submit.prevent="handleLogin" class="space-y-4">
            <!-- Username Input -->
            <div>
              <label for="login-username" class="block text-3xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 mb-1.5">
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
                  :placeholder="$t('auth.usernamePlaceholder') || 'e.g. admin or operator_user'"
                  class="login-field w-full pl-10 pr-4 py-2.5 rounded-xl text-xs font-mono transition-all"
                />
              </div>
            </div>

            <!-- Password Input -->
            <div>
              <label for="login-password" class="block text-3xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300 mb-1.5">
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
                  :placeholder="$t('auth.passwordPlaceholder') || 'Enter password'"
                  class="login-field w-full pl-10 pr-10 py-2.5 rounded-xl text-xs font-mono transition-all"
                />
                <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="absolute inset-y-0 right-0 pr-3.5 flex items-center text-slate-400 hover:text-slate-700 dark:hover:text-slate-200 transition-colors cursor-pointer"
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

            <!-- Submit Button -->
            <div class="pt-2">
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

          <!-- Development & Testing Quick Fill Accounts Strip -->
          <div class="mt-6 pt-5 border-t border-slate-200 dark:border-slate-800/80">
            <div class="flex items-center gap-2 mb-3">
              <div class="h-px flex-1 bg-slate-200/80 dark:bg-slate-800"></div>
              <span class="text-4xs font-mono font-bold uppercase tracking-wider text-slate-500 dark:text-slate-400">
                {{ $t('auth.testAccounts') || 'Testing & Demo Accounts (1-Click Fill)' }}
              </span>
              <div class="h-px flex-1 bg-slate-200/80 dark:bg-slate-800"></div>
            </div>

            <div class="grid grid-cols-2 gap-2 text-2xs">
              <!-- 1. Superadmin -->
              <button
                type="button"
                @click="fillAccount('admin', 'Password123!')"
                class="login-role-card p-2.5 rounded-xl text-left transition-all group cursor-pointer"
                :class="form.username === 'admin' ? 'login-role-active-purple' : ''"
              >
                <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-purple-700 dark:group-hover:text-purple-300 flex items-center justify-between">
                  <span class="flex items-center gap-1.5">
                    <span class="w-2 h-2 rounded-full bg-purple-600 dark:bg-purple-400"></span>
                    <span>SUPERADMIN</span>
                  </span>
                  <span v-if="form.username === 'admin'" class="text-purple-600 dark:text-purple-400 text-3xs font-mono">✓</span>
                </div>
                <div class="text-4xs font-mono text-slate-500 dark:text-slate-400 mt-0.5">@admin &bull; Full Access</div>
              </button>

              <!-- 2. Admin -->
              <button
                type="button"
                @click="fillAccount('admin_user', 'Password123!')"
                class="login-role-card p-2.5 rounded-xl text-left transition-all group cursor-pointer"
                :class="form.username === 'admin_user' ? 'login-role-active-blue' : ''"
              >
                <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-blue-700 dark:group-hover:text-blue-300 flex items-center justify-between">
                  <span class="flex items-center gap-1.5">
                    <span class="w-2 h-2 rounded-full bg-blue-600 dark:bg-blue-400"></span>
                    <span>ADMIN</span>
                  </span>
                  <span v-if="form.username === 'admin_user'" class="text-blue-600 dark:text-blue-400 text-3xs font-mono">✓</span>
                </div>
                <div class="text-4xs font-mono text-slate-500 dark:text-slate-400 mt-0.5">@admin_user &bull; Ops Admin</div>
              </button>

              <!-- 3. Engineer -->
              <button
                type="button"
                @click="fillAccount('engineer_user', 'Password123!')"
                class="login-role-card p-2.5 rounded-xl text-left transition-all group cursor-pointer"
                :class="form.username === 'engineer_user' ? 'login-role-active-amber' : ''"
              >
                <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-amber-700 dark:group-hover:text-amber-300 flex items-center justify-between">
                  <span class="flex items-center gap-1.5">
                    <span class="w-2 h-2 rounded-full bg-amber-500 dark:bg-amber-400"></span>
                    <span>ENGINEER</span>
                  </span>
                  <span v-if="form.username === 'engineer_user'" class="text-amber-600 dark:text-amber-400 text-3xs font-mono">✓</span>
                </div>
                <div class="text-4xs font-mono text-slate-500 dark:text-slate-400 mt-0.5">@engineer_user &bull; Sensors</div>
              </button>

              <!-- 4. Operator -->
              <button
                type="button"
                @click="fillAccount('operator_user', 'Password123!')"
                class="login-role-card p-2.5 rounded-xl text-left transition-all group cursor-pointer"
                :class="form.username === 'operator_user' ? 'login-role-active-emerald' : ''"
              >
                <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-emerald-700 dark:group-hover:text-emerald-300 flex items-center justify-between">
                  <span class="flex items-center gap-1.5">
                    <span class="w-2 h-2 rounded-full bg-emerald-600 dark:bg-emerald-400"></span>
                    <span>OPERATOR</span>
                  </span>
                  <span v-if="form.username === 'operator_user'" class="text-emerald-600 dark:text-emerald-400 text-3xs font-mono">✓</span>
                </div>
                <div class="text-4xs font-mono text-slate-500 dark:text-slate-400 mt-0.5">@operator_user &bull; Dashboard</div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Minimal System Footer -->
    <footer class="w-full max-w-6xl mx-auto px-6 pb-5 pt-2 text-center z-10">
      <div class="inline-flex flex-wrap items-center justify-center gap-2.5 text-3xs font-mono text-slate-500 dark:text-slate-400">
        <span class="inline-flex items-center gap-1.5">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
          <span class="font-semibold text-slate-600 dark:text-slate-300">Datalogger-27 Edge v1.0.0</span>
        </span>
        <span class="text-slate-300 dark:text-slate-700">&bull;</span>
        <span>{{ $t('auth.footerSystemLabel') || 'Autonomous Edge Telemetry Gateway' }}</span>
      </div>
    </footer>
  </div>
</template>

<script>
export default {
  name: 'LoginView',
  data() {
    return {
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
      return typeof document !== 'undefined' && document.documentElement.classList.contains('dark');
    },
    currentLanguage() {
      return this.$store.state.uiPreferences?.language || 'id';
    },
  },
  methods: {
    toggleLanguage() {
      const next = this.currentLanguage === 'id' ? 'en' : 'id';
      this.$store.dispatch('setLanguage', next);
    },
    toggleTheme() {
      const next = this.isDarkMode ? 'light' : 'dark';
      this.$store.dispatch('setTheme', next);
    },
    fillAccount(username, password) {
      this.form.username = username;
      this.form.password = password;
      this.errorMessage = '';
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
        const isOperator = role === 'OPERATOR';
        const defaultLanding = isSuper
          ? '/dashboard'
          : (isOperator ? '/monitoring/dashboard-operator' : '/monitoring/operator');

        if (redirect && redirect !== '/login') {
          // Check if destination is restricted for the logged-in role
          const isDevRoute = redirect.startsWith('/development');
          const isUsersRoute = redirect.startsWith('/administration/users');
          const isRolesRoute = redirect.startsWith('/administration/roles');
          const isQualityRoute = redirect === '/dashboard' || redirect.startsWith('/dashboard') || redirect === '/quality' || redirect.startsWith('/monitoring/quality');
          const sys = (this.$store && this.$store.state && this.$store.state.systemStatus) || {};
          const isCloudMode = Boolean(sys.cloud_mode || String(sys.monitoring_domain || '').toUpperCase() === 'CLOUD');
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
/* Shell Background (Light & Dark — Vue 2 scoped compatible via .login-dark / .login-light) */
.login-shell {
  background: linear-gradient(145deg, #F8FAFC 0%, #EFF6FF 48%, #F1F5F9 100%);
  color: #0F172A;
}

.login-shell.login-dark {
  background: radial-gradient(circle at 50% 0%, #0E1B33 0%, #070D18 58%, #050912 100%);
  color: #F8FAFC;
}

/* Precision Architectural Grid Overlay */
.login-grid-pattern {
  background-image:
    linear-gradient(to right, rgba(15, 23, 42, 0.035) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(15, 23, 42, 0.035) 1px, transparent 1px);
  background-size: 32px 32px;
  mask-image: radial-gradient(ellipse at center, rgba(0, 0, 0, 0.85) 25%, transparent 80%);
  -webkit-mask-image: radial-gradient(ellipse at center, rgba(0, 0, 0, 0.85) 25%, transparent 80%);
}

.login-dark .login-grid-pattern {
  background-image:
    linear-gradient(to right, rgba(56, 189, 248, 0.045) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(56, 189, 248, 0.045) 1px, transparent 1px);
}

/* Soft Ambient Orbs */
.login-orb {
  position: absolute;
  width: 28rem;
  height: 28rem;
  border-radius: 9999px;
  filter: blur(72px);
}

.login-orb-top {
  top: -10rem;
  right: -8rem;
  background: rgba(37, 99, 235, 0.12);
}

.login-orb-bottom {
  bottom: -10rem;
  left: -8rem;
  background: rgba(14, 165, 233, 0.11);
}

.login-dark .login-orb-top {
  background: rgba(37, 99, 235, 0.18);
}

.login-dark .login-orb-bottom {
  background: rgba(6, 182, 212, 0.13);
}

/* Brand Mark */
.login-brand-mark {
  background: linear-gradient(135deg, #1D4ED8 0%, #0284C7 100%);
  box-shadow: 0 8px 18px -4px rgba(29, 78, 216, 0.35), inset 0 1px 0 rgba(255, 255, 255, 0.25);
  color: #FFFFFF !important;
}

/* Top Controls */
.login-top-control {
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid #CBD5E1;
  box-shadow: 0 1px 2px rgba(15, 23, 42, 0.05);
}

.login-top-control:hover {
  border-color: #3B82F6;
  background: #FFFFFF;
}

.login-dark .login-top-control {
  background: rgba(15, 23, 42, 0.88);
  border: 1px solid rgba(51, 65, 85, 0.9);
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.45);
}

.login-dark .login-top-control:hover {
  border-color: #38BDF8;
  background: #0F172A;
}

.login-seg-active {
  background: #EFF6FF;
  color: #1D4ED8;
  box-shadow: inset 0 0 0 1px #BFDBFE;
}

.login-dark .login-seg-active {
  background: rgba(56, 189, 248, 0.18);
  color: #38BDF8;
  box-shadow: inset 0 0 0 1px rgba(56, 189, 248, 0.35);
}

/* Main Login Card */
.login-card {
  background: rgba(255, 255, 255, 0.97);
  border: 1px solid #CBD5E1;
  box-shadow:
    0 24px 48px -12px rgba(15, 23, 42, 0.10),
    0 4px 12px -2px rgba(15, 23, 42, 0.05);
  backdrop-filter: blur(12px);
}

.login-dark .login-card {
  background: rgba(13, 21, 38, 0.95);
  border: 1px solid rgba(51, 65, 85, 0.85);
  box-shadow:
    0 28px 56px -12px rgba(0, 0, 0, 0.6),
    0 0 0 1px rgba(56, 189, 248, 0.08);
}

/* Input Group & Focus Highlighting */
.login-input-icon {
  color: #64748B;
  transition: color 0.15s ease;
}

.login-input-group:focus-within .login-input-icon {
  color: #2563EB;
}

.login-dark .login-input-icon {
  color: #94A3B8;
}

.login-dark .login-input-group:focus-within .login-input-icon {
  color: #38BDF8;
}

.login-field {
  background-color: #F8FAFC !important;
  border: 1px solid #CBD5E1 !important;
  color: #0F172A !important;
}

.login-field:focus {
  background-color: #FFFFFF !important;
  border-color: #2563EB !important;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.14) !important;
  outline: none;
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

/* Browser Autofill Contrast Preservation (Light & Dark) */
.login-light .login-field:-webkit-autofill,
.login-light .login-field:-webkit-autofill:hover,
.login-light .login-field:-webkit-autofill:focus,
.login-light .login-field:-webkit-autofill:active {
  -webkit-text-fill-color: #0F172A !important;
  -webkit-box-shadow: 0 0 0 1000px #F8FAFC inset !important;
  box-shadow: 0 0 0 1000px #F8FAFC inset !important;
  border: 1px solid #CBD5E1 !important;
  caret-color: #0F172A;
  transition: background-color 5000s ease-in-out 0s;
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

/* Primary Sign-In Button */
.login-primary-btn {
  background: linear-gradient(135deg, #1D4ED8 0%, #2563EB 55%, #0284C7 100%);
  color: #FFFFFF !important;
  box-shadow:
    0 10px 20px -6px rgba(37, 99, 235, 0.42),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.login-primary-btn:hover {
  background: linear-gradient(135deg, #1E40AF 0%, #1D4ED8 55%, #0369A1 100%);
  box-shadow:
    0 12px 24px -6px rgba(29, 78, 216, 0.52),
    inset 0 1px 0 rgba(255, 255, 255, 0.25);
}

.login-dark .login-primary-btn {
  background: linear-gradient(135deg, #2563EB 0%, #0284C7 100%);
  box-shadow:
    0 10px 24px -6px rgba(2, 132, 199, 0.4),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.login-dark .login-primary-btn:hover {
  background: linear-gradient(135deg, #3B82F6 0%, #0EA5E9 100%);
}

/* Quick-Fill Role Cards */
.login-role-card {
  background: #F8FAFC;
  border: 1px solid #E2E8F0;
}

.login-role-card:hover {
  background: #FFFFFF;
  border-color: #CBD5E1;
  box-shadow: 0 4px 10px -2px rgba(15, 23, 42, 0.06);
}

.login-dark .login-role-card {
  background: rgba(15, 23, 42, 0.72);
  border: 1px solid rgba(51, 65, 85, 0.78);
}

.login-dark .login-role-card:hover {
  background: rgba(30, 41, 59, 0.88);
  border-color: rgba(71, 85, 105, 0.95);
}

.login-role-active-purple {
  background: #FAF5FF !important;
  border-color: #C084FC !important;
  box-shadow: 0 0 0 1px rgba(147, 51, 234, 0.15);
}
.login-dark .login-role-active-purple {
  background: rgba(147, 51, 234, 0.16) !important;
  border-color: rgba(168, 85, 247, 0.6) !important;
}

.login-role-active-blue {
  background: #EFF6FF !important;
  border-color: #60A5FA !important;
  box-shadow: 0 0 0 1px rgba(37, 99, 235, 0.15);
}
.login-dark .login-role-active-blue {
  background: rgba(37, 99, 235, 0.18) !important;
  border-color: rgba(59, 130, 246, 0.6) !important;
}

.login-role-active-amber {
  background: #FFFBEB !important;
  border-color: #FBBF24 !important;
  box-shadow: 0 0 0 1px rgba(217, 119, 6, 0.15);
}
.login-dark .login-role-active-amber {
  background: rgba(245, 158, 11, 0.16) !important;
  border-color: rgba(251, 191, 36, 0.6) !important;
}

.login-role-active-emerald {
  background: #ECFDF5 !important;
  border-color: #34D399 !important;
  box-shadow: 0 0 0 1px rgba(5, 150, 105, 0.15);
}
.login-dark .login-role-active-emerald {
  background: rgba(16, 185, 129, 0.16) !important;
  border-color: rgba(52, 211, 153, 0.6) !important;
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
