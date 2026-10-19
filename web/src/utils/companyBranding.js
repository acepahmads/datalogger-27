/**
 * Company Branding & Organization Persistence Utility
 * 
 * Provides robust client-side persistence for both Customer (Client) and Maker (Vendor)
 * branding using IndexedDB for image blobs and localStorage for lightweight metadata.
 * Ensures large images are NOT stored as base64 in localStorage (preventing quota exhaustion).
 * Performs client-side image validation and downscaling to compact high-DPI dimensions.
 */

const DB_NAME = 'datalogger_storage';
const DB_VERSION = 1;
const STORE_NAME = 'branding';
const CUSTOMER_LOGO_KEY = 'customer_logo';
const MAKER_LOGO_KEY = 'maker_logo';
const LEGACY_LOGO_KEY = 'company_logo';
const METADATA_KEY = 'datalogger.company_branding_meta';
const ORGANIZATION_BRANDING_KEY = 'datalogger.organization_branding';

// Max dimensions for compact header logo (preserving aspect ratio)
const MAX_LOGO_WIDTH = 256;
const MAX_LOGO_HEIGHT = 96;
const MAX_FILE_SIZE_BYTES = 1024 * 1024; // 1 MB limit before compression

export const DEFAULT_BRANDING = {
  customer: {
    name: 'PT. Mitra Industri Enviro',
    subtitle: 'AQMS',
    hasLogo: false
  },
  maker: {
    name: 'PT. Centra Biotech Indonesia',
    subtitle: 'CBI Industrial IoT',
    hasLogo: false
  }
};

/**
 * Open or initialize the IndexedDB database
 */
function openDB() {
  return new Promise((resolve, reject) => {
    if (typeof window === 'undefined' || !window.indexedDB) {
      reject(new Error('IndexedDB is not supported in this environment.'));
      return;
    }
    const request = window.indexedDB.open(DB_NAME, DB_VERSION);
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME);
      }
    };
    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

/**
 * Optimize an image file to a compact canvas blob
 */
function optimizeImage(file) {
  return new Promise((resolve, reject) => {
    // If SVG, keep raw text/blob directly to maintain vector fidelity
    if (file.type === 'image/svg+xml') {
      const reader = new FileReader();
      reader.onload = () => resolve({ blob: file, dataUrl: reader.result, width: 0, height: 0 });
      reader.onerror = reject;
      reader.readAsDataURL(file);
      return;
    }

    const img = new Image();
    const objectUrl = URL.createObjectURL(file);
    img.onload = () => {
      URL.revokeObjectURL(objectUrl);
      let { width, height } = img;

      // Scale down if larger than maximum bounds
      const ratio = Math.min(MAX_LOGO_WIDTH / width, MAX_LOGO_HEIGHT / height, 1);
      const targetW = Math.round(width * ratio);
      const targetH = Math.round(height * ratio);

      const canvas = document.createElement('canvas');
      canvas.width = targetW;
      canvas.height = targetH;
      const ctx = canvas.getContext('2d');
      ctx.imageSmoothingEnabled = true;
      ctx.imageSmoothingQuality = 'high';
      ctx.drawImage(img, 0, 0, targetW, targetH);

      canvas.toBlob((blob) => {
        if (!blob) {
          reject(new Error('Image processing failed'));
          return;
        }
        const reader = new FileReader();
        reader.onload = () => resolve({ blob, dataUrl: reader.result, width: targetW, height: targetH });
        reader.onerror = reject;
        reader.readAsDataURL(blob);
      }, 'image/png');
    };
    img.onerror = () => {
      URL.revokeObjectURL(objectUrl);
      reject(new Error('Failed to load image file.'));
    };
    img.src = objectUrl;
  });
}

/**
 * Read blob from IndexedDB for a given key
 */
async function getBlobFromDB(key) {
  try {
    const db = await openDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readonly');
      const store = tx.objectStore(STORE_NAME);
      const req = store.get(key);
      req.onsuccess = () => resolve(req.result || null);
      req.onerror = () => reject(req.error);
    });
  } catch (e) {
    console.warn(`Could not read ${key} from IndexedDB:`, e);
    return null;
  }
}

/**
 * Save blob into IndexedDB for a given key
 */
async function putBlobIntoDB(key, blob) {
  const db = await openDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    const store = tx.objectStore(STORE_NAME);
    const req = store.put(blob, key);
    req.onsuccess = resolve;
    req.onerror = () => reject(req.error);
  });
}

/**
 * Delete blob from IndexedDB for a given key
 */
async function deleteBlobFromDB(key) {
  try {
    const db = await openDB();
    return new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readwrite');
      const store = tx.objectStore(STORE_NAME);
      const req = store.delete(key);
      req.onsuccess = resolve;
      req.onerror = () => reject(req.error);
    });
  } catch (e) {
    console.warn(`Could not delete ${key} from IndexedDB:`, e);
  }
}

/**
 * Helper to convert Blob to dataUrl
 */
function blobToDataUrl(blob) {
  return new Promise((resolve) => {
    if (!blob) return resolve(null);
    const reader = new FileReader();
    reader.onload = () => resolve(reader.result);
    reader.onerror = () => resolve(null);
    reader.readAsDataURL(blob);
  });
}

/**
 * Helper to fetch database branding from backend REST API
 */
export async function fetchBrandingFromDatabase() {
  if (typeof window === 'undefined') return null;
  try {
    const res = await fetch('/api/system/branding');
    if (!res.ok) return null;
    const json = await res.json();
    if (json && json.success && json.data) {
      return json.data;
    }
    return json.data || json;
  } catch (e) {
    return null;
  }
}

/**
 * Helper to sync branding changes to backend database REST API
 */
export async function syncBrandingToDatabase(payload) {
  if (typeof window === 'undefined') return;
  try {
    await fetch('/api/system/branding', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload)
    });
  } catch (e) {
    console.warn('Could not sync branding to database API:', e);
  }
}

/**
 * Read current organization metadata from localStorage with fallback defaults
 */
export function getOrganizationBrandingMeta() {
  try {
    if (typeof localStorage === 'undefined') return JSON.parse(JSON.stringify(DEFAULT_BRANDING));
    const raw = localStorage.getItem(ORGANIZATION_BRANDING_KEY);
    if (!raw) {
      // Check legacy single company metadata
      const legacyRaw = localStorage.getItem(METADATA_KEY);
      if (legacyRaw) {
        const legacy = JSON.parse(legacyRaw);
        return {
          customer: { ...DEFAULT_BRANDING.customer },
          maker: {
            name: legacy.companyName !== undefined ? legacy.companyName : DEFAULT_BRANDING.maker.name,
            subtitle: DEFAULT_BRANDING.maker.subtitle,
            hasLogo: !!legacy.hasLogo
          }
        };
      }
      return JSON.parse(JSON.stringify(DEFAULT_BRANDING));
    }
    const parsed = JSON.parse(raw);
    return {
      customer: {
        name: (parsed.customer && parsed.customer.name !== undefined) ? parsed.customer.name : DEFAULT_BRANDING.customer.name,
        subtitle: (parsed.customer && parsed.customer.subtitle !== undefined) ? (parsed.customer.subtitle === 'Local Autonomous Edge' ? '' : parsed.customer.subtitle) : DEFAULT_BRANDING.customer.subtitle,
        hasLogo: !!(parsed.customer && parsed.customer.hasLogo)
      },
      maker: {
        name: (parsed.maker && parsed.maker.name !== undefined) ? parsed.maker.name : DEFAULT_BRANDING.maker.name,
        subtitle: (parsed.maker && parsed.maker.subtitle !== undefined) ? parsed.maker.subtitle : DEFAULT_BRANDING.maker.subtitle,
        hasLogo: !!(parsed.maker && parsed.maker.hasLogo)
      }
    };
  } catch (e) {
    return JSON.parse(JSON.stringify(DEFAULT_BRANDING));
  }
}

/**
 * Notify all components of branding change
 */
function notifyBrandingChanged() {
  if (typeof window !== 'undefined') {
    try {
      window.dispatchEvent(new CustomEvent('organization-branding-updated', {
        detail: getOrganizationBrandingMeta()
      }));
    } catch (e) {
      // ignore
    }
  }
}

/**
 * Load complete organization branding (Customer + Maker) including logo data URLs
 * Synchronizes with backend database while maintaining local cached copy for zero-latency paint.
 */
export async function loadOrganizationBranding() {
  let meta = getOrganizationBrandingMeta();

  // Load customer logo from local cache
  let customerBlob = await getBlobFromDB(CUSTOMER_LOGO_KEY);
  let customerLogoUrl = await blobToDataUrl(customerBlob);

  // Load maker logo (with fallback to legacy key)
  let makerBlob = await getBlobFromDB(MAKER_LOGO_KEY);
  if (!makerBlob) {
    makerBlob = await getBlobFromDB(LEGACY_LOGO_KEY);
  }
  let makerLogoUrl = await blobToDataUrl(makerBlob);

  // Attempt database sync from backend API
  try {
    const remote = await fetchBrandingFromDatabase();
    if (remote) {
      let changed = false;
      if (remote.customer_name !== undefined && remote.customer_name !== meta.customer.name) {
        meta.customer.name = remote.customer_name;
        changed = true;
      }
      if (remote.customer_subtitle !== undefined && remote.customer_subtitle !== meta.customer.subtitle) {
        meta.customer.subtitle = remote.customer_subtitle === 'Local Autonomous Edge' ? '' : remote.customer_subtitle;
        changed = true;
      }
      if (remote.customer_has_logo !== undefined && remote.customer_has_logo !== meta.customer.hasLogo) {
        meta.customer.hasLogo = !!remote.customer_has_logo;
        changed = true;
      }
      if (remote.customer_logo && remote.customer_logo !== customerLogoUrl) {
        customerLogoUrl = remote.customer_logo;
      }

      if (remote.maker_name !== undefined && remote.maker_name !== meta.maker.name) {
        meta.maker.name = remote.maker_name;
        changed = true;
      }
      if (remote.maker_subtitle !== undefined && remote.maker_subtitle !== meta.maker.subtitle) {
        meta.maker.subtitle = remote.maker_subtitle;
        changed = true;
      }
      if (remote.maker_has_logo !== undefined && remote.maker_has_logo !== meta.maker.hasLogo) {
        meta.maker.hasLogo = !!remote.maker_has_logo;
        changed = true;
      }
      if (remote.maker_logo && remote.maker_logo !== makerLogoUrl) {
        makerLogoUrl = remote.maker_logo;
      }

      if (changed) {
        try {
          localStorage.setItem(ORGANIZATION_BRANDING_KEY, JSON.stringify(meta));
        } catch (e) {
          // ignore
        }
        notifyBrandingChanged();
      }
    }
  } catch (err) {
    // Graceful offline fallback
  }

  return {
    customer: {
      name: meta.customer.name,
      subtitle: meta.customer.subtitle,
      hasLogo: !!customerLogoUrl,
      logoUrl: customerLogoUrl
    },
    maker: {
      name: meta.maker.name,
      subtitle: meta.maker.subtitle,
      hasLogo: !!makerLogoUrl,
      logoUrl: makerLogoUrl
    }
  };
}

/**
 * Save Customer Logo & Name (persists locally and to database)
 */
export async function saveCustomerBranding({ file = null, name = '', subtitle = '' }) {
  const meta = getOrganizationBrandingMeta();
  let logoUrl = null;

  if (file) {
    const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg', 'image/svg+xml', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      throw new Error('Unsupported format. Please upload PNG, JPG, SVG, or WebP.');
    }
    if (file.size > MAX_FILE_SIZE_BYTES) {
      throw new Error('File size exceeds 1 MB limit.');
    }
    const optimized = await optimizeImage(file);
    await putBlobIntoDB(CUSTOMER_LOGO_KEY, optimized.blob);
    logoUrl = optimized.dataUrl;
    meta.customer.hasLogo = true;
  }

  if (name !== undefined) meta.customer.name = (name || '').trim();
  if (subtitle !== undefined) meta.customer.subtitle = (subtitle || '').trim();
  meta.customer.updatedAt = new Date().toISOString();

  try {
    localStorage.setItem(ORGANIZATION_BRANDING_KEY, JSON.stringify(meta));
  } catch (e) {
    console.warn('Could not save organization branding to localStorage', e);
  }

  // Persist to backend database API
  const dbPayload = {
    customer_name: meta.customer.name,
    customer_subtitle: meta.customer.subtitle,
    customer_has_logo: meta.customer.hasLogo
  };
  if (logoUrl) {
    dbPayload.customer_logo = logoUrl;
  }
  await syncBrandingToDatabase(dbPayload);

  notifyBrandingChanged();
  return { customer: meta.customer, logoUrl };
}

/**
 * Remove Customer Logo
 */
export async function removeCustomerLogo() {
  await deleteBlobFromDB(CUSTOMER_LOGO_KEY);
  const meta = getOrganizationBrandingMeta();
  meta.customer.hasLogo = false;
  meta.customer.updatedAt = new Date().toISOString();
  try {
    localStorage.setItem(ORGANIZATION_BRANDING_KEY, JSON.stringify(meta));
  } catch (e) {
    // ignore
  }
  await syncBrandingToDatabase({
    customer_has_logo: false,
    customer_logo: ''
  });
  notifyBrandingChanged();
}

/**
 * Save Maker Logo & Name (persists locally and to database)
 */
export async function saveMakerBranding({ file = null, name = '', subtitle = '' }) {
  const meta = getOrganizationBrandingMeta();
  let logoUrl = null;

  if (file) {
    const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg', 'image/svg+xml', 'image/webp'];
    if (!allowedTypes.includes(file.type)) {
      throw new Error('Unsupported format. Please upload PNG, JPG, SVG, or WebP.');
    }
    if (file.size > MAX_FILE_SIZE_BYTES) {
      throw new Error('File size exceeds 1 MB limit.');
    }
    const optimized = await optimizeImage(file);
    await putBlobIntoDB(MAKER_LOGO_KEY, optimized.blob);
    logoUrl = optimized.dataUrl;
    meta.maker.hasLogo = true;
  }

  if (name !== undefined) meta.maker.name = (name || '').trim();
  if (subtitle !== undefined) meta.maker.subtitle = (subtitle || '').trim();
  meta.maker.updatedAt = new Date().toISOString();

  try {
    localStorage.setItem(ORGANIZATION_BRANDING_KEY, JSON.stringify(meta));
    // Also sync legacy company name for any existing dependencies
    localStorage.setItem(METADATA_KEY, JSON.stringify({
      hasLogo: meta.maker.hasLogo,
      companyName: meta.maker.name,
      updatedAt: meta.maker.updatedAt
    }));
  } catch (e) {
    console.warn('Could not save organization branding to localStorage', e);
  }

  // Persist to backend database API
  const dbPayload = {
    maker_name: meta.maker.name,
    maker_subtitle: meta.maker.subtitle,
    maker_has_logo: meta.maker.hasLogo
  };
  if (logoUrl) {
    dbPayload.maker_logo = logoUrl;
  }
  await syncBrandingToDatabase(dbPayload);

  notifyBrandingChanged();
  return { maker: meta.maker, logoUrl };
}

/**
 * Remove Maker Logo
 */
export async function removeMakerLogo() {
  await deleteBlobFromDB(MAKER_LOGO_KEY);
  await deleteBlobFromDB(LEGACY_LOGO_KEY);
  const meta = getOrganizationBrandingMeta();
  meta.maker.hasLogo = false;
  meta.maker.updatedAt = new Date().toISOString();
  try {
    localStorage.setItem(ORGANIZATION_BRANDING_KEY, JSON.stringify(meta));
    localStorage.setItem(METADATA_KEY, JSON.stringify({
      hasLogo: false,
      companyName: meta.maker.name,
      updatedAt: meta.maker.updatedAt
    }));
  } catch (e) {
    // ignore
  }
  await syncBrandingToDatabase({
    maker_has_logo: false,
    maker_logo: ''
  });
  notifyBrandingChanged();
}

/**
 * Save All Organization Text Branding (Both Customer & Maker)
 */
export async function saveOrganizationTextBranding({ customerName, customerSubtitle, makerName, makerSubtitle }) {
  const meta = getOrganizationBrandingMeta();
  if (customerName !== undefined) meta.customer.name = (customerName || '').trim();
  if (customerSubtitle !== undefined) meta.customer.subtitle = (customerSubtitle || '').trim();
  if (makerName !== undefined) meta.maker.name = (makerName || '').trim();
  if (makerSubtitle !== undefined) meta.maker.subtitle = (makerSubtitle || '').trim();
  meta.updatedAt = new Date().toISOString();

  try {
    localStorage.setItem(ORGANIZATION_BRANDING_KEY, JSON.stringify(meta));
    localStorage.setItem(METADATA_KEY, JSON.stringify({
      hasLogo: meta.maker.hasLogo,
      companyName: meta.maker.name,
      updatedAt: meta.updatedAt
    }));
  } catch (e) {
    console.warn('Could not save organization branding to localStorage', e);
  }

  await syncBrandingToDatabase({
    customer_name: meta.customer.name,
    customer_subtitle: meta.customer.subtitle,
    maker_name: meta.maker.name,
    maker_subtitle: meta.maker.subtitle
  });

  notifyBrandingChanged();
  return meta;
}

/**
 * Reset Organization Branding to Default
 */
export async function resetOrganizationBranding() {
  await deleteBlobFromDB(CUSTOMER_LOGO_KEY);
  await deleteBlobFromDB(MAKER_LOGO_KEY);
  await deleteBlobFromDB(LEGACY_LOGO_KEY);
  try {
    localStorage.setItem(ORGANIZATION_BRANDING_KEY, JSON.stringify(DEFAULT_BRANDING));
    localStorage.removeItem(METADATA_KEY);
  } catch (e) {
    // ignore
  }
  await syncBrandingToDatabase({ reset: true });
  notifyBrandingChanged();
  return JSON.parse(JSON.stringify(DEFAULT_BRANDING));
}

// ---------------------------------------------------------------------------
// BACKWARDS-COMPATIBILITY EXPORTS
// ---------------------------------------------------------------------------

export async function saveCompanyLogo(file, companyName = '') {
  const res = await saveMakerBranding({ file, name: companyName });
  return { dataUrl: res.logoUrl, metadata: { ...res.maker, companyName: res.maker.name } };
}

export async function loadCompanyLogo() {
  const branding = await loadOrganizationBranding();
  return {
    dataUrl: branding.maker.logoUrl,
    metadata: {
      hasLogo: branding.maker.hasLogo,
      companyName: branding.maker.name
    }
  };
}

export async function removeCompanyLogo() {
  return removeMakerLogo();
}

export function updateCompanyName(companyName) {
  const meta = saveOrganizationTextBranding({ makerName: companyName });
  return { companyName: meta.maker.name };
}

export function getCompanyBrandingMeta() {
  const meta = getOrganizationBrandingMeta();
  return {
    hasLogo: meta.maker.hasLogo,
    companyName: meta.maker.name
  };
}
