/**
 * Company Branding & Logo Persistence Utility
 * 
 * Provides robust client-side persistence for company logo & name using IndexedDB.
 * Ensures large images are NOT stored as base64 in localStorage (preventing quota exhaustion).
 * Performs client-side image validation and downscaling to compact high-DPI dimensions.
 */

const DB_NAME = 'datalogger_storage';
const DB_VERSION = 1;
const STORE_NAME = 'branding';
const LOGO_KEY = 'company_logo';
const METADATA_KEY = 'datalogger.company_branding_meta';

// Max dimensions for compact header logo (preserving aspect ratio)
const MAX_LOGO_WIDTH = 256;
const MAX_LOGO_HEIGHT = 96;
const MAX_FILE_SIZE_BYTES = 1024 * 1024; // 1 MB limit before compression

/**
 * Open or initialize the IndexedDB database
 */
function openDB() {
  return new Promise((resolve, reject) => {
    if (!window.indexedDB) {
      reject(new Error('IndexedDB is not supported in this browser.'));
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
 * Save company logo to IndexedDB and update lightweight metadata
 */
export async function saveCompanyLogo(file, companyName = '') {
  if (!file) throw new Error('No file provided.');

  const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg', 'image/svg+xml', 'image/webp'];
  if (!allowedTypes.includes(file.type)) {
    throw new Error('Unsupported format. Please upload PNG, JPG, SVG, or WebP.');
  }

  if (file.size > MAX_FILE_SIZE_BYTES) {
    throw new Error('File size exceeds 1 MB limit.');
  }

  const { blob, dataUrl, width, height } = await optimizeImage(file);
  const db = await openDB();

  await new Promise((resolve, reject) => {
    const tx = db.transaction(STORE_NAME, 'readwrite');
    const store = tx.objectStore(STORE_NAME);
    const req = store.put(blob, LOGO_KEY);
    req.onsuccess = resolve;
    req.onerror = () => reject(req.error);
  });

  const metadata = {
    hasLogo: true,
    companyName: companyName.trim() || 'CBI Industrial',
    fileName: file.name,
    mimeType: file.type,
    width,
    height,
    updatedAt: new Date().toISOString()
  };

  try {
    localStorage.setItem(METADATA_KEY, JSON.stringify(metadata));
  } catch (e) {
    console.warn('Could not save metadata to localStorage', e);
  }

  return { dataUrl, metadata };
}

/**
 * Load company logo from IndexedDB
 */
export async function loadCompanyLogo() {
  try {
    const rawMeta = localStorage.getItem(METADATA_KEY);
    const metadata = rawMeta ? JSON.parse(rawMeta) : { hasLogo: false, companyName: 'CBI Industrial' };

    const db = await openDB();
    const blob = await new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readonly');
      const store = tx.objectStore(STORE_NAME);
      const req = store.get(LOGO_KEY);
      req.onsuccess = () => resolve(req.result);
      req.onerror = () => reject(req.error);
    });

    if (!blob) {
      return { dataUrl: null, metadata: { ...metadata, hasLogo: false } };
    }

    const dataUrl = await new Promise((resolve) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result);
      reader.onerror = () => resolve(null);
      reader.readAsDataURL(blob);
    });

    return { dataUrl, metadata };
  } catch (err) {
    console.warn('Error loading company logo from IndexedDB:', err);
    return { dataUrl: null, metadata: { hasLogo: false, companyName: 'CBI Industrial' } };
  }
}

/**
 * Remove company logo from IndexedDB
 */
export async function removeCompanyLogo() {
  try {
    const db = await openDB();
    await new Promise((resolve, reject) => {
      const tx = db.transaction(STORE_NAME, 'readwrite');
      const store = tx.objectStore(STORE_NAME);
      const req = store.delete(LOGO_KEY);
      req.onsuccess = resolve;
      req.onerror = () => reject(req.error);
    });
  } catch (err) {
    console.warn('Could not delete logo from IndexedDB:', err);
  }

  try {
    const rawMeta = localStorage.getItem(METADATA_KEY);
    const prev = rawMeta ? JSON.parse(rawMeta) : {};
    localStorage.setItem(METADATA_KEY, JSON.stringify({
      ...prev,
      hasLogo: false,
      updatedAt: new Date().toISOString()
    }));
  } catch (e) {
    // ignore
  }
}

/**
 * Update company name in metadata
 */
export function updateCompanyName(companyName) {
  try {
    const rawMeta = localStorage.getItem(METADATA_KEY);
    const prev = rawMeta ? JSON.parse(rawMeta) : { hasLogo: false };
    const updated = {
      ...prev,
      companyName: companyName.trim() || 'CBI Industrial',
      updatedAt: new Date().toISOString()
    };
    localStorage.setItem(METADATA_KEY, JSON.stringify(updated));
    return updated;
  } catch (e) {
    return { companyName };
  }
}

/**
 * Get current company branding metadata synchronously
 */
export function getCompanyBrandingMeta() {
  try {
    const rawMeta = localStorage.getItem(METADATA_KEY);
    return rawMeta ? JSON.parse(rawMeta) : { hasLogo: false, companyName: 'CBI Industrial' };
  } catch (e) {
    return { hasLogo: false, companyName: 'CBI Industrial' };
  }
}
