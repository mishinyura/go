package service

import (
	"context"

	"gitlab.com/education/gateway/internal/model"
	budgetv1 "gitlab.com/education/gateway/internal/pb/budget/v1"
)

type BudgetGatewayService interface {
	ExportBudget(ctx context.Context, req model.ExportSimpleRequest) (*model.ExportSimpleResponse, error)
	ImportBudget(ctx context.Context, spreadsheetID, sheetName string) ([]model.SimpleRow, error)
	DownloadBudget(ctx context.Context) ([]model.SimpleRow, error)
}

type budgetGatewayService struct {
	client budgetv1.BudgetServiceClient
}

func NewBudgetGatewayService(client budgetv1.BudgetServiceClient) BudgetGatewayService {
	if client == nil {
		panic("budget gateway service requires gRPC client")
	}
	return &budgetGatewayService{client: client}
}

func (s *budgetGatewayService) ExportBudget(ctx context.Context, req model.ExportSimpleRequest) (*model.ExportSimpleResponse, error) {
	resp, err := s.client.ExportBudget(ctx, &budgetv1.ExportBudgetRequest{
		SpreadsheetId: req.SpreadsheetID,
		SheetName:     req.SheetName,
		Clear:         req.Clear,
		Rows:          toPbRows(req.Rows),
	})
	if err != nil {
		return nil, err
	}
	return &model.ExportSimpleResponse{
		WrittenRange:   resp.WrittenRange,
		Rows:           int(resp.Rows),
		SpreadsheetURL: resp.SpreadsheetUrl,
	}, nil
}

func (s *budgetGatewayService) ImportBudget(ctx context.Context, spreadsheetID, sheetName string) ([]model.SimpleRow, error) {
	resp, err := s.client.ImportBudget(ctx, &budgetv1.ImportBudgetRequest{
		SpreadsheetId: spreadsheetID,
		SheetName:     sheetName,
	})
	if err != nil {
		return nil, err
	}
	return fromPbRows(resp.GetRows()), nil
}

func (s *budgetGatewayService) DownloadBudget(ctx context.Context) ([]model.SimpleRow, error) {
	resp, err := s.client.DownloadBudget(ctx, &budgetv1.DownloadBudgetRequest{})
	if err != nil {
		return nil, err
	}
	return fromPbRows(resp.GetRows()), nil
}

func toPbRows(rows []model.SimpleRow) []*budgetv1.SimpleRow {
	out := make([]*budgetv1.SimpleRow, 0, len(rows))
	for _, r := range rows {
		row := r
		out = append(out, &budgetv1.SimpleRow{
			Category: row.Category,
			Amount:   row.Amount,
		})
	}
	return out
}

func fromPbRows(rows []*budgetv1.SimpleRow) []model.SimpleRow {
	out := make([]model.SimpleRow, 0, len(rows))
	for _, r := range rows {
		if r == nil {
			continue
		}
		out = append(out, model.SimpleRow{
			Category: r.Category,
			Amount:   r.Amount,
		})
	}
	return out
}
