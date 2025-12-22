package grpcserver

import (
	"context"

	"gitlab.com/education/users-api/internal/model"
	pb "gitlab.com/education/users-api/internal/pb/budget/v1"
	"gitlab.com/education/users-api/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BudgetServer struct {
	pb.UnimplementedBudgetServiceServer
	budgetService service.SheetDemoService
}

var _ pb.BudgetServiceServer = (*BudgetServer)(nil)

func NewBudgetServer(svc service.SheetDemoService) *BudgetServer {
	return &BudgetServer{budgetService: svc}
}

func (s *BudgetServer) ExportBudget(ctx context.Context, req *pb.ExportBudgetRequest) (*pb.ExportBudgetResponse, error) {
	if req.GetSpreadsheetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "spreadsheet_id is required")
	}
	if req.GetSheetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "sheet_name is required")
	}
	if len(req.GetRows()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "rows must not be empty")
	}

	modelReq := model.ExportSimpleRequest{
		SpreadsheetID: req.GetSpreadsheetId(),
		SheetName:     req.GetSheetName(),
		Clear:         req.GetClear(),
		Rows:          toModelRows(req.GetRows()),
	}

	resp, err := s.budgetService.ExportDemo(ctx, modelReq)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "export budget: %v", err)
	}

	return &pb.ExportBudgetResponse{
		WrittenRange:   resp.WrittenRange,
		Rows:           int32(resp.Rows),
		SpreadsheetUrl: resp.SpreadsheetURL,
	}, nil
}

func (s *BudgetServer) ImportBudget(ctx context.Context, req *pb.ImportBudgetRequest) (*pb.ImportBudgetResponse, error) {
	if req.GetSpreadsheetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "spreadsheet_id is required")
	}

	sheetName := req.GetSheetName()
	if sheetName == "" {
		sheetName = "Report"
	}

	rows, err := s.budgetService.ImportDemo(ctx, req.GetSpreadsheetId(), sheetName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "import budget: %v", err)
	}

	return &pb.ImportBudgetResponse{
		Rows: toProtoRows(rows),
	}, nil
}

func (s *BudgetServer) DownloadBudget(ctx context.Context, _ *pb.DownloadBudgetRequest) (*pb.DownloadBudgetResponse, error) {
	rows, err := s.budgetService.DownloadDefault(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "download budget: %v", err)
	}

	return &pb.DownloadBudgetResponse{
		Rows: toProtoRows(rows),
	}, nil
}

func toModelRows(protoRows []*pb.SimpleRow) []model.SimpleRow {
	modelRows := make([]model.SimpleRow, 0, len(protoRows))
	for _, r := range protoRows {
		modelRows = append(modelRows, model.SimpleRow{
			Category: r.GetCategory(),
			Amount:   r.GetAmount(),
		})
	}
	return modelRows
}

func toProtoRows(modelRows []model.SimpleRow) []*pb.SimpleRow {
	protoRows := make([]*pb.SimpleRow, 0, len(modelRows))
	for i := range modelRows {
		protoRows = append(protoRows, &pb.SimpleRow{
			Category: modelRows[i].Category,
			Amount:   modelRows[i].Amount,
		})
	}
	return protoRows
}
