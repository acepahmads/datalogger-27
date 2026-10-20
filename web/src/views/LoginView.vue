<template>
  <div class="min-h-screen w-full flex flex-col justify-between bg-slate-100 dark:bg-[#070D18] text-slate-800 dark:text-slate-100 relative overflow-hidden font-sans transition-colors duration-200">
    <!-- Ambient SCADA Glow Background Gradients -->
    <div class="absolute -top-40 -right-40 w-96 h-96 bg-blue-500/10 dark:bg-blue-600/15 rounded-full blur-3xl pointer-events-none"></div>
    <div class="absolute -bottom-40 -left-40 w-96 h-96 bg-cyan-500/10 dark:bg-cyan-500/10 rounded-full blur-3xl pointer-events-none"></div>

    <!-- Top Navigation Bar for Theme & Language Quick Selection -->
    <header class="w-full max-w-7xl mx-auto px-6 py-5 flex items-center justify-between z-10">
      <div class="flex items-center space-x-3">
        <div class="w-9 h-9 rounded-xl bg-gradient-to-tr from-blue-600 to-cyan-500 flex items-center justify-center text-white shadow-md shadow-blue-500/20">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
        </div>
        <div>
          <div class="text-xs font-black tracking-widest uppercase text-slate-900 dark:text-white flex items-center gap-1.5 font-mono">
            DATALOGGER-27
            <span class="px-1.5 py-0.5 rounded text-4xs bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20 font-bold">AQMS &bull; SCADA</span>
          </div>
          <div class="text-3xs text-slate-500 dark:text-slate-400">Continuous Environmental Monitoring Platform</div>
        </div>
      </div>

      <!-- Theme & Language Toggles -->
      <div class="flex items-center space-x-2">
        <!-- Language Switcher -->
        <button
          type="button"
          @click="toggleLanguage"
          class="flex items-center space-x-1.5 px-3 py-1.5 text-xs font-mono font-semibold rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 hover:border-blue-500 transition-colors shadow-xs"
          :title="$t('header.toggleLang') || 'Switch Language'"
        >
          <span>🌐</span>
          <span class="uppercase">{{ currentLanguage }}</span>
        </button>

        <!-- Theme Switcher -->
        <button
          type="button"
          @click="toggleTheme"
          class="p-2 rounded-xl bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 hover:border-blue-500 transition-colors shadow-xs"
          :title="$t('header.toggleTheme') || 'Switch Theme'"
        >
          <span v-if="isDarkMode">🌙</span>
          <span v-else>☀️</span>
        </button>
      </div>
    </header>

    <!-- Main Form Center Card -->
    <main class="w-full max-w-md mx-auto px-4 py-8 z-10 flex flex-col justify-center">
      <div class="bg-white/90 dark:bg-[#0E1626]/90 backdrop-blur-md border border-slate-200/80 dark:border-slate-800/80 rounded-3xl p-6 sm:p-8 shadow-2xl shadow-slate-900/10 dark:shadow-black/40">
        
        <!-- Header in Card -->
        <div class="text-center mb-6">
          <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-3xs font-mono font-bold uppercase tracking-wider bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20 mb-3">
            <span class="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></span>
            {{ $t('auth.systemAccess') || 'SYSTEM ACCESS CONTROL' }}
          </div>
          <h1 class="text-xl sm:text-2xl font-black text-slate-900 dark:text-white tracking-tight">
            {{ $t('auth.signInTitle') || 'Sign In to Datalogger' }}
          </h1>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-1">
            {{ $t('auth.signInSubtitle') || 'Enter your operational credentials to access telemetry and monitoring systems' }}
          </p>
        </div>

        <!-- Error Feedback Toast -->
        <div
          v-if="errorMessage"
          class="mb-5 p-3 rounded-2xl bg-rose-500/10 border border-rose-500/30 text-rose-600 dark:text-rose-400 text-xs flex items-start gap-2.5 animate-shake"
        >
          <svg class="w-4 h-4 flex-shrink-0 mt-0.5 text-rose-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <circle cx="12" cy="12" r="10" stroke-width="2" />
            <line x1="12" y1="8" x2="12" y2="12" stroke-width="2" />
            <line x1="12" y1="16" x2="12.01" y2="16" stroke-width="2" />
          </svg>
          <div class="flex-1 font-medium leading-relaxed">{{ errorMessage }}</div>
          <button @click="errorMessage = ''" class="text-rose-400 hover:text-rose-600">&times;</button>
        </div>

        <!-- Credentials Form -->
        <form @submit.prevent="handleLogin" class="space-y-4">
          <!-- Username Input -->
          <div>
            <label class="block text-2xs font-semibold uppercase tracking-wider text-slate-600 dark:text-slate-400 mb-1.5">
              {{ $t('auth.username') || 'Username' }}
            </label>
            <div class="relative">
              <span class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
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
                placeholder="e.g. admin or operator_user"
                class="w-full pl-10 pr-4 py-2.5 bg-slate-50 dark:bg-[#070D18] border border-slate-200 dark:border-slate-800 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all font-mono"
              />
            </div>
          </div>

          <!-- Password Input -->
          <div>
            <label class="block text-2xs font-semibold uppercase tracking-wider text-slate-600 dark:text-slate-400 mb-1.5">
              {{ $t('auth.password') || 'Password' }}
            </label>
            <div class="relative">
              <span class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none text-slate-400">
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
                placeholder="Enter password"
                class="w-full pl-10 pr-10 py-2.5 bg-slate-50 dark:bg-[#070D18] border border-slate-200 dark:border-slate-800 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all font-mono"
              />
              <button
                type="button"
                @click="showPassword = !showPassword"
                class="absolute inset-y-0 right-0 pr-3.5 flex items-center text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 cursor-pointer"
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
              class="w-full py-2.5 px-4 bg-gradient-to-r from-blue-600 to-cyan-600 hover:from-blue-500 hover:to-cyan-500 text-white font-bold text-xs rounded-xl shadow-lg shadow-blue-500/25 transition-all transform active:scale-98 disabled:opacity-50 disabled:pointer-events-none flex items-center justify-center space-x-2"
            >
              <svg v-if="isLoading" class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v8H4z"></path>
              </svg>
              <span>{{ isLoading ? ($t('auth.signingIn') || 'Authenticating...') : ($t('auth.signInButton') || 'Sign In') }}</span>
              <svg v-if="!isLoading" class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M14 5l7 7m0 0l-7 7m7-7H3" />
              </svg>
            </button>
          </div>
        </form>

        <!-- Development & Testing Quick Fill Accounts Strip -->
        <div class="mt-6 pt-5 border-t border-slate-200/80 dark:border-slate-800/80">
          <div class="text-3xs font-mono font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500 text-center mb-3">
            {{ $t('auth.testAccounts') || 'Testing & Demo Accounts (1-Click Fill)' }}
          </div>

          <div class="grid grid-cols-2 gap-2 text-2xs">
            <!-- 1. Superadmin -->
            <button
              type="button"
              @click="fillAccount('admin', 'Password123!')"
              class="p-2 rounded-xl text-left border border-slate-200 dark:border-slate-800 hover:border-purple-500/50 bg-slate-50/50 dark:bg-slate-900/50 hover:bg-purple-500/5 transition-all group"
            >
              <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-purple-600 dark:group-hover:text-purple-400 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-purple-500"></span>
                SUPERADMIN
              </div>
              <div class="text-4xs font-mono text-slate-400">@admin &bull; Full Dev Access</div>
            </button>

            <!-- 2. Admin -->
            <button
              type="button"
              @click="fillAccount('admin_user', 'Password123!')"
              class="p-2 rounded-xl text-left border border-slate-200 dark:border-slate-800 hover:border-blue-500/50 bg-slate-50/50 dark:bg-slate-900/50 hover:bg-blue-500/5 transition-all group"
            >
              <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-blue-600 dark:group-hover:text-blue-400 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
                ADMIN
              </div>
              <div class="text-4xs font-mono text-slate-400">@admin_user &bull; Ops Admin</div>
            </button>

            <!-- 3. Engineer -->
            <button
              type="button"
              @click="fillAccount('engineer_user', 'Password123!')"
              class="p-2 rounded-xl text-left border border-slate-200 dark:border-slate-800 hover:border-amber-500/50 bg-slate-50/50 dark:bg-slate-900/50 hover:bg-amber-500/5 transition-all group"
            >
              <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-amber-600 dark:group-hover:text-amber-400 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
                ENGINEER
              </div>
              <div class="text-4xs font-mono text-slate-400">@engineer_user &bull; Sensors</div>
            </button>

            <!-- 4. Operator -->
            <button
              type="button"
              @click="fillAccount('operator_user', 'Password123!')"
              class="p-2 rounded-xl text-left border border-slate-200 dark:border-slate-800 hover:border-emerald-500/50 bg-slate-50/50 dark:bg-slate-900/50 hover:bg-emerald-500/5 transition-all group"
            >
              <div class="font-bold text-slate-800 dark:text-slate-200 group-hover:text-emerald-600 dark:group-hover:text-emerald-400 flex items-center gap-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                OPERATOR
              </div>
              <div class="text-4xs font-mono text-slate-400">@operator_user &bull; Console</div>
            </button>
          </div>
        </div>

      </div>
    </main>

    <!-- Footer System Status & Compliance Info -->
    <footer class="w-full max-w-7xl mx-auto px-6 py-4 text-center z-10">
      <div class="text-3xs font-mono text-slate-400 dark:text-slate-500 flex flex-wrap items-center justify-center gap-3">
        <span>Datalogger-27 Edge v1.0.0</span>
        <span>&bull;</span>
        <span>KLHK Permen LHK No. 14/2020 Compliant</span>
        <span>&bull;</span>
        <span>SPARING Direct Integration Ready</span>
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
      return this.$store.state.uiPreferences?.effectiveTheme === 'dark';
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

        if (redirect && redirect !== '/login') {
          // Check if destination is restricted for non-superadmin
          const isDevRoute = redirect.startsWith('/development');
          const isUsersRoute = redirect.startsWith('/administration/users');
          const isRolesRoute = redirect.startsWith('/administration/roles');

          if (isDevRoute && role !== 'SUPERADMIN' && role !== 'ADMINISTRATOR') {
            this.$router.push('/dashboard');
            return;
          }
          if (isUsersRoute && role !== 'SUPERADMIN' && role !== 'ADMINISTRATOR' && role !== 'ADMIN') {
            this.$router.push('/dashboard');
            return;
          }
          if (isRolesRoute && role !== 'SUPERADMIN' && role !== 'ADMINISTRATOR') {
            this.$router.push('/dashboard');
            return;
          }

          this.$router.push(redirect);
        } else {
          // If operator, default to operator console or dashboard
          if (role === 'OPERATOR') {
            this.$router.push('/monitoring/operator');
          } else {
            this.$router.push('/dashboard');
          }
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
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  20%, 60% { transform: translateX(-4px); }
  40%, 80% { transform: translateX(4px); }
}
.animate-shake {
  animation: shake 0.4s ease-in-out;
}
</style>
