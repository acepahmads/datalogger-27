<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-3xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
        <div>
          <h2 class="text-base font-bold text-white tracking-wide">
            {{ isEditing ? 'Edit Industrial Device' : 'Register New Device' }}
          </h2>
          <p class="text-xs text-slate-400 mt-0.5">
            {{ isEditing ? 'Update device profile and communication parameters.' : 'Define physical sensor, meter, or gateway specifications and connection foundation.' }}
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
          <span>General Info</span>
        </button>
        <button
          type="button"
          @click="activeTab = 'location'"
          :class="activeTab === 'location' ? 'border-b-2 border-blue-500 text-blue-400 font-semibold' : 'text-slate-400 hover:text-slate-200'"
          class="py-3 px-3 transition-colors flex items-center space-x-1.5"
        >
          <span>Location & Timezone</span>
        </button>
        <button
          type="button"
          @click="activeTab = 'connection'"
          :class="activeTab === 'connection' ? 'border-b-2 border-blue-500 text-blue-400 font-semibold' : 'text-slate-400 hover:text-slate-200'"
          class="py-3 px-3 transition-colors flex items-center space-x-1.5"
        >
          <span>Communication & Protocol</span>
          <span class="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
        </button>
      </div>

      <!-- Modal Body -->
      <form @submit.prevent="submitForm" class="flex-1 overflow-y-auto p-6 space-y-5">
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
                Device Code <span class="text-rose-400">*</span>
              </label>
              <input
                v-model="form.device_code"
                type="text"
                required
                placeholder="e.g. PM-01, SHT-02"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Device Name -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Device Name <span class="text-rose-400">*</span>
              </label>
              <input
                v-model="form.device_name"
                type="text"
                required
                placeholder="e.g. Main Power Quality Meter"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <!-- Device Type -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Device Type <span class="text-rose-400">*</span>
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
                Admin Status
              </label>
              <select
                v-model="form.status"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white focus:outline-none focus:border-blue-500 font-mono"
              >
                <option value="ACTIVE">ACTIVE</option>
                <option value="INACTIVE">INACTIVE</option>
                <option value="MAINTENANCE">MAINTENANCE</option>
                <option value="DISABLED">DISABLED</option>
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
                  <span class="text-xs font-semibold text-slate-200">Device Enabled</span>
                  <span class="block text-3xs text-slate-400">Include in monitoring cycle</span>
                </div>
              </label>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Manufacturer -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Manufacturer
              </label>
              <input
                v-model="form.manufacturer"
                type="text"
                placeholder="e.g. Schneider Electric, Siemens"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>

            <!-- Model -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Model Number
              </label>
              <input
                v-model="form.model"
                type="text"
                placeholder="e.g. PM5560, SHT-35"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
              />
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <!-- Serial Number -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Serial Number
              </label>
              <input
                v-model="form.serial_number"
                type="text"
                placeholder="e.g. SN-9812-441"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>

            <!-- Firmware Version -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Firmware Version
              </label>
              <input
                v-model="form.firmware_version"
                type="text"
                placeholder="e.g. v2.1.0"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
          </div>

          <!-- Description -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              Description / Notes
            </label>
            <textarea
              v-model="form.description"
              rows="2"
              placeholder="Operational context, installation details, maintenance notes..."
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 resize-none"
            ></textarea>
          </div>
        </div>

        <!-- TAB 2: Location & Timezone -->
        <div v-show="activeTab === 'location'" class="space-y-4">
          <!-- Physical Location -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              Physical Location
            </label>
            <input
              v-model="form.location"
              type="text"
              placeholder="e.g. Building A - Switchgear Room 02, Rack 4"
              class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
            />
          </div>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <!-- Latitude -->
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Latitude (-90 to 90)
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
                Longitude (-180 to 180)
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
                Timezone
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
            <span class="font-medium">Protocol Adapter Architecture</span>
            <span class="text-3xs font-mono px-2 py-0.5 rounded bg-blue-500/20 text-blue-200">Phase 2.1 Configuration Foundation</span>
          </div>

          <!-- Protocol Selector -->
          <div>
            <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
              Communication Protocol <span class="text-rose-400">*</span>
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
                Host / IP / URL <span class="text-rose-400">*</span>
              </label>
              <input
                v-model="form.connection.host"
                type="text"
                placeholder="e.g. 192.168.1.100 or meter.local"
                class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
              />
            </div>
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Port
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
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Serial Port <span class="text-rose-400">*</span>
                </label>
                <input
                  v-model="form.connection.serial_port"
                  type="text"
                  placeholder="e.g. COM1, COM3, /dev/ttyUSB0"
                  class="w-full px-3 py-2 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500 font-mono"
                />
              </div>
              <div>
                <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                  Baud Rate
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
                  Data Bits
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
                  Parity
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
                  Stop Bits
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

          <!-- Common Timing Configuration -->
          <div class="grid grid-cols-1 md:grid-cols-3 gap-4 pt-2 border-t border-slate-800/80">
            <div>
              <label class="block text-2xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5">
                Timeout (ms)
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
                Retry Count
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
                Polling Interval (ms)
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
            Cancel
          </button>
          <div class="flex items-center space-x-2">
            <button
              v-if="activeTab !== 'basic'"
              type="button"
              @click="prevTab"
              class="px-4 py-2 rounded-xl text-xs font-medium text-slate-300 bg-slate-800 hover:bg-slate-700 transition"
            >
              Previous
            </button>
            <button
              v-if="activeTab !== 'connection'"
              type="button"
              @click="nextTab"
              class="px-4 py-2 rounded-xl text-xs font-semibold text-white bg-blue-600 hover:bg-blue-500 transition"
            >
              Next
            </button>
            <button
              v-if="activeTab === 'connection'"
              type="submit"
              :disabled="loading"
              class="px-5 py-2 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-500 transition flex items-center space-x-1.5 disabled:opacity-50"
            >
              <span v-if="loading" class="w-3.5 h-3.5 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
              <span>{{ isEditing ? 'Save Changes' : 'Register Device' }}</span>
            </button>
          </div>
        </div>
      </form>
    </div>
  </div>
</template>

<script>
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
  },
  watch: {
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
  methods: {
    closeModal() {
      this.error = null;
      this.activeTab = 'basic';
      this.$emit('close');
    },
    nextTab() {
      if (this.activeTab === 'basic') this.activeTab = 'location';
      else if (this.activeTab === 'location') this.activeTab = 'connection';
    },
    prevTab() {
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
      this.loading = true;
      this.error = null;

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
