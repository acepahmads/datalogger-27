import * as XLSX from 'xlsx';
import { jsPDF } from 'jspdf';
import autoTable from 'jspdf-autotable';

export function triggerReportBlobDownload(blob, filename) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  setTimeout(() => URL.revokeObjectURL(url), 1500);
}

function fmtNum(val, decimals = 2) {
  if (val === null || val === undefined || Number.isNaN(Number(val))) return '-';
  return Number(val).toFixed(decimals);
}

function fmtDelta(firstVal, lastVal) {
  if (
    firstVal === null ||
    firstVal === undefined ||
    lastVal === null ||
    lastVal === undefined ||
    Number.isNaN(Number(firstVal)) ||
    Number.isNaN(Number(lastVal))
  ) {
    return '-';
  }
  const diff = Number(lastVal) - Number(firstVal);
  const sign = diff > 0 ? '+' : '';
  return `${fmtNum(firstVal, 2)} -> ${fmtNum(lastVal, 2)} (${sign}${diff.toFixed(2)})`;
}

function fmtTs(ts) {
  if (!ts) return '-';
  const d = new Date(ts);
  if (Number.isNaN(d.getTime())) return String(ts);
  return d.toLocaleString('en-GB', { hour12: false });
}

export function exportOperationalReportXLS(reportData, reportType = 'COMPREHENSIVE') {
  if (!reportData) return;

  const wb = XLSX.utils.book_new();
  const comm = reportData.communication_summary || {};
  const alm = reportData.alarm_summary || {};
  const params = Array.isArray(reportData.parameter_summaries) ? reportData.parameter_summaries : [];
  const devices = Array.isArray(reportData.device_summaries) ? reportData.device_summaries : [];
  const alarms = Array.isArray(alm.recent_alarms) ? alm.recent_alarms : [];

  // 1. Executive Summary Sheet
  const summaryRows = [
    ['OPERATIONAL TELEMETRY & COMPLIANCE REPORT', ''],
    ['Report Title', reportData.title || 'Operational Telemetry & Performance Report'],
    ['Report Focus', reportType || reportData.report_type || 'COMPREHENSIVE'],
    ['Reporting Period', reportData.period_formatted || '-'],
    ['Generated At', reportData.generated_at_formatted || '-'],
    ['Timezone', reportData.timezone || 'Asia/Jakarta'],
    ['', ''],
    ['EXECUTIVE KPI METRICS', 'VALUE'],
    ['Fleet Availability (%)', Number((reportData.overall_availability || 0).toFixed(2))],
    ['Good Data Quality (%)', Number((reportData.overall_quality_good || 0).toFixed(2))],
    ['Total Valid Samples', Number(reportData.total_samples || 0)],
    ['Monitored Devices', Number(reportData.total_devices || 0)],
    ['Monitored Parameters', Number(reportData.total_parameters || 0)],
    ['Communication Frames OK', Number(comm.success_frames || 0)],
    ['Communication Frames ERR', Number(comm.error_frames || 0)],
    ['Average Poll Latency (ms)', Number((comm.avg_latency_ms || 0).toFixed(2))],
    ['Active Alarms', Number(alm.active_count || 0)],
    ['Total Recorded Alarms', Number(alm.total_alarms || 0)],
    ['Critical Alarms', Number(alm.critical_count || 0)],
    ['Warning Alarms', Number(alm.warning_count || 0)],
    ['Cleared Alarms', Number(alm.cleared_count || 0)],
  ];
  const wsSummary = XLSX.utils.aoa_to_sheet(summaryRows);
  wsSummary['!cols'] = [{ wch: 32 }, { wch: 48 }];
  XLSX.utils.book_append_sheet(wb, wsSummary, 'Ringkasan_Eksekutif');

  // 2. Parameter Telemetry Sheet
  const paramHeaders = [
    'No',
    'Device Code',
    'Device Name',
    'Parameter Code',
    'Parameter Name',
    'Unit',
    'Samples',
    'Min',
    'Average',
    'Max',
    'First Value',
    'Last Value',
    'Delta (Last - First)',
    'Good (%)',
    'Bad (%)',
    'Uncertain (%)',
    'Data Source',
  ];
  const paramRows = params.map((p, idx) => {
    const hasDelta = p.first_value !== null && p.first_value !== undefined && p.last_value !== null && p.last_value !== undefined;
    const deltaVal = hasDelta ? Number((Number(p.last_value) - Number(p.first_value)).toFixed(4)) : '-';
    return [
      idx + 1,
      p.device_code || '-',
      p.device_name || '-',
      p.parameter_code || '-',
      p.parameter_name || '-',
      p.unit || '-',
      Number(p.sample_count || 0),
      p.min_value !== null && p.min_value !== undefined ? Number(Number(p.min_value).toFixed(4)) : '-',
      p.avg_value !== null && p.avg_value !== undefined ? Number(Number(p.avg_value).toFixed(4)) : '-',
      p.max_value !== null && p.max_value !== undefined ? Number(Number(p.max_value).toFixed(4)) : '-',
      p.first_value !== null && p.first_value !== undefined ? Number(Number(p.first_value).toFixed(4)) : '-',
      p.last_value !== null && p.last_value !== undefined ? Number(Number(p.last_value).toFixed(4)) : '-',
      deltaVal,
      Number((p.good_percent || 0).toFixed(2)),
      Number((p.bad_percent || 0).toFixed(2)),
      Number((p.uncertain_percent || 0).toFixed(2)),
      p.data_source === 'RAW_TELEMETRY' ? 'RAW' : (p.data_source || '-'),
    ];
  });
  const wsParams = XLSX.utils.aoa_to_sheet([paramHeaders, ...paramRows]);
  wsParams['!cols'] = [
    { wch: 5 }, { wch: 20 }, { wch: 24 }, { wch: 16 }, { wch: 22 },
    { wch: 10 }, { wch: 11 }, { wch: 12 }, { wch: 12 }, { wch: 12 },
    { wch: 12 }, { wch: 12 }, { wch: 18 }, { wch: 11 }, { wch: 10 },
    { wch: 13 }, { wch: 14 },
  ];
  XLSX.utils.book_append_sheet(wb, wsParams, 'Statistik_Parameter');

  // 3. Fleet Communication Audit Sheet
  const devHeaders = [
    'No',
    'Device Code',
    'Device Name',
    'Status',
    'Connection',
    'Latency (ms)',
    'Packets OK',
    'Packets ERR',
    'Active Parameters',
    'Availability (%)',
  ];
  const devRows = devices.map((d, idx) => [
    idx + 1,
    d.device_code || '-',
    d.device_name || '-',
    d.status || '-',
    d.connection_status || '-',
    Number(d.latency_ms || 0),
    Number(d.success_count || 0),
    Number(d.failed_count || 0),
    Number(d.parameters_count || 0),
    Number((d.availability_percent || 0).toFixed(2)),
  ]);
  const wsDevs = XLSX.utils.aoa_to_sheet([devHeaders, ...devRows]);
  wsDevs['!cols'] = [
    { wch: 5 }, { wch: 22 }, { wch: 26 }, { wch: 12 }, { wch: 14 },
    { wch: 14 }, { wch: 13 }, { wch: 13 }, { wch: 18 }, { wch: 16 },
  ];
  XLSX.utils.book_append_sheet(wb, wsDevs, 'Audit_Komunikasi_Fleet');

  // 4. Alarms Audit Sheet
  const almHeaders = [
    'No',
    'Alarm Code',
    'Device',
    'Severity',
    'Status',
    'Message',
    'Triggered Time',
    'Acknowledged Time',
    'Cleared Time',
  ];
  const almRows = alarms.map((a, idx) => [
    idx + 1,
    a.alarm_code || '-',
    a.device_name || '-',
    a.severity || '-',
    a.status || '-',
    a.message || '-',
    fmtTs(a.triggered_at),
    fmtTs(a.acknowledged_at),
    fmtTs(a.cleared_at),
  ]);
  const wsAlarms = XLSX.utils.aoa_to_sheet([almHeaders, ...almRows]);
  wsAlarms['!cols'] = [
    { wch: 5 }, { wch: 18 }, { wch: 22 }, { wch: 12 }, { wch: 14 },
    { wch: 38 }, { wch: 20 }, { wch: 20 }, { wch: 20 },
  ];
  XLSX.utils.book_append_sheet(wb, wsAlarms, 'Log_Alarm_Operasional');

  const xlsBuffer = XLSX.write(wb, { bookType: 'xlsx', type: 'array' });
  const blob = new Blob([xlsBuffer], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  });
  const slug = String(reportType || 'comprehensive').toLowerCase();
  triggerReportBlobDownload(blob, `laporan_operasional_${slug}_${Date.now()}.xlsx`);
}

export function exportOperationalReportPDF(reportData, reportType = 'COMPREHENSIVE') {
  if (!reportData) return;

  const comm = reportData.communication_summary || {};
  const alm = reportData.alarm_summary || {};
  const params = Array.isArray(reportData.parameter_summaries) ? reportData.parameter_summaries : [];
  const devices = Array.isArray(reportData.device_summaries) ? reportData.device_summaries : [];
  const alarms = Array.isArray(alm.recent_alarms) ? alm.recent_alarms : [];

  const doc = new jsPDF({
    orientation: 'landscape',
    unit: 'mm',
    format: 'a4',
  });

  // Top Header Banner
  doc.setFillColor(15, 23, 42);
  doc.rect(0, 0, 297, 20, 'F');

  doc.setFont('helvetica', 'bold');
  doc.setFontSize(12);
  doc.setTextColor(255, 255, 255);
  doc.text(
    `${(reportData.title || 'OPERATIONAL TELEMETRY & PERFORMANCE REPORT').toUpperCase()} [${reportType}]`,
    12,
    9
  );

  doc.setFont('helvetica', 'normal');
  doc.setFontSize(8);
  doc.setTextColor(203, 213, 225);
  doc.text(`Periode: ${reportData.period_formatted || '-'}`, 12, 15);

  doc.setFontSize(7.5);
  doc.setTextColor(148, 163, 184);
  doc.text(`Generated: ${reportData.generated_at_formatted || new Date().toLocaleString()}`, 285, 9, {
    align: 'right',
  });
  doc.text('HIGH-INTEGRITY SERVER AGGREGATION', 285, 15, { align: 'right' });

  // Executive KPI Summary Table
  autoTable(doc, {
    startY: 24,
    head: [['FLEET AVAILABILITY', 'GOOD DATA QUALITY', 'ACTIVE / TOTAL ALARMS', 'MONITORED FLEET SCOPE']],
    body: [[
      `${(reportData.overall_availability || 0).toFixed(1)}% (${comm.success_frames || 0} OK / ${comm.error_frames || 0} ERR, Avg ${(comm.avg_latency_ms || 0).toFixed(1)} ms)`,
      `${(reportData.overall_quality_good || 0).toFixed(1)}% (${Number(reportData.total_samples || 0).toLocaleString()} Samples)`,
      `${alm.active_count || 0} Active / ${alm.total_alarms || 0} Total (${alm.critical_count || 0} Crit, ${alm.warning_count || 0} Warn, ${alm.cleared_count || 0} Cleared)`,
      `${reportData.total_devices || 0} Devices • ${reportData.total_parameters || 0} Active Parameters`,
    ]],
    theme: 'grid',
    styles: { fontSize: 7.5, cellPadding: 2, textColor: [15, 23, 42], fontStyle: 'bold' },
    headStyles: { fillColor: [30, 58, 138], textColor: [255, 255, 255], fontSize: 7 },
    margin: { left: 12, right: 12 },
  });

  let cursorY = (doc.lastAutoTable && doc.lastAutoTable.finalY ? doc.lastAutoTable.finalY : 38) + 5;

  const showTelemetry = reportType === 'COMPREHENSIVE' || reportType === 'TELEMETRY';
  const showComm = reportType === 'COMPREHENSIVE' || reportType === 'COMMUNICATION';
  const showAlarms = reportType === 'COMPREHENSIVE' || reportType === 'ALARMS';

  if (showTelemetry && params.length > 0) {
    doc.setFont('helvetica', 'bold');
    doc.setFontSize(8.5);
    doc.setTextColor(15, 23, 42);
    doc.text(`PARAMETER TELEMETRY & STATISTICAL QUALITY METRICS (${params.length} parameters)`, 12, cursorY);
    cursorY += 2;

    autoTable(doc, {
      startY: cursorY,
      head: [[
        'Parameter',
        'Device',
        'Unit',
        'Samples',
        'Min',
        'Avg',
        'Max',
        'First -> Last (Delta)',
        'Quality (G / B / U)',
        'Source',
      ]],
      body: params.map((p) => [
        `${p.parameter_name} (${p.parameter_code})`,
        p.device_code || '-',
        p.unit || '-',
        Number(p.sample_count || 0).toLocaleString(),
        fmtNum(p.min_value, 2),
        fmtNum(p.avg_value, 2),
        fmtNum(p.max_value, 2),
        fmtDelta(p.first_value, p.last_value),
        `${(p.good_percent || 0).toFixed(0)}% / ${(p.bad_percent || 0).toFixed(0)}% / ${(p.uncertain_percent || 0).toFixed(0)}%`,
        p.data_source === 'RAW_TELEMETRY' ? 'RAW' : (p.data_source || '-'),
      ]),
      theme: 'grid',
      styles: { fontSize: 7, cellPadding: 1.5, textColor: [30, 41, 59] },
      headStyles: { fillColor: [15, 23, 42], textColor: [255, 255, 255], fontStyle: 'bold' },
      alternateRowStyles: { fillColor: [248, 250, 252] },
      margin: { left: 12, right: 12 },
    });
    cursorY = (doc.lastAutoTable && doc.lastAutoTable.finalY ? doc.lastAutoTable.finalY : cursorY) + 6;
  }

  if (showComm && devices.length > 0) {
    if (cursorY > 175) {
      doc.addPage();
      cursorY = 15;
    }
    doc.setFont('helvetica', 'bold');
    doc.setFontSize(8.5);
    doc.setTextColor(15, 23, 42);
    doc.text(`DEVICE FLEET AVAILABILITY & COMMUNICATION AUDIT (${devices.length} nodes)`, 12, cursorY);
    cursorY += 2;

    autoTable(doc, {
      startY: cursorY,
      head: [['Device Code', 'Device Name', 'Status', 'Latency', 'Packets (OK / ERR)', 'Parameters', 'Availability']],
      body: devices.map((d) => [
        d.device_code || '-',
        d.device_name || '-',
        d.status || '-',
        `${d.latency_ms || 0} ms`,
        `${d.success_count || 0} / ${d.failed_count || 0}`,
        `${d.parameters_count || 0} active`,
        `${(d.availability_percent || 0).toFixed(1)}%`,
      ]),
      theme: 'grid',
      styles: { fontSize: 7, cellPadding: 1.5, textColor: [30, 41, 59] },
      headStyles: { fillColor: [30, 41, 59], textColor: [255, 255, 255], fontStyle: 'bold' },
      alternateRowStyles: { fillColor: [248, 250, 252] },
      margin: { left: 12, right: 12 },
    });
    cursorY = (doc.lastAutoTable && doc.lastAutoTable.finalY ? doc.lastAutoTable.finalY : cursorY) + 6;
  }

  if (showAlarms) {
    if (cursorY > 175) {
      doc.addPage();
      cursorY = 15;
    }
    doc.setFont('helvetica', 'bold');
    doc.setFontSize(8.5);
    doc.setTextColor(15, 23, 42);
    doc.text(`OPERATIONAL ALARMS & EXCEPTION LOG (${alm.total_alarms || 0} recorded)`, 12, cursorY);
    cursorY += 2;

    autoTable(doc, {
      startY: cursorY,
      head: [['Alarm Code', 'Device', 'Severity', 'Status', 'Message', 'Triggered Time']],
      body: alarms.length > 0
        ? alarms.map((a) => [
            a.alarm_code || '-',
            a.device_name || '-',
            a.severity || '-',
            a.status || '-',
            a.message || '-',
            fmtTs(a.triggered_at),
          ])
        : [['-', '-', '-', '-', 'No critical or warning alarms occurred during this reporting interval.', '-']],
      theme: 'grid',
      styles: { fontSize: 7, cellPadding: 1.5, textColor: [30, 41, 59] },
      headStyles: { fillColor: [127, 29, 29], textColor: [255, 255, 255], fontStyle: 'bold' },
      alternateRowStyles: { fillColor: [248, 250, 252] },
      margin: { left: 12, right: 12 },
    });
  }

  const totalPages = doc.getNumberOfPages();
  for (let p = 1; p <= totalPages; p++) {
    doc.setPage(p);
    doc.setFontSize(7);
    doc.setTextColor(148, 163, 184);
    doc.text(`Halaman ${p} dari ${totalPages}`, 285, 204, { align: 'right' });
  }

  const slug = String(reportType || 'comprehensive').toLowerCase();
  triggerReportBlobDownload(doc.output('blob'), `laporan_operasional_${slug}_${Date.now()}.pdf`);
}

