<template>
  <div v-if="shouldShowBanner" class="relative z-30 select-none transition-all duration-200" role="region" aria-label="Global Alarm Notification">
    <!-- Main Banner Bar -->
    <div
      class="px-4 py-2 sm:px-6 flex items-center justify-between shadow-md border-b text-xs transition-colors duration-150"
      :class="bannerColorClasses"
    >
      <!-- Left: Severity Icon, Priority Badge & Alarm Info -->
      <div class="flex items-center space-x-3 overflow-hidden flex-1 min-w-0 pr-2">
        <!-- Severity Icon & Pulsing Dot -->
        <div class="flex items-center space-x-1.5 flex-shrink-0">
          <span class="text-sm sm:text-base leading-none" :aria-label="highestAlarm.severity">
            {{ severityIcon(highestAlarm.severity) }}
          </span>
          <span
            v-if="highestAlarm.severity === 'CRITICAL' && highestAlarm.status === 'ACTIVE'"
            class="relative flex h-2 w-2"
          >
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-white opacity-80"></span>
            <span class="relative inline-flex rounded-full h-2 w-2 bg-white"></span>
          </span>
        </div>

        <!-- Severity Badge -->
        <span
          class="px-2 py-0.5 rounded-full font-mono font-bold text-3xs tracking-wider uppercase flex-shrink-0"
          :class="badgeColorClasses"
        >
          {{ highestAlarm.severity }}
        </span>

        <!-- Device & Alarm Message Content -->
        <div class="flex items-center space-x-2 truncate">
          <!-- Device Identifier Link -->
          <router-link
            v-if="highestAlarm.device_id || (highestAlarm.Device && highestAlarm.Device.id)"
            :to="`/monitoring/devices/${highestAlarm.device_id || highestAlarm.Device.id}`"
            class="font-mono font-bold underline hover:opacity-80 transition-opacity flex-shrink-0"
            :title="$t('alarms.viewDevice') || 'View Device'"
          >
            [{{ deviceLabel(highestAlarm) }}]
          </router-link>
          <span v-else class="font-mono font-bold flex-shrink-0">
            [{{ deviceLabel(highestAlarm) }}]
          </span>

          <!-- Parameter / Metric -->
          <span v-if="highestAlarm.Parameter && (highestAlarm.Parameter.name || highestAlarm.Parameter.code)" class="font-semibold hidden md:inline">
            {{ highestAlarm.Parameter.name || highestAlarm.Parameter.code }}:
          </span>

          <!-- Alarm Message -->
          <span class="truncate font-medium" :title="highestAlarm.message">
            {{ highestAlarm.message || highestAlarm.alarm_code }}
          </span>

          <!-- Timestamp Relative -->
          <span class="opacity-75 text-3xs font-mono hidden sm:inline flex-shrink-0" :title="formatFullDate(highestAlarm.triggered_at)">
            · {{ formatRelativeTime(highestAlarm.triggered_at) }}
          </span>

          <!-- Acknowledged Tag if already acked -->
          <span
            v-if="highestAlarm.status === 'ACKNOWLEDGED'"
            class="bg-black/20 text-3xs px-1.5 py-0.5 rounded font-mono font-semibold uppercase flex-shrink-0"
          >
            ✓ {{ $t('alarms.acknowledgedState') || 'Acked' }}
          </span>
        </div>

        <!-- Multi-Alarm Counter & Expand Toggle -->
        <button
          v-if="activeAlarmsCount > 1"
          @click="isExpanded = !isExpanded"
          class="flex items-center space-x-1 px-2 py-0.5 rounded-full bg-black/20 hover:bg-black/30 text-3xs font-mono font-bold transition-colors flex-shrink-0 ml-1"
          :title="$t('alarms.toggleAlarmList') || 'Toggle active alarms list'"
        >
          <span>+{{ activeAlarmsCount - 1 }} {{ $t('alarms.more') || 'more' }}</span>
          <svg class="w-3 h-3 transition-transform duration-200" :class="{ 'rotate-180': isExpanded }" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>
      </div>

      <!-- Right: Action Controls -->
      <div class="flex items-center space-x-1.5 sm:space-x-2 flex-shrink-0">
        <!-- Autoplay Unlock Prompt if browser suspended audio -->
        <button
          v-if="!isAudioUnlocked && prefs.soundEnabled"
          @click="unlockAudio"
          class="flex items-center space-x-1 px-2 py-1 rounded-lg bg-amber-400 hover:bg-amber-300 text-slate-950 font-bold text-3xs shadow-sm transition-colors animate-pulse"
          :title="$t('alarms.audioBlockedTooltip') || 'Click to enable audio chimes in this browser'"
        >
          <span>🔇</span>
          <span class="hidden lg:inline">{{ $t('alarms.enableAudioBtn') || 'Enable Sound' }}</span>
        </button>

        <!-- Quick Acknowledge Button -->
        <button
          v-if="highestAlarm.status === 'ACTIVE'"
          @click="openAckModal(highestAlarm)"
          class="px-2.5 py-1 rounded-lg bg-white/20 hover:bg-white/30 text-3xs font-bold font-mono tracking-wide transition-colors flex items-center space-x-1 shadow-sm"
          :title="$t('alarms.acknowledgeTooltip') || 'Acknowledge this alarm incident'"
        >
          <span>✓</span>
          <span class="hidden sm:inline">{{ $t('alarms.acknowledge') || 'Acknowledge' }}</span>
        </button>

        <!-- View Alarm Console Link -->
        <router-link
          to="/alarm/active"
          class="px-2 py-1 rounded-lg bg-black/15 hover:bg-black/25 text-3xs font-medium transition-colors hidden sm:flex items-center space-x-1"
          :title="$t('alarms.viewAlarmConsole') || 'Go to Alarm Console'"
        >
          <span>📋</span>
          <span class="hidden md:inline">{{ $t('alarms.console') || 'Console' }}</span>
        </router-link>

        <!-- Mute / Unmute Toggle Button -->
        <button
          @click="toggleMute"
          class="p-1.5 rounded-lg hover:bg-black/20 text-3xs transition-colors flex items-center justify-center"
          :title="isMuted ? ($t('alarms.unmute') || 'Unmute Sound') : ($t('alarms.mute') || 'Mute Sound')"
          :aria-label="isMuted ? 'Unmute' : 'Mute'"
        >
          <span class="text-xs">{{ isMuted ? '🔇' : '🔊' }}</span>
        </button>

        <!-- Notification Settings (Gear) Button -->
        <button
          @click="isSettingsModalOpen = true"
          class="p-1.5 rounded-lg hover:bg-black/20 text-3xs transition-colors flex items-center justify-center"
          :title="$t('alarms.notificationSettings') || 'Notification Settings'"
          aria-label="Notification Settings"
        >
          <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
          </svg>
        </button>

        <!-- Dismiss Banner Button -->
        <button
          @click="dismissHighestAlarm"
          class="p-1 rounded-lg hover:bg-black/25 text-3xs transition-colors flex items-center justify-center font-bold"
          :title="$t('alarms.dismissBanner') || 'Dismiss this banner (keeps alarm in console)'"
          aria-label="Dismiss banner"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Expandable List of All Active Incidents -->
    <div
      v-if="isExpanded && activeAlarmsCount > 1"
      class="border-b border-slate-200 dark:border-slate-800 bg-white/95 dark:bg-[#0B0F19]/95 backdrop-blur-md shadow-lg max-h-60 overflow-y-auto divide-y divide-slate-100 dark:divide-slate-800/60 text-xs"
    >
      <div
        v-for="item in activeAlarmsList"
        :key="item.id"
        class="px-5 py-2.5 flex items-center justify-between hover:bg-slate-50 dark:hover:bg-slate-800/40 transition-colors"
        :class="{ 'opacity-60': item.status === 'ACKNOWLEDGED' }"
      >
        <div class="flex items-center space-x-3 truncate flex-1 min-w-0 pr-3">
          <span class="text-xs">{{ severityIcon(item.severity) }}</span>
          <span
            class="px-1.5 py-0.5 rounded text-4xs font-mono font-bold uppercase"
            :class="itemSeverityBadgeClass(item.severity)"
          >
            {{ item.severity }}
          </span>
          <span class="font-mono font-bold text-slate-800 dark:text-slate-200 flex-shrink-0">
            [{{ deviceLabel(item) }}]
          </span>
          <span class="truncate text-slate-700 dark:text-slate-300">
            {{ item.message || item.alarm_code }}
          </span>
          <span class="text-4xs text-slate-400 font-mono flex-shrink-0">
            {{ formatRelativeTime(item.triggered_at) }}
          </span>
        </div>

        <div class="flex items-center space-x-2 flex-shrink-0">
          <span v-if="item.status === 'ACKNOWLEDGED'" class="text-3xs text-emerald-600 dark:text-emerald-400 font-mono">
            ✓ {{ item.acknowledged_by || 'Acked' }}
          </span>
          <button
            v-else
            @click="openAckModal(item)"
            class="px-2 py-0.5 rounded bg-blue-50 hover:bg-blue-100 dark:bg-blue-900/30 dark:hover:bg-blue-900/50 text-blue-600 dark:text-blue-300 text-3xs font-semibold font-mono transition-colors"
          >
            {{ $t('alarms.acknowledge') || 'Acknowledge' }}
          </button>
          <button
            @click="dismissSpecificAlarm(item.id)"
            class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-3xs px-1"
            :title="$t('alarms.dismissBanner') || 'Dismiss'"
          >
            ✕
          </button>
        </div>
      </div>
    </div>

    <!-- Quick Operator Acknowledgement Modal -->
    <div
      v-if="showQuickAckModal && ackTargetAlarm"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
    >
      <div class="relative w-full max-w-md rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 shadow-2xl p-6 text-slate-800 dark:text-slate-100 space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <div class="flex items-center space-x-2">
            <span class="text-base">🛡️</span>
            <h3 class="text-sm font-bold text-slate-900 dark:text-white">
              {{ $t('alarms.ackModalTitle') || 'Acknowledge Alarm Incident' }}
            </h3>
          </div>
          <button @click="showQuickAckModal = false" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200">✕</button>
        </div>

        <div class="p-3 rounded-xl bg-slate-50 dark:bg-slate-900/40 border border-slate-200 dark:border-slate-800 space-y-1 text-xs">
          <div class="flex items-center justify-between font-mono">
            <span class="font-bold text-slate-900 dark:text-white">[{{ deviceLabel(ackTargetAlarm) }}]</span>
            <span class="font-bold" :class="ackTargetAlarm.severity === 'CRITICAL' ? 'text-rose-500' : 'text-amber-500'">
              {{ ackTargetAlarm.severity }}
            </span>
          </div>
          <div class="text-3xs text-slate-600 dark:text-slate-300 font-sans">
            {{ ackTargetAlarm.message || ackTargetAlarm.alarm_code }}
          </div>
        </div>

        <div class="space-y-1.5">
          <label class="block text-3xs font-semibold text-slate-700 dark:text-slate-300">
            {{ $t('alarms.operatorNotes') || 'Operator Notes (Optional)' }}
          </label>
          <textarea
            v-model="quickAckNotes"
            rows="2"
            class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 text-xs focus:ring-2 focus:ring-blue-500 focus:outline-none"
            :placeholder="$t('alarms.notesPlaceholder') || 'Field observations, actions taken...'"
          ></textarea>
        </div>

        <div class="flex items-center justify-end space-x-2 pt-2 border-t border-slate-100 dark:border-slate-800">
          <button
            type="button"
            @click="showQuickAckModal = false"
            class="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-3xs font-semibold text-slate-600 dark:text-slate-400 hover:bg-slate-100 dark:hover:bg-slate-800"
          >
            {{ $t('common.cancel') || 'Cancel' }}
          </button>
          <button
            type="button"
            @click="confirmQuickAck"
            :disabled="isAckSubmitting"
            class="px-4 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-bold text-3xs shadow-md transition-colors"
          >
            {{ isAckSubmitting ? 'Acknowledging...' : ($t('alarms.confirmAck') || 'Confirm Acknowledgement') }}
          </button>
        </div>
      </div>
    </div>

    <!-- Notification Settings Modal -->
    <AlarmNotificationModal
      :is-open="isSettingsModalOpen"
      @close="isSettingsModalOpen = false"
    />
  </div>
</template>

<script>
import audioAlert from '../utils/audioAlert';
import { getStoredAlarmPreferences, applyAlarmPreferences } from '../utils/alarmPreferences';
import AlarmNotificationModal from './AlarmNotificationModal.vue';

export default {
  name: 'GlobalAlarmBanner',
  components: {
    AlarmNotificationModal,
  },
  data() {
    return {
      isExpanded: false,
      isSettingsModalOpen: false,
      dismissedMap: {},
      isAudioUnlocked: true,
      unsubscribeAudioState: null,
      showQuickAckModal: false,
      ackTargetAlarm: null,
      quickAckNotes: '',
      isAckSubmitting: false,
      reconcileInterval: null,
    };
  },
  computed: {
    username() {
      return (this.$store.state.user && this.$store.state.user.username) || 'admin';
    },
    prefs() {
      return this.$store.state.alarmPreferences || getStoredAlarmPreferences(this.username);
    },
    isMuted() {
      return this.prefs.isMuted;
    },
    // Raw alarms from store or API
    activeAlarms() {
      return this.$store.state.activeAlarms || [];
    },
    // Filter out dismissed or cleared alarms for this session
    filteredAlarms() {
      return this.activeAlarms.filter(
        (a) => a && a.status !== 'CLEARED' && !this.dismissedMap[Number(a.id)]
      );
    },
    // Prioritize: CRITICAL (3) > WARNING (2) > INFO (1), then newest triggered_at
    sortedAlarms() {
      const severityScore = { CRITICAL: 3, WARNING: 2, INFO: 1 };
      return [...this.filteredAlarms].sort((a, b) => {
        const scoreA = severityScore[a.severity] || 0;
        const scoreB = severityScore[b.severity] || 0;
        if (scoreA !== scoreB) return scoreB - scoreA;
        const timeA = new Date(a.triggered_at || 0).getTime();
        const timeB = new Date(b.triggered_at || 0).getTime();
        return timeB - timeA;
      });
    },
    highestAlarm() {
      return this.sortedAlarms.length > 0 ? this.sortedAlarms[0] : null;
    },
    activeAlarmsCount() {
      return this.sortedAlarms.length;
    },
    activeAlarmsList() {
      return this.sortedAlarms;
    },
    shouldShowBanner() {
      return (
        this.prefs.bannerEnabled !== false &&
        this.highestAlarm !== null
      );
    },
    bannerColorClasses() {
      if (!this.highestAlarm) return '';
      switch (this.highestAlarm.severity) {
        case 'CRITICAL':
          return 'bg-gradient-to-r from-rose-600 to-red-700 text-white border-rose-800 dark:from-rose-950/95 dark:to-red-950/90 dark:text-rose-100 dark:border-rose-800/80';
        case 'WARNING':
          return 'bg-gradient-to-r from-amber-500 to-amber-600 text-slate-950 border-amber-600 dark:from-amber-950/95 dark:to-orange-950/90 dark:text-amber-100 dark:border-amber-800/80';
        case 'INFO':
        default:
          return 'bg-gradient-to-r from-blue-600 to-indigo-700 text-white border-blue-800 dark:from-blue-950/95 dark:to-indigo-950/90 dark:text-blue-100 dark:border-blue-800/80';
      }
    },
    badgeColorClasses() {
      if (!this.highestAlarm) return '';
      switch (this.highestAlarm.severity) {
        case 'CRITICAL':
          return 'bg-rose-900/60 text-white border border-rose-300/40 dark:bg-rose-500/20 dark:text-rose-300 dark:border-rose-500/40';
        case 'WARNING':
          return 'bg-amber-900/40 text-slate-950 border border-amber-700/40 dark:bg-amber-500/20 dark:text-amber-300 dark:border-amber-500/40';
        case 'INFO':
        default:
          return 'bg-blue-900/60 text-white border border-blue-300/40 dark:bg-blue-500/20 dark:text-blue-300 dark:border-blue-500/40';
      }
    },
  },
  watch: {
    // Watch highest alarm changes to trigger audio alerts
    highestAlarm: {
      handler(newAlarm, oldAlarm) {
        if (!newAlarm) {
          audioAlert.stopAllLoops();
          return;
        }

        // If new or different alarm became highest
        const isNewIncident = !oldAlarm || oldAlarm.id !== newAlarm.id;
        if (isNewIncident) {
          if (newAlarm.severity === 'CRITICAL' && newAlarm.status === 'ACTIVE') {
            audioAlert.startCriticalAlarmLoop(newAlarm.id);
          } else if (newAlarm.severity === 'WARNING' && newAlarm.status === 'ACTIVE') {
            audioAlert.playWarningAlert();
          }
        } else if (newAlarm.status === 'ACKNOWLEDGED') {
          // If status transitioned to acknowledged, stop critical audio loop
          audioAlert.stopCriticalAlarmLoop(newAlarm.id);
        }
      },
      deep: true,
    },
  },
  mounted() {
    // Initialize preferences and apply to audio engine
    const initialPrefs = getStoredAlarmPreferences(this.username);
    this.$store.dispatch('updateAlarmPreferences', initialPrefs);
    applyAlarmPreferences(initialPrefs);

    // Initial fetch of active alarms
    this.$store.dispatch('fetchActiveAlarms');

    // Subscribe to audio engine state changes
    this.unsubscribeAudioState = audioAlert.onStateChange((state) => {
      this.isAudioUnlocked = state.isUnlocked;
    });

    // Reconcile periodically (every 30 seconds) in case of missed WebSocket frames
    this.reconcileInterval = setInterval(() => {
      this.$store.dispatch('fetchActiveAlarms');
    }, 30000);

    // Global click listener to unlock Web Audio on first operator interaction
    window.addEventListener('click', this.handleUserInteractionUnlock, { once: true });
  },
  beforeDestroy() {
    if (this.unsubscribeAudioState) {
      this.unsubscribeAudioState();
    }
    if (this.reconcileInterval) {
      clearInterval(this.reconcileInterval);
    }
    audioAlert.stopAllLoops();
    window.removeEventListener('click', this.handleUserInteractionUnlock);
  },
  methods: {
    severityIcon(severity) {
      switch (severity) {
        case 'CRITICAL': return '🚨';
        case 'WARNING': return '⚠️';
        case 'INFO':
        default: return 'ℹ️';
      }
    },
    itemSeverityBadgeClass(severity) {
      switch (severity) {
        case 'CRITICAL': return 'bg-rose-500/20 text-rose-600 dark:text-rose-400 border border-rose-500/30';
        case 'WARNING': return 'bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/30';
        case 'INFO':
        default: return 'bg-blue-500/20 text-blue-600 dark:text-blue-400 border border-blue-500/30';
      }
    },
    deviceLabel(alarm) {
      if (!alarm) return 'System';
      if (alarm.Device && (alarm.Device.device_code || alarm.Device.name)) {
        return alarm.Device.device_code || alarm.Device.name;
      }
      if (alarm.device_id) return `Device #${alarm.device_id}`;
      return 'System';
    },
    formatFullDate(dt) {
      if (!dt) return '';
      try {
        return new Date(dt).toLocaleString();
      } catch (e) {
        return String(dt);
      }
    },
    formatRelativeTime(dt) {
      if (!dt) return '';
      try {
        const sec = Math.floor((Date.now() - new Date(dt).getTime()) / 1000);
        if (sec < 5) return this.$t('alarms.justNow') || 'just now';
        if (sec < 60) return `${sec}s ago`;
        const min = Math.floor(sec / 60);
        if (min < 60) return `${min}m ago`;
        const hrs = Math.floor(min / 60);
        if (hrs < 24) return `${hrs}h ago`;
        return `${Math.floor(hrs / 24)}d ago`;
      } catch (e) {
        return '';
      }
    },
    async unlockAudio() {
      const ok = await audioAlert.unlockAudio();
      this.isAudioUnlocked = ok;
    },
    handleUserInteractionUnlock() {
      if (!this.isAudioUnlocked) {
        audioAlert.unlockAudio().then((ok) => {
          this.isAudioUnlocked = ok;
        });
      }
    },
    toggleMute() {
      const newMuted = audioAlert.toggleMute();
      this.$store.dispatch('updateAlarmPreferences', {
        ...this.prefs,
        isMuted: newMuted,
        muteUntil: audioAlert.muteUntil,
      });
    },
    dismissHighestAlarm() {
      if (this.highestAlarm && this.highestAlarm.id) {
        this.dismissSpecificAlarm(this.highestAlarm.id);
      }
    },
    dismissSpecificAlarm(alarmId) {
      if (!alarmId) return;
      const idNum = Number(alarmId);
      this.$set(this.dismissedMap, idNum, true);
      audioAlert.stopCriticalAlarmLoop(idNum);
    },
    openAckModal(alarm) {
      this.ackTargetAlarm = alarm;
      this.quickAckNotes = '';
      this.showQuickAckModal = true;
    },
    async confirmQuickAck() {
      if (!this.ackTargetAlarm) return;
      this.isAckSubmitting = true;
      try {
        await this.$store.dispatch('acknowledgeAlarm', {
          id: this.ackTargetAlarm.id,
          notes: this.quickAckNotes,
        });
        audioAlert.stopCriticalAlarmLoop(this.ackTargetAlarm.id);
        this.showQuickAckModal = false;
        this.ackTargetAlarm = null;
        // Refresh active alarms
        await this.$store.dispatch('fetchActiveAlarms');
      } catch (err) {
        console.error('Failed to acknowledge alarm:', err);
      } finally {
        this.isAckSubmitting = false;
      }
    },
  },
};
</script>
