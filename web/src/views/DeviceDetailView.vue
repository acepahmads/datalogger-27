<template>
  <div class="space-y-6 font-sans">
    <!-- Top Bar Navigation & Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <div class="flex items-center space-x-2 text-xs text-slate-400 mb-1">
          <router-link to="/monitoring/devices" class="hover:text-blue-400 flex items-center space-x-1">
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
            </svg>
            <span>Devices & Sensors</span>
          </router-link>
          <span>/</span>
          <span class="text-slate-200 font-mono font-semibold">{{ device ? (device.device_code || device.code) : 'Device' }}</span>
        </div>
        <div class="flex items-center space-x-3">
          <h1 class="text-xl font-bold tracking-tight text-white">
            {{ device ? (device.device_name || device.name) : 'Loading Device...' }}
          </h1>
          <span v-if="device" class="px-2.5 py-0.5 rounded-md bg-blue-500/10 text-blue-400 border border-blue-500/20 font-mono text-2xs font-bold">
            {{ device.device_code || device.code }}
          </span>
        </div>
      </div>

      <!-- Quick Actions -->
      <div v-if="device" class="flex items-center space-x-2.5">
        <button
          @click="toggleEnabled"
          :class="device.enabled ? 'bg-amber-500/10 text-amber-400 border-amber-500/20 hover:bg-amber-500/20' : 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20 hover:bg-emerald-500/20'"
          class="px-3 py-1.5 rounded-xl border text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <span class="w-2 h-2 rounded-full" :class="device.enabled ? 'bg-amber-400' : 'bg-emerald-400'"></span>
          <span>{{ device.enabled ? 'Disable Device' : 'Enable Device' }}</span>
        </button>

        <button
          @click="openEditModal"
          class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
          </svg>
          <span>Edit</span>
        </button>

        <button
          @click="confirmDelete"
          class="px-3 py-1.5 rounded-xl bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 border border-rose-500/20 text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
          </svg>
          <span>Delete</span>
        </button>
      </div>
    </div>

    <!-- Status Cards Header -->
    <div v-if="device" class="grid grid-cols-2 md:grid-cols-5 gap-3">
      <!-- Admin Status -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">Administrative Status</div>
        <div class="flex items-center space-x-2">
          <span class="px-2 py-0.5 rounded-md text-2xs font-bold font-mono" :class="statusBadgeClass(device.status)">
            {{ device.status }}
          </span>
        </div>
      </div>

      <!-- Comm Status -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">Communication Status</div>
        <div class="flex items-center space-x-1.5">
          <span class="w-2 h-2 rounded-full" :class="commDotClass(device.connection_status)"></span>
          <span class="text-xs font-mono font-bold" :class="commTextClass(device.connection_status)">
            {{ device.connection_status || 'UNKNOWN' }}
          </span>
        </div>
      </div>

      <!-- Protocol -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">Protocol</div>
        <div class="text-xs font-mono font-bold text-blue-400">
          {{ (device.connection && device.connection.protocol) || device.device_type || 'MODBUS_TCP' }}
        </div>
      </div>

      <!-- Last Seen -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">Last Communication</div>
        <div class="text-xs font-mono text-slate-200">
          {{ formatTimestamp(device.last_seen_at || device.last_communication) }}
        </div>
      </div>

      <!-- Last Data -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">Last Data Packet</div>
        <div class="text-xs font-mono text-slate-200">
          {{ formatTimestamp(device.last_data_at) }}
        </div>
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div v-if="device" class="border-b border-slate-800 flex space-x-6 text-xs font-semibold">
      <button
        @click="activeTab = 'overview'"
        :class="activeTab === 'overview' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>Overview</span>
      </button>
      <button
        @click="activeTab = 'connection'"
        :class="activeTab === 'connection' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>Connection Config</span>
      </button>
      <button
        @click="activeTab = 'parameters'"
        :class="activeTab === 'parameters' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>Parameters</span>
        <span class="px-1.5 py-0.2 rounded-full bg-slate-800 text-slate-300 text-3xs font-mono">
          {{ device.parameters ? device.parameters.length : 0 }}
        </span>
      </button>
      <button
        @click="switchTab('telemetry')"
        :class="activeTab === 'telemetry' ? 'border-b-2 border-emerald-500 text-emerald-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span class="w-2 h-2 rounded-full bg-emerald-400" :class="{ 'animate-pulse': livePacketBlink }"></span>
        <span>Live Telemetry</span>
      </button>
      <button
        @click="switchTab('history')"
        :class="activeTab === 'history' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>Historical Data</span>
      </button>
      <button
        @click="switchTab('raw')"
        :class="activeTab === 'raw' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>Raw Telemetry</span>
      </button>
      <button
        @click="activeTab = 'activity'"
        :class="activeTab === 'activity' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>Activity & Audit</span>
      </button>
      <button
        @click="activeTab = 'health'"
        :class="activeTab === 'health' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>Health Diagnostics</span>
      </button>
    </div>

    <!-- TAB 1: OVERVIEW -->
    <div v-if="device && activeTab === 'overview'" class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Hardware Profile -->
      <div class="saas-card p-5 space-y-4">
        <h3 class="text-xs font-bold text-white uppercase tracking-wider border-b border-slate-800/80 pb-2.5">
          Equipment Profile
        </h3>
        <dl class="grid grid-cols-2 gap-3 text-xs">
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Manufacturer</dt>
            <dd class="text-slate-200 font-medium mt-0.5">{{ device.manufacturer || '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Model</dt>
            <dd class="text-slate-200 font-medium mt-0.5">{{ device.model || '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Serial Number</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.serial_number || '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Firmware Version</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.firmware_version || '--' }}</dd>
          </div>
          <div class="col-span-2">
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Description</dt>
            <dd class="text-slate-300 mt-0.5">{{ device.description || 'No description provided.' }}</dd>
          </div>
        </dl>
      </div>

      <!-- Location & Metadata -->
      <div class="saas-card p-5 space-y-4">
        <h3 class="text-xs font-bold text-white uppercase tracking-wider border-b border-slate-800/80 pb-2.5">
          Location & Geography
        </h3>
        <dl class="grid grid-cols-2 gap-3 text-xs">
          <div class="col-span-2">
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Physical Location</dt>
            <dd class="text-slate-200 font-medium mt-0.5">{{ device.location || 'Not specified' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Latitude</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.latitude !== null && device.latitude !== undefined ? device.latitude : '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Longitude</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.longitude !== null && device.longitude !== undefined ? device.longitude : '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Timezone</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.timezone || 'UTC' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">Created At</dt>
            <dd class="text-slate-200 font-mono text-3xs mt-0.5">{{ formatTimestamp(device.created_at) }}</dd>
          </div>
        </dl>
      </div>
    </div>

    <!-- TAB 2: CONNECTION CONFIG -->
    <div v-if="device && activeTab === 'connection'" class="saas-card p-6 space-y-6">
      <!-- Header & Actions -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-4">
        <div>
          <div class="flex items-center space-x-2">
            <h3 class="text-sm font-bold text-white">Modbus RTU / TCP Communication Engine</h3>
            <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
              Phase 2.2 Active
            </span>
            <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              Auto-Connect Active
            </span>
          </div>
          <p class="text-2xs text-slate-400 mt-0.5">Physical serial & socket lifecycle, auto-reconnect backoff, and continuous background register polling</p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <!-- Force Reconnect (Manual Override) -->
          <button
            @click="handleReconnect"
            :disabled="commActionLoading"
            class="px-3 py-1.5 rounded-xl bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 border border-blue-500/30 text-xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
            title="Sistem sudah otomatis menyambung di background. Tombol ini untuk memicu penyambungan ulang instan."
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>Reconnect Now</span>
          </button>

          <!-- Test Link Button -->
          <button
            @click="handleTestConnection"
            :disabled="commActionLoading"
            class="px-3 py-1.5 rounded-xl bg-amber-500/10 hover:bg-amber-500/20 text-amber-400 border border-amber-500/20 text-xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span>Test Link</span>
          </button>

          <!-- Edit Config -->
          <button
            @click="openEditModal"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 text-xs font-semibold transition"
          >
            Edit Settings
          </button>
        </div>
      </div>

      <!-- Auto-Connect & Background Polling Explanatory Banner -->
      <div
        class="p-3.5 rounded-xl border flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs"
        :class="((commStatus && commStatus.status) || device.connection_status) === 'ONLINE' ? 'bg-emerald-500/10 border-emerald-500/20 text-emerald-300' : 'bg-amber-500/10 border-amber-500/20 text-amber-300'"
      >
        <div class="flex items-center space-x-2.5">
          <span
            class="w-2.5 h-2.5 rounded-full flex-shrink-0"
            :class="((commStatus && commStatus.status) || device.connection_status) === 'ONLINE' ? 'bg-emerald-400 animate-pulse' : 'bg-amber-400 animate-ping'"
          ></span>
          <div>
            <div class="font-bold flex items-center space-x-2">
              <span>Mode: Auto-Connect &amp; Auto-Polling Aktif (Background Loop)</span>
              <span class="px-1.5 py-0.5 rounded text-3xs font-mono font-medium uppercase bg-slate-800 text-slate-300">
                Tidak Perlu Connect Manual
              </span>
            </div>
            <p class="text-3xs opacity-85 mt-0.5">
              {{ ((commStatus && commStatus.status) || device.connection_status) === 'ONLINE' ? 'Komunikasi serial/network terhubung normal. Data telemetry diambil setiap detik dan dicatat ke log.' : 'Port serial/network terputus atau offline. Background engine terus mencoba menyambung ulang otomatis secara berkala (Auto-Reconnect). Semua kegagalan dicatat ke System Log & Warning.' }}
            </p>
          </div>
        </div>
        <div class="flex items-center space-x-3 text-3xs font-mono text-slate-300 flex-shrink-0">
          <span>Interval: {{ (device.connection && device.connection.polling_interval) || 1000 }}ms</span>
          <span>•</span>
          <span>Auto-Recovery: ON</span>
        </div>
      </div>

      <!-- Action Feedback Banner -->
      <div
        v-if="commFeedback"
        :class="commFeedback.type === 'success' ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-300' : 'bg-rose-500/10 border-rose-500/30 text-rose-300'"
        class="p-3 rounded-xl border text-xs flex items-center justify-between"
      >
        <div class="flex items-center space-x-2">
          <span class="w-2 h-2 rounded-full" :class="commFeedback.type === 'success' ? 'bg-emerald-400' : 'bg-rose-400'"></span>
          <span>{{ commFeedback.message }}</span>
        </div>
        <button @click="commFeedback = null" class="text-slate-400 hover:text-white text-3xs font-mono">Dismiss</button>
      </div>

      <!-- Live Engine Diagnostics Strip -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-1">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Engine Lifecycle State</div>
          <div class="flex items-center space-x-2">
            <span class="w-2 h-2 rounded-full" :class="commDotClass(commStatus ? commStatus.status : device.connection_status)"></span>
            <span class="text-xs font-mono font-bold" :class="commTextClass(commStatus ? commStatus.status : device.connection_status)">
              {{ (commStatus && commStatus.status) || device.connection_status || 'UNKNOWN' }}
            </span>
          </div>
        </div>

        <div class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-1">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Roundtrip Latency</div>
          <div class="text-xs font-mono font-bold text-emerald-400">
            {{ commStatus ? commStatus.latency_ms : (device.latency_ms || 0) }} ms
          </div>
        </div>

        <div class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-1">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Consecutive Retries</div>
          <div class="text-xs font-mono font-bold text-slate-200">
            {{ commStatus ? commStatus.consecutive_errors : 0 }} / {{ (device.connection && device.connection.retry_count) || 3 }} max
          </div>
        </div>

        <div class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-1">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Last Comm Error</div>
          <div class="text-xs font-mono truncate text-rose-400" :title="(commStatus && commStatus.last_error) || 'None'">
            {{ (commStatus && commStatus.last_error) || 'None' }}
          </div>
        </div>
      </div>

      <!-- Detailed Configuration Grid -->
      <div v-if="device.connection" class="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs font-sans">
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Protocol</div>
          <div class="text-xs font-mono font-bold text-blue-400 mt-1">{{ device.connection.protocol }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Connection Type</div>
          <div class="text-xs font-mono text-slate-200 mt-1">{{ device.connection.connection_type }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Modbus Slave ID (Unit)</div>
          <div class="text-xs font-mono font-bold text-amber-400 mt-1">#{{ device.connection.slave_id || 1 }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Byte / Word Order</div>
          <div class="text-xs font-mono font-bold text-indigo-400 mt-1">{{ device.connection.byte_order || 'ABCD' }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Host / Address</div>
          <div class="text-xs font-mono text-slate-200 mt-1">{{ device.connection.host || device.connection.address || '--' }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Port</div>
          <div class="text-xs font-mono text-slate-200 mt-1">{{ device.connection.port || '--' }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Serial Port</div>
          <div class="text-xs font-mono text-slate-200 mt-1">{{ device.connection.serial_port || '--' }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Baud Rate & Framing</div>
          <div class="text-xs font-mono text-slate-200 mt-1">
            {{ device.connection.baud_rate || 9600 }} ({{ device.connection.data_bits || 8 }}-{{ device.connection.parity || 'N' }}-{{ device.connection.stop_bits || 1 }})
          </div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Timeout</div>
          <div class="text-xs font-mono text-slate-200 mt-1">
            {{ device.connection.timeout || device.connection.timeout_ms || 1000 }} ms
          </div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Retry Policy</div>
          <div class="text-xs font-mono text-slate-200 mt-1">
            {{ device.connection.retry_count || 3 }} retries with exponential backoff
          </div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Polling Interval</div>
          <div class="text-xs font-mono text-slate-200 mt-1">
            {{ device.connection.polling_interval || device.connection.poll_interval_ms || 1000 }} ms
          </div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">Adapter Status</div>
          <div class="text-xs font-mono text-emerald-400 mt-1">
            {{ (commStatus && commStatus.adapter_status) || 'ACTIVE' }}
          </div>
        </div>
      </div>
      <div v-else class="text-center py-8 text-xs text-slate-400">
        No connection configuration found for this device.
      </div>
    </div>

    <!-- TAB 3: PARAMETERS -->
    <div v-if="device && activeTab === 'parameters'" class="saas-card p-5 space-y-4">
      <div class="flex items-center justify-between border-b border-slate-800 pb-3">
        <div>
          <h3 class="text-sm font-bold text-white">Telemetry Parameters ({{ device.parameters ? device.parameters.length : 0 }})</h3>
          <p class="text-2xs text-slate-400 mt-0.5">Engineering parameter definitions, scaling factor, and registers</p>
        </div>
        <button
          @click="openAddParamModal"
          class="px-3.5 py-1.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
          </svg>
          <span>Add Parameter</span>
        </button>
      </div>

      <!-- Parameters Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs font-sans">
          <thead>
            <tr class="border-b border-slate-800 text-3xs uppercase font-semibold text-slate-400 tracking-wider">
              <th class="py-2.5 px-3">Code</th>
              <th class="py-2.5 px-3">Name</th>
              <th class="py-2.5 px-3">Data Type</th>
              <th class="py-2.5 px-3">Unit</th>
              <th class="py-2.5 px-3">Register</th>
              <th class="py-2.5 px-3">Scale / Offset</th>
              <th class="py-2.5 px-3">Limits</th>
              <th class="py-2.5 px-3">Current Value</th>
              <th class="py-2.5 px-3">Enabled</th>
              <th class="py-2.5 px-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60">
            <tr v-for="param in device.parameters" :key="param.id" class="hover:bg-slate-800/30 transition-colors">
              <td class="py-2.5 px-3 font-mono font-bold text-blue-400">{{ param.parameter_code || param.code }}</td>
              <td class="py-2.5 px-3 font-medium text-slate-200">{{ param.parameter_name || param.name }}</td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-300">{{ param.data_type }}</td>
              <td class="py-2.5 px-3 font-medium text-slate-400">{{ param.unit || '--' }}</td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-400">#{{ param.register_address }}</td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-400">×{{ param.scale || param.scale_factor || 1.0 }} + {{ param.offset || 0 }}</td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-400">
                {{ param.min_value !== null && param.min_value !== undefined ? param.min_value : (param.low_limit || '-∞') }} ..
                {{ param.max_value !== null && param.max_value !== undefined ? param.max_value : (param.high_limit || '+∞') }}
              </td>
              <td class="py-2.5 px-3 font-mono font-bold text-slate-200">
                {{ param.current_value !== null && param.current_value !== undefined ? param.current_value.toFixed(param.precision || 2) : '--' }}
              </td>
              <td class="py-2.5 px-3">
                <button
                  @click="toggleParamEnabled(param)"
                  :class="param.enabled ? 'text-emerald-400 hover:text-emerald-300' : 'text-slate-500 hover:text-slate-400'"
                  class="font-mono text-3xs font-semibold transition"
                >
                  {{ param.enabled ? 'Active' : 'Disabled' }}
                </button>
              </td>
              <td class="py-2.5 px-3 text-right space-x-2">
                <button
                  @click="openTestReadModal(param)"
                  class="text-emerald-400 hover:text-emerald-300 transition text-2xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/20"
                >
                  Test Read
                </button>
                <button @click="openEditParamModal(param)" class="text-blue-400 hover:text-blue-300 transition text-2xs font-semibold">
                  Edit
                </button>
                <button @click="confirmDeleteParam(param)" class="text-rose-400 hover:text-rose-300 transition text-2xs font-semibold">
                  Delete
                </button>
              </td>
            </tr>
            <tr v-if="!device.parameters || device.parameters.length === 0">
              <td colspan="10" class="py-8 text-center text-xs text-slate-400">
                No parameters configured yet. Click "Add Parameter" to configure your first telemetry register.
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- TAB 4: ACTIVITY & AUDIT -->
    <div v-if="device && activeTab === 'activity'" class="saas-card p-5 space-y-4">
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800 pb-3">
        <div>
          <div class="flex items-center space-x-2">
            <h3 class="text-sm font-bold text-white">Device Activity &amp; Audit Log</h3>
            <span class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
              Live Audit Stream
            </span>
          </div>
          <p class="text-2xs text-slate-400 mt-0.5">Riwayat lengkap perubahan konfigurasi, event komunikasi, dan status operasional perangkat</p>
        </div>

        <div class="flex items-center space-x-3 text-xs">
          <span class="text-3xs text-slate-400 font-mono flex items-center gap-1.5">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
            Auto-Sync 3s
          </span>
          <button
            @click="loadActivity"
            :disabled="activityLoading"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-2xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': activityLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>Segarkan Log</span>
          </button>
        </div>
      </div>

      <div class="space-y-3">
        <div
          v-for="trail in activities"
          :key="trail.id"
          class="p-4 bg-[#0B0F19] rounded-xl border transition space-y-2.5 text-xs shadow-sm"
          :class="getAuditCardBorder(trail.action)"
        >
          <!-- Top Row: Action Badge, Operator, Timestamp -->
          <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-1.5 border-b border-slate-800/60 pb-2">
            <div class="flex items-center space-x-2.5">
              <span
                class="px-2.5 py-0.5 rounded-lg text-3xs font-mono font-bold flex items-center space-x-1"
                :class="getAuditActionClass(trail.action)"
              >
                <span>{{ formatAuditActionName(trail.action) }}</span>
              </span>
              <span class="text-slate-400 text-2xs">
                Oleh: <strong class="text-slate-200 font-semibold">{{ trail.username || 'System' }}</strong>
              </span>
              <span v-if="trail.ip_address" class="text-3xs text-slate-500 font-mono">
                ({{ trail.ip_address }})
              </span>
            </div>
            <span class="text-slate-400 font-mono text-3xs flex items-center space-x-1.5">
              <svg class="w-3 h-3 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"></path>
              </svg>
              <span>{{ formatTimestamp(trail.created_at) }}</span>
            </span>
          </div>

          <!-- Human-Readable Summary -->
          <div class="space-y-2">
            <p class="text-slate-200 font-medium text-xs leading-relaxed flex items-start space-x-2">
              <span class="w-1.5 h-1.5 rounded-full mt-1.5 flex-shrink-0" :class="trail.action.includes('ERROR') ? 'bg-rose-400' : (trail.action.includes('RESTORED') ? 'bg-emerald-400' : 'bg-blue-400')"></span>
              <span>{{ formatAuditDescription(trail) }}</span>
            </p>

            <!-- Attribute Chips -->
            <div v-if="getAuditChips(trail).length > 0" class="flex flex-wrap gap-1.5 pl-3.5">
              <span
                v-for="(chip, idx) in getAuditChips(trail)"
                :key="idx"
                class="px-2.5 py-1 rounded-lg bg-slate-800/80 border border-slate-700/60 text-3xs font-mono text-slate-300 flex items-center space-x-1.5 shadow-sm"
              >
                <span class="text-slate-400 font-sans font-medium">{{ chip.label }}:</span>
                <span class="font-bold text-white font-mono">{{ chip.val }}</span>
              </span>
            </div>

            <!-- Optional Collapsible JSON (Technical Details) -->
            <div v-if="hasJsonDetails(trail.details)" class="pl-3.5 pt-1">
              <button
                type="button"
                @click="toggleRawAudit(trail.id)"
                class="text-3xs font-mono text-blue-400 hover:text-blue-300 flex items-center space-x-1 transition"
              >
                <span>{{ showRawAudit[trail.id] ? '▲ Sembunyikan Detail Teknis JSON' : '▼ Lihat Detail Teknis JSON' }}</span>
              </button>
              <pre
                v-if="showRawAudit[trail.id]"
                class="mt-2 p-3 rounded-xl bg-[#060910] border border-slate-800 text-3xs font-mono text-slate-400 overflow-x-auto whitespace-pre-wrap break-all"
              >{{ formatRawJson(trail.details) }}</pre>
            </div>
          </div>
        </div>

        <div v-if="activities.length === 0" class="text-center py-10 text-xs text-slate-400 saas-card">
          <svg class="w-8 h-8 text-slate-600 mx-auto mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
          </svg>
          <span class="font-medium text-slate-300">Belum ada riwayat aktivitas yang tercatat untuk perangkat ini.</span>
          <p class="text-3xs text-slate-500 mt-0.5">Semua perubahan parameter, status koneksi, dan event operasional akan muncul di sini secara realtime.</p>
        </div>
      </div>
    </div>

    <!-- TAB 5: HEALTH DIAGNOSTICS -->
    <div v-if="device && activeTab === 'health'" class="saas-card p-6 space-y-6">
      <div class="border-b border-slate-800 pb-3">
        <h3 class="text-sm font-bold text-white">Device Health & Diagnostic Foundation</h3>
        <p class="text-2xs text-slate-400 mt-0.5">Real hardware counters, latency tracking, and communication stability</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2">
          <div class="text-3xs text-slate-400 uppercase font-semibold">Communication Status</div>
          <div class="text-sm font-mono font-bold" :class="commTextClass(device.connection_status)">
            {{ device.connection_status || 'UNKNOWN' }}
          </div>
          <p class="text-3xs text-slate-400">Tracked by health supervisor</p>
        </div>

        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2">
          <div class="text-3xs text-slate-400 uppercase font-semibold">Observed Latency</div>
          <div class="text-sm font-mono font-bold text-emerald-400">
            {{ device.latency_ms || 0 }} ms
          </div>
          <p class="text-3xs text-slate-400">Roundtrip request/response delay</p>
        </div>

        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2">
          <div class="text-3xs text-slate-400 uppercase font-semibold">Packets Delivery</div>
          <div class="text-sm font-mono font-bold text-slate-200">
            {{ device.success_count || 0 }} ok / {{ device.failed_count || 0 }} err
          </div>
          <p class="text-3xs text-slate-400">Accumulated transmission packets</p>
        </div>
      </div>

      <div class="p-4 rounded-xl bg-blue-500/10 border border-blue-500/20 text-xs text-blue-300 space-y-1">
        <div class="font-bold flex items-center space-x-1.5">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>Phase 2.1 Health Architecture</span>
        </div>
        <p class="text-3xs text-blue-200">
          Uptime calculations, continuous error rate monitoring, and data quality flags will be fed dynamically when Phase 2.2 / 2.3 communication engines are connected. No fake telemetry metrics are generated.
        </p>
      </div>
    </div>

    <!-- TAB 6: LIVE TELEMETRY (Phase 3.1) -->
    <div v-if="device && activeTab === 'telemetry'" class="space-y-4">
      <!-- Live Status Bar -->
      <div class="saas-card p-4 flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-[#0F172A]/90">
        <div class="flex items-center space-x-3">
          <div class="p-2 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center">
            <span class="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse"></span>
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h3 class="text-xs font-bold text-white uppercase tracking-wider">Realtime Telemetry Stream</h3>
              <span class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                WEBSOCKET LIVE
              </span>
            </div>
            <p class="text-2xs text-slate-400 mt-0.5">
              Live engineering measurements updated reactively from field polling engine without page reload.
            </p>
          </div>
        </div>

        <div class="flex items-center space-x-4 text-xs font-mono">
          <div class="text-right">
            <span class="text-3xs text-slate-500 uppercase block">Last Received Packet</span>
            <span class="text-slate-200">{{ formatTimestamp(lastLivePacketTime || device.last_data_at) }}</span>
          </div>
          <button
            @click="fetchLatestTelemetry"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-2xs font-semibold transition flex items-center space-x-1.5"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>Refresh</span>
          </button>
        </div>
      </div>

      <!-- Live Parameter Cards Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <div
          v-for="param in displayParameters"
          :key="param.id"
          class="saas-card p-4 space-y-3 relative overflow-hidden transition-all duration-300 hover:border-slate-700"
          :class="{ 'border-emerald-500/40 shadow-[0_0_15px_rgba(16,185,129,0.15)]': liveBlinks[param.id] }"
        >
          <!-- Card Header -->
          <div class="flex items-start justify-between">
            <div>
              <div class="flex items-center space-x-2">
                <span class="font-mono text-2xs font-bold text-blue-400 bg-blue-500/10 px-2 py-0.5 rounded border border-blue-500/20">
                  {{ param.parameter_code }}
                </span>
                <span class="text-3xs text-slate-500 font-mono">
                  #{{ param.register_address }} ({{ param.data_type }})
                </span>
              </div>
              <h4 class="text-sm font-semibold text-white mt-1.5 truncate">{{ param.parameter_name }}</h4>
            </div>
            <span
              class="px-2 py-0.5 rounded text-3xs font-mono font-bold"
              :class="qualityBadgeClass(getParamQuality(param))"
            >
              {{ getParamQuality(param) }}
            </span>
          </div>

          <!-- Primary Metric Value -->
          <div class="py-2">
            <div class="flex items-baseline space-x-2">
              <span class="text-3xl font-mono font-bold text-white tracking-tight">
                {{ formatParamValue(param) }}
              </span>
              <span class="text-sm font-medium text-slate-400 font-mono">{{ param.unit }}</span>
            </div>
          </div>

          <!-- Raw & Decoded Diagnostics Sub-bar -->
          <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800/80 grid grid-cols-2 gap-2 text-3xs font-mono text-slate-400">
            <div>
              <span class="text-slate-500 block uppercase">Raw / Hex</span>
              <span class="text-slate-300 truncate block">
                {{ getParamRaw(param) }}
              </span>
            </div>
            <div>
              <span class="text-slate-500 block uppercase">Timestamp</span>
              <span class="text-slate-300 truncate block">
                {{ formatTimestamp(getParamTime(param)) }}
              </span>
            </div>
          </div>

          <!-- Card Action -->
          <div class="pt-1 flex items-center justify-between border-t border-slate-800/60 text-2xs">
            <span class="text-slate-500 font-mono">Source: {{ (device.connection && device.connection.protocol) || 'MODBUS' }}</span>
            <button
              @click="inspectHistoryForParam(param.id)"
              class="text-blue-400 hover:text-blue-300 font-semibold flex items-center space-x-1"
            >
              <span>View History</span>
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
              </svg>
            </button>
          </div>
        </div>

        <div v-if="!device.parameters || device.parameters.length === 0" class="col-span-full py-12 text-center text-xs text-slate-400">
          No parameters configured for this device. Add parameters in the Parameters tab to start telemetry ingestion.
        </div>
      </div>
    </div>

    <!-- TAB 7: HISTORICAL DATA (Phase 3.1) -->
    <div v-if="device && activeTab === 'history'" class="space-y-4">
      <!-- Filter Bar -->
      <div class="saas-card p-4 flex flex-col lg:flex-row lg:items-center justify-between gap-3 bg-[#0F172A]/90 border border-slate-800 rounded-2xl">
        <div class="flex flex-wrap items-center gap-3">
          <!-- Parameter Filter -->
          <div>
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">Parameter</label>
            <select
              v-model="historyParamId"
              @change="onFilterChange"
              class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
            >
              <option value="">Semua Parameter (All)</option>
              <option v-for="p in device.parameters" :key="p.id" :value="p.id">
                {{ p.parameter_code }} — {{ p.parameter_name }} ({{ p.unit }})
              </option>
            </select>
          </div>

          <!-- Time Range Quick Buttons -->
          <div>
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">Rentang Waktu</label>
            <div class="flex items-center space-x-1">
              <button
                v-for="tr in timeRangeOptions"
                :key="tr.value"
                @click="setTimeRange(tr.value)"
                class="px-2.5 py-1.5 rounded-lg text-2xs font-semibold transition"
                :class="historyTimeRange === tr.value
                  ? 'bg-blue-600 text-white shadow-sm shadow-blue-500/30'
                  : 'bg-[#0B0F19] text-slate-400 hover:text-slate-200 border border-slate-700/60'"
              >
                {{ tr.label }}
              </button>
            </div>
          </div>

          <!-- Quality Filter -->
          <div>
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">Kualitas Data</label>
            <select
              v-model="historyQuality"
              @change="fetchHistory"
              class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
            >
              <option value="">Semua Kualitas</option>
              <option value="GOOD">GOOD (Normal)</option>
              <option value="BAD">BAD (Error)</option>
              <option value="UNCERTAIN">UNCERTAIN</option>
            </select>
          </div>

          <!-- Resolution / Sample Limit -->
          <div>
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">Kerapatan Titik</label>
            <select
              v-model="historyPageSize"
              @change="onPageSizeChange"
              class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500 font-mono"
            >
              <option :value="20">20 Sampel</option>
              <option :value="50">50 Sampel</option>
              <option :value="100">100 Sampel</option>
              <option :value="250">250 Sampel</option>
            </select>
          </div>
        </div>

        <div class="flex items-center space-x-2 pt-2 lg:pt-0">
          <button
            @click="fetchHistory"
            :disabled="historyLoading"
            class="px-3.5 py-1.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50 shadow-sm shadow-blue-500/20"
          >
            <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': historyLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>Segarkan Grafik</span>
          </button>
        </div>
      </div>

      <!-- Historical Trend Chart Card -->
      <div class="saas-card p-4 sm:p-5 space-y-4">
        <!-- Card Top Bar: Title, Parameter Quick Tabs, Controls -->
        <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 border-b border-slate-800 pb-3">
          <div>
            <div class="flex items-center space-x-2">
              <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20 uppercase tracking-wider">
                Historical Trend Curve
              </span>
              <span class="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-3xs font-sans font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                <span>Live Stream</span>
              </span>
            </div>
            <h3 class="text-sm font-bold text-white mt-1.5 flex items-center space-x-2 font-sans">
              <span>{{ chartActiveName }}</span>
              <span class="text-blue-400 font-mono font-medium">({{ chartActiveCode }})</span>
            </h3>
            <p class="text-3xs text-slate-400 font-sans mt-0.5">
              Grafik dinamika nilai sensor terhadap waktu &bull; Jendela: <span class="text-slate-300 font-medium">{{ historyTimeRangeLabel }}</span>
            </p>
          </div>

          <!-- Parameter Switcher Pills (If multiple parameters) -->
          <div v-if="device.parameters && device.parameters.length > 1" class="flex items-center space-x-1.5 flex-wrap">
            <span class="text-3xs text-slate-400 uppercase font-semibold mr-1 font-sans">Parameter:</span>
            <button
              v-for="p in device.parameters"
              :key="p.id"
              @click="selectChartParam(p.id)"
              class="px-2.5 py-1 rounded-lg text-2xs font-medium transition flex items-center space-x-1 font-sans"
              :class="activeChartParamId === p.id
                ? 'bg-blue-600/20 text-blue-300 border border-blue-500/40 font-semibold shadow-sm'
                : 'bg-slate-800/60 text-slate-400 hover:text-slate-200 border border-slate-700/60'"
            >
              <span>{{ p.parameter_code }}</span>
              <span class="text-3xs text-slate-400">({{ p.unit || '--' }})</span>
            </button>
          </div>

          <!-- Chart Visual Controls Toolbar -->
          <div class="flex items-center space-x-2 text-2xs font-sans">
            <!-- Curve Type Toggle -->
            <div class="bg-slate-900 p-0.5 rounded-lg border border-slate-800 flex items-center">
              <button
                @click="chartCurveType = 'smooth'"
                class="px-2.5 py-1 rounded-md text-3xs font-semibold transition"
                :class="chartCurveType === 'smooth' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'"
                title="Kurva Spline Halus (Bezier)"
              >
                Halus
              </button>
              <button
                @click="chartCurveType = 'linear'"
                class="px-2.5 py-1 rounded-md text-3xs font-semibold transition"
                :class="chartCurveType === 'linear' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'"
                title="Garis Sudut Linear"
              >
                Garis
              </button>
            </div>

            <!-- Toggle Dots -->
            <button
              @click="chartShowPoints = !chartShowPoints"
              class="px-2.5 py-1 rounded-lg border text-3xs font-medium transition flex items-center space-x-1.5"
              :class="chartShowPoints
                ? 'bg-blue-600/15 text-blue-400 border-blue-500/30'
                : 'bg-slate-900 text-slate-500 border-slate-800 hover:text-slate-300'"
              title="Tampilkan / Sembunyikan Titik Data"
            >
              <span class="w-1.5 h-1.5 rounded-full" :class="chartShowPoints ? 'bg-blue-400' : 'bg-slate-600'"></span>
              <span>Titik</span>
            </button>

            <!-- Toggle Avg Line -->
            <button
              @click="chartShowAvgLine = !chartShowAvgLine"
              class="px-2.5 py-1 rounded-lg border text-3xs font-medium transition flex items-center space-x-1.5"
              :class="chartShowAvgLine
                ? 'bg-amber-500/10 text-amber-400 border-amber-500/30'
                : 'bg-slate-900 text-slate-500 border-slate-800 hover:text-slate-300'"
              title="Garis Referensi Rata-Rata"
            >
              <span>Rata-Rata</span>
            </button>
          </div>
        </div>

        <!-- 5 KPI Cards Row -->
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 font-sans">
          <!-- KPI 1: Nilai Terkini -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider">
              <span>Nilai Terkini</span>
              <span class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold" :class="qualityBadgeClass(chartStats.latestQuality)">
                {{ chartStats.latestQuality }}
              </span>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-mono font-bold text-white tracking-tight">
                {{ chartStats.latest }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans flex items-center justify-between">
              <span>Jam {{ chartStats.latestTime }}</span>
              <span class="text-emerald-400 font-medium">Live</span>
            </div>
          </div>

          <!-- KPI 2: Minimum -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider">
              <span>Minimum (Valley)</span>
              <svg class="w-3.5 h-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 14l-7 7m0 0l-7-7m7 7V3"></path>
              </svg>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-mono font-bold text-emerald-400 tracking-tight">
                {{ chartStats.min }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans">
              Terendah @ {{ chartStats.minTime }}
            </div>
          </div>

          <!-- KPI 3: Maximum -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider">
              <span>Maksimum (Peak)</span>
              <svg class="w-3.5 h-3.5 text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 10l7-7m0 0l7 7m-7-7v18"></path>
              </svg>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-mono font-bold text-rose-400 tracking-tight">
                {{ chartStats.max }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans">
              Puncak @ {{ chartStats.maxTime }}
            </div>
          </div>

          <!-- KPI 4: Average -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider">
              <span>Rata-Rata (Mean)</span>
              <span class="text-3xs text-blue-400 font-bold font-sans">AVG</span>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-mono font-bold text-blue-400 tracking-tight">
                {{ chartStats.avg }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans">
              Rerata Seluruh Sampel
            </div>
          </div>

          <!-- KPI 5: Samples & Delta -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1 col-span-2 sm:col-span-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider">
              <span>Variasi Rentang</span>
              <span class="text-3xs text-slate-400 font-bold font-sans">DELTA</span>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-mono font-bold text-slate-200 tracking-tight">
                &Delta; {{ chartStats.delta }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans">
              Total <b class="text-slate-300 font-mono">{{ chartStats.count }}</b> Sampel Data
            </div>
          </div>
        </div>

        <!-- SVG Trend Chart Canvas Area -->
        <div
          ref="chartContainer"
          class="w-full h-72 sm:h-80 bg-[#0B0F19] rounded-xl border border-slate-800 relative p-1 sm:p-2 select-none overflow-hidden group font-sans"
        >
          <!-- Top Left Unit Watermark -->
          <div class="absolute top-2.5 left-4 z-10 flex items-center space-x-1.5 text-3xs font-sans text-slate-400 pointer-events-none">
            <span class="w-1.5 h-1.5 rounded-full bg-blue-500"></span>
            <span>Skala Y: <b class="text-slate-300 font-mono">{{ chartActiveUnit || 'Nilai' }}</b></span>
          </div>

          <!-- Empty State if no records -->
          <div v-if="chartPoints.length === 0" class="w-full h-full flex flex-col items-center justify-center space-y-2 text-center">
            <svg class="w-10 h-10 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"></path>
            </svg>
            <p class="text-xs text-slate-400 font-medium font-sans">Belum ada data rekaman telemetri untuk parameter ini.</p>
            <p class="text-3xs text-slate-500 font-sans">Data akan muncul secara otomatis saat polling engine mengirimkan pembacaan sensor.</p>
          </div>

          <!-- SVG Chart -->
          <svg
            v-else
            class="w-full h-full overflow-visible"
            viewBox="0 0 900 280"
            preserveAspectRatio="none"
            @mousemove="onChartMouseMove"
            @mouseleave="onChartMouseLeave"
          >
            <defs>
              <!-- Gradient Area Fill -->
              <linearGradient id="trendAreaGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stop-color="#3B82F6" stop-opacity="0.22" />
                <stop offset="60%" stop-color="#3B82F6" stop-opacity="0.05" />
                <stop offset="100%" stop-color="#3B82F6" stop-opacity="0.00" />
              </linearGradient>

              <!-- Stroke Gradient -->
              <linearGradient id="trendStrokeGradient" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%" stop-color="#60A5FA" />
                <stop offset="100%" stop-color="#3B82F6" />
              </linearGradient>

              <!-- Subtle Soft Drop Filter -->
              <filter id="neonCurveGlow" x="-10%" y="-10%" width="120%" height="120%">
                <feGaussianBlur stdDeviation="1.5" result="blur" />
                <feMerge>
                  <feMergeNode in="blur" />
                  <feMergeNode in="SourceGraphic" />
                </feMerge>
              </filter>
            </defs>

            <!-- Horizontal Grid Lines & Y-Axis Labels -->
            <g class="grid-lines">
              <g v-for="(tick, idx) in chartGridTicks" :key="'grid-' + idx">
                <line
                  :x1="65"
                  :y1="tick.y"
                  :x2="860"
                  :y2="tick.y"
                  stroke="#1E293B"
                  stroke-dasharray="4 4"
                  stroke-width="1"
                />
                <text
                  :x="55"
                  :y="tick.y + 4"
                  text-anchor="end"
                  fill="#94A3B8"
                  font-size="11"
                  font-family="'JetBrains Mono', 'Inter', monospace"
                  font-weight="500"
                >
                  {{ tick.label }}
                </text>
              </g>
            </g>

            <!-- Vertical Ticks & Timestamps on X-Axis -->
            <g class="time-ticks">
              <g v-for="(tick, idx) in chartTimeTicks" :key="'ttick-' + idx">
                <line
                  :x1="tick.x"
                  :y1="232"
                  :x2="tick.x"
                  :y2="238"
                  stroke="#334155"
                  stroke-width="1.2"
                />
                <text
                  :x="tick.x"
                  :y="254"
                  text-anchor="middle"
                  fill="#94A3B8"
                  font-size="11"
                  font-family="'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"
                  font-weight="500"
                >
                  {{ tick.time }}
                </text>
              </g>
            </g>

            <!-- Area Fill Under Curve -->
            <path
              :d="chartAreaPath"
              fill="url(#trendAreaGradient)"
              class="transition-all duration-300"
            />

            <!-- Average Reference Line -->
            <g v-if="chartShowAvgLine && chartStats.count > 0">
              <line
                :x1="65"
                :y1="chartAvgY"
                :x2="860"
                :y2="chartAvgY"
                stroke="#F59E0B"
                stroke-dasharray="5 4"
                stroke-width="1.2"
                opacity="0.8"
              />
              <rect
                :x="790"
                :y="Math.max(6, chartAvgY - 9)"
                width="70"
                height="18"
                rx="5"
                fill="#1E293B"
                stroke="#F59E0B"
                stroke-width="1"
                opacity="0.95"
              />
              <text
                :x="825"
                :y="Math.max(6, chartAvgY - 9) + 12"
                text-anchor="middle"
                fill="#FBBF24"
                font-size="10"
                font-family="'Inter', -apple-system, sans-serif"
                font-weight="600"
              >
                AVG {{ chartStats.avg }}
              </text>
            </g>

            <!-- Main Trend Curve Line -->
            <path
              :d="chartCurvePath"
              fill="none"
              stroke="url(#trendStrokeGradient)"
              stroke-width="2.5"
              stroke-linecap="round"
              stroke-linejoin="round"
              filter="url(#neonCurveGlow)"
              class="transition-all duration-200"
            />

            <!-- Peak (Max) Annotation Badge -->
            <g v-if="chartShowMinMax && chartMaxPoint && chartStats.count > 1">
              <circle
                :cx="chartMaxPoint.x"
                :cy="chartMaxPoint.y"
                r="5"
                fill="#F43F5E"
                stroke="#FFFFFF"
                stroke-width="1.5"
              />
              <rect
                :x="Math.max(65, Math.min(chartMaxPoint.x - 34, 795))"
                :y="Math.max(6, chartMaxPoint.y - 24)"
                width="68"
                height="19"
                rx="6"
                fill="#1E293B"
                stroke="#F43F5E"
                stroke-width="1.2"
                opacity="0.95"
              />
              <text
                :x="Math.max(99, Math.min(chartMaxPoint.x, 829))"
                :y="Math.max(6, chartMaxPoint.y - 24) + 13"
                text-anchor="middle"
                fill="#FDA4AF"
                font-size="10"
                font-family="'Inter', -apple-system, sans-serif"
                font-weight="600"
              >
                MAX {{ chartMaxPoint.value.toFixed(2) }}
              </text>
            </g>

            <!-- Valley (Min) Annotation Badge -->
            <g v-if="chartShowMinMax && chartMinPoint && chartStats.count > 1 && chartMinPoint.y !== chartMaxPoint.y">
              <circle
                :cx="chartMinPoint.x"
                :cy="chartMinPoint.y"
                r="5"
                fill="#10B981"
                stroke="#FFFFFF"
                stroke-width="1.5"
              />
              <rect
                :x="Math.max(65, Math.min(chartMinPoint.x - 34, 795))"
                :y="Math.min(214, chartMinPoint.y + 7)"
                width="68"
                height="19"
                rx="6"
                fill="#1E293B"
                stroke="#10B981"
                stroke-width="1.2"
                opacity="0.95"
              />
              <text
                :x="Math.max(99, Math.min(chartMinPoint.x, 829))"
                :y="Math.min(214, chartMinPoint.y + 7) + 13"
                text-anchor="middle"
                fill="#6EE7B7"
                font-size="10"
                font-family="'Inter', -apple-system, sans-serif"
                font-weight="600"
              >
                MIN {{ chartMinPoint.value.toFixed(2) }}
              </text>
            </g>

            <!-- Data Point Dots -->
            <g v-if="chartShowPoints">
              <circle
                v-for="(p, i) in chartPoints"
                :key="'dot-' + i"
                :cx="p.x"
                :cy="p.y"
                :r="chartPoints.length > 50 ? 2 : 3"
                fill="#3B82F6"
                stroke="#0B0F19"
                stroke-width="1.5"
                class="hover:opacity-100"
              />
            </g>

            <!-- Active Mouse Crosshair Guidelines -->
            <g v-if="hoveredPoint">
              <line
                :x1="hoveredPoint.x"
                :y1="32"
                :x2="hoveredPoint.x"
                :y2="232"
                stroke="#60A5FA"
                stroke-width="1.2"
                stroke-dasharray="3 3"
                opacity="0.8"
              />
              <line
                :x1="65"
                :y1="hoveredPoint.y"
                :x2="860"
                :y2="hoveredPoint.y"
                stroke="#60A5FA"
                stroke-width="1"
                stroke-dasharray="3 3"
                opacity="0.5"
              />
              <!-- Pulsing beacon on the hovered point -->
              <circle
                :cx="hoveredPoint.x"
                :cy="hoveredPoint.y"
                r="7"
                fill="none"
                stroke="#60A5FA"
                stroke-width="2"
                class="animate-ping"
              />
              <circle
                :cx="hoveredPoint.x"
                :cy="hoveredPoint.y"
                r="5"
                fill="#3B82F6"
                stroke="#FFFFFF"
                stroke-width="1.5"
              />
            </g>
          </svg>

          <!-- Floating Glassmorphism Tooltip -->
          <div
            v-if="hoveredPoint"
            class="absolute pointer-events-none z-30 transition-all duration-75"
            :style="tooltipStyle"
          >
            <div class="bg-[#111827]/95 backdrop-blur-md border border-slate-700 shadow-2xl rounded-xl p-3 text-xs w-60 space-y-1.5 font-sans ring-1 ring-white/10">
              <div class="flex items-center justify-between border-b border-slate-800 pb-1.5">
                <span class="text-3xs font-mono font-bold text-blue-400 uppercase tracking-wider">
                  {{ getParamCode(hoveredPoint.rawRecord.parameter_id) }}
                </span>
                <span
                  class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold"
                  :class="qualityBadgeClass(hoveredPoint.quality)"
                >
                  {{ hoveredPoint.quality }}
                </span>
              </div>
              <div class="text-3xs text-slate-300 font-medium truncate font-sans">
                {{ getParamName(hoveredPoint.rawRecord.parameter_id) }}
              </div>
              <div class="flex items-baseline space-x-1 pt-0.5">
                <span class="text-2xl font-mono font-bold text-white tracking-tight">
                  {{ hoveredPoint.value !== undefined ? hoveredPoint.value.toFixed(2) : '--' }}
                </span>
                <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
              </div>
              <div class="pt-1.5 border-t border-slate-800 flex items-center justify-between text-3xs text-slate-400 font-sans">
                <span>{{ formatTimestamp(hoveredPoint.timestamp) }}</span>
                <span class="text-slate-500 font-mono">#{{ hoveredPoint.index + 1 }}/{{ chartPoints.length }}</span>
              </div>
              <div v-if="hoveredPoint.rawRecord.raw_value !== undefined" class="text-4xs text-slate-500 font-mono flex items-center justify-between">
                <span>Raw: {{ hoveredPoint.rawRecord.raw_value }}</span>
                <span>{{ hoveredPoint.rawRecord.source || 'MODBUS' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Legend and Operator Guide -->
        <div class="flex flex-wrap items-center justify-between gap-3 text-3xs text-slate-400 pt-1 font-sans">
          <div class="flex items-center space-x-4 flex-wrap">
            <span class="flex items-center space-x-1.5">
              <span class="w-3 h-1 rounded bg-blue-500 inline-block"></span>
              <span>Tren Telemetri ({{ chartActiveUnit }})</span>
            </span>
            <span v-if="chartShowAvgLine" class="flex items-center space-x-1.5">
              <span class="w-3 h-0.5 border-t border-dashed border-amber-400 inline-block"></span>
              <span>Rata-Rata ({{ chartStats.avg }})</span>
            </span>
            <span v-if="chartShowMinMax" class="flex items-center space-x-1.5">
              <span class="w-2 h-2 rounded-full bg-rose-500 inline-block"></span>
              <span>Puncak Max</span>
            </span>
            <span v-if="chartShowMinMax" class="flex items-center space-x-1.5">
              <span class="w-2 h-2 rounded-full bg-emerald-500 inline-block"></span>
              <span>Titik Min</span>
            </span>
          </div>
          <div class="text-slate-500 italic flex items-center space-x-1">
            <svg class="w-3 h-3 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"></path>
            </svg>
            <span>Gerakkan kursor pada grafik untuk membaca rincian telemetri pada setiap detik.</span>
          </div>
        </div>
      </div>

      <!-- Historical Data Table -->
      <div class="saas-card overflow-hidden">
        <div class="p-3 border-b border-slate-800 flex items-center justify-between text-xs">
          <span class="font-bold text-white">Historical Telemetry Records ({{ historyTotal }} total)</span>
          <span class="text-3xs text-slate-400 font-mono">Page {{ historyPage }} of {{ Math.ceil(historyTotal / historyPageSize) || 1 }}</span>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs font-sans">
            <thead class="bg-[#0B0F19] text-3xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-800">
              <tr>
                <th class="py-2.5 px-4">Received Time (UTC / Local)</th>
                <th class="py-2.5 px-4">Parameter</th>
                <th class="py-2.5 px-4">Value</th>
                <th class="py-2.5 px-4">Raw Decoded</th>
                <th class="py-2.5 px-4">Quality</th>
                <th class="py-2.5 px-4">Protocol Source</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 font-mono text-2xs">
              <tr v-for="rec in historyRecords" :key="rec.id" class="hover:bg-slate-800/30 transition">
                <td class="py-2.5 px-4 text-slate-300">
                  {{ formatTimestamp(rec.received_at || rec.timestamp) }}
                </td>
                <td class="py-2.5 px-4">
                  <span class="text-blue-400 font-semibold">{{ getParamCode(rec.parameter_id) }}</span>
                </td>
                <td class="py-2.5 px-4 text-white font-bold">
                  {{ rec.value !== undefined ? rec.value.toFixed(2) : '--' }}
                  <span class="text-slate-400 font-normal ml-0.5">{{ getParamUnit(rec.parameter_id) }}</span>
                </td>
                <td class="py-2.5 px-4 text-slate-400">
                  {{ rec.raw_value !== undefined ? rec.raw_value : '--' }}
                </td>
                <td class="py-2.5 px-4">
                  <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold" :class="qualityBadgeClass(rec.quality)">
                    {{ rec.quality || 'GOOD' }}
                  </span>
                </td>
                <td class="py-2.5 px-4 text-slate-400">
                  {{ rec.source || 'MODBUS_TCP' }}
                </td>
              </tr>
              <tr v-if="historyRecords.length === 0">
                <td colspan="6" class="py-8 text-center text-xs text-slate-400 font-sans">
                  {{ historyLoading ? 'Loading telemetry records...' : 'No telemetry records found for selected query filter.' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination Controls -->
        <div class="p-3 border-t border-slate-800 flex items-center justify-between text-xs">
          <button
            @click="prevHistoryPage"
            :disabled="historyPage <= 1 || historyLoading"
            class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold"
          >
            Previous
          </button>
          <span class="text-slate-400 text-2xs font-mono">
            Showing {{ (historyPage - 1) * historyPageSize + 1 }} to {{ Math.min(historyPage * historyPageSize, historyTotal) }} of {{ historyTotal }}
          </span>
          <button
            @click="nextHistoryPage"
            :disabled="historyPage * historyPageSize >= historyTotal || historyLoading"
            class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold"
          >
            Next
          </button>
        </div>
      </div>
    </div>

    <!-- TAB 8: RAW TELEMETRY (Phase 3.1) -->
    <div v-if="device && activeTab === 'raw'" class="space-y-4">
      <div class="saas-card p-4 flex flex-col md:flex-row md:items-center justify-between gap-3 bg-[#0F172A]/90">
        <div>
          <h3 class="text-xs font-bold text-white uppercase tracking-wider">Raw Telemetry & Frame Inspection</h3>
          <p class="text-2xs text-slate-400 mt-0.5">
            Diagnostic wire hex inspection for Modbus register verification and field troubleshooting.
          </p>
        </div>
        <div class="flex items-center space-x-2">
          <select
            v-model="rawParamId"
            @change="fetchRawTelemetry"
            class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option value="">All Parameters</option>
            <option v-for="p in device.parameters" :key="p.id" :value="p.id">
              {{ p.parameter_code }}
            </option>
          </select>
          <button
            @click="fetchRawTelemetry"
            :disabled="rawLoading"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition"
          >
            Refresh
          </button>
        </div>
      </div>

      <!-- Raw Inspection Table -->
      <div class="saas-card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs font-mono">
            <thead class="bg-[#0B0F19] text-3xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-800">
              <tr>
                <th class="py-2.5 px-4">Received Time</th>
                <th class="py-2.5 px-4">Parameter</th>
                <th class="py-2.5 px-4">Raw Hex Stream</th>
                <th class="py-2.5 px-4">Raw Float</th>
                <th class="py-2.5 px-4">Scaled Value</th>
                <th class="py-2.5 px-4">Quality</th>
                <th class="py-2.5 px-4">Protocol</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 text-2xs">
              <tr v-for="rec in rawRecords" :key="rec.id" class="hover:bg-slate-800/30 transition">
                <td class="py-2.5 px-4 text-slate-300">
                  {{ formatTimestamp(rec.received_at || rec.timestamp) }}
                </td>
                <td class="py-2.5 px-4 text-blue-400 font-bold">
                  {{ getParamCode(rec.parameter_id) }}
                </td>
                <td class="py-2.5 px-4 text-emerald-400 font-bold tracking-widest">
                  {{ rec.raw_hex || '--' }}
                </td>
                <td class="py-2.5 px-4 text-slate-400">
                  {{ rec.raw_value !== undefined ? rec.raw_value : '--' }}
                </td>
                <td class="py-2.5 px-4 text-white font-bold">
                  {{ rec.value !== undefined ? rec.value.toFixed(2) : '--' }}
                </td>
                <td class="py-2.5 px-4">
                  <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold" :class="qualityBadgeClass(rec.quality)">
                    {{ rec.quality || 'GOOD' }}
                  </span>
                </td>
                <td class="py-2.5 px-4 text-slate-400">
                  {{ rec.source || 'MODBUS_TCP' }}
                </td>
              </tr>
              <tr v-if="rawRecords.length === 0">
                <td colspan="7" class="py-8 text-center text-xs text-slate-400 font-sans">
                  {{ rawLoading ? 'Loading raw telemetry...' : 'No raw telemetry records captured yet.' }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Raw Pagination -->
        <div class="p-3 border-t border-slate-800 flex items-center justify-between text-xs">
          <button
            @click="prevRawPage"
            :disabled="rawPage <= 1 || rawLoading"
            class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold"
          >
            Previous
          </button>
          <span class="text-slate-400 text-2xs font-mono">
            Page {{ rawPage }} of {{ Math.ceil(rawTotal / rawPageSize) || 1 }} ({{ rawTotal }} records)
          </span>
          <button
            @click="nextRawPage"
            :disabled="rawPage * rawPageSize >= rawTotal || rawLoading"
            class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold"
          >
            Next
          </button>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <DeviceModal
      :is-open="isEditModalOpen"
      :device-to-edit="device"
      @close="isEditModalOpen = false"
      @saved="loadDevice"
    />

    <ParameterModal
      v-if="device"
      :is-open="isParamModalOpen"
      :device-id="device.id"
      :param-to-edit="paramToEdit"
      @close="isParamModalOpen = false"
      @saved="loadDevice"
    />

    <!-- Modbus Parameter Test Read Diagnostic Modal -->
    <div v-if="isTestReadModalOpen && testParam" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
      <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden flex flex-col">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
          <div class="flex items-center space-x-2.5">
            <div class="p-1.5 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
              </svg>
            </div>
            <div>
              <h2 class="text-sm font-bold text-white tracking-wide">Modbus Diagnostic Parameter Read</h2>
              <p class="text-3xs text-slate-400 font-mono">Device: {{ device.device_code }} | Param: {{ testParam.parameter_code }}</p>
            </div>
          </div>
          <button @click="closeTestReadModal" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
          </button>
        </div>

        <!-- Body -->
        <div class="p-6 space-y-4 text-xs font-sans">
          <!-- Parameter Specs -->
          <div class="grid grid-cols-2 gap-3 p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
            <div>
              <span class="text-3xs text-slate-400 font-semibold uppercase">Parameter</span>
              <div class="text-xs text-white font-medium mt-0.5 truncate">{{ testParam.parameter_name }}</div>
            </div>
            <div>
              <span class="text-3xs text-slate-400 font-semibold uppercase">Data Type</span>
              <div class="text-xs font-mono font-bold text-blue-400 mt-0.5">{{ testParam.data_type }}</div>
            </div>
            <div>
              <span class="text-3xs text-slate-400 font-semibold uppercase">Register & Type</span>
              <div class="text-xs font-mono text-slate-200 mt-0.5">#{{ testParam.register_address }} ({{ testParam.register_type }})</div>
            </div>
            <div>
              <span class="text-3xs text-slate-400 font-semibold uppercase">Scale & Offset</span>
              <div class="text-xs font-mono text-slate-200 mt-0.5">×{{ testParam.scale || 1 }} + {{ testParam.offset || 0 }} ({{ testParam.unit || '--' }})</div>
            </div>
          </div>

          <!-- Loading state -->
          <div v-if="testReadLoading" class="py-8 flex flex-col items-center justify-center space-y-3">
            <div class="w-7 h-7 border-2 border-emerald-500/20 border-t-emerald-500 rounded-full animate-spin"></div>
            <span class="text-xs text-slate-400 font-mono">Executing Modbus frame transaction...</span>
          </div>

          <!-- Result card -->
          <div v-else-if="testReadResult" class="space-y-3">
            <div
              :class="testReadResult.success ? 'bg-emerald-500/10 border-emerald-500/30' : 'bg-rose-500/10 border-rose-500/30'"
              class="p-4 rounded-xl border space-y-2.5"
            >
              <div class="flex items-center justify-between">
                <span
                  :class="testReadResult.success ? 'text-emerald-400' : 'text-rose-400'"
                  class="text-xs font-bold uppercase tracking-wider flex items-center space-x-1.5"
                >
                  <span class="w-2 h-2 rounded-full" :class="testReadResult.success ? 'bg-emerald-400' : 'bg-rose-400'"></span>
                  <span>{{ testReadResult.success ? 'Transaction Succeeded' : 'Transaction Failed' }}</span>
                </span>
                <span class="text-3xs font-mono text-slate-400">{{ testReadResult.response_time_ms || 0 }} ms</span>
              </div>

              <!-- Error display if failed -->
              <div v-if="!testReadResult.success" class="text-xs text-rose-300 font-mono bg-rose-950/40 p-2.5 rounded-lg border border-rose-500/20">
                {{ testReadResult.error || testReadResult.error_message || 'Modbus communication failed' }}
              </div>

              <!-- Values display if success -->
              <div v-else class="grid grid-cols-2 gap-3 pt-1">
                <div class="p-2.5 bg-[#0B0F19]/80 rounded-lg border border-slate-800">
                  <div class="text-3xs text-slate-400 uppercase font-semibold">Decoded Value</div>
                  <div class="text-base font-mono font-bold text-emerald-400 mt-1">
                    {{ testReadResult.decoded_value !== undefined && testReadResult.decoded_value !== null ? testReadResult.decoded_value : '--' }}
                    <span class="text-xs font-normal text-slate-300 ml-1">{{ testParam.unit }}</span>
                  </div>
                </div>
                <div class="p-2.5 bg-[#0B0F19]/80 rounded-lg border border-slate-800">
                  <div class="text-3xs text-slate-400 uppercase font-semibold">Raw PDU Value</div>
                  <div class="text-xs font-mono text-slate-200 mt-1">
                    {{ testReadResult.raw_value !== undefined ? testReadResult.raw_value : '--' }}
                    <span v-if="testReadResult.raw_bytes_hex" class="block text-3xs text-slate-400 font-mono mt-0.5">
                      Bytes: {{ testReadResult.raw_bytes_hex }}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            <!-- Diagnostics Metadata -->
            <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800 grid grid-cols-3 gap-2 text-3xs font-mono text-slate-400">
              <div>
                <span class="text-slate-500 block uppercase">Function Code</span>
                <span class="text-slate-300 font-bold">FC {{ testReadResult.function_code || '--' }}</span>
              </div>
              <div>
                <span class="text-slate-500 block uppercase">PDU Offset</span>
                <span class="text-slate-300 font-bold">{{ testReadResult.register_address !== undefined ? testReadResult.register_address : '--' }}</span>
              </div>
              <div>
                <span class="text-slate-500 block uppercase">Timestamp</span>
                <span class="text-slate-300">{{ formatTimestamp(testReadResult.timestamp) }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-3 border-t border-slate-800 flex items-center justify-between bg-[#0B0F19]/60">
          <button
            type="button"
            @click="closeTestReadModal"
            class="px-3.5 py-1.5 rounded-xl text-xs font-medium text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            Close
          </button>
          <button
            type="button"
            @click="executeTestRead"
            :disabled="testReadLoading"
            class="px-4 py-1.5 rounded-xl text-xs font-bold text-white bg-emerald-600 hover:bg-emerald-500 transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>Read Again</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';
import DeviceModal from '../components/DeviceModal.vue';
import ParameterModal from '../components/ParameterModal.vue';

export default {
  name: 'DeviceDetailView',
  components: {
    DeviceModal,
    ParameterModal,
  },
  data() {
    return {
      device: null,
      activities: [],
      activityLoading: false,
      showRawAudit: {},
      activityPollTimer: null,
      activeTab: 'overview',
      isEditModalOpen: false,
      isParamModalOpen: false,
      paramToEdit: null,
      commStatus: null,
      commFeedback: null,
      commActionLoading: false,
      isTestReadModalOpen: false,
      testParam: null,
      testReadLoading: false,
      testReadResult: null,

      // Phase 3.1 Live Telemetry
      liveTelemetry: {},
      liveBlinks: {},
      lastLivePacketTime: null,
      livePacketBlink: false,

      // Phase 3.1 Historical Telemetry
      historyRecords: [],
      historyTotal: 0,
      historyPage: 1,
      historyPageSize: 20,
      historyParamId: '',
      historyQuality: '',
      historyTimeRange: '1h',
      historyLoading: false,

      // Historical Chart Enhancements
      chartCurveType: 'smooth',
      chartShowPoints: true,
      chartShowAvgLine: true,
      chartShowMinMax: true,
      chartParamId: '',
      hoveredPoint: null,
      hoverMouseClientX: 0,
      hoverMouseClientY: 0,
      containerWidth: 800,
      containerHeight: 320,
      timeRangeOptions: [
        { label: '15m', value: '15m' },
        { label: '1j', value: '1h' },
        { label: '6j', value: '6h' },
        { label: '24j', value: '24h' },
        { label: 'Semua', value: 'all' },
      ],

      // Phase 3.1 Raw Telemetry
      rawRecords: [],
      rawTotal: 0,
      rawPage: 1,
      rawPageSize: 20,
      rawParamId: '',
      rawLoading: false,
    };
  },
  computed: {
    displayParameters() {
      if (!this.device || !this.device.parameters) return [];
      return this.device.parameters;
    },
    historyTimeRangeLabel() {
      switch (this.historyTimeRange) {
        case '15m': return '15 Menit Terakhir';
        case '1h': return '1 Jam Terakhir';
        case '6h': return '6 Jam Terakhir';
        case '24h': return '24 Jam Terakhir';
        default: return 'Semua Riwayat Tersimpan';
      }
    },
    activeChartParamId() {
      if (this.chartParamId) return Number(this.chartParamId);
      if (this.historyParamId) return Number(this.historyParamId);
      if (this.historyRecords && this.historyRecords.length > 0) {
        return this.historyRecords[0].parameter_id;
      }
      if (this.device && this.device.parameters && this.device.parameters.length > 0) {
        return this.device.parameters[0].id;
      }
      return null;
    },
    chartActiveParam() {
      if (!this.device || !this.device.parameters || !this.activeChartParamId) return null;
      return this.device.parameters.find(p => p.id === this.activeChartParamId) || null;
    },
    chartActiveUnit() {
      return this.chartActiveParam ? (this.chartActiveParam.unit || '') : '';
    },
    chartActiveName() {
      return this.chartActiveParam
        ? (this.chartActiveParam.parameter_name || this.chartActiveParam.name || this.chartActiveParam.parameter_code)
        : 'Parameter Sensor';
    },
    chartActiveCode() {
      return this.chartActiveParam ? (this.chartActiveParam.parameter_code || this.chartActiveParam.code) : '';
    },
    chartFilteredRecords() {
      if (!this.historyRecords || this.historyRecords.length === 0) return [];
      if (this.activeChartParamId) {
        const filtered = this.historyRecords.filter(r => r.parameter_id === this.activeChartParamId);
        if (filtered.length > 0) return filtered;
      }
      return this.historyRecords;
    },
    chartPoints() {
      const records = this.chartFilteredRecords;
      if (!records || records.length === 0) return [];

      // Telemetry is fetched newest first (DESC); reverse for chronological left-to-right plotting
      const chronological = [...records].reverse();
      const vals = chronological.map(r => (r.value !== undefined && r.value !== null ? Number(r.value) : 0));
      let min = Math.min(...vals);
      let max = Math.max(...vals);

      // If all values are identical (flat line)
      if (min === max) {
        min = min - 1;
        max = max + 1;
      }

      const range = max - min || 1;
      const left = 65;
      const right = 860;
      const top = 32;
      const bottom = 232;
      const plotWidth = right - left;
      const plotHeight = bottom - top;

      const count = chronological.length;
      const maxVal = Math.max(...vals);
      const minVal = Math.min(...vals);

      return chronological.map((rec, idx) => {
        const val = (rec.value !== undefined && rec.value !== null) ? Number(rec.value) : 0;
        const x = count === 1 ? (left + plotWidth / 2) : (left + (idx / (count - 1)) * plotWidth);
        const normalizedY = (val - min) / range;
        const y = bottom - (normalizedY * plotHeight);

        return {
          x: Number(x.toFixed(1)),
          y: Number(y.toFixed(1)),
          value: val,
          rawRecord: rec,
          timestamp: rec.received_at || rec.timestamp,
          quality: rec.quality || 'GOOD',
          index: idx,
          isMax: val === maxVal,
          isMin: val === minVal,
        };
      });
    },
    chartCurvePath() {
      const pts = this.chartPoints;
      if (!pts || pts.length === 0) return '';
      if (pts.length === 1) return `M ${pts[0].x - 5},${pts[0].y} L ${pts[0].x + 5},${pts[0].y}`;

      if (this.chartCurveType === 'linear') {
        return 'M ' + pts.map(p => `${p.x},${p.y}`).join(' L ');
      }

      // Cubic Bezier Spline
      let d = `M ${pts[0].x},${pts[0].y}`;
      for (let i = 1; i < pts.length; i++) {
        const prev = pts[i - 1];
        const curr = pts[i];
        const dx = curr.x - prev.x;
        const cp1x = (prev.x + dx * 0.35).toFixed(1);
        const cp1y = prev.y.toFixed(1);
        const cp2x = (curr.x - dx * 0.35).toFixed(1);
        const cp2y = curr.y.toFixed(1);
        d += ` C ${cp1x},${cp1y} ${cp2x},${cp2y} ${curr.x},${curr.y}`;
      }
      return d;
    },
    chartAreaPath() {
      const curve = this.chartCurvePath;
      const pts = this.chartPoints;
      if (!curve || !pts || pts.length === 0) return '';
      const bottom = 232;
      const first = pts[0];
      const last = pts[pts.length - 1];
      return `${curve} L ${last.x},${bottom} L ${first.x},${bottom} Z`;
    },
    chartGridTicks() {
      const records = this.chartFilteredRecords;
      if (!records || records.length === 0) {
        return [
          { y: 32, label: '100.00' },
          { y: 82, label: '75.00' },
          { y: 132, label: '50.00' },
          { y: 182, label: '25.00' },
          { y: 232, label: '0.00' },
        ];
      }
      const vals = records.map(r => (r.value !== undefined && r.value !== null ? Number(r.value) : 0));
      let min = Math.min(...vals);
      let max = Math.max(...vals);
      if (min === max) {
        min = min - 1;
        max = max + 1;
      }
      const range = max - min;
      const top = 32;
      const bottom = 232;
      const plotHeight = bottom - top;

      const ratios = [1.0, 0.75, 0.5, 0.25, 0.0];
      return ratios.map(r => {
        const val = min + range * r;
        const y = Number((bottom - r * plotHeight).toFixed(1));
        return {
          y,
          value: val,
          label: val.toFixed(2),
        };
      });
    },
    chartTimeTicks() {
      const pts = this.chartPoints;
      if (!pts || pts.length === 0) return [];
      const n = pts.length;
      if (n <= 6) {
        return pts.map(p => ({
          x: p.x,
          time: this.formatTimeShort(p.timestamp),
        }));
      }
      const stepIndices = [
        0,
        Math.floor((n - 1) * 0.25),
        Math.floor((n - 1) * 0.5),
        Math.floor((n - 1) * 0.75),
        n - 1,
      ];
      return stepIndices.map(idx => ({
        x: pts[idx].x,
        time: this.formatTimeShort(pts[idx].timestamp),
      }));
    },
    chartStats() {
      const records = this.chartFilteredRecords;
      if (!records || records.length === 0) {
        return {
          latest: '--',
          latestQuality: 'UNKNOWN',
          latestTime: '--',
          min: '0.00',
          max: '0.00',
          avg: '0.00',
          delta: '0.00',
          count: 0,
          minTime: '--',
          maxTime: '--',
        };
      }

      // records[0] is the newest telemetry record
      const latestRec = records[0];
      const vals = records.map(r => (r.value !== undefined && r.value !== null ? Number(r.value) : 0));
      const minVal = Math.min(...vals);
      const maxVal = Math.max(...vals);
      const sum = vals.reduce((a, b) => a + b, 0);
      const avgVal = sum / vals.length;
      const deltaVal = maxVal - minVal;

      const minRec = records.find(r => r.value === minVal);
      const maxRec = records.find(r => r.value === maxVal);

      return {
        latest: (latestRec.value !== undefined && latestRec.value !== null) ? Number(latestRec.value).toFixed(2) : '--',
        latestQuality: latestRec.quality || 'GOOD',
        latestTime: this.formatTimeShort(latestRec.received_at || latestRec.timestamp),
        min: minVal.toFixed(2),
        max: maxVal.toFixed(2),
        avg: avgVal.toFixed(2),
        delta: deltaVal.toFixed(2),
        count: vals.length,
        minTime: minRec ? this.formatTimeShort(minRec.received_at || minRec.timestamp) : '--',
        maxTime: maxRec ? this.formatTimeShort(maxRec.received_at || maxRec.timestamp) : '--',
      };
    },
    chartMaxPoint() {
      const pts = this.chartPoints;
      if (!pts || pts.length === 0) return null;
      return pts.find(p => p.isMax) || null;
    },
    chartMinPoint() {
      const pts = this.chartPoints;
      if (!pts || pts.length === 0) return null;
      return pts.find(p => p.isMin) || null;
    },
    chartAvgY() {
      const stats = this.chartStats;
      const records = this.chartFilteredRecords;
      if (!records || records.length === 0) return 132;
      const vals = records.map(r => (r.value !== undefined && r.value !== null ? Number(r.value) : 0));
      let min = Math.min(...vals);
      let max = Math.max(...vals);
      if (min === max) {
        min = min - 1;
        max = max + 1;
      }
      const range = max - min || 1;
      const avg = Number(stats.avg);
      const top = 32;
      const bottom = 232;
      const plotHeight = bottom - top;
      const ratio = (avg - min) / range;
      return Number((bottom - ratio * plotHeight).toFixed(1));
    },
    tooltipStyle() {
      if (!this.hoveredPoint) return {};
      const tooltipWidth = 240;
      let left = this.hoverMouseClientX + 16;
      if (left + tooltipWidth > this.containerWidth) {
        left = this.hoverMouseClientX - tooltipWidth - 16;
      }
      if (left < 10) left = 10;

      let top = this.hoverMouseClientY - 70;
      if (top < 10) top = 10;
      if (top + 130 > this.containerHeight) top = this.containerHeight - 140;

      return {
        left: `${left}px`,
        top: `${top}px`,
      };
    },
    trendLinePoints() {
      const pts = this.chartPoints;
      if (!pts || pts.length === 0) return '';
      return pts.map(p => `${p.x},${p.y}`).join(' ');
    },
    historyStats() {
      return this.chartStats;
    },
  },
  watch: {
    '$route.query.tab'(newTab) {
      if (newTab) {
        this.switchTab(newTab);
      }
    },
  },
  mounted() {
    if (this.$route.query.tab) {
      this.activeTab = this.$route.query.tab;
    }
    this.loadDevice();
    if (this.activeTab === 'activity') {
      this.startActivityPolling();
    }
    window.addEventListener('device-comm-event', this.onCommEvent);
    window.addEventListener('device-telemetry-event', this.onTelemetryEvent);
  },
  beforeDestroy() {
    this.stopActivityPolling();
    window.removeEventListener('device-comm-event', this.onCommEvent);
    window.removeEventListener('device-telemetry-event', this.onTelemetryEvent);
  },
  methods: {
    async loadDevice() {
      const id = this.$route.params.id;
      try {
        const res = await this.$store.dispatch('fetchDeviceByID', id);
        this.device = res;
        this.loadActivity();
        this.fetchCommStatus();
        if (this.activeTab === 'telemetry') {
          this.fetchLatestTelemetry();
        } else if (this.activeTab === 'history') {
          this.fetchHistory();
        } else if (this.activeTab === 'raw') {
          this.fetchRawTelemetry();
        }
      } catch (err) {
        console.error('Failed to load device:', err);
      }
    },
    async loadActivity() {
      if (!this.device) return;
      this.activityLoading = true;
      try {
        const res = await this.$store.dispatch('fetchDeviceActivity', {
          deviceId: this.device.id,
          limit: 50,
        });
        this.activities = res || [];
      } catch (err) {
        console.error('Failed to load device activities:', err);
      } finally {
        this.activityLoading = false;
      }
    },
    startActivityPolling() {
      this.stopActivityPolling();
      this.activityPollTimer = setInterval(() => {
        if (this.activeTab === 'activity' && this.device) {
          this.loadActivity();
        }
      }, 3000);
    },
    stopActivityPolling() {
      if (this.activityPollTimer) {
        clearInterval(this.activityPollTimer);
        this.activityPollTimer = null;
      }
    },
    openEditModal() {
      this.isEditModalOpen = true;
    },
    openAddParamModal() {
      this.paramToEdit = null;
      this.isParamModalOpen = true;
    },
    openEditParamModal(param) {
      this.paramToEdit = param;
      this.isParamModalOpen = true;
    },
    async toggleEnabled() {
      if (!this.device) return;
      try {
        await this.$store.dispatch('toggleDeviceEnabled', {
          id: this.device.id,
          enabled: !this.device.enabled,
        });
        this.loadDevice();
      } catch (err) {
        alert('Failed to toggle device state: ' + err.message);
      }
    },
    async confirmDelete() {
      if (!confirm(`Are you sure you want to delete device ${this.device.device_code || this.device.code}? Historical data will be preserved.`)) {
        return;
      }
      try {
        await this.$store.dispatch('deleteDevice', this.device.id);
        this.$router.push('/monitoring/devices');
      } catch (err) {
        alert('Failed to delete device: ' + err.message);
      }
    },
    async toggleParamEnabled(param) {
      try {
        await this.$store.dispatch('toggleParameterEnabled', {
          deviceId: this.device.id,
          paramId: param.id,
          enabled: !param.enabled,
        });
        this.loadDevice();
      } catch (err) {
        alert('Failed to toggle parameter: ' + err.message);
      }
    },
    async confirmDeleteParam(param) {
      if (!confirm(`Delete parameter ${param.parameter_code || param.code}?`)) {
        return;
      }
      try {
        await this.$store.dispatch('deleteParameter', {
          deviceId: this.device.id,
          paramId: param.id,
        });
        this.loadDevice();
      } catch (err) {
        alert('Failed to delete parameter: ' + err.message);
      }
    },
    statusBadgeClass(status) {
      switch (status) {
        case 'ACTIVE': return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
        case 'INACTIVE': return 'bg-slate-500/10 text-slate-400 border border-slate-500/20';
        case 'MAINTENANCE': return 'bg-amber-500/10 text-amber-400 border border-amber-500/20';
        case 'DISABLED': return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
        default: return 'bg-slate-500/10 text-slate-400';
      }
    },
    commDotClass(status) {
      switch (status) {
        case 'ONLINE': return 'bg-emerald-400 animate-pulse';
        case 'CONNECTING': return 'bg-blue-400 animate-pulse';
        case 'ERROR': return 'bg-rose-400';
        case 'OFFLINE': return 'bg-slate-500';
        default: return 'bg-purple-400';
      }
    },
    commTextClass(status) {
      switch (status) {
        case 'ONLINE': return 'text-emerald-400';
        case 'CONNECTING': return 'text-blue-400';
        case 'ERROR': return 'text-rose-400';
        case 'OFFLINE': return 'text-slate-400';
        default: return 'text-purple-400';
      }
    },
    formatTimestamp(ts) {
      if (!ts) return '--';
      const d = new Date(ts);
      return d.toLocaleString('en-GB', {
        day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit', second: '2-digit',
      });
    },
    async fetchCommStatus() {
      if (!this.device) return;
      try {
        const res = await this.$store.dispatch('fetchCommunicationStatus', this.device.id);
        if (res) {
          this.commStatus = (res.data !== undefined && typeof res.data === 'object') ? res.data : res;
        }
      } catch (err) {
        // silent fail
      }
    },
    async handleConnect() {
      if (!this.device) return;
      this.commActionLoading = true;
      this.commFeedback = null;
      try {
        const res = await this.$store.dispatch('connectDevice', this.device.id);
        this.commFeedback = { type: 'success', message: res.message || 'Device communication connected' };
        await this.loadDevice();
      } catch (err) {
        this.commFeedback = { type: 'error', message: (err.response && err.response.data && err.response.data.error) || err.message };
      } finally {
        this.commActionLoading = false;
      }
    },
    async handleDisconnect() {
      if (!this.device) return;
      this.commActionLoading = true;
      this.commFeedback = null;
      try {
        const res = await this.$store.dispatch('disconnectDevice', this.device.id);
        this.commFeedback = { type: 'success', message: res.message || 'Device communication disconnected' };
        await this.loadDevice();
      } catch (err) {
        this.commFeedback = { type: 'error', message: (err.response && err.response.data && err.response.data.error) || err.message };
      } finally {
        this.commActionLoading = false;
      }
    },
    async handleReconnect() {
      if (!this.device) return;
      this.commActionLoading = true;
      this.commFeedback = null;
      try {
        const res = await this.$store.dispatch('reconnectDevice', this.device.id);
        this.commFeedback = { type: 'success', message: res.message || 'Device reconnection initiated' };
        await this.loadDevice();
      } catch (err) {
        this.commFeedback = { type: 'error', message: (err.response && err.response.data && err.response.data.error) || err.message };
      } finally {
        this.commActionLoading = false;
      }
    },
    async handleTestConnection() {
      if (!this.device) return;
      this.commActionLoading = true;
      this.commFeedback = null;
      try {
        const res = await this.$store.dispatch('testDeviceConnection', this.device.id);
        const data = res.data;
        if (data && data.connected) {
          this.commFeedback = { type: 'success', message: `Physical link verified: connected in ${data.latency_ms} ms (status: ${data.status})` };
        } else {
          this.commFeedback = { type: 'error', message: `Link check failed: ${data ? data.error : 'Connection error'}` };
        }
        await this.fetchCommStatus();
      } catch (err) {
        this.commFeedback = { type: 'error', message: (err.response && err.response.data && err.response.data.error) || err.message };
      } finally {
        this.commActionLoading = false;
      }
    },
    openTestReadModal(param) {
      this.testParam = param;
      this.testReadResult = null;
      this.isTestReadModalOpen = true;
      this.executeTestRead();
    },
    closeTestReadModal() {
      this.isTestReadModalOpen = false;
      this.testParam = null;
      this.testReadResult = null;
    },
    async executeTestRead() {
      if (!this.device || !this.testParam) return;
      this.testReadLoading = true;
      this.testReadResult = null;
      try {
        const res = await this.$store.dispatch('testReadParameter', {
          deviceId: this.device.id,
          paramId: this.testParam.id,
          parameterId: this.testParam.id,
        });
        this.testReadResult = (res && res.data && typeof res.data === 'object' && res.data.success !== undefined) ? res.data : res;
        if (this.testReadResult && this.testReadResult.success && this.testReadResult.decoded_value !== undefined) {
          this.testParam.current_value = this.testReadResult.decoded_value;
        }
      } catch (err) {
        this.testReadResult = {
          success: false,
          error: (err.response && err.response.data && err.response.data.error) || err.message,
          timestamp: new Date().toISOString(),
          response_time_ms: 0,
        };
      } finally {
        this.testReadLoading = false;
      }
    },
    onCommEvent(e) {
      if (!this.device || !e.detail || !e.detail.data) return;
      const data = e.detail.data;
      if (data.device_id === this.device.id) {
        if (data.connection_status) this.device.connection_status = data.connection_status;
        if (data.latency_ms !== undefined) this.device.latency_ms = data.latency_ms;
        if (data.last_seen_at) this.device.last_seen_at = data.last_seen_at;
        if (data.last_data_at) this.device.last_data_at = data.last_data_at;
        if (data.parameter_id && this.device.parameters) {
          const p = this.device.parameters.find(x => x.id === data.parameter_id);
          if (p && data.decoded_value !== undefined) {
            p.current_value = data.decoded_value;
          }
        }
        if (this.commStatus) {
          if (data.connection_status) this.commStatus.status = data.connection_status;
          if (data.latency_ms !== undefined) this.commStatus.latency_ms = data.latency_ms;
        }
      }
    },
    switchTab(tab) {
      this.activeTab = tab;
      if (tab === 'activity') {
        this.loadActivity();
        this.startActivityPolling();
      } else {
        this.stopActivityPolling();
      }
      if (tab === 'connection') {
        this.fetchCommStatus();
      } else if (tab === 'telemetry') {
        this.fetchLatestTelemetry();
      } else if (tab === 'history') {
        this.fetchHistory();
      } else if (tab === 'raw') {
        this.fetchRawTelemetry();
      }
    },
    async fetchLatestTelemetry() {
      if (!this.device) return;
      try {
        const res = await axios.get(`/api/devices/${this.device.id}/telemetry/latest`);
        if (res.data && res.data.data) {
          for (const item of res.data.data) {
            this.$set(this.liveTelemetry, item.parameter_id, item);
          }
        }
      } catch (err) {
        console.error('Failed to fetch latest telemetry:', err);
      }
    },
    onTelemetryEvent(e) {
      if (!this.device || !e.detail) return;
      const data = e.detail;
      if (data.device_id === this.device.id) {
        this.$set(this.liveTelemetry, data.parameter_id, data);
        this.$set(this.liveBlinks, data.parameter_id, true);
        setTimeout(() => {
          this.$set(this.liveBlinks, data.parameter_id, false);
        }, 1200);

        this.lastLivePacketTime = data.received_at;
        this.livePacketBlink = true;
        setTimeout(() => {
          this.livePacketBlink = false;
        }, 800);

        this.device.last_data_at = data.received_at;

        // Reactive stream into Raw Telemetry table if on page 1
        if (this.activeTab === 'raw' && this.rawPage === 1 && (!this.rawParamId || Number(this.rawParamId) === data.parameter_id)) {
          const rawItem = {
            id: Date.now(),
            device_id: data.device_id,
            parameter_id: data.parameter_id,
            value: data.value,
            raw_value: data.raw_value,
            raw_hex: data.raw_hex,
            quality: data.quality || 'GOOD',
            source: data.source || 'MODBUS_RTU',
            received_at: data.received_at,
          };
          this.rawRecords.unshift(rawItem);
          if (this.rawRecords.length > this.rawPageSize) {
            this.rawRecords.pop();
          }
          this.rawTotal++;
        }

        // Reactive stream into Historical Data table & SVG Trend Chart if on page 1
        if (this.activeTab === 'history' && this.historyPage === 1 && (!this.historyParamId || Number(this.historyParamId) === data.parameter_id)) {
          const histItem = {
            id: Date.now(),
            device_id: data.device_id,
            parameter_id: data.parameter_id,
            value: data.value,
            raw_value: data.raw_value,
            quality: data.quality || 'GOOD',
            source: data.source || 'MODBUS_RTU',
            received_at: data.received_at,
          };
          this.historyRecords.unshift(histItem);
          if (this.historyRecords.length > this.historyPageSize) {
            this.historyRecords.pop();
          }
          this.historyTotal++;
        }
      }
    },
    async fetchHistory() {
      if (!this.device) return;
      this.historyLoading = true;
      try {
        let startTimeStr = '';
        if (this.historyTimeRange !== 'all') {
          const now = new Date();
          let ms = 60 * 60 * 1000;
          if (this.historyTimeRange === '15m') ms = 15 * 60 * 1000;
          if (this.historyTimeRange === '6h') ms = 6 * 60 * 60 * 1000;
          if (this.historyTimeRange === '24h') ms = 24 * 60 * 60 * 1000;
          const start = new Date(now.getTime() - ms);
          startTimeStr = start.toISOString();
        }

        const params = {
          page: this.historyPage,
          page_size: this.historyPageSize,
        };
        if (this.historyParamId) params.parameter_id = this.historyParamId;
        if (this.historyQuality) params.quality = this.historyQuality;
        if (startTimeStr) params.start_time = startTimeStr;

        const res = await axios.get(`/api/devices/${this.device.id}/telemetry/history`, { params });
        if (res.data && res.data.data) {
          this.historyRecords = res.data.data.items || [];
          this.historyTotal = res.data.data.total || 0;
        }
      } catch (err) {
        console.error('Failed to query historical telemetry:', err);
      } finally {
        this.historyLoading = false;
      }
    },
    async fetchRawTelemetry() {
      if (!this.device) return;
      this.rawLoading = true;
      try {
        const params = {
          page: this.rawPage,
          page_size: this.rawPageSize,
        };
        if (this.rawParamId) params.parameter_id = this.rawParamId;

        const res = await axios.get(`/api/devices/${this.device.id}/telemetry/raw`, { params });
        if (res.data && res.data.data) {
          this.rawRecords = res.data.data.items || [];
          this.rawTotal = res.data.data.total || 0;
        }
      } catch (err) {
        console.error('Failed to query raw telemetry:', err);
      } finally {
        this.rawLoading = false;
      }
    },
    setTimeRange(range) {
      this.historyTimeRange = range;
      this.historyPage = 1;
      this.fetchHistory();
    },
    onFilterChange() {
      this.chartParamId = this.historyParamId;
      this.historyPage = 1;
      this.fetchHistory();
    },
    onPageSizeChange() {
      this.historyPage = 1;
      this.fetchHistory();
    },
    selectChartParam(paramId) {
      this.chartParamId = paramId;
      this.historyParamId = paramId;
      this.historyPage = 1;
      this.fetchHistory();
    },
    getParamName(paramId) {
      if (!this.device || !this.device.parameters) return `Parameter #${paramId}`;
      const p = this.device.parameters.find(x => x.id === paramId);
      return p ? (p.parameter_name || p.name || p.parameter_code) : `Parameter #${paramId}`;
    },
    formatTimeShort(ts) {
      if (!ts) return '';
      const d = new Date(ts);
      return d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    onChartMouseMove(e) {
      const container = this.$refs.chartContainer;
      if (!container || !this.chartPoints || this.chartPoints.length === 0) return;
      const rect = container.getBoundingClientRect();
      const mouseX = Math.max(0, Math.min(e.clientX - rect.left, rect.width));
      const mouseY = Math.max(0, Math.min(e.clientY - rect.top, rect.height));

      this.hoverMouseClientX = mouseX;
      this.hoverMouseClientY = mouseY;
      this.containerWidth = rect.width;
      this.containerHeight = rect.height;

      // Convert mouseX to SVG coordinate space (viewBox width 900)
      const svgX = (mouseX / rect.width) * 900;

      // Find nearest point along the X axis
      let nearest = this.chartPoints[0];
      let minDiff = Math.abs(nearest.x - svgX);

      for (let i = 1; i < this.chartPoints.length; i++) {
        const pt = this.chartPoints[i];
        const diff = Math.abs(pt.x - svgX);
        if (diff < minDiff) {
          minDiff = diff;
          nearest = pt;
        }
      }

      this.hoveredPoint = nearest;
    },
    onChartMouseLeave() {
      this.hoveredPoint = null;
    },
    inspectHistoryForParam(paramId) {
      this.historyParamId = paramId;
      this.chartParamId = paramId;
      this.activeTab = 'history';
      this.fetchHistory();
    },
    prevHistoryPage() {
      if (this.historyPage > 1) {
        this.historyPage--;
        this.fetchHistory();
      }
    },
    nextHistoryPage() {
      if (this.historyPage * this.historyPageSize < this.historyTotal) {
        this.historyPage++;
        this.fetchHistory();
      }
    },
    prevRawPage() {
      if (this.rawPage > 1) {
        this.rawPage--;
        this.fetchRawTelemetry();
      }
    },
    nextRawPage() {
      if (this.rawPage * this.rawPageSize < this.rawTotal) {
        this.rawPage++;
        this.fetchRawTelemetry();
      }
    },
    getParamCode(paramId) {
      if (!this.device || !this.device.parameters) return `#${paramId}`;
      const p = this.device.parameters.find(x => x.id === paramId);
      return p ? (p.parameter_code || p.code) : `#${paramId}`;
    },
    getParamUnit(paramId) {
      if (!this.device || !this.device.parameters) return '';
      const p = this.device.parameters.find(x => x.id === paramId);
      return p ? p.unit : '';
    },
    formatParamValue(param) {
      if (this.liveTelemetry[param.id] !== undefined) {
        const item = this.liveTelemetry[param.id];
        if (item.value_text) return item.value_text;
        if (item.value !== undefined) return item.value.toFixed(param.precision || 2);
      }
      if (param.current_value !== null && param.current_value !== undefined) {
        return param.current_value.toFixed(param.precision || 2);
      }
      return '--';
    },
    getParamQuality(param) {
      if (this.liveTelemetry[param.id] && this.liveTelemetry[param.id].quality) {
        return this.liveTelemetry[param.id].quality;
      }
      return param.current_quality || 'UNKNOWN';
    },
    getParamRaw(param) {
      if (this.liveTelemetry[param.id]) {
        const item = this.liveTelemetry[param.id];
        if (item.raw_hex) return item.raw_hex;
        if (item.raw_value !== undefined) return `Raw: ${item.raw_value}`;
      }
      return '--';
    },
    getParamTime(param) {
      if (this.liveTelemetry[param.id] && this.liveTelemetry[param.id].received_at) {
        return this.liveTelemetry[param.id].received_at;
      }
      return param.current_received_at || param.last_updated;
    },
    qualityBadgeClass(quality) {
      switch (quality) {
        case 'GOOD': return 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20';
        case 'BAD': return 'bg-rose-500/10 text-rose-400 border border-rose-500/20';
        case 'UNCERTAIN': return 'bg-amber-500/10 text-amber-400 border border-amber-500/20';
        default: return 'bg-slate-500/10 text-slate-400 border border-slate-500/20';
      }
    },
    formatAuditActionName(action) {
      if (!action) return 'Aktivitas';
      switch (action) {
        case 'UPDATE_PARAMETER': return 'Parameter Diperbarui';
        case 'CREATE_PARAMETER': return 'Parameter Baru Ditambahkan';
        case 'DELETE_PARAMETER': return 'Parameter Dihapus';
        case 'UPDATE_DEVICE': return 'Konfigurasi Perangkat Diperbarui';
        case 'CREATE_DEVICE': return 'Perangkat Baru Dibuat';
        case 'DELETE_DEVICE': return 'Perangkat Dihapus';
        case 'COMMUNICATION_ERROR': return 'Koneksi Terputus / Error';
        case 'COMMUNICATION_RESTORED': return 'Koneksi Pulih (Online)';
        case 'CONNECT_DEVICE': return 'Perangkat Dihubungkan';
        case 'DISCONNECT_DEVICE': return 'Perangkat Diputuskan';
        default: return action.replace(/_/g, ' ');
      }
    },
    getAuditActionClass(action) {
      if (!action) return 'bg-slate-500/10 text-slate-300 border border-slate-500/20';
      if (action.includes('ERROR') || action.includes('DELETE')) {
        return 'bg-rose-500/15 text-rose-400 border border-rose-500/30';
      }
      if (action.includes('RESTORED') || action.includes('CREATE')) {
        return 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30';
      }
      if (action.includes('UPDATE') || action.includes('CONNECT')) {
        return 'bg-amber-500/15 text-amber-400 border border-amber-500/30';
      }
      return 'bg-blue-500/15 text-blue-400 border border-blue-500/30';
    },
    getAuditCardBorder(action) {
      if (!action) return 'border-slate-800/80';
      if (action.includes('ERROR')) return 'border-rose-900/50 bg-rose-950/10';
      if (action.includes('RESTORED')) return 'border-emerald-900/50 bg-emerald-950/10';
      return 'border-slate-800/80';
    },
    parseSafeJson(val) {
      if (!val) return null;
      if (typeof val === 'object') return val;
      if (typeof val === 'string') {
        try {
          const res = JSON.parse(val);
          if (typeof res === 'string') {
            return this.parseSafeJson(res);
          }
          return res;
        } catch (e) {
          return null;
        }
      }
      return null;
    },
    formatAuditDescription(trail) {
      if (!trail) return '';
      if (trail.action === 'COMMUNICATION_ERROR') {
        return trail.details || 'Komunikasi serial/network gagal atau port terputus. Sistem terus mencoba rekoneksi otomatis.';
      }
      if (trail.action === 'COMMUNICATION_RESTORED') {
        return 'Koneksi komunikasi pulih dan stream data sensor kembali normal.';
      }
      const parsed = this.parseSafeJson(trail.details);
      if (parsed) {
        const after = this.parseSafeJson(parsed.after);
        if (after && typeof after === 'object') {
          if (trail.action.includes('PARAMETER')) {
            const pName = after.parameter_name || after.name || after.parameter_code || 'Parameter';
            const pCode = after.parameter_code || after.code || '';
            return `Memperbarui konfigurasi sensor: "${pName}" (${pCode}).`;
          }
          if (trail.action.includes('DEVICE')) {
            const dName = after.device_name || after.name || after.device_code || 'Perangkat';
            return `Memperbarui profil perangkat: "${dName}".`;
          }
        }
      }
      return trail.details || 'Aktivitas konfigurasi dicatat.';
    },
    getAuditChips(trail) {
      if (!trail || !trail.details) return [];
      const chips = [];
      const parsed = this.parseSafeJson(trail.details);
      if (parsed) {
        const after = this.parseSafeJson(parsed.after) || parsed;
        if (after && typeof after === 'object') {
          if (after.parameter_code) chips.push({ label: 'Kode', val: after.parameter_code });
          if (after.parameter_name) chips.push({ label: 'Nama', val: after.parameter_name });
          if (after.data_type) chips.push({ label: 'Tipe Data', val: after.data_type });
          if (after.register_address !== undefined) chips.push({ label: 'Register', val: `#${after.register_address}` });
          if (after.scale !== undefined) chips.push({ label: 'Scale', val: after.scale });
          if (after.offset !== undefined) chips.push({ label: 'Offset', val: after.offset });
          if (after.unit) chips.push({ label: 'Satuan', val: after.unit });
          if (after.enabled !== undefined) chips.push({ label: 'Status', val: after.enabled ? 'Aktif' : 'Non-Aktif' });
          if (after.device_name) chips.push({ label: 'Nama Perangkat', val: after.device_name });
          if (after.device_code) chips.push({ label: 'Kode Perangkat', val: after.device_code });
          if (after.status) chips.push({ label: 'Status Admin', val: after.status });
          if (after.serial_port) chips.push({ label: 'Port', val: after.serial_port });
          if (after.baud_rate) chips.push({ label: 'Baud', val: after.baud_rate });
        }
      }
      return chips;
    },
    hasJsonDetails(details) {
      if (!details || typeof details !== 'string') return false;
      const s = details.trim();
      return (s.startsWith('{') && s.endsWith('}')) || (s.startsWith('[') && s.endsWith(']'));
    },
    formatRawJson(details) {
      try {
        const parsed = JSON.parse(details);
        return JSON.stringify(parsed, null, 2);
      } catch (e) {
        return details;
      }
    },
    toggleRawAudit(id) {
      this.$set(this.showRawAudit, id, !this.showRawAudit[id]);
    },
  },
};
</script>
