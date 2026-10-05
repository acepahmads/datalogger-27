<template>
  <div class="space-y-6 font-sans">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans">
          Devices & Sensors
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          Industrial equipment telemetry, connection health, and live channel values.
        </p>
      </div>

      <span class="text-2xs font-mono text-slate-400 px-3 py-1 rounded-lg bg-[#111827] border border-slate-800">
        Total Devices: {{ devices.length }}
      </span>
    </div>

    <!-- Device Cards Grid -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div v-for="dev in devices" :key="dev.id"
           class="saas-card p-5 space-y-4">
        <!-- Device Card Header -->
        <div class="flex items-start justify-between border-b border-slate-800/80 pb-3">
          <div>
            <div class="flex items-center space-x-2">
              <span class="px-2 py-0.5 rounded-md bg-blue-500/10 text-blue-400 border border-blue-500/20 font-mono text-3xs font-semibold">
                {{ dev.code }}
              </span>
              <span class="flex items-center text-3xs font-mono font-semibold" :class="dev.status === 'ONLINE' ? 'text-emerald-400' : 'text-slate-400'">
                <span class="w-1.5 h-1.5 rounded-full mr-1.5" :class="dev.status === 'ONLINE' ? 'bg-emerald-400 animate-pulse' : 'bg-slate-500'"></span>
                {{ dev.status }}
              </span>
            </div>
            <h3 class="text-sm font-bold text-white mt-1.5 font-sans">{{ dev.name }}</h3>
            <div class="text-2xs text-slate-400 mt-0.5 font-sans">{{ dev.location }}</div>
          </div>
        </div>

        <!-- Connection Details -->
        <div class="bg-[#0B0F19]/80 p-3 rounded-xl border border-slate-800/80 text-xs font-sans space-y-1.5">
          <div class="flex justify-between">
            <span class="text-slate-400">Protocol:</span>
            <span class="text-blue-300 font-semibold font-mono text-2xs">{{ dev.device_type ? dev.device_type.name : 'MODBUS' }}</span>
          </div>
          <div class="flex justify-between" v-if="dev.connection">
            <span class="text-slate-400">Address:</span>
            <span class="font-mono text-slate-200 text-2xs">{{ dev.connection.address }}{{ dev.connection.port ? ':' + dev.connection.port : '' }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Latency:</span>
            <span class="font-mono text-emerald-400 font-semibold text-2xs">{{ dev.latency_ms }} ms</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Packets:</span>
            <span class="font-mono text-slate-300 text-2xs">{{ dev.success_count }} ok / {{ dev.failed_count }} err</span>
          </div>
        </div>

        <!-- Parameter Telemetry List -->
        <div class="space-y-2">
          <div class="text-3xs uppercase tracking-wider text-slate-400 font-semibold">Live Parameters</div>
          <div class="space-y-1.5">
            <div v-for="param in dev.parameters" :key="param.id"
                 class="p-2.5 rounded-lg bg-[#0B0F19]/60 border border-slate-800/80 flex items-center justify-between text-xs">
              <div>
                <span class="text-slate-200 font-medium font-sans text-xs">{{ param.name }}</span>
                <span class="text-slate-500 block font-mono text-3xs">#{{ param.register_address }}</span>
              </div>
              <div class="text-right">
                <span class="text-xs font-bold text-blue-400 font-mono tabular-nums">
                  {{ param.current_value !== null ? param.current_value.toFixed(1) : '--' }}
                </span>
                <span class="text-slate-400 text-3xs ml-0.5 font-sans">{{ param.unit }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'DevicesView',
  computed: {
    devices() {
      return this.$store.state.devices;
    },
  },
  mounted() {
    this.$store.dispatch('fetchDevices');
  },
};
</script>
