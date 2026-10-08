<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-3xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
        <div>
          <h2 class="text-base font-bold text-white tracking-wide">
            {{ isEditing ? $t('deviceModal.titleEdit') : $t('deviceModal.titleAdd') }}
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            {{ isEditing ? $t('deviceModal.subtitleEdit') : $t('deviceModal.subtitleAdd') }}
          </p>
        </div>
        <button @click="closeModal" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Navigation Tabs -->
      <div class="flex border-b border-slate-800 bg-[#0B0F19]/40 px-6 text-xs font-medium">
        <button
          type="button"
          @click="activeTab = 'basic'"
          :class="activeTab === 'basic' ? 'border-b-2 border-blue-500 text-blue-400 font-semibold' : 'text-slate-400 hover:text-slate-200'"
          class="py-3 px-3 transition-colors flex items-center space-x-1.5"
        >
          <span>{{ $t('deviceModal.tabGeneral') }}</span>
        </button>
        <button
          type="button"
          @click="activeTab = 'location'"
          :class="activeTab === 'location' ? 'border-b-2 border-blue-500 text-blue-400 font-semibold' : 'text-slate-400 hover:text-slate-200'"
          class="py-3 px-3 transition-colors flex items-center space-x-1.5"
        >
          <span>{{ $t('deviceModal.tabLocation') }}</span>
        </button>
        <button
          type="button"
          @click="activeTab = 'connection'"
          :class="activeTab === 'connection' ? 'border-b-2 border-blue-500 text-blue-400 font-semibold' : 'text-slate-400 hover:text-slate-200'"
          class="py-3 px-3 transition-colors flex items-center space-x-1.5"
        >
          <span>{{ $t('deviceModal.tabComm') }}</span>
          <span class="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
        </button>
      </div>

      <!-- Modal Body -->
      <form novalidate @submit.prevent="submitForm" class="flex-1 overflow-y-auto p-6 space-y-5">
        <!-- Error Alert -->
        <div v-if="error" class="p-3 bg-rose-500/10 border border-rose-500/30 rounded-xl text-xs text-rose-400 flex items-center space-x-2">
          <svg class="w-4 h-4 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>{{ error }}</span>
        </div>

        <!-- TAB 1: General Info -->
        <div v-show="activeTab === 'basic'" class="space-y-4">
          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Device Code -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.deviceCode') }}
              </label>
              <input
                v-model="form.device_code"
                type="text"
                :placeholder="$t('deviceModal.placeholders.code')"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Device Name -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.deviceName') }}
              </label>
              <input
                v-model="form.device_name"
                type="text"
                :placeholder="$t('deviceModal.placeholders.name')"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <!-- Device Type -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.deviceType') }}
              </label>
              <select
                v-model="form.device_type"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500"
              >
                <option value="MODBUS_TCP">Modbus TCP Meter</option>
                <option value="MODBUS_RTU">Modbus RTU Sensor</option>
                <option value="MQTT">MQTT Gateway</option>
                <option value="HTTP">HTTP/REST Sensor</option>
                <option value="TCP">Raw TCP Socket</option>
                <option value="UDP">UDP Broadcast</option>
                <option value="SERIAL">Serial RS485/RS232</option>
                <option value="WEBSOCKET">WebSocket Stream</option>
                <option value="CUSTOM">Custom Device</option>
              </select>
            </div>

            <!-- Administrative Status -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.adminStatus') }}
              </label>
              <select
                v-model="form.status"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
              >
                <option value="ACTIVE">{{ $t('status.active') }}</option>
                <option value="INACTIVE">{{ $t('status.inactive') }}</option>
                <option value="MAINTENANCE">{{ $t('status.maintenance') }}</option>
                <option value="DISABLED">{{ $t('status.disabled') }}</option>
              </select>
            </div>

            <!-- Enabled Switch -->
            <div class="flex flex-col justify-end">
              <label class="flex items-center space-x-3 cursor-pointer p-2 rounded-xl bg-[#0B0F19] border border-slate-800">
                <input
                  v-model="form.enabled"
                  type="checkbox"
                  class="w-4 h-4 rounded text-blue-600 focus:ring-0 bg-slate-900 border-slate-700"
                />
                <div>
                  <span class="text-xs font-semibold text-slate-200">{{ $t('deviceModal.deviceEnabled') }}</span>
                  <span class="block text-3xs text-slate-400">{{ $t('deviceModal.includeInCycle') }}</span>
                </div>
              </label>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Manufacturer -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.manufacturer') }}
              </label>
              <input
                v-model="form.manufacturer"
                type="text"
                :placeholder="$t('deviceModal.placeholders.manufacturer')"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>

            <!-- Model -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.model') }}
              </label>
              <input
                v-model="form.model"
                type="text"
                :placeholder="$t('deviceModal.placeholders.model')"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Serial Number -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.serialNumber') }}
              </label>
              <input
                v-model="form.serial_number"
                type="text"
                :placeholder="$t('deviceModal.placeholders.serialNumber')"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Firmware Version -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.firmwareVersion') }}
              </label>
              <input
                v-model="form.firmware_version"
                type="text"
                :placeholder="$t('deviceModal.placeholders.firmware')"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
          </div>

          <!-- Description -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('deviceModal.description') }}
            </label>
            <textarea
              v-model="form.description"
              rows="2"
              :placeholder="$t('deviceModal.placeholders.description')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 resize-none"
            ></textarea>
          </div>
        </div>

        <!-- TAB 2: Location & Timezone -->
        <div v-show="activeTab === 'location'" class="space-y-4">
          <!-- Physical Location -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('deviceModal.location') }}
            </label>
            <input
              v-model="form.location"
              type="text"
              :placeholder="$t('deviceModal.placeholders.location')"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <!-- Latitude -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.latitude') }}
              </label>
              <input
                v-model.number="form.latitude"
                type="number"
                step="0.000001"
                min="-90"
                max="90"
                placeholder="-6.2088"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Longitude -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.longitude') }}
              </label>
              <input
                v-model.number="form.longitude"
                type="number"
                step="0.000001"
                min="-180"
                max="180"
                placeholder="106.8456"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Timezone -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.timezone') }}
              </label>
              <select
                v-model="form.timezone"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
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
          <div class="p-3 bg-blue-500/10 border border-blue-500/20 rounded-xl flex items-center justify-between text-xs text-blue-300">
            <span class="font-medium">{{ $t('deviceModal.protocolArchitecture') }}</span>
            <span class="text-3xs font-mono px-2 py-0.5 rounded bg-blue-500/20 text-blue-200">{{ $t('deviceModal.phase21Foundation') }}</span>
          </div>

          <!-- Protocol Selector -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              {{ $t('deviceModal.protocol') }}
            </label>
            <select
              v-model="form.connection.protocol"
              @change="onProtocolChange"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
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
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.hostIp') }}
              </label>
              <input
                v-model="form.connection.host"
                type="text"
                :placeholder="$t('deviceModal.placeholders.host')"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.port') }}
              </label>
              <input
                v-model.number="form.connection.port"
                type="number"
                placeholder="502"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
          </div>

          <!-- 2. For Serial (MODBUS_RTU, SERIAL) -->
          <div v-if="isSerialProtocol" class="space-y-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <div class="flex items-center justify-between mb-1.5">
                  <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider">
                    {{ $t('deviceModal.hardwareAddress') }}
                  </label>
                  <button
                    type="button"
                    @click="fetchAvailableSerialPorts"
                    :disabled="loadingPorts"
                    class="text-3xs text-blue-400 hover:text-blue-300 flex items-center space-x-1"
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
                      class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
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
                    class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
                  />

                  <!-- Field stability alert / recommendation tip -->
                  <div class="p-2.5 rounded-lg bg-blue-500/10 border border-blue-500/20 text-[10px] text-blue-300 font-sans flex items-start space-x-2 mt-1.5">
                    <svg class="w-3.5 h-3.5 text-blue-400 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd"></path>
                    </svg>
                    <div class="space-y-0.5">
                      <span class="font-semibold text-blue-200">{{ $t('deviceModal.resilienceTipTitle') }}</span>
                      <p class="text-slate-300">
                        {{ $t('deviceModal.resilienceTipText') }}
                      </p>
                    </div>
                  </div>

                  <div class="flex items-center justify-between text-3xs text-slate-400 pt-0.5">
                    <span v-if="systemPorts.length > 0 && !useCustomPort" class="text-emerald-400 font-mono flex items-center space-x-1">
                      <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                      <span>{{ $t('deviceModal.pathsDetected', { count: systemPorts.length }) }}</span>
                    </span>
                    <span v-else-if="!useCustomPort" class="text-slate-500">
                      {{ $t('deviceModal.noPortsDetected') }}
                    </span>
                    <span v-else class="text-slate-400">
                      {{ $t('deviceModal.customPortMode') }}
                    </span>
                    <button
                      type="button"
                      @click="useCustomPort = !useCustomPort"
                      class="text-blue-400 hover:text-blue-300 underline ml-auto text-3xs"
                    >
                      {{ useCustomPort ? $t('deviceModal.chooseFromList') : $t('deviceModal.typeCustomPort') }}
                    </button>
                  </div>
                </div>
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  {{ $t('deviceModal.baudRate') }}
                </label>
                <select
                  v-model.number="form.connection.baud_rate"
                  class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
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
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  {{ $t('deviceModal.dataBits') }}
                </label>
                <select
                  v-model.number="form.connection.data_bits"
                  class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
                >
                  <option :value="8">8 (Standard)</option>
                  <option :value="7">7</option>
                </select>
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  {{ $t('deviceModal.parity') }}
                </label>
                <select
                  v-model="form.connection.parity"
                  class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
                >
                  <option value="N">None (N)</option>
                  <option value="E">Even (E)</option>
                  <option value="O">Odd (O)</option>
                </select>
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  {{ $t('deviceModal.stopBits') }}
                </label>
                <select
                  v-model.number="form.connection.stop_bits"
                  class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
                >
                  <option :value="1">1 (Standard)</option>
                  <option :value="2">2</option>
                </select>
              </div>
            </div>
          </div>

          <!-- Modbus Specific Settings (for MODBUS_TCP, MODBUS_RTU) -->
          <div v-if="form.connection.protocol === 'MODBUS_TCP' || form.connection.protocol === 'MODBUS_RTU'" class="grid grid-cols-1 md:grid-cols-2 gap-4 p-3 bg-slate-900/60 rounded-xl border border-slate-800">
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.slaveId') }}
              </label>
              <input
                v-model.number="form.connection.slave_id"
                type="number"
                min="1"
                max="247"
                placeholder="1"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
              <span class="text-3xs text-slate-400 mt-1 block">{{ $t('deviceModal.slaveAddressHint') }}</span>
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.byteOrder') }}
              </label>
              <select
                v-model="form.connection.byte_order"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-blue-500"
              >
                <option value="ABCD">ABCD — Big Endian (Standard)</option>
                <option value="CDAB">CDAB — Word Swap (Modicon / ABB)</option>
                <option value="BADC">BADC — Byte Swap</option>
                <option value="DCBA">DCBA — Little Endian</option>
              </select>
              <span class="text-3xs text-slate-400 mt-1 block">{{ $t('deviceModal.byteOrderHint') }}</span>
            </div>
          </div>

          <!-- Common Timing Configuration -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2 border-t border-slate-800/80">
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.timeout') }}
              </label>
              <input
                v-model.number="form.connection.timeout"
                type="number"
                min="50"
                step="50"
                placeholder="1000"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.retryCount') }}
              </label>
              <input
                v-model.number="form.connection.retry_count"
                type="number"
                min="0"
                max="10"
                placeholder="3"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                {{ $t('deviceModal.pollInterval') }}
              </label>
              <input
                v-model.number="form.connection.polling_interval"
                type="number"
                min="100"
                step="100"
                placeholder="1000"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
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
          <div class="flex items-center space-x-2">
            <button
              v-if="activeTab !== 'basic'"
              type="button"
              @click="prevTab"
              class="px-4 py-2 rounded-xl text-xs font-medium text-slate-300 bg-slate-800 hover:bg-slate-700 transition"
            >
              {{ $t('common.previous') }}
            </button>
            <button
              v-if="activeTab !== 'connection'"
              type="button"
              @click="nextTab"
              class="px-4 py-2 rounded-xl text-xs font-semibold text-white bg-blue-600 hover:bg-blue-500 transition"
            >
              {{ $t('common.next') }}
            </button>
            <button
              v-if="activeTab === 'connection'"
              type="submit"
              :disabled="loading"
              class="px-5 py-2 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-500 transition flex items-center space-x-1.5 disabled:opacity-50"
            >
              <span v-if="loading" class="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
              <span>{{ isEditing ? $t('common.saveChanges') : $t('deviceModal.registerDevice') }}</span>
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
        },
      },
    };
  },
  computed: {
    isEditing() {
      return !!this.deviceToEdit;
    },
    isNetworkProtocol() {
      const p = this.form.connection.protocol;
      return ['MODBUS_TCP', 'TCP', 'UDP', 'HTTP', 'MQTT', 'WEBSOCKET'].includes(p);
    },
    isSerialProtocol() {
      const p = this.form.connection.protocol;
      return ['MODBUS_RTU', 'SERIAL'].includes(p);
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
        }
      },
    },
    deviceToEdit: {
      immediate: true,
      handler(newVal) {
        if (newVal) {
          this.form = {
            device_code: newVal.device_code || newVal.code || '',
            device_name: newVal.device_name || newVal.name || '',
            device_type: newVal.device_type || 'MODBUS_TCP',
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
              protocol: (newVal.connection && newVal.connection.protocol) || 'MODBUS_TCP',
              connection_type: (newVal.connection && newVal.connection.connection_type) || 'ETHERNET',
              host: (newVal.connection && (newVal.connection.host || newVal.connection.address)) || '',
              port: (newVal.connection && newVal.connection.port) || 502,
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
      if (p === 'MODBUS_TCP') {
        this.form.connection.port = 502;
        this.form.connection.connection_type = 'ETHERNET';
      } else if (p === 'MODBUS_RTU') {
        this.form.connection.connection_type = 'SERIAL';
        this.form.connection.baud_rate = 9600;
        this.form.connection.data_bits = 8;
        this.form.connection.parity = 'N';
        this.form.connection.stop_bits = 1;
        this.fetchAvailableSerialPorts();
      } else if (p === 'MQTT') {
        this.form.connection.port = 1883;
        this.form.connection.connection_type = 'ETHERNET';
      } else if (p === 'HTTP') {
        this.form.connection.port = 80;
        this.form.connection.connection_type = 'ETHERNET';
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
        },
      };
    },
    async submitForm() {
      this.error = null;

      // Validate Basic Info
      if (!this.form.device_code || !this.form.device_code.trim()) {
        this.activeTab = 'basic';
        this.error = 'Device Code is required.';
        return;
      }
      if (!this.form.device_name || !this.form.device_name.trim()) {
        this.activeTab = 'basic';
        this.error = 'Device Name is required.';
        return;
      }

      // Validate Connection Info
      if (this.isNetworkProtocol && (!this.form.connection.host || !this.form.connection.host.trim())) {
        this.activeTab = 'connection';
        this.error = 'Host / IP Address is required for network protocols.';
        return;
      }
      if (this.isSerialProtocol && (!this.form.connection.serial_port || !this.form.connection.serial_port.trim())) {
        this.activeTab = 'connection';
        this.error = 'Serial Port (e.g. COM1, COM3, or /dev/ttyUSB0) is required.';
        return;
      }

      this.loading = true;

      try {
        if (this.isEditing) {
          await this.$store.dispatch('updateDevice', {
            id: this.deviceToEdit.id,
            updates: this.form,
          });
        } else {
          await this.$store.dispatch('createDevice', this.form);
        }
        this.$emit('saved');
        this.closeModal();
      } catch (err) {
        this.error = (err.response && err.response.data && err.response.data.error) || err.message || 'Operation failed';
      } finally {
        this.loading = false;
      }
    },
  },
};
</script>
