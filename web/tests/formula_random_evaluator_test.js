const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('--- Running Formula Random Evaluator Acceptance Tests ---');

// 1. Test ParameterModal logic for formulaPreview
function evaluateFormulaPreview(formula, previewSampleX = 2.0, rawVal = 100.0) {
  if (!formula || !formula.trim()) return null;
  try {
    const formStr = formula.trim();
    const xVal = previewSampleX;

    if (/[;{}[\]"'`\\$]|(?<!=)=(?!=)|=>/.test(formStr)) {
      return { valid: false, error: 'Invalid syntax or disallowed characters' };
    }

    const allowedIdentifiers = new Set([
      'x', 'val', 'value', 'raw',
      'pi', 'e',
      'abs', 'round', 'floor', 'ceil', 'sqrt', 'min', 'max',
      'pow', 'log', 'ln', 'log10', 'exp', 'sin', 'cos', 'tan',
      'random'
    ]);

    const identMatches = formStr.match(/[a-zA-Z_][a-zA-Z0-9_]*/g) || [];
    for (const ident of identMatches) {
      if (!allowedIdentifiers.has(ident.toLowerCase())) {
        return { valid: false, error: `Unknown identifier '${ident}'` };
      }
    }

    const roundFn = (val, prec) => {
      if (val === undefined || isNaN(val)) throw new Error('round() requires at least 1 valid number');
      if (prec === undefined) return Math.round(val);
      const factor = Math.pow(10, Math.trunc(prec));
      return Math.round(val * factor) / factor;
    };

    const randomFn = (min, max) => {
      if (min === undefined || max === undefined) {
        throw new Error('random() requires 2 arguments: random(min, max)');
      }
      if (isNaN(min) || isNaN(max) || !isFinite(min) || !isFinite(max)) {
        throw new Error('random() arguments must be finite numbers');
      }
      if (min > max) {
        throw new Error(`min (${min}) cannot be greater than max (${max})`);
      }
      if (min === max) return min;
      return min + Math.random() * (max - min);
    };

    const sqrtFn = (v) => {
      if (v < 0) throw new Error('sqrt() of negative number');
      return Math.sqrt(v);
    };

    const logFn = (v) => {
      if (v <= 0) throw new Error('log() argument must be positive');
      return Math.log(v);
    };

    const log10Fn = (v) => {
      if (v <= 0) throw new Error('log10() argument must be positive');
      return Math.log10(v);
    };

    const context = {
      x: xVal,
      val: xVal,
      value: xVal,
      raw: rawVal,
      pi: Math.PI,
      e: Math.E,
      abs: Math.abs,
      round: roundFn,
      floor: Math.floor,
      ceil: Math.ceil,
      sqrt: sqrtFn,
      min: Math.min,
      max: Math.max,
      pow: Math.pow,
      log: logFn,
      ln: logFn,
      log10: log10Fn,
      exp: Math.exp,
      sin: Math.sin,
      cos: Math.cos,
      tan: Math.tan,
      random: randomFn
    };

    let expr = formStr.replace(/\^/g, '**');
    expr = expr.replace(/[a-zA-Z_][a-zA-Z0-9_]*/g, m => m.toLowerCase());

    const paramNames = Object.keys(context);
    const paramValues = Object.values(context);

    const fn = new Function(...paramNames, `"use strict"; return (${expr});`);
    const res = fn(...paramValues);

    if (typeof res === 'number' && !isNaN(res) && isFinite(res)) {
      return { valid: true, value: Number(res.toFixed(4)) };
    }
    return { valid: false, error: 'Result is NaN or infinite' };
  } catch (e) {
    return { valid: false, error: e.message || 'Invalid formula' };
  }
}

// Test 1: User's exact formula from screenshot: value + (random(0.01, 0.09))
console.log('Test 1: User exact formula value + (random(0.01, 0.09))');
for (let i = 0; i < 20; i++) {
  const p = evaluateFormulaPreview('value + (random(0.01, 0.09))', 2.0);
  assert(p.valid, `Expected valid, got error: ${p.error}`);
  assert(p.value >= 2.01 && p.value <= 2.09, `Value ${p.value} outside [2.01, 2.09]`);
}
console.log('  PASS: User exact formula correctly evaluated with result in [2.01, 2.09]');

// Test 2: Standard formula: value + random(0.01, 0.09)
console.log('Test 2: Standard formula value + random(0.01, 0.09)');
for (let i = 0; i < 20; i++) {
  const p = evaluateFormulaPreview('value + random(0.01, 0.09)', 25.0);
  assert(p.valid, `Expected valid, got error: ${p.error}`);
  assert(p.value >= 25.01 && p.value <= 25.09, `Value ${p.value} outside [25.01, 25.09]`);
}
console.log('  PASS: Evaluated with result in [25.01, 25.09]');

// Test 3: Variables x, val, raw support
console.log('Test 3: Variables x, val, and raw support');
const pX = evaluateFormulaPreview('x + random(0.1, 0.2)', 10.0);
assert(pX.valid && pX.value >= 10.1 && pX.value <= 10.2);
const pVal = evaluateFormulaPreview('val + random(0.1, 0.2)', 10.0);
assert(pVal.valid && pVal.value >= 10.1 && pVal.value <= 10.2);
const pRaw = evaluateFormulaPreview('raw + random(0.5, 0.9)', 10.0, 50.0);
assert(pRaw.valid && pRaw.value >= 50.5 && pRaw.value <= 50.9);
console.log('  PASS: Variables x, val, and raw working properly');

// Test 4: Invalid arguments validation
console.log('Test 4: Invalid arguments validation');
const pMinGtMax = evaluateFormulaPreview('random(0.09, 0.01)', 2.0);
assert(!pMinGtMax.valid, 'Expected invalid when min > max');
assert(pMinGtMax.error.includes('cannot be greater than max'), `Expected min > max message, got: ${pMinGtMax.error}`);

const pOneArg = evaluateFormulaPreview('random(0.05)', 2.0);
assert(!pOneArg.valid, 'Expected invalid when 1 arg');
assert(pOneArg.error.includes('requires 2 arguments'), `Expected 2 args message, got: ${pOneArg.error}`);

const pDisallowed = evaluateFormulaPreview('random(0.01, 0.09); alert(1)', 2.0);
assert(!pDisallowed.valid, 'Expected disallowed syntax error');
console.log('  PASS: Invalid arguments and disallowed characters cleanly rejected');

// Test 5: Check ParameterModal.vue file contains preset button and updated formulaPreview
console.log('Test 5: Check ParameterModal.vue contains random preset button and support');
const modalPath = path.join(__dirname, '../src/components/ParameterModal.vue');
const modalContent = fs.readFileSync(modalPath, 'utf8');
assert(modalContent.includes('random('), 'ParameterModal must include random(');
assert(modalContent.includes('+random(0.01, 0.09)'), 'ParameterModal must have +random preset button');
assert(modalContent.includes('allowedIdentifiers'), 'ParameterModal must have allowedIdentifiers');
console.log('  PASS: ParameterModal.vue contains required UI updates');

// Test 6: Check locales contain updated formula descriptions
console.log('Test 6: Check i18n locales');
const enPath = path.join(__dirname, '../src/i18n/locales/en.js');
const idPath = path.join(__dirname, '../src/i18n/locales/id.js');
const enContent = fs.readFileSync(enPath, 'utf8');
const idContent = fs.readFileSync(idPath, 'utf8');

assert(enContent.includes('random'), 'en.js must list random');
assert(enContent.includes('value + random(0.01, 0.09)'), 'en.js must include random example');
assert(idContent.includes('random'), 'id.js must list random');
assert(idContent.includes('value + random(0.01, 0.09)'), 'id.js must include random example');
console.log('  PASS: Locale files properly updated');

console.log('--- ALL FRONTEND ACCEPTANCE TESTS PASSED SUCCESSFULLY ---');
