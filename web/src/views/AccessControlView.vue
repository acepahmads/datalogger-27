<template>
  <div class="space-y-4 font-sans text-slate-800 dark:text-slate-200">
    <!-- Header with Title, Badges & Tab Switcher -->
    <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center space-x-3">
        <div class="w-10 h-10 rounded-xl bg-purple-500/10 border border-purple-500/20 text-purple-600 dark:text-purple-400 flex items-center justify-center flex-shrink-0">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
          </svg>
        </div>
        <div>
          <h1 class="text-base font-bold text-slate-900 dark:text-white tracking-tight flex items-center gap-2">
            {{ $t('accessControl.title') || 'Access Control & Administration' }}
            <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20">
              #80 #81 #82
            </span>
          </h1>
          <p class="text-xs text-slate-500 dark:text-slate-400">
            {{ $t('accessControl.subtitle') || 'User credential management, RBAC role definitions, and granular capability matrices' }}
          </p>
        </div>
      </div>

      <!-- Tab Switcher -->
      <div class="flex items-center p-1 bg-slate-100 dark:bg-[#0B0F19] rounded-xl border border-slate-200 dark:border-slate-800 text-xs font-medium">
        <button id="tab-users"
                @click="activeTab = 'users'"
                class="px-3 py-1.5 rounded-lg transition-all flex items-center gap-1.5"
                :class="activeTab === 'users' ? 'bg-white dark:bg-slate-800 text-blue-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
          {{ $t('accessControl.usersTab') || 'Users' }} ({{ usersTotal }})
        </button>

        <button id="tab-roles"
                v-if="isSuperAdmin"
                @click="activeTab = 'roles'"
                class="px-3 py-1.5 rounded-lg transition-all flex items-center gap-1.5"
                :class="activeTab === 'roles' ? 'bg-white dark:bg-slate-800 text-blue-600 dark:text-white font-semibold shadow-xs' : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'">
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
            <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
          </svg>
          {{ $t('accessControl.rolesTab') || 'Roles & Permissions' }} ({{ roles.length }})
        </button>
      </div>
    </div>

    <!-- Feedback Message Toast -->
    <div v-if="toastMessage" class="p-3 rounded-xl text-xs font-semibold flex items-center justify-between transition-all"
         :class="toastType === 'error' ? 'bg-rose-500/10 border border-rose-500/20 text-rose-500' : 'bg-emerald-500/10 border border-emerald-500/20 text-emerald-500'">
      <span>{{ toastMessage }}</span>
      <button @click="toastMessage = ''" class="hover:opacity-75">&times;</button>
    </div>

    <!-- ============================================================== -->
    <!-- TAB 1: USER MANAGEMENT (#80)                                   -->
    <!-- ============================================================== -->
    <div v-if="activeTab === 'users'" class="space-y-4">
      <!-- Filter Bar -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
        <div class="flex flex-wrap items-center gap-3 flex-1">
          <!-- Search input -->
          <div class="relative flex-1 min-w-[200px] max-w-md">
            <svg class="w-4 h-4 text-slate-400 absolute left-3 top-2.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <circle cx="11" cy="11" r="8"></circle>
              <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
            </svg>
            <input type="text"
                   v-model="userSearch"
                   @input="debounceFetchUsers"
                   :placeholder="$t('accessControl.searchPlaceholder') || 'Search username, full name, email...'"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl pl-9 pr-3 py-2 text-slate-800 dark:text-slate-200 focus:ring-1 focus:ring-blue-500" />
          </div>

          <!-- Status Filter -->
          <div class="flex items-center gap-1 p-1 bg-slate-100 dark:bg-[#0B0F19] rounded-xl border border-slate-200 dark:border-slate-800">
            <button @click="setUserStatusFilter('ALL')"
                    class="px-2.5 py-1 rounded-lg text-2xs transition-all"
                    :class="userStatusFilter === 'ALL' ? 'bg-white dark:bg-slate-800 text-blue-600 dark:text-white font-semibold shadow-xs' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'">
              All
            </button>
            <button @click="setUserStatusFilter('ACTIVE')"
                    class="px-2.5 py-1 rounded-lg text-2xs transition-all"
                    :class="userStatusFilter === 'ACTIVE' ? 'bg-white dark:bg-slate-800 text-emerald-600 dark:text-emerald-400 font-semibold shadow-xs' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'">
              Active
            </button>
            <button @click="setUserStatusFilter('INACTIVE')"
                    class="px-2.5 py-1 rounded-lg text-2xs transition-all"
                    :class="userStatusFilter === 'INACTIVE' ? 'bg-white dark:bg-slate-800 text-rose-600 dark:text-rose-400 font-semibold shadow-xs' : 'text-slate-500 hover:text-slate-900 dark:hover:text-white'">
              Disabled
            </button>
          </div>
        </div>

        <!-- Add User Button -->
        <button id="btn-add-user"
                @click="openCreateUserModal"
                class="px-4 py-2 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-semibold shadow-xs transition-colors flex items-center gap-1.5 self-start sm:self-auto">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          {{ $t('accessControl.createUser') || 'Create User' }}
        </button>
      </div>

      <!-- Users Table -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl overflow-hidden shadow-sm">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs">
            <thead class="bg-slate-50 dark:bg-[#0B0F19]/60 text-3xs font-semibold uppercase tracking-wider text-slate-500 border-b border-slate-200 dark:border-slate-800">
              <tr>
                <th class="px-4 py-3">User</th>
                <th class="px-4 py-3">Role</th>
                <th class="px-4 py-3">Status</th>
                <th class="px-4 py-3">Email</th>
                <th class="px-4 py-3">Last Login</th>
                <th class="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800/60 font-mono text-2xs">
              <tr v-if="usersLoading">
                <td colspan="6" class="px-4 py-8 text-center text-slate-500 font-sans">
                  <div class="flex items-center justify-center gap-2">
                    <svg class="w-4 h-4 animate-spin text-blue-500" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    <span>Loading users...</span>
                  </div>
                </td>
              </tr>
              <tr v-else-if="users.length === 0">
                <td colspan="6" class="px-4 py-8 text-center text-slate-500 font-sans">
                  No users found matching current filters.
                </td>
              </tr>
              <tr v-for="u in users" :key="u.id" class="hover:bg-slate-50/50 dark:hover:bg-slate-800/30 font-sans">
                <!-- User Profile -->
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2.5">
                    <div class="w-7 h-7 rounded-full bg-slate-200 dark:bg-slate-700 flex items-center justify-center font-bold text-2xs text-slate-700 dark:text-slate-300">
                      {{ u.username.substring(0, 2).toUpperCase() }}
                    </div>
                    <div>
                      <div class="font-bold text-slate-900 dark:text-white">{{ u.full_name || u.username }}</div>
                      <div class="font-mono text-3xs text-slate-400">@{{ u.username }}</div>
                    </div>
                  </div>
                </td>

                <!-- Role -->
                <td class="px-4 py-3">
                  <span class="inline-flex items-center px-2 py-0.5 rounded-full text-3xs font-semibold"
                        :class="getRoleBadgeClass(u.role ? u.role.name : '')">
                    {{ u.role ? u.role.name : 'No Role' }}
                  </span>
                </td>

                <!-- Status -->
                <td class="px-4 py-3">
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-3xs font-bold"
                        :class="u.is_active ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-500 border border-rose-500/20'">
                    <span class="w-1.5 h-1.5 rounded-full" :class="u.is_active ? 'bg-emerald-500' : 'bg-rose-500'"></span>
                    {{ u.is_active ? 'Active' : 'Disabled' }}
                  </span>
                </td>

                <!-- Email -->
                <td class="px-4 py-3 font-mono text-2xs text-slate-500">
                  {{ u.email || '-' }}
                </td>

                <!-- Last Login -->
                <td class="px-4 py-3 font-mono text-3xs text-slate-500">
                  {{ formatTimestamp(u.last_login) }}
                </td>

                <!-- Actions -->
                <td class="px-4 py-3 text-right">
                  <div v-if="canManageUser(u)" class="flex items-center justify-end gap-1.5">
                    <!-- Edit Button -->
                    <button @click="openEditUserModal(u)"
                            class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500 hover:text-blue-500 transition-colors"
                            title="Edit User">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                        <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                      </svg>
                    </button>

                    <!-- Reset Password Button -->
                    <button @click="openResetPasswordModal(u)"
                            class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500 hover:text-amber-500 transition-colors"
                            title="Reset Password">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                        <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                      </svg>
                    </button>

                    <!-- Toggle Status Button -->
                    <button @click="toggleUserStatus(u)"
                            class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
                            :class="u.is_active ? 'text-slate-500 hover:text-rose-500' : 'text-slate-500 hover:text-emerald-500'"
                            :title="u.is_active ? 'Disable Account' : 'Enable Account'">
                      <svg v-if="u.is_active" class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <circle cx="12" cy="12" r="10"></circle>
                        <line x1="4.93" y1="4.93" x2="19.07" y2="19.07"></line>
                      </svg>
                      <svg v-else class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <polyline points="20 6 9 17 4 12"></polyline>
                      </svg>
                    </button>

                    <!-- Delete Button -->
                    <button @click="confirmDeleteUser(u)"
                            class="p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 text-slate-500 hover:text-rose-600 transition-colors"
                            title="Delete User">
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                        <polyline points="3 6 5 6 21 6"></polyline>
                        <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
                      </svg>
                    </button>
                  </div>
                  <div v-else class="text-3xs font-mono text-slate-400 italic">
                    Protected
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- TAB 2: ROLE & PERMISSION MANAGEMENT (#81, #82)                 -->
    <!-- ============================================================== -->
    <div v-if="activeTab === 'roles'" class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <!-- Roles List Sidebar -->
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-4 shadow-sm space-y-3">
        <div class="flex items-center justify-between pb-2 border-b border-slate-100 dark:border-slate-800">
          <h2 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">System Roles</h2>
          <button id="btn-create-role"
                  @click="openCreateRoleModal"
                  class="px-2.5 py-1 bg-blue-600 hover:bg-blue-500 text-white rounded-lg text-2xs font-semibold transition-colors flex items-center gap-1">
            + New Role
          </button>
        </div>

        <div class="space-y-1.5">
          <div v-for="r in roles" :key="r.id"
               @click="selectRole(r)"
               class="p-3 rounded-xl border transition-all cursor-pointer flex items-center justify-between"
               :class="selectedRole && selectedRole.id === r.id ? 'bg-blue-50/50 dark:bg-blue-600/10 border-blue-500/40 text-blue-600 dark:text-blue-400' : 'bg-slate-50 dark:bg-[#0B0F19]/60 border-slate-200 dark:border-slate-800 text-slate-700 dark:text-slate-300 hover:border-slate-300 dark:hover:border-slate-700'">
            <div>
              <div class="font-bold text-xs flex items-center gap-2">
                {{ r.name }}
                <span v-if="isSystemRole(r.name)" class="text-3xs uppercase font-mono px-1.5 py-0.5 rounded bg-slate-200 dark:bg-slate-800 text-slate-500">
                  SYSTEM
                </span>
              </div>
              <div class="text-3xs text-slate-500 mt-0.5 max-w-[180px] truncate">{{ r.description || 'No description' }}</div>
            </div>
            <span class="text-3xs font-mono font-bold px-2 py-0.5 rounded-full bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
              {{ r.permissions ? r.permissions.length : 0 }} perms
            </span>
          </div>
        </div>
      </div>

      <!-- Permission Matrix Editor -->
      <div class="md:col-span-2 bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl p-5 shadow-sm space-y-4">
        <div v-if="selectedRole" class="space-y-4">
          <!-- Role Details Header -->
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pb-3 border-b border-slate-100 dark:border-slate-800">
            <div>
              <h2 class="text-sm font-bold text-slate-900 dark:text-white flex items-center gap-2">
                {{ selectedRole.name }}
                <span v-if="isSystemRole(selectedRole.name)" class="px-2 py-0.5 rounded text-3xs font-mono bg-blue-500/10 text-blue-500 border border-blue-500/20 font-bold">
                  PROTECTED ROLE
                </span>
              </h2>
              <p class="text-2xs text-slate-500 mt-0.5">{{ selectedRole.description }}</p>
            </div>

            <!-- Actions -->
            <div class="flex items-center gap-2">
              <button v-if="!isSystemRole(selectedRole.name)"
                      @click="deleteRole(selectedRole)"
                      class="px-3 py-1.5 bg-rose-500/10 hover:bg-rose-500/20 text-rose-500 rounded-xl text-xs font-semibold transition-colors">
                Delete Role
              </button>

              <button @click="saveRolePermissions"
                      :disabled="savingPermissions"
                      class="px-4 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl text-xs font-semibold shadow-xs transition-colors flex items-center gap-1.5 disabled:opacity-50">
                <span v-if="savingPermissions" class="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
                Save Permissions
              </button>
            </div>
          </div>

          <!-- Admin Protection Info Banner -->
          <div v-if="selectedRole.name === 'Administrator'"
               class="p-3.5 rounded-xl bg-blue-500/10 border border-blue-500/20 text-xs text-blue-600 dark:text-blue-400 space-y-1">
            <div class="font-bold flex items-center gap-1.5">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
                <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
              </svg>
              Immutable Administrative Authority
            </div>
            <p class="text-3xs text-blue-600/90 dark:text-blue-300/80 leading-relaxed font-sans">
              The Administrator role maintains permanent full authority across all system resources. Sensitive governance permissions (user.manage, role.manage, config.manage) cannot be revoked, safeguarding against accidental administrative lock-out.
            </p>
          </div>

          <!-- Permission Categories Matrix -->
          <div class="space-y-4 max-h-[600px] overflow-y-auto pr-1">
            <div v-for="cat in permissionCategories" :key="cat.category"
                 class="p-3.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19]/60 border border-slate-200 dark:border-slate-800/80 space-y-2.5">
              <div class="flex items-center justify-between">
                <div class="font-bold text-xs uppercase tracking-wider text-slate-700 dark:text-slate-300 flex items-center gap-2">
                  <span class="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
                  {{ cat.category }}
                </div>
                <div class="flex items-center gap-2 text-3xs text-slate-500">
                  <button @click="toggleCategory(cat, true)" class="hover:text-blue-500">Select All</button>
                  <span>&bull;</span>
                  <button @click="toggleCategory(cat, false)" class="hover:text-blue-500">Clear</button>
                </div>
              </div>

              <!-- Permission Checkbox Grid -->
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
                <label v-for="p in cat.permissions" :key="p.id"
                       class="flex items-start gap-2 p-2 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/50 cursor-pointer select-none transition-colors">
                  <input type="checkbox"
                         :value="p.id"
                         v-model="selectedPermissionIDs"
                         class="mt-0.5 rounded border-slate-300 text-blue-600 focus:ring-blue-500 dark:bg-slate-800 dark:border-slate-700" />
                  <div>
                    <span class="font-medium text-slate-800 dark:text-slate-200 block">{{ p.name }}</span>
                    <span class="font-mono text-3xs text-slate-400 block">{{ p.code }}</span>
                  </div>
                </label>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="text-center py-12 text-slate-500 text-xs">
          Select a role from the left sidebar to inspect and configure permissions.
        </div>
      </div>
    </div>

    <!-- ============================================================== -->
    <!-- MODALS: User Create / Edit / Password Reset / Role Create     -->
    <!-- ============================================================== -->

    <!-- Create User Modal -->
    <div v-if="showCreateUserModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs font-sans">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-md p-5 shadow-xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-sm text-slate-900 dark:text-white">Create New User</h3>
          <button @click="showCreateUserModal = false" class="text-slate-400 hover:text-white text-lg">&times;</button>
        </div>

        <form @submit.prevent="submitCreateUser" class="space-y-3 text-xs">
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Username *</label>
            <input type="text" v-model="userForm.username" required placeholder="e.g. operator_shift1"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200" />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Full Name</label>
            <input type="text" v-model="userForm.full_name" placeholder="e.g. John Doe"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200" />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Email</label>
            <input type="email" v-model="userForm.email" placeholder="john@plant.local"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200" />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Password * (min 6 chars)</label>
            <input type="password" v-model="userForm.password" required minlength="6" placeholder="••••••••"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200 font-mono" />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Assign Role *</label>
            <select v-model="userForm.role_id" required
                    class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200">
              <option v-for="r in assignableRoles" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </div>

          <div class="flex items-center gap-2 pt-1">
            <input type="checkbox" id="user-active-toggle" v-model="userForm.is_active" class="rounded text-blue-600" />
            <label for="user-active-toggle" class="text-xs text-slate-700 dark:text-slate-300">Active account immediately</label>
          </div>

          <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
            <button type="button" @click="showCreateUserModal = false" class="px-3 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Cancel</button>
            <button type="submit" class="px-4 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-semibold">Create Account</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Edit User Modal -->
    <div v-if="showEditUserModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs font-sans">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-md p-5 shadow-xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-sm text-slate-900 dark:text-white">Edit User: {{ editingUser.username }}</h3>
          <button @click="showEditUserModal = false" class="text-slate-400 hover:text-white text-lg">&times;</button>
        </div>

        <form @submit.prevent="submitEditUser" class="space-y-3 text-xs">
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Full Name</label>
            <input type="text" v-model="editUserForm.full_name"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200" />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Email</label>
            <input type="email" v-model="editUserForm.email"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200" />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Role</label>
            <select v-model="editUserForm.role_id" required
                    class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200">
              <option v-for="r in assignableRoles" :key="r.id" :value="r.id">{{ r.name }}</option>
            </select>
          </div>

          <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
            <button type="button" @click="showEditUserModal = false" class="px-3 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Cancel</button>
            <button type="submit" class="px-4 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-semibold">Save Changes</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Reset Password Modal -->
    <div v-if="showPasswordModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs font-sans">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-sm p-5 shadow-xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-sm text-slate-900 dark:text-white">Reset Password</h3>
          <button @click="showPasswordModal = false" class="text-slate-400 hover:text-white text-lg">&times;</button>
        </div>

        <form @submit.prevent="submitResetPassword" class="space-y-3 text-xs">
          <p class="text-2xs text-slate-500">
            Set a new secure password for <strong>{{ passwordUser.username }}</strong>.
          </p>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">New Password (min 6 chars)</label>
            <input type="password" v-model="newPassword" required minlength="6" placeholder="••••••••"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200 font-mono" />
          </div>

          <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
            <button type="button" @click="showPasswordModal = false" class="px-3 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Cancel</button>
            <button type="submit" class="px-4 py-1.5 bg-amber-600 hover:bg-amber-500 text-white rounded-xl font-semibold">Update Password</button>
          </div>
        </form>
      </div>
    </div>

    <!-- Create Role Modal -->
    <div v-if="showCreateRoleModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-xs font-sans">
      <div class="bg-white dark:bg-[#111827] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-md p-5 shadow-xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h3 class="font-bold text-sm text-slate-900 dark:text-white">Create Custom Role</h3>
          <button @click="showCreateRoleModal = false" class="text-slate-400 hover:text-white text-lg">&times;</button>
        </div>

        <form @submit.prevent="submitCreateRole" class="space-y-3 text-xs">
          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Role Name *</label>
            <input type="text" v-model="roleForm.name" required placeholder="e.g. Quality Auditor"
                   class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200" />
          </div>

          <div>
            <label class="block text-3xs font-semibold uppercase text-slate-500 mb-1">Description</label>
            <textarea v-model="roleForm.description" rows="3" placeholder="Explain the role duties and intended permissions..."
                      class="w-full bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 rounded-xl px-3 py-2 text-slate-800 dark:text-slate-200"></textarea>
          </div>

          <div class="flex items-center justify-end gap-2 pt-3 border-t border-slate-100 dark:border-slate-800">
            <button type="button" @click="showCreateRoleModal = false" class="px-3 py-1.5 rounded-xl bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Cancel</button>
            <button type="submit" class="px-4 py-1.5 bg-blue-600 hover:bg-blue-500 text-white rounded-xl font-semibold">Create Role</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'AccessControlView',
  data() {
    return {
      activeTab: 'users',
      toastMessage: '',
      toastType: 'success',

      // Users State
      users: [],
      usersTotal: 0,
      usersLoading: false,
      userSearch: '',
      userStatusFilter: 'ALL',
      searchTimeout: null,

      // Roles & Permissions State
      roles: [],
      selectedRole: null,
      permissionCategories: [],
      selectedPermissionIDs: [],
      savingPermissions: false,

      // Modals
      showCreateUserModal: false,
      showEditUserModal: false,
      showPasswordModal: false,
      showCreateRoleModal: false,

      userForm: {
        username: '',
        full_name: '',
        email: '',
        password: '',
        role_id: 1,
        is_active: true,
      },
      editingUser: {},
      editUserForm: {
        full_name: '',
        email: '',
        role_id: 1,
      },
      passwordUser: {},
      newPassword: '',

      roleForm: {
        name: '',
        description: '',
      },
    };
  },
  computed: {
    userRole() {
      const u = this.$store.state.user || {};
      return (u.role || '').toUpperCase();
    },
    isSuperAdmin() {
      return this.userRole === 'SUPERADMIN' || this.userRole === 'ADMINISTRATOR';
    },
    isAdmin() {
      return this.userRole === 'ADMIN';
    },
    assignableRoles() {
      if (this.isSuperAdmin) {
        return this.roles;
      }
      return this.roles.filter(r => {
        const name = (r.name || '').toUpperCase();
        return name !== 'SUPERADMIN' && name !== 'ADMINISTRATOR';
      });
    },
  },
  mounted() {
    this.fetchUsers();
    this.fetchRoles();
    this.fetchPermissions();
  },
  methods: {
    showToast(msg, type = 'success') {
      this.toastMessage = msg;
      this.toastType = type;
      setTimeout(() => {
        if (this.toastMessage === msg) {
          this.toastMessage = '';
        }
      }, 5000);
    },
    isSystemRole(name) {
      if (!name) return false;
      const lower = name.toLowerCase();
      return lower === 'administrator' || lower === 'engineer' || lower === 'operator';
    },
    getRoleBadgeClass(roleName) {
      if (!roleName) return 'bg-slate-100 text-slate-600';
      const lower = roleName.toLowerCase();
      if (lower.includes('admin')) {
        return 'bg-purple-500/10 text-purple-600 dark:text-purple-400 border border-purple-500/20';
      }
      if (lower.includes('engineer')) {
        return 'bg-blue-500/10 text-blue-600 dark:text-blue-400 border border-blue-500/20';
      }
      if (lower.includes('operator')) {
        return 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20';
      }
      return 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20';
    },
    formatTimestamp(ts) {
      if (!ts) return 'Never';
      const d = new Date(ts);
      return d.toLocaleString('en-GB', { hour12: false });
    },

    // Users Methods
    async fetchUsers() {
      this.usersLoading = true;
      try {
        const params = new URLSearchParams();
        if (this.userSearch) params.append('search', this.userSearch);
        if (this.userStatusFilter !== 'ALL') params.append('status', this.userStatusFilter);
        params.append('page', '1');
        params.append('page_size', '50');

        const res = await axios.get(`/api/users?${params.toString()}`);
        if (res.data && res.data.data) {
          this.users = res.data.data.users || [];
          this.usersTotal = res.data.data.total || 0;
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      } finally {
        this.usersLoading = false;
      }
    },
    debounceFetchUsers() {
      clearTimeout(this.searchTimeout);
      this.searchTimeout = setTimeout(() => {
        this.fetchUsers();
      }, 300);
    },
    setUserStatusFilter(status) {
      this.userStatusFilter = status;
      this.fetchUsers();
    },
    canManageUser(u) {
      if (this.isSuperAdmin) return true;
      if (!u) return false;
      const roleName = (u.role ? u.role.name : '').toUpperCase();
      return roleName !== 'SUPERADMIN' && roleName !== 'ADMINISTRATOR';
    },
    openCreateUserModal() {
      const defaultRole = this.assignableRoles.length > 0 ? this.assignableRoles[0].id : 1;
      this.userForm = {
        username: '',
        full_name: '',
        email: '',
        password: '',
        role_id: defaultRole,
        is_active: true,
      };
      this.showCreateUserModal = true;
    },
    async submitCreateUser() {
      if (!this.isSuperAdmin) {
        const targetRole = this.roles.find(r => r.id === this.userForm.role_id);
        if (targetRole && (targetRole.name.toUpperCase() === 'SUPERADMIN' || targetRole.name.toUpperCase() === 'ADMINISTRATOR')) {
          this.showToast('You do not have permission to assign the SUPERADMIN role.', 'error');
          return;
        }
      }
      try {
        const res = await axios.post('/api/users', this.userForm);
        if (res.data && res.data.success) {
          this.showToast(`User '${this.userForm.username}' created successfully.`);
          this.showCreateUserModal = false;
          this.fetchUsers();
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      }
    },
    openEditUserModal(user) {
      this.editingUser = user;
      this.editUserForm = {
        full_name: user.full_name || '',
        email: user.email || '',
        role_id: user.role_id,
      };
      this.showEditUserModal = true;
    },
    async submitEditUser() {
      if (!this.isSuperAdmin) {
        const targetRole = this.roles.find(r => r.id === this.editUserForm.role_id);
        if (targetRole && (targetRole.name.toUpperCase() === 'SUPERADMIN' || targetRole.name.toUpperCase() === 'ADMINISTRATOR')) {
          this.showToast('You do not have permission to assign the SUPERADMIN role.', 'error');
          return;
        }
      }
      try {
        const res = await axios.put(`/api/users/${this.editingUser.id}`, this.editUserForm);
        if (res.data && res.data.success) {
          this.showToast(`User '${this.editingUser.username}' updated successfully.`);
          this.showEditUserModal = false;
          this.fetchUsers();
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      }
    },
    openResetPasswordModal(user) {
      this.passwordUser = user;
      this.newPassword = '';
      this.showPasswordModal = true;
    },
    async submitResetPassword() {
      try {
        const res = await axios.put(`/api/users/${this.passwordUser.id}/password`, {
          password: this.newPassword,
        });
        if (res.data && res.data.success) {
          this.showToast(`Password for '${this.passwordUser.username}' updated successfully.`);
          this.showPasswordModal = false;
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      }
    },
    async toggleUserStatus(user) {
      const newStatus = !user.is_active;
      const action = newStatus ? 'enable' : 'disable';
      if (!confirm(`Are you sure you want to ${action} user '${user.username}'?`)) {
        return;
      }
      try {
        const res = await axios.put(`/api/users/${user.id}/enable`, { is_active: newStatus });
        if (res.data && res.data.success) {
          this.showToast(`User '${user.username}' status set to ${newStatus ? 'Active' : 'Disabled'}.`);
          this.fetchUsers();
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      }
    },
    async confirmDeleteUser(user) {
      if (!confirm(`Are you sure you want to permanently delete user '${user.username}'? This cannot be undone.`)) {
        return;
      }
      try {
        const res = await axios.delete(`/api/users/${user.id}`);
        if (res.data && res.data.success) {
          this.showToast(`User '${user.username}' deleted.`);
          this.fetchUsers();
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      }
    },

    // Roles & Permissions Methods
    async fetchRoles() {
      try {
        const res = await axios.get('/api/roles');
        if (res.data && res.data.data) {
          this.roles = res.data.data;
          if (!this.selectedRole && this.roles.length > 0) {
            this.selectRole(this.roles[0]);
          }
        }
      } catch (err) {
        this.showToast('Failed to fetch roles: ' + err.message, 'error');
      }
    },
    async fetchPermissions() {
      try {
        const res = await axios.get('/api/permissions');
        if (res.data && res.data.data) {
          this.permissionCategories = res.data.data.grouped || [];
        }
      } catch (err) {
        this.showToast('Failed to fetch permissions: ' + err.message, 'error');
      }
    },
    selectRole(role) {
      this.selectedRole = role;
      if (role.permissions) {
        this.selectedPermissionIDs = role.permissions.map(p => p.id);
      } else {
        this.selectedPermissionIDs = [];
      }
    },
    toggleCategory(cat, selectAll) {
      const catIDs = cat.permissions.map(p => p.id);
      if (selectAll) {
        catIDs.forEach(id => {
          if (!this.selectedPermissionIDs.includes(id)) {
            this.selectedPermissionIDs.push(id);
          }
        });
      } else {
        this.selectedPermissionIDs = this.selectedPermissionIDs.filter(id => !catIDs.includes(id));
      }
    },
    async saveRolePermissions() {
      if (!this.selectedRole) return;
      this.savingPermissions = true;
      try {
        const res = await axios.put(`/api/roles/${this.selectedRole.id}/permissions`, {
          permission_ids: this.selectedPermissionIDs,
        });
        if (res.data && res.data.success) {
          this.showToast(`Permission matrix updated for '${this.selectedRole.name}'.`);
          this.fetchRoles();
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      } finally {
        this.savingPermissions = false;
      }
    },
    openCreateRoleModal() {
      this.roleForm = { name: '', description: '' };
      this.showCreateRoleModal = true;
    },
    async submitCreateRole() {
      try {
        const res = await axios.post('/api/roles', this.roleForm);
        if (res.data && res.data.success) {
          this.showToast(`Custom role '${this.roleForm.name}' created.`);
          this.showCreateRoleModal = false;
          await this.fetchRoles();
          const created = this.roles.find(r => r.name === this.roleForm.name);
          if (created) this.selectRole(created);
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      }
    },
    async deleteRole(role) {
      if (!confirm(`Are you sure you want to delete role '${role.name}'?`)) {
        return;
      }
      try {
        const res = await axios.delete(`/api/roles/${role.id}`);
        if (res.data && res.data.success) {
          this.showToast(`Role '${role.name}' deleted.`);
          this.selectedRole = null;
          this.fetchRoles();
        }
      } catch (err) {
        this.showToast(err.response?.data?.error || err.message, 'error');
      }
    },
  },
};
</script>
