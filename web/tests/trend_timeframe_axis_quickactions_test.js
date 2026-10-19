const fs = require('fs');
const path = require('path');
const assert = require('assert');

console.log('===============================================================');
console.log('RUNNING TEST SUITE: TREND TIMEFRAME HEIGHT, X-AXIS LEGEND, & COMPACT QUICK ACTIONS');
console.log('===============================================================');

const occViewPath = path.join(__dirname, '..', 'src', 'views', 'OperatorCommandCenterView.vue');
const occContent = fs.readFileSync(occViewPath, 'utf8');

// 1. Timeframe Buttons & Action Icon Buttons Height Uniformity
console.log('Checking Requirement 1: Timeframe buttons & icon buttons height matching...');

// Timeframe pill container has h-[24px]
assert(
  occContent.includes('h-[24px] flex items-center bg-slate-100 dark:bg-[#101B2B] p-0.5 rounded border border-slate-200 dark:border-[#1E3048]'),
  'Timeframe preset buttons container must have h-[24px] to match adjacent icons'
);

// Timeframe buttons have h-full
assert(
  occContent.includes('class="h-full px-2 flex items-center justify-center rounded transition font-semibold leading-none cursor-pointer"'),
  'Timeframe buttons inside pill container must fill the 24px height with h-full'
);

// Download button has w-[24px] h-[24px]
assert(
  occContent.includes('w-[24px] h-[24px] flex items-center justify-center rounded border border-slate-200 dark:border-[#1E3048] bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-slate-600 dark:text-slate-400 hover:text-cyan-600 dark:hover:text-cyan-400 transition leading-none cursor-pointer'),
  'Download button must be sized at 24px x 24px'
);

// Fullscreen button has w-[24px] h-[24px]
assert(
  occContent.includes('@click="openTrendFullscreen"\n                class="w-[24px] h-[24px] flex items-center justify-center rounded border border-slate-200 dark:border-[#1E3048] bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#16253B] text-slate-600 dark:text-slate-400 hover:text-cyan-600 dark:hover:text-cyan-400 transition leading-none cursor-pointer"'),
  'Fullscreen button must be sized at 24px x 24px matching timeframe pill'
);

console.log('✓ PASS: 1.1: Timeframe toolbar container and adjacent download/fullscreen icon buttons have uniform 24px height');

// 2. Trend Analysis X-Axis Legend Clearance (no sinking or clipping)
console.log('Checking Requirement 2: X-axis time legend clearance...');

// In trendChartDims, padBottom must be at least 24
const padBottomMatch = occContent.match(/trendChartDims\(\)\s*\{[\s\S]*?padBottom\s*=\s*(\d+)/);
assert(padBottomMatch, 'padBottom must be configured in trendChartDims');
const padBottomVal = parseInt(padBottomMatch[1], 10);
assert(padBottomVal >= 24, `padBottom must be >= 24 to give sufficient clearance (got ${padBottomVal})`);

// X-axis text y coordinate inside time-ticks must provide clearance above bottom of canvas
assert(
  occContent.includes('<line :x1="tick.x" :y1="trendChartDims.baseY" :x2="tick.x" :y2="trendChartDims.baseY + 3.5" :stroke="axisStroke" stroke-width="1" />'),
  'Tick marker line must drop 3.5px from baseY'
);
assert(
  occContent.includes(':y="trendChartDims.baseY + 15"'),
  'X-axis timestamp label must be positioned comfortably below baseY within padBottom'
);

console.log('✓ PASS: 2.1: Trend chart X-axis labels have padBottom >= 24 and baseY + 15 clearance, completely preventing clipping');

// 3. Compact Quick Actions Bar & Buttons
console.log('Checking Requirement 3: Compact Quick Actions bar & buttons...');

// Docked SCADA footer bar height reduced to h-[28px]
assert(
  occContent.includes('fixed bottom-0 left-0 right-0 z-30 bg-white/95 dark:bg-[#0B1422]/95 backdrop-blur-md border-t border-slate-200 dark:border-[#20354B] px-2 sm:px-3 py-0.5 flex items-center justify-between gap-1 shadow-lg transition-colors h-[28px] select-none'),
  'Docked quick actions footer must have compact h-[28px]'
);

// Export Data button is compact h-[22px]
assert(
  occContent.includes('class="h-[22px] px-2 rounded bg-cyan-600 hover:bg-cyan-500 text-slate-950 font-bold text-[9.5px] font-mono transition-colors shadow-xs flex items-center space-x-1 leading-none cursor-pointer"'),
  'Export Data button must have compact h-[22px] with text-[9.5px]'
);

// Generate Report button is compact h-[22px]
assert(
  occContent.includes('class="h-[22px] px-2 rounded bg-blue-600 hover:bg-blue-500 text-white font-bold text-[9.5px] font-mono transition-colors flex items-center space-x-1 shadow-xs leading-none cursor-pointer"'),
  'Generate Report button must have compact h-[22px] with text-[9.5px]'
);

// Alarm History button is compact h-[22px]
assert(
  occContent.includes('class="h-[22px] px-1.5 sm:px-2 rounded bg-slate-100 hover:bg-slate-200 dark:bg-[#101B2B] dark:hover:bg-[#18283E] text-slate-700 dark:text-slate-200 border border-slate-300 dark:border-[#20354B] text-[9.5px] font-mono font-medium transition hidden md:flex items-center space-x-1 leading-none"'),
  'Alarm History link must have compact h-[22px] with text-[9.5px]'
);

// Sound On/Off toggle is compact h-[22px]
assert(
  occContent.includes('class="h-[22px] px-2 rounded border text-[9px] font-mono font-bold flex items-center space-x-1 leading-none transition-colors cursor-pointer"'),
  'Sound On/Off toggle button must have compact h-[22px] with text-[9px]'
);

console.log('✓ PASS: 3.1: Quick actions footer and operational buttons are smaller and more compact (h-22px, text-9.5px/9px)');

console.log('===============================================================');
console.log('TEST SUMMARY: ALL 3 REQUIREMENTS PASSED (TOTAL: 3)');
console.log('===============================================================');
