<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-[#0B1422] border border-[#20354B] rounded-2xl w-full max-w-4xl max-h-[92vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-[#20354B] flex items-center justify-between bg-[#080E18]">
        <div class="flex items-center space-x-2.5">
          <div class="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
            </svg>
          </div>
          <div>
            <h2 class="text-sm font-bold text-white tracking-wide font-mono uppercase">
              {{ $t('categoryManagement.title') || 'PARAMETER CATEGORY MANAGEMENT' }}
            </h2>
            <p class="text-3xs text-slate-400 mt-0.5">
              {{ $t('categoryManagement.subtitle') || 'Configure domain categories, display ordering, responsive spans, and assignments.' }}
            </p>
          </div>
        </div>

        <button @click="closeModal" class="text-slate-400 hover:text-white p-1.5 rounded-lg hover:bg-[#1E2E42] transition">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Tab Strip & Actions Bar -->
      <div class="px-6 py-2.5 bg-[#0e192a] border-b border-[#20354B] flex flex-wrap items-center justify-between gap-3 text-2xs font-mono">
        <div class="flex items-center space-x-1.5 bg-[#080E18] p-1 rounded-xl border border-[#20354B]">
          <button
            @click="activeTab = 'list'"
            class="px-3 py-1.5 rounded-lg transition font-semibold flex items-center space-x-1.5"
            :class="activeTab === 'list' ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-xs' : 'text-slate-400 hover:text-white'"
          >
            <span>▦</span>
            <span>{{ $t('categoryManagement.tabs.categories') || 'Categories & Layout' }}</span>
            <span class="px-1.5 py-0.2 rounded-full bg-[#18283D] text-slate-300 text-3xs">{{ categories.length }}</span>
          </button>

          <button
            @click="activeTab = 'assignments'"
            class="px-3 py-1.5 rounded-lg transition font-semibold flex items-center space-x-1.5"
            :class="activeTab === 'assignments' ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/40 shadow-xs' : 'text-slate-400 hover:text-white'"
          >
            <span>⇄</span>
            <span>{{ $t('categoryManagement.tabs.assignments') || 'Parameter Assignments' }}</span>
          </button>
        </div>

        <button
          v-if="canManage"
          @click="openCreateForm"
          class="px-3 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold transition flex items-center space-x-1 shadow-md shadow-cyan-900/20"
        >
          <span>+</span>
          <span>{{ $t('categoryManagement.actions.add') || 'Add Category' }}</span>
        </button>
      </div>

      <!-- Main Body Area -->
      <div class="flex-1 overflow-y-auto p-6 space-y-4 custom-scroller">
        <!-- Notification Banner -->
        <div v-if="noticeMessage" class="p-3 bg-emerald-500/15 border border-emerald-500/30 rounded-xl text-xs text-emerald-300 flex items-center justify-between animate-fade-in">
          <span>✓ {{ noticeMessage }}</span>
          <button @click="noticeMessage = ''" class="text-emerald-400 hover:text-white text-3xs">✕</button>
        </div>

        <!-- Error Alert -->
        <div v-if="errorMessage" class="p-3 bg-rose-500/15 border border-rose-500/30 rounded-xl text-xs text-rose-300 flex items-center justify-between animate-fade-in">
          <span>⚠️ {{ errorMessage }}</span>
          <button @click="errorMessage = ''" class="text-rose-400 hover:text-white text-3xs">✕</button>
        </div>

        <!-- ==================================================== -->
        <!-- TAB 1: CATEGORY LIST & LAYOUT REORDERING              -->
        <!-- ==================================================== -->
        <div v-if="activeTab === 'list'" class="space-y-3">
          <div class="text-3xs text-slate-400 font-mono flex items-center justify-between">
            <span>Arrange category display order and configure desktop column span (1 to 4 columns).</span>
            <span v-if="hasReorderChanges" class="text-amber-400 font-semibold animate-pulse">● Unsaved order changes</span>
          </div>

          <div class="space-y-2">
            <div
              v-for="(cat, idx) in categories"
              :key="cat.id"
              class="p-3 rounded-xl border transition flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-[#080E18] border-[#20354B] hover:border-slate-600"
              :class="!cat.enabled ? 'opacity-60 bg-slate-900/40' : ''"
            >
              <!-- Left: Order Controls, Icon, Name, Code, Span Badge -->
              <div class="flex items-center space-x-3 min-w-0 flex-1">
                <!-- Move Up / Down Buttons -->
                <div v-if="canManage" class="flex flex-col space-y-0.5 flex-shrink-0">
                  <button
                    @click="moveCategoryUp(idx)"
                    :disabled="idx === 0"
                    class="p-1 rounded hover:bg-[#1E2E42] text-slate-400 hover:text-white disabled:opacity-20 text-3xs"
                    :title="$t('categoryManagement.actions.moveUp') || 'Move Up'"
                  >
                    ▲
                  </button>
                  <button
                    @click="moveCategoryDown(idx)"
                    :disabled="idx === categories.length - 1"
                    class="p-1 rounded hover:bg-[#1E2E42] text-slate-400 hover:text-white disabled:opacity-20 text-3xs"
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
                  <div class="flex items-center space-x-2">
                    <span class="text-xs font-bold text-white font-mono truncate">{{ cat.name }}</span>
                    <span class="text-4xs font-mono px-1.5 py-0.5 rounded bg-[#18283D] text-slate-400 border border-slate-700">
                      {{ cat.code }}
                    </span>
                    <span
                      class="text-4xs font-mono px-1.5 py-0.5 rounded font-bold uppercase"
                      :class="getDomainBadgeClass(cat.domain)"
                    >
                      {{ cat.domain || 'AQMS' }}
                    </span>
                    <span
                      class="text-4xs font-mono px-1.5 py-0.5 rounded font-bold uppercase"
                      :class="cat.enabled ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-slate-700/40 text-slate-400 border border-slate-600'"
                    >
                      {{ cat.enabled ? 'Active' : 'Disabled' }}
                    </span>
                  </div>
                  <div class="flex items-center space-x-3 text-3xs text-slate-400 mt-1 font-mono">
                    <span class="truncate max-w-[280px]">{{ cat.description || 'No description provided' }}</span>
                    <span class="text-slate-600">·</span>
                    <span class="text-cyan-400">{{ cat.parameter_count || 0 }} parameters</span>
                  </div>
                </div>
              </div>

              <!-- Right: Column Span Selector & Action Buttons -->
              <div class="flex items-center space-x-2 flex-shrink-0 self-end sm:self-center font-mono text-2xs">
                <!-- Column Span Select -->
                <div class="flex items-center space-x-1 bg-[#101B2B] px-2 py-1 rounded-lg border border-[#20354B]">
                  <span class="text-3xs text-slate-400">Span:</span>
                  <select
                    v-model="cat.default_column_span"
                    @change="onSpanChanged(cat)"
                    :disabled="!canManage"
                    class="bg-transparent text-cyan-300 font-bold focus:outline-none text-3xs"
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
                  @click="toggleCategoryEnabled(cat)"
                  class="px-2 py-1 rounded-lg text-3xs font-semibold transition border"
                  :class="cat.enabled
                    ? 'bg-slate-800 text-slate-300 border-slate-700 hover:text-white'
                    : 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40'"
                  :title="cat.enabled ? 'Disable Category' : 'Enable Category'"
                >
                  {{ cat.enabled ? 'Disable' : 'Enable' }}
                </button>

                <!-- Edit Button -->
                <button
                  v-if="canManage"
                  @click="openEditForm(cat)"
                  class="px-2 py-1 rounded-lg bg-blue-500/10 text-blue-400 hover:bg-blue-500/20 border border-blue-500/30 text-3xs transition"
                >
                  {{ $t('categoryManagement.actions.edit') || 'Edit' }}
                </button>

                <!-- Delete Button -->
                <button
                  v-if="canManage"
                  @click="openDeleteDialog(cat)"
                  class="px-2 py-1 rounded-lg bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 border border-rose-500/30 text-3xs transition"
                >
                  {{ $t('categoryManagement.actions.delete') || 'Delete' }}
                </button>
              </div>
            </div>
          </div>

          <!-- Bottom Save Order Button if reordered -->
          <div v-if="hasReorderChanges && canManage" class="flex justify-end pt-2">
            <button
              @click="saveReorderedCategories"
              :disabled="isSaving"
              class="px-4 py-2 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-2xs transition font-mono flex items-center space-x-1.5"
            >
              <span>💾</span>
              <span>{{ isSaving ? $t('categoryManagement.actions.saving') : $t('categoryManagement.actions.reorder') || 'Save Order' }}</span>
            </button>
          </div>
        </div>

        <!-- ==================================================== -->
        <!-- TAB 2: PARAMETER ASSIGNMENT MATRIX                    -->
        <!-- ==================================================== -->
        <div v-if="activeTab === 'assignments'" class="space-y-3 font-mono">
          <div class="flex flex-wrap items-center justify-between gap-2 text-2xs">
            <input
              type="text"
              v-model="assignSearch"
              placeholder="Search parameters or devices..."
              class="px-3 py-1.5 bg-[#080E18] border border-[#20354B] rounded-lg text-white text-3xs focus:outline-none focus:border-cyan-500 min-w-[240px]"
            />
            <div class="text-3xs text-slate-400">
              Total Parameters: {{ allParametersList.length }} | Uncategorized: {{ uncategorizedCount }}
            </div>
          </div>

          <div class="border border-[#20354B] rounded-xl overflow-hidden bg-[#080E18]">
            <div class="max-h-[420px] overflow-y-auto custom-scroller">
              <table class="w-full text-left text-3xs">
                <thead class="bg-[#101B2B] text-slate-400 uppercase tracking-wider sticky top-0 border-b border-[#20354B]">
                  <tr>
                    <th class="p-2.5">Device</th>
                    <th class="p-2.5">Parameter</th>
                    <th class="p-2.5">Current Category</th>
                    <th class="p-2.5 text-right">Assign Category</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-[#18283D] text-slate-200">
                  <tr
                    v-for="param in filteredAssignParams"
                    :key="param.id"
                    class="hover:bg-[#101B2B]/60 transition"
                  >
                    <td class="p-2.5 text-cyan-400/90 font-bold whitespace-nowrap">
                      {{ param.device_code || ('DEV-' + param.device_id) }}
                    </td>
                    <td class="p-2.5">
                      <div class="flex items-center space-x-1.5">
                        <ParameterIcon :icon="param.icon" size-class="w-3.5 h-3.5 flex-shrink-0" />
                        <span class="font-bold text-white">{{ param.parameter_name || param.parameter_code }}</span>
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
                      <span v-else class="text-slate-500 text-4xs italic">Uncategorized</span>
                    </td>
                    <td class="p-2.5 text-right whitespace-nowrap">
                      <select
                        :value="param.category_id || ''"
                        @change="onParamCategoryChanged(param, $event.target.value)"
                        :disabled="!canManage"
                        class="px-2 py-1 bg-[#101B2B] border border-[#20354B] rounded text-white text-3xs focus:outline-none focus:border-cyan-500"
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
        <div v-if="isFormOpen" class="fixed inset-0 z-60 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-fade-in font-sans">
          <div class="bg-[#0B1422] border border-[#20354B] rounded-2xl w-full max-w-2xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
            <!-- Form Header -->
            <div class="px-6 py-3.5 border-b border-[#20354B] flex items-center justify-between bg-[#080E18]">
              <div class="flex items-center space-x-2">
                <span class="w-2.5 h-2.5 rounded-full" :class="getCategoryTheme(form.color_key).dot"></span>
                <h3 class="text-xs font-bold text-white font-mono uppercase tracking-wider">
                  {{ editingId ? ($t('categoryManagement.tabs.edit') || 'Edit Category') : ($t('categoryManagement.tabs.create') || 'Create Category') }}
                </h3>
              </div>
              <button @click="isFormOpen = false" class="text-slate-400 hover:text-white p-1 rounded-lg">✕</button>
            </div>

            <!-- Form Content -->
            <form @submit.prevent="submitCategoryForm" class="flex-1 overflow-y-auto p-6 space-y-4 font-mono text-2xs custom-scroller">
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                <!-- Category Code -->
                <div>
                  <label class="block text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.code') || 'Category Code' }}
                  </label>
                  <input
                    v-model="form.code"
                    type="text"
                    required
                    :disabled="!!editingId"
                    placeholder="e.g. water_quality"
                    class="w-full px-3 py-2 bg-[#080E18] border border-[#20354B] rounded-xl text-white focus:outline-none focus:border-cyan-500 disabled:opacity-50"
                  />
                  <span class="text-4xs text-slate-400 mt-1 block">Stable alphanumeric code (letters, numbers, _ -)</span>
                </div>

                <!-- Category Display Name -->
                <div>
                  <label class="block text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.name') || 'Display Name' }}
                  </label>
                  <input
                    v-model="form.name"
                    type="text"
                    required
                    placeholder="e.g. Water Quality"
                    class="w-full px-3 py-2 bg-[#080E18] border border-[#20354B] rounded-xl text-white focus:outline-none focus:border-cyan-500"
                  />
                  <span class="text-4xs text-slate-400 mt-1 block">Display label shown in dashboard</span>
                </div>

                <!-- Monitoring Domain -->
                <div>
                  <label class="block text-slate-300 font-semibold mb-1">
                    Monitoring Domain
                  </label>
                  <select
                    v-model="form.domain"
                    class="w-full px-3 py-2 bg-[#080E18] border border-[#20354B] rounded-xl text-white focus:outline-none focus:border-cyan-500 uppercase font-mono"
                  >
                    <option value="AQMS">AQMS (Air Quality)</option>
                    <option value="WQMS">WQMS (Water Quality)</option>
                    <option value="CUSTOM">CUSTOM (Configurable)</option>
                    <option value="ALL">ALL (Universal)</option>
                  </select>
                  <span class="text-4xs text-slate-400 mt-1 block">Target monitoring domain profile</span>
                </div>
              </div>

              <!-- Description -->
              <div>
                <label class="block text-slate-300 font-semibold mb-1">
                  {{ $t('categoryManagement.fields.description') || 'Description' }}
                </label>
                <textarea
                  v-model="form.description"
                  rows="2"
                  placeholder="Operational purpose of this monitoring domain..."
                  class="w-full px-3 py-2 bg-[#080E18] border border-[#20354B] rounded-xl text-white focus:outline-none focus:border-cyan-500"
                ></textarea>
              </div>

              <!-- Vector Icon & Color Picker Row -->
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <!-- Icon Picker -->
                <div>
                  <label class="block text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.icon') || 'Category Vector Icon' }}
                  </label>
                  <div class="flex items-center space-x-2">
                    <select
                      v-model="form.icon_key"
                      class="flex-1 px-3 py-2 bg-[#080E18] border border-[#20354B] rounded-xl text-white focus:outline-none focus:border-cyan-500 capitalize"
                    >
                      <option v-for="ic in iconCatalogList" :key="ic.key" :value="ic.key">
                        {{ ic.label }}
                      </option>
                    </select>
                    <div class="w-9 h-9 rounded-xl flex items-center justify-center border bg-[#101B2B] border-[#20354B] flex-shrink-0">
                      <ParameterIcon :icon="form.icon_key" size-class="w-4 h-4 text-cyan-400" />
                    </div>
                  </div>
                </div>

                <!-- Theme Color Picker -->
                <div>
                  <label class="block text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.color') || 'Theme Accent Color' }}
                  </label>
                  <div class="flex items-center space-x-2">
                    <select
                      v-model="form.color_key"
                      class="flex-1 px-3 py-2 bg-[#080E18] border border-[#20354B] rounded-xl text-white focus:outline-none focus:border-cyan-500 capitalize"
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
                  <label class="block text-slate-300 font-semibold mb-1">
                    {{ $t('categoryManagement.fields.span') || 'Default Column Span' }}
                  </label>
                  <select
                    v-model="form.default_column_span"
                    class="w-full px-3 py-2 bg-[#080E18] border border-[#20354B] rounded-xl text-white focus:outline-none focus:border-cyan-500"
                  >
                    <option :value="1">1 Column (Standard 3-4 Grid)</option>
                    <option :value="2">2 Columns (Double Width Span)</option>
                    <option :value="3">3 Columns (Wide SCADA Panel)</option>
                    <option :value="4">4 Columns (Full Desktop Row)</option>
                  </select>
                </div>

                <div class="flex items-center space-x-3 pt-6">
                  <label class="flex items-center space-x-2 cursor-pointer">
                    <input
                      type="checkbox"
                      v-model="form.enabled"
                      class="w-4 h-4 rounded text-cyan-500 focus:ring-0 bg-[#080E18] border-[#20354B]"
                    />
                    <span class="text-slate-300 font-semibold">{{ $t('categoryManagement.fields.enabled') || 'Category Enabled' }}</span>
                  </label>
                </div>
              </div>

              <!-- LIVE PREVIEW SECTION -->
              <div class="p-3.5 bg-[#080E18] border border-[#20354B] rounded-xl space-y-2 mt-2">
                <span class="text-4xs text-slate-400 font-bold uppercase tracking-wider block">
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
                    <span class="text-4xs px-1.5 py-0.2 rounded-full bg-[#18283D] text-slate-300 border border-slate-700">
                      4 parameters
                    </span>
                  </div>
                  <span class="text-3xs text-slate-400">{{ form.default_column_span }} Col Span</span>
                </div>
              </div>

              <!-- Form Buttons -->
              <div class="flex justify-end space-x-2 pt-2 border-t border-[#20354B]">
                <button
                  type="button"
                  @click="isFormOpen = false"
                  class="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition text-2xs"
                >
                  {{ $t('categoryManagement.actions.cancel') || 'Cancel' }}
                </button>
                <button
                  type="submit"
                  :disabled="isSaving"
                  class="px-5 py-2 rounded-xl bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold transition text-2xs shadow-lg shadow-cyan-900/30"
                >
                  {{ isSaving ? $t('categoryManagement.actions.saving') : $t('categoryManagement.actions.save') || 'Save Category' }}
                </button>
              </div>
            </form>
          </div>
        </div>

        <!-- ==================================================== -->
        <!-- SAFE DELETE CONFIRMATION MODAL                        -->
        <!-- ==================================================== -->
        <div v-if="deleteTarget" class="fixed inset-0 z-70 flex items-center justify-center p-4 bg-black/85 backdrop-blur-md animate-fade-in font-sans">
          <div class="bg-[#0B1422] border border-rose-500/40 rounded-2xl w-full max-w-lg p-6 shadow-2xl space-y-4 font-mono">
            <div class="flex items-center space-x-2 text-rose-400">
              <svg class="w-6 h-6 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
              <h3 class="text-sm font-bold text-white uppercase tracking-wider">
                {{ $t('categoryManagement.deleteModal.title') || 'Delete Category' }}
              </h3>
            </div>

            <p class="text-2xs text-slate-300 leading-relaxed">
              {{ $t('categoryManagement.deleteModal.warning', { name: deleteTarget.name }) || `Are you sure you want to delete category '${deleteTarget.name}'?` }}
            </p>

            <div v-if="deleteTarget.parameter_count > 0" class="p-3 bg-[#080E18] border border-[#20354B] rounded-xl space-y-2 text-2xs">
              <span class="text-cyan-400 font-semibold block">
                {{ $t('categoryManagement.deleteModal.reassignPrompt', { count: deleteTarget.parameter_count }) || `Action for ${deleteTarget.parameter_count} assigned parameter(s):` }}
              </span>
              <label class="flex items-start space-x-2 cursor-pointer">
                <input type="radio" v-model="deleteReassignMode" value="unassign" class="mt-0.5 text-cyan-500" />
                <span class="text-slate-300 text-3xs">{{ $t('categoryManagement.deleteModal.optionUnassign') || 'Keep as Uncategorized (Safely unassigns without deleting parameters)' }}</span>
              </label>
              <label class="flex items-start space-x-2 cursor-pointer">
                <input type="radio" v-model="deleteReassignMode" value="reassign" class="mt-0.5 text-cyan-500" />
                <span class="text-slate-300 text-3xs">{{ $t('categoryManagement.deleteModal.optionReassign') || 'Reassign to another category:' }}</span>
              </label>
              <select
                v-if="deleteReassignMode === 'reassign'"
                v-model="deleteTargetReassignId"
                class="w-full mt-1 px-2.5 py-1.5 bg-[#101B2B] border border-[#20354B] rounded-lg text-white text-3xs focus:outline-none"
              >
                <option v-for="c in categories.filter(c => c.id !== deleteTarget.id)" :key="c.id" :value="c.id">
                  {{ c.name }}
                </option>
              </select>
            </div>

            <div class="p-2.5 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-3xs text-emerald-300">
              🛡️ {{ $t('categoryManagement.deleteModal.safetyNotice') || 'Data Safety: Deleting a category never deletes parameters or telemetry history.' }}
            </div>

            <div class="flex justify-end space-x-2 pt-2 border-t border-[#20354B]">
              <button
                type="button"
                @click="deleteTarget = null"
                class="px-3.5 py-1.5 rounded-lg bg-slate-800 text-slate-300 text-2xs hover:bg-slate-700"
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
  },
  data() {
    return {
      activeTab: 'list',
      categories: [],
      hasReorderChanges: false,
      isSaving: false,
      noticeMessage: '',
      errorMessage: '',
      assignSearch: '',

      // Form state
      isFormOpen: false,
      editingId: null,
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
      const user = this.$store.state.user;
      if (!user) return true;
      const role = user.role || user.role_name || '';
      return role === 'Administrator' || role === 'admin' || role === 'Engineer';
    },
    iconCatalogList() {
      return PARAMETER_ICON_CATALOG;
    },
    colorOptions() {
      return COLOR_LIST;
    },
    allParametersList() {
      const list = [];
      for (const dev of this.devicesList || []) {
        for (const p of dev.parameters || []) {
          list.push({
            ...p,
            device_id: dev.id,
            device_code: dev.device_code || dev.code,
            device_name: dev.device_name || dev.name,
            category: this.categories.find(c => c.id === p.category_id) || null,
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
        this.fetchCategories();
      }
    },
    initialCategories: {
      immediate: true,
      handler(val) {
        if (val && val.length > 0) {
          this.categories = JSON.parse(JSON.stringify(val));
        }
      },
    },
  },
  methods: {
    getCategoryTheme(colorKey) {
      return getCategoryColor(colorKey);
    },
    closeModal() {
      this.isFormOpen = false;
      this.deleteTarget = null;
      this.$emit('close');
    },
    async fetchCategories() {
      try {
        const res = await apiClient.get('/categories');
        if (res.data && res.data.success) {
          this.categories = res.data.data || [];
          this.hasReorderChanges = false;
          this.$emit('categories-updated', this.categories);
        }
      } catch (err) {
        console.warn('Failed to load categories:', err);
      }
    },
    moveCategoryUp(idx) {
      if (idx <= 0) return;
      const item = this.categories.splice(idx, 1)[0];
      this.categories.splice(idx - 1, 0, item);
      this.hasReorderChanges = true;
    },
    moveCategoryDown(idx) {
      if (idx >= this.categories.length - 1) return;
      const item = this.categories.splice(idx, 1)[0];
      this.categories.splice(idx + 1, 0, item);
      this.hasReorderChanges = true;
    },
    async onSpanChanged(cat) {
      try {
        await apiClient.put(`/categories/${cat.id}`, {
          default_column_span: cat.default_column_span,
        });
        this.showNotice(`Updated span for '${cat.name}'`);
        this.$emit('categories-updated', this.categories);
      } catch (err) {
        this.showError('Failed to update span: ' + (err.response?.data?.error || err.message));
      }
    },
    async saveReorderedCategories() {
      this.isSaving = true;
      try {
        const ids = this.categories.map(c => c.id);
        await apiClient.put('/categories/reorder', { category_ids: ids });
        this.hasReorderChanges = false;
        this.showNotice(this.$t('categoryManagement.notifications.reordered') || 'Category order saved successfully');
        this.$emit('categories-updated', this.categories);
      } catch (err) {
        this.showError('Failed to save order: ' + (err.response?.data?.error || err.message));
      } finally {
        this.isSaving = false;
      }
    },
    async toggleCategoryEnabled(cat) {
      const targetState = !cat.enabled;
      try {
        await apiClient.put(`/categories/${cat.id}/toggle`, { enabled: targetState });
        cat.enabled = targetState;
        this.showNotice(`Category '${cat.name}' ${targetState ? 'enabled' : 'disabled'}`);
        this.$emit('categories-updated', this.categories);
      } catch (err) {
        this.showError('Failed to toggle category: ' + (err.response?.data?.error || err.message));
      }
    },
    getDomainBadgeClass(domain) {
      const d = (domain || 'AQMS').toUpperCase();
      if (d === 'AQMS') return 'bg-cyan-500/20 text-cyan-400 border border-cyan-500/30';
      if (d === 'WQMS') return 'bg-blue-500/20 text-blue-400 border border-blue-500/30';
      if (d === 'CUSTOM') return 'bg-purple-500/20 text-purple-400 border border-purple-500/30';
      return 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30';
    },
    openCreateForm() {
      this.editingId = null;
      this.form = {
        code: '',
        name: '',
        description: '',
        icon_key: 'activity',
        color_key: 'cyan',
        default_column_span: 1,
        enabled: true,
        domain: 'AQMS',
      };
      this.isFormOpen = true;
    },
    openEditForm(cat) {
      this.editingId = cat.id;
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
      this.isSaving = true;
      this.errorMessage = '';
      try {
        if (this.editingId) {
          const res = await apiClient.put(`/categories/${this.editingId}`, {
            name: this.form.name,
            description: this.form.description,
            icon_key: this.form.icon_key,
            color_key: this.form.color_key,
            default_column_span: this.form.default_column_span,
            enabled: this.form.enabled,
            domain: this.form.domain,
          });
          this.showNotice(this.$t('categoryManagement.notifications.updated', { name: this.form.name }) || `Updated category '${this.form.name}'`);
        } else {
          const res = await apiClient.post('/categories', this.form);
          this.showNotice(this.$t('categoryManagement.notifications.created', { name: this.form.name }) || `Created category '${this.form.name}'`);
        }
        this.isFormOpen = false;
        await this.fetchCategories();
      } catch (err) {
        this.showError(err.response?.data?.error || err.message);
      } finally {
        this.isSaving = false;
      }
    },
    openDeleteDialog(cat) {
      this.deleteTarget = cat;
      this.deleteReassignMode = 'unassign';
      const other = this.categories.find(c => c.id !== cat.id);
      this.deleteTargetReassignId = other ? other.id : null;
    },
    async confirmDeleteCategory() {
      if (!this.deleteTarget) return;
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
        this.$emit('refresh-devices');
      } catch (err) {
        this.showError(err.response?.data?.error || err.message);
      } finally {
        this.isSaving = false;
      }
    },
    async onParamCategoryChanged(param, newCatId) {
      try {
        const catId = (newCatId !== null && newCatId !== undefined && newCatId !== '' && Number(newCatId) > 0) ? Number(newCatId) : null;
        await apiClient.put(`/devices/${param.device_id}/parameters/${param.id}`, {
          category_id: catId,
          clear_category_id: !catId,
        });
        param.category_id = catId;
        param.category = this.categories.find(c => c.id === catId) || null;
        this.showNotice(`Assigned '${param.parameter_name || param.parameter_code}' to ${param.category ? param.category.name : 'Uncategorized'}`);
        await this.fetchCategories();
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
