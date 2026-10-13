/**
 * Industrial Audio Alert Engine — Pure Web Audio API Synthesis
 * Datalogger Analysis Application
 *
 * Capabilities:
 * - Professional industrial tones (synthesized sine/triangle waves, zero external assets)
 * - Distinct acoustic patterns for WARNING (single soft dual-tone) and CRITICAL (urgent dual-pulse)
 * - Controlled repeat policy for unacknowledged critical alarms
 * - Autoplay restriction detection & user interaction unlock handler
 * - Volume control and temporary mute duration support
 * - Zero memory leaks: oscillators are scheduled and garbage collected cleanly
 */

class AudioAlertEngine {
  constructor() {
    this.audioCtx = null;
    this.masterGain = null;
    this.isMuted = false;
    this.muteUntil = null;
    this.volume = 0.7; // Default 70% volume
    this.soundEnabled = true;
    this.repeatIntervalSeconds = 5; // Default 5s repeat for unacknowledged alarms

    // Track active unacknowledged alarms: Map<alarmId, { severity, lastUpdated }>
    this.activeAlarmMap = new Map();
    this.activeCriticalAlarmIds = new Set();
    this.currentLoopSeverity = null;
    this.masterAlarmTimer = null;
    this.masterCriticalTimer = null;
    this.lastWarningTime = 0;
    this.lastCriticalTime = 0;

    // Pulse event listeners
    this.onPulseCallbacks = new Set();

    // Track autoplay availability
    this.isUnlocked = false;
    this.onStateChangeCallbacks = new Set();
  }

  /**
   * Initializes or returns the existing AudioContext
   */
  getAudioContext() {
    if (typeof window === 'undefined') return null;

    if (!this.audioCtx) {
      const AudioContextClass = window.AudioContext || window.webkitAudioContext;
      if (!AudioContextClass) {
        console.warn('[AudioAlert] Web Audio API is not supported in this browser.');
        return null;
      }
      this.audioCtx = new AudioContextClass();

      // Master Gain Node for volume and mute
      this.masterGain = this.audioCtx.createGain();
      this.masterGain.gain.setValueAtTime(this.effectiveGainValue(), this.audioCtx.currentTime);
      this.masterGain.connect(this.audioCtx.destination);

      // Listen for state changes (e.g. suspended -> running)
      if (this.audioCtx.onstatechange !== undefined) {
        this.audioCtx.onstatechange = () => {
          this.notifyStateChange();
        };
      }
    }

    if (this.audioCtx.state === 'running') {
      this.isUnlocked = true;
    }

    return this.audioCtx;
  }

  /**
   * Calculate effective master gain factoring volume and mute state
   */
  effectiveGainValue() {
    if (!this.soundEnabled || this.checkMuteActive()) {
      return 0;
    }
    return Math.max(0, Math.min(1, this.volume));
  }

  /**
   * Update master gain immediately
   */
  updateMasterGain() {
    if (this.audioCtx && this.masterGain) {
      const targetGain = this.effectiveGainValue();
      const now = this.audioCtx.currentTime;
      this.masterGain.gain.cancelScheduledValues(now);
      this.masterGain.gain.setTargetAtTime(targetGain, now, 0.02);
    }
    this.notifyStateChange();
  }

  /**
   * Check if mute is currently active (including temporary mute expiration)
   */
  checkMuteActive() {
    if (this.isMuted) {
      if (this.muteUntil) {
        const now = Date.now();
        if (now > this.muteUntil) {
          // Mute period expired
          this.isMuted = false;
          this.muteUntil = null;
          return false;
        }
      }
      return true;
    }
    return false;
  }

  /**
   * Explicitly unlock / resume AudioContext after user interaction
   */
  async unlockAudio() {
    const ctx = this.getAudioContext();
    if (!ctx) return false;

    if (ctx.state === 'suspended') {
      try {
        await ctx.resume();
        this.isUnlocked = ctx.state === 'running';
        this.notifyStateChange();
        // If there are active alarms looping, trigger a pulse now that audio is unlocked
        if (this.isUnlocked && this.activeAlarmMap.size > 0 && this.currentLoopSeverity) {
          this.triggerPulse(this.currentLoopSeverity);
        }
        return this.isUnlocked;
      } catch (err) {
        console.warn('[AudioAlert] Failed to resume AudioContext:', err);
        return false;
      }
    }
    this.isUnlocked = true;
    this.notifyStateChange();
    return true;
  }

  /**
   * Check if audio is currently playable
   */
  canPlay() {
    if (!this.soundEnabled || this.checkMuteActive()) return false;
    const ctx = this.getAudioContext();
    return ctx !== null && ctx.state === 'running';
  }

  /**
   * Plays a synthesized tone burst with clean attack and decay
   */
  playTone(frequency, startTime, duration, type = 'sine', peakGain = 0.5) {
    const ctx = this.getAudioContext();
    if (!ctx || !this.masterGain) return;

    try {
      const osc = ctx.createOscillator();
      const gainNode = ctx.createGain();

      osc.type = type;
      osc.frequency.setValueAtTime(frequency, startTime);

      // Smooth attack (5ms) to prevent speaker click
      gainNode.gain.setValueAtTime(0.0001, startTime);
      gainNode.gain.exponentialRampToValueAtTime(peakGain, startTime + 0.008);

      // Smooth exponential decay
      gainNode.gain.exponentialRampToValueAtTime(0.0001, startTime + duration);

      osc.connect(gainNode);
      gainNode.connect(this.masterGain);

      osc.start(startTime);
      osc.stop(startTime + duration + 0.05);

      // Clean up nodes after completion
      osc.onended = () => {
        try {
          osc.disconnect();
          gainNode.disconnect();
        } catch (e) {}
      };
    } catch (e) {
      console.warn('[AudioAlert] Error creating oscillator:', e);
    }
  }

  /**
   * Play WARNING sound: Soft industrial dual-tone chime (D5 -> A5)
   * Enforces Single Audio Stream: debounced so multiple concurrent warning alarms sound only once
   */
  playWarningAlert(force = false) {
    if (!this.soundEnabled || this.checkMuteActive()) return;

    // SCADA Priority Rule: If any critical alarm is actively repeating and not forced, suppress
    if (!force && this.currentLoopSeverity === 'CRITICAL') {
      return;
    }

    // Debounce warning alert so an alarm storm (multiple alarms at once) plays only 1 chime
    const now = Date.now();
    if (!force && (now - this.lastWarningTime < 3000)) {
      return;
    }
    this.lastWarningTime = now;

    const ctx = this.getAudioContext();
    if (!ctx) return;

    // If context is suspended, attempt resume silently if triggered by user event
    if (ctx.state === 'suspended') {
      ctx.resume().catch(() => {});
      return;
    }

    const t0 = ctx.currentTime + 0.02;
    // Tone 1: 587.33 Hz (D5), 110ms
    this.playTone(587.33, t0, 0.11, 'sine', 0.4);
    // Tone 2: 880.00 Hz (A5), 140ms
    this.playTone(880.00, t0 + 0.11, 0.14, 'triangle', 0.35);
  }

  /**
   * Play CRITICAL sound: Urgent dual-pulse industrial alert (A5 + C6)
   * High attention-commanding acoustic signature
   */
  playCriticalAlert(force = false) {
    if (!this.soundEnabled || this.checkMuteActive()) return;

    // Throttle successive critical alert chimes to prevent sound overlap/cacophony
    const now = Date.now();
    if (!force && (now - this.lastCriticalTime < 600)) {
      return;
    }
    this.lastCriticalTime = now;

    const ctx = this.getAudioContext();
    if (!ctx) return;

    if (ctx.state === 'suspended') {
      ctx.resume().catch(() => {});
      return;
    }

    const t0 = ctx.currentTime + 0.02;

    // Pulse 1: 880 Hz + overtone 1046.5 Hz, 90ms
    this.playTone(880.00, t0, 0.09, 'sine', 0.5);
    this.playTone(1046.50, t0, 0.09, 'triangle', 0.3);

    // 60ms gap

    // Pulse 2: 1046.50 Hz + overtone 1318.5 Hz, 120ms
    this.playTone(1046.50, t0 + 0.15, 0.12, 'sine', 0.55);
    this.playTone(1318.51, t0 + 0.15, 0.12, 'triangle', 0.35);
  }

  /**
   * Executes a synchronized alarm pulse (audio tone + visual event).
   * Scheduled at the exact same millisecond. Even if audio is muted or volume is 0,
   * the visual pulse still fires to maintain 100% synchronized cadence.
   */
  triggerPulse(severity = 'WARNING') {
    const normSeverity = (severity || 'WARNING').toUpperCase();

    // 1. Synthesize audio chime if audio is enabled and unmuted
    if (normSeverity === 'CRITICAL') {
      this.playCriticalAlert(true);
    } else {
      this.playWarningAlert(true);
    }

    // 2. Broadcast synchronized visual pulse event
    const payload = {
      severity: normSeverity,
      timestamp: Date.now(),
    };

    for (const cb of this.onPulseCallbacks) {
      try { cb(payload); } catch (e) {}
    }

    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('scada-alarm-pulse', { detail: payload }));
    }
  }

  /**
   * Restart master timer at the specified severity (CRITICAL or WARNING).
   */
  restartMasterTimer(severity) {
    if (this.masterAlarmTimer) {
      clearInterval(this.masterAlarmTimer);
      this.masterAlarmTimer = null;
      this.masterCriticalTimer = null;
    }

    this.currentLoopSeverity = severity;

    // Immediate pulse on first occurrence or escalation
    this.triggerPulse(severity);

    // If repeat is disabled (interval <= 0), do not schedule recurring timer
    if (this.repeatIntervalSeconds <= 0) {
      return;
    }

    const intervalMs = Math.max(1000, this.repeatIntervalSeconds * 1000);
    this.masterAlarmTimer = setInterval(() => {
      if (this.activeAlarmMap.size > 0 && this.currentLoopSeverity) {
        this.triggerPulse(this.currentLoopSeverity);
      } else {
        this.stopAllLoops();
      }
    }, intervalMs);
    this.masterCriticalTimer = this.masterAlarmTimer;
  }

  /**
   * Synchronize the entire active unacknowledged alarm list from Vuex or API.
   * Runs single master annunciator loop with synchronized audio & visual pulse.
   */
  syncActiveAlarms(alarms = []) {
    if (!Array.isArray(alarms) || alarms.length === 0) {
      if (this.activeAlarmMap.size > 0) {
        this.stopAllLoops();
      }
      return;
    }

    const unacked = alarms.filter((a) => a && (a.status === 'ACTIVE' || !a.status));
    if (unacked.length === 0) {
      if (this.activeAlarmMap.size > 0) {
        this.stopAllLoops();
      }
      return;
    }

    const currentIds = new Set();
    let hasCritical = false;

    for (const a of unacked) {
      const id = Number(a.id);
      currentIds.add(id);
      const sev = (a.severity || 'WARNING').toUpperCase();
      this.activeAlarmMap.set(id, { severity: sev, lastUpdated: Date.now() });
      if (sev === 'CRITICAL') {
        hasCritical = true;
      }
    }

    // Prune alarms that are no longer unacknowledged
    for (const id of this.activeAlarmMap.keys()) {
      if (!currentIds.has(id)) {
        this.activeAlarmMap.delete(id);
      }
    }

    if (this.activeAlarmMap.size === 0) {
      this.stopAllLoops();
      return;
    }

    const targetSeverity = hasCritical ? 'CRITICAL' : 'WARNING';
    if (!this.masterAlarmTimer || this.currentLoopSeverity !== targetSeverity) {
      this.restartMasterTimer(targetSeverity);
    }
  }

  /**
   * Start or register an active unacknowledged alarm incident.
   */
  startAlarmLoop(alarmId, severity = 'WARNING') {
    if (!alarmId) return;
    const id = Number(alarmId);
    const sev = (severity || 'WARNING').toUpperCase();
    this.activeAlarmMap.set(id, { severity: sev, lastUpdated: Date.now() });

    let hasCritical = false;
    for (const item of this.activeAlarmMap.values()) {
      if (item.severity === 'CRITICAL') {
        hasCritical = true;
        break;
      }
    }
    const targetSeverity = hasCritical ? 'CRITICAL' : 'WARNING';
    if (!this.masterAlarmTimer || this.currentLoopSeverity !== targetSeverity) {
      this.restartMasterTimer(targetSeverity);
    }
  }

  /**
   * Stop an acknowledged, cleared, or dismissed alarm from repeating.
   */
  stopAlarmLoop(alarmId) {
    if (!alarmId) return;
    const id = Number(alarmId);
    this.activeAlarmMap.delete(id);

    if (this.activeAlarmMap.size === 0) {
      this.stopAllLoops();
    } else {
      let hasCritical = false;
      for (const item of this.activeAlarmMap.values()) {
        if (item.severity === 'CRITICAL') {
          hasCritical = true;
          break;
        }
      }
      const targetSeverity = hasCritical ? 'CRITICAL' : 'WARNING';
      if (this.currentLoopSeverity !== targetSeverity) {
        this.restartMasterTimer(targetSeverity);
      }
    }
  }

  /**
   * Compatibility shims for legacy callers
   */
  startCriticalAlarmLoop(alarmId) {
    if (!alarmId) return;
    const isFirstCritical = this.activeCriticalAlarmIds.size === 0 && this.masterCriticalTimer === null;
    this.activeCriticalAlarmIds.add(alarmId);
    this.startAlarmLoop(alarmId, 'CRITICAL');
  }

  stopCriticalAlarmLoop(alarmId) {
    if (!alarmId) return;
    this.activeCriticalAlarmIds.delete(alarmId);
    this.stopAlarmLoop(alarmId);
  }

  /**
   * Stop all active repeating loops (e.g. on mute, logout, or clear all)
   */
  stopAllLoops() {
    if (this.masterAlarmTimer) {
      clearInterval(this.masterAlarmTimer);
      this.masterAlarmTimer = null;
    }
    if (this.masterCriticalTimer) {
      clearInterval(this.masterCriticalTimer);
      this.masterCriticalTimer = null;
    }
    this.activeAlarmMap.clear();
    this.activeCriticalAlarmIds.clear();
    this.currentLoopSeverity = null;
  }

  /**
   * Toggle global mute
   */
  toggleMute(durationMinutes = null) {
    if (this.checkMuteActive()) {
      // Unmute
      this.isMuted = false;
      this.muteUntil = null;
    } else {
      // Mute
      this.isMuted = true;
      if (durationMinutes && durationMinutes > 0) {
        this.muteUntil = Date.now() + durationMinutes * 60 * 1000;
      } else {
        this.muteUntil = null; // Indefinite
      }
    }
    this.updateMasterGain();
    return this.isMuted;
  }

  /**
   * Explicitly set mute state with optional duration
   */
  setMute(muted, durationMinutes = null) {
    this.isMuted = !!muted;
    if (this.isMuted && durationMinutes && durationMinutes > 0) {
      this.muteUntil = Date.now() + durationMinutes * 60 * 1000;
    } else {
      this.muteUntil = null;
    }
    this.updateMasterGain();
  }

  /**
   * Update volume (0.0 to 1.0)
   */
  setVolume(vol) {
    this.volume = Math.max(0, Math.min(1, parseFloat(vol) || 0));
    this.updateMasterGain();
  }

  /**
   * Enable or disable all alarm sound
   */
  setSoundEnabled(enabled) {
    this.soundEnabled = !!enabled;
    this.updateMasterGain();
  }

  /**
   * Set repeat interval for unacknowledged alarms (in seconds)
   */
  setRepeatInterval(seconds) {
    const val = parseInt(seconds, 10);
    this.repeatIntervalSeconds = isNaN(val) ? 5 : val;

    if (this.activeAlarmMap.size > 0 && this.currentLoopSeverity) {
      this.restartMasterTimer(this.currentLoopSeverity);
    }
    this.notifyStateChange();
  }

  /**
   * Sound preview for testing from settings UI
   */
  previewSound(severity = 'WARNING') {
    // Ensure context is unlocked on click
    this.unlockAudio().then(() => {
      // Temporarily unmute if needed for preview test
      const wasMuted = this.isMuted;

      if (this.masterGain) {
        this.masterGain.gain.setValueAtTime(Math.max(0.1, this.volume), this.audioCtx.currentTime);
      }

      this.triggerPulse(severity);

      // Restore gain after preview duration
      setTimeout(() => {
        if (wasMuted || !this.soundEnabled) {
          this.updateMasterGain();
        }
      }, 500);
    });
  }

  /**
   * Add listener for synchronized pulse events
   */
  onPulse(callback) {
    this.onPulseCallbacks.add(callback);
    return () => this.onPulseCallbacks.delete(callback);
  }

  /**
   * Add listener for state changes (unlocked, muted, volume changes)
   */
  onStateChange(callback) {
    this.onStateChangeCallbacks.add(callback);
    return () => this.onStateChangeCallbacks.delete(callback);
  }

  notifyStateChange() {
    const state = {
      isUnlocked: this.isUnlocked || (this.audioCtx && this.audioCtx.state === 'running'),
      isMuted: this.checkMuteActive(),
      muteUntil: this.muteUntil,
      volume: this.volume,
      soundEnabled: this.soundEnabled,
      repeatIntervalSeconds: this.repeatIntervalSeconds,
    };
    for (const cb of this.onStateChangeCallbacks) {
      try { cb(state); } catch (e) {}
    }
  }
}

// Singleton export
export const audioAlert = new AudioAlertEngine();
export default audioAlert;
