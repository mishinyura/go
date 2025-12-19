package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gitlab.com/education/users-api/internal/model"
	"gitlab.com/education/users-api/internal/service"
)

// DemoSheetHandler показывает минимальный обмен с Google Sheets (2 колонки).
type DemoSheetHandler struct {
	service service.SheetDemoService
}

func NewDemoSheetHandler(s service.SheetDemoService) *DemoSheetHandler {
	return &DemoSheetHandler{service: s}
}

func (h *DemoSheetHandler) Register(r *gin.RouterGroup) {
	demo := r.Group("/demo")
	{
		// export → запись данных в Google Sheets
		demo.POST("/export", h.ExportDemo)
		// import → чтение указанной таблицы по query
		demo.GET("/import", h.ImportDemo)
		// download → сценарий “по кнопке” без параметров
		demo.GET("/download", h.DownloadDefault)
	}
}

// ExportDemo godoc
// @Summary      Заполнить таблицу демо-данными
// @Description  Пишет три строки с категориями и суммами в указанный лист
// @Tags         demo
// @Accept       json
// @Produce      json
// @Param        payload  body      model.ExportSimpleRequest  true  "Настройки листа"
// @Success      200      {object}  model.ExportSimpleResponse
// @Router       /demo/export [post]
func (h *DemoSheetHandler) ExportDemo(c *gin.Context) {
	var req model.ExportSimpleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.Rows) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "rows must not be empty"})
		return
	}
	resp, err := h.service.ExportDemo(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ImportDemo godoc
// @Summary      Прочитать значения из двух колонок
// @Description  Возвращает все заполненные строки из колонок A и B
// @Tags         demo
// @Produce      json
// @Param        spreadsheet_id  query     string  true  "ID таблицы"
// @Param        sheet_name      query     string  false "Название листа" default(Report)
// @Success      200             {object}  map[string][]model.SimpleRow
// @Failure      400             {object}  map[string]string
// @Router       /demo/import [get]
func (h *DemoSheetHandler) ImportDemo(c *gin.Context) {
	spreadsheetID := c.Query("spreadsheet_id")
	if spreadsheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "spreadsheet_id is required"})
		return
	}
	sheetName := c.DefaultQuery("sheet_name", "Report")

	rows, err := h.service.ImportDemo(c.Request.Context(), spreadsheetID, sheetName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": rows})
}

// DownloadDefault godoc
// @Summary      Download (без параметров)
// @Description  Считывает столбцы A/B из таблицы, указанной в переменных окружения
// @Tags         demo
// @Produce      json
// @Success      200  {object}  map[string][]model.SimpleRow
// @Failure      500  {object}  map[string]string
// @Router       /demo/download [get]
func (h *DemoSheetHandler) DownloadDefault(c *gin.Context) {
	// запрос ничего не передаёт, поэтому используем зашитые в сервис параметры
	rows, err := h.service.DownloadDefault(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"rows": rows})
}
