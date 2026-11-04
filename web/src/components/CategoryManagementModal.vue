<template>
  <div v-if="isOpen" class="fixed inset-0 z-[70] flex items-center justify-center p-3 sm:p-4 bg-black/75 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] rounded-2xl w-full max-w-4xl max-h-[92vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-200 dark:border-[#20354B] flex items-center justify-between bg-slate-50/95 dark:bg-[#080E18]">
        <div class="flex items-center space-x-2.5">
          <div class="p-2 rounded-xl bg-blue-500/10 text-blue-600 border border-blue-500/20 dark:bg-cyan-500/10 dark:text-cyan-400 dark:border-cyan-500/20">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
            </svg>
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="text-sm font-bold text-slate-900 dark:text-white tracking-tight uppercase">
                {{ $t('categoryManagement.title') || 'PARAMETER CATEGORY MANAGEMENT' }}
              </h2>
              <span
                v-if="!canManage"
                class="px-2 py-0.5 rounded text-4xs font-bold uppercase bg-amber-500/15 text-amber-700 dark:text-amber-300 border border-amber-500/30"
              >
                Read-Only (Operator)
              </span>
            </div>
            <p class="text-3xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ $t('categoryManagement.subtitle') || 'Configure domain categories, display ordering, responsive spans, and assignments.' }}
            </p>
          </div>
        </div>

        <button
          type="button"
          @click="closeModal"
          class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1.5 rounded-lg hover:bg-slate-200/70 dark:hover:bg-[#1E2E42] transition"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Tab Strip & Actions Bar -->
      <div class="px-6 py-2.5 bg-slate-100/80 dark:bg-[#0e192a] border-b border-slate-200 dark:border-[#20354B] flex flex-wrap items-center justify-between gap-3 text-2xs">
        <div class="flex items-center space-x-1.5 bg-white dark:bg-[#080E18] p-1 rounded-xl border border-slate-200 dark:border-[#20354B]">
          <button
            type="button"
            @click="activeTab = 'list'"
            class="px-3 py-1.5 rounded-lg transition font-semibold flex items-center space-x-1.5"
            :class="activeTab === 'list'
              ? 'bg-blue-600/10 text-blue-700 border border-blue-500/30 dark:bg-cyan-500/20 dark:text-cyan-300 dark:border-cyan-500/40 shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            <span>▦</span>
            <span>{{ $t('categoryManagement.tabs.categories') || 'Categories & Layout' }}</span>
            <span class="px-1.5 py-0.2 rounded-full bg-slate-100 dark:bg-[#18283D] text-slate-700 dark:text-slate-300 text-3xs font-mono tabular-nums">{{ categories.length }}</span>
          </button>

          <button
            type="button"
            @click="activeTab = 'assignments'"
            class="px-3 py-1.5 rounded-lg transition font-semibold flex items-center space-x-1.5"
            :class="activeTab === 'assignments'
              ? 'bg-blue-600/10 text-blue-700 border border-blue-500/30 dark:bg-cyan-500/20 dark:text-cyan-300 dark:border-cyan-500/40 shadow-xs'
              : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
          >
            <span>⇄</span>
            <span>{{ $t('categoryManagement.tabs.assignments') || 'Parameter Assignments' }}</span>
          </button>
        </div>

        <button
          v-if="canManage"
          type="button"
          @click="openCreateForm"
          class="px-3.5 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white dark:bg-cyan-600 dark:hover:bg-cyan-500 dark:text-slate-950 font-bold transition flex items-center space-x-1.5 shadow-sm"
        >
          <span class="text-sm leading-none">+</span>
          <span>{{ $t('categoryManagement.actions.add') || 'Add Category' }}</span>
        </button>
      </div>

      <!-- Main Body Area -->
      <div class="flex-1 overflow-y-auto p-5 sm:p-6 space-y-4 custom-scroller bg-slate-50/40 dark:bg-transparent">
        <!-- Notification Banner -->
        <div v-if="noticeMessage" class="p-3 bg-emerald-50 dark:bg-emerald-500/15 border border-emerald-300 dark:border-emerald-500/30 rounded-xl text-xs text-emerald-800 dark:text-emerald-300 flex items-center justify-between animate-fade-in">
          <span>✓ {{ noticeMessage }}</span>
          <button type="button" @click="noticeMessage = ''" class="text-emerald-600 dark:text-emerald-400 hover:opacity-80 text-3xs">✕</button>
        </div>

        <!-- Error Alert -->
        <div v-if="errorMessage" class="p-3 bg-rose-50 dark:bg-rose-500/15 border border-rose-300 dark:border-rose-500/30 rounded-xl text-xs text-rose-800 dark:text-rose-300 flex items-center justify-between animate-fade-in">
          <span>⚠️ {{ errorMessage }}</span>
          <button type="button" @click="errorMessage = ''" class="text-rose-600 dark:text-rose-400 hover:opacity-80 text-3xs">✕</button>
        </div>

        <!-- ==================================================== -->
        <!-- TAB 1: CATEGORY LIST & LAYOUT REORDERING              -->
        <!-- ==================================================== -->
        <div v-if="activeTab === 'list'" class="space-y-3">
          <div class="flex flex-wrap items-center justify-between gap-2 text-3xs text-slate-500 dark:text-slate-400">
            <div class="flex items-center gap-2 flex-1 min-w-[220px]">
              <input
                v-model="categorySearch"
                type="text"
                placeholder="Search category name, code, or domain..."
                class="px-3 py-1.5 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-lg text-slate-900 dark:text-white text-3xs focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500 w-full max-w-xs"
              />
              <button
                v-if="categorySearch"
                type="button"
                @click="categorySearch = ''"
                class="text-3xs text-slate-400 hover:text-slate-700 dark:hover:text-white px-1.5 py-1"
              >
                Clear
              </button>
            </div>
            <div class="flex items-center gap-3">
              <span>Arrange category order and desktop column span (1–4 columns).</span>
              <span v-if="hasReorderChanges" class="text-amber-600 dark:text-amber-400 font-semibold animate-pulse">● Unsaved order changes</span>
            </div>
          </div>

          <div v-if="filteredCategories.length === 0" class="p-6 text-center rounded-xl border border-dashed border-slate-300 dark:border-[#20354B] text-xs text-slate-500 dark:text-slate-400 bg-white dark:bg-[#080E18]">
            No categories match your filter.
            <button
              v-if="canManage"
              type="button"
              @click="openCreateForm"
              class="ml-1 text-blue-600 dark:text-cyan-400 font-semibold hover:underline"
            >
              + Create a new category
            </button>
          </div>

          <div class="space-y-2">
            <div
              v-for="(cat, idx) in filteredCategories"
              :key="cat.id"
              class="p-3 rounded-xl border transition flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-white dark:bg-[#080E18] border-slate-200 dark:border-[#20354B] hover:border-slate-400 dark:hover:border-slate-600 shadow-2xs"
              :class="!cat.enabled ? 'opacity-60 bg-slate-100/70 dark:bg-slate-900/40' : ''"
            >
              <!-- Left: Order Controls, Icon, Name, Code, Span Badge -->
              <div class="flex items-center space-x-3 min-w-0 flex-1">
                <!-- Move Up / Down Buttons -->
                <div v-if="canManage && !categorySearch.trim()" class="flex flex-col space-y-0.5 flex-shrink-0">
                  <button
                    type="button"
                    @click="moveCategoryUp(idx)"
                    :disabled="idx === 0"
                    class="p-1 rounded hover:bg-slate-100 dark:hover:bg-[#1E2E42] text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white disabled:opacity-20 text-3xs"
                    :title="$t('categoryManagement.actions.moveUp') || 'Move Up'"
                  >
                    ▲
                  </button>
                  <button
                    type="button"
                    @click="moveCategoryDown(idx)"
                    :disabled="idx === categories.length - 1"
                    class="p-1 rounded hover:bg-slate-100 dark:hover:bg-[#1E2E42] text-slate-500 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white disabled:opacity-20 text-3xs"
                    :title="$t('categoryManagement.actions.moveDown') || 'Move Down'"
                  >
                    ▼
                  </button>
                </div>

                <!-- Category Icon & Color Badge -->
                <div
                  class="w-9 h-9 rounded-xl flex items-center justify-center flex-shrink-0 border"
                  :class="[getCategoryTheme(cat.color_key).bg, getCategoryTheme(cat.color_key).border]"
                >
                  <ParameterIcon :icon="cat.icon_key" size-class="w-5 h-5" :override-color="getCategoryTheme(cat.color_key).text" />
                </div>

                <!-- Info -->
                <div class="min-w-0 flex-1">
                  <div class="flex flex-wrap items-center gap-1.5">
                    <span class="text-xs font-bold text-slate-900 dark:text-white truncate">{{ cat.name }}</span>
                    <span class="text-4xs font-mono px-1.5 py-0.5 rounded bg-slate-100 dark:bg-[#18283D] text-slate-600 dark:text-slate-400 border border-slate-200 dark:border-slate-700">
                      {{ cat.code }}
                    </span>
                    <span
                      class="text-4xs font-mono px-1.5 py-0.5 rounded font-bold uppercase"
                      :class="getDomainBadgeClass(cat.domain)"
                    >
                      {{ cat.domain || 'AQMS' }}
                    </span>
                    <span
                      class="text-4xs px-1.5 py-0.5 rounded font-bold uppercase"
                      :class="cat.enabled
                        ? 'bg-emerald-500/15 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400 border border-emerald-500/30'
                        : 'bg-slate-200 text-slate-600 dark:bg-slate-700/40 dark:text-slate-400 border border-slate-300 dark:border-slate-600'"
                    >
                      {{ cat.enabled ? 'Active' : 'Disabled' }}
                    </span>
                  </div>
                  <div class="flex items-center space-x-2.5 text-3xs text-slate-500 dark:text-slate-400 mt-1">
                    <span class="truncate max-w-[280px]">{{ cat.description || 'No description provided' }}</span>
                    <span class="text-slate-300 dark:text-slate-600">·</span>
                    <span class="text-blue-600 dark:text-cyan-400"><span class="font-mono tabular-nums font-semibold">{{ cat.parameter_count || 0 }}</span> parameters</span>
                  </div>
                </div>
              </div>

              <!-- Right: Column Span Selector & Action Buttons -->
              <div class="flex items-center space-x-2 flex-shrink-0 self-end sm:self-center text-2xs">
                <!-- Column Span Select -->
                <div class="flex items-center space-x-1 bg-slate-50 dark:bg-[#101B2B] px-2 py-1 rounded-lg border border-slate-200 dark:border-[#20354B]">
                  <span class="text-3xs text-slate-500 dark:text-slate-400">Span:</span>
                  <select
                    v-model="cat.default_column_span"
                    @change="onSpanChanged(cat)"
                    :disabled="!canManage"
                    class="bg-transparent text-blue-700 dark:text-cyan-300 font-bold focus:outline-none text-3xs disabled:opacity-60"
                  >
                    <option :value="1">1 Col</option>
                    <option :value="2">2 Cols</option>
                    <option :value="3">3 Cols</option>
                    <option :value="4">4 Cols (Full)</option>
                  </select>
                </div>

                <!-- Enable / Disable Toggle -->
                <button
                  v-if="canManage"
                  type="button"
                  @click="toggleCategoryEnabled(cat)"
                  class="px-2.5 py-1 rounded-lg text-3xs font-semibold transition border"
                  :class="cat.enabled
                    ? 'bg-slate-100 text-slate-700 border-slate-200 hover:bg-slate-200 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700 dark:hover:text-white'
                    : 'bg-emerald-500/15 text-emerald-700 border-emerald-500/30 dark:bg-emerald-500/20 dark:text-emerald-300 dark:border-emerald-500/40'"
                  :title="cat.enabled ? 'Disable Category' : 'Enable Category'"
                >
                  {{ cat.enabled ? 'Disable' : 'Enable' }}
                </button>

                <!-- Edit Button -->
                <button
                  v-if="canManage"
                  type="button"
                  @click="openEditForm(cat)"
                  class="px-2.5 py-1 rounded-lg bg-blue-500/10 text-blue-700 hover:bg-blue-500/20 dark:text-blue-400 border border-blue-500/30 text-3xs font-semibold transition"
                >
                  {{ $t('categoryManagement.actions.edit') || 'Edit' }}
                </button>

                <!-- Delete Button -->
                <button
                  v-if="canManage"
                  type="button"
                  @click="openDeleteDialog(cat)"
                  class="px-2.5 py-1 rounded-lg bg-rose-500/10 text-rose-700 hover:bg-rose-500/20 dark:text-rose-400 border border-rose-500/30 text-3xs font-semibold transition"
                >
                  {{ $t('categoryManagement.actions.delete') || 'Delete' }}
                </button>
              </div>
            </div>
          </div>

          <!-- Bottom Save Order Button if reordered -->
          <div v-if="hasReorderChanges && canManage" class="flex justify-end pt-2">
            <button
              type="button"
              @click="saveReorderedCategories"
              :disabled="isSaving"
              class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white dark:bg-cyan-600 dark:hover:bg-cyan-500 dark:text-slate-950 font-bold text-2xs transition flex items-center space-x-1.5 shadow-sm"
            >
              <span>💾</span>
              <span>{{ isSaving ? $t('categoryManagement.actions.saving') : $t('categoryManagement.actions.reorder') || 'Save Order' }}</span>
            </button>
          </div>
        </div>

        <!-- ==================================================== -->
        <!-- TAB 2: PARAMETER ASSIGNMENT MATRIX                    -->
        <!-- ==================================================== -->
        <div v-if="activeTab === 'assignments'" class="space-y-3">
          <div class="flex flex-wrap items-center justify-between gap-2 text-2xs">
            <input
              type="text"
              v-model="assignSearch"
              placeholder="Search parameters or devices..."
              class="px-3 py-1.5 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-lg text-slate-900 dark:text-white text-3xs focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500 min-w-[240px]"
            />
            <div class="text-3xs text-slate-500 dark:text-slate-400">
              Total Parameters: <span class="font-mono tabular-nums font-semibold text-slate-800 dark:text-slate-200">{{ allParametersList.length }}</span> | Uncategorized: <span class="font-mono tabular-nums font-semibold text-amber-600 dark:text-amber-400">{{ uncategorizedCount }}</span>
            </div>
          </div>

          <div class="border border-slate-200 dark:border-[#20354B] rounded-xl overflow-hidden bg-white dark:bg-[#080E18]">
            <div class="max-h-[420px] overflow-y-auto custom-scroller">
              <table class="w-full text-left text-3xs">
                <thead class="bg-slate-100 dark:bg-[#101B2B] text-slate-600 dark:text-slate-400 uppercase tracking-wider sticky top-0 border-b border-slate-200 dark:border-[#20354B]">
                  <tr>
                    <th class="p-2.5">Device</th>
                    <th class="p-2.5">Parameter</th>
                    <th class="p-2.5">Current Category</th>
                    <th class="p-2.5 text-right">Assign Category</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-slate-100 dark:divide-[#18283D] text-slate-800 dark:text-slate-200">
                  <tr v-if="filteredAssignParams.length === 0">
                    <td colspan="4" class="p-6 text-center text-slate-400">
                      No parameters available to assign.
                    </td>
                  </tr>
                  <tr
                    v-for="param in filteredAssignParams"
                    :key="param.id"
                    class="hover:bg-slate-50 dark:hover:bg-[#101B2B]/60 transition"
                  >
                    <td class="p-2.5 text-blue-600 dark:text-cyan-400/90 font-bold font-mono whitespace-nowrap">
                      {{ param.device_code || ('DEV-' + param.device_id) }}
                    </td>
                    <td class="p-2.5">
                      <div class="flex items-center space-x-1.5">
                        <ParameterIcon :icon="param.icon" size-class="w-3.5 h-3.5 flex-shrink-0" />
                        <span class="font-bold text-slate-900 dark:text-white">{{ param.parameter_name || param.parameter_code }}</span>
                        <span class="text-slate-500 text-4xs font-mono">({{ param.parameter_code }})</span>
                      </div>
                    </td>
                    <td class="p-2.5 whitespace-nowrap">
                      <span
                        v-if="param.category"
                        class="px-2 py-0.5 rounded text-4xs font-bold border"
                        :class="getCategoryTheme(param.category.color_key).pill"
                      >
                        {{ param.category.name }}
                      </span>
                      <span v-else class="text-slate-400 dark:text-slate-500 text-4xs italic">Uncategorized</span>
                    </td>
                    <td class="p-2.5 text-right whitespace-nowrap">
                      <select
                        :value="param.category_id || ''"
                        @change="onParamCategoryChanged(param, $event.target.value)"
                        :disabled="!canManage"
                        class="px-2 py-1 bg-white dark:bg-[#101B2B] border border-slate-300 dark:border-[#20354B] rounded text-slate-900 dark:text-white text-3xs focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500"
                      >
                        <option value="">— Uncategorized —</option>
                        <option v-for="c in categories" :key="c.id" :value="c.id">
                          {{ c.name }} {{ !c.enabled ? '(Disabled)' : '' }}
                        </option>
                      </select>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <!-- ==================================================== -->
        <!-- FORM MODAL: CREATE / EDIT CATEGORY                    -->
        <!-- ==================================================== -->
        <div v-if="isFormOpen" class="fixed inset-0 z-[80] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-fade-in font-sans">
          <div class="bg-white dark:bg-[#0B1422] border border-slate-200 dark:border-[#20354B] rounded-2xl w-full max-w-2xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
            <!-- Form Header -->
            <div class="px-6 py-3.5 border-b border-slate-200 dark:border-[#20354B] flex items-center justify-between bg-slate-50 dark:bg-[#080E18]">
              <div class="flex items-center space-x-2">
                <span class="w-2.5 h-2.5 rounded-full" :class="getCategoryTheme(form.color_key).dot"></span>
                <h3 class="text-xs font-bold text-slate-900 dark:text-white uppercase tracking-wider">
                  {{ editingId ? ($t('categoryManagement.tabs.edit') || 'Edit Category') : ($t('categoryManagement.tabs.create') || 'Create Category') }}
                </h3>
              </div>
              <button type="button" @click="isFormOpen = false" class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1 rounded-lg">✕</button>
            </div>

            <!-- Form Content -->
            <form @submit.prevent="submitCategoryForm" class="flex-1 overflow-y-auto p-6 space-y-4 text-2xs custom-scroller">
              <!-- Inline Form Error -->
              <div v-if="formErrorMessage" class="p-2.5 bg-rose-50 dark:bg-rose-500/15 border border-rose-300 dark:border-rose-500/30 rounded-xl text-3xs text-rose-700 dark:text-rose-300">
                ⚠️ {{ formErrorMessage }}
              </div>

              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <!-- Category Display Name -->
                <div>
                  <label class="block text-slate-700 dark:text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.name') || 'Display Name' }} <span class="text-rose-500">*</span>
                  </label>
                  <input
                    v-model="form.name"
                    @input="onFormNameInput"
                    type="text"
                    required
                    placeholder="e.g. Water Quality"
                    class="w-full px-3 py-2 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-xl text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500"
                  />
                  <span class="text-4xs text-slate-500 dark:text-slate-400 mt-1 block">Display label shown in dashboard & dropdown</span>
                </div>

                <!-- Category Code -->
                <div>
                  <label class="block text-slate-700 dark:text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.code') || 'Category Code' }} <span class="text-rose-500">*</span>
                  </label>
                  <input
                    v-model="form.code"
                    @input="codeManuallyEdited = true"
                    type="text"
                    required
                    placeholder="e.g. water_quality"
                    class="w-full px-3 py-2 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-xl text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500"
                  />
                  <span class="text-4xs text-slate-500 dark:text-slate-400 mt-1 block">Unique code (letters, numbers, _ -)</span>
                </div>

                <!-- Monitoring Domain -->
                <div>
                  <label class="block text-slate-700 dark:text-slate-300 font-semibold mb-1">
                    Monitoring Domain
                  </label>
                  <select
                    v-model="form.domain"
                    class="w-full px-3 py-2 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-xl text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500 uppercase font-mono"
                  >
                    <option value="AQMS">AQMS (Air Quality)</option>
                    <option value="WQMS">WQMS (Water Quality)</option>
                    <option value="SPARING">SPARING (Wastewater / Air Limbah)</option>
                    <option value="CUSTOM">CUSTOM (Configurable)</option>
                    <option value="ALL">ALL (Universal)</option>
                  </select>
                  <span class="text-4xs text-slate-500 dark:text-slate-400 mt-1 block">Target monitoring domain profile</span>
                </div>
              </div>

              <!-- Description -->
              <div>
                <label class="block text-slate-700 dark:text-slate-300 font-semibold mb-1">
                  {{ $t('categoryManagement.fields.description') || 'Description' }}
                </label>
                <textarea
                  v-model="form.description"
                  rows="2"
                  placeholder="Operational purpose of this monitoring domain..."
                  class="w-full px-3 py-2 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-xl text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500"
                ></textarea>
              </div>

              <!-- Vector Icon & Color Picker Row -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Icon Picker -->
                <div>
                  <label class="block text-slate-700 dark:text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.icon') || 'Category Vector Icon' }}
                  </label>
                  <div class="flex items-center space-x-2">
                    <select
                      v-model="form.icon_key"
                      class="flex-1 px-3 py-2 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-xl text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500 capitalize"
                    >
                      <option v-for="ic in iconCatalogList" :key="ic.key" :value="ic.key">
                        {{ ic.label }}
                      </option>
                    </select>
                    <div class="w-9 h-9 rounded-xl flex items-center justify-center border bg-slate-50 dark:bg-[#101B2B] border-slate-200 dark:border-[#20354B] flex-shrink-0">
                      <ParameterIcon :icon="form.icon_key" size-class="w-4 h-4 text-blue-600 dark:text-cyan-400" />
                    </div>
                  </div>
                </div>

                <!-- Theme Color Picker -->
                <div>
                  <label class="block text-slate-700 dark:text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.color') || 'Theme Accent Color' }}
                  </label>
                  <div class="flex items-center space-x-2">
                    <select
                      v-model="form.color_key"
                      class="flex-1 px-3 py-2 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-xl text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500 capitalize"
                    >
                      <option v-for="c in colorOptions" :key="c.key" :value="c.key">
                        {{ c.name }}
                      </option>
                    </select>
                    <div
                      class="w-9 h-9 rounded-xl flex items-center justify-center border flex-shrink-0"
                      :class="[getCategoryTheme(form.color_key).bg, getCategoryTheme(form.color_key).border]"
                    >
                      <span class="w-3 h-3 rounded-full" :class="getCategoryTheme(form.color_key).dot"></span>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Column Span & Enabled State -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label class="block text-slate-700 dark:text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.span') || 'Default Column Span' }}
                  </label>
                  <select
                    v-model="form.default_column_span"
                    class="w-full px-3 py-2 bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] rounded-xl text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 dark:focus:border-cyan-500"
                  >
                    <option :value="1">1 Column (Standard 3-4 Grid)</option>
                    <option :value="2">2 Columns (Double Width Span)</option>
                    <option :value="3">3 Columns (Wide CBI Panel)</option>
                    <option :value="4">4 Columns (Full Desktop Row)</option>
                  </select>
                </div>

                <div class="flex items-center space-x-3 pt-6">
                  <label class="flex items-center space-x-2 cursor-pointer">
                    <input
                      type="checkbox"
                      v-model="form.enabled"
                      class="w-4 h-4 rounded text-blue-600 dark:text-cyan-500 focus:ring-0 bg-white dark:bg-[#080E18] border-slate-300 dark:border-[#20354B]"
                    />
                    <span class="text-slate-700 dark:text-slate-300 font-semibold">{{ $t('categoryManagement.fields.enabled') || 'Category Enabled' }}</span>
                  </label>
                </div>
              </div>

              <!-- LIVE PREVIEW SECTION -->
              <div class="p-3.5 bg-slate-50 dark:bg-[#080E18] border border-slate-200 dark:border-[#20354B] rounded-xl space-y-2 mt-2">
                <span class="text-4xs text-slate-500 dark:text-slate-400 font-bold uppercase tracking-wider block">
                  {{ $t('categoryManagement.preview.sectionHeader') || 'Live Operator Command Center Preview' }}
                </span>

                <!-- Category Section Banner Preview -->
                <div
                  class="p-2 rounded-lg border flex items-center justify-between"
                  :class="[getCategoryTheme(form.color_key).bg, getCategoryTheme(form.color_key).border]"
                >
                  <div class="flex items-center space-x-2 min-w-0">
                    <span class="w-2 h-2 rounded-full" :class="getCategoryTheme(form.color_key).dot"></span>
                    <ParameterIcon :icon="form.icon_key" size-class="w-4 h-4" :override-color="getCategoryTheme(form.color_key).text" />
                    <span class="text-xs font-bold uppercase tracking-wider" :class="getCategoryTheme(form.color_key).text">
                      {{ form.name || 'Category Name' }}
                    </span>
                    <span class="text-4xs px-1.5 py-0.2 rounded-full bg-white/80 dark:bg-[#18283D] text-slate-700 dark:text-slate-300 border border-slate-200 dark:border-slate-700">
                      4 parameters
                    </span>
                  </div>
                  <span class="text-3xs text-slate-500 dark:text-slate-400">{{ form.default_column_span }} Col Span</span>
                </div>
              </div>

              <!-- Form Buttons -->
              <div class="flex justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
                <button
                  type="button"
                  @click="isFormOpen = false"
                  class="px-4 py-2 rounded-xl bg-slate-100 hover:bg-slate-200 text-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 transition text-2xs font-medium"
                >
                  {{ $t('categoryManagement.actions.cancel') || 'Cancel' }}
                </button>
                <button
                  type="submit"
                  :disabled="isSaving"
                  class="px-5 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white dark:bg-cyan-600 dark:hover:bg-cyan-500 dark:text-slate-950 font-bold transition text-2xs shadow-sm"
                >
                  {{ isSaving ? $t('categoryManagement.actions.saving') : (editingId ? ($t('categoryManagement.actions.save') || 'Save Changes') : ($t('categoryManagement.actions.add') || 'Create Category')) }}
                </button>
              </div>
            </form>
          </div>
        </div>

        <!-- ==================================================== -->
        <!-- SAFE DELETE CONFIRMATION MODAL                        -->
        <!-- ==================================================== -->
        <div v-if="deleteTarget" class="fixed inset-0 z-[90] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-fade-in font-sans">
          <div class="bg-white dark:bg-[#0B1422] border border-rose-300 dark:border-rose-500/40 rounded-2xl w-full max-w-lg p-6 shadow-2xl space-y-4">
            <div class="flex items-center space-x-2 text-rose-600 dark:text-rose-400">
              <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <h3 class="text-sm font-bold text-slate-900 dark:text-white uppercase tracking-wider">
                {{ $t('categoryManagement.deleteModal.title') || 'Delete Category' }}
              </h3>
            </div>

            <p class="text-2xs text-slate-700 dark:text-slate-300 leading-relaxed">
              {{ $t('categoryManagement.deleteModal.warning', { name: deleteTarget.name }) || `Are you sure you want to delete category '${deleteTarget.name}'?` }}
            </p>

            <div v-if="deleteTarget.parameter_count > 0" class="p-3 bg-slate-50 dark:bg-[#080E18] border border-slate-200 dark:border-[#20354B] rounded-xl space-y-2 text-2xs">
              <span class="text-blue-700 dark:text-cyan-400 font-semibold block">
                {{ $t('categoryManagement.deleteModal.reassignPrompt', { count: deleteTarget.parameter_count }) || `Action for ${deleteTarget.parameter_count} assigned parameter(s):` }}
              </span>
              <label class="flex items-start space-x-2 cursor-pointer">
                <input type="radio" v-model="deleteReassignMode" value="unassign" class="mt-0.5 text-blue-600 dark:text-cyan-500" />
                <span class="text-slate-700 dark:text-slate-300 text-3xs">{{ $t('categoryManagement.deleteModal.optionUnassign') || 'Keep as Uncategorized (Safely unassigns without deleting parameters)' }}</span>
              </label>
              <label v-if="categories.filter(c => c.id !== deleteTarget.id).length > 0" class="flex items-start space-x-2 cursor-pointer">
                <input type="radio" v-model="deleteReassignMode" value="reassign" class="mt-0.5 text-blue-600 dark:text-cyan-500" />
                <span class="text-slate-700 dark:text-slate-300 text-3xs">{{ $t('categoryManagement.deleteModal.optionReassign') || 'Reassign to another category:' }}</span>
              </label>
              <select
                v-if="deleteReassignMode === 'reassign'"
                v-model="deleteTargetReassignId"
                class="w-full mt-1 px-2.5 py-1.5 bg-white dark:bg-[#101B2B] border border-slate-300 dark:border-[#20354B] rounded-lg text-slate-900 dark:text-white text-3xs focus:outline-none"
              >
                <option v-for="c in categories.filter(c => c.id !== deleteTarget.id)" :key="c.id" :value="c.id">
                  {{ c.name }}
                </option>
              </select>
            </div>

            <div class="p-2.5 bg-emerald-50 dark:bg-emerald-500/10 border border-emerald-200 dark:border-emerald-500/20 rounded-xl text-3xs text-emerald-800 dark:text-emerald-300">
              🛡️ {{ $t('categoryManagement.deleteModal.safetyNotice') || 'Data Safety: Deleting a category never deletes parameters or telemetry history.' }}
            </div>

            <div class="flex justify-end space-x-2 pt-2 border-t border-slate-200 dark:border-[#20354B]">
              <button
                type="button"
                @click="deleteTarget = null"
                class="px-3.5 py-1.5 rounded-lg bg-slate-100 hover:bg-slate-200 text-slate-700 dark:bg-slate-800 dark:text-slate-300 text-2xs dark:hover:bg-slate-700"
              >
                {{ $t('categoryManagement.actions.cancel') || 'Cancel' }}
              </button>
              <button
                type="button"
                @click="confirmDeleteCategory"
                :disabled="isSaving"
                class="px-4 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 text-white font-bold text-2xs transition"
              >
                {{ isSaving ? 'Deleting...' : ($t('categoryManagement.deleteModal.confirm') || 'Confirm Delete') }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import ParameterIcon from './ParameterIcon.vue';
import { PARAMETER_ICON_CATALOG } from '../utils/parameterIcons.js';
import { CATEGORY_COLORS, COLOR_LIST, getCategoryColor } from '../utils/categoryTheme.js';
import apiClient from '../services/api.js';

export default {
  name: 'CategoryManagementModal',
  components: {
    ParameterIcon,
  },
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
    initialCategories: {
      type: Array,
      default: () => [],
    },
    devicesList: {
      type: Array,
      default: () => [],
    },
    devices: {
      type: Array,
      default: () => [],
    },
  },
  data() {
    return {
      activeTab: 'list',
      categories: [],
      fallbackDevices: [],
      hasReorderChanges: false,
      isSaving: false,
      noticeMessage: '',
      errorMessage: '',
      formErrorMessage: '',
      categorySearch: '',
      assignSearch: '',

      // Form state
      isFormOpen: false,
      editingId: null,
      codeManuallyEdited: false,
      form: {
        code: '',
        name: '',
        description: '',
        icon_key: 'activity',
        color_key: 'cyan',
        default_column_span: 1,
        enabled: true,
        domain: 'AQMS',
      },

      // Delete confirmation state
      deleteTarget: null,
      deleteReassignMode: 'unassign',
      deleteTargetReassignId: null,
    };
  },
  computed: {
    canManage() {
      const user = this.$store && this.$store.state ? this.$store.state.user : null;
      if (!user) return false;
      const rawRole = typeof user.role === 'string'
        ? user.role
        : ((user.role && user.role.name) || user.role_name || (user.Role && user.Role.name) || '');
      const normalized = String(rawRole).trim().toUpperCase();
      return Boolean(normalized && normalized !== 'OPERATOR');
    },
    iconCatalogList() {
      return PARAMETER_ICON_CATALOG;
    },
    colorOptions() {
      return COLOR_LIST;
    },
    filteredCategories() {
      const q = (this.categorySearch || '').toLowerCase().trim();
      if (!q) return this.categories;
      return this.categories.filter(c =>
        (c.name || '').toLowerCase().includes(q) ||
        (c.code || '').toLowerCase().includes(q) ||
        (c.domain || '').toLowerCase().includes(q) ||
        (c.description || '').toLowerCase().includes(q)
      );
    },
    effectiveDevicesList() {
      if (Array.isArray(this.devicesList) && this.devicesList.length > 0) return this.devicesList;
      if (Array.isArray(this.devices) && this.devices.length > 0) return this.devices;
      if (this.$store && this.$store.state && Array.isArray(this.$store.state.devices) && this.$store.state.devices.length > 0) {
        return this.$store.state.devices;
      }
      return this.fallbackDevices || [];
    },
    allParametersList() {
      const list = [];
      for (const dev of this.effectiveDevicesList || []) {
        for (const p of dev.parameters || []) {
          const pid = p.id || p.param_id;
          list.push({
            ...p,
            id: pid,
            parameter_code: p.parameter_code || p.param_code,
            parameter_name: p.parameter_name || p.param_name,
            device_id: dev.id || dev.device_id,
            device_code: dev.device_code || dev.code,
            device_name: dev.device_name || dev.name,
            category: this.categories.find(c => Number(c.id) === Number(p.category_id)) || null,
          });
        }
      }
      return list;
    },
    uncategorizedCount() {
      return this.allParametersList.filter(p => !p.category_id).length;
    },
    filteredAssignParams() {
      let list = this.allParametersList;
      if (this.assignSearch && this.assignSearch.trim()) {
        const q = this.assignSearch.toLowerCase().trim();
        list = list.filter(p =>
          (p.parameter_code || '').toLowerCase().includes(q) ||
          (p.parameter_name || '').toLowerCase().includes(q) ||
          (p.device_code || '').toLowerCase().includes(q)
        );
      }
      return list;
    },
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.errorMessage = '';
        this.formErrorMessage = '';
        this.fetchCategories();
        if (this.effectiveDevicesList.length === 0) {
          this.fetchFallbackDevices();
        }
      }
    },
    initialCategories: {
      immediate: true,
      handler(val) {
        if (Array.isArray(val) && val.length > 0 && !this.isOpen) {
          this.categories = JSON.parse(JSON.stringify(val));
        }
      },
    },
  },
  methods: {
    getCategoryTheme(colorKey) {
      return getCategoryColor(colorKey);
    },
    slugifyCategoryCode(name) {
      return String(name || '')
        .toLowerCase()
        .trim()
        .replace(/[^a-z0-9]+/g, '_')
        .replace(/^_+|_+$/g, '')
        .slice(0, 60);
    },
    onFormNameInput() {
      if (!this.editingId && !this.codeManuallyEdited) {
        this.form.code = this.slugifyCategoryCode(this.form.name);
      }
    },
    closeModal() {
      this.isFormOpen = false;
      this.deleteTarget = null;
      this.$emit('close');
    },
    async fetchFallbackDevices() {
      try {
        const res = await apiClient.get('/devices', { params: { page_size: 100 } });
        const list = res.data?.data?.items || res.data?.data || res.data || [];
        if (Array.isArray(list)) {
          this.fallbackDevices = list;
        }
      } catch (e) {
        // Ignore silent fallback error
      }
    },
    async fetchCategories() {
      try {
        const res = await apiClient.get('/categories');
        const list = res.data?.data !== undefined ? res.data.data : res.data;
        if (Array.isArray(list)) {
          this.categories = list;
          this.hasReorderChanges = false;
          this.$emit('categories-updated', this.categories);
        }
      } catch (err) {
        console.warn('Failed to load categories:', err);
      }
    },
    moveCategoryUp(idx) {
      if (!this.canManage || idx <= 0) return;
      const item = this.categories.splice(idx, 1)[0];
      this.categories.splice(idx - 1, 0, item);
      this.hasReorderChanges = true;
    },
    moveCategoryDown(idx) {
      if (!this.canManage || idx >= this.categories.length - 1) return;
      const item = this.categories.splice(idx, 1)[0];
      this.categories.splice(idx + 1, 0, item);
      this.hasReorderChanges = true;
    },
    async onSpanChanged(cat) {
      if (!this.canManage) return;
      try {
        await apiClient.put(`/categories/${cat.id}`, {
          default_column_span: cat.default_column_span,
        });
        this.showNotice(`Updated span for '${cat.name}'`);
        this.$emit('categories-updated', this.categories);
        this.$emit('saved', this.categories);
      } catch (err) {
        this.showError('Failed to update span: ' + (err.response?.data?.error || err.message));
      }
    },
    async saveReorderedCategories() {
      if (!this.canManage) return;
      this.isSaving = true;
      try {
        const ids = this.categories.map(c => c.id);
        await apiClient.put('/categories/reorder', { category_ids: ids });
        this.hasReorderChanges = false;
        this.showNotice(this.$t('categoryManagement.notifications.reordered') || 'Category order saved successfully');
        this.$emit('categories-updated', this.categories);
        this.$emit('saved', this.categories);
      } catch (err) {
        this.showError('Failed to save order: ' + (err.response?.data?.error || err.message));
      } finally {
        this.isSaving = false;
      }
    },
    async toggleCategoryEnabled(cat) {
      if (!this.canManage) return;
      const targetState = !cat.enabled;
      try {
        await apiClient.put(`/categories/${cat.id}/toggle`, { enabled: targetState });
        cat.enabled = targetState;
        this.showNotice(`Category '${cat.name}' ${targetState ? 'enabled' : 'disabled'}`);
        this.$emit('categories-updated', this.categories);
        this.$emit('saved', this.categories);
      } catch (err) {
        this.showError('Failed to toggle category: ' + (err.response?.data?.error || err.message));
      }
    },
    getDomainBadgeClass(domain) {
      const d = (domain || 'AQMS').toUpperCase();
      if (d === 'AQMS') return 'bg-cyan-500/15 text-cyan-700 dark:bg-cyan-500/20 dark:text-cyan-400 border border-cyan-500/30';
      if (d === 'WQMS') return 'bg-blue-500/15 text-blue-700 dark:bg-blue-500/20 dark:text-blue-400 border border-blue-500/30';
      if (d === 'SPARING') return 'bg-teal-500/15 text-teal-700 dark:bg-teal-500/20 dark:text-teal-300 border border-teal-500/30';
      if (d === 'CUSTOM') return 'bg-purple-500/15 text-purple-700 dark:bg-purple-500/20 dark:text-purple-400 border border-purple-500/30';
      return 'bg-emerald-500/15 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-400 border border-emerald-500/30';
    },
    defaultDomainFromSystem() {
      const sysDom = (this.$store?.state?.systemStatus?.monitoring_domain || 'AQMS').toUpperCase();
      if (['AQMS', 'WQMS', 'SPARING', 'CUSTOM', 'ALL'].includes(sysDom)) {
        return sysDom;
      }
      return 'AQMS';
    },
    openCreateForm() {
      if (!this.canManage) return;
      this.editingId = null;
      this.codeManuallyEdited = false;
      this.formErrorMessage = '';
      this.form = {
        code: '',
        name: '',
        description: '',
        icon_key: 'activity',
        color_key: 'cyan',
        default_column_span: 1,
        enabled: true,
        domain: this.defaultDomainFromSystem(),
      };
      this.isFormOpen = true;
    },
    openEditForm(cat) {
      if (!this.canManage) return;
      this.editingId = cat.id;
      this.codeManuallyEdited = true;
      this.formErrorMessage = '';
      this.form = {
        code: cat.code,
        name: cat.name,
        description: cat.description || '',
        icon_key: cat.icon_key || 'activity',
        color_key: cat.color_key || 'cyan',
        default_column_span: cat.default_column_span || 1,
        enabled: cat.enabled !== false,
        domain: cat.domain || 'AQMS',
      };
      this.isFormOpen = true;
    },
    async submitCategoryForm() {
      if (!this.canManage) return;
      this.isSaving = true;
      this.errorMessage = '';
      this.formErrorMessage = '';

      const payloadCode = (this.form.code || this.slugifyCategoryCode(this.form.name)).trim();
      const payloadName = (this.form.name || '').trim();
      if (!payloadName || !payloadCode) {
        this.formErrorMessage = 'Category Name and Category Code are required.';
        this.isSaving = false;
        return;
      }

      try {
        let createdCat = null;
        if (this.editingId) {
          await apiClient.put(`/categories/${this.editingId}`, {
            code: payloadCode,
            name: payloadName,
            description: this.form.description,
            icon_key: this.form.icon_key,
            color_key: this.form.color_key,
            default_column_span: Number(this.form.default_column_span) || 1,
            enabled: this.form.enabled,
            domain: this.form.domain,
          });
          this.showNotice(this.$t('categoryManagement.notifications.updated', { name: payloadName }) || `Updated category '${payloadName}'`);
        } else {
          const res = await apiClient.post('/categories', {
            ...this.form,
            code: payloadCode,
            name: payloadName,
            default_column_span: Number(this.form.default_column_span) || 1,
          });
          createdCat = res.data?.data || res.data || null;
          this.showNotice(this.$t('categoryManagement.notifications.created', { name: payloadName }) || `Created category '${payloadName}'`);
        }
        this.isFormOpen = false;
        await this.fetchCategories();
        if (createdCat && createdCat.id) {
          this.$emit('category-created', createdCat);
        }
        this.$emit('saved', this.categories);
      } catch (err) {
        this.formErrorMessage = err.response?.data?.error || err.message;
      } finally {
        this.isSaving = false;
      }
    },
    openDeleteDialog(cat) {
      if (!this.canManage) return;
      this.deleteTarget = cat;
      this.deleteReassignMode = 'unassign';
      const other = this.categories.find(c => c.id !== cat.id);
      this.deleteTargetReassignId = other ? other.id : null;
    },
    async confirmDeleteCategory() {
      if (!this.canManage || !this.deleteTarget) return;
      this.isSaving = true;
      try {
        let url = `/categories/${this.deleteTarget.id}`;
        if (this.deleteReassignMode === 'reassign' && this.deleteTargetReassignId) {
          url += `?reassign_to=${this.deleteTargetReassignId}`;
        }
        await apiClient.delete(url);
        this.showNotice(this.$t('categoryManagement.notifications.deleted') || 'Category deleted successfully');
        this.deleteTarget = null;
        await this.fetchCategories();
        this.$emit('saved', this.categories);
        this.$emit('refresh-devices');
      } catch (err) {
        this.showError(err.response?.data?.error || err.message);
      } finally {
        this.isSaving = false;
      }
    },
    async onParamCategoryChanged(param, newCatId) {
      if (!this.canManage) return;
      try {
        const catId = (newCatId !== null && newCatId !== undefined && newCatId !== '' && Number(newCatId) > 0) ? Number(newCatId) : null;
        const targetRoute = catId ? `/categories/${catId}/parameters` : '/categories/unassign/parameters';
        await apiClient.post(targetRoute, {
          parameter_ids: [Number(param.id)],
        });
        param.category_id = catId;
        param.category = this.categories.find(c => Number(c.id) === Number(catId)) || null;
        this.showNotice(`Assigned '${param.parameter_name || param.parameter_code}' to ${param.category ? param.category.name : 'Uncategorized'}`);
        await this.fetchCategories();
        this.$emit('saved', this.categories);
        this.$emit('refresh-devices');
      } catch (err) {
        this.showError('Failed to assign parameter: ' + (err.response?.data?.error || err.message));
      }
    },
    showNotice(msg) {
      this.noticeMessage = msg;
      setTimeout(() => {
        if (this.noticeMessage === msg) this.noticeMessage = '';
      }, 4000);
    },
    showError(msg) {
      this.errorMessage = msg;
      setTimeout(() => {
        if (this.errorMessage === msg) this.errorMessage = '';
      }, 6000);
    },
  },
};
</script>
