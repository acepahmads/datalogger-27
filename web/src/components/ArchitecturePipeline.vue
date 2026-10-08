<template>
  <div class="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-2.5 font-sans">
    <div v-for="(node, idx) in pipelineNodes" :key="node.id"
         class="p-2.5 rounded-lg bg-[#0B0F19]/70 border border-slate-800/80 hover:border-slate-700/80 transition-all flex flex-col justify-between">
      <div>
        <div class="flex items-center justify-between mb-1.5">
          <span class="text-3xs font-mono text-slate-500">0{{ idx + 1 }}</span>
          <span class="w-1.5 h-1.5 rounded-full"
                :class="node.status === 'ONLINE' ? 'bg-emerald-400' : 'bg-amber-400'"></span>
        </div>
        <div class="text-xs font-bold text-slate-100 font-sans tracking-tight">{{ node.name }}</div>
        <div class="text-3xs text-slate-400 mt-0.5 truncate font-sans">{{ node.subtext }}</div>
      </div>

      <div class="mt-3 pt-2 border-t border-slate-800/60 text-3xs font-sans space-y-0.5">
        <div class="flex justify-between">
          <span class="text-slate-500">{{ $t('common.health') }}:</span>
          <span class="text-emerald-400 font-semibold font-mono">{{ node.health }}</span>
        </div>
        <div class="flex justify-between">
          <span class="text-slate-500">{{ $t('common.activity') }}:</span>
          <span class="text-slate-300 font-mono">{{ node.lastActivity }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ArchitecturePipeline',
  data() {
    return {
      pipelineNodes: [
        { id: 1, name: 'DEVICE', subtext: 'Modbus / Serial', status: 'ONLINE', health: '100%', lastActivity: '1s ago' },
        { id: 2, name: 'COMM', subtext: 'TCP/IP & RS485', status: 'ONLINE', health: '99.9%', lastActivity: '1s ago' },
        { id: 3, name: 'ACQUIRE', subtext: 'Polling Engine', status: 'ONLINE', health: '100%', lastActivity: '1s ago' },
        { id: 4, name: 'PROCESS', subtext: 'Formula & Scale', status: 'ONLINE', health: '100%', lastActivity: '1s ago' },
        { id: 5, name: 'STORAGE', subtext: 'MariaDB Edge', status: 'ONLINE', health: '100%', lastActivity: '1s ago' },
        { id: 6, name: 'ALARM', subtext: 'Threshold Engine', status: 'ONLINE', health: '100%', lastActivity: '5s ago' },
        { id: 7, name: 'OUTPUT', subtext: 'Bounded Queue', status: 'ONLINE', health: 'OPTIMAL', lastActivity: 'Idle' },
        { id: 8, name: 'WEB APP', subtext: 'Vue 2 SaaS', status: 'ONLINE', health: '100%', lastActivity: 'Now' },
      ],
    };
  },
};
</script>
