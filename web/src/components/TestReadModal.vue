<template>
  <div v-if="isOpen && targetParam" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in font-sans">
    <div class="bg-[#0F172A] border border-slate-700/80 rounded-2xl w-full max-w-3xl shadow-2xl overflow-hidden flex flex-col max-h-[92vh]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-slate-800 flex items-center justify-between bg-[#0B0F19]/90">
        <div class="flex items-center space-x-2.5">
          <div class="p-1.5 rounded-lg bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"></path>
            </svg>
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="text-sm font-bold text-white tracking-wide">{{ $t('deviceDetail.diagnosticReportTitle') || 'Engineering Modbus Diagnostic Report' }}</h2>
              <span class="text-3xs font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
                {{ targetDevice.device_code || targetDevice.code || '--' }} · {{ targetParam.parameter_code || targetParam.param_code || '--' }}
              </span>
            </div>
            <p class="text-3xs text-slate-400 font-mono mt-0.5">
              {{ targetParam.parameter_name || targetParam.param_name || targetParam.parameter_code }} · {{ paramDataType }} ({{ targetParam.unit || '--' }})
            </p>
          </div>
        </div>
        <button @click="closeModal" class="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800 transition">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
          </svg>
        </button>
      </div>

      <!-- Body -->
      <div class="p-6 space-y-4 text-xs font-sans overflow-y-auto max-h-[calc(92vh-130px)] custom-scrollbar">
        <!-- Parameter Specs Header Strip -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-2.5 p-3 bg-[#0B0F19] rounded-xl border border-slate-800 text-3xs font-mono">
          <div>
            <span class="text-slate-500 uppercase block font-sans font-semibold">{{ $t('deviceDetail.registerAndType') || 'Register & Type' }}</span>
            <span class="text-slate-200 font-bold mt-0.5 block">#{{ paramRegisterAddress }} ({{ paramRegisterType }})</span>
          </div>
          <div>
            <span class="text-slate-500 uppercase block font-sans font-semibold">{{ $t('deviceDetail.dataType') || 'Data Type' }}</span>
            <span class="text-blue-400 font-bold mt-0.5 block">{{ paramDataType }} [{{ paramByteOrder }}]</span>
          </div>
          <div>
            <span class="text-slate-500 uppercase block font-sans font-semibold">{{ $t('deviceDetail.scaleAndOffset') || 'Scale & Offset' }}</span>
            <span class="text-slate-200 mt-0.5 block">×{{ paramScale }} + {{ paramOffset }}</span>
          </div>
          <div>
            <span class="text-slate-500 uppercase block font-sans font-semibold">{{ $t('deviceDetail.configuredPort') || 'Port / Connection' }}</span>
            <span class="text-slate-300 truncate mt-0.5 block" :title="configuredPortDisplay">
              {{ configuredPortDisplay }}
            </span>
          </div>
        </div>

        <!-- Loading State -->
        <div v-if="testReadLoading" class="py-12 flex flex-col items-center justify-center space-y-3 bg-[#0B0F19]/50 rounded-xl border border-slate-800">
          <div class="w-8 h-8 border-2 border-emerald-500/20 border-t-emerald-500 rounded-full animate-spin"></div>
          <div class="text-center">
            <span class="text-xs text-slate-300 font-medium font-sans block">{{ $t('deviceDetail.executingTransaction') || 'Executing Transaction...' }}</span>
            <span class="text-3xs text-slate-500 font-mono mt-0.5 block">Verifying port ➔ Testing serial frame ➔ Validating CRC ➔ Decoding</span>
          </div>
        </div>

        <!-- Structured Result Content -->
        <div v-else-if="testReadResult" class="space-y-4">
          <!-- 1. Overall Status & Category Banner -->
          <div
            :class="testReadResult.success ? 'bg-emerald-950/30 border-emerald-500/40 text-emerald-200' : 'bg-rose-950/30 border-rose-500/40 text-rose-200'"
            class="p-4 rounded-xl border flex flex-col md:flex-row items-start md:items-center justify-between gap-3 shadow-lg"
          >
            <div class="flex items-center space-x-3">
              <span
                :class="testReadResult.success ? 'bg-emerald-500/20 text-emerald-400 border-emerald-500/40' : 'bg-rose-500/20 text-rose-400 border-rose-500/40'"
                class="px-2.5 py-1 rounded-lg text-xs font-black tracking-wider uppercase border flex items-center space-x-1.5"
              >
                <span class="w-2 h-2 rounded-full" :class="testReadResult.success ? 'bg-emerald-400 animate-pulse' : 'bg-rose-400'"></span>
                <span>{{ testReadResult.success ? ($t('deviceDetail.statusPass') || 'PASS') : ($t('deviceDetail.statusFail') || 'FAIL') }}</span>
              </span>
              <div>
                <div class="flex items-center space-x-2">
                  <span class="text-xs font-bold text-white font-mono uppercase tracking-wide">
                    {{ testReadResult.failure_category || (testReadResult.success ? 'NONE' : 'COMMUNICATION_ERROR') }}
                  </span>
                  <span class="text-3xs font-mono px-2 py-0.5 rounded bg-black/40 border border-slate-700/60 text-slate-300">
                    {{ testReadResult.response_time_ms || 0 }} ms
                  </span>
                </div>
                <div class="text-3xs text-slate-400 font-mono mt-0.5">
                  <span v-if="testReadResult.last_successful_stage">Last Confirmed: <strong class="text-emerald-400">{{ testReadResult.last_successful_stage }}</strong></span>
                  <span v-if="testReadResult.first_failed_stage" class="ml-2">| Failed at: <strong class="text-rose-400">{{ testReadResult.first_failed_stage }}</strong></span>
                </div>
              </div>
            </div>
            <div class="text-3xs text-slate-400 font-mono">
              {{ formatTimestamp(testReadResult.timestamp) }}
            </div>
          </div>

          <!-- 2. Four-Stage Pipeline Stepper -->
          <div class="grid grid-cols-2 md:grid-cols-4 gap-2">
            <div
              v-for="(st, idx) in getStageSummaryList()"
              :key="idx"
              :class="getStageCardClass(st.status)"
              class="p-2.5 rounded-xl border transition-all text-3xs font-mono"
            >
              <div class="flex items-center justify-between mb-1">
                <span class="text-slate-400 font-sans font-semibold uppercase truncate">{{ st.title }}</span>
                <span :class="getStageBadgeClass(st.status)" class="text-3xs font-bold px-1.5 py-0.2 rounded uppercase">
                  {{ getStageBadgeLabel(st.status) }}
                </span>
              </div>
              <div class="text-slate-300 text-3xs truncate mt-0.5" :title="st.message">
                {{ st.message || '--' }}
              </div>
            </div>
          </div>

          <!-- 3. Actionable Diagnosis & Suggested Action Alert -->
          <div
            v-if="testReadResult.suggested_action || (!testReadResult.success && (testReadResult.error || testReadResult.error_message))"
            :class="testReadResult.success ? 'bg-emerald-950/20 border-emerald-500/30' : 'bg-rose-950/25 border-rose-500/30'"
            class="p-3.5 rounded-xl border space-y-1.5"
          >
            <div class="flex items-center space-x-2">
              <svg v-if="!testReadResult.success" class="w-4 h-4 text-rose-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
              </svg>
              <svg v-else class="w-4 h-4 text-emerald-400 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
              <h4 class="text-xs font-bold text-white uppercase tracking-wider font-sans">
                {{ $t('deviceDetail.actionableDiagnosis') || 'Actionable Diagnosis' }}
              </h4>
            </div>
            <p class="text-xs text-slate-200 font-sans leading-relaxed pl-6">
              {{ testReadResult.suggested_action || testReadResult.error || testReadResult.error_message }}
            </p>
          </div>

          <!-- 4. Detailed Evidence Accordions / Panels -->
          <div class="space-y-3">
            <!-- Panel 1: Port & Physical Transport -->
            <div v-if="testReadResult.port_evidence" class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2.5">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-slate-200 uppercase font-sans flex items-center space-x-1.5">
                  <span class="w-1.5 h-1.5 rounded-full bg-blue-400"></span>
                  <span>{{ $t('deviceDetail.portEvidence') || 'Port Evidence' }}</span>
                </span>
                <span :class="testReadResult.port_evidence.exists ? 'text-emerald-400' : 'text-rose-400'" class="text-3xs font-mono font-bold">
                  {{ testReadResult.port_evidence.exists ? 'EXISTS' : 'NOT FOUND' }}
                </span>
              </div>
              <div class="grid grid-cols-2 md:grid-cols-3 gap-2 text-3xs font-mono">
                <div>
                  <span class="text-slate-500 uppercase block">{{ $t('deviceDetail.configuredPort') || 'Configured Port' }}</span>
                  <span class="text-slate-300 font-bold truncate block">{{ testReadResult.port_evidence.configured_port || '--' }}</span>
                </div>
                <div>
                  <span class="text-slate-500 uppercase block">{{ $t('deviceDetail.resolvedPort') || 'Resolved Port' }}</span>
                  <span class="text-slate-300 font-bold truncate block">{{ testReadResult.port_evidence.resolved_port || '--' }}</span>
                </div>
                <div>
                  <span class="text-slate-500 uppercase block">{{ $t('deviceDetail.symlinkTarget') || 'Symlink Target' }}</span>
                  <span :class="testReadResult.port_evidence.symlink_target_exists ? 'text-slate-300' : 'text-rose-400'" class="truncate block">
                    {{ testReadResult.port_evidence.symlink_target || (testReadResult.port_evidence.is_symlink ? 'Broken' : 'Direct Device') }}
                  </span>
                </div>
              </div>

              <!-- Available system ports if detected -->
              <div v-if="testReadResult.port_evidence.system_available_ports && testReadResult.port_evidence.system_available_ports.length" class="pt-1 border-t border-slate-800/80">
                <span class="text-3xs text-slate-400 uppercase font-semibold block mb-1">{{ $t('deviceDetail.detectedPorts') || 'Detected Ports' }}:</span>
                <div class="flex flex-wrap gap-1.5">
                  <span
                    v-for="(pPath, pIdx) in testReadResult.port_evidence.system_available_ports"
                    :key="pIdx"
                    class="text-3xs font-mono px-2 py-0.5 rounded bg-slate-800/80 text-emerald-300 border border-slate-700"
                  >
                    {{ pPath }}
                  </span>
                </div>
              </div>
            </div>

            <!-- Panel 2: Modbus Frame & CRC Inspection -->
            <div v-if="testReadResult.transaction_evidence" class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2.5">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-slate-200 uppercase font-sans flex items-center space-x-1.5">
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
                  <span>{{ $t('deviceDetail.modbusFrames') || 'Modbus Frames' }}</span>
                </span>
                <span
                  :class="testReadResult.transaction_evidence.crc_validation === 'PASS' ? 'text-emerald-400' : (testReadResult.transaction_evidence.crc_validation === 'FAIL' ? 'text-rose-400' : 'text-slate-500')"
                  class="text-3xs font-mono font-bold"
                >
                  CRC16: {{ testReadResult.transaction_evidence.crc_validation || 'NOT_CHECKED' }}
                </span>
              </div>

              <!-- Request Frame ADU -->
              <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800 space-y-1">
                <div class="flex items-center justify-between text-3xs font-mono">
                  <span class="text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.requestFrame') || 'Request Frame' }}</span>
                  <span :class="testReadResult.transaction_evidence.request_transmitted ? 'text-emerald-400' : 'text-slate-500'">
                    {{ testReadResult.transaction_evidence.request_transmitted ? ($t('deviceDetail.transmitted') || 'Transmitted') : ($t('deviceDetail.notTransmitted') || 'Not Transmitted') }}
                  </span>
                </div>
                <div class="text-xs font-mono font-bold text-blue-300 tracking-wider select-all">
                  {{ testReadResult.transaction_evidence.request_frame_hex || '--' }}
                </div>
              </div>

              <!-- Response Frame ADU -->
              <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800 space-y-1">
                <div class="flex items-center justify-between text-3xs font-mono">
                  <span class="text-slate-400 font-semibold uppercase">{{ $t('deviceDetail.responseFrame') || 'Response Frame' }}</span>
                  <span :class="testReadResult.transaction_evidence.response_received ? 'text-emerald-400' : 'text-rose-400'">
                    {{ testReadResult.transaction_evidence.response_received ? ($t('deviceDetail.received') || 'Received') : ($t('deviceDetail.noResponse') || 'No Response') }}
                  </span>
                </div>
                <div class="text-xs font-mono font-bold text-emerald-300 tracking-wider select-all">
                  {{ testReadResult.transaction_evidence.response_frame_hex || (testReadResult.transaction_evidence.response_received ? '--' : ($t('deviceDetail.noResponse') || 'No response')) }}
                </div>
                <div v-if="testReadResult.transaction_evidence.crc_expected_hex" class="text-3xs text-slate-400 font-mono pt-0.5">
                  {{ $t('deviceDetail.crcExpected') || 'CRC Expected' }}: 0x{{ testReadResult.transaction_evidence.crc_expected_hex }} | {{ $t('deviceDetail.crcReceived') || 'CRC Received' }}: 0x{{ testReadResult.transaction_evidence.crc_received_hex }}
                </div>
              </div>

              <!-- Modbus Exception Alert if any -->
              <div v-if="testReadResult.transaction_evidence.modbus_exception_name" class="p-2.5 bg-amber-950/40 rounded-lg border border-amber-500/30 text-amber-200 text-3xs font-mono">
                <strong>Modbus Exception Response:</strong> {{ testReadResult.transaction_evidence.modbus_exception_name }} (Physical communication link is active!)
              </div>
            </div>

            <!-- Panel 3: Data Interpretation & Values -->
            <div v-if="testReadResult.success" class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2.5">
              <span class="text-xs font-bold text-slate-200 uppercase font-sans flex items-center space-x-1.5">
                <span class="w-1.5 h-1.5 rounded-full bg-cyan-400"></span>
                <span>{{ $t('deviceDetail.decodedValue') || 'Decoded Value' }} & Telemetry</span>
              </span>

              <!-- Prominent Formula Result Card if formula configured -->
              <div v-if="getTestReadFormulaVal() !== null" class="p-3 bg-blue-950/30 rounded-xl border border-blue-500/30 space-y-1">
                <div class="flex items-center justify-between">
                  <span class="text-3xs text-blue-300 font-semibold uppercase tracking-wider flex items-center space-x-1.5">
                    <span class="w-2 h-2 rounded-full bg-blue-400"></span>
                    <span>{{ $t('parameters.formulaValue') || 'Formula Value' }} (ƒ(x))</span>
                  </span>
                  <span class="text-3xs font-mono text-blue-300 bg-blue-900/50 px-2 py-0.5 rounded border border-blue-700/50">
                    {{ testReadResult.formula || (targetParam && targetParam.formula) }}
                  </span>
                </div>
                <div class="flex items-baseline space-x-2 pt-1">
                  <span class="text-2xl font-mono font-extrabold text-blue-400 select-all">
                    {{ getTestReadFormulaVal() }}
                  </span>
                  <span class="text-xs font-bold text-slate-300 font-mono">{{ targetParam.unit }}</span>
                </div>
              </div>

              <div class="grid grid-cols-2 md:grid-cols-3 gap-2.5">
                <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
                  <div class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.modbusScaledValue') || 'Modbus Scaled Value' }}</div>
                  <div class="text-base font-mono font-bold text-emerald-400 mt-1 select-all">
                    {{ getTestReadModbusVal() }}
                    <span class="text-xs font-normal text-slate-300 ml-1">{{ targetParam.unit }}</span>
                  </div>
                </div>
                <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
                  <div class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.rawPduValue') || 'Raw PDU Value' }}</div>
                  <div class="text-xs font-mono text-slate-200 mt-1 select-all">
                    {{ testReadResult.raw_value !== undefined ? testReadResult.raw_value : '--' }}
                  </div>
                </div>
                <div class="p-2.5 bg-black/40 rounded-lg border border-slate-800">
                  <div class="text-3xs text-slate-400 uppercase font-semibold">{{ $t('deviceDetail.rawWords') || 'Raw Words' }}</div>
                  <div class="text-xs font-mono text-slate-200 mt-1 truncate select-all" :title="testReadResult.data_evidence && testReadResult.data_evidence.raw_register_hex ? testReadResult.data_evidence.raw_register_hex.join(' ') : '--'">
                    {{ (testReadResult.data_evidence && testReadResult.data_evidence.raw_register_hex && testReadResult.data_evidence.raw_register_hex.join(' ')) || '--' }}
                  </div>
                </div>
              </div>
            </div>

            <!-- Panel 4: Terminal mbpoll Command for Side-by-Side Comparison -->
            <div v-if="testReadResult.mbpoll_command" class="p-3.5 bg-[#0B0F19] rounded-xl border border-slate-800 space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-xs font-bold text-slate-200 uppercase font-sans flex items-center space-x-1.5">
                  <svg class="w-3.5 h-3.5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
                  </svg>
                  <span>{{ $t('deviceDetail.mbpollCommandTitle') || 'mbpoll Diagnostic Command' }}</span>
                </span>
                <button
                  type="button"
                  @click="copyMbpollCommand"
                  class="px-2.5 py-1 rounded-lg text-3xs font-bold text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 border border-slate-700 transition flex items-center space-x-1"
                >
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 5H6a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2v-1M8 5a2 2 0 002 2h2a2 2 0 002-2M8 5a2 2 0 012-2h2a2 2 0 012 2m0 0h2a2 2 0 012 2v3m2 4H10m0 0l3-3m-3 3l3 3"/>
                  </svg>
                  <span>{{ copyStatusText === 'mbpoll' ? ($t('common.copied') || 'Copied!') : ($t('deviceDetail.copyMbpoll') || 'Copy mbpoll') }}</span>
                </button>
              </div>
              <div class="p-2.5 bg-black/60 rounded-lg border border-slate-800 font-mono text-3xs text-amber-300 select-all overflow-x-auto whitespace-pre">
{{ testReadResult.mbpoll_command }}
              </div>
              <p class="text-3xs text-slate-500 font-sans italic">
                * {{ $t('deviceDetail.safeAccessNotice') || 'Coordinate access; stop or pause active polling before running external commands.' }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="px-6 py-3 border-t border-slate-800 flex items-center justify-between bg-[#0B0F19]/80">
        <div>
          <button
            v-if="testReadResult"
            type="button"
            @click="copyDiagnosticReport"
            class="px-3 py-1.5 rounded-xl text-xs font-semibold text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 border border-slate-700 transition flex items-center space-x-1.5"
          >
            <svg class="w-3.5 h-3.5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"/>
            </svg>
            <span>{{ copyStatusText === 'report' ? ($t('common.copied') || 'Copied!') : ($t('deviceDetail.copyReport') || 'Copy Report') }}</span>
          </button>
        </div>
        <div class="flex items-center space-x-2">
          <button
            type="button"
            @click="closeModal"
            class="px-3.5 py-1.5 rounded-xl text-xs font-medium text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            {{ $t('common.close') || 'Close' }}
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
            <span>{{ $t('deviceDetail.readAgain') || 'Read Again' }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'TestReadModal',
  props: {
    isOpen: {
      type: Boolean,
      default: false,
    },
    device: {
      type: Object,
      default: () => ({}),
    },
    parameter: {
      type: Object,
      default: () => null,
    },
  },
  data() {
    return {
      testReadLoading: false,
      testReadResult: null,
      copyStatusText: null,
      copyTimer: null,
    };
  },
  computed: {
    targetDevice() {
      return this.device || {};
    },
    targetParam() {
      return this.parameter || null;
    },
    paramRegisterAddress() {
      if (!this.targetParam) return '--';
      return this.targetParam.register_address !== undefined
        ? this.targetParam.register_address
        : (this.targetParam._rawParam && this.targetParam._rawParam.register_address !== undefined ? this.targetParam._rawParam.register_address : '--');
    },
    paramRegisterType() {
      if (!this.targetParam) return 'HOLDING_REGISTER';
      return this.targetParam.register_type || (this.targetParam._rawParam && this.targetParam._rawParam.register_type) || 'HOLDING_REGISTER';
    },
    paramDataType() {
      if (!this.targetParam) return 'FLOAT32';
      return this.targetParam.data_type || (this.targetParam._rawParam && this.targetParam._rawParam.data_type) || 'FLOAT32';
    },
    paramByteOrder() {
      if (!this.targetParam) return 'ABCD';
      return this.targetParam.byte_order ||
        (this.targetParam._rawParam && this.targetParam._rawParam.byte_order) ||
        (this.targetDevice.connection && this.targetDevice.connection.byte_order) ||
        'ABCD';
    },
    paramScale() {
      if (!this.targetParam) return 1;
      if (this.targetParam.scale !== undefined) return this.targetParam.scale;
      if (this.targetParam._rawParam && this.targetParam._rawParam.scale !== undefined) return this.targetParam._rawParam.scale;
      return 1;
    },
    paramOffset() {
      if (!this.targetParam) return 0;
      if (this.targetParam.offset !== undefined) return this.targetParam.offset;
      if (this.targetParam._rawParam && this.targetParam._rawParam.offset !== undefined) return this.targetParam._rawParam.offset;
      return 0;
    },
    configuredPortDisplay() {
      if (this.targetDevice.connection) {
        return this.targetDevice.connection.serial_port ||
          (this.targetDevice.connection.host ? (this.targetDevice.connection.host + ':' + this.targetDevice.connection.port) : '--');
      }
      return '--';
    },
  },
  watch: {
    isOpen(newVal) {
      if (newVal) {
        this.executeTestRead();
      } else {
        this.testReadResult = null;
        this.testReadLoading = false;
      }
    },
  },
  methods: {
    closeModal() {
      this.$emit('close');
    },
    async executeTestRead() {
      if (!this.targetDevice || !this.targetParam) return;
      const devId = this.targetDevice.id || this.targetDevice.device_id;
      const paramId = this.targetParam.id || this.targetParam.param_id;
      if (!devId || !paramId) return;

      this.testReadLoading = true;
      this.testReadResult = null;
      try {
        const res = await this.$store.dispatch('testReadParameter', {
          deviceId: devId,
          paramId: paramId,
          parameterId: paramId,
        });
        this.testReadResult = (res && res.data && typeof res.data === 'object' && res.data.success !== undefined) ? res.data : res;
        this.$emit('read-completed', {
          device: this.targetDevice,
          parameter: this.targetParam,
          result: this.testReadResult,
        });
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
    formatTimestamp(ts) {
      if (!ts) return '--';
      try {
        return new Date(ts).toLocaleString();
      } catch (e) {
        return ts;
      }
    },
    getTestReadFormulaValNum() {
      if (!this.testReadResult || !this.testReadResult.success) return null;
      if (this.testReadResult.formula_value !== undefined && this.testReadResult.formula_value !== null) {
        return Number(this.testReadResult.formula_value);
      }
      const formula = this.testReadResult.formula || (this.targetParam && this.targetParam.formula);
      if (!formula || typeof formula !== 'string' || !formula.trim()) return null;
      const x = this.getTestReadModbusValNum();
      const raw = this.testReadResult.raw_value !== undefined ? this.testReadResult.raw_value : x;
      try {
        const roundFn = (val, prec) => {
          if (prec === undefined) return Math.round(val);
          const factor = Math.pow(10, Math.trunc(prec));
          return Math.round(val * factor) / factor;
        };
        const randomFn = (min, max) => {
          if (min === undefined || max === undefined || min > max) return min !== undefined ? min : 0;
          return min + Math.random() * (max - min);
        };
        const context = {
          x, val: x, value: x, raw,
          pi: Math.PI, e: Math.E,
          abs: Math.abs, round: roundFn, floor: Math.floor, ceil: Math.ceil,
          sqrt: Math.sqrt, min: Math.min, max: Math.max, pow: Math.pow,
          log: Math.log, ln: Math.log, log10: Math.log10, exp: Math.exp,
          sin: Math.sin, cos: Math.cos, tan: Math.tan,
          random: randomFn
        };
        let expr = formula.replace(/\^/g, '**');
        expr = expr.replace(/[a-zA-Z_][a-zA-Z0-9_]*/g, m => m.toLowerCase());
        const paramNames = Object.keys(context);
        const paramValues = Object.values(context);
        const fn = new Function(...paramNames, `"use strict"; return (${expr});`);
        const res = fn(...paramValues);
        if (typeof res === 'number' && !isNaN(res) && isFinite(res)) {
          return res;
        }
      } catch (e) {}
      return null;
    },
    getTestReadFormulaVal() {
      const num = this.getTestReadFormulaValNum();
      if (num === null) return null;
      const prec = this.targetParam && this.targetParam.precision !== undefined ? this.targetParam.precision : 2;
      return num.toFixed(prec);
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
      const prec = this.targetParam && this.targetParam.precision !== undefined ? this.targetParam.precision : 2;
      if (this.testReadResult.scaled_value !== undefined && this.testReadResult.scaled_value !== null) {
        return Number(this.testReadResult.scaled_value).toFixed(prec);
      }
      if (this.testReadResult.decoded_value !== undefined && this.testReadResult.decoded_value !== null) {
        return Number(this.testReadResult.decoded_value).toFixed(prec);
      }
      return '--';
    },
    getStageSummaryList() {
      if (!this.testReadResult) return [];
      if (this.testReadResult.stages && Array.isArray(this.testReadResult.stages) && this.testReadResult.stages.length) {
        const stageTitles = {
          PORT_VERIFICATION: this.$t('deviceDetail.stagePort') || 'Port Verification',
          SERIAL_CONFIGURATION: this.$t('deviceDetail.stageSerial') || 'Serial Framing',
          MODBUS_TRANSACTION: this.$t('deviceDetail.stageTransaction') || 'Modbus Transaction',
          DATA_INTERPRETATION: this.$t('deviceDetail.stageData') || 'Data Decoding',
        };
        return this.testReadResult.stages.map(st => ({
          stage: st.stage,
          title: stageTitles[st.stage] || st.stage,
          status: st.status,
          message: st.message,
        }));
      }
      const pass = !!this.testReadResult.success;
      return [
        { stage: 'PORT_VERIFICATION', title: this.$t('deviceDetail.stagePort') || 'Port Verification', status: pass ? 'PASS' : 'FAIL', message: pass ? 'Port accessible' : (this.testReadResult.error || 'Port verification failed') },
        { stage: 'SERIAL_CONFIGURATION', title: this.$t('deviceDetail.stageSerial') || 'Serial Framing', status: pass ? 'PASS' : 'NOT_TESTED', message: pass ? 'Framing valid' : 'Not tested' },
        { stage: 'MODBUS_TRANSACTION', title: this.$t('deviceDetail.stageTransaction') || 'Modbus Transaction', status: pass ? 'PASS' : 'NOT_TESTED', message: pass ? 'Transaction complete' : 'Not tested' },
        { stage: 'DATA_INTERPRETATION', title: this.$t('deviceDetail.stageData') || 'Data Decoding', status: pass ? 'PASS' : 'NOT_TESTED', message: pass ? 'Decoded successfully' : 'Not tested' },
      ];
    },
    getStageCardClass(status) {
      switch (status) {
        case 'PASS':
          return 'bg-emerald-950/20 border-emerald-500/30 text-emerald-300';
        case 'FAIL':
          return 'bg-rose-950/20 border-rose-500/30 text-rose-300';
        case 'INSUFFICIENT_EVIDENCE':
          return 'bg-amber-950/20 border-amber-500/30 text-amber-300';
        case 'NOT_TESTED':
        default:
          return 'bg-slate-900/40 border-slate-800 text-slate-500 opacity-60';
      }
    },
    getStageBadgeClass(status) {
      switch (status) {
        case 'PASS':
          return 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40';
        case 'FAIL':
          return 'bg-rose-500/20 text-rose-400 border border-rose-500/40';
        case 'INSUFFICIENT_EVIDENCE':
          return 'bg-amber-500/20 text-amber-400 border border-amber-500/40';
        case 'NOT_TESTED':
        default:
          return 'bg-slate-800 text-slate-400 border border-slate-700';
      }
    },
    getStageBadgeLabel(status) {
      switch (status) {
        case 'PASS':
          return this.$t('deviceDetail.statusPass') || 'PASS';
        case 'FAIL':
          return this.$t('deviceDetail.statusFail') || 'FAIL';
        case 'INSUFFICIENT_EVIDENCE':
          return this.$t('deviceDetail.statusInsufficient') || 'INSUFFICIENT';
        case 'NOT_TESTED':
        default:
          return this.$t('deviceDetail.statusNotTested') || 'NOT TESTED';
      }
    },
    async copyMbpollCommand() {
      if (!this.testReadResult || !this.testReadResult.mbpoll_command) return;
      try {
        await navigator.clipboard.writeText(this.testReadResult.mbpoll_command);
      } catch (err) {
        this.copyToClipboardFallback(this.testReadResult.mbpoll_command);
      }
      this.flashCopyStatus('mbpoll');
    },
    async copyDiagnosticReport() {
      if (!this.testReadResult) return;
      const r = this.testReadResult;
      const dev = this.targetDevice;
      const p = this.targetParam;

      const reportLines = [
        `# MODBUS ENGINEERING DIAGNOSTIC REPORT`,
        `Generated: ${new Date().toISOString()}`,
        `Diagnostic ID: ${r.diagnostic_id || '--'}`,
        `Device: ${dev.device_code || dev.name || '--'} (ID: ${dev.id || dev.device_id})`,
        `Parameter: ${p.parameter_name || p.param_name || p.parameter_code || '--'} [${p.parameter_code || p.param_code || '--'}]`,
        `Overall Verdict: ${r.success ? 'PASS' : 'FAIL'}`,
        `Response Time: ${r.response_time_ms || 0} ms`,
        `Root Cause Category: ${r.failure_category || (r.success ? 'NONE' : 'COMMUNICATION_ERROR')}`,
        `Last Confirmed Stage: ${r.last_successful_stage || '--'}`,
        `First Failed Stage: ${r.first_failed_stage || '--'}`,
        ``,
        `## PIPELINE STAGES`,
      ];

      const stages = this.getStageSummaryList();
      for (const st of stages) {
        reportLines.push(`- [${st.status}] ${st.title}: ${st.message || '--'}`);
      }

      if (r.port_evidence) {
        reportLines.push(``);
        reportLines.push(`## PORT & HARDWARE EVIDENCE`);
        reportLines.push(`- Configured Port: ${r.port_evidence.configured_port || '--'}`);
        reportLines.push(`- Resolved Port: ${r.port_evidence.resolved_port || '--'}`);
        reportLines.push(`- Exists: ${r.port_evidence.exists}`);
        reportLines.push(`- Symlink Target: ${r.port_evidence.symlink_target || 'N/A'}`);
        reportLines.push(`- Symlink Target Exists: ${r.port_evidence.symlink_target_exists}`);
        reportLines.push(`- Accessible: ${r.port_evidence.can_access}`);
        if (r.port_evidence.permission_error) reportLines.push(`- Permission Error: ${r.port_evidence.permission_error}`);
        reportLines.push(`- Port Busy: ${r.port_evidence.is_busy} ${r.port_evidence.busy_reason ? '(' + r.port_evidence.busy_reason + ')' : ''}`);
        reportLines.push(`- Adapter Owned: ${r.port_evidence.adapter_owned}`);
        if (r.port_evidence.system_available_ports && r.port_evidence.system_available_ports.length) {
          reportLines.push(`- Detected System Ports: ${r.port_evidence.system_available_ports.join(', ')}`);
        }
      }

      if (r.serial_config) {
        reportLines.push(``);
        reportLines.push(`## SERIAL CONFIGURATION`);
        reportLines.push(`- Baud Rate: ${r.serial_config.baud_rate}`);
        reportLines.push(`- Framing: ${r.serial_config.data_bits}${r.serial_config.parity}${r.serial_config.stop_bits}`);
        reportLines.push(`- Timeout: ${r.serial_config.timeout_ms} ms`);
      }

      if (r.transaction_evidence) {
        reportLines.push(``);
        reportLines.push(`## MODBUS TRANSACTION EVIDENCE`);
        reportLines.push(`- Slave ID: ${r.transaction_evidence.slave_id}`);
        reportLines.push(`- Function Code: ${r.transaction_evidence.function_code}`);
        reportLines.push(`- Register: #${r.transaction_evidence.configured_address} (${r.transaction_evidence.register_type}), PDU Offset: ${r.transaction_evidence.pdu_address}, Quantity: ${r.transaction_evidence.quantity}`);
        reportLines.push(`- Request Transmitted: ${r.transaction_evidence.request_transmitted}`);
        if (r.transaction_evidence.request_frame_hex) reportLines.push(`- Request ADU Hex: ${r.transaction_evidence.request_frame_hex}`);
        reportLines.push(`- Response Received: ${r.transaction_evidence.response_received}`);
        if (r.transaction_evidence.response_frame_hex) reportLines.push(`- Response ADU Hex: ${r.transaction_evidence.response_frame_hex}`);
        reportLines.push(`- CRC Validation: ${r.transaction_evidence.crc_validation}`);
        if (r.transaction_evidence.crc_expected_hex) reportLines.push(`- Expected CRC: 0x${r.transaction_evidence.crc_expected_hex} | Received CRC: 0x${r.transaction_evidence.crc_received_hex}`);
        if (r.transaction_evidence.modbus_exception_name) reportLines.push(`- Modbus Exception: ${r.transaction_evidence.modbus_exception_name} (Code: 0x${r.transaction_evidence.modbus_exception_code.toString(16)})`);
      }

      if (r.data_evidence) {
        reportLines.push(``);
        reportLines.push(`## DATA INTERPRETATION`);
        reportLines.push(`- Data Type: ${r.data_evidence.data_type}`);
        reportLines.push(`- Byte Order: ${r.data_evidence.byte_order}`);
        if (r.data_evidence.raw_register_hex && r.data_evidence.raw_register_hex.length) {
          reportLines.push(`- Raw Words: ${r.data_evidence.raw_register_hex.join(' ')}`);
        }
        if (r.data_evidence.raw_value !== undefined && r.data_evidence.raw_value !== null) reportLines.push(`- Raw Value: ${r.data_evidence.raw_value}`);
        if (r.data_evidence.decoded_value !== undefined && r.data_evidence.decoded_value !== null) reportLines.push(`- Decoded Value: ${r.data_evidence.decoded_value}`);
        if (r.data_evidence.scaled_value !== undefined && r.data_evidence.scaled_value !== null) reportLines.push(`- Scaled Value: ${r.data_evidence.scaled_value} ${r.data_evidence.unit || ''}`);
        if (r.data_evidence.formula) reportLines.push(`- Formula: ${r.data_evidence.formula} -> ${r.data_evidence.formula_value}`);
      }

      if (r.suggested_action) {
        reportLines.push(``);
        reportLines.push(`## RECOMMENDATION / NEXT ACTION`);
        reportLines.push(`${r.suggested_action}`);
      }

      if (r.mbpoll_command) {
        reportLines.push(``);
        reportLines.push(`## MBPOLL COMMAND (SAFE TERMINAL REPRODUCTION)`);
        reportLines.push(`\`\`\`bash`);
        reportLines.push(`${r.mbpoll_command}`);
        reportLines.push(`\`\`\``);
        reportLines.push(`Note: Coordinate access; stop/pause datalogger polling before running mbpoll.`);
      }

      const reportText = reportLines.join('\n');
      try {
        await navigator.clipboard.writeText(reportText);
      } catch (err) {
        this.copyToClipboardFallback(reportText);
      }
      this.flashCopyStatus('report');
    },
    copyToClipboardFallback(text) {
      const textarea = document.createElement('textarea');
      textarea.value = text;
      textarea.style.position = 'fixed';
      textarea.style.left = '-9999px';
      document.body.appendChild(textarea);
      textarea.select();
      try {
        document.execCommand('copy');
      } catch (e) {
        console.error('Fallback clipboard copy failed:', e);
      }
      document.body.removeChild(textarea);
    },
    flashCopyStatus(type) {
      this.copyStatusText = type;
      if (this.copyTimer) clearTimeout(this.copyTimer);
      this.copyTimer = setTimeout(() => {
        this.copyStatusText = null;
      }, 2500);
    },
  },
};
</script>
