package model

// SimpleRow — минимальная запись из двух колонок в таблице (категория + сумма).
type SimpleRow struct {
	Category string  `json:"category" example:"Продукты"`
	Amount   float64 `json:"amount" example:"1250.50"`
}

// ExportSimpleRequest описывает лист, куда сервис выгружает демо-данные.
type ExportSimpleRequest struct {
	SpreadsheetID string      `json:"spreadsheet_id" binding:"required" example:"1AbCDefGhijk"`
	SheetName     string      `json:"sheet_name" binding:"required" example:"Report"`
	Clear         bool        `json:"clear"`
	Rows          []SimpleRow `json:"rows" binding:"required,min=1"`
}

// ExportSimpleResponse возвращает диапазон и ссылку после записи.
type ExportSimpleResponse struct {
	WrittenRange   string `json:"written_range" example:"Report!A1:B4"`
	Rows           int    `json:"rows" example:"3"`
	SpreadsheetURL string `json:"spreadsheet_url" example:"https://docs.google.com/spreadsheets/d/<ID>/edit"`
}
