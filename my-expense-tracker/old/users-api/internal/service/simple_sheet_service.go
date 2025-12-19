package service

import (
	"context"
	"errors"

	"gitlab.com/education/users-api/internal/model"
)

// DemoSheetConfig хранит параметры листа, с которым работает демо-поток.
type DemoSheetConfig struct {
	SpreadsheetID string
	SheetName     string
}

// SheetDemoService предоставляет самый простой сценарий обмена с Google Sheets.
type SheetDemoService interface {
	ExportDemo(ctx context.Context, req model.ExportSimpleRequest) (*model.ExportSimpleResponse, error)
	ImportDemo(ctx context.Context, spreadsheetID, sheetName string) ([]model.SimpleRow, error)
	DownloadDefault(ctx context.Context) ([]model.SimpleRow, error)
}

type sheetDemoService struct {
	sheets SheetsClient
	cfg    DemoSheetConfig
}

func NewSheetDemoService(client SheetsClient, cfg DemoSheetConfig) SheetDemoService {
	if client == nil {
		panic("SheetDemoService requires SheetsClient")
	}
	if cfg.SpreadsheetID == "" {
		panic("DemoSheetConfig.SpreadsheetID is required")
	}
	if cfg.SheetName == "" {
		cfg.SheetName = "Report"
	}
	return &sheetDemoService{
		sheets: client,
		cfg:    cfg,
	}
}

func (s *sheetDemoService) ExportDemo(ctx context.Context, req model.ExportSimpleRequest) (*model.ExportSimpleResponse, error) {
	if len(req.Rows) == 0 {
		return nil, errors.New("rows payload must not be empty")
	}
	// отправляем те строки, которые пришли в запросе
	return s.sheets.WriteSimple(ctx, req, req.Rows)
}

func (s *sheetDemoService) ImportDemo(ctx context.Context, spreadsheetID, sheetName string) ([]model.SimpleRow, error) {
	return s.sheets.ReadSimple(ctx, spreadsheetID, sheetName)
}

func (s *sheetDemoService) DownloadDefault(ctx context.Context) ([]model.SimpleRow, error) {
	// чтение “по умолчанию”: параметры берём из окружения, запрос пустой
	return s.sheets.ReadSimple(ctx, s.cfg.SpreadsheetID, s.cfg.SheetName)
}
