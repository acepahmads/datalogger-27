<template>
  <div class="space-y-6 font-sans">
    <!-- Header -->
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white">
          {{ $t('devices.title') }}
        </h1>
        <p class="text-xs text-slate-400 mt-0.5">
          {{ $t('devices.subtitle') }}
        </p>
      </div>

      <div class="flex items-center space-x-3">
        <!-- Add Device Button -->
        <button
          @click="openAddModal"
          class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold transition flex items-center space-x-2 shadow-lg shadow-blue-600/20"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
          </svg>
          <span>{{ $t('devices.registerDevice') }}</span>
        </button>
      </div>
    </div>

    <!-- Quick Stats Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="saas-card p-4 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('devices.totalRegistered') }}</div>
        <div class="text-xl font-bold text-white font-mono">{{ totalDevices }}</div>
        <div class="text-3xs text-slate-500">{{ $t('devices.totalRegisteredDesc') }}</div>
      </div>

      <div class="saas-card p-4 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('devices.onlineHealthy') }}</div>
        <div class="text-xl font-bold text-emerald-400 font-mono flex items-center space-x-2">
          <span>{{ onlineDevicesCount }}</span>
          <span class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></span>
        </div>
        <div class="text-3xs text-slate-500">{{ $t('devices.onlineHealthyDesc') }}</div>
      </div>

      <div class="saas-card p-4 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('devices.offlineError') }}</div>
        <div class="text-xl font-bold text-rose-400 font-mono">{{ offlineDevicesCount }}</div>
        <div class="text-3xs text-slate-500">{{ $t('devices.offlineErrorDesc') }}</div>
      </div>

      <div class="saas-card p-4 space-y-1">
        <div class="text-3xs text-slate-400 font-semibold uppercase tracking-wider">{{ $t('devices.maintenanceInactive') }}</div>
        <div class="text-xl font-bold text-amber-400 font-mono">{{ maintenanceDevicesCount }}</div>
        <div class="text-3xs text-slate-500">{{ $t('devices.maintenanceDesc') }}</div>
      </div>
    </div>

    <!-- Filter & Action Bar -->
    <div class="saas-card p-4 flex flex-col md:flex-row items-center justify-between gap-3 text-xs">
      <div class="flex flex-wrap items-center gap-3 w-full md:w-auto">
        <!-- Search -->
        <div class="relative flex-1 md:w-64">
          <input
            v-model="searchQuery"
            @input="onFilterChange"
            type="text"
            :placeholder="$t('devices.searchPlaceholder')"
            class="w-full pl-9 pr-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-white placeholder-slate-500 focus:outline-none focus:border-blue-500"
          />
          <svg class="w-4 h-4 text-slate-400 absolute left-3 top-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
          </svg>
        </div>

        <!-- Status Filter (Administrative) -->
        <select
          v-model="filterStatus"
          @change="onFilterChange"
          class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-slate-300 focus:outline-none focus:border-blue-500"
        >
          <option value="">{{ $t('devices.allAdminStatuses') }}</option>
          <option value="ACTIVE">{{ $t('status.active') }}</option>
          <option value="INACTIVE">{{ $t('status.inactive') }}</option>
          <option value="MAINTENANCE">{{ $t('status.maintenance') }}</option>
          <option value="DISABLED">{{ $t('status.disabled') }}</option>
        </select>

        <!-- Connection Status Filter (Communication) -->
        <select
          v-model="filterConnStatus"
          @change="onFilterChange"
          class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-slate-300 focus:outline-none focus:border-blue-500"
        >
          <option value="">{{ $t('devices.allCommStatuses') }}</option>
          <option value="ONLINE">{{ $t('status.online') }}</option>
          <option value="OFFLINE">{{ $t('status.offline') }}</option>
          <option value="CONNECTING">{{ $t('status.connecting') }}</option>
          <option value="ERROR">{{ $t('status.error') }}</option>
          <option value="UNKNOWN">{{ $t('status.unknown') }}</option>
        </select>

        <!-- Protocol Filter -->
        <select
          v-model="filterProtocol"
          @change="onFilterChange"
          class="px-3 py-1.5 bg-[#0B0F19] border border-slate-700/80 rounded-xl text-xs text-slate-300 focus:outline-none focus:border-blue-500"
        >
          <option value="">{{ $t('devices.allProtocols') }}</option>
          <option value="MODBUS_TCP">MODBUS_TCP</option>
          <option value="MODBUS_RTU">MODBUS_RTU</option>
          <option value="MQTT">MQTT</option>
          <option value="HTTP">HTTP</option>
          <option value="TCP">TCP</option>
          <option value="UDP">UDP</option>
          <option value="SERIAL">SERIAL</option>
          <option value="WEBSOCKET">WEBSOCKET</option>
          <option value="CUSTOM">CUSTOM</option>
        </select>
      </div>

      <!-- View Switcher & Refresh -->
      <div class="flex items-center space-x-2 self-end md:self-auto">
        <button
          @click="viewMode = 'table'"
          :class="viewMode === 'table' ? 'bg-blue-600 text-white' : 'bg-slate-800 text-slate-400 hover:text-white'"
          class="p-1.5 rounded-lg transition"
          :title="$t('devices.tableView')"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 10h16M4 14h16M4 18h16"></path>
          </svg>
        </button>
        <button
          @click="viewMode = 'grid'"
          :class="viewMode === 'grid' ? 'bg-blue-600 text-white' : 'bg-slate-800 text-slate-400 hover:text-white'"
          class="p-1.5 rounded-lg transition"
          :title="$t('devices.gridView')"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"></path>
          </svg>
        </button>
        <button
          @click="loadDevices"
          class="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white transition"
          :title="$t('common.refresh')"
        >
          <svg class="w-4 h-4" :class="{ 'animate-spin': loading }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path>
          </svg>
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading && devices.length === 0" class="py-16 text-center text-slate-400">
      <div class="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-2"></div>
      <p class="text-xs">{{ $t('devices.loadingDevices') }}</p>
    </div>

    <!-- TABLE VIEW -->
    <div v-else-if="viewMode === 'table'" class="saas-card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-xs font-sans">
          <thead>
            <tr class="border-b border-slate-800/80 bg-[#0B0F19]/40 text-3xs uppercase font-semibold text-slate-400 tracking-wider">
              <th @click="sortBy('device_code')" class="py-3 px-4 cursor-pointer hover:text-slate-200">
                {{ $t('devices.code') }} <span v-if="sortColumn === 'device_code'">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th @click="sortBy('device_name')" class="py-3 px-4 cursor-pointer hover:text-slate-200">
                {{ $t('devices.deviceName') }} <span v-if="sortColumn === 'device_name'">{{ sortDirection === 'asc' ? '↑' : '↓' }}</span>
              </th>
              <th class="py-3 px-4">{{ $t('devices.protocol') }}</th>
              <th class="py-3 px-4">{{ $t('devices.manufacturerModel') }}</th>
              <th class="py-3 px-4">{{ $t('devices.location') }}</th>
              <th @click="sortBy('status')" class="py-3 px-4 cursor-pointer hover:text-slate-200">
                {{ $t('devices.adminStatus') }}
              </th>
              <th @click="sortBy('connection_status')" class="py-3 px-4 cursor-pointer hover:text-slate-200">
                {{ $t('devices.commStatus') }}
              </th>
              <th class="py-3 px-4">{{ $t('devices.lastSeen') }}</th>
              <th class="py-3 px-4">{{ $t('devices.lastData') }}</th>
              <th class="py-3 px-4">{{ $t('devices.enabled') }}</th>
              <th class="py-3 px-4 text-right">{{ $t('devices.actions') }}</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-800/60">
            <tr v-for="dev in devices" :key="dev.id" class="hover:bg-slate-800/30 transition-colors">
              <!-- Code -->
              <td class="py-3 px-4">
                <router-link :to="'/monitoring/devices/' + dev.id" class="font-mono font-bold text-blue-400 hover:text-blue-300">
                  {{ dev.device_code || dev.code }}
                </router-link>
              </td>

              <!-- Name -->
              <td class="py-3 px-4 font-semibold text-white">
                <router-link :to="'/monitoring/devices/' + dev.id" class="hover:text-blue-300">
                  {{ dev.device_name || dev.name }}
                </router-link>
              </td>

              <!-- Protocol / Type -->
              <td class="py-3 px-4">
                <span class="px-2 py-0.5 rounded text-3xs font-mono font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20">
                  {{ (dev.connection && dev.connection.protocol) || dev.device_type || 'MODBUS_TCP' }}
                </span>
              </td>

              <!-- Manufacturer / Model -->
              <td class="py-3 px-4 text-slate-300">
                <div>{{ dev.manufacturer || '--' }}</div>
                <div class="text-3xs text-slate-500 font-mono">{{ dev.model }}</div>
              </td>

              <!-- Location -->
              <td class="py-3 px-4 text-slate-300">
                {{ dev.location || '--' }}
              </td>

              <!-- Admin Status -->
              <td class="py-3 px-4">
                <span class="px-2 py-0.5 rounded text-3xs font-bold font-mono" :class="statusBadgeClass(dev.status)">
                  {{ dev.status }}
                </span>
              </td>

              <!-- Comm Status -->
              <td class="py-3 px-4">
                <span class="flex items-center space-x-1.5 font-mono text-3xs font-bold" :class="commTextClass(dev.connection_status)">
                  <span class="w-1.5 h-1.5 rounded-full" :class="commDotClass(dev.connection_status)"></span>
                  <span>{{ dev.connection_status || 'UNKNOWN' }}</span>
                </span>
              </td>

              <!-- Last Seen -->
              <td class="py-3 px-4 font-mono text-3xs text-slate-400">
                {{ formatTimestamp(dev.last_seen_at || dev.last_communication) }}
              </td>

              <!-- Last Data -->
              <td class="py-3 px-4 font-mono text-3xs text-slate-400">
                {{ formatTimestamp(dev.last_data_at) }}
              </td>

              <!-- Enabled -->
              <td class="py-3 px-4">
                <button
                  @click="toggleDeviceEnabled(dev)"
                  class="relative inline-flex h-4 w-8 flex-shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none"
                  :class="dev.enabled ? 'bg-blue-600' : 'bg-slate-700'"
                >
                  <span
                    class="pointer-events-none inline-block h-3 w-3 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out"
                    :class="dev.enabled ? 'translate-x-4' : 'translate-x-0'"
                  ></span>
                </button>
              </td>

              <!-- Actions -->
              <td class="py-3 px-4 text-right space-x-2">
                <router-link :to="'/monitoring/devices/' + dev.id + '?tab=telemetry'" class="inline-flex items-center space-x-1 text-emerald-400 hover:text-emerald-300 font-semibold text-2xs bg-emerald-950/50 hover:bg-emerald-900/60 px-2 py-0.5 rounded border border-emerald-800/60 transition">
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
                  <span>{{ $t('devices.liveData') }}</span>
                </router-link>
                <router-link :to="'/monitoring/devices/' + dev.id" class="text-blue-400 hover:text-blue-300 font-semibold text-2xs transition">
                  {{ $t('common.view') }}
                </router-link>
                <button @click="openEditModal(dev)" class="text-slate-300 hover:text-white font-semibold text-2xs transition">
                  {{ $t('common.edit') }}
                </button>
                <button @click="confirmDelete(dev)" class="text-rose-400 hover:text-rose-300 font-semibold text-2xs transition">
                  {{ $t('common.delete') }}
                </button>
              </td>
            </tr>

            <tr v-if="devices.length === 0">
              <td colspan="11" class="py-12 text-center text-slate-400">
                {{ $t('devices.noDevices') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination Footer -->
      <div class="px-4 py-3 border-t border-slate-800/80 bg-[#0B0F19]/40 flex flex-col md:flex-row items-center justify-between gap-3 text-xs text-slate-400">
        <div>
          {{ $t('devices.showingCount', { count: devices.length, total: totalDevices }) }}
        </div>

        <div class="flex items-center space-x-2">
          <button
            @click="changePage(currentPage - 1)"
            :disabled="currentPage <= 1"
            class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            {{ $t('common.previous') }}
          </button>
          <span class="px-2 font-mono text-slate-300">{{ $t('common.page') }} {{ currentPage }} / {{ totalPages }}</span>
          <button
            @click="changePage(currentPage + 1)"
            :disabled="currentPage >= totalPages"
            class="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 disabled:opacity-40 disabled:cursor-not-allowed transition"
          >
            {{ $t('common.next') }}
          </button>
        </div>
      </div>
    </div>

    <!-- GRID VIEW -->
    <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div v-for="dev in devices" :key="dev.id" class="saas-card p-5 space-y-4 hover:border-slate-700 transition">
        <!-- Card Header -->
        <div class="flex items-start justify-between border-b border-slate-800/80 pb-3">
          <div>
            <div class="flex items-center space-x-2">
              <span class="px-2 py-0.5 rounded-md bg-blue-500/10 text-blue-400 border border-blue-500/20 font-mono text-3xs font-semibold">
                {{ dev.device_code || dev.code }}
              </span>
              <span class="flex items-center text-3xs font-mono font-semibold" :class="commTextClass(dev.connection_status)">
                <span class="w-1.5 h-1.5 rounded-full mr-1.5" :class="commDotClass(dev.connection_status)"></span>
                {{ dev.connection_status || 'UNKNOWN' }}
              </span>
            </div>
            <router-link :to="'/monitoring/devices/' + dev.id" class="text-sm font-bold text-white mt-1.5 block hover:text-blue-300">
              {{ dev.device_name || dev.name }}
            </router-link>
            <div class="text-2xs text-slate-400 mt-0.5">{{ dev.location || $t('devices.noLocation') }}</div>
          </div>

          <span class="px-2 py-0.5 rounded text-3xs font-bold font-mono" :class="statusBadgeClass(dev.status)">
            {{ dev.status }}
          </span>
        </div>

        <!-- Connection Details -->
        <div class="bg-[#0B0F19]/80 p-3 rounded-xl border border-slate-800/80 text-xs font-sans space-y-1.5">
          <div class="flex justify-between">
            <span class="text-slate-400">{{ $t('devices.protocol') }}:</span>
            <span class="text-blue-300 font-semibold font-mono text-2xs">
              {{ (dev.connection && dev.connection.protocol) || dev.device_type || 'MODBUS_TCP' }}
            </span>
          </div>
          <div class="flex justify-between" v-if="dev.connection">
            <span class="text-slate-400">{{ $t('devices.target') }}:</span>
            <span class="font-mono text-slate-200 text-2xs">
              {{ dev.connection.host || dev.connection.serial_port || dev.connection.address || '--' }}
            </span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">{{ $t('devices.latency') }}:</span>
            <span class="font-mono text-emerald-400 font-semibold text-2xs">{{ dev.latency_ms || 0 }} ms</span>
          </div>
        </div>

        <!-- Parameters Preview -->
        <div v-if="dev.parameters && dev.parameters.length > 0" class="space-y-1.5">
          <div class="text-3xs uppercase tracking-wider text-slate-400 font-semibold">{{ $t('devices.activeChannels') }} ({{ dev.parameters.length }})</div>
          <div v-for="param in dev.parameters.slice(0, 3)" :key="param.id"
               class="p-2 rounded-lg bg-[#0B0F19]/60 border border-slate-800/80 flex items-center justify-between text-xs">
            <span class="text-slate-200 text-2xs font-medium">{{ param.parameter_name || param.name }}</span>
            <span class="text-xs font-mono font-bold text-blue-400">
              {{ param.current_value !== null && param.current_value !== undefined ? param.current_value.toFixed(1) : '--' }}
              <span class="text-slate-400 text-3xs">{{ param.unit }}</span>
            </span>
          </div>
        </div>

        <!-- Card Footer Actions -->
        <div class="pt-3 border-t border-slate-800/80 flex items-center justify-between">
          <div class="flex items-center space-x-2">
            <router-link :to="'/monitoring/devices/' + dev.id" class="text-xs font-bold text-blue-400 hover:text-blue-300">
              {{ $t('devices.details') }} →
            </router-link>
            <router-link :to="'/monitoring/devices/' + dev.id + '?tab=telemetry'" class="inline-flex items-center space-x-1 text-emerald-400 hover:text-emerald-300 font-semibold text-2xs bg-emerald-950/50 hover:bg-emerald-900/60 px-2 py-0.5 rounded border border-emerald-800/60 transition">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
              <span>{{ $t('devices.live') }}</span>
            </router-link>
          </div>
          <div class="space-x-2">
            <button @click="openEditModal(dev)" class="text-slate-400 hover:text-white text-2xs">{{ $t('common.edit') }}</button>
            <button @click="confirmDelete(dev)" class="text-rose-400 hover:text-rose-300 text-2xs">{{ $t('common.delete') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Device Modal (Create/Edit) -->
    <DeviceModal
      :is-open="isModalOpen"
      :device-to-edit="deviceToEdit"
      @close="isModalOpen = false"
      @saved="loadDevices"
    />
  </div>
</template>

<script>
import DeviceModal from '../components/DeviceModal.vue';

export default {
  name: 'DevicesView',
  components: {
    DeviceModal,
  },
  data() {
    return {
      devices: [],
      totalDevices: 0,
      loading: false,
      viewMode: 'table',
      searchQuery: '',
      filterStatus: '',
      filterConnStatus: '',
      filterProtocol: '',
      sortColumn: 'id',
      sortDirection: 'desc',
      currentPage: 1,
      pageSize: 15,
      isModalOpen: false,
      deviceToEdit: null,
    };
  },
  computed: {
    totalPages() {
      return Math.ceil(this.totalDevices / this.pageSize) || 1;
    },
    onlineDevicesCount() {
      return this.devices.filter(d => d.connection_status === 'ONLINE').length;
    },
    offlineDevicesCount() {
      return this.devices.filter(d => d.connection_status === 'OFFLINE' || d.connection_status === 'ERROR').length;
    },
    maintenanceDevicesCount() {
      return this.devices.filter(d => d.status === 'MAINTENANCE' || d.status === 'INACTIVE' || d.status === 'DISABLED').length;
    },
  },
  mounted() {
    this.loadDevices();
  },
  methods: {
    async loadDevices() {
      this.loading = true;
      try {
        const res = await this.$store.dispatch('fetchDevices', {
          page: this.currentPage,
          pageSize: this.pageSize,
          search: this.searchQuery,
          status: this.filterStatus,
          connectionStatus: this.filterConnStatus,
          protocol: this.filterProtocol,
          sortBy: this.sortColumn,
          sortDir: this.sortDirection,
        });

        if (res && res.items) {
          this.devices = res.items;
          this.totalDevices = res.total;
        } else if (Array.isArray(res)) {
          this.devices = res;
          this.totalDevices = res.length;
        }
      } catch (err) {
        console.error('Failed to load devices:', err);
      } finally {
        this.loading = false;
      }
    },
    onFilterChange() {
      this.currentPage = 1;
      this.loadDevices();
    },
    sortBy(column) {
      if (this.sortColumn === column) {
        this.sortDirection = this.sortDirection === 'asc' ? 'desc' : 'asc';
      } else {
        this.sortColumn = column;
        this.sortDirection = 'asc';
      }
      this.loadDevices();
    },
    changePage(newPage) {
      if (newPage >= 1 && newPage <= this.totalPages) {
        this.currentPage = newPage;
        this.loadDevices();
      }
    },
    openAddModal() {
      this.deviceToEdit = null;
      this.isModalOpen = true;
    },
    openEditModal(device) {
      this.deviceToEdit = device;
      this.isModalOpen = true;
    },
    async toggleDeviceEnabled(device) {
      try {
        await this.$store.dispatch('toggleDeviceEnabled', {
          id: device.id,
          enabled: !device.enabled,
        });
        device.enabled = !device.enabled;
      } catch (err) {
        alert('Failed to toggle device state: ' + err.message);
      }
    },
    async confirmDelete(device) {
      if (!confirm(this.$t('devices.deleteConfirm', { name: device.device_code || device.code }))) {
        return;
      }
      try {
        await this.$store.dispatch('deleteDevice', device.id);
        this.loadDevices();
      } catch (err) {
        alert('Failed to delete device: ' + err.message);
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
        day: '2-digit', month: 'short', hour: '2-digit', minute: '2-digit',
      });
    },
  },
};
</script>
