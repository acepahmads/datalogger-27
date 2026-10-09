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
            <span>{{ $t('deviceDetail.devicesAndSensors') }}</span>
          </router-link>
          <span>/</span>
          <span class="text-slate-200 font-mono font-semibold">{{ device ? (device.device_code || device.code) : 'Device' }}</span>
        </div>
        <div class="flex items-center space-x-3">
          <h1 class="text-xl font-bold tracking-tight text-white">
            {{ device ? (device.device_name || device.name) : $t('deviceDetail.loadingDevice') }}
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
          <span>{{ device.enabled ? $t('deviceDetail.disableDevice') : $t('deviceDetail.enableDevice') }}</span>
        </button>

        <button
          @click="openEditModal"
          class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
          </svg>
          <span>{{ $t('common.edit') }}</span>
        </button>

        <button
          @click="confirmDelete"
          class="px-3 py-1.5 rounded-xl bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 border border-rose-500/20 text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
          </svg>
          <span>{{ $t('common.delete') }}</span>
        </button>

        <router-link
          v-if="device"
          :to="'/monitoring/analysis?device_id=' + device.id"
          class="px-3 py-1.5 rounded-xl bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-400 border border-indigo-500/30 text-xs font-semibold transition flex items-center space-x-1.5"
          :title="$t('analysis.openInAnalysis')"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
          </svg>
          <span>{{ $t('analysis.breadcrumb') }}</span>
        </router-link>
      </div>
    </div>

    <!-- Status Cards Header -->
    <div v-if="device" class="grid grid-cols-2 md:grid-cols-5 gap-3">
      <!-- Admin Status -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('deviceDetail.adminStatus') }}</div>
        <div class="flex items-center space-x-2">
          <span class="px-2 py-0.5 rounded-md text-2xs font-bold font-mono" :class="statusBadgeClass(device.status)">
            {{ device.status }}
          </span>
        </div>
      </div>

      <!-- Comm Status -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('deviceDetail.commStatus') }}</div>
        <div class="flex items-center space-x-1.5">
          <span class="w-2 h-2 rounded-full" :class="commDotClass(device.connection_status)"></span>
          <span class="text-xs font-mono font-bold" :class="commTextClass(device.connection_status)">
            {{ device.connection_status || 'UNKNOWN' }}
          </span>
        </div>
      </div>

      <!-- Protocol -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('deviceDetail.protocol') }}</div>
        <div class="text-xs font-mono font-bold text-blue-400">
          {{ (device.connection && device.connection.protocol) || device.device_type || 'MODBUS_TCP' }}
        </div>
      </div>

      <!-- Last Seen -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('deviceDetail.lastComm') }}</div>
        <div class="text-xs font-mono text-slate-200">
          {{ formatTimestamp(device.last_seen_at || device.last_communication) }}
        </div>
      </div>

      <!-- Last Data -->
      <div class="saas-card p-3 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('deviceDetail.lastDataPacket') }}</div>
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
        <span>{{ $t('deviceDetail.overview') }}</span>
      </button>
      <button
        @click="activeTab = 'connection'"
        :class="activeTab === 'connection' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>{{ $t('deviceDetail.connectionConfig') }}</span>
      </button>
      <button
        @click="activeTab = 'parameters'"
        :class="activeTab === 'parameters' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>{{ $t('deviceDetail.parameters') }}</span>
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
        <span>{{ $t('deviceDetail.liveTelemetry') }}</span>
      </button>
      <button
        @click="switchTab('history')"
        :class="activeTab === 'history' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>{{ $t('deviceDetail.historical') }}</span>
      </button>
      <button
        @click="switchTab('raw')"
        :class="activeTab === 'raw' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>{{ $t('deviceDetail.rawTelemetry') }}</span>
      </button>
      <button
        @click="switchTab('aggregation')"
        :class="activeTab === 'aggregation' ? 'border-b-2 border-indigo-500 text-indigo-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>{{ $t('deviceDetail.aggregation') }}</span>
        <span v-if="aggregationDefinitions && aggregationDefinitions.length" class="text-[10px] bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 px-1.5 py-0.5 rounded-full font-mono">{{ aggregationDefinitions.length }}</span>
      </button>
      <button
        @click="activeTab = 'activity'"
        :class="activeTab === 'activity' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>{{ $t('deviceDetail.activityAudit') }}</span>
      </button>
      <button
        @click="activeTab = 'health'"
        :class="activeTab === 'health' ? 'border-b-2 border-blue-500 text-blue-400 pb-2.5' : 'text-slate-400 hover:text-slate-200 pb-2.5'"
        class="transition-colors flex items-center space-x-1.5"
      >
        <span>{{ $t('deviceDetail.healthDiagnostics') }}</span>
      </button>
    </div>

    <!-- TAB 1: OVERVIEW -->
    <div v-if="device && activeTab === 'overview'" class="grid grid-cols-1 md:grid-cols-2 gap-6">
      <!-- Hardware Profile -->
      <div class="saas-card p-5 space-y-4">
        <h3 class="text-xs font-bold text-white uppercase tracking-wider border-b border-slate-800/80 pb-2.5">
          {{ $t('deviceDetail.equipmentProfile') }}
        </h3>
        <dl class="grid grid-cols-2 gap-3 text-xs">
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.manufacturer') }}</dt>
            <dd class="text-slate-200 font-medium mt-0.5">{{ device.manufacturer || '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.model') }}</dt>
            <dd class="text-slate-200 font-medium mt-0.5">{{ device.model || '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.serialNumber') }}</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.serial_number || '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.firmwareVersion') }}</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.firmware_version || '--' }}</dd>
          </div>
          <div class="col-span-2">
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.description') }}</dt>
            <dd class="text-slate-300 mt-0.5">{{ device.description || $t('deviceDetail.noDescription') }}</dd>
          </div>
        </dl>
      </div>

      <!-- Location & Metadata -->
      <div class="saas-card p-5 space-y-4">
        <h3 class="text-xs font-bold text-white uppercase tracking-wider border-b border-slate-800/80 pb-2.5">
          {{ $t('deviceDetail.locationGeography') }}
        </h3>
        <dl class="grid grid-cols-2 gap-3 text-xs">
          <div class="col-span-2">
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.physicalLocation') }}</dt>
            <dd class="text-slate-200 font-medium mt-0.5">{{ device.location || $t('deviceDetail.notSpecified') }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.latitude') }}</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.latitude !== null && device.latitude !== undefined ? device.latitude : '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.longitude') }}</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.longitude !== null && device.longitude !== undefined ? device.longitude : '--' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.timezone') }}</dt>
            <dd class="text-slate-200 font-mono text-2xs mt-0.5">{{ device.timezone || 'UTC' }}</dd>
          </div>
          <div>
            <dt class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.createdAt') }}</dt>
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
            <h3 class="text-sm font-bold text-white">{{ $t('deviceDetail.commEngineTitle') }}</h3>
            <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
              {{ $t('deviceDetail.phase22Active') }}
            </span>
            <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
              {{ $t('deviceDetail.autoConnectActive') }}
            </span>
          </div>
          <p class="text-2xs text-slate-400 mt-0.5">{{ $t('deviceDetail.commSubtitle') }}</p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <!-- Force Reconnect (Manual Override) -->
          <button
            @click="handleReconnect"
            :disabled="commActionLoading"
            class="px-3 py-1.5 rounded-xl bg-blue-600/20 hover:bg-blue-600/30 text-blue-400 border border-blue-500/30 text-xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
            :title="$t('deviceDetail.reconnectTooltip')"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>{{ $t('deviceDetail.reconnectNow') }}</span>
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
            <span>{{ $t('deviceDetail.testLink') }}</span>
          </button>

          <!-- Edit Config -->
          <button
            @click="openEditModal"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 text-xs font-semibold transition"
          >
            {{ $t('deviceDetail.editSettings') }}
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
              <span>{{ $t('deviceDetail.autoConnectBannerTitle') }}</span>
              <span class="px-1.5 py-0.5 rounded text-3xs font-mono font-medium uppercase bg-slate-800 text-slate-300">
                {{ $t('deviceDetail.noManualConnectNeeded') }}
              </span>
            </div>
            <p class="text-3xs opacity-85 mt-0.5">
              {{ ((commStatus && commStatus.status) || device.connection_status) === 'ONLINE' ? $t('deviceDetail.autoConnectOnlineDesc') : $t('deviceDetail.autoConnectOfflineDesc') }}
            </p>
          </div>
        </div>
        <div class="flex items-center space-x-3 text-3xs font-mono text-slate-300 flex-shrink-0">
          <span>{{ $t('deviceDetail.interval') }} {{ (device.connection && device.connection.polling_interval) || 1000 }}ms</span>
          <span>•</span>
          <span>{{ $t('deviceDetail.autoRecoveryOn') }}</span>
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
        <button @click="commFeedback = null" class="text-slate-400 hover:text-white text-3xs font-mono">{{ $t('deviceDetail.dismiss') }}</button>
      </div>

      <!-- Live Engine Diagnostics Strip -->
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-1">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.engineLifecycle') }}</div>
          <div class="flex items-center space-x-2">
            <span class="w-2 h-2 rounded-full" :class="commDotClass(commStatus ? commStatus.status : device.connection_status)"></span>
            <span class="text-xs font-mono font-bold" :class="commTextClass(commStatus ? commStatus.status : device.connection_status)">
              {{ (commStatus && commStatus.status) || device.connection_status || 'UNKNOWN' }}
            </span>
          </div>
        </div>

        <div class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-1">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.roundtripLatency') }}</div>
          <div class="text-xs font-mono font-bold text-emerald-400">
            {{ commStatus ? commStatus.latency_ms : (device.latency_ms || 0) }} ms
          </div>
        </div>

        <div class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-1">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.consecutiveRetries') }}</div>
          <div class="text-xs font-mono font-bold text-slate-200">
            {{ commStatus ? commStatus.consecutive_errors : 0 }} / {{ (device.connection && device.connection.retry_count) || 3 }} {{ $t('deviceDetail.max') }}
          </div>
        </div>

        <div class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-1">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.lastCommError') }}</div>
          <div class="text-xs font-mono truncate text-rose-400" :title="(commStatus && commStatus.last_error) || $t('deviceDetail.none')">
            {{ (commStatus && commStatus.last_error) || $t('deviceDetail.none') }}
          </div>
        </div>
      </div>

      <!-- Detailed Configuration Grid -->
      <div v-if="device.connection" class="grid grid-cols-2 md:grid-cols-4 gap-4 text-xs font-sans">
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.protocol') }}</div>
          <div class="text-xs font-mono font-bold text-blue-400 mt-1">{{ device.connection.protocol }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.connectionType') }}</div>
          <div class="text-xs font-mono text-slate-200 mt-1">{{ device.connection.connection_type }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.slaveIdUnit') }}</div>
          <div class="text-xs font-mono font-bold text-amber-400 mt-1">#{{ device.connection.slave_id || 1 }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.byteWordOrder') }}</div>
          <div class="text-xs font-mono font-bold text-indigo-400 mt-1">{{ device.connection.byte_order || 'ABCD' }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.hostAddress') }}</div>
          <div class="text-xs font-mono text-slate-200 mt-1">{{ device.connection.host || device.connection.address || '--' }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.port') }}</div>
          <div class="text-xs font-mono text-slate-200 mt-1">{{ device.connection.port || '--' }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.serialPort') }}</div>
          <div class="text-xs font-mono text-slate-200 mt-1">{{ device.connection.serial_port || '--' }}</div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.baudAndFraming') }}</div>
          <div class="text-xs font-mono text-slate-200 mt-1">
            {{ device.connection.baud_rate || 9600 }} ({{ device.connection.data_bits || 8 }}-{{ device.connection.parity || 'N' }}-{{ device.connection.stop_bits || 1 }})
          </div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.timeout') }}</div>
          <div class="text-xs font-mono text-slate-200 mt-1">
            {{ device.connection.timeout || device.connection.timeout_ms || 1000 }} ms
          </div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.retryPolicy') }}</div>
          <div class="text-xs font-mono text-slate-200 mt-1">
            {{ $t('deviceDetail.retriesWithBackoff', { count: device.connection.retry_count || 3 }) }}
          </div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.pollingInterval') }}</div>
          <div class="text-xs font-mono text-slate-200 mt-1">
            {{ device.connection.polling_interval || device.connection.poll_interval_ms || 1000 }} ms
          </div>
        </div>
        <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800">
          <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.adapterStatus') }}</div>
          <div class="text-xs font-mono text-emerald-400 mt-1">
            {{ (commStatus && commStatus.adapter_status) || 'ACTIVE' }}
          </div>
        </div>
      </div>
      <div v-else class="text-center py-8 text-xs text-slate-400">
        {{ $t('deviceDetail.noConnConfig') }}
      </div>
    </div>

    <!-- TAB 3: PARAMETERS -->
    <div v-if="device && activeTab === 'parameters'" class="saas-card p-5 space-y-4">
      <div class="flex items-center justify-between border-b border-slate-800 pb-3">
        <div>
          <h3 class="text-sm font-bold text-white">{{ $t('deviceDetail.telemetryParamsCount', { count: device.parameters ? device.parameters.length : 0 }) }}</h3>
          <p class="text-2xs text-slate-400 mt-0.5">{{ $t('deviceDetail.paramDefSubtitle') }}</p>
        </div>
        <button
          @click="openAddParamModal"
          class="px-3.5 py-1.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-semibold transition flex items-center space-x-1.5"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
          </svg>
          <span>{{ $t('deviceDetail.addParam') }}</span>
        </button>
      </div>

      <!-- Parameters Table -->
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs font-sans">
          <thead>
            <tr class="border-b border-slate-800 text-3xs uppercase font-semibold text-slate-400 tracking-wider">
              <th class="py-2.5 px-3">{{ $t('deviceDetail.code') }}</th>
              <th class="py-2.5 px-3">{{ $t('deviceDetail.colName') }}</th>
              <th class="py-2.5 px-3">{{ $t('deviceDetail.dataType') }}</th>
              <th class="py-2.5 px-3">{{ $t('deviceDetail.unit') }}</th>
              <th class="py-2.5 px-3">{{ $t('deviceDetail.register') }}</th>
              <th class="py-2.5 px-3">{{ $t('deviceDetail.colScaleOffset') }}</th>
              <th class="py-2.5 px-3">{{ $t('deviceDetail.colLimits') }}</th>
              <th class="py-2.5 px-3">{{ $t('deviceDetail.colCurrentValue') }}</th>
              <th class="py-2.5 px-3">{{ $t('deviceDetail.colEnabled') }}</th>
              <th class="py-2.5 px-3 text-right">{{ $t('deviceDetail.colActions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60">
            <tr v-for="param in device.parameters" :key="param.id" class="hover:bg-slate-800/30 transition-colors">
              <td class="py-2.5 px-3 font-mono font-bold text-blue-400">{{ param.parameter_code || param.code }}</td>
              <td class="py-2.5 px-3 font-medium text-slate-200">{{ param.parameter_name || param.name }}</td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-300">{{ param.data_type }}</td>
              <td class="py-2.5 px-3 font-medium text-slate-400">{{ param.unit || '--' }}</td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-300">
                <span class="font-bold text-slate-200">#{{ param.register_address !== undefined ? param.register_address : 0 }}</span>
                <span class="block text-4xs text-slate-500 font-sans uppercase">{{ param.register_type || 'HOLDING' }}</span>
              </td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-400">
                <div>×{{ param.scale || param.scale_factor || 1.0 }} + {{ param.offset || 0 }}</div>
                <div v-if="param.formula" class="text-blue-400 font-mono text-3xs truncate max-w-[120px]" :title="'Formula: ' + param.formula">
                  ƒ: {{ param.formula }}
                </div>
                <div v-if="param.hold_last_value_enabled" class="text-amber-400/90 text-3xs flex items-center space-x-1" :title="$t('parameters.heldTooltip')">
                  <span>🛡️</span>
                  <span>{{ param.hold_last_value_seconds || 120 }}s</span>
                </div>
              </td>
              <td class="py-2.5 px-3 font-mono text-3xs text-slate-400">
                <div>Hard: {{ param.min_value !== null && param.min_value !== undefined ? param.min_value : (param.low_limit || '-∞') }} .. {{ param.max_value !== null && param.max_value !== undefined ? param.max_value : (param.high_limit || '+∞') }}</div>
                <div v-if="param.warning_low !== null && param.warning_low !== undefined || param.warning_high !== null && param.warning_high !== undefined" class="text-amber-400/80">
                  Warn: {{ param.warning_low !== null && param.warning_low !== undefined ? param.warning_low : '-∞' }} .. {{ param.warning_high !== null && param.warning_high !== undefined ? param.warning_high : '+∞' }}
                </div>
              </td>
              <td class="py-2.5 px-3 font-mono text-slate-200">
                <div class="font-bold flex items-center space-x-1.5">
                  <span>{{ formatParamValue(param) }}</span>
                  <span
                    class="px-1.5 py-0.2 rounded text-4xs font-mono font-bold uppercase"
                    :class="qualityBadgeClass(getParamQuality(param))"
                  >
                    {{ getParamQuality(param) }}
                  </span>
                  <span v-if="isParamHeld(param)" class="px-1.5 py-0.5 rounded text-3xs font-mono font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30 animate-pulse" :title="$t('parameters.heldTooltip')">
                    {{ $t('parameters.heldBadge') }}
                  </span>
                </div>
                <div v-if="getParamQuality(param) !== 'GOOD' && getParamQualityReason(param)" class="text-4xs font-mono text-amber-400/90 truncate max-w-[140px]" :title="getParamQualityReason(param)">
                  {{ getParamQualityReason(param) }}
                </div>
                <div v-if="getParamFormulaValue(param) !== null" class="text-3xs text-blue-400 font-mono font-medium">
                  ƒ(x): {{ getParamFormulaValue(param) }}
                </div>
              </td>
              <td class="py-2.5 px-3">
                <button
                  @click="toggleParamEnabled(param)"
                  :class="param.enabled ? 'text-emerald-400 hover:text-emerald-300' : 'text-slate-500 hover:text-slate-400'"
                  class="font-mono text-3xs font-semibold transition"
                >
                  {{ param.enabled ? $t('deviceDetail.active') : $t('deviceDetail.disabled') }}
                </button>
              </td>
              <td class="py-2.5 px-3 text-right space-x-2">
                <button
                  @click="openTestReadModal(param)"
                  class="text-emerald-400 hover:text-emerald-300 transition text-2xs font-semibold px-2 py-0.5 rounded bg-emerald-500/10 hover:bg-emerald-500/20 border border-emerald-500/20"
                >
                  {{ $t('deviceDetail.testRead') }}
                </button>
                <button @click="openEditParamModal(param)" class="text-blue-400 hover:text-blue-300 transition text-2xs font-semibold">
                  {{ $t('common.edit') }}
                </button>
                <button @click="confirmDeleteParam(param)" class="text-rose-400 hover:text-rose-300 transition text-2xs font-semibold">
                  {{ $t('common.delete') }}
                </button>
              </td>
            </tr>
            <tr v-if="!device.parameters || device.parameters.length === 0">
              <td colspan="10" class="py-8 text-center text-xs text-slate-400">
                {{ $t('deviceDetail.noParamsConfigured') }}
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
            <h3 class="text-sm font-bold text-white">{{ $t('deviceDetail.activityAuditTitle') }}</h3>
            <span class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
              {{ $t('deviceDetail.liveAuditStream') }}
            </span>
          </div>
          <p class="text-2xs text-slate-400 mt-0.5">{{ $t('deviceDetail.activityAuditSubtitle') }}</p>
        </div>

        <div class="flex items-center space-x-3 text-xs">
          <span class="text-3xs text-slate-400 font-mono flex items-center gap-1.5">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
            {{ $t('deviceDetail.autoSync3s') }}
          </span>
          <button
            @click="loadActivity"
            :disabled="activityLoading"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-2xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
          >
            <svg class="w-3.5 h-3.5" :class="{ 'animate-spin': activityLoading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>{{ $t('deviceDetail.refreshLog') }}</span>
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
                {{ $t('deviceDetail.byOperator') }} <strong class="text-slate-200 font-semibold">{{ trail.username || 'System' }}</strong>
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
                <span>{{ showRawAudit[trail.id] ? $t('deviceDetail.hideJsonDetails') : $t('deviceDetail.showJsonDetails') }}</span>
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
          <span class="font-medium text-slate-300">{{ $t('deviceDetail.noActivityRecorded') }}</span>
          <p class="text-3xs text-slate-500 mt-0.5">{{ $t('deviceDetail.noActivitySub') }}</p>
        </div>
      </div>
    </div>

    <!-- TAB 5: HEALTH DIAGNOSTICS -->
    <div v-if="device && activeTab === 'health'" class="saas-card p-6 space-y-6">
      <div class="border-b border-slate-800 pb-3">
        <h3 class="text-sm font-bold text-white">{{ $t('deviceDetail.healthFoundation') }}</h3>
        <p class="text-2xs text-slate-400 mt-0.5">{{ $t('deviceDetail.healthSubtitle') }}</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2">
          <div class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.commStatus') }}</div>
          <div class="text-sm font-mono font-bold" :class="commTextClass(device.connection_status)">
            {{ device.connection_status || 'UNKNOWN' }}
          </div>
          <p class="text-3xs text-slate-400">{{ $t('deviceDetail.trackedBySupervisor') }}</p>
        </div>

        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2">
          <div class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.observedLatency') }}</div>
          <div class="text-sm font-mono font-bold text-emerald-400">
            {{ device.latency_ms || 0 }} ms
          </div>
          <p class="text-3xs text-slate-400">{{ $t('deviceDetail.roundtripDelay') }}</p>
        </div>

        <div class="p-4 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2">
          <div class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.packetsDelivery') }}</div>
          <div class="text-sm font-mono font-bold text-slate-200">
            {{ device.success_count || 0 }} ok / {{ device.failed_count || 0 }} err
          </div>
          <p class="text-3xs text-slate-400">{{ $t('deviceDetail.accumulatedPackets') }}</p>
        </div>
      </div>

      <div class="p-4 rounded-xl bg-blue-500/10 border border-blue-500/20 text-xs text-blue-300 space-y-1">
        <div class="font-bold flex items-center space-x-1.5">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
          </svg>
          <span>{{ $t('deviceDetail.healthArchitectureTitle') }}</span>
        </div>
        <p class="text-3xs text-blue-200">
          {{ $t('deviceDetail.healthArchitectureDesc') }}
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
              <h3 class="text-xs font-bold text-white uppercase tracking-wider">{{ $t('telemetry.liveStream') }}</h3>
              <span class="px-2 py-0.5 rounded-full text-3xs font-mono font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                {{ $t('telemetry.websocketLive') }}
              </span>
            </div>
            <p class="text-2xs text-slate-400 mt-0.5">
              {{ $t('telemetry.liveStreamDesc') }}
            </p>
          </div>
        </div>

        <div class="flex items-center space-x-4 text-xs font-mono">
          <div class="text-right">
            <span class="text-3xs text-slate-500 uppercase block">{{ $t('telemetry.lastPacket') }}</span>
            <span class="text-slate-200">{{ formatTimestamp(lastLivePacketTime || device.last_data_at) }}</span>
          </div>
          <button
            @click="fetchLatestTelemetry"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-2xs font-semibold transition flex items-center space-x-1.5"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
            </svg>
            <span>{{ $t('common.refresh') }}</span>
          </button>
        </div>
      </div>

      <!-- Device Quality Summary Bar (Phase 3.2) -->
      <div v-if="device && device.parameters && device.parameters.length > 0" class="grid grid-cols-2 sm:grid-cols-4 gap-3">
        <div class="saas-card p-3 space-y-1 bg-[#0F172A]/70 border border-emerald-500/20">
          <div class="text-3xs uppercase tracking-wider font-semibold text-emerald-400">{{ $t('telemetry.qualityGood') }}</div>
          <div class="flex items-baseline space-x-2">
            <span class="text-xl font-mono font-bold text-emerald-400">{{ deviceQualitySummary.good }}</span>
            <span class="text-2xs text-slate-400">/ {{ device.parameters.length }}</span>
          </div>
          <div class="text-3xs text-emerald-400/80">{{ deviceQualitySummary.healthPercent }}% {{ $t('telemetry.healthySensors') }}</div>
        </div>

        <div class="saas-card p-3 space-y-1 bg-[#0F172A]/70 border border-amber-500/20">
          <div class="text-3xs uppercase tracking-wider font-semibold text-amber-400">{{ $t('telemetry.qualityUncertain') }}</div>
          <div class="flex items-baseline space-x-2">
            <span class="text-xl font-mono font-bold text-amber-400">{{ deviceQualitySummary.uncertain }}</span>
          </div>
          <div class="text-3xs text-amber-400/80">{{ $t('telemetry.warningLimits') }}</div>
        </div>

        <div class="saas-card p-3 space-y-1 bg-[#0F172A]/70 border border-rose-500/20">
          <div class="text-3xs uppercase tracking-wider font-semibold text-rose-400">{{ $t('telemetry.qualityBad') }}</div>
          <div class="flex items-baseline space-x-2">
            <span class="text-xl font-mono font-bold text-rose-400">{{ deviceQualitySummary.bad }}</span>
          </div>
          <div class="text-3xs text-rose-400/80">{{ $t('telemetry.hardViolations') }}</div>
        </div>

        <div class="saas-card p-3 space-y-1 bg-[#0F172A]/70 border border-purple-500/20">
          <div class="text-3xs uppercase tracking-wider font-semibold text-purple-400">{{ $t('telemetry.qualityStale') }}</div>
          <div class="flex items-baseline space-x-2">
            <span class="text-xl font-mono font-bold text-purple-400">{{ deviceQualitySummary.stale }}</span>
          </div>
          <div class="text-3xs text-purple-400/80">{{ $t('telemetry.timedOutSensors') }}</div>
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
            <div class="flex flex-col items-end space-y-0.5">
              <span
                class="px-2 py-0.5 rounded text-3xs font-mono font-bold uppercase"
                :class="qualityBadgeClass(getParamQuality(param))"
              >
                {{ getParamQuality(param) }}
              </span>
              <span v-if="getParamQualityReason(param)" class="text-4xs font-mono text-amber-400/90 truncate max-w-[120px]" :title="getParamQualityReason(param)">
                {{ getParamQualityReason(param) }}
              </span>
            </div>
          </div>

          <!-- Primary Metric Value -->
          <div class="py-2 space-y-1">
            <div class="flex items-baseline space-x-2">
              <span class="text-3xl font-mono font-bold text-white tracking-tight">
                {{ formatParamValue(param) }}
              </span>
              <span class="text-sm font-medium text-slate-400 font-mono">{{ param.unit }}</span>
              <span
                v-if="isParamHeld(param)"
                class="px-1.5 py-0.5 rounded text-3xs font-mono font-bold bg-amber-500/20 text-amber-300 border border-amber-500/30 animate-pulse ml-2"
                :title="$t('parameters.heldTooltip')"
              >
                {{ $t('parameters.heldBadge') }}
              </span>
            </div>
            <!-- Formula Result if defined -->
            <div v-if="getParamFormulaValue(param) !== null" class="flex items-center space-x-2 text-xs font-mono text-blue-400">
              <span class="font-bold">ƒ(x):</span>
              <span class="font-bold text-white bg-blue-500/10 px-1.5 py-0.5 rounded border border-blue-500/20">
                {{ getParamFormulaValue(param) }} {{ param.unit }}
              </span>
              <span v-if="param.formula" class="text-3xs text-slate-400 truncate max-w-[150px]">({{ param.formula }})</span>
            </div>
          </div>

          <!-- Raw & Decoded Diagnostics Sub-bar -->
          <div class="p-2.5 rounded-lg bg-[#0B0F19] border border-slate-800/80 grid grid-cols-2 gap-2 text-3xs font-mono text-slate-400">
            <div>
              <span class="text-slate-500 block uppercase">{{ $t('telemetry.rawHex') }}</span>
              <span class="text-slate-300 truncate block">
                {{ getParamRaw(param) }}
              </span>
            </div>
            <div>
              <span class="text-slate-500 block uppercase">{{ $t('common.timestamp') }}</span>
              <span class="text-slate-300 truncate block">
                {{ formatTimestamp(getParamTime(param)) }}
              </span>
            </div>
          </div>

          <!-- Card Action -->
          <div class="pt-1 flex items-center justify-between border-t border-slate-800/60 text-2xs">
            <span class="text-slate-500 font-mono">{{ $t('telemetry.source') }}: {{ (device.connection && device.connection.protocol) || 'MODBUS' }}</span>
            <button
              @click="inspectHistoryForParam(param.id)"
              class="text-blue-400 hover:text-blue-300 font-semibold flex items-center space-x-1"
            >
              <span>{{ $t('deviceDetail.viewHistory') }}</span>
              <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"></path>
              </svg>
            </button>
          </div>
        </div>

        <div v-if="!device.parameters || device.parameters.length === 0" class="col-span-full py-12 text-center text-xs text-slate-400">
          {{ $t('deviceDetail.noParamsTelemetry') }}
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
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">{{ $t('deviceDetail.paramLabel') }}</label>
            <select
              v-model="historyParamId"
              @change="onFilterChange"
              class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
            >
              <option value="">{{ $t('deviceDetail.allParametersOption') }}</option>
              <option v-for="p in device.parameters" :key="p.id" :value="p.id">
                {{ p.parameter_code }} — {{ p.parameter_name }} ({{ p.unit }})
              </option>
            </select>
          </div>

          <!-- Time Range Quick Buttons -->
          <div>
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">{{ $t('deviceDetail.timeRange') }}</label>
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
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">{{ $t('deviceDetail.dataQuality') }}</label>
            <select
              v-model="historyQuality"
              @change="fetchHistory"
              class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
            >
              <option value="">{{ $t('deviceDetail.allQualitiesOption') }}</option>
              <option value="GOOD">GOOD (Normal)</option>
              <option value="BAD">BAD (Error)</option>
              <option value="UNCERTAIN">UNCERTAIN</option>
            </select>
          </div>

          <!-- Resolution / Sample Limit -->
          <div>
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">{{ $t('deviceDetail.pointDensity') }}</label>
            <select
              v-model="historyPageSize"
              @change="onPageSizeChange"
              class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500 font-mono"
            >
              <option :value="20">{{ $t('deviceDetail.samplesOption', { count: 20 }) }}</option>
              <option :value="50">{{ $t('deviceDetail.samplesOption', { count: 50 }) }}</option>
              <option :value="100">{{ $t('deviceDetail.samplesOption', { count: 100 }) }}</option>
              <option :value="250">{{ $t('deviceDetail.samplesOption', { count: 250 }) }}</option>
            </select>
          </div>

          <!-- Downsampling Resolution (Phase 3.3) -->
          <div>
            <label class="block text-3xs uppercase font-semibold text-slate-400 mb-1">{{ $t('deviceDetail.downsamplingResolution') }}</label>
            <select
              v-model="historyDownsampleResolution"
              @change="fetchHistory"
              class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500 font-mono"
            >
              <option value="raw">{{ $t('deviceDetail.resolutionRaw') }}</option>
              <option value="auto">{{ $t('deviceDetail.resolutionAuto') }}</option>
              <option value="5m">5 {{ $t('aggregation.minutes') }}</option>
              <option value="30m">30 {{ $t('aggregation.minutes') }}</option>
              <option value="1h">1 {{ $t('aggregation.hours') }}</option>
              <option value="1d">1 {{ $t('aggregation.day') }}</option>
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
            <span>{{ $t('deviceDetail.refreshChart') }}</span>
          </button>
          <router-link
            :to="'/monitoring/analysis?device_id=' + device.id + '&mode=HISTORICAL'"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition flex items-center space-x-1.5"
          >
            <span>{{ $t('analysis.openInAnalysis') }}</span>
          </router-link>
        </div>
      </div>

      <!-- Historical Trend Chart Card -->
      <div class="saas-card p-4 sm:p-5 space-y-4 chart-container font-sans">
        <!-- Card Top Bar: Title, Parameter Quick Tabs, Controls -->
        <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3 border-b border-slate-800 pb-3">
          <div>
            <div class="flex items-center space-x-2">
              <span class="px-2 py-0.5 rounded text-3xs font-sans font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20 uppercase tracking-wider">
                {{ $t('history.trendCurve') }}
              </span>
              <span class="inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-3xs font-sans font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                <span>{{ $t('history.liveStream') }}</span>
              </span>
            </div>
            <h3 class="text-sm font-bold text-white mt-1.5 flex items-center space-x-2 font-sans">
              <span>{{ chartActiveName }}</span>
              <span class="text-blue-400 font-sans font-medium">({{ chartActiveCode }})</span>
            </h3>
            <p class="text-3xs text-slate-400 font-sans mt-0.5">
              {{ $t('deviceDetail.chartDynamicsSubtitle', { range: historyTimeRangeLabel }) }}
            </p>
          </div>

          <!-- Parameter Switcher Pills (If multiple parameters) -->
          <div v-if="device.parameters && device.parameters.length > 1" class="flex items-center space-x-1.5 flex-wrap">
            <span class="text-3xs text-slate-400 uppercase font-semibold mr-1 font-sans">{{ $t('deviceDetail.paramLabel') }}</span>
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
          <div class="flex items-center space-x-2 text-2xs font-sans flex-wrap gap-y-1.5">
            <!-- Mode Switch: Scaled / Formula vs Raw Value -->
            <div class="bg-slate-900 p-0.5 rounded-lg border border-slate-800 flex items-center shadow-sm">
              <button
                @click="chartValueMode = 'scaled'"
                class="px-2.5 py-1 rounded-md text-3xs font-semibold transition font-sans flex items-center space-x-1"
                :class="chartValueMode === 'scaled' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'"
                :title="$t('deviceDetail.modeScaledTitle')"
              >
                <span>{{ $t('deviceDetail.modeScaled') }}</span>
              </button>
              <button
                @click="chartValueMode = 'raw'"
                class="px-2.5 py-1 rounded-md text-3xs font-semibold transition font-sans flex items-center space-x-1"
                :class="chartValueMode === 'raw' ? 'bg-emerald-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'"
                :title="$t('deviceDetail.modeRawTitle')"
              >
                <span class="w-1.5 h-1.5 rounded-full" :class="chartValueMode === 'raw' ? 'bg-emerald-300' : 'bg-slate-500'"></span>
                <span>{{ $t('deviceDetail.modeRaw') }}</span>
              </button>
            </div>

            <!-- Curve Type Toggle -->
            <div class="bg-slate-900 p-0.5 rounded-lg border border-slate-800 flex items-center">
              <button
                @click="chartCurveType = 'smooth'"
                class="px-2.5 py-1 rounded-md text-3xs font-semibold transition font-sans"
                :class="chartCurveType === 'smooth' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'"
                :title="$t('deviceDetail.curveSmoothTitle')"
              >
                {{ $t('deviceDetail.curveSmooth') }}
              </button>
              <button
                @click="chartCurveType = 'linear'"
                class="px-2.5 py-1 rounded-md text-3xs font-semibold transition font-sans"
                :class="chartCurveType === 'linear' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'"
                :title="$t('deviceDetail.curveLinearTitle')"
              >
                {{ $t('deviceDetail.curveLinear') }}
              </button>
            </div>

            <!-- Toggle Dots -->
            <button
              @click="chartShowPoints = !chartShowPoints"
              class="px-2.5 py-1 rounded-lg border text-3xs font-medium transition flex items-center space-x-1.5 font-sans"
              :class="chartShowPoints
                ? 'bg-blue-600/15 text-blue-400 border-blue-500/30'
                : 'bg-slate-900 text-slate-500 border-slate-800 hover:text-slate-300'"
              :title="$t('deviceDetail.toggleDotsTitle')"
            >
              <span class="w-1.5 h-1.5 rounded-full" :class="chartShowPoints ? 'bg-blue-400' : 'bg-slate-600'"></span>
              <span>{{ $t('deviceDetail.toggleDots') }}</span>
            </button>

            <!-- Toggle Avg Line -->
            <button
              @click="chartShowAvgLine = !chartShowAvgLine"
              class="px-2.5 py-1 rounded-lg border text-3xs font-medium transition flex items-center space-x-1.5 font-sans"
              :class="chartShowAvgLine
                ? 'bg-amber-500/10 text-amber-400 border-amber-500/30'
                : 'bg-slate-900 text-slate-500 border-slate-800 hover:text-slate-300'"
              :title="$t('deviceDetail.toggleAvgLineTitle')"
            >
              <span>{{ $t('deviceDetail.toggleAvgLine') }}</span>
            </button>
          </div>
        </div>

        <!-- 5 KPI Cards Row -->
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3 font-sans">
          <!-- KPI 1: Nilai Terkini -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider font-sans">
              <span>{{ $t('deviceDetail.kpiLatest') }}</span>
              <span class="px-1.5 py-0.2 rounded text-4xs font-sans font-bold" :class="qualityBadgeClass(chartStats.latestQuality)">
                {{ chartStats.latestQuality }}
              </span>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-sans font-bold text-white tracking-tight tabular-nums">
                {{ chartStats.latest }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans flex items-center justify-between">
              <span>{{ $t('deviceDetail.atHour', { time: chartStats.latestTime }) }}</span>
              <span class="text-emerald-400 font-medium">{{ $t('deviceDetail.liveTag') }}</span>
            </div>
          </div>

          <!-- KPI 2: Minimum -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider font-sans">
              <span>{{ $t('deviceDetail.kpiMin') }}</span>
              <svg class="w-3.5 h-3.5 text-emerald-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M19 14l-7 7m0 0l-7-7m7 7V3"></path>
              </svg>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-sans font-bold text-emerald-400 tracking-tight tabular-nums">
                {{ chartStats.min }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans">
              {{ $t('deviceDetail.lowestAt', { time: chartStats.minTime }) }}
            </div>
          </div>

          <!-- KPI 3: Maximum -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider font-sans">
              <span>{{ $t('deviceDetail.kpiMax') }}</span>
              <svg class="w-3.5 h-3.5 text-rose-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M5 10l7-7m0 0l7 7m-7-7v18"></path>
              </svg>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-sans font-bold text-rose-400 tracking-tight tabular-nums">
                {{ chartStats.max }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans">
              {{ $t('deviceDetail.peakAt', { time: chartStats.maxTime }) }}
            </div>
          </div>

          <!-- KPI 4: Average -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider font-sans">
              <span>{{ $t('deviceDetail.kpiAvg') }}</span>
              <span class="text-3xs text-blue-400 font-semibold font-sans">AVG</span>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-sans font-bold text-blue-400 tracking-tight tabular-nums">
                {{ chartStats.avg }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans">
              {{ $t('deviceDetail.allSamplesMean') }}
            </div>
          </div>

          <!-- KPI 5: Samples & Delta -->
          <div class="bg-[#0B0F19] rounded-xl p-3 border border-slate-800 space-y-1 col-span-2 sm:col-span-1">
            <div class="flex items-center justify-between text-3xs text-slate-400 font-semibold uppercase tracking-wider font-sans">
              <span>{{ $t('deviceDetail.kpiDelta') }}</span>
              <span class="text-3xs text-slate-400 font-semibold font-sans">DELTA</span>
            </div>
            <div class="flex items-baseline space-x-1">
              <span class="text-2xl font-sans font-bold text-slate-200 tracking-tight tabular-nums">
                {{ chartStats.delta }}
              </span>
              <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
            </div>
            <div class="text-3xs text-slate-400 font-sans">
              <span v-html="$t('deviceDetail.samplesCount', { count: '<b class=\'text-slate-300 font-medium\'>' + chartStats.count + '</b>' })"></span>
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
            <span class="w-1.5 h-1.5 rounded-full" :class="chartValueMode === 'raw' ? 'bg-emerald-400' : 'bg-blue-500'"></span>
            <span>
              {{ $t('deviceDetail.yScale') }}
              <b class="text-slate-300 font-medium">{{ chartActiveUnit || $t('common.value') }}</b>
              <span class="ml-1 text-slate-500 font-mono text-4xs">[{{ chartValueMode === 'raw' ? $t('deviceDetail.modeRaw') : $t('deviceDetail.modeScaled') }}]</span>
            </span>
          </div>

          <!-- Empty State if no records -->
          <div v-if="chartPoints.length === 0" class="w-full h-full flex flex-col items-center justify-center space-y-2 text-center font-sans">
            <svg class="w-10 h-10 text-slate-700" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"></path>
            </svg>
            <p class="text-xs text-slate-400 font-medium font-sans">{{ $t('deviceDetail.noHistory') }}</p>
            <p class="text-3xs text-slate-500 font-sans">{{ $t('deviceDetail.noHistorySubtitle') }}</p>
          </div>

          <!-- SVG Chart -->
          <svg
            v-else
            class="w-full h-full overflow-visible chart-svg"
            :viewBox="`0 0 ${containerWidth} ${containerHeight}`"
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
                  :x1="60"
                  :y1="tick.y"
                  :x2="containerWidth - 30"
                  :y2="tick.y"
                  stroke="#1E293B"
                  stroke-dasharray="4 4"
                  stroke-width="1"
                />
                <text
                  :x="52"
                  :y="tick.y + 4"
                  text-anchor="end"
                  fill="#94A3B8"
                  font-size="11"
                  font-family="'Inter', 'Manrope', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"
                  font-weight="400"
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
                  :y1="containerHeight - 48"
                  :x2="tick.x"
                  :y2="containerHeight - 42"
                  stroke="#334155"
                  stroke-width="1.2"
                />
                <text
                  :x="tick.x"
                  :y="containerHeight - 26"
                  text-anchor="middle"
                  fill="#94A3B8"
                  font-size="11"
                  font-family="'Inter', 'Manrope', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"
                  font-weight="400"
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
                :x1="60"
                :y1="chartAvgY"
                :x2="containerWidth - 30"
                :y2="chartAvgY"
                stroke="#F59E0B"
                stroke-dasharray="5 4"
                stroke-width="1.2"
                opacity="0.8"
              />
              <rect
                :x="containerWidth - 100"
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
                :x="containerWidth - 65"
                :y="Math.max(6, chartAvgY - 9) + 12"
                text-anchor="middle"
                fill="#FBBF24"
                font-size="10"
                font-family="'Inter', 'Manrope', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"
                font-weight="500"
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
                :x="Math.max(60, Math.min(chartMaxPoint.x - 34, containerWidth - 98))"
                :y="Math.max(6, chartMaxPoint.y - 24)"
                width="68"
                height="19"
                rx="6"
                fill="#1E293B"
                stroke="#F43F5E"
                stroke-width="1"
                opacity="0.95"
              />
              <text
                :x="Math.max(94, Math.min(chartMaxPoint.x, containerWidth - 64))"
                :y="Math.max(6, chartMaxPoint.y - 24) + 13"
                text-anchor="middle"
                fill="#FDA4AF"
                font-size="10"
                font-family="'Inter', 'Manrope', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"
                font-weight="500"
              >
                MAX {{ chartValueMode === 'raw' && Number.isInteger(chartMaxPoint.value) ? chartMaxPoint.value : chartMaxPoint.value.toFixed(2) }}
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
                :x="Math.max(60, Math.min(chartMinPoint.x - 34, containerWidth - 98))"
                :y="Math.min(containerHeight - 70, chartMinPoint.y + 7)"
                width="68"
                height="19"
                rx="6"
                fill="#1E293B"
                stroke="#10B981"
                stroke-width="1"
                opacity="0.95"
              />
              <text
                :x="Math.max(94, Math.min(chartMinPoint.x, containerWidth - 64))"
                :y="Math.min(containerHeight - 70, chartMinPoint.y + 7) + 13"
                text-anchor="middle"
                fill="#6EE7B7"
                font-size="10"
                font-family="'Inter', 'Manrope', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif"
                font-weight="500"
              >
                MIN {{ chartValueMode === 'raw' && Number.isInteger(chartMinPoint.value) ? chartMinPoint.value : chartMinPoint.value.toFixed(2) }}
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
                :fill="chartDotColor(p.quality)"
                stroke="#0B0F19"
                stroke-width="1.5"
                class="hover:opacity-100"
              />
            </g>

            <!-- Active Mouse Crosshair Guidelines -->
            <g v-if="hoveredPoint">
              <line
                :x1="hoveredPoint.x"
                :y1="28"
                :x2="hoveredPoint.x"
                :y2="containerHeight - 48"
                stroke="#60A5FA"
                stroke-width="1.2"
                stroke-dasharray="3 3"
                opacity="0.8"
              />
              <line
                :x1="60"
                :y1="hoveredPoint.y"
                :x2="containerWidth - 30"
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
            class="absolute pointer-events-none z-30 transition-all duration-75 font-sans"
            :style="tooltipStyle"
          >
            <div class="bg-[#111827]/95 backdrop-blur-md border border-slate-700 shadow-2xl rounded-xl p-3 text-xs w-60 space-y-1.5 font-sans ring-1 ring-white/10">
              <div class="flex items-center justify-between border-b border-slate-800 pb-1.5">
                <span class="text-3xs font-sans font-semibold text-blue-400 uppercase tracking-wider">
                  {{ getParamCode(hoveredPoint.rawRecord.parameter_id) }}
                </span>
                <div class="flex flex-col items-end">
                  <span
                    class="px-1.5 py-0.5 rounded text-4xs font-sans font-semibold"
                    :class="qualityBadgeClass(hoveredPoint.quality)"
                  >
                    {{ hoveredPoint.quality }}
                  </span>
                  <span v-if="hoveredPoint.rawRecord && hoveredPoint.rawRecord.quality_reason && hoveredPoint.rawRecord.quality_reason !== 'NONE'" class="text-4xs font-mono text-amber-400 mt-0.5">
                    {{ hoveredPoint.rawRecord.quality_reason }}
                  </span>
                </div>
              </div>
              <div class="text-3xs text-slate-300 font-medium truncate font-sans">
                {{ getParamName(hoveredPoint.rawRecord.parameter_id) }}
              </div>
              <div class="flex items-baseline space-x-1 pt-0.5">
                <span class="text-2xl font-sans font-bold text-white tracking-tight tabular-nums">
                  {{ hoveredPoint.value !== undefined ? (chartValueMode === 'raw' && Number.isInteger(hoveredPoint.value) ? hoveredPoint.value : hoveredPoint.value.toFixed(2)) : '--' }}
                </span>
                <span class="text-xs font-sans text-slate-400 font-normal ml-0.5">{{ chartActiveUnit }}</span>
                <span
                  class="text-4xs font-sans font-bold px-1.5 py-0.5 rounded ml-1.5 uppercase"
                  :class="chartValueMode === 'raw' ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30' : 'bg-blue-500/20 text-blue-300 border border-blue-500/30'"
                >
                  {{ chartValueMode === 'raw' ? 'RAW' : 'SCALED' }}
                </span>
              </div>
              <div class="pt-1.5 border-t border-slate-800 flex items-center justify-between text-3xs text-slate-400 font-sans">
                <span>{{ formatTimestamp(hoveredPoint.timestamp) }}</span>
                <span class="text-slate-500 font-sans">#{{ hoveredPoint.index + 1 }}/{{ chartPoints.length }}</span>
              </div>
              <div class="text-4xs text-slate-400 font-sans flex items-center justify-between pt-1 border-t border-slate-800/60">
                <span v-if="chartValueMode === 'raw'">
                  {{ $t('deviceDetail.modeScaled') }}: <b class="text-blue-300">{{ hoveredPoint.scaledValue !== null ? hoveredPoint.scaledValue.toFixed(2) : '--' }} {{ chartActiveParam ? chartActiveParam.unit : '' }}</b>
                </span>
                <span v-else>
                  {{ $t('deviceDetail.modeRaw') }}: <b class="text-emerald-300">{{ hoveredPoint.rawValue !== null ? hoveredPoint.rawValue : '--' }}</b>
                </span>
                <span class="text-slate-500">{{ hoveredPoint.rawRecord ? (hoveredPoint.rawRecord.source || 'MODBUS') : 'MODBUS' }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Legend and Operator Guide -->
        <div class="flex flex-wrap items-center justify-between gap-3 text-3xs text-slate-400 pt-1 font-sans">
          <div class="flex items-center space-x-4 flex-wrap font-sans">
            <span class="flex items-center space-x-1.5">
              <span class="w-3 h-1 rounded bg-blue-500 inline-block"></span>
              <span>{{ $t('deviceDetail.legendTrend', { unit: chartActiveUnit }) }}</span>
            </span>
            <span v-if="chartShowAvgLine" class="flex items-center space-x-1.5">
              <span class="w-3 h-0.5 border-t border-dashed border-amber-400 inline-block"></span>
              <span>{{ $t('deviceDetail.legendAvg', { avg: chartStats.avg }) }}</span>
            </span>
            <span v-if="chartShowMinMax" class="flex items-center space-x-1.5">
              <span class="w-2 h-2 rounded-full bg-rose-500 inline-block"></span>
              <span>{{ $t('deviceDetail.legendMax') }}</span>
            </span>
            <span v-if="chartShowMinMax" class="flex items-center space-x-1.5">
              <span class="w-2 h-2 rounded-full bg-emerald-500 inline-block"></span>
              <span>{{ $t('deviceDetail.legendMin') }}</span>
            </span>
          </div>
          <div class="text-slate-500 italic flex items-center space-x-1 font-sans">
            <svg class="w-3 h-3 text-slate-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122"></path>
            </svg>
            <span>{{ $t('deviceDetail.hoverGuide') }}</span>
          </div>
        </div>
      </div>

      <!-- Historical Data Table -->
      <div class="saas-card overflow-hidden">
        <div class="p-3 border-b border-slate-800 flex items-center justify-between text-xs">
          <span class="font-bold text-white">{{ $t('deviceDetail.historicalRecordsTitle', { count: historyTotal }) }}</span>
          <span class="text-3xs text-slate-400 font-sans tabular-nums">{{ $t('deviceDetail.pageOf', { page: historyPage, total: Math.ceil(historyTotal / historyPageSize) || 1 }) }}</span>
        </div>

        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs font-sans">
            <thead class="bg-[#0B0F19] text-3xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-800">
              <tr>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.receivedTimeCol') }}</th>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.paramLabel') }}</th>
                <th class="py-2.5 px-4">{{ $t('common.value') }}</th>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.rawDecodedCol') }}</th>
                <th class="py-2.5 px-4">{{ $t('telemetry.quality') }}</th>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.protocolSourceCol') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 font-sans text-2xs tabular-nums">
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
                  <div v-if="rec.quality_reason && rec.quality_reason !== 'NONE'" class="text-4xs font-mono text-amber-400/90 mt-0.5" :title="rec.quality_reason">
                    {{ rec.quality_reason }}
                  </div>
                </td>
                <td class="py-2.5 px-4 text-slate-400">
                  {{ rec.source || 'MODBUS_TCP' }}
                </td>
              </tr>
              <tr v-if="historyRecords.length === 0">
                <td colspan="6" class="py-8 text-center text-xs text-slate-400 font-sans">
                  {{ historyLoading ? $t('deviceDetail.loadingRecords') : $t('deviceDetail.noRecordsFound') }}
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
            {{ $t('common.previous') }}
          </button>
          <span class="text-slate-400 text-2xs font-mono">
            {{ $t('deviceDetail.showingPagination', { start: (historyPage - 1) * historyPageSize + 1, end: Math.min(historyPage * historyPageSize, historyTotal), total: historyTotal }) }}
          </span>
          <button
            @click="nextHistoryPage"
            :disabled="historyPage * historyPageSize >= historyTotal || historyLoading"
            class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold"
          >
            {{ $t('common.next') }}
          </button>
        </div>
      </div>
    </div>

    <!-- TAB 8: RAW TELEMETRY (Phase 3.1) -->
    <div v-if="device && activeTab === 'raw'" class="space-y-4">
      <div class="saas-card p-4 flex flex-col md:flex-row md:items-center justify-between gap-3 bg-[#0F172A]/90">
        <div>
          <h3 class="text-xs font-bold text-white uppercase tracking-wider">{{ $t('deviceDetail.rawInspectionTitle') }}</h3>
          <p class="text-2xs text-slate-400 mt-0.5">
            {{ $t('deviceDetail.rawInspectionSubtitle') }}
          </p>
        </div>
        <div class="flex items-center space-x-2">
          <select
            v-model="rawParamId"
            @change="fetchRawTelemetry"
            class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-blue-500"
          >
            <option value="">{{ $t('deviceDetail.allParametersOption') }}</option>
            <option v-for="p in device.parameters" :key="p.id" :value="p.id">
              {{ p.parameter_code }}
            </option>
          </select>
          <button
            @click="fetchRawTelemetry"
            :disabled="rawLoading"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition"
          >
            {{ $t('common.refresh') }}
          </button>
          <router-link
            :to="'/monitoring/analysis?device_id=' + device.id + '&mode=RAW'"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 text-xs font-semibold transition flex items-center space-x-1.5"
          >
            <span>{{ $t('analysis.openInAnalysis') }}</span>
          </router-link>
        </div>
      </div>

      <!-- Raw Inspection Table -->
      <div class="saas-card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-xs font-mono">
            <thead class="bg-[#0B0F19] text-3xs font-bold text-slate-400 uppercase tracking-wider border-b border-slate-800">
              <tr>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.receivedTime') }}</th>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.paramLabel') }}</th>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.rawHexStream') }}</th>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.rawFloat') }}</th>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.scaledValue') }}</th>
                <th class="py-2.5 px-4">{{ $t('telemetry.quality') }}</th>
                <th class="py-2.5 px-4">{{ $t('deviceDetail.protocol') }}</th>
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
                  {{ rawLoading ? $t('deviceDetail.loadingRaw') : $t('deviceDetail.noRawCaptured') }}
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
            {{ $t('common.previous') }}
          </button>
          <span class="text-slate-400 text-2xs font-mono">
            {{ $t('deviceDetail.pageOfWithRecords', { page: rawPage, totalPages: Math.ceil(rawTotal / rawPageSize) || 1, total: rawTotal }) }}
          </span>
          <button
            @click="nextRawPage"
            :disabled="rawPage * rawPageSize >= rawTotal || rawLoading"
            class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold"
          >
            {{ $t('common.next') }}
          </button>
        </div>
      </div>
    </div>

    <!-- TAB 9: AGGREGATION & ROLLUPS (PHASE 3.3) -->
    <div v-if="device && activeTab === 'aggregation'" class="space-y-6 animate-fade-in">
      <!-- Header Banner & Control Bar -->
      <div class="saas-card p-5 bg-[#0F172A]/90 border border-slate-800 rounded-2xl flex flex-col lg:flex-row lg:items-center justify-between gap-4">
        <div class="flex items-center space-x-3.5">
          <div class="p-2.5 rounded-xl bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
            </svg>
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="text-base font-bold text-white tracking-wide">{{ $t('aggregation.title') }}</h2>
              <span class="text-[10px] uppercase font-bold font-mono px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                Phase 3.3 Reusable Engine
              </span>
            </div>
            <p class="text-xs text-slate-400 mt-0.5">{{ $t('aggregation.subtitle') }}</p>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex flex-wrap items-center gap-2.5">
          <!-- Customer View Mode Switch -->
          <button
            @click="customerViewMode = !customerViewMode"
            class="px-3 py-1.5 rounded-xl text-xs font-semibold transition flex items-center space-x-2 border"
            :class="customerViewMode 
              ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30 shadow-sm shadow-emerald-500/10' 
              : 'bg-slate-800 text-slate-300 border-slate-700 hover:bg-slate-700'"
          >
            <span class="w-2 h-2 rounded-full" :class="customerViewMode ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'"></span>
            <span>{{ customerViewMode ? $t('aggregation.customerMode') : $t('aggregation.engineeringMode') }}</span>
          </button>

          <!-- Auto Refresh Toggle (3s Real-Time) -->
          <button
            @click="toggleAutoRefreshAgg"
            class="px-3 py-1.5 rounded-xl text-xs font-semibold transition flex items-center space-x-1.5 border"
            :class="autoRefreshAgg 
              ? 'bg-emerald-500/15 text-emerald-300 border-emerald-500/40 shadow-sm shadow-emerald-500/10' 
              : 'bg-slate-800 text-slate-400 border-slate-700 hover:text-slate-200'"
          >
            <span class="w-2 h-2 rounded-full" :class="autoRefreshAgg ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'"></span>
            <span>{{ $t('aggregation.autoRefresh') }} (3s)</span>
          </button>

          <!-- Run Buckets -->
          <button
            @click="runAllAggBuckets"
            :title="$t('aggregation.runBucketsTooltip')"
            class="px-3 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-200 text-xs font-semibold transition flex items-center space-x-1.5 shadow-sm"
          >
            <svg class="w-3.5 h-3.5 text-indigo-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <span>{{ $t('aggregation.runBuckets') }}</span>
          </button>

          <!-- New Definition Button -->
          <button
            @click="openCreateAggDefModal"
            class="px-3.5 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold transition flex items-center space-x-1.5 shadow-sm shadow-indigo-500/20"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
            </svg>
            <span>{{ $t('aggregation.newDefinition') }}</span>
          </button>

          <router-link
            :to="'/monitoring/analysis?device_id=' + device.id + '&mode=AGGREGATED'"
            class="px-3 py-1.5 rounded-xl bg-indigo-950/40 hover:bg-indigo-900/40 border border-indigo-800/40 text-indigo-300 text-xs font-semibold transition flex items-center space-x-1.5"
          >
            <span>{{ $t('analysis.openInAnalysis') }}</span>
          </router-link>
        </div>
      </div>

      <!-- CUSTOMER DATA PREVIEW CARD (When customerViewMode is active) -->
      <div v-if="customerViewMode" class="saas-card p-5 bg-[#0B132B]/90 border border-emerald-500/30 rounded-2xl space-y-4">
        <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-emerald-500/20 pb-3">
          <div class="flex items-center space-x-2.5">
            <span class="p-1.5 rounded-lg bg-emerald-500/20 text-emerald-400">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
              </svg>
            </span>
            <div>
              <h3 class="text-sm font-bold text-emerald-300">{{ $t('aggregation.customerPreviewTitle') }}</h3>
              <p class="text-3xs text-slate-300 font-mono">{{ $t('aggregation.customerPreviewSubtitle') }}</p>
            </div>
          </div>
          <div class="flex items-center space-x-2">
            <input
              type="text"
              v-model="customerPreviewIdentifier"
              :placeholder="$t('aggregation.lookupPlaceholder')"
              class="px-3 py-1 bg-[#0F172A] border border-slate-700 rounded-lg text-xs text-white font-mono placeholder-slate-500 focus:outline-none focus:border-emerald-500 w-44"
            />
            <button
              @click="queryCustomerData(customerPreviewIdentifier)"
              :disabled="customerPreviewLoading"
              class="px-3 py-1 bg-emerald-600 hover:bg-emerald-500 text-white rounded-lg text-xs font-semibold transition"
            >
              {{ customerPreviewLoading ? '...' : $t('aggregation.lookupButton') }}
            </button>
          </div>
        </div>

        <div v-if="customerPreviewResult" class="space-y-3">
          <div v-if="customerPreviewResult.error" class="p-3 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-mono">
            {{ customerPreviewResult.error }}
          </div>
          <div v-else class="space-y-3">
            <div class="flex items-center space-x-4 text-xs font-mono">
              <span class="text-slate-400">Identifier: <strong class="text-cyan-400">{{ customerPreviewResult.identifier }}</strong></span>
              <span class="text-slate-400">Window: <strong class="text-slate-200">{{ customerPreviewResult.period_start }} → {{ customerPreviewResult.period_end }}</strong></span>
              <span class="text-emerald-400 bg-emerald-500/10 px-2 py-0.5 rounded text-3xs border border-emerald-500/20">
                ✓ {{ $t('aggregation.safeDeliveryVerified') }}
              </span>
            </div>

            <!-- Customer Data Cards Grid -->
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
              <div
                v-for="(item, idx) in customerPreviewResult.data"
                :key="idx"
                class="p-4 rounded-xl bg-[#0F172A] border border-slate-800 space-y-2"
              >
                <div class="flex items-center justify-between">
                  <span class="text-3xs uppercase font-bold text-slate-400 tracking-wider">{{ item.parameter }}</span>
                  <span class="px-2 py-0.5 rounded text-3xs font-bold" :class="item.quality === 'GOOD' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'">
                    {{ item.quality }}
                  </span>
                </div>
                <div class="text-2xl font-bold font-mono text-white">
                  {{ item.value !== null ? item.value : '--' }} <span class="text-sm font-normal text-slate-400 font-sans">{{ item.unit }}</span>
                </div>
                <div class="text-3xs text-slate-400 flex items-center justify-between border-t border-slate-800/80 pt-1.5 font-mono">
                  <span>Samples: {{ item.sample_count }}</span>
                  <span>Formula Processed</span>
                </div>
              </div>
            </div>

            <!-- JSON Payload Preview -->
            <div class="p-3 rounded-xl bg-black/60 border border-slate-800 text-3xs font-mono text-slate-300 max-h-40 overflow-y-auto">
              <pre>{{ JSON.stringify(customerPreviewResult, null, 2) }}</pre>
            </div>
          </div>
        </div>
      </div>

      <!-- SECTION 1: AGGREGATION DEFINITIONS -->
      <div class="saas-card bg-[#0F172A]/90 border border-slate-800 rounded-2xl overflow-hidden">
        <div class="p-4 border-b border-slate-800/80 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <h3 class="text-xs font-bold text-white uppercase tracking-wider">{{ $t('aggregation.definitionsTitle') }}</h3>
            <span class="text-2xs font-mono text-indigo-400 bg-indigo-500/10 px-2 py-0.5 rounded-full border border-indigo-500/20">
              {{ aggregationDefinitions.length }}
            </span>
          </div>
          <button @click="fetchAggregationDefinitions" class="text-slate-400 hover:text-white text-xs transition">
            {{ $t('aggregation.refresh') }}
          </button>
        </div>

        <div v-if="aggregationDefinitionsLoading" class="p-8 text-center text-slate-400 text-xs">
          Loading aggregation definitions...
        </div>
        <div v-else-if="aggregationDefinitions.length === 0" class="p-8 text-center text-slate-400 text-xs">
          {{ $t('aggregation.noDefinitions') }}
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-xs font-sans">
            <thead class="bg-[#0B0F19]/60 text-slate-400 uppercase text-3xs font-semibold tracking-wider border-b border-slate-800">
              <tr>
                <th class="px-4 py-3">{{ $t('aggregation.name') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.parameter') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.sourceType') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.function') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.interval') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.timezone') }}</th>
                <th class="px-4 py-3">{{ $t('deviceDetail.status') }}</th>
                <th class="px-4 py-3 text-right">{{ $t('aggregation.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 text-slate-200">
              <tr v-for="def in aggregationDefinitions" :key="def.id" class="hover:bg-slate-800/40 transition">
                <td class="px-4 py-3 font-medium text-white">
                  <div>{{ def.name }}</div>
                  <div class="text-3xs font-mono text-slate-400">{{ def.code }}</div>
                </td>
                <td class="px-4 py-3">
                  <div class="font-medium text-slate-200">{{ def.parameter ? def.parameter.parameter_name : getParamName(def.parameter_id) }}</div>
                  <div class="text-3xs font-mono text-slate-400">{{ def.parameter ? def.parameter.parameter_code : getParamCode(def.parameter_id) }}</div>
                </td>
                <td class="px-4 py-3">
                  <span
                    class="px-2 py-0.5 rounded text-3xs font-bold font-mono tracking-wide"
                    :class="def.source_type === 'CUSTOMER_PROCESSED' 
                      ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' 
                      : 'bg-purple-500/10 text-purple-400 border border-purple-500/20'"
                  >
                    {{ def.source_type }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
                    {{ def.function }}
                  </span>
                </td>
                <td class="px-4 py-3 font-mono text-slate-300">
                  {{ def.interval_seconds >= 60 ? (def.interval_seconds / 60) + 'm' : def.interval_seconds + 's' }}
                  <span class="text-3xs text-slate-500 font-sans">({{ def.interval_seconds }}s)</span>
                </td>
                <td class="px-4 py-3 font-mono text-3xs text-slate-400">
                  {{ def.timezone || 'Asia/Jakarta' }}
                </td>
                <td class="px-4 py-3">
                  <button
                    @click="toggleAggDefStatus(def)"
                    class="px-2 py-0.5 rounded-full text-3xs font-bold transition flex items-center space-x-1"
                    :class="def.enabled ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-slate-800 text-slate-400 border border-slate-700'"
                  >
                    <span class="w-1.5 h-1.5 rounded-full" :class="def.enabled ? 'bg-emerald-400' : 'bg-slate-500'"></span>
                    <span>{{ def.enabled ? $t('aggregation.statusActive') : $t('aggregation.statusDisabled') }}</span>
                  </button>
                </td>
                <td class="px-4 py-3 text-right">
                  <div class="flex items-center justify-end space-x-2">
                    <button
                      @click="runAggBucket(def.id)"
                      :title="$t('aggregation.recalculate')"
                      class="p-1 rounded-lg bg-indigo-500/10 text-indigo-400 hover:bg-indigo-500/20 border border-indigo-500/20 transition"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path>
                      </svg>
                    </button>
                    <button
                      @click="openEditAggDefModal(def)"
                      class="p-1 rounded-lg bg-slate-800 text-slate-300 hover:text-white hover:bg-slate-700 transition"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
                      </svg>
                    </button>
                    <button
                      @click="deleteAggregationDefinition(def)"
                      class="p-1 rounded-lg bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20 transition"
                    >
                      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                      </svg>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- SECTION 2: AGGREGATION RESULTS BROWSER -->
      <div class="saas-card bg-[#0F172A]/90 border border-slate-800 rounded-2xl overflow-hidden space-y-4">
        <!-- Filter Bar -->
        <div class="p-4 border-b border-slate-800/80 flex flex-col lg:flex-row lg:items-center justify-between gap-3">
          <div class="flex flex-wrap items-center gap-2.5">
            <!-- Filter by Source -->
            <div>
              <select
                v-model="aggregationFilterSource"
                @change="fetchAggregationResults"
                class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-indigo-500 font-mono"
              >
                <option value="">{{ $t('deviceDetail.allDomains') || 'All Domains' }}</option>
                <option value="CUSTOMER_PROCESSED">CUSTOMER_PROCESSED</option>
                <option value="INTERNAL_RAW">INTERNAL_RAW</option>
              </select>
            </div>

            <!-- Filter by Parameter -->
            <div>
              <select
                v-model="aggregationFilterParamId"
                @change="fetchAggregationResults"
                class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-sans focus:outline-none focus:border-indigo-500"
              >
                <option value="">{{ $t('deviceDetail.allParametersOption') }}</option>
                <option v-for="p in device.parameters" :key="p.id" :value="p.id">
                  {{ p.parameter_code }} ({{ p.parameter_name }})
                </option>
              </select>
            </div>

            <!-- Search by Identifier -->
            <div>
              <input
                type="text"
                v-model="aggregationIdentifierSearch"
                @keyup.enter="fetchAggregationResults"
                :placeholder="$t('aggregation.lookupPlaceholder')"
                class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-mono placeholder-slate-500 focus:outline-none focus:border-indigo-500 w-48"
              />
            </div>

            <button
              @click="fetchAggregationResults"
              class="px-3 py-1.5 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold transition"
            >
              {{ $t('common.filter') || 'Filter' }}
            </button>
          </div>

          <div class="flex items-center space-x-2 text-xs font-mono text-slate-400">
            <span>{{ $t('aggregation.resultsTitle') }}: <strong class="text-white">{{ aggregationTotal }}</strong></span>
          </div>
        </div>

        <!-- Table -->
        <div v-if="aggregationResultsLoading" class="p-8 text-center text-slate-400 text-xs">
          Loading aggregation rollups...
        </div>
        <div v-else-if="aggregationResults.length === 0" class="p-8 text-center text-slate-400 text-xs">
          {{ $t('aggregation.noResults') }}
        </div>
        <div v-else class="overflow-x-auto">
          <table class="w-full text-left text-xs font-sans">
            <thead class="bg-[#0B0F19]/60 text-slate-400 uppercase text-3xs font-semibold tracking-wider border-b border-slate-800">
              <tr>
                <th class="px-4 py-3">{{ $t('aggregation.identifier') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.period') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.parameter') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.sourceType') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.function') }}</th>
                <th class="px-4 py-3">{{ $t('aggregation.aggregatedValue') }}</th>
                <th class="px-4 py-3">Min / Max / Avg</th>
                <th class="px-4 py-3">{{ $t('aggregation.samples') }} & {{ $t('aggregation.quality') }}</th>
                <th class="px-4 py-3 text-right">{{ $t('aggregation.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-800/60 text-slate-200">
              <tr v-for="res in aggregationResults" :key="res.id" class="hover:bg-slate-800/40 transition">
                <!-- Identifier & Status -->
                <td class="px-4 py-3 font-mono font-bold tracking-wider">
                  <div class="text-cyan-400 text-xs">{{ res.identifier }}</div>
                  <div class="mt-1">
                    <span
                      v-if="isBucketActive(res)"
                      class="inline-flex items-center space-x-1 px-1.5 py-0.5 rounded text-[9px] font-bold font-mono bg-emerald-500/20 text-emerald-300 border border-emerald-500/40"
                    >
                      <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span>
                      <span>{{ $t('aggregation.liveInProgress') }}</span>
                    </span>
                    <span
                      v-else
                      class="inline-block px-1.5 py-0.5 rounded text-[9px] font-mono bg-slate-800 text-slate-400 border border-slate-700/80"
                    >
                      {{ $t('aggregation.finalized') }}
                    </span>
                  </div>
                </td>
                <!-- Period -->
                <td class="px-4 py-3 font-mono text-3xs text-slate-300">
                  <div class="font-bold text-slate-200">{{ formatPeriodWindow(res.period_start, res.period_end) }}</div>
                  <div class="text-slate-500">{{ new Date(res.period_start).toLocaleDateString() }}</div>
                </td>
                <!-- Parameter -->
                <td class="px-4 py-3">
                  <div class="font-medium text-slate-200">{{ res.parameter ? res.parameter.parameter_name : getParamName(res.parameter_id) }}</div>
                  <div class="text-3xs font-mono text-slate-400">{{ res.parameter ? res.parameter.parameter_code : getParamCode(res.parameter_id) }}</div>
                </td>
                <!-- Domain Source -->
                <td class="px-4 py-3">
                  <span
                    class="px-2 py-0.5 rounded text-3xs font-bold font-mono tracking-wide"
                    :class="res.source_type === 'CUSTOMER_PROCESSED' 
                      ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' 
                      : 'bg-purple-500/10 text-purple-400 border border-purple-500/20'"
                  >
                    {{ res.source_type }}
                  </span>
                </td>
                <!-- Function -->
                <td class="px-4 py-3">
                  <span class="px-2 py-0.5 rounded text-3xs font-mono font-bold bg-blue-500/10 text-blue-400 border border-blue-500/20">
                    {{ res.aggregation_definition ? res.aggregation_definition.function : 'AVG' }}
                  </span>
                </td>
                <!-- Value -->
                <td class="px-4 py-3 font-mono font-bold text-sm text-white">
                  {{ res.value !== null ? Number(res.value).toFixed(2) : '--' }}
                  <span class="text-3xs text-slate-400 font-sans font-normal">{{ getParamUnit(res.parameter_id) }}</span>
                </td>
                <!-- Min / Max / Avg -->
                <td class="px-4 py-3 font-mono text-3xs text-slate-300">
                  <div>Min: <strong class="text-slate-100">{{ res.min_value !== null ? Number(res.min_value).toFixed(2) : '--' }}</strong></div>
                  <div>Max: <strong class="text-slate-100">{{ res.max_value !== null ? Number(res.max_value).toFixed(2) : '--' }}</strong></div>
                  <div>Avg: <strong class="text-slate-100">{{ res.avg_value !== null ? Number(res.avg_value).toFixed(2) : '--' }}</strong></div>
                </td>
                <!-- Samples & Quality -->
                <td class="px-4 py-3">
                  <div class="flex items-center space-x-2">
                    <span
                      class="px-2 py-0.5 rounded text-3xs font-bold tracking-wide"
                      :class="res.quality === 'GOOD' 
                        ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' 
                        : (res.quality === 'UNCERTAIN' ? 'bg-amber-500/10 text-amber-400 border border-amber-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20')"
                      :title="'Overall Quality: ' + res.quality"
                    >
                      {{ res.quality }}
                    </span>
                    <div class="relative group inline-block">
                      <button
                        @click="openBucketSamplesModal(res)"
                        class="text-3xs font-mono text-cyan-400 hover:text-cyan-300 hover:underline flex items-center space-x-1 cursor-pointer"
                        :title="$t('aggregation.sampleBreakdown') || 'Sample Quality Breakdown'"
                      >
                        <span>N={{ res.sample_count }} (✓{{ res.good_count }} / ?{{ res.uncertain_count }} / ✗{{ res.bad_count }})</span>
                      </button>

                      <!-- Interactive Hover Tooltip Popover -->
                      <div class="absolute left-0 bottom-full mb-2 hidden group-hover:block z-50 w-72 p-3 bg-[#0B0F19] border border-slate-700/80 rounded-xl shadow-2xl text-left pointer-events-none font-sans">
                        <div class="text-3xs font-bold uppercase text-slate-300 border-b border-slate-800 pb-1.5 mb-2 flex items-center justify-between">
                          <span class="flex items-center space-x-1">
                            <span>📊</span>
                            <span>{{ $t('aggregation.sampleBreakdown') || 'Sample Quality Breakdown' }}</span>
                          </span>
                          <span class="text-cyan-400 font-mono font-bold">Total N={{ res.sample_count }}</span>
                        </div>
                        <div class="space-y-1.5 text-3xs font-sans leading-tight">
                          <div class="flex items-start space-x-2">
                            <span class="font-bold font-mono text-emerald-400 shrink-0">✓ {{ res.good_count }}</span>
                            <span class="text-slate-300">{{ $t('aggregation.tooltipGood') }}</span>
                          </div>
                          <div class="flex items-start space-x-2">
                            <span class="font-bold font-mono text-amber-400 shrink-0">? {{ res.uncertain_count }}</span>
                            <span class="text-slate-300">{{ $t('aggregation.tooltipUncertain') }}</span>
                          </div>
                          <div class="flex items-start space-x-2">
                            <span class="font-bold font-mono text-rose-400 shrink-0">✗ {{ res.bad_count }}</span>
                            <span class="text-slate-300">{{ $t('aggregation.tooltipBad') }}</span>
                          </div>
                        </div>
                        <div class="mt-2 pt-1 border-t border-slate-800/80 text-[10px] text-slate-400 italic">
                          Click to view individual telemetry values
                        </div>
                      </div>
                    </div>
                  </div>
                </td>
                <!-- Actions -->
                <td class="px-4 py-3 text-right">
                  <div class="flex items-center justify-end space-x-1.5">
                    <button
                      @click="openBucketSamplesModal(res)"
                      class="p-1 rounded-lg bg-cyan-500/10 text-cyan-400 hover:bg-cyan-500/20 border border-cyan-500/30 transition text-3xs font-mono px-2 py-1 flex items-center space-x-1"
                      title="View Raw Telemetry Values & Samples"
                    >
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                      </svg>
                      <span>{{ $t('aggregation.viewValues') || 'Values' }}</span>
                    </button>
                    <button
                      @click="queryCustomerData(res.identifier)"
                      class="p-1 rounded-lg bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20 border border-emerald-500/20 transition text-3xs font-mono px-2 py-1"
                      title="Query Customer API"
                    >
                      API
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Pagination -->
        <div class="p-4 border-t border-slate-800/80 flex items-center justify-between">
          <button
            @click="prevAggPage"
            :disabled="aggregationPage <= 1 || aggregationResultsLoading"
            class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold text-xs"
          >
            {{ $t('common.previous') }}
          </button>
          <span class="text-slate-400 text-2xs font-mono">
            {{ $t('deviceDetail.pageOfWithRecords', { page: aggregationPage, totalPages: Math.ceil(aggregationTotal / aggregationPageSize) || 1, total: aggregationTotal }) }}
          </span>
          <button
            @click="nextAggPage"
            :disabled="aggregationPage * aggregationPageSize >= aggregationTotal || aggregationResultsLoading"
            class="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 transition font-semibold text-xs"
          >
            {{ $t('common.next') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Modals -->
    <!-- Aggregation Definition Modal (Phase 3.3) -->
    <div v-if="isAggDefModalOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
      <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-lg shadow-2xl overflow-hidden flex flex-col">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
          <div class="flex items-center space-x-2.5">
            <div class="p-1.5 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6V4m0 2a2 2 0 100 4m0-4a2 2 0 110 4m-6 8a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4m6 6v10m6-2a2 2 0 100-4m0 4a2 2 0 110-4m0 4v2m0-6V4"></path>
              </svg>
            </div>
            <div>
              <h2 class="text-sm font-bold text-white tracking-wide">
                {{ isEditingAggDef ? $t('aggregation.modalEditTitle') : $t('aggregation.modalCreateTitle') }}
              </h2>
              <p class="text-3xs text-slate-400 font-mono">{{ $t('aggregation.modalDesc') }}</p>
            </div>
          </div>
          <button @click="isAggDefModalOpen = false" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
          </button>
        </div>

        <!-- Body Form -->
        <div class="p-6 space-y-4 text-xs font-sans">
          <!-- Name & Code -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.name') }} *</label>
              <input
                type="text"
                v-model="aggDefForm.name"
                placeholder="e.g. Temperature 30m Average"
                class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-white focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.code') }} *</label>
              <input
                type="text"
                v-model="aggDefForm.code"
                placeholder="e.g. temp_avg_30m"
                class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-indigo-500"
              />
            </div>
          </div>

          <!-- Parameter Selection -->
          <div>
            <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.parameter') }} *</label>
            <select
              v-model="aggDefForm.parameter_id"
              class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
            >
              <option value="" disabled>{{ $t('aggregation.selectParameter') }}</option>
              <option v-if="!isEditingAggDef && device && device.parameters && device.parameters.length > 0" value="ALL" class="text-cyan-400 font-bold bg-[#0F172A]">
                {{ $t('aggregation.allParameters') }}
              </option>
              <option v-for="p in device.parameters" :key="p.id" :value="p.id">
                {{ p.parameter_code }} — {{ p.parameter_name }} ({{ p.unit || 'no unit' }})
              </option>
            </select>
            <p v-if="aggDefForm.parameter_id === 'ALL'" class="text-3xs text-cyan-400 mt-1 font-sans flex items-center space-x-1">
              <span>⚡</span>
              <span>{{ $t('aggregation.allParametersNotice', { count: device.parameters.length }) }}</span>
            </p>
          </div>

          <!-- Source Domain -->
          <div>
            <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.sourceType') }} *</label>
            <select
              v-model="aggDefForm.source_type"
              class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-mono focus:outline-none focus:border-indigo-500"
            >
              <option value="CUSTOMER_PROCESSED">{{ $t('aggregation.sourceCustomerProcessed') }}</option>
              <option value="INTERNAL_RAW">{{ $t('aggregation.sourceInternalRaw') }}</option>
            </select>
            <p class="text-3xs text-slate-400 mt-1">
              {{ aggDefForm.source_type === 'CUSTOMER_PROCESSED' ? $t('aggregation.customerProcessedDesc') : $t('aggregation.internalRawDesc') }}
            </p>
          </div>

          <!-- Function & Interval Preset -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.function') }} *</label>
              <select
                v-model="aggDefForm.function"
                class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-mono focus:outline-none focus:border-indigo-500"
              >
                <option value="AVG">AVG (Average)</option>
                <option value="MIN">MIN (Minimum)</option>
                <option value="MAX">MAX (Maximum)</option>
                <option value="SUM">SUM (Total Sum)</option>
                <option value="COUNT">COUNT (Sample Count)</option>
                <option value="FIRST">FIRST (First Valid)</option>
                <option value="LAST">LAST (Last Valid)</option>
              </select>
            </div>
            <div>
              <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.interval') }} *</label>
              <select
                v-model="aggDefForm.interval_preset"
                @change="onIntervalPresetChange"
                class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 focus:outline-none focus:border-indigo-500"
              >
                <option value="2m">{{ $t('aggregation.interval2m') }}</option>
                <option value="5m">{{ $t('aggregation.interval5m') }}</option>
                <option value="10m">{{ $t('aggregation.interval10m') }}</option>
                <option value="15m">{{ $t('aggregation.interval15m') }}</option>
                <option value="30m">{{ $t('aggregation.interval30m') }}</option>
                <option value="60m">{{ $t('aggregation.interval60m') }}</option>
                <option value="custom">{{ $t('aggregation.intervalCustom') }}</option>
              </select>
            </div>
          </div>

          <!-- Custom Interval Seconds (if custom) -->
          <div v-if="aggDefForm.interval_preset === 'custom'">
            <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.customInterval') }}</label>
            <input
              type="number"
              v-model="aggDefForm.interval_seconds"
              min="10"
              step="10"
              class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-indigo-500"
            />
          </div>

          <!-- Timezone & Quality Policy -->
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.timezone') }}</label>
              <input
                type="text"
                v-model="aggDefForm.timezone"
                placeholder="Asia/Jakarta"
                class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-white font-mono focus:outline-none focus:border-indigo-500"
              />
            </div>
            <div>
              <label class="block text-3xs font-semibold text-slate-400 uppercase mb-1">{{ $t('aggregation.qualityPolicy') }}</label>
              <select
                v-model="aggDefForm.quality_policy"
                class="w-full px-3 py-1.5 bg-[#0B0F19] border border-slate-700 rounded-xl text-xs text-slate-200 font-mono focus:outline-none focus:border-indigo-500"
              >
                <option value="STRICT">STRICT (Requires Good Samples)</option>
                <option value="BEST_EFFORT">BEST_EFFORT (Include Uncertain)</option>
              </select>
            </div>
          </div>

          <!-- Enabled Toggle -->
          <div class="flex items-center space-x-2 pt-1">
            <input
              type="checkbox"
              id="aggDefEnabled"
              v-model="aggDefForm.enabled"
              class="rounded bg-[#0B0F19] border-slate-700 text-indigo-600 focus:ring-0 w-4 h-4 cursor-pointer"
            />
            <label for="aggDefEnabled" class="text-xs text-slate-300 cursor-pointer select-none">
              {{ $t('aggregation.statusActive') }}
            </label>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-4 border-t border-slate-800 flex items-center justify-end space-x-3 bg-[#0B0F19]/50">
          <button
            @click="isAggDefModalOpen = false"
            class="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition"
          >
            {{ $t('common.cancel') }}
          </button>
          <button
            @click="saveAggregationDefinition"
            class="px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold transition shadow-sm shadow-indigo-500/20"
          >
            {{ $t('common.save') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Bucket Samples & Data Values Inspection Modal -->
    <div v-if="isBucketSamplesModalOpen && selectedBucketResult" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
      <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-4xl shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
        <!-- Header -->
        <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
          <div class="flex items-center space-x-3">
            <div class="p-2 rounded-xl bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
              </svg>
            </div>
            <div>
              <div class="flex items-center space-x-2">
                <h2 class="text-sm font-bold text-white tracking-wide">
                  {{ $t('aggregation.bucketSamplesTitle') || 'Bucket Telemetry Samples & Data Values' }}
                </h2>
                <span
                  v-if="isBucketActive(selectedBucketResult)"
                  class="inline-flex items-center space-x-1 px-1.5 py-0.5 rounded text-3xs font-bold font-mono bg-emerald-500/20 text-emerald-300 border border-emerald-500/40"
                >
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span>
                  <span>{{ $t('aggregation.liveInProgress') }}</span>
                </span>
                <span
                  v-else
                  class="inline-block px-1.5 py-0.5 rounded text-3xs font-mono bg-slate-800 text-slate-400 border border-slate-700/80"
                >
                  {{ $t('aggregation.finalized') }}
                </span>
              </div>
              <p class="text-3xs text-slate-400 font-mono mt-0.5">
                Window: <strong class="text-slate-200">{{ formatPeriodWindow(selectedBucketResult.period_start, selectedBucketResult.period_end) }}</strong>
                • Target: <strong class="text-cyan-400">{{ selectedBucketResult.identifier }}</strong>
                • Parameter: <strong class="text-indigo-300">{{ selectedBucketResult.parameter ? selectedBucketResult.parameter.parameter_name : getParamName(selectedBucketResult.parameter_id) }}</strong>
              </p>
            </div>
          </div>
          <button @click="closeBucketSamplesModal" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
            </svg>
          </button>
        </div>

        <!-- KPI Cards -->
        <div class="px-6 py-3 bg-[#0B0F19]/40 border-b border-slate-800/60 grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
          <div class="p-2.5 rounded-xl bg-[#0F172A] border border-slate-800">
            <span class="text-3xs text-slate-400 uppercase font-mono block">Sample Count</span>
            <span class="text-base font-bold font-mono text-cyan-400">
              N={{ bucketSamplesData ? bucketSamplesData.sample_count : (selectedBucketResult.sample_count || 0) }}
            </span>
          </div>
          <div class="p-2.5 rounded-xl bg-[#0F172A] border border-slate-800">
            <span class="text-3xs text-slate-400 uppercase font-mono block">Calculated Avg</span>
            <span class="text-base font-bold font-mono text-white">
              {{ (bucketSamplesData && bucketSamplesData.avg_value !== null) ? Number(bucketSamplesData.avg_value).toFixed(2) : (selectedBucketResult.value !== null ? Number(selectedBucketResult.value).toFixed(2) : '--') }}
              <span class="text-3xs font-normal text-slate-400">{{ bucketSamplesData ? bucketSamplesData.unit : '' }}</span>
            </span>
          </div>
          <div class="p-2.5 rounded-xl bg-[#0F172A] border border-slate-800">
            <span class="text-3xs text-slate-400 uppercase font-mono block">Min / Max</span>
            <span class="text-xs font-mono text-slate-200">
              {{ selectedBucketResult.min_value !== null ? Number(selectedBucketResult.min_value).toFixed(2) : '--' }} / {{ selectedBucketResult.max_value !== null ? Number(selectedBucketResult.max_value).toFixed(2) : '--' }}
            </span>
          </div>
          <div class="p-2.5 rounded-xl bg-[#0F172A] border border-slate-800 flex items-center justify-between">
            <div>
              <span class="text-3xs text-slate-400 uppercase font-mono block">Quality</span>
              <div class="flex items-center space-x-1.5 mt-0.5">
                <span class="text-xs font-bold font-mono text-emerald-400">
                  {{ selectedBucketResult.quality || 'GOOD' }}
                </span>
                <span class="text-3xs font-mono text-slate-400">
                  (✓{{ selectedBucketResult.good_count || 0 }} / ?{{ selectedBucketResult.uncertain_count || 0 }} / ✗{{ selectedBucketResult.bad_count || 0 }})
                </span>
              </div>
            </div>
            <button
              @click="fetchBucketSamples"
              :disabled="bucketSamplesLoading"
              class="px-2 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 text-3xs font-semibold border border-slate-700 transition"
            >
              Refresh
            </button>
          </div>
        </div>

        <!-- Samples Table -->
        <div class="flex-1 overflow-y-auto p-6">
          <div v-if="bucketSamplesLoading && (!bucketSamplesData || !bucketSamplesData.samples)" class="p-12 text-center text-slate-400 text-xs">
            <span class="animate-pulse">Loading telemetry values...</span>
          </div>
          <div v-else-if="!bucketSamplesData || !bucketSamplesData.samples || bucketSamplesData.samples.length === 0" class="p-12 text-center text-slate-400 text-xs">
            No telemetry records found for this time window.
          </div>
          <div v-else class="overflow-x-auto">
            <table class="w-full text-left text-xs font-sans">
              <thead class="bg-[#0B0F19]/80 text-slate-400 uppercase text-3xs font-semibold tracking-wider border-b border-slate-800 sticky top-0">
                <tr>
                  <th class="px-3 py-2">#</th>
                  <th class="px-3 py-2">Received Time</th>
                  <th class="px-3 py-2">Raw Hex Stream</th>
                  <th class="px-3 py-2">Raw Float</th>
                  <th class="px-3 py-2">Scaled Value</th>
                  <th class="px-3 py-2 text-right">Quality</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-800/60 text-slate-200 font-mono text-xs">
                <tr v-for="(sample, sIdx) in bucketSamplesData.samples" :key="sample.id || sIdx" class="hover:bg-slate-800/30 transition">
                  <td class="px-3 py-1.5 text-slate-500 text-3xs">{{ sIdx + 1 }}</td>
                  <td class="px-3 py-1.5 text-slate-300">
                    {{ formatTimestampExact(sample.received_at) }}
                  </td>
                  <td class="px-3 py-1.5 text-emerald-400 text-3xs">
                    {{ sample.raw_hex || '--' }}
                  </td>
                  <td class="px-3 py-1.5 text-slate-300">
                    {{ sample.raw_value !== null ? Number(sample.raw_value).toFixed(4) : '--' }}
                  </td>
                  <td class="px-3 py-1.5 font-bold text-white">
                    {{ sample.processed_value !== null ? Number(sample.processed_value).toFixed(2) : '--' }}
                    <span class="text-3xs font-normal text-slate-400 font-sans">{{ bucketSamplesData.unit }}</span>
                  </td>
                  <td class="px-3 py-1.5 text-right">
                    <span
                      class="px-2 py-0.5 rounded text-3xs font-bold"
                      :class="sample.quality === 'GOOD' ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'"
                    >
                      {{ sample.quality || 'GOOD' }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Footer -->
        <div class="px-6 py-3 border-t border-slate-800/80 bg-[#0B0F19]/80 flex items-center justify-between text-xs font-mono text-slate-400">
          <span>Total records: <strong class="text-white">{{ bucketSamplesData ? bucketSamplesData.samples.length : 0 }}</strong></span>
          <button @click="closeBucketSamplesModal" class="px-4 py-1.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-white font-semibold transition">
            Close
          </button>
        </div>
      </div>
    </div>
    <DeviceModal
      :is-open="isEditModalOpen"
      :device-to-edit="device"
      @close="isEditModalOpen = false"
      @saved="handleDeviceSaved"
    />

    <ParameterModal
      v-if="device"
      :is-open="isParamModalOpen"
      :device-id="device.id"
      :param-to-edit="paramToEdit"
      @close="isParamModalOpen = false"
      @saved="handleParamSaved"
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
              <h2 class="text-sm font-bold text-white tracking-wide">{{ $t('deviceDetail.testReadTitle') }}</h2>
              <p class="text-3xs text-slate-400 font-mono">{{ $t('deviceDetail.testReadDeviceParam', { device: device.device_code, param: testParam.parameter_code }) }}</p>
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
              <span class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.paramLabel') }}</span>
              <div class="text-xs text-white font-medium mt-0.5 truncate">{{ testParam.parameter_name }}</div>
            </div>
            <div>
              <span class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.dataType') }}</span>
              <div class="text-xs font-mono font-bold text-blue-400 mt-0.5">{{ testParam.data_type }}</div>
            </div>
            <div>
              <span class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.registerAndType') }}</span>
              <div class="text-xs font-mono text-slate-200 mt-0.5">#{{ testParam.register_address }} ({{ testParam.register_type }})</div>
            </div>
            <div>
              <span class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.scaleAndOffset') }}</span>
              <div class="text-xs font-mono text-slate-200 mt-0.5">×{{ testParam.scale || 1 }} + {{ testParam.offset || 0 }} ({{ testParam.unit || '--' }})</div>
            </div>
          </div>

          <!-- Loading state -->
          <div v-if="testReadLoading" class="py-8 flex flex-col items-center justify-center space-y-3">
            <div class="w-7 h-7 border-2 border-emerald-500/20 border-t-emerald-500 rounded-full animate-spin"></div>
            <span class="text-xs text-slate-400 font-mono">{{ $t('deviceDetail.executingTransaction') }}</span>
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
                  <span>{{ testReadResult.success ? $t('deviceDetail.transactionSuccess') : $t('deviceDetail.transactionFailed') }}</span>
                </span>
                <span class="text-3xs font-mono text-slate-400">{{ testReadResult.response_time_ms || 0 }} ms</span>
              </div>

              <!-- Error display if failed -->
              <div v-if="!testReadResult.success" class="text-xs text-rose-300 font-mono bg-rose-950/40 p-2.5 rounded-lg border border-rose-500/20">
                {{ testReadResult.error || testReadResult.error_message || $t('deviceDetail.modbusFailed') }}
              </div>

              <!-- Values display if success -->
              <div v-else class="space-y-3 pt-1">
                <!-- Prominent Formula Result Card if formula is configured -->
                <div v-if="getTestReadFormulaVal() !== null" class="p-3 bg-blue-950/30 rounded-xl border border-blue-500/30 space-y-1">
                  <div class="flex items-center justify-between">
                    <span class="text-3xs text-blue-300 font-semibold uppercase tracking-wider flex items-center space-x-1.5">
                      <span class="w-2 h-2 rounded-full bg-blue-400"></span>
                      <span>{{ $t('parameters.formulaValue') }} (ƒ(x))</span>
                    </span>
                    <span class="text-3xs font-mono text-blue-300 bg-blue-900/50 px-2 py-0.5 rounded border border-blue-700/50">
                      {{ testReadResult.formula || (testParam && testParam.formula) }}
                    </span>
                  </div>
                  <div class="flex items-baseline space-x-2 pt-1">
                    <span class="text-2xl font-mono font-extrabold text-blue-400">
                      {{ getTestReadFormulaVal() }}
                    </span>
                    <span class="text-xs font-bold text-slate-300 font-mono">{{ testParam.unit }}</span>
                  </div>
                </div>

                <div class="grid grid-cols-2 gap-3">
                  <div class="p-2.5 bg-[#0B0F19]/80 rounded-lg border border-slate-800">
                    <div class="text-3xs text-slate-400 font-semibold uppercase">
                      {{ getTestReadFormulaVal() !== null ? $t('deviceDetail.modbusScaledValue') : $t('deviceDetail.decodedValue') }}
                    </div>
                    <div class="text-base font-mono font-bold text-emerald-400 mt-1">
                      {{ getTestReadModbusVal() }}
                      <span class="text-xs font-normal text-slate-300 ml-1">{{ testParam.unit }}</span>
                    </div>
                  </div>
                  <div class="p-2.5 bg-[#0B0F19]/80 rounded-lg border border-slate-800">
                    <div class="text-3xs text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.rawPduValue') }}</div>
                    <div class="text-xs font-mono text-slate-200 mt-1">
                      {{ testReadResult.raw_value !== undefined ? testReadResult.raw_value : '--' }}
                      <span v-if="testReadResult.raw_bytes_hex" class="block text-3xs text-slate-400 font-mono mt-0.5">
                        {{ $t('deviceDetail.bytes') }} {{ testReadResult.raw_bytes_hex }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- Diagnostics Metadata -->
            <div class="p-3 bg-[#0B0F19] rounded-xl border border-slate-800 grid grid-cols-3 gap-2 text-3xs font-mono text-slate-400">
              <div>
                <span class="text-slate-500 block uppercase">{{ $t('deviceDetail.functionCode') }}</span>
                <span class="text-slate-300 font-bold">FC {{ testReadResult.function_code || '--' }}</span>
              </div>
              <div>
                <span class="text-slate-500 block uppercase">{{ $t('deviceDetail.pduOffset') }}</span>
                <span class="text-slate-300 font-bold">{{ testReadResult.register_address !== undefined ? testReadResult.register_address : '--' }}</span>
              </div>
              <div>
                <span class="text-slate-500 block uppercase">{{ $t('common.timestamp') }}</span>
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
            {{ $t('common.close') }}
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
            <span>{{ $t('deviceDetail.readAgain') }}</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Global Floating Toast Notification -->
    <div
      v-if="toastMessage"
      class="fixed bottom-6 right-6 z-50 flex items-center space-x-3 px-4 py-3 rounded-2xl shadow-2xl border text-xs font-semibold backdrop-blur-md transition-all animate-fade-in"
      :class="toastType === 'error' ? 'bg-rose-950/90 border-rose-500/50 text-rose-200 shadow-rose-950/40' : 'bg-emerald-950/90 border-emerald-500/50 text-emerald-200 shadow-emerald-950/40'"
    >
      <svg v-if="toastType === 'error'" class="w-5 h-5 text-rose-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
      </svg>
      <svg v-else class="w-5 h-5 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
      </svg>
      <span>{{ toastMessage }}</span>
      <button type="button" @click="toastMessage = null" class="ml-2 hover:opacity-80 text-3xs font-mono">✕</button>
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
      toastMessage: null,
      toastType: 'success',
      toastTimer: null,
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
      chartValueMode: 'scaled', // 'scaled' (formula) or 'raw' (register data)
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
      chartResizeObserver: null,
      // Phase 3.1 Raw Telemetry
      rawRecords: [],
      rawTotal: 0,
      rawPage: 1,
      rawPageSize: 20,
      rawParamId: '',
      rawLoading: false,

      // Phase 3.3 Aggregation & Rollups
      aggregationDefinitions: [],
      aggregationDefinitionsLoading: false,
      aggregationResults: [],
      aggregationResultsLoading: false,
      aggregationTotal: 0,
      aggregationPage: 1,
      aggregationPageSize: 20,
      aggregationFilterDefId: '',
      aggregationFilterSource: '',
      aggregationFilterParamId: '',
      aggregationIdentifierSearch: '',
      customerViewMode: false,
      isAggDefModalOpen: false,
      isEditingAggDef: false,
      aggDefForm: {
        id: null,
        name: '',
        code: '',
        parameter_id: '',
        source_type: 'CUSTOMER_PROCESSED',
        function: 'AVG',
        interval_preset: '30m',
        interval_seconds: 1800,
        timezone: 'Asia/Jakarta',
        quality_policy: 'STRICT',
        enabled: true,
      },
      customerPreviewIdentifier: '',
      customerPreviewResult: null,
      historyDownsampleResolution: 'raw',
      autoRefreshAgg: true,
      aggPollTimer: null,

      // Bucket Samples Inspection Modal
      isBucketSamplesModalOpen: false,
      selectedBucketResult: null,
      bucketSamplesLoading: false,
      bucketSamplesData: null,
      bucketSamplesTimer: null,
    };
  },
  computed: {
    timeRangeOptions() {
      return [
        { label: '15m', value: '15m' },
        { label: this.$t('deviceDetail.timeRange1h'), value: '1h' },
        { label: this.$t('deviceDetail.timeRange6h'), value: '6h' },
        { label: this.$t('deviceDetail.timeRange24h'), value: '24h' },
        { label: this.$t('deviceDetail.timeRangeAll'), value: 'all' },
      ];
    },
    displayParameters() {
      if (!this.device || !this.device.parameters) return [];
      return this.device.parameters;
    },
    deviceQualitySummary() {
      if (!this.device || !this.device.parameters || this.device.parameters.length === 0) {
        return { good: 0, uncertain: 0, bad: 0, stale: 0, healthPercent: 100 };
      }
      let good = 0;
      let uncertain = 0;
      let bad = 0;
      let stale = 0;
      for (const p of this.device.parameters) {
        const q = this.getParamQuality(p);
        if (q === 'GOOD') good++;
        else if (q === 'UNCERTAIN') uncertain++;
        else if (q === 'BAD') bad++;
        else if (q === 'STALE') stale++;
        else good++;
      }
      const total = this.device.parameters.length;
      const healthPercent = total > 0 ? Math.round((good / total) * 100) : 100;
      return { good, uncertain, bad, stale, healthPercent };
    },
    historyTimeRangeLabel() {
      switch (this.historyTimeRange) {
        case '15m': return this.$t('deviceDetail.last15MinOption');
        case '1h': return this.$t('deviceDetail.last1HourOption');
        case '6h': return this.$t('deviceDetail.last6HoursOption');
        case '24h': return this.$t('deviceDetail.last24HoursOption');
        default: return this.$t('deviceDetail.allStoredHistory');
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
      if (this.chartValueMode === 'raw') {
        return 'RAW';
      }
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
      const vals = chronological.map(r => this.getTelemetryVal(r));
      let min = Math.min(...vals);
      let max = Math.max(...vals);

      // If all values are identical (flat line)
      if (min === max) {
        min = min - 1;
        max = max + 1;
      }

      const range = max - min || 1;
      const left = 60;
      const right = Math.max(left + 100, (this.containerWidth || 800) - 30);
      const top = 30;
      const bottom = Math.max(top + 80, (this.containerHeight || 320) - 48);
      const plotWidth = right - left;
      const plotHeight = bottom - top;

      const count = chronological.length;
      const maxVal = Math.max(...vals);
      const minVal = Math.min(...vals);

      return chronological.map((rec, idx) => {
        const val = this.getTelemetryVal(rec);
        const x = count === 1 ? (left + plotWidth / 2) : (left + (idx / (count - 1)) * plotWidth);
        const normalizedY = (val - min) / range;
        const y = bottom - (normalizedY * plotHeight);

        return {
          x: Number(x.toFixed(1)),
          y: Number(y.toFixed(1)),
          value: val,
          rawValue: (rec.raw_value !== undefined && rec.raw_value !== null) ? Number(rec.raw_value) : null,
          scaledValue: (rec.value !== undefined && rec.value !== null) ? Number(rec.value) : null,
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
      const bottom = Math.max(80, (this.containerHeight || 320) - 48);
      const first = pts[0];
      const last = pts[pts.length - 1];
      return `${curve} L ${last.x},${bottom} L ${first.x},${bottom} Z`;
    },
    chartGridTicks() {
      const top = 30;
      const bottom = Math.max(top + 80, (this.containerHeight || 320) - 48);
      const plotHeight = bottom - top;

      const records = this.chartFilteredRecords;
      if (!records || records.length === 0) {
        return [
          { y: top, label: '100.00' },
          { y: Math.round(top + plotHeight * 0.25), label: '75.00' },
          { y: Math.round(top + plotHeight * 0.50), label: '50.00' },
          { y: Math.round(top + plotHeight * 0.75), label: '25.00' },
          { y: bottom, label: '0.00' },
        ];
      }
      const vals = records.map(r => this.getTelemetryVal(r));
      let min = Math.min(...vals);
      let max = Math.max(...vals);
      if (min === max) {
        min = min - 1;
        max = max + 1;
      }
      const range = max - min;

      const ratios = [1.0, 0.75, 0.5, 0.25, 0.0];
      return ratios.map(r => {
        const val = min + range * r;
        const y = Number((bottom - r * plotHeight).toFixed(1));
        const label = (this.chartValueMode === 'raw' && Number.isInteger(val)) ? val.toString() : val.toFixed(2);
        return {
          y,
          value: val,
          label,
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
      const vals = records.map(r => this.getTelemetryVal(r));
      const minVal = Math.min(...vals);
      const maxVal = Math.max(...vals);
      const sum = vals.reduce((a, b) => a + b, 0);
      const avgVal = sum / vals.length;
      const deltaVal = maxVal - minVal;

      const minRec = records.find(r => this.getTelemetryVal(r) === minVal);
      const maxRec = records.find(r => this.getTelemetryVal(r) === maxVal);
      const latestVal = this.getTelemetryVal(latestRec);

      const fmtVal = (val) => {
        if (this.chartValueMode === 'raw' && Number.isInteger(val)) {
          return val.toString();
        }
        return val.toFixed(2);
      };

      return {
        latest: latestRec ? fmtVal(latestVal) : '--',
        latestQuality: latestRec.quality || 'GOOD',
        latestTime: this.formatTimeShort(latestRec.received_at || latestRec.timestamp),
        min: fmtVal(minVal),
        max: fmtVal(maxVal),
        avg: avgVal.toFixed(2),
        delta: fmtVal(deltaVal),
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
      const top = 30;
      const bottom = Math.max(top + 80, (this.containerHeight || 320) - 48);
      const plotHeight = bottom - top;
      if (!records || records.length === 0) return Math.round(top + plotHeight / 2);
      const vals = records.map(r => this.getTelemetryVal(r));
      let min = Math.min(...vals);
      let max = Math.max(...vals);
      if (min === max) {
        min = min - 1;
        max = max + 1;
      }
      const range = max - min || 1;
      const avg = Number(stats.avg);
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
    window.addEventListener('aggregation-result-event', this.onAggregationResultEvent);
    this.initChartResizeObserver();
  },
  beforeDestroy() {
    this.stopActivityPolling();
    this.stopAggregationPolling();
    window.removeEventListener('device-comm-event', this.onCommEvent);
    window.removeEventListener('device-telemetry-event', this.onTelemetryEvent);
    window.removeEventListener('aggregation-result-event', this.onAggregationResultEvent);
    if (this.chartResizeObserver) {
      this.chartResizeObserver.disconnect();
      this.chartResizeObserver = null;
    }
  },
  methods: {
    getTelemetryVal(rec) {
      if (!rec) return 0;
      if (this.chartValueMode === 'raw') {
        if (rec.raw_value !== undefined && rec.raw_value !== null) {
          return Number(rec.raw_value);
        }
      }
      return (rec.value !== undefined && rec.value !== null) ? Number(rec.value) : 0;
    },
    initChartResizeObserver() {
      this.$nextTick(() => {
        this.updateChartDimensions();
        if (window.ResizeObserver && this.$refs.chartContainer) {
          this.chartResizeObserver = new ResizeObserver(() => {
            this.updateChartDimensions();
          });
          this.chartResizeObserver.observe(this.$refs.chartContainer);
        }
      });
    },
    updateChartDimensions() {
      if (this.$refs.chartContainer) {
        const rect = this.$refs.chartContainer.getBoundingClientRect();
        if (rect.width > 0) this.containerWidth = Math.round(rect.width);
        if (rect.height > 0) this.containerHeight = Math.round(rect.height);
      }
    },
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
        } else if (this.activeTab === 'aggregation') {
          this.fetchAggregationDefinitions();
          this.fetchAggregationResults();
          this.startAggregationPolling();
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
    showToast(msg, type = 'success') {
      this.toastMessage = msg;
      this.toastType = type;
      if (this.toastTimer) clearTimeout(this.toastTimer);
      this.toastTimer = setTimeout(() => {
        this.toastMessage = null;
      }, 4000);
    },
    handleDeviceSaved() {
      this.showToast(this.$t('deviceModal.savedSuccess') || 'Device saved successfully!', 'success');
      this.loadDevice();
    },
    handleParamSaved() {
      this.showToast(this.$t('parameterModal.savedSuccess') || 'Parameter saved successfully!', 'success');
      this.loadDevice();
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
        alert(this.$t('deviceDetail.toggleDeviceError') + ': ' + err.message);
      }
    },
    async confirmDelete() {
      if (!confirm(this.$t('deviceDetail.deleteConfirm', { code: this.device.device_code || this.device.code }))) {
        return;
      }
      try {
        await this.$store.dispatch('deleteDevice', this.device.id);
        this.$router.push('/monitoring/devices');
      } catch (err) {
        alert(this.$t('deviceDetail.deleteDeviceError') + ': ' + err.message);
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
        alert(this.$t('deviceDetail.toggleParamError') + ': ' + err.message);
      }
    },
    async confirmDeleteParam(param) {
      if (!confirm(this.$t('deviceDetail.deleteParamConfirm', { code: param.parameter_code || param.code }))) {
        return;
      }
      try {
        await this.$store.dispatch('deleteParameter', {
          deviceId: this.device.id,
          paramId: param.id,
        });
        this.loadDevice();
      } catch (err) {
        alert(this.$t('deviceDetail.deleteParamError') + ': ' + err.message);
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
        case 'ONLINE':
        case 'CONNECTED': return 'bg-emerald-400 animate-pulse';
        case 'CONNECTING': return 'bg-blue-400 animate-pulse';
        case 'RECONNECTING': return 'bg-indigo-400 animate-pulse';
        case 'DEGRADED': return 'bg-amber-400 animate-pulse';
        case 'DISABLED': return 'bg-rose-400/80';
        case 'ERROR': return 'bg-rose-400';
        case 'OFFLINE':
        case 'DISCONNECTED': return 'bg-slate-500';
        default: return 'bg-purple-400';
      }
    },
    commTextClass(status) {
      switch (status) {
        case 'ONLINE':
        case 'CONNECTED': return 'text-emerald-400';
        case 'CONNECTING': return 'text-blue-400';
        case 'RECONNECTING': return 'text-indigo-400';
        case 'DEGRADED': return 'text-amber-400';
        case 'DISABLED': return 'text-rose-400/80';
        case 'ERROR': return 'text-rose-400';
        case 'OFFLINE':
        case 'DISCONNECTED': return 'text-slate-400';
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
        const data = (res && res.data !== undefined && typeof res.data === 'object' && res.data !== null) ? res.data : res;
        if (data && (data.connected || data.code === 'SUCCESS')) {
          const detail = data.message || `Physical link verified: connected in ${data.latency_ms} ms (status: ${data.code || 'SUCCESS'})`;
          this.commFeedback = { type: 'success', message: detail };
        } else {
          const code = (data && data.code) ? ` [${data.code}]` : '';
          const detail = (data && (data.error || data.message)) || 'Connection error';
          this.commFeedback = { type: 'error', message: `Link check failed${code}: ${detail}` };
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
        if (this.testReadResult && this.testReadResult.success) {
          const formulaVal = this.getTestReadFormulaValNum();
          if (formulaVal !== null) {
            this.testParam.current_value = formulaVal;
            this.testParam.current_formula_value = formulaVal;
          } else if (this.testReadResult.decoded_value !== undefined) {
            this.testParam.current_value = this.testReadResult.decoded_value;
          }
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
        this.$nextTick(() => {
          this.updateChartDimensions();
        });
      } else if (tab === 'raw') {
        this.fetchRawTelemetry();
        this.stopAggregationPolling();
      } else if (tab === 'aggregation') {
        this.fetchAggregationDefinitions();
        this.fetchAggregationResults();
        this.startAggregationPolling();
      } else {
        this.stopAggregationPolling();
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
            quality_reason: data.quality_reason,
            quality_flags: data.quality_flags,
            processed_value: data.processed_value,
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
            quality_reason: data.quality_reason,
            quality_flags: data.quality_flags,
            processed_value: data.processed_value,
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

        let url = `/api/devices/${this.device.id}/telemetry/history`;
        if (this.historyDownsampleResolution && this.historyDownsampleResolution !== 'raw') {
          url = `/api/devices/${this.device.id}/telemetry/downsampled`;
          params.resolution = this.historyDownsampleResolution;
        }

        const res = await axios.get(url, { params });
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
      this.containerWidth = Math.round(rect.width);
      this.containerHeight = Math.round(rect.height);

      // SVG coordinates are 1:1 with container screen pixels
      const svgX = mouseX;

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
    isParamHeld(param) {
      if (!param) return false;
      if (this.liveTelemetry[param.id] && this.liveTelemetry[param.id].is_held_value) {
        return true;
      }
      return !!param.is_current_held;
    },
    getParamFormulaValue(param) {
      if (!param) return null;
      if (this.liveTelemetry[param.id] && this.liveTelemetry[param.id].formula_value !== undefined && this.liveTelemetry[param.id].formula_value !== null) {
        return Number(this.liveTelemetry[param.id].formula_value).toFixed(param.precision || 2);
      }
      if (param.current_formula_value !== undefined && param.current_formula_value !== null) {
        return Number(param.current_formula_value).toFixed(param.precision || 2);
      }
      return null;
    },
    getTestReadFormulaValNum() {
      if (!this.testReadResult) return null;
      if (this.testReadResult.formula_value !== undefined && this.testReadResult.formula_value !== null) {
        return Number(this.testReadResult.formula_value);
      }
      const formulaStr = (this.testReadResult.formula || (this.testParam && this.testParam.formula) || '').trim();
      if (!formulaStr) return null;
      try {
        const x = this.getTestReadModbusValNum();
        const raw = this.testReadResult.raw_value !== undefined ? this.testReadResult.raw_value : x;
        const expr = formulaStr
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
        const res = Function(`"use strict"; return (${expr});`)();
        if (typeof res === 'number' && !isNaN(res) && isFinite(res)) {
          return res;
        }
      } catch (e) {}
      return null;
    },
    getTestReadFormulaVal() {
      const num = this.getTestReadFormulaValNum();
      if (num === null) return null;
      return num.toFixed(this.testParam ? this.testParam.precision : 2);
    },
    getTestReadModbusValNum() {
      if (!this.testReadResult) return 0;
      if (this.testReadResult.scaled_value !== undefined && this.testReadResult.scaled_value !== null) {
        return Number(this.testReadResult.scaled_value);
      }
      if (this.testReadResult.decoded_value !== undefined && this.testReadResult.decoded_value !== null) {
        return Number(this.testReadResult.decoded_value);
      }
      return 0;
    },
    getTestReadModbusVal() {
      if (!this.testReadResult) return '--';
      if (this.testReadResult.scaled_value !== undefined && this.testReadResult.scaled_value !== null) {
        return Number(this.testReadResult.scaled_value).toFixed(this.testParam ? this.testParam.precision : 2);
      }
      if (this.testReadResult.decoded_value !== undefined && this.testReadResult.decoded_value !== null) {
        return Number(this.testReadResult.decoded_value).toFixed(this.testParam ? this.testParam.precision : 2);
      }
      return '--';
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
        case 'STALE': return 'bg-purple-500/10 text-purple-400 border border-purple-500/20';
        default: return 'bg-slate-500/10 text-slate-400 border border-slate-500/20';
      }
    },
    chartDotColor(quality) {
      switch (quality) {
        case 'BAD': return '#F43F5E';
        case 'UNCERTAIN': return '#F59E0B';
        case 'STALE': return '#A855F7';
        case 'GOOD': default: return '#3B82F6';
      }
    },
    getParamQualityReason(param) {
      const q = this.getParamQuality(param);
      if (q === 'GOOD') {
        return '';
      }
      if (this.liveTelemetry[param.id] && this.liveTelemetry[param.id].quality_reason && this.liveTelemetry[param.id].quality_reason !== 'NONE') {
        return this.liveTelemetry[param.id].quality_reason;
      }
      if (param.current_quality === 'GOOD') {
        return '';
      }
      if (param.current_quality_reason && param.current_quality_reason !== 'NONE') {
        return param.current_quality_reason;
      }
      return '';
    },
    formatAuditActionName(action) {
      if (!action) return this.$t('deviceDetail.auditActionDefault');
      const actionKey = 'deviceDetail.auditAction_' + action.toLowerCase();
      const val = this.$t(actionKey);
      if (val && val !== actionKey) return val;
      return action.replace(/_/g, ' ');
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
        return trail.details || this.$t('deviceDetail.auditCommErrorDesc');
      }
      if (trail.action === 'COMMUNICATION_RESTORED') {
        return this.$t('deviceDetail.auditCommRestoredDesc');
      }
      const parsed = this.parseSafeJson(trail.details);
      if (parsed) {
        const after = this.parseSafeJson(parsed.after);
        if (after && typeof after === 'object') {
          if (trail.action.includes('PARAMETER')) {
            const pName = after.parameter_name || after.name || after.parameter_code || 'Parameter';
            const pCode = after.parameter_code || after.code || '';
            return this.$t('deviceDetail.auditUpdateParamDesc', { name: pName, code: pCode });
          }
          if (trail.action.includes('DEVICE')) {
            const dName = after.device_name || after.name || after.device_code || 'Device';
            return this.$t('deviceDetail.auditUpdateDeviceDesc', { name: dName });
          }
        }
      }
      return trail.details || this.$t('deviceDetail.auditDefaultDesc');
    },
    getAuditChips(trail) {
      if (!trail || !trail.details) return [];
      const chips = [];
      const parsed = this.parseSafeJson(trail.details);
      if (parsed) {
        const after = this.parseSafeJson(parsed.after) || parsed;
        if (after && typeof after === 'object') {
          if (after.parameter_code) chips.push({ label: this.$t('deviceDetail.code'), val: after.parameter_code });
          if (after.parameter_name) chips.push({ label: this.$t('deviceDetail.colName'), val: after.parameter_name });
          if (after.data_type) chips.push({ label: this.$t('deviceDetail.dataType'), val: after.data_type });
          if (after.register_address !== undefined) chips.push({ label: this.$t('deviceDetail.register'), val: `#${after.register_address}` });
          if (after.scale !== undefined) chips.push({ label: 'Scale', val: after.scale });
          if (after.offset !== undefined) chips.push({ label: 'Offset', val: after.offset });
          if (after.unit) chips.push({ label: this.$t('deviceDetail.unit'), val: after.unit });
          if (after.enabled !== undefined) chips.push({ label: this.$t('deviceDetail.status'), val: after.enabled ? this.$t('deviceDetail.active') : this.$t('deviceDetail.disabled') });
          if (after.device_name) chips.push({ label: this.$t('deviceDetail.deviceName'), val: after.device_name });
          if (after.device_code) chips.push({ label: this.$t('deviceDetail.deviceCode'), val: after.device_code });
          if (after.status) chips.push({ label: this.$t('deviceDetail.adminStatus'), val: after.status });
          if (after.serial_port) chips.push({ label: this.$t('deviceDetail.serialPort'), val: after.serial_port });
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

    // ==========================================
    // Phase 3.3 Aggregation & Rollups Methods
    // ==========================================
    async fetchAggregationDefinitions() {
      if (!this.device) return;
      this.aggregationDefinitionsLoading = true;
      try {
        const res = await axios.get('/api/aggregations/definitions', {
          params: { device_id: this.device.id },
        });
        if (res.data && res.data.data) {
          this.aggregationDefinitions = res.data.data;
        }
      } catch (err) {
        console.error('Failed to fetch aggregation definitions:', err);
      } finally {
        this.aggregationDefinitionsLoading = false;
      }
    },
    async fetchAggregationResults(silent = false) {
      if (!this.device) return;
      if (!silent) this.aggregationResultsLoading = true;
      try {
        const params = {
          device_id: this.device.id,
          page: this.aggregationPage,
          page_size: this.aggregationPageSize,
        };
        if (this.aggregationFilterDefId) params.definition_id = this.aggregationFilterDefId;
        if (this.aggregationFilterSource) params.source_type = this.aggregationFilterSource;
        if (this.aggregationFilterParamId) params.parameter_id = this.aggregationFilterParamId;
        if (this.aggregationIdentifierSearch) params.identifier = this.aggregationIdentifierSearch.trim();

        const res = await axios.get('/api/aggregations/results', { params });
        if (res.data && res.data.data) {
          this.aggregationResults = res.data.data.items || [];
          this.aggregationTotal = res.data.data.total || 0;
        }
      } catch (err) {
        if (!silent) console.error('Failed to fetch aggregation results:', err);
      } finally {
        if (!silent) this.aggregationResultsLoading = false;
      }
    },
    openCreateAggDefModal() {
      this.isEditingAggDef = false;
      this.aggDefForm = {
        id: null,
        name: '',
        code: '',
        parameter_id: this.device.parameters && this.device.parameters.length ? this.device.parameters[0].id : '',
        source_type: 'CUSTOMER_PROCESSED',
        function: 'AVG',
        interval_preset: '30m',
        interval_seconds: 1800,
        timezone: 'Asia/Jakarta',
        quality_policy: 'STRICT',
        enabled: true,
      };
      this.isAggDefModalOpen = true;
    },
    openEditAggDefModal(def) {
      this.isEditingAggDef = true;
      let preset = 'custom';
      if (def.interval_seconds === 120) preset = '2m';
      else if (def.interval_seconds === 300) preset = '5m';
      else if (def.interval_seconds === 600) preset = '10m';
      else if (def.interval_seconds === 900) preset = '15m';
      else if (def.interval_seconds === 1800) preset = '30m';
      else if (def.interval_seconds === 3600) preset = '60m';

      this.aggDefForm = {
        id: def.id,
        name: def.name,
        code: def.code,
        parameter_id: def.parameter_id,
        source_type: def.source_type,
        function: def.function,
        interval_preset: preset,
        interval_seconds: def.interval_seconds,
        timezone: def.timezone || 'Asia/Jakarta',
        quality_policy: def.quality_policy || 'STRICT',
        enabled: def.enabled !== false,
      };
      this.isAggDefModalOpen = true;
    },
    onIntervalPresetChange() {
      const p = this.aggDefForm.interval_preset;
      if (p === '2m') this.aggDefForm.interval_seconds = 120;
      else if (p === '5m') this.aggDefForm.interval_seconds = 300;
      else if (p === '10m') this.aggDefForm.interval_seconds = 600;
      else if (p === '15m') this.aggDefForm.interval_seconds = 900;
      else if (p === '30m') this.aggDefForm.interval_seconds = 1800;
      else if (p === '60m') this.aggDefForm.interval_seconds = 3600;
    },
    async saveAggregationDefinition() {
      if (!this.aggDefForm.name || !this.aggDefForm.code || !this.aggDefForm.parameter_id) {
        alert(this.$t('validation.required') || 'Required field missing');
        return;
      }
      try {
        if (this.aggDefForm.parameter_id === 'ALL') {
          const baseName = this.aggDefForm.name.trim();
          const baseCode = this.aggDefForm.code.trim();
          const paramsList = this.device.parameters || [];
          if (paramsList.length === 0) {
            alert('No parameters found on this device');
            return;
          }

          let createdCount = 0;
          for (const p of paramsList) {
            const pSuffix = p.parameter_code.toLowerCase().replace(/[^a-z0-9_]/g, '_');
            const itemCode = `${baseCode}_${pSuffix}`;
            const itemName = `${baseName} - ${p.parameter_name}`;
            const payload = {
              device_id: this.device.id,
              parameter_id: Number(p.id),
              name: itemName,
              code: itemCode,
              source_type: this.aggDefForm.source_type,
              function: this.aggDefForm.function,
              interval_seconds: Number(this.aggDefForm.interval_seconds),
              timezone: this.aggDefForm.timezone,
              quality_policy: this.aggDefForm.quality_policy,
              enabled: this.aggDefForm.enabled,
            };
            await axios.post('/api/aggregations/definitions', payload);
            createdCount++;
          }
          this.showToast(this.$t('aggregation.allParametersSuccess', { count: createdCount }) || `Created ${createdCount} aggregation definitions!`);
        } else {
          const payload = {
            device_id: this.device.id,
            parameter_id: Number(this.aggDefForm.parameter_id),
            name: this.aggDefForm.name.trim(),
            code: this.aggDefForm.code.trim(),
            source_type: this.aggDefForm.source_type,
            function: this.aggDefForm.function,
            interval_seconds: Number(this.aggDefForm.interval_seconds),
            timezone: this.aggDefForm.timezone,
            quality_policy: this.aggDefForm.quality_policy,
            enabled: this.aggDefForm.enabled,
          };

          if (this.isEditingAggDef && this.aggDefForm.id) {
            await axios.put(`/api/aggregations/definitions/${this.aggDefForm.id}`, payload);
            this.showToast(this.$t('deviceModal.savedSuccess') || 'Saved successfully!');
          } else {
            await axios.post('/api/aggregations/definitions', payload);
            this.showToast(this.$t('deviceModal.savedSuccess') || 'Created successfully!');
          }
        }
        this.isAggDefModalOpen = false;
        await this.fetchAggregationDefinitions();
        await this.fetchAggregationResults();
      } catch (err) {
        alert((err.response && err.response.data && err.response.data.error) || err.message);
      }
    },
    async toggleAggDefStatus(def) {
      try {
        await axios.put(`/api/aggregations/definitions/${def.id}`, {
          enabled: !def.enabled,
        });
        def.enabled = !def.enabled;
        this.showToast(this.$t('deviceModal.savedSuccess') || 'Status updated!');
      } catch (err) {
        alert(err.message);
      }
    },
    async deleteAggregationDefinition(def) {
      if (!confirm(this.$t('aggregation.deleteDefConfirm', { name: def.name }))) return;
      try {
        await axios.delete(`/api/aggregations/definitions/${def.id}`);
        this.showToast(this.$t('deviceModal.savedSuccess') || 'Deleted successfully!');
        await this.fetchAggregationDefinitions();
        await this.fetchAggregationResults();
      } catch (err) {
        alert(err.message);
      }
    },
    async runAggBucket(defId) {
      try {
        await axios.post('/api/aggregations/run', { definition_id: defId });
        this.showToast(this.$t('aggregation.recalculatedSuccess') || 'Bucket calculated successfully!');
        await this.fetchAggregationResults();
      } catch (err) {
        alert((err.response && err.response.data && err.response.data.error) || err.message);
      }
    },
    async runAllAggBuckets() {
      try {
        await axios.post('/api/aggregations/run', { device_id: this.device.id });
        this.showToast(this.$t('aggregation.runningAllSuccess') || 'Aggregation job completed!');
        await this.fetchAggregationResults();
      } catch (err) {
        alert((err.response && err.response.data && err.response.data.error) || err.message);
      }
    },
    async queryCustomerData(identifier) {
      const idToLookup = identifier || this.customerPreviewIdentifier || (this.aggregationResults[0] && this.aggregationResults[0].identifier);
      if (!idToLookup) return;
      this.customerPreviewIdentifier = idToLookup;
      this.customerPreviewLoading = true;
      try {
        const res = await axios.get(`/api/customer/aggregated-data/${idToLookup}`);
        this.customerPreviewResult = res.data;
      } catch (err) {
        this.customerPreviewResult = {
          error: (err.response && err.response.data && err.response.data.error) || err.message,
          status: err.response ? err.response.status : 500,
        };
      } finally {
        this.customerPreviewLoading = false;
      }
    },
    prevAggPage() {
      if (this.aggregationPage > 1) {
        this.aggregationPage--;
        this.fetchAggregationResults();
      }
    },
    nextAggPage() {
      if (this.aggregationPage * this.aggregationPageSize < this.aggregationTotal) {
        this.aggregationPage++;
        this.fetchAggregationResults();
      }
    },
    startAggregationPolling() {
      this.stopAggregationPolling();
      if (!this.autoRefreshAgg) return;
      this.aggPollTimer = setInterval(() => {
        if (this.activeTab === 'aggregation' && !this.aggregationResultsLoading) {
          this.fetchAggregationResults(true);
        }
      }, 3000);
    },
    stopAggregationPolling() {
      if (this.aggPollTimer) {
        clearInterval(this.aggPollTimer);
        this.aggPollTimer = null;
      }
    },
    toggleAutoRefreshAgg() {
      this.autoRefreshAgg = !this.autoRefreshAgg;
      if (this.autoRefreshAgg) {
        this.startAggregationPolling();
        this.showToast((this.$t('aggregation.autoRefresh') || 'Auto Refresh') + ' Enabled (3s)');
      } else {
        this.stopAggregationPolling();
        this.showToast((this.$t('aggregation.autoRefresh') || 'Auto Refresh') + ' Disabled');
      }
    },
    isBucketActive(res) {
      if (!res || !res.period_end) return false;
      return new Date(res.period_end) > new Date();
    },
    onAggregationResultEvent(e) {
      if (!this.device || !e.detail) return;
      const res = e.detail;
      if (res.device_id !== this.device.id) return;

      const idx = this.aggregationResults.findIndex(r => r.identifier === res.identifier && r.parameter_id === res.parameter_id);
      if (idx !== -1) {
        this.$set(this.aggregationResults, idx, res);
      } else {
        this.aggregationResults.unshift(res);
        this.aggregationTotal++;
      }
    },
    formatPeriodWindow(start, end) {
      if (!start || !end) return '';
      const s = new Date(start);
      // Samples in bucket [start, end) end at end - 1s (e.g. 26:00 to 27:59)
      const e = new Date(new Date(end).getTime() - 1000);
      const sStr = s.toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' });
      const eStr = e.toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' });
      return `${sStr} → ${eStr}`;
    },
    formatTimestampExact(ts) {
      if (!ts) return '';
      const d = new Date(ts);
      return d.toLocaleTimeString([], { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' });
    },
    async openBucketSamplesModal(res) {
      this.selectedBucketResult = res;
      this.isBucketSamplesModalOpen = true;
      await this.fetchBucketSamples();
      if (this.isBucketActive(res)) {
        this.startBucketSamplesPolling();
      }
    },
    closeBucketSamplesModal() {
      this.isBucketSamplesModalOpen = false;
      this.selectedBucketResult = null;
      this.bucketSamplesData = null;
      this.stopBucketSamplesPolling();
    },
    async fetchBucketSamples() {
      if (!this.selectedBucketResult) return;
      this.bucketSamplesLoading = true;
      try {
        const params = {
          device_id: this.device.id,
          parameter_id: this.selectedBucketResult.parameter_id,
          period_start: this.selectedBucketResult.period_start,
          period_end: this.selectedBucketResult.period_end,
        };
        let url = '/api/aggregations/samples';
        if (this.selectedBucketResult.id) {
          url = `/api/aggregations/results/${this.selectedBucketResult.id}/samples`;
        }
        const res = await axios.get(url, { params });
        if (res.data && res.data.data) {
          this.bucketSamplesData = res.data.data;
        }
      } catch (err) {
        console.error('Failed to fetch bucket samples:', err);
      } finally {
        this.bucketSamplesLoading = false;
      }
    },
    startBucketSamplesPolling() {
      this.stopBucketSamplesPolling();
      this.bucketSamplesTimer = setInterval(() => {
        if (this.isBucketSamplesModalOpen && this.selectedBucketResult && this.isBucketActive(this.selectedBucketResult)) {
          this.fetchBucketSamples();
        }
      }, 2000);
    },
    stopBucketSamplesPolling() {
      if (this.bucketSamplesTimer) {
        clearInterval(this.bucketSamplesTimer);
        this.bucketSamplesTimer = null;
      }
    },
  },
};
</script>
