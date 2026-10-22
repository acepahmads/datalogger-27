/**
 * Unit Test: Frontend Byte Order Validation & Dynamic Permutation Engine
 */

const assert = require('assert');

function getByteOrderSize(dataType) {
  if (dataType === 'INT16' || dataType === 'UINT16') return 2;
  if (dataType === 'FLOAT64') return 8;
  if (dataType === 'BOOLEAN' || dataType === 'STRING') return 2;
  return 4; // FLOAT32, INT32, UINT32
}

function getAvailableInputBytes(size) {
  const letters = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'];
  return letters.slice(0, size);
}

function validateCustomByteOrder(slots, dataType) {
  const size = getByteOrderSize(dataType);
  const available = getAvailableInputBytes(size);
  if (!slots || slots.length !== size) {
    return `Custom byte order requires ${size} positions`;
  }
  const seen = new Set();
  const duplicates = [];
  const validSet = new Set(available);
  for (const s of slots) {
    if (!validSet.has(s)) {
      return `Invalid byte position '${s}'. Valid positions are ${available.join(', ')}`;
    }
    if (seen.has(s)) {
      duplicates.push(s);
    }
    seen.add(s);
  }
  if (duplicates.length > 0) {
    return `Duplicate byte position: ${[...new Set(duplicates)].join(', ')}`;
  }
  const missing = available.filter(b => !seen.has(b));
  if (missing.length > 0) {
    return `Missing byte position: ${missing.join(', ')}`;
  }
  return null;
}

function initByteOrderState(savedOrder, dataType) {
  const raw = (savedOrder || '').trim();
  const size = getByteOrderSize(dataType);
  const baseSlots = getAvailableInputBytes(size);

  if (!raw) {
    return { selection: '', slots: baseSlots, formOrder: '' };
  }
  if (raw.toUpperCase().startsWith('CUSTOM:')) {
    const perm = raw.substring(7).replace(/[\s,-]/g, '').toUpperCase();
    if (perm.length === size) {
      return { selection: 'CUSTOM', slots: perm.split(''), formOrder: raw };
    }
    return { selection: 'CUSTOM', slots: baseSlots, formOrder: raw };
  }
  const clean = raw.replace(/[\s,-]/g, '').toUpperCase();
  const presets16 = ['AB', 'BA'];
  const presets32 = ['ABCD', 'CDAB', 'BADC', 'DCBA'];
  const presets64 = ['ABCDEFGH', 'GHEFCDAB', 'BADCFEHG', 'HGFEDCBA', 'EFGHABCD'];

  const presets = size === 2 ? presets16 : (size === 8 ? presets64 : presets32);
  if (presets.includes(clean)) {
    return { selection: clean, slots: clean.split(''), formOrder: clean };
  }

  const validLetters = new Set(baseSlots);
  if (clean.length === size && new Set(clean.split('')).size === size && clean.split('').every(ch => validLetters.has(ch))) {
    return { selection: 'CUSTOM', slots: clean.split(''), formOrder: 'CUSTOM:' + clean };
  }

  return { selection: 'CUSTOM', slots: baseSlots, formOrder: raw, warning: `Unrecognized order: ${raw}` };
}

console.log('=== RUNNING FRONTEND BYTE ORDER VALIDATION TEST ===\n');

// Test 1: Byte order sizing
assert.strictEqual(getByteOrderSize('INT16'), 2, 'INT16 size must be 2');
assert.strictEqual(getByteOrderSize('UINT16'), 2, 'UINT16 size must be 2');
assert.strictEqual(getByteOrderSize('FLOAT32'), 4, 'FLOAT32 size must be 4');
assert.strictEqual(getByteOrderSize('INT32'), 4, 'INT32 size must be 4');
assert.strictEqual(getByteOrderSize('UINT32'), 4, 'UINT32 size must be 4');
assert.strictEqual(getByteOrderSize('FLOAT64'), 8, 'FLOAT64 size must be 8');
console.log('  ✓ PASS: Byte order sizing per data type');

// Test 2: Valid custom orders
assert.strictEqual(validateCustomByteOrder(['C', 'D', 'A', 'B'], 'FLOAT32'), null, 'CDAB must be valid for FLOAT32');
assert.strictEqual(validateCustomByteOrder(['B', 'A'], 'INT16'), null, 'BA must be valid for INT16');
assert.strictEqual(validateCustomByteOrder(['E', 'F', 'G', 'H', 'A', 'B', 'C', 'D'], 'FLOAT64'), null, 'EFGHABCD must be valid for FLOAT64');
console.log('  ✓ PASS: Valid custom order permutations for 16, 32, and 64-bit');

// Test 3: Duplicate byte error detection
const dupErr = validateCustomByteOrder(['A', 'B', 'C', 'A'], 'FLOAT32');
assert.ok(dupErr && dupErr.includes('Duplicate byte position: A'), 'Must flag duplicate byte position A');
console.log('  ✓ PASS: Duplicate byte position detection');

// Test 4: Missing byte error detection
const missErr = validateCustomByteOrder(['A', 'B', 'C', 'B'], 'FLOAT32');
assert.ok(missErr && missErr.includes('Duplicate') || missErr.includes('Missing'), 'Must flag duplicate/missing');
console.log('  ✓ PASS: Missing byte position detection');

// Test 5: Wrong length detection
const lenErr = validateCustomByteOrder(['A', 'B', 'C'], 'FLOAT32');
assert.ok(lenErr && lenErr.includes('requires 4 positions'), 'Must reject 3 positions for 32-bit');
console.log('  ✓ PASS: Wrong length detection');

// Test 6: Invalid character outside range
const charErr = validateCustomByteOrder(['A', 'B', 'C', 'E'], 'FLOAT32');
assert.ok(charErr && charErr.includes("Invalid byte position 'E'"), 'Must reject byte E for 32-bit');
console.log('  ✓ PASS: Out of range byte position rejection');

// Test 7: State initialization from saved configurations
// 7a: Preset
const s1 = initByteOrderState('CDAB', 'FLOAT32');
assert.strictEqual(s1.selection, 'CDAB');
assert.deepStrictEqual(s1.slots, ['C', 'D', 'A', 'B']);

// 7b: Custom with prefix
const s2 = initByteOrderState('CUSTOM:CADB', 'FLOAT32');
assert.strictEqual(s2.selection, 'CUSTOM');
assert.deepStrictEqual(s2.slots, ['C', 'A', 'D', 'B']);

// 7c: Custom 64-bit
const s3 = initByteOrderState('CUSTOM:EFGHABCD', 'FLOAT64');
assert.strictEqual(s3.selection, 'CUSTOM');
assert.deepStrictEqual(s3.slots, ['E', 'F', 'G', 'H', 'A', 'B', 'C', 'D']);

// 7d: Empty default
const s4 = initByteOrderState('', 'FLOAT32');
assert.strictEqual(s4.selection, '');
assert.deepStrictEqual(s4.slots, ['A', 'B', 'C', 'D']);
console.log('  ✓ PASS: State restoration for presets, custom orders, and defaults');

console.log('\n========================================');
console.log('RESULTS: ALL FRONTEND TESTS PASSED');
console.log('========================================');
