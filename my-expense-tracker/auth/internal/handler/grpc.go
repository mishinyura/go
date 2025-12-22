package handler

import (
	"context"

	"github.com/yuramishin/expense-tracker/auth/internal/service"
	pb "github.com/yuramishin/expense-tracker/proto/pb_auth"
)

type GrpcHandler struct {
	pb.UnimplementedAuthServiceServer
	svc *service.AuthService
}

func NewGrpcHandler(svc *service.AuthService) *GrpcHandler {
	return &GrpcHandler{svc: svc}
}

func (h *GrpcHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	id, err := h.svc.Register(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.RegisterResponse{UserId: id}, nil
}

func (h *GrpcHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	token, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginResponse{Token: token}, nil
}

func (h *GrpcHandler) Validate(ctx context.Context, req *pb.ValidateRequest) (*pb.ValidateResponse, error) {
	userID, valid := h.svc.Validate(ctx, req.Token)
	return &pb.ValidateResponse{Valid: valid, UserId: userID}, nil
}

func (h *GrpcHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	err := h.svc.Logout(ctx, req.Token)
	return &pb.LogoutResponse{Success: err == nil}, err
}
