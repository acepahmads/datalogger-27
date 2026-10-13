<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-sm animate-fade-in" role="dialog" aria-modal="true" :aria-labelledby="'alarm-settings-title'">
    <div class="relative w-full max-w-lg rounded-2xl bg-white dark:bg-[#0F172A] border border-slate-200 dark:border-slate-800 shadow-2xl overflow-hidden transition-all duration-200 text-slate-800 dark:text-slate-100">
      <!-- Modal Header -->
      <div class="flex items-center justify-between px-6 py-4 border-b border-slate-100 dark:border-slate-800/80 bg-slate-50/50 dark:bg-slate-900/40">
        <div class="flex items-center space-x-3">
          <div class="w-9 h-9 rounded-xl bg-blue-600/10 dark:bg-blue-500/20 text-blue-600 dark:text-blue-400 flex items-center justify-center text-lg font-bold">
            🔔
          </div>
          <div>
            <h2 id="alarm-settings-title" class="text-sm font-bold text-slate-900 dark:text-white">
              {{ $t('alarms.notificationSettings') || 'Alarm Notification Settings' }}
            </h2>
            <p class="text-3xs text-slate-500 dark:text-slate-400">
              {{ $t('alarms.notificationSettingsDesc') || 'Configure industrial audio chimes, global banner visibility, and temporary mute.' }}
            </p>
          </div>
        </div>
        <button
          @click="close"
          class="p-1.5 rounded-lg text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition-colors"
          :title="$t('common.close') || 'Close'"
          aria-label="Close"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 space-y-5 max-h-[75vh] overflow-y-auto text-xs">
        <!-- Autoplay Warning / Unlock Banner if suspended -->
        <div
          v-if="!isAudioUnlocked"
          class="flex items-center justify-between p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-800 dark:text-amber-200"
        >
          <div class="flex items-center space-x-2.5">
            <span class="text-base">⚠️</span>
            <div class="text-3xs leading-relaxed">
              <span class="font-bold">{{ $t('alarms.audioBlockedTitle') || 'Browser Audio Blocked' }}</span>:
              {{ $t('alarms.audioBlockedDesc') || 'Browsers require operator interaction before sounds can play.' }}
            </div>
          </div>
          <button
            @click="unlockAudio"
            class="px-3 py-1.5 rounded-lg bg-amber-500 hover:bg-amber-600 text-slate-950 font-bold text-3xs shadow-sm transition-colors whitespace-nowrap ml-2"
          >
            {{ $t('alarms.enableAudioBtn') || 'Enable Sound' }}
          </button>
        </div>

        <!-- Master Toggles Grid -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
          <!-- Sound Alert Toggle -->
          <div class="p-3.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/30 flex items-center justify-between">
            <div class="space-y-0.5 pr-2">
              <div class="font-bold text-slate-900 dark:text-white flex items-center space-x-1.5">
                <span>🔊</span>
                <span>{{ $t('alarms.soundAlerts') || 'Audio Alerts' }}</span>
              </div>
              <div class="text-4xs text-slate-500 dark:text-slate-400">
                {{ $t('alarms.soundAlertsDesc') || 'Synthesize industrial chimes on active incidents' }}
              </div>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" v-model="localPrefs.soundEnabled" @change="onSoundToggle" class="sr-only peer" />
              <div class="w-10 h-5 bg-slate-300 peer-focus:outline-none rounded-full peer dark:bg-slate-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>

          <!-- Global Banner Toggle -->
          <div class="p-3.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-900/30 flex items-center justify-between">
            <div class="space-y-0.5 pr-2">
              <div class="font-bold text-slate-900 dark:text-white flex items-center space-x-1.5">
                <span>🏷️</span>
                <span>{{ $t('alarms.globalBanner') || 'Visual Banner' }}</span>
              </div>
              <div class="text-4xs text-slate-500 dark:text-slate-400">
                {{ $t('alarms.globalBannerDesc') || 'Display urgent alarm banner in top layout header' }}
              </div>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" v-model="localPrefs.bannerEnabled" class="sr-only peer" />
              <div class="w-10 h-5 bg-slate-300 peer-focus:outline-none rounded-full peer dark:bg-slate-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-slate-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-600"></div>
            </label>
          </div>
        </div>

        <!-- Volume Slider -->
        <div class="p-4 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/20 space-y-2.5" :class="{ 'opacity-50 pointer-events-none': !localPrefs.soundEnabled }">
          <div class="flex items-center justify-between">
            <span class="font-semibold text-slate-800 dark:text-slate-200 flex items-center space-x-1.5">
              <span>{{ volumeIcon }}</span>
              <span>{{ $t('alarms.alertVolume') || 'Alert Volume' }}</span>
            </span>
            <span class="font-mono text-3xs font-bold text-blue-600 dark:text-blue-400">
              {{ Math.round(localPrefs.volume * 100) }}%
            </span>
          </div>
          <input
            type="range"
            min="0"
            max="1"
            step="0.05"
            v-model.number="localPrefs.volume"
            @input="onVolumeChange"
            class="w-full h-1.5 bg-slate-200 dark:bg-slate-700 rounded-lg appearance-none cursor-pointer accent-blue-600"
          />
          <div class="flex justify-between text-4xs text-slate-400 font-mono">
            <span>0% (Silent)</span>
            <span>50%</span>
            <span>100% (Full)</span>
          </div>
        </div>

        <!-- Critical Repeat Interval & Temporary Mute -->
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
          <!-- Critical Repeat Policy -->
          <div class="space-y-1.5" :class="{ 'opacity-50 pointer-events-none': !localPrefs.soundEnabled }">
            <label class="block font-semibold text-slate-800 dark:text-slate-200 text-3xs">
              {{ $t('alarms.criticalRepeatInterval') || 'Critical Alarm Sound Repeat' }}
            </label>
            <select
              v-model.number="localPrefs.repeatIntervalSeconds"
              class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 text-xs focus:ring-2 focus:ring-blue-500 focus:outline-none"
            >
              <option :value="1">{{ $t('alarms.repeatEvery1s') || 'Every 1 second' }}</option>
              <option :value="3">{{ $t('alarms.repeatEvery3s') || 'Every 3 seconds' }}</option>
              <option :value="5">{{ $t('alarms.repeatEvery5s') || 'Every 5 seconds' }}</option>
              <option :value="10">{{ $t('alarms.repeatEvery10s') || 'Every 10 seconds' }}</option>
              <option :value="15">{{ $t('alarms.repeatEvery15s') || 'Every 15 seconds' }}</option>
              <option :value="30">{{ $t('alarms.repeatEvery30s') || 'Every 30 seconds (Default)' }}</option>
              <option :value="60">{{ $t('alarms.repeatEvery60s') || 'Every 60 seconds' }}</option>
              <option :value="0">{{ $t('alarms.repeatOnceOnly') || 'Play once only' }}</option>
            </select>
            <p class="text-4xs text-slate-400 leading-tight">
              {{ $t('alarms.repeatPolicyNote') || 'Repeats only while unacknowledged. Acknowledging stops repeating immediately.' }}
            </p>
          </div>

          <!-- Temporary Mute Duration -->
          <div class="space-y-1.5">
            <label class="block font-semibold text-slate-800 dark:text-slate-200 text-3xs">
              {{ $t('alarms.muteDuration') || 'Mute Audio Temporarily' }}
            </label>
            <select
              v-model.number="localPrefs.muteDurationMinutes"
              @change="onMuteDurationChange"
              class="w-full px-3 py-2 rounded-xl border border-slate-200 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-800 dark:text-slate-200 text-xs focus:ring-2 focus:ring-blue-500 focus:outline-none"
            >
              <option :value="0">{{ $t('alarms.unmuted') || 'Unmuted (Normal Audio)' }}</option>
              <option :value="5">{{ $t('alarms.mute5m') || 'Mute for 5 minutes' }}</option>
              <option :value="15">{{ $t('alarms.mute15m') || 'Mute for 15 minutes' }}</option>
              <option :value="30">{{ $t('alarms.mute30m') || 'Mute for 30 minutes' }}</option>
              <option :value="60">{{ $t('alarms.mute1h') || 'Mute for 1 hour' }}</option>
              <option :value="-1">{{ $t('alarms.muteIndefinite') || 'Mute indefinitely' }}</option>
            </select>
            <p v-if="localPrefs.isMuted && muteRemainingText" class="text-4xs font-mono text-amber-500 font-bold">
              🔇 {{ muteRemainingText }}
            </p>
            <p v-else class="text-4xs text-slate-400 leading-tight">
              {{ $t('alarms.muteNote') || 'Muting silences sound only; visual alarms and data ingestion remain 100% active.' }}
            </p>
          </div>
        </div>

        <!-- Sound Preview Testing -->
        <div class="p-3.5 rounded-xl border border-slate-200 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-900/20 space-y-2">
          <div class="text-3xs font-semibold text-slate-500 uppercase tracking-wider">
            {{ $t('alarms.testAudioChimes') || 'Audio Pattern Preview' }}
          </div>
          <div class="flex flex-wrap gap-2.5">
            <button
              type="button"
              @click="testSound('WARNING')"
              class="flex items-center space-x-2 px-3 py-1.5 rounded-lg border border-amber-500/30 bg-amber-500/10 hover:bg-amber-500/20 text-amber-700 dark:text-amber-300 font-medium text-3xs transition-colors"
            >
              <span>🔔</span>
              <span>{{ $t('alarms.testWarningChime') || 'Test Warning Sound' }}</span>
            </button>
            <button
              type="button"
              @click="testSound('CRITICAL')"
              class="flex items-center space-x-2 px-3 py-1.5 rounded-lg border border-rose-500/30 bg-rose-500/10 hover:bg-rose-500/20 text-rose-700 dark:text-rose-300 font-medium text-3xs transition-colors"
            >
              <span>🚨</span>
              <span>{{ $t('alarms.testCriticalAlert') || 'Test Critical Alert' }}</span>
            </button>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="flex items-center justify-between px-6 py-3.5 border-t border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-900/60">
        <button
          type="button"
          @click="resetDefaults"
          class="text-3xs text-slate-500 hover:text-slate-800 dark:hover:text-slate-200 font-semibold underline transition-colors"
        >
          {{ $t('alarms.restoreDefaults') || 'Restore Default Settings' }}
        </button>

        <div class="flex items-center space-x-2">
          <button
            type="button"
            @click="saveAndClose"
            class="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-700 text-white font-bold text-xs shadow-md shadow-blue-500/20 transition-colors"
          >
            {{ $t('common.save') || 'Save & Close' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import audioAlert from '../utils/audioAlert';
import {
  getStoredAlarmPreferences,
  saveAlarmPreferences,
  resetAlarmPreferences,
} from '../utils/alarmPreferences';

export default {
  name: 'AlarmNotificationModal',
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      localPrefs: {
        soundEnabled: true,
        bannerEnabled: true,
        volume: 0.7,
        repeatIntervalSeconds: 30,
        isMuted: false,
        muteDurationMinutes: 0,
        muteUntil: null,
      },
      isAudioUnlocked: true,
      unsubscribeAudioState: null,
      timerInterval: null,
      muteRemainingText: '',
    };
  },
  computed: {
    username() {
      return (this.$store.state.user && this.$store.state.user.username) || 'admin';
    },
    volumeIcon() {
      if (this.localPrefs.isMuted || !this.localPrefs.soundEnabled || this.localPrefs.volume === 0) return '🔇';
      if (this.localPrefs.volume < 0.4) return '🔈';
      if (this.localPrefs.volume < 0.75) return '🔉';
      return '🔊';
    },
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.loadPreferences();
        this.checkAudioState();
        this.startMuteCountdownWatcher();
      } else {
        if (this.timerInterval) {
          clearInterval(this.timerInterval);
          this.timerInterval = null;
        }
      }
    },
  },
  mounted() {
    this.loadPreferences();
    this.unsubscribeAudioState = audioAlert.onStateChange((state) => {
      this.isAudioUnlocked = state.isUnlocked;
    });
    this.checkAudioState();
  },
  beforeDestroy() {
    if (this.unsubscribeAudioState) {
      this.unsubscribeAudioState();
    }
    if (this.timerInterval) {
      clearInterval(this.timerInterval);
    }
  },
  methods: {
    loadPreferences() {
      const prefs = getStoredAlarmPreferences(this.username);
      this.localPrefs = { ...prefs };
    },
    checkAudioState() {
      const ctx = audioAlert.getAudioContext();
      this.isAudioUnlocked = ctx ? ctx.state === 'running' : true;
    },
    async unlockAudio() {
      const ok = await audioAlert.unlockAudio();
      this.isAudioUnlocked = ok;
    },
    onSoundToggle() {
      audioAlert.setSoundEnabled(this.localPrefs.soundEnabled);
    },
    onVolumeChange() {
      audioAlert.setVolume(this.localPrefs.volume);
    },
    onMuteDurationChange() {
      const dur = this.localPrefs.muteDurationMinutes;
      if (dur === 0) {
        this.localPrefs.isMuted = false;
        this.localPrefs.muteUntil = null;
      } else if (dur === -1) {
        this.localPrefs.isMuted = true;
        this.localPrefs.muteUntil = null;
      } else {
        this.localPrefs.isMuted = true;
        this.localPrefs.muteUntil = Date.now() + dur * 60 * 1000;
      }
      audioAlert.setMute(this.localPrefs.isMuted, dur > 0 ? dur : null);
      this.updateMuteRemaining();
    },
    updateMuteRemaining() {
      if (!this.localPrefs.isMuted) {
        this.muteRemainingText = '';
        return;
      }
      if (!this.localPrefs.muteUntil) {
        this.muteRemainingText = this.$t('alarms.mutedIndefinite') || 'Muted Indefinitely';
        return;
      }
      const diffMs = this.localPrefs.muteUntil - Date.now();
      if (diffMs <= 0) {
        this.localPrefs.isMuted = false;
        this.localPrefs.muteUntil = null;
        this.localPrefs.muteDurationMinutes = 0;
        this.muteRemainingText = '';
        audioAlert.setMute(false);
      } else {
        const mins = Math.ceil(diffMs / 60000);
        this.muteRemainingText = `${this.$t('alarms.mutedRemaining') || 'Muted'}: ~${mins}m remaining`;
      }
    },
    startMuteCountdownWatcher() {
      this.updateMuteRemaining();
      if (this.timerInterval) clearInterval(this.timerInterval);
      this.timerInterval = setInterval(() => {
        this.updateMuteRemaining();
      }, 5000);
    },
    testSound(severity) {
      audioAlert.previewSound(severity);
    },
    resetDefaults() {
      const defaults = resetAlarmPreferences(this.username);
      this.localPrefs = { ...defaults };
      this.$store.dispatch('updateAlarmPreferences', defaults);
    },
    saveAndClose() {
      const saved = saveAlarmPreferences(this.localPrefs, this.username);
      this.$store.dispatch('updateAlarmPreferences', saved);
      this.close();
    },
    close() {
      this.$emit('close');
    },
  },
};
</script>
