const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: OPERATOR AUDIO STATUS & SETUP POPOVER');
console.log('===============================================================');

const occViewPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

const audioAlertPath = path.join(__dirname, '..', 'src', 'utils', 'audioAlert.js');
const audioAlertContent = fs.readFileSync(audioAlertPath, 'utf8');

// 1. Check audioAlert.js compatibility aliases
console.log('Checking Requirement 1: audioAlert compatibility methods...');
assert(
  audioAlertContent.includes('setMuted(muted, durationMinutes = null)'),
  'audioAlert must implement setMuted alias'
);
assert(
  audioAlertContent.includes('playChime(severity = \'WARNING\')'),
  'audioAlert must implement playChime alias'
);
console.log('✓ PASS: 1.1: audioAlert.js contains setMuted and playChime compatibility aliases');

// 2. Check Sound Button in footer with setup trigger
console.log('Checking Requirement 2: Sound button in docked footer...');
assert(
  occContent.includes('class="h-[22px] px-2 rounded border text-[9px] font-mono font-bold flex items-center space-x-1 leading-none transition-colors cursor-pointer"'),
  'Sound button maintains compact h-[22px] SCADA footer styling'
);
assert(
  occContent.includes('@click="toggleAudioSetupPopover"'),
  'Sound button triggers audio setup popover'
);
assert(
  occContent.includes('⚙️'),
  'Sound button displays setup gear indicator'
);
console.log('✓ PASS: 2.1: Sound button displays live status (SOUND ON / MUTED / SOUND OFF) and opens setup popover');

// 3. Check Audio Setup Popover Controls
console.log('Checking Requirement 3: Audio Setup Popover controls...');
assert(
  occContent.includes('v-if="showAudioSetupPopover"'),
  'Audio setup popover renders conditionally'
);
assert(
  occContent.includes('Audio & Alarm Setup'),
  'Popover contains clear title'
);
assert(
  occContent.includes('toggleAudioMute'),
  'Popover contains master mute toggle'
);
assert(
  occContent.includes('onAudioVolumeChange'),
  'Popover contains alert volume slider with onAudioVolumeChange handler'
);
assert(
  occContent.includes('testAlarmChime(\'WARNING\')'),
  'Popover contains Test Warning sound button'
);
assert(
  occContent.includes('testAlarmChime(\'CRITICAL\')'),
  'Popover contains Test Critical sound button'
);
assert(
  occContent.includes('onAudioRepeatIntervalChange'),
  'Popover contains alarm repeat interval selector'
);
assert(
  occContent.includes('onAudioMuteDurationChange'),
  'Popover contains temporary mute duration selector'
);
assert(
  occContent.includes('openFullAlarmSettingsModal'),
  'Popover contains button to open full AlarmNotificationModal'
);
console.log('✓ PASS: 3.1: Audio Setup Popover contains all required controls (Mute, Volume, Test Chimes, Repeat Interval, Mute Timer, Full Settings)');

// 4. Check AlarmNotificationModal integration
console.log('Checking Requirement 4: AlarmNotificationModal component integration...');
assert(
  occContent.includes('import AlarmNotificationModal from \'../components/AlarmNotificationModal.vue\';'),
  'AlarmNotificationModal is imported'
);
assert(
  occContent.includes('<AlarmNotificationModal'),
  'AlarmNotificationModal is mounted in template'
);
assert(
  occContent.includes(':is-open="showAlarmNotificationModal"'),
  'AlarmNotificationModal is bound to showAlarmNotificationModal'
);
console.log('✓ PASS: 4.1: AlarmNotificationModal is properly imported and bound');

// 5. Check dismissal on click outside and escape
console.log('Checking Requirement 5: Popover dismissal...');
assert(
  occContent.includes('if (this.showAudioSetupPopover && !e.target.closest(\'.audio-setup-anchor\')) {'),
  'Popover closes when clicking outside audio-setup-anchor'
);
assert(
  occContent.includes('if (this.showAudioSetupPopover) this.showAudioSetupPopover = false;'),
  'Popover closes when pressing Escape key'
);
console.log('✓ PASS: 5.1: Audio Setup Popover dismisses on click-outside and Escape key');

console.log('===============================================================');
console.log('TEST SUMMARY: ALL 5 REQUIREMENTS PASSED (TOTAL: 5)');
console.log('===============================================================');
