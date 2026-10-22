<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-700/80 rounded-2xl w-full max-w-3xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-200 dark:border-slate-800 flex items-center justify-between bg-slate-50/90 dark:bg-[#0B0F19]/80 backdrop-blur-xs">
        <div>
          <h2 class="text-base font-bold text-slate-900 dark:text-white tracking-wide">
            {{ modalTitle }}
          </h2>
          <p class="text-xs text-slate-500 dark:text-slate-400 mt-0.5">
            {{ modalSubtitle }}
          </p>
        </div>
        <button @click="closeModal" class="text-slate-400 hover:text-slate-700 dark:hover:text-white p-1 rounded-lg hover:bg-slate-100 dark:hover:bg-slate-800 transition">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Navigation Tabs -->
      <div class="flex border-b border-slate-200 dark:border-slate-800 bg-slate-50/60 dark:bg-[#0B0F19]/40 px-6 text-xs font-medium">
        <button
          type="button"
          @click="activeTab = 'basic'"
          :class="activeTab === 'basic' ? 'border-b-2 border-blue-600 text-blue-600 dark:border-blue-500 dark:text-blue-400 font-semibold' : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200'"
          class="py-3 px-3 transition-colors flex items-center space-x-1.5"
        >
          <span>{{ $t('deviceModal.tabGeneral') }}</span>
        </button>
        <button
          type="button"
          @click="activeTab = 'location'"
          :class="activeTab === 'location' ? 'border-b-2 border-blue-600 text-blue-600 dark:border-blue-500 dark:text-blue-400 font-semibold' : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200'"
          class="py-3 px-3 transition-colors flex items-center space-x-1.5"
        >
          <span>{{ $t('deviceModal.tabLocation') }}</span>
        </button>
        <button
          type="button"
          @click="activeTab = 'connection'"
          :class="activeTab === 'connection' ? 'border-b-2 border-blue-600 text-blue-600 dark:border-blue-500 dark:text-blue-400 font-semibold' : 'text-slate-500 hover:text-slate-700 dark:text-slate-400 dark:hover:text-slate-200'"
          class="py-3 px-3 transition-colors flex items-center space-x-1.5"
        >
          <span>{{ $t('deviceModal.tabComm') }}</span>
          <span class="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
        </button>
      </div>

      <!-- Modal Body -->
      <form ref="deviceFormBody" novalidate @submit.prevent="submitForm" class="flex-1 overflow-y-auto p-6 space-y-5">
        <!-- Copy Source Indicator Banner -->
        <div v-if="isCopying && sourceDevice" class="p-3.5 bg-cyan-50 dark:bg-cyan-950/40 border border-cyan-200 dark:border-cyan-700/60 rounded-xl flex items-start space-x-3 text-xs shadow-xs animate-fade-in">
          <span class="text-cyan-600 dark:text-cyan-400 mt-0.5 text-base leading-none">📋</span>
          <div class="flex-1">
            <div class="font-bold text-cyan-950 dark:text-cyan-200">
              {{ $t('deviceModal.copyBannerTitle', { code: sourceDevice.device_code || sourceDevice.code }) }}
              <span class="font-normal text-cyan-700 dark:text-cyan-400">({{ sourceDevice.device_name || sourceDevice.name }})</span>
            </div>
            <div class="text-3xs text-cyan-800 dark:text-cyan-300/80 mt-1 leading-relaxed">
              {{ $t('deviceModal.copyBannerDesc') }}
            </div>
          </div>
        </div>

        <!-- Error Alert -->
        <div v-if="error" class="p-3.5 bg-rose-50 dark:bg-rose-500/15 border-2 border-rose-200 dark:border-rose-500/40 rounded-xl text-xs text-rose-900 dark:text-rose-300 flex items-start space-x-2.5 shadow-xs animate-fade-in">
          <svg class="w-4 h-4 text-rose-500 dark:text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"></path>
          </svg>
          <div class="flex-1">
            <span class="font-bold text-rose-950 dark:text-rose-200 block text-2xs">{{ $t('deviceModal.cannotSaveTitle') }}</span>
            <span class="text-3xs text-rose-800 dark:text-rose-300 mt-0.5 block leading-relaxed whitespace-pre-wrap">{{ error }}</span>
          </div>
          <button type="button" @click="error = null" class="text-rose-500 hover:text-rose-700 dark:hover:text-white text-3xs font-mono ml-2">✕</button>
        </div>

        <!-- TAB 1: General Info -->
        <div v-show="activeTab === 'basic'" class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Device Code -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.deviceCode') }}
              </label>
              <input
                v-model="form.device_code"
                type="text"
                :placeholder="$t('deviceModal.placeholders.code')"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Device Name -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.deviceName') }}
              </label>
              <input
                v-model="form.device_name"
                type="text"
                :placeholder="$t('deviceModal.placeholders.name')"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Administrative Status -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.adminStatus') }}
              </label>
              <select
                v-model="form.status"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
              >
                <option value="ACTIVE">{{ $t('status.active') }}</option>
                <option value="INACTIVE">{{ $t('status.inactive') }}</option>
                <option value="MAINTENANCE">{{ $t('status.maintenance') }}</option>
                <option value="DISABLED">{{ $t('status.disabled') }}</option>
              </select>
            </div>

            <!-- Enabled Switch -->
            <div class="flex flex-col justify-end">
              <label class="flex items-center space-x-3 cursor-pointer p-2.5 rounded-xl bg-slate-50 dark:bg-[#0B0F19] border border-slate-200 dark:border-slate-800 hover:border-slate-300 dark:hover:border-slate-700 transition">
                <input
                  v-model="form.enabled"
                  type="checkbox"
                  class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-white dark:bg-slate-900 border-slate-300 dark:border-slate-700"
                />
                <div>
                  <span class="text-xs font-semibold text-slate-900 dark:text-slate-200">{{ $t('deviceModal.deviceEnabled') }}</span>
                  <span class="block text-3xs text-slate-500 dark:text-slate-400">{{ $t('deviceModal.includeInCycle') }}</span>
                </div>
              </label>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Manufacturer -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.manufacturer') }}
              </label>
              <input
                v-model="form.manufacturer"
                type="text"
                :placeholder="$t('deviceModal.placeholders.manufacturer')"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>

            <!-- Model -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.model') }}
              </label>
              <input
                v-model="form.model"
                type="text"
                :placeholder="$t('deviceModal.placeholders.model')"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Serial Number -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.serialNumber') }}
              </label>
              <input
                v-model="form.serial_number"
                type="text"
                :placeholder="$t('deviceModal.placeholders.serialNumber')"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Firmware Version -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.firmwareVersion') }}
              </label>
              <input
                v-model="form.firmware_version"
                type="text"
                :placeholder="$t('deviceModal.placeholders.firmware')"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
          </div>

          <!-- Description -->
          <div>
            <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('deviceModal.description') }}
            </label>
            <textarea
              v-model="form.description"
              rows="2"
              :placeholder="$t('deviceModal.placeholders.description')"
              class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 resize-none"
            ></textarea>
          </div>
        </div>

        <!-- TAB 2: Location & Timezone -->
        <div v-show="activeTab === 'location'" class="space-y-4">
          <!-- Physical Location -->
          <div>
            <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('deviceModal.location') }}
            </label>
            <input
              v-model="form.location"
              type="text"
              :placeholder="$t('deviceModal.placeholders.location')"
              class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <!-- Latitude -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.latitude') }}
              </label>
              <input
                v-model.number="form.latitude"
                type="number"
                step="0.000001"
                min="-90"
                max="90"
                placeholder="-6.2088"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Longitude -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.longitude') }}
              </label>
              <input
                v-model.number="form.longitude"
                type="number"
                step="0.000001"
                min="-180"
                max="180"
                placeholder="106.8456"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Timezone -->
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.timezone') }}
              </label>
              <select
                v-model="form.timezone"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white focus:outline-none focus:border-blue-500 font-mono"
              >
                <option value="UTC">UTC (Coordinated Universal Time)</option>
                <option value="Asia/Jakarta">Asia/Jakarta (WIB, UTC+7)</option>
                <option value="Asia/Makassar">Asia/Makassar (WITA, UTC+8)</option>
                <option value="Asia/Jayapura">Asia/Jayapura (WIT, UTC+9)</option>
                <option value="Asia/Singapore">Asia/Singapore (SGT, UTC+8)</option>
                <option value="Europe/London">Europe/London (GMT/BST)</option>
                <option value="America/New_York">America/New_York (EST/EDT)</option>
              </select>
            </div>
          </div>
        </div>

        <!-- TAB 3: Communication & Protocol -->
        <div v-show="activeTab === 'connection'" class="space-y-4">
          <div class="p-3 bg-blue-50/80 dark:bg-blue-500/10 border border-blue-200 dark:border-blue-500/20 rounded-xl flex items-center justify-between text-xs text-blue-900 dark:text-blue-300 shadow-2xs">
            <span class="font-medium">{{ $t('deviceModal.protocolArchitecture') }}</span>
            <span class="text-3xs font-mono px-2 py-0.5 rounded bg-blue-100 text-blue-800 border border-blue-200 dark:bg-blue-500/20 dark:text-blue-200 dark:border-transparent font-semibold">{{ $t('deviceModal.phase21Foundation') }}</span>
          </div>

          <!-- Protocol Selector -->
          <div>
            <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('deviceModal.protocol') }}
            </label>
            <select
              v-model="form.connection.protocol"
              @change="onProtocolChange"
              class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
            >
              <option value="MODBUS_TCP">MODBUS_TCP (Ethernet / WiFi Master)</option>
              <option value="MODBUS_RTU">MODBUS_RTU (RS485 / RS232 Serial Master)</option>
              <option value="TCP">TCP (Raw Industrial Socket)</option>
              <option value="UDP">UDP (Datagram Receiver)</option>
              <option value="HTTP">HTTP / REST (Polling Client)</option>
              <option value="MQTT">MQTT (Topic Subscriber)</option>
              <option value="SERIAL">SERIAL (Raw UART / COM Port)</option>
              <option value="WEBSOCKET">WEBSOCKET (Live Stream Client)</option>
              <option value="CUSTOM">CUSTOM (Proprietary Protocol)</option>
            </select>
          </div>

          <!-- DYNAMIC PROTOCOL FIELDS -->
          <!-- 1. For Networked (MODBUS_TCP, TCP, UDP, HTTP, MQTT, WEBSOCKET) -->
          <div v-if="isNetworkProtocol" class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div class="md:col-span-2">
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.hostIp') }}
              </label>
              <input
                v-model="form.connection.host"
                type="text"
                :placeholder="$t('deviceModal.placeholders.host')"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.port') }}
              </label>
              <input
                v-model.number="form.connection.port"
                type="number"
                placeholder="502"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
          </div>

          <!-- 2. For Serial (MODBUS_RTU, SERIAL) -->
          <div v-if="isSerialProtocol" class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <div class="flex items-center justify-between mb-1.5">
                  <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider">
                    {{ $t('deviceModal.hardwareAddress') }}
                  </label>
                  <button
                    type="button"
                    @click="fetchAvailableSerialPorts"
                    :disabled="loadingPorts"
                    class="text-3xs text-blue-600 dark:text-blue-400 hover:text-blue-500 dark:hover:text-blue-300 font-semibold flex items-center space-x-1"
                    title="Scan available hardware COM ports on system"
                  >
                    <svg class="w-3 h-3" :class="{ 'animate-spin': loadingPorts }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                    </svg>
                    <span>{{ loadingPorts ? $t('deviceModal.scanning') : $t('deviceModal.scanPorts') }}</span>
                  </button>
                </div>

                <div class="space-y-1.5">
                  <div v-if="!useCustomPort" class="relative">
                    <select
                      v-model="form.connection.serial_port"
                      class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
                    >
                      <!-- 1. Persistent by-ID (Chip Serial Number: Recommended for field use) -->
                      <optgroup v-if="byIDPorts.length > 0" :label="$t('deviceModal.optgroupById')">
                        <option v-for="port in byIDPorts" :key="'byid-' + port.path" :value="port.path">
                          {{ port.description || port.path }}
                        </option>
                      </optgroup>

                      <!-- 2. Persistent by-Path (Physical USB Socket on Pi) -->
                      <optgroup v-if="byPathPorts.length > 0" :label="$t('deviceModal.optgroupByPath')">
                        <option v-for="port in byPathPorts" :key="'bypath-' + port.path" :value="port.path">
                          {{ port.description || port.path }}
                        </option>
                      </optgroup>

                      <!-- 3. Standard Direct Ports -->
                      <optgroup v-if="standardPorts.length > 0" :label="$t('deviceModal.optgroupDirect')">
                        <option v-for="port in standardPorts" :key="'std-' + port.path" :value="port.path">
                          {{ port.path }} — {{ port.description }}
                        </option>
                      </optgroup>

                      <!-- Fallback list if nothing detected -->
                      <optgroup v-if="detailedPorts.length === 0" :label="$t('deviceModal.optgroupStandard')">
                        <option v-for="port in fallbackPorts" :key="'fb-' + port" :value="port">
                          {{ port }}
                        </option>
                      </optgroup>
                    </select>
                  </div>

                  <input
                    v-else
                    v-model="form.connection.serial_port"
                    type="text"
                    :placeholder="$t('deviceModal.placeholders.hardwareAddress')"
                    class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
                  />

                  <!-- Field stability alert / recommendation tip -->
                  <div class="p-2.5 rounded-lg bg-blue-50 dark:bg-blue-500/10 border border-blue-200 dark:border-blue-500/20 text-[10px] text-blue-900 dark:text-blue-300 font-sans flex items-start space-x-2 mt-1.5">
                    <svg class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path>
                    </svg>
                    <div class="space-y-0.5">
                      <span class="font-semibold text-blue-800 dark:text-blue-200">{{ $t('deviceModal.resilienceTipTitle') }}</span>
                      <p class="text-blue-700/90 dark:text-slate-300">
                        {{ $t('deviceModal.resilienceTipText') }}
                      </p>
                    </div>
                  </div>

                  <!-- Serial Port Conflict Warning (Connection safety Rule 3E) -->
                  <div v-if="serialConflictWarning" class="p-2.5 rounded-lg bg-amber-50 dark:bg-amber-500/15 border border-amber-200 dark:border-amber-500/30 text-3xs text-amber-900 dark:text-amber-200 font-sans flex items-start space-x-2 mt-1.5 shadow-2xs">
                    <svg class="w-3.5 h-3.5 text-amber-600 dark:text-amber-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
                    </svg>
                    <span class="leading-relaxed">{{ serialConflictWarning }}</span>
                  </div>

                  <div class="flex items-center justify-between text-3xs text-slate-500 dark:text-slate-400 pt-0.5">
                    <span v-if="systemPorts.length > 0 && !useCustomPort" class="text-emerald-600 dark:text-emerald-400 font-mono flex items-center space-x-1 font-medium">
                      <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 dark:bg-emerald-400 animate-pulse"></span>
                      <span>{{ $t('deviceModal.pathsDetected', { count: systemPorts.length }) }}</span>
                    </span>
                    <span v-else-if="!useCustomPort" class="text-slate-500">
                      {{ $t('deviceModal.noPortsDetected') }}
                    </span>
                    <span v-else class="text-slate-500 dark:text-slate-400">
                      {{ $t('deviceModal.customPortMode') }}
                    </span>
                    <button
                      type="button"
                      @click="useCustomPort = !useCustomPort"
                      class="text-blue-600 dark:text-blue-400 hover:text-blue-500 dark:hover:text-blue-300 font-medium underline ml-auto text-3xs"
                    >
                      {{ useCustomPort ? $t('deviceModal.chooseFromList') : $t('deviceModal.typeCustomPort') }}
                    </button>
                  </div>
                </div>
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                  {{ $t('deviceModal.baudRate') }}
                </label>
                <select
                  v-model.number="form.connection.baud_rate"
                  class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
                >
                  <option :value="2400">2400</option>
                  <option :value="4800">4800</option>
                  <option :value="9600">9600 (Standard)</option>
                  <option :value="19200">19200</option>
                  <option :value="38400">38400</option>
                  <option :value="57600">57600</option>
                  <option :value="115200">115200</option>
                </select>
              </div>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                  {{ $t('deviceModal.dataBits') }}
                </label>
                <select
                  v-model.number="form.connection.data_bits"
                  class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
                >
                  <option :value="8">8 (Standard)</option>
                  <option :value="7">7</option>
                </select>
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                  {{ $t('deviceModal.parity') }}
                </label>
                <select
                  v-model="form.connection.parity"
                  class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
                >
                  <option value="N">None (N)</option>
                  <option value="E">Even (E)</option>
                  <option value="O">Odd (O)</option>
                </select>
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                  {{ $t('deviceModal.stopBits') }}
                </label>
                <select
                  v-model.number="form.connection.stop_bits"
                  class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
                >
                  <option :value="1">1 (Standard)</option>
                  <option :value="2">2</option>
                </select>
              </div>
            </div>
          </div>

          <!-- 3. For Custom Protocol -->
          <div v-if="isCustomProtocol" class="p-3.5 bg-slate-50 dark:bg-slate-900/60 rounded-xl border border-slate-200 dark:border-slate-800 space-y-3">
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.connectionMedium') || 'Connection Medium' }}
              </label>
              <select
                v-model="form.connection.connection_type"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-blue-500"
              >
                <option value="ETHERNET">Ethernet / IP Socket</option>
                <option value="SERIAL">Serial / RS485 / UART</option>
              </select>
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.extraConfig') || 'Custom Configuration (JSON / Framing Options)' }}
              </label>
              <textarea
                v-model="form.connection.extra_config"
                rows="2"
                placeholder='e.g. {"delimiter": "\r\n", "header": "0xAA"}'
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono resize-none"
              ></textarea>
            </div>
          </div>

          <!-- Common Timing Configuration -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2 border-t border-slate-200 dark:border-slate-800/80">
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.timeout') }}
              </label>
              <input
                v-model.number="form.connection.timeout"
                type="number"
                min="50"
                step="50"
                placeholder="1000"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.retryCount') }}
              </label>
              <input
                v-model.number="form.connection.retry_count"
                type="number"
                min="0"
                max="10"
                placeholder="3"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.pollInterval') }}
              </label>
              <input
                v-model.number="form.connection.polling_interval"
                type="number"
                min="100"
                step="100"
                placeholder="1000"
                class="w-full px-3 py-2 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white placeholder-slate-400 dark:placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
          </div>

          <!-- 4. Multi-Address Batch Read Configuration (Modbus Protocols) -->
          <div v-if="isModbusProtocol" class="p-3.5 bg-gradient-to-br from-indigo-50/70 to-blue-50/40 dark:from-indigo-950/20 dark:to-blue-950/20 border border-indigo-200/80 dark:border-indigo-800/40 rounded-xl space-y-3 shadow-2xs">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-2.5">
                <div class="w-7 h-7 rounded-lg bg-indigo-500/10 dark:bg-indigo-500/20 flex items-center justify-center text-indigo-600 dark:text-indigo-400">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                  </svg>
                </div>
                <div>
                  <span class="text-xs font-bold text-slate-900 dark:text-white">{{ $t('deviceModal.batchRead') || 'Multi-Address Batch Read' }}</span>
                  <p class="text-3xs text-slate-500 dark:text-slate-400">{{ $t('deviceModal.batchReadDesc') || 'Consolidate multiple parameter requests into optimal contiguous Modbus frames' }}</p>
                </div>
              </div>
              <label class="relative inline-flex items-center cursor-pointer">
                <input
                  v-model="form.connection.batch_read_enabled"
                  type="checkbox"
                  class="sr-only peer"
                />
                <div class="w-9 h-5 bg-slate-200 peer-focus:outline-none rounded-full peer dark:bg-slate-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-indigo-600"></div>
              </label>
            </div>

            <!-- Batch Fine-tuning Parameters (Only when Batch Read is Enabled) -->
            <div v-if="form.connection.batch_read_enabled" class="grid grid-cols-1 md:grid-cols-2 gap-3 pt-2.5 border-t border-indigo-100 dark:border-indigo-900/40">
              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                  {{ $t('deviceModal.maxBatchRegisters') || 'Max Batch Registers' }}
                </label>
                <input
                  v-model.number="form.connection.max_batch_registers"
                  type="number"
                  min="1"
                  max="120"
                  placeholder="32"
                  class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-indigo-500"
                />
                <span class="text-3xs text-slate-500 dark:text-slate-400 mt-0.5 block">{{ $t('deviceModal.maxBatchRegistersHelp') || 'Default 32 (Clamped 1 - 120 registers)' }}</span>
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-700 dark:text-slate-300 uppercase tracking-wider mb-1">
                  {{ $t('deviceModal.maxRegisterGap') || 'Max Register Gap' }}
                </label>
                <input
                  v-model.number="form.connection.max_register_gap"
                  type="number"
                  min="0"
                  max="10"
                  placeholder="0"
                  class="w-full px-3 py-1.5 bg-white dark:bg-[#0B0F19] border border-slate-300 dark:border-slate-700/80 rounded-xl text-xs text-slate-900 dark:text-white font-mono focus:outline-none focus:border-indigo-500"
                />
                <span class="text-3xs text-slate-500 dark:text-slate-400 mt-0.5 block">{{ $t('deviceModal.maxRegisterGapHelp') || 'Default 0 (Unmapped address gap tolerance 0 - 10)' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Sticky Warning Alert Above Buttons -->
        <div v-if="error" class="p-3.5 bg-rose-50 dark:bg-rose-500/15 border-2 border-rose-200 dark:border-rose-500/50 rounded-xl text-xs text-rose-900 dark:text-rose-200 flex items-start space-x-2.5 shadow-xs animate-fade-in">
          <svg class="w-5 h-5 text-rose-500 dark:text-rose-400 flex-shrink-0 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
          </svg>
          <div class="flex-1">
            <div class="font-bold flex items-center justify-between text-rose-950 dark:text-rose-100">
              <span class="text-xs">{{ $t('deviceModal.cannotSaveTitle') }}</span>
              <button type="button" @click="error = null" class="text-rose-600 dark:text-rose-400 hover:text-rose-800 dark:hover:text-white text-3xs underline ml-2">{{ $t('common.cancel') }}</button>
            </div>
            <p class="mt-1 text-2xs text-rose-800 dark:text-rose-300 leading-relaxed font-sans whitespace-pre-wrap">{{ error }}</p>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="pt-4 border-t border-slate-200 dark:border-slate-800 flex items-center justify-between">
          <button
            type="button"
            @click="closeModal"
            class="px-4 py-2 rounded-xl text-xs font-medium text-slate-600 dark:text-slate-400 hover:text-slate-900 dark:hover:text-white hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          >
            {{ $t('common.cancel') }}
          </button>
          <div class="flex items-center space-x-2">
            <button
              v-if="activeTab !== 'basic'"
              type="button"
              @click="prevTab"
              class="px-4 py-2 rounded-xl text-xs font-semibold text-slate-700 dark:text-slate-300 bg-white hover:bg-slate-100 dark:bg-slate-800 dark:hover:bg-slate-700 border border-slate-300 dark:border-slate-700 shadow-xs transition"
            >
              {{ $t('common.previous') }}
            </button>
            <button
              v-if="activeTab !== 'connection'"
              type="button"
              @click="nextTab"
              class="px-4 py-2 rounded-xl text-xs font-semibold text-white bg-blue-600 hover:bg-blue-500 shadow-sm transition"
            >
              {{ $t('common.next') }}
            </button>
            <button
              v-if="activeTab === 'connection'"
              type="submit"
              :disabled="loading"
              class="px-5 py-2 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-500 shadow-md shadow-emerald-900/20 transition flex items-center space-x-1.5 disabled:opacity-50"
            >
              <span v-if="loading" class="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
              <span>{{ saveButtonText }}</span>
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'DeviceModal',
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
    deviceToEdit: {
      type: Object,
      default: null,
    },
    isCopying: {
      type: Boolean,
      default: false,
    },
    sourceDevice: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      activeTab: 'basic',
      loading: false,
      error: null,
      systemPorts: [],
      detailedPorts: [],
      loadingPorts: false,
      useCustomPort: false,
      fallbackPorts: ['COM1', 'COM2', 'COM3', 'COM4', 'COM5', 'COM6', 'COM7', 'COM8', '/dev/ttyUSB0', '/dev/ttyUSB1', '/dev/ttyACM0', '/dev/ttyACM1', '/dev/ttyS0'],
      form: {
        device_code: '',
        device_name: '',
        device_type: 'MODBUS_TCP',
        manufacturer: '',
        model: '',
        serial_number: '',
        firmware_version: '',
        description: '',
        location: '',
        latitude: null,
        longitude: null,
        timezone: 'UTC',
        status: 'ACTIVE',
        enabled: true,
        connection: {
          protocol: 'MODBUS_TCP',
          connection_type: 'ETHERNET',
          host: '',
          port: 502,
          slave_id: 1,
          byte_order: 'ABCD',
          serial_port: 'COM1',
          baud_rate: 9600,
          data_bits: 8,
          parity: 'N',
          stop_bits: 1,
          timeout: 1000,
          retry_count: 3,
          polling_interval: 1000,
          enabled: true,
          batch_read_enabled: false,
          max_batch_registers: 32,
          max_register_gap: 0,
        },
      },
    };
  },
  computed: {
    isEditing() {
      return !this.isCopying && !!this.deviceToEdit;
    },
    modalTitle() {
      if (this.isCopying) return this.$t('deviceModal.titleCopy') || 'Copy Edge Device Configuration';
      if (this.isEditing) return this.$t('deviceModal.titleEdit');
      return this.$t('deviceModal.titleAdd');
    },
    modalSubtitle() {
      if (this.isCopying) return this.$t('deviceModal.subtitleCopy') || 'Clone device configuration to create a new independent device.';
      if (this.isEditing) return this.$t('deviceModal.subtitleEdit');
      return this.$t('deviceModal.subtitleAdd');
    },
    saveButtonText() {
      if (this.isCopying) return this.$t('deviceModal.copyButton') || 'Create Copied Device';
      if (this.isEditing) return this.$t('common.saveChanges');
      return this.$t('deviceModal.registerDevice');
    },
    serialConflictWarning() {
      if (!this.isSerialProtocol || !this.form.connection || !this.form.connection.serial_port) return null;
      const port = this.form.connection.serial_port.trim().toLowerCase();
      if (!port) return null;
      const devices = this.$store.state.devices || [];
      const editId = this.isEditing ? this.deviceToEdit?.id : null;
      const conflict = devices.find(d => {
        if (editId && d.id === editId) return false;
        const dPort = (d.connection && d.connection.serial_port) || '';
        return dPort.trim().toLowerCase() === port;
      });
      if (conflict) {
        return this.$t('deviceModal.serialPortConflictWarning', {
          port: this.form.connection.serial_port,
          device: conflict.device_code || conflict.code || conflict.device_name || conflict.name,
        });
      }
      return null;
    },
    isCustomProtocol() {
      return this.form.connection.protocol === 'CUSTOM';
    },
    isModbusProtocol() {
      const p = this.form.connection && this.form.connection.protocol;
      return p === 'MODBUS_TCP' || p === 'MODBUS_RTU';
    },
    isNetworkProtocol() {
      const p = this.form.connection.protocol;
      return ['MODBUS_TCP', 'TCP', 'UDP', 'HTTP', 'MQTT', 'WEBSOCKET'].includes(p) || (p === 'CUSTOM' && this.form.connection.connection_type === 'ETHERNET');
    },
    isSerialProtocol() {
      const p = this.form.connection.protocol;
      return ['MODBUS_RTU', 'SERIAL'].includes(p) || (p === 'CUSTOM' && this.form.connection.connection_type === 'SERIAL');
    },
    isModbusProtocol() {
      const p = this.form.connection.protocol;
      return ['MODBUS_TCP', 'MODBUS_RTU'].includes(p);
    },
    byIDPorts() {
      return this.detailedPorts.filter(p => p.type === 'BY_ID');
    },
    byPathPorts() {
      return this.detailedPorts.filter(p => p.type === 'BY_PATH');
    },
    standardPorts() {
      return this.detailedPorts.filter(p => p.type === 'STANDARD');
    },
  },
  watch: {
    isOpen: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.fetchAvailableSerialPorts();
          if (this.isCopying && this.sourceDevice) {
            this.populateFromSource(this.sourceDevice);
          }
        }
      },
    },
    sourceDevice: {
      immediate: true,
      handler(newVal) {
        if (this.isCopying && newVal) {
          this.populateFromSource(newVal);
        }
      },
    },
    isCopying: {
      immediate: true,
      handler(newVal) {
        if (newVal && this.sourceDevice) {
          this.populateFromSource(this.sourceDevice);
        }
      },
    },
    deviceToEdit: {
      immediate: true,
      handler(newVal) {
        if (this.isCopying) return;
        if (newVal) {
          let proto = (newVal.connection && newVal.connection.protocol) || newVal.device_type || 'MODBUS_TCP';
          if (proto === 'HTTP/REST' || proto === 'REST') {
            proto = 'HTTP';
          }
          this.form = {
            device_code: newVal.device_code || newVal.code || '',
            device_name: newVal.device_name || newVal.name || '',
            device_type: proto,
            manufacturer: newVal.manufacturer || '',
            model: newVal.model || '',
            serial_number: newVal.serial_number || '',
            firmware_version: newVal.firmware_version || '',
            description: newVal.description || '',
            location: newVal.location || '',
            latitude: newVal.latitude || null,
            longitude: newVal.longitude || null,
            timezone: newVal.timezone || 'UTC',
            status: newVal.status || 'ACTIVE',
            enabled: newVal.enabled !== undefined ? newVal.enabled : true,
            connection: {
              protocol: proto,
              connection_type: (newVal.connection && newVal.connection.connection_type) || (proto === 'MODBUS_RTU' || proto === 'SERIAL' ? 'SERIAL' : 'ETHERNET'),
              host: (newVal.connection && (newVal.connection.host || newVal.connection.address)) || '',
              port: (newVal.connection && newVal.connection.port) || (proto === 'MODBUS_TCP' ? 502 : (proto === 'MQTT' ? 1883 : (proto === 'HTTP' ? 80 : 502))),
              slave_id: (newVal.connection && newVal.connection.slave_id) || 1,
              byte_order: (newVal.connection && newVal.connection.byte_order) || 'ABCD',
              serial_port: (newVal.connection && (newVal.connection.serial_port || newVal.connection.address)) || 'COM1',
              baud_rate: (newVal.connection && newVal.connection.baud_rate) || 9600,
              data_bits: (newVal.connection && newVal.connection.data_bits) || 8,
              parity: (newVal.connection && newVal.connection.parity) || 'N',
              stop_bits: (newVal.connection && newVal.connection.stop_bits) || 1,
              timeout: (newVal.connection && (newVal.connection.timeout || newVal.connection.timeout_ms)) || 1000,
              retry_count: (newVal.connection && newVal.connection.retry_count) || 3,
              polling_interval: (newVal.connection && (newVal.connection.polling_interval || newVal.connection.poll_interval_ms)) || 1000,
              enabled: newVal.connection ? newVal.connection.enabled : true,
              extra_config: (newVal.connection && newVal.connection.extra_config) || '',
              batch_read_enabled: !!(newVal.connection && newVal.connection.batch_read_enabled),
              max_batch_registers: (newVal.connection && newVal.connection.max_batch_registers) || 32,
              max_register_gap: (newVal.connection && newVal.connection.max_register_gap !== undefined) ? newVal.connection.max_register_gap : 0,
            },
          };
        } else {
          this.resetForm();
        }
      },
    },
  },
  mounted() {
    this.fetchAvailableSerialPorts();
  },
  methods: {
    async fetchAvailableSerialPorts() {
      this.loadingPorts = true;
      try {
        const res = await axios.get('/api/system/serial-ports/details');
        if (res.data && res.data.data) {
          this.detailedPorts = res.data.data || [];
          this.systemPorts = this.detailedPorts.map((p) => p.path);
          if (this.systemPorts.length > 0) {
            // If current port is empty or default COM1 and COM1 is not detected:
            if (!this.form.connection.serial_port || (this.form.connection.serial_port === 'COM1' && !this.systemPorts.includes('COM1'))) {
              // Prefer recommended persistent by-id or by-path port if available
              const rec = this.detailedPorts.find((p) => p.recommended);
              this.form.connection.serial_port = rec ? rec.path : this.systemPorts[0];
            }
          }
        }
      } catch (err) {
        console.warn('Failed to fetch detailed serial ports, trying basic endpoint:', err);
        try {
          const fallbackRes = await axios.get('/api/system/serial-ports');
          if (fallbackRes.data && fallbackRes.data.data) {
            this.systemPorts = fallbackRes.data.data || [];
          }
        } catch (fbErr) {
          console.warn('Failed to fetch fallback serial ports:', fbErr);
        }
      } finally {
        this.loadingPorts = false;
      }
    },
    generateSuggestedCode(sourceCode) {
      if (!sourceCode) return 'DEV-01';
      const existingCodes = (this.$store.state.devices || []).map(d => (d.device_code || d.code || '').toUpperCase());
      const match = sourceCode.match(/^(.*?)(\d+)$/);
      if (match) {
        const prefix = match[1];
        const numStr = match[2];
        const padLen = numStr.length;
        let nextNum = parseInt(numStr, 10) + 1;
        let candidate = `${prefix}${String(nextNum).padStart(padLen, '0')}`;
        while (existingCodes.includes(candidate.toUpperCase())) {
          nextNum++;
          candidate = `${prefix}${String(nextNum).padStart(padLen, '0')}`;
        }
        return candidate;
      }
      let candidate = `${sourceCode}_COPY`;
      let counter = 2;
      while (existingCodes.includes(candidate.toUpperCase())) {
        candidate = `${sourceCode}_COPY_${counter}`;
        counter++;
      }
      return candidate;
    },
    populateFromSource(source) {
      let proto = (source.connection && source.connection.protocol) || source.device_type || 'MODBUS_TCP';
      if (proto === 'HTTP/REST' || proto === 'REST') proto = 'HTTP';
      const suggestedCode = this.generateSuggestedCode(source.device_code || source.code);
      const baseName = source.device_name || source.name || 'Device';
      const suggestedName = `${baseName} (Copy)`;

      this.form = {
        device_code: suggestedCode,
        device_name: suggestedName,
        device_type: proto,
        manufacturer: source.manufacturer || '',
        model: source.model || '',
        serial_number: '',
        firmware_version: source.firmware_version || '',
        description: source.description || '',
        location: source.location || '',
        latitude: source.latitude || null,
        longitude: source.longitude || null,
        timezone: source.timezone || 'UTC',
        status: 'ACTIVE',
        enabled: false, // Rule 3E: disabled by default to prevent polling collision
        connection: {
          protocol: proto,
          connection_type: (source.connection && source.connection.connection_type) || (proto === 'MODBUS_RTU' || proto === 'SERIAL' ? 'SERIAL' : 'ETHERNET'),
          host: (source.connection && (source.connection.host || source.connection.address)) || '',
          port: (source.connection && source.connection.port) || (proto === 'MODBUS_TCP' ? 502 : 502),
          slave_id: (source.connection && source.connection.slave_id) || 1,
          byte_order: (source.connection && source.connection.byte_order) || 'ABCD',
          serial_port: (source.connection && (source.connection.serial_port || source.connection.address)) || 'COM1',
          baud_rate: (source.connection && source.connection.baud_rate) || 9600,
          data_bits: (source.connection && source.connection.data_bits) || 8,
          parity: (source.connection && source.connection.parity) || 'N',
          stop_bits: (source.connection && source.connection.stop_bits) || 1,
          timeout: (source.connection && (source.connection.timeout || source.connection.timeout_ms)) || 1000,
          retry_count: (source.connection && source.connection.retry_count) || 3,
          polling_interval: (source.connection && (source.connection.polling_interval || source.connection.poll_interval_ms)) || 1000,
          enabled: false,
          extra_config: (source.connection && source.connection.extra_config) || '',
          batch_read_enabled: !!(source.connection && source.connection.batch_read_enabled),
          max_batch_registers: (source.connection && source.connection.max_batch_registers) || 32,
          max_register_gap: (source.connection && source.connection.max_register_gap !== undefined) ? source.connection.max_register_gap : 0,
        },
      };
    },
    closeModal() {
      this.error = null;
      this.activeTab = 'basic';
      this.$emit('close');
    },
    nextTab() {
      this.error = null;
      if (this.activeTab === 'basic') {
        if (!this.form.device_code || !this.form.device_code.trim()) {
          this.error = 'Device Code is required before continuing.';
          return;
        }
        if (!this.form.device_name || !this.form.device_name.trim()) {
          this.error = 'Device Name is required before continuing.';
          return;
        }
        this.activeTab = 'location';
      } else if (this.activeTab === 'location') {
        this.activeTab = 'connection';
      }
    },
    prevTab() {
      this.error = null;
      if (this.activeTab === 'connection') this.activeTab = 'location';
      else if (this.activeTab === 'location') this.activeTab = 'basic';
    },
    onProtocolChange() {
      const p = this.form.connection.protocol;
      this.form.device_type = p;
      if (p === 'MODBUS_TCP') {
        this.form.connection.connection_type = 'ETHERNET';
        if (!this.form.connection.port || [1883, 80, 8080, 8000].includes(this.form.connection.port)) {
          this.form.connection.port = 502;
        }
        if (!this.form.connection.slave_id) this.form.connection.slave_id = 1;
        if (!this.form.connection.byte_order) this.form.connection.byte_order = 'ABCD';
      } else if (p === 'MODBUS_RTU') {
        this.form.connection.connection_type = 'SERIAL';
        if (!this.form.connection.baud_rate) this.form.connection.baud_rate = 9600;
        if (!this.form.connection.data_bits) this.form.connection.data_bits = 8;
        if (!this.form.connection.parity) this.form.connection.parity = 'N';
        if (!this.form.connection.stop_bits) this.form.connection.stop_bits = 1;
        if (!this.form.connection.slave_id) this.form.connection.slave_id = 1;
        if (!this.form.connection.byte_order) this.form.connection.byte_order = 'ABCD';
        this.fetchAvailableSerialPorts();
      } else if (p === 'MQTT') {
        this.form.connection.connection_type = 'ETHERNET';
        if (!this.form.connection.port || [502, 80, 8080, 8000].includes(this.form.connection.port)) {
          this.form.connection.port = 1883;
        }
      } else if (p === 'HTTP') {
        this.form.connection.connection_type = 'ETHERNET';
        if (!this.form.connection.port || [502, 1883, 8080, 8000].includes(this.form.connection.port)) {
          this.form.connection.port = 80;
        }
      } else if (p === 'TCP' || p === 'UDP') {
        this.form.connection.connection_type = 'ETHERNET';
        if (!this.form.connection.port || [502, 1883, 80].includes(this.form.connection.port)) {
          this.form.connection.port = 8000;
        }
      } else if (p === 'WEBSOCKET') {
        this.form.connection.connection_type = 'ETHERNET';
        if (!this.form.connection.port || [502, 1883, 80].includes(this.form.connection.port)) {
          this.form.connection.port = 8080;
        }
      } else if (p === 'SERIAL') {
        this.form.connection.connection_type = 'SERIAL';
        if (!this.form.connection.baud_rate) this.form.connection.baud_rate = 9600;
        if (!this.form.connection.data_bits) this.form.connection.data_bits = 8;
        if (!this.form.connection.parity) this.form.connection.parity = 'N';
        if (!this.form.connection.stop_bits) this.form.connection.stop_bits = 1;
        this.fetchAvailableSerialPorts();
      } else if (p === 'CUSTOM') {
        if (!this.form.connection.connection_type) {
          this.form.connection.connection_type = 'ETHERNET';
        }
      }
    },
    resetForm() {
      this.form = {
        device_code: '',
        device_name: '',
        device_type: 'MODBUS_TCP',
        manufacturer: '',
        model: '',
        serial_number: '',
        firmware_version: '',
        description: '',
        location: '',
        latitude: null,
        longitude: null,
        timezone: 'UTC',
        status: 'ACTIVE',
        enabled: true,
        connection: {
          protocol: 'MODBUS_TCP',
          connection_type: 'ETHERNET',
          host: '',
          port: 502,
          slave_id: 1,
          byte_order: 'ABCD',
          serial_port: 'COM1',
          baud_rate: 9600,
          data_bits: 8,
          parity: 'N',
          stop_bits: 1,
          timeout: 1000,
          retry_count: 3,
          polling_interval: 1000,
          enabled: true,
          extra_config: '',
          batch_read_enabled: false,
          max_batch_registers: 32,
          max_register_gap: 0,
        },
      };
    },
    scrollToError() {
      this.$nextTick(() => {
        if (this.$refs.deviceFormBody) {
          this.$refs.deviceFormBody.scrollTo({ top: 0, behavior: 'smooth' });
        }
      });
    },
    async submitForm() {
      this.error = null;

      // Validate Basic Info
      if (!this.form.device_code || !this.form.device_code.trim()) {
        this.activeTab = 'basic';
        this.error = this.$t('deviceModal.codeRequired');
        this.scrollToError();
        return;
      }
      if (!this.form.device_name || !this.form.device_name.trim()) {
        this.activeTab = 'basic';
        this.error = this.$t('deviceModal.nameRequired');
        this.scrollToError();
        return;
      }

      // Validate Connection Info
      if (this.isNetworkProtocol && (!this.form.connection.host || !this.form.connection.host.trim())) {
        this.activeTab = 'connection';
        this.error = this.$t('deviceModal.hostRequired');
        this.scrollToError();
        return;
      }
      if (this.isSerialProtocol && (!this.form.connection.serial_port || !this.form.connection.serial_port.trim())) {
        this.activeTab = 'connection';
        this.error = this.$t('deviceModal.portRequired');
        this.scrollToError();
        return;
      }

      // Synchronize device_type with connection.protocol for backward compatibility
      this.form.device_type = this.form.connection.protocol;

      this.loading = true;

      try {
        if (this.isEditing) {
          await this.$store.dispatch('updateDevice', {
            id: this.deviceToEdit.id,
            updates: this.form,
          });
          this.$emit('saved');
        } else {
          const payload = {
            ...this.form,
            source_device_id: (this.isCopying && this.sourceDevice) ? this.sourceDevice.id : undefined,
          };
          const createdDev = await this.$store.dispatch('createDevice', payload);
          this.$emit('saved', {
            ...(createdDev || {}),
            isCopy: this.isCopying,
            deviceCode: (createdDev && (createdDev.device_code || createdDev.code)) || this.form.device_code,
            deviceName: (createdDev && (createdDev.device_name || createdDev.name)) || this.form.device_name,
            id: createdDev ? createdDev.id : undefined,
          });
        }
        this.closeModal();
      } catch (err) {
        this.error = (err.response && err.response.data && err.response.data.error) || err.message || 'Operation failed';
        this.scrollToError();
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
