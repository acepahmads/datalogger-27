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
            name: legacy.companyName || DEFAULT_BRANDING.maker.name,
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
        name: (parsed.customer && parsed.customer.name) || DEFAULT_BRANDING.customer.name,
        subtitle: (parsed.customer && parsed.customer.subtitle && parsed.customer.subtitle !== 'Local Autonomous Edge') ? parsed.customer.subtitle : DEFAULT_BRANDING.customer.subtitle,
        hasLogo: !!(parsed.customer && parsed.customer.hasLogo)
      },
      maker: {
        name: (parsed.maker && parsed.maker.name) || DEFAULT_BRANDING.maker.name,
        subtitle: (parsed.maker && parsed.maker.subtitle) || DEFAULT_BRANDING.maker.subtitle,
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
 */
export async function loadOrganizationBranding() {
  const meta = getOrganizationBrandingMeta();

  // Load customer logo
  let customerBlob = await getBlobFromDB(CUSTOMER_LOGO_KEY);
  let customerLogoUrl = await blobToDataUrl(customerBlob);

  // Load maker logo (with fallback to legacy key)
  let makerBlob = await getBlobFromDB(MAKER_LOGO_KEY);
  if (!makerBlob) {
    makerBlob = await getBlobFromDB(LEGACY_LOGO_KEY);
  }
  let makerLogoUrl = await blobToDataUrl(makerBlob);

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
 * Save Customer Logo & Name
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

  if (name.trim()) meta.customer.name = name.trim();
  if (subtitle.trim()) meta.customer.subtitle = subtitle.trim();
  meta.customer.updatedAt = new Date().toISOString();

  try {
    localStorage.setItem(ORGANIZATION_BRANDING_KEY, JSON.stringify(meta));
  } catch (e) {
    console.warn('Could not save organization branding to localStorage', e);
  }

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
  notifyBrandingChanged();
}

/**
 * Save Maker Logo & Name
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

  if (name.trim()) meta.maker.name = name.trim();
  if (subtitle.trim()) meta.maker.subtitle = subtitle.trim();
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
  notifyBrandingChanged();
}

/**
 * Save All Organization Text Branding (Both Customer & Maker)
 */
export function saveOrganizationTextBranding({ customerName, customerSubtitle, makerName, makerSubtitle }) {
  const meta = getOrganizationBrandingMeta();
  if (customerName != null) meta.customer.name = customerName.trim() || DEFAULT_BRANDING.customer.name;
  if (customerSubtitle != null) meta.customer.subtitle = customerSubtitle.trim() || DEFAULT_BRANDING.customer.subtitle;
  if (makerName != null) meta.maker.name = makerName.trim() || DEFAULT_BRANDING.maker.name;
  if (makerSubtitle != null) meta.maker.subtitle = makerSubtitle.trim() || DEFAULT_BRANDING.maker.subtitle;
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
