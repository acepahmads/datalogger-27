package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"datalogger/internal/service"
	"datalogger/pkg/response"

	"github.com/gin-gonic/gin"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) parseReportRequest(c *gin.Context) service.ReportRequest {
	var req service.ReportRequest

	// Bind from JSON body if present
	if c.Request.Method == http.MethodPost && c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&req)
	}

	// Override or populate with query params if provided
	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			req.StartTime = t
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			req.EndTime = t
		}
	}
	if devIDStr := c.Query("device_id"); devIDStr != "" {
		if id, err := strconv.ParseUint(devIDStr, 10, 32); err == nil {
			req.DeviceIDs = append(req.DeviceIDs, uint(id))
		}
	}
	if devIDsStr := c.Query("device_ids"); devIDsStr != "" {
		parts := strings.Split(devIDsStr, ",")
		for _, p := range parts {
			if id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32); err == nil {
				req.DeviceIDs = append(req.DeviceIDs, uint(id))
			}
		}
	}
	if paramIDStr := c.Query("parameter_id"); paramIDStr != "" {
		if id, err := strconv.ParseUint(paramIDStr, 10, 32); err == nil {
			req.ParameterIDs = append(req.ParameterIDs, uint(id))
		}
	}
	if paramIDsStr := c.Query("parameter_ids"); paramIDsStr != "" {
		parts := strings.Split(paramIDsStr, ",")
		for _, p := range parts {
			if id, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32); err == nil {
				req.ParameterIDs = append(req.ParameterIDs, uint(id))
			}
		}
	}
	if repType := c.Query("report_type"); repType != "" {
		req.ReportType = repType
	}
	if tz := c.Query("timezone"); tz != "" {
		req.Timezone = tz
	}
	if fmtStr := c.Query("format"); fmtStr != "" {
		req.Format = fmtStr
	}

	return req
}

// PreviewReport returns structured JSON report data for UI rendering
func (h *ReportHandler) PreviewReport(c *gin.Context) {
	req := h.parseReportRequest(c)

	data, err := h.reportService.BuildReport(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Failed to compile report: "+err.Error())
		return
	}

	response.OK(c, data)
}

// GenerateReport returns HTML, CSV, or JSON for export and printing
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	req := h.parseReportRequest(c)

	data, err := h.reportService.BuildReport(c.Request.Context(), req)
	if err != nil {
		response.InternalError(c, "Failed to compile report: "+err.Error())
		return
	}

	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format == "" {
		format = strings.ToLower(c.Query("format"))
	}
	if format == "" {
		format = "html"
	}

	switch format {
	case "csv":
		csvBytes, err := h.reportService.GenerateCSV(data)
		if err != nil {
			response.InternalError(c, "Failed to generate CSV report: "+err.Error())
			return
		}
		filename := fmt.Sprintf("report-%s-%s.csv", strings.ToLower(data.ReportType), time.Now().Format("20060102-150405"))
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Data(http.StatusOK, "text/csv; charset=utf-8", csvBytes)

	case "html":
		htmlStr, err := h.reportService.GenerateHTML(data)
		if err != nil {
			response.InternalError(c, "Failed to generate HTML report: "+err.Error())
			return
		}
		// If download query param is set, trigger file download, else render inline
		if c.Query("download") == "true" || c.Query("download") == "1" {
			filename := fmt.Sprintf("report-%s-%s.html", strings.ToLower(data.ReportType), time.Now().Format("20060102-150405"))
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(htmlStr))

	default: // json
		response.OK(c, data)
	}
}
