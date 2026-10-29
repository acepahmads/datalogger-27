<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-white dark:bg-[#0c121e] border border-slate-200 dark:border-slate-800 rounded-2xl w-full max-w-4xl max-h-[92vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-3.5 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50/90 dark:bg-[#070b13] backdrop-blur-md">
        <div class="flex items-center space-x-3">
          <div class="w-9 h-9 rounded-xl flex items-center justify-center bg-blue-500/10 text-blue-600 dark:bg-cyan-500/15 dark:text-cyan-400 border border-blue-500/20 dark:border-cyan-500/30 shadow-xs">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4" />
            </svg>
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="text-sm sm:text-base font-bold text-slate-900 dark:text-white tracking-wide">
                {{ modalTitle }}
              </h2>
              <span v-if="isCopying" class="px-2 py-0.5 text-4xs font-mono font-bold uppercase rounded bg-cyan-100 text-cyan-800 border border-cyan-300 dark:bg-cyan-950/80 dark:text-cyan-300 dark:border-cyan-800/80">
                CLONE
              </span>
              <span v-if="currentDevice" class="hidden sm:inline-flex items-center space-x-1 px-2 py-0.5 rounded text-4xs font-mono font-medium bg-slate-100 text-slate-700 border border-slate-200 dark:bg-slate-800/80 dark:text-slate-300 dark:border-slate-700/60">
                <span class="text-slate-400">DEV:</span>
                <span class="font-bold text-cyan-600 dark:text-cyan-400">{{ currentDevice.device_code || currentDevice.code }}</span>
              </span>
            </div>
            <p class="text-2xs text-slate-500 dark:text-slate-400 mt-0.5">
              {{ modalSubtitle }}
            </p>
          </div>
        </div>
        <button
          type="button"
          @click="closeModal"
          class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1.5 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800/80 transition"
          aria-label="Close modal"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Modal Body -->
      <form ref="formBody" @submit.prevent="submitForm" class="flex-1 overflow-y-auto p-5 sm:p-6 space-y-4">
        <!-- Top Error Alert -->
        <div v-if="error" class="p-3 bg-rose-50 dark:bg-rose-500/15 border border-rose-300 dark:border-rose-500/40 rounded-xl text-xs text-rose-800 dark:text-rose-300 flex items-start space-x-2.5 shadow-md shadow-rose-950/10 dark:shadow-rose-950/20 animate-fade-in">
          <svg class="w-4 h-4 text-rose-500 dark:text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
          </svg>
          <div class="flex-1">
            <span class="font-bold text-rose-900 dark:text-rose-200 block text-2xs">{{ $t('parameterModal.cannotSaveTitle') }}</span>
            <span class="text-3xs text-rose-800 dark:text-rose-300 mt-0.5 block leading-relaxed whitespace-pre-wrap">{{ error }}</span>
          </div>
          <button type="button" @click="error = null" class="text-rose-500 hover:text-rose-800 dark:text-rose-400 dark:hover:text-white text-3xs font-mono ml-2">✕</button>
        </div>

        <!-- Copying Parameter Notice Banner -->
        <div v-if="isCopying" class="p-3.5 bg-cyan-50 dark:bg-cyan-950/40 border border-cyan-200 dark:border-cyan-800/80 rounded-xl space-y-2.5 animate-fade-in">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2 text-cyan-900 dark:text-cyan-200 text-xs">
              <svg class="w-4 h-4 text-cyan-600 dark:text-cyan-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
              </svg>
              <span class="font-semibold">{{ $t('parameterModal.copyBannerTitle', { code: sourceParam ? (sourceParam.parameter_name || sourceParam.parameter_code) : '' }) }}</span>
            </div>
            <span class="text-3xs uppercase tracking-wider px-2 py-0.5 rounded bg-cyan-100 text-cyan-800 font-bold border border-cyan-300 dark:bg-cyan-900/60 dark:text-cyan-300 dark:border-cyan-700/50">
              CONFIG CLONE
            </span>
          </div>
          <p class="text-3xs text-slate-600 dark:text-slate-400 leading-relaxed">
            {{ $t('parameterModal.copyBannerDesc') }}
          </p>

          <!-- Destination Device Selector -->
          <div v-if="devicesList && devicesList.length > 0" class="pt-2 border-t border-slate-200 dark:border-slate-800/80">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                  {{ $t('parameterModal.destinationDevice') }}
                </label>
                <span class="text-3xs text-slate-500 dark:text-slate-400">{{ $t('parameterModal.destinationDeviceHelp') }}</span>
              </div>
              <select
                v-model="targetDeviceId"
                class="px-2.5 py-1.5 bg-white dark:bg-[#0B0F19] border border-cyan-300 dark:border-cyan-800/80 rounded-lg text-xs text-slate-900 dark:text-white focus:outline-none focus:border-cyan-500 font-mono min-w-[200px]"
              >
                <option v-for="dev in devicesList" :key="dev.id" :value="dev.id">
                  {{ dev.device_code || dev.code }} ({{ dev.communication_protocol || dev.protocol }})
                </option>
              </select>
            </div>

            <!-- Protocol Mismatch Warning -->
            <div v-if="destProtocolMismatch" class="mt-2 p-2 bg-amber-50 dark:bg-amber-500/10 border border-amber-200 dark:border-amber-500/30 rounded-lg text-3xs text-amber-800 dark:text-amber-300 flex items-start space-x-2">
              <span class="text-amber-600 dark:text-amber-400 font-bold">⚠️</span>
              <span>{{ $t('parameterModal.destProtocolMismatch', { destProto: destProtocolName, srcProto: srcProtocolName }) }}</span>
            </div>
          </div>
        </div>

        <!-- ========================================== -->
        <!-- CARD 1: GENERAL & IDENTITY                 -->
        <!-- ========================================== -->
        <div class="p-4 bg-slate-50/70 dark:bg-[#070b13]/80 border border-slate-200 dark:border-slate-800/80 rounded-xl space-y-3.5">
          <div class="flex items-center justify-between border-b border-slate-200/80 dark:border-slate-800/60 pb-2">
            <div class="flex items-center space-x-2">
              <span class="text-blue-600 dark:text-cyan-400 text-xs font-mono font-bold">#</span>
              <h3 class="text-2xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider font-mono">
                Parameter Identity & Profile
              </h3>
            </div>
            <label class="flex items-center space-x-2 cursor-pointer select-none">
              <input
                v-model="form.enabled"
                type="checkbox"
                class="w-3.5 h-3.5 rounded text-blue-600 dark:text-cyan-500 focus:ring-0 bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-700"
              />
              <span class="text-2xs font-semibold" :class="form.enabled ? 'text-emerald-600 dark:text-emerald-400 font-bold' : 'text-slate-400'">
                {{ form.enabled ? $t('deviceDetail.active') : $t('deviceDetail.disabled') }}
              </span>
            </label>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-3.5">
            <!-- Parameter Code -->
            <div>
              <div class="flex items-center justify-between mb-1">
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                  {{ $t('parameterModal.code') }} <span class="text-rose-500 font-bold">*</span>
                </label>
                <span v-if="isCopying" class="text-4xs font-mono text-cyan-800 dark:text-cyan-400 bg-cyan-100 dark:bg-cyan-950/60 px-1.5 py-0.2 rounded border border-cyan-300 dark:border-cyan-800/60">
                  Unique Code
                </span>
              </div>
              <input
                v-model="form.parameter_code"
                type="text"
                required
                :placeholder="$t('parameterModal.placeholders.code')"
                :class="isCopying ? 'border-cyan-500/80 focus:border-cyan-400' : 'border-slate-300 dark:border-slate-700/80 focus:border-blue-500'"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none font-mono transition"
              />
            </div>

            <!-- Parameter Name -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('parameterModal.name') }} <span class="text-rose-500 font-bold">*</span>
              </label>
              <input
                v-model="form.parameter_name"
                type="text"
                required
                :placeholder="$t('parameterModal.placeholders.name')"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 transition"
              />
            </div>

            <!-- Notes / Description -->
            <div class="md:col-span-2">
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('common.description') }}
              </label>
              <input
                v-model="form.description"
                type="text"
                :placeholder="$t('parameterModal.placeholders.description')"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 transition"
              />
            </div>
          </div>
        </div>

        <!-- ========================================== -->
        <!-- CARD 2: CATEGORIES & LAYOUT PROFILE        -->
        <!-- ========================================== -->
        <div class="p-4 bg-slate-50/70 dark:bg-[#070b13]/80 border border-slate-200 dark:border-slate-800/80 rounded-xl space-y-3">
          <div class="flex flex-wrap items-center justify-between gap-2 border-b border-slate-200/80 dark:border-slate-800/60 pb-2">
            <div class="flex items-center space-x-2">
              <span class="text-purple-600 dark:text-purple-400 text-xs">🏷️</span>
              <h3 class="text-2xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider font-mono">
                {{ $t('operator.categorySettings') || 'Category & Sensor Presentation' }}
              </h3>
            </div>
            <!-- Live Preview Badge with Category -->
            <div class="flex items-center space-x-1.5 text-2xs font-mono px-2 py-0.5 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-700/80 rounded-lg shadow-2xs">
              <span class="text-slate-400 text-3xs">{{ $t('parameterModal.iconCardPreview') || 'Card Preview' }}:</span>
              <span class="text-cyan-600 dark:text-cyan-400/90 font-semibold text-3xs">{{ currentDevice ? (currentDevice.device_code || currentDevice.code) : 'DEV-01' }}</span>
              <span class="text-slate-400">·</span>
              <div class="flex items-center space-x-1 font-bold text-slate-900 dark:text-white">
                <ParameterIcon :icon="form.icon" :code="form.parameter_code" :name="form.parameter_name" :unit="form.unit" size-class="w-3 h-3" />
                <span class="truncate max-w-[120px]">{{ form.parameter_name || form.parameter_code || 'Preview' }}</span>
              </div>
              <span
                v-if="selectedCategoryObject"
                class="px-1.5 py-0.2 rounded text-4xs font-bold border ml-0.5"
                :class="getCategoryTheme(selectedCategoryObject.color_key).pill"
              >
                {{ selectedCategoryObject.name }}
              </span>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-3.5">
            <!-- Category Selector -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('categoryManagement.fields.name') || 'Domain Category' }}
              </label>
              <select
                v-model="form.category_id"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
              >
                <option :value="null">— {{ $t('operator.uncategorized') || 'Uncategorized' }} —</option>
                <option v-for="cat in availableCategories" :key="cat.id" :value="cat.id">
                  {{ cat.name }} {{ !cat.enabled ? '(Disabled)' : '' }}
                </option>
              </select>
            </div>

            <!-- Icon Selector -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('parameterModal.iconLabel') || 'Sensor Icon Type' }}
              </label>
              <select
                v-model="form.icon"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
              >
                <option value="">✨ {{ $t('parameterModal.iconAuto') || 'Auto-Detect (Based on Code/Name/Unit)' }}</option>
                <option v-for="cat in iconCatalog" :key="cat.key" :value="cat.key">
                  {{ cat.label }}
                </option>
              </select>
            </div>
          </div>

          <div class="text-3xs text-slate-500 dark:text-slate-400 flex items-center space-x-1.5 font-mono pt-0.5">
            <span>{{ $t('parameterModal.iconResolved') || 'Resolved Icon' }}:</span>
            <span class="font-bold text-cyan-600 dark:text-cyan-400 capitalize">{{ currentResolvedIconKey }}</span>
            <span v-if="!form.icon" class="text-slate-400 dark:text-slate-500 font-sans italic">(automatic heuristic)</span>
            <span v-else class="text-emerald-600 dark:text-emerald-400 font-sans italic">(manually configured)</span>
          </div>
        </div>

        <!-- ========================================== -->
        <!-- CARD 3: MODBUS COMMUNICATION & POLLING     -->
        <!-- ========================================== -->
        <div class="p-4 bg-slate-50/70 dark:bg-[#070b13]/80 border border-slate-200 dark:border-slate-800/80 rounded-xl space-y-3.5">
          <div class="flex items-center justify-between border-b border-slate-200/80 dark:border-slate-800/60 pb-2">
            <div class="flex items-center space-x-2">
              <span class="text-cyan-600 dark:text-cyan-400 text-xs">⚡</span>
              <h3 class="text-2xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider font-mono">
                Modbus Communication & Polling Configuration
              </h3>
            </div>
            <span class="px-2 py-0.5 rounded text-4xs font-mono font-bold uppercase tracking-wider bg-slate-100 text-slate-700 border border-slate-200 dark:bg-slate-800/90 dark:text-slate-300 dark:border-slate-700">
              {{ currentDevice ? (currentDevice.communication_protocol || currentDevice.protocol || 'MODBUS') : 'MODBUS' }} · SLAVE #{{ form.slave_id }}
            </span>
          </div>

          <!-- Row 1: Register Mapping (3 Columns - Spacious, No Truncation) -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3.5">
            <!-- Register Address -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('parameterModal.addressOffset') }} <span class="text-rose-500 font-bold">*</span>
              </label>
              <input
                v-model.number="form.register_address"
                type="number"
                min="0"
                placeholder="e.g. 40001, 30001"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono transition"
              />
            </div>

            <!-- Register Type -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('parameterModal.registerType') }} <span class="text-rose-500 font-bold">*</span>
              </label>
              <select
                v-model="form.register_type"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
              >
                <option value="HOLDING_REGISTER">Holding Register (4x)</option>
                <option value="INPUT_REGISTER">Input Register (3x)</option>
                <option value="COIL">Coil (0x)</option>
                <option value="DISCRETE_INPUT">Discrete Input (1x)</option>
                <option value="VARIABLE">Virtual / Internal Variable</option>
              </select>
            </div>

            <!-- Slave ID / Unit ID -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1 flex items-center justify-between">
                <span>{{ $t('parameterModal.slaveId') || 'Slave ID / Unit ID' }}</span>
                <span class="text-4xs text-slate-400 font-mono font-normal">1-247</span>
              </label>
              <input
                v-model.number="form.slave_id"
                type="number"
                min="1"
                max="247"
                placeholder="1"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono transition"
              />
            </div>
          </div>

          <!-- Row 2: Data Encoding & Framing (3 Columns) -->
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3.5">
            <!-- Data Type -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('parameterModal.dataType') }} <span class="text-rose-500 font-bold">*</span>
              </label>
              <select
                v-model="form.data_type"
                @change="handleDataTypeChange"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
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

            <!-- Byte / Word Order -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1 flex items-center justify-between">
                <span>{{ $t('parameterModal.byteOrderOverride') }}</span>
                <span class="text-4xs text-slate-500 dark:text-slate-400 font-mono font-normal">({{ byteOrderSize }}B / {{ byteOrderSize / 2 }}W)</span>
              </label>
              <select
                :value="byteOrderSelection"
                @change="onByteOrderSelectionChange($event.target.value)"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
              >
                <option value="">{{ $t('parameterModal.defaultDeviceConn') }}</option>
                <option v-for="preset in relevantPresets" :key="preset.value" :value="preset.value">
                  {{ preset.label }}
                </option>
                <option value="CUSTOM">★ {{ $t('parameterModal.customByteOrder') || 'Custom Byte Order' }}</option>
              </select>
            </div>

            <!-- Batch Polling Mode -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1 flex items-center justify-between">
                <span>{{ $t('parameterModal.batchMode') || 'Batch Mode' }}</span>
                <span class="text-4xs font-mono font-bold" :class="form.batch_mode === 'AUTO' ? 'text-indigo-600 dark:text-indigo-400' : 'text-amber-600 dark:text-amber-400'">
                  {{ form.batch_mode === 'AUTO' ? 'AUTO' : 'SINGLE' }}
                </span>
              </label>
              <select
                v-model="form.batch_mode"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white focus:outline-none focus:border-indigo-500 font-mono"
              >
                <option value="AUTO">{{ $t('parameterModal.batchModeAuto') || 'AUTO (Batch Candidate)' }}</option>
                <option value="SINGLE">{{ $t('parameterModal.batchModeSingle') || 'SINGLE (Force Single Read)' }}</option>
              </select>
            </div>
          </div>

          <!-- Row 3: POLLING INTERVAL (NEW FEATURE) & PRECISION (Spacious Layout) -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5 pt-1 border-t border-slate-200/60 dark:border-slate-800/50">
            <!-- Polling Interval (ms) -->
            <div>
              <div class="flex items-center justify-between mb-1">
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                  {{ $t('parameterModal.pollingInterval') }}
                </label>
                <!-- Status Tag: Inherited vs Custom -->
                <div class="flex items-center space-x-1.5">
                  <span
                    v-if="isPollingIntervalInherited"
                    class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-slate-100 text-slate-600 border border-slate-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700/80"
                    :title="'Default: Not set. Inheriting ' + devicePollingInterval + ' ms from parent device'"
                  >
                    ⚡ {{ $t('parameterModal.pollingInherited', { interval: devicePollingInterval }) }}
                  </span>
                  <span
                    v-else
                    class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-cyan-100 text-cyan-800 border border-cyan-300 dark:bg-cyan-950/80 dark:text-cyan-300 dark:border-cyan-800/80"
                  >
                    ⚙️ {{ $t('parameterModal.pollingCustom', { interval: form.polling_interval_ms }) }}
                  </span>
                  <button
                    v-if="!isPollingIntervalInherited"
                    type="button"
                    @click="form.polling_interval_ms = null"
                    class="text-4xs text-blue-600 dark:text-cyan-400 hover:underline font-mono"
                    :title="$t('parameterModal.resetToDeviceInterval')"
                  >
                    {{ $t('parameterModal.resetToDeviceInterval') || 'Reset' }}
                  </button>
                </div>
              </div>
              <input
                v-model.number="form.polling_interval_ms"
                type="number"
                min="200"
                step="100"
                :placeholder="$t('parameterModal.pollingIntervalPlaceholder', { interval: devicePollingInterval })"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-cyan-500 font-mono transition"
              />
              <span class="text-3xs text-slate-500 dark:text-slate-400 mt-1 block leading-relaxed">
                {{ $t('parameterModal.pollingIntervalHelp', { interval: devicePollingInterval }) }}
              </span>
            </div>

            <!-- Decimal Precision -->
            <div>
              <div class="flex items-center justify-between mb-1">
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                  {{ $t('parameterModal.precision') }}
                </label>
                <span class="text-4xs text-slate-400 font-mono font-normal">0 - 6 digits</span>
              </div>
              <input
                v-model.number="form.precision"
                type="number"
                min="0"
                max="6"
                placeholder="2"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono transition"
              />
              <span class="text-3xs text-slate-500 dark:text-slate-400 mt-1 block leading-relaxed">
                Controls decimal display precision across operator cards and live telemetry tables.
              </span>
            </div>
          </div>
        </div>

        <!-- Custom Byte Order Mapping Editor & Preview (Conditional) -->
        <div v-if="isCustomByteOrder" class="p-4 rounded-xl border border-blue-200 dark:border-blue-500/30 bg-blue-50/70 dark:bg-blue-950/20 backdrop-blur-sm space-y-3.5 transition">
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2.5">
            <div class="flex items-center space-x-2 flex-wrap gap-y-1">
              <span class="w-2 h-2 rounded-full bg-blue-500 animate-pulse"></span>
              <span class="text-xs font-bold text-slate-900 dark:text-white font-mono uppercase tracking-wider">
                {{ $t('parameterModal.byteOrderEditor') || 'Custom Byte Order Mapping' }}
              </span>
              <span class="text-3xs font-mono font-bold text-blue-700 dark:text-blue-300 bg-blue-100 dark:bg-blue-500/20 px-2 py-0.5 rounded border border-blue-200 dark:border-blue-500/30">
                {{ form.data_type }} ({{ byteOrderSize }} Bytes)
              </span>
            </div>
            <!-- Quick Preset Action Buttons -->
            <div class="flex items-center space-x-1.5 flex-wrap gap-y-1">
              <span class="text-3xs text-slate-500 dark:text-slate-400 font-mono mr-0.5">{{ $t('parameterModal.quickPresets') || 'Presets' }}:</span>
              <button
                type="button"
                @click="applyCustomQuickPreset('DEFAULT')"
                class="px-2 py-0.5 text-3xs rounded-md font-mono font-semibold transition border shadow-2xs bg-white hover:bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 dark:border-slate-700"
                title="Direct Standard Big Endian order"
              >
                Standard
              </button>
              <button
                v-if="byteOrderSize >= 4"
                type="button"
                @click="applyCustomQuickPreset('WORD_SWAP')"
                class="px-2 py-0.5 text-3xs rounded-md font-mono font-semibold transition border shadow-2xs bg-white hover:bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 dark:border-slate-700"
                title="Swap 16-bit register word pairs"
              >
                Word Swap
              </button>
              <button
                type="button"
                @click="applyCustomQuickPreset('BYTE_SWAP')"
                class="px-2 py-0.5 text-3xs rounded-md font-mono font-semibold transition border shadow-2xs bg-white hover:bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 dark:border-slate-700"
                title="Swap high/low byte within registers"
              >
                Byte Swap
              </button>
              <button
                type="button"
                @click="applyCustomQuickPreset('REVERSE')"
                class="px-2 py-0.5 text-3xs rounded-md font-mono font-semibold transition border shadow-2xs bg-white hover:bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-200 dark:border-slate-700"
                title="Full reverse little endian order"
              >
                Reverse
              </button>
            </div>
          </div>

          <!-- Position Slot Selectors -->
          <div class="grid gap-2" :class="byteOrderSize === 8 ? 'grid-cols-4 sm:grid-cols-8' : byteOrderSize === 4 ? 'grid-cols-2 sm:grid-cols-4' : 'grid-cols-2'">
            <div
              v-for="(slotVal, idx) in customByteSlots"
              :key="'slot-' + idx"
              class="bg-white dark:bg-[#0B0F19]/90 p-2 rounded-xl border transition shadow-xs"
              :class="isSlotDuplicate(slotVal, idx) ? 'border-rose-400 bg-rose-50/50 dark:border-rose-500/60 dark:bg-rose-950/20' : 'border-slate-200 dark:border-slate-800'"
            >
              <div class="text-4xs uppercase tracking-wider text-slate-500 dark:text-slate-400 font-mono mb-1.5 flex items-center justify-between">
                <span class="font-bold">SLOT {{ idx + 1 }}</span>
                <span class="text-slate-400 dark:text-slate-500 font-mono text-4xs">Out[{{ idx }}]</span>
              </div>
              <div class="relative">
                <select
                  :value="slotVal"
                  @change="onCustomSlotChange(idx, $event.target.value)"
                  class="w-full text-center font-mono font-bold text-xs py-1.5 px-1 rounded-lg border appearance-none cursor-pointer focus:outline-none transition shadow-2xs"
                  :class="isSlotDuplicate(slotVal, idx) ? 'border-rose-400 bg-rose-100/60 text-rose-800 focus:border-rose-500 dark:border-rose-500 dark:bg-rose-950/40 dark:text-rose-300' : 'border-slate-300 bg-slate-50 hover:bg-slate-100 text-slate-900 focus:border-blue-500 dark:border-slate-700 dark:bg-[#101826] dark:hover:bg-[#152033] dark:text-white'"
                  :title="`Slot ${idx + 1} takes Wire Byte ${slotVal}`"
                >
                  <option v-for="b in availableInputBytes" :key="b" :value="b">
                    Byte {{ b }} (Wire {{ b.charCodeAt(0) - 65 }})
                  </option>
                </select>
                <div class="pointer-events-none absolute inset-y-0 right-1 flex items-center text-slate-400 dark:text-slate-500 text-3xs">
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
                  </svg>
                </div>
              </div>
            </div>
          </div>

          <!-- Interactive Input -> Output Preview -->
          <div class="p-3 rounded-xl bg-white dark:bg-[#090D16] border border-slate-200 dark:border-slate-800/80 shadow-xs space-y-2 text-2xs font-mono">
            <div class="flex flex-col md:flex-row md:items-center justify-between gap-2.5">
              <div class="flex items-center space-x-2 flex-wrap gap-y-1">
                <span class="text-slate-500 dark:text-slate-400 uppercase text-3xs font-bold tracking-wider whitespace-nowrap">
                  {{ $t('parameterModal.inputOrder') || 'Wire Input Order' }}:
                </span>
                <div class="flex items-center space-x-1 flex-wrap gap-y-1">
                  <span
                    v-for="b in availableInputBytes"
                    :key="'in-' + b"
                    class="px-2 py-0.5 rounded-md font-bold font-mono text-3xs border shadow-2xs"
                    :class="getByteBadgeClass(b)"
                    :title="`Wire Index ${b.charCodeAt(0) - 65}: Byte ${b}`"
                  >
                    {{ b }}
                  </span>
                </div>
              </div>

              <div class="hidden md:flex text-blue-600 dark:text-cyan-400 font-bold text-sm items-center px-1">
                ➔
              </div>

              <div class="flex items-center space-x-2 flex-wrap gap-y-1">
                <span class="text-slate-500 dark:text-slate-400 uppercase text-3xs font-bold tracking-wider whitespace-nowrap">
                  {{ $t('parameterModal.outputOrder') || 'Decoded Output Order' }}:
                </span>
                <div class="flex items-center space-x-1 flex-wrap gap-y-1">
                  <span
                    v-for="(b, idx) in customByteSlots"
                    :key="'out-' + idx"
                    class="px-2 py-0.5 rounded-md font-bold font-mono text-3xs border shadow-2xs"
                    :class="getByteBadgeClass(b)"
                    :title="`Decoded Byte ${idx}: Takes Wire ${b}`"
                  >
                    {{ b }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <div v-if="byteOrderValidationError" class="flex items-center space-x-2 text-3xs text-rose-600 dark:text-rose-400 font-mono font-medium">
            <span>⚠️</span>
            <span>{{ byteOrderValidationError }}</span>
          </div>
          <div v-else class="flex items-center space-x-2 text-3xs text-emerald-600 dark:text-emerald-400 font-mono font-medium">
            <span>✓</span>
            <span>{{ $t('parameterModal.validPermutation') || 'Valid Permutation' }}: <strong class="font-bold tracking-wider">{{ customByteOrderString }}</strong></span>
          </div>
        </div>

        <div v-if="unknownByteOrderWarning" class="p-2.5 rounded-lg border border-amber-300 dark:border-amber-500/40 bg-amber-50 dark:bg-amber-500/10 text-3xs text-amber-800 dark:text-amber-300 font-mono flex items-center space-x-2">
          <span class="text-amber-600 dark:text-amber-400 font-bold">⚠️</span>
          <span>{{ unknownByteOrderWarning }}</span>
        </div>

        <!-- ========================================== -->
        <!-- CARD 4: SCALING & MEASUREMENT BOUNDS       -->
        <!-- ========================================== -->
        <div class="p-4 bg-slate-50/70 dark:bg-[#070b13]/80 border border-slate-200 dark:border-slate-800/80 rounded-xl space-y-3.5">
          <div class="flex items-center justify-between border-b border-slate-200/80 dark:border-slate-800/60 pb-2">
            <div class="flex items-center space-x-2">
              <span class="text-blue-600 dark:text-blue-400 text-xs">📐</span>
              <h3 class="text-2xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider font-mono">
                Scaling & Measurement Bounds
              </h3>
            </div>
            <span class="text-4xs text-slate-400 font-mono">y = (raw × Scale) + Offset</span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-3 gap-3.5">
            <!-- Engineering Unit -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('parameterModal.unit') }}
              </label>
              <input
                v-model="form.unit"
                type="text"
                :placeholder="$t('parameterModal.placeholders.unit')"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono transition"
              />
            </div>

            <!-- Scale Factor (Multiplier) -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('parameterModal.scaleFactor') }}
              </label>
              <input
                v-model.number="form.scale"
                type="number"
                step="0.0001"
                placeholder="1.0"
                :class="scaleZeroWarning ? 'border-rose-400 focus:border-rose-500 dark:border-rose-500/80 dark:focus:border-rose-400' : 'border-slate-300 dark:border-slate-700/80 focus:border-blue-500'"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none font-mono transition"
              />
              <span v-if="scaleZeroWarning" class="text-4xs text-rose-600 dark:text-rose-400 mt-1 block">⚠️ {{ scaleZeroWarning }}</span>
            </div>

            <!-- Offset (Adder) -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                {{ $t('parameterModal.offsetAdder') }}
              </label>
              <input
                v-model.number="form.offset"
                type="number"
                step="0.01"
                placeholder="0.0"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono transition"
              />
            </div>
          </div>

          <!-- Hard Bounds: Min & Max Value Limit -->
          <div class="pt-1 border-t border-slate-200/60 dark:border-slate-800/50 space-y-2">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                  {{ $t('parameterModal.minLimit') }}
                </label>
                <input
                  v-model.number="form.min_value"
                  type="number"
                  step="any"
                  placeholder="e.g. 0.0"
                  :class="hardLimitWarning ? 'border-rose-400 focus:border-rose-500 dark:border-rose-500/80' : 'border-slate-300 dark:border-slate-700/80 focus:border-blue-500'"
                  class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none font-mono transition"
                />
              </div>

              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                  {{ $t('parameterModal.maxLimit') }}
                </label>
                <input
                  v-model.number="form.max_value"
                  type="number"
                  step="any"
                  placeholder="e.g. 500.0"
                  :class="hardLimitWarning ? 'border-rose-400 focus:border-rose-500 dark:border-rose-500/80' : 'border-slate-300 dark:border-slate-700/80 focus:border-blue-500'"
                  class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none font-mono transition"
                />
              </div>
            </div>

            <!-- Hard Limit Validation Warning -->
            <div v-if="hardLimitWarning" class="p-2.5 rounded-lg bg-rose-50 dark:bg-rose-500/15 border border-rose-200 dark:border-rose-500/40 text-2xs text-rose-700 dark:text-rose-300 flex items-start space-x-2 animate-fade-in">
              <svg class="w-4 h-4 text-rose-500 dark:text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
              </svg>
              <span class="leading-relaxed">{{ hardLimitWarning }}</span>
            </div>
          </div>
        </div>

        <!-- ========================================== -->
        <!-- CARD 5: ALARM THRESHOLDS & QUALITY LIMITS  -->
        <!-- ========================================== -->
        <div class="p-4 bg-slate-50/70 dark:bg-[#070b13]/80 border border-slate-200 dark:border-slate-800/80 rounded-xl space-y-3.5">
          <div class="flex items-center justify-between border-b border-slate-200/80 dark:border-slate-800/60 pb-2">
            <div class="flex items-center space-x-2">
              <span class="text-amber-500 text-xs">🔔</span>
              <h3 class="text-2xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider font-mono">
                Alarm Thresholds & Quality Limits
              </h3>
            </div>
            <span class="px-2 py-0.5 rounded text-4xs font-mono font-bold uppercase tracking-wider bg-amber-500/10 text-amber-700 dark:text-amber-400 border border-amber-500/20">
              Soft Limits · UNCERTAIN
            </span>
          </div>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
            <div>
              <label class="block text-2xs font-semibold text-amber-700 dark:text-amber-400/90 uppercase tracking-wider mb-1 flex items-center justify-between">
                <span>{{ $t('parameterModal.warningLow') }}</span>
                <span class="text-4xs text-amber-700 dark:text-amber-500/80 font-mono font-semibold">UNCERTAIN LOW</span>
              </label>
              <input
                v-model.number="form.warning_low"
                type="number"
                step="any"
                :placeholder="$t('parameterModal.placeholders.warningLow')"
                :class="warningLimitWarning ? 'border-rose-400 focus:border-rose-500' : 'border-amber-300 dark:border-amber-500/30 focus:border-amber-500'"
                class="w-full px-3 py-1.5 bg-amber-50/20 dark:bg-[#0B0F19] border rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none font-mono transition"
              />
            </div>

            <div>
              <label class="block text-2xs font-semibold text-amber-700 dark:text-amber-400/90 uppercase tracking-wider mb-1 flex items-center justify-between">
                <span>{{ $t('parameterModal.warningHigh') }}</span>
                <span class="text-4xs text-amber-700 dark:text-amber-500/80 font-mono font-semibold">UNCERTAIN HIGH</span>
              </label>
              <input
                v-model.number="form.warning_high"
                type="number"
                step="any"
                :placeholder="$t('parameterModal.placeholders.warningHigh')"
                :class="warningLimitWarning ? 'border-rose-400 focus:border-rose-500' : 'border-amber-300 dark:border-amber-500/30 focus:border-amber-500'"
                class="w-full px-3 py-1.5 bg-amber-50/20 dark:bg-[#0B0F19] border rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none font-mono transition"
              />
            </div>
          </div>

          <!-- Warning Limits Alerts -->
          <div v-if="warningLimitWarning" class="p-2.5 rounded-lg bg-rose-50 dark:bg-rose-500/15 border border-rose-200 dark:border-rose-500/40 text-2xs text-rose-700 dark:text-rose-300 flex items-start space-x-2 animate-fade-in">
            <svg class="w-4 h-4 text-rose-500 dark:text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
            </svg>
            <span class="leading-relaxed">{{ warningLimitWarning }}</span>
          </div>
          <div v-else-if="softBeyondHardNotice" class="p-2.5 rounded-lg bg-amber-50 dark:bg-amber-500/10 border border-amber-200 dark:border-amber-500/30 text-2xs text-amber-800 dark:text-amber-300 flex items-start space-x-2 animate-fade-in">
            <svg class="w-4 h-4 text-amber-600 dark:text-amber-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
            <span class="leading-relaxed">{{ softBeyondHardNotice }}</span>
          </div>

          <!-- Zero-Value Alarm Policy Configuration -->
          <div class="pt-2 border-t border-slate-200/60 dark:border-slate-800/50 space-y-2">
            <div class="flex items-center justify-between">
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                {{ $t('parameterModal.zeroValueAlarmPolicy') }}
              </label>
              <span
                class="text-4xs font-mono font-bold px-2 py-0.5 rounded shadow-2xs"
                :class="form.zero_value_alarm_policy === 'SUPPRESS_THRESHOLD' ? 'bg-amber-100 text-amber-800 border border-amber-300 dark:bg-amber-950/60 dark:text-amber-300 dark:border-amber-800/60' : (form.zero_value_alarm_policy === 'INVALID_READING' ? 'bg-rose-100 text-rose-800 border border-rose-300 dark:bg-rose-950/60 dark:text-rose-300 dark:border-rose-800/60' : 'bg-slate-100 text-slate-700 border border-slate-300 dark:bg-slate-800 dark:text-slate-300 dark:border-slate-700')"
              >
                {{ form.zero_value_alarm_policy || 'NORMAL' }}
              </span>
            </div>
            <select
              v-model="form.zero_value_alarm_policy"
              class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="NORMAL">{{ $t('parameterModal.zeroPolicyNormal') }}</option>
              <option value="SUPPRESS_THRESHOLD">{{ $t('parameterModal.zeroPolicySuppress') }}</option>
              <option value="INVALID_READING">{{ $t('parameterModal.zeroPolicyInvalid') }}</option>
            </select>
            <p class="text-3xs leading-relaxed">
              <span v-if="form.zero_value_alarm_policy === 'SUPPRESS_THRESHOLD'" class="text-amber-700 dark:text-amber-300 font-medium">
                ℹ️ {{ $t('parameterModal.zeroPolicySuppressHelp') }}
              </span>
              <span v-else-if="form.zero_value_alarm_policy === 'INVALID_READING'" class="text-rose-700 dark:text-rose-300 font-medium">
                ⚠️ {{ $t('parameterModal.zeroPolicyInvalidHelp') }}
              </span>
              <span v-else class="text-slate-500 dark:text-slate-400">
                {{ $t('parameterModal.zeroPolicyNormalHelp') }}
              </span>
            </p>
          </div>
        </div>

        <!-- ========================================== -->
        <!-- CARD 6: ADVANCED SIGNAL CONDITIONING       -->
        <!-- ========================================== -->
        <div class="border border-slate-200 dark:border-slate-800/80 rounded-xl overflow-hidden bg-slate-50/70 dark:bg-[#070b13]/80 transition">
          <button
            type="button"
            @click="showAdvancedSettings = !showAdvancedSettings"
            class="w-full p-4 flex items-center justify-between hover:bg-slate-100/60 dark:hover:bg-slate-800/40 transition text-left"
          >
            <div class="flex items-center space-x-2">
              <span class="text-indigo-600 dark:text-indigo-400 text-xs font-mono font-bold">ƒ(x)</span>
              <div>
                <h3 class="text-2xs font-bold text-slate-800 dark:text-slate-200 uppercase tracking-wider font-mono">
                  Advanced: Formulas, Hold Protection & Spike Filter
                </h3>
                <span class="text-3xs text-slate-400 dark:text-slate-500 block">
                  Optional signal conditioning, Python-style math evaluation, anomaly suppression, and spike window filter.
                </span>
              </div>
            </div>
            <div class="flex items-center space-x-2">
              <span v-if="form.formula" class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-blue-100 text-blue-800 dark:bg-blue-950/80 dark:text-blue-300">FORMULA</span>
              <span v-if="form.hold_last_value_enabled" class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold bg-amber-100 text-amber-800 dark:bg-amber-950/80 dark:text-amber-300">HOLD</span>
              <span class="text-slate-400 text-xs transform transition" :class="showAdvancedSettings ? 'rotate-180' : ''">▼</span>
            </div>
          </button>

          <div v-show="showAdvancedSettings" class="p-4 pt-1 border-t border-slate-200/80 dark:border-slate-800/60 space-y-4">
            <!-- Mathematical Formula Evaluation -->
            <div class="space-y-2">
              <div class="flex items-center justify-between">
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                  {{ $t('parameterModal.formula') }}
                </label>
                <span class="text-3xs text-blue-600 dark:text-cyan-400 font-mono font-bold">eval(x, raw)</span>
              </div>
              <input
                v-model="form.formula"
                type="text"
                :placeholder="$t('parameterModal.formulaPlaceholder')"
                class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono transition"
              />
              <p class="text-3xs text-slate-500 dark:text-slate-400 leading-relaxed">
                {{ $t('parameterModal.formulaDesc') }}
              </p>

              <!-- Quick Presets -->
              <div class="flex items-center space-x-1.5 flex-wrap gap-y-1.5 pt-0.5">
                <span class="text-3xs text-slate-500 dark:text-slate-400 mr-1 font-medium">{{ $t('parameterModal.presets') }}:</span>
                <button
                  type="button"
                  @click="form.formula = 'x * 1.8 + 32'"
                  class="px-2 py-0.5 rounded-md text-3xs font-mono font-medium transition border shadow-2xs bg-white hover:bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-700"
                >
                  °C➔°F (x*1.8+32)
                </button>
                <button
                  type="button"
                  @click="form.formula = '(raw - 4.0) * (100.0 / 16.0)'"
                  class="px-2 py-0.5 rounded-md text-3xs font-mono font-medium transition border shadow-2xs bg-white hover:bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-700"
                >
                  4-20mA➔%
                </button>
                <button
                  type="button"
                  @click="form.formula = 'round(x, 2)'"
                  class="px-2 py-0.5 rounded-md text-3xs font-mono font-medium transition border shadow-2xs bg-white hover:bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-700"
                >
                  round(x, 2)
                </button>
                <button
                  type="button"
                  @click="form.formula = 'x / 1000'"
                  class="px-2 py-0.5 rounded-md text-3xs font-mono font-medium transition border shadow-2xs bg-white hover:bg-slate-100 text-slate-700 border-slate-300 dark:bg-slate-800 dark:hover:bg-slate-700 dark:text-slate-300 dark:border-slate-700"
                >
                  W➔kW (x/1000)
                </button>
                <button
                  type="button"
                  @click="form.formula = 'value + random(0.01, 0.09)'"
                  class="px-2 py-0.5 rounded-md text-3xs font-mono font-bold transition border shadow-2xs bg-blue-50 hover:bg-blue-100 text-blue-700 border-blue-300 dark:bg-blue-900/40 dark:hover:bg-blue-800/60 dark:border-blue-500/40 dark:text-blue-300"
                  title="Add random jitter simulation [min, max]"
                >
                  +random(0.01, 0.09)
                </button>
              </div>

              <!-- Live Preview Banner -->
              <div v-if="form.formula && form.formula.trim()" class="p-2.5 rounded-lg bg-white dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 flex items-center justify-between text-2xs font-mono shadow-xs">
                <span class="text-slate-500 dark:text-slate-400">{{ $t('parameterModal.formulaPreview') }} (x={{ previewSampleX }}):</span>
                <span v-if="formulaPreview && formulaPreview.valid" class="font-bold text-emerald-600 dark:text-emerald-400">
                  ➔ {{ formulaPreview.value }} {{ form.unit }}
                </span>
                <span v-else class="text-rose-600 dark:text-rose-400 font-semibold">
                  ⚠️ {{ (formulaPreview && formulaPreview.error) || 'Invalid expression' }}
                </span>
              </div>
            </div>

            <!-- Hold Last Good Value -->
            <div class="pt-3 border-t border-slate-200/60 dark:border-slate-800/50 space-y-2.5">
              <label class="flex items-center space-x-3 cursor-pointer">
                <input
                  v-model="form.hold_last_value_enabled"
                  type="checkbox"
                  class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-700"
                />
                <div>
                  <span class="text-xs font-semibold text-slate-900 dark:text-slate-200">{{ $t('parameterModal.holdLastValueTitle') }}</span>
                  <span class="block text-3xs text-slate-500 dark:text-slate-400 mt-0.5">{{ $t('parameterModal.holdLastValueDesc') }}</span>
                </div>
              </label>

              <div v-if="form.hold_last_value_enabled" class="pl-7 pt-1 space-y-1.5">
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
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
                    class="w-32 px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-lg text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
                  />
                  <span class="text-3xs text-slate-500 dark:text-slate-400 font-mono">
                    ({{ Math.round((form.hold_last_value_seconds || 120) / 60 * 10) / 10 }} min)
                  </span>
                </div>
                <p class="text-3xs text-amber-700 dark:text-amber-400/90 leading-relaxed font-medium">
                  {{ $t('parameterModal.holdDurationHelp') }}
                </p>
              </div>
            </div>

            <!-- Quality & Spike Detection Settings (Phase 3.2) -->
            <div class="pt-3 border-t border-slate-200/60 dark:border-slate-800/50 space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-2xs font-bold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                  {{ $t('parameterModal.qualitySettingsTitle') }}
                </span>
                <span class="text-3xs text-blue-600 dark:text-cyan-400 font-mono font-bold">Phase 3.2</span>
              </div>

              <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <label class="flex items-center space-x-3 cursor-pointer p-2.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-[#0B0F19]">
                  <input
                    v-model="form.quality_validation_enabled"
                    type="checkbox"
                    class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-700"
                  />
                  <div>
                    <span class="text-xs font-semibold text-slate-900 dark:text-slate-200">{{ $t('parameterModal.qualityValidationEnabled') }}</span>
                    <span class="block text-4xs text-slate-500 dark:text-slate-400 mt-0.5">Validate hard and soft threshold limits</span>
                  </div>
                </label>

                <label class="flex items-center space-x-3 cursor-pointer p-2.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-[#0B0F19]">
                  <input
                    v-model="form.spike_detection_enabled"
                    type="checkbox"
                    class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-700"
                  />
                  <div>
                    <span class="text-xs font-semibold text-slate-900 dark:text-slate-200">{{ $t('parameterModal.spikeDetectionEnabled') }}</span>
                    <span class="block text-4xs text-slate-500 dark:text-slate-400 mt-0.5">Filter transient abnormal telemetry spikes</span>
                  </div>
                </label>
              </div>

              <div v-if="form.spike_detection_enabled" class="grid grid-cols-1 sm:grid-cols-2 gap-3 p-3 bg-white dark:bg-[#0B0F19] rounded-lg border border-slate-200 dark:border-slate-800">
                <div>
                  <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                    {{ $t('parameterModal.spikeThreshold') }}
                  </label>
                  <input
                    v-model.number="form.spike_threshold"
                    type="number"
                    step="any"
                    placeholder="e.g. 20.0"
                    class="w-full px-3 py-1.5 bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
                  />
                </div>
                <div>
                  <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                    {{ $t('parameterModal.spikeWindowSize') }}
                  </label>
                  <input
                    v-model.number="form.spike_window_size"
                    type="number"
                    min="2"
                    max="10"
                    placeholder="3"
                    class="w-full px-3 py-1.5 bg-slate-50 dark:bg-slate-900 border border-slate-300 dark:border-slate-700 rounded-lg text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Sticky Error Alert Above Save -->
        <div v-if="activeErrorMessage" class="p-3 bg-rose-50 dark:bg-rose-500/15 border-2 border-rose-300 dark:border-rose-500/50 rounded-xl text-xs text-rose-800 dark:text-rose-200 flex items-start space-x-2.5 shadow-md shadow-rose-950/10 dark:shadow-rose-950/40 animate-fade-in">
          <svg class="w-5 h-5 text-rose-500 dark:text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
          </svg>
          <div class="flex-1">
            <div class="font-bold flex items-center justify-between text-rose-900 dark:text-rose-100">
              <span class="text-xs">{{ $t('parameterModal.cannotSaveTitle') }}</span>
              <button v-if="error" type="button" @click="error = null" class="text-rose-600 hover:text-rose-800 dark:text-rose-400 dark:hover:text-white text-3xs underline ml-2">{{ $t('parameterModal.closeWarning') }}</button>
            </div>
            <p class="mt-1 text-2xs text-rose-700 dark:text-rose-300 leading-relaxed font-sans whitespace-pre-wrap">{{ activeErrorMessage }}</p>
          </div>
        </div>

        <!-- Modal Footer Actions -->
        <div class="pt-4 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between bg-white dark:bg-[#0c121e]">
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-2 rounded-xl text-xs font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800/80 transition"
          >
            {{ $t('common.cancel') }}
          </button>
          <button
            type="submit"
            :disabled="loading || !!clientValidationError"
            :class="clientValidationError ? 'bg-slate-200 dark:bg-slate-800 text-slate-400 dark:text-slate-500 cursor-not-allowed opacity-75' : 'bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-900/30 dark:shadow-cyan-950/40'"
            class="px-5 py-2.5 rounded-xl text-xs font-bold transition flex items-center space-x-1.5 shadow-md"
            :title="clientValidationError || ''"
          >
            <span v-if="loading" class="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
            <span v-if="clientValidationError" class="text-2xs text-amber-600 dark:text-amber-300 font-semibold">⚠️ {{ $t('parameterModal.fixLimitsBeforeSave') }}</span>
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
import { getCategoryColor } from '../utils/categoryTheme.js';
import apiClient from '../services/api.js';

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
    device: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      loading: false,
      error: null,
      targetDeviceId: null,
      availableCategories: [],
      byteOrderSelection: '',
      customByteSlots: ['A', 'B', 'C', 'D'],
      unknownByteOrderWarning: null,
      showAdvancedSettings: false,
      form: {
        category_id: null,
        parameter_code: '',
        parameter_name: '',
        data_type: 'FLOAT32',
        unit: '',
        icon: '',
        description: '',
        slave_id: 1,
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
        polling_interval_ms: null,
        enabled: true,
        formula: '',
        hold_last_value_enabled: false,
        hold_last_value_seconds: 120,
        zero_value_alarm_policy: 'NORMAL',
        batch_mode: 'AUTO',
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
      return this.devicesList.find(d => Number(d.id) === devId) || this.sourceDevice || this.device || null;
    },
    targetDevice() {
      const tgtId = Number(this.targetDeviceId || this.deviceId);
      return this.devicesList.find(d => Number(d.id) === tgtId) || null;
    },
    devicePollingInterval() {
      const dev = this.currentDevice;
      if (!dev) return 1000;
      if (dev.connection && (dev.connection.polling_interval || dev.connection.poll_interval_ms)) {
        return dev.connection.polling_interval || dev.connection.poll_interval_ms;
      }
      if (dev.polling_interval || dev.poll_interval_ms) {
        return dev.polling_interval || dev.poll_interval_ms;
      }
      return 1000;
    },
    isPollingIntervalInherited() {
      const v = this.form.polling_interval_ms;
      return v === null || v === undefined || v === '' || Number(v) <= 0;
    },
    effectivePollingInterval() {
      if (!this.isPollingIntervalInherited) {
        return Number(this.form.polling_interval_ms);
      }
      return this.devicePollingInterval;
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
    selectedCategoryObject() {
      if (!this.form.category_id) return null;
      return this.availableCategories.find(c => Number(c.id) === Number(this.form.category_id)) || null;
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
        const xVal = this.previewSampleX;
        const rawVal = 100.0;

        // Basic character safety check: reject dangerous characters
        if (/[;{}[\]"'`\\$]|(?<!=)=(?!=)|=>/.test(formStr)) {
          return { valid: false, error: 'Invalid syntax or disallowed characters' };
        }

        // Allowed identifiers: variables, constants, and math functions
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

        // Helpers matching Go backend formula evaluator
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

        // Replace ^ with ** for exponents
        let expr = formStr.replace(/\^/g, '**');

        // Lowercase identifiers for case-insensitivity
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
      if (this.hardLimitWarning) {
        return this.hardLimitWarning;
      }
      if (this.warningLimitWarning) {
        return this.warningLimitWarning;
      }
      if (this.scaleZeroWarning) {
        return this.scaleZeroWarning;
      }
      if (this.byteOrderValidationError) {
        return this.byteOrderValidationError;
      }
      return null;
    },
    activeErrorMessage() {
      return this.error || this.clientValidationError || null;
    },
    byteOrderSize() {
      const dt = this.form.data_type;
      if (dt === 'INT16' || dt === 'UINT16') return 2;
      if (dt === 'FLOAT64') return 8;
      if (dt === 'BOOLEAN' || dt === 'STRING') return 2;
      return 4; // FLOAT32, INT32, UINT32
    },
    availableInputBytes() {
      return ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'].slice(0, this.byteOrderSize);
    },
    relevantPresets() {
      const size = this.byteOrderSize;
      if (size === 2) {
        return [
          { label: 'AB — Big Endian (Standard)', value: 'AB' },
          { label: 'BA — Little Endian (Byte Swap)', value: 'BA' },
        ];
      }
      if (size === 4) {
        return [
          { label: 'ABCD — Big Endian (Motorola)', value: 'ABCD' },
          { label: 'CDAB — Word Swap (Modicon)', value: 'CDAB' },
          { label: 'BADC — Byte Swap (Mid-Little)', value: 'BADC' },
          { label: 'DCBA — Little Endian (Intel)', value: 'DCBA' },
        ];
      }
      if (size === 8) {
        return [
          { label: 'ABCDEFGH — Big Endian (Standard 64-bit)', value: 'ABCDEFGH' },
          { label: 'GHEFCDAB — Word Swap (Reverse 16-bit Words)', value: 'GHEFCDAB' },
          { label: 'HGFEDCBA — Little Endian (Full Reverse)', value: 'HGFEDCBA' },
        ];
      }
      return [];
    },
    isCustomByteOrder() {
      return this.byteOrderSelection === 'CUSTOM';
    },
    customByteOrderString() {
      if (!this.customByteSlots) return '';
      return this.customByteSlots.join('');
    },
    byteOrderValidationError() {
      if (this.byteOrderSelection !== 'CUSTOM') return null;
      const size = this.byteOrderSize;
      if (!this.customByteSlots || this.customByteSlots.length !== size) {
        return this.$t('parameterModal.invalidByteOrderLength', { count: size }) || `Custom byte order requires ${size} positions`;
      }
      const seen = new Set();
      const duplicates = [];
      for (const b of this.customByteSlots) {
        if (!b) return 'All byte positions must be assigned';
        if (seen.has(b)) duplicates.push(b);
        seen.add(b);
      }
      const expected = this.availableInputBytes;
      const missing = expected.filter(b => !seen.has(b));
      if (duplicates.length > 0) {
        return this.$t('parameterModal.duplicateByteError', { bytes: [...new Set(duplicates)].join(', ') }) || `Duplicate byte position: ${[...new Set(duplicates)].join(', ')}`;
      }
      if (missing.length > 0) {
        return this.$t('parameterModal.missingByteError', { bytes: missing.join(', ') }) || `Missing byte position: ${missing.join(', ')}`;
      }
      return null;
    },
  },
  watch: {
    isOpen: {
      immediate: true,
      handler(val) {
        if (val) {
          this.fetchCategories();
          if (this.isCopying && this.sourceParam) {
            this.populateFromSource(this.sourceParam);
          } else if (this.paramToEdit) {
            this.populateFromParam(this.paramToEdit);
          } else {
            this.resetForm();
          }
        } else {
          this.error = null;
        }
      },
    },
    paramToEdit(newVal) {
      if (this.isOpen && newVal && !this.isCopying) {
        this.populateFromParam(newVal);
      }
    },
    sourceParam(newVal) {
      if (this.isOpen && newVal && this.isCopying) {
        this.populateFromSource(newVal);
      }
    },
  },
  methods: {
    getCategoryTheme(colorKey) {
      return getCategoryColor(colorKey);
    },
    async fetchCategories() {
      try {
        const res = await apiClient.get('/parameter-categories');
        this.availableCategories = (res && res.data && res.data.data) ? res.data.data : [];
      } catch (err) {
        console.error('Failed to load categories in ParameterModal:', err);
      }
    },
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
      if (str.includes('scale cannot be zero') || str.includes('scale factor cannot be zero')) {
        return this.$t('parameterModal.scaleZero');
      }
      if (str.includes('parameter_code is required')) {
        return this.$t('parameterModal.codeRequired');
      }
      if (str.includes('parameter_name is required')) {
        return this.$t('parameterModal.nameRequired');
      }
      return str;
    },
    scrollToError() {
      this.$nextTick(() => {
        if (this.$refs.formBody) {
          this.$refs.formBody.scrollTop = 0;
        }
      });
    },
    closeModal() {
      this.error = null;
      this.$emit('close');
    },
    resetForm() {
      const dev = this.currentDevice;
      this.form = {
        category_id: null,
        parameter_code: '',
        parameter_name: '',
        data_type: 'FLOAT32',
        unit: '',
        icon: '',
        description: '',
        slave_id: (dev && dev.connection && dev.connection.slave_id) || 1,
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
        polling_interval_ms: null,
        enabled: true,
        formula: '',
        hold_last_value_enabled: false,
        hold_last_value_seconds: 120,
        zero_value_alarm_policy: 'NORMAL',
        batch_mode: 'AUTO',
      };
      this.showAdvancedSettings = false;
      this.initByteOrderFromParam('', 'FLOAT32');
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
      const dt = param.data_type || 'FLOAT32';
      const dev = this.currentDevice;
      this.form = {
        category_id: param.category_id !== undefined && param.category_id !== null ? Number(param.category_id) : (param.category ? Number(param.category.id) : null),
        parameter_code: param.parameter_code || param.code || '',
        parameter_name: param.parameter_name || param.name || '',
        data_type: dt,
        unit: param.unit || '',
        icon: param.icon || '',
        description: param.description || '',
        slave_id: (param.slave_id !== undefined && param.slave_id !== null && param.slave_id > 0) ? param.slave_id : ((dev && dev.connection && dev.connection.slave_id) || 1),
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
        polling_interval_ms: (param.polling_interval_ms !== undefined && param.polling_interval_ms !== null && param.polling_interval_ms > 0) ? param.polling_interval_ms : null,
        enabled: param.enabled !== undefined ? param.enabled : true,
        formula: param.formula || '',
        hold_last_value_enabled: param.hold_last_value_enabled !== undefined ? param.hold_last_value_enabled : false,
        hold_last_value_seconds: param.hold_last_value_seconds || 120,
        zero_value_alarm_policy: param.zero_value_alarm_policy || 'NORMAL',
        batch_mode: param.batch_mode || 'AUTO',
      };
      this.showAdvancedSettings = !!(param.formula || param.hold_last_value_enabled || param.spike_detection_enabled);
      this.initByteOrderFromParam(param.byte_order, dt);
    },
    populateFromSource(source) {
      this.targetDeviceId = Number(this.deviceId);
      const code = source.parameter_code || source.code || '';
      const name = source.parameter_name || source.name || '';
      const dt = source.data_type || 'FLOAT32';
      this.form = {
        category_id: source.category_id !== undefined && source.category_id !== null ? Number(source.category_id) : (source.category ? Number(source.category.id) : null),
        parameter_code: this.generateSuggestedCode(code),
        parameter_name: this.generateSuggestedName(name),
        data_type: dt,
        unit: source.unit || '',
        icon: source.icon || '',
        description: source.description || '',
        slave_id: (source.slave_id !== undefined && source.slave_id !== null && source.slave_id > 0) ? source.slave_id : 1,
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
        polling_interval_ms: (source.polling_interval_ms !== undefined && source.polling_interval_ms !== null && source.polling_interval_ms > 0) ? source.polling_interval_ms : null,
        enabled: source.enabled !== undefined ? source.enabled : true,
        formula: source.formula || '',
        hold_last_value_enabled: source.hold_last_value_enabled !== undefined ? source.hold_last_value_enabled : false,
        hold_last_value_seconds: source.hold_last_value_seconds || 120,
        zero_value_alarm_policy: source.zero_value_alarm_policy || 'NORMAL',
        batch_mode: source.batch_mode || 'AUTO',
      };
      this.showAdvancedSettings = !!(source.formula || source.hold_last_value_enabled || source.spike_detection_enabled);
      this.initByteOrderFromParam(source.byte_order, dt);
    },
    onByteOrderSelectionChange(val) {
      this.byteOrderSelection = val;
      this.unknownByteOrderWarning = null;
      if (val === 'CUSTOM') {
        if (!this.customByteSlots || this.customByteSlots.length !== this.byteOrderSize) {
          this.customByteSlots = this.availableInputBytes.slice();
        }
      }
      this.syncByteOrderToForm();
    },
    handleDataTypeChange() {
      const size = this.byteOrderSize;
      this.unknownByteOrderWarning = null;
      if (this.byteOrderSelection === 'CUSTOM') {
        if (!this.customByteSlots || this.customByteSlots.length !== size) {
          this.customByteSlots = this.availableInputBytes.slice();
        }
      } else if (this.byteOrderSelection && this.byteOrderSelection !== '') {
        const exists = this.relevantPresets.some(p => p.value === this.byteOrderSelection);
        if (!exists) {
          if (size === 2 && (this.byteOrderSelection === 'ABCD' || this.byteOrderSelection === 'CDAB')) {
            this.byteOrderSelection = 'AB';
          } else if (size === 2 && (this.byteOrderSelection === 'BADC' || this.byteOrderSelection === 'DCBA')) {
            this.byteOrderSelection = 'BA';
          } else if (size === 8 && this.byteOrderSelection === 'ABCD') {
            this.byteOrderSelection = 'ABCDEFGH';
          } else {
            this.byteOrderSelection = '';
          }
        }
      }
      this.syncByteOrderToForm();
    },
    syncByteOrderToForm() {
      if (this.byteOrderSelection === 'CUSTOM') {
        this.form.byte_order = 'CUSTOM:' + (this.customByteSlots || []).join('');
      } else {
        this.form.byte_order = this.byteOrderSelection || '';
      }
    },
    onCustomSlotChange(idx, val) {
      this.$set(this.customByteSlots, idx, val);
      this.syncByteOrderToForm();
    },
    applyCustomQuickPreset(presetType) {
      const size = this.byteOrderSize;
      const base = this.availableInputBytes.slice();
      if (presetType === 'DEFAULT') {
        this.customByteSlots = base;
      } else if (presetType === 'REVERSE') {
        this.customByteSlots = base.slice().reverse();
      } else if (presetType === 'WORD_SWAP') {
        if (size === 4) {
          this.customByteSlots = ['C', 'D', 'A', 'B'];
        } else if (size === 8) {
          this.customByteSlots = ['G', 'H', 'E', 'F', 'C', 'D', 'A', 'B'];
        }
      } else if (presetType === 'BYTE_SWAP') {
        if (size === 2) {
          this.customByteSlots = ['B', 'A'];
        } else if (size === 4) {
          this.customByteSlots = ['B', 'A', 'D', 'C'];
        } else if (size === 8) {
          this.customByteSlots = ['B', 'A', 'D', 'C', 'F', 'E', 'H', 'G'];
        }
      }
      this.syncByteOrderToForm();
    },
    isSlotDuplicate(val, slotIdx) {
      if (!this.customByteSlots) return false;
      return this.customByteSlots.some((s, i) => s === val && i !== slotIdx);
    },
    getByteBadgeClass(byteChar) {
      const map = {
        'A': 'bg-blue-100 text-blue-800 border-blue-300 dark:bg-blue-500/20 dark:text-blue-300 dark:border-blue-500/40',
        'B': 'bg-cyan-100 text-cyan-800 border-cyan-300 dark:bg-cyan-500/20 dark:text-cyan-300 dark:border-cyan-500/40',
        'C': 'bg-amber-100 text-amber-800 border-amber-300 dark:bg-amber-500/20 dark:text-amber-300 dark:border-amber-500/40',
        'D': 'bg-emerald-100 text-emerald-800 border-emerald-300 dark:bg-emerald-500/20 dark:text-emerald-300 dark:border-emerald-500/40',
        'E': 'bg-purple-100 text-purple-800 border-purple-300 dark:bg-purple-500/20 dark:text-purple-300 dark:border-purple-500/40',
        'F': 'bg-rose-100 text-rose-800 border-rose-300 dark:bg-rose-500/20 dark:text-rose-300 dark:border-rose-500/40',
        'G': 'bg-teal-100 text-teal-800 border-teal-300 dark:bg-teal-500/20 dark:text-teal-300 dark:border-teal-500/40',
        'H': 'bg-indigo-100 text-indigo-800 border-indigo-300 dark:bg-indigo-500/20 dark:text-indigo-300 dark:border-indigo-500/40',
      };
      return map[byteChar] || 'bg-slate-100 text-slate-800 border-slate-300 dark:bg-slate-700 dark:text-slate-300 dark:border-slate-600';
    },
    initByteOrderFromParam(byteOrderStr, dataType) {
      const raw = (byteOrderStr || '').trim();
      this.unknownByteOrderWarning = null;
      const size = (dataType === 'INT16' || dataType === 'UINT16') ? 2 : (dataType === 'FLOAT64' ? 8 : ((dataType === 'BOOLEAN' || dataType === 'STRING') ? 2 : 4));
      const baseSlots = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'].slice(0, size);

      if (!raw) {
        this.byteOrderSelection = '';
        this.customByteSlots = baseSlots;
        this.form.byte_order = '';
        return;
      }
      if (raw.toUpperCase().startsWith('CUSTOM:')) {
        this.byteOrderSelection = 'CUSTOM';
        const perm = raw.substring(7).replace(/[\s,-]/g, '').toUpperCase();
        if (perm.length === size) {
          this.customByteSlots = perm.split('');
        } else {
          this.customByteSlots = baseSlots;
        }
        this.form.byte_order = raw;
        return;
      }
      const clean = raw.replace(/[\s,-]/g, '').toUpperCase();
      const match = this.relevantPresets.find(p => p.value === clean);
      if (match) {
        this.byteOrderSelection = clean;
        this.customByteSlots = clean.split('');
        this.form.byte_order = clean;
        return;
      }
      const validLetters = new Set(baseSlots);
      if (clean.length === size && new Set(clean.split('')).size === size && clean.split('').every(ch => validLetters.has(ch))) {
        this.byteOrderSelection = 'CUSTOM';
        this.customByteSlots = clean.split('');
        this.form.byte_order = 'CUSTOM:' + clean;
        return;
      }
      // Unrecognized order
      this.unknownByteOrderWarning = this.$t('parameterModal.unknownByteOrderWarning', { order: raw, type: dataType }) || `Saved byte order '${raw}' is unrecognized for ${dataType}`;
      this.byteOrderSelection = 'CUSTOM';
      this.customByteSlots = baseSlots;
      this.form.byte_order = raw;
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
      const pollInt = this.cleanNum(this.form.polling_interval_ms);
      const clearPollInt = pollInt === null || pollInt <= 0;

      const rawCat = this.form.category_id;
      const catId = (rawCat !== null && rawCat !== undefined && rawCat !== '' && Number(rawCat) > 0)
        ? Number(rawCat)
        : null;

      const payload = {
        category_id: catId,
        clear_category_id: catId === null,
        parameter_code: this.form.parameter_code ? this.form.parameter_code.trim() : '',
        parameter_name: this.form.parameter_name ? this.form.parameter_name.trim() : '',
        data_type: this.form.data_type,
        unit: this.form.unit ? this.form.unit.trim() : '',
        icon: this.form.icon ? this.form.icon.trim() : '',
        description: this.form.description ? this.form.description.trim() : '',
        slave_id: this.cleanNum(this.form.slave_id) || 1,
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
        polling_interval_ms: clearPollInt ? null : pollInt,
        clear_polling_interval_ms: clearPollInt,
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
        zero_value_alarm_policy: this.form.zero_value_alarm_policy || 'NORMAL',
        batch_mode: this.form.batch_mode || 'AUTO',
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
