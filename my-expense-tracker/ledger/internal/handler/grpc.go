package handler

import (
	"context"

	"github.com/yuramishin/expense-tracker/ledger/internal/domain"
	"github.com/yuramishin/expense-tracker/ledger/internal/service"
	pb "github.com/yuramishin/expense-tracker/proto/pb_ledger"
)

type GrpcHandler struct {
	pb.UnimplementedLedgerServiceServer
	service *service.LedgerService
}

func NewGrpcHandler(s *service.LedgerService) *GrpcHandler {
	return &GrpcHandler{service: s}
}

func (h *GrpcHandler) CreateTransaction(ctx context.Context, req *pb.TransactionRequest) (*pb.TransactionResponse, error) {
	t := &domain.Transaction{
		UserID:      req.UserId,
		Amount:      req.Amount,
		Category:    req.Category,
		Description: req.Description,
	}
	success, msg := h.service.CreateTransaction(ctx, t)
	return &pb.TransactionResponse{Success: success, Message: msg}, nil
}

func (h *GrpcHandler) GetReport(ctx context.Context, req *pb.ReportRequest) (*pb.ReportResponse, error) {
	data, err := h.service.GetReport(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.ReportResponse{ByCategory: data}, nil
}

func (h *GrpcHandler) SetBudget(ctx context.Context, req *pb.BudgetRequest) (*pb.BudgetResponse, error) {
	err := h.service.SetBudget(ctx, req.UserId, req.Category, req.LimitAmount)
	if err != nil {
		return &pb.BudgetResponse{Success: false, Message: err.Error()}, nil
	}
	return &pb.BudgetResponse{Success: true, Message: "Budget Set"}, nil
}

func (h *GrpcHandler) GetBudgets(ctx context.Context, req *pb.GetBudgetsRequest) (*pb.BudgetList, error) {
	list, err := h.service.GetBudgets(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	var pbList []*pb.Budget
	for _, b := range list {
		pbList = append(pbList, &pb.Budget{Category: b.Category, LimitAmount: b.LimitAmount})
	}
	return &pb.BudgetList{Budgets: pbList}, nil
}
