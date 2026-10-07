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
          </div>
          <p class="text-2xs text-slate-400 mt-0.5">Physical serial & socket lifecycle, timeout, retry backoff, and register polling engine</p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <!-- Connect Button -->
          <button
            @click="handleConnect"
            :disabled="commActionLoading"
            class="px-3 py-1.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
            </svg>
            <span>Connect</span>
          </button>

          <!-- Disconnect Button -->
          <button
            @click="handleDisconnect"
            :disabled="commActionLoading"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"></path>
            </svg>
            <span>Disconnect</span>
          </button>

          <!-- Reconnect Button -->
          <button
            @click="handleReconnect"
            :disabled="commActionLoading"
            class="px-3 py-1.5 rounded-xl bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 border border-blue-500/30 text-xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>Reconnect</span>
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
      <div class="flex items-center justify-between border-b border-slate-800 pb-3">
        <div>
          <h3 class="text-sm font-bold text-white">Device Audit Trail</h3>
          <p class="text-2xs text-slate-400 mt-0.5">Immutable historical record of configuration updates and status changes</p>
        </div>
      </div>

      <div class="space-y-2">
        <div v-for="trail in activities" :key="trail.id" class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800 flex items-start justify-between text-xs">
          <div>
            <div class="flex items-center space-x-2">
              <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
                {{ trail.action }}
              </span>
              <span class="text-slate-400 text-3xs">by {{ trail.username }}</span>
            </div>
            <p class="text-slate-300 mt-1 font-sans text-xs break-all">{{ trail.details }}</p>
          </div>
          <span class="text-slate-500 font-mono text-3xs flex-shrink-0 ml-4">{{ formatTimestamp(trail.created_at) }}</span>
        </div>
        <div v-if="activities.length === 0" class="text-center py-8 text-xs text-slate-400">
          No audit entries recorded for this device yet.
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
    };
  },
  mounted() {
    this.loadDevice();
    window.addEventListener('device-comm-event', this.onCommEvent);
  },
  beforeDestroy() {
    window.removeEventListener('device-comm-event', this.onCommEvent);
  },
  methods: {
    async loadDevice() {
      const id = this.$route.params.id;
      try {
        const res = await this.$store.dispatch('fetchDeviceByID', id);
        this.device = res;
        this.loadActivity();
        this.fetchCommStatus();
      } catch (err) {
        console.error('Failed to load device:', err);
      }
    },
    async loadActivity() {
      if (!this.device) return;
      try {
        const res = await this.$store.dispatch('fetchDeviceActivity', {
          deviceId: this.device.id,
          limit: 30,
        });
        this.activities = res || [];
      } catch (err) {
        console.error('Failed to load device activities:', err);
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
        if (res && res.data) {
          this.commStatus = res.data;
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
          parameterId: this.testParam.id,
        });
        this.testReadResult = res.data;
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
  },
};
</script>
