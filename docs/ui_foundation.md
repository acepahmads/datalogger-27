# UI Foundation & Cross-Cutting Architecture
> **Datalogger Analysis Application — Industrial Edge Platform**  
> *Internationalization (i18n), Light / Dark / System Appearance, and Component Consistency*

---

## 1. Executive Summary

This specification establishes the cross-cutting UI foundation for the **Industrial Datalogger Analysis Application**. This work delivers:
1. **Full Internationalization (i18n):** Seamless English (`en`) and Indonesian (`id`) switching without page reloads, backed by 268 verified translation keys.
2. **Three Appearance Modes:** Dedicated **Light Mode (Siang)**, **Dark Mode (Malam)**, and **System Mode (Otomatis)** that follows the operating system's color scheme preference in real-time.
3. **Persistent User Preferences:** Zero-login persistence using `localStorage` keys (`datalogger.language` and `datalogger.theme`, with backwards-compatible legacy fallback).
4. **Zero Startup Theme/Language Flashing:** Pre-hydration script embedded directly in `index.html` head evaluates preferences before DOM rendering.
5. **Component Consistency:** Unified typography scale, standard card tokens, high-contrast semantic status badges, and chart readability across all views.
6. **Protection of Completed Milestones:** Preserves 100% of Phase 1, Phase 2, Phase 2.1, Phase 2.2, Phase 2.3, and Phase 3.1 functionalities, retaining canonical database status strings (`ACTIVE`, `DONE`, `ONLINE`, etc.).

---

## 2. UI & Frontend Architecture

The frontend is built on **Vue 2.7** (leveraging modern Composition & reactive Observable APIs) with **Vite 4**, **Vue Router 3**, and **Vuex 3**, styled with **TailwindCSS 3** and custom high-contrast industrial utility tokens.

```
                    ┌────────────────────────────┐
                    │        index.html          │
                    │ (Pre-render Theme/Lang JS) │
                    └─────────────┬──────────────┘
                                  │
                                  ▼
                    ┌────────────────────────────┐
                    │          main.js           │
                    │  (Vue.use(i18n), Watcher)  │
                    └─────────────┬──────────────┘
                                  │
          ┌───────────────────────┴───────────────────────┐
          ▼                                               ▼
┌───────────────────────────┐                   ┌───────────────────────────┐
│     i18n Reactive Bus     │                   │     Theme Manager Bus     │
│  src/i18n/index.js        │                   │  src/utils/theme.js       │
│  • en.js (268 keys)       │                   │  • light / dark / system  │
│  • id.js (268 keys)       │                   │  • matchMedia listener    │
│  • $t(key, params)        │                   │  • applyTheme()           │
└─────────┬─────────────────┘                   └───────────┬───────────────┘
          │                                                 │
          └───────────────────────┬─────────────────────────┘
                                  │
                                  ▼
                    ┌────────────────────────────┐
                    │      Vuex Store Index      │
                    │  state.uiPreferences       │
                    │  • language & theme        │
                    └─────────────┬──────────────┘
                                  │
                                  ▼
      ┌───────────────────────────┴───────────────────────────┐
      │                                                       │
┌─────┴────────────────┐                               ┌──────┴────────────────┐
│      Layout Shell    │                               │     Feature Views     │
│  • Header.vue        │                               │  • DashboardView      │
│  • Sidebar.vue       │                               │  • DeviceDetailView   │
│  • App.vue           │                               │  • TelemetryMonitor   │
└──────────────────────┘                               │  • DevelopmentView    │
                                                       └───────────────────────┘
```

---

## 3. Translation Architecture (i18n)

### 3.1 Directory Structure
```
web/src/i18n/
├── index.js          # Core reactive plugin ($t, setLocale, $locale)
└── locales/
    ├── en.js         # English translations (source of truth)
    └── id.js         # Indonesian translations (100% key parity)
```

### 3.2 Reactive Key Resolution & Fallback
The `$t(path, params)` helper resolves nested dictionary paths using dot notation (e.g., `$t('devices.deviceName')`).
- If a key is requested in Indonesian (`id`) but is missing, it automatically falls back to English (`en`).
- If missing in both, it safely returns the raw key path without throwing exceptions.
- Dynamic string interpolation is supported using `{variable}` syntax:
  ```javascript
  // en.js: deleteConfirm: 'Are you sure you want to delete device "{name}"?'
  // id.js: deleteConfirm: 'Apakah Anda yakin ingin menghapus perangkat "{name}"?'
  $t('devices.deleteConfirm', { name: device.name })
  ```

### 3.3 Translation Categories
Translations are organized systematically into 17 standard namespaces:
1. `navigation`: Sidebar menus, branding, and database footer labels.
2. `header`: Application title, host metrics, clock, and selector labels.
3. `dashboard`: Edge metrics, KPI cards, pipelines, and activity log labels.
4. `devices`: Hardware registries, connection parameters, and serial settings.
5. `parameters`: Modbus register configurations, multipliers, and units.
6. `telemetry`: Live ingestion cards, quality filters, and WebSocket badges.
7. `historical`: Date-time ranges, statistical aggregations (MIN/MAX/AVG), and CSV exports.
8. `rawData`: Hex packet sniffers, TX/RX direction, and byte length analyzers.
9. `development`: Roadmap phases, milestone cards, and acceptance criteria.
10. `alarms`: Event thresholds, acknowledgement controls, and severity tags.
11. `appearance`: Light, Dark, and System appearance descriptions.
12. `common`: Universal actions (`save`, `cancel`, `edit`, `delete`, `refresh`, `loading`).
13. `status`: Canonical status mapping (`online`, `offline`, `done`, `working`, `healthy`).
14. `validation`: Input form constraints and error messages.

---

## 4. Theme Architecture (Light / Dark / System)

### 4.1 Mode Definitions

| Appearance Mode | Trigger Condition | Visual Characteristics | Best Used For |
| :--- | :--- | :--- | :--- |
| **☀ Light (Siang)** | User manually chooses "Light" | Clean white (`#FFFFFF`) card surfaces, soft slate-50 (`#F8FAFC`) canvas, subtle borders (`#E2E8F0`), high-contrast dark text (`#0F172A`). | High-ambient outdoor daylight or bright control rooms. |
| **🌙 Dark (Malam)** | User manually chooses "Dark" | Deep navy background (`#0B0F19`), slate-900 surface cards (`#111827`), soft borders (`#1E293B`), light slate text (`#F8FAFC`). | 24/7 industrial NOCs, control centers, and night shifts. |
| **◐ System (Otomatis)** | User selects "System" (Default) | Dynamically synchronizes with the OS scheme using `window.matchMedia('(prefers-color-scheme: dark)')`. | Laptops and tablets with automatic day/night OS switching. |

### 4.2 System Theme Reactivity
When in `system` mode, `initSystemThemeWatcher()` binds to the browser's `change` event:
```javascript
const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
mediaQuery.addEventListener('change', (e) => {
  if (getStoredTheme() === 'system') {
    applyTheme('system');
  }
});
```
When the user's OS switches from day mode to dark mode at sunset, the Datalogger UI adapts immediately without needing a manual refresh.

---

## 5. User Preference Persistence

Preferences are stored directly in the browser's `localStorage` and synchronized with Vuex:
- **Language Key:** `datalogger.language` (with legacy fallback to `datalogger_lang`).
- **Theme Key:** `datalogger.theme` (with legacy fallback to `datalogger_theme`).

### First-Time Visitor Logic:
1. **Language:** Checks `navigator.language`. If it starts with `id` (e.g. `id-ID`), defaults to Indonesian (`id`). Otherwise defaults to English (`en`).
2. **Theme:** Defaults to `system` (or `dark` if `matchMedia` is unavailable).
3. **Subsequent Visits:** The user's explicit selection overrides system defaults and persists permanently across reboots and browser updates.

---

## 6. Global CSS & Semantic Tokens

All styling tokens are integrated into [`web/src/assets/main.css`](file:///d:/cbi-project-src/datalogger-27/web/src/assets/main.css):

### 6.1 Semantic Card & Input Classes
- `.saas-card`: White card in light mode, dark charcoal in dark mode.
- `.saas-input`: High-contrast form field with blue focus ring.
- `.chart-svg text`: Automatically switches text fill from `#64748B` (light mode) to `#94A3B8` (dark mode).
- `.chart-svg line`: Automatically switches grid line stroke from `#E2E8F0` (light mode) to `#1E293B` (dark mode).

### 6.2 High-Contrast Status Badges

| Status Token | Light Mode Presentation | Dark Mode Presentation |
| :--- | :--- | :--- |
| `.badge-success` (GOOD/ACTIVE) | Green text (`#047857`) on soft green (`#ECFDF5`) | Neon emerald (`#34D399`) on dark emerald tint (`#064E3B`/10%) |
| `.badge-warning` (UNCERTAIN) | Amber text (`#B45309`) on soft amber (`#FFFBEB`) | Neon amber (`#FBBF24`) on dark amber tint (`#78350F`/10%) |
| `.badge-danger` (BAD/ERROR) | Rose text (`#BE123C`) on soft rose (`#FFF1F2`) | Neon rose (`#FB7185`) on dark rose tint (`#881337`/10%) |
| `.badge-info` (ONLINE/TX) | Blue text (`#1D4ED8`) on soft blue (`#EFF6FF`) | Electric blue (`#60A5FA`) on dark blue tint (`#1E3A8A`/10%) |
| `.badge-neutral` (OFFLINE/PENDING) | Slate text (`#334155`) on soft slate (`#F1F5F9`) | Muted slate (`#94A3B8`) on dark slate (`#1E293B`) |

---

## 7. Developer Guide

### 7.1 Adding a New Translation Key
1. Open [`web/src/i18n/locales/en.js`](file:///d:/cbi-project-src/datalogger-27/web/src/i18n/locales/en.js):
   ```javascript
   devices: {
     myNewAction: 'Calibrate Sensor',
   }
   ```
2. Add the corresponding key in [`web/src/i18n/locales/id.js`](file:///d:/cbi-project-src/datalogger-27/web/src/i18n/locales/id.js):
   ```javascript
   devices: {
     myNewAction: 'Kalibrasi Sensor',
   }
   ```
3. Reference in Vue templates using `$t`:
   ```vue
   <button>{{ $t('devices.myNewAction') }}</button>
   ```

### 7.2 Building Theme-Aware Components
Follow these patterns when authoring new Vue components:
- **Card containers:** Use `<div class="saas-card p-4">` instead of hardcoded background utilities.
- **Headings:** Use `<h3 class="text-sm font-bold text-slate-900 dark:text-white">`.
- **Muted text:** Use `<span class="text-xs text-slate-500 dark:text-slate-400">`.
- **Form fields:** Use standard inputs; `main.css` will style them automatically based on the active theme.
