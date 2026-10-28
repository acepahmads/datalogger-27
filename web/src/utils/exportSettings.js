/**
 * Telemetry Export Governance & Setup Utility
 * 
 * Provides centralized configuration for maximum export records and auto-pagination batching.
 * Permission Rule: All roles can configure this setup EXCEPT 'OPERATOR'.
 * Operators are strictly read-only and bound to the setup configured by Administrator / Engineer.
 */

export const EXPORT_SETTINGS_STORAGE_KEY = 'datalogger.export_settings';

export const EXPORT_LIMIT_PRESETS = [
  { value: 1000, label: '1.000 data (Cepat / 1 batch)' },
  { value: 5000, label: '5.000 data (5 batch)' },
  { value: 10000, label: '10.000 data (Standar / Rekomendasi)' },
  { value: 25000, label: '25.000 data (25 batch)' },
  { value: 50000, label: '50.000 data (50 batch)' },
  { value: 100000, label: '100.000 data (100 batch)' },
  { value: 0, label: 'Semua Data (Tanpa Batas)' },
];

export const DEFAULT_EXPORT_SETTINGS = {
  maxRecords: 10000, // 10,000 records default
  batchSize: 1000,    // 1,000 records per pagination call (matches server pagination limit)
  updatedBy: 'SYSTEM',
  updatedAt: null,
};

/**
 * Check if the given role is authorized to modify export configuration.
 * All roles can setup EXCEPT 'OPERATOR'.
 */
export function canUserManageExportSettings(role) {
  if (!role) return false;
  const r = String(role).trim().toUpperCase();
  if (r === 'OPERATOR') return false;
  return ['SUPERADMIN', 'ADMINISTRATOR', 'ADMIN', 'ENGINEER'].includes(r);
}

/**
 * Retrieve current export configuration.
 */
export function getExportSettings() {
  try {
    const raw = localStorage.getItem(EXPORT_SETTINGS_STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw);
      return {
        ...DEFAULT_EXPORT_SETTINGS,
        ...parsed,
        maxRecords: parsed.maxRecords !== undefined ? Number(parsed.maxRecords) : DEFAULT_EXPORT_SETTINGS.maxRecords,
      };
    }
  } catch (e) {
    console.warn('[ExportSettings] Failed reading export settings from localStorage:', e);
  }
  return { ...DEFAULT_EXPORT_SETTINGS };
}

/**
 * Save export configuration with role validation.
 * Rejects modifications from 'OPERATOR'.
 */
export function saveExportSettings(newSettings, userRole, username = 'admin') {
  if (!canUserManageExportSettings(userRole)) {
    return {
      success: false,
      error: 'Akses ditolak: Hanya peran Administrator dan Engineer yang dapat mengubah setup batas ekspor data.',
    };
  }

  try {
    const current = getExportSettings();
    const updated = {
      ...current,
      ...newSettings,
      maxRecords: newSettings.maxRecords !== undefined ? Number(newSettings.maxRecords) : current.maxRecords,
      updatedBy: username || userRole,
      updatedAt: new Date().toISOString(),
    };

    localStorage.setItem(EXPORT_SETTINGS_STORAGE_KEY, JSON.stringify(updated));

    // Dispatch custom event for real-time synchronization across views
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('datalogger:export-settings-changed', { detail: updated }));
    }

    return { success: true, settings: updated };
  } catch (e) {
    return { success: false, error: e.message || 'Gagal menyimpan konfigurasi ekspor.' };
  }
}

/**
 * Format label for human-readable display of max record limit.
 */
export function formatExportLimitLabel(limit) {
  const num = Number(limit);
  if (num === 0) return 'Semua Data (Tanpa Batas)';
  if (num >= 1000) return `${num.toLocaleString('id-ID')} data`;
  return `${num} data`;
}
