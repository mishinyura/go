package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gitlab.com/education/gateway/internal/model"
	"gitlab.com/education/gateway/internal/service"
)

type BudgetHandler struct {
	service service.BudgetGatewayService
}

func NewBudgetHandler(s service.BudgetGatewayService) *BudgetHandler {
	if s == nil {
		panic("BudgetHandler requires service")
	}
	return &BudgetHandler{service: s}
}

func (h *BudgetHandler) Register(r *gin.RouterGroup) {
	budget := r.Group("/budget")
	{
		budget.POST("/export", h.ExportBudget)
		budget.GET("/import", h.ImportBudget)
		budget.GET("/download", h.DownloadBudget)
	}
}

func (h *BudgetHandler) ExportBudget(c *gin.Context) {
	var req model.ExportSimpleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rows must not be empty"})
		return
	}
	resp, err := h.service.ExportBudget(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *BudgetHandler) ImportBudget(c *gin.Context) {
	spreadsheetID := c.Query("spreadsheet_id")
	if spreadsheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "spreadsheet_id is required"})
		return
	}
	sheetName := c.DefaultQuery("sheet_name", "Report")
	rows, err := h.service.ImportBudget(c.Request.Context(), spreadsheetID, sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": rows})
}

func (h *BudgetHandler) DownloadBudget(c *gin.Context) {
	rows, err := h.service.DownloadBudget(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": rows})
}
