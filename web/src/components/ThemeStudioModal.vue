<template>
  <transition name="fade">
    <div
      v-if="visible"
      id="theme-studio-modal"
      @click.self="$emit('close')"
      class="fixed inset-0 z-[110] flex items-center justify-center p-3 sm:p-4 bg-slate-950/55 backdrop-blur-xs font-sans"
    >
      <div
        class="w-full max-w-2xl rounded-2xl bg-white dark:bg-[#0E1626] border border-slate-200 dark:border-slate-700/90 shadow-2xl overflow-hidden flex flex-col max-h-[92vh]"
      >
        <!-- Header -->
        <div class="px-5 py-3.5 border-b border-slate-100 dark:border-slate-800 flex items-center justify-between gap-3">
          <div class="flex items-center gap-2.5 min-w-0">
            <span class="w-8 h-8 rounded-xl bg-sky-500/10 border border-sky-500/25 text-sky-600 dark:text-sky-400 flex items-center justify-center text-base flex-shrink-0">
              🎨
            </span>
            <div class="min-w-0">
              <h2 class="text-sm sm:text-base font-extrabold text-slate-900 dark:text-white tracking-tight truncate">
                Tampilan &amp; Tema Aplikasi
              </h2>
              <p class="text-3xs text-slate-500 dark:text-slate-400 truncate">
                Pilih gaya visual untuk Dashboard Operator, Export Data, Pengiriman Outbound &amp; Setup
              </p>
            </div>
          </div>
          <button
            type="button"
            id="btn-close-theme-studio"
            @click="$emit('close')"
            class="w-7 h-7 rounded-lg bg-slate-100 hover:bg-slate-200 dark:bg-slate-800 dark:hover:bg-slate-700 text-slate-500 dark:text-slate-300 flex items-center justify-center text-xs cursor-pointer flex-shrink-0"
            title="Tutup"
          >
            ✕
          </button>
        </div>

        <!-- Scrollable Body -->
        <div class="p-4 sm:p-5 space-y-5 overflow-y-auto">
          <!-- Section 1: Lighting Mode (Light / Dark / System) -->
          <div class="space-y-2">
            <div class="text-xs font-bold text-slate-800 dark:text-slate-100">Mode Pencahayaan</div>
            <div class="grid grid-cols-3 gap-2">
              <button
                v-for="mode in lightingModes"
                :key="mode.id"
                type="button"
                :id="'btn-theme-lightmode-' + mode.id"
                @click="selectLightingMode(mode.id)"
                class="px-3 py-2 rounded-xl border text-xs font-bold flex items-center justify-center gap-1.5 transition cursor-pointer"
                :class="currentThemeMode === mode.id
                  ? 'bg-sky-600 text-white btn-text border-sky-600 shadow-xs'
                  : 'bg-slate-50 hover:bg-slate-100 dark:bg-[#0B0F19] dark:hover:bg-slate-800/90 text-slate-700 dark:text-slate-200 border-slate-200 dark:border-slate-700/80'"
              >
                <span>{{ mode.icon }}</span>
                <span>{{ mode.label }}</span>
              </button>
            </div>
          </div>

          <!-- Section 2: 2x2 Visual Theme Presets Grid -->
          <div class="space-y-2">
            <div class="flex items-center justify-between gap-2">
              <div class="text-xs font-bold text-slate-800 dark:text-slate-100">Pilih Gaya &amp; Tema Desain</div>
              <span class="text-4xs font-mono font-semibold text-slate-500 dark:text-slate-400">
                Live Preview Aktif
              </span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
              <button
                v-for="preset in presets"
                :key="preset.id"
                type="button"
                :id="'btn-theme-preset-' + preset.id"
                @click="selectPreset(preset.id)"
                class="group text-left p-2.5 rounded-2xl border-2 transition cursor-pointer flex flex-col gap-2 bg-white dark:bg-[#0B0F19]"
                :class="currentPreset === preset.id
                  ? 'border-emerald-500 ring-2 ring-emerald-500/20 shadow-md'
                  : 'border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700'"
              >
                <!-- Miniature UI Preview Thumbnail -->
                <div
                  class="relative w-full h-24 rounded-xl overflow-hidden border border-slate-200/80 dark:border-slate-700/70 p-2 flex flex-col gap-1.5"
                  :style="{ background: effectiveDark ? preset.previewBgDark : preset.previewBgLight }"
                >
                  <!-- Active Checkmark -->
                  <span
                    v-if="currentPreset === preset.id"
                    class="absolute top-1.5 right-1.5 w-5 h-5 rounded-full bg-emerald-600 text-white badge-text text-3xs font-extrabold flex items-center justify-center shadow-sm"
                  >
                    ✓
                  </span>

                  <!-- Mini Header Bar -->
                  <div class="flex items-center gap-1.5">
                    <span
                      class="w-3 h-3 rounded-full flex-shrink-0"
                      :style="{ background: preset.previewAccent }"
                    ></span>
                    <span class="h-1.5 w-14 rounded-full bg-slate-400/40"></span>
                    <span class="ml-auto h-1.5 w-6 rounded-full bg-slate-400/30"></span>
                  </div>

                  <!-- Mini Bento Hero Row -->
                  <div class="grid grid-cols-3 gap-1.5 flex-1">
                    <div
                      class="col-span-2 p-1.5 flex flex-col justify-between shadow-2xs"
                      :style="{
                        background: preset.previewAccent,
                        borderRadius: preset.previewRadius,
                      }"
                    >
                      <span class="h-1 w-10 rounded-full bg-white/75"></span>
                      <div class="flex items-end gap-0.5 h-5">
                        <span class="w-1.5 h-2 rounded-xs bg-white/60"></span>
                        <span class="w-1.5 h-3.5 rounded-xs bg-white/85"></span>
                        <span class="w-1.5 h-2.5 rounded-xs bg-white/65"></span>
                        <span class="w-1.5 h-4 rounded-xs bg-white/95"></span>
                        <span class="w-1.5 h-3 rounded-xs bg-white/75"></span>
                      </div>
                    </div>
                    <div
                      class="p-1.5 flex flex-col justify-between border border-slate-300/60 dark:border-slate-600/50"
                      :style="{
                        background: effectiveDark ? 'rgba(15, 23, 42, 0.72)' : 'rgba(255, 255, 255, 0.85)',
                        borderRadius: preset.previewRadius,
                      }"
                    >
                      <span class="h-1 w-6 rounded-full bg-slate-400/50"></span>
                      <span
                        class="h-2.5 w-8 rounded-full"
                        :style="{ background: preset.swatches[1] }"
                      ></span>
                    </div>
                  </div>
                </div>

                <!-- Preset Metadata & Swatches -->
                <div class="px-0.5 space-y-1">
                  <div class="flex items-center justify-between gap-2">
                    <div class="flex items-center gap-1.5 min-w-0">
                      <span class="text-xs font-extrabold text-slate-900 dark:text-white truncate">
                        {{ preset.name }}
                      </span>
                      <span
                        class="px-1.5 py-0.2 rounded-full text-4xs font-mono font-bold border flex-shrink-0"
                        :class="currentPreset === preset.id
                          ? 'bg-emerald-500/15 text-emerald-700 dark:text-emerald-300 border-emerald-500/30'
                          : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 border-slate-200 dark:border-slate-700'"
                      >
                        {{ currentPreset === preset.id ? 'AKTIF' : preset.badge }}
                      </span>
                    </div>
                    <!-- Color Palette Dots -->
                    <div class="flex items-center gap-1 flex-shrink-0">
                      <span
                        v-for="(hex, idx) in preset.swatches"
                        :key="preset.id + '-sw-' + idx"
                        class="w-3 h-3 rounded-full border border-slate-300/80 dark:border-slate-600/80"
                        :style="{ background: hex }"
                      ></span>
                    </div>
                  </div>
                  <p class="text-3xs text-slate-500 dark:text-slate-400 leading-snug">
                    {{ preset.description }}
                  </p>
                </div>
              </button>
            </div>
          </div>

          <!-- Section 3: Density & Scope Notice -->
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-1">
            <div class="space-y-1">
              <div class="text-xs font-bold text-slate-800 dark:text-slate-100">Kepadatan Tampilan</div>
              <div class="inline-flex items-center p-0.5 rounded-xl bg-slate-100 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800">
                <button
                  type="button"
                  id="btn-density-comfortable"
                  @click="selectDensity('comfortable')"
                  class="px-3 py-1 rounded-lg text-2xs font-bold transition cursor-pointer"
                  :class="currentDensity === 'comfortable'
                    ? 'bg-sky-600 text-white btn-text shadow-2xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                >
                  Nyaman
                </button>
                <button
                  type="button"
                  id="btn-density-compact"
                  @click="selectDensity('compact')"
                  class="px-3 py-1 rounded-lg text-2xs font-bold transition cursor-pointer"
                  :class="currentDensity === 'compact'
                    ? 'bg-sky-600 text-white btn-text shadow-2xs'
                    : 'text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white'"
                >
                  Padat / Compact
                </button>
              </div>
            </div>

            <div class="flex-1 sm:max-w-xs px-3 py-2 rounded-xl bg-sky-500/8 dark:bg-sky-500/10 border border-sky-500/20 text-3xs text-slate-600 dark:text-slate-300 leading-relaxed">
              Otomatis tersimpan di browser dan berlaku langsung untuk <strong>Dashboard Operator</strong>, <strong>Export Data</strong>, <strong>Pengiriman Outbound</strong> &amp; <strong>Setup</strong>.
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-5 py-3 border-t border-slate-100 dark:border-slate-800 bg-slate-50/80 dark:bg-[#0B0F19]/80 flex items-center justify-between gap-2">
          <button
            type="button"
            id="btn-reset-theme-studio"
            @click="resetToDefault"
            class="px-3 py-1.5 rounded-xl text-xs font-semibold text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-white hover:bg-slate-200/60 dark:hover:bg-slate-800 transition cursor-pointer"
          >
            Reset Default
          </button>
          <button
            type="button"
            @click="$emit('close')"
            class="px-4 py-1.5 rounded-xl bg-sky-600 hover:bg-sky-500 text-white btn-text text-xs font-extrabold shadow-xs transition cursor-pointer"
          >
            Selesai
          </button>
        </div>
      </div>
    </div>
  </transition>
</template>

<script>
import { THEME_PRESETS } from '../utils/theme';

export default {
  name: 'ThemeStudioModal',
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      presets: THEME_PRESETS,
      lightingModes: [
        { id: 'light', label: 'Terang (Light)', icon: '☀' },
        { id: 'dark', label: 'Gelap (Dark)', icon: '🌙' },
        { id: 'system', label: 'Ikuti Sistem', icon: '💻' },
      ],
    };
  },
  computed: {
    currentThemeMode() {
      return (this.$store && this.$store.state.uiPreferences && this.$store.state.uiPreferences.theme) || 'system';
    },
    effectiveDark() {
      return ((this.$store && this.$store.state.uiPreferences && this.$store.state.uiPreferences.effectiveTheme) || 'light') === 'dark';
    },
    currentPreset() {
      return (this.$store && this.$store.state.uiPreferences && this.$store.state.uiPreferences.themePreset) || 'oceanic';
    },
    currentDensity() {
      return (this.$store && this.$store.state.uiPreferences && this.$store.state.uiPreferences.density) || 'comfortable';
    },
  },
  methods: {
    selectLightingMode(mode) {
      this.$store.dispatch('setTheme', mode);
    },
    selectPreset(presetId) {
      this.$store.dispatch('setThemePreset', presetId);
    },
    selectDensity(density) {
      this.$store.dispatch('setDensity', density);
    },
    resetToDefault() {
      this.$store.dispatch('setThemePreset', 'oceanic');
      this.$store.dispatch('setDensity', 'comfortable');
      this.$store.dispatch('setTheme', 'system');
    },
  },
};
</script>
