<template>
  <div class="space-y-6 font-sans">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white font-sans">
          Phase 1 Foundation — Architecture & Release Verification
        </h1>
        <p class="text-xs text-slate-400 mt-0.5 font-sans">
          Certified baseline for production local-first edge datalogger deployments with MariaDB.
        </p>
      </div>

      <span class="px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-semibold font-mono">
        PHASE 1 VERIFIED · MARIADB EDGE
      </span>
    </div>

    <!-- Phase 1 Deliverables Checklist Card -->
    <div class="saas-card p-5 space-y-4">
      <div class="border-b border-slate-800/80 pb-3">
        <h2 class="text-sm font-bold text-white tracking-tight">Phase 1 Foundation Deliverables Checklist</h2>
        <p class="text-2xs text-slate-400 mt-0.5">Single source of truth tracking completed deliverables</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
        <div v-for="(item, idx) in deliverables" :key="idx"
             class="p-3 rounded-lg bg-[#0B0F19]/70 border border-slate-800/80 flex items-start space-x-3">
          <div class="w-4 h-4 rounded-full bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 text-3xs mt-0.5 flex-shrink-0 font-bold">
            ✓
          </div>
          <div>
            <div class="font-semibold text-slate-100 font-sans text-xs">{{ item.title }}</div>
            <div class="text-2xs text-slate-400 mt-0.5 font-sans leading-relaxed">{{ item.desc }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Quick Run & Installation Instructions -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Windows Setup -->
      <div class="saas-card p-5 space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-xs font-bold text-slate-200 uppercase font-sans">Windows Local Installation</h3>
          <span class="bg-blue-500/10 text-blue-400 border border-blue-500/20 px-2 py-0.5 rounded text-3xs font-mono font-medium">PowerShell</span>
        </div>
        <p class="text-2xs text-slate-400 font-sans">Run the automated installer script to set up data directories and start the application:</p>
        <pre class="bg-[#0B0F19] p-3 rounded-lg border border-slate-800 font-mono text-3xs text-blue-300 overflow-x-auto">powershell -ExecutionPolicy Bypass -File .\scripts\install-windows.ps1</pre>
      </div>

      <!-- Linux / Raspberry Pi Setup -->
      <div class="saas-card p-5 space-y-3">
        <div class="flex items-center justify-between">
          <h3 class="text-xs font-bold text-slate-200 uppercase font-sans">Linux / Raspberry Pi OS Service</h3>
          <span class="bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 px-2 py-0.5 rounded text-3xs font-mono font-medium">systemd</span>
        </div>
        <p class="text-2xs text-slate-400 font-sans">Deploy as a self-restarting systemd background service on Raspberry Pi or Linux server:</p>
        <pre class="bg-[#0B0F19] p-3 rounded-lg border border-slate-800 font-mono text-3xs text-emerald-300 overflow-x-auto">sudo bash ./scripts/install-linux.sh</pre>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ReleaseDocView',
  data() {
    return {
      deliverables: [
        { title: '1. Modular Go Backend', desc: 'Gin framework with cmd/, internal/, pkg/ layered architecture and low memory footprint.' },
        { title: '2. Modern Vue 2 Industrial UI Shell', desc: 'Modern industrial SaaS interface with Inter typography, KPI cards, timeline, and dark theme.' },
        { title: '3. MariaDB Edge Storage Engine', desc: 'Production MariaDB storage with connection pooling, health checks, auto-migration, and low-latency queries.' },
        { title: '4. REST API Suite', desc: 'REST endpoints for phases, tasks, progress calculation, database management, hardware health, and logs.' },
        { title: '5. Authentication Foundation', desc: 'Bcrypt hashing, JWT tokens with 24h expiration, role-based authorization matrix.' },
        { title: '6. Dashboard Shell & Architecture', desc: 'End-to-end 8-stage industrial pipeline visualization and real-time status.' },
        { title: '7. 10 Development Phases Management', desc: 'Preloaded with all 10 phases from specification and full task breakdown.' },
        { title: '8. Task Management & Kanban Engine', desc: 'CRUD operations, 7 lifecycle states, priority, owner, test results, and task logs.' },
        { title: '9. Automated Progress Calculation', desc: 'Subphase, phase, and overall completion percentages computed on task status change.' },
        { title: '10. Real-time Activity Timeline', desc: 'Audited log stream of development tasks and status transitions.' },
        { title: '11. System Health Monitoring', desc: 'Real-time hardware telemetry (CPU, RAM, Disk, Uptime) and MariaDB connectivity.' },
        { title: '12. Basic Local Installer & Service', desc: 'Windows PowerShell and Linux systemd service scripts for headless operation.' },
      ],
    };
  },
};
</script>
