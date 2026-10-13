<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
        <div>
          <div class="flex items-center space-x-2">
            <h2 class="text-base font-bold text-white tracking-wide">
              {{ modalTitle }}
            </h2>
            <span v-if="isCopying" class="px-2 py-0.5 text-3xs font-mono font-bold uppercase rounded bg-cyan-950/80 text-cyan-300 border border-cyan-800/80">
              CLONE
            </span>
          </div>
          <p class="text-xs text-slate-400 mt-0.5">
            {{ modalSubtitle }}
          </p>
        </div>
        <button @click="closeModal" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Modal Body -->
      <form ref="formBody" @submit.prevent="submitForm" class="flex-1 overflow-y-auto p-6 space-y-4">
        <!-- Top Error Alert -->
        <div v-if="error" class="p-3.5 bg-rose-500/15 border-2 border-rose-500/40 rounded-xl text-xs text-rose-300 flex items-start space-x-2.5 shadow-lg shadow-rose-950/20 animate-fade-in">
          <svg class="w-4 h-4 text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
          </svg>
          <div class="flex-1">
            <span class="font-bold text-rose-200 block text-2xs">{{ $t('parameterModal.cannotSaveTitle') }}</span>
            <span class="text-3xs text-rose-300 mt-0.5 block leading-relaxed whitespace-pre-wrap">{{ error }}</span>
          </div>
          <button type="button" @click="error = null" class="text-rose-400 hover:text-white text-3xs font-mono ml-2">✕</button>
        </div>

        <!-- Copying Parameter Notice Banner -->
        <div v-if="isCopying" class="p-3.5 bg-cyan-950/40 border border-cyan-800/80 rounded-xl space-y-2.5 animate-fade-in">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2 text-cyan-200 text-xs">
              <svg class="w-4 h-4 text-cyan-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
              </svg>
              <span class="font-semibold">{{ $t('parameterModal.copyBannerTitle', { code: sourceParam ? (sourceParam.parameter_name || sourceParam.parameter_code) : '' }) }}</span>
            </div>
            <span class="text-3xs uppercase tracking-wider px-2 py-0.5 rounded bg-cyan-900/60 text-cyan-300 font-bold border border-cyan-700/50">
              CONFIG CLONE
            </span>
          </div>
          <p class="text-3xs text-slate-400 leading-relaxed">
            {{ $t('parameterModal.copyBannerDesc') }}
          </p>

          <!-- Destination Device Selector -->
          <div v-if="devicesList && devicesList.length > 0" class="pt-2 border-t border-slate-800/80">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
                  {{ $t('parameterModal.destinationDevice') }}
                </label>
                <span class="text-3xs text-slate-400">{{ $t('parameterModal.destinationDeviceHelp') }}</span>
              </div>
              <select
                v-model="targetDeviceId"
                class="px-2.5 py-1.5 bg-[#0B0F19] border border-cyan-800/80 rounded-lg text-xs text-white focus:outline-none focus:border-cyan-500 font-mono min-w-[200px]"
              >
                <option v-for="dev in devicesList" :key="dev.id" :value="dev.id">
                  {{ dev.device_code || dev.code }} ({{ dev.communication_protocol || dev.protocol }})
                </option>
              </select>
            </div>

            <!-- Protocol Mismatch Warning -->
            <div v-if="destProtocolMismatch" class="mt-2 p-2 bg-amber-500/10 border border-amber-500/30 rounded-lg text-3xs text-amber-300 flex items-start space-x-2">
              <span class="text-amber-400 font-bold">⚠️</span>
              <span>{{ $t('parameterModal.destProtocolMismatch', { destProto: destProtocolName, srcProto: srcProtocolName }) }}</span>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Parameter Code -->
          <div>
            <div class="flex items-center justify-between mb-1.5">
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
                {{ $t('parameterModal.code') }}
              </label>
              <span v-if="isCopying" class="text-3xs font-mono text-cyan-400 bg-cyan-950/60 px-1.5 py-0.5 rounded border border-cyan-800/60">
                Must be unique
              </span>
            </div>
            <input
              v-model="form.parameter_code"
              type="text"
              required
              :placeholder="$t('parameterModal.placeholders.code')"
              :class="isCopying ? 'border-cyan-500/80 focus:border-cyan-400' : 'border-slate-700/80 focus:border-blue-500'"
              class="w-full px-3 py-2 bg-[#0B0F19] border rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none font-mono"
            />
          </div>

          <!-- Parameter Name -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.name') }}
            </label>
            <input
              v-model="form.parameter_name"
              type="text"
              required
              :placeholder="$t('parameterModal.placeholders.name')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>
        </div>

        <!-- Parameter Icon & Type Configuration Row -->
        <div class="p-3 bg-[#0B0F19]/60 border border-slate-800 rounded-xl space-y-2">
          <div class="flex flex-wrap items-center justify-between gap-2">
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.iconLabel') || 'Parameter Icon / Sensor Type' }}
            </label>
            <!-- Live Preview Badge -->
            <div class="flex items-center space-x-1.5 text-2xs font-mono px-2 py-1 bg-slate-900 border border-slate-700/80 rounded-lg">
              <span class="text-slate-400 text-3xs">{{ $t('parameterModal.iconCardPreview') || 'Card Preview' }}:</span>
              <span class="text-cyan-400/80 font-semibold text-3xs">AQMS-01</span>
              <span class="text-slate-500">·</span>
              <div class="flex items-center space-x-1 font-bold text-white">
                <ParameterIcon :icon="form.icon" :code="form.parameter_code" :name="form.parameter_name" :unit="form.unit" size-class="w-3.5 h-3.5" />
                <span class="truncate max-w-[130px]">{{ form.parameter_name || form.parameter_code || 'Preview' }}</span>
              </div>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-3 items-center">
            <select
              v-model="form.icon"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="">{{ $t('parameterModal.iconAuto') || '✨ Auto-Detect (Based on Code/Name/Unit)' }}</option>
              <option v-for="cat in iconCatalog" :key="cat.key" :value="cat.key">
                {{ cat.label }}
              </option>
            </select>
            <div class="text-3xs text-slate-400 flex items-center space-x-1.5">
              <span>{{ $t('parameterModal.iconResolved') || 'Resolved Icon' }}:</span>
              <span class="font-bold text-cyan-400 capitalize">{{ currentResolvedIconKey }}</span>
              <span v-if="!form.icon" class="text-slate-500 font-sans italic">(automatic heuristic)</span>
              <span v-else class="text-emerald-400 font-sans italic">(manually configured)</span>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <!-- Data Type -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.dataType') }}
            </label>
            <select
              v-model="form.data_type"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="FLOAT32">FLOAT32 (Single Precision)</option>
              <option value="FLOAT64">FLOAT64 (Double Precision)</option>
              <option value="INT16">INT16 (Signed 16-bit)</option>
              <option value="INT32">INT32 (Signed 32-bit)</option>
              <option value="UINT16">UINT16 (Unsigned 16-bit)</option>
              <option value="UINT32">UINT32 (Unsigned 32-bit)</option>
              <option value="BOOLEAN">BOOLEAN (Digital State)</option>
              <option value="STRING">STRING (ASCII)</option>
            </select>
          </div>

          <!-- Unit -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.unit') }}
            </label>
            <input
              v-model="form.unit"
              type="text"
              :placeholder="$t('parameterModal.placeholders.unit')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <!-- Decimal Precision -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.precision') }}
            </label>
            <input
              v-model.number="form.precision"
              type="number"
              min="0"
              max="6"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <!-- Register Address -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.addressOffset') }}
            </label>
            <input
              v-model.number="form.register_address"
              type="number"
              placeholder="e.g. 30001, 40001"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>

          <!-- Register Type -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.registerType') }}
            </label>
            <select
              v-model="form.register_type"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="HOLDING_REGISTER">Holding Register (4x)</option>
              <option value="INPUT_REGISTER">Input Register (3x)</option>
              <option value="COIL">Coil (0x)</option>
              <option value="DISCRETE_INPUT">Discrete Input (1x)</option>
              <option value="VARIABLE">Virtual / Internal Variable</option>
            </select>
          </div>

          <!-- Byte Order Override -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.byteOrderOverride') }}
            </label>
            <select
              v-model="form.byte_order"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="">{{ $t('parameterModal.defaultDeviceConn') }}</option>
              <option value="ABCD">ABCD (Big Endian)</option>
              <option value="CDAB">CDAB (Word Swap)</option>
              <option value="BADC">BADC (Byte Swap)</option>
              <option value="DCBA">DCBA (Little Endian)</option>
            </select>
          </div>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- Scale Factor -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.scaleFactor') }}
            </label>
            <input
              v-model.number="form.scale"
              type="number"
              step="0.001"
              placeholder="1.0"
              :class="scaleZeroWarning ? 'border-rose-500/80 focus:border-rose-400' : 'border-slate-700/80 focus:border-blue-500'"
              class="w-full px-3 py-2 bg-[#0B0F19] border rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none font-mono transition"
            />
            <span v-if="scaleZeroWarning" class="text-3xs text-rose-400 mt-1 block">⚠️ {{ scaleZeroWarning }}</span>
          </div>

          <!-- Offset -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('parameterModal.offsetAdder') }}
            </label>
            <input
              v-model.number="form.offset"
              type="number"
              step="0.01"
              placeholder="0.0"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
            />
          </div>
        </div>

        <!-- Minimum & Maximum Value Limit (Hard Bounds) -->
        <div class="space-y-2">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Min Value -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('parameterModal.minLimit') }}
              </label>
              <input
                v-model.number="form.min_value"
                type="number"
                step="any"
                placeholder="e.g. 0.0"
                :class="hardLimitWarning ? 'border-rose-500/80 focus:border-rose-400' : 'border-slate-700/80 focus:border-blue-500'"
                class="w-full px-3 py-2 bg-[#0B0F19] border rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none font-mono transition"
              />
            </div>

            <!-- Max Value -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('parameterModal.maxLimit') }}
              </label>
              <input
                v-model.number="form.max_value"
                type="number"
                step="any"
                placeholder="e.g. 500.0"
                :class="hardLimitWarning ? 'border-rose-500/80 focus:border-rose-400' : 'border-slate-700/80 focus:border-blue-500'"
                class="w-full px-3 py-2 bg-[#0B0F19] border rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none font-mono transition"
              />
            </div>
          </div>
          <!-- Hard Limit Validation Warning -->
          <div v-if="hardLimitWarning" class="p-2.5 rounded-xl bg-rose-500/15 border border-rose-500/40 text-2xs text-rose-300 flex items-start space-x-2 animate-fade-in">
            <svg class="w-4 h-4 text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
            </svg>
            <span class="leading-relaxed">{{ hardLimitWarning }}</span>
          </div>
        </div>

        <!-- Warning Limits (Soft Limits: Phase 3.2 UNCERTAIN Quality) -->
        <div class="space-y-2">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-2xs font-semibold text-amber-400/90 uppercase tracking-wider mb-1.5 flex items-center justify-between">
                <span>{{ $t('parameterModal.warningLow') }}</span>
                <span class="text-3xs text-amber-500/80 font-mono">UNCERTAIN</span>
              </label>
              <input
                v-model.number="form.warning_low"
                type="number"
                step="any"
                :placeholder="$t('parameterModal.placeholders.warningLow')"
                :class="warningLimitWarning ? 'border-rose-500/80 focus:border-rose-400' : 'border-amber-500/30 focus:border-amber-400'"
                class="w-full px-3 py-2 bg-[#0B0F19] border rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none font-mono transition"
              />
            </div>

            <div>
              <label class="block text-2xs font-semibold text-amber-400/90 uppercase tracking-wider mb-1.5 flex items-center justify-between">
                <span>{{ $t('parameterModal.warningHigh') }}</span>
                <span class="text-3xs text-amber-500/80 font-mono">UNCERTAIN</span>
              </label>
              <input
                v-model.number="form.warning_high"
                type="number"
                step="any"
                :placeholder="$t('parameterModal.placeholders.warningHigh')"
                :class="warningLimitWarning ? 'border-rose-500/80 focus:border-rose-400' : 'border-amber-500/30 focus:border-amber-400'"
                class="w-full px-3 py-2 bg-[#0B0F19] border rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none font-mono transition"
              />
            </div>
          </div>

          <!-- Warning Limits Validation Callout -->
          <div v-if="warningLimitWarning" class="p-2.5 rounded-xl bg-rose-500/15 border border-rose-500/40 text-2xs text-rose-300 flex items-start space-x-2 animate-fade-in">
            <svg class="w-4 h-4 text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
            </svg>
            <span class="leading-relaxed">{{ warningLimitWarning }}</span>
          </div>
          <div v-else-if="softBeyondHardNotice" class="p-2.5 rounded-xl bg-amber-500/10 border border-amber-500/30 text-2xs text-amber-300 flex items-start space-x-2 animate-fade-in">
            <svg class="w-4 h-4 text-amber-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <span class="leading-relaxed">{{ softBeyondHardNotice }}</span>
          </div>
        </div>

        <!-- Mathematical Formula Evaluation (Python-style eval) -->
        <div class="p-3.5 bg-slate-900/60 rounded-xl border border-slate-800 space-y-2.5">
          <div class="flex items-center justify-between">
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.formula') }}
            </label>
            <span class="text-3xs text-blue-400 font-mono">eval(x, raw)</span>
          </div>
          <input
            v-model="form.formula"
            type="text"
            :placeholder="$t('parameterModal.formulaPlaceholder')"
            class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
          />
          <p class="text-3xs text-slate-400 leading-relaxed">
            {{ $t('parameterModal.formulaDesc') }}
          </p>

          <!-- Quick Presets -->
          <div class="flex items-center space-x-1.5 flex-wrap gap-y-1.5 pt-1">
            <span class="text-3xs text-slate-400 mr-1">{{ $t('parameterModal.presets') }}:</span>
            <button
              type="button"
              @click="form.formula = 'x * 1.8 + 32'"
              class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              °C➔°F (x*1.8+32)
            </button>
            <button
              type="button"
              @click="form.formula = '(raw - 4.0) * (100.0 / 16.0)'"
              class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              4-20mA➔%
            </button>
            <button
              type="button"
              @click="form.formula = 'round(x, 2)'"
              class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              round(x, 2)
            </button>
            <button
              type="button"
              @click="form.formula = 'x / 1000'"
              class="px-2 py-0.5 rounded text-3xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              W➔kW (x/1000)
            </button>
          </div>

          <!-- Live Preview Banner -->
          <div v-if="form.formula && form.formula.trim()" class="p-2 rounded-lg bg-[#0B0F19] border border-slate-800 flex items-center justify-between text-2xs font-mono">
            <span class="text-slate-400">{{ $t('parameterModal.formulaPreview') }} (x={{ previewSampleX }}):</span>
            <span v-if="formulaPreview && formulaPreview.valid" class="font-bold text-emerald-400">
              ➔ {{ formulaPreview.value }} {{ form.unit }}
            </span>
            <span v-else class="text-rose-400 font-semibold">
              ⚠️ {{ (formulaPreview && formulaPreview.error) || 'Invalid expression' }}
            </span>
          </div>
        </div>

        <!-- Anomaly & Timeout Hold-Last-Good-Value Protection -->
        <div class="p-3.5 bg-slate-900/60 rounded-xl border border-slate-800 space-y-3">
          <label class="flex items-center space-x-3 cursor-pointer">
            <input
              v-model="form.hold_last_value_enabled"
              type="checkbox"
              class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
            />
            <div>
              <span class="text-xs font-semibold text-slate-200">{{ $t('parameterModal.holdLastValueTitle') }}</span>
              <span class="block text-3xs text-slate-400 mt-0.5">{{ $t('parameterModal.holdLastValueDesc') }}</span>
            </div>
          </label>

          <!-- Duration Input (Conditional when enabled) -->
          <div v-if="form.hold_last_value_enabled" class="pl-7 pt-2 border-t border-slate-800/80 space-y-1.5">
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.holdDurationSeconds') }}
            </label>
            <div class="flex items-center space-x-3">
              <input
                v-model.number="form.hold_last_value_seconds"
                type="number"
                min="10"
                max="86400"
                step="10"
                placeholder="120"
                class="w-32 px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
              />
              <span class="text-3xs text-slate-400 font-mono">
                ({{ Math.round((form.hold_last_value_seconds || 120) / 60 * 10) / 10 }} min)
              </span>
            </div>
            <p class="text-3xs text-amber-400/90 leading-relaxed">
              {{ $t('parameterModal.holdDurationHelp') }}
            </p>
          </div>
        </div>

        <!-- Phase 3.2 Data Quality & Processing Settings -->
        <div class="p-3.5 bg-slate-900/60 rounded-xl border border-slate-800 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-2xs font-bold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.qualitySettingsTitle') }}
            </span>
            <span class="text-3xs text-blue-400 font-mono">Phase 3.2</span>
          </div>

          <!-- Quality Validation Toggle -->
          <label class="flex items-center space-x-3 cursor-pointer">
            <input
              v-model="form.quality_validation_enabled"
              type="checkbox"
              class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
            />
            <div>
              <span class="text-xs font-semibold text-slate-200">{{ $t('parameterModal.qualityValidationEnabled') }}</span>
              <span class="block text-3xs text-slate-400">{{ $t('parameterModal.qualityValidationDesc') }}</span>
            </div>
          </label>

          <!-- Stale Timeout Input -->
          <div class="pt-2 border-t border-slate-800/80 space-y-1.5">
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
              {{ $t('parameterModal.staleTimeoutSeconds') }}
            </label>
            <div class="flex items-center space-x-3">
              <input
                v-model.number="form.stale_timeout_seconds"
                type="number"
                min="10"
                max="86400"
                step="10"
                placeholder="120"
                class="w-32 px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
              />
              <span class="text-3xs text-slate-400 font-mono">
                ({{ Math.round((form.stale_timeout_seconds || 120) / 60 * 10) / 10 }} min)
              </span>
            </div>
            <p class="text-3xs text-slate-400 leading-relaxed">
              {{ $t('parameterModal.staleTimeoutHelp') }}
            </p>
          </div>

          <!-- Spike Detection Toggle & Config -->
          <div class="pt-2 border-t border-slate-800/80 space-y-2">
            <label class="flex items-center space-x-3 cursor-pointer">
              <input
                v-model="form.spike_detection_enabled"
                type="checkbox"
                class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
              />
              <div>
                <span class="text-xs font-semibold text-slate-200">{{ $t('parameterModal.spikeDetectionEnabled') }}</span>
                <span class="block text-3xs text-slate-400">{{ $t('parameterModal.spikeDetectionDesc') }}</span>
              </div>
            </label>

            <div v-if="form.spike_detection_enabled" class="pl-7 grid grid-cols-1 md:grid-cols-2 gap-3 pt-1">
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                  {{ $t('parameterModal.spikeThreshold') }}
                </label>
                <input
                  v-model.number="form.spike_threshold"
                  type="number"
                  step="any"
                  placeholder="e.g. 20.0"
                  class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
                />
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                  {{ $t('parameterModal.spikeWindowSize') }}
                </label>
                <input
                  v-model.number="form.spike_window_size"
                  type="number"
                  min="2"
                  max="10"
                  placeholder="3"
                  class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Description -->
        <div>
          <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
            {{ $t('common.description') }}
          </label>
          <input
            v-model="form.description"
            type="text"
            :placeholder="$t('parameterModal.placeholders.description')"
            class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
          />
        </div>

        <!-- Enabled Switch -->
        <div>
          <label class="flex items-center space-x-3 cursor-pointer p-2.5 rounded-xl bg-[#0B0F19] border border-slate-800">
            <input
              v-model="form.enabled"
              type="checkbox"
              class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
            />
            <div>
              <span class="text-xs font-semibold text-slate-200">{{ $t('parameterModal.paramActive') }}</span>
              <span class="block text-3xs text-slate-400">{{ $t('parameterModal.collectChannel') }}</span>
            </div>
          </label>
        </div>

        <!-- Sticky Warning Alert Above Buttons -->
        <div v-if="activeErrorMessage" class="p-3.5 bg-rose-500/15 border-2 border-rose-500/50 rounded-xl text-xs text-rose-200 flex items-start space-x-2.5 shadow-lg shadow-rose-950/40 animate-fade-in">
          <svg class="w-5 h-5 text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
          </svg>
          <div class="flex-1">
            <div class="font-bold flex items-center justify-between text-rose-100">
              <span class="text-xs">{{ $t('parameterModal.cannotSaveTitle') }}</span>
              <button v-if="error" type="button" @click="error = null" class="text-rose-400 hover:text-white text-3xs underline ml-2">{{ $t('parameterModal.closeWarning') }}</button>
            </div>
            <p class="mt-1 text-2xs text-rose-300 leading-relaxed font-sans whitespace-pre-wrap">{{ activeErrorMessage }}</p>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="pt-4 border-t border-slate-800 flex items-center justify-between">
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-2 rounded-xl text-xs font-medium text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            {{ $t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="loading || !!clientValidationError"
            :class="clientValidationError ? 'bg-slate-700 text-slate-400 cursor-not-allowed opacity-75' : 'bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-900/30'"
            class="px-5 py-2 rounded-xl text-xs font-bold transition flex items-center space-x-1.5"
            :title="clientValidationError || ''"
          >
            <span v-if="loading" class="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
            <span v-if="clientValidationError" class="text-2xs text-amber-300">⚠️ {{ $t('parameterModal.fixLimitsBeforeSave') }}</span>
            <span v-else>{{ isCopying ? $t('parameterModal.copyButton') : (isEditing ? $t('parameterModal.updateParam') : $t('parameterModal.addParam')) }}</span>
          </button>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import ParameterIcon from './ParameterIcon.vue';
import { PARAMETER_ICON_CATALOG, resolveParameterIconKey } from '../utils/parameterIcons.js';

export default {
  name: 'ParameterModal',
  components: {
    ParameterIcon,
  },
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
    deviceId: {
      type: [Number, String],
      required: true,
    },
    paramToEdit: {
      type: Object,
      default: null,
    },
    isCopying: {
      type: Boolean,
      default: false,
    },
    sourceParam: {
      type: Object,
      default: null,
    },
    sourceDevice: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      loading: false,
      error: null,
      targetDeviceId: null,
      form: {
        parameter_code: '',
        parameter_name: '',
        data_type: 'FLOAT32',
        unit: '',
        icon: '',
        description: '',
        register_address: 0,
        register_type: 'HOLDING_REGISTER',
        byte_order: '',
        min_value: null,
        max_value: null,
        warning_low: null,
        warning_high: null,
        quality_validation_enabled: true,
        processing_enabled: true,
        stale_timeout_seconds: 120,
        spike_detection_enabled: false,
        spike_threshold: 0.0,
        spike_window_size: 3,
        precision: 2,
        scale: 1.0,
        offset: 0.0,
        enabled: true,
        formula: '',
        hold_last_value_enabled: false,
        hold_last_value_seconds: 120,
      },
    };
  },
  computed: {
    isEditing() {
      return !this.isCopying && !!this.paramToEdit;
    },
    modalTitle() {
      if (this.isCopying) return this.$t('parameterModal.titleCopy');
      return this.isEditing ? this.$t('parameterModal.titleEdit') : this.$t('parameterModal.titleAdd');
    },
    modalSubtitle() {
      if (this.isCopying) {
        const code = this.sourceParam ? (this.sourceParam.parameter_name || this.sourceParam.parameter_code) : '';
        return this.$t('parameterModal.subtitleCopy', { source: code });
      }
      return this.isEditing ? this.$t('parameterModal.subtitleEdit') : this.$t('parameterModal.subtitleAdd');
    },
    devicesList() {
      return this.$store.state.devices || [];
    },
    currentDevice() {
      const devId = Number(this.deviceId);
      return this.devicesList.find(d => Number(d.id) === devId) || this.sourceDevice || null;
    },
    targetDevice() {
      const tgtId = Number(this.targetDeviceId || this.deviceId);
      return this.devicesList.find(d => Number(d.id) === tgtId) || null;
    },
    srcProtocolName() {
      const dev = this.currentDevice;
      return dev ? (dev.communication_protocol || dev.protocol || 'UNKNOWN') : 'UNKNOWN';
    },
    destProtocolName() {
      const dev = this.targetDevice;
      return dev ? (dev.communication_protocol || dev.protocol || 'UNKNOWN') : 'UNKNOWN';
    },
    destProtocolMismatch() {
      if (!this.isCopying) return false;
      if (!this.targetDeviceId || Number(this.targetDeviceId) === Number(this.deviceId)) return false;
      return this.srcProtocolName !== this.destProtocolName;
    },
    iconCatalog() {
      return PARAMETER_ICON_CATALOG;
    },
    currentResolvedIconKey() {
      return resolveParameterIconKey(this.form.icon, this.form.parameter_code, this.form.parameter_name, this.form.unit);
    },
    previewSampleX() {
      if (this.paramToEdit && this.paramToEdit.current_value !== null && this.paramToEdit.current_value !== undefined) {
        return this.paramToEdit.current_value;
      }
      return this.form.min_value !== null && this.form.min_value !== undefined ? this.form.min_value + 10 : 25.0;
    },
    formulaPreview() {
      if (!this.form.formula || !this.form.formula.trim()) return null;
      try {
        const formStr = this.form.formula.trim();
        const x = this.previewSampleX;
        const raw = 100.0;
        const expr = formStr
          .replace(/\b(pi)\b/gi, 'Math.PI')
          .replace(/\b(e)\b/gi, 'Math.E')
          .replace(/\b(abs)\b/gi, 'Math.abs')
          .replace(/\b(round)\b/gi, 'Math.round')
          .replace(/\b(floor)\b/gi, 'Math.floor')
          .replace(/\b(ceil)\b/gi, 'Math.ceil')
          .replace(/\b(sqrt)\b/gi, 'Math.sqrt')
          .replace(/\b(min)\b/gi, 'Math.min')
          .replace(/\b(max)\b/gi, 'Math.max')
          .replace(/\b(log|ln)\b/gi, 'Math.log')
          .replace(/\b(exp)\b/gi, 'Math.exp')
          .replace(/\b(sin)\b/gi, 'Math.sin')
          .replace(/\b(cos)\b/gi, 'Math.cos')
          .replace(/\b(tan)\b/gi, 'Math.tan')
          .replace(/\^/g, '**')
          .replace(/\b(x|val|value)\b/gi, `(${x})`)
          .replace(/\b(raw)\b/gi, `(${raw})`);

        if (/[^0-9\.\+\-\*\/\(\)\,\sMathPIEabseroundfloceilsqrtminxlgexpstnc]/.test(expr)) {
          return { valid: false, error: 'Invalid syntax or unsupported characters' };
        }
        const fn = new Function(`return (${expr})`);
        const res = fn();
        if (typeof res === 'number' && !isNaN(res) && isFinite(res)) {
          return { valid: true, value: Number(res.toFixed(4)) };
        }
        return { valid: false, error: 'Result is NaN or infinite' };
      } catch (e) {
        return { valid: false, error: e.message };
      }
    },
    hardLimitWarning() {
      const min = this.cleanNum(this.form.min_value);
      const max = this.cleanNum(this.form.max_value);
      if (min !== null && max !== null && min > max) {
        return this.$t('parameterModal.minGreaterThanMax', { min, max });
      }
      return null;
    },
    warningLimitWarning() {
      const low = this.cleanNum(this.form.warning_low);
      const high = this.cleanNum(this.form.warning_high);
      if (low !== null && high !== null && low > high) {
        return this.$t('parameterModal.warningLowGreaterThanHigh', { low, high });
      }
      return null;
    },
    softBeyondHardNotice() {
      const wLow = this.cleanNum(this.form.warning_low);
      const min = this.cleanNum(this.form.min_value);
      if (wLow !== null && min !== null && wLow < min) {
        return this.$t('parameterModal.warnLowBelowMin', { low: wLow, min });
      }
      const wHigh = this.cleanNum(this.form.warning_high);
      const max = this.cleanNum(this.form.max_value);
      if (wHigh !== null && max !== null && wHigh > max) {
        return this.$t('parameterModal.warnHighAboveMax', { high: wHigh, max });
      }
      return null;
    },
    scaleZeroWarning() {
      const s = this.cleanNum(this.form.scale);
      if (s !== null && s === 0) {
        return this.$t('parameterModal.scaleZero');
      }
      return null;
    },
    clientValidationError() {
      if (!this.form.parameter_code || !this.form.parameter_code.trim()) {
        return this.$t('parameterModal.codeRequired');
      }
      if (!this.form.parameter_name || !this.form.parameter_name.trim()) {
        return this.$t('parameterModal.nameRequired');
      }
      if (this.hardLimitWarning) return this.hardLimitWarning;
      if (this.warningLimitWarning) return this.warningLimitWarning;
      if (this.scaleZeroWarning) return this.scaleZeroWarning;
      return null;
    },
    activeErrorMessage() {
      return this.error || this.clientValidationError || null;
    },
  },
  watch: {
    isOpen: {
      immediate: true,
      handler(val) {
        if (val) {
          if (!this.$store.state.devices || this.$store.state.devices.length === 0) {
            this.$store.dispatch('fetchDevices');
          }
          if (this.isCopying && this.sourceParam) {
            this.populateFromSource(this.sourceParam);
          } else if (this.paramToEdit) {
            this.populateFromParam(this.paramToEdit);
          } else {
            this.resetForm();
          }
        }
      },
    },
    paramToEdit: {
      handler(newVal) {
        if (!this.isCopying && newVal) {
          this.populateFromParam(newVal);
        } else if (!this.isCopying && !newVal) {
          this.resetForm();
        }
      },
    },
    sourceParam: {
      handler(newVal) {
        if (this.isCopying && newVal) {
          this.populateFromSource(newVal);
        }
      },
    },
  },
  methods: {
    cleanNum(val) {
      if (val === null || val === undefined || val === '') return null;
      const n = Number(val);
      return isNaN(n) ? null : n;
    },
    formatErrorMessage(rawErr) {
      if (!rawErr) return 'Operation failed';
      const str = String(rawErr);
      if (str.includes('warning_low cannot be greater than warning_high')) {
        return this.$t('parameterModal.warningLowGreaterThanHigh', {
          low: this.cleanNum(this.form.warning_low) ?? 'Min',
          high: this.cleanNum(this.form.warning_high) ?? 'Max',
        });
      }
      if (str.includes('min_value cannot be greater than max_value')) {
        return this.$t('parameterModal.minGreaterThanMax', {
          min: this.cleanNum(this.form.min_value) ?? 'Min',
          max: this.cleanNum(this.form.max_value) ?? 'Max',
        });
      }
      if (str.includes('scale factor cannot be zero')) {
        return this.$t('parameterModal.scaleZero');
      }
      if (str.includes('parameter_code cannot be empty')) {
        return this.$t('parameterModal.codeRequired');
      }
      if (str.includes('parameter_name cannot be empty')) {
        return this.$t('parameterModal.nameRequired');
      }
      return str;
    },
    scrollToError() {
      this.$nextTick(() => {
        if (this.$refs.formBody) {
          this.$refs.formBody.scrollTo({ top: 0, behavior: 'smooth' });
        }
      });
    },
    closeModal() {
      this.error = null;
      this.$emit('close');
    },
    resetForm() {
      this.form = {
        parameter_code: '',
        parameter_name: '',
        data_type: 'FLOAT32',
        unit: '',
        icon: '',
        description: '',
        register_address: 0,
        register_type: 'HOLDING_REGISTER',
        byte_order: '',
        min_value: null,
        max_value: null,
        warning_low: null,
        warning_high: null,
        quality_validation_enabled: true,
        processing_enabled: true,
        stale_timeout_seconds: 120,
        spike_detection_enabled: false,
        spike_threshold: 0.0,
        spike_window_size: 3,
        precision: 2,
        scale: 1.0,
        offset: 0.0,
        enabled: true,
        formula: '',
        hold_last_value_enabled: false,
        hold_last_value_seconds: 120,
      };
    },
    generateSuggestedCode(sourceCode) {
      if (!sourceCode) return 'param_01';
      const match = sourceCode.match(/^(.*?)(\d+)$/);
      if (match) {
        const prefix = match[1];
        const num = parseInt(match[2], 10);
        const nextNum = String(num + 1).padStart(match[2].length, '0');
        return `${prefix}${nextNum}`;
      }
      return `${sourceCode}_copy`;
    },
    generateSuggestedName(sourceName) {
      if (!sourceName) return '';
      const match = sourceName.match(/^(.*?)(\d+)$/);
      if (match) {
        const prefix = match[1];
        const num = parseInt(match[2], 10);
        return `${prefix}${num + 1}`;
      }
      return `${sourceName} (Copy)`;
    },
    populateFromParam(param) {
      this.targetDeviceId = Number(this.deviceId);
      this.form = {
        parameter_code: param.parameter_code || param.code || '',
        parameter_name: param.parameter_name || param.name || '',
        data_type: param.data_type || 'FLOAT32',
        unit: param.unit || '',
        icon: param.icon || '',
        description: param.description || '',
        register_address: param.register_address !== undefined ? param.register_address : 0,
        register_type: param.register_type || 'HOLDING_REGISTER',
        byte_order: param.byte_order || '',
        min_value: param.min_value !== undefined ? param.min_value : (param.low_limit || null),
        max_value: param.max_value !== undefined ? param.max_value : (param.high_limit || null),
        warning_low: param.warning_low !== undefined ? param.warning_low : null,
        warning_high: param.warning_high !== undefined ? param.warning_high : null,
        quality_validation_enabled: param.quality_validation_enabled !== undefined ? param.quality_validation_enabled : true,
        processing_enabled: param.processing_enabled !== undefined ? param.processing_enabled : true,
        stale_timeout_seconds: param.stale_timeout_seconds || 120,
        spike_detection_enabled: param.spike_detection_enabled !== undefined ? param.spike_detection_enabled : false,
        spike_threshold: param.spike_threshold || 0.0,
        spike_window_size: param.spike_window_size || 3,
        precision: param.precision !== undefined ? param.precision : 2,
        scale: param.scale !== undefined ? param.scale : (param.scale_factor || 1.0),
        offset: param.offset !== undefined ? param.offset : 0.0,
        enabled: param.enabled !== undefined ? param.enabled : true,
        formula: param.formula || '',
        hold_last_value_enabled: param.hold_last_value_enabled !== undefined ? param.hold_last_value_enabled : false,
        hold_last_value_seconds: param.hold_last_value_seconds || 120,
      };
    },
    populateFromSource(source) {
      this.targetDeviceId = Number(this.deviceId);
      const code = source.parameter_code || source.code || '';
      const name = source.parameter_name || source.name || '';
      this.form = {
        parameter_code: this.generateSuggestedCode(code),
        parameter_name: this.generateSuggestedName(name),
        data_type: source.data_type || 'FLOAT32',
        unit: source.unit || '',
        icon: source.icon || '',
        description: source.description || '',
        register_address: source.register_address !== undefined ? source.register_address : 0,
        register_type: source.register_type || 'HOLDING_REGISTER',
        byte_order: source.byte_order || '',
        min_value: source.min_value !== undefined ? source.min_value : (source.low_limit || null),
        max_value: source.max_value !== undefined ? source.max_value : (source.high_limit || null),
        warning_low: source.warning_low !== undefined ? source.warning_low : null,
        warning_high: source.warning_high !== undefined ? source.warning_high : null,
        quality_validation_enabled: source.quality_validation_enabled !== undefined ? source.quality_validation_enabled : true,
        processing_enabled: source.processing_enabled !== undefined ? source.processing_enabled : true,
        stale_timeout_seconds: source.stale_timeout_seconds || 120,
        spike_detection_enabled: source.spike_detection_enabled !== undefined ? source.spike_detection_enabled : false,
        spike_threshold: source.spike_threshold || 0.0,
        spike_window_size: source.spike_window_size || 3,
        precision: source.precision !== undefined ? source.precision : 2,
        scale: source.scale !== undefined ? source.scale : (source.scale_factor || 1.0),
        offset: source.offset !== undefined ? source.offset : 0.0,
        enabled: source.enabled !== undefined ? source.enabled : true,
        formula: source.formula || '',
        hold_last_value_enabled: source.hold_last_value_enabled !== undefined ? source.hold_last_value_enabled : false,
        hold_last_value_seconds: source.hold_last_value_seconds || 120,
      };
    },
    async submitForm() {
      if (this.clientValidationError) {
        this.error = this.clientValidationError;
        this.scrollToError();
        return;
      }

      this.loading = true;
      this.error = null;

      const minVal = this.cleanNum(this.form.min_value);
      const maxVal = this.cleanNum(this.form.max_value);
      const warnLow = this.cleanNum(this.form.warning_low);
      const warnHigh = this.cleanNum(this.form.warning_high);

      const payload = {
        parameter_code: this.form.parameter_code ? this.form.parameter_code.trim() : '',
        parameter_name: this.form.parameter_name ? this.form.parameter_name.trim() : '',
        data_type: this.form.data_type,
        unit: this.form.unit ? this.form.unit.trim() : '',
        icon: this.form.icon ? this.form.icon.trim() : '',
        description: this.form.description ? this.form.description.trim() : '',
        register_address: this.cleanNum(this.form.register_address) || 0,
        register_type: this.form.register_type,
        byte_order: this.form.byte_order || '',
        min_value: minVal,
        max_value: maxVal,
        warning_low: warnLow,
        warning_high: warnHigh,
        clear_min_value: minVal === null,
        clear_max_value: maxVal === null,
        clear_warning_low: warnLow === null,
        clear_warning_high: warnHigh === null,
        quality_validation_enabled: !!this.form.quality_validation_enabled,
        processing_enabled: !!this.form.processing_enabled,
        stale_timeout_seconds: this.cleanNum(this.form.stale_timeout_seconds) || 120,
        spike_detection_enabled: !!this.form.spike_detection_enabled,
        spike_threshold: this.cleanNum(this.form.spike_threshold) || 0.0,
        spike_window_size: this.cleanNum(this.form.spike_window_size) || 3,
        precision: this.cleanNum(this.form.precision) !== null ? this.cleanNum(this.form.precision) : 2,
        scale: this.cleanNum(this.form.scale) !== null ? this.cleanNum(this.form.scale) : 1.0,
        offset: this.cleanNum(this.form.offset) !== null ? this.cleanNum(this.form.offset) : 0.0,
        enabled: !!this.form.enabled,
        formula: this.form.formula ? this.form.formula.trim() : '',
        hold_last_value_enabled: !!this.form.hold_last_value_enabled,
        hold_last_value_seconds: this.cleanNum(this.form.hold_last_value_seconds) || 120,
      };

      if (this.isCopying && this.sourceParam) {
        payload.source_parameter_id = this.sourceParam.id;
      }

      const effectiveDeviceId = (this.isCopying && this.targetDeviceId) ? this.targetDeviceId : this.deviceId;

      try {
        if (this.isEditing) {
          await this.$store.dispatch('updateParameter', {
            deviceId: effectiveDeviceId,
            paramId: this.paramToEdit.id,
            updates: payload,
          });
        } else {
          await this.$store.dispatch('createParameter', {
            deviceId: effectiveDeviceId,
            paramData: payload,
          });
        }
        this.$emit('saved', {
          isCopy: this.isCopying,
          deviceId: effectiveDeviceId,
          paramCode: payload.parameter_code,
          paramName: payload.parameter_name,
        });
        this.closeModal();
      } catch (err) {
        const rawMsg = (err.response && err.response.data && err.response.data.error) || err.message || 'Operation failed';
        this.error = this.formatErrorMessage(rawMsg);
        this.scrollToError();
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
