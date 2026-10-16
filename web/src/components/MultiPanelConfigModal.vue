<template>
  <div
    v-if="isOpen"
    id="multi-panel-config-modal"
    class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans select-none"
  >
    <div
      class="bg-white dark:bg-[#0B1422] border border-slate-300 dark:border-[#20354B] rounded-2xl w-full max-w-4xl max-h-[92vh] flex flex-col shadow-2xl overflow-hidden transition-colors"
    >
      <!-- Modal Header -->
      <div class="px-5 py-3.5 border-b border-slate-200 dark:border-[#20354B] flex items-center justify-between bg-slate-50 dark:bg-[#080E18]">
        <div class="flex items-center space-x-2.5">
          <div class="p-2 rounded-xl bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border border-cyan-500/30 flex-shrink-0">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M4 5a1 1 0 011-1h14a1 1 0 011 1v14a1 1 0 01-1 1H5a1 1 0 01-1-1V5z" />
              <path stroke-linecap="round" stroke-linejoin="round" d="M12 4v16m-8-8h16" />
            </svg>
          </div>
          <div>
            <h2 class="text-sm font-bold text-slate-900 dark:text-white tracking-wider font-mono uppercase leading-tight">
              {{ $t('operator.multiPanel') || 'MULTI-PANEL MONITORING CONFIGURATION' }}
            </h2>
            <p class="text-[11px] text-slate-500 dark:text-slate-400 font-mono leading-tight mt-0.5">
              {{ $t('operator.panelsConfig') || 'Configure horizontal monitoring panels, category/parameter assignments, and ordering.' }}
            </p>
          </div>
        </div>

        <button
          id="btn-close-panel-modal"
          @click="closeModal"
          class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1.5 rounded-lg hover:bg-slate-200 dark:hover:bg-[#1E2E42] transition"
          title="Close modal"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Quick Panel Count Segmented Control & Actions -->
      <div class="px-5 py-2.5 bg-slate-100/80 dark:bg-[#0E192A] border-b border-slate-200 dark:border-[#20354B] flex flex-wrap items-center justify-between gap-3 text-2xs font-mono">
        <!-- Panel Count Preset Buttons (1, 2, 3, 4 Panels) -->
        <div class="flex items-center space-x-1.5">
          <span class="text-slate-500 dark:text-slate-400 text-3xs font-semibold uppercase tracking-wider">
            {{ $t('operator.panelCount') || 'Panel Count' }}:
          </span>
          <div class="flex items-center bg-white dark:bg-[#080E18] p-0.5 rounded-xl border border-slate-200 dark:border-[#20354B] shadow-xs">
            <button
              v-for="cnt in [1, 2, 3, 4]"
              :key="'btn-cnt-' + cnt"
              :id="'btn-select-panel-count-' + cnt"
              @click="changePanelCount(cnt)"
              class="px-3 py-1 rounded-lg transition font-semibold flex items-center space-x-1 text-3xs"
              :class="localPanels.length === cnt
                ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 border border-cyan-500/40 shadow-xs'
                : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
            >
              <span>{{ cnt }}P</span>
              <span class="hidden sm:inline">
                ({{ cnt === 1 ? ($t('operator.preset1P') || 'Single') : cnt === 2 ? ($t('operator.preset2P') || 'Dual') : cnt === 3 ? ($t('operator.preset3P') || 'Triple') : ($t('operator.preset4P') || 'Quad') }})
              </span>
            </button>
          </div>
        </div>

        <!-- Add Panel & Reset to Defaults -->
        <div class="flex items-center space-x-2">
          <button
            v-if="localPanels.length < 6"
            id="btn-add-panel"
            @click="addNewPanel"
            class="px-2.5 py-1 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold transition flex items-center space-x-1 text-3xs shadow-xs"
          >
            <span>+</span>
            <span>{{ $t('operator.addPanel') || 'Add Panel' }}</span>
          </button>
          <button
            id="btn-reset-panels"
            @click="resetToDefaults"
            class="px-2.5 py-1 rounded-lg bg-slate-200 hover:bg-slate-300 dark:bg-[#18283D] dark:hover:bg-[#20354B] text-slate-700 dark:text-slate-300 transition text-3xs border border-slate-300 dark:border-[#20354B]"
            :title="$t('operator.resetPanels') || 'Reset to Smart Defaults'"
          >
            <span>↺</span>
            <span>{{ $t('operator.resetPanels') || 'Reset Defaults' }}</span>
          </button>
        </div>
      </div>

      <!-- Main Body: Configurable Panels Grid / List -->
      <div class="flex-1 overflow-y-auto p-4 sm:p-5 space-y-4 custom-scroller">
        <!-- Unassigned Parameters Guard Notice -->
        <div
          v-if="unassignedParamsList.length > 0"
          class="p-2.5 bg-amber-500/10 border border-amber-500/30 rounded-xl text-xs text-amber-700 dark:text-amber-300 flex flex-wrap items-center justify-between gap-2 animate-fade-in font-mono"
        >
          <div class="flex items-center space-x-2 min-w-0">
            <span class="text-amber-500 font-bold text-sm">⚠️</span>
            <span class="text-3xs">
              <strong>{{ unassignedParamsList.length }}</strong> {{ $t('operator.unassignedNotice', { count: unassignedParamsList.length }) || 'parameter(s) not assigned to any panel:' }}
              <span class="font-normal opacity-90 truncate inline-block max-w-md align-bottom">
                {{ unassignedParamsList.map(p => p.param_name || p.param_code).slice(0, 4).join(', ') }}{{ unassignedParamsList.length > 4 ? '...' : '' }}
              </span>
            </span>
          </div>
          <button
            @click="assignUnassignedToLastPanel"
            class="px-2 py-0.5 rounded bg-amber-500/20 hover:bg-amber-500/30 text-amber-800 dark:text-amber-200 font-bold text-4xs uppercase transition border border-amber-500/40"
          >
            {{ $t('operator.autoAssignUnassigned') || 'Auto-Assign to Last Panel' }}
          </button>
        </div>

        <!-- Panels List -->
        <div class="grid grid-cols-1 gap-3.5">
          <div
            v-for="(panel, pIdx) in localPanels"
            :key="panel.id"
            :id="'panel-card-config-' + (pIdx + 1)"
            class="rounded-xl border p-3.5 bg-slate-50 dark:bg-[#0E192A] border-slate-200 dark:border-[#20354B] transition-all shadow-xs flex flex-col space-y-3"
            :class="{ 'ring-1 ring-cyan-500/50': activeEditingPanelIdx === pIdx }"
          >
            <!-- Panel Card Header: Position, Title, Accent, and Order Actions -->
            <div class="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200/80 dark:border-[#1E3048] pb-2.5">
              <div class="flex items-center space-x-2 flex-1 min-w-0">
                <!-- Panel Badge -->
                <span class="px-2 py-0.5 rounded-md text-[9px] font-mono font-bold bg-cyan-500/15 text-cyan-700 dark:text-cyan-300 border border-cyan-500/30 uppercase tracking-wider flex-shrink-0">
                  Panel {{ pIdx + 1 }}
                </span>

                <!-- Title Input -->
                <div class="flex-1 min-w-[180px]">
                  <input
                    v-model="panel.title"
                    type="text"
                    class="w-full px-2.5 py-1 rounded-lg bg-white dark:bg-[#080E18] border border-slate-300 dark:border-[#20354B] text-slate-900 dark:text-white text-xs font-mono font-bold focus:outline-none focus:border-cyan-500"
                    :placeholder="'Panel ' + (pIdx + 1) + ' Title'"
                  />
                </div>

                <!-- Color Accent Selector -->
                <div class="flex items-center space-x-1 flex-shrink-0" title="Panel Theme Color">
                  <button
                    v-for="color in colorPalette"
                    :key="color.key"
                    @click="panel.colorKey = color.key"
                    class="w-4 h-4 rounded-full transition-transform hover:scale-125 border"
                    :class="[
                      color.dot,
                      panel.colorKey === color.key ? 'ring-2 ring-white dark:ring-slate-300 scale-110' : 'opacity-60 border-transparent'
                    ]"
                  ></button>
                </div>
              </div>

              <!-- Reorder & Delete Actions -->
              <div class="flex items-center space-x-1.5 flex-shrink-0">
                <button
                  v-if="pIdx > 0"
                  @click="movePanel(pIdx, -1)"
                  class="p-1 rounded bg-slate-200 dark:bg-[#18283D] text-slate-600 dark:text-slate-300 hover:text-cyan-500 text-3xs font-mono transition"
                  :title="$t('operator.moveLeft') || 'Move Left'"
                >
                  ◀
                </button>
                <button
                  v-if="pIdx < localPanels.length - 1"
                  @click="movePanel(pIdx, 1)"
                  class="p-1 rounded bg-slate-200 dark:bg-[#18283D] text-slate-600 dark:text-slate-300 hover:text-cyan-500 text-3xs font-mono transition"
                  :title="$t('operator.moveRight') || 'Move Right'"
                >
                  ▶
                </button>
                <button
                  v-if="localPanels.length > 1"
                  @click="removePanel(pIdx)"
                  class="p-1 rounded bg-rose-500/10 hover:bg-rose-500/20 text-rose-600 dark:text-rose-400 text-3xs font-mono transition border border-rose-500/20"
                  :title="$t('operator.removePanel') || 'Remove Panel'"
                >
                  ✕
                </button>
              </div>
            </div>

            <!-- Assignment Mode Switcher (By Category vs By Parameters) -->
            <div class="flex flex-wrap items-center justify-between gap-2 text-3xs font-mono">
              <div class="flex items-center space-x-1 bg-white dark:bg-[#080E18] p-0.5 rounded-lg border border-slate-200 dark:border-[#20354B]">
                <button
                  @click="panel.assignmentMode = 'category'"
                  class="px-2.5 py-1 rounded transition font-semibold"
                  :class="panel.assignmentMode === 'category' ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 font-bold' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'"
                >
                  {{ $t('operator.byCategory') || 'By Category' }}
                </button>
                <button
                  @click="panel.assignmentMode = 'custom'"
                  class="px-2.5 py-1 rounded transition font-semibold"
                  :class="panel.assignmentMode === 'custom' ? 'bg-cyan-500/20 text-cyan-700 dark:text-cyan-300 font-bold' : 'text-slate-500 hover:text-slate-800 dark:hover:text-slate-200'"
                >
                  {{ $t('operator.byParameter') || 'Individual Parameters' }}
                </button>
              </div>

              <!-- Parameter Count badge for this panel -->
              <span class="text-slate-500 dark:text-slate-400 font-mono text-3xs">
                {{ getParametersAssignedToPanel(panel).length }} / {{ allParameters.length }} parameters assigned
              </span>
            </div>

            <!-- MODE 1: Category Checkboxes -->
            <div v-if="panel.assignmentMode === 'category'" class="space-y-1.5">
              <div class="text-4xs font-mono uppercase text-slate-400 tracking-wider">
                {{ $t('operator.assignedCategories') || 'Assigned Categories' }}:
              </div>
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-2">
                <label
                  v-for="cat in availableCategoryOptions"
                  :key="cat.id"
                  class="flex items-center space-x-2 p-2 rounded-lg bg-white dark:bg-[#080E18] border border-slate-200 dark:border-[#1E3048] hover:border-cyan-500/40 cursor-pointer text-3xs font-mono transition"
                  :class="{ 'border-cyan-500/50 bg-cyan-500/5 dark:bg-cyan-950/20': isCategorySelectedInPanel(panel, cat.id) }"
                >
                  <input
                    type="checkbox"
                    :checked="isCategorySelectedInPanel(panel, cat.id)"
                    @change="toggleCategoryInPanel(panel, cat.id)"
                    class="rounded text-cyan-600 focus:ring-0 focus:outline-none w-3.5 h-3.5"
                  />
                  <span class="truncate font-semibold text-slate-800 dark:text-slate-200" :title="cat.name">
                    {{ cat.name }}
                  </span>
                  <span class="text-[8px] font-bold text-slate-400 ml-auto">
                    ({{ getCategoryParamCount(cat.id) }})
                  </span>
                </label>
              </div>
            </div>

            <!-- MODE 2: Individual Parameter Checklist with Search -->
            <div v-else class="space-y-2">
              <div class="flex items-center justify-between gap-2">
                <input
                  v-model="paramSearchQuery"
                  type="text"
                  placeholder="Filter parameters..."
                  class="px-2 py-1 rounded bg-white dark:bg-[#080E18] border border-slate-200 dark:border-[#1E3048] text-slate-900 dark:text-white text-3xs font-mono w-48 focus:outline-none focus:border-cyan-500"
                />
                <div class="flex items-center space-x-1.5 text-4xs font-mono">
                  <button
                    @click="selectAllParamsForPanel(panel)"
                    class="text-cyan-600 dark:text-cyan-400 hover:underline"
                  >
                    Select All
                  </button>
                  <span>·</span>
                  <button
                    @click="clearAllParamsForPanel(panel)"
                    class="text-slate-400 hover:text-rose-500"
                  >
                    Clear All
                  </button>
                </div>
              </div>

              <!-- Parameter Grid Selection -->
              <div class="max-h-40 overflow-y-auto grid grid-cols-2 sm:grid-cols-3 gap-1.5 p-1 rounded-lg bg-white dark:bg-[#080E18] border border-slate-200 dark:border-[#1E3048] custom-scroller">
                <label
                  v-for="param in filteredParamList"
                  :key="param.param_id"
                  class="flex items-center space-x-1.5 p-1.5 rounded hover:bg-slate-50 dark:hover:bg-[#101C2E] cursor-pointer text-3xs font-mono truncate"
                  :class="{ 'text-cyan-600 dark:text-cyan-400 font-bold': isParamSelectedInPanel(panel, param.param_id) }"
                >
                  <input
                    type="checkbox"
                    :checked="isParamSelectedInPanel(panel, param.param_id)"
                    @change="toggleParamInPanel(panel, param.param_id)"
                    class="rounded text-cyan-600 focus:ring-0 focus:outline-none w-3 h-3 flex-shrink-0"
                  />
                  <span class="truncate flex-1" :title="param.param_name + ' (' + param.param_code + ')'">
                    {{ param.param_name || param.param_code }}
                  </span>
                  <span class="text-4xs text-slate-400 flex-shrink-0">
                    {{ param.unit }}
                  </span>
                </label>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-5 py-3 border-t border-slate-200 dark:border-[#20354B] bg-slate-50 dark:bg-[#080E18] flex items-center justify-between">
        <button
          @click="closeModal"
          class="px-4 py-1.5 rounded-lg text-xs font-mono font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white transition"
        >
          Cancel
        </button>

        <div class="flex items-center space-x-2">
          <button
            id="btn-save-panel-layout"
            @click="saveConfiguration"
            class="px-5 py-1.5 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold font-mono text-xs transition flex items-center space-x-1.5 shadow-md shadow-cyan-900/30"
          >
            <span>✓</span>
            <span>Save Layout</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MultiPanelConfigModal',
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
    panelCount: {
      type: Number,
      default: 2,
    },
    panels: {
      type: Array,
      default: () => [],
    },
    categories: {
      type: Array,
      default: () => [],
    },
    allParameters: {
      type: Array,
      default: () => [],
    },
    isDark: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      localPanels: [],
      activeEditingPanelIdx: 0,
      paramSearchQuery: '',
      colorPalette: [
        { key: 'cyan', dot: 'bg-cyan-500' },
        { key: 'blue', dot: 'bg-blue-500' },
        { key: 'amber', dot: 'bg-amber-500' },
        { key: 'purple', dot: 'bg-purple-500' },
        { key: 'emerald', dot: 'bg-emerald-500' },
        { key: 'rose', dot: 'bg-rose-500' },
        { key: 'slate', dot: 'bg-slate-400' },
      ],
    };
  },
  computed: {
    availableCategoryOptions() {
      const list = (this.categories || []).map(c => ({
        id: c.id,
        name: c.name,
        color_key: c.color_key,
      }));
      // Add Uncategorized category
      list.push({
        id: 'uncategorized',
        name: this.$t('operator.uncategorized') || 'Uncategorized',
        color_key: 'slate',
      });
      return list;
    },
    filteredParamList() {
      if (!this.paramSearchQuery) return this.allParameters;
      const q = this.paramSearchQuery.toLowerCase();
      return this.allParameters.filter(p =>
        (p.param_name && p.param_name.toLowerCase().includes(q)) ||
        (p.param_code && p.param_code.toLowerCase().includes(q)) ||
        (p.device_code && p.device_code.toLowerCase().includes(q))
      );
    },
    unassignedParamsList() {
      const assignedIds = new Set();
      this.localPanels.forEach(panel => {
        const params = this.getParametersAssignedToPanel(panel);
        params.forEach(p => assignedIds.add(p.param_id));
      });
      return this.allParameters.filter(p => !assignedIds.has(p.param_id));
    },
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.initializeLocalPanels();
      }
    },
    panels: {
      deep: true,
      handler() {
        if (this.isOpen) {
          this.initializeLocalPanels();
        }
      },
    },
  },
  mounted() {
    this.initializeLocalPanels();
  },
  methods: {
    initializeLocalPanels() {
      if (Array.isArray(this.panels) && this.panels.length > 0) {
        this.localPanels = JSON.parse(JSON.stringify(this.panels));
      } else {
        this.localPanels = this.buildDefaultPanels(this.panelCount || 2);
      }
    },
    buildDefaultPanels(count) {
      const categories = this.categories || [];
      const gasCat = categories.find(c => /gas|pollutant|udara/i.test(c.name)) || categories[0];
      const partCat = categories.find(c => /particul|debu|pm/i.test(c.name));
      const weatherCat = categories.find(c => /weather|cuaca|meteo/i.test(c.name)) || (categories.length > 1 ? categories[1] : null);

      if (count === 1) {
        return [
          {
            id: 'panel-1',
            title: this.$t('operator.allCategories') || 'All Monitoring Parameters',
            colorKey: 'cyan',
            assignmentMode: 'category',
            categoryIds: categories.map(c => c.id).concat(['uncategorized']),
            parameterIds: [],
            allowDuplicates: false,
          },
        ];
      }

      if (count === 2) {
        const p1Cats = gasCat ? [gasCat.id] : (categories[0] ? [categories[0].id] : [1]);
        const p2Cats = categories.filter(c => !p1Cats.includes(c.id)).map(c => c.id).concat(['uncategorized']);
        return [
          {
            id: 'panel-1',
            title: (gasCat && gasCat.name) || 'Gas & Pollutants',
            colorKey: 'cyan',
            assignmentMode: 'category',
            categoryIds: p1Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
          {
            id: 'panel-2',
            title: (weatherCat && weatherCat.name) || 'Weather & Meteorology',
            colorKey: 'blue',
            assignmentMode: 'category',
            categoryIds: p2Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
        ];
      }

      if (count === 3) {
        const p1Cats = gasCat ? [gasCat.id] : [1];
        const p2Cats = partCat ? [partCat.id] : (categories.length > 2 ? [categories[2].id] : [3]);
        const p3Cats = categories.filter(c => !p1Cats.includes(c.id) && !p2Cats.includes(c.id)).map(c => c.id).concat(['uncategorized']);
        return [
          {
            id: 'panel-1',
            title: (gasCat && gasCat.name) || 'Gas & Pollutants',
            colorKey: 'cyan',
            assignmentMode: 'category',
            categoryIds: p1Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
          {
            id: 'panel-2',
            title: (partCat && partCat.name) || 'Particulate Matter',
            colorKey: 'amber',
            assignmentMode: 'category',
            categoryIds: p2Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
          {
            id: 'panel-3',
            title: (weatherCat && weatherCat.name) || 'Weather & Meteorology',
            colorKey: 'blue',
            assignmentMode: 'category',
            categoryIds: p3Cats,
            parameterIds: [],
            allowDuplicates: false,
          },
        ];
      }

      // count >= 4
      const p1Cats = gasCat ? [gasCat.id] : [1];
      const p2Cats = partCat ? [partCat.id] : [3];
      const p3Cats = weatherCat ? [weatherCat.id] : [2];
      const p4Cats = categories.filter(c => !p1Cats.includes(c.id) && !p2Cats.includes(c.id) && !p3Cats.includes(c.id)).map(c => c.id).concat(['uncategorized']);
      return [
        {
          id: 'panel-1',
          title: (gasCat && gasCat.name) || 'Gas & Pollutants',
          colorKey: 'cyan',
          assignmentMode: 'category',
          categoryIds: p1Cats,
          parameterIds: [],
          allowDuplicates: false,
        },
        {
          id: 'panel-2',
          title: (partCat && partCat.name) || 'Particulate Matter',
          colorKey: 'amber',
          assignmentMode: 'category',
          categoryIds: p2Cats,
          parameterIds: [],
          allowDuplicates: false,
        },
        {
          id: 'panel-3',
          title: (weatherCat && weatherCat.name) || 'Weather & Meteorology',
          colorKey: 'blue',
          assignmentMode: 'category',
          categoryIds: p3Cats,
          parameterIds: [],
          allowDuplicates: false,
        },
        {
          id: 'panel-4',
          title: this.$t('operator.uncategorized') || 'System & Aux',
          colorKey: 'slate',
          assignmentMode: 'category',
          categoryIds: p4Cats,
          parameterIds: [],
          allowDuplicates: false,
        },
      ];
    },
    changePanelCount(count) {
      if (this.localPanels.length === count) return;
      if (this.localPanels.length < count) {
        // Add default panels
        const defaults = this.buildDefaultPanels(count);
        while (this.localPanels.length < count) {
          const nextIdx = this.localPanels.length;
          const def = defaults[nextIdx] || {
            id: 'panel-' + (nextIdx + 1),
            title: 'Panel ' + (nextIdx + 1),
            colorKey: 'cyan',
            assignmentMode: 'category',
            categoryIds: [],
            parameterIds: [],
            allowDuplicates: false,
          };
          this.localPanels.push(def);
        }
      } else {
        // Reduce panels count
        this.localPanels = this.localPanels.slice(0, count);
      }
    },
    addNewPanel() {
      const nextIdx = this.localPanels.length + 1;
      this.localPanels.push({
        id: 'panel-' + Date.now(),
        title: 'Panel ' + nextIdx,
        colorKey: 'cyan',
        assignmentMode: 'category',
        categoryIds: [],
        parameterIds: [],
        allowDuplicates: false,
      });
    },
    removePanel(idx) {
      if (this.localPanels.length <= 1) return;
      this.localPanels.splice(idx, 1);
    },
    movePanel(idx, delta) {
      const target = idx + delta;
      if (target < 0 || target >= this.localPanels.length) return;
      const item = this.localPanels.splice(idx, 1)[0];
      this.localPanels.splice(target, 0, item);
    },
    resetToDefaults() {
      this.localPanels = this.buildDefaultPanels(this.localPanels.length || 2);
    },
    isCategorySelectedInPanel(panel, catId) {
      return (panel.categoryIds || []).includes(catId);
    },
    toggleCategoryInPanel(panel, catId) {
      if (!panel.categoryIds) panel.categoryIds = [];
      const idx = panel.categoryIds.indexOf(catId);
      if (idx !== -1) {
        panel.categoryIds.splice(idx, 1);
      } else {
        panel.categoryIds.push(catId);
      }
    },
    isParamSelectedInPanel(panel, paramId) {
      return (panel.parameterIds || []).includes(paramId);
    },
    toggleParamInPanel(panel, paramId) {
      if (!panel.parameterIds) panel.parameterIds = [];
      const idx = panel.parameterIds.indexOf(paramId);
      if (idx !== -1) {
        panel.parameterIds.splice(idx, 1);
      } else {
        panel.parameterIds.push(paramId);
      }
    },
    selectAllParamsForPanel(panel) {
      panel.parameterIds = this.allParameters.map(p => p.param_id);
    },
    clearAllParamsForPanel(panel) {
      panel.parameterIds = [];
    },
    getCategoryParamCount(catId) {
      if (catId === 'uncategorized') {
        return this.allParameters.filter(p => !p.category_id).length;
      }
      return this.allParameters.filter(p => Number(p.category_id) === Number(catId)).length;
    },
    getParametersAssignedToPanel(panel) {
      if (panel.assignmentMode === 'custom') {
        const idSet = new Set((panel.parameterIds || []).map(String));
        return this.allParameters.filter(p => idSet.has(String(p.param_id)));
      }
      const catSet = new Set((panel.categoryIds || []).map(c => String(c).toLowerCase()));
      return this.allParameters.filter(p => {
        if (!p.category_id && catSet.has('uncategorized')) return true;
        return catSet.has(String(p.category_id).toLowerCase());
      });
    },
    assignUnassignedToLastPanel() {
      if (this.localPanels.length === 0) return;
      const last = this.localPanels[this.localPanels.length - 1];
      const unassigned = this.unassignedParamsList;
      if (last.assignmentMode === 'custom') {
        if (!last.parameterIds) last.parameterIds = [];
        unassigned.forEach(p => {
          if (!last.parameterIds.includes(p.param_id)) last.parameterIds.push(p.param_id);
        });
      } else {
        if (!last.categoryIds) last.categoryIds = [];
        unassigned.forEach(p => {
          const catId = p.category_id || 'uncategorized';
          if (!last.categoryIds.includes(catId)) last.categoryIds.push(catId);
        });
      }
    },
    saveConfiguration() {
      const config = {
        panelCount: this.localPanels.length,
        panels: JSON.parse(JSON.stringify(this.localPanels)),
      };
      this.$emit('save', config);
      this.closeModal();
    },
    closeModal() {
      this.$emit('close');
    },
  },
};
</script>
