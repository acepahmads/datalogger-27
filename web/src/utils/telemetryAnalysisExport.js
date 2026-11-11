import * as XLSX from 'xlsx';
import { jsPDF } from 'jspdf';
import autoTable from 'jspdf-autotable';

export function triggerDownloadBlob(blob, filename) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  setTimeout(() => URL.revokeObjectURL(url), 1500);
}

export function triggerDownloadDataUrl(dataUrl, filename) {
  const a = document.createElement('a');
  a.href = dataUrl;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}

export function buildAnalysisExportTable(ctx) {
  const {
    activeMode,
    records = [],
    isOperator = false,
    canAccessRawMode = false,
    getDeviceCode,
    getParamCode,
    getParamUnit,
    formatTimestamp,
    formatTimeOnly,
    formatVal,
    formatUncalibratedVal,
    formatCalibrationDelta,
    formatInterval,
  } = ctx;

  if (activeMode === 'HISTORICAL') {
    const headers = ['No', 'Timestamp'];
    if (!isOperator) headers.push('Device');
    headers.push('Parameter', 'Processed Value');
    if (canAccessRawMode && !isOperator) headers.push('Raw Register');
    headers.push('Unit', 'Quality', 'Reason / Flags');

    const rows = records.map((r, idx) => {
      const flags = [
        r.is_held_value ? '[HELD]' : '',
        r.quality_reason && r.quality_reason !== 'NONE' ? r.quality_reason : '',
      ].filter(Boolean).join(' ') || '--';
      const procVal = r.processed_value !== undefined && r.processed_value !== null ? r.processed_value : r.value;
      const row = [idx + 1, formatTimestamp(r.received_at || r.timestamp)];
      if (!isOperator) row.push(getDeviceCode(r.device_id));
      row.push(getParamCode(r.parameter_id), Number.isFinite(Number(procVal)) ? Number(Number(procVal).toFixed(4)) : formatVal(procVal));
      if (canAccessRawMode && !isOperator) {
        row.push(r.raw_value !== undefined && r.raw_value !== null ? Number(r.raw_value) : '--');
      }
      row.push(getParamUnit(r.parameter_id) || '-', r.quality || 'GOOD', flags);
      return row;
    });

    const colWidths = headers.map((h) => {
      if (h === 'No') return { wch: 6 };
      if (h === 'Timestamp') return { wch: 22 };
      if (h === 'Reason / Flags') return { wch: 20 };
      return { wch: 15 };
    });

    return { headers, rows, colWidths };
  }

  if (activeMode === 'RAW' && canAccessRawMode && !isOperator) {
    const headers = [
      'No',
      'Received Time',
      'Device',
      'Parameter',
      'Raw Hex',
      'Raw Register',
      'Sebelum Kalibrasi (x)',
      'Scaled Value',
      'Selisih Kalibrasi (Δ)',
      'Unit',
      'Quality',
      'Protocol',
      'UUID',
    ];

    const rows = records.map((r, idx) => [
      idx + 1,
      formatTimestamp(r.received_at || r.timestamp),
      getDeviceCode(r.device_id),
      getParamCode(r.parameter_id),
      r.raw_hex || '--',
      r.raw_value !== undefined && r.raw_value !== null ? Number(r.raw_value) : '--',
      formatUncalibratedVal(r),
      r.value !== undefined && r.value !== null && !isNaN(r.value) ? Number(Number(r.value).toFixed(4)) : '--',
      formatCalibrationDelta(r),
      getParamUnit(r.parameter_id) || '-',
      r.quality || 'GOOD',
      r.source || 'MODBUS_TCP',
      r.record_uuid || String(r.id || ''),
    ]);

    const colWidths = [
      { wch: 6 }, { wch: 22 }, { wch: 14 }, { wch: 16 }, { wch: 14 },
      { wch: 14 }, { wch: 18 }, { wch: 14 }, { wch: 18 }, { wch: 10 },
      { wch: 11 }, { wch: 14 }, { wch: 24 },
    ];
    return { headers, rows, colWidths };
  }

  // AGGREGATED mode
  const headers = ['No', 'Period Window', 'Identifier'];
  if (!isOperator) headers.push('Device');
  headers.push('Parameter', 'Value', 'Unit', 'Function', 'Interval', 'Min / Max', 'Samples', 'Quality');

  const rows = records.map((r, idx) => {
    const row = [
      idx + 1,
      `${formatTimestamp(r.period_start)} -> ${formatTimeOnly(r.period_end)}`,
      r.identifier || '--',
    ];
    if (!isOperator) row.push(getDeviceCode(r.device_id));
    row.push(
      getParamCode(r.parameter_id),
      r.value !== undefined && r.value !== null && !isNaN(r.value) ? Number(Number(r.value).toFixed(4)) : '--',
      getParamUnit(r.parameter_id) || '-',
      r.function || 'AVG',
      formatInterval(r.interval_seconds),
      `${formatVal(r.min_value, 1)} / ${formatVal(r.max_value, 1)}`,
      `${r.sample_count || 0} (${r.good_count || 0}G)`,
      r.quality || 'GOOD'
    );
    return row;
  });

  const colWidths = headers.map((h) => {
    if (h === 'No') return { wch: 6 };
    if (h === 'Period Window') return { wch: 30 };
    return { wch: 15 };
  });

  return { headers, rows, colWidths };
}

export function buildSeriesSummaryTable(seriesList = [], isOperator = false) {
  const headers = ['No'];
  if (!isOperator) headers.push('Device');
  headers.push('Parameter', 'Name', 'Unit', 'Latest', 'Min', 'Max', 'Average', 'Samples', 'Good (%)');

  const rows = seriesList.map((s, idx) => {
    const st = s.stats || {};
    const row = [idx + 1];
    if (!isOperator) row.push(s.deviceCode || '--');
    row.push(
      s.paramCode || '--',
      s.paramName || '--',
      s.unit || '-',
      st.latest !== null && st.latest !== undefined ? st.latest : '--',
      st.min !== null && st.min !== undefined ? st.min : '--',
      st.max !== null && st.max !== undefined ? st.max : '--',
      st.avg !== null && st.avg !== undefined ? st.avg : '--',
      st.sampleCount || 0,
      `${st.goodPercent !== undefined ? st.goodPercent : 100}%`
    );
    return row;
  });

  const colWidths = headers.map((h) => {
    if (h === 'No') return { wch: 6 };
    if (h === 'Name') return { wch: 24 };
    return { wch: 14 };
  });

  return { headers, rows, colWidths };
}

export function exportAnalysisXLS(ctx) {
  const { activeMode, seriesList = [], isOperator = false } = ctx;
  const { headers, rows, colWidths } = buildAnalysisExportTable(ctx);
  const summary = buildSeriesSummaryTable(seriesList, isOperator);

  const wb = XLSX.utils.book_new();

  const wsData = XLSX.utils.aoa_to_sheet([headers, ...rows]);
  wsData['!cols'] = colWidths;
  XLSX.utils.book_append_sheet(wb, wsData, 'Data_Telemetri');

  if (summary.rows.length > 0) {
    const wsSummary = XLSX.utils.aoa_to_sheet([summary.headers, ...summary.rows]);
    wsSummary['!cols'] = summary.colWidths;
    XLSX.utils.book_append_sheet(wb, wsSummary, 'Ringkasan_Statistik');
  }

  const xlsBuffer = XLSX.write(wb, { bookType: 'xlsx', type: 'array' });
  const blob = new Blob([xlsBuffer], {
    type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
  });
  triggerDownloadBlob(blob, `telemetry_${activeMode.toLowerCase()}_${Date.now()}.xlsx`);
}

export async function renderAnalysisChartDataUrl({
  svgEl,
  isDark = false,
  format = 'png',
  titleText = 'Telemetry Trend Analysis',
  subtitleText = '',
  visibleSeries = [],
  isOperator = false,
}) {
  if (!svgEl) {
    throw new Error('Elemen grafik belum tersedia untuk diekspor.');
  }

  const clone = svgEl.cloneNode(true);
  const crosshair = clone.querySelector('.chart-hover-crosshair');
  if (crosshair && crosshair.parentNode) {
    crosshair.parentNode.removeChild(crosshair);
  }

  const vb = (svgEl.getAttribute('viewBox') || '').trim().split(/\s+/).map(Number);
  const svgW = (vb.length === 4 && vb[2] > 0) ? vb[2] : (svgEl.clientWidth || 900);
  const svgH = (vb.length === 4 && vb[3] > 0) ? vb[3] : (svgEl.clientHeight || 260);

  clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
  clone.setAttribute('width', String(svgW));
  clone.setAttribute('height', String(svgH));

  const serializer = new XMLSerializer();
  const svgString = serializer.serializeToString(clone);

  const headerH = 46;
  const legendRows = Math.max(1, Math.ceil(visibleSeries.length / 4));
  const footerH = visibleSeries.length > 0 ? 18 + legendRows * 20 : 12;
  const totalW = Math.max(svgW + 32, 860);
  const totalH = headerH + svgH + footerH;

  const canvas = document.createElement('canvas');
  canvas.width = totalW * 2;
  canvas.height = totalH * 2;
  const ctx = canvas.getContext('2d');
  ctx.scale(2, 2);

  // Solid background (essential for JPEG & clean PNG reports)
  ctx.fillStyle = isDark ? '#0B0F19' : '#FFFFFF';
  ctx.fillRect(0, 0, totalW, totalH);

  // Header divider & title
  ctx.fillStyle = isDark ? '#0F172A' : '#F8FAFC';
  ctx.fillRect(0, 0, totalW, headerH - 4);
  ctx.strokeStyle = isDark ? '#1E293B' : '#E2E8F0';
  ctx.lineWidth = 1;
  ctx.beginPath();
  ctx.moveTo(0, headerH - 4);
  ctx.lineTo(totalW, headerH - 4);
  ctx.stroke();

  ctx.fillStyle = isDark ? '#F8FAFC' : '#0F172A';
  ctx.font = 'bold 13px sans-serif';
  ctx.fillText(titleText, 16, 20);

  if (subtitleText) {
    ctx.fillStyle = isDark ? '#94A3B8' : '#64748B';
    ctx.font = '11px monospace';
    ctx.fillText(subtitleText, 16, 35);
  }

  // Draw SVG chart onto canvas
  await new Promise((resolve, reject) => {
    const svgBlob = new Blob([svgString], { type: 'image/svg+xml;charset=utf-8' });
    const url = URL.createObjectURL(svgBlob);
    const img = new Image();
    img.onload = () => {
      const drawW = totalW - 32;
      ctx.drawImage(img, 16, headerH, drawW, svgH);
      URL.revokeObjectURL(url);
      resolve();
    };
    img.onerror = (err) => {
      URL.revokeObjectURL(url);
      reject(err);
    };
    img.src = url;
  });

  // Draw compact legend strip at bottom
  if (visibleSeries.length > 0) {
    const colW = Math.floor((totalW - 32) / Math.min(4, visibleSeries.length));
    visibleSeries.forEach((s, idx) => {
      const col = idx % 4;
      const row = Math.floor(idx / 4);
      const lx = 16 + col * colW;
      const ly = headerH + svgH + 16 + row * 20;

      ctx.fillStyle = s.color || '#3B82F6';
      ctx.beginPath();
      ctx.arc(lx + 5, ly - 4, 4, 0, Math.PI * 2);
      ctx.fill();

      const labelPrefix = !isOperator && s.deviceCode ? `${s.deviceCode}:` : '';
      const st = s.stats || {};
      const statTxt = st.latest !== null && st.latest !== undefined
        ? `${st.latest}${s.unit ? ' ' + s.unit : ''} (avg ${st.avg ?? '--'})`
        : (s.unit || '');
      ctx.fillStyle = isDark ? '#E2E8F0' : '#1E293B';
      ctx.font = 'bold 10px monospace';
      ctx.fillText(`${labelPrefix}${s.paramCode}  ${statTxt}`.slice(0, 34), lx + 14, ly);
    });
  }

  const isJpg = format === 'jpg' || format === 'jpeg';
  return canvas.toDataURL(isJpg ? 'image/jpeg' : 'image/png', 0.95);
}

export async function exportAnalysisChartImage(opts) {
  const fmt = (opts.format || 'png').toLowerCase() === 'jpg' ? 'jpg' : 'png';
  const dataUrl = await renderAnalysisChartDataUrl({ ...opts, format: fmt });
  const modeSlug = (opts.activeMode || 'historical').toLowerCase();
  triggerDownloadDataUrl(dataUrl, `grafik_telemetri_${modeSlug}_${Date.now()}.${fmt}`);
}

export async function exportAnalysisPDF(ctx) {
  const {
    activeMode,
    timeWindowLabel = '',
    svgEl = null,
    visibleSeries = [],
    seriesList = [],
    isDark = false,
    isOperator = false,
  } = ctx;

  const doc = new jsPDF({
    orientation: 'landscape',
    unit: 'mm',
    format: 'a4',
  });

  // Top Navy Header Banner
  doc.setFillColor(15, 23, 42);
  doc.rect(0, 0, 297, 19, 'F');
  doc.setFont('helvetica', 'bold');
  doc.setFontSize(12);
  doc.setTextColor(255, 255, 255);
  doc.text(`LAPORAN ANALISIS TELEMETRI — ${activeMode}`, 12, 9);

  doc.setFont('helvetica', 'normal');
  doc.setFontSize(8);
  doc.setTextColor(203, 213, 225);
  doc.text('INDUSTRIAL ENVIRONMENT MONITORING & TELEMETRY SYSTEM', 12, 14.5);

  doc.setFontSize(7.5);
  doc.setTextColor(148, 163, 184);
  if (timeWindowLabel) {
    doc.text(`Periode: ${timeWindowLabel}`, 285, 9, { align: 'right' });
  }
  doc.text(`Dicetak: ${new Date().toLocaleString()}`, 285, 14.5, { align: 'right' });

  let cursorY = 23;

  // 1. Embed Trend Chart if series are visible
  if (svgEl && visibleSeries.length > 0) {
    try {
      const chartPng = await renderAnalysisChartDataUrl({
        svgEl,
        isDark,
        format: 'png',
        titleText: `Telemetry Trend (${activeMode})`,
        subtitleText: timeWindowLabel,
        visibleSeries,
        isOperator,
      });
      doc.addImage(chartPng, 'PNG', 12, cursorY, 273, 76);
      cursorY += 80;
    } catch (err) {
      console.warn('[TelemetryAnalysis PDF] Chart image embed skipped:', err);
    }
  }

  // 2. Per-Series Statistical Summary Table
  const summary = buildSeriesSummaryTable(seriesList, isOperator);
  if (summary.rows.length > 0) {
    autoTable(doc, {
      startY: cursorY,
      head: [summary.headers],
      body: summary.rows,
      theme: 'grid',
      styles: { fontSize: 7, cellPadding: 1.3, textColor: [30, 41, 59] },
      headStyles: { fillColor: [30, 41, 59], textColor: [255, 255, 255], fontStyle: 'bold' },
      alternateRowStyles: { fillColor: [248, 250, 252] },
      margin: { left: 12, right: 12 },
    });
    cursorY = (doc.lastAutoTable && doc.lastAutoTable.finalY ? doc.lastAutoTable.finalY : cursorY) + 5;
  }

  // 3. Telemetry Data Records Table
  const dataTbl = buildAnalysisExportTable(ctx);
  if (dataTbl.rows.length > 0) {
    autoTable(doc, {
      startY: cursorY,
      head: [dataTbl.headers],
      body: dataTbl.rows,
      theme: 'grid',
      styles: { fontSize: 6.8, cellPadding: 1.2, textColor: [30, 41, 59] },
      headStyles: { fillColor: [15, 23, 42], textColor: [255, 255, 255], fontStyle: 'bold' },
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

  const pdfBlob = doc.output('blob');
  triggerDownloadBlob(pdfBlob, `laporan_telemetri_${activeMode.toLowerCase()}_${Date.now()}.pdf`);
}

