package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gitlab.com/education/users-api/internal/model"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// SheetsClient инкапсулирует самые простые операции с таблицами.
type SheetsClient interface {
	WriteSimple(ctx context.Context, req model.ExportSimpleRequest, rows []model.SimpleRow) (*model.ExportSimpleResponse, error)
	ReadSimple(ctx context.Context, spreadsheetID, sheetName string) ([]model.SimpleRow, error)
}

type sheetsClient struct {
	svc *sheets.Service
}

func NewSheetsClient(ctx context.Context, credentialsPath string) (SheetsClient, error) {
	if credentialsPath == "" {
		return nil, errors.New("переменная GOOGLE_APPLICATION_CREDENTIALS не задана")
	}
	// создаём официальный клиент Google Sheets по JSON-ключу
	service, err := sheets.NewService(ctx, option.WithCredentialsFile(credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("create sheets service: %w", err)
	}
	return &sheetsClient{svc: service}, nil
}

func (c *sheetsClient) WriteSimple(ctx context.Context, req model.ExportSimpleRequest, rows []model.SimpleRow) (*model.ExportSimpleResponse, error) {
	rangeA1 := fmt.Sprintf("%s!A1:B", quoteSheetName(req.SheetName))
	if req.Clear {
		// очищаем диапазон, чтобы видеть только свежие данные
		if _, err := c.svc.Spreadsheets.Values.Clear(req.SpreadsheetID, rangeA1, &sheets.ClearValuesRequest{}).Context(ctx).Do(); err != nil {
			return nil, fmt.Errorf("clear range: %w", err)
		}
	}

	values := [][]interface{}{{"Категория", "Сумма"}}
	for _, row := range rows {
		values = append(values, []interface{}{row.Category, row.Amount})
	}

	body := &sheets.ValueRange{Values: values}
	res, err := c.svc.Spreadsheets.Values.Update(req.SpreadsheetID, rangeA1, body).
		ValueInputOption("USER_ENTERED").
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("update values: %w", err)
	}

	return &model.ExportSimpleResponse{
		WrittenRange:   res.UpdatedRange,
		Rows:           len(rows),
		SpreadsheetURL: fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit", req.SpreadsheetID),
	}, nil
}

func (c *sheetsClient) ReadSimple(ctx context.Context, spreadsheetID, sheetName string) ([]model.SimpleRow, error) {
	rangeA1 := fmt.Sprintf("%s!A2:B", quoteSheetName(sheetName))
	// Google отдаёт “сырой” массив значений, преобразуем его в SimpleRow
	resp, err := c.svc.Spreadsheets.Values.Get(spreadsheetID, rangeA1).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("read values: %w", err)
	}
	if resp == nil || len(resp.Values) == 0 {
		return nil, nil
	}
	rows := make([]model.SimpleRow, 0, len(resp.Values))
	for _, raw := range resp.Values {
		if len(raw) == 0 {
			continue
		}
		category := fmt.Sprint(raw[0])
		if category == "" {
			continue
		}
		var amount float64
		if len(raw) > 1 {
			switch v := raw[1].(type) {
			case float64:
				amount = v
			case string:
				if parsed, err := strconv.ParseFloat(v, 64); err == nil {
					amount = parsed
				}
			}
		}
		rows = append(rows, model.SimpleRow{
			Category: category,
			Amount:   amount,
		})
	}
	return rows, nil
}

func quoteSheetName(name string) string {
	escaped := strings.ReplaceAll(name, "'", "''")
	return fmt.Sprintf("'%s'", escaped)
}
